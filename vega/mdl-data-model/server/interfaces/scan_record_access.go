// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

type ScanRecordStatus struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

//go:generate mockgen -source ../interfaces/scan_record_access.go -destination ../interfaces/mock/mock_scan_record_access.go
type ScanRecordAccess interface {
	ListScanRecords(ctx context.Context, params *PaginationQueryParameters) (scanRecords []*ScanRecord, err error)
	GetByDataSourceId(ctx context.Context, datasourceId string) (scanRecord *ScanRecord, exist bool, err error)
	GetByTaskIds(ctx context.Context, taskIds []string) (scanRecord []*ScanRecord, err error)
	GetByDataSourceIdAndScanner(ctx context.Context, datasourceId string, scanner string) (scanRecord []*ScanRecord, err error)
	CreateScanRecord(ctx context.Context, scanRecord *ScanRecord) error
	UpdateScanRecord(ctx context.Context, scanRecord *ScanRecord) error
	UpdateScanRecordStatus(ctx context.Context, status *ScanRecordStatus) error
	DeleteByDataSourceId(ctx context.Context, datasourceId string) error
}
