// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	dmock "data-model/interfaces/mock"
)

type InvalidDataDict struct {
	Name int
}

func MockNewDataDictRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra,
	dds interfaces.DataDictService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		dds:        dds,
	}
	return r
}

func Test_DataDictRestHandler_ListDataDicts(t *testing.T) {
	Convey("Test ListDataDicts\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dds := dmock.NewMockDataDictService(mockCtrl)

		handler := MockNewDataDictRestHandler(appSetting, hydra, dds)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-dicts"

		Convey("Success ListDataDicts \n", func() {
			dds.EXPECT().ListDataDicts(gomock.Any(), gomock.Any()).AnyTimes().
				Return([]interfaces.DataDict{}, int64(1), nil)

			url = url + "?direction=desc&sort=update_time&limit=1000&offset=0"
			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("offest invalid \n", func() {
			url = url + "?direction=desc&sort=update_time&limit=1000&offset=a"
			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("ListDataDicts Failed \n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: "zh-CN",
				BaseError: rest.BaseError{
					ErrorCode: derrors.DataModel_MetricModel_InternalError,
				},
			}

			dds.EXPECT().ListDataDicts(gomock.Any(), gomock.Any()).AnyTimes().
				Return([]interfaces.DataDict{}, int64(0), err)

			url = url + "?direction=desc&sort=update_time&limit=1000&offset=0"
			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictRestHandler_GetDataDicts(t *testing.T) {
	Convey("Test DictHandler GetDataDicts\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dds := dmock.NewMockDataDictService(mockCtrl)

		handler := MockNewDataDictRestHandler(appSetting, hydra, dds)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-dicts/1,2,3,4"

		Convey("GetDataDicts Success \n", func() {
			dds.EXPECT().GetDataDicts(gomock.Any(), gomock.Any()).
				AnyTimes().Return([]interfaces.DataDict{}, nil)

			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("GetDataDicts dictIDstrs invalid\n", func() {
			errUrl := "/api/mdl-data-model/v1/data-dicts/,"

			req := httptest.NewRequest(http.MethodGet, errUrl, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("GetDataDicts error\n", func() {
			httpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_DataDict_InternalError)
			dds.EXPECT().GetDataDicts(gomock.Any(), gomock.Any()).Return([]interfaces.DataDict{}, httpErr)

			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictRestHandler_CreateDataDict(t *testing.T) {
	Convey("test DictHandler CreateDataDict\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dds := dmock.NewMockDataDictService(mockCtrl)

		handler := MockNewDataDictRestHandler(appSetting, hydra, dds)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-dicts"

		dictInfos := []interfaces.DataDict{
			{
				DictID:    "1",
				DictName:  "test",
				DictType:  interfaces.DATA_DICT_TYPE_KV,
				UniqueKey: true,
				Tags:      []string{"a", "b", "c"},
				Dimension: interfaces.DATA_DICT_KV_DIMENSION,
				DictStore: "t_data_dict_item",
				Comment:   "",
			}, {
				DictID:    "2",
				DictName:  "testd",
				DictType:  interfaces.DATA_DICT_TYPE_DIMENSION,
				UniqueKey: true,
				Tags:      []string{"a", "b", "c"},
				Dimension: interfaces.DATA_DICT_KV_DIMENSION,
				DictStore: "t_dim2323424334",
				Comment:   "",
			},
		}
		Convey("CreateDataDict Success\n", func() {
			dds.EXPECT().CheckDictExistByName(gomock.Any(), gomock.Any()).AnyTimes().Return(false, nil)
			dds.EXPECT().CreateDataDict(gomock.Any(), gomock.Any()).AnyTimes().Return("1", nil)

			reqParamByte, _ := sonic.Marshal(dictInfos)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("CreateDataDict Success With Items\n", func() {
			dictItemInfos := []map[string]string{
				{"comment": "comment", "key": "key", "value": "value"},
				{"comment": "comment0", "key": "key0", "value": "value0"},
			}
			dictInfos[0].DictItems = dictItemInfos
			dictInfos[1].DictItems = dictItemInfos
			dds.EXPECT().CheckDictExistByName(gomock.Any(), gomock.Any()).AnyTimes().Return(false, nil)
			dds.EXPECT().CreateDataDict(gomock.Any(), gomock.Any()).AnyTimes().Return("1", nil)

			reqParamByte, _ := sonic.Marshal(dictInfos)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Failed CreateDataDict ShouldBind Error\n", func() {
			invalidDataDict := InvalidDataDict{
				Name: 123,
			}
			reqParamByte, _ := sonic.Marshal(invalidDataDict)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed CreateDataDict DictName nil\n", func() {
			dictDuplicateInfo := interfaces.DataDict{
				DictName: "",
				Comment:  "ds5a452",
			}
			dictInfos = append(dictInfos, dictDuplicateInfo)

			reqParamByte, _ := sonic.Marshal(dictInfos)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed CreateDataDict ValidateDuplicate DictName Error\n", func() {
			dictDuplicateInfo := interfaces.DataDict{
				DictName: "test",
				Comment:  "ds5a452",
			}
			dictInfos = append(dictInfos, dictDuplicateInfo)

			reqParamByte, _ := sonic.Marshal(dictInfos)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed CreateDataDict ValidateComment Dict Error\n", func() {
			dictCommentInfo := interfaces.DataDict{
				DictName: "test",
				Comment:  "01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789",
			}
			dictInfos = append(dictInfos, dictCommentInfo)

			reqParamByte, _ := sonic.Marshal(dictInfos)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed CreateDataDict ValidateKeyValue Error\n", func() {
			dictItemInfos := []map[string]string{
				{"comment": "comment", "key": "key", "value": "value"},
				{"comment": "comment0", "key": "key", "value": "value0"},
			}
			dictInfos[0].DictItems = dictItemInfos

			reqParamByte, _ := sonic.Marshal(dictInfos)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed CheckDictExistByName Error\n", func() {
			httpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_DataDict_Duplicated_DictName)

			dds.EXPECT().CheckDictExistByName(gomock.Any(), gomock.Any()).AnyTimes().Return(true, httpErr)

			reqParamByte, _ := sonic.Marshal(dictInfos)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed CreateDataDict Error\n", func() {
			httpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError)
			dds.EXPECT().CheckDictExistByName(gomock.Any(), gomock.Any()).AnyTimes().Return(false, nil)
			dds.EXPECT().CreateDataDict(gomock.Any(), gomock.Any()).Return("0", httpErr)

			reqParamByte, _ := sonic.Marshal(dictInfos)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictRestHandler_UpdateDataDict(t *testing.T) {
	Convey("test DictHandler UpdateDataDict\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dds := dmock.NewMockDataDictService(mockCtrl)

		handler := MockNewDataDictRestHandler(appSetting, hydra, dds)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-dicts/11"

		dictInfo := interfaces.DataDict{
			DictID:    "1",
			DictName:  "test",
			DictType:  interfaces.DATA_DICT_TYPE_DIMENSION,
			UniqueKey: true,
			Tags:      []string{"a", "b", "c"},
			Dimension: interfaces.DATA_DICT_KV_DIMENSION,
			DictStore: "t_dim234434",
			Comment:   "",
		}

		Convey("UpdateDataDict Success\n", func() {
			dds.EXPECT().UpdateDataDict(gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(dictInfo)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Failed UpdateDataDict ShouldBindJSON Error\n", func() {
			invalidDataDict := InvalidDataDict{
				Name: 123,
			}
			reqParamByte, _ := sonic.Marshal(invalidDataDict)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed UpdateDataDict ValidateDictName Error\n", func() {
			dictInfo.DictName = ""
			reqParamByte, _ := sonic.Marshal(dictInfo)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed UpdateDataDict ValidateComment Error\n", func() {
			dictInfo.Comment = "01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"
			reqParamByte, _ := sonic.Marshal(dictInfo)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed UpdateDataDict Error\n", func() {
			httpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_DataDict_InternalError)

			dds.EXPECT().UpdateDataDict(gomock.Any(), gomock.Any()).Return(httpErr)

			reqParamByte, _ := sonic.Marshal(dictInfo)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictRestHandler_DeleteDataDicts(t *testing.T) {
	Convey("test DictHandler DeleteDataDicts\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dds := dmock.NewMockDataDictService(mockCtrl)

		handler := MockNewDataDictRestHandler(appSetting, hydra, dds)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-dicts/1,2,3"

		Convey("DeleteDataDicts Success \n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.DataDict{}, nil)
			dds.EXPECT().DeleteDataDict(gomock.Any(), gomock.Any()).AnyTimes().Return(int64(1), nil)

			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("DeleteDataDicts Failed not found\n", func() {
			httpErr := rest.NewHTTPError(testCtx, http.StatusNotFound, derrors.DataModel_DataDict_DictNotFound)
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.DataDict{}, httpErr)
			dds.EXPECT().DeleteDataDict(gomock.Any(), gomock.Any()).AnyTimes().Return(int64(0), nil)

			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("DeleteDataDicts Failed \n", func() {
			httpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_DataDict_InternalError)

			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.DataDict{}, nil)
			dds.EXPECT().DeleteDataDict(gomock.Any(), gomock.Any()).AnyTimes().Return(int64(100), httpErr)

			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
