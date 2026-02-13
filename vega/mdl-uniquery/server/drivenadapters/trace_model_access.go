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
	"sync"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"uniquery/common"
	"uniquery/interfaces"
)

var (
	tmAccessOnce sync.Once
	tmAccess     interfaces.TraceModelAccess
)

type traceModelAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewTraceModelAccess(appSetting *common.AppSetting) interfaces.TraceModelAccess {
	tmAccessOnce.Do(func() {
		tmAccess = &traceModelAccess{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
		}
	})
	return tmAccess
}

// 根据链路模型ID获取对象
func (tma *traceModelAccess) GetTraceModelByID(ctx context.Context, modelID string) (model interfaces.TraceModel, isExist bool, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 调用DataModel服务获取链路模型", trace.WithSpanKind(trace.SpanKindClient))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 声明models, 链路模型查询接口返回的是数组结构
	models := make([]interfaces.TraceModel, 1)

	// 2. 拼接fullPath
	fullPath := fmt.Sprintf("%s/in/v1/trace-models/%v", tma.appSetting.DataModelUrl, modelID)

	span.SetAttributes(
		attr.Key("request_url").String(fullPath),
		attr.Key("model_id").String(modelID),
	)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	// 3. http request
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		"X-Language":                        rest.GetLanguageByCtx(ctx),
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	respCode, respData, err := tma.httpClient.GetNoUnmarshal(ctx, fullPath, nil, headers)

	// 4. 分情况处理http response
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get trace model by http client: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return models[0], false, err
	}

	if respCode == http.StatusNotFound {
		errDetails := fmt.Sprintf("Failed to get trace model by http client: the trace model whose id equal to %s was not found", modelID)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return models[0], false, nil
	}

	if respCode != http.StatusOK {
		err := fmt.Errorf("failed to get trace model by http client: %s", string(respData))
		logger.Error(err.Error())
		o11y.Error(ctx, err.Error())
		return models[0], false, err
	}

	if err = sonic.Unmarshal(respData, &models); err != nil {
		errDetails := fmt.Sprintf("Unmarshal http response failed: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return models[0], false, err
	}

	// 5. 转换traceModel
	model, err = tma.commonProcessTraceModel(ctx, models[0])
	if err != nil {
		return model, false, err
	}

	return model, true, nil
}

// 模拟创建链路模型, 若创建成功, 会返回链路模型依赖的数据视图ID
func (tma *traceModelAccess) SimulateCreateTraceModel(ctx context.Context, model interfaces.TraceModel) (newModel interfaces.TraceModel, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 调用DataModel服务模拟创建链路模型", trace.WithSpanKind(trace.SpanKindClient))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 拼接fullPath
	fullPath := fmt.Sprintf("%s/in/v1/simulate-trace-models", tma.appSetting.DataModelUrl)

	span.SetAttributes(
		attr.Key("request_url").String(fullPath),
	)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	// 2. 发送http request
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		"X-Language":                        rest.GetLanguageByCtx(ctx),
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}
	respCode, respData, err := tma.httpClient.PostNoUnmarshal(ctx, fullPath, headers, model)

	// 3. 分情况处理http response
	simulateModel := interfaces.TraceModel{}
	if err != nil {
		errDetails := fmt.Sprintf("Failed to simulate create trace model by http client: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return simulateModel, err
	}

	if respCode != http.StatusOK {
		errDetails := fmt.Sprintf("Failed to simulate create trace model by http client: %s", string(respData))
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return simulateModel, errors.New(errDetails)
	}

	if err = sonic.Unmarshal(respData, &simulateModel); err != nil {
		errDetails := fmt.Sprintf("Unmarshal http response failed: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return simulateModel, err
	}

	// 4. 转换traceModel
	simulateModel, err = tma.commonProcessTraceModel(ctx, simulateModel)
	if err != nil {
		return simulateModel, err
	}

	return simulateModel, nil
}

// 模拟创建链路模型, 若创建成功, 会返回链路模型依赖的数据视图ID
func (tma *traceModelAccess) SimulateUpdateTraceModel(ctx context.Context, modelID string, model interfaces.TraceModel) (newModel interfaces.TraceModel, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 调用DataModel服务模拟修改链路模型", trace.WithSpanKind(trace.SpanKindClient))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 拼接fullPath
	fullPath := fmt.Sprintf("%s/in/v1/simulate-trace-models/%v", tma.appSetting.DataModelUrl, modelID)

	span.SetAttributes(
		attr.Key("request_url").String(fullPath),
		attr.Key("model_id").String(modelID),
	)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	// 2. 发送http request
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		"X-Language":                        rest.GetLanguageByCtx(ctx),
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}
	respCode, respData, err := tma.httpClient.PutNoUnmarshal(ctx, fullPath, headers, model)

	// 3. 分情况处理http response
	simulateModel := interfaces.TraceModel{}
	if err != nil {
		errDetails := fmt.Sprintf("Failed to simulate update trace model by http client: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return simulateModel, err
	}

	if respCode != http.StatusOK {
		errDetails := fmt.Sprintf("Failed to simulate update trace model by http client: %s", string(respData))
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return simulateModel, errors.New(errDetails)
	}

	if err = sonic.Unmarshal(respData, &simulateModel); err != nil {
		errDetails := fmt.Sprintf("Unmarshal http response failed: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return simulateModel, err
	}

	// 4. 转换traceModel
	simulateModel, err = tma.commonProcessTraceModel(ctx, simulateModel)
	if err != nil {
		return simulateModel, err
	}

	return simulateModel, nil
}

/*
	私有方法
*/

// 根据span和related_log的source_type, 做不同的转换处理
func (tma *traceModelAccess) commonProcessTraceModel(ctx context.Context, model interfaces.TraceModel) (newModel interfaces.TraceModel, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 根据source_type转换从DataModel服务获取到的链路模型")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	b, err := sonic.Marshal(model.SpanConfig)
	if err != nil {
		errDetails := fmt.Sprintf("Marshal field span_config failed, err: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return model, err
	}

	switch model.SpanSourceType {
	case interfaces.SOURCE_TYPE_DATA_VIEW:
		spanConf := interfaces.SpanConfigWithDataView{}
		err = sonic.Unmarshal(b, &spanConf)
		if err != nil {
			errDetails := fmt.Sprintf("Field span_config cannot be unmarshaled to SpanConfigWithDataView, err: %v", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return model, err
		}
		model.SpanConfig = spanConf
	case interfaces.SOURCE_TYPE_DATA_CONNECTION:
		spanConf := interfaces.SpanConfigWithDataConnection{}
		err = sonic.Unmarshal(b, &spanConf)
		if err != nil {
			errDetails := fmt.Sprintf("Field span_config cannot be unmarshaled to SpanConfigWithDataConnection, err: %v", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return model, err
		}
		model.SpanConfig = spanConf
	default:
		errDetails := fmt.Sprintf("Invalid span_source_type: %s", model.SpanSourceType)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return model, errors.New(errDetails)
	}

	if model.EnabledRelatedLog == interfaces.RELATED_LOG_OPEN {
		b, err := sonic.Marshal(model.RelatedLogConfig)
		if err != nil {
			errDetails := fmt.Sprintf("Marshal field related_log_config failed, err: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return model, err
		}

		switch model.RelatedLogSourceType {
		case interfaces.SOURCE_TYPE_DATA_VIEW:
			relatedLogConf := interfaces.RelatedLogConfigWithDataView{}
			err = sonic.Unmarshal(b, &relatedLogConf)
			if err != nil {
				errDetails := fmt.Sprintf("Field related_log_config cannot be unmarshaled to RelatedLogConfigWithDataView, err: %v", err.Error())
				logger.Error(errDetails)
				o11y.Error(ctx, errDetails)
				return model, err
			}
			model.RelatedLogConfig = relatedLogConf
		default:
			errDetails := fmt.Sprintf("Invalid related_log_source_type: %s", model.RelatedLogSourceType)
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return model, errors.New(errDetails)
		}
	}

	return model, nil
}
