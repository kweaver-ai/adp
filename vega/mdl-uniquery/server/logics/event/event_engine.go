// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package event

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"

	"uniquery/common"
	operatation "uniquery/common/compute"
	"uniquery/common/convert"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics"
)

var PreOrderEvent sync.Map

var (
	engineOnce sync.Once
	engine     interfaces.EventEngine
)

type EventEngine struct {
	EventModel     interfaces.EventModel
	DataSource     any
	DataSourceType string
	DVAccess       interfaces.DataViewAccess
	KAccess        interfaces.KafkaAccess
	appSetting     *common.AppSetting
}

func (ds *eventService) InitEngine(ctx context.Context, event_model interfaces.EventModel) error {

	//FIXME:这里需要更新预览标记，进行查询引擎初始化。
	engineOnce.Do(func() {
		engine = &EventEngine{
			appSetting:     ds.appSetting,
			EventModel:     event_model,
			DataSource:     event_model.DataSource,
			DataSourceType: event_model.EventModelType,
			DVAccess:       logics.DVAccess,
			// UAcess:         logics.UAccess,
			KAccess: logics.KAccess,
		}
	})

	ds.engine = engine
	return nil
}

func (eventEngine *EventEngine) Apply(ctx context.Context, query interfaces.EventQuery,
	em interfaces.EventModel) (interfaces.IEvents, interfaces.Records, int) {

	var _events interfaces.IEvents
	var entities interfaces.Records
	var total int

	if query.QueryType == "instant_query" {
		// 实时查询，查事件模型绑定的数据源的数据，再根据检测生成事件数据
		_events, entities, total = eventEngine.InstantQuery(ctx, query, em)
	} else if query.QueryType == "range_query" {
		// 即持久化查询，查事件模型持久化后的数据
		_events, entities, total = eventEngine.RangeQuery(ctx, query, em)
	}

	return _events, entities, total
}

func (eventEngine *EventEngine) InstantQuery(ctx context.Context, query interfaces.EventQuery,
	em interfaces.EventModel) (interfaces.IEvents, interfaces.Records, int) {

	var _events interfaces.IEvents
	var entities interfaces.Records
	var total int

	_events, entities = eventEngine.RuleDetect(ctx, query, em)

	if len(_events) > 0 {
		_events, total, _ = eventEngine.CombineFilter(_events, query.Filters)
		_events, _ = eventEngine.Assemble(_events, query.SortKey, query.Direction, query.Limit, query.Offset)
	}

	return _events, entities, total
}

func (eventEngine *EventEngine) RangeQuery(ctx context.Context, query interfaces.EventQuery, em interfaces.EventModel) (interfaces.IEvents, interfaces.Records, int) {
	// todo: error被忽略，无法返回
	_events, _ := eventEngine.PersistQuery(ctx, query, em, true, true)
	_events, total, _ := eventEngine.CombineFilter(_events, query.Filters)
	_events, _ = eventEngine.Assemble(_events, query.SortKey, query.Direction, query.Limit, query.Offset)
	return _events, nil, total
}

// 规则检测
func (eventEngine *EventEngine) RuleDetect(ctx context.Context, query interfaces.EventQuery, em interfaces.EventModel) (interfaces.IEvents, interfaces.Records) {
	sourceRecords, _, _ := eventEngine.Query(ctx, query, em)
	if len(sourceRecords.Records) == 0 {
		return nil, nil
	}

	sourceRecords, _ = eventEngine.Process(ctx, sourceRecords, em)
	if len(sourceRecords.Records) == 0 {
		return nil, nil
	}

	_events, _, _ := eventEngine.Judge(ctx, sourceRecords, em)
	return _events, nil
}

func (eventEngine *EventEngine) CombineFilter(events interfaces.IEvents, filters []interfaces.Filter) (interfaces.IEvents, int, error) {
	//NOTE：结果过滤过程，仅支持标签和等级过滤
	var finalEvents interfaces.IEvents
	if len(filters) == 0 {
		finalEvents = events
		return finalEvents, len(events), nil
	}
	for _, event := range events {
		ok, _ := eventEngine.mit(event, filters)
		if ok {
			finalEvents = append(finalEvents, event)
		}
	}

	return finalEvents, len(finalEvents), nil
}

func (eventEngine *EventEngine) mit(event interfaces.IEvent, filters []interfaces.Filter) (bool, error) {
	for _, filter := range filters {
		EventValue := reflect.ValueOf(event)
		fieldValue := EventValue.FieldByName(interfaces.FieldNameMap[filter.Name])
		if filter.Name != "level" && filter.Name != "tags" && filter.Name != "type" && filter.Name != "id" {
			logger.Errorf("sorry,filter column %v is unsupported!", filter.Name)
			return true, nil
		}

		switch filter.Operation {
		case "in":
			filterValues := reflect.ValueOf(filter.Value)
			if filterValues.Len() > 0 {
				Elem := filterValues.Index(0)

				if _, ok := Elem.Interface().(string); ok {
					var filterValueStrings []string
					for index := 0; index < filterValues.Len(); index++ {
						filterValueStrings = append(filterValueStrings, filterValues.Index(index).Interface().(string))
					}

					var realFieldValue string
					if _, ok := fieldValue.Interface().(uint64); ok {
						realFieldValue = strconv.FormatUint(fieldValue.Interface().(uint64), 10)
					} else {
						realFieldValue = fieldValue.Interface().(string)
					}

					if hit, _ := operatation.In(realFieldValue, filterValueStrings); !hit {
						return false, nil
					}

				} else if _, ok := Elem.Interface().(float64); ok {
					var filterValueFloat64s []float64
					for index := 0; index < filterValues.Len(); index++ {
						filterValueFloat64s = append(filterValueFloat64s, filterValues.Index(index).Interface().(float64))
					}
					var realFieldValue float64
					if v, ok := fieldValue.Interface().(int); ok {
						realFieldValue = float64(v)
					} else if v, ok := fieldValue.Interface().(float64); ok {
						realFieldValue = v
					}

					if hit, _ := operatation.In(realFieldValue, filterValueFloat64s); !hit {
						return false, nil
					}
				} else {
					var filterValueInts []int
					for index := 0; index < filterValues.Len(); index++ {
						filterValueInts = append(filterValueInts, filterValues.Index(index).Interface().(int))
					}
					realFieldValue := fieldValue.Interface().(int)

					if hit, _ := operatation.In(realFieldValue, filterValueInts); !hit {
						return false, nil
					}
				}
			} else {
				logger.Errorf("the parameter of filter is illegal,please check the event model")
			}

		case "contain":
			filterValues := reflect.ValueOf(filter.Value)
			if fieldValue.Len() > 0 {
				Elem := fieldValue.Index(0)
				if Elem.Kind() == reflect.String {
					var fieldValueStrings, realFilterValue []string
					for index := 0; index < fieldValue.Len(); index++ {
						fieldValueStrings = append(fieldValueStrings, fieldValue.Index(index).Interface().(string))
					}
					for index := 0; index < filterValues.Len(); index++ {
						realFilterValue = append(realFilterValue, filterValues.Index(index).Interface().(string))
					}
					if hit, _ := operatation.Contain(fieldValueStrings, realFilterValue); !hit {
						return false, nil
					}

				} else if Elem.Kind() == reflect.Float64 {
					var fieldValueFloats, realFilterValue []float64
					for index := 0; index < fieldValue.Len(); index++ {
						fieldValueFloats = append(fieldValueFloats, fieldValue.Index(index).Interface().(float64))
					}
					// realFieldValue := fieldValue.Interface().(float64)
					for index := 0; index < filterValues.Len(); index++ {
						realFilterValue = append(realFilterValue, filterValues.Index(index).Interface().(float64))
					}
					if hit, _ := operatation.Contain(fieldValueFloats, realFilterValue); !hit {
						return false, nil
					}

				} else {
					var fieldValueInts, realFilterValue []int
					for index := 0; index < fieldValue.Len(); index++ {
						fieldValueInts = append(fieldValueInts, fieldValue.Index(index).Interface().(int))
					}
					for index := 0; index < filterValues.Len(); index++ {
						realFilterValue = append(realFilterValue, filterValues.Index(index).Interface().(int))
					}
					// realFieldValue := fieldValue.Interface().(int)
					if hit, _ := operatation.Contain(fieldValueInts, realFilterValue); !hit {
						return false, nil
					}

				}
			} else {
				return false, nil
			}

		case "=":
			switch filter.Value.(type) {
			case string:
				if hit, _ := operatation.EqualTo(filter.Value.(string), fieldValue.Interface().(string)); !hit {
					return false, nil
				}

			case float64:
				var realFieldValue float64
				if v, ok := fieldValue.Interface().(int); ok {
					realFieldValue = float64(v)
				} else if v, ok := fieldValue.Interface().(float64); ok {
					realFieldValue = v
				}
				var realFilterValue float64
				if v, ok := filter.Value.(int); ok {
					realFilterValue = float64(v)
				} else if v, ok := filter.Value.(float64); ok {
					realFilterValue = v
				}
				if hit, _ := operatation.EqualTo(realFilterValue, realFieldValue); !hit {
					return false, nil
				}

			case int:
				if hit, _ := operatation.EqualTo(filter.Value.(int), fieldValue.Interface().(int)); !hit {
					return false, nil
				}

			}

		default:
			logger.Errorf("sorry,operation %v is unsupported!", filter.Operation)
		}
	}
	return true, nil
}

func (eventEngine *EventEngine) Assemble(Events interfaces.IEvents, sortKey string, direction string, limit int64, offset int64) (interfaces.IEvents, error) {
	//NOTE：结果sort,limit,offset,direction 处理
	switch sortKey {
	case "@timestamp", "":
		if direction == "asc" {
			sort.SliceStable(Events, func(i, j int) bool {
				return Events[i].GetBaseEvent().CreateTime < Events[j].GetBaseEvent().CreateTime
			})
		} else {
			sort.SliceStable(Events, func(i, j int) bool {
				return Events[i].GetBaseEvent().CreateTime > Events[j].GetBaseEvent().CreateTime
			})
		}

	case "trigger_time":
		if direction == "asc" {
			sort.SliceStable(Events, func(i, j int) bool {
				return Events[i].GetTriggerTime() < Events[j].GetTriggerTime()
			})
		} else {
			sort.SliceStable(Events, func(i, j int) bool {
				return Events[i].GetTriggerTime() > Events[j].GetTriggerTime()
			})
		}

	case "level":
		if direction == "asc" {
			sort.SliceStable(Events, func(i, j int) bool { return Events[i].GetBaseEvent().Level < Events[j].GetBaseEvent().Level })
		} else {
			sort.SliceStable(Events, func(i, j int) bool { return Events[i].GetBaseEvent().Level > Events[j].GetBaseEvent().Level })
		}

	case "event_model_name":
		if direction == "asc" {
			sort.SliceStable(Events, func(i, j int) bool {
				return Events[i].GetBaseEvent().EventModelName < Events[j].GetBaseEvent().EventModelName
			})
		} else {
			sort.SliceStable(Events, func(i, j int) bool {
				return Events[i].GetBaseEvent().EventModelName > Events[j].GetBaseEvent().EventModelName
			})
		}

	case "type":
		if direction == "asc" {
			sort.SliceStable(Events, func(i, j int) bool { return Events[i].GetBaseEvent().EventType < Events[j].GetBaseEvent().EventType })
		} else {
			sort.SliceStable(Events, func(i, j int) bool { return Events[i].GetBaseEvent().EventType > Events[j].GetBaseEvent().EventType })
		}

	case "event_model_id":
		if direction == "asc" {
			sort.SliceStable(Events, func(i, j int) bool {
				return Events[i].GetBaseEvent().EventModelId < Events[j].GetBaseEvent().EventModelId
			})
		} else {
			sort.SliceStable(Events, func(i, j int) bool {
				return Events[i].GetBaseEvent().EventModelId > Events[j].GetBaseEvent().EventModelId
			})
		}

	default:
		logger.Debugf("暂时不支持其他字段排序%s", sortKey)
	}
	start := int64(0)
	if limit == 0 {
		limit = 10
	}

	// if offset != 0 {
	// 	start = offset
	// }
	end := int64(len(Events))
	//NOTE 防止误传
	if offset < 0 || offset > end {
		offset = 0
	}
	if offset > 0 && offset < end {
		start = offset
	}

	//NOTE -1 返回全部
	if limit < 0 {
		limit = end
	}

	if (offset + limit) > int64(len(Events)) {
		end = int64(len(Events))
	} else {
		end = offset + limit
	}
	Events = Events[start:end]
	return Events, nil
}

// 通过数据视图查询事件数据
// @param ctx
// @param query
// @param em
// @param timeLimit  数据视图查询时间范围限制
// @param eventModelIdLimit 事件模型id限制
// @return interfaces.IEvents
// @return error
func (eventEngine *EventEngine) PersistQuery(ctx context.Context, query interfaces.EventQuery, em interfaces.EventModel, timeLimit, eventModelIdLimit bool) (interfaces.IEvents, error) {
	//NOTE: 数据视图查询时间范围限制，无限制则可以查询全部
	if timeLimit {
		//NOTE 如果start没传，按照时间窗口补齐。
		if query.End == 0 {
			query.End = time.Now().UnixNano() / 1e6
		}

		if query.Start == 0 {
			strT := strconv.Itoa(em.DefaultTimeWindow.Interval) + em.DefaultTimeWindow.Unit
			stepT, _ := convert.ParseDuration(strT)
			Step := convert.DurationMilliseconds(stepT)
			query.Start = query.End - Step
		}
		if query.End == query.Start && query.End != -1 {
			strT := strconv.Itoa(em.DefaultTimeWindow.Interval) + em.DefaultTimeWindow.Unit
			stepT, _ := convert.ParseDuration(strT)
			Step := convert.DurationMilliseconds(stepT)
			query.Start = query.End - Step
		}
		//NOTE 批量按ID查询事件数据详情，需要查找整个索引库里面的某一条记录，所有时间要改为空，此时前端模型传入start = -1,end=-1,
		if query.End == -1 && query.Start == -1 {
			query.End = time.Now().UnixNano() / 1e6
			query.Start = query.End - 31536000000 //NOTE: 最大查两年
		}

	}
	//NOTE 直接过滤掉没有事件内容的事件数据。
	if !query.EnableMessageFilter {
		query.Extraction = append(query.Extraction, interfaces.Filter{Name: "event_message", Operation: "!=", Value: ""})
	}

	model := em

	//NOTE: 复用data_view的查询
	model.DataSourceType = "data_view"

	//智能聚合不含数据源
	if len(model.DataSource) == 0 {
		model.DataSource = append(model.DataSource, model.Task.StorageConfig.DataViewId)
	} else {
		model.DataSource[0] = model.Task.StorageConfig.DataViewId
	}

	// NOTE: 事件模型id限制，单个事件模型
	if eventModelIdLimit {
		extraction := interfaces.Filter{
			Name:      "event_model_id",
			Operation: "=",
			Value:     model.EventModelID}
		query.Extraction = append(query.Extraction, extraction)
	}

	dmq := GenerateDataModelQuery(eventEngine.appSetting, model, query)

	//查出来的是原始数据。
	var format = "flat"
	sr, _, err := dmq.FetchSourceRecordsFrom(ctx, format)
	if err != nil {
		logger.Errorf("source records will be null, because source data query function return error,error info is: %v", err)
		return interfaces.IEvents{}, err

	}
	var events interfaces.IEvents
	for _, r := range sr.Records {
		var event interfaces.EventData
		eventType := r["event_type"].(string)

		// 把@timestamp转成int64
		if time_str, ok := r["@timestamp"].(string); ok {
			to, _ := time.Parse("2006-01-02T15:04:05Z", time_str)
			r["@timestamp"] = to.UnixMilli()
		}

		bytes, err := json.Marshal(r)
		if err != nil {
			logger.Errorf("Marshal records error: %v", err)
			return interfaces.IEvents{}, err

		}

		err = json.Unmarshal(bytes, &event)
		if err != nil {
			logger.Errorf("UnMarshal record bytes to  event error: %v", err)
			return interfaces.IEvents{}, err
		}
		if timestamp, ok := r["@timestamp"].(int64); ok {
			event.CreateTime = timestamp
		} else {
			if time_str, ok := r["@timestamp"].(string); ok {
				to, _ := time.Parse("2006-01-02T15:04:05Z", time_str)
				event.CreateTime = to.UnixMilli()
			}
		}

		defaultTimeWindow, err := event.SetDefaultTimeWindow()
		if err != nil {
			logger.Errorf("SetDefaultTimeWindow error: %v", err)
			return interfaces.IEvents{}, err
		}
		event.DefaultTimeWindow = defaultTimeWindow
		labels, err := event.SetLabels()
		if err != nil {
			logger.Errorf("SetLabels error: %v", err)
			return interfaces.IEvents{}, err
		}
		event.Labels = labels

		relations, err := event.SetRelations()
		if err != nil {
			logger.Errorf("SetRelations error: %v", err)
			return interfaces.IEvents{}, err
		}
		event.BaseEvent.Relations = relations

		triggerData, err := event.SetTriggerData()
		if err != nil {
			logger.Errorf("SetTriggerData error: %v", err)
			return interfaces.IEvents{}, err
		}

		event.TriggerData = triggerData

		schedule, err := event.SetSchedule()
		if err != nil {
			logger.Errorf("SetSchedule error: %v", err)
			return interfaces.IEvents{}, err
		}
		event.Schedule = schedule

		event.EventType = eventType
		//NOTE 断言
		context, err := event.SetContext()
		if err != nil {
			logger.Errorf("SetContext error: %v", err)
			return interfaces.IEvents{}, err
		}
		event.Context = context

		var eventResp interfaces.EventRespData
		//NOTE:单个事件模型
		if eventModelIdLimit {
			eventResp = interfaces.EventRespData{
				BaseEvent:     event.BaseEvent,
				Message:       GetMessage(event),
				Context:       event.Context,
				TriggerTime:   event.TriggerTime,
				TriggerData:   event.TriggerData,
				AggregateType: event.AggregateType,
				AggregateAlgo: event.AggregateAlgo,
				DetectType:    event.DetectType,
				DetectAlgo:    event.DetectAlgo,
				Relations:     event.BaseEvent.Relations,
			}
		} else {
			eventResp = interfaces.EventRespData{
				BaseEvent:     event.BaseEvent,
				Message:       GetMessage(event),
				TriggerTime:   event.TriggerTime,
				AggregateType: event.AggregateType,
				AggregateAlgo: event.AggregateAlgo,
				DetectType:    event.DetectType,
				DetectAlgo:    event.DetectAlgo,
			}
		}

		events = append(events, eventResp)
	}
	return events, nil
}

// FIXME: 此处参数DataModelQuery是父类，传入的是子类。实现一个动态获取。
func (rd *EventEngine) Query(ctx context.Context, query interfaces.EventQuery, em interfaces.EventModel) (interfaces.SourceRecords, interfaces.Record, error) {
	//NOTE: 提取源数据
	//NOTE 如果start没传，按照时间窗口补齐。
	if query.End == 0 {
		query.End = time.Now().UnixNano() / 1e6
	}

	if query.Start == 0 {
		strT := strconv.Itoa(em.DefaultTimeWindow.Interval) + em.DefaultTimeWindow.Unit
		stepT, _ := convert.ParseDuration(strT)
		Step := convert.DurationMilliseconds(stepT)
		query.Start = query.End - Step
	}
	if query.End == query.Start {
		strT := strconv.Itoa(em.DefaultTimeWindow.Interval) + em.DefaultTimeWindow.Unit
		stepT, _ := convert.ParseDuration(strT)
		Step := convert.DurationMilliseconds(stepT)
		query.Start = query.End - Step
	}

	//NOTE: 根据数据源类型产生对应的数据访问器

	dmq := GenerateDataModelQuery(rd.appSetting, em, query)

	//Fetch source Records from data source
	var format = "flat"
	sr, model, err := dmq.FetchSourceRecordsFrom(ctx, format)
	if err != nil {
		logger.Errorf("source records will be Null, because  source data query function return error,error info is: %v", err)
		return interfaces.SourceRecords{}, nil, err
	}
	return sr, model, nil
}

func (rd *EventEngine) Process(ctx context.Context, sr interfaces.SourceRecords, em interfaces.EventModel) (interfaces.SourceRecords, error) {
	//NOTE：中间处理过程，暂为空，留个口子。
	return sr, nil
}

func (rd *EventEngine) Judge(ctx context.Context, sr interfaces.SourceRecords, em interfaces.EventModel) (interfaces.IEvents, int, error) {
	var events []interfaces.IEvent
	//NOTE：对结果进行判定，判定是否满足触发条件
	if em.EventModelType == "atomic" {

		formula := em.DetectRule.Formula
		sort.Sort(interfaces.Formula(formula))
		for _, r := range sr.Records {
			lastEventLevel := interfaces.EVENT_MODEL_LEVEL_NORMAL
			result, formulaItem, element := rd.Call(r, em, formula)
			//NOTE: hit,
			if result {
				// 清除的事件的逻辑：如果前序事件是紧急，如当前事件是正常，则产生一个等级为清除的事件
				labels := make(map[string]string)
				for k, v := range r {
					if strings.HasPrefix(k, "labels.") {
						labels[k] = v.(string)
					}
				}
				key := common.GenerateUniqueKey(em.EventModelID, labels)
				if formulaItem.Level == interfaces.EVENT_MODEL_LEVEL_CLEARED {
					if event, ok := PreOrderEvent.Load(key); ok {
						if event.(interfaces.PreOrderEvent).Level == interfaces.EVENT_MODEL_LEVEL_NORMAL ||
							event.(interfaces.PreOrderEvent).Level == interfaces.EVENT_MODEL_LEVEL_CLEARED {

							PreOrderEvent.Store(key, interfaces.PreOrderEvent{
								Level: interfaces.EVENT_MODEL_LEVEL_NORMAL,
							})
							continue
						} else {
							//如果有前序级则赋值，如果没有则获取前序等级
							lastEventLevel = event.(interfaces.PreOrderEvent).Level
						}
					} else {
						lastEventLevel = rd.GetLastEventLevel(ctx, key, em)
						if lastEventLevel == interfaces.EVENT_MODEL_LEVEL_NORMAL || lastEventLevel == interfaces.EVENT_MODEL_LEVEL_CLEARED {
							PreOrderEvent.Store(key, interfaces.PreOrderEvent{
								Level: interfaces.EVENT_MODEL_LEVEL_NORMAL,
							})
							continue
						}
					}
				}
				event, _ := rd.GenerationAtomicEvent(ctx, sr, em, formulaItem, element, r)
				if event.Level == interfaces.EVENT_MODEL_LEVEL_CLEARED {
					event.Context.PreOrderEvent.Level = lastEventLevel
					if preOrderEvent, ok := PreOrderEvent.Load(key); ok {
						event.Context.PreOrderEvent.Id = preOrderEvent.(interfaces.PreOrderEvent).Id
						logger.Debugf("当前清除事件id为:%d,前序事件id为:%v", event.Id, event.Context.PreOrderEvent.Id)
					}
				}
				events = append(events, event)
				PreOrderEvent.Store(key, interfaces.PreOrderEvent{
					Level: event.Level,
					Id:    event.Id,
				})
			}
		}
		return events, len(events), nil

	} else if em.EventModelType == "aggregate" && em.AggregateRule.Type == "healthy_compute" {
		if em.AggregateRule.AggregateAlgo == "MaxLevelMap" {
			hits, level, score := operatation.MaxLevelMap(sr.Records)
			if len(hits) == 0 {
				return events, len(events), nil
			}
			context := Combine(hits, level, score)
			event, _ := rd.GenerationAggregateEvent(ctx, sr, em, context, []string{})
			events = append(events, event)
			return events, len(events), nil
		}

	} else if em.EventModelType == "aggregate" && em.AggregateRule.Type == "group_aggregation" {
		var groupRecords map[string]interfaces.Records

		if em.AggregateRule.AggregateAlgo == "EventDataGroupAggregation" {
			groupRecords = operatation.EventDataGroupAggregation(sr.Records, em.AggregateRule.GroupFields)

		} else if em.AggregateRule.AggregateAlgo == "SourceDataGroupAggregation" {
			groupRecords = operatation.SourceDataGroupAggregation(sr.Records, em.AggregateRule.GroupFields)
		}
		if len(groupRecords) == 0 {
			return events, len(events), nil
		}

		for group_field, hits := range groupRecords {
			group_fields := strings.Split(group_field, ",")
			context := GroupCombine(hits, em.AggregateRule.GroupFields)
			event, _ := rd.GenerationAggregateEvent(ctx, sr, em, context, group_fields)
			events = append(events, event)
		}

	}

	return events, len(events), nil

}

func (rd *EventEngine) Call(record map[string]any, em interfaces.EventModel, f []interfaces.FormulaItem) (hit bool, hitFormulaItem interfaces.FormulaItem, CompareElement map[string]any) {

	for _, formulaItem := range f {
		filter := formulaItem.Filter
		//执行操作符，得到是否命中结果
		// var fieldValue any = record[formulaItem.Filter.FilterExpress.Name]
		var aggreFieldMap map[string]any
		hit, aggreFieldMap = rd.Traversal(filter, record)

		if hit {
			hitFormulaItem = formulaItem   //返回命中的等级规则
			CompareElement = aggreFieldMap //返回指定字段的字段值
			break
		}
	}
	return
}

func (rd *EventEngine) Traversal(filter interfaces.LogicFilter, record map[string]any) (bool, map[string]any) {
	if len(filter.Children) > 0 && (filter.LogicOperator == "and" || filter.LogicOperator == "or") {

		if filter.LogicOperator == "and" {
			var b = true
			var aggreFieldMap = make(map[string]any, len(filter.Children))
			for _, filter := range filter.Children {
				a, fieldMap := rd.Traversal(filter, record)
				b = b && a
				for key, value := range fieldMap {
					aggreFieldMap[key] = value
				}
			}

			return b, aggreFieldMap
		} else {

			var b = false
			var aggreFieldMap = make(map[string]any, len(filter.Children))
			for _, filter := range filter.Children {
				a, fieldMap := rd.Traversal(filter, record)
				b = b || a
				for key, value := range fieldMap {
					aggreFieldMap[key] = value
				}
			}
			return b, aggreFieldMap
		}

	} else {
		var fieldMap = make(map[string]any, 1)
		if value, ok := record[filter.FilterExpress.Name]; ok {
			fieldMap[filter.FilterExpress.Name] = value
			return operatation.Exec(filter.FilterExpress.Operation, value, filter.FilterExpress.Value), fieldMap
		} else {
			return false, fieldMap
		}

	}
}

func (rd *EventEngine) GetLastEventLevel(ctx context.Context, key string, em interfaces.EventModel) int {
	query := interfaces.EventQuery{
		Id:        em.EventModelID,
		Direction: "desc",
	}
	//执行频率
	shcedule, _ := convert.ParseDuration(em.Task.Schedule.Expression)
	shceduleStep := convert.DurationMilliseconds(shcedule)

	query.End = time.Now().UnixNano() / 1e6
	query.Start = query.End - 2*shceduleStep
	//前序等级
	lastEventLevel := interfaces.EVENT_MODEL_LEVEL_NORMAL
	lastEvents, _ := rd.PersistQuery(ctx, query, em, true, true)
	if len(lastEvents) == 0 {
		return lastEventLevel
	}
	//查询时默认升序；倒序遍历查找前序等级
	for i := len(lastEvents) - 1; i >= 0; i-- {
		key1 := common.GenerateUniqueKey(em.EventModelID, lastEvents[i].GetBaseEvent().Labels)
		if key1 == key {
			lastEventLevel = lastEvents[i].GetLevel()
			break
		}
	}
	return lastEventLevel
}

func (rd *EventEngine) GenerationAtomicEvent(ctx context.Context, sr interfaces.SourceRecords, em interfaces.EventModel,
	f interfaces.FormulaItem, element map[string]any, record map[string]any) (interfaces.AtomicEvent, error) {

	id := xid.New().String()

	message := Compose(ctx, em, f, element, record) //基于已有的信息构造事件消息内容

	// recordStr, _ := json.Marshal(record)
	var trigger_time int64
	if timestamp, ok := record["@timestamp"].(float64); ok {
		trigger_time = int64(timestamp)
	} else if timestamp, ok := record["@timestamp"].(int64); ok {
		trigger_time = timestamp
	} else {
		if time_str, ok := record["@timestamp"].(string); ok {
			to, _ := time.Parse("2006-01-02T15:04:05Z", time_str)
			// to, _ := time.Parse("2006-01-02T15:04:05+08:00", time_str)
			trigger_time = to.UnixMilli()
		}
	}
	language := rest.GetLanguageByCtx(ctx)
	var title string
	if language == "zh-CN" {
		title = em.EventModelName + "_" + interfaces.EVENT_MODEL_LEVEL_ZH_CN[f.Level]
	} else {
		title = em.EventModelName + "_" + interfaces.EVENT_MODEL_LEVEL_EN_US[f.Level]
	}

	labels := make(map[string]string)
	for key, value := range record {
		if strings.HasPrefix(key, "labels.") {
			labels[key] = value.(string)
		}
	}

	event := interfaces.AtomicEvent{
		BaseEvent: interfaces.BaseEvent{
			Id:                  id,
			Title:               title,
			EventModelId:        em.EventModelID,
			EventModelName:      em.EventModelName,
			Level:               f.Level,
			LevelName:           interfaces.EVENT_MODEL_LEVEL_EN_US[f.Level],
			GenerateType:        ComputeGenerateType(em),
			Tags:                em.EventModelTags,
			CreateTime:          time.Now().UnixMilli(),
			EventType:           em.EventModelType,
			DataSource:          em.DataSource,
			DataSourceType:      em.DataSourceType,
			DataSourceName:      em.DataSourceName,
			DataSourceGroupName: em.DataSourceGroupName,
			DefaultTimeWindow:   em.DefaultTimeWindow,
			Schedule:            em.Task.Schedule,
			Labels:              labels,
		},
		DetectAlgo:  em.DetectRule.Type, //range_detect | status_detect
		DetectType:  "rule_detect",      //agi_detct | rule_detect
		TriggerTime: trigger_time,
		TriggerData: interfaces.Records{record},
		Message:     message,
	}
	return event, nil
}

func (rd *EventEngine) GenerationAggregateEvent(ctx context.Context, sr interfaces.SourceRecords, em interfaces.EventModel,
	context interfaces.EventContext, additionalTags []string) (interfaces.AggregateEvent, error) {

	id := xid.New().String()

	timestamp, _ := sr.Records[0]["@timestamp"].(int64)
	language := rest.GetLanguageByCtx(ctx)
	var title string
	if language == "zh-CN" {
		title = em.EventModelName + "_" + interfaces.EVENT_MODEL_LEVEL_ZH_CN[context.Level]
		// level = EVENT_MODEL_LEVEL_ZH_CN[context.Level]
	} else {
		title = em.EventModelName + "_" + interfaces.EVENT_MODEL_LEVEL_EN_US[context.Level]
		// level = EVENT_MODEL_LEVEL_EN_US[context.Level]
	}
	//TODO: 根据分组字段进行分组。
	additionalTags = append(additionalTags, em.EventModelTags...)
	event := interfaces.AggregateEvent{
		BaseEvent: interfaces.BaseEvent{
			Id:                  id,
			Title:               title,
			EventModelId:        em.EventModelID,
			EventModelName:      em.EventModelName,
			Level:               context.Level,
			LevelName:           interfaces.EVENT_MODEL_LEVEL_EN_US[context.Level],
			Tags:                additionalTags,
			GenerateType:        ComputeGenerateType(em),
			CreateTime:          time.Now().UnixMilli(),
			EventType:           em.EventModelType,
			DataSource:          em.DataSource,
			DataSourceType:      em.DataSourceType,
			DataSourceName:      em.DataSourceName,
			DataSourceGroupName: em.DataSourceGroupName,
			DefaultTimeWindow:   em.DefaultTimeWindow,
			Schedule:            em.Task.Schedule,
		},
		AggregateType: em.AggregateRule.Type,
		AggregateAlgo: em.AggregateRule.AggregateAlgo,
		TriggerTime:   timestamp,
		TriggerData:   interfaces.Records{},
		Context:       context,
		Message:       GenerateMessage(em.AggregateRule.Type, context, em.AggregateRule.GroupFields, interfaces.EVENT_MODEL_LEVEL_ZH_CN[context.Level]),
	}

	return event, nil
}

func GetMessage(event interfaces.EventData) string {
	if event.EventMessage != "" {
		return event.EventMessage
	}
	return event.Message
}

func ComputeGenerateType(em interfaces.EventModel) string {
	// if em.EnableSubscribe == 0 && em.IsActive == 1 {
	// 	return "batch"
	// }
	if em.EnableSubscribe == 1 && em.IsActive == 0 {
		return interfaces.GENERATE_TYPE_STREAMING
	}
	return interfaces.GENERATE_TYPE_BATCH

}

func (rd *EventEngine) QuerySingleEventByEventId(ctx context.Context, em interfaces.EventModel,
	query interfaces.EventDetailsQueryReq) (interfaces.IEvent, error) {
	//NOTE: 复用data_view的查询
	em.DataSourceType = "data_view"
	if len(em.DataSource) == 0 {
		em.DataSource = append(em.DataSource, em.Task.StorageConfig.DataViewId)
	} else {
		em.DataSource[0] = em.Task.StorageConfig.DataViewId
	}

	dmq := GenerateDataModelQuery(rd.appSetting, em,
		interfaces.EventQuery{
			Start: query.Start,
			End:   query.End,
			Extraction: []interfaces.Filter{
				{
					Name:      "id",
					Operation: "=",
					Value:     query.EventID,
				},
				{
					Name:      "event_model_id",
					Operation: "=",
					Value:     query.EventModelID,
				},
			},
		},
	)

	//NOTE 数据转为IEvent
	var event interfaces.IEvent
	format := "flat"
	sr, _, err := dmq.FetchSourceRecordsFrom(ctx, format)
	if err != nil {
		logger.Errorf("source records will be Null, because  source data query function return error,error info is: %v", err)
		return event, err

	}
	if len(sr.Records) == 0 {
		//NOTE： 针对问题事件，因为索引库不同，需要再单独查一遍。
		logger.Errorf("get event by id failed, event not found")
		return nil, errors.New(uerrors.Uniquery_EventModel_EventNotFound)
	}

	r := sr.Records[0]
	var eventData interfaces.EventData
	eventType := r["event_type"].(string)

	bytes, err := json.Marshal(r)
	if err != nil {
		logger.Errorf("Marshal records error: %v", err)
		return event, err
	}

	err = json.Unmarshal(bytes, &eventData)
	if err != nil {
		logger.Errorf("UnMarshal record bytes to event error: %v", err)
		return event, err
	}

	if timestamp, ok := r["@timestamp"].(int64); ok {
		eventData.CreateTime = timestamp
	} else {
		if time_str, ok := r["@timestamp"].(string); ok {
			to, _ := time.Parse("2006-01-02T15:04:05Z", time_str)
			eventData.CreateTime = to.UnixMilli()
		}
	}
	defaultTimeWindow, err := eventData.SetDefaultTimeWindow()
	if err != nil {
		logger.Errorf("SetDefaultTimeWindow error: %v", err)
		return event, err
	}
	eventData.DefaultTimeWindow = defaultTimeWindow

	labels, err := eventData.SetLabels()
	if err != nil {
		logger.Errorf("SetLabels error: %v", err)
		return event, err
	}
	eventData.Labels = labels

	relations, err := eventData.SetRelations()
	if err != nil {
		logger.Errorf("SetRelations error: %v", err)
		return event, err
	}
	eventData.BaseEvent.Relations = relations
	triggerData, err := eventData.SetTriggerData()
	if err != nil {
		logger.Errorf("SetTriggerData error: %v", err)
		return event, err
	}
	eventData.TriggerData = triggerData

	schedule, err := eventData.SetSchedule()
	if err != nil {
		logger.Errorf("SetSchedule error: %v", err)
		return event, err
	}
	eventData.Schedule = schedule

	//NOTE 放在开头会被覆盖
	eventData.EventType = eventType
	context, err := eventData.SetContext()
	if err != nil {
		logger.Errorf("SetContext error: %v", err)
		return event, err
	}
	eventData.Context = context

	eventResp := interfaces.EventRespData{
		BaseEvent:     eventData.BaseEvent,
		Message:       GetMessage(eventData),
		Context:       eventData.Context,
		TriggerTime:   eventData.TriggerTime,
		TriggerData:   eventData.TriggerData,
		AggregateType: eventData.AggregateType,
		AggregateAlgo: eventData.AggregateAlgo,
		DetectType:    eventData.DetectType,
		DetectAlgo:    eventData.DetectAlgo,
		Relations:     eventData.BaseEvent.Relations,
	}

	return eventResp, nil
}
