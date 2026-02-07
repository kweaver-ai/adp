// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_view

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	"data-model/interfaces"
	dmock "data-model/interfaces/mock"
)

func MockNewDataViewGroupService(appSetting *common.AppSetting,
	dvga interfaces.DataViewGroupAccess,
	dva interfaces.DataViewAccess,
	dvs interfaces.DataViewService,
	ps interfaces.PermissionService) (*dataViewGroupService, sqlmock.Sqlmock) {

	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	return &dataViewGroupService{
		appSetting: appSetting,
		dvga:       dvga,
		dva:        dva,
		db:         db,
		dvs:        dvs,
		ps:         ps,
	}, smock
}

func Test_DataViewGroupService_CreateDataViewGroup(t *testing.T) {
	Convey("Test CreateDataViewGroup", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dva := dmock.NewMockDataViewAccess(mockCtrl)
		dvga := dmock.NewMockDataViewGroupAccess(mockCtrl)
		psMock := dmock.NewMockPermissionService(mockCtrl)
		dvsMock := dmock.NewMockDataViewService(mockCtrl)
		dvgs, smock := MockNewDataViewGroupService(appSetting, dvga, dva, dvsMock, psMock)

		Convey("Create data view group failed", func() {
			groupTest := &interfaces.DataViewGroup{
				GroupName: "test",
			}

			smock.ExpectBegin()
			dvga.EXPECT().CreateDataViewGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some errors"))
			smock.ExpectRollback()

			tx, _ := dvgs.db.Begin()
			_, httpErr := dvgs.CreateDataViewGroup(testCtx, tx, groupTest)
			So(httpErr, ShouldNotBeNil)
		})

		Convey("Create data view group succeed", func() {
			groupTest := &interfaces.DataViewGroup{
				GroupName: "test",
			}

			smock.ExpectBegin()
			dvga.EXPECT().CreateDataViewGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			tx, _ := dvgs.db.Begin()
			_, httpErr := dvgs.CreateDataViewGroup(testCtx, tx, groupTest)
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_DataViewGroupService_DeleteDataViewGroup(t *testing.T) {
	Convey("Test DeleteDataViewGroup", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dva := dmock.NewMockDataViewAccess(mockCtrl)
		dvga := dmock.NewMockDataViewGroupAccess(mockCtrl)
		psMock := dmock.NewMockPermissionService(mockCtrl)
		dvsMock := dmock.NewMockDataViewService(mockCtrl)

		dvgs, smock := MockNewDataViewGroupService(appSetting, dvga, dva, dvsMock, psMock)

		dataViews := []*interfaces.SimpleDataView{
			{ViewID: "1a"},
			{ViewID: "2b"},
		}

		resrcs := map[string]interfaces.ResourceOps{
			"1a": {ResourceID: "1a"},
			"2b": {ResourceID: "2b"},
		}

		Convey("Delete failed, caused by the error from driven method GetSimpleDataViewSByGroupID", func() {
			expectedErr := errors.New("some errors")
			dva.EXPECT().GetSimpleDataViewsByGroupID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			groupID := "x"
			includeDataViews := false
			_, httpErr := dvgs.DeleteDataViewGroup(testCtx, groupID, includeDataViews)

			So(httpErr, ShouldNotBeNil)
		})

		Convey("Delete failed, caused by there are data views in the group and includeDataViews is false", func() {
			dva.EXPECT().GetSimpleDataViewsByGroupID(gomock.Any(), gomock.Any(), gomock.Any()).Return(dataViews, nil)

			groupID := "x"
			includeDataViews := false
			_, httpErr := dvgs.DeleteDataViewGroup(testCtx, groupID, includeDataViews)
			So(httpErr, ShouldNotBeNil)
		})

		Convey("Delete failed, caused by the error begin transaction", func() {
			dva.EXPECT().GetSimpleDataViewsByGroupID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
			smock.ExpectBegin().WillReturnError(errors.New("some errors"))

			groupID := "x"
			includeDataViews := false
			_, httpErr := dvgs.DeleteDataViewGroup(testCtx, groupID, includeDataViews)
			So(httpErr, ShouldNotBeNil)
		})

		Convey("Delete failed, caused by the error from driven method DeleteDataViewGroup", func() {
			dva.EXPECT().GetSimpleDataViewsByGroupID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
			smock.ExpectBegin()
			dvga.EXPECT().DeleteDataViewGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some errors"))
			smock.ExpectRollback()

			groupID := "x"
			includeDataViews := false
			_, httpErr := dvgs.DeleteDataViewGroup(testCtx, groupID, includeDataViews)
			So(httpErr, ShouldNotBeNil)
		})

		Convey("Delete succeed, there are data views in the group and includeDataViews is true", func() {
			psMock.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(resrcs, nil)
			dva.EXPECT().GetSimpleDataViewsByGroupID(gomock.Any(), gomock.Any(), gomock.Any()).Return(dataViews, nil)

			smock.ExpectBegin()
			dvga.EXPECT().DeleteDataViewGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dva.EXPECT().DeleteDataViews(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			groupID := "x"
			includeDataViews := true
			_, httpErr := dvgs.DeleteDataViewGroup(testCtx, groupID, includeDataViews)

			So(httpErr, ShouldBeNil)
		})

		Convey("Delete succeed, there are no data views in the group and includeDataViews is false", func() {
			dva.EXPECT().GetSimpleDataViewsByGroupID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

			smock.ExpectBegin()
			dvga.EXPECT().DeleteDataViewGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			groupID := "x"
			includeDataViews := false
			_, httpErr := dvgs.DeleteDataViewGroup(testCtx, groupID, includeDataViews)

			So(httpErr, ShouldBeNil)
		})

		Convey("Delete succeed, there are no data views in the group and includeDataViews is true", func() {
			dva.EXPECT().GetSimpleDataViewsByGroupID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

			smock.ExpectBegin()
			dvga.EXPECT().DeleteDataViewGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			groupID := "x"
			includeDataViews := true
			_, httpErr := dvgs.DeleteDataViewGroup(testCtx, groupID, includeDataViews)

			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_DataViewGroupService_UpdateDataViewGroup(t *testing.T) {
	Convey("Test UpdateDataViewGroup", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dva := dmock.NewMockDataViewAccess(mockCtrl)
		dvga := dmock.NewMockDataViewGroupAccess(mockCtrl)
		psMock := dmock.NewMockPermissionService(mockCtrl)
		dvsMock := dmock.NewMockDataViewService(mockCtrl)

		dvgs, _ := MockNewDataViewGroupService(appSetting, dvga, dva, dvsMock, psMock)

		Convey("Update failed, caused by the error from driven method UpdateDataViewGroup", func() {
			expectedErr := errors.New("some errors")
			dvga.EXPECT().UpdateDataViewGroup(gomock.Any(), gomock.Any()).Return(expectedErr)

			groupTest := &interfaces.DataViewGroup{
				GroupID:   "x",
				GroupName: "test",
			}

			httpErr := dvgs.UpdateDataViewGroup(testCtx, groupTest)
			So(httpErr, ShouldNotBeNil)
		})

		Convey("Update succeed", func() {
			dvga.EXPECT().UpdateDataViewGroup(gomock.Any(), gomock.Any()).Return(nil)

			groupTest := &interfaces.DataViewGroup{
				GroupID:   "x",
				GroupName: "test",
			}

			httpErr := dvgs.UpdateDataViewGroup(testCtx, groupTest)
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_DataViewGroupService_ListDataViewGroups(t *testing.T) {
	Convey("Test ListDataViewGroups", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dva := dmock.NewMockDataViewAccess(mockCtrl)
		dvga := dmock.NewMockDataViewGroupAccess(mockCtrl)
		psMock := dmock.NewMockPermissionService(mockCtrl)
		dvsMock := dmock.NewMockDataViewService(mockCtrl)

		dvgs, _ := MockNewDataViewGroupService(appSetting, dvga, dva, dvsMock, psMock)

		Convey("Get failed, caused by the error from driven method ListDataViewGroups", func() {
			expectedErr := errors.New("some errors")
			expectedEntries := []*interfaces.DataViewGroup{
				{GroupID: "1", GroupName: "test1"},
				{GroupID: "2", GroupName: "test2"},
			}

			params := &interfaces.ListViewGroupQueryParams{
				Builtin: []bool{false},
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit: 2,
				},
			}

			dvga.EXPECT().ListDataViewGroups(gomock.Any(), gomock.Any()).Return(expectedEntries, expectedErr)
			_, _, httpErr := dvgs.ListDataViewGroups(testCtx, params, true)

			So(httpErr, ShouldNotBeNil)
		})

		Convey("Get failed, caused by the error from driven method GetDataViewGroupsTotal", func() {
			expectedErr := errors.New("some errors")
			expectedEntries := []*interfaces.DataViewGroup{
				{GroupID: "1", GroupName: "test1"},
				{GroupID: "2", GroupName: "test2"},
			}

			params := &interfaces.ListViewGroupQueryParams{
				Builtin: []bool{false},
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit: 2,
				},
			}

			dvga.EXPECT().ListDataViewGroups(gomock.Any(), gomock.Any()).Return(expectedEntries, nil)
			dvga.EXPECT().GetDataViewGroupsTotal(gomock.Any(), gomock.Any()).Return(0, expectedErr)
			_, _, httpErr := dvgs.ListDataViewGroups(testCtx, params, true)

			So(httpErr, ShouldNotBeNil)
		})

		Convey("Get succeed", func() {
			expectedEntries := []*interfaces.DataViewGroup{
				{GroupID: "1", GroupName: "test1"},
				{GroupID: "2", GroupName: "test2"},
			}

			params := &interfaces.ListViewGroupQueryParams{
				Builtin: []bool{false},
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit: -1,
				},
			}

			SimpleDataView := &interfaces.SimpleDataView{
				ViewID:    "1",
				ViewName:  "16",
				GroupID:   "1",
				GroupName: "group1",
				Tags:      []string{"a", "s", "s", "s", "s"},
			}

			dvga.EXPECT().ListDataViewGroups(gomock.Any(), gomock.Any()).Return(expectedEntries, nil)
			dvga.EXPECT().GetDataViewGroupsTotal(gomock.Any(), gomock.Any()).Return(2, nil)
			dvsMock.EXPECT().ListDataViews(gomock.Any(), gomock.Any()).Return([]*interfaces.SimpleDataView{SimpleDataView}, 1, nil)
			entries, _, httpErr := dvgs.ListDataViewGroups(testCtx, params, true)

			So(entries, ShouldResemble, expectedEntries)
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_DataViewGroupService_GetDataViewGroupByID(t *testing.T) {
	Convey("Test GetDataViewGroupByID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dvga := dmock.NewMockDataViewGroupAccess(mockCtrl)
		psMock := dmock.NewMockPermissionService(mockCtrl)
		dvsMock := dmock.NewMockDataViewService(mockCtrl)

		dvgs, _ := MockNewDataViewGroupService(appSetting, dvga, nil, dvsMock, psMock)

		Convey("Get failed, caused by the error from driven method GetDataViewGroupByID", func() {
			expectedErr := errors.New("some errors")
			dvga.EXPECT().GetDataViewGroupByID(gomock.Any(), gomock.Any()).Return(nil, false, expectedErr)

			groupID := "x"
			_, httpErr := dvgs.GetDataViewGroupByID(testCtx, groupID)
			So(httpErr, ShouldNotBeNil)
		})

		Convey("Get failed, caused by the group is not found", func() {
			dvga.EXPECT().GetDataViewGroupByID(gomock.Any(), gomock.Any()).Return(nil, false, nil)

			groupID := "x"
			_, httpErr := dvgs.GetDataViewGroupByID(testCtx, groupID)
			So(httpErr, ShouldNotBeNil)
		})

		Convey("Get succeed", func() {
			expectedGroup := &interfaces.DataViewGroup{
				GroupID:   "1",
				GroupName: "test1",
			}

			dvga.EXPECT().GetDataViewGroupByID(gomock.Any(), gomock.Any()).Return(expectedGroup, true, nil)

			groupID := "x"
			_, httpErr := dvgs.GetDataViewGroupByID(testCtx, groupID)
			So(httpErr, ShouldBeNil)
		})
	})
}

func Test_DataViewGroupService_CheckDataViewGroupExistByName(t *testing.T) {
	Convey("Test CheckDataViewGroupExistByName", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		groupName := "testGroup"
		group := &interfaces.DataViewGroup{
			GroupID:   "x",
			GroupName: "testGroup",
			Builtin:   false,
		}

		appSetting := &common.AppSetting{}
		dva := dmock.NewMockDataViewAccess(mockCtrl)
		dvga := dmock.NewMockDataViewGroupAccess(mockCtrl)
		psMock := dmock.NewMockPermissionService(mockCtrl)
		dvsMock := dmock.NewMockDataViewService(mockCtrl)

		dvgs, _ := MockNewDataViewGroupService(appSetting, dvga, dva, dvsMock, psMock)

		Convey("Check failed, caused by the error from driven method CheckDataViewGroupExistByName", func() {
			expectedErr := errors.New("some error")

			dvga.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(group, false, expectedErr)

			_, _, err := dvgs.CheckDataViewGroupExistByName(testCtx, nil, groupName, false)
			So(err, ShouldNotBeNil)
		})

		Convey("Check succeed", func() {
			dvga.EXPECT().CheckDataViewGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(group, true, nil)

			_, _, err := dvgs.CheckDataViewGroupExistByName(testCtx, nil, groupName, false)
			So(err, ShouldBeNil)
		})
	})
}
