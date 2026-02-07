// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"context"
	"fmt"
	"sync"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/mitchellh/mapstructure"

	"data-model-job/common"
	"data-model-job/interfaces"
)

var (
	dvsOnce sync.Once
	dvs     interfaces.DataViewService
)

type dataViewService struct {
	appSetting *common.AppSetting
	ibAccess   interfaces.IndexBaseAccess
}

func NewDataViewService(appSetting *common.AppSetting) interfaces.DataViewService {
	dvsOnce.Do(func() {
		dvs = &dataViewService{
			appSetting: appSetting,
			ibAccess:   IBAccess,
		}
	})

	return dvs
}

// 获取索引库信息
func (dvService *dataViewService) GetIndexBases(ctx context.Context, view *interfaces.DataView) ([]interfaces.IndexBase, error) {
	switch view.DataSource["type"].(string) {
	case interfaces.INDEX_BASE:
		var bases []interfaces.SimpleIndexBase
		err := mapstructure.Decode(view.DataSource[interfaces.INDEX_BASE], &bases)
		if err != nil {
			return nil, fmt.Errorf("mapstructure decode dataSource failed, %v", err)
		}

		baseTypes := make([]string, 0, len(bases))
		for _, base := range bases {
			baseTypes = append(baseTypes, base.BaseType)
		}

		// 根据索引库类型获取索引库信息
		baseInfos, err := dvService.ibAccess.GetIndexBasesByTypes(ctx, baseTypes)
		if err != nil {
			logger.Errorf("Get index bases failed, %s", err.Error())
			return nil, fmt.Errorf("get index bases failed, %v", err)
		}

		return baseInfos, nil

	default:
		return nil, fmt.Errorf("unsupported dataSource type, only 'index_base' is supported currently")
	}
}

// 合并索引库字段
func mergeIndexBaseFields(mappings interfaces.Mappings) []interfaces.IndexBaseField {
	capacity := len(mappings.DynamicMappings) + len(mappings.MetaMappings) + len(mappings.UserDefinedMappings)
	allBaseFields := make([]interfaces.IndexBaseField, 0, capacity)

	allBaseFields = append(allBaseFields, mappings.MetaMappings...)
	allBaseFields = append(allBaseFields, mappings.DynamicMappings...)
	allBaseFields = append(allBaseFields, mappings.UserDefinedMappings...)

	return allBaseFields
}
