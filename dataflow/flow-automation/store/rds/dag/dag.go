package dagmodel

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/db"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/event"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
	data "github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/utils/data"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
	"github.com/shiningrush/goevent"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

const (
	DAG_TABLENAME                = "t_dag"
	DAGINSTANCE_TABLENAME        = "t_dag_instance"
	TASKINSTANCE_TABLENAME       = "t_task_instance"
	DAGVAR_TABLENAME             = "t_dag_var"
	DAGVERSIONS_TABLENAME        = "t_dag_versions"
	DAGSTEPINDEX_TABLENAME       = "t_dag_step_index"
	DAGTRIGGERINDEX_TABLENAME    = "t_dag_trigger_config_index"
	DAGACCESSORINDEX_TABLENAME   = "t_dag_accessor_index"
	DAGINSTANCEKEYWORD_TABLENAME = "t_dag_instance_keyword"
	OUTBOXMESSAGE_TABLENAME      = "t_outbox"
	INBOXMESSAGE_TABLENAME       = "t_inbox"
)

// Dag 流程配置数据库模型
type Dag struct {
	ID            uint64 `json:"f_id" gorm:"column:f_id;primaryKey"`
	CreatedAt     int64  `json:"f_created_at" gorm:"column:f_created_at"`
	UpdatedAt     int64  `json:"f_updated_at" gorm:"column:f_updated_at"`
	UserID        string `json:"f_user_id" gorm:"column:f_user_id"`
	Name          string `json:"f_name" gorm:"column:f_name"`
	Desc          string `json:"f_desc" gorm:"column:f_desc;type:longtext"`
	Trigger       string `json:"f_trigger" gorm:"column:f_trigger;type:longtext"`
	Cron          string `json:"f_cron" gorm:"column:f_cron"`
	Vars          string `json:"f_vars" gorm:"column:f_vars;type:longtext"`
	Status        string `json:"f_status" gorm:"column:f_status"`
	Tasks         string `json:"f_tasks" gorm:"column:f_tasks;type:longtext"`
	Steps         string `json:"f_steps" gorm:"column:f_steps;type:longtext"`
	Description   string `json:"f_description" gorm:"column:f_description;type:longtext"`
	Shortcuts     string `json:"f_shortcuts" gorm:"column:f_shortcuts;type:longtext"`
	Accessors     string `json:"f_accessors" gorm:"column:f_accessors;type:longtext"`
	Type          string `json:"f_type" gorm:"column:f_type"`
	PolicyType    string `json:"f_policy_type" gorm:"column:f_policy_type"`
	AppInfo       string `json:"f_appinfo" gorm:"column:f_appinfo;type:longtext"`
	Priority      string `json:"f_priority" gorm:"column:f_priority"`
	Removed       bool   `json:"f_removed" gorm:"column:f_removed"`
	Emails        string `json:"f_emails" gorm:"column:f_emails;type:longtext"`
	Template      string `json:"f_template" gorm:"column:f_template"`
	Published     bool   `json:"f_published" gorm:"column:f_published"`
	TriggerConfig string `json:"f_trigger_config" gorm:"column:f_trigger_config;type:longtext"`
	SubIDs        string `json:"f_sub_ids" gorm:"column:f_sub_ids;type:longtext"`
	ExecMode      string `json:"f_exec_mode" gorm:"column:f_exec_mode"`
	Category      string `json:"f_category" gorm:"column:f_category"`
	OutPuts       string `json:"f_outputs" gorm:"column:f_outputs;type:longtext"`
	Instructions  string `json:"f_instructions" gorm:"column:f_instructions;type:longtext"`
	OperatorID    string `json:"f_operator_id" gorm:"column:f_operator_id"`
	IncValues     string `json:"f_inc_values" gorm:"column:f_inc_values;type:longtext"`
	Version       string `json:"f_version" gorm:"column:f_version;type:longtext"`
	VersionID     string `json:"f_version_id" gorm:"column:f_version_id"`
	ModifyBy      string `json:"f_modify_by" gorm:"column:f_modify_by"`
	IsDebug       bool   `json:"f_is_debug" gorm:"column:f_is_debug"`
	DeBugID       string `json:"f_debug_id" gorm:"column:f_debug_id"`
	BizDomainID   string `json:"f_biz_domain_id" gorm:"column:f_biz_domain_id"`
}

// DagVar 流程配置变量数据库模型
type DagVar struct {
	ID           uint64 `json:"f_id" gorm:"column:f_id;primaryKey;autoIncrement"`
	DagID        uint64 `json:"f_dag_id" gorm:"column:f_dag_id"`
	VarName      string `json:"f_var_name" gorm:"column:f_var_name"`
	DefaultValue string `json:"f_default_value" gorm:"column:f_default_value;type:text"`
	VarType      string `json:"f_var_type" gorm:"column:f_var_type"`
	Description  string `json:"f_description" gorm:"column:f_description"`
}

// DagInstanceKeyword dag instance keyword table model
type DagInstanceKeyword struct {
	ID       uint64 `json:"f_id" gorm:"column:f_id;primaryKey"`
	DagInsID uint64 `json:"f_dag_ins_id" gorm:"column:f_dag_ins_id"`
	Keyword  string `json:"f_keyword" gorm:"column:f_keyword"`
}

type DagStepIndex struct {
	ID            uint64 `gorm:"column:f_id;primaryKey"`
	DagID         uint64 `gorm:"column:f_dag_id"`
	Operator      string `gorm:"column:f_operator"`
	SourceID      string `gorm:"column:f_source_id"`
	HasDatasource bool   `gorm:"column:f_has_datasource"`
}

type DagTriggerConfigIndex struct {
	ID       uint64 `gorm:"column:f_id;primaryKey"`
	DagID    uint64 `gorm:"column:f_dag_id"`
	Operator string `gorm:"column:f_operator"`
	SourceID string `gorm:"column:f_source_id"`
}

type DagAccessorIndex struct {
	ID         uint64 `gorm:"column:f_id;primaryKey"`
	DagID      uint64 `gorm:"column:f_dag_id"`
	AccessorID string `gorm:"column:f_accessor_id"`
}

type DagVersion struct {
	ID        uint64 `json:"f_id" gorm:"column:f_id;primaryKey"`
	CreatedAt int64  `json:"f_created_at" gorm:"column:f_created_at"`
	UpdatedAt int64  `json:"f_updated_at" gorm:"column:f_updated_at"`
	DagID     string `json:"f_dag_id" gorm:"column:f_dag_id"`
	UserID    string `json:"f_user_id" gorm:"column:f_user_id"`
	Version   string `json:"f_version" gorm:"column:f_version"`
	VersionID string `json:"f_version_id" gorm:"column:f_version_id"`
	ChangeLog string `json:"f_change_log" gorm:"column:f_change_log"`
	Config    string `json:"f_config" gorm:"column:f_config"`
	SortTime  int64  `json:"f_sort_time" gorm:"column:f_sort_time"`
}

// DagInstance 对应数据库表 t_dag_instance
type DagInstance struct {
	ID               uint64 `gorm:"column:f_id;primaryKey" json:"f_id"`
	CreatedAt        int64  `gorm:"column:f_created_at" json:"f_created_at"`
	UpdatedAt        int64  `gorm:"column:f_updated_at" json:"f_updated_at"`
	DagID            uint64 `gorm:"column:f_dag_id" json:"f_dag_id"`
	Trigger          string `gorm:"column:f_trigger" json:"f_trigger,omitempty"`
	Worker           string `gorm:"column:f_worker" json:"f_worker"`
	Source           string `gorm:"column:f_source" json:"f_source"`
	Vars             string `gorm:"column:f_vars" json:"f_vars,omitempty"`
	Keywords         string `gorm:"column:f_keywords" json:"f_keywords,omitempty"`
	EventPersistence int    `gorm:"column:f_event_persistence" json:"f_event_persistence,omitempty"`
	EventOssPath     string `gorm:"column:f_event_oss_path" json:"f_event_oss_path,omitempty"`
	ShareData        string `gorm:"column:f_share_data" json:"f_share_data,omitempty"`
	ShareDataExt     string `gorm:"column:f_share_data_ext" json:"f_share_data_ext,omitempty"`
	Status           string `gorm:"column:f_status" json:"f_status"`
	Reason           string `gorm:"column:f_reason" json:"f_reason,omitempty"`
	Cmd              string `gorm:"column:f_cmd" json:"f_cmd,omitempty"`
	HasCmd           bool   `gorm:"column:f_has_cmd" json:"f_has_cmd"`
	BatchRunID       string `gorm:"column:f_batch_run_id" json:"f_batch_run_id"`
	UserID           string `gorm:"column:f_user_id" json:"f_user_id"`
	EndedAt          int64  `gorm:"column:f_ended_at" json:"f_ended_at"`
	DagType          string `gorm:"column:f_dag_type" json:"f_dag_type"`
	PolicyType       string `gorm:"column:f_policy_type" json:"f_policy_type"`
	AppInfo          string `gorm:"column:f_appinfo" json:"f_app_info,omitempty"`
	Priority         string `gorm:"column:f_priority" json:"f_priority"`
	Mode             int    `gorm:"column:f_mode" json:"f_mode"`
	Dump             string `gorm:"column:f_dump" json:"f_dump,omitempty"`
	DumpExt          string `gorm:"column:f_dump_ext" json:"f_dump_ext,omitempty"`
	SuccessCallback  string `gorm:"column:f_success_callback" json:"f_success_callback,omitempty"`
	ErrorCallback    string `gorm:"column:f_error_callback" json:"f_error_callback,omitempty"`
	CallChain        string `gorm:"column:f_call_chain" json:"f_call_chain,omitempty"`
	ResumeData       string `gorm:"column:f_resume_data" json:"f_resume_data,omitempty"`
	ResumeStatus     string `gorm:"column:f_resume_status" json:"f_resume_status"`
	Version          string `gorm:"column:f_version" json:"f_version,omitempty"`
	VersionID        string `gorm:"column:f_version_id" json:"f_version_id"`
	BizDomainID      string `gorm:"column:f_biz_domain_id" json:"f_biz_domain_id"`
}

type InBox struct {
	ID        uint64 `gorm:"column:f_id;primaryKey" json:"f_id"`
	CreatedAt int64  `gorm:"column:f_created_at" json:"f_created_at"`
	UpdatedAt int64  `gorm:"column:f_updated_at" json:"f_updated_at"`
	Msg       string `gorm:"column:f_msg" json:"f_msg"`
	Topic     string `gorm:"column:f_topic" json:"f_topic"`
	DocID     string `gorm:"column:f_docid" json:"f_doc_id"`
	Dags      string `gorm:"column:f_dag" json:"f_dags"`
}

type TaskInstanceModel struct {
	ID             uint64 `gorm:"column:f_id;primaryKey" json:"f_id"`
	CreatedAt      int64  `gorm:"column:f_created_at" json:"f_created_at"`
	UpdatedAt      int64  `gorm:"column:f_updated_at" json:"f_updated_at"`
	TaskID         string `gorm:"column:f_task_id" json:"f_task_id"`
	DagInsID       uint64 `gorm:"column:f_dag_ins_id" json:"f_dag_ins_id"`
	Name           string `gorm:"column:f_name" json:"f_name"`
	DependOn       string `gorm:"column:f_depend_on" json:"f_depend_on"`
	ActionName     string `gorm:"column:f_action_name" json:"f_action_name"`
	TimeoutSecs    int    `gorm:"column:f_timeout_secs" json:"f_timeout_secs"`
	Params         string `gorm:"column:f_params" json:"f_params"`
	Traces         string `gorm:"column:f_traces" json:"f_traces"`
	Status         string `gorm:"column:f_status" json:"f_status"`
	Reason         string `gorm:"column:f_reason" json:"f_reason"`
	PreChecks      string `gorm:"column:f_pre_checks" json:"f_pre_checks"`
	Results        string `gorm:"column:f_results" json:"f_results"`
	Steps          string `gorm:"column:f_steps" json:"f_steps"`
	LastModifiedAt int64  `gorm:"column:f_last_modified_at" json:"f_last_modified_at"`
	RenderedParams string `gorm:"column:f_rendered_params" json:"f_rendered_params"`
	Hash           string `gorm:"column:f_hash" json:"f_hash"`
	Settings       string `gorm:"column:f_settings" json:"f_settings"`
	MetaData       string `gorm:"column:f_metadata" json:"f_metadata"`
}

type TokenModel struct {
	ID           uint64 `gorm:"column:f_id;primaryKey" json:"f_id"`
	CreatedAt    int64  `gorm:"column:f_created_at" json:"f_created_at"`
	UpdatedAt    int64  `gorm:"column:f_updated_at" json:"f_updated_at"`
	UserID       string `gorm:"column:f_user_id" json:"f_user_id"`
	UserName     string `gorm:"column:f_user_name" json:"f_user_name"`
	RefreshToken string `gorm:"column:f_refresh_token" json:"f_refresh_token"`
	Token        string `gorm:"column:f_token" json:"f_token"`
	ExpiresIn    int    `gorm:"column:f_expires_in" json:"f_expires_in"`
	LoginIP      string `gorm:"column:f_login_ip" json:"f_login_ip"`
	IsApp        bool   `gorm:"column:f_is_app" json:"f_is_app"`
}

type ClientModel struct {
	ID           uint64 `gorm:"column:f_id;primaryKey" json:"f_id"`
	CreatedAt    int64  `gorm:"column:f_created_at" json:"f_created_at"`
	UpdatedAt    int64  `gorm:"column:f_updated_at" json:"f_updated_at"`
	ClientName   string `gorm:"column:f_client_name" json:"f_client_name"`
	ClientID     string `gorm:"column:f_client_id" json:"f_client_id"`
	ClientSecret string `gorm:"column:f_client_secret" json:"f_client_secret"`
}

type SwitchModel struct {
	ID        uint64 `gorm:"column:f_id;primaryKey" json:"f_id"`
	CreatedAt int64  `gorm:"column:f_created_at" json:"f_created_at"`
	UpdatedAt int64  `gorm:"column:f_updated_at" json:"f_updated_at"`
	Name      string `gorm:"column:f_name" json:"f_name"`
	Status    bool   `gorm:"column:f_status" json:"f_status"`
}

type LogModel struct {
	ID        uint64 `gorm:"column:f_id;primaryKey" json:"f_id"`
	CreatedAt int64  `gorm:"column:f_created_at" json:"f_created_at"`
	UpdatedAt int64  `gorm:"column:f_updated_at" json:"f_updated_at"`
	OssID     string `gorm:"column:f_oss_id" json:"f_oss_id"`
	Key       string `gorm:"column:f_key" json:"f_key"`
	FileName  string `gorm:"column:f_file_name" json:"f_file_name"`
}

type Closer interface {
	Close()
}

type DagRepository interface {
	CreateDag(ctx context.Context, dag *entity.Dag) (string, error)
	CreateDagIns(ctx context.Context, dagIns *entity.DagInstance) (string, error)
	CreateDagVars(ctx context.Context, dagVars []*DagVar) error
	UpdateDag(ctx context.Context, dag *entity.Dag) error
	GetDag(ctx context.Context, dagId string) (*entity.Dag, error)
	GetDagByFields(ctx context.Context, params map[string]interface{}) (*entity.Dag, error)
	GetDagWithOptionalVersion(ctx context.Context, dagID, versionID string) (*entity.Dag, error)
	ListDagInstance(ctx context.Context, input *mod.ListDagInstanceInput) ([]*entity.DagInstance, error)

	DeleteDag(ctx context.Context, id ...string) error
	CreateDagVersion(ctx context.Context, dagVersion *entity.DagVersion) (string, error)
	GetHistoryDagByVersionID(ctx context.Context, dagID, versionID string) (*entity.DagVersion, error)

	Closer
	WithTransaction(ctx context.Context, fn func(context.Context, mod.Store) error) error
	CreateToken(token *entity.Token) error
	UpdateToken(token *entity.Token) error
	DeleteToken(id string) error
	GetTokenByUserID(userID string) (*entity.Token, error)
	CreateClient(clientName, clientID, clientSecret string) error
	GetClient(clientName string) (client *entity.Client, err error)
	RemoveClient(clientName string) (err error)
	BatchCreateDag(ctx context.Context, dags []*entity.Dag) ([]*entity.Dag, error)
	BatchCreateDagIns(ctx context.Context, dagIns []*entity.DagInstance) ([]*entity.DagInstance, error)
	BatchDeleteDagIns(ctx context.Context, ids []string) error
	CreateTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error
	BatchCreateTaskIns(ctx context.Context, taskIns []*entity.TaskInstance) ([]*entity.TaskInstance, error)
	PatchTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error
	PatchDagIns(ctx context.Context, dagIns *entity.DagInstance, mustsPatchFields ...string) error
	UpdateDagIncValue(ctx context.Context, dagId string, incKey string, incValue any) error
	UpdateDagIns(ctx context.Context, dagIns *entity.DagInstance) error
	UpdateTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error
	BatchUpdateDagIns(ctx context.Context, dagIns []*entity.DagInstance) error
	BatchUpdateTaskIns(taskIns []*entity.TaskInstance) error
	BatchDeleteTaskIns(ctx context.Context, ids []string) error
	GetTaskIns(ctx context.Context, taskIns string) (*entity.TaskInstance, error)
	GetDagInstance(ctx context.Context, dagInsId string) (*entity.DagInstance, error)
	GetDagInstanceByFields(ctx context.Context, params map[string]interface{}) (*entity.DagInstance, error)
	ListDag(ctx context.Context, input *mod.ListDagInput) ([]*entity.Dag, error)
	ListDagByFields(ctx context.Context, filter bson.M, opt options.FindOptions) ([]*entity.Dag, error)
	DisdinctDagInstance(input *mod.ListDagInstanceInput) ([]interface{}, error)
	ListTaskInstance(ctx context.Context, input *mod.ListTaskInstanceInput) ([]*entity.TaskInstance, error)
	Marshal(obj interface{}) ([]byte, error)
	Unmarshal(bytes []byte, ptr interface{}) error
	BatchDeleteDagWithTransaction(ctx context.Context, ids []string) error
	GetDagCount(ctx context.Context, params map[string]interface{}) (int64, error)
	ListDagCount(ctx context.Context, input *mod.ListDagInput) (int64, error)
	ListDagCountByFields(ctx context.Context, filter bson.M) (int64, error)
	GetDagInstanceCount(ctx context.Context, params map[string]interface{}) (int64, error)
	CreateInbox(ctx context.Context, msg *entity.InBox) error
	DeleteInbox(ctx context.Context, ids []string) error
	GetInbox(ctx context.Context, id string) (*entity.InBox, error)
	ListInbox(ctx context.Context, input *mod.ListInboxInput) ([]*entity.InBox, error)
	GetSwitchStatus() (bool, error)
	SetSwitchStatus(status bool) error
	CreateLogs(ctx context.Context, ossLogs []*entity.Log) error
	ListHistoryDagIns(ctx context.Context, params map[string]interface{}, dataChannel chan []bson.M) error
	ListHistoryTaskIns(ctx context.Context, params map[string]interface{}, dataChannel chan []bson.M) error
	DeleteDagInsByID(ctx context.Context, params map[string]interface{}) error
	DeleteTaskInsByID(ctx context.Context, params map[string]interface{}) error
	DeleteTaskInsByDagInsID(ctx context.Context, dagInsID string) error
	GetTaskInstanceCount(ctx context.Context, params map[string]interface{}) (int64, error)
	CreatOutBoxMessage(ctx context.Context, outBox *entity.OutBox) error
	BatchCreatOutBoxMessage(ctx context.Context, outBox []*entity.OutBox) error
	DeleteOutBoxMessage(ctx context.Context, ids []string) error
	ListOutBoxMessage(ctx context.Context, input *entity.OutBoxInput) ([]*entity.OutBox, error)
	ListDagInstanceInRangeTime(ctx context.Context, status string, begin, end int64) ([]*entity.DagInstance, error)
	ListExistDagInsID(ctx context.Context, dagInsIDs []string) ([]string, error)
	ListExistDagID(ctx context.Context, dagIDs []string) ([]string, error)
	GroupDagInstance(ctx context.Context, input *mod.GroupInput) ([]*entity.DagInstanceGroup, error)
	RetryDagIns(ctx context.Context, dagInsID string, taskInsIDs []string) error

	// DeleteDag 删除Dag配置,仅在组合算子注册失败时，删除dag配置时使用
	ListDagVersions(ctx context.Context, input *mod.ListDagVersionInput) ([]entity.DagVersion, error)

	// 事务操作
	// WithTransaction(ctx context.Context, fn func(DagRepository) error) error
}

type dag struct {
	db   *gorm.DB
	isTX bool
}

var (
	dagOnce sync.Once
	dagRep  DagRepository
)

func NewDagRepository() DagRepository {
	dagOnce.Do(func() {
		dagRep = &dag{
			db: db.NewDB().Debug(),
		}
	})
	return dagRep
}

func ToDagModel(dag *entity.Dag, isupdate bool) *Dag {
	if isupdate {
		dag.Update()
	} else {
		dag.Initial()
	}
	id, _ := strconv.ParseUint(dag.ID, 10, 64)
	dagVarBytes, _ := json.Marshal(dag.Vars)
	tasksBytes, _ := json.Marshal(dag.Tasks)
	stepsBytes, _ := json.Marshal(dag.Steps)
	shortcutsBytes, _ := json.Marshal(dag.Shortcuts)
	accessorBytes, _ := json.Marshal(dag.Accessors)
	appInfoBytes, _ := json.Marshal(dag.AppInfo)
	emailsBytes, _ := json.Marshal(dag.Emails)
	triggerConfigBytes, _ := json.Marshal(dag.TriggerConfig)
	subIDsBytes, _ := json.Marshal(dag.SubIDs)
	outputsBytes, _ := json.Marshal(dag.OutPuts)
	instructionsBytes, _ := json.Marshal(dag.Instructions)
	incValuesBytes, _ := json.Marshal(dag.IncValues)

	return &Dag{
		ID:            id,
		CreatedAt:     dag.CreatedAt,
		UpdatedAt:     dag.UpdatedAt,
		UserID:        dag.UserID,
		Name:          dag.Name,
		Desc:          dag.Desc,
		Trigger:       string(dag.Trigger),
		Cron:          dag.Cron,
		Vars:          string(dagVarBytes),
		Status:        string(dag.Status),
		Tasks:         string(tasksBytes),
		Steps:         string(stepsBytes),
		Description:   dag.Description,
		Shortcuts:     string(shortcutsBytes),
		Accessors:     string(accessorBytes),
		Type:          dag.Type,
		PolicyType:    dag.PolicyType,
		AppInfo:       string(appInfoBytes),
		Priority:      dag.Priority,
		Emails:        string(emailsBytes),
		Template:      dag.Template,
		Published:     dag.Published,
		TriggerConfig: string(triggerConfigBytes),
		SubIDs:        string(subIDsBytes),
		ExecMode:      dag.ExecMode,
		Category:      dag.Category,
		OutPuts:       string(outputsBytes),
		Instructions:  string(instructionsBytes),
		OperatorID:    dag.OperatorID,
		IncValues:     string(incValuesBytes),
		Version:       dag.Version.ToString(),
		VersionID:     dag.VersionID,
		ModifyBy:      dag.ModifyBy,
		IsDebug:       dag.IsDebug,
		DeBugID:       dag.DeBugID,
		BizDomainID:   dag.BizDomainID,
	}
}

func ToDagInstanceModel(dagIns *entity.DagInstance, isupdate bool) *DagInstance {
	if isupdate {
		dagIns.Update()
	} else {
		dagIns.Initial()
	}

	id, _ := strconv.ParseUint(dagIns.ID, 10, 64)
	dagInsID, _ := strconv.ParseUint(dagIns.DagID, 10, 64)
	varsBytes, _ := json.Marshal(dagIns.Vars)
	keywordsBytes, _ := json.Marshal(dagIns.Keywords)
	sharedataBytes, _ := json.Marshal(dagIns.ShareData)
	sharedataextBytes, _ := json.Marshal(dagIns.ShareDataExt)
	cmdBytes, _ := json.Marshal(dagIns.Cmd)
	appInfoBytes, _ := json.Marshal(dagIns.AppInfo)
	dumpextbytes, _ := json.Marshal(dagIns.DumpExt)
	callchainBytes, _ := json.Marshal(dagIns.CallChain)
	hasCmd := dagIns.Cmd != nil && !reflect.DeepEqual(*dagIns.Cmd, entity.Command{})
	batchRunID := ""
	if val, ok := dagIns.Vars["batch_run_id"]; ok {
		batchRunID = val.Value
	}

	return &DagInstance{
		ID:               id,
		CreatedAt:        dagIns.CreatedAt,
		UpdatedAt:        dagIns.UpdatedAt,
		DagID:            dagInsID,
		Trigger:          string(dagIns.Trigger),
		Worker:           dagIns.Worker,
		Source:           dagIns.Source,
		Vars:             string(varsBytes),
		Keywords:         string(keywordsBytes),
		EventPersistence: int(dagIns.EventPersistence),
		EventOssPath:     dagIns.EventOssPath,
		ShareData:        string(sharedataBytes),
		ShareDataExt:     string(sharedataextBytes),
		Status:           string(dagIns.Status),
		Reason:           dagIns.Reason,
		Cmd:              string(cmdBytes),
		HasCmd:           hasCmd,
		BatchRunID:       batchRunID,
		UserID:           dagIns.UserID,
		EndedAt:          dagIns.EndedAt,
		DagType:          dagIns.DagType,
		PolicyType:       dagIns.PolicyType,
		AppInfo:          string(appInfoBytes),
		Priority:         dagIns.Priority,
		Mode:             int(dagIns.Mode),
		Dump:             dagIns.Dump,
		DumpExt:          string(dumpextbytes),
		SuccessCallback:  dagIns.SuccessCallback,
		ErrorCallback:    dagIns.ErrorCallback,
		CallChain:        string(callchainBytes),
		ResumeData:       dagIns.ResumeData,
		ResumeStatus:     string(dagIns.ResumeStatus),
		Version:          dagIns.Version.ToString(),
		VersionID:        dagIns.VersionID,
		BizDomainID:      dagIns.BizDomainID,
	}
}

func ToTaskInstanceModel(taskIns *entity.TaskInstance, isupdate bool) *TaskInstanceModel {
	if isupdate {
		taskIns.Update()
	} else {
		taskIns.Initial()
	}

	id, _ := strconv.ParseUint(taskIns.ID, 10, 64)
	dagInsID, _ := strconv.ParseUint(taskIns.DagInsID, 10, 64)

	return &TaskInstanceModel{
		ID:             id,
		CreatedAt:      taskIns.CreatedAt,
		UpdatedAt:      taskIns.UpdatedAt,
		TaskID:         taskIns.TaskID,
		DagInsID:       dagInsID,
		Name:           taskIns.Name,
		DependOn:       marshalToString(taskIns.DependOn),
		ActionName:     taskIns.ActionName,
		TimeoutSecs:    taskIns.TimeoutSecs,
		Params:         marshalToString(taskIns.Params),
		Traces:         marshalToString(taskIns.Traces),
		Status:         string(taskIns.Status),
		Reason:         marshalToString(taskIns.Reason),
		PreChecks:      marshalToString(taskIns.PreChecks),
		Results:        marshalToString(taskIns.Results),
		Steps:          marshalToString(taskIns.Steps),
		LastModifiedAt: taskIns.LastModifiedAt,
		RenderedParams: marshalToString(taskIns.RenderedParams),
		Hash:           taskIns.Hash,
		Settings:       marshalToString(taskIns.Settings),
		MetaData:       marshalToString(taskIns.MetaData),
	}
}

func ToEntity(src, dest interface{}) error {
	return copyFields(src, dest)
}

func marshalToString(val interface{}) string {
	if val == nil {
		return ""
	}
	bytes, err := json.Marshal(val)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func copyFields(src interface{}, dest interface{}) error {
	srcVal := reflect.ValueOf(src)
	destVal := reflect.ValueOf(dest)

	if srcVal.Kind() != reflect.Ptr || destVal.Kind() != reflect.Ptr {
		return fmt.Errorf("both src and dest must be pointers")
	}

	srcVal = srcVal.Elem()
	destVal = destVal.Elem()

	if srcVal.Kind() != reflect.Struct || destVal.Kind() != reflect.Struct {
		return fmt.Errorf("both src and dest must be structs")
	}

	srcType := srcVal.Type()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		srcFieldType := srcType.Field(i)

		destField := destVal.FieldByName(srcFieldType.Name)
		if !destField.IsValid() || !destField.CanSet() {
			continue
		}

		// 1. 类型完全一致，直接赋值
		if srcField.Type().AssignableTo(destField.Type()) {
			destField.Set(srcField)
			continue
		}

		// 2. 数值 → 字符串类型（包括自定义字符串类型如 type MyStr string）
		if isNumeric(srcField) && isStringKind(destField) {
			strVal := numericToString(srcField)
			destField.Set(reflect.ValueOf(strVal).Convert(destField.Type()))
			continue
		}

		// 3. 字符串类型 → 数值（如 "12345" → uint64）
		if isStringKind(srcField) && isNumeric(destField) {
			if converted, ok := stringToNumeric(srcField.String(), destField.Type()); ok {
				destField.Set(converted)
				continue
			}
		}

		// 4. 底层类型相同的转换（string ↔ MyStr, int ↔ MyInt 等，但排除整数→string的误转换）
		if safeConvertible(srcField.Type(), destField.Type()) {
			destField.Set(srcField.Convert(destField.Type()))
			continue
		}

		// 5. 指针与非指针之间的转换
		if handlePtrConversion(srcField, destField) {
			continue
		}

		// 6. 字符串 → 复杂类型，尝试 JSON 反序列化
		if isStringKind(srcField) && srcField.String() != "" {
			strValue := srcField.String()
			destFieldType := destField.Type()

			var destInstancePtr reflect.Value
			var isPtr bool

			if destFieldType.Kind() == reflect.Ptr {
				destInstancePtr = reflect.New(destFieldType.Elem())
				isPtr = true
			} else {
				destInstancePtr = reflect.New(destFieldType)
				isPtr = false
			}

			if err := json.Unmarshal([]byte(strValue), destInstancePtr.Interface()); err == nil {
				if isPtr {
					destField.Set(destInstancePtr)
				} else {
					destField.Set(destInstancePtr.Elem())
				}
			}
		}
	}

	return nil
}

// ======================== 辅助函数 ========================

// isNumeric 判断是否为数值类型（包括自定义数值类型如 type MyInt int）
func isNumeric(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// isStringKind 判断底层是否为字符串类型（包括 type MyStr string）
func isStringKind(v reflect.Value) bool {
	return v.Kind() == reflect.String
}

// numericToString 将数值转为字符串
func numericToString(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	}
	return fmt.Sprintf("%v", v.Interface())
}

// stringToNumeric 将字符串转为目标数值类型
func stringToNumeric(s string, targetType reflect.Type) (reflect.Value, bool) {
	// 获取底层 Kind
	kind := targetType.Kind()

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return reflect.Value{}, false
		}
		val := reflect.New(targetType).Elem()
		val.SetInt(n)
		return val, true

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return reflect.Value{}, false
		}
		val := reflect.New(targetType).Elem()
		val.SetUint(n)
		return val, true

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return reflect.Value{}, false
		}
		val := reflect.New(targetType).Elem()
		val.SetFloat(n)
		return val, true

	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return reflect.Value{}, false
		}
		val := reflect.New(targetType).Elem()
		val.SetBool(b)
		return val, true
	}

	return reflect.Value{}, false
}

// safeConvertible 安全的类型转换判断，排除整数→字符串的 Unicode 码点误转换
func safeConvertible(srcType, destType reflect.Type) bool {
	if !srcType.ConvertibleTo(destType) {
		return false
	}

	srcKind := srcType.Kind()
	destKind := destType.Kind()

	// 排除: 整数 → 字符串（Go 会把整数当 rune 转换，不是我们要的行为）
	if isIntKind(srcKind) && destKind == reflect.String {
		return false
	}
	// 排除: 字符串 → 整数（Convert 本身也不支持，但以防万一）
	if srcKind == reflect.String && isIntKind(destKind) {
		return false
	}

	return true
}

func isIntKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}

// handlePtrConversion 处理指针与非指针之间的转换
func handlePtrConversion(srcField, destField reflect.Value) bool {
	srcType := srcField.Type()
	destType := destField.Type()

	// 非指针 → 指针
	if srcType.Kind() != reflect.Ptr && destType.Kind() == reflect.Ptr {
		elemType := destType.Elem()
		if safeConvertible(srcType, elemType) {
			newVal := reflect.New(elemType)
			newVal.Elem().Set(srcField.Convert(elemType))
			destField.Set(newVal)
			return true
		}
	}

	// 指针 → 非指针
	if srcType.Kind() == reflect.Ptr && destType.Kind() != reflect.Ptr {
		if !srcField.IsNil() {
			srcElem := srcField.Elem()
			if safeConvertible(srcElem.Type(), destType) {
				destField.Set(srcElem.Convert(destType))
				return true
			}
		}
	}

	// 指针 → 指针
	if srcType.Kind() == reflect.Ptr && destType.Kind() == reflect.Ptr {
		if !srcField.IsNil() {
			srcElem := srcField.Elem()
			destElemType := destType.Elem()
			if safeConvertible(srcElem.Type(), destElemType) {
				newVal := reflect.New(destElemType)
				newVal.Elem().Set(srcElem.Convert(destElemType))
				destField.Set(newVal)
				return true
			}
		}
	}

	return false
}

func updateJSONMapString(raw, key string, val any) (string, error) {
	m := map[string]any{}
	if raw != "" {
		if err := jsoniter.UnmarshalFromString(raw, &m); err != nil {
			return "", err
		}
	}
	m[key] = val
	bytes, err := jsoniter.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func parseUint64Slice(ids []string) []uint64 {
	var res []uint64
	for _, id := range ids {
		v, _ := strconv.ParseUint(id, 10, 64)
		res = append(res, v)
	}
	return res
}

func toInt64(val interface{}) int64 {
	switch v := val.(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case []byte:
		i, _ := strconv.ParseInt(string(v), 10, 64)
		return i
	case string:
		i, _ := strconv.ParseInt(v, 10, 64)
		return i
	default:
		return 0
	}
}

func ToDagVersionModel(dagVersion *entity.DagVersion) *DagVersion {
	dagVersion.Initial()
	id, _ := strconv.ParseUint(dagVersion.ID, 10, 64)

	return &DagVersion{
		ID:        id,
		CreatedAt: dagVersion.CreatedAt,
		UpdatedAt: dagVersion.UpdatedAt,
		DagID:     dagVersion.DagID,
		UserID:    dagVersion.UserID,
		Version:   dagVersion.Version.ToString(),
		VersionID: dagVersion.VersionID,
		ChangeLog: dagVersion.ChangeLog,
		Config:    string(dagVersion.Config),
		SortTime:  0,
	}
}

// TransactionWithContext 带 Context 的事务
func (d *dag) WithTransaction(ctx context.Context, fn func(context.Context, mod.Store) error) error {
	txCli := &dag{
		db:   d.db.Begin(),
		isTX: true,
	}

	tx := txCli.db
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := fn(ctx, txCli); err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("rollback failed: %s, original: %s", rbErr.Error(), err.Error())
		}
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit failed: %s", err.Error())
	}

	return nil
}

func (d *dag) CreateDag(ctx context.Context, dag *entity.Dag) (string, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	fn := func(dag *entity.Dag) error {
		// 准备 SQL 语句，使用参数化查询防止 SQL 注入
		sql := `INSERT INTO t_dag (
			f_id, f_created_at, f_updated_at, f_user_id, f_name, f_desc, f_trigger,
			f_cron, f_vars, f_status, f_tasks, f_steps, f_description, f_shortcuts,
			f_accessors, f_type, f_policy_type, f_appinfo, f_priority, f_removed,
			f_emails, f_template, f_published, f_trigger_config, f_sub_ids, f_exec_mode,
			f_category, f_outputs, f_instructions, f_operator_id, f_inc_values, f_version,
			f_version_id, f_modify_by, f_is_debug, f_debug_id, f_biz_domain_id
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		// 执行 SQL 语句
		t := ToDagModel(dag, false)
		msgStr, _ := jsoniter.MarshalToString(t)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

		err = dag.CheckRootNode(dag.Tasks)
		if err != nil {
			return err
		}

		err = d.db.Exec(sql,
			t.ID,
			t.CreatedAt,
			t.UpdatedAt,
			t.UserID,
			t.Name,
			t.Desc,
			t.Trigger,
			t.Cron,
			t.Vars,
			t.Status,
			t.Tasks,
			t.Steps,
			t.Description,
			t.Shortcuts,
			t.Accessors,
			t.Type,
			t.PolicyType,
			t.AppInfo,
			t.Priority,
			t.Removed,
			t.Emails,
			t.Template,
			t.Published,
			t.TriggerConfig,
			t.SubIDs,
			t.ExecMode,
			t.Category,
			t.OutPuts,
			t.Instructions,
			t.OperatorID,
			t.IncValues,
			t.Version,
			t.VersionID,
			t.ModifyBy,
			t.IsDebug,
			t.DeBugID,
			t.BizDomainID,
		).Error
		if err != nil {
			return err
		}

		err = d.CreateDagVars(newCtx, BuildDagVars(dag))
		if err != nil {
			return err
		}

		err = d.refreshDagIndexes(newCtx, dag)
		if err != nil {
			return err
		}

		return nil
	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(context.Context, mod.Store) error {
			return fn(dag)
		})
	} else {
		err = fn(dag)
	}

	return dag.ID, err
}

// func (d *dag) BatchCreateDag(ctx context.Context, dags []*entity.Dag) ([]*entity.Dag, error) {
// 	var err error
// 	newCtx, span := trace.StartInternalSpan(ctx)
// 	msgStr, _ := jsoniter.MarshalToString(dags)
// 	defer func() { trace.TelemetrySpanEnd(span, err) }()
// }

func (d *dag) CreateDagVars(ctx context.Context, dagVars []*DagVar) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	msgStr, _ := jsoniter.MarshalToString(dagVars)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	fn := func(dagVars []*DagVar) error {
		if len(dagVars) == 0 {
			return nil
		}

		dagID := dagVars[0].DagID
		sqlStr := `DELETE FROM t_dag_vars WHERE f_dag_id = ?`
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGVAR_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", dagID)))
		err = d.db.Exec(sqlStr, dagID).Error
		if err != nil {
			return err
		}

		sqlStr = `INSERT INTO t_dag_vars (f_id, f_dag_id, f_var_name, f_default_value, f_var_type) VALUES `
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGVAR_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, msgStr))
		values := make([]any, 0, len(dagVars)*5)
		for _, data := range dagVars {
			sqlStr += "(?, ?, ?, ?, ?),"
			values = append(values, data.ID, data.DagID, data.VarName, data.DefaultValue, data.VarType)
		}

		sqlStr = sqlStr[:len(sqlStr)-1]

		err = d.db.Exec(sqlStr, values...).Error
		if err != nil {
			return err
		}

		return nil

	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(context.Context, mod.Store) error {
			return fn(dagVars)
		})
	} else {
		err = fn(dagVars)
	}

	return err
}

func (d *dag) deleteDagInstanceKeywords(ctx context.Context, dagInsID uint64) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sqlStr := `DELETE FROM t_dag_instance_keyword WHERE f_dag_ins_id = ?`
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCEKEYWORD_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", dagInsID)))
	err = d.db.Exec(sqlStr, dagInsID).Error
	return err
}

func (d *dag) insertDagInstanceKeywords(ctx context.Context, dagInsID uint64, keywords []string) error {
	var err error
	if len(keywords) == 0 {
		return nil
	}
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sqlStr := `INSERT INTO t_dag_instance_keyword (f_id, f_dag_ins_id, f_keyword) VALUES `
	values := make([]any, 0, len(keywords)*3)
	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}
		id, _ := utils.GetUniqueID()
		sqlStr += "(?, ?, ?),"
		values = append(values, id, dagInsID, keyword)
	}
	if len(values) == 0 {
		return nil
	}
	sqlStr = strings.TrimSuffix(sqlStr, ",")
	msgStr, _ := jsoniter.MarshalToString(values)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCEKEYWORD_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, msgStr))
	err = d.db.Exec(sqlStr, values...).Error
	return err
}

func (d *dag) replaceDagInstanceKeywords(ctx context.Context, dagInsID uint64, keywords []string) error {
	if err := d.deleteDagInstanceKeywords(ctx, dagInsID); err != nil {
		return err
	}
	return d.insertDagInstanceKeywords(ctx, dagInsID, keywords)
}

func (d *dag) refreshDagIndexes(ctx context.Context, dag *entity.Dag) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if dag == nil {
		return nil
	}

	dagID, parseErr := strconv.ParseUint(dag.ID, 10, 64)
	if parseErr != nil {
		return parseErr
	}

	stepRows := BuildDagStepIndex(dag)
	triggerRows := BuildDagTriggerConfigIndex(dag)
	accessorRows := BuildDagAccessorIndex(dag)

	deleteIndexes := func(table string) error {
		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_dag_id = ?", table)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, table), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", dagID)))
		return d.db.Exec(sqlStr, dagID).Error
	}

	if err = deleteIndexes(DAGSTEPINDEX_TABLENAME); err != nil {
		return err
	}
	if err = deleteIndexes(DAGTRIGGERINDEX_TABLENAME); err != nil {
		return err
	}
	if err = deleteIndexes(DAGACCESSORINDEX_TABLENAME); err != nil {
		return err
	}

	insertStepRows := func(rows []*DagStepIndex) error {
		if len(rows) == 0 {
			return nil
		}
		sqlStr := fmt.Sprintf("INSERT INTO %s (f_id, f_dag_id, f_operator, f_source_id, f_has_datasource) VALUES ", DAGSTEPINDEX_TABLENAME)
		values := make([]any, 0, len(rows)*5)
		for _, row := range rows {
			sqlStr += "(?, ?, ?, ?, ?),"
			values = append(values, row.ID, row.DagID, row.Operator, row.SourceID, row.HasDatasource)
		}
		sqlStr = strings.TrimSuffix(sqlStr, ",")
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGSTEPINDEX_TABLENAME), attribute.String(trace.DB_SQL, sqlStr))
		return d.db.Exec(sqlStr, values...).Error
	}

	insertTriggerRows := func(rows []*DagTriggerConfigIndex) error {
		if len(rows) == 0 {
			return nil
		}
		sqlStr := fmt.Sprintf("INSERT INTO %s (f_id, f_dag_id, f_operator, f_source_id) VALUES ", DAGTRIGGERINDEX_TABLENAME)
		values := make([]any, 0, len(rows)*4)
		for _, row := range rows {
			sqlStr += "(?, ?, ?, ?),"
			values = append(values, row.ID, row.DagID, row.Operator, row.SourceID)
		}
		sqlStr = strings.TrimSuffix(sqlStr, ",")
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGTRIGGERINDEX_TABLENAME), attribute.String(trace.DB_SQL, sqlStr))
		return d.db.Exec(sqlStr, values...).Error
	}

	insertAccessorRows := func(rows []*DagAccessorIndex) error {
		if len(rows) == 0 {
			return nil
		}
		sqlStr := fmt.Sprintf("INSERT INTO %s (f_id, f_dag_id, f_accessor_id) VALUES ", DAGACCESSORINDEX_TABLENAME)
		values := make([]any, 0, len(rows)*3)
		for _, row := range rows {
			sqlStr += "(?, ?, ?),"
			values = append(values, row.ID, row.DagID, row.AccessorID)
		}
		sqlStr = strings.TrimSuffix(sqlStr, ",")
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGACCESSORINDEX_TABLENAME), attribute.String(trace.DB_SQL, sqlStr))
		return d.db.Exec(sqlStr, values...).Error
	}

	if err = insertStepRows(stepRows); err != nil {
		return err
	}
	if err = insertTriggerRows(triggerRows); err != nil {
		return err
	}
	if err = insertAccessorRows(accessorRows); err != nil {
		return err
	}

	return nil
}

func (d *dag) CreateDagVersion(ctx context.Context, dagVersion *entity.DagVersion) (string, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)

	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sqlStr := `INSERT INTO t_dag_versions (
		f_id, f_created_at, f_updated_at, f_dag_id,
		f_user_id, f_version, f_version_id, f_change_log, f_config, f_sort_time)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	t := ToDagVersionModel(dagVersion)
	msgStr, _ := jsoniter.MarshalToString(t)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGVERSIONS_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, msgStr))
	err = d.db.Exec(sqlStr,
		t.ID,
		t.CreatedAt,
		t.UpdatedAt,
		t.DagID,
		t.UserID,
		t.Version,
		t.VersionID,
		t.ChangeLog,
		t.Config,
		t.SortTime,
	).Error
	if err != nil {
		return "", err
	}

	return dagVersion.ID, nil
}

func (d *dag) UpdateDag(ctx context.Context, dag *entity.Dag) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	fn := func(dag *entity.Dag) error {
		// 准备 SQL 语句，使用参数化查询防止 SQL 注入
		sql := `UPDATE t_dag SET
			f_created_at = ?, f_updated_at = ?, f_user_id = ?, f_name = ?, f_desc = ?,
			f_trigger = ?, f_cron = ?, f_vars = ?, f_status = ?, f_tasks = ?, f_steps = ?,
			f_description = ?, f_shortcuts = ?, f_accessors = ?, f_type = ?, f_policy_type = ?,
			f_appinfo = ?, f_priority = ?, f_removed = ?, f_emails = ?, f_template = ?, f_published = ?,
			f_trigger_config = ?, f_sub_ids = ?, f_exec_mode = ?, f_category = ?, f_outputs = ?,
			f_instructions = ?, f_operator_id = ?, f_inc_values = ?, f_version = ?, f_version_id = ?,
			f_modify_by = ?, f_is_debug = ?, f_debug_id = ?, f_biz_domain_id = ?
			WHERE f_id = ?`

		// 执行 SQL 语句
		t := ToDagModel(dag, true)
		msgStr, _ := jsoniter.MarshalToString(t)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

		err = dag.CheckRootNode(dag.Tasks)
		if err != nil {
			return err
		}

		err = d.db.Exec(sql,
			t.CreatedAt,
			t.UpdatedAt,
			t.UserID,
			t.Name,
			t.Desc,
			t.Trigger,
			t.Cron,
			t.Vars,
			t.Status,
			t.Tasks,
			t.Steps,
			t.Description,
			t.Shortcuts,
			t.Accessors,
			t.Type,
			t.PolicyType,
			t.AppInfo,
			t.Priority,
			t.Removed,
			t.Emails,
			t.Template,
			t.Published,
			t.TriggerConfig,
			t.SubIDs,
			t.ExecMode,
			t.Category,
			t.OutPuts,
			t.Instructions,
			t.OperatorID,
			t.IncValues,
			t.Version,
			t.VersionID,
			t.ModifyBy,
			t.IsDebug,
			t.DeBugID,
			t.BizDomainID,
			t.ID,
		).Error
		if err != nil {
			return err
		}

		err = d.CreateDagVars(newCtx, BuildDagVars(dag))
		if err != nil {
			return err
		}

		err = d.refreshDagIndexes(newCtx, dag)
		if err != nil {
			return err
		}

		return nil
	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(context.Context, mod.Store) error {
			return fn(dag)
		})
	} else {
		err = fn(dag)
	}

	return err
}

func (d *dag) GetDag(ctx context.Context, dagId string) (*entity.Dag, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	return d.GetDagByFields(newCtx, map[string]interface{}{"f_id": dagId})
}

func (d *dag) GetDagByFields(ctx context.Context, params map[string]interface{}) (*entity.Dag, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	result, err := NewConverter(DAG_TABLENAME, WithAutoConvert(true)).Convert(params)
	if err != nil {
		return nil, err
	}

	query, _ := jsoniter.MarshalToString(result.Params)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, result.SQL), attribute.String(trace.DB_QUERY, query))

	dag := &Dag{}
	err = d.db.Raw(result.SQL, result.Params...).Scan(dag).Error
	if err != nil {
		return nil, err
	}

	if dag.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	dest := &entity.Dag{}
	err = ToEntity(dag, dest)
	return dest, err
}

func (d *dag) GetDagWithOptionalVersion(ctx context.Context, dagID, versionID string) (*entity.Dag, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	var sql string
	if versionID != "" {
		var config entity.Config
		sql = `SELECT f_config FROM t_dag_versions WHERE f_dag_id = ? AND f_version_id = ?`
		err = d.db.Raw(sql, dagID, versionID).Scan(&config).Error
		if err != nil {
			return nil, err
		}

		return config.ParseToDag()
	} else {
		return d.GetDag(newCtx, dagID)
	}
}

func (d *dag) DeleteDag(ctx context.Context, id ...string) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	msgStr, _ := jsoniter.MarshalToString(id)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	fn := func(ids ...string) error {
		if len(ids) == 0 {
			return nil
		}

		sqlStr := `DELETE FROM t_dag WHERE f_id IN ?`
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, msgStr))

		err = d.db.Exec(sqlStr, ids).Error
		if err != nil {
			return err
		}

		sqlStr = `DELETE FROM t_dag_vars WHERE f_dag_id IN ?`
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGVAR_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, msgStr))
		err = d.db.Exec(sqlStr, ids).Error
		if err != nil {
			return err
		}

		sqlStr = `DELETE FROM t_dag_versions WHERE f_dag_id IN ?`
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGVERSIONS_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, msgStr))
		err = d.db.Exec(sqlStr, ids).Error
		if err != nil {
			return err
		}

		return nil
	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(context.Context, mod.Store) error {
			return fn(id...)
		})
	} else {
		err = fn(id...)
	}

	return err
}

func (d *dag) ListDagInstance(ctx context.Context, input *mod.ListDagInstanceInput) ([]*entity.DagInstance, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sqlStr, args := BuildListDagInstanceQuery(input, false)
	query, _ := jsoniter.MarshalToString(args)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, query))

	dagInstances := make([]*DagInstance, 0)
	err = d.db.Raw(sqlStr, args...).Scan(&dagInstances).Error
	if err != nil {
		return nil, err
	}

	dest := make([]*entity.DagInstance, 0)
	for _, dag := range dagInstances {
		dagIns := &entity.DagInstance{}
		err = ToEntity(dag, dagIns)
		if err != nil {
			return nil, err
		}
		dest = append(dest, dagIns)
	}

	return dest, nil

}

func (d *dag) CreateDagIns(ctx context.Context, dagIns *entity.DagInstance) (string, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sql := `INSERT INTO t_dag_instance (
		f_id, f_created_at, f_updated_at, f_dag_id, f_trigger, f_worker, f_source,
		f_vars, f_keywords, f_event_persistence, f_event_oss_path, f_share_data, f_share_data_ext,
		f_status, f_reason, f_cmd, f_has_cmd, f_batch_run_id, f_user_id, f_ended_at, f_dag_type, f_policy_type, f_appinfo,
		f_priority, f_mode, f_dump, f_dump_ext, f_success_callback, f_error_callback, f_call_chain,
		f_resume_data, f_resume_status, f_version, f_version_id, f_biz_domain_id)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// 执行 SQL 语句
	t := ToDagInstanceModel(dagIns, false)
	msgStr, _ := jsoniter.MarshalToString(t)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

	err = d.db.Exec(sql,
		t.ID,
		t.CreatedAt,
		t.UpdatedAt,
		t.DagID,
		t.Trigger,
		t.Worker,
		t.Source,
		t.Vars,
		t.Keywords,
		t.EventPersistence,
		t.EventOssPath,
		t.ShareData,
		t.ShareDataExt,
		t.Status,
		t.Reason,
		t.Cmd,
		t.HasCmd,
		t.BatchRunID,
		t.UserID,
		t.EndedAt,
		t.DagType,
		t.PolicyType,
		t.AppInfo,
		t.Priority,
		t.Mode,
		t.Dump,
		t.DumpExt,
		t.SuccessCallback,
		t.ErrorCallback,
		t.CallChain,
		t.ResumeData,
		t.ResumeStatus,
		t.Version,
		t.VersionID,
		t.BizDomainID,
	).Error

	if err != nil {
		return "", err
	}

	if err = d.insertDagInstanceKeywords(newCtx, t.ID, dagIns.Keywords); err != nil {
		return "", err
	}

	return dagIns.ID, nil
}

func (d *dag) GetHistoryDagByVersionID(ctx context.Context, dagID, versionID string) (*entity.DagVersion, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sql := `SELECT
		f_id, f_created_at, f_updated_at, f_dag_id, f_user_id, f_version,
		f_version_id, f_change_log, f_config, f_sort_time FROM t_dag_versions
		WHERE f_dag_id = ? AND f_version_id = ?`
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGVERSIONS_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, dagID), attribute.String(trace.DB_QUERY, versionID))

	dagVersion := &DagVersion{}
	err = d.db.Raw(sql, dagID, versionID).Scan(dagVersion).Error
	if err != nil {
		return nil, err
	}

	dest := &entity.DagVersion{}
	err = ToEntity(dagVersion, dest)
	if err != nil {
		return nil, err
	}

	return dest, nil
}

// BatchCreatOutBoxMessage implements [DagRepository].
func (d *dag) BatchCreatOutBoxMessage(ctx context.Context, outBox []*entity.OutBox) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if len(outBox) == 0 {
		return nil
	}

	sqlStr := `INSERT INTO t_outbox (f_id, f_topic, f_msg, f_created_at, f_updated_at) VALUES `

	values := make([]any, 0, len(outBox)*5)
	for _, data := range outBox {
		sqlStr += "(?, ?, ?, ?, ?),"
		values = append(values, data.ID, data.Topic, data.Msg, data.CreatedAt, data.UpdatedAt)
	}

	sqlStr = sqlStr[:len(sqlStr)-1]

	msgStr, _ := jsoniter.MarshalToString(values)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGVAR_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, msgStr))

	err = d.db.Exec(sqlStr, values...).Error

	return err
}

// BatchCreateDag implements [DagRepository].
func (d *dag) BatchCreateDag(ctx context.Context, dags []*entity.Dag) ([]*entity.Dag, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if len(dags) == 0 {
		return dags, nil
	}

	fn := func(ctx context.Context, d mod.Store) error {
		for _, dag := range dags {
			dag.Initial()
			// check task's connection
			_, err = mod.BuildRootNode(mod.MapTasksToGetter(dag.Tasks))
			if err != nil {
				return err
			}

			_, err = d.CreateDag(ctx, dag)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(context.Context, mod.Store) error {
			return fn(newCtx, d)
		})
	} else {
		err = fn(newCtx, d)
	}

	if err != nil {
		return nil, err
	}

	return dags, nil
}

// BatchCreateDagIns implements [DagRepository].
func (d *dag) BatchCreateDagIns(ctx context.Context, dagIns []*entity.DagInstance) ([]*entity.DagInstance, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sqlStr := `INSERT INTO t_dag_instance (
		f_id, f_created_at, f_updated_at, f_dag_id, f_trigger, f_worker, f_source,
		f_vars, f_keywords, f_event_persistence, f_event_oss_path, f_share_data, f_share_data_ext,
		f_status, f_reason, f_cmd, f_has_cmd, f_batch_run_id, f_user_id, f_ended_at, f_dag_type, f_policy_type, f_appinfo,
		f_priority, f_mode, f_dump, f_dump_ext, f_success_callback, f_error_callback, f_call_chain,
		f_resume_data, f_resume_status, f_version, f_version_id, f_biz_domain_id)
		VALUES `

	values := make([]any, 0, len(dagIns)*35)
	models := make([]*DagInstance, 0, len(dagIns))
	for _, data := range dagIns {
		sqlStr += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?), "
		t := ToDagInstanceModel(data, false)
		models = append(models, t)
		values = append(values,
			t.ID,
			t.CreatedAt,
			t.UpdatedAt,
			t.DagID,
			t.Trigger,
			t.Worker,
			t.Source,
			t.Vars,
			t.Keywords,
			t.EventPersistence,
			t.EventOssPath,
			t.ShareData,
			t.ShareDataExt,
			t.Status,
			t.Reason,
			t.Cmd,
			t.HasCmd,
			t.BatchRunID,
			t.UserID,
			t.EndedAt,
			t.DagType,
			t.PolicyType,
			t.AppInfo,
			t.Priority,
			t.Mode,
			t.Dump,
			t.DumpExt,
			t.SuccessCallback,
			t.ErrorCallback,
			t.CallChain,
			t.ResumeData,
			t.ResumeStatus,
			t.Version,
			t.VersionID,
			t.BizDomainID,
		)
	}

	sqlStr = sqlStr[:len(sqlStr)-1]

	msgStr, _ := jsoniter.MarshalToString(values)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGVAR_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, msgStr))

	err = d.db.Exec(sqlStr, values...).Error
	if err != nil {
		return nil, err
	}

	for i, data := range dagIns {
		if err = d.insertDagInstanceKeywords(newCtx, models[i].ID, data.Keywords); err != nil {
			return nil, err
		}
	}

	return dagIns, nil
}

// BatchCreateTaskIns implements [DagRepository].
func (d *dag) BatchCreateTaskIns(ctx context.Context, taskIns []*entity.TaskInstance) ([]*entity.TaskInstance, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if len(taskIns) == 0 {
		return taskIns, nil
	}

	batchSize := 1000
	for i := 0; i < len(taskIns); i += batchSize {
		end := i + batchSize
		if end > len(taskIns) {
			end = len(taskIns)
		}
		batch := taskIns[i:end]

		sqlStr := `INSERT INTO t_task_instance (
			f_id, f_created_at, f_updated_at, f_task_id, f_dag_ins_id, f_name, f_depend_on,
			f_action_name, f_timeout_secs, f_params, f_traces, f_status, f_reason, f_pre_checks,
			f_results, f_steps, f_last_modified_at, f_rendered_params, f_hash, f_settings, f_metadata
		) VALUES `

		values := make([]any, 0, len(batch)*21)
		for _, data := range batch {
			t := ToTaskInstanceModel(data, false)
			sqlStr += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"
			values = append(values,
				t.ID,
				t.CreatedAt,
				t.UpdatedAt,
				t.TaskID,
				t.DagInsID,
				t.Name,
				t.DependOn,
				t.ActionName,
				t.TimeoutSecs,
				t.Params,
				t.Traces,
				t.Status,
				t.Reason,
				t.PreChecks,
				t.Results,
				t.Steps,
				t.LastModifiedAt,
				t.RenderedParams,
				t.Hash,
				t.Settings,
				t.MetaData,
			)
		}
		sqlStr = strings.TrimSuffix(sqlStr, ",")

		msgStr, _ := jsoniter.MarshalToString(values)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, msgStr))

		if err = d.db.Exec(sqlStr, values...).Error; err != nil {
			return taskIns, err
		}
	}

	return taskIns, nil
}

// BatchDeleteDagIns implements [DagRepository].
func (d *dag) BatchDeleteDagIns(ctx context.Context, ids []string) error {
	return d.delete(ctx, ids, DAGINSTANCE_TABLENAME)
}

func (d *dag) delete(ctx context.Context, ids []string, tableName string) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sql := `DELETE FROM ` + tableName + ` WHERE f_id IN ?`
	msgStr, _ := jsoniter.MarshalToString(ids)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, msgStr))

	err = d.db.Exec(sql, ids).Error

	return err
}

// BatchDeleteDagWithTransaction implements [DagRepository].
func (d *dag) BatchDeleteDagWithTransaction(ctx context.Context, ids []string) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	fn := func() error {
		sql := `UPDATE t_dags SET f_removed = 1 WHERE f_id IN (?) AND f_type NOT IN (?)`
		msgStr, _ := jsoniter.MarshalToString(ids)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, msgStr), attribute.String(trace.DB_QUERY, common.DagTypeSecurityPolicy))

		err = d.db.Exec(sql, ids, common.DagTypeSecurityPolicy).Error
		if err != nil {
			return err
		}

		batchSize := 1000
		lastID := uint64(0) // 用于游标的最后一个ID

		for {
			// 获取要删除的 DAG 实例 ID（使用游标分页）
			var dagInsIDs []uint64
			var query string
			var queryArgs = []interface{}{ids}
			var cnt int64

			if lastID == 0 {
				// 第一次查询
				query = `SELECT f_id, f_dag_type
				FROM t_dag_instances
				WHERE f_dag_id IN ?
				ORDER BY f_id ASC
				LIMIT ?`
				queryArgs = append(queryArgs, batchSize)
			} else {
				// 后续查询，使用游标
				query = `SELECT f_id, f_dag_type
				FROM t_dag_instances
				WHERE f_dag_id IN ? AND f_id > ?
				ORDER BY f_id ASC
				LIMIT ?`
				queryArgs = append(queryArgs, lastID, batchSize)
			}

			rows, err := d.db.Raw(query, queryArgs).Rows()
			if err != nil {
				return fmt.Errorf("failed to query dag instances: %w", err)
			}

			for rows.Next() {
				var id uint64
				var dagType string
				if err := rows.Scan(&id, &dagType); err != nil {
					rows.Close()
					return err
				}
				cnt++
				if dagType == common.DagTypeSecurityPolicy {
					continue
				}
				dagInsIDs = append(dagInsIDs, id)
				lastID = id // 更新游标位置
			}
			rows.Close()

			if cnt == 0 {
				break // 没有更多记录可删除
			}

			// 删除 DAG 实例
			deleteDagInsQuery := `DELETE FROM t_dag_instances WHERE f_id IN ?`
			err = d.db.Exec(deleteDagInsQuery, dagInsIDs).Error
			if err != nil {
				return err
			}

			// 删除相关的任务实例
			deleteTaskInsQuery := `DELETE FROM t_task_instance WHERE f_dag_ins_id IN ?`
			err = d.db.Exec(deleteTaskInsQuery, dagInsIDs).Error
			if err != nil {
				return err
			}
		}

		return nil
	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(context.Context, mod.Store) error {
			return fn()
		})
	} else {
		err = fn()
	}

	return err
}

// BatchDeleteTaskIns implements [DagRepository].
func (d *dag) BatchDeleteTaskIns(ctx context.Context, ids []string) error {
	return d.delete(ctx, ids, TASKINSTANCE_TABLENAME)
}

// BatchUpdateDagIns implements [DagRepository].
func (d *dag) BatchUpdateDagIns(ctx context.Context, dagIns []*entity.DagInstance) error {
	var err error
	_, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	batchSize := 1000

	for i := 0; i < len(dagIns); i += batchSize {
		end := i + batchSize
		if end > len(dagIns) {
			end = len(dagIns)
		}

		batch := dagIns[i:end]

		for _, dagIns := range batch {
			dagIns.Update()

			// 方法1：使用UPDATE完全替换（假设文档必须存在）
			sql := `UPDATE t_dag_instance SET
				f_updated_at = ?, f_trigger = ?, f_worker = ?, f_source = ?,
				f_vars = ?, f_keywords = ?, f_event_persistence = ?, f_event_oss_path = ?, f_share_data = ?, f_share_data_ext = ?,
				f_status = ?, f_reason = ?, f_cmd = ?, f_has_cmd = ?, f_batch_run_id = ?, f_user_id = ?, f_ended_at = ?, f_dag_type = ?, f_policy_type = ?, f_appinfo = ?,
				f_priority = ?, f_mode = ?, f_dump = ?, f_dump_ext = ?, f_success_callback = ?, f_error_callback = ?, f_call_chain = ?,
				f_resume_data = ?, f_resume_status = ?, f_version = ?, f_version_id = ?, f_biz_domain_id = ?
				WHERE f_id = ?`

			// 执行 SQL 语句
			t := ToDagInstanceModel(dagIns, false)

			err = d.db.Exec(sql,
				t.UpdatedAt,
				t.Trigger,
				t.Worker,
				t.Source,
				t.Vars,
				t.Keywords,
				t.EventPersistence,
				t.EventOssPath,
				t.ShareData,
				t.ShareDataExt,
				t.Status,
				t.Reason,
				t.Cmd,
				t.HasCmd,
				t.BatchRunID,
				t.UserID,
				t.EndedAt,
				t.DagType,
				t.PolicyType,
				t.AppInfo,
				t.Priority,
				t.Mode,
				t.Dump,
				t.DumpExt,
				t.SuccessCallback,
				t.ErrorCallback,
				t.CallChain,
				t.ResumeData,
				t.ResumeStatus,
				t.Version,
				t.VersionID,
				t.BizDomainID,
				t.ID,
			).Error
			if err != nil {
				return err
			}

			if err = d.replaceDagInstanceKeywords(ctx, t.ID, dagIns.Keywords); err != nil {
				return err
			}
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// BatchUpdateTaskIns implements [DagRepository].
func (d *dag) BatchUpdateTaskIns(taskIns []*entity.TaskInstance) error {
	if len(taskIns) == 0 {
		return nil
	}
	for i := range taskIns {
		taskIns[i].Update()
		t := ToTaskInstanceModel(taskIns[i], true)

		sql := `UPDATE t_task_instance SET
			f_updated_at = ?, f_task_id = ?, f_dag_ins_id = ?, f_name = ?, f_depend_on = ?,
			f_action_name = ?, f_timeout_secs = ?, f_params = ?, f_traces = ?, f_status = ?,
			f_reason = ?, f_pre_checks = ?, f_results = ?, f_steps = ?, f_last_modified_at = ?,
			f_rendered_params = ?, f_hash = ?, f_settings = ?, f_metadata = ?
			WHERE f_id = ?`

		if err := d.db.Exec(sql,
			t.UpdatedAt,
			t.TaskID,
			t.DagInsID,
			t.Name,
			t.DependOn,
			t.ActionName,
			t.TimeoutSecs,
			t.Params,
			t.Traces,
			t.Status,
			t.Reason,
			t.PreChecks,
			t.Results,
			t.Steps,
			t.LastModifiedAt,
			t.RenderedParams,
			t.Hash,
			t.Settings,
			t.MetaData,
			t.ID,
		).Error; err != nil {
			return err
		}
	}

	return nil
}

// Close implements [DagRepository].
func (d *dag) Close() {
	if d == nil || d.db == nil {
		return
	}
	sqlDB, err := d.db.DB()
	if err != nil {
		return
	}
	_ = sqlDB.Close()
}

// CreatOutBoxMessage implements [DagRepository].
func (d *dag) CreatOutBoxMessage(ctx context.Context, outBox *entity.OutBox) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sql := `INSERT INTO t_outbox (f_id, f_topic, f_msg, f_created_at, f_updated_at) VALUES (?, ?, ?, ?, ?)`

	msgStr, _ := jsoniter.MarshalToString(outBox)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, OUTBOXMESSAGE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr), attribute.String(trace.DB_QUERY, msgStr))

	err = d.db.Exec(sql,
		outBox.ID,
		outBox.Topic,
		outBox.Msg,
		outBox.CreatedAt,
		outBox.UpdatedAt,
	).Error

	return err
}

// CreateClient implements [DagRepository].
func (d *dag) CreateClient(clientName string, clientID string, clientSecret string) error {
	id, err := utils.GetUniqueID()
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	sql := `INSERT INTO t_client (f_id, f_created_at, f_updated_at, f_client_name, f_client_id, f_client_secret) VALUES (?, ?, ?, ?, ?, ?)`
	return d.db.Exec(sql, id, now, now, clientName, clientID, clientSecret).Error
}

// CreateInbox implements [DagRepository].
func (d *dag) CreateInbox(ctx context.Context, msg *entity.InBox) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sql := `INSERT INTO t_inbox (f_id, f_msg, f_topic, f_docid, f_dag, f_created_at, f_updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`

	msgStr, _ := jsoniter.MarshalToString(msg)

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, INBOXMESSAGE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

	dagStr, _ := jsoniter.MarshalToString(msg.Dags)
	err = d.db.Exec(sql,
		msg.ID,
		msg.Msg,
		msg.Topic,
		msg.DocID,
		dagStr,
		msg.CreatedAt,
		msg.UpdatedAt,
	).Error

	return err
}

// CreateLogs implements [DagRepository].
func (d *dag) CreateLogs(ctx context.Context, ossLogs []*entity.Log) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if len(ossLogs) == 0 {
		return nil
	}

	batchSize := 1000
	for i := 0; i < len(ossLogs); i += batchSize {
		end := i + batchSize
		if end > len(ossLogs) {
			end = len(ossLogs)
		}
		batch := ossLogs[i:end]

		sqlStr := `INSERT INTO t_log (f_id, f_created_at, f_updated_at, f_ossid, f_key, f_filename) VALUES `
		values := make([]any, 0, len(batch)*6)
		for _, logItem := range batch {
			logItem.Initial()
			id, _ := strconv.ParseUint(logItem.ID, 10, 64)
			sqlStr += "(?, ?, ?, ?, ?, ?),"
			values = append(values, id, logItem.CreatedAt, logItem.UpdatedAt, logItem.OssID, logItem.Key, logItem.FileName)
		}
		sqlStr = strings.TrimSuffix(sqlStr, ",")

		msgStr, _ := jsoniter.MarshalToString(values)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, "t_log"), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, msgStr))

		if err = d.db.Exec(sqlStr, values...).Error; err != nil {
			return err
		}
	}
	return nil
}

// CreateTaskIns implements [DagRepository].
func (d *dag) CreateTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sql := `INSERT INTO t_task_instance (
		f_id, f_created_at, f_updated_at, f_task_id, f_dag_ins_id, f_name, f_depend_on,
		f_action_name, f_timeout_secs, f_params, f_traces, f_status, f_reason, f_pre_checks,
		f_results, f_steps, f_last_modified_at, f_rendered_params, f_hash, f_settings, f_metadata
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	t := ToTaskInstanceModel(taskIns, false)
	msgStr, _ := jsoniter.MarshalToString(t)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

	return d.db.Exec(sql,
		t.ID,
		t.CreatedAt,
		t.UpdatedAt,
		t.TaskID,
		t.DagInsID,
		t.Name,
		t.DependOn,
		t.ActionName,
		t.TimeoutSecs,
		t.Params,
		t.Traces,
		t.Status,
		t.Reason,
		t.PreChecks,
		t.Results,
		t.Steps,
		t.LastModifiedAt,
		t.RenderedParams,
		t.Hash,
		t.Settings,
		t.MetaData,
	).Error
}

// CreateToken implements [DagRepository].
func (d *dag) CreateToken(token *entity.Token) error {
	baseInfo := token.GetBaseInfo()
	baseInfo.Initial()
	id, _ := strconv.ParseUint(baseInfo.ID, 10, 64)

	sql := `INSERT INTO t_token (
		f_id, f_created_at, f_updated_at, f_user_id, f_user_name, f_refresh_token,
		f_token, f_expires_in, f_login_ip, f_is_app
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	if err := d.db.Exec(sql,
		id,
		baseInfo.CreatedAt,
		baseInfo.UpdatedAt,
		token.UserID,
		token.UserName,
		token.RefreshToken,
		token.Token,
		token.ExpiresIn,
		token.LoginIP,
		token.IsApp,
	).Error; err != nil {
		return err
	}
	return nil
}

// DeleteDagInsByID implements [DagRepository].
func (d *dag) DeleteDagInsByID(ctx context.Context, params map[string]interface{}) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	result, err := NewConverter(DAGINSTANCE_TABLENAME, WithAutoConvert(true)).Convert(params)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf("DELETE FROM t_dag_instance WHERE %s", result.Conds)
	msyBytes, _ := jsoniter.MarshalToString(result.Params)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, msyBytes))

	err = d.db.Exec(sql, result.Params...).Error

	return err
}

// DeleteInbox implements [DagRepository].
func (d *dag) DeleteInbox(ctx context.Context, ids []string) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sql := `DELETE FROM t_inbox WHERE f_id IN ?`

	msgStr, _ := jsoniter.MarshalToString(ids)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, INBOXMESSAGE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, msgStr))

	err = d.db.Exec(sql, ids).Error

	return err
}

// DeleteOutBoxMessage implements [DagRepository].
func (d *dag) DeleteOutBoxMessage(ctx context.Context, ids []string) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sql := `DELETE FROM t_outbox WHERE f_id IN ?`

	msgStr, _ := jsoniter.MarshalToString(ids)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, OUTBOXMESSAGE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, msgStr))

	err = d.db.Exec(sql, ids).Error

	return err
}

// DeleteTaskInsByDagInsID implements [DagRepository].
func (d *dag) DeleteTaskInsByDagInsID(ctx context.Context, dagInsID string) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	id, _ := strconv.ParseUint(dagInsID, 10, 64)
	sql := `DELETE FROM t_task_instance WHERE f_dag_ins_id = ?`
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, dagInsID))

	return d.db.Exec(sql, id).Error
}

// DeleteTaskInsByID implements [DagRepository].
func (d *dag) DeleteTaskInsByID(ctx context.Context, params map[string]interface{}) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	status, _ := params["status"].([]entity.TaskInstanceStatus)
	var statusVals []string
	for _, s := range status {
		statusVals = append(statusVals, string(s))
	}

	dagInsIDs, _ := params["dagInsIDs"].([]string)
	var dagInsU64 []uint64
	for _, id := range dagInsIDs {
		v, _ := strconv.ParseUint(id, 10, 64)
		dagInsU64 = append(dagInsU64, v)
	}
	if len(dagInsU64) == 0 || len(statusVals) == 0 {
		return nil
	}

	updatedAt, _ := params["updatedAt"].(int64)
	maxIDStr, _ := params["_id"].(string)
	maxID, _ := strconv.ParseUint(maxIDStr, 10, 64)

	sql := `DELETE FROM t_task_instance WHERE f_id <= ? AND f_dag_ins_id IN ? AND f_status IN ? AND f_updated_at <= ?`
	msgStr, _ := jsoniter.MarshalToString(params)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, msgStr))

	return d.db.Exec(sql, maxID, dagInsU64, statusVals, updatedAt).Error
}

// DeleteToken implements [DagRepository].
func (d *dag) DeleteToken(id string) error {
	sql := `DELETE FROM t_token WHERE f_id = ?`
	uid, _ := strconv.ParseUint(id, 10, 64)
	return d.db.Exec(sql, uid).Error
}

// DisdinctDagInstance implements [DagRepository].
func (d *dag) DisdinctDagInstance(input *mod.ListDagInstanceInput) ([]interface{}, error) {
	var conds []string
	var args []interface{}

	if len(input.Status) > 0 {
		var status []string
		for _, v := range input.Status {
			status = append(status, string(v))
		}
		conds = append(conds, "f_status IN ?")
		args = append(args, status)
	}
	if len(input.DagIDs) > 0 {
		var ids []uint64
		for _, v := range input.DagIDs {
			id, _ := strconv.ParseUint(v, 10, 64)
			ids = append(ids, id)
		}
		conds = append(conds, "f_dag_id IN ?")
		args = append(args, ids)
	}
	if input.Worker != "" {
		conds = append(conds, "f_worker = ?")
		args = append(args, input.Worker)
	}
	if input.UpdatedEnd > 0 {
		conds = append(conds, "f_updated_at <= ?")
		args = append(args, input.UpdatedEnd)
	}
	if input.ExcludeModeVM {
		conds = append(conds, "f_mode <> ?")
		args = append(args, entity.DagInstanceModeVM)
	}

	field := camelToFSnake(input.DistinctField)
	sql := fmt.Sprintf("SELECT DISTINCT %s FROM t_dag_instance", field)
	if len(conds) > 0 {
		sql += " WHERE " + strings.Join(conds, " AND ")
	}
	if input.SortBy != "" {
		dir := utils.IfNot(input.Order == 0, "DESC", "ASC")
		sql += fmt.Sprintf(" ORDER BY %s %s", camelToFSnake(input.SortBy), dir)
	}
	if input.Limit > 0 {
		sql += " LIMIT ? OFFSET ?"
		args = append(args, input.Limit, input.Limit*input.Offset)
	}

	var res []interface{}
	if err := d.db.Raw(sql, args...).Scan(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

// GetClient implements [DagRepository].
func (d *dag) GetClient(clientName string) (client *entity.Client, err error) {
	sql := `SELECT f_id, f_created_at, f_updated_at, f_client_name, f_client_id, f_client_secret FROM t_client WHERE f_client_name = ?`
	model := &ClientModel{}
	if err = d.db.Raw(sql, clientName).Scan(model).Error; err != nil {
		return nil, err
	}
	if model.ID == 0 {
		return &entity.Client{}, nil
	}
	dest := &entity.Client{}
	if err = ToEntity(model, dest); err != nil {
		return nil, err
	}
	return dest, nil
}

// GetDagCount implements [DagRepository].
func (d *dag) GetDagCount(ctx context.Context, params map[string]interface{}) (int64, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if params == nil {
		params = map[string]interface{}{}
	}
	params["type"] = bson.M{"$ne": common.DagTypeSecurityPolicy}

	result, err := NewConverter(DAG_TABLENAME, WithAutoConvert(true)).ConvertConds(params)
	if err != nil {
		return 0, err
	}

	conds := result.Conds
	if conds == "" {
		conds = "1=1"
	}
	conds += " AND f_removed <> 1 AND f_is_debug <> 1"

	sql := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", DAG_TABLENAME, conds)
	query, _ := jsoniter.MarshalToString(result.Params)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, query))

	var count int64
	err = d.db.Raw(sql, result.Params...).Scan(&count).Error
	return count, err
}

// GetDagInstance implements [DagRepository].
func (d *dag) GetDagInstance(ctx context.Context, dagInsId string) (*entity.DagInstance, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	return d.GetDagInstanceByFields(newCtx, map[string]interface{}{"f_id": dagInsId})
}

// GetDagInstanceByFields implements [DagRepository].
func (d *dag) GetDagInstanceByFields(ctx context.Context, params map[string]interface{}) (*entity.DagInstance, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	result, err := NewConverter(DAGINSTANCE_TABLENAME, WithAutoConvert(true)).Convert(params)
	if err != nil {
		return nil, err
	}

	query, _ := jsoniter.MarshalToString(result.Params)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, result.SQL), attribute.String(trace.DB_QUERY, query))

	dagIns := &DagInstance{}
	err = d.db.Raw(result.SQL, result.Params...).Scan(dagIns).Error
	if err != nil {
		return nil, err
	}
	if dagIns.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	dest := &entity.DagInstance{}
	err = ToEntity(dagIns, dest)
	return dest, err
}

// GetDagInstanceCount implements [DagRepository].
func (d *dag) GetDagInstanceCount(ctx context.Context, params map[string]interface{}) (int64, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sql, args, err := BuildDagInstanceCountQueryFromParams(params)
	if err != nil {
		return 0, err
	}
	query, _ := jsoniter.MarshalToString(args)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, query))

	var count int64
	err = d.db.Raw(sql, args...).Scan(&count).Error
	return count, err
}

// GetInbox implements [DagRepository].
func (d *dag) GetInbox(ctx context.Context, id string) (*entity.InBox, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sql := `SELECT f_id, f_msg, f_topic, f_docid, f_dag, f_created_at, f_updated_at FROM t_inbox WHERE f_id = ?`

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, INBOXMESSAGE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, id))

	inbox := &InBox{}
	err = d.db.Raw(sql, id).Scan(inbox).Error
	if err != nil {
		return nil, err
	}

	var docMsg common.DocMsg
	_ = jsoniter.UnmarshalFromString(inbox.Msg, &docMsg)
	var dags []string
	_ = jsoniter.UnmarshalFromString(inbox.Dags, &dags)

	return &entity.InBox{
		BaseInfo: entity.BaseInfo{
			ID:        fmt.Sprintf("%v", inbox.ID),
			CreatedAt: inbox.CreatedAt,
			UpdatedAt: inbox.UpdatedAt,
		},
		Msg:   docMsg,
		Topic: inbox.Topic,
		DocID: inbox.DocID,
		Dags:  dags,
	}, nil
}

// GetSwitchStatus implements [DagRepository].
func (d *dag) GetSwitchStatus() (bool, error) {
	sql := `SELECT f_id, f_created_at, f_updated_at, f_name, f_status FROM t_switch WHERE f_name = ?`
	sw := &SwitchModel{}
	if err := d.db.Raw(sql, entity.SwitchName).Scan(sw).Error; err != nil {
		return false, fmt.Errorf("get switch status failed: %w", err)
	}
	if sw.ID == 0 {
		return true, nil
	}
	return sw.Status, nil
}

// GetTaskIns implements [DagRepository].
func (d *dag) GetTaskIns(ctx context.Context, taskIns string) (*entity.TaskInstance, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sql := `SELECT * FROM t_task_instance WHERE f_id = ?`
	id, _ := strconv.ParseUint(taskIns, 10, 64)
	model := &TaskInstanceModel{}
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, taskIns))

	if err = d.db.Raw(sql, id).Scan(model).Error; err != nil {
		return nil, err
	}
	if model.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	dest := &entity.TaskInstance{}
	if err = ToEntity(model, dest); err != nil {
		return nil, err
	}
	return dest, nil
}

// GetTaskInstanceCount implements [DagRepository].
func (d *dag) GetTaskInstanceCount(ctx context.Context, params map[string]interface{}) (int64, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	result, err := NewConverter(TASKINSTANCE_TABLENAME, WithAutoConvert(true)).ConvertConds(params)
	if err != nil {
		return 0, err
	}

	conds := result.Conds
	if conds == "" {
		conds = "1=1"
	}
	sql := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", TASKINSTANCE_TABLENAME, conds)
	query, _ := jsoniter.MarshalToString(result.Params)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, query))

	var count int64
	err = d.db.Raw(sql, result.Params...).Scan(&count).Error
	return count, err
}

// GetTokenByUserID implements [DagRepository].
func (d *dag) GetTokenByUserID(userID string) (*entity.Token, error) {
	sql := `SELECT f_id, f_created_at, f_updated_at, f_user_id, f_user_name, f_refresh_token, f_token, f_expires_in, f_login_ip, f_is_app FROM t_token WHERE f_user_id = ?`
	model := &TokenModel{}
	if err := d.db.Raw(sql, userID).Scan(model).Error; err != nil {
		return &entity.Token{}, fmt.Errorf("get token failed: %w", err)
	}
	if model.ID == 0 {
		return &entity.Token{}, nil
	}
	dest := &entity.Token{}
	if err := ToEntity(model, dest); err != nil {
		return nil, err
	}
	return dest, nil
}

// GroupDagInstance implements [DagRepository].
func (d *dag) GroupDagInstance(ctx context.Context, input *mod.GroupInput) ([]*entity.DagInstanceGroup, error) {
	var err error
	_, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	conds := make([]string, 0)
	args := make([]interface{}, 0)
	conv := NewConverter(DAGINSTANCE_TABLENAME, WithAutoConvert(true))

	for _, opt := range input.SearchOptions {
		m := map[string]interface{}{
			opt.Field: map[string]interface{}{
				opt.Condition: opt.Value,
			},
		}
		res, cerr := conv.ConvertConds(m)
		if cerr != nil {
			return nil, cerr
		}
		if res.Conds != "" {
			conds = append(conds, res.Conds)
			args = append(args, res.Params...)
		}
	}
	if input.TimeRange != nil {
		m := map[string]interface{}{
			input.TimeRange.Field: map[string]interface{}{
				"$gte": input.TimeRange.Begin,
				"$lte": input.TimeRange.End,
			},
		}
		res, cerr := conv.ConvertConds(m)
		if cerr != nil {
			return nil, cerr
		}
		if res.Conds != "" {
			conds = append(conds, res.Conds)
			args = append(args, res.Params...)
		}
	}

	groupCols := []string{}
	if input.GroupBy != "" {
		groupCols = append(groupCols, camelToFSnake(input.GroupBy))
	}
	for _, g := range input.GroupBys {
		groupCols = append(groupCols, camelToFSnake(g))
	}
	if len(groupCols) == 0 {
		return nil, nil
	}

	sortCol := "f_updated_at"
	if input.SortBy != "" {
		sortCol = camelToFSnake(input.SortBy)
	}
	order := "DESC"
	if input.Order > 0 {
		order = "ASC"
	}

	baseSQL := "FROM t_dag_instance"
	if len(conds) > 0 {
		baseSQL += " WHERE " + strings.Join(conds, " AND ")
	}

	groupSQL := fmt.Sprintf("SELECT %s, COUNT(*) AS total, MAX(%s) AS max_sort %s GROUP BY %s ORDER BY max_sort %s",
		strings.Join(groupCols, ", "), sortCol, baseSQL, strings.Join(groupCols, ", "), order,
	)
	if input.Limit > 0 {
		groupSQL += " LIMIT ?"
		args = append(args, input.Limit)
	}

	type groupRow struct {
		Total   int64
		MaxSort int64
	}
	rows := make([]map[string]interface{}, 0)
	if err = d.db.Raw(groupSQL, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	var result []*entity.DagInstanceGroup
	for _, row := range rows {
		totalVal := toInt64(row["total"])

		conds2 := make([]string, 0, len(groupCols)+1)
		args2 := make([]interface{}, 0, len(groupCols)+1)
		for _, col := range groupCols {
			conds2 = append(conds2, fmt.Sprintf("%s = ?", col))
			args2 = append(args2, row[col])
		}
		conds2 = append(conds2, fmt.Sprintf("%s = ?", sortCol))
		args2 = append(args2, row["max_sort"])

		sql := "SELECT * FROM t_dag_instance WHERE " + strings.Join(conds2, " AND ") + " ORDER BY f_id DESC LIMIT 1"
		model := &DagInstance{}
		if err = d.db.Raw(sql, args2...).Scan(model).Error; err != nil {
			return nil, err
		}
		if model.ID == 0 {
			continue
		}
		dagIns := &entity.DagInstance{}
		if err = ToEntity(model, dagIns); err != nil {
			return nil, err
		}
		result = append(result, &entity.DagInstanceGroup{
			Total:  totalVal,
			DagIns: dagIns,
		})
	}

	return result, nil
}

// ListDag implements [DagRepository].
func (d *dag) ListDag(ctx context.Context, input *mod.ListDagInput) ([]*entity.Dag, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if input == nil {
		return nil, nil
	}

	var conds []string
	var args []interface{}

	// default type exclude
	if input.Type != "" {
		if input.Type != "all" {
			conds = append(conds, "f_type = ?")
			args = append(args, input.Type)
		}
	} else {
		conds = append(conds, "f_type NOT IN ?")
		args = append(args, []string{
			common.DagTypeSecurityPolicy,
			common.DagTypeDataFlow,
			common.DagTypeDataFlowForBot,
			common.DagTypeComboOperator,
		})
	}

	if input.TriggerType != "" {
		conds = append(conds, "f_trigger = ?")
		args = append(args, input.TriggerType)
	}
	if len(input.TriggerTypes) > 0 {
		conds = append(conds, "f_trigger IN ?")
		args = append(args, input.TriggerTypes)
	}
	if input.UserID != "" && input.Scope != "all" {
		conds = append(conds, "f_user_id = ?")
		args = append(args, input.UserID)
	}
	if input.KeyWord != "" {
		conds = append(conds, "f_name LIKE ?")
		args = append(args, "%"+input.KeyWord+"%")
	}
	if len(input.DagIDs) > 0 {
		conds = append(conds, "f_id IN ?")
		args = append(args, parseUint64Slice(input.DagIDs))
	}
	if len(input.Status) > 0 {
		var status []string
		for _, s := range input.Status {
			status = append(status, string(s))
		}
		conds = append(conds, "f_status IN ?")
		args = append(args, status)
	}

	conds = append(conds, "f_removed <> 1", "f_is_debug <> 1")

	indexCond, indexArgs := BuildDagIndexSubquery(input)
	if indexCond != "" {
		conds = append(conds, indexCond)
		args = append(args, indexArgs...)
	}

	sql := "SELECT * FROM t_dag"
	if len(conds) > 0 {
		sql += " WHERE " + strings.Join(conds, " AND ")
	}
	if input.SortBy != "" {
		sql += fmt.Sprintf(" ORDER BY %s %s", camelToFSnake(input.SortBy), utils.IfNot(input.Order == 0, "DESC", "ASC"))
	}
	if input.Limit > 0 {
		sql += " LIMIT ? OFFSET ?"
		args = append(args, input.Limit, input.Limit*input.Offset)
	}

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql))

	models := make([]*Dag, 0)
	if err = d.db.Raw(sql, args...).Scan(&models).Error; err != nil {
		return nil, err
	}

	var res []*entity.Dag
	for _, model := range models {
		if model.ID == 0 {
			continue
		}
		dag := &entity.Dag{}
		if err = ToEntity(model, dag); err != nil {
			return nil, err
		}
		res = append(res, dag)
	}
	return res, nil
}

// ListDagByFields implements [DagRepository].
func (d *dag) ListDagByFields(ctx context.Context, filter bson.M, opt options.FindOptions) ([]*entity.Dag, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if filter == nil {
		filter = bson.M{}
	}
	if _, ok := filter["removed"]; !ok {
		filter["removed"] = bson.M{"$ne": true}
	}
	if _, ok := filter["is_debug"]; !ok {
		filter["is_debug"] = bson.M{"$ne": true}
	}

	result, err := NewConverter(DAG_TABLENAME, WithAutoConvert(true)).Convert(filter)
	if err != nil {
		return nil, err
	}

	sql := result.SQL
	args := result.Params

	if opt.Sort != nil {
		switch v := opt.Sort.(type) {
		case map[string]interface{}:
			for k, order := range v {
				dir := "ASC"
				if ord, ok := order.(int); ok && ord < 0 {
					dir = "DESC"
				}
				sql += fmt.Sprintf(" ORDER BY %s %s", camelToFSnake(k), dir)
				break
			}
		case bson.D:
			if len(v) > 0 {
				dir := "ASC"
				if v[0].Value.(int32) < 0 {
					dir = "DESC"
				}
				sql += fmt.Sprintf(" ORDER BY %s %s", camelToFSnake(v[0].Key), dir)
			}
		}
	}
	if opt.Limit != nil && *opt.Limit > 0 {
		sql += " LIMIT ?"
		args = append(args, *opt.Limit)
		if opt.Skip != nil && *opt.Skip > 0 {
			sql += " OFFSET ?"
			args = append(args, *opt.Skip)
		}
	}

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql))

	models := make([]*Dag, 0)
	if err = d.db.Raw(sql, args...).Scan(&models).Error; err != nil {
		return nil, err
	}
	var res []*entity.Dag
	for _, model := range models {
		if model.ID == 0 {
			continue
		}
		dag := &entity.Dag{}
		if err = ToEntity(model, dag); err != nil {
			return nil, err
		}
		res = append(res, dag)
	}
	return res, nil
}

// ListDagCount implements [DagRepository].
func (d *dag) ListDagCount(ctx context.Context, input *mod.ListDagInput) (int64, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if input == nil {
		return 0, nil
	}

	var conds []string
	var args []interface{}

	conds = append(conds, "f_type NOT IN ?")
	args = append(args, []string{
		common.DagTypeSecurityPolicy,
		common.DagTypeDataFlow,
		common.DagTypeDataFlowForBot,
		common.DagTypeComboOperator,
	})

	if input.Type != "" {
		if input.Type == "all" {
			conds = conds[:0]
			args = args[:0]
		} else {
			conds = []string{"f_type = ?"}
			args = []interface{}{input.Type}
		}
	}

	if input.TriggerType != "" {
		conds = append(conds, "f_trigger = ?")
		args = append(args, input.TriggerType)
	}
	if input.UserID != "" {
		conds = append(conds, "f_user_id = ?")
		args = append(args, input.UserID)
	}
	if input.KeyWord != "" {
		conds = append(conds, "f_name LIKE ?")
		args = append(args, "%"+input.KeyWord+"%")
	}
	if len(input.DagIDs) > 0 {
		conds = append(conds, "f_id IN ?")
		args = append(args, parseUint64Slice(input.DagIDs))
	}
	if len(input.Status) > 0 {
		var status []string
		for _, s := range input.Status {
			status = append(status, string(s))
		}
		conds = append(conds, "f_status IN ?")
		args = append(args, status)
	}
	if input.BizDomainID != "" {
		if input.BizDomainID == common.BizDomainDefaultID {
			conds = append(conds, "(f_biz_domain_id = '' OR f_biz_domain_id = ? OR f_biz_domain_id IS NULL)")
			args = append(args, common.BizDomainDefaultID)
		} else {
			conds = append(conds, "f_biz_domain_id = ?")
			args = append(args, input.BizDomainID)
		}
	}

	conds = append(conds, "f_removed <> 1", "f_is_debug <> 1")

	if len(input.Sources) != 0 && len(input.Trigger) > 0 {
		conds = append(conds, "f_id IN (SELECT f_dag_id FROM t_dag_step_index WHERE f_operator IN ? AND f_source_id IN ?)")
		args = append(args, input.Trigger, input.Sources)
	}

	if input.Accessors != nil && input.UserID == "" {
		conds = append(conds, "f_id IN (SELECT f_dag_id FROM t_dag_accessor_index WHERE f_accessor_id IN ?)")
		args = append(args, input.Accessors)
	}

	sql := "SELECT COUNT(*) FROM t_dag"
	if len(conds) > 0 {
		sql += " WHERE " + strings.Join(conds, " AND ")
	}

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql))

	var count int64
	err = d.db.Raw(sql, args...).Scan(&count).Error
	return count, err
}

// ListDagCountByFields implements [DagRepository].
func (d *dag) ListDagCountByFields(ctx context.Context, filter bson.M) (int64, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if filter == nil {
		filter = bson.M{}
	}
	if _, ok := filter["removed"]; !ok {
		filter["removed"] = bson.M{"$ne": true}
	}
	if _, ok := filter["is_debug"]; !ok {
		filter["is_debug"] = bson.M{"$ne": true}
	}

	result, err := NewConverter(DAG_TABLENAME, WithAutoConvert(true)).ConvertConds(filter)
	if err != nil {
		return 0, err
	}
	conds := result.Conds
	if conds == "" {
		conds = "1=1"
	}
	sql := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", DAG_TABLENAME, conds)
	query, _ := jsoniter.MarshalToString(result.Params)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, query))

	var count int64
	err = d.db.Raw(sql, result.Params...).Scan(&count).Error
	return count, err
}

// ListDagInstanceInRangeTime implements [DagRepository].
func (d *dag) ListDagInstanceInRangeTime(ctx context.Context, status string, begin int64, end int64) ([]*entity.DagInstance, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sql := `SELECT * FROM t_dag_instance WHERE f_status = ? AND f_updated_at >= ? AND f_updated_at <= ?`
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql))

	models := make([]*DagInstance, 0)
	if err = d.db.Raw(sql, status, begin, end).Scan(&models).Error; err != nil {
		return nil, err
	}
	var res []*entity.DagInstance
	for _, model := range models {
		if model.ID == 0 {
			continue
		}
		dagIns := &entity.DagInstance{}
		if err = ToEntity(model, dagIns); err != nil {
			return nil, err
		}
		res = append(res, dagIns)
	}
	return res, nil
}

// ListDagVersions implements [DagRepository].
func (d *dag) ListDagVersions(ctx context.Context, input *mod.ListDagVersionInput) ([]entity.DagVersion, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sql := "SELECT * FROM t_dag_versions WHERE 1=1"
	args := make([]interface{}, 0)
	if input.DagID != "" {
		sql += " AND f_dag_id = ?"
		args = append(args, input.DagID)
	}
	if input.SortBy != "" {
		sql += fmt.Sprintf(" ORDER BY %s %s", camelToFSnake(input.SortBy), utils.IfNot(input.Order == 0, "DESC", "ASC"))
	}
	if input.Limit > 0 {
		sql += " LIMIT ? OFFSET ?"
		args = append(args, input.Limit, input.Limit*input.Offset)
	}

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGVERSIONS_TABLENAME), attribute.String(trace.DB_SQL, sql))

	var models []DagVersion
	if err = d.db.Raw(sql, args...).Scan(&models).Error; err != nil {
		return nil, err
	}
	var res []entity.DagVersion
	for _, m := range models {
		item := entity.DagVersion{}
		if err = ToEntity(&m, &item); err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

// ListExistDagID implements [DagRepository].
func (d *dag) ListExistDagID(ctx context.Context, dagIDs []string) ([]string, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if len(dagIDs) == 0 {
		return nil, nil
	}

	sql := `SELECT f_id FROM t_dag WHERE f_id IN ?`
	ids := parseUint64Slice(dagIDs)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql))

	var res []uint64
	if err = d.db.Raw(sql, ids).Scan(&res).Error; err != nil {
		return nil, err
	}
	var out []string
	for _, id := range res {
		out = append(out, strconv.FormatUint(id, 10))
	}
	return out, nil
}

// ListExistDagInsID implements [DagRepository].
func (d *dag) ListExistDagInsID(ctx context.Context, dagInsIDs []string) ([]string, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if len(dagInsIDs) == 0 {
		return nil, nil
	}

	sql := `SELECT f_id FROM t_dag_instance WHERE f_id IN ?`
	ids := parseUint64Slice(dagInsIDs)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql))

	var res []uint64
	if err = d.db.Raw(sql, ids).Scan(&res).Error; err != nil {
		return nil, err
	}
	var out []string
	for _, id := range res {
		out = append(out, strconv.FormatUint(id, 10))
	}
	return out, nil
}

// ListHistoryDagIns implements [DagRepository].
func (d *dag) ListHistoryDagIns(ctx context.Context, params map[string]interface{}, dataChannel chan []bson.M) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	status, _ := params["status"].([]entity.DagInstanceStatus)
	var statusVals []string
	for _, s := range status {
		statusVals = append(statusVals, string(s))
	}
	updatedAt, _ := params["updatedAt"].(int64)

	lastID := uint64(0)
	batchSize := common.DefaultQuerySize

	for {
		select {
		case <-newCtx.Done():
			close(dataChannel)
			return nil
		default:
		}

		sql := `SELECT * FROM t_dag_instance WHERE f_status IN ? AND f_updated_at <= ? AND f_id > ? ORDER BY f_id ASC LIMIT ?`
		var models []DagInstance
		if err = d.db.Raw(sql, statusVals, updatedAt, lastID, batchSize).Scan(&models).Error; err != nil {
			return err
		}
		if len(models) == 0 {
			close(dataChannel)
			return nil
		}

		var batch []bson.M
		for _, m := range models {
			lastID = m.ID
			e := &entity.DagInstance{}
			if err = ToEntity(&m, e); err != nil {
				return err
			}
			b, _ := bson.Marshal(e)
			var doc bson.M
			_ = bson.Unmarshal(b, &doc)
			batch = append(batch, doc)
		}
		if len(batch) > 0 {
			dataChannel <- batch
		}
		if len(models) < batchSize {
			close(dataChannel)
			return nil
		}
	}
}

// ListHistoryTaskIns implements [DagRepository].
func (d *dag) ListHistoryTaskIns(ctx context.Context, params map[string]interface{}, dataChannel chan []bson.M) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	status, _ := params["status"].([]entity.TaskInstanceStatus)
	var statusVals []string
	for _, s := range status {
		statusVals = append(statusVals, string(s))
	}
	updatedAt, _ := params["updatedAt"].(int64)

	lastID := uint64(0)
	batchSize := common.DefaultQuerySize

	for {
		select {
		case <-newCtx.Done():
			close(dataChannel)
			return nil
		default:
		}

		sql := `SELECT * FROM t_task_instance WHERE f_status IN ? AND f_updated_at <= ? AND f_id > ? ORDER BY f_id ASC LIMIT ?`
		var models []TaskInstanceModel
		if err = d.db.Raw(sql, statusVals, updatedAt, lastID, batchSize).Scan(&models).Error; err != nil {
			return err
		}
		if len(models) == 0 {
			close(dataChannel)
			return nil
		}

		var batch []bson.M
		for _, m := range models {
			lastID = m.ID
			e := &entity.TaskInstance{}
			if err = ToEntity(&m, e); err != nil {
				return err
			}
			b, _ := bson.Marshal(e)
			var doc bson.M
			_ = bson.Unmarshal(b, &doc)
			batch = append(batch, doc)
		}
		if len(batch) > 0 {
			dataChannel <- batch
		}
		if len(models) < batchSize {
			close(dataChannel)
			return nil
		}
	}
}

// ListInbox implements [DagRepository].
func (d *dag) ListInbox(ctx context.Context, input *mod.ListInboxInput) ([]*entity.InBox, error) {
	// 构建 SQL 查询
	sqlQuery := "SELECT * FROM in_box WHERE 1=1"
	var args []interface{}

	// 应用筛选条件
	if input != nil {
		// 根据用户ID筛选
		if input.DocID != "" {
			sqlQuery += " AND f_docid = ?"
			args = append(args, input.DocID)
		}

		// 根据消息类型筛选
		if len(input.Topics) > 0 {
			sqlQuery += " AND f_topic IN ?"
			args = append(args, input.Topics)
		}

		// 根据状态筛选
		if input.Now > 0 {
			sqlQuery += " AND f_created_at <= ?"
			// input.Now - 2*60
			args = append(args, input.Now-2*60)
		}

		// 排序
		orderBy := "f_created_at"
		order := "DESC"
		if input.SortBy != "" {
			orderBy = input.SortBy
		}
		if input.Order > 0 {
			order = "ASC"
		}
		sqlQuery += fmt.Sprintf(" ORDER BY %s %s", orderBy, order)

		// 分页处理（放在最后）
		if input.Limit >= 0 && input.Offset >= 0 {
			offset := input.Offset * input.Limit
			sqlQuery += " LIMIT ? OFFSET ?"
			args = append(args, input.Limit, offset)
		}
	}

	// 执行原生 SQL 查询
	var inboxes []*InBox
	if err := d.db.Raw(sqlQuery, args...).Scan(&inboxes).Error; err != nil {
		return nil, err
	}

	var res []*entity.InBox
	for _, inbox := range inboxes {
		var docMsg common.DocMsg
		_ = jsoniter.UnmarshalFromString(inbox.Msg, &docMsg)
		var dags []string
		_ = jsoniter.UnmarshalFromString(inbox.Dags, &dags)

		res = append(res, &entity.InBox{
			BaseInfo: entity.BaseInfo{
				ID:        fmt.Sprintf("%v", inbox.ID),
				CreatedAt: inbox.CreatedAt,
				UpdatedAt: inbox.UpdatedAt,
			},
			Msg:   docMsg,
			Topic: inbox.Topic,
			DocID: inbox.DocID,
			Dags:  dags,
		})
	}

	return res, nil
}

// ListOutBoxMessage implements [DagRepository].
// TODO : 数据结构标签不一致需要转换
func (d *dag) ListOutBoxMessage(ctx context.Context, input *entity.OutBoxInput) ([]*entity.OutBox, error) {
	var err error
	var msgs []*entity.OutBox

	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx = newCtx

	sqlStr := `SELECT f_id, f_topic, f_msg, f_created_at, f_updated_at FROM t_outbox WHERE 1 == 1 `

	values := make([]interface{}, 0)
	if input.CreateTime > 0 {
		sqlStr += " AND f_created_at <= ?"
		values = append(values, input.CreateTime)
	}

	if input.Limit > 0 {
		sqlStr += " LIMIT ?"
		values = append(values, input.CreateTime)
	}

	err = d.db.Raw(sqlStr, values...).Scan(&msgs).Error
	if err != nil {
		return nil, err
	}

	return msgs, nil
}

// ListTaskInstance implements [DagRepository].
func (d *dag) ListTaskInstance(ctx context.Context, input *mod.ListTaskInstanceInput) ([]*entity.TaskInstance, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	conds := make([]string, 0)
	args := make([]interface{}, 0)

	if len(input.IDs) > 0 {
		var ids []uint64
		for _, id := range input.IDs {
			v, _ := strconv.ParseUint(id, 10, 64)
			ids = append(ids, v)
		}
		conds = append(conds, "f_id IN ?")
		args = append(args, ids)
	}

	var actionConds []string
	var actionArgs []interface{}
	if len(input.ActionName) > 0 {
		actionConds = append(actionConds, "f_action_name IN ?")
		actionArgs = append(actionArgs, input.ActionName)
	}
	if input.ActionNameRegex != "" {
		actionConds = append(actionConds, "f_action_name LIKE ?")
		actionArgs = append(actionArgs, regexToLike(input.ActionNameRegex))
	}
	if len(actionConds) > 0 {
		conds = append(conds, "("+strings.Join(actionConds, " OR ")+")")
		args = append(args, actionArgs...)
	}

	if len(input.Status) > 0 {
		var status []string
		for _, s := range input.Status {
			status = append(status, string(s))
		}
		conds = append(conds, "f_status IN ?")
		args = append(args, status)
	}

	if input.Expired {
		conds = append(conds, "f_updated_at <= ? - f_timeout_secs")
		args = append(args, time.Now().Unix()-5)
	}

	if input.DagInsID != "" {
		id, _ := strconv.ParseUint(input.DagInsID, 10, 64)
		conds = append(conds, "f_dag_ins_id = ?")
		args = append(args, id)
	} else if len(input.DagInsIDs) > 0 {
		var ids []uint64
		for _, id := range input.DagInsIDs {
			v, _ := strconv.ParseUint(id, 10, 64)
			ids = append(ids, v)
		}
		conds = append(conds, "f_dag_ins_id IN ?")
		args = append(args, ids)
	}

	if input.Hash != "" {
		conds = append(conds, "f_hash = ?")
		args = append(args, input.Hash)
	}

	sql := "SELECT * FROM t_task_instance"
	if len(conds) > 0 {
		sql += " WHERE " + strings.Join(conds, " AND ")
	}
	if input.SortBy != "" {
		sql += fmt.Sprintf(" ORDER BY %s %s", camelToFSnake(input.SortBy), utils.IfNot(input.Order == 0, "DESC", "ASC"))
	}
	if input.Limit > 0 {
		sql += " LIMIT ? OFFSET ?"
		args = append(args, input.Limit, input.Limit*input.Offset)
	}

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql))

	models := make([]*TaskInstanceModel, 0)
	if err = d.db.Raw(sql, args...).Scan(&models).Error; err != nil {
		return nil, err
	}

	var res []*entity.TaskInstance
	for _, model := range models {
		if model.ID == 0 {
			continue
		}
		task := &entity.TaskInstance{}
		if err = ToEntity(model, task); err != nil {
			return nil, err
		}
		res = append(res, task)
	}

	return res, nil
}

// Marshal implements [DagRepository].
func (d *dag) Marshal(obj interface{}) ([]byte, error) {
	return jsoniter.Marshal(obj)
}

// PatchDagIns implements [DagRepository].
func (d *dag) PatchDagIns(ctx context.Context, dagIns *entity.DagInstance, mustsPatchFields ...string) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if dagIns.ID == "" {
		return fmt.Errorf("id cannot be empty")
	}

	setClauses := []string{"f_updated_at = ?"}
	values := []any{time.Now().Unix()}
	updateKeywords := false

	if dagIns.EndedAt != 0 {
		setClauses = append(setClauses, "f_ended_at = ?")
		values = append(values, dagIns.EndedAt)
	}

	if dagIns.EventPersistence == 0 {
		if dagIns.ShareData != nil {
			if dagIns.ShareDataExt != nil {
				setClauses = append(setClauses, "f_share_data_ext = ?", "f_share_data = ?")
				values = append(values, marshalToString(dagIns.ShareDataExt), "")
			} else {
				setClauses = append(setClauses, "f_share_data_ext = ?", "f_share_data = ?")
				values = append(values, "", marshalToString(dagIns.ShareData))
			}
		}

		if dagIns.Dump != "" {
			if dagIns.DumpExt != nil {
				setClauses = append(setClauses, "f_dump_ext = ?", "f_dump = ?")
				values = append(values, marshalToString(dagIns.DumpExt), "")
			} else {
				setClauses = append(setClauses, "f_dump_ext = ?", "f_dump = ?")
				values = append(values, "", dagIns.Dump)
			}
		}
	} else {
		setClauses = append(setClauses, "f_event_persistence = ?")
		values = append(values, int(dagIns.EventPersistence))
	}

	if dagIns.EventOssPath != "" {
		setClauses = append(setClauses, "f_event_oss_path = ?")
		values = append(values, dagIns.EventOssPath)
	}
	if dagIns.Status != "" {
		setClauses = append(setClauses, "f_status = ?")
		values = append(values, string(dagIns.Status))
	}
	if utils.IsContain("Cmd", mustsPatchFields) || dagIns.Cmd != nil {
		setClauses = append(setClauses, "f_cmd = ?", "f_has_cmd = ?")
		values = append(values, marshalToString(dagIns.Cmd), dagIns.Cmd != nil)
	}
	if dagIns.Worker != "" {
		setClauses = append(setClauses, "f_worker = ?")
		values = append(values, dagIns.Worker)
	}
	if utils.IsContain("Reason", mustsPatchFields) || dagIns.Reason != "" {
		setClauses = append(setClauses, "f_reason = ?")
		values = append(values, dagIns.Reason)
	}
	if dagIns.ResumeData != "" {
		setClauses = append(setClauses, "f_resume_data = ?")
		values = append(values, dagIns.ResumeData)
	}
	if dagIns.ResumeStatus != "" {
		setClauses = append(setClauses, "f_resume_status = ?")
		values = append(values, string(dagIns.ResumeStatus))
	}
	if dagIns.Source != "" {
		setClauses = append(setClauses, "f_source = ?")
		values = append(values, dagIns.Source)
	}
	if len(dagIns.Keywords) > 0 {
		setClauses = append(setClauses, "f_keywords = ?")
		values = append(values, marshalToString(dagIns.Keywords))
		updateKeywords = true
	}

	sql := fmt.Sprintf("UPDATE t_dag_instance SET %s WHERE f_id = ?", strings.Join(setClauses, ", "))
	values = append(values, dagIns.ID)

	msgStr, _ := jsoniter.MarshalToString(values)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

	if err = d.db.Exec(sql, values...).Error; err != nil {
		return err
	}

	if updateKeywords {
		dagInsID, _ := strconv.ParseUint(dagIns.ID, 10, 64)
		if err = d.replaceDagInstanceKeywords(newCtx, dagInsID, dagIns.Keywords); err != nil {
			return err
		}
	}

	goevent.Publish(&event.DagInstancePatched{
		Payload:         dagIns,
		MustPatchFields: mustsPatchFields,
	})

	return nil
}

// PatchTaskIns implements [DagRepository].
func (d *dag) PatchTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if taskIns.ID == "" {
		return fmt.Errorf("id cannot be empty")
	}

	setClauses := []string{"f_updated_at = ?"}
	values := []any{time.Now().Unix()}

	if taskIns.Status != "" {
		setClauses = append(setClauses, "f_status = ?")
		values = append(values, string(taskIns.Status))
	}
	if taskIns.Reason != "" {
		setClauses = append(setClauses, "f_reason = ?")
		values = append(values, marshalToString(taskIns.Reason))
	}
	if len(taskIns.Traces) > 0 {
		setClauses = append(setClauses, "f_traces = ?")
		values = append(values, marshalToString(taskIns.Traces))
	}
	if taskIns.Results != nil {
		setClauses = append(setClauses, "f_results = ?")
		values = append(values, marshalToString(taskIns.Results))
	}
	if taskIns.LastModifiedAt != 0 {
		setClauses = append(setClauses, "f_last_modified_at = ?")
		values = append(values, taskIns.LastModifiedAt)
	}
	if taskIns.RenderedParams != nil {
		setClauses = append(setClauses, "f_rendered_params = ?")
		values = append(values, marshalToString(taskIns.RenderedParams))
	}
	if taskIns.DependOn != nil {
		setClauses = append(setClauses, "f_depend_on = ?")
		values = append(values, marshalToString(taskIns.DependOn))
	}
	if taskIns.Hash != "" {
		setClauses = append(setClauses, "f_hash = ?")
		values = append(values, taskIns.Hash)
	}
	if taskIns.MetaData != nil {
		setClauses = append(setClauses, "f_metadata = ?")
		values = append(values, marshalToString(taskIns.MetaData))
	}

	sql := fmt.Sprintf("UPDATE t_task_instance SET %s WHERE f_id = ?", strings.Join(setClauses, ", "))
	values = append(values, taskIns.ID)

	msgStr, _ := jsoniter.MarshalToString(values)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

	return d.db.Exec(sql, values...).Error
}

// RemoveClient implements [DagRepository].
func (d *dag) RemoveClient(clientName string) (err error) {
	sql := `DELETE FROM t_client WHERE f_client_name = ?`
	return d.db.Exec(sql, clientName).Error
}

// RetryDagIns implements [DagRepository].
func (d *dag) RetryDagIns(ctx context.Context, dagInsID string, taskInsIDs []string) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	fn := func() error {
		now := time.Now().Unix()
		if len(taskInsIDs) > 0 {
			var ids []uint64
			for _, id := range taskInsIDs {
				v, _ := strconv.ParseUint(id, 10, 64)
				ids = append(ids, v)
			}
			sqlTask := `UPDATE t_task_instance SET f_updated_at = ?, f_status = ? WHERE f_id IN ?`
			if err = d.db.Exec(sqlTask, now, string(entity.TaskInstanceStatusInit), ids).Error; err != nil {
				return err
			}
		}

		sqlDag := `UPDATE t_dag_instance SET f_updated_at = ?, f_status = ?, f_ended_at = ? WHERE f_id = ?`
		return d.db.Exec(sqlDag, now, string(entity.DagInstanceStatusInit), now, dagInsID).Error
	}

	if !d.isTX {
		return d.WithTransaction(newCtx, func(context.Context, mod.Store) error {
			return fn()
		})
	}
	return fn()
}

// SetSwitchStatus implements [DagRepository].
func (d *dag) SetSwitchStatus(status bool) error {
	now := time.Now().Unix()
	sql := `INSERT INTO t_switch (f_id, f_created_at, f_updated_at, f_name, f_status)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE f_status = VALUES(f_status), f_updated_at = VALUES(f_updated_at)`
	id, _ := utils.GetUniqueID()
	return d.db.Exec(sql, id, now, now, entity.SwitchName, status).Error
}

// Unmarshal implements [DagRepository].
func (d *dag) Unmarshal(bytes []byte, ptr interface{}) error {
	return jsoniter.Unmarshal(bytes, ptr)
}

// UpdateDagIncValue implements [DagRepository].
func (d *dag) UpdateDagIncValue(ctx context.Context, dagId string, incKey string, incValue any) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	fn := func() error {
		var raw string
		if err = d.db.Raw(`SELECT f_inc_values FROM t_dag WHERE f_id = ?`, dagId).Scan(&raw).Error; err != nil {
			return err
		}
		updated, uerr := updateJSONMapString(raw, incKey, incValue)
		if uerr != nil {
			return uerr
		}
		return d.db.Exec(`UPDATE t_dag SET f_inc_values = ? WHERE f_id = ?`, updated, dagId).Error
	}

	if !d.isTX {
		return d.WithTransaction(newCtx, func(context.Context, mod.Store) error {
			return fn()
		})
	}
	return fn()
}

// UpdateDagIns implements [DagRepository].
func (d *dag) UpdateDagIns(ctx context.Context, dagIns *entity.DagInstance) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sql := `UPDATE t_dag_instance SET
		f_updated_at = ?, f_dag_id = ?, f_trigger = ?, f_worker = ?, f_source = ?,
		f_vars = ?, f_keywords = ?, f_event_persistence = ?, f_event_oss_path = ?, f_share_data = ?, f_share_data_ext = ?,
		f_status = ?, f_reason = ?, f_cmd = ?, f_has_cmd = ?, f_batch_run_id = ?, f_user_id = ?, f_ended_at = ?, f_dag_type = ?, f_policy_type = ?, f_appinfo = ?,
		f_priority = ?, f_mode = ?, f_dump = ?, f_dump_ext = ?, f_success_callback = ?, f_error_callback = ?, f_call_chain = ?,
		f_resume_data = ?, f_resume_status = ?, f_version = ?, f_version_id = ?, f_biz_domain_id = ?
		WHERE f_id = ?`

	t := ToDagInstanceModel(dagIns, true)
	msgStr, _ := jsoniter.MarshalToString(t)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

	if err = d.db.Exec(sql,
		t.UpdatedAt,
		t.DagID,
		t.Trigger,
		t.Worker,
		t.Source,
		t.Vars,
		t.Keywords,
		t.EventPersistence,
		t.EventOssPath,
		t.ShareData,
		t.ShareDataExt,
		t.Status,
		t.Reason,
		t.Cmd,
		t.HasCmd,
		t.BatchRunID,
		t.UserID,
		t.EndedAt,
		t.DagType,
		t.PolicyType,
		t.AppInfo,
		t.Priority,
		t.Mode,
		t.Dump,
		t.DumpExt,
		t.SuccessCallback,
		t.ErrorCallback,
		t.CallChain,
		t.ResumeData,
		t.ResumeStatus,
		t.Version,
		t.VersionID,
		t.BizDomainID,
		t.ID,
	).Error; err != nil {
		return err
	}

	if err = d.replaceDagInstanceKeywords(newCtx, t.ID, dagIns.Keywords); err != nil {
		return err
	}

	goevent.Publish(&event.DagInstanceUpdated{Payload: dagIns})
	return nil
}

// UpdateTaskIns implements [DagRepository].
func (d *dag) UpdateTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	t := ToTaskInstanceModel(taskIns, true)

	sql := `UPDATE t_task_instance SET
		f_updated_at = ?, f_task_id = ?, f_dag_ins_id = ?, f_name = ?, f_depend_on = ?,
		f_action_name = ?, f_timeout_secs = ?, f_params = ?, f_traces = ?, f_status = ?,
		f_reason = ?, f_pre_checks = ?, f_results = ?, f_steps = ?, f_last_modified_at = ?,
		f_rendered_params = ?, f_hash = ?, f_settings = ?, f_metadata = ?
		WHERE f_id = ?`

	msgStr, _ := jsoniter.MarshalToString(t)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

	return d.db.Exec(sql,
		t.UpdatedAt,
		t.TaskID,
		t.DagInsID,
		t.Name,
		t.DependOn,
		t.ActionName,
		t.TimeoutSecs,
		t.Params,
		t.Traces,
		t.Status,
		t.Reason,
		t.PreChecks,
		t.Results,
		t.Steps,
		t.LastModifiedAt,
		t.RenderedParams,
		t.Hash,
		t.Settings,
		t.MetaData,
		t.ID,
	).Error
}

// UpdateToken implements [DagRepository].
func (d *dag) UpdateToken(token *entity.Token) error {
	baseInfo := token.GetBaseInfo()
	baseInfo.Update()

	sql := `UPDATE t_token SET f_updated_at = ?, f_token = ?, f_expires_in = ? WHERE f_user_id = ?`
	res := d.db.Exec(sql, baseInfo.UpdatedAt, token.Token, token.ExpiresIn, token.UserID)
	if res.Error != nil {
		return fmt.Errorf("update token failed: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("t_token has no key[ %s ] to update: %w", baseInfo.ID, data.ErrDataNotFound)
	}
	return nil
}
