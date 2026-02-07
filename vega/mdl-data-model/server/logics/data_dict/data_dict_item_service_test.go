// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_dict

import (
	"database/sql"
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
	dmock "data-model/interfaces/mock"
)

func MockNewDataDictItemService(appSetting *common.AppSetting,
	ddia interfaces.DataDictItemAccess,
	dda interfaces.DataDictAccess,
	ps interfaces.PermissionService) (*dataDictItemService, sqlmock.Sqlmock) {

	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	ddis := &dataDictItemService{
		appSetting: appSetting,
		dda:        dda,
		ddia:       ddia,
		db:         db,
		ps:         ps,
	}
	return ddis, smock
}

func Test_DataDictItemService_GetKVDictItems(t *testing.T) {
	Convey("test DictItemService GetKVDictItems\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ddia := dmock.NewMockDataDictItemAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ddis, _ := MockNewDataDictItemService(appSetting, ddia, dda, ps)
		errors := errors.New("error")

		Convey("Success GetKVDictItems\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().GetKVDictItems(gomock.Any(), gomock.Any()).Return([]map[string]string{}, nil)

			_, err := ddis.GetKVDictItems(testCtx, "1")
			So(err, ShouldBeNil)
		})

		Convey("GetKVDictItems Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().GetKVDictItems(gomock.Any(), gomock.Any()).Return([]map[string]string{}, errors)

			_, err := ddis.GetKVDictItems(testCtx, "1")
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictItemService_GetDimensionDictItems(t *testing.T) {
	Convey("test DictItemService GetDimensionDictItems\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ddia := dmock.NewMockDataDictItemAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ddis, _ := MockNewDataDictItemService(appSetting, ddia, dda, ps)
		errors := errors.New("error")

		dimension := interfaces.Dimension{
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

		Convey("Success GetDimensionDictItems\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().GetDimensionDictItems(gomock.Any(), gomock.Any(), gomock.Any()).Return([]map[string]string{}, nil)

			_, err := ddis.GetDimensionDictItems(testCtx, "1", "1", dimension)
			So(err, ShouldBeNil)
		})

		Convey("GetDimensionDictItems Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().GetDimensionDictItems(gomock.Any(), gomock.Any(), gomock.Any()).Return([]map[string]string{}, errors)

			_, err := ddis.GetDimensionDictItems(testCtx, "1", "1", dimension)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictItemService_UpdateDimension(t *testing.T) {
	Convey("test DictItemService UpdateDimension\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ddia := dmock.NewMockDataDictItemAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ddis, _ := MockNewDataDictItemService(appSetting, ddia, dda, ps)
		errors := errors.New("error")

		new := []interfaces.DimensionItem{
			{
				ID:   "",
				Name: "key111",
			},
		}
		old := []interfaces.DimensionItem{
			{
				ID:   "",
				Name: "value111",
			},
		}

		Convey("Success UpdateDimension\n", func() {
			ddia.EXPECT().AddDimensionColumn(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().DropDimensionColumn(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			_, _, err := ddis.UpdateDimension(testCtx, "1", "prefix", new, old)
			So(err, ShouldBeNil)
		})

		Convey("UpdateDimension AddDimensionColumn Failed\n", func() {
			ddia.EXPECT().AddDimensionColumn(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors)

			_, _, err := ddis.UpdateDimension(testCtx, "1", "prefix", new, old)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})

		Convey("UpdateDimension DropDimensionColumn Failed\n", func() {
			ddia.EXPECT().AddDimensionColumn(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().DropDimensionColumn(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors)

			_, _, err := ddis.UpdateDimension(testCtx, "1", "prefix", new, old)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictItemService_DeleteDataDictItems(t *testing.T) {
	Convey("test DictItemService DeleteDataDictItems\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ddia := dmock.NewMockDataDictItemAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ddis, _ := MockNewDataDictItemService(appSetting, ddia, dda, ps)
		errors := errors.New("error")

		Convey("Success DeleteDataDictItems\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().DeleteDataDictItems(gomock.Any(), gomock.Any()).Return(nil)

			err := ddis.DeleteDataDictItems(testCtx, interfaces.DataDict{
				DictID:    "1",
				DictType:  interfaces.DATA_DICT_TYPE_KV,
				DictStore: interfaces.DATA_DICT_STORE_DEFAULT,
			})
			So(err, ShouldBeNil)
		})

		Convey("DeleteDataDictItems Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().DeleteDataDictItems(gomock.Any(), gomock.Any()).Return(errors)

			err := ddis.DeleteDataDictItems(testCtx, interfaces.DataDict{
				DictID:    "1",
				DictType:  interfaces.DATA_DICT_TYPE_KV,
				DictStore: interfaces.DATA_DICT_STORE_DEFAULT,
			})
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})

		Convey("DeleteDataDictItems DeleteDimensionTable Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().DeleteDimensionTable(gomock.Any(), gomock.Any()).Return(errors)

			err := ddis.DeleteDataDictItems(testCtx,
				interfaces.DataDict{
					DictID:    "1",
					DictType:  interfaces.DATA_DICT_TYPE_DIMENSION,
					DictStore: "store",
				})
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictItemService_ListDataDictItems(t *testing.T) {
	Convey("test DictItemService ListDataDictItems\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ddia := dmock.NewMockDataDictItemAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ddis, _ := MockNewDataDictItemService(appSetting, ddia, dda, ps)
		errors := errors.New("error")

		dictItemQuery := interfaces.DataDictItemQueryParams{
			Patterns: []interfaces.DataDictItemQueryPattern{
				{
					QueryField:   "key",
					QueryPattern: "",
				},
				{
					QueryField:   "value",
					QueryPattern: "",
				},
			},
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit:     interfaces.MAX_LIMIT,
				Offset:    interfaces.MIN_OFFSET,
				Sort:      interfaces.DATA_DICT_ITEM_SORT["id"],
				Direction: interfaces.DESC_DIRECTION,
			},
		}

		Convey("Success ListDataDictItems\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().ListDataDictItems(gomock.Any(), gomock.Any(), gomock.Any()).Return([]map[string]string{}, nil)
			ddia.EXPECT().GetDictItemTotal(gomock.Any(), gomock.Any(), gomock.Any()).Return(1, nil)

			_, total, err := ddis.ListDataDictItems(testCtx, interfaces.DataDict{DictID: "1"}, dictItemQuery)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
		})

		Convey("ListDataDictItems ListDataDictItems Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().ListDataDictItems(gomock.Any(), gomock.Any(), gomock.Any()).Return([]map[string]string{}, errors)

			_, _, err := ddis.ListDataDictItems(testCtx, interfaces.DataDict{DictID: "1"}, dictItemQuery)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})

		Convey("ListDataDictItems GetDictItemTotal Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().ListDataDictItems(gomock.Any(), gomock.Any(), gomock.Any()).Return([]map[string]string{}, nil)
			ddia.EXPECT().GetDictItemTotal(gomock.Any(), gomock.Any(), gomock.Any()).Return(1, errors)

			_, _, err := ddis.ListDataDictItems(testCtx, interfaces.DataDict{DictID: "1"}, dictItemQuery)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictItemService_CreateDataDictItem(t *testing.T) {

	Convey("test CreateDataDictItem\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ddia := dmock.NewMockDataDictItemAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ddis, _ := MockNewDataDictItemService(appSetting, ddia, dda, ps)
		errors := errors.New("error")

		dimension := interfaces.Dimension{
			Keys: []interfaces.DimensionItem{
				{
					ID:    "",
					Name:  "key111",
					Value: "dsada",
				},
			},
			Values: []interfaces.DimensionItem{
				{
					ID:    "",
					Name:  "value111",
					Value: "dsadadsa",
				},
			},
			Comment: "",
		}

		Convey("CreateDataDictItem Success\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(0, nil)
			ddia.EXPECT().CreateDataDictItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			dda.EXPECT().UpdateDictUpdateTime(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			_, err := ddis.CreateDataDictItem(testCtx, interfaces.DataDict{DictID: "1", UniqueKey: true}, dimension)
			So(err, ShouldBeNil)
		})

		Convey("CreateDataDictItem CountDictItemByKey Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(0, errors)

			_, err := ddis.CreateDataDictItem(testCtx, interfaces.DataDict{DictID: "1", UniqueKey: true}, dimension)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})

		Convey("CreateDataDictItem CountDictItemByKey exist\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(1, nil)

			_, err := ddis.CreateDataDictItem(testCtx, interfaces.DataDict{DictID: "1", UniqueKey: true}, dimension)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusForbidden)
		})

		Convey("CreateDataDictItem Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(0, nil)
			ddia.EXPECT().CreateDataDictItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors)

			_, err := ddis.CreateDataDictItem(testCtx, interfaces.DataDict{DictID: "1", UniqueKey: true}, dimension)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})

		Convey("CreateDataDictItem UpdateDictUpdateTime Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(0, nil)
			ddia.EXPECT().CreateDataDictItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			dda.EXPECT().UpdateDictUpdateTime(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors)

			_, err := ddis.CreateDataDictItem(testCtx, interfaces.DataDict{DictID: "1", UniqueKey: true}, dimension)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictItemService_ImportDataDictItems(t *testing.T) {

	Convey("test ImportDataDictItems\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ddia := dmock.NewMockDataDictItemAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ddis, smock := MockNewDataDictItemService(appSetting, ddia, dda, ps)

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
		errors := errors.New("error")

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

		dimensionItems := []map[string]string{
			{"comment": "comment", "department": "key0", "group": "key1", "name": "value0"},
		}
		dimensionDict := interfaces.DataDict{
			DictName:  "test",
			DictType:  interfaces.DATA_DICT_TYPE_DIMENSION,
			UniqueKey: true,
			DictStore: interfaces.DATA_DICT_DIMENSION_PREFIX_TABLE + "527800737703436444",
			Comment:   "a",
			Dimension: DIMENSION,
			DictItems: dimensionItems,
		}

		Convey("ImportDataDictItems KV Success\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(0, nil)
			ddia.EXPECT().CreateKVDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dda.EXPECT().UpdateDictUpdateTime(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			err := ddis.ImportDataDictItems(testCtx, &dictInfo, items, interfaces.ImportMode_Normal)
			So(err, ShouldBeNil)
		})

		Convey("ImportDataDictItems Begin Transaction Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin().WillReturnError(errors)

			err := ddis.ImportDataDictItems(testCtx, &dictInfo, items, interfaces.ImportMode_Normal)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_InternalError_BeginTransactionFailed)
		})

		Convey("ImportDataDictItems Commit Transaction Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(0, nil)
			ddia.EXPECT().CreateKVDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dda.EXPECT().UpdateDictUpdateTime(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit().WillReturnError(errors)

			err := ddis.ImportDataDictItems(testCtx, &dictInfo, items, interfaces.ImportMode_Normal)
			So(err, ShouldResemble, errors)
		})

		Convey("ImportDataDictItems UpdateDictUpdateTime Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(0, nil)
			ddia.EXPECT().CreateKVDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dda.EXPECT().UpdateDictUpdateTime(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors)
			smock.ExpectRollback()

			err := ddis.ImportDataDictItems(testCtx, &dictInfo, items, interfaces.ImportMode_Normal)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_InternalError)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})

		Convey("ImportDataDictItems dictItemAccess CreateKVDictItems Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(0, nil)
			ddia.EXPECT().CreateKVDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors)
			smock.ExpectRollback()

			err := ddis.ImportDataDictItems(testCtx, &dictInfo, items, interfaces.ImportMode_Normal)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_InternalError)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})

		Convey("ImportDataDictItems CountDictItemByKey Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(0, errors)
			smock.ExpectRollback()

			err := ddis.ImportDataDictItems(testCtx, &dictInfo, items, interfaces.ImportMode_Normal)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_InternalError)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})

		Convey("ImportDataDictItems kv dict CountDictItemByKey cnt > 0, normal mode\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(2, nil)
			smock.ExpectRollback()

			err := ddis.ImportDataDictItems(testCtx, &dictInfo, items, interfaces.ImportMode_Normal)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_Duplicated_DictItemKey)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusForbidden)
		})

		Convey("ImportDataDictItems kv dict CountDictItemByKey cnt > 0, ignore mode\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(),
				[]interfaces.DimensionItem{{ID: "item_key", Name: "key", Value: "key"}}).AnyTimes().Return(2, nil)
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(),
				[]interfaces.DimensionItem{{ID: "item_key", Name: "key", Value: "key0"}}).AnyTimes().Return(0, nil)
			ddia.EXPECT().CreateKVDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dda.EXPECT().UpdateDictUpdateTime(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			err := ddis.ImportDataDictItems(testCtx, &dictInfo, items, interfaces.ImportMode_Ignore)
			So(err, ShouldBeNil)
		})

		Convey("ImportDataDictItems kv dict CountDictItemByKey cnt > 0, overwrite mode\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(),
				[]interfaces.DimensionItem{{ID: "item_key", Name: "key", Value: "key"}}).AnyTimes().Return(2, nil)
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(),
				[]interfaces.DimensionItem{{ID: "item_key", Name: "key", Value: "key0"}}).AnyTimes().Return(0, nil)
			ddia.EXPECT().GetDictItemIDByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"1", "2"}, nil)
			ddia.EXPECT().DeleteDataDictItemsByItemIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().CreateKVDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dda.EXPECT().UpdateDictUpdateTime(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			err := ddis.ImportDataDictItems(testCtx, &dictInfo, items, interfaces.ImportMode_Overwrite)
			So(err, ShouldBeNil)
		})

		Convey("ImportDataDictItems dimension dict CountDictItemByKey cnt > 0, normal mode\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(2, nil)
			smock.ExpectRollback()

			err := ddis.ImportDataDictItems(testCtx, &dimensionDict, dimensionItems, interfaces.ImportMode_Normal)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldResemble, derrors.DataModel_DataDict_Duplicated_DictItemKey)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusForbidden)
		})

		Convey("ImportDataDictItems dimension dict CountDictItemByKey cnt > 0, ignore mode\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(1, nil)
			ddia.EXPECT().CreateDimensionDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			dda.EXPECT().UpdateDictUpdateTime(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			err := ddis.ImportDataDictItems(testCtx, &dimensionDict, dimensionItems, interfaces.ImportMode_Ignore)
			So(err, ShouldBeNil)
		})

		Convey("ImportDataDictItems dimension dict CountDictItemByKey cnt > 0, overwrite mode\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(1, nil)
			ddia.EXPECT().GetDictItemIDByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"1"}, nil)
			ddia.EXPECT().DeleteDataDictItemsByItemIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().CreateDimensionDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dda.EXPECT().UpdateDictUpdateTime(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			err := ddis.ImportDataDictItems(testCtx, &dimensionDict, dimensionItems, interfaces.ImportMode_Overwrite)
			So(err, ShouldBeNil)
		})
	})
}

func Test_DataDictItemService_UpdateDataDictItem(t *testing.T) {

	Convey("test UpdateDataDictItem\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ddia := dmock.NewMockDataDictItemAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ddis, _ := MockNewDataDictItemService(appSetting, ddia, dda, ps)
		errors := errors.New("error")

		dimension := interfaces.Dimension{
			Keys: []interfaces.DimensionItem{
				{
					ID:    "",
					Name:  "key111",
					Value: "dsada",
				},
			},
			Values: []interfaces.DimensionItem{
				{
					ID:    "",
					Name:  "value111",
					Value: "dsadadsa",
				},
			},
			Comment: "",
		}

		Convey("UpdateDataDictItem Success\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().GetDictItemByItemID(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(map[string]string{}, nil)
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(0, nil)
			ddia.EXPECT().UpdateDataDictItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			dda.EXPECT().UpdateDictUpdateTime(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			err := ddis.UpdateDataDictItem(testCtx, interfaces.DataDict{DictID: "1", UniqueKey: true}, "1", dimension)
			So(err, ShouldBeNil)
		})

		Convey("UpdateDataDictItem GetDictItemByItemID Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().GetDictItemByItemID(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(map[string]string{}, errors)

			err := ddis.UpdateDataDictItem(testCtx, interfaces.DataDict{DictID: "1", UniqueKey: true}, "1", dimension)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusBadRequest)
		})

		Convey("UpdateDataDictItem CountDictItemByKey Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().GetDictItemByItemID(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(map[string]string{}, nil)
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(2, errors)

			err := ddis.UpdateDataDictItem(testCtx, interfaces.DataDict{DictID: "1", UniqueKey: true}, "1", dimension)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})

		Convey("UpdateDataDictItem Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().GetDictItemByItemID(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(map[string]string{}, nil)
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(0, nil)
			ddia.EXPECT().UpdateDataDictItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors)

			err := ddis.UpdateDataDictItem(testCtx, interfaces.DataDict{DictID: "1", UniqueKey: true}, "1", dimension)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})

		Convey("UpdateDataDictItem UpdateDictUpdateTime Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().GetDictItemByItemID(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(map[string]string{}, nil)
			ddia.EXPECT().CountDictItemByKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(0, nil)
			ddia.EXPECT().UpdateDataDictItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			dda.EXPECT().UpdateDictUpdateTime(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors)

			err := ddis.UpdateDataDictItem(testCtx, interfaces.DataDict{DictID: "1", UniqueKey: true}, "1", dimension)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictItemService_DeleteDataDictItem(t *testing.T) {
	Convey("test DeleteDataDictItem\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ddia := dmock.NewMockDataDictItemAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ddis, _ := MockNewDataDictItemService(appSetting, ddia, dda, ps)
		errors := errors.New("error")

		Convey("DeleteDataDictItem Success\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().DeleteDataDictItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(int64(1), nil)
			dda.EXPECT().UpdateDictUpdateTime(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			err := ddis.DeleteDataDictItem(testCtx, interfaces.DataDict{DictID: "1", DictStore: "store"}, "1")
			So(err, ShouldBeNil)

		})

		Convey("DeleteDataDictItem Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().DeleteDataDictItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(int64(1), errors)

			err := ddis.DeleteDataDictItem(testCtx, interfaces.DataDict{DictID: "1", DictStore: "store"}, "1")
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})

		Convey("DeleteDataDictItem UpdateDictUpdateTime Failed\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ddia.EXPECT().DeleteDataDictItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(int64(1), nil)
			dda.EXPECT().UpdateDictUpdateTime(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors)

			err := ddis.DeleteDataDictItem(testCtx, interfaces.DataDict{DictID: "1", DictStore: "store"}, "1")
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictItemService_GetDictItemsByItemIDs(t *testing.T) {
	Convey("test DictItemService GetDataDictItems\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ddia := dmock.NewMockDataDictItemAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ddis, _ := MockNewDataDictItemService(appSetting, ddia, dda, ps)
		errors := errors.New("error")

		var items = []string{"1", "2", "3"}

		Convey("Success GetDictItemsByItemIDs\n", func() {
			ddia.EXPECT().GetDictItemByItemID(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(map[string]string{"aa": "1"}, nil)

			_, err := ddis.GetDictItemsByItemIDs(testCtx, "", items)
			So(err, ShouldBeNil)
		})

		Convey("GetDictItemsByItemIDs Failed\n", func() {
			ddia.EXPECT().GetDictItemByItemID(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(map[string]string{}, errors)

			_, err := ddis.GetDictItemsByItemIDs(testCtx, "", items)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusNotFound)
		})
	})
}

func Test_DataDictItemService_CreateKVDictItems(t *testing.T) {

	Convey("test CreateKVDictItems\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ddia := dmock.NewMockDataDictItemAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ddis, _ := MockNewDataDictItemService(appSetting, ddia, dda, ps)
		errors := errors.New("error")

		dictItems := []map[string]string{
			{"comment": "comment", "key": "key", "value": "value"},
			{"comment": "comment0", "key": "key0", "value": "value0"},
		}

		Convey("CreateKVDictItems Success\n", func() {
			ddia.EXPECT().CreateKVDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			err := ddis.CreateKVDictItems(testCtx, &sql.Tx{}, "1", dictItems)
			So(err, ShouldBeNil)
		})

		Convey("CreateKVDictItems Failed\n", func() {
			ddia.EXPECT().CreateKVDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors)

			err := ddis.CreateKVDictItems(testCtx, &sql.Tx{}, "1", dictItems)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})
	})
}

func Test_DataDictItemService_CreateDimensionDictItems(t *testing.T) {

	Convey("test CreateDimensionDictItems\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dda := dmock.NewMockDataDictAccess(mockCtrl)
		ddia := dmock.NewMockDataDictItemAccess(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		ddis, _ := MockNewDataDictItemService(appSetting, ddia, dda, ps)
		errors := errors.New("error")

		dictItems := []map[string]string{
			{"comment": "comment", "key": "key", "value": "value"},
			{"comment": "comment0", "key": "key0", "value": "value0"},
		}
		dimension := interfaces.Dimension{
			ItemID: "1",
			Keys: []interfaces.DimensionItem{
				{
					ID:    "",
					Name:  "key111",
					Value: "testkey",
				},
			},
			Values: []interfaces.DimensionItem{
				{
					ID:    "",
					Name:  "value111",
					Value: "testvalue",
				},
			},
			Comment: "testcomment"}
		dictStore := "t_dim254243243"

		Convey("CreateDimensionDictItems Success\n", func() {
			ddia.EXPECT().CreateDimensionDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			err := ddis.CreateDimensionDictItems(testCtx, &sql.Tx{}, "1", dictStore, dimension, dictItems)
			So(err, ShouldBeNil)
		})

		Convey("CreateDimensionDictItems Failed\n", func() {
			ddia.EXPECT().CreateDimensionDictItems(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors)

			err := ddis.CreateDimensionDictItems(testCtx, &sql.Tx{}, "1", dictStore, dimension, dictItems)
			So(err.(*rest.HTTPError).HTTPCode, ShouldResemble, http.StatusInternalServerError)
		})
	})
}
