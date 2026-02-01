---
name: bkn-modeling
description: 编写和验证 BKN（业务知识网络）建模文件。用于创建实体类、关系类、行动类定义，或修改现有的 .bkn 文件。当用户提到业务知识网络、知识网络、BKN 文件时使用。
---

# BKN 业务知识网络建模

辅助编写 BKN (Business Knowledge Network) 业务知识网络建模文件。

## 快速开始

### 创建新的知识网络

1. 使用模板创建：`docs/ontology/bkn_docs/templates/network.bkn.template`
2. 参考样例：`docs/ontology/bkn_docs/examples/k8s-topology.bkn`
3. 查阅规范：`docs/ontology/bkn_docs/SPECIFICATION.md`

### BKN 文件结构

```markdown
---
type: network
id: my-network
name: 我的网络
version: 1.0.0
---

# 网络标题

## Entity: entity_id
...

## Relation: relation_id
...

## Action: action_id
...
```

## 三种类型

### 实体类 (Entity)

描述业务对象，直接映射数据视图：

```markdown
## Entity: pod

**Pod实例** - 描述

### 数据来源

| 类型 | ID |
|------|-----|
| data_view | view_id |

> **主键**: `id` | **显示属性**: `pod_name`
```

**配置模式**:
- **完全映射**: 只声明数据来源，自动继承所有字段
- **属性覆盖**: 仅声明需要特殊配置的属性
- **完整定义**: 完整声明所有属性（精细控制）

### 关系类 (Relation)

描述两个实体之间的关联：

```markdown
## Relation: pod_belongs_node

**Pod属于节点**

| 起点 | 终点 | 类型 |
|------|------|------|
| pod | node | direct |

### 映射规则

| 起点属性 | 终点属性 |
|----------|----------|
| pod_node_name | node_name |
```

### 行动类 (Action)

描述可执行的操作：

```markdown
## Action: restart_pod

**重启Pod**

| 绑定实体 | 行动类型 |
|----------|----------|
| pod | modify |

### 触发条件

```yaml
field: pod_status
operation: in
value: [Unknown, Failed]
`` `

### 工具配置

| 类型 | 工具ID |
|------|--------|
| tool | kubectl_delete_pod |

### 参数绑定

| 参数 | 来源 | 绑定 |
|------|------|------|
| pod | property | pod_name |
```

## 常用操作

### 添加实体

1. 在 `# 实体定义` section 下添加 `## Entity: {id}`
2. 声明数据来源（必须）
3. 声明主键和显示属性（必须）
4. 可选：属性覆盖、逻辑属性

### 添加关系

1. 在 `# 关系定义` section 下添加 `## Relation: {id}`
2. 声明起点、终点、类型（必须）
3. 声明映射规则（必须）

### 添加行动

1. 在 `# 行动定义` section 下添加 `## Action: {id}`
2. 声明绑定实体和行动类型（必须）
3. 声明工具配置和参数绑定（必须）
4. 可选：触发条件、调度配置

## 增量导入

BKN 支持动态增量导入，任何 `.bkn` 文件可直接导入到已有的知识网络。

### 文件类型

| type | 用途 |
|------|------|
| `entity` | 单个实体，独立文件 |
| `relation` | 单个关系，独立文件 |
| `action` | 单个行动，独立文件 |
| `fragment` | 混合片段，包含多个定义 |
| `delete` | 删除标记 |

### 导入行为

- **ID 不存在**: 新增定义
- **ID 已存在**: 更新定义（覆盖）
- **type: delete**: 删除指定定义

### 稳定性要求（重要）

- **幂等**：同一文件重复导入，结果必须一致
- **确定性**：同一批输入文件（不依赖目录遍历顺序）导入结果必须一致
- **冲突处理**：同一个 `(network, type, id)` 在同一批次重复定义时，默认应 fail-fast（推荐）

### 行动治理（重要）

行动连接执行面（tool/mcp），建议遵循：

- 导入不等于启用；启用/执行需要独立权限与审计
- Action 文档中明确：触发、影响范围、权限与前置条件、回滚/失败策略

### 新增实体示例

创建独立文件 `deployment.bkn`:

```markdown
---
type: entity
id: deployment
name: Deployment
network: k8s-network
---

# Deployment

## 数据来源

| 类型 | ID |
|------|-----|
| data_view | deployment_view |

> **主键**: `id` | **显示属性**: `deployment_name`
```

导入后自动添加到 `k8s-network`。

## 命名规范

| 类型 | 格式 | 示例 |
|------|------|------|
| 网络 ID | snake_case | `k8s_topology` |
| 实体 ID | snake_case | `pod`, `node` |
| 关系 ID | snake_case | `pod_belongs_node` |
| 行动 ID | snake_case | `restart_pod` |

## 验证清单

编写完成后检查：

- [ ] Frontmatter 包含 type, id, name, version
- [ ] 每个实体有数据来源、主键、显示属性
- [ ] 每个关系有起点、终点、映射规则
- [ ] 每个行动有绑定实体、工具配置、参数绑定
- [ ] ID 使用 snake_case 格式
- [ ] 表格格式正确

## 参考文档

- [架构设计](../../docs/ontology/bkn_docs/ARCHITECTURE.md)
- [语言规范](../../docs/ontology/bkn_docs/SPECIFICATION.md)
- 样例：
  - [单文件模式](../../docs/ontology/bkn_docs/examples/k8s-topology.bkn)
  - [按类型拆分](../../docs/ontology/bkn_docs/examples/k8s-network/)
  - [每定义一文件](../../docs/ontology/bkn_docs/examples/k8s-modular/) - 推荐大规模场景
- 模板文件：
  - [网络模板](../../docs/ontology/bkn_docs/templates/network.bkn.template)
  - [实体模板](../../docs/ontology/bkn_docs/templates/entity.bkn.template)
  - [关系模板](../../docs/ontology/bkn_docs/templates/relation.bkn.template)
  - [行动模板](../../docs/ontology/bkn_docs/templates/action.bkn.template)
