// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package vega_view

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	rest "github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"uniquery/common"
	cond "uniquery/common/condition"
	"uniquery/interfaces"
	dmock "uniquery/interfaces/mock"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

func MockNewVegaService(appSetting *common.AppSetting, vva interfaces.VegaAccess) (ts *vegaService) {
	ts = &vegaService{
		appSetting: appSetting,
		vva:        vva,
	}
	return ts
}

func Test_VegaService_GetVegaViewFieldsByID(t *testing.T) {
	Convey("Test GetVegaViewFieldsByID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		vva := dmock.NewMockVegaAccess(mockCtrl)
		vvs := MockNewVegaService(appSetting, vva)

		Convey("When GetVegaViewFieldsByID error", func() {
			vva.EXPECT().GetVegaViewFieldsByID(gomock.Any(), gomock.Any()).Return(interfaces.VegaViewWithFields{}, fmt.Errorf("error"))

			resp, err := vvs.GetVegaViewFieldsByID(testCtx, "test-id")
			So(err, ShouldEqual, fmt.Errorf("error"))
			So(resp, ShouldResemble, interfaces.VegaViewWithFields{})
		})

		Convey("success", func() {
			vegaViewFieldsE := interfaces.VegaViewWithFields{
				Fields: []interfaces.VegaViewField{
					{
						Name: "f1",
						Type: "varchar",
					},
					{
						Name: "f2",
						Type: "varchar",
					},
				},
				VegaFieldMap: map[string]interfaces.VegaViewField{
					"f1": {
						Name: "f1",
						Type: "varchar",
					},
					"f2": {
						Name: "f2",
						Type: "varchar",
					},
				},
			}
			vva.EXPECT().GetVegaViewFieldsByID(gomock.Any(), gomock.Any()).Return(interfaces.VegaViewWithFields{
				Fields: []interfaces.VegaViewField{
					{
						Name: "f1",
						Type: "varchar",
					},
					{
						Name: "f2",
						Type: "varchar",
					},
				}}, nil)

			resp, err := vvs.GetVegaViewFieldsByID(testCtx, "test-id")
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, vegaViewFieldsE)
		})

	})
}

func Test_VegaService_FetchDatasFromVega(t *testing.T) {
	Convey("Test FetchDatasFromVega", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		vva := dmock.NewMockVegaAccess(mockCtrl)
		vvs := MockNewVegaService(appSetting, vva)

		Convey("When GetVegaViewFFetchDatasFromVegaieldsByID error", func() {
			vva.EXPECT().FetchDatasFromVega(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.VegaFetchData{}, fmt.Errorf("error"))

			resp, err := vvs.FetchDatasFromVega(testCtx, "test-id")
			So(err, ShouldEqual, fmt.Errorf("error"))
			So(resp, ShouldResemble, interfaces.VegaFetchData{})
		})

		Convey("failed when nextUri error", func() {
			nextUri := "http://vega-gateway:8099/api/virtual_engine_service/v1/statement/executing/1"
			vegaData := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2024-02-01", 2.2},
					{"b", "2024-02-01", 3.2},
				},
				Columns: []cond.ViewField{
					{
						Name: "f1",
						Type: "varchar",
					},
					{
						Name: "__time",
						Type: "varchar",
					},
					{
						Name: "__value",
						Type: "float",
					},
				},
				NextUri: nextUri,
			}
			vva.EXPECT().FetchDatasFromVega(gomock.Any(), "", gomock.Any()).Times(1).Return(vegaData, nil)
			vva.EXPECT().FetchDatasFromVega(gomock.Any(), "1", gomock.Any()).Times(1).Return(vegaData, fmt.Errorf("error"))

			_, err := vvs.FetchDatasFromVega(testCtx, "test-id")
			So(err, ShouldEqual, fmt.Errorf("error"))
		})

		Convey("success", func() {
			vegaDataE := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2024-02-01", 2.2},
					{"b", "2024-02-01", 3.2},
					{"c", "2024-02-01", 2.2},
					{"d", "2024-02-01", 3.2},
				},
				Columns: []cond.ViewField{
					{
						Name: "f1",
						Type: "varchar",
					},
					{
						Name: "__time",
						Type: "varchar",
					},
					{
						Name: "__value",
						Type: "float",
					},
				},
				NextUri: "http://vega-gateway:8099/api/virtual_engine_service/v1/statement/executing/1",
			}

			nextUri := "http://vega-gateway:8099/api/virtual_engine_service/v1/statement/executing/1"
			vegaData := interfaces.VegaFetchData{
				Data: [][]any{
					{"a", "2024-02-01", 2.2},
					{"b", "2024-02-01", 3.2},
				},
				Columns: []cond.ViewField{
					{
						Name: "f1",
						Type: "varchar",
					},
					{
						Name: "__time",
						Type: "varchar",
					},
					{
						Name: "__value",
						Type: "float",
					},
				},
				NextUri: nextUri,
			}

			vegaData2 := interfaces.VegaFetchData{
				Data: [][]any{
					{"c", "2024-02-01", 2.2},
					{"d", "2024-02-01", 3.2},
				},
				Columns: []cond.ViewField{
					{
						Name: "f1",
						Type: "varchar",
					},
					{
						Name: "__time",
						Type: "varchar",
					},
					{
						Name: "__value",
						Type: "float",
					},
				},
			}
			vva.EXPECT().FetchDatasFromVega(gomock.Any(), "", gomock.Any()).Times(1).Return(vegaData, nil)
			vva.EXPECT().FetchDatasFromVega(gomock.Any(), "1", gomock.Any()).Times(1).Return(vegaData2, nil)

			resp, err := vvs.FetchDatasFromVega(testCtx, "test-id")
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, vegaDataE)
		})

	})
}
