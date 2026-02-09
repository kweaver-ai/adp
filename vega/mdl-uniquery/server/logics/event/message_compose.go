// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package event

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"uniquery/common/compute"
	"uniquery/interfaces"
)

func flatteFieldMap(dataSourceType string, element map[string]any) string {
	var flatteFieldMapStr []string
	for key, value := range element {
		if v, ok := value.(float64); ok {
			if dataSourceType == "metric_model" {
				flatteFieldMapStr = append(flatteFieldMapStr, fmt.Sprintf(`当前值为'%.2f'`, v))
			} else {
				flatteFieldMapStr = append(flatteFieldMapStr, fmt.Sprintf(`'%s':'%.2f'`, key, v))
			}
		} else {
			flatteFieldMapStr = append(flatteFieldMapStr, fmt.Sprintf(`'%s':'%v'`, key, value))
		}
	}
	return strings.Join(flatteFieldMapStr, ",")
}

func flatteFilters(f interfaces.LogicFilter) string {
	var aggreFilter []string
	if len(f.Children) > 0 && (f.LogicOperator == "and" || f.LogicOperator == "or") {
		if f.LogicOperator == "and" {
			for _, filter := range f.Children {
				filterStr := flatteFilters(filter)
				aggreFilter = append(aggreFilter, filterStr)
			}
			return "(" + strings.Join(aggreFilter, " and ") + ")"
		} else {
			for _, filter := range f.Children {
				filterStr := flatteFilters(filter)
				aggreFilter = append(aggreFilter, filterStr)
			}
			return "(" + strings.Join(aggreFilter, " or ") + ")"
		}

	} else {
		return fmt.Sprintf(`('%s' %s '%v')`, f.FilterExpress.Name, f.FilterExpress.Operation, f.FilterExpress.Value)
	}
}
func GetKeysFromFilters(f interfaces.LogicFilter) []string {
	var keys []string
	if len(f.Children) > 0 && (f.LogicOperator == "and" || f.LogicOperator == "or") {
		for _, filter := range f.Children {
			childKeys := GetKeysFromFilters(filter)
			keys = append(keys, childKeys...)
		}
	} else {
		keys = append(keys, f.FilterExpress.Name)
	}
	return keys
}

func RemoveDuplicates(elements []string) []string {
	uniqueElements := make(map[string]bool)
	for _, elem := range elements {
		uniqueElements[elem] = true
	}
	var result []string
	for key := range uniqueElements {
		result = append(result, key)
	}
	return result
}

func Compose(ctx context.Context, em interfaces.EventModel, f interfaces.FormulaItem, element map[string]any, record map[string]any) string {
	//NOTE： 构造事件描述信息
	language := rest.GetLanguageByCtx(ctx)
	zhMessage := ""
	enMessage := ""
	var DataSourceName []string
	if em.DataSourceName == nil {
		DataSourceName = em.DataSource
	} else {
		DataSourceName = em.DataSourceName
	}
	fieldMapStr := flatteFieldMap(em.DataSourceType, element)
	// filterStr := flatteFilters(f.Filter)
	keys := GetKeysFromFilters(f.Filter)
	uniKeys := RemoveDuplicates(keys)
	sort.Strings(uniKeys)
	keysStr := ""
	for _, key := range uniKeys {
		if keysStr == "" {
			keysStr = key
		} else {
			keysStr = keysStr + "," + key
		}
	}

	labels := ""
	var labelsKey []string
	for key := range record {
		if strings.HasPrefix(key, "labels.") {
			labelsKey = append(labelsKey, key)
		}
	}
	sort.Strings(labelsKey)
	for _, key := range labelsKey {
		if labels == "" {
			labels = record[key].(string)
		} else {
			labels = labels + "," + record[key].(string)
		}
	}

	if f.Level == interfaces.EVENT_MODEL_LEVEL_CLEARED {
		if labels == "" {
			if em.DataSourceType == "metric_model" {
				zhMessage = fmt.Sprintf("监控对象(未知)的监控项(%s)已经恢复正常(%s)", DataSourceName, fieldMapStr)
				enMessage = fmt.Sprintf("The monitored object (unknown) monitoring item(%s) has returned to normal(%s)", DataSourceName, fieldMapStr)
			} else if em.DataSourceType == "data_view" {
				zhMessage = fmt.Sprintf("监控对象(未知)的监控项(%s)已经恢复正常(%s)", keysStr, fieldMapStr)
				enMessage = fmt.Sprintf("The monitored object (unknown) monitoring item(%s) has returned to normal(%s)", keysStr, fieldMapStr)
			}
		} else {
			if em.DataSourceType == "metric_model" {
				zhMessage = fmt.Sprintf("监控对象(%s)的监控项(%s)已经恢复正常(%s)", labels, DataSourceName, fieldMapStr)
				enMessage = fmt.Sprintf("The monitored object (%s) monitoring item(%s) has returned to normal(%s)", labels, DataSourceName, fieldMapStr)
			} else if em.DataSourceType == "data_view" {
				zhMessage = fmt.Sprintf("监控对象(%s)的监控项(%s)已经恢复正常(%s)", labels, keysStr, fieldMapStr)
				enMessage = fmt.Sprintf("The monitored object (%s) monitoring item(%s) has returned to normal(%s)", labels, keysStr, fieldMapStr)
			}
		}
	} else {
		if labels == "" {
			if em.DataSourceType == "metric_model" {
				zhMessage = fmt.Sprintf("监控对象(未知)的监控项(%s)产生了异常(%s)", DataSourceName, fieldMapStr)
				enMessage = fmt.Sprintf("The monitored object (unknown) monitoring item(%s) is abnormal(%s)", DataSourceName, fieldMapStr)
			} else if em.DataSourceType == "data_view" {
				zhMessage = fmt.Sprintf("监控对象(未知)的监控项(%s)产生了异常(%s)", keysStr, fieldMapStr)
				enMessage = fmt.Sprintf("The monitored object (unknown) monitoring item(%s) is abnormal(%s)", keysStr, fieldMapStr)
			}

		} else {
			if em.DataSourceType == "metric_model" {
				zhMessage = fmt.Sprintf("监控对象(%s)的监控项(%s)产生了异常(%s)", labels, DataSourceName, fieldMapStr)
				enMessage = fmt.Sprintf("The monitored object (%s) monitoring item(%s) is abnormal(%s)", labels, DataSourceName, fieldMapStr)
			} else if em.DataSourceType == "data_view" {
				zhMessage = fmt.Sprintf("监控对象(%s)的监控项(%s)产生了异常(%s)", labels, keysStr, fieldMapStr)
				enMessage = fmt.Sprintf("The monitored object (%s) monitoring item(%s) is abnormal(%s)", labels, keysStr, fieldMapStr)
			}
		}
	}
	if language == "zh-CN" {
		return zhMessage
	} else {
		return enMessage
	}
}
func GenerateMessage(aggregateRuleType string, context interfaces.EventContext, groupFields []string, level string) string {
	if aggregateRuleType == "healthy_compute" {
		return fmt.Sprintf("%s,生成了一个等级为%s的聚合事件", GenerateBasicMessage(context, groupFields, level), level)
	}
	//分组字段改为","分隔
	groupFieldsStr := ""
	for _, v := range groupFields {
		if groupFieldsStr == "" {
			groupFieldsStr = v
		} else {
			groupFieldsStr = groupFieldsStr + "," + v
		}
	}
	return fmt.Sprintf("%s,分组字段为[%s],生成了一个等级为%s的聚合事件", GenerateBasicMessage(context, groupFields, level), groupFieldsStr, level)
}

func GenerateBasicMessage(context interfaces.EventContext, groupFields []string, level string) string {
	message := "基于"
	cnt := make([]int, 6)
	for _, record := range context.SourceRecords {
		var index int
		if v, ok := record["level"].(string); ok {
			index, _ = strconv.Atoi(v)
		} else {
			index = record["level"].(int)
		}
		cnt[index-1]++
	}
	// for _, level := range []string{"紧急", "主要", "次要", "提示", "不明确", "清除"} {
	for index, level := range []string{
		interfaces.EVENT_MODEL_LEVEL_ZH_CN[interfaces.EVENT_MODEL_LEVEL_CRITICAL],
		interfaces.EVENT_MODEL_LEVEL_ZH_CN[interfaces.EVENT_MODEL_LEVEL_MAJOR],
		interfaces.EVENT_MODEL_LEVEL_ZH_CN[interfaces.EVENT_MODEL_LEVEL_MINOR],
		interfaces.EVENT_MODEL_LEVEL_ZH_CN[interfaces.EVENT_MODEL_LEVEL_WARNING],
		interfaces.EVENT_MODEL_LEVEL_ZH_CN[interfaces.EVENT_MODEL_LEVEL_INDETERMINATE],
		interfaces.EVENT_MODEL_LEVEL_ZH_CN[interfaces.EVENT_MODEL_LEVEL_CLEARED]} {
		if cnt[index] > 0 {
			message += fmt.Sprintf("%d个%s事件 ", cnt[index], level)
		}
	}
	return message
}

func Combine(hits interfaces.Records, level int, score float64) interfaces.EventContext {
	return interfaces.EventContext{
		SourceRecords: hits,
		Level:         level,
		Score:         score,
		GroupFields:   []string{},
	}
}

func GroupCombine(hits interfaces.Records, group_fields []string) interfaces.EventContext {
	var LevelMapSet = map[string]interfaces.Records{
		"Critical":      {},
		"Major":         {},
		"Minor":         {},
		"Warning":       {},
		"Indeterminate": {},
		"Cleared":       {},
	}
	for _, s := range hits {
		var level_id int
		level_id, ok := s["level"].(int)
		if !ok {
			level_str, _ := s["level"].(string)
			level_id, _ = strconv.Atoi(level_str)
		}

		level := compute.EventModelMap[level_id]
		LevelMapSet[level] = append(LevelMapSet[level], s)
	}
	var hits_level int
	var hits_score float64
	for _, level := range []string{"Critical", "Major", "Minor", "Warning", "Indeterminate", "Cleared"} {
		if len(LevelMapSet[level]) > 0 {

			hits_level = compute.LevelMap[level].Id
			hits_score = math.Max(compute.LevelMap[level].UpperBound-float64(len(LevelMapSet[level])), compute.LevelMap[level].LowerBound)
			break
		}

	}
	return interfaces.EventContext{
		SourceRecords: hits,
		Level:         hits_level,
		Score:         hits_score,
		GroupFields:   group_fields,
	}
}
