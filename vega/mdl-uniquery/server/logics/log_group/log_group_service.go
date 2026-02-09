// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package log_group

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"

	"uniquery/common"
	"uniquery/interfaces"
	"uniquery/logics"
)

var (
	lgServiceOnce sync.Once
	lgService     interfaces.LogGroupService
)

type logGroupService struct {
	setting        *common.AppSetting
	dataManagerUrl string
	searchUrl      string
	dmAccess       interfaces.DataManagerAccess
	sAccess        interfaces.SearchAccess
}

func NewLogGroupService(setting *common.AppSetting) interfaces.LogGroupService {
	lgServiceOnce.Do(func() {
		lgService = &logGroupService{
			setting:        setting,
			dataManagerUrl: setting.DataManagerUrl,
			searchUrl:      setting.SearchUrl,
			dmAccess:       logics.DMAccess,
			sAccess:        logics.SAccess,
		}
	})
	return lgService
}

func (lgs *logGroupService) GetLogGroupRootsByConn(ctx context.Context, conn *interfaces.DataConnection) (any, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get single view data for internal module")
	defer span.End()

	host := ""
	accessToken := ""
	if conn != nil {
		arConf, ok := conn.DataSourceConfig.(map[string]any)
		if !ok {
			errDetails := fmt.Sprintf("invalid config for type convert: %v", conn)
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return nil, errors.New(errDetails)
		}

		host = fmt.Sprintf("%s://%s", arConf["protocol"].(string), arConf["address"].(string))
		accessToken = arConf["access_token"].(string)
	} else {
		host = lgs.dataManagerUrl
	}
	data, err := lgs.dmAccess.GetLogGroupRoots(ctx, host, accessToken)
	return data, err
}

func (lgs *logGroupService) GetLogGroupTreeByConn(ctx context.Context,
	userID string, conn *interfaces.DataConnection) (any, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get single view data for internal module")
	defer span.End()

	host := ""
	accessToken := ""
	if conn != nil {
		arConf, ok := conn.DataSourceConfig.(map[string]any)
		if !ok {
			errDetails := fmt.Sprintf("invalid config for type convert: %v", conn)
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return nil, errors.New(errDetails)
		}

		host = fmt.Sprintf("%s://%s", arConf["protocol"].(string), arConf["address"].(string))
		accessToken = arConf["access_token"].(string)
	} else {
		host = lgs.dataManagerUrl
	}
	data, err := lgs.dmAccess.GetLogGroupTree(ctx, userID, host, accessToken)
	return data, err
}

func (lgs *logGroupService) GetLogGroupChildrenByConn(ctx context.Context,
	logGroupID string, conn *interfaces.DataConnection) (any, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get single view data for internal module")
	defer span.End()

	host := ""
	accessToken := ""
	if conn != nil {
		arConf, ok := conn.DataSourceConfig.(map[string]any)
		if !ok {
			errDetails := fmt.Sprintf("invalid config for type convert: %v", conn)
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return nil, errors.New(errDetails)
		}

		host = fmt.Sprintf("%s://%s", arConf["protocol"].(string), arConf["address"].(string))
		accessToken = arConf["access_token"].(string)
	} else {
		host = lgs.dataManagerUrl
	}
	data, err := lgs.dmAccess.GetLogGroupChildren(ctx, logGroupID, host, accessToken)
	return data, err
}

func (lgs *logGroupService) SearchSubmitByConn(ctx context.Context,
	queryBody any, userID string, conn *interfaces.DataConnection) (any, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get single view data for internal module")
	defer span.End()

	host := ""
	accessToken := ""
	if conn != nil {
		arConf, ok := conn.DataSourceConfig.(map[string]any)
		if !ok {
			errDetails := fmt.Sprintf("invalid config for type convert: %v", conn)
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return nil, errors.New(errDetails)
		}

		host = fmt.Sprintf("%s://%s", arConf["protocol"].(string), arConf["address"].(string))
		accessToken = arConf["access_token"].(string)
	} else {
		host = lgs.searchUrl
	}
	data, err := lgs.sAccess.SearchSubmit(ctx, queryBody, userID, host, accessToken)
	return data, err
}

func (lgs *logGroupService) SearchFetchByConn(ctx context.Context,
	jobID string, queryParams url.Values, conn *interfaces.DataConnection) (any, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get single view data for internal module")
	defer span.End()

	host := ""
	accessToken := ""
	if conn != nil {
		arConf, ok := conn.DataSourceConfig.(map[string]any)
		if !ok {
			errDetails := fmt.Sprintf("invalid config for type convert: %v", conn)
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return nil, errors.New(errDetails)
		}

		host = fmt.Sprintf("%s://%s", arConf["protocol"].(string), arConf["address"].(string))
		accessToken = arConf["access_token"].(string)
	} else {
		host = lgs.searchUrl
	}
	data, err := lgs.sAccess.SearchFetch(ctx, jobID, queryParams, host, accessToken)
	return data, err
}

func (lgs *logGroupService) SearchFetchFieldsByConn(ctx context.Context,
	jobID string, queryParams url.Values, conn *interfaces.DataConnection) (any, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get single view data for internal module")
	defer span.End()

	host := ""
	accessToken := ""
	if conn != nil {
		arConf, ok := conn.DataSourceConfig.(map[string]any)
		if !ok {
			errDetails := fmt.Sprintf("invalid config for type convert: %v", conn)
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return nil, errors.New(errDetails)
		}

		host = fmt.Sprintf("%s://%s", arConf["protocol"].(string), arConf["address"].(string))
		accessToken = arConf["access_token"].(string)
	} else {
		host = lgs.searchUrl
	}
	data, err := lgs.sAccess.SearchFetchFields(ctx, jobID, queryParams, host, accessToken)
	return data, err
}

func (lgs *logGroupService) SearchFetchSameFieldsByConn(ctx context.Context,
	jobID string, conn *interfaces.DataConnection) (any, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get single view data for internal module")
	defer span.End()

	host := ""
	accessToken := ""
	if conn != nil {
		arConf, ok := conn.DataSourceConfig.(map[string]any)
		if !ok {
			errDetails := fmt.Sprintf("invalid config for type convert: %v", conn)
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return nil, errors.New(errDetails)
		}

		host = fmt.Sprintf("%s://%s", arConf["protocol"].(string), arConf["address"].(string))
		accessToken = arConf["access_token"].(string)
	} else {
		host = lgs.searchUrl
	}
	data, err := lgs.sAccess.SearchFetchSameFields(ctx, jobID, host, accessToken)
	return data, err
}

func (lgs *logGroupService) SearchContextByConn(ctx context.Context,
	queryParams url.Values, userID string, conn *interfaces.DataConnection) (any, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Get single view data for internal module")
	defer span.End()

	host := ""
	accessToken := ""
	if conn != nil {
		arConf, ok := conn.DataSourceConfig.(map[string]any)
		if !ok {
			errDetails := fmt.Sprintf("invalid config for type convert: %v", conn)
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return nil, errors.New(errDetails)
		}

		host = fmt.Sprintf("%s://%s", arConf["protocol"].(string), arConf["address"].(string))
		accessToken = arConf["access_token"].(string)
	} else {
		host = lgs.searchUrl
	}
	data, err := lgs.sAccess.SearchContext(ctx, queryParams, userID, host, accessToken)
	return data, err
}
