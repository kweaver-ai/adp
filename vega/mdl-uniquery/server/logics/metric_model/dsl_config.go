// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric_model

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"

	"uniquery/common"
	cond "uniquery/common/condition"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	dtype "uniquery/interfaces/data_type"
)

const (
	MAX_BUCKET_SIZE int = 10
)

// 解析 dsl 查询语句
func generateDslByConfig(ctx context.Context, query *interfaces.MetricModelQuery, dataView *interfaces.DataView) (dsl map[string]any, err error) {
	if query.FormulaConfig == nil {
		err = errors.New("missing formula_config")
		return
	}
	dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)

	bktInfos := dslConfig.Buckets
	aggrInfo := dslConfig.Aggregation
	dhInfo := dslConfig.DateHistogram
	if len(bktInfos) > MAX_BUCKET_SIZE {
		err = errors.New("too many buckets in formula_config")
		return
	}
	if aggrInfo == nil {
		err = errors.New("missing aggregation in formula_config")
		return
	}
	if !query.IsInstantQuery && dhInfo == nil {
		err = errors.New("missing date historgram in formula_config")
		return
	}

	if query.DateField != "" {
		_, ok := dataView.FieldsMap[query.DateField]
		if !ok {
			err = fmt.Errorf("date field '%s' is not in data view", query.DateField)
			return
		}
	}

	// if query.IsInstantQuery {
	// 	query.Start = query.Time - query.LookBackDelta
	// 	query.End = query.Time
	// }

	dsl = map[string]any{
		"size": 0,
	}
	crtDsl := dsl

	// parse buckets
	for bktIdx, bktInfo := range bktInfos {
		bucketDsl := map[string]any{}

		switch bktInfo.Type {
		case interfaces.BUCKET_TYPE_TERMS:
			/*
				"terms": {
					"field": "fieldname",
					"size":  1000,
					"missing": "-",
					"value_type": "string",
					"order": {
						"_key": "asc"
					}
				}
			*/
			field, ok := dataView.FieldsMap[bktInfo.Field]
			if !ok {
				err = fmt.Errorf("field '%s' is not in data view", bktInfo.Field)
				return
			}

			bktDsl := map[string]any{
				"field": wrapFieldName(field),
			}
			if dtype.DataType_IsNumber(field.Type) {
				bktDsl["missing"] = 0
				bktDsl["value_type"] = "number"
			} else if dtype.DataType_IsString(field.Type) {
				bktDsl["missing"] = "--"
				bktDsl["value_type"] = "string"
			}

			if bktInfo.Size != 0 {
				bktDsl["size"] = bktInfo.Size
			} else {
				bktDsl["size"] = math.MaxInt32
			}

			direction := interfaces.DEFAULT_DIRECTION
			if bktInfo.Direction != "" {
				if bktInfo.Direction != interfaces.ASC_DIRECTION && bktInfo.Direction != interfaces.DESC_DIRECTION {
					err = fmt.Errorf("invalid direction '%s'", bktInfo.Direction)
					return
				}
				direction = bktInfo.Direction
			}
			if bktInfo.Order != "" {
				switch bktInfo.Order {
				case interfaces.TERMS_ORDER_TYPE_FIELD:
					bktDsl["order"] = map[string]any{
						"_key": direction,
					}
				case interfaces.TERMS_ORDER_TYPE_VALUE:
					if aggrInfo.Type == interfaces.AGGR_TYPE_DOC_COUNT {
						bktDsl["order"] = map[string]any{
							"_count": direction,
						}
					} else {
						bktDsl["order"] = map[string]any{
							interfaces.VALUE_FIELD: direction,
						}
					}
				case interfaces.TERMS_ORDER_TYPE_COUNT:
					bktDsl["order"] = map[string]any{
						"_count": direction,
					}
				default:
					err = fmt.Errorf("invalid order type '%s'", bktInfo.Order)
					return
				}
			}
			bucketDsl[interfaces.BUCKET_TYPE_TERMS] = bktDsl

		case interfaces.BUCKET_TYPE_RANGE:
			/*
				"range": {
					"field": "fieldname",
					"ranges": [
						{
							"key":"low",
							"to": 1234
						},
						{
							"key":"mid",
							"from": 1234,
							"to": 12345
						},
						{
							"key":"high",
							"from": 12345
						}
					]
				}
			*/
			field, ok := dataView.FieldsMap[bktInfo.Field]
			if !ok {
				err = fmt.Errorf("field '%s' is not in data view", bktInfo.Field)
				return
			}
			if !dtype.DataType_IsNumber(field.Type) {
				err = fmt.Errorf("bucket range is not support field '%s' type '%s'", bktInfo.Field, field.Type)
				return
			}
			if len(bktInfo.Ranges) == 0 {
				err = fmt.Errorf("missing ranges in bucket range")
				return
			}

			for idx, rg := range bktInfo.Ranges {
				if rg.Key == "" {
					from := rg.From
					if from == nil {
						from = "*"
					}
					to := rg.To
					if to == nil {
						to = "*"
					}
					key := fmt.Sprintf("%v-%v", from, to)
					bktInfo.Ranges[idx].Key = key
				}
			}
			bucketDsl[interfaces.BUCKET_TYPE_RANGE] = map[string]any{
				"field":  bktInfo.Field,
				"ranges": bktInfo.Ranges,
			}

		case interfaces.BUCKET_TYPE_FILTERS:
			/*
				"filters": {
					"other_bucket": true,
					"other_bucket_key": "__other",
					"filters": {
						"1":{
							"query_string": {"query": "value:1"}
						},
						"12":{
							"query_string": {"query": "value:12"}
						},
						"123":{
							"query_string": {"query": "value:123"}
						}
					}
				}
			*/
			if bktInfo.Name == "" {
				bktInfo.Name = fmt.Sprintf("filters%d", bktIdx)
			}
			if len(bktInfo.Filters) == 0 {
				err = fmt.Errorf("missing filters in bucket filters")
				return
			}
			filters := map[string]any{}
			for name, filter := range bktInfo.Filters {
				if filter.QueryString == "" {
					err = fmt.Errorf("missing query_string in bucket filters")
					return
				}
				filters[name] = map[string]any{
					"query_string": map[string]any{
						"query":            filter.QueryString,
						"analyze_wildcard": true,
					},
				}
			}
			bucketDsl[interfaces.BUCKET_TYPE_FILTERS] = map[string]any{
				"other_bucket":     bktInfo.OtherBucket,
				"other_bucket_key": interfaces.OTHER_FIELD,
				"filters":          filters,
			}

		case interfaces.BUCKET_TYPE_GEOHASH_GRID:
			/*
				"geohash_grid": {
					"field": "geo.coordinates",
					"precision": 4
				}
			*/
			field, ok := dataView.FieldsMap[bktInfo.Field]
			if !ok {
				err = fmt.Errorf("field '%s' is not in data view", bktInfo.Field)
				return
			}
			if field.Type != dtype.DataType_Point {
				err = fmt.Errorf("bucket geohash_grid is not support field '%s' type '%s'", bktInfo.Field, field.Type)
				return
			}
			if bktInfo.Precision < interfaces.BUCKET_GEOHASH_GRID_MIN_PRECISION ||
				bktInfo.Precision > interfaces.BUCKET_GEOHASH_GRID_MAX_PRECISION {
				err = fmt.Errorf("bucket geohash_grid is not support Precision: %d", bktInfo.Precision)
				return
			}

			bucketDsl[interfaces.BUCKET_TYPE_GEOHASH_GRID] = map[string]any{
				"field":     bktInfo.Field,
				"precision": bktInfo.Precision,
			}

		case interfaces.BUCKET_TYPE_DATE_RANGE:
			/*
				"date_range": {
					"field": "@timestamp",
					"ranges": [
						{
							"key":"old",
							"from": "2024-09-10T00:00:00+08:00",
							"to": "2024-09-10T03:00:00+08:00"
						},
						{
							"key":"new",
							"from": "2024-09-10T03:00:00+08:00",
							"to": "2024-09-10T06:00:00+08:00"
						}
					]
				}
			*/
			field, ok := dataView.FieldsMap[bktInfo.Field]
			if !ok {
				err = fmt.Errorf("field '%s' is not in data view", bktInfo.Field)
				return
			}
			if field.Type != dtype.DataType_Datetime {
				err = fmt.Errorf("bucket date_range is not support field '%s' type '%s'", bktInfo.Field, field.Type)
				return
			}
			if len(bktInfo.Ranges) == 0 {
				err = fmt.Errorf("missing ranges in bucket date_range")
				return
			}

			for idx, rg := range bktInfo.Ranges {
				if rg.Key == "" {
					from := rg.From
					if from == nil {
						from = "*"
					}
					to := rg.To
					if to == nil {
						to = "*"
					}
					key := fmt.Sprintf("%v-%v", from, to)
					bktInfo.Ranges[idx].Key = key
				}
			}

			bucketDsl[interfaces.BUCKET_TYPE_DATE_RANGE] = map[string]any{
				"field":  bktInfo.Field,
				"ranges": bktInfo.Ranges,
			}

		default:
			err = fmt.Errorf("invalid bucket type: %s", bktInfo.Type)
			return
		}

		if bktInfo.Name == "" {
			bktInfo.Name = bktInfo.Field
		}
		if bktInfo.Name == "" {
			err = fmt.Errorf("missing bucket name, bucket idx: %d, bucket type: %s", bktIdx, bktInfo.Type)
			return
		}
		bktInfo.BktName = strings.ReplaceAll(bktInfo.Name, ".", "_")
		crtDsl["aggs"] = map[string]any{
			bktInfo.BktName: bucketDsl,
		}
		crtDsl = bucketDsl
	}

	// parse date_histogram
	if !query.IsInstantQuery {
		/*
			"date_histogram": {
				"field": "@timestamp",
				"calendar_interval": "hour",
				// "fixed_interval": "1h",
			}
		*/
		field, ok := dataView.FieldsMap[dhInfo.Field]
		if !ok {
			err = fmt.Errorf("field '%s' is not in data view", dhInfo.Field)
			return
		}
		if field.Type != dtype.DataType_Datetime {
			err = fmt.Errorf("date_histogram is not support field '%s' type '%s'", dhInfo.Field, field.Type)
			return
		}

		if query.DateField == "" {
			query.DateField = dhInfo.Field
		}

		bktDsl := map[string]any{
			"field":     dhInfo.Field,
			"time_zone": interfaces.DEFAULT_QUERY_TIME_ZONE.String(),
		}
		switch dhInfo.IntervalType {
		case interfaces.INTERVAL_TYPE_CALENDAR:
			query.IsCalendar = true
			interval := dhInfo.IntervalValue
			if interval == interfaces.AUTO_INTERVAL {
				interval = *query.StepStr
			}
			intervalN, ok := interfaces.CALENDAR_INTERVALS[interval]
			if !ok {
				err = fmt.Errorf("invalid interval_value or step: '%s'", interval)
				return
			}
			query.StepStr = &intervalN
			dhInfo.IntervalValue = intervalN
			bktDsl["calendar_interval"] = intervalN

		case interfaces.INTERVAL_TYPE_FIXED:
			query.IsCalendar = false
			interval := dhInfo.IntervalValue
			if interval == interfaces.AUTO_INTERVAL {
				interval = *query.StepStr
			}
			var intervalD time.Duration
			intervalD, err = convert.ParseDuration(interval)
			if err != nil {
				err = fmt.Errorf("invalid interval_value or step: '%s', %s", interval, err.Error())
				return
			}
			query.StepStr = &interval
			dhInfo.IntervalValue = interval
			step := intervalD.Milliseconds()
			query.Step = &step
			bktDsl["fixed_interval"] = interval

		default:
			err = fmt.Errorf("invalid interval_type: %s", dhInfo.IntervalType)
			return
		}

		bucketDsl := map[string]any{
			interfaces.BUCKET_TYPE_DATE_HISTOGRAM: bktDsl,
		}

		crtDsl["aggs"] = map[string]any{
			interfaces.DATE_HISTOGRAM_FIELD: bucketDsl,
		}
		crtDsl = bucketDsl
	}

	// parse aggr
	if aggrInfo.Type == interfaces.AGGR_TYPE_DOC_COUNT {
		if query.IsInstantQuery && len(bktInfos) == 0 {
			query.TraceTotalHits = true
		}
	} else {
		field, ok := dataView.FieldsMap[aggrInfo.Field]
		if !ok {
			err = fmt.Errorf("field '%s' is not in data view", aggrInfo.Field)
			return
		}

		aggrDsl := map[string]any{}
		switch aggrInfo.Type {
		case interfaces.AGGR_TYPE_VALUE_COUNT:
			aggrDsl[aggrInfo.Type] = map[string]any{
				"field": wrapFieldName(field),
			}
		case interfaces.AGGR_TYPE_CARDINALITY:
			aggrDsl[aggrInfo.Type] = map[string]any{
				"field":               wrapFieldName(field),
				"precision_threshold": interfaces.DEFAULT_CARDINALITY_PRECISION_THRESHOLD,
			}
		case interfaces.AGGR_TYPE_SUM, interfaces.AGGR_TYPE_AVG:
			if !dtype.DataType_IsNumber(field.Type) {
				err = fmt.Errorf("aggregation '%s' does not support field '%s' type '%s'",
					aggrInfo.Type, field.Name, field.Type)
				return
			}
			aggrDsl[aggrInfo.Type] = map[string]any{
				"field": aggrInfo.Field,
			}
		case interfaces.AGGR_TYPE_MAX, interfaces.AGGR_TYPE_MIN:
			if !dtype.DataType_IsNumber(field.Type) && field.Type != dtype.DataType_Datetime {
				err = fmt.Errorf("aggregation '%s' does not support field '%s' type '%s'",
					aggrInfo.Type, field.Name, field.Type)
				return
			}
			aggrDsl[aggrInfo.Type] = map[string]any{
				"field": aggrInfo.Field,
			}
		case interfaces.AGGR_TYPE_PERCENTILES:
			if !dtype.DataType_IsNumber(field.Type) {
				err = fmt.Errorf("aggregation percentiles does not support field '%s' type '%s'", field.Name, field.Type)
				return
			}
			if len(aggrInfo.Percents) == 0 {
				err = fmt.Errorf("missing percents in aggregation percentiles")
				return
			}
			for _, p := range aggrInfo.Percents {
				if p < 0 || p > 100 {
					err = fmt.Errorf("invalid percents '%f' in aggregation percentiles", p)
					return
				}
			}
			aggrDsl[aggrInfo.Type] = map[string]any{
				"field":    aggrInfo.Field,
				"percents": aggrInfo.Percents,
			}
		default:
			err = fmt.Errorf("invalid aggregation type: %s", aggrInfo.Type)
			return
		}

		crtDsl["aggs"] = map[string]any{
			interfaces.VALUE_FIELD: aggrDsl,
		}
	}

	// 拼接过滤条件
	err = ParseQuery(ctx, query, dsl, dataView)
	if err != nil {
		return
	}

	output, _ := sonic.MarshalString(&dsl)
	logger.Debugf("%s", output)
	return
}

// DSL 的返回结构转换为统一结构
func parseDSLResult2UniresponseForDslConfig(ctx context.Context, dslRes []byte, model interfaces.MetricModel,
	query interfaces.MetricModelQuery) (interfaces.MetricModelUniResponse, error) {

	_, span := ar_trace.Tracer.Start(ctx, "Parse dsl result to uniresponse")
	defer span.End()

	dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)

	bktInfos := dslConfig.Buckets
	aggrInfo := dslConfig.Aggregation
	//dhInfo := query.FormulaConfig.DateHistogram

	// 时间修正
	dhInfo := interfaces.AggInfo{ZoneLocation: interfaces.DEFAULT_QUERY_TIME_ZONE}
	fixedStart, fixedEnd := correctingTime(query, dhInfo.ZoneLocation)
	query.FixedStart = fixedStart
	query.FixedEnd = fixedEnd

	// var resp *interfaces.MetricModelUniResponse
	// 递归获取 terms 层次的 agg, 组装 labels map 和 tsValueMap
	logger.Debugf("%s", string(dslRes))
	rootNode, err := sonic.Get(dslRes)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("sonic.GetFromString error: %s", err.Error()))
		return interfaces.MetricModelUniResponse{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_MetricModel_InternalError).WithErrorDetails(err.Error())
	}

	aggrNode := rootNode.Get("aggregations")
	if !aggrNode.Valid() {
		if aggrInfo.Type == interfaces.AGGR_TYPE_DOC_COUNT && len(bktInfos) == 0 {
			docCount, err := rootNode.GetByPath("hits", "total", "value").Int64()
			if err != nil {
				o11y.Error(ctx, "root.GetByPath(\"hits\", \"total\", \"value\") error")
				return interfaces.MetricModelUniResponse{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					uerrors.Uniquery_MetricModel_InternalError).
					WithErrorDetails("root.GetByPath(\"hits\", \"total\", \"value\") error")
			}
			newAggrNode, _ := sonic.GetFromString(fmt.Sprintf(`{"doc_count": %d}`, docCount))
			aggrNode = &newAggrNode
		} else {
			o11y.Error(ctx, "root.Get(aggregations) error")
			return interfaces.MetricModelUniResponse{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_MetricModel_InternalError).WithErrorDetails("root.Get(aggregations) error")
		}
	}

	labels := make(map[string]string)
	datas := make([]interfaces.MetricModelData, 0)

	// return mapResult
	err = TraversalBucket(aggrNode, query, 0, labels, &datas)
	if err != nil {
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("IteratorAggs2 error: %s", err.Error()))

		return interfaces.MetricModelUniResponse{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_MetricModel_InternalError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	res := interfaces.MetricModelUniResponse{
		Datas:      datas,
		Step:       query.StepStr,
		IsVariable: query.IsVariable,
		IsCalendar: query.IsCalendar,
	}
	if query.IncludeModel {
		res.Model = model
	}
	return processOrderHaving(res, query, model), nil
}

// 递归各层聚合结果，把 dsl 的查询结果组装成统一格式
func TraversalBucket(aggrNode *ast.Node, query interfaces.MetricModelQuery,
	bktIdx int, labels map[string]string, datas *[]interfaces.MetricModelData) error {

	dslConfig := query.FormulaConfig.(interfaces.MetricModelFormulaConfig)
	bktInfos := dslConfig.Buckets
	if bktIdx >= len(bktInfos) {
		if query.IsInstantQuery {
			return TraversalAggregation(aggrNode, query, labels, datas)
		} else {
			return TraversalDateHistogram(aggrNode, query, labels, datas)
		}
	}

	bucketCfg := bktInfos[bktIdx]
	// 取桶
	switch bucketCfg.Type {
	case interfaces.BUCKET_TYPE_TERMS, interfaces.BUCKET_TYPE_RANGE,
		interfaces.BUCKET_TYPE_GEOHASH_GRID, interfaces.BUCKET_TYPE_DATE_RANGE:

		bktNodes, err := aggrNode.GetByPath(bucketCfg.BktName, "buckets").ArrayUseNode()
		if err != nil {
			return err
		}

		for _, bktNode := range bktNodes {
			bktKey, err := bktNode.Get("key").String()
			if err != nil {
				return err
			}
			labels[bucketCfg.Name] = bktKey
			// terms 下得有 date_histogram（因为当前 dsl 作为时序分析，校验时判断了必须含有 date_histogram）
			err = TraversalBucket(&bktNode, query, bktIdx+1, labels, datas)
			if err != nil {
				return err
			}
		}

	case interfaces.BUCKET_TYPE_FILTERS:
		bktNodes, err := aggrNode.GetByPath(bucketCfg.Name, "buckets").MapUseNode()
		if err != nil {
			return err
		}

		for bktKey, bktNode := range bktNodes {
			labels[bucketCfg.Name] = bktKey
			err := TraversalBucket(&bktNode, query, bktIdx+1, labels, datas)
			if err != nil {
				return err
			}
		}

		// case interfaces.MULTI_TERMS:
		//do nothing

	}

	return nil
}

// 递归各层聚合结果，把 dsl 的查询结果组装成统一格式
func TraversalDateHistogram(aggrNode *ast.Node, query interfaces.MetricModelQuery,
	labels map[string]string, datas *[]interfaces.MetricModelData) error {

	bktNodes, err := aggrNode.GetByPath(interfaces.DATE_HISTOGRAM_FIELD, "buckets").ArrayUseNode()
	if err != nil {
		return err
	}

	if len(bktNodes) == 0 {
		return nil
	}

	dateV1, err := bktNodes[0].Get("key").Int64()
	if err != nil {
		return err
	}
	start := min(dateV1, query.FixedStart)

	dateV2, err := bktNodes[len(bktNodes)-1].Get("key").Int64()
	if err != nil {
		return err
	}
	end := max(dateV2, query.FixedEnd)

	arrLen := getPointTimeIndex(query, start, end) + 1
	if arrLen > interfaces.MAX_DATE_HISTOGRAM_BUCKET_SIZE {
		return errors.New("too many buckets in current date histogram aggreation")
	}

	dateValues := make([]interface{}, arrLen)
	currentTime := start
	for i := range dateValues {
		dateValues[i] = currentTime
		currentTime = getNextPointTime(query, currentTime)
	}
	aggrInfo := query.FormulaConfig.(interfaces.MetricModelFormulaConfig).Aggregation
	switch aggrInfo.Type {
	case interfaces.AGGR_TYPE_DOC_COUNT:
		values := make([]interface{}, arrLen)
		for i := range values {
			values[i] = nil
		}
		for _, bktNode := range bktNodes {
			dateV, err := bktNode.Get("key").Int64()
			if err != nil {
				return err
			}

			value, err := bktNode.Get("doc_count").Float64()
			if err != nil {
				return err
			}

			arrIdx := getPointTimeIndex(query, start, dateV)
			values[arrIdx] = convert.WrapMetricValue(value)
		}

		labelsClone := common.CloneStringMap(labels)
		*datas = append(*datas, interfaces.MetricModelData{
			Labels: labelsClone,
			Times:  dateValues,
			Values: values,
		})

	case interfaces.AGGR_TYPE_VALUE_COUNT, interfaces.AGGR_TYPE_CARDINALITY,
		interfaces.AGGR_TYPE_SUM, interfaces.AGGR_TYPE_AVG,
		interfaces.AGGR_TYPE_MAX, interfaces.AGGR_TYPE_MIN:

		values := make([]interface{}, arrLen)
		for i := range values {
			values[i] = nil
		}
		for _, bktNode := range bktNodes {
			dateV, err := bktNode.Get("key").Int64()
			if err != nil {
				return err
			}

			value, err := bktNode.GetByPath(interfaces.VALUE_FIELD, "value").Float64()
			if err != nil {
				return err
			}

			arrIdx := getPointTimeIndex(query, start, dateV)
			values[arrIdx] = convert.WrapMetricValue(value)
		}

		labelsClone := common.CloneStringMap(labels)
		*datas = append(*datas, interfaces.MetricModelData{
			Labels: labelsClone,
			Times:  dateValues,
			Values: values,
		})

	case interfaces.AGGR_TYPE_PERCENTILES:
		valuesNode, err := bktNodes[0].GetByPath(interfaces.VALUE_FIELD, "values").MapUseNode()
		if err != nil {
			return err
		}

		for pKey := range valuesNode {
			values := make([]interface{}, arrLen)
			for i := range values {
				values[i] = nil
			}

			for _, bktNode := range bktNodes {
				dateV, err := bktNode.Get("key").Int64()
				if err != nil {
					return err
				}

				value, err := bktNode.GetByPath(interfaces.VALUE_FIELD, "values", pKey).Float64()
				if err != nil {
					return err
				}

				arrIdx := getPointTimeIndex(query, start, dateV)
				values[arrIdx] = convert.WrapMetricValue(value)
			}

			// 构造 datas，labels是一个map，map 是引用，构造得时候需要把map重新赋值
			labelsClone := common.CloneStringMap(labels)
			labelsClone[interfaces.PERCENT_FIELD] = pKey

			*datas = append(*datas, interfaces.MetricModelData{
				Labels: labelsClone,
				Times:  dateValues,
				Values: values,
			})
		}
	}

	return nil
}

// 递归各层聚合结果，把 dsl 的查询结果组装成统一格式
func TraversalAggregation(aggrNode *ast.Node, query interfaces.MetricModelQuery,
	labels map[string]string, datas *[]interfaces.MetricModelData) error {

	aggrInfo := query.FormulaConfig.(interfaces.MetricModelFormulaConfig).Aggregation
	switch aggrInfo.Type {
	case interfaces.AGGR_TYPE_DOC_COUNT:
		value, err := aggrNode.Get("doc_count").Float64()
		if err != nil {
			return err
		}

		// 构造 datas，labels是一个map，map 是引用，构造得时候需要把map重新赋值
		labelsClone := common.CloneStringMap(labels)
		*datas = append(*datas, interfaces.MetricModelData{
			Labels: labelsClone,
			Times:  []interface{}{query.Time},
			Values: []interface{}{convert.WrapMetricValue(value)},
		})

	case interfaces.AGGR_TYPE_VALUE_COUNT, interfaces.AGGR_TYPE_CARDINALITY,
		interfaces.AGGR_TYPE_SUM, interfaces.AGGR_TYPE_AVG,
		interfaces.AGGR_TYPE_MAX, interfaces.AGGR_TYPE_MIN:

		value, err := aggrNode.GetByPath(interfaces.VALUE_FIELD, "value").Float64()
		if err != nil {
			return err
		}
		// 构造 datas，labels是一个map，map 是引用，构造得时候需要把map重新赋值
		labelsClone := common.CloneStringMap(labels)
		*datas = append(*datas, interfaces.MetricModelData{
			Labels: labelsClone,
			Times:  []interface{}{query.Time},
			Values: []interface{}{convert.WrapMetricValue(value)},
		})

	case interfaces.AGGR_TYPE_PERCENTILES:
		valuesNodes, err := aggrNode.GetByPath(interfaces.VALUE_FIELD, "values").MapUseNode()
		if err != nil {
			return err
		}

		for pKey, pValueNode := range valuesNodes {
			value, err := pValueNode.Float64()
			if err != nil {
				return err
			}

			// 构造 datas，labels是一个map，map 是引用，构造得时候需要把map重新赋值
			labelsClone := common.CloneStringMap(labels)
			labelsClone[interfaces.PERCENT_FIELD] = pKey

			*datas = append(*datas, interfaces.MetricModelData{
				Labels: labelsClone,
				Times:  []interface{}{query.Time},
				Values: []interface{}{convert.WrapMetricValue(value)},
			})
		}
	}

	return nil
}

// getPointTimeIndex 获取下一个时间点
func getPointTimeIndex(query interfaces.MetricModelQuery, startTime int64, currentTime int64) int {

	if query.IsCalendar {
		// 将时间戳转换为时间对象
		switch *query.StepStr {
		case interfaces.CALENDAR_STEP_MINUTE:
			return int((currentTime - startTime) / time.Minute.Milliseconds())
		case interfaces.CALENDAR_STEP_HOUR:
			return int((currentTime - startTime) / time.Hour.Milliseconds())
		case interfaces.CALENDAR_STEP_DAY:
			return int((currentTime - startTime) / (time.Hour * 24).Milliseconds())
		case interfaces.CALENDAR_STEP_WEEK:
			return int((currentTime - startTime) / (time.Hour * 24 * 7).Milliseconds())
		case interfaces.CALENDAR_STEP_MONTH:
			t1 := time.UnixMilli(startTime)
			t2 := time.UnixMilli(currentTime)
			return calculateMonthDifference(t1, t2)
		case interfaces.CALENDAR_STEP_QUARTER:
			t1 := time.UnixMilli(startTime)
			t2 := time.UnixMilli(currentTime)
			return calculateMonthDifference(t1, t2) / 3
		case interfaces.CALENDAR_STEP_YEAR:
			y1 := time.UnixMilli(startTime).Year()
			y2 := time.UnixMilli(currentTime).Year()
			return (y2 - y1)
		}
	} else {
		return int((currentTime - startTime) / *query.Step)
	}

	return 0
}

func calculateMonthDifference(t1, t2 time.Time) int {
	// 获取两个时间的年份和月份
	y1, m1, _ := t1.Date()
	y2, m2, _ := t2.Date()

	// 计算年份和月份差
	years := y2 - y1
	months := int(m2) - int(m1)

	// 调整月份差，确保结果总是正数
	if months < 0 {
		years--
		months += 12
	}

	// 返回年份和月份差的总和
	return years*12 + months
}

// 转换 FieldName
func wrapFieldName(field *cond.ViewField) string {
	name := field.Name
	if field.Type == dtype.DataType_Text {
		name = field.Name + "." + dtype.KEYWORD_SUFFIX
	}
	return name
}
