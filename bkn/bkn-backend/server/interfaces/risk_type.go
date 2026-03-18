// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

const (
	// 风险等级常量
	RISK_LEVEL_SAFE     = "safe"
	RISK_LEVEL_LOW      = "low"
	RISK_LEVEL_MEDIUM   = "medium"
	RISK_LEVEL_HIGH     = "high"
	RISK_LEVEL_CRITICAL = "critical"

	// 内置风险评估函数信息
	BuiltinToolBoxID         = "bkn-internal_risk-assessment"
	BuiltinToolBoxName       = "BKN风险评估工具"
	BuiltinToolConfigVersion = "0.5.0"
	BuiltinToolToolID        = "bkn_common_risk_assessment_tool"
)

var (
	RiskLevelOrder = map[string]int{
		RISK_LEVEL_SAFE:     1,
		RISK_LEVEL_LOW:      2,
		RISK_LEVEL_MEDIUM:   3,
		RISK_LEVEL_HIGH:     4,
		RISK_LEVEL_CRITICAL: 5,
	}
)

// RiskType 风险类
type RiskType struct {
	RTID               string `json:"id" mapstructure:"id"`
	RTName             string `json:"name" mapstructure:"name"`
	CommonInfo         `mapstructure:",squash"`
	KNID               string        `json:"kn_id" mapstructure:"kn_id"`
	Branch             string        `json:"branch" mapstructure:"branch"`
	MaxAcceptableLevel string        `json:"max_acceptable_level" mapstructure:"max_acceptable_level"`
	Parameters         []ParamDef    `json:"parameters" mapstructure:"parameters"`
	RiskRules          []RiskRule    `json:"risk_rules" mapstructure:"risk_rules"`
	RiskFunction       *RiskFunction `json:"risk_function" mapstructure:"risk_function"`
	Creator            AccountInfo   `json:"creator" mapstructure:"creator"`
	CreateTime         int64         `json:"create_time" mapstructure:"create_time"`
	Updater            AccountInfo   `json:"updater" mapstructure:"updater"`
	UpdateTime         int64         `json:"update_time" mapstructure:"update_time"`
	ModuleType         string        `json:"module_type" mapstructure:"module_type"`

	Vector []float32 `json:"_vector,omitempty"`
	Score  *float64  `json:"_score,omitempty"` // opensearch检索的得分，在概念搜索时使用
}

// ParamDef 参数定义（RiskType 与 RiskFunction 共用）
type ParamDef struct {
	Name        string `json:"name" mapstructure:"name"`
	Type        string `json:"type" mapstructure:"type"`
	Required    bool   `json:"required" mapstructure:"required"`
	Description string `json:"description" mapstructure:"description"`
	Default     any    `json:"default,omitempty" mapstructure:"default"`
}

// RiskRule 风险规则
type RiskRule struct {
	ID          string        `json:"id" mapstructure:"id"`
	Name        string        `json:"name" mapstructure:"name"`
	Description string        `json:"description" mapstructure:"description"`
	When        *RiskRuleWhen `json:"when" mapstructure:"when"`
	Decision    string        `json:"decision" mapstructure:"decision"`
	Message     string        `json:"message" mapstructure:"message"`
}

// RiskRuleWhen 命中条件
type RiskRuleWhen struct {
	Type            string   `json:"type" mapstructure:"type"` // condition | natural_language
	Condition       *CondCfg `json:"condition,omitempty" mapstructure:"condition"`
	NaturalLanguage string   `json:"natural_language,omitempty" mapstructure:"natural_language"`
}

// RiskFunction 风险评估函数
type RiskFunction struct {
	Type       string      `json:"type" mapstructure:"type"` // tool
	BoxID      string      `json:"box_id,omitempty" mapstructure:"box_id"`
	ToolID     string      `json:"tool_id,omitempty" mapstructure:"tool_id"`
	Parameters []Parameter `json:"parameters,omitempty" mapstructure:"parameters"` // 扁平列表，每个参数的 source 标记位置 path/query/header/body
}

// RiskTypesQueryParams 风险类查询参数
type RiskTypesQueryParams struct {
	PaginationQueryParameters
	NamePattern string
	Tag         string
	Branch      string
	KNID        string
}

var (
	RiskTypeSort = map[string]string{
		"name":        "f_name",
		"update_time": "f_update_time",
	}
)

// RiskTypes 风险类列表
type RiskTypes struct {
	Entries     []*RiskType `json:"entries"`
	TotalCount  int64       `json:"total_count,omitempty"`
	SearchAfter []any       `json:"search_after,omitempty"`
	OverallMs   int64       `json:"overall_ms"`
}

// RiskTypeAccess 风险类数据访问接口
type RiskTypeAccess interface {
	CheckRiskTypeExistByID(ctx context.Context, knID string, branch string, rtID string) (string, bool, error)
	CheckRiskTypeExistByName(ctx context.Context, knID string, branch string, rtName string) (string, bool, error)
	CreateRiskType(ctx context.Context, tx *sql.Tx, riskType *RiskType) error
	ListRiskTypes(ctx context.Context, query RiskTypesQueryParams) ([]*RiskType, error)
	GetRiskTypesTotal(ctx context.Context, query RiskTypesQueryParams) (int, error)
	GetRiskTypesByIDs(ctx context.Context, knID string, branch string, rtIDs []string) ([]*RiskType, error)
	UpdateRiskType(ctx context.Context, tx *sql.Tx, riskType *RiskType) error
	DeleteRiskTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, rtIDs []string) (int64, error)
	GetAllRiskTypesByKnID(ctx context.Context, knID string, branch string) ([]*RiskType, error)
	DeleteRiskTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) (int64, error)
}
