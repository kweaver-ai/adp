// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	ierrors "uniquery/errors"
	"uniquery/interfaces"
)

// 基于事件模型的事件数据预览(内部)
func (r *restHandler) QueryByIn(c *gin.Context) {
	logger.Debug("Handler QueryByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor

	visitor := GenerateVisitor(c)
	r.Query(c, visitor)
}

// 基于事件模型的事件数据预览（外部）
func (r *restHandler) QueryByEx(c *gin.Context) {
	logger.Debug("Handler QueryByEx Start")

	// 校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.Query(c, visitor)
}

// 基于事件模型的事件数据预览，默认查询最近5m的数据
func (r *restHandler) Query(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler EventModel Start")
	startTime := time.Now()
	ctx := rest.GetLanguageCtx(c)

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	//接收绑定参数
	queryReq := interfaces.EventQueryReq{}
	err := c.ShouldBindJSON(&queryReq)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, ierrors.Uniquery_EventModel_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())
		endTime := time.Now()
		logger.Infof("Api [Query]: request time [%d] ms, status code [%d]",
			endTime.UnixNano()/(int64(time.Millisecond))-startTime.UnixNano()/(int64(time.Millisecond)), httpErr.HTTPCode)
		rest.ReplyError(c, httpErr)
		return
	}

	//NOTE: 查询请求
	total, entries, entities, err := r.eService.Query(ctx, queryReq)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		endTime := time.Now()
		logger.Infof("Api [Query]: request time [%d] ms, status code [%d]",
			endTime.UnixNano()/(int64(time.Millisecond))-startTime.UnixNano()/(int64(time.Millisecond)), httpErr.HTTPCode)
		rest.ReplyError(c, httpErr)
		return
	}
	result := map[string]interface{}{
		"total_count": total,
		"entries":     entries,
		"entities":    entities,
	}

	logger.Infof(fmt.Sprintf("Handler Event query  end, query result,event cnt:%d,entities cnt:%d", total, len(entities)))
	//NOTE: 运行时间;接口信息;返回码
	endTime := time.Now()
	logger.Infof("Api [Query]: request time [%d] ms, status code [%d]",
		endTime.UnixNano()/(int64(time.Millisecond))-startTime.UnixNano()/(int64(time.Millisecond)), http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// 基于事件ID的查询事件详情(内部)
func (r *restHandler) QuerySingleEventByEventIdByIn(c *gin.Context) {
	logger.Debug("Handler QuerySingleEventByEventIdByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor

	visitor := GenerateVisitor(c)
	r.QuerySingleEventByEventId(c, visitor)
}

// 基于事件ID的查询事件详情（外部）
func (r *restHandler) QuerySingleEventByEventIdByEx(c *gin.Context) {
	logger.Debug("Handler QuerySingleEventByEventIdByEx Start")

	// 校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.QuerySingleEventByEventId(c, visitor)
}

// 基于事件ID的查询事件详情
func (r *restHandler) QuerySingleEventByEventId(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler Event query Start")
	startTime := time.Now()
	ctx := rest.GetLanguageCtx(c)

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	query := interfaces.EventDetailsQueryReq{}
	err := c.ShouldBindUri(&query)
	if err != nil {
		logger.Errorf("QuerySingleEventByEventId failed,binding uri failed:%v", err)
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, ierrors.Uniquery_EventModel_InvalidParameter).
			WithErrorDetails("Binding  error:" + err.Error())
		endTime := time.Now()
		logger.Infof("Api [QuerySingleEventByEventId]: request time [%d] ms, status code [%d]",
			endTime.UnixNano()/(int64(time.Millisecond))-startTime.UnixNano()/(int64(time.Millisecond)), httpErr.HTTPCode)
		rest.ReplyError(c, httpErr)
		return
	}
	err = c.ShouldBindQuery(&query)
	if err != nil {
		logger.Errorf("QuerySingleEventByEventId failed,binding json failed:%v", err)
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, ierrors.Uniquery_EventModel_InvalidParameter).
			WithErrorDetails("Binding  error:" + err.Error())
		endTime := time.Now()
		logger.Infof("Api [QuerySingleEventByEventId]: request time [%d] ms, status code [%d]",
			endTime.UnixNano()/(int64(time.Millisecond))-startTime.UnixNano()/(int64(time.Millisecond)), httpErr.HTTPCode)
		rest.ReplyError(c, httpErr)
		return
	}

	result, err := r.eService.QuerySingleEventByEventId(ctx, query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		logger.Errorf("QuerySingleEventByEventId failed:%v", httpErr)
		endTime := time.Now()
		logger.Infof("Api [QuerySingleEventByEventId]: request time [%d] ms, status code [%d]",
			endTime.UnixNano()/(int64(time.Millisecond))-startTime.UnixNano()/(int64(time.Millisecond)), httpErr.HTTPCode)
		rest.ReplyError(c, httpErr)
		return
	}
	endTime := time.Now()
	logger.Infof("Api [QuerySingleEventByEventId]: request time [%d] ms, status code [%d]",
		endTime.UnixNano()/(int64(time.Millisecond))-startTime.UnixNano()/(int64(time.Millisecond)), http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}
