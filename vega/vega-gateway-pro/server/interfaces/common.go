package interfaces

import (
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"math"
	"strings"
)

type contextKey string // 自定义专属的key类型

const (
	CONTENT_TYPE_NAME = "Content-Type"
	CONTENT_TYPE_JSON = "application/json"
	CONTENT_TYPE_FORM = "application/x-www-form-urlencoded"

	HTTP_HEADER_METHOD_OVERRIDE = "x-http-method-override"
	HTTP_HEADER_REQUEST_TOOK    = "x-request-took"
	HTTP_HEADER_ACCOUNT_ID      = "x-account-id"
	HTTP_HEADER_ACCOUNT_TYPE    = "x-account-type"
	HTTP_HEADER_AUTHORIZATION   = "Authorization"

	ACCOUNT_INFO_KEY contextKey = "x-account-info" // 避免直接使用string

	ADMIN_ID   string = "266c6a42-6131-4d62-8f39-853e7093701c"
	ADMIN_TYPE        = string(rest.VisitorType_User)

	DESC_DIRECTION    string = "desc"
	ASC_DIRECTION     string = "asc"
	DEFAULT_DIRECTION string = "desc"

	// 事件模型,指标模型持久化管道的topic
	MODEL_PERSIST_INPUT = "%s.sdp.mdl-model-persistence.input"
)

var (
	// 正负无穷
	NEG_INF float64 = math.Inf(-1)
	POS_INF float64 = math.Inf(1)

	// sql的字符串转义
	Special = strings.NewReplacer(`\`, `\\\\`, `'`, `\'`, `%`, `\%`, `_`, `\_`)
)

type AccountInfo struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}
