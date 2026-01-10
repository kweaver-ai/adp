# Autoflow

Autoflow is a comprehensive automation platform that provides two core capabilities: **Data Processing Flow** and **Workflow Management**. It enables users to build, orchestrate, and execute automated data pipelines and business process workflows through a unified platform.

[中文文档](README_zh.md)

## Overview

Autoflow combines the power of data processing automation with flexible workflow orchestration, providing a complete solution for enterprise automation needs. Whether you need to process large volumes of data, automate business processes, or integrate complex systems, Autoflow provides the tools and services to accomplish your goals.

## Core Capabilities

### 1. Data Processing Flow

Build and execute automated data pipelines with visual workflow design, code execution, and data transformation capabilities.

**Key Features:**
- Visual workflow designer for data pipelines
- Sandboxed Python code execution
- Data transformation and analysis
- Document processing (Word, Excel, PDF)
- OCR and text extraction
- Scheduled and event-driven execution
- Real-time data streaming

**Use Cases:**
- ETL (Extract, Transform, Load) pipelines
- Data quality validation and cleansing
- Automated report generation
- Document processing and analysis
- Image and text recognition workflows

### 2. Workflow Management

Orchestrate business processes with BPMN 2.0 workflows, task management, and audit capabilities.

**Key Features:**
- BPMN 2.0 workflow modeling
- Process instance management
- Task assignment and approval
- Audit process orchestration
- Department-based rules
- Multi-tenant support
- Integration with external systems

**Use Cases:**
- Business process automation
- Approval workflows
- Audit and compliance processes
- Document review and approval
- Multi-step business operations

## Architecture

Autoflow is built as a microservices architecture with the following components:

```
┌─────────────────────────────────────────────────────────┐
│                  Frontend Layer                         │
│  - dia-flow-web: Data flow visual designer              │
│  - workflow-manage-front: Workflow management UI        │
│  - doc-audit-client: Audit client interface             │
│  - workflow-manage-client: Workflow client              │
└─────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────┐
│               Data Processing Services                  │
│  - flow-automation: Data flow orchestration             │
│  - coderunner: Sandboxed code execution                 │
│  - flow-stream-data-pipeline: Real-time streaming       │
└─────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────┐
│              Workflow Management Services               │
│  - workflow: BPMN workflow engine (Activiti)            │
│  - ecron: Scheduled task management                     │
│  - workflow-config: Workflow configuration              │
└─────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────┐
│                 Shared Libraries                        │
│  - ide-go-lib: Common Go libraries                      │
└─────────────────────────────────────────────────────────┘
```

## Services Overview

### Data Processing Flow Services

#### [flow-automation](flow-automation/)
Core data flow orchestration service that manages the complete lifecycle of data pipeline executions.
- **Language**: Go
- **Framework**: Gin
- **Key Features**: DAG management, executor management, trigger system, data connections

#### [coderunner](coderunner/)
Sandboxed Python code execution service for running custom data processing logic.
- **Language**: Python 3.9
- **Key Features**: RestrictedPython execution, package management, document processing, OCR

#### [flow-stream-data-pipeline](flow-stream-data-pipeline/)
Real-time data streaming pipeline service.
- **Key Features**: Stream processing, real-time data transformation

### Workflow Management Services

#### [workflow](workflow/)
BPMN 2.0 workflow engine for business process automation.
- **Language**: Java 8
- **Framework**: Spring Boot + Activiti
- **Key Features**: Process definition, instance management, task assignment, audit management

#### [ecron](ecron/)
Distributed cron job scheduling and execution service.
- **Language**: Go
- **Key Features**: Cron-based scheduling, immediate execution, task monitoring, multi-node support

#### [workflow-config](workflow-config/)
Workflow configuration management service.

### Frontend Applications

#### [dia-flow-web](dia-flow-web/)
Visual designer for building data processing flows.
- **Technology**: Modern web framework
- **Features**: Drag-and-drop workflow design, node configuration, execution monitoring

#### [workflow-manage-front](workflow-manage-front/)
Workflow management user interface.
- **Features**: Process modeling, instance monitoring, task management

#### [doc-audit-client](doc-audit-client/)
Document audit client interface.

#### [workflow-manage-client](workflow-manage-client/)
Workflow management client application.

### Shared Libraries

#### [ide-go-lib](ide-go-lib/)
Common Go libraries shared across Go-based services.

## Tech Stack

### Backend Services
- **Go**: flow-automation, ecron, ide-go-lib
- **Python**: coderunner, flow-stream-data-pipeline
- **Java**: workflow

### Frameworks & Libraries
- **Go**: Gin, MongoDB, Redis, Kafka
- **Python**: Tornado, RestrictedPython, pandas, SQLAlchemy
- **Java**: Spring Boot, Activiti, MyBatis Plus

### Infrastructure
- **Databases**: MongoDB, MySQL/MariaDB, Redis
- **Message Queues**: Kafka, NSQ
- **Container Orchestration**: Kubernetes (Helm)
- **Authentication**: OAuth2

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Kubernetes cluster (for production deployment)
- MongoDB, MySQL/MariaDB, Redis
- Kafka or NSQ message queue

### Quick Start with Docker Compose

```bash
# Clone the repository
git clone <repository-url>
cd autoflow

# Start all services
docker-compose up -d

# Access the applications
# Data Flow Designer: http://localhost:3000
# Workflow Manager: http://localhost:3001
```

### Individual Service Setup

Each service can be run independently. Refer to the README in each service directory for specific setup instructions:

- [flow-automation/README.md](flow-automation/README.md)
- [workflow/README.md](workflow/README.md)
- [coderunner/README.md](coderunner/README.md)
- [ecron/README.md](ecron/README.md)

## Deployment

### Kubernetes Deployment

Each service includes Helm charts for Kubernetes deployment:

```bash
# Deploy flow-automation
cd flow-automation/helm
helm install flow-automation . -f values.yaml

# Deploy workflow service
cd workflow/helm
helm install workflow . -f values.yaml

# Deploy coderunner
cd coderunner/helm
helm install coderunner . -f values.yaml

# Deploy ecron
cd ecron/helm
helm install ecron . -f values.yaml
```

### Configuration

Each service uses environment variables or configuration files. Key configuration areas:

- **Database connections**: MongoDB, MySQL, Redis
- **Message queue**: Kafka/NSQ endpoints
- **Authentication**: OAuth2 service endpoints
- **Service discovery**: Internal service URLs

## Integration

### Data Flow + Workflow Integration

Autoflow services are designed to work together:

1. **Trigger Workflows from Data Flows**: Use flow-automation to trigger workflow processes based on data events
2. **Execute Code in Workflows**: Call coderunner from workflow tasks for custom logic
3. **Schedule Workflows**: Use ecron to schedule periodic workflow executions
4. **Audit Data Operations**: Use workflow audit capabilities to track data processing operations

### External System Integration

- **OAuth2 Authentication**: Integrate with external identity providers
- **Message Queues**: Connect to Kafka/NSQ for event-driven architectures
- **REST APIs**: All services expose RESTful APIs for integration
- **Webhooks**: Configure webhooks for event notifications

## Development

### Project Structure

```
autoflow/
├── flow-automation/       # Data flow orchestration (Go)
├── coderunner/           # Code execution service (Python)
├── workflow/             # Workflow engine (Java)
├── ecron/                # Scheduled tasks (Go)
├── flow-stream-data-pipeline/  # Streaming pipeline (Python)
├── dia-flow-web/         # Data flow UI
├── workflow-manage-front/  # Workflow UI
├── doc-audit-client/     # Audit client
├── workflow-manage-client/  # Workflow client
├── workflow-config/      # Workflow config
└── ide-go-lib/          # Shared Go libraries
```

### Contributing

1. Choose the service you want to contribute to
2. Follow the development guide in the service's README
3. Write tests for your changes
4. Submit a pull request

### Code Style

- **Go**: Follow Go standard conventions, use `golangci-lint`
- **Python**: Follow PEP 8, use `black` and `pylint`
- **Java**: Follow Java conventions, use Spring Boot best practices

## Documentation

- [Flow Automation Documentation](flow-automation/README.md)
- [Workflow Service Documentation](workflow/README.md)
- [CodeRunner Documentation](coderunner/README.md)
- [ECron Documentation](ecron/README.md)

## Use Case Examples

### Example 1: Automated Data Processing Pipeline

1. Design a data flow in dia-flow-web
2. Configure data source connections
3. Add transformation nodes with custom Python code
4. Set up scheduled execution via ecron
5. Monitor execution in flow-automation dashboard

### Example 2: Document Approval Workflow

1. Model approval process in workflow-manage-front
2. Define approval rules and auditors
3. Deploy workflow definition
4. Trigger workflow instances via API
5. Track approval progress and history

### Example 3: Hybrid Automation

1. Create data processing flow to extract and analyze data
2. Trigger workflow approval based on analysis results
3. Execute additional data operations after approval
4. Generate reports and notifications

## Monitoring and Observability

- **Health Checks**: All services expose health endpoints
- **Metrics**: Prometheus metrics for monitoring
- **Logging**: Structured logging across all services
- **Tracing**: Distributed tracing support

## Security

- **Authentication**: OAuth2-based authentication
- **Authorization**: Role-based access control
- **Code Execution**: Sandboxed execution environment
- **Data Isolation**: Multi-tenant data isolation
- **Audit Trails**: Comprehensive audit logging

## Support

For questions and support:
- Check service-specific README files
- Review API documentation
- Contact the development team
