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

func MockNewConceptGroupRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra,
	cgs interfaces.ConceptGroupService,
	kns interfaces.KNService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		cgs:        cgs,
		kns:        kns,
	}
	return r
}

func Test_ConceptGroupRestHandler_CreateConceptGroup(t *testing.T) {
	Convey("Test ConceptGroupHandler CreateConceptGroup\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		cgs := dmock.NewMockConceptGroupService(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, hydra, cgs, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		knID := "kn1"
		url := "/api/ontology-manager/v1/knowledge-networks/" + knID + "/concept-groups"

		conceptGroup := interfaces.ConceptGroup{
			CGName: "group1",
			CommonInfo: interfaces.CommonInfo{
				Comment: "test comment",
			},
		}

		Convey("Success CreateConceptGroup \n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CreateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("cg1", nil)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Failed CreateConceptGroup ShouldBind Error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			reqParamByte, _ := sonic.Marshal([]interfaces.ConceptGroup{conceptGroup})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("CG name is null\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			reqParamByte, _ := sonic.Marshal(interfaces.ConceptGroup{})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("CheckKNExistByID failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: oerrors.OntologyManager_KnowledgeNetwork_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, expectedErr)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ConceptGroupRestHandler_UpdateConceptGroup(t *testing.T) {
	Convey("Test ConceptGroupHandler UpdateConceptGroup\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		cgs := dmock.NewMockConceptGroupService(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, hydra, cgs, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		knID := "kn1"
		cgID := "cg1"
		url := "/api/ontology-manager/v1/knowledge-networks/" + knID + "/concept-groups/" + cgID

		conceptGroup := interfaces.ConceptGroup{
			CGID:   cgID,
			CGName: "group1",
		}

		Convey("Success UpdateConceptGroup\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("group2", true, nil)
			cgs.EXPECT().CheckConceptGroupExistByName(gomock.Any(), knID, gomock.Any(), conceptGroup.CGName).Return("", false, nil)
			cgs.EXPECT().UpdateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Failed UpdateConceptGroup ShouldBind Error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader([]byte("invalid json")))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("ConceptGroup not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})
	})
}

func Test_ConceptGroupRestHandler_DeleteConceptGroup(t *testing.T) {
	Convey("Test ConceptGroupHandler DeleteConceptGroup\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		cgs := dmock.NewMockConceptGroupService(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, hydra, cgs, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		knID := "kn1"
		cgID := "cg1"
		url := "/api/ontology-manager/v1/knowledge-networks/" + knID + "/concept-groups/" + cgID

		Convey("Success DeleteConceptGroup\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("group1", true, nil)
			cgs.EXPECT().DeleteConceptGroupByID(gomock.Any(), gomock.Any(), knID, gomock.Any(), cgID).Return(int64(1), nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("ConceptGroup not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("", false, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})
	})
}

func Test_ConceptGroupRestHandler_ListConceptGroups(t *testing.T) {
	Convey("Test ConceptGroupHandler ListConceptGroups\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		cgs := dmock.NewMockConceptGroupService(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, hydra, cgs, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		knID := "kn1"
		url := "/api/ontology-manager/v1/knowledge-networks/" + knID + "/concept-groups"

		Convey("Success ListConceptGroups\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().ListConceptGroups(gomock.Any(), gomock.Any()).Return([]*interfaces.ConceptGroup{}, 0, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})
	})
}

func Test_ConceptGroupRestHandler_GetConceptGroup(t *testing.T) {
	Convey("Test ConceptGroupHandler GetConceptGroup\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		cgs := dmock.NewMockConceptGroupService(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, hydra, cgs, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		knID := "kn1"
		cgID := "cg1"
		url := "/api/ontology-manager/v1/knowledge-networks/" + knID + "/concept-groups/" + cgID

		Convey("Success GetConceptGroup\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().GetConceptGroupByID(gomock.Any(), knID, gomock.Any(), cgID, gomock.Any()).Return(&interfaces.ConceptGroup{}, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})
	})
}

func Test_ConceptGroupRestHandler_AddObjectTypesToConceptGroup(t *testing.T) {
	Convey("Test ConceptGroupHandler AddObjectTypesToConceptGroup\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		cgs := dmock.NewMockConceptGroupService(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, hydra, cgs, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		knID := "kn1"
		cgID := "cg1"
		url := "/api/ontology-manager/v1/knowledge-networks/" + knID + "/concept-groups/" + cgID + "/object-types"

		requestData := struct {
			Entries []interfaces.ID `json:"entries"`
		}{
			Entries: []interfaces.ID{{ID: "ot1"}},
		}

		Convey("Success AddObjectTypesToConceptGroup\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("group1", true, nil)
			cgs.EXPECT().AddObjectTypesToConceptGroup(gomock.Any(), gomock.Any(), knID, gomock.Any(), cgID, gomock.Any(), gomock.Any()).Return([]string{"ot1"}, nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})
	})
}

func Test_ConceptGroupRestHandler_DeleteObjectTypesFromGroup(t *testing.T) {
	Convey("Test ConceptGroupHandler DeleteObjectTypesFromGroup\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		cgs := dmock.NewMockConceptGroupService(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, hydra, cgs, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		knID := "kn1"
		cgID := "cg1"
		otIDs := "ot1,ot2"
		url := "/api/ontology-manager/v1/knowledge-networks/" + knID + "/concept-groups/" + cgID + "/object-types/" + otIDs

		Convey("Success DeleteObjectTypesFromGroup\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("group1", true, nil)
			cgs.EXPECT().DeleteObjectTypesFromGroup(gomock.Any(), gomock.Any(), knID, gomock.Any(), cgID, gomock.Any()).Return(int64(2), nil)
			cgs.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any()).Return([]interfaces.ConceptGroupRelation{
				{
					CGID:      cgID,
					ConceptID: "ot1",
				},
				{
					CGID:      cgID,
					ConceptID: "ot2",
				},
			}, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})
	})
}
