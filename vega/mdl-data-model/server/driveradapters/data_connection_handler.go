// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/singleflight"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
)

// 创建数据连接
func (r *restHandler) CreateDataConnection(c *gin.Context) {
	logger.Debug("Handler CreateDataConnection Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 创建数据连接", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler CreateDataConnection End")
	}()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. 接收参数
	reqConn := interfaces.DataConnection{}
	err = c.ShouldBindJSON(&reqConn)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_InvalidParameter_RequestBody).
			WithErrorDetails("Binding paramter failed: " + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataConnectionAuditObject("", ""), &httpErr.BaseError)

		// 记录错误log
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 2. 校验数据连接, 检查参数合法性
	err = validateDataConnectionWhenCreate(ctx, r, &reqConn)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataConnectionAuditObject("", reqConn.Name), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 3. 调用logic层, 创建数据连接
	connID, err := r.dcs.CreateDataConnection(ctx, &reqConn)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataConnectionAuditObject("", reqConn.Name), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 4. 构造创建/导入成功的返回结果, 并记录审计日志
	// 因为分布式ID是18位整数, 前端js不能准确接收, 所以返回的链路模型id要转为字符串形式,
	result := map[string]string{"id": connID}

	audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		GenerateDataConnectionAuditObject(connID, reqConn.Name), "")

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	// c.Writer.Header().Set("Location", "/api/mdl-data-model/v1/data-connections/"+connID)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// 批量删除数据连接
func (r *restHandler) DeleteDataConnections(c *gin.Context) {
	logger.Debug("Handler DeleteDataConnections Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 批量删除数据连接", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler DeleteDataConnections End")
	}()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. 获取查询参数
	connIDsStr := c.Param("conn_ids")

	// 2. 将ID字符串转换为[]string
	connIDs := common.StringToStringSlice(connIDsStr)
	// 如果未传入数据连接ID, 应报错
	if len(connIDs) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataConnection_InvalidParameter_ConnectionIDs).
			WithErrorDetails("No invalid data connection id was passed in")

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataConnectionAuditObject("", ""), &httpErr.BaseError)

		// 记录错误log
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 3. 根据数据连接ID去校验数据连接存在性
	id2Name, err := r.dcs.GetMapAboutID2Name(ctx, connIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataConnectionAuditObject(connIDsStr, ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	notExistIDs := make([]string, 0)
	for _, connID := range connIDs {
		if _, ok := id2Name[connID]; !ok {
			notExistIDs = append(notExistIDs, connID)
		}
	}

	if len(notExistIDs) > 0 {
		errDetails := fmt.Sprintf("The data connection whose id is in %v does not exist!", notExistIDs)
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound,
			derrors.DataModel_DataConnection_DataConnectionNotFound).
			WithErrorDetails(errDetails)

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataConnectionAuditObject(connIDsStr, ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 4. 批量删除数据连接
	err = r.dcs.DeleteDataConnections(ctx, connIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataConnectionAuditObject(connIDsStr, ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}

	// 5. 循环记录删除成功的审计日志
	for id, name := range id2Name {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataConnectionAuditObject(id, name), audit.SUCCESS, "")
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 修改数据连接
func (r *restHandler) UpdateDataConnection(c *gin.Context) {
	logger.Debug("Handler UpdateDataConnection Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 修改数据连接", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler UpdateDataConnection End")
	}()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. 获取查询参数
	connID := c.Param("conn_id")

	// 2. 将ID字符串转换为 string
	if connID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataConnection_InvalidParameter_ConnectionIDs).
			WithErrorDetails("No invalid data connection id was passed in")

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataConnectionAuditObject("", ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 上锁, 保证修改和查询过程互斥
	common.GLock.Lock(connID)
	defer common.GLock.Unlock(connID)

	// 3. 根据数据连接ID去校验数据连接存在性
	preConn, isExist, err := r.dcs.GetDataConnection(ctx, connID, true)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataConnectionAuditObject(connID, ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if !isExist {
		errDetails := fmt.Sprintf("The data connection whose id equal to %v was not found", connID)
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound,
			derrors.DataModel_DataConnection_DataConnectionNotFound).
			WithErrorDetails(errDetails)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataConnectionAuditObject(connID, ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 4. 接收request body
	reqConn := interfaces.DataConnection{}
	err = c.ShouldBindJSON(&reqConn)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
			WithErrorDetails("Binding paramter failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataConnectionAuditObject(connID, ""), &httpErr.BaseError)

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 5. 校验请求体
	err = validateDataConnectionWhenUpdate(ctx, r, &reqConn, preConn)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataConnectionAuditObject(connID, reqConn.Name), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 6. 设置reqConn的ID
	reqConn.ID = connID

	// 7. 调用logic层修改数据连接
	err = r.dcs.UpdateDataConnection(ctx, &reqConn, preConn)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataConnectionAuditObject(connID, reqConn.Name), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 8. 记录修改成功的审计日志, 并返回结果
	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		GenerateDataConnectionAuditObject(connID, reqConn.Name), "")

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 查询数据连接
func (r *restHandler) GetDataConnection(c *gin.Context) {
	logger.Debug("Handler GetDataConnection Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 查询数据连接", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler GetDataConnection End")
	}()

	_, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. 获取查询参数
	connID := c.Param("conn_id")
	withAuthInfoStr := c.DefaultQuery("with_auth_info", "false")

	if connID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataConnection_InvalidParameter_ConnectionIDs).
			WithErrorDetails("No invalid data connection id was passed in")
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	sf := singleflight.Group{}
	conn, err, _ := sf.Do(connID+withAuthInfoStr, func() (interface{}, error) {
		// 2. 处理查询参数
		// 2.2 将withAuthInfoStr转为bool类型
		withAuthInfo := (withAuthInfoStr == "true")

		// 上锁, 保证查询和修改过程互斥
		common.GLock.Lock(connID)
		defer common.GLock.Unlock(connID)

		// 3. 查询数据连接
		conn, isExist, err := r.dcs.GetDataConnection(ctx, connID, withAuthInfo)
		if err != nil {
			return conn, err
		}

		if !isExist {
			errDetails := fmt.Sprintf("The data connection whose id equal to %v was not found", connID)
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound,
				derrors.DataModel_DataConnection_DataConnectionNotFound).
				WithErrorDetails(errDetails)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
			return conn, httpErr
		}

		return conn, nil
	})

	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, conn)
}

// 查询数据连接列表
func (r *restHandler) ListDataConnections(c *gin.Context) {
	logger.Debug("Handler ListDataConnections Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver层: 查询数据连接列表与总数", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
		logger.Debug("Handler ListDataConnections End")
	}()

	_, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. 获取查询参数
	namePattern := c.Query("name_pattern")
	name := c.Query("name")
	tag := c.Query("tag")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", interfaces.DEFAULT_SORT)
	direction := c.DefaultQuery("direction", interfaces.DEFAULT_DIRECTION)
	applicationScopeStr := strings.Split(c.Query("application_scope"), ",")

	// 2. 处理applicationScope, 去除每个元素左右空格
	applicationScope := make([]string, 0)
	for _, obj := range applicationScopeStr {
		v := strings.TrimSpace(obj)
		if v != "" {
			applicationScope = append(applicationScope, v)
		}
	}

	// 3. 校验name_pattern和name
	err = validateNameandNamePattern(ctx, name, namePattern)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}

	// 4. 校验分页查询参数
	pageParams, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.DATA_CONNECTION_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}

	// 5. 构造链路模型列表查询参数的结构体
	para := interfaces.DataConnectionListQueryParams{
		CommonListQueryParams: interfaces.CommonListQueryParams{
			NamePattern: namePattern,
			Name:        name,
			Tag:         tag,
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Offset:    pageParams.Offset,
				Limit:     pageParams.Limit,
				Sort:      pageParams.Sort,
				Direction: pageParams.Direction,
			},
		},
		ApplicationScope: applicationScope,
	}

	// 6. 获取数据连接列表
	list, total, err := r.dcs.ListDataConnections(ctx, para)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}

	// 7. 构造返回结果
	result := map[string]interface{}{
		"entries":     list,
		"total_count": total,
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}
