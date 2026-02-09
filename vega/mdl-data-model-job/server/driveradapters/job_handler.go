// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"net/http"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/gin-gonic/gin"

	"data-model-job/common"
	derrors "data-model-job/errors"
	"data-model-job/interfaces"
)

func (r *restHandler) StartJob(c *gin.Context) {
	logger.Info("Start job")

	ctx := rest.GetLanguageCtx(c)
	jobInfo := interfaces.JobInfo{}
	err := c.ShouldBindJSON(&jobInfo)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModelJob_InvalidParameter_Job).
			WithErrorDetails(err.Error())

		rest.ReplyError(c, httpErr)
		return
	}

	// 启动任务
	err = r.jobService.StartJob(ctx, &jobInfo)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModelJob_InternalError_StartJobFailed).
			WithErrorDetails(err.Error())

		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]string{"id": jobInfo.JobId}
	c.Writer.Header().Set("Location", "/api/mdl-data-model-job/v1/jobs/"+jobInfo.JobId)
	rest.ReplyOK(c, http.StatusCreated, result)
}

func (r *restHandler) UpdateJob(c *gin.Context) {
	logger.Info("Update job")

	ctx := rest.GetLanguageCtx(c)
	jobInfo := interfaces.JobInfo{}
	err := c.ShouldBindJSON(&jobInfo)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModelJob_InvalidParameter_Job).
			WithErrorDetails(err.Error())

		rest.ReplyError(c, httpErr)
		return
	}

	jobId := c.Param("id")
	if jobId == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModelJob_InvalidParameter_JobId).
			WithErrorDetails("Job id is required")

		rest.ReplyError(c, httpErr)
		return
	}

	jobInfo.JobId = jobId

	// 根据job_id修改任务信息
	err = r.jobService.UpdateJob(ctx, &jobInfo)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModelJob_InternalError_UpdateJobFailed).
			WithErrorDetails(err.Error())

		rest.ReplyError(c, httpErr)
		return
	}

	rest.ReplyOK(c, http.StatusNoContent, nil)
}

func (r *restHandler) StopJobs(c *gin.Context) {
	logger.Info("Stop jobs")

	ctx := rest.GetLanguageCtx(c)

	jobIdsStr := c.Param("ids")
	jobIds := common.StringToStringSlice(jobIdsStr)

	// 循环根据job_id停止任务
	for _, jobId := range jobIds {
		err := r.jobService.StopJob(ctx, jobId)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, derrors.DataModelJob_InternalError_StopJobFailed).
				WithErrorDetails(err.Error())

			rest.ReplyError(c, httpErr)
			return
		}
	}

	rest.ReplyOK(c, http.StatusNoContent, nil)
}
