// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	cond "data-model-job/common/condition"
)

// 视图对象信息
type DataView struct {
	ViewId     string                 `json:"id"`
	DataSource map[string]any         `json:"data_source"`
	FieldScope uint8                  `json:"field_scope"`
	Fields     []*cond.Field          `json:"fields"`
	Condition  *cond.CondCfg          `json:"filters"`
	FieldsMap  map[string]*cond.Field `json:"-"`
	Creator    AccountInfo            `json:"creator"`
}

//go:generate mockgen -source ../interfaces/data_view_service.go -destination ../interfaces/mock/mock_data_view_service.go
type DataViewService interface {
	GetIndexBases(ctx context.Context, view *DataView) ([]IndexBase, error)
}
