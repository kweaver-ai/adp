# OSS 网关 API 文档

## 基础信息

**基础 URL:** `http://localhost:8080/api/v1`

**协议:** HTTP/HTTPS

**请求格式:** JSON

**响应格式:** JSON

---

## 响应规范

所有 API 响应遵循统一格式：

### 成功响应
```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 错误响应
```json
{
  "code": 400001,
  "message": "Invalid parameter",
  "cause": "storage_id is required"
}
```

### 错误码表

| 错误码   | HTTP 状态码 | 描述              |
|---------|------------|-------------------|
| 0       | 200        | 成功               |
| 400000  | 400        | 错误的请求          |
| 400001  | 400        | 无效的参数          |
| 400002  | 400        | 无效的文件大小       |
| 404000  | 404        | 未找到             |
| 404001  | 404        | 存储配置未找到       |
| 500000  | 500        | 内部服务器错误       |
| 500001  | 500        | 连接失败            |

---

## 存储管理 API

### 1. 获取存储列表

获取所有存储配置。

**接口地址:** `GET /storages`

**查询参数:**

| 参数名       | 类型    | 必填 | 说明                                       |
|-------------|---------|------|-------------------------------------------|
| vendor_type | string  | 否   | 厂商类型过滤 (OSS/OBS/ECEPH)               |
| enabled     | boolean | 否   | 启用状态过滤 (true/false)                  |
| page        | int     | 否   | 页码，从1开始，默认1                        |
| size        | int     | 否   | 每页大小，默认10，最大1000                  |
| order       | string  | 否   | 排序方向 (asc/desc)，默认 desc             |
| rule        | string  | 否   | 排序字段 (create_time/update_time/storage_name)，默认 update_time |
| name        | string  | 否   | 存储名称模糊搜索                            |

**响应示例:**
```json
{
  "code": 0,
  "message": "success",
  "count": 1,
  "data": [
    {
      "storage_id": "A1B2C3D4E5F6G7H8",
      "storage_name": "阿里云存储1",
      "vendor_type": "OSS",
      "endpoint": "https://oss-cn-hangzhou.aliyuncs.com",
      "bucket_name": "my-bucket",
      "region": "cn-hangzhou",
      "is_default": true,
      "is_enabled": true,
      "internal_endpoint": "",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### 2. 获取单个存储配置

根据 ID 获取单个存储配置。

**接口地址:** `GET /storages/:id`

**路径参数:**

| 参数名 | 类型   | 必填 | 说明     |
|--------|--------|------|----------|
| id     | string | 是   | 存储 ID  |

**响应示例:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "storage_id": "A1B2C3D4E5F6G7H8",
    "storage_name": "阿里云存储1",
    "vendor_type": "OSS",
    "endpoint": "https://oss-cn-hangzhou.aliyuncs.com",
    "bucket_name": "my-bucket",
    "region": "cn-hangzhou",
    "is_default": true,
    "is_enabled": true,
    "internal_endpoint": "",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 3. 创建存储配置

创建新的存储配置。

**接口地址:** `POST /storages`

**请求体:**
```json
{
  "storage_name": "阿里云存储1",
  "vendor_type": "OSS",
  "endpoint": "https://oss-cn-hangzhou.aliyuncs.com",
  "bucket_name": "my-bucket",
  "access_key_id": "LTAI5xxx",
  "access_key_secret": "your-secret",
  "region": "cn-hangzhou",
  "is_default": false,
  "internal_endpoint": ""
}
```

**字段说明:**

| 字段名            | 类型    | 必填 | 说明                                      |
|------------------|---------|------|------------------------------------------|
| storage_name     | string  | 是   | 存储配置的显示名称                         |
| vendor_type      | string  | 是   | 厂商类型 (OSS/OBS/ECEPH)                  |
| endpoint         | string  | 是   | 服务端点 URL (必须以 http:// 或 https:// 开头) |
| bucket_name      | string  | 是   | 存储桶名称                                |
| access_key_id    | string  | 是   | 访问密钥 ID                               |
| access_key_secret| string  | 是   | 访问密钥                                  |
| region           | string  | 条件 | 区域标识符 (OSS/OBS必填，ECEPH可选)        |
| is_default       | boolean | 否   | 是否设为默认存储                           |
| internal_endpoint| string  | 否   | 内网访问端点                              |

**响应示例:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "storage_id": "A1B2C3D4E5F6G7H8",
    "status": "ok",
    "id": "A1B2C3D4E5F6G7H8"
  }
}
```

---

### 4. 更新存储配置

更新现有的存储配置。

**接口地址:** `PUT /storages/:id`

**路径参数:**

| 参数名 | 类型   | 必填 | 说明     |
|--------|--------|------|----------|
| id     | string | 是   | 存储 ID  |

**请求体:** (所有字段可选)
```json
{
  "storage_name": "新名称",
  "is_enabled": true
}
```

**响应示例:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok",
    "id": "A1B2C3D4E5F6G7H8"
  }
}
```

---

### 5. 删除存储配置

删除指定的存储配置。

**接口地址:** `DELETE /storages/:id`

**路径参数:**

| 参数名 | 类型   | 必填 | 说明     |
|--------|--------|------|----------|
| id     | string | 是   | 存储 ID  |

**响应示例:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok",
    "id": "A1B2C3D4E5F6G7H8"
  }
}
```

---

### 6. 检查存储连接

测试与对象存储服务的连接。

**接口地址:** `POST /storages/:id/check`

**路径参数:**

| 参数名 | 类型   | 必填 | 说明     |
|--------|--------|------|----------|
| id     | string | 是   | 存储 ID  |

**响应示例:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "connected": true
  }
}
```

---

## 对象元数据 API

### 7. 获取对象元数据 URL

获取对象 HEAD 请求的预签名 URL。

**接口地址:** `GET /head/:storageId/:key`

**路径参数:**

| 参数名     | 类型   | 必填 | 说明                      |
|-----------|--------|------|--------------------------|
| storageId | string | 是   | 存储 ID                   |
| key       | string | 是   | 对象 Key (需 URL 编码)     |

**查询参数:**

| 参数名           | 类型    | 必填 | 说明                          |
|-----------------|---------|------|------------------------------|
| internal_request| boolean | 否   | 是否使用内网端点，默认 false    |
| expires         | int64   | 否   | URL 过期时间（秒），默认 3600  |

**响应示例:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "method": "HEAD",
    "url": "https://my-bucket.oss-cn-hangzhou.aliyuncs.com/path/to/file.txt?...",
    "headers": {}
  }
}
```

---

### 8. 批量获取对象元数据 URL

批量获取多个对象的 HEAD 请求预签名 URL。

**接口地址:** `POST /head/:storageId`

**路径参数:**

| 参数名     | 类型   | 必填 | 说明     |
|-----------|--------|------|----------|
| storageId | string | 是   | 存储 ID  |

**请求体:**
```json
{
  "keys": [
    "path/to/file1.txt",
    "path/to/file2.txt",
    "path/to/file3.txt"
  ],
  "internal_request": false,
  "expires": 3600
}
```

**字段说明:**

| 字段名           | 类型    | 必填 | 说明                                |
|-----------------|---------|------|-------------------------------------|
| keys            | array   | 是   | 对象 Key 列表，最多 100 个           |
| internal_request| boolean | 否   | 是否使用内网端点，默认 false          |
| expires         | int64   | 否   | URL 过期时间（秒），默认 3600        |

**响应示例:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "file1.txt": {
      "method": "HEAD",
      "url": "https://...",
      "headers": {}
    },
    "file2.txt": {
      "method": "HEAD",
      "url": "https://...",
      "headers": {}
    }
  }
}
```

---

## 单文件上传 API

### 9. 获取单文件上传 URL

获取单文件上传的预签名 URL。

**接口地址:** `GET /upload/:storageId/:key`

**路径参数:**

| 参数名     | 类型   | 必填 | 说明                      |
|-----------|--------|------|--------------------------|
| storageId | string | 是   | 存储 ID                   |
| key       | string | 是   | 对象 Key (需 URL 编码)     |

**查询参数:**

| 参数名           | 类型    | 必填 | 说明                    |
|-----------------|---------|------|-----------------------|
| request_method  | string  | 否   | 上传方式 (POST/PUT)，默认 PUT |
| expires         | int64   | 否   | URL 过期时间（秒）           |
| internal_request| boolean | 否   | 是否使用内网端点，默认 false     |

**响应示例 (PUT):**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "method": "PUT",
    "url": "https://my-bucket.oss-cn-hangzhou.aliyuncs.com/test/file.txt?...",
    "headers": {
      "Content-Type": "application/octet-stream"
    },
    "form_field": {}
  }
}
```

**响应示例 (POST):**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "method": "POST",
    "url": "https://my-bucket.oss-cn-hangzhou.aliyuncs.com/",
    "headers": {},
    "form_field": {
      "OSSAccessKeyId": "LTAI5xxx",
      "policy": "eyJleHBpcmF0aW9uIjoi...",
      "Signature": "xxx",
      "key": "test/file.txt",
      "x-oss-signature-version": "OSS2",
      "x-oss-algorithm": "OSS2-HMAC-SHA256",
      "x-oss-credential": "...",
      "x-oss-date": "..."
    }
  }
}
```

---

## 分片上传 API

### 10. 初始化分片上传

初始化分片上传任务。

**接口地址:** `GET /initmultiupload/:storageId/:key`

**路径参数:**

| 参数名     | 类型   | 必填 | 说明                      |
|-----------|--------|------|--------------------------|
| storageId | string | 是   | 存储 ID                   |
| key       | string | 是   | 对象 Key (需 URL 编码)     |

**查询参数:**

| 参数名 | 类型  | 必填 | 说明                |
|--------|-------|------|---------------------|
| size   | int64 | 是   | 文件总大小（字节）    |

**响应示例:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "upload_id": "0004B9894A22E5B1F2A2A778BC7D4C19",
    "part_size": 5242880,
    "key": "large-file.zip"
  }
}
```

**字段说明:**

| 字段名     | 类型   | 说明                          |
|-----------|--------|------------------------------|
| upload_id | string | 上传会话 ID                   |
| part_size | int64  | 建议的分片大小（字节）          |
| key       | string | 对象 Key                     |

---

### 11. 获取分片上传 URL

获取各个分片的上传 URL。

**接口地址:** `POST /uploadpart/:storageId/:key`

**路径参数:**

| 参数名     | 类型   | 必填 | 说明                      |
|-----------|--------|------|--------------------------|
| storageId | string | 是   | 存储 ID                   |
| key       | string | 是   | 对象 Key (需 URL 编码)     |

**请求体:**
```json
{
  "upload_id": "0004B9894A22E5B1F2A2A778BC7D4C19",
  "part_id": [1, 2, 3, 4, 5],
  "internal_request": false
}
```

**字段说明:**

| 字段名           | 类型    | 必填 | 说明                              |
|-----------------|---------|------|----------------------------------|
| upload_id       | string  | 是   | 初始化时返回的 upload_id           |
| part_id         | array   | 是   | 分片编号列表 (1-10000)             |
| internal_request| boolean | 否   | 是否使用内网端点，默认 false        |

**响应示例:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "authrequest": {
      "1": {
        "method": "PUT",
        "url": "https://...?partNumber=1&uploadId=xxx",
        "headers": {
          "Content-Type": "application/octet-stream"
        }
      },
      "2": {
        "method": "PUT",
        "url": "https://...?partNumber=2&uploadId=xxx",
        "headers": {
          "Content-Type": "application/octet-stream"
        }
      }
    }
  }
}
```

---

### 12. 完成分片上传

完成分片上传并合并文件。

**接口地址:** `POST /completeupload/:storageId/:key`

**路径参数:**

| 参数名     | 类型   | 必填 | 说明                      |
|-----------|--------|------|--------------------------|
| storageId | string | 是   | 存储 ID                   |
| key       | string | 是   | 对象 Key (需 URL 编码)     |

**查询参数:**

| 参数名     | 类型   | 必填 | 说明        |
|-----------|--------|------|-------------|
| upload_id | string | 是   | 上传 ID     |

**请求体:**
```json
{
  "1": "\"5eb63bbbe01eeed093cb22bb8f5acdc3\"",
  "2": "\"5eb63bbbe01eeed093cb22bb8f5acdc4\"",
  "3": "\"5eb63bbbe01eeed093cb22bb8f5acdc5\""
}
```

**注意:** Key 为分片编号（字符串），Value 为对应分片的 ETag（必须包含双引号）。

**响应示例:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "method": "POST",
    "url": "https://...?uploadId=xxx",
    "headers": {
      "Content-Type": "application/xml"
    },
    "request_body": "<CompleteMultipartUpload><Part><PartNumber>1</PartNumber><ETag>\"etag1\"</ETag></Part>...</CompleteMultipartUpload>"
  }
}
```

---

## 下载 API

### 13. 获取下载 URL

获取文件下载的预签名 URL。

**接口地址:** `GET /download/:storageId/:key`

**路径参数:**

| 参数名     | 类型   | 必填 | 说明                      |
|-----------|--------|------|--------------------------|
| storageId | string | 是   | 存储 ID                   |
| key       | string | 是   | 对象 Key (需 URL 编码)     |

**查询参数:**

| 参数名           | 类型    | 必填 | 说明                              |
|-----------------|---------|------|----------------------------------|
| expires         | int64   | 否   | URL 过期时间（秒）                 |
| save_name       | string  | 否   | 下载时的文件名 (需 URL 编码)        |
| internal_request| boolean | 否   | 是否使用内网端点，默认 false        |

**响应示例:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "method": "GET",
    "url": "https://my-bucket.oss-cn-hangzhou.aliyuncs.com/test/file.txt?...",
    "headers": {}
  }
}
```

---

## 删除 API

### 14. 获取删除 URL

获取文件删除的预签名 URL。

**接口地址:** `GET /delete/:storageId/:key`

**路径参数:**

| 参数名     | 类型   | 必填 | 说明                      |
|-----------|--------|------|--------------------------|
| storageId | string | 是   | 存储 ID                   |
| key       | string | 是   | 对象 Key (需 URL 编码)     |

**查询参数:**

| 参数名           | 类型    | 必填 | 说明                          |
|-----------------|---------|------|------------------------------|
| expires         | int64   | 否   | URL 过期时间（秒）             |
| internal_request| boolean | 否   | 是否使用内网端点，默认 false    |

**响应示例:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "method": "DELETE",
    "url": "https://my-bucket.oss-cn-hangzhou.aliyuncs.com/test/file.txt?...",
    "headers": {}
  }
}
```

---

## 健康检查 API

### 15. 就绪检查

检查服务是否准备好接收流量（Kubernetes Readiness Probe）。

**接口地址:** `GET /health/ready`

**响应示例（就绪）:**
```json
{
  "status": "ok",
  "checks": {
    "database": "ok"
  }
}
```

**响应示例（未就绪）:**
```json
{
  "status": "not_ready",
  "checks": {
    "database": "failed"
  },
  "message": "Some services are not ready"
}
```

---

### 16. 存活检查

检查服务进程是否正常运行（Kubernetes Liveness Probe）。

**接口地址:** `GET /health/alive`

**响应示例:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

---

## 注意事项

### URL 编码

对象 Key 中的特殊字符需要进行 URL 编码：
- `/` → `%2F`
- ` ` (空格) → `%20`
- `+` → `%2B`
- `?` → `%3F`
- `&` → `%26`
- `=` → `%3D`

### ETag 格式

完成分片上传时，ETag 必须包含双引号：
```json
{
  "1": "\"5eb63bbbe01eeed093cb22bb8f5acdc3\"",
  "2": "\"5eb63bbbe01eeed093cb22bb8f5acdc4\""
}
```

### 文件大小限制

- 单文件上传：最大 5GB（取决于厂商限制）
- 分片上传：推荐用于 > 100MB 的文件
- 最大文件大小：50TB（通过分片上传）

### URL 过期时间

- 默认：3600 秒（1 小时）
- 最小：1 秒
- 最大：604800 秒（7 天）

### 内网与外网

使用 `internal_request=true` 的场景：
- 网关和对象存储在同一 VPC 内
- 需要更快的上传/下载速度
- 希望节省带宽成本

---