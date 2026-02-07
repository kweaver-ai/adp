// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
)

// TODO : 迁移到interfaces
var (
	//TODO:  统一事件模型对象和
	//TODO:  数据内容不做国际化
	EVENT_MODEL_LEVEL_ZH_CN = []map[string]any{
		{"id": 1, "event_level": "紧急", "health_level": "不可用", "health_score": []int{20, 40}},
		{"id": 2, "event_level": "主要", "health_level": "故障", "health_score": []int{40, 60}},
		{"id": 3, "event_level": "次要", "health_level": "错误", "health_score": []int{60, 80}},
		{"id": 4, "event_level": "提示", "health_level": "警示", "health_score": []int{80, 100}},
		{"id": 5, "event_level": "不明确", "health_level": "不明确", "health_score": []int{0, 0}},
		{"id": 6, "event_level": "清除", "health_level": "健康", "health_score": []int{100, 100}},
	}
	EVENT_MODEL_LEVEL_EN_US = []map[string]any{
		{"id": 1, "event_level": "Critical", "health_level": "Unavailable", "health_score": []int{20, 40}},
		{"id": 2, "event_level": "Major", "health_level": "Fault", "health_score": []int{40, 60}},
		{"id": 3, "event_level": "Minor", "health_level": "Error", "health_score": []int{60, 80}},
		{"id": 4, "event_level": "Warning", "health_level": "Warning", "health_score": []int{80, 100}},
		{"id": 5, "event_level": "Indeterminate", "health_level": "Indeterminate", "health_score": []int{0, 0}},
		{"id": 6, "event_level": "Cleared", "health_level": "Healthy", "health_score": []int{100, 100}},
	}
)

// 创建事件模型(内部)
func (r *restHandler) CreateEventModelByIn(c *gin.Context) {
	logger.Debug("Handler CreateEventModelByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.CreateEventModel(c, visitor)
}

// 创建事件模型（外部）
func (r *restHandler) CreateEventModelByEx(c *gin.Context) {
	logger.Debug("Handler CreateEventModelByEx Start")

	// 校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.CreateEventModel(c, visitor)
}

// 创建事件模型
func (r *restHandler) CreateEventModel(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Create EventModel Start")
	ctx := rest.GetLanguageCtx(c)

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	//NOTE 接收请求参数并绑定
	emcrs := []interfaces.EventModelCreateRequest{}
	err := c.ShouldBindJSON(&emcrs)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateEventModelAuditObject("", ""), &httpErr.BaseError)

		rest.ReplyError(c, httpErr)
		return
	}

	now := time.Now().UnixMilli()
	ems := []interfaces.EventModel{}
	// var err error
	for _, emcr := range emcrs {
		//NOTE 生成事件模型对象分布式id
		modelID := xid.New().String()

		//NOTE 生成检测规则对象分布式id
		ruleID := xid.New().String()

		if emcr.EventModelType == "atomic" {
			emcr.DetectRule.DetectRuleID = ruleID
		} else {
			emcr.AggregateRule.AggregateRuleID = ruleID
		}

		//NOTE 当为实时订阅时，补充持久化的默认参数，方便前端回显。
		if emcr.EnableSubscribe == 1 {
			emcr.DefaultTimeWindow = interfaces.TimeInterval{
				Interval: 5,
				Unit:     "m",
			}
			emcr.EventTaskRequest.Schedule = interfaces.Schedule{
				Type:       "FIX_RATE",
				Expression: "5m",
			}
			emcr.EventTaskRequest.StorageConfig.IndexBase = interfaces.DEFAULT_INDEX_BASE
			emcr.EventTaskRequest.StorageConfig.DataViewName = interfaces.DEFAULT_DATA_VIEW_NAME
			emcr.EventTaskRequest.StorageConfig.DataViewID = interfaces.DEFAULT_DATA_VIEW_ID
		}

		//TODO 查询任务表，获取任务id
		var DownstreamDependentTask []string

		if len(emcr.DownstreamDependentModel) > 0 {
			//TODO: 根据模型id获取对应任务ID，定期执行才需要转换，事件驱动场景下不需要转换。
			// var task interfaces.EventTask
			for _, em_id := range emcr.DownstreamDependentModel {
				eventTask, _, err := r.ems.GetEventTaskByModelID(ctx, em_id)
				if err != nil {
					logger.Errorf("Find Task ID failed due to %v by event model id:%s", err, em_id)
					continue
				}
				DownstreamDependentTask = append(DownstreamDependentTask, eventTask.TaskID)
			}
		}

		if len(emcr.DownstreamDependentModel) != len(DownstreamDependentTask) {
			logger.Errorf("the count of Task of Event model not equal to the count of event model,please check event model task! ")
		}

		em := interfaces.EventModel{
			EventModelID:        modelID,
			EventModelName:      emcr.EventModelName,
			EventModelType:      emcr.EventModelType,
			EventModelTags:      emcr.EventModelTags,
			EventModelComment:   emcr.EventModelComment,
			DataSource:          emcr.DataSource,
			DataSourceName:      emcr.DataSourceName,
			DataSourceGroupName: emcr.DataSourceGroupName,
			DataSourceType:      emcr.DataSourceType,
			DefaultTimeWindow: interfaces.TimeInterval{
				Interval: emcr.DefaultTimeWindow.Interval,
				Unit:     emcr.DefaultTimeWindow.Unit,
			}, //解析时间窗口字符串
			DetectRule:               emcr.DetectRule,
			AggregateRule:            emcr.AggregateRule,
			DownstreamDependentModel: emcr.DownstreamDependentModel,
			Task: interfaces.EventTask{
				Schedule:                emcr.EventTaskRequest.Schedule,
				ExecuteParameter:        emcr.EventTaskRequest.ExecuteParameter,
				StorageConfig:           emcr.EventTaskRequest.StorageConfig,
				DispatchConfig:          emcr.EventTaskRequest.DispatchConfig,
				DownstreamDependentTask: DownstreamDependentTask,
				Creator:                 accountInfo,
			},
			IsActive:        emcr.IsActive,
			IsCustom:        1, //默认设置创建时为用户自定义模板
			EnableSubscribe: emcr.EnableSubscribe,
			Status:          emcr.Status,
			Creator:         accountInfo,
		}

		//NOTE: 此处做个兼容操作，将分析应用的ID 改为基于问题识别的algo_app_id来计算，
		//NOTE: 目前前端页面上还不能通过场景来区分算法，是通过开关来判断是否开启分析模型，所以需要后端来进行一下场景判断。
		//NOTE: 等到页面上支持按场景传递对应的算法id，再删除此段逻辑。
		//NOTE: 场景判断逻辑
		//NOTE: 溯源分析algo_app_id =  问题识别的algo_app_id + 1
		//NOTE: 问题收敛 algo_app_id =  问题识别的algo_app_id + 2
		algo_app_id, _ := strconv.Atoi(em.AggregateRule.AggregateAlgo)
		if em.AggregateRule.AnalysisAlgo["traceability_analysis"] != "" {
			em.AggregateRule.AnalysisAlgo["traceability_analysis"] = strconv.FormatInt(int64(algo_app_id+1), 10)
		}
		if em.AggregateRule.AnalysisAlgo["problem_convergence"] != "" {
			em.AggregateRule.AnalysisAlgo["problem_convergence"] = strconv.FormatInt(int64(algo_app_id+2), 10)
		}

		//NOTE 标签合法性
		for _, tag := range em.EventModelTags {
			if isInvalid := strings.ContainsAny(interfaces.NAME_INVALID_CHARACTER, tag); isInvalid {
				httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).
					WithErrorDetails("Event model tag contains special characters, such as /:?\\\"<>|：？‘’“”！《》#[]{}%&*$^!=.''")

				audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
					GenerateEventModelAuditObject("", ""), &httpErr.BaseError)

				rest.ReplyError(c, httpErr)
				return
			}
		}
		//NOTE 持久化任务参数校验
		err = validateEventTask(ctx, em.Task)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateEventModelAuditObject("", ""), &httpErr.BaseError)

			rest.ReplyError(c, httpErr)
			return
		}

		//NOTE: 兼容旧事件模型导入时无dataviewID
		if em.Task.StorageConfig.DataViewID == "" {
			em.Task.StorageConfig.DataViewID = fmt.Sprintf("__%s", em.Task.StorageConfig.IndexBase)
		}

		em.CreateTime = now
		em.UpdateTime = now
		em.DetectRule.CreateTime = now
		em.DetectRule.UpdateTime = now
		em.AggregateRule.CreateTime = now
		em.AggregateRule.UpdateTime = now
		ems = append(ems, em)
	}

	//NOTE 批量创建事件模型，返回特定的服务异常
	var atomicEventModels []interfaces.EventModel
	var aggregateEventModels []interfaces.EventModel
	var AnalysisEventModels []interfaces.EventModel
	for _, event_model := range ems {
		if event_model.EventModelType == "atomic" {
			//NOTE: 原子事件序列
			if len(event_model.DownstreamDependentModel) > 0 {
				atomicEventModels = append(atomicEventModels, event_model)
			} else {
				appendModels := []interfaces.EventModel{event_model}
				atomicEventModels = append(appendModels, atomicEventModels...)
			}
		} else {
			//NOTE: 溯源事件序列
			if event_model.AggregateRule.AggregateAlgo == interfaces.DEFAULT_AGGREGATE_TYPE_FOR_ROOT_CAUSE_ANALYSIS {
				AnalysisEventModels = append(AnalysisEventModels, event_model)
			} else if len(event_model.DownstreamDependentModel) > 0 {
				//NOTE: 问题事件序列
				aggregateEventModels = append(aggregateEventModels, event_model)
			} else {
				appendModels := []interfaces.EventModel{event_model}
				aggregateEventModels = append(appendModels, aggregateEventModels...)
			}
		}
	}

	//NOTE: 先导入原子事件
	//NOTE: 校验原子事件的数据源
	for index, event_model := range atomicEventModels {
		//NOTE 进行业务校验，
		//NOTE 包含名称重复，数据源存在，检测规则合理
		em, httpErr := r.ems.EventModelCreateValidate(ctx, event_model)
		if httpErr != nil {
			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateEventModelAuditObject("", ""), &httpErr.BaseError)

			rest.ReplyError(c, httpErr)
			return
		}

		//NOTE 更新新的数据源id和启动状态
		atomicEventModels[index].DataSource = em.DataSource
		atomicEventModels[index].IsActive = em.IsActive
		atomicEventModels[index].Status = em.Status
	}

	var atomicModelInfos = make([]map[string]any, len(atomicEventModels))
	if len(atomicEventModels) > 0 {
		atomicModelInfos, err = r.ems.CreateEventModels(ctx, atomicEventModels)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateEventModelAuditObject("", ""), &httpErr.BaseError)

			rest.ReplyError(c, httpErr)
			return
		}

		for _, modelInfo := range atomicModelInfos {
			model_name, _ := modelInfo["name"].(string)
			audit.NewWarnLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateEventModelAuditObject("", model_name), audit.SUCCESS, "")
		}
	}

	//NOTE: 再导入聚合事件模型，因为有可能某次导入中有依赖。
	//NOTE:校验聚合事件的数据源
	var aggreModelInfos = make([]map[string]any, len(aggregateEventModels))
	if len(aggregateEventModels) > 0 {
		for index, event_model := range aggregateEventModels {
			//NOTE 包含名称重复，数据源存在，检测规则合理
			em, httpErr := r.ems.EventModelCreateValidate(ctx, event_model)
			if httpErr != nil {
				audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
					GenerateEventModelAuditObject(em.EventModelID, em.EventModelName), &httpErr.BaseError)

				rest.ReplyError(c, httpErr)
				return
			}
			aggregateEventModels[index].DataSource = em.DataSource
		}

		aggreModelInfos, err = r.ems.CreateEventModels(ctx, aggregateEventModels)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateEventModelAuditObject("", ""), &httpErr.BaseError)

			rest.ReplyError(c, httpErr)
			return
		}

		for _, modelInfo := range aggreModelInfos {
			model_name, _ := modelInfo["name"].(string)
			audit.NewWarnLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateEventModelAuditObject("", model_name), audit.SUCCESS, "")
		}
	}

	//NOTE: 再导入溯源分析事件模型，因为有可能某次导入中有问题事件模型依赖。
	if len(AnalysisEventModels) > 0 {
		for index, event_model := range AnalysisEventModels {
			//NOTE 包含名称重复，数据源存在，检测规则合理
			em, httpErr := r.ems.EventModelCreateValidate(ctx, event_model)
			if httpErr != nil {
				audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
					GenerateEventModelAuditObject(em.EventModelID, em.EventModelName), &httpErr.BaseError)

				rest.ReplyError(c, httpErr)
				return
			}
			AnalysisEventModels[index].DataSource = em.DataSource
		}

		aggreModelInfos, err = r.ems.CreateEventModels(ctx, AnalysisEventModels)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateEventModelAuditObject("", ""), &httpErr.BaseError)

			rest.ReplyError(c, httpErr)
			return
		}

		for _, modelInfo := range aggreModelInfos {
			model_name, _ := modelInfo["name"].(string)
			audit.NewWarnLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateEventModelAuditObject("", model_name), audit.SUCCESS, "")
		}
	}

	// c.Writer.Header().Set("Location", "/api/mdl-data-model/v1/event-models/"+strconv.Itoa(int(modelID)))
	atomicModelInfos = append(atomicModelInfos, aggreModelInfos...)

	logger.Debug("Handler CreateEventModels Success")
	rest.ReplyOK(c, http.StatusCreated, atomicModelInfos)
}

// 更新事件模型(内部)
func (r *restHandler) UpdateEventModelByIn(c *gin.Context) {
	logger.Debug("Handler UpdateEventModelByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.UpdateEventModel(c, visitor)
}

// 更新事件模型（外部）
func (r *restHandler) UpdateEventModelByEx(c *gin.Context) {
	logger.Debug("Handler CreateMetricModelsByEx Start")

	// 校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.UpdateEventModel(c, visitor)
}

// 更新事件模型
func (r *restHandler) UpdateEventModel(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler UpdateEventModel Start")
	ctx := rest.GetLanguageCtx(c)

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	//接收绑定参数
	emur := interfaces.EventModelUpateRequest{
		EventModelID: "",
	}
	err := c.ShouldBindJSON(&emur)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateEventModelAuditObject("", emur.EventModelName), &httpErr.BaseError)

		rest.ReplyError(c, httpErr)
		return
	}

	err = c.ShouldBindUri(&emur)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateEventModelAuditObject("", emur.EventModelName), &httpErr.BaseError)

		rest.ReplyError(c, httpErr)
		return
	}

	//NOTE 进行常规校验
	//NOTE 名称合法校验
	// if isInvalid := strings.ContainsAny(interfaces.DATA_TAG_NAME_INVALID_CHARACTER, emur.EventModelName); isInvalid {
	// 	httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).
	// 		WithErrorDetails("Event Model name contains special characters, such as /:?\\\"<>|：？‘’“”！《》,#[]{}%&*$^!=.''")
	// 	rest.ReplyError(c, httpErr)
	// 	return
	// }
	//NOTE 说明合法校验，事件模型未做，保持统一
	// if isInvalid := strings.ContainsAny(interfaces.DATA_TAG_NAME_INVALID_CHARACTER, emur.EventModelComment); isInvalid {
	// 	httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).
	// 		WithErrorDetails("Event model comment contains special characters, such as /:?\\\"<>|：？‘’“”！《》,#[]{}%&*$^!=.''")
	// 	rest.ReplyError(c, httpErr)
	// 	return
	// }

	//NOTE 标签合法性
	for _, tag := range emur.EventModelTags {
		if isInvalid := strings.ContainsAny(interfaces.NAME_INVALID_CHARACTER, tag); isInvalid {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).
				WithErrorDetails("Event model comment contains special characters, such as /:?\\\"<>|：？‘’“”！《》,#[]{}%&*$^!=.''")

			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateEventModelAuditObject("", emur.EventModelName), &httpErr.BaseError)

			rest.ReplyError(c, httpErr)
			return
		}
	}

	task := interfaces.EventTask{
		Schedule:         emur.Schedule,
		ExecuteParameter: emur.ExecuteParameter,
		StorageConfig:    emur.StorageConfig,
		DispatchConfig:   emur.DispatchConfig,
	}
	err = validateEventTask(ctx, task)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateEventModelAuditObject("", emur.EventModelName), &httpErr.BaseError)

		rest.ReplyError(c, err)
		return
	}

	//NOTE: 兼容旧事件模型导入时无dataviewID
	if task.StorageConfig.DataViewID == "" {
		task.StorageConfig.DataViewID = fmt.Sprintf("__%s", task.StorageConfig.IndexBase)
	}

	//NOTE 进行业务校验
	httpErr := r.ems.EventModelUpdateValidate(ctx, emur)
	if httpErr != nil {
		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateEventModelAuditObject("", emur.EventModelName), &httpErr.BaseError)

		rest.ReplyError(c, httpErr)
		return
	}

	//NOTE 更新事件模型
	err = r.ems.UpdateEventModel(ctx, emur)
	if err != nil {
		httpErr = err.(*rest.HTTPError)
		if httpErr.BaseError.ErrorCode == derrors.EventModel_EventModelNotFound {
			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateEventModelAuditObject("", emur.EventModelName), &httpErr.BaseError)

			rest.ReplyError(c, httpErr)
			return
		} else {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).
				WithErrorDetails("update event model failed")

			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateEventModelAuditObject("", emur.EventModelName), &httpErr.BaseError)

			rest.ReplyError(c, httpErr)
			return
		}
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		GenerateEventModelAuditObject("", emur.EventModelName), "")

	logger.Debug("Handler UpdateEventModel Success")
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// NOTE 删除事件模型
func (r *restHandler) DeleteEventModels(c *gin.Context) {
	logger.Debug("Handler DeleteEventModels Start")
	ctx := rest.GetLanguageCtx(c)

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

	//接收绑定参数
	emdr := interfaces.EventModelDeleteRequest{}
	err = c.ShouldBindUri(&emdr)
	//NOTE: 处理请求参数异常
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateEventModelAuditObject(emdr.EventModelIDs, ""), &httpErr.BaseError)

		rest.ReplyError(c, httpErr)
		return
	}

	//NOTE 事件模型ID序列 -> []string{}
	eventModelIDs := common.StringToStringSlice(emdr.EventModelIDs)
	if len(eventModelIDs) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).
			WithErrorDetails("Type Conversion Failed:")

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateEventModelAuditObject(emdr.EventModelIDs, ""), &httpErr.BaseError)

		rest.ReplyError(c, httpErr)
		return
	}

	//NOTE 统一删除事件模型序列
	ems, err := r.ems.DeleteEventModels(ctx, eventModelIDs)
	if err != nil {
		//NOTE 处理未找到对象错误
		httpErr := err.(*rest.HTTPError)
		if httpErr.BaseError.ErrorCode == derrors.EventModel_EventModelNotFound {
			audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				GenerateEventModelAuditObject(emdr.EventModelIDs, ""), &httpErr.BaseError)
			rest.ReplyError(c, httpErr)
			return
		} else if httpErr.BaseError.ErrorCode == derrors.EventModel_RefByOther {
			audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				GenerateEventModelAuditObject(emdr.EventModelIDs, ""), &httpErr.BaseError)
			rest.ReplyError(c, httpErr)
			return
		} else {
			//NOTE 处理内部错误
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).
				WithErrorDetails("delete event model failed")

			audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				GenerateEventModelAuditObject(emdr.EventModelIDs, ""), &httpErr.BaseError)

			rest.ReplyError(c, httpErr)
			return
		}
	}

	//INFO 循环记录审计日志
	for _, model := range ems {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateEventModelAuditObject(model.EventModelID, model.EventModelName), audit.SUCCESS, "")
	}

	logger.Debug("Handler DeleteEventModels Success")
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 分页获取事件模型列表(内部)
func (r *restHandler) QueryEventModelsByIn(c *gin.Context) {
	logger.Debug("Handler QueryEventModelsByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.QueryEventModels(c, visitor)
}

// 分页获取事件模型列表（外部）
func (r *restHandler) QueryEventModelsByEx(c *gin.Context) {
	logger.Debug("Handler CreateMetricModelsByEx Start")

	// 校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.QueryEventModels(c, visitor)
}

// 查询事件模型列表
func (r *restHandler) QueryEventModels(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Query EventModels Start")
	ctx := rest.GetLanguageCtx(c)

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	emqr := interfaces.EventModelQueryRequest{
		IsActive:           "", // 设置默认值,方便后面拼接sql
		EnableSubscribe:    "", // 设置默认值,方便后面拼接sql
		Status:             "",
		IsCustom:           -1,            // 设置默认值,方便后面拼接sql
		Direction:          "desc",        // 设置默认值,方便后面拼接sql
		SortKey:            "update_time", // 设置默认值,方便后面拼接sql
		Limit:              10,            // 设置默认值,方便后面拼接sql
		Offset:             0,             // 设置默认值,方便后面拼接sql
		ScheduleSyncStatus: "",            // 设置默认值,方便后面拼接sql
		TaskStatus:         "",            // 设置默认值,方便后面拼接sql
		EventModelTag:      "",            // 设置默认值,方便后面拼接sql
		DetectType:         "",            // 设置默认值,方便后面拼接sql
		AggregateType:      "",            // 设置默认值,方便后面拼接sql
	}

	err := c.ShouldBindQuery(&emqr)
	if emqr.SortKey == "name" {
		emqr.SortKey = "event_model_name"
	}
	emqr.EventModelTag = strings.Trim(emqr.EventModelTag, " ")
	//NOTE ： 处理请求参数错误
	if err != nil {
		logger.Errorf("find event model failed,%#v", emqr)
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).
			WithErrorDetails("Binding Query Paramter Failed:" + err.Error())
		rest.ReplyError(c, httpErr)
		return
	}

	//NOTE : 查询事件模型
	eventModelList, total, err := r.ems.QueryEventModels(ctx, emqr)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		if httpErr.BaseError.ErrorCode == derrors.EventModel_InternalError {
			logger.Errorf("find event model failed,%#v", emqr)
			rest.ReplyError(c, httpErr)
			return
		}
		if httpErr.BaseError.ErrorCode == derrors.EventModel_EventModelNotFound {
			logger.Errorf("event model not found,%#v", emqr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	result := map[string]interface{}{"entries": eventModelList, "total_count": total}

	logger.Debug("Query EventModels Success")
	rest.ReplyOK(c, http.StatusOK, result)
}

// 按 id 获取事件模型对象信息(内部)
func (r *restHandler) QueryEventModelByIDByIn(c *gin.Context) {
	logger.Debug("Handler QueryEventModelByIDByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.QueryEventModelByID(c, visitor)
}

// 按 id 获取事件模型对象信息（外部）
func (r *restHandler) QueryEventModelByIDByEx(c *gin.Context) {
	logger.Debug("Handler CreateMetricModelsByEx Start")

	// 校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.QueryEventModelByID(c, visitor)
}

// 按 id 获取事件模型对象信息
func (r *restHandler) QueryEventModelByID(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler GetEventModel Start")
	ctx := rest.GetLanguageCtx(c)

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	emqr := interfaces.EventModelItemQueryRequest{}
	err := c.ShouldBindUri(&emqr)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())
		logger.Errorf("find event model failed,%#v", emqr)
		rest.ReplyError(c, httpErr)
		return
	}
	//FIXME: 此处需要改为批量联表查询。
	//NOTE 获取事件模型 by id
	ids := strings.Split(emqr.EventModelIDs, ",")
	var ems []interfaces.EventModel
	for _, id := range ids {
		em, httpErr := r.ems.GetEventModelByID(ctx, id)

		if httpErr != nil {
			if httpErr.BaseError.ErrorCode == derrors.EventModel_EventModelNotFound {
				// rest.ReplyError(c, httpErr)
				logger.Errorf("find event model failed,%#v", emqr)
				ems = append(ems, interfaces.EventModel{})
				continue
			} else {
				// httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).
				// WithErrorDetails("query event model failed")
				logger.Errorf("find event model failed,%#v", emqr)
				// rest.ReplyError(c, httpErr)
				ems = append(ems, interfaces.EventModel{})
				continue
			}
		}
		ems = append(ems, em)
	}

	//NOTE 统计有效的对象的个数
	var cnt = 0
	for _, rst := range ems {
		if rst.EventModelID != "" {
			cnt = cnt + 1
			continue
		}
	}
	//NOTE 一个有效的对象都没有
	if cnt == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, derrors.EventModel_EventModelNotFound).
			WithErrorDetails("query event model failed")
		logger.Errorf("find event model failed,%#v", emqr)
		rest.ReplyError(c, httpErr)
		return
	}
	logger.Debug("Query EventModel By ID Success")
	rest.ReplyOK(c, http.StatusOK, ems)
}

// 事件级别定义接口
func (r *restHandler) QueryEventLevel(c *gin.Context) {
	logger.Debug("Handler GetEventModel Start")
	ctx := rest.GetLanguageCtx(c)

	_, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	// 如果是中文环境,返回中文等级
	var event_level []map[string]any
	lang := rest.GetLanguageByCtx(ctx)
	//FIXME: save to db
	//NOTE: 返回不同语言的级别列表
	if lang == "zh-CN" {
		event_level = EVENT_MODEL_LEVEL_ZH_CN
	} else {
		event_level = EVENT_MODEL_LEVEL_EN_US
	}
	logger.Debug("Query Event Level Success")
	rest.ReplyOK(c, http.StatusOK, event_level)
}

// 更新任务的执行状态(内部)
func (r *restHandler) UpdateEventTaskStatusByIn(c *gin.Context) {
	logger.Debug("Handler UpdateEventTaskStatusByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.UpdateEventTaskStatus(c, visitor)
}

// 更新任务的执行状态（外部）
func (r *restHandler) UpdateEventTaskStatusByEx(c *gin.Context) {
	logger.Debug("Handler CreateMetricModelsByEx Start")

	// 校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.UpdateEventTaskStatus(c, visitor)
}

func (r *restHandler) UpdateEventTaskStatus(c *gin.Context, visitor rest.Visitor) {
	ctx := rest.GetLanguageCtx(c)

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	//获取参数字符串
	taskID := c.Param("id")

	//接收绑定参数
	task := interfaces.EventTask{}
	err := c.ShouldBindJSON(&task)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		rest.ReplyError(c, httpErr)
		return
	}
	task.TaskID = taskID

	//根据id修改信息
	httpErr := r.ems.UpdateEventTaskAttributes(ctx, task)
	if httpErr != nil {

		rest.ReplyError(c, httpErr)
		return
	}

	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 分页获取目标模型资源列表
func (r *restHandler) ListEventModelSrcs(c *gin.Context) {
	logger.Debug("ListEventModelSrcs Start")
	ctx := rest.GetLanguageCtx(c)

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

	emqr := interfaces.EventModelQueryRequest{
		IsActive:           "", // 设置默认值,方便后面拼接sql
		EnableSubscribe:    "", // 设置默认值,方便后面拼接sql
		Status:             "",
		IsCustom:           -1,     // 设置默认值,方便后面拼接sql
		Direction:          "asc",  // 设置默认值,方便后面拼接sql
		SortKey:            "name", // 设置默认值,方便后面拼接sql
		Limit:              50,     // 设置默认值,方便后面拼接sql
		Offset:             0,      // 设置默认值,方便后面拼接sql
		ScheduleSyncStatus: "",     // 设置默认值,方便后面拼接sql
		TaskStatus:         "",     // 设置默认值,方便后面拼接sql
		EventModelTag:      "",     // 设置默认值,方便后面拼接sql
		DetectType:         "",     // 设置默认值,方便后面拼接sql
		AggregateType:      "",     // 设置默认值,方便后面拼接sql
	}

	err = c.ShouldBindQuery(&emqr)
	if emqr.SortKey == "name" {
		emqr.SortKey = "event_model_name"
	}

	namePattern := c.Query(RESOURCES_KEYWOED)
	emqr.EventModelNamePattern = namePattern

	emqr.EventModelTag = strings.Trim(emqr.EventModelTag, " ")
	//NOTE ： 处理请求参数错误
	if err != nil {
		logger.Errorf("find event model failed,%#v", emqr)
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).
			WithErrorDetails("Binding Query Paramter Failed:" + err.Error())
		rest.ReplyError(c, httpErr)
		return
	}

	//NOTE : 查询事件模型
	eventModelList, total, err := r.ems.ListEventModelSrcs(ctx, emqr)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		if httpErr.BaseError.ErrorCode == derrors.EventModel_InternalError {
			logger.Errorf("find event model failed,%#v", emqr)
			rest.ReplyError(c, httpErr)
			return
		}
		if httpErr.BaseError.ErrorCode == derrors.EventModel_EventModelNotFound {
			logger.Errorf("event model not found,%#v", emqr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	result := map[string]interface{}{"entries": eventModelList, "total_count": total}
	rest.ReplyOK(c, http.StatusOK, result)
}
