package logic_view

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
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
	case interfaces.ResourceCategoryTable, interfaces.ResourceCategoryLogicView:
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

// // 视图数据预览
//
//	func (lvs *logicViewService) Simulate(ctx context.Context, query *interfaces.LogicViewSimulateQuery) (*interfaces.ViewUniResponseV2, error) {
//		ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Simulate view data")
//		defer span.End()
//
//		// 决策权限, 预览的时候还没有视图id，此时的预览校验用新建或者编辑
//		ops, err := lvs.ps.GetResourcesOperations(ctx, interfaces.RESOURCE_TYPE_RESOURCE,
//			[]string{interfaces.RESOURCE_ID_ALL})
//		if err != nil {
//			return nil, err
//		}
//
//		if len(ops) != 1 {
//			// 无权限
//			return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
//				WithErrorDetails("Access denied: insufficient permissions for data view's create or modify operation")
//		}
//		// 从 ops 里找新建或编辑的权限
//		for _, op := range ops[0].Operations {
//			if op != interfaces.OPERATION_TYPE_CREATE && op != interfaces.OPERATION_TYPE_MODIFY {
//				return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
//					WithErrorDetails("Access denied: insufficient permissions for data view's create or modify operation")
//			}
//		}
//
//		view := &interfaces.LogicView{
//			Type:           query.Type,
//			QueryType:      query.QueryType,
//			TechnicalName:  query.TechnicalName,
//			DataSourceType: query.DataSourceType,
//			DataSourceID:   query.DataSourceID,
//			FileName:       query.FileName,
//			ExcelConfig:    query.ExcelConfig,
//			DataScope:      query.LogicDefinition,
//			Fields:         query.Fields,
//		}
//
//		// query.NeedTotal = true
//		// 设置预览的默认format为flat
//		if query.Format == "" {
//			query.Format = interfaces.Format_Flat
//		}
//
//		switch query.QueryType {
//		case interfaces.QueryType_IndexBase:
//			return lvs.SimulateByIndexBase(ctx, query, view)
//		case interfaces.QueryType_DSL:
//			return lvs.SimulateByDSL(ctx, query, view)
//		case interfaces.QueryType_SQL:
//			return lvs.SimulateBySQL(ctx, query, view)
//		default:
//			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
//				WithErrorDetails("query type must be DSL or SQL or IndexBase")
//		}
//	}
//
// // SQL类视图数据预览
// func (lvs *logicViewService) SimulateBySQL(ctx context.Context, query *interfaces.LogicViewSimulateQuery,
//
//		view *interfaces.LogicView) (*interfaces.ViewUniResponseV2, error) {
//		ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Simulate view data by SQL")
//		defer span.End()
//
//		switch view.Type {
//		case interfaces.ViewType_Atomic:
//			if query.DataSourceID == "" {
//				span.SetStatus(codes.Error, "Data source ID is empty")
//				return nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
//					rest.PublicError_BadRequest).WithErrorDetails("Data source ID is empty")
//			}
//
//			// 获取数据源信息
//			dataSource, err := lvs.vdsAccess.GetDataSourceByID(ctx, query.DataSourceID)
//			if err != nil {
//				span.SetStatus(codes.Error, "Get data source by ID failed")
//				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
//					WithErrorDetails(fmt.Sprintf("Get data source by ID failed, %v", err))
//			}
//
//			// 构造视图的meta_table_name
//			catalogName := dataSource.BinData.CatalogName
//			schemaName := dataSource.BinData.Schema
//			// 先用schema，没有再用database
//			if schemaName == "" {
//				schemaName = dataSource.BinData.DataBaseName
//			}
//			// database也没有使用默认值 default
//			if schemaName == "" {
//				schemaName = interfaces.DefaultSchema
//			}
//
//			view.MetaTableName = fmt.Sprintf("%s.%s.%s", catalogName, common.QuotationMark(schemaName),
//				common.QuotationMark(view.TechnicalName))
//
//		case interfaces.ViewType_Custom:
//			_, _, httpErr := validateDataScope(ctx, dvs, view)
//			if httpErr != nil {
//				span.SetStatus(codes.Error, "Validate logic definition failed")
//				return nil, httpErr
//			}
//		}
//
//		// 将字段转为map
//		viewFieldsMap := make(map[string]*interfaces.ViewProperty)
//		for _, field := range view.Fields {
//			// init field path
//			field.InitFieldPath()
//			viewFieldsMap[field.Name] = field
//		}
//		view.FieldsMap = viewFieldsMap
//
//		resBytes, total, err := lvs.queryBySQL(ctx, query, view)
//		if err != nil {
//			span.SetStatus(codes.Error, "Query by SQL failed")
//			return nil, err
//		}
//
//		// 转成视图统一结构
//		res, httpErr := convertToViewUniResponse(ctx, query, view, resBytes, total)
//		if httpErr != nil {
//			o11y.Error(ctx, httpErr.Error())
//			span.SetStatus(codes.Error, "Convert to view uniResponse failed")
//			return nil, httpErr
//		}
//
//		span.SetStatus(codes.Ok, "")
//		return res, nil
//	}
//
// // DSL类视图数据预览
// func (lvs *logicViewService) SimulateByDSL(ctx context.Context, query *interfaces.LogicViewSimulateQuery,
//
//		view *interfaces.LogicView) (*interfaces.ViewUniResponseV2, error) {
//		ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Simulate view data by SQL")
//		defer span.End()
//
//		switch view.Type {
//		case interfaces.ViewType_Atomic:
//			if query.DataSourceID == "" {
//				span.SetStatus(codes.Error, "Data source ID is empty")
//				return nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
//					rest.PublicError_BadRequest).WithErrorDetails("Data source ID is empty")
//			}
//
//			// 获取数据源信息
//			dataSource, err := lvs.vdsAccess.GetDataSourceByID(ctx, query.DataSourceID)
//			if err != nil {
//				span.SetStatus(codes.Error, "Get data source by ID failed")
//				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
//					WithErrorDetails(fmt.Sprintf("Get data source by ID failed, %v", err))
//			}
//
//			// 补充视图的 catalog
//			view.DataSourceCatalog = dataSource.BinData.CatalogName
//
//		case interfaces.ViewType_Custom:
//			_, _, httpErr := validateDataScope(ctx, dvs, view)
//			if httpErr != nil {
//				span.SetStatus(codes.Error, "Validate logic definition failed")
//				return nil, httpErr
//			}
//		}
//
//		// 将字段转为map
//		viewFieldsMap := make(map[string]*interfaces.ViewProperty)
//		for _, field := range view.Fields {
//			// init field path
//			field.InitFieldPath()
//			viewFieldsMap[field.Name] = field
//		}
//		view.FieldsMap = viewFieldsMap
//
//		resBytes, total, err := lvs.queryByDSL(ctx, query, view)
//		if err != nil {
//			span.SetStatus(codes.Error, "Query by SQL failed")
//			return nil, err
//		}
//
//		// 转成视图统一结构
//		res, httpErr := convertToViewUniResponse(ctx, query, view, resBytes, total)
//		if httpErr != nil {
//			o11y.Error(ctx, httpErr.Error())
//			span.SetStatus(codes.Error, "Convert to view uniResponse failed")
//			return nil, httpErr
//		}
//
//		span.SetStatus(codes.Ok, "")
//		return res, nil
//	}
//
// // 获取单个视图数据
//
//	func (lvs *logicViewService) GetSingleViewData(ctx context.Context, viewID string, query interfaces.ViewQueryInterface) (*interfaces.ViewUniResponseV2, error) {
//		ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get single view data")
//		defer span.End()
//
//		span.SetAttributes(attr.Key("view_id").String(viewID))
//
//		// 决策当前视图id的数据查询权限
//		hasPermission, err := lvs.ps.CheckPermissionWithResult(ctx, interfaces.Resource{
//			ID:   viewID,
//			Type: interfaces.RESOURCE_TYPE_DATA_VIEW,
//		}, []string{interfaces.OPERATION_TYPE_DATA_QUERY})
//
//		if err != nil {
//			return nil, err
//		}
//
//		// 如果有data_query权限，返回视图的全量数据
//		// 如果没有data_query权限，则获取视图下的所有行列规则，
//		// 决策当前用户具有rule_apply权限的规则，执行规则过滤查询
//		if !hasPermission {
//			// 获取视图下的所有行列规则
//			rowColumnRules, err := lvs.dvrcrAccess.GetRulesByViewID(ctx, viewID)
//			if err != nil {
//				span.SetStatus(codes.Error, "Get row column rules by view ID failed")
//				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
//					WithErrorDetails(err.Error())
//			}
//
//			// 过滤视图下的行列规则，返回当前用户具有rule_apply权限的规则
//			filteredRules, httpErr := lvs.FilterRowColumnRules(ctx, rowColumnRules)
//			if httpErr != nil {
//				span.SetStatus(codes.Error, "Filter row column rules failed")
//				return nil, httpErr
//			}
//
//			if len(filteredRules) == 0 {
//				errDetails := fmt.Sprintf("Neither data query permission nor row column rules with rule_apply permission for view ID %s", viewID)
//				span.SetStatus(codes.Error, errDetails)
//				return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
//					WithErrorDetails(errDetails)
//			}
//
//			// 设置查询参数中的行列规则
//			query.SetRowColumnRules(filteredRules)
//		}
//
//		// data-model服务会检查基础权限(data_view_id,'view_detail')
//		view, httpErr := lvs.GetDataViewByID(ctx, viewID, true)
//		if httpErr != nil {
//			span.SetStatus(codes.Error, "Get data view by ID failed")
//			return nil, httpErr
//		}
//
//		// 查询数据
//		resBytes, total, httpErr := lvs.querySingleViewData(ctx, query, view)
//		if httpErr != nil {
//			span.SetStatus(codes.Error, "Query single view data failed")
//			return nil, httpErr
//		}
//
//		// 转成视图统一结构
//		res, httpErr := convertToViewUniResponse(ctx, query, view, resBytes, total)
//		if httpErr != nil {
//			o11y.Error(ctx, httpErr.Error())
//			span.SetStatus(codes.Error, "Convert to view uniResponse failed")
//			return nil, httpErr
//		}
//
//		span.SetStatus(codes.Ok, "")
//		return res, nil
//	}
//
// // 获取单个视图对象信息
//
//	func (lvs *logicViewService) GetDataViewByID(ctx context.Context, viewID string, includeDataScopeView bool) (*interfaces.LogicView, error) {
//		views, err := lvs.dvAccess.GetDataViewsByIDs(ctx, viewID, includeDataScopeView)
//		if err != nil {
//			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//				verrors.Uniquery_DataView_InternalError_GetDataViewByIDFailed).WithErrorDetails(err.Error())
//		}
//
//		if len(views) == 0 {
//			logger.Errorf("Data view %s not found", viewID)
//			return nil, rest.NewHTTPError(ctx, http.StatusNotFound,
//				verrors.Uniquery_DataView_DataViewNotFound).WithErrorDetails(fmt.Sprintf("view %s not found", viewID))
//		}
//
//		view := views[0]
//
//		// 补充自定义视图的来源原子视图是否来自同一个数据源
//		if includeDataScopeView && view.Type == interfaces.ViewType_Custom {
//			dataSourceIDMap := make(map[string]struct{})
//			for _, node := range view.LogicDefinition {
//				if node.Type == interfaces.LogicDefinitionNodeType_View {
//					var viewNodeConfig interfaces.ViewNodeCfg
//					err := mapstructure.Decode(node.Config, &viewNodeConfig)
//					if err != nil {
//						logger.Errorf("Decode view node config failed, err: %v", err)
//						return nil, err
//					}
//
//					if viewNodeConfig.View == nil {
//						logger.Errorf("View node config view is nil")
//						return nil, fmt.Errorf("view node config view is nil")
//					}
//
//					dataSourceIDMap[viewNodeConfig.View.DataSourceID] = struct{}{}
//					view.LogicDefinitionAdvancedParams.LogicDefinitionDataSourceID = viewNodeConfig.View.DataSourceID
//				}
//			}
//
//			view.IsSingleSource = len(dataSourceIDMap) == 1
//		}
//
//		return view, nil
//	}
//
// // 获取单个视图对象信息
//
//	func (lvs *logicViewService) GetDataViewsByIDs(ctx context.Context, viewIDs []string, includeDataScopeView bool) (map[string]*interfaces.LogicView, error) {
//		if len(viewIDs) == 0 {
//			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
//				WithErrorDetails("view id list is empty")
//		}
//
//		ids := strings.Join(viewIDs, ",")
//
//		views, err := lvs.dvAccess.GetDataViewsByIDs(ctx, ids, includeDataScopeView)
//		if err != nil {
//			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//				verrors.Uniquery_DataView_InternalError_GetDataViewByIDFailed).WithErrorDetails(err.Error())
//		}
//
//		if len(views) != len(viewIDs) {
//			logger.Errorf("Data view %s not found", ids)
//			return nil, rest.NewHTTPError(ctx, http.StatusNotFound,
//				verrors.Uniquery_DataView_DataViewNotFound).WithErrorDetails(fmt.Sprintf("view %s not found", ids))
//		}
//
//		// 转成视图id和视图对象的map返回
//		viewMap := make(map[string]*interfaces.LogicView)
//		for _, view := range views {
//			viewMap[view.ViewID] = view
//		}
//
//		return viewMap, nil
//	}
//
// // 单个查询视图数据和批量查询通用的函数
// func (lvs *logicViewService) querySingleViewData(ctx context.Context, query interfaces.ViewQueryInterface,
//
//		view *interfaces.LogicView) (resBytes []byte, total int64, err error) {
//		ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Query single view data")
//		defer span.End()
//
//		globalFilters := query.GetGlobalFilters()
//		commonParams := query.GetCommonParams()
//		format := commonParams.Format
//
//		allowNonExistField := query.GetQueryParams()[interfaces.QueryParam_AllowNonExistField].(bool)
//
//		viewFieldsMap := make(map[string]*interfaces.ViewProperty)
//		for _, field := range view.Fields {
//			newField := field
//			// init field path
//			newField.InitFieldPath()
//			viewFieldsMap[newField.Name] = newField
//		}
//		// 将视图字段转成map
//		view.FieldsMap = viewFieldsMap
//
//		// 非严格模式下，如果全局过滤条件的字段不在视图字段列表里，数据返回空
//		fieldName, exist := checkConditionFieldExist(viewFieldsMap, globalFilters)
//		if !exist {
//			if allowNonExistField {
//				return []byte{}, total, nil
//			} else {
//				return []byte{}, total, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.Uniquery_DataView_InvalidFilterField_FieldNotInView).
//					WithErrorDetails(fmt.Sprintf("condition config field name '%s' must in view original fields", fieldName))
//			}
//		}
//
//		// 如果查询参数没有设置format，则默认设置成flat
//		if format == "" {
//			query.SetFormat(interfaces.Format_Flat)
//		}
//
//		switch view.QueryType {
//		case interfaces.QueryType_DSL:
//			return lvs.queryByDSL(ctx, query, view)
//		case interfaces.QueryType_SQL:
//			return lvs.queryBySQL(ctx, query, view)
//		default:
//			return nil, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.Uniquery_DataView_InvalidParameter_QueryType).
//				WithErrorDetails("query type must be DSL or SQL or IndexBase")
//		}
//	}
//
// func (lvs *logicViewService) queryByDSL(ctx context.Context, query interfaces.ViewQueryInterface,
//
//		view *interfaces.LogicView) (resBytes []byte, total int64, err error) {
//		ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Query single view data")
//		defer span.End()
//
//		commonParams := query.GetCommonParams()
//		searchAfterParams := query.GetSearchAfterParams()
//
//		// 如果没传 limit，传了 search_after 参数，设置默认limit 10000,  这处会影响拼接dsl
//		if commonParams.Limit == 0 && searchAfterParams != nil && len(searchAfterParams.SearchAfter) > 0 {
//			query.SetLimit(interfaces.SearchAfter_Limit)
//		}
//
//		// 获取索引列表, 视图 ID 到索引列表的映射
//		catalogName, indices, viewIndicesMap, err := lvs.getIndicesByView(view)
//		if err != nil {
//			span.SetStatus(codes.Error, "Get indices failed")
//			return []byte{}, total, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//				verrors.Uniquery_DataView_InternalError_GetIndicesFailed).WithErrorDetails(err.Error())
//		}
//
//		// 如果索引列表为空，则返回空数据, 不需要下面拼接dsl
//		if len(indices) == 0 {
//			span.SetStatus(codes.Ok, "No indices found")
//			return []byte{}, total, nil
//		}
//
//		// 转成 DSL
//		dsl, httpErr := buildDSL(ctx, query, view, viewIndicesMap)
//		if httpErr != nil {
//			o11y.Error(ctx, httpErr.Error())
//			span.SetStatus(codes.Error, "Convert to DSL failed")
//			return []byte{}, total, httpErr
//		}
//
//		// 记录查询vega耗时
//		startTime := time.Now()
//		defer func() {
//			elapsed := time.Since(startTime).Milliseconds()
//			logger.Infof("query vega data cost time is %dms", elapsed)
//			query.SetVegaDuration(elapsed)
//		}()
//
//		// 向vega执行dsl查询
//		fetchParams := &interfaces.FetchVegaDataParams{
//			IsSingleDataSource: isSingleDataSource(view),
//			QueryType:          interfaces.QueryType_DSL,
//			DataSourceID:       getQueryDataSourceID(view),
//			CatalogName:        catalogName,
//			TableNames:         indices,
//			Dsl:                dsl,
//		}
//		dataBatch, err := lvs.vgAccess.FetchDataNoUnmarshal(ctx, fetchParams)
//		if err != nil {
//			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.Uniquery_DataView_InternalError_FetchDataFromVegaFailed).
//				WithErrorDetails(err.Error())
//		}
//
//		// 根据NeedTotal参数决定是否查询total, vega没提供total接口，DSL查询暂不支持total
//		// if query.GetCommonParams().NeedTotal {
//		// 	total, httpErr = lvs.GetTotalByDSL(ctx, dsl, indices)
//		// 	if httpErr != nil {
//		// 		return nil, total, httpErr
//		// 	}
//		// }
//
//		span.SetStatus(codes.Ok, "")
//		return dataBatch, total, nil
//	}
//
// func (lvs *logicViewService) queryBySQL(ctx context.Context, query interfaces.ViewQueryInterface,
//
//		view *interfaces.LogicView) (resBytes []byte, total int64, err error) {
//		commonParams := query.GetCommonParams()
//
//		// 优先使用查询接口指定的 sql
//		selectSql := commonParams.SqlStr
//		if selectSql == "" {
//			if view.SQLStr != "" {
//				// 原子视图还存着sql_str
//				selectSql = view.SQLStr
//			} else {
//				// 实时生成sql
//				selectSql, err = buildViewSql(ctx, view)
//				if err != nil {
//					return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
//						WithErrorDetails(err.Error())
//				}
//			}
//		}
//
//		// 添加时间过滤
//		timeFilterSql := buildTimeFilterSql(commonParams.DateField, commonParams.Start, commonParams.End)
//		// 全局过滤条件, 全局过滤条件选择的字段要在视图列表里
//		globalFilterSql, err := buildSQLCondition(ctx, query.GetGlobalFilters(), view.Type, view.FieldsMap)
//		if err != nil {
//			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
//				WithErrorDetails(err.Error())
//		}
//		// 视图的行列规则不会应用在给指标模型生成的sql上，所以在这里添加行列规则过滤
//		rowColumnRules := query.GetRowColumnRules()
//		rowColumnRulesSQL, newFields, newFieldsMap, err := buildRowColumnRulesSQL(ctx, rowColumnRules, view)
//		if err != nil {
//			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
//				WithErrorDetails(err.Error())
//		}
//		// 更新视图字段为行列规则定义的字段
//		defer func() {
//			view.Fields = newFields
//			view.FieldsMap = newFieldsMap
//		}()
//
//		// 将全局过滤条件、时间过滤和视图sql一起拼sql，全局过滤条件、时间过滤均为可选项
//		var sqlStr string
//		var whereClauses []string
//		// 收集非空的过滤条件
//		if timeFilterSql != "" {
//			whereClauses = append(whereClauses, timeFilterSql)
//		}
//		if globalFilterSql != "" {
//			whereClauses = append(whereClauses, globalFilterSql)
//		}
//		if rowColumnRulesSQL != "" {
//			whereClauses = append(whereClauses, rowColumnRulesSQL)
//		}
//
//		builder := NewSQLBuilder(selectSql)
//		builder.AddWheres(whereClauses)
//		sqlStr = builder.Build()
//
//		// 记录查询vega耗时
//		startTime := time.Now()
//		defer func() {
//			elapsed := time.Since(startTime).Milliseconds()
//			logger.Infof("query vega data cost time is %dms", elapsed)
//			query.SetVegaDuration(elapsed)
//		}()
//
//		// 查询总数
//		if commonParams.NeedTotal {
//			countSql := buildCountSql(sqlStr)
//			logger.Infof("get total count sqlStr is %s", countSql)
//			result, err := lvs.vgAccess.FetchDataNoUnmarshal(ctx, &interfaces.FetchVegaDataParams{
//				IsSingleDataSource: isSingleDataSource(view),
//				QueryType:          interfaces.QueryType_SQL,
//				DataSourceID:       getQueryDataSourceID(view),
//				SqlStr:             countSql,
//				NextUri:            "",
//				UseSearchAfter:     false,
//			})
//			if err != nil {
//				return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.Uniquery_DataView_InternalError_FetchDataFromVegaFailed).
//					WithErrorDetails(err.Error())
//			}
//
//			// 读取count的结果
//			total, err = readCountResult(ctx, isSingleDataSource(view), result)
//			if err != nil {
//				return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
//					WithErrorDetails(err.Error())
//			}
//			logger.Infof("total count is %d", total)
//
//			// 如果总条数为 0, 不需要再查询, columns的元数据在视图对象详情里
//			if total == 0 {
//				return []byte{}, total, nil
//			}
//		}
//
//		var keys []string
//		sfParams := query.GetSearchAfterParams()
//		if sfParams != nil {
//			for _, key := range sfParams.SearchAfter {
//				keys = append(keys, fmt.Sprintf("%v", key))
//			}
//		}
//
//		finalSql := sqlStr
//		// 拼接排序
//		sortParams := prepareSQLSortParams(query.GetSortParams(), view.FieldsMap)
//		if len(sortParams) > 0 {
//			sortSql := buildSQLSortParams(sortParams)
//			if sortSql != "" {
//				finalSql = fmt.Sprintf("%s ORDER BY %s", finalSql, sortSql)
//			}
//		}
//
//		// 如果不翻页查询，并且传递了limit参数，并且sql语句里没有limit, sql里拼接上limit
//		if !commonParams.UseSearchAfter && commonParams.Limit > 0 {
//			finalSql = AddLimitIfMissing(finalSql, commonParams.Limit)
//			// finalSql = fmt.Sprintf("%s LIMIT %d", finalSql, commonParams.Limit)
//		}
//
//		logger.Infof("fetch data sqlStr is [%s]", finalSql)
//
//		timeout := query.GetQueryParams()[interfaces.QueryParam_Timeout].(time.Duration)
//		timeoutSecond := int64(timeout.Seconds())
//		nextUri := strings.Join(keys, "/")
//		fetchParams := &interfaces.FetchVegaDataParams{
//			IsSingleDataSource: isSingleDataSource(view),
//			QueryType:          interfaces.QueryType_SQL,
//			DataSourceID:       getQueryDataSourceID(view),
//			NextUri:            nextUri,
//			SqlStr:             finalSql,
//			UseSearchAfter:     commonParams.UseSearchAfter,
//			Limit:              commonParams.Limit,
//			Timeout:            timeoutSecond,
//		}
//		dataBatch, err := lvs.vgAccess.FetchDataNoUnmarshal(ctx, fetchParams)
//		if err != nil {
//			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.Uniquery_DataView_InternalError_FetchDataFromVegaFailed).
//				WithErrorDetails(err.Error())
//		}
//
//		return dataBatch, total, nil
//	}
//
//	type taskContext struct {
//		hasOutputFields   bool
//		includeIndexField bool
//		includeScoreField bool
//		finalFieldsMap    map[string]*interfaces.ViewProperty
//		results           []map[string]any
//
//		wg    sync.WaitGroup
//		errCh chan error
//	}
//
// // 将结果转成视图查询的统一格式
// func convertToViewUniResponse(ctx context.Context, query interfaces.ViewQueryInterface,
//
//		view *interfaces.LogicView, content []byte, total int64) (*interfaces.ViewUniResponseV2, error) {
//
//		includeView := query.GetQueryParams()[interfaces.QueryParam_IncludeView].(bool)
//
//		if len(content) == 0 {
//			if !includeView {
//				view = nil
//			}
//
//			// 如果不需要 total 计数，total 设为 nil, 不返回此参数
//			var totalCount *int64
//			if query.GetCommonParams().NeedTotal {
//				totalCount = &total
//			} else {
//				totalCount = nil
//			}
//
//			return &interfaces.ViewUniResponseV2{
//				PitID:       "",
//				SearchAfter: nil,
//				View:        view,
//				Entries:     []map[string]any{},
//				TotalCount:  totalCount,
//				ScrollId:    "",
//			}, nil
//		}
//
//		switch view.QueryType {
//		case interfaces.QueryType_DSL:
//			return convertToViewUniResponseByDSL(ctx, query, view, content, total)
//		case interfaces.QueryType_SQL:
//			return convertToViewUniResponseBySQL(ctx, query, view, content, total)
//		case interfaces.QueryType_IndexBase:
//			return convertToViewUniResponseByIndexBase(ctx, query, view, content, total)
//		default:
//			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
//				WithErrorDetails("view query type must be DSL or SQL or IndexBase")
//		}
//	}
//
// func convertToViewUniResponseByDSL(ctx context.Context, query interfaces.ViewQueryInterface,
//
//		view *interfaces.LogicView, content []byte, total int64) (*interfaces.ViewUniResponseV2, error) {
//
//		rootNode, err := sonic.Get(content)
//		if err != nil {
//			detail := fmt.Sprintf("SQL parse root failed, %s", err.Error())
//			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//				verrors.Uniquery_DataView_InternalError_GetDocumentsFailed).WithErrorDetails(detail)
//		}
//
//		var data []ast.Node
//		dataNode := rootNode.Get("data")
//		if dataNode.Exists() {
//			data, err = dataNode.ArrayUseNode()
//			if err != nil {
//				detail := fmt.Sprintf("SQL dataNode convert to arrayNode failed, %s", err.Error())
//				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//					verrors.Uniquery_DataView_InternalError_GetDocumentsFailed).WithErrorDetails(detail)
//			}
//		}
//
//		if len(data) == 0 {
//			return &interfaces.ViewUniResponseV2{
//				PitID:       "",
//				SearchAfter: nil,
//				View:        view,
//				Entries:     []map[string]any{},
//				TotalCount:  nil,
//				ScrollId:    "",
//			}, nil
//		}
//
//		// total 从 trace_total_hits的结果来
//		if query.GetCommonParams().NeedTotal {
//			trackTotalHitsResult, err := data[0].GetByPath("hits", "total", "value").Int64()
//			if err != nil {
//				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
//					WithErrorDetails(fmt.Sprintf("Get DSL query total value failed, err: %s", err.Error()))
//			}
//			total = trackTotalHitsResult
//		}
//
//		resp, err := commonConvertToViewDSLUniResponse(ctx, query, view, data[0], total)
//		if err != nil {
//			return nil, err
//		}
//
//		return resp, nil
//	}
//
// func commonConvertToViewDSLUniResponse(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.LogicView,
//
//		rootNode ast.Node, total int64) (*interfaces.ViewUniResponseV2, error) {
//		docs, err := rootNode.GetByPath("hits", "hits").ArrayUseNode()
//		if err != nil {
//			detail := fmt.Sprintf("DSL docsNode convert to arrayNode failed, %s", err.Error())
//			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//				verrors.Uniquery_DataView_InternalError_GetDocumentsFailed).WithErrorDetails(detail)
//		}
//
//		results := make([]map[string]interface{}, len(docs))
//		errCh := make(chan error, len(docs))
//		defer close(errCh)
//
//		var includeIndexField, includeScoreField bool
//		outputFields := query.GetCommonParams().OutputFields
//		finalFieldsMap := make(map[string]*interfaces.ViewProperty)
//		if len(outputFields) == 0 {
//			// 没有指定输出字段，默认输出视图的字段
//			finalFieldsMap = view.FieldsMap
//			includeIndexField = true
//			includeScoreField = true
//		} else {
//			// 有指定输出字段，检查是否包含 __index 和 _score 字段
//			for _, of := range outputFields {
//				if value, ok := view.FieldsMap[of]; ok {
//					finalFieldsMap[of] = value
//				} else {
//					logger.Errorf("SQL output field '%s' not found in view '%s'", of, view.ViewName)
//				}
//
//				if of == "__index" {
//					includeIndexField = true
//				}
//				if of == "_score" {
//					includeScoreField = true
//				}
//			}
//		}
//
//		taskCtx := &taskContext{
//			hasOutputFields:   len(outputFields) > 0,
//			includeIndexField: includeIndexField,
//			includeScoreField: includeScoreField,
//			finalFieldsMap:    finalFieldsMap,
//			results:           results,
//			wg:                sync.WaitGroup{},
//			errCh:             errCh,
//		}
//
//		for i, doc := range docs {
//			taskCtx.wg.Add(1)
//
//			err := viewPool.Submit(processDocByDSL(taskCtx, view, i, doc, query.GetCommonParams().Format))
//			if err != nil {
//				detail := fmt.Sprintf("DSL submit task of processing a document failed, %s", err.Error())
//				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//					verrors.Uniquery_DataView_InternalError_SubmitTaskFailed).WithErrorDetails(detail)
//			}
//		}
//		taskCtx.wg.Wait()
//
//		if len(taskCtx.errCh) > 0 {
//			err := <-taskCtx.errCh
//
//			detail := fmt.Sprintf("DSL process a document failed, %s", err.Error())
//			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//				verrors.Uniquery_DataView_InternalError_ProcessDocFailed).WithErrorDetails(detail)
//		}
//
//		var scrollId string
//		scrollIdNode := rootNode.Get("_scroll_id")
//		if scrollIdNode.Exists() {
//			scrollId, err = scrollIdNode.String()
//			if err != nil {
//				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//					verrors.Uniquery_DataView_InternalError_GetScrollIdFailed).WithErrorDetails(err.Error())
//			}
//		}
//
//		var pitID string
//		pitIDNode := rootNode.Get("pit_id")
//		if pitIDNode.Exists() {
//			pitID, err = pitIDNode.String()
//			if err != nil {
//				logger.Errorf("DSL get pit_id failed, %s", err.Error())
//				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//					verrors.Uniquery_DataView_InternalError_GetPitIdFailed).WithErrorDetails(err.Error())
//			}
//		}
//
//		var searchAfter []any
//		if len(docs) > 0 {
//			searchAfterNode := docs[len(docs)-1].Get("sort")
//			if searchAfterNode.Exists() {
//				searchAfter, err = searchAfterNode.Array()
//				if err != nil {
//					logger.Errorf("DSL get search_after failed, %s", err.Error())
//					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//						verrors.Uniquery_DataView_InternalError_GetSearchAfterValueFailed).WithErrorDetails(err.Error())
//				}
//			}
//		} else {
//			searchAfter = nil
//		}
//
//		includeView := query.GetQueryParams()[interfaces.QueryParam_IncludeView].(bool)
//		if !includeView {
//			view = nil
//		}
//
//		// 如果不需要 total 计数，total 设为 nil, 不返回此参数
//		var totalCount *int64
//		if query.GetCommonParams().NeedTotal {
//			totalCount = &total
//		} else {
//			totalCount = nil
//		}
//
//		return &interfaces.ViewUniResponseV2{
//			PitID:          pitID,
//			SearchAfter:    searchAfter,
//			View:           view,
//			Entries:        results,
//			TotalCount:     totalCount,
//			ScrollId:       scrollId,
//			VegaDurationMs: query.GetVegaDuration(),
//		}, nil
//	}
//
// func convertToViewUniResponseBySQL(ctx context.Context, query interfaces.ViewQueryInterface,
//
//		view *interfaces.LogicView, content []byte, total int64) (*interfaces.ViewUniResponseV2, error) {
//		rootNode, err := sonic.Get(content)
//		if err != nil {
//			detail := fmt.Sprintf("SQL parse root failed, %s", err.Error())
//			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//				rest.PublicError_InternalServerError).WithErrorDetails(detail)
//		}
//
//		// 解析SQL查询结果
//		currentTotal, err := parseSQLTotalCount(rootNode)
//		if err != nil {
//			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//				rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
//		}
//		logger.Infof("SQL return %d documents", currentTotal)
//
//		// 获取列信息和文档数据
//		columns, err := parseSQLColumns(rootNode)
//		if err != nil {
//			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//				rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
//		}
//
//		docs, err := parseSQLDocuments(rootNode, isSingleDataSource(view))
//		if err != nil {
//			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//				verrors.Uniquery_DataView_InternalError_GetDocumentsFailed).WithErrorDetails(err.Error())
//		}
//
//		// 处理字段映射
//		err = processViewFields(view, columns, query.GetCommonParams().SqlStr)
//		if err != nil {
//			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//				rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
//		}
//
//		results := make([]map[string]interface{}, len(docs))
//		errCh := make(chan error, len(docs))
//		defer close(errCh)
//
//		outputFields := query.GetCommonParams().OutputFields
//		finalFieldsMap := make(map[string]*interfaces.ViewProperty)
//		if len(outputFields) == 0 {
//			finalFieldsMap = view.FieldsMap
//		} else {
//			for _, of := range outputFields {
//				if value, ok := view.FieldsMap[of]; ok {
//					finalFieldsMap[of] = value
//				} else {
//					logger.Errorf("SQL output field '%s' not found in view '%s'", of, view.ViewName)
//				}
//			}
//		}
//
//		taskCtx := &taskContext{
//			hasOutputFields: len(outputFields) > 0,
//			finalFieldsMap:  finalFieldsMap,
//			results:         results,
//			wg:              sync.WaitGroup{},
//			errCh:           errCh,
//		}
//
//		for i, doc := range docs {
//			taskCtx.wg.Add(1)
//
//			err := viewPool.Submit(processDocBySQL(taskCtx, view, columns, i, doc, query.GetCommonParams().Format))
//			if err != nil {
//				detail := fmt.Sprintf("SQL submit task of processing a document failed, %s", err.Error())
//				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//					verrors.Uniquery_DataView_InternalError_SubmitTaskFailed).WithErrorDetails(detail)
//			}
//		}
//		taskCtx.wg.Wait()
//
//		if len(taskCtx.errCh) > 0 {
//			err := <-taskCtx.errCh
//
//			detail := fmt.Sprintf("SQL process a document failed, %s", err.Error())
//			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
//				verrors.Uniquery_DataView_InternalError_ProcessDocFailed).WithErrorDetails(detail)
//		}
//
//		var searchAfter []any
//		if query.GetCommonParams().UseSearchAfter {
//			searchAfter = extractSearchAfterParams(rootNode, isSingleDataSource(view))
//		}
//
//		includeView := query.GetQueryParams()[interfaces.QueryParam_IncludeView].(bool)
//		if !includeView {
//			view = nil
//		}
//
//		// 如果不需要 total 计数，total 设为 nil, 不返回此参数
//		var totalCount *int64
//		if query.GetCommonParams().NeedTotal {
//			totalCount = &total
//		} else {
//			totalCount = nil
//		}
//
//		return &interfaces.ViewUniResponseV2{
//			PitID:          "",
//			SearchAfter:    searchAfter,
//			View:           view,
//			Entries:        results,
//			TotalCount:     totalCount,
//			ScrollId:       "",
//			VegaDurationMs: query.GetVegaDuration(),
//		}, nil
//	}
//
// // ProcessDocByDSL 处理基于 DSL 查询的文档
//
//	func processDocByDSL(taskCtx *taskContext, view *interfaces.LogicView, i int, doc ast.Node, format string) func() {
//		return func() {
//			defer taskCtx.wg.Done()
//
//			originNode := doc.Get("_source")
//			originString, err := originNode.Raw()
//			if err != nil {
//				taskCtx.errCh <- err
//				return
//			}
//
//			var origin map[string]any
//			d := decoder.NewDecoder(originString)
//			d.UseInt64()
//			if err = d.Decode(&origin); err != nil {
//				logger.Errorf("processDocByDSL unmarshal result failed: %s", err.Error())
//				taskCtx.errCh <- err
//				return
//			}
//
//			// pick 为视图最终输出数据
//			pick := make(map[string]any)
//			switch format {
//			case interfaces.Format_Flat:
//				if view.FieldScope == interfaces.FieldScope_All {
//					err = flatten(true, "", origin, pick)
//				} else {
//					// 平铺的时候过滤字段
//					err = flattenWithPickField(true, "", origin, pick, taskCtx.finalFieldsMap)
//				}
//				if err != nil {
//					taskCtx.errCh <- err
//					return
//				}
//			case interfaces.Format_Original:
//				if (view.Type == interfaces.ViewType_Atomic || view.FieldScope == interfaces.FieldScope_All) && !taskCtx.hasOutputFields {
//					pick = origin
//				} else {
//					err = pickData(origin, pick, taskCtx.finalFieldsMap)
//					if err != nil {
//						taskCtx.errCh <- err
//						return
//					}
//				}
//			default:
//				err = fmt.Errorf("unsupported format: %s", format)
//				taskCtx.errCh <- err
//			}
//
//			// 添加 _index和_score 字段
//			if taskCtx.includeIndexField {
//				docIndex, err := doc.Get("_index").String()
//				if err != nil {
//					taskCtx.errCh <- err
//					return
//				}
//				pick["__index"] = docIndex
//			}
//
//			if taskCtx.includeScoreField {
//				docScore, err := doc.Get("_score").Float64()
//				if err != nil {
//					taskCtx.errCh <- err
//					return
//				}
//				pick["_score"] = docScore
//			}
//
//			taskCtx.results[i] = pick
//		}
//	}
//
// // ProcessDocBySQL 处理基于 SQL 的文档
//
//	func processDocBySQL(taskCtx *taskContext, view *interfaces.LogicView, columns []ast.Node, i int, doc ast.Node, format string) func() {
//		return func() {
//			defer taskCtx.wg.Done()
//
//			values, err := doc.ArrayUseNode()
//			if err != nil {
//				logger.Errorf("SQL valuesNode convert to arrayNode failed, %s", err.Error())
//				taskCtx.errCh <- err
//				return
//			}
//
//			origin := make(map[string]any)
//			// 遍历每列，将每列的值变成 key-value 对象
//			for j, col := range columns {
//				fieldName, _ := col.Get("name").String()
//				// fieldType, _ := col.Get("vega_type").String()
//
//				if j < len(values) {
//					node := values[j]
//					val, err := node.InterfaceUseNumber()
//					if err != nil {
//						logger.Errorf("SQL valueNode convert to any type failed, %s", err.Error())
//						taskCtx.errCh <- err
//						return
//					}
//
//					if _, ok := origin[fieldName]; ok {
//						for k, v := range view.FieldsMap {
//							// 对于join，可能会有原始字段重复，二次出现的重复字段拼接为name_srcNodeName
//							if strings.HasPrefix(k, fmt.Sprintf("%s_", fieldName)) && v.OriginalName == fieldName {
//								origin[v.Name] = val
//							}
//						}
//					} else {
//						origin[fieldName] = val
//					}
//				}
//			}
//
//			// pick 为视图最终输出数据
//			pick := make(map[string]any)
//			switch format {
//			case interfaces.Format_Flat:
//				// 平铺的时候过滤字段
//				if view.FieldScope == interfaces.FieldScope_All {
//					err = flatten(true, "", origin, pick)
//				} else {
//					err = flattenWithPickField(true, "", origin, pick, taskCtx.finalFieldsMap)
//				}
//				if err != nil {
//					taskCtx.errCh <- err
//					return
//				}
//			case interfaces.Format_Original:
//				// 当视图类型为原子视图且未指定输出字段时，直接返回原始文档
//				if view.FieldScope == interfaces.FieldScope_All && !taskCtx.hasOutputFields {
//					pick = origin
//				} else {
//					err := pickData(origin, pick, taskCtx.finalFieldsMap)
//					if err != nil {
//						taskCtx.errCh <- err
//						return
//					}
//				}
//			default:
//				err := fmt.Errorf("unsupported format: %s", format)
//
//				taskCtx.errCh <- err
//			}
//
//			taskCtx.results[i] = pick
//		}
//	}
//
// // 判断全局过滤条件的字段在不在视图字段里，在返回true，不在则返回false
//
//	func checkConditionFieldExist(viewFieldsMap map[string]*interfaces.ViewProperty, cfg *filter_condition.FilterCondCfg) (string, bool) {
//		if cfg == nil {
//			return "", true
//		}
//
//		// 判断过滤器是否为空对象 {}
//		if cfg.Name == "" && cfg.Operation == "" && len(cfg.SubConds) == 0 && cfg.ValueFrom == "" && cfg.Value == nil {
//			return "", true
//		}
//
//		switch cfg.Operation {
//		case filter_condition.OperationAnd, filter_condition.OperationOr:
//			for _, subCond := range cfg.SubConds {
//				fieldName, res := checkConditionFieldExist(viewFieldsMap, subCond)
//				if !res {
//					return fieldName, false
//				}
//			}
//		case filter_condition.OperationMultiMatch:
//			// do nothing，校验留给下层做
//		default:
//			// 判断除 * 之外的字段权限
//			if cfg.Name != interfaces.AllField {
//				if _, ok := viewFieldsMap[cfg.Name]; !ok {
//					return cfg.Name, false
//				}
//			}
//		}
//
//		return cfg.Name, true
//	}
//
// // processViewFields 处理视图字段映射逻辑
//
//	func processViewFields(view *interfaces.LogicView, columns []ast.Node, sqlStr string) error {
//		// 预览情况下，sql node的字段列表需要字段列表
//		if view.FieldScope == interfaces.FieldScope_All {
//			return processAllFields(view, columns)
//		} else if view.HasDataScopeSQLNode {
//			return processSQLNodeFields(view, columns)
//		}
//
//		return nil
//	}
//
// // processAllFields 处理所有字段的情况
//
//	func processAllFields(view *interfaces.LogicView, columns []ast.Node) error {
//		fieldsArr := make([]*interfaces.ViewProperty, 0)
//		fieldsMap := make(map[string]*interfaces.ViewProperty)
//		fieldCount := make(map[string]int)
//
//		for _, col := range columns {
//			fieldName, err := col.Get("name").String()
//			if err != nil {
//				return fmt.Errorf("SQL column name convert to string failed, %s", err.Error())
//			}
//
//			fieldType, err := col.Get("vega_type").String()
//			if err != nil {
//				return fmt.Errorf("SQL column type convert to string failed, %s", err.Error())
//			}
//
//			// 获取原始字段名
//			originalName := fieldName
//
//			// 如果columns里有重复的字段名，为后面的字段设置别名
//			if _, exist := fieldsMap[fieldName]; exist {
//				fieldCount[fieldName]++
//				// 为重复字段创建新的唯一名称（添加后缀）
//				newFieldName := fmt.Sprintf("%s_%d", fieldName, fieldCount[fieldName])
//				f := &interfaces.ViewProperty{
//					Name:         newFieldName, // 唯一字段名
//					DisplayName:  newFieldName,
//					OriginalName: originalName, // 保存原始字段名
//					Type:         fieldType,
//				}
//				fieldsMap[newFieldName] = f
//				fieldsArr = append(fieldsArr, f)
//			} else {
//				fieldCount[fieldName] = 0
//				f := &interfaces.ViewProperty{
//					Name:         fieldName,
//					DisplayName:  fieldName,
//					OriginalName: originalName,
//					Type:         fieldType,
//				}
//				fieldsMap[fieldName] = f
//				fieldsArr = append(fieldsArr, f)
//			}
//		}
//
//		view.Fields = fieldsArr
//		view.FieldsMap = fieldsMap
//		return nil
//	}
//
// // processSQLNodeFields 处理SQL节点字段的情况
//
//	func processSQLNodeFields(view *interfaces.LogicView, columns []ast.Node) error {
//		fieldsArr := make([]*interfaces.ViewProperty, 0)
//		fieldsMap := make(map[string]*interfaces.ViewProperty)
//
//		for _, col := range columns {
//			fieldName, err := col.Get("name").String()
//			if err != nil {
//				return fmt.Errorf("SQL column name convert to string failed, %s", err.Error())
//			}
//			fieldType, err := col.Get("vega_type").String()
//			if err != nil {
//				return fmt.Errorf("SQL column type convert to string failed, %s", err.Error())
//			}
//
//			// 非*情况下，已解析出字段，补齐字段类型
//			if f, ok := view.FieldsMap[fieldName]; ok {
//				f.Type = fieldType
//				fieldsMap[fieldName] = f
//			}
//		}
//
//		for _, af := range fieldsMap {
//			fieldsArr = append(fieldsArr, af)
//		}
//
//		view.Fields = fieldsArr
//		view.FieldsMap = fieldsMap
//		return nil
//	}
//
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

		case interfaces.LogicDefinitionNodeType_Join:
			err := validateJoinNode(ctx, node, nodeMap)
			if err != nil {
				return err
			}
		case interfaces.LogicDefinitionNodeType_Union:
			err := validateUnionNode(ctx, "sql", node, nodeMap)
			if err != nil {
				return err
			}
		case interfaces.LogicDefinitionNodeType_Sql:
			_, err := validateSqlNode(ctx, node, nodeMap)
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

func validateJoinNode(ctx context.Context, node *interfaces.LogicDefinitionNode, nodeMap map[string]struct{}) error {
	// 仅支持两个视图join
	if len(node.Inputs) != 2 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition join config is invalid, only support two views join")
	}

	// 校验输入节点是否重复
	inputNodesMap := make(map[string]struct{})
	for _, inputNode := range node.Inputs {
		if _, ok := inputNodesMap[inputNode]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The logic definition join config is invalid, input_nodes must be unique")
		}
		inputNodesMap[inputNode] = struct{}{}
	}

	// 校验输入节点是否存在
	for _, inputNode := range node.Inputs {
		if _, ok := nodeMap[inputNode]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails(fmt.Sprintf("The logic definition join config is invalid, input_node '%s' is not exist", inputNode))
		}
	}

	// mapstructure 解析 join_on
	var cfg interfaces.JoinNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition join config is invalid")
	}

	// join_type 只能为 inner, left, right, full outer
	if _, ok := interfaces.JoinTypeMap[cfg.JoinType]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition join config is invalid, join_type must be inner, left, right, full outer")
	}

	// join_on 校验
	if len(cfg.JoinOn) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition join config is invalid, join_on must be set")
	}

	// join_on 校验
	for _, joinOn := range cfg.JoinOn {
		if joinOn.LeftField == "" || joinOn.RightField == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The logic definition join config is invalid, join_on left_field and right_field must be set")
		}

		// 操作符必须只为=
		if joinOn.Operator != "=" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The logic definition join config is invalid, join_on operator must be =")
		}
	}

	return nil
}

func validateUnionNode(ctx context.Context, qType string, node *interfaces.LogicDefinitionNode, nodeMap map[string]struct{}) error {
	// 当前仅支持两个视图union
	if len(node.Inputs) < 2 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition union config is invalid, need at least two views union")
	}

	// 校验输入节点是否重复
	inputNodesMap := make(map[string]struct{})
	for _, inputNode := range node.Inputs {
		if _, ok := inputNodesMap[inputNode]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The logic definition union config is invalid, input_nodes must be unique")
		}
		inputNodesMap[inputNode] = struct{}{}
	}

	// 校验输入节点是否存在
	for _, inputNode := range node.Inputs {
		if _, ok := nodeMap[inputNode]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails(fmt.Sprintf("The logic definition union config is invalid, input_node '%s' is not exist", inputNode))
		}
	}

	// mapstructure 解析 union config
	var cfg interfaces.UnionNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition union config is invalid")
	}

	if _, ok := interfaces.UnionTypeMap[cfg.UnionType]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition union config is invalid, union_type must be all, distinct")
	}

	return nil
}

func validateSqlNode(ctx context.Context, node *interfaces.LogicDefinitionNode, nodeMap map[string]struct{}) (bool, error) {
	// 输入节点不能为空
	if len(node.Inputs) == 0 {
		return false, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition sql config is invalid, input_nodes must be set")
	}

	// 校验输入节点是否重复
	inputNodesMap := make(map[string]struct{})
	for _, inputNode := range node.Inputs {
		if _, ok := inputNodesMap[inputNode]; ok {
			return false, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The logic definition sql config is invalid, input_nodes must be unique")
		}
		inputNodesMap[inputNode] = struct{}{}
	}

	// 校验输入节点是否存在
	for _, inputNode := range node.Inputs {
		if _, ok := nodeMap[inputNode]; !ok {
			return false, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails(fmt.Sprintf("The logic definition sql config is invalid, input_node '%s' is not exist", inputNode))
		}
	}

	// mapstructure 解析 sql config
	var cfg interfaces.SQLNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return false, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition sql config is invalid")
	}

	// 校验 sql_str 是否为空
	if cfg.SQL == "" {
		return false, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition sql config is invalid, sql_expression must be set")
	}

	// 解析sql里的字段，补全sql节点的输出字段列表
	processedSQL := replaceTablePlaceholders(cfg.SQL, "`")
	logger.Infof("processedSQL for parse fields from sql_expression: %s", processedSQL)
	parser := NewSQLFieldParser()
	info := parser.Parse(processedSQL)

	// 输出 JSON 格式
	logger.Infof("\nJSON 格式:\n%s\n\n", info.FormatAsJSON())

	// 组装sql的输出字段
	outputFields := make([]*interfaces.ViewProperty, 0)
	for _, field := range info.Fields {
		name := field.Name
		if field.Alias != "" {
			name = field.Alias
		}

		outputFields = append(outputFields, &interfaces.ViewProperty{
			Property: interfaces.Property{
				Name: name,
			},
		})
	}

	node.OutputFields = outputFields

	return info.HasStar, nil
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

	// 如果没传fields字段列表，默认使用output节点的输出字段
	// if len(node.OutputFields) == 0 {
	// 	// return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
	// 	// 	WithErrorDetails("The output node must have output fields")
	// }

	// 校验name不能重复，display_name 不能重复, original_name 可以重复
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
