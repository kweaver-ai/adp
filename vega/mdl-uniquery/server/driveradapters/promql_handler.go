// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"uniquery/common"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics/promql/static"
)

// promql query_range 查询
func (r *restHandler) PromqlQueryRange(c *gin.Context) {
	ctx := rest.GetLanguageCtx(c)
	// 参数校验
	start, err := convert.ParseTime(c.PostForm("start"))
	if err != nil {
		common.ReplyError(c, http.StatusBadRequest, uerrors.InvalidParamError(err, "start"))
		return
	}

	end, err := convert.ParseTime(c.PostForm("end"))
	if err != nil {
		common.ReplyError(c, http.StatusBadRequest, uerrors.InvalidParamError(err, "end"))
		return
	}

	// 校验start end。判断场景  start > now, end > now, end < start
	end, err = validateTimeParams(start, end)
	if err != nil {
		common.ReplyError(c, http.StatusBadRequest, err)
		return
	}

	step, err := convert.ParseDuration(c.PostForm("step"))
	if err != nil {
		common.ReplyError(c, http.StatusBadRequest, uerrors.InvalidParamError(err, "step"))
		return
	}

	if step <= 0 {
		common.ReplyError(c, http.StatusBadRequest, uerrors.InvalidParamError(errors.New("zero or negative query resolution step widths are not accepted. Try a positive integer"), "step"))
		return
	}

	// 如果step超过30m,提示用户将step调整为5分钟的倍数
	if step > interfaces.SHARD_ROUTING_30M && step%interfaces.DEFAULT_STEP_DIVISOR != 0 {
		err = errors.New("step should a multiple of 5 minutes when step > 30min")
		common.ReplyError(c, http.StatusBadRequest, uerrors.InvalidParamError(err, "step"))
		return
	}

	// For safety, limit the number of returned points per timeseries.
	// This is sufficient for 60s resolution for a week or 1h resolution for a year.
	if end.Sub(start)/step > 11000 {
		err = errors.New("exceeded maximum resolution of 11,000 points per timeseries. Try decreasing the query resolution (?step=XX)")
		common.ReplyError(c, http.StatusBadRequest, uerrors.PromQLError{
			Typ: uerrors.ErrorBadData,
			Err: err})
		return
	}

	// 控制并发数。暂不考虑
	// sem <- struct{}{}        // 获取信号
	// defer func() { <-sem }() // 释放信号

	var timeoutMS int64
	if to := c.PostForm("timeout"); to != "" {
		timeout, err := convert.ParseDuration(to)
		if err != nil {
			common.ReplyError(c, http.StatusBadRequest, uerrors.InvalidParamError(err, "timeout"))
			return
		}
		timeoutMS = timeout.Milliseconds()
	}

	ts := time.Now().UnixNano()
	// 解析query语句并执行
	_, res, status, err := r.promqlService.Exec(ctx, interfaces.Query{
		QueryStr:   c.PostForm("query"),
		Start:      start.UnixMilli(),
		End:        end.UnixMilli(),
		Interval:   step.Milliseconds(),
		LogGroupId: c.PostForm("ar_dataview"),
		Limit:      -1,
	})
	// todo: 超时的暂时方案, 下个迭代解决此问题
	if timeoutMS > 0 && (time.Now().UnixNano()-ts)/1e6 >= timeoutMS {
		common.ReplyError(c, http.StatusServiceUnavailable, uerrors.PromQLError{
			Typ: uerrors.ErrorTimeout,
			Err: errors.New("query timed out in expression evaluation"),
		})
		return
	}
	if err != nil {
		common.ReplyError(c, status, err)
		return
	}
	common.ReplyOK(c, status, res)
}

// todo: 并发数50，定义为全局变量。是否需定义为配置项呢？
// var sem = make(chan struct{}, 50)

// promql query 查询
func (r *restHandler) PromqlQuery(c *gin.Context) {
	ctx := rest.GetLanguageCtx(c)
	// 参数校验
	tsSecond, err := convert.ParseTimeParam(c.PostForm("time"), "time", time.Now())
	if err != nil {
		common.ReplyError(c, http.StatusBadRequest, uerrors.InvalidParamError(err, "time"))
		return
	}

	// 控制并发数。暂不考虑
	// sem <- struct{}{}        // 获取信号
	// defer func() { <-sem }() // 释放信号

	var timeoutMS int64
	if to := c.PostForm("timeout"); to != "" {
		timeout, err := convert.ParseDuration(to)
		if err != nil {
			common.ReplyError(c, http.StatusBadRequest, uerrors.InvalidParamError(err, "timeout"))
			return
		}
		timeoutMS = timeout.Milliseconds()
	}

	ts := time.Now().UnixNano()
	// 解析query语句并执行
	_, res, status, err := r.promqlService.Exec(ctx, interfaces.Query{
		QueryStr:       c.PostForm("query"),
		Start:          tsSecond.UnixMilli(),
		End:            tsSecond.UnixMilli(),
		Interval:       1,
		IsInstantQuery: true,
		LogGroupId:     c.PostForm("ar_dataview"),
		Limit:          -1,
	})
	// todo: 超时的暂时方案, 下个迭代解决此问题
	if timeoutMS > 0 && (time.Now().UnixNano()-ts)/1e6 >= timeoutMS {
		common.ReplyError(c, http.StatusServiceUnavailable, uerrors.PromQLError{
			Typ: uerrors.ErrorTimeout,
			Err: errors.New("query timed out in expression evaluation"),
		})
		return
	}
	if err != nil {
		common.ReplyError(c, status, err)
		return
	}
	common.ReplyOK(c, status, res)
}

// promql series 查询
func (r *restHandler) PromqlSeries(c *gin.Context) {
	// 参数校验
	if len(c.PostFormArray("match[]")) == 0 {
		common.ReplyError(c, http.StatusBadRequest, uerrors.PromQLError{
			Typ: uerrors.ErrorBadData,
			Err: fmt.Errorf("no match[] parameter provided"),
		})
		return
	}

	start, err := convert.ParseTimeParam(c.PostForm("start"), "start", convert.MinTime)
	if err != nil {
		common.ReplyError(c, http.StatusBadRequest, uerrors.InvalidParamError(err, "start"))
		return
	}

	// end 默认是 now
	end, err := convert.ParseTimeParam(c.PostForm("end"), "end", time.Now())
	if err != nil {
		common.ReplyError(c, http.StatusBadRequest, uerrors.InvalidParamError(err, "end"))
		return
	}

	// 校验start end。判断场景  start > now, end > now, end < start
	end, err = validateTimeParams(start, end)
	if err != nil {
		common.ReplyError(c, http.StatusBadRequest, err)
		return
	}

	// match[] 解析
	matcherSets, err := static.ParseMatchersParam(c.PostFormArray("match[]"))
	if err != nil {
		common.ReplyError(c, http.StatusBadRequest, uerrors.InvalidParamError(err, "match[]"))
		return
	}

	// 根据序列选择器集合去查找序列列表
	res, status, err := r.promqlService.Series(interfaces.Matchers{
		MatcherSet: matcherSets,
		Start:      convert.FromTime(start),
		End:        convert.FromTime(end),
		LogGroupId: c.PostForm("ar_dataview"),
	})

	if err != nil {
		common.ReplyError(c, status, err)
		return
	}
	common.ReplyOK(c, status, res)
}

func validateTimeParams(start time.Time, end time.Time) (time.Time, error) {
	// start 是未来时间就抛异常,附带当前时间
	currentTime := time.Now()
	if start.After(currentTime) {
		return time.Time{}, uerrors.InvalidParamError(
			fmt.Errorf("start is greater than current time, current time is %v", currentTime.Unix()), "start")
	}

	// 如果end 大于 current_time，那么end = current
	if end.After(currentTime) {
		end = currentTime
	}

	if end.Before(start) {
		return time.Time{}, uerrors.InvalidParamError(errors.New("end timestamp must not be before start time"), "end")
	}

	return end, nil
}
