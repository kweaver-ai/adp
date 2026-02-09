// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package interfaces defines entities, DTOs, and service interfaces.
package interfaces

import "context"

// DiscoveryTaskService defines discovery task business logic interface.
type DiscoveryTaskService interface {
	// Create creates a new DiscoveryTask and sends message to Kafka.
	Create(ctx context.Context, catalogID string) (string, error)
	// GetByID retrieves a DiscoveryTask by ID.
	GetByID(ctx context.Context, id string) (*DiscoveryTask, error)
	// List lists DiscoveryTasks for a catalog.
	List(ctx context.Context, params DiscoveryTaskQueryParams) ([]*DiscoveryTask, int64, error)
	// UpdateStatus updates a DiscoveryTask's status.
	UpdateStatus(ctx context.Context, id string, status string, message string, stime int64) error
	// UpdateResult updates a DiscoveryTask's result.
	UpdateResult(ctx context.Context, id string, result *DiscoveryResult, stime int64) error
}
