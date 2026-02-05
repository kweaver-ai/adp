# 面向 AI 时代的轻量级数据虚拟化与物化服务
# (Lightweight Data Virtualization & Materialization Service for the AI Era)

# Part 1: Top-Level Design (顶层设计)

## 1. 产品愿景 (Vision)
本服务旨在构建一个 **AI Native 的数据供给层**。它不试图复刻庞大的企业级数据操作系统（如 Palantir Foundry），而是专注于解决 AI 时代对数据的核心诉求：**多源连接**、**极速探索**与**向量化加速**。

*   **轻量级 (Lightweight)**: 极简极速，**易于部署**。采用单二进制架构，无复杂依赖，可秒级拉起，轻松适配各类基础设施。
*   **虚拟化 (Virtualization)**: 消除数据搬运壁垒，通过联邦机制一键连接。
*   **物化加速 (Materialization)**: 当需要性能时，通过简单的 "Sync" 动作将数据转化为高性能的**OpenSearch** 索引，解锁向量检索与随机访问能力。

---

## 2. 核心概念模型 (Core Concepts)

### 2.1. Catalog (目录 / 联邦网关)
*   **定义**: 管理数据源连接（Connection）与命名空间。
*   **策略**: **Virtual First**。配置即联通，无需预先 ETL。
*   **示例**: `Production_MySQL`, `AWS_S3`, `Corp_Kafka`.

#### 2.1.1. Catalog 类型
系统支持两类 Catalog，用于区分物理数据源和逻辑命名空间：

| 类型 | 说明 | Asset 来源 | 可写性 | 示例 |
| :--- | :--- | :--- | :--- | :--- |
| **Physical Catalog** | 对应真实数据源连接 | 系统自动发现（表、文件等） | 只读（Asset 由源端定义） | `mysql_prod`, `s3_logs` |
| **Logical Catalog** | 逻辑命名空间 | 用户手动创建（View、派生 Asset） | 可写 | `analytics`, `ml_features` |

**Catalog 配置示例**:
```yaml
catalogs:
  # Physical Catalog: 对应真实数据源
  - name: mysql_prod
    type: physical
    connection:
      type: mysql
      host: db.example.com
      database: production

  - name: s3_logs
    type: physical
    connection:
      type: s3
      bucket: company-logs
      region: us-east-1

  # Physical Catalog: 文档管理系统
  - name: feishu_corp
    type: physical
    connection:
      type: feishu
      app_id: "cli_xxx"
      app_secret: "${FEISHU_APP_SECRET}"
      tenant_key: "xxx"
    inventory:
      auto_discover: true
      scan_interval: "1h"

  - name: notion_team
    type: physical
    connection:
      type: notion
      integration_token: "${NOTION_TOKEN}"
    inventory:
      auto_discover: true

  # Logical Catalog: 存放派生资产
  - name: analytics
    type: logical
    description: "跨源分析视图"
    owner: data-team

  - name: ml_features
    type: logical
    description: "ML 特征工程"
    owner: ml-team

  # 系统预置的默认逻辑 Catalog
  - name: default
    type: logical
    system: true
    description: "默认逻辑命名空间"
```

#### 2.1.2. 连接状态管理 (Connection Status)

Physical Catalog 需要维护与数据源的连接状态，系统提供完整的状态监控和维护能力。

**连接状态定义**:

| 状态 | 说明 | 颜色 | 触发条件 |
| :--- | :--- | :--- | :--- |
| `healthy` | 连接正常 | 🟢 绿色 | 健康检查通过，延迟正常 |
| `degraded` | 性能降级 | 🟡 黄色 | 连接可用但延迟高或部分功能受限 |
| `unhealthy` | 连接异常 | 🔴 红色 | 健康检查失败，无法正常查询 |
| `offline` | 离线 | ⚫ 灰色 | 数据源不可达或网络中断 |
| `disabled` | 已禁用 | ⬜ 白色 | 用户主动禁用，不进行健康检查 |

**连接状态模型**:

```json
{
  "catalog_id": "mysql_prod",
  "connection_status": {
    "status": "healthy",
    "last_check_at": "2024-01-15T10:30:00Z",
    "last_success_at": "2024-01-15T10:30:00Z",
    "last_failure_at": null,
    "consecutive_failures": 0,
    "latency_ms": 45,
    "message": "Connection healthy",
    "details": {
      "server_version": "MySQL 8.0.32",
      "max_connections": 151,
      "active_connections": 23,
      "uptime_seconds": 8640000
    }
  }
}
```

**健康检查配置**:

```yaml
catalog:
  name: mysql_prod
  type: physical
  connection:
    type: mysql
    host: db.example.com
    database: production

  # 连接健康检查配置
  health_check:
    enabled: true

    # ========== 检查策略 ==========
    interval: "30s"                    # 检查间隔
    timeout: "5s"                      # 单次检查超时

    # ========== 探活方式 ==========
    probe:
      type: "query"                    # ping | query | custom
      query: "SELECT 1"                # query 模式的探活 SQL
      # custom_endpoint: "/health"     # custom 模式的健康检查端点

    # ========== 状态判定阈值 ==========
    thresholds:
      latency_warning_ms: 500          # 延迟超过此值标记为 degraded
      latency_critical_ms: 2000        # 延迟超过此值标记为 unhealthy
      failure_threshold: 3             # 连续失败次数达到此值标记为 unhealthy
      recovery_threshold: 2            # 连续成功次数达到此值恢复为 healthy

    # ========== 重试策略 ==========
    retry:
      max_attempts: 3                  # 单次检查最大重试次数
      backoff: "exponential"           # fixed | exponential
      initial_delay: "1s"
      max_delay: "30s"
```

**不同数据源的探活方式**:

| 数据源类型 | 默认探活方式 | 探活命令/请求 |
| :--- | :--- | :--- |
| MySQL/MariaDB | `query` | `SELECT 1` |
| PostgreSQL | `query` | `SELECT 1` |
| ClickHouse | `query` | `SELECT 1` |
| S3 | `api` | `HeadBucket` |
| Kafka | `api` | `DescribeCluster` |
| OpenSearch | `http` | `GET /_cluster/health` |
| 飞书 | `http` | `GET /open-apis/auth/v3/tenant_access_token` |
| Notion | `http` | `GET /v1/users/me` |

**连接状态 API**:

```http
# 获取单个 Catalog 连接状态
GET /api/v1/catalogs/{id}/status

# 响应
{
  "catalog_id": "mysql_prod",
  "name": "mysql_prod",
  "type": "physical",
  "connection_status": {
    "status": "healthy",
    "last_check_at": "2024-01-15T10:30:00Z",
    "latency_ms": 45,
    "message": "Connection healthy"
  }
}

# 获取所有 Catalog 连接状态概览
GET /api/v1/catalogs/status/summary

# 响应
{
  "total": 10,
  "by_status": {
    "healthy": 7,
    "degraded": 1,
    "unhealthy": 1,
    "offline": 0,
    "disabled": 1
  },
  "catalogs": [
    { "id": "mysql_prod", "name": "mysql_prod", "status": "healthy", "latency_ms": 45 },
    { "id": "pg_analytics", "name": "pg_analytics", "status": "degraded", "latency_ms": 850 },
    { "id": "kafka_events", "name": "kafka_events", "status": "unhealthy", "message": "Connection refused" }
  ]
}
```

**连接维护操作**:

```http
# 手动测试连接
POST /api/v1/catalogs/{id}/test-connection
{
  "timeout": "10s"
}

# 响应
{
  "success": true,
  "latency_ms": 52,
  "server_info": {
    "version": "MySQL 8.0.32",
    "server_id": "db-prod-01"
  }
}

# 禁用连接 (停止健康检查和查询)
POST /api/v1/catalogs/{id}/disable
{
  "reason": "Planned maintenance"
}

# 启用连接
POST /api/v1/catalogs/{id}/enable

# 强制重连 (断开现有连接并重新建立)
POST /api/v1/catalogs/{id}/reconnect
{
  "drain_timeout": "30s"      # 等待现有查询完成的超时时间
}

# 更新连接配置 (热更新)
PATCH /api/v1/catalogs/{id}/connection
{
  "pool_size": 20,
  "connection_timeout": "10s"
}
```

**状态变更通知**:

```yaml
catalog:
  name: mysql_prod

  # 状态变更通知配置
  status_notifications:
    enabled: true

    # 通知渠道
    channels:
      webhook: "https://hooks.example.com/catalog-status"
      slack: "#data-platform-alerts"
      email: "data-team@example.com"

    # 通知规则
    rules:
      - from: ["healthy"]
        to: ["degraded", "unhealthy", "offline"]
        severity: "warning"

      - from: ["degraded"]
        to: ["unhealthy", "offline"]
        severity: "critical"

      - from: ["unhealthy", "offline"]
        to: ["healthy"]
        severity: "info"
        message_template: "Connection recovered after {downtime}"

    # 通知抑制 (避免告警风暴)
    suppression:
      min_interval: "5m"           # 同一状态最小通知间隔
      flapping_threshold: 3        # 短时间内状态切换次数阈值
      flapping_window: "10m"       # 抖动检测时间窗口
```

**连接池状态监控**:

```http
# 获取连接池详细状态
GET /api/v1/catalogs/{id}/connection-pool

# 响应
{
  "catalog_id": "mysql_prod",
  "pool_status": {
    "max_connections": 50,
    "active_connections": 12,
    "idle_connections": 28,
    "waiting_requests": 0,
    "total_connections_created": 156,
    "total_connections_closed": 106,
    "avg_acquisition_time_ms": 2.3,
    "avg_connection_lifetime_s": 3600
  }
}
```

### 2.2. Asset (资产 - 统一实体)
我们采用 **Asset** 作为通过数据创造价值的统一实体。系统支持**八大类** Asset，且收敛为统一的存储形态。

| 资产类型 | 对应源 (Source) | 语义 | Virtual 行为 (Remote) | Local 行为 (Native / OpenSearch) |
| :--- | :--- | :--- | :--- | :--- |
| **Table** | MySQL, PG | **结构化表** | JDBC 联邦查询 | **CDC Sync** (实时 Binlog -> OpenSearch) |
| **Fileset**| S3, HDFS, 飞书, Notion | **非结构化文件集** | 浏览/预览/API 代理 | **ETL Pipeline** (解析 -> OpenSearch Index) |
| **API** | REST, GraphQL | **应用接口** | Debug/Viewing | **Polling Job** (轮询 -> Flatten -> OpenSearch) |
| **Metric** | Influx, Prom | **时序指标** | PromQL 下推 | **Batch Archives** (归档为 OpenSearch Index) |
| **Topic** | Kafka, Pulsar| **实时流** | 实时采样 (Sampling) | **Micro-batch** (追加写入 OpenSearch) |
| **Index** | ES, OpenSearch | **搜索引擎** | Search DSL 透传 | **Reindex** (远程索引 -> 本地 OpenSearch) |
| **View** | SQL Logic | **逻辑算子/视图**| **Trino/DuckDB** (智能路由) | **OpenSearch Index** (CTAS) |
| **Dataset** | API 写入 | **原生可写数据集** | N/A (仅 Local) | **直接存储** (API -> OpenSearch) |

#### 2.2.1. Asset 状态管理 (Asset Status)

每个 Asset 拥有独立的状态，用于控制其可用性和行为。

**Asset 状态定义**:

| 状态 | 说明 | 查询行为 | Sync 行为 | 可转换至 |
| :--- | :--- | :--- | :--- | :--- |
| `active` | 正常可用 | ✅ 正常执行 | ✅ 正常同步 | `disabled`, `deprecated` |
| `disabled` | 已禁用 | ❌ 返回错误 | ❌ 暂停同步 | `active` |
| `deprecated` | 已废弃 | ⚠️ 返回警告 + 结果 | ✅ 继续同步 | `active`, `disabled` |
| `stale` | 数据过期 | ⚠️ 返回警告 + 结果 | ✅ 继续同步 | `active` (自动) |

**状态模型**:

```json
{
  "asset_id": "ast_orders_001",
  "name": "orders",
  "catalog": "mysql_prod",
  "status": {
    "state": "active",
    "reason": null,
    "disabled_at": null,
    "disabled_by": null,
    "last_state_change": "2024-01-10T08:00:00Z"
  }
}
```

**禁用时的查询行为**:

```http
# 查询被禁用的 Asset
POST /api/v1/query
{
  "asset": "mysql_prod.orders",
  "operation": "select",
  "fields": ["*"]
}

# 响应 (HTTP 403)
{
  "error": {
    "code": "ASSET_DISABLED",
    "message": "Asset 'mysql_prod.orders' is disabled",
    "details": {
      "disabled_at": "2024-01-15T10:00:00Z",
      "disabled_by": "admin@example.com",
      "reason": "Data quality issues under investigation"
    }
  }
}
```

**废弃状态的查询行为**:

```http
# 查询废弃的 Asset
POST /api/v1/query
{
  "asset": "mysql_prod.old_orders",
  "operation": "select",
  "fields": ["*"]
}

# 响应 (HTTP 200 + 警告头)
HTTP/1.1 200 OK
X-Asset-Warning: Asset 'mysql_prod.old_orders' is deprecated, use 'mysql_prod.orders_v2' instead
Deprecation: true
Sunset: 2024-06-01

{
  "data": [...],
  "warnings": [
    {
      "code": "ASSET_DEPRECATED",
      "message": "This asset is deprecated and will be removed on 2024-06-01",
      "suggestion": "Migrate to 'mysql_prod.orders_v2'"
    }
  ]
}
```

**状态管理 API**:

```http
# 禁用 Asset
POST /api/v1/assets/{id}/disable
{
  "reason": "Data quality issues under investigation",
  "notify_subscribers": true
}

# 启用 Asset
POST /api/v1/assets/{id}/enable

# 标记为废弃
POST /api/v1/assets/{id}/deprecate
{
  "sunset_date": "2024-06-01",
  "replacement": "mysql_prod.orders_v2",
  "migration_guide": "https://docs.example.com/migration/orders"
}

# 取消废弃标记
POST /api/v1/assets/{id}/undeprecate
```

**与 Catalog 状态的关系（级联效果）**:

| Catalog 状态 | Asset 状态 | 实际行为 |
| :--- | :--- | :--- |
| `healthy` | `active` | ✅ 正常 |
| `healthy` | `disabled` | ❌ Asset 禁用错误 |
| `disabled` | `active` | ❌ Catalog 禁用错误 (优先) |
| `disabled` | `disabled` | ❌ Catalog 禁用错误 (优先) |
| `unhealthy` | `active` | ⚠️ 尝试查询，可能失败 |

> **规则**: Catalog 状态优先于 Asset 状态。Catalog 禁用时，其下所有 Asset 均不可访问。

**禁用对相关功能的影响**:

| 功能 | Asset 禁用时的行为 |
| :--- | :--- |
| **Virtual 查询** | 返回 `ASSET_DISABLED` 错误 |
| **Local 查询** | 返回 `ASSET_DISABLED` 错误 |
| **Sync 任务** | 暂停，不触发新任务 |
| **CDC 同步** | 暂停消费，保留 offset |
| **Inventory 扫描** | 跳过，保持 disabled 状态 |
| **View 引用** | 依赖此 Asset 的 View 查询失败 |
| **血缘查询** | 正常返回，状态标记为 disabled |

**批量状态操作**:

```http
# 批量禁用
POST /api/v1/assets/bulk-disable
{
  "asset_ids": ["ast_001", "ast_002", "ast_003"],
  "reason": "Scheduled maintenance"
}

# 按 Catalog 批量禁用
POST /api/v1/catalogs/{id}/assets/disable-all
{
  "reason": "Database migration in progress",
  "exclude": ["critical_table"]      # 排除列表
}
```

**状态变更审计**:

```json
{
  "event_type": "asset_status_changed",
  "timestamp": "2024-01-15T10:00:00Z",
  "asset_id": "ast_orders_001",
  "changes": {
    "from": "active",
    "to": "disabled"
  },
  "actor": {
    "user_id": "user_admin",
    "email": "admin@example.com"
  },
  "reason": "Data quality issues under investigation",
  "ip_address": "10.0.1.100"
}
```

#### 2.2.2. 数据类型映射规则 (Type Mapping)

VEGA 引擎采用统一的类型系统，将来自不同数据源的异构类型映射为标准 VEGA 类型，确保跨源数据的类型一致性和互操作性。

**VEGA 类型系统定义**:

| VEGA 类型 | 描述 | 存储范围 | 物化到 OpenSearch | 备注 |
| :--- | :--- | :--- | :--- | :--- |
| **integer** | 有符号整数 | -2^63 ~ 2^63-1 | long | 包含 tinyint, smallint, int, bigint 等 |
| **unsigned_integer** | 无符号整数 | 0 ~ 2^64-1 | unsigned_long | 仅 MySQL/MariaDB/ClickHouse 原生支持 |
| **float** | 浮点数 | IEEE 754 单精度/双精度 | double | 包含 float, double, real 等 |
| **decimal** | 任意精度数 | 精度可配置 (默认 38,18) | scaled_float | 用于金融精度计算 |
| **string** | 短字符串 | 可变长度，一般 < 65KB | keyword | 用于精确匹配、排序、聚合 |
| **text** | 长文本 | 可变长度，最大 2GB | text | 支持全文搜索和分词 |
| **date** | 日期 | 日期部分 (年月日) | date | 格式：YYYY-MM-DD |
| **datetime** | 日期时间 | 日期+时间（含时区）| date | 格式：RFC3339 |
| **time** | 时间 | 时间部分 (时分秒) | keyword | 格式：HH:mm:ss |
| **boolean** | 布尔值 | true / false | boolean | - |
| **binary** | 二进制数据 | 字节数组 | binary | 大对象存储为外部引用 |
| **json** | JSON 对象 | 结构化 JSON | object | 支持嵌套查询 |
| **vector** | 向量 | 浮点数组（维度固定） | knn_vector | 用于 AI/ML 场景 |
| **point** | 地理点 | 经纬度坐标 | geo_point | 暂不支持 |
| **shape** | 地理形状 | 多边形/线 | geo_shape | 暂不支持 |
| **ip** | IP 地址 | IPv4/IPv6 | ip | 暂不支持，作为 string 处理 |

**原始数据库类型到 VEGA 类型映射**:

**1) MySQL / MariaDB**

| VEGA 类型 | MySQL/MariaDB 类型 |
| :--- | :--- |
| **integer** | tinyint, smallint, int, integer, mediumint, bigint, year |
| **unsigned_integer** | tinyint unsigned, smallint unsigned, mediumint unsigned, int unsigned, integer unsigned, bigint unsigned |
| **float** | float, double, real, double precision |
| **decimal** | decimal(m,d), numeric, fixed, dec |
| **string** | char(n), varchar(n) |
| **text** | text, tinytext, mediumtext, longtext |
| **date** | date |
| **datetime** | timestamp, datetime |
| **time** | time |
| **boolean** | boolean, bool, bit |
| **binary** | binary, varbinary, tinyblob, blob, mediumblob, longblob, bit |
| **json** | json |

**2) PostgreSQL**

| VEGA 类型 | PostgreSQL 类型 |
| :--- | :--- |
| **integer** | smallint, integer, int, int4, bigint, int8, serial, bigserial |
| **float** | real, float4, float8, double precision, money |
| **decimal** | numeric(p,s), decimal, money(p,s) |
| **string** | char(n), character(n), varchar(n), character varying(n), bpchar(n), interval(p) |
| **text** | text |
| **date** | date |
| **datetime** | timestamp, timestamp(p), timestamp with time zone, timestamp with local time zone, timestamptz(p) |
| **time** | time, time(p), time with time zone, time with time zone(p), timetz |
| **boolean** | boolean, bool |
| **binary** | bytea, bit |
| **json** | json, jsonb |

**3) Oracle**

| VEGA 类型 | Oracle 类型 |
| :--- | :--- |
| **integer** | smallint, int, integer, pls_integer, number (无小数位) |
| **float** | float, binary_double, binary_float |
| **decimal** | number(m,n), decimal(m,n) |
| **string** | char(n), nchar(n), varchar2(n), nvarchar2(n), char, varchar, rowid |
| **text** | clob, nclob |
| **datetime** | date, timestamp, timestamp(n), timestamp with time zone, timestamp with local time zone |
| **boolean** | - (不支持，使用 number(1) 模拟) |
| **binary** | raw, long raw, blob, bfile |

**4) SQL Server**

| VEGA 类型 | SQL Server 类型 |
| :--- | :--- |
| **integer** | tinyint, smallint, int, bigint |
| **float** | real, float, money |
| **decimal** | decimal(m,n), numeric |
| **string** | char(n), nchar(n), varchar(n), nvarchar(n) |
| **text** | text, ntext |
| **date** | date |
| **datetime** | datetime, datetime2, smalldatetime, datetimeoffset |
| **time** | time |
| **boolean** | bit |
| **binary** | binary, varbinary, image |

**5) ClickHouse**

| VEGA 类型 | ClickHouse 类型 |
| :--- | :--- |
| **integer** | int8, int16, int32, int64 |
| **unsigned_integer** | uint8, uint16, uint32, uint64 |
| **float** | float32, float64 |
| **decimal** | decimal(m,n) |
| **string** | string, fixedstring |
| **date** | date |
| **datetime** | datetime |
| **boolean** | boolean |

> 注：ClickHouse 不支持 text 类型，长文本使用 string 类型存储

**6) Doris**

| VEGA 类型 | Doris 类型 |
| :--- | :--- |
| **integer** | tinyint, smallint, int, bigint |
| **float** | float, double |
| **decimal** | decimal(m,d), numeric, number(m,n) |
| **string** | char(n), varchar(n), string, varchar2(n) |
| **date** | date |
| **datetime** | datetime |
| **boolean** | boolean |
| **json** | json, jsonb |

> 注：Doris 不支持 text 类型，长文本使用 string 类型存储

**7) MaxCompute**

| VEGA 类型 | MaxCompute 类型 |
| :--- | :--- |
| **integer** | tinyint, smallint, int, bigint |
| **float** | float, double |
| **decimal** | decimal(m,d) |
| **string** | char(n), varchar(n), string |
| **date** | date |
| **datetime** | datetime, timestamp |
| **boolean** | boolean |
| **binary** | binary, byte |

> 注：MaxCompute 不支持 text 类型，长文本使用 string 类型存储

**8) GBase8s**

| VEGA 类型 | GBase8s 类型 |
| :--- | :--- |
| **integer** | smallint, integer, serial, int8, bigint, serial8, bigserial |
| **float** | float, double precision, smallfloat, real |
| **decimal** | decimal(m,d), dec, numeric(p,s), money(p,s) |
| **string** | char(n), nchar(n), varchar(m,r), varchar2(m,r), nvarchar(m,r), character, character varying(m,r), interval_year_month, interval_day_time, rowid |
| **date** | date |
| **datetime** | datetime, timestamp, timestamp with time zone, timestamp with local time zone |
| **time** | time, time with time zone |
| **boolean** | boolean |
| **binary** | binary, blob, byte |

> 注：GBase8s 不支持 text 类型，长文本使用 string 类型存储

**9) Dameng (达梦)**

| VEGA 类型 | Dameng 类型 |
| :--- | :--- |
| **integer** | tinyint, smallint, int, bigint |
| **float** | float, double, double precision, real |
| **decimal** | decimal(m,n), number(m,n) |
| **string** | char(n), character(n), varchar(n), varchar2(n), nvarchar2(n), char, varchar |
| **text** | clob |
| **date** | date |
| **datetime** | timestamp, timestamp with time zone, timestamp with local time zone |
| **time** | time |
| **boolean** | bit |
| **binary** | binary, varbinary, raw |

**10) GaussDB**

| VEGA 类型 | GaussDB 类型 |
| :--- | :--- |
| **integer** | tinyint, smallint, integer, bigint, int1, int2, int4, int8 |
| **float** | float, double precision, real, float4, float8, float64 |
| **decimal** | decimal(m,n), numeric(m,d), number(m,n) |
| **string** | char(n), nchar(n), bpchar(n), varchar(n), character(n), character varying(n), varchar2(n), nvarchar2(n), interval |
| **text** | text |
| **date** | date |
| **datetime** | timestamp(p), timestamp with time zone, timestamptz(p), smalldatetime, datetime |
| **time** | time(p), time with time zone(p), timetz |
| **boolean** | boolean, bool |
| **binary** | bytea, blob, clob, binary |
| **json** | json, jsonb |

**11) OpenSearch (物化目标)**

| VEGA 类型 | OpenSearch 类型 |
| :--- | :--- |
| **integer** | byte, short, integer, long |
| **unsigned_integer** | unsigned_long |
| **float** | double, float, half_float, scaled_float |
| **string** | keyword |
| **text** | text |
| **date** | date |
| **datetime** | date |
| **boolean** | boolean |
| **binary** | binary |
| **vector** | knn_vector |
| **point** | geo_point |
| **shape** | geo_shape |
| **ip** | ip |

**类型映射规则说明**:

1.  **整数类型**:
    *   自动识别有符号/无符号，unsigned 类型映射到 `unsigned_integer`
    *   Serial 类型（PostgreSQL/GBase8s）映射为对应的整数类型
    *   Year 类型（MySQL/MariaDB）映射为 `integer`

2.  **精度数值**:
    *   `decimal` 未指定精度时，默认映射为 `decimal(38,18)`
    *   `numeric` 未指定精度时，默认映射为 `decimal(38,18)`
    *   `money` 类型（PostgreSQL/SQL Server/GBase8s）映射为 `decimal` 并保留精度
    *   Oracle `number` 类型根据精度自动选择 `integer` 或 `decimal`

3.  **字符串类型**:
    *   **string**: 固定长度字符串（char）和可变长度字符串（varchar）映射为 `string`，用于精确匹配、排序和聚合
    *   **text**: 长文本类型（text, clob, nclob, longtext, mediumtext）映射为 `text`，支持全文搜索和分词
    *   国际化字符类型（nchar, nvarchar, nvarchar2）自动处理 UTF-8 编码，映射为 `string`
    *   Interval 类型（PostgreSQL/GBase8s/GaussDB）作为 `string` 处理
    *   部分数据库（ClickHouse, Doris, MaxCompute, GBase8s）不支持 `text` 类型，长文本使用 `string` 类型存储

4.  **时间类型**:
    *   带时区的时间戳（timestamp with time zone, timestamptz）保留时区信息
    *   `datetime` 类型统一映射为 `datetime`
    *   `smalldatetime`（SQL Server/GaussDB）精度损失警告（精度到分钟）
    *   PostgreSQL `interval` 类型暂不支持，映射为 `string`

5.  **二进制数据**:
    *   大对象类型（BLOB）根据大小决定是否内联存储
    *   超过 32KB 的二进制数据建议存储为外部引用
    *   `bfile`（Oracle）作为外部文件引用处理

6.  **特殊类型**:
    *   **向量类型** (`vector`): OpenSearch 原生支持 (`knn_vector`)，其他数据库通过扩展字段存储
    *   **空间类型** (`point`, `shape`): 当前版本暂不支持，未来版本将支持 PostGIS 和 Oracle Spatial
    *   **IP 类型**: 当前版本暂不支持，作为 `string` 处理
    *   **JSON 类型**: 原生支持 MySQL/PostgreSQL/MariaDB/Doris/GaussDB 的 JSON/JSONB 类型

**类型转换示例**:

```yaml
# MySQL 表定义
CREATE TABLE orders (
  id BIGINT UNSIGNED PRIMARY KEY,
  amount DECIMAL(10,2),
  customer_name VARCHAR(100),
  created_at TIMESTAMP,
  metadata JSON
);

# 映射到内部 Schema
{
  "catalog": "mysql_prod",
  "asset": "orders",
  "fields": [
    {"name": "id", "type": "unsigned_integer", "nullable": false},
    {"name": "amount", "type": "decimal(10,2)", "nullable": true},
    {"name": "customer_name", "type": "string", "nullable": true},
    {"name": "created_at", "type": "datetime", "nullable": true},
    {"name": "metadata", "type": "json", "nullable": true}
  ],
  "primary_key": ["id"]
}

# 物化到 OpenSearch Index Mapping
{
  "mappings": {
    "properties": {
      "id": {"type": "unsigned_long"},
      "amount": {"type": "scaled_float", "scaling_factor": 100},
      "customer_name": {
        "type": "text",
        "fields": {"keyword": {"type": "keyword", "ignore_above": 256}}
      },
      "created_at": {"type": "date"},
      "metadata": {"type": "object", "enabled": true}
    }
  }
}
```

**跨数据库类型兼容性注意事项**:
*   **unsigned 整数**: 仅 MySQL/MariaDB/ClickHouse 原生支持，其他数据库需使用 CHECK 约束模拟
*   **JSON 类型**: 不支持的数据库（Oracle/SQL Server）使用 CLOB/NVARCHAR(MAX) + 应用层解析
*   **时区处理**: 建议统一使用 UTC 存储，避免跨时区数据一致性问题
*   **精度损失**: Money 类型（固定精度）到 Decimal 转换时需验证精度是否满足需求

#### 2.2.3. Table: 结构化数据映射
**定义**: 传统的行列式二维表。
*   **Virtual Mode**: 通过 JDBC/ODBC 协议进行联邦查询。系统不存储数据，仅直连查询。
*   **Local Mode**: 通过 CDC (Change Data Capture) 实时同步到 OpenSearch。
*   **Schema 映射**:
    *   Primary Key -> Document ID (`_id`)
    *   Columns -> Document Fields
    *   JSON/Text Column -> Nested Field / Text Field (Standard Analyzer)

#### 2.2.4. Fileset: 非结构化文档映射 (文档系统)
当对接飞书、Notion、S3 等系统时，映射关系如下：

**Catalog 层级 → 租户/Bucket 级别**
*   飞书: 企业租户 (`feishu_corp`)
*   S3: Bucket (`s3_logs`)

**Asset 层级 → 知识库/目录前缀**
*   飞书: 知识库 (`feishu_corp.tech_wiki`)
*   S3: 目录前缀 (`s3_logs.nginx_access`)

**配置示例**:
```yaml
asset:
  name: tech_wiki
  type: fileset
  source: { type: feishu_wiki, space_id: "7xxx" }
  sync: { mode: incremental, schedule: "0 */6 * * *" }
  embedding_config:
    enabled: true
    source_fields: ["title", "content"]
```

#### 2.2.5. API: 接口数据扁平化
**定义**: 将返回 JSON 列表的 API 映射为数据表。
*   **Virtual Mode**: 实时调用 API (带有 Pagination)，返回前 N 条作为预览。
*   **Local Mode**: 定时轮询 (Polling)，将结果扁平化存储。
*   **Mapping**: 指定 `root_path` (如 `data.items`)，数组中每个对象映射为一个 Document。

**示例**:
```yaml
asset:
  name: weather_history
  type: api
  source:
    url: "https://api.weather.com/v1/history"
    method: "GET"
    pagination: { type: "offset", limit: 100 }
    roots: "data"  # JSON path to array
```

#### 2.2.6. Metric: 时序指标归档
**定义**: Prometheus/InfluxDB 中的时序数据流。
*   **Virtual Mode**: 透传 PromQL/InfluxQL 查询。
*   **Local Mode**: 定时降采样 (Downsample) 并归档到 OpenSearch。
*   **Mapping**:
    *   Datetime -> `@timestamp`
    *   Value -> `value` (double)
    *   Labels/Tags -> `labels` (nested object or flattened)

#### 2.2.7. Topic: 消息流接入
**定义**: Kafka/Pulsar 中的 Topic。
*   **Virtual Mode**: 采样 (Sampling)。消费最新的 100 条消息用于预览 Schema 和内容。
*   **Local Mode**: 启动消费者组，微批写入 OpenSearch。
*   **Mapping**:
    *   Key -> `_key`
    *   Payload (JSON) -> Document Fields
    *   Headers -> Metadata

#### 2.2.8. Index: 搜索引擎代理
**定义**: 外部 ElasticSearch/OpenSearch 索引。
*   **Virtual Mode**: 代理查询请求，提供统一的网关鉴权和审计。
*   **Local Mode**: Reindex 操作，将远程数据完整迁移到本地集群。

#### 2.2.9. View: 逻辑视图映射
View 作为派生资产，其归属 Catalog 遵循以下规则：

*   **单源 View**: 默认归属源 Catalog，可覆盖归属到 Logical Catalog。
*   **跨源 View**: 必须显式归属到 Logical Catalog (如 `analytics`)。

**创建示例**:
```json
{
  "name": "user_behavior_360",
  "type": "view",
  "catalog": "analytics",
  "definition": {
    "operation": "join",
    "left": {"asset": "mysql_prod.users"},
    "right": {"asset": "s3_logs.click_events"}
  }
}
```

#### 2.2.10. Dataset: 原生可写数据集

**定义**: 系统原生管理的可读写数据集，数据通过 API 直接写入 OpenSearch，无需外部数据源。

**核心特性**:
*   **API 驱动**: 通过 REST API 直接进行 CRUD 操作，无需依赖外部数据源
*   **Schema 定义**: 用户自定义字段结构、类型约束、主键
*   **仅 Local 模式**: 数据直接存储于 OpenSearch，无 Virtual 模式
*   **归属要求**: 必须归属于 Logical Catalog

**与其他 Asset 类型的区别**:

| 维度 | 其他 Asset (Table/Fileset 等) | Dataset |
| :--- | :--- | :--- |
| **数据来源** | 外部数据源同步 | API 直接写入 |
| **写入方式** | Sync (单向同步) | API (可读写) |
| **Virtual 模式** | 支持 (联邦查询) | 不支持 |
| **数据所有权** | 源端拥有 | 系统拥有 |

**典型使用场景**:
*   **AI 产出存储**: RAG 应用提取的知识片段、LLM 生成的结构化数据
*   **用户自定义数据**: 手动上传的数据、应用程序写入的事件
*   **中间结果持久化**: 跨源 Join 的结果保存、预计算的聚合数据
*   **标注/反馈数据**: 人工标注的训练数据、用户反馈

**Schema 定义示例**:
```yaml
asset:
  name: "user_feedback"
  type: "dataset"
  catalog: "app_data"           # 必须归属 Logical Catalog

  schema:
    fields:
      - name: "id"
        type: "string"
        nullable: false
        primary_key: true       # 主键，用于 Upsert 去重

      - name: "user_id"
        type: "string"
        nullable: false

      - name: "content"
        type: "text"
        nullable: false
        features:
          - type: "fulltext"
            config: { analyzer: "ik_max_word" }
          - type: "vector"
            config: { dimension: 768, auto_generate: true }

      - name: "rating"
        type: "integer"
        nullable: true

      - name: "metadata"
        type: "json"
        nullable: true

      - name: "created_at"
        type: "datetime"
        nullable: false
        default: "now()"

  settings:
    shards: 3
    replicas: 1

  embedding_config:
    enabled: true
    source_fields: ["content"]
    model: "text-embedding-3-small"
```

**Schema 约束支持**:

| 约束类型 | 描述 | 示例 |
| :--- | :--- | :--- |
| `nullable` | 是否允许空值 | `nullable: false` |
| `primary_key` | 主键（支持复合主键） | 用于 Upsert 去重 |
| `default` | 默认值 | `"now()"`, `0`, `"unknown"` |

**写入流程**:
```
API Request
    │
    ▼
┌──────────────┐
│ 1. Validate  │ ← Schema 校验、类型检查
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ 2. Transform │ ← 默认值填充、类型转换
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ 3. Embedding │ ← (可选) 生成向量
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ 4. Bulk Write│ ← OpenSearch Bulk API
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ 5. Response  │ ← 返回写入统计
└──────────────┘
```

**Catalog 配置扩展**:
```yaml
catalogs:
  - name: app_data
    type: logical
    description: "应用数据层"
    dataset_config:
      enabled: true                      # 允许创建 Dataset
      storage:
        index_prefix: "dataset_app_data_"
        default_shards: 3
        default_replicas: 1
      quotas:
        max_datasets: 100
        max_total_docs: 100000000
        max_write_batch_size: 10000
```

#### 2.2.11. 字段特征系统 (Field Features)

为支持高级数据能力（向量检索、多字段映射、分词器配置、精确匹配等），系统在 `ViewField` 结构中引入 `Features` 扩展，为字段赋予超越基础类型的语义能力。

**适用范围**:
*   **Virtual 模式**: 关系型数据库（MySQL/PG 等）可通过 `RefField` 引用已有字段定义特征，实现逻辑层的能力扩展。
*   **Local 模式 (物化后)**: 所有 Asset 类型均可定义完整特征，物化到 OpenSearch 时自动生成对应 Mapping。

**核心数据结构**:
```go
type FieldFeature struct {
    Name     string            `json:"name"`      // 特征名称
    Type     FieldFeatureType  `json:"type"`      // keyword, fulltext, vector
    Description  string            `json:"description"`   // 备注
    RefField string            `json:"ref_field"` // 引用的物理字段
    Enabled  bool              `json:"enabled"`   // 是否启用（同类型仅一个为true）
    IsNative bool              `json:"is_native"` // true:系统同步, false:用户扩展
    Config   map[string]any    `json:"config"`    // 特征配置
}

type FieldFeatureType string // "keyword" | "fulltext" | "vector"
```

**特征类型说明**:

| 类型 | 用途 | 典型配置 | Virtual 模式 | Local 模式 |
| :--- | :--- | :--- | :--- | :--- |
| **keyword** | 精确匹配、排序、聚合 | `ignore_above_len: 2048` | 引用其他字段 | 原生支持 |
| **fulltext** | 全文检索、中文分词 | `analyzer: "ik_max_word"` | 引用 text 字段 | 原生支持 |
| **vector** | 向量语义搜索 | `dimension: 768, space_type: "cosinesimil"` | 引用 vector 字段 | 原生支持 |

**设计亮点**:
1.  **物理零侵入**: 无需修改源端 Schema 即可为字段"挂载"向量搜索或全文检索能力，Virtual 模式下通过引用实现。
2.  **权责清晰**: `IsNative` 参数完美隔离"系统自动同步"与"人工逻辑扩展"的边界，避免同步覆盖问题。
3.  **引用透明性**: `RefField` 使字段可灵活"借用"其他列的能力，对上层业务隐藏底层物理结构。
4.  **热切换能力**: 支持为一个字段配置多个向量模型或分词器，通过 `Enabled` 状态位实现搜索能力的"热切换"。
5.  **物化增强**: Virtual 模式下定义的特征，在物化时自动转换为 OpenSearch 原生 Mapping，无缝升级。

**字段特征示例**:
```json
{
  "name": "document",
  "type": "text",
  "features": [
    {
      "type": "fulltext",
      "ref_field": "document",
      "enabled": true,
      "is_native": true,
      "config": { "analyzer": "ik_max_word" }
    },
    {
      "type": "keyword",
      "ref_field": "document.keyword",
      "enabled": true,
      "is_native": true,
      "config": { "ignore_above_len": 2048 }
    },
    {
      "type": "vector",
      "ref_field": "document_vector_768",
      "enabled": true,
      "is_native": false,
      "config": {}
    }
  ]
}
```

**同步规则 (Auto-Sync)**:
*   扫描底层 Mapping 时，系统构造 `IsNative: true` 的特征。
*   若 Features 中不存在该 Type 的特征，设为 `Enabled: true`。
*   若已存在手动添加的特征 (`IsNative: false`)，则新同步的特征设为 `Enabled: false`。
*   **原则**: 系统永远不自动覆盖用户的 Enabled 状态，只负责增量同步 IsNative 特征。

**排他性约束**:
*   同一 `ViewField` 下，同一 `Type` 的特征中，`Enabled == true` 的元素个数 ≤ 1。
*   UI 层面启用某特征时，自动关闭该字段下同类型的其他特征。

#### 2.2.12. Asset 自动发现机制 (Inventory)

Inventory Worker 扫描数据源元数据，自动发现并注册 Asset。

**扫描方式 (按 Asset 类型)**:
| Asset 类型 | 扫描方法 | 元数据来源 |
| :--- | :--- | :--- |
| **Table** | `SHOW TABLES` / `information_schema` | MySQL, PG, ClickHouse 等 |
| **Fileset** | `ListBuckets` / `ListObjects` / `GetSpaces` | S3, 飞书, Notion 等 |
| **Topic** | `ListTopics` / `DescribeTopics` | Kafka, Pulsar 等 |
| **Index** | `GET /_cat/indices` | OpenSearch, ElasticSearch |

**发现策略配置**:

```yaml
catalog:
  name: mysql_prod
  type: physical
  connection: { ... }

  # 自动发现配置
  inventory:
    # ========== 触发策略 ==========
    trigger:
      mode: "scheduled"              # manual | scheduled | event_driven
      schedule: "0 */6 * * *"        # cron 表达式 (scheduled 模式)
      on_connection_test: true       # 连接测试时触发扫描

    # ========== 发现模式 ==========
    discovery_mode: "incremental"    # full | incremental
    # full: 全量扫描，对比完整列表
    # incremental: 增量扫描，仅检测变化

    # ========== 变更处理策略 ==========
    changes:
      on_new_asset: "auto_register"        # auto_register | pending_review | ignore
      on_deleted_asset: "mark_stale"       # auto_remove | mark_stale | ignore
      on_schema_change: "auto_update"      # auto_update | pending_review | ignore

    # ========== 过滤规则 ==========
    filters:
      include_patterns:              # 包含规则 (正则)
        - "^(?!_).*"                 # 排除下划线开头的表
        - "orders_.*"                # 包含 orders_ 前缀的表
      exclude_patterns:              # 排除规则 (正则)
        - ".*_backup$"               # 排除 _backup 后缀
        - ".*_tmp$"                  # 排除 _tmp 后缀
        - "^sys_.*"                  # 排除系统表
      exclude_schemas:               # 排除的 Schema/Database
        - "information_schema"
        - "performance_schema"
        - "mysql"

    # ========== 高级选项 ==========
    options:
      scan_timeout: "5m"             # 单次扫描超时
      max_assets_per_scan: 10000     # 单次最大发现数量
      schema_sample_rows: 100        # Schema 推断采样行数
      parallel_workers: 4            # 并行扫描线程数
```

**触发模式说明**:

| 模式 | 说明 | 适用场景 |
| :--- | :--- | :--- |
| `manual` | 仅通过 API 手动触发 | 变更频率低、需严格控制的环境 |
| `scheduled` | 按 cron 表达式定时触发 | 常规生产环境 |
| `event_driven` | 监听源端变更事件 (如 DDL 事件) | 支持事件推送的数据源 |

**变更处理策略说明**:

| 事件 | 策略选项 | 行为 |
| :--- | :--- | :--- |
| **新增 Asset** | `auto_register` | 自动注册为 Virtual Asset |
| | `pending_review` | 标记为待审核，需人工确认 |
| | `ignore` | 忽略，不注册 |
| **删除 Asset** | `auto_remove` | 自动删除 Asset 记录 |
| | `mark_stale` | 标记为 stale，保留元数据 |
| | `ignore` | 忽略，保持原状 |
| **Schema 变更** | `auto_update` | 自动更新 Schema 定义 |
| | `pending_review` | 标记为待审核 |
| | `ignore` | 忽略变更 |

**手动触发 API**:

```http
# 触发单个 Catalog 的发现
POST /api/v1/catalogs/{id}/inventory/scan
{
  "mode": "full",                    # full | incremental
  "dry_run": false,                  # 仅预览，不实际变更
  "filters": {                       # 可选：覆盖默认过滤规则
    "include_patterns": ["orders_.*"]
  }
}

# 响应
{
  "scan_id": "scan_abc123",
  "status": "running",
  "progress_url": "/api/v1/jobs/scan_abc123"
}

# 查询扫描结果
GET /api/v1/catalogs/{id}/inventory/scans/{scan_id}
{
  "status": "completed",
  "summary": {
    "total_discovered": 150,
    "new_assets": 5,
    "deleted_assets": 2,
    "schema_changes": 3,
    "unchanged": 140
  },
  "changes": [
    { "asset": "orders_2024", "change_type": "new", "action": "registered" },
    { "asset": "old_logs", "change_type": "deleted", "action": "marked_stale" },
    { "asset": "users", "change_type": "schema_changed", "action": "updated" }
  ]
}
```

**发现事件通知**:

```yaml
inventory:
  notifications:
    on_new_asset:
      webhook: "https://hooks.example.com/new-asset"
      slack: "#data-alerts"
    on_schema_change:
      webhook: "https://hooks.example.com/schema-change"
      severity: "warning"
    on_deleted_asset:
      slack: "#data-alerts"
      severity: "info"
```

#### 2.2.13. 层级设计决策
采用 **二级命名空间** (`catalog.asset`)，而非三级。
*   **Catalog**: 数据源实例 / 租户 / 逻辑分组
*   **Asset**: 表 / 文件集 / 视图 / 接口
通过 Asset 内部的 `path` 和 `tags` 字段实现更细粒度的逻辑分组能力。



### 2.3. 虚拟计算引擎: 适应性双引擎 (Adaptive Compute)
为了应对从小数据预览到海量数据 Join 的不同需求，我们采用双引擎策略：
1.  **DuckDB (Embedded)**:
    *   **场景**: 数据探索 (Preview), 小文件分析 (<10GB), 读时解析 CSV/JSON。
    *   **优势**: 进程内零延迟，部署简单。
2.  **Trino (External Cluster)**:
    *   **场景**: **跨库大数据量 Join** (如 MySQL Join Hive), 复杂过滤。
    *   **优势**: 分布式内存计算，也是业界成熟的 Data Virtualization 标准引擎。
    *   **集成**: 我们的 Go Control Plane 负责向 Trino 提交 SQL，并流式获取结果。

### 2.4. 双模态设计 (Dual Mode)
每个 Asset 对象逻辑上包含两个视图，系统根据查询场景智能路由：
1.  **Virtual View (v1)**: 实时指向源端 (Federated)。适合：`SELECT * FROM order WHERE id=1` (即时查)。
2.  **Local View (v2)**: 指向本地 OpenSearch 索引 (Materialized)。适合：`Vector_Search(desc_embedding)` 或 `Scan(last_year_data)` (AI 分析)。

> **核心原则**: 物化 (Materialization) 是 Asset 的一种**状态/能力**，而非一种特殊的 Asset 类型。
> *   **1:1 加速**: 直接对 Physical Asset (如 MySQL Table) 开启 Sync，系统自动维护 Local View，无需创建新 Asset。
> *   **复杂加工**: 创建 Logical View Asset (定义 SQL 逻辑) 并开启 Sync，实现类似 "Materialized View" 的效果。

### 2.5. 查询路由决策逻辑 (Query Routing Strategy)
系统根据以下规则自动选择执行引擎，**优先使用本地物化数据**以获得最佳性能：

```
┌─────────────────────────────────────────────────────────────────┐
│                        Query Router                              │
├─────────────────────────────────────────────────────────────────┤
│  Input: DSL Query                                                │
│                                                                  │
│  Step 1: 检查物化状态 (Check Materialization)                     │
│    ├── Local View 存在且新鲜度 < TTL → 优先使用 OpenSearch (Local) │
│    └── 否则 → 进入 Step 2 (Virtual Mode)                         │
│                                                                  │
│  Step 2: 评估数据规模 & 类型                                      │
│    ├── Vector Search → OpenSearch (直接执行)                     │
│    ├── Full-text Search → ElasticSearch DSL (透传)               │
│    └── SQL Query → Step 3                                        │
│                                                                  │
│  Step 3: 评估 SQL 复杂度                                          │
│    ├── 单表查询 → 源端直连 (MySQL/PG/ClickHouse)                  │
│    ├── 同源 Union/Join → 下推至源端 (Pushdown)                    │
│    └── 跨源 Union/Join → Step 4                                  │
│                                                                  │
│  Step 4: 选择计算引擎 (Cross-Source Compute)                      │
│    ├── 数据量 < 1GB (小数据/预览) → DuckDB (Embedded)             │
│    └── 数据量 ≥ 1GB (大数据/批处理) → Trino (Distributed Cluster) │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**路由逻辑详解**:
1.  **Check Materialization** (Highest Priority): 如果数据已经同步到 OpenSearch 且未过期，直接查 OpenSearch。这是 AI 场景下性能最高的路径。
2.  **Single/Same Source**: 如果必须查源端（实时性要求高或未物化），且是单表或同源 Join，直接下推到源数据库执行，避免数据搬运。
3.  **Cross Source**: 只有在跨源 Join 时才启用虚拟计算引擎 (DuckDB/Trino)。

**路由配置示例**:
```yaml
routing:
  thresholds:
    duckdb_max_scan_size: 10GB
    duckdb_max_join_size: 1GB
    local_view_ttl: 1h
  preferences:
    prefer_local_for_vector: true
    prefer_local_for_aggregation: true
    force_pushdown_for_realtime: false # 是否强制下推以获取最新数据
```

### 2.6. 统一查询语言 (Unified DSL)
为了屏蔽底层引擎（MySQL, Trino, OpenSearch）的语法差异，系统对外提供一套**自定义 DSL**（基于 JSON 的结构化查询语言）。
*   **AST 中间层**: API 网关接收 DSL -> 解析为 AST -> 转译器 (Transpiler) -> 目标方言 (SQL / PromQL / ES DSL)。
*   **设计目标**: 让前端或 AI Agent 以面向对象的方式组装查询，而无需拼接 SQL 字符串。

**DSL 语法示例**:

```json
// 示例 1: 简单查询
// 命名空间格式: {catalog}.{asset} (二级结构)
{
  "asset": "sales.orders",
  "operation": "select",
  "fields": ["order_id", "customer_name", "amount"],
  "filter": {
    "and": [
      {"field": "status", "op": "eq", "value": "completed"},
      {"field": "amount", "op": "gt", "value": 1000}
    ]
  },
  "sort": [{"field": "created_at", "order": "desc"}],
  "limit": 100
}

// 转译为 SQL:
// SELECT order_id, customer_name, amount FROM sales.orders
// WHERE status = 'completed' AND amount > 1000
// ORDER BY created_at DESC LIMIT 100
```

```json
// 示例 2: 向量检索
{
  "asset": "knowledge.documents",
  "operation": "vector_search",
  "vector_field": "content_embedding",
  "query_text": "如何配置 Kubernetes 网络策略",
  "top_k": 10,
  "filter": {"field": "doc_type", "op": "eq", "value": "tutorial"},
  "return_fields": ["title", "content", "url"]
}

// 转译为 OpenSearch Query:
// { "query": { "knn": ... } }
```

```json
// 示例 3: 跨源 Join
{
  "operation": "join",
  "left": {
    "asset": "mysql_prod.users",
    "alias": "u"
  },
  "right": {
    "asset": "s3_logs.user_events",
    "alias": "e"
  },
  "join_type": "left",
  "on": {"left": "u.user_id", "op": "eq", "right": "e.user_id"},
  "fields": ["u.name", "u.email", {"agg": "count", "field": "e.event_id", "alias": "event_count"}],
  "group_by": ["u.user_id", "u.name", "u.email"]
}
```

```json
// 示例 4: 时序指标查询
{
  "asset": "prometheus.http_requests",
  "operation": "metric_query",
  "aggregation": "rate",
  "window": "5m",
  "filter": {"field": "status_code", "op": "regex", "value": "5.."},
  "group_by": ["service", "endpoint"],
  "time_range": {"start": "-1h", "end": "now"}
}

// 转译为 PromQL:
// sum(rate(http_requests{status_code=~"5.."}[5m])) by (service, endpoint)
```

---

## 3. 系统架构 (System Architecture)

采用 **Golang Native** 架构，并新增 **Query Transpiler** 层。

```mermaid
graph TD
    subgraph "Control Plane (Go Microservices)"
        API[API Gateway (DSL Parser)]
        Transpiler[Query Transpiler]
        CatalogSvc[Catalog Service]
        JobSched[Scheduler]
        AuthSvc[Auth & RBAC Service]
    end

    subgraph "Data Plane (Go Workers)"
        Inventory[Inventory Worker]
        ETL_Worker[ETL / Vectorize Worker]
        Stream_Worker[CDC / Stream Worker]
        DuckDB[DuckDB (Embedded)]
        EmbedSvc[Embedding Service]
    end

    subgraph "External Compute"
        Trino[Trino Cluster]
    end

    subgraph "Storage Layer"
        OpenSearch[(OpenSearch Cluster)]
        MetaDB[(PostgreSQL - Metadata)]
    end

    subgraph "Observability"
        Metrics[Prometheus]
        Traces[Jaeger]
        Logs[Loki]
    end

    User --> API
    API -- "Authenticate" --> AuthSvc
    API -- "DSL" --> Transpiler
    Transpiler -- "Get Schema" --> CatalogSvc
    Transpiler -- "Gen SQL" --> DuckDB
    Transpiler -- "Gen Plan" --> Trino
    Transpiler -- "Gen Request" --> Stream_Worker

    %% Query Flow
    Transpiler -- "Small Query" --> DuckDB
    Transpiler -- "Heavy Query" --> Trino
    DuckDB -- "Federated Read" --> MySQL_Source & S3_Source & OpenSearch

    %% Materialization Flow
    JobSched -- "Trigger Sync" --> ETL_Worker
    ETL_Worker -- "1. Submit SQL" --> Trino
    Trino -- "2. Stream Result (Arrow)" --> ETL_Worker
    ETL_Worker -- "3. Generate Embedding" --> EmbedSvc
    ETL_Worker -- "4. Write" --> OpenSearch

    %% Lineage & Observability
    ETL_Worker -- "Report Provenance" --> CatalogSvc
    API & ETL_Worker & Stream_Worker --> Metrics & Traces & Logs
```

### 3.1. 为何选择 Golang?
- **High Performance**: Go 的协程模型 (Goroutines) 非常适合处理大量的 IO 密集型任务（如 CDC 消费、API 轮询、文件下载）。
- **Data Engineering Eco**: Go 拥有成熟的数据库驱动 (MySQL/PG) 和云原生生态 (Kubernetes/MinIO)。
- **OpenSearch Integration**: Go 拥有成熟的 OpenSearch 官方客户端 (`opensearch-go`)，可高效处理批量写入与查询。

### 3.2. 向量嵌入生成流程 (Embedding Pipeline)
物化过程中的向量生成采用可插拔的 Embedding Service：

```
┌─────────────────────────────────────────────────────────────────┐
│                    Embedding Pipeline                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. ETL Worker 提取文本字段                                       │
│     └── 根据 Asset Schema 中的 embedding_config 配置              │
│                                                                  │
│  2. 批量发送至 Embedding Service                                  │
│     ├── External API: OpenAI / Cohere / Azure                    │
│     └── Internal Service: Go-native ONNX Runtime (Future)        │
│                                                                  │
│  3. 向量写入 OpenSearch                                          │
│     └── 利用 k-NN 插件构建索引                                    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Embedding 配置示例**:
```yaml
asset:
  name: knowledge_base
  embedding_config:
    enabled: true
    source_fields: ["title", "content"]
    target_field: "content_embedding"
    model:
      provider: openai           # openai | local | custom
      model_name: text-embedding-3-small
      dimensions: 1536
    batch_size: 100
    index:
      type: IVF_PQ
      num_partitions: 256
      num_sub_vectors: 96
```

### 3.3. CDC 同步机制 (Change Data Capture)

```
┌─────────────────────────────────────────────────────────────────┐
│                      CDC Sync Flow                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  MySQL/PG ──[Binlog/WAL]──> Debezium ──[Kafka]──> Stream Worker  │
│                                                                  │
│  Stream Worker 处理流程:                                          │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ 1. 消费 CDC 事件 (INSERT/UPDATE/DELETE)                     │ │
│  │ 2. 应用 Schema 映射 (字段转换、类型适配)                      │ │
│  │ 3. 攒批 (Micro-batch): 按时间窗口(5s) 或条数(1000) 触发      │ │
│  │ 4. 写入 OpenSearch (Upsert for UPDATE/DELETE)               │ │
│  │ 5. OpenSearch 自动处理索引刷新和段合并                        │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  一致性保证:                                                      │
│  ├── At-least-once: Kafka offset 提交在写入成功后                 │
│  ├── 幂等写入: 使用 (source_table, pk, version) 作为去重键        │
│  └── 最终一致: Compaction 后数据与源端一致 (延迟 < 30s)           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 4. 关键特性 (Key Features)

### 4.1. Local Mode: Why OpenSearch?
我们选择 **OpenSearch** 作为物化层的核心存储，这是综合考虑向量检索、全文搜索、结构化查询、读写分离、水平扩展等需求后的最优选择。

#### 4.1.1. 核心优势

*   **向量检索**: 支持 k-NN (HNSW/IVF/PQ)，GPU 加速，十亿级向量规模。
*   **全文搜索**: 原生倒排索引，分词/高亮/facet，中文分词 (IK) 支持。
*   **混合检索**: 向量 + BM25 混合查询原生支持，RAG 场景最佳实践。
*   **结构化查询**: DSL 支持复杂过滤、聚合分析。
*   **读写分离**: 集群天然支持 Primary/Replica 分离。
*   **水平扩展**: 分片机制，支持 PB 级数据。
*   **Trino 集成**: 官方 Connector，联邦查询无需适配。
*   **Go SDK**: 官方支持，生产级成熟度。

#### 4.1.2. 存储层选型对比

在物化存储选型时，我们对比了多种方案：

| 维度 | OpenSearch | pgvector | Qdrant | Lance |
| :--- | :--- | :--- | :--- | :--- |
| **向量规模** | ✅ 十亿级 (GPU) | ⚠️ 千万级 | ✅ 亿级 | ✅ 亿级 |
| **向量性能** | ✅ 高 | ⚠️ 中 | ✅ 高 | ✅ 高 |
| **全文搜索** | ✅✅ 原生强项 | ✅ tsvector | ❌ 无 | ⚠️ 基础 |
| **混合检索** | ✅ 原生支持 | ⚠️ 需手动 | ⚠️ 需手动 | ⚠️ 需手动 |
| **字段过滤** | ✅ DSL | ✅ SQL | ✅ Filter | ✅ SQL-like |
| **聚合查询** | ✅ 强 | ✅ SQL | ❌ 无 | ⚠️ 基础 |
| **读写分离** | ✅ 副本 | ✅ 副本 | ✅ 副本 | ⚠️ 需自建 |
| **水平扩展** | ✅ 分片 | ⚠️ 有限 | ✅ 分片 | ⚠️ 单机 |
| **Go SDK** | ✅ 官方 | ✅ 原生 | ✅ 官方 | ⚠️ 实验性 |
| **Trino 集成** | ✅ 官方 | ✅ 官方 | ❌ 需自研 | ❌ 需自研 |
| **云托管** | ✅ AWS/阿里云 | ✅ 各云 | ✅ Cloud | ✅ Cloud |

**设计决策**：选择 **OpenSearch** 作为物化存储，理由如下：

1. **混合检索原生支持**：向量 + 全文搜索是 RAG 场景核心需求，OpenSearch 原生支持 Hybrid Search
2. **十亿级规模**：GPU 加速索引构建，分层存储（内存/磁盘/S3），支持超大规模数据
3. **Trino 无缝集成**：官方 Connector，跨源 JOIN 无需额外适配
4. **Go SDK 成熟**：官方支持，生产级稳定性，无语言绑定风险
5. **读写分离**：集群天然支持 Primary/Replica，满足高可用需求
6. **运维成熟**：AWS OpenSearch Service / 阿里云等托管服务完善

虽然 OpenSearch 本身不轻量，但我们将它视为**像 MySQL 一样的外部基础设施**。
- **服务本身 (Control Plane)** 保持极简、单二进制、无状态。
- **数据状态 (Data Plane)** 下沉到成熟的 OpenSearch 集群。
- 这符合现代云原生架构：应用与状态分离。对于用户，只需提供一个 OpenSearch 连接串即可。

#### 4.1.3. 索引设计

**Mapping 示例**：

```json
{
  "mappings": {
    "properties": {
      "id": { "type": "keyword" },
      "catalog_id": { "type": "keyword" },
      "asset_id": { "type": "keyword" },

      "title": {
        "type": "text",
        "analyzer": "ik_max_word",
        "fields": { "keyword": { "type": "keyword" } }
      },
      "content": {
        "type": "text",
        "analyzer": "ik_max_word"
      },
      "doc_type": { "type": "keyword" },
      "status": { "type": "keyword" },
      "tags": { "type": "keyword" },
      "metadata": { "type": "object", "enabled": true },
      "path": { "type": "keyword" },
      "created_at": { "type": "date" },
      "updated_at": { "type": "date" },

      "embedding": {
        "type": "knn_vector",
        "dimension": 1536,
        "method": {
          "name": "hnsw",
          "space_type": "cosinesimil",
          "engine": "faiss",
          "parameters": {
            "ef_construction": 256,
            "m": 16
          }
        }
      }
    }
  },
  "settings": {
    "index": {
      "knn": true,
      "number_of_shards": 5,
      "number_of_replicas": 1
    }
  }
}
```

#### 4.1.4. 查询示例

```json
// 1. 纯向量检索
{
  "size": 10,
  "query": {
    "knn": {
      "embedding": { "vector": [0.1, 0.2, ...], "k": 10 }
    }
  }
}

// 2. 向量 + 字段过滤
{
  "size": 10,
  "query": {
    "bool": {
      "must": [
        { "knn": { "embedding": { "vector": [...], "k": 100 } } }
      ],
      "filter": [
        { "term": { "catalog_id": "feishu_corp" } },
        { "term": { "status": "active" } },
        { "range": { "created_at": { "gte": "2024-01-01" } } }
      ]
    }
  }
}

// 3. 混合检索（向量 + 全文）
{
  "size": 10,
  "query": {
    "hybrid": {
      "queries": [
        { "knn": { "embedding": { "vector": [...], "k": 50 } } },
        { "match": { "content": "API 设计规范" } }
      ]
    }
  }
}

// 4. 聚合分析
{
  "size": 0,
  "query": { "term": { "catalog_id": "feishu_corp" } },
  "aggs": {
    "by_type": { "terms": { "field": "doc_type" } },
    "by_month": { "date_histogram": { "field": "created_at", "calendar_interval": "month" } }
  }
}
```

#### 4.1.5. 集群架构

```
┌─────────────────────────────────────────────────────────────────┐
│                    OpenSearch 集群架构                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Go Service (无状态)                                             │
│       │                                                          │
│       ├── 写入请求 ────────────> Coordinating Node               │
│       │                              │                           │
│       │                              ▼                           │
│       │                         Primary Shards                   │
│       │                              │                           │
│       │                              ▼ (异步复制)                 │
│       │                         Replica Shards                   │
│       │                              │                           │
│       └── 读取请求 ────────────> Coordinating Node               │
│                                      │                           │
│                                      ▼                           │
│                              Replica Shards (负载均衡)           │
│                                                                  │
│  节点角色:                                                       │
│  ├── Master Node: 集群管理                                       │
│  ├── Data Node: 存储 + 计算                                      │
│  ├── Coordinating Node: 查询路由                                 │
│  └── ML Node: GPU 加速 (可选)                                    │
│                                                                  │
│  分片策略:                                                       │
│  ├── Primary Shards: 按 catalog_id 路由                          │
│  └── Replica Shards: 每个 Primary 1-2 个副本                     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

#### 4.1.6. 配置示例

```yaml
# OpenSearch 连接配置
storage:
  type: opensearch
  opensearch:
    # 集群连接
    endpoints:
      - "https://opensearch-node1:9200"
      - "https://opensearch-node2:9200"
      - "https://opensearch-node3:9200"
    username: "${OPENSEARCH_USER}"
    password: "${OPENSEARCH_PASSWORD}"

    # 索引配置
    index:
      prefix: "vega_"                    # 索引前缀
      shards: 5                          # 主分片数
      replicas: 1                        # 副本数
      refresh_interval: "1s"             # 刷新间隔

    # 向量配置
    vector:
      engine: faiss                      # faiss | nmslib | lucene
      algorithm: hnsw                    # hnsw | ivf
      dimension: 1536                    # 向量维度
      space_type: cosinesimil            # 相似度计算
      ef_construction: 256               # 索引构建参数
      m: 16                              # HNSW 参数

    # 性能配置
    bulk:
      batch_size: 1000                   # 批量写入大小
      flush_interval: "5s"               # 刷新间隔
      concurrent_requests: 4             # 并发请求数

    # 查询配置
    query:
      timeout: "30s"                     # 查询超时
      max_result_window: 10000           # 最大返回条数
```

### 4.2. Query Transpiler (DSL 引擎)
这是系统的"翻译官"，负责将用户的 JSON DSL 翻译为不同引擎的语言：
*   **To SQL**: 针对 Table (MySQL), View (Trino/DuckDB)。
*   **To OpenSearch DSL**: 针对 Local Asset (OpenSearch)，翻译为向量检索 / 全文搜索 / 混合查询。
*   **To ES DSL**: 针对 Index Asset (外部 ES/OpenSearch)，翻译为 `{ "query": { "match": ... } }`。
*   **To PromQL**: 针对 Metric Asset，翻译为时序聚合查询。
这使得上层应用（特别是 AI Agent）只需掌握一种 DSL 即可查询所有资产。

---

---

