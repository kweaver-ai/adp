// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	cond "uniquery/common/condition"
)

type VegaViewWithFields struct {
	Catalog      string          `json:"view_source_catalog_name"` //视图源
	Table        string          `json:"technical_name"`           //表技术名称
	ViewName     string          `json:"business_name"`            //视图名称
	Fields       []VegaViewField `json:"fields"`
	VegaFieldMap map[string]VegaViewField
}

type VegaViewField struct {
	Name string `json:"technical_name"`
	Type string `json:"data_type"`
}

type VegaFetchData struct {
	Data    [][]any          `json:"data"` // data里是数组的数组,每个数组里面是一行数据,每行数据的位置与coumns里的定义的字段顺序一致
	Columns []cond.ViewField `json:"columns"`
	NextUri string           `json:"nextUri"`
}

//go:generate mockgen -source ../interfaces/vega_service.go -destination ../interfaces/mock/mock_vega_service.go
type VegaService interface {
	GetVegaViewFieldsByID(ctx context.Context, viewID string) (VegaViewWithFields, error)
	FetchDatasFromVega(ctx context.Context, sql string) (VegaFetchData, error)
}
