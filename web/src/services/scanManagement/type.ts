export type Status = 'wait' | 'running' | 'success' | 'fail' | 'all';

export interface ScanRequest {
  scan_name: string;
  type: number;
  ds_info: { ds_id: string; ds_type: string; scan_strategy?: string[] }; // 扫描策略 insert update delete
  cron_expression?: CronExpression; // cron表达式
  status?: string; // 任务状态'open' | 'close'
  use_default_template?: boolean;
  field_list_when_change?: string[];
  use_multi_threads?: boolean;
  tables?: string[];
}

export interface ScanResponse {
  id: string;
  ds_id: string;
  status: Status;
}

export interface TaskProcessInfo {
  table_count: number;
  success_count: number;
  fail_count: number;
}

export interface TaskResultInfo {
  table_count: number;
  success_count: number;
  fail_count: number;
  fail_stage: number;
  error_stack: string;
}

export interface TaskStatusResponse {
  id: string;
  status: Status;
  data: {
    task_process_info: TaskProcessInfo;
    task_result_info: TaskResultInfo;
  };
}

export interface TableStatus {
  table_id: string;
  table_name: string;
  status: Status;
  start_time: string;
}

export interface TableStatusResponse {
  id: string;
  tables: TableStatus[];
}

export interface RetryRequest {
  id: string;
  tables: string[];
}

export interface RetryResponse {
  id: string;
  tables: Array<{
    table_id: string;
    status: Status;
  }>;
}

export interface ScanTaskItem {
  id: string;
  name: string;
  create_user: string;
  scan_status: Status;
  start_time: string;
  schedule_id: string;
  type: number;
  ds_type: string;
  allow_multi_table_scan: boolean;
  task_process_info: string;
  task_result_info: string;
  task_status: 'open' | 'close'; // 任务状态：open开启，close关闭
  status: 'open' | 'close'; // 任务状态：open开启，close关闭
  operations: string[];
}

export interface ScanTaskListResponse {
  total_count: number;
  entries: ScanTaskItem[];
}

export interface TableInfo {
  id: string;
  name: string;
  create_time?: string;
  advanced_params: GetExcelColumnsRequest;
}

export interface TableListResponse {
  total_count: number;
  entries: TableInfo[];
}

export interface ColumnInfo {
  id: string;
  name: string;
  comment: string;
  type: string;
}

export interface ColumnListResponse {
  total_count: number;
  entries: ColumnInfo[];
}

export interface PageQueryParams {
  limit?: number;
  offset?: number;
  keyword?: string;
  sort?: string;
  direction?: string;
  name?: string;
  ds_id?: string;
  status?: Status;
  filters?: Record<string, any>;
}

export interface ScanTaskInfoParams {
  limit?: number;
  offset?: number;
  keyword?: string;
  sort?: string;
  direction?: string;
  status?: Status;
}

export interface ScanTaskInfoResponseItem {
  task_id: string;
  table_id: string;
  table_name: string;
  scan_status: Status;
  start_time: string;
}

export interface ScanTaskInfoResponse {
  total_count: number;
  entries: ScanTaskInfoResponseItem[];
}

export interface CreateExcelTableRequest {
  datasource_id: string;
  catalog: string;
  file_name: string;
  table_name: string;
  sheet?: string;
  all_sheet?: boolean;
  sheet_as_new_column?: boolean;
  start_cell: string;
  end_cell: string;
  has_headers?: boolean;
  columns: Array<{
    column: string;
    type: string;
  }>;
}

export interface ExcelTableResponse {
  tableId: string;
  tableName: string;
}

export interface Column {
  column: string;
  type: string;
}

export interface ExcelColumnsResponse {
  data: Column[];
}

export interface GetExcelColumnsRequest {
  id?: string;
  create_time?: string;
  datasource_id?: string;
  catalog: string;
  file_name: string;
  sheet?: string;
  all_sheet?: boolean;
  sheet_as_new_column?: boolean;
  start_cell: string;
  end_cell: string;
  has_headers?: boolean;
}

export interface ExcelFileListResponse {
  data: string[];
  total: number;
}

export interface ExcelSheetListResponse {
  data: string[];
  total: number;
}

export interface ScheduleScanStatusResponse {
  last_scan_task_id: string; // 上次扫描任务ID
  duration: string; // 持续时间
  scan_strategy: string[]; // 扫描策略 insert update delete : 'open' | 'close'; // 任务状态：open开启，close关闭
  cron_expression: CronExpression; // cron表达式 CronExpression
  task_status: string; // 任务状态
  scan_status: string; // 扫描状态
  start_time: string; // 开始时间
  end_time: string; // 结束时间
}

export interface ScheduleHistoryItem {
  duration: string; // 持续时间
  task_id: string; // 任务ID
  scan_status: string; // 扫描状态
  start_time: string; // 开始时间
  end_time: string; // 结束时间
  task_process_info: string; // 任务处理信息
  task_result_info: string; // 任务结果信息
}

export interface ScheduleHistoryListResponse {
  entries: ScheduleHistoryItem[];
  total_count: number;
}

export interface UpdateScheduleStatusRequest {
  schedule_id: string; // 定时任务ID
  status: 'open' | 'close'; // 状态：enable或disable
}

export interface UpdateScheduleStatusResponse {
  status: 'open' | 'close';
}

export interface CronExpression {
  type: string;
  expression: string;
}

export interface UpdateScheduleRequest {
  schedule_id: string; // 定时任务ID
  scan_strategy: string[]; // 扫描策略 insert update delete
  cron_expression: CronExpression; // cron表达式
  status?: 'open' | 'close'; // 任务状态
}

export interface UpdateScheduleResponse {
  status: 'open' | 'close'; // 执行状态
  schedule_id: string; // 定时任务ID
}
