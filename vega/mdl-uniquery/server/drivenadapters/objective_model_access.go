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
	omAccessOnce sync.Once
	omAccess     interfaces.ObjectiveModelAccess
)

type objectiveModelAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewObjectiveModelAccess(appSetting *common.AppSetting) interfaces.ObjectiveModelAccess {
	omAccessOnce.Do(func() {
		omAccess = &objectiveModelAccess{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
		}
	})
	return omAccess
}

// 获取目标模型信息，包含日志分组的过滤条件
func (oma *objectiveModelAccess) GetObjectiveModel(ctx context.Context, modelId string) (interfaces.ObjectiveModel, bool, error) {
	httpUrl := fmt.Sprintf("%s/in/v1/objective-models/%s", oma.appSetting.DataModelUrl, modelId)
	// http client 发送请求时，在 RoundTrip 时是用 transport 在 RoundTrip，此时的 transport 是 otelhttp.NewTransport 的，
	// otelhttp.NewTransport 的 RoundTrip 时会对 propagator 做 inject, 即 t.propagators.Inject
	ctx, span := ar_trace.Tracer.Start(ctx, "请求 data-model 获取目标模型信息", trace.WithSpanKind(trace.SpanKindClient))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         httpUrl,
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
	// httpClient 的请求新增参数支持上下文的处理请求的函数
	respCode, result, err := oma.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, headers)

	var emptyModel interfaces.ObjectiveModel
	if err != nil {
		logger.Errorf("get request method failed: %v", err)
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Get Failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get objective model request failed: %v", err))

		return emptyModel, false, fmt.Errorf("get request method failed: %v", err)
	}

	if respCode == http.StatusNotFound {
		logger.Errorf("objective model %s not exists", modelId)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, fmt.Sprintf("objective model [%s] not found", modelId))

		return emptyModel, false, nil
	}

	if respCode != http.StatusOK {
		logger.Errorf("get objective model failed: %v", result)

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
		o11y.Error(ctx, fmt.Sprintf("Get objective model failed: %v", err))

		return emptyModel, false, fmt.Errorf("get objective model failed: %v", baseError.ErrorDetails)
	}

	if result == nil {
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, "Http response body is null")

		return emptyModel, false, nil
	}

	// 处理返回结果 result
	var objectiveModels []interfaces.ObjectiveModel
	if err := sonic.Unmarshal(result, &objectiveModels); err != nil {
		logger.Errorf("unmalshal objective model info failed: %v\n", err)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal objective model info failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Unmalshal objective model info failed: %v", err))

		return emptyModel, false, err
	}

	if len(objectiveModels) == 0 {
		return emptyModel, false, nil
	}
	objectiveModel := objectiveModels[0]

	if objectiveModel.ObjectiveType == interfaces.SLO {
		// 把 objective_config转成 SLOObjective
		var sloObjective interfaces.SLOObjective
		jsonData, err := json.Marshal(objectiveModel.ObjectiveConfig)
		if err != nil {
			return emptyModel, false, err
		}
		err = json.Unmarshal(jsonData, &sloObjective)
		if err != nil {
			return emptyModel, false, err
		}
		objectiveModel.ObjectiveConfig = sloObjective
	} else {
		// 把 objective_config转成 SLOObjective
		var kpiObjective interfaces.KPIObjective
		jsonData, err := json.Marshal(objectiveModel.ObjectiveConfig)
		if err != nil {
			return emptyModel, false, err
		}
		err = json.Unmarshal(jsonData, &kpiObjective)
		if err != nil {
			return emptyModel, false, err
		}
		objectiveModel.ObjectiveConfig = kpiObjective
	}

	// 添加成功时的 trace 属性
	o11y.AddHttpAttrs4Ok(span, respCode)

	return objectiveModel, true, nil
}
