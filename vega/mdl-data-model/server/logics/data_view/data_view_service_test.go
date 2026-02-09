// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_view

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	dcond "data-model/interfaces/condition"
	dmock "data-model/interfaces/mock"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

func MockNewDataViewService(appSetting *common.AppSetting,
	dvga interfaces.DataViewGroupAccess,
	dva interfaces.DataViewAccess,
	iba interfaces.IndexBaseAccess,
	dmja interfaces.DataModelJobAccess,
	ps interfaces.PermissionService) (*dataViewService, sqlmock.Sqlmock) {

	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	return &dataViewService{
		appSetting: appSetting,
		dvga:       dvga,
		dva:        dva,
		iba:        iba,
		dmja:       dmja,
		db:         db,
		ps:         ps,
	}, smock
}

// func Test_DataViewService_CreateDataViews(t *testing.T) {
// 	Convey("Test CreateDataViews", t, func() {
// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		appSetting := &common.AppSetting{}
// 		dva := dmock.NewMockDataViewAccess(mockCtrl)
// 		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
// 		dvga := dmock.NewMockDataViewGroupAccess(mockCtrl)
// 		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
// 		ps := dmock.NewMockPermissionService(mockCtrl)
// 		dvs, smock := MockNewDataViewService(appSetting, dvga, dva, iba, dmja, ps)

// 		views := []*interfaces.DataView{
// 			{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					DataSource: map[string]any{
// 						"type": "index_base",
// 						"index_base": []any{
// 							interfaces.SimpleIndexBase{BaseType: "x"},
// 							interfaces.SimpleIndexBase{BaseType: "y"},
// 						}},
// 					GroupName: "test",
// 				},
// 				Fields: []*interfaces.ViewField{
// 					{
// 						Name: "@timestamp",
// 						Type: "date",
// 					},
// 					{
// 						Name: "zzz",
// 						Type: "long",
// 					},
// 				},
// 			},
// 		}

// 		sqlStr := "select * from vdm_mysql_0leb5aws.default.test"

// 		group := &interfaces.DataViewGroup{
// 			GroupName: "xxx",
// 			GroupID:   "xxx",
// 			Builtin:   false,
// 		}

// 		Convey("Create failed, caused by the error getIndexBaseByNameFailed", func() {
// 			expectedErr := errors.New("some errors")
// 			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_GetIndexBaseByTypeFailed).
// 				WithErrorDetails(expectedErr.Error())

// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

// 			_, httpErr := dvs.CreateDataViews(testCtx, nil, views, sqlStr, interfaces.ImportMode_Normal)
// 			So(httpErr, ShouldResemble, expectedHttpErr)
// 		})

// 		Convey("Create failed, caused by the error field type conflict", func() {
// 			baseInfos := []interfaces.IndexBase{
// 				{
// 					SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "xxx"},
// 					Mappings: interfaces.Mappings{
// 						DynamicMappings: []interfaces.IndexBaseField{
// 							{Field: "zzz", Type: "long"},
// 						},
// 					},
// 				},
// 				{
// 					SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "xxx"},
// 					Mappings: interfaces.Mappings{
// 						DynamicMappings: []interfaces.IndexBaseField{
// 							{Field: "zzz", Type: "text"},
// 						}},
// 				},
// 			}
// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(baseInfos, nil)
// 			_, httpErr := dvs.CreateDataViews(testCtx, nil, views, sqlStr, interfaces.ImportMode_Normal)
// 			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_FieldTypeConflict)
// 		})

// 		Convey("Create failed, caused by the error from begin db transaction", func() {
// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			expectedErr := errors.New("some errors")
// 			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, nil)
// 			smock.ExpectBegin().WillReturnError(expectedErr)

// 			_, httpErr := dvs.CreateDataViews(testCtx,nil, views,sqlStr, interfaces.ImportMode_Normal)
// 			So(httpErr, ShouldNotBeNil)
// 		})

// 		Convey("Create succeed and field scope is 0", func() {
// 			// 分组名称为 ""，不需要检查分组名称是否存在，直接返回的未分组的 id
// 			viewsTest := []*interfaces.DataView{
// 				{
// 					SimpleDataView: interfaces.SimpleDataView{
// 						DataSource: map[string]any{
// 							"type": "index_base",
// 							"index_base": []any{
// 								interfaces.SimpleIndexBase{BaseType: "x"},
// 								interfaces.SimpleIndexBase{BaseType: "y"},
// 							}},
// 						GroupName: "",
// 					},
// 					Fields: []*interfaces.ViewField{
// 						{
// 							Name: "@timestamp",
// 							Type: "date",
// 						},
// 					},
// 				},
// 			}

// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, nil)
// 			smock.ExpectBegin()
// 			dva.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
// 			dva.EXPECT().CheckDataViewExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
// 			dvga.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(group, true, nil)
// 			dva.EXPECT().CreateDataViews(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			smock.ExpectCommit()
// 			ps.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

// 			_, httpErr := dvs.CreateDataViews(testCtx,nil,  viewsTest,sqlStr, interfaces.ImportMode_Normal)
// 			So(httpErr, ShouldBeNil)
// 		})

// 		Convey("Create succeed and field scope is 1", func() {
// 			// 分组名称为 "test"，检查分组名称是否存在，mock返回存在
// 			viewsTest := []*interfaces.DataView{
// 				{
// 					SimpleDataView: interfaces.SimpleDataView{
// 						DataSource: map[string]any{
// 							"type": "index_base",
// 							"index_base": []any{
// 								interfaces.SimpleIndexBase{BaseType: "x"},
// 								interfaces.SimpleIndexBase{BaseType: "y"},
// 							}},
// 						GroupName: "test",
// 					},
// 					Fields: []*interfaces.ViewField{
// 						{
// 							Name: "@timestamp",
// 							Type: "date",
// 						},
// 					},
// 				},
// 			}

// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, nil)
// 			smock.ExpectBegin()
// 			dva.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
// 			dva.EXPECT().CheckDataViewExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
// 			dvga.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(group, true, nil)
// 			dva.EXPECT().CreateDataViews(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			smock.ExpectCommit()
// 			ps.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

// 			_, httpErr := dvs.CreateDataViews(testCtx, nil,viewsTest, sqlStr, interfaces.ImportMode_Normal)
// 			So(httpErr, ShouldBeNil)
// 		})
// 	})
// }

// func Test_DataViewService_DeleteDataViews(t *testing.T) {
// 	Convey("Test DeleteDataViews", t, func() {
// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		appSetting := &common.AppSetting{}
// 		dva := dmock.NewMockDataViewAccess(mockCtrl)
// 		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
// 		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
// 		dvga := dmock.NewMockDataViewGroupAccess(mockCtrl)
// 		ps := dmock.NewMockPermissionService(mockCtrl)
// 		dvs, smock := MockNewDataViewService(appSetting, dvga, dva, iba, dmja, ps)

// 		jobs := []*interfaces.JobInfo{
// 			{JobID: "2q"},
// 			{JobID: "3s"},
// 		}
// 		resrc := map[string]interfaces.ResourceOps{
// 			"0": {
// 				ResourceID: "0",
// 			},
// 			"1": {
// 				ResourceID: "1",
// 			},
// 		}

// 		Convey("Delete succeed", func() {
// 			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
// 				Return(resrc, nil)
// 			ps.EXPECT().DeleteResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dva.EXPECT().GetJobsByDataViewIDs(gomock.Any(), gomock.Any()).Return(jobs, nil)
// 			smock.ExpectBegin()
// 			dva.EXPECT().DeleteDataViews(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dmja.EXPECT().DeleteDataModelJobs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			smock.ExpectCommit()

// 			dmja.EXPECT().StopJobs(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

// 			httpErr := dvs.DeleteDataViews(testCtx, []string{"0", "1"})

// 			So(httpErr, ShouldBeNil)
// 		})

// 		Convey("Delete failed", func() {
// 			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
// 				Return(resrc, nil)
// 			expectedErr := errors.New("some errors")
// 			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_DeleteDataViewsFailed).
// 				WithErrorDetails(expectedErr.Error())

// 			dva.EXPECT().GetJobsByDataViewIDs(gomock.Any(), gomock.Any()).Return(jobs, nil)
// 			dmja.EXPECT().StopJobs(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
// 			smock.ExpectBegin()
// 			dva.EXPECT().DeleteDataViews(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)
// 			smock.ExpectRollback()

// 			httpErr := dvs.DeleteDataViews(testCtx, []string{"0", "1"})

// 			So(httpErr, ShouldResemble, expectedHttpErr)
// 		})
// 	})
// }

// func Test_DataViewService_UpdateDataView(t *testing.T) {
// 	Convey("Test UpdateDataView", t, func() {
// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		appSetting := &common.AppSetting{}
// 		dva := dmock.NewMockDataViewAccess(mockCtrl)
// 		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
// 		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
// 		dvga := dmock.NewMockDataViewGroupAccess(mockCtrl)
// 		ps := dmock.NewMockPermissionService(mockCtrl)
// 		dvs, smock := MockNewDataViewService(appSetting, dvga, dva, iba, dmja, ps)

// 		oldView := &interfaces.DataView{
// 			SimpleDataView: interfaces.SimpleDataView{
// 				DataSource: map[string]any{
// 					"type": "index_base",
// 					"index_base": []any{
// 						interfaces.SimpleIndexBase{BaseType: "x"},
// 						interfaces.SimpleIndexBase{BaseType: "y"},
// 					}},
// 			},
// 			Fields: []*interfaces.ViewField{
// 				{
// 					Name: "@timestamp",
// 					Type: "date",
// 				},
// 			},
// 		}
// 		view := &interfaces.DataView{
// 			SimpleDataView: interfaces.SimpleDataView{
// 				DataSource: map[string]any{
// 					"type": "index_base",
// 					"index_base": []any{
// 						interfaces.SimpleIndexBase{BaseType: "x"},
// 						interfaces.SimpleIndexBase{BaseType: "y"},
// 					}},
// 				GroupID: "nnn",
// 			},
// 			Fields: []*interfaces.ViewField{
// 				{
// 					Name: "@timestamp",
// 					Type: "date",
// 				},
// 				{
// 					Name: "zzz",
// 					Type: "long",
// 				},
// 			},
// 		}

// 		group := &interfaces.DataViewGroup{
// 			GroupName: "xxx",
// 			GroupID:   "xxx",
// 			Builtin:   false,
// 		}

// 		Convey("Update failed, caused by the error getIndexBaseByNameFailed", func() {
// 			expectedErr := errors.New("some errors")
// 			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_GetIndexBaseByTypeFailed).
// 				WithErrorDetails(expectedErr.Error())

// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dva.EXPECT().GetDataViews(gomock.Any(), gomock.Any()).Return([]*interfaces.DataView{oldView}, nil)
// 			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

// 			httpErr := dvs.UpdateDataView(testCtx, nil, view)
// 			So(httpErr, ShouldResemble, expectedHttpErr)
// 		})

// 		Convey("Update failed, caused by the error field type conflict", func() {
// 			baseInfos := []interfaces.IndexBase{
// 				{
// 					SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "xxx"},
// 					Mappings: interfaces.Mappings{
// 						DynamicMappings: []interfaces.IndexBaseField{
// 							{Field: "zzz", Type: "long"},
// 						},
// 					},
// 				},
// 				{
// 					SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "xxx"},
// 					Mappings: interfaces.Mappings{
// 						DynamicMappings: []interfaces.IndexBaseField{
// 							{Field: "zzz", Type: "text"},
// 						}},
// 				},
// 			}

// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dva.EXPECT().GetDataViews(gomock.Any(), gomock.Any()).Return([]*interfaces.DataView{oldView}, nil)
// 			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(baseInfos, nil)

// 			err := dvs.UpdateDataView(testCtx, nil, view)
// 			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_FieldTypeConflict)
// 		})

// 		Convey("Update failed, cased by the error from begin tx failed", func() {
// 			expectedErr := errors.New("some errors")
// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dva.EXPECT().GetDataViews(gomock.Any(), gomock.Any()).Return([]*interfaces.DataView{oldView}, nil)
// 			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, nil)
// 			smock.ExpectBegin().WillReturnError(expectedErr)

// 			err := dvs.UpdateDataView(testCtx, nil, view)
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("Update failed, cased by the error from retrive group id by group name failed", func() {
// 			expectedErr := errors.New("some errors")
// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dva.EXPECT().GetDataViews(gomock.Any(), gomock.Any()).Return([]*interfaces.DataView{oldView}, nil)
// 			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, nil)
// 			smock.ExpectBegin()
// 			dvga.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(group, true, expectedErr)
// 			smock.ExpectCommit()

// 			err := dvs.UpdateDataView(testCtx, nil, view)
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("Update failed, cased by the error from check data view exist by name failed", func() {
// 			expectedErr := errors.New("some errors")
// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dva.EXPECT().GetDataViews(gomock.Any(), gomock.Any()).Return([]*interfaces.DataView{oldView}, nil)
// 			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, nil)
// 			smock.ExpectBegin()
// 			dvga.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(group, true, nil)
// 			dva.EXPECT().CheckDataViewExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, expectedErr)
// 			smock.ExpectCommit()

// 			err := dvs.UpdateDataView(testCtx, nil, view)
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("Update failed, cased by the error from exist in same group failed", func() {
// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dva.EXPECT().GetDataViews(gomock.Any(), gomock.Any()).Return([]*interfaces.DataView{oldView}, nil)
// 			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, nil)
// 			smock.ExpectBegin()
// 			dvga.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(group, true, nil)
// 			dva.EXPECT().CheckDataViewExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", true, nil)
// 			smock.ExpectCommit()

// 			err := dvs.UpdateDataView(testCtx, nil, view)
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("Update failed, caused by the error from driven method UpdateDataView", func() {
// 			expectedErr := errors.New("some errors")
// 			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_UpdateDataViewFailed).
// 				WithErrorDetails(expectedErr.Error())
// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			dva.EXPECT().GetDataViews(gomock.Any(), gomock.Any()).Return([]*interfaces.DataView{oldView}, nil)
// 			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, nil)
// 			smock.ExpectBegin()
// 			dvga.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(group, true, nil)
// 			dva.EXPECT().CheckDataViewExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
// 			dva.EXPECT().UpdateDataView(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)
// 			smock.ExpectCommit()

// 			err := dvs.UpdateDataView(testCtx, nil, view)
// 			So(err, ShouldResemble, expectedHttpErr)
// 		})

// 		Convey("Update succeed and field scope is 0", func() {
// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, nil)
// 			dva.EXPECT().GetDataViews(gomock.Any(), gomock.Any()).Return([]*interfaces.DataView{oldView}, nil)
// 			smock.ExpectBegin()
// 			dvga.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(group, true, nil)
// 			dva.EXPECT().CheckDataViewExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
// 			dva.EXPECT().UpdateDataView(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			smock.ExpectCommit()

// 			err := dvs.UpdateDataView(testCtx, nil, view)
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("Update succeed and field scope is 1", func() {
// 			view := &interfaces.DataView{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					DataSource: map[string]any{
// 						"type": "index_base",
// 						"index_base": []any{
// 							interfaces.SimpleIndexBase{BaseType: "x"},
// 							interfaces.SimpleIndexBase{BaseType: "y"},
// 						}},
// 				},
// 				Fields: []*interfaces.ViewField{
// 					{
// 						Name: "@timestamp",
// 						Type: "date",
// 					},
// 				},
// 			}

// 			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, nil)
// 			dva.EXPECT().GetDataViews(gomock.Any(), gomock.Any()).Return([]*interfaces.DataView{oldView}, nil)
// 			smock.ExpectBegin()
// 			dvga.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(group, true, nil)
// 			dva.EXPECT().CheckDataViewExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
// 			dva.EXPECT().UpdateDataView(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
// 			smock.ExpectCommit()

// 			err := dvs.UpdateDataView(testCtx, nil, view)
// 			So(err, ShouldBeNil)
// 		})
// 	})
// }

// func Test_DataViewService_GetDataViews(t *testing.T) {
// 	Convey("Test GetDataViews", t, func() {
// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		appSetting := &common.AppSetting{}
// 		dva := dmock.NewMockDataViewAccess(mockCtrl)
// 		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
// 		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
// 		dvga := dmock.NewMockDataViewGroupAccess(mockCtrl)
// 		ps := dmock.NewMockPermissionService(mockCtrl)
// 		dvs, _ := MockNewDataViewService(appSetting, dvga, dva, iba, dmja, ps)

// 		baseInfos := []interfaces.IndexBase{
// 			{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "xxx"}},
// 			{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "xxx"}},
// 		}

// 		expectedRes := []*interfaces.DataView{
// 			{
// 				SimpleDataView: interfaces.SimpleDataView{
// 					ViewID: "1a",
// 					DataSource: map[string]any{
// 						"type": "index_base",
// 						"index_base": []any{
// 							interfaces.SimpleIndexBase{BaseType: "1a"},
// 							interfaces.SimpleIndexBase{BaseType: "2c"},
// 						},
// 					},
// 				},
// 			},
// 		}

// 		resrc := map[string]interfaces.ResourceOps{
// 			"1": {
// 				ResourceID: "1",
// 			},
// 		}

// 		Convey("Get failed, caused by the error from driven method GetDataViews", func() {
// 			expectedErr := errors.New("some errors")
// 			dva.EXPECT().GetDataViews(gomock.Any(), gomock.Any()).Return(nil, expectedErr)
// 			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_GetDataViewsFailed).
// 				WithErrorDetails(expectedErr.Error())

// 			modelIDs := []string{"1"}
// 			_, httpErr := dvs.GetDataViews(testCtx, modelIDs)
// 			So(httpErr, ShouldResemble, expectedHttpErr)
// 		})

// 		Convey("Get failed, caused by some data views are not found", func() {
// 			dva.EXPECT().GetDataViews(gomock.Any(), gomock.Any()).Return(nil, nil)

// 			modelIDs := []string{"1"}
// 			_, err := dvs.GetDataViews(testCtx, modelIDs)
// 			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusNotFound)
// 			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_DataView_DataViewNotFound)
// 		})

// 		Convey("GetIndexBasesByTypes failed, return nil", func() {
// 			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(resrc, nil)
// 			modelIDs := []string{"1"}

// 			expectedErr := errors.New("some errors")
// 			dva.EXPECT().GetDataViews(gomock.Any(), gomock.Any()).Return(expectedRes, nil)
// 			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

// 			_, httpErr := dvs.GetDataViews(testCtx, modelIDs)
// 			So(httpErr, ShouldBeNil)
// 		})

// 		Convey("Get succeed", func() {
// 			modelIDs := []string{"1"}

// 			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(resrc, nil)
// 			dva.EXPECT().GetDataViews(gomock.Any(), gomock.Any()).Return(expectedRes, nil)
// 			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(baseInfos, nil)

// 			_, httpErr := dvs.GetDataViews(testCtx, modelIDs)
// 			So(httpErr, ShouldBeNil)
// 		})
// 	})
// }

// func Test_DataViewService_ListDataViews(t *testing.T) {
// 	Convey("Test ListDataViews", t, func() {
// 		mockCtrl := gomock.NewController(t)
// 		defer mockCtrl.Finish()

// 		appSetting := &common.AppSetting{}
// 		dva := dmock.NewMockDataViewAccess(mockCtrl)
// 		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
// 		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
// 		dvga := dmock.NewMockDataViewGroupAccess(mockCtrl)
// 		ps := dmock.NewMockPermissionService(mockCtrl)
// 		dvs, _ := MockNewDataViewService(appSetting, dvga, dva, iba, dmja, ps)

// 		// baseInfos := []interfaces.IndexBase{
// 		// 	{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "xxx"}},
// 		// 	{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "xxx"}},
// 		// }

// 		Convey("Get failed, caused by the error from getting the entries of data views", func() {
// 			resrc := map[string]interfaces.ResourceOps{
// 				"1": {
// 					ResourceID: "1",
// 				},
// 			}

// 			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
// 				Return(resrc, nil).AnyTimes()

// 			expectedErr := errors.New("some errors")
// 			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_ListDataViewsFailed).
// 				WithErrorDetails(expectedErr.Error())

// 			expectedEntries := []*interfaces.SimpleDataView{}

// 			dva.EXPECT().ListDataViews(gomock.Any(), gomock.Any()).Return(expectedEntries, expectedErr)
// 			_, _, httpErr := dvs.ListDataViews(testCtx, &interfaces.ListViewQueryParams{})

// 			So(httpErr, ShouldResemble, expectedHttpErr)
// 		})

// 		// Convey("Get failed, caused by the error GetIndexBasesByTypes", func() {
// 		// 	expectedErr := errors.New("some errors")
// 		// 	expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_GetIndexBaseByTypeFailed).
// 		// 		WithErrorDetails(expectedErr.Error())
// 		// 	expectedEntries := []interfaces.SimpleDataView{
// 		// 		{
// 		// 			ViewName:   "test1",
// 		// 			ViewID:     "101",
// 		// 			UpdateTime: "2006-01-02 15:04:05",
// 		// 			DataSource: map[string]any{
// 		// 				"type": "index_base",
// 		// 				"index_base": []any{
// 		// 					interfaces.SimpleIndexBase{BaseType: "1a"},
// 		// 					interfaces.SimpleIndexBase{BaseType: "2c"},
// 		// 				}},
// 		// 		},
// 		// 	}

// 		// 	dva.EXPECT().ListDataViews(gomock.Any()).Return(expectedEntries, nil)
// 		// 	iba.EXPECT().GetIndexBasesByTypes(gomock.Any()).Return(nil, expectedErr)

// 		// 	_, _, httpErr := dvs.ListDataViews(testCtx, interfaces.ListViewQueryParams{})

// 		// 	So(httpErr, ShouldResemble, expectedHttpErr)
// 		// })

// 		// Convey("Get failed, caused by the error from getting the total of data views", func() {
// 		// 	resrc := map[string]interfaces.ResourceOps{
// 		// 		"1": {
// 		// 			ResourceID: "1",
// 		// 		},
// 		// 	}

// 		// 	ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
// 		// 		Return(resrc, nil)

// 		// 	expectedErr := errors.New("some errors")
// 		// 	expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
// 		// 		derrors.DataModel_DataView_InternalError_GetDataViewsTotalFailed).WithErrorDetails(expectedErr.Error())
// 		// 	expectedEntries := []*interfaces.SimpleDataView{
// 		// 		{
// 		// 			ViewName:   "test1",
// 		// 			ViewID:     "101",
// 		// 			UpdateTime: int64(1729651127346),
// 		// 			DataSource: map[string]any{
// 		// 				"type": "index_base",
// 		// 				"index_base": []any{
// 		// 					interfaces.SimpleIndexBase{BaseType: "1a"},
// 		// 					interfaces.SimpleIndexBase{BaseType: "2c"},
// 		// 				},
// 		// 			},
// 		// 		},
// 		// 	}

// 		// 	dva.EXPECT().ListDataViews(gomock.Any(), gomock.Any()).Return(expectedEntries, nil)
// 		// 	dva.EXPECT().GetDataViewsTotal(gomock.Any(), gomock.Any()).Return(0, expectedErr).AnyTimes()
// 		// 	_, _, httpErr := dvs.ListDataViews(testCtx, &interfaces.ListViewQueryParams{})

// 		// 	So(httpErr, ShouldResemble, expectedHttpErr)
// 		// })

// 		Convey("Get succeed", func() {
// 			resrc := map[string]interfaces.ResourceOps{
// 				"1": {
// 					ResourceID: "1",
// 				},
// 			}

// 			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
// 				Return(resrc, nil)

// 			expectedEntries := []*interfaces.SimpleDataView{
// 				{
// 					ViewName:   "test1",
// 					ViewID:     "101",
// 					UpdateTime: int64(1729651127346),
// 					DataSource: map[string]any{
// 						"type": "index_base",
// 						"index_base": []any{
// 							interfaces.SimpleIndexBase{BaseType: "1a"},
// 							interfaces.SimpleIndexBase{BaseType: "2c"},
// 						},
// 					},
// 				},
// 			}

// 			// expectedTotal := 1

// 			dva.EXPECT().ListDataViews(gomock.Any(), gomock.Any()).Return(expectedEntries, nil)
// 			// dva.EXPECT().GetDataViewsTotal(gomock.Any(), gomock.Any()).Return(expectedTotal, nil)
// 			_, _, httpErr := dvs.ListDataViews(testCtx, &interfaces.ListViewQueryParams{})

// 			// So(entries, ShouldResemble, expectedEntries)
// 			// So(total, ShouldEqual, expectedTotal)
// 			So(httpErr, ShouldBeNil)
// 		})
// 	})
// }

func Test_DataViewService_CheckDataViewExistByName(t *testing.T) {
	Convey("Test CheckDataViewExistByName", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		viewID := "1"
		viewName := "test"
		viewGroupName := "testGroup"

		appSetting := &common.AppSetting{}
		dva := dmock.NewMockDataViewAccess(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		dvga := dmock.NewMockDataViewGroupAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		dvs, _ := MockNewDataViewService(appSetting, dvga, dva, iba, dmja, ps)

		Convey("Check failed, caused by the error from driven method CheckDataViewExistByName", func() {
			expectedErr := errors.New("some error")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_CheckViewIfExistFailed).
				WithErrorDetails(expectedErr.Error())

			dva.EXPECT().CheckDataViewExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(viewID, false, expectedErr)

			tx, _ := dvs.db.Begin()
			_, _, err := dvs.CheckDataViewExistByName(testCtx, tx, viewName, viewGroupName)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Get succeed", func() {
			dva.EXPECT().CheckDataViewExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(viewID, false, nil)

			tx, _ := dvs.db.Begin()
			_, _, err := dvs.CheckDataViewExistByName(testCtx, tx, viewName, viewGroupName)
			So(err, ShouldBeNil)
		})
	})
}

func Test_DataViewService_CheckDataViewExistByID(t *testing.T) {
	Convey("Test CheckDataViewExistByID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dva := dmock.NewMockDataViewAccess(mockCtrl)
		iba := dmock.NewMockIndexBaseAccess(mockCtrl)
		dmja := dmock.NewMockDataModelJobAccess(mockCtrl)
		dvga := dmock.NewMockDataViewGroupAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		dvs, _ := MockNewDataViewService(appSetting, dvga, dva, iba, dmja, ps)

		Convey("Check failed, caused by the error from driven method CheckDataViewExistByID", func() {
			expectedErr := errors.New("some errors")
			dva.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, expectedErr)

			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_DataView_InternalError_CheckViewIfExistFailed).
				WithErrorDetails(expectedErr.Error())

			tx, _ := dvs.db.Begin()
			_, _, err := dvs.CheckDataViewExistByID(testCtx, tx, "a")
			So(err, ShouldResemble, expectedHttpErr)
		})

		// Convey("Get failed, caused by data view does not exist", func() {
		// 	expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusNotFound, derrors.DataModel_DataView_DataViewNotFound)
		// 	dva.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any()).Return("", false, nil)

		// 	_, _, err := dvs.CheckDataViewExistByID(testCtx, "a")
		// 	So(err, ShouldResemble, expectedHttpErr)

		// })

		Convey("Get succeed", func() {
			expectedName := "test"
			dva.EXPECT().CheckDataViewExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedName, true, nil)

			tx, _ := dvs.db.Begin()
			name, _, err := dvs.CheckDataViewExistByID(testCtx, tx, "a")
			So(name, ShouldEqual, expectedName)
			So(err, ShouldBeNil)
		})
	})
}

func Test_DataViewService_MergeIndexBaseFields(t *testing.T) {
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

func Test_DataViewService_DeepEqualJobCondition(t *testing.T) {
	Convey("Test deepEqualJobCondition", t, func() {
		Convey("a is nil && b is nil", func() {
			res := deepEqualJobCondition(nil, nil)
			So(res, ShouldBeTrue)
		})

		Convey("a is nil and b is not nil", func() {
			b := &interfaces.CondCfg{}
			res := deepEqualJobCondition(nil, b)
			So(res, ShouldBeFalse)
		})

		Convey("a is not nil and b is nil", func() {
			a := &interfaces.CondCfg{
				Name:      "xx",
				Operation: dcond.OperationExist,
			}
			res := deepEqualJobCondition(a, nil)
			So(res, ShouldBeFalse)
		})

		Convey("name is not equal", func() {
			a := &interfaces.CondCfg{
				Name:      "x",
				Operation: dcond.OperationExist,
			}

			b := &interfaces.CondCfg{
				Name:      "y",
				Operation: dcond.OperationExist,
			}
			res := deepEqualJobCondition(a, b)
			So(res, ShouldBeFalse)
		})

		Convey("operation is not equal", func() {
			a := &interfaces.CondCfg{
				Name:      "x",
				Operation: dcond.OperationExist,
			}

			b := &interfaces.CondCfg{
				Name:      "x",
				Operation: dcond.OperationNotIn,
			}
			res := deepEqualJobCondition(a, b)
			So(res, ShouldBeFalse)
		})

		Convey("value from is not equal", func() {
			a := &interfaces.CondCfg{
				Name:      "x",
				Operation: dcond.OperationEq,
				ValueOptCfg: interfaces.ValueOptCfg{
					ValueFrom: "const",
				},
			}

			b := &interfaces.CondCfg{
				Name:      "x",
				Operation: dcond.OperationEq,
				ValueOptCfg: interfaces.ValueOptCfg{
					ValueFrom: "field",
				},
			}
			res := deepEqualJobCondition(a, b)
			So(res, ShouldBeFalse)
		})

		Convey("value is not equal", func() {
			a := &interfaces.CondCfg{
				Name:      "x",
				Operation: dcond.OperationEq,
				ValueOptCfg: interfaces.ValueOptCfg{
					ValueFrom: "const",
					Value:     "melody",
				},
			}

			b := &interfaces.CondCfg{
				Name:      "x",
				Operation: dcond.OperationEq,
				ValueOptCfg: interfaces.ValueOptCfg{
					ValueFrom: "const",
					Value:     "annie",
				},
			}
			res := deepEqualJobCondition(a, b)
			So(res, ShouldBeFalse)
		})

		Convey("sub conditions length is not equal", func() {
			a := &interfaces.CondCfg{
				Operation: dcond.OperationAnd,
				SubConds:  []*interfaces.CondCfg{},
			}

			b := &interfaces.CondCfg{
				Operation: dcond.OperationAnd,
				SubConds: []*interfaces.CondCfg{
					{Name: "x", Operation: dcond.OperationExist},
				},
			}
			res := deepEqualJobCondition(a, b)
			So(res, ShouldBeFalse)
		})

		Convey("sub conditions are not equal", func() {
			a := &interfaces.CondCfg{
				Operation: dcond.OperationAnd,
				SubConds: []*interfaces.CondCfg{
					{Name: "x", Operation: dcond.OperationExist},
					{Name: "y", Operation: dcond.OperationExist},
					{Name: "z", Operation: dcond.OperationExist},
				},
			}

			b := &interfaces.CondCfg{
				Operation: dcond.OperationAnd,
				SubConds: []*interfaces.CondCfg{
					{Name: "x", Operation: dcond.OperationExist},
					{Name: "y", Operation: dcond.OperationExist},
					{Name: "z", Operation: dcond.OperationNotExist},
				},
			}
			res := deepEqualJobCondition(a, b)
			So(res, ShouldBeFalse)
		})

		Convey("condition struct is not equal", func() {
			a := &interfaces.CondCfg{
				Name:      "x",
				Operation: dcond.OperationEq,
				ValueOptCfg: interfaces.ValueOptCfg{
					ValueFrom: "const",
					Value:     "melody",
				},
			}

			b := &interfaces.CondCfg{
				Operation: dcond.OperationAnd,
				SubConds: []*interfaces.CondCfg{
					{Name: "x", Operation: dcond.OperationExist},
				},
			}
			res := deepEqualJobCondition(a, b)
			So(res, ShouldBeFalse)
		})
	})
}

// func Test_DataViewService_ValidateCond(t *testing.T) {
// 	Convey("Test validateCond", t, func() {

// 		Convey("cfg is nil", func() {
// 			err := validateCond(testCtx, nil, nil)
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("filter field is not in view fields list", func() {
// 			fieldsMap := map[string]string{
// 				"a": dtype.DataType_Text,
// 				"b": dtype.DataType_Long,
// 			}

// 			cfg := &interfaces.CondCfg{
// 				Name:      "c",
// 				Operation: dcond.OperationEq,
// 			}

// 			err := validateCond(testCtx, cfg, fieldsMap)
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("filter field is binary type", func() {
// 			fieldsMap := map[string]string{
// 				"a": dtype.DataType_Binary,
// 			}

// 			cfg := &interfaces.CondCfg{
// 				Operation: dcond.OperationAnd,
// 				SubConds: []*interfaces.CondCfg{
// 					{Name: "a", Operation: dcond.OperationEq},
// 					{Name: "b", Operation: dcond.OperationEq},
// 				},
// 			}

// 			err := validateCond(testCtx, cfg, fieldsMap)
// 			So(err, ShouldNotBeNil)
// 		})

// 		Convey("filter field is empty or not_empty, but field type is not string", func() {
// 			fieldsMap := map[string]string{
// 				"a": dtype.DataType_Text,
// 				"b": dtype.DataType_Long,
// 			}

// 			cfg := &interfaces.CondCfg{
// 				Name:      "b",
// 				Operation: dcond.OperationEmpty,
// 			}

// 			err := validateCond(testCtx, cfg, fieldsMap)
// 			So(err, ShouldNotBeNil)
// 		})
// 	})
// }
