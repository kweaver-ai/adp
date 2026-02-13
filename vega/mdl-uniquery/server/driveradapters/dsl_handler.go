// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"time"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"uniquery/common"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
)

// GetResult dsl查询
func (r *restHandler) DslGetResult(c *gin.Context) {
	ctx := rest.GetLanguageCtx(c)

	param := c.Query("scroll")
	var scroll time.Duration
	var err error
	if param != "" {
		scroll, err = convert.IntToDuration(param)
		if err != nil {
			oerr := uerrors.NewOpenSearchError(uerrors.IllegalArgumentException).WithReason(err.Error())
			common.ReplyError(c, oerr.StatusCode, oerr)
			return
		}
	}

	var dsl map[string]interface{}
	err = sonic.ConfigDefault.NewDecoder(c.Request.Body).Decode(&dsl)
	if err != nil {
		oerr := uerrors.NewOpenSearchError(uerrors.IllegalArgumentException).WithReason("illegal json body")
		common.ReplyError(c, oerr.StatusCode, oerr)
		return
	}

	pathLib := c.Param("index")
	res, status, err := r.dslService.Search(ctx, dsl, pathLib, scroll)
	if err != nil {
		common.ReplyError(c, status, err)
		return
	}

	common.ReplyOK(c, status, res)
}

// Scroll scroll查询
func (r *restHandler) DslScroll(c *gin.Context) {
	ctx := rest.GetLanguageCtx(c)

	reqBody := interfaces.Scroll{}
	err := c.ShouldBindJSON(&reqBody)
	if err != nil {
		oerr := uerrors.NewOpenSearchError(uerrors.IllegalArgumentException).WithReason(err.Error())
		common.ReplyError(c, oerr.StatusCode, oerr)
		return
	}

	res, status, err := r.dslService.ScrollSearch(ctx, reqBody)
	if err != nil {
		common.ReplyError(c, status, err)
		return
	}

	common.ReplyOK(c, status, res)
}

// GetCount 获取查询数据的总数
func (r *restHandler) DslGetCount(c *gin.Context) {
	ctx := rest.GetLanguageCtx(c)

	var dsl map[string]interface{}
	var err error
	err = sonic.ConfigDefault.NewDecoder(c.Request.Body).Decode(&dsl)
	if err != nil {
		oerr := uerrors.NewOpenSearchError(uerrors.IllegalArgumentException).WithReason("illegal json body")
		common.ReplyError(c, oerr.StatusCode, oerr)
		return
	}

	pathLib := c.Param("index")
	res, status, err := r.dslService.Count(ctx, dsl, pathLib)
	if err != nil {
		common.ReplyError(c, status, err)
		return
	}

	common.ReplyOK(c, status, res)
}

// DeleteScroll 删除scroll查询
func (r *restHandler) DslDeleteScroll(c *gin.Context) {
	ctx := rest.GetLanguageCtx(c)

	reqBody := interfaces.DeleteScroll{}
	err := c.ShouldBindJSON(&reqBody)
	if err != nil {
		oerr := uerrors.NewOpenSearchError(uerrors.IllegalArgumentException).WithReason("illegal json body")
		common.ReplyError(c, oerr.StatusCode, oerr)
		return
	}

	res, status, err := r.dslService.DeleteScroll(ctx, reqBody)
	if err != nil {
		common.ReplyError(c, status, err)
		return
	}

	common.ReplyOK(c, status, res)
}

// DeleteAllScroll 删除所有scroll查询
func (r *restHandler) DslDeleteAllScroll(c *gin.Context) {
	ctx := rest.GetLanguageCtx(c)

	reqBody := interfaces.DeleteScroll{
		ScrollId: []string{"_all"},
	}
	res, status, err := r.dslService.DeleteScroll(ctx, reqBody)
	if err != nil {
		common.ReplyError(c, status, err)
		return
	}

	common.ReplyOK(c, status, res)
}
