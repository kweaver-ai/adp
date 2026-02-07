// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"strconv"
	"strings"
	uerrors "vega-gateway-pro/errors"
	"vega-gateway-pro/interfaces"
)

// FetchQuery @Summary 查询数据
// @Description 查询数据
// @Tags fetch
// @Accept json
// @Produce json
// @Param req body interfaces.FetchQueryReq true "查询数据请求"
// @Success 200 {object} interfaces.FetchResp "查询数据响应"
// @Failure 400 {object} rest.HTTPError "请求参数错误"
// @Failure 500 {object} rest.HTTPError "内部服务器错误"
// @Router /fetch [post]
func (r *restHandler) FetchQuery(c *gin.Context) {
	logger.Debug("Handler FetchQuery Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver layer: FetchQuery",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()
	query := interfaces.FetchQueryReq{}
	err := c.ShouldBindJSON(&query)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.InvalidParameter_RequestBody)
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errorDetails []string
			for _, fieldError := range validationErrors {
				errorDetails = append(errorDetails, fmt.Sprintf(
					"field '%s' validate failed: rule '%s' (param: %v), current value: %v",
					fieldError.Field(),
					fieldError.ActualTag(),
					fieldError.Param(),
					fieldError.Value(),
				))
			}
			httpErr = httpErr.WithErrorDetails("Binding Parameter Failed: " + strings.Join(errorDetails, "; "))
		} else {
			httpErr = httpErr.WithErrorDetails("Binding Parameter Failed:" + err.Error())
		}
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if query.BatchSize == nil {
		defaultBatchSize := 10000
		query.BatchSize = &defaultBatchSize
	}
	if query.Timeout == nil {
		defaultTimeout := 0
		query.Timeout = &defaultTimeout
	}

	res, err := r.fetchService.FetchQuery(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, res)
}

// NextQuery @Summary 查询下一页数据
// @Description 查询下一页数据
// @Tags fetch
// @Accept json
// @Produce json
// @Param query_id path string true "查询ID"
// @Param slug path string true "查询签名"
// @Param token path string true "查询令牌"
// @Param batch_size query int false "每页数据量"
// @Success 200 {object} interfaces.FetchResp "查询下一页数据响应"
// @Failure 400 {object} rest.HTTPError "请求参数错误"
// @Failure 500 {object} rest.HTTPError "内部服务器错误"
// @Router /fetch/{query_id}/{slug}/{token} [get]
func (r *restHandler) NextQuery(c *gin.Context) {
	logger.Debug("Handler NextQuery Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver layer: NextQuery",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	nextQueryReq := interfaces.NextQueryReq{}
	err := c.ShouldBindUri(&nextQueryReq)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.InvalidParameter_RequestBody).
			WithErrorDetails("Binding Parameter Failed:" + err.Error())
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	batchSizeStr := c.Query("batch_size")
	batchSize := 10000
	if batchSizeStr != "" {
		batchSize, err = strconv.Atoi(batchSizeStr)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.InvalidParameter_RequestBody).
				WithErrorDetails("Binding Parameter Failed:" + err.Error())
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		// 限制batchSize在1-10000之间
		if batchSize <= 0 || batchSize > 10000 {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.InvalidParameter_RequestBody).
				WithErrorDetails("batch_size must be between 1 and 10000")
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}
	nextQueryReq.BatchSize = batchSize

	res, err := r.fetchService.NextQuery(ctx, &nextQueryReq)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, res)
}
