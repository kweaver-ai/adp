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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	uerrors "uniquery/errors"
	"uniquery/interfaces"
)

// 基于目标模型的指标数据预览(内部)
func (r *restHandler) ObjectiveSimulateByIn(c *gin.Context) {
	logger.Debug("Handler ObjectiveSimulateByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor

	visitor := GenerateVisitor(c)
	r.ObjectiveSimulate(c, visitor)
}

// 基于目标模型的指标数据预览（外部）
func (r *restHandler) ObjectiveSimulateByEx(c *gin.Context) {
	logger.Debug("Handler ObjectiveSimulateByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"目标模型数据预览 API", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ObjectiveSimulate(c, visitor)
}

// 基于目标模型的指标数据预览，默认查询最近半小时的数据，步长5分钟
func (r *restHandler) ObjectiveSimulate(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler ObjectiveModel Start")
	startTime := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "基于目标模型的指标数据预览 API", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	attrs := []attribute.KeyValue{
		attribute.Key("handler").String(c.FullPath()),
		attribute.Key("method").String(c.Request.Method),
		attribute.Key("method_override").String(c.GetHeader(X_HTTP_METHOD_OVERRIDE)),
		attribute.Key("service_name").String(SERVICE_NAME),
		attribute.Key("module_name").String(OBJECTIVE_MODEL_MODULE),
	}
	defer func() {
		r.reqDurHistogram.Record(ctx, time.Since(startTime).Milliseconds(), metric.WithAttributes(attrs...))
		r.reqCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}()

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("目标模型的指标数据预览请求参数: [%s,%v]", c.Request.RequestURI, c.Request.Body))

	//接收绑定参数
	query := interfaces.ObjectiveModelQuery{}
	err := c.ShouldBindJSON(&query)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		rest.ReplyError(c, httpErr)

		attrs = append(attrs, attribute.Int("status_code", httpErr.HTTPCode))
		return
	}

	// 校验 methodOverride
	err = ValidateHeaderMethodOverride(ctx, c.GetHeader(X_HTTP_METHOD_OVERRIDE))
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)

		attrs = append(attrs, attribute.Int("status_code", httpErr.HTTPCode))
		return
	}

	// 校验参数
	err = ValidateObjectiveModelSimulate(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		rest.ReplyError(c, httpErr)

		attrs = append(attrs, attribute.Int("status_code", httpErr.HTTPCode))
		return
	}

	// 预览的model，不返回，include_model 为false
	// 请求
	result, err := r.omService.Simulate(ctx, query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		rest.ReplyError(c, httpErr)

		attrs = append(attrs, attribute.Int("status_code", httpErr.HTTPCode))
		return
	}
	// 设置 trace 的成功信息的 attributes
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)

	rest.ReplyOK(c, http.StatusOK, result)
	attrs = append(attrs, attribute.Int("status_code", http.StatusOK))
}

// 基于目标模型的指标数据查询(内部)
func (r *restHandler) GetObjectiveModelDataByIn(c *gin.Context) {
	logger.Debug("Handler GetObjectiveModelDataByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor

	visitor := GenerateVisitor(c)
	r.GetObjectiveModelData(c, visitor)
}

// 基于目标模型的指标数据查询（外部）
func (r *restHandler) GetObjectiveModelDataByEx(c *gin.Context) {
	logger.Debug("Handler GetObjectiveModelDataByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"目标模型数据查询 API", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetObjectiveModelData(c, visitor)
}

// 基于目标模型的指标数据查询
func (r *restHandler) GetObjectiveModelData(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler GetObjectiveModelData Start")
	startTime := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "基于目标模型的指标数据查询 API", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	attrs := []attribute.KeyValue{
		attribute.Key("handler").String(c.FullPath()),
		attribute.Key("method").String(c.Request.Method),
		attribute.Key("method_override").String(c.GetHeader(X_HTTP_METHOD_OVERRIDE)),
		attribute.Key("service_name").String(SERVICE_NAME),
		attribute.Key("module_name").String(OBJECTIVE_MODEL_MODULE),
	}
	defer func() {
		r.reqDurHistogram.Record(ctx, time.Since(startTime).Milliseconds(), metric.WithAttributes(attrs...))
		r.reqCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}()

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("目标模型的指标数据查询请求参数: [%s,%v]", c.Request.RequestURI, c.Request.Body))

	// 校验 methodOverride
	err := ValidateHeaderMethodOverride(ctx, c.GetHeader(X_HTTP_METHOD_OVERRIDE))
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)

		attrs = append(attrs, attribute.Int("status_code", httpErr.HTTPCode))
		return
	}

	// instant query 参数， time（即start 和 end），isInstantQuery, interval = 1
	//接收绑定参数
	query := interfaces.ObjectiveModelQuery{}
	err = c.ShouldBindJSON(&query)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Binding Paramter Failed:%s", err.Error()))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		rest.ReplyError(c, httpErr)

		attrs = append(attrs, attribute.Int("status_code", httpErr.HTTPCode))
		return
	}

	modelID := c.Param("model_id")
	// ignoring_store_cache 是否优先匹配持久化任务查询持久化缓存下来的数据
	ignoringStoreCache := c.DefaultQuery("ignoring_store_cache", interfaces.DEFAULT_IGNORING_STORE_CACHE)
	// ignoring_memory_cache 是否优先匹配持久化任务查询持久化缓存下来的数据
	ignoringMemoryCache := c.DefaultQuery("ignoring_memory_cache", interfaces.DEFAULT_IGNORING_STORE_CACHE)
	// include_model 是否优先匹配持久化任务查询持久化缓存下来的数据
	includeModel := c.DefaultQuery("include_model", interfaces.DEFAULT_INCLUDE_MODEL)
	// include_metrics 请求需要的指标
	includeMetrics := c.QueryArray("include_metrics")

	queryParam, err := validateObjectiveModelQueryParameters(ctx, ignoringMemoryCache, ignoringStoreCache,
		includeModel, includeMetrics)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		rest.ReplyError(c, httpErr)

		attrs = append(attrs, attribute.Int("status_code", httpErr.HTTPCode))
		return
	}

	query.ObjectiveModelID = modelID
	query.ObjectiveModelQueryParameters = queryParam
	err = validateObjectiveModelData(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		rest.ReplyError(c, httpErr)

		attrs = append(attrs, attribute.Int("status_code", httpErr.HTTPCode))
		return
	}

	// 执行查询
	result, err := r.omService.Exec(ctx, query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		rest.ReplyError(c, httpErr)

		attrs = append(attrs, attribute.Int("status_code", httpErr.HTTPCode))
		return
	}

	// 设置 trace 的成功信息的 attributes
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	span.SetAttributes(attribute.Key(SERIES_TOTAL).Int(result.SeriesTotal),
		attribute.Key(POINT_TOTAL_PER_SERIES).Int(result.PointTotalPerSeries))

	rest.ReplyOK(c, http.StatusOK, result)

	attrs = append(attrs, attribute.Int("status_code", http.StatusOK))

}
