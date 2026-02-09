// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"data-model/interfaces"
	dcond "data-model/interfaces/condition"
)

var (
	// 持久化任务的固定频率调度时间和时间窗口的正则匹配，支持单位有分钟，小时，天
	DurationDayHourMinuteRE = regexp.MustCompile("^(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?$")

	// 持久化任务的追溯时长的正则匹配，支持单位有小时，天
	DurationDayHourRE = regexp.MustCompile("^(([0-9]+)d)?(([0-9]+)h)?$")
)

// string 转 []string
func StringToStringSlice(str string) []string {
	if str == "" {
		return []string{}
	}

	strSlice := []string{}
	strs := strings.Split(str, ",")
	for _, v := range strs {
		v = strings.Trim(v, " ")
		if v != "" {
			strSlice = append(strSlice, v)
		}
	}
	return strSlice
}

// map 转 string
func Map2String(mapV map[string]string) (string, error) {
	ca := ""
	if mapV != nil {
		caStr, err := sonic.Marshal(mapV)
		if err != nil {
			return "", err
		}
		return string(caStr), nil
	}
	return ca, nil
}

// string 转 map
func String2Map(str string) (map[string]string, error) {
	var tempMap map[string]string
	if str != "" {
		if err := sonic.Unmarshal([]byte(str), &tempMap); err != nil {
			return tempMap, err
		}
		return tempMap, nil
	}
	return tempMap, nil
}

func ExtractTimeUnitFormString(s string) string {

	string_re := regexp.MustCompile("[a-zA-Z]+")
	unit := strings.Join(string_re.FindStringSubmatch(s), "")
	return unit
}

func ExtractTimeIntervalFormString(s string) int {
	number_re := regexp.MustCompile("^-?[0-9]+")
	n := number_re.FindStringSubmatch(s)[0]
	interval, _ := strconv.Atoi(n)
	return interval
}

func In(target string, str_array []string) bool {
	sort.Strings(str_array)
	index := sort.SearchStrings(str_array, target)
	if index < len(str_array) && str_array[index] == target {
		return true
	}
	return false
}

// 时间区间解析
func ParseDuration(s string, reg *regexp.Regexp, containMinute bool) (time.Duration, error) {
	if d, err := strconv.ParseFloat(s, 64); err == nil {
		ts := d * float64(time.Second)
		if ts > float64(math.MaxInt64) || ts < float64(math.MinInt64) {
			return 0, fmt.Errorf("cannot parse %q to a valid duration. It overflows int64", s)
		}
		return time.Duration(ts), nil
	}

	switch s {
	case "0":
		// Allow 0 without a unit.
		return 0, nil
	case "":
		return 0, fmt.Errorf("empty duration string")
	}

	matches := reg.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("not a valid duration string: %q", s)
	}
	var dur time.Duration

	// Parse the match at pos `pos` in the regex and use `mult` to turn that
	// into ms, then add that value to the total parsed duration.
	var overflowErr error
	m := func(pos int, mult time.Duration) {
		if matches[pos] == "" {
			return
		}
		n, _ := strconv.Atoi(matches[pos])

		// Check if the provided duration overflows time.Duration (> ~ 290years).
		if n > int((1<<63-1)/mult/time.Millisecond) {
			overflowErr = errors.New("duration out of range")
		}
		d := time.Duration(n) * time.Millisecond
		dur += d * mult

		if dur < 0 {
			overflowErr = errors.New("duration out of range")
		}
	}

	m(2, 1000*60*60*24) // d
	m(4, 1000*60*60)    // h
	if containMinute {
		m(6, 1000*60) // m
	}

	if overflowErr == nil {
		return time.Duration(dur), nil
	}
	return 0, fmt.Errorf("cannot parse %q to a valid duration", s)
}

func ConvertFiltersToCondition(filters []interfaces.Filter) (condition *interfaces.CondCfg) {
	if len(filters) == 0 {
		return
	} else if len(filters) == 1 {
		filter := filters[0]
		if filter.Operation == dcond.Operation_EQ {
			filter.Operation = dcond.OperationEq
		}

		condition = &interfaces.CondCfg{
			Operation: filter.Operation,
			Name:      filter.Name,
			ValueOptCfg: interfaces.ValueOptCfg{
				ValueFrom: dcond.ValueFrom_Const,
				Value:     filter.Value,
			},
		}
	} else {
		subConds := make([]*interfaces.CondCfg, 0, len(filters))
		for i := 0; i < len(filters); i++ {
			filter := filters[i]
			if filter.Operation == dcond.Operation_EQ {
				filter.Operation = dcond.OperationEq
			}

			subConds = append(subConds, &interfaces.CondCfg{
				Operation: filter.Operation,
				Name:      filter.Name,
				ValueOptCfg: interfaces.ValueOptCfg{
					ValueFrom: dcond.ValueFrom_Const,
					Value:     filter.Value,
				},
			})
		}

		condition = &interfaces.CondCfg{
			Operation: dcond.OperationAnd,
			SubConds:  subConds,
		}
	}

	return
}

// hasOverlappingRanges 判断数组中的区间是否存在重叠
func HasOverlappingRanges(ranges []interfaces.Range) bool {
	if len(ranges) < 2 {
		return false // 少于两个区间，不可能重叠
	}

	// 按照区间的起始点排序（如果起始点相同，则按照结束点排序）
	sort.Slice(ranges, func(i, j int) bool {
		if ranges[i].From == ranges[j].From {
			return *ranges[i].To < *ranges[j].To
		}
		return *ranges[i].From < *ranges[j].From
	})

	// 遍历排序后的区间，检查是否存在重叠
	for i := 1; i < len(ranges); i++ {
		if isOverlapping(ranges[i-1], ranges[i]) {
			return true
		}
	}

	return false
}

// isOverlapping 判断两个区间是否重叠
func isOverlapping(i1, i2 interfaces.Range) bool {
	return *i1.To > *i2.From // 发现重叠
}

// AssertFloat64 尝试将 interface{} 转换为 float64，如果转换失败则返回错误。
func AssertFloat64(v any) (float64, error) {
	switch v := v.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(v).Int()), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, errors.New("cannot assert type to float64")
	}
}

func ProcessUngroupedName(ctx context.Context, groupName string, objectName string) string {
	//NOTE: 返回不同语言的未分组
	lang := rest.GetLanguageByCtx(ctx)
	ungroupedName := interfaces.UNGROUPED_EN_US
	if lang == rest.SimplifiedChinese {
		ungroupedName = interfaces.UNGROUPED_ZH_CN
	}

	name := fmt.Sprintf("%s/%s", groupName, objectName)
	if groupName == "" {
		name = fmt.Sprintf("%s/%s", ungroupedName, objectName)
	}
	return name
}

// 将带引号的时间戳字符串转换为int64
// 支持单引号或双引号包裹，例如 "1672531200" 或 '1672531200000'
func QuotedTimestampToInt64(quotedStr string) (int64, error) {
	// 去除首尾的引号（支持单引号和双引号）
	trimmed := strings.Trim(quotedStr, `"'`)

	// 转换为int64
	timestamp, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("convert failed: %v", err)
	}
	return timestamp, nil
}

// 提取所有模型 ID，并去除 {{}} 内的前后空格
func ExtractModelIDs(input string) []string {
	// 正则表达式匹配 {{ 任意内容（非贪婪）}}，并捕获中间部分
	re := regexp.MustCompile(`\{\{\s*(.*?)\s*\}\}`)

	// 查找所有匹配项
	matches := re.FindAllStringSubmatch(input, -1)

	// 提取捕获组（即 {{}} 内的内容）
	var modelIDs []string
	for _, match := range matches {
		if len(match) >= 2 {
			// 去除捕获组的前后空格
			modelID := strings.TrimSpace(match[1])
			modelIDs = append(modelIDs, modelID)
		}
	}

	return modelIDs
}

// 对字符串数组去重
func DuplicateSlice(strSlice []string) []string {
	keys := make(map[string]struct{})
	list := make([]string, 0, len(strSlice))

	for _, item := range strSlice {
		if _, ok := keys[item]; !ok {
			keys[item] = struct{}{}
			list = append(list, item)
		}
	}
	return list
}

func IsSameType(arr []any) bool {
	if len(arr) == 0 {
		return true
	}

	firstType := reflect.TypeOf(arr[0])
	for _, v := range arr {
		if reflect.TypeOf(v) != firstType {
			return false
		}
	}

	return true
}
