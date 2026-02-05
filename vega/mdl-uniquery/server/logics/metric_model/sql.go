package metric_model

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"

	"uniquery/common"
	cond "uniquery/common/condition"
	"uniquery/common/convert"
	vopt "uniquery/common/value_opt"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
)

// 根据配置信息生成sql
func generateSQL(ctx context.Context, query interfaces.MetricModelQuery, dataView *interfaces.DataView,
	viewQuery4Metric interfaces.ViewQuery4Metric, model interfaces.MetricModel) (string, error) {

	// sql := fmt.Sprintf("select %s,%s,%s from %s.%s where %s group by %s,%s limit 1", group, analysisDim, aggrExpr,
	// 	vegaViewFields.Catalog, vegaViewFields.Table, condStr, group, analysisDim)

	// fieldsMap := make(map[string]*cond.ViewField)
	sqlConfig := query.FormulaConfig.(interfaces.SQLConfig)

	condStr := sqlConfig.ConditionStr
	if sqlConfig.ConditionStr != "" && query.ConditionStr != "" {
		condStr = fmt.Sprintf("%s AND %s", sqlConfig.ConditionStr, query.ConditionStr)
	} else if query.ConditionStr != "" {
		condStr = query.ConditionStr
	}

	condtions := sqlConfig.Condition
	if sqlConfig.Condition != nil && query.Condition != nil {
		condtions = &cond.CondCfg{
			Operation: cond.OperationAnd,
			SubConds:  []*cond.CondCfg{sqlConfig.Condition, query.Condition},
		}
	} else if query.Condition != nil {
		condtions = query.Condition
	}
	// 1. 模型身上的过滤条件
	if condtions != nil {
		// 把 condition 转成where表达式的子句

		// 2. 构建condition
		// 创建一个包含查询类型的上下文
		ctx = context.WithValue(ctx, cond.CtxKey_QueryType, interfaces.QueryType_SQL)
		CondCfg, _, err := cond.NewCondition(ctx, condtions, dataView.Type, dataView.FieldsMap)
		if err != nil {
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_Filters).
				WithErrorDetails(fmt.Sprintf("New condition failed, %v", err))
		}

		// 3. 生成sql
		if CondCfg != nil {
			condStr, err = CondCfg.Convert2SQL(ctx)
			if err != nil {
				return "", rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_Filters).
					WithErrorDetails(fmt.Sprintf("Convert condition to dsl failed, %v", err))
			}
		}
	}

	// 2. 再拼上时间字段的过滤的条件, AND的关系
	// 需要把时间戳格式化后传递到过滤条件中
	// 过滤方式1: to_unixtime(l_commitdate) between 785088000 and 785174340. unixtime是到秒的时间戳
	// 过滤方式2: 把start和end格式化后传输:
	// l_commitdate between cast('1994-11-18 00:00:00.200'  AS TIMESTAMP) and cast('1994-11-18 23:59:59.999'  AS TIMESTAMP)
	// timeFilterSql := fmt.Sprintf(`to_unixtime(%s) BETWEEN %d AND %d`, query.DateField, query.Start/1000, query.End/1000)
	timeFilterSql := ""
	dateField := ""
	if query.DateField != "" {
		if field, exist := dataView.FieldsMap[query.DateField]; exist {
			dateField = field.OriginalName
			timeFilterSql = fmt.Sprintf(`"%s" BETWEEN from_unixtime(%d) AND from_unixtime(%d)`, dateField,
				*query.Start/1000, *query.End/1000)
		}
	}

	// 3. 查询时附加的过滤条件: filters的过滤条件
	reqFilterStr, err := transFilter2SQL(ctx, query, dataView)
	if err != nil {
		return "", err
	}

	// 4. 分组字段,先分组字段,分析维度在其后拼接
	// groupFs := sqlConfig.GroupByFields
	groupFs := []string{}
	groupMap := make(map[string]*cond.ViewField)
	for _, f := range sqlConfig.GroupByFields {
		if field, exist := dataView.FieldsMap[f]; exist {
			groupMap[f] = field
			groupFs = append(groupFs, fmt.Sprintf(`"%s"`, field.OriginalName))
		}
	}
	// 拼分析维度
	for _, dim := range query.AnalysisDims {
		if _, exist := groupMap[dim]; !exist {
			groupFs = append(groupFs, fmt.Sprintf(`"%s"`, dim))
		}
	}
	// 如果是范围查询,则需要组装一个时间字段按指定步长分组的group by
	dateFmt := ""
	if !query.IsInstantQuery {
		if query.DateField == "" {
			// 报错
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_MetricModel_NullParameter_DateField).
				WithErrorDetails("date field is required for range query")
		}
		// 组装时间字段的group by. 时间的group by字段放在最后 as __time
		// date_format(timestamp, format) -> varchar
		switch *query.StepStr {
		case interfaces.CALENDAR_STEP_MINUTE:
			dateFmt = fmt.Sprintf(`date_format("%s",'%s')`, dateField, `%Y-%m-%d %H:%i`)
		case interfaces.CALENDAR_STEP_HOUR:
			dateFmt = fmt.Sprintf(`date_format("%s",'%s')`, dateField, `%Y-%m-%d %H`)
		case interfaces.CALENDAR_STEP_DAY:
			dateFmt = fmt.Sprintf(`date_format("%s",'%s')`, dateField, `%Y-%m-%d`)
		case interfaces.CALENDAR_STEP_WEEK:
			dateFmt = fmt.Sprintf(`date_format("%s",'%s')`, dateField, `%x-%v`)
		case interfaces.CALENDAR_STEP_MONTH:
			dateFmt = fmt.Sprintf(`date_format("%s",'%s')`, dateField, `%Y-%m`)
		case interfaces.CALENDAR_STEP_QUARTER:
			dateFmt = `format('%d-Q%d',` + fmt.Sprintf(`year("%s"),quarter("%s"))`, dateField, dateField)
		case interfaces.CALENDAR_STEP_YEAR:
			dateFmt = fmt.Sprintf(`date_format("%s",'%s')`, dateField, `%Y`)
		}
		groupFs = append(groupFs, dateFmt)
	}
	group := strings.Join(groupFs, ",")

	// 5. 度量聚合
	aggrStr4Having := sqlConfig.AggrExprStr
	aggrExpr := sqlConfig.AggrExprStr
	if sqlConfig.AggrExpr != nil {
		// 注意去重计数的函数，即函数映射关系
		aggField := sqlConfig.AggrExpr.Field
		if field, exist := dataView.FieldsMap[aggField]; exist {
			aggField = field.OriginalName
		}
		if sqlConfig.AggrExpr.Aggr == interfaces.AGGR_TYPE_COUNT_DISTINCT {
			aggrStr4Having = fmt.Sprintf(`count(distinct "%s")`, aggField)
			aggrExpr = fmt.Sprintf(`count(distinct "%s") as %s`, aggField, interfaces.VALUE_FIELD)
		} else {
			aggrStr4Having = fmt.Sprintf(`%s("%s")`, sqlConfig.AggrExpr.Aggr, aggField)
			aggrExpr = fmt.Sprintf(`%s("%s") as %s`, sqlConfig.AggrExpr.Aggr, aggField, interfaces.VALUE_FIELD)
		}
	}
	if sqlConfig.AggrExprStr != "" {
		// as __value
		aggrExpr = fmt.Sprintf("%s as %s", sqlConfig.AggrExprStr, interfaces.VALUE_FIELD)
	}

	// 6. 排序部分下沉
	orderByMap := map[string]bool{}
	allOrderBys := []interfaces.OrderField{}
	// 指标模型配置的排序
	for _, orderBy := range model.OrderByFields {
		orderByMap[orderBy.Name] = true
		allOrderBys = append(allOrderBys, orderBy)
	}
	// 请求的排序
	for _, orderBy := range query.OrderByFields {
		// 已经配置在模型中的排序字段，忽略，以模型的为准
		if !orderByMap[orderBy.Name] {
			orderByMap[orderBy.Name] = true
			allOrderBys = append(allOrderBys, orderBy)
		}
	}

	orderByFs := []string{}
	for _, orderBy := range allOrderBys {
		// 排序字段的名称是视图名称，需换成技术名
		orderByName := orderBy.Name
		if field, exist := dataView.FieldsMap[orderByName]; exist {
			orderByName = field.OriginalName
		}
		orderByFs = append(orderByFs, fmt.Sprintf(`"%s" %s`, orderByName, orderBy.Direction))
	}
	// 还需要拼接上按时间分组字段排序
	if dateFmt != "" {
		orderByFs = append(orderByFs, dateFmt)
	}
	orderBy := strings.Join(orderByFs, ",")

	orderByStr := ""
	if orderBy != "" {
		orderByStr = fmt.Sprintf("ORDER BY %s", orderBy)
	}

	// 7. having 过滤。不带同环比的可以把 having 拼上。范围查询，不拼having，跳过,即时查询才可以把 having 拼上
	havingFs := []string{}
	if query.RequestMetrics == nil && query.IsInstantQuery {
		havingFs = generateHavingStr(query, model, aggrStr4Having)
	}
	having := strings.Join(havingFs, " AND ")
	havingStr := ""
	if having != "" {
		havingStr = fmt.Sprintf("HAVING %s", having)
	}
	// 拼sql
	// 1. select 部分
	sql := fmt.Sprintf("SELECT %s", aggrExpr)
	groupByStr := ""
	if group != "" {
		sql = fmt.Sprintf("%s, %s", sql, group)
		groupByStr = fmt.Sprintf("GROUP BY %s", group)
		if dateFmt != "" {
			sql = fmt.Sprintf("%s as %s", sql, interfaces.TIME_FIELD)
		}
	}

	// 2. from 的是视图的 view_source_catalog_name.technical_name
	sql = fmt.Sprintf("%s FROM (%s) WHERE 1=1", sql, strings.TrimSuffix(viewQuery4Metric.QueryStr, ";"))
	// 3. where 过滤条件
	if timeFilterSql != "" {
		sql = fmt.Sprintf("%s AND %s", sql, timeFilterSql)
	}
	if condStr != "" {
		sql = fmt.Sprintf("%s AND %s", sql, condStr)
	}
	if reqFilterStr != "" {
		sql = fmt.Sprintf("%s AND %s", sql, reqFilterStr)
	}
	// 4. group by 部分
	sql = fmt.Sprintf("%s %s %s %s", sql, groupByStr, havingStr, orderByStr)

	// 计算公式有效性检查时,拼上 limit 1
	if query.IsModelRequest {
		sql = fmt.Sprintf("%s LIMIT 1", sql)
	}
	logger.Debugf("生成的sql语句为: %s", sql)
	return sql, nil
}

// 生成 having 子句
func generateHavingStr(query interfaces.MetricModelQuery, metricModel interfaces.MetricModel, aggrStr4Having string) []string {

	havingFs := []string{}
	// 模型配置的 having 和查询请求的 having 做 and 的合并
	havingConditions := []*cond.CondCfg{query.HavingCondition, metricModel.HavingCondition}
	for _, condition := range havingConditions {
		if condition == nil {
			continue
		}
		switch condition.Operation {
		case cond.OperationEq:
			havingFs = append(havingFs, fmt.Sprintf(`%s %s %v`, aggrStr4Having, "=", condition.Value))
		case cond.OperationGt, cond.OperationGte, cond.OperationLt, cond.OperationLte, cond.OperationNotEq:
			havingFs = append(havingFs, fmt.Sprintf(`%s %s %v`, aggrStr4Having, condition.Operation, condition.Value))
		case cond.OperationIn, cond.OperationNotIn:
			values := condition.Value.([]any)

			valueList := make([]string, len(values))
			for i, v := range values {
				vStr, ok := v.(string)
				if ok {
					valueList[i] = fmt.Sprintf(`'%v'`, vStr)
				} else {
					valueList[i] = fmt.Sprintf(`%v`, v)
				}
			}
			value := strings.Join(valueList, ",")

			op := "IN"
			if condition.Operation == cond.OperationNotIn {
				op = "NOT IN"
			}

			havingFs = append(havingFs, fmt.Sprintf(`%s %s (%v)`, aggrStr4Having, op, value))

		case cond.OperationRange, cond.OperationOutRange:
			values := condition.Value.([]any)

			left := values[0]
			right := values[1]

			// 处理字符串类型的值，需要用单引号包裹
			leftStr, ok := left.(string)
			if ok {
				leftStr = interfaces.Special.Replace(fmt.Sprintf("%q", leftStr))
			} else {
				leftStr = fmt.Sprintf("%v", left)
			}

			rightStr, ok := right.(string)
			if ok {
				rightStr = interfaces.Special.Replace(fmt.Sprintf("%q", rightStr))
			} else {
				rightStr = fmt.Sprintf("%v", right)
			}

			if condition.Operation == cond.OperationRange {
				// 构建SQL条件：字段名 >= 左边界 AND 字段名 < 右边界
				havingFs = append(havingFs, fmt.Sprintf(`%s >= %s AND %s < %s `, aggrStr4Having, leftStr, aggrStr4Having, rightStr))
			} else {
				// 构建SQL条件：字段名 < 左边界 OR 字段名 >= 右边界
				havingFs = append(havingFs, fmt.Sprintf(`%s < %s OR %s >= %s `, aggrStr4Having, leftStr, aggrStr4Having, rightStr))
			}
		default:
			return havingFs
		}

	}
	return havingFs
}

// 把请求的过滤条件转换成sql的过滤条件
func transFilter2SQL(ctx context.Context, query interfaces.MetricModelQuery, dataView *interfaces.DataView) (string, error) {
	filters := query.Filters
	if len(filters) <= 0 {
		return "", nil
	}
	var whereArgs []string
	// 过滤器间是 and，过滤器内是 or
	for _, filter := range filters {
		_, ok := dataView.FieldsMap[filter.Name]
		if !ok {
			// 字段不存在
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_Filters).
				WithErrorDetails(fmt.Sprintf("filter field[%s] not exists in data view[%s]", filter.Name, query.DataSource.ID))
		}

		var condtioni cond.Condition
		var err error
		operation := filter.Operation
		switch filter.Operation {
		// 对于filter和condition定义不同的部分,需要做映射
		case interfaces.OPERATION_EQ:
			operation = cond.OperationEq
		case cond.OperationRange:
			// do nothing. sql 暂不支持此操作符
			operation = ""
		case cond.OperationOutRange:
			// do nothing. sql 暂不支持此操作符
			operation = ""
		}

		if operation != "" {
			// 创建一个condition
			// 创建一个包含查询类型的上下文
			ctx = context.WithValue(ctx, cond.CtxKey_QueryType, interfaces.QueryType_SQL)
			condtioni, _, err = cond.NewCondition(ctx, &cond.CondCfg{
				Name:      filter.Name,
				Operation: operation,
				ValueOptCfg: vopt.ValueOptCfg{
					ValueFrom: vopt.ValueFrom_Const,
					Value:     filter.Value,
				},
			}, dataView.Type, dataView.FieldsMap)
			if err != nil {
				return "", rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_Filters).
					WithErrorDetails(err.Error())
			}
			// 转换到sql
			inSql, err := condtioni.Convert2SQL(ctx)
			if err != nil {
				return "", rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_Filters).
					WithErrorDetails(err.Error())
			}
			whereArgs = append(whereArgs, inSql)
		}
	}
	whereSql := strings.Join(whereArgs, " AND ")
	return whereSql, nil
}

// 解析vega的结果为统一的指标结果格式
func parseVegaResult2Uniresponse(ctx context.Context, vegaData, samePeriodDatas interfaces.VegaFetchData,
	query interfaces.MetricModelQuery, vegaDuration int64, model interfaces.MetricModel) (interfaces.MetricModelUniResponse, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Parse vega result to uniresponse")
	defer span.End()

	if !query.IsInstantQuery {
		// 范围查询的时间字段一定是 __time, 值字段一定是 __value,其余字段是维度字段,且要按维度字段组成序列
		return convertVegaDatas2TimeSeries(ctx, vegaData, samePeriodDatas, query, vegaDuration)
	}

	var err error
	resp := interfaces.MetricModelUniResponse{
		SeriesTotal:    len(vegaData.Data),
		VegaDurationMs: vegaDuration,
	}

	// 如果要计算占比, 计算总和
	var total float64
	hasGrowthValue := false
	hasGrowthRate := false
	samePeriodMap := make(map[string]float64)
	if query.RequestMetrics != nil {
		switch query.RequestMetrics.Type {
		case interfaces.METRICS_PROPORTION:
			// 查找__value列的索引
			total, err = getProportionTotal(ctx, vegaData)
			if err != nil {
				return resp, err
			}
		case interfaces.METRICS_SAMEPERIOD:
			// 检查method数组，确定需要计算哪些指标
			methods := query.RequestMetrics.SamePeriodCfg.Method
			for _, method := range methods {
				if method == interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE {
					hasGrowthValue = true
				}
				if method == interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_RATE {
					hasGrowthRate = true
				}
			}

			// 创建同期数据的查找映射
			for _, row := range samePeriodDatas.Data {
				key := ""
				var value float64
				// 构建维度键和指标值
				for i, field := range vegaData.Columns {
					if field.Name == interfaces.VALUE_FIELD {
						value, err = getFloat64Value(ctx, row[i])
						if err != nil {
							return resp, err
						}
						continue
					}
					key += fmt.Sprintf("%v|", row[i])
				}
				samePeriodMap[key] = value
			}
		}
	}

	datas := make([]interfaces.MetricModelData, 0)
	// 处理本期数据
	for _, row := range vegaData.Data {
		// 第i行数据
		// 遍历字段,把第一行组装成labels和values
		labels := make(map[string]string)
		values := make([]any, 0)
		growthValues := make([]any, 0)
		growthRates := make([]any, 0)
		proportions := make([]any, 0)
		// 构建维度键
		key := ""
		var currentValue float64
		// 把每行处理成时序的结构
		for i, col := range vegaData.Columns {
			// 聚合字段用 __value 标识,group by字段自行定义
			if col.Name == interfaces.VALUE_FIELD {
				currentValue, err = appendValues(ctx, row[i], &values)
				if err != nil {
					return resp, err
				}
				continue
			}
			key += fmt.Sprintf("%v|", row[i])
			// 维度字段
			labels[col.Name] = fmt.Sprintf("%v", row[i])
		}

		// 计算增长值\增长率
		if query.RequestMetrics != nil {
			switch query.RequestMetrics.Type {
			case interfaces.METRICS_SAMEPERIOD:

				if samePeriod, exists := samePeriodMap[key]; exists {
					// 同期数据存在
					if hasGrowthValue {
						growthValues = append(growthValues, currentValue-samePeriod)
					}
					// 计算增长率（避免除以0）
					if hasGrowthRate {
						if samePeriod != 0 {
							growthRates = append(growthRates, (currentValue-samePeriod)/samePeriod*100)
							// 保留两位小数
							// growthRate = math.Round(growthRate.(float64)*100) / 100
						} else {
							growthRates = append(growthRates, nil)
						}
					}
				} else {
					// 请求了同环比,但是同期值不存在,只有method中包含对应方法才添加null
					if hasGrowthValue {
						growthValues = append(growthValues, nil)
					}
					if hasGrowthRate {
						growthRates = append(growthRates, nil)
					}
				}
			case interfaces.METRICS_PROPORTION:
				// 请求了占比,占比是每行除以全部行的和.全部行的和
				if total != 0 {
					proportions = append(proportions, currentValue/total*100)
					// 保留两位小数
					// growthRate = math.Round(growthRate.(float64)*100) / 100
				} else {
					proportions = append(proportions, nil)
				}
			}
		}

		timeStr := convert.FormatRFC3339Milli(query.Time)
		mData := interfaces.MetricModelData{
			Labels:       labels,
			Times:        []any{query.Time},
			TimeStrs:     []string{timeStr},
			Values:       values,
			GrowthRates:  growthRates,
			GrowthValues: growthValues,
			Proportions:  proportions,
		}
		datas = append(datas, mData)
	}

	resp.Datas = datas
	span.SetStatus(codes.Ok, "")
	return processOrderHaving(resp, query, model), nil
}

// 获取占比的总数(分母)
func getProportionTotal(ctx context.Context, vegaData interfaces.VegaFetchData) (float64, error) {
	var total float64
	// 查找__value列的索引
	valueIndex := -1
	for i, col := range vegaData.Columns {
		if col.Name == interfaces.VALUE_FIELD {
			valueIndex = i
			break
		}
	}
	if valueIndex == -1 {
		return total, nil
	}
	for _, row := range vegaData.Data {
		if valueIndex >= len(row) {
			continue
		}
		val, err := getFloat64Value(ctx, row[valueIndex])
		if err != nil {
			return 0, err
		}
		total += val
	}
	return total, nil
}

// 获取序列在每个时间点上占比的总数(分母)
func getSeriesProportionTotal(ctx context.Context, seriesMap map[string]interfaces.MetricModelData) (map[string]float64, error) {
	totalMap := make(map[string]float64)

	for _, series := range seriesMap {
		for i, timeStr := range series.TimeStrs {
			val, err := getFloat64Value(ctx, series.Values[i])
			if err != nil {
				return nil, err
			}
			totalMap[timeStr] += val
		}

	}

	return totalMap, nil
}

// 范围查询时,把vega数据转成时序数据的格式
func convertVegaDatas2TimeSeries(ctx context.Context, vegaData, samePeriodDatas interfaces.VegaFetchData,
	query interfaces.MetricModelQuery, vegaDuration int64) (interfaces.MetricModelUniResponse, error) {

	resp := interfaces.MetricModelUniResponse{
		Step:           query.StepStr,
		VegaDurationMs: vegaDuration,
	}

	currentSeriesMap, err := convert2TimeSeries(ctx, vegaData, query, false)
	if err != nil {
		return resp, err
	}

	if query.RequestMetrics == nil {
		// 将map转换为slice
		datas := make([]interfaces.MetricModelData, 0, len(currentSeriesMap))
		for _, ts := range currentSeriesMap {
			datas = append(datas, ts)
		}
		resp.Datas = datas
		resp.SeriesTotal = len(datas)
		return resp, nil
	}

	// 同环比\占比
	switch query.RequestMetrics.Type {
	case interfaces.METRICS_SAMEPERIOD:
		// 同期值
		previousMap, err := convert2TimeSeries(ctx, samePeriodDatas, query, true)
		if err != nil {
			return interfaces.MetricModelUniResponse{}, err
		}
		datas, err := calcSamePeriodValue(ctx, currentSeriesMap, previousMap, query.RequestMetrics)
		if err != nil {
			return resp, err
		}
		resp.Datas = datas
		resp.SeriesTotal = len(datas)

		return resp, nil
	case interfaces.METRICS_PROPORTION:
		datas, err := calcProportionValue(ctx, currentSeriesMap)
		if err != nil {
			return resp, err
		}
		resp.Datas = datas
		resp.SeriesTotal = len(datas)
		return resp, nil
	}

	return resp, nil
}

func convert2TimeSeries(ctx context.Context, vegaData interfaces.VegaFetchData,
	query interfaces.MetricModelQuery, isSamePeriod bool) (map[string]interfaces.MetricModelData, error) {

	// 确定维度列、时间列和值列的位置
	timeIndex, valueIndex := -1, -1
	dimensionIndices := make([]int, 0)
	dimensionNames := make([]string, 0)
	for i, col := range vegaData.Columns {
		if col.Name == interfaces.TIME_FIELD {
			timeIndex = i
		} else if col.Name == interfaces.VALUE_FIELD {
			valueIndex = i
		} else {
			dimensionNames = append(dimensionNames, col.Name)
			dimensionIndices = append(dimensionIndices, i)
		}
	}
	if timeIndex == -1 || valueIndex == -1 {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_MetricModel_InternalError_ParseResultFailed).WithErrorDetails("missing __time or __value column")
	}

	// 处理成序列之和,对空缺的步长点补null
	// 时间修正, 同期值和本期值不同,同期值需要把时间置到同期的时间轴上
	fixedStart, fixedEnd := correctingTime(query, common.APP_LOCATION)
	// 如果是同期数据,需要把start和end对应到同期的start end
	if isSamePeriod {
		fixedStart = calcComparisonTime(time.UnixMilli(fixedStart).In(common.APP_LOCATION),
			*query.RequestMetrics.SamePeriodCfg).UnixMilli()
		fixedEnd = calcComparisonTime(time.UnixMilli(fixedEnd).In(common.APP_LOCATION),
			*query.RequestMetrics.SamePeriodCfg).UnixMilli()
	}
	// 生成完整的时间点
	allTimes := make([]any, 0)
	allTimeStrs := make([]string, 0)
	for currentTime := fixedStart; currentTime <= fixedEnd; {
		allTimes = append(allTimes, currentTime)
		allTimeStrs = append(allTimeStrs, convert.FormatTimeMiliis(currentTime, *query.StepStr))
		// 格式化时间
		currentTime = getNextPointTime(query, currentTime)
	}

	// 使用map来按维度分组数据
	seriesMap := make(map[string]interfaces.MetricModelData)
	for _, row := range vegaData.Data {
		// 构建labels的key
		var key string
		labels := make(map[string]string)
		for i, idx := range dimensionIndices {
			labelValue := fmt.Sprintf("%v", row[idx])
			labels[dimensionNames[i]] = labelValue
			key += labelValue + "|"
		}

		// 获取或创建TimeSeries
		ts, exists := seriesMap[key]
		if !exists {
			if query.FillNull {
				ts = interfaces.MetricModelData{
					Labels:   labels,
					Times:    allTimes,
					TimeStrs: allTimeStrs,
					Values:   make([]any, len(allTimes)),
				}
				// 初始化所有时间点的值为null
				for i := range ts.Values {
					ts.Values[i] = nil
				}
			} else {
				ts = interfaces.MetricModelData{
					Labels:   labels,
					Times:    make([]any, 0),
					TimeStrs: make([]string, 0),
					Values:   make([]any, 0),
				}
			}
		}

		// 添加时间和值
		timeStr, ok := row[timeIndex].(string)
		if !ok {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
				uerrors.Uniquery_MetricModel_InvalidParameter).
				WithErrorDetails(fmt.Sprintf(`date format result [%v] not a string`, row[timeIndex]))
		}
		if query.FillNull {
			idx := findTimeStrIndex(allTimeStrs, timeStr)
			if idx == -1 {
				continue
			}
			if row[valueIndex] != nil {
				val, err := convert.AssertFloat64(row[valueIndex])
				if err != nil {
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						uerrors.Uniquery_InternalError_AssertFloat64Failed).
						WithErrorDetails(fmt.Sprintf("err: %v, vega metric value is %v", err, row[valueIndex]))
				}
				ts.Values[idx] = val
			}
		} else {
			timei, err := convert.ParseTimeToMillis(timeStr, *query.StepStr)
			if err != nil {
				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					uerrors.Uniquery_MetricModel_InternalError_ParseResultFailed).
					WithErrorDetails(fmt.Sprintf(`parse time[%s] to unixtime with[%s] failed`, timeStr, *query.StepStr))
			}

			ts.Times = append(ts.Times, timei)
			ts.TimeStrs = append(ts.TimeStrs, timeStr)

			_, err = appendValues(ctx, row[valueIndex], &ts.Values)
			if err != nil {
				return nil, err
			}
		}

		seriesMap[key] = ts
	}

	return seriesMap, nil
}

func calcComparisonTime(t time.Time, granlarCfg interfaces.SamePeriodCfg) time.Time {
	switch granlarCfg.TimeGranularity {
	case interfaces.METRICS_SAMEPERIOD_TIME_GRANULARITY_DAY:
		return t.AddDate(0, 0, -granlarCfg.Offset)
	case interfaces.METRICS_SAMEPERIOD_TIME_GRANULARITY_MONTH:
		// 月环比, k个月同期
		newTime := t.AddDate(0, -granlarCfg.Offset, 0)
		// 处理月末日期不存在的情况
		if t.Day() != newTime.Day() {
			newTime = convert.LastDayOfMonth(newTime)
			newTime = time.Date(newTime.Year(), newTime.Month(), newTime.Day(),
				t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1e6, newTime.Location())
		}
		return newTime
	case interfaces.METRICS_SAMEPERIOD_TIME_GRANULARITY_QUARTER:
		// 上k个季度
		newTime := t.AddDate(0, -3*granlarCfg.Offset, 0)

		// 处理季度末日期不存在的情况
		if t.Day() != newTime.Day() {
			newTime = convert.LastDayOfMonth(newTime)
			newTime = time.Date(newTime.Year(), newTime.Month(), newTime.Day(),
				t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1e6, newTime.Location())
		}
		return newTime
	case interfaces.METRICS_SAMEPERIOD_TIME_GRANULARITY_YEAR:
		// 上k年
		newTime := t.AddDate(-granlarCfg.Offset, 0, 0)

		// 处理闰年2月29日的情况
		if t.Month() == time.February && t.Day() == 29 && !convert.IsLeap(newTime.Year()) {
			newTime = time.Date(newTime.Year(), time.February, 28,
				t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1e6, newTime.Location())
		}
		return newTime
	}

	return t
}

// 把vega的指标值添加到统一结构的values中
func appendValues(ctx context.Context, value any, values *[]any) (float64, error) {
	var val float64
	var err error
	if value == nil {
		*values = append(*values, nil)
	} else {
		val, err = convert.AssertFloat64(value)
		if err != nil {
			return val, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_InternalError_AssertFloat64Failed).
				WithErrorDetails(fmt.Sprintf("err: %v, vega metric value is %v", err, value))
		}
		*values = append(*values, convert.WrapMetricValue(val))
	}
	return val, nil
}

// 从vega中读取指标值,并把其转成float64
func getFloat64Value(ctx context.Context, value any) (float64, error) {
	if value == nil {
		return 0, nil
	} else {
		val, err := convert.AssertFloat64(value)
		if err != nil {
			return val, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_InternalError_AssertFloat64Failed).
				WithErrorDetails(fmt.Sprintf("err: %v, vega metric value is %v", err, value))
		}
		return val, nil
	}
}

// 在时间点序列中查找索引
func findTimeStrIndex(timePoints []string, timeStr string) int {
	for i, t := range timePoints {
		if t == timeStr {
			return i
		}
	}
	return -1
}

func calcSamePeriodValue(ctx context.Context, currentSeriesMap, previousMap map[string]interfaces.MetricModelData,
	requestMetrics *interfaces.RequestMetrics) ([]interfaces.MetricModelData, error) {

	// 检查method数组，确定需要计算哪些指标
	methods := requestMetrics.SamePeriodCfg.Method
	hasGrowthValue := false
	hasGrowthRate := false
	for _, method := range methods {
		if method == interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE {
			hasGrowthValue = true
		}
		if method == interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_RATE {
			hasGrowthRate = true
		}
	}

	datas := make([]interfaces.MetricModelData, 0)
	// 生成结果
	for key, currentPoints := range currentSeriesMap {
		// 获取对比期数据
		previousPoints := previousMap[key]

		// 创建时间序列
		ts := interfaces.MetricModelData{
			Labels:       currentPoints.Labels,
			Times:        make([]any, 0),
			TimeStrs:     make([]string, 0),
			Values:       make([]any, 0),
			GrowthValues: make([]any, 0),
			GrowthRates:  make([]any, 0),
		}
		// 处理每个时间点
		for i := range currentPoints.Times {
			ts.Times = append(ts.Times, currentPoints.Times[i])
			ts.TimeStrs = append(ts.TimeStrs, currentPoints.TimeStrs[i])
			ts.Values = append(ts.Values, currentPoints.Values[i])

			// 查找对比期数据
			compareDate := calcComparisonTime(time.UnixMilli(currentPoints.Times[i].(int64)).In(common.APP_LOCATION),
				*requestMetrics.SamePeriodCfg).UnixMilli()

			var previousV any
			for j, pt := range previousPoints.Times {
				if compareDate == pt {
					previousV = previousPoints.Values[j]
					break
				}
			}

			// 计算增长值和增长率
			if previousV != nil && currentPoints.Values[i] != nil {
				currentVal, err := getFloat64Value(ctx, currentPoints.Values[i])
				if err != nil {
					return nil, err
				}
				previousVal, err := getFloat64Value(ctx, previousV)
				if err != nil {
					return nil, err
				}
				if previousVal != 0 {
					growthValue := currentVal - previousVal
					growthRate := (growthValue / previousVal) * 100

					if hasGrowthValue {
						ts.GrowthValues = append(ts.GrowthValues, convert.WrapMetricValue(growthValue))
					}
					if hasGrowthRate {
						ts.GrowthRates = append(ts.GrowthRates, convert.WrapMetricValue(growthRate))
					}
				} else {
					if hasGrowthValue {
						ts.GrowthValues = append(ts.GrowthValues, nil)
					}
					if hasGrowthRate {
						ts.GrowthRates = append(ts.GrowthRates, nil)
					}
				}
			} else {
				if hasGrowthValue {
					ts.GrowthValues = append(ts.GrowthValues, nil)
				}
				if hasGrowthRate {
					ts.GrowthRates = append(ts.GrowthRates, nil)
				}
			}
		}
		datas = append(datas, ts)
	}
	return datas, nil
}

func calcProportionValue(ctx context.Context, currentSeriesMap map[string]interfaces.MetricModelData) ([]interfaces.MetricModelData, error) {
	// 计算每个时间点上的序列的值之和
	timeTotals, err := getSeriesProportionTotal(ctx, currentSeriesMap)
	if err != nil {
		return nil, err
	}

	datas := []interfaces.MetricModelData{}
	for _, ts := range currentSeriesMap {
		// 遍历,计算每个序列在每个时间点上的占比
		for i, timeStr := range ts.TimeStrs {
			// 计算当前序列在当前时间点的占比
			if total, ok := timeTotals[timeStr]; ok && total != 0 {
				val, err := getFloat64Value(ctx, ts.Values[i])
				if err != nil {
					return nil, err
				}

				percentage := (val / total) * 100
				ts.Proportions = append(ts.Proportions, convert.WrapMetricValue(percentage))
			} else {
				ts.Proportions = append(ts.Proportions, nil)
			}
		}
		datas = append(datas, ts)
	}
	return datas, nil
}
