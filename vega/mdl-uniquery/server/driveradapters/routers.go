// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_metric"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/middleware"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/metric"

	"uniquery/common"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics/data_connection"
	"uniquery/logics/data_view"
	"uniquery/logics/dsl"
	"uniquery/logics/event"
	"uniquery/logics/log_group"
	"uniquery/logics/metric_model"
	"uniquery/logics/objective_model"
	"uniquery/logics/promql"
	utrace "uniquery/logics/trace"
	"uniquery/logics/trace_model"
	"uniquery/version"
)

const (
	CONTENT_TYPE_FORM      = "application/x-www-form-urlencoded"
	CONTENT_TYPE_JSON      = "application/json"
	X_HTTP_METHOD_OVERRIDE = "X-HTTP-Method-Override"
	SERVICE_NAME           = "uniquery"
	METRIC_MODEL_MODULE    = "metric_model"
	OBJECTIVE_MODEL_MODULE = "objective_model"

	// 指标模型的指标数据查询预览时， trace 添加的属性
	SERIES_TOTAL           = "current_request_series_total"
	POINT_TOTAL            = "current_request_point_total"
	POINT_TOTAL_PER_SERIES = "current_request_point_total_per_series"

	// 可观测性指标
	REQUEST_DURATION_NAME        = "uniquery_http_request_duration_milliseconds"
	REQUEST_COUNT_NAME           = "uniquery_http_request_count"
	REQUEST_DURATION_DESCRIPTION = "a histogram with uniquery http request duration"
	REQUEST_COUNT_DESCRIPTION    = "http request count"
	DURATION_UNIT                = "milliseconds"
	COUNT_UNIT                   = "count"
)

var (
	REQUEST_DURATION_BUCKETS = []float64{25, 50, 75, 100, 250, 500, 1000, 1500, 2000, 2500, 5000, 10000, 20000}
)

type RestHandler interface {
	RegisterPublic(engine *gin.Engine)
}

type SubHandler struct {
	appSetting *common.AppSetting
	subService interfaces.EventSubService
}

type restHandler struct {
	appSetting    *common.AppSetting
	hydra         rest.Hydra
	dcService     interfaces.DataConnectionService
	dvService     interfaces.DataViewService
	dslService    interfaces.DslService
	eService      interfaces.EventService
	lgService     interfaces.LogGroupService
	mmService     interfaces.MetricModelService
	omService     interfaces.ObjectiveModelService
	promqlService interfaces.PromQLService
	tService      interfaces.TraceService
	tmService     interfaces.TraceModelService

	// 北极星指标
	reqDurHistogram metric.Int64Histogram
	reqCounter      metric.Int64Counter
}

func NewRestHandler(appSetting *common.AppSetting) RestHandler {
	mmService := metric_model.NewMetricModelService(appSetting)
	r := &restHandler{
		appSetting:    appSetting,
		hydra:         rest.NewHydra(appSetting.HydraAdminSetting),
		dcService:     data_connection.NewDataConnectionService(appSetting),
		dvService:     data_view.NewDataViewService(appSetting),
		dslService:    dsl.NewDslService(appSetting),
		eService:      event.NewEventService(appSetting),
		lgService:     log_group.NewLogGroupService(appSetting),
		mmService:     mmService,
		omService:     objective_model.NewobjectiveModelService(appSetting),
		promqlService: promql.NewPromQLService(appSetting, mmService),
		tService:      utrace.NewTraceService(appSetting),
		tmService:     trace_model.NewTraceModelService(appSetting),
	}

	r.InitMetric()
	return r
}

func NewSubscribeHandler(appSetting *common.AppSetting) SubHandler {
	return SubHandler{
		appSetting: appSetting,
		subService: event.NewEventSubService(appSetting),
	}
}

func (r *restHandler) InitMetric() {
	histogram, _ := ar_metric.Meter.Int64Histogram(
		REQUEST_DURATION_NAME,
		metric.WithUnit(DURATION_UNIT),
		metric.WithDescription(REQUEST_DURATION_DESCRIPTION),
		metric.WithExplicitBucketBoundaries(REQUEST_DURATION_BUCKETS...),
	)

	counter, _ := ar_metric.Meter.Int64Counter(
		REQUEST_COUNT_NAME,
		metric.WithUnit(COUNT_UNIT),
		metric.WithDescription(REQUEST_COUNT_DESCRIPTION),
	)

	r.reqDurHistogram = histogram
	r.reqCounter = counter
}

func (r *restHandler) RegisterPublic(c *gin.Engine) {
	c.Use(middleware.TracingMiddleware())

	c.GET("/health", r.HealthCheck)

	// apiV1 := c.Group("/api/mdl-uniquery/v1", r.verifyOAuthMiddleWare())
	apiV1 := c.Group("/api/mdl-uniquery/v1")
	{
		apiV1.POST("/dsl/:index/_search", r.dslMiddleWare(), r.DslGetResult)
		apiV1.POST("/dsl/_search", r.dslMiddleWare(), r.DslGetResult)
		apiV1.POST("/dsl/_search/scroll", r.dslMiddleWare(), r.DslScroll)
		apiV1.POST("/dsl/:index/_count", r.dslMiddleWare(), r.DslGetCount)
		apiV1.POST("/dsl/_count", r.dslMiddleWare(), r.DslGetCount)
		apiV1.DELETE("/dsl/_search/scroll", r.dslMiddleWare(), r.DslDeleteScroll)
		apiV1.DELETE("/dsl/_search/scroll/_all", r.DslDeleteAllScroll)

		apiV1.POST("/promql/query_range", r.promqlMiddleWare(), r.PromqlQueryRange)
		apiV1.POST("/promql/query", r.promqlMiddleWare(), r.PromqlQuery)
		apiV1.POST("/promql/series", r.promqlMiddleWare(), r.PromqlSeries)

		// 指标查询接口
		apiV1.POST("/metric-model", r.verifyJsonContentTypeMiddleWare(), r.SimulateByEx)
		apiV1.POST("/metric-models/:model_ids", r.verifyJsonContentTypeMiddleWare(), r.GetMetricModelDataByEx)
		apiV1.GET("/metric-models/:model_ids/fields", r.GetMetricModelFields)
		apiV1.GET("/metric-models/:model_ids/field_values/:field_name", r.GetMetricModelFieldValues)
		apiV1.GET("/metric-models/:model_ids/labels", r.GetMetricModelLabels)

		// 获取span列表
		// apiV1.GET("/spans", r.GetSpanList)
		// 查看trace详情
		//apiV1.GET("/traces/:trace_id", middleware.PermissionVerifyMiddleware, r.GetTraceDetail) // Deprecated, 待链路模型上线后去除
		apiV1.GET("/traces/:trace_id", r.GetTraceDetail) // Deprecated, 待链路模型上线后去除

		// 视图查询接口
		apiV1.POST("/data-views", r.verifyJsonContentTypeMiddleWare(), r.ViewSimulateByEx)
		// apiV1.POST("/data-views/:view_ids", r.verifyJsonContentTypeMiddleWare(), r.GetViewDataV1)
		apiV1.POST("/data-views/:view_ids", r.verifyJsonContentTypeMiddleWare(), r.GetViewDataByEx)
		apiV1.POST("/data-view-pits", r.verifyJsonContentTypeMiddleWare(), r.DeleteDataViewPitsByEx)

		// 链路查询接口
		// (1) 预览span列表
		apiV1.POST("/simulate-traces/:trace_id/spans", r.verifyJsonContentTypeMiddleWare(), r.PreviewSpanListByEx)
		// (2) 查询span列表
		apiV1.POST("/trace-models/:trace_model_id/traces/:trace_id/spans", r.verifyJsonContentTypeMiddleWare(), r.GetSpanListByEx)
		// (3) 预览trace详情
		apiV1.POST("/simulate-traces/:trace_id", r.verifyJsonContentTypeMiddleWare(), r.PreviewTraceByEx)
		// (4) 查询trace详情
		apiV1.POST("/trace-models/:trace_model_id/traces/:trace_id", r.verifyJsonContentTypeMiddleWare(), r.GetTraceByEx)
		// (5) 预览span详情
		apiV1.POST("/simulate-traces/:trace_id/spans/:span_id", r.verifyJsonContentTypeMiddleWare(), r.PreviewSpanByEx)
		// (6) 查询span详情
		apiV1.GET("/trace-models/:trace_model_id/traces/:trace_id/spans/:span_id", r.GetSpanByEx)
		// (7) 预览span的关联日志列表
		apiV1.POST("/simulate-traces/:trace_id/spans/:span_id/related-logs", r.verifyJsonContentTypeMiddleWare(), r.PreviewSpanRelatedLogListByEx)
		// (8) 查询span的关联日志列表
		apiV1.POST("/trace-models/:trace_model_id/traces/:trace_id/spans/:span_id/related-logs", r.GetSpanRelatedLogListByEx)

		// 目标模型的指标查询接口
		apiV1.POST("/objective-models", r.verifyJsonContentTypeMiddleWare(), r.ObjectiveSimulateByEx)
		apiV1.POST("/objective-models/:model_id", r.verifyJsonContentTypeMiddleWare(), r.GetObjectiveModelDataByEx)

		// 事件模型的数据查询接口
		apiV1.POST("/events", r.QueryByEx)
		apiV1.GET("/event-models/:event_model_id/events/:event_id", r.QuerySingleEventByEventIdByEx)
	}

	apiInV1 := c.Group("/api/mdl-uniquery/in/v1")
	{
		// 指标查询接口
		apiInV1.POST("/metric-model", r.verifyJsonContentTypeMiddleWare(), r.SimulateByIn)
		apiInV1.POST("/metric-models/:model_ids", r.verifyJsonContentTypeMiddleWare(), r.GetMetricModelDataByIn)

		// 视图查询接口
		apiInV1.POST("/data-views", r.verifyJsonContentTypeMiddleWare(), r.ViewSimulateByIn)
		apiInV1.POST("/data-views/:view_ids", r.verifyJsonContentTypeMiddleWare(), r.GetViewDataByIn)
		apiInV1.POST("/data-view-pits", r.verifyJsonContentTypeMiddleWare(), r.DeleteDataViewPitsByIn)

		// 目标模型的指标查询接口
		apiInV1.POST("/objective-models", r.verifyJsonContentTypeMiddleWare(), r.ObjectiveSimulateByIn)
		apiInV1.POST("/objective-models/:model_id", r.verifyJsonContentTypeMiddleWare(), r.GetObjectiveModelDataByIn)

		// 事件模型的数据查询接口
		apiInV1.POST("/events", r.QueryByIn)
		apiInV1.GET("/event-models/:event_model_id/events/:event_id", r.QuerySingleEventByEventIdByIn)

		// 链路查询内部接口
		// (1) 预览span列表
		apiInV1.POST("/simulate-traces/:trace_id/spans", r.verifyJsonContentTypeMiddleWare(), r.PreviewSpanListByIn)
		// (2) 查询span列表
		apiInV1.POST("/trace-models/:trace_model_id/traces/:trace_id/spans", r.verifyJsonContentTypeMiddleWare(), r.GetSpanListByIn)
		// (3) 预览trace详情
		apiInV1.POST("/simulate-traces/:trace_id", r.verifyJsonContentTypeMiddleWare(), r.PreviewTraceByIn)
		// (4) 查询trace详情
		apiInV1.POST("/trace-models/:trace_model_id/traces/:trace_id", r.verifyJsonContentTypeMiddleWare(), r.GetTraceByIn)
		// (5) 预览span详情
		apiInV1.POST("/simulate-traces/:trace_id/spans/:span_id", r.verifyJsonContentTypeMiddleWare(), r.PreviewSpanByIn)
		// (6) 查询span详情
		apiInV1.GET("/trace-models/:trace_model_id/traces/:trace_id/spans/:span_id", r.GetSpanByIn)
		// (7) 预览span的关联日志列表
		apiInV1.POST("/simulate-traces/:trace_id/spans/:span_id/related-logs", r.verifyJsonContentTypeMiddleWare(), r.PreviewSpanRelatedLogListByIn)
		// (8) 查询span的关联日志列表
		apiInV1.POST("/trace-models/:trace_model_id/traces/:trace_id/spans/:span_id/related-logs", r.GetSpanRelatedLogListByIn)
	}

	// promql api 遵循prometheus api规则开放对应api给grafana
	apiV1_promql := c.Group("/api/v1")
	{
		apiV1_promql.POST("/query_range", r.promqlMiddleWare(), r.PromqlQueryRange)
		apiV1_promql.POST("/query", r.promqlMiddleWare(), r.PromqlQuery)
		apiV1_promql.POST("/series", r.promqlMiddleWare(), r.PromqlSeries)
	}

	apiV1_loggroup := c.Group("/api/mdl-uniquery/v1/loggroup")
	{
		apiV1_loggroup.GET("/manager/roots", r.ProxyGetLogGroupRoots)
		apiV1_loggroup.GET("/manager/:id/tree", r.ProxyGetLogGroupTree)
		apiV1_loggroup.GET("/manager/:id/children", r.ProxyGetLogGroupChildren)

		apiV1_loggroup.POST("/search/submit", r.ProxySearchSubmit)
		apiV1_loggroup.GET("/search/fetch/:job_id", r.ProxySearchFetch)
		apiV1_loggroup.GET("/search/fetch/:job_id/fields", r.ProxySearchFetchFields)
		apiV1_loggroup.GET("/search/fetch/:job_id/fields/sameFields", r.ProxySearchFetchSameFields)
		apiV1_loggroup.GET("/search/context", r.ProxySearchContext)
	}

	logger.Info("RestHandler RegisterPublic")
}

// HealthCheck 健康检查
func (r *restHandler) HealthCheck(c *gin.Context) {
	// 返回服务信息
	serverInfo := o11y.ServerInfo{
		ServerName:    version.ServerName,
		ServerVersion: version.ServerVersion,
		Language:      version.LanguageGo,
		GoVersion:     version.GoVersion,
		GoArch:        version.GoArch,
	}
	rest.ReplyOK(c, http.StatusOK, serverInfo)
}

// gin中间件 校验content type
func (r *restHandler) verifyJsonContentTypeMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		//拦截请求，判断ContentType是否为XXX
		if c.ContentType() != CONTENT_TYPE_JSON {
			httpErr := rest.NewHTTPError(c, http.StatusNotAcceptable, uerrors.Uniquery_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("Content-Type header [%s] is not supported, expected is [application/json].", c.ContentType()))
			rest.ReplyError(c, httpErr)

			c.Abort()
		}

		//执行后续操作
		c.Next()
	}
}

// gin中间件 校验content type
func (r *restHandler) dslMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		//拦截请求，判断ContentType是否为XXX
		if c.ContentType() != CONTENT_TYPE_JSON {
			common.ReplyError(c, http.StatusNotAcceptable,
				uerrors.PromQLError{
					Typ: uerrors.ErrorStatusNotAcceptable,
					Err: fmt.Errorf("Content-Type header [%s] is not supported, expected is [application/json]", c.ContentType()),
				},
			)
			c.Abort()
		}

		//执行后续操作
		c.Next()
	}
}

// gin中间件 校验content type
func (r *restHandler) promqlMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		//拦截请求，判断ContentType是否为XXX
		if c.ContentType() != CONTENT_TYPE_FORM {
			common.ReplyError(c, http.StatusNotAcceptable,
				uerrors.PromQLError{
					Typ: uerrors.ErrorStatusNotAcceptable,
					Err: fmt.Errorf("Content-Type header [%s] is not supported, expected is [application/x-www-form-urlencoded]", c.ContentType()),
				},
			)
			c.Abort()
		}

		//执行后续操作
		c.Next()
	}
}

// gin中间件 校验oauth
// func (r *restHandler) verifyOAuthMiddleWare() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ctx := rest.GetLanguageCtx(c)
// 		_, err := r.hydra.VerifyToken(ctx, c)
// 		if err != nil {
// 			httpError := rest.NewHTTPError(ctx, http.StatusUnauthorized, rest.PublicError_Unauthorized).
// 				WithErrorDetails(err.Error())
// 			rest.ReplyError(c, httpError)
// 			c.Abort()
// 			return
// 		}

// 		//执行后续操作
// 		c.Next()
// 	}
// }

// 校验oauth
func (r *restHandler) verifyOAuth(ctx context.Context, c *gin.Context) (rest.Visitor, error) {
	vistor, err := r.hydra.VerifyToken(ctx, c)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusUnauthorized, rest.PublicError_Unauthorized).
			WithErrorDetails(err.Error())
		rest.ReplyError(c, httpErr)
		return vistor, err
	}

	return vistor, nil
}

func GenerateVisitor(c *gin.Context) rest.Visitor {
	accountInfo := interfaces.AccountInfo{
		ID:   c.GetHeader(interfaces.HTTP_HEADER_ACCOUNT_ID),
		Type: c.GetHeader(interfaces.HTTP_HEADER_ACCOUNT_TYPE),
	}
	visitor := rest.Visitor{
		ID:         accountInfo.ID,
		Type:       rest.VisitorType(accountInfo.Type),
		TokenID:    "", // 无token
		IP:         c.ClientIP(),
		Mac:        c.GetHeader("X-Request-MAC"),
		UserAgent:  c.GetHeader("User-Agent"),
		ClientType: rest.ClientType_Linux,
	}
	return visitor
}
