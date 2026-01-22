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

	"ontology-manager/common"
	oerrors "ontology-manager/errors"
	"ontology-manager/interfaces"
	"ontology-manager/logics/action_schedule"
	"ontology-manager/logics/action_type"
	"ontology-manager/logics/concept_group"
	"ontology-manager/logics/job"
	"ontology-manager/logics/knowledge_network"
	"ontology-manager/logics/object_type"
	"ontology-manager/logics/relation_type"
	"ontology-manager/version"
)

type RestHandler interface {
	RegisterPublic(engine *gin.Engine)
}

type restHandler struct {
	appSetting *common.AppSetting
	hydra      rest.Hydra
	ass        interfaces.ActionScheduleService
	ats        interfaces.ActionTypeService
	cgs        interfaces.ConceptGroupService
	js         interfaces.JobService
	kns        interfaces.KNService
	ots        interfaces.ObjectTypeService
	rts        interfaces.RelationTypeService
}

func NewRestHandler(appSetting *common.AppSetting) RestHandler {
	r := &restHandler{
		appSetting: appSetting,
		hydra:      rest.NewHydra(appSetting.HydraAdminSetting),
		ass:        action_schedule.NewActionScheduleService(appSetting),
		ats:        action_type.NewActionTypeService(appSetting),
		cgs:        concept_group.NewConceptGroupService(appSetting),
		js:         job.NewJobService(appSetting),
		kns:        knowledge_network.NewKNService(appSetting),
		ots:        object_type.NewObjectTypeService(appSetting),
		rts:        relation_type.NewRelationTypeService(appSetting),
	}
	return r
}

func (r *restHandler) RegisterPublic(c *gin.Engine) {
	c.Use(r.AccessLog())
	c.Use(middleware.TracingMiddleware())

	c.GET("/health", r.HealthCheck)

	apiV1 := c.Group("/api/ontology-manager/v1")
	{
		// 业务知识网络
		apiV1.POST("/knowledge-networks", r.verifyJsonContentTypeMiddleWare(), r.CreateKNByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id", r.DeleteKN)
		apiV1.PUT("/knowledge-networks/:kn_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateKNByEx)
		apiV1.GET("/knowledge-networks", r.ListKNsByEx)
		apiV1.GET("/knowledge-networks/:kn_id", r.GetKNByEx)
		apiV1.POST("/knowledge-networks/:kn_id/relation-type-paths", r.GetRelationTypePathsByEx)

		// 概念分组
		apiV1.POST("/knowledge-networks/:kn_id/concept-groups", r.verifyJsonContentTypeMiddleWare(), r.CreateConceptGroupByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id/concept-groups/:cg_id", r.DeleteConceptGroup) // 不支持批量删
		apiV1.PUT("/knowledge-networks/:kn_id/concept-groups/:cg_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateConceptGroupByEx)
		apiV1.GET("/knowledge-networks/:kn_id/concept-groups", r.ListConceptGroupsByEx)
		apiV1.GET("/knowledge-networks/:kn_id/concept-groups/:cg_id", r.GetConceptGroupByEx)
		apiV1.POST("/knowledge-networks/:kn_id/concept-groups/:cg_id/object-types", r.AddObjectTypesToConceptGroupByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id/concept-groups/:cg_id/object-types/:ot_ids", r.DeleteObjectTypesFromGroupByEx)

		// 对象类
		apiV1.POST("/knowledge-networks/:kn_id/object-types", r.verifyJsonContentTypeMiddleWare(), r.HandleObjectTypeGetOverrideByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id/object-types/:ot_ids", r.DeleteObjectTypes) // path上用kn_ids接，实际上只能传一个id
		apiV1.PUT("/knowledge-networks/:kn_id/object-types/:ot_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateObjectTypeByEx)
		apiV1.PUT("/knowledge-networks/:kn_id/object-types/:ot_id/data_properties/:property_names", r.verifyJsonContentTypeMiddleWare(), r.UpdateDataProperties)
		apiV1.GET("/knowledge-networks/:kn_id/object-types", r.ListObjectTypesByEx)        // path上用kn_ids接，实际上只能传一个id
		apiV1.GET("/knowledge-networks/:kn_id/object-types/:ot_ids", r.GetObjectTypesByEx) // path上用kn_ids接，实际上只能传一个id

		// 关系类
		apiV1.POST("/knowledge-networks/:kn_id/relation-types", r.verifyJsonContentTypeMiddleWare(), r.HandleRelationTypeGetOverrideByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id/relation-types/:rt_ids", r.DeleteRelationTypes)
		apiV1.PUT("/knowledge-networks/:kn_id/relation-types/:rt_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateRelationTypeByEx)
		apiV1.GET("/knowledge-networks/:kn_id/relation-types", r.ListRelationTypesByEx)
		apiV1.GET("/knowledge-networks/:kn_id/relation-types/:rt_ids", r.GetRelationTypesByEx)

		// 行动类
		apiV1.POST("/knowledge-networks/:kn_id/action-types", r.verifyJsonContentTypeMiddleWare(), r.HandleActionTypeGetOverrideByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id/action-types/:at_ids", r.DeleteActionTypes)
		apiV1.PUT("/knowledge-networks/:kn_id/action-types/:at_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateActionTypeByEx)
		apiV1.GET("/knowledge-networks/:kn_id/action-types", r.ListActionTypesByEx)
		apiV1.GET("/knowledge-networks/:kn_id/action-types/:at_ids", r.GetActionTypesByEx)

		// 任务管理
		apiV1.POST("/knowledge-networks/:kn_id/jobs", r.verifyJsonContentTypeMiddleWare(), r.CreateJobByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id/jobs/:job_ids", r.DeleteJobsByEx)
		apiV1.GET("/knowledge-networks/:kn_id/jobs", r.ListJobsByEx)
		apiV1.GET("/knowledge-networks/:kn_id/jobs/:job_id/tasks", r.ListTasksByEx)

		// 行动计划管理
		apiV1.POST("/knowledge-networks/:kn_id/action-schedules", r.verifyJsonContentTypeMiddleWare(), r.CreateActionScheduleByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id/action-schedules/:schedule_ids", r.DeleteActionSchedulesByEx)
		apiV1.PUT("/knowledge-networks/:kn_id/action-schedules/:schedule_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateActionScheduleByEx)
		apiV1.PUT("/knowledge-networks/:kn_id/action-schedules/:schedule_id/status", r.verifyJsonContentTypeMiddleWare(), r.UpdateActionScheduleStatusByEx)
		apiV1.GET("/knowledge-networks/:kn_id/action-schedules", r.ListActionSchedulesByEx)
		apiV1.GET("/knowledge-networks/:kn_id/action-schedules/:schedule_id", r.GetActionScheduleByEx)

		// 业务知识网络资源示例列表
		apiV1.GET("/resources", r.ListResources)
	}

	apiInV1 := c.Group("/api/ontology-manager/in/v1")
	{
		// 业务知识网络
		apiInV1.POST("/knowledge-networks", r.verifyJsonContentTypeMiddleWare(), r.CreateKNByIn)
		apiInV1.PUT("/knowledge-networks/:kn_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateKNByIn)
		apiInV1.GET("/knowledge-networks", r.ListKNsByIn)
		apiInV1.GET("/knowledge-networks/:kn_id", r.GetKNByIn)
		apiInV1.POST("/knowledge-networks/:kn_id/relation-type-paths", r.GetRelationTypePathsByIn)

		// 概念分组
		apiInV1.POST("/knowledge-networks/:kn_id/concept-groups", r.verifyJsonContentTypeMiddleWare(), r.CreateConceptGroupByIn)
		apiInV1.PUT("/knowledge-networks/:kn_id/concept-groups/:cg_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateConceptGroupByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/concept-groups", r.ListConceptGroupsByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/concept-groups/:cg_id", r.GetConceptGroupByIn)
		apiInV1.POST("/knowledge-networks/:kn_id/concept-groups/:cg_id/object-types", r.AddObjectTypesToConceptGroupByIn)
		apiInV1.DELETE("/knowledge-networks/:kn_id/concept-groups/:cg_id/object-types/:ot_ids", r.DeleteObjectTypesFromGroupByIn)

		// 对象类
		apiInV1.POST("/knowledge-networks/:kn_id/object-types", r.verifyJsonContentTypeMiddleWare(), r.HandleObjectTypeGetOverrideByIn)
		apiInV1.PUT("/knowledge-networks/:kn_id/object-types/:ot_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateObjectTypeByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/object-types", r.ListObjectTypesByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/object-types/:ot_ids", r.GetObjectTypesByIn) // path上用kn_ids接，实际上只能传一个id

		// 关系类
		apiInV1.POST("/knowledge-networks/:kn_id/relation-types", r.verifyJsonContentTypeMiddleWare(), r.HandleRelationTypeGetOverrideByIn)
		apiInV1.PUT("/knowledge-networks/:kn_id/relation-types/:rt_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateRelationTypeByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/relation-types", r.ListRelationTypesByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/relation-types/:rt_ids", r.GetRelationTypesByIn)

		// 行动类
		apiInV1.POST("/knowledge-networks/:kn_id/action-types", r.verifyJsonContentTypeMiddleWare(), r.HandleActionTypeGetOverrideByIn)
		apiInV1.PUT("/knowledge-networks/:kn_id/action-types/:at_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateActionTypeByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/action-types", r.ListActionTypesByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/action-types/:at_ids", r.GetActionTypesByIn)

		// 行动计划管理
		apiInV1.POST("/knowledge-networks/:kn_id/action-schedules", r.verifyJsonContentTypeMiddleWare(), r.CreateActionScheduleByIn)
		apiInV1.DELETE("/knowledge-networks/:kn_id/action-schedules/:schedule_ids", r.DeleteActionSchedulesByIn)
		apiInV1.PUT("/knowledge-networks/:kn_id/action-schedules/:schedule_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateActionScheduleByIn)
		apiInV1.PUT("/knowledge-networks/:kn_id/action-schedules/:schedule_id/status", r.verifyJsonContentTypeMiddleWare(), r.UpdateActionScheduleStatusByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/action-schedules", r.ListActionSchedulesByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/action-schedules/:schedule_id", r.GetActionScheduleByIn)

		// 任务管理
		apiInV1.POST("/knowledge-networks/:kn_id/jobs", r.verifyJsonContentTypeMiddleWare(), r.CreateJobByIn)
		apiInV1.DELETE("/knowledge-networks/:kn_id/jobs/:job_ids", r.DeleteJobsByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/jobs", r.ListJobsByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/jobs/:job_id/tasks", r.ListTasksByIn)
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
		if c.ContentType() != interfaces.CONTENT_TYPE_JSON {
			httpErr := rest.NewHTTPError(c, http.StatusNotAcceptable, oerrors.OntologyManager_InvalidRequestHeader_ContentType).
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
