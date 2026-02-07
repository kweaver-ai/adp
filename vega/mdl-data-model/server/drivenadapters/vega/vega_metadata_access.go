// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package vega

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

	"data-model/common"
	"data-model/interfaces"
)

var (
	vmAccessOnce sync.Once
	vmAccess     interfaces.VegaMetadataAccess
)

type vegaMetadataAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewVegaMetadataAccess(appSetting *common.AppSetting) interfaces.VegaMetadataAccess {
	vmAccessOnce.Do(func() {
		vmAccess = &vegaMetadataAccess{
			appSetting: appSetting,
			// 设置超时时间60min
			httpClient: rest.NewHTTPClient(),
		}
	})

	return vmAccess
}

func (vma *vegaMetadataAccess) ListMetadataTablesBySourceID(ctx context.Context, params *interfaces.ListMetadataTablesParams) ([]interfaces.SimpleMetadataTable, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "list vega metadata tables", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	var queryValues url.Values = make(url.Values)
	queryValues.Add("limit", fmt.Sprintf("%d", params.Limit))
	queryValues.Add("offset", fmt.Sprintf("%d", params.Offset))
	queryValues.Add("update_time", params.UpdateTime)

	queryStr := queryValues.Encode()
	urlStr := fmt.Sprintf("%s/table/batch?%s", vma.appSetting.VegaMetadataUrl, queryStr)

	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         urlStr,
		HttpMethod:      http.MethodPost,
		HttpContentType: rest.ContentTypeJson,
	})

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	requestBody := struct {
		DataSourceIDs []string `json:"ds_ids"`
	}{
		DataSourceIDs: []string{params.DataSourceId},
	}

	respCode, respData, err := vma.httpClient.PostNoUnmarshal(ctx, urlStr, headers, requestBody)
	logger.Debugf("get %s finished, response code is %d, request body is %+v, error is %v",
		urlStr, respCode, requestBody, err)

	if err != nil {
		logger.Errorf("DrivenMetadata ListMetadataTables request failed: %v", err)

		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Post Failed")
		o11y.Error(ctx, fmt.Sprintf("DrivenMetadata ListMetadataTables request failed: %v", err))

		return nil, fmt.Errorf("DrivenMetadata ListMetadataTables request failed: %v", err)
	}

	// 错误码结构是 code, description, detail，需要对错误码做转换
	if respCode != http.StatusOK {
		// 转成 baseerror. vega返回的错误码跟我们当前的不同，暂时先用
		var vegaError VegaError
		if err = sonic.Unmarshal(respData, &vegaError); err != nil {
			logger.Errorf("unmalshal VegaError failed: %v\n", err)

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
		logger.Errorf("List Vega Metadata Tables Error: %v", httpErr.Error())

		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		o11y.Error(ctx, fmt.Sprintf("List Vega Metadata Tables failed: %v", vegaError))

		return nil, fmt.Errorf("list Vega Metadata Tables Error: %v", httpErr.Error())
	}

	if respData == nil {
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, "Http response body is null")

		return nil, nil
	}

	result := struct {
		Entries []interfaces.SimpleMetadataTable `json:"entries"`
		Total   int                              `json:"total_count"`
	}{}
	if err = sonic.Unmarshal(respData, &result); err != nil {
		logger.Errorf("DrivenMetadata ListMetadataTables sonic.Unmarshal error: %v", err)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal vega metadata tables failed")
		o11y.Error(ctx, fmt.Sprintf("Unmalshal vega metadata tables failed: %v", err))

		return nil, err
	}
	logger.Debugf("list vega metadata tables result count is %d", len(result.Entries))

	o11y.AddHttpAttrs4Ok(span, respCode)
	return result.Entries, nil
}

func (vma *vegaMetadataAccess) GetMetadataTablesByIDs(ctx context.Context, tableIDs []string) ([]interfaces.MetadataTable, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "get vega metadata tables by ids", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	urlStr := fmt.Sprintf("%s/tableAndField/batch", vma.appSetting.VegaMetadataUrl)

	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         urlStr,
		HttpMethod:      http.MethodPost,
		HttpContentType: rest.ContentTypeJson,
	})

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	requestBody := struct {
		TableIDs []string `json:"table_ids"`
	}{
		TableIDs: tableIDs,
	}

	respCode, respData, err := vma.httpClient.PostNoUnmarshal(ctx, urlStr, headers, requestBody)
	logger.Debugf("get %s finished, response code is %d, request body is %+v, error is %v",
		urlStr, respCode, requestBody, err)

	if err != nil {
		logger.Errorf("DrivenMetadata GetMetadataTablesByIDs request failed: %v", err)

		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Post Failed")
		o11y.Error(ctx, fmt.Sprintf("DrivenMetadata GetMetadataTablesByIDs request failed: %v", err))

		return nil, fmt.Errorf("DrivenMetadata GetMetadataTablesByIDs request failed: %v", err)
	}

	// 错误码结构是 code, description, detail，需要对错误码做转换
	if respCode != http.StatusOK {
		// 转成 baseerror. vega返回的错误码跟我们当前的不同，暂时先用
		var vegaError VegaError
		if err = sonic.Unmarshal(respData, &vegaError); err != nil {
			logger.Errorf("unmalshal VegaError failed: %v\n", err)

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
		logger.Errorf("Get Vega Metadata Tables By IDs Error: %v", httpErr.Error())

		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		o11y.Error(ctx, fmt.Sprintf("Get Vega Metadata Tables By IDs failed: %v", vegaError))

		return nil, fmt.Errorf("get Vega Metadata Tables By IDs Error: %v", httpErr.Error())
	}

	if respData == nil {
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, "Http response body is null")

		return nil, nil
	}

	result := struct {
		Entries []interfaces.MetadataTable `json:"entries"`
		Total   int                        `json:"total_count"`
	}{}
	if err = sonic.Unmarshal(respData, &result); err != nil {
		logger.Errorf("DrivenMetadata GetMetadataTablesByIDs sonic.Unmarshal error: %v", err)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal vega metadata tables failed")
		o11y.Error(ctx, fmt.Sprintf("Unmalshal vega metadata tables failed: %v", err))

		return nil, err
	}
	logger.Debugf("get vega metadata tables by ids result count is %d", len(result.Entries))

	o11y.AddHttpAttrs4Ok(span, respCode)
	return result.Entries, nil
}

// func serializeAdvancedParams(tables []interfaces.MetadataTable) (err error) {
// 	for _, data := range tables {
// 		err = sonic.Unmarshal([]byte(data.Table.AdvancedParams), &data.Table.AdvancedParamsStruct)
// 		if err != nil {
// 			return err
// 		}
// 		for _, field := range data.FieldList {
// 			err = sonic.Unmarshal([]byte(field.AdvancedParams), &field.AdvancedParamsStruct)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }
