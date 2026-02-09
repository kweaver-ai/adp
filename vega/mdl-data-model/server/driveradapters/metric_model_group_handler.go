// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	derrors "data-model/errors"
	"data-model/interfaces"
)

// 创建指标模型分组
func (r *restHandler) CreateMetricModelGroup(c *gin.Context) {
	logger.Debug("Handler CreateMetricModelGroup Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"创建指标模型分组", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 接受绑定参数
	metricModelGroup := interfaces.MetricModelGroup{}
	err = c.ShouldBindJSON(&metricModelGroup)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModelGroup_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateMetricModelGroupAuditObject("", ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("创建指标模型分组请求参数: [%s,%v]", c.Request.RequestURI, metricModelGroup.GroupName))

	// 校验 指标模型分组信息
	err = validateMetricModelGroup(ctx, &metricModelGroup)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateMetricModelGroupAuditObject("", metricModelGroup.GroupName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Validate metric model  group name [%s] failed: %s. %v", metricModelGroup.GroupName,
			httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		span.SetAttributes(attr.Key("group_name").String(metricModelGroup.GroupName))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 调用当前 service 的校验放在handler中,调用其他service做的校验放在当前handler对应的service中
	// 校验 请求体与现有指标模型分组名称的重复性
	exist, err := r.mmgs.CheckMetricModelGroupExist(ctx, metricModelGroup.GroupName)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateMetricModelGroupAuditObject("", metricModelGroup.GroupName), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_MetricModelGroup_GroupNameExisted)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateMetricModelGroupAuditObject("", metricModelGroup.GroupName), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//调用创建
	groupID, err := r.mmgs.CreateMetricModelGroup(ctx, metricModelGroup)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateMetricModelGroupAuditObject("", metricModelGroup.GroupName), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	metricModelGroup.GroupID = groupID

	//每次成功创建 记录审计日志
	audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		GenerateMetricModelGroupAuditObject(groupID, metricModelGroup.GroupName), "")

	result := map[string]string{"id": groupID}

	logger.Debug("Handler CreateMetricModelGroup Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	// c.Writer.Header().Set("Location", "/api/mdl-data-model/v1/metric-models-groups/"+groupID)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// 更新指标模型分组
func (r *restHandler) UpdateMetricModelGroup(c *gin.Context) {
	logger.Debug("Handler UpdateMetricModelGroup Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"修改指标模型分组", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 接受绑定参数
	group := interfaces.MetricModelGroup{}
	err = c.ShouldBindJSON(&group)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModelGroup_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateMetricModelGroupAuditObject("", ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("修改指标模型分组请求参数: [%s,%v]", c.Request.RequestURI, group))

	//接受 group_id 参数
	groupID := c.Param("group_id")
	span.SetAttributes(attr.Key("group_id").String(groupID))

	//禁止修改默认分组
	if strings.Trim(groupID, " ") == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModelGroup_InvalidParameter).
			WithErrorDetails(" You can not change the default group!")

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateMetricModelGroupAuditObject(groupID, group.GroupName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	group.GroupID = groupID

	// 校验 指标模型分组信息的合法性, 非空、长度、特殊字符
	err = validateMetricModelGroup(ctx, &group)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateMetricModelGroupAuditObject(groupID, group.GroupName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Validate metric model group name [%s] failed: %s. %v", group.GroupName,
			httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		span.SetAttributes(attr.Key("group_name").String(group.GroupName))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 先按id获取原对象
	oldGroup, err := r.mmgs.GetMetricModelGroupByID(ctx, groupID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateMetricModelGroupAuditObject(groupID, group.GroupName), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 内置分组不能修改
	// if oldMetricModelGroup.Builtin {
	// 	httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_MetricModelGroup_ForbiddenDeleteBuiltinGroup).
	// 		WithErrorDetails(fmt.Sprintf("The group id[%s] is built-in group, is not allowed to be updated", groupID))

	// 	audit.NewWarnLogWithError(c, audit.OPERATION, audit.DELETE,
	// 		interfaces.OBJECTTYPE_METRIC_MODEL_GROUP, groupID, audit.FAILED, &httpErr.BaseError)

	// 	// 设置 trace 的错误信息的 attributes
	// 	span.SetAttributes(attr.Key("group_id").String(groupID))
	// 	o11y.AddHttpAttrs4Error(span, httpErr.HTTPCode, httpErr.BaseError.ErrorCode,
	// 		fmt.Sprintf("The group id[%s] is built-in group, is not allowed to be deleted", groupID))

	// 	rest.ReplyError(c, httpErr)
	// 	return
	// }

	//不同的话需要校验名称重复性
	if oldGroup.GroupName != group.GroupName {
		exist, err := r.mmgs.CheckMetricModelGroupExist(ctx, group.GroupName)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateMetricModelGroupAuditObject(groupID, group.GroupName), &httpErr.BaseError)

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		if exist {
			httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_MetricModelGroup_GroupNameExisted)

			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateMetricModelGroupAuditObject(groupID, group.GroupName), &httpErr.BaseError)

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	//根据id修改信息
	err = r.mmgs.UpdateMetricModelGroup(ctx, group)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateMetricModelGroupAuditObject(groupID, group.GroupName), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		GenerateMetricModelGroupAuditObject(groupID, group.GroupName), "")

	logger.Debug("Handler UpdateMetricModelGroup Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 查询指标模型分组列表
func (r *restHandler) ListMetricModelGroups(c *gin.Context) {
	logger.Debug("ListMetricModelGroups Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"获取指标模型分组列表", trace.WithSpanKind(trace.SpanKindServer))
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

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("获取指标模型分组列表请求参数: [%s]", c.Request.RequestURI))

	// 获取分页参数
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", interfaces.DEFAULT_METRIC_MODEL_GROUP_SORT)
	direction := c.DefaultQuery("direction", interfaces.ASC_DIRECTION)
	builtinQueryArr := c.QueryArray("builtin")

	// 校验分页查询参数
	pageParam, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.METRIC_MODEL_GROUP_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验builtin参数
	builtinArr := make([]bool, 0, len(builtinQueryArr))
	for _, val := range builtinQueryArr {
		if val == "" {
			continue
		}
		builtin, err := strconv.ParseBool(val)
		if err != nil {
			errDetails := fmt.Sprintf(`The value of param 'builtin' should be bool type, but got '%s'`, val)
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
				derrors.DataModel_MetricModelGroup_InvalidParameter_Builtin).
				WithErrorDetails(errDetails)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		builtinArr = append(builtinArr, builtin)
	}

	queryParam := interfaces.ListMetricGroupQueryParams{
		Builtin:                   builtinArr,
		PaginationQueryParameters: pageParam,
	}

	// 获取指标模型
	entries, total, err := r.mmgs.ListMetricModelGroups(ctx, queryParam)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]interface{}{
		"total_count": total,
		"entries":     entries,
	}

	logger.Debug("Handler ListMetricModelGroups Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// 刪除指标模型分组
func (r *restHandler) DeleteMetricModelGroup(c *gin.Context) {
	logger.Debug("Handler DeleteMetricModelGroup Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"删除指标模型分组", trace.WithSpanKind(trace.SpanKindServer))
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

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("删除指标模型分组请求参数: [%s]", c.Request.RequestURI))

	//获取参数字符串
	groupID := c.Param("group_id")
	span.SetAttributes(attr.Key("group_id").String(groupID))

	if strings.Trim(groupID, " ") == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModelGroup_InvalidParameter).
			WithErrorDetails(" You can not delete the default group!")

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateMetricModelGroupAuditObject(groupID, ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//获取force 并校验
	forceStr := c.DefaultQuery("force", interfaces.DEFAULT_FORCE)
	force, err := strconv.ParseBool(forceStr)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_MetricModelGroup_InvalidParameter_Force).
			WithErrorDetails(fmt.Sprintf("The forc:%s is invalid", forceStr))

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateMetricModelGroupAuditObject(groupID, ""), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//检查 groupID 是否存在
	group, err := r.mmgs.GetMetricModelGroupByID(ctx, groupID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateMetricModelGroupAuditObject(groupID, ""), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验分组 ID：内置分组不能删除
	// if metricModelGroup.Builtin {
	// 	httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_MetricModelGroup_ForbiddenDeleteBuiltinGroup).
	// 		WithErrorDetails(fmt.Sprintf("The group id[%s] is built-in group, is not allowed to be deleted", groupID))

	// 	audit.NewWarnLogWithError(c, audit.OPERATION, audit.DELETE,
	// 		interfaces.OBJECTTYPE_METRIC_MODEL_GROUP, groupID, audit.FAILED, &httpErr.BaseError)

	// 	// 设置 trace 的错误信息的 attributes
	// 	span.SetAttributes(attr.Key("group_id").String(groupID))
	// 	o11y.AddHttpAttrs4Error(span, httpErr.HTTPCode, httpErr.BaseError.ErrorCode,
	// 		fmt.Sprintf("The group id[%s] is built-in group, is not allowed to be deleted", groupID))

	// 	rest.ReplyError(c, httpErr)
	// 	return
	// }

	rowsAffect, metricModels, err := r.mmgs.DeleteMetricModelGroup(ctx, groupID, force)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateMetricModelGroupAuditObject(groupID, ""), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//删除成功记录审计日志
	audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
		GenerateMetricModelGroupAuditObject(groupID, group.GroupName), audit.SUCCESS, "")
	if rowsAffect != 0 {
		for _, model := range metricModels {
			audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				GenerateMetricModelAuditObject(model.ModelID, model.ModelName), audit.SUCCESS, "")
		}
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 导出分组下的指标模型
func (r *restHandler) GetMetricModelsInGroup(c *gin.Context) {
	logger.Debug("Handler GetMetricModelsInGroup Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Get metric models", trace.WithSpanKind(trace.SpanKindServer))
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

	groupID := c.Param("group_name") // url路径参数用group_name来接，实参是 group_id
	span.SetAttributes(attr.Key("group_id").String(groupID))

	//检查 groupID 是否存在
	_, err = r.mmgs.GetMetricModelGroupByID(ctx, groupID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	models, err := r.mms.GetMetricModelsDetailByGroupID(ctx, groupID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler GetMetricModelsInGroup Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, models)
}
