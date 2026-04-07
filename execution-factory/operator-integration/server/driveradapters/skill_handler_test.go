package driveradapters

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/driveradapters/skill"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSkillRestHandlerRegisterPublic(t *testing.T) {
	Convey("SkillRestHandler RegisterPublic does not expose runtime profile routes", t, func() {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		group := router.Group("/api/v1")

		r := &skillRestHandler{
			SkillHandler:          &noopSkillHandler{},
		}

		r.RegisterPublic(group)

		routes := router.Routes()
		So(hasRoute(routes, http.MethodGet, "/api/v1/skills/:skill_id/runtime-profiles/:entrypoint"), ShouldBeFalse)
		So(hasRoute(routes, http.MethodPost, "/api/v1/skills/:skill_id/runtime-profiles/:entrypoint"), ShouldBeFalse)
		So(hasRoute(routes, http.MethodPost, "/api/v1/skills/:skill_id/runtime-profiles/:entrypoint/execute"), ShouldBeFalse)
	})
}

func TestSkillRestHandlerRegisterPrivate(t *testing.T) {
	Convey("SkillRestHandler RegisterPrivate exposes runtime profile routes", t, func() {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		group := router.Group("/api/agent-operator-integration/internal-v1")

		r := &skillRestHandler{
			SkillHandler: &noopSkillHandler{},
		}

		r.RegisterPrivate(group)

		routes := router.Routes()
		So(hasRoute(routes, http.MethodGet, "/api/agent-operator-integration/internal-v1/skills/:skill_id/runtime-profiles/:entrypoint"), ShouldBeTrue)
		So(hasRoute(routes, http.MethodPost, "/api/agent-operator-integration/internal-v1/skills/:skill_id/runtime-profiles/:entrypoint"), ShouldBeTrue)
		So(hasRoute(routes, http.MethodPost, "/api/agent-operator-integration/internal-v1/skills/:skill_id/runtime-profiles/:entrypoint/execute"), ShouldBeTrue)
	})
}

func hasRoute(routes gin.RoutesInfo, method, path string) bool {
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}

type noopSkillHandler struct{}

func (n *noopSkillHandler) RegisterSkill(c *gin.Context)                   {}
func (n *noopSkillHandler) DeleteSkill(c *gin.Context)                     {}
func (n *noopSkillHandler) UpdateSkillStatus(c *gin.Context)               {}
func (n *noopSkillHandler) DownloadSkill(c *gin.Context)                   {}
func (n *noopSkillHandler) QuerySkillList(c *gin.Context)                  {}
func (n *noopSkillHandler) QuerySkillMarketList(c *gin.Context)            {}
func (n *noopSkillHandler) GetSkillMarketDetail(c *gin.Context)            {}
func (n *noopSkillHandler) GetSkillDetail(c *gin.Context)                  {}
func (n *noopSkillHandler) GetSkillContent(c *gin.Context)                 {}
func (n *noopSkillHandler) ReadSkillFile(c *gin.Context)                   {}
func (n *noopSkillHandler) UpsertSkillRuntimeProfile(c *gin.Context)       {}
func (n *noopSkillHandler) GetSkillRuntimeProfile(c *gin.Context)          {}
func (n *noopSkillHandler) ExecuteSkill(c *gin.Context)                    {}

var _ skill.SkillHandler = (*noopSkillHandler)(nil)
