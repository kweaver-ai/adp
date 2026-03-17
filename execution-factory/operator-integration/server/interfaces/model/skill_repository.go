package model

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
)

//go:generate mockgen -source=skill_repository.go -destination=../../mocks/model_skill_repository.go -package=mocks

// SkillStatus Skill 状态
type SkillStatus string

const (
	SkillStatusDraft    SkillStatus = "draft"
	SkillStatusActive   SkillStatus = "active"
	SkillStatusError    SkillStatus = "error"
	SkillStatusDeleting SkillStatus = "deleting"
	SkillStatusDeleted  SkillStatus = "deleted"
)

// SkillRepositoryDB Skill 主表
type SkillRepositoryDB struct {
	ID           int64       `json:"id" db:"f_id"`
	SkillID      string      `json:"skill_id" db:"f_skill_id"`
	Name         string      `json:"name" db:"f_name"`
	Description  string      `json:"description" db:"f_description"`
	Instructions string      `json:"instructions" db:"f_instructions"`
	Version      string      `json:"version" db:"f_version"`
	Status       SkillStatus `json:"status" db:"f_status"`
	Source       string      `json:"source" db:"f_source"`
	OwnerType    string      `json:"owner_type" db:"f_owner_type"`
	OwnerID      string      `json:"owner_id" db:"f_owner_id"`
	ExtendInfo   string      `json:"extend_info" db:"f_extend_info"`
	Dependencies string      `json:"dependencies" db:"f_dependencies"`
	FileManifest string      `json:"file_manifest" db:"f_file_manifest"`
	CreateTime   int64       `json:"create_time" db:"f_create_time"`
	CreateUser   string      `json:"create_user" db:"f_create_user"`
	UpdateTime   int64       `json:"update_time" db:"f_update_time"`
	UpdateUser   string      `json:"update_user" db:"f_update_user"`
	DeleteTime   int64       `json:"delete_time" db:"f_delete_time"`
	DeleteUser   string      `json:"delete_user" db:"f_delete_user"`
}

// GetBizID 获取业务 ID
func (s *SkillRepositoryDB) GetBizID() string {
	return s.SkillID
}

// ISkillRepository Skill 主表接口
type ISkillRepository interface {
	InsertSkill(ctx context.Context, tx *sql.Tx, skill *SkillRepositoryDB) (skillID string, err error)
	UpdateSkill(ctx context.Context, tx *sql.Tx, skill *SkillRepositoryDB) error
	UpdateSkillStatus(ctx context.Context, tx *sql.Tx, skillID string, status SkillStatus, updateUser string) error
	SelectSkillByID(ctx context.Context, tx *sql.Tx, skillID string) (skill *SkillRepositoryDB, err error)
	SelectSkillListPage(ctx context.Context, tx *sql.Tx, filter map[string]interface{},
		sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) (skills []*SkillRepositoryDB, err error)
	CountByWhereClause(ctx context.Context, tx *sql.Tx, filter map[string]interface{}) (count int64, err error)
	DeleteSkillByID(ctx context.Context, tx *sql.Tx, skillID string) error
}
