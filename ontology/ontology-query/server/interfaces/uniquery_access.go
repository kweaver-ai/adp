package interfaces

import (
	"context"
	"encoding/json"
	cond "ontology-query/common/condition"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
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

// SearchAfterArray 自定义类型，用于保持大整数精度
type SearchAfterArray []any

// UnmarshalJSON 自定义反序列化方法，确保大整数不会丢失精度
func (s *SearchAfterArray) UnmarshalJSON(data []byte) error {
	var result []any

	// 使用 UseInt64 选项
	cfg := sonic.Config{UseInt64: true}.Froze()
	if err := cfg.Unmarshal(data, &result); err != nil {
		logger.Errorf("Unmarshal Search After failed, %s", err)
		return err
	}

	*s = result
	return nil
}

// MarshalJSON 自定义序列化方法，确保正确输出
func (s SearchAfterArray) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any(s))
}

type SearchAfterParams struct {
	SearchAfter SearchAfterArray `json:"search_after"`
	// PitID        string `json:"pit_id"`
	// PitKeepAlive string `json:"pit_keep_alive"`
}

type ViewData struct {
	Datas       []map[string]any `json:"entries"`
	TotalCount  int64            `json:"total_count"`
	SearchAfter []any            `json:"search_after,omitempty"`
}

type MetricQuery struct {
	Start          *int64   `json:"start"`
	End            *int64   `json:"end"`
	StepStr        *string  `json:"step"`
	IsInstantQuery bool     `json:"instant"`
	Filters        []Filter `json:"filters"`
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
	Values []interface{}     `json:"values"`
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
