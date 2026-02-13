# MySQL Resource AT 测试

## 概述

本目录包含 MySQL Resource 的验收测试（AT 测试），运行通用 Resource 测试用例。

## 测试文件

| 文件 | 描述 |
|------|------|
| `resource_test.go` | 通用测试入口，运行所有 RMxxx 系列测试 |

## 测试用例清单

### 通用测试（RMxxx）

来自 `resource/internal/test_cases.go`，适用于所有 connector 类型。

#### 创建测试（RM101-109）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| RM101 | 创建 resource - 基本场景 | 201 Created |
| RM102 | 创建 resource - 最小字段 | 201 Created |
| RM103 | 创建 resource - 完整字段 | 201 Created |
| RM104 | 创建后立即查询 | 创建成功，查询返回一致数据 |
| RM105 | 创建带 category 的 resource | 201 Created |
| RM106 | Tags 数组测试（空/单个/多个） | 全部 201 Created |
| RM107 | 特殊字符名称测试 | 中文、连字符等均成功 |
| RM108 | 创建多个 resource，列表查询 | 列表返回正确 total |
| RM109 | 创建 resource - 带 description | 201 Created |

#### 负向测试（RM121-134）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| RM121 | 缺少必填字段 - name | 400 Bad Request |
| RM122 | 缺少必填字段 - catalog_id | 400 Bad Request |
| RM123 | 重复的 resource 名称（同一 catalog 内） | 409 Conflict |
| RM124 | 无效 JSON 格式 | 400 Bad Request |
| RM125 | 错误的 Content-Type | 406 Not Acceptable |
| RM126 | 超长 name 字段（>128字符） | 400 Bad Request |
| RM127 | 超长 description 字段（>1000字符） | 400 Bad Request |
| RM128 | name 为空字符串 | 400 Bad Request |
| RM129 | name 只有空格 | 201 Created（有效字符） |
| RM130 | tags 包含空字符串 | 400 Bad Request |
| RM131 | tags 包含非法字符 | 400 Bad Request |
| RM132 | 单个 tag 超长（41字符） | 400 Bad Request |
| RM133 | 无效的 catalog_id | 400 Bad Request |
| RM134 | 请求体为空 | 400 Bad Request |

#### 边界测试（RM141-149）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| RM141 | name 长度边界 - 1字符 | 201 Created |
| RM142 | name 长度边界 - 128字符 | 201 Created |
| RM143 | name 长度边界 - 129字符 | 400 Bad Request |
| RM144 | description 长度边界 - 1000字符 | 201 Created |
| RM145 | description 长度边界 - 1001字符 | 400 Bad Request |
| RM146 | tags 数量边界 - 5个 | 201 Created |
| RM147 | tags 数量边界 - 6个 | 400 Bad Request |
| RM148 | 单个 tag 长度边界 - 40字符 | 201 Created |
| RM149 | 单个 tag 长度边界 - 41字符 | 400 Bad Request |

#### 安全测试（RM161-163）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| RM161 | SQL 注入尝试 | 201 Created（安全处理） |
| RM162 | XSS 尝试 | 201 Created（安全处理） |
| RM163 | 路径遍历尝试 | 201 Created（安全处理） |

#### 读取测试（RM201-206）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| RM201 | 获取存在的 resource | 200 OK |
| RM202 | 获取不存在的 resource | 404 Not Found |
| RM203 | 列表分页测试 | 正确分页返回 |
| RM204 | 批量获取 resources | 200 OK |
| RM205 | 批量获取 - 部分 ID 不存在 | 404 Not Found |
| RM206 | 按 catalog_id 过滤列表 | 200 OK |

#### 更新测试（RM301-305）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| RM301 | 更新 resource 名称 | 204 No Content |
| RM302 | 更新 resource 注释 | 204 No Content |
| RM303 | 更新不存在的 resource | 404 Not Found |
| RM304 | 更新 tags | 204 No Content |
| RM305 | 更新 category | 204 No Content |

#### 删除测试（RM401-406）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| RM401 | 删除存在的 resource | 204 No Content |
| RM402 | 删除不存在的 resource | 404 Not Found |
| RM403 | 重复删除同一 resource | 404 Not Found |
| RM404 | 批量删除 resources | 204 No Content |
| RM405 | 批量删除 - 部分 ID 不存在 | 404 Not Found |
| RM406 | 删除后列表中不再显示 | 列表中不存在 |

#### 名称唯一性测试（RM501-503）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| RM501 | 同一 catalog 内重名冲突 | 409 Conflict |
| RM502 | 不同 catalog 内同名共存 | 201 Created |
| RM503 | 删除后重建同名 resource | 201 Created |

## 运行测试

```bash
# 运行所有 MySQL Resource 测试
go test -v ./tests/at/resource/mysql/...

# 运行特定测试套件
go test -v ./tests/at/resource/mysql/... -run TestMySQLResourceCommon

# 运行单个用例
go test -v ./tests/at/resource/mysql/... -run RM101
```
