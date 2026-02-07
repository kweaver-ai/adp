// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"errors"
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
	dvAccessOnce sync.Once
	dvAccess     interfaces.DataViewAccess
)

type dataViewAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewDataViewAccess(appSetting *common.AppSetting) interfaces.DataViewAccess {
	dvAccessOnce.Do(func() {
		dvAccess = &dataViewAccess{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
		}
	})
	return dvAccess
}

// 批量根据 id 获取视图列表
func (dva *dataViewAccess) GetDataViewsByIDs(ctx context.Context, ids string, includeDataScopeView bool) ([]*interfaces.DataView, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get views by IDs from data-model service", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	httpUrl := fmt.Sprintf("%s/%s", dva.appSetting.DataViewUrl, ids)

	span.SetAttributes(attr.Key("view_ids").String(ids))
	var queryValues url.Values = make(url.Values)
	if includeDataScopeView {
		queryValues.Add("include_data_scope_views", "true")
	}

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

	respCode, respData, err := dva.httpClient.GetNoUnmarshal(ctx, httpUrl, queryValues, headers)
	if err != nil {
		errDetails := fmt.Sprintf("GetDataViewByIDs http request failed: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http get views failed")

		return nil, fmt.Errorf("get request method failed: %s", err)
	}

	if respCode != http.StatusOK {
		logger.Errorf("get data view failed: %s", respData)

		var baseError rest.BaseError
		if err := sonic.Unmarshal(respData, &baseError); err != nil {
			logger.Errorf("Unmalshal baesError failed: %s", err)
			o11y.Error(ctx, err.Error())
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal baseError failed")
			return nil, err
		}

		o11y.Error(ctx, fmt.Sprintf("%s. %v", baseError.Description, baseError.ErrorDetails))
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status code is not 200")
		return nil, fmt.Errorf("GetDataViewByIDs failed: %s", baseError.ErrorDetails)
	}

	var views []*interfaces.DataView
	if err = sonic.Unmarshal(respData, &views); err != nil {
		logger.Errorf("Unmarshal data view failed: %s", err)
		o11y.Error(ctx, err.Error())
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal data view info failed")
		return nil, err
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return views, nil
}

// 根据视图名称获取视图ID
func (dva *dataViewAccess) GetDataViewIDByName(ctx context.Context, viewName string) (string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get view id by name from data-model service", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	httpUrl := fmt.Sprintf("%s?name=%s&include_builtin=true", dva.appSetting.DataViewUrl, viewName)

	span.SetAttributes(attr.Key("view_name").String(viewName))
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

	respCode, respData, err := dva.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, headers)
	if err != nil {
		errDetails := fmt.Sprintf("GetDataViewIDByName http request failed: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http get views failed")

		return "", fmt.Errorf("get request method failed: %s", err)
	}

	if respCode != http.StatusOK {
		logger.Errorf("get data view ID by name failed: %v", respData)

		var baseError rest.BaseError
		if err := sonic.Unmarshal(respData, &baseError); err != nil {
			logger.Errorf("Unmalshal baesError failed: %s", err)
			o11y.Error(ctx, err.Error())
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal baseError failed")
			return "", err
		}

		o11y.Error(ctx, fmt.Sprintf("%s. %v", baseError.Description, baseError.ErrorDetails))
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status code is not 200")
		return "", fmt.Errorf("GetDataViewIDByName failed: %s", baseError.ErrorDetails)
	}

	result := struct {
		Entries []interfaces.DataView `json:"entries"`
		Total   int                   `json:"total_count"`
	}{}
	if err = sonic.Unmarshal(respData, &result); err != nil {
		logger.Errorf("Unmarshal data view failed: %s", err)
		o11y.Error(ctx, err.Error())
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal data view info failed")
		return "", err
	}
	if len(result.Entries) > 0 {
		o11y.AddHttpAttrs4Ok(span, respCode)
		return result.Entries[0].ViewID, nil
	} else {
		errDetails := "data view result is empty"
		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", errDetails)
		logger.Errorf(errDetails)
		return "", errors.New(errDetails)
	}
}
