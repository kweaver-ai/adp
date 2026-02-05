package driveradapters

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"ontology-query/common"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
	dmock "ontology-query/interfaces/mock"
)

func Test_RestHandler_GetActionsInActionTypeByIn(t *testing.T) {
	Convey("Test RestHandler GetActionsInActionTypeByIn", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		ats := dmock.NewMockActionTypeService(mockCtrl)
		kns := dmock.NewMockKnowledgeNetworkService(mockCtrl)
		ots := dmock.NewMockObjectTypeService(mockCtrl)

		handler := MockNewRestHandler(appSetting, hydra, ats, kns, ots)
		handler.RegisterPublic(engine)

		knID := "kn1"
		atID := "at1"
		url := "/api/ontology-query/in/v1/knowledge-networks/" + knID + "/action-types/" + atID

		actionQuery := interfaces.ActionQuery{
			InstanceIdentities: []map[string]any{
				{"id": "1"},
			},
		}

		Convey("成功 - 获取行动数据", func() {
			ats.EXPECT().GetActionsByActionTypeID(gomock.Any(), gomock.Any()).Return(interfaces.Actions{
				Actions: []interfaces.ActionParam{
					{
						Parameters: map[string]any{"key": "value"},
					},
				},
				TotalCount: 1,
			}, nil)

			reqParamByte, _ := sonic.Marshal(actionQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_TYPE, "user")
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("失败 - 参数绑定失败", func() {
			reqParamByte := []byte("invalid json")
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_TYPE, "user")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("失败 - 唯一标识为空", func() {
			invalidQuery := interfaces.ActionQuery{
				InstanceIdentities: []map[string]any{},
			}
			reqParamByte, _ := sonic.Marshal(invalidQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_TYPE, "user")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("失败 - Service返回错误", func() {
			ats.EXPECT().GetActionsByActionTypeID(gomock.Any(), gomock.Any()).Return(interfaces.Actions{}, rest.NewHTTPError(context.TODO(), http.StatusInternalServerError, oerrors.OntologyQuery_ActionType_InternalError))

			reqParamByte, _ := sonic.Marshal(actionQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_TYPE, "user")
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_RestHandler_GetActionsInActionTypeByEx(t *testing.T) {
	Convey("Test RestHandler GetActionsInActionTypeByEx", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		ats := dmock.NewMockActionTypeService(mockCtrl)
		kns := dmock.NewMockKnowledgeNetworkService(mockCtrl)
		ots := dmock.NewMockObjectTypeService(mockCtrl)

		handler := MockNewRestHandler(appSetting, hydra, ats, kns, ots)
		handler.RegisterPublic(engine)

		knID := "kn1"
		atID := "at1"
		url := "/api/ontology-query/v1/knowledge-networks/" + knID + "/action-types/" + atID

		actionQuery := interfaces.ActionQuery{
			InstanceIdentities: []map[string]any{
				{"id": "1"},
			},
		}

		Convey("成功 - Token验证通过，获取行动数据", func() {
			visitor := rest.Visitor{
				ID:   "user1",
				Type: rest.VisitorType_User,
			}
			hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(visitor, nil)
			ats.EXPECT().GetActionsByActionTypeID(gomock.Any(), gomock.Any()).Return(interfaces.Actions{
				Actions: []interfaces.ActionParam{
					{
						Parameters: map[string]any{"key": "value"},
					},
				},
				TotalCount: 1,
			}, nil)

			reqParamByte, _ := sonic.Marshal(actionQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("失败 - Token验证失败", func() {
			hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(rest.Visitor{}, rest.NewHTTPError(context.TODO(), http.StatusUnauthorized, rest.PublicError_Unauthorized))

			reqParamByte, _ := sonic.Marshal(actionQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusUnauthorized)
		})
	})
}

func Test_RestHandler_GetActionsInActionType(t *testing.T) {
	Convey("Test RestHandler GetActionsInActionType", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		ats := dmock.NewMockActionTypeService(mockCtrl)
		kns := dmock.NewMockKnowledgeNetworkService(mockCtrl)
		ots := dmock.NewMockObjectTypeService(mockCtrl)

		handler := MockNewRestHandler(appSetting, hydra, ats, kns, ots)

		knID := "kn1"
		atID := "at1"
		url := "/api/ontology-query/v1/knowledge-networks/" + knID + "/action-types/" + atID

		actionQuery := interfaces.ActionQuery{
			InstanceIdentities: []map[string]any{
				{"id": "1"},
			},
		}

		Convey("成功 - 参数验证通过，获取行动数据", func() {
			ats.EXPECT().GetActionsByActionTypeID(gomock.Any(), gomock.Any()).Return(interfaces.Actions{
				Actions: []interfaces.ActionParam{
					{
						Parameters: map[string]any{"key": "value"},
					},
				},
				TotalCount: 1,
			}, nil)

			reqParamByte, _ := sonic.Marshal(actionQuery)
			req := httptest.NewRequest(http.MethodPost, url+"?branch=main", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "kn_id", Value: knID},
				{Key: "at_id", Value: atID},
			}

			visitor := rest.Visitor{
				ID:   "user1",
				Type: rest.VisitorType_User,
			}
			handler.GetActionsInActionType(c, visitor)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("失败 - includeTypeInfo参数无效", func() {
			reqParamByte, _ := sonic.Marshal(actionQuery)
			req := httptest.NewRequest(http.MethodPost, url+"?include_type_info=invalid", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "kn_id", Value: knID},
				{Key: "at_id", Value: atID},
			}

			visitor := rest.Visitor{
				ID:   "user1",
				Type: rest.VisitorType_User,
			}
			handler.GetActionsInActionType(c, visitor)

			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("失败 - Method Override无效", func() {
			reqParamByte, _ := sonic.Marshal(actionQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "POST")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "kn_id", Value: knID},
				{Key: "at_id", Value: atID},
			}

			visitor := rest.Visitor{
				ID:   "user1",
				Type: rest.VisitorType_User,
			}
			handler.GetActionsInActionType(c, visitor)

			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})
	})
}
