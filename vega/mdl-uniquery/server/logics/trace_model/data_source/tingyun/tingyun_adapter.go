// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_source

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	"uniquery/common"
	cond "uniquery/common/condition"
	vopt "uniquery/common/value_opt"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics"
)

const (
	TINGYUN_TIME_FORMAT string = "2006-01-02 15:04"
	DEFAULT_TIME_PERIOD int64  = 10080 // 7天
)

type TingYunClientConfig struct {
	Address     string
	Protocol    string
	AccessToken string
	QueryParams string
}

type TingYunDetailedConfig struct {
	Address        string    `json:"address"`
	Protocol       string    `json:"protocol"`
	ApiKey         string    `json:"api_key"`
	SecretKey      string    `json:"secret_key,omitempty"`
	AccessToken    string    `json:"access_token,omitempty"`
	ExpirationTime time.Time `json:"expiration_time,omitempty"`
}

var (
	tyaOnce sync.Once
	tya     interfaces.TraceModelAdapter
)

type tingYunwAdapter struct {
	appSetting *common.AppSetting
	dcAccess   interfaces.DataConnectionAccess
	httpClient rest.HTTPClient
}

func NewTingYunAdapter(appSetting *common.AppSetting) interfaces.TraceModelAdapter {
	tyaOnce.Do(func() {
		tya = &tingYunwAdapter{
			appSetting: appSetting,
			dcAccess:   logics.DCAccess,
			httpClient: common.NewHTTPClient(),
		}
	})
	return tya
}

func (tyAdapter *tingYunwAdapter) GetSpanList(ctx context.Context, model interfaces.TraceModel, params interfaces.SpanListQueryParams) (spanList []interfaces.SpanListEntry, total int64, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 通过tingYunwAdapter获取span列表")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 生成queryParams, 供听云查询使用
	queryParams, err := tyAdapter.convertQueryParams(ctx, params)
	if err != nil {
		return []interfaces.SpanListEntry{}, 0, err
	}

	// 2. 查询数据连接详情
	spanConf, _ := model.SpanConfig.(interfaces.SpanConfigWithDataConnection)
	conn, isExist, err := tyAdapter.dcAccess.GetDataConnectionByID(ctx, spanConf.DataConnection.ID)
	if err != nil {
		logger.Errorf("Get data connection by id failed, err: %s", err.Error())
		return []interfaces.SpanListEntry{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetDataConnectionByIDFailed).
			WithErrorDetails(err.Error())
	}

	if !isExist {
		errDetails := fmt.Sprintf("Data connection whose id equal to %s was not found", spanConf.DataConnection.ID)
		logger.Error(errDetails)
		return []interfaces.SpanListEntry{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetDataConnectionByIDFailed).
			WithErrorDetails(errDetails)
	}

	err = tyAdapter.processDataConnection(ctx, conn)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return []interfaces.SpanListEntry{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_ProcessDataConnectionFailed).
			WithErrorDetails(err.Error())
	}

	// 3. 调用听云API获取trace列表
	tingYunDetailedConfig, _ := conn.DataSourceConfig.(TingYunDetailedConfig)
	clientConf := TingYunClientConfig{
		Address:     tingYunDetailedConfig.Address,
		Protocol:    tingYunDetailedConfig.Protocol,
		AccessToken: tingYunDetailedConfig.AccessToken,
		QueryParams: queryParams,
	}
	entries, total, err := tyAdapter.getTraceList(ctx, clientConf)
	if err != nil {
		logger.Errorf("Get tingyun trace list failed, err: %v", err.Error())
		return []interfaces.SpanListEntry{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetTingYunTraceListFailed).
			WithErrorDetails(err.Error())
	}

	// 4. 只取offset到offset+limit这中间的内容
	if len(entries) <= params.Offset {
		return []interfaces.SpanListEntry{}, 0, nil
	}

	// 5. 解析听云API返回的内容
	entries = entries[params.Offset:]
	spanList = make([]interfaces.SpanListEntry, len(entries))

	g, _ := errgroup.WithContext(ctx)
	g.SetLimit(40)
	for i, rawTraceEntry := range entries {
		i, rawTraceEntry := i, rawTraceEntry
		g.Go(
			func() error {
				abstractSpan := tyAdapter.extractRawTraceEntry(ctx, rawTraceEntry)
				spanList[i] = tyAdapter.genSpanDetail(rawTraceEntry, abstractSpan)
				return nil
			},
		)
	}

	_ = g.Wait()
	return spanList, total, nil
}

func (tyAdapter *tingYunwAdapter) GetSpan(ctx context.Context, model interfaces.TraceModel, params interfaces.SpanQueryParams) (spanDetail interfaces.SpanDetail, err error) {
	_, span := ar_trace.Tracer.Start(ctx, "logic层: 通过tingYunwAdapter获取span详情")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	return interfaces.SpanDetail{}, nil
}

func (tyAdapter *tingYunwAdapter) GetSpanMap(ctx context.Context, model interfaces.TraceModel,
	params interfaces.TraceQueryParams) (briefSpanMap map[string]*interfaces.BriefSpan_,
	detailSpanMap map[string]interfaces.SpanDetail, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 通过tingYunwAdapter获取span map")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 查询数据连接详情
	spanConf, _ := model.SpanConfig.(interfaces.SpanConfigWithDataConnection)
	conn, isExist, err := tyAdapter.dcAccess.GetDataConnectionByID(ctx, spanConf.DataConnection.ID)
	if err != nil {
		logger.Errorf("Get data connection by id failed, err: %v", err.Error())
		return nil, nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetDataConnectionByIDFailed).
			WithErrorDetails(err.Error())
	}

	if !isExist {
		errDetails := fmt.Sprintf("Data connection whose id equal to %s was not found", spanConf.DataConnection.ID)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return nil, nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetDataConnectionByIDFailed).
			WithErrorDetails(errDetails)
	}

	err = tyAdapter.processDataConnection(ctx, conn)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return nil, nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_ProcessDataConnectionFailed).
			WithErrorDetails(err.Error())
	}

	// 2. 调用听云API获取trace详情
	tingYunDetailedConfig, _ := conn.DataSourceConfig.(TingYunDetailedConfig)
	clientConf := TingYunClientConfig{
		Address:     tingYunDetailedConfig.Address,
		Protocol:    tingYunDetailedConfig.Protocol,
		AccessToken: tingYunDetailedConfig.AccessToken,
		QueryParams: tyAdapter.genTraceDetailQueryParams(params),
	}

	tyTraceDetail, isExist, err := tyAdapter.getTraceDetail(ctx, clientConf)
	if err != nil {
		logger.Errorf("Get tingyun trace detail failed, err: %v", err.Error())
		return nil, nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_TraceModel_InternalError_GetTingYunTraceDetailFailed).
			WithErrorDetails(err.Error())
	}

	if !isExist {
		errDetails := fmt.Sprintf("The trace whose id equal to %v was not found in tingyun system", params.TraceID)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return nil, nil, rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_TraceModel_TraceNotFound).
			WithErrorDetails(errDetails)
	}

	// 3. 解析听云API返回的内容
	briefSpanMap = make(map[string]*interfaces.BriefSpan_, 0)
	detailSpanMap = make(map[string]interfaces.SpanDetail, 0)
	err = tyAdapter.parseRawTraceDetail(ctx, "-2", params.TraceID, tyTraceDetail, briefSpanMap, detailSpanMap)
	if err != nil {
		o11y.Error(ctx, err.Error())
		return nil, nil, err
	}

	return briefSpanMap, detailSpanMap, nil
}

// 暂不支持通过数据连接查询听云log相关内容
func (tyAdapter *tingYunwAdapter) GetRelatedLogCountMap(ctx context.Context, model interfaces.TraceModel, params interfaces.TraceQueryParams) (countMap map[string]int64, err error) {
	_, span := ar_trace.Tracer.Start(ctx, "logic层: 通过tingYunwAdapter获取关联日志的统计信息")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	return map[string]int64{}, nil
}

// 暂不支持通过数据连接查询听云log相关内容
func (tyAdapter *tingYunwAdapter) GetSpanRelatedLogList(ctx context.Context, model interfaces.TraceModel, params interfaces.RelatedLogListQueryParams) (entries []interfaces.RelatedLogListEntry, total int64, err error) {
	_, span := ar_trace.Tracer.Start(ctx, "logic层: 通过tingYunwAdapter获取关联日志列表")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	return []interfaces.RelatedLogListEntry{}, 0, nil
}

/*
	私有方法
*/

// 将SpanListQueryParams转成听云的查询参数
func (tyAdapter *tingYunwAdapter) convertQueryParams(ctx context.Context, params interfaces.SpanListQueryParams) (queryParams string, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 通过tingYunwAdapter获取span列表")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	queryParams = ""

	// 处理sort和direction
	if params.Sort != interfaces.DEFAULT_SORT {
		errDetails := fmt.Sprintf("TingYun does not support this sort field %v", params.Sort)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		return "", rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_InvalidParameter_Sort).
			WithErrorDetails(errDetails)
	}
	queryParams += fmt.Sprintf("sortField=timestamp&sortDirection=%s", params.Direction)

	// 处理offset和limit
	queryParams += fmt.Sprintf("&pageNumber=1&pageSize=%v", params.Offset+params.Limit)

	// 处理conditdion
	if params.Condition != nil {
		queryCond, err := tyAdapter.convertQueryCondition(ctx, params.Condition)
		if err != nil {
			logger.Error(err.Error())
			o11y.Error(ctx, err.Error())
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, uerrors.Uniquery_DataView_InvalidParameter_Filters).
				WithErrorDetails(err.Error())
		}
		queryParams += queryCond
	}

	if !strings.Contains(queryParams, "endTime") {
		endTimeStr := time.Now().Format(TINGYUN_TIME_FORMAT)
		// 默认查7天的数据
		queryParams += fmt.Sprintf("&timePeriod=%d&endTime=%s", DEFAULT_TIME_PERIOD, endTimeStr)
	}

	// 处理traceID
	if params.TraceID != "_all" {
		queryParams += fmt.Sprintf("&traceGuid=%s", params.TraceID)
	}

	return queryParams, nil
}

// 将condition.CondCfg转成听云的过滤条件
func (tyAdapter *tingYunwAdapter) convertQueryCondition(ctx context.Context, condCfg *cond.CondCfg) (queryParams string, err error) {
	switch condCfg.Operation {
	case cond.OperationAnd:
		for _, subCond := range condCfg.SubConds {
			subParams, err := tyAdapter.convertQueryCondition(ctx, subCond)
			if err != nil {
				return queryParams, err
			}
			queryParams += subParams
		}
		return queryParams, nil
	case cond.OperationRange:
		if condCfg.Name != interfaces.MetaField_Timestamp && condCfg.Name != "timestamp" {
			return queryParams, errors.New("the tingyun only supports range queries for fields in [__start_time, @timestamp, timestamp]")
		}

		if condCfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
			return queryParams, fmt.Errorf("the range condition does not support value from type(%s)", condCfg.ValueFrom)
		}

		val, ok := condCfg.ValueOptCfg.Value.([]any)
		if !ok || len(val) != 2 {
			return queryParams, errors.New("the range condition right value should be an array of length 2")
		}

		startTimeFloat, ok := val[0].(float64)
		if !ok {
			return queryParams, errors.New("the start time format is incorrect, not float64")
		}
		startTime := int64(startTimeFloat)

		if num := startTime / 1e12; num < 1 || num >= 10 {
			return queryParams, errors.New("the start time format is incorrect, not unix_milli")
		}

		endTimeFloat, ok := val[1].(float64)
		if !ok {
			return queryParams, errors.New("the end time format is incorrect, not float64")
		}
		endTime := int64(endTimeFloat)

		if num := endTime / 1e12; num < 1 || num >= 10 {
			return queryParams, errors.New("the end time format is incorrect, not unix_milli")
		}

		if startTime > endTime {
			return queryParams, errors.New("the start time is longer than the end time")
		}

		// AnyRobot选择时间过滤时, 如果选择"今年", 会传入未来时间戳.
		// 而听云不接受未来时间, 会报错, 下述代码为兼容上述场景而增加.
		currentTime := time.Now().UnixMilli()
		if endTime > currentTime {
			endTime = currentTime
		}

		if startTime > currentTime {
			startTime = currentTime
		}

		endTimeStr := time.UnixMilli(endTime).Format(TINGYUN_TIME_FORMAT)
		timePeriod := (endTime - startTime) / 60000
		timePeriod = min(timePeriod, tyAdapter.appSetting.ThirdParty.TingYunMaxTimePeriod)
		return fmt.Sprintf("&timePeriod=%d&endTime=%v", timePeriod, endTimeStr), nil
	case cond.OperationEq:
		if condCfg.ValueOptCfg.ValueFrom != vopt.ValueFrom_Const {
			return queryParams, fmt.Errorf("the tingyun condition does not support value from type(%s)", condCfg.ValueFrom)
		}
		return fmt.Sprintf("&%s=%v", condCfg.Name, condCfg.Value), nil
	default:
		return "", fmt.Errorf("the tingyun does not support operation %v", condCfg.Operation)
	}
}

// 将听云Trace列表中的Item抽象为Span
func (tyAdapter *tingYunwAdapter) extractRawTraceEntry(_ context.Context, rawTraceEntry map[string]any) interfaces.AbstractSpan {
	abstractSpan := interfaces.AbstractSpan{}

	// 1. 提取Name
	abstractSpan.Name = common.Any2String(rawTraceEntry["actionName"])

	// 2. 提取TraceID
	abstractSpan.TraceID = common.Any2String(rawTraceEntry["traceGuid"])

	// 3. 提取SpanID
	abstractSpan.SpanID = abstractSpan.TraceID + "_-1"

	// 4. 提取ParentSpanID
	abstractSpan.ParentSpanID = ""

	// 5. 提取StartTime, EndTime和Duration
	respTime, _ := rawTraceEntry["respTime"].(float64)
	abstractSpan.Duration = int64(respTime * 1e3)
	timestamp, _ := rawTraceEntry["timestamp"].(float64)
	abstractSpan.StartTime = int64(timestamp * 1e3)
	abstractSpan.EndTime = abstractSpan.StartTime + abstractSpan.Duration

	// 6. 提取Kind
	actionType := common.Any2String(rawTraceEntry["actionType"])
	switch actionType {
	case "IF":
		abstractSpan.Kind = interfaces.SPAN_KIND_SERVER
	case "BG", "TX":
		abstractSpan.Kind = interfaces.SPAN_KIND_INTERNAL
	default:
		if val, ok := interfaces.SPAN_KIND_MAP[strings.ToLower(actionType)]; ok {
			abstractSpan.Kind = val
		} else {
			abstractSpan.Kind = interfaces.SPAN_KIND_UNSPECIFIED
		}
	}

	// 7. 提取Status
	if errCount := common.Any2String(rawTraceEntry["errorCount"]); errCount == "0" || errCount == "" {
		abstractSpan.Status = interfaces.SPAN_STATUS_OK
	} else {
		abstractSpan.Status = interfaces.SPAN_STATUS_ERROR
	}

	// 8. 提取ServiceName
	abstractSpan.ServiceName = common.Any2String((rawTraceEntry["applicationName"]))

	return abstractSpan
}

// 转换听云Trace详情
func (tyAdapter *tingYunwAdapter) parseRawTraceDetail(ctx context.Context, parentSeqStr string, traceID string, rawTraceDetail map[string]any, briefSpanMap map[string]*interfaces.BriefSpan_, detailSpanMap map[string]interfaces.SpanDetail) error {
	if len(rawTraceDetail) == 0 {
		return nil
	}

	var (
		abstractSpan interfaces.AbstractSpan
		isRootSpan   bool
		seqStr       string
	)

	seqInf, ok := rawTraceDetail["seq"]
	if !ok {
		isRootSpan = true
		seqStr = "-1"
	} else {
		seqStr = common.Any2String(seqInf)
	}

	if isRootSpan { // 说明是root span
		abstractSpan = tyAdapter.extractRawTraceEntry(ctx, rawTraceDetail)
	} else {
		// 1. 提取Name
		abstractSpan.Name = common.Any2String(rawTraceDetail["clasz"]) + common.Any2String(rawTraceDetail["method"])

		// 2. 提取TraceID
		abstractSpan.TraceID = traceID

		// 3. 提取SpanID
		abstractSpan.SpanID = traceID + "_" + seqStr

		// 4. 提取ParentSpanID
		abstractSpan.ParentSpanID = traceID + "_" + parentSeqStr

		// 5. 提取StartTime, EndTime和Duration
		totalTime, _ := rawTraceDetail["totalTime"].(float64)
		abstractSpan.Duration = int64(totalTime * 1e3)
		startTime, _ := rawTraceDetail["startTime"].(float64)
		abstractSpan.StartTime = int64(startTime * 1e3)
		abstractSpan.EndTime = abstractSpan.StartTime + abstractSpan.Duration

		// 6. 提取Kind
		// Service,Exception,External,Database,NoSQL,Pool,MQ,Code,Dataitem
		actionType := common.Any2String(rawTraceDetail["metricType"])
		switch strings.ToLower(actionType) {
		case "service":
			abstractSpan.Kind = interfaces.SPAN_KIND_SERVER
		case "code", "pool", "exception", "dataitem":
			abstractSpan.Kind = interfaces.SPAN_KIND_INTERNAL
		case "external", "database", "nosql":
			abstractSpan.Kind = interfaces.SPAN_KIND_CLIENT
		case "mq":
			if strings.Contains(abstractSpan.Name, "producer") {
				abstractSpan.Kind = interfaces.SPAN_KIND_PRODUCER
			} else if strings.Contains(abstractSpan.Name, "consumer") {
				abstractSpan.Kind = interfaces.SPAN_KIND_CONSUMER
			} else {
				abstractSpan.Kind = interfaces.SPAN_KIND_UNSPECIFIED
			}
			abstractSpan.Kind = interfaces.SPAN_KIND_INTERNAL
		default:
			if val, ok := interfaces.SPAN_KIND_MAP[strings.ToLower(actionType)]; ok {
				abstractSpan.Kind = val
			} else {
				abstractSpan.Kind = interfaces.SPAN_KIND_UNSPECIFIED
			}
		}

		// 7. 提取Status
		if _, ok := rawTraceDetail["errors"]; !ok {
			abstractSpan.Status = interfaces.SPAN_STATUS_OK
		} else {
			abstractSpan.Status = interfaces.SPAN_STATUS_ERROR
		}

		// 8. 提取ServiceName
		callerApplicationInf, ok := rawTraceDetail["callerApplication"]
		if ok {
			callerApplication, _ := callerApplicationInf.(map[string]any)
			abstractSpan.ServiceName = common.Any2String(callerApplication["name"])
		}
	}

	briefSpan := &interfaces.BriefSpan_{
		Key:          abstractSpan.SpanID,
		Name:         abstractSpan.Name,
		SpanID:       abstractSpan.SpanID,
		ParentSpanID: abstractSpan.ParentSpanID,
		StartTime:    abstractSpan.StartTime,
		EndTime:      abstractSpan.EndTime,
		Duration:     abstractSpan.Duration,
		Kind:         abstractSpan.Kind,
		Status:       abstractSpan.Status,
		ServiceName:  abstractSpan.ServiceName,
		Children:     make([]*interfaces.BriefSpan_, 0),
	}
	briefSpanMap[abstractSpan.SpanID] = briefSpan

	if isRootSpan { // root span没有seq
		timeLine, _ := rawTraceDetail["timeLine"].(map[string]any)
		err := tyAdapter.parseRawTraceDetail(ctx, "-1", traceID, timeLine, briefSpanMap, detailSpanMap)
		if err != nil {
			return err
		}
		delete(rawTraceDetail, "timeline")
	} else {
		subTimelineInfs, _ := rawTraceDetail["subTimeLines"].([]any)
		for _, subTimelineInf := range subTimelineInfs {
			subTimeline, _ := subTimelineInf.(map[string]any)
			err := tyAdapter.parseRawTraceDetail(ctx, common.Any2String(seqStr), traceID, subTimeline, briefSpanMap, detailSpanMap)
			if err != nil {
				return err
			}
		}
		delete(rawTraceDetail, "subTimeLines")
	}

	detailSpanMap[abstractSpan.SpanID] = tyAdapter.genSpanDetail(rawTraceDetail, abstractSpan)
	return nil
}

// 根据abstractSpan补充rawSpan
func (tyAdapter *tingYunwAdapter) genSpanDetail(rawSpan map[string]any, abstractSpan interfaces.AbstractSpan) map[string]any {
	if rawSpan == nil {
		rawSpan = make(map[string]any)
	}

	rawSpan["__trace_id"] = abstractSpan.TraceID
	rawSpan["__span_id"] = abstractSpan.SpanID
	rawSpan["__parent_span_id"] = abstractSpan.ParentSpanID
	rawSpan["__name"] = abstractSpan.Name
	rawSpan["__start_time"] = abstractSpan.StartTime
	rawSpan["__end_time"] = abstractSpan.EndTime
	rawSpan["__duration"] = abstractSpan.Duration
	rawSpan["__kind"] = abstractSpan.Kind
	rawSpan["__status"] = abstractSpan.Status
	rawSpan["__service_name"] = abstractSpan.ServiceName

	return rawSpan
}

func (tyAdapter *tingYunwAdapter) genTraceDetailQueryParams(params interfaces.TraceQueryParams) string {
	tyParams := fmt.Sprintf("traceGuid=%s&filterList=Service,Exception,External,Database,NoSQL,Pool,MQ,Code,Dataitem", params.TraceID)

	// 情况1: 传了合法的__start_time和__end_time
	startTimeInf, ok1 := params.Context["__start_time"]
	if !ok1 {
		logger.Warn("The __start_time is not exist in filed context")
	}

	startTimeFloat, ok1 := startTimeInf.(float64)
	if !ok1 {
		logger.Warn("The __start_time type in filed context is incorrect, not float64")
	}

	startTimeInt := int64(startTimeFloat)
	if num := startTimeInt / 1e15; num < 1 || num >= 10 {
		logger.Warnf("The __start_time format in filed context is incorrect, not unix_micro")
	}

	endTimeInf, ok2 := params.Context["__end_time"]
	if !ok2 {
		logger.Warn("The __end_time is not exist in filed context")
	}

	endTimeFloat, ok2 := endTimeInf.(float64)
	if !ok2 {
		logger.Warn("The __end_time type in filed context is incorrect, not float64")
	}

	endTimeInt := int64(endTimeFloat)
	if num := endTimeInt / 1e15; num < 1 || num >= 10 {
		logger.Warnf("The __end_time format in filed context is incorrect, not unix_micro")
	}

	if ok1 && ok2 {
		// 结束时间往后延长10分钟
		endTime := time.UnixMicro(endTimeInt).Add(10 * time.Minute).Format(TINGYUN_TIME_FORMAT)
		// 时间间隔补20分钟, 相当于startTime往前移10分钟
		timePeriod := (endTimeInt - startTimeInt + 20*60*1e6) / (60 * 1e6)
		timePeriod = min(timePeriod, tyAdapter.appSetting.ThirdParty.TingYunMaxTimePeriod)

		tyParams += fmt.Sprintf("&endTime=%v&timePeriod=%v", endTime, timePeriod)
		return tyParams
	}

	// 情况2: __start_time和__end_time不合法/未传, 但传了合法的@timestamp
	timestampInf, ok3 := params.Context["@timestamp"]
	if !ok3 {
		logger.Warn("The @timestamp is not exist in filed context")
	}

	timestampString, ok3 := timestampInf.(string)
	if !ok3 {
		logger.Warn("The @timestamp type in filed context is incorrect, not string")
	}

	timestamp, err := time.Parse("2006-01-02T15:04:05.999Z07:00", timestampString)
	if err != nil {
		logger.Warn("Failed to convert @timestamp to time.Time, err: %s", err.Error())
	}

	if ok3 && err == nil {
		timeZone, _ := time.LoadLocation(os.Getenv("TZ"))
		// 以timestamp为轴, 向左右各取10分钟作为start_time和end_time
		endTime := timestamp.Add(10 * time.Minute).In(timeZone).Format(TINGYUN_TIME_FORMAT)
		timePeriod := 20

		tyParams += fmt.Sprintf("&endTime=%v&timePeriod=%v", endTime, timePeriod)
		return tyParams
	}

	// 情况1和情况2都不满足, 默认查询时间范围为近1天
	endTime := time.Now().Format(TINGYUN_TIME_FORMAT)
	timePeriod := 1440

	tyParams += fmt.Sprintf("&endTime=%v&timePeriod=%v", endTime, timePeriod)
	return tyParams
}

func (tyAdapter *tingYunwAdapter) getTraceList(ctx context.Context, cfg TingYunClientConfig) (traceList []map[string]any, total int64, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 查询听云系统获取trace列表", trace.WithSpanKind(trace.SpanKindClient))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 请求听云, 获取trace列表
	tyURL := fmt.Sprintf("%s://%s/server-api/action/trace?%s", cfg.Protocol, cfg.Address, cfg.QueryParams)
	headers := map[string]string{
		"Authorization": "Bearer " + cfg.AccessToken,
		"Content-Type":  "application/json",
		"X-Language":    rest.GetLanguageByCtx(ctx),
	}

	span.SetAttributes(
		attr.Key("tingyun_url").String(tyURL),
	)

	queryValues := url.Values{}
	respCode, respBody, err := tyAdapter.httpClient.GetNoUnmarshal(ctx, tyURL, queryValues, headers)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get tingyun trace list: %s", err)
		logger.Errorf(errDetails)
		return []map[string]any{}, 0, err
	}

	if respCode != http.StatusOK {
		err := fmt.Errorf("failed to get tingyun trace list: %s", string(respBody))
		logger.Error(err.Error())
		return []map[string]any{}, 0, err
	}

	// 2. 解析resBody
	dataInfo := struct {
		Code int `json:"code"`
		Data struct {
			Total   int              `json:"totalElements"`
			Content []map[string]any `json:"content"`
		} `json:"data"`
	}{}
	err = sonic.Unmarshal(respBody, &dataInfo)
	if err != nil {
		errWrap := fmt.Errorf("failed to unmarshal respBody after getting tingyun trace list, err: %v, resp is: %s", err.Error(), respBody)
		logger.Error(errWrap.Error())
		return []map[string]any{}, 0, errWrap
	}

	if dataInfo.Code != http.StatusOK {
		err := fmt.Errorf("failed to get tingyun trace list: %s", string(respBody))
		logger.Error(err.Error())
		return []map[string]any{}, 0, err
	}

	return dataInfo.Data.Content, int64(dataInfo.Data.Total), nil
}

func (tyAdapter *tingYunwAdapter) getTraceDetail(ctx context.Context, cfg TingYunClientConfig) (traceDetail map[string]any, isExist bool, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic层: 查询听云系统获取trace详情", trace.WithSpanKind(trace.SpanKindClient))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	// 1. 请求听云, 获取trace详情
	tyURL := fmt.Sprintf("%s://%s/server-api/action/trace/detail?%s", cfg.Protocol, cfg.Address, cfg.QueryParams)
	headers := map[string]string{
		"Authorization": "Bearer " + cfg.AccessToken,
		"Content-Type":  "application/json",
		"X-Language":    rest.GetLanguageByCtx(ctx),
	}

	span.SetAttributes(
		attr.Key("tingyun_url").String(tyURL),
	)

	queryValues := url.Values{}
	respCode, respBody, err := tyAdapter.httpClient.GetNoUnmarshal(ctx, tyURL, queryValues, headers)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get tingyun trace detail: %s", err)
		logger.Errorf(errDetails)
		return nil, false, err
	}

	if respCode != http.StatusOK {
		err := fmt.Errorf("failed to get tingyun trace detail: %s", string(respBody))
		logger.Error(err.Error())
		return nil, false, err
	}

	// 2. 解析resBody
	node, err := sonic.Get(respBody, "code")
	if err != nil {
		errDetails := fmt.Sprintf("Using sonic to get code failed: %s", err.Error())
		logger.Errorf(errDetails)
		return nil, false, err
	}

	code, err := node.Int64()
	if err != nil {
		errDetails := fmt.Sprintf("Using sonic to get code failed: %s", err.Error())
		logger.Errorf(errDetails)
		return nil, false, err
	}

	if code == http.StatusNotFound {
		return nil, false, nil
	}

	if code != http.StatusOK {
		dataInfo := struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}{}
		err = sonic.Unmarshal(respBody, &dataInfo)
		if err != nil {
			errDetails := fmt.Sprintf("Failed to unmarshal respBody after getting tingyun trace detail, err: %v", err.Error())
			logger.Error(errDetails)
			return nil, false, err
		}

		return nil, false, errors.New(dataInfo.Msg)
	}

	dataInfo := struct {
		Code int            `json:"code"`
		Data map[string]any `json:"data"`
	}{}
	err = sonic.Unmarshal(respBody, &dataInfo)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to unmarshal respBody after getting tingyun trace detail, err: %v", err.Error())
		logger.Error(errDetails)
		return nil, false, err
	}

	return dataInfo.Data, true, nil
}

func (tyAdapter *tingYunwAdapter) processDataConnection(_ context.Context, conn *interfaces.DataConnection) error {

	if conn.DataSourceType != interfaces.SOURCE_TYPE_TINGYUN {
		errDetails := fmt.Sprintf("Invalid data_source_type: %v", conn.DataSourceType)
		logger.Error(errDetails)
		return errors.New(errDetails)
	}

	b, err := sonic.Marshal(conn.DataSourceConfig)
	if err != nil {
		errDetails := fmt.Sprintf("Marshal field config failed, err: %v", err.Error())
		logger.Error(errDetails)
		return err
	}

	conf := TingYunDetailedConfig{}
	err = sonic.Unmarshal(b, &conf)
	if err != nil {
		errDetails := fmt.Sprintf("Field config cannot be unmarshaled to TingYunDetailedConfig, err: %v", err.Error())
		logger.Error(errDetails)
		return err
	}

	conn.DataSourceConfig = conf
	return nil
}
