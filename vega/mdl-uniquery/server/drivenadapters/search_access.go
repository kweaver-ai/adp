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
	"net/url"
	"sync"
	"uniquery/common"
	"uniquery/interfaces"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	sAccessOnce sync.Once
	sAccess     interfaces.SearchAccess
)

type searchAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewSearchAccess(appSetting *common.AppSetting) interfaces.SearchAccess {
	sAccessOnce.Do(func() {
		sAccess = &searchAccess{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
		}
	})
	return sAccess
}

func (sa *searchAccess) SearchSubmit(ctx context.Context, queryBody any,
	userID string, host string, accessToken string) (data any, err error) {

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
	fullPath := fmt.Sprintf("%s/v1/search/submit", host)

	span.SetAttributes(
		attr.Key("request_url").String(fullPath),
	)

	// 2. http request
	headers := map[string]string{
		"Content-Type":   "application/json",
		"X-Language":     rest.GetLanguageByCtx(ctx),
		"User":           userID,
		"X-Access-Token": accessToken,
	}
	respCode, respData, err := sa.httpClient.Post(ctx, fullPath, headers, queryBody)

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

func (sa *searchAccess) SearchFetch(ctx context.Context, jobID string,
	queryParams url.Values, host string, accessToken string) (data any, err error) {

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
	fullPath := fmt.Sprintf("%s/v1/search/fetch/%s", host, jobID)

	span.SetAttributes(
		attr.Key("request_url").String(fullPath),
	)

	// 2. http request
	headers := map[string]string{
		"Content-Type":   "application/json",
		"X-Language":     rest.GetLanguageByCtx(ctx),
		"X-Access-Token": accessToken,
	}
	respCode, respData, err := sa.httpClient.Get(ctx, fullPath, queryParams, headers)

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

func (sa *searchAccess) SearchFetchFields(ctx context.Context, jobID string,
	queryParams url.Values, host string, accessToken string) (data any, err error) {

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
	fullPath := fmt.Sprintf("%s/v1/search/fetch/%s/fields", host, jobID)

	span.SetAttributes(
		attr.Key("request_url").String(fullPath),
	)

	// 2. http request
	headers := map[string]string{
		"Content-Type":   "application/json",
		"X-Language":     rest.GetLanguageByCtx(ctx),
		"X-Access-Token": accessToken,
	}
	respCode, respData, err := sa.httpClient.Get(ctx, fullPath, queryParams, headers)

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

func (sa *searchAccess) SearchFetchSameFields(ctx context.Context,
	jobID string, host string, accessToken string) (data any, err error) {

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
	fullPath := fmt.Sprintf("%s/v1/search/fetch/%s/fields/sameFields", host, jobID)

	span.SetAttributes(
		attr.Key("request_url").String(fullPath),
	)

	// 2. http request
	headers := map[string]string{
		"Content-Type":   "application/json",
		"X-Language":     rest.GetLanguageByCtx(ctx),
		"X-Access-Token": accessToken,
	}
	respCode, respData, err := sa.httpClient.Get(ctx, fullPath, nil, headers)

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

func (sa *searchAccess) SearchContext(ctx context.Context, queryParams url.Values,
	userID string, host string, accessToken string) (data any, err error) {

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
	fullPath := fmt.Sprintf("%s/v1/search/context", host)

	span.SetAttributes(
		attr.Key("request_url").String(fullPath),
	)

	// 2. http request
	headers := map[string]string{
		"Content-Type":   "application/json",
		"X-Language":     rest.GetLanguageByCtx(ctx),
		"User":           userID,
		"X-Access-Token": accessToken,
	}
	respCode, respData, err := sa.httpClient.Get(ctx, fullPath, queryParams, headers)

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
