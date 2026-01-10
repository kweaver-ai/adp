# Workflow Service (工作流服务)

Workflow Service 是一个基于 Spring Boot 和 Activiti 构建的综合工作流和审核管理系统。它提供工作流定义、流程实例管理、审核流程编排和基于部门的审核人规则管理功能。

[English Documentation](README.md)

## 核心架构

Workflow Service 采用多模块 Maven 架构:

```
┌─────────────────────────────────────────────────────────┐
│                    workflow-rest                        │
│              (REST API 与 Web 层)                       │
│  - RESTful API 端点                                     │
│  - 请求/响应处理                                         │
│  - 认证与授权                                            │
│  - 异常处理                                              │
└─────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────┐
│                   workflow-code                         │
│              (核心业务逻辑)                              │
│  - 工作流引擎集成 (Activiti)                             │
│  - 流程定义管理                                          │
│  - 任务分配与执行                                        │
│  - 审核规则管理                                          │
└─────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────┐
│                  doc-audit-rest                         │
│              (审核管理模块)                              │
│  - 审核流程编排                                          │
│  - 审核策略管理                                          │
│  - 审核人分配                                            │
└─────────────────────────────────────────────────────────┘
```

### 模块结构

- **`workflow-rest/`**: REST API 层,包含控制器、配置和异常处理
- **`workflow-code/`**: 核心业务逻辑和 Activiti 工作流引擎集成
- **`doc-audit-rest/`**: 通用审核流程管理模块
- **`helm/`**: Kubernetes 部署 Helm Charts
- **`migrations/`**: 数据库迁移脚本

## 主要功能

### 1. 工作流管理
- **流程定义**: 创建、更新、部署和删除工作流定义
- **BPMN 2.0 支持**: 完全支持 BPMN 2.0 工作流建模
- **流程部署**: 将工作流定义部署到 Activiti 引擎
- **版本控制**: 管理工作流定义的多个版本
- **流程复制**: 克隆现有工作流定义

### 2. 流程实例管理
- **实例创建**: 启动新的工作流实例
- **实例追踪**: 监控流程实例执行状态
- **任务分配**: 将任务分配给用户或组
- **实例取消**: 取消正在运行的流程实例
- **执行历史**: 追踪完整的流程执行历史

### 3. 审核管理
- **审核策略**: 配置各种业务场景的审核策略
- **审核人分配**: 基于部门或角色分配审核人
- **审核规则**: 定义基于部门的审核人规则
- **免审部门**: 配置免于审核的部门
- **第三方审核**: 与外部审核系统集成

### 4. 任务管理
- **任务分配**: 将任务分配给特定用户
- **会签支持**: 多用户审批(会签)任务
- **任务完成**: 通过批准或拒绝完成任务
- **任务委托**: 将任务委托给其他用户
- **任务过滤**: 按发起人、受理人或状态过滤任务

### 5. 集成能力
- **OAuth2 认证**: 基于 OAuth2 服务的 Token 认证
- **用户管理**: 与外部用户管理系统集成
- **消息队列**: Kafka/NSQ 支持异步处理
- **Redis 缓存**: 分布式缓存以优化性能

### 6. 可观测性
- **Actuator 端点**: Spring Boot Actuator 健康检查
- **日志记录**: 使用 Logback 的结构化日志
- **API 文档**: Knife4j/Swagger API 文档(可配置)
- **审计日志**: 所有操作的完整审计追踪

## 技术栈

- **语言**: Java 8
- **框架**: Spring Boot 2.7.15
- **工作流引擎**: Activiti 7.0.x
- **ORM**: MyBatis Plus 3.5.3.1
- **数据库**: MySQL/MariaDB
- **缓存**: Redis (集群/哨兵/单机)
- **消息队列**: Kafka (主要), NSQ (可选)
- **API 文档**: Knife4j 3.0.2, Swagger 3.0.0
- **构建工具**: Maven
- **容器编排**: Kubernetes (Helm Charts)

## 前置要求

- Java 8+
- Maven 3.6+
- MySQL/MariaDB 5.7+
- Redis 5.0+
- Kafka 或 NSQ 消息队列
- Hydra OAuth2 服务

## 配置说明

服务使用 Spring Boot 属性文件进行配置:

### 主配置 (`application.properties`)

```properties
# 服务器配置
server.port=8081
server.additionalPorts=8083
server.servlet.context-path=/api/workflow-rest

# 数据库配置
server.dbtype=mysql
server.rds-host=localhost
server.rds-port=3306
server.rds-database=workflow
spring.datasource.username=root
spring.datasource.password=password
spring.datasource.maximumPoolSize=10

# Redis 配置
redis.connectType=sentinel
redis.sentinel.sentinelHost=proton-redis-proton-redis-sentinel.resource
redis.sentinel.sentinelPort=26379
redis.sentinel.masterGroupName=mymaster

# 消息队列 (Kafka/NSQ)
nsq.type=kafka
nsq.produce-host=kafka-broker.example.com
nsq.produce-port=9092

# 外部服务集成
oauth.admin-endpoint=http://oauth-admin:4445/admin
user.management-endpoint=http://user-management:8080

# API 文档
knife4j.enable=true
knife4j.production=false
```

### 环境特定配置

- `application-test.properties`: 测试环境配置
- `application-as_test.properties`: AnyShare 测试环境
- `application-ut.properties`: 单元测试配置

## API 文档

### 流程定义管理

- `POST /api/workflow-rest/v1/process-definitions` - 创建工作流定义
- `PUT /api/workflow-rest/v1/process-definitions/{id}` - 更新工作流定义
- `GET /api/workflow-rest/v1/process-definitions/{id}` - 获取工作流定义详情
- `DELETE /api/workflow-rest/v1/process-definitions/{id}` - 删除工作流定义
- `GET /api/workflow-rest/v1/process-definitions` - 列出工作流定义
- `POST /api/workflow-rest/v1/process-definitions/{id}/deploy` - 部署工作流定义
- `POST /api/workflow-rest/v1/process-definitions/{id}/copy` - 复制工作流定义

### 流程实例管理

- `POST /api/workflow-rest/v1/process-instances` - 启动流程实例
- `GET /api/workflow-rest/v1/process-instances/{id}` - 获取流程实例详情
- `DELETE /api/workflow-rest/v1/process-instances/{id}` - 取消流程实例
- `GET /api/workflow-rest/v1/process-instances` - 列出流程实例
- `GET /api/workflow-rest/v1/process-instances/{id}/trace` - 获取流程追踪日志

### 任务管理

- `POST /api/workflow-rest/v1/tasks/{id}/complete` - 完成任务
- `POST /api/workflow-rest/v1/tasks/{id}/assign` - 分配任务
- `GET /api/workflow-rest/v1/tasks` - 列出任务
- `GET /api/workflow-rest/v1/tasks/{id}` - 获取任务详情

### 审核策略管理

- `POST /api/workflow-rest/v1/audit-strategies` - 创建审核策略
- `PUT /api/workflow-rest/v1/audit-strategies/{id}` - 更新审核策略
- `DELETE /api/workflow-rest/v1/audit-strategies/{id}` - 删除审核策略
- `GET /api/workflow-rest/v1/audit-strategies` - 列出审核策略

### 部门审核人规则

- `POST /api/workflow-rest/v1/dept-auditor-rules` - 创建部门审核人规则
- `PUT /api/workflow-rest/v1/dept-auditor-rules/{id}` - 更新部门审核人规则
- `DELETE /api/workflow-rest/v1/dept-auditor-rules/{id}` - 删除部门审核人规则
- `GET /api/workflow-rest/v1/dept-auditor-rules` - 查询部门审核人规则

### 健康检查

- `GET /actuator/health` - 服务健康状态

## 构建和运行

### 使用 Maven 构建

```bash
# 构建所有模块
mvn clean package

# 构建特定模块
cd workflow-rest
mvn clean package

# 跳过测试
mvn clean package -DskipTests
```

### 本地运行

```bash
# 运行 workflow-rest 模块
cd workflow-rest
mvn spring-boot:run

# 或运行 JAR
java -jar target/workflow-rest-0.0.1-SNAPSHOT.jar
```

应用将在以下端口启动:
- 公共 API: `http://localhost:8081/api/workflow-rest`
- 私有 API: `http://localhost:8083/api/workflow-rest`

### Docker 构建

```bash
# 构建 Docker 镜像
docker build -t workflow-rest:latest -f workflow-rest/Dockerfile .
```

## 部署

### 使用 Helm

```bash
cd helm
helm install workflow . -f values.yaml
```

### 环境变量

可以覆盖配置的关键环境变量:

- `SERVER_PORT`: 服务器端口 (默认: 8081)
- `SERVER_RDS_HOST`: 数据库主机
- `SERVER_RDS_PORT`: 数据库端口
- `SERVER_RDS_DATABASE`: 数据库名称
- `SPRING_DATASOURCE_USERNAME`: 数据库用户名
- `SPRING_DATASOURCE_PASSWORD`: 数据库密码
- `REDIS_CONNECTTYPE`: Redis 连接类型 (cluster/sentinel/standalone)
- `NSQ_TYPE`: 消息队列类型 (kafka/nsq)

## 开发指南

### 代码结构

```
workflow-rest/
├── src/main/java/com/aishu/wf/api/
│   ├── config/          # 配置类
│   ├── rest/            # REST 控制器
│   ├── model/           # 请求/响应模型
│   ├── exception/       # 异常处理
│   ├── log/             # 审计日志处理器
│   └── listener/        # 事件监听器
└── src/main/resources/
    ├── application.properties
    └── logback.xml

workflow-code/
└── src/main/java/com/aishu/wf/core/
    ├── service/         # 业务服务
    ├── dao/             # 数据访问对象
    └── common/          # 通用工具
```

### 测试

```bash
# 运行所有测试
mvn test

# 运行测试并生成覆盖率报告
mvn test jacoco:report

# 运行特定测试类
mvn test -Dtest=ProcessDefinitionServiceTest
```

### API 文档

当 `knife4j.enable=true` 时,可访问 API 文档:
- Knife4j UI: `http://localhost:8081/api/workflow-rest/doc.html`
- Swagger UI: `http://localhost:8081/api/workflow-rest/swagger-ui.html`

### 代码风格

1. 遵循 Java 标准规范
2. 使用 Lombok 注解减少样板代码
3. 编写完整的 JavaDoc 注释
4. 使用 MyBatis Plus 进行数据库操作

## 数据库迁移

数据库迁移脚本位于 `migrations/` 目录。使用您首选的迁移工具(Flyway、Liquibase 等)应用迁移。

## 架构详解

### Activiti 工作流引擎集成

服务集成了 Activiti 7.0.x 工作流引擎以提供:
- BPMN 2.0 流程执行
- 任务生命周期管理
- 流程变量管理
- 自定义逻辑的事件监听器

### 审计日志

为以下操作实现了完整的审计日志:
- 流程定义操作(创建、更新、删除、复制)
- 流程实例操作(启动、取消)
- 任务操作(分配、完成)
- 审核策略操作(创建、更新、删除)
- 部门审核人规则操作
- 所有业务审核流程

### 多租户支持

服务通过以下方式支持多租户:
- 基于租户的数据隔离
- 与外部用户管理系统集成
- 基于部门的访问控制
