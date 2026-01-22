// Package interfaces 定义接口
// @file drivenadapters.go
// @description: 入站接口定义
package interfaces

//go:generate mockgen -source=drivenadapters.go -destination=../mocks/drivenadapters.go -package=mocks
import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	// SystemUser 系统
	SystemUser = "system"
	// UnknownUser 未知
	UnknownUser = "unknown"
)

// VisitorType 访问者类型
type VisitorType string

// 访问者类型定义
const (
	RealName  VisitorType = "realname"  // 实名用户
	Anonymous VisitorType = "anonymous" // 匿名用户
	Business  VisitorType = "business"  // 应用账户
)

// ToAccessorType 转换为访问者类型
func (v VisitorType) ToAccessorType() AccessorType {
	switch v {
	case RealName:
		return AccessorTypeUser
	case Business:
		return AccessorTypeApp
	case Anonymous:
		return AccessorTypeAnonymous
	default:
		// 未知访问者类型，默认匿名用户
		return AccessorTypeAnonymous
	}
}

// AccountType 登录账号类型
type AccountType int32

// 登录账号类型定义
const (
	Other  AccountType = 0
	IDCard AccountType = 1
)

const (
	// AccessedByUser 实名用户
	AccessedByUser string = "accessed_by_users"
	// AccessedByAnyOne 匿名用户
	AccessedByAnyOne string = "accessed_by_anyone"
)

// ClientType 设备类型
type ClientType int32

// ClientTypeMap 客户端类型表
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

// ReverseClientTypeMap 客户端类型字符串反查表
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

// AccountTypeMap 账户类型表
var AccountTypeMap = map[AccountType]string{
	Other:  "other_category",
	IDCard: "id_card",
}

// ReverseAccountTypeMap 账户类型字符串反查表
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

// 设备类型定义
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

// TokenInfo 授权验证信息
type TokenInfo struct {
	Active     bool        // 令牌状态
	VisitorID  string      // 访问者ID
	Scope      string      // 权限范围
	ClientID   string      // 客户端ID
	VisitorTyp VisitorType // 访问者类型
	// 以下字段只在visitorType=realname，即实名用户时才存在
	LoginIP     string      // 登陆IP
	Udid        string      // 设备码
	AccountTyp  AccountType // 账户类型
	ClientTyp   ClientType  // 设备类型
	PhoneNumber string      // 匿名用户的电话号码
	VisitorName string      // 匿名外链，访问者的昵称
	MAC         string      // MAC地址
	UserAgent   string      // 代理信息
}

// Hydra 授权服务接口
type Hydra interface {
	Introspect(ctx context.Context, token string) (tokenInfo *TokenInfo, err error)
}

const (
	// DisplayName 用户显示名称
	DisplayName = "name"
)

// UserInfo 用户信息
type UserInfo struct {
	UserID      string   `json:"id"`    // 用户ID
	DisplayName string   `json:"name"`  // 用户显示名称
	Roles       []string `json:"roles"` // 角色
	Account     string   `json:"account"`
}

// AppInfo 应用账号信息
type AppInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  struct {
		IDs []string `json:"ids"`
	} `json:"detail"`
}

// FlowAutomationImportReq 	数据流导入请求
// @Description 用于组合算子配置（流程）导入
type FlowAutomationImportReq struct {
	Mode    string `json:"mode" validate:"oneof=create upsert"` // 导入模式
	Configs []any  `json:"configs" validate:"required"`         // 配置数据
}

// FlowAutomationExportResp 数据流导出响应
// @Description 用于组合算子配置（流程）导出
type FlowAutomationExportResp struct {
	Configs     []any    `json:"configs"`      // 配置数据
	OperatorIDs []string `json:"operator_ids"` // 依赖算子ID列表
}

type FlowAutomation interface {
	Export(ctx context.Context, dagIDs []string) (resp *FlowAutomationExportResp, err error)
	Import(ctx context.Context, req *FlowAutomationImportReq, userID string) (err error)
}

// UserManagement 用户管理接口
type UserManagement interface {
	GetAppInfo(ctx context.Context, appID string) (appInfo *AppInfo, err error)
	GetUserInfo(ctx context.Context, userID string, fields ...string) (info *UserInfo, err error)
	GetUsersInfo(ctx context.Context, userIDs []string, fields []string) (infos []*UserInfo, err error)
	GetUsersName(ctx context.Context, userIDs []string) (userMap map[string]string, err error)
}

type MCPClient interface {
	// GetInitInfo 获取初始化信息
	GetInitInfo(ctx context.Context) *mcp.InitializeResult
	// ListTools 获取工具列表
	ListTools(ctx context.Context, req mcp.ListToolsRequest) (*mcp.ListToolsResult, error)
	// CallTool 调用工具
	CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

type MCPToolConfig struct {
	ToolID      string          `json:"tool_id"`      // 工具ID
	Name        string          `json:"name"`         // 工具名称
	Description string          `json:"description"`  // 工具描述
	InputSchema json.RawMessage `json:"input_schema"` // 输入模式
}

type MCPInstanceCreateRequest struct {
	MCPID        string           `json:"mcp_id"`
	Version      int              `json:"version"`
	Name         string           `json:"name"`
	Instructions string           `json:"instructions"`
	ToolConfigs  []*MCPToolConfig `json:"tools"`
}

type MCPInstanceCreateResponse struct {
	MCPID     string `json:"mcp_id"`
	Version   int    `json:"version"`
	StreamURL string `json:"stream_url"`
	SSEURL    string `json:"sse_url"`
}

type MCPInstanceUpdateRequest struct {
	MCPServerName string           `json:"name"`
	Instructions  string           `json:"instructions"`
	ToolConfigs   []*MCPToolConfig `json:"tools"`
}

type MCPInstanceUpdateResponse struct {
	MCPID      string `json:"mcp_id"`
	MCPVersion int    `json:"version"`
	StreamURL  string `json:"stream_url"`
	SSEURL     string `json:"sse_url"`
}

// AgentOperatorApp MCP实例管理接口
type AgentOperatorApp interface {
	// 创建MCP实例
	CreateMCPInstance(ctx context.Context, req *MCPInstanceCreateRequest) (*MCPInstanceCreateResponse, error)
	// 删除MCP实例
	DeleteMCPInstance(ctx context.Context, mcpID string, mcpVersion int) error
	// 更新MCP实例
	UpdateMCPInstance(ctx context.Context, mcpID string, mcpVersion int, req *MCPInstanceUpdateRequest) (*MCPInstanceUpdateResponse, error)
	// 删除该MCP所有实例
	DeleteAllMCPInstances(ctx context.Context, mcpID string) error
	// 升级MCP实例
	UpgradeMCPInstance(ctx context.Context, req *MCPInstanceCreateRequest) (*MCPInstanceCreateResponse, error)
}

// AccessorType 访问类型
type AccessorType string

const (
	AccessorTypeUser       AccessorType = "user"       // 实名用户
	AccessorTypeDepartment AccessorType = "department" // 部门
	AccessorTypeGroup      AccessorType = "group"      // 组织
	AccessorTypeRole       AccessorType = "role"       // 角色
	AccessorTypeApp        AccessorType = "app"        // 应用账户
	AccessorTypeAnonymous  AccessorType = "anonymous"  // 匿名访问
)

// ToVisitorType 将AccessorType转换为VisitorType
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

const (
	// AccessorRootDepartmentID 根部门ID
	AccessorRootDepartmentID string = "00000000-0000-0000-0000-000000000000"
)

// AuthMethod 授权方法
type AuthMethod = string

// 支持的授权方法
const (
	AuthMethodGet    AuthMethod = "GET"
	AuthMethodDelete AuthMethod = "DELETE"
)

// AuthAccessor 访问者信息
type AuthAccessor struct {
	ID   string       `json:"id"`   // 唯一标识ID
	Type AccessorType `json:"type"` // 访问类型
	Name string       `json:"name"` // 访问者名称
}

// AuthResource 资源信息
type AuthResource struct {
	ID   string `json:"id"`   // 唯一标识ID
	Type string `json:"type"` // 资源类型
	Name string `json:"name"` // 资源名称
}

// AuthOperationCheckRequest 操作检查请求
type AuthOperationCheckRequest struct {
	Accessor  *AuthAccessor       `json:"accessor"`  // 访问者信息
	Resource  *AuthResource       `json:"resource"`  // 资源信息
	Operation []AuthOperationType `json:"operation"` // 检查的操作
	Method    string              `json:"method"`    // 方法
}

// AuthOperationCheckResponse 操作检查响应
type AuthOperationCheckResponse struct {
	Result bool `json:"result"` // 检查结果
}

// ResourceListRequest 资源列举请求
type ResourceListRequest struct {
	Accessor  *AuthAccessor       `json:"accessor"`  // 访问者信息
	Resource  *AuthResource       `json:"resource"`  // 资源信息
	Operation []AuthOperationType `json:"operation"` // 检查的操作
	Method    string              `json:"method"`    // 方法
}

// AuthResourceFilterRequest 资源过滤请求
type AuthResourceFilterRequest struct {
	Accessor   *AuthAccessor       `json:"accessor"`  // 访问者信息
	Resources  []*AuthResource     `json:"resources"` // 资源列表
	Operations []AuthOperationType `json:"operation"` // 检查的操作列表
	Method     string              `json:"method"`    // 方法
}

type AuthOperation struct {
	ID   string `json:"id"`   // 唯一标识ID
	Name string `json:"name"` // 操作名称
}

type PolicyOperation struct {
	Allow []*AuthOperation `json:"allow"` // 允许的操作
	Deny  []*AuthOperation `json:"deny"`  // 拒绝的操作
}

// AuthCreatePolicyRequest 新建策略请求
type AuthCreatePolicyRequest struct {
	Accessor  *AuthAccessor    `json:"accessor"`             // 访问者信息
	Resource  *AuthResource    `json:"resource"`             // 资源信息
	Operation *PolicyOperation `json:"operation"`            // 策略操作
	Condition string           `json:"condition,omitempty"`  // 条件
	ExpiresAt string           `json:"expires_at,omitempty"` // 到期时间(秒级)，RFC3339格式，UNIX TIME时间纪元(1970-01-01T08:00:00+08:00)表示永久有效
}

// AuthDeletePolicyRequest 删除策略请求
type AuthDeletePolicyRequest struct {
	Method    string          `json:"method"`    // 方法
	Resources []*AuthResource `json:"resources"` // 资源列表
}

// AuthResourceResult 资源结果
type AuthResourceResult struct {
	ID string `json:"id"` // 唯一标识ID
}

// Authorization 授权服务接口
type Authorization interface {
	// 单个决策
	OperationCheck(ctx context.Context, req *AuthOperationCheckRequest) (*AuthOperationCheckResponse, error)
	// 资源过滤
	ResourceFilter(ctx context.Context, req *AuthResourceFilterRequest) ([]*AuthResourceResult, error)
	// 资源列举
	ResourceList(ctx context.Context, req *ResourceListRequest) ([]*AuthResourceResult, error)
	// 新建策略
	CreatePolicy(ctx context.Context, req []*AuthCreatePolicyRequest) error
	// 策略删除
	DeletePolicy(ctx context.Context, req *AuthDeletePolicyRequest) error
}

// AuditLogModel 审计日志模型
type AuditLogModel struct {
	Operation   string               `json:"operation" validate:"required"`          // 操作类型
	Description string               `json:"description" validate:"required"`        // 字符串描述，最大长度65,535
	OpTime      int64                `json:"op_time" validate:"required"`            // 操作时间（通过mq上报的必需）精确到纳秒
	Operator    AuditLogOperatorInfo `json:"operator" validate:"required"`           // 操作者信息
	Object      AuditLogObjectInfo   `json:"object,omitempty"`                       // 操作对象信息
	LogFrom     LogFrom              `json:"log_from" validate:"required"`           // 日志来源
	Detail      interface{}          `json:"detail,omitempty"`                       // 细节
	ExMsg       string               `json:"ex_msg,omitempty"`                       // 附加信息，最大长度65,535
	Level       LoggerLevel          `json:"level" validate:"required"`              // 日志级别，默认INFO
	OutBizID    string               `json:"out_biz_id" validate:"required,max=128"` // 外部唯一业务ID，用于防抖，格式不限 最长128
	Type        AuditLogType         `json:"type" validate:"required"`               // 日志类型，最大长度128
}

// LogFrom 日志来源
type LogFrom struct {
	Package string      `json:"package" validate:"required"` // 大包名
	Service ServiceInfo `json:"service" validate:"required"` // 服务信息
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name string `json:"name" validate:"required"` // 服务名称
}

// LoggerLevel 日志级别
type LoggerLevel string

const (
	// LoggerLevelInfo 信息
	LoggerLevelInfo LoggerLevel = "INFO"
	// LoggerLevelWarn 警告
	LoggerLevelWarn LoggerLevel = "WARN"
)

// AuditLogObjectInfo 操作对象信息
type AuditLogObjectInfo struct {
	Type string `json:"type" validate:"required"` // 操作对象类型
	Name string `json:"name"`                     // 操作对象名称，最大长度128
	ID   string `json:"id"`                       // 操作对象ID，最大长度40
}

// AuditLogOperatoAgent 操作者代理信息
type AuditLogOperatoAgent struct {
	Type string `json:"type" validate:"required"` // 操作者客户端类型
	IP   string `json:"ip" validate:"required"`   // 操作者设备IP
	MAC  string `json:"mcp" validate:"required"`  // 操作者设备mac地址
}

// AuditLogOperatorInfo 操作者信息
type AuditLogOperatorInfo struct {
	ID    string               `json:"id" validate:"required,max=40"`    // 操作者ID，最大长度40
	Name  string               `json:"name" validate:"required,max=128"` // 操作者名称，以传入数据为准，最大长度128,type为internal_service必传
	Type  AuditLogOperatorType `json:"type" validate:"required"`         // 操作者类型
	Agent AuditLogOperatoAgent `json:"agent" validate:"required"`        // 操作者代理信息
}

// AuditLogOperatorType 操作者类型
type AuditLogOperatorType string

const (
	// AuthenticatedUser 实名用户
	AuthenticatedUser AuditLogOperatorType = "authenticated_user"
	// AnonymousUser 匿名用户
	AnonymousUser AuditLogOperatorType = "anonymous_user"
	// AppUser 应用账户
	AppUser AuditLogOperatorType = "app"
	// InternalService 内部服务
	InternalService AuditLogOperatorType = "internal_service"
)

// AuditLogOperationType 审计日志操作类型
type AuditLogOperationType string

const (
	// AuditLogOperationTypeCreate 新建
	AuditLogOperationTypeCreate AuditLogOperationType = "create"
	// AuditLogOperationTypeDelete 删除
	AuditLogOperationTypeDelete AuditLogOperationType = "delete"
	// AuditLogOperationTypeModify 编辑
	AuditLogOperationTypeModify AuditLogOperationType = "modify"
	// AuditLogOperationTypePublish 发布
	AuditLogOperationTypePublish AuditLogOperationType = "publish"
	// AuditLogOperationTypeUnpublish 下架
	AuditLogOperationTypeUnpublish AuditLogOperationType = "unpublish"
	// AuditLogOperationTypeExecute 执行
	AuditLogOperationTypeExecute AuditLogOperationType = "execute"
)

// AuditLogType 日志类型
type AuditLogType string

const (
	// AuditLogOperation 操作日志
	AuditLogOperation AuditLogType = "operation" // 操作日志
)

// BusinessDomainResource 业务域资源信息
type BusinessDomainResource struct {
	BDID string `json:"bd_id"` // 业务域ID
	ID   string `json:"id"`    // 资源ID
	Type string `json:"type"`  // 资源类型
}

// BusinessDomainResourceListRequest 业务域资源列表查询请求
type BusinessDomainResourceListRequest struct {
	BDID   string `json:"bd_id"`  // 业务域ID
	ID     string `json:"id"`     // 资源ID
	Type   string `json:"type"`   // 资源类型
	Limit  int    `json:"limit"`  // 数据量，默认：20，-1代表不进行分页，全量查询
	Offset int    `json:"offset"` // 数据偏移量，默认0
}

// BusinessDomainResourceListResponse 业务域资源列表查询响应
type BusinessDomainResourceListResponse struct {
	Limit  int                       `json:"limit"`  // 数据量
	Offset int                       `json:"offset"` // 数据偏移量
	Total  int                       `json:"total"`  // 数据总数
	Items  []*BusinessDomainResource `json:"items"`  // 数据内容
}

// BusinessDomainResourceAssociateRequest 业务域资源关联请求
type BusinessDomainResourceAssociateRequest struct {
	BDID string `json:"bd_id"` // 业务域ID
	ID   string `json:"id"`    // 资源ID
	Type string `json:"type"`  // 资源类型
}

// BusinessDomainResourceDisassociateRequest 业务域资源取消关联请求
type BusinessDomainResourceDisassociateRequest struct {
	BDID string `json:"bd_id"` // 业务域ID
	ID   string `json:"id"`    // 资源ID
	Type string `json:"type"`  // 资源类型
}

// BusinessDomainManagement 业务域管理服务接口
type BusinessDomainManagement interface {
	// 资源关联
	AssociateResource(ctx context.Context, req *BusinessDomainResourceAssociateRequest) error
	// 资源取消关联
	DisassociateResource(ctx context.Context, req *BusinessDomainResourceDisassociateRequest) error
	// 资源列表查询
	ResourceList(ctx context.Context, req *BusinessDomainResourceListRequest) (*BusinessDomainResourceListResponse, error)
}

// ExecuteCodeReq 执行代码请求
type ExecuteCodeReq struct {
	HandlerCode    string         `json:"handler_code" validate:"required"` // 执行代码
	Event          map[string]any `json:"event" validate:"required"`        // 事件
	ExecuteContext ExecuteContext `json:"context"`                          // 执行上下文
}

// ExecuteContext 执行上下文
type ExecuteContext struct {
	FunctionName          string `json:"function_name"`            // 函数名称
	FunctionVersion       string `json:"function_version"`         // 函数版本
	RemainingTimeInMillis int64  `json:"remaining_time_in_millis"` // 最大执行超时时间（毫秒）
	MemoryLimitInMB       int64  `json:"memory_limit_in_mb"`       // 内存限制，单位MB
	LogGroupName          string `json:"log_group_name"`           // 日志组名称
}

// ExecuteCodeResp 执行代码响应
type ExecuteCodeResp struct {
	Stdout  string `json:"stdout"`  // 标准输出
	Stderr  string `json:"stderr"`  // 标准错误输出
	Result  any    `json:"result"`  // 执行结果
	Metrics any    `json:"metrics"` // 执行指标
}

// SandBoxConfigReq 沙箱环境配置请求
type SandBoxConfigReq struct {
	Timeout int            `json:"timeout"` // 超时时间，单位秒
	Body    any            `json:"body"`    // 请求体
	Headers map[string]any `json:"headers"` // 请求头
	Code    string         `json:"code"`    // 执行代码
}

// SandBoxEnv 沙箱环境接口
type SandBoxEnv interface {
	// 获取沙箱服务器路由
	GetSandBoxServerRouter() *APIRouter
	// 获取沙箱环境执行请求配置
	GetSandBoxRequestConfig(ctx context.Context, req *SandBoxConfigReq) (*HTTPRequest, error)
	// 执行代码
	ExecuteCode(ctx context.Context, req *ExecuteCodeReq) (*ExecuteCodeResp, error)
}

// ChatCompletionReq 聊天完成请求
type ChatCompletionReq struct {
	Model            string                  `json:"model"`             // 模型名称，当传空字符串（""）并且不传model_id的时候代表调用默认模型（如果admin没有配置全局默认模型，调用会报错）
	Messages         []ChatCompletionMessage `json:"messages"`          // 消息列表
	Stream           bool                    `json:"stream"`            // 是否流式返回，默认false
	TopK             int                     `json:"top_k"`             // 采样池大小，仅从概率最高的前 k 个 token 中选择，k 为整数。限制生成时的候选范围。k=1 时等价于贪心搜索（完全确定）；k=50 时允许更多样性，但可能降低相关性。取值范围1~∞
	TopP             float64                 `json:"top_p"`             // 核采样，取值范围0~1，平衡生成结果的多样性和质量。值越小，输出越集中（如 0.9 仅保留概率最高的部分）；值越大，输出越随机。
	FrequencyPenalty float64                 `json:"frequency_penalty"` // 频率惩罚，降低重复出现 token 的概率，范围通常为 -2.0~2.0。抑制重复内容。正值（如 0.5）惩罚重复词；负值鼓励重复（较少使用）
	PresencePenalty  float64                 `json:"presence_penalty"`  // 存在惩罚，降低已出现过的 token 的概率，范围通常为 -2.0~2.0。鼓励生成新话题或词汇。例如设为 0.2 时，模型会避免重复使用已生成的词语。
	Temperature      float64                 `json:"temperature"`       // 控制随机性（高=创意，低=严谨），0.1 生成保守结果，1.0 更灵活0~1，部分saas模型不支持0值
	MaxTokens        int                     `json:"max_tokens"`        // 最大生成长度,，取值范围不可以超出模型最大上下文长度
	ModelID          string                  `json:"model_id"`          // 模型ID，当传空字符串（""）并且不传model的时候代表调用默认模型（如果admin没有配置全局默认模型，调用会报错）
}

// ChatCompletionResp 聊天完成响应
type ChatCompletionResp struct {
	ID      string                 `json:"id"`      // 响应ID
	Object  string                 `json:"object"`  // 对象类型，固定值"chat.completion"
	Created int64                  `json:"created"` // 创建时间戳
	Model   string                 `json:"model"`   // 模型名称
	Choices []ChatCompletionChoice `json:"choices"` // 生成结果列表
	Usage   ChatCompletionUsage    `json:"usage"`   // 消耗统计
}

type ChatCompletionChoice struct {
	Index        int                   `json:"index"`             // 结果索引
	Message      ChatCompletionMessage `json:"message,omitempty"` // 消息内容, 流式返回时为空
	Delta        ChatCompletionMessage `json:"delta,omitempty"`   // 增量消息内容, 非流式返回时为空
	FinishReason string                `json:"finish_reason"`     // 完成原因
	Flag         int                   `json:"flag"`              // 标志位
}

// ChatCompletionMessage 消息结构体
type ChatCompletionMessage struct {
	Role    string `json:"role,omitempty"`    // 角色
	Content string `json:"content,omitempty"` // 内容
}

// ChatCompletionUsage 消耗统计结构体
type ChatCompletionUsage struct {
	PromptTokens        int                       `json:"prompt_tokens"`         // 提示词 token 数
	CompletionTokens    int                       `json:"completion_tokens"`     // 完成 token 数
	TotalTokens         int                       `json:"total_tokens"`          // 总 token 数
	PromptTokensDetails ChatCompletionTokenDetail `json:"prompt_tokens_details"` // 提示词 token 数详情
}

// ChatCompletionTokenDetail 提示词 token 数详情结构体
type ChatCompletionTokenDetail struct {
	CachedTokens   int `json:"cached_tokens"`   // 缓存 token 数
	UncachedTokens int `json:"uncached_tokens"` // 未缓存 token 数
}

// MFModelAPIClient 模型管理API接口
type MFModelAPIClient interface {
	// 调用模型
	ChatCompletion(ctx context.Context, req *ChatCompletionReq) (resp *ChatCompletionResp, err error)
	// 调用模型流式返回
	StreamChatCompletion(ctx context.Context, req *ChatCompletionReq) (chan string, chan error, error)
}

// GetPromptResp 获取提示词响应
type GetPromptResp struct {
	PromptID   string `json:"prompt_id"`   // 提示词ID
	PromptName string `json:"prompt_name"` // 提示词名称
	ModelID    string `json:"model_id"`    // 模型ID
	ModelName  string `json:"model_name"`  // 模型名称
	Messages   string `json:"messages"`    // 提示词内容
}

// MFModelManager 模型管理接口
type MFModelManager interface {
	// 获取提示词
	GetPromptByPromptID(ctx context.Context, promptID string) (resp *GetPromptResp, err error)
}
