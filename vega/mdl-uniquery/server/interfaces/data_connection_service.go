// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

//go:generate mockgen -source ../interfaces/data_connection_service.go -destination ../interfaces/mock/mock_data_connection_service.go
type DataConnectionService interface {
	GetDataConnectionByID(ctx context.Context, connID string) (*DataConnection, bool, error)
}
