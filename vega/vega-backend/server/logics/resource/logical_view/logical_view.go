// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logical_view

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"reflect"

	"github.com/dlclark/regexp2"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	derrors "vega-backend/errors"
	"vega-backend/interfaces"
)

// 对和已有重复的数据视图做处理
func (dvs *dataViewService) handleDataViewImportMode(ctx context.Context, tx *sql.Tx, mode string,
	views []*interfaces.DataView) ([]*interfaces.DataView, []*interfaces.DataView, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "data view import mode logic")
	defer span.End()

	createdViews, updatedViews := []*interfaces.DataView{}, []*interfaces.DataView{}

	switch mode {
	case interfaces.ImportMode_Normal:
		for i := 0; i < len(views); i++ {
			view := views[i]
			createdViews = append(createdViews, view)

			// 检查id在系统里是否已经存在
			_, idExist, httpErr := dvs.CheckDataViewExistByID(ctx, tx, view.ViewID)
			if httpErr != nil {
				return createdViews, updatedViews, httpErr
			}

			if idExist {
				errDetails := fmt.Sprintf("DataView ID '%s' already exists", view.ViewID)
				logger.Error(errDetails)
				span.SetStatus(codes.Error, errDetails)
				return createdViews, updatedViews, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataView_Existed_ViewID).
					WithErrorDetails(errDetails)
			}

			// 校验视图名称在分组内是否已存在
			_, nameExist, httpErr := dvs.CheckDataViewExistByName(ctx, tx, view.ViewName, view.GroupName)
			if httpErr != nil {
				return createdViews, updatedViews, httpErr
			}

			if nameExist {
				errDetails := fmt.Sprintf("DataView name '%s' already exists in group '%s'", view.ViewName, view.GroupName)
				logger.Error(errDetails)
				span.SetStatus(codes.Error, errDetails)
				return createdViews, updatedViews, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataView_Existed_ViewName).
					WithDescription(map[string]any{"ViewName": view.ViewName, "GroupName": view.GroupName}).
					WithErrorDetails(errDetails)
			}
		}
	case interfaces.ImportMode_Ignore:
		for i := 0; i < len(views); i++ {
			view := views[i]
			_, idExist, httpErr := dvs.CheckDataViewExistByID(ctx, tx, view.ViewID)
			if httpErr != nil {
				span.SetStatus(codes.Error, "Check data view exist by id failed")
				return createdViews, updatedViews, httpErr
			}

			_, nameExist, httpErr := dvs.CheckDataViewExistByName(ctx, tx, view.ViewName, view.GroupName)
			if httpErr != nil {
				span.SetStatus(codes.Error, "Check data view exist by name failed")
				return createdViews, updatedViews, httpErr
			}

			// 存在重复的就跳过
			if !idExist && !nameExist {
				createdViews = append(createdViews, view)
			}
		}
	case interfaces.ImportMode_Overwrite:
		for i := 0; i < len(views); i++ {
			view := views[i]
			createdViews = append(createdViews, view)

			_, idExist, httpErr := dvs.CheckDataViewExistByID(ctx, tx, view.ViewID)
			if httpErr != nil {
				span.SetStatus(codes.Error, "Check data view exist by id failed")
				return createdViews, updatedViews, httpErr
			}

			existViewID, nameExist, httpErr := dvs.CheckDataViewExistByName(ctx, tx, view.ViewName, view.GroupName)
			if httpErr != nil {
				span.SetStatus(codes.Error, "Check data view exist by name failed")
				return createdViews, updatedViews, httpErr
			}

			if idExist && nameExist {
				// 如果 id 和名称都存在，但是存在的名称对应的视图 id 和当前视图 id 不一样，则报错
				if existViewID != view.ViewID {
					errDetails := fmt.Sprintf("DataView ID '%s' and name '%s' already exist, but the exist view id is '%s'",
						view.ViewID, view.ViewName, existViewID)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return createdViews, updatedViews, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataView_Existed_ViewName).
						WithDescription(map[string]any{"ViewName": view.ViewName, "GroupName": view.GroupName}).
						WithErrorDetails(errDetails)
				} else {
					// 如果 id 和名称都存在，存在的名称对应的视图 id 和当前视图 id 一样，则覆盖更新
					createdViews = createdViews[:len(createdViews)-1]
					updatedViews = append(updatedViews, view)
				}
			}

			// id 已存在，且名称不存在，覆盖更新
			if idExist && !nameExist {
				// 从create数组中删除, 放到更新数组中
				createdViews = createdViews[:len(createdViews)-1]
				updatedViews = append(updatedViews, views[i])
			}

			// 如果 id 不存在，name 存在，报错
			if !idExist && nameExist {
				errDetails := fmt.Sprintf("DataView ID '%s' does not exist, but name '%s' already exists", view.ViewID, view.ViewName)
				logger.Error(errDetails)
				span.SetStatus(codes.Error, errDetails)
				return createdViews, updatedViews, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataView_Existed_ViewName).
					WithDescription(map[string]any{"ViewName": view.ViewName, "GroupName": view.GroupName}).
					WithErrorDetails(errDetails)
			}

			// 如果 id 不存在，name不存在，不需要做什么，由后续批量创建
			// if !idExist && !nameExist {}

		}

	default:
		return createdViews, updatedViews, rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataView_InvalidParameter_ImportMode).WithErrorDetails(fmt.Sprintf("unsupport import_mode %s", mode))
	}

	span.SetStatus(codes.Ok, "")
	return createdViews, updatedViews, nil
}

// 初始化视图分组对象
func initViewGroupReq(dView *interfaces.DataView) *interfaces.ViewGroupReq {
	var builtin bool
	if dView.Type == interfaces.ViewType_Atomic {
		builtin = true
	} else {
		builtin = false
	}

	return &interfaces.ViewGroupReq{
		GroupName: dView.GroupName,
		GroupID:   dView.GroupID,
		Builtin:   builtin,
	}
}

func validateViewNode(ctx context.Context, dvs *dataViewService, node *interfaces.DataScopeNode,
	dataScopeView map[string]*interfaces.DataView) error {
	// 视图节点输入节点必须为空
	if len(node.InputNodes) != 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The view node must have no input node")
	}

	var cfg interfaces.ViewNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("decode view node config failed, %v", err))
	}

	// 判断自定义视图的来源视图是否存在，从这个函数能够拿到字段列表
	atomicView, err := dvs.GetDataView(ctx, cfg.ViewID)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_InvalidReferenceView).
			WithErrorDetails(fmt.Sprintf("get data view %s failed, %v", cfg.ViewID, err))
	}

	if atomicView.Type != interfaces.ViewType_Atomic {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
			WithErrorDetails(fmt.Sprintf("The source view of the custom view '%s' is not an atomic view", cfg.ViewID))
	}

	dataScopeView[atomicView.ViewID] = atomicView

	// fieldsMap 是字段name和字段的映射
	fieldsMap := make(map[string]*interfaces.ViewField)
	for _, viewField := range atomicView.Fields {
		fieldsMap[viewField.Name] = viewField
	}

	// 校验过滤条件
	httpErr := validateCond(ctx, cfg.Filters, fieldsMap)
	if httpErr != nil {
		return httpErr
	}

	// 校验去重配置
	if cfg.Distinct.Enable {
		// 如果视图的查询类型是DSL或索引基类，去重配置不能开启
		if atomicView.QueryType == interfaces.QueryType_DSL || atomicView.QueryType == interfaces.QueryType_IndexBase {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The data scope view query type is DSL or IndexBase, distinct config is not supported")
		}

		// 校验去重字段是否在视图字段列表里，去重字段接口传递的是name
		for _, field := range cfg.Distinct.Fields {
			if _, ok := fieldsMap[field]; !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
					WithErrorDetails(fmt.Sprintf("The field '%s' is not in the view '%s' field list", field, atomicView.ViewName))
			}
		}
	}

	// 校验输出字段是否在视图字段列表里
	for _, field := range node.OutputFields {
		if _, ok := fieldsMap[field.Name]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails(fmt.Sprintf("The field '%s' is not in the view '%s' field list", field.Name, atomicView.ViewName))
		}
	}

	return nil
}

func validateJoinNode(ctx context.Context, node *interfaces.DataScopeNode, nodeMap map[string]struct{}) error {
	// 仅支持两个视图join
	if len(node.InputNodes) != 2 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope join config is invalid, only support two views join")
	}

	// 校验输入节点是否重复
	inputNodesMap := make(map[string]struct{})
	for _, inputNode := range node.InputNodes {
		if _, ok := inputNodesMap[inputNode]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The data scope join config is invalid, input_nodes must be unique")
		}
		inputNodesMap[inputNode] = struct{}{}
	}

	// 校验输入节点是否存在
	for _, inputNode := range node.InputNodes {
		if _, ok := nodeMap[inputNode]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
				WithErrorDetails(fmt.Sprintf("The data scope join config is invalid, input_node '%s' is not exist", inputNode))
		}
	}

	// mapstructure 解析 join_on
	var cfg interfaces.JoinNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope join config is invalid")
	}

	// join_type 只能为 inner, left, right, full outer
	if _, ok := interfaces.JoinTypeMap[cfg.JoinType]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope join config is invalid, join_type must be inner, left, right, full outer")
	}

	// join_on 校验
	if len(cfg.JoinOn) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope join config is invalid, join_on must be set")
	}

	// join_on 校验
	for _, joinOn := range cfg.JoinOn {
		if joinOn.LeftField == "" || joinOn.RightField == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The data scope join config is invalid, join_on left_field and right_field must be set")
		}

		// 操作符必须只为=
		if joinOn.Operator != "=" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The data scope join config is invalid, join_on operator must be =")
		}
	}

	return nil
}

func validateUnionNode(ctx context.Context, qType string, node *interfaces.DataScopeNode, nodeMap map[string]struct{}) error {
	// 当前仅支持两个视图union
	if len(node.InputNodes) < 2 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope union config is invalid, need at least two views union")
	}

	// 校验输入节点是否重复
	inputNodesMap := make(map[string]struct{})
	for _, inputNode := range node.InputNodes {
		if _, ok := inputNodesMap[inputNode]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The data scope union config is invalid, input_nodes must be unique")
		}
		inputNodesMap[inputNode] = struct{}{}
	}

	// 校验输入节点是否存在
	for _, inputNode := range node.InputNodes {
		if _, ok := nodeMap[inputNode]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
				WithErrorDetails(fmt.Sprintf("The data scope union config is invalid, input_node '%s' is not exist", inputNode))
		}
	}

	// mapstructure 解析 union config
	var cfg interfaces.UnionNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope union config is invalid")
	}

	if _, ok := interfaces.UnionTypeMap[cfg.UnionType]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope union config is invalid, union_type must be all, distinct")
	}

	// 如果查询类型是DSL或索引基类，只允许union all
	if qType == interfaces.QueryType_DSL || qType == interfaces.QueryType_IndexBase {
		if cfg.UnionType != interfaces.UnionType_All {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The data scope union config is invalid, DSL or IndexBase view only support union all")
		}
	}

	if qType == interfaces.QueryType_SQL {
		// 校验fields列表长度
		if len(cfg.UnionFields) != len(node.InputNodes) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The data scope union config is invalid, union fields count not equal input nodes count")
		}

		// 校验合并字段是否数量和类型一致
		firstFields := cfg.UnionFields[0]
		for _, uFields := range cfg.UnionFields {
			if len(firstFields) != len(uFields) {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
					WithErrorDetails("The data scope union config is invalid, union fields count not equal")
			}
		}
	}

	return nil
}

func validateSqlNode(ctx context.Context, node *interfaces.DataScopeNode, nodeMap map[string]struct{}) error {
	// 输入节点不能为空
	if len(node.InputNodes) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope sql config is invalid, input_nodes must be set")
	}

	// 校验输入节点是否重复
	inputNodesMap := make(map[string]struct{})
	for _, inputNode := range node.InputNodes {
		if _, ok := inputNodesMap[inputNode]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The data scope sql config is invalid, input_nodes must be unique")
		}
		inputNodesMap[inputNode] = struct{}{}
	}

	// 校验输入节点是否存在
	for _, inputNode := range node.InputNodes {
		if _, ok := nodeMap[inputNode]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
				WithErrorDetails(fmt.Sprintf("The data scope sql config is invalid, input_node '%s' is not exist", inputNode))
		}
	}

	// mapstructure 解析 sql config
	var cfg interfaces.SQLNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope sql config is invalid")
	}

	// 校验 sql_str 是否为空
	if cfg.SQLExpression == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope sql config is invalid, sql_expression must be set")
	}

	return nil
}

func validateOutputNode(ctx context.Context, node *interfaces.DataScopeNode, nodeMap map[string]struct{}) error {
	// 输入节点只能有一个
	if len(node.InputNodes) != 1 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The output node must have one input node")
	}

	// 校验输入节点是否存在
	inputNode := node.InputNodes[0]
	if _, ok := nodeMap[inputNode]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
			WithErrorDetails(fmt.Sprintf("The output node input_node '%s' is not exist", inputNode))
	}

	// 如果没传fields字段列表，默认使用output节点的输出字段
	if len(node.OutputFields) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The output node must have output fields")
	}

	// 校验name不能重复，display_name 不能重复
	nameMap := make(map[string]struct{})
	// originalNameMap := make(map[string]struct{})
	displayNameMap := make(map[string]struct{})
	for _, field := range node.OutputFields {
		if _, ok := nameMap[field.Name]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The output node field name is repeated")
		}
		nameMap[field.Name] = struct{}{}

		// if _, ok := originalNameMap[field.OriginalName]; ok {
		// 	return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
		// 		WithErrorDetails("The output node field original_name is repeated")
		// }
		// originalNameMap[field.OriginalName] = struct{}{}

		if _, ok := displayNameMap[field.DisplayName]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The output node field display_name is repeated")
		}
		displayNameMap[field.DisplayName] = struct{}{}
	}

	return nil
}

// 相比handler层的校验，补充对过滤条件字段类型的校验
// 后续扩充对字段类型和输入字段值是否匹配的校验
func validateCond(ctx context.Context, cfg *interfaces.FilterCondCfg, fieldsMap map[string]*interfaces.ViewField) error {
	if cfg == nil {
		return nil
	}

	// 判断过滤器是否为空对象 {}
	if cfg.Name == "" && cfg.Operation == "" && len(cfg.SubConds) == 0 && cfg.ValueFrom == "" && cfg.Value == nil {
		return nil
	}

	// 过滤条件字段不允许 __id 和 __routing
	if cfg.Name == "__id" || cfg.Name == "__routing" {
		return rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_Forbidden_FilterField).
			WithErrorDetails("The filter field '__id' and '__routing' is not allowed")
	}

	// 过滤操作符
	if cfg.Operation == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_NullParameter_FilterOperation)
	}

	_, exists := dcond.OperationMap[cfg.Operation]
	if !exists {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_UnsupportFilterOperation).
			WithErrorDetails(fmt.Sprintf("unsupport condition operation %s", cfg.Operation))
	}

	switch cfg.Operation {
	case interfaces.OperationAnd, interfaces.OperationOr:
		// 子过滤条件不能超过10个
		if len(cfg.SubConds) > dcond.MaxSubCondition {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_CountExceeded_Filters).
				WithErrorDetails(fmt.Sprintf("The number of subConditions exceeds %d", dcond.MaxSubCondition))
		}

		for _, subCond := range cfg.SubConds {
			err := validateCond(ctx, subCond, fieldsMap)
			if err != nil {
				return err
			}
		}
	default:
		// 过滤字段名称不能为空
		if cfg.Name == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_NullParameter_FilterName)
		}

		// 除了 exist, not_exist, empty, not_empty 外需要校验 value_from
		if _, ok := dcond.NotRequiredValueOperationMap[cfg.Operation]; !ok {
			if cfg.ValueFrom != dcond.ValueFrom_Const {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_ValueFrom).
					WithErrorDetails(fmt.Sprintf("condition does not support value_from type('%s')", cfg.ValueFrom))
			}
		}
	}

	switch cfg.Operation {
	case dcond.OperationEq, dcond.OperationNotEq, dcond.OperationGt, dcond.OperationGte,
		dcond.OperationLt, dcond.OperationLte, dcond.OperationLike, dcond.OperationNotLike,
		dcond.OperationRegex, dcond.OperationMatch, dcond.OperationMatchPhrase, dcond.OperationCurrent:
		// 右侧值为单个值
		_, ok := cfg.Value.([]interface{})
		if ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a single value", cfg.Operation))
		}

		if cfg.Operation == dcond.OperationLike || cfg.Operation == dcond.OperationNotLike ||
			cfg.Operation == dcond.OperationPrefix || cfg.Operation == dcond.OperationNotPrefix {
			_, ok := cfg.Value.(string)
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
					WithErrorDetails("[like not_like prefix not_prefix] operation's value should be a string")
			}
		}

		if cfg.Operation == dcond.OperationRegex {
			val, ok := cfg.Value.(string)
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
					WithErrorDetails("[regex] operation's value should be a string")
			}

			_, err := regexp2.Compile(val, regexp2.RE2)
			if err != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
					WithErrorDetails(fmt.Sprintf("[regex] operation regular expression error: %s", err.Error()))
			}

		}

	case dcond.OperationIn, dcond.OperationNotIn:
		// 当 operation 是 in, not_in 时，value 为任意基本类型的数组，且长度大于等于1；
		_, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[in not_in] operation's value must be an array")
		}

		if len(cfg.Value.([]interface{})) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[in not_in] operation's value should contains at least 1 value")
		}
	case dcond.OperationRange, dcond.OperationOutRange, dcond.OperationBetween:
		// 当 operation 是 range 时，value 是个由范围的下边界和上边界组成的长度为 2 的数值型数组
		// 当 operation 是 out_range 时，value 是个长度为 2 的数值类型的数组，查询的数据范围为 (-inf, value[0]) || [value[1], +inf)
		v, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[range, out_range, between] operation's value must be an array")
		}

		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[range, out_range, between] operation's value must contain 2 values")
		}
	case dcond.OperationBefore:
		// before时, 长度为2的数组，第一个值为时间长度，数值型；第二个值为时间单位，字符串
		v, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[before] operation's value must be an array")
		}

		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[before] operation's value must contain 2 values")
		}
		_, err := common.AssertFloat64(v[0])
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[before] operation's first value should be a number")
		}

		_, ok = v[1].(string)
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[before] operation's second value should be a string")
		}
	}

	switch cfg.Operation {
	case dcond.OperationAnd, dcond.OperationOr:
		for _, subCond := range cfg.SubConds {
			err := validateCond(ctx, subCond, fieldsMap)
			if err != nil {
				return err
			}
		}
	default:
		// 除 * 之外的过滤字段可以在视图字段列表里
		if cfg.Name != interfaces.AllField {
			cField, ok := fieldsMap[cfg.Name]
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataView_InvalidFieldPermission_Filters).
					WithDescription(map[string]any{"FieldName": cfg.Name}).
					WithErrorDetails(fmt.Sprintf("Filter field '%s' is not in view fields list", cfg.Name))
			}

			fieldType := cField.Type
			// binary 类型的字段不支持过滤
			if fieldType == dtype.DataType_Binary {
				return rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataView_FilterBinaryFieldsForbidden).
					WithErrorDetails("Binary fields do not support filtering")
			}

			// empty, not_empty 的字段类型必须为 string
			if cfg.Operation == dcond.OperationEmpty || cfg.Operation == dcond.OperationNotEmpty {
				if !dtype.DataType_IsString(fieldType) {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_FilterFieldTypeMisMatchOperation).
						WithDescription(map[string]any{"FieldName": cfg.Name, "FieldType": fieldType, "Operation": cfg.Operation}).
						WithErrorDetails("Filter field must be of string type when using 'empty' or 'not_empty' operation")
				}
			}
		} else {
			// 如果字段为 *，则只允许使用 match 和 match_phrase 操作符
			if cfg.Operation != dcond.OperationMatch && cfg.Operation != dcond.OperationMatchPhrase {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_FilterFieldTypeMisMatchOperation).
					WithDescription(map[string]any{"FieldName": cfg.Name, "FieldType": "", "Operation": cfg.Operation}).
					WithErrorDetails("Filter field '*' only supports 'match' and 'match_phrase' operations")
			}
		}
	}

	return nil
}
