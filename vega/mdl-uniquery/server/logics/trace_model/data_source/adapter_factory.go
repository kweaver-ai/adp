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

	"uniquery/common"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	data_view "uniquery/logics/trace_model/data_source/data_view"
	tingyun "uniquery/logics/trace_model/data_source/tingyun"
)

func NewTraceModelAdapter(ctx context.Context, dataSourceType string, appSetting *common.AppSetting) (interfaces.TraceModelAdapter, error) {
	switch dataSourceType {
	case interfaces.SOURCE_TYPE_DATA_VIEW:
		return data_view.NewDataViewAdapter(appSetting), nil
	case interfaces.SOURCE_TYPE_TINGYUN:
		return tingyun.NewTingYunAdapter(appSetting), nil
	default:
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError).
			WithErrorDetails(fmt.Sprintf("Invalid data_source_type %s, TraceModelAdapter cannot be manufactured based on this data_source_type", dataSourceType))
	}
}
