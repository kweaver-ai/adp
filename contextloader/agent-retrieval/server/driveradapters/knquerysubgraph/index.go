package knquerysubgraph

import (
	"net/http"
	"sync"

	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/infra/config"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/infra/errors"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/infra/rest"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/interfaces"
	logicskn "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/logics/knquerysubgraph"
	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// KnQuerySubgraphHandler 子图查询处理器
type KnQuerySubgraphHandler interface {
	QueryInstanceSubgraph(c *gin.Context)
}

type knQuerySubgraphHandler struct {
	Logger                 interfaces.Logger
	KnQuerySubgraphService interfaces.IKnQuerySubgraphService
}

var (
	kqsOnce    sync.Once
	kqsHandler KnQuerySubgraphHandler
)

// NewKnQuerySubgraphHandler 新建 KnQuerySubgraphHandler
func NewKnQuerySubgraphHandler() KnQuerySubgraphHandler {
	kqsOnce.Do(func() {
		conf := config.NewConfigLoader()
		kqsHandler = &knQuerySubgraphHandler{
			Logger:                 conf.GetLogger(),
			KnQuerySubgraphService: logicskn.NewKnQuerySubgraphService(),
		}
	})
	return kqsHandler
}

// QueryInstanceSubgraph 查询对象子图
func (h *knQuerySubgraphHandler) QueryInstanceSubgraph(c *gin.Context) {
	var err error
	req := &interfaces.QueryInstanceSubgraphReq{}

	// 绑定 Header
	if err = c.ShouldBindHeader(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	// 绑定 Path Parameters
	if err = c.ShouldBindUri(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	// 绑定 Query Parameters
	if err = c.ShouldBindQuery(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	// 绑定 JSON Body
	if err = c.ShouldBindJSON(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	// 设置默认值
	if err = defaults.Set(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	// 参数校验
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	// 调用业务逻辑
	resp, err := h.KnQuerySubgraphService.QueryInstanceSubgraph(c.Request.Context(), req)
	if err != nil {
		h.Logger.Errorf("[KnQuerySubgraphHandler#QueryInstanceSubgraph] QueryInstanceSubgraph failed, err: %v", err)
		rest.ReplyError(c, err)
		return
	}

	// 返回成功响应
	rest.ReplyOK(c, http.StatusOK, resp)
}
