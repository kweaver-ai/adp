# Flow Automation

[中文](README.zh.md) | English

Flow Automation is a core service responsible for orchestrating and executing automated workflows. It manages the complete lifecycle of workflow executions, integrates with various infrastructure components, and provides comprehensive APIs for workflow management.

## Core Architecture

This project follows a Hexagonal Architecture (Ports and Adapters), separating core business logic from external interfaces:

```
┌─────────────────────────────────────────────────────────┐
│                    Driver Adapters                      │
│          (HTTP API, Message Queue Consumers)            │
│  ┌───────────────────────────────────────────────────┐  │
│  │              Core Business Logic                  │  │
│  │         (logics/, module/, pkg/)                  │  │
│  └───────────────────────────────────────────────────┘  │
│                   Driven Adapters                       │
│        (Database, External Services, Cache)             │
└─────────────────────────────────────────────────────────┘
```

### Directory Structure

- **`logics/`**: Core business logic layer, containing workflow management, executors, triggers, etc.
- **`driveradapters/`**: Driver adapters (inbound), handling HTTP requests, message queue consumption, etc.
- **`drivenadapters/`**: Driven adapters (outbound), interacting with databases, external services, cache, etc.
- **`module/`**: Domain modules, including initialization and configuration
- **`pkg/`**: Shared utility packages and entity definitions

## Key Features

### 1. Workflow Management (DAG Management)
- **Create & Edit**: Support for creating, updating, and deleting workflows (DAGs)
- **Version Control**: Workflow version management
- **Execution Management**: Start, cancel, and pause workflow instances
- **Debug Mode**: Support for single-step and full debugging
- **Batch Operations**: Batch query and operations on workflows

### 2. Custom Executors
- **Executor Management**: Create, update, and delete custom executors
- **Action Management**: Add, modify, and delete actions for executors
- **Access Control**: User permission-based executor access control
- **Agent Import**: Support for importing agents from app store

### 3. Trigger System
- **Scheduled Triggers**: Cron-based scheduled task triggers
- **Event Triggers**: Support for external event-driven workflow execution
- **Manual Triggers**: Support for user-initiated workflow execution
- **Callback Handling**: Process asynchronous task callbacks

### 4. Authentication & Authorization
- **Token Authentication**: OAuth2 authentication via Hydra
- **Permission Verification**: Fine-grained resource access control
- **Business Domain Isolation**: Multi-tenant business domain isolation

### 5. Security Policies
- **Access Control**: API access security policies
- **Data Isolation**: Tenant data isolation
- **Audit Logging**: Operation audit trails

### 6. Observability
- **Metrics Monitoring**: Prometheus metrics exposure
- **Distributed Tracing**: OpenTelemetry distributed tracing
- **Log Management**: Structured logging and streaming logs
- **Health Checks**: Service health status monitoring

### 7. Data Flow Management
- **Data Connections**: Manage data source connection configurations
- **Data Transformation**: Support for data flow transformation and processing

### 8. Operator Management
- **Operator Registration**: Register and manage workflow operators
- **Operator Execution**: Execute various data processing operators

## Tech Stack

- **Language**: Go 1.24.0
- **Web Framework**: Gin
- **Databases**: 
  - MongoDB (workflow definitions and instance storage)
  - MariaDB/MySQL (relational data)
  - Redis (caching and distributed locks)
- **Message Queue**: Kafka
- **Configuration**: Viper, godotenv
- **Observability**: OpenTelemetry, Prometheus
- **Container Orchestration**: Kubernetes (Helm Charts)

## Prerequisites

- Go 1.24+
- Docker & Docker Compose (for local infrastructure)

## Getting Started

For detailed local development setup instructions, including running infrastructure and mock services, please refer to [LOCAL_DEVELOPMENT.md](LOCAL_DEVELOPMENT.md).

### Quick Run

1. **Start Dependencies**:
   ```bash
   docker-compose up -d
   ```

2. **Run Application**:
   ```bash
   go run main.go
   ```

The application will start two services:
- Public API Server: `http://localhost:8082`
- Private API Server: `http://localhost:8083`

## Configuration

The application uses environment variables for configuration:

- **`.env`**: Default configuration (committed to Git)
- **`.env.local`**: Local development overrides (gitignored)

Key configuration items:
- `API_SERVER_PORT`: Public API port (default 8082)
- `API_SERVER_PRIVATE_PORT`: Private API port (default 8083)
- `MONGODB_HOST`: MongoDB address
- `REDIS_HOST`: Redis address
- `KAFKA_BROKERS`: Kafka cluster addresses

## Project Structure

```text
.
├── common/             # Common utilities and constants
├── conf/               # Configuration files
├── docs/               # Documentation
├── drivenadapters/     # Outbound adapters (DB, external APIs)
├── driveradapters/     # Inbound adapters (HTTP, Event Consumers)
│   ├── admin/          # Admin interfaces
│   ├── alarm/          # Alarm interfaces
│   ├── auth/           # Authentication interfaces
│   ├── executor/       # Executor interfaces
│   ├── mgnt/           # Workflow management interfaces
│   ├── trigger/        # Trigger interfaces
│   └── ...
├── helm/               # Helm deployment charts
├── logics/             # Business logic
│   ├── mgnt/           # Workflow management logic
│   ├── executor/       # Executor logic
│   ├── cronjob/        # Scheduled task logic
│   └── ...
├── mock-server/        # Mock services for local development
├── module/             # Domain modules
├── pkg/                # Shared packages
│   ├── actions/        # Action definitions
│   ├── entity/         # Entity definitions
│   └── ...
├── resource/           # Static resources
├── schema/             # Database schemas
├── scripts/            # Helper scripts
├── store/              # Data storage interfaces
└── utils/              # General utilities
```

## API Documentation

Main API endpoints:

### Workflow Management
- `POST /api/automation/v1/dags` - Create workflow
- `PUT /api/automation/v1/dags/:id` - Update workflow
- `GET /api/automation/v1/dags/:id` - Get workflow details
- `DELETE /api/automation/v1/dags/:id` - Delete workflow
- `GET /api/automation/v1/dags` - List workflows

### Workflow Execution
- `POST /api/automation/v1/dags/:id/run` - Execute workflow
- `POST /api/automation/v1/dags/:id/cancel` - Cancel execution
- `GET /api/automation/v1/dags/:id/instances` - Query execution instances

### Executor Management
- `POST /api/automation/v1/executors` - Create executor
- `PUT /api/automation/v1/executors/:id` - Update executor
- `GET /api/automation/v1/executors` - List executors

## Deployment

The project uses Helm for Kubernetes deployment:

```bash
cd helm/flow-automation
helm install flow-automation . -f values.yaml
```

## Development Guide

1. **Code Style**: Follow Go standard code conventions, use `golangci-lint` for code checking
2. **Testing**: Use `go test` to run unit tests
3. **Mock Generation**: Use `go.uber.org/mock` to generate test mocks