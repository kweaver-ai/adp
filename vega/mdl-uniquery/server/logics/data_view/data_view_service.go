package data_view

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/bytedance/sonic/decoder"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"
	"github.com/panjf2000/ants/v2"
	"github.com/tidwall/gjson"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"

	"uniquery/common"
	cond "uniquery/common/condition"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	dtype "uniquery/interfaces/data_type"
	"uniquery/logics"
	"uniquery/logics/permission"
)

var (
	dvServiceOnce sync.Once
	dvService     interfaces.DataViewService
	viewPool      *ants.Pool
)

type multiFieldStatsDSLParas struct {
	baseFilterStr   string
	timeFilterStr   string
	aggTermStr      string
	scriptFilterStr string
}

type Bucket struct {
	Key         any    `json:"key"`
	KeyAsString string `json:"key_as_string"`
	DocCount    int64  `json:"doc_count"`
}

type dataViewService struct {
	appSetting  *common.AppSetting
	dvAccess    interfaces.DataViewAccess
	dvrcrAccess interfaces.DataViewRowColumnRuleAccess
	ibAccess    interfaces.IndexBaseAccess
	osAccess    interfaces.OpenSearchAccess
	vdsAccess   interfaces.VegaDataSourceAccess
	vgAccess    interfaces.VegaGatewayAccess
	ps          interfaces.PermissionService
}

func NewDataViewService(appSetting *common.AppSetting) interfaces.DataViewService {
	dvServiceOnce.Do(func() {
		dvService = &dataViewService{
			appSetting:  appSetting,
			dvAccess:    logics.DVAccess,
			dvrcrAccess: logics.DVRCRAccess,
			ibAccess:    logics.IBAccess,
			osAccess:    logics.OSAccess,
			vdsAccess:   logics.VDSAccess,
			vgAccess:    logics.VGAccess,
			ps:          permission.NewPermissionService(appSetting),
		}

		InitViewPool(appSetting.PoolSetting)
	})
	return dvService
}

// 初始化协程池
func InitViewPool(poolSetting common.PoolSetting) {
	pool, err := ants.NewPool(poolSetting.ViewPoolSize, ants.WithPreAlloc(true), ants.WithNonblocking(false))
	if err != nil {
		logger.Fatalf("Init parser pool failed, %s", err.Error())
		panic(err)
	}

	viewPool = pool
}

// 视图数据预览
func (dvs *dataViewService) Simulate(ctx context.Context, query *interfaces.DataViewSimulateQuery) (*interfaces.ViewUniResponseV2, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Simulate view data")
	defer span.End()

	// 决策权限, 预览的时候还没有视图id，此时的预览校验用新建或者编辑
	ops, err := dvs.ps.GetResourcesOperations(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW,
		[]string{interfaces.RESOURCE_ID_ALL})
	if err != nil {
		return nil, err
	}

	if len(ops) != 1 {
		// 无权限
		return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: insufficient permissions for data view's create or modify operation")
	}
	// 从 ops 里找新建或编辑的权限
	for _, op := range ops[0].Operations {
		if op != interfaces.OPERATION_TYPE_CREATE && op != interfaces.OPERATION_TYPE_MODIFY {
			return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for data view's create or modify operation")
		}
	}

	view := &interfaces.DataView{
		Type:           query.Type,
		QueryType:      query.QueryType,
		TechnicalName:  query.TechnicalName,
		DataSourceType: query.DataSourceType,
		DataSourceID:   query.DataSourceID,
		FileName:       query.FileName,
		ExcelConfig:    query.ExcelConfig,
		DataScope:      query.DataScope,
		Fields:         query.Fields,
	}

	// query.NeedTotal = true
	// 设置预览的默认format为flat
	if query.Format == "" {
		query.Format = interfaces.Format_Flat
	}

	switch query.QueryType {
	case interfaces.QueryType_IndexBase:
		return dvs.SimulateByIndexBase(ctx, query, view)
	case interfaces.QueryType_DSL:
		return dvs.SimulateByDSL(ctx, query, view)
	case interfaces.QueryType_SQL:
		return dvs.SimulateBySQL(ctx, query, view)
	default:
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails("query type must be DSL or SQL or IndexBase")
	}
}

// 索引库类视图数据预览
func (dvs *dataViewService) SimulateByIndexBase(ctx context.Context, query *interfaces.DataViewSimulateQuery,
	view *interfaces.DataView) (*interfaces.ViewUniResponseV2, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Simulate view data by DSL")
	defer span.End()

	var baseTypes []string
	var baseTypeViewMap map[string]string // 视图 id 和索引库的映射
	switch view.Type {
	case interfaces.ViewType_Atomic:
		// 原子视图，实时获取索引库的字段填充视图 fields
		span.SetAttributes(attr.Key("base_type").String(query.TechnicalName))

		// 获取索引库信息
		baseInfos, err := dvs.ibAccess.GetIndexBasesByTypes(ctx, []string{query.TechnicalName})
		if err != nil {
			span.SetStatus(codes.Error, "Get index bases by types failed")
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_InternalError_GetIndexBaseByTypeFailed).WithErrorDetails(err.Error())
		}

		viewFields, viewFieldsMap := convertIndexBaseFieldsToViewFields(baseInfos)
		view.Fields = viewFields
		view.FieldsMap = viewFieldsMap

		baseTypes = []string{query.TechnicalName}
		baseTypeViewMap = map[string]string{query.TechnicalName: view.ViewID}
	case interfaces.ViewType_Custom:
		// 校验 data_scope
		var httpErr error
		baseTypes, baseTypeViewMap, httpErr = validateDataScope(ctx, dvs, view)
		if httpErr != nil {
			span.SetStatus(codes.Error, "Validate data scope failed")
			return nil, httpErr
		}

		// 组装fieldsMap
		fieldsMap := make(map[string]*cond.ViewField)
		for _, field := range view.Fields {
			if existField, ok := fieldsMap[field.Name]; ok {
				if field.Type != existField.Type {
					errDetails := fmt.Sprintf("Field '%s' has two different types: '%s' and '%s'", field.Name, field.Type, existField.Type)
					logger.Errorf(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
						uerrors.Uniquery_DataView_FieldTypeConflict).WithErrorDetails(errDetails)
				}
			} else {
				fieldsMap[field.Name] = field
			}
		}

		view.FieldsMap = fieldsMap

		// 校验是否包含默认字段
		for _, key := range interfaces.REQUIRED_META_FIELDS {
			if _, ok := view.FieldsMap[key]; !ok {
				return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_MissingRequiredField).
					WithDescription(map[string]any{"FieldName": key}).
					WithErrorDetails(fmt.Sprintf("The default required meta field '%s' is missing", key))
			}
		}

		// 获取索引库信息
		baseInfos, err := dvs.ibAccess.GetIndexBasesByTypes(ctx, baseTypes)
		if err != nil {
			span.SetStatus(codes.Error, "Get index bases by types failed")
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_InternalError_GetIndexBaseByTypeFailed).WithErrorDetails(err.Error())
		}

		// 校验字段在不同索引库是否存在类型冲突, 在索引库字段里遍历类型，如果是不同类别的类型，直接报错
		for _, base := range baseInfos {
			baseFields := mergeIndexBaseFields(base.Mappings)

			for _, baseField := range baseFields {
				// 校验视图字段类型是否和索引库里的字段类型一样
				if viewField, ok := view.FieldsMap[baseField.Field]; ok {
					baseFieldType, ok := dtype.IndexBase_DataType_Map[baseField.Type]
					if !ok {
						baseFieldType = baseField.Type
					}

					if viewField.Type != baseFieldType {
						errDetails := fmt.Sprintf("Field '%s' with type '%s' is different from type '%s' in index base",
							baseField.Field, viewField.Type, baseFieldType)
						o11y.Error(ctx, errDetails)
						span.SetStatus(codes.Error, "Field type conflict")
						return nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
							uerrors.Uniquery_DataView_FieldTypeConflict).WithErrorDetails(errDetails)
					}
				}
			}
		}
	default:
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).WithErrorDetails("invalid view type")
	}

	// 使用索引优化接口获取索引
	_, indices, _, err := dvs.GetIndices(ctx, baseTypes, query.Start, query.End)
	if err != nil {
		span.SetStatus(codes.Error, "Get indices failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_GetIndicesFailed).WithErrorDetails(err.Error())
	}

	// 如果索引列表为空，则返回空数据
	if len(indices) == 0 {
		total := int64(0)
		span.SetStatus(codes.Ok, "No indices found")
		return &interfaces.ViewUniResponseV2{
			PitID:       "",
			SearchAfter: nil,
			View:        view,
			Entries:     []map[string]any{},
			TotalCount:  &total,
		}, nil
	}

	// 视图 ID 到索引列表的映射
	viewIndicesMap, err := getViewIndicesMap(indices, baseTypeViewMap)
	if err != nil {
		span.SetStatus(codes.Error, "Get view indices map failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	// 转成 DSL
	dsl, err := buildDSL(ctx, query, view, viewIndicesMap)
	if err != nil {
		o11y.Error(ctx, err.Error())
		span.SetStatus(codes.Error, "Convert to DSL failed")
		return nil, err
	}

	dslQueryBuffer, err := marshalDSL(dsl)
	if err != nil {
		span.SetStatus(codes.Error, "Marshal DSL failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}
	logger.Infof("view '%s' DSL string is %s", view.ViewName, dslQueryBuffer.String())

	// 调用osAccess查询
	resBytes, _, err := dvs.osAccess.SearchSubmitWithBuffer(ctx, dslQueryBuffer,
		indices, 0, interfaces.DEFAULT_PREFERENCE)
	if err != nil {
		span.SetStatus(codes.Error, "Search submit failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_InternalError_SearchSubmitFailed).WithErrorDetails(err.Error())
	}

	// 预览接口返回total
	total, httpErr := dvs.GetTotalByIndexBase(ctx, dslQueryBuffer, indices)
	if httpErr != nil {
		return nil, httpErr
	}

	// 转成视图统一结构
	res, err := convertToViewUniResponse(ctx, query, view, resBytes, total)
	if err != nil {
		o11y.Error(ctx, err.Error())
		span.SetStatus(codes.Error, "Convert to view uniResponse failed")
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return res, nil
}

// SQL类视图数据预览
func (dvs *dataViewService) SimulateBySQL(ctx context.Context, query *interfaces.DataViewSimulateQuery,
	view *interfaces.DataView) (*interfaces.ViewUniResponseV2, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Simulate view data by SQL")
	defer span.End()

	switch view.Type {
	case interfaces.ViewType_Atomic:
		if query.DataSourceID == "" {
			span.SetStatus(codes.Error, "Data source ID is empty")
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
				rest.PublicError_BadRequest).WithErrorDetails("Data source ID is empty")
		}

		// 获取数据源信息
		dataSource, err := dvs.vdsAccess.GetDataSourceByID(ctx, query.DataSourceID)
		if err != nil {
			span.SetStatus(codes.Error, "Get data source by ID failed")
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails(fmt.Sprintf("Get data source by ID failed, %v", err))
		}

		// 构造视图的meta_table_name
		catalogName := dataSource.BinData.CatalogName
		schemaName := dataSource.BinData.Schema
		// 先用schema，没有再用database
		if schemaName == "" {
			schemaName = dataSource.BinData.DataBaseName
		}
		// database也没有使用默认值 default
		if schemaName == "" {
			schemaName = interfaces.DefaultSchema
		}

		view.MetaTableName = fmt.Sprintf("%s.%s.%s", catalogName, common.QuotationMark(schemaName),
			common.QuotationMark(view.TechnicalName))

	case interfaces.ViewType_Custom:
		_, _, httpErr := validateDataScope(ctx, dvs, view)
		if httpErr != nil {
			span.SetStatus(codes.Error, "Validate data scope failed")
			return nil, httpErr
		}
	}

	// 将字段转为map
	viewFieldsMap := make(map[string]*cond.ViewField)
	for _, field := range view.Fields {
		// init field path
		field.InitFieldPath()
		viewFieldsMap[field.Name] = field
	}
	view.FieldsMap = viewFieldsMap

	resBytes, total, err := dvs.queryBySQL(ctx, query, view)
	if err != nil {
		span.SetStatus(codes.Error, "Query by SQL failed")
		return nil, err
	}

	// 转成视图统一结构
	res, httpErr := convertToViewUniResponse(ctx, query, view, resBytes, total)
	if httpErr != nil {
		o11y.Error(ctx, httpErr.Error())
		span.SetStatus(codes.Error, "Convert to view uniResponse failed")
		return nil, httpErr
	}

	span.SetStatus(codes.Ok, "")
	return res, nil
}

// DSL类视图数据预览
func (dvs *dataViewService) SimulateByDSL(ctx context.Context, query *interfaces.DataViewSimulateQuery,
	view *interfaces.DataView) (*interfaces.ViewUniResponseV2, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Simulate view data by SQL")
	defer span.End()

	switch view.Type {
	case interfaces.ViewType_Atomic:
		if query.DataSourceID == "" {
			span.SetStatus(codes.Error, "Data source ID is empty")
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
				rest.PublicError_BadRequest).WithErrorDetails("Data source ID is empty")
		}

		// 获取数据源信息
		dataSource, err := dvs.vdsAccess.GetDataSourceByID(ctx, query.DataSourceID)
		if err != nil {
			span.SetStatus(codes.Error, "Get data source by ID failed")
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails(fmt.Sprintf("Get data source by ID failed, %v", err))
		}

		// 补充视图的 catalog
		view.DataSourceCatalog = dataSource.BinData.CatalogName

	case interfaces.ViewType_Custom:
		_, _, httpErr := validateDataScope(ctx, dvs, view)
		if httpErr != nil {
			span.SetStatus(codes.Error, "Validate data scope failed")
			return nil, httpErr
		}
	}

	// 将字段转为map
	viewFieldsMap := make(map[string]*cond.ViewField)
	for _, field := range view.Fields {
		// init field path
		field.InitFieldPath()
		viewFieldsMap[field.Name] = field
	}
	view.FieldsMap = viewFieldsMap

	resBytes, total, err := dvs.queryByDSL(ctx, query, view)
	if err != nil {
		span.SetStatus(codes.Error, "Query by SQL failed")
		return nil, err
	}

	// 转成视图统一结构
	res, httpErr := convertToViewUniResponse(ctx, query, view, resBytes, total)
	if httpErr != nil {
		o11y.Error(ctx, httpErr.Error())
		span.SetStatus(codes.Error, "Convert to view uniResponse failed")
		return nil, httpErr
	}

	span.SetStatus(codes.Ok, "")
	return res, nil
}

// 获取单个视图数据
func (dvs *dataViewService) GetSingleViewData(ctx context.Context, viewID string, query interfaces.ViewQueryInterface) (*interfaces.ViewUniResponseV2, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get single view data")
	defer span.End()

	span.SetAttributes(attr.Key("view_id").String(viewID))

	// 决策当前视图id的数据查询权限
	hasPermission, err := dvs.ps.CheckPermissionWithResult(ctx, interfaces.Resource{
		ID:   viewID,
		Type: interfaces.RESOURCE_TYPE_DATA_VIEW,
	}, []string{interfaces.OPERATION_TYPE_DATA_QUERY})

	if err != nil {
		return nil, err
	}

	// 如果有data_query权限，返回视图的全量数据
	// 如果没有data_query权限，则获取视图下的所有行列规则，
	// 决策当前用户具有rule_apply权限的规则，执行规则过滤查询
	if !hasPermission {
		// 获取视图下的所有行列规则
		rowColumnRules, err := dvs.dvrcrAccess.GetRulesByViewID(ctx, viewID)
		if err != nil {
			span.SetStatus(codes.Error, "Get row column rules by view ID failed")
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails(err.Error())
		}

		// 过滤视图下的行列规则，返回当前用户具有rule_apply权限的规则
		filteredRules, httpErr := dvs.FilterRowColumnRules(ctx, rowColumnRules)
		if httpErr != nil {
			span.SetStatus(codes.Error, "Filter row column rules failed")
			return nil, httpErr
		}

		if len(filteredRules) == 0 {
			errDetails := fmt.Sprintf("Neither data query permission nor row column rules with rule_apply permission for view ID %s", viewID)
			span.SetStatus(codes.Error, errDetails)
			return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails(errDetails)
		}

		// 设置查询参数中的行列规则
		query.SetRowColumnRules(filteredRules)
	}

	// data-model服务会检查基础权限(data_view_id,'view_detail')
	view, httpErr := dvs.GetDataViewByID(ctx, viewID, true)
	if httpErr != nil {
		span.SetStatus(codes.Error, "Get data view by ID failed")
		return nil, httpErr
	}

	// 查询数据
	resBytes, total, httpErr := dvs.querySingleViewData(ctx, query, view)
	if httpErr != nil {
		span.SetStatus(codes.Error, "Query single view data failed")
		return nil, httpErr
	}

	// 转成视图统一结构
	res, httpErr := convertToViewUniResponse(ctx, query, view, resBytes, total)
	if httpErr != nil {
		o11y.Error(ctx, httpErr.Error())
		span.SetStatus(codes.Error, "Convert to view uniResponse failed")
		return nil, httpErr
	}

	span.SetStatus(codes.Ok, "")
	return res, nil
}

// 获取单个视图数据, 服务内部 trace model 调用
func (dvs *dataViewService) RetrieveSingleViewData(ctx context.Context, viewID string, query *interfaces.DataViewQueryV1) (*interfaces.ViewInternalResponse, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get single view data for internal module")
	defer span.End()

	// 决策当前视图id的数据查询权限
	err := dvs.ps.CheckPermission(ctx, interfaces.Resource{
		ID:   viewID,
		Type: interfaces.RESOURCE_TYPE_DATA_VIEW,
	}, []string{interfaces.OPERATION_TYPE_DATA_QUERY})
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attr.Key("view_id").String(viewID))

	view, httpErr := dvs.GetDataViewByID(ctx, viewID, true)
	if httpErr != nil {
		span.SetStatus(codes.Error, "Get data view by ID failed")
		return nil, httpErr
	}

	// start1 := time.Now()
	// fmt.Printf("[logic]开始调用querySingleViewData方法查询数据, 当前时间%v\n", start1)
	resBytes, total, httpErr := dvs.querySingleViewData(ctx, query, view)
	// end1 := time.Now()
	// fmt.Printf("[logic]结束调用querySingleViewData方法查询数据, 当前时间%v, 共耗时%v\n", end1, end1.Sub(start1))
	if httpErr != nil {
		span.SetStatus(codes.Error, "Query single view data failed")
		return nil, httpErr
	}

	// 转成内部模块调用的结构
	// start2 := time.Now()
	// fmt.Printf("[logic]开始转换成视图统一结构, 当前时间%v\n", start2)
	res, httpErr := dvs.GetViewResInternal(ctx, query, view, resBytes, total)
	// end2 := time.Now()
	// fmt.Printf("[logic]结束转换成视图统一结构, 当前时间%v, 共耗时%v\n", end2, end2.Sub(start2))
	if httpErr != nil {
		span.SetStatus(codes.Error, "Get internal view response failed")
		return nil, httpErr
	}

	span.SetStatus(codes.Ok, "")
	return res, nil
}

// 获取单个视图对象信息
func (dvs *dataViewService) GetDataViewByID(ctx context.Context, viewID string, includeDataScopeView bool) (*interfaces.DataView, error) {
	views, err := dvs.dvAccess.GetDataViewsByIDs(ctx, viewID, includeDataScopeView)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_GetDataViewByIDFailed).WithErrorDetails(err.Error())
	}

	if len(views) == 0 {
		logger.Errorf("Data view %s not found", viewID)
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound,
			uerrors.Uniquery_DataView_DataViewNotFound).WithErrorDetails(fmt.Sprintf("view %s not found", viewID))
	}

	view := views[0]

	// 补充自定义视图的来源原子视图是否来自同一个数据源
	if includeDataScopeView && view.Type == interfaces.ViewType_Custom {
		dataSourceIDMap := make(map[string]struct{})
		for _, node := range view.DataScope {
			if node.Type == interfaces.DataScopeNodeType_View {
				var viewNodeConfig interfaces.ViewNodeCfg
				err := mapstructure.Decode(node.Config, &viewNodeConfig)
				if err != nil {
					logger.Errorf("Decode view node config failed, err: %v", err)
					return nil, err
				}

				if viewNodeConfig.View == nil {
					logger.Errorf("View node config view is nil")
					return nil, fmt.Errorf("view node config view is nil")
				}

				dataSourceIDMap[viewNodeConfig.View.DataSourceID] = struct{}{}
				view.DataScopeAdvancedParams.DataScopeDataSourceID = viewNodeConfig.View.DataSourceID
			}
		}

		view.IsSingleSource = len(dataSourceIDMap) == 1
	}

	return view, nil
}

// 获取单个视图对象信息
func (dvs *dataViewService) GetDataViewsByIDs(ctx context.Context, viewIDs []string, includeDataScopeView bool) (map[string]*interfaces.DataView, error) {
	if len(viewIDs) == 0 {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails("view id list is empty")
	}

	ids := strings.Join(viewIDs, ",")

	views, err := dvs.dvAccess.GetDataViewsByIDs(ctx, ids, includeDataScopeView)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_GetDataViewByIDFailed).WithErrorDetails(err.Error())
	}

	if len(views) != len(viewIDs) {
		logger.Errorf("Data view %s not found", ids)
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound,
			uerrors.Uniquery_DataView_DataViewNotFound).WithErrorDetails(fmt.Sprintf("view %s not found", ids))
	}

	// 转成视图id和视图对象的map返回
	viewMap := make(map[string]*interfaces.DataView)
	for _, view := range views {
		viewMap[view.ViewID] = view
	}

	return viewMap, nil
}

// 单个查询视图数据和批量查询通用的函数
func (dvs *dataViewService) querySingleViewData(ctx context.Context, query interfaces.ViewQueryInterface,
	view *interfaces.DataView) (resBytes []byte, total int64, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Query single view data")
	defer span.End()

	globalFilters := query.GetGlobalFilters()
	commonParams := query.GetCommonParams()
	format := commonParams.Format

	allowNonExistField := query.GetQueryParams()[interfaces.QueryParam_AllowNonExistField].(bool)

	viewFieldsMap := make(map[string]*cond.ViewField)
	for _, field := range view.Fields {
		newField := field
		// init field path
		newField.InitFieldPath()
		viewFieldsMap[newField.Name] = newField
	}
	// 将视图字段转成map
	view.FieldsMap = viewFieldsMap

	// 非严格模式下，如果全局过滤条件的字段不在视图字段列表里，数据返回空
	fieldName, exist := checkConditionFieldExist(viewFieldsMap, globalFilters)
	if !exist {
		if allowNonExistField {
			return []byte{}, total, nil
		} else {
			return []byte{}, total, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidFilterField_FieldNotInView).
				WithErrorDetails(fmt.Sprintf("condition config field name '%s' must in view original fields", fieldName))
		}
	}

	// 如果查询参数没有设置format，则默认设置成flat
	if format == "" {
		query.SetFormat(interfaces.Format_Flat)
	}

	switch view.QueryType {
	case interfaces.QueryType_IndexBase:
		return dvs.queryByIndexBase(ctx, query, view)
	case interfaces.QueryType_DSL:
		return dvs.queryByDSL(ctx, query, view)
	case interfaces.QueryType_SQL:
		return dvs.queryBySQL(ctx, query, view)
	default:
		return nil, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_QueryType).
			WithErrorDetails("query type must be DSL or SQL or IndexBase")
	}
}

func (dvs *dataViewService) queryByIndexBase(ctx context.Context, query interfaces.ViewQueryInterface,
	view *interfaces.DataView) (resBytes []byte, total int64, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Query single view data")
	defer span.End()

	commonParams := query.GetCommonParams()
	start := commonParams.Start
	end := commonParams.End
	scrollParams := query.GetScroll()
	searchAfterParams := query.GetSearchAfterParams()

	// 如果有 scroll_id，则直接调用 scroll 查询，不需要拼接 dsl
	if scrollParams.ScrollId != "" {
		scrollQuery := interfaces.Scroll{
			Scroll:   scrollParams.Scroll,
			ScrollId: scrollParams.ScrollId,
		}

		resBytes, _, err = dvs.osAccess.Scroll(ctx, scrollQuery)
		if err != nil {
			span.SetStatus(codes.Error, "Scroll query failed")
			return []byte{}, total, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_InternalError_ScrollFailed).WithErrorDetails(err.Error())
		}

		span.SetStatus(codes.Ok, "")
		return resBytes, total, nil
	}

	// 获取视图的索引库列表
	baseTypes, baseTypeViewMap, err := GetBaseTypes(view)
	if err != nil {
		span.SetStatus(codes.Error, "Get base types failed")
		return []byte{}, total, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	// 使用索引优化接口获取索引
	_, indices, _, err := dvs.GetIndices(ctx, baseTypes, start, end)
	if err != nil {
		span.SetStatus(codes.Error, "Get indices failed")
		return []byte{}, total, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_GetIndicesFailed).WithErrorDetails(err.Error())
	}

	// 如果索引列表为空，则返回空数据, 不需要下面拼接dsl
	if len(indices) == 0 {
		span.SetStatus(codes.Ok, "No indices found")
		return []byte{}, total, nil
	}

	// 如果 pit_id 为空，且 pit_keep_alive 参数不为空，则创建一个 pit
	// 创建 pit 参数 keep_alive 为 required, 这处会影响拼接dsl
	if searchAfterParams != nil && searchAfterParams.PitID == "" && searchAfterParams.PitKeepAlive != "" {
		keepAlive, err := convert.IntToDuration(searchAfterParams.PitKeepAlive)
		if err != nil {
			span.SetStatus(codes.Error, "[Query.SearchAfter.PitKeepAlive] convert string type to time duration failed")
			return []byte{}, total, rest.NewHTTPError(ctx, http.StatusBadRequest,
				uerrors.Uniquery_DataView_InvalidParameter_PitKeepAlive).WithErrorDetails(err.Error())
		}

		createPitRes, pitID, _, err := dvs.osAccess.CreatePointInTime(ctx, indices, keepAlive)
		logger.Debugf("[logic] create pit response: %s", string(createPitRes))
		if err != nil {
			span.SetStatus(codes.Error, "Create point in time failed")
			return []byte{}, total, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_InternalError_CreatePointInTimeFailed).WithErrorDetails(err.Error())
		}

		searchAfterParams.PitID = pitID
	}

	// 如果没传 limit，传了 search_after 参数，设置默认limit 10000,  这处会影响拼接dsl
	if commonParams.Limit == 0 && searchAfterParams != nil && len(searchAfterParams.SearchAfter) > 0 {
		query.SetLimit(interfaces.SearchAfter_Limit)
	}

	// 视图 ID 到索引列表的映射
	viewIndicesMap, err := getViewIndicesMap(indices, baseTypeViewMap)
	if err != nil {
		span.SetStatus(codes.Error, "Get view indices map failed")
		return []byte{}, total, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	// 转成 DSL
	dsl, httpErr := buildDSL(ctx, query, view, viewIndicesMap)
	if httpErr != nil {
		o11y.Error(ctx, httpErr.Error())
		span.SetStatus(codes.Error, "Convert to DSL failed")
		return []byte{}, total, httpErr
	}

	dslQueryBuffer, err := marshalDSL(dsl)
	if err != nil {
		span.SetStatus(codes.Error, "Marshal DSL failed")
		return []byte{}, total, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}
	logger.Infof("view '%s' DSL string is %s", view.ViewName, dslQueryBuffer.String())

	// 如果有 scroll 参数, 则转成 time.duration 的形式
	var scroll time.Duration
	if scrollParams.Scroll != "" {
		scroll, err = convert.IntToDuration(scrollParams.Scroll)
		if err != nil {
			span.SetStatus(codes.Error, "[Query.Scroll] convert string type to time duration failed")
			return []byte{}, total, rest.NewHTTPError(ctx, http.StatusBadRequest,
				uerrors.Uniquery_DataView_InvalidParameter_Scroll).WithErrorDetails(err.Error())
		}
	}

	// // from+limit 查询
	// // scroll 翻页第一次查询
	// // 带 pit 的 search_after 翻页第一次查询需要返回 total  ==> 接口通过 need_total 参数显示传递
	// query.SetNeedTotal(true)

	// 使用 pit 查询
	if searchAfterParams != nil && searchAfterParams.PitID != "" {
		// 如果 search_after 翻页传递了 pit_id 和 search_after 参数，则不需要返回 total
		if len(searchAfterParams.SearchAfter) > 0 {
			query.SetNeedTotal(false)
		}

		var osStatusCode int
		resBytes, osStatusCode, err = dvs.osAccess.SearchWithPit(ctx, dslQueryBuffer)
		if osStatusCode == http.StatusNotFound && strings.Contains(err.Error(), interfaces.SearchError_SearchContextMissingException) {
			return []byte{}, total, rest.NewHTTPError(ctx, http.StatusNotFound,
				uerrors.Uniquery_DataView_PointInTimeSearchContextNotFound).WithErrorDetails(err.Error())
		}

	} else {
		resBytes, _, err = dvs.osAccess.SearchSubmitWithBuffer(ctx, dslQueryBuffer, indices,
			scroll, interfaces.DEFAULT_PREFERENCE)
	}
	if err != nil {
		span.SetStatus(codes.Error, "Search submit failed")
		return []byte{}, total, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_InternalError_SearchSubmitFailed).WithErrorDetails(err.Error())
	}

	// 根据NeedTotal参数决定是否查询total
	if query.GetCommonParams().NeedTotal {
		total, httpErr = dvs.GetTotalByIndexBase(ctx, dslQueryBuffer, indices)
		if httpErr != nil {
			return nil, total, httpErr
		}
	}

	span.SetStatus(codes.Ok, "")
	return resBytes, total, nil
}

func (dvs *dataViewService) queryByDSL(ctx context.Context, query interfaces.ViewQueryInterface,
	view *interfaces.DataView) (resBytes []byte, total int64, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Query single view data")
	defer span.End()

	commonParams := query.GetCommonParams()
	searchAfterParams := query.GetSearchAfterParams()

	// 如果没传 limit，传了 search_after 参数，设置默认limit 10000,  这处会影响拼接dsl
	if commonParams.Limit == 0 && searchAfterParams != nil && len(searchAfterParams.SearchAfter) > 0 {
		query.SetLimit(interfaces.SearchAfter_Limit)
	}

	// 获取索引列表, 视图 ID 到索引列表的映射
	catalogName, indices, viewIndicesMap, err := dvs.getIndicesByView(view)
	if err != nil {
		span.SetStatus(codes.Error, "Get indices failed")
		return []byte{}, total, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_GetIndicesFailed).WithErrorDetails(err.Error())
	}

	// 如果索引列表为空，则返回空数据, 不需要下面拼接dsl
	if len(indices) == 0 {
		span.SetStatus(codes.Ok, "No indices found")
		return []byte{}, total, nil
	}

	// 转成 DSL
	dsl, httpErr := buildDSL(ctx, query, view, viewIndicesMap)
	if httpErr != nil {
		o11y.Error(ctx, httpErr.Error())
		span.SetStatus(codes.Error, "Convert to DSL failed")
		return []byte{}, total, httpErr
	}

	// 记录查询vega耗时
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime).Milliseconds()
		logger.Infof("query vega data cost time is %dms", elapsed)
		query.SetVegaDuration(elapsed)
	}()

	// 向vega执行dsl查询
	fetchParams := &interfaces.FetchVegaDataParams{
		IsSingleDataSource: isSingleDataSource(view),
		QueryType:          interfaces.QueryType_DSL,
		DataSourceID:       getQueryDataSourceID(view),
		CatalogName:        catalogName,
		TableNames:         indices,
		Dsl:                dsl,
	}
	dataBatch, err := dvs.vgAccess.FetchDataNoUnmarshal(ctx, fetchParams)
	if err != nil {
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(err.Error())
	}

	// 根据NeedTotal参数决定是否查询total, vega没提供total接口，DSL查询暂不支持total
	// if query.GetCommonParams().NeedTotal {
	// 	total, httpErr = dvs.GetTotalByDSL(ctx, dsl, indices)
	// 	if httpErr != nil {
	// 		return nil, total, httpErr
	// 	}
	// }

	span.SetStatus(codes.Ok, "")
	return dataBatch, total, nil
}

func (dvs *dataViewService) queryBySQL(ctx context.Context, query interfaces.ViewQueryInterface,
	view *interfaces.DataView) (resBytes []byte, total int64, err error) {
	commonParams := query.GetCommonParams()

	// 优先使用查询接口指定的 sql
	selectSql := commonParams.SqlStr
	if selectSql == "" {
		if view.SQLStr != "" {
			// 原子视图还存着sql_str
			selectSql = view.SQLStr
		} else {
			// 实时生成sql
			selectSql, err = buildViewSql(ctx, view)
			if err != nil {
				return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
					WithErrorDetails(err.Error())
			}
		}
	}

	// 添加时间过滤
	timeFilterSql := buildTimeFilterSql(commonParams.DateField, commonParams.Start, commonParams.End)
	// 全局过滤条件, 全局过滤条件选择的字段要在视图列表里
	globalFilterSql, err := buildSQLCondition(ctx, query.GetGlobalFilters(), view.Type, view.FieldsMap)
	if err != nil {
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(err.Error())
	}
	// 视图的行列规则不会应用在给指标模型生成的sql上，所以在这里添加行列规则过滤
	rowColumnRules := query.GetRowColumnRules()
	rowColumnRulesSQL, newFields, newFieldsMap, err := buildRowColumnRulesSQL(ctx, rowColumnRules, view)
	if err != nil {
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(err.Error())
	}
	// 更新视图字段为行列规则定义的字段
	defer func() {
		view.Fields = newFields
		view.FieldsMap = newFieldsMap
	}()

	// 将全局过滤条件、时间过滤和视图sql一起拼sql，全局过滤条件、时间过滤均为可选项
	var sqlStr string
	var whereClauses []string
	// 收集非空的过滤条件
	if timeFilterSql != "" {
		whereClauses = append(whereClauses, timeFilterSql)
	}
	if globalFilterSql != "" {
		whereClauses = append(whereClauses, globalFilterSql)
	}
	if rowColumnRulesSQL != "" {
		whereClauses = append(whereClauses, rowColumnRulesSQL)
	}

	builder := NewSQLBuilder(selectSql)
	builder.AddWheres(whereClauses)
	sqlStr = builder.Build()

	// 记录查询vega耗时
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime).Milliseconds()
		logger.Infof("query vega data cost time is %dms", elapsed)
		query.SetVegaDuration(elapsed)
	}()

	// 查询总数
	if commonParams.NeedTotal {
		countSql := buildCountSql(sqlStr)
		logger.Infof("get total count sqlStr is %s", countSql)
		result, err := dvs.vgAccess.FetchDataNoUnmarshal(ctx, &interfaces.FetchVegaDataParams{
			IsSingleDataSource: isSingleDataSource(view),
			QueryType:          interfaces.QueryType_SQL,
			DataSourceID:       getQueryDataSourceID(view),
			SqlStr:             countSql,
			NextUri:            "",
			UseSearchAfter:     false,
		})
		if err != nil {
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails(err.Error())
		}

		// 读取count的结果
		total, err = readCountResult(ctx, isSingleDataSource(view), result)
		if err != nil {
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails(err.Error())
		}
		logger.Infof("total count is %d", total)

		// 如果总条数为 0, 不需要再查询, columns的元数据在视图对象详情里
		if total == 0 {
			return []byte{}, total, nil
		}
	}

	var keys []string
	sfParams := query.GetSearchAfterParams()
	if sfParams != nil {
		for _, key := range sfParams.SearchAfter {
			keys = append(keys, fmt.Sprintf("%v", key))
		}
	}

	finalSql := sqlStr
	// 拼接排序
	sortParams := prepareSQLSortParams(query.GetSortParams(), view.FieldsMap)
	if len(sortParams) > 0 {
		sortSql := buildSQLSortParams(sortParams)
		if sortSql != "" {
			finalSql = fmt.Sprintf("%s ORDER BY %s", finalSql, sortSql)
		}
	}

	// 如果不翻页查询，并且传递了limit参数，并且sql语句里没有limit, sql里拼接上limit
	if !commonParams.UseSearchAfter && commonParams.Limit > 0 {
		finalSql = AddLimitIfMissing(finalSql, commonParams.Limit)
		// finalSql = fmt.Sprintf("%s LIMIT %d", finalSql, commonParams.Limit)
	}

	logger.Infof("fetch data sqlStr is [%s]", finalSql)

	timeout := query.GetQueryParams()[interfaces.QueryParam_Timeout].(time.Duration)
	timeoutSecond := int64(timeout.Seconds())
	nextUri := strings.Join(keys, "/")
	fetchParams := &interfaces.FetchVegaDataParams{
		IsSingleDataSource: isSingleDataSource(view),
		QueryType:          interfaces.QueryType_SQL,
		DataSourceID:       getQueryDataSourceID(view),
		NextUri:            nextUri,
		SqlStr:             finalSql,
		UseSearchAfter:     commonParams.UseSearchAfter,
		Limit:              commonParams.Limit,
		Timeout:            timeoutSecond,
	}
	dataBatch, err := dvs.vgAccess.FetchDataNoUnmarshal(ctx, fetchParams)
	if err != nil {
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(err.Error())
	}

	return dataBatch, total, nil
}

func (dvs *dataViewService) GetTotalByIndexBase(ctx context.Context, dslBuffer bytes.Buffer, indices []string) (total int64, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Query single view data")
	defer span.End()

	// 反序列化 DSL
	var dsl map[string]any
	err = sonic.Unmarshal(dslBuffer.Bytes(), &dsl)
	if err != nil {
		logger.Errorf("Unmarshal data view query dslBytes failed, %s", err.Error())
		logger.Errorf("Current view DSL json: %q", dslBuffer.Bytes())
		return 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_UnmarshalFailed).WithErrorDetails(err.Error())
	}

	delete(dsl, "pit")
	delete(dsl, "from")
	delete(dsl, "size")
	delete(dsl, "sort")
	delete(dsl, "track_scores")
	delete(dsl, "track_total_hits")
	totalBytes, _, err := dvs.osAccess.Count(ctx, dsl, indices)
	if err != nil {
		span.SetStatus(codes.Error, "Search total documents count failed")
		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError_CountFailed).
			WithErrorDetails(err.Error())
	}

	totalNode, err := sonic.Get(totalBytes, "count")
	if err != nil {
		span.SetStatus(codes.Error, "Get total documents count failed")
		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError_CountFailed).
			WithErrorDetails(err.Error())
	}

	total, err = totalNode.Int64()
	if err != nil {
		span.SetStatus(codes.Error, "Convert total documents count to type int64 failed")
		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError_CountFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return total, nil
}

// func (dvs *dataViewService) GetTotalByDSL(ctx context.Context, dsl interfaces.DSLCfg, indices []string) (total int64, err error) {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Query single view data")
// 	defer span.End()

// 	delete(dsl, "pit")
// 	delete(dsl, "from")
// 	delete(dsl, "size")
// 	delete(dsl, "sort")
// 	totalBytes, _, err := dvs.osAccess.Count(ctx, dsl, indices)
// 	if err != nil {
// 		span.SetStatus(codes.Error, "Search total documents count failed")
// 		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError_CountFailed).
// 			WithErrorDetails(err.Error())
// 	}

// 	totalNode, err := sonic.Get(totalBytes, "count")
// 	if err != nil {
// 		span.SetStatus(codes.Error, "Get total documents count failed")
// 		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError_CountFailed).
// 			WithErrorDetails(err.Error())
// 	}

// 	total, err = totalNode.Int64()
// 	if err != nil {
// 		span.SetStatus(codes.Error, "Convert total documents count to type int64 failed")
// 		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError_CountFailed).
// 			WithErrorDetails(err.Error())
// 	}

// 	span.SetStatus(codes.Ok, "")
// 	return total, nil
// }

// 从视图获取索引列表, 返回 catalogName， viewIndicesMap（视图id到索引列表的映射）
func (dvs *dataViewService) getIndicesByView(view *interfaces.DataView) (string, []string, map[string][]string, error) {
	// 原子视图的索引列表是原子视图的技术名称，自定义视图的索引列表是dataScope参数里视图节点的技术名称组合
	switch view.Type {
	case interfaces.ViewType_Atomic:
		return view.DataSourceCatalog, []string{view.TechnicalName}, map[string][]string{view.ViewID: {view.TechnicalName}}, nil
	case interfaces.ViewType_Custom:
		var catalog string
		catalogMap := map[string]struct{}{}
		indices := []string{}
		viewIndicesMap := map[string][]string{}
		// 判断多个 view node 的 catalog 是否一致
		for _, node := range view.DataScope {
			if node.Type == interfaces.DataScopeNodeType_View {
				var cfg interfaces.ViewNodeCfg
				err := mapstructure.Decode(node.Config, &cfg)
				if err != nil {
					return "", nil, nil, fmt.Errorf("failed to decode view node config, %s", err.Error())
				}

				if cfg.View == nil {
					return "", nil, nil, fmt.Errorf("view is nil")
				}
				indices = append(indices, cfg.View.TechnicalName)
				viewIndicesMap[cfg.View.ViewID] = append(viewIndicesMap[cfg.View.ViewID], cfg.View.TechnicalName)

				catalog = cfg.View.DataSourceCatalog
				catalogMap[catalog] = struct{}{}
			}
		}

		if len(catalogMap) > 1 {
			return "", nil, nil, fmt.Errorf("custom view %s has different catalog %v", view.ViewID, catalogMap)
		}

		return catalog, indices, viewIndicesMap, nil
	default:
		return "", nil, nil, fmt.Errorf("data view type %s is not supported", view.Type)
	}
}

// 从索引库获取索引和分片信息
func (dvs *dataViewService) GetIndices(ctx context.Context, baseTypes []string, start int64,
	end int64) ([]*interfaces.IndexShards, []string, int, error) {

	if len(baseTypes) == 0 {
		return nil, nil, http.StatusOK, nil
	}

	indicesResult, respCode, err := dvs.ibAccess.GetIndices(ctx, baseTypes, start, end)
	if err != nil {
		return nil, nil, respCode, err
	}

	var indices []string
	var indexShardsArr []*interfaces.IndexShards

	for _, indice := range indicesResult["indices"] {
		indexShardsArr = append(indexShardsArr, &interfaces.IndexShards{
			IndexName: indice.IndexName,
			Pri:       strconv.Itoa(indice.ShardNum),
			ShardNum:  indice.ShardNum,
		})

		indices = append(indices, indice.IndexName)
	}

	return indexShardsArr, indices, respCode, nil
}

// 内部模块调用视图返回的视图结果
func (dvs *dataViewService) GetViewResInternal(ctx context.Context, query *interfaces.DataViewQueryV1,
	view *interfaces.DataView, content []byte, total int64) (*interfaces.ViewInternalResponse, error) {

	if len(content) == 0 {
		return &interfaces.ViewInternalResponse{
			ScrollId: "",
			View:     view,
			Total:    0,
			Datas:    []*ast.Node{},
		}, nil
	}

	rootNode, err := sonic.Get(content)
	if err != nil {
		detail := fmt.Sprintf("parse root failed, %s", err.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_GetDocumentsFailed).WithErrorDetails(detail)
	}

	docs, err := rootNode.GetByPath("hits", "hits").ArrayUseNode()
	if err != nil {
		detail := fmt.Sprintf("docsNode convert to arrayNode failed, %s", err.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_GetDocumentsFailed).WithErrorDetails(detail)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(40)

	results := make([]*ast.Node, len(docs))
	for i, doc := range docs {
		i, doc := i, doc

		g.Go(func() error {
			originNode := doc.Get("_source")
			// 添加 _index 字段
			docIndex := doc.Get("_index")
			docScore := doc.Get("_score")

			// val := originNode.Get("__id")
			// if !val.Exists() {
			// 	docId := doc.Get("_id")
			// 	_, err = originNode.Set("__id", *docId)
			// 	if err != nil {
			// 		detail := fmt.Sprintf("set '__id' to source field failed, %s", err.Error())
			// 		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			// 			uerrors.Uniquery_DataView_InternalError_SetDocIdFailed).WithErrorDetails(detail)
			// 	}
			// }

			_, err = originNode.Set("__index", *docIndex)
			if err != nil {
				detail := fmt.Sprintf("set '__index' to source field failed, %s", err.Error())
				return rest.NewHTTPError(ctx, http.StatusInternalServerError,
					uerrors.Uniquery_DataView_InternalError_SetDocIndexFailed).WithErrorDetails(detail)
			}

			_, err = originNode.Set("_score", *docScore)
			if err != nil {
				detail := fmt.Sprintf("set '_score' to source field failed, %s", err.Error())
				return rest.NewHTTPError(ctx, http.StatusInternalServerError,
					rest.PublicError_InternalServerError).WithErrorDetails(detail)
			}

			results[i] = originNode
			return nil
		})
	}

	err = g.Wait()
	if err != nil {
		return nil, err
	}

	var scrollId string
	scrollIdNode := rootNode.Get("_scroll_id")
	if scrollIdNode.Exists() {
		scrollId, err = scrollIdNode.String()
		if err != nil {
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_InternalError_GetScrollIdFailed).WithErrorDetails(err.Error())
		}
	}

	internalResponse := &interfaces.ViewInternalResponse{
		ScrollId: scrollId,
		View:     view,
		Total:    total,
		Datas:    results,
	}

	return internalResponse, nil
}

type taskContext struct {
	hasOutputFields   bool
	includeIndexField bool
	includeScoreField bool
	finalFieldsMap    map[string]*cond.ViewField
	results           []map[string]any

	wg    sync.WaitGroup
	errCh chan error
}

// 将结果转成视图查询的统一格式
func convertToViewUniResponse(ctx context.Context, query interfaces.ViewQueryInterface,
	view *interfaces.DataView, content []byte, total int64) (*interfaces.ViewUniResponseV2, error) {

	includeView := query.GetQueryParams()[interfaces.QueryParam_IncludeView].(bool)

	if len(content) == 0 {
		if !includeView {
			view = nil
		}

		// 如果不需要 total 计数，total 设为 nil, 不返回此参数
		var totalCount *int64
		if query.GetCommonParams().NeedTotal {
			totalCount = &total
		} else {
			totalCount = nil
		}

		return &interfaces.ViewUniResponseV2{
			PitID:       "",
			SearchAfter: nil,
			View:        view,
			Entries:     []map[string]any{},
			TotalCount:  totalCount,
			ScrollId:    "",
		}, nil
	}

	switch view.QueryType {
	case interfaces.QueryType_DSL:
		return convertToViewUniResponseByDSL(ctx, query, view, content, total)
	case interfaces.QueryType_SQL:
		return convertToViewUniResponseBySQL(ctx, query, view, content, total)
	case interfaces.QueryType_IndexBase:
		return convertToViewUniResponseByIndexBase(ctx, query, view, content, total)
	default:
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails("view query type must be DSL or SQL or IndexBase")
	}
}

func convertToViewUniResponseByDSL(ctx context.Context, query interfaces.ViewQueryInterface,
	view *interfaces.DataView, content []byte, total int64) (*interfaces.ViewUniResponseV2, error) {

	rootNode, err := sonic.Get(content)
	if err != nil {
		detail := fmt.Sprintf("SQL parse root failed, %s", err.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_GetDocumentsFailed).WithErrorDetails(detail)
	}

	var data []ast.Node
	dataNode := rootNode.Get("data")
	if dataNode.Exists() {
		data, err = dataNode.ArrayUseNode()
		if err != nil {
			detail := fmt.Sprintf("SQL dataNode convert to arrayNode failed, %s", err.Error())
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_InternalError_GetDocumentsFailed).WithErrorDetails(detail)
		}
	}

	if len(data) == 0 {
		return &interfaces.ViewUniResponseV2{
			PitID:       "",
			SearchAfter: nil,
			View:        view,
			Entries:     []map[string]any{},
			TotalCount:  nil,
			ScrollId:    "",
		}, nil
	}

	// total 从 trace_total_hits的结果来
	if query.GetCommonParams().NeedTotal {
		trackTotalHitsResult, err := data[0].GetByPath("hits", "total", "value").Int64()
		if err != nil {
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails(fmt.Sprintf("Get DSL query total value failed, err: %s", err.Error()))
		}
		total = trackTotalHitsResult
	}

	resp, err := commonConvertToViewDSLUniResponse(ctx, query, view, data[0], total)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func convertToViewUniResponseByIndexBase(ctx context.Context, query interfaces.ViewQueryInterface,
	view *interfaces.DataView, content []byte, total int64) (*interfaces.ViewUniResponseV2, error) {
	rootNode, err := sonic.Get(content)
	if err != nil {
		detail := fmt.Sprintf("DSL parse root failed, %s", err.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_GetDocumentsFailed).WithErrorDetails(detail)
	}

	resp, err := commonConvertToViewDSLUniResponse(ctx, query, view, rootNode, total)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func commonConvertToViewDSLUniResponse(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView,
	rootNode ast.Node, total int64) (*interfaces.ViewUniResponseV2, error) {
	docs, err := rootNode.GetByPath("hits", "hits").ArrayUseNode()
	if err != nil {
		detail := fmt.Sprintf("DSL docsNode convert to arrayNode failed, %s", err.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_GetDocumentsFailed).WithErrorDetails(detail)
	}

	results := make([]map[string]interface{}, len(docs))
	errCh := make(chan error, len(docs))
	defer close(errCh)

	var includeIndexField, includeScoreField bool
	outputFields := query.GetCommonParams().OutputFields
	finalFieldsMap := make(map[string]*cond.ViewField)
	if len(outputFields) == 0 {
		// 没有指定输出字段，默认输出视图的字段
		finalFieldsMap = view.FieldsMap
		includeIndexField = true
		includeScoreField = true
	} else {
		// 有指定输出字段，检查是否包含 __index 和 _score 字段
		for _, of := range outputFields {
			if value, ok := view.FieldsMap[of]; ok {
				finalFieldsMap[of] = value
			} else {
				logger.Errorf("SQL output field '%s' not found in view '%s'", of, view.ViewName)
			}

			if of == "__index" {
				includeIndexField = true
			}
			if of == "_score" {
				includeScoreField = true
			}
		}
	}

	taskCtx := &taskContext{
		hasOutputFields:   len(outputFields) > 0,
		includeIndexField: includeIndexField,
		includeScoreField: includeScoreField,
		finalFieldsMap:    finalFieldsMap,
		results:           results,
		wg:                sync.WaitGroup{},
		errCh:             errCh,
	}

	for i, doc := range docs {
		taskCtx.wg.Add(1)

		err := viewPool.Submit(processDocByDSL(taskCtx, view, i, doc, query.GetCommonParams().Format))
		if err != nil {
			detail := fmt.Sprintf("DSL submit task of processing a document failed, %s", err.Error())
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_InternalError_SubmitTaskFailed).WithErrorDetails(detail)
		}
	}
	taskCtx.wg.Wait()

	if len(taskCtx.errCh) > 0 {
		err := <-taskCtx.errCh

		detail := fmt.Sprintf("DSL process a document failed, %s", err.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_ProcessDocFailed).WithErrorDetails(detail)
	}

	var scrollId string
	scrollIdNode := rootNode.Get("_scroll_id")
	if scrollIdNode.Exists() {
		scrollId, err = scrollIdNode.String()
		if err != nil {
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_InternalError_GetScrollIdFailed).WithErrorDetails(err.Error())
		}
	}

	var pitID string
	pitIDNode := rootNode.Get("pit_id")
	if pitIDNode.Exists() {
		pitID, err = pitIDNode.String()
		if err != nil {
			logger.Errorf("DSL get pit_id failed, %s", err.Error())
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_InternalError_GetPitIdFailed).WithErrorDetails(err.Error())
		}
	}

	var searchAfter []any
	if len(docs) > 0 {
		searchAfterNode := docs[len(docs)-1].Get("sort")
		if searchAfterNode.Exists() {
			searchAfter, err = searchAfterNode.Array()
			if err != nil {
				logger.Errorf("DSL get search_after failed, %s", err.Error())
				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					uerrors.Uniquery_DataView_InternalError_GetSearchAfterValueFailed).WithErrorDetails(err.Error())
			}
		}
	} else {
		searchAfter = nil
	}

	includeView := query.GetQueryParams()[interfaces.QueryParam_IncludeView].(bool)
	if !includeView {
		view = nil
	}

	// 如果不需要 total 计数，total 设为 nil, 不返回此参数
	var totalCount *int64
	if query.GetCommonParams().NeedTotal {
		totalCount = &total
	} else {
		totalCount = nil
	}

	return &interfaces.ViewUniResponseV2{
		PitID:          pitID,
		SearchAfter:    searchAfter,
		View:           view,
		Entries:        results,
		TotalCount:     totalCount,
		ScrollId:       scrollId,
		VegaDurationMs: query.GetVegaDuration(),
	}, nil
}

func convertToViewUniResponseBySQL(ctx context.Context, query interfaces.ViewQueryInterface,
	view *interfaces.DataView, content []byte, total int64) (*interfaces.ViewUniResponseV2, error) {
	rootNode, err := sonic.Get(content)
	if err != nil {
		detail := fmt.Sprintf("SQL parse root failed, %s", err.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(detail)
	}

	// 解析SQL查询结果
	currentTotal, err := parseSQLTotalCount(rootNode)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}
	logger.Infof("SQL return %d documents", currentTotal)

	// 获取列信息和文档数据
	columns, err := parseSQLColumns(rootNode)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	docs, err := parseSQLDocuments(rootNode, isSingleDataSource(view))
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_GetDocumentsFailed).WithErrorDetails(err.Error())
	}

	// 处理字段映射
	err = processViewFields(view, columns, query.GetCommonParams().SqlStr)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	// 检查SQL返回的字段与视图定义的字段数量是否一致
	// if len(columns) != len(view.Fields) {
	// 	detail := fmt.Sprintf("SQL columns number not equal to view fields number, %d != %d", len(columns), len(view.Fields))
	// 	return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
	// 		uerrors.Uniquery_DataView_InternalError_GetDocumentsFailed).WithErrorDetails(detail)
	// }

	results := make([]map[string]interface{}, len(docs))
	errCh := make(chan error, len(docs))
	defer close(errCh)

	outputFields := query.GetCommonParams().OutputFields
	finalFieldsMap := make(map[string]*cond.ViewField)
	if len(outputFields) == 0 {
		finalFieldsMap = view.FieldsMap
	} else {
		for _, of := range outputFields {
			if value, ok := view.FieldsMap[of]; ok {
				finalFieldsMap[of] = value
			} else {
				logger.Errorf("SQL output field '%s' not found in view '%s'", of, view.ViewName)
			}
		}
	}

	taskCtx := &taskContext{
		hasOutputFields: len(outputFields) > 0,
		finalFieldsMap:  finalFieldsMap,
		results:         results,
		wg:              sync.WaitGroup{},
		errCh:           errCh,
	}

	for i, doc := range docs {
		taskCtx.wg.Add(1)

		err := viewPool.Submit(processDocBySQL(taskCtx, view, columns, i, doc, query.GetCommonParams().Format))
		if err != nil {
			detail := fmt.Sprintf("SQL submit task of processing a document failed, %s", err.Error())
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_InternalError_SubmitTaskFailed).WithErrorDetails(detail)
		}
	}
	taskCtx.wg.Wait()

	if len(taskCtx.errCh) > 0 {
		err := <-taskCtx.errCh

		detail := fmt.Sprintf("SQL process a document failed, %s", err.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_ProcessDocFailed).WithErrorDetails(detail)
	}

	var searchAfter []any
	if query.GetCommonParams().UseSearchAfter {
		searchAfter = extractSearchAfterParams(rootNode, isSingleDataSource(view))
	}

	includeView := query.GetQueryParams()[interfaces.QueryParam_IncludeView].(bool)
	if !includeView {
		view = nil
	}

	// 如果不需要 total 计数，total 设为 nil, 不返回此参数
	var totalCount *int64
	if query.GetCommonParams().NeedTotal {
		totalCount = &total
	} else {
		totalCount = nil
	}

	return &interfaces.ViewUniResponseV2{
		PitID:          "",
		SearchAfter:    searchAfter,
		View:           view,
		Entries:        results,
		TotalCount:     totalCount,
		ScrollId:       "",
		VegaDurationMs: query.GetVegaDuration(),
	}, nil
}

// ProcessDocByDSL 处理基于 DSL 查询的文档
func processDocByDSL(taskCtx *taskContext, view *interfaces.DataView, i int, doc ast.Node, format string) func() {
	return func() {
		defer taskCtx.wg.Done()

		originNode := doc.Get("_source")
		originString, err := originNode.Raw()
		if err != nil {
			taskCtx.errCh <- err
			return
		}

		var origin map[string]any
		d := decoder.NewDecoder(originString)
		d.UseInt64()
		if err = d.Decode(&origin); err != nil {
			logger.Errorf("processDocByDSL unmarshal result failed: %s", err.Error())
			taskCtx.errCh <- err
			return
		}

		// pick 为视图最终输出数据
		pick := make(map[string]any)
		switch format {
		case interfaces.Format_Flat:
			if view.FieldScope == interfaces.FieldScope_All {
				err = flatten(true, "", origin, pick)
			} else {
				// 平铺的时候过滤字段
				err = flattenWithPickField(true, "", origin, pick, taskCtx.finalFieldsMap)
			}
			if err != nil {
				taskCtx.errCh <- err
				return
			}
		case interfaces.Format_Original:
			// if view.FieldScope == interfaces.FieldScope_All {
			if (view.Type == interfaces.ViewType_Atomic || view.FieldScope == interfaces.FieldScope_All) && !taskCtx.hasOutputFields {
				pick = origin
			} else {
				err = pickData(origin, pick, taskCtx.finalFieldsMap)
				if err != nil {
					taskCtx.errCh <- err
					return
				}
			}
		default:
			err = fmt.Errorf("unsupported format: %s", format)
			taskCtx.errCh <- err
		}

		// _, ok := pick["__id"]
		// if !ok {
		// 	docId, err := doc.Get("_id").String()
		// 	if err != nil {
		// 		taskCtx.errCh <- err
		// 		return
		// 	}
		// 	pick["__id"] = docId
		// }

		// 添加 _index和_score 字段
		if taskCtx.includeIndexField {
			docIndex, err := doc.Get("_index").String()
			if err != nil {
				taskCtx.errCh <- err
				return
			}
			pick["__index"] = docIndex
		}

		if taskCtx.includeScoreField {
			docScore, err := doc.Get("_score").Float64()
			if err != nil {
				taskCtx.errCh <- err
				return
			}
			pick["_score"] = docScore
		}

		taskCtx.results[i] = pick
	}
}

// ProcessDocBySQL 处理基于 SQL 的文档
func processDocBySQL(taskCtx *taskContext, view *interfaces.DataView, columns []ast.Node, i int, doc ast.Node, format string) func() {
	return func() {
		defer taskCtx.wg.Done()

		values, err := doc.ArrayUseNode()
		if err != nil {
			logger.Errorf("SQL valuesNode convert to arrayNode failed, %s", err.Error())
			taskCtx.errCh <- err
			return
		}

		origin := make(map[string]any)
		// 遍历每列，将每列的值变成 key-value 对象
		for j, col := range columns {
			fieldName, _ := col.Get("name").String()
			// fieldType, _ := col.Get("type").String()

			if j < len(values) {
				node := values[j]
				val, err := node.InterfaceUseNumber()
				if err != nil {
					logger.Errorf("SQL valueNode convert to any type failed, %s", err.Error())
					taskCtx.errCh <- err
					return
				}

				if _, ok := origin[fieldName]; ok {
					for k, v := range view.FieldsMap {
						// 对于join，可能会有原始字段重复，二次出现的重复字段拼接为name_srcNodeName
						if strings.HasPrefix(k, fmt.Sprintf("%s_", fieldName)) && v.OriginalName == fieldName {
							origin[v.Name] = val
						}
					}
				} else {
					origin[fieldName] = val
				}
			}
		}

		// pick 为视图最终输出数据
		pick := make(map[string]any)
		switch format {
		case interfaces.Format_Flat:
			// 平铺的时候过滤字段
			if view.FieldScope == interfaces.FieldScope_All {
				err = flatten(true, "", origin, pick)
			} else {
				err = flattenWithPickField(true, "", origin, pick, taskCtx.finalFieldsMap)
			}
			if err != nil {
				taskCtx.errCh <- err
				return
			}
		case interfaces.Format_Original:
			// if view.FieldScope == interfaces.FieldScope_All {
			// 当视图类型为原子视图且未指定输出字段时，直接返回原始文档
			if (view.Type == interfaces.ViewType_Atomic || view.FieldScope == interfaces.FieldScope_All) && !taskCtx.hasOutputFields {
				pick = origin
			} else {
				err := pickData(origin, pick, taskCtx.finalFieldsMap)
				if err != nil {
					taskCtx.errCh <- err
					return
				}
			}
		default:
			err := fmt.Errorf("unsupported format: %s", format)

			taskCtx.errCh <- err
		}

		taskCtx.results[i] = pick
	}
}

// 对数据做过滤，只包含视图字段
func pickData(origin, pick map[string]any, fieldsMap map[string]*cond.ViewField) error {
	// 过滤字段， 转成pick
	for _, field := range fieldsMap {
		myField := field
		value, isSliceValue, err := getData(origin, myField)
		if err != nil {
			return err
		}

		err = setData(myField, pick, value, isSliceValue)
		if err != nil {
			return err
		}
	}

	return nil
}

// 合并索引库字段
func mergeIndexBaseFields(mappings interfaces.Mappings) []interfaces.IndexBaseField {
	capacity := len(mappings.MetaMappings) + len(mappings.DynamicMappings) + len(mappings.UserDefinedMappings)
	allBaseFields := make([]interfaces.IndexBaseField, 0, capacity)

	allBaseFields = append(allBaseFields, mappings.MetaMappings...)
	allBaseFields = append(allBaseFields, mappings.DynamicMappings...)
	allBaseFields = append(allBaseFields, mappings.UserDefinedMappings...)

	return allBaseFields
}

// 拼接新的字段名: a.b.c
func joinKey(top bool, prefix, subkey string) string {
	if subkey == "" {
		return prefix
	}

	key := prefix

	if top {
		key += subkey
	} else {
		key += "." + subkey
	}

	return key
}

func getData(origin map[string]any, field *cond.ViewField) ([]any, bool, error) {
	// field.InitFieldPath()
	oDatas, isSliceValue, err := GetDatasByPath(origin, field.Path)
	if err != nil {
		return nil, isSliceValue, err
	}
	if len(oDatas) == 0 {
		logger.Debugf("this field %s does not exist in the original data", strings.Join(field.Path, "."))
		return nil, isSliceValue, nil
	}

	// vData := []any{}
	// for _, oData := range oDatas {
	// 	v, err := ConvertValueByType(field.Type, oData)
	// 	if err != nil {
	// 		return nil, isSliceValue, err
	// 	}
	// 	vData = append(vData, v)
	// }

	return oDatas, isSliceValue, nil
}

// result coule be single value or slice value
// result type should only be string, time.Time, int64, float64, bool
// func ConvertValueByType(vType string, vData any) (any, error) {
// 	if vData == nil {
// 		return nil, fmt.Errorf("vData is nil")
// 	}
// 	switch v := vData.(type) {
// 	case string:
// 		switch vType {
// 		case interfaces.DATATYPE_DATETIME:
// 			v, err := time.Parse(time.RFC3339Nano, vData.(string))
// 			if err != nil {
// 				return nil, err
// 			}
// 			return v, nil
// 		case interfaces.DATATYPE_KEYWORD, interfaces.DATATYPE_TEXT:
// 			return v, nil
// 		default:
// 			return nil, fmt.Errorf("invalid value")
// 		}

// 	case float64:
// 		switch vType {
// 		case interfaces.DATATYPE_SHORT:
// 			v := int64(int16(vData.(float64)))
// 			return v, nil
// 		case interfaces.DATATYPE_INTEGER:
// 			v := int64(int32(vData.(float64)))
// 			return v, nil
// 		case interfaces.DATATYPE_LONG:
// 			v := int64(vData.(float64))
// 			return v, nil
// 		case interfaces.DATATYPE_FLOAT:
// 			v := float64(float32(vData.(float64)))
// 			return v, nil
// 		case interfaces.DATATYPE_DOUBLE:
// 			return v, nil
// 		default:
// 			return nil, fmt.Errorf("invalid value")
// 		}

// 	case bool:
// 		switch vType {
// 		case interfaces.DATATYPE_BOOLEAN:
// 			return v, nil
// 		default:
// 			return nil, fmt.Errorf("invalid value")
// 		}

// 	default:
// 		return nil, fmt.Errorf("invalid value")
// 	}
// }

// array 里面相同类型 可以获取内部数据，如果非相同类型，则nil
func GetDatasByPath(obj any, path []string) ([]any, bool, error) {
	if reflect.TypeOf(obj) == nil {
		return []any{}, false, nil
	}

	current := obj
	for idx := 0; idx < len(path); idx++ {
		switch reflect.TypeOf(current).Kind() {
		case reflect.Map:
			value, ok := current.(map[string]any)[path[idx]]
			if !ok || value == nil {
				return []any{}, false, nil
			}
			// found
			current = value

		case reflect.Slice:
			res := []any{}
			for _, sub := range current.([]any) {
				subRes, isSliceValue, err := GetDatasByPath(sub, path[idx:])
				if err != nil {
					return []any{}, isSliceValue, err
				}
				res = append(res, subRes...)
			}
			return res, true, nil

		default:
			// invalid path
			return []any{}, false, nil
		}
	}

	// path is empty
	return GetLastDatas(current)
}

func GetLastDatas(obj any) (res []any, isSliceValue bool, err error) {
	if obj == nil {
		return []any{}, isSliceValue, nil
	}
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Slice:
		isSliceValue = true
		for _, sub := range obj.([]any) {
			subRes, isSliceValue, err := GetLastDatas(sub)
			if err != nil {
				return []any{}, isSliceValue, err
			}
			res = append(res, subRes...)
		}
	default:
		res = append(res, obj)
		return res, isSliceValue, nil
	}

	return res, isSliceValue, nil
}

// 写入k v
// 子节点不存在则构建树结构往根树并入
// 比如 {"a":{"b":"c"}} 写入 ["a", "d"] value 为 123
// 则最终为 {"a":{"b":"c","d":123}}
func setData(field *cond.ViewField, obj map[string]any, data []any, isSliceValue bool) error {
	if len(data) == 0 {
		return nil
	}

	current := obj
	// field.InitFieldPath()
	for idx := 0; idx < len(field.Path)-1; idx++ {
		if value, ok := current[field.Path[idx]]; ok {
			if reflect.TypeOf(value).Kind() == reflect.Map {
				current = value.(map[string]any)
			} else {
				return fmt.Errorf("current path is not a map: %s", field.Path[idx])
			}
		} else {
			tmp := make(map[string]interface{})
			current[field.Path[idx]] = tmp
			current = tmp
		}
	}
	if len(data) == 1 && !isSliceValue {
		current[field.Path[len(field.Path)-1]] = data[0]
	} else {
		current[field.Path[len(field.Path)-1]] = data
	}

	return nil
}

func (dvs *dataViewService) CountMultiFields(ctx context.Context, viewID string, query *interfaces.DataViewQueryV1, fields []string, sep string) (countMap map[string]int64, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: count by multiple fields")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 判断field个数
	if len(fields) == 0 {
		return nil, nil
	}

	// 2. parseQuery
	view, err := dvs.parseQuery(ctx, viewID)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return nil, err
	}

	for i, field := range fields {
		v, ok := view.FieldsMap[field]
		if !ok {
			return nil, nil
		}

		if v.Type == interfaces.TEXT_TYPE {
			field = field + "." + dtype.KEYWORD_SUFFIX
		}

		fields[i] = field
	}

	// 3. 获取数据视图底层的索引, 暂时仅支持元数据视图的 DSL 视图
	indices, err := dvs.acquireIndices(ctx, view.DataSourceID, query.Start, query.End)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return nil, err
	}

	// 如果索引列表为空，则返回空数据
	if len(indices) == 0 {
		span.SetStatus(codes.Ok, "No indices found")
		return map[string]int64{}, nil
	}

	// 4. prepare dsl parameters
	paras, err := dvs.prepareDSLParas(ctx, query, view, fields)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return nil, err
	}

	// 5. 查找结果
	stats := make(map[string]int64)
	for {
		// 5.1 转成dsl
		dsl, err := dvs.generateMultiFieldStatsDSL(ctx, paras, fields)
		if err != nil {
			o11y.Error(ctx, err.Error())
			return nil, err
		}

		// 5.2 调用SearchSubmit 查询
		resBytes, _, err := dvs.osAccess.SearchSubmit(ctx, dsl, indices, 0, interfaces.DEFAULT_PREFERENCE, false)
		if err != nil {
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_InternalError_SearchSubmitFailed).WithErrorDetails(err.Error())
		}

		// 5.3 处理返回结果, 更新
		resJson := string(resBytes)
		bucketStr := gjson.Get(resJson, "aggregations.group_by_fields.buckets").String()

		var buckets []Bucket
		err = sonic.Unmarshal([]byte(bucketStr), &buckets)
		if err != nil {
			o11y.Error(ctx, err.Error())
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_InternalError_UnmarshalFailed).WithErrorDetails(err.Error())
		}

		for _, bucket := range buckets {
			if keyArray, ok := bucket.Key.([]any); ok {
				vals := make([]string, 0)
				for i := 0; i < len(keyArray); i++ {
					vals = append(vals, fmt.Sprint(keyArray[i]))
				}

				spanID := strings.Join(vals, sep)
				stats[spanID] = bucket.DocCount
			} else {
				spanID := fmt.Sprint(bucket.Key)
				stats[spanID] = bucket.DocCount
			}
		}

		// 5.4 判断是否需要继续分页查询
		length := len(buckets)
		if length < interfaces.MAX_SEARCH_RELATED_LOGS_BUCKET {
			break
		}

		// 5.5 找到criticalKey
		criticalKey := ""
		if len(fields) == 1 {
			criticalKey = fmt.Sprint(buckets[length-1].Key)
		} else {
			criticalKey = buckets[length-1].KeyAsString
		}

		// 5.6 构造script, 并更新paras.scriptFilterStr
		str := ""
		for i, field := range fields {
			if i == len(fields) {
				str = str + "doc['" + field + "'].value"
			} else {
				str = str + "doc['" + field + "'].value + '|'"
			}
		}

		script := fmt.Sprintf("String str = %s; str.compareTo('%s') > 0", str, criticalKey)
		scriptFilterStr := fmt.Sprintf(`
				{
					"script": {
						"script": {
							"source": %s
						}
					}
				},`, script)
		paras.scriptFilterStr = scriptFilterStr
	}

	return stats, nil
}

func (dvs *dataViewService) parseQuery(ctx context.Context, viewID string) (*interfaces.DataView, error) {
	// 1. 获取视图对象信息
	view, err := dvs.GetDataViewByID(ctx, viewID, false)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_GetDataViewByIDFailed).WithErrorDetails(err.Error())
	}

	// 2. 构建map, key为fieldName, value为fieldType
	viewFieldsMap := make(map[string]*cond.ViewField)
	for _, field := range view.Fields {
		viewFieldsMap[field.Name] = field
	}

	// 3. 更新view
	view.FieldsMap = viewFieldsMap

	return view, nil
}

// func (dvs *dataViewService) acquireIndices(ctx context.Context, dataSource map[string]any, start, end int64) ([]string, error) {
// TODO 暂时仅支持元数据视图的 DSL 视图
func (dvs *dataViewService) acquireIndices(ctx context.Context, dataSourceID string, start, end int64) ([]string, error) {
	// 1. 获取索引库信息
	// var bases interfaces.SimpleIndexBases
	// err := mapstructure.Decode(dataSource[interfaces.INDEX_BASE], &bases)
	// if err != nil {
	// 	return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataSource).
	// 		WithErrorDetails(fmt.Sprintf("mapstructure decode dataSource failed, %s", err.Error()))
	// }

	// baseTypes := make([]string, 0, len(bases))
	// for _, base := range bases {
	// 	baseTypes = append(baseTypes, base.BaseType)
	// }

	baseTypes := []string{dataSourceID}

	// 2. 使用索引优化接口获取索引
	_, indices, _, err := dvs.GetIndices(ctx, baseTypes, start, end)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_GetIndicesFailed).WithErrorDetails(err.Error())
	}

	return indices, nil
}

func (dvs *dataViewService) prepareDSLParas(ctx context.Context, query *interfaces.DataViewQueryV1, view *interfaces.DataView, fields []string) (multiFieldStatsDSLParas, error) {
	var paras multiFieldStatsDSLParas

	// 1. 拼接baseFilter字符串，暂时仅支持元数据视图的 DSL 视图
	// var cfg *cond.CondCfg
	// if query.GlobalFilters != nil || view.Condition != nil {
	// 	if query.GlobalFilters == nil {
	// 		cfg = view.Condition
	// 	} else if view.Condition == nil {
	// 		cfg = query.GlobalFilters
	// 	} else {
	// 		cfg = &cond.CondCfg{
	// 			Operation: cond.OperationAnd,
	// 			SubConds: []*cond.CondCfg{
	// 				query.GlobalFilters,
	// 				view.Condition,
	// 			},
	// 		}
	// 	}
	// }

	cfg := query.GlobalFilters
	// 创建一个包含查询类型的上下文
	ctx = context.WithValue(ctx, cond.CtxKey_QueryType, view.QueryType)

	// 将过滤条件拼接到 dsl 的 query bool filter中, 链路模型使用，不需要打分
	CondCfg, _, err := cond.NewCondition(ctx, cfg, view.Type, view.FieldsMap)
	if err != nil {
		return paras, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_Filters).
			WithErrorDetails(fmt.Sprintf("failed to new condition, %s", err.Error()))
	}

	if CondCfg != nil {
		paras.baseFilterStr, err = CondCfg.Convert(ctx)
		if err != nil {
			return paras, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_Filters).
				WithErrorDetails(fmt.Sprintf("failed to convert condition to dsl, %s", err.Error()))
		}
	}

	// 2. 拼接timeFilter字符串
	timeFilterStr := ""
	if query.Start != 0 && query.End != 0 {
		timeFilterStr = fmt.Sprintf(`
				{
					"range": {
						"@timestamp": {
							"gte": %d,
							"lte": %d
						}
					}
				}`, query.Start, query.End)
	} else {
		timeFilterStr = `
				{
					"range": {
						"@timestamp": {}
					}
				}`
	}
	paras.timeFilterStr = timeFilterStr

	// 3. 拼接aggs terms字符串
	aggTermStr := ""
	for i, field := range fields {
		if i == len(fields)-1 {
			aggTermStr = fmt.Sprintf("%s%s", aggTermStr, fmt.Sprintf(`
					{
						"field": "%s"
					}`, field))
			break
		}

		aggTermStr = fmt.Sprintf("%s%s", aggTermStr, fmt.Sprintf(`
				{
					"field": "%s"
				},`, field))
	}

	paras.aggTermStr = aggTermStr
	return paras, nil
}

func (dvs *dataViewService) generateMultiFieldStatsDSL(ctx context.Context, paras multiFieldStatsDSLParas, fields []string) (map[string]any, error) {
	var s []string
	if paras.baseFilterStr != "" {
		s = append(s, paras.baseFilterStr)
	}

	if paras.scriptFilterStr != "" {
		s = append(s, paras.scriptFilterStr)
	}

	if paras.timeFilterStr != "" {
		s = append(s, paras.timeFilterStr)
	}

	var dslStr string
	// todo: 适配1.x版本的opensearch, 不支持multi_terms
	if len(fields) == 1 {
		dslStr = fmt.Sprintf(`{
			"size": 0,
			"query": {
				"bool": {
					"filter": [
						%s
					]
				}
			},
			"aggs": {
				"group_by_fields": {
					"terms": {
						"field": "%s",
						"size": %d
					}
				}
			}
		}`, strings.Join(s, ","), fields[0], interfaces.MAX_SEARCH_RELATED_LOGS_BUCKET)
	} else {
		dslStr = fmt.Sprintf(`{
		"size": 0,
		"query": {
			"bool": {
				"filter": [
					%s
				]
			}
		},
		"aggs": {
			"group_by_fields": {
				"multi_terms": {
					"terms":[
						%s
					],
					"size": %d
				}
			}
		}
	}`, strings.Join(s, ","), paras.aggTermStr, interfaces.MAX_SEARCH_RELATED_LOGS_BUCKET)
	}

	var dslBuffer bytes.Buffer
	dslBuffer.WriteString(dslStr)

	var dsl map[string]any
	err := sonic.Unmarshal(dslBuffer.Bytes(), &dsl)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_UnmarshalFailed).WithErrorDetails(err.Error())
	}

	return dsl, nil
}

// 判断全局过滤条件的字段在不在视图字段里，在返回true，不在则返回false
func checkConditionFieldExist(viewFieldsMap map[string]*cond.ViewField, cfg *cond.CondCfg) (string, bool) {
	if cfg == nil {
		return "", true
	}

	// 判断过滤器是否为空对象 {}
	if cfg.Name == "" && cfg.Operation == "" && len(cfg.SubConds) == 0 && cfg.ValueFrom == "" && cfg.Value == nil {
		return "", true
	}

	switch cfg.Operation {
	case cond.OperationAnd, cond.OperationOr:
		for _, subCond := range cfg.SubConds {
			fieldName, res := checkConditionFieldExist(viewFieldsMap, subCond)
			if !res {
				return fieldName, false
			}
		}
	case cond.OperationMultiMatch:
		// do nothing，校验留给下层做
	default:
		// 判断除 * 之外的字段权限
		if cfg.Name != cond.AllField {
			if _, ok := viewFieldsMap[cfg.Name]; !ok {
				return cfg.Name, false
			}
		}
	}

	return cfg.Name, true
}

// 从opensearch查数据，内部模块使用
func (dvs *dataViewService) GetDataFromOpenSearch(ctx context.Context, query map[string]any,
	indices []string, scroll time.Duration, preference string, trackTotalHits bool) ([]byte, int, error) {

	res, statusCode, err := dvs.osAccess.SearchSubmit(ctx,
		query, indices, scroll, preference, trackTotalHits)
	if err != nil {
		return []byte{}, statusCode, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_InternalError_SearchSubmitFailed).WithErrorDetails(err.Error())
	}

	return res, statusCode, nil
}

// 从opensearch查数据，内部模块使用
func (dvs *dataViewService) GetDataFromOpenSearchWithBuffer(ctx context.Context, query bytes.Buffer,
	indices []string, scroll time.Duration, preference string) ([]byte, int, error) {

	res, statusCode, err := dvs.osAccess.SearchSubmitWithBuffer(ctx,
		query, indices, scroll, preference)
	if err != nil {
		return []byte{}, statusCode, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_InternalError_SearchSubmitFailed).WithErrorDetails(err.Error())
	}

	return res, statusCode, nil
}

// 指标模型调用视图构建dsl或sql
func (dvs *dataViewService) BuildViewQuery4MetricModel(ctx context.Context, start, end int64,
	view *interfaces.DataView) (interfaces.ViewQuery4Metric, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Build view query for metric model")
	defer span.End()

	switch view.QueryType {
	case interfaces.QueryType_IndexBase:
		// 获取视图的索引库列表
		baseTypes, baseTypeViewMap, err := GetBaseTypes(view)
		if err != nil {
			span.SetStatus(codes.Error, "Get base types failed")
			return interfaces.ViewQuery4Metric{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
		}

		// 使用索引优化接口获取索引
		indexShards, indices, _, err := dvs.GetIndices(ctx, baseTypes, start, end)
		if err != nil {
			span.SetStatus(codes.Error, "Get indices failed")
			return interfaces.ViewQuery4Metric{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_InternalError_GetIndicesFailed).WithErrorDetails(err.Error())
		}

		// 如果索引列表为空，则返回空数据, 不需要下面拼接dsl
		if len(indices) == 0 {
			span.SetStatus(codes.Ok, "No indices found")
			return interfaces.ViewQuery4Metric{}, nil
		}

		// 视图 ID 到索引列表的映射
		viewIndicesMap, err := getViewIndicesMap(indices, baseTypeViewMap)
		if err != nil {
			span.SetStatus(codes.Error, "Get view indices map failed")
			return interfaces.ViewQuery4Metric{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
		}

		// 构建查询条件
		queryDSL, err := buildDSLQuery(ctx, view, viewIndicesMap)
		if err != nil {
			return interfaces.ViewQuery4Metric{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				rest.PublicError_InternalServerError).
				WithErrorDetails(fmt.Sprintf("failed to build query dsl, %s", err.Error()))
		}

		// 合并查询条件到主DSL结构体
		queryDSLStr, err := sonic.MarshalString(queryDSL.Query)
		if err != nil {
			return interfaces.ViewQuery4Metric{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
		}

		return interfaces.ViewQuery4Metric{
			IndexShards: indexShards,
			Indices:     indices,
			QueryStr:    queryDSLStr,
			BaseTypes:   baseTypes,
		}, nil

	case interfaces.QueryType_SQL:
		selectSql, err := buildViewSql(ctx, view)
		if err != nil {
			return interfaces.ViewQuery4Metric{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails(err.Error())
		}

		return interfaces.ViewQuery4Metric{
			QueryStr: selectSql,
		}, nil

	case interfaces.QueryType_DSL:
		// 获取索引列表, 视图 ID 到索引列表的映射
		catalogName, indices, viewIndicesMap, err := dvs.getIndicesByView(view)
		if err != nil {
			span.SetStatus(codes.Error, "Get indices failed")
			return interfaces.ViewQuery4Metric{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_InternalError_GetIndicesFailed).WithErrorDetails(err.Error())
		}

		// 如果索引列表为空，则返回空数据, 不需要下面拼接dsl
		if len(indices) == 0 {
			span.SetStatus(codes.Ok, "No indices found")
			return interfaces.ViewQuery4Metric{}, nil
		}

		// 构建查询条件
		queryDSL, err := buildDSLQuery(ctx, view, viewIndicesMap)
		if err != nil {
			return interfaces.ViewQuery4Metric{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				rest.PublicError_InternalServerError).
				WithErrorDetails(fmt.Sprintf("failed to build query dsl, %s", err.Error()))
		}

		// 合并查询条件到主DSL结构体
		queryDSLStr, err := sonic.MarshalString(queryDSL.Query)
		if err != nil {
			return interfaces.ViewQuery4Metric{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
		}

		return interfaces.ViewQuery4Metric{
			Catalog:  catalogName,
			Indices:  indices,
			QueryStr: queryDSLStr,
		}, nil

	default:
		return interfaces.ViewQuery4Metric{}, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_QueryType).
			WithErrorDetails("query type must be DSL or SQL or IndexBase")
	}
}

// 从 opensearch 获取索引和分片信息, GET _cat/indices/metricbeat*?v&h=index,pri&format=json
func (dvs *dataViewService) LoadIndexShards(ctx context.Context, indices string) ([]byte, int, error) {
	resBytes, status, err := dvs.osAccess.LoadIndexShards(ctx, indices)
	if err != nil {
		return nil, status, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_LoadIndexShardsFailed).WithErrorDetails(err.Error())
	}

	return resBytes, status, nil
}

func (dvs *dataViewService) DeleteDataViewPits(ctx context.Context, pits *interfaces.DeletePits) (*interfaces.DeletePitsResp, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Delete Data View Pits")
	defer span.End()

	deletePitIDs := pits.PitIDs
	// 如果 pit id 包含 __all，则删除所有 pit，重置 pit_ids，只包含一个值为 _all
	for _, pitID := range deletePitIDs {
		if pitID == interfaces.All_Pits_DataView {
			deletePitIDs = []string{interfaces.All_Pits_OpenSearch}
			break
		}
	}

	res, _, err := dvs.osAccess.DeletePointInTime(ctx, deletePitIDs)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_DataView_InternalError_DeletePointInTimeFailed).
			WithErrorDetails(fmt.Sprintf("delete pit ids %v failed, %s", deletePitIDs, err.Error()))
	}

	return res, nil
}

// 过滤视图下的行列规则，返回当前用户具有rule_apply权限的规则
func (dvs *dataViewService) FilterRowColumnRules(ctx context.Context, rules []*interfaces.DataViewRowColumnRule) ([]*interfaces.DataViewRowColumnRule, error) {
	if len(rules) == 0 {
		return rules, nil
	}

	// 根据权限过滤有查看权限的对象
	ruleIDs := make([]string, 0)
	for _, v := range rules {
		ruleIDs = append(ruleIDs, v.RuleID)
	}
	matchResouces, err := dvs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_DATA_VIEW_ROW_COLUMN_RULE,
		ruleIDs, []string{interfaces.OPERATION_TYPE_RULE_APPLY}, true)
	if err != nil {
		return nil, err
	}

	// 所有有权限的模型数组
	idMap := make(map[string]interfaces.ResourceOps)
	for _, resourceOps := range matchResouces {
		idMap[resourceOps.ResourceID] = resourceOps
	}

	// 遍历对象
	results := make([]*interfaces.DataViewRowColumnRule, 0)
	for _, v := range rules {
		if resrc, exist := idMap[v.RuleID]; exist {
			v.Operations = resrc.Operations // 用户当前有权限的操作

			results = append(results, v)
		}
	}

	return results, nil
}

// extractSearchAfterParams 从nextUri中提取searchAfter参数
// 解析nextUri的倒数三个部分，例如：/api/internal/vega-gateway/v2/fetch/{query_id}/{slug}/{token} 解析出 query_id、slug、token
func extractSearchAfterParams(rootNode ast.Node, isSingleDataSource bool) []any {
	var nextUri string
	if isSingleDataSource {
		nextUri, _ = rootNode.Get("next_uri").String()
	} else {
		nextUri, _ = rootNode.Get("nextUri").String()
	}

	if nextUri == "" {
		return nil
	}

	// 解析URI路径，获取最后三个部分
	uriParts := strings.Split(strings.Trim(nextUri, "/"), "/")
	if len(uriParts) < 3 {
		return nil
	}

	// 取倒数三个部分
	startIdx := len(uriParts) - 3
	if startIdx < 0 {
		startIdx = 0
	}
	searchAfterParams := uriParts[startIdx:]

	// 转换为any类型数组
	searchAfter := make([]any, 0, len(searchAfterParams))
	for _, param := range searchAfterParams {
		searchAfter = append(searchAfter, param)
	}

	return searchAfter
}

// parseSQLTotalCount 解析SQL查询结果中的total_count
func parseSQLTotalCount(rootNode ast.Node) (int64, error) {
	countNode := rootNode.Get("total_count")
	if !countNode.Exists() {
		return 0, errors.New("SQL total_count not exists in response")
	}

	total, err := countNode.Int64()
	if err != nil {
		return 0, fmt.Errorf("SQL total_count convert to int64 failed, %s", err.Error())
	}

	return total, nil
}

// parseSQLColumns 解析SQL查询结果中的columns信息
func parseSQLColumns(rootNode ast.Node) ([]ast.Node, error) {
	columnsNode := rootNode.Get("columns")
	if !columnsNode.Exists() {
		return nil, errors.New("SQL columnsNode not exists in response")
	}

	if !columnsNode.Valid() {
		return nil, errors.New("SQL columnsNode is invalid")
	}

	//判断返回的是否为 "columns": null, V_NULL   = 2 (json value `null`, key exists)
	if columnsNode.TypeSafe() == ast.V_NULL {
		return nil, errors.New("SQL columnsNode is null")
	}

	columns, err := columnsNode.ArrayUseNode()
	if err != nil {
		return nil, fmt.Errorf("SQL columnsNode convert to arrayNode failed, %s", err.Error())
	}

	return columns, nil
}

// parseSQLDocuments 解析SQL查询结果中的文档数据
func parseSQLDocuments(rootNode ast.Node, isSingleDataSource bool) ([]ast.Node, error) {
	var docsNode *ast.Node
	if isSingleDataSource {
		docsNode = rootNode.Get("entries")
	} else {
		docsNode = rootNode.Get("data")
	}

	if !docsNode.Exists() {
		return nil, errors.New("SQL dataNode not exists in response")
	}

	if !docsNode.Valid() {
		return nil, errors.New("SQL dataNode is invalid")
	}

	//判断返回的是否为 "data": null, V_NULL   = 2 (json value `null`, key exists)
	if docsNode.TypeSafe() == ast.V_NULL {
		return nil, errors.New("SQL dataNode is null")
	}

	docs, err := docsNode.ArrayUseNode()
	if err != nil {
		return nil, fmt.Errorf("SQL dataNode convert to arrayNode failed, %s", err.Error())
	}

	return docs, nil
}

// processViewFields 处理视图字段映射逻辑
func processViewFields(view *interfaces.DataView, columns []ast.Node, sqlStr string) error {
	// 如果是通过接口自己传的sql语句，大概率包含聚合、函数比较复杂的操作
	// 这时候不使用视图字段，直接返回vega返回的字段
	if sqlStr != "" {
		view.FieldScope = interfaces.FieldScope_All
	}

	// 预览情况下，sql node的字段列表需要字段列表
	if view.FieldScope == interfaces.FieldScope_All {
		return processAllFields(view, columns)
	} else if view.HasDataScopeSQLNode {
		return processSQLNodeFields(view, columns)
	}

	return nil
}

// processAllFields 处理所有字段的情况
func processAllFields(view *interfaces.DataView, columns []ast.Node) error {
	fieldsArr := make([]*cond.ViewField, 0)
	fieldsMap := make(map[string]*cond.ViewField)
	fieldCount := make(map[string]int)

	for _, col := range columns {
		fieldName, err := col.Get("name").String()
		if err != nil {
			return fmt.Errorf("SQL column name convert to string failed, %s", err.Error())
		}

		fieldType, err := col.Get("type").String()
		if err != nil {
			return fmt.Errorf("SQL column type convert to string failed, %s", err.Error())
		}

		// 获取原始字段名
		originalName := fieldName

		// 如果columns里有重复的字段名，为后面的字段设置别名
		if _, exist := fieldsMap[fieldName]; exist {
			fieldCount[fieldName]++
			// 为重复字段创建新的唯一名称（添加后缀）
			newFieldName := fmt.Sprintf("%s_%d", fieldName, fieldCount[fieldName])
			f := &cond.ViewField{
				Name:         newFieldName, // 唯一字段名
				DisplayName:  newFieldName,
				OriginalName: originalName, // 保存原始字段名
				Type:         fieldType,
			}
			fieldsMap[newFieldName] = f
			fieldsArr = append(fieldsArr, f)
		} else {
			fieldCount[fieldName] = 0
			f := &cond.ViewField{
				Name:         fieldName,
				DisplayName:  fieldName,
				OriginalName: originalName,
				Type:         fieldType,
			}
			fieldsMap[fieldName] = f
			fieldsArr = append(fieldsArr, f)
		}
	}

	view.Fields = fieldsArr
	view.FieldsMap = fieldsMap
	return nil
}

// processSQLNodeFields 处理SQL节点字段的情况
func processSQLNodeFields(view *interfaces.DataView, columns []ast.Node) error {
	fieldsArr := make([]*cond.ViewField, 0)
	fieldsMap := make(map[string]*cond.ViewField)

	for _, col := range columns {
		fieldName, err := col.Get("name").String()
		if err != nil {
			return fmt.Errorf("SQL column name convert to string failed, %s", err.Error())
		}
		fieldType, err := col.Get("type").String()
		if err != nil {
			return fmt.Errorf("SQL column type convert to string failed, %s", err.Error())
		}

		// 非*情况下，已解析出字段，补齐字段类型
		if f, ok := view.FieldsMap[fieldName]; ok {
			f.Type = fieldType
			fieldsMap[fieldName] = f
		}
	}

	for _, af := range fieldsMap {
		fieldsArr = append(fieldsArr, af)
	}

	view.Fields = fieldsArr
	view.FieldsMap = fieldsMap
	return nil
}

// 读取count结果
func readCountResult(ctx context.Context, isSingleDataSource bool, result []byte) (int64, error) {
	var val any
	if isSingleDataSource {
		var vegaFetchData interfaces.VegaGatewayProFetchDataRes
		d := decoder.NewDecoder(string(result))
		d.UseInt64()
		if err := d.Decode(&vegaFetchData); err != nil {
			logger.Errorf("unmalshal vega datas info failed: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Unmalshal vega datas failed: %v", err))

			return 0, fmt.Errorf("unmarshal vega datas info failed: %v", err)
		}

		// 总数
		val = vegaFetchData.Entries[0][0]
	} else {
		var vegaFetchData interfaces.DataConnFetchDataRes
		d := decoder.NewDecoder(string(result))
		d.UseInt64()
		if err := d.Decode(&vegaFetchData); err != nil {
			logger.Errorf("unmalshal vega datas info failed: %v", err)
			o11y.Error(ctx, fmt.Sprintf("Unmalshal vega datas failed: %v", err))

			return 0, fmt.Errorf("unmarshal vega datas info failed: %v", err)
		}

		// 总数
		val = vegaFetchData.Data[0][0]
	}

	var totalCountAll int64
	// 已知val实际上是int类型
	if tmp, ok := val.(int64); ok {
		totalCountAll = tmp
	} else {
		logger.Errorf("data.Data[0][0] is not a int64, conversion failed")
		return 0, fmt.Errorf("data count is not a int64, conversion failed")
	}

	return totalCountAll, nil
}
