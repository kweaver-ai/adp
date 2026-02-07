// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

const (
	//模块类型
	MODULE_TYPE_OBJECTIVE_MODEL = "objective_model"

	OBJECTTYPE_OBJECTIVE_MODEL = "ID_AUDIT_OBJECTIVE_MODEL"

	// objective type
	SLO = "slo"
	KPI = "kpi"

	// 目标单位
	UNIT_NUM_NONE    = "none"
	UNIT_NUM_PERCENT = "%"

	// 综合计算指标最大个数
	COMPREHENSIVE_METRIC_TOTAL = 10
	// 附加计算指标最大个数
	ADDITIONAL_METRIC_TOTAL = 5
	// 状态个数
	STATUS_TOTAL = 10
)

var (
	OBJECTIVE_MODEL_SORT = map[string]string{
		"update_time": "f_update_time",
		"model_name":  "f_model_name",
	}
)

// 目标模型结构体
type ObjectiveModelInfo struct {
	ModelID         string      `json:"id"`
	ModelName       string      `json:"name"`
	ObjectiveType   string      `json:"objective_type"`
	ObjectiveConfig any         `json:"objective_config"`
	Tags            []string    `json:"tags"`
	Comment         string      `json:"comment"`
	Creator         AccountInfo `json:"creator"`
	CreateTime      int64       `json:"create_time"`
	UpdateTime      int64       `json:"update_time"`

	// 操作权限
	Operations   []string `json:"operations"`
	IfNameModify bool     `json:"-"`
}

type CreateObjectiveModel struct {
	ObjectiveModelInfo
	Task *CreateMetricTask `json:"task"`
}

type ObjectiveModel struct {
	ObjectiveModelInfo
	Task *MetricTask `json:"task"`
}

type SLOObjective struct {
	Objective        *float64               `json:"objective"`
	Period           *int64                 `json:"period"`
	GoodMetricModel  *BundleMetricModel     `json:"good_metric_model"`
	TotalMetricModel *BundleMetricModel     `json:"total_metric_model"`
	StatusConfig     *ObjectiveStatusConfig `json:"status_config"`
}

type KPIObjective struct {
	Objective                 *float64                   `json:"objective"`
	Unit                      string                     `json:"unit"`
	ComprehensiveMetricModels []ComprehensiveMetricModel `json:"comprehensive_metric_models"`
	AdditionalMetricModels    []BundleMetricModel        `json:"additional_metric_models"`
	ScoreMax                  *float64                   `json:"score_max,omitempty"`
	ScoreMin                  *float64                   `json:"score_min,omitempty"`
	StatusConfig              *ObjectiveStatusConfig     `json:"status_config"`
}

type BundleMetricModel struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	UnitType string `json:"unit_type,omitempty"`
	Unit     string `json:"unit,omitempty"`
}

type ComprehensiveMetricModel struct {
	ID     string `json:"id"`
	Weight *int64 `json:"weight"`
	Name   string `json:"name"`
}

type ObjectiveStatusConfig struct {
	Ranges      []Range `json:"ranges"`
	OtherStatus string  `json:"other_status"`
}

type Range struct {
	From   *float64 `json:"from"`
	To     *float64 `json:"to"`
	Status string   `json:"status"`
}

// 目标模型列表查询参数
type ObjectiveModelsQueryParams struct {
	PaginationQueryParameters
	NamePattern   string
	Name          string
	ObjectiveType string
	Tag           string
}

func IsValidObjectiveType(m string) bool {
	return m == SLO || m == KPI
}
