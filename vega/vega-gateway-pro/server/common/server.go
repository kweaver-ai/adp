package common

import (
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
)

// ReplyOK 响应成功
func ReplyOK(c *gin.Context, statusCode int, body interface{}) {
	var bodyStr string
	if body != nil {
		if v, ok := body.([]byte); ok {
			bodyStr = string(v)
		} else {
			b, _ := sonic.Marshal(body)
			bodyStr = string(b)
		}
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.String(statusCode, bodyStr)
}

// ReplyError 响应错误
func ReplyError(c *gin.Context, statusCode int, err error) {
	body := err.Error()

	c.Writer.Header().Set("Content-Type", "application/json")
	c.String(statusCode, body)
}
