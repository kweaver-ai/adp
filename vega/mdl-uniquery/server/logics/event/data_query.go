// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package event

import (
	"context"
	"encoding/json"
	"reflect"
	"uniquery/common"
	cond "uniquery/common/condition"
	"uniquery/common/convert"
	vopt "uniquery/common/value_opt"
	"uniquery/interfaces"
	"uniquery/logics"
	"uniquery/logics/data_view"
	"uniquery/logics/metric_model"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

type MetricDataQuery struct {
	DataSource     []string
	DataSourceType string
	TimeInterval   interfaces.TimeInterval
	Start          int64
	End            int64
	Step           string
	Filters        []interfaces.Filter
	mmService      interfaces.MetricModelService
}

func NewMetricDataQuery(appSetting *common.AppSetting, dataSource []string, dataSourceType string, t interfaces.TimeInterval, query interfaces.EventQuery) interfaces.DataModelQuery {

	return &MetricDataQuery{
		DataSource:     dataSource,
		DataSourceType: query.DataSourceType,
		TimeInterval:   t,
		End:            query.End,
		Start:          query.Start,
		Step:           query.Step,
		Filters:        query.Extraction,
		// mmService:      logics.MetricModelService,
		mmService: metric_model.NewMetricModelService(appSetting),
	}
}

func (md *MetricDataQuery) FetchSourceRecordsFrom(ctx context.Context, format string) (interfaces.SourceRecords, interfaces.Record, error) {

	//NOTE:不传，走指标模型即时查询
	var query interfaces.MetricModelQuery

	if md.Step == "" {
		query = interfaces.MetricModelQuery{
			QueryTimeParams: interfaces.QueryTimeParams{
				IsInstantQuery: true,
				Start:          &md.Start,
				End:            &md.End,
				// LookBackDelta:  md.End - md.Start,
				// Time: md.End,
			},
			MetricModelID: md.DataSource[0],
		}
	} else { //走范围查询
		stepT, _ := convert.ParseDuration(md.Step)
		Step := stepT.Milliseconds()
		query = interfaces.MetricModelQuery{
			QueryTimeParams: interfaces.QueryTimeParams{
				IsInstantQuery: false,
				Start:          &md.Start,
				End:            &md.End,
				StepStr:        &md.Step,
				Step:           &Step,
			},
			MetricModelID: md.DataSource[0],
		}
	}
	// 事件模型查询指标模型数据，不分页
	query.Limit = interfaces.DEFAULT_SERIES_LIMIT_INT

	uniResponse, _, _, err := md.mmService.Exec(ctx, &query)

	if err != nil {
		logger.Errorf("source records will be Null, because metric model return error,error info is: %v", err)
		return interfaces.SourceRecords{}, interfaces.Record{}, err
	}
	var sr interfaces.SourceRecords

	//解构指标模型，前两个循环都是一次
	for _, rs := range uniResponse.Datas {
		for index, st := range rs.Times {
			logger.Debugf("source records for event-> val:%#v,time:%#v,label: %#v\n", index, rs.Values[index], st, rs.Labels)
			_, ok := rs.Values[index].(string) //剔除正负无穷大等字符串值
			if ok {
				continue
			}
			if rs.Values[index] == nil { //剔除null等无效值
				continue
			}

			new_record := map[string]interface{}{"value": rs.Values[index], "@timestamp": st}
			for key, value := range rs.Labels {
				new_record["labels."+key] = value
			}

			sr.Records = append(sr.Records, new_record)

		}
	}

	return sr, nil, nil
}

type LogDataQuery struct {
	DataSource      []string
	DataSourceType  string
	Start           int64
	End             int64
	TimeInterval    interfaces.TimeInterval
	Condition       *cond.CondCfg
	dataViewService interfaces.DataViewService
	osAccess        interfaces.OpenSearchAccess
}

func NewLogDataQuery(appSetting *common.AppSetting, dataSource []string, dataSourceType string, t interfaces.TimeInterval, query interfaces.EventQuery) interfaces.DataModelQuery {

	return &LogDataQuery{
		DataSource:     dataSource,
		DataSourceType: "data_view",
		Start:          query.Start,
		End:            query.End,
		TimeInterval:   t,
		Condition:      convertFiltersToCondition(query.Extraction),
		// LookBackDelta:  query.LookBackDelta,
		dataViewService: data_view.NewDataViewService(appSetting),
		osAccess:        logics.OSAccess,
	}
}

func (dv *LogDataQuery) FetchSourceRecordsFrom(ctx context.Context, format string) (interfaces.SourceRecords, interfaces.Record, error) {
	//NOTE 构造视图模型查询参数
	id := dv.DataSource[0]
	logger.Debugf("read data view durration: [%d,%d]", dv.Start, dv.End)

	var query = &interfaces.DataViewQueryV1{
		ViewQueryCommonParams: interfaces.ViewQueryCommonParams{
			Start:  dv.Start,
			End:    dv.End,
			Limit:  10000,
			Format: format,
		},
		// NOTE: 这个不要改，影响"清除的事件"，"清除的事件"的生成依赖这个排序
		SortParamsV1: interfaces.SortParamsV1{
			Sort:      "@timestamp",
			Direction: "asc",
		},
		Scroll:             "2m",
		GlobalFilters:      dv.Condition,
		AllowNonExistField: true,
	}
	var sr interfaces.SourceRecords
	var flag = true
	var lastRequest = true
	var uniResponse interfaces.ViewUniResponseV1
	var scrollId string
	scrollIds := make([]string, 0)
	var model = interfaces.Record{}
	for {
		if flag {
			respV2, err := dv.dataViewService.GetSingleViewData(ctx, id, query)
			if err != nil {
				logger.Errorf("source records will be null, because data view model return error,error info is: %v", err)
				return interfaces.SourceRecords{}, nil, err
			}
			uniResponse = interfaces.ViewUniResponseV1{
				ScrollId: respV2.ScrollId,
				View:     respV2.View,
				Datas: []interfaces.ViewData{{
					Total:  respV2.TotalCount,
					Values: respV2.Entries,
				}},
			}

			b, _ := json.Marshal(uniResponse.View)
			err = json.Unmarshal(b, &model)
			if err != nil {
				logger.Errorf("unmarshal error: %v", err)
				return interfaces.SourceRecords{}, nil, err
			}
			// if len(uniResponses[0].Datas[0].Values)==0{
			// 	logger.Errorf("source records will be null, because data view model return null")
			// 	return interfaces.SourceRecords{}, err
			// }

			scrollId = uniResponse.ScrollId
		} else {
			var scrollQuery = &interfaces.DataViewQueryV1{
				Scroll:             "2m",
				ScrollId:           scrollId,
				GlobalFilters:      dv.Condition,
				AllowNonExistField: true,
			}
			respV2, err := dv.dataViewService.GetSingleViewData(ctx, id, scrollQuery)
			if err != nil {
				logger.Errorf("source records will be null, because data view return error,error info is: %v", err)
				return interfaces.SourceRecords{}, nil, err
			}
			uniResponse = interfaces.ViewUniResponseV1{
				ScrollId: respV2.ScrollId,
				View:     respV2.View,
				Datas: []interfaces.ViewData{{
					Total:  respV2.TotalCount,
					Values: respV2.Entries,
				}},
			}

			if len(uniResponse.Datas) == 0 {
				logger.Errorf("source records will be null, because data view model return null")
				return interfaces.SourceRecords{}, nil, err
			}
		}

		//TOO: 增加无字段属性时，跳出循环。
		for _, rs := range uniResponse.Datas {
			if len(rs.Values) == 0 && !flag {
				lastRequest = false
				break
			}
			for _, st := range rs.Values {
				if _, ok := reflect.TypeOf(rs).FieldByName("Labels"); ok {
					lables := reflect.ValueOf(rs).FieldByName("Labels").Interface().(map[string]string)
					for key, value := range lables {
						st[key] = value
					}
				}
				sr.Records = append(sr.Records, st)
			}
		}
		flag = false
		if !lastRequest {
			break
		}
	}
	scrollIds = append(scrollIds, scrollId)
	if len(scrollIds) > 0 {
		go clearScrollIds(dv, scrollIds)
	}

	return sr, model, nil
}

type EventDataQuery struct {
	DataSources    []string
	DataSourceType string
	TimeInterval   interfaces.TimeInterval
	// LookBackDelta  string
	// QueryTime      int64
	Start        int64
	End          int64
	Extraction   []interfaces.Filter
	Filters      []interfaces.Filter
	eventService interfaces.EventService
}

func NewEventDataQuery(appSetting *common.AppSetting, dataSources []string, dataSourceType string, t interfaces.TimeInterval, query interfaces.EventQuery) interfaces.DataModelQuery {
	return &EventDataQuery{
		DataSources:    dataSources,
		DataSourceType: "event_model",
		TimeInterval:   t,
		Start:          query.Start,
		End:            query.End,
		Extraction:     query.Extraction,
		Filters:        query.Filters,
		eventService:   NewEventService(appSetting),
	}
}

func (md *EventDataQuery) FetchSourceRecordsFrom(ctx context.Context, format string) (interfaces.SourceRecords, interfaces.Record, error) {
	// var eventSlices any
	var querys []interfaces.EventQuery
	for _, dataSource := range md.DataSources {
		query := interfaces.EventQuery{
			QueryType:  "range_query",
			Start:      md.Start,
			End:        md.End,
			Id:         dataSource,
			Limit:      -1,
			Offset:     0,
			Filters:    md.Filters,
			Extraction: md.Extraction,
		}
		querys = append(querys, query)
	}
	queryRes := interfaces.EventQueryReq{
		Querys: querys,
		Limit:  0,
		Offset: 0,
	}

	_, eventSlices, _, _ := md.eventService.Query(ctx, queryRes)
	eventSlicess, ok := eventSlices.([]interfaces.IEvents)

	var sr interfaces.SourceRecords
	if ok {
		for _, events := range eventSlicess {
			for _, event := range events {
				sr.Records = append(sr.Records, map[string]interface{}{
					"id":               event.GetBaseEvent().Id,
					"type":             event.GetBaseEvent().EventType,
					"@timestamp":       event.GetBaseEvent().CreateTime,
					"trigger_time":     event.GetTriggerTime(),
					"level":            event.GetBaseEvent().Level,
					"event_model_id":   event.GetBaseEvent().EventModelId,
					"event_model_name": event.GetBaseEvent().EventModelName,
					"tags":             event.GetTag(),
					"title":            event.GetBaseEvent().Title,
					"trigger_data":     event.GetTriggerData(),
					"relations":        event.GetBaseEvent().Relations,
				})
			}
		}
	}

	return sr, nil, nil
}

// NOTE: 这里有空可以改用工厂模式，摈弃这种简单工厂模式
func GenerateDataModelQuery(appSetting *common.AppSetting, em interfaces.EventModel, query interfaces.EventQuery) (_query interfaces.DataModelQuery) {
	if em.DataSourceType == "metric_model" {
		return NewMetricDataQuery(appSetting, em.DataSource, em.DataSourceType, em.DefaultTimeWindow, query)
	} else if em.DataSourceType == "data_view" {
		return NewLogDataQuery(appSetting, em.DataSource, em.DataSourceType, em.DefaultTimeWindow, query)
	} else {
		return NewEventDataQuery(appSetting, em.DataSource, em.DataSourceType, em.DefaultTimeWindow, query)
	}
}

func convertFiltersToCondition(filters []interfaces.Filter) (condition *cond.CondCfg) {
	if len(filters) == 0 {
		return
	} else if len(filters) == 1 {
		filter := filters[0]
		if filter.Operation == interfaces.OPERATION_EQ {
			filter.Operation = cond.OperationEq
		}

		condition = &cond.CondCfg{
			Operation: filter.Operation,
			Name:      filter.Name,
			ValueOptCfg: vopt.ValueOptCfg{
				ValueFrom: vopt.ValueFrom_Const,
				Value:     filter.Value,
			},
		}
	} else {
		subConds := make([]*cond.CondCfg, 0, len(filters))
		for i := 0; i < len(filters); i++ {
			filter := filters[i]
			if filter.Operation == interfaces.OPERATION_EQ {
				filter.Operation = cond.OperationEq
			}

			subConds = append(subConds, &cond.CondCfg{
				Operation: filter.Operation,
				Name:      filter.Name,
				ValueOptCfg: vopt.ValueOptCfg{
					ValueFrom: vopt.ValueFrom_Const,
					Value:     filter.Value,
				},
			})
		}

		condition = &cond.CondCfg{
			Operation: cond.OperationAnd,
			SubConds:  subConds,
		}
	}

	return
}

func clearScrollIds(dv *LogDataQuery, scrollIds []string) {
	para := interfaces.DeleteScroll{
		ScrollId: scrollIds,
	}
	logger.Debugf("clear scroll ids: %v", scrollIds)

	// 必须接返回值, 否则代码检查通不过
	_, _, err := dv.osAccess.DeleteScroll(context.Background(), para)
	if err != nil {
		logger.Errorf("LogDataQuery clear scroll ids failed: %v", err)
	}
}
