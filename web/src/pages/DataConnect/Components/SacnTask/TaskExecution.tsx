import React, { useState, useEffect, useCallback } from 'react';
import intl from 'react-intl-universal';
import { DatabaseOutlined, LoadingOutlined } from '@ant-design/icons';
import { Table, Button, Empty, Badge, Radio, Breadcrumb, Input, RadioChangeEvent, Typography } from 'antd';
import dayjs from 'dayjs';
import { PAGINATION_DEFAULT, DATE_FORMAT } from '@/hooks/useConstants';
import { getScanTaskInfo, getScheduleHistoryList } from '@/services/scanManagement';
import * as ScanTaskType from '@/services/scanManagement/type';
import { getScanStatusColor } from '../../utils';
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table';

const { Text } = Typography;

interface TaskExecutionProps {
  scanDetail: ScanTaskType.ScanTaskItem;
  scheduleStatus?: ScanTaskType.ScheduleScanStatusResponse | null;
  visible: boolean;
}

const TaskExecution: React.FC<TaskExecutionProps> = ({ scanDetail, scheduleStatus, visible }) => {
  // 状态管理
  const [mode, setMode] = useState('current'); // current: 当前执行, history: 历史记录
  const [selectedHistoryTask, setSelectedHistoryTask] = useState<ScanTaskType.ScheduleHistoryItem | null>(null);
  const [subTaskData, setSubTaskData] = useState<ScanTaskType.ScanTaskInfoResponseItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [pagination, setPagination] = useState<TablePaginationConfig>({
    current: PAGINATION_DEFAULT.current,
    pageSize: PAGINATION_DEFAULT.pageSize,
    total: 0,
  });

  // Format duration in seconds to human-readable format: d h m s
  const formatDuration = (seconds: number | string | undefined) => {
    if (!seconds) return '--';
    const secs = Number(seconds);
    if (isNaN(secs) || secs < 0) return '--';

    const days = Math.floor(secs / (24 * 3600));
    const hours = Math.floor((secs % (24 * 3600)) / 3600);
    const minutes = Math.floor((secs % 3600) / 60);
    const remainingSeconds = Math.floor(secs % 60);

    const parts: string[] = [];
    if (days > 0) parts.push(`${days}d`);
    if (hours > 0) parts.push(`${hours}h`);
    if (minutes > 0) parts.push(`${minutes}m`);
    if (remainingSeconds > 0 || parts.length === 0) parts.push(`${remainingSeconds}s`);

    return parts.join('');
  };

  // 定时扫描状态和历史记录
  const [historyData, setHistoryData] = useState<ScanTaskType.ScheduleHistoryItem[]>([]);
  const [historyPagination, setHistoryPagination] = useState<TablePaginationConfig>({
    current: PAGINATION_DEFAULT.current,
    pageSize: PAGINATION_DEFAULT.pageSize,
    total: 0,
  });
  const [historyLoading, setHistoryLoading] = useState(false);
  const [showHistoryDetail, setShowHistoryDetail] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [historyDataTotal, setHistoryDataTotal] = useState(0);

  // 从 scanDetail 中解析统计数据
  const { table_count = 0, success_count = 0, fail_count = 0 } = JSON.parse(scanDetail.task_result_info || '{}');

  // 获取扫描任务子表数据
  const fetchSubTaskData = async (page: number = 1, pageSize: number = PAGINATION_DEFAULT.pageSize, taskId?: string, keyword?: string) => {
    setIsLoading(true);
    try {
      const params: ScanTaskType.ScanTaskInfoParams = {
        limit: pageSize,
        offset: (page - 1) * pageSize,
        keyword: keyword || searchText, // 添加搜索关键词参数
      };

      // 使用传入的taskId或默认的主任务ID
      const currentTaskId = taskId || scanDetail.id;
      const response = await getScanTaskInfo(currentTaskId, params);
      setSubTaskData(response.entries);
      setPagination((prev) => ({
        ...prev,
        total: response.total_count,
        current: page,
      }));
    } catch (error) {
      console.error('Failed to get scan task details:', error);
    } finally {
      setIsLoading(false);
    }
  };

  // 处理分页变化
  const handleTableChange = (newPagination: TablePaginationConfig) => {
    setPagination(newPagination);
    fetchSubTaskData(newPagination.current, newPagination.pageSize, selectedHistoryTask?.task_id, searchText);
  };

  // 防抖搜索函数
  // eslint-disable-next-line @typescript-eslint/no-unsafe-function-type
  const debounce = (func: Function, wait: number) => {
    let timeout: NodeJS.Timeout;
    return function executedFunction(...args: any[]) {
      const later = () => {
        clearTimeout(timeout);
        func(...args);
      };
      clearTimeout(timeout);
      timeout = setTimeout(later, wait);
    };
  };

  // 防抖搜索处理
  const debouncedSearch = useCallback(
    debounce((value: string, taskId?: string) => {
      fetchSubTaskData(1, pagination.pageSize, taskId, value);
    }, 300),
    [pagination.pageSize]
  );

  // 处理搜索
  const handleSearch = (value: string) => {
    setSearchText(value);
    debouncedSearch(value, selectedHistoryTask?.task_id);
  };

  // 获取定时扫描历史记录
  const fetchHistoryList = async (page: number = 1, pageSize: number = PAGINATION_DEFAULT.pageSize) => {
    if (!scanDetail.schedule_id) return;
    setHistoryLoading(true);
    try {
      const params = {
        limit: pageSize,
        offset: (page - 1) * pageSize,
      };
      const response = await getScheduleHistoryList(scanDetail.schedule_id, params);
      setHistoryData(response.entries);
      setHistoryDataTotal(response.total_count);
      setHistoryPagination((prev) => ({
        ...prev,
        total: response.total_count,
        current: page,
      }));
    } catch (error) {
      console.error('Failed to get history list:', error);
    } finally {
      setHistoryLoading(false);
    }
  };

  // 处理历史记录分页变化
  const handleHistoryTableChange = (newPagination: TablePaginationConfig) => {
    setHistoryPagination(newPagination);
    fetchHistoryList(newPagination.current, newPagination.pageSize);
  };

  // 当visible变化时，重新获取数据
  useEffect(() => {
    if (visible) {
      // 初始加载时清空搜索
      setSearchText('');
      // 只调用一次，不依赖searchText
      fetchSubTaskData(1, pagination.pageSize, undefined, '');
    }
  }, [visible]);

  // 查看历史任务详情
  const handleViewHistoryTask = (record: ScanTaskType.ScheduleHistoryItem) => {
    setSelectedHistoryTask(record);
    setShowHistoryDetail(true);
    // 查看历史详情时清空搜索
    setSearchText('');
    // 获取历史任务的子表列表，不使用之前的搜索条件
    fetchSubTaskData(1, pagination.pageSize, record.task_id, '');
  };

  // 返回历史记录列表
  const handleBackToHistory = () => {
    setShowHistoryDetail(false);
    setSelectedHistoryTask(null);
    // 返回历史列表时清空搜索
    setSearchText('');
    // 返回到历史列表时，恢复显示当前任务的子表数据，不使用之前的搜索条件
    fetchSubTaskData(1, pagination.pageSize, undefined, '');
  };

  // 表格列定义 - 当前执行
  const currentColumns: ColumnsType<ScanTaskType.ScanTaskInfoResponseItem> = [
    {
      title: '表名',
      dataIndex: 'table_name',
      key: 'table_name',
      ellipsis: true,
      render: (text) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <DatabaseOutlined />
          <Text ellipsis={{ tooltip: text }}>{text}</Text>
        </div>
      ),
    },
    {
      title: '扫描状态',
      dataIndex: 'scan_status',
      key: 'scan_status',
      render: (text: string) => {
        const { label, color } = getScanStatusColor(text);
        return <Badge status={color} text={label} />;
      },
    },
    {
      title: '开始时间',
      dataIndex: 'start_time',
      key: 'start_time',
      sorter: (a, b) => new Date(a.start_time || 0).getTime() - new Date(b.start_time || 0).getTime(),
      render: (text: string) => (text ? dayjs(text).format(DATE_FORMAT.DEFAULT) : '--'),
    },
    {
      title: '结束时间',
      dataIndex: 'end_time',
      key: 'end_time',
      render: (text: string) => (text ? dayjs(text).format(DATE_FORMAT.DEFAULT) : '--'),
    },
  ];

  // 历史记录表格列定义
  const historyColumns: ColumnsType<ScanTaskType.ScheduleHistoryItem> = [
    {
      title: '执行状态',
      dataIndex: 'scan_status',
      key: 'scan_status',
      render: (text: string) => {
        const { label, color } = getScanStatusColor(text);
        return <Badge status={color} text={label} />;
      },
    },
    {
      title: '耗时',
      dataIndex: 'duration',
      key: 'duration',
      render: (duration) => formatDuration(duration),
    },
    {
      title: '开始时间',
      dataIndex: 'start_time',
      key: 'start_time',
      sorter: (a, b) => new Date(a.start_time || 0).getTime() - new Date(b.start_time || 0).getTime(),
      render: (text: string) => (text ? dayjs(text).format(DATE_FORMAT.DEFAULT) : '--'),
    },
    {
      title: '结束时间',
      dataIndex: 'end_time',
      key: 'end_time',
      render: (text: string) => (text ? dayjs(text).format(DATE_FORMAT.DEFAULT) : '--'),
    },
    {
      title: '详情',
      key: 'action',
      render: (_, record) => (
        <Button type="link" size="small" onClick={() => handleViewHistoryTask(record)}>
          查看
        </Button>
      ),
    },
  ];

  // 处理模式切换
  const handleModeChange = (e: RadioChangeEvent) => {
    setMode(e.target.value);
    // 切换页签时清空搜索
    setSearchText('');
  };

  // 当类型变化时，确保模式正确
  useEffect(() => {
    if (mode === 'history' && scanDetail.type !== 2) {
      setMode('current');
    }
  }, [mode, scanDetail.type]);

  // 渲染组件内容
  return (
    <>
      {/* 布局切换 */}
      {scanDetail.type === 2 && (
        <Radio.Group onChange={handleModeChange} value={mode} style={{ marginBottom: 8 }}>
          <Radio.Button value="current">当前执行</Radio.Button>
          <Radio.Button value="history">历史记录({historyDataTotal})</Radio.Button>
        </Radio.Group>
      )}

      {/* 内容区域 */}
      {mode === 'current' ? (
        /* 当前执行视图 */
        <>
          {/* 顶部概览 */}
          <div style={{ marginBottom: 16, padding: '8px 0', display: 'flex', alignItems: 'center', justifyContent: 'flex-start', gap: '24px' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <div style={{ fontSize: 14, color: '#666', marginRight: '8px' }}>状态:</div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '4px', fontSize: 14 }}>
                {scheduleStatus?.scan_status === 'running' && <LoadingOutlined spin />}
                <span style={{ color: getScanStatusColor(scheduleStatus?.scan_status || scanDetail.scan_status || '').color }}>
                  {getScanStatusColor(scheduleStatus?.scan_status || scanDetail.scan_status || '').label || '--'}
                </span>
              </div>
            </div>
            <div style={{ width: '1px', height: '16px', backgroundColor: '#e8e8e8' }}></div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <div style={{ fontSize: 14, color: '#666', marginRight: '8px' }}>耗时:</div>
              <div style={{ fontSize: 14 }}>{formatDuration(scheduleStatus?.duration)}</div>
            </div>
            <div style={{ width: '1px', height: '16px', backgroundColor: '#e8e8e8' }}></div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <div style={{ fontSize: 14, color: '#666', marginRight: '8px' }}>开始时间:</div>
              <div style={{ fontSize: 14 }}>{scheduleStatus?.start_time ? dayjs(scheduleStatus.start_time).format(DATE_FORMAT.DEFAULT) : '--'}</div>
            </div>
            <div style={{ width: '1px', height: '16px', backgroundColor: '#e8e8e8' }}></div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <div style={{ fontSize: 14, color: '#666', marginRight: '8px' }}>结束时间:</div>
              <div style={{ fontSize: 14 }}>{scheduleStatus?.end_time ? dayjs(scheduleStatus.end_time).format(DATE_FORMAT.DEFAULT) : '--'}</div>
            </div>
          </div>

          {/* 扫描目标列表 */}
          <div style={{ marginBottom: 16 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
              <h4 style={{ margin: 0, fontWeight: 500 }}>{intl.get('Global.scanTarget')}</h4>
              <Input.Search
                placeholder={intl.get('Global.search')}
                allowClear
                style={{ width: 240 }}
                value={searchText}
                onSearch={handleSearch}
                onChange={(e) => handleSearch(e.target.value || '')}
              />
            </div>
            <Table
              columns={currentColumns}
              dataSource={subTaskData}
              rowKey="id"
              loading={isLoading}
              size="small"
              scroll={{ y: 400 }}
              pagination={{
                ...pagination,
                showSizeChanger: PAGINATION_DEFAULT.showSizeChanger,
                showQuickJumper: PAGINATION_DEFAULT.showQuickJumper,
                showTotal: (total) => intl.get('Global.total', { total }),
                pageSizeOptions: [...PAGINATION_DEFAULT.pageSizeOptions],
              }}
              onChange={handleTableChange}
              locale={{
                emptyText: <Empty description={intl.get('Global.noData')} />,
              }}
            />
          </div>
        </>
      ) : (
        /* 历史记录视图 */
        <>
          {/* 历史记录详情视图 */}
          {showHistoryDetail && selectedHistoryTask ? (
            <>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
                <Breadcrumb separator="/">
                  <Breadcrumb.Item onClick={handleBackToHistory}>
                    <span style={{ cursor: 'pointer' }}>历史记录</span>
                  </Breadcrumb.Item>
                  <Breadcrumb.Item>{`开始时间: ${dayjs(selectedHistoryTask.start_time).format(DATE_FORMAT.DEFAULT)}`}</Breadcrumb.Item>
                </Breadcrumb>
                <Input.Search
                  placeholder={'搜索'}
                  allowClear
                  style={{ width: 240 }}
                  value={searchText}
                  onSearch={handleSearch}
                  onChange={(e) => handleSearch(e.target.value || '')}
                />
              </div>

              {/* 顶部统计概览 */}
              <div style={{ marginBottom: 16, padding: '8px 0', display: 'flex', alignItems: 'center', justifyContent: 'flex-start', gap: '24px' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <div style={{ fontSize: 14, color: '#666', marginRight: '8px' }}>表总数:</div>
                  <div style={{ fontSize: 14 }}>{table_count || 0}</div>
                </div>
                <div style={{ width: '1px', height: '16px', backgroundColor: '#e8e8e8' }}></div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <div style={{ fontSize: 14, color: '#666', marginRight: '8px' }}>成功数:</div>
                  <div style={{ fontSize: 14, color: '#52c41a' }}>{success_count || 0}</div>
                </div>
                <div style={{ width: '1px', height: '16px', backgroundColor: '#e8e8e8' }}></div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <div style={{ fontSize: 14, color: '#666', marginRight: '8px' }}>失败数:</div>
                  <div style={{ fontSize: 14, color: '#ff4d4f' }}>{fail_count || 0}</div>
                </div>
              </div>
              <Table
                columns={currentColumns}
                dataSource={subTaskData}
                rowKey="id"
                loading={isLoading}
                size="small"
                scroll={{ y: 500 }}
                pagination={{
                  ...pagination,
                  showSizeChanger: PAGINATION_DEFAULT.showSizeChanger,
                  showQuickJumper: PAGINATION_DEFAULT.showQuickJumper,
                  showTotal: (total) => intl.get('Global.total', { total }),
                  pageSizeOptions: [...PAGINATION_DEFAULT.pageSizeOptions],
                }}
                onChange={handleTableChange}
                locale={{
                  emptyText: <Empty description={intl.get('Global.noData')} />,
                }}
              />
            </>
          ) : (
            /* 历史记录列表视图 */
            <>
              {/* 历史执行记录列表 */}
              <Table
                columns={historyColumns}
                dataSource={historyData}
                rowKey="task_id"
                loading={historyLoading}
                size="small"
                scroll={{ y: 500 }}
                pagination={{
                  ...historyPagination,
                  showSizeChanger: PAGINATION_DEFAULT.showSizeChanger,
                  showQuickJumper: PAGINATION_DEFAULT.showQuickJumper,
                  showTotal: (total) => intl.get('Global.total', { total }),
                  pageSizeOptions: [...PAGINATION_DEFAULT.pageSizeOptions],
                }}
                onChange={handleHistoryTableChange}
                locale={{
                  emptyText: <Empty description={intl.get('Global.noData')} />,
                }}
              />
            </>
          )}
        </>
      )}
    </>
  );
};

export default TaskExecution;
