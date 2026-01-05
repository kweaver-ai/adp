// Package interfaces 定义业务知识网络行动召回接口
package interfaces

import "context"

// ==================== 常量定义 ====================

const (
	// ResultProcessStrategyKnActionRecall 业务知识网络行动召回结果处理策略
	ResultProcessStrategyKnActionRecall = "kn_action_recall"
)

// ActionSource 类型常量
const (
	// ActionSourceTypeTool 工具类型的行动源
	ActionSourceTypeTool = "tool"
	// ActionSourceTypeMCP MCP 类型的行动源（下个版本支持）
	ActionSourceTypeMCP = "mcp"
)

// ==================== 请求和响应结构 ====================

// KnActionRecallRequest 业务知识网络行动召回请求
type KnActionRecallRequest struct {
	// Query Parameters
	KnID string `json:"kn_id" validate:"required"` // 业务知识网络ID
	AtID string `json:"at_id" validate:"required"` // 行动类型ID

	// Request Body
	UniqueIdentity map[string]interface{} `json:"unique_identity" validate:"required,min=1"` // 对象唯一标识

	// Header 字段
	AccountID   string `json:"-" header:"x-account-id" validate:"required"`
	AccountType string `json:"-" header:"x-account-type" validate:"required"`
}

// KnActionRecallResponse 业务知识网络行动召回响应
type KnActionRecallResponse struct {
	Headers      map[string]string `json:"headers"` // HTTP Header 参数
	DynamicTools []KnDynamicTool   `json:"_dynamic_tools"`
}

// KnDynamicTool 动态工具定义
type KnDynamicTool struct {
	Name            string                 `json:"name"`              // 工具名称
	Description     string                 `json:"description"`       // 工具描述
	Parameters      map[string]interface{} `json:"parameters"`        // OpenAI Function Call Schema
	ApiURL          string                 `json:"api_url"`           // 工具执行代理URL
	OriginalSchema  map[string]interface{} `json:"original_schema"`   // 原始 OpenAPI 定义
	FixedParams     KnFixedParams          `json:"fixed_params"`      // 固定参数
	ApiCallStrategy string                 `json:"api_call_strategy"` // 结果处理策略， 固定值为 kn_action_recall
}

// KnFixedParams 固定参数结构
type KnFixedParams struct {
	Header map[string]interface{} `json:"header"` // HTTP Header 参数
	Path   map[string]interface{} `json:"path"`   // URL Path 参数
	Query  map[string]interface{} `json:"query"`  // URL Query 参数
	Body   map[string]interface{} `json:"body"`   // Request Body 参数
}

// ==================== 行动查询相关结构 ====================

// QueryActionsRequest 行动查询请求
type QueryActionsRequest struct {
	KnID                string                   `json:"kn_id"`
	AtID                string                   `json:"at_id"`
	UniqueIdentities    []map[string]interface{} `json:"unique_identities"`
	IncludeTypeInfo     bool                     `json:"include_type_info"`
	XHTTPMethodOverride string                   `json:"-"` // 固定为 GET
}

// QueryActionsResponse 行动查询响应
type QueryActionsResponse struct {
	ActionType   *ActionTypeInfo `json:"action_type,omitempty"` // 行动类信息
	ActionSource *ActionSource   `json:"action_source"`         // 行动源
	Actions      []ActionParams  `json:"actions"`               // 行动参数列表
	TotalCount   int             `json:"total_count"`
	OverallMs    int             `json:"overall_ms"`
}

// ActionTypeInfo 行动类信息
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

// ActionTypeParam 行动类型参数
type ActionTypeParam struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Source    string `json:"source"`
	ValueFrom string `json:"value_from"` // property/input/const
	Value     string `json:"value,omitempty"`
}

// ActionSource 行动源
type ActionSource struct {
	Type   string `json:"type"`    // tool/mcp
	BoxID  string `json:"box_id"`  // 工具箱ID
	ToolID string `json:"tool_id"` // 工具ID
}

// ActionParams 行动参数
type ActionParams struct {
	Parameters    map[string]interface{} `json:"parameters"`     // 已实例化的参数
	DynamicParams map[string]interface{} `json:"dynamic_params"` // 动态参数（值为null）
}

// ==================== Service 接口 ====================

// IKnActionRecallService 业务知识网络行动召回服务接口
type IKnActionRecallService interface {
	// GetActionInfo 获取行动信息（行动召回）
	GetActionInfo(ctx context.Context, req *KnActionRecallRequest) (*KnActionRecallResponse, error)
}
