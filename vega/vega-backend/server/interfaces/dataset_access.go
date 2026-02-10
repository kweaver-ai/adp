// Package interfaces defines entities, DTOs, and service interfaces.
package interfaces

import "context"

// DatasetAccess 定义 dataset 数据访问接口
type DatasetAccess interface {
	Create(ctx context.Context, name string, schemaDefinition []Property) error
	Update(ctx context.Context, name string, schemaDefinition []Property) error
	Delete(ctx context.Context, name string) error
	CheckExist(ctx context.Context, name string) (bool, error)
	ListDocuments(ctx context.Context, name string, params *QueryParams) ([]map[string]any, int64, error)
	CreateDocuments(ctx context.Context, name string, documents []map[string]any) ([]string, error)
	GetDocument(ctx context.Context, name string, docID string) (map[string]any, error)
	UpdateDocument(ctx context.Context, name string, docID string, document map[string]any) error
	DeleteDocument(ctx context.Context, name string, docID string) error
}
