# context-loader 工具使用指南（v5.0.4）

面向开发人员的 context-loader 工具集说明文档，用于理解 context-loader 的定位、能力边界，以及在 Agent/服务中如何调用这些接口。

## 文档信息

| 字段 | 值 |
| :--- | :--- |
| 文档版本 | v1.0 |
| 适用版本 | context-loader v5.0.4（internal 5004） |
| 发布日期 | 2026-01-04 |
| 状态 | 正式发布 |

| 修订日期 | 修订说明 |
| :--- | :--- |
| 2026-01-04 | 首次发布 |

## 1. 什么是 context-loader

### 1.1 定位

context-loader 的目标不是直接回答用户问题，而是为 Agent 提供来自 BKN（业务知识网络）的高质量、最小且完备的上下文子集，让最终回答尽可能基于事实、降低幻觉。

### 1.2 双模检索（Dual-Mode Retrieval）

- 探索发现模式（Exploratory Search）：解决概念盲区，快速“点亮地图”（实例语义召回为后续规划）
- 精确结构化查询模式（Structural Query）：按图索骥，用确定性查询获取可审计的事实

## 2. 能力边界

| 维度 | context-loader 负责 | context-loader 不负责 |
| :--- | :--- | :--- |
| 意图与推理 | 提供可用的检索工具与稳定输出 | 用户意图理解、规划与复杂推理 |
| 数据获取 | Schema 检索、实例检索、结构化查询、逻辑属性解析 | 最终自然语言答案生成 |
| 可靠性 | 提供确定性的结构化查询原子能力 | “自动把所有参数都推断出来”的完全自治 |

## 3. 快速开始

### 3.1 服务地址

默认服务地址：

```
http://agent-retrieval:30779
```

### 3.2 认证与通用 Header

多数接口要求在 Header 中携带：

| Header | 必填 | 说明 |
| :--- | :--- | :--- |
| `x-account-id` | 是（以接口定义为准） | 账户 ID |
| `x-account-type` | 是（以接口定义为准） | 账户类型（如 user/app/system/anonymous） |

### 3.3 最小调用示例：先查概念，再找入口实例

1）查概念（Schema，推荐 v2：kn_search）：

```bash
curl -X POST "http://agent-retrieval:30779/api/agent-retrieval/in/v1/kn/kn_search" \
  -H "Content-Type: application/json" \
  -H "x-account-id: <your-account-id>" \
  -H "x-account-type: user" \
  -d '{
    "kn_id": "kn_medical",
    "query": "头晕吃什么药",
    "only_schema": true
  }'
```

2）精确查询实例（用 query_object_instance 定位入口实例）：

```bash
curl -X POST "http://agent-retrieval:30779/api/agent-retrieval/in/v1/kn/query_object_instance?kn_id=kn_medical&ot_id=disease" \
  -H "Content-Type: application/json" \
  -H "x-account-id: <your-account-id>" \
  -H "x-account-type: user" \
  -d '{
    "limit": 10,
    "condition": {
      "operation": "and",
      "sub_conditions": [
        { "field": "name", "operation": "like", "value_from": "const", "value": "高血压" }
      ]
    }
  }'
```

## 4. 用法概览（如何选择工具）

### 4.1 典型调用链

```
用户问题
  └─ Agent 规划
      ├─ 探索发现：kn_schema_search / kn_search
      ├─ 精确查询：query_object_instance / query_instance_subgraph
      ├─ 逻辑属性：get_logic_properties_values（需要动态参数时）
      └─ 行动召回：get_action_info（需要动态工具发现时）
```

### 4.2 工具总览

| 工具 | 核心作用 | 何时用 |
| :--- | :--- | :--- |
| `kn_schema_search` | 语义检索概念（Schema） | 不确定有哪些对象类/关系类/动作类时 |
| `kn_search` | 概念召回（Schema，v2） | 需要更强的概念召回控制能力（多轮/精简Schema等） |
| `query_object_instance` | 单对象类实例过滤查询 | 已知对象类与过滤条件，要查列表时 |
| `query_instance_subgraph` | 沿关系路径拉取子图 | 需要跨关系找关联对象/多跳事实时 |
| `get_logic_properties_values` | 逻辑属性解析（指标/算子） | 值需要按上下文动态计算时 |
| `get_action_info` | 动态工具发现（Function Call 定义） | 针对具体对象实例，想知道“能做什么动作”时 |

### 4.3 工具依赖（双模检索如何衔接）

- 精确查询依赖探索发现：
  - 结构化查询需要 `ot_id`（对象类 ID）与 Schema 信息（字段/主键/关系方向/动作绑定对象类）
  - `ot_id` 与 Schema 通常来自 `kn_schema_search` 或 `kn_search` 的返回（object_types / relation_types / action_types）
- 逻辑属性与行动召回依赖 Schema + 精确查询数据：
  - `get_logic_properties_values` 需要 `ot_id` + `unique_identities`（来自 `query_object_instance`/`query_instance_subgraph` 的实例结果），以及逻辑属性定义（来自 Schema）
  - `get_action_info` 需要 `at_id`（来自 Schema）+ `unique_identity`（来自精确查询的实例主键/唯一标识）

## 5. 实现原理概览

### 5.1 Schema 与 Data 分层

- Schema：对象类/关系类/动作类定义（用于让 Agent“理解世界结构”）
- Data：对象实例与关联事实（用于让 Agent“拿到确定性证据”）

### 5.2 为什么要“先探索，再结构化”

- 探索发现用于降低盲区：先找到候选概念/入口实例
- 结构化查询用于保证可追溯：推理链的每一步都有确定输入与确定输出

## 6. 工具参考（Tool Reference）

本节仅给出开发接入时最常用的信息：用途、关键参数与最小示例。完整字段与响应结构以本目录下对应的 OpenAPI YAML 文件为准。

### 6.1 kn_schema_search（语义检索 / 概念召回，v1）

> 接口定义：[toolset/openapi/kn_schema_search.yaml](./toolset/openapi/kn_schema_search.yaml)

- API：`POST /api/agent-retrieval/in/v1/kn/semantic-search`
- 作用：根据 query 返回与之相关的概念信息（Schema）

请求体（关键字段）：

| 字段 | 必填 | 说明 |
| :--- | :--- | :--- |
| `kn_id` | 是 | 业务知识网络 ID |
| `query` | 是 | 用户自然语言查询 |
| `search_scope` | 否 | 限定概念分组、是否包含对象类/关系类/行动类 |
| `max_concepts` | 否 | 最大概念数量（默认 10） |
| `rerank_action` | 否 | 重排策略（default/vector/llm） |
| `return_query_understanding` | 否 | 是否返回 query_understanding |

返回要点：

- `concepts`：相关概念列表
- `query_understanding`：可选，包含意图拆解与策略信息

Data Agent 配置（建议）：

| 配置项 | 推荐类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| `x-account-id` | 应用变量 | Header 参数 | `header.x-account-id` |
| `x-account-type` | 固定值/应用变量 | Header 参数 | `user` 或 `header.x-account-type` |
| `kn_id` | 固定值/应用变量 | 知识网络 ID | `"kn_medical"` 或 `self_config.data_source.knowledge_network[0].knowledge_network_id` |
| `query` | 模型生成 | 用户问题/关键词 | `模型生成` |
| `previous_queries` | 模型生成 | 请求体参数（可选） | `模型生成` |
| `search_scope` | 模型生成 | 请求体参数（可选） | `模型生成` |
| `max_concepts` | 固定值 | 最大概念数 | `10` |
| `rerank_action` | 固定值/模型生成 | 请求体参数（可选） | `default` |
| `return_query_understanding` | 固定值 | 请求体参数（可选） | `false` |

### 6.2 kn_search（知识网络检索 / 概念召回，v2）

> 接口定义：[toolset/openapi/kn_search.yaml](./toolset/openapi/kn_search.yaml)

- API：`POST /api/agent-retrieval/in/v1/kn/kn_search`
- 作用：返回 Schema（object_types / relation_types / action_types）
- 说明：本指南暂时只覆盖“概念召回”。语义实例召回作为后续规划；需要实例请优先使用 `query_object_instance` / `query_instance_subgraph`。

请求体（关键字段）：

| 字段 | 必填 | 说明 |
| :--- | :--- | :--- |
| `kn_id` | 是 | 业务知识网络 ID |
| `query` | 是 | 问题或关键词（多关键词用空格分隔） |
| `only_schema` | 否 | 建议设置为 true（仅概念召回） |
| `session_id` | 否 | 多轮会话 ID（用于维护历史召回） |
| `additional_context` | 否 | 二次检索补充上下文 |
| `retrieval_config` | 否 | 召回配置（概念召回精简、并集策略等） |
| `enable_rerank` | 否 | 是否启用重排序（默认 true） |

返回要点：

- `object_types / relation_types / action_types`：概念列表（Schema）

Data Agent 配置（建议）：

| 配置项 | 推荐类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| `session_id` | 应用变量 | 会话 ID（多轮时用于历史召回） | `self_config.conversation_id` |
| `x-account-id` | 应用变量 | Header 参数（接口定义为可选） | `header.x-account-id` |
| `x-account-type` | 固定值/应用变量 | Header 参数（接口定义为可选） | `user` 或 `header.x-account-type` |
| `kn_id` | 固定值/应用变量 | 知识网络 ID | `"kn_medical"` 或 `self_config.data_source.knowledge_network[0].knowledge_network_id` |
| `query` | 模型生成 | 用户问题/关键词 | `模型生成` |
| `only_schema` | 固定值 | 仅概念召回 | `true` |
| `retrieval_config.concept_retrieval.schema_brief` | 固定值 | 精简 Schema | `true` |
| `retrieval_config.concept_retrieval.return_union` | 固定值 | 请求体参数（可选） | `false` |
| `additional_context` | 模型生成 | 请求体参数（可选） | `模型生成` |
| `enable_rerank` | 固定值 | 请求体参数（可选） | `true` |

### 6.3 query_object_instance（对象实例查询）

> 接口定义：[toolset/openapi/query_object_instance.yaml](./toolset/openapi/query_object_instance.yaml)

- API：`POST /api/agent-retrieval/in/v1/kn/query_object_instance`
- Query 参数：
  - `kn_id`（必填）：业务知识网络 ID
  - `ot_id`（必填）：对象类 ID
  - `include_type_info`（可选）：是否包含对象类信息（默认 false）
  - `include_logic_params`（可选）：是否返回逻辑属性计算参数（默认 false）
- 作用：在指定对象类内，按过滤条件查询实例列表（支持分页）

请求体（FirstQueryWithSearchAfter，关键字段）：

| 字段 | 必填 | 说明 |
| :--- | :--- | :--- |
| `limit` | 是 | 返回数量（默认 10，范围以接口定义为准） |
| `condition` | 否 | 过滤条件（支持 and/or/比较/集合/like/match 等） |
| `sort` | 否 | 排序字段列表 |
| `need_total` | 否 | 是否返回总数 |
| `properties` | 否 | 指定返回的属性字段列表 |

Condition 规则要点：

- `value_from` 与 `value` 必须同时出现
- `value_from` 当前仅支持 `"const"`

示例：

```bash
curl -X POST "http://agent-retrieval:30779/api/agent-retrieval/in/v1/kn/query_object_instance?kn_id=kn_medical&ot_id=drug" \
  -H "Content-Type: application/json" \
  -H "x-account-id: <your-account-id>" \
  -H "x-account-type: user" \
  -d '{
    "limit": 10,
    "condition": {
      "operation": "and",
      "sub_conditions": [
        { "field": "name", "operation": "like", "value_from": "const", "value": "阿司匹林" }
      ]
    }
  }'
```

Data Agent 配置（建议）：

| 配置项 | 推荐类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| `x-account-id` | 应用变量 | Header 参数 | `header.x-account-id` |
| `x-account-type` | 固定值/应用变量 | Header 参数 | `user` 或 `header.x-account-type` |
| `kn_id` | 固定值/应用变量 | Query 参数 | `"kn_medical"` 或 `self_config.data_source.knowledge_network[0].knowledge_network_id` |
| `ot_id` | 模型生成 | Query 参数（对象类 ID） | `模型生成` |
| `include_type_info` | 固定值 | Query 参数（可选） | `false` |
| `include_logic_params` | 固定值 | Query 参数（可选） | `false` |
| `limit` | 固定值 | 请求体参数 | `10` |
| `condition` | 模型生成 | 请求体参数（可选） | `模型生成` |
| `sort` | 模型生成 | 请求体参数（可选） | `模型生成` |
| `need_total` | 固定值 | 请求体参数（可选） | `false` |
| `properties` | 模型生成 | 请求体参数（可选） | `模型生成` |

### 6.4 query_instance_subgraph（实例子图查询）

> 接口定义：[toolset/openapi/query_instance_subgraph.yaml](./toolset/openapi/query_instance_subgraph.yaml)

- API：`POST /api/agent-retrieval/in/v1/kn/query_instance_subgraph`
- Query 参数：
  - `kn_id`（必填）：业务知识网络 ID
  - `include_logic_params`（可选）：是否返回逻辑属性计算参数（默认 false）
- 作用：基于关系路径查询对象子图；支持多条路径，每条路径返回独立子图

使用要点：

- 请求体必须提供 `relation_type_paths`（以接口定义为准），用于描述关系路径模板
- `relation_type_paths[].object_types` 与 `relation_type_paths[].relation_types` 的数组顺序必须严格对应
- Condition 结构与 query_object_instance 一致（同样需要 value_from + value 配对）

Data Agent 配置（建议）：

| 配置项 | 推荐类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| `x-account-id` | 应用变量 | Header 参数 | `header.x-account-id` |
| `x-account-type` | 固定值/应用变量 | Header 参数 | `user` 或 `header.x-account-type` |
| `kn_id` | 固定值/应用变量 | Query 参数 | `"kn_medical"` 或 `self_config.data_source.knowledge_network[0].knowledge_network_id` |
| `include_logic_params` | 固定值 | Query 参数（可选） | `false` |
| `relation_type_paths` | 模型生成 | 请求体参数 | `模型生成` |

### 6.5 get_logic_properties_values（逻辑属性解析）

- API：`POST /api/agent-retrieval/in/v1/kn/logic-property-resolver`
- 作用：针对某对象类的一个或多个实例，批量计算/查询逻辑属性（metric/operator）

请求体（关键字段）：

| 字段 | 必填 | 说明 |
| :--- | :--- | :--- |
| `kn_id` | 是 | 业务知识网络 ID |
| `ot_id` | 是 | 对象类 ID |
| `query` | 是 | 用户原始问题（用于生成 dynamic_params） |
| `unique_identities` | 是 | 实例主键数组（支持批量） |
| `properties` | 是 | 逻辑属性名列表（metric/operator） |
| `additional_context` | 否 | 推荐传结构化 JSON 字符串，补充时间/对象上下文等 |
| `options` | 否 | 高级选项（如 return_debug、max_concurrency） |

返回形态：

- 成功：返回 `datas`
- 缺参：返回 `error_code=MISSING_INPUT_PARAMS` 与 `missing` 清单，按 hint 补充 query 或 additional_context 后重试

缺参示例（节选）：

```json
{
  "error_code": "MISSING_INPUT_PARAMS",
  "message": "dynamic_params 缺少必需的 input 参数",
  "missing": [
    {
      "property": "approved_drug_count",
      "params": [
        { "name": "start", "type": "INTEGER", "hint": "在 additional_context 中补充时间范围，或在 query 中明确时间信息" }
      ]
    }
  ],
  "trace_id": "3f5d6c1c-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

Data Agent 配置（建议）：

| 配置项 | 推荐类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| `x-account-id` | 应用变量 | Header 参数 | `header.x-account-id` |
| `x-account-type` | 固定值/应用变量 | Header 参数 | `user` 或 `header.x-account-type` |
| `kn_id` | 固定值/应用变量 | 请求体参数 | `"kn_medical"` 或 `self_config.data_source.knowledge_network[0].knowledge_network_id` |
| `ot_id` | 模型生成 | 请求体参数 | `模型生成` |
| `query` | 模型生成 | 请求体参数 | `模型生成` |
| `unique_identities` | 模型生成 | 请求体参数 | `模型生成` |
| `properties` | 模型生成 | 请求体参数 | `模型生成` |
| `additional_context` | 模型生成 | 请求体参数（可选） | `模型生成` |
| `options` | 模型生成 | 请求体参数（可选） | `模型生成` |

### 6.6 get_action_info（行动信息召回 / 动态工具发现）

- API：`POST /api/agent-retrieval/in/v1/kn/get_action_info`
- 作用：针对对象实例，召回可执行行动，并转换为 OpenAI Function Call 规范的工具定义列表

请求体：

| 字段 | 必填 | 说明 |
| :--- | :--- | :--- |
| `kn_id` | 是 | 业务知识网络 ID |
| `at_id` | 是 | 行动类型 ID |
| `unique_identity` | 是 | 对象唯一标识（键值对，至少一个字段） |

返回要点：

- `_dynamic_tools`：动态工具列表（每个工具包含 name/description/parameters/api_url/fixed_params 等）

当前版本限制：

- 仅支持 type=tool 的行动源（MCP 下版本支持）
- 仅处理 actions[0]
- 不处理 dynamic_params（由 LLM 侧生成）

Data Agent 配置（建议）：

| 配置项 | 推荐类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| `x-account-id` | 应用变量 | Header 参数 | `header.x-account-id` |
| `x-account-type` | 固定值/应用变量 | Header 参数 | `user` 或 `header.x-account-type` |
| `kn_id` | 固定值/应用变量 | 请求体参数 | `"kn_medical"` 或 `self_config.data_source.knowledge_network[0].knowledge_network_id` |
| `at_id` | 模型生成 | 请求体参数 | `模型生成` |
| `unique_identity` | 模型生成 | 请求体参数 | `模型生成` |

## 7. 集成场景与最佳实践

### 7.1 场景：从问题到可审计事实链

1）探索概念：用 `kn_schema_search` / `kn_search` 确认对象类/关系类  
2）精确定位实例：用 `query_object_instance`（单类过滤）或 `query_instance_subgraph`（跨关系/多跳）获取入口实例与事实  
3）补充指标：用 `get_logic_properties_values` 获取逻辑属性值（必要时补 additional_context）  
4）动态动作：用 `get_action_info` 获取与实例关联的可执行行动

## 8. 附录

### 8.1 本目录 OpenAPI 定义文件

- `kn_schema_search.yaml`
- `kn_search.yaml`
- `query_object_instance.yaml`
- `query_instance_subgraph.yaml`
- `get_logic_properties_values.yaml`
- `get_action_info.yaml`
