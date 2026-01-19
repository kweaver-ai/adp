import request from '../request';
import * as ScanManagement from './type';

const BASE_URL = '/api/data-connection/v1/metadata';
const EXCEL_BASE_URL = '/api/data-connection/v1/gateway/excel';

export const createScanTask = (params: ScanManagement.ScanRequest) => {
  const obj = {
    ...params,
    use_default_template: true,
    field_list_when_change: [],
    use_multi_threads: true,
    tables: params.tables || [],
    status: params.status || 'open',
  };
  return request.post<ScanManagement.ScanResponse>(`${BASE_URL}/scan`, obj);
};

export const batchCreateScanTask = (params: ScanManagement.ScanRequest[]) => {
  const obj = params.map((item) => ({
    ...item,
    use_default_template: true,
    field_list_when_change: [],
    use_multi_threads: true,
    tables: item.tables || [],
  }));
  return request.post<ScanManagement.ScanResponse>(`${BASE_URL}/scan/batch`, obj);
};

export const getScanTaskStatus = (taskId: string) => {
  return request.get<ScanManagement.TaskStatusResponse>(`${BASE_URL}/scan/status/${taskId}`);
};

export const getScanTaskTableStatus = (params: { id: string; type: string[] }) => {
  return request.post<ScanManagement.TableStatusResponse>(`${BASE_URL}/scan/status`, params);
};

export const retryScanTask = (params: ScanManagement.RetryRequest) => {
  return request.post<ScanManagement.RetryResponse>(`${BASE_URL}/scan/retry`, params);
};

export const getScanTaskList = (params: ScanManagement.PageQueryParams) => {
  const curSort = params.sort || 'start_time';
  const queryParams = {
    sort: curSort,
    direction: params.direction || 'desc',
    limit: params.limit || -1,
    keyword: params.keyword,
    offset: params.offset || 0,
    ds_id: params.ds_id,
    status: !params.status || params.status === 'all' ? '' : params.status,
    ...params.filters,
  };
  return request.get<ScanManagement.ScanTaskListResponse>(`${BASE_URL}/scan`, queryParams);
};

export const getDataSourceTables = (id: string, params?: ScanManagement.PageQueryParams) => {
  return request.get<ScanManagement.TableListResponse>(`${BASE_URL}/data-source/${id}`, params);
};

export const getScanTaskInfo = (taskId: string, params?: ScanManagement.ScanTaskInfoParams) => {
  return request.get<ScanManagement.ScanTaskInfoResponse>(`${BASE_URL}/scan/info/${taskId}`, params);
};

export const getTableColumns = (tableId: string, params?: ScanManagement.PageQueryParams) => {
  return request.get<ScanManagement.ColumnListResponse>(`${BASE_URL}/table/${tableId}`, params);
};

export const createExcelTable = (params: ScanManagement.CreateExcelTableRequest) => {
  return request.post<ScanManagement.ExcelTableResponse>(`${EXCEL_BASE_URL}/table`, params);
};

export const deleteExcelTable = (tableId: string) => {
  return request.delete<ScanManagement.ExcelTableResponse>(`${EXCEL_BASE_URL}/table/${tableId}`);
};

export const getExcelColumns = (params: ScanManagement.GetExcelColumnsRequest) => {
  return request.post<ScanManagement.ExcelColumnsResponse>(`${EXCEL_BASE_URL}/columns`, params, { timeout: 60000 });
};

export const getExcelSheets = (catalog: string, fileName: string) => {
  return request.get<ScanManagement.ExcelSheetListResponse>(
    `${EXCEL_BASE_URL}/sheet`,
    {
      catalog,
      file_name: fileName,
    },
    { timeout: 60000 }
  );
};

/**
 * 获取Excel文件列表
 * @param catalog - 目录名称
 */
export const getExcelFiles = (catalog: string) => {
  return request.get<ScanManagement.ExcelFileListResponse>(`${EXCEL_BASE_URL}/files/${catalog}`, {}, { timeout: 60000 });
};

/**
 * 获取定时扫描状态
 * @param scheduleId - 定时任务ID
 */
export const getScheduleScanStatus = (scheduleId: string, type: number) => {
  return request.get<ScanManagement.ScheduleScanStatusResponse>(`${BASE_URL}/scan/schedule/${scheduleId}?type=${type}`);
};

/**
 * 获取定时扫描历史记录列表
 * @param scheduleId - 定时任务ID
 * @param params - 分页查询参数
 */
export const getScheduleHistoryList = (scheduleId: string, params?: ScanManagement.PageQueryParams) => {
  return request.get<ScanManagement.ScheduleHistoryListResponse>(`${BASE_URL}/scan/schedule/task/${scheduleId}`, params);
};

/**
 * 更新定时任务状态
 * @param params - 更新状态请求参数
 */
export const updateScheduleStatus = (params: ScanManagement.UpdateScheduleStatusRequest) => {
  return request.put<ScanManagement.UpdateScheduleStatusResponse>(`${BASE_URL}/scan/schedule/status`, params);
};

/**
 * 更新定时任务配置
 * @param params - 更新请求参数
 */
export const updateSchedule = (params: ScanManagement.UpdateScheduleRequest) => {
  return request.put<ScanManagement.UpdateScheduleResponse>(`${BASE_URL}/scan/schedule`, params);
};

export default {
  createScanTask,
  batchCreateScanTask,
  getScanTaskStatus,
  getScanTaskTableStatus,
  retryScanTask,
  getScanTaskList,
  getDataSourceTables,
  getTableColumns,
  createExcelTable,
  deleteExcelTable,
  getExcelColumns,
  getExcelSheets,
  getExcelFiles,
  getScheduleScanStatus,
  getScheduleHistoryList,
  updateScheduleStatus,
  updateSchedule,
};
