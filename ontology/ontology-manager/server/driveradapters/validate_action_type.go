package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	cond "ontology-manager/common/condition"
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

	// 行动条件非空时，校验行动条件
	if actionType.Condition != nil {
		err = validateActionCondition(ctx, actionType.Condition, actionType.ObjectTypeID)
		if err != nil {
			return err
		}
	}

	return nil
}

// 校验行动条件的合法性
func validateActionCondition(ctx context.Context, cfg *interfaces.CondCfg, objectTypeID string) error {
	if cfg == nil {
		return nil
	}

	// 如果行动条件不给对象类id，那么就默认使用行动类的对象类id
	if cfg.ObjectTypeID == "" {
		cfg.ObjectTypeID = objectTypeID
	}
	// if cfg.ObjectTypeID == "" {
	// 	return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyManager_ActionType_InvalidParameter).
	// 		WithErrorDetails("行动条件的对象类不能为空")
	// }

	// 过滤操作符
	if cfg.Operation == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyManager_ActionType_InvalidParameter).
			WithErrorDetails("行动条件的过滤条件不能为空")
	}

	_, exists := interfaces.ActionCondOperationMap[cfg.Operation]
	if !exists {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyManager_ActionType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("行动条件的操作符[%s]不支持", cfg.Operation))
	}

	switch cfg.Operation {
	case cond.OperationAnd, cond.OperationOr:
		// 子过滤条件不能超过100个
		if len(cfg.SubConds) > cond.MaxSubCondition {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyManager_CountExceeded_Conditions).
				WithErrorDetails(fmt.Sprintf("行动条件的子条件不能超过 %d 个", cond.MaxSubCondition))
		}

		for _, subCond := range cfg.SubConds {
			err := validateActionCondition(ctx, subCond, objectTypeID)
			if err != nil {
				return err
			}
		}
	default:
		// 过滤字段名称不能为空
		if cfg.Field == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyManager_ActionType_InvalidParameter).
				WithErrorDetails("行动条件的过滤字段不能为空")

		}
	}

	switch cfg.Operation {
	case cond.OperationEq, cond.OperationNotEq, cond.OperationGt, cond.OperationGte, cond.OperationLt, cond.OperationLte:
		// 右侧值为单个值
		_, ok := cfg.Value.([]any)
		if ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyManager_ActionType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a single value", cfg.Operation))
		}

	case cond.OperationIn, cond.OperationNotIn:
		// 当 operation 是 in, not_in 时，value 为任意基本类型的数组，且长度大于等于1；
		_, ok := cfg.Value.([]any)
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyManager_ActionType_InvalidParameter).
				WithErrorDetails("[in not_in] operation's value must be an array")
		}

		if len(cfg.Value.([]any)) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyManager_ActionType_InvalidParameter).
				WithErrorDetails("[in not_in] operation's value should contains at least 1 value")
		}
	case cond.OperationRange, cond.OperationOutRange, cond.OperationBefore, cond.OperationBetween:
		// 当 operation 是 range 时，value 是个由范围的下边界和上边界组成的长度为 2 的数值型数组
		// 当 operation 是 out_range 时，value 是个长度为 2 的数值类型的数组，查询的数据范围为 (-inf, value[0]) || [value[1], +inf)
		v, ok := cfg.Value.([]any)
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyManager_ActionType_InvalidParameter).
				WithErrorDetails("[range, out_range] operation's value must be an array")
		}

		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyManager_ActionType_InvalidParameter).
				WithErrorDetails("[range, out_range] operation's value must contain 2 values")
		}
	}

	return nil
}
