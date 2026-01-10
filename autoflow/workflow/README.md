# Workflow Service

Workflow Service is a comprehensive workflow and audit management system built on Spring Boot and Activiti. It provides workflow definition, process instance management, audit process orchestration, and department-based auditor rule management capabilities.

[中文文档](README_zh.md)

## Core Architecture

The Workflow Service follows a multi-module Maven architecture:

```
┌─────────────────────────────────────────────────────────┐
│                    workflow-rest                        │
│              (REST API & Web Layer)                     │
│  - RESTful API endpoints                                │
│  - Request/Response handling                            │
│  - Authentication & Authorization                       │
│  - Exception handling                                   │
└─────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────┐
│                   workflow-code                         │
│            (Core Business Logic)                        │
│  - Workflow engine integration (Activiti)               │
│  - Process definition management                        │
│  - Task assignment & execution                          │
│  - Audit rule management                                │
└─────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────┐
│                  doc-audit-rest                         │
│            (Audit Management Module)                    │
│  - Audit process orchestration                          │
│  - Audit strategy management                            │
│  - Auditor assignment                                   │
└─────────────────────────────────────────────────────────┘
```

### Module Structure

- **`workflow-rest/`**: REST API layer with controllers, configuration, and exception handling
- **`workflow-code/`**: Core business logic and Activiti workflow engine integration
- **`doc-audit-rest/`**: General audit process management module
- **`helm/`**: Kubernetes deployment Helm charts
- **`migrations/`**: Database migration scripts

## Key Features

### 1. Workflow Management
- **Process Definition**: Create, update, deploy, and delete workflow definitions
- **BPMN 2.0 Support**: Full support for BPMN 2.0 workflow modeling
- **Process Deployment**: Deploy workflow definitions to Activiti engine
- **Version Control**: Manage multiple versions of workflow definitions
- **Process Copying**: Clone existing workflow definitions

### 2. Process Instance Management
- **Instance Creation**: Start new workflow instances
- **Instance Tracking**: Monitor process instance execution status
- **Task Assignment**: Assign tasks to users or groups
- **Instance Cancellation**: Cancel running process instances
- **Execution History**: Track complete process execution history

### 3. Audit Management
- **Audit Strategies**: Configure audit strategies for various business scenarios
- **Auditor Assignment**: Assign auditors based on departments or roles
- **Audit Rules**: Define department-based auditor rules
- **Free Audit Departments**: Configure departments exempt from audit
- **Third-party Audit**: Integration with external audit systems

### 4. Task Management
- **Task Assignment**: Assign tasks to specific users
- **Countersign Support**: Multi-user approval (countersign) tasks
- **Task Completion**: Complete tasks with approval or rejection
- **Task Delegation**: Delegate tasks to other users
- **Task Filtering**: Filter tasks by initiator, assignee, or status

### 5. Integration Capabilities
- **OAuth2 Authentication**: Token-based authentication via OAuth2 service
- **User Management**: Integration with external user management systems
- **Message Queue**: Kafka/NSQ support for asynchronous processing
- **Redis Caching**: Distributed caching for performance optimization

### 6. Observability
- **Actuator Endpoints**: Spring Boot Actuator health checks
- **Logging**: Structured logging with Logback
- **API Documentation**: Knife4j/Swagger API documentation (configurable)
- **Audit Logging**: Comprehensive audit trail for all operations

## Tech Stack

- **Language**: Java 8
- **Framework**: Spring Boot 2.7.15
- **Workflow Engine**: Activiti 7.0.x
- **ORM**: MyBatis Plus 3.5.3.1
- **Database**: MySQL/MariaDB
- **Cache**: Redis (Cluster/Sentinel/Standalone)
- **Message Queue**: Kafka (primary), NSQ (alternative)
- **API Documentation**: Knife4j 3.0.2, Swagger 3.0.0
- **Build Tool**: Maven
- **Container Orchestration**: Kubernetes (Helm Charts)

## Prerequisites

- Java 8+
- Maven 3.6+
- MySQL/MariaDB 5.7+
- Redis 5.0+
- Kafka or NSQ message queue
- Hydra OAuth2 service

## Configuration

The service uses Spring Boot properties files for configuration:

### Main Configuration (`application.properties`)

```properties
# Server Configuration
server.port=8081
server.additionalPorts=8083
server.servlet.context-path=/api/workflow-rest

# Database Configuration
server.dbtype=mysql
server.rds-host=localhost
server.rds-port=3306
server.rds-database=workflow
spring.datasource.username=root
spring.datasource.password=password
spring.datasource.maximumPoolSize=10

# Redis Configuration
redis.connectType=sentinel
redis.sentinel.sentinelHost=proton-redis-proton-redis-sentinel.resource
redis.sentinel.sentinelPort=26379
redis.sentinel.masterGroupName=mymaster

# Message Queue (Kafka/NSQ)
nsq.type=kafka
nsq.produce-host=kafka-broker.example.com
nsq.produce-port=9092

# External Service Integration
oauth.admin-endpoint=http://oauth-admin:4445/admin
user.management-endpoint=http://user-management:8080

# API Documentation
knife4j.enable=true
knife4j.production=false
```

### Environment-Specific Profiles

- `application-test.properties`: Test environment configuration
- `application-as_test.properties`: AnyShare test environment
- `application-ut.properties`: Unit test configuration

## API Documentation

### Process Definition Management

- `POST /api/workflow-rest/v1/process-definitions` - Create workflow definition
- `PUT /api/workflow-rest/v1/process-definitions/{id}` - Update workflow definition
- `GET /api/workflow-rest/v1/process-definitions/{id}` - Get workflow definition details
- `DELETE /api/workflow-rest/v1/process-definitions/{id}` - Delete workflow definition
- `GET /api/workflow-rest/v1/process-definitions` - List workflow definitions
- `POST /api/workflow-rest/v1/process-definitions/{id}/deploy` - Deploy workflow definition
- `POST /api/workflow-rest/v1/process-definitions/{id}/copy` - Copy workflow definition

### Process Instance Management

- `POST /api/workflow-rest/v1/process-instances` - Start process instance
- `GET /api/workflow-rest/v1/process-instances/{id}` - Get process instance details
- `DELETE /api/workflow-rest/v1/process-instances/{id}` - Cancel process instance
- `GET /api/workflow-rest/v1/process-instances` - List process instances
- `GET /api/workflow-rest/v1/process-instances/{id}/trace` - Get process trace log

### Task Management

- `POST /api/workflow-rest/v1/tasks/{id}/complete` - Complete task
- `POST /api/workflow-rest/v1/tasks/{id}/assign` - Assign task
- `GET /api/workflow-rest/v1/tasks` - List tasks
- `GET /api/workflow-rest/v1/tasks/{id}` - Get task details

### Audit Strategy Management

- `POST /api/workflow-rest/v1/audit-strategies` - Create audit strategy
- `PUT /api/workflow-rest/v1/audit-strategies/{id}` - Update audit strategy
- `DELETE /api/workflow-rest/v1/audit-strategies/{id}` - Delete audit strategy
- `GET /api/workflow-rest/v1/audit-strategies` - List audit strategies

### Department Auditor Rules

- `POST /api/workflow-rest/v1/dept-auditor-rules` - Create department auditor rule
- `PUT /api/workflow-rest/v1/dept-auditor-rules/{id}` - Update department auditor rule
- `DELETE /api/workflow-rest/v1/dept-auditor-rules/{id}` - Delete department auditor rule
- `GET /api/workflow-rest/v1/dept-auditor-rules` - Query department auditor rules

### Health Checks

- `GET /actuator/health` - Service health status

## Building and Running

### Build with Maven

```bash
# Build all modules
mvn clean package

# Build specific module
cd workflow-rest
mvn clean package

# Skip tests
mvn clean package -DskipTests
```

### Run Locally

```bash
# Run workflow-rest module
cd workflow-rest
mvn spring-boot:run

# Or run the JAR
java -jar target/workflow-rest-0.0.1-SNAPSHOT.jar
```

The application will start on:
- Public API: `http://localhost:8081/api/workflow-rest`
- Private API: `http://localhost:8083/api/workflow-rest`

### Docker Build

```bash
# Build Docker image
docker build -t workflow-rest:latest -f workflow-rest/Dockerfile .
```

## Deployment

### Using Helm

```bash
cd helm
helm install workflow . -f values.yaml
```

### Environment Variables

Key environment variables that can override configuration:

- `SERVER_PORT`: Server port (default: 8081)
- `SERVER_RDS_HOST`: Database host
- `SERVER_RDS_PORT`: Database port
- `SERVER_RDS_DATABASE`: Database name
- `SPRING_DATASOURCE_USERNAME`: Database username
- `SPRING_DATASOURCE_PASSWORD`: Database password
- `REDIS_CONNECTTYPE`: Redis connection type (cluster/sentinel/standalone)
- `NSQ_TYPE`: Message queue type (kafka/nsq)

## Development Guide

### Code Structure

```
workflow-rest/
├── src/main/java/com/aishu/wf/api/
│   ├── config/          # Configuration classes
│   ├── rest/            # REST controllers
│   ├── model/           # Request/Response models
│   ├── exception/       # Exception handling
│   ├── log/             # Audit log handlers
│   └── listener/        # Event listeners
└── src/main/resources/
    ├── application.properties
    └── logback.xml

workflow-code/
└── src/main/java/com/aishu/wf/core/
    ├── service/         # Business services
    ├── dao/             # Data access objects
    └── common/          # Common utilities
```

### Testing

```bash
# Run all tests
mvn test

# Run tests with coverage
mvn test jacoco:report

# Run specific test class
mvn test -Dtest=ProcessDefinitionServiceTest
```

### API Documentation

When `knife4j.enable=true`, access API documentation at:
- Knife4j UI: `http://localhost:8081/api/workflow-rest/doc.html`
- Swagger UI: `http://localhost:8081/api/workflow-rest/swagger-ui.html`

### Code Style

1. Follow Java standard conventions
2. Use Lombok annotations to reduce boilerplate
3. Write comprehensive JavaDoc comments
4. Use MyBatis Plus for database operations

## Database Migrations

Database migration scripts are located in the `migrations/` directory. Apply migrations using your preferred migration tool (Flyway, Liquibase, etc.).

## Architecture Details

### Activiti Workflow Engine Integration

The service integrates with Activiti 7.0.x workflow engine to provide:
- BPMN 2.0 process execution
- Task lifecycle management
- Process variable management
- Event listeners for custom logic

### Audit Logging

Comprehensive audit logging is implemented for:
- Process definition operations (create, update, delete, copy)
- Process instance operations (start, cancel)
- Task operations (assign, complete)
- Audit strategy operations (create, update, delete)
- Department auditor rule operations
- All business audit processes

### Multi-tenancy Support

The service supports multi-tenancy through:
- Tenant-based data isolation
- Integration with external user management systems
- Department-based access control
