// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric_model

import (
	"context"
	"database/sql"
	"sync"

	"data-model/common"
	"data-model/interfaces"
	"data-model/logics"
)

var (
	mmtServiceOnce sync.Once
	mmtService     interfaces.MetricModelTaskService
)

type metricModelTaskService struct {
	appSetting *common.AppSetting
	mmta       interfaces.MetricModelTaskAccess
}

func NewMetricModelTaskService(appSetting *common.AppSetting) interfaces.MetricModelTaskService {
	mmtServiceOnce.Do(func() {
		mmtService = &metricModelTaskService{
			appSetting: appSetting,
			mmta:       logics.MMTA,
		}
	})
	return mmtService
}

// 批量创建持久化任务
func (mmts *metricModelTaskService) CreateMetricTask(ctx context.Context, tx *sql.Tx, task interfaces.MetricTask) error {
	return mmts.mmta.CreateMetricTask(ctx, tx, task)
}

// 按指标模型id获取持久化任务id列表
func (mmts *metricModelTaskService) GetMetricTaskIDsByModelIDs(ctx context.Context, modelIDs []string) ([]string, error) {
	return mmts.mmta.GetMetricTaskIDsByModelIDs(ctx, modelIDs)
}

// 更新任务
func (mmts *metricModelTaskService) UpdateMetricTask(ctx context.Context, tx *sql.Tx, task interfaces.MetricTask) error {
	return mmts.mmta.UpdateMetricTask(ctx, tx, task)
}

// 按任务id批量删除持久化任务，逻辑删除，把状态设置为删除中
// func (mmts *metricModelTaskService) SetTaskSyncStatusByTaskIDs(ctx context.Context, tx *sql.Tx, taskSyncStatus interfaces.TaskSyncStatus) error {
// 	return mmts.mmta.SetTaskSyncStatusByTaskIDs(ctx, tx, taskSyncStatus)
// }

// 按模型id批量删除持久化任务，逻辑删除，把状态设置为删除中
// func (mmts *metricModelTaskService) SetTaskSyncStatusByModelIDs(ctx context.Context, tx *sql.Tx, taskSyncStatus interfaces.TaskSyncStatus) error {
// 	return mmts.mmta.SetTaskSyncStatusByModelIDs(ctx, tx, taskSyncStatus)
// }

// 按任务id批量获取任务信息
func (mmts *metricModelTaskService) GetMetricTasksByTaskIDs(ctx context.Context, taskIDs []string) ([]interfaces.MetricTask, error) {
	return mmts.mmta.GetMetricTasksByTaskIDs(ctx, taskIDs)
}

// 按模型id批量获取任务信息
func (mmts *metricModelTaskService) GetMetricTasksByModelIDs(ctx context.Context, modelIDs []string) (map[string]interfaces.MetricTask, error) {
	return mmts.mmta.GetMetricTasksByModelIDs(ctx, modelIDs)
}

// 获取正在进行中的任务
// func (mmts *metricModelTaskService) GetProcessingMetricTasks(ctx context.Context) ([]interfaces.MetricTask, error) {
// 	return mmts.mmta.GetProcessingMetricTasks(ctx)
// }

// 更新任务状态为完成，更新调度id
func (mmts *metricModelTaskService) UpdateMetricTaskStatusInFinish(ctx context.Context, task interfaces.MetricTask) error {
	return mmts.mmta.UpdateMetricTaskStatusInFinish(ctx, task)
}

func (mmts *metricModelTaskService) UpdateMetricTaskAttributes(ctx context.Context, task interfaces.MetricTask) error {
	return mmts.mmta.UpdateMetricTaskAttributes(ctx, task)
}

// 根据任务id，物理删除任务
func (mmts *metricModelTaskService) DeleteMetricTaskByTaskIDs(ctx context.Context, tx *sql.Tx, taskIDs []string) error {
	return mmts.mmta.DeleteMetricTaskByTaskIDs(ctx, tx, taskIDs)
}

// 校验模型下的任务名称是否已存在
func (mmts *metricModelTaskService) CheckMetricModelTaskExistByName(ctx context.Context, task interfaces.MetricTask, deleteTaskIDs []string) (bool, error) {
	return mmts.mmta.CheckMetricModelTaskExistByName(ctx, task, deleteTaskIDs)
}
