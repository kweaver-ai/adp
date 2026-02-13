// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package index_base

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get index bases by types", trace.WithSpanKind(trace.SpanKindClient))
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
		errDetails := fmt.Sprintf("Get indexbases by base types '%s' failed, %s", baseTypes, err.Error())
		logger.Error(errDetails)

		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http GET index bases by types failed")

		return nil, err
	}

	if respCode != http.StatusOK {
		var baseError rest.BaseError
		if err := sonic.Unmarshal(respData, &baseError); err != nil {
			errDetails := fmt.Sprintf("Unmalshal baesError failed: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal baesError failed")

			return nil, err
		}

		logger.Errorf("Get indexbases '%s' failed, respCode: %d, error: %v", baseTypes, respCode, baseError)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", baseError.Description, baseError.ErrorDetails))
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status code is not 200")
		return nil, fmt.Errorf("get indexbases '%s' failed, errDetails: %v", baseTypes, baseError.ErrorDetails)
	}

	var bases []interfaces.IndexBase
	if err := sonic.Unmarshal(respData, &bases); err != nil {
		errDetails := fmt.Sprintf("Unmarshal indexbase respData failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal indexbase info failed")

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
		logger.Warn(errDetails)
		o11y.Warn(ctx, errDetails)
		// 如果有的存在，有的不存在，不报错，返回存在的索引库信息
		// return nil, fmt.Errorf(errDetails)
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return bases, nil
}

// 根据索引库类型获取索引库详情
func (iba *indexBaseAccess) GetManyIndexBasesByTypes(ctx context.Context, types []string) ([]interfaces.IndexBase, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get index bases by types", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	baseTypes := strings.Join(types, ",")

	span.SetAttributes(attr.Key("base_types").String(baseTypes))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         iba.appSetting.IndexBaseUrl,
		HttpMethod:      http.MethodPost,
		HttpContentType: rest.ContentTypeJson,
	})

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:           interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:      accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE:    accountInfo.Type,
		interfaces.HTTP_HEADER_METHOD_OVERRIDE: http.MethodGet,
	}

	reqBody := map[string]any{
		"base_types": baseTypes,
	}
	respCode, respData, err := iba.httpClient.PostNoUnmarshal(ctx, iba.appSetting.IndexBaseUrl, headers, reqBody)
	if err != nil {
		errDetails := fmt.Sprintf("Get indexbases by base types '%s' failed, %s", baseTypes, err.Error())
		logger.Error(errDetails)

		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http GET index bases by types failed")

		return nil, err
	}

	if respCode != http.StatusOK {
		var baseError rest.BaseError
		if err := sonic.Unmarshal(respData, &baseError); err != nil {
			errDetails := fmt.Sprintf("Unmalshal baesError failed: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal baesError failed")

			return nil, err
		}

		o11y.Error(ctx, fmt.Sprintf("%s. %v", baseError.Description, baseError.ErrorDetails))
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status code is not 200")
		return nil, fmt.Errorf("get indexbases '%s' failed, errDetails: %v", baseTypes, baseError.ErrorDetails)
	}

	var bases []interfaces.IndexBase
	if err := sonic.Unmarshal(respData, &bases); err != nil {
		errDetails := fmt.Sprintf("Unmarshal indexbase respData failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal indexbase info failed")

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
		logger.Warn(errDetails)
		o11y.Warn(ctx, errDetails)
		// 如果有的存在，有的不存在，不报错，返回存在的索引库信息
		// return nil, fmt.Errorf(errDetails)
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return bases, nil
}

// 根据索引库类型获取索引库简单信息
func (iba *indexBaseAccess) GetSimpleIndexBasesByTypes(ctx context.Context, types []string) ([]interfaces.SimpleIndexBase, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "根据索引库类型获取索引库信息", trace.WithSpanKind(trace.SpanKindClient))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         iba.appSetting.IndexBaseUrl,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})

	span.SetAttributes(attr.Key("base_types").StringSlice(types))
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

	baseTypes := strings.Join(types, ",")
	url := fmt.Sprintf("%s/%s", iba.appSetting.IndexBaseUrl, baseTypes)
	respCode, respData, err := iba.httpClient.GetNoUnmarshal(ctx, url, nil, headers)
	if err != nil {
		logger.Errorf("Get index base by base types '%s' failed, %s", baseTypes, err)
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Get Failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get index base request failed: %v", err))

		return nil, err
	}

	if respCode != http.StatusOK {
		// 转成 baseerror
		var baseError rest.BaseError
		if err := sonic.Unmarshal(respData, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)

			// 添加异常时的 trace 属性
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal BaesError failed")
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Unmalshal BaesError failed: %v", err))

			return nil, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode, BaseError: baseError}
		logger.Errorf("Get index base return error: %v", httpErr.Error())

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("get index base %v return error: %v", types, httpErr.Error()))

		return nil, fmt.Errorf("get index base %v return error %v", types, httpErr.Error())
	}

	if respData == nil {
		return nil, fmt.Errorf("get index base %v return null", types)
	}

	var bases []interfaces.SimpleIndexBase
	if err := sonic.Unmarshal(respData, &bases); err != nil {
		logger.Errorf("Unmarshal indexbase respData failed, %s", err)
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal indexbase respData failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Unmalshal indexbase respData failed: %v", err))

		return nil, err
	}

	if len(bases) < len(types) {
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "indexbase doesn't exist")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Have any IndexBase[%v] doesn't exist, expect number is %d, actual number is %d", types, len(types), len(bases)))
		return nil, fmt.Errorf("have any IndexBase[%v] doesn't exist, expect number is %d, actual number is %d", types, len(types), len(bases))
	}

	// 添加成功时的 trace 属性
	o11y.AddHttpAttrs4Ok(span, respCode)
	return bases, nil
}

func (iba *indexBaseAccess) ListIndexBases(ctx context.Context) ([]interfaces.SimpleIndexBase, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: List index bases", trace.WithSpanKind(trace.SpanKindClient))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         iba.appSetting.IndexBaseUrl,
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
	queryValues := url.Values{}
	queryValues.Add("limit", "-1")

	respCode, respData, err := iba.httpClient.GetNoUnmarshal(ctx, iba.appSetting.IndexBaseUrl, queryValues, headers)
	if err != nil {
		logger.Errorf("List index bases failed, %s", err)
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Get Failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("List index bases request failed: %v", err))

		return nil, err
	}

	if respCode != http.StatusOK {
		// 转成 baseerror
		var baseError rest.BaseError
		if err := sonic.Unmarshal(respData, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v", err)

			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal BaesError failed")
			o11y.Error(ctx, fmt.Sprintf("Unmalshal BaesError failed: %v", err))

			return nil, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode, BaseError: baseError}
		logger.Errorf("List index bases return error: %v", httpErr.Error())

		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		o11y.Error(ctx, fmt.Sprintf("List index bases return error: %v", httpErr.Error()))

		return nil, fmt.Errorf("list index bases return error %v", httpErr.Error())
	}

	if respData == nil {
		return nil, fmt.Errorf("list index bases return null")
	}

	var bases []interfaces.SimpleIndexBase
	if err := sonic.Unmarshal(respData, &bases); err != nil {
		logger.Errorf("Unmarshal indexbase respData failed, %s", err)
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal indexbase respData failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Unmalshal indexbase respData failed: %v", err))

		return nil, err
	}

	// 添加成功时的 trace 属性
	o11y.AddHttpAttrs4Ok(span, respCode)
	return bases, nil
}
