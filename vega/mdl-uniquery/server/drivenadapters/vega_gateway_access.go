package drivenadapters

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"

	"uniquery/common"
	"uniquery/interfaces"
)

var (
	vgAccessOnce sync.Once
	vgAccess     interfaces.VegaGatewayAccess
)

type vegaGatewayAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewVegaGatewayAccess(appSetting *common.AppSetting) interfaces.VegaGatewayAccess {
	vgAccessOnce.Do(func() {
		vgAccess = &vegaGatewayAccess{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
			// httpClient: common.NewHTTPClientWithOptions(rest.HttpClientOptions{
			// 	TimeOut: 1800,
			// }),
		}
	})

	return vgAccess
}

func (vga *vegaGatewayAccess) FetchDataNoUnmarshal(ctx context.Context, params *interfaces.FetchVegaDataParams) ([]byte, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "执行sql取数", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	// 获取账户信息和构建请求头
	vegaViewHeaders := vga.buildRequestHeaders(ctx)

	// 根据查询类型执行不同的请求逻辑
	respCode, result, err := vga.executeRequest(ctx, span, params, vegaViewHeaders)

	// 统一处理响应
	return vga.handleResponse(ctx, span, respCode, result, err)
}

// buildRequestHeaders 构建Vega网关请求头
func (vga *vegaGatewayAccess) buildRequestHeaders(ctx context.Context) map[string]string {
	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	return map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
		"X-Presto-User":                     "admin",
	}
}

// executeRequest 执行Vega网关请求
func (vga *vegaGatewayAccess) executeRequest(ctx context.Context, span trace.Span, params *interfaces.FetchVegaDataParams,
	headers map[string]string) (int, []byte, error) {
	switch params.QueryType {
	case interfaces.QueryType_DSL:
		return vga.executeDSLRequest(ctx, span, params, headers)
	case interfaces.QueryType_SQL:
		return vga.executeSQLRequest(ctx, span, params, headers)
	default:
		return 0, nil, fmt.Errorf("query type %s is not supported", params.QueryType)
	}
}

// executeDSLRequest 执行DSL查询请求
func (vga *vegaGatewayAccess) executeDSLRequest(ctx context.Context, span trace.Span, params *interfaces.FetchVegaDataParams,
	headers map[string]string) (int, []byte, error) {
	body := map[string]any{
		"type":         interfaces.VegaQueryType_DSL,
		"catalog_name": params.CatalogName,
		"table_name":   params.TableNames,
		"dsl":          params.Dsl,
	}

	logger.Infof("request type is [%d], catalog_name is [%s], table_name is [%s], dsl is [%s]",
		interfaces.VegaQueryType_DSL, params.CatalogName, params.TableNames, params.Dsl)

	urlStr := fmt.Sprintf("%s/fetch", vga.appSetting.DataConnGatewayUrl)
	vga.setSpanAttributes(span, urlStr, http.MethodPost)

	respCode, result, err := vga.httpClient.PostNoUnmarshal(ctx, urlStr, headers, body)
	logger.Debugf("post [%s] finished, request body is %v, response code is [%d], error is [%v]",
		urlStr, body, respCode, err)

	return respCode, result, err
}

// executeSQLRequest 执行SQL查询请求
func (vga *vegaGatewayAccess) executeSQLRequest(ctx context.Context, span trace.Span, params *interfaces.FetchVegaDataParams,
	headers map[string]string) (int, []byte, error) {
	if params.NextUri == "" {
		return vga.executeInitialSQLRequest(ctx, span, params, headers)
	} else {
		return vga.executeNextBatchRequest(ctx, span, params, headers)
	}
}

// executeInitialSQLRequest 执行初始SQL请求
func (vga *vegaGatewayAccess) executeInitialSQLRequest(ctx context.Context, span trace.Span, params *interfaces.FetchVegaDataParams,
	headers map[string]string) (int, []byte, error) {
	body := vga.buildSQLRequestBody(params)
	urlStr := vga.buildSQLRequestURL(params)

	vga.setSpanAttributes(span, urlStr, http.MethodPost)

	respCode, result, err := vga.httpClient.PostNoUnmarshal(ctx, urlStr, headers, body)
	logger.Debugf("post [%s] finished, request sql is [%s], request body is %v, response code is [%d], error is [%v]",
		urlStr, params.SqlStr, body, respCode, err)

	return respCode, result, err
}

// executeNextBatchRequest 执行下一批数据请求
func (vga *vegaGatewayAccess) executeNextBatchRequest(ctx context.Context, span trace.Span, params *interfaces.FetchVegaDataParams,
	headers map[string]string) (int, []byte, error) {
	httpUrl := vga.buildNextBatchURL(params)

	vga.setSpanAttributes(span, httpUrl, http.MethodGet)

	respCode, result, err := vga.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, headers)
	logger.Debugf("get [%s] finished, response code is [%d], error is [%v]",
		httpUrl, respCode, err)

	return respCode, result, err
}

// buildSQLRequestBody 构建SQL请求体
func (vga *vegaGatewayAccess) buildSQLRequestBody(params *interfaces.FetchVegaDataParams) map[string]any {
	body := map[string]any{"sql": params.SqlStr}

	if params.UseSearchAfter {
		if params.IsSingleDataSource {
			// 原子视图走下推数据源查询，type = 2 表示流式查询
			body["type"] = interfaces.VegaDataSourceQueryType_Stream
			body["data_source_id"] = params.DataSourceID
		} else {
			// 非原子视图走data-connection查询，type = 1 表示流式查询
			body["type"] = interfaces.VegaQueryType_Stream
		}
		if params.Timeout != 0 {
			body["timeout"] = params.Timeout
		}
		if params.Limit > 0 {
			body["batch_size"] = params.Limit
		}
	} else {
		if params.IsSingleDataSource {
			// 原子视图走下推数据源查询，type = 1 同步查询
			body["type"] = interfaces.VegaDataSourceQueryType_Sync
			body["data_source_id"] = params.DataSourceID
		} else {
			// 非原子视图走data-connection查询, type = 0 同步查询
			body["type"] = interfaces.VegaQueryType_Sync
		}
	}

	return body
}

// buildSQLRequestURL 构建SQL请求URL
func (vga *vegaGatewayAccess) buildSQLRequestURL(params *interfaces.FetchVegaDataParams) string {
	if params.IsSingleDataSource {
		return fmt.Sprintf("%s/fetch", vga.appSetting.VegaGatewayProUrl)
	} else {
		return fmt.Sprintf("%s/fetch", vga.appSetting.DataConnGatewayUrl)
	}
}

// buildNextBatchURL 构建下一批数据请求URL
func (vga *vegaGatewayAccess) buildNextBatchURL(params *interfaces.FetchVegaDataParams) string {
	queryValues := make(url.Values)

	if params.IsSingleDataSource {
		if params.Limit > 0 {
			queryValues.Add("batch_size", fmt.Sprintf("%d", params.Limit))
		}
		queryStr := queryValues.Encode()
		return fmt.Sprintf("%s/fetch/%s?%s", vga.appSetting.VegaGatewayProUrl, params.NextUri, queryStr)
	} else {
		if params.Limit > 0 {
			queryValues.Add("batchSize", fmt.Sprintf("%d", params.Limit))
		}
		queryStr := queryValues.Encode()
		return fmt.Sprintf("%s/statement/executing/%s?%s", vga.appSetting.DataConnGatewayUrl, params.NextUri, queryStr)
	}
}

// setSpanAttributes 设置Span属性
func (vga *vegaGatewayAccess) setSpanAttributes(span trace.Span, url string, method string) {
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         url,
		HttpMethod:      method,
		HttpContentType: rest.ContentTypeJson,
	})
}

// handleResponse 统一处理响应
func (vga *vegaGatewayAccess) handleResponse(ctx context.Context, span trace.Span, respCode int, result []byte, err error) ([]byte, error) {
	if err != nil {
		vga.logError(ctx, span, respCode, "Http Post Failed", err)
		return nil, fmt.Errorf("query data from vega gateway failed: %v", err)
	}

	if respCode != http.StatusOK {
		return vga.handleErrorResponse(ctx, span, respCode, result)
	}

	if result == nil {
		o11y.AddHttpAttrs4Ok(span, respCode)
		o11y.Warn(ctx, "Http response body is null")
		return nil, nil
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return result, nil
}

// handleErrorResponse 处理错误响应
func (vga *vegaGatewayAccess) handleErrorResponse(ctx context.Context, span trace.Span, respCode int, result []byte) ([]byte, error) {
	var vegaError *VegaError
	if err := sonic.Unmarshal(result, &vegaError); err != nil {
		vga.logError(ctx, span, respCode, "Unmalshal VegaError failed", err)
		return nil, err
	}

	vga.logError(ctx, span, respCode, "Http status is not 200", vegaError)
	return nil, fmt.Errorf("query data from vega gateway error: %v", vegaError.Error())
}

// logError 统一错误日志记录
func (vga *vegaGatewayAccess) logError(ctx context.Context, span trace.Span, respCode int, errorType string, err error) {
	logger.Errorf("query data from vega gateway failed: %v", err)
	o11y.AddHttpAttrs4Error(span, respCode, "InternalError", errorType)
	o11y.Error(ctx, fmt.Sprintf("query data from vega gateway failed: %v", err))
}
