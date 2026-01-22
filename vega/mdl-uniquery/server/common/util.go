package common

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	"github.com/bytedance/sonic"
	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/modern-go/reflect2"
)

// 创建一个局部的 API 实例，不影响全局 jsoniter.ConfigDefault
var jsonExt = jsoniter.ConfigCompatibleWithStandardLibrary

type intToStringExtension struct {
	jsoniter.DummyExtension
}

func init() {
	// 仅在局部实例中注册扩展
	jsonExt.RegisterExtension(&intToStringExtension{})
}

// 为 int64 和 json.Number 类型指定自定义编码器
func (extension *intToStringExtension) CreateEncoder(typ reflect2.Type) jsoniter.ValEncoder {
	if typ.Kind() == reflect.Int64 {
		return &int64AsStringEncoder{}
	}

	if typ.Kind() == reflect.String {
		return &jsonNumberAsStringEncoder{}
	}

	return nil
}

type int64AsStringEncoder struct{}

func (enc *int64AsStringEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	val := *((*int64)(ptr))
	stream.WriteString(strconv.FormatInt(val, 10))
}

func (enc *int64AsStringEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*int64)(ptr)) == 0
}

type jsonNumberAsStringEncoder struct{}

func (enc *jsonNumberAsStringEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	val := *((*json.Number)(ptr))
	stream.WriteString(val.String())
}

func (enc *jsonNumberAsStringEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*json.Number)(ptr)) == ""
}

// 判断是否为浏览器请求
func isBrowser(userAgent string) bool {
	return strings.Contains(userAgent, "Mozilla") ||
		strings.Contains(userAgent, "Chrome") ||
		strings.Contains(userAgent, "Safari")
}

// 如果浏览器请求，使用jsoniter进行编码, 否则使用sonic进行编码
func Marshal(userAgent string, v interface{}) ([]byte, error) {
	if isBrowser(userAgent) {
		return jsonExt.Marshal(v)
	}
	return sonic.Marshal(v)
}

// CE 条件表达式函数（泛型版本）
// condition: 条件表达式
// trueVal: 条件为真时返回的值
// falseVal: 条件为假时返回的值
func CE[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

// 给字符串加双引号
func QuotationMark(s string) string {
	if strings.HasPrefix(s, "\"") || strings.HasSuffix(s, "\"") { //防止拼接过情况
		return s
	}
	return "\"" + s + "\""
}

func GenerateUniqueKey(id string, label map[string]string) string {
	var keys []string
	for k := range label {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	key := id
	for _, k := range keys {
		key = key + "-" + k + ":" + label[k]
	}
	return key
}

func IsSlice(i any) bool {
	kind := reflect.ValueOf(i).Kind()
	return kind == reflect.Slice || kind == reflect.Array
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

func SplitString2InterfaceArray(s, sep string) []any {
	count := strings.Count(s, sep) + 1
	anys := make([]any, count)
	idx := 0
	for {
		idx = strings.Index(s, sep)
		if idx == -1 {
			anys[len(anys)-1] = s
			break
		}
		anys[len(anys)-count] = s[:idx]
		s = s[idx+len(sep):]
		count--
	}
	return anys

	// strs := strings.Split(s, sep)
	// anys := make([]any, len(strs))

	// for i, subStr := range strs {
	// 	anys[i] = subStr
	// }
	// return anys
}

func Any2String(val any) string {
	switch t := val.(type) {
	case string:
		return val.(string)
	case uint:
		it := val.(uint)
		return strconv.Itoa(int(it))
	case uint8:
		it := val.(uint8)
		return strconv.Itoa(int(it))
	case uint16:
		it := val.(uint16)
		return strconv.Itoa(int(it))
	case uint32:
		it := val.(uint32)
		return strconv.Itoa(int(it))
	case uint64:
		it := val.(uint64)
		return strconv.FormatUint(it, 10)
	case int:
		it := val.(int)
		return strconv.Itoa(it)
	case int8:
		it := val.(int8)
		return strconv.Itoa(int(it))
	case int16:
		it := val.(int16)
		return strconv.Itoa(int(it))
	case int32:
		it := val.(int32)
		return strconv.Itoa(int(it))
	case int64:
		it := val.(int64)
		return strconv.FormatInt(it, 10)
	case float32:
		ft := val.(float32)
		return strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case float64:
		ft := val.(float64)
		return strconv.FormatFloat(ft, 'f', -1, 64)
	case []byte:
		return string(val.([]byte))
	default:
		logger.Warnf("unspported interface dynamic type: %v", t)
		return ""
	}
}

func CloneStringMap(originalMap map[string]string) map[string]string {
	newMap := make(map[string]string)
	for k, v := range originalMap {
		newMap[k] = v
	}
	return newMap
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
