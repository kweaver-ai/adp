// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package anyrobot

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
)

var (
	arcProcessorOnce sync.Once
	arcProcessor     interfaces.DataConnectionProcessor
)

// AnyRobot客户端配置
type anyrobotClientConfig struct {
	Address     string `json:"address"`
	Protocol    string `json:"protocol"`
	AppID       string `json:"api_key"`
	AppSecret   string `json:"secret_key,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
}

type anyrobotConnectionProcessor struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewAnyRobotConnectionProcessor(appSetting *common.AppSetting) interfaces.DataConnectionProcessor {
	arcProcessorOnce.Do(func() {
		arcProcessor = &anyrobotConnectionProcessor{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
		}
	})
	return arcProcessor
}

func (arcp *anyrobotConnectionProcessor) ValidateWhenCreate(ctx context.Context, conn *interfaces.DataConnection) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 处理待创建的数据连接")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. detailedConfig由any转为anyrobotClientConfig
	arConf, err := arcp.any2AnyRobotClientConfig(ctx, conn.DataSourceConfig)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	// 2. 校验anyrobotClientConfig
	err = arcp.commonValidateWhenCreateAndUpdate(ctx, arConf, false)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	// 3. secret_key加密
	arConf.AppSecret = common.EncryptPassword(arConf.AppSecret)

	conn.DataSourceConfig = arConf
	return nil
}

func (arcp *anyrobotConnectionProcessor) ValidateWhenUpdate(ctx context.Context,
	conn *interfaces.DataConnection, preConn *interfaces.DataConnection) (err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 处理待修改的数据连接")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. detailedConfig由any转为anyrobotClientConfig
	arConf, err := arcp.any2AnyRobotClientConfig(ctx, conn.DataSourceConfig)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	preConf, err := arcp.any2AnyRobotClientConfig(ctx, preConn.DataSourceConfig)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	// 2. 校验anyrobotClientConfig
	err = arcp.commonValidateWhenCreateAndUpdate(ctx, arConf, true)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	// 3. 如果传入了新的secret_key, 则需要加密
	if arConf.AppSecret == "" {
		arConf.AppSecret = preConf.AppSecret
	} else {
		arConf.AppSecret = common.EncryptPassword(arConf.AppSecret)
	}

	conn.DataSourceConfig = arConf
	return nil
}

func (arcp *anyrobotConnectionProcessor) ComputeConfigMD5(ctx context.Context,
	conn *interfaces.DataConnection) (md5 string, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 计算详细配置的md5")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. detailedConfig由any转为anyrobotClientConfig
	arConf, err := arcp.any2AnyRobotClientConfig(ctx, conn.DataSourceConfig)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return "", err
	}

	// Decrypt AppSecret
	appSecret := common.DecryptPassword(arConf.AppSecret)

	// 2. 生成config的md5
	str := arConf.Address + arConf.Protocol + arConf.AppID + appSecret
	return common.MD532Lower(str), nil
}

func (arcp *anyrobotConnectionProcessor) GenerateAuthInfoAndStatus(ctx context.Context,
	conn *interfaces.DataConnection) (err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 生成auth_info和连接状态")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. detailedConfig由any转为anyrobotClientConfig
	arConf, err := arcp.any2AnyRobotClientConfig(ctx, conn.DataSourceConfig)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	// 2. 获取access_token, 并校验连通性
	err = arcp.ping(ctx, arConf)
	if err != nil {
		errDetails := fmt.Sprintf("Verify the connectivity of anyrobot failed, err: %v", err.Error())
		logger.Error(errDetails)
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataConnection_InternalError_VerifyConnectivityFailed).WithErrorDetails(errDetails)
	}

	// 3. 更新conn和status, 并返回结果
	conn.DataConnectionStatus.Status = "ok"
	conn.DataConnectionStatus.DetectionTime = time.Now().UnixMilli()

	conn.DataSourceConfig = arConf
	return nil
}

func (arcp *anyrobotConnectionProcessor) UpdateAuthInfoAndStatus(ctx context.Context, conn *interfaces.DataConnection) (needWriteBack bool, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 更新auth_info和连接状态")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. []byte转anyrobotClientConfig
	arConf, err := arcp.any2AnyRobotClientConfig(ctx, conn.DataSourceConfig)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return false, err
	}

	err = arcp.ping(ctx, arConf)
	if err != nil {
		conn.DataConnectionStatus.Status = "error"
		errDetails := fmt.Sprintf("Verify the connectivity of anyrobot failed, err: %v", err.Error())
		logger.Error(errDetails)
		return false, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataConnection_InternalError_VerifyConnectivityFailed).WithErrorDetails(errDetails)
	}

	conn.DataSourceConfig = arConf
	return false, nil
}

func (arcp *anyrobotConnectionProcessor) HideAuthInfo(ctx context.Context, conn *interfaces.DataConnection) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 隐藏auth_info")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	arConf, err := arcp.any2AnyRobotClientConfig(ctx, conn.DataSourceConfig)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	arConf.AppSecret = ""
	conn.DataSourceConfig = arConf

	return nil
}

func (arcp *anyrobotConnectionProcessor) any2AnyRobotClientConfig(ctx context.Context, conf any) (*anyrobotClientConfig, error) {
	switch t := conf.(type) {
	case anyrobotClientConfig:
		return conf.(*anyrobotClientConfig), nil
	case []byte:
		arConf := anyrobotClientConfig{}
		err := sonic.Unmarshal(conf.([]byte), &arConf)
		if err != nil {
			errDetails := fmt.Sprintf("Field config cannot be unmarshaled to anyrobotClientConfig, err: %v", err.Error())
			logger.Error(errDetails)
			return &arConf, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_UnmarshalDataFailed).WithErrorDetails(errDetails)
		}
		return &arConf, nil
	default:
		b, err := sonic.Marshal(conf)
		if err != nil {
			errDetails := fmt.Sprintf("Marshal field config with field type %v failed, err: %v", t, err.Error())
			logger.Error(errDetails)
			return &anyrobotClientConfig{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_MarshalDataFailed).WithErrorDetails(errDetails)
		}

		arConf := anyrobotClientConfig{}
		err = sonic.Unmarshal(b, &arConf)
		if err != nil {
			errDetails := fmt.Sprintf("Field config with field type %v cannot be unmarshaled to anyrobotClientConfig, err: %v", t, err.Error())
			logger.Error(errDetails)
			return &arConf, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_UnmarshalDataFailed).WithErrorDetails(errDetails)
		}
		return &arConf, nil
	}
}

func (arcp *anyrobotConnectionProcessor) commonValidateWhenCreateAndUpdate(ctx context.Context,
	arConf *anyrobotClientConfig, isUpdate bool) error {

	// 2.1 校验address
	if arConf.Address == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataConnection_NullParameter_Address)
	}

	// 2.2 校验protocol
	if arConf.Protocol != "http" && arConf.Protocol != "https" {
		errDetails := fmt.Sprintf("The protocol %v is invalid, valid protocol is http or https", arConf.Protocol)
		logger.Error(errDetails)
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataConnection_InvalidParameter_Protocol).WithErrorDetails(errDetails)
	}

	// 2.3 校验api_key
	if arConf.AppID == "" {
		errDetails := "The api_key is null"
		logger.Error(errDetails)
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataConnection_NullParameter_ApiKey).WithErrorDetails(errDetails)
	}

	// 2.4 校验secret_key
	if !isUpdate && arConf.AppSecret == "" {
		errDetails := "The secret_key is null"
		logger.Error(errDetails)
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataConnection_NullParameter_SecretKey).WithErrorDetails(errDetails)
	}

	return nil
}

func (arcp *anyrobotConnectionProcessor) ping(ctx context.Context, arConf *anyrobotClientConfig) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: ping", trace.WithSpanKind(trace.SpanKindClient))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// Decrypt AppSecret
	appSecret := common.DecryptPassword(arConf.AppSecret)

	// 1. 拼接url
	currentTime := time.Now().UnixMilli()
	str := fmt.Sprintf("app_id=%s&app_secret=%s&timestamp=%v", arConf.AppID, appSecret, currentTime)
	url := fmt.Sprintf("%s://%s/openapi/developer/%s/token?auth=%s&timestamp=%v",
		arConf.Protocol, arConf.Address, arConf.AppID, common.MD532Lower(str), currentTime)

	span.SetAttributes(attr.Key("anyrobot_url").String(url))

	respCode, respBody, err := arcp.httpClient.GetNoUnmarshal(ctx, url, nil, nil)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to ping anyrobot system: %s", err)
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	if respCode != http.StatusOK {
		errDetails := fmt.Sprintf("Failed to ping anyrobot system: %s", string(respBody))
		err := errors.New(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	arConf.AccessToken = string(respBody)
	return nil
}
