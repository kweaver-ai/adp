// Package mysql 数据库操作
package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/event"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/utils"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/utils/data"
	"github.com/shiningrush/goevent"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// StoreOption Store 配置项
type StoreOption struct {
	DSN      string
	MaxConns int
	MinConns int
}

// Store 结构体
type Store struct {
	config *common.Config
	opt    *StoreOption
	db     *sql.DB
}

// NewStore 创建Store实例
func NewStore(option *StoreOption) *Store {
	return &Store{
		config: common.NewConfig(),
		opt:    option,
	}
}

// Init store 初始化
func (s *Store) Init() error {
	var err error
	s.db, err = sql.Open("mysql", s.opt.DSN)
	if err != nil {
		return fmt.Errorf("connect to mysql failed: %w", err)
	}

	s.db.SetMaxOpenConns(s.opt.MaxConns)
	s.db.SetMaxIdleConns(s.opt.MinConns)
	s.db.SetConnMaxLifetime(time.Hour)

	if err := s.db.Ping(); err != nil {
		return fmt.Errorf("ping mysql failed: %w", err)
	}

	if err := s.createTables(); err != nil {
		return fmt.Errorf("create tables failed: %w", err)
	}

	return nil
}

// createTables 创建表结构
func (s *Store) createTables() error {
	// 创建dags表
	dagTableSQL := `
	CREATE TABLE IF NOT EXISTS dags (
		id VARCHAR(36) PRIMARY KEY,
		userid VARCHAR(255) NOT NULL,
		name VARCHAR(255) NOT NULL,
		desc TEXT,
		trigger VARCHAR(50) NOT NULL,
		cron VARCHAR(100),
		status VARCHAR(50) NOT NULL DEFAULT 'normal',
		tasks JSON NOT NULL,
		steps JSON,
		description TEXT,
		shortcuts JSON,
		accessors JSON,
		type VARCHAR(100),
		policy_type VARCHAR(100),
		appinfo JSON,
		priority VARCHAR(50),
		removed BOOLEAN NOT NULL DEFAULT FALSE,
		emails JSON,
		template VARCHAR(255),
		published BOOLEAN NOT NULL DEFAULT FALSE,
		trigger_config JSON,
		sub_ids JSON,
		exec_mode VARCHAR(50),
		category VARCHAR(100),
		outputs JSON,
		instructions JSON,
		operator_id VARCHAR(36),
		inc_values JSON,
		version VARCHAR(50),
		version_id VARCHAR(36),
		modify_by VARCHAR(255),
		is_debug BOOLEAN NOT NULL DEFAULT FALSE,
		debug_id VARCHAR(36),
		biz_domain_id VARCHAR(36),
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 创建dag_instances表
	dagInstanceTableSQL := `
	CREATE TABLE IF NOT EXISTS dag_instances (
		id VARCHAR(36) PRIMARY KEY,
		dag_id VARCHAR(36) NOT NULL,
		trigger VARCHAR(50) NOT NULL,
		worker VARCHAR(255),
		source VARCHAR(255),
		vars JSON,
		keywords JSON,
		event_persistence INT NOT NULL DEFAULT 0,
		event_oss_path VARCHAR(512),
		share_data JSON,
		status VARCHAR(50) NOT NULL DEFAULT 'init',
		reason TEXT,
		cmd JSON,
		userid VARCHAR(255) NOT NULL,
		ended_at BIGINT,
		dag_type VARCHAR(100),
		policy_type VARCHAR(100),
		appinfo JSON,
		priority VARCHAR(50),
		mode INT NOT NULL DEFAULT 0,
		dump TEXT,
		success_callback VARCHAR(512),
		error_callback VARCHAR(512),
		call_chain JSON,
		resume_data TEXT,
		resume_status VARCHAR(50),
		version VARCHAR(50),
		version_id VARCHAR(36),
		biz_domain_id VARCHAR(36),
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 创建task_instances表
	taskInstanceTableSQL := `
	CREATE TABLE IF NOT EXISTS task_instances (
		id VARCHAR(36) PRIMARY KEY,
		dag_ins_id VARCHAR(36) NOT NULL,
		task_id VARCHAR(36) NOT NULL,
		status VARCHAR(50) NOT NULL DEFAULT 'init',
		reason TEXT,
		traces JSON,
		results JSON,
		last_modified_at BIGINT,
		rendered_params JSON,
		depend_on JSON,
		hash VARCHAR(255),
		metadata JSON,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 创建tokens表
	tokenTableSQL := `
	CREATE TABLE IF NOT EXISTS tokens (
		id VARCHAR(36) PRIMARY KEY,
		userid VARCHAR(255) NOT NULL,
		token TEXT NOT NULL,
		expires_in INT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 创建inboxes表
	inboxTableSQL := `
	CREATE TABLE IF NOT EXISTS inboxes (
		id VARCHAR(36) PRIMARY KEY,
		docid VARCHAR(255) NOT NULL,
		topic VARCHAR(255) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 创建outboxes表
	outboxTableSQL := `
	CREATE TABLE IF NOT EXISTS outboxes (
		id VARCHAR(36) PRIMARY KEY,
		topic VARCHAR(255) NOT NULL,
		message JSON NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 创建clients表
	clientTableSQL := `
	CREATE TABLE IF NOT EXISTS clients (
		id VARCHAR(36) PRIMARY KEY,
		client_name VARCHAR(255) NOT NULL,
		client_id VARCHAR(255) NOT NULL,
		client_secret VARCHAR(255) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 创建switches表
	switchTableSQL := `
	CREATE TABLE IF NOT EXISTS switches (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		status BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 创建logs表
	logTableSQL := `
	CREATE TABLE IF NOT EXISTS logs (
		id VARCHAR(36) PRIMARY KEY,
		level VARCHAR(50) NOT NULL,
		message TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 创建dag_versions表
	dagVersionTableSQL := `
	CREATE TABLE IF NOT EXISTS dag_versions (
		id VARCHAR(36) PRIMARY KEY,
		dag_id VARCHAR(36) NOT NULL,
		version_id VARCHAR(36) NOT NULL,
		config JSON NOT NULL,
		sort_time BIGINT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 执行创建表语句
	tables := []string{
		dagTableSQL,
		dagInstanceTableSQL,
		taskInstanceTableSQL,
		tokenTableSQL,
		inboxTableSQL,
		outboxTableSQL,
		clientTableSQL,
		switchTableSQL,
		logTableSQL,
		dagVersionTableSQL,
	}

	for _, tableSQL := range tables {
		_, err := s.db.Exec(tableSQL)
		if err != nil {
			return err
		}
	}

	// 创建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_dags_userid ON dags (userid);",
		"CREATE INDEX IF NOT EXISTS idx_dags_status ON dags (status);",
		"CREATE INDEX IF NOT EXISTS idx_dags_type ON dags (type);",
		"CREATE INDEX IF NOT EXISTS idx_dags_trigger ON dags (trigger);",
		"CREATE INDEX IF NOT EXISTS idx_dags_removed ON dags (removed);",
		"CREATE INDEX IF NOT EXISTS idx_dag_instances_dag_id ON dag_instances (dag_id);",
		"CREATE INDEX IF NOT EXISTS idx_dag_instances_status ON dag_instances (status);",
		"CREATE INDEX IF NOT EXISTS idx_dag_instances_userid ON dag_instances (userid);",
		"CREATE INDEX IF NOT EXISTS idx_dag_instances_created_at ON dag_instances (created_at);",
		"CREATE INDEX IF NOT EXISTS idx_dag_instances_ended_at ON dag_instances (ended_at);",
		"CREATE INDEX IF NOT EXISTS idx_dag_instances_updated_at ON dag_instances (updated_at);",
		"CREATE INDEX IF NOT EXISTS idx_dag_instances_priority ON dag_instances (priority);",
		"CREATE INDEX IF NOT EXISTS idx_dag_instances_mode ON dag_instances (mode);",
		"CREATE INDEX IF NOT EXISTS idx_task_instances_dag_ins_id ON task_instances (dag_ins_id);",
		"CREATE INDEX IF NOT EXISTS idx_task_instances_status ON task_instances (status);",
		"CREATE INDEX IF NOT EXISTS idx_task_instances_task_id ON task_instances (task_id);",
		"CREATE INDEX IF NOT EXISTS idx_task_instances_updated_at ON task_instances (updated_at);",
		"CREATE INDEX IF NOT EXISTS idx_task_instances_last_modified_at ON task_instances (last_modified_at);",
		"CREATE INDEX IF NOT EXISTS idx_task_instances_hash ON task_instances (hash);",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_tokens_userid ON tokens (userid);",
		"CREATE INDEX IF NOT EXISTS idx_inboxes_docid ON inboxes (docid);",
		"CREATE INDEX IF NOT EXISTS idx_inboxes_topic ON inboxes (topic);",
		"CREATE INDEX IF NOT EXISTS idx_inboxes_created_at ON inboxes (created_at);",
		"CREATE INDEX IF NOT EXISTS idx_outboxes_topic ON outboxes (topic);",
		"CREATE INDEX IF NOT EXISTS idx_outboxes_created_at ON outboxes (created_at);",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_client_name ON clients (client_name);",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_client_id ON clients (client_id);",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_switches_name ON switches (name);",
		"CREATE INDEX IF NOT EXISTS idx_logs_level ON logs (level);",
		"CREATE INDEX IF NOT EXISTS idx_logs_created_at ON logs (created_at);",
		"CREATE INDEX IF NOT EXISTS idx_dag_versions_dag_id ON dag_versions (dag_id);",
		"CREATE INDEX IF NOT EXISTS idx_dag_versions_version_id ON dag_versions (version_id);",
		"CREATE INDEX IF NOT EXISTS idx_dag_versions_sort_time ON dag_versions (sort_time);",
	}

	for _, indexSQL := range indexes {
		_, err := s.db.Exec(indexSQL)
		if err != nil {
			// 忽略索引已存在的错误
			if !strings.Contains(err.Error(), "Duplicate key name") {
				return err
			}
		}
	}

	return nil
}

// Close 关闭存储连接
func (s *Store) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

// GetDB 根据上下文获取数据库连接（支持事务）
func (s *Store) GetDB(ctx context.Context) interface{} {
	if tx, ok := ctx.Value("tx").(*sql.Tx); ok {
		return tx
	}
	return s.db
}

// WithTransaction 通用事务封装，自动管理事务的开始、提交和回滚
func (s *Store) WithTransaction(ctx context.Context, fn func(sessCtx interface{}) error) error {
	// 开始事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}

	// 执行事务函数
	err = fn(tx)
	if err != nil {
		// 回滚事务
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("rollback transaction failed: %w, original error: %v", rollbackErr, err)
		}
		return err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}

	return nil
}

// JSONMarshal 封装JSON序列化
func JSONMarshal(v interface{}) string {
	if v == nil {
		return "null"
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return string(data)
}

// JSONUnmarshal 封装JSON反序列化
func JSONUnmarshal(data string, v interface{}) error {
	if data == "" || data == "null" {
		return nil
	}
	return json.Unmarshal([]byte(data), v)
}

// CreateToken 创建token记录
func (s *Store) CreateToken(token *entity.Token) error {
	token.Initial()

	db := s.GetDB(context.Background())
	var err error

	if tx, ok := db.(*sql.Tx); ok {
		_, err = tx.Exec(
			"INSERT INTO tokens (id, userid, token, expires_in, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())",
			token.ID, token.UserID, token.Token, token.ExpiresIn,
		)
	} else if dbConn, ok := db.(*sql.DB); ok {
		_, err = dbConn.Exec(
			"INSERT INTO tokens (id, userid, token, expires_in, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())",
			token.ID, token.UserID, token.Token, token.ExpiresIn,
		)
	}

	if err != nil {
		return fmt.Errorf("insert token failed: %w", err)
	}

	return nil
}

// UpdateToken 更新token记录
func (s *Store) UpdateToken(token *entity.Token) error {
	token.Update()

	db := s.GetDB(context.Background())
	var err error

	if tx, ok := db.(*sql.Tx); ok {
		_, err = tx.Exec(
			"UPDATE tokens SET token = ?, expires_in = ?, updated_at = NOW() WHERE userid = ?",
			token.Token, token.ExpiresIn, token.UserID,
		)
	} else if dbConn, ok := db.(*sql.DB); ok {
		_, err = dbConn.Exec(
			"UPDATE tokens SET token = ?, expires_in = ?, updated_at = NOW() WHERE userid = ?",
			token.Token, token.ExpiresIn, token.UserID,
		)
	}

	if err != nil {
		return fmt.Errorf("update token failed: %w", err)
	}

	return nil
}

// DeleteToken 删除token记录
func (s *Store) DeleteToken(id string) error {
	db := s.GetDB(context.Background())
	var err error

	if tx, ok := db.(*sql.Tx); ok {
		_, err = tx.Exec("DELETE FROM tokens WHERE id = ?", id)
	} else if dbConn, ok := db.(*sql.DB); ok {
		_, err = dbConn.Exec("DELETE FROM tokens WHERE id = ?", id)
	}

	if err != nil {
		return fmt.Errorf("delete token failed: %w", err)
	}

	return nil
}

// GetTokenByUserID 获取token记录
func (s *Store) GetTokenByUserID(userID string) (*entity.Token, error) {
	token := &entity.Token{}

	db := s.GetDB(context.Background())
	var row *sql.Row

	if tx, ok := db.(*sql.Tx); ok {
		row = tx.QueryRow("SELECT id, userid, token, expires_in, created_at, updated_at FROM tokens WHERE userid = ?", userID)
	} else if dbConn, ok := db.(*sql.DB); ok {
		row = dbConn.QueryRow("SELECT id, userid, token, expires_in, created_at, updated_at FROM tokens WHERE userid = ?", userID)
	}

	err := row.Scan(&token.ID, &token.UserID, &token.Token, &token.ExpiresIn, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return token, nil
		}
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	return token, nil
}

// CreateClient 创建client记录
func (s *Store) CreateClient(clientName, clientID, clientSecret string) error {
	id := utils.GetUUID()

	db := s.GetDB(context.Background())
	var err error

	if tx, ok := db.(*sql.Tx); ok {
		_, err = tx.Exec(
			"INSERT INTO clients (id, client_name, client_id, client_secret, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())",
			id, clientName, clientID, clientSecret,
		)
	} else if dbConn, ok := db.(*sql.DB); ok {
		_, err = dbConn.Exec(
			"INSERT INTO clients (id, client_name, client_id, client_secret, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())",
			id, clientName, clientID, clientSecret,
		)
	}

	if err != nil {
		return fmt.Errorf("insert client failed: %w", err)
	}

	return nil
}

// GetClient 获取client记录
func (s *Store) GetClient(clientName string) (client *entity.Client, err error) {
	client = &entity.Client{}

	db := s.GetDB(context.Background())
	var row *sql.Row

	if tx, ok := db.(*sql.Tx); ok {
		row = tx.QueryRow("SELECT id, client_name, client_id, client_secret FROM clients WHERE client_name = ?", clientName)
	} else if dbConn, ok := db.(*sql.DB); ok {
		row = dbConn.QueryRow("SELECT id, client_name, client_id, client_secret FROM clients WHERE client_name = ?", clientName)
	}

	err = row.Scan(&client.ID, &client.ClientName, &client.ClientID, &client.ClientSecret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrDataNotFound
		}
		return nil, fmt.Errorf("get client failed: %w", err)
	}

	return client, nil
}

// RemoveClient 删除client记录
func (s *Store) RemoveClient(clientName string) (err error) {
	db := s.GetDB(context.Background())
	var res sql.Result

	if tx, ok := db.(*sql.Tx); ok {
		res, err = tx.Exec("DELETE FROM clients WHERE client_name = ?", clientName)
	} else if dbConn, ok := db.(*sql.DB); ok {
		res, err = dbConn.Exec("DELETE FROM clients WHERE client_name = ?", clientName)
	}

	if err != nil {
		return fmt.Errorf("delete client failed: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows failed: %w", err)
	}

	if affected == 0 {
		return data.ErrDataNotFound
	}

	return nil
}

// CreateDag 创建dag记录
func (s *Store) CreateDag(ctx context.Context, dag *entity.Dag) (string, error) {
	dag.Initial()

	// 检查任务连接
	_, err := mod.BuildRootNode(mod.MapTasksToGetter(dag.Tasks))
	if err != nil {
		return "", err
	}

	db := s.GetDB(ctx)
	var execErr error

	if tx, ok := db.(*sql.Tx); ok {
		_, execErr = tx.Exec(
			`INSERT INTO dags (
				id, userid, name, `+"`desc`"+`, trigger, cron, status, tasks, steps, description, 
				shortcuts, accessors, type, policy_type, appinfo, priority, removed, emails, 
				template, published, trigger_config, sub_ids, exec_mode, category, outputs, 
				instructions, operator_id, inc_values, version, version_id, modify_by, 
				is_debug, debug_id, biz_domain_id, created_at, updated_at
			) VALUES (
				?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
			)`,
			dag.ID, dag.UserID, dag.Name, dag.Desc, string(dag.Trigger), dag.Cron, string(dag.Status),
			JSONMarshal(dag.Tasks), JSONMarshal(dag.Steps), dag.Description,
			JSONMarshal(dag.Shortcuts), JSONMarshal(dag.Accessors), dag.Type, dag.PolicyType, JSONMarshal(dag.AppInfo),
			dag.Priority, dag.Removed, JSONMarshal(dag.Emails), dag.Template, dag.Published,
			JSONMarshal(dag.TriggerConfig), JSONMarshal(dag.SubIDs), dag.ExecMode, dag.Category,
			JSONMarshal(dag.OutPuts), JSONMarshal(dag.Instructions), dag.OperatorID, JSONMarshal(dag.IncValues),
			dag.Version.String(), dag.VersionID, dag.ModifyBy, dag.IsDebug, dag.DeBugID, dag.BizDomainID,
		)
	} else if dbConn, ok := db.(*sql.DB); ok {
		_, execErr = dbConn.Exec(
			`INSERT INTO dags (
				id, userid, name, `+"`desc`"+`, trigger, cron, status, tasks, steps, description, 
				shortcuts, accessors, type, policy_type, appinfo, priority, removed, emails, 
				template, published, trigger_config, sub_ids, exec_mode, category, outputs, 
				instructions, operator_id, inc_values, version, version_id, modify_by, 
				is_debug, debug_id, biz_domain_id, created_at, updated_at
			) VALUES (
				?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
			)`,
			dag.ID, dag.UserID, dag.Name, dag.Desc, string(dag.Trigger), dag.Cron, string(dag.Status),
			JSONMarshal(dag.Tasks), JSONMarshal(dag.Steps), dag.Description,
			JSONMarshal(dag.Shortcuts), JSONMarshal(dag.Accessors), dag.Type, dag.PolicyType, JSONMarshal(dag.AppInfo),
			dag.Priority, dag.Removed, JSONMarshal(dag.Emails), dag.Template, dag.Published,
			JSONMarshal(dag.TriggerConfig), JSONMarshal(dag.SubIDs), dag.ExecMode, dag.Category,
			JSONMarshal(dag.OutPuts), JSONMarshal(dag.Instructions), dag.OperatorID, JSONMarshal(dag.IncValues),
			dag.Version.String(), dag.VersionID, dag.ModifyBy, dag.IsDebug, dag.DeBugID, dag.BizDomainID,
		)
	}

	if execErr != nil {
		return "", fmt.Errorf("insert dag failed: %w", execErr)
	}

	return dag.ID, nil
}

// BatchCreateDag 批量创建dag记录
func (s *Store) BatchCreateDag(ctx context.Context, dags []*entity.Dag) ([]*entity.Dag, error) {
	if len(dags) == 0 {
		return dags, nil
	}

	return dags, s.WithTransaction(ctx, func(tx interface{}) error {
		txConn := tx.(*sql.Tx)

		stmt, err := txConn.Prepare(
			`INSERT INTO dags (
				id, userid, name, ` + "`desc`" + `, trigger, cron, status, tasks, steps, description, 
				shortcuts, accessors, type, policy_type, appinfo, priority, removed, emails, 
				template, published, trigger_config, sub_ids, exec_mode, category, outputs, 
				instructions, operator_id, inc_values, version, version_id, modify_by, 
				is_debug, debug_id, biz_domain_id, created_at, updated_at
			) VALUES (
				?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
			)`,
		)
		if err != nil {
			return fmt.Errorf("prepare statement failed: %w", err)
		}
		defer stmt.Close()

		for _, dag := range dags {
			dag.Initial()

			// 检查任务连接
			_, err := mod.BuildRootNode(mod.MapTasksToGetter(dag.Tasks))
			if err != nil {
				return err
			}

			_, err = stmt.Exec(
				dag.ID, dag.UserID, dag.Name, dag.Desc, string(dag.Trigger), dag.Cron, string(dag.Status),
				JSONMarshal(dag.Tasks), JSONMarshal(dag.Steps), dag.Description,
				JSONMarshal(dag.Shortcuts), JSONMarshal(dag.Accessors), dag.Type, dag.PolicyType, JSONMarshal(dag.AppInfo),
				dag.Priority, dag.Removed, JSONMarshal(dag.Emails), dag.Template, dag.Published,
				JSONMarshal(dag.TriggerConfig), JSONMarshal(dag.SubIDs), dag.ExecMode, dag.Category,
				JSONMarshal(dag.OutPuts), JSONMarshal(dag.Instructions), dag.OperatorID, JSONMarshal(dag.IncValues),
				dag.Version.String(), dag.VersionID, dag.ModifyBy, dag.IsDebug, dag.DeBugID, dag.BizDomainID,
			)
			if err != nil {
				return fmt.Errorf("insert dag failed: %w", err)
			}
		}

		return nil
	})
}

// CreateDagIns 创建dag instance记录
func (s *Store) CreateDagIns(ctx context.Context, dagIns *entity.DagInstance) (string, error) {
	dagIns.Initial()

	db := s.GetDB(ctx)
	var execErr error

	if tx, ok := db.(*sql.Tx); ok {
		_, execErr = tx.Exec(
			`INSERT INTO dag_instances (
				id, dag_id, trigger, worker, source, vars, keywords, event_persistence, 
				event_oss_path, share_data, status, reason, cmd, userid, ended_at, 
				dag_type, policy_type, appinfo, priority, mode, dump, success_callback, 
				error_callback, call_chain, resume_data, resume_status, version, version_id, 
				biz_domain_id, created_at, updated_at
			) VALUES (
				?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
			)`,
			dagIns.ID, dagIns.DagID, string(dagIns.Trigger), dagIns.Worker, dagIns.Source,
			JSONMarshal(dagIns.Vars), JSONMarshal(dagIns.Keywords), dagIns.EventPersistence,
			dagIns.EventOssPath, JSONMarshal(dagIns.ShareData), string(dagIns.Status),
			dagIns.Reason, JSONMarshal(dagIns.Cmd), dagIns.UserID, dagIns.EndedAt,
			dagIns.DagType, dagIns.PolicyType, JSONMarshal(dagIns.AppInfo), dagIns.Priority,
			dagIns.Mode, dagIns.Dump, dagIns.SuccessCallback, dagIns.ErrorCallback,
			JSONMarshal(dagIns.CallChain), dagIns.ResumeData, string(dagIns.ResumeStatus),
			dagIns.Version.String(), dagIns.VersionID, dagIns.BizDomainID,
		)
	} else if dbConn, ok := db.(*sql.DB); ok {
		_, execErr = dbConn.Exec(
			`INSERT INTO dag_instances (
				id, dag_id, trigger, worker, source, vars, keywords, event_persistence, 
				event_oss_path, share_data, status, reason, cmd, userid, ended_at, 
				dag_type, policy_type, appinfo, priority, mode, dump, success_callback, 
				error_callback, call_chain, resume_data, resume_status, version, version_id, 
				biz_domain_id, created_at, updated_at
			) VALUES (
				?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
			)`,
			dagIns.ID, dagIns.DagID, string(dagIns.Trigger), dagIns.Worker, dagIns.Source,
			JSONMarshal(dagIns.Vars), JSONMarshal(dagIns.Keywords), dagIns.EventPersistence,
			dagIns.EventOssPath, JSONMarshal(dagIns.ShareData), string(dagIns.Status),
			dagIns.Reason, JSONMarshal(dagIns.Cmd), dagIns.UserID, dagIns.EndedAt,
			dagIns.DagType, dagIns.PolicyType, JSONMarshal(dagIns.AppInfo), dagIns.Priority,
			dagIns.Mode, dagIns.Dump, dagIns.SuccessCallback, dagIns.ErrorCallback,
			JSONMarshal(dagIns.CallChain), dagIns.ResumeData, string(dagIns.ResumeStatus),
			dagIns.Version.String(), dagIns.VersionID, dagIns.BizDomainID,
		)
	}

	if execErr != nil {
		return "", fmt.Errorf("insert dag instance failed: %w", execErr)
	}

	return dagIns.ID, nil
}

// BatchCreateDagIns 批量创建dag instance
func (s *Store) BatchCreateDagIns(ctx context.Context, dagIns []*entity.DagInstance) ([]*entity.DagInstance, error) {
	if len(dagIns) == 0 {
		return dagIns, nil
	}

	return dagIns, s.WithTransaction(ctx, func(tx interface{}) error {
		txConn := tx.(*sql.Tx)

		stmt, err := txConn.Prepare(
			`INSERT INTO dag_instances (
				id, dag_id, trigger, worker, source, vars, keywords, event_persistence, 
				event_oss_path, share_data, status, reason, cmd, userid, ended_at, 
				dag_type, policy_type, appinfo, priority, mode, dump, success_callback, 
				error_callback, call_chain, resume_data, resume_status, version, version_id, 
				biz_domain_id, created_at, updated_at
			) VALUES (
				?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
			)`,
		)
		if err != nil {
			return fmt.Errorf("prepare statement failed: %w", err)
		}
		defer stmt.Close()

		for _, ins := range dagIns {
			ins.Initial()

			_, err = stmt.Exec(
				ins.ID, ins.DagID, string(ins.Trigger), ins.Worker, ins.Source,
				JSONMarshal(ins.Vars), JSONMarshal(ins.Keywords), ins.EventPersistence,
				ins.EventOssPath, JSONMarshal(ins.ShareData), string(ins.Status),
				ins.Reason, JSONMarshal(ins.Cmd), ins.UserID, ins.EndedAt,
				ins.DagType, ins.PolicyType, JSONMarshal(ins.AppInfo), ins.Priority,
				ins.Mode, ins.Dump, ins.SuccessCallback, ins.ErrorCallback,
				JSONMarshal(ins.CallChain), ins.ResumeData, string(ins.ResumeStatus),
				ins.Version.String(), ins.VersionID, ins.BizDomainID,
			)
			if err != nil {
				return fmt.Errorf("insert dag instance failed: %w", err)
			}
		}

		return nil
	})
}

// BatchDeleteDagIns 批量删除dag instance
func (s *Store) BatchDeleteDagIns(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	return s.WithTransaction(ctx, func(tx interface{}) error {
		txConn := tx.(*sql.Tx)

		// 构建IN子句
		placeholders := make([]string, len(ids))
		args := make([]interface{}, len(ids))
		for i, id := range ids {
			placeholders[i] = "?"
			args[i] = id
		}

		query := fmt.Sprintf("DELETE FROM dag_instances WHERE id IN (%s)", strings.Join(placeholders, ","))
		_, err := txConn.Exec(query, args...)
		if err != nil {
			return fmt.Errorf("delete dag instances failed: %w", err)
		}

		return nil
	})
}

// CreateTaskIns 创建task instance
func (s *Store) CreateTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error {
	taskIns.Initial()

	db := s.GetDB(ctx)
	var execErr error

	if tx, ok := db.(*sql.Tx); ok {
		_, execErr = tx.Exec(
			`INSERT INTO task_instances (
				id, dag_ins_id, task_id, status, reason, traces, results, last_modified_at, 
				rendered_params, depend_on, hash, metadata, created_at, updated_at
			) VALUES (
				?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
			)`,
			taskIns.ID, taskIns.DagInsID, taskIns.TaskID, string(taskIns.Status),
			taskIns.Reason, JSONMarshal(taskIns.Traces), JSONMarshal(taskIns.Results),
			taskIns.LastModifiedAt, JSONMarshal(taskIns.RenderedParams),
			JSONMarshal(taskIns.DependOn), taskIns.Hash, JSONMarshal(taskIns.MetaData),
		)
	} else if dbConn, ok := db.(*sql.DB); ok {
		_, execErr = dbConn.Exec(
			`INSERT INTO task_instances (
				id, dag_ins_id, task_id, status, reason, traces, results, last_modified_at, 
				rendered_params, depend_on, hash, metadata, created_at, updated_at
			) VALUES (
				?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
			)`,
			taskIns.ID, taskIns.DagInsID, taskIns.TaskID, string(taskIns.Status),
			taskIns.Reason, JSONMarshal(taskIns.Traces), JSONMarshal(taskIns.Results),
			taskIns.LastModifiedAt, JSONMarshal(taskIns.RenderedParams),
			JSONMarshal(taskIns.DependOn), taskIns.Hash, JSONMarshal(taskIns.MetaData),
		)
	}

	if execErr != nil {
		return fmt.Errorf("insert task instance failed: %w", execErr)
	}

	return nil
}

// BatchCreateTaskIns 批量创建task instance
func (s *Store) BatchCreateTaskIns(ctx context.Context, taskIns []*entity.TaskInstance) ([]*entity.TaskInstance, error) {
	if len(taskIns) == 0 {
		return taskIns, nil
	}

	return taskIns, s.WithTransaction(ctx, func(tx interface{}) error {
		txConn := tx.(*sql.Tx)

		stmt, err := txConn.Prepare(
			`INSERT INTO task_instances (
				id, dag_ins_id, task_id, status, reason, traces, results, last_modified_at, 
				rendered_params, depend_on, hash, metadata, created_at, updated_at
			) VALUES (
				?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
			)`,
		)
		if err != nil {
			return fmt.Errorf("prepare statement failed: %w", err)
		}
		defer stmt.Close()

		for _, task := range taskIns {
			task.Initial()

			_, err = stmt.Exec(
				task.ID, task.DagInsID, task.TaskID, string(task.Status),
				task.Reason, JSONMarshal(task.Traces), JSONMarshal(task.Results),
				task.LastModifiedAt, JSONMarshal(task.RenderedParams),
				JSONMarshal(task.DependOn), task.Hash, JSONMarshal(task.MetaData),
			)
			if err != nil {
				return fmt.Errorf("insert task instance failed: %w", err)
			}
		}

		return nil
	})
}

// PatchTaskIns 修改task instance
func (s *Store) PatchTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error {
	if taskIns.ID == "" {
		return fmt.Errorf("id cannot be empty")
	}

	return s.WithTransaction(ctx, func(tx interface{}) error {
		txConn := tx.(*sql.Tx)

		query := "UPDATE task_instances SET updated_at = NOW()"
		args := []interface{}{}

		if taskIns.Status != "" {
			query += ", status = ?"
			args = append(args, string(taskIns.Status))
		}

		if taskIns.Reason != "" {
			query += ", reason = ?"
			args = append(args, taskIns.Reason)
		}

		if len(taskIns.Traces) > 0 {
			query += ", traces = ?"
			args = append(args, JSONMarshal(taskIns.Traces))
		}

		if taskIns.Results != nil {
			query += ", results = ?"
			args = append(args, JSONMarshal(taskIns.Results))
		}

		if taskIns.LastModifiedAt != 0 {
			query += ", last_modified_at = ?"
			args = append(args, taskIns.LastModifiedAt)
		}

		if taskIns.RenderedParams != nil {
			query += ", rendered_params = ?"
			args = append(args, JSONMarshal(taskIns.RenderedParams))
		}

		if taskIns.DependOn != nil {
			query += ", depend_on = ?"
			args = append(args, JSONMarshal(taskIns.DependOn))
		}

		if taskIns.Hash != "" {
			query += ", hash = ?"
			args = append(args, taskIns.Hash)
		}

		if taskIns.MetaData != nil {
			query += ", metadata = ?"
			args = append(args, JSONMarshal(taskIns.MetaData))
		}

		query += " WHERE id = ?"
		args = append(args, taskIns.ID)

		_, err := txConn.Exec(query, args...)
		if err != nil {
			return fmt.Errorf("patch task instance failed: %w", err)
		}

		return nil
	})
}

// PatchDagIns 修改dag instance
func (s *Store) PatchDagIns(ctx context.Context, dagIns *entity.DagInstance, mustsPatchFields ...string) error {
	return s.WithTransaction(ctx, func(tx interface{}) error {
		txConn := tx.(*sql.Tx)

		query := "UPDATE dag_instances SET updated_at = NOW()"
		args := []interface{}{}

		if dagIns.EndedAt != 0 {
			query += ", ended_at = ?"
			args = append(args, dagIns.EndedAt)
		}

		if dagIns.EventPersistence == 0 {
			if dagIns.ShareData != nil {
				if dagIns.ShareDataExt != nil {
					query += ", share_data = NULL"
				} else {
					query += ", share_data = ?"
					args = append(args, JSONMarshal(dagIns.ShareData))
				}
			}

			if dagIns.Dump != "" {
				if dagIns.DumpExt != nil {
					query += ", dump = ''"
				} else {
					query += ", dump = ?"
					args = append(args, dagIns.Dump)
				}
			}
		}

		if dagIns.EventPersistence != 0 {
			query += ", event_persistence = ?"
			args = append(args, dagIns.EventPersistence)
		}

		if dagIns.EventOssPath != "" {
			query += ", event_oss_path = ?"
			args = append(args, dagIns.EventOssPath)
		}

		if dagIns.Status != "" {
			query += ", status = ?"
			args = append(args, string(dagIns.Status))
		}

		if utils.StringsContain(mustsPatchFields, "Cmd") || dagIns.Cmd != nil {
			query += ", cmd = ?"
			args = append(args, JSONMarshal(dagIns.Cmd))
		}

		if dagIns.Worker != "" {
			query += ", worker = ?"
			args = append(args, dagIns.Worker)
		}

		if utils.StringsContain(mustsPatchFields, "Reason") || dagIns.Reason != "" {
			query += ", reason = ?"
			args = append(args, dagIns.Reason)
		}

		if dagIns.ResumeData != "" {
			query += ", resume_data = ?"
			args = append(args, dagIns.ResumeData)
		}

		if dagIns.ResumeStatus != "" {
			query += ", resume_status = ?"
			args = append(args, string(dagIns.ResumeStatus))
		}

		if dagIns.Source != "" {
			query += ", source = ?"
			args = append(args, dagIns.Source)
		}

		query += " WHERE id = ?"
		args = append(args, dagIns.ID)

		_, err := txConn.Exec(query, args...)
		if err != nil {
			return fmt.Errorf("patch dag instance failed: %w", err)
		}

		goevent.Publish(&event.DagInstancePatched{
			Payload:         dagIns,
			MustPatchFields: mustsPatchFields,
		})

		return nil
	})
}

// UpdateDag 更新dag
func (s *Store) UpdateDag(ctx context.Context, dag *entity.Dag) error {
	// 检查任务连接
	_, err := mod.BuildRootNode(mod.MapTasksToGetter(dag.Tasks))
	if err != nil {
		return err
	}

	dag.Update()

	return s.WithTransaction(ctx, func(tx interface{}) error {
		txConn := tx.(*sql.Tx)

		_, err := txConn.Exec(
			`UPDATE dags SET
				userid = ?, name = ?, `+"`desc`"+` = ?, trigger = ?, cron = ?, status = ?, tasks = ?, steps = ?, 
				description = ?, shortcuts = ?, accessors = ?, type = ?, policy_type = ?, appinfo = ?, 
				priority = ?, removed = ?, emails = ?, template = ?, published = ?, trigger_config = ?, 
				sub_ids = ?, exec_mode = ?, category = ?, outputs = ?, instructions = ?, operator_id = ?, 
				inc_values = ?, version = ?, version_id = ?, modify_by = ?, is_debug = ?, debug_id = ?, 
				biz_domain_id = ?, updated_at = NOW()
			WHERE id = ?`,
			dag.UserID, dag.Name, dag.Desc, string(dag.Trigger), dag.Cron, string(dag.Status),
			JSONMarshal(dag.Tasks), JSONMarshal(dag.Steps), dag.Description,
			JSONMarshal(dag.Shortcuts), JSONMarshal(dag.Accessors), dag.Type, dag.PolicyType, JSONMarshal(dag.AppInfo),
			dag.Priority, dag.Removed, JSONMarshal(dag.Emails), dag.Template, dag.Published,
			JSONMarshal(dag.TriggerConfig), JSONMarshal(dag.SubIDs), dag.ExecMode, dag.Category,
			JSONMarshal(dag.OutPuts), JSONMarshal(dag.Instructions), dag.OperatorID, JSONMarshal(dag.IncValues),
			dag.Version.String(), dag.VersionID, dag.ModifyBy, dag.IsDebug, dag.DeBugID, dag.BizDomainID,
			dag.ID,
		)
		if err != nil {
			return fmt.Errorf("update dag failed: %w", err)
		}

		return nil
	})
}

// UpdateDagIncValue 更新dag的inc_value
func (s *Store) UpdateDagIncValue(ctx context.Context, dagId string, incKey string, incValue any) error {
	return s.WithTransaction(ctx, func(tx interface{}) error {
		txConn := tx.(*sql.Tx)

		// 首先获取当前的inc_values
		var incValuesStr string
		err := txConn.QueryRow("SELECT inc_values FROM dags WHERE id = ?", dagId).Scan(&incValuesStr)
		if err != nil {
			return fmt.Errorf("get dag inc_values failed: %w", err)
		}

		// 解析inc_values
		incValues := make(map[string]any)
		if incValuesStr != "" && incValuesStr != "null" {
			if err := JSONUnmarshal(incValuesStr, &incValues); err != nil {
				return fmt.Errorf("unmarshal inc_values failed: %w", err)
			}
		}

		// 更新inc_value
		incValues[incKey] = incValue

		// 保存回数据库
		_, err = txConn.Exec(
			"UPDATE dags SET inc_values = ?, updated_at = NOW() WHERE id = ?",
			JSONMarshal(incValues), dagId,
		)
		if err != nil {
			return fmt.Errorf("update dag inc_value failed: %w", err)
		}

		return nil
	})
}

// UpdateDagIns 更新dag instance
func (s *Store) UpdateDagIns(ctx context.Context, dagIns *entity.DagInstance) error {
	dagIns.Update()

	return s.WithTransaction(ctx, func(tx interface{}) error {
		txConn := tx.(*sql.Tx)

		_, err := txConn.Exec(
			`UPDATE dag_instances SET
				dag_id = ?, trigger = ?, worker = ?, source = ?, vars = ?, keywords = ?, 
				event_persistence = ?, event_oss_path = ?, share_data = ?, status = ?, 
				reason = ?, cmd = ?, userid = ?, ended_at = ?, dag_type = ?, policy_type = ?, 
				appinfo = ?, priority = ?, mode = ?, dump = ?, success_callback = ?, 
				error_callback = ?, call_chain = ?, resume_data = ?, resume_status = ?, 
				version = ?, version_id = ?, biz_domain_id = ?, updated_at = NOW()
			WHERE id = ?`,
			dagIns.DagID, string(dagIns.Trigger), dagIns.Worker, dagIns.Source,
			JSONMarshal(dagIns.Vars), JSONMarshal(dagIns.Keywords), dagIns.EventPersistence,
			dagIns.EventOssPath, JSONMarshal(dagIns.ShareData), string(dagIns.Status),
			dagIns.Reason, JSONMarshal(dagIns.Cmd), dagIns.UserID, dagIns.EndedAt,
			dagIns.DagType, dagIns.PolicyType, JSONMarshal(dagIns.AppInfo), dagIns.Priority,
			dagIns.Mode, dagIns.Dump, dagIns.SuccessCallback, dagIns.ErrorCallback,
			JSONMarshal(dagIns.CallChain), dagIns.ResumeData, string(dagIns.ResumeStatus),
			dagIns.Version.String(), dagIns.VersionID, dagIns.BizDomainID,
			dagIns.ID,
		)
		if err != nil {
			return fmt.Errorf("update dag instance failed: %w", err)
		}

		goevent.Publish(&event.DagInstanceUpdated{Payload: dagIns})
		return nil
	})
}

// UpdateTaskIns 更新task instance
func (s *Store) UpdateTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error {
	taskIns.Update()

	return s.WithTransaction(ctx, func(tx interface{}) error {
		txConn := tx.(*sql.Tx)

		_, err := txConn.Exec(
			`UPDATE task_instances SET
				dag_ins_id = ?, task_id = ?, status = ?, reason = ?, traces = ?, results = ?, 
				last_modified_at = ?, rendered_params = ?, depend_on = ?, hash = ?, 
				metadata = ?, updated_at = NOW()
			WHERE id = ?`,
			taskIns.DagInsID, taskIns.TaskID, string(taskIns.Status), taskIns.Reason,
			JSONMarshal(taskIns.Traces), JSONMarshal(taskIns.Results), taskIns.LastModifiedAt,
			JSONMarshal(taskIns.RenderedParams), JSONMarshal(taskIns.DependOn), taskIns.Hash,
			JSONMarshal(taskIns.MetaData), taskIns.ID,
		)
		if err != nil {
			return fmt.Errorf("update task instance failed: %w", err)
		}

		return nil
	})
}

// BatchUpdateDagIns 批量更新dag instance
func (s *Store) BatchUpdateDagIns(ctx context.Context, dagIns []*entity.DagInstance) error {
	if len(dagIns) == 0 {
		return nil
	}

	return s.WithTransaction(ctx, func(tx interface{}) error {
		txConn := tx.(*sql.Tx)

		stmt, err := txConn.Prepare(
			`UPDATE dag_instances SET
				dag_id = ?, trigger = ?, worker = ?, source = ?, vars = ?, keywords = ?, 
				event_persistence = ?, event_oss_path = ?, share_data = ?, status = ?, 
				reason = ?, cmd = ?, userid = ?, ended_at = ?, dag_type = ?, policy_type = ?, 
				appinfo = ?, priority = ?, mode = ?, dump = ?, success_callback = ?, 
				error_callback = ?, call_chain = ?, resume_data = ?, resume_status = ?, 
				version = ?, version_id = ?, biz_domain_id = ?, updated_at = NOW()
			WHERE id = ?`,
		)
		if err != nil {
			return fmt.Errorf("prepare statement failed: %w", err)
		}
		defer stmt.Close()

		for _, ins := range dagIns {
			ins.Update()

			_, err = stmt.Exec(
				ins.DagID, string(ins.Trigger), ins.Worker, ins.Source,
				JSONMarshal(ins.Vars), JSONMarshal(ins.Keywords), ins.EventPersistence,
				ins.EventOssPath, JSONMarshal(ins.ShareData), string(ins.Status),
				ins.Reason, JSONMarshal(ins.Cmd), ins.UserID, ins.EndedAt,
				ins.DagType, ins.PolicyType, JSONMarshal(ins.AppInfo), ins.Priority,
				ins.Mode, ins.Dump, ins.SuccessCallback, ins.ErrorCallback,
				JSONMarshal(ins.CallChain), ins.ResumeData, string(ins.ResumeStatus),
				ins.Version.String(), ins.VersionID, ins.BizDomainID,
				ins.ID,
			)
			if err != nil {
				return fmt.Errorf("update dag instance failed: %w", err)
			}
		}

		return nil
	})
}

// BatchUpdateTaskIns 批量更新task instance
func (s *Store) BatchUpdateTaskIns(taskIns []*entity.TaskInstance) error {
	if len(taskIns) == 0 {
		return nil
	}

	return s.WithTransaction(context.Background(), func(tx interface{}) error {
		txConn := tx.(*sql.Tx)

		stmt, err := txConn.Prepare(
			`UPDATE task_instances SET
				dag_ins_id = ?, task_id = ?, status = ?, reason = ?, traces = ?, results = ?, 
				last_modified_at = ?, rendered_params = ?, depend_on = ?, hash = ?, 
				metadata = ?, updated_at = NOW()
			WHERE id = ?`,
		)
		if err != nil {
			return fmt.Errorf("prepare statement failed: %w", err)
		}
		defer stmt.Close()

		for _, task := range taskIns {
			task.Update()

			_, err = stmt.Exec(
				task.DagInsID, task.TaskID, string(task.Status), task.Reason,
				JSONMarshal(task.Traces), JSONMarshal(task.Results), task.LastModifiedAt,
				JSONMarshal(task.RenderedParams), JSONMarshal(task.DependOn), task.Hash,
				JSONMarshal(task.MetaData), task.ID,
			)
			if err != nil {
				return fmt.Errorf("update task instance failed: %w", err)
			}
		}

		return nil
	})
}

// BatchDeleteTaskIns 批量删除task instance
func (s *Store) BatchDeleteTaskIns(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	return s.WithTransaction(ctx, func(tx interface{}) error {
		txConn := tx.(*sql.Tx)

		// 构建IN子句
		placeholders := make([]string, len(ids))
		args := make([]interface{}, len(ids))
		for i, id := range ids {
			placeholders[i] = "?"
			args[i] = id
		}

		query := fmt.Sprintf("DELETE FROM task_instances WHERE id IN (%s)", strings.Join(placeholders, ","))
		_, err := txConn.Exec(query, args...)
		if err != nil {
			return fmt.Errorf("delete task instances failed: %w", err)
		}

		return nil
	})
}

// GetTaskIns 获取task instance
func (s *Store) GetTaskIns(ctx context.Context, taskInsID string) (*entity.TaskInstance, error) {
	taskIns := &entity.TaskInstance{}

	db := s.GetDB(ctx)
	var row *sql.Row

	if tx, ok := db.(*sql.Tx); ok {
		row = tx.QueryRow(
			`SELECT id, dag_ins_id, task_id, status, reason, traces, results, last_modified_at, 
				rendered_params, depend_on, hash, metadata, created_at, updated_at 
			FROM task_instances WHERE id = ?`,
			taskInsID,
		)
	} else if dbConn, ok := db.(*sql.DB); ok {
		row = dbConn.QueryRow(
			`SELECT id, dag_ins_id, task_id, status, reason, traces, results, last_modified_at, 
				rendered_params, depend_on, hash, metadata, created_at, updated_at 
			FROM task_instances WHERE id = ?`,
			taskInsID,
		)
	}

	var status, tracesStr, resultsStr, renderedParamsStr, dependOnStr, metadataStr string
	err := row.Scan(
		&taskIns.ID, &taskIns.DagInsID, &taskIns.TaskID, &status, &taskIns.Reason,
		&tracesStr, &resultsStr, &taskIns.LastModifiedAt, &renderedParamsStr,
		&dependOnStr, &taskIns.Hash, &metadataStr, &taskIns.CreatedAt, &taskIns.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrDataNotFound
		}
		return nil, fmt.Errorf("get task instance failed: %w", err)
	}

	// 解析JSON字段
	taskIns.Status = entity.TaskInstanceStatus(status)
	JSONUnmarshal(tracesStr, &taskIns.Traces)
	JSONUnmarshal(resultsStr, &taskIns.Results)
	JSONUnmarshal(renderedParamsStr, &taskIns.RenderedParams)
	JSONUnmarshal(dependOnStr, &taskIns.DependOn)
	JSONUnmarshal(metadataStr, &taskIns.MetaData)

	return taskIns, nil
}

// GetDag 获取dag
func (s *Store) GetDag(ctx context.Context, dagID string) (*entity.Dag, error) {
	dag := &entity.Dag{}

	db := s.GetDB(ctx)
	var row *sql.Row

	if tx, ok := db.(*sql.Tx); ok {
		row = tx.QueryRow(
			`SELECT id, userid, name, `+"`desc`"+`, trigger, cron, status, tasks, steps, description, 
				shortcuts, accessors, type, policy_type, appinfo, priority, removed, emails, 
				template, published, trigger_config, sub_ids, exec_mode, category, outputs, 
				instructions, operator_id, inc_values, version, version_id, modify_by, 
				is_debug, debug_id, biz_domain_id, created_at, updated_at 
			FROM dags WHERE id = ? AND removed = false`,
			dagID,
		)
	} else if dbConn, ok := db.(*sql.DB); ok {
		row = dbConn.QueryRow(
			`SELECT id, userid, name, `+"`desc`"+`, trigger, cron, status, tasks, steps, description, 
				shortcuts, accessors, type, policy_type, appinfo, priority, removed, emails, 
				template, published, trigger_config, sub_ids, exec_mode, category, outputs, 
				instructions, operator_id, inc_values, version, version_id, modify_by, 
				is_debug, debug_id, biz_domain_id, created_at, updated_at 
			FROM dags WHERE id = ? AND removed = false`,
			dagID,
		)
	}

	var trigger, status, typeStr, policyType, template, execMode, category, operatorID, version, versionID, modifyBy, debugID, bizDomainID string
	var removed, published, isDebug bool
	var tasksStr, stepsStr, description, shortcutsStr, accessorsStr, appinfoStr, priority, emailsStr, triggerConfigStr, subIDsStr, outputsStr, instructionsStr, incValuesStr string
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&dag.ID, &dag.UserID, &dag.Name, &dag.Desc, &trigger, &dag.Cron, &status,
		&tasksStr, &stepsStr, &description, &shortcutsStr, &accessorsStr, &typeStr,
		&policyType, &appinfoStr, &priority, &removed, &emailsStr, &template, &published,
		&triggerConfigStr, &subIDsStr, &execMode, &category, &outputsStr, &instructionsStr,
		&operatorID, &incValuesStr, &version, &versionID, &modifyBy, &isDebug, &debugID,
		&bizDomainID, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrDataNotFound
		}
		return nil, fmt.Errorf("get dag failed: %w", err)
	}

	// 解析字段
	dag.Trigger = entity.DagTrigger(trigger)
	dag.Status = entity.DagStatus(status)
	dag.Type = typeStr
	dag.PolicyType = policyType
	dag.Template = template
	dag.ExecMode = execMode
	dag.Category = category
	dag.OperatorID = operatorID
	dag.Version = entity.NewVersion(version)
	dag.VersionID = versionID
	dag.ModifyBy = modifyBy
	dag.Removed = removed
	dag.Published = published
	dag.IsDebug = isDebug
	dag.DeBugID = debugID
	dag.BizDomainID = bizDomainID
	dag.CreatedAt = createdAt
	dag.UpdatedAt = updatedAt
	dag.Description = description
	dag.Priority = priority

	// 解析JSON字段
	JSONUnmarshal(tasksStr, &dag.Tasks)
	JSONUnmarshal(stepsStr, &dag.Steps)
	JSONUnmarshal(shortcutsStr, &dag.Shortcuts)
	JSONUnmarshal(accessorsStr, &dag.Accessors)
	JSONUnmarshal(appinfoStr, &dag.AppInfo)
	JSONUnmarshal(emailsStr, &dag.Emails)
	JSONUnmarshal(triggerConfigStr, &dag.TriggerConfig)
	JSONUnmarshal(subIDsStr, &dag.SubIDs)
	JSONUnmarshal(outputsStr, &dag.OutPuts)
	JSONUnmarshal(instructionsStr, &dag.Instructions)
	JSONUnmarshal(incValuesStr, &dag.IncValues)

	return dag, nil
}

// GetDagByFields 根据字段获取dag
func (s *Store) GetDagByFields(ctx context.Context, params map[string]interface{}) (*entity.Dag, error) {
	query := "SELECT id, userid, name, `desc`, trigger, cron, status, tasks, steps, description, shortcuts, accessors, type, policy_type, appinfo, priority, removed, emails, template, published, trigger_config, sub_ids, exec_mode, category, outputs, instructions, operator_id, inc_values, version, version_id, modify_by, is_debug, debug_id, biz_domain_id, created_at, updated_at FROM dags WHERE removed = false"
	args := []interface{}{}

	for key, value := range params {
		switch key {
		case "id":
			query += " AND id = ?"
			args = append(args, value)
		case "userid":
			query += " AND userid = ?"
			args = append(args, value)
		case "name":
			query += " AND name = ?"
			args = append(args, value)
		case "status":
			query += " AND status = ?"
			args = append(args, value)
		case "type":
			query += " AND type = ?"
			args = append(args, value)
		case "trigger":
			query += " AND trigger = ?"
			args = append(args, value)
		}
	}

	row := s.db.QueryRow(query, args...)

	dag := &entity.Dag{}
	var trigger, status, typeStr, policyType, template, execMode, category, operatorID, version, versionID, modifyBy, debugID, bizDomainID string
	var removed, published, isDebug bool
	var tasksStr, stepsStr, description, shortcutsStr, accessorsStr, appinfoStr, priority, emailsStr, triggerConfigStr, subIDsStr, outputsStr, instructionsStr, incValuesStr string
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&dag.ID, &dag.UserID, &dag.Name, &dag.Desc, &trigger, &dag.Cron, &status,
		&tasksStr, &stepsStr, &description, &shortcutsStr, &accessorsStr, &typeStr,
		&policyType, &appinfoStr, &priority, &removed, &emailsStr, &template, &published,
		&triggerConfigStr, &subIDsStr, &execMode, &category, &outputsStr, &instructionsStr,
		&operatorID, &incValuesStr, &version, &versionID, &modifyBy, &isDebug, &debugID,
		&bizDomainID, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrDataNotFound
		}
		return nil, fmt.Errorf("get dag by fields failed: %w", err)
	}

	dag.Trigger = entity.DagTrigger(trigger)
	dag.Status = entity.DagStatus(status)
	dag.Type = typeStr
	dag.PolicyType = policyType
	dag.Template = template
	dag.ExecMode = execMode
	dag.Category = category
	dag.OperatorID = operatorID
	dag.Version = entity.NewVersion(version)
	dag.VersionID = versionID
	dag.ModifyBy = modifyBy
	dag.Removed = removed
	dag.Published = published
	dag.IsDebug = isDebug
	dag.DeBugID = debugID
	dag.BizDomainID = bizDomainID
	dag.CreatedAt = createdAt
	dag.UpdatedAt = updatedAt
	dag.Description = description
	dag.Priority = priority

	JSONUnmarshal(tasksStr, &dag.Tasks)
	JSONUnmarshal(stepsStr, &dag.Steps)
	JSONUnmarshal(shortcutsStr, &dag.Shortcuts)
	JSONUnmarshal(accessorsStr, &dag.Accessors)
	JSONUnmarshal(appinfoStr, &dag.AppInfo)
	JSONUnmarshal(emailsStr, &dag.Emails)
	JSONUnmarshal(triggerConfigStr, &dag.TriggerConfig)
	JSONUnmarshal(subIDsStr, &dag.SubIDs)
	JSONUnmarshal(outputsStr, &dag.OutPuts)
	JSONUnmarshal(instructionsStr, &dag.Instructions)
	JSONUnmarshal(incValuesStr, &dag.IncValues)

	return dag, nil
}

// GetDagWithOptionalVersion 根据版本获取dag
func (s *Store) GetDagWithOptionalVersion(ctx context.Context, dagID, versionID string) (*entity.Dag, error) {
	if versionID == "" {
		return s.GetDag(ctx, dagID)
	}

	var dagVersion entity.DagVersion
	err := s.db.QueryRow(
		"SELECT id, dag_id, version_id, config, sort_time, created_at, updated_at FROM dag_versions WHERE dag_id = ? AND version_id = ?",
		dagID, versionID,
	).Scan(&dagVersion.ID, &dagVersion.DagID, &dagVersion.VersionID, &dagVersion.Config, &dagVersion.SortTime, &dagVersion.CreatedAt, &dagVersion.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.GetDag(ctx, dagID)
		}
		return nil, fmt.Errorf("get dag version failed: %w", err)
	}

	dag := &entity.Dag{}
	if err := json.Unmarshal([]byte(dagVersion.Config.(string)), dag); err != nil {
		return nil, fmt.Errorf("unmarshal dag config failed: %w", err)
	}

	return dag, nil
}

// GetDagInstance 获取dag instance
func (s *Store) GetDagInstance(ctx context.Context, dagInsId string) (*entity.DagInstance, error) {
	dagIns := &entity.DagInstance{}

	row := s.db.QueryRow(
		`SELECT id, dag_id, trigger, worker, source, vars, keywords, event_persistence, 
			event_oss_path, share_data, status, reason, cmd, userid, ended_at, 
			dag_type, policy_type, appinfo, priority, mode, dump, success_callback, 
			error_callback, call_chain, resume_data, resume_status, version, version_id, 
			biz_domain_id, created_at, updated_at 
		FROM dag_instances WHERE id = ?`,
		dagInsId,
	)

	var trigger, status, dagType, policyType, priority, resumeStatus, version, versionID, bizDomainID string
	var varsStr, keywordsStr, shareDataStr, cmdStr, appinfoStr, callChainStr string
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&dagIns.ID, &dagIns.DagID, &trigger, &dagIns.Worker, &dagIns.Source,
		&varsStr, &keywordsStr, &dagIns.EventPersistence, &dagIns.EventOssPath,
		&shareDataStr, &status, &dagIns.Reason, &cmdStr, &dagIns.UserID, &dagIns.EndedAt,
		&dagType, &policyType, &appinfoStr, &priority, &dagIns.Mode, &dagIns.Dump,
		&dagIns.SuccessCallback, &dagIns.ErrorCallback, &callChainStr, &dagIns.ResumeData,
		&resumeStatus, &version, &versionID, &bizDomainID, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrDataNotFound
		}
		return nil, fmt.Errorf("get dag instance failed: %w", err)
	}

	dagIns.Trigger = entity.DagTrigger(trigger)
	dagIns.Status = entity.DagInstanceStatus(status)
	dagIns.DagType = dagType
	dagIns.PolicyType = policyType
	dagIns.Priority = priority
	dagIns.ResumeStatus = entity.DagInstanceStatus(resumeStatus)
	dagIns.Version = entity.NewVersion(version)
	dagIns.VersionID = versionID
	dagIns.BizDomainID = bizDomainID
	dagIns.CreatedAt = createdAt
	dagIns.UpdatedAt = updatedAt

	JSONUnmarshal(varsStr, &dagIns.Vars)
	JSONUnmarshal(keywordsStr, &dagIns.Keywords)
	JSONUnmarshal(shareDataStr, &dagIns.ShareData)
	JSONUnmarshal(cmdStr, &dagIns.Cmd)
	JSONUnmarshal(appinfoStr, &dagIns.AppInfo)
	JSONUnmarshal(callChainStr, &dagIns.CallChain)

	return dagIns, nil
}

// GetDagInstanceByFields 根据字段获取dag instance
func (s *Store) GetDagInstanceByFields(ctx context.Context, params map[string]interface{}) (*entity.DagInstance, error) {
	query := `SELECT id, dag_id, trigger, worker, source, vars, keywords, event_persistence, 
		event_oss_path, share_data, status, reason, cmd, userid, ended_at, 
		dag_type, policy_type, appinfo, priority, mode, dump, success_callback, 
		error_callback, call_chain, resume_data, resume_status, version, version_id, 
		biz_domain_id, created_at, updated_at FROM dag_instances WHERE 1=1`
	args := []interface{}{}

	for key, value := range params {
		switch key {
		case "id":
			query += " AND id = ?"
			args = append(args, value)
		case "dag_id":
			query += " AND dag_id = ?"
			args = append(args, value)
		case "userid":
			query += " AND userid = ?"
			args = append(args, value)
		case "status":
			query += " AND status = ?"
			args = append(args, value)
		}
	}

	row := s.db.QueryRow(query, args...)

	dagIns := &entity.DagInstance{}
	var trigger, status, dagType, policyType, priority, resumeStatus, version, versionID, bizDomainID string
	var varsStr, keywordsStr, shareDataStr, cmdStr, appinfoStr, callChainStr string
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&dagIns.ID, &dagIns.DagID, &trigger, &dagIns.Worker, &dagIns.Source,
		&varsStr, &keywordsStr, &dagIns.EventPersistence, &dagIns.EventOssPath,
		&shareDataStr, &status, &dagIns.Reason, &cmdStr, &dagIns.UserID, &dagIns.EndedAt,
		&dagType, &policyType, &appinfoStr, &priority, &dagIns.Mode, &dagIns.Dump,
		&dagIns.SuccessCallback, &dagIns.ErrorCallback, &callChainStr, &dagIns.ResumeData,
		&resumeStatus, &version, &versionID, &bizDomainID, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrDataNotFound
		}
		return nil, fmt.Errorf("get dag instance by fields failed: %w", err)
	}

	dagIns.Trigger = entity.DagTrigger(trigger)
	dagIns.Status = entity.DagInstanceStatus(status)
	dagIns.DagType = dagType
	dagIns.PolicyType = policyType
	dagIns.Priority = priority
	dagIns.ResumeStatus = entity.DagInstanceStatus(resumeStatus)
	dagIns.Version = entity.NewVersion(version)
	dagIns.VersionID = versionID
	dagIns.BizDomainID = bizDomainID
	dagIns.CreatedAt = createdAt
	dagIns.UpdatedAt = updatedAt

	JSONUnmarshal(varsStr, &dagIns.Vars)
	JSONUnmarshal(keywordsStr, &dagIns.Keywords)
	JSONUnmarshal(shareDataStr, &dagIns.ShareData)
	JSONUnmarshal(cmdStr, &dagIns.Cmd)
	JSONUnmarshal(appinfoStr, &dagIns.AppInfo)
	JSONUnmarshal(callChainStr, &dagIns.CallChain)

	return dagIns, nil
}

// ListDag 列出dag
func (s *Store) ListDag(ctx context.Context, input *mod.ListDagInput) ([]*entity.Dag, error) {
	query := "SELECT id, userid, name, `desc`, trigger, cron, status, tasks, steps, description, shortcuts, accessors, type, policy_type, appinfo, priority, removed, emails, template, published, trigger_config, sub_ids, exec_mode, category, outputs, instructions, operator_id, inc_values, version, version_id, modify_by, is_debug, debug_id, biz_domain_id, created_at, updated_at FROM dags WHERE removed = false"
	args := []interface{}{}

	if input.UserID != "" {
		query += " AND userid = ?"
		args = append(args, input.UserID)
	}

	if len(input.DagIDs) > 0 {
		placeholders := make([]string, len(input.DagIDs))
		for i, id := range input.DagIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		query += fmt.Sprintf(" AND id IN (%s)", strings.Join(placeholders, ","))
	}

	if len(input.Trigger) > 0 {
		placeholders := make([]string, len(input.Trigger))
		for i, t := range input.Trigger {
			placeholders[i] = "?"
			args = append(args, t)
		}
		query += fmt.Sprintf(" AND trigger IN (%s)", strings.Join(placeholders, ","))
	}

	if len(input.Status) > 0 {
		placeholders := make([]string, len(input.Status))
		for i, s := range input.Status {
			placeholders[i] = "?"
			args = append(args, string(s))
		}
		query += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ","))
	}

	if input.Type != "" {
		query += " AND type = ?"
		args = append(args, input.Type)
	}

	if input.BizDomainID != "" {
		query += " AND biz_domain_id = ?"
		args = append(args, input.BizDomainID)
	}

	if input.KeyWord != "" {
		query += " AND (name LIKE ? OR description LIKE ?)"
		args = append(args, "%"+input.KeyWord+"%", "%"+input.KeyWord+"%")
	}

	if input.SortBy != "" {
		order := "ASC"
		if input.Order == -1 {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", input.SortBy, order)
	} else {
		query += " ORDER BY created_at DESC"
	}

	if input.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, input.Limit)
	}

	if input.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, input.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list dag failed: %w", err)
	}
	defer rows.Close()

	var dags []*entity.Dag
	for rows.Next() {
		dag := &entity.Dag{}
		var trigger, status, typeStr, policyType, template, execMode, category, operatorID, version, versionID, modifyBy, debugID, bizDomainID string
		var removed, published, isDebug bool
		var tasksStr, stepsStr, description, shortcutsStr, accessorsStr, appinfoStr, priority, emailsStr, triggerConfigStr, subIDsStr, outputsStr, instructionsStr, incValuesStr string
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&dag.ID, &dag.UserID, &dag.Name, &dag.Desc, &trigger, &dag.Cron, &status,
			&tasksStr, &stepsStr, &description, &shortcutsStr, &accessorsStr, &typeStr,
			&policyType, &appinfoStr, &priority, &removed, &emailsStr, &template, &published,
			&triggerConfigStr, &subIDsStr, &execMode, &category, &outputsStr, &instructionsStr,
			&operatorID, &incValuesStr, &version, &versionID, &modifyBy, &isDebug, &debugID,
			&bizDomainID, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan dag failed: %w", err)
		}

		dag.Trigger = entity.DagTrigger(trigger)
		dag.Status = entity.DagStatus(status)
		dag.Type = typeStr
		dag.PolicyType = policyType
		dag.Template = template
		dag.ExecMode = execMode
		dag.Category = category
		dag.OperatorID = operatorID
		dag.Version = entity.NewVersion(version)
		dag.VersionID = versionID
		dag.ModifyBy = modifyBy
		dag.Removed = removed
		dag.Published = published
		dag.IsDebug = isDebug
		dag.DeBugID = debugID
		dag.BizDomainID = bizDomainID
		dag.CreatedAt = createdAt
		dag.UpdatedAt = updatedAt
		dag.Description = description
		dag.Priority = priority

		JSONUnmarshal(tasksStr, &dag.Tasks)
		JSONUnmarshal(stepsStr, &dag.Steps)
		JSONUnmarshal(shortcutsStr, &dag.Shortcuts)
		JSONUnmarshal(accessorsStr, &dag.Accessors)
		JSONUnmarshal(appinfoStr, &dag.AppInfo)
		JSONUnmarshal(emailsStr, &dag.Emails)
		JSONUnmarshal(triggerConfigStr, &dag.TriggerConfig)
		JSONUnmarshal(subIDsStr, &dag.SubIDs)
		JSONUnmarshal(outputsStr, &dag.OutPuts)
		JSONUnmarshal(instructionsStr, &dag.Instructions)
		JSONUnmarshal(incValuesStr, &dag.IncValues)

		dags = append(dags, dag)
	}

	return dags, nil
}

// ListDagByFields 根据字段列出dag
func (s *Store) ListDagByFields(ctx context.Context, filter bson.M, opt options.FindOptions) ([]*entity.Dag, error) {
	query := "SELECT id, userid, name, `desc`, trigger, cron, status, tasks, steps, description, shortcuts, accessors, type, policy_type, appinfo, priority, removed, emails, template, published, trigger_config, sub_ids, exec_mode, category, outputs, instructions, operator_id, inc_values, version, version_id, modify_by, is_debug, debug_id, biz_domain_id, created_at, updated_at FROM dags WHERE removed = false"
	args := []interface{}{}

	for key, value := range filter {
		switch key {
		case "userid":
			query += " AND userid = ?"
			args = append(args, value)
		case "type":
			query += " AND type = ?"
			args = append(args, value)
		case "trigger":
			query += " AND trigger = ?"
			args = append(args, value)
		case "status":
			query += " AND status = ?"
			args = append(args, value)
		}
	}

	if opt.Sort != nil {
		if sortMap, ok := opt.Sort.(bson.M); ok {
			for field, order := range sortMap {
				ord := "ASC"
				if ordInt, ok := order.(int); ok && ordInt == -1 {
					ord = "DESC"
				}
				query += fmt.Sprintf(" ORDER BY %s %s", field, ord)
				break
			}
		}
	}

	if opt.Limit != nil {
		query += " LIMIT ?"
		args = append(args, *opt.Limit)
	}

	if opt.Skip != nil {
		query += " OFFSET ?"
		args = append(args, *opt.Skip)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list dag by fields failed: %w", err)
	}
	defer rows.Close()

	var dags []*entity.Dag
	for rows.Next() {
		dag := &entity.Dag{}
		var trigger, status, typeStr, policyType, template, execMode, category, operatorID, version, versionID, modifyBy, debugID, bizDomainID string
		var removed, published, isDebug bool
		var tasksStr, stepsStr, description, shortcutsStr, accessorsStr, appinfoStr, priority, emailsStr, triggerConfigStr, subIDsStr, outputsStr, instructionsStr, incValuesStr string
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&dag.ID, &dag.UserID, &dag.Name, &dag.Desc, &trigger, &dag.Cron, &status,
			&tasksStr, &stepsStr, &description, &shortcutsStr, &accessorsStr, &typeStr,
			&policyType, &appinfoStr, &priority, &removed, &emailsStr, &template, &published,
			&triggerConfigStr, &subIDsStr, &execMode, &category, &outputsStr, &instructionsStr,
			&operatorID, &incValuesStr, &version, &versionID, &modifyBy, &isDebug, &debugID,
			&bizDomainID, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan dag failed: %w", err)
		}

		dag.Trigger = entity.DagTrigger(trigger)
		dag.Status = entity.DagStatus(status)
		dag.Type = typeStr
		dag.PolicyType = policyType
		dag.Template = template
		dag.ExecMode = execMode
		dag.Category = category
		dag.OperatorID = operatorID
		dag.Version = entity.NewVersion(version)
		dag.VersionID = versionID
		dag.ModifyBy = modifyBy
		dag.Removed = removed
		dag.Published = published
		dag.IsDebug = isDebug
		dag.DeBugID = debugID
		dag.BizDomainID = bizDomainID
		dag.CreatedAt = createdAt
		dag.UpdatedAt = updatedAt
		dag.Description = description
		dag.Priority = priority

		JSONUnmarshal(tasksStr, &dag.Tasks)
		JSONUnmarshal(stepsStr, &dag.Steps)
		JSONUnmarshal(shortcutsStr, &dag.Shortcuts)
		JSONUnmarshal(accessorsStr, &dag.Accessors)
		JSONUnmarshal(appinfoStr, &dag.AppInfo)
		JSONUnmarshal(emailsStr, &dag.Emails)
		JSONUnmarshal(triggerConfigStr, &dag.TriggerConfig)
		JSONUnmarshal(subIDsStr, &dag.SubIDs)
		JSONUnmarshal(outputsStr, &dag.OutPuts)
		JSONUnmarshal(instructionsStr, &dag.Instructions)
		JSONUnmarshal(incValuesStr, &dag.IncValues)

		dags = append(dags, dag)
	}

	return dags, nil
}

// ListDagInstance 列出dag instance
func (s *Store) ListDagInstance(ctx context.Context, input *mod.ListDagInstanceInput) ([]*entity.DagInstance, error) {
	query := `SELECT id, dag_id, trigger, worker, source, vars, keywords, event_persistence, 
		event_oss_path, share_data, status, reason, cmd, userid, ended_at, 
		dag_type, policy_type, appinfo, priority, mode, dump, success_callback, 
		error_callback, call_chain, resume_data, resume_status, version, version_id, 
		biz_domain_id, created_at, updated_at FROM dag_instances WHERE 1=1`
	args := []interface{}{}

	if input.Worker != "" {
		query += " AND worker = ?"
		args = append(args, input.Worker)
	}

	if len(input.DagIDs) > 0 {
		placeholders := make([]string, len(input.DagIDs))
		for i, id := range input.DagIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		query += fmt.Sprintf(" AND dag_id IN (%s)", strings.Join(placeholders, ","))
	}

	if len(input.Status) > 0 {
		placeholders := make([]string, len(input.Status))
		for i, s := range input.Status {
			placeholders[i] = "?"
			args = append(args, string(s))
		}
		query += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ","))
	}

	if len(input.UserIDs) > 0 {
		placeholders := make([]string, len(input.UserIDs))
		for i, id := range input.UserIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		query += fmt.Sprintf(" AND userid IN (%s)", strings.Join(placeholders, ","))
	}

	if input.TimeRange != nil {
		query += fmt.Sprintf(" AND %s >= ? AND %s <= ?", input.TimeRange.Field, input.TimeRange.Field)
		args = append(args, input.TimeRange.Begin, input.TimeRange.End)
	}

	if input.HasCmd {
		query += " AND cmd IS NOT NULL"
	}

	if input.SortBy != "" {
		order := "ASC"
		if input.Order == -1 {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", input.SortBy, order)
	} else {
		query += " ORDER BY created_at DESC"
	}

	if input.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, input.Limit)
	}

	if input.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, input.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list dag instance failed: %w", err)
	}
	defer rows.Close()

	var dagInstances []*entity.DagInstance
	for rows.Next() {
		dagIns := &entity.DagInstance{}
		var trigger, status, dagType, policyType, priority, resumeStatus, version, versionID, bizDomainID string
		var varsStr, keywordsStr, shareDataStr, cmdStr, appinfoStr, callChainStr string
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&dagIns.ID, &dagIns.DagID, &trigger, &dagIns.Worker, &dagIns.Source,
			&varsStr, &keywordsStr, &dagIns.EventPersistence, &dagIns.EventOssPath,
			&shareDataStr, &status, &dagIns.Reason, &cmdStr, &dagIns.UserID, &dagIns.EndedAt,
			&dagType, &policyType, &appinfoStr, &priority, &dagIns.Mode, &dagIns.Dump,
			&dagIns.SuccessCallback, &dagIns.ErrorCallback, &callChainStr, &dagIns.ResumeData,
			&resumeStatus, &version, &versionID, &bizDomainID, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan dag instance failed: %w", err)
		}

		dagIns.Trigger = entity.DagTrigger(trigger)
		dagIns.Status = entity.DagInstanceStatus(status)
		dagIns.DagType = dagType
		dagIns.PolicyType = policyType
		dagIns.Priority = priority
		dagIns.ResumeStatus = entity.DagInstanceStatus(resumeStatus)
		dagIns.Version = entity.NewVersion(version)
		dagIns.VersionID = versionID
		dagIns.BizDomainID = bizDomainID
		dagIns.CreatedAt = createdAt
		dagIns.UpdatedAt = updatedAt

		JSONUnmarshal(varsStr, &dagIns.Vars)
		JSONUnmarshal(keywordsStr, &dagIns.Keywords)
		JSONUnmarshal(shareDataStr, &dagIns.ShareData)
		JSONUnmarshal(cmdStr, &dagIns.Cmd)
		JSONUnmarshal(appinfoStr, &dagIns.AppInfo)
		JSONUnmarshal(callChainStr, &dagIns.CallChain)

		dagInstances = append(dagInstances, dagIns)
	}

	return dagInstances, nil
}

// DisdinctDagInstance 获取dag instance的去重值
func (s *Store) DisdinctDagInstance(input *mod.ListDagInstanceInput) ([]interface{}, error) {
	if input.DistinctField == "" {
		return nil, fmt.Errorf("distinct field is required")
	}

	query := fmt.Sprintf("SELECT DISTINCT %s FROM dag_instances WHERE 1=1", input.DistinctField)
	args := []interface{}{}

	if input.Worker != "" {
		query += " AND worker = ?"
		args = append(args, input.Worker)
	}

	if len(input.Status) > 0 {
		placeholders := make([]string, len(input.Status))
		for i, s := range input.Status {
			placeholders[i] = "?"
			args = append(args, string(s))
		}
		query += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ","))
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("distinct dag instance failed: %w", err)
	}
	defer rows.Close()

	var results []interface{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, fmt.Errorf("scan distinct value failed: %w", err)
		}
		results = append(results, value)
	}

	return results, nil
}

// ListTaskInstance 列出task instance
func (s *Store) ListTaskInstance(ctx context.Context, input *mod.ListTaskInstanceInput) ([]*entity.TaskInstance, error) {
	query := `SELECT id, dag_ins_id, task_id, status, reason, traces, results, last_modified_at, 
		rendered_params, depend_on, hash, metadata, created_at, updated_at FROM task_instances WHERE 1=1`
	args := []interface{}{}

	if input.DagInsID != "" {
		query += " AND dag_ins_id = ?"
		args = append(args, input.DagInsID)
	}

	if len(input.IDs) > 0 {
		placeholders := make([]string, len(input.IDs))
		for i, id := range input.IDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		query += fmt.Sprintf(" AND id IN (%s)", strings.Join(placeholders, ","))
	}

	if len(input.Status) > 0 {
		placeholders := make([]string, len(input.Status))
		for i, s := range input.Status {
			placeholders[i] = "?"
			args = append(args, string(s))
		}
		query += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ","))
	}

	if input.Hash != "" {
		query += " AND hash = ?"
		args = append(args, input.Hash)
	}

	if input.SortBy != "" {
		order := "ASC"
		if input.Order == -1 {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", input.SortBy, order)
	} else {
		query += " ORDER BY created_at DESC"
	}

	if input.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, input.Limit)
	}

	if input.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, input.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list task instance failed: %w", err)
	}
	defer rows.Close()

	var taskInstances []*entity.TaskInstance
	for rows.Next() {
		taskIns := &entity.TaskInstance{}
		var status, tracesStr, resultsStr, renderedParamsStr, dependOnStr, metadataStr string
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&taskIns.ID, &taskIns.DagInsID, &taskIns.TaskID, &status, &taskIns.Reason,
			&tracesStr, &resultsStr, &taskIns.LastModifiedAt, &renderedParamsStr,
			&dependOnStr, &taskIns.Hash, &metadataStr, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan task instance failed: %w", err)
		}

		taskIns.Status = entity.TaskInstanceStatus(status)
		taskIns.CreatedAt = createdAt
		taskIns.UpdatedAt = updatedAt

		JSONUnmarshal(tracesStr, &taskIns.Traces)
		JSONUnmarshal(resultsStr, &taskIns.Results)
		JSONUnmarshal(renderedParamsStr, &taskIns.RenderedParams)
		JSONUnmarshal(dependOnStr, &taskIns.DependOn)
		JSONUnmarshal(metadataStr, &taskIns.MetaData)

		taskInstances = append(taskInstances, taskIns)
	}

	return taskInstances, nil
}

// Marshal 序列化
func (s *Store) Marshal(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

// Unmarshal 反序列化
func (s *Store) Unmarshal(bytes []byte, ptr interface{}) error {
	return json.Unmarshal(bytes, ptr)
}

// BatchDeleteDagWithTransaction 使用事务批量删除dag
func (s *Store) BatchDeleteDagWithTransaction(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	return s.WithTransaction(ctx, func(tx interface{}) error {
		txConn := tx.(*sql.Tx)

		placeholders := make([]string, len(ids))
		args := make([]interface{}, len(ids))
		for i, id := range ids {
			placeholders[i] = "?"
			args[i] = id
		}

		query := fmt.Sprintf("UPDATE dags SET removed = true, updated_at = NOW() WHERE id IN (%s)", strings.Join(placeholders, ","))
		_, err := txConn.Exec(query, args...)
		if err != nil {
			return fmt.Errorf("delete dags failed: %w", err)
		}

		return nil
	})
}

// GetDagCount 获取dag数量
func (s *Store) GetDagCount(ctx context.Context, params map[string]interface{}) (int64, error) {
	query := "SELECT COUNT(*) FROM dags WHERE removed = false"
	args := []interface{}{}

	for key, value := range params {
		switch key {
		case "userid":
			query += " AND userid = ?"
			args = append(args, value)
		case "type":
			query += " AND type = ?"
			args = append(args, value)
		case "status":
			query += " AND status = ?"
			args = append(args, value)
		}
	}

	var count int64
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get dag count failed: %w", err)
	}

	return count, nil
}

// ListDagCount 列出dag数量
func (s *Store) ListDagCount(ctx context.Context, input *mod.ListDagInput) (int64, error) {
	query := "SELECT COUNT(*) FROM dags WHERE removed = false"
	args := []interface{}{}

	if input.UserID != "" {
		query += " AND userid = ?"
		args = append(args, input.UserID)
	}

	if len(input.Trigger) > 0 {
		placeholders := make([]string, len(input.Trigger))
		for i, t := range input.Trigger {
			placeholders[i] = "?"
			args = append(args, t)
		}
		query += fmt.Sprintf(" AND trigger IN (%s)", strings.Join(placeholders, ","))
	}

	if input.Type != "" {
		query += " AND type = ?"
		args = append(args, input.Type)
	}

	if input.BizDomainID != "" {
		query += " AND biz_domain_id = ?"
		args = append(args, input.BizDomainID)
	}

	if input.KeyWord != "" {
		query += " AND (name LIKE ? OR description LIKE ?)"
		args = append(args, "%"+input.KeyWord+"%", "%"+input.KeyWord+"%")
	}

	var count int64
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("list dag count failed: %w", err)
	}

	return count, nil
}

// ListDagCountByFields 根据字段列出dag数量
func (s *Store) ListDagCountByFields(ctx context.Context, filter bson.M) (int64, error) {
	query := "SELECT COUNT(*) FROM dags WHERE removed = false"
	args := []interface{}{}

	for key, value := range filter {
		switch key {
		case "userid":
			query += " AND userid = ?"
			args = append(args, value)
		case "type":
			query += " AND type = ?"
			args = append(args, value)
		case "trigger":
			query += " AND trigger = ?"
			args = append(args, value)
		}
	}

	var count int64
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("list dag count by fields failed: %w", err)
	}

	return count, nil
}

// GetDagInstanceCount 获取dag instance数量
func (s *Store) GetDagInstanceCount(ctx context.Context, params map[string]interface{}) (int64, error) {
	query := "SELECT COUNT(*) FROM dag_instances WHERE 1=1"
	args := []interface{}{}

	for key, value := range params {
		switch key {
		case "dag_id":
			query += " AND dag_id = ?"
			args = append(args, value)
		case "userid":
			query += " AND userid = ?"
			args = append(args, value)
		case "status":
			query += " AND status = ?"
			args = append(args, value)
		}
	}

	var count int64
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get dag instance count failed: %w", err)
	}

	return count, nil
}

// CreateInbox 创建inbox
func (s *Store) CreateInbox(ctx context.Context, msg *entity.InBox) error {
	msg.Initial()

	_, err := s.db.Exec(
		"INSERT INTO inboxes (id, docid, topic, created_at) VALUES (?, ?, ?, NOW())",
		msg.ID, msg.DocID, msg.Topic,
	)
	if err != nil {
		return fmt.Errorf("create inbox failed: %w", err)
	}

	return nil
}

// DeleteInbox 删除inbox
func (s *Store) DeleteInbox(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("DELETE FROM inboxes WHERE id IN (%s)", strings.Join(placeholders, ","))
	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("delete inbox failed: %w", err)
	}

	return nil
}

// GetInbox 获取inbox
func (s *Store) GetInbox(ctx context.Context, id string) (*entity.InBox, error) {
	inbox := &entity.InBox{}

	err := s.db.QueryRow(
		"SELECT id, docid, topic, created_at FROM inboxes WHERE id = ?",
		id,
	).Scan(&inbox.ID, &inbox.DocID, &inbox.Topic, &inbox.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrDataNotFound
		}
		return nil, fmt.Errorf("get inbox failed: %w", err)
	}

	return inbox, nil
}

// ListInbox 列出inbox
func (s *Store) ListInbox(ctx context.Context, input *mod.ListInboxInput) ([]*entity.InBox, error) {
	query := "SELECT id, docid, topic, created_at FROM inboxes WHERE 1=1"
	args := []interface{}{}

	if input.DocID != "" {
		query += " AND docid = ?"
		args = append(args, input.DocID)
	}

	if len(input.Topics) > 0 {
		placeholders := make([]string, len(input.Topics))
		for i, t := range input.Topics {
			placeholders[i] = "?"
			args = append(args, t)
		}
		query += fmt.Sprintf(" AND topic IN (%s)", strings.Join(placeholders, ","))
	}

	if input.Now > 0 {
		query += " AND created_at <= FROM_UNIXTIME(?)"
		args = append(args, input.Now)
	}

	if input.SortBy != "" {
		order := "ASC"
		if input.Order == -1 {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", input.SortBy, order)
	} else {
		query += " ORDER BY created_at ASC"
	}

	if input.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, input.Limit)
	}

	if input.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, input.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list inbox failed: %w", err)
	}
	defer rows.Close()

	var inboxes []*entity.InBox
	for rows.Next() {
		inbox := &entity.InBox{}
		err := rows.Scan(&inbox.ID, &inbox.DocID, &inbox.Topic, &inbox.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan inbox failed: %w", err)
		}
		inboxes = append(inboxes, inbox)
	}

	return inboxes, nil
}

// GetSwitchStatus 获取开关状态
func (s *Store) GetSwitchStatus() (bool, error) {
	var status bool
	err := s.db.QueryRow("SELECT status FROM switches WHERE name = 'mysql_switch'").Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("get switch status failed: %w", err)
	}

	return status, nil
}

// SetSwitchStatus 设置开关状态
func (s *Store) SetSwitchStatus(status bool) error {
	_, err := s.db.Exec(
		"INSERT INTO switches (id, name, status, created_at, updated_at) VALUES (?, 'mysql_switch', ?, NOW(), NOW()) ON DUPLICATE KEY UPDATE status = ?, updated_at = NOW()",
		utils.GetUUID(), status, status,
	)
	if err != nil {
		return fmt.Errorf("set switch status failed: %w", err)
	}

	return nil
}

// CreateLogs 创建日志
func (s *Store) CreateLogs(ctx context.Context, ossLogs []*entity.Log) error {
	if len(ossLogs) == 0 {
		return nil
	}

	for _, log := range ossLogs {
		_, err := s.db.Exec(
			"INSERT INTO logs (id, level, message, created_at) VALUES (?, ?, ?, NOW())",
			utils.GetUUID(), log.Level, log.Message,
		)
		if err != nil {
			return fmt.Errorf("create log failed: %w", err)
		}
	}

	return nil
}

// ListHistoryDagIns 列出历史dag instance
func (s *Store) ListHistoryDagIns(ctx context.Context, params map[string]interface{}, dataChannel chan []bson.M) error {
	defer close(dataChannel)

	query := `SELECT id, dag_id, trigger, worker, source, vars, keywords, event_persistence, 
		event_oss_path, share_data, status, reason, cmd, userid, ended_at, 
		dag_type, policy_type, appinfo, priority, mode, dump, success_callback, 
		error_callback, call_chain, resume_data, resume_status, version, version_id, 
		biz_domain_id, created_at, updated_at FROM dag_instances WHERE 1=1`
	args := []interface{}{}

	for key, value := range params {
		switch key {
		case "dag_id":
			query += " AND dag_id = ?"
			args = append(args, value)
		case "userid":
			query += " AND userid = ?"
			args = append(args, value)
		case "status":
			query += " AND status = ?"
			args = append(args, value)
		}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return fmt.Errorf("list history dag instances failed: %w", err)
	}
	defer rows.Close()

	var results []bson.M
	for rows.Next() {
		var id, dagID, trigger, worker, source, status, reason, userID, dagType, policyType, priority, resumeStatus, version, versionID, bizDomainID string
		var varsStr, keywordsStr, shareDataStr, cmdStr, appinfoStr, callChainStr string
		var eventPersistence int
		var eventOssPath, dump, successCallback, errorCallback, resumeData string
		var endedAt int64
		var mode int
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&id, &dagID, &trigger, &worker, &source, &varsStr, &keywordsStr, &eventPersistence,
			&eventOssPath, &shareDataStr, &status, &reason, &cmdStr, &userID, &endedAt,
			&dagType, &policyType, &appinfoStr, &priority, &mode, &dump, &successCallback,
			&errorCallback, &callChainStr, &resumeData, &resumeStatus, &version, &versionID,
			&bizDomainID, &createdAt, &updatedAt,
		)
		if err != nil {
			return fmt.Errorf("scan history dag instance failed: %w", err)
		}

		result := bson.M{
			"id":         id,
			"dag_id":     dagID,
			"trigger":    trigger,
			"worker":     worker,
			"source":     source,
			"status":     status,
			"reason":     reason,
			"userid":     userID,
			"ended_at":   endedAt,
			"created_at": createdAt,
			"updated_at": updatedAt,
		}

		results = append(results, result)
	}

	dataChannel <- results
	return nil
}

// ListHistoryTaskIns 列出历史task instance
func (s *Store) ListHistoryTaskIns(ctx context.Context, params map[string]interface{}, dataChannel chan []bson.M) error {
	defer close(dataChannel)

	query := `SELECT id, dag_ins_id, task_id, status, reason, traces, results, last_modified_at, 
		rendered_params, depend_on, hash, metadata, created_at, updated_at FROM task_instances WHERE 1=1`
	args := []interface{}{}

	for key, value := range params {
		switch key {
		case "dag_ins_id":
			query += " AND dag_ins_id = ?"
			args = append(args, value)
		case "status":
			query += " AND status = ?"
			args = append(args, value)
		}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return fmt.Errorf("list history task instances failed: %w", err)
	}
	defer rows.Close()

	var results []bson.M
	for rows.Next() {
		var id, dagInsID, taskID, status, reason, hash string
		var tracesStr, resultsStr, renderedParamsStr, dependOnStr, metadataStr string
		var lastModifiedAt int64
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&id, &dagInsID, &taskID, &status, &reason, &tracesStr, &resultsStr,
			&lastModifiedAt, &renderedParamsStr, &dependOnStr, &hash, &metadataStr,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return fmt.Errorf("scan history task instance failed: %w", err)
		}

		result := bson.M{
			"id":               id,
			"dag_ins_id":       dagInsID,
			"task_id":          taskID,
			"status":           status,
			"reason":           reason,
			"hash":             hash,
			"last_modified_at": lastModifiedAt,
			"created_at":       createdAt,
			"updated_at":       updatedAt,
		}

		results = append(results, result)
	}

	dataChannel <- results
	return nil
}

// DeleteDagInsByID 根据ID删除dag instance
func (s *Store) DeleteDagInsByID(ctx context.Context, params map[string]interface{}) error {
	query := "DELETE FROM dag_instances WHERE 1=1"
	args := []interface{}{}

	for key, value := range params {
		switch key {
		case "id":
			query += " AND id = ?"
			args = append(args, value)
		case "dag_id":
			query += " AND dag_id = ?"
			args = append(args, value)
		}
	}

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("delete dag instance by id failed: %w", err)
	}

	return nil
}

// DeleteTaskInsByID 根据ID删除task instance
func (s *Store) DeleteTaskInsByID(ctx context.Context, params map[string]interface{}) error {
	query := "DELETE FROM task_instances WHERE 1=1"
	args := []interface{}{}

	for key, value := range params {
		switch key {
		case "id":
			query += " AND id = ?"
			args = append(args, value)
		case "dag_ins_id":
			query += " AND dag_ins_id = ?"
			args = append(args, value)
		}
	}

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("delete task instance by id failed: %w", err)
	}

	return nil
}

// DeleteTaskInsByDagInsID 根据dag instance ID删除task instance
func (s *Store) DeleteTaskInsByDagInsID(ctx context.Context, dagInsID string) error {
	_, err := s.db.Exec("DELETE FROM task_instances WHERE dag_ins_id = ?", dagInsID)
	if err != nil {
		return fmt.Errorf("delete task instance by dag instance id failed: %w", err)
	}

	return nil
}

// GetTaskInstanceCount 获取task instance数量
func (s *Store) GetTaskInstanceCount(ctx context.Context, params map[string]interface{}) (int64, error) {
	query := "SELECT COUNT(*) FROM task_instances WHERE 1=1"
	args := []interface{}{}

	for key, value := range params {
		switch key {
		case "dag_ins_id":
			query += " AND dag_ins_id = ?"
			args = append(args, value)
		case "status":
			query += " AND status = ?"
			args = append(args, value)
		}
	}

	var count int64
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get task instance count failed: %w", err)
	}

	return count, nil
}

// CreatOutBoxMessage 创建outbox消息
func (s *Store) CreatOutBoxMessage(ctx context.Context, outBox *entity.OutBox) error {
	outBox.Initial()

	_, err := s.db.Exec(
		"INSERT INTO outboxes (id, topic, message, created_at) VALUES (?, ?, ?, NOW())",
		outBox.ID, outBox.Topic, JSONMarshal(outBox.Message),
	)
	if err != nil {
		return fmt.Errorf("create outbox message failed: %w", err)
	}

	return nil
}

// BatchCreatOutBoxMessage 批量创建outbox消息
func (s *Store) BatchCreatOutBoxMessage(ctx context.Context, outBox []*entity.OutBox) error {
	if len(outBox) == 0 {
		return nil
	}

	for _, ob := range outBox {
		ob.Initial()

		_, err := s.db.Exec(
			"INSERT INTO outboxes (id, topic, message, created_at) VALUES (?, ?, ?, NOW())",
			ob.ID, ob.Topic, JSONMarshal(ob.Message),
		)
		if err != nil {
			return fmt.Errorf("create outbox message failed: %w", err)
		}
	}

	return nil
}

// DeleteOutBoxMessage 删除outbox消息
func (s *Store) DeleteOutBoxMessage(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("DELETE FROM outboxes WHERE id IN (%s)", strings.Join(placeholders, ","))
	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("delete outbox message failed: %w", err)
	}

	return nil
}

// ListOutBoxMessage 列出outbox消息
func (s *Store) ListOutBoxMessage(ctx context.Context, input *entity.OutBoxInput) ([]*entity.OutBox, error) {
	query := "SELECT id, topic, message, created_at FROM outboxes WHERE 1=1"
	args := []interface{}{}

	if input.ID != "" {
		query += " AND id = ?"
		args = append(args, input.ID)
	}

	if input.Now > 0 {
		query += " AND created_at <= FROM_UNIXTIME(?)"
		args = append(args, input.Now)
	}

	if input.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, input.Limit)
	}

	if input.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, input.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list outbox message failed: %w", err)
	}
	defer rows.Close()

	var outboxes []*entity.OutBox
	for rows.Next() {
		ob := &entity.OutBox{}
		var messageStr string
		var createdAt time.Time

		err := rows.Scan(&ob.ID, &ob.Topic, &messageStr, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("scan outbox message failed: %w", err)
		}

		ob.CreatedAt = createdAt
		JSONUnmarshal(messageStr, &ob.Message)

		outboxes = append(outboxes, ob)
	}

	return outboxes, nil
}

// ListDagInstanceInRangeTime 根据时间范围列出dag instance
func (s *Store) ListDagInstanceInRangeTime(ctx context.Context, status string, begin, end int64) ([]*entity.DagInstance, error) {
	query := `SELECT id, dag_id, trigger, worker, source, vars, keywords, event_persistence, 
		event_oss_path, share_data, status, reason, cmd, userid, ended_at, 
		dag_type, policy_type, appinfo, priority, mode, dump, success_callback, 
		error_callback, call_chain, resume_data, resume_status, version, version_id, 
		biz_domain_id, created_at, updated_at FROM dag_instances WHERE status = ? AND created_at >= FROM_UNIXTIME(?) AND created_at <= FROM_UNIXTIME(?)`

	rows, err := s.db.Query(query, status, begin, end)
	if err != nil {
		return nil, fmt.Errorf("list dag instance in range time failed: %w", err)
	}
	defer rows.Close()

	var dagInstances []*entity.DagInstance
	for rows.Next() {
		dagIns := &entity.DagInstance{}
		var trigger, dagStatus, dagType, policyType, priority, resumeStatus, version, versionID, bizDomainID string
		var varsStr, keywordsStr, shareDataStr, cmdStr, appinfoStr, callChainStr string
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&dagIns.ID, &dagIns.DagID, &trigger, &dagIns.Worker, &dagIns.Source,
			&varsStr, &keywordsStr, &dagIns.EventPersistence, &dagIns.EventOssPath,
			&shareDataStr, &dagStatus, &dagIns.Reason, &cmdStr, &dagIns.UserID, &dagIns.EndedAt,
			&dagType, &policyType, &appinfoStr, &priority, &dagIns.Mode, &dagIns.Dump,
			&dagIns.SuccessCallback, &dagIns.ErrorCallback, &callChainStr, &dagIns.ResumeData,
			&resumeStatus, &version, &versionID, &bizDomainID, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan dag instance failed: %w", err)
		}

		dagIns.Trigger = entity.DagTrigger(trigger)
		dagIns.Status = entity.DagInstanceStatus(dagStatus)
		dagIns.DagType = dagType
		dagIns.PolicyType = policyType
		dagIns.Priority = priority
		dagIns.ResumeStatus = entity.DagInstanceStatus(resumeStatus)
		dagIns.Version = entity.NewVersion(version)
		dagIns.VersionID = versionID
		dagIns.BizDomainID = bizDomainID
		dagIns.CreatedAt = createdAt
		dagIns.UpdatedAt = updatedAt

		JSONUnmarshal(varsStr, &dagIns.Vars)
		JSONUnmarshal(keywordsStr, &dagIns.Keywords)
		JSONUnmarshal(shareDataStr, &dagIns.ShareData)
		JSONUnmarshal(cmdStr, &dagIns.Cmd)
		JSONUnmarshal(appinfoStr, &dagIns.AppInfo)
		JSONUnmarshal(callChainStr, &dagIns.CallChain)

		dagInstances = append(dagInstances, dagIns)
	}

	return dagInstances, nil
}

// ListExistDagInsID 列出存在的dag instance ID
func (s *Store) ListExistDagInsID(ctx context.Context, dagInsIDs []string) ([]string, error) {
	if len(dagInsIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(dagInsIDs))
	args := make([]interface{}, len(dagInsIDs))
	for i, id := range dagInsIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("SELECT id FROM dag_instances WHERE id IN (%s)", strings.Join(placeholders, ","))
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list exist dag instance id failed: %w", err)
	}
	defer rows.Close()

	var existIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan dag instance id failed: %w", err)
		}
		existIDs = append(existIDs, id)
	}

	return existIDs, nil
}

// ListExistDagID 列出存在的dag ID
func (s *Store) ListExistDagID(ctx context.Context, dagIDs []string) ([]string, error) {
	if len(dagIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(dagIDs))
	args := make([]interface{}, len(dagIDs))
	for i, id := range dagIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("SELECT id FROM dags WHERE id IN (%s) AND removed = false", strings.Join(placeholders, ","))
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list exist dag id failed: %w", err)
	}
	defer rows.Close()

	var existIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan dag id failed: %w", err)
		}
		existIDs = append(existIDs, id)
	}

	return existIDs, nil
}

// GroupDagInstance 分组dag instance
func (s *Store) GroupDagInstance(ctx context.Context, input *mod.GroupInput) ([]*entity.DagInstanceGroup, error) {
	if input.GroupBy == "" && len(input.GroupBys) == 0 {
		return nil, fmt.Errorf("group by field is required")
	}

	var groupField string
	if input.GroupBy != "" {
		groupField = input.GroupBy
	} else {
		groupField = input.GroupBys[0]
	}

	query := fmt.Sprintf("SELECT %s, COUNT(*) as total FROM dag_instances WHERE 1=1", groupField)
	args := []interface{}{}

	for _, opt := range input.SearchOptions {
		switch opt.Condition {
		case "$eq":
			query += fmt.Sprintf(" AND %s = ?", opt.Field)
			args = append(args, opt.Value)
		case "$ne":
			query += fmt.Sprintf(" AND %s != ?", opt.Field)
			args = append(args, opt.Value)
		case "$in":
			query += fmt.Sprintf(" AND %s IN (?)", opt.Field)
			args = append(args, opt.Value)
		}
	}

	if input.TimeRange != nil {
		query += fmt.Sprintf(" AND %s >= ? AND %s <= ?", input.TimeRange.Field, input.TimeRange.Field)
		args = append(args, input.TimeRange.Begin, input.TimeRange.End)
	}

	query += fmt.Sprintf(" GROUP BY %s", groupField)

	if input.SortBy != "" {
		order := "ASC"
		if input.Order == -1 {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", input.SortBy, order)
	}

	if input.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, input.Limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("group dag instance failed: %w", err)
	}
	defer rows.Close()

	var groups []*entity.DagInstanceGroup
	for rows.Next() {
		group := &entity.DagInstanceGroup{}
		var fieldValue string
		var total int64

		err := rows.Scan(&fieldValue, &total)
		if err != nil {
			return nil, fmt.Errorf("scan group failed: %w", err)
		}

		group.ID = fieldValue
		group.Total = total

		groups = append(groups, group)
	}

	return groups, nil
}

// RetryDagIns 重试dag instance
func (s *Store) RetryDagIns(ctx context.Context, dagInsID string, taskInsIDs []string) error {
	return s.WithTransaction(ctx, func(tx interface{}) error {
		txConn := tx.(*sql.Tx)

		_, err := txConn.Exec(
			"UPDATE dag_instances SET status = 'init', reason = '', updated_at = NOW() WHERE id = ?",
			dagInsID,
		)
		if err != nil {
			return fmt.Errorf("update dag instance failed: %w", err)
		}

		if len(taskInsIDs) > 0 {
			placeholders := make([]string, len(taskInsIDs))
			args := make([]interface{}, len(taskInsIDs))
			for i, id := range taskInsIDs {
				placeholders[i] = "?"
				args[i] = id
			}

			query := fmt.Sprintf("UPDATE task_instances SET status = 'init', reason = '', updated_at = NOW() WHERE id IN (%s)", strings.Join(placeholders, ","))
			_, err = txConn.Exec(query, args...)
			if err != nil {
				return fmt.Errorf("update task instances failed: %w", err)
			}
		}

		return nil
	})
}

// DeleteDag 删除dag
func (s *Store) DeleteDag(ctx context.Context, id ...string) error {
	if len(id) == 0 {
		return nil
	}

	placeholders := make([]string, len(id))
	args := make([]interface{}, len(id))
	for i, idVal := range id {
		placeholders[i] = "?"
		args[i] = idVal
	}

	query := fmt.Sprintf("DELETE FROM dags WHERE id IN (%s)", strings.Join(placeholders, ","))
	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("delete dag failed: %w", err)
	}

	return nil
}

// CreateDagVersion 创建dag版本
func (s *Store) CreateDagVersion(ctx context.Context, dagVersion *entity.DagVersion) (string, error) {
	dagVersion.Initial()

	configStr := JSONMarshal(dagVersion.Config)

	_, err := s.db.Exec(
		"INSERT INTO dag_versions (id, dag_id, version_id, config, sort_time, created_at, updated_at) VALUES (?, ?, ?, ?, ?, NOW(), NOW())",
		dagVersion.ID, dagVersion.DagID, dagVersion.VersionID, configStr, dagVersion.SortTime,
	)
	if err != nil {
		return "", fmt.Errorf("create dag version failed: %w", err)
	}

	return dagVersion.ID, nil
}

// ListDagVersions 列出dag版本
func (s *Store) ListDagVersions(ctx context.Context, input *mod.ListDagVersionInput) ([]entity.DagVersion, error) {
	query := "SELECT id, dag_id, version_id, config, sort_time, created_at, updated_at FROM dag_versions WHERE dag_id = ?"
	args := []interface{}{input.DagID}

	if input.SortBy != "" {
		order := "ASC"
		if input.Order == -1 {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", input.SortBy, order)
	} else {
		query += " ORDER BY sort_time DESC"
	}

	if input.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, input.Limit)
	}

	if input.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, input.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list dag versions failed: %w", err)
	}
	defer rows.Close()

	var versions []entity.DagVersion
	for rows.Next() {
		var version entity.DagVersion
		var configStr string
		var createdAt, updatedAt time.Time

		err := rows.Scan(&version.ID, &version.DagID, &version.VersionID, &configStr, &version.SortTime, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan dag version failed: %w", err)
		}

		version.CreatedAt = createdAt
		version.UpdatedAt = updatedAt

		if configStr != "" && configStr != "null" {
			json.Unmarshal([]byte(configStr), &version.Config)
		}

		versions = append(versions, version)
	}

	return versions, nil
}

// GetHistoryDagByVersionID 根据版本ID获取历史dag
func (s *Store) GetHistoryDagByVersionID(ctx context.Context, dagID, versionID string) (*entity.DagVersion, error) {
	var dagVersion entity.DagVersion
	var configStr string
	var createdAt, updatedAt time.Time

	err := s.db.QueryRow(
		"SELECT id, dag_id, version_id, config, sort_time, created_at, updated_at FROM dag_versions WHERE dag_id = ? AND version_id = ?",
		dagID, versionID,
	).Scan(&dagVersion.ID, &dagVersion.DagID, &dagVersion.VersionID, &configStr, &dagVersion.SortTime, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrDataNotFound
		}
		return nil, fmt.Errorf("get history dag by version id failed: %w", err)
	}

	dagVersion.CreatedAt = createdAt
	dagVersion.UpdatedAt = updatedAt

	if configStr != "" && configStr != "null" {
		json.Unmarshal([]byte(configStr), &dagVersion.Config)
	}

	return &dagVersion, nil
}
