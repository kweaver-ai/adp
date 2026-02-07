// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package event

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"uniquery/common"
	uerrors "uniquery/errors"
	"uniquery/interfaces"
	"uniquery/logics"
	"uniquery/logics/permission"
)

var (
	eventOnce sync.Once
	es        interfaces.EventService
)

type eventService struct {
	appSetting *common.AppSetting
	engine     interfaces.EventEngine
	emAccess   interfaces.EventModelAccess
	dvAccess   interfaces.DataViewAccess
	// uAccess    interfaces.UniqueryAccess
	ibAccess interfaces.IndexBaseAccess
	ps       interfaces.PermissionService
}

func NewEventService(appSetting *common.AppSetting) interfaces.EventService {
	eventOnce.Do(func() {
		es = &eventService{
			appSetting: appSetting,
			engine:     nil,
			emAccess:   logics.EMAccess,
			dvAccess:   logics.DVAccess,
			// uAccess:    logics.UAccess,
			ibAccess: logics.IBAccess,
			ps:       permission.NewPermissionService(appSetting),
		}
	})
	return es
}

func ComposeEventModelByQueryParam(query interfaces.EventQuery) interfaces.EventModel {
	return interfaces.EventModel{
		EventModelID:      "0",
		EventModelName:    query.EventModelName,
		EventModelType:    query.EventModelType,
		UpdateTime:        time.Now().UnixMilli(),
		EventModelTags:    query.EventModelTags,
		DataSourceType:    query.DataSourceType,
		DataSource:        query.DataSource,
		DetectRule:        query.DetectRule,
		AggregateRule:     query.AggregateRule,
		Comment:           query.Comment,
		DefaultTimeWindow: query.DefaultTimeWindow,
		IsActive:          0,
		IsCustom:          1,
	}
}

// 事件数据查询接口
func (ds *eventService) Query(ctx context.Context, queryReq interfaces.EventQueryReq) (int, any, []interfaces.Records, error) {
	//NOTE: 初始化处理对象, 目前默认每次查询都是同类型查询，比如都是范围查询,
	//NOTE：所以可以使用querys的第一个对象进行处理对象的初始化

	// 决策权限。 预览的时候还没有模型id，此时的预览校验用新建或者编辑，
	ops, err := ds.ps.GetResourcesOperations(ctx, interfaces.RESOURCE_TYPE_EVENT_MODEL,
		[]string{interfaces.RESOURCE_ID_ALL})
	if err != nil {
		return 0, []interfaces.IEvents{}, []interfaces.Records{}, err
	}
	if len(ops) != 1 {
		// 无权限
		return 0, []interfaces.IEvents{}, []interfaces.Records{}, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: insufficient permissions for event model's create or modify operation.")
	}
	// 从 ops 里找新建或编辑的权限
	for _, op := range ops[0].Operations {
		if op != interfaces.OPERATION_TYPE_CREATE && op != interfaces.OPERATION_TYPE_DELETE {
			return 0, []interfaces.IEvents{}, []interfaces.Records{}, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails("Access denied: insufficient permissions for event model's create or modify operation.")
		}
	}

	//NOTE: 构建查询器
	var compose_event_model interfaces.EventModel
	querys := queryReq.Querys
	if len(querys) == 1 && querys[0].Preview == 1 && querys[0].Id == "" { //新建预览查询
		compose_event_model = ComposeEventModelByQueryParam(querys[0])
	} else {
		if len(querys) == 0 {
			return 0, []interfaces.IEvents{}, []interfaces.Records{}, nil
		}
		compose_event_models, err := ds.emAccess.GetEventModelById(ctx, querys[0].Id)

		if len(compose_event_models) == 0 {
			return 0, nil, nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				uerrors.Uniquery_EventModel_InternalError_GetEventModelByIdFailed).WithErrorDetails(err.Error())
		} else {
			compose_event_model = compose_event_models[0]
		}

		//普通查询
	}
	err = ds.InitEngine(ctx, compose_event_model)
	if err != nil {
		return 0, nil, nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_InternalError)
	}

	var result []interfaces.IEvents
	var entities []interfaces.Records
	var cnt = 0
	// Preview=1 是预览
	if len(querys) == 1 && querys[0].Preview == 1 {
		//NOTE： 如果是数据预览，则获取到一个事件数据即可返回
		events, entity, _ := ds.engine.Apply(ctx, querys[0], compose_event_model)
		if len(events) == 0 {
			result = append(result, interfaces.IEvents{})
			entities = append(entities, interfaces.Records{})
			cnt = 0
		} else {
			result = append(result, events[:1])
			if len(entity) != 0 {
				entities = append(entities, entity[:1])
			}
			cnt = cnt + len(events[:1])
		}
		return cnt, result, entities, nil

	} else {
		//NOTE: 开启循环遍历查询
		for _, query := range querys {
			event_models, err1 := ds.emAccess.GetEventModelById(ctx, query.Id)

			if err1 != nil || len(event_models) == 0 {
				result = append(result, interfaces.IEvents{})
				continue
			}
			events, entity, query_total := ds.engine.Apply(ctx, query, event_models[0])

			if len(events) == 0 {
				result = append(result, interfaces.IEvents{})
				entities = append(entities, interfaces.Records{})
			} else {
				result = append(result, events)
				entities = append(entities, entity)
			}
			cnt = cnt + query_total

		}
		// //NOTE: 再做一次外部分页，全局请求参数中的分页
		if queryReq.Limit != 0 || queryReq.Offset != 0 {
			var flatteResult interfaces.IEvents = interfaces.IEvents{}
			for _, event := range result {
				flatteResult = append(flatteResult, event...)
			}

			flatteResult, _ = ds.engine.Assemble(flatteResult, queryReq.SortKey, queryReq.Direction, queryReq.Limit, queryReq.Offset)
			cnt = len(flatteResult)
			return cnt, flatteResult, entities, nil
		}
	}

	return cnt, result, entities, nil
}

// 根据ID 查询事件详情
func (ds *eventService) QuerySingleEventByEventId(ctx context.Context, query interfaces.EventDetailsQueryReq) (interfaces.IEvent, error) {
	// 决策当前模型id的数据查询权限
	err := ds.ps.CheckPermission(ctx, interfaces.Resource{
		ID:   query.EventModelID,
		Type: interfaces.RESOURCE_TYPE_EVENT_MODEL,
	}, []string{interfaces.OPERATION_TYPE_DATA_QUERY})
	if err != nil {
		return nil, err
	}

	//NOTE  查事件模型
	eventModels, err := ds.emAccess.GetEventModelById(ctx, query.EventModelID)
	if err != nil {
		logger.Errorf("QueryEventById failed ,because GetEventModelById function return error,error info is: %v", err)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			uerrors.Uniquery_EventModel_InternalError_GetEventModelByIdFailed).WithErrorDetails(err.Error())
	}
	if len(eventModels) == 0 {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, uerrors.Uniquery_EventModel_EventModelNotFound)
	}
	model := eventModels[0]
	_ = ds.InitEngine(ctx, model)
	var events interfaces.IEvent
	events, err = ds.engine.QuerySingleEventByEventId(ctx, model, query)
	// if err != nil && err.Error() == uerrors.Uniquery_EventModel_EventNotFound {
	// 	//NOTE： 从默认问题库里面再查一次。
	// 	model.Task.StorageConfig.DataViewId = interfaces.INCIDENT_DATA_VIEW_ID
	// 	events, err = ds.engine.QuerySingleEventByEventId(ctx, model, query)
	// 	if err != nil {
	// 		logger.Errorf("QuerySingleEventByEventId failed,error info is: %v", err)
	// 		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_EventModel_InternalError)
	// 	}
	// } else
	if err != nil {
		logger.Errorf("QuerySingleEventByEventId failed,error info is: %v", err)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, uerrors.Uniquery_EventModel_InternalError)
	}
	return events, nil

}
