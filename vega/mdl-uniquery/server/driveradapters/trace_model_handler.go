// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"

	uerrors "uniquery/errors"
	"uniquery/interfaces"
)

const X_REQUEST_TOOK string = "X-Request-Took"

func (r *restHandler) PreviewSpanListByEx(c *gin.Context) {
	logger.Debug("Handler PreviewSpanListByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 预览span列表", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.PreviewSpanList(c, visitor)
}

func (r *restHandler) PreviewSpanListByIn(c *gin.Context) {
	logger.Debug("Handler PreviewSpanListByIn Start")

	visitor := GenerateVisitor(c)
	r.PreviewSpanList(c, visitor)
}

// 预览span列表
func (r *restHandler) PreviewSpanList(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler PreviewSpanList Start")
	start := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver层: 预览span列表", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler PreviewSpanList End")
	}()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. check重载请求头
	method := c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE)
	if method != http.MethodGet {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_OverrideMethod)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 2. 获取url上的trace_id
	traceID := c.Param("trace_id")

	// 3. 初始化queryParams
	queryParams := interfaces.SpanListPreviewParams{
		SpanListQueryParams: interfaces.SpanListQueryParams{
			TraceID: traceID,
		},
	}

	// 4. 接收request body
	err := c.ShouldBindJSON(&queryParams)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_RequestBody).
			WithErrorDetails("Binding paramter failed: " + err.Error())
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 5. 校验span列表查询参数
	queryParams.SpanListQueryParams, err = validateParamsWhenGetSpanList(ctx, queryParams.SpanListQueryParams)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, err, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 6. 校验传入的链路模型是否有效
	// modelID=0, 说明是创建链路模型时预览, 否则为修改链路模型时预览
	var model interfaces.TraceModel
	if queryParams.TraceModel.ID == "" {
		model, err = r.tmService.SimulateCreateTraceModel(ctx, queryParams.TraceModel)
	} else {
		model, err = r.tmService.SimulateUpdateTraceModel(ctx, queryParams.TraceModel.ID, queryParams.TraceModel)
	}

	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, err, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 7. 调用logic层, 查询span列表
	spanList, total, err := r.tmService.GetSpanList(ctx, model, queryParams.SpanListQueryParams)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, err, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 8. 构造返回结果
	result := map[string]interface{}{
		"entries":     spanList,
		"total_count": total,
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOkWithHeaders(c, http.StatusOK, result, map[string]string{
		X_REQUEST_TOOK: time.Since(start).String(),
	})
}

func (r *restHandler) GetSpanListByEx(c *gin.Context) {
	logger.Debug("Handler GetSpanListByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 查询span列表", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetSpanList(c, visitor)
}

func (r *restHandler) GetSpanListByIn(c *gin.Context) {
	logger.Debug("Handler GetSpanListByIn Start")

	visitor := GenerateVisitor(c)
	r.GetSpanList(c, visitor)
}

// 查询span列表
func (r *restHandler) GetSpanList(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler GetSpanList Start")
	start := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver层: 查询span列表", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler GetSpanList End")
	}()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. check重载请求头
	method := c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE)
	if method != http.MethodGet {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_OverrideMethod)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 2. 获取url上的trace_model_id与trace_id
	modelID := c.Param("trace_model_id")
	traceID := c.Param("trace_id")

	// 3. 校验trace_model_id
	if modelID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_TraceModel_InvalidParameter_ModelID).
			WithErrorDetails("No invalid trace model id was passed in")
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 4. 初始化queryParams
	queryParams := interfaces.SpanListQueryParams{
		TraceID: traceID,
	}

	// 5. 接收request body
	err := c.ShouldBindJSON(&queryParams)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_RequestBody).
			WithErrorDetails("Binding paramter failed: " + err.Error())
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 6. 校验span列表查询参数
	queryParams, err = validateParamsWhenGetSpanList(ctx, queryParams)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 7. 根据modelID查询链路模型对象
	model, err := r.tmService.GetTraceModelByID(ctx, modelID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 8. 调用logic层, 查询span列表
	spanList, total, err := r.tmService.GetSpanList(ctx, model, queryParams)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 9. 构造返回结果
	result := map[string]interface{}{
		"entries":     spanList,
		"total_count": total,
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOkWithHeaders(c, http.StatusOK, result, map[string]string{
		X_REQUEST_TOOK: time.Since(start).String(),
	})
}

func (r *restHandler) PreviewTraceByEx(c *gin.Context) {
	logger.Debug("Handler PreviewTraceByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 预览trace详情", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.PreviewTrace(c, visitor)
}

func (r *restHandler) PreviewTraceByIn(c *gin.Context) {
	logger.Debug("Handler PreviewTraceByIn Start")

	visitor := GenerateVisitor(c)
	r.PreviewTrace(c, visitor)
}

// 预览单条trace
func (r *restHandler) PreviewTrace(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler PreviewTrace Start")
	start := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver层: 预览trace详情", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler PreviewTrace End")
	}()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. check重载请求头
	method := c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE)
	if method != http.MethodGet {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_OverrideMethod)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 2. 获取url上的trace_id
	traceID := c.Param("trace_id")

	// 3. 接收request body
	queryParams := interfaces.TracePreviewParams{
		TraceQueryParams: interfaces.TraceQueryParams{
			TraceID: traceID,
		},
	}
	err := c.ShouldBindJSON(&queryParams)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_RequestBody).
			WithErrorDetails("Binding paramter failed: " + err.Error())
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 4. 校验传入的链路模型是否有效
	// modelID=0, 说明是创建链路模型时预览, 否则为修改链路模型时预览
	var model interfaces.TraceModel
	if queryParams.TraceModel.ID == "" {
		model, err = r.tmService.SimulateCreateTraceModel(ctx, queryParams.TraceModel)
	} else {
		model, err = r.tmService.SimulateUpdateTraceModel(ctx, queryParams.TraceModel.ID, queryParams.TraceModel)
	}

	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, err, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 5. 调用logic层, 查询trace详情
	traceDetail, err := r.tmService.GetTrace(ctx, model, queryParams.TraceQueryParams)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, err, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOkWithHeaders(c, http.StatusOK, traceDetail, map[string]string{
		X_REQUEST_TOOK: time.Since(start).String(),
	})
}

func (r *restHandler) GetTraceByEx(c *gin.Context) {
	logger.Debug("Handler GetTraceByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 查询trace详情", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetTrace(c, visitor)
}

func (r *restHandler) GetTraceByIn(c *gin.Context) {
	logger.Debug("Handler GetTraceByIn Start")

	visitor := GenerateVisitor(c)
	r.GetTrace(c, visitor)
}

// 查询单条trace
func (r *restHandler) GetTrace(c *gin.Context, visitor rest.Visitor) {
	// start1 := time.Now()
	// fmt.Printf("[driver]开始查询Trace详情, 当前时间%v\n", start1)
	// defer func() {
	// 	end1 := time.Now()
	// 	fmt.Printf("[driver]结束查询Trace详情, 当前时间%v, 共耗时%v\n", end1, end1.Sub(start1))
	// 	fmt.Println("------------------\n------------------")
	// }()

	logger.Debug("Handler GetTrace Start")
	start := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver层: 查询trace详情", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler GetTrace End")
	}()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. check重载请求头
	method := c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE)
	if method != http.MethodGet {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_OverrideMethod)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 2. 获取url上的trace_model_id与trace_id
	modelID := c.Param("trace_model_id")
	traceID := c.Param("trace_id")

	// 3. 校验trace_model_id
	if modelID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_TraceModel_InvalidParameter_ModelID).
			WithErrorDetails("No invalid trace model id was passed in")
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 4. 初始化queryParams
	queryParams := interfaces.TraceQueryParams{
		TraceID: traceID,
	}

	// 5. 接收request body
	err := c.ShouldBindJSON(&queryParams)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_RequestBody).
			WithErrorDetails("Binding paramter failed: " + err.Error())
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 6. 根据modelID查询链路模型对象
	model, err := r.tmService.GetTraceModelByID(ctx, modelID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 7. 调用logic层, 查询trace详情
	traceDetail, err := r.tmService.GetTrace(ctx, model, queryParams)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// end1 := time.Now()
	// fmt.Printf("[driver]结束查询Trace详情, 当前时间%v, 共耗时%v\n", end1, end1.Sub(start1))
	// fmt.Println("------------------\n------------------")

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOkWithHeaders(c, http.StatusOK, traceDetail, map[string]string{
		X_REQUEST_TOOK: time.Since(start).String(),
	})
}

func (r *restHandler) PreviewSpanByEx(c *gin.Context) {
	logger.Debug("Handler PreviewSpanByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 预览span详情", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.PreviewSpan(c, visitor)
}

func (r *restHandler) PreviewSpanByIn(c *gin.Context) {
	logger.Debug("Handler PreviewSpanByIn Start")

	visitor := GenerateVisitor(c)
	r.PreviewSpan(c, visitor)
}

// 预览单条span
func (r *restHandler) PreviewSpan(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler PreviewSpan Start")
	start := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver层: 预览span详情", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler PreviewSpan End")
	}()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. check重载请求头
	method := c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE)
	if method != http.MethodGet {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_OverrideMethod)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 2. 获取url上的trace_id与span_id
	traceID := c.Param("trace_id")
	spanID := c.Param("span_id")

	// 3. 接收request body
	queryParams := interfaces.SpanPreviewParams{}
	err := c.ShouldBindJSON(&queryParams)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_RequestBody).
			WithErrorDetails("Binding paramter failed: " + err.Error())
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 4. 校验传入的链路模型是否有效
	// modelID=0, 说明是创建链路模型时预览, 否则为修改链路模型时预览
	var model interfaces.TraceModel
	if queryParams.TraceModel.ID == "" {
		model, err = r.tmService.SimulateCreateTraceModel(ctx, queryParams.TraceModel)
	} else {
		model, err = r.tmService.SimulateUpdateTraceModel(ctx, queryParams.TraceModel.ID, queryParams.TraceModel)
	}

	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, err, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 5. 调用logic层, 查询span详情
	spanQueryParams := interfaces.SpanQueryParams{
		TraceID: traceID,
		SpanID:  spanID,
	}
	spanDetail, err := r.tmService.GetSpan(ctx, model, spanQueryParams)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, err, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOkWithHeaders(c, http.StatusOK, spanDetail, map[string]string{
		X_REQUEST_TOOK: time.Since(start).String(),
	})
}

func (r *restHandler) GetSpanByEx(c *gin.Context) {
	logger.Debug("Handler GetSpanByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 查询span详情", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetSpan(c, visitor)
}

func (r *restHandler) GetSpanByIn(c *gin.Context) {
	logger.Debug("Handler GetSpanByIn Start")

	visitor := GenerateVisitor(c)
	r.GetSpan(c, visitor)
}

// 查询单条span
func (r *restHandler) GetSpan(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler GetSpan Start")
	start := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver层: 查询span详情", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler GetSpan End")
	}()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. 获取url上的trace_model_id与trace_id
	modelID := c.Param("trace_model_id")
	traceID := c.Param("trace_id")
	spanID := c.Param("span_id")

	// 2. 校验trace_model_id
	if modelID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_TraceModel_InvalidParameter_ModelID).
			WithErrorDetails("No invalid trace model id was passed in")
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 3. 根据modelID查询链路模型对象
	model, err := r.tmService.GetTraceModelByID(ctx, modelID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, err, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 4. 调用logic层, 查询span详情
	spanQueryParams := interfaces.SpanQueryParams{
		TraceID: traceID,
		SpanID:  spanID,
	}
	spanDetail, err := r.tmService.GetSpan(ctx, model, spanQueryParams)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, err, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOkWithHeaders(c, http.StatusOK, spanDetail, map[string]string{
		X_REQUEST_TOOK: time.Since(start).String(),
	})
}

func (r *restHandler) PreviewSpanRelatedLogListByEx(c *gin.Context) {
	logger.Debug("Handler PreviewSpanRelatedLogListByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 预览span关联日志列表", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.PreviewSpanRelatedLogList(c, visitor)
}

func (r *restHandler) PreviewSpanRelatedLogListByIn(c *gin.Context) {
	logger.Debug("Handler PreviewSpanRelatedLogListByIn Start")

	visitor := GenerateVisitor(c)
	r.PreviewSpanRelatedLogList(c, visitor)
}

// 预览span关联日志列表
func (r *restHandler) PreviewSpanRelatedLogList(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler PreviewSpanRelatedLogList Start")
	start := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver层: 预览span关联日志列表", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler PreviewSpanRelatedLogList End")
	}()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. check重载请求头
	method := c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE)
	if method != http.MethodGet {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_OverrideMethod)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 2. 获取url上的trace_id与span_id
	traceID := c.Param("trace_id")
	spanID := c.Param("span_id")

	// 3. 接收request body
	queryParams := interfaces.RelatedLogListPreviewParams{}
	err := c.ShouldBindJSON(&queryParams)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_RequestBody).
			WithErrorDetails("Binding paramter failed: " + err.Error())
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 4. 更新queryParams
	queryParams.TraceID = traceID
	queryParams.SpanID = spanID

	// 5. 校验关联日志列表查询参数
	queryParams.RelatedLogListQueryParams, err = validateParamsWhenGetRelatedLogList(ctx, queryParams.RelatedLogListQueryParams)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, err, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 6. 校验传入的链路模型是否有效
	// modelID=0, 说明是创建链路模型时预览, 否则为修改链路模型时预览
	var model interfaces.TraceModel
	if queryParams.TraceModel.ID == "" {
		model, err = r.tmService.SimulateCreateTraceModel(ctx, queryParams.TraceModel)
	} else {
		model, err = r.tmService.SimulateUpdateTraceModel(ctx, queryParams.TraceModel.ID, queryParams.TraceModel)
	}

	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, err, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 7. 调用logic层, 查询关联日志列表
	relatedLogList, total, err := r.tmService.GetSpanRelatedLogList(ctx, model, queryParams.RelatedLogListQueryParams)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, err, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 8. 构造返回结果
	result := map[string]interface{}{
		"entries":     relatedLogList,
		"total_count": total,
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOkWithHeaders(c, http.StatusOK, result, map[string]string{
		X_REQUEST_TOOK: time.Since(start).String(),
	})
}

func (r *restHandler) GetSpanRelatedLogListByEx(c *gin.Context) {
	logger.Debug("Handler GetSpanRelatedLogListByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 查询span关联日志列表", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetSpanRelatedLogList(c, visitor)
}

func (r *restHandler) GetSpanRelatedLogListByIn(c *gin.Context) {
	logger.Debug("Handler GetSpanRelatedLogListByIn Start")

	visitor := GenerateVisitor(c)
	r.GetSpanRelatedLogList(c, visitor)
}

// 查询span关联日志列表
func (r *restHandler) GetSpanRelatedLogList(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler GetSpanRelatedLogList Start")
	start := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver层: 查询span关联日志列表", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler GetSpanRelatedLogList End")
	}()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. check重载请求头
	method := c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE)
	if method != http.MethodGet {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_OverrideMethod)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 2. 获取url上的trace_model_id, trace_id与span_id
	modelID := c.Param("trace_model_id")
	traceID := c.Param("trace_id")
	spanID := c.Param("span_id")

	// 3. 校验trace_model_id
	if modelID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_TraceModel_InvalidParameter_ModelID).
			WithErrorDetails("No invalid trace model id was passed in")
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 4. 初始化queryParams
	queryParams := interfaces.RelatedLogListQueryParams{
		TraceID: traceID,
		SpanID:  spanID,
	}

	// 5. 接收request body
	err := c.ShouldBindJSON(&queryParams)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_RequestBody).
			WithErrorDetails("Binding paramter failed: " + err.Error())
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 3. 校验关联日志列表查询参数
	queryParams, err = validateParamsWhenGetRelatedLogList(ctx, queryParams)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}

	// 4. 更新queryParams
	queryParams.TraceID = traceID
	queryParams.SpanID = spanID

	// 5. 根据modelID查询链路模型对象
	model, err := r.tmService.GetTraceModelByID(ctx, modelID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, err, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 6. 调用logic层, 查询关联日志列表
	relatedLogList, total, err := r.tmService.GetSpanRelatedLogList(ctx, model, queryParams)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, err, map[string]string{
			X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	// 7. 构造返回结果
	result := map[string]interface{}{
		"entries":     relatedLogList,
		"total_count": total,
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOkWithHeaders(c, http.StatusOK, result, map[string]string{
		X_REQUEST_TOOK: time.Since(start).String(),
	})
}
