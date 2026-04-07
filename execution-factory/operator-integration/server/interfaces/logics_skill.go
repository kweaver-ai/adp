package interfaces

import (
	"context"
	"encoding/json"
)

//go:generate mockgen -source=logics_skill.go -destination=../mocks/logics_skill.go -package=mocks

// RegisterSkillReq 注册 Skill 请求
type RegisterSkillReq struct {
	BusinessDomainID string          `header:"x-business-domain" validate:"required"`
	UserID           string          `header:"user_id" validate:"required"`
	FileType         string          `form:"file_type" validate:"required,oneof=zip content"`
	File             json.RawMessage `form:"file" validate:"required"`
	Category         BizCategory     `form:"category" default:"other_category" validate:"required"`
	Source           string          `form:"source" default:"custom" validate:"oneof=custom internal"`
	ExtendInfo       json.RawMessage `form:"extend_info"`
}

// RegisterSkillResp 注册 Skill 响应
type RegisterSkillResp struct {
	SkillID     string    `json:"skill_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Status      BizStatus `json:"status"`
	Files       []string  `json:"files"`
}

// DeleteSkillReq 删除 Skill 请求
type DeleteSkillReq struct {
	BusinessDomainID string `header:"x-business-domain" validate:"required"`
	UserID           string `header:"user_id" validate:"required"`
	SkillID          string `uri:"skill_id" validate:"required"`
}

// DownloadSkillReq 下载 Skill 请求
type DownloadSkillReq struct {
	BusinessDomainID string `header:"x-business-domain" validate:"required"`
	UserID           string `header:"user_id"`
	SkillID          string `uri:"skill_id" validate:"required"`
}

// DownloadSkillResp 下载 Skill 响应
type DownloadSkillResp struct {
	SkillID  string `json:"skill_id"`
	FileName string `json:"file_name"`
	Content  []byte `json:"content"`
}

// QuerySkillListReq Skill 列表查询
type QuerySkillListReq struct {
	BusinessDomainID string      `header:"x-business-domain" validate:"required"`
	UserID           string      `header:"user_id"`
	Name             string      `form:"name"`
	Status           BizStatus   `form:"status" validate:"omitempty,oneof=unpublish published offline"`
	Category         BizCategory `form:"category"`
	CreateUser       string      `form:"create_user"`
	CommonPageParams `json:",inline"`
}

// SkillInfo Skill 详情
type SkillInfo struct {
	SkillID          string         `json:"skill_id"`
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	Version          string         `json:"version"`
	Status           BizStatus      `json:"status"`
	Source           string         `json:"source"`
	Dependencies     map[string]any `json:"dependencies,omitempty"`
	ExtendInfo       map[string]any `json:"extend_info,omitempty"`
	CreateUser       string         `json:"create_user"`
	CreateTime       int64          `json:"create_time"`
	UpdateUser       string         `json:"update_user"`
	UpdateTime       int64          `json:"update_time"`
	Category         BizCategory    `json:"category,omitempty"`
	CategoryName     string         `json:"category_name,omitempty"`
	BusinessDomainID string         `json:"business_domain_id"`
}

// SkillSummary Skill 列表摘要
type SkillSummary struct {
	SkillID          string      `json:"skill_id"`
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	Version          string      `json:"version"`
	Status           BizStatus   `json:"status"`
	Source           string      `json:"source"`
	CreateUser       string      `json:"create_user"`
	CreateTime       int64       `json:"create_time"`
	UpdateUser       string      `json:"update_user"`
	UpdateTime       int64       `json:"update_time"`
	BusinessDomainID string      `json:"business_domain_id"`
	Category         BizCategory `json:"category,omitempty"`
	CategoryName     string      `json:"category_name,omitempty"`
}

// SkillFileSummary Skill 文件摘要
type SkillFileSummary struct {
	RelPath  string `json:"rel_path"`
	FileType string `json:"file_type"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

// QuerySkillListResp Skill 列表响应
type QuerySkillListResp struct {
	CommonPageResult `json:",inline"`
	Data             []*SkillSummary `json:"data"`
}

// QuerySkillMarketListReq Skill 市场列表查询
type QuerySkillMarketListReq struct {
	BusinessDomainID string      `header:"x-business-domain" validate:"required"`
	UserID           string      `header:"user_id"`
	Name             string      `form:"name"`
	Category         BizCategory `form:"category"`
	CreateUser       string      `form:"create_user"`
	CommonPageParams `json:",inline"`
}

// QuerySkillMarketListResp Skill 市场列表响应
type QuerySkillMarketListResp struct {
	CommonPageResult `json:",inline"`
	Data             []*SkillSummary `json:"data"`
}

// GetSkillDetailReq Skill 详情查询
type GetSkillDetailReq struct {
	BusinessDomainID string `header:"x-business-domain" validate:"required"`
	UserID           string `header:"user_id"`
	SkillID          string `uri:"skill_id" validate:"required"`
}

// GetSkillMarketDetailReq Skill 市场详情查询
type GetSkillMarketDetailReq struct {
	BusinessDomainID string `header:"x-business-domain" validate:"required"`
	UserID           string `header:"user_id"`
	SkillID          string `uri:"skill_id" validate:"required"`
}

// GetSkillContentReq Skill 内容查询
type GetSkillContentReq struct {
	BusinessDomainID string `header:"x-business-domain" validate:"required"`
	UserID           string `header:"user_id"`
	SkillID          string `uri:"skill_id" validate:"required"`
}

// GetSkillContentResp Skill 内容响应
type GetSkillContentResp struct {
	SkillID             string                    `json:"skill_id"`
	URL                 string                    `json:"url"`
	Files               []*SkillFileSummary       `json:"files"`
	Status              BizStatus                 `json:"status"`
	RuntimeCapabilities []*SkillRuntimeCapability `json:"runtime_capabilities,omitempty"`
}

// ReadSkillFileReq 读取 Skill 文件请求
type ReadSkillFileReq struct {
	BusinessDomainID string `header:"x-business-domain" validate:"required"`
	UserID           string `header:"user_id"`
	SkillID          string `uri:"skill_id" validate:"required"`
	RelPath          string `json:"rel_path" validate:"required"`
}

// ReadSkillFileResp 读取 Skill 文件响应
type ReadSkillFileResp struct {
	SkillID  string `json:"skill_id"`
	RelPath  string `json:"rel_path"`
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	FileType string `json:"file_type"`
}

// SkillRuntimeProfileInfo Skill 可执行配置
type SkillRuntimeProfileInfo struct {
	SkillID         string         `json:"skill_id"`
	SkillVersion    string         `json:"skill_version"`
	Entrypoint      string         `json:"entrypoint"`
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	RuntimeType     string         `json:"runtime_type"`
	CommandTemplate []string       `json:"command_template"`
	InputSchema     map[string]any `json:"input_schema,omitempty"`
	OutputSchema    map[string]any `json:"output_schema,omitempty"`
	Timeout         int            `json:"timeout"`
	Status          string         `json:"status"`
	ExtendInfo      map[string]any `json:"extend_info,omitempty"`
	CreateUser      string         `json:"create_user"`
	CreateTime      int64          `json:"create_time"`
	UpdateUser      string         `json:"update_user"`
	UpdateTime      int64          `json:"update_time"`
}

// SkillRuntimeCapability 面向模型暴露的可执行能力摘要
type SkillRuntimeCapability struct {
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	RuntimeType  string         `json:"runtime_type,omitempty"`
	InputSchema  map[string]any `json:"input_schema,omitempty"`
	OutputSchema map[string]any `json:"output_schema,omitempty"`
}

// UpsertSkillRuntimeProfileReq 新增/更新 Skill 执行配置请求
type UpsertSkillRuntimeProfileReq struct {
	BusinessDomainID string         `header:"x-business-domain" validate:"required"`
	UserID           string         `header:"user_id" validate:"required"`
	SkillID          string         `uri:"skill_id" validate:"required"`
	Entrypoint       string         `uri:"entrypoint" validate:"required"`
	Name             string         `json:"name" validate:"required"`
	Description      string         `json:"description" validate:"required"`
	RuntimeType      string         `json:"runtime_type" default:"python" validate:"required"`
	CommandTemplate  []string       `json:"command_template" validate:"required"`
	InputSchema      map[string]any `json:"input_schema"`
	OutputSchema     map[string]any `json:"output_schema"`
	Timeout          int            `json:"timeout"`
	Status           BizStatus      `json:"status" default:"published" validate:"omitempty,oneof=unpublish published offline"`
	ExtendInfo       map[string]any `json:"extend_info"`
}

// UpsertSkillRuntimeProfileResp Skill 执行配置响应
type UpsertSkillRuntimeProfileResp struct {
	Profile *SkillRuntimeProfileInfo `json:"profile"`
}

// GetSkillRuntimeProfileReq 查询 Skill 执行配置请求
type GetSkillRuntimeProfileReq struct {
	BusinessDomainID string `header:"x-business-domain" validate:"required"`
	UserID           string `header:"user_id" validate:"required"`
	SkillID          string `uri:"skill_id" validate:"required"`
	Entrypoint       string `uri:"entrypoint" validate:"required"`
}

// GetSkillRuntimeProfileResp 查询 Skill 执行配置响应
type GetSkillRuntimeProfileResp struct {
	Profile *SkillRuntimeProfileInfo `json:"profile"`
}

// ExecuteSkillReq Skill 执行请求
type ExecuteSkillReq struct {
	BusinessDomainID string         `header:"x-business-domain" validate:"required"`
	UserID           string         `header:"user_id" validate:"required"`
	SkillID          string         `uri:"skill_id" validate:"required"`
	Entrypoint       string         `uri:"entrypoint" validate:"required"`
	Inputs           map[string]any `json:"inputs"`
	Timeout          int            `json:"timeout"`
}

// ExecuteSkillResp Skill 执行响应
type ExecuteSkillResp struct {
	SkillID       string                   `json:"skill_id"`
	SkillVersion  string                   `json:"skill_version"`
	Entrypoint    string                   `json:"entrypoint"`
	SessionID     string                   `json:"session_id"`
	RuntimeType   string                   `json:"runtime_type"`
	ExitCode      int                      `json:"exit_code"`
	ErrorMessage  string                   `json:"error_message,omitempty"`
	ExecutionTime float64                  `json:"execution_time"`
	Stdout        string                   `json:"stdout"`
	Stderr        string                   `json:"stderr"`
	ReturnValue   any                      `json:"return_value"`
	Profile       *SkillRuntimeProfileInfo `json:"profile,omitempty"`
}

// UpdateSkillStatusReq 更新 Skill 状态请求
type UpdateSkillStatusReq struct {
	BusinessDomainID string    `header:"x-business-domain" validate:"required"`
	UserID           string    `header:"user_id"`
	SkillID          string    `uri:"skill_id" validate:"required"`
	Status           BizStatus `json:"status" validate:"required,oneof=unpublish published offline"`
}

// UpdateSkillStatusResp 更新 Skill 状态响应
type UpdateSkillStatusResp struct {
	SkillID string    `json:"skill_id"`
	Status  BizStatus `json:"status"`
}

// SkillRegistry Skill 管理接口
type SkillRegistry interface {
	RegisterSkill(ctx context.Context, req *RegisterSkillReq) (*RegisterSkillResp, error)
	DeleteSkill(ctx context.Context, req *DeleteSkillReq) error
	DownloadSkill(ctx context.Context, req *DownloadSkillReq) (*DownloadSkillResp, error)
	QuerySkillList(ctx context.Context, req *QuerySkillListReq) (*QuerySkillListResp, error)
	GetSkillDetail(ctx context.Context, req *GetSkillDetailReq) (*SkillInfo, error)
	// 更新 Skill 状态
	UpdateSkillStatus(ctx context.Context, req *UpdateSkillStatusReq) (*UpdateSkillStatusResp, error)
}

// SkillMarket Skill 市场接口
type SkillMarket interface {
	QuerySkillMarketList(ctx context.Context, req *QuerySkillMarketListReq) (*QuerySkillMarketListResp, error)
	GetSkillMarketDetail(ctx context.Context, req *GetSkillMarketDetailReq) (*SkillInfo, error)
}

// SkillReader Skill 只读接口
type SkillReader interface {
	GetSkillContent(ctx context.Context, req *GetSkillContentReq) (*GetSkillContentResp, error)
	ReadSkillFile(ctx context.Context, req *ReadSkillFileReq) (*ReadSkillFileResp, error)
}

// SkillRuntimeProfileService Skill 执行配置管理接口
type SkillRuntimeProfileService interface {
	UpsertSkillRuntimeProfile(ctx context.Context, req *UpsertSkillRuntimeProfileReq) (*UpsertSkillRuntimeProfileResp, error)
	GetSkillRuntimeProfile(ctx context.Context, req *GetSkillRuntimeProfileReq) (*GetSkillRuntimeProfileResp, error)
}

// SkillExecutionService Skill 执行编排接口
type SkillExecutionService interface {
	ExecuteSkill(ctx context.Context, req *ExecuteSkillReq) (*ExecuteSkillResp, error)
}
