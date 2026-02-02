// Package interfaces defines entities, DTOs, and service interfaces.
package interfaces

const (
	ConnectorModeLocal  string = "local"  // 内置在 vega-manager 进程内运行
	ConnectorModeRemote string = "remote" // 作为独立服务运行，通过 HTTP 调用
)

const (
	ConnectorCategoryTable   string = "table"   // 关系型数据库
	ConnectorCategoryIndex   string = "index"   // 搜索引擎
	ConnectorCategoryTopic   string = "topic"   // 消息队列
	ConnectorCategoryFile    string = "file"    // 文件
	ConnectorCategoryFileset string = "fileset" // 文件集
	ConnectorCategoryMetric  string = "metric"  // 时序数据库
	ConnectorCategoryAPI     string = "api"     // API 服务
)

// ConnectorType 表示一个已注册的 connector 类型
type ConnectorType struct {
	Type     string `json:"type"`
	Name     string `json:"name"`     // mysql, postgresql, kafka...
	Comment  string `json:"comment"`  // 类型描述
	Mode     string `json:"mode"`     // local | remote
	Category string `json:"category"` // table | index | topic | file | fileset | metric | api
	Endpoint string `json:"endpoint"` // 仅 remote 模式，远程服务地址
	Enabled  bool   `json:"enabled"`  // 是否启用
}

// ConnectorTypesQueryParams 查询参数
type ConnectorTypesQueryParams struct {
	PaginationParams
	Mode     string // 按模式筛选
	Category string // 按分类筛选
	Enabled  *bool  // 按启用状态筛选
}

// ConnectorType 表示一个已注册的 connector 类型
type ConnectorTypeReq struct {
	Type     string `json:"type"`
	Name     string `json:"name"`     // mysql, postgresql, kafka...
	Comment  string `json:"comment"`  // 类型描述
	Mode     string `json:"mode"`     // local | remote
	Category string `json:"category"` // table | index | topic | file | fileset | metric | api
	Endpoint string `json:"endpoint"` // 仅 remote 模式，远程服务地址
	Enabled  bool   `json:"enabled"`  // 是否启用

	OriginConnectorType *ConnectorType `json:"-"`
}
