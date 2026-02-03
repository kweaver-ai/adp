import Request from '../request';
import * as ActionType from './type';

const BASE_URL = '/api/ontology-manager/v1/knowledge-networks';
const BASE_URL_QUERY = '/api/ontology-query/v1/knowledge-networks';

/**
 * 获取行动类列表
 * @param knId 知识网络ID
 * @param params 查询参数
 */
export const getActionTypes = (knId: string, params: ActionType.GetActionTypesRequest): Promise<ActionType.GetActionTypesResponse> => {
  return Request.get(`${BASE_URL}/${knId}/action-types`, params);
};

/**
 * 删除行动类
 * @param knId 知识网络ID
 * @param atIds 行动类ID列表
 */
export const deleteActionType = (knId: string, atIds: string[]): Promise<void> => {
  return Request.delete(`${BASE_URL}/${knId}/action-types/${atIds.join(',')}`);
};

/**
 * 新建行动类
 * @param knId 知识网络ID
 * @param data 创建数据
 */
export const createActionType = (knId: string, data: ActionType.CreateActionTypeRequest): Promise<void> => {
  return Request.post(`${BASE_URL}/${knId}/action-types`, { entries: data });
};

/**
 * 编辑行动类
 * @param knId 知识网络ID
 * @param atId 行动类ID
 * @param data 编辑数据
 */
export const editActionType = (knId: string, atId: string, data: ActionType.EditActionTypeRequest): Promise<void> => {
  return Request.put(`${BASE_URL}/${knId}/action-types/${atId}`, data);
};

/**
 * 获取行动类详情
 * @param knId 知识网络ID
 * @param atIds 行动类ID列表
 */
export const getActionTypeDetail = (knId: string, atIds: string[]): Promise<ActionType.ActionType[]> => {
  return Request.get<{ entries: ActionType.ActionType[] }>(`${BASE_URL}/${knId}/action-types/${atIds.join(',')}`).then((response) => response.entries);
};

// /**
//  * 创建行动计划
//  * @param knId 知识网络ID
//  * @param data 创建数据
//  */
// export const createActionSchedule = (knId: string, data: ActionType.CreateActionScheduleRequest): Promise<{ id: string }> => {
//   return Request.post(`${BASE_URL}/${knId}/action-schedules`, data);
// };

// /**
//  * 删除行动计划
//  * @param knId 知识网络ID
//  * @param scheduleIds 行动计划ID列表
//  */
// export const deleteActionSchedules = (knId: string, scheduleIds: string[]): Promise<void> => {
//   return Request.delete(`${BASE_URL}/${knId}/action-schedules/${scheduleIds.join(',')}`);
// };

// /**
//  * 更新行动计划
//  * @param knId 知识网络ID
//  * @param scheduleId 行动计划ID
//  * @param data 更新数据
//  */
// export const updateActionSchedule = (knId: string, scheduleId: string, data: ActionType.UpdateActionScheduleRequest): Promise<void> => {
//   return Request.put(`${BASE_URL}/${knId}/action-schedules/${scheduleId}`, data);
// };

// /**
//  * 更新行动计划状态
//  * @param knId 知识网络ID
//  * @param scheduleId 行动计划ID
//  * @param status 状态
//  */
// export const updateActionScheduleStatus = (knId: string, scheduleId: string, status: ActionType.ActionScheduleStatusEnum): Promise<void> => {
//   return Request.put(`${BASE_URL}/${knId}/action-schedules/${scheduleId}/status`, { status });
// };

/**
 * 获取行动计划列表
 * @param knId 知识网络ID
 * @param params 查询参数
 */
export const getActionSchedules = (knId: string, params: ActionType.GetActionSchedulesRequest): Promise<ActionType.GetActionSchedulesResponse> => {
  return Request.get(`${BASE_URL}/${knId}/action-schedules`, params);
};

/**
 * 获取行动计划详情
 * @param knId 知识网络ID
 * @param scheduleId 行动计划ID
 */
export const getActionSchedule = (knId: string, scheduleId: string): Promise<ActionType.ActionSchedule> => {
  return Request.get(`${BASE_URL}/${knId}/action-schedules/${scheduleId}`);
};

/**
 * 执行行动类
 * @param knId 知识网络ID
 * @param atId 行动类ID
 * @param data 执行参数
 */
export const executeActionType = (knId: string, atId: string, data: ActionType.ActionExecutionRequest): Promise<ActionType.ActionExecutionResponse> => {
  return Request.post(`${BASE_URL_QUERY}/${knId}/action-types/${atId}/execute`, data);
};

/**
 * 获取行动执行状态
 * @param knId 知识网络ID
 * @param executionId 执行ID
 */
export const getActionExecutionStatus = (knId: string, executionId: string): Promise<ActionType.ActionExecution> => {
  return Request.get(`${BASE_URL_QUERY}/${knId}/action-executions/${executionId}`);
};

/**
 * 查询行动日志
 * @param knId 知识网络ID
 * @param params 查询参数
 */
export const queryActionLogs = (knId: string, params: ActionType.QueryActionLogsRequest): Promise<ActionType.ActionExecutionList> => {
  return Request.get(`${BASE_URL_QUERY}/${knId}/action-logs`, params);
};

/**
 * 获取行动执行日志详情
 * @param knId 知识网络ID
 * @param logId 日志ID
 */
export const getActionExecutionLogDetail = (knId: string, logId: string): Promise<ActionType.ActionExecutionLogDetail> => {
  return Request.get(`${BASE_URL_QUERY}/${knId}/action-logs/${logId}`);
};

/**
 * 取消行动执行
 * @param knId 知识网络ID
 * @param logId 日志ID (执行ID)
 */
export const cancelActionExecution = (knId: string, logId: string): Promise<ActionType.CancelExecutionResponse> => {
  return Request.post(`${BASE_URL_QUERY}/${knId}/action-logs/${logId}/cancel`);
};

export default {
  getActionTypes,
  deleteActionType,
  createActionType,
  editActionType,
  getActionTypeDetail,
  // createActionSchedule,
  // deleteActionSchedules,
  // updateActionSchedule,
  // updateActionScheduleStatus,
  getActionSchedules,
  getActionSchedule,
  executeActionType,
  getActionExecutionStatus,
  queryActionLogs,
  getActionExecutionLogDetail,
  cancelActionExecution,
};
