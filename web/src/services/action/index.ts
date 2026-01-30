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

/**
 * 获取行动类任务列表
 * @param knId 知识网络ID
 * @param atId 行动类ID
 * @param params 查询参数
 */
export const getActionTasks = (knId: string, atId: string, params: ActionType.GetTasksRequest): Promise<ActionType.GetTasksResponse> => {
  return Request.get(`${BASE_URL}/${knId}/action-types/${atId}/tasks`, params);
};

/**
 * 获取行动类任务详情
 * @param knId 知识网络ID
 * @param atId 行动类ID
 * @param taskId 任务ID
 */
export const getActionTaskDetail = (knId: string, atId: string, taskId: string): Promise<ActionType.Task> => {
  return Request.get(`${BASE_URL}/${knId}/action-types/${atId}/tasks/${taskId}`);
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
  getActionTasks,
  getActionTaskDetail,
  executeActionType,
  getActionExecutionStatus,
  queryActionLogs,
  cancelActionExecution,
};
