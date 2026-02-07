// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_connection

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	dmock "data-model/interfaces/mock"
	"data-model/logics/data_connection/data_source"
)

var (
	testENCtx = context.WithValue(context.Background(), rest.XLangKey, rest.AmericanEnglish)
)

func MockNewDataConnectionService(appSetting *common.AppSetting,
	dca interfaces.DataConnectionAccess) (*dataConnectionService, sqlmock.Sqlmock) {

	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	dcs := &dataConnectionService{
		appSetting: appSetting,
		db:         db,
		dca:        dca,
	}
	return dcs, smock
}

func Test_DataConnectionService_CreateDataConnection(t *testing.T) {
	Convey("Test CreateDataConnection", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dca := dmock.NewMockDataConnectionAccess(mockCtrl)
		dcs, smock := MockNewDataConnectionService(appSetting, dca)
		dcp := dmock.NewMockDataConnectionProcessor(mockCtrl)

		Convey("Create failed, caused by the error from method `ValidateWhenCreate`", func() {
			reqConn := interfaces.DataConnection{}
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_UnmarshalDataFailed)

			patch := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch.Reset()

			dcp.EXPECT().ValidateWhenCreate(gomock.Any(), gomock.Any()).Return(expectedErr)

			_, err := dcs.CreateDataConnection(testENCtx, &reqConn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Create failed, caused by the error from method `ComputeConfigMD5`", func() {
			reqConn := interfaces.DataConnection{}
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_UnmarshalDataFailed)

			patch := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch.Reset()

			dcp.EXPECT().ValidateWhenCreate(gomock.Any(), gomock.Any()).Return(nil)
			dcp.EXPECT().ComputeConfigMD5(gomock.Any(), gomock.Any()).Return("", expectedErr)

			_, err := dcs.CreateDataConnection(testENCtx, &reqConn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Create failed, caused by the error from method `getDataConnectionsByConfigMD5`", func() {
			reqConn := interfaces.DataConnection{}
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_UnmarshalDataFailed)

			patch1 := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&dataConnectionService{}, "getDataConnectionsByConfigMD5",
				func(dcService *dataConnectionService, ctx context.Context, configMD5 string) (connMap map[string]*interfaces.DataConnection, err error) {
					return nil, expectedErr
				},
			)
			defer patch2.Reset()

			dcp.EXPECT().ValidateWhenCreate(gomock.Any(), gomock.Any()).Return(nil)
			dcp.EXPECT().ComputeConfigMD5(gomock.Any(), gomock.Any()).Return("pre", nil)

			_, err := dcs.CreateDataConnection(testENCtx, &reqConn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Create failed, caused by the same config md5 was found", func() {
			reqConn := interfaces.DataConnection{}
			expectedConnMap := map[string]*interfaces.DataConnection{
				"1": {ID: "1"},
			}
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusBadRequest,
				derrors.DataModel_DataConnection_DuplicatedParameter_Config).
				WithErrorDetails(fmt.Sprintf("Same config whose conn_id in %s already exists in the database", expectedConnMap["1"].ID))
			patch1 := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&dataConnectionService{}, "getDataConnectionsByConfigMD5",
				func(dcService *dataConnectionService, ctx context.Context, configMD5 string) (connMap map[string]*interfaces.DataConnection, err error) {
					return expectedConnMap, nil
				},
			)
			defer patch2.Reset()

			dcp.EXPECT().ValidateWhenCreate(gomock.Any(), gomock.Any()).Return(nil)
			dcp.EXPECT().ComputeConfigMD5(gomock.Any(), gomock.Any()).Return("pre", nil)

			_, err := dcs.CreateDataConnection(testENCtx, &reqConn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Create failed, caused by the error from method `GenerateAuthInfoAndStatus`", func() {
			reqConn := interfaces.DataConnection{}
			expectedConnMap := map[string]interfaces.DataConnection{
				"1": {},
			}
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusBadRequest,
				derrors.DataModel_DataConnection_DuplicatedParameter_Config).WithErrorDetails(fmt.Sprintf("Same config whose conn_id is %v already exists in the database", expectedConnMap["0"].ID))
			patch1 := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&dataConnectionService{}, "getDataConnectionsByConfigMD5",
				func(dcService *dataConnectionService, ctx context.Context, configMD5 string) (connMap map[string]*interfaces.DataConnection, err error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			dcp.EXPECT().ValidateWhenCreate(gomock.Any(), gomock.Any()).Return(nil)
			dcp.EXPECT().ComputeConfigMD5(gomock.Any(), gomock.Any()).Return("pre", nil)
			dcp.EXPECT().GenerateAuthInfoAndStatus(gomock.Any(), gomock.Any()).Return(expectedErr)

			_, err := dcs.CreateDataConnection(testENCtx, &reqConn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Create failed, because the transaction begin failed", func() {
			reqConn := interfaces.DataConnection{}
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError, derrors.DataModel_InternalError_BeginTransactionFailed).
				WithErrorDetails(fmt.Sprintf("Begin transaction failed when creating a data connection, err: %v", expectedErr.Error()))

			patch1 := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&dataConnectionService{}, "getDataConnectionsByConfigMD5",
				func(dcService *dataConnectionService, ctx context.Context, configMD5 string) (connMap map[string]*interfaces.DataConnection, err error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			dcp.EXPECT().ValidateWhenCreate(gomock.Any(), gomock.Any()).Return(nil)
			dcp.EXPECT().ComputeConfigMD5(gomock.Any(), gomock.Any()).Return("pre", nil)
			dcp.EXPECT().GenerateAuthInfoAndStatus(gomock.Any(), gomock.Any()).Return(nil)

			smock.ExpectBegin().WillReturnError(expectedErr)

			_, err := dcs.CreateDataConnection(testENCtx, &reqConn)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Create failed, caused by the error from method `CreateDataConnection`", func() {
			reqConn := interfaces.DataConnection{}
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_CreateDataConnectionFailed).WithErrorDetails(expectedErr.Error())

			patch1 := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&dataConnectionService{}, "getDataConnectionsByConfigMD5",
				func(dcService *dataConnectionService, ctx context.Context, configMD5 string) (connMap map[string]*interfaces.DataConnection, err error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			dcp.EXPECT().ValidateWhenCreate(gomock.Any(), gomock.Any()).Return(nil)
			dcp.EXPECT().ComputeConfigMD5(gomock.Any(), gomock.Any()).Return("pre", nil)
			dcp.EXPECT().GenerateAuthInfoAndStatus(gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin().WillReturnError(nil)
			dca.EXPECT().CreateDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)

			_, err := dcs.CreateDataConnection(testENCtx, &reqConn)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Create failed, caused by the error from method `CreateDataConnectionStatus`", func() {
			reqConn := interfaces.DataConnection{}
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_CreateDataConnectionStatusFailed).WithErrorDetails(expectedErr.Error())

			patch1 := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&dataConnectionService{}, "getDataConnectionsByConfigMD5",
				func(dcService *dataConnectionService, ctx context.Context, configMD5 string) (connMap map[string]*interfaces.DataConnection, err error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			dcp.EXPECT().ValidateWhenCreate(gomock.Any(), gomock.Any()).Return(nil)
			dcp.EXPECT().ComputeConfigMD5(gomock.Any(), gomock.Any()).Return("pre", nil)
			dcp.EXPECT().GenerateAuthInfoAndStatus(gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin().WillReturnError(nil)
			dca.EXPECT().CreateDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dca.EXPECT().CreateDataConnectionStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)

			_, err := dcs.CreateDataConnection(testENCtx, &reqConn)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Create succeed, but transaction commit failed", func() {
			reqConn := interfaces.DataConnection{}
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_CommitTransactionFailed).WithErrorDetails(fmt.Sprintf("Commit transaction failed when creating a data connection, err: %v", expectedErr))

			patch1 := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&dataConnectionService{}, "getDataConnectionsByConfigMD5",
				func(dcService *dataConnectionService, ctx context.Context, configMD5 string) (connMap map[string]*interfaces.DataConnection, err error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			dcp.EXPECT().ValidateWhenCreate(gomock.Any(), gomock.Any()).Return(nil)
			dcp.EXPECT().ComputeConfigMD5(gomock.Any(), gomock.Any()).Return("pre", nil)
			dcp.EXPECT().GenerateAuthInfoAndStatus(gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin().WillReturnError(nil)
			smock.ExpectCommit().WillReturnError(expectedErr)
			dca.EXPECT().CreateDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dca.EXPECT().CreateDataConnectionStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			_, err := dcs.CreateDataConnection(testENCtx, &reqConn)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Create succeed, and transaction commit succeed", func() {
			reqConn := interfaces.DataConnection{}

			patch1 := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&dataConnectionService{}, "getDataConnectionsByConfigMD5",
				func(dcService *dataConnectionService, ctx context.Context, configMD5 string) (connMap map[string]*interfaces.DataConnection, err error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			dcp.EXPECT().ValidateWhenCreate(gomock.Any(), gomock.Any()).Return(nil)
			dcp.EXPECT().ComputeConfigMD5(gomock.Any(), gomock.Any()).Return("pre", nil)
			dcp.EXPECT().GenerateAuthInfoAndStatus(gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin().WillReturnError(nil)
			smock.ExpectCommit().WillReturnError(nil)
			dca.EXPECT().CreateDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dca.EXPECT().CreateDataConnectionStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			_, err := dcs.CreateDataConnection(testENCtx, &reqConn)
			So(err, ShouldBeNil)
		})
	})
}

func Test_DataConnectionService_DeleteDataConnections(t *testing.T) {
	Convey("Test DeleteDataConnections", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dca := dmock.NewMockDataConnectionAccess(mockCtrl)
		dcs, smock := MockNewDataConnectionService(appSetting, dca)

		Convey("Delete failed, because the transaction begin failed", func() {
			connIDs := []string{"1", "2"}
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_BeginTransactionFailed).WithErrorDetails(expectedErr.Error())

			smock.ExpectBegin().WillReturnError(expectedErr)

			err := dcs.DeleteDataConnections(testENCtx, connIDs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Delete failed, caused by the error from method `DeleteDataConnections`", func() {
			connIDs := []string{"1", "2"}
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_DeleteDataConnectionsFailed).WithErrorDetails(expectedErr.Error())

			smock.ExpectBegin().WillReturnError(nil)
			smock.ExpectRollback().WillReturnError(expectedErr)

			dca.EXPECT().DeleteDataConnections(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)

			err := dcs.DeleteDataConnections(testENCtx, connIDs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Delete failed, caused by the error from method `DeleteDataConnectionStatuses`", func() {
			connIDs := []string{"1", "2"}
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_DeleteDataConnectionStatusesFailed).WithErrorDetails(expectedErr.Error())

			smock.ExpectBegin().WillReturnError(nil)
			smock.ExpectRollback().WillReturnError(expectedErr)

			dca.EXPECT().DeleteDataConnections(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dca.EXPECT().DeleteDataConnectionStatuses(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)

			err := dcs.DeleteDataConnections(testENCtx, connIDs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Delete succeed, but transaction commit failed", func() {
			connIDs := []string{"1", "2"}
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_CommitTransactionFailed).WithErrorDetails(fmt.Sprintf("Commit transaction failed when deleting data connections, err: %v", expectedErr))

			smock.ExpectBegin().WillReturnError(nil)
			smock.ExpectCommit().WillReturnError(expectedErr)

			dca.EXPECT().DeleteDataConnections(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dca.EXPECT().DeleteDataConnectionStatuses(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := dcs.DeleteDataConnections(testENCtx, connIDs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Delete succeed, and transaction commit succeed", func() {
			connIDs := []string{"1", "2"}
			smock.ExpectBegin().WillReturnError(nil)
			smock.ExpectCommit().WillReturnError(nil)

			dca.EXPECT().DeleteDataConnections(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dca.EXPECT().DeleteDataConnectionStatuses(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := dcs.DeleteDataConnections(testENCtx, connIDs)
			So(err, ShouldBeNil)
		})
	})
}

func Test_DataConnectionService_UpdateDataConnection(t *testing.T) {
	Convey("Test UpdateDataConnection", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dca := dmock.NewMockDataConnectionAccess(mockCtrl)
		dcs, _ := MockNewDataConnectionService(appSetting, dca)
		dcp := dmock.NewMockDataConnectionProcessor(mockCtrl)

		Convey("Update failed, caused by the error from method `ValidateWhenUpdate`", func() {
			reqConn, preConn := interfaces.DataConnection{}, interfaces.DataConnection{}
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_UnmarshalDataFailed)
			patch := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch.Reset()

			dcp.EXPECT().ValidateWhenUpdate(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)

			err := dcs.UpdateDataConnection(testENCtx, &reqConn, &preConn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Update failed, caused by the error from method `ComputeConfigMD5`", func() {
			reqConn, preConn := interfaces.DataConnection{}, interfaces.DataConnection{}
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_UnmarshalDataFailed)
			patch := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch.Reset()

			dcp.EXPECT().ValidateWhenUpdate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dcp.EXPECT().ComputeConfigMD5(gomock.Any(), gomock.Any()).Return("", expectedErr)

			err := dcs.UpdateDataConnection(testENCtx, &reqConn, &preConn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Update failed, caused by the error from method `getDataConnectionsByConfigMD5`", func() {
			reqConn, preConn := interfaces.DataConnection{}, interfaces.DataConnection{}
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_UnmarshalDataFailed)
			patch1 := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&dataConnectionService{}, "getDataConnectionsByConfigMD5",
				func(dcService *dataConnectionService, ctx context.Context, configMD5 string) (connMap map[string]*interfaces.DataConnection, err error) {
					return nil, expectedErr
				},
			)
			defer patch2.Reset()

			dcp.EXPECT().ValidateWhenUpdate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dcp.EXPECT().ComputeConfigMD5(gomock.Any(), gomock.Any()).Return("pre", nil)

			err := dcs.UpdateDataConnection(testENCtx, &reqConn, &preConn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Update failed, caused by the same config md5 was found", func() {
			reqConn, preConn := interfaces.DataConnection{}, interfaces.DataConnection{}
			expectedConnMap := map[string]*interfaces.DataConnection{
				"1": {ID: "1"},
			}
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusBadRequest,
				derrors.DataModel_DataConnection_DuplicatedParameter_Config).
				WithErrorDetails(fmt.Sprintf("Same config whose conn_id is %v already exists in the database", expectedConnMap["1"].ID))
			patch1 := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&dataConnectionService{}, "getDataConnectionsByConfigMD5",
				func(dcService *dataConnectionService, ctx context.Context, configMD5 string) (connMap map[string]*interfaces.DataConnection, err error) {
					return expectedConnMap, nil
				},
			)
			defer patch2.Reset()

			dcp.EXPECT().ValidateWhenUpdate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dcp.EXPECT().ComputeConfigMD5(gomock.Any(), gomock.Any()).Return("pre", nil)

			err := dcs.UpdateDataConnection(testENCtx, &reqConn, &preConn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Update failed, caused by the error from method `GenerateAuthInfoAndStatus`", func() {
			reqConn, preConn := interfaces.DataConnection{}, interfaces.DataConnection{}
			expectedConnMap := map[string]interfaces.DataConnection{
				"1": {},
			}
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusBadRequest,
				derrors.DataModel_DataConnection_DuplicatedParameter_Config).
				WithErrorDetails(fmt.Sprintf("Same config whose conn_id is %v already exists in the database", expectedConnMap["0"].ID))
			patch1 := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&dataConnectionService{}, "getDataConnectionsByConfigMD5",
				func(dcService *dataConnectionService, ctx context.Context, configMD5 string) (connMap map[string]*interfaces.DataConnection, err error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			dcp.EXPECT().ValidateWhenUpdate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dcp.EXPECT().ComputeConfigMD5(gomock.Any(), gomock.Any()).Return("pre", nil)
			dcp.EXPECT().GenerateAuthInfoAndStatus(gomock.Any(), gomock.Any()).Return(expectedErr)

			err := dcs.UpdateDataConnection(testENCtx, &reqConn, &preConn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Update failed, caused by the error from method `updateDataConnectionAndStatus`", func() {
			reqConn, preConn := interfaces.DataConnection{}, interfaces.DataConnection{}
			expectedConnMap := map[string]interfaces.DataConnection{
				"1": {},
			}
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusBadRequest,
				derrors.DataModel_DataConnection_DuplicatedParameter_Config).
				WithErrorDetails(fmt.Sprintf("Same config whose conn_id is %v already exists in the database", expectedConnMap["0"].ID))
			patch1 := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&dataConnectionService{}, "getDataConnectionsByConfigMD5",
				func(dcService *dataConnectionService, ctx context.Context, configMD5 string) (connMap map[string]*interfaces.DataConnection, err error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&dataConnectionService{}, "updateDataConnectionAndStatus",
				func(dcService *dataConnectionService, ctx context.Context, conn interfaces.DataConnection) (err error) {
					return expectedErr
				},
			)
			defer patch3.Reset()

			dcp.EXPECT().ValidateWhenUpdate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dcp.EXPECT().ComputeConfigMD5(gomock.Any(), gomock.Any()).Return("pre", nil)
			dcp.EXPECT().GenerateAuthInfoAndStatus(gomock.Any(), gomock.Any()).Return(nil)

			err := dcs.UpdateDataConnection(testENCtx, &reqConn, &preConn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Update succeed", func() {
			reqConn, preConn := interfaces.DataConnection{}, interfaces.DataConnection{}
			patch1 := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&dataConnectionService{}, "getDataConnectionsByConfigMD5",
				func(dcService *dataConnectionService, ctx context.Context, configMD5 string) (connMap map[string]*interfaces.DataConnection, err error) {
					return nil, nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&dataConnectionService{}, "updateDataConnectionAndStatus",
				func(dcService *dataConnectionService, ctx context.Context, conn interfaces.DataConnection) (err error) {
					return nil
				},
			)
			defer patch3.Reset()

			dcp.EXPECT().ValidateWhenUpdate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dcp.EXPECT().ComputeConfigMD5(gomock.Any(), gomock.Any()).Return("pre", nil)
			dcp.EXPECT().GenerateAuthInfoAndStatus(gomock.Any(), gomock.Any()).Return(nil)

			err := dcs.UpdateDataConnection(testENCtx, &reqConn, &preConn)
			So(err, ShouldBeNil)
		})
	})
}

func Test_DataConnectionService_GetDataConnection(t *testing.T) {
	Convey("Test GetDataConnection", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dca := dmock.NewMockDataConnectionAccess(mockCtrl)
		dcs, _ := MockNewDataConnectionService(appSetting, dca)
		dcp := dmock.NewMockDataConnectionProcessor(mockCtrl)

		Convey("Get failed, caused by the error from dca method `GetDataConnection`", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_GetDataConnectionsFailed).WithErrorDetails(expectedErr.Error())

			dca.EXPECT().GetDataConnection(gomock.Any(), gomock.Any()).Return(&interfaces.DataConnection{}, false, expectedErr)

			_, _, err := dcs.GetDataConnection(testENCtx, "1", true)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Get failed, because the conn does not found", func() {
			dca.EXPECT().GetDataConnection(gomock.Any(), gomock.Any()).Return(&interfaces.DataConnection{}, false, nil)

			_, isExist, err := dcs.GetDataConnection(testENCtx, "1", true)
			So(isExist, ShouldBeFalse)
			So(err, ShouldBeNil)
		})

		Convey("Get failed, caused by the error from func `updateDataConnectionAndStatus`", func() {
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_UnmarshalDataFailed)
			dca.EXPECT().GetDataConnection(gomock.Any(), gomock.Any()).Return(&interfaces.DataConnection{}, true, nil)
			dcp.EXPECT().UpdateAuthInfoAndStatus(gomock.Any(), gomock.Any()).Return(true, nil)

			patch1 := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&dataConnectionService{}, "updateDataConnectionAndStatus",
				func(ctx context.Context, conn interfaces.DataConnection) (err error) {
					return expectedErr
				},
			)
			defer patch2.Reset()

			_, _, err := dcs.GetDataConnection(testENCtx, "1", true)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by the error from method `HideAuthInfo`", func() {
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_UnmarshalDataFailed)
			dca.EXPECT().GetDataConnection(gomock.Any(), gomock.Any()).Return(&interfaces.DataConnection{}, true, nil)
			dcp.EXPECT().UpdateAuthInfoAndStatus(gomock.Any(), gomock.Any()).Return(true, nil)
			dcp.EXPECT().HideAuthInfo(gomock.Any(), gomock.Any()).Return(expectedErr)

			patch1 := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&dataConnectionService{}, "updateDataConnectionAndStatus",
				func(ctx context.Context, conn interfaces.DataConnection) (err error) {
					return nil
				},
			)
			defer patch2.Reset()

			_, _, err := dcs.GetDataConnection(testENCtx, "1", false)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed", func() {
			expectedErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_UnmarshalDataFailed)
			dca.EXPECT().GetDataConnection(gomock.Any(), gomock.Any()).Return(&interfaces.DataConnection{}, true, nil)
			dcp.EXPECT().UpdateAuthInfoAndStatus(gomock.Any(), gomock.Any()).Return(true, nil)
			dcp.EXPECT().HideAuthInfo(gomock.Any(), gomock.Any()).Return(expectedErr)

			patch1 := ApplyFuncReturn(data_source.NewDataConnectionProcessor, dcp, nil)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&dataConnectionService{}, "updateDataConnectionAndStatus",
				func(ctx context.Context, conn interfaces.DataConnection) (err error) {
					return nil
				},
			)
			defer patch2.Reset()

			_, _, err := dcs.GetDataConnection(testENCtx, "1", false)
			So(err, ShouldResemble, expectedErr)
		})
	})
}

func Test_DataConnectionService_ListDataConnections(t *testing.T) {
	Convey("Test ListDataConnections", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dca := dmock.NewMockDataConnectionAccess(mockCtrl)
		dcs, _ := MockNewDataConnectionService(appSetting, dca)

		Convey("List failed, caused by the error from dca method `ListDataConnections`", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_ListDataConnectionsFailed).WithErrorDetails(expectedErr.Error())
			queryParams := interfaces.DataConnectionListQueryParams{}

			dca.EXPECT().ListDataConnections(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			_, _, err := dcs.ListDataConnections(testENCtx, queryParams)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("List failed, caused by the error from tma method `GetDataConnectionTotal`", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_GetDataConnectionTotalFailed).WithErrorDetails(expectedErr.Error())
			queryParams := interfaces.DataConnectionListQueryParams{}

			dca.EXPECT().ListDataConnections(gomock.Any(), gomock.Any()).Return(nil, nil)
			dca.EXPECT().GetDataConnectionTotal(gomock.Any(), gomock.Any()).Return(0, expectedErr)

			_, _, err := dcs.ListDataConnections(testENCtx, queryParams)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("List succeed", func() {
			queryParams := interfaces.DataConnectionListQueryParams{}

			dca.EXPECT().ListDataConnections(gomock.Any(), gomock.Any()).Return(nil, nil)
			dca.EXPECT().GetDataConnectionTotal(gomock.Any(), gomock.Any()).Return(0, nil)

			_, _, err := dcs.ListDataConnections(testENCtx, queryParams)
			So(err, ShouldBeNil)
		})
	})
}

func Test_DataConnectionService_GetMapAboutName2ID(t *testing.T) {
	Convey("Test GetMapAboutName2ID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dca := dmock.NewMockDataConnectionAccess(mockCtrl)
		dcs, _ := MockNewDataConnectionService(appSetting, dca)

		Convey("Get failed, caused by the error from dca method 'GetMapAboutName2ID'", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError, derrors.DataModel_DataConnection_InternalError_GetMapAboutName2IDFailed).
				WithErrorDetails(expectedErr.Error())

			dca.EXPECT().GetMapAboutName2ID(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			connNames := []string{"conn1"}
			_, err := dcs.GetMapAboutName2ID(testENCtx, connNames)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Get succeed", func() {
			expectedConnMap := map[string]string{}
			dca.EXPECT().GetMapAboutName2ID(gomock.Any(), gomock.Any()).Return(expectedConnMap, nil)

			connNames := []string{"conn1"}
			_, err := dcs.GetMapAboutName2ID(testENCtx, connNames)
			So(err, ShouldBeNil)
		})
	})
}

func Test_DataConnectionService_GetMapAboutID2Name(t *testing.T) {
	Convey("Test GetMapAboutID2Name", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dca := dmock.NewMockDataConnectionAccess(mockCtrl)
		dcs, _ := MockNewDataConnectionService(appSetting, dca)

		Convey("Get failed, caused by the error from dca method 'GetMapAboutID2Name'", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError, derrors.DataModel_DataConnection_InternalError_GetMapAboutID2NameFailed).
				WithErrorDetails(expectedErr.Error())

			dca.EXPECT().GetMapAboutID2Name(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			connIDs := []string{"1"}
			_, err := dcs.GetMapAboutID2Name(testENCtx, connIDs)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Get succeed", func() {
			expectedConnMap := map[string]string{}
			dca.EXPECT().GetMapAboutID2Name(gomock.Any(), gomock.Any()).Return(expectedConnMap, nil)

			connIDs := []string{"1"}
			_, err := dcs.GetMapAboutID2Name(testENCtx, connIDs)
			So(err, ShouldBeNil)
		})
	})
}

func Test_DataConnectionService_GetDataConnectionSourceType(t *testing.T) {
	Convey("Test GetDataConnectionSourceType", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dca := dmock.NewMockDataConnectionAccess(mockCtrl)
		dcs, _ := MockNewDataConnectionService(appSetting, dca)

		Convey("Get failed, caused by the error from dca method 'GetDataConnectionSourceType'", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError, derrors.DataModel_DataConnection_InternalError_GetDataConnectionSourceTypeFailed).
				WithErrorDetails(expectedErr.Error())

			dca.EXPECT().GetDataConnectionSourceType(gomock.Any(), gomock.Any()).Return("", false, expectedErr)

			connID := "1"
			_, _, err := dcs.GetDataConnectionSourceType(testENCtx, connID)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Get succeed", func() {
			dca.EXPECT().GetDataConnectionSourceType(gomock.Any(), gomock.Any()).Return("", true, nil)

			connID := "1"
			_, _, err := dcs.GetDataConnectionSourceType(testENCtx, connID)
			So(err, ShouldBeNil)
		})
	})
}

/*
	私有方法
*/

func Test_DataConnectionService_UpdateDataConnectionAndStatus(t *testing.T) {
	Convey("Test UpdateDataConnectionAndStatus", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dca := dmock.NewMockDataConnectionAccess(mockCtrl)
		dcs, smock := MockNewDataConnectionService(appSetting, dca)

		Convey("Update failed, because the transaction begin failed", func() {
			reqConn := interfaces.DataConnection{}
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_BeginTransactionFailed).WithErrorDetails(fmt.Sprintf("Begin transaction failed when updating data connection config and status in database, err: %v", expectedErr.Error()))

			smock.ExpectBegin().WillReturnError(expectedErr)

			err := dcs.updateDataConnectionAndStatus(testENCtx, &reqConn)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Update failed, caused by the error from method `UpdateDataConnection`", func() {
			reqConn := interfaces.DataConnection{}
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_UpdateDataConnectionFailed).WithErrorDetails(expectedErr.Error())

			smock.ExpectBegin().WillReturnError(nil)
			smock.ExpectRollback().WillReturnError(expectedErr)

			dca.EXPECT().UpdateDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)

			err := dcs.updateDataConnectionAndStatus(testENCtx, &reqConn)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Update failed, caused by the error from method `UpdateDataConnectionStatus`", func() {
			reqConn := interfaces.DataConnection{}
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_UpdateDataConnectionStatusFailed).WithErrorDetails(expectedErr.Error())

			smock.ExpectBegin().WillReturnError(nil)
			smock.ExpectRollback().WillReturnError(expectedErr)

			dca.EXPECT().UpdateDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dca.EXPECT().UpdateDataConnectionStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)

			err := dcs.updateDataConnectionAndStatus(testENCtx, &reqConn)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Update succeed, but transaction commit failed", func() {
			reqConn := interfaces.DataConnection{}
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_CommitTransactionFailed).WithErrorDetails(fmt.Sprintf("Commit transaction failed when updating data connection config and status in database, err: %v", expectedErr))

			smock.ExpectBegin().WillReturnError(nil)
			smock.ExpectCommit().WillReturnError(expectedErr)

			dca.EXPECT().UpdateDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dca.EXPECT().UpdateDataConnectionStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := dcs.updateDataConnectionAndStatus(testENCtx, &reqConn)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Update succeed, and transaction commit succeed", func() {
			reqConn := interfaces.DataConnection{}

			smock.ExpectBegin().WillReturnError(nil)
			smock.ExpectCommit().WillReturnError(nil)

			dca.EXPECT().UpdateDataConnection(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			dca.EXPECT().UpdateDataConnectionStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := dcs.updateDataConnectionAndStatus(testENCtx, &reqConn)
			So(err, ShouldBeNil)
		})
	})
}

func Test_DataConnectionService_GetDataConnectionsByConfigMD5(t *testing.T) {
	Convey("Test GetDataConnectionsByConfigMD5", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		dca := dmock.NewMockDataConnectionAccess(mockCtrl)
		dcs, _ := MockNewDataConnectionService(appSetting, dca)

		Convey("Get failed, caused by the error from dca method 'GetDataConnectionsByConfigMD5'", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testENCtx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_GetDataConnectionsFailed).WithErrorDetails(expectedErr.Error())

			dca.EXPECT().GetDataConnectionsByConfigMD5(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			_, err := dcs.getDataConnectionsByConfigMD5(testENCtx, "")
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Get succeed", func() {
			dca.EXPECT().GetDataConnectionsByConfigMD5(gomock.Any(), gomock.Any()).Return(nil, nil)

			_, err := dcs.getDataConnectionsByConfigMD5(testENCtx, "")
			So(err, ShouldBeNil)
		})
	})
}
