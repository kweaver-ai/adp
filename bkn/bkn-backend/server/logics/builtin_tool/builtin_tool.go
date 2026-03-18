// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package builtin_tool

import (
	"bkn-backend/interfaces"
	_ "embed"
	"encoding/json"
	"fmt"
)

// function.json 与 logics/function.json 保持同步，用于 embed 嵌入
//
//go:embed function.json
var functionJSON []byte

// BuildRegisterInternalToolReq 从 function.json 构建内置工具注册请求体
func BuildRegisterInternalToolReq() ([]byte, error) {
	var raw map[string]any
	if err := json.Unmarshal(functionJSON, &raw); err != nil {
		return nil, fmt.Errorf("parse function.json: %w", err)
	}

	useRule, _ := raw["use_rule"].(string)
	if useRule == "" {
		useRule = "基于参数和规则进行风险评估"
	}

	funcInput, ok := raw["function_input"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("function_input not found in function.json")
	}

	stripParamIDs(funcInput)

	req := map[string]any{
		"box_id":         interfaces.BuiltinToolBoxID,
		"box_name":       interfaces.BuiltinToolBoxName,
		"box_desc":       useRule,
		"metadata_type":  "function",
		"source":         "internal",
		"config_version": interfaces.BuiltinToolConfigVersion,
		"config_source":  "auto",
		"functions":      []any{funcInput},
	}

	return json.Marshal(req)
}

func stripParamIDs(obj map[string]any) {
	if obj == nil {
		return
	}
	delete(obj, "id")
	for _, v := range obj {
		if m, ok := v.(map[string]any); ok {
			stripParamIDs(m)
		}
		if arr, ok := v.([]any); ok {
			for _, item := range arr {
				if m, ok := item.(map[string]any); ok {
					stripParamIDs(m)
				}
			}
		}
	}
}
