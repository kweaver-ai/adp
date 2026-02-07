// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	dmock "data-model/interfaces/mock"
)

// 数据连接
func Test_ValidateDataConnection_ValidateDataConnectionWhenCreate(t *testing.T) {
	Convey("Test validateDataConnectionWhenCreate", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)

		handler := MockNewDataConnectionRestHandler(appSetting, hydra, dcs)

		Convey("Invalid conn, caused by the error from func `validateObjectName`", func() {
			expectedErr := errors.New("some errors")

			patch := ApplyFuncReturn(validateObjectName, expectedErr)
			defer patch.Reset()

			err := validateDataConnectionWhenCreate(testCtx, handler, &interfaces.DataConnection{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid conn, caused by the error from dcService method `GetMapAboutName2ID`", func() {
			expectedErr := errors.New("some errors")

			patch := ApplyFuncReturn(validateObjectName, nil)
			defer patch.Reset()

			dcs.EXPECT().GetMapAboutName2ID(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			err := validateDataConnectionWhenCreate(testCtx, handler, &interfaces.DataConnection{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid conn, caused by the conn name has already exist", func() {
			conn := interfaces.DataConnection{
				Name: "conn1",
			}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusForbidden, derrors.DataModel_DataConnection_ConnectionNameExisted).
				WithErrorDetails(fmt.Sprintf("Data connection name %s already exists", conn.Name))
			expectedConnMap := map[string]string{"conn1": "1"}

			patch := ApplyFuncReturn(validateObjectName, nil)
			defer patch.Reset()

			dcs.EXPECT().GetMapAboutName2ID(gomock.Any(), gomock.Any()).Return(expectedConnMap, nil)

			err := validateDataConnectionWhenCreate(testCtx, handler, &conn)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid conn, caused by the invalid data_source_type", func() {
			conn := interfaces.DataConnection{
				Name: "conn1",
			}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_DataConnection_InvalidParameter_DataSourceType)
			expectedConnMap := map[string]string{"m1": "1"}

			patch := ApplyFuncReturn(validateObjectName, nil)
			defer patch.Reset()

			dcs.EXPECT().GetMapAboutName2ID(gomock.Any(), gomock.Any()).Return(expectedConnMap, nil)

			err := validateDataConnectionWhenCreate(testCtx, handler, &conn)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid conn, caused by the error from func `validateObjectTags`", func() {
			conn := interfaces.DataConnection{
				Name:           "conn1",
				DataSourceType: interfaces.SOURCE_TYPE_TINGYUN,
			}
			expectedConnMap := map[string]string{"m1": "1"}
			expectedErr := errors.New("some errors")

			patch1 := ApplyFuncReturn(validateObjectName, nil)
			defer patch1.Reset()

			dcs.EXPECT().GetMapAboutName2ID(gomock.Any(), gomock.Any()).Return(expectedConnMap, nil)

			patch2 := ApplyFuncReturn(validateObjectTags, expectedErr)
			defer patch2.Reset()

			err := validateDataConnectionWhenCreate(testCtx, handler, &conn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid conn, caused by the error from func `validateObjectComment`", func() {
			conn := interfaces.DataConnection{
				Name:           "conn1",
				DataSourceType: interfaces.SOURCE_TYPE_TINGYUN,
			}
			expectedConnMap := map[string]string{"m1": "1"}
			expectedErr := errors.New("some errors")

			patch1 := ApplyFuncReturn(validateObjectName, nil)
			defer patch1.Reset()

			dcs.EXPECT().GetMapAboutName2ID(gomock.Any(), gomock.Any()).Return(expectedConnMap, nil)

			patch2 := ApplyFuncReturn(validateObjectTags, nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(validateObjectComment, expectedErr)
			defer patch3.Reset()

			err := validateDataConnectionWhenCreate(testCtx, handler, &conn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Valid conn", func() {
			conn := interfaces.DataConnection{
				Name:           "conn1",
				DataSourceType: interfaces.SOURCE_TYPE_TINGYUN,
			}
			expectedConnMap := map[string]string{"m1": "1"}

			patch1 := ApplyFuncReturn(validateObjectName, nil)
			defer patch1.Reset()

			dcs.EXPECT().GetMapAboutName2ID(gomock.Any(), gomock.Any()).Return(expectedConnMap, nil)

			patch2 := ApplyFuncReturn(validateObjectTags, nil)
			defer patch2.Reset()

			patch3 := ApplyFuncReturn(validateObjectComment, nil)
			defer patch3.Reset()

			err := validateDataConnectionWhenCreate(testCtx, handler, &conn)
			So(err, ShouldBeNil)
		})
	})
}

func Test_ValidateDataConnection_ValidateDataConnectionWhenUpdate(t *testing.T) {
	Convey("Test validateDataConnectionWhenUpdate", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		hydra := rmock.NewMockHydra(mockCtrl)
		dcs := dmock.NewMockDataConnectionService(mockCtrl)

		handler := MockNewDataConnectionRestHandler(appSetting, hydra, dcs)

		Convey("Invalid conn, caused by the error from func `validateObjectName`", func() {
			preConn := interfaces.DataConnection{
				Name:           "conn1",
				DataSourceType: interfaces.SOURCE_TYPE_TINGYUN,
			}
			conn := interfaces.DataConnection{
				Name:           "conn2",
				DataSourceType: interfaces.SOURCE_TYPE_TINGYUN,
			}
			expectedErr := errors.New("some errors")

			patch := ApplyFuncReturn(validateObjectName, expectedErr)
			defer patch.Reset()

			err := validateDataConnectionWhenUpdate(testCtx, handler, &preConn, &conn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid conn, caused by the error from dcService method `GetMapAboutName2ID`", func() {
			preConn := interfaces.DataConnection{
				Name:           "conn1",
				DataSourceType: interfaces.SOURCE_TYPE_TINGYUN,
			}
			conn := interfaces.DataConnection{
				Name:           "conn2",
				DataSourceType: interfaces.SOURCE_TYPE_TINGYUN,
			}
			expectedErr := errors.New("some errors")

			patch := ApplyFuncReturn(validateObjectName, nil)
			defer patch.Reset()

			dcs.EXPECT().GetMapAboutName2ID(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			err := validateDataConnectionWhenUpdate(testCtx, handler, &conn, &preConn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid conn, caused by the conn name has already exist", func() {
			preConn := interfaces.DataConnection{
				Name:           "conn1",
				DataSourceType: interfaces.SOURCE_TYPE_TINGYUN,
			}
			conn := interfaces.DataConnection{
				Name:           "conn2",
				DataSourceType: interfaces.SOURCE_TYPE_TINGYUN,
			}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusForbidden, derrors.DataModel_DataConnection_ConnectionNameExisted).
				WithErrorDetails(fmt.Sprintf("Data connection name %s already exists", conn.Name))
			expectedConnMap := map[string]string{"conn2": "1"}

			patch := ApplyFuncReturn(validateObjectName, nil)
			defer patch.Reset()

			dcs.EXPECT().GetMapAboutName2ID(gomock.Any(), gomock.Any()).Return(expectedConnMap, nil)

			err := validateDataConnectionWhenUpdate(testCtx, handler, &conn, &preConn)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Invalid conn, caused by the error from func `validateObjectTags`", func() {
			preConn := interfaces.DataConnection{
				Name:           "conn1",
				DataSourceType: interfaces.SOURCE_TYPE_TINGYUN,
			}
			conn := interfaces.DataConnection{
				Name:           "conn1",
				DataSourceType: interfaces.SOURCE_TYPE_TINGYUN,
			}
			expectedErr := errors.New("some errors")

			patch := ApplyFuncReturn(validateObjectTags, expectedErr)
			defer patch.Reset()

			err := validateDataConnectionWhenUpdate(testCtx, handler, &conn, &preConn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Invalid conn, caused by the error from func `validateObjectComment`", func() {
			preConn := interfaces.DataConnection{
				Name:           "conn1",
				DataSourceType: interfaces.SOURCE_TYPE_TINGYUN,
			}
			conn := interfaces.DataConnection{
				Name:           "conn1",
				DataSourceType: interfaces.SOURCE_TYPE_TINGYUN,
			}
			expectedErr := errors.New("some errors")

			patch1 := ApplyFuncReturn(validateObjectTags, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateObjectComment, expectedErr)
			defer patch2.Reset()

			err := validateDataConnectionWhenUpdate(testCtx, handler, &conn, &preConn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Valid conn", func() {
			preConn := interfaces.DataConnection{
				Name:           "conn1",
				DataSourceType: interfaces.SOURCE_TYPE_TINGYUN,
			}
			conn := interfaces.DataConnection{
				Name:           "conn1",
				DataSourceType: interfaces.SOURCE_TYPE_TINGYUN,
			}

			patch1 := ApplyFuncReturn(validateObjectTags, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(validateObjectComment, nil)
			defer patch2.Reset()

			err := validateDataConnectionWhenUpdate(testCtx, handler, &conn, &preConn)
			So(err, ShouldBeNil)
		})
	})
}
