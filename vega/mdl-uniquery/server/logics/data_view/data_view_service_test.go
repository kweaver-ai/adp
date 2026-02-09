// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_view

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	cond "uniquery/common/condition"
	cmock "uniquery/common/condition/mock"
	vopt "uniquery/common/value_opt"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	dtype "uniquery/interfaces/data_type"
	mock "uniquery/interfaces/mock"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

func MockNewDataViewService(appSetting *common.AppSetting, dvAccess interfaces.DataViewAccess,
	ibAccess interfaces.IndexBaseAccess, osAccess interfaces.OpenSearchAccess, ps interfaces.PermissionService) *dataViewService {
	return &dataViewService{
		appSetting: appSetting,
		dvAccess:   dvAccess,
		ibAccess:   ibAccess,
		osAccess:   osAccess,
		ps:         ps,
	}
}

// var requiredFields = []*cond.ViewField{
// 	{Name: "@timestamp"},
// 	{Name: "__data_type"},
// 	{Name: "__index_base"},
// 	{Name: "__write_time"},
// 	{Name: "__category"},
// 	{Name: "tags"},
// 	{Name: "__id"},
// 	{Name: "__routing"},
// 	{Name: "__tsid"},
// 	{Name: "__pipeline_id"},
// }

// func TestSimulate(t *testing.T) {
// 	Convey("Test Simulate", t, func() {
// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		dvaMock := mock.NewMockDataViewAccess(mockCtrl)
// 		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
// 		osaMock := mock.NewMockOpenSearchAccess(mockCtrl)
// 		psMock := mock.NewMockPermissionService(mockCtrl)
// 		appSetting := &common.AppSetting{}

// 		dvsMock := MockNewDataViewService(appSetting, dvaMock, ibaMock, osaMock, psMock)
// 		ops := []interfaces.ResourceOps{
// 			{
// 				ResourceID: interfaces.RESOURCE_ID_ALL,
// 				Operations: []string{interfaces.OPERATION_TYPE_CREATE},
// 			},
// 		}

// 		fields := requiredFields
// 		fields = append(fields, &cond.ViewField{Name: "zzz"})
// 		query := &interfaces.DataViewSimulateQuery{
// 			// DataSource: map[string]any{"index_base": []interfaces.SimpleIndexBase{
// 			// 	{BaseType: "x"},
// 			// 	{BaseType: "y"},
// 			// }},
// 			Fields:     fields,
// 		}

// 		indicesResult := map[string]map[string]interfaces.Indice{
// 			"indices": {
// 				"x": {
// 					IndexName: "x",
// 					ShardNum:  3,
// 				},
// 				"y": {
// 					IndexName: "y",
// 					ShardNum:  3,
// 				},
// 			},
// 		}

// 		var dslBuf bytes.Buffer
// 		dslStr := `
// 			{
// 				"from": 0,
// 				"size": 500,
// 				"sort": [
// 					{
// 					"@timestamp": "asc"
// 					}
// 				],
// 				"query": {
// 					"bool": {
// 					"filter": [
// 						{
// 						"range": {
// 							"@timestamp": {}
// 						}
// 						}
// 					],
// 					"must": []
// 					}
// 				}
// 			}
// 			`
// 		dslBuf.WriteString(dslStr)

// 		totalResult := `
// 		{
// 			"count": 1042,
// 			"_shards": {
// 				"total": 1,
// 				"successful": 1,
// 				"skipped": 0,
// 				"failed": 0
// 			}
// 		}
// 		`
// 		totalBytes := []byte(totalResult)

// 		Convey("Simulate failed, mapstructure failed", func() {
// 			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)

// 			query1 := &interfaces.DataViewSimulateQuery{
// 				// DataSource: map[string]any{"index_base": []string{"x"}},
// 				// FieldScope: 0,
// 			}

// 			_, httpErr := dvsMock.Simulate(testCtx, query1)
// 			So(httpErr, ShouldNotBeNil)
// 		})

// 		Convey("Simulate failed, get index base error", func() {
// 			expectedErr := errors.New("some errors")
// 			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
// 				uerrors.Uniquery_DataView_InternalError_GetIndexBaseByTypeFailed).WithErrorDetails(expectedErr.Error())

// 			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
// 			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

// 			_, httpErr := dvsMock.Simulate(testCtx, query)
// 			So(httpErr, ShouldResemble, expectedHttpErr)
// 		})

// 		Convey("Simulate failed, get indices error", func() {
// 			expectedErr := errors.New("some errors")
// 			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
// 				uerrors.Uniquery_DataView_InternalError_GetIndicesFailed).WithErrorDetails(expectedErr.Error())

// 			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
// 			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, nil)
// 			ibaMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(indicesResult, 200, expectedErr)

// 			_, httpErr := dvsMock.Simulate(testCtx, query)
// 			So(httpErr, ShouldResemble, expectedHttpErr)
// 		})

// 		Convey("Simulate failed, field type conflict", func() {
// 			baseInfos := []interfaces.IndexBase{
// 				{
// 					Mappings: interfaces.Mappings{
// 						DynamicMappings: []interfaces.IndexBaseField{
// 							{
// 								Field: "zzz",
// 								Type:  "long",
// 							},
// 						},
// 						UserDefinedMappings: []interfaces.IndexBaseField{
// 							{
// 								Field: "zzz",
// 								Type:  "text",
// 							},
// 						},
// 					},
// 				},
// 			}

// 			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
// 			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(baseInfos, nil)

// 			query1 := &interfaces.DataViewSimulateQuery{
// 				// DataSource: map[string]any{"index_base": []interfaces.SimpleIndexBase{
// 				// 	{BaseType: "x"},
// 				// 	{BaseType: "y"},
// 				// }},
// 				// FieldScope: 0,
// 				Fields:     fields,
// 			}

// 			_, httpErr := dvsMock.Simulate(testCtx, query1)
// 			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_DataView_FieldTypeConflict)

// 		})

// 		Convey("Simulate failed, convert To DSL error", func() {
// 			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
// 			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, nil)

// 			expectedErr := errors.New("some errors")
// 			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest,
// 				uerrors.Uniquery_DataView_InvalidParameter_Filters).WithErrorDetails(expectedErr.Error())

// 			patch1 := ApplyFunc(buildDSL,
// 				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView) (bytes.Buffer, error) {
// 					return bytes.Buffer{}, expectedHttpErr
// 				},
// 			)
// 			defer patch1.Reset()

// 			_, httpErr := dvsMock.Simulate(testCtx, query)
// 			So(httpErr, ShouldResemble, expectedHttpErr)
// 		})

// 		Convey("Simulate failed, search submit error", func() {
// 			query := &interfaces.DataViewSimulateQuery{
// 				// DataSource: map[string]any{
// 				// 	"type": "index_base",
// 				// 	"index_base": []any{
// 				// 		interfaces.SimpleIndexBase{
// 				// 			BaseType: "x",
// 				// 		},
// 				// 	},
// 				// },
// 				// FieldScope: 1,
// 			}

// 			expectedErr := errors.New("some errors")
// 			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
// 				uerrors.Uniquery_InternalError_SearchSubmitFailed).WithErrorDetails(expectedErr.Error())

// 			patch1 := ApplyFunc(buildDSL,
// 				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView) (bytes.Buffer, error) {
// 					return bytes.Buffer{}, nil
// 				},
// 			)
// 			defer patch1.Reset()

// 			expectedBaseInfos := []interfaces.IndexBase{
// 				{
// 					SimpleIndexBase: interfaces.SimpleIndexBase{
// 						BaseType: "1a",
// 					},
// 				},

// 				{
// 					SimpleIndexBase: interfaces.SimpleIndexBase{
// 						BaseType: "2b",
// 					},
// 				},
// 			}
// 			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
// 			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(expectedBaseInfos, nil)
// 			ibaMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(indicesResult, 200, nil)
// 			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
// 				gomock.Any(), gomock.Any()).Return([]byte{}, 0, expectedErr)

// 			_, httpErr := dvsMock.Simulate(testCtx, query)
// 			So(httpErr, ShouldResemble, expectedHttpErr)
// 		})

// 		Convey("Simulate failed, convert to view uniRepsonse error", func() {
// 			expectedErr := errors.New("some errors")
// 			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
// 				uerrors.Uniquery_DataView_InternalError_ConvertToViewUniResponseFailed).WithErrorDetails(expectedErr.Error())

// 			patch1 := ApplyFunc(buildDSL,
// 				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView) (bytes.Buffer, error) {
// 					return dslBuf, nil
// 				},
// 			)
// 			defer patch1.Reset()

// 			patch4 := ApplyFunc(convertToViewUniResponse,
// 				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView, content []byte) (*interfaces.ViewUniResponseV2, error) {
// 					return nil, expectedHttpErr
// 				},
// 			)
// 			defer patch4.Reset()

// 			expectedBaseInfos := []interfaces.IndexBase{
// 				{
// 					SimpleIndexBase: interfaces.SimpleIndexBase{
// 						BaseType: "1a",
// 					},
// 				},

// 				{
// 					SimpleIndexBase: interfaces.SimpleIndexBase{
// 						BaseType: "2b",
// 					},
// 				},
// 			}
// 			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
// 			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(expectedBaseInfos, nil)
// 			ibaMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(indicesResult, 200, nil)
// 			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
// 				gomock.Any(), gomock.Any()).Return([]byte{}, 0, nil)
// 			osaMock.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(totalBytes, 200, nil)

// 			_, httpErr := dvsMock.Simulate(testCtx, query)
// 			So(httpErr, ShouldResemble, expectedHttpErr)
// 		})

// 		Convey("Simulate success, field scope is 0", func() {
// 			patch1 := ApplyFunc(buildDSL,
// 				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView) (bytes.Buffer, error) {
// 					return dslBuf, nil
// 				},
// 			)
// 			defer patch1.Reset()

// 			patch4 := ApplyFunc(convertToViewUniResponse,
// 				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView, content []byte) (*interfaces.ViewUniResponseV2, error) {
// 					return &interfaces.ViewUniResponseV2{}, nil
// 				},
// 			)
// 			defer patch4.Reset()

// 			expectedBaseInfos := []interfaces.IndexBase{
// 				{
// 					SimpleIndexBase: interfaces.SimpleIndexBase{
// 						BaseType: "1a",
// 					},
// 				},

// 				{
// 					SimpleIndexBase: interfaces.SimpleIndexBase{
// 						BaseType: "2b",
// 					},
// 				},
// 			}
// 			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
// 			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(expectedBaseInfos, nil)
// 			ibaMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(indicesResult, 200, nil)
// 			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
// 				gomock.Any(), gomock.Any()).Return([]byte{}, 0, nil)
// 			osaMock.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(totalBytes, 200, nil)

// 			result, httpErr := dvsMock.Simulate(testCtx, query)
// 			So(result, ShouldNotBeNil)
// 			So(httpErr, ShouldBeNil)
// 		})

// 		Convey("Simulate success, field scope is 1", func() {
// 			patch1 := ApplyFunc(buildDSL,
// 				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView) (bytes.Buffer, error) {
// 					return dslBuf, nil
// 				},
// 			)
// 			defer patch1.Reset()

// 			patch4 := ApplyFunc(convertToViewUniResponse,
// 				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView, content []byte) (*interfaces.ViewUniResponseV2, error) {
// 					return &interfaces.ViewUniResponseV2{}, nil
// 				},
// 			)
// 			defer patch4.Reset()

// 			expectedBaseInfos := []interfaces.IndexBase{
// 				{
// 					SimpleIndexBase: interfaces.SimpleIndexBase{
// 						BaseType: "1a",
// 					},
// 					Mappings: interfaces.Mappings{
// 						DynamicMappings: []interfaces.IndexBaseField{
// 							{
// 								Field: "zzz",
// 								Type:  "long",
// 							},
// 						},
// 					},
// 				},

// 				{
// 					SimpleIndexBase: interfaces.SimpleIndexBase{
// 						BaseType: "2b",
// 					},
// 				},
// 			}
// 			psMock.EXPECT().GetResourcesOperations(gomock.Any(), gomock.Any(), gomock.Any()).Return(ops, nil)
// 			ibaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(expectedBaseInfos, nil)
// 			ibaMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(indicesResult, 200, nil)
// 			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
// 				gomock.Any(), gomock.Any()).Return([]byte{}, 0, nil)
// 			osaMock.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(totalBytes, 200, nil)

// 			query1 := &interfaces.DataViewSimulateQuery{
// 				// DataSource: map[string]any{"index_base": []interfaces.SimpleIndexBase{
// 				// 	{BaseType: "x"},
// 				// 	{BaseType: "y"},
// 				// }},
// 				// FieldScope: 1,
// 			}

// 			result, httpErr := dvsMock.Simulate(testCtx, query1)
// 			So(result, ShouldNotBeNil)
// 			So(httpErr, ShouldBeNil)
// 		})
// 	})
// }

func TestGetSingleViewData(t *testing.T) {
	Convey("Test GetSingleViewData", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		dvaMock := mock.NewMockDataViewAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		osaMock := mock.NewMockOpenSearchAccess(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		appSetting := &common.AppSetting{}

		dvsMock := MockNewDataViewService(appSetting, dvaMock, ibaMock, osaMock, psMock)

		view := interfaces.DataView{
			ViewID:    "1",
			QueryType: interfaces.QueryType_IndexBase,
			Type:      interfaces.ViewType_Atomic,
		}
		query := &interfaces.DataViewQueryV1{}

		indicesResult := map[string]map[string]interfaces.Indice{
			"indices": {
				"x": {
					IndexName: "x",
					ShardNum:  3,
				},
				"y": {
					IndexName: "y",
					ShardNum:  3,
				},
			},
		}
		var dslBuf bytes.Buffer
		dslStr := `
			{
				"from": 0,
				"size": 500,
				"sort": [
					{
					"@timestamp": "asc"
					}
				],
				"query": {
					"bool": {
					"filter": [
						{
						"range": {
							"@timestamp": {}
						}
						}
					],
					"must": []
					}
				}
			}
			`
		dslBuf.WriteString(dslStr)

		totalResult := `
		{
			"count": 1042,
			"_shards": {
				"total": 1,
				"successful": 1,
				"skipped": 0,
				"failed": 0
			}
		}
		`
		totalBytes := []byte(totalResult)

		expectedViewRes := []*interfaces.DataView{
			{
				ViewID:    "1",
				QueryType: interfaces.QueryType_IndexBase,
				Type:      interfaces.ViewType_Atomic,
				// DataSource: map[string]any{
				// 	"type": "index_base",
				// 	"index_base": []any{
				// 		interfaces.SimpleIndexBase{
				// 			BaseType: "x",
				// 		},
				// 	},
				// },
			},
			{
				ViewID:    "2",
				QueryType: interfaces.QueryType_IndexBase,
				Type:      interfaces.ViewType_Atomic,
				// DataSource: map[string]any{
				// 	"type": "index_base",
				// 	"index_base": []any{
				// 		interfaces.SimpleIndexBase{
				// 			BaseType: "x",
				// 		},
				// 	},
				// },
			},
		}

		Convey("GetSingleViewData failed, get data views by IDs error", func() {
			expectedErr := errors.New("some error")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_InternalError_GetDataViewByIDFailed).WithErrorDetails(expectedErr.Error())

			psMock.EXPECT().CheckPermissionWithResult(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			dvaMock.EXPECT().GetDataViewsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			_, httpErr := dvsMock.GetSingleViewData(testCtx, view.ViewID, query)
			So(httpErr, ShouldResemble, expectedHttpErr)
		})

		Convey("GetSingleViewData failed, data view not found", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_DataViewNotFound)

			expectedRes := []*interfaces.DataView{}
			psMock.EXPECT().CheckPermissionWithResult(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			dvaMock.EXPECT().GetDataViewsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedRes, nil)

			_, httpErr := dvsMock.GetSingleViewData(testCtx, view.ViewID, query)
			So(httpErr, ShouldNotBeNil)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, expectedHttpErr.BaseError.ErrorCode)
		})

		Convey("GetSingleViewData failed, global condition not in view fields", func() {
			expectedRes := []*interfaces.DataView{
				{
					ViewID: "1",
					Fields: []*cond.ViewField{
						{Name: "x", Type: dtype.DataType_Integer},
					},
					QueryType: interfaces.QueryType_IndexBase,
					Type:      interfaces.ViewType_Atomic,
				},
			}
			psMock.EXPECT().CheckPermissionWithResult(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			dvaMock.EXPECT().GetDataViewsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedRes, nil)

			view := &interfaces.DataView{
				ViewID: "1",
				// FieldScope: 1,
				// Condition: &cond.CondCfg{
				// 	Name:        "y",
				// 	Operation:   cond.OperationLike,
				// 	ValueOptCfg: vopt.ValueOptCfg{Value: "ai", ValueFrom: "const"},
				// },
				// Fields: []*cond.ViewField{
				// 	{Name: "x", Type: dtype.DATATYPE_TEXT},
				// 	{Name: "y", Type: dtype.DATATYPE_TEXT},
				// },
			}

			query1 := &interfaces.DataViewQueryV1{
				AllowNonExistField: true,
				GlobalFilters: &cond.CondCfg{
					Name:        "y",
					Operation:   cond.OperationLike,
					ValueOptCfg: vopt.ValueOptCfg{Value: "ai", ValueFrom: "const"},
				},
			}

			expectedViewResponse := &interfaces.ViewUniResponseV2{
				View:       expectedRes[0],
				TotalCount: nil,
				Entries:    []map[string]any{},
			}

			res, httpErr := dvsMock.GetSingleViewData(testCtx, view.ViewID, query1)
			So(httpErr, ShouldBeNil)
			So(res, ShouldResemble, expectedViewResponse)
		})

		Convey("GetSingleViewData failed, get indices error", func() {
			expectedErr := errors.New("some error")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_InternalError_GetIndicesFailed).WithErrorDetails(expectedErr.Error())
			psMock.EXPECT().CheckPermissionWithResult(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			dvaMock.EXPECT().GetDataViewsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedViewRes, nil)
			ibaMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(indicesResult, 200, expectedErr)

			patch1 := ApplyFunc(buildDSL,
				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView, viewIndicesMap map[string][]string) (interfaces.DSLCfg, error) {
					return interfaces.DSLCfg{}, expectedHttpErr
				},
			)
			defer patch1.Reset()

			_, httpErr := dvsMock.GetSingleViewData(testCtx, view.ViewID, query)
			So(httpErr, ShouldResemble, expectedHttpErr)
		})

		Convey("GetSingleViewData failed, convert to dsl error", func() {
			expectedErr := errors.New("some error")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest,
				uerrors.Uniquery_DataView_InvalidParameter_Filters).WithErrorDetails(expectedErr.Error())
			psMock.EXPECT().CheckPermissionWithResult(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			dvaMock.EXPECT().GetDataViewsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedViewRes, nil)
			ibaMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(indicesResult, 200, nil)

			patch1 := ApplyFunc(buildDSL,
				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView, viewIndicesMap map[string][]string) (interfaces.DSLCfg, error) {
					return interfaces.DSLCfg{}, expectedHttpErr
				},
			)
			defer patch1.Reset()

			_, httpErr := dvsMock.GetSingleViewData(testCtx, view.ViewID, query)
			So(httpErr, ShouldResemble, expectedHttpErr)
		})

		Convey("GetSingleViewData failed, search submit error", func() {
			expectedErr := errors.New("some error")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				uerrors.Uniquery_InternalError_SearchSubmitFailed).WithErrorDetails(expectedErr.Error())
			psMock.EXPECT().CheckPermissionWithResult(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			dvaMock.EXPECT().GetDataViewsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedViewRes, nil)
			ibaMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(indicesResult, 200, nil)
			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).Return([]byte{}, 0, expectedErr)
			osaMock.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(totalBytes, 200, nil).AnyTimes()

			patch1 := ApplyFunc(buildDSL,
				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView, viewIndicesMap map[string][]string) (interfaces.DSLCfg, error) {
					return interfaces.DSLCfg{}, nil
				},
			)
			defer patch1.Reset()

			_, httpErr := dvsMock.GetSingleViewData(testCtx, view.ViewID, query)
			So(httpErr, ShouldResemble, expectedHttpErr)
		})

		Convey("GetSingleViewData failed,convert to view uniResponse error", func() {
			expectedErr := errors.New("some error")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_InternalError_ConvertToViewUniResponseFailed).WithErrorDetails(expectedErr.Error())

			psMock.EXPECT().CheckPermissionWithResult(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			dvaMock.EXPECT().GetDataViewsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedViewRes, nil)
			ibaMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(indicesResult, 200, nil)
			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).Return([]byte{}, 0, nil)
			osaMock.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(totalBytes, 200, nil).AnyTimes()

			patch1 := ApplyFunc(buildDSL,
				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView, viewIndicesMap map[string][]string) (interfaces.DSLCfg, error) {
					return interfaces.DSLCfg{}, nil
				},
			)
			defer patch1.Reset()

			patch3 := ApplyFunc(convertToViewUniResponse,
				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView, content []byte) (*interfaces.ViewUniResponseV2, error) {
					return nil, expectedHttpErr
				},
			)
			defer patch3.Reset()

			_, httpErr := dvsMock.GetSingleViewData(testCtx, view.ViewID, query)
			So(httpErr, ShouldResemble, expectedHttpErr)
		})

		Convey("GetSingleViewData success", func() {
			psMock.EXPECT().CheckPermissionWithResult(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			dvaMock.EXPECT().GetDataViewsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedViewRes, nil)
			ibaMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(indicesResult, 200, nil)
			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).Return([]byte{}, 0, nil).AnyTimes()
			osaMock.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(totalBytes, 200, nil).AnyTimes()

			patch1 := ApplyFunc(buildDSL,
				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView, viewIndicesMap map[string][]string) (interfaces.DSLCfg, error) {
					return interfaces.DSLCfg{}, nil
				},
			)
			defer patch1.Reset()

			patch3 := ApplyFunc(convertToViewUniResponse,
				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView, content []byte) (*interfaces.ViewUniResponseV2, error) {
					return &interfaces.ViewUniResponseV2{}, nil
				},
			)
			defer patch3.Reset()

			result, httpErr := dvsMock.GetSingleViewData(testCtx, view.ViewID, query)
			So(result, ShouldNotBeNil)
			So(httpErr, ShouldBeNil)
		})

	})
}

// func TestRetrieveSingleData(t *testing.T) {
// 	Convey("Test RetrieveSingleViewData", t, func() {
// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		dvaMock := mock.NewMockDataViewAccess(mockCtrl)
// 		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
// 		osaMock := mock.NewMockOpenSearchAccess(mockCtrl)
// 		psMock := mock.NewMockPermissionService(mockCtrl)
// 		appSetting := &common.AppSetting{}

// 		dvsMock := MockNewDataViewService(appSetting, dvaMock, ibaMock, osaMock, psMock)

// 		indicesResult := map[string]map[string]interfaces.Indice{
// 			"indices": {
// 				"x": {
// 					IndexName: "x",
// 					ShardNum:  3,
// 				},
// 				"y": {
// 					IndexName: "y",
// 					ShardNum:  3,
// 				},
// 			},
// 		}
// 		var dslBuf bytes.Buffer
// 		dslStr := `
// 			{
// 				"from": 0,
// 				"size": 500,
// 				"sort": [
// 					{
// 					"@timestamp": "asc"
// 					}
// 				],
// 				"query": {
// 					"bool": {
// 					"filter": [
// 						{
// 						"range": {
// 							"@timestamp": {}
// 						}
// 						}
// 					],
// 					"must": []
// 					}
// 				}
// 			}
// 			`
// 		dslBuf.WriteString(dslStr)

// 		totalResult := `
// 		{
// 			"count": 1042,
// 			"_shards": {
// 				"total": 1,
// 				"successful": 1,
// 				"skipped": 0,
// 				"failed": 0
// 			}
// 		}
// 		`
// 		totalBytes := []byte(totalResult)

// 		view := interfaces.DataView{
// 			ViewID:     "1",
// 			FieldScope: 1,
// 			Fields: []*cond.ViewField{
// 				{Name: "@timestamp", Type: dtype.DATATYPE_DATETIME},
// 				{Name: "__id", Type: dtype.DATATYPE_KEYWORD},
// 				{Name: "__write_time", Type: dtype.DATATYPE_DATETIME},
// 			},
// 		}

// 		query := &interfaces.DataViewQueryV1{}

// 		Convey("RetrieveSingleViewData failed, get data view by ID error", func() {
// 			expectedErr := errors.New("some error")
// 			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
// 				uerrors.Uniquery_DataView_InternalError_GetDataViewByIDFailed).WithErrorDetails(expectedErr.Error())

// 			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dvaMock.EXPECT().GetDataViewsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedErr)

// 			_, httpErr := dvsMock.RetrieveSingleViewData(testCtx, view.ViewID, query)
// 			So(httpErr, ShouldResemble, expectedHttpErr)
// 		})

// 		Convey("RetrieveSingleViewData failed, query single view data-to dsl error", func() {
// 			expectedView := []*interfaces.DataView{
// 				{
// 					ViewID: "1",
// 					// DataSource: map[string]any{
// 					// 	"type": "index_base",
// 					// 	"index_base": []any{
// 					// 		interfaces.SimpleIndexBase{
// 					// 			BaseType: "x",
// 					// 		},
// 					// 	},
// 					// },
// 					// Condition: &cond.CondCfg{Operation: "unsupport"},
// 					Fields: []*cond.ViewField{
// 						{Name: "@timestamp", Type: dtype.DATATYPE_DATETIME},
// 						{Name: "__id", Type: dtype.DATATYPE_KEYWORD},
// 						{Name: "__write_time", Type: dtype.DATATYPE_DATETIME},
// 					},
// 				},
// 			}
// 			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dvaMock.EXPECT().GetDataViewsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedView, nil)
// 			ibaMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(indicesResult, 200, nil)

// 			_, httpErr := dvsMock.RetrieveSingleViewData(testCtx, view.ViewID, query)
// 			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusBadRequest)
// 			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_DataView_InvalidParameter_Filters)
// 		})

// 		Convey("RetrieveSingleViewData failed, get view res internal error", func() {
// 			expectedView := []*interfaces.DataView{
// 				{
// 					ViewID: "1",
// 					// DataSource: map[string]any{
// 					// 	"type": "index_base",
// 					// 	"index_base": []any{
// 					// 		interfaces.SimpleIndexBase{
// 					// 			BaseType: "x",
// 					// 		},
// 					// 	},
// 					// },
// 					// FieldScope: 1,
// 					// Condition:  &cond.CondCfg{Operation: cond.OperationNotIn, Name: "aa"},
// 					Fields: []*cond.ViewField{{Name: "aa"},
// 						{Name: "@timestamp"},
// 						{Name: "__id"},
// 						{Name: "__write_time", Type: dtype.DATATYPE_DATETIME},
// 					},
// 				},
// 			}
// 			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dvaMock.EXPECT().GetDataViewsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedView, nil)
// 			ibaMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(indicesResult, 200, nil)

// 			_, httpErr := dvsMock.RetrieveSingleViewData(testCtx, view.ViewID, query)
// 			So(httpErr, ShouldNotBeNil)
// 		})

// 		Convey("RetrieveSingleViewData success,", func() {
// 			expectedView := []*interfaces.DataView{
// 				{
// 					ViewID: "1",
// 					// DataSource: map[string]any{
// 					// 	"type": "index_base",
// 					// 	"index_base": []any{
// 					// 		interfaces.SimpleIndexBase{
// 					// 			BaseType: "x",
// 					// 		},
// 					// 	},
// 					// },
// 					FieldScope: 1,
// 					// Condition:  &cond.CondCfg{Operation: cond.OperationExist, Name: "aa"},
// 					Fields: []*cond.ViewField{
// 						{Name: "aa"},
// 						{Name: "@timestamp"},
// 						{Name: "__id"},
// 						{Name: "__write_time", Type: dtype.DATATYPE_DATETIME},
// 					},
// 				},
// 			}
// 			psMock.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dvaMock.EXPECT().GetDataViewsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedView, nil)
// 			ibaMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(indicesResult, 200, nil)
// 			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
// 				gomock.Any(), gomock.Any()).Return(osResult, 0, nil).AnyTimes()
// 			osaMock.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(totalBytes, 200, nil).AnyTimes()

// 			patch1 := ApplyFunc(buildDSL,
// 				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView) (bytes.Buffer, error) {
// 					return dslBuf, nil
// 				},
// 			)
// 			defer patch1.Reset()

// 			patch3 := ApplyFunc(convertToViewUniResponse,
// 				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView, content []byte) (*interfaces.ViewUniResponseV2, error) {
// 					return &interfaces.ViewUniResponseV2{}, nil
// 				},
// 			)
// 			defer patch3.Reset()

// 			result, httpErr := dvsMock.RetrieveSingleViewData(testCtx, view.ViewID, query)
// 			So(result, ShouldNotBeNil)
// 			So(httpErr, ShouldBeNil)
// 		})
// 	})
// }

func TestQuerySingleViewData(t *testing.T) {
	Convey("Test querySingleViewData", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		dvaMock := mock.NewMockDataViewAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		osaMock := mock.NewMockOpenSearchAccess(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		appSetting := &common.AppSetting{}

		dvsMock := MockNewDataViewService(appSetting, dvaMock, ibaMock, osaMock, psMock)

		view := &interfaces.DataView{
			QueryType: interfaces.QueryType_IndexBase,
		}
		indicesResult := map[string]map[string]interfaces.Indice{
			"indices": {
				"x": {
					IndexName: "x",
					ShardNum:  3,
				},
				"y": {
					IndexName: "y",
					ShardNum:  3,
				},
			},
		}
		var dslBuf bytes.Buffer
		dslStr := `
			{
				"from": 0,
				"size": 500,
				"sort": [
					{
					"@timestamp": "asc"
					}
				],
				"query": {
					"bool": {
					"filter": [
						{
						"range": {
							"@timestamp": {}
						}
						}
					],
					"must": []
					}
				}
			}
			`
		dslBuf.WriteString(dslStr)

		totalResult := `
		{
			"count": 1042,
			"_shards": {
				"total": 1,
				"successful": 1,
				"skipped": 0,
				"failed": 0
			}
		}
		`
		totalBytes := []byte(totalResult)

		Convey("querySingleViewData failed, scroll query", func() {
			query := &interfaces.DataViewQueryV1{
				Scroll:   "1m",
				ScrollId: "abcdefg",
			}

			osaMock.EXPECT().Scroll(gomock.Any(), gomock.Any()).Return([]byte{}, 0, nil)

			_, _, err := dvsMock.querySingleViewData(testCtx, query, view)
			So(err, ShouldBeNil)
		})

		Convey("querySingleViewData failed, scroll query failed", func() {
			query := &interfaces.DataViewQueryV1{
				Scroll:   "1m",
				ScrollId: "abcdefg",
			}

			expectedErr := errors.New("some error")
			osaMock.EXPECT().Scroll(gomock.Any(), gomock.Any()).Return([]byte{}, 0, expectedErr)

			_, _, err := dvsMock.querySingleViewData(testCtx, query, view)
			So(err, ShouldNotBeNil)
		})

		Convey("querySingleViewData failed, mapstructure decode failed", func() {
			patch1 := ApplyFunc(buildDSL,
				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView, viewIndicesMap map[string][]string) (interfaces.DSLCfg, error) {
					return interfaces.DSLCfg{}, nil
				},
			)
			defer patch1.Reset()

			view := &interfaces.DataView{
				// DataSource: map[string]any{
				// 	interfaces.INDEX_BASE: "wrong format",
				// },
			}
			query := &interfaces.DataViewQueryV1{}

			res, _, err := dvsMock.querySingleViewData(testCtx, query, view)
			So(len(res), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("querySingleViewData failed, scroll string convert to time duration failed", func() {
			view := &interfaces.DataView{
				// ViewID: "1", DataSource: map[string]any{
				// 	"type": "index_base",
				// 	"index_base": []any{
				// 		interfaces.SimpleIndexBase{
				// 			BaseType: "x",
				// 		},
				// 	},
				// },
			}
			ibaMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(indicesResult, 200, nil)
			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).Return([]byte{}, 0, nil).AnyTimes()

			patch1 := ApplyFunc(buildDSL,
				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView, viewIndicesMap map[string][]string) (interfaces.DSLCfg, error) {
					return interfaces.DSLCfg{}, nil
				},
			)
			defer patch1.Reset()

			patch3 := ApplyFunc(convertToViewUniResponse,
				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView, content []byte) (*interfaces.ViewUniResponseV2, error) {
					return &interfaces.ViewUniResponseV2{}, nil
				},
			)
			defer patch3.Reset()

			query := &interfaces.DataViewQueryV1{
				Scroll: "2mmmm",
			}

			res, _, httpErr := dvsMock.querySingleViewData(testCtx, query, view)
			So(len(res), ShouldEqual, 0)
			So(httpErr, ShouldNotBeNil)
		})

		Convey("querySingleViewData success", func() {
			view := &interfaces.DataView{
				ViewID:    "1",
				QueryType: interfaces.QueryType_IndexBase,
				Type:      interfaces.ViewType_Atomic,
				// DataSource: map[string]any{
				// 	"type": "index_base",
				// 	"index_base": []any{
				// 		interfaces.SimpleIndexBase{
				// 			BaseType: "x",
				// 		},
				// 	},
				// },
			}
			ibaMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(indicesResult, 200, nil)
			osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any()).Return([]byte{}, 0, nil).AnyTimes()
			osaMock.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(totalBytes, 200, nil).AnyTimes()

			patch1 := ApplyFunc(buildDSL,
				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView, viewIndicesMap map[string][]string) (interfaces.DSLCfg, error) {
					return interfaces.DSLCfg{}, nil
				},
			)
			defer patch1.Reset()

			patch3 := ApplyFunc(convertToViewUniResponse,
				func(ctx context.Context, query interfaces.ViewQueryInterface, view *interfaces.DataView, content []byte) (*interfaces.ViewUniResponseV2, error) {
					return &interfaces.ViewUniResponseV2{}, nil
				},
			)
			defer patch3.Reset()

			query := &interfaces.DataViewQueryV1{
				Scroll: "2m",
				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
					NeedTotal: true,
				},
			}

			res, _, httpErr := dvsMock.querySingleViewData(testCtx, query, view)
			So(len(res), ShouldEqual, 0)
			So(httpErr, ShouldBeNil)
		})
	})
}

// func TestbuildDSL(t *testing.T) {
// 	Convey("Test buildDSL", t, func() {
//

// 		Convey("buildDSL failed, unsupported filter operation type", func() {
// 			view := &interfaces.DataView{
// 				Condition: &cond.CondCfg{
// 					Name:      "a",
// 					Operation: "*",
// 					ValueOptCfg: vopt.ValueOptCfg{
// 						Value:     "10",
// 						ValueFrom: "const",
// 					},
// 				},
// 			}

// 			query := &interfaces.DataViewQueryV1{
// 				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
// 					Start:  1693408963000,
// 					End:    1693408964000,
// 					Offset: 0,
// 					Limit:  10,
// 				},
// 			}

// 			res, err := buildDSL(testCtx,query, view)
// 			So(res, ShouldResemble, bytes.Buffer{})
// 			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, uerrors.Uniquery_DataView_InvalidParameter_Filters)
// 		})

// 		Convey("buildDSL success", func() {
// 			view := &interfaces.DataView{
// 				Condition: &cond.CondCfg{
// 					Name:      "a",
// 					Operation: cond.OperationEq,
// 					ValueOptCfg: vopt.ValueOptCfg{
// 						Value:     6,
// 						ValueFrom: "const",
// 					},
// 				},
// 				FieldsMap: map[string]*cond.ViewField{
// 					"a":          {Name: "a", Type: dtype.DATATYPE_LONG},
// 					"b":          {Name: "b", Type: dtype.DATATYPE_TEXT},
// 					"@timestamp": {Name: "@timestamp", Type: dtype.DATATYPE_DATETIME},
// 				},
// 			}

// 			query := &interfaces.DataViewQueryV1{
// 				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
// 					Start:  1693408963000,
// 					End:    1693408964000,
// 					Offset: 0,
// 					Limit:  10,
// 				},
// 				SortParamsV1: interfaces.SortParamsV1{
// 					Sort:      "@timestamp",
// 					Direction: "desc",
// 				},
// 			}

// 			queryStr := `
// 			{
// 				"from": 0,
// 				"size": 10,
// 				"sort": [
// 				  { "@timestamp": {"order": "desc"} }
// 				],
// 				"query": {
// 				  "bool": {
// 					"filter": [
// 					  {
// 						"term": {
// 							"a": {
// 								"value": 6
// 							}
// 						}
// 					  },
// 					  {
// 						"range": {
// 						  "@timestamp": {
// 							"gte": 1693408963000,
// 							"lte": 1693408964000
// 						  }
// 						}
// 					  }
// 					],
// 					"must":
// 					  	[]
// 				  }
// 				}
// 			  }
// 			`
// 			var expectedQueryBuffer bytes.Buffer
// 			expectedQueryBuffer.WriteString(queryStr)

// 			dsl, err := buildDSL(testCtx,query, view)

// 			So(dsl, ShouldResemble, expectedQueryBuffer)
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("buildDSL success, global condition and view condition all exist", func() {
// 			view := &interfaces.DataView{
// 				Condition: &cond.CondCfg{
// 					Name:      "b",
// 					Operation: cond.OperationNotExist,
// 				},
// 				FieldsMap: map[string]*cond.ViewField{
// 					"a": {Name: "a", Type: dtype.DATATYPE_LONG},
// 					"b": {Name: "b", Type: dtype.DATATYPE_TEXT},
// 					"c": {Name: "c", Type: dtype.DATATYPE_TEXT},
// 				},
// 			}

// 			query := &interfaces.DataViewQueryV1{
// 				GlobalFilters: &cond.CondCfg{
// 					Name:      "a",
// 					Operation: cond.OperationEq,
// 					ValueOptCfg: vopt.ValueOptCfg{
// 						Value:     6,
// 						ValueFrom: "const",
// 					},
// 				},
// 				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
// 					Start:  1693408963000,
// 					End:    1693408964000,
// 					Offset: 0,
// 					Limit:  10,
// 				},
// 				SortParamsV1: interfaces.SortParamsV1{
// 					Sort:      "c",
// 					Direction: "desc",
// 				},
// 			}

// 			queryStr := `
// 			{
// 				"from": 0,
// 				"size": 10,
// 				"sort": [
// 					{
// 						"c.keyword": {
// 							"order": "desc"
// 						}
// 					}
// 				],
// 				"query": {
// 					"bool": {
// 						"filter": [
// 							{
// 								"bool": {
// 									"filter": [
// 										{
// 											"term": {
// 												"a": {
// 													"value": 6
// 												}
// 											}
// 										},
// 										{
// 											"bool": {
// 												"must_not": [
// 													{
// 														"exists": {
// 															"field": "b"
// 														}
// 													}
// 												]
// 											}
// 										}
// 									]
// 								}
// 							},
// 							{
// 								"range": {
// 									"@timestamp": {
// 										"gte": 1693408963000,
// 										"lte": 1693408964000
// 									}
// 								}
// 							}
// 						],
// 						"must": []
// 					}
// 				}
// 			}
// 			`

// 			var expectedQueryBuffer bytes.Buffer
// 			expectedQueryBuffer.WriteString(queryStr)

// 			dsl, err := buildDSL(testCtx,query, view)

// 			So(dsl, ShouldResemble, expectedQueryBuffer)
// 			So(err, ShouldBeNil)
// 		})
// 	})
// }

func TestGetViewResInternal(t *testing.T) {
	total := int64(8888)
	Convey("Test GetViewResInternal", t, func() {
		content := []byte(`{
			"took" : 14,
			"timed_out" : false,
			"_shards" : {
				"total" : 3,
				"successful" : 3,
				"skipped" : 0,
				"failed" : 0
			},
			"hits" : {
				"total" : {
					"value" : 3,
					"relation" : "eq"
				},
				"max_score" : 1.0,
				"hits" : [
					{
						"_index" : "ar_audit_log_en_us-2023.07-0",
						"_type" : "_doc",
						"_id" : "m2nIHokBDG8c24cLzoZK",
						"_score" : 1.0,
						"_source" : {
							"日志描述" : "Login Failure. System Error",
							"失败原因" : "System Error",
							"IP地址" : "localhost",
							"级别" : "Alert",
							"arAgentType" : "audit",
							"host" : "localhost",
							"签名算法" : "RSA",
							"@timestamp" : "2023-07-04T02:42:11.147Z",
							"时间" : "2023-07-04 10:42:10",
							"日志类型" : "Access Log",
							"结果" : "Failure",
							"type" : "ar_audit_log_en_us",
							"动作" : "Login",
							"用户" : "admin",
							"语言" : "en_US",
							"签名" : "XwCX/67J40/+FyvZPqEgfSfulfsruuhd6xlmhDfOh3HGTP8eQV5X5eg0D/TirRv4lAUylf1qV5aZqSVuL3R24mGddBOf6acQbcRhpsAcq424Ml7azLMJ7OqP2UchDtki3G6ScKecUMfF13RArfIfzKPq7sEDIIw0EQNcUBWz+6s=",
							"@version" : "1"
						}
					},
					{
						"_index" : "ar_audit_log_en_us-2023.07-0",
						"_type" : "_doc",
						"_id" : "nGnIHokBDG8c24cLz4Zw",
						"_score" : 1.0,
						"_source" : {
							"日志描述" : "Login Failure. System Error",
							"失败原因" : "System Error",
							"IP地址" : "localhost",
							"级别" : "Alert",
							"arAgentType" : "audit",
							"host" : "localhost",
							"签名算法" : "RSA",
							"@timestamp" : "2023-07-04T02:42:42.269Z",
							"时间" : "2023-07-04 10:42:42",
							"日志类型" : "Access Log",
							"结果" : "Failure",
							"type" : "ar_audit_log_en_us",
							"动作" : "Login",
							"用户" : "admin",
							"语言" : "en_US",
							"签名" : "HdYPvvyriqgq3XiU3GQKBAdYRFhuIIaIh/qKEHe4PKHLC5PberCFSzlpZ7JLwgSDudIo7jfUcSljeuREOw2lvKF8+7Oml2w/XeGZaW0xnQn+dlPjJ4c0tNZMf94XCJyZBOv805ZD8MYcvgGdg6ZNfQ6Td5f2qOzRLP9gos/f5x8=",
							"@version" : "1"
						}
					}
				]
			}
		}`)

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		dvaMock := mock.NewMockDataViewAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		osaMock := mock.NewMockOpenSearchAccess(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		appSetting := &common.AppSetting{}

		dvsMock := MockNewDataViewService(appSetting, dvaMock, ibaMock, osaMock, psMock)

		Convey("GetViewResInternal success", func() {
			view := &interfaces.DataView{
				Type: interfaces.ViewType_Custom,
				Fields: []*cond.ViewField{
					{Name: "@timestamp"},
					{Name: "arAgentType"},
					{Name: "日志描述"},
					{Name: "签名"},
				},
				FieldsMap: map[string]*cond.ViewField{
					"@timestamp":  {Name: "@timestamp"},
					"arAgentType": {Name: "arAgentType"},
					"日志描述":        {Name: "日志描述"},
					"签名":          {Name: "签名"},
				},
			}
			query := &interfaces.DataViewQueryV1{}

			res, err := dvsMock.GetViewResInternal(testCtx, query, view, content, total)
			So(res, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
	})
}

func TestConvertToViewUniResponse(t *testing.T) {
	Convey("Test ConvertToViewUniResponse", t, func() {
		InitViewPool(common.PoolSetting{
			ViewPoolSize: 10,
		})

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		// dvaMock := mock.NewMockDataViewAccess(mockCtrl)
		// ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		// osaMock := mock.NewMockOpenSearchAccess(mockCtrl)
		// appSetting := &common.AppSetting{}

		// dvsMock := MockNewDataViewService(appSetting, dvaMock, ibaMock, osaMock)

		osResultStr := `
			{
				"took" : 14,
				"timed_out" : false,
				"_shards" : {
				  "total" : 3,
				  "successful" : 3,
				  "skipped" : 0,
				  "failed" : 0
				},
				"hits" : {
				  "total" : {
					"value" : 3,
					"relation" : "eq"
				  },
				  "max_score" : 1.0,
				  "hits" : [
					{
					  "_index" : "ar_audit_log_en_us-2023.07-0",
					  "_type" : "_doc",
					  "_id" : "m2nIHokBDG8c24cLzoZK",
					  "_score" : 1.0,
					  "_source" : {
						"日志描述" : "Login Failure. System Error",
						"失败原因" : "System Error",
						"IP地址" : "localhost",
						"级别" : "Alert",
						"arAgentType" : "audit",
						"host" : "localhost",
						"签名算法" : "RSA",
						"@timestamp" : "2023-07-04T02:42:11.147Z",
						"时间" : "2023-07-04 10:42:10",
						"日志类型" : "Access Log",
						"结果" : "Failure",
						"type" : "ar_audit_log_en_us",
						"动作" : "Login",
						"用户" : "admin",
						"语言" : "en_US",
						"签名" : "XwCX/67J40/+FyvZPqEgfSfulfsruuhd6xlmhDfOh3HGTP8eQV5X5eg0D/TirRv4lAUylf1qV5aZqSVuL3R24mGddBOf6acQbcRhpsAcq424Ml7azLMJ7OqP2UchDtki3G6ScKecUMfF13RArfIfzKPq7sEDIIw0EQNcUBWz+6s=",
						"@version" : "1"
					  }
					},
					{
					  "_index" : "ar_audit_log_en_us-2023.07-0",
					  "_type" : "_doc",
					  "_id" : "nGnIHokBDG8c24cLz4Zw",
					  "_score" : 1.0,
					  "_source" : {
						"日志描述" : "Login Failure. System Error",
						"失败原因" : "System Error",
						"IP地址" : "localhost",
						"级别" : "Alert",
						"arAgentType" : "audit",
						"host" : "localhost",
						"签名算法" : "RSA",
						"@timestamp" : "2023-07-04T02:42:42.269Z",
						"时间" : "2023-07-04 10:42:42",
						"日志类型" : "Access Log",
						"结果" : "Failure",
						"type" : "ar_audit_log_en_us",
						"动作" : "Login",
						"用户" : "admin",
						"语言" : "en_US",
						"签名" : "HdYPvvyriqgq3XiU3GQKBAdYRFhuIIaIh/qKEHe4PKHLC5PberCFSzlpZ7JLwgSDudIo7jfUcSljeuREOw2lvKF8+7Oml2w/XeGZaW0xnQn+dlPjJ4c0tNZMf94XCJyZBOv805ZD8MYcvgGdg6ZNfQ6Td5f2qOzRLP9gos/f5x8=",
						"@version" : "1"
					  }
					},
					{
					  "_index" : "ar_audit_log_en_us-2023.07-0",
					  "_type" : "_doc",
					  "_id" : "mWnIHokBDG8c24cLzYbt",
					  "_score" : 1.0,
					  "_source" : {
						"日志描述" : "Login Failure. System Error",
						"失败原因" : "System Error",
						"IP地址" : "localhost",
						"级别" : "Alert",
						"arAgentType" : "audit",
						"host" : "localhost",
						"签名算法" : "RSA",
						"@timestamp" : "2023-07-04T02:42:18.529Z",
						"时间" : "2023-07-04 10:42:18",
						"日志类型" : "Access Log",
						"结果" : "Failure",
						"type" : "ar_audit_log_en_us",
						"动作" : "Login",
						"用户" : "admin",
						"语言" : "en_US",
						"签名" : "jO72jDmZJg1ZSLsXXU/Bg13HDDNch+bFocuW4x8rDV0O0urzfb53hcB7n9q2RW1K0ykCGlTv9TJOKKCe6drScHBQMoX4yDz6uHXtdZJrfsFhqN24DuDWSVdUCYwSeAIEWJEWTWpcmFvQNkLsOeCVny5a10Wu8J5zrwDxe+YGUQU=",
						"@version" : "1"
					  }
					}
				  ]
				}
			  }

			`
		total := int64(8888)
		content := []byte(osResultStr)

		view := &interfaces.DataView{
			Fields: []*cond.ViewField{
				{Name: "@timestamp"},
				{Name: "arAgentType"},
				{Name: "日志描述"},
				{Name: "签名"},
			},
			FieldsMap: map[string]*cond.ViewField{
				"@timestamp":  {Name: "@timestamp"},
				"arAgentType": {Name: "arAgentType"},
				"日志描述":        {Name: "日志描述"},
				"签名":          {Name: "签名"},
			},
			QueryType: interfaces.QueryType_IndexBase,
			Type:      interfaces.ViewType_Atomic,
		}
		query := &interfaces.DataViewQueryV1{
			ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
				Format: interfaces.Format_Flat,
			},
		}

		Convey("Convert failed, get hits failed", func() {
			patch := ApplyFunc(sonic.Get, func(src []byte, path ...interface{}) (ast.Node, error) {
				return ast.Node{}, errors.New("get hits failed")
			})
			defer patch.Reset()

			_, err := convertToViewUniResponse(testCtx, query, view, content, total)
			So(err, ShouldNotBeNil)
		})

		Convey("Convert success1", func() {
			// expectedRes := &interfaces.ViewUniResponseV2{
			// 	View:       view,
			// 	TotalCount: &total,
			// 	Entries: []map[string]any{
			// 		{
			// 			"@timestamp":  "2023-07-04T02:42:11.147Z",
			// 			"__index":     "ar_audit_log_en_us-2023.07-0",
			// 			"__id":        "m2nIHokBDG8c24cLzoZK",
			// 			"arAgentType": "audit",
			// 			"日志描述":        "Login Failure. System Error",
			// 			"签名":          "XwCX/67J40/+FyvZPqEgfSfulfsruuhd6xlmhDfOh3HGTP8eQV5X5eg0D/TirRv4lAUylf1qV5aZqSVuL3R24mGddBOf6acQbcRhpsAcq424Ml7azLMJ7OqP2UchDtki3G6ScKecUMfF13RArfIfzKPq7sEDIIw0EQNcUBWz+6s=",
			// 		},
			// 		{
			// 			"@timestamp":  "2023-07-04T02:42:42.269Z",
			// 			"arAgentType": "audit",
			// 			"__index":     "ar_audit_log_en_us-2023.07-0",
			// 			"__id":        "nGnIHokBDG8c24cLz4Zw",
			// 			"日志描述":        "Login Failure. System Error",
			// 			"签名":          "HdYPvvyriqgq3XiU3GQKBAdYRFhuIIaIh/qKEHe4PKHLC5PberCFSzlpZ7JLwgSDudIo7jfUcSljeuREOw2lvKF8+7Oml2w/XeGZaW0xnQn+dlPjJ4c0tNZMf94XCJyZBOv805ZD8MYcvgGdg6ZNfQ6Td5f2qOzRLP9gos/f5x8=",
			// 		},
			// 		{
			// 			"@timestamp":  "2023-07-04T02:42:18.529Z",
			// 			"__index":     "ar_audit_log_en_us-2023.07-0",
			// 			"arAgentType": "audit",
			// 			"__id":        "mWnIHokBDG8c24cLzYbt",
			// 			"日志描述":        "Login Failure. System Error",
			// 			"签名":          "jO72jDmZJg1ZSLsXXU/Bg13HDDNch+bFocuW4x8rDV0O0urzfb53hcB7n9q2RW1K0ykCGlTv9TJOKKCe6drScHBQMoX4yDz6uHXtdZJrfsFhqN24DuDWSVdUCYwSeAIEWJEWTWpcmFvQNkLsOeCVny5a10Wu8J5zrwDxe+YGUQU=",
			// 		},
			// 	},
			// }
			_, err := convertToViewUniResponse(testCtx, query, view, content, total)
			// So(res, ShouldResemble, expectedRes)
			So(err, ShouldBeNil)
		})

		Convey("Convert success2, format is original", func() {
			view := &interfaces.DataView{
				QueryType: interfaces.QueryType_IndexBase,
				Type:      interfaces.ViewType_Atomic,
			}
			query := &interfaces.DataViewQueryV1{
				ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
					Format: interfaces.Format_Original,
				},
			}

			osResultStr := `
			{
				"took": 14,
				"timed_out": false,
				"_shards": {
					"total": 3,
					"successful": 3,
					"skipped": 0,
					"failed": 0
				},
				"hits": {
					"total": {
						"value": 3,
						"relation": "eq"
					},
					"max_score": 1.0,
					"hits": [
						{
							"_index": "mdl-metricbeat-2023.11.09-000000",
							"_type": "_doc",
							"_id": "bbRMs4sBwsj5g1XWFHA1",
							"_score": null,
							"_routing": "944177",
							"_source": {
								"@version": "1",
								"__data_type": "metricbeat",
								"arPort": 20010,
								"metrics": {
									"node_network_info": 1
								},
								"category": "metric",
								"type": "metricbeat",
								"__write_time": "2023-11-09T08:56:29.312Z",
								"arAgentType": "metricbeat",
								"metricset": {
									"name": "collector",
									"period": 30000
								},
								"service": {
									"address": "http://localhost:9102/metrics",
									"type": "prometheus"
								},
								"labels": {
									"operstate": "up",
									"device": "cali1a992d74725",
									"address": "ee:ee:ee:ee:ee:ee",
									"broadcast": "ff:ff:ff:ff:ff:ff",
									"duplex": "full",
									"instance": "localhost:9102",
									"job": "prometheus"
								},
								"__index_base": "metricbeat",
								"__routing": "944177",
								"host": {
									"name": "metricbeat-metricbeat-6dfb85478-964fg"
								},
								"@timestamp": "2023-11-09T08:56:27.023Z",
								"__labels_str": "address=ee:ee:ee:ee:ee:ee,broadcast=ff:ff:ff:ff:ff:ff,device=cali1a992d74725,duplex=full,instance=localhost:9102,job=prometheus,operstate=up",
								"tags": [
									"beats_input_raw_event"
								]
							},
							"sort": [
								1699520187023
							]
						},
						{
							"_index": "mdl-metricbeat-2023.11.09-000000",
							"_type": "_doc",
							"_id": "brRMs4sBwsj5g1XWFHA1",
							"_score": null,
							"_routing": "944177",
							"_source": {
								"@version": "1",
								"__data_type": "metricbeat",
								"arPort": 20010,
								"metrics": {
									"node_filesystem_device_error": 1
								},
								"category": "metric",
								"type": "metricbeat",
								"__write_time": "2023-11-09T08:56:29.312Z",
								"arAgentType": "metricbeat",
								"metricset": {
									"name": "collector",
									"period": 30000
								},
								"service": {
									"address": "http://localhost:9102/metrics",
									"type": "prometheus"
								},
								"labels": {
									"fstype": "tmpfs",
									"device": "shm",
									"mountpoint": "/mnt/proton_data/cs_docker_data/containers/f5b518fe997fae0790898e1ade16068346133f4209c1a82d75e878891704fe03/mounts/shm",
									"instance": "localhost:9102",
									"job": "prometheus"
								},
								"__index_base": "metricbeat",
								"__routing": "944177",
								"host": {
									"name": "metricbeat-metricbeat-6dfb85478-964fg"
								},
								"@timestamp": "2023-11-09T08:56:27.023Z",
								"__labels_str": "device=shm,fstype=tmpfs,instance=localhost:9102,job=prometheus,mountpoint=/mnt/proton_data/cs_docker_data/containers/f5b518fe997fae0790898e1ade16068346133f4209c1a82d75e878891704fe03/mounts/shm",
								"tags": [
									"beats_input_raw_event"
								]
							},
							"sort": [
								1699520187023
							]
						}
					]
				}
			}
			`
			content := []byte(osResultStr)

			_, err := convertToViewUniResponse(testCtx, query, view, content, total)
			// So(res.Datas, ShouldNotBeNil)
			// So(len(res.Datas[0].Values[0]), ShouldEqual, 19)
			So(err, ShouldBeNil)
		})
	})
}

func TestMergeIndexBaseFields(t *testing.T) {
	Convey("Test mergeIndexBaseFields", t, func() {
		Convey("Convert succeed", func() {
			mappings := interfaces.Mappings{
				DynamicMappings: []interfaces.IndexBaseField{

					{
						Field: "a",
						Type:  "text",
					},
					{
						Field: "b.ip",
						Type:  "ip",
					},
					{
						Field: "b.latitude",
						Type:  "half_float",
					},
					{
						Field: "c",
						Type:  "long",
					},
				},
				MetaMappings: []interfaces.IndexBaseField{
					{
						Field: "__data_type",
						Type:  "keyword",
					},
				},
			}

			fields := mergeIndexBaseFields(mappings)

			expectedFields := []interfaces.IndexBaseField{
				{
					Field: "__data_type",
					Type:  "keyword",
				},
				{
					Field: "a",
					Type:  "text",
				},
				{
					Field: "b.ip",
					Type:  "ip",
				},
				{
					Field: "b.latitude",
					Type:  "half_float",
				},
				{
					Field: "c",
					Type:  "long",
				},
			}

			So(fields, ShouldResemble, expectedFields)
		})
	})
}

func TestFlatten(t *testing.T) {
	Convey("Test Flatten", t, func() {
		Convey("success, value is {}", func() {
			top := true
			prefix := ""
			srcStr := `{"a": {"b": 1, "c": 2, "d": {}}}`
			var src map[string]any
			err := sonic.Unmarshal([]byte(srcStr), &src)
			if err != nil {
				t.Fatal(err)
			}
			dest := make(map[string]any)
			fieldsMap := map[string]*cond.ViewField{
				"a.b": {
					Name: "a.b",
					Type: "integer",
				},
				"a.c": {
					Name: "a.c",
					Type: "integer",
				},
				"a.d": {
					Name: "a.d",
					Type: "integer",
				},
			}

			var expectRes map[string]any
			expectResStr := `{"a.b": 1, "a.c": 2, "a.d": {}}`
			err = sonic.Unmarshal([]byte(expectResStr), &expectRes)
			if err != nil {
				t.Fatal(err)
			}

			err = flattenWithPickField(top, prefix, src, dest, fieldsMap)
			So(err, ShouldBeNil)

			So(dest, ShouldResemble, expectRes)
		})

		Convey("success, value is []", func() {
			top := true
			prefix := ""
			srcStr := `{"a": {"b": 1, "c": 2, "d": []}}`
			var src map[string]any
			err := sonic.Unmarshal([]byte(srcStr), &src)
			if err != nil {
				t.Fatal(err)
			}
			dest := make(map[string]any)
			fieldsMap := map[string]*cond.ViewField{
				"a.b": {
					Name: "a.b",
					Type: "integer",
				},
				"a.c": {
					Name: "a.c",
					Type: "integer",
				},
				"a.d": {
					Name: "a.d",
					Type: "integer",
				},
			}

			var expectRes map[string]any
			expectResStr := `{"a.b": 1, "a.c": 2, "a.d": []}`
			err = sonic.Unmarshal([]byte(expectResStr), &expectRes)
			if err != nil {
				t.Fatal(err)
			}

			err = flattenWithPickField(top, prefix, src, dest, fieldsMap)
			So(err, ShouldBeNil)

			So(dest, ShouldResemble, expectRes)
		})

		Convey("success, value is [object1, object2]", func() {
			top := true
			prefix := ""
			srcStr := `{"a": {"b": 1, "c": 2, "d": [{"e": 3}, {"f": 4, "e": 5}]}}`
			var src map[string]any
			err := sonic.Unmarshal([]byte(srcStr), &src)
			if err != nil {
				t.Fatal(err)
			}
			dest := make(map[string]any)
			fieldsMap := map[string]*cond.ViewField{
				"a.b": {
					Name: "a.b",
					Type: "integer",
				},
				"a.c": {
					Name: "a.c",
					Type: "integer",
				},
				"a.d.e": {
					Name: "a.d",
					Type: "integer",
				},
				"a.d.f": {
					Name: "a.d",
					Type: "integer",
				},
			}

			var expectRes map[string]any
			expectResStr := `{"a.b": 1, "a.c": 2, "a.d.e": [3, 5], "a.d.f": 4}`
			err = sonic.Unmarshal([]byte(expectResStr), &expectRes)
			if err != nil {
				t.Fatal(err)
			}

			err = flattenWithPickField(top, prefix, src, dest, fieldsMap)
			So(err, ShouldBeNil)

			So(dest, ShouldResemble, expectRes)
		})
	})
}

func TestJoinKey(t *testing.T) {
	Convey("Test JoinKey", t, func() {
		Convey("subkey is ''", func() {
			top := true
			prefix := "a"
			subkey := ""
			want := joinKey(top, prefix, subkey)
			So(want, ShouldEqual, "a")
		})

		Convey("top is true", func() {
			top := true
			prefix := ""
			subkey := "a"
			want := joinKey(top, prefix, subkey)
			So(want, ShouldEqual, "a")
		})

		Convey("top is false", func() {
			top := false
			prefix := "a"
			subkey := "b"
			want := joinKey(top, prefix, subkey)
			So(want, ShouldEqual, "a.b")
		})
	})
}

func TestPickData(t *testing.T) {
	Convey("Test pickData", t, func() {
		fields := map[string]*cond.ViewField{
			"name": {
				Name: "name",
				Type: "keyword",
				Path: []string{"name"},
			},
			"age": {
				Name: "age",
				Type: "keyword",
				Path: []string{"age"},
			},
		}
		origin := make(map[string]any)
		pick := make(map[string]any)

		Convey("get data error", func() {
			patches := ApplyFuncReturn(getData, nil, false, errors.New("error"))
			defer patches.Reset()

			err := pickData(origin, pick, fields)
			So(err, ShouldNotBeNil)
		})

		Convey("set data error", func() {
			patches1 := ApplyFuncReturn(getData, nil, false, nil)
			defer patches1.Reset()

			patches2 := ApplyFuncReturn(setData, errors.New("error"))
			defer patches2.Reset()

			err := pickData(origin, pick, fields)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			patches1 := ApplyFuncReturn(getData, nil, false, nil)
			defer patches1.Reset()

			patches2 := ApplyFuncReturn(setData, nil)
			defer patches2.Reset()

			err := pickData(origin, pick, fields)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetData(t *testing.T) {
	Convey("getData", t, func() {
		field := &cond.ViewField{
			Name: "age",
			Type: "keyword",
			Path: []string{"age"},
		}
		origin := make(map[string]any)
		Convey("GetDatasByPath error", func() {
			patches := ApplyFuncReturn(GetDatasByPath, nil, false, errors.New("error"))
			defer patches.Reset()

			_, _, err := getData(origin, field)
			So(err, ShouldNotBeNil)
		})

		Convey("datas is null", func() {
			patches := ApplyFuncReturn(GetDatasByPath, []any{}, false, nil)
			defer patches.Reset()

			_, _, err := getData(origin, field)
			So(err, ShouldBeNil)
		})

		Convey("success", func() {
			patches := ApplyFuncReturn(GetDatasByPath, []any{1}, false, nil)
			defer patches.Reset()

			_, _, err := getData(origin, field)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetDatasByPath(t *testing.T) {
	Convey("GetDatasByPath", t, func() {

		Convey("GetDatasByPath: simple json", func() {
			var jsonStr = `{
				"a":{
					"b":{
						"c": {
							"d": "d"
						}
					}
				}
			}`

			root := map[string]any{}
			_ = sonic.UnmarshalString(jsonStr, &root)
			path := []string{"a", "b", "c", "d"}

			dest, _, err := GetDatasByPath(root, path)
			So(err, ShouldBeNil)
			So(len(dest), ShouldEqual, 1)
		})

		Convey("GetDatasByPath: invalid json", func() {
			var jsonStr = `{
				"a":{
					"b":{
						"c": null
					}
				}
			}`

			root := map[string]any{}
			_ = sonic.UnmarshalString(jsonStr, &root)
			path := []string{"a", "b", "c", "d"}

			dest, _, err := GetDatasByPath(root, path)
			So(err, ShouldBeNil)
			So(len(dest), ShouldEqual, 0)
		})

		Convey("getNodeByPath: none value", func() {
			var jsonStr = `{
				"a":{
					"b":{
						"c": "c"
					}
				}
			}`

			root := map[string]any{}
			_ = sonic.UnmarshalString(jsonStr, &root)
			path := []string{"a", "b", "c", "d"}

			dest, _, err := GetDatasByPath(root, path)
			So(err, ShouldBeNil)
			So(len(dest), ShouldEqual, 0)
		})

		Convey("getNodeByPath: complex json", func() {
			var jsonStr = `{
				"a":[
					{
						"b":[
							{
								"c":[
									{
										"d": "d1"
									},{
										"d": ["d2"]
									}
								]
							},
							{
								"c":[
									{
										"d": ["d3", "d4"]
									}
								]
							}
						]
					}, {
						"b":[
							{
								"c":[
									{
										"d": [["d5"]]
									},{
										"d": [[], ["d6"]]
									}
								]
							},
							{
								"c":[
									{
										"d": ["d7", ["d8"]]
									},
									{
										"d": [["d9"], ["d10", "d11"]]
									}
								]
							}
						]
					}
				]
			}`

			root := map[string]any{}
			_ = sonic.UnmarshalString(jsonStr, &root)
			path := []string{"a", "b", "c", "d"}

			dest, _, err := GetDatasByPath(root, path)
			So(err, ShouldBeNil)
			So(len(dest), ShouldBeGreaterThan, 0)
		})
	})
}

func TestTask_GetLastDatas(t *testing.T) {
	Convey("Test GetLastDatas", t, func() {
		Convey("obj is nil", func() {
			var obj any = nil
			_, _, err := GetLastDatas(obj)
			So(err, ShouldBeNil)
		})

		Convey("obj is slice", func() {
			obj := []any{1, 2}
			_, isSliceValue, err := GetLastDatas(obj)
			So(isSliceValue, ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("obj is not slice", func() {
			obj := 1
			_, isSliceValue, err := GetLastDatas(obj)
			So(isSliceValue, ShouldBeFalse)
			So(err, ShouldBeNil)
		})
	})
}

func TestSetData(t *testing.T) {
	Convey("Test setData", t, func() {
		Convey("data length is 0", func() {
			field := &cond.ViewField{}
			obj := map[string]any{}
			data := []any{}
			isSliceValue := false
			err := setData(field, obj, data, isSliceValue)
			So(err, ShouldBeNil)
		})

		Convey("set a.b.c", func() {
			field := &cond.ViewField{
				Name: "a.b.c",
				Path: []string{"a", "b", "c"},
			}
			obj := map[string]any{}
			data := []any{"wahaha"}
			isSliceValue := false
			err := setData(field, obj, data, isSliceValue)
			expectedObj := []byte(`{"a": {"b": {"c": "wahaha"}}}`)
			var expected map[string]any
			if err := sonic.Unmarshal(expectedObj, &expected); err != nil {
				t.Fatal(err.Error())
			}
			So(obj, ShouldResemble, expected)
			So(err, ShouldBeNil)
		})

		Convey("set a.b.c, value is an array", func() {
			field := &cond.ViewField{
				Name: "a.b.c",
				Path: []string{"a", "b", "c"},
			}
			obj := map[string]any{}
			data := []any{"wahaha"}
			isSliceValue := true
			err := setData(field, obj, data, isSliceValue)
			expectedObj := []byte(`{"a": {"b": {"c": ["wahaha"]}}}`)
			var expected map[string]any
			if err := sonic.Unmarshal(expectedObj, &expected); err != nil {
				t.Fatal(err.Error())
			}
			So(obj, ShouldResemble, expected)
			So(err, ShouldBeNil)
		})
	})
}

// func TestCountMultiFields(t *testing.T) {
// 	Convey("Test CountMultiFields", t, func() {
// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		dvaMock := mock.NewMockDataViewAccess(mockCtrl)
// 		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
// 		osaMock := mock.NewMockOpenSearchAccess(mockCtrl)
// 		psMock := mock.NewMockPermissionService(mockCtrl)

// 		appSetting := &common.AppSetting{}

// 		dvs := MockNewDataViewService(appSetting, dvaMock, ibaMock, osaMock, psMock)

// 		dataView := &interfaces.DataView{
// 			ViewID: "1",
// 			FieldsMap: map[string]*cond.ViewField{
// 				"f1": {Name: "f1", Type: dtype.DATATYPE_KEYWORD},
// 				"f2": {Name: "f2", Type: dtype.DATATYPE_KEYWORD},
// 			},
// 		}

// 		Convey("Count succeed, but no field", func() {
// 			viewID := "1"
// 			query := &interfaces.DataViewQueryV1{}
// 			_, err := dvs.CountMultiFields(testCtx, viewID, query, []string{}, "_")
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("Count failed, caused by the error from method 'parseQuery'", func() {
// 			expectedErr := errors.New("some errors")

// 			patch := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "parseQuery",
// 				func(_ *dataViewService, ctx context.Context, viewID string) (*interfaces.DataView, error) {
// 					return nil, expectedErr
// 				})
// 			defer patch.Reset()

// 			viewID := "1"
// 			query := &interfaces.DataViewQueryV1{}
// 			_, err := dvs.CountMultiFields(testCtx, viewID, query, []string{"f1", "f2"}, "_")
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("Count succeed, but no results", func() {
// 			patch1 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "parseQuery",
// 				func(_ *dataViewService, ctx context.Context, viewID string) (*interfaces.DataView, error) {
// 					return dataView, nil
// 				})
// 			defer patch1.Reset()

// 			viewID := "1"
// 			query := &interfaces.DataViewQueryV1{}
// 			stats, err := dvs.CountMultiFields(testCtx, viewID, query, []string{"f1", "f2"}, "_")
// 			So(err, ShouldBeNil)
// 			So(len(stats), ShouldEqual, 0)
// 		})

// 		Convey("Count failed, caused by the error from method 'acquireIndices'", func() {
// 			expectedErr := errors.New("some errors")

// 			patch1 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "parseQuery",
// 				func(_ *dataViewService, ctx context.Context, viewID string) (*interfaces.DataView, error) {
// 					return dataView, nil
// 				})
// 			defer patch1.Reset()

// 			patch2 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "acquireIndices",
// 				func(_ *dataViewService, ctx context.Context, query *interfaces.DataViewQueryV1) ([]string, error) {
// 					return []string{}, expectedErr
// 				})
// 			defer patch2.Reset()

// 			viewID := "1"
// 			query := &interfaces.DataViewQueryV1{}
// 			_, err := dvs.CountMultiFields(testCtx, viewID, query, []string{"f1", "f2"}, "_")
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("Count failed, caused by the error from method 'prepareDSLParas'", func() {
// 			expectedErr := errors.New("some errors")

// 			patch1 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "parseQuery",
// 				func(_ *dataViewService, ctx context.Context, viewID string) (*interfaces.DataView, error) {
// 					return dataView, nil
// 				})
// 			defer patch1.Reset()

// 			patch2 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "acquireIndices",
// 				func(_ *dataViewService, ctx context.Context, query *interfaces.DataViewQueryV1) ([]string, error) {
// 					return []string{"index"}, nil
// 				})
// 			defer patch2.Reset()

// 			patch3 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "prepareDSLParas",
// 				func(ctx context.Context, query *interfaces.DataViewQueryV1, fields []string) (multiFieldStatsDSLParas, error) {
// 					return multiFieldStatsDSLParas{}, expectedErr
// 				})
// 			defer patch3.Reset()

// 			query := &interfaces.DataViewQueryV1{}
// 			viewID := "1"
// 			_, err := dvs.CountMultiFields(testCtx, viewID, query, []string{"f1", "f2"}, "_")
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("Count failed, caused by the error from method 'generateMultiFieldStatsDSL'", func() {
// 			expectedErr := errors.New("some errors")

// 			patch1 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "parseQuery",
// 				func(_ *dataViewService, ctx context.Context, viewID string) (*interfaces.DataView, error) {
// 					return dataView, nil
// 				})
// 			defer patch1.Reset()

// 			patch2 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "acquireIndices",
// 				func(_ *dataViewService, ctx context.Context, query *interfaces.DataViewQueryV1) ([]string, error) {
// 					return []string{"index"}, nil
// 				})
// 			defer patch2.Reset()

// 			patch3 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "prepareDSLParas",
// 				func(ctx context.Context, query *interfaces.DataViewQueryV1, fields []string) (multiFieldStatsDSLParas, error) {
// 					return multiFieldStatsDSLParas{}, nil
// 				})
// 			defer patch3.Reset()

// 			patch4 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "generateMultiFieldStatsDSL",
// 				func(ctx context.Context, paras multiFieldStatsDSLParas, fields []string) (map[string]any, error) {
// 					return map[string]any{}, expectedErr
// 				})
// 			defer patch4.Reset()

// 			query := &interfaces.DataViewQueryV1{}
// 			viewID := "1"
// 			_, err := dvs.CountMultiFields(testCtx, viewID, query, []string{"f1", "f2"}, "_")
// 			So(err, ShouldResemble, expectedErr)
// 		})

// 		Convey("Count failed, caused by the error from method 'SearchSubmit'", func() {
// 			expectedErr := errors.New("some errors")

// 			patch1 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "parseQuery",
// 				func(_ *dataViewService, ctx context.Context, viewID string) (*interfaces.DataView, error) {
// 					return dataView, nil
// 				})
// 			defer patch1.Reset()

// 			patch2 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "acquireIndices",
// 				func(_ *dataViewService, ctx context.Context, query *interfaces.DataViewQueryV1) ([]string, error) {
// 					return []string{"index"}, nil
// 				})
// 			defer patch2.Reset()

// 			patch3 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "prepareDSLParas",
// 				func(ctx context.Context, query *interfaces.DataViewQueryV1, fields []string) (multiFieldStatsDSLParas, error) {
// 					return multiFieldStatsDSLParas{}, nil
// 				})
// 			defer patch3.Reset()

// 			patch4 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "generateMultiFieldStatsDSL",
// 				func(ctx context.Context, paras multiFieldStatsDSLParas, fields []string) (map[string]any, error) {
// 					return map[string]any{}, nil
// 				})
// 			defer patch4.Reset()

// 			osaMock.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(),
// 				gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte(nil), 0, expectedErr)

// 			viewID := "1"
// 			query := &interfaces.DataViewQueryV1{}
// 			_, err := dvs.CountMultiFields(testCtx, viewID, query, []string{"f1", "f2"}, "_")
// 			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError,
// 				uerrors.Uniquery_InternalError_SearchSubmitFailed).WithErrorDetails(expectedErr.Error()))
// 		})

// 		Convey("Count failed, caused by the error from func 'sonic.Unmarshal'", func() {
// 			expectedErr := errors.New("some errors")

// 			patch1 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "parseQuery",
// 				func(_ *dataViewService, ctx context.Context, viewID string) (*interfaces.DataView, error) {
// 					return dataView, nil
// 				})
// 			defer patch1.Reset()

// 			patch2 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "acquireIndices",
// 				func(_ *dataViewService, ctx context.Context, query *interfaces.DataViewQueryV1) ([]string, error) {
// 					return []string{"index"}, nil
// 				})
// 			defer patch2.Reset()

// 			patch3 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "prepareDSLParas",
// 				func(ctx context.Context, query *interfaces.DataViewQueryV1, fields []string) (multiFieldStatsDSLParas, error) {
// 					return multiFieldStatsDSLParas{}, nil
// 				})
// 			defer patch3.Reset()

// 			patch4 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "generateMultiFieldStatsDSL",
// 				func(ctx context.Context, paras multiFieldStatsDSLParas, fields []string) (map[string]any, error) {
// 					return map[string]any{}, nil
// 				})
// 			defer patch4.Reset()

// 			osaMock.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(),
// 				gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte(nil), 0, nil)

// 			patch5 := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
// 			defer patch5.Reset()

// 			viewID := "1"
// 			query := &interfaces.DataViewQueryV1{}
// 			_, err := dvs.CountMultiFields(testCtx, viewID, query, []string{"f1", "f2"}, "_")
// 			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError,
// 				uerrors.Uniquery_DataView_InternalError_UnmarshalFailed).WithErrorDetails(expectedErr.Error()))
// 		})

// 		Convey("Count succeed", func() {
// 			patch1 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "parseQuery",
// 				func(_ *dataViewService, ctx context.Context, viewID string) (*interfaces.DataView, error) {
// 					return dataView, nil
// 				})
// 			defer patch1.Reset()

// 			patch2 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "acquireIndices",
// 				func(_ *dataViewService, ctx context.Context, query *interfaces.DataViewQueryV1) ([]string, error) {
// 					return []string{"index"}, nil
// 				})
// 			defer patch2.Reset()

// 			patch3 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "prepareDSLParas",
// 				func(ctx context.Context, query *interfaces.DataViewQueryV1, fields []string) (multiFieldStatsDSLParas, error) {
// 					return multiFieldStatsDSLParas{}, nil
// 				})
// 			defer patch3.Reset()

// 			patch4 := ApplyPrivateMethod(reflect.TypeOf(&dataViewService{}), "generateMultiFieldStatsDSL",
// 				func(ctx context.Context, paras multiFieldStatsDSLParas, fields []string) (map[string]any, error) {
// 					return map[string]any{}, nil
// 				})
// 			defer patch4.Reset()

// 			osaMock.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(),
// 				gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte(nil), 0, nil)

// 			patch5 := ApplyFuncReturn(sonic.Unmarshal, nil)
// 			defer patch5.Reset()

// 			viewID := "1"
// 			query := &interfaces.DataViewQueryV1{}
// 			_, err := dvs.CountMultiFields(testCtx, viewID, query, []string{"f1", "f2"}, "_")
// 			So(err, ShouldBeNil)
// 		})
// 	})
// }

func TestParseQuery(t *testing.T) {
	Convey("Test parseQuery", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		dvaMock := mock.NewMockDataViewAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		osaMock := mock.NewMockOpenSearchAccess(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)

		appSetting := &common.AppSetting{}

		dvs := MockNewDataViewService(appSetting, dvaMock, ibaMock, osaMock, psMock)

		Convey("Parse failed, caused by the error from method 'GetDataViewByID'", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyMethodReturn(&dataViewService{}, "GetDataViewByID", &interfaces.DataView{}, expectedErr)
			defer patch.Reset()

			viewID := "1"
			_, err := dvs.parseQuery(testCtx, viewID)
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_InternalError_GetDataViewByIDFailed).WithErrorDetails(expectedErr.Error()))
		})

		Convey("Parse succeed", func() {
			patch := ApplyMethodReturn(&dataViewService{}, "GetDataViewByID", &interfaces.DataView{}, nil)
			defer patch.Reset()

			viewID := "1"
			_, err := dvs.parseQuery(testCtx, viewID)
			So(err, ShouldBeNil)
		})
	})
}

// func TestAcquireIndices(t *testing.T) {
// 	Convey("Test acquireIndices", t, func() {
// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		dvaMock := mock.NewMockDataViewAccess(mockCtrl)
// 		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
// 		osaMock := mock.NewMockOpenSearchAccess(mockCtrl)
// 		psMock := mock.NewMockPermissionService(mockCtrl)

// 		appSetting := &common.AppSetting{}
// 		indicesResult := map[string]map[string]interfaces.Indice{
// 			"indices": {
// 				"x": {
// 					IndexName: "x",
// 					ShardNum:  3,
// 				},
// 				"y": {
// 					IndexName: "y",
// 					ShardNum:  3,
// 				},
// 			},
// 		}

// 		dvs := MockNewDataViewService(appSetting, dvaMock, ibaMock, osaMock, psMock)

// 		Convey("Acquire failed, caused by the error from method 'Decode'", func() {
// 			expectedErr := errors.New("some errors")
// 			patch := ApplyFuncReturn(mapstructure.Decode, expectedErr)
// 			defer patch.Reset()

// 			dataSource := "test"
// 			start := int64(0)
// 			end := int64(0)
// 			_, err := dvs.acquireIndices(testCtx, dataSource, start, end)
// 			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_DataSource).
// 				WithErrorDetails(fmt.Sprintf("mapstructure decode dataSource failed, %s", expectedErr.Error())))
// 		})

// 		Convey("Acquire failed, caused by the error from method 'GetIndices'", func() {
// 			expectedErr := errors.New("some errors")
// 			patch := ApplyFuncReturn(mapstructure.Decode, nil)
// 			defer patch.Reset()

// 			ibaMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
// 				Return(indicesResult, 0, expectedErr).AnyTimes()

// 			dataSource := "test"
// 			start := int64(1716576234000)
// 			end := int64(1716662634000)
// 			_, err := dvs.acquireIndices(testCtx, dataSource, start, end)
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("Acquire succeed'", func() {
// 			patch := ApplyFuncReturn(mapstructure.Decode, nil)
// 			defer patch.Reset()

// 			ibaMock.EXPECT().GetIndices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
// 				Return(indicesResult, 0, nil).AnyTimes()

// 			dataSource := "test"
// 			start := int64(1716576234000)
// 			end := int64(1716662634000)
// 			_, err := dvs.acquireIndices(testCtx, dataSource, start, end)
// 			So(err, ShouldBeNil)
// 		})
// 	})
// }

func TestPrepareDSLParas(t *testing.T) {
	Convey("Test prepareDSLParas", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		dvaMock := mock.NewMockDataViewAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		osaMock := mock.NewMockOpenSearchAccess(mockCtrl)
		conditionMock := cmock.NewMockCondition(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)

		appSetting := &common.AppSetting{}

		dvs := MockNewDataViewService(appSetting, dvaMock, ibaMock, osaMock, psMock)

		Convey("Prepare failed, caused by the error from method 'NewCondition'", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyFuncReturn(cond.NewCondition, conditionMock, false, expectedErr)
			defer patch.Reset()

			view := &interfaces.DataView{}
			_, err := dvs.prepareDSLParas(testCtx, &interfaces.DataViewQueryV1{}, view, []string{"f1"})
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_Filters).
				WithErrorDetails(fmt.Sprintf("failed to new condition, %s", expectedErr.Error())))
		})

		Convey("Prepare failed, caused by the error from method 'Convert'", func() {
			expectedErr := errors.New("some errors")

			patch1 := ApplyFuncReturn(cond.NewCondition, conditionMock, false, nil)
			defer patch1.Reset()

			conditionMock.EXPECT().Convert(gomock.Any()).Return("", expectedErr)

			view := &interfaces.DataView{}
			_, err := dvs.prepareDSLParas(testCtx, &interfaces.DataViewQueryV1{}, view, []string{"f1"})
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_Filters).
				WithErrorDetails(fmt.Sprintf("failed to convert condition to dsl, %s", expectedErr.Error())))
		})

		Convey("Prepare succeed", func() {
			patch := ApplyFuncReturn(cond.NewCondition, conditionMock, false, nil)
			defer patch.Reset()

			conditionMock.EXPECT().Convert(gomock.Any()).Return("", nil)

			view := &interfaces.DataView{}
			_, err := dvs.prepareDSLParas(testCtx, &interfaces.DataViewQueryV1{}, view, []string{"f1"})
			So(err, ShouldBeNil)
		})
	})
}

func TestGenerateMultiFieldStatsDSL(t *testing.T) {
	Convey("Test generateMultiFieldStatsDSL", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		dvaMock := mock.NewMockDataViewAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		osaMock := mock.NewMockOpenSearchAccess(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		appSetting := &common.AppSetting{}

		dvs := MockNewDataViewService(appSetting, dvaMock, ibaMock, osaMock, psMock)

		Convey("Generate failed", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch.Reset()

			_, err := dvs.generateMultiFieldStatsDSL(testCtx, multiFieldStatsDSLParas{
				baseFilterStr:   "xxx",
				timeFilterStr:   "xxx",
				aggTermStr:      "xxx",
				scriptFilterStr: "xxx",
			}, []string{"f1"})
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				uerrors.Uniquery_DataView_InternalError_UnmarshalFailed).WithErrorDetails(expectedErr.Error()))
		})

		Convey("Generate succeed", func() {
			patch := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patch.Reset()

			_, err := dvs.generateMultiFieldStatsDSL(testCtx, multiFieldStatsDSLParas{
				baseFilterStr:   "xxx",
				timeFilterStr:   "xxx",
				aggTermStr:      "xxx",
				scriptFilterStr: "xxx",
			}, []string{"f1", "f2"})
			So(err, ShouldBeNil)
		})
	})
}

func TestHasFieldPermission(t *testing.T) {
	Convey("Test HasFieldPermission", t, func() {
		viewFieldsMap := map[string]*cond.ViewField{
			"a": {},
			"b": {},
			"c": {},
			"d": {},
			"e": {},
			"f": {},
			"g": {},
			"h": {},
		}

		Convey("simple cfg, has Permission", func() {
			cfg := &cond.CondCfg{
				Name: "a",
			}
			_, res := checkConditionFieldExist(viewFieldsMap, cfg)
			So(res, ShouldBeTrue)
		})

		Convey("nested cfg, has Permission testcase 1", func() {
			cfg := &cond.CondCfg{
				Operation: cond.OperationAnd,
				SubConds: []*cond.CondCfg{
					{Name: "a"},
					{
						Operation: cond.OperationOr,
						SubConds: []*cond.CondCfg{
							{
								Operation: cond.OperationAnd,
								SubConds: []*cond.CondCfg{
									{Name: "h"},
								},
							},
						},
					},
				},
			}
			_, res := checkConditionFieldExist(viewFieldsMap, cfg)
			So(res, ShouldBeTrue)
		})

		Convey("nested cfg, has Permission testcase 2", func() {
			cfg := &cond.CondCfg{
				Operation: cond.OperationAnd,
				SubConds: []*cond.CondCfg{
					{Name: "a"},
					{Name: "e"},
					{
						Operation: cond.OperationOr,
						SubConds: []*cond.CondCfg{
							{Name: "g"},
							{Name: "h"},
							{
								Operation: cond.OperationAnd,
								SubConds: []*cond.CondCfg{
									{Name: "h"},
									{Name: "d"},
									{
										Operation: cond.OperationOr,
										SubConds: []*cond.CondCfg{
											{Name: "e"},
											{Name: "f"},
											{
												Operation: cond.OperationAnd,
												SubConds: []*cond.CondCfg{
													{Name: "c"},
													{Name: "b"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}
			_, res := checkConditionFieldExist(viewFieldsMap, cfg)
			So(res, ShouldBeTrue)
		})

		Convey("simple cfg,, does not has Permission", func() {
			cfg := &cond.CondCfg{
				Name: "aaa",
			}
			_, res := checkConditionFieldExist(viewFieldsMap, cfg)
			So(res, ShouldBeFalse)
		})

		Convey("nested cfg, does not has Permission testcase 1", func() {
			cfg := &cond.CondCfg{
				Operation: cond.OperationAnd,
				SubConds: []*cond.CondCfg{
					{Name: "a"},
					{
						Operation: cond.OperationOr,
						SubConds: []*cond.CondCfg{
							{
								Operation: cond.OperationAnd,
								SubConds: []*cond.CondCfg{
									{Name: "hhhhh"},
								},
							},
						},
					},
				},
			}
			_, res := checkConditionFieldExist(viewFieldsMap, cfg)
			So(res, ShouldBeFalse)
		})

		Convey("nested cfg, does not has Permission testcase 2", func() {
			cfg := &cond.CondCfg{
				Operation: cond.OperationAnd,
				SubConds: []*cond.CondCfg{
					{Name: "a"},
					{Name: "aaaa"},
					{
						Operation: cond.OperationOr,
						SubConds: []*cond.CondCfg{
							{
								Operation: cond.OperationAnd,
								SubConds: []*cond.CondCfg{
									{Name: "h"},
								},
							},
						},
					},
				},
			}
			_, res := checkConditionFieldExist(viewFieldsMap, cfg)
			So(res, ShouldBeFalse)
		})

		Convey("nested cfg, does not has Permission testcase 3", func() {
			cfg := &cond.CondCfg{
				Operation: cond.OperationAnd,
				SubConds: []*cond.CondCfg{
					{Name: "a"},
					{Name: "e"},
					{
						Operation: cond.OperationOr,
						SubConds: []*cond.CondCfg{
							{Name: "g"},
							{Name: "h"},
							{
								Operation: cond.OperationAnd,
								SubConds: []*cond.CondCfg{
									{Name: "h"},
									{Name: "d"},
									{
										Operation: cond.OperationOr,
										SubConds: []*cond.CondCfg{
											{Name: "e"},
											{Name: "f"},
											{
												Operation: cond.OperationAnd,
												SubConds: []*cond.CondCfg{
													{Name: "c"},
													{Name: "ddddd"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}
			_, res := checkConditionFieldExist(viewFieldsMap, cfg)
			So(res, ShouldBeFalse)
		})

		Convey("nested cfg, does not has Permission testcase 4", func() {
			cfg := &cond.CondCfg{
				Operation: cond.OperationAnd,
				SubConds: []*cond.CondCfg{
					{Name: "a"},
					{Name: "e"},
					{
						Operation: cond.OperationOr,
						SubConds: []*cond.CondCfg{
							{Name: "g"},
							{Name: "h"},
							{
								Operation: cond.OperationAnd,
								SubConds: []*cond.CondCfg{
									{Name: "hhh"},
									{Name: "d"},
									{
										Operation: cond.OperationOr,
										SubConds: []*cond.CondCfg{
											{Name: "e"},
											{Name: "f"},
											{
												Operation: cond.OperationAnd,
												SubConds: []*cond.CondCfg{
													{Name: "c"},
													{Name: "ddddd"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}
			_, res := checkConditionFieldExist(viewFieldsMap, cfg)
			So(res, ShouldBeFalse)
		})
	})
}

func TestGetSearchAfterDSL(t *testing.T) {
	Convey("Test getSearchAfterDSL", t, func() {
		Convey("Get succeed, because search_after params is nil", func() {
			_, err := getSearchAfterDSL(nil)
			So(err, ShouldBeNil)
		})

		Convey("Get succeed, have pit and have search_after", func() {
			dsl, err := getSearchAfterDSL(&interfaces.SearchAfterParams{
				SearchAfter:  []any{"1", "2"},
				PitID:        "aaa",
				PitKeepAlive: "5m",
			})
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
		})

		Convey("Get succeed, have pit and no search_after", func() {
			dsl, err := getSearchAfterDSL(&interfaces.SearchAfterParams{
				PitID:        "aaa",
				PitKeepAlive: "5m",
			})

			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
		})

		Convey("Get succeed, no pit and have search_after", func() {
			dsl, err := getSearchAfterDSL(&interfaces.SearchAfterParams{
				SearchAfter: []any{"1", "2"},
			})
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
		})
	})
}

func TestGetDataFromOpenSearch(t *testing.T) {
	Convey("Test GetDataFromOpenSearch", t, func() {
		Convey("Get succeed", func() {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			dvaMock := mock.NewMockDataViewAccess(mockCtrl)
			ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
			osaMock := mock.NewMockOpenSearchAccess(mockCtrl)
			psMock := mock.NewMockPermissionService(mockCtrl)
			appSetting := &common.AppSetting{}

			dvsMock := MockNewDataViewService(appSetting, dvaMock, ibaMock, osaMock, psMock)

			Convey("Get succeed", func() {
				osaMock.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, 200, nil)
				query := map[string]any{}
				indices := []string{"index1", "index2"}

				_, _, err := dvsMock.GetDataFromOpenSearch(testCtx, query, indices, 0, "", false)
				So(err, ShouldBeNil)
			})

			Convey("Get failed, because search submit failed", func() {
				osaMock.EXPECT().SearchSubmit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(nil, 500, errors.New("some errors"))
				query := map[string]any{}
				indices := []string{"index1", "index2"}
				_, _, err := dvsMock.GetDataFromOpenSearch(testCtx, query, indices, 0, "", false)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestGetDataFromOpenSearchWithBuffer(t *testing.T) {
	Convey("Test GetDataFromOpenSearchWithBuffer", t, func() {
		Convey("Get succeed", func() {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			dvaMock := mock.NewMockDataViewAccess(mockCtrl)
			ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
			osaMock := mock.NewMockOpenSearchAccess(mockCtrl)
			psMock := mock.NewMockPermissionService(mockCtrl)
			appSetting := &common.AppSetting{}

			dvsMock := MockNewDataViewService(appSetting, dvaMock, ibaMock, osaMock, psMock)

			Convey("Get succeed", func() {
				osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return([]byte{}, 200, nil)
				query := bytes.Buffer{}
				indices := []string{"index1", "index2"}

				_, _, err := dvsMock.GetDataFromOpenSearchWithBuffer(testCtx, query, indices, 0, "")
				So(err, ShouldBeNil)
			})

			Convey("Get failed, because search submit failed", func() {
				osaMock.EXPECT().SearchSubmitWithBuffer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 500, errors.New("some errors"))
				query := bytes.Buffer{}
				indices := []string{"index1", "index2"}
				_, _, err := dvsMock.GetDataFromOpenSearchWithBuffer(testCtx, query, indices, 0, "")
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestDeleteDataViewPits(t *testing.T) {
	Convey("Test DeleteDataViewPits", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		dvaMock := mock.NewMockDataViewAccess(mockCtrl)
		ibaMock := mock.NewMockIndexBaseAccess(mockCtrl)
		osaMock := mock.NewMockOpenSearchAccess(mockCtrl)
		psMock := mock.NewMockPermissionService(mockCtrl)
		appSetting := &common.AppSetting{}

		dvsMock := MockNewDataViewService(appSetting, dvaMock, ibaMock, osaMock, psMock)

		Convey("Delete succeed, pit id is __all", func() {
			osaMock.EXPECT().DeletePointInTime(gomock.Any(), gomock.Any()).Return(&interfaces.DeletePitsResp{}, 200, nil)
			pits := &interfaces.DeletePits{
				PitIDs: []string{interfaces.All_Pits_DataView},
			}
			_, err := dvsMock.DeleteDataViewPits(testCtx, pits)
			So(err, ShouldBeNil)
		})

		Convey("Delete succeed, pit id is not __all", func() {
			osaMock.EXPECT().DeletePointInTime(gomock.Any(), gomock.Any()).Return(&interfaces.DeletePitsResp{}, 200, nil)
			pits := &interfaces.DeletePits{
				PitIDs: []string{"1", "2"},
			}
			_, err := dvsMock.DeleteDataViewPits(testCtx, pits)
			So(err, ShouldBeNil)
		})

		Convey("Delete succeed, pit id is __all and other pit id is not __all", func() {
			osaMock.EXPECT().DeletePointInTime(gomock.Any(), gomock.Any()).Return(&interfaces.DeletePitsResp{}, 200, nil)
			pits := &interfaces.DeletePits{
				PitIDs: []string{interfaces.All_Pits_DataView, "1", "2"},
			}
			_, err := dvsMock.DeleteDataViewPits(testCtx, pits)
			So(err, ShouldBeNil)
		})
	})
}
