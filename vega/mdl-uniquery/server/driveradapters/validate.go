package driveradapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/dlclark/regexp2"
	"github.com/go-viper/mapstructure/v2"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"uniquery/common"
	cond "uniquery/common/condition"
	"uniquery/common/convert"
	vopt "uniquery/common/value_opt"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
)

/*
	公共的校验函数, 适用于所有模块. 包括:
	  	(1) 分页查询列表参数的校验.
*/

// 分页查询列表参数的校验
func validatePaginationQueryParams(ctx context.Context, params interfaces.PaginationQueryParams, category string) (interfaces.PaginationQueryParams, error) {
	// 1. 校验offset
	if params.Offset < interfaces.MIN_OFFSET_NUM_OF_LIST {
		return params, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Offset).
			WithErrorDetails(fmt.Sprintf("The offset is not greater than %d", interfaces.MIN_OFFSET_NUM_OF_LIST))
	}

	// 2. 校验limit
	if params.Limit == 0 {
		params.Limit = interfaces.DEFAULT_LIMIT_NUM_OF_LIST
	}

	if !(params.Limit >= interfaces.MIN_LIMIT_NUM_OF_LIST && params.Limit <= interfaces.MAX_LIMIT_NUM_OF_LIST) {
		return params, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Limit).
			WithErrorDetails(fmt.Sprintf("The number per page does not in the range of [%d,%d]", interfaces.MIN_LIMIT_NUM_OF_LIST, interfaces.MAX_LIMIT_NUM_OF_LIST))
	}

	// 3. 校验offset+limit
	if sum := params.Offset + params.Limit; sum > interfaces.MAX_SEARCH_SIZE {
		return params, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_OffestAndLimitSum).
			WithErrorDetails(fmt.Sprintf("Search scope is too large, from + size must be less than or equal to: [%v] but was [%v]", interfaces.MAX_SEARCH_SIZE, sum))
	}

	if params.Sort == "" {
		params.Sort = interfaces.DEFAULT_SORT
	}

	// 4. 校验sort
	switch category {
	case interfaces.QUERY_CATEGORY_SPAN_LIST:
		val, ok := interfaces.SPAN_LIST_SORT[params.Sort]
		if !ok {
			types := make([]string, 0)
			for t := range interfaces.SPAN_LIST_SORT {
				types = append(types, t)
			}
			return params, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Sort).
				WithErrorDetails(fmt.Sprintf("Wrong sort type, does not belong to any item in set %v ", types))
		}
		params.Sort = val
	case interfaces.QUERY_CATEGORY_RELATED_LOG_LIST:
		val, ok := interfaces.RELATED_LOG_LIST_SORT[params.Sort]
		if !ok {
			types := make([]string, 0)
			for t := range interfaces.RELATED_LOG_LIST_SORT {
				types = append(types, t)
			}
			return params, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Sort).
				WithErrorDetails(fmt.Sprintf("Wrong sort type, does not belong to any item in set %v ", types))
		}
		params.Sort = val
	default:
		return params, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InternalError).
			WithErrorDetails("Invalid query category of func 'validatePaginationQueryParams'")
	}

	// 5. 校验direction
	if params.Direction == "" {
		params.Direction = interfaces.DEFAULT_DIRECTION
	}

	if params.Direction != interfaces.DESC_DIRECTION && params.Direction != interfaces.ASC_DIRECTION {
		return params, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Direction).
			WithErrorDetails("The sort direction is not desc or asc")
	}

	return params, nil
}

// 校验ARDataView
func ValidateDataView(ctx context.Context, dataViewId, dataViewType string) error {
	if dataViewId == "" {
		switch dataViewType {
		case "trace":
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_Trace_InvalidParameter_TraceDataViewID).
				WithErrorDetails("The trace_data_view_id is null")
		case "log":
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_Trace_InvalidParameter_LogDataViewID).
				WithErrorDetails("The log_data_view_id is null")
		}
	}
	return nil
}

// 校验spanStatuses
// func ValidateSpanStatuses(ctx context.Context, spanStatuses string) (map[string]int, error) {
// 	statusMap := map[string]int{}
// 	if spanStatuses == "" {
// 		return statusMap, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_Trace_InvalidParameter_SpanStatuses).
// 			WithErrorDetails("The span_statuses is null")
// 	}

// 	// 去掉字符串左右两边的括号和空格
// 	spanStatuses = strings.Trim(spanStatuses, "{} []<>")

// 	// 按逗号分隔
// 	statusSlice := strings.Split(spanStatuses, ",")

// 	for _, status := range statusSlice {
// 		if status != "Ok" && status != "Error" && status != "Unset" {
// 			return statusMap, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_Trace_InvalidParameter_SpanStatuses).
// 				WithErrorDetails("Not all of the elements in span_statuses belong to the set of [" + interfaces.DEFAULT_SPAN_STATUSES + "]")
// 		}

// 		statusMap[status] = 1
// 	}
// 	return statusMap, nil
// }

// 校验startTime和endTime
// func ValidateSpanQueryTime(ctx context.Context, startTime, endTime string) (int64, int64, error) {
// 	if startTime == "" {
// 		return 0, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_Trace_InvalidParameter_StartTime).
// 			WithErrorDetails("The start_time is null")
// 	}

// 	start, err := strconv.ParseInt(startTime, 10, 64)
// 	if err != nil {
// 		return 0, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_Trace_InvalidParameter_StartTime).
// 			WithErrorDetails(fmt.Sprintf("The start_time cannot be converted to decimal int64, err: %v", err.Error()))
// 	}

// 	if len(startTime) != interfaces.UNIX_MILLISECOND_TIMESTAMP_STR_LENGTH {
// 		return 0, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_Trace_InvalidParameter_StartTime).
// 			WithErrorDetails(fmt.Sprintf("The start_time is not a Unix millisecond timestamp because its length is not equal to %d", interfaces.UNIX_MILLISECOND_TIMESTAMP_STR_LENGTH))
// 	}

// 	if endTime == "" {
// 		return 0, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_Trace_InvalidParameter_EndTime).
// 			WithErrorDetails("The end_time is null")
// 	}

// 	end, err := strconv.ParseInt(endTime, 10, 64)
// 	if err != nil {
// 		return 0, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_Trace_InvalidParameter_EndTime).
// 			WithErrorDetails(fmt.Sprintf("The end_time cannot be converted to decimal int64, err: %v", err.Error()))
// 	}

// 	if len(endTime) != interfaces.UNIX_MILLISECOND_TIMESTAMP_STR_LENGTH {
// 		return 0, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_Trace_InvalidParameter_EndTime).
// 			WithErrorDetails(fmt.Sprintf("The end_time is not a Unix millisecond timestamp because its length is not equal to %d", interfaces.UNIX_MILLISECOND_TIMESTAMP_STR_LENGTH))
// 	}

// 	// end比start小, 抛异常
// 	if end < start {
// 		return 0, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_Trace_InvalidParameter_EndTime).
// 			WithErrorDetails("The end_time is before start_time")
// 	}

// 	// 以下代码暂时注释掉的原因: 前端界面上允许选择未来时间
// 	// end是未来时间, 抛异常, 附带当前时间
// 	// currentTime := time.Now()
// 	// current := currentTime.UnixMilli()
// 	// if end > current {
// 	// 	return 0, 0, rest.NewHTTPError(http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_EndTime).
// 	// 		WithErrorDetails(fmt.Sprintf("The end_time is greater than current time, current timestamp is %v.", current))
// 	// }

// 	return start, end, nil
// }

// 校验分页参数offset和limit
// func ValidateOffsetAndLimit(ctx context.Context, offsetStr, limitStr string) (int, int, error) {
// 	if offsetStr == "" {
// 		return 0, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Offset).
// 			WithErrorDetails("The offset is null")
// 	}
// 	offset, err := strconv.Atoi(offsetStr)
// 	if err != nil {
// 		return 0, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Offset).
// 			WithErrorDetails(fmt.Sprintf("The offset cannot be converted to an integer, err:%v", err))
// 	}
// 	if offset < interfaces.MIN_OFFSET {
// 		return 0, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Offset).
// 			WithErrorDetails("The offset is not greater than " + fmt.Sprint(interfaces.MIN_OFFSET))
// 	}

// 	if limitStr == "" {
// 		return 0, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Limit).
// 			WithErrorDetails("The limit is null")
// 	}

// 	limit, err := strconv.Atoi(limitStr)
// 	if err != nil {
// 		return 0, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Limit).
// 			WithErrorDetails(fmt.Sprintf("The limit cannot be converted to an integer, err:%v", err))
// 	}
// 	if limit < interfaces.MIN_LIMIT || limit > interfaces.MAX_LIMIT {
// 		errStr := fmt.Sprintf("The number per page is not in the range of [%d,%d]", interfaces.MIN_LIMIT, interfaces.MAX_LIMIT)
// 		return 0, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Limit).
// 			WithErrorDetails(errStr)
// 	}

// 	return offset, limit, nil
// }

// 校验 methodOverride. 以 header 为准，header 中的 method 为空，则取 body 中的 method 进行比较
// func validateMethodOverride(ctx context.Context, headerMethod string, query *interfaces.MetricModelQuery) error {
// 	err := ValidateHeaderMethodOverride(ctx, headerMethod)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// 校验 x-http-method-override 重载方法，只在header里传递 method
func ValidateHeaderMethodOverride(ctx context.Context, headerMethod string) error {
	// 校验 method
	if headerMethod == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_NullParameter_OverrideMethod)
	}
	if headerMethod != "GET" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_OverrideMethod).
			WithErrorDetails(fmt.Sprintf("X-HTTP-Method-Override is expected to be GET, but it is actually %s", headerMethod))
	}

	return nil
}

// 指标模型必要参数的非空校验
func ValidateMetricModelSimulate(ctx context.Context, query *interfaces.MetricModelQuery) error {

	// 校验指标类型非空
	if query.MetricType == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_MetricType)
	}
	// 指标类型是枚举中的某个值
	if !interfaces.IsValidMetricType(query.MetricType) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_UnsupportMetricType).
			WithErrorDetails(fmt.Sprintf("Unsupport metric type %s", query.MetricType))
	}
	// 原子指标时数据视图非空
	switch query.MetricType {
	case interfaces.ATOMIC_METRIC:
		if query.DataSource == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_DataSource)
		} else {
			// 校验数据源类型
			// if !interfaces.IsValidDataSourceType(query.DataSource.Type) {
			// 	return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_DataSourceType)
			// }
			// 数据源id不为空
			if query.DataSource.ID == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_DataSourceID)
			}
		}

		// 查询语言非空
		if query.QueryType == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_QueryType)
		}
		// 查询语言是枚举中的某个值
		if !interfaces.IsValidQueryType(query.QueryType) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_UnsupportQueryType).
				WithErrorDetails(fmt.Sprintf("Unsupport query type %s", query.QueryType))
		}

		// sql的计算公式是在sql_config中
		if query.QueryType == interfaces.SQL {
			// 校验sql
			if query.FormulaConfig == nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_FormulaConfig)
			}
			//  把 formula_config转成 SQLConfig
			var sqlConfig interfaces.SQLConfig
			jsonData, err := json.Marshal(query.FormulaConfig)
			if err != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_SqlConfig).
					WithErrorDetails(fmt.Sprintf("SQL Config Marshal error: %s", err.Error()))
			}
			err = json.Unmarshal(jsonData, &sqlConfig)
			if err != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_SqlConfig).
					WithErrorDetails(fmt.Sprintf("SQL Config Unmarshal error: %s", err.Error()))
			}

			// 度量计算不为空
			if sqlConfig.AggrExpr == nil && sqlConfig.AggrExprStr == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_SQLAggrExpression)
			}

			// aggr 的 code 和 config不能同时存在
			if sqlConfig.AggrExpr != nil {
				// code为空,配置不为空时,聚合函数和聚合字段都不能为空
				if sqlConfig.AggrExprStr == "" {
					if sqlConfig.AggrExpr.Aggr == "" || sqlConfig.AggrExpr.Field == "" {
						return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_SQLAggrExpression)
					}
				} else {
					// code不为空,且配置项也都不为空,报错(不能同时存在)
					if sqlConfig.AggrExpr.Aggr != "" && sqlConfig.AggrExpr.Field != "" {
						return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_SQLAggrExpression)
					}
					// 配置化对象不为空,配置项不完整,code不为空,则以code为准.配置话置空
					sqlConfig.AggrExpr = nil
				}
			}

			// code 和 config不能同时存在
			if sqlConfig.Condition != nil && sqlConfig.ConditionStr != "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_SQLCondition)
			}

			// 校验视图过滤条件(与数据视图中的一样)：操作符、字段类型和操作符是否匹配
			err = validateCond(ctx, sqlConfig.Condition)
			if err != nil {
				return err
			}

			// FormulaConfig 赋值为 SqlConfig
			query.FormulaConfig = sqlConfig

			// 时间字段非空
			// if query.DateField == "" {
			// 	return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_DateField)
			// }
			// sql的数据源类型当前只有vega逻辑视图 vega_view
			// if query.DataSource.Type != interfaces.QueryType_SQL {
			// 	return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_DataSourceType)
			// }
		} else {
			// 计算公式非空
			if query.QueryType != interfaces.DSL_CONFIG && query.Formula == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_Formula)
			}
			// 非sql的数据源类型是数据视图data_view
			// if query.DataSource.Type != interfaces.DSL {
			// 	return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_DataSourceType)
			// }
		}

		// 预览时还需要校验计算公式有效性, 预览是直接从前端请求接口。
		// 如果是保存的有效性检查，dsl的规则已经在data-model中做过了，无需再做，此时应该增加 flag 绕过。
		if query.QueryType == interfaces.DSL && !query.IsModelRequest {
			err := validateDSL(ctx, query)
			if err != nil {
				return err
			}
		}

	case interfaces.DERIVED_METRIC:
		// 校验衍生指标配置
		if query.FormulaConfig == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_FormulaConfig)
		}
		// 把 formula_config转成 DerivedConfig
		var derivedConfig interfaces.DerivedConfig
		jsonData, err := sonic.Marshal(query.FormulaConfig)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_SqlConfig).
				WithErrorDetails(fmt.Sprintf("SQL Config Marshal error: %s", err.Error()))
		}
		err = json.Unmarshal(jsonData, &derivedConfig)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_SqlConfig).
				WithErrorDetails(fmt.Sprintf("SQL Config Unmarshal error: %s", err.Error()))
		}

		// 依赖模型不能为空
		if derivedConfig.DependMetricModel == nil || derivedConfig.DependMetricModel.ID == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_DependMetricModel)
		}

		if derivedConfig.ConditionStr == "" {
			if derivedConfig.DateCondition == nil && derivedConfig.BusinessCondition == nil {
				// 都为空，error
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_DerivedCondition)
			}
			// 校验过滤条件(与数据视图中的一样)：操作符、字段类型和操作符是否匹配
			if derivedConfig.DateCondition != nil {
				err = validateCond(ctx, derivedConfig.DateCondition)
				if err != nil {
					return err
				}
			}

			if derivedConfig.BusinessCondition != nil {
				err = validateCond(ctx, derivedConfig.BusinessCondition)
				if err != nil {
					return err
				}
			}
		}

		// code 和 config不能同时存在
		if (derivedConfig.DateCondition != nil || derivedConfig.BusinessCondition != nil) && derivedConfig.ConditionStr != "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_DerivedCondition)
		}

		// 存储时只存依赖模型的ID，去除名称
		derivedConfig.DependMetricModel = &interfaces.DependMetricModel{ID: derivedConfig.DependMetricModel.ID}
		// FormulaConfig 赋值为 DerivedConfig
		query.FormulaConfig = derivedConfig

	case interfaces.COMPOSITED_METRIC:
		// do nothing
	}

	// 范围查询: 若start end step 都是空，则默认最近半小时的数据，步长5分钟，兼容前面版本的指标数据预览
	if !query.IsInstantQuery {
		if query.IsCalendarInterval == 1 {
			// 如果是 calendar_interval,那么就不做step的校验
			if query.StepStr == nil {
				// 如果为空被修正默认修正为 minute
				min := interfaces.CALENDAR_STEP_MINUTE
				query.StepStr = &min
			}
			if _, exists := interfaces.CALENDAR_INTERVALS[*query.StepStr]; !exists {
				// 预览时不合法的日历间隔被修正默认修正为 minute
				min := interfaces.CALENDAR_STEP_MINUTE
				query.StepStr = &min
			}
		}
	}

	// 校验查询时间范围的相关参数
	err := validateQueryTimeParam(ctx, &query.QueryTimeParams)
	if err != nil {
		return err
	}

	// 校验 filters
	err = validateFilters(ctx, query.Filters)
	if err != nil {
		return err
	}

	// 校验同环比参数
	err = validateRequestMetrics(ctx, query)
	if err != nil {
		return err
	}

	// 校验排序字段的非空
	for _, orderBy := range query.OrderByFields {
		if orderBy.Name == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_OrderByName)
		}
		if orderBy.Direction == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_OrderByDirection)
		}
		if orderBy.Direction != interfaces.DESC_DIRECTION && orderBy.Direction != interfaces.ASC_DIRECTION {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_OrderByDirection).
				WithErrorDetails("The order direction is not desc or asc")
		}
	}

	// 校验 havingCondition 的有效性
	if query.HavingCondition != nil {
		err = validateHavingCondition(ctx, query.HavingCondition)
		if err != nil {
			return err
		}
	}

	return nil
}

// start end 的大小校验
func validateTimes(ctx context.Context, start time.Time, end time.Time) (time.Time, error) {
	// start 是未来时间就抛异常,附带当前时间
	// currentTime := time.Now()
	// if start.After(currentTime) {
	// 	return time.Time{}, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Start).
	// 		WithErrorDetails(fmt.Sprintf("Start is greater than current time, current time is %v.", currentTime.UnixMilli()))
	// }

	// // 如果end 大于 current_time，那么end = current
	// if end.After(currentTime) {
	// 	end = currentTime
	// }

	if end.Before(start) {
		return time.Time{}, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Start).
			WithErrorDetails("end timestamp must not be before start time")

	}

	return end, nil
}

// 基于指标模型的指标数据查询的参数校验
func validateMetricModelData(ctx context.Context, query *interfaces.MetricModelQuery) error {

	// 校验查询时间范围的相关参数
	err := validateQueryTimeParam(ctx, &query.QueryTimeParams)
	if err != nil {
		return err
	}

	// 校验 filters
	err = validateFilters(ctx, query.Filters)
	if err != nil {
		return err
	}

	// 校验同环比参数
	err = validateRequestMetrics(ctx, query)
	if err != nil {
		return err
	}

	// 校验排序字段的非空
	for _, orderBy := range query.OrderByFields {
		if orderBy.Name == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_OrderByName)
		}
		if orderBy.Direction == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_OrderByDirection)
		}
		if orderBy.Direction != interfaces.DESC_DIRECTION && orderBy.Direction != interfaces.ASC_DIRECTION {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_OrderByDirection).
				WithErrorDetails("The order direction is not desc or asc")
		}
	}

	// 校验 havingCondition 的有效性
	if query.HavingCondition != nil {
		err = validateHavingCondition(ctx, query.HavingCondition)
		if err != nil {
			return err
		}
	}

	return nil
}

// 对过滤器的过滤条件做校验
func validateFilters(ctx context.Context, filters []interfaces.Filter) error {
	for _, filter := range filters {
		// 非 multi_match才校验 field，建议这些校验都下沉到各个condition里校验，不统一校验
		if filter.Operation != cond.OperationMultiMatch && filter.Name == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_NullParameter_FilterName)
		}

		if filter.Value == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_NullParameter_FilterValue)
		}

		if filter.Operation == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_NullParameter_FilterOperation)
		}

		_, exists := cond.OperationMap[filter.Operation]
		if !exists {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_UnsupportFilterOperation)
		}

		// 当 operation 是 in 时，value 为任意基本类型的数组，且长度大于等于1；
		if filter.Operation == interfaces.OPERATION_IN {
			_, ok := filter.Value.([]interface{})
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
					WithErrorDetails("When Filter Operation is in, The Value must be an array. ")
			}

			if len(filter.Value.([]interface{})) <= 0 {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
					WithErrorDetails("When Filter Operation is in, The Value should contains at least 1 value")
			}
		} else if filter.Operation == cond.OperationRange {
			// 当 operation 是 range 时，value 是个由范围的下边界和上边界组成的长度为 2 的数值型数组
			v, ok := filter.Value.([]interface{})
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
					WithErrorDetails("When Filter Operation is range, The Value must be an array. ")
			}

			if len(v) != 2 {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
					WithErrorDetails("When Filter Operation is range, The Value must contains 2 value")
			}
		} else if filter.Operation == cond.OperationOutRange {
			// 当 operation 是 out_range 时，value 是个长度为 2 的数值类型的数组，查询的数据范围为 (-inf, value[0]) || [value[1], +inf)。
			v, ok := filter.Value.([]interface{})
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
					WithErrorDetails("When Filter Operation is out_range, The Value must be an array. ")
			}

			if len(v) != 2 {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
					WithErrorDetails("When Filter Operation is out_range, The Value must contains 2 value")
			}
		} else {
			// 当 operation 是 = 或 != 时，value 为任意基本类型的值
			_, ok := filter.Value.([]interface{})
			if ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
					WithErrorDetails("When Filter Operation is one of [=, !=, like, not like, >, >=, <, <=], The Value must not be an array, should be a basic type. ")
			}
			// }
		}

	}
	return nil
}

// 校验查询的时间相关的条件，预览和查询接口都可使用
func validateQueryTimeParam(ctx context.Context, query *interfaces.QueryTimeParams) error {

	if query.IsInstantQuery {
		// 优先看start end
		// 看start end是否为空，若start end不为空，则用end-start
		if query.Start == nil || query.End == nil {
			// 兼容旧的instant_query: 把 time, look_back_delta 转成 start end。
			// instant query 校验 time 即可。 time <= 0 修正为 now
			// time 是 end
			end := query.Time
			if end == 0 {
				end = time.Now().UnixMilli()
			}
			if end < 0 {
				httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Time)
				return httpErr
			}

			// 校验 look_back_delta。look_back_delta 是非必选
			var lookBackDelta int64
			if query.LookBackDeltaStr != "" {
				LookBackDeltaT, err := convert.ParseDuration(query.LookBackDeltaStr)
				if err != nil {
					httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_LookBackDelta).
						WithErrorDetails(err.Error())
					return httpErr
				}

				if LookBackDeltaT <= 0 {
					httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_LookBackDelta).
						WithErrorDetails("zero or negative lookBackDelta is not accepted.")
					return httpErr
				}
				lookBackDelta = LookBackDeltaT.Milliseconds()
			} else {
				// query.LookBackDeltaStr == ""
				lookBackDelta = interfaces.DEFAULT_LOOK_BACK_DELTA.Milliseconds()
			}
			if end <= lookBackDelta {
				httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Time)
				return httpErr
			}
			//
			start := end - lookBackDelta
			query.End = &end
			query.Start = &start
		}
	}

	// 校验 start
	if query.Start == nil || *query.Start <= 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Start)
		return httpErr
	}
	startT := time.UnixMilli(*query.Start)

	// 校验 end
	if query.End == nil || *query.End <= 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_End)
		return httpErr
	}
	endT := time.UnixMilli(*query.End)

	// 校验start end。判断场景  start > now, end > now, end < start
	endT, err := validateTimes(ctx, startT, endT)
	if err != nil {
		return err
	}
	et := endT.UnixMilli()
	query.End = &et

	// 趋势查询还需校验 step
	if !query.IsInstantQuery {
		if query.StepStr == nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_Step)
			return httpErr
		}
		// sql-原子指标，衍生指标，复合指标的步长只支持日历步长,不支持固定步长，在此处没法校验，需到service获取到指标模型信息之后校验
		// step 应属于系统步长集 15s, 30s, 1m, 2m, 5m, 10m, 15m, 20m, 30m, 1h, 2h, 3h, 6h, 12h, 1d, 1y,
		// minute, hour, day, month, quarter, year
		_, fixedExists := common.FixedStepsMap[*query.StepStr]
		_, calendarExists := interfaces.CALENDAR_INTERVALS[*query.StepStr]
		// 请求步长不在固定步长和日历步长集中，不合法
		if !fixedExists && !calendarExists {
			stepArr := make([]string, 0)
			stepArr = append(stepArr, common.FixedSteps...)
			stepArr = append(stepArr, []string{"minute", "hour", "day", "week", "month", "quarter", "year"}...)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Step).
				WithErrorDetails(fmt.Sprintf("expect steps is one of {%v}, actaul is %s", stepArr, *query.StepStr))
		}
		// 如果是 calendar_interval,那么就不做step的校验
		if fixedExists {
			err = validStep(ctx, query, endT, startT)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func validStep(ctx context.Context, query *interfaces.QueryTimeParams, endT time.Time, startT time.Time) error {
	stepT, err := convert.ParseDuration(*query.StepStr)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Step).
			WithErrorDetails(err.Error())
		return httpErr
	}

	if stepT <= 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Step).
			WithErrorDetails("zero or negative query resolution step widths are not accepted. Try a positive integer")
		return httpErr
	}

	// 如果step超过2h,提示用户将step调整为5分钟的倍数。原逻辑上step超过30min是5m的倍数，所以当路由改为2h后，此处取都能兼容的，用30m的约束
	if stepT > interfaces.SHARD_ROUTING_30M && stepT%interfaces.DEFAULT_STEP_DIVISOR != 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Step).
			WithErrorDetails("step should a multiple of 5 minutes when step > 30min")
		return httpErr
	}

	// For safety, limit the number of returned points per timeseries.
	// This is sufficient for 60s resolution for a week or 1h resolution for a year.
	if endT.Sub(startT)/stepT > 11000 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Step).
			WithErrorDetails("exceeded maximum resolution of 11,000 points per timeseries. Try decreasing the query resolution (?step=XX)")
		return httpErr
	}
	stepMs := stepT.Milliseconds()
	query.Step = &stepMs
	return nil
}

// 校验请求指标(同环比\占比)的参数
func validateRequestMetrics(ctx context.Context, query *interfaces.MetricModelQuery) error {

	if query.RequestMetrics == nil {
		return nil
	}

	if query.RequestMetrics.Type == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_RequestMetricsType)
	}

	if query.RequestMetrics.Type != interfaces.METRICS_SAMEPERIOD && query.RequestMetrics.Type != interfaces.METRICS_PROPORTION {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_RequestMetricsType)
	}

	if query.RequestMetrics.Type == interfaces.METRICS_SAMEPERIOD {
		if query.RequestMetrics.SamePeriodCfg == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_SamePeriodCfg)
		}

		if len(query.RequestMetrics.SamePeriodCfg.Method) == 0 {
			// 不报错，给默认值，都计算
			query.RequestMetrics.SamePeriodCfg.Method = []string{
				interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE,
				interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_RATE,
			}
			// return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_SamePeriodMethod)
		}

		for _, method := range query.RequestMetrics.SamePeriodCfg.Method {
			if method != interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE &&
				method != interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_RATE {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_SamePeriodMethod)
			}
		}

		if query.RequestMetrics.SamePeriodCfg.Offset <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_SamePeriodOffset)
		}

		if !interfaces.IsValidTimeGranularity(query.RequestMetrics.SamePeriodCfg.TimeGranularity) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_SamePeriodTimeGranularity)
		}

	}

	return nil
}

// 校验 dsl 是否满足规则
func validateDSL(ctx context.Context, query *interfaces.MetricModelQuery) error {
	// 解析 dsl 的 aggregations 部分，做 dsl 相关的校验
	// 1. 如果 interval 是变量，那么先把变量替换成具体值
	stepStr := ""
	if query.StepStr != nil {
		stepStr = *query.StepStr
	}
	dslStr := strings.Replace(query.Formula, interfaces.VARIABLE_INTERVAL, stepStr, -1)

	// 2. 把 dsl 语句转成 json，结构为 map[string]interfaces{}
	var dsl map[string]interface{}
	err := sonic.Unmarshal([]byte(dslStr), &dsl)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			WithErrorDetails(fmt.Sprintf("dsl Unmarshal error: %s", err.Error()))
	}

	//  dsl 的 size 置0
	if dsl["size"] == nil || dsl["size"] != float64(0) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			WithErrorDetails(fmt.Sprintf("The size of dsl expected 0, actual is %v", dsl["size"]))
	}

	// 3. 递归读取 json
	aggs, exist := dsl["aggs"]
	if !exist {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("dsl missing aggregation.")
	}
	// 4. 校验aggs
	aggsLayers := 0
	aggNameMap := make(map[string]int, 0)
	dateHistogramAggName := make([]string, 0)
	metricAggName := make([]string, 0)
	err = validAggs(ctx, aggs, query, &aggsLayers, aggNameMap, &dateHistogramAggName, &metricAggName)
	if err != nil {
		return err
	}

	// 5. 最大聚合层数7，6层分桶+1层值
	if aggsLayers > 7 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("The number of aggregation layers in DSL does not exceed 7.")
	}

	// 6. 度量字段
	if len(metricAggName) != 1 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("A metric aggregation should be included in dsl")
	}
	if query.MeasureField == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_MeasureField)
	} else if query.MeasureField != metricAggName[0] {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_MeasureField).
			WithErrorDetails("Measure Field should be the name of metric aggregation in dsl")
	}

	// 7. 时间字段
	if len(dateHistogramAggName) != 1 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("Only one date_histogram aggregation should be included in dsl")
	}

	// 9. date_histogram 下是值聚合
	if aggNameMap[dateHistogramAggName[0]]+1 != aggNameMap[metricAggName[0]] {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("The sub aggregation of date_histogram aggregation needs to be a metric aggregation.")
	}

	return nil
}

// 校验指标模型的查询参数
func validateMetricModelQueryParameters(ctx context.Context, offset, limit,
	ignoringMemoryCache, ignoringStoreCache, ignoringHCTSStr, fillNullStr,
	includeModel string) (interfaces.MetricModelQueryParameters, error) {

	queryParams := interfaces.MetricModelQueryParameters{}

	incModel, err := strconv.ParseBool(includeModel)
	if err != nil {
		return queryParams, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_IncludeModel).
			WithErrorDetails(fmt.Sprintf("The include_model:%s is invalid", includeModel))
	}

	off, err := strconv.ParseInt(offset, 10, 64)
	if err != nil {
		return queryParams, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Offset).
			WithErrorDetails(err.Error())
	}
	if off < interfaces.MIN_OFFSET {
		return queryParams, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Offset).
			WithErrorDetails(fmt.Sprintf("The offset is not greater than %d", interfaces.MIN_OFFSET))
	}

	lim, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		return queryParams, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Limit).
			WithErrorDetails(err.Error())
	}
	if !(limit == interfaces.DEFAULT_SERIES_LIMIT || (lim >= interfaces.MIN_LIMIT && lim <= interfaces.MAX_LIMIT)) {
		return queryParams, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Limit).
			WithErrorDetails(fmt.Sprintf("The number per page does not equal %s is not in the range of [%d,%d]",
				interfaces.DEFAULT_SERIES_LIMIT, interfaces.MIN_LIMIT, interfaces.MAX_LIMIT))
	}

	ignoringStore, err := strconv.ParseBool(ignoringStoreCache)
	if err != nil {
		return queryParams, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_IgnoringStoreCache).
			WithErrorDetails(fmt.Sprintf("The ignoring_store_cache:%s is invalid", ignoringStoreCache))
	}

	ignoringMemory, err := strconv.ParseBool(ignoringMemoryCache)
	if err != nil {
		return queryParams, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_IgnoringMemoryCache).
			WithErrorDetails(fmt.Sprintf("The ignoring_memory_cache:%s is invalid", ignoringMemoryCache))
	}

	ignoringHCTS, err := strconv.ParseBool(ignoringHCTSStr)
	if err != nil {
		return queryParams, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_IgnoringHCTS).
			WithErrorDetails(fmt.Sprintf("The ignoring_hcts:%s is invalid", ignoringHCTSStr))
	}

	fillNull, err := strconv.ParseBool(fillNullStr)
	if err != nil {
		return queryParams, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_FillNull).
			WithErrorDetails(fmt.Sprintf("The fill_null:%s is invalid", fillNullStr))
	}

	return interfaces.MetricModelQueryParameters{
		Offset:              off,
		Limit:               lim,
		IncludeModel:        incModel,
		IgnoringHCTS:        ignoringHCTS,
		IgnoringMemoryCache: ignoringMemory,
		IgnoringStoreCache:  ignoringStore,
		FillNull:            fillNull,
	}, nil
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
func validAggs(ctx context.Context, aggs interface{}, query *interfaces.MetricModelQuery, aggsLayers *int, aggNameMap map[string]int,
	dateHistogramAggName *[]string, metricAggName *[]string) error {

	aggsDetail, ok := aggs.(map[string]interface{})
	if !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("The aggregation of dsl is not a map")
	}
	if len(aggsDetail) != 1 {
		// 并行聚合不支持
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("Multiple aggregation is not supported")
	}
	// aggsDetail 的 key 是聚合名称，value 是个定义了聚合的 map
	for aggName, value := range aggsDetail {
		// key 是聚合名称
		// 各层聚合名称不能相同
		if _, exists := aggNameMap[aggName]; exists {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
				WithErrorDetails("The aggregation names of each layer aggregation cannot be the same")
		}
		*aggsLayers++
		aggNameMap[aggName] = *aggsLayers

		// value 是个map，map 中有两个元素，一个 key 是 aggType，一个 key 是 aggs 或者 aggregations
		aggValues, ok := value.(map[string]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
				WithErrorDetails("The aggregation of dsl is not a map")
		}
		for aggType, aggValue := range aggValues {
			switch aggType {
			case interfaces.AGGS:
				// 递归遍历子聚合
				err := validAggs(ctx, aggValue, query, aggsLayers, aggNameMap, dateHistogramAggName, metricAggName)
				if err != nil {
					return err
				}
			case interfaces.BUCKET_TYPE_TERMS, interfaces.BUCKET_TYPE_FILTERS,
				interfaces.BUCKET_TYPE_RANGE, interfaces.BUCKET_TYPE_DATE_RANGE:
				// 分桶聚合类型： date_histogram（日期直方图）、terms（词条）、filters（过滤）、range（范围）、date_range（日期范围）、multi_terms（多词条）
				// do nothing
			case interfaces.BUCKET_TYPE_DATE_HISTOGRAM:
				// 日期直方图的聚合名称需是时间字段的名称，此处解析就把名称留下
				dateHisto := aggValue.(map[string]interface{})

				if _, intervalExists := dateHisto["interval"]; intervalExists {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
						WithErrorDetails("The interval has been abandoned")
				}

				_, fixedIntervalExists := dateHisto["fixed_interval"]
				_, calendarIntervalExists := dateHisto["calendar_interval"]
				if fixedIntervalExists == calendarIntervalExists {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
						WithErrorDetails("The date_histogram aggregation statement is incorrect")
				}
				if calendarIntervalExists {
					query.IsCalendarInterval = 1
				}

				*dateHistogramAggName = append(*dateHistogramAggName, aggName)
			case interfaces.AGGR_TYPE_VALUE_COUNT, interfaces.AGGR_TYPE_CARDINALITY, interfaces.AGGR_TYPE_SUM, interfaces.AGGR_TYPE_AVG, interfaces.AGGR_TYPE_MAX, interfaces.AGGR_TYPE_MIN:
				// 值聚合, 需要把值聚合的聚合名称留下，需与配置的度量字段相同
				*metricAggName = append(*metricAggName, aggName)
			case interfaces.AGGR_TYPE_TOP_HITS:
				// 判断度量字段是不是在includes字段里。
				query.ContainTopHits = true
				err := parseTopHits(ctx, aggValue, query.MeasureField, aggsLayers, aggNameMap, metricAggName)
				if err != nil {
					return err
				}
			default:
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
					WithErrorDetails("Unsupport aggregation type in dsl.")
			}
		}
	}
	return nil
}

// 解析 top_hits
func parseTopHits(ctx context.Context, aggValue interface{}, measureField string, aggsLayers *int,
	aggNameMap map[string]int, metricAggName *[]string) error {

	var topHit interfaces.TopHits
	topAgg, err := sonic.Marshal(aggValue)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			WithErrorDetails(fmt.Sprintf("TopHits marshal error: %s", err.Error()))
	}
	err = sonic.Unmarshal(topAgg, &topHit)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			WithErrorDetails(fmt.Sprintf("TopHits Unmarshal error: %s", err.Error()))
	}

	if topHit.Size <= 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("TopHits's numHits must be > 0")
	}

	if topHit.Source.Includes == nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("TopHits's includes must not be empty")
	}

	equal := false
	for _, field := range topHit.Source.Includes {
		if measureField == field {
			equal = true
			break
		}
	}
	if equal {
		*metricAggName = append(*metricAggName, measureField)
		aggNameMap[measureField] = *aggsLayers
	} else {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_Formula).
			WithErrorDetails("The measure field must be one of the includes in top_hits.")
	}
	return nil
}

// 视图数据预览参数校验
func ValidateDataViewSimulate(ctx context.Context, query *interfaces.DataViewSimulateQuery) error {
	// 校验format是否为 original 或者 flat
	err := validateFormat(ctx, query.Format)
	if err != nil {
		return err
	}

	err = validatePaginationParams(ctx, query.ViewQueryCommonParams.Offset, query.ViewQueryCommonParams.Limit)
	if err != nil {
		return err
	}

	err = validateSortParamsV2(ctx, query.Sort)
	if err != nil {
		return err
	}

	// err = validateViewDataSource(ctx, query.DataSource, query.FieldScope)
	// if err != nil {
	// 	return err
	// }

	// err = validateViewFields(ctx, query.FieldScope, query.Fields)
	// if err != nil {
	// 	return err
	// }

	err = validateViewTime(ctx, query.ViewQueryCommonParams.Start, query.ViewQueryCommonParams.End)
	if err != nil {
		return err
	}

	// 校验行列规则的过滤条件
	for _, rule := range query.GetRowColumnRules() {
		err = validateCond(ctx, rule.RowFilters)
		if err != nil {
			return err
		}
	}

	return nil
}

// 视图数据查询参数校验 V2
func ValidateDataViewQueryV2(ctx context.Context, query *interfaces.DataViewQueryV2) error {
	// 校验format是否为 original 或者 flat
	err := validateFormat(ctx, query.Format)
	if err != nil {
		return err
	}

	// 校验分页参数
	err = validatePaginationParams(ctx, query.ViewQueryCommonParams.Offset, query.ViewQueryCommonParams.Limit)
	if err != nil {
		return err
	}

	// 校验排序参数
	err = validateSortParamsV2(ctx, query.Sort)
	if err != nil {
		return err
	}

	// 校验视图查询的 start 和 end
	err = validateViewTime(ctx, query.ViewQueryCommonParams.Start, query.ViewQueryCommonParams.End)
	if err != nil {
		return err
	}

	// 校验 search_after 和 pit 参数
	err = validateSearchAfterAndPit(ctx, query.SearchAfterParams, query.ViewQueryCommonParams.Offset)
	if err != nil {
		return err
	}

	// 过滤条件用map接，然后再decode到condCfg中
	var actualCond *cond.CondCfg
	err = mapstructure.Decode(query.GlobalFilters, &actualCond)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Filter).
			WithErrorDetails(fmt.Sprintf("mapstructure decode filters failed: %s", err.Error()))
	}
	query.ActualCondition = actualCond

	// 校验全局过滤条件：操作符、字段类型和操作符是否匹配
	err = validateCond(ctx, query.ActualCondition)
	if err != nil {
		return err
	}

	return nil
}

// 视图数据查询参数校验
func ValidateDataViewQueryV1(ctx context.Context, query *interfaces.DataViewQueryV1) error {
	// 校验format是否为 original 或者 flat
	err := validateFormat(ctx, query.Format)
	if err != nil {
		return err
	}

	// 有 scrollId，只需校验 scroll参数，无需其他参数
	if query.ScrollId != "" {
		if query.Scroll == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_NullParameter_Scroll).
				WithErrorDetails("When you have the 'scroll_id' param, 'scroll' param cannot be empty")
		}

	} else {
		// 校验分页参数
		err = validatePaginationParams(ctx, query.ViewQueryCommonParams.Offset, query.ViewQueryCommonParams.Limit)
		if err != nil {
			return err
		}

		// 校验排序方向
		err = validateSortParamsV1(ctx, query.Direction)
		if err != nil {
			return err
		}

		// 校验视图查询的 start 和 end
		err = validateViewTime(ctx, query.ViewQueryCommonParams.Start, query.ViewQueryCommonParams.End)
		if err != nil {
			return err
		}

		// 校验 scroll
		err = validateScroll(ctx, query.Scroll, query.ViewQueryCommonParams.Offset)
		if err != nil {
			return err
		}

		// 校验全局过滤条件：操作符、字段类型和操作符是否匹配
		err = validateCond(ctx, query.GlobalFilters)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateFormat(ctx context.Context, format string) error {
	if format != interfaces.Format_Original && format != interfaces.Format_Flat {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_Format).
			WithErrorDetails(fmt.Sprintf("The output format should be %s or %s", interfaces.Format_Original, interfaces.Format_Flat))
	}

	return nil
}

// 校验 search_after 和 pit 参数
func validateSearchAfterAndPit(ctx context.Context, params interfaces.SearchAfterParams, offset int) error {
	// 如果传递了 search_after，则 offset 必须为 0
	if len(params.SearchAfter) > 0 {
		if offset != 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Offset).
				WithErrorDetails("'offset' param must be set to 0 when 'search_after' is used")
		}
	}

	// keep_alive 不能超过 24h
	if params.PitKeepAlive != "" {
		keepAlive, err := convert.IntToDuration(params.PitKeepAlive)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				uerrors.Uniquery_DataView_InvalidParameter_PitKeepAlive).WithErrorDetails(err.Error())
		}

		if keepAlive > 24*time.Hour {
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				uerrors.Uniquery_DataView_InvalidParameter_PitKeepAlive).WithErrorDetails("The keep_alive parameter cannot exceed 24 hours")
		}

	}
	return nil
}

func validateScroll(ctx context.Context, scroll string, offset int) error {
	// scroll 查询
	if scroll != "" {
		if offset > 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_OffsetNotAllowedWithScroll).
				WithErrorDetails("Using 'offset' is not allowed in a scroll context")
		}
	}

	return nil
}

// 分页排序参数校验
func validatePaginationParams(ctx context.Context, offset, limit int) error {
	// from + size 查询校验
	if offset < 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Offset).
			WithErrorDetails("When execute From + size query, 'offset' should be >= 0")
	}

	if limit < interfaces.MIN_LIMIT || limit > interfaces.MAX_SEARCH_SIZE {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Limit).
			WithErrorDetails(fmt.Sprintf("Limit should be in the range of [%d,%d]", interfaces.MIN_LIMIT, interfaces.MAX_SEARCH_SIZE))
	}

	if offset+limit > interfaces.MAX_SEARCH_SIZE {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Limit)
	}

	return nil
}

func validateSortParamsV1(ctx context.Context, direction string) error {
	if direction != interfaces.ASC_DIRECTION && direction != interfaces.DESC_DIRECTION {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_Direction).
			WithErrorDetails("The sort direction should be desc or asc")
	}

	return nil
}

func validateSortParamsV2(ctx context.Context, sort []*interfaces.SortParamsV2) error {
	for _, s := range sort {
		if s.Direction != interfaces.ASC_DIRECTION && s.Direction != interfaces.DESC_DIRECTION {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_Direction).
				WithErrorDetails("The sort direction should be desc or asc")
		}
	}

	return nil
}

// 校验视图数据源
func validateViewDataSource(ctx context.Context, dataSource map[string]any, fieldScope string) error {
	var dataSourceType string
	// 视图数据源结构、非空校验
	if value, ok := dataSource["type"]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataSource).
			WithErrorDetails("The dataSource type is null")
	} else {
		if dataSourceType, ok = value.(string); !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataSource).
				WithErrorDetails("The dataSource type is not string")
		}
	}

	// 数据源类型校验、索引库校验
	switch dataSourceType {
	case interfaces.INDEX_BASE:
		if value, ok := dataSource[interfaces.INDEX_BASE]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataSource).
				WithErrorDetails("There is no 'index_base' parameter in the dataSource")
		} else {

			if bases, ok := value.([]any); ok {
				if len(bases) <= 0 {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataSource).
						WithErrorDetails("The number of index base must be at least one")
				}

				// 如果选择全部字段，来源索引库只能选一个，以减少字段类型冲突
				if fieldScope == interfaces.FieldScope_All && len(bases) > 1 {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataSource).
						WithErrorDetails("When the field scope is 'all fields', there can be only one index base")
				}
			} else {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataSource).
					WithErrorDetails("The index base names are not a list")
			}
		}
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_UnsupportDataSourceType).
			WithErrorDetails("Only 'index_base' is supported currently")
	}

	return nil
}

// 校验视图的查询时间
func validateViewTime(ctx context.Context, start, end int64) error {
	if start < 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Start)
		return httpErr
	}

	if end < 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_End)
		return httpErr
	}

	startTime := time.UnixMilli(start)
	endTime := time.UnixMilli(end)
	currentTime := time.Now()

	// start 是未来时间就抛异常
	if startTime.After(currentTime) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Start).
			WithErrorDetails(fmt.Sprintf("Start is greater than current time, current time is %v.", currentTime.UnixMilli()))
	}

	if end != 0 && endTime.Before(startTime) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Start).
			WithErrorDetails("end timestamp must not be before start time")

	}

	return nil
}

func validateCond(ctx context.Context, cfg *cond.CondCfg) error {
	if cfg == nil {
		return nil
	}

	// 判断过滤器是否为空对象 {}
	if cfg.Name == "" && cfg.Operation == "" && len(cfg.SubConds) == 0 && cfg.ValueFrom == "" && cfg.Value == nil {
		return nil
	}

	// 暂时性不支持过滤条件字段 __id 和 __routing，等索引库元字段完善后再放开
	// if cfg.Name == "__id" || cfg.Name == "__routing" {
	// 	return rest.NewHTTPError(ctx, http.StatusForbidden, uerrors.Uniquery_Forbidden_FilterField).
	// 		WithErrorDetails("The filter field '__id' and '__routing' is not allowed")
	// }

	// 过滤操作符
	if cfg.Operation == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_NullParameter_FilterOperation)
	}

	_, exists := cond.OperationMap[cfg.Operation]
	if !exists {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_UnsupportFilterOperation)
	}

	switch cfg.Operation {
	case cond.OperationAnd, cond.OperationOr:
		// 子过滤条件不能超过10个
		if len(cfg.SubConds) > cond.MaxSubCondition {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_CountExceeded_Filters).
				WithErrorDetails(fmt.Sprintf("The number of subConditions exceeds %d", cond.MaxSubCondition))
		}

		for _, subCond := range cfg.SubConds {
			err := validateCond(ctx, subCond)
			if err != nil {
				return err
			}
		}
	default:
		// 过滤字段名称不能为空
		if cfg.Operation != cond.OperationMultiMatch && cfg.Name == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_NullParameter_FilterName)
		}

		if _, ok := cond.NotRequiredValueOperationMap[cfg.Operation]; !ok {
			if cfg.ValueFrom == "" {
				cfg.ValueFrom = vopt.ValueFrom_Const
			}
			if cfg.ValueFrom != vopt.ValueFrom_Const {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_ValueFrom).
					WithErrorDetails(fmt.Sprintf("condition does not support value_from type('%s')", cfg.ValueFrom))
			}
		}
	}

	switch cfg.Operation {
	case cond.OperationEq, cond.OperationNotEq, cond.OperationGt, cond.OperationGte, cond.OperationLt, cond.OperationLte,
		cond.OperationLike, cond.OperationNotLike, cond.OperationPrefix, cond.OperationNotPrefix, cond.OperationRegex,
		cond.OperationMatch, cond.OperationMatchPhrase, cond.OperationCurrent, cond.OperationMultiMatch:
		// 右侧值为单个值
		_, ok := cfg.Value.([]interface{})
		if ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a single value", cfg.Operation))
		}

		if cfg.Operation == cond.OperationLike || cfg.Operation == cond.OperationNotLike ||
			cfg.Operation == cond.OperationPrefix || cfg.Operation == cond.OperationNotPrefix {
			// 如果有 real_value 则跳过 value 的校验
			if cfg.RealValue == nil {
				_, ok := cfg.Value.(string)
				if !ok {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
						WithErrorDetails("[like not_like prefix not_prefix] operation's value should be a string")
				}
			} else {
				_, ok := cfg.RealValue.(string)
				if !ok {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
						WithErrorDetails("[like not_like prefix not_prefix] operation's real_value should be a string")
				}
			}
		}

		if cfg.Operation == cond.OperationRegex {
			val, ok := cfg.Value.(string)
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
					WithErrorDetails("[regex] operation's value should be a string")
			}

			_, err := regexp2.Compile(val, regexp2.RE2)
			if err != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
					WithErrorDetails(fmt.Sprintf("[regex] operation regular expression error: %s", err.Error()))
			}

		}

	case cond.OperationIn, cond.OperationNotIn:
		// 当 operation 是 in, not_in 时，value 为任意基本类型的数组，且长度大于等于1；
		_, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("[in not_in] operation's value must be an array")
		}

		if len(cfg.Value.([]interface{})) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("[in not_in] operation's value should contains at least 1 value")
		}
	case cond.OperationRange, cond.OperationOutRange, cond.OperationBefore, cond.OperationBetween:
		// 当 operation 是 range 时，value 是个由范围的下边界和上边界组成的长度为 2 的数值型数组
		// 当 operation 是 out_range 时，value 是个长度为 2 的数值类型的数组，查询的数据范围为 (-inf, value[0]) || [value[1], +inf)
		v, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("[range, out_range] operation's value must be an array")
		}

		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("[range, out_range] operation's value must contain 2 values")
		}

	}

	return nil
}

/*
基于链路模型预览/查询的校验函数. 包括:
	(1) Span列表查询参数的校验
	(2) Span关联日志列表预览参数的校验
*/

// 基于链路模型查询的校验函数(1): : Span列表查询参数的校验
func validateParamsWhenGetSpanList(ctx context.Context, params interfaces.SpanListQueryParams) (interfaces.SpanListQueryParams, error) {
	// 1. 校验通用分页查询参数
	subParams, err := validatePaginationQueryParams(ctx, params.PaginationQueryParams, "span_list")
	if err != nil {
		return params, err
	}
	params.PaginationQueryParams = subParams

	// 2. 校验condition
	err = validateCond(ctx, params.Condition)
	if err != nil {
		return params, err
	}

	return params, nil
}

// 基于链路模型查询的校验函数(2): : Span关联日志列表预览参数的校验
func validateParamsWhenGetRelatedLogList(ctx context.Context, params interfaces.RelatedLogListQueryParams) (interfaces.RelatedLogListQueryParams, error) {
	// 1. 校验分页查询参数
	subParams, err := validatePaginationQueryParams(ctx, params.PaginationQueryParams, "related_log_list")
	if err != nil {
		return params, err
	}

	// 更新params
	params.PaginationQueryParams = subParams

	// 2. 校验condition
	err = validateCond(ctx, params.Condition)
	if err != nil {
		return params, err
	}

	return params, nil
}

// 目标模型数据预览参数校验
func ValidateObjectiveModelSimulate(ctx context.Context, query *interfaces.ObjectiveModelQuery) error {
	// 校验指标类型非空
	if query.ObjectiveType == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_NullParameter_ObjectiveType)
	}
	// 目标类型是枚举中的某个值
	if !interfaces.IsValidObjectiveType(query.ObjectiveType) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_UnsupportObjectiveType)
	}

	// 计算公式非空
	if query.ObjectiveConfig == nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_NullParameter_ObjectiveConfig)
	}

	// 查询语言是 promql，则时间维度必须是 @timestamp，时间格式必须是 epoch_millis
	if query.ObjectiveType == interfaces.SLO {
		// 把 objective_config转成 SLOObjective
		var sloObjective interfaces.SLOObjective
		jsonData, err := json.Marshal(query.ObjectiveConfig)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_InvalidParameter_ObjectiveConfig).
				WithErrorDetails(fmt.Sprintf("ObjectiveConfig Marshal error: %s", err.Error()))
		}
		err = json.Unmarshal(jsonData, &sloObjective)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_InvalidParameter_ObjectiveConfig).
				WithErrorDetails(fmt.Sprintf("SLO Objective Unmarshal error: %s", err.Error()))
		}

		// 目标非空，大于0
		if sloObjective.Objective == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_NullParameter_Objective)
		}
		if *(sloObjective.Objective) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_InvalidParameter_Objective)
		}

		// 周期非空
		if sloObjective.Period == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_NullParameter_Period)
		}

		// 良好指标非空
		if sloObjective.GoodMetricModel == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_NullParameter_GoodMetricModel)
		}
		if sloObjective.GoodMetricModel.Id == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_NullParameter_GoodMetricModelID)
		}

		// 总指标非空
		if sloObjective.TotalMetricModel == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_NullParameter_TotalMetricModel)
		}
		if sloObjective.TotalMetricModel.Id == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_NullParameter_TotalMetricModelID)
		}

		// 状态阈值状态校验
		if sloObjective.StatusConfig != nil {
			// 状态非空, 区间不能重叠
			err = validStatusConfig(ctx, sloObjective.StatusConfig.Ranges)
			if err != nil {
				return err
			}
		}
		query.ObjectiveConfig = sloObjective
	} else {
		// kpi
		// 把 objective_config转成 KPIObjective
		var kpiObjective interfaces.KPIObjective
		jsonData, err := json.Marshal(query.ObjectiveConfig)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_InvalidParameter_ObjectiveConfig).
				WithErrorDetails(fmt.Sprintf("ObjectiveConfig Marshal error: %s", err.Error()))
		}
		err = json.Unmarshal(jsonData, &kpiObjective)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_InvalidParameter_ObjectiveConfig).
				WithErrorDetails(fmt.Sprintf("KPI Objective Unmarshal error: %s", err.Error()))
		}

		// 目标非空，大于0
		if kpiObjective.Objective == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_NullParameter_Objective)
		}
		if *(kpiObjective.Objective) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_InvalidParameter_Objective)
		}

		// 目标单位非空
		if kpiObjective.Unit == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_NullParameter_ObjectiveUnit)
		}

		// 目标单位无效
		if kpiObjective.Unit != interfaces.UNIT_NUM_NONE && kpiObjective.Unit != interfaces.UNIT_NUM_PERCENT {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_InvalidParameter_ObjectiveUnit).
				WithErrorDetails(fmt.Sprintf(`expected objective unit is one of [none, %%], actual unit is %s`, kpiObjective.Unit))
		}

		// 综合计算指标非空
		if kpiObjective.ComprehensiveMetricModels == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_NullParameter_ComprehensiveMetricModels)
		}

		// 综合计算指标个数大于10
		if len(kpiObjective.ComprehensiveMetricModels) > interfaces.COMPREHENSIVE_METRIC_TOTAL {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_CountExceeded_ComprehensiveMetricModelsTotal)
		}

		var weight int64
		for _, mm := range kpiObjective.ComprehensiveMetricModels {
			// id非空
			if mm.Id == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_NullParameter_ComprehensiveMetricModelID)
			}
			// 权重非空且大于0
			if mm.Weight == nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_NullParameter_ComprehensiveWeight)
			}
			if *(mm.Weight) <= 0 {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_InvalidParameter_ComprehensiveWeight)
			}
			weight += *(mm.Weight)
		}
		// 权重之和等于100
		if weight != 100 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				uerrors.Uniquery_ObjectiveModel_InvalidParameter_ComprehensiveWeight).
				WithErrorDetails(fmt.Sprintf("The sum of weights must equal 100, actual is %d", weight))
		}

		// 附加计算指标
		if len(kpiObjective.AdditionalMetricModels) > interfaces.ADDITIONAL_METRIC_TOTAL {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_CountExceeded_ComprehensiveMetricModelsTotal)
		}

		for _, mm := range kpiObjective.AdditionalMetricModels {
			// id非空
			if mm.Id == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_NullParameter_AdditionalMetricModelID)
			}
		}
		// 状态阈值状态校验
		if kpiObjective.StatusConfig != nil {
			// 状态非空, 区间不能重叠
			err = validStatusConfig(ctx, kpiObjective.StatusConfig.Ranges)
			if err != nil {
				return err
			}
		}
		query.ObjectiveConfig = kpiObjective
	}

	// 校验查询时间范围的相关参数
	err := validateQueryTimeParam(ctx, &query.QueryTimeParams)
	if err != nil {
		return err
	}

	return nil
}

// 基于目标模型的指标数据查询的参数校验
func validateObjectiveModelData(ctx context.Context, query *interfaces.ObjectiveModelQuery) error {

	// 校验查询时间范围的相关参数
	err := validateQueryTimeParam(ctx, &query.QueryTimeParams)
	if err != nil {
		return err
	}

	return nil
}

// 校验目标模型的查询参数
func validateObjectiveModelQueryParameters(ctx context.Context, ignoringMemoryCache, ignoringStoreCache,
	includeModel string, includeMetrics []string) (interfaces.ObjectiveModelQueryParameters, error) {

	queryParams := interfaces.ObjectiveModelQueryParameters{}

	incModel, err := strconv.ParseBool(includeModel)
	if err != nil {
		return queryParams, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_InvalidParameter_IncludeModel).
			WithErrorDetails(fmt.Sprintf("The include_model:%s is invalid", includeModel))
	}

	ignoringStore, err := strconv.ParseBool(ignoringStoreCache)
	if err != nil {
		return queryParams, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_InvalidParameter_IgnoringStoreCache).
			WithErrorDetails(fmt.Sprintf("The ignoring_store_cache:%s is invalid", ignoringStoreCache))
	}

	ignoringMemory, err := strconv.ParseBool(ignoringMemoryCache)
	if err != nil {
		return queryParams, rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_InvalidParameter_IgnoringMemoryCache).
			WithErrorDetails(fmt.Sprintf("The ignoring_memory_cache:%s is invalid", ignoringMemoryCache))
	}

	return interfaces.ObjectiveModelQueryParameters{
		MetricModelQueryParameters: interfaces.MetricModelQueryParameters{
			IncludeModel:        incModel,
			IgnoringMemoryCache: ignoringMemory,
			IgnoringStoreCache:  ignoringStore,
		},
		IncludeMetrics: includeMetrics,
	}, nil
}

func validStatusConfig(ctx context.Context, ranges []interfaces.Range) error {
	if len(ranges) > interfaces.STATUS_TOTAL {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_CountExceeded_StatusTotal)
	}
	rgs := make([]interfaces.Range, 0)
	rgs = append(rgs, ranges...)
	for i, rg := range rgs {
		if rg.Status == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_InvalidParameter_StatusRanges).
				WithErrorDetails("Exist empty status in the status config")
		}
		if rg.From == nil && rg.To == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_InvalidParameter_StatusRanges).
				WithErrorDetails("Exist both from and to are empty in the status config")
		} else if rg.From == nil {
			rgs[i].From = &interfaces.NEG_INF
		} else if rg.To == nil {
			rgs[i].To = &interfaces.POS_INF
		} else if *rg.From >= *rg.To {
			// 如果from比to大，报错
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_InvalidParameter_StatusRanges).
				WithErrorDetails("Range's From is greater than Range's To in the status config")
		}
	}

	// 区间不能重叠
	if convert.HasOverlappingRanges(rgs) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_ObjectiveModel_InvalidParameter_StatusRanges).
			WithErrorDetails("Exist overlapping range in the kpi's status config")
	}
	return nil
}

func validateHavingCondition(ctx context.Context, havingCondition *cond.CondCfg) error {
	if havingCondition.Name != interfaces.VALUE_FIELD {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_InvalidParameter_HavingConditionName)
	}
	if havingCondition.Operation == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_HavingConditionOperation)
	}
	// operation需是有效的
	_, exists := cond.HavingOperationMap[havingCondition.Operation]
	if !exists {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_UnsupportHavingConditionOperation).
			WithErrorDetails(fmt.Sprintf("unsupport having condition operation %s", havingCondition.Operation))
	}

	if havingCondition.Value == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_HavingConditionOperation)
	}

	switch havingCondition.Operation {
	case cond.OperationEq, cond.OperationNotEq, cond.OperationGt, cond.OperationGte,
		cond.OperationLt, cond.OperationLte:
		// 右侧值为单个值
		_, ok := havingCondition.Value.([]interface{})
		if ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a single value", havingCondition.Operation))
		}
		// 值为数值类型
		_, err := convert.AssertFloat64(havingCondition.Value)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a number in having condition", havingCondition.Operation))
		}
	case cond.OperationIn, cond.OperationNotIn:
		// 当 operation 是 in, not_in 时，value 为任意基本类型的数组，且长度大于等于1；
		_, ok := havingCondition.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("[in not_in] operation's value must be an array")
		}

		if len(havingCondition.Value.([]interface{})) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("[in not_in] operation's value should contains at least 1 value")
		}

		if !cond.IsSameType(havingCondition.Value.([]any)) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("condition [not_in] right value should be an array composed of elements of same type")
		}
		// 值为数值类型
		_, err := convert.AssertFloat64(havingCondition.Value.([]any)[0])
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a number slice in having condition", havingCondition.Operation))
		}
	case cond.OperationRange, cond.OperationOutRange:
		// 当 operation 是 range 时，value 是个由范围的下边界和上边界组成的长度为 2 的数值型数组
		// 当 operation 是 out_range 时，value 是个长度为 2 的数值类型的数组，查询的数据范围为 (-inf, value[0]) || [value[1], +inf)
		v, ok := havingCondition.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("[range, out_range] operation's value must be an array")
		}

		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("[range, out_range] operation's value must contain 2 values")
		}
		if !cond.IsSameType(havingCondition.Value.([]any)) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails("condition [range, out_range] right value should be an array composed of elements of same type")
		}
		// 值为数值类型
		_, err := convert.AssertFloat64(havingCondition.Value.([]any)[0])
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_FilterValue).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a number slice in having condition", havingCondition.Operation))
		}
	}
	return nil
}
