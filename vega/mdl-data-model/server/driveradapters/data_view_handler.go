package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
)

func (r *restHandler) HandleDataViewPostOverrideByEx(c *gin.Context) {
	switch c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE) {
	case "", http.MethodPost:
		r.CreateDataViewsByEx(c)
	case http.MethodGet:
		r.GetDataViewsByEx(c)
	case http.MethodDelete:
		r.DeleteDataViewsByEx(c)
	default:
		httpErr := rest.NewHTTPError(rest.GetLanguageCtx(c), http.StatusBadRequest,
			derrors.DataModel_InvalidParameter_OverrideMethod)
		rest.ReplyError(c, httpErr)
	}
}

func (r *restHandler) HandleDataViewPostOverrideByIn(c *gin.Context) {
	switch c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE) {
	case "", http.MethodPost:
		r.CreateDataViewsByIn(c)
	case http.MethodGet:
		r.GetDataViewsByIn(c)
	case http.MethodDelete:
		r.DeleteDataViewsByIn(c)
	default:
		httpErr := rest.NewHTTPError(rest.GetLanguageCtx(c), http.StatusBadRequest,
			derrors.DataModel_InvalidParameter_OverrideMethod)
		rest.ReplyError(c, httpErr)
	}
}

// 创建数据视图（外部）
func (r *restHandler) CreateDataViewsByEx(c *gin.Context) {
	logger.Debug("Handler CreateDataViewsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Create data views by ex", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.CreateDataViews(c, visitor)
}

// 创建数据视图(内部)
func (r *restHandler) CreateDataViewsByIn(c *gin.Context) {
	logger.Debug("Handler CreateDataViewsByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.CreateDataViews(c, visitor)
}

// 创建数据视图
func (r *restHandler) CreateDataViews(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler CreateDataViews Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Create data views", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 查询参数
	mode := c.DefaultQuery(interfaces.QueryParam_ImportMode, interfaces.ImportMode_Normal)
	httpErr := validateImportMode(ctx, mode)
	if httpErr != nil {
		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject("", ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	viewsReqBody := []interfaces.CreateDataView{}
	err := c.ShouldBindJSON(&viewsReqBody)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
			WithErrorDetails("Binding paramter failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject("", ""), &httpErr.BaseError)

		// 记录错误log
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置attributes和status
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 如果传入的视图为[], 应报错
	if len(viewsReqBody) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
			WithErrorDetails("No data view was passed in")

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject("", ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	nameMap := make(map[string]any)
	// techNameMap := make(map[string]any)
	idMap := make(map[string]any)
	dataViews := make([]*interfaces.DataView, 0, len(viewsReqBody))
	for i := 0; i < len(viewsReqBody); i++ {
		// 校验导入视图时模块是否是数据视图
		if viewsReqBody[i].ModuleType != "" && viewsReqBody[i].ModuleType != interfaces.MODULE_TYPE_DATA_VIEW {
			httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_InvalidParameter_ModuleType).
				WithErrorDetails("Model type is not 'data_view'")

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataViewAuditObject("", ""), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		viewID := viewsReqBody[i].ViewID
		viewName := viewsReqBody[i].ViewName
		// techName := viewsReqBody[i].TechnicalName
		groupName := viewsReqBody[i].GroupName
		uk_view_name := fmt.Sprintf("%s_%s", groupName, viewName)
		// uk_technical_name := fmt.Sprintf("%s_%s", groupName, techName)

		// 校验请求体中多个视图 ID 是否重复
		if _, ok := idMap[viewID]; !ok {
			idMap[viewID] = nil
		} else {
			errDetails := fmt.Sprintf("DataView ID '%s' already exists in the file", viewID)
			httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataView_Duplicated_ViewIDInFile).
				WithDescription(map[string]any{"ViewID": viewID}).
				WithErrorDetails(errDetails)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataViewAuditObject(viewID, viewName), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		// 校验请求体中多个视图名称是否在分组内重复
		if _, ok := nameMap[uk_view_name]; !ok {
			nameMap[uk_view_name] = nil
		} else {
			errDetails := fmt.Sprintf("DataView '%s' already exists within the group '%s' in the file", viewName, groupName)
			httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataView_Duplicated_ViewNameInFile).
				WithDescription(map[string]any{"ViewName": viewName, "GroupName": groupName}).
				WithErrorDetails(errDetails)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataViewAuditObject(viewID, viewName), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		// 校验请求体中多个视图技术名称是否在分组内重复
		// if _, ok := techNameMap[uk_technical_name]; !ok {
		// 	techNameMap[uk_technical_name] = nil
		// } else {
		// 	errDetails := fmt.Sprintf("DataView technical name '%s' already exists within the group '%s' in the file", techName, groupName)
		// 	httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataView_Duplicated_ViewTechnicalNameInFile).
		// 		WithDescription(map[string]any{"TechnicalName": techName, "GroupName": groupName}).
		// 		WithErrorDetails(errDetails)

		// 	audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		// 		GenerateDataViewAuditObject(viewID, viewName), &httpErr.BaseError)

		// 	o11y.AddHttpAttrs4HttpError(span, httpErr)
		// 	rest.ReplyError(c, httpErr)
		// 	return
		// }

		dataView := &interfaces.DataView{
			SimpleDataView: interfaces.SimpleDataView{
				ViewID:        viewsReqBody[i].ViewID,
				ViewName:      viewsReqBody[i].ViewName,
				TechnicalName: viewsReqBody[i].TechnicalName,
				GroupID:       viewsReqBody[i].GroupID,
				GroupName:     viewsReqBody[i].GroupName,
				Type:          viewsReqBody[i].Type,
				QueryType:     strings.ToUpper(viewsReqBody[i].QueryType),
				Tags:          viewsReqBody[i].Tags,
				Comment:       viewsReqBody[i].Comment,
				DataSource:    viewsReqBody[i].DataSource, // 索引库创建视图还有这个字段，暂时保留，切换为扫描索引库时删除
				DataSourceID:  viewsReqBody[i].DataSourceID,
				FileName:      viewsReqBody[i].FileName,
			},
			ExcelConfig: viewsReqBody[i].ExcelConfig,
			DataScope:   viewsReqBody[i].DataScope,
			// FieldScope:      viewsReqBody[i].FieldScope,
			Fields:      viewsReqBody[i].Fields,
			PrimaryKeys: viewsReqBody[i].PrimaryKeys,
			// LogGroupFilters: viewsReqBody[i].LogGroupFilters,
		}

		// 校验 builtin, 统一返回 bool 类型
		builtinBool, err := validateBuiltin(ctx, viewsReqBody[i].Builtin)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataViewAuditObject(viewID, viewName), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		// 更新视图 builtin 属性
		dataView.Builtin = builtinBool

		// 校验视图过滤条件类型,统一返回新过滤器格式
		// cfg, err := validateViewFiltersType(ctx, viewsReqBody[i].Condition)
		// if err != nil {
		// 	httpErr := err.(*rest.HTTPError)

		// 	audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		// 		GenerateDataViewAuditObject(viewID, viewName), &httpErr.BaseError)

		// 	o11y.AddHttpAttrs4HttpError(span, httpErr)
		// 	rest.ReplyError(c, httpErr)
		// 	return
		// }

		// 更新视图过滤条件，能够兼容旧过滤条件并转成新的格式
		// dataView.Condition = cfg

		// 校验数据视图必要创建参数的合法性
		err = ValidateDataView(ctx, dataView)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataViewAuditObject(viewID, viewName), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		dataViews = append(dataViews, dataView)
	}

	// 批量创建
	viewIDs, err := r.dvs.CreateDataViews(ctx, dataViews, mode, true)
	if err != nil {
		// 失败，发送一条
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject(dataViews[0].ViewID, dataViews[0].ViewName), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 成功，发送多条
	for _, view := range dataViews {
		audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject(view.ViewID, view.ViewName), "")
	}

	// viewsIDsStr := strings.Join(viewIDs, ",")
	result := make([]interface{}, 0, len(viewIDs))
	for _, viewID := range viewIDs {
		result = append(result, map[string]interface{}{"id": viewID})
	}

	logger.Debug("Handler CreateDataViews Success")
	// c.Writer.Header().Set("Location", "/api/mdl-data-model/v1/data-views/"+viewsIDsStr)
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// 删除数据视图（外部）
func (r *restHandler) DeleteDataViewsByEx(c *gin.Context) {
	logger.Debug("Handler DeleteDataViewsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Delete data views by ex", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.DeleteDataViews(c, visitor)
}

// 删除数据视图(内部)
func (r *restHandler) DeleteDataViewsByIn(c *gin.Context) {
	logger.Debug("Handler DeleteDataViewsByIn Start")

	visitor := GenerateVisitor(c)
	r.DeleteDataViews(c, visitor)
}

// 删除数据视图，支持批量删除
func (r *restHandler) DeleteDataViews(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler DeleteDataViews Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Delete data views", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	viewIDsStr := c.Param("view_ids")
	viewIDs := common.StringToStringSlice(viewIDsStr)

	// 如果路径参数不存在，则为重载接口
	if len(viewIDs) == 0 {
		var viewIDsReq interfaces.ViewIDsReq
		err := c.ShouldBindJSON(&viewIDsReq)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
				WithErrorDetails("Failed to bind param: " + err.Error())

			audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				GenerateDataViewAuditObject("", ""), &httpErr.BaseError)

			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		viewIDs = viewIDsReq.IDs
	}

	span.SetAttributes(attr.Key("view_ids").String(fmt.Sprintf("%v", viewIDs)))

	// 校验视图是否存在，有一个不存在则返回错误
	viewNames := make(map[string]string)
	for _, viewID := range viewIDs {
		viewName, eixst, err := r.dvs.CheckDataViewExistByID(ctx, nil, viewID)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				GenerateDataViewAuditObject(viewID, viewName), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		if !eixst {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataView_DataViewNotFound).
				WithErrorDetails(fmt.Sprintf("Data view %s not exist", viewID))

			audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				GenerateDataViewAuditObject(viewID, viewName), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		viewNames[viewID] = viewName
	}

	err := r.dvs.DeleteDataViews(ctx, viewIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject("", ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 成功，发送多条
	for _, viewID := range viewIDs {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject(viewID, viewNames[viewID]), audit.SUCCESS, "")
	}

	logger.Debug("Handler DeleteDataViews Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 更新数据视图（外部）
func (r *restHandler) UpdateDataViewByEx(c *gin.Context) {
	logger.Debug("Handler UpdateDataViewByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Update a data view by ex", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.UpdateDataView(c, visitor)
}

// 更新数据视图(内部)
func (r *restHandler) UpdateDataViewByIn(c *gin.Context) {
	logger.Debug("Handler UpdateDataViewByIn Start")

	visitor := GenerateVisitor(c)
	r.UpdateDataView(c, visitor)
}

// 更新数据视图
func (r *restHandler) UpdateDataView(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler UpdateDataView Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Update a data view", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	viewID := c.Param("view_id")
	span.SetAttributes(attr.Key("view_id").String(viewID))

	viewInfo := &interfaces.DataView{}
	err := c.ShouldBindJSON(&viewInfo)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
			WithErrorDetails("Failed to bind param: " + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject(viewID, ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	viewInfo.ViewID = viewID

	// 参数校验
	err = ValidateDataView(ctx, viewInfo)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject(viewID, viewInfo.ViewName), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	err = r.dvs.UpdateDataView(ctx, nil, viewInfo)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject(viewID, viewInfo.ViewName), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		GenerateDataViewAuditObject(viewID, viewInfo.ViewName), "")

	logger.Debug("Handler updateDataView Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 按 id 获取数据视图对象信息（外部）
func (r *restHandler) GetDataViewsByEx(c *gin.Context) {
	logger.Debug("Handler GetDataViewsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Get data views by ex", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetDataViews(c, visitor)
}

// 按 id 获取数据视图对象信息(内部)
func (r *restHandler) GetDataViewsByIn(c *gin.Context) {
	logger.Debug("Handler GetDataViewsByIn Start")

	visitor := GenerateVisitor(c)
	r.GetDataViews(c, visitor)
}

// 获取数据视图详情
func (r *restHandler) GetDataViews(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler GetDataViews Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Get data views", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	viewIDsStr := c.Param("view_ids")
	viewIDs := common.StringToStringSlice(viewIDsStr)

	// 如果路径参数不存在，则为重载接口
	if len(viewIDs) == 0 {
		var viewIDsReq interfaces.ViewIDsReq
		err := c.ShouldBindJSON(&viewIDsReq)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
				WithErrorDetails("Failed to bind param: " + err.Error())

			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		viewIDs = viewIDsReq.IDs
	}

	includeDataScopeViewsParam := c.DefaultQuery("include_data_scope_views", "false")
	includeDataScopeView, err := strconv.ParseBool(includeDataScopeViewsParam)
	if err != nil {
		errDetails := fmt.Sprintf(`The value of param 'include_data_scope_views' should be bool type, but got '%s'`, includeDataScopeViewsParam)

		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails(errDetails)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	views, err := r.dvs.GetDataViews(ctx, viewIDs, includeDataScopeView)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler RetrieveDataView Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, views)
}

// 分页获取数据视图列表（外部）
func (r *restHandler) ListDataViewsByEx(c *gin.Context) {
	logger.Debug("Handler ListDataViewsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: List data views by ex", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ListDataViews(c, visitor)
}

// 分页获取数据视图列表(内部)
func (r *restHandler) ListDataViewsByIn(c *gin.Context) {
	logger.Debug("Handler ListDataViewsByIn Start")

	visitor := GenerateVisitor(c)
	r.ListDataViews(c, visitor)
}

// 获取数据视图列表
func (r *restHandler) ListDataViews(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler ListDataViews Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: List data views", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	name := c.Query("name")
	namePattern := c.Query("name_pattern")
	techName := c.Query("technical_name")
	keyword := c.Query("keyword")
	tag := c.Query("tag")
	dataSourceType := c.Query("data_source_type")
	dataSourceID := c.Query("data_source_id")
	fileName := c.Query("file_name")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", "update_time")
	direction := c.DefaultQuery("direction", interfaces.DESC_DIRECTION)
	builtinQueryArr := c.QueryArray("builtin")
	viewType := c.Query("type")
	queryType := c.Query("query_type")
	statusArr := c.QueryArray("status")
	createTimeStartStr := c.Query("create_time_start")
	createTimeEndStr := c.Query("create_time_end")
	updateTimeStartStr := c.Query("update_time_start")
	updateTimeEndStr := c.Query("update_time_end")
	groupID := c.DefaultQuery("group_id", interfaces.GroupID_All)
	groupName := c.DefaultQuery("group_name", interfaces.GroupName_All)
	operationsArr := c.QueryArray("operations")

	// 校验builtin参数
	builtinArr := make([]bool, 0, len(builtinQueryArr))
	for _, val := range builtinQueryArr {
		if val == "" {
			continue
		}
		builtin, err := strconv.ParseBool(val)
		if err != nil {
			errDetails := fmt.Sprintf(`The value of param 'builtin' should be bool type, but got '%s'`, val)
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_Builtin).
				WithErrorDetails(errDetails)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		builtinArr = append(builtinArr, builtin)
	}

	createTimeStart, _ := common.QuotedTimestampToInt64(createTimeStartStr)
	createTimeEnd, _ := common.QuotedTimestampToInt64(createTimeEndStr)
	updateTimeStart, _ := common.QuotedTimestampToInt64(updateTimeStartStr)
	updateTimeEnd, _ := common.QuotedTimestampToInt64(updateTimeEndStr)

	err := validateNameandNamePattern(ctx, name, namePattern)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 分页参数校验
	PageParam, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.DATA_VIEW_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验field_scope参数
	// if fieldScopeStr != "" && (fieldScopeStr != fmt.Sprintf("%d", interfaces.FieldScope_Custom) &&
	// 	fieldScopeStr != fmt.Sprintf("%d", interfaces.FieldScope_All)) {

	// 	httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_FieldScope).
	// 		WithErrorDetails("The value of param 'field_scope' should be one of the following options: '0', '1'")

	// 	o11y.AddHttpAttrs4HttpError(span, httpErr)
	// 	rest.ReplyError(c, httpErr)
	// 	return
	// }

	//去掉标签前后的所有空格进行搜索
	tag = strings.Trim(tag, " ")

	viewQueryParam := &interfaces.ListViewQueryParams{
		Type:                      viewType,
		QueryType:                 queryType,
		DataSourceType:            dataSourceType,
		DataSourceID:              dataSourceID,
		FileName:                  fileName,
		Keyword:                   keyword,
		Name:                      name,
		NamePattern:               namePattern,
		TechnicalName:             techName,
		GroupID:                   groupID,
		GroupName:                 groupName,
		Status:                    statusArr,
		CreateTimeStart:           createTimeStart,
		CreateTimeEnd:             createTimeEnd,
		UpdateTimeStart:           updateTimeStart,
		UpdateTimeEnd:             updateTimeEnd,
		Tag:                       tag,
		Builtin:                   builtinArr,
		Operations:                operationsArr,
		PaginationQueryParameters: PageParam,
		// OpenStreaming:             openStreamingArr,
		// FieldScopeStr:             fieldScopeStr,
	}

	views, total, err := r.dvs.ListDataViews(ctx, viewQueryParam)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]interface{}{
		"entries":     views,
		"total_count": total,
	}

	logger.Debug("Handler ListDataViews Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// 更新视图的属性字段
func (r *restHandler) UpdateDataViewAttrFields(c *gin.Context) {
	fieldsStr := c.Param("fields")

	// 空字符串split之后的长度为 1，所以先判断 fieldsStr 是否为空字符串
	fieldsArr := common.StringToStringSlice(fieldsStr)
	if len(fieldsArr) == 0 {
		httpErr := rest.NewHTTPError(rest.GetLanguageCtx(c), http.StatusBadRequest,
			derrors.DataModel_DataView_NullParameter_AttributeFields).
			WithErrorDetails("No attribute field is passed")

		rest.ReplyError(c, httpErr)
		return
	}

	for _, field := range fieldsArr {
		// 目前只支持group_name字段
		if _, ok := interfaces.AttrFieldsMap[field]; !ok {
			httpErr := rest.NewHTTPError(rest.GetLanguageCtx(c), http.StatusBadRequest,
				derrors.DataModel_DataView_InvalidParameter_AttributeFields).
				WithErrorDetails(fmt.Sprintf("field '%s' is not supported", field))

			rest.ReplyError(c, httpErr)
			return
		}
	}

	// if len(fieldsArr) != 1 {
	// 	httpErr := rest.NewHTTPError(rest.GetLanguageCtx(c), http.StatusBadRequest,
	// 		derrors.DataModel_DataView_InvalidParameter_AttributeFields).
	// 		WithErrorDetails("Modifying multiple parameters simultaneously is not supported")

	// 	rest.ReplyError(c, httpErr)
	// 	return
	// }

	if fieldsArr[0] == interfaces.AttrFields_GroupName {
		r.UpdateDataViewsGroupName(c)
	} else {
		r.UpdateAtomicDataView(c)
	}

	// if fieldsArr[0] == interfaces.AttrFields_OpenStreaming {
	// 	r.UpdateDataViewRealTimeStreaming(c)
	// }
}

// 更新原子视图，暂时不支持批量
func (r *restHandler) UpdateAtomicDataView(c *gin.Context) {
	logger.Debug("Handler UpdateAtomicDataViews Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: update atomic data view", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	viewIDsStr := c.Param("view_id")
	span.SetAttributes(attr.Key("view_id").String(viewIDsStr))
	viewIDs := common.StringToStringSlice(viewIDsStr)

	if len(viewIDs) != 1 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_ViewID).
			WithErrorDetails("Batch updates are not supported")

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject(viewIDsStr, ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	viewUpdateReq := interfaces.AtomicViewUpdateReq{}
	err = c.ShouldBindJSON(&viewUpdateReq)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
			WithErrorDetails("Failed to bind param: " + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject(viewIDsStr, ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验原子视图的更新参数
	err = validateAtomicViewUpdateReq(ctx, &viewUpdateReq)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject(viewIDs[0], viewUpdateReq.ViewName), &httpErr.BaseError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	viewUpdateReq.ViewID = viewIDs[0]
	err = r.dvs.UpdateAtomicDataViews(ctx, &viewUpdateReq)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject(viewIDsStr, ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 成功，发送多条
	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		GenerateDataViewAuditObject(viewIDs[0], ""), "")

	logger.Debug("Handler UpdateAtomicDataViews Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 批量移动视图的分组
func (r *restHandler) UpdateDataViewsGroupName(c *gin.Context) {
	logger.Debug("Handler UpdateDataViewsGroupName Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Bulk move the grouping of data view", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	viewIDsStr := c.Param("view_id")
	span.SetAttributes(attr.Key("view_id").String(viewIDsStr))

	viewIDs := common.StringToStringSlice(viewIDsStr)

	viewGroupReq := interfaces.ViewGroupReq{}
	err = c.ShouldBindJSON(&viewGroupReq)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
			WithErrorDetails("Failed to bind param: " + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject(viewIDsStr, ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 批量移动视图分组，只能是自定义视图的分组，因此分组builtin为false
	viewGroupReq.Builtin = false

	// 判断数据视图是否存在
	viewMap, err := r.dvs.GetSimpleDataViewsByIDs(ctx, viewIDs, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject(viewIDsStr, ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 判断批量移动的视图的名称是否存在重复
	viewNameSet := make(map[string]struct{})
	for _, simpleView := range viewMap {
		if _, ok := viewNameSet[simpleView.ViewName]; ok {
			errDetails := fmt.Sprintf("Data view name '%s' is duplicated in group '%s'", simpleView.ViewName, viewGroupReq.GroupName)
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_Duplicated_ViewName).
				WithDescription(map[string]any{"ViewName": simpleView.ViewName, "GroupName": viewGroupReq.GroupName}).
				WithErrorDetails(errDetails)

			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateDataViewAuditObject(viewIDsStr, ""), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		viewNameSet[simpleView.ViewName] = struct{}{}
	}

	// 校验分组名称
	err = validateGroupName(ctx, viewGroupReq.GroupName)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject(viewIDsStr, ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	err = r.dvs.UpdateDataViewsGroup(ctx, viewMap, &viewGroupReq)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject(viewIDsStr, ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 成功，发送多条
	for _, view := range viewMap {
		audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewAuditObject(view.ViewID, view.ViewName), "")
	}

	logger.Debug("Handler updateDataViewsGroup Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 分页获取数据视图资源列表
func (r *restHandler) ListDataViewSrcs(c *gin.Context) {
	logger.Debug("ListDataViewSrcs Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: List data view sources", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))
	o11y.Info(ctx, fmt.Sprintf("driver layer: List data view sources request parameters: [%s]", c.Request.RequestURI))

	// 获取分页参数
	namePattern := c.Query(RESOURCES_KEYWOED)
	name := c.Query("name")
	tag := c.Query("tag")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", RESOURCES_PAGE_LIMIT)
	sort := c.DefaultQuery("sort", "group_name")
	direction := c.DefaultQuery("direction", interfaces.ASC_DIRECTION)
	builtinQueryArr := c.QueryArray("builtin")
	groupID := c.DefaultQuery("group_id", interfaces.GroupID_All)
	groupName := c.DefaultQuery("group_name", interfaces.GroupName_All)

	// 校验builtin参数
	builtinArr := make([]bool, 0, len(builtinQueryArr))
	for _, val := range builtinQueryArr {
		if val == "" {
			continue
		}
		builtin, err := strconv.ParseBool(val)
		if err != nil {
			errDetails := fmt.Sprintf(`The value of param 'builtin' should be bool type, but got '%s'`, val)
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_InvalidParameter_Builtin).
				WithErrorDetails(errDetails)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		builtinArr = append(builtinArr, builtin)
	}

	err = validateNameandNamePattern(ctx, name, namePattern)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 分页参数校验
	PageParam, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.DATA_VIEW_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//去掉标签前后的所有空格进行搜索
	tag = strings.Trim(tag, " ")

	viewQueryParam := &interfaces.ListViewQueryParams{
		Builtin:                   builtinArr,
		Name:                      name,
		NamePattern:               namePattern,
		GroupID:                   groupID,
		GroupName:                 groupName,
		Tag:                       tag,
		PaginationQueryParameters: PageParam,
	}

	resources, total, err := r.dvs.ListDataViewSrcs(ctx, viewQueryParam)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	result := map[string]interface{}{"entries": resources, "total_count": total}

	logger.Debug("Handler ListDataViewSrcs Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}
