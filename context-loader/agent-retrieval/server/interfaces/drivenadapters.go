// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package interfaces defines interfaces
// @file drivenadapters.go
// @description: Inbound interface definition
package interfaces

//go:generate mockgen -source=drivenadapters.go -destination=../mocks/drivenadapters.go -package=mocks
import (
	"context"
)

// AccountAuthContext Account authentication context
type AccountAuthContext struct {
	// AccountID Account unique identifier
	AccountID string `json:"account_id"`
	// AccountType Account Type
	AccountType AccessorType `json:"account_type"`
	// TokenInfo Token information
	TokenInfo *TokenInfo `json:"token_info"`
}

const (
	// SystemUser System
	SystemUser = "system"
	// UnknownUser Unknown
	UnknownUser = "unknown"
)

// VisitorType Visitor Type
type VisitorType string

// Visitor type definitions
const (
	RealName  VisitorType = "realname"  // Real-name user
	Anonymous VisitorType = "anonymous" // Anonymous user
	Business  VisitorType = "business"  // Application account
)

// ToAccessorType Converts to AccessorType
func (v VisitorType) ToAccessorType() AccessorType {
	switch v {
	case RealName:
		return AccessorTypeUser
	case Business:
		return AccessorTypeApp
	case Anonymous:
		return AccessorTypeAnonymous
	default:
		// Unknown visitor type, default to anonymous user
		return AccessorTypeAnonymous
	}
}

// AccessorType Accessor Type
type AccessorType string

const (
	AccessorTypeUser       AccessorType = "user"       // Real-name user
	AccessorTypeDepartment AccessorType = "department" // Department
	AccessorTypeGroup      AccessorType = "group"      // Organization
	AccessorTypeRole       AccessorType = "role"       // Role
	AccessorTypeApp        AccessorType = "app"        // Application account
	AccessorTypeAnonymous  AccessorType = "anonymous"  // Anonymous access
)

// ToVisitorType Converts AccessorType to VisitorType
func (a AccessorType) ToVisitorType() VisitorType {
	switch a {
	case AccessorTypeUser:
		return RealName
	case AccessorTypeApp:
		return Business
	case AccessorTypeAnonymous:
		return Anonymous
	case AccessorTypeDepartment, AccessorTypeGroup, AccessorTypeRole:
		return ""
	default:
		return ""
	}
}

// AccountType Login account type
type AccountType int32

// Login account type definition
const (
	Other  AccountType = 0
	IDCard AccountType = 1
)

const (
	// AccessedByUser Real-name user
	AccessedByUser string = "accessed_by_users"
	// AccessedByAnyOne Anonymous user
	AccessedByAnyOne string = "accessed_by_anyone"
)

// ClientType Device type
type ClientType int32

// ClientTypeMap Client type map
var ClientTypeMap = map[ClientType]string{
	Unknown:      "unknown",
	IOS:          "ios",
	Android:      "android",
	WindowsPhone: "windows_phone",
	Windows:      "windows",
	MacOS:        "mac_os",
	Web:          "web",
	MobileWeb:    "mobile_web",
	Nas:          "nas",
	ConsoleWeb:   "console_web",
	DeployWeb:    "deploy_web",
	Linux:        "linux",
	APP:          "app",
}

// ReverseClientTypeMap Reverse client type map
var ReverseClientTypeMap = map[string]ClientType{
	"unknown":       Unknown,
	"ios":           IOS,
	"android":       Android,
	"windows_phone": WindowsPhone,
	"windows":       Windows,
	"mac_os":        MacOS,
	"web":           Web,
	"mobile_web":    MobileWeb,
	"nas":           Nas,
	"console_web":   ConsoleWeb,
	"deploy_web":    DeployWeb,
	"linux":         Linux,
	"app":           APP,
}

// AccountTypeMap Account type map
var AccountTypeMap = map[AccountType]string{
	Other:  "other_category",
	IDCard: "id_card",
}

// ReverseAccountTypeMap Reverse account type map
var ReverseAccountTypeMap = map[string]AccountType{
	"other_category": Other,
	"id_card":        IDCard,
}

func (typ ClientType) String() string {
	str, ok := ClientTypeMap[typ]
	if !ok {
		str = ClientTypeMap[Unknown]
	}
	return str
}

// Device type definition
const (
	Unknown ClientType = iota
	IOS
	Android
	WindowsPhone
	Windows
	MacOS
	Web
	MobileWeb
	Nas
	ConsoleWeb
	DeployWeb
	Linux
	APP
)

// TokenInfo Authorization verification information
type TokenInfo struct {
	Active     bool        // Token status
	VisitorID  string      // Visitor ID
	Scope      string      // Permission scope
	ClientID   string      // Client ID
	VisitorTyp VisitorType // Visitor type
	// Following fields exist only when visitorType=realname (real-name user)
	LoginIP     string      // Login IP
	Udid        string      // Device ID
	AccountTyp  AccountType // Account type
	ClientTyp   ClientType  // Device type
	PhoneNumber string      // Phone number for anonymous users
	VisitorName string      // Nickname for anonymous visitors
	MAC         string      // MAC address
	UserAgent   string      // User agent info
}

// Hydra Authorization service interface
type Hydra interface {
	Introspect(ctx context.Context, token string) (tokenInfo *TokenInfo, err error)
}

const (
	// DisplayName User display name
	DisplayName = "name"
)

// UserInfo User information
type UserInfo struct {
	UserID      string   `json:"id"`    // User ID
	DisplayName string   `json:"name"`  // User display name
	Roles       []string `json:"roles"` // Roles
	Account     string   `json:"account"`
}

// AppInfo App account information
type AppInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ErrorResponse Error response
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  struct {
		IDs []string `json:"ids"`
	} `json:"detail"`
}

// UserManagement User management interface
type UserManagement interface {
	GetAppInfo(ctx context.Context, appID string) (appInfo *AppInfo, err error)
	GetUserInfo(ctx context.Context, userID string, fields ...string) (info *UserInfo, err error)
	GetUsersInfo(ctx context.Context, userIDs []string, fields []string) (infos []*UserInfo, err error)
	GetUsersName(ctx context.Context, userIDs []string) (userMap map[string]string, err error)
}

// KnowledgeRerankActionType Result set rerank type based on business knowledge network
type KnowledgeRerankActionType string

const (
	KnowledgeRerankActionLLM     KnowledgeRerankActionType = "llm"     // Rerank based on LLM
	KnowledgeRerankActionVector  KnowledgeRerankActionType = "vector"  // Rerank based on vector
	KnowledgeRerankActionDefault KnowledgeRerankActionType = "default" // Default rerank
)

// KnowledgeRerankReq Knowledge rerank request
type KnowledgeRerankReq struct {
	QueryUnderstanding *QueryUnderstanding       `json:"query_understanding" validate:"required"`                      // Query understanding
	KnowledgeConcepts  []*ConceptResult          `json:"concepts" validate:"required"`                                 // Business knowledge network concepts
	Action             KnowledgeRerankActionType `json:"action" validate:"required,oneof=llm vector" default:"vector"` // Action: llm based rerank, vector based rerank
}

// KnDataSourceConfig Knowledge network data source configuration
type KnDataSourceConfig struct {
	KnowledgeNetworkID string `json:"knowledge_network_id"` // Knowledge Network ID
}

// KnSearchReq kn_search request
type KnSearchReq struct {
	// Header Parameters
	XAccountID   string `header:"x-account-id"`
	XAccountType string `header:"x-account-type"`

	// Body Parameters - use any to avoid defining complex structures explicitly
	// Corresponds to the complete request structure of data-retrieval interface
	Query             string                `json:"query" validate:"required"`
	KnID              string                `json:"kn_id" validate:"required"`
	knIDs             []*KnDataSourceConfig // Internal use, converted from KnID, not exposed
	SessionID         *string               `json:"session_id,omitempty"`
	AdditionalContext *string               `json:"additional_context,omitempty"`
	RetrievalConfig   any                   `json:"retrieval_config,omitempty"`
	OnlySchema        *bool                 `json:"only_schema,omitempty"`
	EnableRerank      *bool                 `json:"enable_rerank,omitempty"`
}

// SetKnIDs Sets knIDs (internal use, converted from KnID)
func (r *KnSearchReq) SetKnIDs(knIDs []*KnDataSourceConfig) {
	r.knIDs = knIDs
}

// GetKnIDs Gets knIDs (internal use)
func (r *KnSearchReq) GetKnIDs() []*KnDataSourceConfig {
	return r.knIDs
}

// KnSearchResp kn_search response
type KnSearchResp struct {
	// Use any to directly return the original structure from the underlying interface
	// Corresponds to the complete response structure of data-retrieval interface
	ObjectTypes   any     `json:"object_types,omitempty"`
	RelationTypes any     `json:"relation_types,omitempty"`
	ActionTypes   any     `json:"action_types,omitempty"`
	Nodes         any     `json:"nodes,omitempty"`
	Message       *string `json:"message,omitempty"`
}

// DataRetrieval Data retrieval interface
type DataRetrieval interface {
	KnowledgeRerank(ctx context.Context, req *KnowledgeRerankReq) (results []*ConceptResult, err error)
	// KnSearch Knowledge network retrieval
	KnSearch(ctx context.Context, req *KnSearchReq) (resp *KnSearchResp, err error)
}
