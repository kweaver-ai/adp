package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/object_type_service.go -destination ../interfaces/mock/mock_object_type_service.go
type ObjectTypeService interface {
	CheckObjectTypeExistByID(ctx context.Context, knID string, branch string, otID string) (string, bool, error)
	CheckObjectTypeExistByName(ctx context.Context, knID string, branch string, otName string) (string, bool, error)
	CreateObjectTypes(ctx context.Context, tx *sql.Tx, objectTypes []*ObjectType, mode string, needCreateConceptGroupRelation bool) ([]string, error)
	ListObjectTypes(ctx context.Context, tx *sql.Tx, query ObjectTypesQueryParams) ([]*ObjectType, int, error)
	GetObjectTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, otIDs []string) ([]*ObjectType, error)
	UpdateObjectType(ctx context.Context, tx *sql.Tx, objectType *ObjectType) error
	UpdateDataProperties(ctx context.Context, objectType *ObjectType, dataProperties []*DataProperty) error
	DeleteObjectTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, otIDs []string) error

	GetObjectTypeByID(ctx context.Context, tx *sql.Tx, knID string, branch string, otID string) (*ObjectType, error)
	GetAllObjectTypesByKnID(ctx context.Context, knID string, branch string) (map[string]*ObjectType, error)
	GetObjectTypeIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error)
	DeleteObjectTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) error

	SearchObjectTypes(ctx context.Context, query *ConceptsQuery) (ObjectTypes, error)

	// 获取对象类基本信息（无翻译依赖资源）
	GetObjectTypesMapByIDs(ctx context.Context, knID string, branch string, otIDs []string, needPropMap bool) (map[string]*ObjectType, error)

	// 对象类写索引
	InsertOpenSearchData(ctx context.Context, objectTypes []*ObjectType) error
}
