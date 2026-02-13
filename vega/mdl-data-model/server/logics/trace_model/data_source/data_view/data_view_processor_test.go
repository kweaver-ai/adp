// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_view

import (
	"context"
	"fmt"
	"net/http"
	"testing"

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

func MockNewDataViewTraceProcessor(appSetting *common.AppSetting,
	dvService interfaces.DataViewService) *dataViewTraceProcessor {
	return &dataViewTraceProcessor{
		appSetting: appSetting,
		dvService:  dvService,
	}
}

func Test_DataViewProcessor_GetSpanFieldInfo(t *testing.T) {
	Convey("Test GetSpanFieldInfo", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dvtp := MockNewDataViewTraceProcessor(appSetting, dvs)

		Convey("Get failed, caused by the error from method `GetDataViews`", func() {
			model := interfaces.TraceModel{
				SpanConfig: interfaces.SpanConfigWithDataView{
					DataView: interfaces.DataViewConfig{
						Name: "1",
						ID:   "1",
					},
				},
			}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusNotFound,
				derrors.DataModel_DataView_InternalError_GetDataViewsFailed)
			dvs.EXPECT().GetDataViews(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*interfaces.DataView{}, expectedErr)

			_, err := dvtp.GetSpanFieldInfo(testCtx, model)
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_TraceModel_DependentDataViewNotFound).
				WithErrorDetails(fmt.Sprintf("The data view whose id equal to %v was not found", "1")))
		})

		Convey("Get succeed", func() {
			model := interfaces.TraceModel{
				SpanConfig: interfaces.SpanConfigWithDataView{
					DataView: interfaces.DataViewConfig{
						Name: "1",
						ID:   "1",
					},
				},
			}
			dvs.EXPECT().GetDataViews(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*interfaces.DataView{{
				Fields: []*interfaces.ViewField{
					{},
				},
			}}, nil)

			_, err := dvtp.GetSpanFieldInfo(testCtx, model)
			So(err, ShouldBeNil)
		})
	})
}

func Test_DataViewProcessor_GetRelatedLogFieldInfo(t *testing.T) {
	Convey("Test GetRelatedLogFieldInfo", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dvs := dmock.NewMockDataViewService(mockCtrl)
		dvtp := MockNewDataViewTraceProcessor(appSetting, dvs)

		Convey("Get failed, caused by the error from method `GetDataViews`", func() {
			model := interfaces.TraceModel{
				RelatedLogConfig: interfaces.RelatedLogConfigWithDataView{
					DataView: interfaces.DataViewConfig{
						Name: "1",
						ID:   "1",
					},
				},
			}
			expectedErr := rest.NewHTTPError(testCtx, http.StatusNotFound, derrors.DataModel_DataView_InternalError_GetDataViewsFailed)
			dvs.EXPECT().GetDataViews(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*interfaces.DataView{}, expectedErr)

			_, err := dvtp.GetRelatedLogFieldInfo(testCtx, model)
			So(err, ShouldResemble, rest.NewHTTPError(testCtx, http.StatusInternalServerError, derrors.DataModel_TraceModel_DependentDataViewNotFound).
				WithErrorDetails(fmt.Sprintf("The data view whose id equal to %v was not found", "1")))
		})

		Convey("Get succeed", func() {
			model := interfaces.TraceModel{
				RelatedLogConfig: interfaces.RelatedLogConfigWithDataView{
					DataView: interfaces.DataViewConfig{
						Name: "1",
						ID:   "1",
					},
				},
			}
			dvs.EXPECT().GetDataViews(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*interfaces.DataView{{
				Fields: []*interfaces.ViewField{
					{},
				},
			}}, nil)

			_, err := dvtp.GetRelatedLogFieldInfo(testCtx, model)
			So(err, ShouldBeNil)
		})
	})
}
