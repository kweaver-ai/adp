// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bytedance/sonic"
	"github.com/dlclark/regexp2"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	dcond "data-model/interfaces/condition"
)

// 指标模型分组信息校验
func validateMetricModelGroup(ctx context.Context, metricModelGroup *interfaces.MetricModelGroup) error {
	// 校验 指标模型分组名称的合法性, 非空、长度
	metricModelGroup.GroupName = strings.TrimSpace(metricModelGroup.GroupName)
	err := validateObjectName(ctx, metricModelGroup.GroupName, interfaces.METRIC_MODEL_GROUP_MODULE)
	if err != nil {
		return err
	}
	// 分组名称不能为内置视图分组名称, 内置分组在数据库中，后面校验重名时会校验到
	// if _, ok := interfaces.RESERVED_METRIC_MODEL_GROUP_NAME[metricModelGroup.GroupName]; ok {
	// 	return rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_MetricModelGroup_ForbiddenBuiltinGroupName).
	// 		WithErrorDetails("The group name cannot be empty string,  'event' or 'anyrobot_observability' as they are built-in groups")
	// }

	// 校验不能包含这些特殊字符，*"\/<>:|?#
	if strings.ContainsAny(metricModelGroup.GroupName, "*\"\\/<>:|?#") {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataViewGroup_InvalidParameter_GroupName).
			WithErrorDetails("The group name cannot contain special characters: *\"\\/<>:|?#")
	}

	//校验 指标模型分组备注的合法性，长度
	err = validateObjectComment(ctx, metricModelGroup.Comment)
	if err != nil {
		return err
	}
	return nil
}

// 指标模型必要创建参数的非空校验。bool 为 dsl 语句中是否使用了 top_hits 的标识。
func ValidateMetricModel(ctx context.Context, metricModel *interfaces.MetricModel) (bool, error) {
	// 校验名称合法性
	// 去掉模型名称的前后空格
	metricModel.ModelName = strings.TrimSpace(metricModel.ModelName)
	err := validateObjectName(ctx, metricModel.ModelName, interfaces.METRIC_MODEL_MODULE)
	if err != nil {
		return false, err
	}
	// 度量名称校验: 为兼容历史模型的导入，后端接口对空值不做校验，当为空时，在service层，把 __m.模型id 赋值给这个字段
	err = validateMeasureName(ctx, metricModel.MeasureName)
	if err != nil {
		return false, err
	}

	// 校验指标类型非空
	if metricModel.MetricType == "" {
		return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_MetricType)
	}
	// 指标类型是枚举中的某个值
	if !interfaces.IsValidMetricType(metricModel.MetricType) {
		return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_UnsupportMetricType)
	}

	// 校验原子指标
	switch metricModel.MetricType {
	case interfaces.ATOMIC_METRIC:
		// 数据源非空
		if metricModel.DataSource == nil {
			return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_DataSource)
		} else {
			// 校验数据源类型.数据源类型不填,由统一视图的 query_type来赋值,后端存储有这个字段
			// if !interfaces.IsValidDataSourceType(metricModel.DataSource.Type) {
			// 	return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_DataSourceType)
			// }
			// 数据源id不为空
			if metricModel.DataSource.ID == "" {
				return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_DataSourceID)
			}
		}

		// 查询语言非空
		if metricModel.QueryType == "" {
			return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_QueryType)
		}
		// 查询语言是枚举中的某个值
		if !interfaces.IsValidQueryType(metricModel.QueryType) {
			return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_UnsupportQueryType)
		}

		// sql的计算公式是在dsl_config中
		if metricModel.QueryType == interfaces.SQL {
			// 校验sql
			if metricModel.FormulaConfig == nil {
				return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_FormulaConfig)
			}
			// 把 formula_config转成 SQLConfig
			var sqlConfig interfaces.SQLConfig
			jsonData, err := sonic.Marshal(metricModel.FormulaConfig)
			if err != nil {
				return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_FormulaConfig).
					WithErrorDetails(fmt.Sprintf("[%s]'s SQL Config Marshal error: %s", metricModel.ModelName, err.Error()))
			}
			err = sonic.Unmarshal(jsonData, &sqlConfig)
			if err != nil {
				return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_FormulaConfig).
					WithErrorDetails(fmt.Sprintf("[%s]'s SQL Config Unmarshal error: %s", metricModel.ModelName, err.Error()))
			}

			// 度量计算不为空
			if sqlConfig.AggrExpr == nil && sqlConfig.AggrExprStr == "" {
				return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_SQLAggrExpression)
			}
			// aggr 的 code 和 config不能同时存在
			if sqlConfig.AggrExpr != nil {
				// code为空,配置不为空时,聚合函数和聚合字段都不能为空
				if sqlConfig.AggrExprStr == "" {
					if sqlConfig.AggrExpr.Aggr == "" || sqlConfig.AggrExpr.Field == "" {
						return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_SQLAggrExpression)
					}
				} else {
					// code不为空,且配置项也都不为空,报错(不能同时存在)
					if sqlConfig.AggrExpr.Aggr != "" && sqlConfig.AggrExpr.Field != "" {
						return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_SQLAggrExpression)
					}
					// 配置化对象不为空,配置项不完整,code不为空,则以code为准.配置话置空
					sqlConfig.AggrExpr = nil
				}
			}

			// code 和 config不能同时存在
			if sqlConfig.Condition != nil && sqlConfig.ConditionStr != "" {
				return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_SQLCondition)
			}

			// 校验过滤条件(与数据视图中的一样)：操作符、字段类型和操作符是否匹配
			err = validateCond(ctx, sqlConfig.Condition)
			if err != nil {
				return false, err
			}

			// FormulaConfig 赋值为 SqlConfig
			metricModel.FormulaConfig = sqlConfig

			// 时间字段非空
			// if metricModel.DateField == "" {
			// 	return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_DateField)
			// }

			metricModel.MeasureField = interfaces.SQL_METRICFIELD

			// sql当前只支持日历间隔
			metricModel.IsCalendarInterval = 1
		} else {
			// 非sql的数据源类型是数据视图data_view
			// if metricModel.DataSource.Type != interfaces.DATA_SOURCE_DATA_VIEW {
			// 	return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_DataSourceType)
			// }
			// 计算公式非空
			if metricModel.Formula == "" {
				return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_Formula)
			}
		}

		// 查询语言是 promql，则时间维度必须是 @timestamp，时间格式必须是 epoch_millis
		if metricModel.QueryType == interfaces.PROMQL {
			// 时间字段
			if metricModel.DateField == "" {
				metricModel.DateField = interfaces.PROMQL_DATEFIELD
			} else if metricModel.DateField != interfaces.PROMQL_DATEFIELD {
				return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_DateField)
			}

			// 度量字段
			if metricModel.MeasureField == "" {
				metricModel.MeasureField = interfaces.PROMQL_METRICFIELD
			} else if metricModel.MeasureField != interfaces.PROMQL_METRICFIELD {
				return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_MeasureField)
			}
		}

		// 查询语言是 dsl(时序)，则时间维度必须是 date_histogram 的聚合名称，时间格式必须是 epoch_millis
		var containTopHits bool
		if metricModel.QueryType == interfaces.DSL {
			err := validateDSL(ctx, metricModel, &containTopHits)
			// containTopHits = containTopHitsTmp
			if err != nil {
				return containTopHits, err
			}
		}

		// 除了 promql 的情况外，对度量字段做非空校验
		if metricModel.MeasureField == "" {
			return containTopHits, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_MeasureField)
		}

	case interfaces.DERIVED_METRIC:
		// 校验衍生指标配置
		if metricModel.FormulaConfig == nil {
			return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_FormulaConfig)
		}
		// 把 formula_config转成 DerivedConfig
		var derivedConfig interfaces.DerivedConfig
		jsonData, err := sonic.Marshal(metricModel.FormulaConfig)
		if err != nil {
			return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_FormulaConfig).
				WithErrorDetails(fmt.Sprintf("[%s]'s Derived Config Marshal error: %s", metricModel.ModelName, err.Error()))
		}
		err = sonic.Unmarshal(jsonData, &derivedConfig)
		if err != nil {
			return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_FormulaConfig).
				WithErrorDetails(fmt.Sprintf("[%s]'s Derived Config Unmarshal error: %s", metricModel.ModelName, err.Error()))
		}

		// 依赖模型不能为空
		if derivedConfig.DependMetricModel == nil || derivedConfig.DependMetricModel.ID == "" {
			return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_DependMetricModel)
		}

		if derivedConfig.ConditionStr == "" {
			if derivedConfig.DateCondition == nil && derivedConfig.BusinessCondition == nil {
				// 都为空，error
				return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_DerivedCondition)
			}
			// 校验过滤条件(与数据视图中的一样)：操作符、字段类型和操作符是否匹配
			if derivedConfig.DateCondition != nil {
				err = validateCond(ctx, derivedConfig.DateCondition)
				if err != nil {
					return false, err
				}
			}

			if derivedConfig.BusinessCondition != nil {
				err = validateCond(ctx, derivedConfig.BusinessCondition)
				if err != nil {
					return false, err
				}
			}
		}

		// code 和 config不能同时存在
		if (derivedConfig.DateCondition != nil || derivedConfig.BusinessCondition != nil) && derivedConfig.ConditionStr != "" {
			return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_DerivedCondition)
		}

		// 存储时只存依赖模型的ID，去除名称
		derivedConfig.DependMetricModel = &interfaces.DependMetricModel{ID: derivedConfig.DependMetricModel.ID}
		// FormulaConfig 赋值为 DerivedConfig
		metricModel.FormulaConfig = derivedConfig

	case interfaces.COMPOSITED_METRIC:
		// 复合指标，计算公式有效性检查交在service层检查

	}

	// 单位类型非空
	if metricModel.UnitType == "" {
		return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_UnitType)
	}

	// 单位类型有效
	if !interfaces.IsValidUnitType(metricModel.UnitType) {
		return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_UnitType)
	}

	// 如果导入的单位是%，把其单位类型改成 PERCENTAGE_UNIT
	if metricModel.Unit == "%" {
		metricModel.UnitType = interfaces.PERCENTAGE_UNIT
	}

	// 单位非空
	if metricModel.Unit == "" {
		return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_Unit)
	}

	// 单位有效。
	ok, newUnit := interfaces.IsValidUnit(metricModel.UnitType, metricModel.Unit)
	if !ok {
		return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_UnitType)
	}
	metricModel.Unit = newUnit

	// 若输入了 tags，校验 tags 的合法性
	err = ValidateTags(ctx, metricModel.Tags)
	if err != nil {
		return false, err
	}

	// 去掉tag前后空格以及数组去重
	metricModel.Tags = libCommon.TagSliceTransform(metricModel.Tags)

	// 校验comment合法性
	err = validateObjectComment(ctx, metricModel.Comment)
	if err != nil {
		return false, err
	}

	// SQL 不支持配置任务，若接口设置了任务，则置空处理
	if metricModel.QueryType == interfaces.SQL {
		metricModel.Task = nil
	}
	// 校验任务信息
	// 当任务不为空时，才校验。兼容老的指标模型的导入
	if metricModel.Task != nil {
		err = validateTask(ctx, metricModel.QueryType, *metricModel.Task)
		if err != nil {
			return false, err
		}
	}

	// 校验排序字段的非空
	for _, orderBy := range metricModel.OrderByFields {
		if orderBy.Name == "" {
			return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_OrderByName)
		}
		if orderBy.Direction == "" {
			return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_OrderByDirection)
		}
		if orderBy.Direction != interfaces.DESC_DIRECTION && orderBy.Direction != interfaces.ASC_DIRECTION {
			return false, rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_OrderByDirection).
				WithErrorDetails("The order direction is not desc or asc")
		}
	}

	// 校验 havingCondition 的有效性
	if metricModel.HavingCondition != nil {
		err = validateHavingCondition(ctx, metricModel.HavingCondition)
		if err != nil {
			return false, err
		}
	}

	return false, nil
}

func validateMeasureName(ctx context.Context, measureName string) error {
	// 1. 度量名称校验: 为兼容历史模型的导入，后端接口对空值不做校验，当为空时，在service层，把 __m.模型id 赋值给这个字段
	if measureName == "" {
		return nil
	}

	// 长度校验
	if utf8.RuneCountInString(measureName) > interfaces.OBJECT_NAME_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_LengthExceeded_MeasureName).
			WithErrorDetails(fmt.Sprintf("The length of the measure name exceeds %v", interfaces.OBJECT_NAME_MAX_LENGTH))
	}

	// 2. 须以 __m. 为前缀, 只允许使用字母、数字、下划线,且自定义部分只能以字母或数字开头
	reg1, err := regexp.Compile(interfaces.MEASURE_NAME_RULE)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_MetricModel_InternalError_MeasureNameRuleCompileFailed).
			WithErrorDetails(fmt.Sprintf("regexp compile rule[%s] error:%v", interfaces.MEASURE_NAME_RULE, err.Error()))
	}
	if regexMatch := reg1.MatchString(measureName); !regexMatch {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_MeasureName).
			WithErrorDetails("The measure name can only support English letters and numbers and _, and can only start with a letter or an number")
	}

	return nil
}

// 指标模型持久化任务信息校验
func validateTask(ctx context.Context, queryType string, task interfaces.MetricTask) error {
	// 任务间的step和过滤条件不能相同
	stepsMap := make(map[string]interface{})

	// 1. 任务名称长度限制
	if utf8.RuneCountInString(task.TaskName) > interfaces.OBJECT_NAME_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_LengthExceeded_TaskName).
			WithErrorDetails(fmt.Sprintf("The length of the task_name named %v exceeds %v", task.TaskName, interfaces.OBJECT_NAME_MAX_LENGTH))
	}

	// 2. 执行频率
	if task.Schedule == (interfaces.Schedule{}) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_Schedule)
	}
	// 2.1 执行类型
	if task.Schedule.Type == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_ScheduleType)
	}
	if task.Schedule.Type != interfaces.SCHEDULE_TYPE_FIXED && task.Schedule.Type != interfaces.SCHEDULE_TYPE_CRON {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_ScheduleType)
	}
	// 2.2 执行表达式
	if task.Schedule.Expression == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_ScheduleExpression)
	}
	if task.Schedule.Type == interfaces.SCHEDULE_TYPE_FIXED {
		// 不为空时校验有效性,time duration的格式，单位支持 m - 分钟； h - 小时； d - 天
		err := validateDuration(ctx, task.Schedule.Expression, common.DurationDayHourMinuteRE,
			derrors.DataModel_MetricModel_InvalidParameter_ScheduleExpression, "schedule expression ", true)
		if err != nil {
			return err
		}
	} else {
		// cron 表达式的校验。只支持6位，不支持年的指定。
		_, err := CronParser.Parse(task.Schedule.Expression)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_ScheduleExpression).
				WithErrorDetails(err.Error())
		}
	}

	// 4.0 时间窗口在 promql 时不能配置，只在 dsl 时配置
	if queryType == interfaces.PROMQL && len(task.TimeWindows) != 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_TimeWindows).
			WithErrorDetails("when the query type is PromQL, time windows should be null")
	}

	// 4. 时间窗口。支持 m - 分钟； h - 小时； d - 天；还支持使用 前1小时(previous_hour)、
	// 前1天(previous_day)、前一周(previous_week)、前一个月(previous_month)
	// 一个任务中的时间窗口字符串不能重复
	if queryType == interfaces.DSL {
		windowMap := make(map[string]string)
		if len(task.TimeWindows) == 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_TimeWindows).
				WithErrorDetails("when the query type is PromQL, time windows should not be null")
		} else {
			for _, timeWindow := range task.TimeWindows {
				if timeWindow == "" {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_TimeWindows)
				} else {
					// 不为空时校验有效性,time duration的格式，单位支持 m - 分钟； h - 小时前1小时(previous_hour)、
					// 前1天(previous_day)、前一周(previous_week)、前一个月(previous_month)
					if _, exist := windowMap[timeWindow]; !exist {
						windowMap[timeWindow] = timeWindow
					} else {
						// 存在重复的时间窗口字符串
						return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_Duplicated_TimeWindows)
					}
					// _, exists := interfaces.PREVIOUS_TIMEWINDOW[timeWindow]
					// // 如果时间窗口不是前一个时间单位的窗口，就校验时间区间的有效性；如果是，无需校验，遍历下一个。
					// if !exists {
					err := validateDuration(ctx, timeWindow, common.DurationDayHourMinuteRE,
						derrors.DataModel_MetricModel_InvalidParameter_TimeWindows, "time window", true)
					if err != nil {
						return err
					}
					// }
				}
			}
		}
	}

	// 7. 追溯时长。不为空时，校验时间区间的有效性， h - 小时； d - 天；
	var retraceDurationV time.Duration
	if task.RetraceDuration != "" {
		//追溯时长不为空时，需要校验有效性
		err := validateDuration(ctx, task.RetraceDuration, common.DurationDayHourRE,
			derrors.DataModel_MetricModel_InvalidParameter_RetraceDuration, "retrace duration", false)
		if err != nil {
			return err
		}

		// 解析追溯时长
		retraceDurationV, err = common.ParseDuration(task.RetraceDuration, common.DurationDayHourRE, false)
		if err != nil {
			return err
		}
	}

	// 5. 持久化步长
	if len(task.Steps) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_Step)
	} else {
		// 是否有效，是否重复。
		for _, step := range task.Steps {
			_, exists := common.PersistStepsMap[step]
			if !exists {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Step).
					WithErrorDetails(fmt.Sprintf("expect persist steps is one of %v, actaul is %s",
						common.PersistSteps, step))
			}
			// 不允许有重复的步长
			if _, ok := stepsMap[step]; !ok {
				stepsMap[step] = nil
			} else {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_Duplicated_TaskStep)
			}

			// 解析持久化步长
			stepV, err := common.ParseDuration(step, common.DurationDayHourMinuteRE, true)
			if err != nil {
				return err
			}

			// 能追溯的数据时间点数小于1w
			if retraceDurationV > 0 && (float64(retraceDurationV)/float64(stepV)) > interfaces.MAX_RETRACE_POINTS_NUM {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_RetraceDuration).
					WithErrorDetails(fmt.Sprintf("The task of retrace data point can not exceed 10000, actual is [%s]", retraceDurationV/stepV))
			}
		}

	}
	// 6. 索引库类型
	if task.IndexBase == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_IndexBase)
	}

	// 8. 备注
	if utf8.RuneCountInString(task.Comment) > interfaces.COMMENT_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_LengthExceeded_TaskComment).
			WithErrorDetails(fmt.Sprintf("The length of the task comment exceeds %v", interfaces.COMMENT_MAX_LENGTH))
	}

	return nil
}

// 校验时间区间
func validateDuration(ctx context.Context, durationStr string, reg *regexp.Regexp, errCode string, objectName string, containMinute bool) error {
	durationV, err := common.ParseDuration(durationStr, reg, containMinute)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, errCode).
			WithErrorDetails(err.Error())
		return httpErr
	}

	if durationV <= 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, errCode).
			WithErrorDetails(fmt.Sprintf("zero or negative %s is not accepted. Try a positive integer", objectName))
		return httpErr
	}
	return nil
}

// 指标模型中dsl语句的聚合表达式的校验
func validateDSL(ctx context.Context, metricModel *interfaces.MetricModel, containTopHits *bool) error {
	// 解析 dsl 的 aggregations 部分，做 dsl 相关的校验

	// 1. 把 dsl 语句转成 json，结构为 map[string]interfaces{}
	var dsl map[string]interface{}
	err := sonic.Unmarshal([]byte(metricModel.Formula), &dsl)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
			WithErrorDetails(fmt.Sprintf("DSL Unmarshal error: %s", err.Error()))
	}

	//  2. dsl 的 size 置0
	if dsl["size"] == nil || dsl["size"] != float64(0) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
			WithErrorDetails(fmt.Sprintf("The size of dsl expected 0, actual is %v", dsl["size"]))
	}

	// 3. 递归读取 json
	aggs, exist := dsl[interfaces.AGGS]
	if !exist {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("dsl missing aggregation.")
	}
	// 4. 校验aggs
	aggsLayers := 0
	aggNameMap := make(map[string]int, 0)
	dateHistogramAggName := make([]string, 0)
	metricAggName := make([]string, 0)
	err = validAggs(ctx, aggs, metricModel, &aggsLayers, aggNameMap, &dateHistogramAggName, &metricAggName, containTopHits)
	if err != nil {
		return err
	}

	// 5. 最大聚合层数7，6层分桶+1层值
	if aggsLayers > 7 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("The number of aggregation layers in DSL does not exceed 7.")
	}

	// 6. 度量字段
	if len(metricAggName) != 1 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("A metric aggregation should be included in dsl")
	}
	if metricModel.MeasureField == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_MeasureField)
	} else if metricModel.MeasureField != metricAggName[0] {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_MeasureField).
			WithErrorDetails("Measure Field should be the name of metric aggregation in dsl")
	}

	// 7. 时间字段
	if len(dateHistogramAggName) != 1 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("Only one date_histogram aggregation should be included in dsl")
	}
	if metricModel.DateField == "" {
		metricModel.DateField = dateHistogramAggName[0]
	} else if metricModel.DateField != dateHistogramAggName[0] {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_DateField)
	}

	// 9. date_histogram 下是值聚合
	if aggNameMap[dateHistogramAggName[0]]+1 != aggNameMap[metricAggName[0]] {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("The sub aggregation of date_histogram aggregation needs to be a metric aggregation.")
	}

	return nil
}

// validAggs
// @Description: 逐层校验子聚合
// @param ctx
// @param aggs: dsl 中 "aggs" 对应的值内容
// @param measureFiled: 指标模型中指定的度量字段
// @param aggsLayers: 表示聚合所在的层数
// @param aggNameMap: 存储的是<各层的聚合名称, 层数>
// @param dateHistogramAggName: date_histogram 的聚合名称，用于校验时间字段
// @param metricAggName: 值聚合的聚合名称，用于校验度量字段
// @param containTopHits: dsl 语句中是否包含 tophits 聚合
// @return error
func validAggs(ctx context.Context, aggs interface{}, metricModel *interfaces.MetricModel, aggsLayers *int, aggNameMap map[string]int,
	dateHistogramAggName *[]string, metricAggName *[]string, containTopHits *bool) error {

	aggsDetail, ok := aggs.(map[string]interface{})
	if !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("The aggregation of dsl is not a map")
	}
	if len(aggsDetail) != 1 {
		// 并行聚合不支持
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("Multiple aggregation is not supported")
	}
	// aggsDetail 的 key 是聚合名称，value 是个定义了聚合的 map
	for aggName, value := range aggsDetail {
		// key 是聚合名称
		// 各层聚合名称不能相同
		if _, exists := aggNameMap[aggName]; exists {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
				WithErrorDetails("The aggregation names of each layer aggregation cannot be the same")
		}
		*aggsLayers++
		aggNameMap[aggName] = *aggsLayers

		// value 是个map，map 中有两个元素，一个 key 是 aggType，一个 key 是 aggs 或者 aggregations
		aggValues, ok := value.(map[string]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
				WithErrorDetails("The aggregation of dsl is not a map")
		}
		for aggType, aggValue := range aggValues {
			switch aggType {
			case interfaces.AGGS:
				// 递归遍历子聚合
				err := validAggs(ctx, aggValue, metricModel, aggsLayers, aggNameMap, dateHistogramAggName, metricAggName, containTopHits)
				if err != nil {
					return err
				}
			case interfaces.TERMS, interfaces.FILTERS, interfaces.RANGE, interfaces.DATE_RANGE:
				// 分桶聚合类型： date_histogram（日期直方图）、terms（词条）、filters（过滤）、range（范围）、date_range（日期范围）、multi_terms（多词条）
				// do nothing
			case interfaces.DATE_HISTOGRAM:
				// 日期直方图的聚合名称需是时间字段的名称，此处解析就把名称留下
				dateHisto := aggValue.(map[string]interface{})

				if _, intervalExists := dateHisto["interval"]; intervalExists {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
						WithErrorDetails("The interval has been abandoned")
				}

				_, fixedIntervalExists := dateHisto["fixed_interval"]
				_, calendarIntervalExists := dateHisto["calendar_interval"]
				if fixedIntervalExists == calendarIntervalExists {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
						WithErrorDetails("The date_histogram aggregation statement is incorrect")
				}
				if calendarIntervalExists {
					metricModel.IsCalendarInterval = 1
				}

				*dateHistogramAggName = append(*dateHistogramAggName, aggName)
			case interfaces.VALUE_COUNT, interfaces.CARDINALITY, interfaces.SUM, interfaces.AVG, interfaces.MAX, interfaces.MIN:
				// 值聚合, 需要把值聚合的聚合名称留下，需与配置的度量字段相同
				*metricAggName = append(*metricAggName, aggName)
			case interfaces.TOP_HITS:
				// 判断度量字段是不是在includes字段里。
				*containTopHits = true
				err := validTopHits(ctx, aggValue, metricModel.MeasureField, metricAggName, aggNameMap, aggsLayers)
				if err != nil {
					return err
				}
			default:
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
					WithErrorDetails("Unsupport aggregation type in dsl.")
			}
		}
	}
	return nil
}

// top_hits 的参数校验
func validTopHits(ctx context.Context, aggValue interface{}, measureFiled string, metricAggName *[]string,
	aggNameMap map[string]int, aggsLayers *int) error {

	var topHit topHits
	topAgg, err := sonic.Marshal(aggValue)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
			WithErrorDetails(fmt.Sprintf("TopHits marshal error: %s", err.Error()))
	}
	err = sonic.Unmarshal(topAgg, &topHit)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
			WithErrorDetails(fmt.Sprintf("TopHits Unmarshal error: %s", err.Error()))
	}

	if topHit.Size <= 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("TopHits's numHits must be > 0")
	}

	if topHit.Source.Includes == nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("TopHits's includes must not be empty")
	}

	// 校验模型的独立字段是否包含在 top_hits 的 includes 中
	equal := false
	for _, field := range topHit.Source.Includes {
		if measureFiled == field {
			equal = true
			break
		}
	}
	if equal {
		*metricAggName = append(*metricAggName, measureFiled)
		aggNameMap[measureFiled] = *aggsLayers
	} else {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("The measure field must be one of the includes in top_hits")
	}
	return nil
}

type topHits struct {
	Size   int         `json:"size"`
	Sort   interface{} `json:"sort"`
	Source source      `json:"_source"`
}

type source struct {
	Includes []string `json:"includes"`
}

// 目标模型必要创建参数的非空校验。
func ValidateObjectiveModel(ctx context.Context, objectiveModel *interfaces.ObjectiveModel) error {
	// 校验目标模型id的合法性
	err := validateModelID(ctx, objectiveModel.ModelID, 0)
	if err != nil {
		return err
	}

	// 校验名称合法性: 不为空且不超过40个字符
	err = validateObjectName(ctx, objectiveModel.ModelName, interfaces.OBJECTTYPE_OBJECTIVE_MODEL)
	if err != nil {
		return err
	}

	// 校验目标类型非空
	if objectiveModel.ObjectiveType == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_ObjectiveType)
	}
	// 目标类型是枚举中的某个值
	if !interfaces.IsValidObjectiveType(objectiveModel.ObjectiveType) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_UnsupportObjectiveType)
	}

	// 目标配置非空
	if objectiveModel.ObjectiveConfig == nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_ObjectiveConfig)
	}

	// 查询语言是 promql，则时间维度必须是 @timestamp，时间格式必须是 epoch_millis
	if objectiveModel.ObjectiveType == interfaces.SLO {
		// 把 objective_config转成 SLOObjective
		var sloObjective interfaces.SLOObjective
		jsonData, err := sonic.Marshal(objectiveModel.ObjectiveConfig)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_ObjectiveConfig).
				WithErrorDetails(fmt.Sprintf("[%s]'s ObjectiveConfig Marshal error: %s", objectiveModel.ModelName, err.Error()))
		}
		err = sonic.Unmarshal(jsonData, &sloObjective)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_ObjectiveConfig).
				WithErrorDetails(fmt.Sprintf("[%s]'s SLO Objective Unmarshal error: %s", objectiveModel.ModelName, err.Error()))
		}

		// 目标非空，大于0
		if sloObjective.Objective == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_Objective)
		}
		if *(sloObjective.Objective) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_Objective)
		}

		// 周期非空
		if sloObjective.Period == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_Period)
		}
		// 周期小于0
		if *(sloObjective.Period) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_Period)
		}

		// 良好指标非空
		if sloObjective.GoodMetricModel == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_GoodMetricModel)
		}
		if sloObjective.GoodMetricModel.ID == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_GoodMetricModelID)
		}

		// 总指标非空
		if sloObjective.TotalMetricModel == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_TotalMetricModel)
		}
		if sloObjective.TotalMetricModel.ID == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_TotalMetricModelID)
		}

		// 状态阈值状态校验
		if sloObjective.StatusConfig != nil {
			// 状态非空, 区间不能重叠
			err = validStatusConfig(ctx, sloObjective.StatusConfig.Ranges, objectiveModel)
			if err != nil {
				return err
			}
		}
		objectiveModel.ObjectiveConfig = sloObjective
	} else {
		// kpi
		// 把 objective_config转成 KPIObjective
		var kpiObjective interfaces.KPIObjective
		jsonData, err := sonic.Marshal(objectiveModel.ObjectiveConfig)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_ObjectiveConfig).
				WithErrorDetails(fmt.Sprintf("[%s]'s ObjectiveConfig Marshal error: %s", objectiveModel.ModelName, err.Error()))
		}
		err = sonic.Unmarshal(jsonData, &kpiObjective)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_ObjectiveConfig).
				WithErrorDetails(fmt.Sprintf("[%s]'s KPI Objective Unmarshal error: %s", objectiveModel.ModelName, err.Error()))
		}

		// 目标非空，大于0
		if kpiObjective.Objective == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_Objective)
		}
		if *(kpiObjective.Objective) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_Objective)
		}

		// 目标单位非空
		if kpiObjective.Unit == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_ObjectiveUnit)
		}

		// 目标单位无效
		if kpiObjective.Unit != interfaces.UNIT_NUM_NONE && kpiObjective.Unit != interfaces.UNIT_NUM_PERCENT {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_ObjectiveUnit).
				WithErrorDetails(fmt.Sprintf(`expected objective unit is one of [none, %%], actual unit is %s`, kpiObjective.Unit))
		}

		// 综合计算指标非空
		if kpiObjective.ComprehensiveMetricModels == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_ComprehensiveMetricModels)
		}

		// 综合计算指标个数大于10
		if len(kpiObjective.ComprehensiveMetricModels) > interfaces.COMPREHENSIVE_METRIC_TOTAL {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_CountExceeded_ComprehensiveMetricModelsTotal)
		}

		var weight int64
		for _, mm := range kpiObjective.ComprehensiveMetricModels {
			// id非空
			if mm.ID == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_ComprehensiveMetricModelID)
			}
			// 权重非空且大于0
			if mm.Weight == nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_ComprehensiveWeight)
			}
			if *(mm.Weight) <= 0 {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_ComprehensiveWeight).
					WithErrorDetails("the weight parameter must greater than zero")
			}
			weight += *(mm.Weight)
		}
		// 权重之和等于100
		if weight != 100 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				derrors.DataModel_ObjectiveModel_InvalidParameter_ComprehensiveWeight).
				WithErrorDetails(fmt.Sprintf("The sum of weights must equal 100, actual is %d", weight))
		}

		// 附加计算指标
		if len(kpiObjective.AdditionalMetricModels) > interfaces.ADDITIONAL_METRIC_TOTAL {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_CountExceeded_AdditionalMetricModelsTotal)
		}
		for _, mm := range kpiObjective.AdditionalMetricModels {
			// id非空
			if mm.ID == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_AdditionalMetricModelID)
			}
		}
		// 状态阈值状态校验
		if kpiObjective.StatusConfig != nil {
			// 状态非空, 区间不能重叠
			err = validStatusConfig(ctx, kpiObjective.StatusConfig.Ranges, objectiveModel)
			if err != nil {
				return err
			}
		}
		objectiveModel.ObjectiveConfig = kpiObjective
	}

	// 若输入了 tags，校验 tags 的合法性
	err = ValidateTags(ctx, objectiveModel.Tags)
	if err != nil {
		return err
	}

	// 去掉tag前后空格以及数组去重
	objectiveModel.Tags = libCommon.TagSliceTransform(objectiveModel.Tags)

	// 校验comment合法性
	err = validateObjectComment(ctx, objectiveModel.Comment)
	if err != nil {
		return err
	}

	// 校验任务信息，任务不能为空
	if objectiveModel.Task == nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_Task)
	}
	err = validateObjectiveTask(ctx, objectiveModel.Task)
	if err != nil {
		return err
	}

	return nil
}

// 校验状态区间配置
func validStatusConfig(ctx context.Context, ranges []interfaces.Range, objectiveModel *interfaces.ObjectiveModel) error {
	if len(ranges) > interfaces.STATUS_TOTAL {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_CountExceeded_StatusTotal)
	}

	rgs := make([]interfaces.Range, 0)
	rgs = append(rgs, ranges...)
	for i, rg := range rgs {
		if rg.Status == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_StatusRanges).
				WithErrorDetails(fmt.Sprintf("Exist empty status in the [%s]'s status config", objectiveModel.ModelName))
		}

		if rg.From == nil && rg.To == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_StatusRanges).
				WithErrorDetails(fmt.Sprintf("Exist both from and to are empty in the [%s]'s status config", objectiveModel.ModelName))
		} else if rg.From == nil {
			rgs[i].From = &interfaces.NEG_INF
		} else if rg.To == nil {
			rgs[i].To = &interfaces.POS_INF
		} else if *rg.From >= *rg.To {
			// 如果from比to大，报错
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_StatusRanges).
				WithErrorDetails(fmt.Sprintf("Range's From is greater than Range's To in the [%s]'s status config", objectiveModel.ModelName))
		}
	}

	// 区间不能重叠
	if common.HasOverlappingRanges(rgs) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_StatusRanges).
			WithErrorDetails(fmt.Sprintf("Exist overlapping range in the [%s]'s status config", objectiveModel.ModelName))
	}
	return nil
}

// 目标模型持久化任务信息校验
func validateObjectiveTask(ctx context.Context, task *interfaces.MetricTask) error {

	// 1. 执行频率
	// if task.Schedule == (interfaces.Schedule{}) {
	// 	return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_Schedule)
	// }
	// 1.1 执行类型
	if task.Schedule.Type == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_ScheduleType)
	}
	if task.Schedule.Type != interfaces.SCHEDULE_TYPE_FIXED && task.Schedule.Type != interfaces.SCHEDULE_TYPE_CRON {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_ScheduleType)
	}
	// 1.2 执行表达式
	if task.Schedule.Expression == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_ScheduleExpression)
	}
	if task.Schedule.Type == interfaces.SCHEDULE_TYPE_FIXED {
		// 不为空时校验有效性,time duration的格式，单位支持 m - 分钟； h - 小时； d - 天
		err := validateDuration(ctx, task.Schedule.Expression, common.DurationDayHourMinuteRE,
			derrors.DataModel_ObjectiveModel_InvalidParameter_ScheduleExpression, "schedule expression ", true)
		if err != nil {
			return err
		}
	} else {
		// cron 表达式的校验。只支持6位，不支持年的指定。
		_, err := CronParser.Parse(task.Schedule.Expression)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_ScheduleExpression).
				WithErrorDetails(err.Error())
		}
	}

	// 2. 追溯时长。不为空时，校验时间区间的有效性， h - 小时； d - 天；
	var retraceDurationV time.Duration
	if task.RetraceDuration != "" {
		//追溯时长不为空时，需要校验有效性
		err := validateDuration(ctx, task.RetraceDuration, common.DurationDayHourRE,
			derrors.DataModel_ObjectiveModel_InvalidParameter_RetraceDuration, "retrace duration", false)
		if err != nil {
			return err
		}

		// 解析追溯时长
		retraceDurationV, err = common.ParseDuration(task.RetraceDuration, common.DurationDayHourRE, false)
		if err != nil {
			return err
		}
	}

	// 3. 持久化步长
	if len(task.Steps) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_Step)
	} else {
		// 是否有效，是否重复。
		stepsMap := make(map[string]interface{})
		for _, step := range task.Steps {
			_, exists := common.PersistStepsMap[step]
			if !exists {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_Step).
					WithErrorDetails(fmt.Sprintf("expect persist steps is one of %v, actaul is %s",
						common.PersistSteps, step))
			}
			// 不允许有重复的步长
			if _, ok := stepsMap[step]; !ok {
				stepsMap[step] = nil
			} else {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_Duplicated_TaskStep)
			}

			// 解析持久化步长
			stepV, err := common.ParseDuration(step, common.DurationDayHourMinuteRE, true)
			if err != nil {
				return err
			}

			// 能追溯的数据时间点数小于1w
			if retraceDurationV > 0 && (float64(retraceDurationV)/float64(stepV)) > interfaces.MAX_RETRACE_POINTS_NUM {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter_RetraceDuration).
					WithErrorDetails(fmt.Sprintf("The task of retrace data point can not exceed 10000, actual is [%s]", retraceDurationV/stepV))
			}
		}

	}
	// 4. 索引库类型
	if task.IndexBase == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_NullParameter_IndexBase)
	}

	// 5. 备注
	if utf8.RuneCountInString(task.Comment) > interfaces.COMMENT_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_LengthExceeded_TaskComment).
			WithErrorDetails(fmt.Sprintf("The length of the task comment exceeds %v", interfaces.COMMENT_MAX_LENGTH))
	}

	return nil
}

// 校验过滤条件
func validateCond(ctx context.Context, cfg *interfaces.CondCfg) error {
	if cfg == nil {
		return nil
	}

	// 判断过滤器是否为空对象 {}
	if cfg.Name == "" && cfg.Operation == "" && len(cfg.SubConds) == 0 && cfg.ValueFrom == "" && cfg.Value == nil {
		return nil
	}

	// 过滤条件字段不允许 __id 和 __routing
	if cfg.Name == "__id" || cfg.Name == "__routing" {
		return rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_Forbidden_FilterField).
			WithErrorDetails("The filter field '__id' and '__routing' is not allowed")
	}

	// 过滤操作符
	if cfg.Operation == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_NullParameter_FilterOperation)
	}

	_, exists := dcond.OperationMap[cfg.Operation]
	if !exists {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_UnsupportFilterOperation).
			WithErrorDetails(fmt.Sprintf("unsupport condition operation %s", cfg.Operation))
	}

	switch cfg.Operation {
	case dcond.OperationAnd, dcond.OperationOr:
		// 子过滤条件不能超过10个
		if len(cfg.SubConds) > dcond.MaxSubCondition {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataView_CountExceeded_Filters).
				WithErrorDetails(fmt.Sprintf("The number of subConditions exceeds %d", dcond.MaxSubCondition))
		}

		for _, subCond := range cfg.SubConds {
			err := validateCond(ctx, subCond)
			if err != nil {
				return err
			}
		}
	default:
		// 过滤字段名称不能为空
		if cfg.Name == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_NullParameter_FilterName)
		}

		// 除了 exist, not_exist, empty, not_empty 外需要校验 value_from
		if _, ok := dcond.NotRequiredValueOperationMap[cfg.Operation]; !ok {
			if cfg.ValueFrom != dcond.ValueFrom_Const {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_ValueFrom).
					WithErrorDetails(fmt.Sprintf("condition does not support value_from type('%s')", cfg.ValueFrom))
			}
		}
	}

	switch cfg.Operation {
	case dcond.OperationEq, dcond.OperationNotEq, dcond.OperationGt, dcond.OperationGte,
		dcond.OperationLt, dcond.OperationLte, dcond.OperationLike, dcond.OperationNotLike,
		dcond.OperationRegex, dcond.OperationMatch, dcond.OperationMatchPhrase, dcond.OperationCurrent:
		// 右侧值为单个值
		_, ok := cfg.Value.([]interface{})
		if ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a single value", cfg.Operation))
		}

		if cfg.Operation == dcond.OperationLike || cfg.Operation == dcond.OperationNotLike ||
			cfg.Operation == dcond.OperationPrefix || cfg.Operation == dcond.OperationNotPrefix {
			_, ok := cfg.Value.(string)
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
					WithErrorDetails("[like not_like prefix not_prefix] operation's value should be a string")
			}
		}

		if cfg.Operation == dcond.OperationRegex {
			val, ok := cfg.Value.(string)
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
					WithErrorDetails("[regex] operation's value should be a string")
			}

			_, err := regexp2.Compile(val, regexp2.RE2)
			if err != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
					WithErrorDetails(fmt.Sprintf("[regex] operation regular expression error: %s", err.Error()))
			}

		}

	case dcond.OperationIn, dcond.OperationNotIn:
		// 当 operation 是 in, not_in 时，value 为任意基本类型的数组，且长度大于等于1；
		_, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[in not_in] operation's value must be an array")
		}

		if len(cfg.Value.([]interface{})) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[in not_in] operation's value should contains at least 1 value")
		}
	case dcond.OperationRange, dcond.OperationOutRange, dcond.OperationBetween:
		// 当 operation 是 range 时，value 是个由范围的下边界和上边界组成的长度为 2 的数值型数组
		// 当 operation 是 out_range 时，value 是个长度为 2 的数值类型的数组，查询的数据范围为 (-inf, value[0]) || [value[1], +inf)
		v, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[range, out_range, between] operation's value must be an array")
		}

		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[range, out_range, between] operation's value must contain 2 values")
		}
	case dcond.OperationBefore:
		// before时, 长度为2的数组，第一个值为时间长度，数值型；第二个值为时间单位，字符串
		v, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[before] operation's value must be an array")
		}

		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[before] operation's value must contain 2 values")
		}
		_, err := common.AssertFloat64(v[0])
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[before] operation's first value should be a number")
		}

		_, ok = v[1].(string)
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[before] operation's second value should be a string")
		}
	}

	return nil
}

func validateHavingCondition(ctx context.Context, havingCondition *interfaces.CondCfg) error {
	if havingCondition.Name != interfaces.VALUE_FIELD_NAME {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter_HavingConditionName)
	}
	if havingCondition.Operation == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_HavingConditionOperation)
	}
	// operation需是有效的
	_, exists := dcond.HavingOperationMap[havingCondition.Operation]
	if !exists {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_UnsupportHavingConditionOperation).
			WithErrorDetails(fmt.Sprintf("unsupport having condition operation %s", havingCondition.Operation))
	}

	if havingCondition.Value == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_NullParameter_HavingConditionOperation)
	}

	switch havingCondition.Operation {
	case dcond.OperationEq, dcond.OperationNotEq, dcond.OperationGt, dcond.OperationGte,
		dcond.OperationLt, dcond.OperationLte:
		// 右侧值为单个值
		_, ok := havingCondition.Value.([]interface{})
		if ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a single value", havingCondition.Operation))
		}
		// 值为数值类型
		_, err := common.AssertFloat64(havingCondition.Value)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a number in having condition", havingCondition.Operation))
		}
	case dcond.OperationIn, dcond.OperationNotIn:
		// 当 operation 是 in, not_in 时，value 为任意基本类型的数组，且长度大于等于1；
		_, ok := havingCondition.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[in not_in] operation's value must be an array")
		}

		if len(havingCondition.Value.([]interface{})) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[in not_in] operation's value should contains at least 1 value")
		}

		if !common.IsSameType(havingCondition.Value.([]any)) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("condition [not_in] right value should be an array composed of elements of same type")
		}
		// 值为数值类型
		_, err := common.AssertFloat64(havingCondition.Value.([]any)[0])
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a number slice in having condition", havingCondition.Operation))
		}
	case dcond.OperationRange, dcond.OperationOutRange:
		// 当 operation 是 range 时，value 是个由范围的下边界和上边界组成的长度为 2 的数值型数组
		// 当 operation 是 out_range 时，value 是个长度为 2 的数值类型的数组，查询的数据范围为 (-inf, value[0]) || [value[1], +inf)
		v, ok := havingCondition.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[range, out_range] operation's value must be an array")
		}

		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("[range, out_range] operation's value must contain 2 values")
		}

		if !common.IsSameType(havingCondition.Value.([]any)) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails("condition [range, out_range] right value should be an array composed of elements of same type")
		}
		// 值为数值类型
		_, err := common.AssertFloat64(havingCondition.Value.([]any)[0])
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_FilterValue).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a number slice in having condition", havingCondition.Operation))
		}
	}
	return nil
}
