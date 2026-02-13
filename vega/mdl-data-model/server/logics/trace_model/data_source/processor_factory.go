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
	"data-model/logics/trace_model/data_source/data_view"
	"data-model/logics/trace_model/data_source/tingyun"
)

func NewTraceModelProcessor(ctx context.Context, appSetting *common.AppSetting, dataSourceType string) (interfaces.TraceModelProcessor, error) {
	switch dataSourceType {
	case interfaces.SOURCE_TYPE_DATA_VIEW:
		return data_view.NewDataViewTraceProcessor(appSetting), nil
	case interfaces.SOURCE_TYPE_TINGYUN:
		return tingyun.NewTingYunTraceProcessor(appSetting), nil
	default:
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModel_TraceModel_InternalError_InitTraceModelProcessor).
			WithErrorDetails(fmt.Sprintf("Invalid data_source_type: %v", dataSourceType))
	}
}
