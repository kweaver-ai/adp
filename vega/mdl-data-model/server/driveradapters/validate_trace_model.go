// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"

	derrors "data-model/errors"
	"data-model/interfaces"
	dcond "data-model/interfaces/condition"
)

/*
	链路模型模块的校验函数. 包括:
		(1)  创建/导入时的链路模型校验
		(2)  修改时的链路模型校验
		(3)  创建/导入/修改时的链路校验(名称)
		(4)  创建/导入/修改时的链路校验(除名称以外的其它参数)
		(5)  创建/导入/修改时的跨度校验
		(6)  创建/导入/修改时的跨度基础属性校验
		(7)  创建/导入/修改时的关联日志配置校验
		(8)  创建/导入/修改时的时间格式校验
		(9)  创建/导入/修改时的耗时单位校验
		(10) 创建/导入/修改时的前置条件校验
		(11) 创建/导入/修改时的前置条件中value_from字段校验
		(12) span数据来源校验
*/

// 链路模型校验函数(1): 创建/导入时的链路模型校验
func validateTraceModelsWhenCreate(ctx context.Context, r *restHandler, reqModels []interfaces.TraceModel) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driver层: 校验传入的链路模型")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 校验链路模型(名称)
	err = validateTraceModelsWithName(ctx, r, reqModels)
	if err != nil {
		return err
	}

	// 2. 校验链路模型(除名称外的其它参数)
	err = validateTraceModelsWithoutName(ctx, reqModels)
	if err != nil {
		return err
	}

	return nil
}

// 链路模型校验函数(2): 修改时的链路模型校验
func validateTraceModelWhenUpdate(ctx context.Context, r *restHandler, preName string, reqModel interfaces.TraceModel) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driver层: 校验传入的链路模型")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	reqModels := []interfaces.TraceModel{reqModel}

	// 待修改的链路模型名称与之前不一样, 则需进行名称的合法性校验
	if preName != reqModel.Name {
		// 1. 校验链路模型(名称)
		err := validateTraceModelsWithName(ctx, r, reqModels)
		if err != nil {
			return err
		}
	}

	// 2. 校验链路模型(除名称外的其它参数)
	err = validateTraceModelsWithoutName(ctx, reqModels)
	if err != nil {
		return err
	}

	return nil
}

// 链路模型校验函数(3): 创建/导入/修改时的链路校验(名称)
func validateTraceModelsWithName(ctx context.Context, r *restHandler, reqModels []interfaces.TraceModel) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driver层: 校验所有链路模型名称")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 初始化参数
	// 统计同批次所有链路模型的名称, 用于重复性校验
	modelNameStat := make(map[string]bool)
	// 存储同批次所有链路模型的名称, 供后续校验在数据库中的存在性时使用
	modelNames := make([]string, 0)

	// 2. 校验每一个链路模型
	for _, reqModel := range reqModels {
		// 2.1 校验链路模型名称是否有效
		err := validateObjectName(ctx, reqModel.Name, interfaces.TRACE_MODEL)
		if err != nil {
			o11y.Error(ctx, err.Error())
			return err
		}

		// 2.2 校验是否与同批次其它待创建/导入的链路模型重复
		if _, ok := modelNameStat[reqModel.Name]; ok {
			errDetails := fmt.Sprintf("The trace model named %v is not unique among all the trace models to be created or imported", reqModel.Name)
			o11y.Error(ctx, errDetails)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_NotUniqueInBatch_ModelName).
				WithErrorDetails(errDetails)
		}

		// 2.3 更新modelNameStat与modelNames
		modelNameStat[reqModel.Name] = true
		modelNames = append(modelNames, reqModel.Name)
	}

	// 3. 检查链路模型名称在数据库中的存在性
	modelMap, err := r.tms.GetSimpleTraceModelMapByNames(ctx, modelNames)
	if err != nil {
		return err
	}

	if length := len(modelMap); length != 0 {
		modelNames := make([]string, 0, length)
		for name := range modelMap {
			modelNames = append(modelNames, name)
		}

		errDetails := fmt.Sprintf("The trace model whose name is in %v already exist in the database!", modelNames)
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_TraceModel_ModelNameExisted).
			WithErrorDetails(errDetails)
	}

	return nil
}

// 链路模型校验函数(4): 创建/导入/修改时的链路校验(除名称以外的其它参数)
func validateTraceModelsWithoutName(ctx context.Context, reqModels []interfaces.TraceModel) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driver层: 校验所有链路模型除名称外的其它参数")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	for i := range reqModels {
		// 1. 校验tags
		err := validateObjectTags(ctx, reqModels[i].Tags)
		if err != nil {
			o11y.Error(ctx, err.Error())
			return err
		}

		// 2. 校验comment
		err = validateObjectComment(ctx, reqModels[i].Comment)
		if err != nil {
			o11y.Error(ctx, err.Error())
			return err
		}

		// 3. 校验spanConfig
		err = validateSpanConfig(ctx, reqModels[i].SpanConfig, reqModels[i].SpanSourceType)
		if err != nil {
			o11y.Error(ctx, err.Error())
			return err
		}

		// 判断是否开启关联日志的配置
		if reqModels[i].EnabledRelatedLog == interfaces.RELATED_LOG_OPEN {
			// 4. 校验relatedLogConfig
			err = validateRelatedLogConfig(ctx, reqModels[i].RelatedLogConfig, reqModels[i].RelatedLogSourceType)
			if err != nil {
				o11y.Error(ctx, err.Error())
				return err
			}
		} else {
			// 如果没有开启, 需要处理脏数据, 避免数据库操作问题
			reqModels[i].RelatedLogSourceType = ""
			reqModels[i].RelatedLogConfig = struct{}{}
		}
	}

	return nil
}

// 链路模型校验函数(5): 创建/导入/修改时的跨度校验
func validateSpanConfig(ctx context.Context, spanConf any, spanSourceType string) error {
	if spanSourceType == interfaces.SOURCE_TYPE_DATA_VIEW {
		conf := spanConf.(interfaces.SpanConfigWithDataView)
		// 校验span所在数据视图名称
		err := validateObjectName(ctx, conf.DataView.Name, interfaces.MODULE_TYPE_DATA_VIEW)
		if err != nil {
			return err
		}

		// 校验span的basic_attributes中的必填配置
		err = validateSpanBasicAttrs(ctx, conf)
		if err != nil {
			return err
		}
	} else {
		conf := spanConf.(interfaces.SpanConfigWithDataConnection)
		// 校验数据连接名称
		err := validateObjectName(ctx, conf.DataConnection.Name, interfaces.DATA_CONNECTION)
		if err != nil {
			return err
		}
	}

	return nil
}

// 链路模型校验函数(6): 创建/导入/修改时的跨度基础属性校验
func validateSpanBasicAttrs(ctx context.Context, spanConf interfaces.SpanConfigWithDataView) error {
	/*
		part1. 必填字段校验
		包括: TraceID, SpanID, ParentSpanID, Name, StartTime, ServiceName
	*/

	// 1.1 TraceID校验
	if spanConf.TraceID.FieldName == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_TraceID).
			WithErrorDetails("The field name is empty")
	}

	// 1.2 SpanID校验
	fieldNames := spanConf.SpanID.FieldNames
	if len(fieldNames) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_SpanID).
			WithErrorDetails("The field name array is empty")
	}

	m := make(map[string]struct{})
	for _, fieldName := range fieldNames {
		if fieldName == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_SpanID).
				WithErrorDetails("Some field names in the field name array are empty")
		}

		if _, ok := m[fieldName]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_SpanID).
				WithErrorDetails("Field names in field_names must be unique")
		}
		m[fieldName] = struct{}{}
	}

	// 1.3 ParentSpanID校验
	configs := spanConf.ParentSpanID
	if len(configs) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ParentSpanID).
			WithErrorDetails("The configuration array is empty")
	}

	for _, config := range configs {
		// 校验前置条件
		if err := validPrecond(config.Precond); err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ParentSpanID).
				WithErrorDetails(fmt.Sprintf("Condition err: %v", err))
		}

		// 校验FieldNames
		if len(config.FieldNames) == 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ParentSpanID).
				WithErrorDetails("The field name array is empty")
		}

		m := make(map[string]struct{})
		for _, fieldName := range config.FieldNames {
			if fieldName == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ParentSpanID).
					WithErrorDetails("Some field names in the field name array are empty")
			}

			if _, ok := m[fieldName]; ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ParentSpanID).
					WithErrorDetails("Field names in field_names must be unique")
			}
			m[fieldName] = struct{}{}
		}
	}

	// 1.4 Name校验
	if spanConf.Name.FieldName == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Name).
			WithErrorDetails("The field name is empty")
	}

	// 1.5 StartTime校验
	startTimeConf := spanConf.StartTime
	if startTimeConf.FieldName == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_StartTime).
			WithErrorDetails("The field name is empty")
	}

	if !isValidTimeFormat(startTimeConf.FieldFormat) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_StartTime).
			WithErrorDetails(fmt.Sprintf("The time format is invalid, valid time format is in %v", interfaces.VALID_TIME_FORMATS))
	}

	// 1.6 ServiceName校验
	if spanConf.ServiceName.FieldName == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ServiceName).
			WithErrorDetails("The field name is empty")
	}

	/*
		part2. 非必填字段校验
		包括: EndTime, Duration, Kind, Status.
		其中: SpanKind和SpanStatus的FieldName可为空, 无需校验;
		而EndTime的FieldName不为空时, 其FieldFormat要合法,
		Duration的FieldName不为空时, 其Unit要合法,
		EndTime和Duration两者有且仅能配置一个.
	*/

	// 2.1 EndTime校验
	endTimeConf := spanConf.EndTime
	if endTimeConf.FieldName != "" && !isValidTimeFormat(endTimeConf.FieldFormat) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_EndTime).
			WithErrorDetails(fmt.Sprintf("The time format is invalid, valid time format is in %v", interfaces.VALID_TIME_FORMATS))
	}

	// 2.2 Duration校验
	durationConf := spanConf.Duration
	if durationConf.FieldName != "" && !isValidDurationUnit(durationConf.FieldUnit) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Duration).
			WithErrorDetails(fmt.Sprintf("The duration unit is invalid, valid duration unit is in %v", interfaces.VALID_DURATION_UNITS))
	}

	// 同时配置EndTime和Duration, 报错
	if endTimeConf.FieldName != "" && durationConf.FieldName != "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig).
			WithErrorDetails("Both end_time and duration are configured")
	}

	// 同时未配置EndTime和Duration, 报错
	if endTimeConf.FieldName == "" && durationConf.FieldName == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig).
			WithErrorDetails("Both end_time and duration are not configured")
	}

	return nil
}

// 链路模型校验函数(7): 创建/导入/修改时的关联日志配置校验
func validateRelatedLogConfig(ctx context.Context, relatedLogConf any, sourceType string) error {
	if sourceType == interfaces.SOURCE_TYPE_DATA_VIEW {
		conf := relatedLogConf.(interfaces.RelatedLogConfigWithDataView)

		// 1. 校验关联日志所在数据视图名称
		err := validateObjectName(ctx, conf.DataView.Name, interfaces.MODULE_TYPE_DATA_VIEW)
		if err != nil {
			return err
		}

		// 2. TraceID校验
		if conf.TraceID.FieldName == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidRelatedLogConfig_TraceID).
				WithErrorDetails("The field name is empty")
		}

		// 3. SpanID校验
		fieldNames := conf.SpanID.FieldNames
		if len(fieldNames) == 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidRelatedLogConfig_SpanID).
				WithErrorDetails("The field name array is empty")
		}

		m := make(map[string]struct{})
		for _, fieldName := range fieldNames {
			if fieldName == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidRelatedLogConfig_SpanID).
					WithErrorDetails("Some field names in the field name array are empty")
			}

			if _, ok := m[fieldName]; ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidRelatedLogConfig_SpanID).
					WithErrorDetails("Field names in field_names must be unique")
			}
			m[fieldName] = struct{}{}
		}
	}
	return nil
}

// 链路模型校验函数(8): 创建/导入/修改时的时间格式校验
func isValidTimeFormat(format string) bool {
	for _, validFormat := range interfaces.VALID_TIME_FORMATS {
		if format == validFormat {
			return true
		}
	}
	return false
}

// 链路模型校验函数(9): 创建/导入/修改时的耗时单位校验
func isValidDurationUnit(unit string) bool {
	for _, validUnit := range interfaces.VALID_DURATION_UNITS {
		if unit == validUnit {
			return true
		}
	}
	return false
}

// 链路模型校验函数(10): 创建/导入/修改时的前置条件校验
func validPrecond(precond *interfaces.CondCfg) error {
	// precond == nil, 认为没配置前置条件
	if precond == nil {
		return nil
	}

	switch precond.Operation {
	case dcond.OperationAnd, dcond.OperationOr:
		for _, subCond := range precond.SubConds {
			if err := validPrecond(subCond); err != nil {
				return err
			}
		}
		return nil
	case dcond.Operation_RANGE:
		if precond.Name == "" {
			return errors.New("name is empty")
		}

		if precond.ValueFrom != dcond.ValueFrom_Const {
			return fmt.Errorf("invalid value_from, valid value_from is %v", dcond.ValueFrom_Const)
		}
		return nil
	case dcond.OperationEq, dcond.OperationNotEq:
		if precond.Name == "" {
			return errors.New("name is empty")
		}

		if !isValidValueFrom(precond.ValueFrom) {
			return fmt.Errorf("invalid value_from, valid value_from is in %v", interfaces.VALID_PRECONDITION_VALUE_FROM)
		}

		if precond.ValueFrom == dcond.ValueFrom_Field {
			if valueStr, ok := precond.Value.(string); !ok || valueStr == "" {
				return errors.New("value_from is field, but value is invalid")
			}
		}
		return nil
	default:
		return fmt.Errorf("invalid operation, valid operation is in %v", interfaces.VALID_PRECONDITION_OPERATIONS)
	}
}

// 链路模型校验函数(11): 创建/导入/修改时的前置条件中value_from字段校验
func isValidValueFrom(valueFrom string) bool {
	for _, validValueFrom := range interfaces.VALID_PRECONDITION_VALUE_FROM {
		if valueFrom == validValueFrom {
			return true
		}
	}
	return false
}

// 链路模型校验函数(12): span数据来源校验
func validateSpanSourceType(ctx context.Context, spanSourceType string) error {
	switch spanSourceType {
	case "data_view", "data_connection":
		return nil
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidParameter_SpanSourceType).
			WithErrorDetails("span_source_type is invalid, valid span_source_type is data_view or data_connection")
	}
}
