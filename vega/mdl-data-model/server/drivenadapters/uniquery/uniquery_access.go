// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package uniquery

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"data-model/common"
	"data-model/interfaces"
)

var (
	uAccessOnce sync.Once
	uAccess     interfaces.UniqueryAccess
)

type uniqueryAccess struct {
	appSetting  *common.AppSetting
	uniqueryUrl string
	httpClient  rest.HTTPClient
}

func NewUniqueryAccess(appSetting *common.AppSetting) interfaces.UniqueryAccess {
	uAccessOnce.Do(func() {
		uAccess = &uniqueryAccess{
			appSetting:  appSetting,
			uniqueryUrl: appSetting.UniQueryUrl,
			httpClient:  common.NewHTTPClient(),
		}
	})

	return uAccess
}

// 计算公式有效性检查
func (ua *uniqueryAccess) CheckFormulaByUniquery(ctx context.Context, query interfaces.MetricModelQuery) (bool, string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "计算公式有效性校验", trace.WithSpanKind(trace.SpanKindClient))

	urlStr := fmt.Sprintf("%s/metric-model", ua.uniqueryUrl)
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:            urlStr,
		HttpMethod:         http.MethodPost,
		HttpContentType:    rest.ContentTypeJson,
		HttpMethodOverride: http.MethodGet,
	})

	span.SetAttributes(
		attr.Key("metric_type").String(query.MetricType),
		// attr.Key("data_source_type").String(query.DataSource.Type),
		// attr.Key("data_source_id").String(query.DataSource.ID),
		attr.Key("query_type").String(query.QueryType),
		attr.Key("formula").String(query.Formula),
		attr.Key("measure_field").String(query.MeasureField),
		attr.Key("is_model_request").Bool(query.IsModelRequest),
	)
	defer span.End()

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	uniqueryMetricModelHeaders := map[string]string{
		interfaces.CONTENT_TYPE_NAME:           interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_METHOD_OVERRIDE: http.MethodGet,
		interfaces.HTTP_HEADER_ACCOUNT_ID:      accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE:    accountInfo.Type,
	}

	respCode, result, err := ua.httpClient.PostNoUnmarshal(ctx, urlStr, uniqueryMetricModelHeaders, query)
	logger.Debugf("post [%s] finished, response code is [%d], result is [%s], error is [%v]", urlStr, respCode, result, err)

	if err != nil {
		logger.Errorf("Post metric model simulate request failed: %v", err)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Post Failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Post metric model simulate request failed: %v", err))

		return false, "", fmt.Errorf("post metric model simulate request failed: %v", err)
	}
	if respCode != http.StatusOK {
		// 转成 baseerror
		var baseError rest.BaseError
		if err := sonic.Unmarshal(result, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)

			// 添加异常时的 trace 属性
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal BaesError failed")
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Unmalshal BaesError failed: %v", err))

			return false, "", err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode, BaseError: baseError}
		logger.Errorf("Formula invalid: %v", httpErr.Error())

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Post metric model simulate failed: %v", baseError))

		return false, fmt.Sprintf("Formula invalid: %v", httpErr.Error()), nil
	}

	// 添加成功时的 trace 属性
	o11y.AddHttpAttrs4Ok(span, respCode)

	return true, "", nil
}

// 计算公式有效性检查
func (ua *uniqueryAccess) BuildDataViewSql(ctx context.Context, view *interfaces.DataView) (string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: build data view sql", trace.WithSpanKind(trace.SpanKindClient))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         ua.uniqueryUrl,
		HttpMethod:      http.MethodPost,
		HttpContentType: rest.ContentTypeJson,
	})

	urlStr := fmt.Sprintf("%s/data-view-sql", ua.uniqueryUrl)

	span.SetAttributes(
		attr.Key("view_type").String(view.Type),
		attr.Key("query_type").String(view.QueryType),
		attr.Key("view_id").String(view.ViewID),
		attr.Key("view_name").String(view.ViewName),
	)
	defer span.End()

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	uniqueryMetricModelHeaders := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	respCode, result, err := ua.httpClient.PostNoUnmarshal(ctx, urlStr, uniqueryMetricModelHeaders, view)
	logger.Debugf("post [%s] finished, response code is [%d], result is [%s], error is [%v]", urlStr, respCode, result, err)

	if err != nil {
		logger.Errorf("Post data view sql request failed: %v", err)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Post Failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Post data view sql request failed: %v", err))

		return "", fmt.Errorf("post data view sql request failed: %v", err)
	}
	if respCode != http.StatusOK {
		// 转成 baseerror
		var baseError rest.BaseError
		if err := sonic.Unmarshal(result, &baseError); err != nil {
			logger.Errorf("unmarshal BaseError failed: %v\n", err)

			// 添加异常时的 trace 属性
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal BaseError failed")
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Unmarshal BaseError failed: %v", err))

			return "", err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode, BaseError: baseError}
		logger.Errorf("data view config invalid: %v", httpErr.Error())

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Post data view sql failed: %v", baseError))

		return "", fmt.Errorf("data view config invalid: %v", httpErr.Error())
	}

	var dataViewSql interfaces.DataViewSql
	if err := sonic.Unmarshal(result, &dataViewSql); err != nil {
		logger.Errorf("unmarshal DataViewSql failed: %v\n", err)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal DataViewSql failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Unmarshal DataViewSql failed: %v", err))

		return "", err
	}

	// 添加成功时的 trace 属性
	o11y.AddHttpAttrs4Ok(span, respCode)
	return dataViewSql.SqlStr, nil
}
