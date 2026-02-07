// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
)

// 分页查询数据字典(内部)
func (r *restHandler) ListDataDictsByIn(c *gin.Context) {
	logger.Debug("Handler ListDataDictsByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.ListDataDicts(c, visitor)
}

// 分页查询数据字典（外部）
func (r *restHandler) ListDataDictsByEx(c *gin.Context) {
	logger.Debug("Handler ListDataDictsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"分页获取数据字典列表", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ListDataDicts(c, visitor)
}

// 分页查询数据字典
func (r *restHandler) ListDataDicts(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler ListDicts Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "分页获取数据字典列表",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("分页获取数据字典列表请求参数: [%s]", c.Request.RequestURI))

	// 接收参数
	namePattern := c.Query("name_pattern")
	name := c.Query("name")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.NO_LIMIT)
	sort := c.DefaultQuery("sort", interfaces.DEFAULT_SORT)
	direction := c.DefaultQuery("direction", interfaces.DESC_DIRECTION)

	tag := c.Query("tag")
	// 去掉标签前后的所有空格进行搜索
	tag = strings.Trim(tag, " ")

	// 根据类型过滤参数
	types := c.Query("type")
	typeSlice := strings.Split(types, ",")
	if len(typeSlice) > 1 || len(typeSlice) == 0 {
		types = ""
	}

	// 分页参数校验
	pageParam, err := validatePaginationQueryParameters(ctx, offset, limit, sort, direction, interfaces.DATA_DICT_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	// name_pattern 和 name 不能同时存在
	if namePattern != "" && name != "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter).
			WithErrorDetails("name_pattern and name cannot exists at the same time")
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	listQuery := interfaces.DataDictQueryParams{
		Type:        types,
		Tag:         tag,
		NamePattern: namePattern,
		Name:        name,
	}
	listQuery.Sort = pageParam.Sort
	listQuery.Direction = pageParam.Direction
	listQuery.Limit = pageParam.Limit
	listQuery.Offset = pageParam.Offset

	// 调用Service
	dicts, total, err := r.dds.ListDataDicts(ctx, listQuery)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]interface{}{
		"total_count": total,
		"entries":     dicts,
	}

	logger.Debug("Handler ListDicts Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// 批量 获取/导出 数据字典(内部)
func (r *restHandler) GetDataDictsByIn(c *gin.Context) {
	logger.Debug("Handler GetDataDictsByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.GetDataDicts(c, visitor)
}

// 批量 获取/导出 数据字典（外部）
func (r *restHandler) GetDataDictsByEx(c *gin.Context) {
	logger.Debug("Handler GetDataDictsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"获取数据字典", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetDataDicts(c, visitor)
}

// 批量 获取/导出 数据字典
func (r *restHandler) GetDataDicts(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler GetDataDicts Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "获取数据字典",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("获取数据字典请求参数: [%s]", c.Request.RequestURI))

	// 获取参数字符串 <id1,id2,id3>
	dictIDstrs := c.Param("dict_id")
	span.SetAttributes(attr.Key("dict_ids").String(dictIDstrs))

	// 解析字符串 转换为 []string
	dictIDs := common.StringToStringSlice(dictIDstrs)
	if len(dictIDs) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictIDs).
			WithErrorDetails("Type Conversion Failed:")
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	result, err := r.dds.GetDataDicts(ctx, dictIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler GetDataDicts Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// 创建数据字典(内部)
func (r *restHandler) CreateDataDictsByIn(c *gin.Context) {
	logger.Debug("Handler CreateDataDictsByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.CreateDataDicts(c, visitor)
}

// 创建数据字典（外部）
func (r *restHandler) CreateDataDictsByEx(c *gin.Context) {
	logger.Debug("Handler CreateDataDictsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"创建数据字典", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.CreateDataDicts(c, visitor)
}

// 创建数据字典
func (r *restHandler) CreateDataDicts(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler CreateDataDict Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "创建数据字典",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 接收绑定参数
	dataDicts := []interfaces.DataDict{}
	err := c.ShouldBindJSON(&dataDicts)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject("", ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("创建数据字典请求参数: [%s,%v]", c.Request.RequestURI, dataDicts))

	// 校验 请求体中字典合法性
	// 请求体字典名称
	dictNameArr := []string{}
	for i := 0; i < len(dataDicts); i++ {
		// 适配旧json
		if dataDicts[i].OldDictName != "" {
			dataDicts[i].DictName = dataDicts[i].OldDictName
			dataDicts[i].OldDictName = ""
		}
		if len(dataDicts[i].OldDictItems) != 0 {
			dataDicts[i].DictItems = dataDicts[i].OldDictItems
			dataDicts[i].OldDictItems = nil
		}

		// 校验字典名称合法性、维度限制
		err = ValidateDict(ctx, dataDicts[i])
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataDictAuditObject("", dataDicts[i].DictName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Validate data dict[%s] failed: %s. %v", dataDicts[i].DictName,
				httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			span.SetAttributes(attr.Key("dict_name").String(dataDicts[i].DictName))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		// 校验comment合法性
		err = validateObjectComment(ctx, dataDicts[i].Comment)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataDictAuditObject("", dataDicts[i].DictName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Validate data dict[%s] failed: %s. %v", dataDicts[i].DictName,
				httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			span.SetAttributes(attr.Key("dict_name").String(dataDicts[i].DictName))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		// 去掉tag前后空格以及数组去重
		dataDicts[i].Tags = libCommon.TagSliceTransform(dataDicts[i].Tags)
		// 校验标签
		err = validateObjectTags(ctx, dataDicts[i].Tags)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataDictAuditObject("", dataDicts[i].DictName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Validate data dict[%s] failed: %s. %v", dataDicts[i].DictName,
				httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			span.SetAttributes(attr.Key("dict_name").String(dataDicts[i].DictName))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		// 分类校验字典项
		switch dataDicts[i].DictType {
		case interfaces.DATA_DICT_TYPE_DIMENSION:
			httpErr := validateDimensionDictItems(ctx, dataDicts[i])
			if httpErr != nil {

				audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
					GenerateDataDictAuditObject("", dataDicts[i].DictName), &httpErr.BaseError)

				// 记录异常日志
				o11y.Error(ctx, fmt.Sprintf("Validate data dict[%s] failed: %s. %v", dataDicts[i].DictName,
					httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

				// 设置 trace 的错误信息的 attributes
				span.SetAttributes(attr.Key("dict_name").String(dataDicts[i].DictName))
				o11y.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}

		default:
			httpErr := validateKVDictItems(ctx, dataDicts[i])
			if httpErr != nil {

				audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
					GenerateDataDictAuditObject("", dataDicts[i].DictName), &httpErr.BaseError)

				// 记录异常日志
				o11y.Error(ctx, fmt.Sprintf("Validate data dict[%s] failed: %s. %v", dataDicts[i].DictName,
					httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

				// 设置 trace 的错误信息的 attributes
				span.SetAttributes(attr.Key("dict_name").String(dataDicts[i].DictName))
				o11y.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}

		dictNameArr = append(dictNameArr, dataDicts[i].DictName)
	}

	// 校验 请求体自身中字典名称重复性
	err = ValidateDuplicate(ctx, dictNameArr)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_Duplicated_DictName).
			WithErrorDetails("Dictionary Name duplicated:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject("", strings.Join(dictNameArr, ",")), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Validate data dicts [%v] failed: %s. %v", dictNameArr,
			httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		span.SetAttributes(attr.Key("dicts").StringSlice(dictNameArr))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验 请求体与现有字典名称的重复性
	for i := 0; i < len(dataDicts); i++ {
		// 校验字典名称是否已存在
		_, err := r.dds.CheckDictExistByName(ctx, dataDicts[i].DictName)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataDictAuditObject("", dataDicts[i].DictName), &httpErr.BaseError)

			// 设置 trace 的错误信息的 attributes
			span.SetAttributes(attr.Key("dict_name").String(dataDicts[i].DictName))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	// 调用创建
	for i := 0; i < len(dataDicts); i++ {
		dictID, err := r.dds.CreateDataDict(ctx, dataDicts[i])
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataDictAuditObject(dictID, dataDicts[i].DictName), &httpErr.BaseError)

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		dataDicts[i].DictID = dictID
		// 每次成功创建 记录审计日志
		audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject(dictID, dataDicts[i].DictName), "")
	}

	result := []map[string]interface{}{}
	for i := 0; i < len(dataDicts); i++ {
		result = append(result, map[string]interface{}{
			"dict_id":   dataDicts[i].DictID,
			"dict_name": dataDicts[i].DictName,
		})
	}

	logger.Debug("Handler CreateDataDict Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// 更新单个数据字典(内部)
func (r *restHandler) UpdateDataDictByIn(c *gin.Context) {
	logger.Debug("Handler UpdateDataDictByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.UpdateDataDict(c, visitor)
}

// 更新单个数据字典（外部）
func (r *restHandler) UpdateDataDictByEx(c *gin.Context) {
	logger.Debug("Handler UpdateDataDictByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"修改数据字典", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.UpdateDataDict(c, visitor)
}

// 更新单个数据字典
func (r *restHandler) UpdateDataDict(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler UpdateDataDict Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "修改数据字典",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 接受dict_id参数
	dictID := c.Param("dict_id")
	if dictID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictIDs).
			WithErrorDetails("dict_id is empty")

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject("", ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 接收绑定参数
	dict := interfaces.DataDict{}
	err := c.ShouldBindJSON(&dict)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_Dict).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject(dictID, ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	dict.DictID = dictID

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("修改数据字典请求参数: [%s, %v]", c.Request.RequestURI, dict))

	// 校验字典名称合法性
	err = ValidateDict(ctx, dict)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject(dictID, dict.DictName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Validate data dict[%s] failed: %s. %v", dict.DictName,
			httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		span.SetAttributes(attr.Key("dict_name").String(dict.DictName))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验comment合法性
	err = validateObjectComment(ctx, dict.Comment)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject(dictID, dict.DictName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Validate data dict[%s] failed: %s. %v", dict.DictName,
			httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 去掉tag前后空格以及数组去重
	dict.Tags = libCommon.TagSliceTransform(dict.Tags)
	// 校验标签
	err = validateObjectTags(ctx, dict.Tags)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject(dictID, dict.DictName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Validate data dict[%s] failed: %s. %v", dict.DictName,
			httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验维度名称重复性
	// 请求体维度名称
	if dict.DictType == interfaces.DATA_DICT_TYPE_DIMENSION {
		dimensionArr := []string{}
		for _, k := range dict.Dimension.Keys {
			dimensionArr = append(dimensionArr, k.Name)
		}

		for _, v := range dict.Dimension.Values {
			dimensionArr = append(dimensionArr, v.Name)
		}

		err = ValidateDuplicate(ctx, dimensionArr)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_Duplicated_DictDimension).
				WithErrorDetails("Dictionary Dimension duplicated:" + err.Error())

			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateDataDictAuditObject(dictID, dict.DictName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Validate data dict[%s] failed: %s. %v", dict.DictName,
				httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	// 根据id修改信息
	err = r.dds.UpdateDataDict(ctx, dict)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject(dictID, dict.DictName), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		GenerateDataDictAuditObject(dictID, dict.DictName), "")

	logger.Debug("Handler UpdateDataDict Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 批量删除数据字典
func (r *restHandler) DeleteDataDicts(c *gin.Context) {
	logger.Debug("Handler DeleteDataDicts Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "删除数据字典",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("删除数据字典请求参数: [%s]", c.Request.RequestURI))

	// 获取参数字符串 <id1,id2,id3>
	dictIDsStr := c.Param("dict_id")

	span.SetAttributes(attr.Key("dict_ids").String(dictIDsStr))

	// 解析字符串 转换为 []string
	dictIDs := common.StringToStringSlice(dictIDsStr)
	if len(dictIDs) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictIDs).
			WithErrorDetails("dictIDs is empty")

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject("", ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 检查dictIDs是否都存在
	var dicts []interfaces.DataDict
	for i := 0; i < len(dictIDs); i++ {
		dict, err := r.dds.GetDataDictByID(ctx, dictIDs[i])
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataDict_DictNotFound).
				WithErrorDetails("Dictionary " + dictIDs[i] + " not found!")

			audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				GenerateDataDictAuditObject(dictIDs[i], ""), &httpErr.BaseError)

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		dicts = append(dicts, dict)
	}

	// 循环删除数据字典
	for i := 0; i < len(dicts); i++ {
		rowsAffect, err := r.dds.DeleteDataDict(ctx, dicts[i])
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				GenerateDataDictAuditObject(dicts[i].DictID, dicts[i].DictName), &httpErr.BaseError)

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if rowsAffect != 0 {
			audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				GenerateDataDictAuditObject(dicts[i].DictID, dicts[i].DictName), audit.SUCCESS, "")
		}
	}

	logger.Debug("Handler DeleteDataDicts Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 分页获取数据字典资源列表
func (r *restHandler) ListDataDictSrcs(c *gin.Context) {
	logger.Debug("Handler ListDataDictSrcs Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "分页获取数据字典列表",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("分页获取数据字典列表请求参数: [%s]", c.Request.RequestURI))

	// 接收参数
	namePattern := c.Query(RESOURCES_KEYWOED)
	name := c.Query("name")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", RESOURCES_PAGE_LIMIT)
	sort := c.DefaultQuery("sort", "name")
	direction := c.DefaultQuery("direction", interfaces.ASC_DIRECTION)

	tag := c.Query("tag")
	// 去掉标签前后的所有空格进行搜索
	tag = strings.Trim(tag, " ")

	// 根据类型过滤参数
	types := c.Query("type")
	typeSlice := strings.Split(types, ",")
	if len(typeSlice) > 1 || len(typeSlice) == 0 {
		types = ""
	}

	// 分页参数校验
	pageParam, err := validatePaginationQueryParameters(ctx, offset, limit, sort, direction, interfaces.DATA_DICT_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	// name_pattern 和 name 不能同时存在
	if namePattern != "" && name != "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter).
			WithErrorDetails("name_pattern and name cannot exists at the same time")
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	listQuery := interfaces.DataDictQueryParams{
		Type:        types,
		Tag:         tag,
		NamePattern: namePattern,
		Name:        name,
	}
	listQuery.Sort = pageParam.Sort
	listQuery.Direction = pageParam.Direction
	listQuery.Limit = pageParam.Limit
	listQuery.Offset = pageParam.Offset

	// 调用Service
	dicts, total, err := r.dds.ListDataDictSrcs(ctx, listQuery)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]interface{}{
		"total_count": total,
		"entries":     dicts,
	}

	logger.Debug("Handler ListDataDictSrcs Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}
