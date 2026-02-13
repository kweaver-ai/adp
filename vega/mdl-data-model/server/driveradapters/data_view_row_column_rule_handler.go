// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
)

// 创建数据视图行列权限（外部）
func (r *restHandler) CreateDataViewRowColumnRulesByEx(c *gin.Context) {
	logger.Debug("Handler CreateDataViewRowColumnRulesByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Create data view row column rules by ex", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.CreateDataViewRowColumnRules(c, visitor)
}

// 创建数据视图行列权限(内部)
func (r *restHandler) CreateDataViewRowColumnRulesByIn(c *gin.Context) {
	logger.Debug("Handler CreateDataViewRowColumnRulesByIn Start")
	visitor := GenerateVisitor(c)
	r.CreateDataViewRowColumnRules(c, visitor)
}

// 创建数据视图行列权限
func (r *restHandler) CreateDataViewRowColumnRules(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler CreateDataViewRowColumnRules Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Create data view row column rules", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	reqBody := []interfaces.DataViewRowColumnRule{}
	err := c.ShouldBindJSON(&reqBody)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
			WithErrorDetails("Binding paramter failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataViewRowColumnRuleAuditObject("", ""), &httpErr.BaseError)

		// 记录错误log
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置attributes和status
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 如果传入的视图行列规则为[], 应报错
	if len(reqBody) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
			WithErrorDetails("No data view row column rule was passed in")

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataViewRowColumnRuleAuditObject("", ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	nameMap := make(map[string]any)
	idMap := make(map[string]any)
	rules := make([]*interfaces.DataViewRowColumnRule, 0, len(reqBody))
	for i := 0; i < len(reqBody); i++ {
		ruleID := reqBody[i].RuleID
		ruleName := reqBody[i].RuleName
		viewID := reqBody[i].ViewID
		uk_rule_name := fmt.Sprintf("%s_%s", ruleName, viewID)

		// 如果传了规则ID，校验请求体中多个规则 ID 是否重复
		if ruleID != "" {
			if _, ok := idMap[ruleID]; !ok {
				idMap[ruleID] = nil
			} else {
				errDetails := fmt.Sprintf("data view row column rule ID '%s' already exists in the request body", ruleID)
				httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
					WithErrorDetails(errDetails)

				audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
					GenerateDataViewRowColumnRuleAuditObject(ruleID, ruleName), &httpErr.BaseError)

				o11y.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}

		// 校验请求体中多个规则名称是否在视图内重复
		if _, ok := nameMap[uk_rule_name]; !ok {
			nameMap[uk_rule_name] = nil
		} else {
			errDetails := fmt.Sprintf("data view row column rule name '%s' already exists within the view '%s' in the request body", ruleName, viewID)
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails(errDetails)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataViewRowColumnRuleAuditObject(ruleID, ruleName), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		rule := &interfaces.DataViewRowColumnRule{
			RuleID:     ruleID,
			RuleName:   ruleName,
			ViewID:     viewID,
			Tags:       reqBody[i].Tags,
			Comment:    reqBody[i].Comment,
			Fields:     reqBody[i].Fields,
			RowFilters: reqBody[i].RowFilters,
		}

		// 校验数据视图必要创建参数的合法性
		err = validateDataViewRowColumnRule(ctx, rule)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataViewRowColumnRuleAuditObject(ruleID, ruleName), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		rules = append(rules, rule)
	}

	// 批量创建
	ruleIDs, err := r.dvrcs.CreateDataViewRowColumnRules(ctx, rules)
	if err != nil {
		// 失败，发送一条
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataViewRowColumnRuleAuditObject(rules[0].RuleID, rules[0].RuleName), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 成功，发送多条
	for _, rule := range rules {
		audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataViewRowColumnRuleAuditObject(rule.RuleID, rule.RuleName), "")
	}

	result := make([]interface{}, 0, len(ruleIDs))
	for _, ruleID := range ruleIDs {
		result = append(result, map[string]interface{}{"id": ruleID})
	}

	logger.Debug("Handler CreateDataViewRowColumnRules Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// 删除数据视图行列权限（外部）
func (r *restHandler) DeleteDataViewRowColumnRulesByEx(c *gin.Context) {
	logger.Debug("Handler DeleteDataViewRowColumnRulesByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Delete data view row column rules by ex", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.DeleteDataViewRowColumnRules(c, visitor)
}

// 删除数据视图行列权限(内部)
func (r *restHandler) DeleteDataViewRowColumnRulesByIn(c *gin.Context) {
	logger.Debug("Handler DeleteDataViewRowColumnRulesByIn Start")

	visitor := GenerateVisitor(c)
	r.DeleteDataViewRowColumnRules(c, visitor)
}

// 删除数据视图行列权限，支持批量删除
func (r *restHandler) DeleteDataViewRowColumnRules(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler DeleteDataViewRowColumnRules Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Delete data view row column rules", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	ruleIDsStr := c.Param("rule_ids")
	ruleIDs := common.StringToStringSlice(ruleIDsStr)

	// 校验rule_ids是否为空
	if len(ruleIDs) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails("rule_ids is required")

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataViewRowColumnRuleAuditObject("", ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	span.SetAttributes(attr.Key("rule_ids").String(fmt.Sprintf("%v", ruleIDs)))

	// 校验规则是否存在，有一个不存在则返回错误
	ruleIDNameMap := make(map[string]string)
	for _, ruleID := range ruleIDs {
		ruleName, err := r.dvrcs.CheckDataViewRowColumnRuleExistByID(ctx, ruleID)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				GenerateDataViewRowColumnRuleAuditObject(ruleID, ruleName), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		ruleIDNameMap[ruleID] = ruleName
	}

	err := r.dvrcs.DeleteDataViewRowColumnRules(ctx, ruleIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataViewRowColumnRuleAuditObject("", ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 成功，发送多条
	for ruleID, ruleName := range ruleIDNameMap {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataViewRowColumnRuleAuditObject(ruleID, ruleName), audit.SUCCESS, "")
	}

	logger.Debug("Handler DeleteDataViewRowColumnRules Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 更新数据视图行列权限（外部）
func (r *restHandler) UpdateDataViewRowColumnRuleByEx(c *gin.Context) {
	logger.Debug("Handler UpdateDataViewRowColumnRuleByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Update a data view row column rule by ex", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.UpdateDataViewRowColumnRule(c, visitor)
}

// 更新数据视图行列权限(内部)
func (r *restHandler) UpdateDataViewRowColumnRuleByIn(c *gin.Context) {
	logger.Debug("Handler UpdateDataViewRowColumnRuleByIn Start")

	visitor := GenerateVisitor(c)
	r.UpdateDataViewRowColumnRule(c, visitor)
}

// 更新数据视图行列权限
func (r *restHandler) UpdateDataViewRowColumnRule(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler UpdateDataViewRowColumnRule Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Update a data view row column rules", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	ruleID := c.Param("rule_id")
	span.SetAttributes(attr.Key("rule_id").String(ruleID))

	if ruleID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataViewRowColumnRule_NullParameter_RuleID).
			WithErrorDetails("rule_id is required")

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewRowColumnRuleAuditObject(ruleID, ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	ruleInfo := &interfaces.DataViewRowColumnRule{}
	err := c.ShouldBindJSON(&ruleInfo)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
			WithErrorDetails("Failed to bind param: " + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewRowColumnRuleAuditObject(ruleID, ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	ruleInfo.RuleID = ruleID

	// 参数校验
	err = validateDataViewRowColumnRule(ctx, ruleInfo)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewRowColumnRuleAuditObject(ruleID, ruleInfo.RuleName), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	err = r.dvrcs.UpdateDataViewRowColumnRule(ctx, ruleInfo)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewRowColumnRuleAuditObject(ruleID, ruleInfo.RuleName), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		GenerateDataViewRowColumnRuleAuditObject(ruleID, ruleInfo.RuleName), "")

	logger.Debug("Handler updateDataView Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 按 id 获取数据视图行列权限对象信息（外部）
func (r *restHandler) GetDataViewRowColumnRulesByEx(c *gin.Context) {
	logger.Debug("Handler GetDataViewRowColumnRulesByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Get data view row column rules by ex", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetDataViewRowColumnRules(c, visitor)
}

// 按 id 获取数据视图行列权限对象信息(内部)
func (r *restHandler) GetDataViewRowColumnRulesByIn(c *gin.Context) {
	logger.Debug("Handler GetDataViewRowColumnRulesByIn Start")

	visitor := GenerateVisitor(c)
	r.GetDataViewRowColumnRules(c, visitor)
}

// 获取数据视图行列权限详情
func (r *restHandler) GetDataViewRowColumnRules(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler GetDataViewRowColumnRules Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Get data view row column rules", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	ruleIDsStr := c.Param("rule_ids")
	ruleIDs := common.StringToStringSlice(ruleIDsStr)

	// 如果路径参数不存在，则为重载接口
	if len(ruleIDs) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
			WithErrorDetails("Param 'rule_ids' is required")

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	ruleInfos, err := r.dvrcs.GetDataViewRowColumnRules(ctx, ruleIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler GetDataViewRowColumnRules Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, ruleInfos)
}

// 分页获取数据视图行列权限列表（外部）
func (r *restHandler) ListDataViewRowColumnRulesByEx(c *gin.Context) {
	logger.Debug("Handler ListDataViewRowColumnRulesByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: List data view row column rules by ex", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ListDataViewRowColumnRules(c, visitor, false)
}

// 分页获取数据视图行列权限列表(内部)
func (r *restHandler) ListDataViewRowColumnRulesByIn(c *gin.Context) {
	logger.Debug("Handler ListDataViewRowColumnRulesByIn Start")

	visitor := GenerateVisitor(c)
	r.ListDataViewRowColumnRules(c, visitor, true)
}

// 获取数据视图行列权限列表
func (r *restHandler) ListDataViewRowColumnRules(c *gin.Context, visitor rest.Visitor, isInnerRequest bool) {
	logger.Debug("Handler ListDataViewRowColumnRules Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: List data view row column rules", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	name := c.Query("name")
	namePattern := c.Query("name_pattern")
	viewID := c.Query("view_id")
	tag := c.Query("tag")

	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", "update_time")
	direction := c.DefaultQuery("direction", interfaces.DESC_DIRECTION)

	err := validateNameandNamePattern(ctx, name, namePattern)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 分页参数校验
	PageParam, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.DATA_VIEW_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//去掉标签前后的所有空格进行搜索
	tag = strings.Trim(tag, " ")

	param := &interfaces.ListRowColumnRuleQueryParams{
		Name:                      name,
		NamePattern:               namePattern,
		ViewID:                    viewID,
		Tag:                       tag,
		IsInnerRequest:            isInnerRequest,
		PaginationQueryParameters: PageParam,
	}

	rules, total, err := r.dvrcs.ListDataViewRowColumnRules(ctx, param)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]interface{}{
		"entries":     rules,
		"total_count": total,
	}

	logger.Debug("Handler ListDataViewRowColumnRules Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// 分页获取数据视图行列权限资源列表
func (r *restHandler) ListDataViewRowColumnRuleSrcs(c *gin.Context) {
	logger.Debug("Handler ListDataViewRowColumnRuleSrcs Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: List data view row column rule sources", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))
	o11y.Info(ctx, fmt.Sprintf("driver layer: List data view sources request parameters: [%s]", c.Request.RequestURI))

	// 获取分页参数
	namePattern := c.Query(RESOURCES_KEYWOED)
	name := c.Query("name")
	tag := c.Query("tag")
	viewID := c.Query("view_id")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", RESOURCES_PAGE_LIMIT)
	sort := c.DefaultQuery("sort", "group_name")
	direction := c.DefaultQuery("direction", interfaces.ASC_DIRECTION)

	err = validateNameandNamePattern(ctx, name, namePattern)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 分页参数校验
	PageParam, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.DATA_VIEW_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//去掉标签前后的所有空格进行搜索
	tag = strings.Trim(tag, " ")

	param := &interfaces.ListRowColumnRuleQueryParams{
		Name:                      name,
		NamePattern:               namePattern,
		ViewID:                    viewID,
		Tag:                       tag,
		PaginationQueryParameters: PageParam,
	}

	resources, total, err := r.dvrcs.ListDataViewRowColumnRuleSrcs(ctx, param)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	result := map[string]interface{}{"entries": resources, "total_count": total}

	logger.Debug("Handler ListDataViewRowColumnRulesSrcs Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}
