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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
)

// 基于指标模型的指标数据预览(内部)
func (r *restHandler) SimulateByIn(c *gin.Context) {
	logger.Debug("Handler SimulateByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor

	visitor := GenerateVisitor(c)
	r.Simulate(c, visitor)
}

// 基于指标模型的指标数据预览（外部）
func (r *restHandler) SimulateByEx(c *gin.Context) {
	logger.Debug("Handler SimulateByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"基于指标模型的指标数据预览 API", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.Simulate(c, visitor)
}

// 基于指标模型的指标数据预览，默认查询最近半小时的数据，步长5分钟
func (r *restHandler) Simulate(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler MetricModel Start")
	startTime := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "基于指标模型的指标数据预览 API", trace.WithSpanKind(trace.SpanKindServer))
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
		attribute.Key("module_name").String(METRIC_MODEL_MODULE),
	}
	defer func() {
		r.reqDurHistogram.Record(ctx, time.Since(startTime).Milliseconds(), metric.WithAttributes(attrs...))
		r.reqCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}()

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("指标数据预览请求参数: [%s,%v]", c.Request.RequestURI, c.Request.Body))

	//接收绑定参数
	query := interfaces.MetricModelQuery{}
	err := c.ShouldBindJSON(&query)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		rest.ReplyError(c, httpErr)

		attrs = append(attrs, attribute.Int("status_code", httpErr.HTTPCode))
		return
	}

	// 处理对data_view_id的兼容
	if query.DataSource == nil {
		// 数据源为空，尝试读取data_view_id
		if query.DataViewID != "" {
			// 视图id不为空，则把data_view_id转成数据源
			query.DataSource = &interfaces.MetricDataSource{
				Type: interfaces.DSL,
				ID:   query.DataViewID,
			}
		}
	}

	// ignoring_hcts 是否忽略高基序列的查询
	ignoringHCTSStr := c.DefaultQuery("ignoring_hcts", fmt.Sprintf("%v", r.appSetting.ServerSetting.IgnoringHcts))
	ignoringHCTS, err := strconv.ParseBool(ignoringHCTSStr)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_IgnoringHCTS).
			WithErrorDetails(fmt.Sprintf("The ignoring_hcts:%v is invalid", ignoringHCTSStr))
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		rest.ReplyError(c, httpErr)

		attrs = append(attrs, attribute.Int("status_code", httpErr.HTTPCode))
		return
	}
	query.IgnoringHCTS = ignoringHCTS

	// include_model 是否优先匹配持久化任务查询持久化缓存下来的数据
	includeModel := c.DefaultQuery("include_model", interfaces.DEFAULT_INCLUDE_MODEL)
	incModel, err := strconv.ParseBool(includeModel)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_IncludeModel).
			WithErrorDetails(fmt.Sprintf("The include_model:%s is invalid", includeModel))
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		rest.ReplyError(c, httpErr)

		attrs = append(attrs, attribute.Int("status_code", httpErr.HTTPCode))
		return
	}
	query.IncludeModel = incModel

	// 校验 methodOverride
	err = ValidateHeaderMethodOverride(ctx, c.GetHeader(X_HTTP_METHOD_OVERRIDE))
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

	// 校验参数
	err = ValidateMetricModelSimulate(ctx, &query)
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

	// 预览不支持分页。
	query.Limit = -1
	// 预览的model，不返回，include_model 为false
	// 请求
	result, err := r.mmService.Simulate(ctx, query)
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

	result.OverallMs = time.Now().UnixMilli() - startTime.UnixMilli()
	rest.ReplyOK(c, http.StatusOK, result)
	attrs = append(attrs, attribute.Int("status_code", http.StatusOK))
}

// 基于指标模型的指标数据预览(内部)
func (r *restHandler) GetMetricModelDataByIn(c *gin.Context) {
	logger.Debug("Handler GetMetricModelDataByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor

	visitor := GenerateVisitor(c)
	r.GetMetricModelData(c, visitor)
}

// 基于指标模型的指标数据预览（外部）
func (r *restHandler) GetMetricModelDataByEx(c *gin.Context) {
	logger.Debug("Handler GetMetricModelDataByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "获取指标模型[%s]的指标数据API",
		trace.WithSpanKind(trace.SpanKindServer))

	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetMetricModelData(c, visitor)
}

// 基于指标模型的指标数据查询
func (r *restHandler) GetMetricModelData(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler GetMetricModelData Start")
	startTime := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "获取指标模型[%s]的指标数据API", trace.WithSpanKind(trace.SpanKindServer))
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
		attribute.Key("module_name").String(METRIC_MODEL_MODULE),
	}
	defer func() {
		r.reqDurHistogram.Record(ctx, time.Since(startTime).Milliseconds(), metric.WithAttributes(attrs...))
		r.reqCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}()

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("指标数据查询请求参数: [%s,%v]", c.Request.RequestURI, c.Request.Body))

	modelIDStrs := c.Param("model_ids")
	//解析字符串 转换为 []string
	modelIDs := convert.StringToStringSlice(modelIDStrs)

	// ignoring_store_cache 是否优先匹配持久化任务查询持久化缓存下来的数据
	ignoringStoreCache := c.DefaultQuery("ignoring_store_cache", interfaces.DEFAULT_IGNORING_STORE_CACHE)
	// ignoring_memory_cache 是否优先匹配持久化任务查询持久化缓存下来的数据
	ignoringMemoryCache := c.DefaultQuery("ignoring_memory_cache", interfaces.DEFAULT_IGNORING_MEMORY_CACHE)
	// ignoring_hcts 是否忽略高基序列的查询
	ignoringHCTSStr := c.DefaultQuery("ignoring_hcts", fmt.Sprintf("%v", r.appSetting.ServerSetting.IgnoringHcts))
	// fill_null 范围查询时,对于缺失的步长点是否补空
	fillNullStr := c.DefaultQuery("fill_null", interfaces.DEFAULT_FILL_NULL)

	// limit 每页最多可返回的序列数； 分页可选1-1000； -1表示不分页； 默认值为 -1
	// limit := c.DefaultQuery("limit", interfaces.DEFAULT_SERIES_LIMIT)
	// offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	// 排序分页不支持,参数传递不认,按默认不分页的方式走
	limit := interfaces.DEFAULT_SERIES_LIMIT
	offset := interfaces.DEFAULT_OFFEST

	// include_model 是否优先匹配持久化任务查询持久化缓存下来的数据
	includeModel := c.DefaultQuery("include_model", interfaces.DEFAULT_INCLUDE_MODEL)
	// 过滤模式选择默认是normal
	filterMode := c.DefaultQuery("filter_mode", interfaces.FILTER_MODE_NORMAL)

	// 校验查询参数
	queryParam, err := validateMetricModelQueryParameters(ctx, offset, limit,
		ignoringMemoryCache, ignoringStoreCache, ignoringHCTSStr, fillNullStr, includeModel)
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

	err = ValidateHeaderMethodOverride(ctx, c.GetHeader(X_HTTP_METHOD_OVERRIDE))
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
	// instant query 参数， time（即start 和 end），isInstantQuery, interval = 1
	//接收绑定参数
	if len(modelIDs) == 1 {
		query := interfaces.MetricModelQuery{}
		err = c.ShouldBindJSON(&query)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("Binding Paramter Failed:%s", err.Error()))

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))

			rest.ReplyError(c, httpErr)

			attrs = append(attrs, attribute.Int("status_code", httpErr.HTTPCode))
			return
		}

		query.MetricModelQueryParameters = queryParam
		query.MetricModelID = modelIDs[0]
		query.FilterMode = filterMode

		err = validateMetricModelData(ctx, &query)
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
		result, seriesTotal, pointTotal, err := r.mmService.Exec(ctx, &query)
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
		span.SetAttributes(attribute.Key(SERIES_TOTAL).Int(seriesTotal),
			attribute.Key(POINT_TOTAL).Int(pointTotal))

		result.OverallMs = time.Now().UnixMilli() - startTime.UnixMilli()
		rest.ReplyOK(c, http.StatusOK, result)

		attrs = append(attrs, attribute.Int("status_code", http.StatusOK))
	} else {
		query := []interfaces.MetricModelQuery{}
		err = c.ShouldBindJSON(&query)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter).
				WithErrorDetails("Binding Paramter Failed:" + err.Error())

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))

			rest.ReplyError(c, httpErr)

			attrs = append(attrs, attribute.Int("status_code", httpErr.HTTPCode))
			return
		}

		res := make([]interface{}, len(query))
		failed := false
		pointTotal := 0
		seriesTotal := 0

		for i := range query {
			startTimei := time.Now()
			query[i].MetricModelQueryParameters = queryParam
			query[i].MetricModelID = modelIDs[i]

			// 参数校验
			err = validateMetricModelData(ctx, &query[i])
			if err != nil {
				httpErr := err.(*rest.HTTPError)
				// 设置 trace 的错误信息的 attributes
				o11y.AddHttpAttrs4HttpError(span, httpErr)
				o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
					httpErr.BaseError.ErrorDetails))

				// 收集validateMetricModelData的错误
				res[i] = interfaces.UniResponseError{
					StatusCode: httpErr.HTTPCode,
					BaseError:  httpErr.BaseError,
				}
				failed = true
				continue
			}

			// 执行查询
			result, series, point, err := r.mmService.Exec(ctx, &query[i])
			if err != nil {
				httpErr := err.(*rest.HTTPError)
				// 设置 trace 的错误信息的 attributes
				o11y.AddHttpAttrs4HttpError(span, httpErr)
				o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
					httpErr.BaseError.ErrorDetails))

				// 收集Exec的错误
				res[i] = interfaces.UniResponseError{
					StatusCode: httpErr.HTTPCode,
					BaseError:  httpErr.BaseError,
				}
				failed = true
			} else {
				seriesTotal += series
				pointTotal += point
				result.OverallMs = time.Now().UnixMilli() - startTimei.UnixMilli()

				res[i] = result
				// 设置 trace 的成功信息的 attributes
				o11y.AddHttpAttrs4Ok(span, http.StatusOK)
				span.SetAttributes(attribute.Key(SERIES_TOTAL).Int(seriesTotal),
					attribute.Key(POINT_TOTAL).Int(pointTotal))
			}

		}

		if failed {
			rest.ReplyOK(c, http.StatusMultiStatus, res)

			attrs = append(attrs, attribute.Int("status_code", http.StatusMultiStatus))
		} else {
			rest.ReplyOK(c, http.StatusOK, res)

			attrs = append(attrs, attribute.Int("status_code", http.StatusOK))
		}
	}
}

func (r *restHandler) GetMetricModelFields(c *gin.Context) {
	logger.Debug("Handler GetMetricModelFields Start")
	startTime := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "基于指标模型ID获取模型字段列表 API", trace.WithSpanKind(trace.SpanKindServer))
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

	attrs := []attribute.KeyValue{
		attribute.Key("handler").String(c.FullPath()),
		attribute.Key("method").String(c.Request.Method),
		attribute.Key("service_name").String(SERVICE_NAME),
		attribute.Key("module_name").String(METRIC_MODEL_MODULE),
	}
	defer func() {
		r.reqDurHistogram.Record(ctx, time.Since(startTime).Milliseconds(), metric.WithAttributes(attrs...))
		r.reqCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}()

	//获取参数字符串,单个模型id
	modelID := c.Param("model_ids")
	span.SetAttributes(attribute.Key("model_id").String(modelID))

	if strings.Trim(modelID, " ") == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter).
			WithErrorDetails("metric model id is empty")

		rest.ReplyError(c, httpErr)
		return
	}

	// 请求
	result, err := r.mmService.GetMetricModelFields(ctx, modelID)
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

func (r *restHandler) GetMetricModelFieldValues(c *gin.Context) {
	logger.Debug("Handler GetMetricModelFieldValues Start")
	startTime := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "基于指标模型ID获取模型字段值列表 API", trace.WithSpanKind(trace.SpanKindServer))
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

	attrs := []attribute.KeyValue{
		attribute.Key("handler").String(c.FullPath()),
		attribute.Key("method").String(c.Request.Method),
		attribute.Key("service_name").String(SERVICE_NAME),
		attribute.Key("module_name").String(METRIC_MODEL_MODULE),
	}
	defer func() {
		r.reqDurHistogram.Record(ctx, time.Since(startTime).Milliseconds(), metric.WithAttributes(attrs...))
		r.reqCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}()

	//获取参数字符串,单个模型id
	modelID := c.Param("model_ids")
	span.SetAttributes(attribute.Key("model_id").String(modelID))
	fieldName := c.Param("field_name")
	span.SetAttributes(attribute.Key("fieldName").String(fieldName))

	if strings.Trim(modelID, " ") == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter).
			WithErrorDetails("metric model id is empty")

		rest.ReplyError(c, httpErr)
		return
	}

	if fieldName == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_NullParameter_FieldName)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		rest.ReplyError(c, httpErr)

		attrs = append(attrs, attribute.Int("status_code", httpErr.HTTPCode))
		return
	}

	// 请求
	result, err := r.mmService.GetMetricModelFieldValues(ctx, modelID, fieldName)
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

func (r *restHandler) GetMetricModelLabels(c *gin.Context) {
	logger.Debug("Handler GetMetricModelLabels Start")
	startTime := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "基于指标模型ID获取模型的维度字段列表 API", trace.WithSpanKind(trace.SpanKindServer))
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

	attrs := []attribute.KeyValue{
		attribute.Key("handler").String(c.FullPath()),
		attribute.Key("method").String(c.Request.Method),
		attribute.Key("service_name").String(SERVICE_NAME),
		attribute.Key("module_name").String(METRIC_MODEL_MODULE),
	}
	defer func() {
		r.reqDurHistogram.Record(ctx, time.Since(startTime).Milliseconds(), metric.WithAttributes(attrs...))
		r.reqCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}()

	//获取参数字符串,单个模型id
	modelID := c.Param("model_ids")
	span.SetAttributes(attribute.Key("model_id").String(modelID))

	if strings.Trim(modelID, " ") == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter).
			WithErrorDetails("metric model id is empty")

		rest.ReplyError(c, httpErr)
		return
	}

	// 请求
	result, err := r.mmService.GetMetricModelLabels(ctx, modelID)
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
