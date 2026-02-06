package common

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metadata"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/sandbox"
)

// UnifiedProxyHandler 统一代理处理接口
type UnifiedProxyHandler interface {
	FunctionExecuteProxy(c *gin.Context)
	FunctionExecute(c *gin.Context)
}

// unifiedProxyHandler 代理处理实现
type unifiedProxyHandler struct {
	Logger          interfaces.Logger
	MetadataService interfaces.IMetadataService
	SessionPool     sandbox.SessionPool
}

var (
	pOnce        sync.Once
	proxyHandler UnifiedProxyHandler
)

func NewUnifiedProxyHandler() UnifiedProxyHandler {
	pOnce.Do(func() {
		conf := config.NewConfigLoader()
		proxyHandler = &unifiedProxyHandler{
			Logger:          conf.Logger,
			MetadataService: metadata.NewMetadataService(),
			SessionPool:     sandbox.GetSessionPool(),
		}
	})
	return proxyHandler
}

// FunctionExecute 函数执行
func (h *unifiedProxyHandler) FunctionExecute(c *gin.Context) {
	var err error
	req := &interfaces.ExecuteCodeReq{}
	if err = c.ShouldBindJSON(req); err != nil {
		err = errors.NewHTTPError(c.Request.Context(), http.StatusBadRequest, errors.ErrExtDebugParamsInvalid,
			fmt.Sprintf("invalid request body, err: %v", err))
		rest.ReplyError(c, err)
		return
	}

	err = defaults.Set(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, fmt.Sprintf("set default value failed, err: %v", err))
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.SessionPool.ExecuteCode(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	h.Logger.Infof("FunctionExecute resp: %v", resp)
	result := &FunctionExecuteResp{
		Stdout:  resp.Stdout,
		Stderr:  resp.Stderr,
		Result:  resp.ReturnValue,
		Metrics: resp.Metrics,
	}
	rest.ReplyOK(c, http.StatusOK, result)
}

// FunctionExecuteResp 函数执行响应
type FunctionExecuteResp struct {
	Stdout  string `json:"stdout"`  // 标准输出
	Stderr  string `json:"stderr"`  // 标准错误输出
	Result  any    `json:"result"`  // 执行结果值
	Metrics any    `json:"metrics"` // 执行指标
}

// FunctionExecuteProxyReq 函数执行代理请求参数
type FunctionExecuteProxyReq struct {
	Version string `uri:"version" validate:"required,uuid4"`
	Timeout int64  `query:"timeout"` // 毫秒
}

// FunctionExecuteProxy 执行代理请求
func (h *unifiedProxyHandler) FunctionExecuteProxy(c *gin.Context) {
	var err error
	req := &FunctionExecuteProxyReq{}
	if err = c.ShouldBindUri(req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	// 读取请求体
	event := map[string]any{}
	if err = c.ShouldBindJSON(&event); err != nil {
		err = errors.NewHTTPError(c.Request.Context(), http.StatusBadRequest, errors.ErrExtDebugParamsInvalid,
			fmt.Sprintf("invalid request body, err: %v", err))
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	// 获取元数据
	exists, metadata, err := h.MetadataService.CheckMetadataExists(c.Request.Context(), interfaces.MetadataTypeFunc, req.Version)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	if !exists {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusNotFound, fmt.Sprintf("metadata %s not found", req.Version))
		rest.ReplyError(c, err)
		return
	}

	// 执行函数
	code, scriptType, _ := metadata.GetFunctionContent()
	if scriptType != string(interfaces.ScriptTypePython) {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, fmt.Sprintf("script_type %s not supported", scriptType))
		rest.ReplyError(c, err)
		return
	}
	if code == "" {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, fmt.Sprintf("function code is empty for version %s", req.Version))
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.SessionPool.ExecuteCode(c.Request.Context(), &interfaces.ExecuteCodeReq{
		Code:     code,
		Event:    event,
		Timeout:  int(req.Timeout / 1000),
		Language: string(interfaces.ScriptTypePython),
	})
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	h.Logger.Infof("FunctionExecuteProxy resp: %v", resp)
	// 转换为 FunctionExecuteResp
	result := &FunctionExecuteResp{
		Stdout:  resp.Stdout,
		Stderr:  resp.Stderr,
		Result:  resp.ReturnValue,
		Metrics: resp.Metrics,
	}
	rest.ReplyOK(c, http.StatusOK, result)
}
