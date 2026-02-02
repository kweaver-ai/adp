// Package interfaces defines entities, DTOs, and service interfaces.
package interfaces

// DiscoveryTask status constants.
const (
	DiscoveryTaskStatusPending   = "pending"
	DiscoveryTaskStatusRunning   = "running"
	DiscoveryTaskStatusCompleted = "completed"
	DiscoveryTaskStatusFailed    = "failed"

	// KafkaTopic is the topic for discovery task messages.
	DiscoveryTaskTopic = "adp-vega-discovery-task"
)

// DiscoveryTask represents a discovery task entity.
type DiscoveryTask struct {
	ID        string `json:"id"`
	CatalogID string `json:"catalog_id"`

	Status     string           `json:"status"`   // pending/running/completed/failed
	Progress   int              `json:"progress"` // 0-100
	Message    string           `json:"message"`
	StartTime  int64            `json:"start_time,omitempty"`  // 开始执行时间
	FinishTime int64            `json:"finish_time,omitempty"` // 完成时间
	Result     *DiscoveryResult `json:"result,omitempty"`

	Creator    AccountInfo `json:"creator"`
	CreateTime int64       `json:"create_time"`
}

// DiscoveryTaskQueryParams holds discovery task list query parameters.
type DiscoveryTaskQueryParams struct {
	PaginationParams
	CatalogID string
	Status    string
}

// DiscoveryTaskMessage represents the Kafka message for discovery task.
type DiscoveryTaskMessage struct {
	TaskID string `json:"task_id"`
}
