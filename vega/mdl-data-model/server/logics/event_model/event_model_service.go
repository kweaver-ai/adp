// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package event_model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	"data-model/logics"
	"data-model/logics/data_view"
	"data-model/logics/metric_model"
	"data-model/logics/permission"
)

var (
	emServiceOnce sync.Once
	emService     interfaces.EventModelService
)

const (
	DEFAULT_DISPACH_CONFIG_TIME_OUT = 3600
	DEFAULT_FAIL_RETRY_COUNT        = 1
)

type eventModelService struct {
	appSetting *common.AppSetting
	dmja       interfaces.DataModelJobAccess
	ema        interfaces.EventModelAccess
	iba        interfaces.IndexBaseAccess
	mms        interfaces.MetricModelService
	dvs        interfaces.DataViewService
	db         *sql.DB
	ps         interfaces.PermissionService
}

func NewEventModelService(appSetting *common.AppSetting) interfaces.EventModelService {
	emServiceOnce.Do(func() {
		emService = &eventModelService{
			appSetting: appSetting,
			db:         logics.DB,
			dmja:       logics.DMJA,
			dvs:        data_view.NewDataViewService(appSetting),
			ema:        logics.EMA,
			iba:        logics.IBA,
			mms:        metric_model.NewMetricModelService(appSetting),
			ps:         permission.NewPermissionService(appSetting),
		}
	})
	return emService
}

// 根据名称获取事件模型存在性
func (ems *eventModelService) CheckEventModelExistByName(ctx context.Context, modelName string) (bool, error) {

	//NOTE: 构建查询参数，且设置默认值
	emq := interfaces.EventModelQueryRequest{
		EventModelName: modelName,
		Direction:      "desc",
		SortKey:        "update_time",
		Limit:          10,
		Offset:         0,
		IsActive:       "",
		IsCustom:       -1,
	}
	models, err := ems.ema.QueryEventModels(ctx, emq)

	if err != nil {
		logger.Errorf("Query Event Model by ID error: %s", err.Error())
		return false, err
	}
	//NOTE 如果查询找到对应事件模型，则返回已存在状态
	if len(models) > 0 {
		return true, nil
	}
	//NOTE 返回部存在同名事件模型
	return false, nil

}

func (ems *eventModelService) BatchValidateDataSources(ctx context.Context, dataSource []string, event_model_type string, dataSourceType string, DataSourceName []string, DataSourceGroupName []string) ([]string, *rest.HTTPError) {
	//FIXME: 多种数据来源时，需要以反射或者抽象工程来重构

	if event_model_type == "atomic" && len(dataSource) != 1 {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).WithErrorDetails("DataSource count is illegal")
	}
	var newDataSources = []string{}
	for index, sourceID := range dataSource {
		var name, group_name string
		if len(DataSourceName) == 0 {
			name = ""
		} else {
			name = DataSourceName[index]

		}
		if len(DataSourceGroupName) == 0 {
			group_name = ""

		} else {
			group_name = DataSourceGroupName[index]
		}
		newDataSourceID, err := ems.ValidateEventModelDataSource(ctx, sourceID, dataSourceType, name, group_name)
		if err != nil || newDataSourceID == "" || newDataSourceID == "0" {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_DataSourceIllegal).WithErrorDetails(fmt.Sprintf("DataSource: %s is illegal", name))
		}
		newDataSources = append(newDataSources, newDataSourceID)
	}
	return newDataSources, nil
}

// 验证事件模型的数据源对象的合法性，目前主要是判断存不存在。
func (ems *eventModelService) ValidateEventModelDataSource(ctx context.Context, dataSource string, dataSourceType string, DataSourceName string, DataSourceGroupName string) (string, error) {
	//FIXME: 多种数据来源时，需要以反射或者抽象工程来重构

	if dataSourceType == interfaces.EVENT_MODEL_DATA_SOURCE_TYPE_FOR_METRIC_MODEL {
		var NewDataSource string
		mm, err := ems.mms.GetMetricModelByModelID(ctx, dataSource)
		if err != nil {
			NewDataSource, err = ems.mms.GetMetricModelIDByName(ctx, DataSourceGroupName, DataSourceName)
			if err != nil || NewDataSource == "" {
				return "", err
			} else {
				return NewDataSource, nil
			}

		}

		return mm.ModelID, nil
	} else if dataSourceType == interfaces.EVENT_MODEL_DATA_SOURCE_TYPE_FOR_EVENT_MODEL {
		// var NewDataSourceMap = make(map[string]string)
		// var err1 error
		id := dataSource
		em, httpErr := ems.GetEventModelByID(ctx, id)
		if httpErr != nil {
			NewDataSourceMap, err1 := ems.ema.GetEventModelMapByNames([]string{DataSourceName})
			if err1 != nil || len(NewDataSourceMap) == 0 {
				return "", err1
			} else {
				return NewDataSourceMap[DataSourceName], nil
			}
		}

		return em.EventModelID, nil
		// return NewDataSourceMap[DataSourceName], nil
	} else if dataSourceType == interfaces.EVENT_MODEL_DATA_SOURCE_TYPE_FOR_DATE_VIEW {
		_, exist, err := ems.dvs.CheckDataViewExistByID(ctx, nil, dataSource)
		if err != nil {
			return "", err
		}

		if !exist {
			return "", fmt.Errorf("data View with ID: %v not found! ", dataSource)
		}

		return dataSource, nil

	} else {
		return "", nil
	}

}

// FIXME: 需要改成递归检验。
func (ems *eventModelService) ValidateEventModelDetectRule(ctx context.Context, detectRule interfaces.DetectRule, detectRuleType string) *rest.HTTPError {
	//FIXME: 多种检测类型时，需要以反射或者抽象工程来重构
	//TODO:  添加事件级别唯一性校验

	//TODO 将来从SDK的过滤器模块中提取所有过滤器操作符
	RangeOpearationMap := map[string]bool{"range": true, "out_range": true, "in": true, "contain": true}
	SingleOpearationMap := map[string]bool{"=": true, "==": true, "<": true, ">": true, "<=": true, ">=": true, "!=": true, "like": true, "notlike": true, "not_like": true, "NotLike": true}
	for _, formulaItem := range detectRule.Formula {
		//INFO: 多值类型校验
		if RangeOpearationMap[formulaItem.Filter.FilterExpress.Operation] {
			t := reflect.TypeOf(formulaItem.Filter.FilterExpress.Value)

			if t.Kind() != reflect.Array && t.Kind() != reflect.Slice {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).WithErrorDetails("operation and value is mismatched")
			}

		}
		//INFO: 单值类型校验
		if SingleOpearationMap[formulaItem.Filter.FilterExpress.Operation] {
			t := reflect.TypeOf(formulaItem.Filter.FilterExpress.Value)
			if t.Kind() == reflect.String &&
				formulaItem.Filter.FilterExpress.Operation != "=" &&
				formulaItem.Filter.FilterExpress.Operation != "==" &&
				formulaItem.Filter.FilterExpress.Operation != "like" &&
				formulaItem.Filter.FilterExpress.Operation != "notlike" &&
				formulaItem.Filter.FilterExpress.Operation != "not_like" &&
				formulaItem.Filter.FilterExpress.Operation != "NotLike" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).WithErrorDetails("operation and value is mismatched")
			}
			if t.Kind() != reflect.String && t.Kind() != reflect.Float64 {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).WithErrorDetails("operation and value is mismatched")
			}
		}
	}

	return nil
}

func (ems *eventModelService) ValidateEventModelAggregateRule(ctx context.Context, AggregateRule interfaces.AggregateRule, aggregateRuleType string) bool {
	//NOTE: 此处需要提取出去重构，为以后扩展做准备
	if common.In(AggregateRule.Type, []string{"healthy_compute", "group_aggregation"}) && common.In(AggregateRule.AggregateAlgo, []string{"MaxLevelMap", "SourceDataGroupAggregation", "EventDataGroupAggregation"}) {
		return true
	} else {
		return false
	}

}

// 事件模型创建请求参数校验
func (ems *eventModelService) EventModelCreateValidate(ctx context.Context, eventModel interfaces.EventModel) (interfaces.EventModel, *rest.HTTPError) {

	//NOTE 时间窗口合法校验
	var legalDuration bool
	switch eventModel.DefaultTimeWindow.Unit {
	case "d":
		legalDuration = eventModel.DefaultTimeWindow.Interval > 1 || eventModel.DefaultTimeWindow.Interval <= 0

	case "h":
		legalDuration = eventModel.DefaultTimeWindow.Interval > 24 || eventModel.DefaultTimeWindow.Interval <= 0
	case "m":
		legalDuration = eventModel.DefaultTimeWindow.Interval > 1440 || eventModel.DefaultTimeWindow.Interval <= 0
	default:
		legalDuration = false
	}
	if legalDuration {
		return interfaces.EventModel{}, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).WithErrorDetails(" event model time duration is illeagal duplicated")
	}

	//NOTE  名称存在性校验
	isExist, err := ems.CheckEventModelExistByName(ctx, eventModel.EventModelName)

	if err != nil {
		logger.Errorf("event_model exist judge with model name cause an internal error: %s,%v", eventModel.EventModelName, err.Error())
		return interfaces.EventModel{}, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).WithErrorDetails(" internal error when checking event model name")
	}
	if isExist {
		logger.Errorf("event_model exists with model name: %s", eventModel.EventModelName)
		return interfaces.EventModel{}, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_ModelNameExisted).WithErrorDetails(" event model name duplicated")
	}

	//NOTE : 数据源存在校验
	var NewDataSources []string
	var httpErr *rest.HTTPError
	if len(eventModel.DataSource) > 0 {
		NewDataSources, httpErr = ems.BatchValidateDataSources(ctx, eventModel.DataSource, eventModel.EventModelType, eventModel.DataSourceType, eventModel.DataSourceName, eventModel.DataSourceGroupName)
		if httpErr != nil {
			logger.Errorf("event_model data source validate occur an internal error: %s,%v", eventModel.EventModelName, httpErr.BaseError.Description)
			return interfaces.EventModel{}, httpErr
		}
	}
	// if eventModel.AggregateRule.Type != "agi_aggregation" {
	// 	NewDataSources, httpErr = ems.BatchValidateDataSources(ctx, eventModel.DataSource, eventModel.EventModelType, eventModel.DataSourceType, eventModel.DataSourceName, eventModel.DataSourceGroupName)
	// 	if httpErr != nil {
	// 		logger.Errorf("event_model data source validate occur an internal error: %s,%v", eventModel.EventModelName, httpErr.BaseError.Description)
	// 		return interfaces.EventModel{}, httpErr
	// 	}
	// }
	//NOTE：增加依赖任务检测
	//TODO: 增加关联依赖依赖，不允许递归依赖。
	if len(eventModel.DownstreamDependentModel) > 3 {
		logger.Errorf("event_model has downstream depent task more than 3!")
		return interfaces.EventModel{}, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_DependentModelIllegal).WithErrorDetails("event_model has downstream depent task more than 3,  please check downstream dependent task")
	}
	//NOTE: 任务去重依赖
	var MapModel = make(map[string]interfaces.EventModel, len(eventModel.DownstreamDependentModel))
	for _, model_id := range eventModel.DownstreamDependentModel {
		model, err := ems.GetEventModelByID(ctx, model_id)
		if err != nil {
			logger.Errorf("The dependent model of event model not exist yet!")
			return interfaces.EventModel{}, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_DependentModelIllegal).WithErrorDetails(" The dependent model of event model not exist yet!")

		}
		// if model.DownstreamDependentModel
		if len(model.DownstreamDependentModel) > 0 {
			logger.Errorf("The dependent model of event model cannot have downstream model dependencies,Recursive dependencies are not supported yet!")
			return interfaces.EventModel{}, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_DependentModelIllegal).WithErrorDetails(" The dependent model of event model cannot have downstream model dependencies,Recursive dependencies are not supported yet!")
		}
		if v, ok := MapModel[model_id]; ok {
			logger.Errorf("event_model duplicated!")
			return interfaces.EventModel{}, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_DependentModelIllegal).WithErrorDetails(" event_model duplicated, please check downstream dependent task")
		} else {
			MapModel[model_id] = v
		}
	}

	//NOTE 任务不能依赖自己。
	SelfID := eventModel.EventModelID
	if _, ok := MapModel[SelfID]; ok {
		logger.Errorf("event_model can't depent itemself!")
		return interfaces.EventModel{}, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_DependentModelIllegal).WithErrorDetails(" event_model can't depent itemself, please check downstream dependent task")
	}

	//NOTE 当事件模型未开启时， 定期执行和实时订阅不同同时为选中状态
	//NOTE 为了兼容以前的事件模型导入(is_active=0 且无enable_subscribe属性和status属性，会默认置0),所以这里当三者都为0时，允许通过校验，并置is_active=1,status=0。
	if eventModel.Status == 0 && eventModel.IsActive == 1 && eventModel.EnableSubscribe == 1 {
		logger.Errorf("run model of event_model both active or inactive(batch and subscribe): %s", eventModel.EventModelName)
		return interfaces.EventModel{}, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_RunModeIllegal).WithErrorDetails(" please check is_active and enable_subscribe,must not equal to 1 both")
	}
	//NOTE 导入旧事件模型时，如果持久化为开启状态(is_active=1)，则置启用字段为1，执行模式为定期执行方式(status=1,is_active=1)。
	// if eventModel.Status == 0 && eventModel.IsActive == 1 {
	// 	eventModel.Status = 1
	// }
	//NOTE 导入旧事件模型时，默认置为定期执行方式
	if eventModel.Status == 0 && eventModel.IsActive == 0 && eventModel.EnableSubscribe == 0 {
		eventModel.IsActive = 1
	}

	//NOTE 当status为1时，代表一定是当前匹配版本，定期执行和实时订阅不同同时为选中状态或非选中状态，必须选其一。
	if eventModel.Status == 1 && ((eventModel.IsActive == 1 && eventModel.EnableSubscribe == 1) || (eventModel.IsActive == 0 && eventModel.EnableSubscribe == 0)) {
		logger.Errorf("run model of event_model both active or inactive(batch and subscribe): %s", eventModel.EventModelName)
		return interfaces.EventModel{}, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_RunModeIllegal).WithErrorDetails(" please check is_active and enable_subscribe,must not equal to 1 or 0 both")
	}

	//NOTE 当事件模型未开启时， 定期执行和实时订阅不同同时为选中状态
	//NOTE 为了兼容以前的事件模型导入(is_active=0 且无enable_subscribe属性和status属性，会默认置0),所以这里当三者都为0时，允许通过校验，并置is_active=1,status=0。
	if eventModel.Status == 0 && eventModel.IsActive == 1 && eventModel.EnableSubscribe == 1 {
		logger.Errorf("run model of event_model both active or inactive(batch and subscribe): %s", eventModel.EventModelName)
		return interfaces.EventModel{}, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_RunModeIllegal).WithErrorDetails(" please check is_active and enable_subscribe,must not equal to 1 both")
	}
	//NOTE 导入旧事件模型时，如果持久化为开启状态(is_active=1)，则置启用字段为1，执行模式为定期执行方式(status=1,is_active=1)。
	// if eventModel.Status == 0 && eventModel.IsActive == 1 {
	// 	eventModel.Status = 1
	// }
	//NOTE 导入旧事件模型时，默认置为定期执行方式
	if eventModel.Status == 0 && eventModel.IsActive == 0 && eventModel.EnableSubscribe == 0 {
		eventModel.IsActive = 1
	}

	//NOTE 当status为1时，代表一定是当前匹配版本，定期执行和实时订阅不同同时为选中状态或非选中状态，必须选其一。
	if eventModel.Status == 1 && ((eventModel.IsActive == 1 && eventModel.EnableSubscribe == 1) || (eventModel.IsActive == 0 && eventModel.EnableSubscribe == 0)) {
		logger.Errorf("run model of event_model both active or inactive(batch and subscribe): %s", eventModel.EventModelName)
		return interfaces.EventModel{}, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_RunModeIllegal).WithErrorDetails(" please check is_active and enable_subscribe,must not equal to 1 or 0 both")
	}

	//NOTE ：触发条件校验
	if eventModel.EventModelType == "atomic" {
		if eventModel.DetectRule.Type != "agi_detect" {
			httpErr = ems.ValidateEventModelDetectRule(ctx, eventModel.DetectRule, eventModel.DetectRule.Type)
			if httpErr != nil {
				logger.Errorf("event_model detect rule validate occur an internal error: %s,%v", eventModel.EventModelName, httpErr.BaseError.Description)
				return interfaces.EventModel{}, httpErr
			}
		}
	} else {
		if eventModel.AggregateRule.Type != "agi_aggregation" {
			ilLegal := ems.ValidateEventModelAggregateRule(ctx, eventModel.AggregateRule, eventModel.AggregateRule.Type)
			if !ilLegal {
				logger.Errorf("event_model aggregate rule validate occur an internal error: %s,%v", eventModel.EventModelName, "aggregate rule is illegal")
				return interfaces.EventModel{}, httpErr
			}
		}
	}

	if len(NewDataSources) > 0 {
		eventModel.DataSource = NewDataSources
	}

	//NOTE 配置持久化任务
	if !common.IsTaskEmpty(eventModel.Task) {
		//NOTE 索引库及存在性校验
		baseinfos, err := ems.iba.GetSimpleIndexBasesByTypes(ctx, []string{eventModel.Task.StorageConfig.IndexBase})
		if err != nil {
			logger.Errorf("GetSimpleIndexBasesByTypes error: %v \n", err)
			return eventModel, rest.NewHTTPError(ctx, http.StatusBadRequest,
				derrors.EventModel_TaskSyncCreateFailed).WithErrorDetails(err.Error())
		}
		if len(baseinfos) == 0 {
			logger.Errorf("the index_base does not exsit\n")
			return eventModel, rest.NewHTTPError(ctx, http.StatusBadRequest,
				derrors.EventModel_TaskSyncCreateFailed).WithErrorDetails("the index_base does not exsit\n")
		}
		//NOTE: 数据视图存在性校验
		_, exist, err := ems.dvs.CheckDataViewExistByID(ctx, nil, eventModel.Task.StorageConfig.DataViewID)
		if err != nil {
			logger.Errorf("CheckDataViewExistByID error: %v \n", err)
			return eventModel, rest.NewHTTPError(ctx, http.StatusBadRequest,
				derrors.EventModel_TaskSyncCreateFailed).WithErrorDetails(err.Error())
		} else if !exist {
			//说明不存在这个数据视图
			logger.Errorf("data view not ound  \n")
			return eventModel, rest.NewHTTPError(ctx, http.StatusBadRequest,
				derrors.EventModel_TaskSyncCreateFailed).WithErrorDetails("Data view is not exist")
		}

	}

	return eventModel, nil
}

// 事件模型更新请求参数校验
func (ems *eventModelService) EventModelUpdateValidate(ctx context.Context, eventModel interfaces.EventModelUpateRequest) *rest.HTTPError {

	em, httpErr := ems.GetEventModelByID(ctx, eventModel.EventModelID)
	//NOTE： 如果返回不存在错误或内部错误，则直接返回
	if httpErr != nil {
		return httpErr
	}
	//NOTE 如果配置有持久化任务，则状态为完成时才允许更新
	// if em.Task.StorageConfig.IndexBase != "" && em.Task.ScheduleSyncStatus != interfaces.SCHEDULE_SYNC_STATUS_FINISH {
	// 	logger.Errorf("event_model_task schedule sync status is not finish when updating event_model")
	// 	return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_EventModel_InternalError_TaskBeingModified).WithErrorDetails(" event_model_task schedule sync status is not finish when updating event_model")
	// }

	//NOTE  名称存在性校验
	if em.EventModelName != eventModel.EventModelName {
		isExist, err := ems.CheckEventModelExistByName(ctx, eventModel.EventModelName)

		if err != nil {
			logger.Errorf("event_model exist judge with model name cause an internal error: %s,%v", eventModel.EventModelName, err.Error())
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).WithErrorDetails(" internal error when checking event model name")
		}
		if isExist {
			logger.Errorf("event_model exists with model name: %s", eventModel.EventModelName)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_ModelNameExisted).WithErrorDetails(" event model name duplicated")
		}
	}

	if eventModel.IsActive == 1 && eventModel.EnableSubscribe == 1 {
		logger.Errorf("run model of event_model both active(batch and subscribe): %s", eventModel.EventModelName)
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).WithErrorDetails("un model of event_model both active(batch and subscribe)")
	}
	Interval := eventModel.DefaultTimeWindow.Interval
	Unit := eventModel.DefaultTimeWindow.Unit

	//NOTE 时间窗口合法校验
	var legalDuration bool

	switch Unit {
	case "d":
		legalDuration = Interval > 1 || Interval <= 0

	case "h":
		legalDuration = Interval > 24 || Interval <= 0
	case "m":
		legalDuration = Interval > 1440 || Interval <= 0
	default:
		legalDuration = false
	}
	if legalDuration {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_InvalidParameter).WithErrorDetails(" event model time duration is illeagal duplicated")
	}

	sourceIDs := eventModel.DataSource
	ModelNames := make([]string, 0, len(sourceIDs))
	GroupNames := make([]string, 0, len(sourceIDs))
	if eventModel.DataSourceType == "metric_model" {
		IDNameMap, _ := ems.mms.GetMetricModelSimpleInfosByIDs(ctx, eventModel.DataSource)
		for _, value := range IDNameMap {
			ModelNames = append(ModelNames, value.ModelName)
			GroupNames = append(GroupNames, value.GroupName)

		}
	} else if eventModel.DataSourceType == "event_model" {
		ModelNames, _ = ems.ema.GetEventModelNamesByIDs(sourceIDs)
	}

	eventModel.DataSourceName = ModelNames
	eventModel.DataSourceGroupName = GroupNames

	//NOTE : 数据源存在校验
	if eventModel.AggregateRule.Type != "agi_aggregation" {
		_, httpErr = ems.BatchValidateDataSources(ctx, eventModel.DataSource, eventModel.EventModelType, eventModel.DataSourceType, eventModel.DataSourceName, eventModel.DataSourceGroupName)
		if httpErr != nil {
			logger.Errorf("event_model data source validate occur an internal error: %s", httpErr.BaseError.Description)
			return httpErr
		}
	}
	//NOTE  任务依赖检测
	if len(eventModel.DownstreamDependentModel) > 0 {
		//NOTE 根据事件模型ID去获取任务ID，如果两者不匹配，则非法。
		emIDs := eventModel.DownstreamDependentModel
		models, err := ems.GetEventModelMapByIDs(emIDs)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_DependentModelIllegal).WithErrorDetails(" event model dependent model is illeagal,please check")
		}
		if len(models) != len(eventModel.DownstreamDependentModel) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_DependentModelIllegal).WithErrorDetails(" event model dependent model is illeagal,please check")
		}
		//NOTE 重复依赖检测
		var MapModel = make(map[string]interfaces.EventModel, len(eventModel.DownstreamDependentModel))

		for _, model_id := range eventModel.DownstreamDependentModel {
			model, _ := ems.GetEventModelByID(ctx, model_id)
			// if model.DownstreamDependentModel
			if len(model.DownstreamDependentModel) > 0 {
				logger.Errorf("The dependent model of event model cannot have downstream model dependencies,Recursive dependencies are not supported yet!")
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_DependentModelIllegal).WithErrorDetails(" The dependent model of event model cannot have downstream model dependencies,Recursive dependencies are not supported yet!")
			}
			if v, ok := MapModel[model_id]; ok {
				logger.Errorf("event_model duplicated!")
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_DependentModelIllegal).WithErrorDetails(" event_model duplicated, please check downstream dependent task")
			} else {
				MapModel[model_id] = v
			}
		}

		//NOTE 任务不能依赖自己。
		SelfID := eventModel.EventModelID
		if _, ok := MapModel[SelfID]; ok {
			logger.Errorf("event_model can't depent itemself!")
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.EventModel_DependentModelIllegal).WithErrorDetails(" event_model can't depent itemself, please check downstream dependent task")
		}
	}

	//NOTE ：触发条件校验
	if eventModel.EventModelType == "atomic" {
		if eventModel.DetectRule.Type != "agi_detect" {
			httpErr = ems.ValidateEventModelDetectRule(ctx, eventModel.DetectRule, eventModel.DetectRule.Type)
			if httpErr != nil {
				logger.Errorf("event_model detect rule validate occur an internal error: %s,%v", eventModel.EventModelName, httpErr.BaseError.Description)
				return httpErr
			}
		}
	} else {
		if eventModel.AggregateRule.Type != "agi_aggregation" {
			ilLegal := ems.ValidateEventModelAggregateRule(ctx, eventModel.AggregateRule, eventModel.AggregateRule.Type)
			if !ilLegal {
				logger.Errorf("event_model aggregate rule validate occur an internal error: %s,%v", eventModel.EventModelName, "aggregate rule is illegal")
				return httpErr
			}
		}
	}

	//NOTE更新持久化配置
	if !common.IsTaskEmpty(eventModel.EventTaskRequest) {
		//NOTE 索引库及存在性校验
		baseinfos, err := ems.iba.GetSimpleIndexBasesByTypes(ctx, []string{eventModel.StorageConfig.IndexBase})
		if err != nil {
			logger.Errorf("GetSimpleIndexBasesByTypes error: %v \n", err)
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				derrors.EventModel_TaskSyncCreateFailed).WithErrorDetails(err.Error())
		}
		if len(baseinfos) == 0 {
			logger.Errorf("the index_base does not exsit\n")
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				derrors.EventModel_TaskSyncCreateFailed).WithErrorDetails("the index_base does not exsit\n")
		}
		//NOTE: 数据视图存在性校验
		_, exist, err := ems.dvs.CheckDataViewExistByID(ctx, nil, eventModel.StorageConfig.DataViewID)
		if err != nil {
			logger.Errorf("CheckDataViewsExistByIDs error: %v \n", err)
			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_DataView_InternalError_CheckViewIfExistFailed).WithErrorDetails(err.Error())
		} else if !exist {
			//说明不存在这个数据视图
			logger.Errorf("data view not found  \n")
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				derrors.EventModel_TaskSyncCreateFailed).WithErrorDetails("Data view is not exist")
		}
	}

	return nil
}

// 创建多个事件模型
func (ems *eventModelService) CreateEventModels(ctx context.Context, eventModels []interfaces.EventModel) ([]map[string]any, error) {

	// 判断userid是否有创建指标模型的权限（策略决策）
	err := ems.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_EVENT_MODEL,
		ID:   interfaces.RESOURCE_ID_ALL,
	}, []string{interfaces.OPERATION_TYPE_CREATE})
	if err != nil {
		return nil, err
	}

	// 0. 开始事务
	tx, err := ems.db.Begin()
	if err != nil {
		logger.Errorf("Begin transaction error: %s", err.Error())
		return []map[string]any{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_EventModel_InternalError_BeginTransactionFailed).
			WithErrorDetails(err.Error())
	}
	// 0.1 异常时
	defer func() {
		switch err {
		case nil:
			// 提交事务
			err = tx.Commit()
			if err != nil {
				logger.Errorf("CreateEventModel Transaction Commit Failed:%v", err)
			}
		default:
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logger.Errorf("CreateMetricModel Transaction Rollback Error:%v", rollbackErr)
			}
		}
	}()

	// 1. 创建模型
	modelInfo, err := ems.ema.CreateEventModels(tx, eventModels)
	if err != nil {
		logger.Errorf("CreateEventModel error: %s", err.Error())
		logger.Errorf("CreateEventModel error,param: %#v", eventModels)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).WithErrorDetails(err.Error())
	}

	now := time.Now().UnixMilli()
	resrcs := []interfaces.Resource{}
	//2. 创建模型下的任务
	for _, eventModel := range eventModels {
		// 添加资源
		resrcs = append(resrcs, interfaces.Resource{
			ID:   eventModel.EventModelID,
			Type: interfaces.RESOURCE_TYPE_EVENT_MODEL,
			Name: eventModel.EventModelName,
		})

		//若配置了持久化任务,设置默认值
		if !common.IsTaskEmpty(eventModel.Task) {
			eventModel.Task = InitDispatchConfig(eventModel.Task)
		} else {
			//配置默认初始化任务
			eventModel.Task = interfaces.EventTask{
				Schedule: interfaces.Schedule{
					Type:       "FIX_RATE",
					Expression: fmt.Sprint(eventModel.DefaultTimeWindow.Interval) + eventModel.DefaultTimeWindow.Unit,
				},
				StorageConfig: interfaces.StorageConfig{
					IndexBase:    interfaces.DEFAULT_INDEX_BASE,
					DataViewName: interfaces.DEFAULT_DATA_VIEW_NAME,
					DataViewID:   interfaces.DEFAULT_DATA_VIEW_ID,
				},
			}
			eventModel.Task = InitDispatchConfig(eventModel.Task)
		}

		taskID := xid.New().String()

		eventModel.Task.TaskID = taskID
		eventModel.Task.ModelID = eventModel.EventModelID
		eventModel.Task.CreateTime = now
		eventModel.Task.UpdateTime = now
		eventModel.Task.StatusUpdateTime = now
		eventModel.Task.ScheduleSyncStatus = interfaces.SCHEDULE_SYNC_STATUS_FINISH
		err = ems.CreateEventTask(ctx, tx, eventModel.Task)
		if err != nil {
			logger.Errorf("CreateEventModelTask error: %s", err.Error())
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).WithErrorDetails(err.Error())
		}
		// 事件模型在创建时，任务是开启的状态才启动任务，针对导入的场景。新创建的需去列表页打开
		if eventModel.IsActive == 1 && eventModel.Status == 1 {
			// 请求 data-model-job 服务开启任务
			eventJobCfg := &interfaces.DataModelJobCfg{
				JobID:      eventModel.Task.TaskID,
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_EVENT_MODEL,
				EventTask:  &eventModel.Task,
				Schedule:   eventModel.Task.Schedule,
			}
			logger.Infof("event model %s request data-model-job start job %d", eventModel.EventModelID, eventModel.Task.TaskID)
			uncancelableCtx := context.WithoutCancel(ctx)
			go func() {
				err := ems.dmja.StartJob(uncancelableCtx, eventJobCfg)
				if err != nil {
					logger.Errorf("Start event job[%s] failed: %s", eventJobCfg.JobID, err.Error())
				}
			}()
		}
	}

	// 注册资源策略
	err = ems.ps.CreateResources(ctx, resrcs, interfaces.COMMON_OPERATIONS)
	if err != nil {
		return nil, err
	}

	return modelInfo, nil
}

// 修改事件模型
func (ems *eventModelService) UpdateEventModel(ctx context.Context, emr interfaces.EventModelUpateRequest) error {

	// 判断userid是否有创建指标模型的权限（策略决策）
	err := ems.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_EVENT_MODEL,
		ID:   emr.EventModelID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	em, httpErr := ems.GetEventModelByID(ctx, emr.EventModelID)
	//NOTE： 如果返回不存在错误或内部错误，则直接返回
	if httpErr != nil {
		return httpErr
	}

	//NOTE 构造事件模型对象
	if emr.EventModelName != em.EventModelName {
		em.EventModelName = emr.EventModelName
	}
	if emr.EventModelComment != em.EventModelComment {
		em.EventModelComment = emr.EventModelComment
	}

	em.EventModelTags = emr.EventModelTags

	if len(emr.DataSource) != 0 {
		em.DataSource = emr.DataSource
	}
	if emr.DataSourceType != "" {
		em.DataSourceType = emr.DataSourceType
	}
	if emr.DefaultTimeWindow.Interval != 0 {
		em.DefaultTimeWindow = interfaces.TimeInterval{
			Interval: emr.DefaultTimeWindow.Interval,
			Unit:     emr.DefaultTimeWindow.Unit}
	}

	em.DownstreamDependentModel = emr.DownstreamDependentModel

	//是否为启停
	flag := false
	//NOTE 向xxljob同步对应任务状态
	//持久化启停状态
	//TODO: 下面这段if是启动/关闭模型的操作对应的。启停操作应另外参考视图的方式另外处理，不应与更新耦合
	// 判断当前是不是定时任务，然后如果启用状态变化了，就说明是要启停这个任务.
	// 通过启用开关启停模型时，对应的是启动和停止任务，更新模型表的status字段，其余操作没有。
	if emr.Status != em.Status && em.EnableSubscribe == 0 || emr.IsActive == 0 {
		flag = true
		var err error
		// if em.Task.TaskScheduleID == 0 {
		// 	logger.Errorf("Event task does not exist")
		// 	return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).
		// 		WithErrorDetails("Event task does not exist")
		// }
		if emr.Status == 1 && emr.IsActive == 1 {
			// 请求 data-model-job 服务开启任务
			modelJobCfg := &interfaces.DataModelJobCfg{
				JobID:      em.Task.TaskID,
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_EVENT_MODEL,
				EventTask:  &em.Task,
				Schedule:   em.Task.Schedule,
			}
			logger.Infof("event model %s request data-model-job start job %d", em.EventModelName, em.Task.TaskID)

			err = ems.dmja.StartJob(ctx, modelJobCfg)
			if err != nil {
				logger.Errorf("Start event job[%s] failed: %s", modelJobCfg.JobID, err.Error())
			}
		} else {
			// err = stopJob(em.Task.TaskScheduleID)
			logger.Infof("event model %s request data-model-job stop job %d", em.EventModelName, em.Task.TaskID)

			err = ems.dmja.StopJobs(ctx, []string{em.Task.TaskID})
			if err != nil {
				logger.Errorf("Start event job[%s] failed: %s", em.Task.TaskID, err.Error())
			}
		}
		if err != nil {
			logger.Errorf("Update IsActive  error: %s", err.Error())
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).
				WithErrorDetails(err.Error())
		}
	}

	em.IsActive = emr.IsActive
	em.EnableSubscribe = emr.EnableSubscribe
	em.Status = emr.Status

	//NOTE 构造检测规则
	now := time.Now().UnixMilli()
	em.UpdateTime = now
	em.DetectRule.UpdateTime = now
	if emr.DetectRule.Type != "" {
		em.DetectRule.Type = emr.DetectRule.Type
	}
	if emr.DetectRule.Formula != nil {
		em.DetectRule.Formula = emr.DetectRule.Formula
	}
	if emr.DetectRule.DetectAlgo != "" {
		em.DetectRule.DetectAlgo = emr.DetectRule.DetectAlgo
	}
	if emr.DetectRule.AnalysisAlgo != nil {
		em.DetectRule.AnalysisAlgo = emr.DetectRule.AnalysisAlgo
	}

	//NOTE 构造聚合规则
	em.AggregateRule.UpdateTime = now
	if emr.AggregateRule.Type != "" {
		em.AggregateRule.Type = emr.AggregateRule.Type
	}
	if emr.AggregateRule.AggregateAlgo != "" {
		em.AggregateRule.AggregateAlgo = emr.AggregateRule.AggregateAlgo
	}
	if len(emr.AggregateRule.GroupFields) != 0 {
		em.AggregateRule.GroupFields = emr.AggregateRule.GroupFields
	}
	if emr.AggregateRule.AnalysisAlgo != nil {
		em.AggregateRule.AnalysisAlgo = emr.AggregateRule.AnalysisAlgo
	}
	//NOTE 处理分析算法
	algo_app_id, _ := strconv.Atoi(emr.AggregateRule.AggregateAlgo)
	if em.AggregateRule.AnalysisAlgo["traceability_analysis"] != "" {
		em.AggregateRule.AnalysisAlgo["traceability_analysis"] = strconv.FormatInt(int64(algo_app_id+1), 10)
	}
	if em.AggregateRule.AnalysisAlgo["problem_convergence"] != "" {
		em.AggregateRule.AnalysisAlgo["problem_convergence"] = strconv.FormatInt(int64(algo_app_id+2), 10)
	}

	// 0. 开始事务
	tx, err := ems.db.Begin()
	if err != nil {
		logger.Errorf("Begin transaction error: %s", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_EventModel_InternalError_BeginTransactionFailed).
			WithErrorDetails(err.Error())
	}
	// 0.1 异常时
	defer func() {
		switch err {
		case nil:
			// 提交事务
			err = tx.Commit()
			if err != nil {
				logger.Errorf("UpdateEventModel Transaction Commit Failed:%v", err)
			}
		default:
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logger.Errorf("UpdateEventModel Transaction Rollback Error:%v", rollbackErr)
			}
		}
	}()

	//NOTE: 更新事件模型表和检测规则表
	err = ems.ema.UpdateEventModel(tx, em)
	if err != nil {
		logger.Errorf("EventModel update error: %s", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).
			WithErrorDetails(err.Error())
	}
	//如果是启停则不需要修改任务。
	if flag {
		return nil
	}

	var task interfaces.EventTask
	task, exist, err := ems.GetEventTaskByModelID(ctx, emr.EventModelID)
	if err != nil {
		logger.Errorf("GetEventTaskByModelID  error: %s", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).
			WithErrorDetails(err.Error())
	}

	//NOTE 更新持久化任务
	if !common.IsTaskEmpty(emr.EventModelRequest.EventTaskRequest) {
		//NOTE 更新持久化配置
		if emr.EventModelRequest.StorageConfig.IndexBase != "" {
			task.StorageConfig.IndexBase = emr.EventModelRequest.StorageConfig.IndexBase
		}
		if emr.EventModelRequest.StorageConfig.DataViewName != "" {
			task.StorageConfig.DataViewName = emr.EventModelRequest.StorageConfig.DataViewName
		}
		if emr.EventModelRequest.StorageConfig.DataViewID != "" {
			task.StorageConfig.DataViewID = emr.EventModelRequest.StorageConfig.DataViewID
		}
		if emr.EventModelRequest.Schedule.Type != "" {
			task.Schedule.Type = emr.EventModelRequest.Schedule.Type
		}

		if emr.EventModelRequest.Schedule.Expression != "" {
			task.Schedule.Expression = emr.EventModelRequest.Schedule.Expression
		}
		if emr.EventModelRequest.ExecuteParameter != nil {
			task.ExecuteParameter = emr.EventModelRequest.ExecuteParameter
		}

		if emr.EventModelRequest.DispatchConfig.TimeOut != 0 {
			task.DispatchConfig.TimeOut = emr.EventModelRequest.DispatchConfig.TimeOut
		}
		if emr.EventModelRequest.DispatchConfig.FailRetryCount != 0 {
			task.DispatchConfig.FailRetryCount = emr.EventModelRequest.DispatchConfig.FailRetryCount
		}
		if emr.EventModelRequest.DispatchConfig.RouteStrategy != "" {
			task.DispatchConfig.RouteStrategy = emr.EventModelRequest.DispatchConfig.RouteStrategy
		}
		if emr.EventModelRequest.DispatchConfig.BlockStrategy != "" {
			task.DispatchConfig.BlockStrategy = emr.EventModelRequest.DispatchConfig.BlockStrategy
		}

		//NOTE 当任务依赖更新时，需要基于事件模型ID获取对应的调度依赖id，一起更新。
		var TaksIDs []string
		for _, model_id := range emr.EventModelRequest.DownstreamDependentModel {
			task, _, _ := ems.GetEventTaskByModelID(ctx, model_id)
			TaksIDs = append(TaksIDs, task.TaskID)
		}
		task.DownstreamDependentTask = TaksIDs

		//不存在则创建
		if !exist {
			taskID := xid.New().String()

			accountInfo := interfaces.AccountInfo{}
			if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
				accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
			}

			task.TaskID = taskID
			task.ModelID = emr.EventModelID
			task.Creator = accountInfo
			task.UpdateTime = now
			task.StatusUpdateTime = now
			task.ScheduleSyncStatus = interfaces.SCHEDULE_SYNC_STATUS_FINISH

			//设置默认值
			if task.DispatchConfig.TimeOut == 0 {
				task.DispatchConfig.TimeOut = DEFAULT_DISPACH_CONFIG_TIME_OUT
			}
			if task.DispatchConfig.FailRetryCount == 0 {
				task.DispatchConfig.FailRetryCount = DEFAULT_FAIL_RETRY_COUNT
			}
			if task.DispatchConfig.RouteStrategy == "" {
				task.DispatchConfig.RouteStrategy = "ROUND"
			}
			if task.DispatchConfig.BlockStrategy == "" {
				task.DispatchConfig.BlockStrategy = "SERIAL_EXECUTION"
			}

			err = ems.CreateEventTask(ctx, tx, task)
			if err != nil {
				logger.Errorf("CreateEventModelTask error: %s", err.Error())
				return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).WithErrorDetails(err.Error())
			}
			// ems.sendTaskToChan(task.TaskID)
			if em.IsActive == 1 && em.Status == 1 {
				// 请求 data-model-job 服务开启任务
				eventJobCfg := &interfaces.DataModelJobCfg{
					JobID:      task.TaskID,
					JobType:    interfaces.JOB_TYPE_SCHEDULE,
					ModuleType: interfaces.MODULE_TYPE_EVENT_MODEL,
					EventTask:  &task,
					Schedule:   task.Schedule,
				}
				logger.Infof("event model %s request data-model-job start job %d", em.EventModelID, task.TaskID)
				uncancelableCtx := context.WithoutCancel(ctx)
				go func() {
					err := ems.dmja.StartJob(uncancelableCtx, eventJobCfg)
					if err != nil {
						logger.Errorf("Start event job[%s] failed: %s", eventJobCfg.JobID, err.Error())
					}
				}()
			}
			return nil
		}

		task.ScheduleSyncStatus = interfaces.SCHEDULE_SYNC_STATUS_FINISH
		task.UpdateTime = now
		err = ems.UpdateEventTask(ctx, tx, task)
		if err != nil {
			logger.Errorf("UpdateEventTask error: %s", err.Error())
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).WithErrorDetails(err.Error())
		}
		// ems.sendTaskToChan(task.TaskID)
		if em.IsActive == 1 && em.Status == 1 {
			// 只有启用状态的任务变更才更新到data-model-job中，其他的只更新数据库
			// 请求 data-model-job 更新任务
			newEventJobCfg := &interfaces.DataModelJobCfg{
				JobID:      task.TaskID,
				JobType:    interfaces.JOB_TYPE_SCHEDULE,
				ModuleType: interfaces.MODULE_TYPE_EVENT_MODEL,
				EventTask:  &task,
				Schedule:   task.Schedule,
			}
			logger.Infof("event model %s request data-model-job update job %d", em.EventModelName, em.Task.TaskID)

			uncancelableCtx := context.WithoutCancel(ctx)
			go func() {
				err := ems.dmja.UpdateJob(uncancelableCtx, newEventJobCfg)
				if err != nil {
					logger.Errorf("Update event model job[%s] failed, %s", newEventJobCfg.JobID, err.Error())
				}
			}()
		}
	}

	err = ems.ps.UpdateResource(ctx, interfaces.Resource{
		ID:   emr.EventModelID,
		Type: interfaces.RESOURCE_TYPE_EVENT_MODEL,
		Name: emr.EventModelName,
	})
	if err != nil {
		return err
	}

	return nil
}

// 按 id 获取事件模型信息
func (ems *eventModelService) GetEventModelByID(ctx context.Context, modelID string) (interfaces.EventModel, *rest.HTTPError) {

	eventModel, err := ems.ema.GetEventModelByID(modelID)
	if err != nil {
		if err.Error() == derrors.EventModel_EventModelNotFound {
			return interfaces.EventModel{}, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.EventModel_EventModelNotFound).
				WithErrorDetails(err.Error())
		} else {
			return interfaces.EventModel{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).
				WithErrorDetails(err.Error())
		}
	}

	// 判断userid是否有查看模型的权限（策略决策）
	err = ems.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_EVENT_MODEL,
		ID:   eventModel.EventModelID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return interfaces.EventModel{}, err.(*rest.HTTPError)
	}

	ModelNames := make([]string, 0, len(eventModel.DataSource))
	GroupNames := make([]string, 0, len(eventModel.DataSource))
	if eventModel.DataSourceType == "metric_model" {
		IDNameMap, _ := ems.mms.GetMetricModelSimpleInfosByIDs(ctx, eventModel.DataSource)
		for _, value := range IDNameMap {
			ModelNames = append(ModelNames, value.ModelName)
			GroupNames = append(GroupNames, value.GroupName)
		}

	} else if eventModel.DataSourceType == "event_model" {
		ModelNames, _ = ems.ema.GetEventModelNamesByIDs(eventModel.DataSource)
	} else if eventModel.DataSourceType == "data_view" {
		IDNameMap, _ := ems.dvs.GetDataViews(ctx, eventModel.DataSource, false)
		for _, view := range IDNameMap {
			ModelNames = append(ModelNames, view.ViewName)
			GroupNames = append(GroupNames, "")

		}
	}

	eventModel.DataSourceName = ModelNames
	eventModel.DataSourceGroupName = GroupNames

	//NOTE  添加持久化任务信息
	task, exist, err := ems.GetEventTaskByModelID(ctx, modelID)
	if err != nil {
		return interfaces.EventModel{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).
			WithErrorDetails(err.Error())
	}
	if exist {
		eventModel.Task = task
	}
	return eventModel, nil
}

// 批量删除事件模型
func (ems *eventModelService) DeleteEventModels(ctx context.Context, modelIDs []string) ([]interfaces.EventModel, error) {

	// 先获取资源序列 fmt.Sprintf("%s%s", interfaces.METRIC_MODEL_RESOURCE_ID_PREFIX, metricModel.ModelID),
	matchResouces, err := ems.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_EVENT_MODEL, modelIDs,
		[]string{interfaces.OPERATION_TYPE_DELETE}, false)
	if err != nil {
		return nil, err
	}
	// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
	for _, mID := range modelIDs {
		if _, exist := matchResouces[mID]; !exist {
			return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for event model's delete operation.")
		}
	}

	var models []interfaces.EventModel
	// 按modelid获取任务id列表
	taskIDs, err := ems.GetEventTaskIDByModelIDs(ctx, modelIDs)
	if err != nil {
		logger.Errorf("DeleteEventModels error: %s", err.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).
			WithErrorDetails(err.Error())
	}
	// 请求data-model-job 批量停止任务
	if len(taskIDs) > 0 {
		// 请求data-model-job服务批量停止实际运行的任务
		if err = ems.dmja.StopJobs(ctx, taskIDs); err != nil {
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_StopJobFailed).WithErrorDetails(err.Error())
		}
	}

	// 批量查询到对象对象，用于获取检测规则id和事件模型名称
	//NOTE 任一事件模型对象不存在，均会导致所有的删除取消
	//INFO: 检查 modelIDs 是否都存在
	for _, id := range modelIDs {
		em, httpErr := ems.GetEventModelByID(ctx, id)
		if httpErr != nil {
			logger.Errorf("query event model  failed by id: %v", id)
			return nil, httpErr
		}

		refs, httpErr := ems.GetEventModelRefs(ctx, id)
		if httpErr != nil {
			logger.Errorf("query event model  refs by id: %v", id)
			return nil, httpErr
		}
		if refs > 0 {
			logger.Errorf("query event model  refs by id: %v", id)
			httpErr = rest.NewHTTPError(ctx, http.StatusLocked, derrors.EventModel_RefByOther).
				WithErrorDetails(fmt.Sprintf("this event model %s has been reference by another %s event_model", id, strconv.Itoa(refs)))
			return nil, httpErr
		}
		refs, httpErr = ems.GetEventModelDependence(ctx, id)
		if httpErr != nil {
			logger.Errorf("query event model  dependent by id: %v", id)
			return nil, httpErr
		}
		if refs > 0 {
			logger.Errorf("query event model  dependent by id: %v", id)
			httpErr = rest.NewHTTPError(ctx, http.StatusLocked, derrors.EventModel_RefByOther).
				WithErrorDetails(fmt.Sprintf("this event model %s has been dependent by another %s event_model", id, strconv.Itoa(refs)))
			return nil, httpErr
		}
		models = append(models, em)
	}

	// if len(ems) == 0 {

	// 	return nil, rest.NewHTTPError(ctx, http.StatusNotFound,
	// 		derrors.EventModel_EventModelNotFound).WithErrorDetails("no one event model id exist")
	// }

	// 0. 开始事务
	tx, err := ems.db.Begin()
	if err != nil {
		logger.Errorf("Begin transaction error: %s", err.Error())

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_EventModel_InternalError_BeginTransactionFailed).
			WithErrorDetails(err.Error())
	}
	// 0.1 异常时
	defer func() {
		switch err {
		case nil:
			// 提交事务
			err = tx.Commit()
			if err != nil {
				logger.Errorf("DeleteEventModels Transaction Commit Failed:%v", err)
			}
			logger.Infof("DeleteEventModels Transaction Commit Success:%v", modelIDs)

		default:
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logger.Errorf("DeleteEventModels Transaction Rollback Error:%v", rollbackErr)
			}
		}
	}()

	err = ems.ema.DeleteEventModels(tx, models)
	if err != nil {
		logger.Errorf("Delete EventModel error: %s", err.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.EventModel_InternalError).WithErrorDetails(err.Error())
	}

	logger.Infof("Delete Event Model success:  request delete modelID is %v!", modelIDs)

	// 物理删除事件模型下的持久化任务
	err = ems.DeleteEventTaskByTaskIDs(ctx, tx, taskIDs)
	if err != nil {
		logger.Errorf("DeleteEventModels error: %s", err.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).
			WithErrorDetails(err.Error())
	}

	//  清除资源策略
	err = ems.ps.DeleteResources(ctx, interfaces.RESOURCE_TYPE_EVENT_MODEL, modelIDs)
	if err != nil {
		return nil, err
	}

	return models, nil
}

// 分页查询事件模型
func (ems *eventModelService) QueryEventModels(ctx context.Context, params interfaces.EventModelQueryRequest) (
	[]interfaces.EventModel, int, error) {

	//NOTE 根据查询条件查询对应的事件模型,带翻页参数
	eventModels, err := ems.ema.QueryEventModels(ctx, params)
	if err != nil {
		logger.Errorf("Query Event Model Error: %s", err.Error())
		return eventModels, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).WithErrorDetails(err.Error())
	}
	if len(eventModels) == 0 {
		return eventModels, 0, err
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	// 处理资源id
	resMids := []string{}
	for _, m := range eventModels {
		resMids = append(resMids, m.EventModelID)
	}
	matchResoucesMap, err := ems.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_EVENT_MODEL, resMids,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true)
	if err != nil {
		return []interfaces.EventModel{}, 0, nil
	}

	// 遍历对象
	results := make([]interfaces.EventModel, 0)
	for _, model := range eventModels {
		if resrc, exist := matchResoucesMap[model.EventModelID]; exist {
			model.Operations = resrc.Operations // 用户当前有权限的操作
			results = append(results, model)
		}
	}

	// limit = -1,则返回所有
	if params.Limit == -1 {
		return results, len(results), nil
	}

	// 分页
	// 检查起始位置是否越界
	if params.Offset < 0 || params.Offset >= len(results) {
		return []interfaces.EventModel{}, 0, nil
	}
	// 计算结束位置
	end := params.Offset + params.Limit
	if end > len(results) {
		end = len(results)
	}

	return results[params.Offset:end], len(results), nil
}

// 根据事件模型名称获取IDs
func (ems *eventModelService) GetEventModelMapByNames(modelNames []string) (map[string]string, error) {
	return ems.ema.GetEventModelMapByNames(modelNames)
}

// 根据事件模型IDs获取名称
func (ems *eventModelService) GetEventModelMapByIDs(modelIDs []string) (map[string]string, error) {
	return ems.ema.GetEventModelMapByIDs(modelIDs)
}

func (ems *eventModelService) GetEventModelRefs(ctx context.Context, modelID string) (int, *rest.HTTPError) {
	refs, err := ems.ema.GetEventModelRefsByID(modelID)
	if err != nil {

		return 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).
			WithErrorDetails(err.Error())

	}
	return refs, nil

}

func (ems *eventModelService) GetEventModelDependence(ctx context.Context, modelID string) (int, *rest.HTTPError) {
	refs, err := ems.ema.GetEventModelDependenceByID(modelID)
	if err != nil {

		return 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).
			WithErrorDetails(err.Error())

	}
	return refs, nil

}

func (ems *eventModelService) UpdateEventTaskAttributes(ctx context.Context, task interfaces.EventTask) error {
	return ems.ema.UpdateEventTaskAttributes(ctx, task)
}

// 创建持久化任务
func (ems *eventModelService) CreateEventTask(ctx context.Context, tx *sql.Tx, task interfaces.EventTask) error {
	return ems.ema.CreateEventTask(ctx, tx, task)
}

// 按模型id获取持久化任务id
func (ems *eventModelService) GetEventTaskIDByModelIDs(ctx context.Context, modelIDs []string) ([]string, error) {
	return ems.ema.GetEventTaskIDByModelIDs(ctx, modelIDs)
}

// 更新任务
func (ems *eventModelService) UpdateEventTask(ctx context.Context, tx *sql.Tx, task interfaces.EventTask) error {
	return ems.ema.UpdateEventTask(ctx, tx, task)
}

// 按任务id批量删除持久化任务，逻辑删除，把状态设置为删除中
// func (ems *eventModelService) SetTaskSyncStatusByTaskID(ctx context.Context, tx *sql.Tx, taskSyncStatus interfaces.EventTaskSyncStatus) error {
// 	return ems.ema.SetTaskSyncStatusByTaskID(ctx, tx, taskSyncStatus)
// }

// // 按模型id批量删除持久化任务，逻辑删除，把状态设置为删除中
// func (ems *eventModelService) SetTaskSyncStatusByModelID(ctx context.Context, tx *sql.Tx, taskSyncStatus interfaces.EventTaskSyncStatus) error {
// 	return ems.ema.SetTaskSyncStatusByModelID(ctx, tx, taskSyncStatus)
// }

// 按任务id批量获取任务信息
func (ems *eventModelService) GetEventTaskByTaskID(ctx context.Context, taskID string) (interfaces.EventTask, error) {
	return ems.ema.GetEventTaskByTaskID(ctx, taskID)
}

// 按模型id获取任务信息 内部校验调用无需权限
func (ems *eventModelService) GetEventTaskByModelID(ctx context.Context, modelID string) (interfaces.EventTask, bool, error) {
	return ems.ema.GetEventTaskByModelID(ctx, modelID)
}

// 根据任务id，物理删除任务，由删除事件模型的权限来控制，不单独校验
func (ems *eventModelService) DeleteEventTaskByTaskIDs(ctx context.Context, tx *sql.Tx, taskIDs []string) error {
	return ems.ema.DeleteEventTaskByTaskIDs(ctx, tx, taskIDs)
}

// 获取正在进行中的任务
// func (ems *eventModelService) GetProcessingEventTasks(ctx context.Context) ([]interfaces.EventTask, error) {
// 	return ems.ema.GetProcessingEventTasks(ctx)
// }

// 更新任务状态为完成，更新调度id
func (ems *eventModelService) UpdateEventTaskStatusInFinish(ctx context.Context, task interfaces.EventTask) error {
	return ems.ema.UpdateEventTaskStatusInFinish(ctx, task)
}

func InitDispatchConfig(task interfaces.EventTask) interfaces.EventTask {
	//设置默认值
	if task.DispatchConfig.TimeOut == 0 {
		task.DispatchConfig.TimeOut = DEFAULT_DISPACH_CONFIG_TIME_OUT
	}
	if task.DispatchConfig.FailRetryCount == 0 {
		task.DispatchConfig.FailRetryCount = DEFAULT_FAIL_RETRY_COUNT
	}
	if task.DispatchConfig.RouteStrategy == "" {
		task.DispatchConfig.RouteStrategy = "ROUND"
	}
	if task.DispatchConfig.BlockStrategy == "" {
		task.DispatchConfig.BlockStrategy = "SERIAL_EXECUTION"
	}
	return task
}

// todo: 校验参数应放在handler层
func (eventModelService *eventModelService) ValidateExecuteParam(ctx context.Context, executeParam map[string]any) (bool, error) {
	start, ok := executeParam["start"]
	var start_timestamp, end_timestamp float64
	if ok {
		start_timestamp = start.(float64)
	}
	end, ok := executeParam["end"]
	if ok {
		end_timestamp = end.(float64)
	}
	if start_timestamp >= end_timestamp && start_timestamp > 0 {
		return false, errors.New("执行参数不合理，填入的开始时间不能大于等于结束时间")
	}
	return true, nil
}

func (ems *eventModelService) ListEventModelSrcs(ctx context.Context,
	params interfaces.EventModelQueryRequest) ([]interfaces.Resource, int, error) {

	emptyResources := []interfaces.Resource{}

	//NOTE 根据查询条件查询对应的事件模型,带翻页参数
	eventModels, err := ems.ema.QueryEventModels(ctx, params)
	if err != nil {
		logger.Errorf("Query Event Model Error: %s", err.Error())
		return emptyResources, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.EventModel_InternalError).WithErrorDetails(err.Error())
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	// 处理资源id
	resMids := make([]string, 0)
	for _, m := range eventModels {
		resMids = append(resMids, m.EventModelID)
	}
	matchResoucesMap, err := ems.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_EVENT_MODEL, resMids,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, false)
	if err != nil {
		return emptyResources, 0, err
	}

	// 遍历对象
	results := make([]interfaces.Resource, 0)
	for _, model := range eventModels {
		if _, exist := matchResoucesMap[model.EventModelID]; exist {
			results = append(results, interfaces.Resource{
				ID:   model.EventModelID,
				Type: interfaces.RESOURCE_TYPE_EVENT_MODEL,
				Name: model.EventModelName,
			})
		}
	}

	// limit = -1,则返回所有
	if params.Limit == -1 {
		return results, len(results), nil
	}

	// 分页
	// 检查起始位置是否越界
	if params.Offset < 0 || params.Offset >= len(results) {
		return nil, 0, nil
	}
	// 计算结束位置
	end := params.Offset + params.Limit
	if end > len(results) {
		end = len(results)
	}

	return results[params.Offset:end], len(results), nil
}
