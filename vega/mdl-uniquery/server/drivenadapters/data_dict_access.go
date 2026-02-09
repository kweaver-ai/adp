// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"uniquery/common"
	"uniquery/interfaces"
	"uniquery/logics/data_dict"
)

var (
	ddAccessOnce sync.Once
	ddAccess     interfaces.DataDictAccess
)

type dataDictAccess struct {
	appSetting  *common.AppSetting
	dataDictUrl string
	httpClient  rest.HTTPClient
}

func NewDataDictAccess(appSetting *common.AppSetting) interfaces.DataDictAccess {
	ddAccessOnce.Do(func() {
		ddAccess = &dataDictAccess{
			appSetting:  appSetting,
			dataDictUrl: appSetting.DataDictUrl,
			httpClient:  common.NewHTTPClient(),
		}
	})
	return ddAccess
}

// 获取数据字典id
func (dda *dataDictAccess) GetDictInfo(ctx context.Context, dictName string) (interfaces.DataDict, error) {
	httpUrl := fmt.Sprintf("%s?offset=0&limit=-1&name=%s", dda.dataDictUrl, dictName)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		"X-Language":                        rest.GetLanguageByCtx(ctx),
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	// httpClient 的请求新增参数支持上下文的处理请求的函数
	respCode, result, err := dda.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, headers)

	if err != nil {
		logger.Errorf("get request method failed: %v", err)
		return interfaces.DataDict{}, fmt.Errorf("get request method failed: %v", err)
	}
	if respCode == http.StatusNotFound {
		logger.Errorf("data dict %s not exists", dictName)
		return interfaces.DataDict{}, nil
	}
	if respCode != http.StatusOK {
		logger.Errorf("get data dict failed: %v", result)
		var baseError rest.BaseError
		if err := sonic.Unmarshal(result, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)

			return interfaces.DataDict{}, err
		}
		return interfaces.DataDict{}, fmt.Errorf("get data dict failed: %v", baseError.ErrorDetails)
	}
	if result == nil {
		return interfaces.DataDict{}, nil
	}
	list := struct {
		Entries []struct {
			DictID     string               `json:"id"`
			DictName   string               `json:"name"`
			UniqueKey  bool                 `json:"unique_key"`
			CreateTime int64                `json:"create_time"`
			UpdateTime int64                `json:"update_time"`
			Dimension  interfaces.Dimension `json:"dimension"`
			DictItems  []map[string]string  `json:"items"`
		} `json:"entries"`
	}{}

	if err := sonic.Unmarshal([]byte(result), &list); err != nil {
		logger.Errorf("unmalshal data dict info failed: %v\n", err)
		return interfaces.DataDict{}, err
	}

	if len(list.Entries) == 0 {
		return interfaces.DataDict{}, fmt.Errorf("dict does not exist")
	}

	return interfaces.DataDict{
		DictID:     list.Entries[0].DictID,
		DictName:   list.Entries[0].DictName,
		UniqueKey:  list.Entries[0].UniqueKey,
		CreateTime: list.Entries[0].CreateTime,
		UpdateTime: list.Entries[0].UpdateTime,
		Dimension:  list.Entries[0].Dimension,
	}, nil
}

// 获取数据字典信息
func (dda *dataDictAccess) GetDictIteams(ctx context.Context, dictID string) (map[string][]map[string]string, error) {
	httpUrl := fmt.Sprintf("%s/%s", dda.dataDictUrl, dictID)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		"X-Language":                        rest.GetLanguageByCtx(ctx),
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}
	// httpClient 的请求新增参数支持上下文的处理请求的函数
	respCode, result, err := dda.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, headers)

	if err != nil {
		logger.Errorf("get request method failed: %v", err)
		return nil, fmt.Errorf("get request method failed: %v", err)
	}
	if respCode == http.StatusNotFound {
		logger.Errorf("data dict %s not exists", dictID)
		return nil, nil
	}
	if respCode != http.StatusOK {
		logger.Errorf("get data dict failed: %v", result)
		var baseError rest.BaseError
		if err := sonic.Unmarshal(result, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)

			return nil, err
		}
		return nil, fmt.Errorf("get data dict failed: %v", baseError.ErrorDetails)
	}
	if result == nil {
		return nil, nil
	}

	// 处理返回结果 result
	var dicts []interfaces.DataDict
	if err := sonic.Unmarshal([]byte(result), &dicts); err != nil {
		logger.Errorf("unmalshal data dict info failed: %v\n", err)
		return nil, err
	}
	if len(dicts) == 0 {
		logger.Errorf("get data dict failed")
		return nil, err
	}

	data := make(map[string][]map[string]string, 0)

	for _, item := range dicts[0].DictItems {
		keyArr := make([]string, 0)
		for _, key := range dicts[0].Dimension.Keys {
			keyArr = append(keyArr, item[key.Name])
		}

		keys := strings.Join(keyArr, data_dict.Separator)

		if iteams, ok := data[keys]; ok {
			data[keys] = append(iteams, item)
		} else {
			data[keys] = []map[string]string{item}
		}

	}

	return data, nil
}
