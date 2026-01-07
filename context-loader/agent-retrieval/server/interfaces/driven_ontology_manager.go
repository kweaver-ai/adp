// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

// KnOperationType Business knowledge network operator
type KnOperationType string

const (
	KnOperationTypeAnd            KnOperationType = "and"       // AND
	KnOperationTypeOr             KnOperationType = "or"        // OR
	KnOperationTypeEqual          KnOperationType = "=="        // Equal
	KnOperationTypeNotEqual       KnOperationType = "!="        // Not Equal
	KnOperationTypeGreater        KnOperationType = ">"         // Greater than
	KnOperationTypeLess           KnOperationType = "<"         // Less than
	KnOperationTypeGreaterOrEqual KnOperationType = ">="        // Greater than or equal
	KnOperationTypeLessOrEqual    KnOperationType = "<="        // Less than or equal
	KnOperationTypeIn             KnOperationType = "in"        // in
	KnOperationTypeNotIn          KnOperationType = "not_in"    // not_in
	KnOperationTypeLike           KnOperationType = "like"      // like
	KnOperationTypeNotLike        KnOperationType = "not_like"  // not_like
	KnOperationTypeRange          KnOperationType = "range"     // range
	KnOperationTypeOutRange       KnOperationType = "out_range" // out_range
	KnOperationTypeExist          KnOperationType = "exist"     // exist
	KnOperationTypeNotExist       KnOperationType = "not_exist" // not_exist
	KnOperationTypeRegex          KnOperationType = "regex"     // regex
	KnOperationTypeMatch          KnOperationType = "match"     // match
	KnOperationTypeKnn            KnOperationType = "knn"       // knn
)

// LogicPropertyType Logic property type
type LogicPropertyType string

const (
	LogicPropertyTypeMetric   LogicPropertyType = "metric"   // Metric type
	LogicPropertyTypeOperator LogicPropertyType = "operator" // Operator type
)

type KnBaseError struct {
	ErrorCode               string         `json:"error_code"`    // Error code
	Description             string         `json:"description"`   // Error description
	Solution                string         `json:"solution"`      // Solution
	ErrorLink               string         `json:"error_link"`    // Error link
	ErrorDetails            interface{}    `json:"error_details"` // Detail content
	DescriptionTemplateData map[string]any `json:"-"`             // Description parameters
	SolutionTemplateData    map[string]any `json:"-"`             // Solution parameters
}

type ResourceInfo struct {
	Type string `json:"type"` // Data source type
	ID   string `json:"id"`   // Data view ID
	Name string `json:"name"` // View name
}

type SimpleObjectType struct {
	OTID   string `json:"id"`
	OTName string `json:"name"`
	Icon   string `json:"icon"`
	Color  string `json:"color"`
}

// DataProperty Data property structure definition
type DataProperty struct {
	Name                string            `json:"name"`                 // Property name. Can only contain lowercase letters, numbers, underscores (_), hyphens (-), and cannot start with underscore or hyphen
	DisplayName         string            `json:"display_name"`         // Property display name
	Type                string            `json:"type"`                 // Property data type. In addition to view field types, there are metric, objective, event, trace, log, operator
	Comment             string            `json:"comment"`              // Comment
	MappedField         any               `json:"mapped_field"`         // View field info
	ConditionOperations []KnOperationType `json:"condition_operations"` // List of query condition operators supported by this data property
}

// LogicPropertyDef Logic property definition (extracted from object type definition)
type LogicPropertyDef struct {
	Name        string              `json:"name"`
	DisplayName string              `json:"display_name,omitempty"`
	Type        LogicPropertyType   `json:"type"` // Logic property type: metric or operator
	Comment     string              `json:"comment,omitempty"`
	DataSource  map[string]any      `json:"data_source,omitempty"`
	Parameters  []PropertyParameter `json:"parameters,omitempty"`
}

// PropertyParameter Property parameter definition
type PropertyParameter struct {
	Name             string `json:"name"`
	Type             string `json:"type"`
	ValueFrom        string `json:"value_from"` // "input", "property", "const"
	Value            any    `json:"value,omitempty"`
	IfSystemGenerate bool   `json:"if_system_generate,omitempty"`
	Comment          string `json:"comment,omitempty"`
}

// ObjectType Object type structure definition
type ObjectType struct {
	ModuleType      string              `json:"module_type"` // Module type
	ID              string              `json:"id"`          // Object ID
	Name            string              `json:"name"`        // Object name
	Tags            []string            `json:"tags"`        // Tags
	Comment         string              `json:"comment"`     // Comment
	Score           float64             `json:"_score"`      // Score
	DataSource      *ResourceInfo       `json:"data_source"`
	DataProperties  []*DataProperty     `json:"data_properties,omitempty"`  // Data properties
	LogicProperties []*LogicPropertyDef `json:"logic_properties,omitempty"` // Logic properties
	PrimaryKeys     []string            `json:"primary_keys"`               // Primary key fields
}

// RelationType Relation type structure definition
type RelationType struct {
	ModuleType string   `json:"module_type"` // Module type
	ID         string   `json:"id"`          // Relation type ID
	Name       string   `json:"name"`        // Relation type name
	Tags       []string `json:"tags"`        // Tags
	Comment    string   `json:"comment"`     // Comment
	Score      float64  `json:"_score"`      // Score

	SourceObjectTypeId string `json:"source_object_type_id"`        // Source object type ID
	TargetObjectTypeId string `json:"target_object_type_id"`        // Target object type ID
	SourceObjectType   any    `json:"source_object_type,omitempty"` // Provide name when viewing details
	TargetObjectType   any    `json:"target_object_type,omitempty"` // Provide name when viewing details
	MappingRules       any    `json:"mapping_rules"`                // Mapping rules based on type, direct corresponds to []Mapping structure
	Type               string `json:"type"`                         // Relation type
}

// ActionType Action type structure definition
type ActionType struct {
	ModuleType string   `json:"module_type"` // Module type
	ID         string   `json:"id"`          // Action type ID
	Name       string   `json:"name"`        // Action type name
	Tags       []string `json:"tags"`        // Tags
	Comment    string   `json:"comment"`     // Comment
	Score      float64  `json:"_score"`      // Score

	ObjectTypeId string `json:"object_type_id"` // Object type ID bound to action type
}

type KnCondValueFrom string

const (
	CondValueFromConst KnCondValueFrom = "const"
)

type KnCondLimitKey string

const (
	CondLimitKeyK           KnCondLimitKey = "k"            // Pagination key
	CondLimitKeyMinScore    KnCondLimitKey = "min_score"    // Min score
	CondLimitKeyMinDistance KnCondLimitKey = "min_distance" // Min distance
)

// KnCondition Retrieval condition
type KnCondition struct {
	Field         string          `json:"field"`          // Field name
	Operation     KnOperationType `json:"operation"`      // Operator
	SubConditions []*KnCondition  `json:"sub_conditions"` // Sub filtering conditions
	Value         any             `json:"value"`          // Field value
	ValueFrom     KnCondValueFrom `json:"value_from"`     // Field value source
	LimitKey      KnCondLimitKey  `json:"limit_key"`
	LimitValue    any             `json:"limit_value"`
}

type KnSortParams struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

// QueryConceptsReq Query concepts request
type QueryConceptsReq struct {
	KnID      string          `json:"-"`         // Knowledge network ID
	Cond      *KnCondition    `json:"condition"` // Retrieval condition
	Sort      []*KnSortParams `json:"sort"`
	Limit     int             `json:"limit"`      // Return count, default 10. Range 1-10000
	NeedTotal bool            `json:"need_total"` // Whether total count is needed, default false
}

// Concepts Retrieved concepts list
type Concepts struct {
	Entries     []any `json:"entries"`
	TotalCount  int64 `json:"total_count,omitempty"`
	SearchAfter []any `json:"search_after,omitempty"`
	OverallMs   int64 `json:"overall_ms"`
}

// ObjectTypeConcepts Object type concepts list
type ObjectTypeConcepts struct {
	Entries    []*ObjectType `json:"entries"`               // Object type data
	TotalCount int64         `json:"total_count,omitempty"` // Total count
}

// RelationTypeConcepts Relation type concepts list
type RelationTypeConcepts struct {
	Entries    []*RelationType `json:"entries"`               // Relation type data
	TotalCount int64           `json:"total_count,omitempty"` // Total count
}

// ActionTypeConcepts Action type concepts list
type ActionTypeConcepts struct {
	Entries    []*ActionType `json:"entries"`               // Action type data
	TotalCount int64         `json:"total_count,omitempty"` // Total count
}

// OntologyManagerAccess Ontology management interface
type OntologyManagerAccess interface {
	// SearchObjectTypes Search object types
	SearchObjectTypes(ctx context.Context, query *QueryConceptsReq) (objectTypes *ObjectTypeConcepts, err error)
	// GetObjectTypeDetail Get object type details
	GetObjectTypeDetail(ctx context.Context, knId string, otIds []string, includeDetail bool) ([]*ObjectType, error)

	// SearchRelationTypes Search relation types
	SearchRelationTypes(ctx context.Context, query *QueryConceptsReq) (releationTypes *RelationTypeConcepts, err error)
	// GetRelationTypeDetail Get relation type details
	GetRelationTypeDetail(ctx context.Context, knId string, rtIDs []string, includeDetail bool) ([]*RelationType, error)

	// SearchActionTypes Search action types
	SearchActionTypes(ctx context.Context, query *QueryConceptsReq) (actionTypes *ActionTypeConcepts, err error)
	// GetActionTypeDetail Get action type details
	GetActionTypeDetail(ctx context.Context, knId string, atIDs []string, includeDetail bool) ([]*ActionType, error)
}
