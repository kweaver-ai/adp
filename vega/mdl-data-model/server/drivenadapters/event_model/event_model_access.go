// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package event_model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"data-model/common"
	"data-model/drivenadapters/metric_model"
	derrors "data-model/errors"
	"data-model/interfaces"
)

const (
	EVENT_MODEL_TABLE_NAME    = "t_event_models"
	DETECT_RULE_TABLE_NAME    = "t_event_model_detect_rules"
	AGGREGATE_RULE_TABLE_NAME = "t_event_model_aggregate_rules"
	EVENT_TASK_TABLE_NAME     = "t_event_model_task"
)

var (
	emAccessOnce sync.Once
	emAccess     interfaces.EventModelAccess
)

type eventModelAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
	mma        interfaces.MetricModelAccess
}

func NewEventModelAccess(appSetting *common.AppSetting) interfaces.EventModelAccess {
	emAccessOnce.Do(func() {
		emAccess = &eventModelAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
			mma:        metric_model.NewMetricModelAccess(appSetting),
		}
	})
	return emAccess
}

// 创建事件模型
func (ema *eventModelAccess) CreateEventModels(tx *sql.Tx, ems []interfaces.EventModel) ([]map[string]any, error) {

	//NOTE 特定字段序列化和构造插入sql
	DetectRuleQuery := sq.Insert(DETECT_RULE_TABLE_NAME).
		Columns(
			"f_detect_rule_id",
			"f_detect_rule_type",
			"f_rule_priority",
			"f_formula",
			"f_detect_algo",
			"f_detect_analysis_algo",
			"f_create_time",
			"f_update_time")
	AggregateRuleQuery := sq.Insert(AGGREGATE_RULE_TABLE_NAME).
		Columns(
			"f_aggregate_rule_id",
			"f_aggregate_rule_type",
			"f_rule_priority",
			"f_aggregate_algo",
			"f_aggregate_analysis_algo",
			"f_group_fields",
			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_update_time")
	EventModelQuery := sq.Insert(EVENT_MODEL_TABLE_NAME).
		Columns(
			"f_event_model_id",
			"f_event_model_name",
			"f_event_model_group_name",
			"f_event_model_type",
			"f_event_model_comment",
			"f_event_model_tags",
			"f_data_source_type",
			"f_data_source",
			"f_detect_rule_id",
			"f_aggregate_rule_id",
			"f_default_time_window",
			"f_is_active",
			"f_is_custom",
			"f_enable_subscribe",
			"f_status",
			"f_downstream_dependent_model",
			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_update_time")

	var modelInfo = []map[string]any{}
	var atomicFlag = false
	var aggregateFlag = false
	for _, em := range ems {
		modelInfo = append(modelInfo, map[string]any{
			"id":   em.EventModelID,
			"name": em.EventModelName,
		})
		if em.EventModelType == "atomic" {
			atomicFlag = true
			formulaStr, _ := sonic.Marshal(em.DetectRule.Formula)
			analysisAlgo, _ := sonic.Marshal(em.DetectRule.AnalysisAlgo)
			DetectRuleQuery = DetectRuleQuery.Values(
				em.DetectRule.DetectRuleID,
				em.DetectRule.Type,
				99,
				formulaStr,
				em.DetectRule.DetectAlgo,
				analysisAlgo,
				em.DetectRule.CreateTime,
				em.DetectRule.UpdateTime)
		} else {
			aggregateFlag = true
			GroupFieldsStr, _ := sonic.Marshal(em.AggregateRule.GroupFields)
			analysisAlgo, _ := sonic.Marshal(em.AggregateRule.AnalysisAlgo)
			AggregateRuleQuery = AggregateRuleQuery.Values(
				em.AggregateRule.AggregateRuleID,
				em.AggregateRule.Type,
				99,
				em.AggregateRule.AggregateAlgo,
				analysisAlgo,
				GroupFieldsStr,
				em.AggregateRule.CreateTime,
				em.AggregateRule.UpdateTime)
		}
		//NOTE:这里拼接称字符串插入数据库后依然会转成bigint20,所以换种方式存入。
		if em.DataSource == nil {
			em.DataSource = []string{}
		}
		DataSourceStr, _ := sonic.Marshal(em.DataSource)
		// DataSourceStr := strings.Join(em.DataSource, ",")
		tagsStr := strings.Join(em.EventModelTags, ",")
		DefaultTimeWindowStr, _ := sonic.Marshal(em.DefaultTimeWindow)
		DownstreamDependentModel := strings.Join(em.DownstreamDependentModel, ",")
		EventModelQuery = EventModelQuery.Values(
			em.EventModelID,
			em.EventModelName,
			"",
			em.EventModelType,
			em.EventModelComment,
			tagsStr,
			em.DataSourceType,
			DataSourceStr,
			em.DetectRule.DetectRuleID,
			em.AggregateRule.AggregateRuleID,
			DefaultTimeWindowStr,
			em.IsActive,
			1,
			em.EnableSubscribe,
			em.Status,
			DownstreamDependentModel,
			em.Creator.ID,
			em.Creator.Type,
			em.CreateTime,
			em.UpdateTime)
	}

	//NOTE  start db transaction
	if atomicFlag {
		_, err := DetectRuleQuery.RunWith(tx).Exec()
		if err != nil {
			logger.Errorf("insert rule error: %v\n", err)
			return nil, err
		}
	}
	if aggregateFlag {
		_, err := AggregateRuleQuery.RunWith(tx).Exec()
		if err != nil {
			logger.Errorf("insert rule error: %v\n", err)
			return nil, err
		}
	}

	_, err := EventModelQuery.RunWith(tx).Exec()
	if err != nil {
		logger.Errorf("insert event model  error: %v\n", err)
		return nil, err
	}
	return modelInfo, nil
}

// 按 id 获取事件模型信息
func (ema *eventModelAccess) GetEventModelByID(modelID string) (interfaces.EventModel, error) {

	Query := sq.Select(
		"f_event_model_id",
		"f_event_model_name",
		"f_event_model_type",
		"f_event_model_tags",
		"f_event_model_comment",
		"f_data_source_type",
		"f_data_source",
		"f_detect_rule_id",
		"f_aggregate_rule_id",
		"COALESCE(f_downstream_dependent_model,'')",
		"f_default_time_window",
		"f_is_active",
		"f_is_custom",
		"f_enable_subscribe",
		"f_status",
		"f_create_time",
		"f_update_time",
	).From(EVENT_MODEL_TABLE_NAME).
		Where(sq.Eq{"f_event_model_id": modelID}).
		RunWith(ema.db)

	querySql, _, _ := Query.ToSql()
	logger.Debugf("query_sql: %v\n", querySql)

	var EventModelTagsStr, DefaultTimeWindowStr, DataSourceStr, DownstreamDependentModel string
	em := interfaces.EventModel{}

	rows := Query.QueryRow()
	err := rows.Scan(
		&em.EventModelID,
		&em.EventModelName,
		&em.EventModelType,
		&EventModelTagsStr,
		&em.EventModelComment,
		&em.DataSourceType,
		&DataSourceStr,
		&em.DetectRule.DetectRuleID,
		&em.AggregateRule.AggregateRuleID,
		&DownstreamDependentModel,
		&DefaultTimeWindowStr,
		&em.IsActive,
		&em.IsCustom,
		&em.EnableSubscribe,
		&em.Status,
		&em.CreateTime,
		&em.UpdateTime)

	//NOTE 未找到对应的事件模型
	if err == sql.ErrNoRows {
		logger.Errorf("event model not found,ID: %v\n", em.EventModelID)
		return interfaces.EventModel{}, errors.New(derrors.EventModel_EventModelNotFound)

	}
	//NOTE 查询发生了错误
	if err != nil {
		logger.Errorf("query execute failed: %v\n", querySql)
		logger.Errorf("query execute error: %v\n", err)
		return interfaces.EventModel{}, errors.New(derrors.EventModel_InternalError)
	}
	//NOTE 将空字符串转换为空数组-> 前端
	if EventModelTagsStr == "" {
		em.EventModelTags = []string{}
	} else {
		em.EventModelTags = strings.Split(EventModelTagsStr, ",")
	}
	if DownstreamDependentModel == "" {
		em.DownstreamDependentModel = []string{}
	} else {
		em.DownstreamDependentModel = strings.Split(DownstreamDependentModel, ",")
	}
	err = sonic.Unmarshal([]byte(DataSourceStr), &em.DataSource)
	if err != nil {
		//TODO ADD LOG
		logger.Errorf(" serial data source of  event model failed: %v", err.Error())
		return interfaces.EventModel{}, errors.New(derrors.EventModel_InternalError)
	}

	err = sonic.Unmarshal([]byte(DefaultTimeWindowStr), &em.DefaultTimeWindow)

	if err != nil {
		//TODO ADD LOG
		logger.Errorf(" serial time windows of  event model failed: %v", err.Error())
		return interfaces.EventModel{}, errors.New(derrors.EventModel_InternalError)
	}

	if em.EventModelType == "atomic" {
		RQuery := sq.Select(
			"f_detect_rule_id",
			"f_detect_rule_type",
			"f_formula",
			"COALESCE(f_detect_algo,'')",
			"COALESCE(f_detect_analysis_algo,'{}')",
			"f_rule_priority",
			"f_create_time",
			"f_update_time",
		).From(DETECT_RULE_TABLE_NAME).
			Where(sq.Eq{"f_detect_rule_id": em.DetectRule.DetectRuleID}).
			RunWith(ema.db)

		dr := interfaces.DetectRule{}
		var formulaStr string
		var analysisAlgoStr string

		drows := RQuery.QueryRow()
		err = drows.Scan(
			&dr.DetectRuleID,
			&dr.Type,
			&formulaStr,
			&dr.DetectAlgo,
			&analysisAlgoStr,
			&dr.Priority,
			&dr.CreateTime,
			&dr.UpdateTime)
		if err != nil {
			return interfaces.EventModel{}, errors.New(derrors.EventModel_InternalError)
		}
		err = sonic.Unmarshal([]byte(formulaStr), &dr.Formula)
		if err != nil {
			return interfaces.EventModel{}, err
		}

		err = sonic.Unmarshal([]byte(analysisAlgoStr), &dr.AnalysisAlgo)
		if err != nil {
			logger.Errorf("event detect analysis algo parse failed,param: %v", analysisAlgoStr)
			return interfaces.EventModel{}, err
		}
		//NOTE 将检测规则嵌入事件模型

		em.DetectRule = dr
		em.AggregateRule = interfaces.AggregateRule{}

	} else {
		RQuery := sq.Select(
			"f_aggregate_rule_id",
			"f_aggregate_rule_type",
			"f_aggregate_algo",
			"f_aggregate_analysis_algo",
			"f_group_fields",
			"f_rule_priority",
			"f_create_time",
			"f_update_time",
		).From(AGGREGATE_RULE_TABLE_NAME).
			Where(sq.Eq{"f_aggregate_rule_id": em.AggregateRule.AggregateRuleID}).
			RunWith(ema.db)

		ar := interfaces.AggregateRule{}
		var GroupFieldsStr string
		var analysisAlgo string

		drows := RQuery.QueryRow()
		err = drows.Scan(
			&ar.AggregateRuleID,
			&ar.Type,
			&ar.AggregateAlgo,
			&analysisAlgo,
			&GroupFieldsStr,
			&ar.Priority,
			&ar.CreateTime,
			&ar.UpdateTime)
		if err != nil {
			return interfaces.EventModel{}, errors.New(derrors.EventModel_InternalError)
		}
		//NOTE 将聚合规则嵌入事件模型ar
		err = sonic.Unmarshal([]byte(GroupFieldsStr), &ar.GroupFields)
		if err != nil {
			return interfaces.EventModel{}, err
		}
		err = sonic.Unmarshal([]byte(analysisAlgo), &ar.AnalysisAlgo)
		if err != nil {
			logger.Errorf("event detect analysis algo parse failed,param: %v", analysisAlgo)
			return interfaces.EventModel{}, err
		}
		em.AggregateRule = ar
		em.DetectRule = interfaces.DetectRule{}
	}

	return em, nil
}

// 修改事件模型
func (ema *eventModelAccess) UpdateEventModel(tx *sql.Tx, em interfaces.EventModel) error {

	tagsStr := strings.Join(em.EventModelTags, ",")
	DefaultTimeWindowStr, _ := sonic.Marshal(em.DefaultTimeWindow)
	DataSourceStr, _ := sonic.Marshal(em.DataSource)

	//构造sql
	model_sql, model_args, _ := sq.Update(EVENT_MODEL_TABLE_NAME).
		Set("f_event_model_name", em.EventModelName).
		Set("f_event_model_tags", tagsStr).
		Set("f_event_model_comment", em.EventModelComment).
		Set("f_data_source", DataSourceStr).
		Set("f_data_source_type", em.DataSourceType).
		Set("f_default_time_window", DefaultTimeWindowStr).
		Set("f_is_active", em.IsActive).
		Set("f_enable_subscribe", em.EnableSubscribe).
		Set("f_status", em.Status).
		Set("f_update_time", em.UpdateTime).
		Set("f_downstream_dependent_model", strings.Join(em.DownstreamDependentModel, ",")).
		Where(sq.Eq{"f_event_model_id": em.EventModelID}).ToSql()
	// RunWith(ema.db)
	formulaStr, _ := sonic.Marshal(em.DetectRule.Formula)
	detectAnalysisAlgo, _ := sonic.Marshal(em.DetectRule.AnalysisAlgo)

	drule_sql, drule_args, _ := sq.Update(DETECT_RULE_TABLE_NAME).
		Set("f_detect_rule_type", em.DetectRule.Type).
		Set("f_formula", formulaStr).
		Set("f_detect_algo", em.DetectRule.DetectAlgo).
		Set("f_detect_analysis_algo", detectAnalysisAlgo).
		Set("f_update_time", em.DetectRule.UpdateTime).
		Where(sq.Eq{"f_detect_rule_id": em.DetectRule.DetectRuleID}).ToSql()

	GroupFieldsStr, _ := sonic.Marshal(em.AggregateRule.GroupFields)
	aggregateAnalysisAlgo, _ := sonic.Marshal(em.AggregateRule.AnalysisAlgo)
	arule_sql, arule_args, _ := sq.Update(AGGREGATE_RULE_TABLE_NAME).
		Set("f_aggregate_rule_type", em.AggregateRule.Type).
		Set("f_aggregate_algo", em.AggregateRule.AggregateAlgo).
		Set("f_aggregate_analysis_algo", aggregateAnalysisAlgo).
		Set("f_group_fields", GroupFieldsStr).
		Set("f_update_time", em.AggregateRule.UpdateTime).
		Where(sq.Eq{"f_aggregate_rule_id": em.AggregateRule.AggregateRuleID}).ToSql()

	//NOTE 开启事务

	_, err := tx.Exec(model_sql, model_args...)
	if err != nil {
		logger.Errorf("update event model error: %v\n", err)
		return err
	}
	if em.EventModelType == "atomic" {
		_, err = tx.Exec(drule_sql, drule_args...)
		if err != nil {
			logger.Errorf("update  detect rule error: %v\n", err)
			return err
		}
	} else if em.EventModelType == "aggregate" {
		_, err = tx.Exec(arule_sql, arule_args...)
		if err != nil {
			logger.Errorf("update  aggregate rule error: %v\n", err)
			return err
		}
	}
	return nil
}

// 删除事件模型
func (ema *eventModelAccess) DeleteEventModels(tx *sql.Tx, ems []interfaces.EventModel) error {
	//NOTE transaction process
	for _, em := range ems {
		drule_sql, drule_args, err := sq.Delete(DETECT_RULE_TABLE_NAME).
			Where(sq.Eq{"f_detect_rule_id": em.DetectRule.DetectRuleID}).
			ToSql()
		if err != nil {
			logger.Errorf("delete detect rule error: %v\n", err)
			return err
		}

		arule_sql, arule_args, err := sq.Delete(AGGREGATE_RULE_TABLE_NAME).
			Where(sq.Eq{"f_aggregate_rule_id": em.AggregateRule.AggregateRuleID}).
			ToSql()
		if err != nil {
			logger.Errorf("delete aggregate rule error: %v\n", err)
			return err
		}

		model_sql, model_args, err := sq.Delete(EVENT_MODEL_TABLE_NAME).
			Where(sq.Eq{"f_event_model_id": em.EventModelID}).
			ToSql()
		if err != nil {
			logger.Errorf("delete event model error: %v\n", err)
			return err
		}

		_, err = tx.Exec(drule_sql, drule_args...)
		if err != nil {
			logger.Errorf("delete  detect rule error: %v\n", err)
			return err
		}
		_, err = tx.Exec(arule_sql, arule_args...)
		if err != nil {
			logger.Errorf("delete  aggregate rule error: %v\n", err)
			return err
		}

		_, err = tx.Exec(model_sql, model_args...)
		if err != nil {
			logger.Errorf("delete event model error: %v\n", err)
			return err
		}
	}

	return nil
}

// 查询指标模型列表
func (ema *eventModelAccess) QueryEventModels(ctx context.Context, params interfaces.EventModelQueryRequest) ([]interfaces.EventModel, error) {

	Query := sq.Select(
		"f_event_model_id",
		"f_event_model_name",
		"f_event_model_type",
		"f_event_model_tags",
		"f_event_model_comment",
		"f_data_source_type",
		"f_data_source",
		"f_detect_rule_id",
		"f_aggregate_rule_id",
		"COALESCE(f_downstream_dependent_model,'')",
		"COALESCE(f_detect_rule_type,'')",
		"COALESCE(f_detect_algo,'')",
		"COALESCE(f_formula,'')",
		"COALESCE(f_aggregate_rule_type, '')",
		"COALESCE(f_aggregate_algo,'')",
		"COALESCE(f_group_fields,'')",
		"f_default_time_window",
		"f_is_active",
		"f_is_custom",
		"f_enable_subscribe",
		"COALESCE(f_detect_analysis_algo,'{}')",
		"COALESCE(f_aggregate_analysis_algo,'{}')",
		"f_status",
		fmt.Sprintf("%s.f_create_time", EVENT_MODEL_TABLE_NAME),
		fmt.Sprintf("%s.f_update_time", EVENT_MODEL_TABLE_NAME),
		fmt.Sprintf("COALESCE(%s.f_create_time,0)", DETECT_RULE_TABLE_NAME),
		fmt.Sprintf("COALESCE(%s.f_update_time,0)", DETECT_RULE_TABLE_NAME),
		fmt.Sprintf("COALESCE(%s.f_create_time,0)", AGGREGATE_RULE_TABLE_NAME),
		fmt.Sprintf("COALESCE(%s.f_update_time,0)", AGGREGATE_RULE_TABLE_NAME),
		"COALESCE(f_task_id,0)",
		"COALESCE(f_model_id,0)",
		"COALESCE(f_storage_config,'{}')",
		"COALESCE(f_schedule,'{}')",
		"COALESCE(f_dispatch_config,'{}')",
		"COALESCE(f_execute_parameter,'{}')",
		"COALESCE(f_task_status,0)",
		"COALESCE(f_error_details,'')",
		"COALESCE(f_downstream_dependent_task,'')",
		"COALESCE(f_status_update_time,'')",
		"COALESCE(f_schedule_sync_status,0)",
		fmt.Sprintf("COALESCE(%s.f_create_time,0)", EVENT_TASK_TABLE_NAME),
		fmt.Sprintf("COALESCE(%s.f_update_time,0)", EVENT_TASK_TABLE_NAME),
	).From(EVENT_MODEL_TABLE_NAME).
		LeftJoin(fmt.Sprintf("%s USING (f_detect_rule_id)", DETECT_RULE_TABLE_NAME)).
		LeftJoin(fmt.Sprintf("%s USING (f_aggregate_rule_id)", AGGREGATE_RULE_TABLE_NAME)).
		LeftJoin(fmt.Sprintf("%s ON %s.f_event_model_id = %s.f_model_id", EVENT_TASK_TABLE_NAME, EVENT_MODEL_TABLE_NAME, EVENT_TASK_TABLE_NAME)).
		OrderBy(fmt.Sprintf("%s.f_", EVENT_MODEL_TABLE_NAME) + params.SortKey + " " + params.Direction)

	// if params.Limit != -1 {
	// 	Query = Query.Limit(uint64(params.Limit)).Offset(uint64(params.Offset))
	// }
	if params.EventModelNamePattern != "" {
		name := strings.Replace(params.EventModelNamePattern, "%", "\\%", -1)
		name = strings.Replace(name, "_", "\\_", -1)
		Query = Query.Where(sq.Like{"f_event_model_name": fmt.Sprint("%", name, "%")})
	}
	if params.EventModelName != "" {
		Query = Query.Where(sq.Eq{"f_event_model_name": params.EventModelName})
	}

	if params.EventModelType != "" {
		EventModelTypes := strings.Split(params.EventModelType, ",")
		Query = Query.Where(sq.Eq{"f_event_model_type": EventModelTypes})

	}
	if params.DetectType != "" {
		DetectType := strings.Split(params.DetectType, ",")
		Query = Query.Where(sq.Eq{"f_detect_rule_type": DetectType})
	}
	if params.AggregateType != "" {
		AggregateType := strings.Split(params.AggregateType, ",")
		Query = Query.Where(sq.Eq{"f_aggregate_rule_type": AggregateType})
	}
	if params.EventModelTag != "" {
		Query = Query.Where(sq.Expr("instr(f_event_model_tags, ?) > 0", params.EventModelTag))
	}
	if params.DataSourceType != "" {
		DataSourceType := strings.Split(params.DataSourceType, ",")
		Query = Query.Where(sq.Eq{"f_data_source_type": DataSourceType})
	}

	//NOTE： 此处目前只有单个数据源的原子事件查找请求，所以直接匹配即可。
	if params.DataSource != "" {
		Query = Query.Where(sq.Eq{"f_data_source": params.DataSource})
	}

	if params.IsActive != "" {
		isActive := strings.Split(params.IsActive, ",")
		Query = Query.Where(sq.Eq{"f_is_active": isActive})
	}

	if params.EnableSubscribe != "" {
		enableSubscribe := strings.Split(params.EnableSubscribe, ",")
		Query = Query.Where(sq.Eq{"f_enable_subscribe": enableSubscribe})
	}

	if params.Status != "" {
		status := strings.Split(params.Status, ",")
		Query = Query.Where(sq.Eq{"f_status": status})
	}

	if params.IsCustom != -1 {
		Query = Query.Where(sq.Eq{"f_is_custom": params.IsCustom})
	}
	if params.TaskStatus != "" {
		taskStatus := strings.Split(params.TaskStatus, ",")
		Query = Query.Where(sq.Eq{"f_task_status": taskStatus})
	}
	if params.ScheduleSyncStatus != "" {
		syncStatus := strings.Split(params.ScheduleSyncStatus, ",")
		Query = Query.Where(sq.Eq{"f_schedule_sync_status": syncStatus})
	}

	eventModels := []interfaces.EventModel{}

	msql, args, _ := Query.ToSql()
	// fmt.Printf("msql: %v\n", msql)
	logger.Debugf("Query sql: %v,args:%v", msql, args)
	rows, err := Query.RunWith(ema.db).Query()
	// rows, err := ema.db.Query(msql, args)

	if err != nil {
		logger.Errorf("query event model error: %v\n", err)
		return eventModels, err
	}
	defer rows.Close()
	for rows.Next() {
		var em interfaces.EventModel
		var EventModelTagsStr string
		var DefaultTimeWindowStr string
		var DataSourceStr string
		var formula string
		var GroupFieldsStr string
		var DownstreamDependentModel string
		var DownstreamDependentTask string
		var DetectAnalysisAlgo string
		var AggregateAnalysisAlgo string
		var scheduleBytes, storageConfigBytes, dispatchConfigBytes, executeParameterBytes []byte

		err := rows.Scan(
			&em.EventModelID,
			&em.EventModelName,
			&em.EventModelType,
			&EventModelTagsStr,
			&em.EventModelComment,
			&em.DataSourceType,
			&DataSourceStr,
			&em.DetectRule.DetectRuleID,
			&em.AggregateRule.AggregateRuleID,
			&DownstreamDependentModel,
			&em.DetectRule.Type,
			&em.DetectRule.DetectAlgo,
			&formula,
			&em.AggregateRule.Type,
			&em.AggregateRule.AggregateAlgo,
			&GroupFieldsStr,
			&DefaultTimeWindowStr,
			&em.IsActive,
			&em.IsCustom,
			&em.EnableSubscribe,
			&DetectAnalysisAlgo,
			&AggregateAnalysisAlgo,
			&em.Status,
			&em.CreateTime,
			&em.UpdateTime,
			&em.DetectRule.CreateTime,
			&em.DetectRule.UpdateTime,
			&em.AggregateRule.CreateTime,
			&em.AggregateRule.UpdateTime,
			&em.Task.TaskID,
			&em.Task.ModelID,
			&storageConfigBytes,
			&scheduleBytes,
			&dispatchConfigBytes,
			&executeParameterBytes,
			&em.Task.TaskStatus,
			&em.Task.ErrorDetails,
			&DownstreamDependentTask,
			&em.Task.StatusUpdateTime,
			&em.Task.ScheduleSyncStatus,
			&em.Task.CreateTime,
			&em.Task.UpdateTime)
		if err != nil {
			logger.Errorf("event_model scan failed,%v", err.Error())
			return eventModels, err
		}
		//NOTE: 特定字段处理
		if EventModelTagsStr == "" {
			em.EventModelTags = []string{}
		} else {
			em.EventModelTags = strings.Split(EventModelTagsStr, ",")
		}

		if formula == "" {
			em.DetectRule.Formula = []interfaces.FormulaItem{}
		} else {
			_ = sonic.Unmarshal([]byte(formula), &em.DetectRule.Formula)
		}

		err = sonic.Unmarshal([]byte(DefaultTimeWindowStr), &em.DefaultTimeWindow)
		if err != nil {
			//TODO ADD LOG
			logger.Errorf("event_model parse failed,%v", err.Error())
			logger.Errorf("event_model parse failed,param:%v", DefaultTimeWindowStr)
			return eventModels, err
		}

		err = sonic.Unmarshal([]byte(DataSourceStr), &em.DataSource)
		if err != nil {
			//TODO ADD LOG
			logger.Errorf("event_model parse failed,%v", err.Error())
			logger.Errorf("event_model parse failed,param:%v", DataSourceStr)
			return eventModels, err
		}

		err = sonic.Unmarshal([]byte(DetectAnalysisAlgo), &em.DetectRule.AnalysisAlgo)
		if err != nil {
			//TODO ADD LOG
			logger.Errorf("event_model detect analysis algo parse failed,%v", err.Error())
			logger.Errorf("event_model detect analysis algo  parse failed,param:%v", DetectAnalysisAlgo)
			return eventModels, err
		}

		if DownstreamDependentModel == "" {
			em.DownstreamDependentModel = []string{}
		} else {
			em.DownstreamDependentModel = strings.Split(DownstreamDependentModel, ",")
		}
		if DownstreamDependentTask == "" {
			em.Task.DownstreamDependentTask = []string{}
		} else {
			em.Task.DownstreamDependentTask = strings.Split(DownstreamDependentTask, ",")
		}

		err = sonic.Unmarshal([]byte(DetectAnalysisAlgo), &em.DetectRule.AnalysisAlgo)
		if err != nil {
			//TODO ADD LOG
			logger.Errorf("detect rule analysis algo  parse failed,%v", err.Error())
			logger.Errorf("detect rule analysis algo  parse failed,DetectAnalysisAlgo:%v", DetectAnalysisAlgo)
			return eventModels, err
		}

		err = sonic.Unmarshal([]byte(AggregateAnalysisAlgo), &em.AggregateRule.AnalysisAlgo)
		if err != nil {
			//TODO ADD LOG
			logger.Errorf("aggregate rule analysis algo  parse failed,%v", err.Error())
			logger.Errorf("aggregate rule analysis algo  parse failed,AggregateAnalysisAlgo:%v", AggregateAnalysisAlgo)
			return eventModels, err
		}

		if GroupFieldsStr == "" {
			em.AggregateRule.GroupFields = []string{}
		} else {
			err = sonic.Unmarshal([]byte(GroupFieldsStr), &em.AggregateRule.GroupFields)
			if err != nil {
				//TODO ADD LOG
				logger.Errorf("event_model parse failed,%v", err.Error())
				logger.Errorf("event_model parse failed,param:%v", GroupFieldsStr)
				return eventModels, err
			}
		}

		ModelNames := make([]string, 0, len(em.DataSource))
		GroupNames := make([]string, 0, len(em.DataSource))
		if em.DataSourceType == "metric_model" {
			IDNameMap, _ := ema.mma.GetMetricModelSimpleInfosByIDs(ctx, em.DataSource)
			for _, value := range IDNameMap {
				ModelNames = append(ModelNames, value.ModelName)
				GroupNames = append(GroupNames, value.GroupName)
			}
		} else if em.DataSourceType == "event_model" {
			ModelNames, _ = ema.GetEventModelNamesByIDs(em.DataSource)
		}

		em.DataSourceName = ModelNames
		em.DataSourceGroupName = GroupNames
		err = sonic.Unmarshal(scheduleBytes, &em.Task.Schedule)
		if err != nil {
			logger.Errorf("Failed to unmarshal schedule when getting event model, err: %v", err.Error())
			return eventModels, err
		}
		err = sonic.Unmarshal(storageConfigBytes, &em.Task.StorageConfig)
		if err != nil {
			logger.Errorf("Failed to unmarshal storageConfig when getting event model, err: %v", err.Error())
			return eventModels, err
		}
		err = sonic.Unmarshal(dispatchConfigBytes, &em.Task.DispatchConfig)
		if err != nil {
			logger.Errorf("Failed to unmarshal dispatchConfig when getting event model, err: %v", err.Error())
			return eventModels, err
		}
		err = sonic.Unmarshal(executeParameterBytes, &em.Task.ExecuteParameter)
		if err != nil {
			logger.Errorf("Failed to unmarshal executeParameter when getting event model, err: %v", err.Error())
			return eventModels, err
		}
		eventModels = append(eventModels, em)
	}
	return eventModels, nil
}

// 查询事件模型总数
func (ema *eventModelAccess) QueryTotalNumberEventModels(params interfaces.EventModelQueryRequest) (int, error) {
	total := 0
	Query := sq.Select("count(f_event_model_id)").
		From(EVENT_MODEL_TABLE_NAME).
		LeftJoin(fmt.Sprintf("%s USING (f_detect_rule_id)", DETECT_RULE_TABLE_NAME)).
		LeftJoin(fmt.Sprintf("%s USING (f_aggregate_rule_id)", AGGREGATE_RULE_TABLE_NAME)).
		LeftJoin(fmt.Sprintf("%s ON %s.f_event_model_id = %s.f_model_id", EVENT_TASK_TABLE_NAME, EVENT_MODEL_TABLE_NAME, EVENT_TASK_TABLE_NAME))

	//NOTE： 和查询函数的过滤条件保持一致
	if params.EventModelNamePattern != "" {
		Query = Query.Where(sq.Like{"f_event_model_name": fmt.Sprint("%", params.EventModelNamePattern, "%")})
	}
	if params.EventModelName != "" {
		Query = Query.Where(sq.Eq{"f_event_model_name": params.EventModelName})
	}

	// EventModelType        string `json:"type" form:"type" binding:"omitempty"`
	// EventModelTags        string `json:"tags" form:"tags" binding:"omitempty,max=5,dive,max=40"`
	// DataSourceType        string `json:"data_source_type" form:"data_source_type"`
	// DataSource            string `json:"data_source" form:"data_source"`
	if params.EventModelType != "" {
		EventModelTypes := strings.Split(params.EventModelType, ",")
		Query = Query.Where(sq.Eq{"f_event_model_type": EventModelTypes})

	}
	if params.DetectType != "" {
		DetectType := strings.Split(params.DetectType, ",")
		Query = Query.Where(sq.Eq{"f_detect_rule_type": DetectType})
	}
	if params.AggregateType != "" {
		AggregateType := strings.Split(params.AggregateType, ",")
		Query = Query.Where(sq.Eq{"f_aggregate_rule_type": AggregateType})
	}

	if params.EventModelTag != "" {
		Query = Query.Where(sq.Expr("instr(f_event_model_tags, ?) > 0", params.EventModelTag))
	}

	if params.DataSourceType != "" {
		DataSourceTypes := strings.Split(params.DataSourceType, ",")
		Query = Query.Where(sq.Eq{"f_data_source_type": DataSourceTypes})
	}

	//NOTE： 此处目前只有单个数据源的原子事件查找请求，所以直接匹配即可。
	if params.DataSource != "" {

		Query = Query.Where(sq.Eq{"f_data_source": params.DataSource})
	}

	if params.IsActive != "" {
		isActive := strings.Split(params.IsActive, ",")
		Query = Query.Where(sq.Eq{"f_is_active": isActive})
	}

	if params.IsCustom != -1 {
		Query = Query.Where(sq.Eq{"f_is_custom": params.IsCustom})
	}

	if params.EnableSubscribe != "" {
		Query = Query.Where(sq.Eq{"f_enable_subscribe": params.EnableSubscribe})
	}

	if params.Status != "" {
		Query = Query.Where(sq.Eq{"f_status": params.Status})
	}
	if params.TaskStatus != "" {
		taskStatus := strings.Split(params.TaskStatus, ",")
		Query = Query.Where(sq.Eq{"f_task_status": taskStatus})
	}
	if params.ScheduleSyncStatus != "" {
		syncStatus := strings.Split(params.ScheduleSyncStatus, ",")
		Query = Query.Where(sq.Eq{"f_schedule_sync_status": syncStatus})
	}

	rows := Query.RunWith(ema.db).QueryRow()
	err := rows.Scan(&total)
	if err != nil {
		logger.Errorf("query event model total error: %v\n", err)
		return 0, err
	}

	return total, nil
}

func (ema *eventModelAccess) GetEventModelNamesByIDs(modelIDs []string) ([]string, error) {
	// 1. 初始化modelMap
	var names = make([]string, 0, len(modelIDs))

	// 2. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Select("COALESCE(f_event_model_name,'')").
		From(EVENT_MODEL_TABLE_NAME).
		Where(sq.Eq{"f_event_model_id": modelIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to generate a sql statement using the squirrel sdk before get event model names by ids, err: %v", err.Error())
		return nil, err
	}

	// 3. debug日志级别下, 打印完整的sql语句
	logger.Debugf("The detailed sql statement when getting event model names by ids is: %v", sqlStr)

	// 4. 执行查询操作
	rows, err := ema.db.Query(sqlStr, args...)
	if err != nil {
		logger.Errorf("Failed to get event model names by ids, err: %v", err.Error())
		return nil, err
	}

	defer rows.Close()

	// 5. Scan查询结果
	for rows.Next() {
		var modelName string

		err := rows.Scan(
			&modelName,
		)

		if err != nil {
			logger.Errorf("Failed to scan row after executing the sql to get the event model names by ids, err: %v", err.Error())
			return nil, err
		}

		names = append(names, modelName)
	}

	return names, nil
}

// 根据事件模型名称数组获取名称与ID的映射关系
func (ema *eventModelAccess) GetEventModelMapByNames(modelNames []string) (map[string]string, error) {
	// 1. 初始化modelMap
	modelMap := make(map[string]string)

	// 2. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Select(
		"f_event_model_name",
		"f_event_model_id").
		From(EVENT_MODEL_TABLE_NAME).
		Where(sq.Eq{"f_event_model_name": modelNames}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to generate a sql statement using the squirrel sdk before get event model ids by names, err: %v", err.Error())
		return modelMap, err
	}

	// 3. debug日志级别下, 打印完整的sql语句
	logger.Debugf("The detailed sql statement when getting event model ids by names is: %v", sqlStr)

	// 4. 执行查询操作
	rows, err := ema.db.Query(sqlStr, args...)
	if err != nil {
		logger.Errorf("Failed to get event model ids by names, err: %v", err.Error())
		return modelMap, err
	}

	defer rows.Close()

	// 5. Scan查询结果
	for rows.Next() {
		var modelName string
		var modelID string

		err := rows.Scan(
			&modelName,
			&modelID,
		)

		if err != nil {
			logger.Errorf("Failed to scan row after executing the sql to get the event model ids by names, err: %v", err.Error())
			return modelMap, err
		}

		modelMap[modelName] = modelID
	}

	return modelMap, nil
}

// 根据事件模型ID数组获取ID与名称的映射关系
func (ema *eventModelAccess) GetEventModelMapByIDs(modelIDs []string) (map[string]string, error) {
	// 1. 初始化modelMap
	modelMap := make(map[string]string)

	// 2. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Select(
		"f_event_model_id",
		"f_event_model_name").
		From(EVENT_MODEL_TABLE_NAME).
		Where(sq.Eq{"f_event_model_id": modelIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to generate a sql statement using the squirrel sdk before get event model names by ids, err: %v", err.Error())
		return modelMap, err
	}

	// 3. debug日志级别下, 打印完整的sql语句
	logger.Debugf("The detailed sql statement when getting event model names by ids is: %v", sqlStr)

	// 4. 执行查询操作
	rows, err := ema.db.Query(sqlStr, args...)
	if err != nil {
		logger.Errorf("Failed to get event model names by ids, err: %v", err.Error())
		return modelMap, err
	}

	defer rows.Close()

	// 5. Scan查询结果
	for rows.Next() {
		var modelID string
		var modelName string

		err := rows.Scan(
			&modelID,
			&modelName,
		)

		if err != nil {
			logger.Errorf("Failed to scan row after executing the sql to get the event model names by ids, err: %v", err.Error())
			return modelMap, err
		}

		modelMap[modelID] = modelName
	}

	return modelMap, nil
}

func (ema *eventModelAccess) GetEventModelRefsByID(modelID string) (int, error) {

	Query := sq.Select("count(*)").
		From(EVENT_MODEL_TABLE_NAME).
		Where(sq.Like{"f_data_source": "%" + modelID + "%"}).
		RunWith(ema.db)

	rows := Query.RunWith(ema.db).QueryRow()
	var refs int
	err := rows.Scan(&refs)
	if err != nil {
		//TODO ADD LOG
		logger.Errorf("event_model parse failed,%v", err.Error())
		return 0, err
	}
	return refs, nil
}

func (ema *eventModelAccess) GetEventModelDependenceByID(modelID string) (int, error) {

	Query := sq.Select("count(*)").
		From(EVENT_MODEL_TABLE_NAME).
		Where(sq.Like{"f_downstream_dependent_model": "%" + modelID + "%"}).
		RunWith(ema.db)

	rows := Query.RunWith(ema.db).QueryRow()
	var refs int
	err := rows.Scan(&refs)
	if err != nil {
		//TODO ADD LOG
		logger.Errorf("event_model parse failed,%v", err.Error())
		return 0, err
	}
	return refs, nil
}

// 更新任务的执行状态
func (ema *eventModelAccess) UpdateEventTaskAttributes(ctx context.Context, task interfaces.EventTask) error {

	data := map[string]interface{}{
		"f_task_status":        task.TaskStatus,
		"f_error_details":      task.ErrorDetails,
		"f_status_update_time": task.StatusUpdateTime,
	}

	sqlStr, vals, err := sq.Update(EVENT_TASK_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_task_id": task.TaskID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of update event model task by task_id, error: %s", err.Error())

		return err
	}
	// 记录处理的 sql 字符串
	logger.Debugf(fmt.Sprintf("修改事件模型任务执行状态的 sql 语句: %s", sqlStr))

	ret, err := ema.db.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("update event model task error: %v\n", err)
		return err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
	}

	if RowsAffected > 1 {
		// 影响行数不等于1不报错，更新操作已经发生
		logger.Errorf("UPDATE %d RowsAffected more than 1, RowsAffected is %d, eventTask is %v",
			task.TaskID, RowsAffected, task)
	}

	return nil
}

func (ema *eventModelAccess) CreateEventTask(ctx context.Context, tx *sql.Tx, task interfaces.EventTask) error {
	// 1 初始化sql语句
	sqlBuilder := sq.Insert(EVENT_TASK_TABLE_NAME).
		Columns(
			"f_task_id",
			"f_model_id",
			"f_schedule",
			"f_dispatch_config",
			"f_execute_parameter",
			"f_storage_config",
			"f_task_status",
			"f_status_update_time",
			"f_error_details",
			"f_schedule_sync_status",
			"f_downstream_dependent_task",
			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_update_time")

	// 2.0 序列化schedule/dispatch_config/execute_parameter
	scheduleBytes, err := sonic.Marshal(task.Schedule)
	if err != nil {
		logger.Errorf("Failed to marshal schedule, err: %v", err.Error())
		return err
	}
	storageConfigBytes, err := sonic.Marshal(task.StorageConfig)
	if err != nil {
		logger.Errorf("Failed to marshal storageConfig, err: %v", err.Error())
		return err
	}
	dispatchConfigBytes, err := sonic.Marshal(task.DispatchConfig)
	if err != nil {
		logger.Errorf("Failed to marshal dispatchConfig, err: %v", err.Error())
		return err
	}
	executeParameterBytes, err := sonic.Marshal(task.ExecuteParameter)
	if err != nil {
		logger.Errorf("Failed to marshal dispatchConfig, err: %v", err.Error())
		return err
	}
	DownstreamDependent := strings.Join(task.DownstreamDependentTask, ",")
	// 2.1 追加参数
	sqlBuilder = sqlBuilder.Values(
		task.TaskID,
		task.ModelID,
		scheduleBytes,
		dispatchConfigBytes,
		executeParameterBytes,
		storageConfigBytes,
		task.TaskStatus,
		task.StatusUpdateTime,
		task.ErrorDetails,
		task.ScheduleSyncStatus,
		DownstreamDependent,
		task.Creator.ID,
		task.Creator.Type,
		task.CreateTime,
		task.UpdateTime)

	// 3. 生成完整的sql语句和参数列表
	sqlStr, args, err := sqlBuilder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of insert event task, error: %s", err.Error())
		return err
	}
	// 记录处理的 sql 字符串
	logger.Debugf(fmt.Sprintf("创建事件模型任务的 sql 语句: %s", sqlStr))

	// 执行批量insert
	_, err = tx.Exec(sqlStr, args...)
	if err != nil {
		logger.Errorf("insert data error: %v\n", err)
		return err
	}

	return nil
}

// 获取事件模型下的任务id
// 获取事件模型下的任务id
func (ema *eventModelAccess) GetEventTaskIDByModelIDs(ctx context.Context, modelIDs []string) ([]string, error) {

	sqlStr, vals, err := sq.Select("f_task_id").
		From(EVENT_TASK_TABLE_NAME).
		Where(sq.Eq{"f_model_id": modelIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select task, error: %s", err.Error())
		return nil, err
	}

	// 记录处理的 sql 字符串
	logger.Debugf(fmt.Sprintf("查询事件模型任务id的 sql 语句: %s; modelID: %v", sqlStr, modelIDs))

	rows, err := ema.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	taskIDs := make([]string, 0)
	for rows.Next() {
		var taskID string
		err := rows.Scan(&taskID)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			return taskIDs, err
		}

		taskIDs = append(taskIDs, taskID)
	}
	return taskIDs, nil
}

// 更新任务
func (ema *eventModelAccess) UpdateEventTask(ctx context.Context, tx *sql.Tx, task interfaces.EventTask) error {

	// 2.0 序列化schedule
	scheduleBytes, err := sonic.Marshal(task.Schedule)
	if err != nil {
		logger.Errorf("Failed to marshal schedule, err: %v", err.Error())
		return err
	}
	storageConfigBytes, err := sonic.Marshal(task.StorageConfig)
	if err != nil {
		logger.Errorf("Failed to marshal storageConfig, err: %v", err.Error())
		return err
	}
	dispatchConfigBytes, err := sonic.Marshal(task.DispatchConfig)
	if err != nil {
		logger.Errorf("Failed to marshal dispatchConfig, err: %v", err.Error())
		return err
	}
	executeParameterBytes, err := sonic.Marshal(task.ExecuteParameter)
	if err != nil {
		logger.Errorf("Failed to marshal dispatchConfig, err: %v", err.Error())
		return err
	}
	data := map[string]interface{}{
		// "f_model_id":             task.ModelID,
		"f_schedule":                  scheduleBytes,
		"f_storage_config":            storageConfigBytes,
		"f_dispatch_config":           dispatchConfigBytes,
		"f_execute_parameter":         executeParameterBytes,
		"f_schedule_sync_status":      task.ScheduleSyncStatus,
		"f_downstream_dependent_task": strings.Join(task.DownstreamDependentTask, ","),
		"f_update_time":               task.UpdateTime,
	}
	sqlStr, vals, err := sq.Update(EVENT_TASK_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_task_id": task.TaskID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of update event model task by task_id, error: %s", err.Error())
		return err
	}
	// 记录处理的 sql 字符串
	logger.Debugf(fmt.Sprintf("修改事件模型任务的 sql 语句: %s", sqlStr))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("update event model task error: %v\n", err)
		return err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
	}

	if RowsAffected > 1 {
		// 影响行数不等于1不报错，更新操作已经发生
		logger.Errorf("UPDATE %d RowsAffected more than 1, RowsAffected is %d, eventModel is %v",
			task.TaskID, RowsAffected, task)
	}

	return nil
}

// 按任务id逻辑删除持久化任务，任务的状态设置为删除中。
// func (ema *eventModelAccess) SetTaskSyncStatusByTaskID(ctx context.Context, tx *sql.Tx, taskSyncStatus interfaces.EventTaskSyncStatus) error {

// 	if taskSyncStatus.TaskID == 0 {
// 		return nil
// 	}
// 	data := map[string]interface{}{
// 		"f_schedule_sync_status": taskSyncStatus.SyncStatus,
// 		"f_update_time":          taskSyncStatus.UpdateTime,
// 	}
// 	sqlStr, vals, err := sq.Update(EVENT_TASK_TABLE_NAME).
// 		SetMap(data).
// 		Where(sq.Eq{"f_task_id": taskSyncStatus.TaskID}).
// 		ToSql()
// 	if err != nil {
// 		logger.Errorf("Failed to build the sql of update event model task status by task_id, error: %s", err.Error())

// 		return err
// 	}
// 	// 记录处理的 sql 字符串
// 	logger.Debugf(fmt.Sprintf("逻辑删除事件模型的持久化任务的 sql 语句: %s", sqlStr))

// 	ret, err := tx.Exec(sqlStr, vals...)
// 	if err != nil {
// 		logger.Errorf("update event model task status error: %v\n", err)
// 		return err
// 	}

// 	//sql语句影响的行数
// 	RowsAffected, err := ret.RowsAffected()
// 	if err != nil {
// 		logger.Errorf("Get RowsAffected error: %v\n", err)
// 	}

// 	if RowsAffected != 1 {
// 		// 影响行数不等于请求的任务数时不报错，更新操作已经发生
// 		logger.Errorf("UPDATE %v RowsAffected not equals 1, RowsAffected is %d",
// 			taskSyncStatus.TaskID, RowsAffected)
// 	}
// 	return nil
// }

// 按模型id逻辑删除持久化任务，任务的状态设置为删除中。
// func (ema *eventModelAccess) SetTaskSyncStatusByModelID(ctx context.Context, tx *sql.Tx, taskSyncStatus interfaces.EventTaskSyncStatus) error {

// 	if taskSyncStatus.ModelID == 0 {
// 		return nil
// 	}
// 	data := map[string]interface{}{
// 		"f_schedule_sync_status": taskSyncStatus.SyncStatus,
// 		"f_update_time":          taskSyncStatus.UpdateTime,
// 	}
// 	sqlStr, vals, err := sq.Update(EVENT_TASK_TABLE_NAME).
// 		SetMap(data).
// 		Where(sq.Eq{"f_model_id": taskSyncStatus.ModelID}).
// 		ToSql()
// 	if err != nil {
// 		logger.Errorf("Failed to build the sql of update event model task status by f_model_id, error: %s", err.Error())
// 		return err
// 	}
// 	// 记录处理的 sql 字符串
// 	logger.Debugf(fmt.Sprintf("逻辑删除事件模型的持久化任务的 sql 语句: %s", sqlStr))

// 	ret, err := tx.Exec(sqlStr, vals...)
// 	if err != nil {
// 		logger.Errorf("update event model task status error: %v\n", err)
// 		return err
// 	}

// 	//sql语句影响的行数
// 	RowsAffected, err := ret.RowsAffected()
// 	if err != nil {
// 		logger.Errorf("Get RowsAffected error: %v\n", err)
// 		logger.Warn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
// 	}
// 	logger.Debugf(fmt.Sprintf("Update %d RowsAffected is %d", taskSyncStatus.ModelID, RowsAffected))

// 	return nil
// }

// 按任务id获取任务信息
func (ema *eventModelAccess) GetEventTaskByTaskID(ctx context.Context, taskID string) (interfaces.EventTask, error) {

	task := interfaces.EventTask{}

	sqlStr, vals, err := sq.Select(
		"f_task_id",
		"f_model_id",
		"f_schedule",
		"f_dispatch_config",
		"f_execute_parameter",
		"f_storage_config",
		"f_task_status",
		"f_status_update_time",
		"f_error_details",
		"f_schedule_sync_status",
		"f_downstream_dependent_task",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_update_time",
	).From(EVENT_TASK_TABLE_NAME).
		Where(sq.Eq{"f_task_id": taskID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select task, error: %s", err.Error())
		return task, err
	}

	// 记录处理的 sql 字符串
	logger.Debugf(fmt.Sprintf("查询事件模型任务的 sql 语句: %s; taskID: %v", sqlStr, taskID))
	row := ema.db.QueryRow(sqlStr, vals...)

	var scheduleBytes, storageConfigBytes, dispatchConfigBytes, executeParameterBytes []byte
	var DownstreamDependentTask string
	err = row.Scan(
		&task.TaskID,
		&task.ModelID,
		&scheduleBytes,
		&dispatchConfigBytes,
		&executeParameterBytes,
		&storageConfigBytes,
		&task.TaskStatus,
		&task.StatusUpdateTime,
		&task.ErrorDetails,
		&task.ScheduleSyncStatus,
		&DownstreamDependentTask,
		&task.Creator.ID,
		&task.Creator.Type,
		&task.CreateTime,
		&task.UpdateTime,
	)
	if err != nil {
		logger.Errorf("row scan failed, err: %v \n", err)
		return task, err
	}

	// 处理任务信息
	// 2.0 反序列化schedule/dispatch_config/execute_parameter
	err = sonic.Unmarshal(scheduleBytes, &task.Schedule)
	if err != nil {
		logger.Errorf("Failed to unmarshal schedule after getting event model, err: %v", err.Error())
		return interfaces.EventTask{}, err
	}
	err = sonic.Unmarshal(storageConfigBytes, &task.StorageConfig)
	if err != nil {
		logger.Errorf("Failed to unmarshal storageConfig after getting event model, err: %v", err.Error())
		return interfaces.EventTask{}, err
	}
	err = sonic.Unmarshal(dispatchConfigBytes, &task.DispatchConfig)
	if err != nil {
		logger.Errorf("Failed to unmarshal dispatchConfig after getting event model, err: %v", err.Error())
		return interfaces.EventTask{}, err
	}
	err = sonic.Unmarshal(executeParameterBytes, &task.ExecuteParameter)
	if err != nil {
		logger.Errorf("Failed to unmarshal executeParameter after getting event model, err: %v", err.Error())
		return interfaces.EventTask{}, err
	}
	task.DownstreamDependentTask = strings.Split(DownstreamDependentTask, ",")
	return task, nil
}

// 按模型id获取任务信息
func (ema *eventModelAccess) GetEventTaskByModelID(ctx context.Context, modelID string) (interfaces.EventTask, bool, error) {

	task := interfaces.EventTask{}
	sqlStr, vals, err := sq.Select(
		"f_task_id",
		"f_model_id",
		"f_schedule",
		"f_dispatch_config",
		"f_execute_parameter",
		"f_storage_config",
		"f_task_status",
		"f_status_update_time",
		"f_error_details",
		"f_schedule_sync_status",
		"COALESCE(f_downstream_dependent_task, '')",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_update_time").
		From(EVENT_TASK_TABLE_NAME).
		Where(sq.Eq{"f_model_id": modelID}).
		// Where(sq.NotEq{"f_schedule_sync_status": interfaces.SCHEDULE_SYNC_STATUS_DELETE}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select task, error: %s", err.Error())
		return task, false, err
	}

	// 记录处理的 sql 字符串
	logger.Debugf(fmt.Sprintf("查询事件模型任务的 sql 语句: %s; modelID: %v", sqlStr, modelID))

	var scheduleBytes, storageConfigBytes, dispatchConfigBytes, executeParameterBytes []byte
	var DownstreamDependentTask string

	rows := ema.db.QueryRow(sqlStr, vals...)
	err = rows.Scan(
		&task.TaskID,
		&task.ModelID,
		&scheduleBytes,
		&dispatchConfigBytes,
		&executeParameterBytes,
		&storageConfigBytes,
		&task.TaskStatus,
		&task.StatusUpdateTime,
		&task.ErrorDetails,
		&task.ScheduleSyncStatus,
		&DownstreamDependentTask,
		&task.Creator.ID,
		&task.Creator.Type,
		&task.CreateTime,
		&task.UpdateTime,
	)

	if err == sql.ErrNoRows {
		logger.Errorf("query no rows, error: %v \n", err)
		return task, false, nil
	} else if err != nil {
		logger.Errorf("row scan failed, err: %v \n", err)
		return interfaces.EventTask{}, false, err
	}

	// 处理任务信息
	// 2.0 反序列化schedule/dispatch_config/execute_parameter
	err = sonic.Unmarshal(scheduleBytes, &task.Schedule)
	if err != nil {
		logger.Errorf("Failed to unmarshal schedule after getting event model, err: %v", err.Error())
		return interfaces.EventTask{}, false, err
	}
	err = sonic.Unmarshal(storageConfigBytes, &task.StorageConfig)
	if err != nil {
		logger.Errorf("Failed to unmarshal storageConfig after getting event model, err: %v", err.Error())
		return interfaces.EventTask{}, false, err
	}
	err = sonic.Unmarshal(dispatchConfigBytes, &task.DispatchConfig)
	if err != nil {
		logger.Errorf("Failed to unmarshal dispatchConfig after getting event model, err: %v", err.Error())
		return interfaces.EventTask{}, false, err
	}
	err = sonic.Unmarshal(executeParameterBytes, &task.ExecuteParameter)
	if err != nil {
		logger.Errorf("Failed to unmarshal executeParameter after getting event model, err: %v", err.Error())
		return interfaces.EventTask{}, false, err
	}
	if DownstreamDependentTask != "" {
		task.DownstreamDependentTask = strings.Split(DownstreamDependentTask, ",")
	} else {
		task.DownstreamDependentTask = []string{}
	}
	return task, true, nil
}

// 获取正在进行中的任务
// func (ema *eventModelAccess) GetProcessingEventTasks(ctx context.Context) ([]interfaces.EventTask, error) {

// 	tasks := make([]interfaces.EventTask, 0)
// 	processingStatus := []int{
// 		interfaces.SCHEDULE_SYNC_STATUS_CREATE,
// 		interfaces.SCHEDULE_SYNC_STATUS_UPDATE,
// 		interfaces.SCHEDULE_SYNC_STATUS_DELETE,
// 	}
// 	sqlStr, vals, err := sq.Select(
// 		"f_task_id",
// 		"f_model_id",
// 		"f_schedule",
// 		"f_dispatch_config",
// 		"f_execute_parameter",
// 		"f_storage_config",
// 		"f_task_status",
// 		"f_status_update_time",
// 		"f_error_details",
// 		"f_schedule_sync_status",
// 		"f_downstream_dependent_task",
// 		"f_update_time").
// 		From(EVENT_TASK_TABLE_NAME).
// 		Where(sq.Eq{"f_schedule_sync_status": processingStatus}).
// 		OrderBy("f_update_time desc").
// 		ToSql()
// 	if err != nil {
// 		logger.Errorf("Failed to build the sql of select tasks, error: %s", err.Error())
// 		return tasks, err
// 	}

// 	// 记录处理的 sql 字符串
// 	logger.Debugf(fmt.Sprintf("查询正在进行中的事件模型的任务列表的 sql 语句: %s", sqlStr))
// 	rows, err := ema.db.Query(sqlStr, vals...)
// 	if err != nil {
// 		logger.Errorf("list data error: %v\n", err)
// 		return tasks, err
// 	}
// 	defer rows.Close()

// 	var DownstreamDependentTask string
// 	var (
// 		scheduleBytes, storageConfigBytes, dispatchConfigBytes, executeParameterBytes []byte
// 	)
// 	for rows.Next() {
// 		task := interfaces.EventTask{}
// 		err = rows.Scan(
// 			&task.TaskID,
// 			&task.ModelID,
// 			&scheduleBytes,
// 			&dispatchConfigBytes,
// 			&executeParameterBytes,
// 			&storageConfigBytes,
// 			&task.TaskStatus,
// 			&task.StatusUpdateTime,
// 			&task.ErrorDetails,
// 			&task.ScheduleSyncStatus,
// 			&DownstreamDependentTask,
// 			&task.UpdateTime,
// 		)
// 		if err != nil {
// 			logger.Errorf("row scan failed when GetProcessingEventTasks , err: %v \n", err)
// 			return tasks, err
// 		}

// 		// 处理任务信息
// 		// 2.0 反序列化schedule/dispatch_config/execute_parameter
// 		err = sonic.Unmarshal(scheduleBytes, &task.Schedule)
// 		if err != nil {
// 			logger.Errorf("Failed to unmarshal schedule after getting event model, err: %v", err.Error())
// 			return tasks, err
// 		}
// 		err = sonic.Unmarshal(storageConfigBytes, &task.StorageConfig)
// 		if err != nil {
// 			logger.Errorf("Failed to unmarshal storageConfig after getting event model, err: %v", err.Error())
// 			return tasks, err
// 		}
// 		err = sonic.Unmarshal(dispatchConfigBytes, &task.DispatchConfig)
// 		if err != nil {
// 			logger.Errorf("Failed to unmarshal dispatchConfig after getting event model, err: %v", err.Error())
// 			return tasks, err
// 		}
// 		err = sonic.Unmarshal(executeParameterBytes, &task.ExecuteParameter)
// 		if err != nil {
// 			logger.Errorf("Failed to unmarshal executeParameter after getting event model, err: %v", err.Error())
// 			return tasks, err
// 		}
// 		if DownstreamDependentTask != "" {
// 			task.DownstreamDependentTask = strings.Split(DownstreamDependentTask, ",")
// 		} else {
// 			task.DownstreamDependentTask = []string{}
// 		}

// 		tasks = append(tasks, task)
// 	}

// 	return tasks, nil
// }

// 更新任务状态为完成，更新调度id
func (ema *eventModelAccess) UpdateEventTaskStatusInFinish(ctx context.Context, task interfaces.EventTask) error {

	data := map[string]interface{}{
		"f_schedule_sync_status": interfaces.SCHEDULE_SYNC_STATUS_FINISH,
		"f_update_time":          task.UpdateTime,
	}
	sqlStr, vals, err := sq.Update(EVENT_TASK_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_task_id": task.TaskID}).
		Where(sq.LtOrEq{"f_update_time": task.UpdateTime}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of update event model task by task_id, error: %s", err.Error())
		return err
	}
	// 记录处理的 sql 字符串
	logger.Debugf(fmt.Sprintf("修改事件模型任务的 sql 语句: %s", sqlStr))

	ret, err := ema.db.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("update event model task error: %v\n", err)
		return err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
	}

	if RowsAffected > 1 {
		// 影响行数不等于1不报错，更新操作已经发生
		logger.Errorf("UPDATE %d RowsAffected more than 1, RowsAffected is %d, eventModel is %v",
			task.TaskID, RowsAffected, task)
	}

	return nil
}

// 物理删除任务
func (ema *eventModelAccess) DeleteEventTaskByTaskIDs(ctx context.Context, tx *sql.Tx, taskIDs []string) error {

	sqlStr, vals, err := sq.Delete(EVENT_TASK_TABLE_NAME).
		Where(sq.Eq{"f_task_id": taskIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of delete model task by f_task_id, error: %s", err.Error())
		return err
	}

	// 记录处理的 sql 字符串
	logger.Debugf(fmt.Sprintf("删除事件模型持久化任务的 sql 语句: %s; 删除的任务id: %v", sqlStr, taskIDs))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("delete data error: %v\n", err)
		return err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
	}
	logger.Debugf("RowsAffected: %d", RowsAffected)

	return nil
}
