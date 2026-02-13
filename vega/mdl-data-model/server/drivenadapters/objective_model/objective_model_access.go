// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package objective_model

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"data-model/common"
	"data-model/interfaces"
)

const (
	OBJECTIVE_MODEL_TABLE_NAME = "t_objective_model"
)

var (
	omAccessOnce sync.Once
	omAccess     interfaces.ObjectiveModelAccess
)

type objectiveModelAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewObjectiveModelAccess(appSetting *common.AppSetting) interfaces.ObjectiveModelAccess {
	omAccessOnce.Do(func() {
		omAccess = &objectiveModelAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})

	return omAccess
}

// 根据ID获取目标模型存在性
func (oma *objectiveModelAccess) CheckObjectiveModelExistByID(ctx context.Context, modelID string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query objective model", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()
	//查询
	sqlStr, vals, err := sq.Select(
		"f_model_name").
		From(OBJECTIVE_MODEL_TABLE_NAME).
		Where(sq.Eq{"f_model_id": modelID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of get model id by f_mode_id, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of get model id by f_mode_id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")

		return "", false, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("获取目标模型信息的 sql 语句: %s", sqlStr))

	var name string
	err = oma.db.QueryRow(sqlStr, vals...).
		Scan(&name)
	if err == sql.ErrNoRows {
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")

		return "", false, nil
	} else if err != nil {
		logger.Errorf("row scan failed, err: %v\n", err)

		o11y.Error(ctx, fmt.Sprintf("Row scan failed, err: %v", err))
		span.SetStatus(codes.Error, "Row scan failed ")

		return "", false, err
	}

	span.SetStatus(codes.Ok, "")
	return name, true, nil
}

// 根据名称获取目标模型存在性
func (oma *objectiveModelAccess) CheckObjectiveModelExistByName(ctx context.Context, modelName string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query objective model", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()
	//查询
	sqlStr, vals, err := sq.Select(
		"f_model_id").
		From(OBJECTIVE_MODEL_TABLE_NAME).
		Where(sq.Eq{"f_model_name": modelName}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of get model id by name, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of get model id by name, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")

		return "", false, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("获取目标模型信息的 sql 语句: %s", sqlStr))

	modelInfo := interfaces.ObjectiveModel{}
	err = oma.db.QueryRow(sqlStr, vals...).Scan(
		&modelInfo.ModelID,
	)
	if err == sql.ErrNoRows {
		span.SetAttributes(attr.Key("no_rows").Bool(true))
		span.SetStatus(codes.Ok, "")

		return "", false, nil
	} else if err != nil {
		logger.Errorf("row scan failed, err: %v\n", err)

		o11y.Error(ctx, fmt.Sprintf("Row scan failed, err: %v", err))
		span.SetStatus(codes.Error, "Row scan failed ")

		return "", false, err
	}

	span.SetStatus(codes.Ok, "")
	return modelInfo.ModelID, true, nil
}

// 创建目标模型
func (oma *objectiveModelAccess) CreateObjectiveModel(ctx context.Context, tx *sql.Tx, objectiveModel interfaces.ObjectiveModel) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Insert into objective model", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(objectiveModel.Tags)

	// 序列化目标配置
	configBytes, err := sonic.Marshal(objectiveModel.ObjectiveConfig)
	if err != nil {
		logger.Errorf("Failed to marshal objective config, err: %v", err.Error())
		return err
	}

	sqlStr, vals, err := sq.Insert(OBJECTIVE_MODEL_TABLE_NAME).
		Columns(
			"f_model_id",
			"f_model_name",
			"f_objective_type",
			"f_objective_config",
			"f_tags",
			"f_comment",
			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_update_time",
		).
		Values(
			objectiveModel.ModelID,
			objectiveModel.ModelName,
			objectiveModel.ObjectiveType,
			configBytes,
			tagsStr,
			objectiveModel.Comment,
			objectiveModel.Creator.ID,
			objectiveModel.Creator.Type,
			objectiveModel.CreateTime,
			objectiveModel.UpdateTime).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of insert model, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of insert model, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")

		return err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("创建目标模型的 sql 语句: %s", sqlStr))

	_, err = tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("insert data error: %v\n", err)

		span.SetStatus(codes.Error, "Insert data error")
		o11y.Error(ctx, fmt.Sprintf("Insert data error: %v ", err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 查询目标模型列表
func (oma *objectiveModelAccess) ListObjectiveModels(ctx context.Context, modelsQuery interfaces.ObjectiveModelsQueryParams) ([]interfaces.ObjectiveModel, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Select objective models", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	objectiveModels := make([]interfaces.ObjectiveModel, 0)

	subBuilder := sq.Select(
		"f_model_id",
		"f_model_name",
		"f_objective_type",
		"f_objective_config",
		"f_tags",
		"f_comment",
		"f_create_time",
		"f_update_time").
		From(OBJECTIVE_MODEL_TABLE_NAME)

	builder := processQueryCondition(modelsQuery, subBuilder)

	//排序
	builder = builder.OrderBy(fmt.Sprint(modelsQuery.Sort, " ", modelsQuery.Direction))

	// 接入权限后不在数据库查询时分页，需从数据库中获取所有对象
	//添加分页参数 limit = -1 不分页，可选1-1000
	// if modelsQuery.Limit != -1 {
	// 	builder = builder.Offset(uint64(modelsQuery.Offset)).Limit(uint64(modelsQuery.Limit))
	// }

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select models, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select models, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return objectiveModels, err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询目标模型列表的 sql 语句: %s; queryParams: %v", sqlStr, modelsQuery))
	rows, err := oma.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		span.SetStatus(codes.Error, "List data error")
		o11y.Error(ctx, fmt.Sprintf("List data error: %v", err))

		return objectiveModels, err
	}
	defer rows.Close()
	for rows.Next() {
		objectiveModel := interfaces.ObjectiveModel{}
		var oConfigBytes []byte
		tagsStr := ""
		err := rows.Scan(
			&objectiveModel.ModelID,
			&objectiveModel.ModelName,
			&objectiveModel.ObjectiveType,
			&oConfigBytes,
			&tagsStr,
			&objectiveModel.Comment,
			&objectiveModel.CreateTime,
			&objectiveModel.UpdateTime,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			span.SetStatus(codes.Error, "Row scan error")
			o11y.Error(ctx, fmt.Sprintf("Row scan error: %v", err))
			return objectiveModels, err
		}

		// tags string 转成数组的格式
		objectiveModel.Tags = libCommon.TagString2TagSlice(tagsStr)

		// 反序列化 ObjectiveConfig
		if objectiveModel.ObjectiveType == interfaces.SLO {
			var sloObjective interfaces.SLOObjective
			err = sonic.Unmarshal(oConfigBytes, &sloObjective)
			if err != nil {
				logger.Errorf("Failed to unmarshal SLOObjective after getting objective model, err: %v", err.Error())
				return objectiveModels, err
			}
			objectiveModel.ObjectiveConfig = sloObjective
		} else {
			var kpiObjective interfaces.KPIObjective
			err = sonic.Unmarshal(oConfigBytes, &kpiObjective)
			if err != nil {
				logger.Errorf("Failed to unmarshal KPIObjective after getting objective model, err: %v", err.Error())
				return objectiveModels, err
			}
			objectiveModel.ObjectiveConfig = kpiObjective
		}

		objectiveModels = append(objectiveModels, objectiveModel)
	}

	span.SetStatus(codes.Ok, "")
	return objectiveModels, nil
}

func (oma *objectiveModelAccess) GetObjectiveModelsTotal(ctx context.Context, modelsQuery interfaces.ObjectiveModelsQueryParams) (int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Select objective models total number", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	subBuilder := sq.Select("COUNT(f_model_id)").From(OBJECTIVE_MODEL_TABLE_NAME)
	builder := processQueryCondition(modelsQuery, subBuilder)
	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select objective models total, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select objective models total, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询目标模型总数的 sql 语句: %s; queryParams: %v", sqlStr, modelsQuery))

	total := 0
	err = oma.db.QueryRow(sqlStr, vals...).Scan(&total)
	if err != nil {
		logger.Errorf("get objective model totals error: %v\n", err)
		span.SetStatus(codes.Error, "Get objective model totals error")
		o11y.Error(ctx, fmt.Sprintf("Get objective model totals error: %v", err))
		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return total, nil
}

func (oma *objectiveModelAccess) GetObjectiveModelsByModelIDs(ctx context.Context, modelIDs []string) ([]interfaces.ObjectiveModel, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Get objective model[%s]", modelIDs), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	objectiveModels := make([]interfaces.ObjectiveModel, 0)
	//查询
	sqlStr, vals, err := sq.Select(
		"f_model_id",
		"f_model_name",
		"f_objective_type",
		"f_objective_config",
		"f_tags",
		"f_comment",
		"f_creator",
		"f_creator_type",
		"f_create_time",
		"f_update_time").
		From(OBJECTIVE_MODEL_TABLE_NAME).
		Where(sq.Eq{"f_model_id": modelIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select objective model by id, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select objective model by id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")

		return objectiveModels, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询目标模型列表的 sql 语句: %s.", sqlStr))
	rows, err := oma.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		span.SetStatus(codes.Error, "List data error")
		o11y.Error(ctx, fmt.Sprintf("List data error: %v", err))

		return objectiveModels, err
	}
	defer rows.Close()
	for rows.Next() {
		objectiveModel := interfaces.ObjectiveModel{}
		var oConfigBytes []byte
		tagsStr := ""
		err := rows.Scan(
			&objectiveModel.ModelID,
			&objectiveModel.ModelName,
			&objectiveModel.ObjectiveType,
			&oConfigBytes,
			&tagsStr,
			&objectiveModel.Comment,
			&objectiveModel.Creator.ID,
			&objectiveModel.Creator.Type,
			&objectiveModel.CreateTime,
			&objectiveModel.UpdateTime,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			span.SetStatus(codes.Error, "Row scan error")
			o11y.Error(ctx, fmt.Sprintf("Row scan error: %v", err))
			return objectiveModels, err
		}

		// tags string 转成数组的格式
		objectiveModel.Tags = libCommon.TagString2TagSlice(tagsStr)

		// 反序列化 ObjectiveConfig
		if objectiveModel.ObjectiveType == interfaces.SLO {
			var sloObjective interfaces.SLOObjective
			err = sonic.Unmarshal(oConfigBytes, &sloObjective)
			if err != nil {
				logger.Errorf("Failed to unmarshal SLOObjective after getting objective model, err: %v", err.Error())
				return objectiveModels, err
			}
			objectiveModel.ObjectiveConfig = sloObjective
		} else {
			var kpiObjective interfaces.KPIObjective
			err = sonic.Unmarshal(oConfigBytes, &kpiObjective)
			if err != nil {
				logger.Errorf("Failed to unmarshal KPIObjective after getting objective model, err: %v", err.Error())
				return objectiveModels, err
			}
			objectiveModel.ObjectiveConfig = kpiObjective
		}

		objectiveModels = append(objectiveModels, objectiveModel)
	}

	span.SetStatus(codes.Ok, "")
	return objectiveModels, nil
}

func (oma *objectiveModelAccess) UpdateObjectiveModel(ctx context.Context, tx *sql.Tx, objectiveModel interfaces.ObjectiveModel) error {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Update objective model[%s]", objectiveModel.ModelID), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(objectiveModel.Tags)

	// 序列化目标配置
	configBytes, err := sonic.Marshal(objectiveModel.ObjectiveConfig)
	if err != nil {
		logger.Errorf("Failed to marshal objective config, err: %v", err.Error())
		return err
	}

	data := map[string]interface{}{
		"f_model_name":       objectiveModel.ModelName,
		"f_objective_type":   objectiveModel.ObjectiveType,
		"f_objective_config": configBytes,
		"f_tags":             tagsStr,
		"f_comment":          objectiveModel.Comment,
		"f_update_time":      objectiveModel.UpdateTime,
	}
	sqlStr, vals, err := sq.Update(OBJECTIVE_MODEL_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_model_id": objectiveModel.ModelID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of update model by model_id, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of update model by model_id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("修改目标模型的 sql 语句: %s", sqlStr))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("update objective model error: %v\n", err)
		span.SetStatus(codes.Error, "Update data error")
		o11y.Error(ctx, fmt.Sprintf("Update data error: %v ", err))
		return err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		o11y.Warn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
	}

	if RowsAffected > 1 {
		// 影响行数不等于1不报错，更新操作已经发生
		logger.Errorf("UPDATE %d RowsAffected more than 1, RowsAffected is %d, objectiveModel is %v",
			objectiveModel.ModelID, RowsAffected, objectiveModel)

		o11y.Warn(ctx, fmt.Sprintf("Update %s RowsAffected more than 1, RowsAffected is %d, objectiveModel is %v",
			objectiveModel.ModelID, RowsAffected, objectiveModel))
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (oma *objectiveModelAccess) DeleteObjectiveModels(ctx context.Context, tx *sql.Tx, modelIDs []string) (int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete objective models from db", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("model_ids").String(fmt.Sprintf("%v", modelIDs)))
	defer span.End()

	if len(modelIDs) == 0 {
		return 0, nil
	}

	sqlStr, vals, err := sq.Delete(OBJECTIVE_MODEL_TABLE_NAME).
		Where(sq.Eq{"f_model_id": modelIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of delete objective model by model_id, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of delete objective model by model_id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("删除目标模型的 sql 语句: %s; 删除的模型ids: %v", sqlStr, modelIDs))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("delete data error: %v\n", err)
		span.SetStatus(codes.Error, "Delete data error")
		o11y.Error(ctx, fmt.Sprintf("Delete data error: %v ", err))

		return 0, err
	}

	//sql语句影响的行数
	RowsAffected, err := ret.RowsAffected()
	if err != nil {
		logger.Errorf("Get RowsAffected error: %v\n", err)
		span.SetStatus(codes.Error, "Get RowsAffected error")
		o11y.Warn(ctx, fmt.Sprintf("Get RowsAffected error: %v ", err))
	}
	logger.Infof("RowsAffected: %d", RowsAffected)
	span.SetStatus(codes.Ok, "")
	return RowsAffected, nil
}

// 拼接 sql 过滤条件
func processQueryCondition(modelsQuery interfaces.ObjectiveModelsQueryParams, subBuilder sq.SelectBuilder) sq.SelectBuilder {
	if modelsQuery.Name != "" {
		// 名称精确查询
		subBuilder = subBuilder.Where(sq.Eq{"f_model_name": modelsQuery.Name})
	} else if modelsQuery.NamePattern != "" {
		// 模糊查询
		subBuilder = subBuilder.Where(sq.Expr("instr(f_model_name, ?) > 0", modelsQuery.NamePattern))
	}

	// 目标类型过滤
	if modelsQuery.ObjectiveType != "" {
		subBuilder = subBuilder.Where(sq.Eq{"f_objective_type": modelsQuery.ObjectiveType})
	}

	// 指标类型过滤
	if modelsQuery.Tag != "" {
		subBuilder = subBuilder.Where(sq.Expr("instr(f_tags, ?) > 0", `"`+modelsQuery.Tag+`"`))
	}

	return subBuilder
}
