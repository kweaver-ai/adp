// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_dict

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	dmock "data-model/interfaces/mock"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

func MockNewDataDictService(appSetting *common.AppSetting,
	dda interfaces.DataDictAccess,
	ddis interfaces.DataDictItemsService,
	ps interfaces.PermissionService) (*dataDictService, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	dds := &dataDictService{
		appSetting: appSetting,
		dda:        dda,
		ddis:       ddis,
		db:         db,
		ps:         ps,
	}
	return dds, smock
}

func Test_DataDictService_ListDataDicts(t *testing.T) {
	Convey("test DictService ListDataDicts\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ddis := dmock.NewMockDataDictItemsService(mockCtrl)
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		dds, _ := MockNewDataDictService(appSetting, dda, ddis, ps)

		dictQuery := interfaces.DataDictQueryParams{
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit:     interfaces.MAX_LIMIT,
				Offset:    interfaces.MIN_OFFSET,
				Sort:      interfaces.DATA_DICT_SORT["update_time"],
				Direction: interfaces.DESC_DIRECTION,
			},
		}
		resrc := map[string]interfaces.ResourceOps{
			"1": {
				ResourceID: "1",
			},
		}

		Convey("Success ListDataDicts\n", func() {
			dictQuery.NamePattern = "test"
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			dda.EXPECT().ListDataDicts(gomock.Any(), gomock.Any()).Return([]interfaces.DataDict{{DictID: "1"}}, nil)
			// dda.EXPECT().GetDictTotal(gomock.Any(), gomock.Any()).Return(int64(1), nil)
			_, total, err := dds.ListDataDicts(testCtx, dictQuery)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
		})

		Convey("ListDataDicts Failed\n", func() {
			dda.EXPECT().ListDataDicts(gomock.Any(), gomock.Any()).Return([]interfaces.DataDict{}, fmt.Errorf("ListDataDicts Failed"))
			_, _, err := dds.ListDataDicts(testCtx, dictQuery)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusNotFound)
		})

		// Convey("GetDictTotal Failed\n", func() {
		// 	dda.EXPECT().ListDataDicts(gomock.Any(), gomock.Any()).Return([]interfaces.DataDict{}, nil)
		// 	dda.EXPECT().GetDictTotal(gomock.Any(), gomock.Any()).Return(int64(0), fmt.Errorf("GetDictTotal Failed"))
		// 	_, _, err := dds.ListDataDicts(testCtx, dictQuery)
		// 	So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusNotFound)
		// })
	})
}

func Test_DataDictService_GetDataDicts(t *testing.T) {
	Convey("test GetDataDicts\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ddis := dmock.NewMockDataDictItemsService(mockCtrl)
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		dds, _ := MockNewDataDictService(appSetting, dda, ddis, ps)

		dictIDs := []string{"431687149096534788"}
		dictInfo := interfaces.DataDict{
			DictID:    "431687149096534788",
			DictName:  "test",
			DictType:  "kv_dict",
			UniqueKey: true,
			Tags:      []string{"a", "b", "c"},
			Dimension: interfaces.DATA_DICT_KV_DIMENSION,
			DictStore: "t_data_dict_item",
			Comment:   "",
		}
		resrc := map[string]interfaces.ResourceOps{
			"431687149096534788": {
				ResourceID: "431687149096534788",
			},
		}

		Convey("GetDataDicts GetKVDictItems Success \n", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			dda.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(dictInfo, nil)
			ddis.EXPECT().GetKVDictItems(gomock.Any(), gomock.Any()).Return([]map[string]string{}, nil)

			dictInfos, err := dds.GetDataDicts(testCtx, dictIDs)
			So(err, ShouldBeNil)
			So(dictInfos, ShouldNotResemble, []interfaces.DataDict{})
		})

		Convey("GetDataDicts GetDimensionDictItems Success \n", func() {
			dictInfo.DictType = interfaces.DATA_DICT_TYPE_DIMENSION
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			dda.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(dictInfo, nil)
			ddis.EXPECT().GetDimensionDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]map[string]string{}, nil)

			dictInfos, err := dds.GetDataDicts(testCtx, dictIDs)
			So(err, ShouldBeNil)
			So(dictInfos, ShouldNotResemble, []interfaces.DataDict{})
		})

		Convey("GetDataDicts GetDataDictByID Error\n", func() {

			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			dda.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(dictInfo, fmt.Errorf("GetDataDictByID Failed"))

			_, err := dds.GetDataDicts(testCtx, dictIDs)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusNotFound)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_DictNotFound)

		})

		Convey("GetDataDicts GetKVDictItems Error \n", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			dda.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(dictInfo, nil)
			ddis.EXPECT().GetKVDictItems(gomock.Any(), gomock.Any()).Return([]map[string]string{}, fmt.Errorf("GetKVDictItems Failed"))

			_, err := dds.GetDataDicts(testCtx, dictIDs)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusNotFound)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_DictItemNotFound)
		})

		Convey("GetDataDicts GetDimensionDictItems Error \n", func() {
			dictInfo.DictType = interfaces.DATA_DICT_TYPE_DIMENSION
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			dda.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(dictInfo, nil)
			ddis.EXPECT().GetDimensionDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]map[string]string{}, fmt.Errorf("GetDimensionDictItems Failed"))

			_, err := dds.GetDataDicts(testCtx, dictIDs)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusNotFound)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_DictItemNotFound)
		})
	})

}

func Test_DataDictService_CreateDataDict(t *testing.T) {

	Convey("test CreateDataDict\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ddis := dmock.NewMockDataDictItemsService(mockCtrl)
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		dds, smock := MockNewDataDictService(appSetting, dda, ddis, ps)
		items := []map[string]string{
			{"comment": "comment", "key": "key", "value": "value"},
			{"comment": "comment0", "key": "key0", "value": "value0"},
		}
		dictInfo := interfaces.DataDict{
			DictName:  "test",
			DictType:  interfaces.DATA_DICT_TYPE_KV,
			UniqueKey: true,
			DictStore: interfaces.DATA_DICT_STORE_DEFAULT,
			Comment:   "a",
			Dimension: interfaces.DATA_DICT_KV_DIMENSION,
			DictItems: items,
		}
		errors := errors.New("Error")
		Convey("CreateDataDict KV Success\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ps.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			dda.EXPECT().CreateDataDict(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddis.EXPECT().CreateKVDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			smock.ExpectCommit()
			_, err := dds.CreateDataDict(testCtx, dictInfo)
			So(err, ShouldBeNil)
		})

		Convey("CreateDataDict Dimension Success\n", func() {
			dictInfo.DictType = interfaces.DATA_DICT_TYPE_DIMENSION
			dictInfo.Dimension = interfaces.Dimension{
				Keys: []interfaces.DimensionItem{
					{
						ID:   "",
						Name: "key111",
					},
				},
				Values: []interfaces.DimensionItem{
					{
						ID:   "",
						Name: "value111",
					},
				},
				Comment: "",
			}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ps.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dda.EXPECT().CreateDimensionDictStore(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dda.EXPECT().AddDimensionIndex(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			dda.EXPECT().CreateDataDict(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddis.EXPECT().CreateDimensionDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			smock.ExpectCommit()
			_, err := dds.CreateDataDict(testCtx, dictInfo)
			So(err, ShouldBeNil)
		})

		Convey("CreateDataDict Begin Transaction Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin().WillReturnError(errors)
			_, err := dds.CreateDataDict(testCtx, dictInfo)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_InternalError_BeginTransactionFailed)
		})

		Convey("CreateDataDict Commit Transaction Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ps.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			dda.EXPECT().CreateDataDict(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddis.EXPECT().CreateKVDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			smock.ExpectCommit().WillReturnError(errors)
			_, err := dds.CreateDataDict(testCtx, dictInfo)
			So(err, ShouldResemble, errors)
		})

		Convey("CreateDataDict CreateDataDict Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			dda.EXPECT().CreateDataDict(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors)
			smock.ExpectRollback()
			_, err := dds.CreateDataDict(testCtx, dictInfo)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_InternalError)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictService_UpdateDataDict(t *testing.T) {

	Convey("test UpdateDataDict\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ddis := dmock.NewMockDataDictItemsService(mockCtrl)
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		dds, _ := MockNewDataDictService(appSetting, dda, ddis, ps)

		dictInfo := interfaces.DataDict{
			DictName:  "test",
			DictType:  interfaces.DATA_DICT_TYPE_DIMENSION,
			UniqueKey: true,
			DictStore: "t_dim34243244324",
			Comment:   "a",
			Dimension: interfaces.DATA_DICT_KV_DIMENSION,
		}
		errors := errors.New("Error")
		Convey("UpdateDataDict Success \n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ps.EXPECT().UpdateResource(gomock.Any(), gomock.Any()).Return(nil)
			dda.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.DataDict{UniqueKey: true}, nil)
			dda.EXPECT().CheckDictExistByName(gomock.Any(), gomock.Any()).AnyTimes().Return(false, nil)
			ddis.EXPECT().UpdateDimension(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.DimensionItem{}, true, nil)
			dda.EXPECT().DropDimensionIndex(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			dda.EXPECT().AddDimensionIndex(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			dda.EXPECT().UpdateDataDict(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			err := dds.UpdateDataDict(testCtx, dictInfo)
			So(err, ShouldBeNil)
		})

		Convey("UpdateDataDict GetDataDictByID Failed \n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dda.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.DataDict{DictType: interfaces.DATA_DICT_TYPE_DIMENSION}, errors)

			err := dds.UpdateDataDict(testCtx, dictInfo)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusNotFound)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_DictNotFound)
		})

		Convey("UpdateDataDict CheckDictExistByName exist \n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dda.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.DataDict{DictType: interfaces.DATA_DICT_TYPE_DIMENSION}, nil)
			dda.EXPECT().CheckDictExistByName(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

			err := dds.UpdateDataDict(testCtx, dictInfo)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusForbidden)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_DictNameExisted)
		})

		Convey("UpdateDataDict CheckDictExistByName Failed \n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dda.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.DataDict{DictType: interfaces.DATA_DICT_TYPE_DIMENSION}, nil)
			dda.EXPECT().CheckDictExistByName(gomock.Any(), gomock.Any()).AnyTimes().Return(false, errors)

			err := dds.UpdateDataDict(testCtx, dictInfo)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_InternalError)
		})

		Convey("UpdateDataDict UpdateDimension Failed \n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dda.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.DataDict{DictType: interfaces.DATA_DICT_TYPE_DIMENSION}, nil)
			dda.EXPECT().CheckDictExistByName(gomock.Any(), gomock.Any()).AnyTimes().Return(false, nil)
			ddis.EXPECT().UpdateDimension(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.DimensionItem{}, true, errors)

			err := dds.UpdateDataDict(testCtx, dictInfo)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusBadRequest)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_InvalidParameter_DictDimension)
		})

		Convey("UpdateDataDict DropDimensionIndex Failded\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dda.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.DataDict{UniqueKey: true, DictType: interfaces.DATA_DICT_TYPE_DIMENSION}, nil)
			dda.EXPECT().CheckDictExistByName(gomock.Any(), gomock.Any()).AnyTimes().Return(false, nil)
			ddis.EXPECT().UpdateDimension(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.DimensionItem{}, true, nil)
			dda.EXPECT().DropDimensionIndex(gomock.Any(), gomock.Any()).AnyTimes().Return(errors)

			err := dds.UpdateDataDict(testCtx, dictInfo)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_InternalError)
		})

		Convey("UpdateDataDict AddDimensionIndex Failded \n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dda.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.DataDict{UniqueKey: true, DictType: interfaces.DATA_DICT_TYPE_DIMENSION}, nil)
			dda.EXPECT().CheckDictExistByName(gomock.Any(), gomock.Any()).AnyTimes().Return(false, nil)
			ddis.EXPECT().UpdateDimension(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.DimensionItem{}, true, nil)
			dda.EXPECT().DropDimensionIndex(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			dda.EXPECT().AddDimensionIndex(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors)

			err := dds.UpdateDataDict(testCtx, dictInfo)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_InternalError)
		})

		Convey("UpdateDataDict UpdateDataDict Failed \n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dda.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).AnyTimes().Return(interfaces.DataDict{DictType: interfaces.DATA_DICT_TYPE_DIMENSION}, nil)
			dda.EXPECT().CheckDictExistByName(gomock.Any(), gomock.Any()).AnyTimes().Return(false, nil)
			ddis.EXPECT().UpdateDimension(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]interfaces.DimensionItem{}, true, nil)
			dda.EXPECT().DropDimensionIndex(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			dda.EXPECT().AddDimensionIndex(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			dda.EXPECT().UpdateDataDict(gomock.Any(), gomock.Any()).AnyTimes().Return(errors)

			err := dds.UpdateDataDict(testCtx, dictInfo)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_InternalError)
		})
	})
}

func Test_DataDictService_DeleteDataDict(t *testing.T) {
	Convey("test DeleteDataDict\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ddis := dmock.NewMockDataDictItemsService(mockCtrl)
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		dds, _ := MockNewDataDictService(appSetting, dda, ddis, ps)
		errors := errors.New("Error")
		resrc := map[string]interfaces.ResourceOps{
			"1": {
				ResourceID: "1",
			},
		}
		Convey("DeleteDataDict Success rowsAffect 1 \n", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			ps.EXPECT().DeleteResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dda.EXPECT().DeleteDataDict(gomock.Any(), gomock.Any()).AnyTimes().Return(int64(1), nil)
			ddis.EXPECT().DeleteDataDictItems(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			rowsAffect, err := dds.DeleteDataDict(testCtx, interfaces.DataDict{DictID: "1", DictStore: "dsaad"})
			So(err, ShouldBeNil)
			So(rowsAffect, ShouldAlmostEqual, int64(1))
		})

		Convey("DeleteDataDict Failed \n", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			dda.EXPECT().DeleteDataDict(gomock.Any(), gomock.Any()).AnyTimes().Return(int64(0), errors)

			_, err := dds.DeleteDataDict(testCtx, interfaces.DataDict{DictID: "1", DictStore: "dsaad"})
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_InternalError)
		})

		Convey("DeleteDataDict Failed rowsAffect 2 > 1 \n", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			dda.EXPECT().DeleteDataDict(gomock.Any(), gomock.Any()).AnyTimes().Return(int64(2), nil)

			_, err := dds.DeleteDataDict(testCtx, interfaces.DataDict{DictID: "1", DictStore: "dsaad"})
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_InternalError)
		})

		Convey("DeleteDataDict Failed DeleteDataDictItems error \n", func() {
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			dda.EXPECT().DeleteDataDict(gomock.Any(), gomock.Any()).AnyTimes().Return(int64(1), nil)
			ddis.EXPECT().DeleteDataDictItems(gomock.Any(), gomock.Any()).AnyTimes().Return(errors)

			_, err := dds.DeleteDataDict(testCtx, interfaces.DataDict{DictID: "1", DictStore: "dsaad"})
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_InternalError)

		})

	})
}

func Test_DataDictService_GetDataDictByID(t *testing.T) {
	Convey("test DictService GetDataDictByID\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ddis := dmock.NewMockDataDictItemsService(mockCtrl)
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		dds, _ := MockNewDataDictService(appSetting, dda, ddis, ps)
		errors := errors.New("Error")

		Convey("Success GetDataDictByID\n", func() {
			dda.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{}, nil)
			_, err := dds.GetDataDictByID(testCtx, "1")
			So(err, ShouldBeNil)
		})

		Convey("Failed GetDataDictByID\n", func() {
			dda.EXPECT().GetDataDictByID(gomock.Any(), gomock.Any()).Return(interfaces.DataDict{}, errors)
			_, err := dds.GetDataDictByID(testCtx, "1")
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusNotFound)
		})
	})
}

func Test_DataDictService_CheckDictExistByName(t *testing.T) {
	Convey("test DictService CheckDictExistByName\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ddis := dmock.NewMockDataDictItemsService(mockCtrl)
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		dds, _ := MockNewDataDictService(appSetting, dda, ddis, ps)
		errors := errors.New("Error")

		Convey("Success CheckDictExistByName\n", func() {
			dda.EXPECT().CheckDictExistByName(gomock.Any(), gomock.Any()).Return(false, nil)
			_, err := dds.CheckDictExistByName(testCtx, "1")
			So(err, ShouldBeNil)
		})

		Convey("Failed CheckDictExistByName\n", func() {
			dda.EXPECT().CheckDictExistByName(gomock.Any(), gomock.Any()).Return(true, errors)
			_, err := dds.CheckDictExistByName(testCtx, "1")
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusForbidden)
		})
	})
}
