// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

// ValidateRiskTypes 校验风险类创建请求
func ValidateRiskTypes(ctx context.Context, knID string, riskTypes []*interfaces.RiskType) error {
	tmpNameMap := make(map[string]any)
	idMap := make(map[string]any)
	for i := range riskTypes {
		riskType := riskTypes[i]
		if riskType.ModuleType != "" && riskType.ModuleType != interfaces.MODULE_TYPE_RISK_TYPE {
			return rest.NewHTTPError(ctx, http.StatusForbidden, berrors.BknBackend_InvalidParameter_ModuleType).
				WithErrorDetails("Risk type module type is not 'risk_type'")
		}

		rtID := riskType.RTID
		if _, ok := idMap[rtID]; !ok || rtID == "" {
			idMap[rtID] = nil
		} else {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RiskType_RiskTypeIDExisted).
				WithDescription(map[string]any{"riskTypeID": rtID}).
				WithErrorDetails(fmt.Sprintf("RiskType ID '%s' already exists in the request body", rtID))
		}

		err := ValidateRiskType(ctx, riskType)
		if err != nil {
			return err
		}

		if _, ok := tmpNameMap[riskType.RTName]; !ok {
			tmpNameMap[riskType.RTName] = nil
		} else {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RiskType_RiskTypeNameExisted)
		}

		riskType.KNID = knID
	}
	return nil
}

// ValidateRiskType 校验单个风险类
func ValidateRiskType(ctx context.Context, riskType *interfaces.RiskType) error {
	err := validateID(ctx, riskType.RTID)
	if err != nil {
		return err
	}

	riskType.RTName = strings.TrimSpace(riskType.RTName)
	err = validateObjectName(ctx, riskType.RTName, interfaces.MODULE_TYPE_RISK_TYPE)
	if err != nil {
		return err
	}

	if err = ValidateTags(ctx, riskType.Tags); err != nil {
		return err
	}
	riskType.Tags = libCommon.TagSliceTransform(riskType.Tags)

	// 若 risk function 未绑定，则绑定内置风险评估工具，使用内置风险评估工具无需绑定参数，由系统处理,在risk type上定义的参数都会把实参传给工具
	if riskType.RiskFunction == nil {
		riskType.RiskFunction = &interfaces.RiskFunction{
			Type:   "tool",
			BoxID:  interfaces.BuiltinToolBoxID,
			ToolID: interfaces.BuiltinToolToolID,
		}
	}
	if riskType.RiskFunction.BoxID == "" || riskType.RiskFunction.ToolID == "" {
		riskType.RiskFunction.BoxID = interfaces.BuiltinToolBoxID
		riskType.RiskFunction.ToolID = interfaces.BuiltinToolToolID
	}

	if err = validateObjectComment(ctx, riskType.Comment); err != nil {
		return err
	}

	// 校验 max_acceptable_level
	if riskType.MaxAcceptableLevel != "" {
		if _, ok := interfaces.RiskLevelOrder[riskType.MaxAcceptableLevel]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RiskType_InvalidMaxAcceptableLevel).
				WithErrorDetails(fmt.Sprintf("max_acceptable_level must be one of [safe, low, medium, high, critical], got [%s]",
					riskType.MaxAcceptableLevel))
		}
	}

	// 校验 parameters 唯一性
	paramNames := make(map[string]bool)
	for _, p := range riskType.Parameters {
		if p.Name == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RiskType_InvalidParameter).
				WithErrorDetails("RiskType parameter name cannot be empty")
		}
		if paramNames[p.Name] {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RiskType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("Duplicate parameter name: %s", p.Name))
		}
		paramNames[p.Name] = true
	}

	// 校验 RiskFunction.Parameters
	if riskType.RiskFunction != nil && len(riskType.RiskFunction.Parameters) > 0 {
		if err := validateRiskFunctionParameters(ctx, riskType.RiskFunction.Parameters, paramNames); err != nil {
			return err
		}
	}

	return nil
}

// validateRiskFunctionParameters 校验 RiskFunction 的扁平化参数
func validateRiskFunctionParameters(ctx context.Context, params []interfaces.Parameter, riskTypeParamNames map[string]bool) error {
	validSources := map[string]bool{"path": true, "query": true, "header": true, "body": true}
	for i, p := range params {
		source := strings.ToLower(p.Source)
		if source != "" && !validSources[source] {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RiskType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("risk_function.parameters[%d].source must be one of [path, query, header, body], got [%s]", i, p.Source))
		}
		switch strings.ToLower(p.ValueFrom) {
		case interfaces.VALUE_FROM_CONST:
			// value 可为任意类型，无需额外校验
		case interfaces.VALUE_FROM_PARAM, "":
			refName, ok := p.Value.(string)
			if !ok || refName == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RiskType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("risk_function.parameters[%d]: value_from=param requires value to be non-empty string (ParamDef.name)", i))
			}
			if !riskTypeParamNames[refName] {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RiskType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("risk_function.parameters[%d]: value_from=param references unknown RiskType parameter [%s]", i, refName))
			}
		default:
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RiskType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("risk_function.parameters[%d].value_from must be one of [const, param], got [%s]", i, p.ValueFrom))
		}
	}
	return nil
}
