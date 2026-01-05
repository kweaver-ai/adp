package knlogicpropertyresolver

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/infra/config"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/infra/errors"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/infra/rest"
	"devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/interfaces"
	logicskn "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/logics/knlogicpropertyresolver"
	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// KnLogicPropertyResolverHandler 逻辑属性解析 Handler
type KnLogicPropertyResolverHandler interface {
	ResolveLogicProperties(c *gin.Context)
}

type knLogicPropertyResolverHandle struct {
	Logger  interfaces.Logger
	Service interfaces.IKnLogicPropertyResolverService
}

var (
	handlerOnce sync.Once
	handler     KnLogicPropertyResolverHandler
)

// NewKnLogicPropertyResolverHandler 创建 KnLogicPropertyResolverHandler
func NewKnLogicPropertyResolverHandler() KnLogicPropertyResolverHandler {
	handlerOnce.Do(func() {
		conf := config.NewConfigLoader()
		handler = &knLogicPropertyResolverHandle{
			Logger:  conf.GetLogger(),
			Service: logicskn.NewKnLogicPropertyResolverService(),
		}
	})
	return handler
}

// ResolveLogicProperties 解析逻辑属性
// @Summary 解析逻辑属性
// @Description 基于 query + 上下文生成 dynamic_params，并调用底层 ontology-query 接口批量获取逻辑属性值（metric + operator）
// @Tags kn-context-loader
// @Accept json
// @Produce json
// @Param x-account-id header string true "账户ID"
// @Param x-account-type header string true "账户类型"
// @Param x-kn-id header string true "知识网络ID"
// @Param body body interfaces.ResolveLogicPropertiesRequest true "请求参数"
// @Success 200 {object} interfaces.ResolveLogicPropertiesResponse "成功响应"
// @Failure 400 {object} interfaces.MissingParamsError "缺参错误"
// @Failure 404 {object} interfaces.KnBaseError "对象类不存在"
// @Failure 500 {object} interfaces.KnBaseError "服务器错误"
// @Router /api/kn/logic-property-resolver [post]
func (k *knLogicPropertyResolverHandle) ResolveLogicProperties(c *gin.Context) {
	var err error
	req := &interfaces.ResolveLogicPropertiesRequest{
		Options: &interfaces.ResolveOptions{},
	}

	// 绑定 Header 参数
	if err = c.ShouldBindHeader(req); err != nil {
		k.Logger.Errorf("[KnLogicPropertyResolverHandler] Bind header failed: %v", err)
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	// 绑定 JSON Body
	if err = c.ShouldBindJSON(req); err != nil {
		k.Logger.Errorf("[KnLogicPropertyResolverHandler] Bind JSON failed: %v", err)
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	// 设置默认值
	if err = defaults.Set(req.Options); err != nil {
		k.Logger.Errorf("[KnLogicPropertyResolverHandler] Set defaults failed: %v", err)
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	// 参数校验
	err = validator.New().Struct(req)
	if err != nil {
		k.Logger.Errorf("[KnLogicPropertyResolverHandler] Validate failed: %v", err)
		rest.ReplyError(c, err)
		return
	}

	// 📥 记录请求入参（结构化）
	reqJSON, _ := json.Marshal(req)
	k.Logger.Infof("========== [kn-logic-property-resolver] 请求开始 ==========")
	k.Logger.Infof("📥 请求参数: %s", string(reqJSON))

	// 调用 Service 层（记录耗时）
	startTime := time.Now()
	resp, err := k.Service.ResolveLogicProperties(c.Request.Context(), req)
	elapsed := time.Since(startTime).Milliseconds()

	if err != nil {
		k.Logger.Errorf("========== [kn-logic-property-resolver] 请求失败 ========== (耗时: %dms)", elapsed)
		k.Logger.Errorf("❌ 错误信息: %v", err)
		rest.ReplyError(c, err)
		return
	}

	// 📤 记录响应结果
	respJSON, _ := json.Marshal(resp)
	k.Logger.Infof("========== [kn-logic-property-resolver] 请求成功 ========== (耗时: %dms)", elapsed)
	k.Logger.Infof("📤 响应数据: %s", string(respJSON))

	// 返回成功响应
	rest.ReplyOK(c, http.StatusOK, resp)
}
