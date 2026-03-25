package logic_view

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/dlclark/regexp2"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/catalog"
	"vega-backend/logics/connectors"
	"vega-backend/logics/connectors/factory"
	"vega-backend/logics/filter_condition"
	"vega-backend/logics/permission"
	"vega-backend/logics/resource"
)

var (
	lvServiceOnce sync.Once
	lvService     interfaces.LogicViewService
)

type logicViewService struct {
	appSetting *common.AppSetting
	cs         interfaces.CatalogService
	rs         interfaces.ResourceService
	ps         interfaces.PermissionService
}

// NewLogicViewService creates a new ResourceDataService.
func NewLogicViewService(appSetting *common.AppSetting) interfaces.LogicViewService {
	lvServiceOnce.Do(func() {
		lvService = &logicViewService{
			appSetting: appSetting,
			cs:         catalog.NewCatalogService(appSetting),
			rs:         resource.NewResourceService(appSetting),
			ps:         permission.NewPermissionService(appSetting),
		}
	})
	return lvService
}

func (lvs *logicViewService) Query(ctx context.Context, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Query logic view")
	defer span.End()

	logger.Debugf("Query logic view, resourceID: %s, params: %v",
		resource.ID, params)

	switch resource.LogicType {
	case interfaces.LogicType_Derived:
		return lvs.queryDerivedLogicView(ctx, resource, params)
	case interfaces.LogicType_Composite:
		return lvs.queryCompositeLogicView(ctx, resource, params)
	default:
		return nil, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
			WithErrorDetails(fmt.Sprintf("The logic type of the custom view '%s' is not supported", resource.ID))
	}
}

func (lvs *logicViewService) queryDerivedLogicView(ctx context.Context, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query derived logic view")
	defer span.End()

	var inputNode *interfaces.LogicDefinitionNode
	for _, node := range resource.LogicDefinition {
		if node.Type == interfaces.LogicDefinitionNodeType_Resource {
			inputNode = node
			break
		}
	}

	var nodeCfg interfaces.ResourceNodeCfg
	if err := mapstructure.Decode(inputNode.Config, &nodeCfg); err != nil {
		span.SetStatus(codes.Error, "Decode resource node config failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to decode resource node config: %v", err))
	}
	fromResourceFilterCond := nodeCfg.Filters

	fromResource, err := lvs.rs.GetByID(ctx, nodeCfg.ResourceID)
	if err != nil {
		span.SetStatus(codes.Error, "Get source resource failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to get source resource %s: %v", nodeCfg.ResourceID, err))
	}
	if fromResource == nil {
		span.SetStatus(codes.Error, "Source resource not found")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound).
			WithErrorDetails(fmt.Sprintf("source resource %s not found", nodeCfg.ResourceID))
	}

	catalog, err := lvs.cs.GetByID(ctx, fromResource.CatalogID, true)
	if err != nil {
		span.SetStatus(codes.Error, "Get catalog failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to get catalog: %v", err))
	}
	if catalog == nil {
		span.SetStatus(codes.Error, "Catalog not found")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_CatalogNotFound).
			WithErrorDetails(fmt.Sprintf("catalog %s not found", fromResource.CatalogID))
	}

	fieldMap := map[string]*interfaces.Property{}
	outputFields := make([]string, 0, len(resource.SchemaDefinition))
	for _, prop := range resource.SchemaDefinition {
		fieldMap[prop.Name] = prop
		outputFields = append(outputFields, prop.Name)
	}
	params.OutputFields = outputFields

	// 合并资源和查询的 FilterCondCfg, 需要判断下是否为nil
	var mergedFilterCond *interfaces.FilterCondCfg
	if fromResourceFilterCond != nil && params.FilterCondCfg != nil {
		mergedFilterCond = &interfaces.FilterCondCfg{
			Operation: filter_condition.OperationAnd,
			SubConds:  []*interfaces.FilterCondCfg{fromResourceFilterCond, params.FilterCondCfg},
		}
	} else if fromResourceFilterCond != nil {
		mergedFilterCond = fromResourceFilterCond
	} else if params.FilterCondCfg != nil {
		mergedFilterCond = params.FilterCondCfg
	}

	actualFilterCond, err := filter_condition.NewFilterCondition(ctx, mergedFilterCond, fieldMap)
	if err != nil {
		span.SetStatus(codes.Error, "Create filter condition failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(err.Error())
	}
	params.ActualFilterCond = actualFilterCond

	// 交给 querySingleSourceData Connector 处理 SQL push-down
	return querySingleSourceData(ctx, catalog, fromResource, params)
}

func (lv *logicViewService) queryCompositeLogicView(ctx context.Context, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query composite logic view")
	defer span.End()

	return nil, 0, nil
}

func querySingleSourceData(ctx context.Context, catalog *interfaces.Catalog, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Query data")
	defer span.End()

	logger.Debugf("QueryData, resourceID: %s, catalogID: %s, params: %v",
		resource.ID, resource.CatalogID, params)

	connector, err := factory.GetFactory().CreateConnectorInstance(ctx, catalog.ConnectorType, catalog.ConnectorCfg)
	if err != nil {
		span.SetStatus(codes.Error, "Create connector failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to create connector: %v", err))
	}

	if err := connector.Connect(ctx); err != nil {
		span.SetStatus(codes.Error, "Connect to data source failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to connect to data source: %v", err))
	}
	defer connector.Close(ctx)

	switch resource.Category {
	case interfaces.ResourceCategoryTable:
		tableConnector, ok := connector.(connectors.TableConnector)
		if !ok {
			span.SetStatus(codes.Error, "Connector does not support table operations")
			return nil, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
				WithErrorDetails(fmt.Sprintf("connector %s does not support table operations", catalog.ConnectorType))
		}

		result, err := tableConnector.ExecuteQuery(ctx, resource, params)
		if err != nil {
			span.SetStatus(codes.Error, "Execute query failed")
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
				WithErrorDetails(fmt.Sprintf("failed to execute query: %v", err))
		}
		return result.Rows, result.Total, nil

	default:
		span.SetStatus(codes.Error, "Connector does not support table operations")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
			WithErrorDetails(connector.GetCategory())
	}

}

// 预览的时候校验逻辑视图配置
func (lvs *logicViewService) validateLogicDefinition(ctx context.Context, view *interfaces.LogicView) error {
	if view.LogicDefinition == nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails("logic definition is empty")
	}

	// 节点数不能超过20
	if len(view.LogicDefinition) > 20 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails("logic definition node count exceeds 20")
	}

	nodeMap := make(map[string]struct{})
	for _, ds := range view.LogicDefinition {
		nodeMap[ds.ID] = struct{}{}
	}

	for _, node := range view.LogicDefinition {
		switch node.Type {
		case interfaces.LogicDefinitionNodeType_Resource:
			_, err := lvs.validateViewNode(ctx, node)
			if err != nil {
				return err
			}
		case interfaces.LogicDefinitionNodeType_Output:
			err := validateOutputNode(ctx, node, nodeMap)
			if err != nil {
				return err
			}
		default:
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("the logic definition node type is invalid")
		}
	}

	view.IsSingleSource = false

	return nil
}

func (lvs *logicViewService) validateViewNode(ctx context.Context, node *interfaces.LogicDefinitionNode) (*interfaces.LogicView, error) {
	// 视图节点输入节点必须为空
	if len(node.Inputs) != 0 {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The view node must have no input node")
	}

	var cfg interfaces.ResourceNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("decode view node config failed, %v", err))
	}

	// fieldsMap 是字段name和字段的映射
	fieldsMap := make(map[string]*interfaces.ViewProperty)
	for _, field := range node.OutputFields {
		fieldsMap[field.Name] = field
	}

	// 校验过滤条件
	httpErr := validateCond(ctx, cfg.Filters, fieldsMap)
	if httpErr != nil {
		return nil, httpErr
	}

	return nil, nil
}

func validateOutputNode(ctx context.Context, node *interfaces.LogicDefinitionNode, nodeMap map[string]struct{}) error {
	// 输入节点只能有一个
	if len(node.Inputs) != 1 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The output node must have one input node")
	}

	// 校验输入节点是否存在
	inputNode := node.Inputs[0]
	if _, ok := nodeMap[inputNode]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails(fmt.Sprintf("The output node input_node '%s' is not exist", inputNode))
	}

	// 校验name不能重复，display_name 不能重复, original_name 可以重复
	nameMap := make(map[string]struct{})
	displayNameMap := make(map[string]struct{})
	for _, field := range node.OutputFields {
		if _, ok := nameMap[field.Name]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The output node field name is repeated")
		}
		nameMap[field.Name] = struct{}{}

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
func validateCond(ctx context.Context, cfg *interfaces.FilterCondCfg, fieldsMap map[string]*interfaces.ViewProperty) error {
	if cfg == nil {
		return nil
	}

	// 判断过滤器是否为空对象 {}
	if cfg.Name == "" && cfg.Operation == "" && len(cfg.SubConds) == 0 && cfg.ValueFrom == "" && cfg.Value == nil {
		return nil
	}

	// 过滤条件字段不允许 __id 和 __routing
	if cfg.Name == "__id" || cfg.Name == "__routing" {
		return rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("The filter field '__id' and '__routing' is not allowed")
	}

	// 过滤操作符
	if cfg.Operation == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest)
	}

	_, exists := filter_condition.OperationMap[cfg.Operation]
	if !exists {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails(fmt.Sprintf("unsupport condition operation %s", cfg.Operation))
	}

	switch cfg.Operation {
	case filter_condition.OperationAnd, filter_condition.OperationOr:
		// 子过滤条件不能超过10个
		if len(cfg.SubConds) > interfaces.MaxSubCondition {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
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
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest)
		}

		// 除了 exist, not_exist, empty, not_empty 外需要校验 value_from
		if _, ok := map[string]struct{}{filter_condition.OperationExist: {}, filter_condition.OperationNotExist: {}, filter_condition.OperationEmpty: {}, filter_condition.OperationNotEmpty: {}}[cfg.Operation]; !ok {
			if cfg.ValueFrom != interfaces.ValueFrom_Const {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
					WithErrorDetails(fmt.Sprintf("condition does not support value_from type('%s')", cfg.ValueFrom))
			}
		}
	}

	switch cfg.Operation {
	case filter_condition.OperationEqual, filter_condition.OperationNotEqual, filter_condition.OperationGt, filter_condition.OperationGte,
		filter_condition.OperationLt, filter_condition.OperationLte, filter_condition.OperationLike, filter_condition.OperationNotLike,
		filter_condition.OperationRegex, filter_condition.OperationMatch, filter_condition.OperationMatchPhrase, filter_condition.OperationCurrent:
		// 右侧值为单个值
		_, ok := cfg.Value.([]interface{})
		if ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a single value", cfg.Operation))
		}

		if cfg.Operation == filter_condition.OperationLike || cfg.Operation == filter_condition.OperationNotLike ||
			cfg.Operation == filter_condition.OperationPrefix || cfg.Operation == filter_condition.OperationNotPrefix {
			// 如果有 real_value 则跳过 value 的校验
			if cfg.RealValue == nil {
				_, ok := cfg.Value.(string)
				if !ok {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
						WithErrorDetails("[like not_like prefix not_prefix] operation's value should be a string")
				}
			} else {
				_, ok := cfg.RealValue.(string)
				if !ok {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
						WithErrorDetails("[like not_like prefix not_prefix] operation's real_value should be a string")
				}
			}
		}

		if cfg.Operation == filter_condition.OperationRegex {
			val, ok := cfg.Value.(string)
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
					WithErrorDetails("[regex] operation's value should be a string")
			}

			_, err := regexp2.Compile(val, regexp2.RE2)
			if err != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
					WithErrorDetails(fmt.Sprintf("[regex] operation regular expression error: %s", err.Error()))
			}

		}

	case filter_condition.OperationIn, filter_condition.OperationNotIn:
		// 当 operation 是 in, not_in 时，value 为任意基本类型的数组，且长度大于等于1；
		_, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails("[in not_in] operation's value must be an array")
		}

		if len(cfg.Value.([]interface{})) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails("[in not_in] operation's value should contains at least 1 value")
		}
	case filter_condition.OperationRange, filter_condition.OperationOutRange, filter_condition.OperationBetween:
		// 当 operation 是 range 时，value 是个由范围的下边界和上边界组成的长度为 2 的数值型数组
		// 当 operation 是 out_range 时，value 是个长度为 2 的数值类型的数组，查询的数据范围为 (-inf, value[0]) || [value[1], +inf)
		v, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails("[range, out_range, between] operation's value must be an array")
		}

		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails("[range, out_range, between] operation's value must contain 2 values")
		}
	case filter_condition.OperationBefore:
		// before时, 长度为2的数组，第一个值为时间长度，数值型；第二个值为时间单位，字符串
		v, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails("[before] operation's value must be an array")
		}

		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails("[before] operation's value must contain 2 values")
		}
		/*
				_, err := conv.AssertFloat64(v[0])
				if err != nil {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
						WithErrorDetails("[before] operation's first value should be a number")
				}

				_, ok = v[1].(string)
				if !ok {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
						WithErrorDetails("[before] operation's second value should be a string")
			}
		*/
	}

	switch cfg.Operation {
	case filter_condition.OperationAnd, filter_condition.OperationOr:
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
			if fieldType == interfaces.DataType_Binary {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
					WithErrorDetails("Binary fields do not support filtering")
			}

			// empty, not_empty 的字段类型必须为 string
			if cfg.Operation == filter_condition.OperationEmpty || cfg.Operation == filter_condition.OperationNotEmpty {
				if !interfaces.DataType_IsString(fieldType) {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
						WithDescription(map[string]any{"FieldName": cfg.Name, "FieldType": fieldType, "Operation": cfg.Operation}).
						WithErrorDetails("Filter field must be of string type when using 'empty' or 'not_empty' operation")
				}
			}
		} else {
			// 如果字段为 *，则只允许使用 match 和 match_phrase 操作符
			if cfg.Operation != filter_condition.OperationMatch && cfg.Operation != filter_condition.OperationMatchPhrase {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
					WithDescription(map[string]any{"FieldName": cfg.Name, "FieldType": "", "Operation": cfg.Operation}).
					WithErrorDetails("Filter field '*' only supports 'match' and 'match_phrase' operations")
			}
		}
	}

	return nil
}
