package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	oerrors "ontology-manager/errors"
	"ontology-manager/interfaces"
)

// 对象类必要创建参数的非空校验。
func ValidateActionType(ctx context.Context, actionType *interfaces.ActionType) error {
	// 校验id的合法性
	err := validateID(ctx, actionType.ATID)
	if err != nil {
		return err
	}

	// 校验名称合法性
	// 去掉名称的前后空格
	actionType.ATName = strings.TrimSpace(actionType.ATName)
	err = validateObjectName(ctx, actionType.ATName, interfaces.MODULE_TYPE_ACTION_TYPE)
	if err != nil {
		return err
	}

	// 若输入了 tags，校验 tags 的合法性
	err = ValidateTags(ctx, actionType.Tags)
	if err != nil {
		return err
	}

	// 去掉tag前后空格以及数组去重
	actionType.Tags = libCommon.TagSliceTransform(actionType.Tags)

	// 校验comment合法性
	err = validateObjectComment(ctx, actionType.Comment)
	if err != nil {
		return err
	}

	// 校验类型
	if actionType.ActionSource.Type != "" {
		// type 不为空，则代表在配置映射了，则需要校验映射
		if !interfaces.IsValidActionSourceType(actionType.ActionSource.Type) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyManager_ActionType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("The type of action source is expected one of [tool, map], actual is [%s]",
					actionType.ActionSource.Type))
		}
		switch actionType.ActionSource.Type {
		case interfaces.ACTION_TYPE_TOOL:
			// tool 时，mcp_id或者tool_name不为空，则报错
			if actionType.ActionSource.McpID != "" || actionType.ActionSource.ToolName != "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyManager_ActionType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("tool type should not have mcp data, current mcp_id is[%s], tool_name is [%s]",
						actionType.ActionSource.McpID, actionType.ActionSource.ToolName))
			}
		case interfaces.ACTION_TYPE_MCP:
			// map 时，box_id或者tool_id不为空，则报错
			if actionType.ActionSource.BoxID != "" || actionType.ActionSource.ToolID != "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyManager_ActionType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("mcp type should not have tool data, current box_id is[%s], tool_id is [%s]",
						actionType.ActionSource.BoxID, actionType.ActionSource.ToolID))
			}
		}
	}

	// parameters 非空时：参数名称非空
	if len(actionType.Parameters) > 0 {
		for _, param := range actionType.Parameters {
			if param.Name == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyManager_ActionType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("行动类[%s]行动资源参数名称不能为空", actionType.ATName))
			}
		}
	}

	return nil
}
