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

	"github.com/bytedance/sonic"
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
	dcaOnce sync.Once
	dca     interfaces.DataConnectionAccess
)

type dataConnectionAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewDataConnectionAccess(appSetting *common.AppSetting) interfaces.DataConnectionAccess {
	dcaOnce.Do(func() {
		dca = &dataConnectionAccess{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
		}
	})
	return dca
}

// 根据数据连接ID获取对象
func (dcAccess *dataConnectionAccess) GetDataConnectionByID(ctx context.Context,
	connID string) (conn *interfaces.DataConnection, isExist bool, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 调用DataModel服务获取数据连接",
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
	fullPath := fmt.Sprintf("%s/v1/data-connections/%v?with_auth_info=true", dcAccess.appSetting.DataModelUrl, connID)

	span.SetAttributes(
		attr.Key("request_url").String(fullPath),
		attr.Key("conn_id").String(connID),
	)

	// 2. http request
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Language":   rest.GetLanguageByCtx(ctx),
	}
	respCode, respData, err := dcAccess.httpClient.GetNoUnmarshal(ctx, fullPath, nil, headers)

	// 3. 分情况处理http response
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get data connection by http client: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return &interfaces.DataConnection{}, false, err
	}

	if respCode == http.StatusNotFound {
		errDetails := fmt.Sprintf("Failed to get data connection by http client: the data connection whose id equal to %s was not found", connID)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return &interfaces.DataConnection{}, false, nil
	}

	if respCode != http.StatusOK {
		errDetails := fmt.Sprintf("Failed to get data connection by http client: %s", string(respData))
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return &interfaces.DataConnection{}, false, errors.New(errDetails)
	}

	conn = &interfaces.DataConnection{}
	if err = sonic.Unmarshal(respData, &conn); err != nil {
		errDetails := fmt.Sprintf("Unmarshal http response failed: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return &interfaces.DataConnection{}, false, err
	}

	return conn, true, nil
}

// 根据数据连接Name获取对象
func (dcAccess *dataConnectionAccess) GetDataConnectionTypeByName(ctx context.Context,
	connName string) (sourceType string, isExist bool, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 根据名称调用DataModel服务获取数据连接类型",
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
	fullPath := fmt.Sprintf("%s/v1/data-connections?name=%v", dcAccess.appSetting.DataModelUrl, connName)

	span.SetAttributes(
		attr.Key("request_url").String(fullPath),
		attr.Key("conn_name").String(connName),
	)

	// 2. http request
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Language":   rest.GetLanguageByCtx(ctx),
	}
	respCode, respData, err := dcAccess.httpClient.GetNoUnmarshal(ctx, fullPath, nil, headers)

	// 3. 分情况处理http response
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get data connection type by http client: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return "", false, err
	}

	if respCode != http.StatusOK {
		errDetails := fmt.Sprintf("Failed to get data connection type by http client: %v", string(respData))
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return "", false, errors.New(errDetails)
	}

	connLists := struct {
		Entries []struct {
			DataSourceType string `json:"data_source_type"`
		} `json:"entries"`
	}{}

	if err = sonic.Unmarshal(respData, &connLists); err != nil {
		errDetails := fmt.Sprintf("Unmarshal http response failed: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return "", false, err
	}

	if len(connLists.Entries) == 0 {
		return "", false, nil
	}

	return connLists.Entries[0].DataSourceType, true, nil
}
