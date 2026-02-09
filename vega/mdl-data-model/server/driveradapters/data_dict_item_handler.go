// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/xuri/excelize/v2"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
)

const (
	FILE_FORMAT_JSON = "json"
	FILE_FORMAT_CSV  = "csv"
	FILE_FORMAT_XLSX = "xlsx"
)

// 按 id 获取目标模型对象信息(内部)
func (r *restHandler) HandleDataDictListOrExportByIn(c *gin.Context) {
	logger.Debug("Handler HandleDataDictListOrExportByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.HandleDataDictListOrExport(c, visitor)
}

// 按 id 获取目标模型对象信息（外部）
func (r *restHandler) HandleDataDictListOrExportByEx(c *gin.Context) {
	logger.Debug("Handler HandleDataDictListOrExportByEx Start")

	// 校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.HandleDataDictListOrExport(c, visitor)
}

func (r *restHandler) HandleDataDictListOrExport(c *gin.Context, visitor rest.Visitor) {
	requestFormat := c.DefaultQuery("format", FILE_FORMAT_JSON)
	switch requestFormat {
	case FILE_FORMAT_JSON:
		r.ListDataDictItems(c, visitor)
	case FILE_FORMAT_CSV:
		fallthrough
	case FILE_FORMAT_XLSX:
		fallthrough
	default:
		r.ExportDataDictItems(c, requestFormat, visitor)
	}
}

// 创建或导入(内部)
func (r *restHandler) HandleDataDictCreateOrImportByIn(c *gin.Context) {
	logger.Debug("Handler HandleDataDictCreateOrImportByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.HandleDataDictCreateOrImport(c, visitor)
}

// 创建或导入（外部）
func (r *restHandler) HandleDataDictCreateOrImportByEx(c *gin.Context) {
	logger.Debug("Handler HandleDataDictCreateOrImportByEx Start")

	// 校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.HandleDataDictCreateOrImport(c, visitor)
}

func (r *restHandler) HandleDataDictCreateOrImport(c *gin.Context, visitor rest.Visitor) {
	if c.ContentType() != rest.ContentTypeJson {
		r.ImportDataDictItems(c, visitor)
	} else {
		r.CreateDataDictItem(c, visitor)
	}
}

// 分页查询数据字典项
func (r *restHandler) ListDataDictItems(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler GetDataDictItems Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "分页获取数据字典项列表",
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

	dictID := c.Param("dict_id")
	if dictID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictID).
			WithErrorDetails("Type Conversion Failed:")
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	// 接收参数
	field := c.Query("query_field")
	pattern := c.Query("query_pattern")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", interfaces.DATA_DICT_ITEM_DEFAULT_SORT)
	direction := c.DefaultQuery("direction", interfaces.ASC_DIRECTION)

	// 分页参数校验
	pageParam, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.DATA_DICT_ITEM_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	listQuery := interfaces.DataDictItemQueryParams{
		Patterns: []interfaces.DataDictItemQueryPattern{
			{
				QueryField:   field,
				QueryPattern: pattern,
			},
		},
	}
	listQuery.Sort = pageParam.Sort
	listQuery.Direction = pageParam.Direction
	listQuery.Limit = pageParam.Limit
	listQuery.Offset = pageParam.Offset

	// 检查数据字典是否存在
	dict, err := r.dds.GetDataDictByID(ctx, dictID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		httpErr = rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataDict_DictNotFound).
			WithErrorDetails("Dictionary " + dictID + " not found!")
		rest.ReplyError(c, httpErr)
		return
	}

	// 调用Service
	dictItems, total, err := r.ddis.ListDataDictItems(ctx, dict, listQuery)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]interface{}{
		"total_count": total,
		"entries":     dictItems,
	}

	logger.Debug("Handler ListDictItems Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// 创建单个数据字典项
func (r *restHandler) CreateDataDictItem(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler CreateDataDictItem Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "创建数据字典项",
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

	// 接受dict_id
	dictID := c.Param("dict_id")
	if dictID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictID).
			WithErrorDetails("Type Conversion Failed:")

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataDictItemAuditObject("", ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 检查数据字典是否存在
	dict, err := r.dds.GetDataDictByID(ctx, dictID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataDict_DictNotFound).
			WithErrorDetails("Dictionary " + dictID + " not found!")

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataDictItemAuditObject(dictID, ""), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 接受绑定参数
	dataDictItem := map[string]string{}
	err = c.ShouldBindJSON(&dataDictItem)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataDict_InvalidParameter_DictItems).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataDictItemAuditObject(dictID, ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if dataDictItem[interfaces.DATA_DICT_DIMENSION_NAME_COMMENT] != "" {
		dict.Dimension.Comment = dataDictItem[interfaces.DATA_DICT_DIMENSION_NAME_COMMENT]
	}
	// 校验comment是否合法
	err = validateObjectComment(ctx, dict.Dimension.Comment)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataDictItemAuditObject(dictID, dict.DictName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 判断类型
	switch dict.DictType {
	case interfaces.DATA_DICT_TYPE_KV:
		err = validateKVItem(ctx, dataDictItem, dict)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataDictItemAuditObject(dictID, dict.DictName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, err)
			return
		}

	case interfaces.DATA_DICT_TYPE_DIMENSION:
		err = validateDimensionItem(ctx, dataDictItem, dict)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataDictItemAuditObject(dictID, dict.DictName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, err)
			return
		}
	}

	// 调用service
	// 将赋值过的维度结构体传下去
	itemID, err := r.ddis.CreateDataDictItem(ctx, dict, dict.Dimension)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataDictItemAuditObject(dictID, dict.DictName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		GenerateDataDictItemAuditObject(dictID, dict.DictName), "")

	result := map[string]interface{}{
		"dict_id": dictID,
		"item_id": itemID,
	}

	logger.Debug("Handler CreateDataDictItem Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// csv导入数据字典项
func (r *restHandler) ImportDataDictItems(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler ImportDataDictItems Start")
	beginTime := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "导入数据字典项",
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

	// 接受dict_id
	dictID := c.Param("dict_id")
	if dictID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictID).
			WithErrorDetails("dict_id is empty")

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataDictItemAuditObject("", ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 查询参数
	mode := c.DefaultQuery("import_mode", interfaces.ImportMode_Normal)
	httpErr := validateImportMode(ctx, mode)
	if httpErr != nil {
		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataDictItemAuditObject(dictID, ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 检查数据字典是否存在
	dict, err := r.dds.GetDataDictByID(ctx, dictID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataDict_DictNotFound).
			WithErrorDetails("Dictionary " + dictID + " not found!")

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataDictItemAuditObject(dictID, ""), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	fname := c.PostForm("file_name")
	ext := filepath.Ext(fname)
	if ext != "" && ext[0] == '.' {
		ext = ext[1:]
	}
	if ext != FILE_FORMAT_CSV && ext != FILE_FORMAT_XLSX {

		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_FileExt).
			WithErrorDetails("Import file's ext is invalid: " + fname)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataDictItemAuditObject(dictID, dict.DictName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	f, err := c.FormFile("items_file")
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictItemsFile).
			WithErrorDetails("Get Form File Failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataDictItemAuditObject(dictID, dict.DictName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	file, err := f.Open()
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictItemsFile).
			WithErrorDetails("Open File Failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataDictItemAuditObject(dictID, dict.DictName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	var itemsArr []map[string]string
	if ext == FILE_FORMAT_CSV {
		itemsArr, err = r.ParseDataDictItemsFromCSV(ctx, dict, reader)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_BadRequest_ParseCSVFileFailed).
				WithErrorDetails("import csv file failed:" + err.Error())

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataDictItemAuditObject(dictID, dict.DictName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	} else if ext == FILE_FORMAT_XLSX {
		itemsArr, err = r.ParseDataDictItemsFromXlsx(ctx, dict, reader)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_BadRequest_ParseXlsxFileFailed).
				WithErrorDetails("import xlsx file failed:" + err.Error())

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataDictItemAuditObject(dictID, dict.DictName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	dict.DictItems = itemsArr
	// 校验自身
	switch dict.DictType {
	case interfaces.DATA_DICT_TYPE_DIMENSION:
		httpErr := validateDimensionDictItems(ctx, dict)
		if httpErr != nil {
			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataDictItemAuditObject(dictID, dict.DictName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Validate data dict[%s] failed: %s. %v", dict.DictName,
				httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

	default:
		httpErr := validateKVDictItems(ctx, dict)
		if httpErr != nil {
			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateDataDictItemAuditObject(dictID, dict.DictName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Validate data dict[%s] failed: %s. %v", dict.DictName,
				httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	err = r.ddis.ImportDataDictItems(ctx, &dict, itemsArr, mode)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateDataDictItemAuditObject(dictID, dict.DictName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		GenerateDataDictItemAuditObject(dictID, dict.DictName), "")

	logger.Debug("Handler ImportDataDictItems Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)

	took := fmt.Sprint(time.Since(beginTime).Milliseconds())
	c.Writer.Header().Set(interfaces.HTTP_HEADER_REQUEST_TOOK, took)
	rest.ReplyOK(c, http.StatusCreated, nil)
}

func (r *restHandler) ParseDataDictItemsFromCSV(ctx context.Context,
	dict interfaces.DataDict, reader *bufio.Reader) ([]map[string]string, error) {

	// 第一行 即表头
	firstRow, err := reader.ReadString('\n')
	if err != nil {
		return []map[string]string{}, err
	}
	firstRowString := strings.TrimRight(firstRow, "\r\n")
	// 去除bom头
	firstRowString = strings.TrimLeft(firstRowString, "\ufeff")
	headerRow := strings.Split(firstRowString, ",")
	if len(headerRow) <= 0 {
		err := fmt.Errorf("invalid File: Empty! ")
		return []map[string]string{}, err
	}
	logger.Debugf("File Header: %+v", headerRow)

	// 建立维度键的名称-下标map
	headerMap := map[string]int{}
	for _, item := range dict.Dimension.Keys {
		for idx, header := range headerRow {
			if item.Name == header {
				headerMap[header] = idx
			}
		}
	}
	logger.Debugf("Data dict: %+v", dict)
	logger.Debugf("Header Map: %+v", headerMap)
	// 所有key都应该存在于csv文件中
	if len(headerMap) < len(dict.Dimension.Keys) {
		err := fmt.Errorf("invalid File: Lack of Dimension Keys! ")
		return []map[string]string{}, err
	}

	// 维度属性 添加到map
	for _, item := range dict.Dimension.Values {
		found := false
		for idx, header := range headerRow {
			if item.Name == header {
				headerMap[header] = idx
				found = true
			}
		}
		if !found && dict.DictType == interfaces.DATA_DICT_TYPE_KV {
			err := fmt.Errorf("invalid File: Lack of Dimension value! ")
			return []map[string]string{}, err
		}
	}

	// comment 添加到map
	for idx, header := range headerRow {
		if header == interfaces.DATA_DICT_DIMENSION_NAME_COMMENT {
			headerMap[interfaces.DATA_DICT_DIMENSION_NAME_COMMENT] = idx
		}
	}

	// 继续读文件
	itemMaps := []map[string]string{}
	for {
		rowString, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			err = fmt.Errorf("read File Failed: %s", err.Error())
			return []map[string]string{}, err
		}
		// 每一行数据
		rowString = strings.TrimRight(rowString, "\r\n")
		// 按行读 无内容 认为文件结束
		if len(rowString) <= 0 {
			logger.Warn("Empty Line")
			break
		}

		row := strings.Split(rowString, ",")
		// 分割后 与 表头数目不一致 补齐内容
		if len(row) < len(headerRow) {
			row = append(row, make([]string, len(headerRow)-len(row))...)
		}

		itemMap := map[string]string{}
		for header, idx := range headerMap {
			itemMap[header] = row[idx]
		}
		itemMaps = append(itemMaps, itemMap)
		if err == io.EOF {
			logger.Debug("Read File EOF")
			break
		}
	}

	if len(itemMaps) == 0 {
		err := fmt.Errorf("invalid File: No rows")
		return []map[string]string{}, err
	}

	return itemMaps, nil
}

func (r *restHandler) ParseDataDictItemsFromXlsx(ctx context.Context,
	dict interfaces.DataDict, reader *bufio.Reader) ([]map[string]string, error) {

	// 第一行 即表头
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return []map[string]string{}, err
	}

	if len(f.GetSheetList()) != 1 {
		return []map[string]string{}, fmt.Errorf("invalid xlsx file: only support 1 sheet")
	}

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return []map[string]string{}, err
	}
	if len(rows) <= 1 {
		err := fmt.Errorf("invalid File: No rows")
		return []map[string]string{}, err
	}

	headerRow := rows[0]
	logger.Debugf("File Header: %+v", headerRow)

	// 建立维度键的名称-下标map
	headerMap := map[string]int{}
	for _, item := range dict.Dimension.Keys {
		for idx, header := range headerRow {
			if item.Name == header {
				headerMap[header] = idx
			}
		}
	}
	logger.Debugf("Data dict: %+v", dict)
	logger.Debugf("Index Map: %+v", headerMap)
	// 不是所有key都存在于csv
	if len(headerMap) < len(dict.Dimension.Keys) {
		err := fmt.Errorf("invalid File: Lack of Dimension Keys! ")
		return []map[string]string{}, err
	}

	// 维度属性 添加到map
	for _, item := range dict.Dimension.Values {
		found := false
		for idx, header := range headerRow {
			if item.Name == header {
				headerMap[header] = idx
				found = true
			}
		}
		if !found && dict.DictType == interfaces.DATA_DICT_TYPE_KV {
			err := fmt.Errorf("invalid File: Lack of Dimension value! ")
			return []map[string]string{}, err
		}
	}

	// comment 添加到map
	for idx, header := range headerRow {
		if header == interfaces.DATA_DICT_DIMENSION_NAME_COMMENT {
			headerMap[interfaces.DATA_DICT_DIMENSION_NAME_COMMENT] = idx
		}
	}

	// 继续读文件
	itemMaps := []map[string]string{}
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		// 分割后 与 表头数目不一致 补齐内容
		if len(row) < len(headerRow) {
			row = append(row, make([]string, len(headerRow)-len(row))...)
		}

		itemMap := map[string]string{}
		for header, idx := range headerMap {
			itemMap[header] = row[idx]
		}
		itemMaps = append(itemMaps, itemMap)
	}

	return itemMaps, nil
}

// 文件导出数据字典项
func (r *restHandler) ExportDataDictItems(c *gin.Context, requestFormat string, visitor rest.Visitor) {
	logger.Debug("Handler ExportDataDictItems Start")
	beginTime := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "导出数据字典项",
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
	span.SetAttributes(attr.Key("format").String(requestFormat))

	if requestFormat != FILE_FORMAT_CSV && requestFormat != FILE_FORMAT_XLSX {

		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_Format).
			WithErrorDetails("Export Format is invalid: " + requestFormat)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 接受dict_id
	dictID := c.Param("dict_id")
	if dictID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			derrors.DataModel_DataDict_InvalidParameter_DictID).WithErrorDetails("Type Conversion Failed:")

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 检查数据字典是否存在
	dict, err := r.dds.GetDataDictByID(ctx, dictID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataDict_DictNotFound).
			WithErrorDetails("Dictionary " + dictID + " not found!")

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	var result []map[string]string
	switch dict.DictType {
	case interfaces.DATA_DICT_TYPE_DIMENSION:
		itemsMap, err := r.ddis.GetDimensionDictItems(ctx, dict.DictID, dict.DictStore, dict.Dimension)
		result = itemsMap
		logger.Debugf("itemsMap: %+v \n", itemsMap)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

	default:
		itemsMap, err := r.ddis.GetKVDictItems(ctx, dictID)
		result = itemsMap
		logger.Debugf("itemsMap: %+v \n", itemsMap)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}
	logger.Debugf("Get items result: %+v \n", result)

	var buf []byte
	if requestFormat == FILE_FORMAT_CSV {
		buf, err = r.ExportDataDictItemsToCSV(dict, result)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_DataDict_InternalError_ExportCSVFileFailed).
				WithErrorDetails("export csv file failed:" + err.Error())

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	} else if requestFormat == FILE_FORMAT_XLSX {
		buf, err = r.ExportDataDictItemsToXLSX(dict, result)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
				derrors.DataModel_DataDict_InternalError_ExportXlsxFileFailed).
				WithErrorDetails("export excel file failed:" + err.Error())

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	logger.Debug("Handler ExportDataDictItems Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)

	took := fmt.Sprint(time.Since(beginTime).Milliseconds())
	c.Writer.Header().Set(interfaces.HTTP_HEADER_REQUEST_TOOK, took)
	c.Writer.Header().Set("Content-Transfer-Encoding", "binary")
	c.Data(http.StatusOK, "application/octet-stream", buf)
}

// csv导出数据字典项
func (r *restHandler) ExportDataDictItemsToCSV(dict interfaces.DataDict, results []map[string]string) ([]byte, error) {
	// 拼接首行
	headers := make([]string, 0, len(dict.Dimension.Keys)+len(dict.Dimension.Values)+1)
	for i := range dict.Dimension.Keys {
		headers = append(headers, dict.Dimension.Keys[i].Name)
	}
	for j := range dict.Dimension.Values {
		headers = append(headers, dict.Dimension.Values[j].Name)
	}
	headers = append(headers, interfaces.DATA_DICT_DIMENSION_NAME_COMMENT)

	var sb strings.Builder
	sb.WriteString(strings.Join(headers, ","))
	sb.WriteString("\r\n")

	row := make([]string, len(headers))
	if len(results) > 0 {
		for _, result := range results {
			for j, header := range headers {
				row[j] = result[header]
			}

			sb.WriteString(strings.Join(row, ","))
			sb.WriteString("\r\n")
		}
	}

	buf := []byte(sb.String())
	return buf, nil
}

// xlsx导出数据字典项
func (r *restHandler) ExportDataDictItemsToXLSX(dict interfaces.DataDict, results []map[string]string) ([]byte, error) {
	// 拼接首行
	headers := make([]string, 0, len(dict.Dimension.Keys)+len(dict.Dimension.Values)+1)
	for i := range dict.Dimension.Keys {
		headers = append(headers, dict.Dimension.Keys[i].Name)
	}
	for j := range dict.Dimension.Values {
		headers = append(headers, dict.Dimension.Values[j].Name)
	}
	headers = append(headers, interfaces.DATA_DICT_DIMENSION_NAME_COMMENT)

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			logger.Error(err)
		}
	}()

	err := f.SetDefaultFont("宋体")
	if err != nil {
		return []byte{}, err
	}

	sheetName := "Sheet1"
	err = f.SetSheetRow(sheetName, "A1", &headers)
	if err != nil {
		return []byte{}, err
	}

	row := make([]string, len(headers))
	if len(results) > 0 {
		for i, result := range results {
			for j, header := range headers {
				row[j] = result[header]
			}

			rowCell := fmt.Sprintf("A%d", i+2)
			err = f.SetSheetRow(sheetName, rowCell, &row)
			if err != nil {
				return []byte{}, err
			}
		}
	}

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return []byte{}, err
	}

	buf := buffer.Bytes()
	return buf, nil
}

// 按 id 获取目标模型对象信息(内部)
func (r *restHandler) UpdateDataDictItemByIn(c *gin.Context) {
	logger.Debug("Handler UpdateDataDictItemByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.UpdateDataDictItem(c, visitor)
}

// 按 id 获取目标模型对象信息（外部）
func (r *restHandler) UpdateDataDictItemByEx(c *gin.Context) {
	logger.Debug("Handler UpdateDataDictItemByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"修改数据字典项", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.UpdateDataDictItem(c, visitor)
}

// 更新单个数据字典项
func (r *restHandler) UpdateDataDictItem(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler UpdateDataDictItem Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "修改数据字典项",
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

	// 接受_id参数
	dictID := c.Param("dict_id")
	itemID := c.Param("item_id")

	if dictID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictID).
			WithErrorDetails("dictID is empty")

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject("", ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if itemID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_ItemID).
			WithErrorDetails("itemID is empty")

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject(dictID, ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 检查数据字典是否存在
	dict, err := r.dds.GetDataDictByID(ctx, dictID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_DataDict_DictNotFound).
			WithErrorDetails("Dictionary " + dictID + " not found!")

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject(dictID, ""), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 接受绑定参数
	dataDictItem := map[string]string{}
	err = c.ShouldBindJSON(&dataDictItem)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictItems).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject(dictID, dict.DictName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if dataDictItem[interfaces.DATA_DICT_DIMENSION_NAME_COMMENT] != "" {
		dict.Dimension.Comment = dataDictItem[interfaces.DATA_DICT_DIMENSION_NAME_COMMENT]
	}
	// 校验comment是否合法
	err = validateObjectComment(ctx, dict.Dimension.Comment)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject(dictID, dict.DictName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 判断类型
	switch dict.DictType {
	case interfaces.DATA_DICT_TYPE_KV:
		err = validateKVItem(ctx, dataDictItem, dict)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateDataDictAuditObject(dictID, dict.DictName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, err)
			return
		}

	case interfaces.DATA_DICT_TYPE_DIMENSION:
		err = validateDimensionItem(ctx, dataDictItem, dict)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateDataDictAuditObject(dictID, dict.DictName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, err)
			return
		}
	}

	// 调用service修改字典项
	err = r.ddis.UpdateDataDictItem(ctx, dict, itemID, dict.Dimension)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject(dictID, dict.DictName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		GenerateDataDictAuditObject(dictID, dict.DictName), "")

	logger.Debug("Handler UpdateDataDictItem Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 删除多个数据字典项
func (r *restHandler) DeleteDataDictItems(c *gin.Context) {
	logger.Debug("Handler DeleteDataDictItems Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "删除数据字典项",
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

	// 获取item_ids
	dictID := c.Param("dict_id")
	itemIDsStr := c.Param("item_ids")

	if dictID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_DictID).
			WithErrorDetails("dictID is empty")

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject("", ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 解析字符串 转换为 []string
	itemIDs := common.StringToStringSlice(itemIDsStr)
	if len(itemIDs) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_DataDict_InvalidParameter_ItemIDs).
			WithErrorDetails("itemIDs is empty")

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject(itemIDsStr, ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 检查数据字典是否存在
	dict, err := r.dds.GetDataDictByID(ctx, dictID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject(dictID, ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 检查数据字典项是否都存在
	_, err = r.ddis.GetDictItemsByItemIDs(ctx, dict.DictStore, itemIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateDataDictAuditObject(dictID, dict.DictName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 循环删除
	for i := 0; i < len(itemIDs); i++ {
		err := r.ddis.DeleteDataDictItem(ctx, dict, itemIDs[i])
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				GenerateDataDictAuditObject(dictID, dict.DictName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
		GenerateDataDictAuditObject(dictID, dict.DictName), audit.SUCCESS, "")

	logger.Debug("Handler DeleteDataDictItems Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}
