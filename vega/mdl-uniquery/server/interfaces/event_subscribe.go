// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "time"

const (
	EXPIRATION_TIME = 2 * time.Minute
	DELETE_TIME     = 5 * time.Minute
)

type EventSubService interface {
	Subscribe(exitCh chan bool) error
}

type DataSource struct {
	DataSourceId   string
	DataSourceType string
}
