// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"

	"github.com/hibiken/asynq"
)

// DiscoveryResult represents the result of a discovery operation.
type DiscoveryResult struct {
	CatalogID      string `json:"catalog_id"`
	NewCount       int    `json:"new_count"`
	StaleCount     int    `json:"stale_count"`
	UnchangedCount int    `json:"unchanged_count"`
	Message        string `json:"message"`
}

// DiscoveryWorker interface defines discovery execution functionality.
// This worker is called by the task management service to execute the actual discovery.
//
//go:generate mockgen -source ../interfaces/discovery_worker.go -destination ../interfaces/mock/mock_discovery_worker.go
type DiscoveryWorker interface {
	Start()

	Run(ctx context.Context) error
	ProcessTask(ctx context.Context, event *asynq.Task) error
}
