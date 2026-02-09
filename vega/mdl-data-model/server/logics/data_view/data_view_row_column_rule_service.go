// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_view

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	"data-model/logics"
	"data-model/logics/permission"
)

var (
	dvrcrServiceOnce sync.Once
	dvrcrService     interfaces.DataViewRowColumnRuleService
)

type dataViewRowColumnRuleService struct {
	appSetting *common.AppSetting
	ps         interfaces.PermissionService
	dva        interfaces.DataViewAccess
	dvrcra     interfaces.DataViewRowColumnRuleAccess
}

func NewDataViewRowColumnRuleService(appSetting *common.AppSetting) interfaces.DataViewRowColumnRuleService {
	dvrcrServiceOnce.Do(func() {
		dvrcrService = &dataViewRowColumnRuleService{
			appSetting: appSetting,
			ps:         permission.NewPermissionService(appSetting),
			dva:        logics.DVA,
			dvrcra:     logics.DVRCRA,
		}
	})

	return dvrcrService
}

// 创建数据视图行列规则
func (dvrcrs *dataViewRowColumnRuleService) CreateDataViewRowColumnRules(ctx context.Context, dataViewRowColumnRules []*interfaces.DataViewRowColumnRule) ([]string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Create logical view row column rules")
	defer span.End()

	viewIDs := make([]string, 0, len(dataViewRowColumnRules))
	for _, rule := range dataViewRowColumnRules {
		viewIDs = append(viewIDs, rule.ViewID)
	}

	// 判断userid对于当前viewID是否有行列规则管理的权限（策略决策）
	matchResouces, err := dvrcrs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW, viewIDs,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL, interfaces.OPERATION_TYPE_RULE_MANAGE}, false)
	if err != nil {
		return nil, err
	}
	// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
	for _, mID := range viewIDs {
		if _, exist := matchResouces[mID]; !exist {
			return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for row column rule's manage operation")
		}
	}

	currentTime := time.Now().UnixMilli()
	ruleIDs := make([]string, 0, len(dataViewRowColumnRules))
	for _, rule := range dataViewRowColumnRules {
		// 校验创建参数
		err = dvrcrs.validateCreateUpdateParams(ctx, rule)
		if err != nil {
			return nil, err
		}

		// 校验规则名称是否存在
		_, exist, httpErr := dvrcrs.CheckDataViewRowColumnRuleExistByName(ctx, rule.RuleName, rule.ViewID)
		if httpErr != nil {
			return nil, httpErr
		}
		if exist {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
				derrors.DataModel_DataViewRowColumnRule_ExistByName).
				WithErrorDetails(fmt.Sprintf("rule name %s already exist", rule.RuleName))
		}

		// 如果规则ID为空，则生成一个
		if rule.RuleID == "" {
			rule.RuleID = xid.New().String()
		}

		accountInfo := interfaces.AccountInfo{}
		if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
			accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
		}

		rule.Creator = accountInfo
		rule.Updater = accountInfo
		rule.CreateTime = currentTime
		rule.UpdateTime = currentTime

		ruleIDs = append(ruleIDs, rule.RuleID)

	}

	err = dvrcrs.dvrcra.CreateDataViewRowColumnRules(ctx, dataViewRowColumnRules)
	if err != nil {
		logger.Errorf("Create logical view row column rules error: %s", err.Error())
		span.SetStatus(codes.Error, "create logical view row column rules failed")

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return ruleIDs, nil
}

// 删除数据视图行列权限
func (dvrcrs *dataViewRowColumnRuleService) DeleteDataViewRowColumnRules(ctx context.Context, ruleIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Delete logical view row column rules")
	defer span.End()

	// 获取视图ID
	rules, err := dvrcrs.dvrcra.GetSimpleRulesByRuleIDs(ctx, nil, ruleIDs)
	if err != nil {
		return err
	}
	viewIDs := make([]string, 0, len(rules))
	for _, rule := range rules {
		viewIDs = append(viewIDs, rule.ViewID)
	}

	// 判断userid对于当前viewID是否有行列规则管理的权限（策略决策）
	matchResouces, err := dvrcrs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW, viewIDs,
		[]string{interfaces.OPERATION_TYPE_RULE_MANAGE}, false)
	if err != nil {
		return err
	}
	// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
	for _, mID := range viewIDs {
		if _, exist := matchResouces[mID]; !exist {
			return rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for row column rule's delete operation")
		}
	}

	err = dvrcrs.dvrcra.DeleteDataViewRowColumnRules(ctx, nil, ruleIDs)
	if err != nil {
		logger.Errorf("Delete logical view row column rules error: %s", err.Error())
		span.SetStatus(codes.Error, "delete logical view row column rules failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(err.Error())
	}

	// 清除策略
	err = dvrcrs.ps.DeleteResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW_ROW_COLUMN_RULE, ruleIDs)
	if err != nil {
		logger.Errorf("Delete logical view row column rules error: %s", err.Error())
		span.SetStatus(codes.Error, "delete logical view row column rules failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 删除某个数据视图下的行列权限，内部使用，不校验权限
func (dvrcrs *dataViewRowColumnRuleService) DeleteRowColumnRulesByViewIDs(ctx context.Context, tx *sql.Tx, viewIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Delete logical view row column rules")
	defer span.End()

	// 获取视图下所有的行列规则ID
	rules, err := dvrcrs.dvrcra.GetSimpleRulesByViewIDs(ctx, tx, viewIDs)
	if err != nil {
		logger.Errorf("List logical view row column rules error: %s", err.Error())
		span.SetStatus(codes.Error, "list logical view row column rules failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(err.Error())
	}

	ruleIDs := make([]string, 0, len(rules))
	for _, rule := range rules {
		ruleIDs = append(ruleIDs, rule.RuleID)
	}

	err = dvrcrs.dvrcra.DeleteDataViewRowColumnRules(ctx, tx, ruleIDs)
	if err != nil {
		logger.Errorf("Delete logical view row column rules error: %s", err.Error())
		span.SetStatus(codes.Error, "delete logical view row column rules failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(err.Error())
	}

	// 清除策略
	err = dvrcrs.ps.DeleteResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW_ROW_COLUMN_RULE, ruleIDs)
	if err != nil {
		logger.Errorf("Delete logical view row column rules error: %s", err.Error())
		span.SetStatus(codes.Error, "delete logical view row column rules failed")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 修改数据视图行列权限
func (dvrcrs *dataViewRowColumnRuleService) UpdateDataViewRowColumnRule(ctx context.Context, rule *interfaces.DataViewRowColumnRule) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Update logical view row column rule")
	defer span.End()

	span.SetAttributes(
		attr.Key("data_view_row_column_rule_id").String(rule.RuleID),
		attr.Key("data_view_row_column_rule_name").String(rule.RuleName),
	)

	// 判断userid是否有修改数据视图行列权限的权限（策略决策）
	err := dvrcrs.ps.CheckPermission(ctx,
		interfaces.Resource{
			Type: interfaces.RESOURCE_TYPE_DATA_VIEW,
			ID:   rule.ViewID,
		},
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL, interfaces.OPERATION_TYPE_RULE_MANAGE},
	)
	if err != nil {
		span.SetStatus(codes.Error, "Check permission failed")
		return err
	}

	oldRules, err := dvrcrs.GetDataViewRowColumnRules(ctx, []string{rule.RuleID})
	if err != nil {
		span.SetStatus(codes.Error, "Get logical view row column rules failed")
		return err
	}
	if len(oldRules) == 0 {
		errDetails := fmt.Sprintf("Logical view row column rule '%s' not exists", rule.RuleID)
		logger.Errorf(errDetails)
		span.SetStatus(codes.Error, errDetails)
		return rest.NewHTTPError(ctx, http.StatusNotFound, rest.PublicError_NotFound).
			WithErrorDetails(errDetails)
	}

	oldRule := oldRules[0]

	oldRuleName := oldRule.RuleName
	oldViewID := oldRule.ViewID
	newRuleName := rule.RuleName
	newViewID := rule.ViewID

	// 校验更新参数
	err = dvrcrs.validateCreateUpdateParams(ctx, rule)
	if err != nil {
		return err
	}

	// 视图id不允许变更
	if newViewID != oldViewID {
		errDetails := fmt.Sprintf("Logical view row column rule view id '%s' not allow change, old view id: '%s'", rule.ViewID, oldViewID)
		logger.Errorf(errDetails)
		span.SetStatus(codes.Error, errDetails)
		return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails(errDetails)
	}

	// 校验行列规则名称在当前视图下是否已存在
	// if newViewID != oldViewID || newRuleName != oldRuleName {
	if newRuleName != oldRuleName {
		_, exist, httpErr := dvrcrs.CheckDataViewRowColumnRuleExistByName(ctx, newRuleName, newViewID)
		if httpErr != nil {
			span.SetStatus(codes.Error, "Check logical view exist by name failed")
			return httpErr
		}

		if exist {
			errDetails := fmt.Sprintf("Logical view row column rule '%s' already exists in view '%s'", rule.RuleName, rule.ViewID)
			logger.Errorf(errDetails)
			span.SetStatus(codes.Error, errDetails)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails(errDetails)
		}
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	rule.Updater = accountInfo
	rule.UpdateTime = time.Now().UnixMilli()
	err = dvrcrs.dvrcra.UpdateDataViewRowColumnRule(ctx, rule)
	if err != nil {
		logger.Errorf("update logical view row column rule error, %v", err)
		span.SetStatus(codes.Error, "update logical view row column rule failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 查询数据视图行列权限列表
func (dvrcrs *dataViewRowColumnRuleService) ListDataViewRowColumnRules(ctx context.Context,
	params *interfaces.ListRowColumnRuleQueryParams) ([]*interfaces.DataViewRowColumnRule, int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: List logical view row column rules")
	defer span.End()

	rules, err := dvrcrs.dvrcra.ListDataViewRowColumnRules(ctx, params)
	if err != nil {
		logger.Errorf("ListDataViewRowColumnRules error: %s", err.Error())

		span.SetStatus(codes.Error, "List logical view row column rules failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	if len(rules) == 0 {
		return rules, 0, nil
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	// ruleIDs := make([]string, 0)
	viewIDs := make([]string, 0)
	for _, v := range rules {
		// ruleIDs = append(ruleIDs, v.RuleID)
		viewIDs = append(viewIDs, v.ViewID)
	}

	// matchResouces, err := dvrcrs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW,
	// 	viewIDs, []string{interfaces.OPERATION_TYPE_RULE_MANAGE}, true)
	// if err != nil {
	// 	return nil, 0, err
	// }

	// 获取资源操作
	viewOpsMap, err := dvrcrs.ps.GetResourceOps(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW, viewIDs)
	if err != nil {
		return nil, 0, err
	}

	// 决策是否有rule_manage或rule_authorize权限，有其中一个就ok
	checkPermission := func(slice []string, val1, val2 string) bool {
		found1, found2 := false, false
		for _, v := range slice {
			if v == val1 {
				found1 = true
			}
			if v == val2 {
				found2 = true
			}
			// 如果两个都提前找到了，可以立即退出循环
			if found1 && found2 {
				return true
			}
		}

		if !found1 && !found2 {
			return false
		} else {
			return true
		}
	}

	viewIDMap := make(map[string]interfaces.ResourceOps)
	for _, ops := range viewOpsMap {
		if checkPermission(ops.Operations, interfaces.OPERATION_TYPE_RULE_MANAGE, interfaces.OPERATION_TYPE_RULE_AUTHORIZE) {
			viewIDMap[ops.ResourceID] = ops
		}
	}

	// 根据视图id获取视图名称
	viewMap, err := dvrcrs.dva.GetSimpleDataViewMapByIDs(ctx, viewIDs)
	if err != nil {
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	// 如果包含不存在的，不报错，只打印错误日志
	for _, viewID := range viewIDs {
		if _, ok := viewMap[viewID]; !ok {
			logger.Errorf("Data view '%s' does not exist!", viewID)
		}
	}

	// 遍历对象
	results := make([]*interfaces.DataViewRowColumnRule, 0)
	if params.IsInnerRequest {
		results = rules
	} else {
		for _, rl := range rules {
			if _, exist := viewIDMap[rl.ViewID]; exist {
				// 这里返回的是视图的资源类型
				// v.Operations = resrc.Operations
				rl.ViewName = viewMap[rl.ViewID].ViewName

				results = append(results, rl)
			}
		}
	}

	// limit = -1,则返回所有
	if params.Limit == -1 {
		return results, len(results), nil
	}

	// 分页
	// 检查起始位置是否越界
	if params.Offset < 0 || params.Offset >= len(results) {
		return []*interfaces.DataViewRowColumnRule{}, 0, nil
	}
	// 计算结束位置
	end := params.Offset + params.Limit
	if end > len(results) {
		end = len(results)
	}

	span.SetStatus(codes.Ok, "")
	return results[params.Offset:end], len(results), nil
}

// 获取行列规则的资源实例列表
func (dvrcrs *dataViewRowColumnRuleService) ListDataViewRowColumnRuleSrcs(ctx context.Context,
	params *interfaces.ListRowColumnRuleQueryParams) ([]*interfaces.Resource, int, error) {
	listCtx, listSpan := ar_trace.Tracer.Start(ctx, "logic layer: List logical view row column rule resources")
	listSpan.End()

	rules, err := dvrcrs.dvrcra.ListDataViewRowColumnRules(listCtx, params)
	if err != nil {
		logger.Errorf("ListDataViewRowColumnRules error: %s", err.Error())
		listSpan.SetStatus(codes.Error, "List logical view row column rules error")
		listSpan.End()
		return nil, 0, rest.NewHTTPError(listCtx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}
	if len(rules) == 0 {
		return []*interfaces.Resource{}, 0, nil
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	// 处理资源id
	ruleIDs := make([]string, 0)
	viewIDs := make([]string, 0)
	for _, v := range rules {
		ruleIDs = append(ruleIDs, v.RuleID)
		viewIDs = append(viewIDs, v.ViewID)
	}
	// 校验权限管理的操作权限
	matchResouces, err := dvrcrs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW, ruleIDs,
		[]string{interfaces.OPERATION_TYPE_RULE_MANAGE}, false)
	if err != nil {
		return nil, 0, err
	}

	// 所有有权限的模型数组
	idMap := make(map[string]bool)
	for _, resourceOps := range matchResouces {
		idMap[resourceOps.ResourceID] = true
	}

	// 根据视图id获取视图名称, 如果有一个不存在, 则返回错误
	viewMap, err := dvrcrs.dva.GetSimpleDataViewMapByIDs(ctx, viewIDs)
	if err != nil {
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	// 如果包含不存在的，不报错，只打印错误日志
	for _, viewID := range viewIDs {
		if _, ok := viewMap[viewID]; !ok {
			logger.Errorf("Data view '%s' does not exist!", viewID)
		}
	}

	// 遍历对象
	results := make([]*interfaces.Resource, 0)
	for _, rule := range rules {
		if idMap[rule.RuleID] {
			rule.ViewName = viewMap[rule.ViewID].ViewName

			results = append(results, &interfaces.Resource{
				ID:   rule.RuleID,
				Type: interfaces.RESOURCE_TYPE_DATA_VIEW,
				Name: common.ProcessUngroupedName(ctx, rule.ViewName, rule.RuleName),
			})
		}
	}

	// 分页
	// 检查起始位置是否越界
	if params.Offset < 0 || params.Offset >= len(results) {
		return []*interfaces.Resource{}, 0, nil
	}
	// 计算结束位置
	end := params.Offset + params.Limit
	if end > len(results) {
		end = len(results)
	}

	listSpan.SetStatus(codes.Ok, "")
	return results[params.Offset:end], len(results), nil
}

// 按ID获取数据视图行列权限信息
func (dvrcrs *dataViewRowColumnRuleService) GetDataViewRowColumnRules(ctx context.Context, ruleIDs []string) ([]*interfaces.DataViewRowColumnRule, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("logic layer: Get logical view row column rules '%s' info", strings.Join(ruleIDs, ",")))
	defer span.End()

	span.SetAttributes(attr.Key("rule_ids").String(strings.Join(ruleIDs, ",")))

	rules, err := dvrcrs.dvrcra.GetDataViewRowColumnRules(ctx, ruleIDs)
	if err != nil {
		logger.Errorf("Get logical view row column rule by id error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get logical view row column rules '%s' failed", strings.Join(ruleIDs, ",")))
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	// 找到不存在的视图 id，如果有视图 id 不存在，则返回错误
	if len(rules) < len(ruleIDs) {
		ruleMap := make(map[string]struct{}, len(rules))
		for _, r := range rules {
			ruleMap[r.RuleID] = struct{}{}
		}

		for _, ruleID := range ruleIDs {
			if _, ok := ruleMap[ruleID]; !ok {
				errDetails := fmt.Sprintf("The logical view row column rule %s was not found", ruleID)
				logger.Errorf(errDetails)

				o11y.Error(ctx, errDetails)
				span.SetStatus(codes.Error, errDetails)
				return nil, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataView_DataViewNotFound).
					WithErrorDetails(errDetails)
			}
		}
	}

	// 先获取资源序列
	matchResouces, err := dvrcrs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW, ruleIDs,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		return nil, err
	}
	// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
	for _, mID := range ruleIDs {
		if _, exist := matchResouces[mID]; !exist {
			return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for logical view row column rule's view_detail operation.")
		}
	}

	for index, r := range rules {
		// 补充视图的可操作权限、数据源名称
		r.Operations = matchResouces[r.RuleID].Operations

		rules[index] = r
	}

	span.SetStatus(codes.Ok, "")
	return rules, nil
}

// 根据id检查行列规则是否存在
func (dvrcrs *dataViewRowColumnRuleService) CheckDataViewRowColumnRuleExistByID(ctx context.Context, ruleID string) (string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("logic layer: Check logical view row column rule '%s' existence", ruleID))
	defer span.End()

	span.SetAttributes(attr.Key("rule_id").String(ruleID))

	ruleID, exist, err := dvrcrs.dvrcra.CheckDataViewRowColumnRuleExistByID(ctx, ruleID)
	if err != nil {
		logger.Errorf("Check logical view row column rule existence by id error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("failed to check logical view row column rule '%s' existence", ruleID))
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	// 校验规则是否存在
	if !exist {
		errDetails := fmt.Sprintf("The logical view row column rule %s was not found", ruleID)
		logger.Errorf(errDetails)

		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, errDetails)
		return "", rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataView_DataViewNotFound).
			WithErrorDetails(errDetails)
	}

	span.SetStatus(codes.Ok, "")
	return ruleID, nil
}

// 根据名称检查行列规则是否存在
func (dvrcrs *dataViewRowColumnRuleService) CheckDataViewRowColumnRuleExistByName(ctx context.Context, ruleName string, viewID string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("logic layer: Check logical view row column rule '%s' existence", ruleName))
	defer span.End()

	span.SetAttributes(attr.Key("rule_name").String(ruleName))

	ruleID, exist, err := dvrcrs.dvrcra.CheckDataViewRowColumnRuleExistByName(ctx, ruleName, viewID)
	if err != nil {
		logger.Errorf("Check logical view row column rule existence by name error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("failed to check logical view row column rule '%s' existence", ruleName))
		return ruleID, exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return ruleID, exist, nil
}

// 校验创建和更新的参数
func (dvrcrs *dataViewRowColumnRuleService) validateCreateUpdateParams(ctx context.Context, rule *interfaces.DataViewRowColumnRule) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Validate create/update logical view row column rule params")
	defer span.End()

	// 校验视图是否存在
	if _, exist, err := dvrcrs.dva.CheckDataViewExistByID(ctx, nil, rule.ViewID); !exist {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_DataViewNotFound).
			WithErrorDetails(fmt.Sprintf("view '%s' not found", rule.ViewID))
	} else if err != nil {
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	return nil
}
