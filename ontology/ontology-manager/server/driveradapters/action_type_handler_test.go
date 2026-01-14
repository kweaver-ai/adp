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

func MockNewActionTypeRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra,
	ats interfaces.ActionTypeService,
	kns interfaces.KNService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		ats:        ats,
		kns:        kns,
	}
	return r
}

func Test_ActionTypeRestHandler_CreateActionTypes(t *testing.T) {
	Convey("Test ActionTypeHandler CreateActionTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		ats := dmock.NewMockActionTypeService(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewActionTypeRestHandler(appSetting, hydra, ats, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		knID := "kn1"
		url := "/api/ontology-manager/v1/knowledge-networks/" + knID + "/action-types"

		actionType := &interfaces.ActionType{
			ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
				ATID:   "at1",
				ATName: "action1",
			},
		}
		requestData := struct {
			Entries []*interfaces.ActionType `json:"entries"`
		}{
			Entries: []*interfaces.ActionType{actionType},
		}

		Convey("Success CreateActionTypes \n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CreateActionTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"at1"}, nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Failed CreateActionTypes ShouldBind Error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			reqParamByte, _ := sonic.Marshal([]interfaces.ActionType{*actionType})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty entries\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			emptyRequestData := struct {
				Entries []*interfaces.ActionType `json:"entries"`
			}{
				Entries: []*interfaces.ActionType{},
			}
			reqParamByte, _ := sonic.Marshal(emptyRequestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
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

		Convey("CheckKNExistByID failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: oerrors.OntologyManager_KnowledgeNetwork_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, expectedErr)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("CreateActionTypes failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: oerrors.OntologyManager_ActionType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CreateActionTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, err)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ActionTypeRestHandler_UpdateActionType(t *testing.T) {
	Convey("Test ActionTypeHandler UpdateActionType\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		ats := dmock.NewMockActionTypeService(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewActionTypeRestHandler(appSetting, hydra, ats, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		knID := "kn1"
		atID := "at1"
		url := "/api/ontology-manager/v1/knowledge-networks/" + knID + "/action-types/" + atID

		actionType := &interfaces.ActionType{
			ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
				ATID:   atID,
				ATName: "action1",
			},
		}

		Convey("Success UpdateActionType\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), atID).Return("action2", true, nil)
			ats.EXPECT().CheckActionTypeExistByName(gomock.Any(), knID, gomock.Any(), actionType.ATName).Return("", false, nil)
			ats.EXPECT().UpdateActionType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(actionType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Failed UpdateActionType ShouldBind Error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader([]byte("invalid json")))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(actionType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("ActionType not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), atID).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(actionType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("ActionType name already exists\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), atID).Return("oldname", true, nil)
			ats.EXPECT().CheckActionTypeExistByName(gomock.Any(), knID, gomock.Any(), actionType.ATName).Return("action1", true, nil)

			reqParamByte, _ := sonic.Marshal(actionType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("UpdateActionType failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: oerrors.OntologyManager_ActionType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), atID).Return("action1", true, nil)
			ats.EXPECT().UpdateActionType(gomock.Any(), gomock.Any(), gomock.Any()).Return(err)

			reqParamByte, _ := sonic.Marshal(actionType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ActionTypeRestHandler_DeleteActionTypes(t *testing.T) {
	Convey("Test ActionTypeHandler DeleteActionTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		ats := dmock.NewMockActionTypeService(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewActionTypeRestHandler(appSetting, hydra, ats, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		knID := "kn1"
		atIDs := "at1,at2"
		url := "/api/ontology-manager/v1/knowledge-networks/" + knID + "/action-types/" + atIDs

		Convey("Success DeleteActionTypes\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), "at1").Return("action1", true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), "at2").Return("action2", true, nil)
			ats.EXPECT().DeleteActionTypesByIDs(gomock.Any(), gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(int64(2), nil)

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

		Convey("ActionType not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), "at1").Return("", false, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("DeleteActionTypes failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: oerrors.OntologyManager_ActionType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), "at1").Return("action1", true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), "at2").Return("action2", true, nil)
			ats.EXPECT().DeleteActionTypesByIDs(gomock.Any(), gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(int64(0), err)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ActionTypeRestHandler_ListActionTypes(t *testing.T) {
	Convey("Test ActionTypeHandler ListActionTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		ats := dmock.NewMockActionTypeService(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewActionTypeRestHandler(appSetting, hydra, ats, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		knID := "kn1"
		url := "/api/ontology-manager/v1/knowledge-networks/" + knID + "/action-types"

		Convey("Success ListActionTypes\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.ActionType{}, 0, nil)

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

		Convey("ListActionTypes failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: oerrors.OntologyManager_ActionType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return(nil, 0, err)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ActionTypeRestHandler_GetActionTypes(t *testing.T) {
	Convey("Test ActionTypeHandler GetActionTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		ats := dmock.NewMockActionTypeService(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewActionTypeRestHandler(appSetting, hydra, ats, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		knID := "kn1"
		atIDs := "at1,at2"
		url := "/api/ontology-manager/v1/knowledge-networks/" + knID + "/action-types/" + atIDs

		Convey("Success GetActionTypes\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().GetActionTypesByIDs(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return([]*interfaces.ActionType{}, nil)

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

		Convey("GetActionTypes failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: oerrors.OntologyManager_ActionType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().GetActionTypesByIDs(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(nil, err)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ActionTypeRestHandler_SearchActionTypes(t *testing.T) {
	Convey("Test ActionTypeHandler SearchActionTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		ats := dmock.NewMockActionTypeService(mockCtrl)
		kns := dmock.NewMockKNService(mockCtrl)

		handler := MockNewActionTypeRestHandler(appSetting, hydra, ats, kns)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		knID := "kn1"
		url := "/api/ontology-manager/v1/knowledge-networks/" + knID + "/action-types"

		query := interfaces.ConceptsQuery{
			Limit: 10,
		}

		Convey("Success SearchActionTypes\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().SearchActionTypes(gomock.Any(), gomock.Any()).Return(interfaces.ActionTypes{}, nil)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Failed SearchActionTypes ShouldBind Error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("invalid json")))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("SearchActionTypes failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: oerrors.OntologyManager_ActionType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().SearchActionTypes(gomock.Any(), gomock.Any()).Return(interfaces.ActionTypes{}, err)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
