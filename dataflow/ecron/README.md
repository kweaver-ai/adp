# ECron

ECron is a distributed cron job scheduling and execution service that provides scheduled task management, immediate task execution, and task status monitoring capabilities. It consists of two core services: the Analysis Service for task scheduling and the Management Service for task lifecycle management.

[中文文档](README_zh.md)

## Core Architecture

ECron follows a microservices architecture with two main components:

```
┌─────────────────────────────────────────────────────────┐
│                   Analysis Service                      │
│        (Task Scheduling & Status Monitoring)            │
│  - Cron-based task scheduling                           │
│  - Immediate task execution                             │
│  - Task status tracking                                 │
│  - Message queue consumption                            │
└─────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────┐
│                  Management Service                     │
│           (Task CRUD & Execution Management)            │
│  - Task information management                          │
│  - Task execution coordination                          │
│  - RESTful API endpoints                                │
│  - OAuth2 authentication                                │
└─────────────────────────────────────────────────────────┘
```

### Directory Structure

- **`analysis/`**: Analysis service for task scheduling and monitoring
- **`management/`**: Management service for task CRUD operations and execution
- **`common/`**: Shared data structures and utilities
- **`utils/`**: Common utility functions (DB, HTTP, Auth, Logging, etc.)
- **`mock/`**: Mock implementations for testing
- **`migrations/`**: Database migration scripts
- **`helm/`**: Helm charts for Kubernetes deployment

## Key Features

### 1. Task Scheduling
- **Cron-based Scheduling**: Support for cron expressions to schedule recurring tasks
- **Immediate Execution**: Execute tasks immediately via message queue
- **Task Lifecycle Management**: Create, update, enable/disable, and delete tasks
- **Multi-node Support**: Optional multi-node deployment for high availability

### 2. Task Execution
- **HTTP/HTTPS Execution**: Execute tasks via HTTP requests (GET, POST, PUT, DELETE)
- **Command Execution**: Execute shell commands or scripts
- **Kubernetes Job Support**: Create and execute Kubernetes batch jobs
- **Retry Mechanism**: Configurable retry attempts for failed tasks

### 3. Task Monitoring
- **Status Tracking**: Real-time task execution status monitoring
- **Execution History**: Track task execution history and results
- **Webhook Notifications**: Send task execution results to configured webhooks
- **Health Checks**: Service health and readiness endpoints

### 4. Authentication & Authorization
- **OAuth2 Integration**: Token-based authentication via Hydra
- **Multi-tenant Support**: Tenant-based data isolation
- **Permission Control**: Fine-grained access control for task operations

### 5. Message Queue Integration
- **NSQ Support**: Message queue for task notifications and status updates
- **Kafka Support**: Alternative message queue connector
- **Asynchronous Processing**: Non-blocking task execution and status updates

## Tech Stack

- **Language**: Go 1.24.0
- **Web Framework**: Gin
- **Database**: MySQL/MariaDB
- **Message Queue**: NSQ (primary), Kafka (alternative)
- **Cron Library**: robfig/cron/v3
- **Testing**: GoConvey, gomock
- **Container Orchestration**: Kubernetes (Helm Charts)

## Prerequisites

- Go 1.24+
- MySQL/MariaDB 5.7+
- NSQ or Kafka message queue
- OAuth2 service (Hydra)

## Configuration

The service uses a YAML configuration file (`cronsvr.yaml`):

### Key Configuration Items

```yaml
# System settings
lang: zh_CN                    # System language
job_failures: 3                # Maximum task failure attempts
webhook: /api/cronsvr/v1/...   # Task result webhook path

# Service IDs
analysis_service_id: <uuid>    # Analysis service ID
management_service_id: <uuid>  # Management service ID

# Message Queue
mq_host: nsqlookupd.proton-mq.svc.cluster.local
mq_port: 4161
mq_connector_type: nsq         # nsq or kafka

# Cron Service
cron_addr: 0.0.0.0
cron_port: 12345
cron_protocol: http
multi_node: false              # Enable multi-node deployment

# Database
db_addr: localhost
db_port: 3306
db_name: ecron
user_name: root
user_pwd: password
max_open_conns: 30

# OAuth2 Service
oauth_public_addr: 127.0.0.1
oauth_public_port: 4444
oauth_admin_addr: 127.0.0.1
oauth_admin_port: 4445

# SSL Certificate (optional)
ssl_on: false
cert_file: /path/to/cert.crt
key_file: /path/to/cert.key
```

## API Documentation

### Management Service Endpoints

#### Task Management
- `GET /api/ecron-management/v1/jobtotal` - Get total number of tasks
- `GET /api/ecron-management/v1/job` - Query task information (with pagination)
- `POST /api/ecron-management/v1/job` - Create a new task
- `PUT /api/ecron-management/v1/job/:job_id` - Update task information
- `DELETE /api/ecron-management/v1/job/:job_id` - Delete a task

#### Task Status
- `GET /api/ecron-management/v1/jobstatus/:job_id` - Get task execution status
- `PUT /api/ecron-management/v1/jobstatus/:job_id` - Update task status

#### Task Control
- `PUT /api/ecron-management/v1/job/:job_id/enable` - Enable/disable a task
- `PUT /api/ecron-management/v1/job/:job_id/notify` - Update task notification settings

#### Task Execution
- `POST /api/ecron-management/v1/jobexecution` - Execute a task immediately

#### Health Checks
- `GET /health/ready` - Readiness probe
- `GET /health/alive` - Liveness probe

### Task Information Structure

```json
{
  "job_id": "unique-job-id",
  "job_name": "My Scheduled Task",
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
  "remarks": "Task description"
}
```

## Running Unit Tests

### Prerequisites for Testing

1. **Missing Library File**: If you encounter a missing `libtlq9client.so` file error, place the file in the `lib64` directory.

2. **Message Queue Dependency**: If `go get` fails for the Baolande message middleware dependency, clone the `go-msq` repository and run:
   ```bash
   ./go-msq/besmq/deploy_besmq_libs.sh
   ```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./analysis
go test ./management
go test ./utils
```

## Deployment

### Using Helm

```bash
cd helm
helm install ecron . -f values.yaml
```

### Environment Variables

The service can also be configured using environment variables, which override YAML configuration:

- `ECRON_DB_ADDR`: Database address
- `ECRON_DB_PORT`: Database port
- `ECRON_DB_NAME`: Database name
- `ECRON_MQ_HOST`: Message queue host
- `ECRON_CRON_PORT`: Cron service port

## Development Guide

1. **Code Style**: Follow Go standard conventions, use `golangci-lint` for linting
2. **Testing**: Write unit tests using `go.uber.org/mock` for mocking
3. **Mock Generation**: Use `go generate` to regenerate mocks:
   ```bash
   go generate ./...
   ```

## Architecture Details

### Analysis Service

The Analysis Service is responsible for:
- Loading and refreshing task lists from the database
- Scheduling cron-based tasks using the cron library
- Consuming immediate task execution messages from the message queue
- Monitoring and updating task execution status
- Sending task execution results to configured webhooks

### Management Service

The Management Service provides:
- RESTful API for task CRUD operations
- Task execution coordination
- OAuth2-based authentication and authorization
- Database persistence for task information and status
- Message queue publishing for task notifications

### Execution Module

The Execution Module supports:
- **HTTP Mode**: Execute HTTP requests with configurable methods, headers, and body
- **EXE Mode**: Execute shell commands or create Kubernetes jobs
- **Kubernetes Integration**: Create batch jobs in Kubernetes clusters
