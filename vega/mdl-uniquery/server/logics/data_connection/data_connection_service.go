// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_connection

import (
	"context"
	"sync"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"

	"uniquery/common"
	"uniquery/interfaces"
	"uniquery/logics"
)

var (
	dcServiceOnce sync.Once
	dcService     interfaces.DataConnectionService
)

type dataConnectionService struct {
	appSetting *common.AppSetting
	dcAccess   interfaces.DataConnectionAccess
}

func NewDataConnectionService(appsetting *common.AppSetting) interfaces.DataConnectionService {
	dcServiceOnce.Do(func() {
		dcService = &dataConnectionService{
			appSetting: appsetting,
			dcAccess:   logics.DCAccess,
		}
	})
	return dcService
}

func (dcs *dataConnectionService) GetDataConnectionByID(ctx context.Context,
	connID string) (*interfaces.DataConnection, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get single view data for internal module")
	defer span.End()

	conn, exist, err := dcs.dcAccess.GetDataConnectionByID(ctx, connID)

	return conn, exist, err
}
