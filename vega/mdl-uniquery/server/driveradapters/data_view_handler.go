package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"

	"uniquery/common"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
)

// 数据视图数据预览（外部）
func (r *restHandler) ViewSimulateByEx(c *gin.Context) {
	logger.Debug("Handler SimulateByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: View data simulate", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ViewSimulate(c, visitor)
}

// 视图数据查询(内部)
func (r *restHandler) ViewSimulateByIn(c *gin.Context) {
	logger.Debug("Handler SimulateByIn Start")

	visitor := GenerateVisitor(c)
	r.ViewSimulate(c, visitor)
}

// 视图数据预览
func (r *restHandler) ViewSimulate(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler ViewDataSimulate Start")
	startTime := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver layer: View data simulate", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	query := interfaces.DataViewSimulateQuery{}
	err := c.ShouldBindJSON(&query)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_RequestBody).
			WithErrorDetails("Binding Parameter Failed:" + err.Error())

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	err = ValidateHeaderMethodOverride(ctx, c.GetHeader(interfaces.Headers_MethodOverride))
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 补充 sort 字段
	// query.Sort = completeSortParams(query.Sort, query.UseSearchAfter)

	// 设置默认值
	setDefaultValues(&query.ViewQueryCommonParams)

	// 视图预览的参数校验
	err = ValidateDataViewSimulate(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result, err := r.dvService.Simulate(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	result.OverallMs = int64(time.Since(startTime).Milliseconds())
	resultBytes, err := common.Marshal(c.Request.UserAgent(), result)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InvalidParameter_RequestBody).
			WithErrorDetails("Marshal query result Failed:" + err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	common.ReplyOK(c, http.StatusOK, resultBytes)
}

// 视图数据查询（外部）
func (r *restHandler) GetViewDataByEx(c *gin.Context) {
	logger.Debug("Handler GetViewDataEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver layer: Get view data",
		trace.WithSpanKind(trace.SpanKindServer))

	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetViewData(c, visitor)
}

// 视图数据查询 (内部)
func (r *restHandler) GetViewDataByIn(c *gin.Context) {
	logger.Debug("Handler GetViewDataByIn Start")

	visitor := GenerateVisitor(c)
	r.GetViewData(c, visitor)
}

// 视图数据查询 V2 版
func (r *restHandler) GetViewData(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler GetViewData Start")
	startTime := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver layer: Get view data", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	err := ValidateHeaderMethodOverride(ctx, c.GetHeader(interfaces.Headers_MethodOverride))
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//  超时参数，格式为 "30s"
	var timeoutDur time.Duration
	timeoutStr := c.Query(interfaces.QueryParam_Timeout)
	if timeoutStr != "" {
		timeoutDur, err = time.ParseDuration(timeoutStr)
		if err != nil {
			errDetails := fmt.Sprintf(`The value of param '%s' should be duration type, but got '%s'`, interfaces.QueryParam_Timeout, timeoutStr)

			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails(errDetails)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	includeViewParam := c.DefaultQuery(interfaces.QueryParam_IncludeView, "false")
	includeView, err := strconv.ParseBool(includeViewParam)
	if err != nil {
		errDetails := fmt.Sprintf(`The value of param '%s' should be bool type, but got '%s'`, interfaces.QueryParam_IncludeView, includeViewParam)

		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_IncludeView).
			WithErrorDetails(errDetails)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	allowNonExistFieldParam := c.DefaultQuery(interfaces.QueryParam_AllowNonExistField, "false")
	allowNonExistField, err := strconv.ParseBool(allowNonExistFieldParam)
	if err != nil {
		errDetails := fmt.Sprintf(`The value of param '%s' should be bool type, but got '%s'`, interfaces.QueryParam_AllowNonExistField, allowNonExistFieldParam)
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_AllowNonExistField).
			WithErrorDetails(errDetails)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	ids := c.Param("view_ids")
	if ids == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_ViewIDs).
			WithErrorDetails("View id is empty")

		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	idsArr := convert.StringToStringSlice(ids)

	// 单个查询
	if len(idsArr) == 1 {
		query := interfaces.DataViewQueryV2{}
		err := c.ShouldBindJSON(&query)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_RequestBody).
				WithErrorDetails("Binding Parameter Failed:" + err.Error())

			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		viewID := idsArr[0]
		query.Timeout = timeoutDur
		query.IncludeView = includeView
		query.AllowNonExistField = allowNonExistField
		// query.Sort = completeSortParams(query.Sort, query.UseSearchAfter)
		setDefaultValues(&query.ViewQueryCommonParams)

		// 视图查询的参数校验
		err = ValidateDataViewQueryV2(ctx, &query)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		res, err := r.dvService.GetSingleViewData(ctx, viewID, &query)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		res.OverallMs = int64(time.Since(startTime).Milliseconds())
		resBytes, err := common.Marshal(c.Request.UserAgent(), res)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InvalidParameter_RequestBody).
				WithErrorDetails("Marshal query result Failed:" + err.Error())
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		o11y.AddHttpAttrs4Ok(span, http.StatusOK)
		common.ReplyOK(c, http.StatusOK, resBytes)
		return
	} else {
		querys := []interfaces.DataViewQueryV2{}
		err := c.ShouldBindJSON(&querys)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_RequestBody).
				WithErrorDetails("Binding Parameter Failed:" + err.Error())

			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		failed := false
		results := make([]any, len(querys))

		for i := range querys {
			startTimei := time.Now()
			viewID := idsArr[i]
			querys[i].IncludeView = includeView
			querys[i].AllowNonExistField = allowNonExistField
			// querys[i].Sort = completeSortParams(querys[i].Sort, querys[i].UseSearchAfter)
			setDefaultValues(&querys[i].ViewQueryCommonParams)

			err = ValidateDataViewQueryV2(ctx, &querys[i])
			if err != nil {
				httpErr := err.(*rest.HTTPError)
				o11y.AddHttpAttrs4HttpError(span, httpErr)

				results[i] = interfaces.UniResponseError{
					StatusCode: httpErr.HTTPCode,
					BaseError:  httpErr.BaseError,
				}

				failed = true
				continue
			}

			res, err := r.dvService.GetSingleViewData(ctx, viewID, &querys[i])
			if err != nil {
				httpErr := err.(*rest.HTTPError)
				o11y.AddHttpAttrs4HttpError(span, httpErr)

				results[i] = interfaces.UniResponseError{
					StatusCode: httpErr.HTTPCode,
					BaseError:  httpErr.BaseError,
				}

				failed = true
				continue
			}

			res.OverallMs = int64(time.Since(startTimei))
			results[i] = res
			o11y.AddHttpAttrs4Ok(span, http.StatusOK)
		}

		resultsBytes, err := common.Marshal(c.Request.UserAgent(), results)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InvalidParameter_RequestBody).
				WithErrorDetails("Marshal query result Failed:" + err.Error())
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		if failed {
			common.ReplyOK(c, http.StatusMultiStatus, resultsBytes)
		} else {
			common.ReplyOK(c, http.StatusOK, resultsBytes)
		}
	}
}

// 批量删除 pits（外部）
func (r *restHandler) DeleteDataViewPitsByEx(c *gin.Context) {
	logger.Debug("Handler DeleteDataViewPitsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver layer: Delete data view pits",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.DeleteDataViewPits(c, visitor)
}

// 批量删除 pits(内部)
func (r *restHandler) DeleteDataViewPitsByIn(c *gin.Context) {
	logger.Debug("Handler DeleteDataViewPitsByIn Start")

	visitor := GenerateVisitor(c)
	r.DeleteDataViewPits(c, visitor)
}

// 批量删除 pit
func (r *restHandler) DeleteDataViewPits(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler DeleteDataViewPits Start")

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "driver layer: Delete data view pits", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 校验 method
	headerMethod := c.GetHeader(interfaces.Headers_MethodOverride)
	if headerMethod == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_NullParameter_OverrideMethod)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if headerMethod != http.MethodDelete {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_OverrideMethod).
			WithErrorDetails(fmt.Sprintf("%s is expected to be %s, but it is actually %s", interfaces.Headers_MethodOverride, http.MethodDelete, headerMethod))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验请求体
	body := interfaces.DeletePits{}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_RequestBody).
			WithErrorDetails("Binding Parameter Failed:" + err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	res, err := r.dvService.DeleteDataViewPits(ctx, &body)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, err)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	common.ReplyOK(c, http.StatusOK, res)
}

// 设置视图查询参数的默认值
func setDefaultValues(query *interfaces.ViewQueryCommonParams) {
	if query.Limit == 0 {
		if query.UseSearchAfter {
			query.Limit = interfaces.SearchAfter_Limit
		} else {
			query.Limit = interfaces.DEFAULT_LIMIT
		}
	}

	if query.Format == "" {
		query.Format = interfaces.Format_Flat
	}

	// 如果end 大于 current_time，那么end = current
	currentTime := time.Now().UnixMilli()
	if query.End > currentTime {
		query.End = currentTime
	}
}
