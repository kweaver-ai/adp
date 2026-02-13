// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package leafnodes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/tidwall/gjson"

	cond "uniquery/common/condition"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	dtype "uniquery/interfaces/data_type"
	"uniquery/logics/promql/labels"
	"uniquery/logics/promql/parser"
	"uniquery/logics/promql/static"
)

// MakeDSL 构建请求的 dsl 语句，返回要执行的 dsl 语句以及要查询的索引库列表
func makeGetTsidDSL(expr parser.VectorSelector, query *interfaces.Query,
	firstTsid string, mustFilter interface{}) (*bytes.Buffer, int, error) {

	var queryBuffer bytes.Buffer
	// 构造 query 部分
	queryStr, _ := makeDSLQuery(expr)
	queryBuffer.WriteString(queryStr)

	// 将视图的过滤条件转成dsl, 暂时注释，原子视图没有过滤条件，自定义视图时再考虑
	// var dslStr string
	// if query.DataView.Condition != nil {
	// 	cfg := query.DataView.Condition

	// 	// 将过滤条件拼接到 dsl 的 query 中
	// 	CondCfg, err := cond.NewCondition(ctx, cfg, query.DataView.FieldScope, query.DataView.FieldsMap)
	// 	if err != nil {
	// 		return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
	// 			Typ: uerrors.ErrorExec,
	// 			Err: fmt.Errorf("common.MakeDSL new condition error: %s", err.Error()),
	// 		}
	// 	}

	// 	if CondCfg != nil {
	// 		dslStr, err = CondCfg.Convert(ctx)
	// 		if err != nil {
	// 			return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
	// 				Typ: uerrors.ErrorExec,
	// 				Err: fmt.Errorf("common.MakeDSL convert condition to dsl error: %s", err.Error()),
	// 			}
	// 		}
	// 	}
	// }

	dslStr := query.ViewQuery4Metric.QueryStr
	if dslStr != "" {
		dslStr += ","
	}
	queryBuffer.WriteString(dslStr)

	// 如果是基于指标模型的查询，则需把过滤器 filters 的过滤条件拼接到 dsl 的 query 中
	filterStr, status, err := AppendFilters(*query)
	if err != nil {
		return nil, status, err
	}
	queryBuffer.WriteString(filterStr)

	// 构造must_filter
	mstr, err := sonic.Marshal(mustFilter)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
			Typ: uerrors.ErrorExec,
			Err: fmt.Errorf("common.MakeDSL Marshal must_filter error: %s", err.Error()),
		}
	}

	// 构造 tsid range, 分批获取tsid
	queryBuffer.WriteString(`
					{
						"range": {
							"`)
	queryBuffer.WriteString(wrapKeyWordFieldName(interfaces.TSID))
	queryBuffer.WriteString(`": {
								"gt": "`)
	queryBuffer.WriteString(firstTsid)
	queryBuffer.WriteString(`"
							}
						}
					}
				],
				"must":`)
	queryBuffer.WriteString(string(mstr))
	queryBuffer.WriteString(`
			}
		}`)

	// 构造aggregation
	queryBuffer.WriteString(`,
	"aggs": {
		"`)
	queryBuffer.WriteString(interfaces.TSID)
	queryBuffer.WriteString(`": {
			"terms": {
				"field": "`)
	queryBuffer.WriteString(wrapKeyWordFieldName(interfaces.TSID))
	queryBuffer.WriteString(fmt.Sprintf(`",
				"size": %d,`, query.MaxSearchSeriesSize))
	// 如果是来自data-model的请求,是在做计算公式有效性检查,取size为1,此时按深度优先聚合性能更优
	if query.IsModelRequest {
		queryBuffer.WriteString(`"collect_mode": "depth_first",`)
	}
	queryBuffer.WriteString(`"order": {
					"_key": "asc"
				}
			},
			"aggs": {
				"labels": {
				  "top_hits": {
					"size": 1,
					"_source": {
					  "includes": "labels.*"
					}
				  }
				}
			  }
			}
		  }
	  	}`)

	logger.Debugf("获取tsid的查询语句: %s", queryBuffer.String())

	return &queryBuffer, http.StatusOK, nil
}

// 构建请求的 dsl 语句，过滤条件与MakeDSL中的不一样。这里是字段存在、时间范围。然后再由filters聚合中传递tsid做过滤聚合
func makeBatchTsidDSL(expr parser.VectorSelector, query *interfaces.Query, groupBy []string,
	aggregationType string, tsids []string) (*bytes.Buffer, int, error) {

	var queryBuffer bytes.Buffer
	str, _ := json.Marshal(tsids)
	// todo：优化点: 原本过滤条件 + tsid_range
	queryBuffer.WriteString(`{
		"size": 0,
		"query": {
			"bool": {
				"filter": [
					{
						"exists": {
							"field": "`)
	queryBuffer.WriteString(wrapMetricsFieldName(expr.Name))
	queryBuffer.WriteString(`"
						}
					},
					{
						"terms_set": {
							"`)
	queryBuffer.WriteString(wrapKeyWordFieldName(interfaces.TSID))
	queryBuffer.WriteString(`": {
							  "terms": `)
	queryBuffer.WriteString(string(str))
	queryBuffer.WriteString(`,
							  "minimum_should_match_script": {
								"source": "1"
							  }
							}
						}
					},
					{
						"range": {
							"@timestamp": {
								"gte":`)
	queryBuffer.WriteString(fmt.Sprintf(`%d,
								"lt":%d`, query.Start, query.End))

	queryBuffer.WriteString(`
						}
					}
				}
			]
		}
	}`)

	var interval int64
	switch aggregationType {
	case interfaces.IRATE_AGG, interfaces.RATE_AGG, interfaces.CHANGES_AGG, interfaces.AVG_OVER_TIME, interfaces.SUM_OVER_TIME,
		interfaces.MAX_OVER_TIME, interfaces.MIN_OVER_TIME, interfaces.COUNT_OVER_TIME:
		interval = query.SubIntervalWith2h // tsid按2h路由的子步长来查询
	default:
		interval = query.Interval
	}

	// 构造aggregation
	status, err := makeAggregation(groupBy, aggregationType, expr.Name, interval, query.MaxSearchSeriesSize,
		&queryBuffer)
	if err != nil {
		return nil, status, err
	}
	return &queryBuffer, http.StatusOK, nil
}

// makeDSL 构建请求的 dsl 语句，返回要执行的 dsl 语句以及要查询的索引库列表
func makeDSL(expr parser.VectorSelector, query *interfaces.Query,
	groupBy []string, aggregationType string, mustFilter interface{}, isTsid bool) (*bytes.Buffer, int, error) {

	var queryBuffer bytes.Buffer
	// 构造 query 部分
	queryStr, metricName := makeDSLQuery(expr)

	queryBuffer.WriteString(queryStr)

	// 构造must_filter
	mstr, err := sonic.Marshal(mustFilter)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
			Typ: uerrors.ErrorExec,
			Err: fmt.Errorf("common.MakeDSL Marshal must_filter error: %s", err.Error()),
		}
	}

	// 将视图的过滤条件转成dsl, 暂时注释，原子视图没有过滤条件，自定义视图时再考虑
	// var dslStr string
	// if query.DataView.Condition != nil {
	// 	cfg := query.DataView.Condition

	// 	// 将过滤条件拼接到 dsl 的 query 中
	// 	CondCfg, err := cond.NewCondition(ctx, cfg, query.DataView.FieldScope, query.DataView.FieldsMap)
	// 	if err != nil {
	// 		return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
	// 			Typ: uerrors.ErrorExec,
	// 			Err: fmt.Errorf("common.MakeDSL new condition error: %s", err.Error()),
	// 		}
	// 	}

	// 	if CondCfg != nil {
	// 		dslStr, err = CondCfg.Convert(ctx)
	// 		if err != nil {
	// 			return nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
	// 				Typ: uerrors.ErrorExec,
	// 				Err: fmt.Errorf("common.MakeDSL convert condition to dsl error: %s", err.Error()),
	// 			}
	// 		}
	// 	}
	// }

	// if dslStr != "" {
	// 	dslStr += ","
	// }
	// queryBuffer.WriteString(dslStr)

	// 如果是基于指标模型的查询，则需把全局过滤器 filters 的过滤条件拼接到 dsl 的 query 中
	filterStr, status, err := AppendFilters(*query)
	if err != nil {
		return nil, status, err
	}
	queryBuffer.WriteString(filterStr)

	// 构造 time range
	queryBuffer.WriteString(`
					{
						"range": {
							"@timestamp": {
								"gte":`)
	queryBuffer.WriteString(fmt.Sprintf(`%d,
								"lt":%d`, query.Start, query.End))
	queryBuffer.WriteString(`
						  }
						}
					}
				],
				"must":
					`)
	queryBuffer.WriteString(string(mstr))
	queryBuffer.WriteString(`
			}
		}`)

	var interval int64
	switch aggregationType {
	case interfaces.IRATE_AGG, interfaces.RATE_AGG, interfaces.CHANGES_AGG, interfaces.AVG_OVER_TIME, interfaces.SUM_OVER_TIME,
		interfaces.MAX_OVER_TIME, interfaces.MIN_OVER_TIME, interfaces.COUNT_OVER_TIME:
		interval = query.SubIntervalWith30min
		if isTsid {
			interval = query.SubIntervalWith2h
		}
	default:
		interval = query.Interval
	}

	// 构造aggregation
	status, err = makeAggregation(groupBy, aggregationType, metricName, interval,
		query.MaxSearchSeriesSize, &queryBuffer)
	if err != nil {
		return nil, status, err
	}
	logger.Debug(queryBuffer.String())

	return &queryBuffer, http.StatusOK, nil
}

// 构造 dsl 的过滤条。 返回 dsl 中的 query 部分、指标名称
func makeDSLQuery(expr parser.VectorSelector) (string, string) {
	var metricName string

	queryStr := `
	{
		"size":0,
		"query": {
			"bool": {
				"filter": [`

	// 根据 matcher 的 name type value 构建 dsl 过滤条件
	for _, matcher := range expr.LabelMatchers {
		if matcher.Name == labels.MetricName {
			// 指标名称过滤
			queryStr = fmt.Sprintf("%s%s", queryStr, fmt.Sprintf(`
					{
						"exists": {
							"field": "%s"
						}
					},`, wrapMetricsFieldName(matcher.Value)))
			metricName = matcher.Value
		} else {
			// 普通的 label 过滤，支持四种过滤方式
			mstr, _ := sonic.Marshal(matcher.Value)
			expression := fmt.Sprintf(`"%s": %s`, wrapKeyWordFieldName(interfaces.LABELS_PREFIX, matcher.Name), string(mstr))
			switch matcher.Type {
			case labels.MatchEqual:
				// =
				queryStr = fmt.Sprintf("%s%s", queryStr, fmt.Sprintf(`
					{
						"term": {
							%s
						}
				 	},`, expression))
			case labels.MatchNotEqual:
				// !=
				queryStr = fmt.Sprintf("%s%s", queryStr, fmt.Sprintf(`
					{
						"bool" : {
							"must_not" : {
								"term" : {
									%s
								}
							}
						}
					},`, expression))
			case labels.MatchRegexp:
				// =~
				queryStr = fmt.Sprintf("%s%s", queryStr, fmt.Sprintf(`
					{
						"regexp": {
							%s
						}
					},`, expression))
			case labels.MatchNotRegexp:
				// !~
				queryStr = fmt.Sprintf("%s%s", queryStr, fmt.Sprintf(`
					{
						"bool": {
							"must_not": {
								"regexp": {
									%s
								}
							}
						}
					},`, expression))
			}
		}
	}
	return queryStr, metricName
}

// 构造 dsl 的 aggs 部分，返回聚合查询的字符串
func makeAggregation(groupFields []string, aggregationType string, metricName string,
	interval int64, maxSearchSeriesSize int, queryStr *bytes.Buffer) (int, error) {

	var braceStr string
	// 先写 terms aggregation 部分
	for i := 0; i < len(groupFields); i++ {
		// __labels_str 要求在第一层，不用包起来
		termsAgg := fmt.Sprintf(`,
		"aggs": {
			"%s": {
				"terms": {
					"field": "%s",
					"order": { "_key": "asc" },
					"size": %d
				}`, groupFields[i], wrapKeyWordFieldName(groupFields[i]), maxSearchSeriesSize)
		queryStr.WriteString(termsAgg)
		// 拼接结尾的大括号
		braceStr = fmt.Sprintf("%s%s", braceStr, `
			}
		}`)
	}

	// value aggregation 部分根据请求类型不同构造不同
	var valueAgg string
	metricFiled := wrapMetricsFieldName(metricName)
	switch aggregationType {
	case interfaces.SAMPLING_AGG:
		valueAgg = fmt.Sprintf(`
							"sampling": {
								"value": {
									"field": "%s"
								},
								"timestamp": {
									"field": "@timestamp"
								}
							}`, metricFiled)
	case interfaces.IRATE_AGG:
		valueAgg = fmt.Sprintf(`
						"irate_sampling": {
								"value": {
									"field": "%s"
								},
								"timestamp": {
									"field": "@timestamp"
								}
							}`, metricFiled)
	case interfaces.RATE_AGG:
		valueAgg = fmt.Sprintf(`
							"rate_sampling": {
								"value": {
									"field": "%s"
								},
								"timestamp": {
									"field": "@timestamp"
								}
							}`, metricFiled)

	case interfaces.CHANGES_AGG:
		valueAgg = fmt.Sprintf(`
							"changes_sampling": {
								"value": {
									"field": "%s"
								},
								"timestamp": {
									"field": "@timestamp"
								}
							}`, metricFiled)

	case interfaces.AVG_OVER_TIME, interfaces.SUM_OVER_TIME:
		valueAgg = fmt.Sprintf(`
							"sum": {
								"field": "%s"
							}`, metricFiled)
	case interfaces.COUNT_OVER_TIME:
		valueAgg = fmt.Sprintf(`
							"value_count": {
								"field": "%s"
							}`, metricFiled)
	case interfaces.MAX_OVER_TIME:
		valueAgg = fmt.Sprintf(`
							"max": {
								"field": "%s"
							}`, metricFiled)
	case interfaces.MIN_OVER_TIME:
		valueAgg = fmt.Sprintf(`
							"min": {
								"field": "%s"
							}`, metricFiled)
	case interfaces.DELTA_AGG:
		valueAgg = fmt.Sprintf(`
							"delta_sampling": {
								"value": {
									"field": "%s"
								},
								"timestamp": {
									"field": "@timestamp"
								}
							}`, metricFiled)
	default:
		return http.StatusUnprocessableEntity, uerrors.PromQLError{
			Typ: uerrors.ErrorExec,
			Err: errors.New("unsupport aggregation type"),
		}
	}

	queryStr.WriteString(fmt.Sprintf(`,
			"aggs": {
				"time": {
					"date_histogram": {
						"field": "@timestamp",
						"fixed_interval": "%dms",
						"min_doc_count": 1,
						"time_zone":"%s",
						"order": {
							"_key": "asc"
						}
					}`, interval, os.Getenv("TZ")))

	queryStr.WriteString(`
		,
			"aggs": {
				"value": {
					`)
	queryStr.WriteString(valueAgg)
	queryStr.WriteString(`
				}
			  }
		    }
		  }`)
	queryStr.WriteString(braceStr)
	queryStr.WriteString(`}`)
	return http.StatusOK, nil
}

// 转换成 keyword
func wrapKeyWordFieldName(fields ...string) string {
	for _, field := range fields {
		if field == "" {
			logger.Warn("missing field name")
			return ""
		}
	}

	return strings.Join(fields, ".") + "." + interfaces.KEYWORD_SUFFIX
}

// metrics 字段添加 metrics. 前缀
func wrapMetricsFieldName(field string) string {
	if field == "" {
		logger.Warn("missing metric name")
		return ""
	}
	return interfaces.METRICS_PREFIX + "." + field
}

// 按时间窗去重k个数组中的样本点
func duplicatePoint(pointMap map[int64]*static.Point, tsArr [][]gjson.Result) {
	for _, tsArri := range tsArr {
		for _, pointij := range tsArri {
			currentT := pointij.Get("key").Int()
			sampleT := pointij.Get("value.timestamp").Int()
			value := pointij.Get("value.value").Float()
			if point, ok := (pointMap)[currentT]; ok {
				// 合并去重。取时间戳较大者
				if point.T < sampleT {
					(pointMap)[currentT] = &static.Point{
						T: sampleT,
						V: value,
					}
				} else if point.T == sampleT {
					// 如果时间戳相等，那就取值较大者
					(pointMap)[currentT] = &static.Point{
						T: sampleT,
						V: math.Max((pointMap)[currentT].V, value),
					}
				}
			} else {
				(pointMap)[currentT] = &static.Point{
					T: sampleT,
					V: value,
				}
			}
		}
	}
}

// 处理 labels 结构解析，把string解析成[]labels.Label。
// key 就是 __labels_str，形如: "labels.cluster=\"opensearch\",labels.index=\"agent_audit_log-2022.05-0\",labels.job=\"prometheus\""
func parseLabelsStr(key string, labelsMap map[string][]*labels.Label) labels.Labels {
	// 如果从 opensearch 中拿到的 __labels_str 是空串，那么就直接返回空，不需要解析。
	var metric labels.Labels

	if len(strings.TrimSpace(key)) == 0 {
		return metric
	}

	labelsStrArr := strings.Split(key, ",")
	for i := 0; i < len(labelsStrArr); i++ {
		labelArr := strings.Split(labelsStrArr[i], "=")
		// 如果解析出来的labels不符合规范，那么不解析，按__labels_str返回
		if len(labelArr) != 2 {
			lb := labelsMap[key]
			metric = lb
			break
		}

		// 对结果去掉空格
		metric = append(metric, &labels.Label{
			Name:  strings.TrimSpace(labelArr[0]),
			Value: strings.ReplaceAll(strings.TrimSpace(labelArr[1]), `"`, ""),
		})

	}
	return metric.Sort()
}

// 递归获取 terms 层次的 agg, 组装 labels map 和 tsValueMap
func iteratorTermsAgg(aggregations gjson.Result, keystr string, labelsArri []*labels.Label, groupByi int,
	groupBy []string, mapResult *MapResult) {

	for _, bucket := range aggregations.Get(fmt.Sprintf("%s.buckets", groupBy[groupByi])).Array() {
		// 拿到时间序列，先放着待合并
		var keys string
		bukectKey := bucket.Get("key").String()
		labelArr := make([]*labels.Label, 0)
		labelArr = append(labelArr, labelsArri...)
		labelArr = append(labelArr, &labels.Label{
			Name:  groupBy[groupByi],
			Value: bukectKey,
		})

		if groupByi == 0 {
			keys = bukectKey
		} else {
			keys = fmt.Sprintf("%s|%s", keystr, bukectKey)
		}
		if groupByi == len(groupBy)-1 {
			mapResult.LabelsMap[keys] = labelArr
			mapResult.TsValueMap[keys] = append(mapResult.TsValueMap[keys], bucket.Get("time.buckets").Array())
		} else {
			iteratorTermsAgg(bucket, keys, labelArr, groupByi+1, groupBy, mapResult)
		}
	}
}

// 对于rate, irate算子的 instant query: interval=range,
// 因此 range 应满足在 >30m(__labels_str) >2h (__tsid) 的情况下是 5m 的倍数。取能兼容的，还是大于30m时的要求
func checkInstantQueryInterval(interval time.Duration) error {
	if interval > interfaces.SHARD_ROUTING_30M && interval%interfaces.DEFAULT_STEP_DIVISOR != 0 {
		err := errors.New("if instant query and range > 2h, range should a multiple of 5 minutes")
		return uerrors.PromQLError{Typ: uerrors.ErrorBadData, Err: err}
	}

	return nil
}

// 获取所有样本点的最小、最大时间戳, 以及样本点总数
func getPointsStatsInfo(tsArr [][]gjson.Result) (int64, int64, int) {
	var start, end int64
	start, end = math.MaxInt64, math.MinInt64
	totalNum := 0
	for i := 0; i < len(tsArr); i++ {
		totalNum += len(tsArr[i])
		minTs := tsArr[i][0].Get("key").Int()
		if minTs < start {
			start = minTs
		}
		maxTs := tsArr[i][len(tsArr[i])-1].Get("key").Int()
		if maxTs > end {
			end = maxTs
		}
	}

	return start, end, totalNum
}

func processParam(expr *parser.MatrixSelector, query *interfaces.Query) (
	*parser.VectorSelector, *interfaces.Query, int, error) {

	start := query.Start
	exprRange := expr.Range
	if query.IsInstantQuery {
		// if range auto is true,range = lookbackDelta, default 5min
		lookbackDeltaT := exprRange.Milliseconds()
		if expr.Auto && exprRange == 0 {
			// lookbackDeltaT = convert.GetLookBackDelta(query.LookBackDelta, lookbackDelta)
			lookbackDeltaT = query.End - query.Start
			exprRange = time.Duration(lookbackDeltaT) * time.Millisecond
			// exprRange = lookbackDelta
			// if exprRange == 0 {
			// 	exprRange = interfaces.DEFAULT_LOOK_BACK_DELTA
			// }
		}
		// instant query: interval=range, 因此 range 应满足在 >30m 的情况下是 5m 的倍数
		if err := checkInstantQueryInterval(exprRange); err != nil {
			return nil, nil, http.StatusBadRequest, err
		}
		// 结束时间 - range
		start = query.End - lookbackDeltaT
		// start = query.Start - lookbackDeltaT
		query.Interval = lookbackDeltaT
	}

	vs, ok := expr.VectorSelector.(*parser.VectorSelector)
	if !ok {
		return nil, nil, http.StatusUnprocessableEntity, uerrors.PromQLError{
			Typ: uerrors.ErrorInternal,
			Err: errors.New("changes is not VectorSelector type"),
		}
	}

	intervalDuration := time.Duration(query.Interval) * time.Millisecond

	// fix range: if range auto is true, range = step
	if !query.IsInstantQuery && expr.Auto && expr.Range == 0 {
		exprRange = intervalDuration
	}

	// 如果range>30m路由, substep是range和step的公约数, substep需要满足是 <=30m 中的较大的值
	if intervalDuration > interfaces.SHARD_ROUTING_30M {
		subInterval := calcSubInterval(intervalDuration, exprRange, interfaces.SHARD_ROUTING_30M)
		query.SubIntervalWith30min = subInterval.Milliseconds()
	} else {
		// 如果range<=30min路由, substep是range和step的最大公约数
		selRange := exprRange.Milliseconds()
		query.SubIntervalWith30min = getCommonDivisor(selRange, query.Interval)
	}

	// 如果range>2h路由, substep是range和step的公约数, substep需要满足是 <=2h 中的较大的值
	if intervalDuration > interfaces.SHARD_ROUTING_2H {
		subInterval := calcSubInterval(intervalDuration, exprRange, interfaces.SHARD_ROUTING_2H)
		query.SubIntervalWith2h = subInterval.Milliseconds()
	} else {
		// 如果range<=30min路由, substep是range和step的最大公约数
		selRange := exprRange.Milliseconds()
		query.SubIntervalWith2h = getCommonDivisor(selRange, query.Interval)
	}

	newQuery := *query
	newQuery.Start = start
	expr.Range = exprRange

	return vs, &newQuery, http.StatusOK, nil
}

// 根据仪表盘提交的全局过滤器的过滤条件拼接到 dsl 请求的 query 部分
func AppendFilters(query interfaces.Query) (string, int, error) {
	if !query.IsMetricModel {
		return "", http.StatusOK, nil
	}

	filters := query.Filters
	if len(filters) <= 0 {
		return "", http.StatusOK, nil
	}
	filterStr := ""
	// 过滤器间是 and，过滤器内是 or
	for _, filter := range filters {
		filterStri := ""

		// 从日志分组的字段信息中获取字段的类型，如果是 text 就给字段名加 .keyword；
		// 如果是脱敏字段，text 类型的加上 _desensitize.keyword, 其余类型的字段加上 _desensitize
		filterField := filter.Name
		fieldInfo, ok1 := query.DataView.FieldsMap[filter.Name]
		if ok1 {
			filterField = fieldInfo.OriginalName // filter的name是视图字段，过滤时需转成 original name
		}
		desensitizeFiled := filterField + interfaces.DESENSITIZE_FIELD_SUFFIX

		_, ok2 := query.DataView.FieldsMap[desensitizeFiled]
		if ok1 && ok2 {
			// 脱敏字段
			filterField = desensitizeFiled
		}

		var fieldType string
		if fieldInfo != nil {
			fieldType = fieldInfo.Type
		}

		if fieldType == interfaces.TEXT_TYPE {
			filterField = wrapKeyWordFieldName(filterField)
		}

		switch filter.Operation {
		case interfaces.OPERATION_IN:
			value := filter.Value.([]interface{})
			for i := 0; i < len(value); i++ {
				v := value[i]
				_, ok := value[i].(string)
				if ok {
					v = fmt.Sprintf(`"%s"`, value[i])
				}
				filterStri = fmt.Sprintf("%s%s", filterStri, fmt.Sprintf(`
						{
							"term": {
								"%s": {
									"value": %v
								}
							}
						}`, filterField, v))
				if i != len(value)-1 {
					filterStri = fmt.Sprintf("%s,", filterStri)
				}
			}
			filterStri = fmt.Sprintf(`
					{
						"bool": {
							"should": [%s
							]
						}
					},`, filterStri)

		case interfaces.OPERATION_EQ, cond.OperationEq:
			v := filter.Value
			vStr, ok := filter.Value.(string)
			if ok {
				v = fmt.Sprintf(`"%s"`, vStr)
			}
			filterStri = fmt.Sprintf("%s%s", filterStri, fmt.Sprintf(`
					{
						"term": {
							"%s": {
								"value": %v
							}
						}
					},`, filterField, v))
		case cond.OperationNotEq:
			v := filter.Value
			vStr, ok := filter.Value.(string)
			if ok {
				v = fmt.Sprintf(`"%s"`, vStr)
			}

			filterStri = fmt.Sprintf(`
					{
						"bool": {
							"must_not": [
								{
									"term": {
										"%s": {
											"value": %v
										}
									}
								}
							]
						}
					},`, filterField, v)

		case cond.OperationRange:
			value := filter.Value.([]interface{})
			if len(value) != 2 {
				return "", http.StatusUnprocessableEntity, errors.New("when filter's operation is range, the value should be an array with length is 2. ")
				// uerrors.PromQLError{
				// 	Typ: uerrors.ErrorExec,
				// 	Err: errors.New("when filter's operation is range, the value should be an array with length is 2. "),
				// }
			}
			if !cond.IsSameType(value) {
				return "", http.StatusUnprocessableEntity, errors.New("condition [range] right value should be of the same type")
				// uerrors.PromQLError{
				// 	Typ: uerrors.ErrorExec,
				// 	Err: errors.New("condition [range] right value should be of the same type"),
				// }
			}

			gte := value[0]
			lte := value[1]

			if _, ok := gte.(string); ok {
				gte = fmt.Sprintf(`"%s"`, gte)
				lte = fmt.Sprintf(`"%s"`, lte)
			}

			if fieldType == dtype.DataType_Datetime {
				var format string
				switch gte.(type) {
				case string:
					format = "strict_date_optional_time"
				case float64, int64:
					format = "epoch_millis"
				}

				filterStri = fmt.Sprintf("%s%s", filterStri, fmt.Sprintf(`
					{
						"range": {
							"%s": {
								"gte": %v,
								"lt": %v,
								"format": "%s"
							}
						}
					},`, filterField, gte, lte, format))
			} else {
				// range 左闭右开区间
				filterStri = fmt.Sprintf("%s%s", filterStri, fmt.Sprintf(`
					{
						"range": {
							"%s": {
								"gte": %v,
								"lt": %v
							}
						}
					},`, filterField, gte, lte))
			}

		case cond.OperationOutRange:
			value := filter.Value.([]interface{})
			if len(value) != 2 {
				return "", http.StatusUnprocessableEntity, errors.New("when filter's operation is out_range, the value should be an array with length is 2. ")
				// uerrors.PromQLError{
				// 	Typ: uerrors.ErrorExec,
				// 	Err: errors.New("when filter's operation is out_range, the value should be an array with length is 2. "),
				// }
			}
			if !cond.IsSameType(value) {
				return "", http.StatusUnprocessableEntity, errors.New("condition [out_range] right value should be of the same type")
				// uerrors.PromQLError{
				// 	Typ: uerrors.ErrorExec,
				// 	Err: errors.New("condition [out_range] right value should be of the same type"),
				// }
			}

			lt := value[0]
			gte := value[1]

			if _, ok := lt.(string); ok {
				lt = fmt.Sprintf(`"%s"`, lt)
				gte = fmt.Sprintf(`"%s"`, gte)
			}

			if fieldType == dtype.DataType_Datetime {
				var format string
				switch lt.(type) {
				case string:
					format = "strict_date_optional_time"
				case float64:
					format = "epoch_millis"
				default:
					return "", http.StatusUnprocessableEntity, errors.New("unsupport date type")
				}

				filterStri = fmt.Sprintf("%s%s", filterStri, fmt.Sprintf(`
					{
						"bool": {
							"should": [
								{
									"range": {
										"%s": {
											"lt": %v,
											"format": "%s"
										}
									}
								},
								{
									"range": {
										"%s": {
											"gte":  %v,
											"format": "%s"
										}
									}
								}
							]
						}
					},`, filterField, lt, format, filterField, gte, format))
			} else {
				// out_range  (-inf, value[0]] || [value[1], +inf)
				filterStri = fmt.Sprintf("%s%s", filterStri, fmt.Sprintf(`
					{
						"bool": {
							"should": [
								{
									"range": {
										"%s": {
											"lt": %v
										}
									}
								},
								{
									"range": {
										"%s": {
											"gte":  %v
										}
									}
								}
							]
						}
					},`, filterField, lt, filterField, gte))
			}

		case cond.OperationLike:
			v := filter.Value
			v = fmt.Sprintf(`".*%v.*"`, v)
			// vStr, ok := filter.Value.(string)
			// if ok {
			// 	v = fmt.Sprintf(`".*%s.*"`, vStr)
			// }
			filterStri = fmt.Sprintf("%s%s", filterStri, fmt.Sprintf(`
					{
						"regexp": {
							"%s": %v
						}
					},`, filterField, v))

		case cond.OperationNotLike:
			v := filter.Value
			v = fmt.Sprintf(`".*%v.*"`, v)
			// vStr, ok := filter.Value.(string)
			// if ok {
			// 	v = fmt.Sprintf(`".*%s.*"`, vStr)
			// }

			filterStri = fmt.Sprintf(`
					{
						"bool": {
							"must_not": [
								{
									"regexp": {
										"%s": %v
									}
								}
							]
						}
					},`, filterField, v)

		case cond.OperationGt:
			v := filter.Value
			vStr, ok := filter.Value.(string)
			if ok {
				v = fmt.Sprintf(`"%s"`, vStr)
			}
			filterStri = fmt.Sprintf("%s%s", filterStri, fmt.Sprintf(`
					{
						"range": {
							"%s": {
								"gt": %v
							}
						}
					},`, filterField, v))

		case cond.OperationGte:
			v := filter.Value
			vStr, ok := filter.Value.(string)
			if ok {
				v = fmt.Sprintf(`"%s"`, vStr)
			}
			filterStri = fmt.Sprintf("%s%s", filterStri, fmt.Sprintf(`
					{
						"range": {
							"%s": {
								"gte": %v
							}
						}
					},`, filterField, v))

		case cond.OperationLt:
			v := filter.Value
			vStr, ok := filter.Value.(string)
			if ok {
				v = fmt.Sprintf(`"%s"`, vStr)
			}
			filterStri = fmt.Sprintf("%s%s", filterStri, fmt.Sprintf(`
					{
						"range": {
							"%s": {
								"lt": %v
							}
						}
					},`, filterField, v))

		case cond.OperationLte:
			v := filter.Value
			vStr, ok := filter.Value.(string)
			if ok {
				v = fmt.Sprintf(`"%s"`, vStr)
			}
			filterStri = fmt.Sprintf("%s%s", filterStri, fmt.Sprintf(`
					{
						"range": {
							"%s": {
								"lte": %v
							}
						}
					},`, filterField, v))

		default:
			return "", http.StatusUnprocessableEntity, errors.New("unsupport operation type")
			// uerrors.PromQLError{
			// 	Typ: uerrors.ErrorExec,
			// 	Err: errors.New("unsupport operation type"),
			// }
		}

		filterStr = fmt.Sprintf("%s%s", filterStr, filterStri)
	}
	return filterStr, http.StatusOK, nil
}
