// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"uniquery/common"
	"uniquery/interfaces"
)

const STATUS_LOG_GROUP_NOT_FOUND = 3759341573

var (
	lgAccessOnce sync.Once
	lgAccess     interfaces.LogGroupAccess
)

type logGroupAccess struct {
	appSetting     *common.AppSetting
	dataManagerUrl string
	httpClient     rest.HTTPClient
}

func NewLogGroupAccess(appSetting *common.AppSetting) interfaces.LogGroupAccess {
	lgAccessOnce.Do(func() {
		lgAccess = &logGroupAccess{
			appSetting:     appSetting,
			dataManagerUrl: appSetting.DataManagerUrl,
			httpClient:     common.NewHTTPClient(),
		}
	})
	return lgAccess
}

// 根据日志分组id获取分组下的日志类型列表
func (lga *logGroupAccess) GetLogGroupQueryFilters(logGroupID string) (interfaces.LogGroup, bool, error) {
	httpUrl := fmt.Sprintf("%s/manager/loggroup/%s/queryfilters", lga.dataManagerUrl, logGroupID)
	respCode, result, err := lga.httpClient.Get(context.Background(), httpUrl, url.Values{}, make(map[string]string))

	if err != nil {
		logger.Errorf("get request method failed: %v", err)
		return interfaces.LogGroup{}, false, fmt.Errorf("get request method failed: %v", err)
	}

	if respCode != http.StatusOK {
		logger.Errorf("get queryfilters failed: %v", result)

		// Python的202状态码表示日志分组不存在
		if respCode == http.StatusAccepted {
			return interfaces.LogGroup{}, false, nil
		}
		return interfaces.LogGroup{}, false, fmt.Errorf("get queryfilters failed: %v", result)
	}

	indexPattern := make([]string, 0)
	userRes := result.(map[string]interface{})["indices"].(map[string]interface{})

	for _, v := range userRes["index_pattern"].([]interface{}) {
		// 日志类型的需要拼接上 -*
		indexPattern = append(indexPattern, v.(string)+"-*", "mdl-"+v.(string)+"-*")
	}
	for _, v := range userRes["manual_index"].([]interface{}) {
		// 手动索引直接查
		indexPattern = append(indexPattern, v.(string))
	}

	mustFilter := result.(map[string]interface{})["must_filter"]

	return interfaces.LogGroup{
		IndexPattern: indexPattern,
		MustFilter:   mustFilter,
	}, true, nil
}

// 根据名称获取日志分组信息
func (lga *logGroupAccess) GetLogGroupByName(logGroupName string) ([]interfaces.LogGroupInfo, error) {
	// http://localhost/manager/loggroup?name_pattern=^主机监控$
	// http://localhost/manager/loggroup?name_pattern=^%E4%B8%BB%E6%9C%BA%E7%9B%91%E6%8E%A7$
	httpUrl := fmt.Sprintf("%s/manager/loggroup?name_pattern=%s", lga.dataManagerUrl, fmt.Sprintf("^%s$", url.QueryEscape(logGroupName)))
	respCode, result, err := lga.httpClient.GetNoUnmarshal(context.Background(), httpUrl, nil, interfaces.Headers)

	if err != nil {
		logger.Errorf("get request method failed: %v", err)
		return []interfaces.LogGroupInfo{}, fmt.Errorf("get request method failed: %v", err)
	}
	if respCode != http.StatusOK {
		logger.Errorf("get log group by name failed: %v", result)
		return []interfaces.LogGroupInfo{}, fmt.Errorf("get log group by name failed: %v", err)
	}

	if result == nil {
		return []interfaces.LogGroupInfo{}, err
	}

	// 处理返回结果 result
	logGroupInfos := make([]interfaces.LogGroupInfo, 0)
	if err := sonic.Unmarshal(result, &logGroupInfos); err != nil {
		logger.Errorf("unmalshal log group info failed: %v\n", err)
		return []interfaces.LogGroupInfo{}, err
	}

	return logGroupInfos, nil
}

// 根据id获取数据视图的日志类型列表和过滤条件
func (lga *logGroupAccess) GetLogGroupQueryFiltersAndFields(groupId string) (interfaces.LogGroup, bool, error) {
	httpUrl := fmt.Sprintf("%s/manager/loggroup/%s?include_fields=1&desensitize=1", lga.dataManagerUrl, groupId)
	respCode, result, err := lga.httpClient.GetNoUnmarshal(context.Background(), httpUrl, nil, interfaces.Headers)
	logger.Debugf("get [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, result, err)

	if err != nil {
		logger.Errorf("get request method failed: %v", err)
		return interfaces.LogGroup{}, false, fmt.Errorf("get request method failed: %v", err)
	}

	// 日志分组不存在
	if respCode == http.StatusAccepted {
		logger.Errorf("log group %s not exists", groupId)
		return interfaces.LogGroup{}, false, nil
	}

	if respCode != http.StatusOK {
		logger.Errorf("get queryfilters failed: %v", result)
		return interfaces.LogGroup{}, false, fmt.Errorf("get queryfilters failed: %v", err)
	}

	if result == nil {
		return interfaces.LogGroup{}, false, err
	}

	// 处理返回结果 result
	var queryFilters interfaces.QueryFilters
	if err := sonic.Unmarshal(result, &queryFilters); err != nil {
		logger.Errorf("unmalshal log group info failed: %v\n", err)
		return interfaces.LogGroup{}, false, err
	}

	indexPattern := make([]string, 0)
	for _, v := range queryFilters.Indices.IndexPattern {
		// 日志类型的需要拼接上 -*
		indexPattern = append(indexPattern, v+"-*", "mdl-"+v+"-*")
	}
	// 手动索引直接查
	indexPattern = append(indexPattern, queryFilters.Indices.ManualIndex...)
	fields := make(map[string]string, 0)
	err = getFields(queryFilters.ArrayFields, fields)
	if err != nil {
		logger.Errorf("loggroup field type is wrong")
		return interfaces.LogGroup{}, false, err
	}
	queryFilters.Fields = fields
	return interfaces.LogGroup{
		IndexPattern: indexPattern,
		MustFilter:   queryFilters.MustFilter,
		Fields:       queryFilters.Fields,
	}, true, nil
}

// 从日志分组中把字段信息解析出来
func getFields(fieldsInfo map[string][]interfaces.LogGroupField, res map[string]string) error {
	for _, fields := range fieldsInfo {
		err := getValue("", fields, res)
		if err != nil {
			return err
		}
	}
	return nil
}

// 构造单个日志类型的字段map
func getValue(key string, lists []interfaces.LogGroupField, res map[string]string) error {
	for i := 0; i < len(lists); i++ {
		prefix := ""
		if key != "" {
			prefix = key + "."
		}
		switch v := lists[i].Type.(type) {
		case string:
			// 以第一个为准。需判断是否已存在,不存在则添加
			if _, ok := res[prefix+lists[i].Name]; !ok {
				res[prefix+lists[i].Name] = lists[i].Type.(string)
			}
		case []interface{}:
			// convert map to json
			jsonString, err := sonic.Marshal(v)
			if err != nil {
				return err
			}
			// convert json to struct
			fieldArr := make([]interfaces.LogGroupField, 0)
			err = sonic.Unmarshal(jsonString, &fieldArr)
			if err != nil {
				return err
			}

			err = getValue(prefix+lists[i].Name, fieldArr, res)
			if err != nil {
				return err
			}
		default:
			return errors.New("loggroup field type is wrong")
		}
	}
	return nil
}
