package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"vega-gateway-pro/common"
	"vega-gateway-pro/interfaces"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"
)

var (
	vvAccessOnce sync.Once
	vcAccess     interfaces.VegaCalculateAccess
)

type vegaCalculateAccess struct {
	appSetting                  *common.AppSetting
	vegaCalculateCoordinatorUrl string
	httpClient                  rest.HTTPClient
}

func NewVegaCalculateAccess(appSetting *common.AppSetting) interfaces.VegaCalculateAccess {
	vvAccessOnce.Do(func() {
		vcAccess = &vegaCalculateAccess{
			appSetting:                  appSetting,
			vegaCalculateCoordinatorUrl: appSetting.VegaCalculateCoordinatorUrl,
			httpClient:                  common.NewHTTPClientWithOptions(rest.HttpClientOptions{TimeOut: 0}),
		}
	})

	return vcAccess
}

func (vca *vegaCalculateAccess) StatementQuery(ctx context.Context, sql string) (*interfaces.VegaCalculateData, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "执行sql取数", trace.WithSpanKind(trace.SpanKindClient))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         vca.vegaCalculateCoordinatorUrl,
		HttpMethod:      http.MethodPost,
		HttpContentType: rest.ContentTypeJson,
	})
	defer span.End()

	httpUrl := fmt.Sprintf("%s/v1/statement", vca.vegaCalculateCoordinatorUrl)
	vegaCalculateHeaders := map[string]string{
		interfaces.CONTENT_TYPE_NAME: interfaces.CONTENT_TYPE_JSON,
		"X-Presto-User":              "admin",
	}

	respCode, result, err := vca.httpClient.PostNoUnmarshal(ctx, httpUrl, vegaCalculateHeaders, []byte(sql))
	logger.Debugf("post [%s] finished, request sql is [%s], response code is [%d], result is [%s], error is [%v]",
		httpUrl, sql, respCode, result, err)

	return vca.handleResponse(ctx, span, respCode, result, err)
}

func (vca *vegaCalculateAccess) NextUriQuery(ctx context.Context, nextUri string) (*interfaces.VegaCalculateData, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "获取下一页数据", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	vegaCalculateHeaders := map[string]string{
		"X-Presto-User": "admin",
	}

	respCode, result, err := vca.httpClient.GetNoUnmarshal(ctx, nextUri, nil, vegaCalculateHeaders)
	logger.Debugf("get [%s] is finished, response code [%d], error is [%v]", nextUri, respCode, err)

	return vca.handleResponse(ctx, span, respCode, result, err)
}

func (vca *vegaCalculateAccess) handleResponse(ctx context.Context, span trace.Span, respCode int, result []byte, err error) (*interfaces.VegaCalculateData, error) {
	if err != nil {
		logger.Errorf("fetch data from vega calculate failed: %v", err)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Get Failed")
		o11y.Error(ctx, fmt.Sprintf("fetch data from vega calculate failed: %v", err))
		return nil, fmt.Errorf("fetch data from vega calculate failed: %v", err)
	}

	if respCode != http.StatusOK {
		var vegaError interfaces.VegaCalculateError
		if err := json.Unmarshal(result, &vegaError); err != nil {
			logger.Errorf("unmalshal VegaError failed: %v\n", err)
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal VegaError failed")
			o11y.Error(ctx, fmt.Sprintf("Unmalshal VegaError failed: %v", err))
			return nil, err
		}
		httpErr := &rest.HTTPError{
			HTTPCode: http.StatusInternalServerError,
			BaseError: rest.BaseError{
				ErrorCode:    rest.PublicError_InternalServerError,
				Description:  "sql execute failed",
				ErrorDetails: vegaError.Error.Message,
			}}
		logger.Errorf("fetch data from vega calculate Error: %v", httpErr.Error())
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		o11y.Error(ctx, fmt.Sprintf("fetch data from vega calculate failed: %v", vegaError))
		return nil, fmt.Errorf("fetch data from vega calculate Error: %v", httpErr.Error())
	}

	if result == nil {
		o11y.AddHttpAttrs4Ok(span, respCode)
		o11y.Warn(ctx, "Http response body is null")
		return nil, fmt.Errorf("Http response body is null")
	}

	// 先尝试解析错误信息
	var vegaError interfaces.VegaCalculateError
	if err := json.Unmarshal(result, &vegaError); err == nil && vegaError.Stats.State == "FAILED" {
		httpErr := &rest.HTTPError{
			HTTPCode: http.StatusInternalServerError,
			BaseError: rest.BaseError{
				ErrorCode:    rest.PublicError_InternalServerError,
				Description:  "sql execute failed",
				ErrorDetails: vegaError.Error.Message,
			}}
		logger.Errorf("fetch data from vega calculate Error: %v", httpErr.Error())
		o11y.AddHttpAttrs4Error(span, respCode, rest.PublicError_InternalServerError, "sql execute failed")
		o11y.Error(ctx, fmt.Sprintf("fetch data from vega calculate failed: %v", vegaError))
		return nil, fmt.Errorf("fetch data from vega calculate Error: %v", httpErr.Error())
	}

	var vegaFetchData interfaces.VegaCalculateData
	if err := sonic.Unmarshal(result, &vegaFetchData); err != nil {
		logger.Errorf("unmalshal vega datas info failed: %v\n", err)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal vega datas failed")
		o11y.Error(ctx, fmt.Sprintf("Unmalshal vega datas failed: %v", err))
		return nil, err
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return &vegaFetchData, nil
}
