# ECron (定时任务调度服务)

ECron 是一个分布式定时任务调度和执行服务,提供定时任务管理、即时任务执行和任务状态监控功能。它由两个核心服务组成:用于任务调度的分析服务和用于任务生命周期管理的管理服务。

[English Documentation](README.md)

## 核心架构

ECron 采用微服务架构,包含两个主要组件:

```
┌─────────────────────────────────────────────────────────┐
│                   分析服务 (Analysis)                    │
│              (任务调度与状态监控)                         │
│  - 基于 Cron 的任务调度                                  │
│  - 即时任务执行                                          │
│  - 任务状态追踪                                          │
│  - 消息队列消费                                          │
└─────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────┐
│                  管理服务 (Management)                   │
│              (任务 CRUD 与执行管理)                       │
│  - 任务信息管理                                          │
│  - 任务执行协调                                          │
│  - RESTful API 接口                                     │
│  - OAuth2 认证                                          │
└─────────────────────────────────────────────────────────┘
```

### 目录结构说明

- **`analysis/`**: 分析服务,负责任务调度和监控
- **`management/`**: 管理服务,负责任务 CRUD 操作和执行
- **`common/`**: 共享数据结构和工具
- **`utils/`**: 通用工具函数(数据库、HTTP、认证、日志等)
- **`mock/`**: 测试 Mock 实现
- **`migrations/`**: 数据库迁移脚本
- **`helm/`**: Kubernetes 部署 Helm Charts

## 主要功能

### 1. 任务调度
- **Cron 定时调度**: 支持 Cron 表达式配置周期性任务
- **即时执行**: 通过消息队列立即执行任务
- **任务生命周期管理**: 创建、更新、启用/禁用、删除任务
- **多节点支持**: 可选的多节点部署以实现高可用

### 2. 任务执行
- **HTTP/HTTPS 执行**: 通过 HTTP 请求执行任务(GET、POST、PUT、DELETE)
- **命令执行**: 执行 Shell 命令或脚本
- **Kubernetes Job 支持**: 创建和执行 Kubernetes 批处理任务
- **重试机制**: 可配置的失败重试次数

### 3. 任务监控
- **状态追踪**: 实时任务执行状态监控
- **执行历史**: 追踪任务执行历史和结果
- **Webhook 通知**: 将任务执行结果发送到配置的 Webhook
- **健康检查**: 服务健康状态和就绪状态检查端点

### 4. 认证与授权
- **OAuth2 集成**: 基于 Hydra 的 Token 认证
- **多租户支持**: 基于租户的数据隔离
- **权限控制**: 细粒度的任务操作访问控制

### 5. 消息队列集成
- **NSQ 支持**: 用于任务通知和状态更新的消息队列
- **Kafka 支持**: 可选的消息队列连接器
- **异步处理**: 非阻塞的任务执行和状态更新

## 技术栈

- **语言**: Go 1.24.0
- **Web 框架**: Gin
- **数据库**: MySQL/MariaDB
- **消息队列**: NSQ (主要), Kafka (可选)
- **Cron 库**: robfig/cron/v3
- **测试**: GoConvey, gomock
- **容器编排**: Kubernetes (Helm Charts)

## 前置要求

- Go 1.24+
- MySQL/MariaDB 5.7+
- NSQ 或 Kafka 消息队列
- OAuth2 服务 (Hydra)

## 配置说明

服务使用 YAML 配置文件 (`cronsvr.yaml`):

### 主要配置项

```yaml
# 系统设置
lang: zh_CN                    # 系统语言
job_failures: 3                # 任务失败最大重试次数
webhook: /api/cronsvr/v1/...   # 任务结果 Webhook 路径

# 服务 ID
analysis_service_id: <uuid>    # 分析服务 ID
management_service_id: <uuid>  # 管理服务 ID

# 消息队列
mq_host: nsqlookupd.proton-mq.svc.cluster.local
mq_port: 4161
mq_connector_type: nsq         # nsq 或 kafka

# 定时服务
cron_addr: 0.0.0.0
cron_port: 12345
cron_protocol: http
multi_node: false              # 启用多节点部署

# 数据库
db_addr: localhost
db_port: 3306
db_name: ecron
user_name: root
user_pwd: password
max_open_conns: 30

# OAuth2 服务
oauth_public_addr: 127.0.0.1
oauth_public_port: 4444
oauth_admin_addr: 127.0.0.1
oauth_admin_port: 4445

# SSL 证书 (可选)
ssl_on: false
cert_file: /path/to/cert.crt
key_file: /path/to/cert.key
```

## API 文档

### 管理服务端点

#### 任务管理
- `GET /api/ecron-management/v1/jobtotal` - 获取任务总数
- `GET /api/ecron-management/v1/job` - 查询任务信息(支持分页)
- `POST /api/ecron-management/v1/job` - 创建新任务
- `PUT /api/ecron-management/v1/job/:job_id` - 更新任务信息
- `DELETE /api/ecron-management/v1/job/:job_id` - 删除任务

#### 任务状态
- `GET /api/ecron-management/v1/jobstatus/:job_id` - 获取任务执行状态
- `PUT /api/ecron-management/v1/jobstatus/:job_id` - 更新任务状态

#### 任务控制
- `PUT /api/ecron-management/v1/job/:job_id/enable` - 启用/禁用任务
- `PUT /api/ecron-management/v1/job/:job_id/notify` - 更新任务通知设置

#### 任务执行
- `POST /api/ecron-management/v1/jobexecution` - 立即执行任务

#### 健康检查
- `GET /health/ready` - 就绪探针
- `GET /health/alive` - 存活探针

### 任务信息结构

```json
{
  "job_id": "unique-job-id",
  "job_name": "我的定时任务",
  "job_cron_time": "0 0 * * *",
  "job_type": "cron",
  "job_context": {
    "mode": "http",
    "exec": "https://api.example.com/endpoint",
    "info": {
      "method": "POST",
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "key": "value"
      }
    },
    "notify": {
      "webhook": "https://callback.example.com/webhook"
    }
  },
  "tenant_id": "tenant-123",
  "enabled": true,
  "remarks": "任务描述"
}
```

## 运行单元测试

### 测试前置条件

1. **缺少库文件**: 如果运行单元测试时出现缺少 `libtlq9client.so` 文件的错误,可将该文件放到 `lib64` 目录下解决。

2. **消息中间件依赖**: 如果 `go get` 宝兰德消息中间件依赖包失败,请拉取 `go-msq` 仓库代码,执行脚本:
   ```bash
   ./go-msq/besmq/deploy_besmq_libs.sh
   ```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 运行特定包的测试
go test ./analysis
go test ./management
go test ./utils
```

## 部署

### 使用 Helm

```bash
cd helm
helm install ecron . -f values.yaml
```

### 环境变量

服务也可以通过环境变量进行配置,环境变量会覆盖 YAML 配置:

- `ECRON_DB_ADDR`: 数据库地址
- `ECRON_DB_PORT`: 数据库端口
- `ECRON_DB_NAME`: 数据库名称
- `ECRON_MQ_HOST`: 消息队列主机
- `ECRON_CRON_PORT`: 定时服务端口

## 开发指南

1. **代码风格**: 遵循 Go 标准代码规范,使用 `golangci-lint` 进行代码检查
2. **测试**: 使用 `go.uber.org/mock` 编写单元测试
3. **Mock 生成**: 使用 `go generate` 重新生成 Mock:
   ```bash
   go generate ./...
   ```

## 架构详解

### 分析服务 (Analysis Service)

分析服务负责:
- 从数据库加载和刷新任务列表
- 使用 Cron 库调度基于 Cron 表达式的任务
- 从消息队列消费即时任务执行消息
- 监控和更新任务执行状态
- 将任务执行结果发送到配置的 Webhook

### 管理服务 (Management Service)

管理服务提供:
- 任务 CRUD 操作的 RESTful API
- 任务执行协调
- 基于 OAuth2 的认证和授权
- 任务信息和状态的数据库持久化
- 任务通知的消息队列发布

### 执行模块 (Execution Module)

执行模块支持:
- **HTTP 模式**: 执行可配置方法、请求头和请求体的 HTTP 请求
- **EXE 模式**: 执行 Shell 命令或创建 Kubernetes 任务
- **Kubernetes 集成**: 在 Kubernetes 集群中创建批处理任务
