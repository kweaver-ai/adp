package dbaccess

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/db"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

type skillRuntimeProfileDB struct {
	dbPool *sqlx.DB
	dbName string
	orm    *ormhelper.DB
}

var (
	skillRuntimeProfileOnce sync.Once
	skillRuntimeProfileInst model.ISkillRuntimeProfile
)

const tbSkillRuntimeProfile = "t_skill_runtime_profile"

func NewSkillRuntimeProfileDB() model.ISkillRuntimeProfile {
	skillRuntimeProfileOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		dbPool := db.NewDBPool()
		dbName := confLoader.GetDBName()
		orm := ormhelper.New(dbPool, dbName)
		skillRuntimeProfileInst = &skillRuntimeProfileDB{
			dbPool: dbPool,
			dbName: dbName,
			orm:    orm,
		}
	})
	return skillRuntimeProfileInst
}

func (s *skillRuntimeProfileDB) InsertSkillRuntimeProfile(ctx context.Context, tx *sql.Tx, profile *model.SkillRuntimeProfileDB) error {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	now := time.Now().UnixNano()
	profile.CreateTime = now
	profile.UpdateTime = now
	_, err := orm.Insert().Into(tbSkillRuntimeProfile).Values(map[string]interface{}{
		"f_skill_id":         profile.SkillID,
		"f_skill_version":    profile.SkillVersion,
		"f_entrypoint":       profile.Entrypoint,
		"f_name":             profile.Name,
		"f_description":      profile.Description,
		"f_runtime_type":     profile.RuntimeType,
		"f_command_template": profile.CommandTemplate,
		"f_input_schema":     profile.InputSchema,
		"f_output_schema":    profile.OutputSchema,
		"f_timeout":          profile.Timeout,
		"f_status":           profile.Status,
		"f_extend_info":      profile.ExtendInfo,
		"f_create_time":      profile.CreateTime,
		"f_create_user":      profile.CreateUser,
		"f_update_time":      profile.UpdateTime,
		"f_update_user":      profile.UpdateUser,
	}).Execute(ctx)
	return err
}

func (s *skillRuntimeProfileDB) UpdateSkillRuntimeProfile(ctx context.Context, tx *sql.Tx, profile *model.SkillRuntimeProfileDB) error {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	profile.UpdateTime = time.Now().UnixNano()
	_, err := orm.Update(tbSkillRuntimeProfile).SetData(map[string]interface{}{
		"f_name":             profile.Name,
		"f_description":      profile.Description,
		"f_runtime_type":     profile.RuntimeType,
		"f_command_template": profile.CommandTemplate,
		"f_input_schema":     profile.InputSchema,
		"f_output_schema":    profile.OutputSchema,
		"f_timeout":          profile.Timeout,
		"f_status":           profile.Status,
		"f_extend_info":      profile.ExtendInfo,
		"f_update_time":      profile.UpdateTime,
		"f_update_user":      profile.UpdateUser,
	}).WhereEq("f_skill_id", profile.SkillID).WhereEq("f_skill_version", profile.SkillVersion).WhereEq("f_entrypoint", profile.Entrypoint).Execute(ctx)
	return err
}

func (s *skillRuntimeProfileDB) SelectSkillRuntimeProfileBySkillIDAndEntrypoint(ctx context.Context, tx *sql.Tx, skillID, version, entrypoint string) (profile *model.SkillRuntimeProfileDB, err error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	profile = &model.SkillRuntimeProfileDB{}
	err = orm.Select().From(tbSkillRuntimeProfile).WhereEq("f_skill_id", skillID).WhereEq("f_skill_version", version).WhereEq("f_entrypoint", entrypoint).First(ctx, profile)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (s *skillRuntimeProfileDB) SelectSkillRuntimeProfilesBySkillID(ctx context.Context, tx *sql.Tx, skillID, version string) (profiles []*model.SkillRuntimeProfileDB, err error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	profiles = []*model.SkillRuntimeProfileDB{}
	err = orm.Select().From(tbSkillRuntimeProfile).WhereEq("f_skill_id", skillID).WhereEq("f_skill_version", version).Get(ctx, &profiles)
	return profiles, err
}
