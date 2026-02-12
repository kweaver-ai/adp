# v0.3.0 版本交付物（Context Loader）

本文档是 `releases/v0.3.0/` 的**唯一说明文档**（本版本不再单独维护 README）。你可以把它当成“本版本交付清单 + 每个文件怎么用 + 如何验证可用”的入口。

---

## 1. 版本信息

- **版本号**：v0.3.0
- **发布日期**：2026-02-06

---

## 2. 本版本交付的能力（你能做什么）

本版本（v0.3.0）交付能力如下；具体接口与工具清单见后文“4.3 工具定义（快照）”。

- **构建并对外开放 Context Loader MCP Server**：形成标准化的 MCP 服务入口，对外提供 Context Loader 工具链能力。
- **解除对 `data-retrieval` 服务的依赖**：完成对 Context Loader 工具链的工艺管理（以本版本工具集快照为基准进行版本化管理与交付）。
- **新增业务知识网络索引构建能力**：增加索引构建任务创建与状态查询能力（对应工具：`create_kn_index_build_job`、`get_kn_index_build_status`）。

---

## 3. 变更摘要（面向使用方）

相对早期交付形态，本版本最关键的变化可以概括为三点：

- **文档形态调整**：`v0.3.0` 只保留 `overview.md` 作为单一说明入口（不再提供单独 README）。
- **能力增强**：在原“探索发现 → 精确查询 → 逻辑属性解析/行动召回”链路基础上，**新增知识网络索引构建能力**（创建构建任务、查询构建状态）。
- **依赖显式化**：新增 `tool-deps/`，将“执行工厂工具集”作为交付的一部分，避免接入方遗漏关键依赖。

---

## 4. 交付内容（本版本包含什么）

### 4.1 目录结构总览

```text
releases/v0.3.0/
├── overview.md                # 本版本说明文档（唯一入口）
├── toolset/                   # Context Loader 工具集（ADP 快照）；完整 OpenAPI 见 docs/apis/private/
├── tool-deps/                 # Context Loader 依赖的其他工具集（ADP 快照）
├── agent-deps/                # 必选依赖：逻辑属性解析等能力依赖的 Agent（导出包）
└── agent-recall-examples/     # 可选示例：Decision Agent 中使用 Context Loader 的示例（可回放）
```

---

### 4.2 文档

- **`overview.md`**：本版本唯一说明文档（即本文）。

---

### 4.3 工具定义（快照）

- **Context Loader 工具集（本版本快照）**：见 [`toolset/`](./toolset/)
  - **用途**：导入工具平台/运行时后，获得本版本对外开放的 Context Loader 工具集（以 ADP 形式交付）。
  - **包含工具（按功能分组）**：
    - 探索发现：`kn_schema_search`、`kn_search`
    - 精确查询：`query_object_instance`、`query_instance_subgraph`
    - 逻辑属性解析：`get_logic_properties_values`
    - 动态工具发现：`get_action_info`
    - 索引构建：`create_kn_index_build_job`、`get_kn_index_build_status`
  - **重要说明**：
    - ADP 中包含的 `server_url`（例如 `http://agent-retrieval:30779`）通常是**集群内服务地址示例**，接入方应根据实际环境替换/映射。

完整 OpenAPI 说明（长期维护的契约 SSOT，如与快照存在差异，应以 SSOT 为准并补充差异说明）：
- `docs/apis/private/`

---

### 4.4 工具依赖（快照）

- **Context Loader 依赖的其他工具集（本版本依赖快照）**：见 [`tool-deps/`](./tool-deps/)
  - **用途**：提供 Context Loader/其配套 Agent 在运行时所需的外部工具集定义（例如执行工厂相关能力）。
  - **示例工具**：`get_operator_schema`（获取算子市场指定算子详情，供算子类逻辑属性等场景使用）
  - **重要说明**：
    - ADP 中的 `server_url`（例如 `http://agent-operator-integration:9000/...`）同样需要按环境配置。

---

### 4.5 Agent 资产（快照）

#### 4.5.1 必选依赖 Agent（必须导入）

- **依赖 Agent 导入/导出包**：见 [`agent-deps/`](./agent-deps/)
  - **用途**：逻辑属性解析能力（`get_logic_properties_values`）依赖此处的 Agent/子 Agent（例如用于生成 metric/operator 的 dynamic_params）。**该目录为必选依赖，必须导入**，否则相关能力可能不可用或无法稳定运行。
  - **包含内容示例**（以导出包内定义为准）：
    - `generate_metric_logic_dynamic_params`：指标（metric）逻辑属性参数生成器
    - `generate_operator_logic_dynamic_params`：算子（operator）逻辑属性参数生成器（依赖执行工厂工具获取 schema）

#### 4.5.2 在 Decision Agent 中的使用示例（可选，可回放）

- **Decision Agent 使用示例**：见 [`agent-recall-examples/`](./agent-recall-examples/)
  - **用途**：展示在 Decision Agent 中如何使用 Context Loader（例如先 `kn_schema_search` 再查询实例/子图，再做逻辑属性解析或行动召回），可用于接入参考与回放验证。
---

## 5. 使用说明（快速开始）

### 5.1 导入与依赖准备

1. **导入 Context Loader 工具集**：导入 [`toolset/`](./toolset/) 下的 ADP 快照。
2. **导入工具依赖（按需）**：导入 [`tool-deps/`](./tool-deps/) 下的 ADP 快照。
   - 当你需要算子类逻辑属性/执行工厂联动（例如需要算子 schema）时，这是必需的。
3. **导入必选依赖 Agent**：导入 [`agent-deps/`](./agent-deps/) 下的导出包。
   - 逻辑属性解析能力 `get_logic_properties_values` 依赖这些 Agent/子 Agent 来生成 metric/operator 的 `dynamic_params`，因此该项为**必选**。

### 5.2 使用示例

- **Decision Agent 示例回放（推荐）**：见 [`agent-recall-examples/`](./agent-recall-examples/)
  - 用途：展示在 Decision Agent 中如何编排使用 Context Loader 工具链，可作为接入参考与回放验证。

---

## 6. Context Loader MCP Server（怎么使用）

v0.3.0 交付并对外开放 Context Loader MCP Server，供 Cursor / Claude Desktop / Claude Code 等 MCP 客户端直接连接并以“工具”的方式调用。

### 6.1 MCP 端点与必备 Header

- **端点**：`https://{host}:{port}/api/agent-retrieval/v1/mcp`
- **必备 Header**：
  - `Authorization: Bearer <TOKEN>`
  - `X-Kn-ID: <kn_id>`（推荐在连接级配置；未配置时可在工具调用参数中传 `kn_id`）

### 6.2 Cursor / Codex / Trae 等（同类工具配置示例）

以 Cursor 为例（同类工具通常也支持 `mcpServers` 结构）：

**方式 A：直接配置 url + headers**

```json
{
  "mcpServers": {
    "context-loader": {
      "url": "http://{host}:{port}/api/agent-retrieval/v1/mcp",
      "headers": {
        "Authorization": "Bearer 替换为实际token",
        "X-Kn-ID": "替换为实际kn_id"
      }
    }
  }
}
```

**方式 B：通过 `mcp-remote` 适配器传递 Header（通用）**

```json
{
  "mcpServers": {
    "context-loader": {
      "command": "npx",
      "args": [
        "-y",
        "mcp-remote",
        "https://{host}:{port}/api/agent-retrieval/v1/mcp",
        "--header",
        "Authorization: Bearer 替换为实际token",
        "--header",
        "X-Kn-ID: 替换为实际kn_id"
      ]
    }
  }
}
```

### 6.3 Claude Desktop 配置示例

配置文件（macOS）：`~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "context-loader": {
      "command": "npx",
      "args": [
        "-y",
        "mcp-remote",
        "https://{host}:{port}/api/agent-retrieval/v1/mcp",
        "--header",
        "Authorization: Bearer 替换为实际token",
        "--header",
        "X-Kn-ID: 替换为实际kn_id"
      ]
    }
  }
}
```

### 6.4 Claude Code 配置示例

**方式 A：CLI 添加（HTTP Transport + Header）**

```bash
claude mcp add --transport http \
  context-loader https://{host}:{port}/api/agent-retrieval/v1/mcp \
  --header "Authorization: Bearer 替换为实际token" \
  --header "X-Kn-ID: 替换为实际kn_id"
```

**方式 B：项目级 `.mcp.json`（推荐团队共享）**

```json
{
  "mcpServers": {
    "context-loader": {
      "type": "http",
      "url": "https://{host}:{port}/api/agent-retrieval/v1/mcp",
      "headers": {
        "Authorization": "Bearer ${TOKEN}",
        "X-Kn-ID": "${KN_ID}"
      }
    }
  }
}
```
