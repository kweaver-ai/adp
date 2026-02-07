// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

type DataConnection struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Tags             []string `json:"tags"`
	Comment          string   `json:"comment"`
	CreateTime       int64    `json:"create_time"`
	UpdateTime       int64    `json:"update_time"`
	DataSourceType   string   `json:"data_source_type"`
	DataSourceConfig any      `json:"config"`
	ApplicationScope []string `json:"application_scope"`

	DataConnectionStatus `json:",inline"`
}

type DataConnectionStatus struct {
	ID            string `json:"-"`
	Status        string `json:"status"`
	DetectionTime int64  `json:"detection_time"`
}

//go:generate mockgen -source ../interfaces/data_connection_access.go -destination ../interfaces/mock/mock_data_connection_access.go
type DataConnectionAccess interface {
	GetDataConnectionByID(ctx context.Context, connID string) (*DataConnection, bool, error)
	GetDataConnectionTypeByName(ctx context.Context, connName string) (string, bool, error)
}
