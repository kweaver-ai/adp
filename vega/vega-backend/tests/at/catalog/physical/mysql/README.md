# MySQL Catalog AT 测试

## 概述

本目录包含 MySQL Catalog 的验收测试（AT 测试）。MySQL Catalog 是物理 Catalog，连接到实际的 MySQL 数据源。

> **注意**：通用字段测试（name/description/tags 边界验证）已在 `catalog/logical` 中覆盖，此处仅测试 MySQL 特有功能。

## 测试文件

| 文件 | 描述 |
|------|------|
| `catalog_test.go` | MySQL Catalog CRUD 测试入口 |
| `builder.go` | MySQL Payload 构建器 |

## 测试用例清单

### 创建测试（MY1xx）

#### 正向测试（MY101-MY119）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MY101 | 创建 MySQL catalog - 基本场景 | 201 Created |
| MY102 | 创建后验证 connector_type 为 mysql | connector_type = "mysql" |
| MY103 | 创建后验证 type 为 physical | type = "physical" |
| MY104 | 创建 MySQL catalog - 完整字段 | 201 Created |
| MY105 | 创建带 MySQL 特定 options（charset/timeout） | 201 Created |
| MY106 | 创建后立即查询 | 查询返回一致数据 |
| MY107 | MySQL 连接测试成功 | 200 OK |
| MY108 | 获取 MySQL catalog 健康状态 | 200 OK |
| MY109 | 创建实例级 MySQL catalog（不指定 database） | 201 Created |
| MY110 | 实例级 MySQL catalog 连接测试成功 | 200 OK |

#### connector_config 负向测试（MY121-MY129）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MY121 | 缺少 host 字段 | 400 Bad Request |
| MY122 | 缺少 port 字段 | 400 Bad Request |
| MY123 | 缺少 user 字段 | 400 Bad Request |
| MY124 | 空用户名 | 400 Bad Request |
| MY125 | 错误密码 | 400 Bad Request |
| MY126 | 不存在的数据库 | 400 Bad Request |
| MY127 | 无效端口（非数字） | 400 Bad Request |
| MY128 | 超出范围端口（65536） | 400 Bad Request |
| MY129 | 负数端口 | 400 Bad Request |

#### 边界测试（MY131-MY139）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MY131 | port 边界值（1） | 201 Created |
| MY132 | port 边界值（65535） | 201 Created |
| MY133 | database 名称最大长度（64字符） | 201 Created |
| MY134 | database 名称超过最大长度 | 400 Bad Request |
| MY135 | host 为 IP 地址 | 201 Created |
| MY136 | host 为域名 | 201 Created |
| MY137 | 不指定 database（实例级连接） | 201 Created |
| MY138 | password 为空（无密码连接） | 201 Created 或 400 |

---

### 读取测试（MY2xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MY201 | 获取存在的 MySQL catalog | 200 OK |
| MY202 | 列表查询 - 按 type 过滤 physical | 200 OK |
| MY203 | 列表查询 - 按 connector_type 过滤 mysql | 200 OK |
| MY204 | 查询 catalog - 验证所有字段返回 | 200 OK |
| MY205 | 验证 connector_config.password 不返回 | password 字段不存在 |

---

### 更新测试（MY3xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MY301 | 整体更新 connector_config | 204 No Content |
| MY302 | 更新 connector_config 后连接测试 | 200 OK |
| MY303 | 更新 host 为无效地址 | 400 Bad Request |
| MY304 | 更新 port 为无效值 | 400 Bad Request |
| MY305 | 更新 password | 204 No Content |

---

### 删除测试（MY4xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MY401 | 删除 MySQL catalog 后健康状态不可查 | 404 Not Found |
| MY402 | 删除 MySQL catalog 后不能测试连接 | 404 Not Found |

---

### MySQL 特有测试（MY5xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MY501 | MySQL charset 选项测试（utf8mb4） | 201 Created |
| MY502 | MySQL parseTime 选项测试 | 201 Created |
| MY503 | MySQL loc 选项测试（时区） | 201 Created |
| MY504 | MySQL timeout 选项测试 | 201 Created |
| MY505 | MySQL SSL 连接测试 | 201 Created 或 400 |
| MY506 | MySQL collation 选项测试 | 201 Created |

## 运行测试

```bash
# 运行所有 MySQL Catalog 测试
go test -v ./tests/at/catalog/physical/mysql/...

# 运行创建测试
go test -v ./tests/at/catalog/physical/mysql/... -run TestMySQLSpecificCreate

# 运行特定用例
go test -v ./tests/at/catalog/physical/mysql/... -run MY101
```
