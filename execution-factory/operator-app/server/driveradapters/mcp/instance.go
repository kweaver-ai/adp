package mcp

import (
	"net/http"

	infraerrors "github.com/kweaver-ai/adp/execution-factory/operator-app/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-app/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-app/server/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func (h *mcpHnadle) CreateMCPInstance(c *gin.Context) {
	var err error
	req := &interfaces.MCPDeployCreateRequest{}
	err = c.ShouldBindJSON(req)
	if err != nil {
		err = infraerrors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	res, err := h.mcpService.CreateMCPInstance(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, res)
}

func (h *mcpHnadle) DeleteMCPInstance(c *gin.Context) {
	var err error
	req := &interfaces.MCPDeleteRequest{}
	err = c.ShouldBindUri(req)
	if err != nil {
		err = infraerrors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = h.mcpService.DeleteMCPInstance(c.Request.Context(), req.MCPID, req.Version)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, nil)
}

func (h *mcpHnadle) UpdateMCPInstance(c *gin.Context) {
	var err error
	req := &interfaces.MCPDeployUpdateRequest{}
	err = c.ShouldBindJSON(req)
	if err != nil {
		err = infraerrors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	err = c.ShouldBindUri(req)
	if err != nil {
		err = infraerrors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	res, err := h.mcpService.UpdateMCPInstance(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, res)
}

func (h *mcpHnadle) DeleteByMCPID(c *gin.Context) {
	var err error
	req := &interfaces.MCPDeleteByMCPIDReq{}
	err = c.ShouldBindUri(req)
	if err != nil {
		err = infraerrors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	err = h.mcpService.DeleteByMCPID(c.Request.Context(), req.MCPID)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, nil)
}

// UpgradeMCPInstance 升级mcp服务实例
func (h *mcpHnadle) UpgradeMCPInstance(c *gin.Context) {
	var err error
	req := &interfaces.MCPDeployCreateRequest{}
	err = c.ShouldBindJSON(req)
	if err != nil {
		err = infraerrors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	res, err := h.mcpService.UpgradeMCPInstance(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, res)
}
