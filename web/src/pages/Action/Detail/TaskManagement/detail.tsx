import React, { useMemo } from 'react';
import intl from 'react-intl-universal';
import DetailDrawer, { DataItem } from '@/components/DetailDrawer';
import { Table, Tag, Space, Badge, Tooltip } from 'antd';
import dayjs from 'dayjs';
import type { ActionExecutionResult, ActionExecutionLogDetail } from '@/services/action/type';
import * as ActionType from '@/services/action/type';
import { formatMsToHMS } from '@/utils/time';

interface TaskDetailProps {
  visible: boolean;
  onClose: () => void;
  taskData: ActionExecutionLogDetail;
  loading: boolean;
  runStatusDescCom: (props: { success_count?: number; failed_count?: number }) => React.ReactElement;
}

const TaskDetail: React.FC<TaskDetailProps> = ({ visible, onClose, taskData, loading, runStatusDescCom }) => {
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

  const getBadge = (s: string) => {
    switch (s) {
      case 'success':
        return 'success';
      case 'failed':
        return 'error';
      default:
        return 'default';
    }
  };
  const columns: any = [
    {
      title: intl.get('Global.name'),
      dataIndex: '_display',
      key: '_display',
      width: 200,
      render: (name: string) => name || '-',
    },
    {
      title: intl.get('Global.status'),
      dataIndex: 'status',
      key: 'status',
      width: 150,
      render: (status: string) => <Badge status={getBadge(status)} text={intl.get(`Action.${status}`)} />,
    },
    {
      title: intl.get('Action.executionInfo'),
      dataIndex: 'error_message',
      key: 'error_message',
      width: 300,
      render: (errorMessage: string, record: any) => {
        const text = errorMessage || (record.status === 'success' ? intl.get('Global.success') : '');
        return (
          <Tooltip title={text}>
            <div style={{ width: '100%', maxWidth: '300px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{text}</div>
          </Tooltip>
        );
      },
    },
    {
      title: intl.get('Action.executionDuration'),
      dataIndex: 'duration_ms',
      key: 'duration_ms',
      width: 150,
      render: (duration: number) => (duration ? formatMsToHMS(duration > 1000 ? duration : 1000) : '-'),
    },
    {
      title: intl.get('Global.startTime'),
      dataIndex: 'start_time',
      key: 'start_time',
      width: 300,
      render: (time: string) => (time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'),
    },
    {
      title: intl.get('Global.endTime'),
      dataIndex: 'end_time',
      key: 'end_time',
      width: 300,
      render: (time: string) => (time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'),
    },
  ];

  const detailDrawerData = useMemo((): DataItem[] | null => {
    if (!visible || !taskData) return null;

    return [
      {
        title: intl.get('Action.basicInfo'),
        isOpen: true,
        content: [
          {
            name: intl.get('Global.startTime'),
            value: taskData.start_time ? dayjs(taskData.start_time).format('YYYY-MM-DD HH:mm:ss') : '-',
          },
          {
            name: intl.get('Global.endTime'),
            value: taskData.end_time ? dayjs(taskData.end_time).format('YYYY-MM-DD HH:mm:ss') : '-',
          },
          {
            name: intl.get('Action.triggerType'),
            value: intl.get(`Action.${taskData.trigger_type}`),
          },
          {
            name: intl.get('Action.runStatus'),
            value: <Badge status={getBadgeStatus(taskData.status)} text={intl.get(`Action.${taskData.status}`)} />,
          },
          {
            name: intl.get('Action.totalDuration'),
            value: taskData.duration_ms ? formatMsToHMS(taskData.duration_ms > 1000 ? taskData.duration_ms : 1000) : '-',
          },
          {
            name: intl.get('Action.runStatusDesc'),
            value: runStatusDescCom({ success_count: taskData.success_count, failed_count: taskData.failed_count }),
          },
        ],
      },
      {
        title: intl.get('Action.taskExecution'),
        isOpen: true,
        content: [
          {
            isOneLine: true,
            value: (
              <div style={{ width: '100%' }}>
                <Table
                  columns={columns}
                  dataSource={
                    taskData.results?.map((result: ActionExecutionResult, index: number) => ({
                      key: index,
                      unique_identity: result.unique_identity,
                      status: result.status,
                      error_message: result.error_message,
                      duration_ms: result.duration_ms,
                      _display: result._display,
                      start_time: result.start_time ? dayjs(result.start_time).format('YYYY-MM-DD HH:mm:ss') : '-',
                      end_time: result.end_time ? dayjs(result.end_time).format('YYYY-MM-DD HH:mm:ss') : '-',
                    })) || []
                  }
                  pagination={{ pageSize: 20 }}
                  scroll={{ x: 800 }}
                />
              </div>
            ),
          },
        ],
      },
    ];
  }, [visible, taskData]);

  return <DetailDrawer data={detailDrawerData} title={intl.get('Action.taskDetail')} width={1040} onClose={onClose} open={visible} loading={loading} />;
};

export default TaskDetail;
