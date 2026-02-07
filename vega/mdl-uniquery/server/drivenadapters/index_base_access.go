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
	"strings"
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
	ibAccessOnce sync.Once
	ibAccess     interfaces.IndexBaseAccess
)

type indexBaseAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewIndexBaseAccess(appSetting *common.AppSetting) interfaces.IndexBaseAccess {
	ibAccessOnce.Do(func() {
		ibAccess = &indexBaseAccess{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
		}
	})

	return ibAccess
}

// 根据索引库类型获取索引库详情
func (iba *indexBaseAccess) GetIndexBasesByTypes(ctx context.Context, types []string) ([]interfaces.IndexBase, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get index bases from index-base service", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	baseTypes := strings.Join(types, ",")
	url := fmt.Sprintf("%s/%s", iba.appSetting.IndexBaseUrl, baseTypes)

	span.SetAttributes(attr.Key("base_types").String(baseTypes))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         url,
		HttpMethod:      http.MethodGet,
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

	respCode, respData, err := iba.httpClient.GetNoUnmarshal(ctx, url, nil, headers)
	if err != nil {
		logger.Errorf("Get indexbases by base types '%s' failed, %s", baseTypes, err)
		o11y.Error(ctx, err.Error())
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http get index bases failed")
		return nil, err
	}

	if respCode != http.StatusOK {
		var baseError rest.BaseError
		if err := sonic.Unmarshal(respData, &baseError); err != nil {
			logger.Errorf("Unmalshal baesError failed: %s", err)
			o11y.Error(ctx, err.Error())
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal baseError failed")
			return nil, err
		}

		o11y.Error(ctx, fmt.Sprintf("%s. %v", baseError.Description, baseError.ErrorDetails))
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status code is not 200")
		return nil, fmt.Errorf("get indexbases '%s' failed,  errDetails: %s", baseTypes, baseError.ErrorDetails)
	}

	var bases []interfaces.IndexBase
	if err := sonic.Unmarshal(respData, &bases); err != nil {
		logger.Errorf("Unmarshal indexbase respData failed, %s", err)
		o11y.Error(ctx, err.Error())
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal index base info failed")
		return nil, err
	}

	// 索引库接口默认批量查询，只返回存在的索引库信息
	if len(bases) < len(types) {
		nonexistentBases := make([]string, 0)
		typesMap := make(map[string]struct{})

		for _, base := range bases {
			typesMap[base.BaseType] = struct{}{}
		}

		for _, baseType := range types {
			if _, ok := typesMap[baseType]; !ok {
				nonexistentBases = append(nonexistentBases, baseType)
			}
		}

		errDetails := fmt.Sprintf("IndexBases %v doesn't exist", nonexistentBases)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Index base was not found")
		return nil, errors.New(errDetails)
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return bases, nil
}

// 通过索引优化接口获取索引
func (iba *indexBaseAccess) GetIndices(ctx context.Context, types []string, start,
	end int64) (map[string]map[string]interfaces.Indice, int, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get indices from index-base service", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	baseTypes := strings.Join(types, ",")
	url := fmt.Sprintf("%s/%s/index_filters?from=%d&to=%d&merge=false", iba.appSetting.IndexBaseUrl, baseTypes, start, end)

	span.SetAttributes(attr.Key("base_types").String(baseTypes))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         url,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})

	respCode, respData, err := iba.httpClient.GetNoUnmarshal(ctx, url, nil, interfaces.Headers)
	if err != nil {
		logger.Errorf("Get indices by base types '%s' failed, %s", baseTypes, err)
		o11y.Error(ctx, err.Error())
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http get indices failed")
		return nil, respCode, err
	}

	if respCode != http.StatusOK {
		var baseError rest.BaseError
		if err := sonic.Unmarshal(respData, &baseError); err != nil {
			logger.Errorf("Unmarshal baesError failed: %s", err)
			o11y.Error(ctx, err.Error())
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal baseError failed")
			return nil, respCode, err
		}

		o11y.Error(ctx, fmt.Sprintf("%s. %v", baseError.Description, baseError.ErrorDetails))
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status code is not 200")
		return nil, respCode, fmt.Errorf("get indices by %s failed,  errDetails: %s", baseTypes, baseError.ErrorDetails)
	}

	var indicesResult map[string]map[string]interfaces.Indice
	if err := sonic.Unmarshal(respData, &indicesResult); err != nil {
		logger.Errorf("Unmarshal indices respData failed, %s", err)
		o11y.Error(ctx, err.Error())
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal indices info failed")
		return nil, respCode, err
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return indicesResult, respCode, nil

}
