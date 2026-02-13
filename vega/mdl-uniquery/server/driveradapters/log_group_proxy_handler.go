// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"

	uerrors "uniquery/errors"
	"uniquery/interfaces"
)

func (r *restHandler) ProxyGetLogGroupRoots(c *gin.Context) {
	logger.Debug("Handler ProxyGetLogGroupRoots Start")

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver layer: ProxyGetLogGroupRoots", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	var conn *interfaces.DataConnection = nil
	connID := c.GetHeader("X-Data-Connection-ID")
	if connID != "" {
		var err error
		var exist bool
		conn, exist, err = r.dcService.GetDataConnectionByID(ctx, connID)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_LogGroup_InternalError).
				WithErrorDetails(err.Error())
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))
			rest.ReplyError(c, httpErr)
			return
		}
		if !exist {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_LogGroup_DataConnectionNotFound).
				WithErrorDetails("data connection is not exist")
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))
			rest.ReplyError(c, httpErr)
			return
		}
	}

	data, err := r.lgService.GetLogGroupRootsByConn(ctx, conn)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_LogGroup_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, data)
}

func (r *restHandler) ProxyGetLogGroupTree(c *gin.Context) {
	logger.Debug("Handler ProxyGetLogGroupTree Start")

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver layer: ProxyGetLogGroupTree", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	userID := c.Param("id")
	if userID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_LogGroup_InvalidParameter_UserID).
			WithErrorDetails("user id is empty")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	var conn *interfaces.DataConnection = nil
	connID := c.GetHeader("X-Data-Connection-ID")
	if connID != "" {
		userID = interfaces.AR_ADMIN_ID

		var err error
		var exist bool
		conn, exist, err = r.dcService.GetDataConnectionByID(ctx, connID)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_LogGroup_InternalError).
				WithErrorDetails(err.Error())
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))
			rest.ReplyError(c, httpErr)
			return
		}
		if !exist {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_LogGroup_DataConnectionNotFound).
				WithErrorDetails("data connection is not exist")
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))
			rest.ReplyError(c, httpErr)
			return
		}
	}

	data, err := r.lgService.GetLogGroupTreeByConn(ctx, userID, conn)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_LogGroup_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, data)
}

func (r *restHandler) ProxyGetLogGroupChildren(c *gin.Context) {
	logger.Debug("Handler ProxyGetLogGroupChildren Start")

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver layer: ProxyGetLogGroupChildren", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	logGroupID := c.Param("id")
	if logGroupID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_LogGroup_InvalidParameter_LogGroupID).
			WithErrorDetails("log group id is empty")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	var conn *interfaces.DataConnection = nil
	connID := c.GetHeader("X-Data-Connection-ID")
	if connID != "" {
		var err error
		var exist bool
		conn, exist, err = r.dcService.GetDataConnectionByID(ctx, connID)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_LogGroup_InternalError).
				WithErrorDetails(err.Error())
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))
			rest.ReplyError(c, httpErr)
			return
		}
		if !exist {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_LogGroup_DataConnectionNotFound).
				WithErrorDetails("data connection is not exist")
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))
			rest.ReplyError(c, httpErr)
			return
		}
	}

	data, err := r.lgService.GetLogGroupChildrenByConn(ctx, logGroupID, conn)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_LogGroup_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, data)
}

func (r *restHandler) ProxySearchSubmit(c *gin.Context) {
	logger.Debug("Handler ProxySearchSubmit Start")

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver layer: ProxySearchSubmit", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	userID := c.GetHeader("User")

	var conn *interfaces.DataConnection = nil
	connID := c.GetHeader("X-Data-Connection-ID")
	if connID != "" {
		userID = interfaces.AR_ADMIN_ID

		var err error
		var exist bool
		conn, exist, err = r.dcService.GetDataConnectionByID(ctx, connID)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_LogGroup_InternalError).
				WithErrorDetails(err.Error())
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))
			rest.ReplyError(c, httpErr)
			return
		}
		if !exist {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_LogGroup_DataConnectionNotFound).
				WithErrorDetails("data connection is not exist")
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))
			rest.ReplyError(c, httpErr)
			return
		}
	}

	queryBody := []any{}
	err := c.ShouldBindJSON(&queryBody)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_RequestBody).
			WithErrorDetails("Binding Parameter Failed:" + err.Error())

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	data, err := r.lgService.SearchSubmitByConn(ctx, queryBody, userID, conn)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_LogGroup_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, data)
}

func (r *restHandler) ProxySearchFetch(c *gin.Context) {
	logger.Debug("Handler ProxySearchFetch Start")

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver layer: ProxySearchFetch", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	jobID := c.Param("job_id")
	if jobID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_LogGroup_InvalidParameter_JobID).
			WithErrorDetails("job id is empty")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	var conn *interfaces.DataConnection = nil
	connID := c.GetHeader("X-Data-Connection-ID")
	if connID != "" {
		var err error
		var exist bool
		conn, exist, err = r.dcService.GetDataConnectionByID(ctx, connID)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_LogGroup_InternalError).
				WithErrorDetails(err.Error())
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))
			rest.ReplyError(c, httpErr)
			return
		}
		if !exist {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_LogGroup_DataConnectionNotFound).
				WithErrorDetails("data connection is not exist")
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))
			rest.ReplyError(c, httpErr)
			return
		}
	}

	queryParams := url.Values{}
	if value, ok := c.GetQuery("page"); ok {
		queryParams.Add("page", value)
	}
	if value, ok := c.GetQuery("size"); ok {
		queryParams.Add("size", value)
	}

	data, err := r.lgService.SearchFetchByConn(ctx, jobID, queryParams, conn)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_LogGroup_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, data)
}

func (r *restHandler) ProxySearchFetchFields(c *gin.Context) {
	logger.Debug("Handler ProxySearchFetchFields Start")

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver layer: ProxySearchFetchFields", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	jobID := c.Param("job_id")
	if jobID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_LogGroup_InvalidParameter_JobID).
			WithErrorDetails("job id is empty")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	var conn *interfaces.DataConnection = nil
	connID := c.GetHeader("X-Data-Connection-ID")
	if connID != "" {
		var err error
		var exist bool
		conn, exist, err = r.dcService.GetDataConnectionByID(ctx, connID)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_LogGroup_InternalError).
				WithErrorDetails(err.Error())
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))
			rest.ReplyError(c, httpErr)
			return
		}
		if !exist {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_LogGroup_DataConnectionNotFound).
				WithErrorDetails("data connection is not exist")
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))
			rest.ReplyError(c, httpErr)
			return
		}
	}

	queryParams := url.Values{}
	if value, ok := c.GetQuery("field_name"); ok {
		queryParams.Add("field_name", value)
	}
	if value, ok := c.GetQuery("field_type"); ok {
		queryParams.Add("field_type", value)
	}
	if value, ok := c.GetQuery("logLibrary"); ok {
		queryParams.Add("logLibrary", value)
	}
	if value, ok := c.GetQuery("aggNum"); ok {
		queryParams.Add("aggNum", value)
	}
	if value, ok := c.GetQuery("type_list"); ok {
		queryParams.Add("type_list", value)
	}

	data, err := r.lgService.SearchFetchFieldsByConn(ctx, jobID, queryParams, conn)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_LogGroup_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, data)
}

func (r *restHandler) ProxySearchFetchSameFields(c *gin.Context) {
	logger.Debug("Handler ProxySearchFetchSameFields Start")

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver layer: ProxySearchFetchSameFields", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	jobID := c.Param("job_id")
	if jobID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_LogGroup_InvalidParameter_JobID).
			WithErrorDetails("job id is empty")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	var conn *interfaces.DataConnection = nil
	connID := c.GetHeader("X-Data-Connection-ID")
	if connID != "" {
		var err error
		var exist bool
		conn, exist, err = r.dcService.GetDataConnectionByID(ctx, connID)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_LogGroup_InternalError).
				WithErrorDetails(err.Error())
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))
			rest.ReplyError(c, httpErr)
			return
		}
		if !exist {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_LogGroup_DataConnectionNotFound).
				WithErrorDetails("data connection is not exist")
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))
			rest.ReplyError(c, httpErr)
			return
		}
	}

	data, err := r.lgService.SearchFetchSameFieldsByConn(ctx, jobID, conn)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_LogGroup_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, data)
}

func (r *restHandler) ProxySearchContext(c *gin.Context) {
	logger.Debug("Handler ProxySearchContext Start")

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver layer: ProxySearchContext", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	userID := c.GetHeader("User")

	queryParams := url.Values{}
	if value, ok := c.GetQuery("id"); ok {
		queryParams.Add("id", value)
	}
	if value, ok := c.GetQuery("index"); ok {
		queryParams.Add("index", value)
	}
	if value, ok := c.GetQuery("type"); ok {
		queryParams.Add("type", value)
	}
	if value, ok := c.GetQuery("size"); ok {
		queryParams.Add("size", value)
	}
	if value, ok := c.GetQuery("orient"); ok {
		queryParams.Add("orient", value)
	}

	var conn *interfaces.DataConnection = nil
	connID := c.GetHeader("X-Data-Connection-ID")
	if connID != "" {
		userID = interfaces.AR_ADMIN_ID

		var err error
		var exist bool
		conn, exist, err = r.dcService.GetDataConnectionByID(ctx, connID)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_LogGroup_InternalError).
				WithErrorDetails(err.Error())
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))
			rest.ReplyError(c, httpErr)
			return
		}
		if !exist {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_LogGroup_DataConnectionNotFound).
				WithErrorDetails("data connection is not exist")
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
				httpErr.BaseError.ErrorDetails))
			rest.ReplyError(c, httpErr)
			return
		}
	}

	data, err := r.lgService.SearchContextByConn(ctx, queryParams, userID, conn)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_LogGroup_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, data)
}
