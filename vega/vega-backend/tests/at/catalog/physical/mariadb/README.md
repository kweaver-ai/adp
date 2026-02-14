# MariaDB Catalog AT 测试

## 概述

本目录包含 MariaDB Catalog 的验收测试（AT 测试）。MariaDB Catalog 是物理 Catalog，连接到实际的 MariaDB 数据源。

> **注意**：通用字段测试（name/description/tags 边界验证）已在 `catalog/logical` 中覆盖，此处仅测试 MariaDB 特有功能。

## 测试文件

| 文件 | 描述 |
|------|------|
| `catalog_test.go` | MariaDB Catalog CRUD 测试入口 |
| `builder.go` | MariaDB Payload 构建器 |

## 测试用例清单

### 创建测试（MD1xx）

#### 正向测试（MD101-MD119）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MD101 | 创建 MariaDB catalog - 基本场景 | 201 Created |
| MD102 | 创建后验证 connector_type 为 mariadb | connector_type = "mariadb" |
| MD103 | 创建后验证 type 为 physical | type = "physical" |
| MD104 | 创建 MariaDB catalog - 完整字段 | 201 Created |
| MD105 | 创建带 MariaDB 特定 options（charset/timeout） | 201 Created |
| MD106 | 创建后立即查询 | 查询返回一致数据 |
| MD107 | MariaDB 连接测试成功 | 200 OK |
| MD108 | 获取 MariaDB catalog 健康状态 | 200 OK |
| MD109 | 创建实例级 MariaDB catalog（不指定 database） | 201 Created |
| MD110 | 实例级 MariaDB catalog 连接测试成功 | 200 OK |

#### connector_config 负向测试（MD121-MD129）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MD121 | 缺少 host 字段 | 400 Bad Request |
| MD122 | 缺少 port 字段 | 400 Bad Request |
| MD123 | 缺少 user 字段 | 400 Bad Request |
| MD124 | 空用户名 | 400 Bad Request |
| MD125 | 错误密码 | 400 Bad Request |
| MD126 | 不存在的数据库 | 400 Bad Request |
| MD127 | 无效端口（非数字） | 400 Bad Request |
| MD128 | 超出范围端口（65536） | 400 Bad Request |
| MD129 | 负数端口 | 400 Bad Request |

#### 边界测试（MD131-MD139）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MD131 | port 边界值（1） | 201 Created |
| MD132 | port 边界值（65535） | 201 Created |
| MD133 | database 名称最大长度（64字符） | 201 Created |
| MD134 | database 名称超过最大长度 | 400 Bad Request |
| MD135 | host 为 IP 地址 | 201 Created |
| MD136 | host 为域名 | 201 Created |
| MD137 | 不指定 database（实例级连接） | 201 Created |
| MD138 | password 为空（无密码连接） | 201 Created 或 400 |

---

### 读取测试（MD2xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MD201 | 获取存在的 MariaDB catalog | 200 OK |
| MD202 | 列表查询 - 按 type 过滤 physical | 200 OK |
| MD203 | 列表查询 - 按 connector_type 过滤 mariadb | 200 OK |
| MD204 | 查询 catalog - 验证所有字段返回 | 200 OK |
| MD205 | 验证 connector_config.password 不返回 | password 字段不存在 |

---

### 更新测试（MD3xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MD301 | 整体更新 connector_config | 204 No Content |
| MD302 | 更新 connector_config 后连接测试 | 200 OK |
| MD303 | 更新 host 为无效地址 | 400 Bad Request |
| MD304 | 更新 port 为无效值 | 400 Bad Request |
| MD305 | 更新 password | 204 No Content |

---

### 删除测试（MD4xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MD401 | 删除 MariaDB catalog 后健康状态不可查 | 404 Not Found |
| MD402 | 删除 MariaDB catalog 后不能测试连接 | 404 Not Found |

---

### MariaDB 特有测试（MD5xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MD501 | MariaDB charset 选项测试（utf8mb4） | 201 Created |
| MD502 | MariaDB parseTime 选项测试 | 201 Created |
| MD503 | MariaDB loc 选项测试（时区） | 201 Created |
| MD504 | MariaDB timeout 选项测试 | 201 Created |
| MD505 | MariaDB SSL 连接测试 | 201 Created 或 400 |
| MD506 | MariaDB collation 选项测试 | 201 Created |

## 运行测试

```bash
# 运行所有 MariaDB Catalog 测试
go test -v ./tests/at/catalog/physical/mariadb/...

# 运行创建测试
go test -v ./tests/at/catalog/physical/mariadb/... -run TestMariaDBSpecificCreate

# 运行特定用例
go test -v ./tests/at/catalog/physical/mariadb/... -run MD101
```
