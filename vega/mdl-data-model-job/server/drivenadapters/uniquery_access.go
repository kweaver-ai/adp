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
	uAccessOnce sync.Once
	uAccess     interfaces.UniqueryAccess
)

type uniqueryAccess struct {
	uniqueryUrl string
	// insightEventUrl string
	httpClient rest.HTTPClient
}

func NewUniqueryAccess(appSetting *common.AppSetting) interfaces.UniqueryAccess {
	opts := rest.HttpClientOptions{
		TimeOut: 3600,
	}
	uAccessOnce.Do(func() {
		uAccess = &uniqueryAccess{
			uniqueryUrl: appSetting.UniQueryUrl,
			httpClient:  common.NewHTTPClientWithOptions(opts),
		}
	})

	return uAccess
}

func (ua *uniqueryAccess) GetMetricModelData(ctx context.Context, modelId string, query interfaces.MetricModelQuery) (interfaces.UniResponse, error) {

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	uniqueryMetricModelHeaders := map[string]string{
		interfaces.CONTENT_TYPE_NAME:           interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_METHOD_OVERRIDE: http.MethodGet,
		interfaces.HTTP_HEADER_ACCOUNT_ID:      accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE:    accountInfo.Type,
	}

	url := fmt.Sprintf("%s/metric-models/%s?ignoring_store_cache=true", ua.uniqueryUrl, modelId)

	respCode, result, err := ua.httpClient.PostNoUnmarshal(ctx, url, uniqueryMetricModelHeaders, query)
	logger.Debugf("post [%s] with headers[%v] finished, request is [%v] response code is [%d], result is [%s], error is [%v]",
		ua.uniqueryUrl, uniqueryMetricModelHeaders, query, respCode, result, err)

	metricData := interfaces.UniResponse{}

	if err != nil {
		logger.Errorf("get request method failed: %v", err)

		return metricData, fmt.Errorf("get request method failed: %v", err)
	}
	if respCode != http.StatusOK {
		// 转成 baseerror
		var baseError rest.BaseError
		if err := json.Unmarshal(result, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)
			return metricData, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode, BaseError: baseError}
		logger.Errorf("Formula invalid: %v", httpErr.Error())

		return metricData, fmt.Errorf("get metric data %s return error %v", modelId, httpErr.Error())
	}

	if result == nil {
		return metricData, fmt.Errorf("get metric data %v return null", modelId)
	}

	if err := json.Unmarshal(result, &metricData); err != nil {
		logger.Errorf("Unmarshal Metric Model Data failed, %s", err)

		return metricData, err
	}

	return metricData, nil
}

func (ua *uniqueryAccess) GetObjectiveModelData(ctx context.Context, modelId string, query interfaces.MetricModelQuery) (interfaces.ObjectiveModelUniResponse, error) {

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	uniqueryMetricModelHeaders := map[string]string{
		interfaces.CONTENT_TYPE_NAME:           interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_METHOD_OVERRIDE: http.MethodGet,
		interfaces.HTTP_HEADER_ACCOUNT_ID:      accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE:    accountInfo.Type,
	}

	url := fmt.Sprintf("%s/objective-models/%s?ignoring_store_cache=true&include_model=true", ua.uniqueryUrl, modelId)

	respCode, result, err := ua.httpClient.PostNoUnmarshal(ctx, url, uniqueryMetricModelHeaders, query)
	logger.Debugf("post [%s] finished, request is [%v] response code is [%d], result is [%s], error is [%v]", ua.uniqueryUrl,
		query, respCode, result, err)

	objectiveData := interfaces.ObjectiveModelUniResponse{}

	if err != nil {
		logger.Errorf("get request method failed: %v", err)

		return objectiveData, fmt.Errorf("get request method failed: %v", err)
	}
	if respCode != http.StatusOK {
		// 转成 baseerror
		var baseError rest.BaseError
		if err := json.Unmarshal(result, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)
			return objectiveData, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode, BaseError: baseError}
		logger.Errorf("Formula invalid: %v", httpErr.Error())

		return objectiveData, fmt.Errorf("get objective data %s return error %v", modelId, httpErr.Error())
	}

	if result == nil {
		return objectiveData, fmt.Errorf("get objective data %v return null", modelId)
	}

	if err := json.Unmarshal(result, &objectiveData); err != nil {
		logger.Errorf("Unmarshal Objective Model Data failed, %s", err)

		return objectiveData, err
	}

	return objectiveData, nil
}

// 查询事件模型数据
func (ua *uniqueryAccess) GetEventModelData(ctx context.Context, query interfaces.EventModelQueryRequest) (interfaces.EventModelResponse, error) {

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	uniqueryEventModelHeaders := map[string]string{
		interfaces.CONTENT_TYPE_NAME:           interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_METHOD_OVERRIDE: http.MethodGet,
		interfaces.HTTP_HEADER_ACCOUNT_ID:      accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE:    accountInfo.Type,
	}

	url := fmt.Sprintf("%s/events", ua.uniqueryUrl)

	respCode, result, err := ua.httpClient.PostNoUnmarshal(ctx, url, uniqueryEventModelHeaders, query)
	// logger.Debugf("post [%s] finished, response code is [%d], result is [%s], error is [%v]", url, respCode, result, err)
	eventData := interfaces.EventModelResponse{}

	if err != nil {
		logger.Errorf("get request method failed: %v", err)

		return eventData, fmt.Errorf("get request method failed: %v", err)
	}
	if respCode != http.StatusOK {
		// 转成 baseerror
		var baseError rest.BaseError
		if err := json.Unmarshal(result, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)
			return eventData, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode, BaseError: baseError}
		logger.Errorf("Formula invalid: %v", httpErr.Error())

		return eventData, fmt.Errorf("get event data %s return error %v", query.Querys[0].Id, httpErr.Error())
	}

	if result == nil {
		logger.Errorf("get event data  %v failed,return null ", query.Querys[0].Id)
		return eventData, fmt.Errorf("get event data %s return null", query.Querys[0].Id)
	}

	if err := json.Unmarshal(result, &eventData); err != nil {
		logger.Errorf("Unmarshal Event Model Data failed, %s", err)

		return eventData, err
	}

	return eventData, nil
}
