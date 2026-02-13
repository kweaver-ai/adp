// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package middleware

import (
	"bytes"
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"go.uber.org/zap"
	"io"
	"time"
)

type AccessLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w AccessLogWriter) Write(p []byte) (int, error) {
	if n, err := w.body.Write(p); err != nil {
		return n, err
	}
	return w.ResponseWriter.Write(p)
}

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyWriter := &AccessLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bodyWriter
		req, err := c.GetRawData()
		if err != nil {
			logger.Errorf("accesslog:%s", err)
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(req))
		beginTime := time.Now()
		c.Next()
		endTime := time.Now()
		durTime := endTime.Sub(beginTime).Seconds()
		response, _ := JsonToMap(bodyWriter.body.String())
		request, _ := JsonToMap(string(req))
		if bodyWriter.Status() != 200 {
			logger.GetLogger().With(
				zap.String("request", c.Request.URL.Path),
				zap.Any("params", c.Param("index")),
				zap.Any("query", c.Query("scroll")),
				zap.Any("request", request),
				zap.Any("response", response)).Errorf(
				"access log: method: %s, status_code: %d, begin_time: %s, end_time: %s, subTime: %f",
				c.Request.Method,
				bodyWriter.Status(),
				beginTime.Format("2006-01-02T15:04:05.000Z0700"),
				endTime.Format("2006-01-02T15:04:05.000Z0700"),
				durTime,
			)
		} else {
			logger.GetLogger().With(zap.String("request", c.Request.URL.Path)).Debugf(
				"access log: method: %s, status_code: %d, begin_time: %s, end_time: %s, subTime: %f",
				c.Request.Method,
				bodyWriter.Status(),
				beginTime.Format("2006-01-02T15:04:05.000Z0700"),
				endTime.Format("2006-01-02T15:04:05.000Z0700"),
				durTime,
			)
		}
	}
}

func JsonToMap(jsonStr string) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	err := sonic.Unmarshal([]byte(jsonStr), &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
