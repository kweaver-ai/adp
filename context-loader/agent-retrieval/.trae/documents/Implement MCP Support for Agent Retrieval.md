我将执行以下计划，首先完成文档的编写，然后进行代码实现。

### 1. 文档编写阶段 (立即执行)
*   **生成计划文档**: 将本计划详细内容写入 `/Users/cc/Desktop/work/aishu/adp/context-loader/agent-retrieval/docs/prd/feature-803607/plan.md`。
*   **编写详细设计文档**: 创建 `/Users/cc/Desktop/work/aishu/adp/context-loader/agent-retrieval/docs/prd/feature-803607/implementation_design.md`，内容包含：
    *   **需求分析**: 
        *   支持 `action_source.type = mcp`。
        *   获取 MCP 工具详情 (`GetMCPToolDetail`)。
        *   Schema 格式转换 (OpenAPI/JSON Schema -> OpenAI Function Call)。
        *   代理执行接口封装 (URL 构造)。
    *   **详细设计**:
        *   **接口定义**: `DrivenOperatorIntegration` 接口变更，新增请求/响应结构体。
        *   **代理 URL 封装设计**: 
            *   **设计**: 构造 URL `.../mcp/proxy/{mcp_id}/tools/{tool_name}/call`。
            *   **约定**: 该 URL 需被 `agent-operator-integration` 服务支持（或作为后续需求提交给该服务），它将接收 LLM 的扁平化参数，自动封装为 `{"tool_name": "...", "parameters": {...}}` 格式并转发给 MCP 执行接口。
        *   **Schema 转换逻辑**: 详解 `resolveMCPSchema` 算法，如何处理根节点 `$defs` 及递归引用。
    *   **测试计划**:
        *   单元测试用例：覆盖 Schema 转换的各种边界情况（嵌套、数组、引用）。
        *   集成测试用例：模拟完整调用流程。

### 2. 代码实现阶段
*   **接口定义更新**: 
    *   修改 `server/interfaces/kn_action_recall.go` (`ActionSource` 结构体)。
    *   修改 `server/interfaces/driven_operator_integration.go` (新增 `GetMCPToolDetail` 接口)。
*   **基础设施实现**:
    *   修改 `server/drivenadapters/operator_integration.go` 实现 MCP 工具列表获取与过滤。
*   **核心逻辑实现**:
    *   新建/修改 `server/logics/knactionrecall/schema_converter.go` 实现 MCP Schema 转换。
    *   修改 `server/logics/knactionrecall/get_action_info.go` 集成 MCP 处理流程。

### 3. 验证阶段
*   **编写单元测试**: 创建 `server/logics/knactionrecall/schema_converter_test.go` 并执行测试。
