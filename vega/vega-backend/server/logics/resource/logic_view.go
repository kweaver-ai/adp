// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package resource

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dlclark/regexp2"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"
	"go.opentelemetry.io/otel/codes"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	fcond "vega-backend/logics/filter_condition"
)

// 创建和更新视图的一些通用操作
func (rs *resourceService) validateLogicDefinition(ctx context.Context, view *interfaces.ResourceRequest) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Common operation for creating and updating views")
	defer span.End()

	// 自定义视图
	if view.LogicDefinition == nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails("Logic definition is empty")
	}

	nodeMap := make(map[string]struct{})
	for _, ds := range view.LogicDefinition {
		nodeMap[ds.ID] = struct{}{}
	}

	dataScopeViewMap := make(map[string]*interfaces.Resource)

	for _, node := range view.LogicDefinition {
		switch node.Type {
		case interfaces.DataScopeNodeType_Resource:
			// 校验资源节点
			err := validateResourceNode(ctx, rs, node, dataScopeViewMap)
			if err != nil {
				return err
			}
		case interfaces.DataScopeNodeType_Join:
			err := validateJoinNode(ctx, node, nodeMap)
			if err != nil {
				return err
			}
		case interfaces.DataScopeNodeType_Union:
			err := validateUnionNode(ctx, view.Category, node, nodeMap)
			if err != nil {
				return err
			}
		case interfaces.DataScopeNodeType_Sql:
			if view.Category != interfaces.ResourceCategoryTable {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
					WithErrorDetails("The sql node is only supported in sql query type")
			}

			err := validateSqlNode(ctx, node, nodeMap)
			if err != nil {
				return err
			}
		case interfaces.DataScopeNodeType_Output:
			err := validateOutputNode(ctx, node, nodeMap)
			if err != nil {
				return err
			}
		default:
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The data scope node type is invalid")
		}
	}

	dataScopeViewCategory := make(map[string]struct{})
	dataScopeViewDataSourceID := make(map[string]struct{})
	for _, dsView := range dataScopeViewMap {
		dataScopeViewDataSourceID[dsView.CatalogID] = struct{}{}
		dataScopeViewCategory[dsView.Category] = struct{}{}
	}

	if len(dataScopeViewCategory) != 1 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The source view of the custom view must have the same category")
	}

	// 如果数据源类型是opensearch，则不能跨opensearch数据源选择
	if view.Category == interfaces.ResourceCategoryIndex && len(dataScopeViewDataSourceID) > 1 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The source view of query type DSL must have the same data source when create custom view")
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func validateResourceNode(ctx context.Context, dvs *resourceService, node *interfaces.DataScopeNode,
	dataScopeView map[string]*interfaces.Resource) error {
	// 资源节点输入节点必须为空
	if len(node.InputNodes) != 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The resource node must have no input node")
	}

	var cfg interfaces.ResourceNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("decode resource node config failed, %v", err))
	}

	// 判断自定义视图的来源表是否存在，从这个函数能够拿到字段列表
	atomicView, err := dvs.GetByID(ctx, cfg.ResourceID)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails(fmt.Sprintf("get resource %s failed, %v", cfg.ResourceID, err))
	}

	// 校验来源视图的类型
	switch atomicView.Category {
	case interfaces.ResourceCategoryTable:
	case interfaces.ResourceCategoryFile:
	case interfaces.ResourceCategoryFileset:
	case interfaces.ResourceCategoryAPI:
	case interfaces.ResourceCategoryTopic:
	case interfaces.ResourceCategoryIndex:
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails(fmt.Sprintf("The source resource of the custom view '%s' is not supported", cfg.ResourceID))

	}

	dataScopeView[atomicView.ID] = atomicView

	// fieldsMap 是字段name和字段的映射
	fieldsMap := make(map[string]*interfaces.Property)
	for _, viewField := range atomicView.SchemaDefinition {
		fieldsMap[viewField.Name] = viewField
	}

	// 校验过滤条件
	httpErr := validateCond(ctx, cfg.Filters, fieldsMap)
	if httpErr != nil {
		return httpErr
	}

	// 校验去重配置, 只有 table 去重配置
	if cfg.Distinct.Enable {
		if atomicView.Category != interfaces.ResourceCategoryTable {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The data scope view category is not table, distinct config is not supported")
		}

		// 校验去重字段是否在视图字段列表里，去重字段接口传递的是name
		for _, field := range cfg.Distinct.Fields {
			if _, ok := fieldsMap[field]; !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
					WithErrorDetails(fmt.Sprintf("The field '%s' is not in the view '%s' field list", field, atomicView.Name))
			}
		}
	}

	// 校验输出字段是否在视图字段列表里
	for _, field := range node.OutputFields {
		if _, ok := fieldsMap[field.Name]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails(fmt.Sprintf("The field '%s' is not in the view '%s' field list", field.Name, atomicView.Name))
		}
	}

	return nil
}

func validateJoinNode(ctx context.Context, node *interfaces.DataScopeNode, nodeMap map[string]struct{}) error {
	// 仅支持两个视图join
	if len(node.InputNodes) != 2 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The data scope join config is invalid, only support two views join")
	}

	// 校验输入节点是否重复
	inputNodesMap := make(map[string]struct{})
	for _, inputNode := range node.InputNodes {
		if _, ok := inputNodesMap[inputNode]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The data scope join config is invalid, input_nodes must be unique")
		}
		inputNodesMap[inputNode] = struct{}{}
	}

	// 校验输入节点是否存在
	for _, inputNode := range node.InputNodes {
		if _, ok := nodeMap[inputNode]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails(fmt.Sprintf("The data scope join config is invalid, input_node '%s' is not exist", inputNode))
		}
	}

	// mapstructure 解析 join_on
	var cfg interfaces.JoinNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The data scope join config is invalid")
	}

	// join_type 只能为 inner, left, right, full outer
	if _, ok := interfaces.JoinTypeMap[cfg.JoinType]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The data scope join config is invalid, join_type must be inner, left, right, full outer")
	}

	// join_on 校验
	if len(cfg.JoinOn) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The data scope join config is invalid, join_on must be set")
	}

	// join_on 校验
	for _, joinOn := range cfg.JoinOn {
		if joinOn.LeftField == "" || joinOn.RightField == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The data scope join config is invalid, join_on left_field and right_field must be set")
		}

		// 操作符必须只为=
		if joinOn.Operator != "=" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The data scope join config is invalid, join_on operator must be =")
		}
	}

	return nil
}

func validateUnionNode(ctx context.Context, category string, node *interfaces.DataScopeNode, nodeMap map[string]struct{}) error {
	// 当前仅支持两个视图union
	if len(node.InputNodes) < 2 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The data scope union config is invalid, need at least two views union")
	}

	// 校验输入节点是否重复
	inputNodesMap := make(map[string]struct{})
	for _, inputNode := range node.InputNodes {
		if _, ok := inputNodesMap[inputNode]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The data scope union config is invalid, input_nodes must be unique")
		}
		inputNodesMap[inputNode] = struct{}{}
	}

	// 校验输入节点是否存在
	for _, inputNode := range node.InputNodes {
		if _, ok := nodeMap[inputNode]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails(fmt.Sprintf("The data scope union config is invalid, input_node '%s' is not exist", inputNode))
		}
	}

	// mapstructure 解析 union config
	var cfg interfaces.UnionNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The data scope union config is invalid")
	}

	if _, ok := interfaces.UnionTypeMap[cfg.UnionType]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The data scope union config is invalid, union_type must be all, distinct")
	}

	// 如果查询类型是DSL或索引基类，只允许union all
	if category == interfaces.ResourceCategoryIndex {
		if cfg.UnionType != interfaces.UnionType_All {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The data scope union config is invalid, DSL or IndexBase view only support union all")
		}
	}

	if category == interfaces.ResourceCategoryTable {
		// 校验fields列表长度
		if len(cfg.UnionFields) != len(node.InputNodes) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The data scope union config is invalid, union fields count not equal input nodes count")
		}

		// 校验合并字段是否数量和类型一致
		firstFields := cfg.UnionFields[0]
		for _, uFields := range cfg.UnionFields {
			if len(firstFields) != len(uFields) {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
					WithErrorDetails("The data scope union config is invalid, union fields count not equal")
			}
		}
	}

	return nil
}

func validateSqlNode(ctx context.Context, node *interfaces.DataScopeNode, nodeMap map[string]struct{}) error {
	// 输入节点不能为空
	if len(node.InputNodes) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The data scope sql config is invalid, input_nodes must be set")
	}

	// 校验输入节点是否重复
	inputNodesMap := make(map[string]struct{})
	for _, inputNode := range node.InputNodes {
		if _, ok := inputNodesMap[inputNode]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The data scope sql config is invalid, input_nodes must be unique")
		}
		inputNodesMap[inputNode] = struct{}{}
	}

	// 校验输入节点是否存在
	for _, inputNode := range node.InputNodes {
		if _, ok := nodeMap[inputNode]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails(fmt.Sprintf("The data scope sql config is invalid, input_node '%s' is not exist", inputNode))
		}
	}

	// mapstructure 解析 sql config
	var cfg interfaces.SQLNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The data scope sql config is invalid")
	}

	// 校验 sql_str 是否为空
	if cfg.SQLExpression == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The data scope sql config is invalid, sql_expression must be set")
	}

	return nil
}

func validateOutputNode(ctx context.Context, node *interfaces.DataScopeNode, nodeMap map[string]struct{}) error {
	// 输入节点只能有一个
	if len(node.InputNodes) != 1 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The output node must have one input node")
	}

	// 校验输入节点是否存在
	inputNode := node.InputNodes[0]
	if _, ok := nodeMap[inputNode]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails(fmt.Sprintf("The output node input_node '%s' is not exist", inputNode))
	}

	// 如果没传fields字段列表，默认使用output节点的输出字段
	if len(node.OutputFields) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The output node must have output fields")
	}

	// 校验name不能重复，display_name 不能重复
	nameMap := make(map[string]struct{})
	// originalNameMap := make(map[string]struct{})
	displayNameMap := make(map[string]struct{})
	for _, field := range node.OutputFields {
		if _, ok := nameMap[field.Name]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The output node field name is repeated")
		}
		nameMap[field.Name] = struct{}{}

		// if _, ok := originalNameMap[field.OriginalName]; ok {
		// 	return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
		// 		WithErrorDetails("The output node field original_name is repeated")
		// }
		// originalNameMap[field.OriginalName] = struct{}{}

		if _, ok := displayNameMap[field.DisplayName]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The output node field display_name is repeated")
		}
		displayNameMap[field.DisplayName] = struct{}{}
	}

	return nil
}

// 相比handler层的校验，补充对过滤条件字段类型的校验
// 后续扩充对字段类型和输入字段值是否匹配的校验
func validateCond(ctx context.Context, cfg *interfaces.FilterCondCfg, fieldsMap map[string]*interfaces.Property) error {
	if cfg == nil {
		return nil
	}

	// 判断过滤器是否为空对象 {}
	if cfg.Name == "" && cfg.Operation == "" && len(cfg.SubConds) == 0 && cfg.ValueFrom == "" && cfg.Value == nil {
		return nil
	}

	// 过滤条件字段不允许 __id 和 __routing
	if cfg.Name == "__id" || cfg.Name == "__routing" {
		return rest.NewHTTPError(ctx, http.StatusForbidden, verrors.VegaBackend_InvalidParameter_FilterCondition).
			WithErrorDetails("The filter field '__id' and '__routing' is not allowed")
	}

	// 过滤操作符
	if cfg.Operation == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_NullParameter_FilterConditionOperation)
	}

	_, exists := fcond.OperationMap[cfg.Operation]
	if !exists {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_UnsupportFilterConditionOperation).
			WithErrorDetails(fmt.Sprintf("unsupport condition operation %s", cfg.Operation))
	}

	switch cfg.Operation {
	case fcond.OperationAnd, fcond.OperationOr:
		// 子过滤条件不能超过10个
		if len(cfg.SubConds) > interfaces.MaxSubCondition {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_CountExceeded_FilterConditionSubConds).
				WithErrorDetails(fmt.Sprintf("The number of subConditions exceeds %d", interfaces.MaxSubCondition))
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
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_NullParameter_FilterConditionName)
		}
	}

	switch cfg.Operation {
	case fcond.OperationEqual, fcond.OperationNotEqual, fcond.OperationGt, fcond.OperationGte,
		fcond.OperationLt, fcond.OperationLte, fcond.OperationLike, fcond.OperationNotLike,
		fcond.OperationRegex, fcond.OperationMatch, fcond.OperationMatchPhrase, fcond.OperationCurrent:
		// 右侧值为单个值
		_, ok := cfg.Value.([]interface{})
		if ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a single value", cfg.Operation))
		}

		if cfg.Operation == fcond.OperationLike || cfg.Operation == fcond.OperationNotLike ||
			cfg.Operation == fcond.OperationPrefix || cfg.Operation == fcond.OperationNotPrefix {
			_, ok := cfg.Value.(string)
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
					WithErrorDetails("[like not_like prefix not_prefix] operation's value should be a string")
			}
		}

		if cfg.Operation == fcond.OperationRegex {
			val, ok := cfg.Value.(string)
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
					WithErrorDetails("[regex] operation's value should be a string")
			}

			_, err := regexp2.Compile(val, regexp2.RE2)
			if err != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
					WithErrorDetails(fmt.Sprintf("[regex] operation regular expression error: %s", err.Error()))
			}

		}

	case fcond.OperationIn, fcond.OperationNotIn:
		// 当 operation 是 in, not_in 时，value 为任意基本类型的数组，且长度大于等于1；
		_, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
				WithErrorDetails("[in not_in] operation's value must be an array")
		}

		if len(cfg.Value.([]interface{})) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
				WithErrorDetails("[in not_in] operation's value should contains at least 1 value")
		}
	case fcond.OperationRange, fcond.OperationOutRange, fcond.OperationBetween:
		// 当 operation 是 range 时，value 是个由范围的下边界和上边界组成的长度为 2 的数值型数组
		// 当 operation 是 out_range 时，value 是个长度为 2 的数值类型的数组，查询的数据范围为 (-inf, value[0]) || [value[1], +inf)
		v, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
				WithErrorDetails("[range, out_range, between] operation's value must be an array")
		}

		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
				WithErrorDetails("[range, out_range, between] operation's value must contain 2 values")
		}
	case fcond.OperationBefore:
		// before时, 长度为2的数组，第一个值为时间长度，数值型；第二个值为时间单位，字符串
		_, ok := cfg.Value.(float64)
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
				WithErrorDetails("[before] operation's value must be an array")
		}

		_, ok = cfg.RemainCfg["unit"]
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
				WithErrorDetails("[before] operation's remain cfg must contain unit")
		}
	}

	switch cfg.Operation {
	case fcond.OperationAnd, fcond.OperationOr:
		for _, subCond := range cfg.SubConds {
			err := validateCond(ctx, subCond, fieldsMap)
			if err != nil {
				return err
			}
		}
	default:
		// 除 * 之外的过滤字段在视图字段列表里
		if cfg.Name != interfaces.AllField {
			cField, ok := fieldsMap[cfg.Name]
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusForbidden, verrors.VegaBackend_InvalidParameter_FilterCondition).
					WithErrorDetails(fmt.Sprintf("Filter field '%s' is not in view fields list", cfg.Name))
			}

			fieldType := cField.Type
			// binary 类型的字段不支持过滤
			if fieldType == interfaces.DataType_Binary {
				return rest.NewHTTPError(ctx, http.StatusForbidden, verrors.VegaBackend_InvalidParameter_FilterCondition).
					WithErrorDetails("Binary fields do not support filtering")
			}

			// empty, not_empty 的字段类型必须为 string
			if cfg.Operation == fcond.OperationEmpty || cfg.Operation == fcond.OperationNotEmpty {
				if !interfaces.DataType_IsString(fieldType) {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterCondition).
						WithErrorDetails("Filter field must be of string type when using 'empty' or 'not_empty' operation")
				}
			}
		} else {
			// 如果字段为 *，则只允许使用 match 和 match_phrase 操作符
			if cfg.Operation != fcond.OperationMatch && cfg.Operation != fcond.OperationMatchPhrase &&
				cfg.Operation != fcond.OperationMultiMatch {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterCondition).
					WithErrorDetails("Filter field '*' only supports 'match', 'match_phrase' and 'multi_match' operations")
			}
		}
	}

	return nil
}

// func (rs *resourceService) getLogicViewSource(ctx context.Context, resource *interfaces.Resource) (*interfaces.Resource, error) {
// 	// 给每个原子视图添加对应的技术名称（DSL类视图技术名称对应的是来源索引库），uniquery查询数据时需要
// 	for _, node := range resource.LogicDefinition {
// 		if node.Type != interfaces.DataScopeNodeType_Resource {
// 			continue
// 		}

// 		var viewID string
// 		var ok bool
// 		if viewID, ok = node.Config["resource_id"].(string); !ok {
// 			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
// 				WithErrorDetails("resource_id is not string")
// 		}

// 		// if includeDataScopeViews {

// 		// 获取原子视图的信息
// 		resource, err := rs.GetByID(ctx, viewID)
// 		if err != nil {
// 			return nil, err
// 		}

// 		// fieldsMap := make(map[string]*interfaces.ViewProperty)
// 		// for _, vf := range resource.SchemaDefinition {
// 		// 	fieldsMap[vf.Name] = &interfaces.ViewProperty{
// 		// 		Property:    *vf,
// 		// 		SrcNodeID:   node.ID,
// 		// 		SrcNodeName: node.Title,
// 		// 	}
// 		// }
// 		// resource.FieldsMap = fieldsMap

// 		node.Config["resource"] = resource
// 		// }
// 	}

// 	// fieldsMap := make(map[string]*interfaces.ViewProperty)
// 	// for _, vf := range resource.SchemaDefinition {
// 	// 	// name 作为 key
// 	// 	fieldsMap[vf.Name] = vf
// 	// }

// 	// resource.FieldsMap = fieldsMap

// 	return resource, nil
// }
