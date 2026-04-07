package model

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source=skill_runtime_profile.go -destination=../../mocks/model_skill_runtime_profile.go -package=mocks

// SkillRuntimeProfileDB Skill 执行配置表
type SkillRuntimeProfileDB struct {
	ID              int64  `json:"id" db:"f_id"`
	SkillID         string `json:"skill_id" db:"f_skill_id"`
	SkillVersion    string `json:"skill_version" db:"f_skill_version"`
	Entrypoint      string `json:"entrypoint" db:"f_entrypoint"`
	Name            string `json:"name" db:"f_name"`
	Description     string `json:"description" db:"f_description"`
	RuntimeType     string `json:"runtime_type" db:"f_runtime_type"`
	CommandTemplate string `json:"command_template" db:"f_command_template"`
	InputSchema     string `json:"input_schema" db:"f_input_schema"`
	OutputSchema    string `json:"output_schema" db:"f_output_schema"`
	Timeout         int    `json:"timeout" db:"f_timeout"`
	Status          string `json:"status" db:"f_status"`
	ExtendInfo      string `json:"extend_info" db:"f_extend_info"`
	CreateTime      int64  `json:"create_time" db:"f_create_time"`
	CreateUser      string `json:"create_user" db:"f_create_user"`
	UpdateTime      int64  `json:"update_time" db:"f_update_time"`
	UpdateUser      string `json:"update_user" db:"f_update_user"`
}

// ISkillRuntimeProfile Skill 执行配置接口
type ISkillRuntimeProfile interface {
	InsertSkillRuntimeProfile(ctx context.Context, tx *sql.Tx, profile *SkillRuntimeProfileDB) error
	UpdateSkillRuntimeProfile(ctx context.Context, tx *sql.Tx, profile *SkillRuntimeProfileDB) error
	SelectSkillRuntimeProfileBySkillIDAndEntrypoint(ctx context.Context, tx *sql.Tx, skillID, version, entrypoint string) (profile *SkillRuntimeProfileDB, err error)
	SelectSkillRuntimeProfilesBySkillID(ctx context.Context, tx *sql.Tx, skillID, version string) (profiles []*SkillRuntimeProfileDB, err error)
}
