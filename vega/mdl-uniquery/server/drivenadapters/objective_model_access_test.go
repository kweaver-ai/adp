// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"fmt"
	"net/http"
	"testing"
	"uniquery/common"
	"uniquery/interfaces"

	"github.com/golang/mock/gomock"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetObjectiveModel(t *testing.T) {
	Convey("Test GetObjectiveModel", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		httpClientMock := rmock.NewMockHTTPClient(mockCtrl)
		oma := &objectiveModelAccess{
			appSetting: &common.AppSetting{},
			httpClient: httpClientMock}

		Convey("GetObjectiveModel failed because http client error", func() {
			httpClientMock.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, nil, fmt.Errorf("http client error"))

			model, exists, err := oma.GetObjectiveModel(testCtx, "123")
			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)
			So(model, ShouldResemble, interfaces.ObjectiveModel{})
		})

		Convey("GetObjectiveModel returns not found", func() {
			httpClientMock.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusNotFound, []byte{}, nil)

			model, exists, err := oma.GetObjectiveModel(testCtx, "123")
			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
			So(model, ShouldResemble, interfaces.ObjectiveModel{})
		})

		Convey("GetObjectiveModel failed with non-200 status", func() {
			httpClientMock.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusInternalServerError, []byte(`{"error":"internal error"}`), nil)

			model, exists, err := oma.GetObjectiveModel(testCtx, "123")
			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)
			So(model, ShouldResemble, interfaces.ObjectiveModel{})
		})

		Convey("GetObjectiveModel success", func() {
			httpClientMock.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, []byte(`[{"objective_type":"SLO"}]`), nil)

			model, exists, err := oma.GetObjectiveModel(testCtx, "123")
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
			So(model.ObjectiveType, ShouldEqual, "SLO")
		})
	})
}
