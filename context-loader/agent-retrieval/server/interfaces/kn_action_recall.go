// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package interfaces defines interfaces for business knowledge network action recall
package interfaces

import "context"

// ==================== Constant Definitions ====================

const (
	// ResultProcessStrategyKnActionRecall Result processing strategy for knowledge network action recall
	ResultProcessStrategyKnActionRecall = "kn_action_recall"
)

// ActionSource Type Constants
const (
	// ActionSourceTypeTool Tool type action source
	ActionSourceTypeTool = "tool"
	// ActionSourceTypeMCP MCP type action source (supported in next version)
	ActionSourceTypeMCP = "mcp"
)

// ==================== Request and Response Structures ====================

// KnActionRecallRequest Knowledge Network Action Recall Request
type KnActionRecallRequest struct {
	// Query Parameters
	KnID string `json:"kn_id" validate:"required"` // Knowledge Network ID
	AtID string `json:"at_id" validate:"required"` // Action Type ID

	// Request Body
	UniqueIdentity map[string]interface{} `json:"unique_identity" validate:"required,min=1"` // Object Unique Identity

	// Header Fields
	AccountID   string `json:"-" header:"x-account-id" validate:"required"`
	AccountType string `json:"-" header:"x-account-type" validate:"required"`
}

// KnActionRecallResponse Knowledge Network Action Recall Response
type KnActionRecallResponse struct {
	Headers      map[string]string `json:"headers"` // HTTP Header Parameters
	DynamicTools []KnDynamicTool   `json:"_dynamic_tools"`
}

// KnDynamicTool Dynamic Tool Definition
type KnDynamicTool struct {
	Name            string                 `json:"name"`              // Tool Name
	Description     string                 `json:"description"`       // Tool Description
	Parameters      map[string]interface{} `json:"parameters"`        // OpenAI Function Call Schema
	ApiURL          string                 `json:"api_url"`           // Tool Execution Proxy URL
	OriginalSchema  map[string]interface{} `json:"original_schema"`   // Original OpenAPI Definition
	FixedParams     KnFixedParams          `json:"fixed_params"`      // Fixed Parameters
	ApiCallStrategy string                 `json:"api_call_strategy"` // Result Processing Strategy, fixed value: kn_action_recall
}

// KnFixedParams Fixed Parameters Structure
type KnFixedParams struct {
	Header map[string]interface{} `json:"header"` // HTTP Header Parameters
	Path   map[string]interface{} `json:"path"`   // URL Path Parameters
	Query  map[string]interface{} `json:"query"`  // URL Query Parameters
	Body   map[string]interface{} `json:"body"`   // Request Body Parameters
}

// ==================== Action Query Related Structures ====================

// QueryActionsRequest Action Query Request
type QueryActionsRequest struct {
	KnID                string                   `json:"kn_id"`
	AtID                string                   `json:"at_id"`
	UniqueIdentities    []map[string]interface{} `json:"unique_identities"`
	IncludeTypeInfo     bool                     `json:"include_type_info"`
	XHTTPMethodOverride string                   `json:"-"` // Fixed to GET
}

// QueryActionsResponse Action Query Response
type QueryActionsResponse struct {
	ActionType   *ActionTypeInfo `json:"action_type,omitempty"` // Action Type Info
	ActionSource *ActionSource   `json:"action_source"`         // Action Source
	Actions      []ActionParams  `json:"actions"`               // Action Parameters List
	TotalCount   int             `json:"total_count"`
	OverallMs    int             `json:"overall_ms"`
}

// ActionTypeInfo Action Type Info
type ActionTypeInfo struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	ActionType   string                 `json:"action_type"` // add/modify/delete
	ObjectTypeID string                 `json:"object_type_id"`
	Parameters   []ActionTypeParam      `json:"parameters"`
	Condition    map[string]interface{} `json:"condition"`
	Affect       map[string]interface{} `json:"affect"`
	Schedule     map[string]interface{} `json:"schedule"`
}

// ActionTypeParam Action Type Parameter
type ActionTypeParam struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Source    string `json:"source"`
	ValueFrom string `json:"value_from"` // property/input/const
	Value     string `json:"value,omitempty"`
}

// ActionSource Action Source
type ActionSource struct {
	Type   string `json:"type"`    // tool/mcp
	BoxID  string `json:"box_id"`  // Tool Box ID
	ToolID string `json:"tool_id"` // Tool ID
}

// ActionParams Action Parameters
type ActionParams struct {
	Parameters    map[string]interface{} `json:"parameters"`     // Instantiated Parameters
	DynamicParams map[string]interface{} `json:"dynamic_params"` // Dynamic Parameters (value is null)
}

// ==================== Service Interfaces ====================

// IKnActionRecallService Knowledge Network Action Recall Service Interface
type IKnActionRecallService interface {
	// GetActionInfo gets action information (action recall)
	GetActionInfo(ctx context.Context, req *KnActionRecallRequest) (*KnActionRecallResponse, error)
}
