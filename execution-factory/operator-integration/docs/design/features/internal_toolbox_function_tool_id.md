# 内部工具箱函数类型支持透传 tool_id

## 1. 背景

当前 `CreateInternalToolBox` 接口支持通过 `metadata_type=function` 创建内部工具箱。历史行为中，工具箱内工具的 `tool_id` 由系统自动生成，业务方无法在创建时显式指定工具 ID。

本次变更补充了函数类型内部工具箱的入参能力：调用方可在 `functions[]` 中为单个函数显式传入 `tool_id`。当该字段存在时，系统在创建内部工具时沿用调用方传入的工具 ID；当该字段为空时，仍保持原有自动生成逻辑。

## 2. 变更范围

本次仅影响内部工具箱创建接口在 `metadata_type=function` 场景下的请求体结构与落库行为：

- 接口：`POST /api/agent-operator-integration/v1/tool-box/intcomp`
- 接口：`POST /api/agent-operator-integration/internal-v1/tool-box/intcomp`
- 影响对象：`CreateInternalToolBoxReq.functions[]`
- 新增字段：`functions[].tool_id`

本次变更不影响：

- `metadata_type=openapi` 场景
- 未传 `tool_id` 的历史调用方式
- 工具箱其他字段和更新逻辑

## 3. 字段说明

函数类型内部工具箱请求中的 `functions[]` 由原先的 `FunctionInput` 扩展为 `InternalFunctionInput`，新增可选字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `tool_id` | string | 否 | 工具 ID。传入时优先使用该值创建内部工具；不传或为空时由系统自动生成 |

其余字段继续沿用 `FunctionInput` 原有定义，例如：

- `name`
- `description`
- `inputs`
- `outputs`
- `script_type`
- `code`

## 4. 行为说明

### 4.1 透传规则

当 `metadata_type=function` 时，服务端会逐个解析 `functions[]`：

1. 按原有逻辑解析函数元数据
2. 如果当前函数传入了非空 `tool_id`，则将 `函数名 -> tool_id` 建立映射
3. 创建工具箱内工具后，如果命中映射，则将生成工具的 `tool_id` 覆盖为调用方传入值

因此，`tool_id` 的透传是按单个函数生效的，而不是对整个工具箱统一生效。

### 4.2 不传时的兼容行为

如果 `functions[].tool_id` 未传或为空：

- 服务端不会报错
- 服务端不会尝试覆盖生成结果
- 工具 ID 仍按历史逻辑自动生成

这意味着已有调用方无需修改请求，也不会因本次变更产生兼容性问题。

### 4.3 使用约束

建议调用方在以下场景传入 `tool_id`：

- 需要在业务侧长期稳定引用某个内部工具 ID
- 需要在多环境或多版本发布时保持工具 ID 一致
- 需要和外部配置、权限、编排信息做稳定绑定

如果业务方不关心工具 ID 的可预测性，则可以继续沿用默认自动生成方式。

## 5. 请求示例

```json
{
  "box_id": "intbox_customer_service",
  "box_name": "客服内部工具箱",
  "box_desc": "客服场景使用的内部函数工具",
  "metadata_type": "function",
  "functions": [
    {
      "tool_id": "tool_query_order",
      "name": "query_order",
      "description": "查询订单信息",
      "script_type": "python",
      "inputs": [
        {
          "name": "order_id",
          "type": "string",
          "required": true,
          "description": "订单ID"
        }
      ],
      "outputs": [
        {
          "name": "result",
          "type": "object",
          "description": "订单信息"
        }
      ],
      "code": "def main(order_id):\n    return {\"order_id\": order_id}"
    },
    {
      "name": "ping_service",
      "description": "健康检查",
      "script_type": "python",
      "inputs": [],
      "outputs": [
        {
          "name": "alive",
          "type": "boolean",
          "description": "服务是否可用"
        }
      ],
      "code": "def main():\n    return {\"alive\": True}"
    }
  ],
  "source": "internal",
  "config_version": "1.0.0",
  "config_source": "manual"
}
```

预期行为：

- `query_order` 对应工具优先使用 `tool_query_order`
- `ping_service` 未传 `tool_id`，仍由系统自动生成工具 ID

## 6. 对接建议

- 如果业务依赖固定工具 ID，建议在函数定义阶段显式传入 `tool_id`
- 如果一个工具箱中存在多个函数，建议确保每个显式传入的 `tool_id` 在业务侧具备唯一性和可读性
- 如果同一批函数需要与外部系统做映射，建议将 `tool_id` 作为稳定主键维护，而不是依赖响应结果中的自动生成值

## 7. 兼容性与风险

### 7.1 兼容性

- 向后兼容旧请求
- 向后兼容未传 `tool_id` 的函数类型工具箱创建方式
- OpenAPI 文档中已同步新增 `InternalFunctionInput.tool_id`

### 7.2 风险提示

- 当前透传逻辑基于函数名建立映射，调用方应保证同一次请求中的函数名具备明确语义，避免维护上的歧义
- 本次文档仅说明“支持透传 `tool_id`”这一行为，不额外定义业务侧 `tool_id` 命名规范，命名约束应由接入方自行管理
