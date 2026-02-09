// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package vega_view

import (
	"context"
	"strings"
	"sync"

	"uniquery/common"
	"uniquery/interfaces"
	"uniquery/logics"
)

var (
	vvServiceOnce sync.Once
	vvService     interfaces.VegaService
)

type vegaService struct {
	appSetting *common.AppSetting
	vva        interfaces.VegaAccess
}

func NewVegaService(appSetting *common.AppSetting) interfaces.VegaService {
	vvServiceOnce.Do(func() {
		vvService = &vegaService{
			appSetting: appSetting,
			vva:        logics.VVA,
		}
	})
	return vvService
}

func (vvs *vegaService) GetVegaViewFieldsByID(ctx context.Context, viewID string) (interfaces.VegaViewWithFields, error) {
	vegaViewFields, err := vvs.vva.GetVegaViewFieldsByID(ctx, viewID)
	if err != nil {
		return vegaViewFields, err
	}
	// 把fields数组处理一份fields的map，key为field id，便于后面处理校验字段存在性
	fieldMap := make(map[string]interfaces.VegaViewField)
	for _, field := range vegaViewFields.Fields {
		fieldMap[field.Name] = field
	}
	vegaViewFields.VegaFieldMap = fieldMap
	return vegaViewFields, nil
}

func (vvs *vegaService) FetchDatasFromVega(ctx context.Context, sql string) (interfaces.VegaFetchData, error) {
	var fetchData interfaces.VegaFetchData
	fetchDatai, err := vvs.vva.FetchDatasFromVega(ctx, "", sql)
	if err != nil {
		return fetchData, err
	}
	fetchData = fetchDatai

	for fetchDatai.NextUri != "" {
		// 用nexturi继续请求下一批数据
		// 从uri中提取出后一段 v1之后的部分
		// http://vega-gateway:8099/api/virtual_engine_service/v1/statement/executing/20250516_051107_00077_dpx8g/xe0786066b47e4aeda3972d483db644f1/1
		dataUri := strings.SplitAfter(fetchDatai.NextUri, "executing/")[1]
		// nextUri = strings.Split(dataUri, "/")
		fetchDatai, err = vvs.vva.FetchDatasFromVega(ctx, dataUri, sql)
		if err != nil {
			return fetchData, err
		}
		fetchData.Data = append(fetchData.Data, fetchDatai.Data...)
	}
	return fetchData, nil
}
