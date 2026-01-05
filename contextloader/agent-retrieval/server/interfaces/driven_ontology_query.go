package interfaces

import "context"

// 业务知识网络概念类型
type KnConceptType string

const (
	KnConceptTypeObject   KnConceptType = "object_type"   // 对象类
	KnConceptTypeRelation KnConceptType = "relation_type" // 关系类
	KnConceptTypeAction   KnConceptType = "action_type"   // 行动类
)

// QueryObjectInstancesReq 检索对象详细数据请求对象
type QueryObjectInstancesReq struct {
	KnID               string       `form:"kn_id"`                          // 知识网络ID
	OtID               string       `form:"ot_id"`                          // 对象类ID
	IncludeTypeInfo    bool         `form:"include_type_info"`              //是否包含对象类信息
	IncludeLogicParams bool         `form:"include_logic_params"`           // 包含逻辑属性的计算参数，默认false，不包含。
	Cond               *KnCondition `json:"condition"`                      // 检索条件
	Limit              int          `json:"limit" validate:"min=1,max=100"` //返回的数量，默认值 10。范围 1-100
}

type QueryObjectInstancesResp struct {
	Data          []any          `json:"datas"`       // 对象实例列表
	ObjectConcept map[string]any `json:"object_type"` // 对象类型
}

// QueryLogicPropertiesReq 查询逻辑属性值请求
type QueryLogicPropertiesReq struct {
	KnID             string                   `json:"kn_id"`
	OtID             string                   `json:"ot_id"`
	UniqueIdentities []map[string]interface{} `json:"unique_identities"`
	Properties       []string                 `json:"properties"`
	DynamicParams    map[string]interface{}   `json:"dynamic_params"`
}

// QueryLogicPropertiesResp 查询逻辑属性值响应
type QueryLogicPropertiesResp struct {
	Datas []map[string]interface{} `json:"datas"`
}

// QueryInstanceSubgraphReq 子图查询请求
type QueryInstanceSubgraphReq struct {
	// Path 参数
	KnID string `form:"kn_id"`

	// Query 参数
	IncludeLogicParams bool `form:"include_logic_params"`

	// Body 参数 - 使用 interface{} 避免明确定义结构
	// 对应 ontology-query 接口的 SubGraphQueryBaseOnTypePath 结构
	RelationTypePaths interface{} `json:"relation_type_paths"`
}

// QueryInstanceSubgraphResp 子图查询响应
type QueryInstanceSubgraphResp struct {
	// 使用 interface{} 直接返回底层接口的原始结构
	// 对应 ontology-query 接口的 PathEntries 结构
	Entries interface{} `json:"entries"`
}

// DrivenOntologyQuery 本体查询接口
type DrivenOntologyQuery interface {
	// QueryObjectInstances 检索指定对象类的对象的详细数据
	QueryObjectInstances(ctx context.Context, req *QueryObjectInstancesReq) (resp *QueryObjectInstancesResp, err error)
	// QueryLogicProperties 查询逻辑属性值
	QueryLogicProperties(ctx context.Context, req *QueryLogicPropertiesReq) (resp *QueryLogicPropertiesResp, err error)
	// QueryActions 查询行动
	QueryActions(ctx context.Context, req *QueryActionsRequest) (resp *QueryActionsResponse, err error)
	// QueryInstanceSubgraph 查询对象子图
	QueryInstanceSubgraph(ctx context.Context, req *QueryInstanceSubgraphReq) (resp *QueryInstanceSubgraphResp, err error)
}
