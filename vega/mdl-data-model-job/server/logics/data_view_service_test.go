// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model-job/common"
	"data-model-job/interfaces"
	dmock "data-model-job/interfaces/mock"
)

func MockNewDataViewService(appSetting *common.AppSetting,
	ibAccess interfaces.IndexBaseAccess) *dataViewService {

	return &dataViewService{
		appSetting: appSetting,
		ibAccess:   ibAccess,
	}
}

func Test_DataViewService_GetIndexBases(t *testing.T) {
	Convey("Test jobService GetIndexBases", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		iba := dmock.NewMockIndexBaseAccess(mockCtl)
		dvs := MockNewDataViewService(&common.AppSetting{}, iba)

		ctx := context.Background()
		baseInfos := []interfaces.IndexBase{
			{SimpleIndexBase: interfaces.SimpleIndexBase{BaseType: "aaaa"}},
		}

		Convey("mapstructure decode failed", func() {
			view := &interfaces.DataView{
				DataSource: map[string]any{
					"type": "index_base",
					"index_base": []any{
						"base1",
					},
				},
			}

			_, err := dvs.GetIndexBases(ctx, view)
			So(err, ShouldNotBeNil)
		})

		Convey("GetIndexBasesByTypes failed", func() {
			view := &interfaces.DataView{
				DataSource: map[string]any{
					"type": "index_base",
					"index_base": []any{
						map[string]any{
							"base_type": "base1",
						},
					},
				},
			}

			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(),
				gomock.Any()).Return(baseInfos, errors.New("error"))

			_, err := dvs.GetIndexBases(ctx, view)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			view := &interfaces.DataView{
				DataSource: map[string]any{
					"type": "index_base",
					"index_base": []any{
						map[string]any{
							"base_type": "base1",
						},
					},
				},
			}

			iba.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).Return(baseInfos, nil)

			_, err := dvs.GetIndexBases(ctx, view)
			So(err, ShouldBeNil)
		})

		Convey("unsupport data source type", func() {
			view := &interfaces.DataView{
				DataSource: map[string]any{
					"type": "es",
				},
			}

			_, err := dvs.GetIndexBases(ctx, view)
			So(err, ShouldNotBeNil)
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
