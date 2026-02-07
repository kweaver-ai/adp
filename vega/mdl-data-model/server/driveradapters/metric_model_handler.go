// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"data-model/common"
	derrors "data-model/errors"
	"data-model/interfaces"
)

// 创建指标模型(内部)
func (r *restHandler) CreateMetricModelsByIn(c *gin.Context) {
	logger.Debug("Handler CreateMetricModelsByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.CreateMetricModels(c, visitor)
}

// 创建指标模型（外部）
func (r *restHandler) CreateMetricModelsByEx(c *gin.Context) {
	logger.Debug("Handler CreateMetricModelsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"创建指标模型", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.CreateMetricModels(c, visitor)
}

// 创建指标模型
func (r *restHandler) CreateMetricModels(c *gin.Context, visitor rest.Visitor) {
	//创建时新增参数groupName ，需要进行校验
	//请求体以及数据库表中重复名称校验需要groupName 和modelName
	//需要通过groupName更新groupID才能创建
	logger.Debug("Handler CreateMetricModels Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"创建指标模型", trace.WithSpanKind(trace.SpanKindServer))
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
			GenerateMetricModelAuditObject("", ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 接受绑定参数
	modelsReqBody := []interfaces.CreateMetricModel{}
	err := c.ShouldBindJSON(&modelsReqBody)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateMetricModelAuditObject("", ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	// 如果传入的模型对象为[], 应报错
	if len(modelsReqBody) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_RequestBody).
			WithErrorDetails("No metric model was passed in")

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateMetricModelAuditObject("", ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("创建指标模型请求参数: [%s,%v]", c.Request.RequestURI, modelsReqBody))

	models := make([]*interfaces.MetricModel, 0)
	// 校验 请求体中指标模型名称合法性
	tmpNameMap := make(map[interfaces.CombinationName]interface{})
	idMap := make(map[string]any)
	// 校验 请求体中指标模型度量名称合法性
	tmpMeasureNameMap := make(map[string]interface{})
	for i := 0; i < len(modelsReqBody); i++ {
		// 校验导入模型时模块是否是指标模型
		if modelsReqBody[i].ModuleType != "" && modelsReqBody[i].ModuleType != interfaces.MODULE_TYPE_METRIC_MODEL {
			httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, derrors.DataModel_InvalidParameter_ModuleType).
				WithErrorDetails("Model name is not 'metric_model'")

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateMetricModelAuditObject("", ""), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		// 0.校验请求体中多个模型 ID 是否重复
		modelID := modelsReqBody[i].ModelID
		if _, ok := idMap[modelID]; !ok || modelID == "" {
			idMap[modelID] = nil
		} else {
			errDetails := fmt.Sprintf("MetricModel ID '%s' already exists in the request body", modelID)
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_Duplicated_ModelIDInFile).
				WithDescription(map[string]any{"ModelID": modelID}).
				WithErrorDetails(errDetails)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateMetricModelAuditObject("", modelsReqBody[i].ModelName), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		// 处理对data_view_id的兼容
		if modelsReqBody[i].DataSource == nil {
			// 数据源为空，尝试读取data_view_id
			if modelsReqBody[i].DataViewID != "" {
				// 视图id不为空，则把data_view_id转成数据源
				modelsReqBody[i].DataSource = &interfaces.CreateMetricDataSource{
					// Type: interfaces.DATA_SOURCE_DATA_VIEW,
					ID: modelsReqBody[i].DataViewID,
				}
			}
		}

		metricModel := &interfaces.MetricModel{
			SimpleMetricModel: interfaces.SimpleMetricModel{
				ModelID:         modelsReqBody[i].ModelID,
				ModelName:       modelsReqBody[i].ModelName,
				CatalogID:       modelsReqBody[i].CatalogID,
				CatalogContent:  modelsReqBody[i].CatalogContent,
				MeasureName:     modelsReqBody[i].MeasureName,
				GroupName:       modelsReqBody[i].GroupName,
				Tags:            modelsReqBody[i].Tags,
				Comment:         modelsReqBody[i].Comment,
				MetricType:      modelsReqBody[i].MetricType,
				QueryType:       modelsReqBody[i].QueryType,
				Formula:         modelsReqBody[i].Formula,
				FormulaConfig:   modelsReqBody[i].FormulaConfig,
				OrderByFields:   modelsReqBody[i].OrderByFields,
				HavingCondition: modelsReqBody[i].HavingCondition,
				AnalysisDims:    modelsReqBody[i].AnalysisDims,
				DateField:       modelsReqBody[i].DateField,
				MeasureField:    modelsReqBody[i].MeasureField,
				UnitType:        modelsReqBody[i].UnitType,
				Unit:            modelsReqBody[i].Unit,
				Builtin:         modelsReqBody[i].Builtin,
			},
		}
		if modelsReqBody[i].DataSource != nil {
			metricModel.DataSource = &interfaces.MetricDataSource{
				Type: modelsReqBody[i].DataSource.Type,
				ID:   modelsReqBody[i].DataSource.ID,
			}
		}
		if modelsReqBody[i].Task != nil && len(modelsReqBody[i].Task.Steps) > 0 {
			metricModel.Task = &interfaces.MetricTask{
				TaskName:        modelsReqBody[i].Task.TaskName,
				Schedule:        modelsReqBody[i].Task.Schedule,
				TimeWindows:     modelsReqBody[i].Task.TimeWindows,
				Steps:           modelsReqBody[i].Task.Steps,
				IndexBase:       modelsReqBody[i].Task.IndexBase,
				RetraceDuration: modelsReqBody[i].Task.RetraceDuration,
				Comment:         modelsReqBody[i].Task.Comment,
			}
		}

		// 1. 校验 指标模型必要创建参数的合法性, 非空、长度、是枚举值
		containTopHits, err := ValidateMetricModel(ctx, metricModel)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateMetricModelAuditObject("", metricModel.ModelName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Validate metric model[%s] failed: %s. %v", metricModel.ModelName,
				httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			span.SetAttributes(
				attr.Key("group_name").String(metricModel.GroupName),
				attr.Key("model_name").String(metricModel.ModelName))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		metricModel.IfContainTopHits = containTopHits

		//2. 校验groupName
		if metricModel.GroupName != "" {
			err = validateObjectName(ctx, metricModel.GroupName, interfaces.METRIC_MODEL_GROUP_MODULE)
			if err != nil {
				httpErr := err.(*rest.HTTPError)

				audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
					GenerateMetricModelAuditObject("", metricModel.ModelName), &httpErr.BaseError)

				// 记录异常日志
				o11y.Error(ctx, fmt.Sprintf("Validate metric model group [%s] failed: %s. %v", metricModel.GroupName,
					httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

				// 设置 trace 的错误信息的 attributes
				span.SetAttributes(
					attr.Key("group_name").String(metricModel.GroupName),
					attr.Key("model_name").String(metricModel.ModelName))

				o11y.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}

		cname := interfaces.CombinationName{
			GroupName: metricModel.GroupName,
			ModelName: metricModel.ModelName,
		}

		// 3. 校验 请求体中指标模型名称重复性
		if _, ok := tmpNameMap[cname]; !ok {
			tmpNameMap[cname] = nil
		} else {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_Duplicated_CombinationName)

			audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
				GenerateMetricModelAuditObject("", metricModel.ModelName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Duplicated metric model combination name: [%s]: %s. %v", fmt.Sprintf("%v", cname),
				httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			span.SetAttributes(
				attr.Key("group_name").String(metricModel.GroupName),
				attr.Key("model_name").String(metricModel.ModelName))

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		// 3. 校验 请求体中指标模型度量名称重复性
		if metricModel.MeasureName != "" {
			if _, ok := tmpMeasureNameMap[metricModel.MeasureName]; !ok {
				tmpMeasureNameMap[metricModel.MeasureName] = nil
			} else {
				httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_Duplicated_MeasureName)

				audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
					GenerateMetricModelAuditObject("", metricModel.ModelName), &httpErr.BaseError)

				// 记录异常日志
				o11y.Error(ctx, fmt.Sprintf("Duplicated metric model measure name: [%s]: %s. %v", metricModel.MeasureName,
					httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

				// 设置 trace 的错误信息的 attributes
				span.SetAttributes(attr.Key("measure_name").String(metricModel.MeasureName))

				o11y.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}
		models = append(models, metricModel)
	}

	//调用创建
	modelIDs, err := r.mms.CreateMetricModels(ctx, models, mode)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateMetricModelAuditObject(models[0].ModelID, models[0].ModelName), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 成功，发送多条
	for _, model := range models {
		//每次成功创建 记录审计日志
		audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			GenerateMetricModelAuditObject(model.ModelID, model.ModelName), "")
	}

	// modelIDsStr := strings.Join(modelIDs, ",")
	result := []interface{}{}
	for _, modelID := range modelIDs {
		result = append(result, map[string]interface{}{"id": modelID})
	}

	logger.Debug("Handler CreateMetricModels Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	// c.Writer.Header().Set("Location", "/api/mdl-data-model/v1/metric-models/"+modelIDsStr)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// 更新指标模型(内部)
func (r *restHandler) UpdateMetricModelByIn(c *gin.Context) {
	logger.Debug("Handler UpdateMetricModelByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.UpdateMetricModel(c, visitor)
}

// 更新指标模型（外部）
func (r *restHandler) UpdateMetricModelByEx(c *gin.Context) {
	logger.Debug("Handler CreateMetricModelsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"修改指标模型", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.UpdateMetricModel(c, visitor)
}

// 更新指标模型
func (r *restHandler) UpdateMetricModel(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler UpdateMetricModel Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"修改指标模型", trace.WithSpanKind(trace.SpanKindServer))
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
	metricModel := interfaces.MetricModel{}
	err := c.ShouldBindJSON(&metricModel)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateMetricModelAuditObject(modelID, ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	metricModel.ModelID = modelID

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("修改指标模型请求参数: [%s, %v]", c.Request.RequestURI, metricModel))

	// 先按id获取原对象
	oldMetricModel, err := r.mms.GetMetricModelByModelID(ctx, modelID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateMetricModelAuditObject(modelID, ""), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 处理对data_view_id的兼容
	if metricModel.DataSource == nil {
		// 数据源为空，尝试读取data_view_id
		if metricModel.DataViewID != "" {
			// 视图id不为空，则把data_view_id转成数据源
			metricModel.DataSource = &interfaces.MetricDataSource{
				// Type: interfaces.DATA_SOURCE_DATA_VIEW,
				ID: metricModel.DataViewID,
			}
		}
	}

	// 校验 指标模型基本参数的合法性, 非空、长度、是枚举值
	containTopHits, err := ValidateMetricModel(ctx, &metricModel)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateMetricModelAuditObject(modelID, metricModel.ModelName), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Validate metric model[%s] failed: %s. %v", metricModel.ModelName,
			httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		span.SetAttributes(attr.Key("model_name").String(metricModel.ModelName))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	metricModel.IfContainTopHits = containTopHits

	//校验 GroupName 是否合法
	if metricModel.GroupName != "" {
		err = validateObjectName(ctx, metricModel.GroupName, interfaces.METRIC_MODEL_GROUP_MODULE)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateMetricModelAuditObject(modelID, metricModel.ModelName), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Validate metric model[%s] failed: %s. %v", metricModel.ModelName,
				httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			span.SetAttributes(attr.Key("model_name").String(metricModel.ModelName))

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	// 名称或分组不同，校验新名称是否已存在
	ifNameModify := false
	if oldMetricModel.ModelName != metricModel.ModelName || oldMetricModel.GroupName != metricModel.GroupName {
		ifNameModify = true
		_, exist, err := r.mms.CheckMetricModelExistByName(ctx, metricModel.GroupName, metricModel.ModelName)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateMetricModelAuditObject(modelID, metricModel.ModelName), &httpErr.BaseError)

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if exist {
			httpErr := rest.NewHTTPError(ctx, http.StatusForbidden,
				derrors.DataModel_MetricModel_CombinationNameExisted)

			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateMetricModelAuditObject(modelID, metricModel.ModelName), &httpErr.BaseError)

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}
	metricModel.IfNameModify = ifNameModify

	// 更新操作时，measureName是旧的，不能修改
	metricModel.MeasureName = oldMetricModel.MeasureName
	//根据id修改信息
	err = r.mms.UpdateMetricModel(ctx, nil, metricModel)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateMetricModelAuditObject(modelID, metricModel.ModelName), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		GenerateMetricModelAuditObject(modelID, metricModel.ModelName), "")

	logger.Debug("Handler UpdateMetricModel Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 批量删除指标模型
func (r *restHandler) DeleteMetricModels(c *gin.Context) {
	logger.Debug("Handler DeleteMetricModels Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"删除指标模型", trace.WithSpanKind(trace.SpanKindServer))
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
	o11y.Info(ctx, fmt.Sprintf("删除指标模型请求参数: [%s]", c.Request.RequestURI))

	//获取参数字符串 <id1,id2,id3>
	modelIDsStr := c.Param("model_ids")
	span.SetAttributes(attr.Key("model_ids").String(modelIDsStr))

	//解析字符串 转换为 []string
	modelIDs := common.StringToStringSlice(modelIDsStr)

	//检查 modelIDs 是否都存在
	var metricModels []interfaces.MetricModel
	for _, modelID := range modelIDs {
		metricModel, err := r.mms.GetMetricModelByModelID(ctx, modelID)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				GenerateMetricModelAuditObject(modelID, ""), &httpErr.BaseError)

			// 设置 trace 的错误信息的 attributes
			o11y.AddHttpAttrs4HttpError(span, httpErr)

			rest.ReplyError(c, httpErr)
			return
		}

		metricModels = append(metricModels, metricModel)
	}

	// 批量删除指标模型
	rowsAffect, err := r.mms.DeleteMetricModels(ctx, nil, modelIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 批量删除的审计日志,一个都每删除成功，填空
		audit.NewWarnLogWithError(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			GenerateMetricModelAuditObject("", ""), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//循环记录审计日志
	if rowsAffect != 0 {
		for _, model := range metricModels {
			audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
				GenerateMetricModelAuditObject(model.ModelID, model.ModelName), audit.SUCCESS, "")
		}
	}

	logger.Debug("Handler DeleteMetricModels Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 分页获取指标模型列表(内部)
func (r *restHandler) ListMetricModelsByIn(c *gin.Context) {
	logger.Debug("Handler ListMetricModelsByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.ListMetricModels(c, visitor)
}

// 分页获取指标模型列表（外部）
func (r *restHandler) ListMetricModelsByEx(c *gin.Context) {
	logger.Debug("Handler CreateMetricModelsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"分页获取指标模型列表", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ListMetricModels(c, visitor)
}

// 分页获取指标模型列表
func (r *restHandler) ListMetricModels(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("ListMetricModels Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"分页获取指标模型列表", trace.WithSpanKind(trace.SpanKindServer))
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
	o11y.Info(ctx, fmt.Sprintf("分页获取指标模型列表请求参数: [%s]", c.Request.RequestURI))

	// 获取分页参数
	namePattern := c.Query("name_pattern")
	name := c.Query("name")
	metricType := c.Query("metric_type")
	queryType := c.Query("query_type")
	tag := c.Query("tag")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", "update_time")
	direction := c.DefaultQuery("direction", interfaces.DESC_DIRECTION)
	groupIDstr := c.DefaultQuery("group_id", interfaces.GroupID_All)

	//去掉标签前后的所有空格进行搜索
	tag = strings.Trim(tag, " ")

	// 校验分页查询参数
	pageParam, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.METRIC_MODEL_SORT)
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
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter).
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
	parameter := interfaces.MetricModelsQueryParams{
		NamePattern: namePattern,
		Name:        name,
		Tag:         tag,
		MetricType:  metricType,
		QueryType:   queryType,
		GroupID:     groupIDstr,
	}
	parameter.Sort = pageParam.Sort
	parameter.Direction = pageParam.Direction
	parameter.Limit = pageParam.Limit
	parameter.Offset = pageParam.Offset

	// var result map[string]interface{}
	// if simpleInfo {
	// 获取指标模型简单信息
	simpleMetricModelList, total, err := r.mms.ListSimpleMetricModels(ctx, parameter)
	result := map[string]interface{}{"entries": simpleMetricModelList, "total_count": total}
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

	logger.Debug("Handler ListMetricModels Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// 按 id 获取指标模型对象信息(内部)
func (r *restHandler) GetMetricModelsByIn(c *gin.Context) {
	logger.Debug("Handler GetMetricModelsByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.GetMetricModels(c, visitor)
}

// 按 id 获取指标模型对象信息（外部）
func (r *restHandler) GetMetricModelsByEx(c *gin.Context) {
	logger.Debug("Handler CreateMetricModelsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"分页获取指标模型列表", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetMetricModels(c, visitor)
}

// 按 id 获取指标模型对象信息
func (r *restHandler) GetMetricModels(c *gin.Context, visitor rest.Visitor) {
	logger.Debug("Handler GetMetricModels Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Get metric models", trace.WithSpanKind(trace.SpanKindServer))
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

	//解析字符串 转换为 []string
	modelIDs := common.StringToStringSlice(modelIDsStr)

	// 详情是否包含视图详细信息
	includeViewParam := c.DefaultQuery("include_view", interfaces.DEFAULT_INCLUDE_VIEW)
	includeView, err := strconv.ParseBool(includeViewParam)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_InvalidParameter_SimpleInfo).
			WithErrorDetails(fmt.Sprintf("The include_view:%s is invalid", includeViewParam))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 获取指标模型的详细信息，根据 include_view 参数来判断是否包含数据视图的过滤条件
	result, err := r.mms.GetMetricModels(ctx, modelIDs, includeView)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	logger.Debug("Handler GetMetricModel Success")
	rest.ReplyOK(c, http.StatusOK, result)
}

// 批量修改指标模型的分组
func (r *restHandler) UpdateMetricModels(c *gin.Context) {
	logger.Debug("Handler UpdateMetricModels Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"批量修改指标模型的分组", trace.WithSpanKind(trace.SpanKindServer))
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

	// 接受绑定参数
	bodyGroup := interfaces.MetricModelGroupName{}
	err = c.ShouldBindJSON(&bodyGroup)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateMetricModelAuditObject("", ""), &httpErr.BaseError)

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("批量修改指标模型的分组请求参数: [%s,%v]", c.Request.RequestURI, bodyGroup.GroupName))

	//获取参数字符串 <id1,id2,id3>
	modelIDsStr := c.Param("model_id")
	span.SetAttributes(attr.Key("model_ids").String(modelIDsStr))

	//解析字符串 转换为 []string
	modelIDs := common.StringToStringSlice(modelIDsStr)

	//校验 GroupName 是否合法
	if bodyGroup.GroupName != "" {
		err := validateObjectName(ctx, bodyGroup.GroupName, interfaces.METRIC_MODEL_GROUP_MODULE)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateMetricModelAuditObject("", ""), &httpErr.BaseError)

			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Validate metric model[%s] failed: %s. %v", bodyGroup.GroupName,
				httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

			// 设置 trace 的错误信息的 attributes
			span.SetAttributes(attr.Key("model_name").String(bodyGroup.GroupName))

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	//检查 modelIDs 是否都存在
	modelMap, err := r.mms.GetMetricModelSimpleInfosByIDs(ctx, modelIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateMetricModelAuditObject("", ""), &httpErr.BaseError)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if len(modelMap) != len(modelIDs) {
		errStr := fmt.Sprintf("Exists any models not found, expect model nums is [%d], actual models num is [%d]", len(modelIDs), len(modelMap))
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound,
			derrors.DataModel_MetricModel_MetricModelNotFound).WithErrorDetails(errStr)

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 判断批量移动的视图的名称是否存在重复，可以选择不同分组的模型移动到指定分组
	modelNameSet := make(map[string]struct{})
	for _, model := range modelMap {
		if _, ok := modelNameSet[model.ModelName]; ok {
			errDetails := fmt.Sprintf("metric model name '%s' is duplicated in group '%s'", model.ModelName, bodyGroup.GroupName)
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_Duplicated_ModelName).
				WithDescription(map[string]any{"MetricModelName": model.ModelName, "GroupName": bodyGroup.GroupName}).
				WithErrorDetails(errDetails)

			audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateMetricModelAuditObject("", model.ModelName), &httpErr.BaseError)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		modelNameSet[model.ModelName] = struct{}{}
	}

	// 批量修改指标模型的分组
	rowsAffect, err := r.mms.UpdateMetricModelsGroup(ctx, modelMap, bodyGroup)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 批量修改的审计日志,一个都没修改成功，填空
		audit.NewWarnLogWithError(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
			GenerateMetricModelAuditObject("", ""), &httpErr.BaseError)

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//循环记录审计日志
	if rowsAffect != 0 {
		for _, model := range modelMap {
			audit.NewWarnLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
				GenerateMetricModelAuditObject(model.ModelID, model.ModelName), audit.SUCCESS, "")
		}
	}

	logger.Debug("Handler UpdateMetricModels Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 按任务 id 获取指标模型持久化任务对象信息(内部)
func (r *restHandler) GetMetricTaskByIn(c *gin.Context) {
	logger.Debug("Handler GetMetricTaskByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.GetMetricTask(c, visitor)
}

// 按任务 id 获取指标模型持久化任务对象信息（外部）
func (r *restHandler) GetMetricTaskByEx(c *gin.Context) {
	logger.Debug("Handler CreateMetricModelsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"按任务id获取指标模型持久化任务信息", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetMetricTask(c, visitor)
}

// 按任务 id 获取指标模型持久化任务对象信息
func (r *restHandler) GetMetricTask(c *gin.Context, visitor rest.Visitor) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"按任务id获取指标模型持久化任务信息", trace.WithSpanKind(trace.SpanKindServer))
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
	taskID := c.Param("task_id")
	span.SetAttributes(attr.Key("task_id").String(taskID))

	// 获取指标模型的详细信息，
	tasks, err := r.mmts.GetMetricTasksByTaskIDs(ctx, []string{taskID})
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
			derrors.DataModel_MetricModel_InternalError_GetMetricTaskByIDFailed).
			WithErrorDetails("Get metric task Failed:" + err.Error())

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if len(tasks) <= 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, derrors.DataModel_MetricModel_MetricTaskNotFound).
			WithErrorDetails(fmt.Sprintf("Metric task[%s] not found", taskID))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, tasks[0])
}

// 更新任务的计划时间和执行状态(内部)
func (r *restHandler) UpdateMetricTaskPlanTimeByIn(c *gin.Context) {
	logger.Debug("Handler UpdateMetricTaskPlanTimeByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := GenerateVisitor(c)
	r.UpdateMetricTaskPlanTime(c, visitor)
}

// 更新任务的计划时间和执行状态（外部）
func (r *restHandler) UpdateMetricTaskPlanTimeByEx(c *gin.Context) {
	logger.Debug("Handler CreateMetricModelsByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"更新任务的计划时间", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.UpdateMetricTaskPlanTime(c, visitor)
}

// 更新任务的计划时间和执行状态
func (r *restHandler) UpdateMetricTaskPlanTime(c *gin.Context, visitor rest.Visitor) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"更新任务的计划时间", trace.WithSpanKind(trace.SpanKindServer))
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
	taskID := c.Param("task_id")
	span.SetAttributes(attr.Key("task_id").String(taskID))

	//接收绑定参数
	task := interfaces.MetricTask{}
	err := c.ShouldBindJSON(&task)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	task.TaskID = taskID
	//根据id修改信息
	err = r.mmts.UpdateMetricTaskAttributes(ctx, task)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 按 id 获取指标模型所绑定的数据源的字段列表
func (r *restHandler) GetMetricModelSourceFields(c *gin.Context) {
	logger.Debug("Handler GetMetricModelSourceFields Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"按 id 获取指标模型所绑定的数据源的字段列表", trace.WithSpanKind(trace.SpanKindServer))
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

	// 1. 接受 model_ids 参数。参数名虽然为 model_ids，但是实际上只支持单个模型的获取
	modelID := c.Param("model_ids")
	span.SetAttributes(attr.Key("model_id").String(modelID))

	// 获取指标模型的详细信息，根据 include_view 参数来判断是否包含数据视图的过滤条件
	result, err := r.mms.GetMetricModelSourceFields(ctx, modelID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	logger.Debug("Handler GetMetricModelSourceFields Success")
	rest.ReplyOK(c, http.StatusOK, result)
}

// 分页获取指标模型资源列表
func (r *restHandler) ListMetricModelSrcs(c *gin.Context) {
	logger.Debug("ListMetricModelSrcs Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"分页获取指标模型资源示例列表", trace.WithSpanKind(trace.SpanKindServer))
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
	o11y.Info(ctx, fmt.Sprintf("分页获取指标模型资源实例列表请求参数: [%s]", c.Request.RequestURI))

	// 获取分页参数
	namePattern := c.Query(RESOURCES_KEYWOED) // 统一资源平台获取资源列表搜索时，用 key 来接
	name := c.Query("name")
	metricType := c.Query("metric_type")
	queryType := c.Query("query_type")
	tag := c.Query("tag")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", RESOURCES_PAGE_LIMIT)
	sort := c.DefaultQuery("sort", "group_name")
	direction := c.DefaultQuery("direction", interfaces.ASC_DIRECTION)
	groupIDstr := c.DefaultQuery("group_id", interfaces.GroupID_All)

	//去掉标签前后的所有空格进行搜索
	tag = strings.Trim(tag, " ")

	// 校验分页查询参数
	pageParam, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.METRIC_MODEL_SORT)
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
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, derrors.DataModel_MetricModel_InvalidParameter).
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
	parameter := interfaces.MetricModelsQueryParams{
		NamePattern: namePattern,
		Name:        name,
		Tag:         tag,
		MetricType:  metricType,
		QueryType:   queryType,
		GroupID:     groupIDstr,
	}
	parameter.Sort = pageParam.Sort
	parameter.Direction = pageParam.Direction
	parameter.Limit = pageParam.Limit
	parameter.Offset = pageParam.Offset

	// var result map[string]interface{}
	// if simpleInfo {
	// 获取指标模型简单信息
	resources, total, err := r.mms.ListMetricModelSrcs(ctx, parameter)
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
	result := map[string]interface{}{"entries": resources, "total_count": total}

	logger.Debug("Handler ListMetricModels Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// 按 id 获取指标模型所绑定的数据源的字段列表
func (r *restHandler) GetMetricModelOrderFields(c *gin.Context) {
	logger.Debug("Handler GetMetricModelOrderFields Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"按 id 获取指标模型可选的排序字段列表", trace.WithSpanKind(trace.SpanKindServer))
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

	// 1. 接受 model_ids 参数。参数名虽然为 model_ids，但是实际上只支持单个模型的获取
	modelIDsStr := c.Param("model_ids")
	//解析字符串 转换为 []string
	modelIDs := common.StringToStringSlice(modelIDsStr)

	span.SetAttributes(attr.Key("model_id").StringSlice(modelIDs))

	// 获取指标模型的详细信息，根据 include_view 参数来判断是否包含数据视图的过滤条件
	result, err := r.mms.GetMetricModelOrderByFields(ctx, modelIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	logger.Debug("Handler GetMetricModelSourceFields Success")
	rest.ReplyOK(c, http.StatusOK, result)
}
