// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/bytedance/sonic"
)

const (
	//特征的配置项
	PropertyFeatureType_Keyword  = "keyword"
	PropertyFeatureType_Fulltext = "fulltext"
	PropertyFeatureType_Vector   = "vector"

	LogicDefinitionNodeType_Resource = "resource"
	LogicDefinitionNodeType_Join     = "join"
	LogicDefinitionNodeType_Union    = "union"
	LogicDefinitionNodeType_Sql      = "sql"
	LogicDefinitionNodeType_Output   = "output"

	// join的类型
	JoinType_Inner     = "inner"
	JoinType_Left      = "left"
	JoinType_Right     = "right"
	JoinType_FullOuter = "full outer"

	// union的类型
	UnionType_All      = "all"
	UnionType_Distinct = "distinct"
)

const (
	// 视图字段名称、字段显示名、字段备注、字段特征备注的最大长度
	MaxLength_ViewPropertyName               = 255
	MaxLength_ViewPropertyDisplayName        = 255
	MaxLength_ViewPropertyFeatureName        = 255
	MaxLength_ViewPropertyDescription        = 1000
	MaxLength_ViewPropertyFeatureDescription = 1000

	RegexPattern_NonBuiltin_ViewID = "^[a-z0-9][a-z0-9_-]{0,39}$"
)

var (
	LogicDefinitionNodeTypeMap = map[string]struct{}{
		LogicDefinitionNodeType_Resource: {},
		LogicDefinitionNodeType_Join:     {},
		LogicDefinitionNodeType_Union:    {},
		LogicDefinitionNodeType_Sql:      {},
		LogicDefinitionNodeType_Output:   {},
	}

	JoinTypeMap = map[string]struct{}{
		JoinType_Inner:     {},
		JoinType_Left:      {},
		JoinType_Right:     {},
		JoinType_FullOuter: {},
	}

	UnionTypeMap = map[string]struct{}{
		UnionType_All:      {},
		UnionType_Distinct: {},
	}

	PropertyFeatureTypeMap = map[string]struct{}{
		PropertyFeatureType_Keyword:  {},
		PropertyFeatureType_Fulltext: {},
		PropertyFeatureType_Vector:   {},
	}
)

type LogicView struct {
	Resource
	FieldsMap map[string]*ViewProperty `json:"fields_map,omitempty" mapstructure:"-"`
}

// LogicDefinitionNode 表示图中的节点
type LogicDefinitionNode struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Type         string          `json:"type"`
	Inputs       []string        `json:"inputs"`
	Config       map[string]any  `json:"config"`
	OutputFields []*ViewProperty `json:"output_fields"`
}

// 节点类型为resource的节点配置
type ResourceNodeCfg struct {
	ResourceID string         `json:"resource_id" mapstructure:"resource_id"`
	Filters    *FilterCondCfg `json:"filters,omitempty" mapstructure:"filters"`
	Distinct   Distinct       `json:"distinct" mapstructure:"distinct"`
	Resource   *Resource      `json:"resource,omitempty" mapstructure:"resource"`
}

type Distinct struct {
	Enable bool     `json:"enable" mapstructure:"enable"`
	Fields []string `json:"fields,omitempty" mapstructure:"fields"`
}

// 节点类型为join的节点配置
type JoinNodeCfg struct {
	JoinType string         `json:"join_type" mapstructure:"join_type"`
	JoinOn   []*JoinOn      `json:"join_on" mapstructure:"join_on"`
	Filters  *FilterCondCfg `json:"filters,omitempty" mapstructure:"filters"`
	Distinct Distinct       `json:"distinct,omitempty" mapstructure:"distinct"`
}

// join on 配置
type JoinOn struct {
	LeftField  string `json:"left_field" mapstructure:"left_field"`   //传递 name
	RightField string `json:"right_field" mapstructure:"right_field"` //传递 name
	Operator   string `json:"operator" mapstructure:"operator"`
}

// 节点类型为union的节点配置
type UnionNodeCfg struct {
	UnionType string         `json:"union_type" mapstructure:"union_type"`
	Filters   *FilterCondCfg `json:"filters,omitempty" mapstructure:"filters"`
}

type SQLNodeCfg struct {
	SQL string `json:"sql" mapstructure:"sql"`
}

// OutputFieldRef 表示 Union 对齐模式中 from 数组的元素
type OutputFieldRef struct {
	From     string `json:"from"`
	FromNode string `json:"from_node"`
}

// 逻辑视图字段
type ViewProperty struct {
	Property
	From     string            `json:"from,omitempty"`      // Join 映射模式：源字段名
	FromNode string            `json:"from_node,omitempty"` // Join 映射模式：源节点ID
	FromList []*OutputFieldRef `json:"from_list,omitempty"` // Union 对齐模式：多源对齐数组
}

// UnmarshalJSON 自定义反序列化，处理 from 字段的多态
// JSON 中 "from" 可以是 string (Join) 或 array (Union)
func (v *ViewProperty) UnmarshalJSON(data []byte) error {
	// 先用一个临时的 map 来探测 from 字段的类型
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		// 如果不是对象，可能是纯字符串（投影模式 / 通配符模式）
		var s string
		if err2 := json.Unmarshal(data, &s); err2 == nil {
			v.Name = s
			return nil
		}
		return err
	}

	// 解码 Property 的字段
	type PropertyAlias Property
	var propAlias PropertyAlias
	if err := json.Unmarshal(data, &propAlias); err != nil {
		return err
	}
	v.Property = Property(propAlias)

	// 解码 from_node
	if rawFromNode, ok := raw["from_node"]; ok {
		if err := json.Unmarshal(rawFromNode, &v.FromNode); err != nil {
			return fmt.Errorf("failed to unmarshal from_node: %w", err)
		}
	}

	// 解码 from: 可能是 string 或 array
	if rawFrom, ok := raw["from"]; ok {
		// 尝试 string
		var fromStr string
		if err := json.Unmarshal(rawFrom, &fromStr); err == nil {
			v.From = fromStr
			return nil
		}

		// 尝试 array (Union 对齐模式)
		var fromList []*OutputFieldRef
		if err := json.Unmarshal(rawFrom, &fromList); err == nil {
			v.FromList = fromList
			return nil
		}

		return fmt.Errorf("failed to unmarshal 'from' field: expected string or array")
	}

	return nil
}

func (v *ViewProperty) String() string {
	return fmt.Sprintf("ViewProperty{name: %s, type: %s, description: %s, display_name: %s, original_name: %s}",
		v.Name, v.Type, v.Description, v.DisplayName, v.OriginalName)
}

type DSLCfg struct {
	From           int              `json:"from"`
	Size           int              `json:"size"`
	Sort           []map[string]any `json:"sort,omitempty"`
	TrackScores    bool             `json:"track_scores,omitempty"`
	TrackTotalHits bool             `json:"track_total_hits,omitempty"`
	SearchAfter    []any            `json:"search_after,omitempty"`
	Query          struct {
		Bool struct {
			Should         []any `json:"should,omitempty"`
			Filter         []any `json:"filter,omitempty"`
			Must           []any `json:"must,omitempty"`
			MinShouldMatch int   `json:"minimum_should_match,omitempty"`
		} `json:"bool"`
	} `json:"query"`
	Pit *struct {
		ID        string `json:"id,omitempty"`
		KeepAlive string `json:"keep_alive,omitempty"`
	} `json:"pit,omitempty"`
}

func (dsl DSLCfg) String() string {
	bytes, _ := sonic.MarshalIndent(dsl, "", "  ")
	return string(bytes)
}

type SearchAfterParams struct {
	SearchAfter  []any  `json:"search_after"`
	PitID        string `json:"pit_id"`
	PitKeepAlive string `json:"pit_keep_alive"`
}
