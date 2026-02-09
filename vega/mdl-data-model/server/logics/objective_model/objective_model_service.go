// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package objective_model

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	"data-model/logics"
	"data-model/logics/metric_model"
	"data-model/logics/permission"
)

var (
	omServiceOnce sync.Once
	omService     interfaces.ObjectiveModelService
)

type objectiveModelService struct {
	appSetting *common.AppSetting
	db         *sql.DB
	dmja       interfaces.DataModelJobAccess
	iba        interfaces.IndexBaseAccess
	mms        interfaces.MetricModelService
	mmts       interfaces.MetricModelTaskService
	oma        interfaces.ObjectiveModelAccess
	ps         interfaces.PermissionService
}

func NewObjectiveModelService(appSetting *common.AppSetting) interfaces.ObjectiveModelService {
	omServiceOnce.Do(func() {
		omService = &objectiveModelService{
			appSetting: appSetting,
			db:         logics.DB,
			dmja:       logics.DMJA,
			iba:        logics.IBA,
			mms:        metric_model.NewMetricModelService(appSetting),
			mmts:       metric_model.NewMetricModelTaskService(appSetting),
			oma:        logics.OMA,
			ps:         permission.NewPermissionService(appSetting),
		}
	})
	return omService
}

func (oms *objectiveModelService) CheckObjectiveModelExistByID(ctx context.Context, modelID string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("校验目标模型[%v]的存在性", modelID))
	span.SetAttributes(
		attr.Key("model_id").String(modelID))
	defer span.End()

	modelName, exist, err := oms.oma.CheckObjectiveModelExistByID(ctx, modelID)
	if err != nil {
		logger.Errorf("CheckObjectiveModelExistByID error: %s", err.Error())

		span.SetStatus(codes.Error, fmt.Sprintf("按ID[%v]获取目标模型失败", modelID))
		// 记录处理的 sql 字符串
		o11y.Error(ctx, fmt.Sprintf("按ID[%v]获取目标模型失败: %v", modelID, err))

		return "", exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_ObjectiveModel_InternalError_CheckModelIfExistFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return modelName, exist, nil
}

func (oms *objectiveModelService) CheckObjectiveModelExistByName(ctx context.Context, modelName string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("校验目标模型[%v]的存在性", modelName))
	span.SetAttributes(
		attr.Key("model_name").String(modelName))
	defer span.End()

	modelID, exist, err := oms.oma.CheckObjectiveModelExistByName(ctx, modelName)
	if err != nil {
		logger.Errorf("CheckObjectiveModelExistByName error: %s", err.Error())

		span.SetStatus(codes.Error, fmt.Sprintf("按名称[%v]获取目标模型失败", modelName))
		// 记录处理的 sql 字符串
		o11y.Error(ctx, fmt.Sprintf("按名称[%v]获取目标模型失败: %v", modelName, err))

		return modelID, exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_ObjectiveModel_InternalError_CheckModelIfExistFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return modelID, exist, nil
}

// 废弃
func (oms *objectiveModelService) CreateObjectiveModel(ctx context.Context, objectiveModel interfaces.ObjectiveModel) (modelID string, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Create metric model")
	defer span.End()

	// 4. 在service层校验依赖的指标模型的存在性
	timeWindow := ""
	switch oConfig := objectiveModel.ObjectiveConfig.(type) {
	case interfaces.SLOObjective:
		// slo 校验good和total
		// 如果关联模型不存在，报错
		err = oms.checkMetricModelExists(ctx, []string{oConfig.GoodMetricModel.ID,
			oConfig.TotalMetricModel.ID})
		if err != nil {
			return "", err
		}
		timeWindow = fmt.Sprintf("%dd", *oConfig.Period)
	case interfaces.KPIObjective:
		// kpi 校验综合计算指标模型和附加计算的指标模型的存在性
		modelIDs := make([]string, 0)
		for _, model := range oConfig.ComprehensiveMetricModels {
			modelIDs = append(modelIDs, model.ID)
		}
		for _, model := range oConfig.AdditionalMetricModels {
			modelIDs = append(modelIDs, model.ID)
		}

		err = oms.checkMetricModelExists(ctx, modelIDs)
		if err != nil {
			return "", err
		}
	}

	// 若提交的模型id为空，生成分布式ID
	if objectiveModel.ModelID == "" {
		objectiveModel.ModelID = xid.New().String()
	}
	currentTime := time.Now().UnixMilli() // 目标模型的update_time是int类型
	objectiveModel.CreateTime = currentTime
	objectiveModel.UpdateTime = currentTime

	// 校验 请求体中的索引库类型是否存在
	err = oms.checkIndexBase(ctx, objectiveModel.Task.IndexBase)
	if err != nil {
		return "", err
	}

	// 生成分布式ID
	objectiveModel.Task.TaskID = xid.New().String()
	objectiveModel.Task.ModuleType = interfaces.MODULE_TYPE_OBJECTIVE_MODEL
	objectiveModel.Task.ModelID = objectiveModel.ModelID
	objectiveModel.Task.UpdateTime = currentTime
	if timeWindow != "" {
		objectiveModel.Task.TimeWindows = []string{timeWindow}
	}
	// 时间窗口对应的计划时间
	// 追溯时长转毫秒，用于计算计划时间
	durationV := int64(0)
	if objectiveModel.Task.RetraceDuration != "" {
		dur, err := common.ParseDuration(objectiveModel.Task.RetraceDuration, common.DurationDayHourRE, false)
		if err != nil {
			logger.Errorf("Failed to parse retrace duration, err: %v", err.Error())
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_RetraceDuration).
				WithErrorDetails(err.Error())
		}
		durationV = int64(dur / (time.Millisecond / time.Nanosecond))
	}
	planTime := time.Now().UnixNano()/int64(time.Millisecond/time.Nanosecond) - durationV

	objectiveModel.Task.PlanTime = planTime
	// 创建，创建的任务状态为创建中
	objectiveModel.Task.ScheduleSyncStatus = interfaces.SCHEDULE_SYNC_STATUS_FINISH

	span.SetAttributes(
		attr.Key("objective_model_id").String(objectiveModel.ModelID),
		attr.Key("objective_model_name").String(objectiveModel.ModelName),
	)

	// 0. 开始事务
	tx, err := oms.db.Begin()
	if err != nil {
		logger.Errorf("Begin transaction error: %s", err.Error())
		span.SetStatus(codes.Error, "事务开启失败")
		o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError_BeginTransactionFailed).
			WithErrorDetails(err.Error())
	}
	// 0.1 异常时
	defer func() {
		switch err {
		case nil:
			// 提交事务
			err = tx.Commit()
			if err != nil {
				logger.Errorf("CreateObjectiveModel Transaction Commit Failed:%v", err)
				span.SetStatus(codes.Error, "提交事务失败")
				o11y.Error(ctx, fmt.Sprintf("CreateObjectiveModel Transaction Commit Failed: %s", err.Error()))

			}
			logger.Infof("CreateObjectiveModel Transaction Commit Success:%v", objectiveModel.ModelName)
			o11y.Debug(ctx, fmt.Sprintf("CreateObjectiveModel Transaction Commit Success: %s", objectiveModel.ModelName))
		default:
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logger.Errorf("CreateObjectiveModel Transaction Rollback Error:%v", rollbackErr)
				span.SetStatus(codes.Error, "事务回滚失败")
				o11y.Error(ctx, fmt.Sprintf("CreateObjectiveModel Transaction Rollback Error: %s", err.Error()))
			}
		}
	}()

	// 1. 创建模型
	err = oms.oma.CreateObjectiveModel(ctx, tx, objectiveModel)
	if err != nil {
		logger.Errorf("CreateObjectiveModel error: %s", err.Error())
		span.SetStatus(codes.Error, "创建目标模型失败")

		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError).
			WithErrorDetails(err.Error())
	}

	// 2. 创建模型下的任务
	err = oms.mmts.CreateMetricTask(ctx, tx, *objectiveModel.Task)
	if err != nil {
		logger.Errorf("CreateMetricTasks error: %s", err.Error())
		span.SetStatus(codes.Error, "创建目标模型持久化任务失败")

		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError).
			WithErrorDetails(err.Error())
	}

	// 3. 请求 data-model-job 服务开启任务
	modelJobCfg := &interfaces.DataModelJobCfg{
		JobID:      objectiveModel.Task.TaskID,
		JobType:    interfaces.JOB_TYPE_SCHEDULE,
		ModuleType: interfaces.MODULE_TYPE_OBJECTIVE_MODEL,
		MetricTask: objectiveModel.Task,
		Schedule:   objectiveModel.Task.Schedule,
	}
	logger.Infof("objective model %s request data-model-job start job %d", objectiveModel.ModelName, objectiveModel.Task.TaskID)
	uncancelableCtx := context.WithoutCancel(ctx)
	go func() {
		err := oms.dmja.StartJob(uncancelableCtx, modelJobCfg)
		if err != nil {
			logger.Errorf("Start objective job[%s] failed: %s", modelJobCfg.JobID, err.Error())
		}
	}()

	span.SetStatus(codes.Ok, "")
	return objectiveModel.ModelID, nil
}

func (oms *objectiveModelService) CreateObjectiveModels(ctx context.Context, objectiveModels []*interfaces.ObjectiveModel, mode string) (modelIDs []string, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Create objective model")
	defer span.End()

	// 判断userid是否有创建指标模型的权限（策略决策）
	err = oms.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_OBJECTIVE_MODEL,
		ID:   interfaces.RESOURCE_ID_ALL,
	}, []string{interfaces.OPERATION_TYPE_CREATE})
	if err != nil {
		return nil, err
	}

	currentTime := time.Now().UnixMilli()
	for _, model := range objectiveModels {
		// 若提交的模型id为空，生成分布式ID
		if model.ModelID == "" {
			model.ModelID = xid.New().String()
		}
		err = oms.processForCreate(ctx, model, currentTime)
		if err != nil {
			return nil, err
		}
	}

	// 0. 开始事务
	tx, err := oms.db.Begin()
	if err != nil {
		logger.Errorf("Begin transaction error: %s", err.Error())
		span.SetStatus(codes.Error, "事务开启失败")
		o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError_BeginTransactionFailed).
			WithErrorDetails(err.Error())
	}
	// 0.1 异常时
	defer func() {
		switch err {
		case nil:
			// 提交事务
			err = tx.Commit()
			if err != nil {
				logger.Errorf("CreateObjectiveModel Transaction Commit Failed:%v", err)
				span.SetStatus(codes.Error, "提交事务失败")
				o11y.Error(ctx, fmt.Sprintf("CreateObjectiveModel Transaction Commit Failed: %s", err.Error()))

			}
			logger.Infof("CreateObjectiveModel Transaction Commit Success")
			o11y.Debug(ctx, "CreateObjectiveModel Transaction Commit Success")
		default:
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logger.Errorf("CreateObjectiveModel Transaction Rollback Error:%v", rollbackErr)
				span.SetStatus(codes.Error, "事务回滚失败")
				o11y.Error(ctx, fmt.Sprintf("CreateObjectiveModel Transaction Rollback Error: %s", err.Error()))
			}
		}
	}()

	createModels, updateModels, err := oms.handleObjectiveModelImportMode(ctx, tx, mode, objectiveModels)
	if err != nil {
		return nil, err
	}

	// 挨个调更新函数（其中校验对象的更新权限，若无，报错）
	for _, model := range updateModels {
		err = oms.UpdateObjectiveModel(ctx, tx, *model)
		if err != nil {
			return nil, err
		}
	}

	resrcs := make([]interfaces.Resource, 0)
	// 1. 创建模型
	for _, objectiveModel := range createModels {
		modelIDs = append(modelIDs, objectiveModel.ModelID)
		err = oms.oma.CreateObjectiveModel(ctx, tx, *objectiveModel)
		if err != nil {
			logger.Errorf("CreateObjectiveModel error: %s", err.Error())
			span.SetStatus(codes.Error, "创建目标模型失败")

			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError).
				WithErrorDetails(err.Error())
		}

		accountInfo := interfaces.AccountInfo{}
		if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
			accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
		}
		objectiveModel.Task.Creator = accountInfo

		// 2. 创建模型下的任务
		err = oms.mmts.CreateMetricTask(ctx, tx, *objectiveModel.Task)
		if err != nil {
			logger.Errorf("CreateMetricTasks error: %s", err.Error())
			span.SetStatus(codes.Error, "创建目标模型持久化任务失败")

			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError).
				WithErrorDetails(err.Error())
		}

		// 3. 请求 data-model-job 服务开启任务
		modelJobCfg := &interfaces.DataModelJobCfg{
			JobID:      objectiveModel.Task.TaskID,
			JobType:    interfaces.JOB_TYPE_SCHEDULE,
			ModuleType: interfaces.MODULE_TYPE_OBJECTIVE_MODEL,
			MetricTask: objectiveModel.Task,
			Schedule:   objectiveModel.Task.Schedule,
		}
		logger.Infof("objective model %s request data-model-job start job %d", objectiveModel.ModelName, objectiveModel.Task.TaskID)
		uncancelableCtx := context.WithoutCancel(ctx)
		go func() {
			err = oms.dmja.StartJob(uncancelableCtx, modelJobCfg)
			if err != nil {
				logger.Errorf("Start objective job[%s] failed: %s", modelJobCfg.JobID, err.Error())
			}
		}()

		resrcs = append(resrcs, interfaces.Resource{
			ID:   objectiveModel.ModelID,
			Type: interfaces.RESOURCE_TYPE_OBJECTIVE_MODEL,
			Name: objectiveModel.ModelName,
		})
	}

	// 注册资源策略
	err = oms.ps.CreateResources(ctx, resrcs, interfaces.COMMON_OPERATIONS)
	if err != nil {
		return nil, err
	}
	span.SetStatus(codes.Ok, "")
	return modelIDs, nil
}

func (oms *objectiveModelService) ListObjectiveModels(ctx context.Context, parameter interfaces.ObjectiveModelsQueryParams) ([]interfaces.ObjectiveModel, int, error) {
	listCtx, listSpan := ar_trace.Tracer.Start(ctx, "查询目标模型列表")

	//获取目标模型列表
	emptyObjectiveModels := make([]interfaces.ObjectiveModel, 0)
	objectiveModelArr, err := oms.oma.ListObjectiveModels(listCtx, parameter)
	if err != nil {
		logger.Errorf("ListObjectiveModels error: %s", err.Error())
		listSpan.SetStatus(codes.Error, "List objective models error")
		listSpan.End()

		return emptyObjectiveModels, 0, rest.NewHTTPError(listCtx, http.StatusInternalServerError,
			derrors.DataModel_ObjectiveModel_InternalError).WithErrorDetails(err.Error())
	}
	if len(objectiveModelArr) == 0 {
		return objectiveModelArr, 0, nil
	}

	// 处理资源id
	modelIDs := make([]string, 0)
	for _, m := range objectiveModelArr {
		modelIDs = append(modelIDs, m.ModelID)
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	matchResoucesMap, err := oms.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_OBJECTIVE_MODEL, modelIDs,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		return emptyObjectiveModels, 0, err
	}

	// 根据模型id获取任务信息
	modelTaskMap, err := oms.mmts.GetMetricTasksByModelIDs(listCtx, modelIDs)
	if err != nil {
		logger.Errorf("GetMetricTasksByModelIDs error: %s", err.Error())
		listSpan.SetStatus(codes.Error, fmt.Sprintf("Get metric model[%s] error: %v", modelIDs, err))
		listSpan.End()

		return emptyObjectiveModels, 0, rest.NewHTTPError(listCtx, http.StatusInternalServerError,
			derrors.DataModel_ObjectiveModel_InternalError_GetMetricTasksByModelIDsFailed).WithErrorDetails(err.Error())
	}
	listSpan.SetStatus(codes.Ok, "")
	listSpan.End()

	// 指标模型 id 换成名称，需要循环遍历查询。
	metricModelCtx, metricModelSpan := ar_trace.Tracer.Start(ctx, "Objective model's metric model id transfer to name")
	objectiveModels := make([]interfaces.ObjectiveModel, 0)
	for _, model := range objectiveModelArr {
		// 只留下有权限的模型
		if resrc, exist := matchResoucesMap[model.ModelID]; exist {
			// 处理模型任务
			task := modelTaskMap[model.ModelID]
			model.Task = &task
			if model.Task != nil {
				// 若存在任务，则把任务中的索引库类型换索引库名称
				// 获取索引库名称
				simpleIndexBases, err := oms.iba.GetSimpleIndexBasesByTypes(ctx, []string{model.Task.IndexBase})
				if err != nil {
					logger.Warnf("GetSimpleIndexBasesByTypes error: %s", err.Error())

					// return emptyObjectiveModels, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					// 	derrors.DataModel_ObjectiveModel_InternalError_GetSimpleIndexBasesByTypesFailed).WithErrorDetails(err.Error())
				}
				// 若索引库查询为空，则不赋值
				if len(simpleIndexBases) > 0 {
					// 遍历模型任务，把索引库名称赋值到任务上
					model.Task.IndexBaseName = simpleIndexBases[0].Name
				}
			}

			// 处理依赖的指标模型的名称
			// slo 和 kpi 的处理方式不同，需分开
			switch oConfig := model.ObjectiveConfig.(type) {
			case interfaces.SLOObjective:
				metricIDs := make([]string, 0)
				metricIDs = append(metricIDs, oConfig.GoodMetricModel.ID, oConfig.TotalMetricModel.ID)

				// 按指标模型id获取指标模型信息
				res, err := oms.mms.GetMetricModelSimpleInfosByIDs(ctx, metricIDs)
				if err != nil {
					return emptyObjectiveModels, 0, err
				}
				goodMetric, ok := res[metricIDs[0]]
				if !ok {
					// 查询时，如果指标模型已经被删除，则指标模型名称字段为空
					o11y.Warn(metricModelCtx, fmt.Sprintf("metric model %s not found", oConfig.GoodMetricModel.ID))
				} else {
					oConfig.GoodMetricModel.Name = goodMetric.ModelName
					oConfig.GoodMetricModel.UnitType = goodMetric.UnitType
					oConfig.GoodMetricModel.Unit = goodMetric.Unit
				}

				totalMetric, ok := res[metricIDs[1]]
				if !ok {
					// 查询时，如果指标模型已经被删除，则指标模型名称字段为空
					o11y.Warn(metricModelCtx, fmt.Sprintf("metric model %s not found", oConfig.TotalMetricModel.ID))
				} else {
					oConfig.TotalMetricModel.Name = totalMetric.ModelName
					oConfig.TotalMetricModel.UnitType = totalMetric.UnitType
					oConfig.TotalMetricModel.Unit = totalMetric.Unit
				}

			case interfaces.KPIObjective:
				metricIDs := make([]string, 0)
				for _, model := range oConfig.ComprehensiveMetricModels {
					metricIDs = append(metricIDs, model.ID)
				}
				for _, model := range oConfig.AdditionalMetricModels {
					metricIDs = append(metricIDs, model.ID)
				}

				// 按指标模型id获取指标模型信息
				res, err := oms.mms.GetMetricModelSimpleInfosByIDs(ctx, metricIDs)
				if err != nil {
					return emptyObjectiveModels, 0, err
				}
				for i, model := range oConfig.ComprehensiveMetricModels {
					metric, ok := res[model.ID]
					if !ok {
						// 查询时，如果指标模型已经被删除，则指标模型名称字段为空
						o11y.Warn(metricModelCtx, fmt.Sprintf("metric model %s not found", model.ID))
					} else {
						oConfig.ComprehensiveMetricModels[i].Name = metric.ModelName
					}
				}
				for i, model := range oConfig.AdditionalMetricModels {
					metric, ok := res[model.ID]
					if !ok {
						// 查询时，如果指标模型已经被删除，则指标模型名称字段为空
						o11y.Warn(metricModelCtx, fmt.Sprintf("metric model %s not found", model.ID))
					} else {
						oConfig.AdditionalMetricModels[i].Name = metric.ModelName
					}
				}
			}

			model.Operations = resrc.Operations // 用户当前有权限的操作
			objectiveModels = append(objectiveModels, model)
		}

	}
	metricModelSpan.SetStatus(codes.Ok, "")
	metricModelSpan.End()

	// limit = -1,则返回所有
	if parameter.Limit == -1 {
		return objectiveModels, len(objectiveModels), nil
	}
	// 分页
	// 检查起始位置是否越界
	if parameter.Offset < 0 || parameter.Offset >= len(objectiveModels) {
		return []interfaces.ObjectiveModel{}, 0, nil
	}
	// 计算结束位置
	end := parameter.Offset + parameter.Limit
	if end > len(objectiveModels) {
		end = len(objectiveModels)
	}

	listSpan.SetStatus(codes.Ok, "")
	return objectiveModels[parameter.Offset:end], len(objectiveModels), nil

}

func (oms *objectiveModelService) GetObjectiveModels(ctx context.Context, modelIDs []string) ([]interfaces.ObjectiveModel, error) {
	// 获取目标模型
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("查询目标模型[%s]信息", modelIDs))
	span.SetAttributes(attr.Key("model_ids").String(fmt.Sprintf("%v", modelIDs)))

	// 获取模型基本信息
	objectiveModelArr, err := oms.oma.GetObjectiveModelsByModelIDs(ctx, modelIDs)
	if err != nil {
		logger.Errorf("GetObjectiveModelsByModelIDs error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get objective model[%s] error: %v", modelIDs, err))
		span.End()

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_ObjectiveModel_InternalError_GetObjectiveModelsByModelIDsFailed).WithErrorDetails(err.Error())
	}

	if len(objectiveModelArr) != len(modelIDs) {
		errStr := fmt.Sprintf("Exists any models not found, expect model nums is [%d], actual models num is [%d]", len(modelIDs), len(objectiveModelArr))
		logger.Errorf(errStr)
		span.SetStatus(codes.Error, errStr)
		span.End()

		return nil, rest.NewHTTPError(ctx, http.StatusNotFound,
			derrors.DataModel_ObjectiveModel_ObjectiveModelNotFound).WithErrorDetails(errStr)
	}

	// 先获取资源序列
	matchResouces, err := oms.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_OBJECTIVE_MODEL, modelIDs,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		return nil, err
	}
	// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
	for _, mID := range modelIDs {
		if _, exist := matchResouces[mID]; !exist {
			return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for objective model's view_detail operation.")
		}
	}

	// 根据模型id获取任务信息
	modelTaskMap, err := oms.mmts.GetMetricTasksByModelIDs(ctx, modelIDs)
	if err != nil {
		logger.Errorf("GetMetricTasksByModelIDs error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get objective model[%s] error: %v", modelIDs, err))
		span.End()

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_ObjectiveModel_InternalError_GetMetricTasksByModelIDsFailed).WithErrorDetails(err.Error())
	}
	span.SetStatus(codes.Ok, "")
	span.End()

	// 指标模型 id 换成名称，需要循环遍历查询。
	metricModelCtx, metricModelSpan := ar_trace.Tracer.Start(ctx, "Objective model's metric model id transfer to name")
	objectiveModels := make([]interfaces.ObjectiveModel, 0)
	for _, model := range objectiveModelArr {
		// model 上附加 opreration 字段
		model.Operations = matchResouces[model.ModelID].Operations

		// 处理模型任务
		task, ok := modelTaskMap[model.ModelID]
		if !ok {
			model.Task = nil
		} else {
			model.Task = &task
			if model.Task != nil {
				// 若存在任务，则把任务中的索引库类型换索引库名称
				// 获取索引库名称
				simpleIndexBases, err := oms.iba.GetSimpleIndexBasesByTypes(ctx, []string{model.Task.IndexBase})
				if err != nil {
					logger.Errorf("GetSimpleIndexBasesByTypes error: %s", err.Error())

					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						derrors.DataModel_ObjectiveModel_InternalError_GetSimpleIndexBasesByTypesFailed).WithErrorDetails(err.Error())
				}
				// 若索引库查询为空，则不赋值
				if len(simpleIndexBases) > 0 {
					// 遍历模型任务，把索引库名称赋值到任务上
					model.Task.IndexBaseName = simpleIndexBases[0].Name
				}
			}
		}

		// 处理依赖的指标模型的名称
		// slo 和 kpi 的处理方式不同，需分开
		switch oConfig := model.ObjectiveConfig.(type) {
		case interfaces.SLOObjective:
			metricIDs := make([]string, 0)
			metricIDs = append(metricIDs, oConfig.GoodMetricModel.ID, oConfig.TotalMetricModel.ID)

			// 按指标模型id获取指标模型信息
			res, err := oms.mms.GetMetricModelSimpleInfosByIDs(ctx, metricIDs)
			if err != nil {
				return nil, err
			}
			goodMetric, ok := res[metricIDs[0]]
			if !ok {
				// 查询时，如果指标模型已经被删除，则指标模型名称字段为空
				o11y.Warn(metricModelCtx, fmt.Sprintf("metric model %s not found", oConfig.GoodMetricModel.ID))
			} else {
				oConfig.GoodMetricModel.Name = goodMetric.ModelName
				oConfig.GoodMetricModel.UnitType = goodMetric.UnitType
				oConfig.GoodMetricModel.Unit = goodMetric.Unit
			}

			totalMetric, ok := res[metricIDs[1]]
			if !ok {
				// 查询时，如果指标模型已经被删除，则指标模型名称字段为空
				o11y.Warn(metricModelCtx, fmt.Sprintf("metric model %s not found", oConfig.TotalMetricModel.ID))
			} else {
				oConfig.TotalMetricModel.Name = totalMetric.ModelName
				oConfig.TotalMetricModel.UnitType = totalMetric.UnitType
				oConfig.TotalMetricModel.Unit = totalMetric.Unit
			}

		case interfaces.KPIObjective:
			metricIDs := make([]string, 0)
			for _, model := range oConfig.ComprehensiveMetricModels {
				metricIDs = append(metricIDs, model.ID)
			}
			for _, model := range oConfig.AdditionalMetricModels {
				metricIDs = append(metricIDs, model.ID)
			}

			// 按指标模型id获取指标模型信息
			res, err := oms.mms.GetMetricModelSimpleInfosByIDs(ctx, metricIDs)
			if err != nil {
				return nil, err
			}
			for i, model := range oConfig.ComprehensiveMetricModels {
				metric, ok := res[model.ID]
				if !ok {
					// 查询时，如果指标模型已经被删除，则指标模型名称字段为空
					o11y.Warn(metricModelCtx, fmt.Sprintf("metric model %s not found", model.ID))
				} else {
					oConfig.ComprehensiveMetricModels[i].Name = metric.ModelName
				}
			}
			for i, model := range oConfig.AdditionalMetricModels {
				metric, ok := res[model.ID]
				if !ok {
					// 查询时，如果指标模型已经被删除，则指标模型名称字段为空
					o11y.Warn(metricModelCtx, fmt.Sprintf("metric model %s not found", model.ID))
				} else {
					oConfig.AdditionalMetricModels[i].Name = metric.ModelName
				}
			}
		}

		objectiveModels = append(objectiveModels, model)
	}
	metricModelSpan.SetStatus(codes.Ok, "")
	metricModelSpan.End()

	span.SetStatus(codes.Ok, "")
	return objectiveModels, nil
}
func (oms *objectiveModelService) UpdateObjectiveModel(ctx context.Context, tx *sql.Tx, objectiveModel interfaces.ObjectiveModel) (err error) {
	updateCtx, updateSpan := ar_trace.Tracer.Start(ctx, "Update objective model")
	defer updateSpan.End()

	// 判断userid是否有创建指标模型的权限（策略决策）
	err = oms.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_OBJECTIVE_MODEL,
		ID:   objectiveModel.ModelID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	// 4. 在service层校验依赖的指标模型的存在性
	timeWindow := ""
	switch oConfig := objectiveModel.ObjectiveConfig.(type) {
	case interfaces.SLOObjective:
		// slo 校验good和total
		// 如果关联模型不存在，报错
		err = oms.checkMetricModelExists(ctx, []string{oConfig.GoodMetricModel.ID,
			oConfig.TotalMetricModel.ID})
		if err != nil {
			return err
		}
		timeWindow = fmt.Sprintf("%dd", *oConfig.Period)
	case interfaces.KPIObjective:
		// kpi 校验综合计算指标模型和附加计算的指标模型的存在性
		modelIDs := make([]string, 0)
		for _, model := range oConfig.ComprehensiveMetricModels {
			modelIDs = append(modelIDs, model.ID)
		}
		for _, model := range oConfig.AdditionalMetricModels {
			modelIDs = append(modelIDs, model.ID)
		}

		err = oms.checkMetricModelExists(ctx, modelIDs)
		if err != nil {
			return err
		}
	}

	currentTime := time.Now().UnixMilli() // 目标模型的update_time是int类型
	objectiveModel.CreateTime = currentTime
	objectiveModel.UpdateTime = currentTime

	// 校验 请求体中的索引库类型是否存在
	err = oms.checkIndexBase(ctx, objectiveModel.Task.IndexBase)
	if err != nil {
		return err
	}

	updateSpan.SetAttributes(
		attr.Key("objective_model_id").String(objectiveModel.ModelID),
		attr.Key("objective_model_name").String(objectiveModel.ModelName))

	if tx == nil {
		// 0. 开始事务
		tx, err = oms.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			updateSpan.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

			return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("UpdateMetricModel Transaction Commit Failed:%v", err)
					updateSpan.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(updateCtx, fmt.Sprintf("UpdateMetricModel Transaction Commit Failed: %s", err.Error()))

				}
				logger.Infof("UpdateMetricModel Transaction Commit Success:%v", objectiveModel.ModelName)
				o11y.Debug(updateCtx, fmt.Sprintf("UpdateMetricModel Transaction Commit Success: %s", objectiveModel.ModelName))
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("UpdateMetricModel Transaction Rollback Error:%v", rollbackErr)
					updateSpan.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(updateCtx, fmt.Sprintf("UpdateMetricModel Transaction Rollback Error: %s", err.Error()))
				}
			}
		}()
	}

	// 更新模型信息
	err = oms.oma.UpdateObjectiveModel(updateCtx, tx, objectiveModel)
	if err != nil {
		logger.Errorf("metricModel error: %s", err.Error())
		updateSpan.SetStatus(codes.Error, "修改目标模型失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError).
			WithErrorDetails(err.Error())
	}

	// 更新操作
	if timeWindow != "" {
		objectiveModel.Task.TimeWindows = []string{timeWindow}
	}
	objectiveModel.Task.ModuleType = interfaces.MODULE_TYPE_OBJECTIVE_MODEL
	objectiveModel.Task.ModelID = objectiveModel.ModelID
	objectiveModel.Task.UpdateTime = currentTime
	objectiveModel.Task.ScheduleSyncStatus = interfaces.SCHEDULE_SYNC_STATUS_FINISH
	// 修改任务
	err = oms.mmts.UpdateMetricTask(updateCtx, tx, *objectiveModel.Task)
	if err != nil {
		logger.Errorf("UpdateMetricTask error: %s", err.Error())
		updateSpan.SetStatus(codes.Error, "修改目标模型持久化任务失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError).
			WithErrorDetails(err.Error())
	}

	// 请求 data-model-job 更新任务
	newMetricJobCfg := &interfaces.DataModelJobCfg{
		JobID:      objectiveModel.Task.TaskID,
		JobType:    interfaces.JOB_TYPE_SCHEDULE,
		ModuleType: interfaces.MODULE_TYPE_OBJECTIVE_MODEL,
		MetricTask: objectiveModel.Task,
		Schedule:   objectiveModel.Task.Schedule,
	}
	uncancelableCtx := context.WithoutCancel(ctx)
	go func() {
		err = oms.dmja.UpdateJob(uncancelableCtx, newMetricJobCfg)
		if err != nil {
			logger.Errorf("Update objective model job[%s] failed, %s", newMetricJobCfg.JobID, err.Error())
		}
	}()

	// 请求更新资源名称的接口，更新资源的名称
	if objectiveModel.IfNameModify {
		err = oms.ps.UpdateResource(ctx, interfaces.Resource{
			ID:   objectiveModel.ModelID,
			Type: interfaces.RESOURCE_TYPE_OBJECTIVE_MODEL,
			Name: objectiveModel.ModelName,
		})
		if err != nil {
			return err
		}
	}

	updateSpan.SetStatus(codes.Ok, "")
	return nil
}

func (oms *objectiveModelService) DeleteObjectiveModels(ctx context.Context, modelIDs []string) (int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete objective models")
	defer span.End()

	// 先获取资源序列 fmt.Sprintf("%s%s", interfaces.METRIC_MODEL_RESOURCE_ID_PREFIX, metricModel.ModelID),
	matchResouces, err := oms.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_OBJECTIVE_MODEL, modelIDs,
		[]string{interfaces.OPERATION_TYPE_DELETE}, false)
	if err != nil {
		return 0, err
	}
	// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
	for _, mID := range modelIDs {
		if _, exist := matchResouces[mID]; !exist {
			return 0, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for objective model's delete operation.")
		}
	}

	// 按modelid获取任务id列表
	taskIDs, err := oms.mmts.GetMetricTaskIDsByModelIDs(ctx, modelIDs)
	if err != nil {
		return 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_ObjectiveModel_InternalError).
			WithErrorDetails(err.Error())
	}

	// 请求data-model-job 批量停止任务
	if len(taskIDs) > 0 {
		// 请求data-model-job服务批量停止实际运行的任务
		if err = oms.dmja.StopJobs(ctx, taskIDs); err != nil {
			span.SetStatus(codes.Error, "Stop objective model jobs failed")
			return 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_StopJobFailed).WithErrorDetails(err.Error())
		}
	}

	// 0. 开始事务
	tx, err := oms.db.Begin()
	if err != nil {
		logger.Errorf("Begin transaction error: %s", err.Error())
		span.SetStatus(codes.Error, "事务开启失败")
		o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

		return 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError_BeginTransactionFailed).
			WithErrorDetails(err.Error())
	}
	// 0.1 异常时
	defer func() {
		switch err {
		case nil:
			// 提交事务
			err = tx.Commit()
			if err != nil {
				logger.Errorf("DeleteObjectiveModels Transaction Commit Failed:%v", err)
				span.SetStatus(codes.Error, "提交事务失败")
				o11y.Error(ctx, fmt.Sprintf("DeleteObjectiveModels Transaction Commit Failed: %s", err.Error()))

			}
			logger.Infof("DeleteObjectiveModels Transaction Commit Success:%v", modelIDs)
			o11y.Debug(ctx, fmt.Sprintf("DeleteObjectiveModels Transaction Commit Success: %v", modelIDs))
		default:
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logger.Errorf("DeleteObjectiveModels Transaction Rollback Error:%v", rollbackErr)
				span.SetStatus(codes.Error, "事务回滚失败")
				o11y.Error(ctx, fmt.Sprintf("DeleteObjectiveModels Transaction Rollback Error: %s", err.Error()))
			}
		}
	}()

	// 删除指标模型
	rowsAffect, err := oms.oma.DeleteObjectiveModels(ctx, tx, modelIDs)
	span.SetAttributes(attr.Key("rows_affect").Int64(rowsAffect))
	if err != nil {
		logger.Errorf("DeleteObjectiveModels error: %s", err.Error())
		span.SetStatus(codes.Error, "删除目标模型失败")

		return rowsAffect, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_ObjectiveModel_InternalError).WithErrorDetails(err.Error())
	}
	logger.Infof("DeleteObjectiveModels: Rows affected is %v, request delete modelIDs is %v!", rowsAffect, len(modelIDs))
	if rowsAffect != int64(len(modelIDs)) {
		logger.Warnf("Delete models number %v not equal requerst models number %v!", rowsAffect, len(modelIDs))

		o11y.Warn(ctx, fmt.Sprintf("Delete models number %v not equal requerst models number %v!", rowsAffect, len(modelIDs)))
	}

	// 删除指标模型下的任务
	err = oms.mmts.DeleteMetricTaskByTaskIDs(ctx, tx, taskIDs)
	if err != nil {
		logger.Errorf("DeleteMetricTaskByTaskIDs error: %s", err.Error())
		span.SetStatus(codes.Error, "批量删除目标模型持久化任务失败")
		return 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_ObjectiveModel_InternalError).
			WithErrorDetails(err.Error())
	}

	//  清除资源策略
	err = oms.ps.DeleteResources(ctx, interfaces.RESOURCE_TYPE_OBJECTIVE_MODEL, modelIDs)
	if err != nil {
		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return rowsAffect, nil
}

func (oms *objectiveModelService) checkMetricModelExists(ctx context.Context, metricIDs []string) error {
	res, err := oms.mms.GetMetricModelSimpleInfosByIDs(ctx, metricIDs)
	if err != nil {
		return err
	}
	unit := ""
	for i, id := range metricIDs {
		metricModel, ok := res[id]

		// 如果关联模型不存在，报错
		if !ok {
			errDetails := fmt.Sprintf("The metric model with id [%s] was not found!", id)
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)

			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_MetricModelNotFound).
				WithErrorDetails(errDetails)
		}
		if i == 0 {
			unit = metricModel.Unit
		}

		if unit != metricModel.Unit {
			errDetails := fmt.Sprintf("Exist some metric model's unit[%s] with whose id is [%s] and name is [%s] was not equal others[%s]!",
				metricModel.Unit, id, metricModel.ModelName, unit)
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)

			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_AssociateMetricModelsUnit_Different).
				WithErrorDetails(errDetails)
		}

	}
	return nil
}

func (oms *objectiveModelService) checkIndexBase(ctx context.Context, indexBaseType string) error {
	// 校验 请求体中的数据视图是否存在
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("获取索引库[%s]信息", indexBaseType))
	span.SetAttributes(attr.Key("index_base_type").String(indexBaseType))
	defer span.End()

	indexbase, err := oms.iba.GetSimpleIndexBasesByTypes(ctx, []string{indexBaseType})
	// 记录模块调用返回日志
	o11y.Info(ctx, fmt.Sprintf("GetSimpleIndexBasesByTypes, request: %s; response: {indexbases: %v, error: %v}",
		indexBaseType, indexbase, err))

	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_ObjectiveModel_InternalError_GetSimpleIndexBasesByTypesFailed).WithErrorDetails(err.Error())

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get index base [%s] error: %s. %v", indexBaseType,
			httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		// 获取视图的span状态设置
		span.SetStatus(codes.Error, fmt.Sprintf("Get index base [%s] error", indexBaseType))

		return httpErr
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 指标模型依赖的视图校验
func (oms *objectiveModelService) processForCreate(ctx context.Context, objectiveModel *interfaces.ObjectiveModel, currentTime int64) error {

	ctx, span := ar_trace.Tracer.Start(ctx, "Create metric models")
	defer span.End()

	// 4. 在service层校验依赖的指标模型的存在性
	timeWindow := ""
	switch oConfig := objectiveModel.ObjectiveConfig.(type) {
	case interfaces.SLOObjective:
		// slo 校验good和total
		// 如果关联模型不存在，报错
		err := oms.checkMetricModelExists(ctx, []string{oConfig.GoodMetricModel.ID,
			oConfig.TotalMetricModel.ID})
		if err != nil {
			return err
		}
		timeWindow = fmt.Sprintf("%dd", *oConfig.Period)
	case interfaces.KPIObjective:
		// kpi 校验综合计算指标模型和附加计算的指标模型的存在性
		modelIDs := make([]string, 0)
		for _, model := range oConfig.ComprehensiveMetricModels {
			modelIDs = append(modelIDs, model.ID)
		}
		for _, model := range oConfig.AdditionalMetricModels {
			modelIDs = append(modelIDs, model.ID)
		}

		err := oms.checkMetricModelExists(ctx, modelIDs)
		if err != nil {
			return err
		}
	}

	objectiveModel.CreateTime = currentTime
	objectiveModel.UpdateTime = currentTime

	// 校验 请求体中的索引库类型是否存在
	err := oms.checkIndexBase(ctx, objectiveModel.Task.IndexBase)
	if err != nil {
		return err
	}

	// 生成分布式ID
	objectiveModel.Task.TaskID = xid.New().String()
	objectiveModel.Task.ModuleType = interfaces.MODULE_TYPE_OBJECTIVE_MODEL
	objectiveModel.Task.ModelID = objectiveModel.ModelID
	objectiveModel.Task.UpdateTime = currentTime
	if timeWindow != "" {
		objectiveModel.Task.TimeWindows = []string{timeWindow}
	}
	// 时间窗口对应的计划时间
	// 追溯时长转毫秒，用于计算计划时间
	durationV := int64(0)
	if objectiveModel.Task.RetraceDuration != "" {
		dur, err := common.ParseDuration(objectiveModel.Task.RetraceDuration, common.DurationDayHourRE, false)
		if err != nil {
			logger.Errorf("Failed to parse retrace duration, err: %v", err.Error())
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_RetraceDuration).
				WithErrorDetails(err.Error())
		}
		durationV = int64(dur / (time.Millisecond / time.Nanosecond))
	}
	planTime := time.Now().UnixNano()/int64(time.Millisecond/time.Nanosecond) - durationV

	objectiveModel.Task.PlanTime = planTime
	// 创建，创建的任务状态为创建中
	objectiveModel.Task.ScheduleSyncStatus = interfaces.SCHEDULE_SYNC_STATUS_FINISH

	span.SetAttributes(
		attr.Key("objective_model_id").String(objectiveModel.ModelID),
		attr.Key("objective_model_name").String(objectiveModel.ModelName),
	)
	span.SetStatus(codes.Ok, "")
	return nil
}

func (oms *objectiveModelService) handleObjectiveModelImportMode(ctx context.Context, tx *sql.Tx, mode string,
	models []*interfaces.ObjectiveModel) ([]*interfaces.ObjectiveModel, []*interfaces.ObjectiveModel, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "objective model import mode logic")
	defer span.End()

	createModels := []*interfaces.ObjectiveModel{}
	updateModels := []*interfaces.ObjectiveModel{}

	// 3. 校验 若模型的id不为空，则用请求体的id与现有模型ID的重复性
	for i := 0; i < len(models); i++ {
		createModels = append(createModels, models[i])
		idExist := false
		_, idExist, err := oms.CheckObjectiveModelExistByID(ctx, models[i].ModelID)
		if err != nil {
			return createModels, updateModels, err
		}

		// 校验 请求体与现有模型名称的重复性
		existID, nameExist, err := oms.CheckObjectiveModelExistByName(ctx, models[i].ModelName)
		if err != nil {
			return createModels, updateModels, err
		}

		// 根据mode来区别，若是ignore，就从结果集中忽略，若是overwrite，就调用update，若是normal就报错。
		if idExist || nameExist {
			switch mode {
			case interfaces.ImportMode_Normal:
				if idExist {
					errDetails := fmt.Sprintf("The objective model with id [%s] already exists!", models[i].ModelID)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return createModels, updateModels, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_ModelIDExisted).
						WithErrorDetails(errDetails)
				}

				if nameExist {
					errDetails := fmt.Sprintf("objective model name '%s' already exists", models[i].ModelName)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return createModels, updateModels, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_ObjectiveModel_ModelNameExisted).
						WithDescription(map[string]any{"ModelName": models[i].ModelName}).
						WithErrorDetails(errDetails)
				}

			case interfaces.ImportMode_Ignore:
				// 存在重复的就跳过
				// 从create数组中删除
				createModels = createModels[:len(createModels)-1]
			case interfaces.ImportMode_Overwrite:
				if idExist && nameExist {
					// 如果 id 和名称都存在，但是存在的名称对应的视图 id 和当前视图 id 不一样，则报错
					if existID != models[i].ModelID {
						errDetails := fmt.Sprintf("ObjectiveModel ID '%s' and name '%s' already exist, but the exist model id is '%s'",
							models[i].ModelID, models[i].ModelID, existID)
						logger.Error(errDetails)
						span.SetStatus(codes.Error, errDetails)
						return createModels, updateModels, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_ObjectiveModel_ModelNameExisted).
							WithErrorDetails(errDetails)
					} else {
						// 如果 id 和名称、度量名称都存在，存在的名称对应的模型 id 和当前模型 id 一样，则覆盖更新
						// 从create数组中删除, 放到更新数组中
						createModels = createModels[:len(createModels)-1]
						updateModels = append(updateModels, models[i])
					}
				}

				// id 已存在，且名称不存在，覆盖更新
				if idExist && !nameExist {
					// 从create数组中删除, 放到更新数组中
					createModels = createModels[:len(createModels)-1]
					updateModels = append(updateModels, models[i])
				}

				// 如果 id 不存在，name 存在，报错
				if !idExist && nameExist {
					errDetails := fmt.Sprintf("ObjectiveModel ID '%s' does not exist, but name '%s' already exists",
						models[i].ModelID, models[i].ModelName)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return createModels, updateModels, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_ObjectiveModel_ModelNameExisted).
						WithErrorDetails(errDetails)
				}

				// 如果 id 不存在，name不存在，度量名称不存在，不需要做什么，创建
				// if !idExist && !nameExist {}
			}
		}
	}
	span.SetStatus(codes.Ok, "")
	return createModels, updateModels, nil
}

func (oms *objectiveModelService) ListObjectiveModelSrcs(ctx context.Context, parameter interfaces.ObjectiveModelsQueryParams) ([]interfaces.Resource, int, error) {
	listCtx, listSpan := ar_trace.Tracer.Start(ctx, "查询目标模型实例列表")
	defer listSpan.End()

	//获取目标模型列表
	empty := []interfaces.Resource{}
	models, err := oms.oma.ListObjectiveModels(listCtx, parameter)
	if err != nil {
		logger.Errorf("ListObjectiveModels error: %s", err.Error())
		listSpan.SetStatus(codes.Error, "List objective models error")
		listSpan.End()

		return empty, 0, rest.NewHTTPError(listCtx, http.StatusInternalServerError,
			derrors.DataModel_ObjectiveModel_InternalError).WithErrorDetails(err.Error())
	}

	// 处理资源id
	modelIDs := make([]string, 0)
	for _, m := range models {
		modelIDs = append(modelIDs, m.ModelID)
	}
	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	matchResoucesMap, err := oms.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_OBJECTIVE_MODEL, modelIDs,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		return empty, 0, err
	}

	// 遍历对象
	results := make([]interfaces.Resource, 0)
	for _, model := range models {
		if _, exist := matchResoucesMap[model.ModelID]; exist {
			results = append(results, interfaces.Resource{
				ID:   model.ModelID,
				Type: interfaces.RESOURCE_TYPE_OBJECTIVE_MODEL,
				Name: model.ModelName,
			})
		}
	}

	// limit = -1,则返回所有
	if parameter.Limit == -1 {
		return results, len(results), nil
	}

	// limit = -1,则返回所有
	if parameter.Limit == -1 {
		return results, len(results), nil
	}

	// 分页
	// 检查起始位置是否越界
	if parameter.Offset < 0 || parameter.Offset >= len(results) {
		return nil, 0, nil
	}
	// 计算结束位置
	end := parameter.Offset + parameter.Limit
	if end > len(results) {
		end = len(results)
	}

	listSpan.SetStatus(codes.Ok, "")
	return results[parameter.Offset:end], len(results), nil
}
