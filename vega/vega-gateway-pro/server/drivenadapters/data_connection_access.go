package drivenadapters

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"sync"
	"vega-gateway-pro/common"
	"vega-gateway-pro/common/rsa"
	"vega-gateway-pro/interfaces"
)

var (
	vdsAccessOnce sync.Once
	dcAccess      interfaces.DataConnectionAccess
)

type dataConnectionAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

type VegaError struct {
	Code        string      `json:"code"`        // 错误码
	Description string      `json:"description"` // 错误描述
	Detail      interface{} `json:"detail"`      // 详细内容
}

func NewDataConnectionAccess(appSetting *common.AppSetting) interfaces.DataConnectionAccess {
	vdsAccessOnce.Do(func() {
		dcAccess = &dataConnectionAccess{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
		}
	})

	return dcAccess
}

// 调用内部接口，根据catalog查询数据源
func (dca *dataConnectionAccess) GetDataSourceById(ctx context.Context, dataSourceId string) (*interfaces.DataSource, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "get data source by id", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	urlStr := fmt.Sprintf("%s/api/internal/data-connection/v1/datasource/%s", dca.appSetting.DataConnectionUrl, dataSourceId)

	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         urlStr,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})
	defer span.End()

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   interfaces.ADMIN_ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: interfaces.ADMIN_TYPE,
	}

	respCode, respData, err := dca.httpClient.GetNoUnmarshal(ctx, urlStr, nil, headers)
	logger.Debugf("get %s finished, response code is %d, result is %s, error is %v", urlStr, respCode, respData, err)

	if err != nil {
		logger.Errorf("DrivenMetadata GetDataSourceById request failed: %v", err)

		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Get Failed")
		o11y.Error(ctx, fmt.Sprintf("DrivenMetadata GetDataSourceById request failed: %v", err))

		return nil, fmt.Errorf("DrivenMetadata GetDataSourceById request failed: %v", err)
	}

	// 错误码结构是 code, description, detail，需要对错误码做转换
	if respCode != http.StatusOK {
		// 转成 baseerror. vega返回的错误码跟我们当前的不同，暂时先用
		var vegaError VegaError
		if err = sonic.Unmarshal(respData, &vegaError); err != nil {
			logger.Errorf("unmalshal VegaError failed: %v", err)

			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal VegaError failed")
			o11y.Error(ctx, fmt.Sprintf("Unmalshal VegaError failed: %v", err))

			return nil, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    vegaError.Code,
				Description:  vegaError.Description,
				ErrorDetails: vegaError.Detail,
			}}
		logger.Errorf("Get Vega DataSource Error: %v", httpErr.Error())

		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		o11y.Error(ctx, fmt.Sprintf("Get Vega DataSource failed: %v", vegaError))

		return nil, fmt.Errorf("get Vega DataSource Error: %v", httpErr.Error())
	}

	var res interfaces.DataSource
	if err = sonic.Unmarshal(respData, &res); err != nil {
		logger.Errorf("DrivenMetadata GetDataSourceById sonic.Unmarshal error: %v", err)

		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal DataSource failed")
		o11y.Error(ctx, fmt.Sprintf("Unmalshal DataSource failed: %v", err))

		return nil, err
	}

	// 解密密码
	res.BinData.Password, err = rsa.Decrypt(res.BinData.Password)
	if err != nil {
		logger.Errorf("Decrypt password failed: %s", err.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(err.Error())
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return &res, nil
}
