// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// ==================== Toolbox Service Related Structures ====================

// GetToolDetailRequest Get tool detail request
type GetToolDetailRequest struct {
	BoxID  string
	ToolID string
}

// GetToolDetailResponse Get tool detail response
type GetToolDetailResponse struct {
	ToolID       string         `json:"tool_id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Status       string         `json:"status"` // enabled/disabled
	MetadataType string         `json:"metadata_type"`
	Metadata     ToolMetadata   `json:"metadata"`
	UseRule      string         `json:"use_rule,omitempty"`
	GlobalParams map[string]any `json:"global_parameters,omitempty"`
	CreateTime   int64          `json:"create_time"`
	UpdateTime   int64          `json:"update_time"`
	CreateUser   string         `json:"create_user"`
	UpdateUser   string         `json:"update_user"`
	ExtendInfo   map[string]any `json:"extend_info,omitempty"`
}

// ToolMetadata Tool metadata
type ToolMetadata struct {
	Version     string         `json:"version"`
	Summary     string         `json:"summary"`
	Description string         `json:"description"`
	ServerURL   string         `json:"server_url"`
	Path        string         `json:"path"`
	Method      string         `json:"method"`
	CreateTime  int64          `json:"create_time"`
	UpdateTime  int64          `json:"update_time"`
	CreateUser  string         `json:"create_user"`
	UpdateUser  string         `json:"update_user"`
	ApiSpec     map[string]any `json:"api_spec"` // OpenAPI specification
}

// ==================== Driven Adapters Interface ====================

// DrivenOperatorIntegration Operator integration service interface
type DrivenOperatorIntegration interface {
	// GetToolDetail Get tool detail
	GetToolDetail(ctx context.Context, req *GetToolDetailRequest) (*GetToolDetailResponse, error)
}
