// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors Resource 模块错误码
package errors

// Resource 错误码
const (
	// LogicView 校验相关
	VegaBackend_LogicalView_InvalidParameter_LogicDefinition   = "VegaBackend.LogicalView.InvalidParameter.LogicDefinition"
	VegaBackend_LogicalView_InvalidParameter_FieldName         = "VegaBackend.LogicalView.InvalidParameter.FieldName"
	VegaBackend_LogicalView_LengthExceeded_FieldName           = "VegaBackend.LogicalView.LengthExceeded.FieldName"
	VegaBackend_LogicalView_LengthExceeded_FieldDisplayName    = "VegaBackend.LogicalView.LengthExceeded.FieldDisplayName"
	VegaBackend_LogicalView_LengthExceeded_FieldComment        = "VegaBackend.LogicalView.LengthExceeded.FieldComment"
	VegaBackend_LogicalView_Duplicated_FieldName               = "VegaBackend.LogicalView.Duplicated.FieldName"
	VegaBackend_LogicalView_Duplicated_FieldDisplayName        = "VegaBackend.LogicalView.Duplicated.FieldDisplayName"
	VegaBackend_LogicalView_InvalidParameter_FieldFeatureName  = "VegaBackend.LogicalView.InvalidParameter.FieldFeatureName"
	VegaBackend_LogicalView_LengthExceeded_FieldFeatureName    = "VegaBackend.LogicalView.LengthExceeded.FieldFeatureName"
	VegaBackend_LogicalView_Duplicated_FieldFeatureName        = "VegaBackend.LogicalView.Duplicated.FieldFeatureName"
	VegaBackend_LogicalView_LengthExceeded_FieldFeatureComment = "VegaBackend.LogicalView.LengthExceeded.FieldFeatureComment"
)

var LogicalViewResourceErrCodeList = []string{
	VegaBackend_LogicalView_InvalidParameter_LogicDefinition,
	VegaBackend_LogicalView_InvalidParameter_FieldName,
	VegaBackend_LogicalView_LengthExceeded_FieldName,
	VegaBackend_LogicalView_LengthExceeded_FieldDisplayName,
	VegaBackend_LogicalView_LengthExceeded_FieldComment,
	VegaBackend_LogicalView_Duplicated_FieldName,
	VegaBackend_LogicalView_Duplicated_FieldDisplayName,
	VegaBackend_LogicalView_InvalidParameter_FieldFeatureName,
	VegaBackend_LogicalView_LengthExceeded_FieldFeatureName,
	VegaBackend_LogicalView_Duplicated_FieldFeatureName,
	VegaBackend_LogicalView_LengthExceeded_FieldFeatureComment,
}
