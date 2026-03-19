package interfaces

import (
	"context"
	"encoding/json"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
)

//go:generate mockgen -source=logics_skill.go -destination=../mocks/logics_skill.go -package=mocks

// SkillFileAccessLevel Skill 文件访问级别
type SkillFileAccessLevel string

const (
	SkillFileAccessLevelPublicManifest SkillFileAccessLevel = "public_manifest" // 公开清单
	SkillFileAccessLevelRuntimeRead    SkillFileAccessLevel = "runtime_read"    // 运行时读取
	SkillFileAccessLevelRestricted     SkillFileAccessLevel = "restricted"      // 受限访问
)

// RegisterSkillReq 注册 Skill 请求
type RegisterSkillReq struct {
	BusinessDomainID string          `header:"x-business-domain" validate:"required"`
	UserID           string          `header:"user_id" validate:"required"`
	FileType         string          `form:"file_type" validate:"required,oneof=zip content"`
	File             json.RawMessage `form:"file" validate:"required"`
	Source           string          `form:"source"`
	ExtendInfo       json.RawMessage `form:"extend_info"`
}

// RegisterSkillResp 注册 Skill 响应
type RegisterSkillResp struct {
	SkillID     string            `json:"skill_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Version     string            `json:"version"`
	Status      model.SkillStatus `json:"status"`
	Files       []string          `json:"files"`
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
	BusinessDomainID string            `header:"x-business-domain" validate:"required"`
	UserID           string            `header:"user_id"`
	Name             string            `form:"name"`
	Status           model.SkillStatus `form:"status" validate:"omitempty,oneof=draft active error deleting"`
	Source           string            `form:"source"`
	CreateUser       string            `form:"create_user"`
	CommonPageParams `json:",inline"`
}

// SkillInfo Skill 详情
type SkillInfo struct {
	SkillID      string                 `json:"skill_id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Version      string                 `json:"version"`
	Status       model.SkillStatus      `json:"status"`
	Source       string                 `json:"source"`
	Dependencies map[string]interface{} `json:"dependencies,omitempty"`
	ExtendInfo   map[string]interface{} `json:"extend_info,omitempty"`
	CreateUser   string                 `json:"create_user"`
	CreateTime   int64                  `json:"create_time"`
	UpdateUser   string                 `json:"update_user"`
	UpdateTime   int64                  `json:"update_time"`
}

// SkillSummary Skill 列表摘要
type SkillSummary struct {
	SkillID     string            `json:"skill_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Version     string            `json:"version"`
	Status      model.SkillStatus `json:"status"`
	Source      string            `json:"source"`
	CreateUser  string            `json:"create_user"`
	CreateTime  int64             `json:"create_time"`
	UpdateUser  string            `json:"update_user"`
	UpdateTime  int64             `json:"update_time"`
}

// SkillFileSummary Skill 文件摘要
type SkillFileSummary struct {
	RelPath     string               `json:"rel_path"`
	FileType    string               `json:"file_type"`
	AccessLevel SkillFileAccessLevel `json:"access_level"`
	Size        int64                `json:"size"`
	MimeType    string               `json:"mime_type"`
}

// QuerySkillListResp Skill 列表响应
type QuerySkillListResp struct {
	CommonPageResult `json:",inline"`
	Data             []*SkillSummary `json:"data"`
}

// QuerySkillMarketListReq Skill 市场列表查询
type QuerySkillMarketListReq struct {
	BusinessDomainID string `header:"x-business-domain" validate:"required"`
	UserID           string `header:"user_id"`
	Name             string `form:"name"`
	Source           string `form:"source"`
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
	SkillID      string              `json:"skill_id"`
	SkillContent string              `json:"skill_content"`
	Files        []*SkillFileSummary `json:"files"`
	Status       model.SkillStatus   `json:"status"`
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
	SkillID     string `json:"skill_id"`
	RelPath     string `json:"rel_path"`
	Content     string `json:"content"`
	MimeType    string `json:"mime_type"`
	FileType    string `json:"file_type"`
	AccessLevel string `json:"access_level"`
}

// AgentSkillBindingReq 查询 Agent 绑定的 Skill 请求
type AgentSkillBindingReq struct {
	BusinessDomainID string `header:"x-business-domain" validate:"required"`
	UserID           string `header:"user_id"`
	AgentID          string `uri:"agent_id" validate:"required"`
}

// AgentSkillBindingResp Agent Skill 绑定结果
type AgentSkillBindingResp struct {
	AgentID string       `json:"agent_id"`
	Skills  []*SkillInfo `json:"skills"`
}

// SkillRegistry Skill 管理接口
type SkillRegistry interface {
	RegisterSkill(ctx context.Context, req *RegisterSkillReq) (*RegisterSkillResp, error)
	DeleteSkill(ctx context.Context, req *DeleteSkillReq) error
	DownloadSkill(ctx context.Context, req *DownloadSkillReq) (*DownloadSkillResp, error)
	QuerySkillList(ctx context.Context, req *QuerySkillListReq) (*QuerySkillListResp, error)
	GetSkillDetail(ctx context.Context, req *GetSkillDetailReq) (*SkillInfo, error)
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

// SkillRuntimeBinder Skill 运行时绑定接口
type SkillRuntimeBinder interface {
	ListAgentSkills(ctx context.Context, req *AgentSkillBindingReq) (*AgentSkillBindingResp, error)
}
