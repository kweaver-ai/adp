// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric_model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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
	METRIC_MODEL_TABLE_NAME = "t_metric_model"
)

var (
	mmAccessOnce sync.Once
	mmAccess     interfaces.MetricModelAccess
)

type metricModelAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
}

func NewMetricModelAccess(appSetting *common.AppSetting) interfaces.MetricModelAccess {
	mmAccessOnce.Do(func() {
		mmAccess = &metricModelAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
		}
	})

	return mmAccess
}

// 根据ID获取指标模型存在性
func (mma *metricModelAccess) CheckMetricModelExistByID(ctx context.Context, modelID string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query metric model", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()
	//查询
	sqlStr, vals, err := sq.Select(
		"f_model_name").
		From(METRIC_MODEL_TABLE_NAME).
		Where(sq.Eq{"f_model_id": modelID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of get model name by id, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of get model name by id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")

		return "", false, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("获取指标模型信息的 sql 语句: %s", sqlStr))

	modelName := ""
	err = mma.db.QueryRow(sqlStr, vals...).Scan(
		&modelName,
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
	return modelName, true, nil
}

// 根据分组名称和模型名称获取指标模型存在性
// func (mma *metricModelAccess) CheckMetricModelExistByName(ctx context.Context, combinationName interfaces.CombinationName) (string, bool, error) {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "Query metric model", trace.WithSpanKind(trace.SpanKindClient))
// 	span.SetAttributes(
// 		attr.Key("db_url").String(libdb.GetDBUrl()),
// 		attr.Key("db_type").String(libdb.GetDBType()))
// 	defer span.End()
// 	//查询
// 	sqlStr, vals, err := sq.Select(
// 		"mm.f_model_id").
// 		From(METRIC_MODEL_GROUP_TABLE_NAME + " " + "AS mmg").
// 		Join(METRIC_MODEL_TABLE_NAME + " " + "AS mm on mmg.f_group_id = mm.f_group_id").
// 		Where(sq.Eq{"mm.f_model_name": combinationName.ModelName}).
// 		Where(sq.Eq{"mmg.f_group_name": combinationName.GroupName}).
// 		ToSql()
// 	if err != nil {
// 		logger.Errorf("Failed to build the sql of get model id by name, error: %s", err.Error())

// 		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of get model id by name, error: %s", err.Error()))
// 		span.SetStatus(codes.Error, "Build sql failed ")

// 		return "", false, err
// 	}

// 	// 记录处理的 sql 字符串
// 	o11y.Info(ctx, fmt.Sprintf("获取指标模型信息的 sql 语句: %s", sqlStr))

// 	modelInfo := interfaces.MetricModel{}
// 	err = mma.db.QueryRow(sqlStr, vals...).Scan(
// 		&modelInfo.ModelID,
// 	)
// 	if err == sql.ErrNoRows {
// 		span.SetAttributes(attr.Key("no_rows").Bool(true))
// 		span.SetStatus(codes.Ok, "")

// 		return "", false, nil
// 	} else if err != nil {
// 		logger.Errorf("row scan failed, err: %v\n", err)

// 		o11y.Error(ctx, fmt.Sprintf("Row scan failed, err: %v", err))
// 		span.SetStatus(codes.Error, "Row scan failed ")

// 		return "", false, err
// 	}

// 	span.SetStatus(codes.Ok, "")
// 	return modelInfo.ModelID, true, nil
// }

// 创建指标模型
func (mma *metricModelAccess) CreateMetricModel(ctx context.Context, tx *sql.Tx, metricModel interfaces.MetricModel) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Insert into metric model", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(metricModel.Tags)

	// 数据源序列化
	dsBytes, err := sonic.Marshal(metricModel.DataSource)
	if err != nil {
		logger.Errorf("Failed to marshal data source, err: %v", err.Error())
		return err
	}
	// sql配置序列化
	formulaConfigBytes, err := sonic.Marshal(metricModel.FormulaConfig)
	if err != nil {
		logger.Errorf("Failed to marshal formula config, err: %v", err.Error())
		return err
	}
	// 排序字段序列化
	orderByFieldsBytes, err := sonic.Marshal(metricModel.OrderByFields)
	if err != nil {
		logger.Errorf("Failed to marshal order by fields, err: %v", err.Error())
		return err
	}
	// having值过滤序列化
	havingConditionBytes, err := sonic.Marshal(metricModel.HavingCondition)
	if err != nil {
		logger.Errorf("Failed to marshal having condition, err: %v", err.Error())
		return err
	}

	// 分析维度序列化
	analysisDimsBytes, err := sonic.Marshal(metricModel.AnalysisDims)
	if err != nil {
		logger.Errorf("Failed to marshal analysis dimensions, err: %v", err.Error())
		return err
	}

	sqlStr, vals, err := sq.Insert(METRIC_MODEL_TABLE_NAME).
		Columns(
			"f_model_id",
			"f_model_name",
			"f_catalog_id",
			"f_catalog_content",
			"f_measure_name",
			"f_metric_type",
			"f_data_source",
			"f_query_type",
			"f_formula",
			"f_formula_config",
			"f_order_by_fields",
			"f_having_condition",
			"f_analysis_dimessions",
			"f_date_field",
			"f_measure_field",
			"f_unit_type",
			"f_unit",
			"f_tags",
			"f_comment",
			"f_group_id",
			"f_builtin",
			"f_calendar_interval",
			"f_creator",
			"f_creator_type",
			"f_create_time",
			"f_update_time",
		).
		Values(
			metricModel.ModelID,
			metricModel.ModelName,
			metricModel.CatalogID,
			metricModel.CatalogContent,
			metricModel.MeasureName,
			metricModel.MetricType,
			dsBytes,
			metricModel.QueryType,
			metricModel.Formula,
			formulaConfigBytes,
			orderByFieldsBytes,
			havingConditionBytes,
			analysisDimsBytes,
			metricModel.DateField,
			metricModel.MeasureField,
			metricModel.UnitType,
			metricModel.Unit,
			tagsStr,
			metricModel.Comment,
			metricModel.GroupID,
			metricModel.Builtin,
			metricModel.IsCalendarInterval,
			metricModel.Creator.ID,
			metricModel.Creator.Type,
			metricModel.CreateTime,
			metricModel.UpdateTime).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of insert model, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of insert model, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")

		return err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("创建指标模型的 sql 语句: %s", sqlStr))

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

// 按 id 获取指标模型信息
func (mma *metricModelAccess) GetMetricModelByModelID(ctx context.Context, modelID string) (interfaces.MetricModel, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Get metric model[%s]", modelID), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	metricModel := interfaces.MetricModel{
		ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
	}
	//查询
	sqlStr, vals, err := sq.Select(
		"mm.f_model_id",
		"mm.f_model_name",
		"mm.f_measure_name",
		"mm.f_metric_type",
		"mm.f_data_source",
		"mm.f_query_type",
		"mm.f_formula",
		"mm.f_formula_config",
		"mm.f_order_by_fields",
		"mm.f_having_condition",
		"mm.f_analysis_dimessions",
		"mm.f_date_field",
		"mm.f_measure_field",
		"mm.f_unit_type",
		"mm.f_unit",
		"mm.f_tags",
		"mm.f_comment",
		"mm.f_group_id",
		"mm.f_builtin",
		"mm.f_calendar_interval",
		"mm.f_create_time",
		"mm.f_update_time",
		"mmg.f_group_name").
		From(METRIC_MODEL_GROUP_TABLE_NAME + " " + "AS mmg").
		Join(METRIC_MODEL_TABLE_NAME + " " + "AS mm on mm.f_group_id = mmg.f_group_id").
		Where(sq.Eq{"mm.f_model_id": modelID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select model by id, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select model by id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")

		return metricModel, false, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("获取指标模型信息的 sql 语句: %s", sqlStr))

	tagsStr := ""
	var (
		dsBytes            []byte
		formulaConfigBytes []byte
		orderByBytes       []byte
		havingCondBytes    []byte
		analysisDimsBytes  []byte
	)
	err = mma.db.QueryRow(sqlStr, vals...).Scan(
		&metricModel.ModelID,
		&metricModel.ModelName,
		&metricModel.MeasureName,
		&metricModel.MetricType,
		&dsBytes,
		&metricModel.QueryType,
		&metricModel.Formula,
		&formulaConfigBytes,
		&orderByBytes,
		&havingCondBytes,
		&analysisDimsBytes,
		&metricModel.DateField,
		&metricModel.MeasureField,
		&metricModel.UnitType,
		&metricModel.Unit,
		&tagsStr,
		&metricModel.Comment,
		&metricModel.GroupID,
		&metricModel.Builtin,
		&metricModel.IsCalendarInterval,
		&metricModel.CreateTime,
		&metricModel.UpdateTime,
		&metricModel.GroupName,
	)

	if err == sql.ErrNoRows {
		logger.Errorf("query no rows, error: %v \n", err)
		span.SetStatus(codes.Error, fmt.Sprintf("Metric model %s not found", modelID))
		o11y.Error(ctx, fmt.Sprintf("Metric model %s not found, sql err: %v", modelID, err))

		return metricModel, false, nil
	} else if err != nil {
		logger.Errorf("row scan failed, error: %v \n", err)
		span.SetStatus(codes.Error, "Row scan failed")
		o11y.Error(ctx, fmt.Sprintf("Row scan failed, error: %v ", err))

		return interfaces.MetricModel{}, false, err
	}

	// tags string 转成数组的格式
	metricModel.Tags = libCommon.TagString2TagSlice(tagsStr)

	// 反序列化数据源信息 ds
	err = sonic.Unmarshal(dsBytes, &metricModel.DataSource)
	if err != nil {
		logger.Errorf("Failed to unmarshal schedule after getting metric model, err: %v", err.Error())
		return interfaces.MetricModel{}, false, err
	}
	// 反序列化 FormulaConfig
	if metricModel.QueryType == interfaces.SQL && formulaConfigBytes != nil {
		var sqlConfig interfaces.SQLConfig
		err = sonic.Unmarshal(formulaConfigBytes, &sqlConfig)
		if err != nil {
			logger.Errorf("Failed to unmarshal Formula Config after getting metric model model, err: %v", err.Error())
			return interfaces.MetricModel{}, false, err
		}
		metricModel.FormulaConfig = sqlConfig
	}
	// 衍生指标
	if metricModel.MetricType == interfaces.DERIVED_METRIC && formulaConfigBytes != nil {
		var derivedConfig interfaces.DerivedConfig
		err = sonic.Unmarshal(formulaConfigBytes, &derivedConfig)
		if err != nil {
			logger.Errorf("Failed to unmarshal Formula Config after getting metric model, err: %v", err.Error())
			return interfaces.MetricModel{}, false, err
		}
		metricModel.FormulaConfig = derivedConfig
	}
	// 反序列化排序字段
	if orderByBytes != nil {
		err = sonic.Unmarshal(orderByBytes, &metricModel.OrderByFields)
		if err != nil {
			logger.Errorf("Failed to unmarshal order by fields after getting metric model, err: %v", err.Error())
			return interfaces.MetricModel{}, false, err
		}
	}
	// 反序列化having过滤
	if havingCondBytes != nil {
		err = sonic.Unmarshal(havingCondBytes, &metricModel.HavingCondition)
		if err != nil {
			logger.Errorf("Failed to unmarshal having condition after getting metric model, err: %v", err.Error())
			return interfaces.MetricModel{}, false, err
		}
	}
	// 反序列化分析维度
	if analysisDimsBytes != nil {
		err = sonic.Unmarshal(analysisDimsBytes, &metricModel.AnalysisDims)
		if err != nil {
			logger.Errorf("Failed to unmarshal analysis dimensions after getting metric model, err: %v", err.Error())
			return interfaces.MetricModel{}, false, err
		}
	}

	span.SetStatus(codes.Ok, "")
	return metricModel, true, nil
}

// 批量获取指标模型信息，不包括任务信息。任务信息单独查询
func (mma *metricModelAccess) GetMetricModelsByModelIDs(ctx context.Context, modelIDs []string) ([]interfaces.MetricModel, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Get metric model[%s]", modelIDs), trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	metricModels := make([]interfaces.MetricModel, 0)
	//查询
	sqlStr, vals, err := sq.Select(
		"mm.f_model_id",
		"mm.f_model_name",
		"mm.f_catalog_id",
		"mm.f_catalog_content",
		"mm.f_measure_name",
		"mm.f_metric_type",
		"mm.f_data_source",
		"mm.f_query_type",
		"mm.f_formula",
		"mm.f_formula_config",
		"mm.f_order_by_fields",
		"mm.f_having_condition",
		"mm.f_analysis_dimessions",
		"mm.f_date_field",
		"mm.f_measure_field",
		"mm.f_unit_type",
		"mm.f_unit",
		"mm.f_tags",
		"mm.f_comment",
		"mm.f_group_id",
		"mm.f_builtin",
		"mm.f_calendar_interval",
		"mm.f_creator",
		"mm.f_creator_type",
		"mm.f_create_time",
		"mm.f_update_time",
		"mmg.f_group_name").
		From(METRIC_MODEL_GROUP_TABLE_NAME + " " + "AS mmg").
		Join(METRIC_MODEL_TABLE_NAME + " " + "AS mm on mm.f_group_id = mmg.f_group_id").
		Where(sq.Eq{"mm.f_model_id": modelIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select model by id, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select model by id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")

		return metricModels, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询指标模型列表的 sql 语句: %s.", sqlStr))
	rows, err := mma.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		span.SetStatus(codes.Error, "List data error")
		o11y.Error(ctx, fmt.Sprintf("List data error: %v", err))

		return metricModels, err
	}
	defer rows.Close()
	for rows.Next() {
		metricModel := interfaces.MetricModel{
			ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
		}
		tagsStr := ""
		var (
			dsBytes             []byte
			formulaConfigBytes  []byte
			orderByBytes        []byte
			havingCondBytes     []byte
			analysisDimsBytes   []byte
			catalogContentBytes []byte
		)
		err := rows.Scan(
			&metricModel.ModelID,
			&metricModel.ModelName,
			&metricModel.CatalogID,
			&catalogContentBytes,
			&metricModel.MeasureName,
			&metricModel.MetricType,
			&dsBytes,
			&metricModel.QueryType,
			&metricModel.Formula,
			&formulaConfigBytes,
			&orderByBytes,
			&havingCondBytes,
			&analysisDimsBytes,
			&metricModel.DateField,
			&metricModel.MeasureField,
			&metricModel.UnitType,
			&metricModel.Unit,
			&tagsStr,
			&metricModel.Comment,
			&metricModel.GroupID,
			&metricModel.Builtin,
			&metricModel.IsCalendarInterval,
			&metricModel.Creator.ID,
			&metricModel.Creator.Type,
			&metricModel.CreateTime,
			&metricModel.UpdateTime,
			&metricModel.GroupName,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			span.SetStatus(codes.Error, "Row scan error")
			o11y.Error(ctx, fmt.Sprintf("Row scan error: %v", err))
			return metricModels, err
		}

		// tags string 转成数组的格式
		metricModel.Tags = libCommon.TagString2TagSlice(tagsStr)

		if catalogContentBytes == nil {
			metricModel.CatalogContent = ""
		} else {
			metricModel.CatalogContent = string(catalogContentBytes)
		}

		// 反序列化数据源信息 ds
		err = sonic.Unmarshal(dsBytes, &metricModel.DataSource)
		if err != nil {
			logger.Errorf("Failed to unmarshal schedule after getting metric model, err: %v", err.Error())
			return metricModels, err
		}
		// 反序列化 FormulaConfig
		if metricModel.QueryType == interfaces.SQL && formulaConfigBytes != nil {
			var sqlConfig interfaces.SQLConfig
			err = sonic.Unmarshal(formulaConfigBytes, &sqlConfig)
			if err != nil {
				logger.Errorf("Failed to unmarshal Formula Config after getting metric model, err: %v", err.Error())
				return metricModels, err
			}
			metricModel.FormulaConfig = sqlConfig
		}
		// 衍生指标
		if metricModel.MetricType == interfaces.DERIVED_METRIC && formulaConfigBytes != nil {
			var derivedConfig interfaces.DerivedConfig
			err = sonic.Unmarshal(formulaConfigBytes, &derivedConfig)
			if err != nil {
				logger.Errorf("Failed to unmarshal Formula Config after getting metric model, err: %v", err.Error())
				return metricModels, err
			}
			metricModel.FormulaConfig = derivedConfig
		}
		// 反序列化排序字段
		if orderByBytes != nil {
			err = sonic.Unmarshal(orderByBytes, &metricModel.OrderByFields)
			if err != nil {
				logger.Errorf("Failed to unmarshal order by fields after getting metric model, err: %v", err.Error())
				return metricModels, err
			}
		}
		// 反序列化having过滤
		if havingCondBytes != nil {
			err = sonic.Unmarshal(havingCondBytes, &metricModel.HavingCondition)
			if err != nil {
				logger.Errorf("Failed to unmarshal having condition after getting metric model, err: %v", err.Error())
				return metricModels, err
			}
		}
		// 反序列化分析维度
		if analysisDimsBytes != nil {
			err = sonic.Unmarshal(analysisDimsBytes, &metricModel.AnalysisDims)
			if err != nil {
				logger.Errorf("Failed to unmarshal analysis dimensions after getting metric model, err: %v", err.Error())
				return metricModels, err
			}
		}

		metricModels = append(metricModels, metricModel)
	}

	span.SetStatus(codes.Ok, "")
	return metricModels, nil
}

// 修改指标模型
func (mma *metricModelAccess) UpdateMetricModel(ctx context.Context, tx *sql.Tx,
	metricModel interfaces.MetricModel) error {

	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Update metric model[%s]", metricModel.ModelID),
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))

	// tags 转成 string 的格式
	tagsStr := libCommon.TagSlice2TagString(metricModel.Tags)
	// 数据源序列化
	dsBytes, err := sonic.Marshal(metricModel.DataSource)
	if err != nil {
		logger.Errorf("Failed to marshal data source, err: %v", err.Error())
		return err
	}
	// sql配置序列化
	formulaConfigBytes, err := sonic.Marshal(metricModel.FormulaConfig)
	if err != nil {
		logger.Errorf("Failed to marshal formula config, err: %v", err.Error())
		return err
	}
	// 排序字段序列化
	orderByFieldsBytes, err := sonic.Marshal(metricModel.OrderByFields)
	if err != nil {
		logger.Errorf("Failed to marshal order by fields, err: %v", err.Error())
		return err
	}
	// having值过滤序列化
	havingConditionBytes, err := sonic.Marshal(metricModel.HavingCondition)
	if err != nil {
		logger.Errorf("Failed to marshal having condition, err: %v", err.Error())
		return err
	}
	// 分析维度序列化
	analysisDimsBytes, err := sonic.Marshal(metricModel.AnalysisDims)
	if err != nil {
		logger.Errorf("Failed to marshal analysis dimensions, err: %v", err.Error())
		return err
	}

	data := map[string]interface{}{
		"f_model_name":      metricModel.ModelName,
		"f_catalog_id":      metricModel.CatalogID,
		"f_catalog_content": metricModel.CatalogContent,
		// "f_measure_name":      metricModel.MeasureName, // measure_name不让修改
		"f_metric_type":         metricModel.MetricType,
		"f_data_source":         dsBytes,
		"f_query_type":          metricModel.QueryType,
		"f_formula":             metricModel.Formula,
		"f_formula_config":      formulaConfigBytes,
		"f_order_by_fields":     orderByFieldsBytes,
		"f_having_condition":    havingConditionBytes,
		"f_analysis_dimessions": analysisDimsBytes,
		"f_date_field":          metricModel.DateField,
		"f_measure_field":       metricModel.MeasureField,
		"f_unit_type":           metricModel.UnitType,
		"f_unit":                metricModel.Unit,
		"f_tags":                tagsStr,
		"f_comment":             metricModel.Comment,
		"f_group_id":            metricModel.GroupID,
		"f_builtin":             metricModel.Builtin,
		"f_calendar_interval":   metricModel.IsCalendarInterval,
		"f_update_time":         metricModel.UpdateTime,
	}
	sqlStr, vals, err := sq.Update(METRIC_MODEL_TABLE_NAME).
		SetMap(data).
		Where(sq.Eq{"f_model_id": metricModel.ModelID}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of update model by model_id, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of update model by model_id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("修改指标模型的 sql 语句: %s", sqlStr))

	ret, err := tx.Exec(sqlStr, vals...)
	if err != nil {
		logger.Errorf("update metric model error: %v\n", err)
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
		logger.Errorf("UPDATE %s RowsAffected more than 1, RowsAffected is %d, metricModel is %v",
			metricModel.ModelID, RowsAffected, metricModel)

		o11y.Warn(ctx, fmt.Sprintf("Update %s RowsAffected more than 1, RowsAffected is %d, metricModel is %v",
			metricModel.ModelID, RowsAffected, metricModel))
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 删除指标模型
func (mma *metricModelAccess) DeleteMetricModels(ctx context.Context, tx *sql.Tx, modelIDs []string) (int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete metric models from db",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("model_ids").String(fmt.Sprintf("%v", modelIDs)))

	if len(modelIDs) == 0 {
		return 0, nil
	}

	sqlStr, vals, err := sq.Delete(METRIC_MODEL_TABLE_NAME).
		Where(sq.Eq{"f_model_id": modelIDs}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of delete model by model_id, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of delete model by model_id, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("删除指标模型的 sql 语句: %s; 删除的模型ids: %v", sqlStr, modelIDs))

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

// 查询指标模型列表
// func (mma *metricModelAccess) ListMetricModels(ctx context.Context, modelsQuery interfaces.MetricModelsQueryParams) ([]interfaces.MetricModel, error) {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "Select metric models", trace.WithSpanKind(trace.SpanKindClient))
// 	span.SetAttributes(
// 		attr.Key("db_url").String(libdb.GetDBUrl()),
// 		attr.Key("db_type").String(libdb.GetDBType()))
// 	defer span.End()

// 	metricModels := make([]interfaces.MetricModel, 0)

// 	subBuilder := sq.Select(
// 		"mm.f_model_id",
// 		"mm.f_model_name",
// 		"mm.f_measure_name",
// 		"mm.f_metric_type",
// 		"mm.f_data_source",
// 		"mm.f_query_type",
// 		"mm.f_formula",
// 		"mm.f_date_field",
// 		"mm.f_measure_field",
// 		"mm.f_unit_type",
// 		"mm.f_unit",
// 		"mm.f_tags",
// 		"mm.f_comment",
// 		"mm.f_group_id",
// 		"mm.f_builtin",
// 		"mm.f_calendar_interval",
// 		"mm.f_create_time",
// 		"mm.f_update_time",
// 		"mmg.f_group_name").
// 		From(METRIC_MODEL_GROUP_TABLE_NAME + " " + "AS mmg").
// 		Join(METRIC_MODEL_TABLE_NAME + " " + "AS mm on mm.f_group_id = mmg.f_group_id")

// 	builder := mma.processQueryCondition(modelsQuery, subBuilder)
// 	//排序
// 	if modelsQuery.Sort == "f_group_name" {
// 		builder = builder.OrderBy(fmt.Sprint("mmg."+modelsQuery.Sort, " ", modelsQuery.Direction, ",mm.f_model_name ", modelsQuery.Direction))
// 	} else if modelsQuery.Sort == "model_name" {
// 		builder = builder.OrderBy(fmt.Sprint("mm."+modelsQuery.Sort, " ", modelsQuery.Direction, ",mmg.f_group_name ", modelsQuery.Direction))
// 	} else {
// 		builder = builder.OrderBy(fmt.Sprint("mm."+modelsQuery.Sort, " ", modelsQuery.Direction))
// 	}

// 	//添加分页参数 limit = -1 不分页，可选1-1000
// 	if modelsQuery.Limit != -1 {
// 		builder = builder.Offset(uint64(modelsQuery.Offset)).Limit(uint64(modelsQuery.Limit))
// 	}

// 	sqlStr, vals, err := builder.ToSql()
// 	if err != nil {
// 		logger.Errorf("Failed to build the sql of select models, error: %s", err.Error())

// 		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select models, error: %s", err.Error()))
// 		span.SetStatus(codes.Error, "Build sql failed ")
// 		return metricModels, err
// 	}
// 	// 记录处理的 sql 字符串
// 	o11y.Info(ctx, fmt.Sprintf("查询指标模型列表的 sql 语句: %s; queryParams: %v", sqlStr, modelsQuery))
// 	rows, err := mma.db.Query(sqlStr, vals...)
// 	if err != nil {
// 		logger.Errorf("list data error: %v\n", err)
// 		span.SetStatus(codes.Error, "List data error")
// 		o11y.Error(ctx, fmt.Sprintf("List data error: %v", err))

// 		return metricModels, err
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		metricModel := interfaces.MetricModel{}
// 		tagsStr := ""
// 		err := rows.Scan(
// 			&metricModel.ModelID,
// 			&metricModel.ModelName,
// 			&metricModel.MeasureName,
// 			&metricModel.MetricType,
// 			&metricModel.DataViewID,
// 			&metricModel.QueryType,
// 			&metricModel.Formula,
// 			&metricModel.DateField,
// 			&metricModel.MeasureField,
// 			&metricModel.UnitType,
// 			&metricModel.Unit,
// 			&tagsStr,
// 			&metricModel.Comment,
// 			&metricModel.GroupID,
// 			&metricModel.Builtin,
// 			&metricModel.IsCalendarInterval,
// 			&metricModel.CreateTime,
// 			&metricModel.UpdateTime,
// 			&metricModel.GroupName,
// 		)
// 		if err != nil {
// 			logger.Errorf("row scan failed, err: %v \n", err)
// 			span.SetStatus(codes.Error, "Row scan error")
// 			o11y.Error(ctx, fmt.Sprintf("Row scan error: %v", err))
// 			return metricModels, err
// 		}

// 		// tags string 转成数组的格式
// 		metricModel.Tags = libCommon.TagString2TagSlice(tagsStr)

// 		metricModels = append(metricModels, metricModel)
// 	}

// 	span.SetStatus(codes.Ok, "")
// 	return metricModels, nil
// }

// 查询指标模型部分字段列表
func (mma *metricModelAccess) ListSimpleMetricModels(ctx context.Context, modelsQuery interfaces.MetricModelsQueryParams) ([]interfaces.SimpleMetricModel, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Select simple metric models", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	simpleMetricModels := make([]interfaces.SimpleMetricModel, 0)

	subBuilder := sq.Select(
		"mm.f_model_id",
		"mm.f_model_name",
		"mm.f_measure_name",
		"mm.f_group_id",
		"mmg.f_group_name",
		"mm.f_tags",
		"mm.f_comment",
		"mm.f_metric_type",
		// "mm.f_data_source",
		"mm.f_query_type",
		"mm.f_formula",
		"mm.f_formula_config",
		"mm.f_order_by_fields",
		"mm.f_having_condition",
		"mm.f_analysis_dimessions",
		"mm.f_date_field",
		"mm.f_measure_field",
		"mm.f_unit_type",
		"mm.f_unit",
		"mm.f_builtin",
		"mm.f_calendar_interval",
		"mm.f_create_time",
		"mm.f_update_time").
		From(METRIC_MODEL_GROUP_TABLE_NAME + " " + "AS mmg").
		Join(METRIC_MODEL_TABLE_NAME + " " + "AS mm on mm.f_group_id = mmg.f_group_id")

	builder := mma.processQueryCondition(modelsQuery, subBuilder)
	//排序
	if modelsQuery.Sort == "f_group_name" {
		builder = builder.OrderBy(fmt.Sprint("mmg."+modelsQuery.Sort, " ", modelsQuery.Direction, ",mm.f_model_name ", modelsQuery.Direction))
	} else if modelsQuery.Sort == "model_name" {
		builder = builder.OrderBy(fmt.Sprint("mm."+modelsQuery.Sort, " ", modelsQuery.Direction, ",mmg.f_group_name ", modelsQuery.Direction))
	} else {
		builder = builder.OrderBy(fmt.Sprint("mm."+modelsQuery.Sort, " ", modelsQuery.Direction))
	}
	// 接入权限后不在数据库查询时分页，需从数据库中获取所有对象
	//添加分页参数 limit = -1 不分页，可选1-1000
	// if modelsQuery.Limit != -1 {
	// 	builder = builder.Offset(uint64(modelsQuery.Offset)).
	// 		Limit(uint64(modelsQuery.Limit))
	// }

	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select models, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select models, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return simpleMetricModels, err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询简单指标模型列表的 sql 语句: %s; queryParams: %v", sqlStr, modelsQuery))

	rows, err := mma.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		span.SetStatus(codes.Error, "List data error")
		o11y.Error(ctx, fmt.Sprintf("List data error: %v", err))

		return simpleMetricModels, err
	}
	defer rows.Close()

	for rows.Next() {
		simpleMetricModel := interfaces.SimpleMetricModel{}
		tagsStr := ""
		var (
			// dsBytes            []byte
			orderByBytes       []byte
			havingCondBytes    []byte
			formulaConfigBytes []byte
			analysisDimsBytes  []byte
		)
		err := rows.Scan(
			&simpleMetricModel.ModelID,
			&simpleMetricModel.ModelName,
			&simpleMetricModel.MeasureName,
			&simpleMetricModel.GroupID,
			&simpleMetricModel.GroupName,
			&tagsStr,
			&simpleMetricModel.Comment,
			&simpleMetricModel.MetricType,
			// &dsBytes,
			&simpleMetricModel.QueryType,
			&simpleMetricModel.Formula,
			&formulaConfigBytes,
			&orderByBytes,
			&havingCondBytes,
			&analysisDimsBytes,
			&simpleMetricModel.DateField,
			&simpleMetricModel.MeasureField,
			&simpleMetricModel.UnitType,
			&simpleMetricModel.Unit,
			&simpleMetricModel.Builtin,
			&simpleMetricModel.IsCalendarInterval,
			&simpleMetricModel.CreateTime,
			&simpleMetricModel.UpdateTime,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			span.SetStatus(codes.Error, "Row scan error")
			o11y.Error(ctx, fmt.Sprintf("Row scan error: %v", err))
			return simpleMetricModels, err
		}
		// tags string 转成数组的格式
		simpleMetricModel.Tags = libCommon.TagString2TagSlice(tagsStr)

		// 反序列化 FormulaConfig
		if simpleMetricModel.QueryType == interfaces.SQL && formulaConfigBytes != nil {
			var sqlConfig interfaces.SQLConfig
			err = sonic.Unmarshal(formulaConfigBytes, &sqlConfig)
			if err != nil {
				logger.Errorf("Failed to unmarshal Formula Config after getting metric model, err: %v", err.Error())
				return simpleMetricModels, err
			}
			simpleMetricModel.FormulaConfig = sqlConfig
		}
		// 衍生指标
		if simpleMetricModel.MetricType == interfaces.DERIVED_METRIC && formulaConfigBytes != nil {
			var derivedConfig interfaces.DerivedConfig
			err = sonic.Unmarshal(formulaConfigBytes, &derivedConfig)
			if err != nil {
				logger.Errorf("Failed to unmarshal Formula Config after getting metric model, err: %v", err.Error())
				return simpleMetricModels, err
			}
			simpleMetricModel.FormulaConfig = derivedConfig
		}
		// 反序列化排序字段
		if orderByBytes != nil {
			err = sonic.Unmarshal(orderByBytes, &simpleMetricModel.OrderByFields)
			if err != nil {
				logger.Errorf("Failed to unmarshal order by fields after getting metric model, err: %v", err.Error())
				return simpleMetricModels, err
			}
		}
		// 反序列化having过滤
		if havingCondBytes != nil {
			err = sonic.Unmarshal(havingCondBytes, &simpleMetricModel.HavingCondition)
			if err != nil {
				logger.Errorf("Failed to unmarshal having condition after getting metric model, err: %v", err.Error())
				return simpleMetricModels, err
			}
		}
		// 反序列化分析维度
		if analysisDimsBytes != nil {
			err = sonic.Unmarshal(analysisDimsBytes, &simpleMetricModel.AnalysisDims)
			if err != nil {
				logger.Errorf("Failed to unmarshal analysis dimensions after getting metric model, err: %v", err.Error())
				return simpleMetricModels, err
			}
		}

		simpleMetricModels = append(simpleMetricModels, simpleMetricModel)
	}

	span.SetStatus(codes.Ok, "")
	return simpleMetricModels, nil

}

// 查询指标模型总数
func (mma *metricModelAccess) GetMetricModelsTotal(ctx context.Context, modelsQuery interfaces.MetricModelsQueryParams) (int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Select metric models total number", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()

	// params, sqlStr := processQueryCondition(modelsQuery, sqlStr)
	subBuilder := sq.Select("COUNT(mm.f_model_id)").From(METRIC_MODEL_TABLE_NAME + " " + "AS mm")
	builder := mma.processQueryCondition(modelsQuery, subBuilder)
	sqlStr, vals, err := builder.ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of select models total, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select models total, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return 0, err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("查询指标模型总数的 sql 语句: %s; queryParams: %v", sqlStr, modelsQuery))

	total := 0
	err = mma.db.QueryRow(sqlStr, vals...).Scan(&total)
	if err != nil {
		logger.Errorf("get metric model totals error: %v\n", err)
		span.SetStatus(codes.Error, "Get metric model totals error")
		o11y.Error(ctx, fmt.Sprintf("Get metric model totals error: %v", err))
		return 0, err
	}

	span.SetStatus(codes.Ok, "")
	return total, nil
}

// 拼接 sql 过滤条件
func (mma *metricModelAccess) processQueryCondition(modelsQuery interfaces.MetricModelsQueryParams, subBuilder sq.SelectBuilder) sq.SelectBuilder {
	if modelsQuery.Name != "" {
		// 名称精确查询
		subBuilder = subBuilder.Where(sq.Eq{"mm.f_model_name": modelsQuery.Name})
	} else if modelsQuery.NamePattern != "" {
		// 模糊查询
		subBuilder = subBuilder.Where(sq.Expr("instr(mm.f_model_name, ?) > 0", modelsQuery.NamePattern))
	}

	// 指标类型过滤
	if modelsQuery.MetricType != "" {
		subBuilder = subBuilder.Where(sq.Eq{"mm.f_metric_type": modelsQuery.MetricType})
	}

	// 指标类型过滤
	if modelsQuery.Tag != "" {
		subBuilder = subBuilder.Where(sq.Expr("instr(mm.f_tags, ?) > 0", `"`+modelsQuery.Tag+`"`))
	}

	// 查询语言过滤，支持多选
	if modelsQuery.QueryType != "" {
		queryTypes := strings.Split(modelsQuery.QueryType, ",")
		subBuilder = subBuilder.Where(sq.Eq{"mm.f_query_type": queryTypes})
	}
	//根据分组查询
	if modelsQuery.GroupID != interfaces.GroupID_All {
		subBuilder = subBuilder.Where(sq.Eq{"mm.f_group_id": modelsQuery.GroupID})
	}

	return subBuilder
}

// 根据指标模型ID获取名称
func (mma *metricModelAccess) GetMetricModelSimpleInfosByIDs(ctx context.Context, modelIDs []string) (map[string]interfaces.SimpleMetricModel, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中根据指标模型ID数组获取其对应的名称", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("model_ids").String(fmt.Sprintf("%v", modelIDs)),
	)
	defer span.End()
	// 1. 初始化modelMap
	modelMap := make(map[string]interfaces.SimpleMetricModel)

	// 2. 生成完整的sql语句和参数列表
	sqlStr, args, err := sq.Select(
		"mm.f_model_id",
		"mm.f_model_name",
		"mmg.f_group_name",
		"mm.f_unit_type",
		"mm.f_unit",
		"mm.f_builtin").
		From(METRIC_MODEL_TABLE_NAME + " " + "AS mm").
		Join(METRIC_MODEL_GROUP_TABLE_NAME + " " + "AS mmg ON mm.f_group_id = mmg.f_group_id").
		Where(sq.Eq{"mm.f_model_id": modelIDs}).ToSql()

	if err != nil {
		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before get metric model names by ids, err: %v", err.Error())
		logger.Errorf(errDetails)
		// 记录错误log
		o11y.Error(ctx, errDetails)

		// 设置status
		span.SetStatus(codes.Error, "Failed to generate a sql statement using the squirrel sdk")
		return modelMap, err
	}

	// 3. debug日志级别下, 打印完整的sql语句
	sqlInfo := fmt.Sprintf("The detailed sql statement when getting metric model names by ids is: %v", sqlStr)
	logger.Debugf(sqlInfo)
	o11y.Info(ctx, sqlInfo)

	// 4. 执行查询操作
	rows, err := mma.db.Query(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get metric model names by ids, err: %v", err.Error())
		logger.Errorf(errDetails)
		// 记录错误log
		o11y.Error(ctx, errDetails)

		// 设置status
		span.SetStatus(codes.Error, "Failed to get metric model names by ids")
		return modelMap, err
	}

	defer rows.Close()

	// 5. Scan查询结果
	for rows.Next() {
		var (
			simpleInfo interfaces.SimpleMetricModel
		)

		err := rows.Scan(
			&simpleInfo.ModelID,
			&simpleInfo.ModelName,
			&simpleInfo.GroupName,
			&simpleInfo.UnitType,
			&simpleInfo.Unit,
			&simpleInfo.Builtin,
		)

		if err != nil {
			errDetails := fmt.Sprintf("Failed to scan row after executing the sql to get the metric model names by ids, err: %v", err.Error())
			logger.Errorf(errDetails)
			// 记录错误log
			o11y.Error(ctx, errDetails)

			// 设置status
			span.SetStatus(codes.Error, "Failed to scan row")
			return modelMap, err
		}

		modelMap[simpleInfo.ModelID] = simpleInfo
	}

	// 设置status
	span.SetStatus(codes.Ok, "")
	return modelMap, nil
}

// // GetMetricModelIDsBySimpleInfos 和 GetMetricModelSimpleInfosByIDs2 仅给结构模型使用，待结构模型下线，删掉
// // 根据指标模型分组名称和模型名称获取ID
// func (mma *metricModelAccess) GetMetricModelIDsBySimpleInfos(ctx context.Context, simpleInfos []interfaces.ModelSimpleInfo) (map[interfaces.ModelSimpleInfo]string, error) {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中根据指标模型名称数组获取其对应的ID", trace.WithSpanKind(trace.SpanKindClient))
// 	span.SetAttributes(
// 		attr.Key("db_url").String(libdb.GetDBUrl()),
// 		attr.Key("db_type").String(libdb.GetDBType()),
// 		attr.Key("simpleInfos").String(fmt.Sprintf("%v", simpleInfos)),
// 	)
// 	defer span.End()
// 	// 1. 初始化modelMap
// 	modelMap := make(map[interfaces.ModelSimpleInfo]string)

// 	// 2. 生成完整的sql语句和参数列表
// 	sqlBuilder := sq.Select(
// 		"mm.f_model_id",
// 		"mm.f_model_name",
// 		"mmg.f_group_name").
// 		From(METRIC_MODEL_TABLE_NAME + " " + "AS mm").
// 		Join(METRIC_MODEL_GROUP_TABLE_NAME + " " + "AS mmg ON mm.f_group_id = mmg.f_group_id")

// 	var conds sq.Or
// 	for _, info := range simpleInfos {
// 		conds = append(conds, sq.Eq{
// 			"mmg.f_group_name": info.GroupName,
// 			"mm.f_model_name":  info.ModelName,
// 		})
// 	}

// 	sqlStr, args, err := sqlBuilder.Where(conds).ToSql()
// 	if err != nil {
// 		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before get metric model ids by simple infos, err: %v", err.Error())
// 		logger.Errorf(errDetails)
// 		// 记录错误log
// 		o11y.Error(ctx, errDetails)

// 		// 设置status
// 		span.SetStatus(codes.Error, "Failed to generate a sql statement using the squirrel sdk")
// 		return modelMap, err
// 	}

// 	// 3. debug日志级别下, 打印完整的sql语句
// 	sqlInfo := fmt.Sprintf("The detailed sql statement when getting metric model ids by simple infos is: %v", sqlStr)
// 	logger.Debugf(sqlInfo)
// 	o11y.Info(ctx, sqlInfo)

// 	// 4. 执行查询操作
// 	rows, err := mma.db.Query(sqlStr, args...)
// 	if err != nil {
// 		errDetails := fmt.Sprintf("Failed to get metric model ids by simple infos, err: %v", err.Error())
// 		logger.Errorf(errDetails)
// 		// 记录错误log
// 		o11y.Error(ctx, errDetails)

// 		// 设置status
// 		span.SetStatus(codes.Error, "Failed to get metric model ids by simple infos")
// 		return modelMap, err
// 	}

// 	defer rows.Close()

// 	// 5. Scan查询结果
// 	for rows.Next() {
// 		var (
// 			modelID    string
// 			simpleInfo interfaces.ModelSimpleInfo
// 		)

// 		err := rows.Scan(
// 			&modelID,
// 			&simpleInfo.ModelName,
// 			&simpleInfo.GroupName,
// 		)

// 		if err != nil {
// 			errDetails := fmt.Sprintf("Failed to scan row after executing the sql to get the metric model ids by names, err: %v", err.Error())
// 			logger.Errorf(errDetails)
// 			// 记录错误log
// 			o11y.Error(ctx, errDetails)

// 			// 设置status
// 			span.SetStatus(codes.Error, "Failed to scan row")
// 			return modelMap, err
// 		}

// 		modelMap[simpleInfo] = modelID
// 	}

// 	// 设置status
// 	span.SetStatus(codes.Ok, "")
// 	return modelMap, nil
// }

// // 根据指标模型ID获取名称
// func (mma *metricModelAccess) GetMetricModelSimpleInfosByIDs2(ctx context.Context, modelIDs []string) (map[string]interfaces.ModelSimpleInfo, error) {
// 	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从数据库中根据指标模型ID数组获取其对应的名称", trace.WithSpanKind(trace.SpanKindClient))
// 	span.SetAttributes(
// 		attr.Key("db_url").String(libdb.GetDBUrl()),
// 		attr.Key("db_type").String(libdb.GetDBType()),
// 		attr.Key("model_ids").String(fmt.Sprintf("%v", modelIDs)),
// 	)
// 	defer span.End()
// 	// 1. 初始化modelMap
// 	modelMap := make(map[string]interfaces.ModelSimpleInfo)

// 	// 2. 生成完整的sql语句和参数列表
// 	sqlStr, args, err := sq.Select(
// 		"mm.f_model_id",
// 		"mm.f_model_name",
// 		"mmg.f_group_name",
// 		"mm.f_unit_type",
// 		"mm.f_unit").
// 		From(METRIC_MODEL_TABLE_NAME + " " + "AS mm").
// 		Join(METRIC_MODEL_GROUP_TABLE_NAME + " " + "AS mmg ON mm.f_group_id = mmg.f_group_id").
// 		Where(sq.Eq{"mm.f_model_id": modelIDs}).ToSql()

// 	if err != nil {
// 		errDetails := fmt.Sprintf("Failed to generate a sql statement using the squirrel sdk before get metric model names by ids, err: %v", err.Error())
// 		logger.Errorf(errDetails)
// 		// 记录错误log
// 		o11y.Error(ctx, errDetails)

// 		// 设置status
// 		span.SetStatus(codes.Error, "Failed to generate a sql statement using the squirrel sdk")
// 		return modelMap, err
// 	}

// 	// 3. debug日志级别下, 打印完整的sql语句
// 	sqlInfo := fmt.Sprintf("The detailed sql statement when getting metric model names by ids is: %v", sqlStr)
// 	logger.Debugf(sqlInfo)
// 	o11y.Info(ctx, sqlInfo)

// 	// 4. 执行查询操作
// 	rows, err := mma.db.Query(sqlStr, args...)
// 	if err != nil {
// 		errDetails := fmt.Sprintf("Failed to get metric model names by ids, err: %v", err.Error())
// 		logger.Errorf(errDetails)
// 		// 记录错误log
// 		o11y.Error(ctx, errDetails)

// 		// 设置status
// 		span.SetStatus(codes.Error, "Failed to get metric model names by ids")
// 		return modelMap, err
// 	}

// 	defer rows.Close()

// 	// 5. Scan查询结果
// 	for rows.Next() {
// 		var (
// 			modelID    string
// 			simpleInfo interfaces.ModelSimpleInfo
// 		)

// 		err := rows.Scan(
// 			&modelID,
// 			&simpleInfo.ModelName,
// 			&simpleInfo.GroupName,
// 			&simpleInfo.UnitType,
// 			&simpleInfo.Unit,
// 		)

// 		if err != nil {
// 			errDetails := fmt.Sprintf("Failed to scan row after executing the sql to get the metric model names by ids, err: %v", err.Error())
// 			logger.Errorf(errDetails)
// 			// 记录错误log
// 			o11y.Error(ctx, errDetails)

// 			// 设置status
// 			span.SetStatus(codes.Error, "Failed to scan row")
// 			return modelMap, err
// 		}

// 		modelMap[modelID] = simpleInfo
// 	}

// 	// 设置status
// 	span.SetStatus(codes.Ok, "")
// 	return modelMap, nil
// }

// 获取一个分组内所有指标模型（可以用来批量删除）
func (mma *metricModelAccess) GetMetricModelsByGroupID(ctx context.Context, groupID string) ([]interfaces.MetricModel, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Select metric models by group_id", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("group_id").String(fmt.Sprintf("%v", groupID)),
	)
	defer span.End()

	//不需要包含groupName
	metricModels := make([]interfaces.MetricModel, 0)

	sqlStr, vals, err := sq.Select(
		"mm.f_model_id",
		"mm.f_model_name",
		"mm.f_catalog_id",
		"mm.f_catalog_content",
		"mm.f_metric_type",
		"mm.f_data_source",
		"mm.f_query_type",
		"mm.f_formula",
		"mm.f_formula_config",
		"mm.f_order_by_fields",
		"mm.f_having_condition",
		"mm.f_analysis_dimessions",
		"mm.f_date_field",
		"mm.f_measure_field",
		"mm.f_unit_type",
		"mm.f_unit",
		"mm.f_tags",
		"mm.f_comment",
		"mm.f_group_id",
		"mm.f_update_time",
		"mmg.f_group_name").
		From(METRIC_MODEL_GROUP_TABLE_NAME + " " + "AS mmg").
		Join(METRIC_MODEL_TABLE_NAME + " " + "AS mm on mm.f_group_id = mmg.f_group_id").
		Where(sq.Eq{"mm.f_group_id": groupID}).ToSql()

	if err != nil {
		logger.Errorf("Failed to build the sql of select models by group_id, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select models by group_id error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return metricModels, err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("根据分组ID查询指标模型的 sql 语句: %s; queryParams: %s", sqlStr, groupID))

	rows, err := mma.db.Query(sqlStr, vals...)
	if err != nil {
		logger.Errorf("list data error: %v\n", err)
		span.SetStatus(codes.Error, "List data error")
		o11y.Error(ctx, fmt.Sprintf("List data error: %v", err))

		return metricModels, err
	}
	defer rows.Close()

	for rows.Next() {
		metricModel := interfaces.MetricModel{
			ModuleType: interfaces.MODULE_TYPE_METRIC_MODEL,
		}
		tagsStr := ""
		var (
			dsBytes             []byte
			formulaConfigBytes  []byte
			orderByBytes        []byte
			havingCondBytes     []byte
			analysisDimsBytes   []byte
			catalogContentBytes []byte
		)
		err := rows.Scan(
			&metricModel.ModelID,
			&metricModel.ModelName,
			&metricModel.CatalogID,
			&catalogContentBytes,
			&metricModel.MetricType,
			&dsBytes,
			&metricModel.QueryType,
			&metricModel.Formula,
			&formulaConfigBytes,
			&orderByBytes,
			&havingCondBytes,
			&analysisDimsBytes,
			&metricModel.DateField,
			&metricModel.MeasureField,
			&metricModel.UnitType,
			&metricModel.Unit,
			&tagsStr,
			&metricModel.Comment,
			&metricModel.GroupID,
			&metricModel.UpdateTime,
			&metricModel.GroupName,
		)
		if err != nil {
			logger.Errorf("row scan failed, err: %v \n", err)
			span.SetStatus(codes.Error, "Row scan error")
			o11y.Error(ctx, fmt.Sprintf("Row scan error: %v", err))
			return metricModels, err
		}

		// tags string 转成数组的格式
		metricModel.Tags = libCommon.TagString2TagSlice(tagsStr)

		if catalogContentBytes == nil {
			metricModel.CatalogContent = ""
		} else {
			metricModel.CatalogContent = string(catalogContentBytes)
		}

		// 反序列化数据源信息 ds
		err = sonic.Unmarshal(dsBytes, &metricModel.DataSource)
		if err != nil {
			logger.Errorf("Failed to unmarshal schedule after getting metric model, err: %v", err.Error())
			return metricModels, err
		}
		// 反序列化sql配置
		// 反序列化 FormulaConfig
		if metricModel.QueryType == interfaces.SQL && formulaConfigBytes != nil {
			var sqlConfig interfaces.SQLConfig
			err = sonic.Unmarshal(formulaConfigBytes, &sqlConfig)
			if err != nil {
				logger.Errorf("Failed to unmarshal Formula Config after getting metric model, err: %v", err.Error())
				return metricModels, err
			}
			metricModel.FormulaConfig = sqlConfig
		}
		// 衍生指标
		if metricModel.MetricType == interfaces.DERIVED_METRIC && formulaConfigBytes != nil {
			var derivedConfig interfaces.DerivedConfig
			err = sonic.Unmarshal(formulaConfigBytes, &derivedConfig)
			if err != nil {
				logger.Errorf("Failed to unmarshal Formula Config after getting metric model, err: %v", err.Error())
				return metricModels, err
			}
			metricModel.FormulaConfig = derivedConfig
		}

		// 反序列化排序字段
		if orderByBytes != nil {
			err = sonic.Unmarshal(orderByBytes, &metricModel.OrderByFields)
			if err != nil {
				logger.Errorf("Failed to unmarshal order by fields after getting metric model, err: %v", err.Error())
				return metricModels, err
			}
		}
		// 反序列化having过滤
		if havingCondBytes != nil {
			err = sonic.Unmarshal(havingCondBytes, &metricModel.HavingCondition)
			if err != nil {
				logger.Errorf("Failed to unmarshal having condition after getting metric model, err: %v", err.Error())
				return metricModels, err
			}
		}

		// 反序列化分析维度
		if analysisDimsBytes != nil {
			err = sonic.Unmarshal(analysisDimsBytes, &metricModel.AnalysisDims)
			if err != nil {
				logger.Errorf("Failed to unmarshal analysis dimensions after getting metric model, err: %v", err.Error())
				return metricModels, err
			}
		}

		metricModels = append(metricModels, metricModel)
	}
	span.SetStatus(codes.Ok, "")
	return metricModels, nil
}

// 修改指标模型的groupID （批量修改分组）
func (mma *metricModelAccess) UpdateMetricModelsGroupID(ctx context.Context, tx *sql.Tx, modelIDs []string, groupID string) (int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update metric models ", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("model_ids").String(fmt.Sprintf("%v", modelIDs)),
		attr.Key("group_id").String(groupID))
	defer span.End()

	if len(modelIDs) != 0 {
		data := map[string]interface{}{
			"f_group_id": groupID,
		}
		sqlStr, vals, err := sq.Update(METRIC_MODEL_TABLE_NAME).
			SetMap(data).
			Where(sq.Eq{"f_model_id": modelIDs}).
			ToSql()
		if err != nil {
			logger.Errorf("Failed to build the sql of update model attributes by model_id, error: %s", err.Error())

			o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of update model attributes by model_id, error: %s", err.Error()))
			span.SetStatus(codes.Error, "Build sql failed ")
			return 0, err
		}

		// 记录处理的 sql 字符串
		o11y.Info(ctx, fmt.Sprintf("修改指标模型的分组的 sql 语句: %s; 修改的模型ids: %v", sqlStr, modelIDs))

		ret, err := tx.Exec(sqlStr, vals...)
		if err != nil {
			logger.Errorf("update data error: %v\n", err)
			span.SetStatus(codes.Error, "Update data error")
			o11y.Error(ctx, fmt.Sprintf("Update data error: %v ", err))
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

	span.SetStatus(codes.Ok, "")
	return 0, nil
}

// 根据名称获取到ID ，可以用于根据名称获取到指标模型对象信息
func (mma *metricModelAccess) GetMetricModelIDByName(ctx context.Context, groupName, modelName string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Select metric model id  by group name and model name ", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("group_name").String(groupName),
		attr.Key("model_name").String(modelName))
	defer span.End()

	sqlStr, vals, err := sq.Select(
		"mm.f_model_id").
		From(METRIC_MODEL_GROUP_TABLE_NAME + " " + "AS mmg").
		Join(METRIC_MODEL_TABLE_NAME + " " + "AS mm on mmg.f_group_id = mm.f_group_id").
		Where(sq.Eq{"mm.f_model_name": modelName}).
		Where(sq.Eq{"mmg.f_group_name": groupName}).
		ToSql()

	if err != nil {
		logger.Errorf("Failed to build the sql of select metric model id by group_name and model_name, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of select metric model id by group_name and model_name, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")
		return "", false, err
	}
	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("根据名称获取指标模型ID sql 语句: %s; groupName: %s modelName: %s", sqlStr, groupName, modelName))

	var modelID string
	err = mma.db.QueryRow(sqlStr, vals...).Scan(&modelID)

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
	return modelID, true, nil

}

// 根据度量名称查询指标模型校验度量名称的唯一性
func (mma *metricModelAccess) CheckMetricModelByMeasureName(ctx context.Context, measureName string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query metric model by measure name", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()))
	defer span.End()
	//查询
	sqlStr, vals, err := sq.Select("f_model_id").
		From(METRIC_MODEL_TABLE_NAME).
		Where(sq.Eq{"f_measure_name": measureName}).
		ToSql()
	if err != nil {
		logger.Errorf("Failed to build the sql of get model id by f_measure_name, error: %s", err.Error())

		o11y.Error(ctx, fmt.Sprintf("Failed to build the sql of get model id by f_measure_name, error: %s", err.Error()))
		span.SetStatus(codes.Error, "Build sql failed ")

		return "", false, err
	}

	// 记录处理的 sql 字符串
	o11y.Info(ctx, fmt.Sprintf("获取指标模型信息的 sql 语句: %s", sqlStr))

	modelInfo := interfaces.MetricModel{}
	err = mma.db.QueryRow(sqlStr, vals...).Scan(
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
