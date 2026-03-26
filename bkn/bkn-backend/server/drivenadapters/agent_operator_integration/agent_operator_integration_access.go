// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package agent_operator_integration

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

const (
	defaultBusinessDomain = "bd_public"
)

var (
	aoiAccessOnce sync.Once
	aoiAccess     interfaces.AgentOperatorIntegrationAccess
)

type agentOperatorIntegrationAccess struct {
	appSetting  *common.AppSetting
	httpClient  rest.HTTPClient
	operatorURL string
}

// NewAgentOperatorIntegrationAccess 创建 agent-operator-integration 访问实例
func NewAgentOperatorIntegrationAccess(appSetting *common.AppSetting) interfaces.AgentOperatorIntegrationAccess {
	aoiAccessOnce.Do(func() {
		aoiAccess = &agentOperatorIntegrationAccess{
			appSetting:  appSetting,
			httpClient:  common.NewHTTPClient(),
			operatorURL: appSetting.AgentOperatorIntegrationUrl,
		}
	})

	return aoiAccess
}

func (a *agentOperatorIntegrationAccess) RegisterInternalTool(ctx context.Context, body []byte) error {
	if a.operatorURL == "" {
		return fmt.Errorf("AgentOperatorIntegrationUrl not configured")
	}

	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Register internal tool",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	// "/api/agent-operator-integration/internal-v1/tool-box/intcomp"
	httpURL := fmt.Sprintf("%s/tool-box/intcomp", a.operatorURL)
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         httpURL,
		HttpMethod:      http.MethodPost,
		HttpContentType: rest.ContentTypeJson,
	})

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:           interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_BUSINESS_DOMAIN: defaultBusinessDomain,
		interfaces.HTTP_HEADER_ACCOUNT_ID:      interfaces.ADMIN_ACCOUNT_ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE:    interfaces.ADMIN_ACCOUNT_TYPE,
	}

	respCode, respData, err := a.httpClient.PostNoUnmarshal(ctx, httpURL, headers, body)
	logger.Debugf("RegisterInternalTool [%s] finished, response code is [%d], result is [%s], error is [%v]",
		httpURL, respCode, respData, err)

	if err != nil {
		errDetails := fmt.Sprintf("RegisterInternalTool http request failed: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetAttributes(attr.Key("error").String(errDetails))
		return fmt.Errorf("RegisterInternalTool http request failed: %s", err)
	}

	if respCode != http.StatusOK {
		errDetails := fmt.Sprintf("RegisterInternalTool failed: status=%d, body=%s", respCode, respData)
		logger.Error(errDetails)
		span.SetAttributes(attr.Key("error").String(errDetails))
		return fmt.Errorf("RegisterInternalTool failed: %s", errDetails)
	}

	return nil
}

// ProbeToolBoxTool sends a minimal POST to tool-box proxy to verify the tool exists (HTTP layer / gateway).
func (a *agentOperatorIntegrationAccess) ProbeToolBoxTool(ctx context.Context, boxID, toolID string) error {
	if a.operatorURL == "" {
		return fmt.Errorf("AgentOperatorIntegrationUrl not configured")
	}
	if boxID == "" || toolID == "" {
		return fmt.Errorf("box_id and tool_id must be non-empty")
	}

	httpURL := fmt.Sprintf("%s/tool-box/%s/tool/%s",
		a.operatorURL, url.PathEscape(boxID), url.PathEscape(toolID))

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	// Short timeout payload; endpoint returns 200 when the proxy accepts the call.
	body := []byte(`{"timeout":3}`)
	respCode, respData, err := a.httpClient.PostNoUnmarshal(ctx, httpURL, headers, body)
	logger.Debugf("ProbeToolBoxTool [%s] response code=%d err=%v", httpURL, respCode, err)

	if err != nil {
		return fmt.Errorf("ProbeToolBoxTool request failed: %w", err)
	}
	if respCode != http.StatusOK {
		return fmt.Errorf("ProbeToolBoxTool failed: status=%d body=%s", respCode, respData)
	}
	return nil
}
