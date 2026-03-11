package driveradapters

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/common"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/rest"
)

func TestMiddlewareResponseFormat_DefaultAndValid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	convey.Convey("middlewareResponseFormat default and valid values", t, func() {
		convey.Convey("no query param defaults to json", func() {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", http.NoBody)

			mw := middlewareResponseFormat()
			mw(c)

			formatVal, ok := common.GetResponseFormatFromCtx(c.Request.Context())
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(formatVal, convey.ShouldEqual, rest.FormatJSON)
			convey.So(w.Code, convey.ShouldEqual, 200) // default status for recorder
		})

		convey.Convey("response_format=json", func() {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test?response_format=json", http.NoBody)

			mw := middlewareResponseFormat()
			mw(c)

			formatVal, ok := common.GetResponseFormatFromCtx(c.Request.Context())
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(formatVal, convey.ShouldEqual, rest.FormatJSON)
		})

		convey.Convey("response_format=toon", func() {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test?response_format=toon", http.NoBody)

			mw := middlewareResponseFormat()
			mw(c)

			formatVal, ok := common.GetResponseFormatFromCtx(c.Request.Context())
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(formatVal, convey.ShouldEqual, rest.FormatTOON)
		})
	})
}

func TestMiddlewareResponseFormat_Invalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	convey.Convey("middlewareResponseFormat invalid value returns 400", t, func() {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test?response_format=xml", http.NoBody)

		mw := middlewareResponseFormat()
		mw(c)

		convey.So(w.Code, convey.ShouldEqual, http.StatusBadRequest)
		convey.So(w.Body.String(), convey.ShouldContainSubstring, "invalid response_format")
	})
}
