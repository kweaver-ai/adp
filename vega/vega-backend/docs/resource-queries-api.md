# 统一查询接口设计文档

## 接口描述

统一查询接口是系统提供的通用数据查询服务，支持多种数据源的查询操作。通过该接口，客户端可以查询任意resource的数据，支持通用SQL查询和方言查询。

该接口设计灵活，支持以下查询类型：
- 通用SQL查询（默认）
- 方言查询（通过`resource_type`指定，如`resource_type=mysql`，`resource_type=opensearch`等）

接口支持流式查询，通过search_after获取后续数据，不同resource_type的search_after格式不同，具体请参考请求示例。

## 请求信息

| 项目 | 内容 |
|------|------|
| 请求方法 | POST |
| 请求路径 | `/api/vega-backend/v1/resource-queries` |
| 内容类型 | `application/json` |

### 请求头

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| x-account-id | string | 是 | 账户ID |
| x-account-type | string | 是 | 账户类型 |

## 查询参数说明

该接口不使用URL查询参数，所有查询参数通过请求体传递。

## 请求体

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| resource_type | string | 否 | 资源类型，可选值：`opensearch`、`mysql`。不填则默认为通用SQL查询 |
| query | string | 是 | 查询语句。当`resource_type=opensearch`时为DSL查询语句；当`resource_type=mysql`时为MySQL语法的查询语句；不填`resource_type`时为通用SQL查询 |

### resource_type 说明（后续会按计划支持其他方言查询）

| 值 | 说明 | query格式 |
|----|------|-----------|
| 不填 | 通用SQL查询 | 标准SQL语句 |
| mysql | MySQL查询 | MySQL语法SQL语句 |
| opensearch | OpenSearch查询 | OpenSearch DSL语句 |

## 响应体

| 字段名 | 类型 | 说明                             |
|--------|------|--------------------------------|
| columns | array | 结果集列定义                         |
| columns[].name | string | 列名                             |
| columns[].type | string | 列数据类型（如：integer、string、boolean等） |
| entries | array | 查询结果数据行                        |
| entries[] | object | 单行数据，键为列名，值为对应数据               |
| search_after | array | 分页游标，用于获取下一页数据，不同resource_type的格式不同，具体请参考请求示例。为空表示无更多数据      |
| total_count | integer | 总记录数                           |

## 请求示例

### 示例1：通用SQL查询第一页数据

```bash
curl -X POST "https://your-domain/api/vega-backend/v1/resource-queries" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token-here" \
  -d '{
    "query": "select * from resource_id offset 0 limit 10"
  }'
```

### 示例2：通用SQL查询第二页数据

```bash
curl -X POST "https://your-domain/api/vega-backend/v1/resource-queries" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token-here" \
  -d '{
    "query": "select * from resource_id offset 10 limit 10"
  }'
```

### 示例3：通用SQL流式查询数据

```bash
curl -X POST "https://your-domain/api/vega-backend/v1/resource-queries" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token-here" \
  -d '{
    "search_after": ["query_id_xxx"],
  }'
```

### 示例4：MySQL查询

```bash
curl -X POST "https://your-domain/api/vega-backend/v1/resource-queries" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token-here" \
  -d '{
    "resource_type": "mysql",
    "query": "select * from resource_id limit 10 offset 0"
  }'
```

### 示例5：OpenSearch DSL查询第一页数据

```bash
curl -X POST "https://your-domain/api/vega-backend/v1/resource-queries" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token-here" \
  -d '{
    "resource_type": "opensearch",
    "query": {
        "query": {
            "match_all": {}
        },
        "size": 10,
        "sort": [
            {
                "id": {
                    "order": "asc"
                }
            }
        ]
    }
}'
```

### 示例6：OpenSearch DSL查询第二页数据

```bash
curl -X POST "https://your-domain/api/vega-backend/v1/resource-queries" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token-here" \
  -d '{
    "resource_type": "opensearch",
    "query": {
        "query": {
            "match_all": {}
        },
        "size": 10,
        "sort": [
            {
                "id": {
                    "order": "asc"
                }
            }
        ],
        "search_after": [
            10
        ], // 替换为第1页最后一条的sort值（如id=10）
        "from": 0 // search_after模式下from必须为0
    }
}'
```

## 响应示例

```json
{
  "columns": [
    {
      "name": "id",
      "type": "integer"
    },
    {
      "name": "name",
      "type": "string"
    }
  ],
  "entries": [
    {
      "id": 1,
      "name": "zs"
    },
    {
      "id": 2,
      "name": "ls"
    }
  ],
  "search_after": [
    "query_id_xxx"
  ],
  "total_count": 10
}
```

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 403 | 无权限访问该资源 |
| 500 | 服务器内部错误 |

## 注意事项

1. 查询语句建议添加适当的限制条件（如LIMIT），避免返回过多数据
2. OpenSearch DSL查询时，query字段需要传入完整的JSON对象
3. 使用search_after进行分页时，需要在下次请求中携带该参数
4. 不同数据源的数据类型映射可能存在差异，请参考columns中的type字段进行数据解析
