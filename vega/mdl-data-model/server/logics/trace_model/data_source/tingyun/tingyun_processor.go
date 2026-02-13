// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package tingyun

import (
	"context"
	"sync"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"go.opentelemetry.io/otel/codes"

	"data-model/common"
	"data-model/interfaces"
	dtype "data-model/interfaces/data_type"
)

var (
	tytProcessorOnce sync.Once
	tytProcessor     interfaces.TraceModelProcessor

	// 听云span字段信息
	tySpanFieldInfo = []interfaces.TraceModelField{
		{
			Name: "actionAlias",
			Type: dtype.DataType_Text,
		},
		{
			Name: "actionGuid",
			Type: dtype.DataType_Text,
		},
		{
			Name: "actionId",
			Type: dtype.DataType_Float,
		},
		{
			Name: "actionName",
			Type: dtype.DataType_Text,
		},
		{
			Name: "actionTraceFlag",
			Type: dtype.DataType_Float,
		},
		{
			Name: "actionType",
			Type: dtype.DataType_Text,
		},
		{
			Name: "applicationId",
			Type: dtype.DataType_Float,
		},
		{
			Name: "applicationName",
			Type: dtype.DataType_Text,
		},
		{
			Name: "bizSystemId",
			Type: dtype.DataType_Float,
		},
		{
			Name: "bizSystemName",
			Type: dtype.DataType_Text,
		},
		{
			Name: "callerApplicationId",
			Type: dtype.DataType_Float,
		},
		{
			Name: "callerBizSystemId",
			Type: dtype.DataType_Float,
		},
		{
			Name: "callerInstanceId",
			Type: dtype.DataType_Float,
		},
		{
			Name: "durationUs",
			Type: dtype.DataType_Float,
		},
		{
			Name: "errorCount",
			Type: dtype.DataType_Float,
		},
		{
			Name: "errorFnNo",
			Type: dtype.DataType_Float,
		},
		{
			Name: "exclusiveTime",
			Type: dtype.DataType_Float,
		},
		{
			Name: "hasError",
			Type: dtype.DataType_Float,
		},
		{
			Name: "id",
			Type: dtype.DataType_Text,
		},
		{
			Name: "instanceId",
			Type: dtype.DataType_Text,
		},
		{
			Name: "instanceName",
			Type: dtype.DataType_Text,
		},
		{
			Name: "isSlowTrace",
			Type: dtype.DataType_Boolean,
		},
		{
			Name: "isSnapshot",
			Type: dtype.DataType_Boolean,
		},
		{
			Name: "requestId",
			Type: dtype.DataType_Text,
		},
		{
			Name: "respTime",
			Type: dtype.DataType_Float,
		},
		{
			Name: "status",
			Type: dtype.DataType_Boolean,
		},
		{
			Name: "timeAccuracy",
			Type: dtype.DataType_Float,
		},
		{
			Name: "timestamp",
			Type: dtype.DataType_Float,
		},
		{
			Name: "traceGuid",
			Type: dtype.DataType_Text,
		},
		{
			Name: "uri",
			Type: dtype.DataType_Text,
		},
		{
			Name: "userId",
			Type: dtype.DataType_Text,
		},
	}
)

type tingYunTraceProcessor struct {
	appSetting *common.AppSetting
}

func NewTingYunTraceProcessor(appSetting *common.AppSetting) interfaces.TraceModelProcessor {
	tytProcessorOnce.Do(func() {
		tytProcessor = &tingYunTraceProcessor{
			appSetting: appSetting,
		}
	})
	return tytProcessor
}

func (tytp *tingYunTraceProcessor) GetSpanFieldInfo(ctx context.Context,
	model interfaces.TraceModel) (fieldInfos []interfaces.TraceModelField, err error) {

	_, span := ar_trace.Tracer.Start(ctx, "logic层: 查询Span字段信息")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	return append(interfaces.SPAN_METADATA, tySpanFieldInfo...), nil
}

func (tytp *tingYunTraceProcessor) GetRelatedLogFieldInfo(ctx context.Context, model interfaces.TraceModel) (fieldInfos []interfaces.TraceModelField, err error) {
	_, span := ar_trace.Tracer.Start(ctx, "logic层: 查询Span关联日志字段信息")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	return []interfaces.TraceModelField(nil), nil
}
