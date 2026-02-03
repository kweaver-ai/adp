// Package discovery_task provides DiscoveryTask business logic.
package discovery_task

import (
	"context"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	mqclient "github.com/kweaver-ai/proton-mq-sdk-go"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/trace"

	"vega-backend/common"
	discoverytaskaccess "vega-backend/drivenadapters/discovery_task"
	"vega-backend/interfaces"
)

var (
	dtsOnce    sync.Once
	dtsService interfaces.DiscoveryTaskService
)

type discoveryTaskService struct {
	appSetting *common.AppSetting
	mqClient   mqclient.ProtonMQClient
	dta        interfaces.DiscoveryTaskAccess
}

// NewDiscoveryTaskService creates or returns the singleton DiscoveryTaskService.
func NewDiscoveryTaskService(appSetting *common.AppSetting) interfaces.DiscoveryTaskService {
	dtsOnce.Do(func() {
		client, err := mqclient.NewProtonMQClient(appSetting.MQSetting.MQHost, appSetting.MQSetting.MQPort,
			appSetting.MQSetting.MQHost, appSetting.MQSetting.MQPort, appSetting.MQSetting.MQType,
			mqclient.UserInfo(appSetting.MQSetting.Auth.Username, appSetting.MQSetting.Auth.Password),
			mqclient.AuthMechanism(appSetting.MQSetting.Auth.Mechanism),
		)
		if err != nil {
			logger.Fatal("failed to create a proton mq client:", err)
		}
		dtsService = &discoveryTaskService{
			appSetting: appSetting,
			mqClient:   client,
			dta:        discoverytaskaccess.NewDiscoveryTaskAccess(appSetting),
		}
	})
	return dtsService
}

// Create creates a new DiscoveryTask and sends message to Kafka.
func (s *discoveryTaskService) Create(ctx context.Context, catalogID string) (string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoveryTaskService.Create",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// Get account info from context
	accountInfo := interfaces.AccountInfo{}
	if ai, ok := ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo); ok {
		accountInfo = ai
	}

	now := time.Now().UnixMilli()
	task := &interfaces.DiscoveryTask{
		ID:          xid.New().String(),
		CatalogID:   catalogID,
		TriggerType: interfaces.DiscoveryTaskTriggerManual,
		Status:      interfaces.DiscoveryTaskStatusPending,
		Progress:    0,
		Message:     "",
		Creator:     accountInfo,
		CreateTime:  now,
	}

	// 1. Write to database
	if err := s.dta.Create(ctx, task); err != nil {
		logger.Errorf("Failed to create discovery task: %v", err)
		o11y.Error(ctx, "Failed to create discovery task")
		return "", err
	}

	// 2. TODO Send message to Kafka
	bytes, err := sonic.Marshal(&interfaces.DiscoveryTaskMessage{
		TaskID: task.ID,
	})
	if err != nil {
		logger.Errorf("Failed to marshal discovery task: %v", err)
		o11y.Error(ctx, "Failed to marshal discovery task")
		return "", err
	}

	err = s.mqClient.Pub(interfaces.DiscoveryTaskTopic, bytes)
	if err != nil {
		logger.Errorf("Failed to send message to Kafka: %v", err)
		o11y.Error(ctx, "Failed to send message to Kafka")
		return "", err
	}

	return task.ID, nil
}

// GetByID retrieves a DiscoveryTask by ID.
func (s *discoveryTaskService) GetByID(ctx context.Context, id string) (*interfaces.DiscoveryTask, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoveryTaskService.GetByID",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return s.dta.GetByID(ctx, id)
}

// List lists DiscoveryTasks for a catalog.
func (s *discoveryTaskService) List(ctx context.Context, params interfaces.DiscoveryTaskQueryParams) ([]*interfaces.DiscoveryTask, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoveryTaskService.List",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return s.dta.List(ctx, params)
}

// UpdateStatus updates a DiscoveryTask's status.
func (s *discoveryTaskService) UpdateStatus(ctx context.Context, id, status, message string, stime int64) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoveryTaskService.UpdateStatus",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return s.dta.UpdateStatus(ctx, id, status, message, stime)
}

// UpdateResult updates a DiscoveryTask's result.
func (s *discoveryTaskService) UpdateResult(ctx context.Context, id string, result *interfaces.DiscoveryResult, stime int64) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DiscoveryTaskService.UpdateResult",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return s.dta.UpdateResult(ctx, id, result, stime)
}
