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

	"uniquery/common"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
)

var (
	emAccessOnce sync.Once
	emAccess     interfaces.EventModelAccess
)

type eventModelAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewEventModelAccess(appSetting *common.AppSetting) interfaces.EventModelAccess {
	emAccessOnce.Do(func() {
		emAccess = &eventModelAccess{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
		}
	})
	return emAccess
}

// 根据 id 获取事件模型列表
func (ema *eventModelAccess) GetEventModelById(ctx context.Context, event_model_id string) ([]interfaces.EventModel, error) {
	httpUrl := ema.appSetting.EventModelUrl + "/" + event_model_id

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	respCode, respData, err := ema.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, headers)
	if err != nil {
		errDetails := fmt.Sprintf("GetEventModelById http request failed: %s", err.Error())
		logger.Error(errDetails)
		return nil, fmt.Errorf("get request method failed: %s", err)
	}

	if respCode == http.StatusNotFound {
		logger.Errorf("Failed to found event_model with input parameter")
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_EventModel_EventModelNotFound).
			WithErrorDetails(fmt.Sprintf("event model %s not found", event_model_id))
	}

	if respCode != http.StatusOK {
		logger.Errorf("get event model failed: %v", err)

		var baseError rest.BaseError
		if err := json.Unmarshal(respData, &baseError); err != nil {
			logger.Errorf("Unmalshal baesError failed: %s", err)
			return nil, err
		}

		return nil, fmt.Errorf("GetEventModelById failed: %s", baseError.ErrorDetails)
	}

	var ems []interfaces.EventModel
	if err = json.Unmarshal(respData, &ems); err != nil {
		logger.Errorf("Unmarshal event model failed: %s", err)
		return nil, err
	}

	return ems, nil
}

type EventModelRecords struct {
	Entries []interfaces.EventModel
	Total   int
}

func IsContain(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

func (ema *eventModelAccess) GetEventModelBySourceId(ctx context.Context, dataSource string) ([]interfaces.EventModel, error) {
	// httpUrl := ema.appSetting.EventModelUrl + "?status=1&enable_subscribe=1"
	httpUrl := fmt.Sprintf("%s?status=%s&enable_subscribe=%s&limit=-1", ema.appSetting.EventModelUrl, "1", "1")

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}
	respCode, respData, err := ema.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, headers)
	if err != nil {
		errDetails := fmt.Sprintf("GetEventModelById http request failed: %s", err.Error())
		logger.Error(errDetails)
		return nil, fmt.Errorf("get request method failed: %s", err)
	}

	if respCode == http.StatusNotFound {
		logger.Errorf("Failed to found event_model with input parameter")
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_EventModel_EventModelNotFound).
			WithErrorDetails("event models not found")
	}

	if respCode != http.StatusOK {
		logger.Errorf("get event model failed: %v", err)

		var baseError rest.BaseError
		if err := json.Unmarshal(respData, &baseError); err != nil {
			logger.Errorf("Unmalshal baesError failed: %s", err)
			return nil, err
		}

		return nil, fmt.Errorf("GetEventModelById failed: %s", baseError.ErrorDetails)
	}

	var ems = EventModelRecords{}

	if err = json.Unmarshal(respData, &ems); err != nil {
		logger.Errorf("Unmarshal event model failed: %s", err)
		return nil, err
	}
	var relationsEms []interfaces.EventModel
	for _, em := range ems.Entries {
		if IsContain(em.DataSource, dataSource) {
			logger.Debugf("Found event_model %s", em.EventModelName)
			relationsEms = append(relationsEms, em)
		}
	}
	logger.Infof("Found event_model total to listen cnt :%d", len(relationsEms))

	return relationsEms, nil
}
