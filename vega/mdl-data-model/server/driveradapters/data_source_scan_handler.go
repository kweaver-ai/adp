// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

// import (
// 	"context"
// 	"fmt"
// 	"net/http"

// 	"github.com/kweaver-ai/kweaver-go-lib/logger"
// 	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
// 	"github.com/kweaver-ai/kweaver-go-lib/rest"
// 	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
// 	"github.com/gin-gonic/gin"
// 	"go.opentelemetry.io/otel/trace"

// 	"data-model/interfaces"
// )

// // 扫描数据源
// func (r *restHandler) ScanDataSource(c *gin.Context) {
// 	logger.Debug("ScanDataSource Start")
// 	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
// 		"driver layer: Scan data source", trace.WithSpanKind(trace.SpanKindServer))
// 	defer span.End()

// 	visitor, err := r.verifyOAuth(ctx, c)
// 	if err != nil {
// 		return
// 	}
// 	accountInfo := interfaces.AccountInfo{
// 		ID:   visitor.ID,
// 		Type: string(visitor.Type),
// 	}
// 	// accountID 存入 context 中
// 	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

// 	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))
// 	o11y.Info(ctx, fmt.Sprintf("driver layer: Scan data sources request parameters: [%s]", c.Request.RequestURI))

// 	var req interfaces.ScanTask
// 	if err = c.ShouldBindJSON(&req); err != nil {
// 		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
// 			WithErrorDetails("Binding paramter failed:" + err.Error())

// 		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

// 		o11y.AddHttpAttrs4HttpError(span, httpErr)
// 		rest.ReplyError(c, httpErr)
// 		return

// 	}

// 	scanRes, err := r.dss.Scan(ctx, &req)
// 	if err != nil {
// 		httpErr := err.(*rest.HTTPError)
// 		o11y.Error(ctx, fmt.Sprintf("driver layer: Scan data sources error: [%s]", httpErr.Error()))
// 		o11y.AddHttpAttrs4HttpError(span, httpErr)
// 		rest.ReplyError(c, httpErr)
// 		return
// 	}

// 	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
// 	rest.ReplyOK(c, http.StatusOK, scanRes)
// }

// // 列出所有数据源
// func (r *restHandler) ListDataSourcesWithScanRecord(c *gin.Context) {
// 	logger.Debug("ListDataSourcesWithScanRecord Start")
// 	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
// 		"driver layer: List data sources with scan record", trace.WithSpanKind(trace.SpanKindServer))
// 	defer span.End()

// 	visitor, err := r.verifyOAuth(ctx, c)
// 	if err != nil {
// 		return
// 	}
// 	accountInfo := interfaces.AccountInfo{
// 		ID:   visitor.ID,
// 		Type: string(visitor.Type),
// 	}
// 	// accountID 存入 context 中
// 	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

// 	queryType := c.Query("query_type")
// 	queryParams := &interfaces.ListDataSourceQueryParams{
// 		QueryType: queryType,
// 	}

// 	dataSources, err := r.dss.ListDataSourcesWithScanRecord(ctx, queryParams)
// 	if err != nil {
// 		httpErr := err.(*rest.HTTPError)
// 		o11y.Error(ctx, fmt.Sprintf("driver layer: List data sources error: [%s]", httpErr.Error()))
// 		o11y.AddHttpAttrs4HttpError(span, httpErr)
// 		rest.ReplyError(c, httpErr)
// 		return
// 	}

// 	logger.Debug("Handler ListDataSourcesWithScanRecord Success")
// 	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
// 	rest.ReplyOK(c, http.StatusOK, dataSources)

// }
