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
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"uniquery/common"
	"uniquery/interfaces"
)

var (
	dvrcrAccessOnce sync.Once
	dvrcrAccess     interfaces.DataViewRowColumnRuleAccess
)

type dataViewRowColumnRuleAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewDataViewRowColumnRuleAccess(appSetting *common.AppSetting) interfaces.DataViewRowColumnRuleAccess {
	dvrcrAccessOnce.Do(func() {
		dvrcrAccess = &dataViewRowColumnRuleAccess{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
		}
	})
	return dvrcrAccess
}

func (dvr *dataViewRowColumnRuleAccess) GetRulesByViewID(ctx context.Context, viewID string) ([]*interfaces.DataViewRowColumnRule, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get row column rules by view ID from data-model service", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	httpUrl := fmt.Sprintf("%s/data-view-row-column-rules", dvr.appSetting.DataModelInUrl)

	span.SetAttributes(attr.Key("view_id").String(viewID))
	var queryValues url.Values = make(url.Values)
	queryValues.Add("view_id", viewID)

	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		"X-Language":                        rest.GetLanguageByCtx(ctx),
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	respCode, respData, err := dvr.httpClient.GetNoUnmarshal(ctx, httpUrl, queryValues, headers)
	if err != nil {
		errDetails := fmt.Sprintf("GetDataViewRowColumnRulesByViewID http request failed: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http get row column rules failed")

		return nil, fmt.Errorf("get request method failed: %s", err)
	}

	if respCode != http.StatusOK {
		logger.Errorf("get row column rules failed: %s", respData)

		var baseError rest.BaseError
		if err := sonic.Unmarshal(respData, &baseError); err != nil {
			logger.Errorf("Unmalshal baesError failed: %s", err)
			o11y.Error(ctx, err.Error())
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal baseError failed")
			return nil, err
		}

		o11y.Error(ctx, fmt.Sprintf("%s. %v", baseError.Description, baseError.ErrorDetails))
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status code is not 200")
		return nil, fmt.Errorf("GetDataViewRowColumnRulesByViewID failed: %s", baseError.ErrorDetails)
	}

	var result interfaces.ListRowColumnRulesResult
	if err = sonic.Unmarshal(respData, &result); err != nil {
		logger.Errorf("Unmarshal row column rules failed: %s", err)
		o11y.Error(ctx, err.Error())
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal row column rules info failed")
		return nil, err
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return result.Entries, nil
}
