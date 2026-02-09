// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package interfaces defines entities, DTOs, and service interfaces.
package interfaces

import "context"

// DiscoveryTaskAccess defines discovery task data access interface.
type DiscoveryTaskAccess interface {
	// Create creates a new DiscoveryTask.
	Create(ctx context.Context, task *DiscoveryTask) error
	// GetByID retrieves a DiscoveryTask by ID.
	GetByID(ctx context.Context, id string) (*DiscoveryTask, error)
	// List lists DiscoveryTasks with filters.
	List(ctx context.Context, params DiscoveryTaskQueryParams) ([]*DiscoveryTask, int64, error)
	// UpdateStatus updates a DiscoveryTask's status and message.
	UpdateStatus(ctx context.Context, id, status, message string, stime int64) error
	// UpdateProgress updates a DiscoveryTask's progress.
	UpdateProgress(ctx context.Context, id string, progress int) error
	// UpdateResult updates a DiscoveryTask's result and sets status to completed.
	UpdateResult(ctx context.Context, id string, result *DiscoveryResult, stime int64) error
}
