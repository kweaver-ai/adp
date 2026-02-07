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
	emAccessOnce sync.Once
	emAccess     interfaces.EventModelAccess
)

type eventModelAccess struct {
	eventTaskUrl  string
	eventModelUrl string
	httpClient    rest.HTTPClient
}

func NewEventModelAccess(appSetting *common.AppSetting) interfaces.EventModelAccess {
	opts := rest.HttpClientOptions{
		TimeOut: 3600,
	}
	emAccessOnce.Do(func() {
		emAccess = &eventModelAccess{
			eventTaskUrl:  appSetting.EventTaskUrl,
			eventModelUrl: appSetting.EventModelUrl,
			httpClient:    common.NewHTTPClientWithOptions(opts),
		}
	})

	return emAccess
}

// 更新任务执行状态
func (ema *eventModelAccess) UpdateEventTaskAttributesById(ctx context.Context, task interfaces.EventTask) error {
	url := fmt.Sprintf("%s/%s/attr", ema.eventTaskUrl, task.TaskID)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	respCode, respData, err := ema.httpClient.PutNoUnmarshal(ctx, url, headers, task)
	logger.Debugf("put [%s] with headers[%v], task[%v] finished, response code is [%d], result is [%s], error is [%v]",
		url, headers, task, respCode, respData, err)
	if err != nil {
		logger.Errorf("Put the task_status of event model task by taskid '%s' failed, %s", task.TaskID, err)
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
		logger.Errorf("Put event task return error: %v", httpErr.Error())

		return fmt.Errorf("put event task %s return error %v", task.TaskID, httpErr.Error())
	}

	return nil
}

// 获取事件模型详情
func (ema *eventModelAccess) GetEventModel(ctx context.Context, modelID string) (interfaces.EventModel, error) {
	url := fmt.Sprintf("%s/%s", ema.eventModelUrl, modelID)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}
	respCode, respData, err := ema.httpClient.GetNoUnmarshal(ctx, url, nil, headers)
	logger.Debugf("get [%s] with headers[%v] finished, response code is [%d], result is [%s], error is [%v]",
		url, headers, respCode, respData, err)
	if err != nil {
		logger.Errorf("Get eventmodel by modelID '%d' failed, %s", modelID, err)
		return interfaces.EventModel{}, err
	}

	if respCode != http.StatusOK {
		// 转成 baseerror
		var baseError rest.BaseError
		if err := json.Unmarshal(respData, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)

			return interfaces.EventModel{}, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode, BaseError: baseError}
		logger.Errorf("Get eventmodel by modelID return error: %v", httpErr.Error())

		return interfaces.EventModel{}, fmt.Errorf("get eventmodel by modelID %s return error %v", modelID, httpErr.Error())
	}
	var eventmodels []interfaces.EventModel
	if err := json.Unmarshal(respData, &eventmodels); err != nil {
		logger.Errorf("unmalshal respData failed: %v\n", err)

		return interfaces.EventModel{}, err
	}
	if len(eventmodels) > 0 {
		return eventmodels[0], nil
	} else {
		return interfaces.EventModel{}, fmt.Errorf("EventModel %s not found", modelID)
	}
}
