// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"github.com/kweaver-ai/kweaver-go-lib/audit"
)

const (
	OBJECT_TYPE_DATA_CONNECTION           = "data_connection"
	OBJECT_TYPE_DATA_DICT                 = "data_dict"
	OBJECT_TYPE_DATA_DICT_ITEM            = "data_dict_item"
	OBJECT_TYPE_DATA_VIEW                 = "data_view"
	OBJECT_TYPE_DATA_VIEW_GROUP           = "data_view_group"
	OBJECT_TYPE_DATA_VIEW_ROW_COLUMN_RULE = "data_view_row_column_rule"
	OBJECT_TYPE_EVENT_MODEL               = "event_model"
	OBJECT_TYPE_METRIC_MODEL              = "metric_model"
	OBJECT_TYPE_METRIC_MODEL_GROUP        = "metric_model_group"
	OBJECT_TYPE_OBJECTIVE_MODEL           = "objective_model"
	OBJECT_TYPE_TRACE_MODEL               = "trace_model"
)

func GenerateDataConnectionAuditObject(id string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: OBJECT_TYPE_DATA_CONNECTION,
		ID:   id,
		Name: name,
	}
}

func GenerateDataDictAuditObject(id string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: OBJECT_TYPE_DATA_DICT,
		ID:   id,
		Name: name,
	}
}

func GenerateDataDictItemAuditObject(id string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: OBJECT_TYPE_DATA_DICT_ITEM,
		ID:   id,
		Name: name,
	}
}

func GenerateDataViewAuditObject(id string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: OBJECT_TYPE_DATA_VIEW,
		ID:   id,
		Name: name,
	}
}

func GenerateDataViewGroupAuditObject(id string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: OBJECT_TYPE_DATA_VIEW_GROUP,
		ID:   id,
		Name: name,
	}
}

func GenerateDataViewRowColumnRuleAuditObject(id string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: OBJECT_TYPE_DATA_VIEW_ROW_COLUMN_RULE,
		ID:   id,
		Name: name,
	}
}

func GenerateEventModelAuditObject(id string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: OBJECT_TYPE_EVENT_MODEL,
		ID:   id,
		Name: name,
	}
}

func GenerateMetricModelAuditObject(id string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: OBJECT_TYPE_METRIC_MODEL,
		ID:   id,
		Name: name,
	}
}

func GenerateMetricModelGroupAuditObject(id string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: OBJECT_TYPE_METRIC_MODEL_GROUP,
		ID:   id,
		Name: name,
	}
}

func GenerateObjectiveModelAuditObject(id string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: OBJECT_TYPE_OBJECTIVE_MODEL,
		ID:   id,
		Name: name,
	}
}

func GenerateTraceModelAuditObject(id string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: OBJECT_TYPE_TRACE_MODEL,
		ID:   id,
		Name: name,
	}
}
