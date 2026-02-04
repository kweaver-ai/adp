package interfaces

import (
	"context"
	cond "ontology-query/common/condition"
)

type ViewQuery struct {
	Filters        *cond.CondCfg `json:"filters"`
	NeedTotal      bool          `json:"need_total"`
	Limit          int           `json:"limit"`
	UseSearchAfter bool          `json:"use_search_after"`
	Sort           []*SortParams `json:"sort"`
	SearchAfterParams
}

type SortParams struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

type SearchAfterParams struct {
	SearchAfter []any `json:"search_after"`
	// PitID        string `json:"pit_id"`
	// PitKeepAlive string `json:"pit_keep_alive"`
}

type ViewData struct {
	Datas       []map[string]any `json:"entries"`
	TotalCount  int64            `json:"total_count"`
	SearchAfter []any            `json:"search_after,omitempty"`
}

type OrderField struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Direction string `json:"direction"` // asc or desc
}

type HavingCondition struct {
	Field     string `json:"field"`     // 只有 __value
	Operation string `json:"operation"` // ==, !=, >, >=, <, <=, in, not_in, range, out_range
	Value     any    `json:"value"`
}

type SameperiodConfig struct {
	Method          []string `json:"method"`           // growth_value or growth_rate
	Offset          int      `json:"offset"`           // 偏移量
	TimeGranularity string   `json:"time_granularity"` // day, month, quarter, year
}

type Metrics struct {
	Type             string            `json:"type"` // sameperiod or proportion
	SameperiodConfig *SameperiodConfig `json:"sameperiod_config,omitempty"`
}

type MetricQuery struct {
	Start              *int64           `json:"start"`
	End                *int64           `json:"end"`
	StepStr            *string          `json:"step"`
	IsInstantQuery     bool             `json:"instant"`
	Filters            []Filter         `json:"filters"`
	AnalysisDimensions []string         `json:"analysis_dimensions,omitempty"`
	OrderByFields      []OrderField     `json:"order_by_fields,omitempty"`
	HavingCondition    *HavingCondition `json:"having_condition,omitempty"`
	Metrics            *Metrics         `json:"metrics,omitempty"`
}

type MetricData struct {
	Model      MetricModel `json:"model,omitempty"`
	Datas      []Data      `json:"datas"`
	Step       string      `json:"step"`
	IsVariable bool        `json:"is_variable"`
	IsCalendar bool        `json:"is_calendar"`
}

type Data struct {
	Labels map[string]string `json:"labels"`
	Times  []interface{}     `json:"times"`
	// TimeStrs     []interface{}     `json:"time_strs"`
	Values       []interface{} `json:"values"`
	GrowthValues []interface{} `json:"growth_values,omitempty"`
	GrowthRates  []interface{} `json:"growth_rates,omitempty"`
	Proportions  []interface{} `json:"proportions,omitempty"`
}

type MetricModel struct {
	UnitType string `json:"unit_type"`
	Unit     string `json:"unit"`
}

//go:generate mockgen -source ../interfaces/uniquery_access.go -destination ../interfaces/mock/mock_uniquery_access.go
type UniqueryAccess interface {
	GetViewDataByID(ctx context.Context, viewID string, viewRequest ViewQuery) (ViewData, error)
	GetMetricDataByID(ctx context.Context, metricID string, metricRequest MetricQuery) (MetricData, error)
}
