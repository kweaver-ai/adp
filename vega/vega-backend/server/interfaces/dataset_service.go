// Package interfaces defines entities, DTOs, and service interfaces.
package interfaces

import "context"

// DatasetService 定义 dataset 业务逻辑接口
type DatasetService interface {
    Create(ctx context.Context, res *Resource) error
    Update(ctx context.Context, res *Resource) error
    Delete(ctx context.Context, res *Resource) error
    ListDocuments(ctx context.Context, id string, params *QueryParams) ([]map[string]any, int64, error)
    CreateDocuments(ctx context.Context, id string, documents []map[string]any) ([]string, error)
    GetDocument(ctx context.Context, id string, docID string) (map[string]any, error)
    UpdateDocument(ctx context.Context, id string, docID string, document map[string]any) error
    DeleteDocument(ctx context.Context, id string, docID string) error
}
