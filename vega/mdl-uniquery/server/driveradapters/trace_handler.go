// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	uerrors "uniquery/errors"
)

// GetSpanList span列表查询
// func (r *restHandler) GetSpanList(c *gin.Context) {
// 	logger.Debug("Handler GetSpanList Start")

// 	// 接收offset, limit参数
// 	offsetStr := c.DefaultQuery("offset", interfaces.DEFAULT_OFFSET_STR)
// 	limitStr := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT_STR)

// 	// 校验offset, limit参数, 并将其转化为int类型
// 	offset, limit, httpErr := ValidateOffsetAndLimit(offsetStr, limitStr)
// 	if httpErr != nil {
// 		rest.ReplyError(c, httpErr)
// 		return
// 	}

// 	// 接收span_statuses参数
// 	spanStatuses := c.DefaultQuery("span_statuses", interfaces.DEFAULT_SPAN_STATUSES)

// 	// 校验span_statuses参数
// 	spanStatusMap, httpErr := ValidateSpanStatuses(spanStatuses)
// 	if httpErr != nil {
// 		rest.ReplyError(c, httpErr)
// 		return
// 	}

// 	// 接收ar_dataview参数
// 	dataViewId, ok := c.GetQuery("ar_dataview")
// 	if !ok {
// 		httpErr := rest.NewHTTPError(http.StatusBadRequest, uerrors.Uniquery_MissingParameter_ARDataView)
// 		rest.ReplyError(c, httpErr)
// 		return
// 	}

// 	// 校验ar_dataview参数
// 	httpErr = ValidateDataView(dataViewId)
// 	if httpErr != nil {
// 		rest.ReplyError(c, httpErr)
// 		return
// 	}

// 	// 接收start_time参数
// 	startTime, ok := c.GetQuery("start_time")
// 	if !ok {
// 		httpErr := rest.NewHTTPError(http.StatusBadRequest, uerrors.Uniquery_MissingParameter_StartTime)
// 		rest.ReplyError(c, httpErr)
// 		return
// 	}

// 	// 接收end_time参数
// 	endTime, ok := c.GetQuery("end_time")
// 	if !ok {
// 		httpErr := rest.NewHTTPError(http.StatusBadRequest, uerrors.Uniquery_MissingParameter_EndTime)
// 		rest.ReplyError(c, httpErr)
// 		return
// 	}

// 	// 校验start_time, end_time参数
// 	start, end, httpErr := ValidateSpanQueryTime(startTime, endTime)
// 	if httpErr != nil {
// 		rest.ReplyError(c, httpErr)
// 		return
// 	}

// 	// 构造span列表查询结构体
// 	spanListQuery := interfaces.SpanListQuery{
// 		DataViewId:    dataViewId,
// 		Offset:        offset,
// 		Limit:         limit,
// 		StartTime:     start,
// 		EndTime:       end,
// 		SpanStatusMap: spanStatusMap,
// 	}

// 	// 调用logic层
// 	spanList, total, httpErr := r.tService.GetSpanList(spanListQuery)
// 	if httpErr != nil {
// 		rest.ReplyError(c, httpErr)
// 		return
// 	}

// 	result := map[string]interface{}{"entries": spanList, "total_count": total}

// 	logger.Debug("Handler GetSpanList Success")
// 	rest.ReplyOK(c, http.StatusOK, result)
// }

// GetTraceDetail 单条trace详情查询
func (r *restHandler) GetTraceDetail(c *gin.Context) {
	// start := time.Now()
	// fmt.Printf("开始查询单条Trace, 当前时间%v\n", start)

	logger.Debug("Handler GetTraceDetail Start")
	ctx := rest.GetLanguageCtx(c)

	// 获取url上的trace_id值
	traceId := c.Param("trace_id")

	// 接收请求参数
	// trace_data_view_id: 链路的日志分组
	traceDataViewId, ok := c.GetQuery("trace_data_view_id")
	if !ok {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_Trace_MissingParameter_TraceDataViewID)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验参数
	httpErr := ValidateDataView(ctx, traceDataViewId, "trace")
	if httpErr != nil {
		rest.ReplyError(c, httpErr)
		return
	}

	// log_data_view_id: 关联日志的日志分组
	logDataViewId, ok := c.GetQuery("log_data_view_id")
	if !ok {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_Trace_MissingParameter_LogDataViewID)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验参数
	httpErr = ValidateDataView(ctx, logDataViewId, "log")
	if httpErr != nil {
		rest.ReplyError(c, httpErr)
		return
	}

	// 调用logic层
	res, httpErr := r.tService.GetTraceDetail(ctx, traceDataViewId, logDataViewId, traceId)
	if httpErr != nil {
		rest.ReplyError(c, httpErr)
		return
	}

	// end := time.Now()
	// fmt.Printf("结束查询单条Trace, 准备返回结果, 当前时间%v, 共耗时%v\n", end, end.Sub(start))
	logger.Debug("Handler GetTraceDetail Success")
	rest.ReplyOK(c, http.StatusOK, res)
}
