package metric_model

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
	"data-model/logics/data_view"
	"data-model/logics/permission"
)

var (
	mmServiceOnce sync.Once
	mmService     interfaces.MetricModelService
)

type metricModelService struct {
	appSetting *common.AppSetting
	ps         interfaces.PermissionService
	db         *sql.DB
	dmja       interfaces.DataModelJobAccess
	iba        interfaces.IndexBaseAccess
	dvs        interfaces.DataViewService
	mma        interfaces.MetricModelAccess
	mmga       interfaces.MetricModelGroupAccess
	mmts       interfaces.MetricModelTaskService
	ua         interfaces.UniqueryAccess
}

func NewMetricModelService(appSetting *common.AppSetting) interfaces.MetricModelService {
	mmServiceOnce.Do(func() {
		mmService = &metricModelService{
			appSetting: appSetting,
			db:         logics.DB,
			dmja:       logics.DMJA,
			iba:        logics.IBA,
			dvs:        data_view.NewDataViewService(appSetting),
			mma:        logics.MMA,
			mmga:       logics.MMGA,
			mmts:       NewMetricModelTaskService(appSetting),
			ps:         permission.NewPermissionService(appSetting),
			ua:         logics.UA,
		}
	})
	return mmService
}

// 创建单个指标模型 废弃
// func (mms *metricModelService) CreateMetricModel(ctx context.Context, metricModel interfaces.MetricModel) (modelID string, err error) {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "Create metric model")
// 	defer span.End()

// 	// 判断userid是否有创建指标模型的权限（策略决策）
// 	err = mms.ps.CheckPermission(ctx, interfaces.Resource{
// 		Type: interfaces.RESOURCE_TYPE_METRIC_MODEL,
// 		ID:   interfaces.RESOURCE_ID_ALL,
// 	}, []string{interfaces.OPERATION_TYPE_CREATE})
// 	if err != nil {
// 		return "", err
// 	}

// 	switch metricModel.DataSource.Type {
// 	case interfaces.DATA_SOURCE_DATA_VIEW:
// 		// 请求体中的数据视图是否存在
// 		_, err := mms.checkDataView(ctx, metricModel.DataSource.ID)
// 		if err != nil {
// 			return "", err
// 		}
// 		// dsl 时校验 top_hits 的度量字段的字段类型是否为数值，其他情况不校验。
// 		if metricModel.IfContainTopHits && metricModel.QueryType == interfaces.DSL {
// 			// 前面校验过dsl的数据源类型是data_view,所以这里可以直接使用
// 			err = mms.checkMeasureFieldType(ctx, metricModel.DataSource.ID, metricModel.MeasureField)
// 			if err != nil {
// 				return "", err
// 			}
// 		}

// 		// 计算公式有效性检查，请求 uniquery，与指标数据预览同一个接口
// 		err = mms.CheckFormula(ctx, metricModel)
// 		if err != nil {
// 			return "", err
// 		}
// 	case interfaces.DATA_SOURCE_VEGA_LOGIC_VIEW:
// 		err = mms.checkSQLView(ctx, metricModel, nil)
// 		if err != nil {
// 			return "", err
// 		}

// 		// 计算公式有效性检查，请求 uniquery，与指标数据预览同一个接口
// 		err = mms.CheckSqlFormulaConfig(ctx, metricModel)
// 		if err != nil {
// 			return "", err
// 		}
// 	}

// 	// 生成分布式ID
// 	if metricModel.ModelID == "" {
// 		metricModel.ModelID = xid.New().String()
// 	}
// 	currentTime := time.Now().UnixMilli()

// 	metricModel.UpdateTime = currentTime
// 	metricModel.CreateTime = currentTime

// 	// 度量名称为空，则赋值模型id
// 	if metricModel.MeasureName == "" {
// 		metricModel.MeasureName = fmt.Sprintf("%s%s", interfaces.MEASURE_PREFIX, metricModel.ModelID)
// 	}

// 	// 把任务id、模型id和更新时间赋值到任务对象中
// 	if metricModel.Task != nil {
// 		// 校验 请求体中的索引库类型是否存在
// 		err := mms.checkIndexBase(ctx, metricModel.Task.IndexBase)
// 		if err != nil {
// 			return "", err
// 		}

// 		// 生成分布式ID
// 		taskID := xid.New().String()
// 		metricModel.Task.TaskID = taskID

// 		metricModel.Task.ModuleType = interfaces.MODULE_TYPE_METRIC_MODEL
// 		metricModel.Task.ModelID = metricModel.ModelID
// 		metricModel.Task.MeasureName = metricModel.MeasureName
// 		metricModel.Task.UpdateTime = currentTime
// 		// 时间窗口对应的计划时间
// 		// 追溯时长转毫秒，用于计算计划时间
// 		durationV := int64(0)
// 		if metricModel.Task.RetraceDuration != "" {
// 			dur, err := common.ParseDuration(metricModel.Task.RetraceDuration, common.DurationDayHourRE, false)
// 			if err != nil {
// 				logger.Errorf("Failed to parse retrace duration, err: %v", err.Error())
// 				return "", rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_RetraceDuration).
// 					WithErrorDetails(err.Error())
// 			}
// 			durationV = int64(dur / (time.Millisecond / time.Nanosecond))
// 		}
// 		planTime := time.Now().UnixNano()/int64(time.Millisecond/time.Nanosecond) - durationV

// 		metricModel.Task.PlanTime = planTime
// 		// 创建，创建的任务状态为完成
// 		metricModel.Task.ScheduleSyncStatus = interfaces.SCHEDULE_SYNC_STATUS_FINISH
// 	}

// 	span.SetAttributes(
// 		attr.Key("metric_model_id").String(metricModel.ModelID),
// 		attr.Key("metric_model_name").String(metricModel.ModelName),
// 		attr.Key("group_name").String(metricModel.GroupName),
// 	)

// 	// 0. 开始事务
// 	tx, err := mms.db.Begin()
// 	if err != nil {
// 		logger.Errorf("Begin transaction error: %s", err.Error())
// 		span.SetStatus(codes.Error, "事务开启失败")
// 		o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

// 		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModel_InternalError_BeginTransactionFailed).
// 			WithErrorDetails(err.Error())
// 	}
// 	// 0.1 异常时
// 	defer func() {
// 		switch err {
// 		case nil:
// 			// 提交事务
// 			err = tx.Commit()
// 			if err != nil {
// 				logger.Errorf("CreateMetricModel Transaction Commit Failed:%v", err)
// 				span.SetStatus(codes.Error, "提交事务失败")
// 				o11y.Error(ctx, fmt.Sprintf("CreateMetricModel Transaction Commit Failed: %s", err.Error()))

// 			}
// 			logger.Infof("CreateMetricModel Transaction Commit Success:%v", metricModel.ModelName)
// 			o11y.Debug(ctx, fmt.Sprintf("CreateMetricModel Transaction Commit Success: %s", metricModel.ModelName))
// 		default:
// 			rollbackErr := tx.Rollback()
// 			if rollbackErr != nil {
// 				logger.Errorf("CreateMetricModel Transaction Rollback Error:%v", rollbackErr)
// 				span.SetStatus(codes.Error, "事务回滚失败")
// 				o11y.Error(ctx, fmt.Sprintf("CreateMetricModel Transaction Rollback Error: %s", err.Error()))
// 			}
// 		}
// 	}()

// 	//创建之前需要先更新一下group_ID
// 	metricGroup, err := mms.RetriveGroupIDByGroupName(ctx, tx, metricModel.GroupName)
// 	if err != nil {
// 		logger.Errorf("RetriveGroupIDByGroupName error: %s", err.Error())
// 		span.SetStatus(codes.Error, "根据groupName更新GroupID失败")
// 		return "", err
// 	}
// 	if metricModel.Builtin != metricGroup.Builtin {
// 		errDetail := "The built-in model can only be placed in a built-in group, while the non-built-in model can only be placed in a non-built-in group."
// 		logger.Errorf(errDetail)
// 		span.SetStatus(codes.Error, "指标模型的内置属性与所绑定的分组的内置属性不匹配")
// 		return "", rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_MetricModel_InvalidParameter_Group).
// 			WithErrorDetails(errDetail)
// 	}
// 	metricModel.GroupID = metricGroup.GroupID

// 	// 1. 创建模型
// 	err = mms.mma.CreateMetricModel(ctx, tx, metricModel)
// 	if err != nil {
// 		logger.Errorf("CreateMetricModel error: %s", err.Error())
// 		span.SetStatus(codes.Error, "创建指标模型失败")

// 		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModel_InternalError).
// 			WithErrorDetails(err.Error())
// 	}

// 	// 2. 创建模型下的任务
// 	if metricModel.Task != nil {
// 		// 写任务表
// 		userID := ""
// 		if ctx.Value(interfaces.USER_KEY) != nil {
// 			userID = ctx.Value(interfaces.USER_KEY).(string)
// 		}
// 		metricModel.Task.Creator = userID
// 		err = mms.mmts.CreateMetricTask(ctx, tx, *metricModel.Task)
// 		if err != nil {
// 			logger.Errorf("CreateMetricTasks error: %s", err.Error())
// 			span.SetStatus(codes.Error, "创建指标模型持久化任务失败")

// 			return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModel_InternalError).
// 				WithErrorDetails(err.Error())
// 		}
// 		// 请求 data-model-job 服务开启任务
// 		modelJobCfg := &interfaces.DataModelJobCfg{
// 			JobID:      metricModel.Task.TaskID,
// 			JobType:    interfaces.JOB_TYPE_SCHEDULE,
// 			ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
// 			MetricTask: metricModel.Task,
// 			Schedule:   metricModel.Task.Schedule,
// 		}
// 		logger.Infof("metric model %s request data-model-job start job %d", metricModel.ModelName, metricModel.Task.TaskID)
// 		uncancelableCtx := context.WithoutCancel(ctx)
// 		go func() {
// 			err := mms.dmja.StartJob(uncancelableCtx, modelJobCfg)
// 			if err != nil {
// 				logger.Errorf("Start metric job[%s] failed: %s", modelJobCfg.JobID, err.Error())
// 			}
// 		}()
// 	}

// 	span.SetStatus(codes.Ok, "")
// 	return metricModel.ModelID, nil
// }

// 批量创建指标模型
func (mms *metricModelService) CreateMetricModels(ctx context.Context, metricModels []*interfaces.MetricModel,
	mode string) (modelIDs []string, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Create metric models")
	defer span.End()

	// 判断userid是否有创建指标模型的权限（策略决策）
	err = mms.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_METRIC_MODEL,
		ID:   interfaces.RESOURCE_ID_ALL,
	}, []string{interfaces.OPERATION_TYPE_CREATE})
	if err != nil {
		return nil, err
	}

	currentTime := time.Now().UnixMilli()
	for _, model := range metricModels {
		// 如果视图ID为空，则生成一个
		if model.ModelID == "" {
			model.ModelID = xid.New().String()
		}
		err = mms.processForCreate(ctx, model, currentTime)
		if err != nil {
			return nil, err
		}
	}

	// 0. 开始事务
	tx, err := mms.db.Begin()
	if err != nil {
		logger.Errorf("Begin transaction error: %s", err.Error())
		span.SetStatus(codes.Error, "事务开启失败")
		o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModel_InternalError_BeginTransactionFailed).
			WithErrorDetails(err.Error())
	}
	// 0.1 异常时
	defer func() {
		switch err {
		case nil:
			// 提交事务
			err = tx.Commit()
			if err != nil {
				logger.Errorf("CreateMetricModel Transaction Commit Failed:%v", err)
				span.SetStatus(codes.Error, "提交事务失败")
				o11y.Error(ctx, fmt.Sprintf("CreateMetricModel Transaction Commit Failed: %s", err.Error()))

			}
			logger.Infof("CreateMetricModel Transaction Commit Success:%v", metricModels)
			o11y.Debug(ctx, fmt.Sprintf("CreateMetricModel Transaction Commit Success: %v", metricModels))
		default:
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logger.Errorf("CreateMetricModel Transaction Rollback Error:%v", rollbackErr)
				span.SetStatus(codes.Error, "事务回滚失败")
				o11y.Error(ctx, fmt.Sprintf("CreateMetricModel Transaction Rollback Error: %s", err.Error()))
			}
		}
	}()

	createModels, updateModels, err := mms.handleMetricModelImportMode(ctx, mode, metricModels)
	if err != nil {
		return nil, err
	}

	// 挨个调更新函数（其中校验对象的更新权限，若无，报错）
	for _, model := range updateModels {
		err = mms.UpdateMetricModel(ctx, tx, *model)
		if err != nil {
			return nil, err
		}
	}

	// 前面已经校验创建权限，可以直接创建
	createSrcs := []interfaces.Resource{}
	for _, model := range createModels {
		//创建之前需要先更新一下group_ID
		metricGroup, err := mms.RetriveGroupIDByGroupName(ctx, tx, model.GroupName)
		if err != nil {
			logger.Errorf("RetriveGroupIDByGroupName error: %s", err.Error())
			span.SetStatus(codes.Error, "根据groupName更新GroupID失败")
			return nil, err
		}
		if model.Builtin != metricGroup.Builtin {
			errDetail := "The built-in model can only be placed in a built-in group, while the non-built-in model can only be placed in a non-built-in group."
			logger.Errorf(errDetail)
			span.SetStatus(codes.Error, "指标模型的内置属性与所绑定的分组的内置属性不匹配")
			return nil, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_MetricModel_InvalidParameter_Group).
				WithErrorDetails(errDetail)
		}
		model.GroupID = metricGroup.GroupID
		modelIDs = append(modelIDs, model.ModelID)

		// 1. 逐个创建模型
		err = mms.mma.CreateMetricModel(ctx, tx, *model)
		if err != nil {
			logger.Errorf("CreateMetricModel error: %s", err.Error())
			span.SetStatus(codes.Error, "创建指标模型失败")

			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModel_InternalError).
				WithErrorDetails(err.Error())
		}

		// 2. 创建模型下的任务
		if model.Task != nil {
			// 写任务表
			accountInfo := interfaces.AccountInfo{}
			if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
				accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
			}
			model.Task.Creator = accountInfo
			err = mms.mmts.CreateMetricTask(ctx, tx, *model.Task)
			if err != nil {
				logger.Errorf("CreateMetricTasks error: %s", err.Error())
				span.SetStatus(codes.Error, "创建指标模型持久化任务失败")

				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModel_InternalError).
					WithErrorDetails(err.Error())
			}
			// 请求 data-model-job 服务开启任务
			modelJobCfg := &interfaces.DataModelJobCfg{
				JobID:      model.Task.TaskID,
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
				MetricTask: model.Task,
				Schedule:   model.Task.Schedule,
			}
			logger.Infof("metric model %s request data-model-job start job %d", model.ModelName, model.Task.TaskID)
			uncancelableCtx := context.WithoutCancel(ctx)
			go func() {
				err := mms.dmja.StartJob(uncancelableCtx, modelJobCfg)
				if err != nil {
					logger.Errorf("Start metric job[%s] failed: %s", modelJobCfg.JobID, err.Error())
				}
			}()
		}

		name := common.ProcessUngroupedName(ctx, model.GroupName, model.ModelName)
		createSrcs = append(createSrcs, interfaces.Resource{
			ID:   model.ModelID,
			Type: interfaces.RESOURCE_TYPE_METRIC_MODEL,
			Name: name,
		})
	}

	// 注册资源策略
	err = mms.ps.CreateResources(ctx, createSrcs, interfaces.COMMON_OPERATIONS)
	if err != nil {
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return modelIDs, nil
}

// 修改指标模型
func (mms *metricModelService) UpdateMetricModel(ctx context.Context, tx *sql.Tx, metricModel interfaces.MetricModel) (err error) {
	updateCtx, updateSpan := ar_trace.Tracer.Start(ctx, "Update metric model")
	defer updateSpan.End()

	// 判断userid是否有创建指标模型的权限（策略决策）
	err = mms.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_METRIC_MODEL,
		ID:   metricModel.ModelID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	err = mms.checkDepends(ctx, &metricModel)
	if err != nil {
		return err
	}

	updateTime := time.Now().UnixMilli()
	metricModel.UpdateTime = updateTime

	updateSpan.SetAttributes(
		attr.Key("metric_model_id").String(metricModel.ModelID),
		attr.Key("metric_model_name").String(metricModel.ModelName))

	if tx == nil {
		// 0. 开始事务
		tx, err = mms.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			updateSpan.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

			return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModel_InternalError_BeginTransactionFailed).
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
				logger.Infof("UpdateMetricModel Transaction Commit Success:%v", metricModel.ModelName)
				o11y.Debug(updateCtx, fmt.Sprintf("UpdateMetricModel Transaction Commit Success: %s", metricModel.ModelName))
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

	// 更新之前先获取下group
	metricGroup, err := mms.RetriveGroupIDByGroupName(ctx, tx, metricModel.GroupName)
	if err != nil {
		logger.Errorf("RetriveGroupIDByGroupName error: %s", err.Error())
		updateSpan.SetStatus(codes.Error, "根据groupName更新GroupID失败")
		return err
	}
	if metricModel.Builtin != metricGroup.Builtin {
		errDetail := "The built-in model can only be placed in a built-in group, while the non-built-in model can only be placed in a non-built-in group."
		logger.Errorf(errDetail)
		updateSpan.SetStatus(codes.Error, "指标模型的内置属性与所绑定的分组的内置属性不匹配")
		return rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_MetricModel_InvalidParameter_Group).
			WithErrorDetails(errDetail)
	}
	metricModel.GroupID = metricGroup.GroupID

	// 更新模型信息
	err = mms.mma.UpdateMetricModel(updateCtx, tx, metricModel)
	if err != nil {
		logger.Errorf("metricModel error: %s", err.Error())
		updateSpan.SetStatus(codes.Error, "修改指标模型失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModel_InternalError).
			WithErrorDetails(err.Error())
	}

	// 按模型id去获取任务，判断模型下是否存在任务
	// 若提交的task为空，模型下存在任务，则是删除操作；若提交的task为空，模型下不存在任务，do nothing
	// 若提交的task不为空，模型下存在任务，则是更新操作；若提交的task不为空，模型下不存在任务，则是创建任务的操作。
	mapTask, err := mms.mmts.GetMetricTasksByModelIDs(ctx, []string{metricModel.ModelID})
	if err != nil {
		return err
	}
	task, exists := mapTask[metricModel.ModelID]

	if metricModel.Task != nil {
		if exists {
			// 更新操作
			metricModel.Task.TaskID = task.TaskID
			metricModel.Task.ModuleType = interfaces.MODULE_TYPE_METRIC_MODEL
			metricModel.Task.ModelID = metricModel.ModelID
			metricModel.Task.MeasureName = metricModel.MeasureName
			metricModel.Task.UpdateTime = updateTime
			metricModel.Task.ScheduleSyncStatus = interfaces.SCHEDULE_SYNC_STATUS_FINISH
			// 修改任务
			err = mms.mmts.UpdateMetricTask(updateCtx, tx, *metricModel.Task)
			if err != nil {
				logger.Errorf("UpdateMetricTask error: %s", err.Error())
				updateSpan.SetStatus(codes.Error, "修改指标模型持久化任务失败")

				return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModel_InternalError).
					WithErrorDetails(err.Error())
			}
			// 请求 data-model-job 更新任务
			newMetricJobCfg := &interfaces.DataModelJobCfg{
				JobID:      metricModel.Task.TaskID,
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
				MetricTask: metricModel.Task,
				Schedule:   metricModel.Task.Schedule,
			}
			uncancelableCtx := context.WithoutCancel(ctx)
			go func() {
				err = mms.dmja.UpdateJob(uncancelableCtx, newMetricJobCfg)
				if err != nil {
					logger.Errorf("Update metric model job[%s] failed, %s", newMetricJobCfg.JobID, err.Error())
				}
			}()
		} else {
			// 创建操作
			accountInfo := interfaces.AccountInfo{}
			if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
				accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
			}
			metricModel.Task.Creator = accountInfo
			// 生成分布式ID
			metricModel.Task.TaskID = xid.New().String()
			metricModel.Task.ModuleType = interfaces.MODULE_TYPE_METRIC_MODEL
			metricModel.Task.ModelID = metricModel.ModelID
			metricModel.Task.MeasureName = metricModel.MeasureName
			metricModel.Task.UpdateTime = updateTime
			// 创建，创建的任务状态为完成
			metricModel.Task.ScheduleSyncStatus = interfaces.SCHEDULE_SYNC_STATUS_FINISH
			// 初始化时间窗口对应的计划时间。只有新建任务和任务调度执行成功之后修改计划时间，其他情况不允许修改
			// 追溯时长转毫秒，用于计算计划时间
			durationV := int64(0)
			if metricModel.Task.RetraceDuration != "" {
				dur, err := common.ParseDuration(metricModel.Task.RetraceDuration, common.DurationDayHourRE, false)
				if err != nil {
					logger.Errorf("Failed to parse retrace duration, err: %v", err.Error())
					return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_RetraceDuration).
						WithErrorDetails(err.Error())
				}
				durationV = int64(dur / (time.Millisecond / time.Nanosecond))
			}
			planTime := time.Now().UnixNano() / int64(time.Millisecond/time.Nanosecond)
			metricModel.Task.PlanTime = planTime - durationV

			err = mms.mmts.CreateMetricTask(updateCtx, tx, *metricModel.Task)
			if err != nil {
				logger.Errorf("CreateMetricTasks error: %s", err.Error())
				updateSpan.SetStatus(codes.Error, "创建指标模型持久化任务失败")
				return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModel_InternalError).
					WithErrorDetails(err.Error())
			}
			// 请求 data-model-job 服务开启任务
			modelJobCfg := &interfaces.DataModelJobCfg{
				JobID:      metricModel.Task.TaskID,
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
				MetricTask: metricModel.Task,
				Schedule:   metricModel.Task.Schedule,
			}
			logger.Infof("metric model %s request data-model-job start job %d", metricModel.ModelName, metricModel.Task.TaskID)
			uncancelableCtx := context.WithoutCancel(ctx)
			go func() {
				err = mms.dmja.StartJob(uncancelableCtx, modelJobCfg)
				if err != nil {
					logger.Errorf("Start metric job[%s] failed: %s", modelJobCfg.JobID, err.Error())
				}
			}()
		}
	} else {
		// 请求的task为空（关闭持久化），且模型下存在任务，则需要删除该任务
		if exists {
			// 删除模型下的任务
			err = mms.mmts.DeleteMetricTaskByTaskIDs(updateCtx, tx, []string{task.TaskID})
			if err != nil {
				logger.Errorf("DeleteMetricTaskByTaskID error: %s", err.Error())
				updateSpan.SetStatus(codes.Error, "删除指标模型持久化任务失败")
				return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModel_InternalError).
					WithErrorDetails(err.Error())
			}
			// 请求 data-model-job 删除任务
			logger.Infof("metric model id '%s' delete persist task, request data-model-job to stop job", metricModel.ModelID)
			uncancelableCtx := context.WithoutCancel(ctx)
			go func() {
				err = mms.dmja.StopJobs(uncancelableCtx, []string{task.TaskID})
				if err != nil {
					logger.Errorf("Stop metric job[%s] failed: %s", task.TaskID, err.Error())
				}
			}()
		}
	}

	// 请求更新资源名称的接口，更新资源的名称
	if metricModel.IfNameModify {
		err = mms.ps.UpdateResource(ctx, interfaces.Resource{
			ID:   metricModel.ModelID,
			Type: interfaces.RESOURCE_TYPE_METRIC_MODEL,
			Name: common.ProcessUngroupedName(ctx, metricModel.GroupName, metricModel.ModelName),
		})
		if err != nil {
			return err
		}
	}

	updateSpan.SetStatus(codes.Ok, "")
	return nil
}

// 按 id 获取指标模型信息。内部调用函数，无需权限验证
func (mms *metricModelService) GetMetricModelByModelID(ctx context.Context, modelID string) (interfaces.MetricModel, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("查询指标模型[%s]信息", modelID))
	span.SetAttributes(attr.Key("model_id").String(modelID))
	defer span.End()

	var mm interfaces.MetricModel
	metricModel, exist, err := mms.mma.GetMetricModelByModelID(ctx, modelID)
	if err != nil {
		logger.Errorf("GetMetricModelByModelID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get metric model[%s] error: %v", modelID, err))
		return mm, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_GetModelByIDFailed).WithErrorDetails(err.Error())
	}
	if !exist {
		logger.Debugf("Metric Model %s not found!", modelID)
		span.SetStatus(codes.Error, fmt.Sprintf("Metric model[%s] not found", modelID))
		return mm, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_MetricModel_MetricModelNotFound)
	}

	span.SetStatus(codes.Ok, "")
	return metricModel, nil
}

// 批量删除指标模型
func (mms *metricModelService) DeleteMetricModels(ctx context.Context, tx *sql.Tx, modelIDs []string) (affectedRows int64, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete metric models")
	defer span.End()

	// 先获取资源序列 fmt.Sprintf("%s%s", interfaces.METRIC_MODEL_RESOURCE_ID_PREFIX, metricModel.ModelID),
	matchResouces, err := mms.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_METRIC_MODEL, modelIDs,
		[]string{interfaces.OPERATION_TYPE_DELETE}, false)
	if err != nil {
		return 0, err
	}
	// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
	for _, mID := range modelIDs {
		if _, exist := matchResouces[mID]; !exist {
			return 0, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for metric model's delete operation.")
		}
	}

	// 按modelid获取任务id列表
	taskIDs, err := mms.mmts.GetMetricTaskIDsByModelIDs(ctx, modelIDs)
	if err != nil {
		return 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError).WithErrorDetails(err.Error())
	}
	// 请求data-model-job 批量停止任务
	if len(taskIDs) > 0 {
		// 请求data-model-job服务批量停止实际运行的任务
		if err = mms.dmja.StopJobs(ctx, taskIDs); err != nil {
			span.SetStatus(codes.Error, "Stop metric model jobs failed")
			return 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_StopJobFailed).WithErrorDetails(err.Error())
		}
	}

	// 事务为空时，开启一个事务
	if tx == nil {
		// 0. 开始事务
		tx, err = mms.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

			return 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModel_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("DeleteMetricModels Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("DeleteMetricModels Transaction Commit Failed: %s", err.Error()))

				}
				logger.Infof("DeleteMetricModels Transaction Commit Success:%v", modelIDs)
				o11y.Debug(ctx, fmt.Sprintf("DeleteMetricModels Transaction Commit Success: %v", modelIDs))
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("DeleteMetricModels Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("DeleteMetricModels Transaction Rollback Error: %s", err.Error()))
				}
			}
		}()
	}

	// 删除指标模型
	rowsAffect, err := mms.mma.DeleteMetricModels(ctx, tx, modelIDs)
	span.SetAttributes(attr.Key("rows_affect").Int64(rowsAffect))
	if err != nil {
		logger.Errorf("DeleteMetricModel error: %s", err.Error())
		span.SetStatus(codes.Error, "删除指标模型失败")

		return rowsAffect, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError).WithErrorDetails(err.Error())
	}
	logger.Infof("DeleteMetricModel: Rows affected is %v, request delete modelIDs is %v!", rowsAffect, len(modelIDs))
	if rowsAffect != int64(len(modelIDs)) {
		logger.Warnf("Delete models number %v not equal requerst models number %v!", rowsAffect, len(modelIDs))

		o11y.Warn(ctx, fmt.Sprintf("Delete models number %v not equal requerst models number %v!", rowsAffect, len(modelIDs)))
	}

	// 删除指标模型下的任务
	err = mms.mmts.DeleteMetricTaskByTaskIDs(ctx, tx, taskIDs)
	if err != nil {
		logger.Errorf("CreateMetricTasks error: %s", err.Error())
		span.SetStatus(codes.Error, "批量删除指标模型持久化任务失败")
		return 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModel_InternalError).
			WithErrorDetails(err.Error())
	}

	//  清除资源策略
	err = mms.ps.DeleteResources(ctx, interfaces.RESOURCE_TYPE_METRIC_MODEL, modelIDs)
	if err != nil {
		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return rowsAffect, nil
}

// 分页查询指标模型
// func (mms *metricModelService) ListMetricModels(ctx context.Context,
// 	parameter interfaces.MetricModelsQueryParams) ([]interfaces.MetricModel, int, error) {

// 	listCtx, listSpan := ar_trace.Tracer.Start(ctx, "查询指标模型列表")

// 	//获取指标模型列表
// 	emptyMetricModels := make([]interfaces.MetricModel, 0)
// 	metricModelArr, err := mms.mma.ListMetricModels(listCtx, parameter)
// 	if err != nil {
// 		logger.Errorf("ListMetricModels error: %s", err.Error())
// 		listSpan.SetStatus(codes.Error, "List metric models error")
// 		listSpan.End()

// 		return emptyMetricModels, 0, rest.NewHTTPError(listCtx, http.StatusInternalServerError,
// 			derrors.DataModel_MetricModel_InternalError).WithErrorDetails(err.Error())
// 	}

// 	modelIDs := make([]string, 0)
// 	for _, model := range metricModelArr {
// 		modelIDs = append(modelIDs, model.ModelID)
// 	}
// 	// 根据模型id获取任务信息
// 	modelTaskMap, err := mms.mmts.GetMetricTasksByModelIDs(listCtx, modelIDs)
// 	if err != nil {
// 		logger.Errorf("GetMetricTasksByModelIDs error: %s", err.Error())
// 		listSpan.SetStatus(codes.Error, fmt.Sprintf("Get metric model[%s] error: %v", modelIDs, err))
// 		listSpan.End()

// 		return emptyMetricModels, 0, rest.NewHTTPError(listCtx, http.StatusInternalServerError,
// 			derrors.DataModel_MetricModel_InternalError_GetMetricTasksByModelIDsFailed).WithErrorDetails(err.Error())
// 	}

// 	listSpan.SetStatus(codes.Ok, "")
// 	listSpan.End()

// 	// 数据视图 id 换成名称，需要循环遍历查询。
// 	dataviewCtx, dataviewSpan := ar_trace.Tracer.Start(ctx, "Metric model's data_view_id transfer to name")
// 	metricModels := make([]interfaces.MetricModel, 0)
// 	for _, model := range metricModelArr {
// 		task, taskExist := modelTaskMap[model.ModelID]
// 		if taskExist {
// 			model.Task = &task
// 			// 若存在任务，则把任务中的索引库类型换索引库名称
// 			// 获取索引库名称
// 			simpleIndexBases, err := mms.iba.GetSimpleIndexBasesByTypes(ctx, []string{model.Task.IndexBase})
// 			if err != nil {
// 				logger.Errorf("GetSimpleIndexBasesByTypes error: %s", err.Error())

// 				return emptyMetricModels, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
// 					derrors.DataModel_MetricModel_InternalError_GetSimpleIndexBasesByTypesFailed).WithErrorDetails(err.Error())
// 			}
// 			// 若索引库查询为空，则不赋值
// 			if len(simpleIndexBases) > 0 {
// 				// 遍历模型任务，把索引库名称赋值到任务上
// 				model.Task.IndexBaseName = simpleIndexBases[0].Name
// 			}
// 		}

// 		viewName, exist, err := mms.dvs.CheckDataViewExistByID(dataviewCtx, nil, model.DataViewID)
// 		if err != nil {
// 			logger.Errorf("GetDataViewByID error: %s", err.Error())
// 			dataviewSpan.SetStatus(codes.Error, fmt.Sprintf("Get data view[%s] error", model.DataViewID))
// 			dataviewSpan.End()

// 			return emptyMetricModels, 0, rest.NewHTTPError(dataviewCtx, http.StatusInternalServerError,
// 				derrors.DataModel_MetricModel_InternalError_GetDataViewByIDFailed).WithErrorDetails(err.Error())
// 		}
// 		if !exist {
// 			// 查询时，如果数据视图已经被删除，则数据视图字段为空
// 			o11y.Warn(dataviewCtx, fmt.Sprintf("Data view %s not found", model.DataViewID))
// 		} else {
// 			model.DataViewName = viewName
// 		}
// 		metricModels = append(metricModels, model)
// 	}
// 	dataviewSpan.SetStatus(codes.Ok, "")
// 	dataviewSpan.End()

// 	//调用driven层，获取总数
// 	totalCtx, totalSpan := ar_trace.Tracer.Start(ctx, "Get metric model total")
// 	defer totalSpan.End()

// 	total, err := mms.mma.GetMetricModelsTotal(totalCtx, parameter)
// 	if err != nil {
// 		logger.Errorf("GetMetricModelsTotal error: %s", err.Error())
// 		totalSpan.SetStatus(codes.Error, "Get metric model total error")

// 		return emptyMetricModels, 0, rest.NewHTTPError(totalCtx, http.StatusInternalServerError,
// 			derrors.DataModel_MetricModel_InternalError).WithErrorDetails(err.Error())
// 	}
// 	totalSpan.SetStatus(codes.Ok, "")

// 	return metricModels, total, nil
// }

// 分页查询指标模型。只展示有权限的资源
func (mms *metricModelService) ListSimpleMetricModels(ctx context.Context, parameter interfaces.MetricModelsQueryParams) (
	[]interfaces.SimpleMetricModel, int, error) {
	listCtx, listSpan := ar_trace.Tracer.Start(ctx, "查询指标模型简单信息列表")
	listSpan.End()

	//获取指标模型列表（不分页，获取所有的指标模型)
	models, err := mms.mma.ListSimpleMetricModels(listCtx, parameter)
	emptySimpleMetricModels := make([]interfaces.SimpleMetricModel, 0)
	if err != nil {
		logger.Errorf("ListMetricModels error: %s", err.Error())
		listSpan.SetStatus(codes.Error, "List simple metric models error")
		listSpan.End()
		return emptySimpleMetricModels, 0, rest.NewHTTPError(listCtx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError).WithErrorDetails(err.Error())
	}

	if len(models) == 0 {
		return models, 0, nil
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	// 处理资源id
	resMids := make([]string, 0)
	for _, m := range models {
		resMids = append(resMids, m.ModelID)
	}
	matchResoucesMap, err := mms.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_METRIC_MODEL, resMids,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		return emptySimpleMetricModels, 0, err
	}

	// 遍历对象
	results := make([]interfaces.SimpleMetricModel, 0)
	for _, model := range models {
		if resrc, exist := matchResoucesMap[model.ModelID]; exist {
			model.Operations = resrc.Operations // 用户当前有权限的操作
			results = append(results, model)
		}
	}

	// limit = -1,则返回所有
	if parameter.Limit == -1 {
		return results, len(results), nil
	}

	// 分页
	// 检查起始位置是否越界
	if parameter.Offset < 0 || parameter.Offset >= len(results) {
		return emptySimpleMetricModels, 0, nil
	}
	// 计算结束位置
	end := parameter.Offset + parameter.Limit
	if end > len(results) {
		end = len(results)
	}

	listSpan.SetStatus(codes.Ok, "")
	return results[parameter.Offset:end], len(results), nil
}

// 获取指标模型信息，通过 include_view 来控制是否包含了数据视图过滤条件
func (mms *metricModelService) GetMetricModels(ctx context.Context, modelIDs []string,
	includeView bool) ([]interfaces.MetricModelWithFilters, error) {

	// 获取指标模型
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get metric models")
	span.SetAttributes(attr.Key("model_ids").String(fmt.Sprintf("%v", modelIDs)))

	// id去重后再查
	mIDs := common.DuplicateSlice(modelIDs)

	// 获取模型基本信息
	metricModels, err := mms.mma.GetMetricModelsByModelIDs(ctx, mIDs)
	mmfilters := make([]interfaces.MetricModelWithFilters, 0)
	if err != nil {
		logger.Errorf("GetMetricModelByModelID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get metric model[%v] error: %v", mIDs, err))
		span.End()

		return mmfilters, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_GetModelByIDFailed).WithErrorDetails(err.Error())
	}
	if len(metricModels) != len(mIDs) {
		errStr := fmt.Sprintf("Exists any models not found, expect model nums is [%d], actual models num is [%d]", len(mIDs), len(metricModels))
		logger.Errorf(errStr)
		span.SetStatus(codes.Error, errStr)
		span.End()

		return mmfilters, rest.NewHTTPError(ctx, http.StatusNotFound,
			derrors.DataModel_MetricModel_MetricModelNotFound).WithErrorDetails(errStr)
	}

	// 先获取资源序列  todo:
	matchResouces, err := mms.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_METRIC_MODEL, mIDs,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		return mmfilters, err
	}
	// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
	for _, model := range metricModels {
		if _, exist := matchResouces[model.ModelID]; !exist {
			return mmfilters, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for metric model's view_detail operation.")
		}
	}

	// 根据模型id获取任务信息
	modelTaskMap, err := mms.mmts.GetMetricTasksByModelIDs(ctx, mIDs)
	if err != nil {
		logger.Errorf("GetMetricTasksByModelIDs error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get metric model[%v] error: %v", mIDs, err))
		span.End()

		return mmfilters, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_GetMetricTasksByModelIDsFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	span.End()

	// 获取数据视图的过滤条件
	dataViewCtx, dataViewSpan := ar_trace.Tracer.Start(ctx, "获取数据视图的过滤条件")
	dsTypes := make([]string, 0)
	dvIDs := make([]string, 0)
	fieldsNum := make([]int, 0)
	for _, model := range metricModels {
		// model 上附加 opreration 字段
		model.Operations = matchResouces[model.ModelID].Operations

		task, taskExist := modelTaskMap[model.ModelID]
		if taskExist {
			model.Task = &task
			// 若存在任务，则把任务中的索引库类型换索引库名称
			// 获取索引库名称
			simpleIndexBases, err := mms.iba.GetSimpleIndexBasesByTypes(ctx, []string{model.Task.IndexBase})
			if err != nil {
				logger.Errorf("GetSimpleIndexBasesByTypes error: %s", err.Error())

				return mmfilters, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					derrors.DataModel_MetricModel_InternalError_GetSimpleIndexBasesByTypesFailed).WithErrorDetails(err.Error())
			}
			// 若索引库查询为空，则不赋值
			if len(simpleIndexBases) > 0 {
				// 遍历模型任务，把索引库名称赋值到任务上
				model.Task.IndexBaseName = simpleIndexBases[0].Name
			}
		}
		switch model.MetricType {
		case interfaces.ATOMIC_METRIC:
			// 原子指标翻译数据源，衍生指标翻译依赖的指标模型
			dsTypes = append(dsTypes, model.DataSource.Type)
			dvIDs = append(dvIDs, model.DataSource.ID)

			if includeView {
				// 获取视图信息
				dataView, err := mms.dvs.GetDataView(dataViewCtx, model.DataSource.ID)
				if err != nil {
					logger.Errorf("GetDataView error: %s", err.Error())
					dataViewSpan.SetAttributes(attr.Key("data_source_ids").StringSlice(dvIDs),
						attr.Key("data_source_types").StringSlice(dsTypes))

					dataViewSpan.SetStatus(codes.Error, "获取数据视图过滤条件失败")
					o11y.Error(dataViewCtx, fmt.Sprintf("Get data view filters failed, dataViewID: %s, error: %v ", model.DataSource.ID, err))
					dataViewSpan.End()
					return mmfilters, rest.NewHTTPError(dataViewCtx, http.StatusInternalServerError,
						derrors.DataModel_MetricModel_InternalError_GetDataViewQueryFiltersFailed).WithErrorDetails(err.Error())
				}
				// 请求不同的数据源去获取不同的信息
				if model.DataSource.Type == interfaces.QueryType_SQL {
					// 分组字段存储的是[]string，交互查询用的是[]Field
					sqlConfig := model.FormulaConfig.(interfaces.SQLConfig)
					// 处理分组字段，技术名、显示名
					groupByFileds := make([]interfaces.Field, 0)
					for _, groupBy := range model.FormulaConfig.(interfaces.SQLConfig).GroupByFields {
						groupByFileds = append(groupByFileds, interfaces.Field{
							Name:        dataView.FieldsMap[groupBy].Name,
							DisplayName: dataView.FieldsMap[groupBy].DisplayName,
							Type:        dataView.FieldsMap[groupBy].Type,
						})
					}
					sqlConfig.GroupByFieldsDetail = groupByFileds
					model.FormulaConfig = sqlConfig
					// 处理分析维度，技术名、显示名
					for i, analysisDim := range model.AnalysisDims {
						model.AnalysisDims[i].Name = dataView.FieldsMap[analysisDim.Name].Name
						model.AnalysisDims[i].DisplayName = dataView.FieldsMap[analysisDim.Name].DisplayName
						model.AnalysisDims[i].Type = dataView.FieldsMap[analysisDim.Name].Type
					}
					// 处理排序字段，显示名
					for i, orderBy := range model.OrderByFields {
						if orderBy.Name == interfaces.VALUE_FIELD_NAME {
							model.OrderByFields[i].DisplayName = interfaces.VALUE_FIELD_DISPLAY_NAME
							continue
						}
						model.OrderByFields[i].Name = dataView.FieldsMap[orderBy.Name].Name
						model.OrderByFields[i].DisplayName = dataView.FieldsMap[orderBy.Name].DisplayName
						model.OrderByFields[i].Type = dataView.FieldsMap[orderBy.Name].Type
					}
				}

				fieldsNum = append(fieldsNum, len(dataView.Fields))

				model.DataSource.Name = dataView.ViewName
				metricModelWithFilters := interfaces.MetricModelWithFilters{
					MetricModel: model,
				}
				metricModelWithFilters.DataView = dataView

				// 如果data_source是data_view,则需要赋值data_view_id
				metricModelWithFilters.DataViewID = model.DataSource.ID
				metricModelWithFilters.DataViewName = dataView.ViewName

				mmfilters = append(mmfilters, metricModelWithFilters)
			} else {
				// 获取视图名称
				viewName, exist, err := mms.dvs.CheckDataViewExistByID(dataViewCtx, nil, model.DataSource.ID)
				if err != nil {
					logger.Errorf("GetDataViewByID error: %s", err.Error())
					dataViewSpan.SetAttributes(attr.Key("data_source_ids").StringSlice(dvIDs),
						attr.Key("data_source_types").StringSlice(dsTypes),
						attr.Key("fields_nums").IntSlice(fieldsNum))

					dataViewSpan.SetStatus(codes.Error, "获取数据视图过滤条件失败")
					dataViewSpan.End()
					o11y.Error(dataViewCtx, fmt.Sprintf("Get data view failed, dataViewID: %s, error: %v ", model.DataSource.ID, err))

					return mmfilters, rest.NewHTTPError(dataViewCtx, http.StatusInternalServerError,
						derrors.DataModel_MetricModel_InternalError_GetDataViewByIDFailed).WithErrorDetails(err.Error())
				}
				if !exist {
					// 查询时，如果数据视图已经被删除，则数据视图字段为空
					o11y.Warn(dataViewCtx, fmt.Sprintf("Data view %s not found", model.DataSource.ID))
				} else {

					viewFields, fieldsMap, err := mms.getMetricViewFieldsAndGroupFields(ctx, model)
					if err != nil {
						// 查询时报错，记录日志
						o11y.Warn(ctx, fmt.Sprintf("getMetricViewFieldsAndGroupFields[%s] err: %s", model.ModelID, err.Error()))

						// logger.Errorf("GetDataView error: %s", err.Error())
						// dataViewSpan.SetAttributes(attr.Key("data_source_ids").StringSlice(dvIDs),
						// 	attr.Key("data_source_types").StringSlice(dsTypes))

						// dataViewSpan.SetStatus(codes.Error, "获取数据视图过滤条件失败")
						// o11y.Error(dataViewCtx, fmt.Sprintf("Get data view filters failed, dataViewID: %s, error: %v ", model.DataSource.ID, err))
						// dataViewSpan.End()
						// return mmfilters, rest.NewHTTPError(dataViewCtx, http.StatusInternalServerError,
						// 	derrors.DataModel_MetricModel_InternalError_GetDataViewQueryFiltersFailed).WithErrorDetails(err.Error())
					}

					// sql-原子：fields_map为指标模型的分组字段 +分析维度 + 排序字段
					// 给分组字段、分析维度、排序字段的 display_name 赋值
					if model.DataSource.Type == interfaces.QueryType_SQL {
						sqlConfig := model.FormulaConfig.(interfaces.SQLConfig)

						// 处理分组字段，技术名、显示名
						groupByFileds := make([]interfaces.Field, 0)
						for _, groupBy := range model.FormulaConfig.(interfaces.SQLConfig).GroupByFields {
							groupByFileds = append(groupByFileds, interfaces.Field{
								Name:        viewFields[groupBy].Name,
								DisplayName: viewFields[groupBy].DisplayName,
								Type:        viewFields[groupBy].Type,
							})

							fieldsMap[viewFields[groupBy].Name] = interfaces.Field{
								Name:        viewFields[groupBy].Name,
								DisplayName: viewFields[groupBy].DisplayName,
								Type:        viewFields[groupBy].Type,
								Comment:     &viewFields[groupBy].Comment,
							}
						}
						sqlConfig.GroupByFieldsDetail = groupByFileds
						model.FormulaConfig = sqlConfig

						// 衍生指标，转换分析维度和排序字段，以及构造指标的字段map
						transferDimAndOrder(ctx, &model, viewFields, fieldsMap)

					} else {
						// dsl / promql：fields_map 为所绑定的视图的字段信息
						for _, field := range viewFields {
							fieldsMap[field.Name] = interfaces.Field{
								Name:        field.Name,
								DisplayName: field.DisplayName,
								Type:        field.Type,
								Comment:     &field.Comment,
							}
						}
					}
					model.FieldsMap = fieldsMap

					model.DataSource.Name = viewName
					// 如果data_source是data_view,则需要赋值data_view_id
					model.DataViewID = model.DataSource.ID
					model.DataViewName = viewName
				}

				mmfilters = append(mmfilters, interfaces.MetricModelWithFilters{
					MetricModel: model,
				})
			}

		case interfaces.DERIVED_METRIC:
			// 按id获取名称。 衍生指标的分析维度依赖原子指标，所以再请求一次原子指标的详情，把其分析维度拿到，作为衍生指标的分析维度的参考集
			derivedConfig := model.FormulaConfig.(interfaces.DerivedConfig)
			dependModels, err := mms.GetMetricModels(ctx, []string{derivedConfig.DependMetricModel.ID}, false)
			if err != nil {
				return mmfilters, err
			}
			if len(dependModels) != 1 {
				// 查询时，如果原子指标已经被删除，则不翻译
				o11y.Warn(ctx, fmt.Sprintf("Depend metric model %s not found", derivedConfig.DependMetricModel.ID))
			} else {
				derivedConfig.DependMetricModel.GroupName = dependModels[0].GroupName
				derivedConfig.DependMetricModel.Name = dependModels[0].ModelName
				model.FormulaConfig = derivedConfig

				// 衍生:  fields_map为依赖的原子指标分组字段 + 本身的分析维度  + 排序字段(需注意值字段)
				// 当前指标模型映射到视图的原始字段集
				viewFields, fieldsMap, err := mms.getMetricViewFieldsAndGroupFields(ctx, model)
				if err != nil {
					// 查询时报错，记录日志
					o11y.Warn(ctx, fmt.Sprintf("getMetricViewFieldsAndGroupFields[%s] err: %s", model.ModelID, err.Error()))
				}

				// 衍生指标，转换分析维度和排序字段，以及构造指标的字段map
				transferDimAndOrder(ctx, &model, viewFields, fieldsMap)
			}

			mmfilters = append(mmfilters, interfaces.MetricModelWithFilters{
				MetricModel: model,
			})
		case interfaces.COMPOSITED_METRIC:
			// 分析维度是公共维度，所以在第一个模型里一定有复合指标的分析维度，若没有，则返回空，需要修正
			modelIDs := common.ExtractModelIDs(model.Formula)
			// 校验model存在、分析维度的交集(复合指标的分析维度需在所有的依赖指标模型中)
			dependModels, err := mms.GetMetricModels(ctx, modelIDs, false)
			if err != nil {
				return mmfilters, err
			}
			if len(dependModels) != len(modelIDs) {
				// 查询时，如果数据视图已经被删除，则数据视图字段为空
				o11y.Warn(ctx, fmt.Sprintf("Exist any depend metric model %v not found", modelIDs))
			} else {
				// 分析维度是公共维度，所以在第一个模型里一定有复合指标的分析维度，若没有，则返回空，需要修正
				// dependsDimMap := map[string]interfaces.Field{}
				// for _, dim := range dependModels[0].AnalysisDims {
				// 	dependsDimMap[dim.Name] = dim
				// }

				// 排序字段是依赖的模型的可选排序字段的交集，所以取第一个模型即可
				viewFields, fieldsMap, err := mms.getMetricViewFieldsAndGroupFields(ctx, model)
				if err != nil {
					// 查询时报错，记录日志
					o11y.Warn(ctx, fmt.Sprintf("getMetricViewFieldsAndGroupFields[%s] err: %s", model.ModelID, err.Error()))
				}

				// 复合指标，转换分析维度和排序字段，以及构造指标的字段map
				transferDimAndOrder(ctx, &model, viewFields, fieldsMap)
			}

			mmfilters = append(mmfilters, interfaces.MetricModelWithFilters{
				MetricModel: model,
			})
		}
	}
	dataViewSpan.SetAttributes(attr.Key("data_source_ids").StringSlice(dvIDs),
		attr.Key("data_source_types").StringSlice(dsTypes),
		attr.Key("fields_nums").IntSlice(fieldsNum))
	dataViewSpan.SetStatus(codes.Ok, "")
	dataViewSpan.End()

	span.SetStatus(codes.Ok, "")
	return mmfilters, nil
}

// 计算公式有效性检查
func (mms *metricModelService) CheckFormula(ctx context.Context, metricModel interfaces.MetricModel) error {
	// 计算公式有效性检查请求的是最近30分钟，step为5min。 日历间隔的查询，生成的步长
	// now := time.Now()
	query := interfaces.MetricModelQuery{
		MetricType:      metricModel.MetricType,
		DataSource:      metricModel.DataSource,
		QueryType:       metricModel.QueryType,
		Formula:         metricModel.Formula,
		MeasureField:    metricModel.MeasureField,
		OrderByFields:   metricModel.OrderByFields,
		HavingCondition: metricModel.HavingCondition,
		// IsInstantQuery: false,
		// Start:          now.Add(-30 * time.Minute).UnixMilli(),
		// End:            now.UnixMilli(),
		// Step:           "5m",
		IsInstantQuery: true,
		LookBackDelta:  "10m",
		IsModelRequest: true,
	}
	if metricModel.IsCalendarInterval == 1 {
		query.Step = "minute"
	}
	valid, errStr, err := mms.ua.CheckFormulaByUniquery(ctx, query)
	if err != nil {
		logger.Errorf("CheckFormulaByUniquery error: %s", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_CheckFormulaFailed).WithErrorDetails(err.Error())
	}
	if !valid {
		logger.Errorf("Formula invalid: %s", errStr)
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_MetricModel_InvalidParameter_Formula).WithErrorDetails(errStr)
	}
	return nil
}

// 计算公式有效性检查
func (mms *metricModelService) CheckSqlFormulaConfig(ctx context.Context, metricModel interfaces.MetricModel) error {
	// 计算公式有效性检查请求的是最近30分钟，step为5min。 日历间隔的查询，生成的步长
	// 计算公式有效性校验时不需要传递分析维度
	query := interfaces.MetricModelQuery{
		MetricType:      metricModel.MetricType,
		DataSource:      metricModel.DataSource,
		QueryType:       metricModel.QueryType,
		FormulaConfig:   metricModel.FormulaConfig,
		DateField:       metricModel.DateField,
		OrderByFields:   metricModel.OrderByFields,
		HavingCondition: metricModel.HavingCondition,
		IsInstantQuery:  true,
		LookBackDelta:   "10m",
		IsModelRequest:  true,
	}

	valid, errStr, err := mms.ua.CheckFormulaByUniquery(ctx, query)
	if err != nil {
		logger.Errorf("CheckFormulaByUniquery error: %s", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_CheckFormulaFailed).WithErrorDetails(err.Error())
	}
	if !valid {
		logger.Errorf("Formula invalid: %s", errStr)
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_MetricModel_InvalidParameter_Formula).WithErrorDetails(errStr)
	}
	return nil
}

func (mms *metricModelService) checkMeasureFieldType(ctx context.Context, dataViewID string, measureField string) error {
	// dsl 时校验 top_hits 的度量字段的字段类型是否为数值，其他情况不校验。
	ctx, span := ar_trace.Tracer.Start(ctx, "校验 top_hits 的度量字段")
	span.SetAttributes(attr.Key("data_view_id").String(dataViewID))
	defer span.End()

	dataView, err := mms.dvs.GetDataView(ctx, dataViewID)
	// 记录模块调用返回日志
	o11y.Info(ctx, fmt.Sprintf("GetDataView, request: %s; response: error: %v}", dataViewID, err))

	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置获取过滤条件的 span 的状态
		span.SetStatus(codes.Error, fmt.Sprintf("Get data view [%s] filters error", dataViewID))

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get data view [%s] filters error: %s. %v", dataViewID,
			httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_GetDataViewQueryFiltersFailed).WithErrorDetails(httpErr)
	}
	// 此处的map可能取到的是nil，需先取，存在才取type
	field, exist := dataView.FieldsMap[measureField]
	if !exist {
		// 字段不存在
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_MetricModel_InvalidParameter_MeasureFieldNotExists).
			WithErrorDetails(fmt.Sprintf("The measure field [%s] not exists in data view[%s]", measureField, dataView.ViewName))
	}
	mType := field.Type
	_, exists := interfaces.MEASURE_FIELD_TYPE[mType]
	if !exists {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_MeasureField).
			WithErrorDetails(fmt.Sprintf("The type of measure field[%s] expect to a number, actual is %s",
				measureField, mType))

		span.SetStatus(codes.Error, fmt.Sprintf("The type of measure field[%s] is not a number", measureField))
		span.End()

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf(" %s. The type of measure field[%s] expect to a number, actual is %s.",
			httpErr.BaseError.Description, measureField, mType))

		return httpErr
	}
	span.SetStatus(codes.Ok, "")
	return nil
}

func (mms *metricModelService) checkIndexBase(ctx context.Context, indexBaseType string) error {
	// 校验 请求体中的数据视图是否存在
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("获取索引库[%s]信息", indexBaseType))
	span.SetAttributes(attr.Key("index_base_type").String(indexBaseType))
	defer span.End()

	indexbase, err := mms.iba.GetSimpleIndexBasesByTypes(ctx, []string{indexBaseType})
	// 记录模块调用返回日志
	o11y.Info(ctx, fmt.Sprintf("GetSimpleIndexBasesByTypes, request: %s; response: {indexbases: %v, error: %v}",
		indexBaseType, indexbase, err))

	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_GetSimpleIndexBasesByTypesFailed).WithErrorDetails(err.Error())

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

// 根据指标模型IDs获取指标模型的id、名称、组名、单位类型、单位、是否内置。事件模型、目标模型调此函数获取简单信息
func (mms *metricModelService) GetMetricModelSimpleInfosByIDs(ctx context.Context, modelIDs []string) (map[string]interfaces.SimpleMetricModel, error) {
	// id去重后再查
	mIDs := common.DuplicateSlice(modelIDs)

	modelMap, err := mms.mma.GetMetricModelSimpleInfosByIDs(ctx, mIDs)
	if err != nil {
		return nil, err
	}
	// 不报错，返回空即可
	if len(modelMap) != len(mIDs) {
		return map[string]interfaces.SimpleMetricModel{}, nil
	}

	// 校验查看权限
	// 先获取资源序列
	matchResoucesMap, err := mms.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_METRIC_MODEL, mIDs,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, false)
	if err != nil {
		return nil, err
	}

	// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
	for _, mID := range modelIDs {
		if _, exist := matchResoucesMap[mID]; !exist {
			return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for metric model's view_detail operation.")
		}
	}

	return modelMap, nil
}

// // 根据指标模型名称获取IDs
// func (mms *metricModelService) GetMetricModelIDsBySimpleInfos(ctx context.Context, simpleInfos []interfaces.ModelSimpleInfo) (map[interfaces.ModelSimpleInfo]string, error) {
// 	return mms.mma.GetMetricModelIDsBySimpleInfos(ctx, simpleInfos)
// }

// // 根据指标模型IDs获取名称，仅给结构模型使用，待结构模型下线，此函数删掉
// func (mms *metricModelService) GetMetricModelSimpleInfosByIDs2(ctx context.Context, modelIDs []string) (map[string]interfaces.ModelSimpleInfo, error) {
// return mms.mma.GetMetricModelSimpleInfosByIDs2(ctx, modelIDs)
// }

// 根据groupID 获取指标模型信息（批量修改模型groupID是调用，此函数不用加权限）
func (mms *metricModelService) GetMetricModelsByGroupID(ctx context.Context,
	groupID string) ([]interfaces.MetricModel, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("查询指标模型分组[%s]内指标模型信息", groupID))
	span.SetAttributes(attr.Key("group_id").String(groupID))
	defer span.End()

	emptyMetricModels := make([]interfaces.MetricModel, 0)
	metricModels, err := mms.mma.GetMetricModelsByGroupID(ctx, groupID)
	if err != nil {
		logger.Errorf("GetMetricModelsByGroupID error: %s", err.Error())
		span.SetStatus(codes.Error, "Get metric models by groupID error")
		span.End()
		return emptyMetricModels, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return metricModels, nil

}

// （创建/导入指标模型时）根据groupName更新GroupID
func (mms *metricModelService) RetriveGroupIDByGroupName(ctx context.Context, tx *sql.Tx, groupName string) (interfaces.MetricModelGroup, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("查询指标模型分组[%s]的分组ID", groupName))
	span.SetAttributes(attr.Key("group_name").String(groupName))
	defer span.End()

	//查询groupName是否存在
	metricGroup, exist, err := mms.mmga.GetMetricModelGroupByName(ctx, tx, groupName)
	if err != nil {
		logger.Errorf("CheckMetricModelGroupExistByName error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("按名称[%s]获取指标模型分组失败", groupName))
		return metricGroup, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModelGroup_InternalError_GetMetricModelGroupIDByNameFailed).WithErrorDetails(err.Error())
	}
	if exist {
		// 存在，校验组的查看权限
		// 判断userid是否有创建指标模型的权限（策略决策）
		// err = mms.ps.CheckPermission(ctx, interfaces.ACCESSOR_TYPE_USER,
		// 	fmt.Sprintf("%s%s", interfaces.METRIC_MODEL_GROUP_RESOURCE_ID_PREFIX, metricGroup.GroupID),
		// 	interfaces.RESOURCE_TYPE_METRIC_MODEL, []string{interfaces.PERMISSION_TYPE_VIEW_DETAIL})
		// if err != nil {
		// 	return metricGroup, err
		// }
		return metricGroup, nil
	}

	//不存在，需要创建分组并返回groupID
	// 校验组的创建权限: 无需单独校验创建组的权限，前面已经校验了创建权限
	// 生成分布式ID
	groupID := xid.New().String()

	metricModelGroup := interfaces.MetricModelGroup{
		GroupID:    groupID,
		GroupName:  groupName,
		UpdateTime: time.Now().UnixMilli(),
	}

	//调用driven层创建指标模型分组
	createCtx, createSpan := ar_trace.Tracer.Start(ctx, "Create metric model group")
	createSpan.SetAttributes(
		attr.Key("metric_model_group_id").String(groupID),
		attr.Key("metric_model_group_name").String(metricModelGroup.GroupName))
	defer createSpan.End()

	err = mms.mmga.CreateMetricModelGroup(createCtx, tx, metricModelGroup)
	if err != nil {
		logger.Errorf("CreateMetricModelGroup error: %s", err.Error())
		createSpan.SetStatus(codes.Error, "创建指标模型分组失败")

		return metricModelGroup, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModelGroup_InternalError).
			WithErrorDetails(err.Error())
	}
	createSpan.SetStatus(codes.Ok, "")

	// 注册当前组的策略
	// err = mms.ps.CreatePolicy(ctx, interfaces.ACCESSOR_TYPE_USER,
	// 	fmt.Sprintf("%s%s", interfaces.METRIC_MODEL_GROUP_RESOURCE_ID_PREFIX, metricModelGroup.GroupID),
	// 	interfaces.RESOURCE_TYPE_METRIC_MODEL, metricModelGroup.GroupName,
	// 	"",
	// 	interfaces.COMMON_OPERATION)
	// if err != nil {
	// 	return metricModelGroup, err
	// }

	return metricModelGroup, nil
}

// 根据名称获取指标模型ID
func (mms *metricModelService) GetMetricModelIDByName(ctx context.Context, groupName string, modelName string) (string, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("根据分组名称[%s]和指标模型名称[%s]查询指标模型ID", groupName, modelName))
	span.SetAttributes(
		attr.Key("group_name").String(groupName),
		attr.Key("model_name").String(modelName))
	defer span.End()

	modelID, exist, err := mms.mma.GetMetricModelIDByName(ctx, groupName, modelName)
	if err != nil {
		logger.Errorf("GetMetricModelIDByName error: %s", err.Error())
		span.SetStatus(codes.Error, "Get metric model id  by group name and model name error")
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_GetModelIDByNameFailed).WithErrorDetails(err.Error())
	}
	if !exist {
		logger.Debugf("Metric Model  %s not found!", modelName)
		span.SetStatus(codes.Error, fmt.Sprintf("Metric model [%s] not found", modelName))
		return "", rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_MetricModel_MetricModelNotFound)
	}

	span.SetStatus(codes.Ok, "")
	return modelID, nil
}

// 批量修改模型groupID
func (mms *metricModelService) UpdateMetricModelsGroup(ctx context.Context,
	modelsMap map[string]interfaces.SimpleMetricModel, group interfaces.MetricModelGroupName) (rowsAffect int64, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Update metric models")
	defer span.End()

	// 0. 开始事务
	tx, err := mms.db.Begin()
	if err != nil {
		logger.Errorf("Begin transaction error: %s", err.Error())
		span.SetStatus(codes.Error, "事务开启失败")
		o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

		return 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModel_InternalError_BeginTransactionFailed).
			WithErrorDetails(err.Error())
	}
	// 0.1 异常时
	defer func() {
		switch err {
		case nil:
			// 提交事务
			err = tx.Commit()
			if err != nil {
				logger.Errorf("UpdateMetricModelsGroup Transaction Commit Failed:%v", err)
				span.SetStatus(codes.Error, "提交事务失败")
				o11y.Error(ctx, fmt.Sprintf("UpdateMetricModelsGroup Transaction Commit Failed: %s", err.Error()))

			}
			logger.Info("UpdateMetricModelsGroup Transaction Commit Success")
			o11y.Debug(ctx, "UpdateMetricModelsGroup Transaction Commit Success")
		default:
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logger.Errorf("UpdateMetricModelsGroup Transaction Rollback Error:%v", rollbackErr)
				span.SetStatus(codes.Error, "事务回滚失败")
				o11y.Error(ctx, fmt.Sprintf("UpdateMetricModelsGroup Transaction Rollback Error: %s", err.Error()))
			}
		}
	}()

	// 根据分组名称获取分组 ID，如果不存在，则创建分组
	groupInfo, httpErr := mms.RetriveGroupIDByGroupName(ctx, tx, group.GroupName)
	if httpErr != nil {
		logger.Errorf("RetriveGroupIDByGroupName error: %s", err.Error())
		span.SetStatus(codes.Error, "根据groupName更新GroupID失败")
		return 0, httpErr
	}

	// 校验当前模型的修改权限.组没有权限,所以校验当前模型的权限即可.
	// 先获取资源列表
	modelIDs := make([]string, 0, len(modelsMap))
	// 如果移入的视图和分组内名称重复，则操作不成功，给出提示
	for _, newModel := range modelsMap {
		modelIDs = append(modelIDs, newModel.ModelID)
	}
	// id去重后再查
	mIDs := common.DuplicateSlice(modelIDs)
	matchResouces, err := mms.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_METRIC_MODEL, mIDs,
		[]string{interfaces.OPERATION_TYPE_MODIFY}, false)
	if err != nil {
		return 0, err
	}

	// 资源过滤后的数量跟请求的数量不等，说明有部分模型没有权限，不能删除
	if len(matchResouces) != len(mIDs) {
		return 0, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: insufficient permissions for metric model's creation operation.")
	}

	//获取目标分组内的指标模型
	existModels, err := mms.GetMetricModelsByGroupID(ctx, groupInfo.GroupID)
	if err != nil {
		logger.Errorf("GetMetricModelsByGroupID error: %s", err.Error())
		span.SetStatus(codes.Error, "Get metric models by groupID error")
		return 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError).WithErrorDetails(err.Error())
	}

	// 如果移入的视图和分组内名称重复，则操作不成功，给出提示
	for _, newModel := range modelsMap {
		// 内置对象只能移到内置分组里，非内置对象只能移动非内置分组里
		if newModel.Builtin != groupInfo.Builtin {
			errDetail := "The built-in model can only be placed in a built-in group, while the non-built-in model can only be placed in a non-built-in group."
			logger.Errorf(errDetail)
			span.SetStatus(codes.Error, "指标模型的内置属性与所绑定的分组的内置属性不匹配")
			return 0, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_MetricModel_InvalidParameter_Group).
				WithErrorDetails(errDetail)
		}
		for _, oldModel := range existModels {

			if newModel.ModelName == oldModel.ModelName {
				errDetails := fmt.Sprintf("Metric model '%s' already exsited in group '%s'", newModel.ModelName, group.GroupName)
				logger.Error(errDetails)
				span.SetStatus(codes.Error, errDetails)
				return 0, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_CombinationNameExisted).
					WithDescription(map[string]any{"ModelName": newModel.ModelName, "GroupName": group.GroupName}).
					WithErrorDetails(errDetails)
			}
		}
	}

	rowsAffect, err = mms.mma.UpdateMetricModelsGroupID(ctx, tx, modelIDs, groupInfo.GroupID)
	span.SetAttributes(attr.Key("rows_affect").Int64(rowsAffect))
	if err != nil {
		logger.Errorf("UpdateMetricModels error: %s", err.Error())
		span.SetStatus(codes.Error, "修改指标模型的分组信息失败")

		return rowsAffect, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModelGroup_InternalError).WithErrorDetails(err.Error())
	}
	logger.Infof("UpdateMetricModels: Rows affected is %v, request update modelIDs is %v!", rowsAffect, len(modelIDs))
	if rowsAffect != int64(len(modelIDs)) {
		logger.Warnf("Update models number %v not equal request models number %v!", rowsAffect, len(modelIDs))

		o11y.Warn(ctx, fmt.Sprintf("Update models number %v not equal request models number %v!", rowsAffect, len(modelIDs)))
	}

	// 批量改模型的分组，不改模型的名称，不用调更新资源名称的接口

	span.SetStatus(codes.Ok, "")
	return rowsAffect, nil
}

// 按模型id校验模型的存在性
func (mms *metricModelService) CheckMetricModelExistByID(ctx context.Context, modelID string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("根据ID[%s]校验指标模型的存在性", modelID))
	span.SetAttributes(
		attr.Key("model_od").String(modelID))
	defer span.End()

	modelID, exist, err := mms.mma.CheckMetricModelExistByID(ctx, modelID)
	if err != nil {
		logger.Errorf("CheckMetricModelExistByID error: %s", err.Error())

		span.SetStatus(codes.Error, fmt.Sprintf("按ID[%s]获取指标模型失败", modelID))
		// 记录处理的 sql 字符串
		o11y.Error(ctx, fmt.Sprintf("按ID[%s]获取指标模型失败: %v", modelID, err))

		return modelID, exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_CheckModelIfExistFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return modelID, exist, nil
}

func (mms *metricModelService) CheckMetricModelExistByName(ctx context.Context, groupName, modelName string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("校验指标模型[%s/%s]的存在性", groupName, modelName))
	span.SetAttributes(
		attr.Key("group_name").String(groupName),
		attr.Key("model_name").String(modelName))
	defer span.End()

	modelID, exist, err := mms.mma.GetMetricModelIDByName(ctx, groupName, modelName)
	if err != nil {
		logger.Errorf("GetMetricModelIDByName error: %s", err.Error())

		span.SetStatus(codes.Error, fmt.Sprintf("按名称[%s/%s]获取指标模型失败", groupName, modelName))
		// 记录处理的 sql 字符串
		o11y.Error(ctx, fmt.Sprintf("按名称[%s/%s]获取指标模型失败: %v", groupName, modelName, err))

		return modelID, exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_CheckModelIfExistFailed).WithErrorDetails(err.Error())
	}
	// if exist {
	// 	logger.Errorf("MetricModel %v already exist!", combinationName)

	// 	span.SetStatus(codes.Error, fmt.Sprintf("指标模型[%v]已存在", combinationName))
	// 	// 记录处理的 sql 字符串
	// 	o11y.Error(ctx, fmt.Sprintf("指标模型[%v]已存在", combinationName))

	// 	return exist, nil
	// }

	span.SetStatus(codes.Ok, "")
	return modelID, exist, nil
}

// 根据度量名称获取指标模型来校验度量名称的唯一性
func (mms *metricModelService) CheckMetricModelByMeasureName(ctx context.Context, measureName string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("校验指标模型的度量名称[%s]的唯一性", measureName))
	span.SetAttributes(attr.Key("measure_name").String(measureName))
	defer span.End()

	modelID, exist, err := mms.mma.CheckMetricModelByMeasureName(ctx, measureName)
	if err != nil {
		logger.Errorf("CheckMetricModelByMeasureName error: %s", err.Error())

		span.SetStatus(codes.Error, fmt.Sprintf("校验度量名称[%s]唯一性失败", measureName))
		// 记录日志
		o11y.Error(ctx, fmt.Sprintf("校验度量名称[%s]唯一性失败: %v", measureName, err))

		return modelID, exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_CheckDuplicateMeasureNameFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return modelID, exist, nil
}

// 根据模型id获取模型所绑定的数据源的字段列表
func (mms *metricModelService) GetMetricModelSourceFields(ctx context.Context, modelID string) ([]*interfaces.ViewField, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("查询指标模型[%s]信息", modelID))
	span.SetAttributes(attr.Key("model_id").String(modelID))
	defer span.End()

	// 校验查询权限
	err := mms.ps.CheckPermission(ctx, interfaces.Resource{
		ID:   modelID,
		Type: interfaces.RESOURCE_TYPE_METRIC_MODEL},
		[]string{interfaces.OPERATION_TYPE_DATA_QUERY})
	if err != nil {
		return nil, err
	}

	// 1. 根据模型id获取模型信息
	model, exist, err := mms.mma.GetMetricModelByModelID(ctx, modelID)
	if err != nil {
		logger.Errorf("GetMetricModelByModelID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get metric model[%s] error: %v", modelID, err))
		return []*interfaces.ViewField{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_GetModelByIDFailed).WithErrorDetails(err.Error())
	}
	if !exist {
		logger.Debugf("Metric Model %s not found!", modelID)
		span.SetStatus(codes.Error, fmt.Sprintf("Metric model[%s] not found", modelID))
		return []*interfaces.ViewField{}, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_MetricModel_MetricModelNotFound)
	}

	// 获取数据源（数据视图）字段列表
	dataViewQueryFilters, err := mms.dvs.GetDataView(ctx, model.DataSource.ID)
	if err != nil {
		logger.Errorf("GetDataView error: %s", err.Error())
		span.SetStatus(codes.Error, "获取数据视图过滤条件失败")
		o11y.Error(ctx, fmt.Sprintf("Get data view filters failed, dataViewID: %s, error: %v ", model.DataSource.ID, err))

		return []*interfaces.ViewField{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_GetDataViewQueryFiltersFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return dataViewQueryFilters.Fields, nil
}

// 根据分组获取模型详细信息，用于导出
func (mms *metricModelService) GetMetricModelsDetailByGroupID(ctx context.Context, groupID string) ([]interfaces.MetricModel, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get metric models by group id")
	defer span.End()

	models, err := mms.mma.GetMetricModelsByGroupID(ctx, groupID)
	if err != nil {
		logger.Errorf("GetMetricModelsByGroupID error: %s", err.Error())
		span.SetStatus(codes.Error, "Get metric models by groupID error")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError).WithErrorDetails(err.Error())
	}

	modelIDs := make([]string, 0)
	for _, model := range models {
		modelIDs = append(modelIDs, model.ModelID)
	}

	// 校验组下的模型有查看权限
	matchResouces, err := mms.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_METRIC_MODEL, modelIDs,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, false)
	if err != nil {
		return nil, err
	}

	// 资源过滤后的数量跟请求的数量不等，说明有部分模型没有权限，不能导出
	if len(matchResouces) != len(modelIDs) {
		return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: insufficient permissions for metric model's view_detail operation.")
	}

	// 根据模型id获取任务信息
	modelTaskMap, err := mms.mmts.GetMetricTasksByModelIDs(ctx, modelIDs)
	if err != nil {
		logger.Errorf("GetMetricTasksByModelIDs error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get metric model[%s] error: %v", modelIDs, err))
		span.End()

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_GetMetricTasksByModelIDsFailed).WithErrorDetails(err.Error())
	}

	// 数据视图 id 换成名称，需要循环遍历查询。
	metricModels := make([]interfaces.MetricModel, 0)
	for _, model := range models {
		task, taskExist := modelTaskMap[model.ModelID]
		if taskExist {
			model.Task = &task
			// 若存在任务，则把任务中的索引库类型换索引库名称
			// 获取索引库名称
			simpleIndexBases, err := mms.iba.GetSimpleIndexBasesByTypes(ctx, []string{model.Task.IndexBase})
			if err != nil {
				logger.Errorf("GetSimpleIndexBasesByTypes error: %s", err.Error())

				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					derrors.DataModel_MetricModel_InternalError_GetSimpleIndexBasesByTypesFailed).WithErrorDetails(err.Error())
			}
			// 若索引库查询为空，则不赋值
			if len(simpleIndexBases) > 0 {
				// 遍历模型任务，把索引库名称赋值到任务上
				model.Task.IndexBaseName = simpleIndexBases[0].Name
			}
		}

		// 导入的时候只认id，不认名称。
		switch model.MetricType {
		case interfaces.ATOMIC_METRIC:
			viewName, exist, err := mms.dvs.CheckDataViewExistByID(ctx, nil, model.DataSource.ID)
			if err != nil {
				logger.Errorf("GetDataViewByID error: %s", err.Error())
				span.SetStatus(codes.Error, fmt.Sprintf("Get data view[%s] error", model.DataSource.ID))
				span.End()

				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					derrors.DataModel_MetricModel_InternalError_GetDataViewByIDFailed).WithErrorDetails(err.Error())
			}
			if !exist {
				// 查询时，如果数据视图已经被删除，则数据视图字段为空
				o11y.Warn(ctx, fmt.Sprintf("Data view %s not found", model.DataSource.ID))
			} else {
				model.DataSource.Name = viewName
			}

		case interfaces.DERIVED_METRIC:
			// 按id获取名称
			derivedConfig := model.FormulaConfig.(interfaces.DerivedConfig)
			modelMap, err := mms.mma.GetMetricModelSimpleInfosByIDs(ctx, []string{derivedConfig.DependMetricModel.ID})
			if err != nil {
				logger.Errorf("CheckMetricModelExistByID error: %s", err.Error())

				span.SetStatus(codes.Error, fmt.Sprintf("按ID[%s]获取指标模型失败", derivedConfig.DependMetricModel.ID))
				// 记录处理的 sql 字符串
				o11y.Error(ctx, fmt.Sprintf("按ID[%s]获取指标模型失败: %v", derivedConfig.DependMetricModel.ID, err))

				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					derrors.DataModel_MetricModel_InternalError_CheckModelIfExistFailed).WithErrorDetails(err.Error())
			}
			if len(modelMap) != 1 {
				// 查询时，如果数据视图已经被删除，则数据视图字段为空
				o11y.Warn(ctx, fmt.Sprintf("Depend metric model %s not found", derivedConfig.DependMetricModel.ID))
			}
			derivedConfig.DependMetricModel.GroupName = modelMap[derivedConfig.DependMetricModel.ID].GroupName
			derivedConfig.DependMetricModel.Name = modelMap[derivedConfig.DependMetricModel.ID].ModelName
			model.FormulaConfig = derivedConfig

		case interfaces.COMPOSITED_METRIC:
			// 无转换 直接拼
		}

		metricModels = append(metricModels, model)
	}

	return metricModels, nil
}

// 校验vega视图的相关配置的有效性
func (mms *metricModelService) checkSQLView(ctx context.Context, metricModel interfaces.MetricModel, dataView *interfaces.DataView) error {
	// 1. 请求vega逻辑视图获取视图字段

	// 2. 校验时间字段是否属于逻辑视图，时间字段非必填
	if _, exist := dataView.FieldsMap[metricModel.DateField]; metricModel.DateField != "" && !exist {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_DateFieldNotExisted).
			WithErrorDetails(fmt.Sprintf("date field[%s] not exists in vega logic view[%s]", metricModel.DateField, metricModel.DataSource.ID))
	}

	// 分组字段存储的是[]string，交互查询用的是[]Field
	// 3. 校验分组字段是否属于逻辑视图
	sqlConfig := metricModel.FormulaConfig.(interfaces.SQLConfig)
	for _, field := range sqlConfig.GroupByFields {
		if _, exist := dataView.FieldsMap[field]; !exist {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_GroupByFieldNotExisted).
				WithErrorDetails(fmt.Sprintf("group by field[%s] not exists in vega logic view[%s]", field, metricModel.DataSource.ID))
		}
	}

	// 5. 校验聚合字段是否属于逻辑视图
	if sqlConfig.AggrExpr != nil {
		if _, exist := dataView.FieldsMap[sqlConfig.AggrExpr.Field]; !exist {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_AggregationNotExisted).
				WithErrorDetails(fmt.Sprintf("aggregation field[%s] not exists in vega logic view[%s]", sqlConfig.AggrExpr.Field, metricModel.DataSource.ID))
		}
	}

	// 4. 校验分析维度是否属于逻辑视图
	for _, field := range metricModel.AnalysisDims {
		if _, exist := dataView.FieldsMap[field.Name]; !exist {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_AnalysisDimensionNotExisted).
				WithErrorDetails(fmt.Sprintf("analysis dimension[%s] not exists in vega logic view[%s]", field.Name, metricModel.DataSource.ID))
		}
	}

	return nil
}

// 指标模型数据来源、计算公式有效性和持久化任务校验
func (mms *metricModelService) processForCreate(ctx context.Context, metricModel *interfaces.MetricModel, currentTime int64) error {

	ctx, span := ar_trace.Tracer.Start(ctx, "Create metric models")
	defer span.End()

	// 原子指标、衍生指标、和复合指标的校验不同，需要兼容
	err := mms.checkDepends(ctx, metricModel)
	if err != nil {
		return err
	}

	metricModel.UpdateTime = currentTime
	metricModel.CreateTime = currentTime

	// 度量名称为空，则赋值模型id
	if metricModel.MeasureName == "" {
		metricModel.MeasureName = fmt.Sprintf("%s%s", interfaces.MEASURE_PREFIX, metricModel.ModelID)
	}

	// 把任务id、模型id和更新时间赋值到任务对象中
	if metricModel.Task != nil {
		// 校验 请求体中的索引库类型是否存在
		err := mms.checkIndexBase(ctx, metricModel.Task.IndexBase)
		if err != nil {
			return err
		}

		// 生成分布式ID
		taskID := xid.New().String()
		metricModel.Task.TaskID = taskID

		metricModel.Task.ModuleType = interfaces.MODULE_TYPE_METRIC_MODEL
		metricModel.Task.ModelID = metricModel.ModelID
		metricModel.Task.MeasureName = metricModel.MeasureName
		metricModel.Task.UpdateTime = currentTime
		// 时间窗口对应的计划时间
		// 追溯时长转毫秒，用于计算计划时间
		durationV := int64(0)
		if metricModel.Task.RetraceDuration != "" {
			dur, err := common.ParseDuration(metricModel.Task.RetraceDuration, common.DurationDayHourRE, false)
			if err != nil {
				logger.Errorf("Failed to parse retrace duration, err: %v", err.Error())
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_RetraceDuration).
					WithErrorDetails(err.Error())
			}
			durationV = int64(dur / (time.Millisecond / time.Nanosecond))
		}
		planTime := time.Now().UnixNano()/int64(time.Millisecond/time.Nanosecond) - durationV

		metricModel.Task.PlanTime = planTime
		// 创建，创建的任务状态为完成
		metricModel.Task.ScheduleSyncStatus = interfaces.SCHEDULE_SYNC_STATUS_FINISH
	}

	span.SetAttributes(
		attr.Key("metric_model_id").String(metricModel.ModelID),
		attr.Key("metric_model_name").String(metricModel.ModelName),
		attr.Key("group_name").String(metricModel.GroupName),
	)

	span.SetStatus(codes.Ok, "")
	return nil
}

func (mms *metricModelService) handleMetricModelImportMode(ctx context.Context, mode string,
	models []*interfaces.MetricModel) ([]*interfaces.MetricModel, []*interfaces.MetricModel, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "metric model import mode logic")
	defer span.End()

	createModels := []*interfaces.MetricModel{}
	updateModels := []*interfaces.MetricModel{}

	// 调用当前 service 的校验放在handler中,调用其他service做的校验放在当前handler对应的service中
	// 3. 校验 若指标模型的id不为空，则用请求体的id与现有指标模型ID的重复性
	for i := 0; i < len(models); i++ {
		createModels = append(createModels, models[i])
		_, idExist, err := mms.CheckMetricModelExistByID(ctx, models[i].ModelID)
		if err != nil {
			return createModels, updateModels, err
		}

		// 校验 请求体与现有指标模型名称的重复性
		existID, nameExist, err := mms.CheckMetricModelExistByName(ctx, models[i].GroupName, models[i].ModelName)
		if err != nil {
			return createModels, updateModels, err
		}

		// 5. 校验度量名称的唯一性
		measureNameExist := false
		modelIDByMeaserName := models[i].ModelID
		if models[i].MeasureName != "" {
			modelIDByMeaserName, measureNameExist, err = mms.CheckMetricModelByMeasureName(ctx, models[i].MeasureName)
			if err != nil {
				return createModels, updateModels, err
			}
		}

		// 根据mode来区别，若是ignore，就从结果集中忽略，若是overwrite，就调用update，若是normal就报错。
		if idExist || nameExist || measureNameExist {
			switch mode {
			case interfaces.ImportMode_Normal:
				if idExist {
					errDetails := fmt.Sprintf("The metric model with id [%s] already exists!", models[i].ModelID)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return createModels, updateModels, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_IDExisted).
						WithErrorDetails(errDetails)
				}

				if nameExist {
					errDetails := fmt.Sprintf("metric model name '%s' already exists in group '%s'", models[i].ModelName, models[i].GroupName)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return createModels, updateModels, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_MetricModel_CombinationNameExisted).
						WithDescription(map[string]any{"ModelName": models[i].ModelName, "GroupName": models[i].GroupName}).
						WithErrorDetails(errDetails)
				}

				if measureNameExist {
					errDetails := fmt.Sprintf("metric model measure name '%s' already exists", models[i].MeasureName)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return createModels, updateModels, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_MetricModel_Duplicated_MeasureName).
						WithErrorDetails(errDetails)
				}

			case interfaces.ImportMode_Ignore:
				// 存在重复的就跳过
				// 从create数组中删除
				createModels = createModels[:len(createModels)-1]
			case interfaces.ImportMode_Overwrite:
				models[i].MeasureName = modelIDByMeaserName
				if idExist && nameExist {
					// 如果 id 和名称都存在，但是存在的名称对应的视图 id 和当前视图 id 不一样，则报错
					if existID != models[i].ModelID {
						errDetails := fmt.Sprintf("MetricModel ID '%s' and name '%s/%s' already exist, but the exist model id is '%s'",
							models[i].ModelID, models[i].GroupName, models[i].ModelName, existID)
						logger.Error(errDetails)
						span.SetStatus(codes.Error, errDetails)
						return createModels, updateModels, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_MetricModel_CombinationNameExisted).
							WithErrorDetails(errDetails)
					} else if measureNameExist && modelIDByMeaserName != models[i].ModelID {
						// 名称对应的id和提交的id相同，度量名称存在，且度量名称对应的id和提交的id不同，报错
						errDetails := fmt.Sprintf("MetricModel ID '%s' with measureName '%s' already exist, but the exist model id is '%s'",
							models[i].ModelID, models[i].MeasureName, modelIDByMeaserName)
						logger.Error(errDetails)
						span.SetStatus(codes.Error, errDetails)
						return createModels, updateModels, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_MetricModel_Duplicated_MeasureName).
							WithErrorDetails(errDetails)
					} else {
						// 如果 id 和名称都存在，存在的名称对应的视图 id 和当前视图 id 一样，则覆盖更新
						// 从create数组中删除, 放到更新数组中
						createModels = createModels[:len(createModels)-1]
						updateModels = append(updateModels, models[i])
					}
				}

				// id 已存在，且名称不存在，覆盖更新
				if idExist && !nameExist {
					if measureNameExist && modelIDByMeaserName != models[i].ModelID {
						// 名称对应的id和提交的id相同，度量名称存在，且度量名称对应的id和提交的id不同，报错
						errDetails := fmt.Sprintf("MetricModel ID '%s' with measureName '%s' already exist, but the exist model id is '%s'",
							models[i].ModelID, models[i].MeasureName, modelIDByMeaserName)

						return createModels, updateModels, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_MetricModel_Duplicated_MeasureName).
							WithErrorDetails(errDetails)
					} else {
						// 从create数组中删除, 放到更新数组中
						createModels = createModels[:len(createModels)-1]
						updateModels = append(updateModels, models[i])
					}
				}

				// 如果 id 不存在，name 存在，报错
				if !idExist && nameExist {
					errDetails := fmt.Sprintf("MetricModel ID '%s' does not exist, but name '%s/%s' already exists",
						models[i].ModelID, models[i].GroupName, models[i].ModelName)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return createModels, updateModels, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_MetricModel_CombinationNameExisted).
						WithErrorDetails(errDetails)
				}

				// 如果 id 不存在，name 不存在，度量名称存在，报错
				if !idExist && !nameExist && measureNameExist {
					// 名称对应的id和提交的id相同，度量名称存在，且度量名称对应的id和提交的id不同，报错
					errDetails := fmt.Sprintf("MetricModel ID '%s' does not exist, but measure name '%v' already exists",
						models[i].ModelID, models[i].MeasureName)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return createModels, updateModels, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_MetricModel_Duplicated_MeasureName).
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

func (mms *metricModelService) ListMetricModelSrcs(ctx context.Context, parameter interfaces.MetricModelsQueryParams) ([]interfaces.Resource, int, error) {
	listCtx, listSpan := ar_trace.Tracer.Start(ctx, "查询指标模型实例列表")
	listSpan.End()

	//获取指标模型列表（不分页，获取所有的指标模型)
	models, err := mms.mma.ListSimpleMetricModels(listCtx, parameter)
	emptyResources := []interfaces.Resource{}
	if err != nil {
		logger.Errorf("ListMetricModels error: %s", err.Error())
		listSpan.SetStatus(codes.Error, "List simple metric models error")
		listSpan.End()
		return emptyResources, 0, rest.NewHTTPError(listCtx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError).WithErrorDetails(err.Error())
	}
	if len(models) == 0 {
		return emptyResources, 0, nil
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	// 处理资源id
	resMids := make([]string, 0)
	for _, m := range models {
		resMids = append(resMids, m.ModelID)
	}
	// 校验权限管理的操作权限
	matchResoucesMap, err := mms.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_METRIC_MODEL, resMids,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, false)
	if err != nil {
		return emptyResources, 0, err
	}

	// 遍历对象
	results := make([]interfaces.Resource, 0)
	for _, model := range models {
		if _, exist := matchResoucesMap[model.ModelID]; exist {
			// 如果是未分组，组名是空，此时需要把其按语言翻译未分组
			name := common.ProcessUngroupedName(ctx, model.GroupName, model.ModelName)

			results = append(results, interfaces.Resource{
				ID:   model.ModelID,
				Type: interfaces.RESOURCE_TYPE_METRIC_MODEL,
				Name: name,
			})
		}
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

// 校验指标模型的依赖(依赖的视图,指标模型)
func (mms *metricModelService) checkDepends(ctx context.Context, metricModel *interfaces.MetricModel) error {
	switch metricModel.MetricType {
	case interfaces.ATOMIC_METRIC:
		// 原子指标时的校验
		viewCtx, viewSpan := ar_trace.Tracer.Start(ctx, fmt.Sprintf("获取数据视图[%s]信息", metricModel.DataSource.ID))
		viewSpan.SetAttributes(attr.Key("data_view_id").String(metricModel.DataSource.ID))
		defer viewSpan.End()

		dataView, err := mms.dvs.GetDataView(viewCtx, metricModel.DataSource.ID)
		o11y.Info(viewCtx, fmt.Sprintf("GetDataView, request: %s; response: {dataView: %v, error: %v}",
			metricModel.DataSource.ID, dataView, err))
		if err != nil {
			return err
		}
		viewSpan.SetStatus(codes.Ok, "")

		// 校验视图有没有数据查询权限
		hasDataQuery := false
		for _, op := range dataView.Operations {
			if op == interfaces.OPERATION_TYPE_DATA_QUERY {
				hasDataQuery = true
			}
		}
		if !hasDataQuery {
			// 报错
			return rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for data view's data-query operation")
		}

		metricModel.DataSource.Type = dataView.QueryType
		switch dataView.QueryType {
		case interfaces.QueryType_DSL, interfaces.QueryType_IndexBase:
			// dsl 时校验 top_hits 的度量字段的字段类型是否为数值，其他情况不校验。
			if metricModel.IfContainTopHits && metricModel.QueryType == interfaces.DSL {
				// 前面校验过dsl的数据源类型是data_view,所以这里可以直接使用
				err = mms.checkMeasureFieldType(ctx, metricModel.DataSource.ID, metricModel.MeasureField)
				if err != nil {
					return err
				}
			}

			// 计算公式有效性检查，请求 uniquery，与指标数据预览同一个接口
			err = mms.CheckFormula(ctx, *metricModel)
			if err != nil {
				return err
			}
		case interfaces.QueryType_SQL:
			err := mms.checkSQLView(ctx, *metricModel, dataView)
			if err != nil {
				return err
			}

			// 计算公式有效性检查，请求 uniquery，与指标数据预览同一个接口
			err = mms.CheckSqlFormulaConfig(ctx, *metricModel)
			if err != nil {
				return err
			}

			// 原子指标的排序字段：列表集来自于前面配置的分组字段 + 值字段
			groupByFieldMap := map[string]bool{
				interfaces.VALUE_FIELD_NAME: true,
			}
			for _, group := range metricModel.FormulaConfig.(interfaces.SQLConfig).GroupByFields {
				groupByFieldMap[group] = true
			}
			for _, order := range metricModel.OrderByFields {
				if !groupByFieldMap[order.Name] {
					// 不属于分组字段集、__value，则报错
					return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_OrderByFieldNotExisted).
						WithErrorDetails(fmt.Sprintf("The atomic metric model's order_by_field[%s] is not belong to the group fields of metric model[%s]!",
							order.Name, metricModel.ModelID))
				}
			}
		}
	case interfaces.DERIVED_METRIC:
		// 衍生指标校验依赖的原子指标的存在性
		err := mms.validDerivedMetricModel(ctx, metricModel)
		if err != nil {
			return err
		}
	case interfaces.COMPOSITED_METRIC:
		// 分析维度需是依赖指标的公共维度
		err := mms.validCompositeMetricModel(ctx, metricModel)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mms *metricModelService) validDerivedMetricModel(ctx context.Context, metricModel *interfaces.MetricModel) error {
	// 衍生指标校验依赖的原子指标的存在性
	dependModelMap, orderByFieldMap, err := mms.getOrderByFields(ctx, *metricModel)
	if err != nil {
		return err
	}
	dependModel := dependModelMap[metricModel.FormulaConfig.(interfaces.DerivedConfig).DependMetricModel.ID]

	// 判断依赖指标是原子指标
	if dependModel.MetricType != interfaces.ATOMIC_METRIC {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_MetricModelNotFound).
			WithErrorDetails(fmt.Sprintf("The Depend model's metric_type expected is atomic, actual is [%s]!", dependModel.MetricType))
	}

	// 当前只支持sql的衍生指标
	if dependModel.QueryType != interfaces.SQL {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_MetricModelNotFound).
			WithErrorDetails(fmt.Sprintf("The Depend model's query_type expected is sql, actual is [%s]!", dependModel.MetricType))
	}

	// 分析维度是依赖的原子指标的子集，即衍生指标配置的分析维度都应在原子指标的分析维度中
	dependDimsMap := map[string]bool{}
	for _, dimension := range dependModel.AnalysisDims {
		dependDimsMap[dimension.Name] = true
	}
	for _, dim := range metricModel.AnalysisDims {
		if !dependDimsMap[dim.Name] {
			// 不属于依赖指标模型的分析维度集，则报错
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_AnalysisDimensionNotExisted).
				WithErrorDetails(fmt.Sprintf("The metric model's analysis_dimensions[%s] is not belong to the dimensions of depend metric model[%s]!",
					dim.Name, dependModel.ModelID))
		}
	}

	for _, order := range metricModel.OrderByFields {
		if !orderByFieldMap[order.Name] {
			// 不属于分组字段集、__value，则报错
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_OrderByFieldNotExisted).
				WithErrorDetails(fmt.Sprintf("The derived metric model's order_by_field[%s] is not belong to the group fields of atomic metric model[%s]!",
					order.Name, dependModel.ModelID))
		}
	}

	// 衍生指标的指标单位与原子指标的单位一致，如果不一致，就修改为一致
	if metricModel.UnitType != dependModel.UnitType || metricModel.Unit != dependModel.Unit {
		metricModel.UnitType = dependModel.UnitType
		metricModel.Unit = dependModel.Unit
	}

	// 计算公式有效性检查，请求 uniquery，与指标数据预览同一个接口
	err = mms.CheckSqlFormulaConfig(ctx, *metricModel)
	if err != nil {
		return err
	}

	// 衍生指标是否是日历步长跟原子指标相同
	metricModel.IsCalendarInterval = dependModel.IsCalendarInterval
	return nil
}

func (mms *metricModelService) validCompositeMetricModel(ctx context.Context, metricModel *interfaces.MetricModel) error {
	// 获取复合指标的排序字段和依赖模型
	dependModelMap, orderByFieldMap, err := mms.getOrderByFields(ctx, *metricModel)
	if err != nil {
		return err
	}
	// 分析维度需是依赖指标的公共维度
	modelIDs := common.ExtractModelIDs(metricModel.Formula)
	// 校验model存在、分析维度的交集(复合指标的分析维度需在所有的依赖指标模型中)
	for _, modelID := range modelIDs {
		// 如果计算公式中的指标模型存在当前模型的id，则报错
		if modelID == metricModel.ModelID {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
				WithErrorDetails("composite metric's formula cannot contain itself!")
		}

		// 依赖模型的分析维度转map
		dependDimMap := map[string]bool{}
		for _, analysisDim := range dependModelMap[modelID].AnalysisDims {
			dependDimMap[analysisDim.Name] = true
		}

		// 校验复合指标的分析维度是否属于依赖指标
		for _, analysisDim := range metricModel.AnalysisDims {
			if !dependDimMap[analysisDim.Name] {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_AnalysisDimensionNotExisted).
					WithErrorDetails(fmt.Sprintf("The analysis dimension with name [%s] not found in metric model[%s]!", analysisDim.Name, dependModelMap[modelID].ModelName))
			}
		}

		for _, order := range metricModel.OrderByFields {
			if !orderByFieldMap[order.Name] {
				// 不属于分组字段集、__value，则报错
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_OrderByFieldNotExisted).
					WithErrorDetails(fmt.Sprintf("The derived metric model's order_by_field[%s] is not belong to the group fields of atomic metric model[%s]!",
						order.Name, dependModelMap[modelID].ModelID))
			}
		}
	}
	// 复合指标的计算公式有效性校验
	// 计算公式有效性检查，请求 uniquery，与指标数据预览同一个接口
	err = mms.CheckFormula(ctx, *metricModel)
	if err != nil {
		return err
	}
	return nil
}

// 获取指标模型的排序字段集
func (mms *metricModelService) getOrderByFields(ctx context.Context, model interfaces.MetricModel) (map[string]interfaces.MetricModel, map[string]bool, error) {
	switch model.MetricType {
	case interfaces.ATOMIC_METRIC:
		return mms.getAtomicSortFields(model)
	case interfaces.DERIVED_METRIC:
		return mms.getDerivedSortFields(ctx, model)
	case interfaces.COMPOSITED_METRIC:
		return mms.getCompositeSortFields(ctx, model)
	default:
		return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_MetricModel_UnsupportMetricType)
	}
}

// 原子指标：排序字段集 = 分组字段集 + 值字段
func (mms *metricModelService) getAtomicSortFields(model interfaces.MetricModel) (map[string]interfaces.MetricModel, map[string]bool, error) {
	orderByFieldMap := map[string]bool{}
	for _, group := range model.FormulaConfig.(interfaces.SQLConfig).GroupByFields {
		orderByFieldMap[group] = true
	}
	// 去掉本指标模型身上配置的排序字段
	for _, order := range model.OrderByFields {
		delete(orderByFieldMap, order.Name)
	}
	// 增加值字段
	orderByFieldMap[interfaces.VALUE_FIELD_NAME] = true
	return map[string]interfaces.MetricModel{
		model.ModelID: model,
	}, orderByFieldMap, nil
}

// 衍生指标：排序字段集 = 依赖原子指标的分组字段 - 依赖原子指标的排序字段 + 值字段
func (mms *metricModelService) getDerivedSortFields(ctx context.Context, model interfaces.MetricModel) (map[string]interfaces.MetricModel, map[string]bool, error) {
	derivedModel := model.FormulaConfig.(interfaces.DerivedConfig)
	// 先获取依赖的原子指标模型
	dependModel, exist, err := mms.mma.GetMetricModelByModelID(ctx, derivedModel.DependMetricModel.ID)
	if err != nil {
		return nil, nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_GetModelByIDFailed).WithErrorDetails(err.Error())

	}
	if !exist {
		return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_MetricModelNotFound).
			WithErrorDetails(fmt.Sprintf("The Depend metric model of derived metric model with id [%s] not exists!", derivedModel.DependMetricModel))
	}
	// 分组字段要一直递归
	// 衍生指标的排序字段：列表集为原子指标的分组字段 - 原子指标的排序字段 + 值字段
	orderByFieldMap := map[string]bool{}
	for _, group := range dependModel.FormulaConfig.(interfaces.SQLConfig).GroupByFields {
		orderByFieldMap[group] = true
	}
	// 去掉原子指标配置的排序字段
	for _, order := range dependModel.OrderByFields {
		delete(orderByFieldMap, order.Name)
	}

	// 去掉本指标模型身上配置的排序字段
	for _, order := range model.OrderByFields {
		delete(orderByFieldMap, order.Name)
	}
	// 增加值字段
	orderByFieldMap[interfaces.VALUE_FIELD_NAME] = true

	// 返回衍生指标依赖的原子指标
	return map[string]interfaces.MetricModel{
		dependModel.ModelID: dependModel,
	}, orderByFieldMap, nil
}

// 复合指标：排序字段集 = 各参与指标排序字段集的交集 + 值字段
func (mms *metricModelService) getCompositeSortFields(ctx context.Context, model interfaces.MetricModel) (map[string]interfaces.MetricModel, map[string]bool, error) {
	modelIDs := common.ExtractModelIDs(model.Formula)
	// 获取所有依赖指标的排序字段集
	orderByFieldMap := map[string]bool{}
	// 获取所有依赖指标的排序字段集
	allDependModel := map[string]interfaces.MetricModel{}
	allSortFields := make([]map[string]bool, 0, len(modelIDs))
	for _, dep := range modelIDs {
		dependModel, exist, err := mms.mma.GetMetricModelByModelID(ctx, dep)
		if err != nil {
			return allDependModel, nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_MetricModel_InternalError_GetModelByIDFailed).WithErrorDetails(err.Error())

		}
		if !exist {
			return allDependModel, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_MetricModelNotFound).
				WithErrorDetails(fmt.Sprintf("The Depend metric model[%s] of composite metric model with id [%s] not exists!", dep, model.ModelID))
		}
		_, sortFields, err := mms.getOrderByFields(ctx, dependModel)
		if err != nil {
			return allDependModel, nil, err
		}
		allDependModel[dep] = dependModel
		allSortFields = append(allSortFields, sortFields)
	}
	// 计算所有排序字段集的交集
	// 使用第一个集合作为基准
	for field := range allSortFields[0] {
		orderByFieldMap[field] = true
	}
	// 与其他集合求交集
	for i := 1; i < len(allSortFields); i++ {
		for field := range orderByFieldMap {
			if !allSortFields[i][field] {
				delete(orderByFieldMap, field)
			}
		}
	}

	// 去掉本指标模型身上配置的排序字段
	for _, order := range model.OrderByFields {
		delete(orderByFieldMap, order.Name)
	}
	// 增加值字段
	orderByFieldMap[interfaces.VALUE_FIELD_NAME] = true

	return allDependModel, orderByFieldMap, nil
}

// 获取指标模型的映射视图的原始字段的字段集
func (mms *metricModelService) getMetricViewFieldsAndGroupFields(ctx context.Context,
	model interfaces.MetricModel) (map[string]*interfaces.ViewField, map[string]interfaces.Field, error) {

	switch model.MetricType {
	case interfaces.ATOMIC_METRIC:
		return mms.getAtomicViewFieldsAndGroupFields(ctx, model)
	case interfaces.DERIVED_METRIC:
		return mms.getDerivedViewFieldsAndGroupFields(ctx, model)
	case interfaces.COMPOSITED_METRIC:
		return mms.getCompositeViewFieldsAndGroupFields(ctx, model)
	default:
		return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_MetricModel_UnsupportMetricType)
	}
}

// 原子指标：视图的字段集
func (mms *metricModelService) getAtomicViewFieldsAndGroupFields(ctx context.Context,
	model interfaces.MetricModel) (map[string]*interfaces.ViewField, map[string]interfaces.Field, error) {
	// 获取视图信息
	dataView, err := mms.dvs.GetDataView(ctx, model.DataSource.ID)
	if err != nil {
		return nil, nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_GetDataViewQueryFiltersFailed).WithErrorDetails(err.Error())
	}

	// sql-原子：分组字段
	groupFieldsMap := map[string]interfaces.Field{}
	if model.DataSource.Type == interfaces.QueryType_SQL {
		// 分组字段
		for _, groupBy := range model.FormulaConfig.(interfaces.SQLConfig).GroupByFields {
			groupFieldsMap[dataView.FieldsMap[groupBy].Name] = interfaces.Field{
				Name:        dataView.FieldsMap[groupBy].Name,
				DisplayName: dataView.FieldsMap[groupBy].DisplayName,
				Type:        dataView.FieldsMap[groupBy].Type,
				Comment:     &dataView.FieldsMap[groupBy].Comment,
			}
		}

		// 聚合字段, 配置化时需要返回聚合字段，用于详情展示
		aggr := model.FormulaConfig.(interfaces.SQLConfig).AggrExpr
		if aggr != nil {
			groupFieldsMap[aggr.Field] = interfaces.Field{
				Name:        dataView.FieldsMap[aggr.Field].Name,
				DisplayName: dataView.FieldsMap[aggr.Field].DisplayName,
				Type:        dataView.FieldsMap[aggr.Field].Type,
				Comment:     &dataView.FieldsMap[aggr.Field].Comment,
			}
		}

		// 日期标识字段
		if model.DateField != "" {
			groupFieldsMap[dataView.FieldsMap[model.DateField].Name] = interfaces.Field{
				Name:        dataView.FieldsMap[model.DateField].Name,
				DisplayName: dataView.FieldsMap[model.DateField].DisplayName,
				Type:        dataView.FieldsMap[model.DateField].Type,
				Comment:     &dataView.FieldsMap[model.DateField].Comment,
			}
		}
	}

	// 原子指标的原始字段集为其绑定的数据视图的字段集
	return dataView.FieldsMap, groupFieldsMap, err
}

// 衍生指标：其原子对应的视图的字段集
func (mms *metricModelService) getDerivedViewFieldsAndGroupFields(ctx context.Context,
	model interfaces.MetricModel) (map[string]*interfaces.ViewField, map[string]interfaces.Field, error) {
	derivedModel := model.FormulaConfig.(interfaces.DerivedConfig)
	// 先获取依赖的原子指标模型
	dependModel, exist, err := mms.mma.GetMetricModelByModelID(ctx, derivedModel.DependMetricModel.ID)
	if err != nil {
		return nil, nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_GetModelByIDFailed).WithErrorDetails(err.Error())

	}
	if !exist {
		return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_MetricModelNotFound).
			WithErrorDetails(fmt.Sprintf("The Depend metric model of derived metric model with id [%s] not exists!", derivedModel.DependMetricModel))
	}
	// 衍生指标的原始字段集是其依赖的原子指标的字段集
	return mms.getMetricViewFieldsAndGroupFields(ctx, dependModel)
}

// 复合指标：其依赖的指标模型的视图字段集的
func (mms *metricModelService) getCompositeViewFieldsAndGroupFields(ctx context.Context,
	model interfaces.MetricModel) (map[string]*interfaces.ViewField, map[string]interfaces.Field, error) {
	modelIDs := common.ExtractModelIDs(model.Formula)

	// 获取所有依赖指标的原始字段集
	allFields := map[string]*interfaces.ViewField{}
	allGroupFields := map[string]interfaces.Field{}
	for _, dep := range modelIDs {
		dependModel, exist, err := mms.mma.GetMetricModelByModelID(ctx, dep)
		if err != nil {
			return nil, nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_MetricModel_InternalError_GetModelByIDFailed).WithErrorDetails(err.Error())

		}
		if !exist {
			return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_MetricModelNotFound).
				WithErrorDetails(fmt.Sprintf("The Depend metric model[%s] of composite metric model with id [%s] not exists!", dep, model.ModelID))
		}
		fields, groupFields, err := mms.getMetricViewFieldsAndGroupFields(ctx, dependModel)
		if err != nil {
			return nil, nil, err
		}

		for k, v := range fields {
			allFields[k] = v
		}

		for k, v := range groupFields {
			allGroupFields[k] = v
		}
	}
	return allFields, allGroupFields, nil
}

func (mms *metricModelService) GetMetricModelOrderByFields(ctx context.Context, modelIDs []string) ([][]*interfaces.ViewField, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("查询指标模型[%s]的排序字段列表", modelIDs))
	span.SetAttributes(attr.Key("model_id").StringSlice(modelIDs))
	defer span.End()

	allFields := [][]*interfaces.ViewField{}

	// id去重后再查
	mIDs := common.DuplicateSlice(modelIDs)

	// 获取模型基本信息
	metricModels, err := mms.mma.GetMetricModelsByModelIDs(ctx, mIDs)
	if err != nil {
		logger.Errorf("GetMetricModelByModelID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get metric model[%v] error: %v", mIDs, err))
		span.End()

		return allFields, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_GetModelByIDFailed).WithErrorDetails(err.Error())
	}
	if len(metricModels) != len(mIDs) {
		errStr := fmt.Sprintf("Exists any models not found, expect model nums is [%d], actual models num is [%d]", len(mIDs), len(metricModels))
		logger.Errorf(errStr)
		span.SetStatus(codes.Error, errStr)
		span.End()

		return allFields, rest.NewHTTPError(ctx, http.StatusNotFound,
			derrors.DataModel_MetricModel_MetricModelNotFound).WithErrorDetails(errStr)
	}

	// 先获取资源序列  todo:
	matchResouces, err := mms.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_METRIC_MODEL, mIDs,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		return allFields, err
	}
	// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
	for _, model := range metricModels {
		if _, exist := matchResouces[model.ModelID]; !exist {
			return allFields, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for metric model's view_detail operation.")
		}
	}

	for _, model := range metricModels {
		fields := []*interfaces.ViewField{}
		_, fieldMap, err := mms.getOrderByFields(ctx, model)
		if err != nil {
			span.SetStatus(codes.Error, "GetOrderByFields error")
			return allFields, err
		}

		// 获取视图字段
		viewFieldMap, _, err := mms.getMetricViewFieldsAndGroupFields(ctx, model)
		if err != nil {
			return allFields, err
		}

		for fieldName := range fieldMap {
			if fieldName == interfaces.VALUE_FIELD_NAME {
				fields = append(fields, &interfaces.ViewField{
					Name:        fieldName,
					DisplayName: interfaces.VALUE_FIELD_DISPLAY_NAME,
				})
			} else {
				fields = append(fields, viewFieldMap[fieldName])
			}
		}
		allFields = append(allFields, fields)
	}

	span.SetStatus(codes.Ok, "")
	return allFields, nil
}

// 翻译分析维度和排序字段
func transferDimAndOrder(ctx context.Context,
	model *interfaces.MetricModel,
	viewFields map[string]*interfaces.ViewField,
	fieldsMap map[string]interfaces.Field) {

	// 分析维度
	for i, dim := range model.AnalysisDims {
		_, exists := viewFields[dim.Name]
		if exists {
			model.AnalysisDims[i].Name = viewFields[dim.Name].Name
			model.AnalysisDims[i].DisplayName = viewFields[dim.Name].DisplayName
			model.AnalysisDims[i].Type = viewFields[dim.Name].Type
			model.AnalysisDims[i].Comment = &viewFields[dim.Name].Comment

			fieldsMap[dim.Name] = interfaces.Field{
				Name:        viewFields[dim.Name].Name,
				DisplayName: viewFields[dim.Name].DisplayName,
				Type:        viewFields[dim.Name].Type,
				Comment:     &viewFields[dim.Name].Comment,
			}
		} else {
			o11y.Warn(ctx, fmt.Sprintf("Analysis dimension[%s] not found in depend metric model not found", dim.Name))
			fieldsMap[dim.Name] = dim
		}
	}
	// 排序字段
	for i, orderBy := range model.OrderByFields {
		if orderBy.Name == interfaces.VALUE_FIELD_NAME {
			model.OrderByFields[i].DisplayName = interfaces.VALUE_FIELD_DISPLAY_NAME

			comment := interfaces.VALUE_FIELD_DISPLAY_NAME
			fieldsMap[orderBy.Name] = interfaces.Field{
				Name:        orderBy.Name,
				DisplayName: interfaces.VALUE_FIELD_DISPLAY_NAME,
				Type:        "float",
				Comment:     &comment,
			}
			continue
		}
		vField, exists := viewFields[orderBy.Name]
		if exists {
			model.OrderByFields[i].DisplayName = vField.DisplayName
			model.OrderByFields[i].Type = vField.Type

			fieldsMap[orderBy.Name] = interfaces.Field{
				Name:        vField.Name,
				DisplayName: vField.DisplayName,
				Type:        vField.Type,
				Comment:     &vField.Comment,
			}
		} else {
			o11y.Warn(ctx, fmt.Sprintf("Order by field[%s] not found in depend metric model not found", orderBy.Name))
			fieldsMap[orderBy.Name] = interfaces.Field{
				Name:        orderBy.Name,
				DisplayName: orderBy.DisplayName,
				Type:        orderBy.Type,
			}
		}
	}
	// 设置指标模型的fields_map
	model.FieldsMap = fieldsMap
}
