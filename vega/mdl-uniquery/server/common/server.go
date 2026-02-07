// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import (
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	uerrors "uniquery/errors"
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
	var body string
	switch e := err.(type) {
	case *uerrors.OpenSearchError:
		if statusCode != e.StatusCode {
			logger.Errorf("OpenSearchError's StatusCode '%d' is not match with argument: '%d'", statusCode, e.StatusCode)
			statusCode = e.StatusCode
		}
		body = e.Error()

	case uerrors.PromQLError:
		b, _ := sonic.Marshal(&uerrors.ResponseError{
			Status:    uerrors.StatusError,
			ErrorType: e.Typ,
			Error:     e.Error(),
		})
		body = string(b)
	default:
		body = err.Error()
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.String(statusCode, body)
}
