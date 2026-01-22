package data_view

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/dlclark/regexp2"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"

	"uniquery/common"
	cond "uniquery/common/condition"
	conv "uniquery/common/convert"
	vopt "uniquery/common/value_opt"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	dtype "uniquery/interfaces/data_type"
)

// 判断视图的数据源是否来自同一个数据源
func isSingleDataSource(view *interfaces.DataView) bool {
	if view.Type == interfaces.ViewType_Atomic {
		return true
	}

	if view.Type == interfaces.ViewType_Custom {
		return view.DataScopeAdvancedParams.IsSingleSource
	}

	return false
}

// 获取查询vega数据的dataSourceID
func getQueryDataSourceID(view *interfaces.DataView) string {
	if view.Type == interfaces.ViewType_Atomic {
		return view.DataSourceID
	}

	if view.Type == interfaces.ViewType_Custom {
		return view.DataScopeAdvancedParams.DataScopeDataSourceID
	}

	return ""
}

func GetBaseTypes(view *interfaces.DataView) ([]string, map[string]string, error) {
	baseTypes := []string{}
	// 索引库和视图 id 的映射
	baseTypeViewMap := make(map[string]string)

	switch view.Type {
	case interfaces.ViewType_Atomic:
		baseTypes = append(baseTypes, view.TechnicalName)
		baseTypeViewMap[view.TechnicalName] = view.ViewID
	case interfaces.ViewType_Custom:
		for _, node := range view.DataScope {
			// 自定义视图的索引库列表
			if node.Type == interfaces.DataScopeNodeType_View {
				var viewNodeConfig interfaces.ViewNodeCfg
				err := mapstructure.Decode(node.Config, &viewNodeConfig)
				if err != nil {
					logger.Errorf("Decode view node config failed, err: %v", err)
					return nil, nil, err
				}

				if viewNodeConfig.View == nil {
					logger.Errorf("View node config view is nil")
					return nil, nil, fmt.Errorf("view node config view is nil")
				}

				baseTypes = append(baseTypes, viewNodeConfig.View.TechnicalName)
				baseTypeViewMap[viewNodeConfig.View.TechnicalName] = viewNodeConfig.ViewID
			}
		}
	}

	return baseTypes, baseTypeViewMap, nil

}

// 将索引库字段转为视图字段
func convertIndexBaseFieldsToViewFields(baseInfos []interfaces.IndexBase) ([]*cond.ViewField, map[string]*cond.ViewField) {
	viewFields := make([]*cond.ViewField, 0)
	viewFieldsMap := make(map[string]*cond.ViewField)
	// 获取日志库字段转成视图字段格式
	for _, base := range baseInfos {
		allBaseFields := mergeIndexBaseFields(base.Mappings)
		for _, field := range allBaseFields {
			displayName := field.DisplayName
			if displayName == "" {
				displayName = field.Field
			}

			fieldType, ok := dtype.IndexBase_DataType_Map[field.Type]
			if !ok {
				fieldType = field.Type
			}

			f := &cond.ViewField{
				Name:         field.Field,
				Type:         fieldType,
				DisplayName:  displayName,
				OriginalName: field.Field,
			}

			viewFields = append(viewFields, f)
			viewFieldsMap[field.Field] = f
		}
	}

	return viewFields, viewFieldsMap
}

// 校验自定义视图配置
func validateDataScope(ctx context.Context, dvs *dataViewService, view *interfaces.DataView) ([]string, map[string]string, error) {
	if view.DataScope == nil {
		return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails("data scope is empty")
	}

	// 节点数不能超过20
	if len(view.DataScope) > 20 {
		return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails("data scope node count exceeds 20")
	}

	nodeMap := make(map[string]struct{})
	for _, ds := range view.DataScope {
		nodeMap[ds.ID] = struct{}{}
	}

	baseTypes := []string{}
	baseTypeViewMap := make(map[string]string) // 视图 id 和索引库的映射
	dataScopeViewQueryType := make(map[string]struct{})
	dataScopeViewDataSourceID := make(map[string]struct{})

	for _, node := range view.DataScope {
		switch node.Type {
		case interfaces.DataScopeNodeType_View:
			// 校验视图节点
			atomicView, err := validateViewNode(ctx, dvs, node)
			if err != nil {
				return nil, nil, err
			}

			baseTypes = append(baseTypes, atomicView.TechnicalName)
			baseTypeViewMap[atomicView.TechnicalName] = atomicView.ViewID
			dataScopeViewQueryType[atomicView.QueryType] = struct{}{}
			dataScopeViewDataSourceID[atomicView.DataSourceID] = struct{}{}
			view.DataScopeAdvancedParams.DataScopeDataSourceID = atomicView.DataSourceID

		case interfaces.DataScopeNodeType_Join:
			err := validateJoinNode(ctx, node, nodeMap)
			if err != nil {
				return nil, nil, err
			}
		case interfaces.DataScopeNodeType_Union:
			err := validateUnionNode(ctx, view.QueryType, node, nodeMap)
			if err != nil {
				return nil, nil, err
			}
		case interfaces.DataScopeNodeType_Sql:
			if view.QueryType != interfaces.QueryType_SQL {
				return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
					WithErrorDetails("The sql node is only supported in sql query type")
			}

			hasStar, err := validateSqlNode(ctx, node, nodeMap)
			if err != nil {
				return nil, nil, err
			}
			// 标识有sql node
			view.HasDataScopeSQLNode = true
			view.HasStar = hasStar
		case interfaces.DataScopeNodeType_Output:
			// if previewDataScopeNodeID != node.ID {
			// 	break
			// }

			err := validateOutputNode(ctx, node, nodeMap)
			if err != nil {
				return nil, nil, err
			}

			// 如果没传fields，使用输出节点的fields
			if len(view.Fields) == 0 {
				view.Fields = node.OutputFields
			}

		default:
			return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
				WithErrorDetails("the data scope node type is invalid")
		}
	}

	if len(dataScopeViewQueryType) != 1 {
		return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
			WithErrorDetails("the source view of the custom view must have the same query type")
	}

	view.IsSingleSource = len(dataScopeViewDataSourceID) == 1

	// 如果输出节点也没有输出字段,字段范围是全部字段
	if view.HasStar || len(view.Fields) == 0 {
		view.FieldScope = interfaces.FieldScope_All
	}

	return baseTypes, baseTypeViewMap, nil
}

func validateViewNode(ctx context.Context, dvs *dataViewService, node *interfaces.DataScopeNode) (*interfaces.DataView, error) {
	// 视图节点输入节点必须为空
	if len(node.InputNodes) != 0 {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The view node must have no input node")
	}

	var cfg interfaces.ViewNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("decode view node config failed, %v", err))
	}

	// 判断自定义视图的来源视图是否存在，从这个函数能够拿到字段列表
	atomicView, err := dvs.GetDataViewByID(ctx, cfg.ViewID, false)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("get data view %s failed, %v", cfg.ViewID, err))
	}

	if atomicView.Type != interfaces.ViewType_Atomic {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
			WithErrorDetails(fmt.Sprintf("The source view of the custom view '%s' is not an atomic view", cfg.ViewID))
	}

	// fieldsMap 是字段name和字段的映射
	fieldsMap := make(map[string]*cond.ViewField)
	for _, viewField := range atomicView.Fields {
		fieldsMap[viewField.Name] = viewField
	}
	// node.Config["fields_map"] = fieldsMap
	// 补充 data scope 的原子视图信息
	node.Config["view"] = atomicView

	// 校验过滤条件
	httpErr := validateCond(ctx, cfg.Filters, fieldsMap)
	if httpErr != nil {
		return nil, httpErr
	}

	// 校验去重配置
	if cfg.Distinct.Enable {
		// 如果视图的查询类型是DSL或索引库查询，去重配置不能开启
		if atomicView.QueryType == interfaces.QueryType_DSL || atomicView.QueryType == interfaces.QueryType_IndexBase {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The data scope view query type is DSL, distinct config is not supported")
		}

		// 校验去重字段是否在视图字段列表里，去重字段接口传递的是name
		for _, field := range cfg.Distinct.Fields {
			if _, ok := fieldsMap[field]; !ok {
				return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
					WithErrorDetails(fmt.Sprintf("The field '%s' is not in the view field list", field))
			}
		}
	}

	// 校验输出字段是否在视图字段列表里
	for _, field := range node.OutputFields {
		if _, ok := fieldsMap[field.Name]; !ok {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails(fmt.Sprintf("The field '%s' is not in the view field list", field.Name))
		}
	}

	return atomicView, nil
}

func validateJoinNode(ctx context.Context, node *interfaces.DataScopeNode, nodeMap map[string]struct{}) error {
	// 仅支持两个视图join
	if len(node.InputNodes) != 2 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope join config is invalid, only support two views join")
	}

	// 校验输入节点是否重复
	inputNodesMap := make(map[string]struct{})
	for _, inputNode := range node.InputNodes {
		if _, ok := inputNodesMap[inputNode]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The data scope join config is invalid, input_nodes must be unique")
		}
		inputNodesMap[inputNode] = struct{}{}
	}

	// 校验输入节点是否存在
	for _, inputNode := range node.InputNodes {
		if _, ok := nodeMap[inputNode]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
				WithErrorDetails(fmt.Sprintf("The data scope join config is invalid, input_node '%s' is not exist", inputNode))
		}
	}

	// mapstructure 解析 join_on
	var cfg interfaces.JoinNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope join config is invalid")
	}

	// join_type 只能为 inner, left, right, full outer
	if _, ok := interfaces.JoinTypeMap[cfg.JoinType]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope join config is invalid, join_type must be inner, left, right, full outer")
	}

	// join_on 校验
	if len(cfg.JoinOn) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope join config is invalid, join_on must be set")
	}

	// join_on 校验
	for _, joinOn := range cfg.JoinOn {
		if joinOn.LeftField == "" || joinOn.RightField == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The data scope join config is invalid, join_on left_field and right_field must be set")
		}

		// 操作符必须只为=
		if joinOn.Operator != "=" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The data scope join config is invalid, join_on operator must be =")
		}
	}

	return nil
}

func validateUnionNode(ctx context.Context, qType string, node *interfaces.DataScopeNode, nodeMap map[string]struct{}) error {
	// 当前仅支持两个视图union
	if len(node.InputNodes) < 2 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope union config is invalid, need at least two views union")
	}

	// 校验输入节点是否重复
	inputNodesMap := make(map[string]struct{})
	for _, inputNode := range node.InputNodes {
		if _, ok := inputNodesMap[inputNode]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The data scope union config is invalid, input_nodes must be unique")
		}
		inputNodesMap[inputNode] = struct{}{}
	}

	// 校验输入节点是否存在
	for _, inputNode := range node.InputNodes {
		if _, ok := nodeMap[inputNode]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
				WithErrorDetails(fmt.Sprintf("The data scope union config is invalid, input_node '%s' is not exist", inputNode))
		}
	}

	// mapstructure 解析 union config
	var cfg interfaces.UnionNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope union config is invalid")
	}

	if _, ok := interfaces.UnionTypeMap[cfg.UnionType]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope union config is invalid, union_type must be all, distinct")
	}

	// 如果查询类型是DSL或索引库查询，只允许union all
	if qType == interfaces.QueryType_DSL || qType == interfaces.QueryType_IndexBase {
		if cfg.UnionType != interfaces.UnionType_All {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The data scope union config is invalid, dsl or IndexBase view only support union all")
		}
	}

	if qType == interfaces.QueryType_SQL {
		// 校验fields列表长度
		if len(cfg.UnionFields) != len(node.InputNodes) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The data scope union config is invalid, union fields count not equal input nodes count")
		}

		// 校验合并字段是否数量和类型一致
		firstFields := cfg.UnionFields[0]
		for _, uFields := range cfg.UnionFields {
			if len(firstFields) != len(uFields) {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
					WithErrorDetails("The data scope union config is invalid, union fields count not equal")
			}
		}
	}

	return nil
}

func validateSqlNode(ctx context.Context, node *interfaces.DataScopeNode, nodeMap map[string]struct{}) (bool, error) {
	// 输入节点不能为空
	if len(node.InputNodes) == 0 {
		return false, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope sql config is invalid, input_nodes must be set")
	}

	// 校验输入节点是否重复
	inputNodesMap := make(map[string]struct{})
	for _, inputNode := range node.InputNodes {
		if _, ok := inputNodesMap[inputNode]; ok {
			return false, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The data scope sql config is invalid, input_nodes must be unique")
		}
		inputNodesMap[inputNode] = struct{}{}
	}

	// 校验输入节点是否存在
	for _, inputNode := range node.InputNodes {
		if _, ok := nodeMap[inputNode]; !ok {
			return false, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
				WithErrorDetails(fmt.Sprintf("The data scope sql config is invalid, input_node '%s' is not exist", inputNode))
		}
	}

	// mapstructure 解析 sql config
	var cfg interfaces.SQLNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return false, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope sql config is invalid")
	}

	// 校验 sql_str 是否为空
	if cfg.SQLExpression == "" {
		return false, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The data scope sql config is invalid, sql_expression must be set")
	}

	// 解析sql里的字段，补全sql节点的输出字段列表
	processedSQL := replaceTablePlaceholders(cfg.SQLExpression, "`")
	logger.Infof("processedSQL for parse fields from sql_expression: %s", processedSQL)
	parser := NewSQLFieldParser()
	info := parser.Parse(processedSQL)

	// 输出 JSON 格式
	logger.Infof("\nJSON 格式:\n%s\n\n", info.FormatAsJSON())

	// 组装sql的输出字段
	outputFields := make([]*cond.ViewField, 0)
	for _, field := range info.Fields {
		outputFields = append(outputFields, &cond.ViewField{
			Name:         common.CE(field.Alias != "", field.Alias, field.Name),
			OriginalName: field.Name,
			DisplayName:  common.CE(field.Alias != "", field.Alias, field.Name),
		})
	}

	node.OutputFields = outputFields

	return info.HasStar, nil
}

func validateOutputNode(ctx context.Context, node *interfaces.DataScopeNode, nodeMap map[string]struct{}) error {
	// 输入节点只能有一个
	if len(node.InputNodes) != 1 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
			WithErrorDetails("The output node must have one input node")
	}

	// 校验输入节点是否存在
	inputNode := node.InputNodes[0]
	if _, ok := nodeMap[inputNode]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
			WithErrorDetails(fmt.Sprintf("The output node input_node '%s' is not exist", inputNode))
	}

	// 如果没传fields字段列表，默认使用output节点的输出字段
	// if len(node.OutputFields) == 0 {
	// 	// return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
	// 	// 	WithErrorDetails("The output node must have output fields")
	// }

	// 校验name不能重复，display_name 不能重复, original_name 可以重复
	nameMap := make(map[string]struct{})
	// originalNameMap := make(map[string]struct{})
	displayNameMap := make(map[string]struct{})
	for _, field := range node.OutputFields {
		if _, ok := nameMap[field.Name]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The output node field name is repeated")
		}
		nameMap[field.Name] = struct{}{}

		// if _, ok := originalNameMap[field.OriginalName]; ok {
		// 	return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
		// 		WithErrorDetails("The output node field original_name is repeated")
		// }
		// originalNameMap[field.OriginalName] = struct{}{}

		if _, ok := displayNameMap[field.DisplayName]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataScope).
				WithErrorDetails("The output node field display_name is repeated")
		}
		displayNameMap[field.DisplayName] = struct{}{}
	}

	return nil
}

// 相比handler层的校验，补充对过滤条件字段类型的校验
// 后续扩充对字段类型和输入字段值是否匹配的校验
func validateCond(ctx context.Context, cfg *cond.CondCfg, fieldsMap map[string]*cond.ViewField) error {
	if cfg == nil {
		return nil
	}

	// 判断过滤器是否为空对象 {}
	if cfg.Name == "" && cfg.Operation == "" && len(cfg.SubConds) == 0 && cfg.ValueFrom == "" && cfg.Value == nil {
		return nil
	}

	// 过滤条件字段不允许 __id 和 __routing
	if cfg.Name == "__id" || cfg.Name == "__routing" {
		return rest.NewHTTPError(ctx, http.StatusForbidden, uerrors.Uniquery_Forbidden_FilterField).
			WithErrorDetails("The filter field '__id' and '__routing' is not allowed")
	}

	// 过滤操作符
	if cfg.Operation == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_NullParameter_FilterOperation)
	}

	_, exists := cond.OperationMap[cfg.Operation]
	if !exists {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_UnsupportFilterOperation).
			WithErrorDetails(fmt.Sprintf("unsupport condition operation %s", cfg.Operation))
	}

	switch cfg.Operation {
	case cond.OperationAnd, cond.OperationOr:
		// 子过滤条件不能超过10个
		if len(cfg.SubConds) > cond.MaxSubCondition {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_CountExceeded_Filters).
				WithErrorDetails(fmt.Sprintf("The number of subConditions exceeds %d", cond.MaxSubCondition))
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
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_NullParameter_FilterName)
		}

		// 除了 exist, not_exist, empty, not_empty 外需要校验 value_from
		if _, ok := cond.NotRequiredValueOperationMap[cfg.Operation]; !ok {
			if cfg.ValueFrom != vopt.ValueFrom_Const {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_ValueFrom).
					WithErrorDetails(fmt.Sprintf("condition does not support value_from type('%s')", cfg.ValueFrom))
			}
		}
	}

	switch cfg.Operation {
	case cond.OperationEq, cond.OperationNotEq, cond.OperationGt, cond.OperationGte,
		cond.OperationLt, cond.OperationLte, cond.OperationLike, cond.OperationNotLike,
		cond.OperationRegex, cond.OperationMatch, cond.OperationMatchPhrase, cond.OperationCurrent:
		// 右侧值为单个值
		_, ok := cfg.Value.([]interface{})
		if ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a single value", cfg.Operation))
		}

		if cfg.Operation == cond.OperationLike || cfg.Operation == cond.OperationNotLike ||
			cfg.Operation == cond.OperationPrefix || cfg.Operation == cond.OperationNotPrefix {
			_, ok := cfg.Value.(string)
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
					WithErrorDetails("[like not_like prefix not_prefix] operation's value should be a string")
			}
		}

		if cfg.Operation == cond.OperationRegex {
			val, ok := cfg.Value.(string)
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
					WithErrorDetails("[regex] operation's value should be a string")
			}

			_, err := regexp2.Compile(val, regexp2.RE2)
			if err != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
					WithErrorDetails(fmt.Sprintf("[regex] operation regular expression error: %s", err.Error()))
			}

		}

	case cond.OperationIn, cond.OperationNotIn:
		// 当 operation 是 in, not_in 时，value 为任意基本类型的数组，且长度大于等于1；
		_, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("[in not_in] operation's value must be an array")
		}

		if len(cfg.Value.([]interface{})) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("[in not_in] operation's value should contains at least 1 value")
		}
	case cond.OperationRange, cond.OperationOutRange, cond.OperationBetween:
		// 当 operation 是 range 时，value 是个由范围的下边界和上边界组成的长度为 2 的数值型数组
		// 当 operation 是 out_range 时，value 是个长度为 2 的数值类型的数组，查询的数据范围为 (-inf, value[0]) || [value[1], +inf)
		v, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("[range, out_range, between] operation's value must be an array")
		}

		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("[range, out_range, between] operation's value must contain 2 values")
		}
	case cond.OperationBefore:
		// before时, 长度为2的数组，第一个值为时间长度，数值型；第二个值为时间单位，字符串
		v, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("[before] operation's value must be an array")
		}

		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("[before] operation's value must contain 2 values")
		}
		_, err := conv.AssertFloat64(v[0])
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("[before] operation's first value should be a number")
		}

		_, ok = v[1].(string)
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("[before] operation's second value should be a string")
		}
	}

	switch cfg.Operation {
	case cond.OperationAnd, cond.OperationOr:
		for _, subCond := range cfg.SubConds {
			err := validateCond(ctx, subCond, fieldsMap)
			if err != nil {
				return err
			}
		}
	default:
		// 除 * 之外的过滤字段可以在视图字段列表里
		if cfg.Name != "*" {
			cField, ok := fieldsMap[cfg.Name]
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
					WithDescription(map[string]any{"FieldName": cfg.Name}).
					WithErrorDetails(fmt.Sprintf("Filter field '%s' is not in view fields list", cfg.Name))
			}

			fieldType := cField.Type

			// 字段类型为空的字段不支持过滤查询
			// if fieldType == "" {
			// 	return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			// 		WithErrorDetails("Empty type fields do not support filtering")
			// }

			// binary 类型的字段不支持过滤
			if fieldType == dtype.DataType_Binary {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
					WithErrorDetails("Binary fields do not support filtering")
			}

			// empty, not_empty 的字段类型必须为 string
			if cfg.Operation == cond.OperationEmpty || cfg.Operation == cond.OperationNotEmpty {
				if !dtype.DataType_IsString(fieldType) {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
						WithDescription(map[string]any{"FieldName": cfg.Name, "FieldType": fieldType, "Operation": cfg.Operation}).
						WithErrorDetails("Filter field must be of string type when using 'empty' or 'not_empty' operation")
				}
			}
		} else {
			// 如果字段为 *，则只允许使用 match 和 match_phrase 操作符
			if cfg.Operation != cond.OperationMatch && cfg.Operation != cond.OperationMatchPhrase {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
					WithDescription(map[string]any{"FieldName": cfg.Name, "FieldType": "", "Operation": cfg.Operation}).
					WithErrorDetails("Filter field '*' only supports 'match' and 'match_phrase' operations")
			}
		}
	}

	return nil
}

// 替换 SQL 中的 {{\.node[\w-]+}} 占位符为带包裹符的表名
func replaceTablePlaceholders(sqlStr, quote string) string {
	// 正则模式：匹配 {{.表名}}，表名允许字母、数字、下划线、连字符
	// 分组 1 用于捕获表名（如从 {{.node_w3uy-}} 中捕获 node_w3uy-）
	re := regexp.MustCompile(`\{\{\.([a-zA-Z0-9_-]+)\}\}`)

	// 对每个匹配的占位符执行替换逻辑
	return re.ReplaceAllStringFunc(sqlStr, func(match string) string {
		// 提取捕获组（表名）
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 2 {
			// 若匹配格式异常（理论上不会触发），返回原始占位符避免破坏 SQL
			return match
		}
		tableName := submatches[1] // 捕获组 1 即表名

		// 用包裹符包裹表名（如 `node_w3uy-`）
		return fmt.Sprintf("%s%s%s", quote, tableName, quote)
	})
}

// 添加辅助函数检查字段列表中是否包含 *
// func containsAsterisk(fields []string) bool {
// 	for _, field := range fields {
// 		if field == "*" {
// 			return true
// 		}
// 	}
// 	return false
// }

// 对 _source的数据平铺展示
func flattenWithPickField(top bool, prefix string, src any, dest map[string]any, fieldsMap map[string]*cond.ViewField) error {
	assign := func(newKey string, val any) error {
		switch value := val.(type) {
		case map[string]any:
			// 保留值为{}的字段
			if len(value) == 0 {
				if _, ok := fieldsMap[newKey]; ok {
					dest[newKey] = value
				}

				break
			}

			if err := flattenWithPickField(false, newKey, value, dest, fieldsMap); err != nil {
				return err
			}

		case []any:
			// 保留值为[]的字段
			if len(value) == 0 {
				if _, ok := fieldsMap[newKey]; ok {
					dest[newKey] = value
				}

				break
			}

			switch value[0].(type) {
			// 如果数组的元素是map或数组，继续向下展开
			case map[string]any, []any:
				if err := flattenWithPickField(false, newKey, value, dest, fieldsMap); err != nil {
					return err
				}
			default:
				if _, ok := fieldsMap[newKey]; ok {
					dest[newKey] = value
				}
			}

		default:
			if existVal, ok := dest[newKey]; !ok {
				if _, ok := fieldsMap[newKey]; ok {
					dest[newKey] = value
				}

			} else {
				// 如果展开后的字段已存在，则将值存成数组
				vals := make([]any, 0)

				switch existedValue := existVal.(type) {
				case []any:
					existedValue = append(existedValue, value)
					if _, ok := fieldsMap[newKey]; ok {
						dest[newKey] = existedValue
					}
				default:
					vals = append(vals, existedValue, value)
					if _, ok := fieldsMap[newKey]; ok {
						dest[newKey] = vals
					}
				}

			}
		}

		return nil

	}

	switch nested := src.(type) {
	case map[string]any:
		for key, val := range nested {
			newKey := joinKey(top, prefix, key)
			err := assign(newKey, val)
			if err != nil {
				return err
			}
		}
	case []any:
		for _, val := range nested {
			newKey := joinKey(top, prefix, "")
			err := assign(newKey, val)
			if err != nil {
				return err
			}

		}
	default:
		return errors.New("not a valid input: map or slice")
	}

	return nil
}

// 对 _source的数据平铺展示
func flatten(top bool, prefix string, src any, dest map[string]any) error {
	assign := func(newKey string, val any) error {
		switch value := val.(type) {
		case map[string]any:
			// 保留值为{}的字段
			if len(value) == 0 {
				// if _, ok := fieldsMap[newKey]; ok {
				dest[newKey] = value
				// }

				break
			}

			if err := flatten(false, newKey, value, dest); err != nil {
				return err
			}

		case []any:
			// 保留值为[]的字段
			if len(value) == 0 {
				// if _, ok := fieldsMap[newKey]; ok {
				dest[newKey] = value
				// }

				break
			}

			switch value[0].(type) {
			// 如果数组的元素是map或数组，继续向下展开
			case map[string]any, []any:
				if err := flatten(false, newKey, value, dest); err != nil {
					return err
				}
			default:
				// if _, ok := fieldsMap[newKey]; ok {
				dest[newKey] = value
				// }
			}

		default:
			if existVal, ok := dest[newKey]; !ok {
				// if _, ok := fieldsMap[newKey]; ok {
				dest[newKey] = value
				// }

			} else {
				// 如果展开后的字段已存在，则将值存成数组
				vals := make([]any, 0)

				switch existedValue := existVal.(type) {
				case []any:
					existedValue = append(existedValue, value)
					// if _, ok := fieldsMap[newKey]; ok {
					dest[newKey] = existedValue
					// }
				default:
					vals = append(vals, existedValue, value)
					// if _, ok := fieldsMap[newKey]; ok {
					dest[newKey] = vals
					// }
				}

			}
		}

		return nil

	}

	switch nested := src.(type) {
	case map[string]any:
		for key, val := range nested {
			newKey := joinKey(top, prefix, key)
			err := assign(newKey, val)
			if err != nil {
				return err
			}
		}
	case []any:
		for _, val := range nested {
			newKey := joinKey(top, prefix, "")
			err := assign(newKey, val)
			if err != nil {
				return err
			}

		}
	default:
		return errors.New("not a valid input: map or slice")
	}

	return nil
}

// // 修正 columns 和视图字段列表对齐，解决字段重复问题
// func fixColumns(view *interfaces.DataView, columns []ast.Node) ([]ast.Node, error){
// 	if len(view.Fields) == 0  || len(view.FieldsMap) ==  0 {
// 		return columns, fmt.Errorf("")
// 	}

// 	for _, col := range columns {
// 		fieldName, _ := col.Get("name").String()
// 		fieldType, _ := col.Get("type").String()

// 		f := &cond.ViewField{
// 			Name:         fieldName,
// 			DisplayName:  fieldName,
// 			OriginalName: fieldName,
// 			Type:         fieldType,
// 		}
// 		view.FieldsMap[fieldName] = f
// 		view.Fields = append(view.Fields, f)
// 	}
// }
