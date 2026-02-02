package driveradapters

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"ontology-manager/common"
	oerrors "ontology-manager/errors"
	"ontology-manager/interfaces"
	dmock "ontology-manager/interfaces/mock"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"
)

func MockNewKnowledgeNetworkRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra,
	kns interfaces.KNService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		kns:        kns,
	}
	return r
}

func Test_KnowledgeNetworkRestHandler_CreateKN(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler CreateKN\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewKnowledgeNetworkRestHandler(appSetting, hydra, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/ontology-manager/v1/knowledge-networks"

		kn := interfaces.KN{
			KNName:  "kn1",
			Comment: "test comment",
			Branch:  "main",
		}

		Convey("Success CreateKN \n", func() {
			kns.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return("kn1", nil)

			reqParamByte, _ := sonic.Marshal(kn)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_BUSINESS_DOMAIN, "domain1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Failed CreateKN ShouldBind Error\n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.KN{kn})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_BUSINESS_DOMAIN, "domain1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN name is null\n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.KN{})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_BUSINESS_DOMAIN, "domain1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Business domain is empty\n", func() {
			reqParamByte, _ := sonic.Marshal(kn)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("CreateKN failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: oerrors.OntologyManager_KnowledgeNetwork_InternalError,
				},
			}

			kns.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return("", err)

			reqParamByte, _ := sonic.Marshal(kn)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_BUSINESS_DOMAIN, "domain1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_KnowledgeNetworkRestHandler_UpdateKN(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler UpdateKN\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewKnowledgeNetworkRestHandler(appSetting, hydra, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		knID := "kn1"
		url := "/api/ontology-manager/v1/knowledge-networks/" + knID

		kn := interfaces.KN{
			KNID:   knID,
			KNName: "kn1",
			Branch: interfaces.MAIN_BRANCH,
		}

		Convey("Success UpdateKN\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("kn2", true, nil)
			kns.EXPECT().CheckKNExistByName(gomock.Any(), kn.KNName, gomock.Any()).Return("", false, nil)
			kns.EXPECT().UpdateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(kn)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Failed UpdateKN ShouldBind Error\n", func() {
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader([]byte("invalid json")))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(kn)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})
	})
}

func Test_KnowledgeNetworkRestHandler_DeleteKN(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler DeleteKN\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewKnowledgeNetworkRestHandler(appSetting, hydra, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		knID := "kn1"
		url := "/api/ontology-manager/v1/knowledge-networks/" + knID

		Convey("Success DeleteKN\n", func() {
			kns.EXPECT().GetKNByID(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(&interfaces.KN{
				KNID:   knID,
				KNName: "kn1",
			}, nil)
			kns.EXPECT().DeleteKN(gomock.Any(), gomock.Any()).Return(nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().GetKNByID(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(nil, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})
}

func Test_KnowledgeNetworkRestHandler_ListKNs(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler ListKNs\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewKnowledgeNetworkRestHandler(appSetting, hydra, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/ontology-manager/v1/knowledge-networks"

		Convey("Success ListKNs\n", func() {
			kns.EXPECT().ListKNs(gomock.Any(), gomock.Any()).Return([]*interfaces.KN{}, 0, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?business_domain=domain1", nil)
			req.Header.Set(interfaces.HTTP_HEADER_BUSINESS_DOMAIN, "domain1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Failed with empty business domain\n", func() {
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func Test_KnowledgeNetworkRestHandler_GetKN(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler GetKN\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewKnowledgeNetworkRestHandler(appSetting, hydra, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		knID := "kn1"
		url := "/api/ontology-manager/v1/knowledge-networks/" + knID

		Convey("Success GetKN\n", func() {
			kns.EXPECT().GetKNByID(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(&interfaces.KN{}, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("KN not found\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusNotFound,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: oerrors.OntologyManager_KnowledgeNetwork_NotFound,
				},
			}

			kns.EXPECT().GetKNByID(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(nil, err)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})
}

func Test_KnowledgeNetworkRestHandler_GetRelationTypePaths(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler GetRelationTypePaths\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewKnowledgeNetworkRestHandler(appSetting, hydra, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		knID := "kn1"
		url := "/api/ontology-manager/v1/knowledge-networks/" + knID + "/relation-type-paths"

		query := interfaces.RelationTypePathsBaseOnSource{
			SourceObjecTypeId: "ot1",
			Direction:         interfaces.DIRECTION_FORWARD,
			PathLength:        2,
		}

		Convey("Success GetRelationTypePaths\n", func() {
			kns.EXPECT().GetKNByID(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(&interfaces.KN{
				KNID:   knID,
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}, nil)
			kns.EXPECT().GetRelationTypePaths(gomock.Any(), gomock.Any()).Return([]interfaces.RelationTypePath{}, nil)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Failed GetRelationTypePaths ShouldBind Error\n", func() {

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("invalid json")))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().GetKNByID(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(nil, nil)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})
}
