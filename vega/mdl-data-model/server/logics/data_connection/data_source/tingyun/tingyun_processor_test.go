// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package tingyun

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	sonic_ast "github.com/bytedance/sonic/ast"
	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

func MockNewTingYunConnectionProcessor(appSetting *common.AppSetting,
	httpClient rest.HTTPClient) *tingYunConnectionProcessor {
	return &tingYunConnectionProcessor{
		appSetting: appSetting,
		httpClient: httpClient,
	}
}

func Test_TingYunConnectionProcessor_ValidateWhenCreate(t *testing.T) {
	Convey("Test ValidateWhenCreate", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		tycp := MockNewTingYunConnectionProcessor(appSetting, httpClient)

		Convey("Validate failed, caused by the error from method `any2TingYunDetailedConfig`", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{}, expectedErr
				},
			)
			defer patch.Reset()

			err := tycp.ValidateWhenCreate(testCtx, &interfaces.DataConnection{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Validate failed, caused by the error from method `commonValidateWhenCreateAndUpdate`", func() {
			expectedErr := errors.New("some errors")
			patch1 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{}, nil
				},
			)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "commonValidateWhenCreateAndUpdate",
				func(t *tingYunConnectionProcessor, ctx context.Context, conf tingYunDetailedConfig, isUpdate bool) error {
					return expectedErr
				},
			)
			defer patch2.Reset()

			err := tycp.ValidateWhenCreate(testCtx, &interfaces.DataConnection{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Validate succeed", func() {
			patch1 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{}, nil
				},
			)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "commonValidateWhenCreateAndUpdate",
				func(t *tingYunConnectionProcessor, ctx context.Context, conf tingYunDetailedConfig, isUpdate bool) error {
					return nil
				},
			)
			defer patch2.Reset()

			conn := interfaces.DataConnection{}
			err := tycp.ValidateWhenCreate(testCtx, &conn)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TingYunConnectionProcessor_ValidateWhenUpdate(t *testing.T) {
	Convey("Test ValidateWhenUpdate", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		tycp := MockNewTingYunConnectionProcessor(appSetting, httpClient)

		Convey("Validate failed, caused by the error from method `any2TingYunDetailedConfig` when first used", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{}, expectedErr
				},
			)
			defer patch.Reset()

			err := tycp.ValidateWhenUpdate(testCtx, &interfaces.DataConnection{}, &interfaces.DataConnection{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Validate failed, caused by the error from method `any2TingYunDetailedConfig` when second used", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					if _, ok := i.(tingYunDetailedConfig); ok {
						return &tingYunDetailedConfig{}, nil
					} else {
						return &tingYunDetailedConfig{}, expectedErr
					}
				},
			)
			defer patch.Reset()

			conn := interfaces.DataConnection{
				DataSourceConfig: tingYunDetailedConfig{},
			}
			preConn := interfaces.DataConnection{}
			err := tycp.ValidateWhenUpdate(testCtx, &conn, &preConn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Validate failed, caused by the error from method `commonValidateWhenCreateAndUpdate`", func() {
			expectedErr := errors.New("some errors")
			patch1 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{}, nil
				},
			)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "commonValidateWhenCreateAndUpdate",
				func(t *tingYunConnectionProcessor, ctx context.Context, conf tingYunDetailedConfig, isUpdate bool) error {
					return expectedErr
				},
			)
			defer patch2.Reset()

			conn := interfaces.DataConnection{}
			preConn := interfaces.DataConnection{}
			err := tycp.ValidateWhenUpdate(testCtx, &conn, &preConn)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Validate succeed", func() {
			patch1 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{}, nil
				},
			)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "commonValidateWhenCreateAndUpdate",
				func(t *tingYunConnectionProcessor, ctx context.Context, conf tingYunDetailedConfig, isUpdate bool) error {
					return nil
				},
			)
			defer patch2.Reset()

			conn := interfaces.DataConnection{}
			preConn := interfaces.DataConnection{}
			err := tycp.ValidateWhenUpdate(testCtx, &conn, &preConn)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TingYunConnectionProcessor_ComputeConfigMD5(t *testing.T) {
	Convey("Test ComputeConfigMD5", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		tycp := MockNewTingYunConnectionProcessor(appSetting, httpClient)

		Convey("Generate failed, caused by the error from method `any2TingYunDetailedConfig`", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{}, expectedErr
				},
			)
			defer patch.Reset()

			_, err := tycp.ComputeConfigMD5(testCtx, &interfaces.DataConnection{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Generate succeed", func() {
			patch := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{
						SecretKey: common.EncryptPassword("password"),
					}, nil
				},
			)
			defer patch.Reset()

			_, err := tycp.ComputeConfigMD5(testCtx, &interfaces.DataConnection{})
			So(err, ShouldBeNil)
		})
	})
}

func Test_TingYunConnectionProcessor_GenerateAuthInfoAndStatus(t *testing.T) {
	Convey("Test GenerateAuthInfoAndStatus", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		tycp := MockNewTingYunConnectionProcessor(appSetting, httpClient)

		Convey("Generate failed, caused by the error from method `any2TingYunDetailedConfig`", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{}, expectedErr
				},
			)
			defer patch.Reset()

			err := tycp.GenerateAuthInfoAndStatus(testCtx, &interfaces.DataConnection{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Generate failed, caused by the error from method `getAccessToken`", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_GetAccessTokenFailed).WithErrorDetails(fmt.Sprintf("Get access token of tingyun failed, err: %v", expectedErr.Error()))

			patch1 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{
						SecretKey: common.EncryptPassword("password"),
					}, nil
				},
			)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "getAccessToken",
				func(t *tingYunConnectionProcessor, ctx context.Context, cfg TingYunClientConfig) (accessToken string, err error) {
					return "", expectedErr
				},
			)
			defer patch2.Reset()

			err := tycp.GenerateAuthInfoAndStatus(testCtx, &interfaces.DataConnection{})
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Generate failed, caused by the error from method `ping`", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_VerifyConnectivityFailed).WithErrorDetails(fmt.Sprintf("Verify the connectivity of tingyun failed, err: %v", expectedErr.Error()))

			patch1 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{
						SecretKey: common.EncryptPassword("password"),
					}, nil
				},
			)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "getAccessToken",
				func(t *tingYunConnectionProcessor, ctx context.Context, cfg TingYunClientConfig) (accessToken string, err error) {
					return "", nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "ping",
				func(t *tingYunConnectionProcessor, ctx context.Context, cfg TingYunClientConfig) (err error) {
					return expectedErr
				},
			)
			defer patch3.Reset()

			err := tycp.GenerateAuthInfoAndStatus(testCtx, &interfaces.DataConnection{})
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Generate succeed", func() {
			patch1 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{
						SecretKey: common.EncryptPassword("password"),
					}, nil
				},
			)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "getAccessToken",
				func(t *tingYunConnectionProcessor, ctx context.Context, cfg TingYunClientConfig) (accessToken string, err error) {
					return "", nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "ping",
				func(t *tingYunConnectionProcessor, ctx context.Context, cfg TingYunClientConfig) (err error) {
					return nil
				},
			)
			defer patch3.Reset()

			err := tycp.GenerateAuthInfoAndStatus(testCtx, &interfaces.DataConnection{})
			So(err, ShouldBeNil)
		})
	})
}

func Test_TingYunConnectionProcessor_UpdateAuthInfoAndStatus(t *testing.T) {
	Convey("Test UpdateAuthInfoAndStatus", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		tycp := MockNewTingYunConnectionProcessor(appSetting, httpClient)

		Convey("Update failed, caused by the error from method `any2TingYunDetailedConfig`", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{}, expectedErr
				},
			)
			defer patch.Reset()

			_, err := tycp.UpdateAuthInfoAndStatus(testCtx, &interfaces.DataConnection{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Update failed, caused by the error from method `getAccessToken`", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_GetAccessTokenFailed).WithErrorDetails(fmt.Sprintf("Get new access token of tingyun failed, err: %v", expectedErr.Error()))

			patch1 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{
						SecretKey: common.EncryptPassword("password"),
					}, nil
				},
			)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "getAccessToken",
				func(t *tingYunConnectionProcessor, ctx context.Context, cfg TingYunClientConfig) (accessToken string, err error) {
					return "", expectedErr
				},
			)
			defer patch2.Reset()

			_, err := tycp.UpdateAuthInfoAndStatus(testCtx, &interfaces.DataConnection{})
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Update failed, caused by the error from method `ping`", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_VerifyConnectivityFailed).WithErrorDetails(fmt.Sprintf("Verify the connectivity of tingyun failed, err: %v", expectedErr.Error()))

			patch1 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{
						SecretKey: common.EncryptPassword("password"),
					}, nil
				},
			)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "getAccessToken",
				func(t *tingYunConnectionProcessor, ctx context.Context, cfg TingYunClientConfig) (accessToken string, err error) {
					return "", nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "ping",
				func(t *tingYunConnectionProcessor, ctx context.Context, cfg TingYunClientConfig) (err error) {
					return expectedErr
				},
			)
			defer patch3.Reset()

			_, err := tycp.UpdateAuthInfoAndStatus(testCtx, &interfaces.DataConnection{})
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Update succeed", func() {
			patch1 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{
						SecretKey: common.EncryptPassword("password"),
					}, nil
				},
			)
			defer patch1.Reset()

			patch2 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "getAccessToken",
				func(t *tingYunConnectionProcessor, ctx context.Context, cfg TingYunClientConfig) (accessToken string, err error) {
					return "", nil
				},
			)
			defer patch2.Reset()

			patch3 := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "ping",
				func(t *tingYunConnectionProcessor, ctx context.Context, cfg TingYunClientConfig) (err error) {
					return nil
				},
			)
			defer patch3.Reset()

			_, err := tycp.UpdateAuthInfoAndStatus(testCtx, &interfaces.DataConnection{})
			So(err, ShouldBeNil)
		})
	})
}

func Test_TingYunConnectionProcessor_HideAuthInfo(t *testing.T) {
	Convey("Test HideAuthInfo", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		tycp := MockNewTingYunConnectionProcessor(appSetting, httpClient)

		Convey("Hide failed, caused by the error from method `any2TingYunDetailedConfig`", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{}, expectedErr
				},
			)
			defer patch.Reset()

			err := tycp.HideAuthInfo(testCtx, &interfaces.DataConnection{})
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Hide succeed", func() {
			patch := ApplyPrivateMethod(&tingYunConnectionProcessor{}, "any2TingYunDetailedConfig",
				func(t *tingYunConnectionProcessor, ctx context.Context, i any) (*tingYunDetailedConfig, error) {
					return &tingYunDetailedConfig{}, nil
				},
			)
			defer patch.Reset()

			err := tycp.HideAuthInfo(testCtx, &interfaces.DataConnection{})
			So(err, ShouldBeNil)
		})
	})
}

func Test_TingYunConnectionProcessor_Any2TingYunDetailedConfig(t *testing.T) {
	Convey("Test Any2TingYunDetailedConfig", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		tycp := MockNewTingYunConnectionProcessor(appSetting, httpClient)

		Convey("Convert succeed, dynamic type is tingYunDetailedConfig", func() {
			_, err := tycp.any2TingYunDetailedConfig(testCtx, &tingYunDetailedConfig{})
			So(err, ShouldBeNil)
		})

		Convey("Convert failed, caused by the error from func `sonic.Unmarshal`, dynamic type is []byte", func() {
			expectedErr := errors.New("some errors")
			errDetails := fmt.Sprintf("Field config cannot be unmarshaled to tingYunDetailedConfig, err: %v", expectedErr.Error())
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_UnmarshalDataFailed).WithErrorDetails(errDetails)

			patch := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch.Reset()

			_, err := tycp.any2TingYunDetailedConfig(testCtx, []byte(nil))
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Convert succeed, dynamic type is []byte", func() {
			patch := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patch.Reset()

			_, err := tycp.any2TingYunDetailedConfig(testCtx, []byte(nil))
			So(err, ShouldBeNil)
		})

		Convey("Convert failed, caused by the error from func `sonic.Marshal`, dynamic type is other", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_MarshalDataFailed).WithErrorDetails(fmt.Sprintf("Marshal field config with field type 1 failed, err: %v", expectedErr.Error()))

			patch := ApplyFuncReturn(sonic.Marshal, nil, expectedErr)
			defer patch.Reset()

			_, err := tycp.any2TingYunDetailedConfig(testCtx, 1)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Convert failed, caused by the error from func `sonic.Unmarshal`, dynamic type is other", func() {
			expectedErr := errors.New("some errors")
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_UnmarshalDataFailed).WithErrorDetails(fmt.Sprintf("Field config with field type 1 cannot be unmarshaled to tingYunDetailedConfig, err: %v", expectedErr.Error()))

			patch1 := ApplyFuncReturn(sonic.Marshal, nil, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(sonic.Unmarshal, expectedErr)
			defer patch2.Reset()

			_, err := tycp.any2TingYunDetailedConfig(testCtx, 1)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Convert succeed, dynamic type is other", func() {
			patch1 := ApplyFuncReturn(sonic.Marshal, nil, nil)
			defer patch1.Reset()

			patch2 := ApplyFuncReturn(sonic.Unmarshal, nil)
			defer patch2.Reset()

			_, err := tycp.any2TingYunDetailedConfig(testCtx, 1)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TingYunConnectionProcessor_CommonValidateBeforeCreateAndUpdate(t *testing.T) {
	Convey("Test CommonValidateBeforeCreateAndUpdate", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		tycp := MockNewTingYunConnectionProcessor(appSetting, httpClient)

		Convey("Validate failed, caused by the empty address", func() {
			tyConfig := tingYunDetailedConfig{}
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest, derrors.DataModel_DataConnection_NullParameter_Address)

			err := tycp.commonValidateWhenCreateAndUpdate(testCtx, &tyConfig, false)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Validate failed, caused by the invalid protocol", func() {
			tyConfig := tingYunDetailedConfig{
				Address: "http://127.0.0.1",
			}
			errDetails := fmt.Sprintf("The protocol %v is invalid, valid protocol is http or https", tyConfig.Protocol)
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest,
				derrors.DataModel_DataConnection_InvalidParameter_Protocol).WithErrorDetails(errDetails)

			err := tycp.commonValidateWhenCreateAndUpdate(testCtx, &tyConfig, false)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Validate failed, caused by the invalid api_key", func() {
			tyConfig := tingYunDetailedConfig{
				Address:  "127.0.0.1",
				Protocol: "http",
			}
			errDetails := "The api_key is null"
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest,
				derrors.DataModel_DataConnection_NullParameter_ApiKey).WithErrorDetails(errDetails)

			err := tycp.commonValidateWhenCreateAndUpdate(testCtx, &tyConfig, false)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Validate failed, caused by the invalid secret_key", func() {
			tyConfig := tingYunDetailedConfig{
				Address:  "127.0.0.1",
				Protocol: "http",
				ApiKey:   "xxx",
			}
			errDetails := "The secret_key is null"
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusBadRequest,
				derrors.DataModel_DataConnection_NullParameter_SecretKey).WithErrorDetails(errDetails)

			err := tycp.commonValidateWhenCreateAndUpdate(testCtx, &tyConfig, false)
			So(err, ShouldResemble, expectedHttpErr)
		})

		Convey("Validate succeed", func() {
			tyConfig := tingYunDetailedConfig{
				Address:   "127.0.0.1",
				Protocol:  "http",
				ApiKey:    "xxx",
				SecretKey: "key",
			}

			err := tycp.commonValidateWhenCreateAndUpdate(testCtx, &tyConfig, false)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TingYunConnectionProcessor_GetAccessToken(t *testing.T) {
	Convey("Test GetAccessToken", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		tycp := MockNewTingYunConnectionProcessor(appSetting, httpClient)

		cfg := TingYunClientConfig{}

		Convey("Get failed, caused by http error", func() {
			expectedErr := errors.New("some errors")
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusInternalServerError, nil, expectedErr)

			_, err := tycp.getAccessToken(testCtx, cfg)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by error from func `sonic.Get` when getting code", func() {
			expectedErr := errors.New("some errors")
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, []byte("{\"code\": 200}"), nil)

			patch := ApplyFuncReturn(sonic.Get, sonic_ast.Node{}, expectedErr)
			defer patch.Reset()

			_, err := tycp.getAccessToken(testCtx, cfg)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by error from func `ast.Node.Int64`", func() {
			expectedErr := errors.New("some errors")
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, []byte("{\"code\": 200}"), nil)

			node := sonic_ast.Node{}
			patch1 := ApplyFuncReturn(sonic.Get, node, nil)
			defer patch1.Reset()

			patch2 := ApplyMethodReturn(&sonic_ast.Node{}, "Int64", int64(1), expectedErr)
			defer patch2.Reset()

			_, err := tycp.getAccessToken(testCtx, cfg)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by respCode != 200", func() {
			expectedRespBody := []byte("some errors")
			expectedErr := fmt.Errorf("failed to get tingyun access token by http client: %s", string(expectedRespBody))
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusInternalServerError, expectedRespBody, nil)

			node := sonic_ast.Node{}
			patch1 := ApplyFuncReturn(sonic.Get, node, nil)
			defer patch1.Reset()

			patch2 := ApplyMethodReturn(&sonic_ast.Node{}, "Int64", int64(1), nil)
			defer patch2.Reset()

			_, err := tycp.getAccessToken(testCtx, cfg)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by error from func `sonic.Get` when getting access_token", func() {
			expectedErr := fmt.Errorf("some errors")
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, []byte(nil), nil)

			count := 0
			patch1 := ApplyFunc(sonic.Get, func(src []byte, path ...interface{}) (sonic_ast.Node, error) {
				if count > 0 {
					return sonic_ast.Node{}, expectedErr
				}
				count++
				return sonic_ast.Node{}, nil
			})
			defer patch1.Reset()

			patch2 := ApplyMethodReturn(&sonic_ast.Node{}, "Int64", int64(200), nil)
			defer patch2.Reset()

			_, err := tycp.getAccessToken(testCtx, cfg)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get failed, caused by error from func `ast.Node.String`", func() {
			expectedErr := fmt.Errorf("some errors")
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, []byte(nil), nil)

			patch1 := ApplyFuncReturn(sonic.Get, sonic_ast.Node{}, nil)
			defer patch1.Reset()

			patch2 := ApplyMethodReturn(&sonic_ast.Node{}, "Int64", int64(200), nil)
			defer patch2.Reset()

			patch3 := ApplyMethodReturn(&sonic_ast.Node{}, "String", "", expectedErr)
			defer patch3.Reset()

			_, err := tycp.getAccessToken(testCtx, cfg)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Get succeed", func() {
			expectedToken := "new access_token"
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, []byte(nil), nil)

			patch1 := ApplyFuncReturn(sonic.Get, sonic_ast.Node{}, nil)
			defer patch1.Reset()

			patch2 := ApplyMethodReturn(&sonic_ast.Node{}, "Int64", int64(200), nil)
			defer patch2.Reset()

			patch3 := ApplyMethodReturn(&sonic_ast.Node{}, "String", expectedToken, nil)
			defer patch3.Reset()

			newAccessToken, err := tycp.getAccessToken(testCtx, cfg)
			So(newAccessToken, ShouldEqual, expectedToken)
			So(err, ShouldBeNil)
		})
	})
}

func Test_TingYunConnectionProcessor_Ping(t *testing.T) {
	Convey("Test Ping", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		httpClient := rmock.NewMockHTTPClient(mockCtrl)
		tycp := MockNewTingYunConnectionProcessor(appSetting, httpClient)

		cfg := TingYunClientConfig{}

		Convey("Ping failed, caused by http error", func() {
			expectedErr := errors.New("some errors")
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, nil, expectedErr)

			err := tycp.ping(testCtx, cfg)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Ping failed, caused by status != 200", func() {
			expectedErr := fmt.Errorf("Failed to ping tingyun system: %s", "{}")
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusInternalServerError, []byte("{}"), nil)

			err := tycp.ping(testCtx, cfg)
			So(err, ShouldResemble, expectedErr)
		})

		Convey("Ping succeed", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				http.StatusOK, []byte("{}"), nil)

			err := tycp.ping(testCtx, cfg)
			So(err, ShouldBeNil)
		})
	})
}
