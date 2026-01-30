import { useState, useEffect } from 'react';
import intl from 'react-intl-universal';
import { EllipsisOutlined } from '@ant-design/icons';
import { Dropdown, Empty, Modal, Drawer, Spin } from 'antd';
import classnames from 'classnames';
import dayjs from 'dayjs';
import { formatMsToHMS } from '@/utils/time';
import * as ActionType from '@/services/action/type';
import createImage from '@/assets/images/common/create.svg';
import emptyImage from '@/assets/images/common/empty.png';
import noSearchResultImage from '@/assets/images/common/no_search_result.svg';
import HOOKS from '@/hooks';
import { Title, Table, Select, Button, IconFont } from '@/web-library/common';
import actionApi from '@/services/action';
import ActionInfo from '../components/ActionInfo';
import styles from './index.module.less';

interface TaskManagementProps {
  knId: string;
  atId: string;
}

const StateItem = ({ state, error }: { state: ActionType.TaskStatusEnum; error?: string }) => {
  const [errorOpen, setErrorOpen] = useState(false);
  const ACTION_TASK_STATE_LABELS: Record<ActionType.TaskStatusEnum, string> = {
    [ActionType.TaskStatusEnum.Pending]: intl.get('Action.statusPending'),
    [ActionType.TaskStatusEnum.Running]: intl.get('Action.statusRunning'),
    [ActionType.TaskStatusEnum.Success]: intl.get('Action.statusSuccess'),
    [ActionType.TaskStatusEnum.Failed]: intl.get('Action.statusFailed'),
    [ActionType.TaskStatusEnum.Canceled]: intl.get('Action.statusCanceled'),
  };

  return (
    <>
      <div className={styles['task-state']}>
        <div className={classnames(styles['task-state-icon'], styles[state])}></div>
        <div className={styles['task-state']}>{ACTION_TASK_STATE_LABELS[state] || state}</div>
        {(state === ActionType.TaskStatusEnum.Failed || state === ActionType.TaskStatusEnum.Canceled) && (
          <IconFont type="icon-dip-Details" onClick={() => setErrorOpen(true)} style={{ cursor: 'pointer' }} />
        )}
      </div>
      <Modal title={intl.get('Action.taskErrorTitle')} open={errorOpen} onCancel={() => setErrorOpen(false)} footer={null}>
        <div className={styles['task-error']}>{error}</div>
      </Modal>
    </>
  );
};

const TaskManagement = ({ knId, atId }: TaskManagementProps) => {
  const ACTION_TASK_STATE_LABELS: Record<ActionType.TaskStatusEnum, string> = {
    [ActionType.TaskStatusEnum.Pending]: intl.get('Action.statusPending'),
    [ActionType.TaskStatusEnum.Running]: intl.get('Action.statusRunning'),
    [ActionType.TaskStatusEnum.Success]: intl.get('Action.statusSuccess'),
    [ActionType.TaskStatusEnum.Failed]: intl.get('Action.statusFailed'),
    [ActionType.TaskStatusEnum.Canceled]: intl.get('Action.statusCanceled'),
  };

  const ACTION_TASK_STATE_OPTIONS = [
    { value: '', label: intl.get('Global.all') },
    { value: ActionType.TaskStatusEnum.Pending, label: ACTION_TASK_STATE_LABELS[ActionType.TaskStatusEnum.Pending] },
    { value: ActionType.TaskStatusEnum.Running, label: ACTION_TASK_STATE_LABELS[ActionType.TaskStatusEnum.Running] },
    { value: ActionType.TaskStatusEnum.Success, label: ACTION_TASK_STATE_LABELS[ActionType.TaskStatusEnum.Success] },
    { value: ActionType.TaskStatusEnum.Failed, label: ACTION_TASK_STATE_LABELS[ActionType.TaskStatusEnum.Failed] },
    { value: ActionType.TaskStatusEnum.Canceled, label: ACTION_TASK_STATE_LABELS[ActionType.TaskStatusEnum.Canceled] },
  ];
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<ActionType.Task[]>([]);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });
  const [filterValues, setFilterValues] = useState<any>({ keyword: '', trigger_type: '', status: '' });
  const [detailVisible, setDetailVisible] = useState(false);
  const [currentTask, setCurrentTask] = useState<ActionType.Task | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);

  const fetchTasks = async (page = 1, pageSize = 10, filters = filterValues) => {
    setLoading(true);
    try {
      const params: ActionType.GetTasksRequest = {
        offset: (page - 1) * pageSize,
        limit: pageSize,
        keyword: filters.keyword,
      };
      if (filters.status && filters.status !== 'all') params.status = filters.status;
      // Note: trigger_type is not in GetTasksRequest yet, but if it was supported:
      // if (filters.trigger_type && filters.trigger_type !== 'all') params.trigger_type = filters.trigger_type;

      const res = await actionApi.getActionTasks(knId, atId, params);
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

  const handleViewDetail = async (record: ActionType.Task) => {
    setDetailVisible(true);
    setDetailLoading(true);
    try {
      const res = await actionApi.getActionTaskDetail(knId, atId, record.id);
      setCurrentTask(res);
    } catch (error) {
      console.error(error);
    } finally {
      setDetailLoading(false);
    }
  };

  const columns: any = [
    {
      title: intl.get('Global.startTime'),
      dataIndex: 'start_time',
      width: 200,
      sorter: true,
      __selected: true,
      render: (time: number) => (time ? dayjs(time * 1000).format('YYYY-MM-DD HH:mm:ss') : '--'),
    },
    {
      title: intl.get('Global.operation'),
      dataIndex: 'operation',
      width: 80,
      __selected: true,
      render: (_value: any, record: ActionType.Task) => {
        const dropdownMenu: any = [{ key: 'view', label: intl.get('Global.view'), visible: true }];
        return (
          <Dropdown
            trigger={['click']}
            menu={{
              items: dropdownMenu.filter((item: { visible: boolean }) => item.visible).map(({ key, label }: any) => ({ key, label })),
              onClick: (event: any) => {
                event.domEvent.stopPropagation();
                if (event.key === 'view') handleViewDetail(record);
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
    },
    {
      title: intl.get('Global.runStatus'),
      dataIndex: 'status',
      width: 150,
      __selected: true,
      render: (status: ActionType.TaskStatusEnum, record: ActionType.Task) => {
        return <StateItem error={record?.result_desc} state={status} />;
      },
    },
    {
      title: intl.get('Action.runStatusDesc'),
      dataIndex: 'result_desc',
      width: 200,
      __selected: true,
      render: (text: string) => text || '--',
    },
    {
      title: intl.get('Action.totalDuration'),
      dataIndex: 'duration',
      width: 100,
      sorter: true,
      __selected: true,
      render: (duration: number) => (duration ? formatMsToHMS(duration * 1000) : '--'),
    },
    {
      title: intl.get('Global.endTime'),
      dataIndex: 'end_time',
      width: 200,
      sorter: true,
      __selected: true,
      render: (time: number) => (time ? dayjs(time * 1000).format('YYYY-MM-DD HH:mm:ss') : '--'),
    },
    {
      title: intl.get('Global.creator'),
      dataIndex: ['operator', 'name'],
      width: 120,
      __selected: true,
      render: (text: string) => text || '--',
    },
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
          onRefresh={() => fetchTasks(1, pagination.pageSize, filterValues)}
        >
          {/* Placeholder for trigger type filter if needed in future */}
          <Select.LabelSelect key="trigger_type" label={intl.get('Action.triggerType')} defaultValue="all" style={{ width: 190 }} options={[]} />
          <Select.LabelSelect key="status" label={intl.get('Global.status')} defaultValue="all" style={{ width: 190 }} options={ACTION_TASK_STATE_OPTIONS} />
        </Table.Operation>
      </Table.PageTable>

      <Drawer title={intl.get('Action.taskDetail')} width="60%" open={detailVisible} onClose={() => setDetailVisible(false)} maskClosable>
        {detailLoading ? (
          <div style={{ display: 'flex', justifyContent: 'center', paddingTop: 50 }}>
            <Spin />
          </div>
        ) : currentTask?.action_config ? (
          <ActionInfo knId={knId} atId={atId} detail={currentTask.action_config} />
        ) : (
          <div>{intl.get('Global.noData')}</div>
        )}
      </Drawer>
    </div>
  );
};

export default TaskManagement;
