# BKN 语言规范

版本: 1.0.0
spec_version: 1.0.0

## 概述

BKN (Business Knowledge Network) 是一种基于 Markdown 的业务知识网络建模语言，用于描述业务知识网络。本文档定义了 BKN 的语法规范。

### 读者路线图（先看什么）

- **业务读者**：先看“实体定义规范 / 关系定义规范 / 行动定义规范”的示例与表格，再看“最佳实践”。
- **工程读者**：先看“增量导入规范（确定性语义）”与“校验/失败策略”，再看各类型字段。
- **Agent/大模型**：优先按“每定义一文件”组织读取，避免一次加载过多内容。

### 术语表（Glossary）

| 术语 | 含义 |
|------|------|
| BKN | Business Knowledge Network，业务知识网络 |
| knowledge_network / 网络 | 一个业务知识网络的整体集合 |
| entity | 业务对象类型（例如 Pod/Node/Service） |
| relation | 连接两个 entity 的关系类型（例如 belongs_to/routes_to） |
| action | 对 entity 执行的操作定义（可能绑定 tool/mcp） |
| data_view | 数据视图（实体/关系可直接映射的数据来源） |
| primary_key | 主键字段（用于唯一定位实例） |
| display_key | 展示字段（用于 UI/检索显示） |
| fragment | 混合片段文件（可包含多个 entity/relation/action） |
| delete | 删除标记文件（显式声明要删除的定义） |

## 文件格式

### 文件扩展名

- `.bkn` - BKN 文件

### 文件编码

- UTF-8

### 基本结构

每个 BKN 文件由两部分组成：

1. **YAML Frontmatter**: 文件元数据
2. **Markdown Body**: 知识网络定义内容

```markdown
---
type: network
id: example-network
name: 示例网络
version: 1.0.0
---

# 网络标题

网络描述...

## Entity: entity_id

实体定义...

## Relation: relation_id

关系定义...

## Action: action_id

行动定义...
```

---

## Frontmatter 规范

### 工程可控性字段（推荐）

为支持规模化协作、审批与审计，建议在定义文件中使用以下字段：

| 字段 | 适用 type | 说明 |
|------|----------|------|
| `spec_version` | all | 该文件使用的规范版本（默认继承文档 spec_version） |
| `namespace` | entity/relation/action/fragment/delete | 命名空间/包名，用于大规模组织与避免冲突（例如 `platform.k8s`） |
| `owner` | entity/relation/action/fragment/delete | 负责人/团队（用于审计与审批路由） |
| `enabled` | action | 是否启用（建议默认 `false`，导入不等于启用） |
| `risk_level` | action | 风险等级（`low|medium|high`，用于审批与发布策略） |
| `requires_approval` | action | 是否需要审批才能启用/执行 |

### 文件类型 (type)

| type | 说明 | 用途 |
|------|------|------|
| `network` | 完整知识网络 | 包含多个定义的网络文件 |
| `entity` | 单个实体定义 | 独立的实体文件，可直接导入 |
| `relation` | 单个关系定义 | 独立的关系文件，可直接导入 |
| `action` | 单个行动定义 | 独立的行动文件，可直接导入 |
| `fragment` | 混合片段 | 包含多个类型的部分定义 |
| `delete` | 删除标记 | 标记要删除的定义 |

### 网络文件 (type: network)

```yaml
---
type: network                    # 完整知识网络
id: string                       # 网络ID，唯一标识
name: string                     # 网络显示名称
version: string                  # 版本号 (semver)
tags: [string]                   # 可选，标签列表
description: string              # 可选，网络描述
includes: [string]               # 可选，引用的其他文件
---
```

### 单实体文件 (type: entity)

```yaml
---
type: entity                     # 单个实体定义
id: string                       # 实体ID，唯一标识
name: string                     # 实体显示名称
version: string                  # 可选，版本号
network: string                  # 所属网络ID（建议必填，保证导入确定性）
namespace: string                # 可选，命名空间/包名
owner: string                    # 可选，负责人/团队
tags: [string]                   # 可选，标签列表
---
```

### 单关系文件 (type: relation)

```yaml
---
type: relation                   # 单个关系定义
id: string                       # 关系ID，唯一标识
name: string                     # 关系显示名称
version: string                  # 可选，版本号
network: string                  # 所属网络ID（建议必填，保证导入确定性）
namespace: string                # 可选，命名空间/包名
owner: string                    # 可选，负责人/团队
---
```

### 单行动文件 (type: action)

```yaml
---
type: action                     # 单个行动定义
id: string                       # 行动ID，唯一标识
name: string                     # 行动显示名称
action_type: add | modify | delete  # 行动类型
version: string                  # 可选，版本号
network: string                  # 所属网络ID（建议必填，保证导入确定性）
namespace: string                # 可选，命名空间/包名
owner: string                    # 可选，负责人/团队
enabled: boolean                 # 可选，是否启用（建议默认 false）
risk_level: low | medium | high  # 可选，风险等级
requires_approval: boolean       # 可选，是否需要审批
---
```

### 混合片段 (type: fragment)

```yaml
---
type: fragment                   # 混合片段
id: string                       # 片段ID
name: string                     # 片段名称
version: string                  # 可选，版本号
network: string                  # 目标网络ID（建议必填，保证导入确定性）
namespace: string                # 可选，命名空间/包名
owner: string                    # 可选，负责人/团队
---
```

### 删除标记 (type: delete)

```yaml
---
type: delete                     # 删除标记
network: string                  # 目标网络ID（建议必填，保证导入确定性）
namespace: string                # 可选，命名空间/包名
owner: string                    # 可选，负责人/团队
targets:                         # 要删除的定义列表
  - entity: pod
  - relation: pod_belongs_node
  - action: restart_pod
---
```

---

## 实体定义规范

### 语法

```markdown
## Entity: {entity_id}

**{显示名称}** - {简短描述}

### 数据来源

| 类型 | ID | 名称 |
|------|-----|------|
| data_view | {view_id} | {view_name} |

> **主键**: `{primary_key}` | **显示属性**: `{display_key}`

### 属性覆盖

(可选) 仅声明需要特殊配置的属性

| 属性名 | 显示名 | 索引配置 | 说明 |
|--------|--------|----------|------|
| ... | ... | ... | ... |

### 逻辑属性

#### {property_name}

- **类型**: metric | operator
- **来源**: {source_id} ({source_type})
- **说明**: {description}

| 参数名 | 来源 | 绑定值 |
|--------|------|--------|
| ... | property | {property_name} |
| ... | input | - |
```

### 字段说明

| 字段 | 必须 | 说明 |
|------|:----:|------|
| entity_id | YES | 实体唯一标识，小写字母、数字、下划线 |
| 显示名称 | YES | 人类可读名称 |
| 数据来源 | YES | 映射的数据视图 |
| 主键 | YES | 主键属性名 |
| 显示属性 | YES | 用于展示的属性名 |
| 属性覆盖 | NO | 需要特殊配置的属性 |
| 逻辑属性 | NO | 指标、算子等扩展属性 |

### 配置模式

#### 模式一：完全映射（最简洁）

直接映射视图，自动继承所有字段：

```markdown
## Entity: node

**Node节点**

### 数据来源

| 类型 | ID |
|------|-----|
| data_view | view_123 |

> **主键**: `id` | **显示属性**: `node_name`
```

#### 模式二：映射 + 属性覆盖

映射视图，仅声明需要特殊配置的属性：

```markdown
## Entity: pod

**Pod实例**

### 数据来源

| 类型 | ID |
|------|-----|
| data_view | view_456 |

> **主键**: `id` | **显示属性**: `pod_name`

### 属性覆盖

| 属性名 | 索引配置 |
|--------|----------|
| pod_status | fulltext + vector |
```

#### 模式三：完整定义

完整声明所有属性：

```markdown
## Entity: service

**Service服务**

### 数据来源

| 类型 | ID |
|------|-----|
| data_view | view_789 |

### 数据属性

| 属性名 | 显示名 | 类型 | 说明 | 主键 | 索引 |
|--------|--------|------|------|:----:|:----:|
| id | ID | int64 | 主键 | YES | YES |
| service_name | 名称 | VARCHAR | 服务名 | | YES |

> **显示属性**: `service_name`
```

---

## 关系定义规范

### 语法

```markdown
## Relation: {relation_id}

**{显示名称}** - {简短描述}

| 起点 | 终点 | 类型 |
|------|------|------|
| {source_entity} | {target_entity} | direct | data_view |

### 映射规则

| 起点属性 | 终点属性 |
|----------|----------|
| {source_prop} | {target_prop} |

### 业务语义

(可选) 关系的业务含义说明...
```

### 字段说明

| 字段 | 必须 | 说明 |
|------|:----:|------|
| relation_id | YES | 关系唯一标识 |
| 起点 | YES | 起点实体 ID |
| 终点 | YES | 终点实体 ID |
| 类型 | YES | `direct` (直接映射) 或 `data_view` (视图映射) |
| 映射规则 | YES | 属性映射关系 |

### 关系类型

#### 直接映射 (direct)

通过属性值匹配建立关联：

```markdown
## Relation: pod_belongs_node

| 起点 | 终点 | 类型 |
|------|------|------|
| pod | node | direct |

### 映射规则

| 起点属性 | 终点属性 |
|----------|----------|
| pod_node_name | node_name |
```

#### 视图映射 (data_view)

通过中间视图建立关联：

```markdown
## Relation: user_likes_post

| 起点 | 终点 | 类型 |
|------|------|------|
| user | post | data_view |

### 映射视图

| 类型 | ID |
|------|-----|
| data_view | user_post_likes_view |

### 起点映射

| 起点属性 | 视图属性 |
|----------|----------|
| user_id | uid |

### 终点映射

| 视图属性 | 终点属性 |
|----------|----------|
| pid | post_id |
```

---

## 行动定义规范

### 语法

```markdown
## Action: {action_id}

**{显示名称}** - {简短描述}

| 绑定实体 | 行动类型 |
|----------|----------|
| {entity_id} | add | modify | delete |

### 触发条件

```yaml
field: {property_name}
operation: == | != | > | < | >= | <= | in | not_in | exist | not_exist
value: {value}
```

### 工具配置

| 类型 | 工具ID |
|------|--------|
| tool | {tool_id} |

或

| 类型 | MCP |
|------|-----|
| mcp | {mcp_id}/{tool_name} |

### 参数绑定

| 参数 | 来源 | 绑定 |
|------|------|------|
| {param_name} | property | {property_name} |
| {param_name} | input | - |
| {param_name} | const | {value} |

### 调度配置

(可选)

| 类型 | 表达式 |
|------|--------|
| FIX_RATE | {interval} |
| CRON | {cron_expr} |

### 执行说明

(可选) 详细的执行流程说明...
```

### 治理要求（强烈建议）

行动定义连接执行面（tool/mcp），为了稳定性与安全性，建议在每个 Action 中**显式写清**以下四类信息，并在工程侧落地相应治理：

1. **触发**：何时触发（手动/定时/条件触发），触发条件是否可复现
2. **影响范围**：影响哪些对象、范围边界、预期副作用
3. **权限与前置条件**：谁可以导入/启用/执行，是否需要审批，依赖的权限/凭据
4. **回滚/失败策略**：失败处理、重试策略、熔断/限流、可撤销性

> 推荐实践：导入不等于启用；启用与执行需要独立的权限与审计日志，并能关联到对应的 BKN 定义版本。

### 字段说明

| 字段 | 必须 | 说明 |
|------|:----:|------|
| action_id | YES | 行动唯一标识 |
| 绑定实体 | YES | 目标实体 ID |
| 行动类型 | YES | `add` / `modify` / `delete` |
| 触发条件 | NO | 自动触发的条件 |
| 工具配置 | YES | 执行的工具或 MCP |
| 参数绑定 | YES | 参数来源配置 |
| 调度配置 | NO | 定时执行配置 |

### 触发条件操作符

| 操作符 | 说明 | 示例 |
|--------|------|------|
| == | 等于 | `value: Running` |
| != | 不等于 | `value: Running` |
| > | 大于 | `value: 100` |
| < | 小于 | `value: 100` |
| >= | 大于等于 | `value: 100` |
| <= | 小于等于 | `value: 100` |
| in | 包含于 | `value: [A, B, C]` |
| not_in | 不包含于 | `value: [A, B, C]` |
| exist | 存在 | (无需 value) |
| not_exist | 不存在 | (无需 value) |
| range | 范围内 | `value: [0, 100]` |

### 参数来源

| 来源 | 说明 |
|------|------|
| property | 从实体属性获取 |
| input | 运行时用户输入 |
| const | 常量值 |

---

## 通用语法元素

### 表格格式

使用标准 Markdown 表格：

```markdown
| 列1 | 列2 | 列3 |
|-----|-----|-----|
| 值1 | 值2 | 值3 |
```

居中对齐（用于布尔值）：

```markdown
| 列1 | 列2 |
|-----|:---:|
| 值1 | YES |
```

### YAML 代码块

用于复杂结构（如条件表达式）：

```markdown
```yaml
condition:
  operation: and
  sub_conditions:
    - field: status
      operation: ==
      value: Failed
    - field: retry_count
      operation: <
      value: 3
`` `
```

### Mermaid 图表

用于可视化关系：

```markdown
```mermaid
graph LR
    A --> B
    B --> C
`` `
```

### 引用块

用于关键信息高亮：

```markdown
> **主键**: `id` | **显示属性**: `name`
```

### 标题层级

- `#` - 网络标题
- `##` - 类型定义 (Entity/Relation/Action)
- `###` - 定义内的主要 section
- `####` - 逻辑属性名等子项

---

## 文件组织

### 模式一：单文件（小型网络）

所有定义在一个 `.bkn` 文件中：

```markdown
---
type: network
id: my-network
---

# My Network

## Entity: entity1
...

## Entity: entity2
...

## Relation: rel1
...

## Action: action1
...
```

### 模式二：按类型拆分（中型网络）

使用 `index.bkn` 引用其他文件：

```markdown
---
type: network
id: my-network
includes:
  - entities.bkn
  - relations.bkn
  - actions.bkn
---

# My Network

网络描述...
```

### 模式三：每定义一文件（大型网络，推荐）

每个实体/关系/行动独立一个文件：

```
k8s-network/
├── index.bkn                    # type: network
├── entities/
│   ├── pod.bkn                  # type: entity
│   ├── node.bkn                 # type: entity
│   └── service.bkn              # type: entity
├── relations/
│   ├── pod_belongs_node.bkn     # type: relation
│   └── service_routes_pod.bkn   # type: relation
└── actions/
    ├── restart_pod.bkn          # type: action
    └── cordon_node.bkn          # type: action
```

**单实体文件示例** (`pod.bkn`):

```markdown
---
type: entity
id: pod
name: Pod实例
network: k8s-network
---

# Pod实例

Kubernetes 中的最小部署单元。

## 数据来源

| 类型 | ID |
|------|-----|
| data_view | view_123 |

> **主键**: `id` | **显示属性**: `pod_name`
```

---

## 增量导入规范

BKN 支持将任何 `.bkn` 文件动态导入到已有的知识网络。

### 导入器能力要求（工程可控性 9+ 的前提）

建议实现一个 **BKN Importer**，将 BKN 文件转换为系统变更，并提供以下能力（缺一不可）：

| 能力 | 说明 | 目的 |
|------|------|------|
| `validate` | 结构/表格/YAML block 校验，引用完整性校验，参数绑定校验 | 阻止错误进入系统 |
| `diff` | 计算变更集（新增/更新/删除）与影响范围 | 让变更可解释、可审计 |
| `dry_run` | 在不落地的情况下执行 validate + diff | 上线前预演 |
| `apply` | 执行落地（按确定性语义与冲突策略） | 可控执行 |
| `export` | 将线上知识网络状态导出为 BKN（可 round-trip） | 防漂移、可回滚、可复现 |

> 要求：所有导入操作必须记录审计信息（操作者、时间、输入文件指纹、变更集、结果）。

### 导入的确定性（必须保证）

为保证多人协作与可回放性，导入语义必须是**确定性的（deterministic）**：

- 对同一组输入文件（不考虑文件系统顺序）导入结果一致
- 同一文件重复导入结果一致（幂等）
- 冲突可解释：要么明确失败（fail-fast），要么有明确规则（例如 last-wins），不得“隐式合并”

### 唯一键与作用域

每个定义的唯一键建议为：

- `key = (network_id, type, id)`

其中 `network_id` 取自：

- 优先使用 frontmatter `network`
- 若缺失，则由导入目标网络（导入命令参数或 `type: network` 的 `id`）补齐

### 更新语义（replace vs merge）

默认建议使用 **replace（整段覆盖）**：

- 当 `key` 已存在时，用导入文件中的定义整体替换旧定义
- **缺失字段不代表删除**：仅代表“该字段不在本次定义中”；删除必须显式声明（见 `type: delete`）

如确有需要，可支持受控的 **merge-by-section（按章节合并）**，但必须满足：

- 仅允许合并少数“附加型章节”（例如 `属性覆盖`、`逻辑属性`）
- 冲突必须可控：同名逻辑属性/同名字段配置冲突时 fail-fast 或 last-wins（需配置）
- 合并策略必须在导入器中显式配置并记录到导入审计日志

### 冲突与优先级

当同一个 `key` 在一次导入批次中被多个文件重复声明：

- 默认：**fail-fast**（推荐，保证稳定性）
- 可选：按显式优先级排序（例如命令行顺序或 `priority` 字段），否则不建议支持

### 导入行为

| 场景 | 行为 |
|------|------|
| ID 不存在 | 创建新定义 |
| ID 已存在 | 更新定义（覆盖） |
| 使用 `type: delete` | 删除指定定义 |

### 导入示例

**场景：向已有网络添加新实体**

创建 `deployment.bkn`:

```markdown
---
type: entity
id: deployment
name: Deployment
network: k8s-network
---

# Deployment

Kubernetes 部署控制器。

## 数据来源

| 类型 | ID |
|------|-----|
| data_view | deployment_view |

> **主键**: `id` | **显示属性**: `deployment_name`
```

导入后，`k8s-network` 将包含新的 `deployment` 实体。

**场景：更新已有实体**

创建同 ID 的文件，导入后自动覆盖：

```markdown
---
type: entity
id: pod
name: Pod实例（更新版）
network: k8s-network
---

# Pod实例

更新后的定义...
```

**场景：删除定义**

```markdown
---
type: delete
network: k8s-network
targets:
  - entity: deprecated_entity
  - relation: old_relation
---

# 删除废弃定义

清理不再使用的定义。
```

**场景：批量导入（fragment）**

```markdown
---
type: fragment
id: monitoring-extension
name: 监控扩展
network: k8s-network
---

# 监控扩展

添加监控相关的实体和行动。

## Entity: alert

**告警**

### 数据来源

| 类型 | ID |
|------|-----|
| data_view | alert_view |

> **主键**: `id` | **显示属性**: `alert_name`

---

## Action: send_alert

**发送告警**

| 绑定实体 | 行动类型 |
|----------|----------|
| alert | add |

### 工具配置

| 类型 | 工具ID |
|------|--------|
| tool | alert_sender |
```

---

## Patch 规范（文件级别）

### 添加操作

```markdown
---
type: patch
id: 2026-01-31-add-metric
target: k8s-topology.bkn
operation: add
---

# 添加CPU指标

在 `## Entity: pod` 的 `### 逻辑属性` 后添加：

#### cpu_usage

- **类型**: metric
- **来源**: cpu_metric
```

### 修改操作

```markdown
---
type: patch
id: 2026-01-31-update-condition
target: k8s-topology.bkn
operation: modify
---

# 更新触发条件

将 `## Action: restart_pod` 的触发条件修改为：

```yaml
field: pod_status
operation: in
value: [Unknown, Failed, CrashLoopBackOff]
`` `
```

### 删除操作

```markdown
---
type: patch
id: 2026-01-31-remove-action
target: k8s-topology.bkn
operation: delete
---

# 删除废弃行动

删除 `## Action: deprecated_action`
```

---

## 最佳实践

### 命名规范

- **ID**: 小写字母、数字、下划线，如 `pod_belongs_node`
- **显示名称**: 简洁明确，如 "Pod属于节点"
- **标签**: 使用统一的标签体系

### 文档结构

1. 网络描述放在文件开头
2. 使用 mermaid 图展示整体拓扑
3. 实体定义在前，关系和行动在后
4. 相关定义放在一起

### 简洁原则

- 优先使用完全映射模式
- 只在需要时声明属性覆盖
- 避免重复信息

### 可读性

- 使用表格呈现结构化数据
- 添加业务语义说明
- 必要时使用 mermaid 图

---

## 参考

- [架构设计](./ARCHITECTURE.md)
- 样例：
  - [单文件模式](./examples/k8s-topology.bkn) - 所有定义在一个文件
  - [按类型拆分](./examples/k8s-network/) - 实体/关系/行动分文件
  - [每定义一文件](./examples/k8s-modular/) - 每个定义独立文件（推荐大规模场景）
