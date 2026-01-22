package drivenadapters

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

var (
	sandboxOnce sync.Once
	sandBoxEnv  interfaces.SandBoxEnv
)

// sandBoxEnvClient 沙箱环境客户端
type sandBoxEnvClient struct {
	baseURL    string
	logger     interfaces.Logger
	httpClient interfaces.HTTPClient
}

const (
	executeCodeURL = "/v2/execute_code"
)

func NewSandBoxEnvClient() interfaces.SandBoxEnv {
	sandboxOnce.Do(func() {
		conf := config.NewConfigLoader()
		fmt.Println(conf.SandboxRuntime)
		fmt.Println(conf.SandboxRuntime)
		sandBoxEnv = &sandBoxEnvClient{
			baseURL: fmt.Sprintf("%s://%s:%d/workspace/se", conf.SandboxRuntime.PrivateProtocol,
				conf.SandboxRuntime.PrivateHost, conf.SandboxRuntime.PrivatePort),
			logger:     conf.GetLogger(),
			httpClient: rest.NewHTTPClient(),
		}
	})
	return sandBoxEnv
}

// GetSandBoxServerRouter 获取沙箱环境服务器路由
func (s *sandBoxEnvClient) GetSandBoxServerRouter() *interfaces.APIRouter {
	return &interfaces.APIRouter{
		HTTPRouter: interfaces.HTTPRouter{
			URL:    executeCodeURL,
			Method: http.MethodPost,
		},
		ServerURL: s.baseURL,
	}
}

// 获取沙箱环境执行请求配置
func (s *sandBoxEnvClient) GetSandBoxRequestConfig(ctx context.Context, req *interfaces.SandBoxConfigReq) (resp *interfaces.HTTPRequest, err error) {
	// 合并上下文头和请求头
	if req.Headers == nil {
		req.Headers = map[string]any{}
	}
	commonHeaders := common.GetHeaderFromCtx(ctx)
	for k, v := range commonHeaders {
		req.Headers[k] = v
	}
	timeout := time.Duration(req.Timeout) * time.Second
	event := map[string]any{}
	err = utils.AnyToObject(req.Body, &event)
	if err != nil {
		s.logger.Errorf("parse event failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	executeReq := &interfaces.ExecuteCodeReq{
		HandlerCode: req.Code,
		Event:       event,
	}
	if req.Timeout > 0 {
		executeReq.ExecuteContext.RemainingTimeInMillis = timeout.Milliseconds()
	}
	resp = &interfaces.HTTPRequest{
		Timeout: timeout,
		HTTPRouter: interfaces.HTTPRouter{
			URL:    executeCodeURL,
			Method: http.MethodPost,
		},
		HTTPRequestParams: interfaces.HTTPRequestParams{
			Headers: req.Headers,
			Body:    executeReq,
		},
	}
	return
}

// https://{host}:{port}/api/workspace/se/v2/execute_code
// 执行代码
func (s *sandBoxEnvClient) ExecuteCode(ctx context.Context, req *interfaces.ExecuteCodeReq) (resp *interfaces.ExecuteCodeResp, err error) {
	headers := common.GetHeaderFromCtx(ctx)
	code, respData, err := s.httpClient.Post(ctx, fmt.Sprintf("%s%s", s.baseURL, executeCodeURL), headers, req)
	if err != nil {
		s.logger.Errorf("execute code failed, err: %v", err)
		err = errors.NewHTTPError(ctx, code, errors.ErrExtSandboxRuntimeExecuteCodeFailed, err.Error())
		return nil, err
	}
	resp = &interfaces.ExecuteCodeResp{}
	err = utils.AnyToObject(respData, resp)
	if err != nil {
		s.logger.Errorf("parse execute code response failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err)
	}
	return
}
