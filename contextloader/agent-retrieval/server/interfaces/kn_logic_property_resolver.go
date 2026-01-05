// Package interfaces 定义kn-logic-property-resolver接口
package interfaces

import "context"

// ==================== 请求和响应结构 ====================

// ResolveLogicPropertiesRequest 逻辑属性解析请求
type ResolveLogicPropertiesRequest struct {
	// 必填字段
	KnID             string                   `json:"kn_id" validate:"required"`
	OtID             string                   `json:"ot_id" validate:"required"`
	Query            string                   `json:"query"`
	UniqueIdentities []map[string]interface{} `json:"unique_identities" validate:"required"`
	Properties       []string                 `json:"properties" validate:"required"`

	// 可选字段
	AdditionalContext string          `json:"additional_context,omitempty"`
	Options           *ResolveOptions `json:"options,omitempty"`

	// Header 字段
	AccountID   string `json:"-" header:"x-account-id"`
	AccountType string `json:"-" header:"x-account-type"`
}

// ResolveOptions 可选配置
type ResolveOptions struct {
	ReturnDebug     bool `json:"return_debug" default:"false"`
	MaxRepairRounds int  `json:"max_repair_rounds" default:"1"`
	MaxConcurrency  int  `json:"max_concurrency" default:"4"`
}

// ResolveLogicPropertiesResponse 逻辑属性解析响应
type ResolveLogicPropertiesResponse struct {
	Datas []map[string]any `json:"datas"`

	// Debug 信息（仅在 return_debug=true 时返回）
	Debug *ResolveDebugInfo `json:"debug,omitempty"`
}

// ResolveDebugInfo Debug 信息（改进版）
type ResolveDebugInfo struct {
	// Agent 生成的所有动态参数
	DynamicParams map[string]any `json:"dynamic_params,omitempty"`

	// Agent 调用信息（按 property 分组）
	AgentInfo map[string]*AgentInfo `json:"agent_info,omitempty"`

	// 服务器时间戳
	NowMs int64 `json:"now_ms,omitempty"`

	// 警告信息
	Warnings []string `json:"warnings,omitempty"`

	// 追踪 ID
	TraceID string `json:"trace_id,omitempty"`
}

// AgentInfo Agent 调用信息
type AgentInfo struct {
	// 属性类型
	PropertyType string `json:"property_type"` // "metric" 或 "operator"

	// Agent 请求参数（直接存储 Agent 请求结构）
	Request AgentRequestDebugInfo `json:"request,omitempty"`

	// Agent 响应信息
	Response *AgentResponseDebugInfo `json:"response,omitempty"`
}

// AgentRequestDebugInfo Agent 请求调试信息
// 直接存储 Agent 的请求结构：MetricDynamicParamsGeneratorReq 或 OperatorDynamicParamsGeneratorReq
type AgentRequestDebugInfo any

// AgentResponseDebugInfo Agent 响应调试信息
// 直接存储 Agent 的响应信息
type AgentResponseDebugInfo struct {
	// Agent 成功响应：动态参数
	DynamicParams map[string]any `json:"dynamic_params,omitempty"`

	// Agent 失败响应：错误信息（对应 _error 字段）
	Error string `json:"_error,omitempty"`
}

// MissingParamsError 缺参错误响应（改进版）
type MissingParamsError struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`

	// 直接返回Agent生成的原始错误消息
	ErrorMsg string `json:"error_msg,omitempty"`

	// Debug信息（仅在开启debug时返回）
	Debug *ResolveDebugInfo `json:"debug,omitempty"`

	Missing []MissingPropertyParams `json:"missing"`
	TraceID string                  `json:"trace_id"`
}

// MissingPropertyParams 缺参信息（简化版）
type MissingPropertyParams struct {
	Property string `json:"property"`
	// 直接返回 Agent 生成的错误消息，不再解析具体参数信息
	ErrorMsg string `json:"error_msg,omitempty"`
}

// ==================== Service 接口 ====================

// IKnLogicPropertyResolverService 逻辑属性解析服务接口
type IKnLogicPropertyResolverService interface {
	// ResolveLogicProperties 解析逻辑属性
	ResolveLogicProperties(ctx context.Context, req *ResolveLogicPropertiesRequest) (*ResolveLogicPropertiesResponse, error)
}

// ==================== LLM 参数生成相关结构 ====================

// LLMPromptInput LLM 提示词输入
type LLMPromptInput struct {
	LogicProperty     *LogicPropertyDef `json:"logic_property"`
	Query             string            `json:"query"`
	UniqueIdentities  []map[string]any  `json:"unique_identities"`
	AdditionalContext string            `json:"additional_context,omitempty"`
	NowMs             int64             `json:"now_ms,omitempty"`
	Timezone          string            `json:"timezone,omitempty"`
}

// LLMPromptOutput LLM 提示词输出
type LLMPromptOutput struct {
	// 成功输出：{ "property_name": { "param1": value1, ... } }
	DynamicParams map[string]any `json:"-"`

	// 缺参输出：{ "_error": "..." }
	Error string `json:"_error,omitempty"`
}
