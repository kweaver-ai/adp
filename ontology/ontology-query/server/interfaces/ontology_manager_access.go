package interfaces

import "context"

// 基于起点、方向和路径长度获取对象子图的请求体
type PathsQueryBaseOnSource struct {
	ConceptGroups     []string `json:"concept_groups,omitempty"`
	SourceObjecTypeId string   `json:"source_object_type_id"`
	Direction         string   `json:"direction"`
	PathLength        int      `json:"path_length"`

	KNID string `json:"-"`
	// IncludeTypeInfo bool   `json:"-"`
}

//go:generate mockgen -source ../interfaces/ontology_manager_access.go -destination ../interfaces/mock/mock_ontology_manager_access.go
type OntologyManagerAccess interface {
	GetObjectType(ctx context.Context, knID string, branch string, otId string) (ObjectType, bool, error)
	GetRelationType(ctx context.Context, knID string, branch string, rtId string) (RelationType, bool, error)
	GetActionType(ctx context.Context, knID string, branch string, atId string) (ActionType, bool, error)
	GetRelationTypePathsBaseOnSource(ctx context.Context, knID string, branch string, query PathsQueryBaseOnSource) ([]RelationTypePath, error)
}
