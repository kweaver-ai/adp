package driveradapters

import (
	"net/http"
	"net/http/httptest"
	"ontology-manager/common"
	"ontology-manager/interfaces"
	dmock "ontology-manager/interfaces/mock"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"
)

func MockNewResourceRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra,
	kns interfaces.KNService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		kns:        kns,
	}
	return r
}

func Test_RestHandler_ListResources(t *testing.T) {
	Convey("Test RestHandler ListResources\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewResourceRestHandler(appSetting, hydra, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/ontology-manager/v1/resources"

		Convey("Success ListResources with KN type\n", func() {
			kns.EXPECT().ListKnSrcs(gomock.Any(), gomock.Any()).Return([]interfaces.Resource{}, 0, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?resource_type="+interfaces.RESOURCE_TYPE_KN, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Success ListResources with unknown type\n", func() {
			req := httptest.NewRequest(http.MethodGet, url+"?resource_type=unknown", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			// 默认情况下不返回错误，只是不处理
			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}
