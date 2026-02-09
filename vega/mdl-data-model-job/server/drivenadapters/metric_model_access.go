// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"data-model-job/common"
	"data-model-job/interfaces"
)

var (
	mmAccessOnce sync.Once
	mmAccess     interfaces.MetricModelAccess
)

type metricModelAccess struct {
	metricTaskUrl string
	httpClient    rest.HTTPClient
}

func NewMetricModelAccess(appSetting *common.AppSetting) interfaces.MetricModelAccess {
	mmAccessOnce.Do(func() {
		mmAccess = &metricModelAccess{
			metricTaskUrl: appSetting.MetricTaskUrl,
			httpClient:    common.NewHTTPClient(),
		}
	})

	return mmAccess
}

// 按任务id获取指标持久化任务的计划时间
func (mma *metricModelAccess) GetTaskPlanTimeById(ctx context.Context, taskId string) (int64, error) {
	url := fmt.Sprintf("%s/%s", mma.metricTaskUrl, taskId)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}
	respCode, respData, err := mma.httpClient.GetNoUnmarshal(ctx, url, nil, headers)
	if err != nil {
		logger.Errorf("Get the plan time of metric model task by taskid '%s' failed, %s", taskId, err)

		return 0, err
	}

	if respCode != http.StatusOK {
		// 转成 baseerror
		var baseError rest.BaseError
		if err := json.Unmarshal(respData, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)

			return 0, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode, BaseError: baseError}
		logger.Errorf("get metric task return error: %v", httpErr.Error())

		return 0, fmt.Errorf("get metric task %s return error %v", taskId, httpErr.Error())
	}

	if respData == nil {
		return 0, fmt.Errorf("get metric task %v return null", taskId)
	}

	var task interfaces.MetricTask
	if err := json.Unmarshal(respData, &task); err != nil {
		logger.Errorf("Unmarshal metric task respData failed, %s", err)

		return 0, err
	}

	return task.PlanTime, nil
}

// 更新任务计划时间为调度执行的时间戳
func (mma *metricModelAccess) UpdateTaskAttributesById(ctx context.Context,
	taskId string, task interfaces.MetricTask) error {

	url := fmt.Sprintf("%s/%s/attr", mma.metricTaskUrl, taskId)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}
	respCode, respData, err := mma.httpClient.PutNoUnmarshal(ctx, url, headers, task)
	logger.Debugf("get [%s] with headers[%v], task[%v] finished, response code is [%d], result is [%s], error is [%v]",
		url, headers, task, respCode, respData, err)
	if err != nil {
		logger.Errorf("Put the plan time of metric model task by taskid '%s' failed, %s", taskId, err)

		return err
	}

	if respCode != http.StatusNoContent {
		// 转成 baseerror
		var baseError rest.BaseError
		if err := json.Unmarshal(respData, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)

			return err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode, BaseError: baseError}
		logger.Errorf("put metric task return error: %v", httpErr.Error())

		return fmt.Errorf("put metric task %s return error %v", taskId, httpErr.Error())
	}

	return nil
}
