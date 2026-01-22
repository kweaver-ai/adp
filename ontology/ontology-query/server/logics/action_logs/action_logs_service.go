package action_logs

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"ontology-query/common"
	"ontology-query/interfaces"
	"ontology-query/logics"
)

var (
	alsOnce    sync.Once
	alsService interfaces.ActionLogsService
)

type actionLogsService struct {
	appSetting *common.AppSetting
	osAccess   interfaces.OpenSearchAccess
}

// NewActionLogsService creates a singleton instance of ActionLogsService
func NewActionLogsService(appSetting *common.AppSetting) interfaces.ActionLogsService {
	alsOnce.Do(func() {
		alsService = &actionLogsService{
			appSetting: appSetting,
			osAccess:   logics.OSA,
		}
	})
	return alsService
}

// CreateExecution creates a new execution record in OpenSearch
func (s *actionLogsService) CreateExecution(ctx context.Context, exec *interfaces.ActionExecution) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "CreateExecution", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.SetAttributes(
		attr.Key("execution_id").String(exec.ID),
		attr.Key("kn_id").String(exec.KNID),
		attr.Key("action_type_id").String(exec.ActionTypeID),
	)

	indexName := interfaces.GetActionExecutionIndex(exec.KNID)

	// Ensure index exists
	if err := s.ensureIndexExists(ctx, indexName); err != nil {
		logger.Errorf("Failed to ensure index exists: %v", err)
		return fmt.Errorf("failed to ensure index exists: %w", err)
	}

	// Insert the execution record
	if err := s.osAccess.InsertData(ctx, indexName, exec.ID, exec); err != nil {
		logger.Errorf("Failed to insert execution record: %v", err)
		return fmt.Errorf("failed to insert execution record: %w", err)
	}

	logger.Debugf("Created execution record: %s", exec.ID)
	return nil
}

// UpdateExecution updates an existing execution record
func (s *actionLogsService) UpdateExecution(ctx context.Context, knID, execID string, updates map[string]any) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "UpdateExecution", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.SetAttributes(
		attr.Key("execution_id").String(execID),
		attr.Key("kn_id").String(knID),
	)

	// Get the current execution
	exec, err := s.GetExecution(ctx, knID, execID)
	if err != nil {
		return err
	}

	// Apply updates
	execMap := structToMap(exec)
	for k, v := range updates {
		execMap[k] = v
	}

	indexName := interfaces.GetActionExecutionIndex(knID)

	// Re-insert with updated values (OpenSearch index API is upsert)
	if err := s.osAccess.InsertData(ctx, indexName, execID, execMap); err != nil {
		logger.Errorf("Failed to update execution record: %v", err)
		return fmt.Errorf("failed to update execution record: %w", err)
	}

	logger.Debugf("Updated execution record: %s", execID)
	return nil
}

// GetExecution retrieves a single execution by ID
func (s *actionLogsService) GetExecution(ctx context.Context, knID, execID string) (*interfaces.ActionExecution, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetExecution", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.SetAttributes(
		attr.Key("execution_id").String(execID),
		attr.Key("kn_id").String(knID),
	)

	indexName := interfaces.GetActionExecutionIndex(knID)

	// Build query to get by ID
	query := map[string]any{
		"query": map[string]any{
			"term": map[string]any{
				"id": execID,
			},
		},
		"size": 1,
	}

	hits, err := s.osAccess.SearchData(ctx, indexName, query)
	if err != nil {
		logger.Errorf("Failed to search execution: %v", err)
		return nil, fmt.Errorf("failed to search execution: %w", err)
	}

	if len(hits) == 0 {
		return nil, fmt.Errorf("execution not found: %s", execID)
	}

	exec, err := mapToActionExecution(hits[0].Source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse execution: %w", err)
	}

	return exec, nil
}

// QueryExecutions queries executions based on filter criteria
func (s *actionLogsService) QueryExecutions(ctx context.Context, query *interfaces.ActionLogQuery) (*interfaces.ActionExecutionList, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "QueryExecutions", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.SetAttributes(attr.Key("kn_id").String(query.KNID))

	indexName := interfaces.GetActionExecutionIndex(query.KNID)

	// Build the must conditions
	mustConditions := []map[string]any{}

	if query.ActionTypeID != "" {
		mustConditions = append(mustConditions, map[string]any{
			"term": map[string]any{
				"action_type_id": query.ActionTypeID,
			},
		})
	}

	if query.Status != "" {
		mustConditions = append(mustConditions, map[string]any{
			"term": map[string]any{
				"status": query.Status,
			},
		})
	}

	if query.TriggerType != "" {
		mustConditions = append(mustConditions, map[string]any{
			"term": map[string]any{
				"trigger_type": query.TriggerType,
			},
		})
	}

	if len(query.StartTimeRange) == 2 {
		mustConditions = append(mustConditions, map[string]any{
			"range": map[string]any{
				"start_time": map[string]any{
					"gte": query.StartTimeRange[0],
					"lte": query.StartTimeRange[1],
				},
			},
		})
	}

	// Build the query
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 1000 {
		limit = 1000
	}

	osQuery := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": mustConditions,
			},
		},
		"size": limit,
		"sort": []map[string]any{
			{"start_time": map[string]any{"order": "desc"}},
			{"id": map[string]any{"order": "asc"}},
		},
	}

	if len(query.SearchAfter) > 0 {
		osQuery["search_after"] = query.SearchAfter
	}

	hits, err := s.osAccess.SearchData(ctx, indexName, osQuery)
	if err != nil {
		logger.Errorf("Failed to query executions: %v", err)
		return nil, fmt.Errorf("failed to query executions: %w", err)
	}

	// Convert hits to executions
	executions := make([]interfaces.ActionExecution, 0, len(hits))
	var lastSort []any

	for _, hit := range hits {
		exec, err := mapToActionExecution(hit.Source)
		if err != nil {
			logger.Warnf("Failed to parse execution, skipping: %v", err)
			continue
		}
		executions = append(executions, *exec)
		lastSort = hit.Sort
	}

	result := &interfaces.ActionExecutionList{
		Entries:     executions,
		SearchAfter: lastSort,
	}

	// Get total count if needed
	if query.NeedTotal {
		countQuery := map[string]any{
			"query": map[string]any{
				"bool": map[string]any{
					"must": mustConditions,
				},
			},
		}
		countBytes, err := s.osAccess.Count(ctx, indexName, countQuery)
		if err == nil {
			var countResult struct {
				Count int `json:"count"`
			}
			if json.Unmarshal(countBytes, &countResult) == nil {
				result.TotalCount = countResult.Count
			}
		}
	}

	return result, nil
}

// ensureIndexExists creates the index if it doesn't exist
func (s *actionLogsService) ensureIndexExists(ctx context.Context, indexName string) error {
	exists, err := s.osAccess.IndexExists(ctx, indexName)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	// Create the index with mappings
	indexBody := map[string]any{
		"settings": map[string]any{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]any{
			"properties": map[string]any{
				"id":                 map[string]any{"type": "keyword"},
				"kn_id":              map[string]any{"type": "keyword"},
				"action_type_id":     map[string]any{"type": "keyword"},
				"action_type_name":   map[string]any{"type": "keyword"},
				"action_source_type": map[string]any{"type": "keyword"},
				"object_type_id":     map[string]any{"type": "keyword"},
				"trigger_type":       map[string]any{"type": "keyword"},
				"status":             map[string]any{"type": "keyword"},
				"total_count":        map[string]any{"type": "integer"},
				"success_count":      map[string]any{"type": "integer"},
				"failed_count":       map[string]any{"type": "integer"},
				"executor_id":        map[string]any{"type": "keyword"},
				"start_time":         map[string]any{"type": "long"},
				"end_time":           map[string]any{"type": "long"},
				"duration_ms":        map[string]any{"type": "long"},
				"results":            map[string]any{"type": "nested"},
				"dynamic_params":     map[string]any{"type": "object", "enabled": false},
				"action_source":      map[string]any{"type": "object", "enabled": false},
			},
		},
	}

	if err := s.osAccess.CreateIndex(ctx, indexName, indexBody); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	logger.Infof("Created index: %s", indexName)

	// Wait a bit for the index to be ready
	time.Sleep(100 * time.Millisecond)

	return nil
}

// structToMap converts a struct to a map
func structToMap(v any) map[string]any {
	data, _ := json.Marshal(v)
	var result map[string]any
	_ = json.Unmarshal(data, &result)
	return result
}

// mapToActionExecution converts a map to ActionExecution
func mapToActionExecution(m map[string]any) (*interfaces.ActionExecution, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	var exec interfaces.ActionExecution
	if err := json.Unmarshal(data, &exec); err != nil {
		return nil, err
	}

	return &exec, nil
}
