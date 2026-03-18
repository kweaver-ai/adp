// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package risk_type

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"ontology-query/common"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
	"ontology-query/logics"
)

const (
	RiskLevelSafe     = "safe"
	RiskLevelLow      = "low"
	RiskLevelMedium   = "medium"
	RiskLevelHigh     = "high"
	RiskLevelCritical = "critical"
)

var riskLevelOrder = map[string]int{
	RiskLevelSafe:     1,
	RiskLevelLow:      2,
	RiskLevelMedium:   3,
	RiskLevelHigh:     4,
	RiskLevelCritical: 5,
}

var (
	rtsOnce sync.Once
	rts     interfaces.RiskTypeService
)

type riskTypeService struct {
	appSetting *common.AppSetting
	omAccess   interfaces.OntologyManagerAccess
	aoAccess   interfaces.AgentOperatorAccess
}

func NewRiskTypeService(appSetting *common.AppSetting) interfaces.RiskTypeService {
	rtsOnce.Do(func() {
		rts = &riskTypeService{
			appSetting: appSetting,
			omAccess:   logics.OMA,
			aoAccess:   logics.AOA,
		}
	})
	return rts
}

// Evaluate 对 ActionType 进行风险评估
// 若 risk_type_configs 为空则直接返回 allow；否则获取 RiskType 并评估
func (r *riskTypeService) Evaluate(ctx context.Context, actionType *interfaces.ActionType, knID string, branch string) (*interfaces.RiskTypeEvalResult, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "RiskType.Evaluate", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	if len(actionType.RiskTypeConfigs) == 0 {
		span.SetStatus(codes.Ok, "no risk_type_configs")
		return &interfaces.RiskTypeEvalResult{Allow: true}, nil
	}

	riskTypeIDs := make([]string, 0, len(actionType.RiskTypeConfigs))
	for _, cfg := range actionType.RiskTypeConfigs {
		if cfg.RiskTypeID != "" {
			riskTypeIDs = append(riskTypeIDs, cfg.RiskTypeID)
		}
	}
	if len(riskTypeIDs) == 0 {
		return &interfaces.RiskTypeEvalResult{Allow: true}, nil
	}

	riskTypes, err := r.omAccess.GetRiskTypesByIDs(ctx, knID, branch, riskTypeIDs)
	if err != nil {
		logger.Errorf("GetRiskTypesByIDs failed: %v", err)
		span.SetStatus(codes.Error, "GetRiskTypesByIDs failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
			WithErrorDetails(fmt.Sprintf("get risk types by ids %v failed: %v", riskTypeIDs, err))
	}

	riskTypeMap := make(map[string]*interfaces.RiskType)
	for i := range riskTypes {
		riskTypeMap[riskTypes[i].RTID] = &riskTypes[i]
	}

	// 校验所有 risk_type_id 存在
	for _, cfg := range actionType.RiskTypeConfigs {
		if cfg.RiskTypeID != "" && riskTypeMap[cfg.RiskTypeID] == nil {
			span.SetStatus(codes.Error, "risk type not found")
			return &interfaces.RiskTypeEvalResult{
					Allow:   false,
					Message: fmt.Sprintf("risk_type_id '%s' not found", cfg.RiskTypeID),
				}, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
					WithErrorDetails(fmt.Sprintf("risk_type_id '%s' does not exist", cfg.RiskTypeID))
		}
	}

	// 遍历每个 RiskTypeConfig 进行评估，并与该 RiskType 的 max_acceptable_level 比较
	for _, cfg := range actionType.RiskTypeConfigs {
		rt := riskTypeMap[cfg.RiskTypeID]
		if rt == nil {
			continue
		}
		// arguments 为 action type 调用时传入的实参（对 risk type 中 input 类型参数的实例化）
		var arguments map[string]any
		if cfg.Params != nil {
			arguments = cfg.Params
		} else {
			arguments = make(map[string]any)
			for _, param := range cfg.Parameters {
				arguments[param.Name] = param.Value
			}
		}
		// 风险类形参定义（rt.Parameters）与 arguments 合并：常量取定义值，输入取 arguments，得到当前调用的实参
		actualParams := mergeRiskTypeParams(rt.Parameters, arguments)

		// 通过内置风险评估工具执行评估
		level, err := r.executeRiskAssessmentTool(ctx, rt, actualParams)
		if err != nil {
			logger.Errorf("executeRiskAssessmentTool failed for risk_type %s: %v", rt.RTID, err)
			span.SetStatus(codes.Error, "risk assessment tool execution failed")
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
				WithErrorDetails(err.Error())
		}
		maxAcceptable := rt.MaxAcceptableLevel
		if maxAcceptable == "" {
			maxAcceptable = RiskLevelCritical
		}
		if riskLevelOrder[level] > riskLevelOrder[maxAcceptable] {
			span.SetStatus(codes.Error, "risk level exceeds max_acceptable_level")
			return &interfaces.RiskTypeEvalResult{
					Allow:   false,
					Message: fmt.Sprintf("risk evaluation failed: level %s exceeds max_acceptable_level %s for risk_type %s", level, maxAcceptable, rt.RTID),
				}, rest.NewHTTPError(ctx, http.StatusForbidden, oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
					WithErrorDetails(fmt.Sprintf("risk level %s exceeds max_acceptable_level %s", level, maxAcceptable))
		}
	}

	span.SetStatus(codes.Ok, "")
	return &interfaces.RiskTypeEvalResult{Allow: true}, nil
}

// mergeRiskTypeParams 将风险类形参定义与 action type 传入的实参合并，得到当前调用的实参
// 形参中 ValueFrom=const 的取 param.Value，ValueFrom=input 或空的取 arguments 中对应值
func mergeRiskTypeParams(paramDefs []interfaces.Parameter, arguments map[string]any) map[string]any {
	if len(paramDefs) == 0 {
		if arguments == nil {
			return make(map[string]any)
		}
		return arguments
	}
	result := make(map[string]any, len(paramDefs))
	for _, param := range paramDefs {
		switch param.ValueFrom {
		case interfaces.LOGIC_PARAMS_VALUE_FROM_CONST:
			result[param.Name] = param.Value
		case interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT, "":
			if arguments != nil {
				if v, ok := arguments[param.Name]; ok {
					result[param.Name] = v
				}
			}
		default:
			// property 等其它来源在 Evaluate 场景无 objectData，从 arguments 取
			if arguments != nil {
				if v, ok := arguments[param.Name]; ok {
					result[param.Name] = v
				}
			}
		}
	}
	return result
}

// convertToParamDefs 将 []Parameter 转为内置工具所需的 param_defs 格式，每项仅保留 name、type
func convertToParamDefs(params []interfaces.Parameter) []map[string]any {
	if len(params) == 0 {
		return nil
	}
	result := make([]map[string]any, 0, len(params))
	for _, p := range params {
		result = append(result, map[string]any{"name": p.Name, "type": p.Type})
	}
	return result
}

// buildRiskFunctionToolRequest 按 RiskFunction.Parameters 的 Source 组装 ToolExecutionRequest
// 统一合约：Body 包含 rules、param_defs、arguments；自定义 tool 在此基础上按 Source 组装 path/query/header/body
func buildRiskFunctionToolRequest(rt *interfaces.RiskType, actualParams map[string]any) interfaces.ToolExecutionRequest {
	request := interfaces.ToolExecutionRequest{
		Header: map[string]any{},
		Query:  map[string]any{},
		Body: map[string]any{
			"arguments":  actualParams,
			"rules":      rt.RiskRules,
			"param_defs": convertToParamDefs(rt.Parameters),
		},
		Path:    map[string]any{},
		Timeout: 300,
	}

	if rt.RiskFunction == nil || len(rt.RiskFunction.Parameters) == 0 {
		return request
	}

	for _, param := range rt.RiskFunction.Parameters {
		var value any
		switch strings.ToLower(param.ValueFrom) {
		case interfaces.LOGIC_PARAMS_VALUE_FROM_CONST:
			value = param.Value
		case interfaces.LOGIC_PARAMS_VALUE_FROM_PARAM, interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT, "":
			if refName, ok := param.Value.(string); ok && actualParams != nil {
				value = actualParams[refName]
			}
		default:
			if actualParams != nil {
				if refName, ok := param.Value.(string); ok {
					value = actualParams[refName]
				}
			}
		}
		if value == nil {
			continue
		}

		source := strings.ToLower(param.Source)
		switch source {
		case interfaces.PARAMETER_HEADER:
			setNestedValueForRisk(request.Header, param.Name, value)
		case interfaces.PARAMETER_QUERY:
			setNestedValueForRisk(request.Query, param.Name, value)
		case interfaces.PARAMETER_PATH:
			setNestedValueForRisk(request.Path, param.Name, value)
		case interfaces.PARAMETER_BODY, "":
			setNestedValueForRisk(request.Body, param.Name, value)
		default:
			setNestedValueForRisk(request.Body, param.Name, value)
		}
	}

	return request
}

func setNestedValueForRisk(target map[string]any, key string, value any) {
	if value == nil {
		return
	}
	if strings.Contains(key, ".") {
		parts := strings.Split(key, ".")
		current := target
		for i, part := range parts {
			if i == len(parts)-1 {
				current[part] = value
				return
			}
			if _, exists := current[part]; !exists {
				current[part] = make(map[string]any)
			}
			if next, ok := current[part].(map[string]any); ok {
				current = next
			} else {
				current[part] = make(map[string]any)
				current = current[part].(map[string]any)
			}
		}
		return
	}
	target[key] = value
}

// executeRiskAssessmentTool 通过 aoAccess.ExecuteTool 调用风险评估工具，返回风险级别
// 内置工具与自定义 tool 统一接收 rules、param_defs、arguments；自定义 tool 在此基础上按 RiskFunction.Parameters 的 Source 组装 path/query/header/body
func (r *riskTypeService) executeRiskAssessmentTool(ctx context.Context, rt *interfaces.RiskType, actualParams map[string]any) (string, error) {
	if r.aoAccess == nil {
		return "", fmt.Errorf("AgentOperatorAccess not configured")
	}
	boxID := interfaces.BuiltinToolBoxID
	toolID := interfaces.BuiltinToolToolID
	if rt.RiskFunction != nil && rt.RiskFunction.BoxID != "" && rt.RiskFunction.ToolID != "" {
		boxID = rt.RiskFunction.BoxID
		toolID = rt.RiskFunction.ToolID
	}

	execRequest := buildRiskFunctionToolRequest(rt, actualParams)

	result, err := r.aoAccess.ExecuteTool(ctx, boxID, toolID, execRequest)
	if err != nil {
		logger.Errorf("ExecuteTool failed for risk_type %s: %v", rt.RTID, err)
		return "", fmt.Errorf("risk assessment tool execution failed: %w", err)
	}

	// 解析返回体：{ risk_level, success, message }
	resMap, ok := result.(map[string]any)
	if !ok {
		return "", fmt.Errorf("risk assessment tool returned invalid result type")
	}
	if success, _ := resMap["result"].(map[string]any)["success"].(bool); !success {
		msg, _ := resMap["result"].(map[string]any)["message"].(string)
		if msg == "" {
			msg, _ = resMap["error"].(string)
		}
		logger.Warnf("Risk assessment tool returned success=false for risk_type %s: %s", rt.RTID, msg)
		return RiskLevelCritical, nil
	}

	level, _ := resMap["result"].(map[string]any)["risk_level"].(string)
	if level == "" {
		return RiskLevelCritical, nil
	}
	return level, nil
}

// MustAllow 若 RiskType 返回 disallow 则返回错误
func (r *riskTypeService) MustAllow(ctx context.Context, actionType *interfaces.ActionType, knID string, branch string) error {
	result, err := r.Evaluate(ctx, actionType, knID, branch)
	if err != nil {
		return err
	}
	if !result.Allow {
		return rest.NewHTTPError(ctx, http.StatusForbidden, oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
			WithErrorDetails(result.Message)
	}
	return nil
}
