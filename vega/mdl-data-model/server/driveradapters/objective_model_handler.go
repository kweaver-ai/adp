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
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
)

// 创建目标模型(内部)
func (r *restHandler) CreateObjectiveModelsByIn(c *gin.Context) {
	logger.Debug("Handler CreateObjectiveModelsByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.CreateObjectiveModels(c, visitor)
}

// 创建目标模型（外部）
func (r *restHandler) CreateObjectiveModelsByEx(c *gin.Context) {
	logger.Debug("Handler CreateObjectiveModelsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"创建目标模型", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.CreateObjectiveModels(c, visitor)
}

// 创建目标模型
func (r *restHandler) CreateObjectiveModels(c *gin.Context, visitor rest.Visitor) {
	//创建时新增参数groupName ，需要进行校验
	//请求体以及数据库表中重复名称校验需要groupName 和modelName
	//需要通过groupName更新groupID才能创建
	logger.Debug("Handler CreateObjectiveModels Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"创建目标模型", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 查询参数
	mode := c.DefaultQuery(interfaces.QueryParam_ImportMode, interfaces.ImportMode_Normal)
	httpErr := validateImportMode(ctx, mode)
	if httpErr != nil {
		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateObjectiveModelAuditObject("", ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 接受绑定参数
	modelsReqBody := []interfaces.CreateObjectiveModel{}
	err := c.ShouldBindJSON(&modelsReqBody)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateObjectiveModelAuditObject("", ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("创建目标模型请求参数: [%s,%v]", c.Request.RequestURI, modelsReqBody))

	createModels := make([]*interfaces.ObjectiveModel, 0)
	// 校验 请求体中目标模型名称合法性
	tmpNameMap := make(map[string]any)
	idMap := make(map[string]any)
	for i := 0; i < len(modelsReqBody); i++ {
		// 0.校验请求体中多个模型 ID 是否重复
		modelID := modelsReqBody[i].ModelID
		if _, ok := idMap[modelID]; !ok {
			idMap[modelID] = nil
		} else {
			errDetails := fmt.Sprintf("ObjectiveModel ID '%s' already exists in the file", modelID)
			httpErr := rest.NewHTTPError(ctx, http.StatusForbidden,
				derrors.DataModel_MetricModel_Duplicated_ModelIDInFile).
				WithDescription(map[string]any{"ModelID": modelID}).
				WithErrorDetails(errDetails)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateObjectiveModelAuditObject(modelsReqBody[i].ModelID, modelsReqBody[i].ModelName), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		model := interfaces.ObjectiveModel{
			ObjectiveModelInfo: modelsReqBody[i].ObjectiveModelInfo,
		}
		if modelsReqBody[i].Task != nil {
			model.Task = &interfaces.MetricTask{
				TaskName:        modelsReqBody[i].Task.TaskName,
				Schedule:        modelsReqBody[i].Task.Schedule,
				TimeWindows:     modelsReqBody[i].Task.TimeWindows,
				Steps:           modelsReqBody[i].Task.Steps,
				IndexBase:       modelsReqBody[i].Task.IndexBase,
				RetraceDuration: modelsReqBody[i].Task.RetraceDuration,
				Comment:         modelsReqBody[i].Task.Comment,
			}
		}

		// 1. 校验 目标模型必要创建参数的合法性, 非空、长度、是枚举值
		err := ValidateObjectiveModel(ctx, &model)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateObjectiveModelAuditObject(model.ModelID, model.ModelName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Validate objective model[%s] failed: %s. %v", model.ModelName,
				httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			span.SetAttributes(attr.Key("model_name").String(model.ModelName))

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		// 2. 校验 请求体中目标模型名称重复性
		if _, ok := tmpNameMap[model.ModelName]; !ok {
			tmpNameMap[model.ModelName] = nil
		} else {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
				derrors.DataModel_ObjectiveModel_Duplicated_ObjectiveModelName)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateObjectiveModelAuditObject(model.ModelID, model.ModelName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Duplicated objective model name: [%s]: %s. %v",
				fmt.Sprintf("%v", model.ModelName),
				httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			span.SetAttributes(attr.Key("model_name").String(model.ModelName))

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		createModels = append(createModels, &model)

	}

	modelIDs, err := r.oms.CreateObjectiveModels(ctx, createModels, mode)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateObjectiveModelAuditObject(createModels[0].ModelID, createModels[0].ModelName), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//调用创建
	for _, model := range createModels {
		//每次成功创建 记录审计日志
		audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateObjectiveModelAuditObject(model.ModelID, model.ModelName), "")
	}

	// modelIDsStr := strings.Join(modelIDs, ",")
	result := []interface{}{}
	for _, modelID := range modelIDs {
		result = append(result, map[string]interface{}{"id": modelID})
	}

	logger.Debug("Handler CreateObjectiveModels Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)

	// c.Writer.Header().Set("Location", "/api/mdl-data-model/v1/objective-models/"+modelIDsStr)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// 分页获取模型列表(内部)
func (r *restHandler) ListObjectiveModelsByIn(c *gin.Context) {
	logger.Debug("Handler ListObjectiveModelsByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.ListObjectiveModels(c, visitor)
}

// 分页获取模型列表（外部）
func (r *restHandler) ListObjectiveModelsByEx(c *gin.Context) {
	logger.Debug("Handler ListObjectiveModelsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"分页获取目标模型列表", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ListObjectiveModels(c, visitor)
}

// 分页获取模型列表
func (r *restHandler) ListObjectiveModels(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("ListObjectiveModels Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"分页获取目标模型列表", trace.WithSpanKind(trace.SpanKindServer))
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
	o11y.Info(ctx, fmt.Sprintf("分页获取目标模型列表请求参数: [%s]", c.Request.RequestURI))

	// 获取分页参数
	namePattern := c.Query("name_pattern")
	name := c.Query("name")
	objectiveType := c.Query("objective_type")
	tag := c.Query("tag")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", "update_time")
	direction := c.DefaultQuery("direction", interfaces.DESC_DIRECTION)

	//去掉标签前后的所有空格进行搜索
	tag = strings.Trim(tag, " ")

	// 校验分页查询参数
	pageParam, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.OBJECTIVE_MODEL_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// name_pattern 和 name 不能同时存在
	if namePattern != "" && name != "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter).
			WithErrorDetails("name_pattern and name cannot exists at the same time")

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 构造标签列表查询参数的结构体
	parameter := interfaces.ObjectiveModelsQueryParams{
		NamePattern:   namePattern,
		Name:          name,
		Tag:           tag,
		ObjectiveType: objectiveType,
	}
	parameter.Sort = pageParam.Sort
	parameter.Direction = pageParam.Direction
	parameter.Limit = pageParam.Limit
	parameter.Offset = pageParam.Offset

	var result map[string]interface{}

	// 获取目标模型
	objectiveModelList, total, err := r.oms.ListObjectiveModels(ctx, parameter)
	result = map[string]interface{}{"entries": objectiveModelList, "total_count": total}
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler ListObjectiveModels Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// 按 id 获取目标模型对象信息(内部)
func (r *restHandler) GetObjectiveModelsByIn(c *gin.Context) {
	logger.Debug("Handler GetObjectiveModelsByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.GetObjectiveModels(c, visitor)
}

// 按 id 获取目标模型对象信息（外部）
func (r *restHandler) GetObjectiveModelsByEx(c *gin.Context) {
	logger.Debug("Handler GetObjectiveModelsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"按id获取目标模型信息", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetObjectiveModels(c, visitor)
}

// 按 id 获取目标模型对象信息
func (r *restHandler) GetObjectiveModels(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler GetObjectiveModels Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"按id获取目标模型信息", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	//获取参数字符串
	modelIDsStr := c.Param("model_ids")
	span.SetAttributes(attr.Key("model_ids").String(modelIDsStr))

	//解析字符串 转换为数组
	modelIDs := common.StringToStringSlice(modelIDsStr)

	// 获取目标模型的详细信息
	result, err := r.oms.GetObjectiveModels(ctx, modelIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	span.SetStatus(codes.Ok, "")

	logger.Debug("Handler GetObjectiveModels Success")
	rest.ReplyOK(c, http.StatusOK, result)
}

// 修改目标模型(内部)
func (r *restHandler) UpdateObjectiveModelByIn(c *gin.Context) {
	logger.Debug("Handler UpdateObjectiveModelByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.UpdateObjectiveModel(c, visitor)
}

// 修改目标模型（外部）
func (r *restHandler) UpdateObjectiveModelByEx(c *gin.Context) {
	logger.Debug("Handler UpdateObjectiveModelByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"修改目标模型", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.UpdateObjectiveModel(c, visitor)
}

// 修改目标模型
func (r *restHandler) UpdateObjectiveModel(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler UpdateObjectiveModel Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"修改目标模型", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. 接受 model_id 参数
	modelID := c.Param("model_id")
	span.SetAttributes(attr.Key("model_id").String(modelID))

	//接收绑定参数
	objectiveModel := interfaces.ObjectiveModel{}
	err := c.ShouldBindJSON(&objectiveModel)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateObjectiveModelAuditObject(modelID, ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	objectiveModel.ModelID = modelID

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("修改目标模型请求参数: [%s, %v]", c.Request.RequestURI, objectiveModel))

	// 校验 目标模型基本参数的合法性, 非空、长度、是枚举值
	err = ValidateObjectiveModel(ctx, &objectiveModel)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateObjectiveModelAuditObject(modelID, objectiveModel.ModelName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Validate objective model[%s] failed: %s. %v", objectiveModel.ModelName,
			httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		span.SetAttributes(attr.Key("model_name").String(objectiveModel.ModelName))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 先按id获取原对象
	oldObjectiveName, exist, err := r.oms.CheckObjectiveModelExistByID(ctx, modelID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateObjectiveModelAuditObject(modelID, objectiveModel.ModelName), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusForbidden,
			derrors.DataModel_ObjectiveModel_ObjectiveModelNotFound)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateObjectiveModelAuditObject(modelID, objectiveModel.ModelName), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 名称或分组不同，校验新名称是否已存在
	ifNameModify := false
	if oldObjectiveName != objectiveModel.ModelName {
		ifNameModify = true
		_, nameExist, err := r.oms.CheckObjectiveModelExistByName(ctx, objectiveModel.ModelName)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateObjectiveModelAuditObject(modelID, objectiveModel.ModelName), &httpErr.BaseError)

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		if nameExist {
			httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_ObjectiveModel_ModelNameExisted).
				WithErrorDetails(fmt.Sprintf("Objective Model %v already exist!", objectiveModel.ModelName))

			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateObjectiveModelAuditObject(modelID, objectiveModel.ModelName), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}
	objectiveModel.IfNameModify = ifNameModify

	//根据id修改信息
	err = r.oms.UpdateObjectiveModel(ctx, nil, objectiveModel)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateObjectiveModelAuditObject(modelID, objectiveModel.ModelName), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		GenerateObjectiveModelAuditObject(modelID, objectiveModel.ModelName), "")

	logger.Debug("Handler UpdateObjectiveModel Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 批量删除目标模型
func (r *restHandler) DeleteObjectiveModels(c *gin.Context) {
	logger.Debug("Handler DeleteObjectiveModels Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"删除目标模型", trace.WithSpanKind(trace.SpanKindServer))
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
	o11y.Info(ctx, fmt.Sprintf("删除目标模型请求参数: [%s]", c.Request.RequestURI))

	//获取参数字符串 <id1,id2,id3>
	modelIDsStr := c.Param("model_ids")
	span.SetAttributes(attr.Key("model_ids").String(modelIDsStr))
	//解析字符串 转换为数组
	modelIDs := common.StringToStringSlice(modelIDsStr)

	//检查 modelIDs 是否都存在
	var modelNames []string
	for _, modelID := range modelIDs {
		// 先按id获取原对象
		objectiveName, exist, err := r.oms.CheckObjectiveModelExistByID(ctx, modelID)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				GenerateObjectiveModelAuditObject(modelID, ""), &httpErr.BaseError)

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		if !exist {
			httpErr := rest.NewHTTPError(ctx, http.StatusForbidden,
				derrors.DataModel_ObjectiveModel_ObjectiveModelNotFound)

			audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				GenerateObjectiveModelAuditObject(modelID, ""), &httpErr.BaseError)

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		modelNames = append(modelNames, objectiveName)
	}

	// 批量删除目标模型
	rowsAffect, err := r.oms.DeleteObjectiveModels(ctx, modelIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 批量删除的审计日志,一个都每删除成功，填空
		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateObjectiveModelAuditObject(modelIDsStr, ""), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//循环记录审计日志
	if rowsAffect != 0 {
		for i := 0; i < len(modelNames); i++ {
			audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				GenerateObjectiveModelAuditObject(modelIDs[i], modelNames[i]), audit.SUCCESS, "")
		}
	}

	logger.Debug("Handler DeleteObjectiveModels Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 分页获取目标模型资源列表
func (r *restHandler) ListObjectiveModelSrcs(c *gin.Context) {
	logger.Debug("ListObjectiveModelSrcs Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"分页获取目标模型资源示例列表", trace.WithSpanKind(trace.SpanKindServer))
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
	o11y.Info(ctx, fmt.Sprintf("分页获取目标模型实例列表请求参数: [%s]", c.Request.RequestURI))

	// 获取分页参数
	namePattern := c.Query(RESOURCES_KEYWOED)
	name := c.Query("name")
	objectiveType := c.Query("objective_type")
	tag := c.Query("tag")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", RESOURCES_PAGE_LIMIT)
	sort := c.DefaultQuery("sort", "model_name")
	direction := c.DefaultQuery("direction", interfaces.ASC_DIRECTION)

	//去掉标签前后的所有空格进行搜索
	tag = strings.Trim(tag, " ")

	// 校验分页查询参数
	pageParam, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.OBJECTIVE_MODEL_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// name_pattern 和 name 不能同时存在
	if namePattern != "" && name != "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_ObjectiveModel_InvalidParameter).
			WithErrorDetails("name_pattern and name cannot exists at the same time")

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 构造标签列表查询参数的结构体
	parameter := interfaces.ObjectiveModelsQueryParams{
		NamePattern:   namePattern,
		Name:          name,
		Tag:           tag,
		ObjectiveType: objectiveType,
	}
	parameter.Sort = pageParam.Sort
	parameter.Direction = pageParam.Direction
	parameter.Limit = pageParam.Limit
	parameter.Offset = pageParam.Offset

	var result map[string]interface{}

	// 获取目标模型
	objectiveModelList, total, err := r.oms.ListObjectiveModelSrcs(ctx, parameter)
	result = map[string]interface{}{"entries": objectiveModelList, "total_count": total}
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler ListObjectiveModelSrcs Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}
