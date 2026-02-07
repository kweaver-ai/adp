// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_model_job

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"data-model/common"
	"data-model/interfaces"
)

const (
	DATA_MODEL_JOB_TABLE_NAME = "t_data_model_job"
)

var (
	dmjAccessOnce sync.Once
	dmjAccess     interfaces.DataModelJobAccess
)

type dataModelJobAccess struct {
	appSetting *common.AppSetting
	db         *sql.DB
	httpClient rest.HTTPClient
}

func NewDataModelJobAccess(appSetting *common.AppSetting) interfaces.DataModelJobAccess {
	dmjAccessOnce.Do(func() {
		dmjAccess = &dataModelJobAccess{
			appSetting: appSetting,
			db:         libdb.NewDB(&appSetting.DBSetting),
			httpClient: common.NewHTTPClient(),
		}
	})

	return dmjAccess
}

// 在数据库中创建实时订阅任务
func (dmja *dataModelJobAccess) CreateDataModelJob(ctx context.Context, tx *sql.Tx, job *interfaces.JobInfo) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Insert a streaming job into DB", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
	)

	builder := sq.Insert(DATA_MODEL_JOB_TABLE_NAME).
		Columns(
			"f_job_id",
			"f_job_type",
			"f_job_config",
			"f_job_status",
			"f_job_status_details",
			"f_create_time",
			"f_update_time",
			"f_creator",
			"f_creator_type",
		)

	configBytes, err := sonic.Marshal(job.JobConfig)
	if err != nil {
		errDetails := fmt.Sprintf("Marshal jobConfig failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Marshal jobConfig failed")

		return err
	}

	builder = builder.Values(
		job.JobID,
		job.JobType,
		configBytes,
		job.JobStatus,
		job.JobStatusDetails,
		job.CreateTime,
		job.UpdateTime,
		job.Creator.ID,
		job.Creator.Type,
	)

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'create a streaming job' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("sql stmt for creating a streaming job is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	_, err = tx.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("insert a streaming job failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Insert a streaming job failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil

}

// 在数据库中批量删除实时订阅任务
func (dmja *dataModelJobAccess) DeleteDataModelJobs(ctx context.Context, tx *sql.Tx, jobIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Delete data model jobs from DB", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("db_url").String(libdb.GetDBUrl()),
		attr.Key("db_type").String(libdb.GetDBType()),
		attr.Key("jobID").String(fmt.Sprintf("%v", jobIDs)),
	)

	sqlStr, args, err := sq.Delete(DATA_MODEL_JOB_TABLE_NAME).
		Where(sq.Eq{"f_job_id": jobIDs}).
		ToSql()
	if err != nil {
		errDetails := fmt.Sprintf("Generate 'delete data model jobs' sql stmt failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Generate sql stmt failed")

		return err
	}

	sqlStmt := fmt.Sprintf("Sql stmt for deleting data model jobs is '%s'", sqlStr)
	logger.Debug(sqlStmt)
	o11y.Info(ctx, sqlStmt)

	_, err = tx.Exec(sqlStr, args...)
	if err != nil {
		errDetails := fmt.Sprintf("Delete data model jobs failed, %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		span.SetStatus(codes.Error, "Delete data model jobs failed")

		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 发起http请求运行job
func (dmja *dataModelJobAccess) StartJob(ctx context.Context, job *interfaces.DataModelJobCfg) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Start job", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	url := dmja.appSetting.DataModelJobUrl

	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         url,
		HttpMethod:      http.MethodPost,
		HttpContentType: rest.ContentTypeJson,
	})

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	respCode, respData, err := dmja.httpClient.PostNoUnmarshal(ctx, url, headers, job)
	if err != nil {
		errDetails := fmt.Sprintf("start job '%s' failed, %s", job.JobID, err.Error())
		logger.Error(errDetails)

		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "start job failed")

		return err
	}

	if respCode != http.StatusCreated {
		var baseError rest.BaseError
		if err := sonic.Unmarshal(respData, &baseError); err != nil {
			errDetails := fmt.Sprintf("Unmalshal baesError failed: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal baesError failed")

			return err
		}

		o11y.Error(ctx, fmt.Sprintf("%s. %v", baseError.Description, baseError.ErrorDetails))
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status code is not 200")
		return fmt.Errorf("start job '%s' failed, errDetails: %v", job.JobID, baseError.ErrorDetails)
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return nil
}

// 发起http请求批量停止job
func (dmja *dataModelJobAccess) StopJobs(ctx context.Context, jobIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Stop jobs", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	jobIDsStr := strings.Join(jobIDs, ",")
	if jobIDsStr == "" {
		return nil
	}

	url := fmt.Sprintf("%s/%s", dmja.appSetting.DataModelJobUrl, jobIDsStr)

	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         url,
		HttpMethod:      http.MethodDelete,
		HttpContentType: rest.ContentTypeJson,
	})

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	respCode, respData, err := dmja.httpClient.DeleteNoUnmarshal(context.Background(), url, headers)
	if err != nil {
		errDetails := fmt.Sprintf("stop jobs '%s' failed, %s", jobIDsStr, err.Error())
		logger.Error(errDetails)

		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "stop jobs failed")

		return err
	}

	if respCode != http.StatusNoContent {
		var baseError rest.BaseError
		if err := sonic.Unmarshal(respData, &baseError); err != nil {
			errDetails := fmt.Sprintf("Unmalshal baesError failed: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal baesError failed")

			return err
		}

		o11y.Error(ctx, fmt.Sprintf("%s. %v", baseError.Description, baseError.ErrorDetails))
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status code is not 200")
		return fmt.Errorf("stop jobs '%s' failed, errDetails: %v", jobIDsStr, baseError.ErrorDetails)
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return nil
}

// 发起http请求更新job
func (dmja *dataModelJobAccess) UpdateJob(ctx context.Context, job *interfaces.DataModelJobCfg) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Update job", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	url := fmt.Sprintf("%s/%s", dmja.appSetting.DataModelJobUrl, job.JobID)

	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         url,
		HttpMethod:      http.MethodPut,
		HttpContentType: rest.ContentTypeJson,
	})

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	respCode, respData, err := dmja.httpClient.PutNoUnmarshal(ctx, url, headers, job)
	if err != nil {
		errDetails := fmt.Sprintf("start job '%s' failed, %s", job.JobID, err.Error())
		logger.Error(errDetails)

		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "start job failed")

		return err
	}

	if respCode != http.StatusNoContent {
		var baseError rest.BaseError
		if err := sonic.Unmarshal(respData, &baseError); err != nil {
			errDetails := fmt.Sprintf("Unmalshal baesError failed: %s", err.Error())
			logger.Error(errDetails)
			o11y.Error(ctx, errDetails)
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal baesError failed")

			return err
		}

		o11y.Error(ctx, fmt.Sprintf("%s. %v", baseError.Description, baseError.ErrorDetails))
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status code is not 200")
		return fmt.Errorf("update job '%s' failed, errDetails: %v", job.JobID, baseError.ErrorDetails)
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return nil
}
