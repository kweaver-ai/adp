package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	vvAccessOnce sync.Once
	vvAccess     interfaces.VegaAccess
)

type vegaViewAccess struct {
	appSetting     *common.AppSetting
	vegaGatewayUrl string
	vegaViewUrl    string
	httpClient     rest.HTTPClient
}

type VegaError struct {
	Code        string      `json:"code"`        // 错误码
	Description string      `json:"description"` // 错误描述
	Detail      interface{} `json:"detail"`      // 详细内容
}

func NewVegaAccess(appSetting *common.AppSetting) interfaces.VegaAccess {
	vvAccessOnce.Do(func() {
		vvAccess = &vegaViewAccess{
			appSetting:     appSetting,
			vegaGatewayUrl: appSetting.VegaGatewaysUrl,
			vegaViewUrl:    appSetting.VegaViewUrl,
			//httpClient:  common.NewHTTPClient(),
			httpClient: common.NewHTTPClient(),
		}
	})

	return vvAccess
}

// 获取Vega视图的字段信息
func (vva *vegaViewAccess) GetVegaViewFieldsByID(ctx context.Context, viewID string) (interfaces.VegaViewWithFields, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "计算公式有效性校验", trace.WithSpanKind(trace.SpanKindClient))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         vva.vegaViewUrl,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})
	defer span.End()

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	vegaViewHeaders := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	httpUrl := fmt.Sprintf("%s/%s", vva.vegaViewUrl, viewID)
	respCode, result, err := vva.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, vegaViewHeaders)
	logger.Debugf("get [%s] finished, response code is [%d], result is [%s], error is [%v]", vva.vegaViewUrl, respCode, result, err)

	if err != nil {
		logger.Errorf("Get Vega View Fields request failed: %v", err)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Get Failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get Vega View Fields request failed: %v", err))

		return interfaces.VegaViewWithFields{}, fmt.Errorf("get Vega View Fields request failed: %v", err)
	}
	// todo: 视图id不存在和uuid无效都是返回的400，错误码结构是 code, description, detail，需要对错误码做转换
	if respCode != http.StatusOK {
		// 转成 baseerror. vega返回的错误码跟我们当前的不同，暂时先用
		var vegaError VegaError
		if err := json.Unmarshal(result, &vegaError); err != nil {
			logger.Errorf("unmalshal VegaError failed: %v\n", err)

			// 添加异常时的 trace 属性
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal VegaError failed")
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Unmalshal VegaError failed: %v", err))

			return interfaces.VegaViewWithFields{}, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    vegaError.Code,
				Description:  vegaError.Description,
				ErrorDetails: vegaError.Detail,
			}}
		logger.Errorf("Get Vega View Error: %v", httpErr.Error())

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get Vega View failed: %v", vegaError))

		return interfaces.VegaViewWithFields{}, fmt.Errorf("get Vega View Error: %v", httpErr.Error())
	}

	if result == nil {
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, "Http response body is null")

		return interfaces.VegaViewWithFields{}, nil
	}

	// 处理返回结果 result, 读取field
	var vegaFields interfaces.VegaViewWithFields
	if err := sonic.Unmarshal(result, &vegaFields); err != nil {
		logger.Errorf("unmalshal vega view fields info failed: %v\n", err)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal vega view fields failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Unmalshal vega view fields failed: %v", err))

		return interfaces.VegaViewWithFields{}, err
	}

	// 添加成功时的 trace 属性
	o11y.AddHttpAttrs4Ok(span, respCode)
	return vegaFields, nil
}

func (vva *vegaViewAccess) FetchDatasFromVega(ctx context.Context, nextUri string, sql string) (interfaces.VegaFetchData, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "执行sql取数", trace.WithSpanKind(trace.SpanKindClient))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         vva.vegaGatewayUrl,
		HttpMethod:      http.MethodPost,
		HttpContentType: rest.ContentTypeJson,
	})
	defer span.End()

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	vegaViewHeaders := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
		"X-Presto-User":                     "admin",
	}
	var respCode int
	var result []byte
	var err error
	if nextUri == "" {
		// vegaGatewayUrl = /api/virtual_engine_service/v1
		// /api/virtual_engine_service/v1/fetch
		httpUrl := fmt.Sprintf("%s/fetch?type=1", vva.vegaGatewayUrl)
		respCode, result, err = vva.httpClient.PostNoUnmarshal(ctx, httpUrl, vegaViewHeaders, []byte(sql))
		logger.Debugf("post [%s] finished, request sql is [%s], response code is [%d], result is [%s], error is [%v]",
			httpUrl, sql, respCode, result, err)
	} else {
		// /api/virtual_engine_service/v1
		// statement/executing/20250516_051107_00077_dpx8g/xe0786066b47e4aeda3972d483db644f1/1
		httpUrl := fmt.Sprintf("%s/statement/executing/%s", vva.vegaGatewayUrl, nextUri)
		respCode, result, err = vva.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, vegaViewHeaders)
		logger.Debugf("get [%s] finished, request sql is [%s], response code is [%d], result is [%s], error is [%v]",
			httpUrl, sql, respCode, result, err)
	}

	if err != nil {
		logger.Errorf("fetch data from vega gateway by sql[%s] failed: %v", sql, err)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Get Failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("fetch data from vega gateway failed: %v", err))

		return interfaces.VegaFetchData{}, fmt.Errorf("fetch data from vega gateway failed: %v", err)
	}
	// todo: 视图id不存在和uuid无效都是返回的400，错误码结构是 code, description, detail，需要对错误码做转换
	if respCode != http.StatusOK {
		// 转成 baseerror. vega返回的错误码跟我们当前的不同，暂时先用
		var vegaError VegaError
		if err := json.Unmarshal(result, &vegaError); err != nil {
			logger.Errorf("unmalshal VegaError failed: %v\n", err)

			// 添加异常时的 trace 属性
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal VegaError failed")
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Unmalshal VegaError failed: %v", err))

			return interfaces.VegaFetchData{}, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    vegaError.Code,
				Description:  vegaError.Description,
				ErrorDetails: vegaError.Detail,
			}}
		logger.Errorf("fetch data from vega gateway by sql[%s] Error: %v", sql, httpErr.Error())

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("fetch data from vega gateway failed: %v", vegaError))

		return interfaces.VegaFetchData{}, fmt.Errorf("fetch data from vega gateway Error: %v", httpErr.Error())
	}

	if result == nil {
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, "Http response body is null")

		return interfaces.VegaFetchData{}, nil
	}

	// 处理返回结果 result, 读取field
	var vegaFetchData interfaces.VegaFetchData
	if err := sonic.Unmarshal(result, &vegaFetchData); err != nil {
		logger.Errorf("unmalshal vega datas info failed: %v\n", err)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal vega datas failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Unmalshal vega datas failed: %v", err))

		return interfaces.VegaFetchData{}, err
	}
	// 遍历直到netxUri为空
	// while

	// 添加成功时的 trace 属性
	o11y.AddHttpAttrs4Ok(span, respCode)
	return vegaFetchData, nil
}
