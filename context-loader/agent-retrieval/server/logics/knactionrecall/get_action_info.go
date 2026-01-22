// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knactionrecall

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/common"
	infraErr "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// GetActionInfo 获取行动信息（行动召回）
func (s *knActionRecallServiceImpl) GetActionInfo(ctx context.Context, req *interfaces.KnActionRecallRequest) (*interfaces.KnActionRecallResponse, error) {
	// 1. 参数转换：unique_identity -> unique_identities (数组)
	uniqueIdentities := []map[string]interface{}{req.UniqueIdentity}

	// 2. 调用行动查询接口
	actionsReq := &interfaces.QueryActionsRequest{
		KnID:             req.KnID,
		AtID:             req.AtID,
		UniqueIdentities: uniqueIdentities,
		IncludeTypeInfo:  false, // 不需要类型信息
	}

	actionsResp, err := s.ontologyQuery.QueryActions(ctx, actionsReq)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("[KnActionRecall#GetActionInfo] QueryActions failed, err: %v", err)
		return nil, err
	}

	// 3. 检查返回结果
	if actionsResp.ActionSource == nil {
		s.logger.WithContext(ctx).Warnf("[KnActionRecall#GetActionInfo] ActionSource is nil")
		return &interfaces.KnActionRecallResponse{
			DynamicTools: []interfaces.KnDynamicTool{},
		}, nil
	}

	if len(actionsResp.Actions) == 0 {
		s.logger.WithContext(ctx).Warnf("[KnActionRecall#GetActionInfo] Actions is empty")
		return &interfaces.KnActionRecallResponse{
			DynamicTools: []interfaces.KnDynamicTool{},
		}, nil
	}

	// 4. 检查 action_source.type
	if actionsResp.ActionSource.Type != interfaces.ActionSourceTypeTool && actionsResp.ActionSource.Type != interfaces.ActionSourceTypeMCP {
		s.logger.WithContext(ctx).Warnf("[KnActionRecall#GetActionInfo] Unsupported action_source type: %s", actionsResp.ActionSource.Type)
		return nil, infraErr.DefaultHTTPError(ctx, http.StatusBadRequest,
			fmt.Sprintf("当前仅支持 type=%s 或 %s 的行动源。当前类型: %s",
				interfaces.ActionSourceTypeTool, interfaces.ActionSourceTypeMCP, actionsResp.ActionSource.Type))
	}

	// 5. 仅处理 actions[0]
	firstAction := actionsResp.Actions[0]

	var dynamicTool interfaces.KnDynamicTool

	if actionsResp.ActionSource.Type == interfaces.ActionSourceTypeTool {
		// 6. 获取工具详情
		toolDetailReq := &interfaces.GetToolDetailRequest{
			BoxID:  actionsResp.ActionSource.BoxID,
			ToolID: actionsResp.ActionSource.ToolID,
		}

		toolDetail, err := s.operatorIntegration.GetToolDetail(ctx, toolDetailReq)
		if err != nil {
			s.logger.WithContext(ctx).Errorf("[KnActionRecall#GetActionInfo] GetToolDetail failed, err: %v", err)
			return nil, err
		}

		// 7. 生成 API URL（从配置读取）
		apiURL := fmt.Sprintf("%s://%s:%d/api/agent-operator-integration/internal-v1/tool-box/%s/proxy/%s",
			s.config.OperatorIntegration.PrivateProtocol,
			s.config.OperatorIntegration.PrivateHost,
			s.config.OperatorIntegration.PrivatePort,
			actionsResp.ActionSource.BoxID,
			actionsResp.ActionSource.ToolID)

		// 8. 映射固定参数
		fixedParams := s.mapFixedParams(ctx, firstAction.Parameters, toolDetail.Metadata.APISpec)

		// 9. 转换 Schema 为 OpenAI Function Call 格式
		parameters, err := s.convertSchemaToFunctionCall(ctx, toolDetail.Metadata.APISpec)
		if err != nil {
			s.logger.WithContext(ctx).Errorf("[KnActionRecall#GetActionInfo] ConvertSchema failed, err: %v", err)
			return nil, infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError,
				fmt.Sprintf("Schema 转换失败: %v", err))
		}

		// 10. 构建 KnDynamicTool
		dynamicTool = interfaces.KnDynamicTool{
			Name:            toolDetail.Name,
			Description:     toolDetail.Description,
			Parameters:      parameters,
			APIURL:          apiURL,
			OriginalSchema:  toolDetail.Metadata.APISpec,
			FixedParams:     fixedParams,
			APICallStrategy: interfaces.ResultProcessStrategyKnActionRecall,
		}
	} else {
		// MCP Logic
		mcpReq := &interfaces.GetMCPToolDetailRequest{
			McpID:    actionsResp.ActionSource.McpID,
			ToolName: actionsResp.ActionSource.ToolName,
		}

		toolDetail, err := s.operatorIntegration.GetMCPToolDetail(ctx, mcpReq)
		if err != nil {
			s.logger.WithContext(ctx).Errorf("[KnActionRecall#GetActionInfo] GetMCPToolDetail failed, err: %v", err)
			return nil, err
		}

		// API URL
		apiURL := fmt.Sprintf("http://%s:%d/api/agent-retrieval/in/v1/mcp/proxy/%s/tools/%s/call",
			s.config.Project.Name, // 使用服务名称作为 Host (如 agent-retrieval)
			s.config.Project.Port,
			actionsResp.ActionSource.McpID,
			actionsResp.ActionSource.ToolName)

		// Fixed Params (Flat)
		fixedParams := firstAction.Parameters

		// Schema Conversion
		parameters, err := s.convertMCPSchemaToFunctionCall(ctx, toolDetail.InputSchema)
		if err != nil {
			s.logger.WithContext(ctx).Errorf("[KnActionRecall#GetActionInfo] ConvertMCPSchema failed, err: %v", err)
			return nil, infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("MCP Schema 转换失败: %v", err))
		}

		dynamicTool = interfaces.KnDynamicTool{
			Name:            toolDetail.Name,
			Description:     toolDetail.Description,
			Parameters:      parameters,
			APIURL:          apiURL,
			OriginalSchema:  toolDetail.InputSchema,
			FixedParams:     fixedParams,
			APICallStrategy: interfaces.ResultProcessStrategyKnActionRecall,
		}
	}

	// 11. 构建headers
	headers := common.GetHeaderFromCtx(ctx)

	return &interfaces.KnActionRecallResponse{
		Headers:      headers,
		DynamicTools: []interfaces.KnDynamicTool{dynamicTool},
	}, nil
}
