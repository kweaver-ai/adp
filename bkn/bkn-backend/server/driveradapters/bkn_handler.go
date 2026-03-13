// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	bknsdk "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"

	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

// UploadBKN 上传 BKN tar 包并导入（外部接口）
func (r *restHandler) UploadBKN(c *gin.Context) {
	logger.Debug("Handler UploadBKN Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"上传BKN并导入", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	_, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_KnowledgeNetwork_InvalidParameter).
			WithErrorDetails("Failed to get uploaded file: " + err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	defer file.Close()

	// 验证文件类型
	if header.Header.Get("Content-Type") != "application/octet-stream" {
		// 尝试通过后缀名判断
		ext := filepath.Ext(header.Filename)
		if ext != ".tar" && ext != ".tgz" && ext != ".tar.gz" {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_KnowledgeNetwork_InvalidParameter).
				WithErrorDetails("Invalid file type, expected tar archive")
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	// 获取表单参数
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)

	logger.Debugf("Upload BKN: branch=%s, filename=%s, size=%d",
		branch, header.Filename, header.Size)

	// 直接从 tar 包加载网络（纯内存，无需写入磁盘）
	bknNetwork, err := bknsdk.LoadNetworkFromTar(file)
	if err != nil {
		logger.Errorf("Failed to load network from tar: %s", err.Error())
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_KnowledgeNetwork_InvalidParameter).
			WithErrorDetails("Failed to load network from tar: " + err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 从header中获取业务域
	businessDomain := c.GetHeader(interfaces.HTTP_HEADER_BUSINESS_DOMAIN)
	if businessDomain == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_KnowledgeNetwork_InvalidParameter_BusinessDomain).
			WithErrorDetails("Business Domain is empty")

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	bknNetwork.Branch = branch
	bknNetwork.BusinessDomain = businessDomain

	// 调用服务处理 tar 包
	knID, err := r.bs.Import(ctx, bknNetwork)
	if err != nil {
		logger.Errorf("Upload BKN failed: %s", err.Error())
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_KnowledgeNetwork_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debugf("Upload BKN completed: kn_id=%s", knID)

	rest.ReplyOK(c, http.StatusOK, map[string]string{"kn_id": knID})
}

// DownloadBKN 下载 BKN tar 包（外部接口）
func (r *restHandler) DownloadBKN(c *gin.Context) {
	logger.Debug("Handler DownloadBKN Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"下载BKN包", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	_, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 获取路径参数
	kn_id := c.Param("kn_id")
	if kn_id == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_KnowledgeNetwork_InvalidParameter).
			WithErrorDetails("kn_id is required")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 获取查询参数
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)

	logger.Debugf("Download BKN: kn_id=%s, branch=%s", kn_id, branch)

	// 调用服务导出为 tar 包
	tarData, err := r.bs.ExportToTar(ctx, kn_id, branch)
	if err != nil {
		logger.Errorf("Download BKN failed: %s", err.Error())
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_KnowledgeNetwork_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	filename := kn_id + "-" + branch + ".tar"

	logger.Debugf("Download BKN completed: filename=%s size=%d", filename, len(tarData))

	// 设置响应头
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/octet-stream", tarData)
}
