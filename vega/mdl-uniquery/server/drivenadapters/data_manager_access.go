// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"uniquery/common"
	"uniquery/interfaces"
)

var (
	dmAccessOnce sync.Once
	dmAccess     interfaces.DataManagerAccess
)

type dataManagerAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewDataManagerAccess(appSetting *common.AppSetting) interfaces.DataManagerAccess {
	dmAccessOnce.Do(func() {
		dmAccess = &dataManagerAccess{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
		}
	})
	return dmAccess
}

func (dma *dataManagerAccess) GetLogGroupRoots(ctx context.Context, host string, accessToken string) (data any, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 调用DataManager服务获取日志分组根分组",
		trace.WithSpanKind(trace.SpanKindClient))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 拼接fullPath
	fullPath := fmt.Sprintf("%s/manager/loggroup/roots", host)

	span.SetAttributes(
		attr.Key("request_url").String(fullPath),
	)

	// 2. http request
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Language":   rest.GetLanguageByCtx(ctx),
	}
	if accessToken != "" {
		headers["X-Access-Token"] = accessToken
	}
	respCode, respData, err := dma.httpClient.Get(ctx, fullPath, nil, headers)

	// 3. 分情况处理http response
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get data by http client: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return nil, err
	}

	if respCode == http.StatusUnauthorized {
		errDetails := fmt.Sprintf("failed to authorized by access-token: %s", accessToken)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return nil, errors.New(errDetails)
	}

	if respCode != http.StatusOK {
		errDetails := fmt.Sprintf("Failed to get data by http client: %v", respData)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return nil, errors.New(errDetails)
	}

	return respData, nil
}

func (dma *dataManagerAccess) GetLogGroupTree(ctx context.Context, userID string, host string, accessToken string) (data any, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 调用DataManager服务获取日志分组根分组",
		trace.WithSpanKind(trace.SpanKindClient))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 拼接fullPath
	fullPath := fmt.Sprintf("%s/manager/loggroup/%s/tree", host, userID)

	span.SetAttributes(
		attr.Key("request_url").String(fullPath),
	)

	// 2. http request
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Language":   rest.GetLanguageByCtx(ctx),
	}
	if accessToken != "" {
		headers["X-Access-Token"] = accessToken
	}
	respCode, respData, err := dma.httpClient.Get(ctx, fullPath, nil, headers)

	// 3. 分情况处理http response
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get data by http client: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return nil, err
	}

	if respCode == http.StatusUnauthorized {
		errDetails := fmt.Sprintf("failed to authorized by access-token: %s", accessToken)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return nil, errors.New(errDetails)
	}

	if respCode != http.StatusOK {
		errDetails := fmt.Sprintf("Failed to get data by http client: %v", respData)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return nil, errors.New(errDetails)
	}

	return respData, nil
}

func (dma *dataManagerAccess) GetLogGroupChildren(ctx context.Context, logGroupID string, host string, accessToken string) (data any, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 调用DataManager服务获取日志分组根分组",
		trace.WithSpanKind(trace.SpanKindClient))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 拼接fullPath
	fullPath := fmt.Sprintf("%s/manager/loggroup/%s/children", host, logGroupID)

	span.SetAttributes(
		attr.Key("request_url").String(fullPath),
	)

	// 2. http request
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Language":   rest.GetLanguageByCtx(ctx),
	}
	if accessToken != "" {
		headers["X-Access-Token"] = accessToken
	}
	respCode, respData, err := dma.httpClient.Get(ctx, fullPath, nil, headers)

	// 3. 分情况处理http response
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get data by http client: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return nil, err
	}

	if respCode == http.StatusUnauthorized {
		errDetails := fmt.Sprintf("failed to authorized by access-token: %s", accessToken)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return nil, errors.New(errDetails)
	}

	if respCode != http.StatusOK {
		errDetails := fmt.Sprintf("Failed to get data by http client: %v", respData)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return nil, errors.New(errDetails)
	}

	return respData, nil
}
