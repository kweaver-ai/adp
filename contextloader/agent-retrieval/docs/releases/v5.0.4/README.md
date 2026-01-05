# v5.0.4 版本交付物

## 1. 版本信息

- 版本号：v5.0.4
- 内部版本映射：5004
- 发布日期：2026-01-04
- Git 基线：待补充（Tag/Commit）

## 2. 变更摘要（面向使用方）

- 交付 Context Loader 工具链：探索发现 → 精确查询 → 逻辑属性解析/行动召回。
- 提供工具定义快照（OpenAPI/ADP）、Agent 导出包与编排示例，便于快速接入与回放验证。

## 3. 交付内容（本版本包含什么）

### 3.1 文档

- 工具使用指南（本版本）：[tool-usage-guide.md](./tool-usage-guide.md)
- 版本概览：[overview.md](./overview.md)

### 3.2 工具定义（快照）

- OpenAPI（本版本快照）：[toolset/openapi/](./toolset/openapi/)
- ADP（本版本快照）：[toolset/adp/contextloader_toolset_v5.0.4.adp](./toolset/adp/contextloader_toolset_v5.0.4.adp)

长期维护的契约 SSOT：
- `docs/apis/`（如与快照存在差异，应以 SSOT 为准并补充差异说明）

### 3.3 Agent 资产（快照）

- Agent 导入/导出包：[agent-deps/agent_export_20251230_175732.json](./agent-deps/agent_export_20251230_175732.json)
- Agent 编排示例：[agent-recall-examples/contextloader_agent_example.json](./agent-recall-examples/contextloader_agent_example.json)

## 4. 验证与证据（如何确认可用）

- 最小可跑通链路：参考 [tool-usage-guide.md](./tool-usage-guide.md) 中的“3. 快速开始”
- 调用链路选择：参考“4. 用法概览（如何选择工具）”
- 回放资产：
  - Agent 编排示例（见上）
  - 工具定义快照（见上）

## 5. 发布与回滚（占位）

- 发布步骤：待补充（镜像/Chart/配置基线、灰度策略）
- 回滚策略：待补充（触发条件、回滚动作、限制）

## 6. 已知问题

- 暂无（待补充）
