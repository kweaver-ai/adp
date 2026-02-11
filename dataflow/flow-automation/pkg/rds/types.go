package rds

import (
	"fmt"
)

const (
	CONF_TABLENAME              = "t_automation_conf"
	AI_MODEL_TABLENAME          = "t_model"
	ALARM_RULE_TABLENAME        = "t_alarm_rule"
	ALARM_USER_TABLENAME        = "t_alarm_user"
	CONTENT_ADMIN_TABLENAME     = "t_content_admin"
	AGENT_TABLENAME             = "t_automation_agent"
	DAG_INSTANCE_EVENT_TABLE    = "t_dag_instance_event"
	DAG_INSTANCE_EXT_DATA_TABLE = "t_automation_dag_instance_ext_data"
	EXECUTOR_TABLENAME          = "t_automation_executor"
	EXECUTOR_ACCESSOR_TABLENAME = "t_automation_executor_accessor"
	EXECUTOR_ACTION_TABLENAME   = "t_automation_executor_action"
)

const (
	TaskCacheTableFormat = `t_task_cache_%s`
)

type ConfModel struct {
	Key   *string `gorm:"column:f_key;type:char(32);primary_key:not null" json:"key"`
	Value *string `gorm:"column:f_value;type:char(255)" json:"value"`
}

type AiModel struct {
	ID          uint64 `gorm:"column:f_id;primary_key:not null" json:"id"`
	CreatedAt   int64  `gorm:"column:f_created_at;type:bigint" json:"created_at"`
	UpdatedAt   int64  `gorm:"column:f_updated_at;type:bigint" json:"updated_at"`
	TrainStatus string `gorm:"column:f_train_status;type:varchar(16)" json:"train_status"`
	Status      int    `gorm:"column:f_status;type:tinyint" json:"status"`
	Rule        string `gorm:"column:f_rule;type:text" json:"rule"`
	Name        string `gorm:"column:f_name;type:varchar(255)" json:"name"`
	Description string `gorm:"column:f_description;type:varchar(300)" json:"description"`
	UserID      string `gorm:"column:f_userid;type:varchar(40)" json:"userID"`
	Type        int    `gorm:"column:f_type;type:tinyint" json:"type"`
}

type TrainFileOSSInfo struct {
	ID        uint64 `gorm:"column:f_id;primary_key:not null" json:"id"`
	TrainID   uint64 `gorm:"column:f_train_id;primary_key:not null" json:"trainID"`
	OSSID     string `gorm:"column:f_oss_id;type:varchar(36)" json:"ossID"`
	Key       string `gorm:"column:f_key;type:varchar(36)" json:"key"`
	CreatedAt int64  `gorm:"column:f_created_at;type:bigint" json:"created_at"`
}

type ContentAdmin struct {
	ID       uint64 `gorm:"column:f_id;primary_key:not null" json:"id"`
	UserID   string `gorm:"column:f_user_id;type:varchar(40)" json:"userID"`
	UserName string `gorm:"column:f_user_name;type:varchar(128)" json:"userName"`
}

type AlarmRule struct {
	ID        uint64 `gorm:"column:f_id;primary_key:not null" json:"id"`
	RuleID    uint64 `gorm:"column:f_rule_id;type:bigint" json:"ruleID"`
	DagID     uint64 `gorm:"column:f_dag_id;type:bigint" json:"dagID"`
	Frequency int    `gorm:"column:f_frequency;type:unsigned smallint" json:"frequency"`
	Threshold int    `gorm:"column:f_threshold;type:unsigned mediumint" json:"threshold"`
	CreatedAt int64  `gorm:"column:f_created_at;type:bigint" json:"created_at"`
}

type AlarmUser struct {
	ID       uint64 `gorm:"column:f_id;primary_key:not null" json:"id"`
	RuleID   uint64 `gorm:"column:f_rule_id;type:bigint" json:"ruleID"`
	UserID   string `gorm:"column:f_user_id;type:varchar(36)" json:"userID"`
	UserName string `gorm:"column:f_user_name;type:varchar(128)" json:"userName"`
	UserType string `gorm:"column:f_user_type;type:varchar(10)" json:"userType"`
}

type AgentModel struct {
	ID      uint64 `gorm:"column:f_id;type:bigint unsigned;primary_key:not null" json:"-"`
	Name    string `gorm:"column:f_name;type:varchar(128);not null;default:''" json:"name"`
	AgentID string `gorm:"column:f_agent_id;type:varchar(64);not null;default:''" json:"agent_id"`
	Version string `gorm:"column:f_version;type:varchar(32);not null;default:''" json:"version"`
}

type DagInstanceEventType uint8

const (
	DagInstanceEventTypeVariable     DagInstanceEventType = 1
	DagInstanceEventTypeTaskStatus   DagInstanceEventType = 2
	DagInstanceEventTypeInstructions DagInstanceEventType = 3
	DagInstanceEventTypeVM           DagInstanceEventType = 4
	DagInstanceEventTypeTrace        DagInstanceEventType = 5
)

type DagInstanceEventVisibility uint8

const (
	DagInstanceEventVisibilityPrivate = 0
	DagInstanceEventVisibilityPublic  = 1
)

type DagInstanceEvent struct {
	ID         uint64                     `gorm:"column:f_id" json:"id,omitempty"`
	Type       DagInstanceEventType       `gorm:"column:f_type" json:"type,omitempty"`
	InstanceID string                     `gorm:"column:f_instance_id" json:"instance_id,omitempty"`
	Operator   string                     `gorm:"column:f_operator" json:"operator,omitempty"`
	TaskID     string                     `gorm:"column:f_task_id" json:"task_id,omitempty"`
	Status     string                     `gorm:"column:f_status" json:"status,omitempty"`
	Name       string                     `gorm:"column:f_name" json:"name,omitempty"`
	Data       string                     `gorm:"column:f_data" json:"data,omitempty"`
	Size       int                        `gorm:"column:f_size" json:"size,omitempty"`
	Inline     bool                       `gorm:"column:f_inline" json:"inline,omitempty"`
	Visibility DagInstanceEventVisibility `gorm:"column:f_visibility" json:"visibility,omitempty"`
	Timestamp  int64                      `gorm:"column:f_timestamp" json:"timestamp,omitempty"`
}

type DagInstanceEventField string

const (
	DagInstanceEventFieldID         DagInstanceEventField = "f_id"
	DagInstanceEventFieldType       DagInstanceEventField = "f_type"
	DagInstanceEventFieldInstanceID DagInstanceEventField = "f_instance_id"
	DagInstanceEventFieldOperator   DagInstanceEventField = "f_operator"
	DagInstanceEventFieldTaskID     DagInstanceEventField = "f_task_id"
	DagInstanceEventFieldStatus     DagInstanceEventField = "f_status"
	DagInstanceEventFieldName       DagInstanceEventField = "f_name"
	DagInstanceEventFieldData       DagInstanceEventField = "f_data"
	DagInstanceEventFieldSize       DagInstanceEventField = "f_size"
	DagInstanceEventFieldInline     DagInstanceEventField = "f_inline"
	DagInstanceEventFieldTimestamp  DagInstanceEventField = "f_timestamp"
	DagInstanceEventFieldVisibility DagInstanceEventField = "f_visibility"
)

var (
	DagInstanceEventFieldAll = []DagInstanceEventField{
		DagInstanceEventFieldID,
		DagInstanceEventFieldType,
		DagInstanceEventFieldInstanceID,
		DagInstanceEventFieldOperator,
		DagInstanceEventFieldTaskID,
		DagInstanceEventFieldStatus,
		DagInstanceEventFieldName,
		DagInstanceEventFieldData,
		DagInstanceEventFieldSize,
		DagInstanceEventFieldInline,
		DagInstanceEventFieldTimestamp,
		DagInstanceEventFieldVisibility,
	}
	DagInstanceEventFieldPublic = []DagInstanceEventField{
		DagInstanceEventFieldType,
		DagInstanceEventFieldOperator,
		DagInstanceEventFieldTaskID,
		DagInstanceEventFieldStatus,
		DagInstanceEventFieldName,
		DagInstanceEventFieldData,
		DagInstanceEventFieldSize,
		DagInstanceEventFieldInline,
		DagInstanceEventFieldTimestamp,
	}
)

type DagInstanceEventListOptions struct {
	DagInstanceID string
	Offset        int
	Limit         int
	Visibilities  []DagInstanceEventVisibility
	Types         []DagInstanceEventType
	Fields        []DagInstanceEventField
	Names         []string
	Inline        *bool
	LatestOnly    bool
}

type DagInstanceExtData struct {
	ID        string `gorm:"column:f_id;primary_key:not null" json:"id" bson:"_id"`
	CreatedAt int64  `gorm:"column:f_created_at;type:bigint" json:"createdAt" bson:"createdAt"`
	UpdatedAt int64  `gorm:"column:f_updated_at;type:bigint" json:"updatedAt" bson:"updatedAt"`
	DagID     string `gorm:"column:f_dag_id;type:varchar(64)" json:"dagId" bson:"dagId"`
	DagInsID  string `gorm:"column:f_dag_ins_id;type:varchar(64)" json:"dagInsId" bson:"dagInsId"`
	Field     string `gorm:"column:f_field;type:varchar(64)" json:"field" bson:"field"`
	OssID     string `gorm:"column:f_oss_id;type:varchar(64)" json:"ossId" bson:"ossId"`
	OssKey    string `gorm:"column:f_oss_key;type:varchar(255)" json:"ossKey" bson:"ossKey"`
	Size      int64  `gorm:"column:f_size;type:bigint" json:"size" bson:"size"`
	Removed   bool   `gorm:"column:f_removed;type:tinyint(1)" json:"removed" bson:"removed"`
}

type ExtDataQueryOptions struct {
	IDs         []string
	DagID       string
	DagInsID    string
	Removed     bool
	Limit       int
	MinID       string
	SelectField []string
}

type ExecutorModel struct {
	ID          *uint64                  `gorm:"column:f_id;primary_key:not null" json:"id"`
	Name        *string                  `gorm:"column:f_name;type:varchar(64)" json:"name"`
	Description *string                  `gorm:"column:f_description;type:varchar(256)" json:"description"`
	CreatorID   *string                  `gorm:"column:f_creator_id;type:varchar(40)" json:"creator_id"`
	Status      *int                     `gorm:"column:f_status;type:tinyint" json:"status"`
	CreatedAt   *int64                   `gorm:"column:f_created_at;type:bigint" json:"created_at"`
	UpdatedAt   *int64                   `gorm:"column:f_updated_at;type:bigint" json:"updated_at"`
	Accessors   []*ExecutorAccessorModel `gorm:"-" json:"accessors"`
	Actions     []*ExecutorActionModel   `gorm:"-" json:"actions"`
}

type ExecutorAccessorModel struct {
	ID           *uint64 `gorm:"column:f_id;primary_key:not null" json:"id"`
	ExecutorID   *uint64 `gorm:"column:f_executor_id;primary_key:not null" json:"executor_id"`
	AccessorID   *string `gorm:"column:f_accessor_id;type:varchar(40)" json:"accessor_id"`
	AccessorType *string `gorm:"column:f_accessor_type;type:varchar(20)" json:"accessor_type"`
}

type ExecutorActionModel struct {
	ID          *uint64 `gorm:"column:f_id;primary_key:not null" json:"id"`
	ExecutorID  *uint64 `gorm:"column:f_executor_id;primary_key:not null" json:"executor_id"`
	Operator    *string `gorm:"column:f_operator;type:varchar(64)" json:"operator"`
	Name        *string `gorm:"column:f_name;type:varchar(64)" json:"name"`
	Description *string `gorm:"column:f_description;type:varchar(64)" json:"description"`
	Group       *string `gorm:"column:f_group;type:varchar(64)" json:"group"`
	Type        *string `gorm:"column:f_type;type:varchar(16)" json:"type"`
	Inputs      *string `gorm:"column:f_inputs;type:text" json:"inputs"`
	Outputs     *string `gorm:"column:f_outputs;type:text" json:"outputs"`
	Config      *string `gorm:"column:f_config;type:text" json:"config"`
	CreatedAt   *int64  `gorm:"column:f_created_at;type:bigint" json:"created_at"`
	UpdatedAt   *int64  `gorm:"column:f_updated_at;type:bigint" json:"updated_at"`
}

type ExecutorWithActionModel struct {
	ID          *uint64 `gorm:"column:f_id;primary_key:not null" json:"id"`
	Name        *string `gorm:"column:f_name;type:varchar(64)" json:"name"`
	Description *string `gorm:"column:f_description;type:varchar(256)" json:"description"`
	CreatorID   *string `gorm:"column:f_creator_id;type:varchar(40)" json:"creator_id"`
	Status      *int    `gorm:"column:f_status;type:tinyint" json:"status"`
	CreatedAt   *int64  `gorm:"column:f_created_at;type:bigint" json:"created_at"`
	UpdatedAt   *int64  `gorm:"column:f_updated_at;type:bigint" json:"updated_at"`

	ActionID          *uint64 `gorm:"column:f_action_id;type:bigint" json:"action_id"`
	ActionOperator    *string `gorm:"column:f_action_operator;type:varchar(64)" json:"action_operator"`
	ActionName        *string `gorm:"column:f_action_name;type:varchar(64)" json:"action_name"`
	ActionDescription *string `gorm:"column:f_action_description;type:varchar(256)" json:"action_description"`
	ActionGroup       *string `gorm:"column:f_action_group;type:varchar(64)" json:"action_group"`
	ActionType        *string `gorm:"column:f_action_type;type:varchar(64)" json:"action_type"`
	ActionInputs      *string `gorm:"column:f_action_inputs;type:varchar(256)" json:"action_inputs"`
	ActionOutputs     *string `gorm:"column:f_action_outputs;type:varchar(256)" json:"action_outputs"`
	ActionConfig      *string `gorm:"column:f_action_config;type:varchar(256)" json:"action_config"`
	ActionCreatedAt   *int64  `gorm:"column:f_action_created_at;type:bigint" json:"action_created_at"`
	ActionUpdatedAt   *int64  `gorm:"column:f_action_updated_at;type:bigint" json:"action_updated_at"`
}

type TaskStatus int8

const (
	TaskStatusPending TaskStatus = 1
	TaskStatusSuccess TaskStatus = 2
	TaskStatusFailed  TaskStatus = 3
)

type TaskCacheItem struct {
	ID         uint64     `gorm:"column:f_id;primaryKey;type:char(64);not null" json:"id"`
	Hash       string     `gorm:"column:f_hash;type:char(40);not null;default:''" json:"hash"`
	Type       string     `gorm:"column:f_type;type:varchar(32);not null;default:''" json:"type"`
	Status     TaskStatus `gorm:"column:f_status;type:tinyint(4);not null;default:0" json:"status"`
	OssID      string     `gorm:"column:f_oss_id;type:char(36);not null;default:''" json:"ossId"`
	OssKey     string     `gorm:"column:f_oss_key;type:varchar(255);not null;default:''" json:"ossKey"`
	Ext        string     `gorm:"column:f_ext;type:char(20);not null;default:''" json:"ext"`
	Size       int64      `gorm:"column:f_size;type:bigint(20);not null;default:0" json:"size"`
	ErrMsg     string     `gorm:"column:f_err_msg;type:text" json:"errMsg"`
	CreateTime int64      `gorm:"column:f_create_time;type:bigint(20);not null;default:0" json:"createTime"`
	ModifyTime int64      `gorm:"column:f_modify_time;type:bigint(20);not null;default:0" json:"modifyTime"`
	ExpireTime int64      `gorm:"column:f_expire_time;type:bigint(20);not null;default:0" json:"expireTime"`
}

type ListTaskCacheOptions struct {
	TableSuffix string
	Expired     *bool
	Limit       int64
	MinID       uint64
}

type Options struct {
	OrderBy       *string
	Order         *string
	Limit         *int64
	Page          *int64
	SearchOptions []*SearchOption
}

type SearchOption struct {
	Col       string
	Val       interface{}
	Condition string
}

type UpdateParams struct {
	Status      *int64  `column:"f_status"`
	Rule        *string `column:"f_rule"`
	Name        *string `column:"f_name"`
	Description *string `column:"f_description"`
}

type UpdateCondition struct {
	ID     *string `column:"f_id"`
	UserID *string `column:"f_userid"`
}

type QueryCondition UpdateCondition

type ListParams struct {
	UserID *string
	Status *int64
	Name   *string
}

func (opt *Options) BuildQuery(baseQuery string) (sqlStr string, searchSqlVal []interface{}) {
	sqlStr = baseQuery
	if opt == nil {
		return
	}

	if len(opt.SearchOptions) != 0 {
		var searchSqlStr string
		for _, val := range opt.SearchOptions {
			searchSqlStr = fmt.Sprintf("AND %s %s ? ", val.Col, val.Condition)
			searchSqlVal = append(searchSqlVal, val.Val)
		}
		sqlStr = fmt.Sprintf("%s %s", sqlStr, searchSqlStr)
	}

	if opt.Order != nil && opt.OrderBy != nil {
		sqlStr = fmt.Sprintf("%s ORDER BY %s %s", sqlStr, *opt.OrderBy, *opt.Order)
	}

	if opt.Limit != nil && opt.Page != nil {
		offset := (*opt.Limit) * (*opt.Page)
		sqlStr = fmt.Sprintf("%s LIMIT %v, %v", sqlStr, offset, *opt.Limit)
	}

	return
}
