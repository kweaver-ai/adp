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

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
)

// 创建链路模型（外部）
func (r *restHandler) CreateTraceModelsByEx(c *gin.Context) {
	logger.Debug("Handler CreateTraceModelsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 批量创建链路模型", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.CreateTraceModels(c, visitor)
}

// 创建链路模型(内部)
func (r *restHandler) CreateTraceModelsByIn(c *gin.Context) {
	logger.Debug("Handler CreateTraceModelsByIn Start")

	visitor := GenerateVisitor(c)
	r.CreateTraceModels(c, visitor)
}

// 批量创建/导入链路模型
func (r *restHandler) CreateTraceModels(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler CreateTraceModels Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 批量创建链路模型", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler CreateTraceModels End")
	}()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. 接收request body
	reqModels := []interfaces.TraceModel{}
	err := c.ShouldBindJSON(&reqModels)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
			WithErrorDetails("Binding paramter failed: " + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject("", ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 如果未传入链路模型, 应报错
	if len(reqModels) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
			WithErrorDetails("No invalid trace model was passed in")

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject("", ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 2. 转换传入的所有链路模型
	err = convertTraceModels(ctx, reqModels)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// temporary solution: 审计日志暂不支持记录对象数组, 所以目前只记录第一个链路模型对象名称
		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject("", reqModels[0].Name), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 3. 校验传入的所有链路模型, 检查参数合法性
	err = validateTraceModelsWhenCreate(ctx, r, reqModels)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// temporary solution: 审计日志暂不支持记录对象数组, 所以目前只记录第一个链路模型对象名称
		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject("", reqModels[0].Name), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 4. 调用logic层, 批量创建/导入链路模型
	modelIDs, err := r.tms.CreateTraceModels(ctx, reqModels)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// temporary solution: 审计日志暂不支持记录对象数组, 所以目前只记录第一个链路模型对象名称
		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject("", reqModels[0].Name), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 5. 构造创建/导入成功的返回结果, 并记录审计日志
	// modelIDsStr := strings.Join(modelIDs, ",")
	result := []map[string]string{}
	for _, modelID := range modelIDs {
		result = append(result, map[string]string{"id": modelID})
	}

	for _, model := range reqModels {
		// 每一个链路模型都要记录审计日志
		audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject(model.ID, model.Name), "")
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	// c.Writer.Header().Set("Location", "/api/mdl-data-model/v1/trace-models/"+modelIDsStr)
	rest.ReplyOK(c, http.StatusCreated, result)
}

func (r *restHandler) SimulateCreateTraceModelByEx(c *gin.Context) {
	logger.Debug("Handler SimulateCreateTraceModelByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 模拟创建链路模型", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.SimulateCreateTraceModel(c, visitor)
}

func (r *restHandler) SimulateCreateTraceModelByIn(c *gin.Context) {
	logger.Debug("Handler SimulateCreateTraceModelByIn Start")

	visitor := GenerateVisitor(c)
	r.SimulateCreateTraceModel(c, visitor)
}

// 模拟创建链路模型
func (r *restHandler) SimulateCreateTraceModel(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler SimulateCreateTraceModel Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 模拟创建链路模型", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler SimulateCreateTraceModel End")
	}()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. 接收request body
	reqModel := interfaces.TraceModel{}
	err := c.ShouldBindJSON(&reqModel)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
			WithErrorDetails("Binding paramter failed: " + err.Error())

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 2. 转换传入的所有链路模型
	reqModels := []interfaces.TraceModel{reqModel}
	err = convertTraceModels(ctx, reqModels)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}

	// 3. 校验传入的所有链路模型, 检查参数合法性
	err = validateTraceModelsWhenCreate(ctx, r, reqModels)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}

	// 3. 调用logic层, 批量创建/导入链路模型
	resModel, err := r.tms.SimulateCreateTraceModel(ctx, reqModels[0])
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}

	// 4. 返回模拟创建成功的结果
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, resModel)
}

func (r *restHandler) DeleteTraceModelsByEx(c *gin.Context) {
	logger.Debug("Handler DeleteTraceModelsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 批量删除链路模型", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.DeleteTraceModels(c, visitor)
}

func (r *restHandler) DeleteTraceModelsByIn(c *gin.Context) {
	logger.Debug("Handler DeleteTraceModelsByIn Start")

	visitor := GenerateVisitor(c)
	r.DeleteTraceModels(c, visitor)
}

// 批量删除链路模型
func (r *restHandler) DeleteTraceModels(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler DeleteTraceModels Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 批量删除链路模型", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler DeleteTraceModels End")
	}()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. 获取模型ID字符串<id1,id2,id3>
	modelIDsStr := c.Param("model_ids")

	// 2. 将ID字符串转换为[]string
	modelIDs := common.StringToStringSlice(modelIDsStr)
	if len(modelIDs) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidParameter_ModelIDs).
			WithErrorDetails("No invalid trace model id was passed in")

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject(modelIDsStr, ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 3. 根据链路模型ID数组去校验链路模型存在性
	modelMap, err := r.tms.GetSimpleTraceModelMapByIDs(ctx, modelIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject(modelIDsStr, ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	notExistIDs := make([]string, 0)
	for _, modelID := range modelIDs {
		if _, ok := modelMap[modelID]; !ok {
			notExistIDs = append(notExistIDs, modelID)
		}
	}

	if len(notExistIDs) > 0 {
		errDetails := fmt.Sprintf("The trace model whose id is in %v does not exist!", notExistIDs)
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_TraceModel_TraceModelNotFound).
			WithErrorDetails(errDetails)

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject(modelIDsStr, ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 4. 批量删除链路模型
	err = r.tms.DeleteTraceModels(ctx, modelIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject(modelIDsStr, ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}

	// 5. 循环记录删除成功的审计日志
	for _, model := range modelMap {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject(model.ID, model.Name), audit.SUCCESS, "")
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

func (r *restHandler) UpdateTraceModelByEx(c *gin.Context) {
	logger.Debug("Handler UpdateTraceModelByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 修改链路模型", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.UpdateTraceModel(c, visitor)
}

func (r *restHandler) UpdateTraceModelByIn(c *gin.Context) {
	logger.Debug("Handler UpdateTraceModelByIn Start")

	visitor := GenerateVisitor(c)
	r.UpdateTraceModel(c, visitor)
}

// 修改链路模型
func (r *restHandler) UpdateTraceModel(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler UpdateTraceModel Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 修改链路模型", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler UpdateTraceModel End")
	}()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. 接收model_id参数
	modelID := c.Param("model_id")

	// 2. 校验model_id
	if modelID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidParameter_ModelIDs).
			WithErrorDetails("Type conversion failed:")

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject("", ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 3. 根据链路模型ID数组去校验链路模型存在性
	modelMap, err := r.tms.GetSimpleTraceModelMapByIDs(ctx, []string{modelID})
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject("", modelID), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if len(modelMap) == 0 {
		errDetails := fmt.Sprintf("The trace model whose id is %v does not exist!", modelID)
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_TraceModel_TraceModelNotFound).
			WithErrorDetails(errDetails)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject("", modelID), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	modelName := modelMap[modelID].Name

	// 4. 接收request body
	reqModel := interfaces.TraceModel{}
	err = c.ShouldBindJSON(&reqModel)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
			WithErrorDetails("Binding paramter failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject(modelID, modelName), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 5. 转换传入的所有链路模型
	reqModel.ID = modelID
	reqModels := []interfaces.TraceModel{reqModel}
	err = convertTraceModels(ctx, reqModels)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject(modelID, modelName), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 6. 校验请求体
	err = validateTraceModelWhenUpdate(ctx, r, modelName, reqModels[0])
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject(modelID, modelName), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 7. 调用logic层修改链路模型
	err = r.tms.UpdateTraceModel(ctx, reqModels[0])
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateTraceModelAuditObject(modelID, modelName), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 8. 记录修改成功的审计日志, 并返回结果
	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		GenerateTraceModelAuditObject(modelID, modelName), "")

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

func (r *restHandler) SimulateUpdateTraceModelByEx(c *gin.Context) {
	logger.Debug("Handler SimulateUpdateTraceModelByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 模拟修改链路模型", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.SimulateUpdateTraceModel(c, visitor)
}

func (r *restHandler) SimulateUpdateTraceModelByIn(c *gin.Context) {
	logger.Debug("Handler SimulateUpdateTraceModelByIn Start")

	visitor := GenerateVisitor(c)
	r.SimulateUpdateTraceModel(c, visitor)
}

// 模拟修改链路模型
func (r *restHandler) SimulateUpdateTraceModel(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler SimulateUpdateTraceModel Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 模拟修改链路模型", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler SimulateUpdateTraceModel End")
	}()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. 接收model_id参数
	modelID := c.Param("model_id")

	// 2. 校验model_id
	if modelID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_TraceModel_InvalidParameter_ModelIDs).
			WithErrorDetails("Type conversion failed:")
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 3. 根据链路模型ID数组去校验链路模型存在性
	var preName string
	modelMap, err := r.tms.GetSimpleTraceModelMapByIDs(ctx, []string{modelID})
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if len(modelMap) == 0 {
		errDetails := fmt.Sprintf("The trace model whose id is %v does not exist!", modelID)
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_TraceModel_TraceModelNotFound).
			WithErrorDetails(errDetails)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	preName = modelMap[modelID].Name

	// 4. 接收request body
	reqModel := interfaces.TraceModel{}
	err = c.ShouldBindJSON(&reqModel)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
			WithErrorDetails("Binding paramter failed:" + err.Error())
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 5. 转换传入的所有链路模型
	reqModel.ID = modelID
	reqModels := []interfaces.TraceModel{reqModel}
	err = convertTraceModels(ctx, reqModels)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 6. 校验请求体
	err = validateTraceModelWhenUpdate(ctx, r, preName, reqModels[0])
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 7. 调用logic层修改链路模型
	resModel, err := r.tms.SimulateUpdateTraceModel(ctx, reqModels[0])
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 8. 返回模拟修改成功的结果
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, resModel)
}

func (r *restHandler) GetTraceModelsByEx(c *gin.Context) {
	logger.Debug("Handler GetTraceModelsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 批量查询链路模型", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetTraceModels(c, visitor)
}

func (r *restHandler) GetTraceModelsByIn(c *gin.Context) {
	logger.Debug("Handler GetTraceModelsByIn Start")

	visitor := GenerateVisitor(c)
	r.GetTraceModels(c, visitor)
}

// 批量查询/导出链路模型
func (r *restHandler) GetTraceModels(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler GetTraceModels Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 批量查询链路模型", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler GetTraceModels End")
	}()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. 获取查询参数
	modelIDstrs := c.Param("model_ids")

	// 2. 将ID字符串转换为[]string
	modelIDs := common.StringToStringSlice(modelIDstrs)
	if len(modelIDs) == 0 {
		o11y.AddHttpAttrs4Ok(span, http.StatusOK)
		rest.ReplyOK(c, http.StatusOK, []interfaces.TraceModel{})
		return
	}

	// 3. 批量获取链路模型
	models, err := r.tms.GetTraceModels(ctx, modelIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, models)
}

func (r *restHandler) ListTraceModelsByEx(c *gin.Context) {
	logger.Debug("Handler ListTraceModelsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 查询链路模型列表与总数", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ListTraceModels(c, visitor)
}

func (r *restHandler) ListTraceModelsByIn(c *gin.Context) {
	logger.Debug("Handler ListTraceModelsByIn Start")

	visitor := GenerateVisitor(c)
	r.ListTraceModels(c, visitor)
}

// 查询链路模型列表, 支持按链路模型名称模糊/精准查询
func (r *restHandler) ListTraceModels(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler ListTraceModels Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 查询链路模型列表与总数", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler ListTraceModels End")
	}()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. 获取查询参数
	namePattern := c.Query("name_pattern")
	name := c.Query("name")
	tag := c.Query("tag")
	spanSourceTypeStr := c.Query("span_source_type")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", interfaces.DEFAULT_SORT)
	direction := c.DefaultQuery("direction", interfaces.DEFAULT_DIRECTION)

	// 2. 校验name_pattern和name
	err := validateNameandNamePattern(ctx, name, namePattern)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}

	// 3. 校验spanSourceType
	spanSourceTypes := []string{}
	if spanSourceTypeStr != "" {
		spanSourceTypes = strings.Split(spanSourceTypeStr, ",")
		for _, spanSourceType := range spanSourceTypes {
			err = validateSpanSourceType(ctx, spanSourceType)
			if err != nil {
				httpErr := err.(*rest.HTTPError)
				o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

				o11y.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, err)
				return
			}
		}
	}

	// 4. 校验分页查询参数
	pageParam, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.TRACE_MODEL_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}

	// 5. 构造链路模型列表查询参数的结构体
	para := interfaces.TraceModelListQueryParams{
		SpanSourceTypes: spanSourceTypes,
		CommonListQueryParams: interfaces.CommonListQueryParams{
			NamePattern: namePattern,
			Name:        name,
			Tag:         tag,
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Offset:    pageParam.Offset,
				Limit:     pageParam.Limit,
				Sort:      pageParam.Sort,
				Direction: pageParam.Direction,
			},
		},
	}

	// 6. 获取链路模型列表
	modelList, total, err := r.tms.ListTraceModels(ctx, para)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 7. 构造返回结果
	result := map[string]interface{}{
		"entries":     modelList,
		"total_count": total,
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

func (r *restHandler) GetTraceModelFieldInfoByEx(c *gin.Context) {
	logger.Debug("Handler GetTraceModelFieldInfoByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 查询链路模型字段信息", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetTraceModelFieldInfo(c, visitor)
}

func (r *restHandler) GetTraceModelFieldInfoByIn(c *gin.Context) {
	logger.Debug("Handler GetTraceModelFieldInfoByIn Start")

	visitor := GenerateVisitor(c)
	r.GetTraceModelFieldInfo(c, visitor)
}

// 查询链路模型字段信息
func (r *restHandler) GetTraceModelFieldInfo(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler GetTraceModelFieldInfo Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 查询链路模型字段信息", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler GetTraceModelFieldInfo End")
	}()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. 接收model_id参数
	modelID := c.Param("model_ids")

	// 2. 校验model_id
	if modelID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_TraceModel_InvalidParameter_ModelIDs).
			WithErrorDetails("Type conversion failed:")

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 3. 查询链路模型字段信息
	fieldInfo, err := r.tms.GetTraceModelFieldInfo(ctx, modelID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, fieldInfo)
}

// 分页获取数据视图资源列表
func (r *restHandler) ListTraceModelSrcs(c *gin.Context) {
	logger.Debug("ListTraceModelSrcs Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: List trace model sources", trace.WithSpanKind(trace.SpanKindServer))
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
	o11y.Info(ctx, fmt.Sprintf("driver layer: List trace model sources request parameters: [%s]", c.Request.RequestURI))

	// 获取分页参数
	namePattern := c.Query(RESOURCES_KEYWOED)
	name := c.Query("name")
	tag := c.Query("tag")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", RESOURCES_PAGE_LIMIT)
	sort := c.DefaultQuery("sort", "name")
	spanSourceTypeStr := c.Query("span_source_type")
	direction := c.DefaultQuery("direction", interfaces.ASC_DIRECTION)

	//去掉标签前后的所有空格进行搜索
	tag = strings.Trim(tag, " ")

	// 校验分页查询参数
	pageParam, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.TRACE_MODEL_SORT)
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

	// name_pattern 和 name 不能同时存在
	if namePattern != "" && name != "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ConflictParameter_NameAndNamePatternCoexist).
			WithErrorDetails("name_pattern and name cannot exists at the same time")

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 3. 校验spanSourceType
	spanSourceTypes := []string{}
	if spanSourceTypeStr != "" {
		spanSourceTypes = strings.Split(spanSourceTypeStr, ",")
		for _, spanSourceType := range spanSourceTypes {
			err = validateSpanSourceType(ctx, spanSourceType)
			if err != nil {
				httpErr := err.(*rest.HTTPError)
				o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

				o11y.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, err)
				return
			}
		}
	}

	para := interfaces.TraceModelListQueryParams{
		SpanSourceTypes: spanSourceTypes,
		CommonListQueryParams: interfaces.CommonListQueryParams{
			NamePattern: namePattern,
			Name:        name,
			Tag:         tag,
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Offset:    pageParam.Offset,
				Limit:     pageParam.Limit,
				Sort:      pageParam.Sort,
				Direction: pageParam.Direction,
			},
		},
	}

	resources, total, err := r.tms.ListTraceModelSrcs(ctx, para)
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
	result := map[string]interface{}{"entries": resources, "total_count": total}

	logger.Debug("Handler ListTraceModelSrcs Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)

}

/*
	私有方法
*/

// 转换传入的trace models
func convertTraceModels(ctx context.Context, reqModels []interfaces.TraceModel) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driver层: 转换传入的链路模型")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	err = convertSpanConfig(ctx, reqModels)
	if err != nil {
		return err
	}

	err = convertRelatedLogConfig(ctx, reqModels)
	if err != nil {
		return err
	}

	return nil
}

// 根据sourceType转换spanConfig的动态类型和动态值
func convertSpanConfig(ctx context.Context, reqModels []interfaces.TraceModel) error {
	for i := range reqModels {
		b, err := sonic.Marshal(reqModels[i].SpanConfig)
		if err != nil {
			errDetails := fmt.Sprintf("Marshal filed span_config failed, err: %v", err.Error())
			o11y.Error(ctx, errDetails)
			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_MarshalDataFailed).WithErrorDetails(errDetails)
		}

		switch reqModels[i].SpanSourceType {
		case interfaces.SOURCE_TYPE_DATA_VIEW:
			conf := interfaces.SpanConfigWithDataView{}
			err = sonic.Unmarshal(b, &conf)
			if err != nil {
				errDetails := fmt.Sprintf("Field span_config cannot be unmarshaled to SpanConfigWithTraceModel, err: %v", err.Error())
				o11y.Error(ctx, errDetails)
				return rest.NewHTTPError(ctx, http.StatusInternalServerError,
					derrors.DataModel_InternalError_UnmarshalDataFailed).WithErrorDetails(errDetails)
			}
			reqModels[i].SpanConfig = conf
		case interfaces.SOURCE_TYPE_DATA_CONNECTION:
			conf := interfaces.SpanConfigWithDataConnection{}
			err = sonic.Unmarshal(b, &conf)
			if err != nil {
				errDetails := fmt.Sprintf("Field span_config cannot be unmarshaled to SpanConfigWithDataConnection, err: %v", err.Error())
				o11y.Error(ctx, errDetails)
				return rest.NewHTTPError(ctx, http.StatusInternalServerError,
					derrors.DataModel_InternalError_UnmarshalDataFailed).WithErrorDetails(errDetails)
			}
			reqModels[i].SpanConfig = conf
		default:
			errDetails := "span_source_type is invalid, valid span_source_type is " + interfaces.SOURCE_TYPE_DATA_VIEW +
				" or " + interfaces.SOURCE_TYPE_DATA_CONNECTION
			o11y.Error(ctx, errDetails)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidParameter_SpanSourceType).
				WithErrorDetails(errDetails)
		}
	}

	return nil
}

// 根据sourceType转换relatedLogConfig的动态类型和动态值
func convertRelatedLogConfig(ctx context.Context, reqModels []interfaces.TraceModel) error {
	for i := range reqModels {
		if reqModels[i].EnabledRelatedLog == interfaces.RELATED_LOG_CLOSE {
			continue
		} else if reqModels[i].EnabledRelatedLog == interfaces.RELATED_LOG_OPEN {
			b, err := sonic.Marshal(reqModels[i].RelatedLogConfig)
			if err != nil {
				errDetails := fmt.Sprintf("Marshal filed related_log_config failed, err: %v", err.Error())
				o11y.Error(ctx, errDetails)
				return rest.NewHTTPError(ctx, http.StatusInternalServerError,
					derrors.DataModel_InternalError_MarshalDataFailed).WithErrorDetails(errDetails)
			}

			switch reqModels[i].RelatedLogSourceType {
			case interfaces.SOURCE_TYPE_DATA_VIEW:
				conf := interfaces.RelatedLogConfigWithDataView{}
				err = sonic.Unmarshal(b, &conf)
				if err != nil {
					errDetails := fmt.Sprintf("Field related_log_config cannot be unmarshaled to RelatedLogConfigWithTraceModel, err: %v", err.Error())
					o11y.Error(ctx, errDetails)
					return rest.NewHTTPError(ctx, http.StatusInternalServerError,
						derrors.DataModel_InternalError_UnmarshalDataFailed).WithErrorDetails(errDetails)
				}
				reqModels[i].RelatedLogConfig = conf
			default:
				errDetails := "related_log_source_type is invalid, valid related_log_source_type is " + interfaces.SOURCE_TYPE_DATA_VIEW
				o11y.Error(ctx, errDetails)
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidParameter_RelatedLogSourceType).
					WithErrorDetails(errDetails)
			}
		} else {
			errDetails := fmt.Sprintf("enabled_related_log is invalid, valid enabled_related_log is %d or %d", interfaces.RELATED_LOG_CLOSE, interfaces.RELATED_LOG_OPEN)
			o11y.Error(ctx, errDetails)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidParameter_EnabledRelatedLog).
				WithErrorDetails(errDetails)
		}
	}

	return nil
}
