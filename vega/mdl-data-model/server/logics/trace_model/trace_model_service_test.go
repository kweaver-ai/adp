// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package trace_model

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	dcond "data-model/interfaces/condition"
	dmock "data-model/interfaces/mock"
	"data-model/logics/trace_model/data_source"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

func MockNewTraceModelService(appSetting *common.AppSetting,
	tma interfaces.TraceModelAccess,
	dvs interfaces.DataViewService,
	dcs interfaces.DataConnectionService,
	ps interfaces.PermissionService) *traceModelService {
	return &traceModelService{
		appSetting: appSetting,
		tma:        tma,
		dvs:        dvs,
		dcs:        dcs,
		ps:         ps,
	}
}

func Test_TraceModelService_CreateTraceModels(t *testing.T) {
	Convey("Test CreateTraceModels", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Create failed, caused by the error from tmService method `getDetailedViewMapByIDs`", func() {
			expectedErr := errors.New("some error")

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, expectedErr
				},
			)
			defer patch2.Reset()

			_, err := tms.CreateTraceModels(testCtx, []interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Create failed, caused by the error from tmService method `getConnectionMapByNames`", func() {
			expectedErr := errors.New("some error")
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, expectedErr
				},
			)
			defer patch4.Reset()

			_, err := tms.CreateTraceModels(testCtx, []interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Create failed, caused by the error from tmService method `validateReqTraceModels`", func() {
			expectedErr := errors.New("some error")

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, nil
				},
			)
			defer patch4.Reset()

			patch5 := ApplyPrivateMethod(&traceModelService{}, "validateReqTraceModels",
				func(_ *traceModelService, ctx context.Context, reqModels []interfaces.TraceModel, viewMap map[string]interfaces.DataView) (err error) {
					return expectedErr
				},
			)
			defer patch5.Reset()

			_, err := tms.CreateTraceModels(testCtx, []interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Create failed, caused by the error from tmService method `modifyReqModels`", func() {
			expectedErr := errors.New("some error")

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, nil
				},
			)
			defer patch4.Reset()

			patch5 := ApplyPrivateMethod(&traceModelService{}, "validateReqTraceModels",
				func(_ *traceModelService, ctx context.Context, reqModels []interfaces.TraceModel, viewMap map[string]interfaces.DataView) (err error) {
					return nil
				},
			)
			defer patch5.Reset()

			patch6 := ApplyPrivateMethod(&traceModelService{}, "modifyReqModels",
				func(_ *traceModelService, ctx context.Context, isUpdate bool, viewMap map[string]interfaces.DataView, connMap map[string]string, reqModels []interfaces.TraceModel) (err error) {
					return expectedErr
				},
			)
			defer patch6.Reset()

			_, err := tms.CreateTraceModels(testCtx, []interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Create failed, caused by the error from tma method `CreateTraceModel`", func() {
			expectedErr := errors.New("some error")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_CreateTraceModelsFailed).
				WithErrorDetails(expectedErr.Error())

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, nil
				},
			)
			defer patch4.Reset()

			patch5 := ApplyPrivateMethod(&traceModelService{}, "validateReqTraceModels",
				func(_ *traceModelService, ctx context.Context, reqModels []interfaces.TraceModel, viewMap map[string]interfaces.DataView) (err error) {
					return nil
				},
			)
			defer patch5.Reset()

			patch6 := ApplyPrivateMethod(&traceModelService{}, "modifyReqModels",
				func(_ *traceModelService, ctx context.Context, isUpdate bool, viewMap map[string]interfaces.DataView, connMap map[string]string, reqModels []interfaces.TraceModel) (err error) {
					return nil
				},
			)
			defer patch6.Reset()

			tma.EXPECT().CreateTraceModels(gomock.Any(), gomock.Any()).Return(expectedErr)

			_, err := tms.CreateTraceModels(testCtx, []interfaces.TraceModel{})
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Create succeed", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ps.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, nil
				},
			)
			defer patch4.Reset()

			patch5 := ApplyPrivateMethod(&traceModelService{}, "validateReqTraceModels",
				func(_ *traceModelService, ctx context.Context, reqModels []interfaces.TraceModel, viewMap map[string]interfaces.DataView) (err error) {
					return nil
				},
			)
			defer patch5.Reset()

			patch6 := ApplyPrivateMethod(&traceModelService{}, "modifyReqModels",
				func(_ *traceModelService, ctx context.Context, isUpdate bool, viewMap map[string]interfaces.DataView, connMap map[string]string, reqModels []interfaces.TraceModel) (err error) {
					return nil
				},
			)
			defer patch6.Reset()

			tma.EXPECT().CreateTraceModels(gomock.Any(), gomock.Any()).Return(nil)

			_, err := tms.CreateTraceModels(testCtx, make([]interfaces.TraceModel, 2))
			So(err, ShouldResemble, nil)
		})
	})
}

func Test_TraceModelService_SimulateCreateTraceModel(t *testing.T) {
	Convey("Test SimulateCreateTraceModel", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Simulate create failed, caused by the error from tmService method `getDetailedViewMapByIDs`", func() {
			expectedErr := errors.New("some error")

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, expectedErr
				},
			)
			defer patch2.Reset()

			_, err := tms.SimulateCreateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Simulate create failed, caused by the error from tmService method `getConnectionMapByNames`", func() {
			expectedErr := errors.New("some error")

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, expectedErr
				},
			)
			defer patch4.Reset()

			_, err := tms.SimulateCreateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Simulate create failed, caused by the error from tmService method `validateReqTraceModels`", func() {
			expectedErr := errors.New("some error")

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, nil
				},
			)
			defer patch4.Reset()

			patch5 := ApplyPrivateMethod(&traceModelService{}, "validateReqTraceModels",
				func(_ *traceModelService, ctx context.Context, reqModels []interfaces.TraceModel, viewMap map[string]interfaces.DataView) (err error) {
					return expectedErr
				},
			)
			defer patch5.Reset()

			_, err := tms.SimulateCreateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Simulate create failed, caused by the error from tmService method `modifyReqModels`", func() {
			expectedErr := errors.New("some error")

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, nil
				},
			)
			defer patch4.Reset()

			patch5 := ApplyPrivateMethod(&traceModelService{}, "validateReqTraceModels",
				func(_ *traceModelService, ctx context.Context, reqModels []interfaces.TraceModel, viewMap map[string]interfaces.DataView) (err error) {
					return nil
				},
			)
			defer patch5.Reset()

			patch6 := ApplyPrivateMethod(&traceModelService{}, "modifyReqModels",
				func(_ *traceModelService, ctx context.Context, isUpdate bool, viewMap map[string]interfaces.DataView, connMap map[string]string, reqModels []interfaces.TraceModel) (err error) {
					return expectedErr
				},
			)
			defer patch6.Reset()

			_, err := tms.SimulateCreateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Simulate create succeed", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, nil
				},
			)
			defer patch4.Reset()

			patch5 := ApplyPrivateMethod(&traceModelService{}, "validateReqTraceModels",
				func(_ *traceModelService, ctx context.Context, reqModels []interfaces.TraceModel, viewMap map[string]interfaces.DataView) (err error) {
					return nil
				},
			)
			defer patch5.Reset()

			patch6 := ApplyPrivateMethod(&traceModelService{}, "modifyReqModels",
				func(_ *traceModelService, ctx context.Context, isUpdate bool, viewMap map[string]interfaces.DataView, connMap map[string]string, reqModels []interfaces.TraceModel) (err error) {
					return nil
				},
			)
			defer patch6.Reset()

			_, err := tms.SimulateCreateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelService_DeleteTraceModels(t *testing.T) {
	Convey("Test DeleteTraceModels", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)
		resrc := map[string]interfaces.ResourceOps{
			"0": {
				ResourceID: "0",
			},
			"1": {
				ResourceID: "1",
			},
		}

		Convey("Update failed, caused by the error from tma method `DeleteTraceModels`", func() {
			expectedErr := errors.New("some errors")

			tma.EXPECT().DeleteTraceModels(gomock.Any(), gomock.Any()).Return(expectedErr)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)

			modelIDs := []string{"0", "1"}
			err := tms.DeleteTraceModels(testCtx, modelIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("Update succeed", func() {
			tma.EXPECT().DeleteTraceModels(gomock.Any(), gomock.Any()).Return(nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			ps.EXPECT().DeleteResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			modelIDs := []string{"0", "1"}
			err := tms.DeleteTraceModels(testCtx, modelIDs)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelService_UpdateTraceModel(t *testing.T) {
	Convey("Test UpdateTraceModel", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Update failed, caused by the error from tmService method `getDetailedViewMapByIDs`", func() {
			expectedErr := errors.New("some error")
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, expectedErr
				},
			)
			defer patch2.Reset()

			err := tms.UpdateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Update failed, caused by the error from tmService method `getConnectionMapByNames`", func() {
			expectedErr := errors.New("some error")
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, expectedErr
				},
			)
			defer patch4.Reset()

			err := tms.UpdateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Update failed, caused by the error from tmService method `validateReqTraceModels`", func() {
			expectedErr := errors.New("some error")
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, nil
				},
			)
			defer patch4.Reset()

			patch5 := ApplyPrivateMethod(&traceModelService{}, "validateReqTraceModels",
				func(_ *traceModelService, ctx context.Context, reqModels []interfaces.TraceModel, viewMap map[string]interfaces.DataView) (err error) {
					return expectedErr
				},
			)
			defer patch5.Reset()

			err := tms.UpdateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Update failed, caused by the error from tmService method `modifyReqModels`", func() {
			expectedErr := errors.New("some error")
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, nil
				},
			)
			defer patch4.Reset()

			patch5 := ApplyPrivateMethod(&traceModelService{}, "validateReqTraceModels",
				func(_ *traceModelService, ctx context.Context, reqModels []interfaces.TraceModel, viewMap map[string]interfaces.DataView) (err error) {
					return nil
				},
			)
			defer patch5.Reset()

			patch6 := ApplyPrivateMethod(&traceModelService{}, "modifyReqModels",
				func(_ *traceModelService, ctx context.Context, isUpdate bool, viewMap map[string]interfaces.DataView, connMap map[string]string, reqModels []interfaces.TraceModel) (err error) {
					return expectedErr
				},
			)
			defer patch6.Reset()

			err := tms.UpdateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Update failed, caused by the error from tma method `UpdateTraceModel`", func() {
			expectedErr := errors.New("some error")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_UpdateTraceModelFailed).
				WithErrorDetails(expectedErr.Error())

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, nil
				},
			)
			defer patch4.Reset()

			patch5 := ApplyPrivateMethod(&traceModelService{}, "validateReqTraceModels",
				func(_ *traceModelService, ctx context.Context, reqModels []interfaces.TraceModel, viewMap map[string]interfaces.DataView) (err error) {
					return nil
				},
			)
			defer patch5.Reset()

			patch6 := ApplyPrivateMethod(&traceModelService{}, "modifyReqModels",
				func(_ *traceModelService, ctx context.Context, isUpdate bool, viewMap map[string]interfaces.DataView, connMap map[string]string, reqModels []interfaces.TraceModel) (err error) {
					return nil
				},
			)
			defer patch6.Reset()

			tma.EXPECT().UpdateTraceModel(gomock.Any(), gomock.Any()).Return(expectedErr)

			err := tms.UpdateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Update succeed", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ps.EXPECT().UpdateResource(gomock.Any(), gomock.Any()).Return(nil)
			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, nil
				},
			)
			defer patch4.Reset()

			patch5 := ApplyPrivateMethod(&traceModelService{}, "validateReqTraceModels",
				func(_ *traceModelService, ctx context.Context, reqModels []interfaces.TraceModel, viewMap map[string]interfaces.DataView) (err error) {
					return nil
				},
			)
			defer patch5.Reset()

			patch6 := ApplyPrivateMethod(&traceModelService{}, "modifyReqModels",
				func(_ *traceModelService, ctx context.Context, isUpdate bool, viewMap map[string]interfaces.DataView, connMap map[string]string, reqModels []interfaces.TraceModel) (err error) {
					return nil
				},
			)
			defer patch6.Reset()

			tma.EXPECT().UpdateTraceModel(gomock.Any(), gomock.Any()).Return(nil)

			err := tms.UpdateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelService_SimulateUpdateTraceModel(t *testing.T) {
	Convey("Test SimulateUpdateTraceModel", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Simulate update failed, caused by the error from tmService method `getDetailedViewMapByIDs`", func() {
			expectedErr := errors.New("some error")
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, expectedErr
				},
			)
			defer patch2.Reset()

			_, err := tms.SimulateUpdateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Simulate update failed, caused by the error from tmService method `getConnectionMapByNames`", func() {
			expectedErr := errors.New("some error")
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, expectedErr
				},
			)
			defer patch4.Reset()

			_, err := tms.SimulateUpdateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Simulate update failed, caused by the error from tmService method `validateReqTraceModels`", func() {
			expectedErr := errors.New("some error")
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, nil
				},
			)
			defer patch4.Reset()

			patch5 := ApplyPrivateMethod(&traceModelService{}, "validateReqTraceModels",
				func(_ *traceModelService, ctx context.Context, reqModels []interfaces.TraceModel, viewMap map[string]interfaces.DataView) (err error) {
					return expectedErr
				},
			)
			defer patch5.Reset()

			_, err := tms.SimulateUpdateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Simulate update failed, caused by the error from tmService method `modifyReqModels`", func() {
			expectedErr := errors.New("some error")
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, nil
				},
			)
			defer patch4.Reset()

			patch5 := ApplyPrivateMethod(&traceModelService{}, "validateReqTraceModels",
				func(_ *traceModelService, ctx context.Context, reqModels []interfaces.TraceModel, viewMap map[string]interfaces.DataView) (err error) {
					return nil
				},
			)
			defer patch5.Reset()

			patch6 := ApplyPrivateMethod(&traceModelService{}, "modifyReqModels",
				func(_ *traceModelService, ctx context.Context, isUpdate bool, viewMap map[string]interfaces.DataView, connMap map[string]string, reqModels []interfaces.TraceModel) (err error) {
					return expectedErr
				},
			)
			defer patch6.Reset()

			_, err := tms.SimulateUpdateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Simulate update succeed", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getDetailedViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewNames []string) (map[string]interfaces.DataView, error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionNames",
				func(_ *traceModelService, reqModels []interfaces.TraceModel) []string {
					return nil
				},
			)
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByNames",
				func(_ *traceModelService, ctx context.Context, connNames []string) (connMap map[string]string, err error) {
					return nil, nil
				},
			)
			defer patch4.Reset()

			patch5 := ApplyPrivateMethod(&traceModelService{}, "validateReqTraceModels",
				func(_ *traceModelService, ctx context.Context, reqModels []interfaces.TraceModel, viewMap map[string]interfaces.DataView) (err error) {
					return nil
				},
			)
			defer patch5.Reset()

			patch6 := ApplyPrivateMethod(&traceModelService{}, "modifyReqModels",
				func(_ *traceModelService, ctx context.Context, isUpdate bool, viewMap map[string]interfaces.DataView, connMap map[string]string, reqModels []interfaces.TraceModel) (err error) {
					return nil
				},
			)
			defer patch6.Reset()

			_, err := tms.SimulateUpdateTraceModel(testCtx, interfaces.TraceModel{})
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelService_GetTraceModels(t *testing.T) {
	Convey("Test GetTraceModels", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)
		resrc := map[string]interfaces.ResourceOps{
			"1": {
				ResourceID: "1",
			},
		}

		Convey("Get failed, caused by the error from tma method `GetDetailedTraceModelMapByIDs`", func() {
			modelIDs := []string{"1"}
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_TraceModel_InternalError_GetDetailedTraceModelMapByIDsFailed).WithErrorDetails(expectedErr.Error())

			tma.EXPECT().GetDetailedTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(nil, expectedErr)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)

			_, err := tms.GetTraceModels(testCtx, modelIDs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Get failed, because some models do not exist", func() {
			modelIDs := []string{"1"}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusNotFound, derrors.DataModel_TraceModel_TraceModelNotFound).
				WithErrorDetails(fmt.Sprintf("The trace model whose id equal to %v was not found", modelIDs[0]))

			tma.EXPECT().GetDetailedTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)

			_, err := tms.GetTraceModels(testCtx, modelIDs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Get failed, caused by the error from tma method `getSimpleViewMapByIDs", func() {
			modelIDs := []string{"1"}
			expectedErr := errors.New("some errors")
			expectedModelMap := map[string]interfaces.TraceModel{"1": {}}

			tma.EXPECT().GetDetailedTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(expectedModelMap, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, models []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getSimpleViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewIDs []string) (viewMap map[string]interfaces.DataView, err error) {
					return nil, expectedErr
				})
			defer patch2.Reset()

			_, err := tms.GetTraceModels(testCtx, modelIDs)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from tma method `getConnectionMapByIDs", func() {
			modelIDs := []string{"1"}
			expectedErr := errors.New("some errors")
			expectedModelMap := map[string]interfaces.TraceModel{"1": {}}

			tma.EXPECT().GetDetailedTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(expectedModelMap, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, models []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getSimpleViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewIDs []string) (viewMap map[string]interfaces.DataView, err error) {
					return nil, nil
				})
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionIDs",
				func(_ *traceModelService, models []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByIDs",
				func(_ *traceModelService, ctx context.Context, connIDs []string) (connMap map[string]string, err error) {
					return nil, expectedErr
				})
			defer patch4.Reset()

			_, err := tms.GetTraceModels(testCtx, modelIDs)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed", func() {
			modelIDs := []string{"1"}
			expectedModelMap := map[string]interfaces.TraceModel{"1": {}}

			tma.EXPECT().GetDetailedTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(expectedModelMap, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "getDependentViewIDs",
				func(_ *traceModelService, models []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "getSimpleViewMapByIDs",
				func(_ *traceModelService, ctx context.Context, viewIDs []string) (viewMap map[string]interfaces.DataView, err error) {
					return nil, nil
				})
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&traceModelService{}, "getDependentConnectionIDs",
				func(_ *traceModelService, models []interfaces.TraceModel) []string {
					return []string{}
				})
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(&traceModelService{}, "getConnectionMapByIDs",
				func(_ *traceModelService, ctx context.Context, connIDs []string) (connMap map[string]string, err error) {
					return nil, nil
				})
			defer patch4.Reset()

			patch5 := ApplyPrivateMethod(&traceModelService{}, "modifyResModels",
				func(_ *traceModelService, ctx context.Context, viewMap map[string]interfaces.DataView, connMap map[string]string, resModels []interfaces.TraceModel) {
				})
			defer patch5.Reset()

			_, err := tms.GetTraceModels(testCtx, modelIDs)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelService_ListTraceModels(t *testing.T) {
	Convey("Test ListTraceModels", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)
		resrc := map[string]interfaces.ResourceOps{
			"1": {
				ResourceID: "1",
			},
		}

		Convey("List failed, caused by the error from tma method `ListTraceModels`", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_TraceModel_InternalError_ListTraceModelsFailed).WithErrorDetails(expectedErr.Error())
			queryParams := interfaces.TraceModelListQueryParams{}

			tma.EXPECT().ListTraceModels(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			_, _, err := tms.ListTraceModels(testCtx, queryParams)
			So(err, ShouldResemble, expectedHttpErr)
		})

		// Convey("List failed, caused by the error from tma method `GetTraceModelTotal`", func() {
		// 	expectedErr := errors.New("some errors")
		// 	expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
		// 		derrors.DataModel_TraceModel_InternalError_GetTraceModelTotalFailed).WithErrorDetails(expectedErr.Error())
		// 	queryParams := interfaces.TraceModelListQueryParams{}

		// 	tma.EXPECT().ListTraceModels(gomock.Any(), gomock.Any()).Return(nil, nil)
		// 	tma.EXPECT().GetTraceModelTotal(gomock.Any(), gomock.Any()).Return(int64(0), expectedErr)

		// 	_, _, err := tms.ListTraceModels(testCtx, queryParams)
		// 	So(err, ShouldResemble, expectedHttpErr)
		// })

		Convey("List succeed", func() {
			queryParams := interfaces.TraceModelListQueryParams{}
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil).AnyTimes()
			tma.EXPECT().ListTraceModels(gomock.Any(), gomock.Any()).Return(nil, nil)
			// tma.EXPECT().GetTraceModelTotal(gomock.Any(), gomock.Any()).Return(int64(0), nil)

			_, _, err := tms.ListTraceModels(testCtx, queryParams)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelService_GetTraceModelFieldInfo(t *testing.T) {
	Convey("Test GetTraceModelFieldInfo", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)
		tmp := dmock.NewMockTraceModelProcessor(mockCtrl)
		resrc := map[string]interfaces.ResourceOps{
			"0": {
				ResourceID: "0",
			},
		}

		Convey("Get failed, caused by the error from tma method `GetDetailedTraceModelMapByIDs`", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_TraceModel_InternalError_GetDetailedTraceModelMapByIDsFailed).WithErrorDetails(expectedErr.Error())

			tma.EXPECT().GetDetailedTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(nil, expectedErr)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)

			modelID := "1"
			_, err := tms.GetTraceModelFieldInfo(testCtx, modelID)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Get failed, caused by the trace model was not found", func() {
			modelID := "1"
			errDetails := fmt.Sprintf("The trace model whose id equal to %v was not found", modelID)
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusNotFound, derrors.DataModel_TraceModel_TraceModelNotFound).
				WithErrorDetails(errDetails)

			tma.EXPECT().GetDetailedTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(nil, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)

			_, err := tms.GetTraceModelFieldInfo(testCtx, modelID)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Get failed, caused by the error from method `getUnderlyingDataSourceType`", func() {
			modelID := "1"
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_GetUnderlyingDataSourceTypeFailed)

			tma.EXPECT().GetDetailedTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.TraceModel{
				"1": {},
			}, nil)
			patch := ApplyPrivateMethod(&traceModelService{}, "getUnderlyingDataSourceType",
				func(_ *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (sourceType string, err error) {
					return "", expectedErr
				},
			)
			defer patch.Reset()

			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)

			_, err := tms.GetTraceModelFieldInfo(testCtx, modelID)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from func `NewTraceModelProcessor`", func() {
			modelID := "1"
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_InitTraceModelProcessor)

			tma.EXPECT().GetDetailedTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.TraceModel{
				"1": {},
			}, nil)
			patch1 := ApplyPrivateMethod(&traceModelService{}, "getUnderlyingDataSourceType",
				func(_ *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (sourceType string, err error) {
					return "", nil
				},
			)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(data_source.NewTraceModelProcessor, tmp, expectedErr)
			defer patch2.Reset()

			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)
			_, err := tms.GetTraceModelFieldInfo(testCtx, modelID)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from method `GetSpanFieldInfo`", func() {
			modelID := "1"
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_DependentDataViewNotFound)

			tma.EXPECT().GetDetailedTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.TraceModel{
				"1": {},
			}, nil)
			patch1 := ApplyPrivateMethod(&traceModelService{}, "getUnderlyingDataSourceType",
				func(_ *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (sourceType string, err error) {
					return "", nil
				},
			)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(data_source.NewTraceModelProcessor, tmp, nil)
			defer patch2.Reset()

			tmp.EXPECT().GetSpanFieldInfo(gomock.Any(), gomock.Any()).Return([]interfaces.TraceModelField(nil), expectedErr)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)

			_, err := tms.GetTraceModelFieldInfo(testCtx, modelID)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from method `getUnderlyingDataSourceType` when used a second time", func() {
			modelID := "1"
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_GetUnderlyingDataSourceTypeFailed)

			tma.EXPECT().GetDetailedTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.TraceModel{
				"1": {
					EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
				},
			}, nil)
			patch1 := ApplyPrivateMethod(&traceModelService{}, "getUnderlyingDataSourceType",
				func(_ *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (sourceType string, err error) {
					if queryCategory == interfaces.QUERY_CATEGORY_RELATED_LOG {
						return "", expectedErr
					} else {
						return "", nil
					}
				},
			)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(data_source.NewTraceModelProcessor, tmp, nil)
			defer patch2.Reset()

			tmp.EXPECT().GetSpanFieldInfo(gomock.Any(), gomock.Any()).Return([]interfaces.TraceModelField(nil), nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)

			_, err := tms.GetTraceModelFieldInfo(testCtx, modelID)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from method `NewTraceModelProcessor` when used a second time", func() {
			modelID := "1"
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_InitTraceModelProcessor)

			tma.EXPECT().GetDetailedTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.TraceModel{
				"1": {
					EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
				},
			}, nil)
			patch1 := ApplyPrivateMethod(&traceModelService{}, "getUnderlyingDataSourceType",
				func(_ *traceModelService, ctx context.Context, queryCategory string, model interfaces.TraceModel) (sourceType string, err error) {
					return "", nil
				},
			)
			defer patch1.Reset()

			patch2 := ApplyFuncSeq(data_source.NewTraceModelProcessor, []OutputCell{
				{Values: Params{tmp, nil}, Times: 1},
				{Values: Params{tmp, expectedErr}, Times: 1},
			})
			defer patch2.Reset()

			tmp.EXPECT().GetSpanFieldInfo(gomock.Any(), gomock.Any()).Return([]interfaces.TraceModelField(nil), nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)

			_, err := tms.GetTraceModelFieldInfo(testCtx, modelID)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from method `GetRelatedLogFieldInfo", func() {
			modelID := "1"
			expectedErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_InitTraceModelProcessor)

			tma.EXPECT().GetDetailedTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.TraceModel{
				"1": {
					EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
				},
			}, nil)
			patch1 := ApplyPrivateMethod(&traceModelService{}, "getUnderlyingDataSourceType",
				func(ctx context.Context, queryCategory string, model interfaces.TraceModel) (sourceType string, err error) {
					return "", nil
				},
			)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(data_source.NewTraceModelProcessor, tmp, nil)
			defer patch2.Reset()

			tmp.EXPECT().GetSpanFieldInfo(gomock.Any(), gomock.Any()).Return([]interfaces.TraceModelField(nil), nil)
			tmp.EXPECT().GetRelatedLogFieldInfo(gomock.Any(), gomock.Any()).Return([]interfaces.TraceModelField(nil), expectedErr)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)

			_, err := tms.GetTraceModelFieldInfo(testCtx, modelID)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed", func() {
			modelID := "1"
			tma.EXPECT().GetDetailedTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(map[string]interfaces.TraceModel{
				"1": {
					EnabledRelatedLog: interfaces.RELATED_LOG_OPEN,
				},
			}, nil)
			patch1 := ApplyPrivateMethod(&traceModelService{}, "getUnderlyingDataSourceType",
				func(ctx context.Context, queryCategory string, model interfaces.TraceModel) (sourceType string, err error) {
					return "", nil
				},
			)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(data_source.NewTraceModelProcessor, tmp, nil)
			defer patch2.Reset()

			tmp.EXPECT().GetSpanFieldInfo(gomock.Any(), gomock.Any()).Return([]interfaces.TraceModelField(nil), nil)
			tmp.EXPECT().GetRelatedLogFieldInfo(gomock.Any(), gomock.Any()).Return([]interfaces.TraceModelField(nil), nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(resrc, nil)

			_, err := tms.GetTraceModelFieldInfo(testCtx, modelID)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelService_GetSimpleTraceModelMapByIDs(t *testing.T) {
	Convey("Test GetSimpleTraceModelMapByIDs", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Get failed, caused by the error from tma method 'GetDetailedTraceModelMapByIDs'", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_GetSimpleTraceModelMapByIDsFailed).
				WithErrorDetails(expectedErr.Error())

			tma.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			modelIDs := []string{"1"}
			_, err := tms.GetSimpleTraceModelMapByIDs(testCtx, modelIDs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Get succeed", func() {
			expectedModelMap := map[string]interfaces.TraceModel{}
			tma.EXPECT().GetSimpleTraceModelMapByIDs(gomock.Any(), gomock.Any()).Return(expectedModelMap, nil)

			modelIDs := []string{"1"}
			_, err := tms.GetSimpleTraceModelMapByIDs(testCtx, modelIDs)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelService_GetSimpleTraceModelMapByNames(t *testing.T) {
	Convey("Test GetSimpleTraceModelMapByNames", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Get failed, caused by the error from dvs method 'GetSimpleDataViewMapByIDs'", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_GetSimpleTraceModelMapByNamesFailed).
				WithErrorDetails(expectedErr.Error())

			tma.EXPECT().GetSimpleTraceModelMapByNames(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			modelNames := []string{"model1"}
			_, err := tms.GetSimpleTraceModelMapByNames(testCtx, modelNames)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Get succeed", func() {
			expectedModelMap := map[string]interfaces.TraceModel{}
			tma.EXPECT().GetSimpleTraceModelMapByNames(gomock.Any(), gomock.Any()).Return(expectedModelMap, nil)

			modelNames := []string{"model1"}
			_, err := tms.GetSimpleTraceModelMapByNames(testCtx, modelNames)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelService_GetDetailedViewMapByIDs(t *testing.T) {
	Convey("Test GetDetailedViewMapByIDs", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Get succeed, but no viewNames was passed in", func() {
			viewIDs := []string{}
			_, err := tms.getDetailedViewMapByIDs(testCtx, viewIDs)
			So(err, ShouldBeNil)
		})

		Convey("Get failed, caused by the error from dvs method `GetDetailedDataViewMapByIDs`", func() {
			expectedErr := errors.New("some errors")

			dvs.EXPECT().GetDetailedDataViewMapByIDs(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			viewIDs := []string{"view1"}
			_, err := tms.getDetailedViewMapByIDs(testCtx, viewIDs)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, because some views do not exist", func() {
			viewIDs := []string{"view1"}
			dvs.EXPECT().GetDetailedDataViewMapByIDs(gomock.Any(), gomock.Any()).Return(nil, nil)

			_, err := tms.getDetailedViewMapByIDs(testCtx, viewIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("Get succeed", func() {
			viewIDs := []string{"view1"}
			expectedViewMap := map[string]*interfaces.DataView{
				"view1": {
					Fields: []*interfaces.ViewField{
						{
							Name:    "f1",
							Type:    "t1",
							Comment: "",
						},
					},
				},
			}
			dvs.EXPECT().GetDetailedDataViewMapByIDs(gomock.Any(), gomock.Any()).Return(expectedViewMap, nil)

			_, err := tms.getDetailedViewMapByIDs(testCtx, viewIDs)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelService_GetDependentConnectionNames(t *testing.T) {
	Convey("Test getDependentConnectionNames", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Get succeed", func() {
			models := []interfaces.TraceModel{
				{
					SpanSourceType: interfaces.SOURCE_TYPE_DATA_CONNECTION,
					SpanConfig: interfaces.SpanConfigWithDataConnection{
						DataConnection: interfaces.DataConnectionConfig{
							Name: "conn1",
						},
					},
				},
			}

			expectedConnNames := []string{"conn1"}
			connNames := tms.getDependentConnectionNames(models)
			So(connNames, ShouldResemble, expectedConnNames)
		})
	})
}

func Test_TraceModelService_GetConnectionMapByNames(t *testing.T) {
	Convey("Test getConnectionMapByNames", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Get succeed, but no connNames was passed in'", func() {
			connNames := []string{}
			_, err := tms.getConnectionMapByNames(testCtx, connNames)
			So(err, ShouldBeNil)
		})

		Convey("Get failed, caused by the error from dcs method 'GetMapAboutName2ID'", func() {
			expectedErr := errors.New("some errors")

			dcs.EXPECT().GetMapAboutName2ID(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			connNames := []string{"1"}
			_, err := tms.getConnectionMapByNames(testCtx, connNames)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the data connection was not found", func() {
			errDetails := "The dependent data connection named 1 does not exist in the database!"
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_DependentDataConnectionNotFound).
				WithErrorDetails(errDetails)
			expetedConnMap := map[string]string{}

			dcs.EXPECT().GetMapAboutName2ID(gomock.Any(), gomock.Any()).Return(expetedConnMap, nil)

			connNames := []string{"1"}
			_, err := tms.getConnectionMapByNames(testCtx, connNames)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed", func() {
			expetedConnMap := map[string]string{"conn1": "1"}
			dcs.EXPECT().GetMapAboutName2ID(gomock.Any(), gomock.Any()).Return(expetedConnMap, nil)

			connNames := []string{"conn1"}
			_, err := tms.getConnectionMapByNames(testCtx, connNames)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelService_ValidateReqTraceModels(t *testing.T) {
	Convey("Test validateReqTraceModels", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Validate failed, caused by the error from func `validateSpanBasicAttrs`", func() {
			models := []interfaces.TraceModel{
				{
					SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
					SpanConfig: interfaces.SpanConfigWithDataView{
						DataView: interfaces.DataViewConfig{
							ID:   "view1",
							Name: "view1",
						},
					},
					EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
					RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
					RelatedLogConfig: interfaces.RelatedLogConfigWithDataView{
						DataView: interfaces.DataViewConfig{
							ID:   "view2",
							Name: "view2",
						},
					},
				},
			}
			viewMap := map[string]*interfaces.DataView{
				"view1": {SimpleDataView: interfaces.SimpleDataView{ViewID: "1"}},
				"view2": {SimpleDataView: interfaces.SimpleDataView{ViewID: "2"}},
			}
			expectedErr := errors.New("some errors")

			patch := ApplyPrivateMethod(&traceModelService{}, "validateSpanBasicAttrs",
				func(tms *traceModelService, ctx context.Context, attrs interfaces.SpanConfigWithDataView, fieldMap map[string]string) error {
					return expectedErr
				})
			defer patch.Reset()

			err := tms.validateReqTraceModels(testCtx, models, viewMap)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Validate failed, caused by the error from func `validateRelatedLogWithDataView`", func() {
			models := []interfaces.TraceModel{
				{
					SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
					SpanConfig: interfaces.SpanConfigWithDataView{
						DataView: interfaces.DataViewConfig{
							ID:   "view1",
							Name: "view1",
						},
					},
					EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
					RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
					RelatedLogConfig: interfaces.RelatedLogConfigWithDataView{
						DataView: interfaces.DataViewConfig{
							ID:   "view2",
							Name: "view2",
						},
					},
				},
			}
			viewMap := map[string]*interfaces.DataView{
				"view1": {SimpleDataView: interfaces.SimpleDataView{ViewID: "1"}},
				"view2": {SimpleDataView: interfaces.SimpleDataView{ViewID: "2"}},
			}
			expectedErr := errors.New("some errors")

			patch1 := ApplyPrivateMethod(&traceModelService{}, "validateSpanBasicAttrs",
				func(tms *traceModelService, ctx context.Context, attrs interfaces.SpanConfigWithDataView, fieldMap map[string]string) error {
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validateRelatedLogWithDataView",
				func(tms *traceModelService, ctx context.Context, relatedLog interfaces.RelatedLogConfigWithDataView, fieldMap map[string]string) error {
					return expectedErr
				})
			defer patch2.Reset()

			err := tms.validateReqTraceModels(testCtx, models, viewMap)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Validate succeed", func() {
			models := []interfaces.TraceModel{
				{
					SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
					SpanConfig: interfaces.SpanConfigWithDataView{
						DataView: interfaces.DataViewConfig{
							ID:   "view1",
							Name: "view1",
						},
					},
					EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
					RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
					RelatedLogConfig: interfaces.RelatedLogConfigWithDataView{
						DataView: interfaces.DataViewConfig{
							ID:   "view2",
							Name: "view2",
						},
					},
				},
			}
			viewMap := map[string]*interfaces.DataView{
				"view1": {SimpleDataView: interfaces.SimpleDataView{ViewID: "1"}},
				"view2": {SimpleDataView: interfaces.SimpleDataView{ViewID: "2"}},
			}

			patch1 := ApplyPrivateMethod(&traceModelService{}, "validateSpanBasicAttrs",
				func(tms *traceModelService, ctx context.Context, attrs interfaces.SpanConfigWithDataView, fieldMap map[string]string) error {
					return nil
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validateRelatedLogWithDataView",
				func(tms *traceModelService, ctx context.Context, relatedLog interfaces.RelatedLogConfigWithDataView, fieldMap map[string]string) error {
					return nil
				})
			defer patch2.Reset()

			err := tms.validateReqTraceModels(testCtx, models, viewMap)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelService_ValidateSpanBasicAttrs(t *testing.T) {
	Convey("Test ValidateSpanBasicAttrs", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Invalid spanBasicAttrs, because traceID corresponding field does not exist", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
			}
			fieldMap := map[string]string{}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_TraceID).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", basicAttrs.TraceID.FieldName))

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid spanBasicAttrs, because traceID corresponding field type is invalid", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
			}
			fieldMap := map[string]string{"f1": "type1"}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_TraceID).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", basicAttrs.TraceID.FieldName, interfaces.VALID_FIELD_TYPES_FOR_TRACE_ID))

			patch := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					return false
				})
			defer patch.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid spanBasicAttrs, because spanID corresponding field does not exist", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
			}
			fieldMap := map[string]string{"f1": "type1"}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_SpanID).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", basicAttrs.SpanID.FieldNames[0]))

			patch := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					return true
				})
			defer patch.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid spanBasicAttrs, because spanID corresponding field type is invalid", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
			}
			fieldMap := map[string]string{"f1": "type1", "f2": "type2"}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_SpanID).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", basicAttrs.SpanID.FieldNames[0], interfaces.VALID_FIELD_TYPES_FOR_SPAN_ID))

			patch := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					switch fieldType {
					case "type1":
						return true
					default:
						return false
					}
				})
			defer patch.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid spanBasicAttrs, because the precond of parentSpanID is invalid", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: make([]interfaces.ParentSpanIDConfig, 1),
			}
			fieldMap := map[string]string{"f1": "type1", "f2": "type2"}
			expectedErr := errors.New("some error")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ParentSpanID).
				WithErrorDetails(expectedErr)

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					return true
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return expectedErr
				})
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the parentSpanID corresponding field does not exist", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						Precond:    &interfaces.CondCfg{},
						FieldNames: []string{"f3"},
					},
				},
			}
			fieldMap := map[string]string{"f1": "type1", "f2": "type2"}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ParentSpanID).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", basicAttrs.ParentSpanID[0].FieldNames[0]))

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					return true
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return nil
				})
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the parentSpanID corresponding field type is invalid", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						Precond:    &interfaces.CondCfg{},
						FieldNames: []string{"f3"},
					},
				},
			}
			fieldMap := map[string]string{"f1": "type1", "f2": "type2", "f3": "type3"}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ParentSpanID).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", basicAttrs.ParentSpanID[0].FieldNames[0], interfaces.VALID_FIELD_TYPES_FOR_PARENT_SPAN_ID))

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					switch fieldType {
					case "type1", "type2":
						return true
					default:
						return false
					}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return nil
				})
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the name corresponding field does not exist", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						Precond:    &interfaces.CondCfg{},
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
			}
			fieldMap := map[string]string{"f1": "type1", "f2": "type2", "f3": "type3"}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Name).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", basicAttrs.Name.FieldName))

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					switch fieldType {
					case "type1", "type2", "type3":
						return true
					default:
						return false
					}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return nil
				})
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the name corresponding field type is invalid", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						Precond:    &interfaces.CondCfg{},
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
			}
			fieldMap := map[string]string{"f1": "type1", "f2": "type2", "f3": "type3", "f4": "type4"}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Name).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", basicAttrs.Name.FieldName, interfaces.VALID_FIELD_TYPES_FOR_NAME))

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					switch fieldType {
					case "type1", "type2", "type3":
						return true
					default:
						return false
					}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return nil
				})
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the startTime corresponding field does not exist", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						Precond:    &interfaces.CondCfg{},
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName: "f5",
				},
			}
			fieldMap := map[string]string{"f1": "type1", "f2": "type2", "f3": "type3", "f4": "type4"}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_StartTime).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", basicAttrs.StartTime.FieldName))

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					switch fieldType {
					case "type1", "type2", "type3", "type4":
						return true
					default:
						return false
					}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return nil
				})
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the startTime corresponding field type is invalid", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						Precond:    &interfaces.CondCfg{},
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName: "f5",
				},
			}
			fieldMap := map[string]string{
				"f1": "type1",
				"f2": "type2",
				"f3": "type3",
				"f4": "type4",
				"f5": "type5",
			}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_StartTime).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", basicAttrs.StartTime.FieldName, interfaces.VALID_FIELD_TYPES_FOR_START_TIME))

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					switch fieldType {
					case "type1", "type2", "type3", "type4":
						return true
					default:
						return false
					}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return nil
				})
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the serviceName corresponding field does not exist", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						Precond:    &interfaces.CondCfg{},
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName: "f5",
				},
				ServiceName: interfaces.ServiceNameConfig{
					FieldName: "f6",
				},
			}
			fieldMap := map[string]string{
				"f1": "type1",
				"f2": "type2",
				"f3": "type3",
				"f4": "type4",
				"f5": "type5",
			}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ServiceName).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", basicAttrs.ServiceName.FieldName))

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					switch fieldType {
					case "type1", "type2", "type3", "type4", "type5":
						return true
					default:
						return false
					}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return nil
				})
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the serviceName corresponding field type is invalid", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						Precond:    &interfaces.CondCfg{},
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName: "f5",
				},
				ServiceName: interfaces.ServiceNameConfig{
					FieldName: "f6",
				},
			}
			fieldMap := map[string]string{
				"f1": "type1",
				"f2": "type2",
				"f3": "type3",
				"f4": "type4",
				"f5": "type5",
				"f6": "type6",
			}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_ServiceName).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", basicAttrs.ServiceName.FieldName, interfaces.VALID_FIELD_TYPES_FOR_SERVICE_NAME))

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					switch fieldType {
					case "type1", "type2", "type3", "type4", "type5":
						return true
					default:
						return false
					}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return nil
				})
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the endTime corresponding field does not exist", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						Precond:    &interfaces.CondCfg{},
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName: "f5",
				},
				ServiceName: interfaces.ServiceNameConfig{
					FieldName: "f6",
				},
				EndTime: interfaces.EndTimeConfig{
					FieldName: "f7",
				},
			}
			fieldMap := map[string]string{
				"f1": "type1",
				"f2": "type2",
				"f3": "type3",
				"f4": "type4",
				"f5": "type5",
				"f6": "type6",
			}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_EndTime).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", basicAttrs.EndTime.FieldName))

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					switch fieldType {
					case "type1", "type2", "type3", "type4", "type5", "type6":
						return true
					default:
						return false
					}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return nil
				})
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the endTime corresponding field type is invalid", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						Precond:    &interfaces.CondCfg{},
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName: "f5",
				},
				ServiceName: interfaces.ServiceNameConfig{
					FieldName: "f6",
				},
				EndTime: interfaces.EndTimeConfig{
					FieldName: "f7",
				},
			}
			fieldMap := map[string]string{
				"f1": "type1",
				"f2": "type2",
				"f3": "type3",
				"f4": "type4",
				"f5": "type5",
				"f6": "type6",
				"f7": "type7",
			}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_EndTime).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", basicAttrs.EndTime.FieldName, interfaces.VALID_FIELD_TYPES_FOR_END_TIME))

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					switch fieldType {
					case "type1", "type2", "type3", "type4", "type5", "type6":
						return true
					default:
						return false
					}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return nil
				},
			)
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the duration corresponding field does not exist", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						Precond:    &interfaces.CondCfg{},
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName: "f5",
				},
				ServiceName: interfaces.ServiceNameConfig{
					FieldName: "f6",
				},
				EndTime: interfaces.EndTimeConfig{
					FieldName: "f7",
				},
				Duration: interfaces.DurationConfig{
					FieldName: "f8",
				},
			}
			fieldMap := map[string]string{
				"f1": "type1",
				"f2": "type2",
				"f3": "type3",
				"f4": "type4",
				"f5": "type5",
				"f6": "type6",
				"f7": "type7",
			}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Duration).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", basicAttrs.Duration.FieldName))

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					switch fieldType {
					case "type1", "type2", "type3", "type4", "type5", "type6",
						"type7":
						return true
					default:
						return false
					}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return nil
				},
			)
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the duration corresponding field type is invalid", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						Precond:    &interfaces.CondCfg{},
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName: "f5",
				},
				ServiceName: interfaces.ServiceNameConfig{
					FieldName: "f6",
				},
				EndTime: interfaces.EndTimeConfig{
					FieldName: "f7",
				},
				Duration: interfaces.DurationConfig{
					FieldName: "f8",
				},
			}
			fieldMap := map[string]string{
				"f1": "type1",
				"f2": "type2",
				"f3": "type3",
				"f4": "type4",
				"f5": "type5",
				"f6": "type6",
				"f7": "type7",
				"f8": "type8",
			}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Duration).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", basicAttrs.Duration.FieldName, interfaces.VALID_FIELD_TYPES_FOR_DURATION))

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					switch fieldType {
					case "type1", "type2", "type3", "type4", "type5", "type6",
						"type7":
						return true
					default:
						return false
					}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return nil
				},
			)
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the kind corresponding field does not exist", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						Precond:    &interfaces.CondCfg{},
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName: "f5",
				},
				ServiceName: interfaces.ServiceNameConfig{
					FieldName: "f6",
				},
				EndTime: interfaces.EndTimeConfig{
					FieldName: "f7",
				},
				Duration: interfaces.DurationConfig{
					FieldName: "f8",
				},
				Kind: interfaces.KindConfig{
					FieldName: "f9",
				},
			}
			fieldMap := map[string]string{
				"f1": "type1",
				"f2": "type2",
				"f3": "type3",
				"f4": "type4",
				"f5": "type5",
				"f6": "type6",
				"f7": "type7",
				"f8": "type8",
			}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Kind).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", basicAttrs.Kind.FieldName))

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					switch fieldType {
					case "type1", "type2", "type3", "type4", "type5", "type6",
						"type7", "type8":
						return true
					default:
						return false
					}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return nil
				},
			)
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the kind corresponding field type is invalid", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						Precond:    &interfaces.CondCfg{},
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName: "f5",
				},
				ServiceName: interfaces.ServiceNameConfig{
					FieldName: "f6",
				},
				EndTime: interfaces.EndTimeConfig{
					FieldName: "f7",
				},
				Duration: interfaces.DurationConfig{
					FieldName: "f8",
				},
				Kind: interfaces.KindConfig{
					FieldName: "f9",
				},
			}
			fieldMap := map[string]string{
				"f1": "type1",
				"f2": "type2",
				"f3": "type3",
				"f4": "type4",
				"f5": "type5",
				"f6": "type6",
				"f7": "type7",
				"f8": "type8",
				"f9": "type9",
			}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Kind).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", basicAttrs.Kind.FieldName, interfaces.VALID_FIELD_TYPES_FOR_KIND))

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					switch fieldType {
					case "type1", "type2", "type3", "type4", "type5", "type6",
						"type7", "type8":
						return true
					default:
						return false
					}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return nil
				},
			)
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the status corresponding field does not exist", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						Precond:    &interfaces.CondCfg{},
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName: "f5",
				},
				ServiceName: interfaces.ServiceNameConfig{
					FieldName: "f6",
				},
				EndTime: interfaces.EndTimeConfig{
					FieldName: "f7",
				},
				Duration: interfaces.DurationConfig{
					FieldName: "f8",
				},
				Kind: interfaces.KindConfig{
					FieldName: "f9",
				},
				Status: interfaces.StatusConfig{
					FieldName: "f10",
				},
			}
			fieldMap := map[string]string{
				"f1": "type1",
				"f2": "type2",
				"f3": "type3",
				"f4": "type4",
				"f5": "type5",
				"f6": "type6",
				"f7": "type7",
				"f8": "type8",
				"f9": "type9",
			}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Status).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", basicAttrs.Status.FieldName))

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					switch fieldType {
					case "type1", "type2", "type3", "type4", "type5", "type6",
						"type7", "type8", "type9":
						return true
					default:
						return false
					}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return nil
				},
			)
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid spanBasicAttrs, because the status corresponding field type is invalid", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						Precond:    &interfaces.CondCfg{},
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName: "f5",
				},
				ServiceName: interfaces.ServiceNameConfig{
					FieldName: "f6",
				},
				EndTime: interfaces.EndTimeConfig{
					FieldName: "f7",
				},
				Duration: interfaces.DurationConfig{
					FieldName: "f8",
				},
				Kind: interfaces.KindConfig{
					FieldName: "f9",
				},
				Status: interfaces.StatusConfig{
					FieldName: "f10",
				},
			}
			fieldMap := map[string]string{
				"f1":  "type1",
				"f2":  "type2",
				"f3":  "type3",
				"f4":  "type4",
				"f5":  "type5",
				"f6":  "type6",
				"f7":  "type7",
				"f8":  "type8",
				"f9":  "type9",
				"f10": "type10",
			}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidBasicAttributeConfig_Status).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", basicAttrs.Status.FieldName, interfaces.VALID_FIELD_TYPES_FOR_STATUS))

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					switch fieldType {
					case "type1", "type2", "type3", "type4", "type5", "type6",
						"type7", "type8", "type9":
						return true
					default:
						return false
					}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return nil
				})
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Valid spanBasicAttrs", func() {
			basicAttrs := interfaces.SpanConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
				ParentSpanID: []interfaces.ParentSpanIDConfig{
					{
						Precond:    &interfaces.CondCfg{},
						FieldNames: []string{"f3"},
					},
				},
				Name: interfaces.NameConfig{
					FieldName: "f4",
				},
				StartTime: interfaces.StartTimeConfig{
					FieldName: "f5",
				},
				ServiceName: interfaces.ServiceNameConfig{
					FieldName: "f6",
				},
				EndTime: interfaces.EndTimeConfig{
					FieldName: "f7",
				},
				Duration: interfaces.DurationConfig{
					FieldName: "f8",
				},
				Kind: interfaces.KindConfig{
					FieldName: "f9",
				},
				Status: interfaces.StatusConfig{
					FieldName: "f10",
				},
			}
			fieldMap := map[string]string{
				"f1":  "type1",
				"f2":  "type2",
				"f3":  "type3",
				"f4":  "type4",
				"f5":  "type5",
				"f6":  "type6",
				"f7":  "type7",
				"f8":  "type8",
				"f9":  "type9",
				"f10": "type10",
			}

			patch1 := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					switch fieldType {
					case "type1", "type2", "type3", "type4", "type5", "type6",
						"type7", "type8", "type9", "type10":
						return true
					default:
						return false
					}
				})
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&traceModelService{}, "validPrecond",
				func(tms *traceModelService, precond interfaces.CondCfg, fieldMap map[string]string) error {
					return nil
				})
			defer patch2.Reset()

			err := tms.validateSpanBasicAttrs(testCtx, basicAttrs, fieldMap)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelService_IsValidFieldType(t *testing.T) {
	Convey("Test IsValidFieldType", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Invalid type", func() {
			flag := tms.isValidFieldType("1", []string{"2"})
			So(flag, ShouldBeFalse)
		})

		Convey("Valid type", func() {
			flag := tms.isValidFieldType("1", []string{"1"})
			So(flag, ShouldBeTrue)
		})
	})
}

func Test_TraceModelService_ValidPrecond(t *testing.T) {
	Convey("Test ValidPrecond", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Invalid precond, because precond.Name corresponding field does not exist", func() {
			precond := &interfaces.CondCfg{
				Operation: dcond.OperationEq,
				Name:      "f1",
				ValueOptCfg: interfaces.ValueOptCfg{
					Value: "v1",
				},
			}
			fieldMap := map[string]string{}

			err := tms.validPrecond(precond, fieldMap)
			So(err, ShouldNotBeNil)
		})

		Convey("Invalid precond, because precond.Value corresponding field does not exist", func() {
			precond := &interfaces.CondCfg{
				Operation: dcond.OperationEq,
				Name:      "f1",
				ValueOptCfg: interfaces.ValueOptCfg{
					ValueFrom: dcond.ValueFrom_Field,
					Value:     "f2",
				},
			}
			fieldMap := map[string]string{"f1": "v1"}

			err := tms.validPrecond(precond, fieldMap)
			So(err, ShouldNotBeNil)
		})

		Convey("Invalid precond, caused by the error from subConditions", func() {
			precond := &interfaces.CondCfg{
				Operation: dcond.OperationAnd,
				SubConds: []*interfaces.CondCfg{
					{
						Operation: "",
					},
					{
						Operation: dcond.OperationEq,
						Name:      "f1",
						ValueOptCfg: interfaces.ValueOptCfg{
							Value: "v1",
						},
					},
					{
						Operation: dcond.OperationEq,
						Name:      "f2",
						ValueOptCfg: interfaces.ValueOptCfg{
							Value: "v2",
						},
					},
				},
			}
			fieldMap := map[string]string{"f1": "v1"}

			err := tms.validPrecond(precond, fieldMap)
			So(err, ShouldNotBeNil)
		})

		Convey("Valid precond", func() {
			precond := &interfaces.CondCfg{
				Operation: dcond.OperationAnd,
				SubConds: []*interfaces.CondCfg{
					{
						Operation: "",
					},
					{
						Operation: dcond.OperationEq,
						Name:      "f1",
						ValueOptCfg: interfaces.ValueOptCfg{
							Value: "v1",
						},
					},
				},
			}
			fieldMap := map[string]string{"f1": "v1"}

			err := tms.validPrecond(precond, fieldMap)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelService_ValidateRelatedLogWithDataView(t *testing.T) {
	Convey("Test validateRelatedLogWithDataView", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Invalid spanRelatedLog, because traceID corresponding field does not exist", func() {
			relatedLog := interfaces.RelatedLogConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
			}
			fieldMap := map[string]string{}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidRelatedLogConfig_TraceID).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", relatedLog.TraceID.FieldName))

			err := tms.validateRelatedLogWithDataView(testCtx, relatedLog, fieldMap)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid spanRelatedLog, because traceID corresponding field type is invalid", func() {
			relatedLog := interfaces.RelatedLogConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
			}
			fieldMap := map[string]string{"f1": "type1"}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidRelatedLogConfig_TraceID).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", relatedLog.TraceID.FieldName, interfaces.VALID_FIELD_TYPES_FOR_TRACE_ID))

			patch := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					return false
				},
			)
			defer patch.Reset()

			err := tms.validateRelatedLogWithDataView(testCtx, relatedLog, fieldMap)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid spanRelatedLog, because spanID corresponding field does not exist", func() {
			relatedLog := interfaces.RelatedLogConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
			}
			fieldMap := map[string]string{"f1": "type1"}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidRelatedLogConfig_SpanID).
				WithErrorDetails(fmt.Sprintf("Field %v does not exist in the selected data view", relatedLog.SpanID.FieldNames[0]))

			patch := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					return true
				},
			)
			defer patch.Reset()

			err := tms.validateRelatedLogWithDataView(testCtx, relatedLog, fieldMap)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid spanRelatedLog, because spanID corresponding field type is invalid", func() {
			relatedLog := interfaces.RelatedLogConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
			}
			fieldMap := map[string]string{"f1": "type1", "f2": "type2"}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_TraceModel_InvalidRelatedLogConfig_SpanID).
				WithErrorDetails(fmt.Sprintf("The type of field %v is invalid, valid type is in %v", relatedLog.SpanID.FieldNames[0], interfaces.VALID_FIELD_TYPES_FOR_SPAN_ID))

			patch := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					return fieldType == "type1"
				},
			)
			defer patch.Reset()

			err := tms.validateRelatedLogWithDataView(testCtx, relatedLog, fieldMap)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Valid spanRelatedLog", func() {
			relatedLog := interfaces.RelatedLogConfigWithDataView{
				TraceID: interfaces.TraceIDConfig{
					FieldName: "f1",
				},
				SpanID: interfaces.SpanIDConfig{
					FieldNames: []string{"f2"},
				},
			}
			fieldMap := map[string]string{"f1": "type1", "f2": "type2"}
			patch := ApplyPrivateMethod(&traceModelService{}, "isValidFieldType",
				func(tms *traceModelService, fieldType string, validDataSet []string) bool {
					return true
				},
			)
			defer patch.Reset()

			err := tms.validateRelatedLogWithDataView(testCtx, relatedLog, fieldMap)
			So(err, ShouldBeNil)
		})

	})
}

func Test_TraceModelService_ModifyReqModels(t *testing.T) {
	Convey("Test modifyReqModels", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Modify succeed", func() {
			models := []interfaces.TraceModel{
				{
					SpanSourceType: interfaces.SOURCE_TYPE_DATA_CONNECTION,
					SpanConfig: interfaces.SpanConfigWithDataConnection{
						DataConnection: interfaces.DataConnectionConfig{
							Name: "conn1",
						},
					},
					EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
					RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
					RelatedLogConfig: interfaces.RelatedLogConfigWithDataView{
						DataView: interfaces.DataViewConfig{
							Name: "view2",
						},
					},
				},
			}

			connMap := map[string]string{"conn1": "1"}

			err := tms.modifyReqModels(testCtx, false, connMap, models)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelService_GetDependentViewIDs(t *testing.T) {
	Convey("Test getDependentViewIDs", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Modify succeed", func() {
			models := []interfaces.TraceModel{
				{
					SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
					SpanConfig: interfaces.SpanConfigWithDataView{
						DataView: interfaces.DataViewConfig{
							ID: "1",
						},
					},
					EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
					RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
					RelatedLogConfig: interfaces.RelatedLogConfigWithDataView{
						DataView: interfaces.DataViewConfig{
							ID: "2",
						},
					},
				},
			}

			expectedViewIDs := []string{"1", "2"}
			viewIDs := tms.getDependentViewIDs(models)
			sort.Slice(viewIDs, func(i, j int) bool {
				return viewIDs[i] < viewIDs[j]
			})
			So(viewIDs, ShouldResemble, expectedViewIDs)
		})
	})
}

func Test_TraceModelService_GetSimpleViewMapByIDs(t *testing.T) {
	Convey("Test getSimpleViewMapByIDs", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Get failed, caused by the error from dvs method 'GetSimpleDataViewMapByIDs'", func() {
			expectedErr := errors.New("some errors")

			dvs.EXPECT().GetSimpleDataViewsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			viewIDs := []string{"1"}
			_, err := tms.getSimpleViewMapByIDs(testCtx, viewIDs)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed, but the data view was not found", func() {
			expetedViewMap := map[string]*interfaces.DataView{}

			dvs.EXPECT().GetSimpleDataViewsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(expetedViewMap, nil)

			viewIDs := []string{"1"}
			_, err := tms.getSimpleViewMapByIDs(testCtx, viewIDs)
			So(err, ShouldBeNil)
		})

		Convey("Get succeed", func() {
			expectedViewMap := map[string]*interfaces.DataView{"1": {}}
			dvs.EXPECT().GetSimpleDataViewsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedViewMap, nil)

			viewIDs := []string{"1"}
			_, err := tms.getSimpleViewMapByIDs(testCtx, viewIDs)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelService_GetDependentConnectionIDs(t *testing.T) {
	Convey("Test getDependentConnectionIDs", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Modify succeed", func() {
			models := []interfaces.TraceModel{
				{
					SpanSourceType: interfaces.SOURCE_TYPE_DATA_CONNECTION,
					SpanConfig: interfaces.SpanConfigWithDataConnection{
						DataConnection: interfaces.DataConnectionConfig{
							ID: "1",
						},
					},
				},
			}

			expectedConnIDs := []string{"1"}
			connIDs := tms.getDependentConnectionIDs(models)
			So(connIDs, ShouldResemble, expectedConnIDs)
		})
	})
}

func Test_TraceModelService_GetConnectionMapByIDs(t *testing.T) {
	Convey("Test getConnectionMapByIDs", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Get succeed, but no connIDs was passed in'", func() {
			connIDs := []string{}
			_, err := tms.getConnectionMapByIDs(testCtx, connIDs)
			So(err, ShouldBeNil)
		})

		Convey("Get failed, caused by the error from dcs method 'GetMapAboutID2Name'", func() {
			expectedErr := errors.New("some errors")

			dcs.EXPECT().GetMapAboutID2Name(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			connIDs := []string{"1"}
			_, err := tms.getConnectionMapByIDs(testCtx, connIDs)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed, but the data connection was not found", func() {
			expetedConnMap := map[string]string{}

			dcs.EXPECT().GetMapAboutID2Name(gomock.Any(), gomock.Any()).Return(expetedConnMap, nil)

			connIDs := []string{"1"}
			_, err := tms.getConnectionMapByIDs(testCtx, connIDs)
			So(err, ShouldBeNil)
		})

		Convey("Get succeed", func() {
			expetedConnMap := map[string]string{"1": "conn1"}
			dcs.EXPECT().GetMapAboutID2Name(gomock.Any(), gomock.Any()).Return(expetedConnMap, nil)

			connIDs := []string{"1"}
			_, err := tms.getConnectionMapByIDs(testCtx, connIDs)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TraceModelService_ModifyResModels(t *testing.T) {
	Convey("Test modifyResModels", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("Modify succeed, and span_source_type is data_view", func() {
			models := []interfaces.TraceModel{
				{
					SpanSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
					SpanConfig: interfaces.SpanConfigWithDataView{
						DataView: interfaces.DataViewConfig{
							ID: "1",
						},
					},
					EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
					RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
					RelatedLogConfig: interfaces.RelatedLogConfigWithDataView{
						DataView: interfaces.DataViewConfig{
							ID: "2",
						},
					},
				},
			}
			viewMap := map[string]*interfaces.DataView{
				"1": {SimpleDataView: interfaces.SimpleDataView{ViewName: "view1"}},
				"2": {SimpleDataView: interfaces.SimpleDataView{ViewName: "view2"}},
			}
			connMap := map[string]string{}

			tms.modifyResModels(testCtx, viewMap, connMap, models)

			spanConf, ok := models[0].SpanConfig.(interfaces.SpanConfigWithDataView)
			So(ok, ShouldBeTrue)
			So(spanConf.DataView.Name, ShouldEqual, "view1")

			relatedLogConf, ok := models[0].RelatedLogConfig.(interfaces.RelatedLogConfigWithDataView)
			So(ok, ShouldBeTrue)
			So(relatedLogConf.DataView.Name, ShouldEqual, "view2")
		})

		Convey("Modify succeed, and span_source_type is data_connection", func() {
			models := []interfaces.TraceModel{
				{
					SpanSourceType: interfaces.SOURCE_TYPE_DATA_CONNECTION,
					SpanConfig: interfaces.SpanConfigWithDataConnection{
						DataConnection: interfaces.DataConnectionConfig{
							ID: "1",
						},
					},
					EnabledRelatedLog:    interfaces.RELATED_LOG_OPEN,
					RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_VIEW,
					RelatedLogConfig: interfaces.RelatedLogConfigWithDataView{
						DataView: interfaces.DataViewConfig{
							ID: "2",
						},
					},
				},
			}
			viewMap := map[string]*interfaces.DataView{
				"1": {SimpleDataView: interfaces.SimpleDataView{ViewName: "view1"}},
				"2": {SimpleDataView: interfaces.SimpleDataView{ViewName: "view2"}},
			}
			connMap := map[string]string{
				"1": "conn1",
			}

			tms.modifyResModels(testCtx, viewMap, connMap, models)

			spanConf, ok := models[0].SpanConfig.(interfaces.SpanConfigWithDataConnection)
			So(ok, ShouldBeTrue)
			So(spanConf.DataConnection.Name, ShouldEqual, "conn1")

			relatedLogConf, ok := models[0].RelatedLogConfig.(interfaces.RelatedLogConfigWithDataView)
			So(ok, ShouldBeTrue)
			So(relatedLogConf.DataView.Name, ShouldEqual, "view2")
		})
	})
}

func Test_TraceModelService_GetUnderlyingDataSouceType(t *testing.T) {
	Convey("Test getUnderlyingDataSouceType", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		tma := dmock.NewMockTraceModelAccess(mockCtrl)
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)
		ps := dmock.NewMockPermissionService(mockCtrl)
		tms := MockNewTraceModelService(appSetting, tma, dvs, dcs, ps)

		Convey("other condition, and method 'GetDataConnectionTypeByName' return error", func() {
			expectedErr := errors.New("some errors")
			dcs.EXPECT().GetDataConnectionSourceType(gomock.Any(), gomock.Any()).Return("", false, expectedErr)

			_, err := tms.getUnderlyingDataSourceType(testCtx, interfaces.QUERY_CATEGORY_SPAN, interfaces.TraceModel{
				RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_CONNECTION,
				SpanConfig:           interfaces.SpanConfigWithDataConnection{},
			})
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_GetUnderlyingDataSourceTypeFailed).
				WithErrorDetails(expectedErr.Error()))
		})

		Convey("other condition, and data connection was not found", func() {
			expectedErr := fmt.Errorf("Get underlying data souce type failed, the data connection whose id equal to %v was not found", "1")
			dcs.EXPECT().GetDataConnectionSourceType(gomock.Any(), gomock.Any()).Return("", false, nil)

			_, err := tms.getUnderlyingDataSourceType(testCtx, interfaces.QUERY_CATEGORY_SPAN, interfaces.TraceModel{
				RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_CONNECTION,
				SpanConfig: interfaces.SpanConfigWithDataConnection{
					DataConnection: interfaces.DataConnectionConfig{
						Name: "conn1",
						ID:   "1",
					},
				},
			})
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_GetUnderlyingDataSourceTypeFailed).
				WithErrorDetails(expectedErr.Error()))
		})

		Convey("other condition, and return correct result", func() {
			dcs.EXPECT().GetDataConnectionSourceType(gomock.Any(), gomock.Any()).Return(interfaces.SOURCE_TYPE_TINGYUN, true, nil)

			sourceType, err := tms.getUnderlyingDataSourceType(testCtx, interfaces.QUERY_CATEGORY_SPAN, interfaces.TraceModel{
				RelatedLogSourceType: interfaces.SOURCE_TYPE_DATA_CONNECTION,
				SpanConfig: interfaces.SpanConfigWithDataConnection{
					DataConnection: interfaces.DataConnectionConfig{
						Name: "conn1",
					},
				},
			})
			So(err, ShouldBeNil)
			So(sourceType, ShouldEqual, interfaces.SOURCE_TYPE_TINGYUN)
		})
	})
}
