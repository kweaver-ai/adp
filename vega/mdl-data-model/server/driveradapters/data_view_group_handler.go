package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	derrors "data-model/errors"
	"data-model/interfaces"
)

// 创建数据视图分组
func (r *restHandler) CreateDataViewGroup(c *gin.Context) {
	logger.Debug("Handler CreateDataViewGroup Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Create data view group", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	group := interfaces.DataViewGroup{}
	err = c.ShouldBindJSON(&group)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataViewGroup_InvalidParameter_RequestBody).
			WithErrorDetails("Binding paramter failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataViewGroupAuditObject("", ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	err = validateGroupName(ctx, group.GroupName)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataViewGroupAuditObject("", group.GroupName), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验分组名称是否已存在
	_, exist, err := r.dvgs.CheckDataViewGroupExistByName(ctx, nil, group.GroupName, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataViewGroupAuditObject("", group.GroupName), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataViewGroup_Existed_GroupName)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataViewGroupAuditObject("", group.GroupName), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	groupID, err := r.dvgs.CreateDataViewGroup(ctx, nil, &group)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataViewGroupAuditObject("", group.GroupName), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		GenerateDataViewGroupAuditObject("", group.GroupName), "")

	logger.Debug("Handler CreateDataViewGroup Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)

	result := map[string]string{"id": groupID}
	// c.Writer.Header().Set("Location", "/api/mdl-data-model/v1/data-view-groups/"+groupID)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// 刪除数据视图分组
func (r *restHandler) DeleteDataViewGroup(c *gin.Context) {
	logger.Debug("Handler DeleteDataViewGroup Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Delete data view group", trace.WithSpanKind(trace.SpanKindServer))
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

	groupID := c.Param("group_id")
	span.SetAttributes(attr.Key("group_id").String(groupID))

	if groupID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataViewGroup_NullParameter_GroupID).
			WithErrorDetails("Group ID cannot be empty")

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataViewGroupAuditObject("", ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 是否删除分组下的数据视图
	deleteViewsStr := c.DefaultQuery("delete_views", interfaces.DEFAULT_FORCE)
	deleteViews, err := strconv.ParseBool(deleteViewsStr)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataViewGroup_InvalidParameter_DeleteViews).
			WithErrorDetails(fmt.Sprintf("Invalid param delete_views '%s'", deleteViewsStr))

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataViewGroupAuditObject(groupID, ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 检查 groupID 是否存在
	group, err := r.dvgs.GetDataViewGroupByID(ctx, groupID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataViewGroupAuditObject(groupID, ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	dataViews, err := r.dvgs.DeleteDataViewGroup(ctx, groupID, deleteViews)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataViewGroupAuditObject(groupID, group.GroupName), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 记录删除分组成功的审计日志
	audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
		GenerateDataViewGroupAuditObject(groupID, group.GroupName), audit.SUCCESS, "")

	// 记录删除数据视图成功的审计日志
	for _, view := range dataViews {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataViewGroupAuditObject(view.ViewID, view.ViewName), audit.SUCCESS, "")
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 更新数据视图分组
func (r *restHandler) UpdateDataViewGroup(c *gin.Context) {
	logger.Debug("Handler UpdateDataViewGroup Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Update data view group", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	groupID := c.Param("group_id")
	span.SetAttributes(attr.Key("group_id").String(groupID))

	// groupID 不能为空
	if groupID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataViewGroup_NullParameter_GroupID).
			WithErrorDetails("The group ID cannot be empty")

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewGroupAuditObject("", ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	group := interfaces.DataViewGroup{}
	err = c.ShouldBindJSON(&group)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataViewGroup_InvalidParameter_RequestBody).
			WithErrorDetails("Binding paramter failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewGroupAuditObject(groupID, ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	group.GroupID = groupID

	// 判断数据视图分组是否存在
	oldGroup, err := r.dvgs.GetDataViewGroupByID(ctx, groupID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewGroupAuditObject(groupID, group.GroupName), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	err = validateGroupName(ctx, group.GroupName)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewGroupAuditObject(groupID, group.GroupName), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 名称不同需要校验名称是否已存在
	if oldGroup.GroupName != group.GroupName {
		_, exist, err := r.dvgs.CheckDataViewGroupExistByName(ctx, nil, group.GroupName, false)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateDataViewGroupAuditObject(groupID, group.GroupName), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		if exist {
			errDetails := fmt.Sprintf("Data view group '%s' already exists", group.GroupName)
			httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataViewGroup_Existed_GroupName).
				WithErrorDetails(errDetails)

			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateDataViewGroupAuditObject(groupID, group.GroupName), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	err = r.dvgs.UpdateDataViewGroup(ctx, &group)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataViewGroupAuditObject(groupID, group.GroupName), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		GenerateDataViewGroupAuditObject(groupID, group.GroupName), "")

	logger.Debug("Handler UpdateDataViewGroup Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 查询数据视图分组列表
func (r *restHandler) ListDataViewGroups(c *gin.Context) {
	logger.Debug("ListDataViewGroups Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: List data view groups", trace.WithSpanKind(trace.SpanKindServer))
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

	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", interfaces.DEFAULT_DATA_VIEW_GROUP_SORT)
	direction := c.DefaultQuery("direction", interfaces.ASC_DIRECTION)
	builtinQueryArr := c.QueryArray("builtin")
	includeDeletedStr := c.DefaultQuery("include_deleted", "false")

	pageParam, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.DATA_VIEW_GROUP_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验builtin参数
	builtinArr := make([]bool, 0, len(builtinQueryArr))
	for _, val := range builtinQueryArr {
		if val == "" {
			continue
		}
		builtin, err := strconv.ParseBool(val)
		if err != nil {
			errDetails := fmt.Sprintf(`The value of param 'builtin' should be bool type, but got '%s'`, val)
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
				derrors.DataModel_DataViewGroup_InvalidParameter_Builtin).
				WithErrorDetails(errDetails)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		builtinArr = append(builtinArr, builtin)
	}

	includeDeleted, err := strconv.ParseBool(includeDeletedStr)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails(fmt.Sprintf("Invalid param include_deleted '%s'", includeDeletedStr))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	queryParam := &interfaces.ListViewGroupQueryParams{
		Builtin:                   builtinArr,
		IncludeDeleted:            includeDeleted,
		PaginationQueryParameters: pageParam,
	}

	entries, total, err := r.dvgs.ListDataViewGroups(ctx, queryParam, true)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]interface{}{
		"total_count": total,
		"entries":     entries,
	}

	logger.Debug("Handler ListDataViewGroups Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// 导出分组下的视图
func (r *restHandler) GetDataViewsInGroup(c *gin.Context) {
	logger.Debug("Handler GetDataViewsInGroup Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Get data views", trace.WithSpanKind(trace.SpanKindServer))
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

	groupID := c.Param("group_id")
	span.SetAttributes(attr.Key("group_id").String(groupID))

	// 检查分组是否存在
	_, err = r.dvgs.GetDataViewGroupByID(ctx, groupID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	views, err := r.dvs.GetDataViewsByGroupID(ctx, groupID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler GetDataViewsInGroup Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, views)
}
