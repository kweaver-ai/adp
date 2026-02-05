package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/middleware"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	"data-model/logics/data_connection"
	"data-model/logics/data_dict"
	"data-model/logics/data_view"
	"data-model/logics/event_model"
	"data-model/logics/metric_model"
	"data-model/logics/objective_model"
	"data-model/logics/trace_model"
	"data-model/version"
	"data-model/worker"
)

type RestHandler interface {
	RegisterPublic(engine *gin.Engine)
}

type restHandler struct {
	appSetting *common.AppSetting
	hydra      rest.Hydra
	dcs        interfaces.DataConnectionService
	dds        interfaces.DataDictService
	ddis       interfaces.DataDictItemsService
	dvs        interfaces.DataViewService
	dvgs       interfaces.DataViewGroupService
	dvms       interfaces.DataViewMonitorService
	dvrcs      interfaces.DataViewRowColumnRuleService
	ems        interfaces.EventModelService
	mms        interfaces.MetricModelService
	mmts       interfaces.MetricModelTaskService
	mmgs       interfaces.MetricModelGroupService
	oms        interfaces.ObjectiveModelService
	tms        interfaces.TraceModelService
}

func NewRestHandler(appSetting *common.AppSetting) RestHandler {
	return &restHandler{
		appSetting: appSetting,
		hydra:      rest.NewHydra(appSetting.HydraAdminSetting),
		dcs:        data_connection.NewDataConnectionService(appSetting),
		dds:        data_dict.NewDataDictService(appSetting),
		ddis:       data_dict.NewDataDictItemService(appSetting),
		dvs:        data_view.NewDataViewService(appSetting),
		dvgs:       data_view.NewDataViewGroupService(appSetting),
		dvms:       worker.NewDataViewMonitorService(appSetting),
		dvrcs:      data_view.NewDataViewRowColumnRuleService(appSetting),
		ems:        event_model.NewEventModelService(appSetting),
		mms:        metric_model.NewMetricModelService(appSetting),
		mmts:       metric_model.NewMetricModelTaskService(appSetting),
		mmgs:       metric_model.NewMetricModelGroupService(appSetting),
		oms:        objective_model.NewObjectiveModelService(appSetting),
		tms:        trace_model.NewTraceModelService(appSetting),
	}
}

func (r *restHandler) RegisterPublic(c *gin.Engine) {
	c.Use(r.AccessLog())
	c.Use(middleware.TracingMiddleware())

	c.GET("/health", r.HealthCheck)

	apiV1 := c.Group("/api/mdl-data-model/v1")
	{
		//指标模型分组
		apiV1.POST("/metric-model-groups", r.verifyJsonContentTypeMiddleWare(), r.CreateMetricModelGroup)
		apiV1.DELETE("/metric-model-groups/:group_id", r.DeleteMetricModelGroup)
		apiV1.PUT("/metric-model-groups/:group_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateMetricModelGroup)
		apiV1.GET("/metric-model-groups", r.ListMetricModelGroups)
		apiV1.GET("/metric-model-groups/:group_name/metric-models", r.GetMetricModelsInGroup) // 路径参数用group_name，实际上是group_id

		// 指标模型
		apiV1.POST("/metric-models", r.verifyJsonContentTypeMiddleWare(), r.CreateMetricModelsByEx)
		apiV1.DELETE("/metric-models/:model_ids", r.DeleteMetricModels)
		apiV1.PUT("/metric-models/:model_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateMetricModelByEx)
		apiV1.PUT("/metric-models/:model_id/attributes", r.verifyJsonContentTypeMiddleWare(), r.UpdateMetricModels)
		apiV1.GET("/metric-models", r.ListMetricModelsByEx)
		apiV1.GET("/metric-models/:model_ids", r.GetMetricModelsByEx)
		apiV1.GET("/metric-models/:model_ids/fields", r.GetMetricModelSourceFields)      // 路径参数用ids，实际上只支持单个
		apiV1.GET("/metric-models/:model_ids/order_fields", r.GetMetricModelOrderFields) // 支持多个

		// 指标模型持久化任务
		apiV1.GET("/metric-tasks/:task_id", r.GetMetricTaskByEx)
		apiV1.PUT("/metric-tasks/:task_id/attr", r.UpdateMetricTaskPlanTimeByEx)

		// event model
		apiV1.POST("/event-models", r.verifyJsonContentTypeMiddleWare(), r.CreateEventModelByEx)                //新增事件模型
		apiV1.DELETE("/event-models/:event_model_ids", r.DeleteEventModels)                                     //批量删除事件模型
		apiV1.PUT("/event-models/:event_model_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateEventModelByEx) // 更新事件模型
		apiV1.PUT("/event-task/:id/attr", r.verifyJsonContentTypeMiddleWare(), r.UpdateEventTaskStatusByEx)     //同步事件模型持久化任务执行状态
		apiV1.GET("/event-models", r.QueryEventModelsByEx)                                                      // 事件模型查询接口
		apiV1.GET("/event-models/:event_model_ids", r.QueryEventModelByIDByEx)                                  // 事件模型查询接口
		apiV1.GET("/event-level", r.QueryEventLevel)                                                            //事件级别查询

		// 数据视图
		apiV1.POST("/data-views", r.verifyJsonContentTypeMiddleWare(), r.HandleDataViewPostOverrideByEx)
		apiV1.DELETE("/data-views/:view_ids", r.DeleteDataViewsByEx)
		apiV1.PUT("/data-views/:view_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateDataViewByEx)
		apiV1.GET("/data-views/:view_ids", r.GetDataViewsByEx)
		apiV1.GET("/data-views", r.ListDataViewsByEx)
		// 路径参数用view_id，实际上是批量接口，写view_ids gin框架会报错
		apiV1.PUT("/data-views/:view_id/attrs/:fields", r.verifyJsonContentTypeMiddleWare(), r.UpdateDataViewAttrFields)
		// 数据视图分组
		apiV1.POST("/data-view-groups", r.verifyJsonContentTypeMiddleWare(), r.CreateDataViewGroup)
		apiV1.DELETE("/data-view-groups/:group_id", r.DeleteDataViewGroup)
		apiV1.PUT("/data-view-groups/:group_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateDataViewGroup)
		apiV1.GET("/data-view-groups", r.ListDataViewGroups)
		apiV1.GET("/data-view-groups/:group_id/data-views", r.GetDataViewsInGroup)

		// 数据视图行列权限
		apiV1.POST("/data-view-row-column-rules", r.verifyJsonContentTypeMiddleWare(), r.CreateDataViewRowColumnRulesByEx)
		apiV1.DELETE("/data-view-row-column-rules/:rule_ids", r.DeleteDataViewRowColumnRulesByEx)
		apiV1.PUT("/data-view-row-column-rules/:rule_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateDataViewRowColumnRuleByEx)
		apiV1.GET("/data-view-row-column-rules/:rule_ids", r.GetDataViewRowColumnRulesByEx)
		apiV1.GET("/data-view-row-column-rules", r.ListDataViewRowColumnRulesByEx)

		// 扫描数据源
		// apiV1.POST("/data-source-scan", r.ScanDataSource)
		// 获取所有数据源，数据源信息包含扫描记录
		// apiV1.GET("/data-sources", r.ListDataSourcesWithScanRecord)

		// 数据字典
		apiV1.POST("/data-dicts", r.verifyJsonContentTypeMiddleWare(), r.CreateDataDictsByEx)
		apiV1.DELETE("/data-dicts/:dict_id", r.DeleteDataDicts)
		apiV1.PUT("/data-dicts/:dict_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateDataDictByEx)
		apiV1.GET("/data-dicts", r.ListDataDictsByEx)
		apiV1.GET("/data-dicts/:dict_id", r.GetDataDictsByEx)
		// 数据字典项
		apiV1.POST("/data-dicts/:dict_id/items", r.HandleDataDictCreateOrImportByEx)
		apiV1.DELETE("/data-dicts/:dict_id/items/:item_ids", r.DeleteDataDictItems)
		apiV1.PUT("/data-dicts/:dict_id/items/:item_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateDataDictItemByEx)
		apiV1.GET("/data-dicts/:dict_id/items", r.HandleDataDictListOrExportByEx)

		// 数据连接
		apiV1.POST("/data-connections", r.verifyJsonContentTypeMiddleWare(), r.CreateDataConnection)
		apiV1.DELETE("/data-connections/:conn_ids", r.DeleteDataConnections)
		apiV1.PUT("/data-connections/:conn_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateDataConnection)
		apiV1.GET("/data-connections/:conn_id", r.GetDataConnection)
		apiV1.GET("/data-connections", r.ListDataConnections)

		// 链路模型
		apiV1.POST("/trace-models", r.verifyJsonContentTypeMiddleWare(), r.CreateTraceModelsByEx)
		apiV1.POST("/simulate-trace-models", r.verifyJsonContentTypeMiddleWare(), r.SimulateCreateTraceModelByEx)
		apiV1.DELETE("/trace-models/:model_ids", r.DeleteTraceModelsByEx)
		apiV1.PUT("/trace-models/:model_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateTraceModelByEx)
		apiV1.PUT("/simulate-trace-models/:model_id", r.verifyJsonContentTypeMiddleWare(), r.SimulateUpdateTraceModelByEx)
		apiV1.GET("/trace-models/:model_ids", r.GetTraceModelsByEx)
		apiV1.GET("/trace-models", r.ListTraceModelsByEx)
		apiV1.GET("/trace-models/:model_ids/field-info", r.GetTraceModelFieldInfoByEx) // path参数不能用:model_id, 会报router conflict

		// 目标模型
		apiV1.POST("/objective-models", r.verifyJsonContentTypeMiddleWare(), r.CreateObjectiveModelsByEx)
		apiV1.GET("/objective-models", r.ListObjectiveModelsByEx)
		apiV1.GET("/objective-models/:model_ids", r.GetObjectiveModelsByEx)
		apiV1.PUT("/objective-models/:model_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateObjectiveModelByEx)
		apiV1.DELETE("/objective-models/:model_ids", r.DeleteObjectiveModels)

		// 临时代码，为防止因tag-mgmt服务不存在，而导致的前端报错503
		apiV1.GET("/object-tags", r.GetObjectTag)

		// 数据模型资源示例列表
		apiV1.GET("/resources", r.ListResources)
	}

	apiInV1 := c.Group("/api/mdl-data-model/in/v1")
	{
		// 指标模型
		apiInV1.POST("/metric-models", r.verifyJsonContentTypeMiddleWare(), r.CreateMetricModelsByIn)
		apiInV1.PUT("/metric-models/:model_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateMetricModelByIn)
		apiInV1.GET("/metric-models", r.ListMetricModelsByIn)
		apiInV1.GET("/metric-models/:model_ids", r.GetMetricModelsByIn)
		// 指标模型持久化任务
		apiInV1.GET("/metric-tasks/:task_id", r.GetMetricTaskByIn)
		apiInV1.PUT("/metric-tasks/:task_id/attr", r.UpdateMetricTaskPlanTimeByIn)

		// 数据视图
		apiInV1.POST("/data-views", r.verifyJsonContentTypeMiddleWare(), r.HandleDataViewPostOverrideByIn)
		apiInV1.DELETE("/data-views/:view_ids", r.DeleteDataViewsByIn)
		apiInV1.PUT("/data-views/:view_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateDataViewByIn)
		apiInV1.GET("/data-views/:view_ids", r.GetDataViewsByIn)
		apiInV1.GET("/data-views", r.ListDataViewsByIn)

		// 数据视图行列权限
		// apiInV1.POST("/data-view-row-column-rules", r.verifyJsonContentTypeMiddleWare(), r.CreateDataViewRowColumnRulesByIn)
		// apiInV1.DELETE("/data-view-row-column-rules/:rule_ids", r.DeleteDataViewRowColumnRulesByIn)
		// apiInV1.PUT("/data-view-row-column-rules/:rule_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateDataViewRowColumnRuleByIn)
		// apiInV1.GET("/data-view-row-column-rules/:rule_ids", r.GetDataViewRowColumnRulesByIn)
		apiInV1.GET("/data-view-row-column-rules", r.ListDataViewRowColumnRulesByIn)

		// 目标模型
		apiInV1.POST("/objective-models", r.verifyJsonContentTypeMiddleWare(), r.CreateObjectiveModelsByIn)
		apiInV1.GET("/objective-models", r.ListObjectiveModelsByIn)
		apiInV1.GET("/objective-models/:model_ids", r.GetObjectiveModelsByIn)
		apiInV1.PUT("/objective-models/:model_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateObjectiveModelByIn)

		// event model
		apiInV1.POST("/event-models", r.verifyJsonContentTypeMiddleWare(), r.CreateEventModelByIn)                //新增事件模型
		apiInV1.PUT("/event-models/:event_model_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateEventModelByIn) // 更新事件模型
		apiInV1.GET("/event-models", r.QueryEventModelsByIn)                                                      // 事件模型列表接口
		apiInV1.GET("/event-models/:event_model_ids", r.QueryEventModelByIDByIn)                                  // 事件模型详情接口
		apiInV1.PUT("/event-task/:id/attr", r.verifyJsonContentTypeMiddleWare(), r.UpdateEventTaskStatusByIn)     // data-model-job同步持久化任务执行状态

		// 链路模型
		apiInV1.POST("/trace-models", r.verifyJsonContentTypeMiddleWare(), r.CreateTraceModelsByIn)
		apiInV1.POST("/simulate-trace-models", r.verifyJsonContentTypeMiddleWare(), r.SimulateCreateTraceModelByIn)
		apiInV1.DELETE("/trace-models/:model_ids", r.DeleteTraceModelsByIn)
		apiInV1.PUT("/trace-models/:model_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateTraceModelByIn)
		apiInV1.PUT("/simulate-trace-models/:model_id", r.verifyJsonContentTypeMiddleWare(), r.SimulateUpdateTraceModelByIn)
		apiInV1.GET("/trace-models/:model_ids", r.GetTraceModelsByIn)
		apiInV1.GET("/trace-models", r.ListTraceModelsByIn)
		apiInV1.GET("/trace-models/:model_ids/field-info", r.GetTraceModelFieldInfoByIn) // path参数不能用:model_id, 会报router conflict

		// 数据字典
		apiInV1.POST("/data-dicts", r.verifyJsonContentTypeMiddleWare(), r.CreateDataDictsByIn)
		apiInV1.PUT("/data-dicts/:dict_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateDataDictByIn)
		apiInV1.GET("/data-dicts", r.ListDataDictsByIn)
		apiInV1.GET("/data-dicts/:dict_id", r.GetDataDictsByIn)
		// 数据字典项
		apiInV1.POST("/data-dicts/:dict_id/items", r.HandleDataDictCreateOrImportByIn)
		apiInV1.PUT("/data-dicts/:dict_id/items/:item_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateDataDictItemByIn)
		apiInV1.GET("/data-dicts/:dict_id/items", r.HandleDataDictListOrExportByIn)
	}

	logger.Info("RestHandler RegisterPublic")
}

func (r *restHandler) GetObjectTag(c *gin.Context) {
	result := map[string]any{
		"entries":     []any{},
		"total_count": 0,
	}
	rest.ReplyOK(c, http.StatusOK, result)
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

// gin中间件 校验content type
func (r *restHandler) verifyJsonContentTypeMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		//拦截请求，判断ContentType是否为XXX
		if c.ContentType() != interfaces.CONTENT_TYPE_JSON {
			httpErr := rest.NewHTTPError(c, http.StatusNotAcceptable, derrors.DataModel_InvalidRequestHeader_ContentType).
				WithErrorDetails(fmt.Sprintf("Content-Type header [%s] is not supported, expected is [application/json].", c.ContentType()))
			rest.ReplyError(c, httpErr)

			c.Abort()
		}

		//执行后续操作
		c.Next()
	}
}

// gin中间件 访问日志
func (r *restHandler) AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {

		beginTime := time.Now()
		c.Next()
		endTime := time.Now()
		durTime := endTime.Sub(beginTime).Seconds()

		logger.Debugf("access log: url: %s, method: %s, begin_time: %s, end_time: %s, subTime: %f",
			c.Request.URL.Path,
			c.Request.Method,
			beginTime.Format(libCommon.RFC3339Milli),
			endTime.Format(libCommon.RFC3339Milli),
			durTime,
		)
	}
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
