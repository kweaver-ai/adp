// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	mmAccessOnce sync.Once
	mmAccess     interfaces.MetricModelAccess
)

type metricModelAccess struct {
	appSetting       *common.AppSetting
	metricModelUrl   string
	httpClient       rest.HTTPClient
	oAuth2HTTPClient rest.HTTPClient
}

func NewMetricModelAccess(appSetting *common.AppSetting) interfaces.MetricModelAccess {
	mmAccessOnce.Do(func() {
		mmAccess = &metricModelAccess{
			appSetting:       appSetting,
			metricModelUrl:   appSetting.DataModelInUrl,
			oAuth2HTTPClient: common.NewHTTPClient(),
			httpClient:       common.NewHTTPClient(),
		}
	})
	return mmAccess
}

// 获取指标模型信息，包含日志分组的过滤条件
func (mma *metricModelAccess) GetMetricModel(ctx context.Context, modelId string) ([]interfaces.MetricModel, bool, error) {
	// ?include_view=true 去掉，变成 include_view=false。先获取指标模型信息再获取视图信息，分开请求
	httpUrl := fmt.Sprintf("%s/metric-models/%s", mma.metricModelUrl, modelId)
	// http client 发送请求时，在 RoundTrip 时是用 transport 在 RoundTrip，此时的 transport 是 otelhttp.NewTransport 的，
	// otelhttp.NewTransport 的 RoundTrip 时会对 propagator 做 inject, 即 t.propagators.Inject
	ctx, span := ar_trace.Tracer.Start(ctx, "请求 data-model 获取指标模型信息", trace.WithSpanKind(trace.SpanKindClient))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})
	defer span.End()

	var (
		respCode int
		result   []byte
		err      error
	)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	metricModelHeaders := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}
	// httpClient 的请求新增参数支持上下文的处理请求的函数
	respCode, result, err = mma.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, metricModelHeaders)

	var emptyModel []interfaces.MetricModel
	if err != nil {
		logger.Errorf("get request method failed: %v", err)
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Get Failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get metric model request failed: %v", err))

		return emptyModel, false, fmt.Errorf("get request method failed: %v", err)
	}

	if respCode == http.StatusNotFound {
		logger.Errorf("metric model %s not exists", modelId)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, fmt.Sprintf("Metric model [%s] not found", modelId))

		return emptyModel, false, nil
	}

	if respCode != http.StatusOK {
		logger.Errorf("get metric model failed: %v", result)

		var baseError rest.BaseError
		if err := sonic.Unmarshal(result, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)

			// 添加异常时的 trace 属性
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal BaesError failed")
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Unmalshal BaesError failed: %v", err))

			return emptyModel, false, err
		}

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get metric model failed: %v", err))

		return emptyModel, false, fmt.Errorf("get metric model failed: %v", baseError.ErrorDetails)
	}

	if result == nil {
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, "Http response body is null")

		return emptyModel, false, nil
	}

	// 处理返回结果 result
	var metricModels []interfaces.MetricModel
	if err := sonic.Unmarshal(result, &metricModels); err != nil {
		logger.Errorf("unmalshal metric model info failed: %v\n", err)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal metric model info failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Unmalshal metric model info failed: %v", err))

		return emptyModel, false, err
	}

	// 如果是config的formula,则在这转换
	for i := range metricModels {
		if metricModels[i].MetricType == interfaces.DERIVED_METRIC {
			//  把 formula_config转成 SQLConfig
			var derivedConfig interfaces.DerivedConfig
			jsonData, err := json.Marshal(metricModels[i].FormulaConfig)
			if err != nil {
				return emptyModel, false, fmt.Errorf("derived Config Marshal error: %s", err.Error())
			}
			err = json.Unmarshal(jsonData, &derivedConfig)
			if err != nil {
				return emptyModel, false, fmt.Errorf("derived Config Unmarshal error: %s", err.Error())
			}
			// FormulaConfig 赋值为 SqlConfig
			metricModels[i].FormulaConfig = derivedConfig
		} else if metricModels[i].QueryType == interfaces.SQL {
			//  把 formula_config转成 SQLConfig
			var sqlConfig interfaces.SQLConfig
			jsonData, err := json.Marshal(metricModels[i].FormulaConfig)
			if err != nil {
				return emptyModel, false, fmt.Errorf("SQL Config Marshal error: %s", err.Error())
			}
			err = json.Unmarshal(jsonData, &sqlConfig)
			if err != nil {
				return emptyModel, false, fmt.Errorf("SQL Config Unmarshal error: %s", err.Error())
			}
			// FormulaConfig 赋值为 SqlConfig
			metricModels[i].FormulaConfig = sqlConfig
		}
	}

	// 添加成功时的 trace 属性
	o11y.AddHttpAttrs4Ok(span, respCode)

	return metricModels, true, nil
}

// 获取指标模型ID
func (mma *metricModelAccess) GetMetricModelIDByName(ctx context.Context, groupName, modelName string) (string, bool, error) {
	httpUrl := fmt.Sprintf("%s/metric-model-groups/%s/metric-models/%s", mma.metricModelUrl, groupName, modelName)

	ctx, span := ar_trace.Tracer.Start(ctx, "请求 data-model 获取指标模型ID", trace.WithSpanKind(trace.SpanKindClient))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})
	defer span.End()
	var (
		respCode int
		result   []byte
		err      error
	)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	metricModelHeaders := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}
	// httpClient 的请求新增参数支持上下文的处理请求的函数
	respCode, result, err = mma.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, metricModelHeaders)

	if err != nil {
		logger.Errorf("get request method failed: %v", err)
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Get Failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get metric model id request failed: %v", err))

		return "", false, fmt.Errorf("get request method failed: %v", err)
	}

	if respCode == http.StatusNotFound {
		logger.Errorf("metric model [%s] not exists", modelName)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, fmt.Sprintf("Metric model [%s] not found", modelName))

		return "", false, nil
	}

	if respCode != http.StatusOK {
		logger.Errorf("get metric model id failed: %v", result)

		var baseError rest.BaseError
		if err := sonic.Unmarshal(result, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)

			// 添加异常时的 trace 属性
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal BaesError failed")
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Unmalshal BaesError failed: %v", err))

			return "", false, err
		}

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get metric model id failed: %v", err))

		return "", false, fmt.Errorf("get metric model id failed: %v", baseError.ErrorDetails)
	}

	if result == nil {
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, "Http response body is null")

		return "", false, nil
	}

	// 处理返回结果 result
	var metricModel interfaces.MetricModel
	if err := sonic.Unmarshal(result, &metricModel); err != nil {
		logger.Errorf("unmalshal metric model info failed: %v\n", err)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal metric model info failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Unmalshal metric model info failed: %v", err))

		return "", false, err
	}

	// 添加成功时的 trace 属性
	o11y.AddHttpAttrs4Ok(span, respCode)

	return metricModel.ModelID, true, nil
}

func (mma *metricModelAccess) GetMetricModels(ctx context.Context, modelIds []string) ([]interfaces.MetricModel, error) {
	ids := strings.Join(modelIds, ",")
	httpUrl := fmt.Sprintf("%s/metric-models/%s", mma.metricModelUrl, ids)
	ctx, span := ar_trace.Tracer.Start(ctx, "请求 data-model 获取指标模型信息", trace.WithSpanKind(trace.SpanKindClient))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})
	defer span.End()

	var (
		respCode int
		result   []byte
		err      error
	)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	metricModelHeaders := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}
	// httpClient 的请求新增参数支持上下文的处理请求的函数
	respCode, result, err = mma.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, metricModelHeaders)

	var emptyModel []interfaces.MetricModel
	if err != nil {
		logger.Errorf("get request method failed: %v", err)
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Get Failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get metric model request failed: %v", err))

		return emptyModel, fmt.Errorf("get request method failed: %v", err)
	}

	if respCode == http.StatusNotFound {
		logger.Warnf("metric model %s not exists", modelIds)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, fmt.Sprintf("Metric model [%v] not found", modelIds))

		return emptyModel, nil
	}

	if respCode != http.StatusOK {
		logger.Errorf("get metric model failed: %v", result)

		var baseError rest.BaseError
		if err := sonic.Unmarshal(result, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)

			// 添加异常时的 trace 属性
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal BaesError failed")
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Unmalshal BaesError failed: %v", err))

			return emptyModel, err
		}

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get metric model failed: %v", err))

		return emptyModel, fmt.Errorf("get metric model failed: %v", baseError.ErrorDetails)
	}

	if result == nil {
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, "Http response body is null")

		return emptyModel, nil
	}

	// 处理返回结果 result
	var metricModels []interfaces.MetricModel
	if err := sonic.Unmarshal(result, &metricModels); err != nil {
		logger.Errorf("unmalshal metric model info failed: %v\n", err)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal metric model info failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Unmalshal metric model info failed: %v", err))

		return emptyModel, err
	}

	// 添加成功时的 trace 属性
	o11y.AddHttpAttrs4Ok(span, respCode)

	return metricModels, nil
}
