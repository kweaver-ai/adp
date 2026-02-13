// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

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
)

var (
	mmgServiceOnce sync.Once
	mmgService     interfaces.MetricModelGroupService
)

type metricModelGroupService struct {
	appSetting *common.AppSetting
	db         *sql.DB
	mma        interfaces.MetricModelAccess
	mmga       interfaces.MetricModelGroupAccess
	mms        interfaces.MetricModelService
}

func NewMetricModelGroupService(appSetting *common.AppSetting) interfaces.MetricModelGroupService {
	mmgServiceOnce.Do(func() {
		mmgService = &metricModelGroupService{
			appSetting: appSetting,
			db:         logics.DB,
			mma:        logics.MMA,
			mmga:       logics.MMGA,
			mms:        NewMetricModelService(appSetting),
		}
	})
	return mmgService
}

// 按ID获取指标模型分组信息
func (mmgs *metricModelGroupService) GetMetricModelGroupByID(ctx context.Context,
	groupID string) (interfaces.MetricModelGroup, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("查询指标模型分组[%s]信息", groupID))
	span.SetAttributes(attr.Key("group_id").String(groupID))
	defer span.End()

	var mmg interfaces.MetricModelGroup
	metricModelGroup, exist, err := mmgs.mmga.GetMetricModelGroupByID(ctx, groupID)
	if err != nil {
		logger.Errorf("GetMetricModelGroupByID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get metric model group [%s] error: %v", groupID, err))
		return mmg, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModelGroup_InternalError_GetGroupByIDFailed).WithErrorDetails(err.Error())
	}
	if !exist {
		logger.Debugf("Metric Model Group %s not found!", groupID)
		span.SetStatus(codes.Error, fmt.Sprintf("Metric model Group [%s] not found", groupID))
		return mmg, rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_MetricModelGroup_GroupNotFound)
	}

	span.SetStatus(codes.Ok, "")
	return metricModelGroup, nil
}

// 根据groupName检查Group是否存在
func (mmgs *metricModelGroupService) CheckMetricModelGroupExist(ctx context.Context, groupName string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("校验指标模型分组[%s]的存在性", groupName))
	span.SetAttributes(attr.Key("group_name").String(groupName))
	defer span.End()

	_, exist, err := mmgs.mmga.CheckMetricModelGroupExist(ctx, groupName)
	if err != nil {
		logger.Errorf("CheckMetricModelGroupExistByName error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("按名称[%s]获取指标模型分组失败", groupName))
		return exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModelGroup_InternalError_CheckGroupIfExistFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return exist, nil

}

// 创建指标模型分组
func (mmgs *metricModelGroupService) CreateMetricModelGroup(ctx context.Context,
	metricModelGroup interfaces.MetricModelGroup) (string, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Create metric model group")
	defer span.End()

	// 生成分布式ID
	metricModelGroup.GroupID = xid.New().String()
	metricModelGroup.UpdateTime = time.Now().UnixMilli()

	span.SetAttributes(
		attr.Key("metric_model_group_id").String(metricModelGroup.GroupID),
		attr.Key("metric_model_group_name").String(metricModelGroup.GroupName))

	//调用driven层创建指标模型分组
	err := mmgs.mmga.CreateMetricModelGroup(ctx, nil, metricModelGroup)
	if err != nil {
		logger.Errorf("CreateMetricModelGroup error: %s", err.Error())
		span.SetStatus(codes.Error, "创建指标模型分组失败")

		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModelGroup_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return metricModelGroup.GroupID, nil
}

// 修改指标模型分组
func (mmgs *metricModelGroupService) UpdateMetricModelGroup(ctx context.Context,
	metricModelGroup interfaces.MetricModelGroup) error {

	ctx, span := ar_trace.Tracer.Start(ctx, "Update metric model group")
	span.SetAttributes(
		attr.Key("metric_model_group_id").String(metricModelGroup.GroupID),
		attr.Key("metric_model_group_name").String(metricModelGroup.GroupName))
	defer span.End()

	metricModelGroup.UpdateTime = time.Now().UnixMilli()
	err := mmgs.mmga.UpdateMetricModelGroup(ctx, metricModelGroup)
	if err != nil {
		logger.Errorf("metricModelGroup error: %s", err.Error())
		span.SetStatus(codes.Error, "修改指标模型分组失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModelGroup_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 查询指标模型分组列表
func (mmgs *metricModelGroupService) ListMetricModelGroups(ctx context.Context,
	params interfaces.ListMetricGroupQueryParams) ([]*interfaces.MetricModelGroup, int, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "查询指标模型分组列表")
	defer span.End()

	//获取指标模型分组列表
	groups, err := mmgs.mmga.ListMetricModelGroups(ctx, params)
	if err != nil {
		logger.Errorf("ListMetricModelGroups error: %s", err.Error())
		span.SetStatus(codes.Error, "List metric model groups error")

		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModelGroup_InternalError_ListGroupsFailed).WithErrorDetails(err.Error())
	}
	total, err := mmgs.mmga.GetMetricModelGroupsTotal(ctx, params)
	if err != nil {
		logger.Errorf("GetDataViewGroupsTotal error: %s", err.Error())

		span.SetStatus(codes.Error, "Get data view groups total failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModelGroup_InternalError_GetGroupsTotalFailed).WithErrorDetails(err.Error())
	}

	// 获取每组下的有查看权限的模型数量
	// 方法1；list所有模型，按权限过滤，对过滤完成的数据按组进行分组计数。
	models, _, err := mmgs.mms.ListSimpleMetricModels(ctx, interfaces.MetricModelsQueryParams{
		GroupID: interfaces.GroupID_All,
		PaginationQueryParameters: interfaces.PaginationQueryParameters{
			Limit: -1,
			Sort:  interfaces.METRIC_MODEL_SORT["group_name"],
		},
	})
	if err != nil {
		return nil, 0, err
	}

	groupMetricCnt := map[string]int{}
	for _, model := range models {
		if cnt, exist := groupMetricCnt[model.GroupID]; !exist {
			groupMetricCnt[model.GroupID] = 1
		} else {
			groupMetricCnt[model.GroupID] = cnt + 1
		}
	}

	// 赋值
	for i := range groups {
		if cnt, exist := groupMetricCnt[groups[i].GroupID]; exist {
			groups[i].MetricModelCount = cnt
		} else {
			groups[i].MetricModelCount = 0
		}
	}

	span.SetStatus(codes.Ok, "")
	return groups, total, nil
}

func (mmgs *metricModelGroupService) DeleteMetricModelGroup(ctx context.Context, groupID string, force bool) (int64, []interfaces.MetricModel, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete metric model groups")
	defer span.End()

	metricModels, err := mmgs.mma.GetMetricModelsByGroupID(ctx, groupID)
	if err != nil {
		logger.Errorf("DeleteMetricModelGroup error: %s", err.Error())
		span.SetStatus(codes.Error, "删除指标模型分组失败")
		return 0, metricModels, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModelGroup_InternalError).WithErrorDetails(err.Error())
	}

	modelIDs := make([]string, 0)
	for _, model := range metricModels {
		modelIDs = append(modelIDs, model.ModelID)
	}
	var rowsAffect int64
	if force {
		// 删除分组同时删除所有指标模型
		rowsAffect, err = mmgs.DeleteMetricModelGroupAndModels(ctx, groupID, modelIDs)
		if err != nil {
			logger.Errorf("DeleteMetricModelGroup error: %s", err.Error())
			span.SetStatus(codes.Error, "删除指标模型分组失败")
			return 0, metricModels, err
		}

	} else if !force && len(modelIDs) == 0 {
		//删除分组
		_, err := mmgs.DeleteMetricModelGroupAndModels(ctx, groupID, modelIDs)
		if err != nil {
			logger.Errorf("DeleteMetricModelGroup error: %s", err.Error())
			span.SetStatus(codes.Error, "删除指标模型分组失败")
			return 0, metricModels, err
		}

	} else {
		// 分组不为空 禁止删除
		return 0, metricModels, rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_MetricModelGroup_GroupNotEmpty).
			WithErrorDetails("Delete Group Failed: The group is not empty")
	}

	span.SetStatus(codes.Ok, "")
	return rowsAffect, metricModels, nil
}

// 删除分组和分组内指标模型
func (mmgs *metricModelGroupService) DeleteMetricModelGroupAndModels(ctx context.Context, groupID string,
	modelIDs []string) (int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete metric model groups")
	defer span.End()

	// 起事务，删除分组，删除组内模型，模型对应的任务更新为删除中
	// 0. 开始事务
	tx, err := mmgs.db.Begin()
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
				logger.Errorf("DeleteMetricModelGroupAndModels Transaction Commit Failed:%v", err)
				span.SetStatus(codes.Error, "提交事务失败")
				o11y.Error(ctx, fmt.Sprintf("DeleteMetricModelGroupAndModels Transaction Commit Failed: %s", err.Error()))

			}
			logger.Infof("DeleteMetricModelGroupAndModels Transaction Commit Success:%v", modelIDs)
			o11y.Debug(ctx, fmt.Sprintf("DeleteMetricModelGroupAndModels Transaction Commit Success: %v", modelIDs))
		default:
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logger.Errorf("DeleteMetricModelGroupAndModels Transaction Rollback Error:%v", rollbackErr)
				span.SetStatus(codes.Error, "事务回滚失败")
				o11y.Error(ctx, fmt.Sprintf("DeleteMetricModelGroupAndModels Transaction Rollback Error: %s", err.Error()))
			}
		}
	}()

	// 删除分组
	//rowsAffect 记录影响分组行数
	rowsAffect, err := mmgs.mmga.DeleteMetricModelGroup(ctx, tx, groupID)
	span.SetAttributes(attr.Key("rows_affect").Int64(rowsAffect))
	if err != nil {
		logger.Errorf("DeleteMetricModelGroupAndModels error: %s", err.Error())
		span.SetStatus(codes.Error, "删除指标模型分组失败")

		return rowsAffect, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModelGroup_InternalError).WithErrorDetails(err.Error())
	}

	// 模型不为空时删除模型（在删除模型中会把模型对应的任务置为删除中）
	if len(modelIDs) != 0 {
		rowsAffect, err = mmgs.mms.DeleteMetricModels(ctx, tx, modelIDs)
		if err != nil {
			logger.Errorf("DeleteMetricModels error: %s", err.Error())
			span.SetStatus(codes.Error, "删除指标模型分组内的指标模型失败")

			return rowsAffect, err
		}
	}

	logger.Infof("DeleteMetricModelGroupAndModels: Rows affected is %v, request delete group is %v!", rowsAffect, 1)
	if rowsAffect != 1 {
		logger.Warnf("Delete model group number %v not equal request model group number %v!", rowsAffect, 1)

		o11y.Warn(ctx, fmt.Sprintf("Delete model group number %v not equal request model group number %v!", rowsAffect, 1))
	}
	span.SetStatus(codes.Ok, "")
	return rowsAffect, nil
}
