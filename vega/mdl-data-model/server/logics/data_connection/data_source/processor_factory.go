// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_source

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
	"data-model/logics/data_connection/data_source/anyrobot"
	"data-model/logics/data_connection/data_source/tingyun"
)

func NewDataConnectionProcessor(ctx context.Context, appSetting *common.AppSetting, dataSourceType string) (interfaces.DataConnectionProcessor, error) {
	switch dataSourceType {
	case interfaces.SOURCE_TYPE_ANYROBOT:
		return anyrobot.NewAnyRobotConnectionProcessor(appSetting), nil
	case interfaces.SOURCE_TYPE_TINGYUN:
		return tingyun.NewTingYunConnectionProcessor(appSetting), nil
	default:
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_DataConnection_InternalError_InitDataConnectionProcessor).
			WithErrorDetails(fmt.Sprintf("Invalid data_source_type: %v", dataSourceType))
	}
}
