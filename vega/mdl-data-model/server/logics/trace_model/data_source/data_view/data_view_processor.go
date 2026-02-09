// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_view

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	"data-model/logics/data_view"
)

var (
	dvtProcessorOnce sync.Once
	dvtProcessor     interfaces.TraceModelProcessor
)

type dataViewTraceProcessor struct {
	appSetting *common.AppSetting
	dvService  interfaces.DataViewService
}

func NewDataViewTraceProcessor(appSetting *common.AppSetting) interfaces.TraceModelProcessor {
	dvtProcessorOnce.Do(func() {
		dvtProcessor = &dataViewTraceProcessor{
			appSetting: appSetting,
			dvService:  data_view.NewDataViewService(appSetting),
		}
	})
	return dvtProcessor
}

func (dvtp *dataViewTraceProcessor) GetSpanFieldInfo(ctx context.Context, model interfaces.TraceModel) (fieldInfos []interfaces.TraceModelField, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 查询Span字段信息")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	spanConf, _ := model.SpanConfig.(interfaces.SpanConfigWithDataView)
	views, err := dvtp.dvService.GetDataViews(ctx, []string{spanConf.DataView.ID}, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		if httpErr.HTTPCode == http.StatusNotFound {
			errDetails := fmt.Sprintf("The data view whose id equal to %v was not found", spanConf.DataView.ID)
			logger.Errorf(errDetails)
			o11y.Error(ctx, errDetails)
			return fieldInfos, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_TraceModel_DependentDataViewNotFound).
				WithErrorDetails(errDetails)
		}
		return fieldInfos, err
	}

	fieldInfos = interfaces.SPAN_METADATA
	for _, field := range views[0].Fields {
		fieldInfos = append(fieldInfos, interfaces.TraceModelField{
			Name: field.Name,
			Type: field.Type,
		})
	}

	return fieldInfos, nil
}

func (dvtp *dataViewTraceProcessor) GetRelatedLogFieldInfo(ctx context.Context, model interfaces.TraceModel) (fieldInfos []interfaces.TraceModelField, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 查询Span关联日志字段信息")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	relatedLogConf, _ := model.RelatedLogConfig.(interfaces.RelatedLogConfigWithDataView)
	views, err := dvtp.dvService.GetDataViews(ctx, []string{relatedLogConf.DataView.ID}, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		if httpErr.HTTPCode == http.StatusNotFound {
			errDetails := fmt.Sprintf("The data view whose id equal to %v was not found", relatedLogConf.DataView.ID)
			logger.Errorf(errDetails)
			o11y.Error(ctx, errDetails)
			return fieldInfos, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_TraceModel_DependentDataViewNotFound).
				WithErrorDetails(errDetails)
		}
		return fieldInfos, err
	}

	fieldInfos = interfaces.RELATED_LOG_METADATA
	for _, field := range views[0].Fields {
		fieldInfos = append(fieldInfos, interfaces.TraceModelField{
			Name: field.Name,
			Type: field.Type,
		})
	}

	return fieldInfos, nil
}
