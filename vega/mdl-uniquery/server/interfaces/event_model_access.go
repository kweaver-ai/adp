// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

//go:generate mockgen -source ../interfaces/event_model_access.go -destination ../interfaces/mock/mock_event_model_access.go
type EventModelAccess interface {
	GetEventModelById(ctx context.Context, id string) ([]EventModel, error)
	GetEventModelBySourceId(ctx context.Context, dataSource string) ([]EventModel, error)
}
