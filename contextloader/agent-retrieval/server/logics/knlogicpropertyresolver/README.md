# kn-logic-property-resolver

## 功能概述

生成逻辑属性（metric 和 operator）的 `dynamic_params` 并调用 ontology-query 查询属性值。

## 核心接口

```go
type IKnLogicPropertyResolverService interface {
    ResolveLogicProperties(ctx context.Context, req *ResolveLogicPropertiesRequest) (*ResolveLogicPropertiesResponse, error)
}
```

## 使用方式

```go
// 创建服务实例
service := knlogicpropertyresolver.NewKnLogicPropertyResolverService()

// 调用接口
resp, err := service.ResolveLogicProperties(ctx, &interfaces.ResolveLogicPropertiesRequest{
    KnID:             "kn_medical",
    OtID:             "company",
    Query:            "查询企业健康度",
    UniqueIdentities: []map[string]interface{}{{"company_id": "company_000001"}},
    Properties:       []string{"business_health_score"},
})
```

## HTTP 路由

```
POST /api/kn/logic-property-resolver
```

## 主要流程

1. 参数校验
2. 获取对象类定义（ontology-manager）
3. 提取逻辑属性定义
4. 生成 dynamic_params（调用 Agent 平台）
5. 参数校验
6. 查询逻辑属性值（ontology-query）
7. 返回结果

## 架构特点

- **并发处理**：按 property 并发生成参数（可配置并发数）
- **Agent 集成**：使用 Agent 平台生成参数（可扩展切换到直接调用 LLM）
- **容错机制**：单个 property 失败不影响其他
- **缺参支持**：返回结构化的缺参信息

## 详细设计文档

完整的需求、设计、实现文档请查看：

- **实现架构设计**：`prd/feature-799460/09-实现架构设计.md`
- **开发实施指南**：`prd/feature-799460/10-开发实施指南.md`
- **PRD 目录**：`prd/feature-799460/`

## 配置要求

在 `config.yaml` 中配置 Agent Key：

```yaml
deploy_agent:
  metric_dynamic_params_generator_key: "your_metric_agent_key"
  operator_dynamic_params_generator_key: "your_operator_agent_key"
```

## 依赖服务

- **ontology-manager**：获取对象类定义
- **ontology-query**：查询逻辑属性值
- **agent-app**：生成 dynamic_params

## 开发状态

当前已完成：
- ✅ 基础架构搭建
- ✅ Agent 平台集成
- ✅ 并发控制实现
- 🔲 参数校验（待实现）
- 🔲 Ontology Query 调用（待实现）

