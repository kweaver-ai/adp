// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"

	derrors "data-model/errors"
	"data-model/interfaces"
)

/*
数据连接模块的校验函数. 包括:
	(1) 创建时的数据连接校验
	(2) 修改时的数据连接校验
*/

// 数据连接校验函数(1): 创建时的数据连接校验
func validateDataConnectionWhenCreate(ctx context.Context, r *restHandler, reqConn *interfaces.DataConnection) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driver层: 校验传入的数据连接")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 校验数据连接名称
	err = validateObjectName(ctx, reqConn.Name, interfaces.DATA_CONNECTION)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	name2ID, err := r.dcs.GetMapAboutName2ID(ctx, []string{reqConn.Name})
	if err != nil {
		return err
	}

	if _, ok := name2ID[reqConn.Name]; ok {
		errDetails := fmt.Sprintf("Data connection name %s already exists", reqConn.Name)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return rest.NewHTTPError(ctx, http.StatusForbidden,
			derrors.DataModel_DataConnection_ConnectionNameExisted).WithErrorDetails(errDetails)
	}

	// 2. 校验data_source_type
	if _, ok := interfaces.DataSourceType2ApplicationScope[reqConn.DataSourceType]; !ok {
		errDetails := fmt.Sprintf("The data_source_type %s is invalid", reqConn.DataSourceType)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataConnection_InvalidParameter_DataSourceType)
	}

	// 3. 校验tags
	err = validateObjectTags(ctx, reqConn.Tags)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	// 4. 校验comment
	err = validateObjectComment(ctx, reqConn.Comment)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	return nil
}

// 数据连接校验函数(2): 修改时的数据连接校验
func validateDataConnectionWhenUpdate(ctx context.Context, r *restHandler,
	reqConn *interfaces.DataConnection, preConn *interfaces.DataConnection) (err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "driver层: 校验传入的数据连接")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 校验data_source_type有无修改
	if reqConn.DataSourceType != preConn.DataSourceType {
		errDetails := "The data_source_type cannot be modified"
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataConnection_ForbiddenUpdateParameter_DataSourceType)
	}

	// 2. 校验connection_name有无修改
	// 若修改, 需校验新名称在数据库中是否存在
	if reqConn.Name != preConn.Name {
		err := validateObjectName(ctx, reqConn.Name, interfaces.DATA_CONNECTION)
		if err != nil {
			o11y.Error(ctx, err.Error())
			return err
		}

		name2ID, err := r.dcs.GetMapAboutName2ID(ctx, []string{reqConn.Name})
		if err != nil {
			return err
		}

		if _, ok := name2ID[reqConn.Name]; ok {
			errDetails := fmt.Sprintf("Data connection name %s already exists", reqConn.Name)
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			return rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_DataConnection_ConnectionNameExisted).
				WithErrorDetails(errDetails)
		}
	}

	// 3. 校验tags
	err = validateObjectTags(ctx, reqConn.Tags)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	// 4. 校验comment
	err = validateObjectComment(ctx, reqConn.Comment)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return err
	}

	return nil
}
