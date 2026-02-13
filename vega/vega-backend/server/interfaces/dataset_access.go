// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// DatasetAccess 定义 dataset 数据访问接口
//
//go:generate mockgen -source ../interfaces/dataset_access.go -destination ../interfaces/mock/mock_dataset_access.go
type DatasetAccess interface {
	Create(ctx context.Context, name string, schemaDefinition []*Property) error
	Update(ctx context.Context, name string, schemaDefinition []*Property) error
	Delete(ctx context.Context, name string) error
	CheckExist(ctx context.Context, name string) (bool, error)
	ListDocuments(ctx context.Context, name string, params *ResourceDataQueryParams) ([]map[string]any, int64, error)
	CreateDocuments(ctx context.Context, name string, documents []map[string]any) ([]string, error)
	GetDocument(ctx context.Context, name string, docID string) (map[string]any, error)
	UpdateDocument(ctx context.Context, name string, docID string, document map[string]any) error
	DeleteDocument(ctx context.Context, name string, docID string) error
	UpdateDocuments(ctx context.Context, name string, updateRequests []map[string]any) error
	DeleteDocuments(ctx context.Context, name string, docIDs string) error
}
