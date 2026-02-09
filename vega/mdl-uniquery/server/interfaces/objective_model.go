// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

const (
	// objective type
	SLO string = "slo"
	KPI string = "kpi"

	STATUS_DEFAULT      = "__未定义"
	STATUS_CODE_DEFAULT = -1 // 未定义时的状态码
	STATUS_CODE_OTHER   = 0  // 定义了其他的状态码

	// 目标单位
	UNIT_NUM_NONE    = "none"
	UNIT_NUM_PERCENT = "%"

	// 综合计算指标最大个数
	COMPREHENSIVE_METRIC_TOTAL = 10
	// 附加计算指标最大个数
	ADDITIONAL_METRIC_TOTAL = 5
	// 状态个数
	STATUS_TOTAL = 10

	// slo的目标最大值
	SLO_MAX_OBJECTIVE = float64(100)
)

// 指标数据预览的查询请求体
type ObjectiveModelQuery struct {
	ObjectiveType   string `json:"objective_type"`
	ObjectiveConfig any    `json:"objective_config"`
	QueryTimeParams
	ObjectiveModelID string `json:"-"`
	ObjectiveModelQueryParameters
}

type ObjectiveModelQueryParameters struct {
	MetricModelQueryParameters
	IncludeMetrics []string
}

// 目标模型的数据返回结构体
type ObjectiveModelUniResponse struct {
	Model               any   `json:"model,omitempty"`
	Datas               []any `json:"datas"`
	PointTotalPerSeries int   `json:"points_total_per_series,omitempty"`
	SeriesTotal         int   `json:"series_total,omitempty"`
}

type SLOObjectiveData struct {
	Labels           map[string]string `json:"labels"`
	Times            []any             `json:"times,omitempty"`
	SLI              []any             `json:"__slis,omitempty"`
	Objective        []float64         `json:"__slo_objectives,omitempty"`
	Good             []any             `json:"__goods,omitempty"`
	Total            []any             `json:"__totals,omitempty"`
	AchiveRate       []any             `json:"__slo_achievement_rates,omitempty"`
	TotalErrorBudget []any             `json:"__total_error_budgets,omitempty"`
	LeftErrorBudget  []any             `json:"__left_error_budgets,omitempty"`
	BurnRate         []any             `json:"__burn_rates,omitempty"`
	Period           []int64           `json:"__slo_periods,omitempty"`
	Status           []string          `json:"__slo_status,omitempty"`
	StatusCode       []int             `json:"__slo_status_code,omitempty"`
}

type KPIObjectiveData struct {
	Labels              map[string]string `json:"labels"`
	Times               []any             `json:"times"`
	KPI                 []any             `json:"__kpis"`
	Objective           []float64         `json:"__kpi_objectives,omitempty"`
	AchiveRate          []any             `json:"__kpi_achievement_rates,omitempty"`
	KPIScore            []any             `json:"__kpi_scores,omitempty"`
	AssociateMetricNums []any             `json:"__kpi_associate_metric_nums,omitempty"`
	Status              []string          `json:"__kpi_status,omitempty"`
	StatusCode          []int             `json:"__kpi_status_code,omitempty"`
}

type ObjectiveModel struct {
	ModelID         string      `json:"id"`
	ModelName       string      `json:"name"`
	ObjectiveType   string      `json:"objective_type"`
	ObjectiveConfig any         `json:"objective_config"`
	Tags            []string    `json:"tags"`
	Comment         string      `json:"comment"`
	UpdateTime      int64       `json:"update_time"`
	Task            *MetricTask `json:"task"`
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
	ScoreMax                  *float64                   `json:"score_max"`
	ScoreMin                  *float64                   `json:"score_min"`
	StatusConfig              *ObjectiveStatusConfig     `json:"status_config"`
}

type BundleMetricModel struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	UnitType string `json:"unit_type,omitempty"`
	Unit     string `json:"unit,omitempty"`
}

type ComprehensiveMetricModel struct {
	Id     string `json:"id"`
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

func IsValidObjectiveType(m string) bool {
	return m == SLO || m == KPI
}
