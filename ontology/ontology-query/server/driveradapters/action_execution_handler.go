package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
)

// ExecuteActionByIn handles action execution request (internal)
func (r *restHandler) ExecuteActionByIn(c *gin.Context) {
	logger.Debug("Handler ExecuteActionByIn Start")
	visitor := GenerateVisitor(c)
	r.ExecuteAction(c, visitor)
}

// ExecuteActionByEx handles action execution request (external)
func (r *restHandler) ExecuteActionByEx(c *gin.Context) {
	logger.Debug("Handler ExecuteActionByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "执行行动类API",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ExecuteAction(c, visitor)
}

// ExecuteAction handles the action execution request
func (r *restHandler) ExecuteAction(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler ExecuteAction Start")
	startTime := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "执行行动类API",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))
	o11y.Info(ctx, fmt.Sprintf("行动执行请求参数: [%s,%v]", c.Request.RequestURI, c.Request.Body))

	// Get path parameters
	knID := c.Param("kn_id")
	atID := c.Param("at_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("at_id").String(atID),
		attr.Key("branch").String(branch),
	)

	// Bind request body
	req := interfaces.ActionExecutionRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ActionExecution_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Binding Parameter Failed: %s", err.Error()))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	req.KNID = knID
	req.Branch = branch
	req.ActionTypeID = atID

	// Note: unique_identities is optional
	// If not provided, the action will apply to all entities matching the action type's conditions

	// Execute action
	result, err := r.ass.ExecuteAction(ctx, &req)
	if err != nil {
		httpErr, ok := err.(*rest.HTTPError)
		if !ok {
			httpErr = rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError).
				WithErrorDetails(err.Error())
		}

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusAccepted)
	logger.Debugf("ExecuteAction completed in %dms", time.Since(startTime).Milliseconds())
	rest.ReplyOK(c, http.StatusAccepted, result)
}

// GetActionExecutionByIn handles get execution status request (internal)
func (r *restHandler) GetActionExecutionByIn(c *gin.Context) {
	logger.Debug("Handler GetActionExecutionByIn Start")
	visitor := GenerateVisitor(c)
	r.GetActionExecution(c, visitor)
}

// GetActionExecutionByEx handles get execution status request (external)
func (r *restHandler) GetActionExecutionByEx(c *gin.Context) {
	logger.Debug("Handler GetActionExecutionByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "获取行动执行状态API",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetActionExecution(c, visitor)
}

// GetActionExecution handles the get execution status request
func (r *restHandler) GetActionExecution(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler GetActionExecution Start")
	startTime := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "获取行动执行状态API",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// Get path parameters
	knID := c.Param("kn_id")
	executionID := c.Param("execution_id")
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("execution_id").String(executionID),
	)

	// Get execution
	result, err := r.ass.GetExecution(ctx, knID, executionID)
	if err != nil {
		httpErr, ok := err.(*rest.HTTPError)
		if !ok {
			httpErr = rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError).
				WithErrorDetails(err.Error())
		}

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	logger.Debugf("GetActionExecution completed in %dms", time.Since(startTime).Milliseconds())
	rest.ReplyOK(c, http.StatusOK, result)
}

// QueryActionLogsOverrideByIn handles query action logs request (internal)
func (r *restHandler) QueryActionLogsOverrideByIn(c *gin.Context) {
	logger.Debug("Handler QueryActionLogsOverrideByIn Start")
	visitor := GenerateVisitor(c)
	r.QueryActionLogsOverride(c, visitor)
}

// QueryActionLogsOverrideByEx handles query action logs request (external)
func (r *restHandler) QueryActionLogsOverrideByEx(c *gin.Context) {
	logger.Debug("Handler QueryActionLogsOverrideByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "查询行动执行日志API",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.QueryActionLogsOverride(c, visitor)
}

// QueryActionLogsOverride handles the query action logs request (POST with method override)
func (r *restHandler) QueryActionLogsOverride(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler QueryActionLogsOverride Start")
	startTime := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "查询行动执行日志API",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))
	o11y.Info(ctx, fmt.Sprintf("行动日志查询请求参数: [%s,%v]", c.Request.RequestURI, c.Request.Body))

	// Validate method override header
	if err := ValidateHeaderMethodOverride(ctx, c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE)); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	// Get path parameters
	knID := c.Param("kn_id")
	span.SetAttributes(attr.Key("kn_id").String(knID))

	// Bind request body
	query := interfaces.ActionLogQuery{}
	if err := c.ShouldBindJSON(&query); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ActionExecution_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Binding Parameter Failed: %s", err.Error()))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	query.KNID = knID

	// Query executions
	result, err := r.als.QueryExecutions(ctx, &query)
	if err != nil {
		httpErr, ok := err.(*rest.HTTPError)
		if !ok {
			httpErr = rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_ActionExecution_QueryExecutionsFailed).
				WithErrorDetails(err.Error())
		}

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	logger.Debugf("QueryActionLogs completed in %dms", time.Since(startTime).Milliseconds())
	rest.ReplyOK(c, http.StatusOK, result)
}

// GetActionLogByIn handles get single action log request (internal)
func (r *restHandler) GetActionLogByIn(c *gin.Context) {
	logger.Debug("Handler GetActionLogByIn Start")
	visitor := GenerateVisitor(c)
	r.GetActionLog(c, visitor)
}

// GetActionLogByEx handles get single action log request (external)
func (r *restHandler) GetActionLogByEx(c *gin.Context) {
	logger.Debug("Handler GetActionLogByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "获取行动执行日志详情API",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetActionLog(c, visitor)
}

// GetActionLog handles the get single action log request
func (r *restHandler) GetActionLog(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler GetActionLog Start")
	startTime := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "获取行动执行日志详情API",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// Get path parameters
	knID := c.Param("kn_id")
	logID := c.Param("log_id")
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("log_id").String(logID),
	)

	// Get execution log
	result, err := r.als.GetExecution(ctx, knID, logID)
	if err != nil {
		httpErr, ok := err.(*rest.HTTPError)
		if !ok {
			httpErr = rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.OntologyQuery_ActionExecution_ExecutionNotFound).
				WithErrorDetails(err.Error())
		}

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	logger.Debugf("GetActionLog completed in %dms", time.Since(startTime).Milliseconds())
	rest.ReplyOK(c, http.StatusOK, result)
}
