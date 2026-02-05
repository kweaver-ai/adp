package drivenadapters

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"uniquery/common"
	"uniquery/interfaces"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"
)

var (
	vdsAccessOnce sync.Once
	vdsAccess     interfaces.VegaDataSourceAccess
)

type vegaDataSourceAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewVegaDataSourceAccess(appSetting *common.AppSetting) interfaces.VegaDataSourceAccess {
	vdsAccessOnce.Do(func() {
		vdsAccess = &vegaDataSourceAccess{
			appSetting: appSetting,
			//httpClient:  common.NewHTTPClient(),
			httpClient: common.NewHTTPClient(),
		}
	})

	return vdsAccess
}

// 调用内部接口，根据id查询数据源
func (vdsa *vegaDataSourceAccess) GetDataSourceByID(ctx context.Context, id string) (datasource *interfaces.VegaDataSource, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "get vega data source by id", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	urlStr := fmt.Sprintf("%s/%s", vdsa.appSetting.DataConnDataSourceUrl, id)

	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         urlStr,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})
	defer span.End()

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	respCode, respData, err := vdsa.httpClient.GetNoUnmarshal(ctx, urlStr, nil, headers)
	logger.Debugf("get %s finished, response code is %d, result is %s, error is %v", urlStr, respCode, respData, err)

	if err != nil {
		logger.Errorf("DrivenMetadata GetDataSourceByID request failed: %v", err)

		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Get Failed")
		o11y.Error(ctx, fmt.Sprintf("DrivenMetadata GetDataSourceByID request failed: %v", err))

		return nil, fmt.Errorf("DrivenMetadata GetDataSourceByID request failed: %v", err)
	}

	// 错误码结构是 code, description, detail，需要对错误码做转换
	if respCode != http.StatusOK {
		// 转成 baseerror. vega返回的错误码跟我们当前的不同，暂时先用
		var vegaError *VegaError
		if err = sonic.Unmarshal(respData, &vegaError); err != nil {
			logger.Errorf("unmalshal VegaError failed: %v", err)

			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal VegaError failed")
			o11y.Error(ctx, fmt.Sprintf("Unmalshal VegaError failed: %v", err))

			return nil, err
		}

		logger.Errorf("Get Vega DataSource Error: %v", vegaError.Error())

		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		o11y.Error(ctx, fmt.Sprintf("Get Vega DataSource failed: %v", vegaError))

		return nil, fmt.Errorf("get Vega DataSource Error: %v", vegaError.Error())
	}

	if respData == nil {
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, "Http response body is null")

		return nil, nil
	}

	var res interfaces.VegaDataSource
	if err = sonic.Unmarshal(respData, &res); err != nil {
		logger.Errorf("DrivenMetadata GetDataSourceByID sonic.Unmarshal error: %v", err)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal vega data source failed")
		o11y.Error(ctx, fmt.Sprintf("Unmalshal vega data source failed: %v", err))

		return nil, err
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return &res, nil
}
