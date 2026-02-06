import { useState, useEffect } from 'react';
import intl from 'react-intl-universal';
import { EllipsisOutlined } from '@ant-design/icons';
import { Dropdown, Empty, Badge } from 'antd';
import dayjs from 'dayjs';
import actionApi from '@/services/action';
import * as ActionType from '@/services/action/type';
import emptyImage from '@/assets/images/common/empty.png';
import noSearchResultImage from '@/assets/images/common/no_search_result.svg';
import { Title, Table, Select, Button, IconFont } from '@/web-library/common';
import { CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';
import TaskDetail from './detail';
import { formatMsToHMS } from '@/utils/time';
import styles from './index.module.less';
import HOOKS from '@/hooks';

interface TaskManagementProps {
  knId: string;
  atId: string;
  refreshTask?: boolean;
  onRefreshComplete?: () => void;
}
const runStatusDescCom = ({ success_count = 0, failed_count = 0 }: { success_count?: number; failed_count?: number }) =>
  <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
            <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
              <CheckCircleOutlined style={{ color: '#52c41a' }} />
              {success_count}
            </span>
            <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
              <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
              {failed_count}
            </span>
          </div>
const TaskManagement = ({ knId, atId, refreshTask, onRefreshComplete }: TaskManagementProps) => {
  const ACTION_EXECUTION_STATE_LABELS: Record<ActionType.ActionExecutionStatusEnum, string> = {
    [ActionType.ActionExecutionStatusEnum.Pending]: intl.get('Action.pending'),
    [ActionType.ActionExecutionStatusEnum.Running]: intl.get('Action.running'),
    [ActionType.ActionExecutionStatusEnum.Completed]: intl.get('Action.completed'),
    [ActionType.ActionExecutionStatusEnum.Failed]: intl.get('Action.failed'),
    [ActionType.ActionExecutionStatusEnum.Cancelled]: intl.get('Action.cancelled')
  };
  const ACTION_EXECUTION_STATE_OPTIONS = [
    { value: '', label: intl.get('Global.all') },
    { value: ActionType.ActionExecutionStatusEnum.Pending, label: ACTION_EXECUTION_STATE_LABELS[ActionType.ActionExecutionStatusEnum.Pending] },
    { value: ActionType.ActionExecutionStatusEnum.Running, label: ACTION_EXECUTION_STATE_LABELS[ActionType.ActionExecutionStatusEnum.Running] },
    { value: ActionType.ActionExecutionStatusEnum.Completed, label: ACTION_EXECUTION_STATE_LABELS[ActionType.ActionExecutionStatusEnum.Completed] },
    { value: ActionType.ActionExecutionStatusEnum.Failed, label: ACTION_EXECUTION_STATE_LABELS[ActionType.ActionExecutionStatusEnum.Failed] },
    { value: ActionType.ActionExecutionStatusEnum.Cancelled, label: ACTION_EXECUTION_STATE_LABELS[ActionType.ActionExecutionStatusEnum.Cancelled] },
  ];
  const TRIGGER_TYPE_OPTIONS = [
    { value: '', label: intl.get('Global.all') },
    { value: 'manual', label: intl.get('Action.manual') },
    { value: 'schedule', label: intl.get('Action.schedule') },
    { value: 'event', label: intl.get('Action.event') },
  ];
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<ActionType.ActionExecutionLog[]>([]);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });
  const [filterValues, setFilterValues] = useState<any>({ keyword: '', trigger_type: '', status: '' });
  const [detailVisible, setDetailVisible] = useState(false);
  const [currentTask, setCurrentTask] = useState<ActionType.ActionExecutionLogDetail | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const { modal } = HOOKS.useGlobalContext();

  const fetchTasks = async (page = 1, pageSize = 10, filters = filterValues) => {
    setLoading(true);
    try {
      const params: ActionType.QueryActionLogsRequest = {
        limit: pageSize,
        need_total: true,
        action_type_id: atId,
        offset: (page - 1) * pageSize,
      };
      if (filters.status && filters.status !== 'all') params.status = filters.status;
      if (filters.trigger_type && filters.trigger_type !== 'all') params.trigger_type = filters.trigger_type;
      if (filters.keyword) params.keyword = filters.keyword;

      const res = await actionApi.queryActionLogs(knId, params);
      setData(res.entries);
      setPagination({
        current: page,
        pageSize,
        total: res.total_count,
      });
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTasks(1, pagination.pageSize, filterValues);
  }, [knId, atId]);

  useEffect(() => {
    if (refreshTask) {
      fetchTasks(1, pagination.pageSize, filterValues);
      if (onRefreshComplete) {
        onRefreshComplete();
      }
    }
  }, [refreshTask, onRefreshComplete, pagination.pageSize, filterValues]);

  const handleTableChange = (pag: any, filters: any, sorter: any) => {
    setPagination({ ...pagination, current: pag.current, pageSize: pag.pageSize });
    fetchTasks(pag.current, pag.pageSize, filterValues);
  };

  const handleSearch = (value: string) => {
    const newFilters = { ...filterValues, keyword: value };
    setFilterValues(newFilters);
    fetchTasks(1, pagination.pageSize, newFilters);
  };

  const handleChangeFilter = (values: any) => {
    const newFilters = { ...filterValues, ...values };
    setFilterValues(newFilters);
    fetchTasks(1, pagination.pageSize, newFilters);
  };

  const handleViewDetail = async (record: ActionType.ActionExecutionLog) => {
    setDetailVisible(true);
    setDetailLoading(true);
    try {
      const res = await actionApi.getActionExecutionLogDetail(knId, record.id);
      setCurrentTask(res);
    } catch (error) {
      console.error(error);
    } finally {
      setDetailLoading(false);
    }
  };

  const handleCancelExecution = async (record: ActionType.ActionExecutionLog) => {
    modal.confirm({
      title: intl.get('Action.confirmCancelExecution'),
      content: intl.get('Action.confirmCancelExecutionDesc'),
      okText: intl.get('Global.confirm'),
      cancelText: intl.get('Global.cancel'),
      onOk: async () => {
        try {
          await actionApi.cancelActionExecution(knId, record.id);
          // 刷新任务列表
          fetchTasks(pagination.current, pagination.pageSize, filterValues);
        } catch (error) {
          console.error(error);
        }
      },
    });
  };

  const columns: any = [
    {
      title: intl.get('Global.startTime'),
      dataIndex: 'start_time',
      width: 200,
      sorter: true,
      __selected: true,
      render: (time: number) => (time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '--'),
    },
    {
      title: intl.get('Global.operation'),
      dataIndex: 'operation',
      width: 80,
      __selected: true,
      render: (_value: any, record: ActionType.ActionExecutionLog) => {
        const isCancellable = [ActionType.ActionExecutionStatusEnum.Pending, ActionType.ActionExecutionStatusEnum.Running].includes(record.status);
        const dropdownMenu: any = [
          { key: 'view', label: intl.get('Global.view'), visible: true },
          { key: 'stop', label: intl.get('Action.stop'), visible: isCancellable }
        ];
        return (
          <Dropdown
            trigger={['click']}
            menu={{
              items: dropdownMenu.filter((item: { visible: boolean }) => item.visible).map(({ key, label }: any) => ({ key, label })),
              onClick: (event: any) => {
                event.domEvent.stopPropagation();
                if (event.key === 'view') handleViewDetail(record);
                if (event.key === 'stop') handleCancelExecution(record);
              },
            }}
          >
            <Button.Icon icon={<EllipsisOutlined style={{ fontSize: 20 }} />} onClick={(event) => event.stopPropagation()} />
          </Dropdown>
        );
      },
    },
    {
      title: intl.get('Action.triggerType'),
      dataIndex: 'trigger_type',
      width: 120,
      __selected: true,
      render: (triggerType: string) => triggerType ? intl.get(`Action.${triggerType}`) : '--',
    },
    {
      title: intl.get('Action.runStatus'),
      dataIndex: 'status',
      width: 150,
      __selected: true,
      render: (status: ActionType.ActionExecutionStatusEnum) => {
        const getBadgeStatus = (s: ActionType.ActionExecutionStatusEnum) => {
          switch (s) {
            case ActionType.ActionExecutionStatusEnum.Completed:
              return 'success';
            case ActionType.ActionExecutionStatusEnum.Pending:
              return 'default';
            case ActionType.ActionExecutionStatusEnum.Failed:
              return 'error';
            case ActionType.ActionExecutionStatusEnum.Running:
              return 'processing';
            case ActionType.ActionExecutionStatusEnum.Cancelled:
              return 'warning';
            default:
              return 'default';
          }
        };
        
        return (
          <Badge 
            status={getBadgeStatus(status)} 
            text={ACTION_EXECUTION_STATE_LABELS[status] || status} 
          />
        );
      },
    },
    {
      title: intl.get('Action.runStatusDesc'),
      dataIndex: 'runStatusDesc',
      width: 200,
      __selected: true,
      render: (_value: any, record: ActionType.ActionExecutionLog) => {
        const { success_count = 0, failed_count = 0 } = record;
        return runStatusDescCom({ success_count, failed_count });
      },
    },
    {
      title: intl.get('Action.totalDuration'),
      dataIndex: 'duration_ms',
      width: 100,
      sorter: true,
      __selected: true,
      render: (duration: number) => (duration ? formatMsToHMS(duration > 1000 ? duration : 1000) : '-'),
    },
    {
      title: intl.get('Global.endTime'),
      dataIndex: 'end_time',
      width: 200,
      sorter: true,
      __selected: true,
      render: (endTime: number, record: ActionType.ActionExecutionLog) => (endTime ? dayjs(endTime).format('YYYY-MM-DD HH:mm:ss') : '--'),
    },
    // {
    //   title: intl.get('Global.creator'),
    //   dataIndex: 'creator',
    //   width: 120,
    //   __selected: true,
    //   render: () => '--',
    // },
  ];

  return (
    <div className={styles['task-management-root']}>
      <Title>{intl.get('Global.taskManagement')}</Title>
      <Table.PageTable
        name="task"
        rowKey="id"
        columns={columns}
        loading={loading}
        dataSource={data}
        pagination={pagination}
        onChange={handleTableChange}
        locale={{
          emptyText:
            filterValues.keyword || filterValues.status !== '' ? (
              <Empty image={noSearchResultImage} description={intl.get('Global.emptyNoSearchResult')} />
            ) : (
              <Empty image={emptyImage} description={intl.get('Global.noData')} />
            ),
        }}
      >
        <Table.Operation
          nameConfig={{ key: 'keyword', placeholder: intl.get('Global.searchName') }}
          initialFilter={filterValues}
          onChange={handleChangeFilter}
          isSearch={false}
          onRefresh={() => fetchTasks(1, pagination.pageSize, filterValues)}
        >
          <Select.LabelSelect
            key="trigger_type"
            label={intl.get('Action.triggerType')}
            defaultValue="all"
            style={{ width: 190 }}
            options={TRIGGER_TYPE_OPTIONS}
          />
          <Select.LabelSelect
            key="status"
            label={intl.get('Global.status')}
            defaultValue="all"
            style={{ width: 190 }}
            options={ACTION_EXECUTION_STATE_OPTIONS}
          />
        </Table.Operation>
      </Table.PageTable>

      {currentTask && (
        <TaskDetail 
          visible={detailVisible} 
          onClose={() => setDetailVisible(false)} 
          taskData={currentTask} 
          loading={detailLoading}
          runStatusDescCom={runStatusDescCom}
        />
      )}
    </div>
  );
};

export default TaskManagement;
