# MariaDB Resource AT 测试

## 概述

本目录包含 MariaDB Resource 的验收测试（AT 测试）。MariaDB Resource 是物理 Resource，通过 Catalog 的 Discovery 机制自动发现，关联到实际的 MariaDB 数据源。

> **注意**：MariaDB Resource 与 Dataset Resource 有本质区别：
> - **不允许创建**：Resource 通过 Catalog Discovery 自动发现，不可手动创建
> - **不允许删除**：物理 Resource 由数据源管理，不可删除
> - **受限更新**：仅允许修改显示名称、描述、标签等元数据，不可修改 schema 等核心字段

## 测试文件

| 文件 | 描述 |
|------|------|
| `resource_test.go` | MariaDB Resource 测试入口 |

## 测试用例清单

### Discovery 测试（MR1xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR101 | Catalog Discovery 后验证 Resource 存在 | Resource 列表包含发现的表 |
| MR102 | 验证 Resource 的 connector_type 为 mariadb | connector_type = "mariadb" |
| MR103 | 验证 Resource 的 category 为 table | category = "table" |
| MR104 | 验证 Resource 的 schema_definition 正确返回 | schema_definition 包含字段列表 |
| MR105 | 验证 Resource 的 original_name 与数据库表名一致 | original_name 正确映射 |
| MR106 | Discovery 后验证多个表都被发现 | 所有表都在 Resource 列表中 |

#### Discovery 负向测试（MR121-MR125）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR121 | 手动创建 MariaDB Resource | 400 Bad Request（不允许创建） |
| MR122 | 无效 Catalog ID 的 Discovery | 404 Not Found |
| MR123 | 连接失败的 Catalog Discovery | 500 Internal Error |
| MR124 | 权限不足的数据库 Discovery | 返回可见的表（部分或空） |
| MR125 | 空数据库的 Discovery | Resource 列表为空 |

---

### 读取测试（MR2xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR201 | 获取存在的 MariaDB Resource | 200 OK |
| MR202 | 获取不存在的 Resource | 404 Not Found |
| MR203 | 列表查询 - 按 catalog_id 过滤 | 200 OK |
| MR204 | 列表查询 - 按 category=table 过滤 | 200 OK |
| MR205 | 列表查询 - 按 connector_type=mariadb 过滤 | 200 OK |
| MR206 | 列表分页测试 | 正确分页返回 |
| MR207 | 验证 Resource 包含完整的 schema 信息 | schema_definition 完整 |
| MR208 | 验证 Resource 包含正确的 catalog 关联 | catalog_id 正确 |

---

### 更新测试（MR3xx）

#### 允许的更新（MR301-MR305）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR301 | 更新 Resource 显示名称 | 204 No Content |
| MR302 | 更新 Resource 描述 | 204 No Content |
| MR303 | 更新 Resource 标签 | 204 No Content |
| MR304 | 同时更新名称、描述、标签 | 204 No Content |
| MR305 | 更新后验证修改生效 | 查询返回更新后的值 |

#### 禁止的更新（MR321-MR328）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR321 | 尝试修改 schema_definition | 400 Bad Request |
| MR322 | 尝试修改 connector_type | 400 Bad Request |
| MR323 | 尝试修改 category | 400 Bad Request |
| MR324 | 尝试修改 catalog_id | 400 Bad Request |
| MR325 | 尝试修改 config | 400 Bad Request |
| MR326 | 尝试修改 original_name | 400 Bad Request |
| MR327 | 更新不存在的 Resource | 404 Not Found |
| MR328 | 空更新请求体 | 400 Bad Request |

---

### 删除测试（MR4xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR401 | 尝试删除 MariaDB Resource | 400 Bad Request（不允许删除） |
| MR402 | 尝试批量删除 MariaDB Resource | 400 Bad Request（不允许删除） |

---

### 数据查询测试（MR5xx）

#### 基础查询（MR501-MR510）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR501 | 查询 Resource 数据 - 基本场景 | 200 OK，返回数据列表 |
| MR502 | 查询 Resource 数据 - 分页 | 200 OK，正确分页 |
| MR503 | 查询 Resource 数据 - 指定字段 | 200 OK，仅返回指定字段 |
| MR504 | 查询 Resource 数据 - 排序 | 200 OK，正确排序 |
| MR505 | 查询 Resource 数据 - 限制返回条数 | 200 OK，返回指定条数 |
| MR506 | 查询空表数据 | 200 OK，entries 为空 |
| MR507 | 查询不存在的 Resource 数据 | 404 Not Found |
| MR508 | 验证返回数据字段与 schema 一致 | 字段类型匹配 |

#### 过滤条件查询（MR511-MR525）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR511 | 查询 - 等于条件（eq） | 200 OK，正确过滤 |
| MR512 | 查询 - 不等于条件（neq） | 200 OK，正确过滤 |
| MR513 | 查询 - 大于条件（gt） | 200 OK，正确过滤 |
| MR514 | 查询 - 大于等于条件（gte） | 200 OK，正确过滤 |
| MR515 | 查询 - 小于条件（lt） | 200 OK，正确过滤 |
| MR516 | 查询 - 小于等于条件（lte） | 200 OK，正确过滤 |
| MR517 | 查询 - IN 条件 | 200 OK，正确过滤 |
| MR518 | 查询 - NOT IN 条件 | 200 OK，正确过滤 |
| MR519 | 查询 - LIKE 条件（模糊匹配） | 200 OK，正确过滤 |
| MR520 | 查询 - IS NULL 条件 | 200 OK，正确过滤 |
| MR521 | 查询 - IS NOT NULL 条件 | 200 OK，正确过滤 |
| MR522 | 查询 - BETWEEN 条件 | 200 OK，正确过滤 |
| MR523 | 查询 - 组合条件（AND） | 200 OK，正确过滤 |
| MR524 | 查询 - 组合条件（OR） | 200 OK，正确过滤 |
| MR525 | 查询 - 嵌套组合条件 | 200 OK，正确过滤 |

#### 查询边界测试（MR531-MR536）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR531 | 查询 - offset 超出范围 | 200 OK，entries 为空 |
| MR532 | 查询 - limit 最大值 | 200 OK |
| MR533 | 查询 - limit=0 | 400 Bad Request |
| MR534 | 查询 - 无效排序字段 | 400 Bad Request |
| MR535 | 查询 - 无效过滤字段 | 400 Bad Request |
| MR536 | 查询 - 无效过滤操作符 | 400 Bad Request |

---

### 统计查询测试（MR6xx）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| MR601 | 统计查询 - 总数统计 | 200 OK，返回 total_count |
| MR602 | 统计查询 - 分组统计 | 200 OK，返回分组结果 |
| MR603 | 统计查询 - 聚合函数（SUM/AVG/MIN/MAX） | 200 OK，返回聚合结果 |
| MR604 | 统计查询 - 带过滤条件 | 200 OK，返回过滤后统计 |

## 运行测试

```bash
# 运行所有 MariaDB Resource 测试
go test -v ./tests/at/resource/table/mairadb/...

# 运行 Discovery 测试
go test -v ./tests/at/resource/table/mairadb/... -run TestMariaDBResourceDiscovery

# 运行读取测试
go test -v ./tests/at/resource/table/mairadb/... -run TestMariaDBResourceRead

# 运行更新测试
go test -v ./tests/at/resource/table/mairadb/... -run TestMariaDBResourceUpdate

# 运行数据查询测试
go test -v ./tests/at/resource/table/mairadb/... -run TestMariaDBResourceQuery

# 运行特定用例（通过用例ID）
go test -v ./tests/at/resource/table/mairadb/... -run MR101
```
