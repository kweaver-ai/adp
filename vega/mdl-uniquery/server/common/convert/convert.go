// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package convert

import (
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
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"

	"uniquery/common"
	"uniquery/interfaces"
)

const (
	// The largest SampleValue that can be converted to an int64 without overflow.
	maxInt64 = 9223372036854774784 // 0x7FFF FFFF FFFF FC00

	// The smallest SampleValue that can be converted to an int64 without underflow.
	minInt64 = -9223372036854775808 // 0x8000 0000 0000 0000

	DEFAULT_LOOK_BACK_DELTA = time.Minute * time.Duration(5)
	// 特殊数值
	POS_INF = "+Inf"
	NEG_INF = "-Inf"
	NaN     = "NaN"
)

var (
	MinTime = time.Unix(math.MinInt64/1000+62135596801, 0).UTC()
	MaxTime = time.Unix(math.MaxInt64/1000-62135596801, 999999999).UTC()

	minTimeFormatted = MinTime.Format(time.RFC3339Nano)
	maxTimeFormatted = MaxTime.Format(time.RFC3339Nano)

	//durationRE = regexp.MustCompile("^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$")
)

func JsonToMap(jsonStr string) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	err := sonic.Unmarshal([]byte(jsonStr), &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func IntToDuration(str string) (time.Duration, error) {
	if len(str) < 2 {
		return 0, errors.New("str to time.duration parse failed due to malformed parameter")
	}

	var index int
	reg, _ := regexp.Compile(`^[\+-]?\d+$`)
	for i := 0; i < len(str); i++ {
		ch := string(str[i])
		match := reg.MatchString(ch)
		if match {
			continue
		} else {
			index = i
		}
	}

	vInt, err := strconv.Atoi(str[:index])
	if err != nil {
		return 0, err
	}

	unit := str[index:]
	match, _ := regexp.MatchString(`[s,m,h]`, unit)
	if !match || len(unit) != 1 {
		return 0, errors.New("str to time.duration parse error,the unit must be 's','m','h'")
	}

	var timeOutDuration time.Duration
	switch unit {
	case "s":
		timeOutDuration = time.Duration(vInt) * time.Second
	case "h":
		timeOutDuration = time.Duration(vInt) * time.Hour
	default:
		timeOutDuration = time.Duration(vInt) * time.Minute
	}
	return timeOutDuration, nil
}

func InterToArray(v interface{}) []string {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()
	var array []string
	value := reflect.ValueOf(v)
	arrayInter := value.Interface().([]interface{})
	if len(arrayInter) == 0 {
		return []string{}
	}
	for _, val := range arrayInter {
		value = reflect.ValueOf(val)
		str := value.Interface().(string)
		array = append(array, str)
	}
	return array
}

func IntersectArray(slice1 []string, slice2 []string) []string {
	m := make(map[string]int)
	nn := make([]string, 0)
	for _, v := range slice1 {
		m[v]++
	}

	for _, v := range slice2 {
		times := m[v]
		if times == 1 {
			nn = append(nn, v)
		}
	}
	return nn
}

func StructConvertMap(obj interface{}, tagName string) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		tagName := t.Field(i).Tag.Get(tagName)
		fmt.Println(tagName)
		if tagName != "" && tagName != "-" {
			data[tagName] = v.Field(i).Interface()
		}
	}
	return data
}

func MapToByte(m map[string]interface{}) []byte {
	temp, _ := sonic.Marshal(m)
	return temp
}

func ParseTimeParam(val string, paramName string, defaultValue time.Time) (time.Time, error) {
	if val == "" {
		return defaultValue, nil
	}
	result, err := ParseTime(val)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time value for '%s': %v", paramName, err)
	}
	return result, nil
}

func ParseTime(s string) (time.Time, error) {
	if t, err := strconv.ParseFloat(s, 64); err == nil {
		s, ns := math.Modf(t)
		ns = math.Round(ns*1000) / 1000
		return time.Unix(int64(s), int64(ns*float64(time.Second))).UTC(), nil
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}

	// Stdlib's time parser can only handle 4 digit years. As a workaround until
	// that is fixed we want to at least support our own boundary times.
	// Context: https://github.com/prometheus/client_golang/issues/614
	// Upstream issue: https://github.com/golang/go/issues/20555
	switch s {
	case minTimeFormatted:
		return MinTime, nil
	case maxTimeFormatted:
		return MaxTime, nil
	}
	return time.Time{}, fmt.Errorf("cannot parse %q to a valid timestamp", s)
}

func isdigit(c byte) bool { return c >= '0' && c <= '9' }

// Units are required to go in order from biggest to smallest.
// This guards against confusion from "1m1d" being 1 minute + 1 day, not 1 month + 1 day.
var unitMap = map[string]struct {
	pos  int
	mult uint64
}{
	"ms": {7, uint64(time.Millisecond)},
	"s":  {6, uint64(time.Second)},
	"m":  {5, uint64(time.Minute)},
	"h":  {4, uint64(time.Hour)},
	"d":  {3, uint64(24 * time.Hour)},
	"w":  {2, uint64(7 * 24 * time.Hour)},
	"y":  {1, uint64(365 * 24 * time.Hour)},
}

// ParseDurationMilli parses a string into a millisecod,
// assuming that a year always has 365d, a week always has 7d, and a day always has 24h.
func ParseDuration(s string) (time.Duration, error) {
	switch s {
	case "0":
		// Allow 0 without a unit.
		return 0, nil
	case "":
		return 0, errors.New("empty duration string")
	}

	orig := s
	var dur uint64
	lastUnitPos := 0

	for s != "" {
		if !isdigit(s[0]) {
			return 0, fmt.Errorf("not a valid duration string: %q", orig)
		}
		// Consume [0-9]*
		i := 0
		for ; i < len(s) && isdigit(s[i]); i++ {
		}
		v, err := strconv.ParseUint(s[:i], 10, 0)
		if err != nil {
			return 0, fmt.Errorf("not a valid duration string: %q", orig)
		}
		s = s[i:]

		// Consume unit.
		for i = 0; i < len(s) && !isdigit(s[i]); i++ {
		}
		if i == 0 {
			return 0, fmt.Errorf("not a valid duration string: %q", orig)
		}
		u := s[:i]
		s = s[i:]
		unit, ok := unitMap[u]
		if !ok {
			return 0, fmt.Errorf("unknown unit %q in duration %q", u, orig)
		}
		if unit.pos <= lastUnitPos { // Units must go in order from biggest to smallest.
			return 0, fmt.Errorf("not a valid duration string: %q", orig)
		}
		lastUnitPos = unit.pos
		// Check if the provided duration overflows time.Duration (> ~ 290years).
		if v > 1<<63/unit.mult {
			return 0, errors.New("duration out of range")
		}
		dur += v * unit.mult
		if dur > 1<<63-1 {
			return 0, errors.New("duration out of range")
		}
	}
	return time.Duration(dur), nil
}

// FromFloatSeconds returns a millisecond timestamp from float seconds.
func FromFloatSeconds(ts float64) int64 {
	return int64(math.Round(ts * 1000))
}

// FromTime returns a new millisecond timestamp from a time.
func FromTime(t time.Time) int64 {
	return t.Unix()*1000 + int64(t.Nanosecond())/int64(time.Millisecond)
}

// convertibleToInt64 returns true if v does not over-/underflow an int64.
func ConvertibleToInt64(v float64) bool {
	return v <= maxInt64 && v >= minInt64
}

func RFC3339ToMicroTimestamp(timeStrs []string) ([]int64, error) {
	microTimestamps := make([]int64, len(timeStrs))
	for index, timeStr := range timeStrs {
		realTime, err := time.Parse(time.RFC3339Nano, timeStr)
		if err != nil {
			return nil, err
		}
		microTimestamps[index] = realTime.UnixMicro()
	}
	return microTimestamps, nil
}

// 获取瞬时查询的回退时间
func GetLookBackDelta(queryDelta int64, configDelta time.Duration) int64 {
	// 优先取查询参数中的回退时间，参数中没有取配置的回退时间，回退时间没有取默认的 5min
	lookBackDelta := queryDelta
	if lookBackDelta == 0 {
		lookBackDelta = configDelta.Milliseconds()
		if lookBackDelta == 0 {
			lookBackDelta = DEFAULT_LOOK_BACK_DELTA.Milliseconds()
		}
	}
	return lookBackDelta
}

// string 转 []string
func StringToStringSlice(str string) []string {
	strSlice := []string{}
	str = strings.Trim(str, "{} <>")
	if str == "" {
		return strSlice
	}

	strs := strings.Split(str, ",")
	for _, v := range strs {
		v = strings.Trim(v, "{} <>")
		if v != "" {
			strSlice = append(strSlice, v)
		}
	}
	return strSlice
}

// string 转 []uint64
func StringToInt64Slice(str string) ([]uint64, error) {

	if str == "" {
		return []uint64{}, nil
	} else {
		//分割
		resStr := strings.Split(str, ",")

		//转换
		res, err := StringSliceToInt64Slice(resStr)
		if err != nil {
			return res, err
		}

		return res, nil
	}
}

// []string 转 []uint64
func StringSliceToInt64Slice(strArr []string) ([]uint64, error) {
	res := make([]uint64, len(strArr))
	var err error
	for index, val := range strArr {
		//去空格
		val = strings.Trim(val, " ")

		res[index], err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			return []uint64{}, fmt.Errorf("[]string %v to []uint64 parse failed", strArr)
		}
	}
	return res, nil
}

// 封装指标值
func WrapMetricValue(v float64) interface{} {
	if math.IsInf(v, 1) {
		return POS_INF
	} else if math.IsInf(v, -1) {
		return NEG_INF
	} else if math.IsNaN(v) {
		return NaN
	} else {
		return v
	}
}

// hasOverlappingRanges 判断数组中的区间是否存在重叠
func HasOverlappingRanges(ranges []interfaces.Range) bool {
	if len(ranges) < 2 {
		return false
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
	return *i1.To > *i2.From
}

// findInterval 查找值所属的区间
func FindRange(value float64, ranges []interfaces.Range) (int, *interfaces.Range) {
	for i, interval := range ranges {
		// 检查值是否在区间内（包括边界）
		if value >= *interval.From && value < *interval.To {
			return i, &interval
		}
	}
	// 如果没有找到所属的区间，返回nil
	return 0, nil
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

// union 函数返回多个 map 的并集
func Union(maps ...map[string]bool) map[string]bool {
	result := make(map[string]bool)

	// 遍历所有 map
	for _, m := range maps {
		// 将当前 map 的键值对添加到结果集中
		for key, value := range m {
			if key != interfaces.ALL_LABELS_FLAG {
				result[key] = value
			}
		}
	}

	return result
}

func DurationMilliseconds(d time.Duration) int64 {
	return int64(d / (time.Millisecond / time.Nanosecond))
}

type SortKey struct {
	Key       string
	Ascending bool
}

// 去重合并两个map数组
func MergeAndDeduplicate(a, b []map[string]string, sortKeys []SortKey) []map[string]string {
	// 合并两个切片
	merged := append(a, b...)
	// 去重
	unique := make([]map[string]string, 0)
	// 创建一个map来跟踪已经存在的元素
	seen := make(map[string]bool)
	for _, m := range merged {
		key := getMapKey(m)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, m)
		}
	}
	// 排序
	sort.Slice(unique, func(i, j int) bool {
		for _, sk := range sortKeys {
			vi := unique[i][sk.Key]
			vj := unique[j][sk.Key]

			if vi != vj {
				if sk.Ascending {
					return vi < vj
				}
				return vi > vj
			}
		}
		return false
	})

	return unique
}

// getMapKey 生成map的唯一标识字符串
func getMapKey(m map[string]string) string {
	// 如果map为空，返回空字符串
	if len(m) == 0 {
		return ""
	}

	// 使用反射来保证顺序一致性
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	// 对key进行排序以保证一致性
	sort.Strings(keys)

	var keyStr string
	for _, k := range keys {
		keyStr += fmt.Sprintf("%s:%s|", k, m[k])
	}

	return keyStr
}

func GetIndexBasePattern(baseTypes []string) []string {
	indexPattern := make([]string, 0)
	for _, index := range baseTypes {
		indexPattern = append(indexPattern, "mdl-"+index+"-*")
	}
	return indexPattern
}

// 将日期时间字符串转换为毫秒时间戳
func ParseTimeToMillis(timeStr string, formatType string) (int64, error) {
	// 预处理字符串，去除多余空格
	timeStr = strings.TrimSpace(timeStr)

	var layout string
	// 根据格式类型选择解析方式
	switch formatType {
	case interfaces.CALENDAR_STEP_MINUTE:
		layout = "2006-01-02 15:04"
	case interfaces.CALENDAR_STEP_HOUR:
		layout = "2006-01-02 15"
	case interfaces.CALENDAR_STEP_DAY:
		layout = "2006-01-02"
	case interfaces.CALENDAR_STEP_WEEK:
		// 处理年周格式 2025-46
		parts := strings.Split(timeStr, "-")
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid year-week format: %s", timeStr)
		}
		year, _ := strconv.ParseInt(parts[0], 10, 64)
		week, _ := strconv.ParseInt(parts[1], 10, 64)
		// 获取该年的第1周的周一
		firstWeekMonday := time.Date(int(year), time.January, 1, 0, 0, 0, 0, common.APP_LOCATION)
		for firstWeekMonday.Weekday() != time.Monday {
			firstWeekMonday = firstWeekMonday.AddDate(0, 0, 1)
		}
		// 计算第week周的周一
		weekNMonday := firstWeekMonday.AddDate(0, 0, (int(week)-1)*7)
		return weekNMonday.UnixMilli(), nil
	case interfaces.CALENDAR_STEP_MONTH:
		layout = "2006-01"
	case interfaces.CALENDAR_STEP_QUARTER:
		// 处理年季度格式 2025-Q1
		parts := strings.Split(timeStr, "-Q")
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid year-quarter format: %s", timeStr)
		}
		year := parts[0]
		quarter := parts[1]
		// 构造为季度的第一天
		var month string
		switch quarter {
		case "1":
			month = "01"
		case "2":
			month = "04"
		case "3":
			month = "07"
		case "4":
			month = "10"
		default:
			return 0, fmt.Errorf("invalid quarter: %s", quarter)
		}
		timeStr = year + "-" + month + "-01"
		layout = "2006-01-02"
	case interfaces.CALENDAR_STEP_YEAR:
		layout = "2006"
	default:
		return 0, fmt.Errorf("unsupported format type: %s", formatType)
	}

	t, err := time.ParseInLocation(layout, timeStr, common.APP_LOCATION)
	if err != nil {
		return 0, fmt.Errorf("failed to parse time: %v", err)
	}
	// 转换为毫秒时间戳
	return t.UnixMilli(), nil
}

func FormatTimeMiliis(ts int64, formatType string) string {
	t := time.UnixMilli(ts).In(common.APP_LOCATION)
	switch formatType {
	case interfaces.CALENDAR_STEP_MINUTE:
		return t.Format("2006-01-02 15:04")
	case interfaces.CALENDAR_STEP_HOUR:
		return t.Format("2006-01-02 15")
	case interfaces.CALENDAR_STEP_DAY:
		return t.Format("2006-01-02")
	case interfaces.CALENDAR_STEP_WEEK:
		// 周：年-周 (如 2025-46)
		year, week := t.ISOWeek()
		return fmt.Sprintf("%d-%02d", year, week)
	case interfaces.CALENDAR_STEP_MONTH:
		return t.Format("2006-01")
	case interfaces.CALENDAR_STEP_QUARTER:
		quarter := (t.Month()-1)/3 + 1
		return fmt.Sprintf("%d-Q%d", t.Year(), quarter)
	case interfaces.CALENDAR_STEP_YEAR:
		return t.Format("2006")
	default:
		return FormatRFC3339Milli(ts)
	}
}

// 按环境变量中的时区格式化
func FormatRFC3339Milli(timestamp int64) string {
	t := time.UnixMilli(timestamp)

	// 转换为指定时区并格式化
	return t.In(common.APP_LOCATION).Format(libCommon.RFC3339Milli)
}

// lastDayOfMonth 返回给定日期所在月份的最后一天
func LastDayOfMonth(t time.Time) time.Time {
	firstOfMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	nextMonth := firstOfMonth.AddDate(0, 1, 0)
	return nextMonth.AddDate(0, 0, -1)
}

// isLeap 判断是否是闰年
func IsLeap(year int) bool {
	return year%400 == 0 || (year%100 != 0 && year%4 == 0)
}

// 复合指标的计算公式中的 {{}} 转换为 metric_model 函数
func TransformExpression(input string) string {
	// 正则表达式匹配 {{任意内容}}，并捕获中间的内容
	re := regexp.MustCompile(`\{\{\s*(.*?)\s*\}\}`)

	// 替换函数，对每个匹配项进行处理
	replacer := func(match string) string {
		// 提取匹配组，去除前后空格
		modelID := re.ReplaceAllString(match, "$1")
		modelID = strings.TrimSpace(modelID)
		return fmt.Sprintf(`metric_model("%s")`, modelID)
	}

	// 执行替换
	return re.ReplaceAllStringFunc(input, replacer)
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
