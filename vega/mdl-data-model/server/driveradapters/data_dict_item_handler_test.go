// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
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

func MockNewDataDictItemRestHandler(appSetting *common.AppSetting,
	hydra rest.Hydra,
	dds interfaces.DataDictService,
	ddis interfaces.DataDictItemsService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		hydra:      hydra,
		dds:        dds,
		ddis:       ddis,
	}
	return r
}

func Test_DataDictItemRestHandler_GetDataDictItems(t *testing.T) {
	Convey("Test DictItemHandler GetDataDictItems\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dds := dmock.NewMockDataDictService(mockCtrl)
		ddis := dmock.NewMockDataDictItemsService(mockCtrl)

		handler := MockNewDataDictItemRestHandler(appSetting, hydra, dds, ddis)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-dicts/11/items"

		Convey("Success ListDataDictItems \n", func() {

			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.DataDict{}, nil)
			ddis.EXPECT().ListDataDictItems(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]map[string]string{}, 3, nil)

			url = url + "?direction=asc&sort=key&limit=1000&offset=0"
			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("offest invalid \n", func() {
			url = url + "?direction=asc&sort=key&limit=1000&offset=a"
			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("ListDataDicts Failed \n", func() {
			httpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError)

			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.DataDict{}, nil)
			ddis.EXPECT().ListDataDictItems(gomock.Any(), gomock.Any(), gomock.Any()).Return([]map[string]string{}, 0, httpErr)

			url = url + "?direction=asc&sort=key&limit=1000&offset=0"
			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictItemRestHandler_ExportDataDictItems(t *testing.T) {
	Convey("Test DictItemHandler ExportDataDictItems\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dds := dmock.NewMockDataDictService(mockCtrl)
		ddis := dmock.NewMockDataDictItemsService(mockCtrl)

		handler := MockNewDataDictItemRestHandler(appSetting, hydra, dds, ddis)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-dicts/11/items?format=csv"

		Convey("Success KV ExportDataDictItems \n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{Dimension: interfaces.DATA_DICT_KV_DIMENSION}, nil)
			ddis.EXPECT().GetKVDictItems(gomock.Any(), gomock.Any()).Return([]map[string]string{}, nil)

			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Success Dimension ExportDataDictItems \n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{DictType: interfaces.DATA_DICT_TYPE_DIMENSION}, nil)
			ddis.EXPECT().GetDimensionDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]map[string]string{}, nil)

			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("KV ExportDataDictItems Failed\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{Dimension: interfaces.DATA_DICT_KV_DIMENSION}, nil)
			ddis.EXPECT().GetKVDictItems(gomock.Any(), gomock.Any()).Return([]map[string]string{}, fmt.Errorf("some error"))

			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Dimension ExportDataDictItems Failed\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{DictType: interfaces.DATA_DICT_TYPE_DIMENSION}, nil)
			ddis.EXPECT().GetDimensionDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]map[string]string{}, fmt.Errorf("some error"))

			req := httptest.NewRequest(http.MethodGet, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictItemRestHandler_CreateDataDictItem(t *testing.T) {
	Convey("test DictItemHandler CreateDataDictItem\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dds := dmock.NewMockDataDictService(mockCtrl)
		ddis := dmock.NewMockDataDictItemsService(mockCtrl)

		handler := MockNewDataDictItemRestHandler(appSetting, hydra, dds, ddis)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-dicts/11/items"
		dictItemInfo := interfaces.KvDictItem{
			Key:     "test key",
			Value:   "test value",
			Comment: "",
		}

		Convey("CreateDataDictItem Success\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{}, nil)
			ddis.EXPECT().CreateDataDictItem(gomock.Any(), gomock.Any(), gomock.Any()).Return("1", nil)

			reqParamByte, _ := sonic.Marshal(dictItemInfo)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)

			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Failed CreateDataDictItem ShouldBind Error\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{}, nil)

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

		Convey("Failed CreateDataDictItem ValidateKeyValue Error\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{DictType: interfaces.DATA_DICT_TYPE_KV}, nil)
			dictItemInfo.Key = ""
			reqParamByte, _ := sonic.Marshal(dictItemInfo)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)

			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed CreateDataDictItem ValidateComment Error\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{}, nil)
			dictItemInfo.Comment = "01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"
			reqParamByte, _ := sonic.Marshal(dictItemInfo)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed CreateDataDictItem Error\n", func() {
			httpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError)

			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{}, nil)
			ddis.EXPECT().CreateDataDictItem(gomock.Any(), gomock.Any(), gomock.Any()).Return("0", httpErr)

			reqParamByte, _ := sonic.Marshal(dictItemInfo)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictItemRestHandler_ImportDataDictItems(t *testing.T) {
	Convey("test DictItemHandler ImportDataDictItems\n", t, func() {

		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dds := dmock.NewMockDataDictService(mockCtrl)
		ddis := dmock.NewMockDataDictItemsService(mockCtrl)

		handler := MockNewDataDictItemRestHandler(appSetting, hydra, dds, ddis)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-dicts/11/items"

		DIMENSION := interfaces.Dimension{
			Keys: []interfaces.DimensionItem{
				{
					ID:   "f_key527800737703501980",
					Name: "department",
				},
				{
					ID:   "f_key527800737703567516",
					Name: "group",
				},
			},
			Values: []interfaces.DimensionItem{
				{
					ID:   "f_value527826197581701276",
					Name: "name",
				},
			},
		}

		Convey("ImportDataDictItems Success\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{Dimension: interfaces.DATA_DICT_KV_DIMENSION}, nil)
			ddis.EXPECT().ImportDataDictItems(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.ImportMode_Normal).Return(nil)

			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			_ = writer.WriteField("file_name", "test.csv")
			fileWriter, _ := writer.CreateFormFile("items_file", "test.csv")

			csvData := "key,value\r\nvalue1,value2\r\nvalue3,value4\r\n"
			buf := bytes.NewBufferString(csvData)
			_, _ = io.Copy(fileWriter, buf)
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, url, &body)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, writer.FormDataContentType())
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("ImportDataDictItems Failed\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{Dimension: interfaces.DATA_DICT_KV_DIMENSION}, nil)
			ddis.EXPECT().ImportDataDictItems(gomock.Any(), gomock.Any(), gomock.Any(), interfaces.ImportMode_Normal).
				Return(rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError))

			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			_ = writer.WriteField("file_name", "test.csv")
			fileWriter, _ := writer.CreateFormFile("items_file", "test.csv")

			csvData := "key,value\r\nvalue1,value2\r\nvalue3,value4\r\n"
			buf := bytes.NewBufferString(csvData)
			_, _ = io.Copy(fileWriter, buf)
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, url, &body)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, writer.FormDataContentType())
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("ImportDataDictItems FormFile failed\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).
				Return(interfaces.DataDict{Dimension: interfaces.DATA_DICT_KV_DIMENSION}, nil)

			var c *gin.Context
			patch := ApplyMethodReturn(c, "FormFile", &multipart.FileHeader{}, fmt.Errorf("error"))
			defer patch.Reset()

			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			_ = writer.WriteField("file_name", "test.csv")
			fileWriter, _ := writer.CreateFormFile("items_file", "test.csv")

			csvData := "key,value\r\nvalue1,value2\r\nvalue3,value4\r\n"
			buf := bytes.NewBufferString(csvData)
			_, _ = io.Copy(fileWriter, buf)
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, url, &body)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, writer.FormDataContentType())
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("ImportDataDictItems Open failed\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{Dimension: interfaces.DATA_DICT_KV_DIMENSION}, nil)

			var m *multipart.FileHeader
			patch := ApplyMethodReturn(m, "Open", nil, fmt.Errorf("error"))
			defer patch.Reset()

			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			_ = writer.WriteField("file_name", "test.csv")
			fileWriter, _ := writer.CreateFormFile("items_file", "test.csv")

			csvData := "key,value\r\nvalue1,value2\r\nvalue3,value4\r\n"
			buf := bytes.NewBufferString(csvData)
			_, _ = io.Copy(fileWriter, buf)
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, url, &body)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, writer.FormDataContentType())
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("ImportDataDictItem ReadString 1 failed\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).
				Return(interfaces.DataDict{Dimension: interfaces.DATA_DICT_KV_DIMENSION}, nil)

			var a *bufio.Reader
			patch := ApplyMethodReturn(a, "ReadString", "", fmt.Errorf("ReadString error"))
			defer patch.Reset()

			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			_ = writer.WriteField("file_name", "test.csv")
			fileWriter, _ := writer.CreateFormFile("items_file", "test.csv")

			csvData := "key,value\r\nvalue1,value2\r\nvalue3,value4\r\n"
			buf := bytes.NewBufferString(csvData)
			_, _ = io.Copy(fileWriter, buf)
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, url, &body)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, writer.FormDataContentType())
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("ImportDataDictItems ReadString 2 failed\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{Dimension: interfaces.DATA_DICT_KV_DIMENSION}, nil)

			var cnt = 0
			var a *bufio.Reader
			patch := ApplyMethod(a, "ReadString",
				func() (string, error) {
					if cnt == 0 {
						cnt++
						return "", nil
					}
					return "", fmt.Errorf("ReadString error")
				},
			)
			defer patch.Reset()

			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			_ = writer.WriteField("file_name", "test.csv")
			fileWriter, _ := writer.CreateFormFile("items_file", "test.csv")

			csvData := "key,value\r\nvalue1,value2\r\nvalue3,value4\r\n"
			buf := bytes.NewBufferString(csvData)
			_, _ = io.Copy(fileWriter, buf)
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, url, &body)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, writer.FormDataContentType())
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("ImportDataDictItems lack of keys\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{Dimension: interfaces.DATA_DICT_KV_DIMENSION}, nil)

			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			_ = writer.WriteField("file_name", "test.csv")
			fileWriter, _ := writer.CreateFormFile("items_file", "test.csv")

			csvData := "keya,value\r\nvalue1,value2\r\nvalue3,value4\r\n"
			buf := bytes.NewBufferString(csvData)
			_, _ = io.Copy(fileWriter, buf)
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, url, &body)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, writer.FormDataContentType())
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("ImportDataDictItems lack of row values\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{Dimension: interfaces.DATA_DICT_KV_DIMENSION}, nil)

			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			_ = writer.WriteField("file_name", "test.csv")
			fileWriter, _ := writer.CreateFormFile("items_file", "test.csv")

			csvData := "key,value\r\nvalue1value2\r\nvalue3,value4\r\n"
			buf := bytes.NewBufferString(csvData)
			_, _ = io.Copy(fileWriter, buf)
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, url, &body)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, writer.FormDataContentType())
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("ImportDataDictItems kv dict duplicate keys\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{Dimension: interfaces.DATA_DICT_KV_DIMENSION, UniqueKey: true}, nil)

			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			_ = writer.WriteField("file_name", "test.csv")
			fileWriter, _ := writer.CreateFormFile("items_file", "test.csv")

			csvData := "key,value\r\nvalue1,value2\r\nvalue1,value4\r\n"
			buf := bytes.NewBufferString(csvData)
			_, _ = io.Copy(fileWriter, buf)
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, url, &body)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, writer.FormDataContentType())
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("ImportDataDictItems dimension dict duplicate keys\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{
				DictType:  interfaces.DATA_DICT_TYPE_DIMENSION,
				Dimension: DIMENSION,
				UniqueKey: true}, nil)

			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			_ = writer.WriteField("file_name", "test.csv")
			fileWriter, _ := writer.CreateFormFile("items_file", "test.csv")

			csvData := "department,group,name\r\nkey1,key2,value1\r\nkey1,key2,value4\r\n"
			buf := bytes.NewBufferString(csvData)
			_, _ = io.Copy(fileWriter, buf)
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, url, &body)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, writer.FormDataContentType())
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func Test_DataDictItemRestHandler_UpdateDataDictItem(t *testing.T) {
	Convey("test DictItemHandler UpdateDataDictItem\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dds := dmock.NewMockDataDictService(mockCtrl)
		ddis := dmock.NewMockDataDictItemsService(mockCtrl)

		handler := MockNewDataDictItemRestHandler(appSetting, hydra, dds, ddis)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-dicts/11/items/111"

		dictItemInfo := interfaces.KvDictItem{
			Key:     "test key",
			Value:   "test value",
			Comment: "",
		}
		Convey("UpdateDataDictItem Success\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).
				Return(interfaces.DataDict{
					DictType:  interfaces.DATA_DICT_TYPE_DIMENSION,
					Dimension: interfaces.DATA_DICT_KV_DIMENSION}, nil)
			ddis.EXPECT().UpdateDataDictItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			reqParamByte, _ := sonic.Marshal(dictItemInfo)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Failed UpdateDataDictItem ShouldBind Error\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{}, nil)

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

		Convey("Failed UpdateDataDictItem ValidateKeyValue Error\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).
				Return(interfaces.DataDict{DictType: interfaces.DATA_DICT_TYPE_KV}, nil)

			dictItemInfo.Key = ""
			reqParamByte, _ := sonic.Marshal(dictItemInfo)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed UpdateDataDictItem ValidateCommentError\n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{}, nil)

			dictItemInfo.Comment = "01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"
			reqParamByte, _ := sonic.Marshal(dictItemInfo)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("UpdateDataDictItem Failed\n", func() {
			httpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_DataDict_InternalError)

			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{}, nil)
			ddis.EXPECT().UpdateDataDictItem(gomock.Any(),
				gomock.Any(), gomock.Any(), gomock.Any()).Return(httpErr)
			reqParamByte, _ := sonic.Marshal(dictItemInfo)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictItemRestHandler_DeleteDataDictItems(t *testing.T) {
	Convey("test DictItemHandler DeleteDataDictItems\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dds := dmock.NewMockDataDictService(mockCtrl)
		ddis := dmock.NewMockDataDictItemsService(mockCtrl)

		handler := MockNewDataDictItemRestHandler(appSetting, hydra, dds, ddis)
		handler.RegisterPublic(engine)

		hydra.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(rest.Visitor{}, nil)

		url := "/api/mdl-data-model/v1/data-dicts/11/items/1,2,3"

		Convey("DeleteDataDictItems Success \n", func() {
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{}, nil)
			ddis.EXPECT().GetDictItemsByItemIDs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]map[string]string{}, nil)
			ddis.EXPECT().DeleteDataDictItem(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("DeleteDataDictItems GetDataDictByID Failed \n", func() {
			httpErr := rest.NewHTTPError(testCtx, http.StatusNotFound, derrors.DataModel_DataDict_DictNotFound)

			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{}, httpErr)

			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("DeleteDataDictItems GetDictItemsByItemIDs Failed \n", func() {
			httpErr := rest.NewHTTPError(testCtx, http.StatusNotFound, derrors.DataModel_DataDict_DictItemNotFound)
			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{}, nil)
			ddis.EXPECT().GetDictItemsByItemIDs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]map[string]string{}, httpErr)

			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("DeleteDataDictItems Failed \n", func() {
			httpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataDict_InternalError)

			dds.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{}, nil)
			ddis.EXPECT().GetDictItemsByItemIDs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]map[string]string{}, nil)
			ddis.EXPECT().DeleteDataDictItem(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(httpErr)

			req := httptest.NewRequest(http.MethodDelete, url, bytes.NewReader(nil))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
