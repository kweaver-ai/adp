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
	tycProcessorOnce sync.Once
	tycProcessor     interfaces.DataConnectionProcessor
)

// 听云数据连接详细配置
type tingYunDetailedConfig struct {
	Address        string `json:"address"`
	Protocol       string `json:"protocol"`
	ApiKey         string `json:"api_key"`
	SecretKey      string `json:"secret_key,omitempty"`
	AccessToken    string `json:"access_token,omitempty"`
	ExpirationTime int64  `json:"expiration_time,omitempty"`
}

// 听云客户端配置
type TingYunClientConfig struct {
	Address     string
	Protocol    string
	ApiKey      string
	SecretKey   string `json:"-"`
	AccessToken string
}

type tingYunConnectionProcessor struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewTingYunConnectionProcessor(appSetting *common.AppSetting) interfaces.DataConnectionProcessor {
	tycProcessorOnce.Do(func() {
		tycProcessor = &tingYunConnectionProcessor{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
		}
	})
	return tycProcessor
}

func (tycp *tingYunConnectionProcessor) ValidateWhenCreate(ctx context.Context, conn *interfaces.DataConnection) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 处理待创建的数据连接")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. detailedConfig由any转为tingYunDetailedConfig
	conf, err := tycp.any2TingYunDetailedConfig(ctx, conn.DataSourceConfig)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	// 2. 校验tingYunDetailedConfig
	err = tycp.commonValidateWhenCreateAndUpdate(ctx, conf, false)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	// 3. secret_key加密
	conf.SecretKey = common.EncryptPassword(conf.SecretKey)

	conn.DataSourceConfig = conf
	return nil
}

func (tycp *tingYunConnectionProcessor) ValidateWhenUpdate(ctx context.Context, conn *interfaces.DataConnection, preConn *interfaces.DataConnection) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 处理待修改的数据连接")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. detailedConfig由any转为tingYunDetailedConfig
	conf, err := tycp.any2TingYunDetailedConfig(ctx, conn.DataSourceConfig)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	preConf, err := tycp.any2TingYunDetailedConfig(ctx, preConn.DataSourceConfig)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	// 2. 校验tingYunDetailedConfig
	err = tycp.commonValidateWhenCreateAndUpdate(ctx, conf, true)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	// 3. 如果传入了新的secret_key, 则需要加密
	if conf.SecretKey == "" {
		conf.SecretKey = preConf.SecretKey
	} else {
		conf.SecretKey = common.EncryptPassword(conf.SecretKey)
	}

	conn.DataSourceConfig = conf
	return nil
}

func (tycp *tingYunConnectionProcessor) ComputeConfigMD5(ctx context.Context, conn *interfaces.DataConnection) (md5 string, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 计算详细配置的md5")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. detailedConfig由any转为tingYunDetailedConfig
	conf, err := tycp.any2TingYunDetailedConfig(ctx, conn.DataSourceConfig)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return "", err
	}

	// 2. 生成config的md5
	str := conf.Address + conf.Protocol + conf.ApiKey + conf.SecretKey
	return common.MD532Lower(str), nil
}

func (tycp *tingYunConnectionProcessor) GenerateAuthInfoAndStatus(ctx context.Context, conn *interfaces.DataConnection) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 生成auth_info和连接状态")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. detailedConfig由any转为tingYunDetailedConfig
	conf, err := tycp.any2TingYunDetailedConfig(ctx, conn.DataSourceConfig)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	// 2. 获取access_token, 并校验连通性
	// 2.1 获取access_token
	tyClientCfg := TingYunClientConfig{
		Address:   conf.Address,
		Protocol:  conf.Protocol,
		ApiKey:    conf.ApiKey,
		SecretKey: common.DecryptPassword(conf.SecretKey),
	}
	accessToken, err := tycp.getAccessToken(ctx, tyClientCfg)
	if err != nil {
		errDetails := fmt.Sprintf("Get access token of tingyun failed, err: %v", err.Error())
		logger.Error(errDetails)
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataConnection_InternalError_GetAccessTokenFailed).WithErrorDetails(errDetails)
	}

	// 2.2 校验连通性
	tyClientCfg.AccessToken = accessToken
	err = tycp.ping(ctx, tyClientCfg)
	if err != nil {
		errDetails := fmt.Sprintf("Verify the connectivity of tingyun failed, err: %v", err.Error())
		logger.Error(errDetails)
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_DataConnection_InternalError_VerifyConnectivityFailed).WithErrorDetails(errDetails)
	}

	// 3. 更新conn和status, 并返回结果
	conn.DataConnectionStatus.Status = "ok"
	conn.DataConnectionStatus.DetectionTime = time.Now().UnixMilli()

	conf.AccessToken = accessToken
	conn.DataSourceConfig = conf
	return nil
}

func (tycp *tingYunConnectionProcessor) UpdateAuthInfoAndStatus(ctx context.Context, conn *interfaces.DataConnection) (needWriteBack bool, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 更新auth_info和连接状态")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. []byte转tingYunDetailedConfig
	conf, err := tycp.any2TingYunDetailedConfig(ctx, conn.DataSourceConfig)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return false, err
	}

	// 2. 判断token是否即将过期
	// 根据配置中的tokenActiveTime来判断, 提前10分钟更新token
	now := time.Now().UnixMilli()
	tokenActiveTime := (time.Duration(tycp.appSetting.ThirdParty.TingYunTokenActiveTime) * time.Minute).Milliseconds()
	if conn.DetectionTime+tokenActiveTime-(10*time.Minute).Milliseconds() <= now {
		var newAccessToken string

		// 2.1 更新access_token
		tyClientCfg := TingYunClientConfig{
			Address:   conf.Address,
			Protocol:  conf.Protocol,
			ApiKey:    conf.ApiKey,
			SecretKey: common.DecryptPassword(conf.SecretKey),
		}

		newAccessToken, err = tycp.getAccessToken(ctx, tyClientCfg)
		if err != nil {
			errDetails := fmt.Sprintf("Get new access token of tingyun failed, err: %v", err.Error())
			logger.Error(errDetails)
			return false, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_GetAccessTokenFailed).WithErrorDetails(errDetails)
		}
		conf.AccessToken = newAccessToken

		// 2.2 校验连通性
		conn.DataConnectionStatus.Status = "ok"
		tyClientCfg.AccessToken = newAccessToken
		err = tycp.ping(ctx, tyClientCfg)
		if err != nil {
			conn.DataConnectionStatus.Status = "error"
			errDetails := fmt.Sprintf("Verify the connectivity of tingyun failed, err: %v", err.Error())
			logger.Error(errDetails)
			return false, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_DataConnection_InternalError_VerifyConnectivityFailed).WithErrorDetails(errDetails)
		}

		conn.DataConnectionStatus.DetectionTime = time.Now().UnixMilli()
		conf.ExpirationTime = conn.DetectionTime + tokenActiveTime
		conn.DataSourceConfig = conf
		return true, nil
	}

	conf.ExpirationTime = conn.DetectionTime + tokenActiveTime
	conn.DataSourceConfig = conf
	return false, nil
}

func (tycp *tingYunConnectionProcessor) HideAuthInfo(ctx context.Context, conn *interfaces.DataConnection) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 隐藏auth_info")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	conf, err := tycp.any2TingYunDetailedConfig(ctx, conn.DataSourceConfig)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	conf.SecretKey = ""
	conf.AccessToken = ""
	conn.DataSourceConfig = conf

	return nil
}

/*
	私有方法
*/

func (tycp *tingYunConnectionProcessor) any2TingYunDetailedConfig(ctx context.Context, i any) (*tingYunDetailedConfig, error) {
	switch t := i.(type) {
	case tingYunDetailedConfig:
		return i.(*tingYunDetailedConfig), nil
	case []byte:
		b := i.([]byte)
		conf := tingYunDetailedConfig{}
		err := sonic.Unmarshal(b, &conf)
		if err != nil {
			errDetails := fmt.Sprintf("Field config cannot be unmarshaled to tingYunDetailedConfig, err: %v", err.Error())
			logger.Error(errDetails)
			return &conf, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_UnmarshalDataFailed).WithErrorDetails(errDetails)
		}
		return &conf, nil
	default:
		b, err := sonic.Marshal(i)
		if err != nil {
			errDetails := fmt.Sprintf("Marshal field config with field type %v failed, err: %v", t, err.Error())
			logger.Error(errDetails)
			return &tingYunDetailedConfig{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_MarshalDataFailed).WithErrorDetails(errDetails)
		}

		conf := tingYunDetailedConfig{}
		err = sonic.Unmarshal(b, &conf)
		if err != nil {
			errDetails := fmt.Sprintf("Field config with field type %v cannot be unmarshaled to tingYunDetailedConfig, err: %v", t, err.Error())
			logger.Error(errDetails)
			return &conf, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_InternalError_UnmarshalDataFailed).WithErrorDetails(errDetails)
		}
		return &conf, nil
	}
}

func (tycp *tingYunConnectionProcessor) commonValidateWhenCreateAndUpdate(ctx context.Context, conf *tingYunDetailedConfig, isUpdate bool) error {
	// 2.1 校验address
	if conf.Address == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataConnection_NullParameter_Address)
	}

	// 2.2 校验protocol
	if conf.Protocol != "http" && conf.Protocol != "https" {
		errDetails := fmt.Sprintf("The protocol %v is invalid, valid protocol is http or https", conf.Protocol)
		logger.Error(errDetails)
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataConnection_InvalidParameter_Protocol).WithErrorDetails(errDetails)
	}

	// 2.3 校验api_key
	if conf.ApiKey == "" {
		errDetails := "The api_key is null"
		logger.Error(errDetails)
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataConnection_NullParameter_ApiKey).WithErrorDetails(errDetails)
	}

	// 2.4 校验secret_key
	if !isUpdate && conf.SecretKey == "" {
		errDetails := "The secret_key is null"
		logger.Error(errDetails)
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataConnection_NullParameter_SecretKey).WithErrorDetails(errDetails)
	}

	return nil
}

func (tycp *tingYunConnectionProcessor) getAccessToken(ctx context.Context, cfg TingYunClientConfig) (accessToken string, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 从听云获取access_token", trace.WithSpanKind(trace.SpanKindClient))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 拼接url
	currentTime := time.Now().UnixMilli()
	str := fmt.Sprintf("api_key=%s&secret_key=%s&timestamp=%v", cfg.ApiKey, cfg.SecretKey, currentTime)
	url := fmt.Sprintf("%s://%s/auth-api/auth/token?api_key=%s&auth=%s&timestamp=%v",
		cfg.Protocol, cfg.Address, cfg.ApiKey, common.MD532Lower(str), currentTime)

	span.SetAttributes(attr.Key("tingyun_url").String(url))

	// 2. 请求听云
	respCode, respBody, err := tycp.httpClient.GetNoUnmarshal(ctx, url, nil, nil)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get tingyun access token by http client: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return "", err
	}

	// 3. 解析resBody, 获取access_code
	node, err := sonic.Get(respBody, "code")
	if err != nil {
		errDetails := fmt.Sprintf("Using sonic to get code failed: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return "", err
	}

	code, err := node.Int64()
	if err != nil {
		errDetails := fmt.Sprintf("Using sonic to get code failed: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return "", err
	}

	if respCode != http.StatusOK || code != http.StatusOK {
		err := fmt.Errorf("failed to get tingyun access token by http client: %s", string(respBody))
		logger.Error(err.Error())
		o11y.Error(ctx, err.Error())
		return "", err
	}

	node, err = sonic.Get(respBody, "access_token")
	if err != nil {
		errDetails := fmt.Sprintf("Using sonic to get access_token failed: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return "", err
	}

	accessToken, err = node.String()
	if err != nil {
		errDetails := fmt.Sprintf("Using sonic to get access_token failed: %s", err.Error())
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return "", err
	}

	return accessToken, nil
}

func (tycp *tingYunConnectionProcessor) ping(ctx context.Context, cfg TingYunClientConfig) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven层: 校验与听云的连通性", trace.WithSpanKind(trace.SpanKindClient))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	url := fmt.Sprintf("%s://%s/server-api/action/trace?pageSize=0", cfg.Protocol, cfg.Address)
	headers := map[string]string{
		"Authorization": "Bearer " + cfg.AccessToken,
	}

	span.SetAttributes(attr.Key("tingyun_url").String(url))

	respCode, respBody, err := tycp.httpClient.GetNoUnmarshal(ctx, url, nil, headers)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to ping tingyun system: %s", err)
		logger.Errorf(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	if respCode != http.StatusOK {
		errDetails := fmt.Sprintf("Failed to ping tingyun system: %s", string(respBody))
		err := errors.New(errDetails)
		o11y.Error(ctx, errDetails)
		return err
	}

	return nil
}
