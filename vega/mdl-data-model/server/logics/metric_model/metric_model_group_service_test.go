// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric_model

import (
	"errors"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/did"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	dmock "data-model/interfaces/mock"
)

func MockNewMetricModelGroupService(appSetting *common.AppSetting,
	mma interfaces.MetricModelAccess,
	mmga interfaces.MetricModelGroupAccess,
	mms interfaces.MetricModelService) (*metricModelGroupService, sqlmock.Sqlmock) {

	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	return &metricModelGroupService{
		appSetting: appSetting,
		mma:        mma,
		mmga:       mmga,
		mms:        mms,
		db:         db,
	}, smock
}

func Test_MetricModelGroupService_GetMetricModelGroupByID(t *testing.T) {
	Convey("Test GetMetricModelGroupByID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		mmgs, _ := MockNewMetricModelGroupService(appSetting, mma, mmga, mms)

		metricModelGroup := interfaces.MetricModelGroup{
			GroupID:    "1",
			GroupName:  "groupOld",
			Comment:    "111",
			UpdateTime: testMetricUpdateTime,
		}
		Convey("Success GetMetricModelGroupByID  \n", func() {
			mmga.EXPECT().GetMetricModelGroupByID(gomock.Any(), gomock.Any()).Return(metricModelGroup, true, nil)

			modelGroup, err := mmgs.GetMetricModelGroupByID(testCtx, "1")
			So(modelGroup, ShouldResemble, metricModelGroup)
			So(err, ShouldBeNil)
		})

		Convey("GetMetricModelGroupByID  failed\n", func() {
			err := errors.New("Get metric models by groupID failed")
			mmga.EXPECT().GetMetricModelGroupByID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModelGroup{}, false, err)
			modelGroup, httpErr := mmgs.GetMetricModelGroupByID(testCtx, "1")
			So(modelGroup, ShouldResemble, interfaces.MetricModelGroup{})
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModelGroup_InternalError_GetGroupByIDFailed)
		})

		Convey("Group Not Exist \n", func() {

			mmga.EXPECT().GetMetricModelGroupByID(gomock.Any(), gomock.Any()).Return(interfaces.MetricModelGroup{}, false, nil)
			modelGroup, httpErr := mmgs.GetMetricModelGroupByID(testCtx, "1")
			So(modelGroup, ShouldResemble, interfaces.MetricModelGroup{})
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusNotFound)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModelGroup_GroupNotFound)
		})
	})
}

func Test_MetricModelGroupService_CheckMetricModelGroupExist(t *testing.T) {
	Convey("Test CheckMetricModelGroupExist", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		mmgs, _ := MockNewMetricModelGroupService(appSetting, mma, mmga, mms)

		Convey("Success CheckMetricModelGroupExist  \n", func() {
			mmga.EXPECT().CheckMetricModelGroupExist(gomock.Any(), gomock.Any()).
				Return(testMetricGroup, true, nil)

			exist, err := mmgs.CheckMetricModelGroupExist(testCtx, "group1")

			So(exist, ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("CheckMetricModelGroupExist  failed\n", func() {
			err := errors.New("CheckMetricModelGroupExist error")
			mmga.EXPECT().CheckMetricModelGroupExist(gomock.Any(), gomock.Any()).
				Return(testMetricGroup, false, err)

			exist, httpErr := mmgs.CheckMetricModelGroupExist(testCtx, "group1")

			So(exist, ShouldBeFalse)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModelGroup_InternalError_CheckGroupIfExistFailed)

		})
	})
}

func Test_MetricModelGroupService_CreateMetricModelGroup(t *testing.T) {
	Convey("Test CreateMetricModelGroup", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		mmgs, smock := MockNewMetricModelGroupService(appSetting, mma, mmga, mms)

		metricModelGroup := interfaces.MetricModelGroup{
			GroupName: "groupOld",
			Comment:   "111",
		}
		Convey("Success CreateMetricModelGroup  \n", func() {
			smock.ExpectBegin()
			mmga.EXPECT().CreateMetricModelGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			_, err := mmgs.CreateMetricModelGroup(testCtx, metricModelGroup)
			So(err, ShouldBeNil)
		})

		Convey("CreateMetricModelGroup  failed  \n", func() {
			err := errors.New("CreateMetricModelGroup failed")
			patch := ApplyFuncReturn(did.GenerateDistributedID, 110, nil)
			defer patch.Reset()

			smock.ExpectBegin()
			mmga.EXPECT().CreateMetricModelGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(err)
			smock.ExpectCommit()

			groupID, httpErr := mmgs.CreateMetricModelGroup(testCtx, metricModelGroup)

			So(groupID, ShouldEqual, "")
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModelGroup_InternalError)

		})
	})
}

func Test_MetricModelGroupService_UpdateMetricModelGroup(t *testing.T) {
	Convey("Test CreateMetricModelGroup", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		mmgs, _ := MockNewMetricModelGroupService(appSetting, mma, mmga, mms)

		metricModelGroup := interfaces.MetricModelGroup{
			GroupName: "groupOld",
			Comment:   "111",
		}
		Convey("Success UpdateMetricModelGroup  \n", func() {

			mmga.EXPECT().UpdateMetricModelGroup(gomock.Any(), gomock.Any()).Return(nil)

			httpErr := mmgs.UpdateMetricModelGroup(testCtx, metricModelGroup)
			So(httpErr, ShouldBeNil)
		})

		Convey("UpdateMetricModelGroup  failed  \n", func() {
			err := errors.New("UpdateMetricModelGroup failed")
			mmga.EXPECT().UpdateMetricModelGroup(gomock.Any(), gomock.Any()).Return(err)

			httpErr := mmgs.UpdateMetricModelGroup(testCtx, metricModelGroup)

			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModelGroup_InternalError)

		})
	})
}

func Test_MetricModelGroupService_ListMetricModelGroups(t *testing.T) {
	Convey("Test ListMetricModelGroups", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		mmgs, _ := MockNewMetricModelGroupService(appSetting, mma, mmga, mms)

		parameter := interfaces.ListMetricGroupQueryParams{
			Builtin: []bool{false},
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit:  -1,
				Offset: 0,
			},
		}

		metricModelGroupWithCount := interfaces.MetricModelGroup{
			GroupID:          "1",
			GroupName:        "group1",
			Comment:          "111",
			UpdateTime:       testMetricUpdateTime,
			MetricModelCount: 1,
		}
		simpleMetricModel := interfaces.SimpleMetricModel{
			ModelID:    "1",
			ModelName:  "16",
			GroupID:    "1",
			GroupName:  "group1",
			MetricType: "atomic",
			QueryType:  "promql",
			Tags:       []string{"a", "s", "s", "s", "s"},
		}
		Convey("Success ListMetricModelGroups  \n", func() {

			mmga.EXPECT().ListMetricModelGroups(gomock.Any(), gomock.Any()).Return([]*interfaces.MetricModelGroup{&metricModelGroupWithCount}, nil)
			mmga.EXPECT().GetMetricModelGroupsTotal(gomock.Any(), gomock.Any()).Return(1, nil)
			mms.EXPECT().ListSimpleMetricModels(gomock.Any(), gomock.Any()).Return([]interfaces.SimpleMetricModel{simpleMetricModel}, 1, nil)

			modelGroups, total, httpErr := mmgs.ListMetricModelGroups(testCtx, parameter)
			So(modelGroups, ShouldResemble, []*interfaces.MetricModelGroup{&metricModelGroupWithCount})
			So(total, ShouldEqual, 1)
			So(httpErr, ShouldBeNil)

		})

		Convey("ListMetricModelGroups  failed  \n", func() {
			err := errors.New("ListMetricModelGroups failed")
			mmga.EXPECT().ListMetricModelGroups(gomock.Any(), gomock.Any()).Return([]*interfaces.MetricModelGroup{}, err)

			modelGroups, total, httpErr := mmgs.ListMetricModelGroups(testCtx, parameter)
			So(modelGroups, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModelGroup_InternalError_ListGroupsFailed)

		})

		Convey("GetMetricModelGroupsTotal  failed  \n", func() {
			err := errors.New("GetMetricModelGroupsTotal failed")
			mmga.EXPECT().ListMetricModelGroups(gomock.Any(), gomock.Any()).Return([]*interfaces.MetricModelGroup{&metricModelGroupWithCount}, nil)
			mmga.EXPECT().GetMetricModelGroupsTotal(gomock.Any(), gomock.Any()).Return(0, err)

			modelGroups, total, httpErr := mmgs.ListMetricModelGroups(testCtx, parameter)
			So(modelGroups, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModelGroup_InternalError_GetGroupsTotalFailed)

		})
	})
}

func Test_MetricModelGroupService_DeleteMetricModelGroup(t *testing.T) {
	Convey("Test DeleteMetricModelGroup", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		mmgs, smock := MockNewMetricModelGroupService(appSetting, mma, mmga, mms)

		model := interfaces.MetricModel{
			SimpleMetricModel: interfaces.SimpleMetricModel{
				ModelID: "1",
			},
		}
		Convey("Success DeleteMetricModelGroup by force is true \n", func() {
			smock.ExpectBegin()
			mma.EXPECT().GetMetricModelsByGroupID(gomock.Any(), gomock.Any()).Return([]interfaces.MetricModel{model}, nil)
			mmga.EXPECT().DeleteMetricModelGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			mms.EXPECT().DeleteMetricModels(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			smock.ExpectCommit()

			rowsAffect, metricModels, httpErr := mmgs.DeleteMetricModelGroup(testCtx, "1", true)

			So(rowsAffect, ShouldEqual, int64(1))
			So(metricModels, ShouldResemble, []interfaces.MetricModel{model})
			So(httpErr, ShouldBeNil)

		})
		Convey("GetMetricModelsByGroupID  failed  \n", func() {
			mockErr := errors.New("mock error")
			smock.ExpectBegin()
			mma.EXPECT().GetMetricModelsByGroupID(gomock.Any(), gomock.Any()).Return([]interfaces.MetricModel{}, mockErr)
			smock.ExpectCommit()

			rowsAffect, metricModels, err := mmgs.DeleteMetricModelGroup(testCtx, "1", true)

			So(rowsAffect, ShouldEqual, int64(0))
			So(metricModels, ShouldResemble, []interfaces.MetricModel{})
			So(err.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModelGroup_InternalError)
		})

		Convey("DeleteMetricModelGroup  failed  \n", func() {
			err := errors.New("DeleteMetricModelGroup failed")
			smock.ExpectBegin()
			mma.EXPECT().GetMetricModelsByGroupID(gomock.Any(), gomock.Any()).Return([]interfaces.MetricModel{model}, nil)
			mmga.EXPECT().DeleteMetricModelGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), err)
			smock.ExpectCommit()

			rowsAffect, metricModels, httpErr := mmgs.DeleteMetricModelGroup(testCtx, "1", true)
			So(rowsAffect, ShouldEqual, int64(0))
			So(metricModels, ShouldResemble, []interfaces.MetricModel{model})
			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModelGroup_InternalError)
		})

		Convey("DeleteMetricModelGroup  failed because of DeleteMetricModels failed \n", func() {
			err := errors.New("DeleteMetricModelGroup failed")
			smock.ExpectBegin()
			mma.EXPECT().GetMetricModelsByGroupID(gomock.Any(), gomock.Any()).Return([]interfaces.MetricModel{model}, nil)
			mmga.EXPECT().DeleteMetricModelGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			mms.EXPECT().DeleteMetricModels(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), err)
			smock.ExpectCommit()

			rowsAffect, metricModels, httpErr := mmgs.DeleteMetricModelGroup(testCtx, "1", true)
			So(rowsAffect, ShouldEqual, int64(0))
			So(metricModels, ShouldResemble, []interfaces.MetricModel{model})
			So(httpErr, ShouldEqual, err)
		})
	})
}

func Test_MetricModelGroupService_DeleteMetricModelGroupAndModels(t *testing.T) {
	Convey("Test DeleteMetricModelGroupAndModels", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		mma := dmock.NewMockMetricModelAccess(mockCtrl)
		mmga := dmock.NewMockMetricModelGroupAccess(mockCtrl)
		mms := dmock.NewMockMetricModelService(mockCtrl)
		mmgs, smock := MockNewMetricModelGroupService(appSetting, mma, mmga, mms)

		Convey("Success DeleteMetricModelGroupAndModels  \n", func() {
			smock.ExpectBegin()
			mmga.EXPECT().DeleteMetricModelGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			mms.EXPECT().DeleteMetricModels(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			smock.ExpectCommit()

			rowsAffect, httpErr := mmgs.DeleteMetricModelGroupAndModels(testCtx, "1", []string{"1", "2"})
			So(rowsAffect, ShouldEqual, int64(1))
			So(httpErr, ShouldBeNil)

		})

		Convey("DeleteMetricModelGroupAndModels  failed  \n", func() {
			err := errors.New("DeleteMetricModelGroupAndModels failed")
			smock.ExpectBegin()
			mmga.EXPECT().DeleteMetricModelGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), err)
			smock.ExpectCommit()

			rowsAffect, httpErr := mmgs.DeleteMetricModelGroupAndModels(testCtx, "1", []string{"1", "2"})
			So(rowsAffect, ShouldEqual, int64(0))

			So(httpErr.(*rest.HTTPError).HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(httpErr.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, derrors.DataModel_MetricModelGroup_InternalError)
		})

		Convey("DeleteMetricModelGroupAndModels failed because of DeleteMetricModels failed \n", func() {
			err := errors.New("DeleteMetricModelGroupAndModels failed")
			smock.ExpectBegin()
			mmga.EXPECT().DeleteMetricModelGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			mms.EXPECT().DeleteMetricModels(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), err)
			smock.ExpectCommit()

			rowsAffect, httpErr := mmgs.DeleteMetricModelGroupAndModels(testCtx, "1", []string{"1", "2"})
			So(rowsAffect, ShouldEqual, int64(0))

			So(httpErr, ShouldEqual, err)
		})
	})
}
