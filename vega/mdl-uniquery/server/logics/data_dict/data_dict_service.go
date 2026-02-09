// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_dict

import (
	"context"
	"strings"
	"sync"

	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"uniquery/common"
	"uniquery/interfaces"
	"uniquery/logics"
)

var (
	// 使用0值字节作为分隔符
	Separator = "\x00"

	mu      sync.RWMutex
	ddsOnce sync.Once
	dds     interfaces.DataDictService

	dictCache map[string]interfaces.DataDict = make(map[string]interfaces.DataDict)
)

type DataDictService struct {
	appSetting *common.AppSetting
	ddAccess   interfaces.DataDictAccess
}

func NewDataDictService(appSetting *common.AppSetting) interfaces.DataDictService {
	ddsOnce.Do(func() {
		dds = &DataDictService{
			appSetting: appSetting,
			ddAccess:   logics.DDAccess,
		}
	})
	return dds
}

func (dds *DataDictService) LoadDict(ctx context.Context, name string) error {

	dict, err := dds.ddAccess.GetDictInfo(ctx, name)
	if err != nil {
		// 如果该字典名称没有找到字典，并且缓存中有该字典信息，把缓存的字典信息删除
		mu.Lock()
		defer mu.Unlock()
		delete(dictCache, name)

		logger.Errorf("failed to find dict: %s ,beacuse %s", name, err.Error())
		return err
	}

	// 判断该字典是否在缓存中存在
	mu.RLock()
	defer mu.RUnlock()
	cachedDict, ok := dictCache[name]

	// 字典不存在、缓存不一致，更新缓存
	if !ok || dict.UpdateTime != cachedDict.UpdateTime {
		go dds.UpdateCache(context.WithoutCancel(ctx), dict)
	}

	return nil
}

func (dds *DataDictService) UpdateCache(ctx context.Context, dict interfaces.DataDict) {
	mu.Lock()
	defer mu.Unlock()
	iteams, err := dds.ddAccess.GetDictIteams(ctx, dict.DictID)
	if err != nil {
		logger.Errorf("get dict ieatms falied")
		return
	}
	dict.DictRecords = iteams
	dictCache[dict.DictName] = dict
}

// 读缓存字典信息
func GetDictByName(dictName string) (interfaces.DataDict, bool) {
	mu.RLock()
	defer mu.RUnlock()
	if cachedDict, ok := dictCache[dictName]; ok {
		return cachedDict, true
	}

	return interfaces.DataDict{}, false
}

// 通过字典的所有key获取字典项
func GetRecordsByKey(dict interfaces.DataDict, values []string) ([]map[string]string, bool) {

	uniqueString := strings.Join(values, Separator)
	records, ok := dict.DictRecords[uniqueString]
	return records, ok
}
