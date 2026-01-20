import React, { useMemo, useState, useEffect } from 'react';
import intl from 'react-intl-universal';
import { FormOutlined } from '@ant-design/icons';
import { Button } from 'antd';
import DetailDrawer, { DataItem } from '@/components/DetailDrawer';
import { getScheduleScanStatus } from '@/services/scanManagement';
import * as ScanTaskType from '@/services/scanManagement/type';
import TaskExecution from './TaskExecution';
import ScanTaskConfig from '../ScanTaskConfig';

interface ScanDetailProps {
  scanDetail: ScanTaskType.ScanTaskItem;
  visible: boolean;
  onClose: () => void;
  getTableType: (type: string, val: string) => JSX.Element | string;
  isEdit?: boolean;
}

const ScanDetail: React.FC<ScanDetailProps> = ({ visible, onClose, scanDetail, getTableType, isEdit = false }) => {
  // 状态管理
  const [scheduleStatus, setScheduleStatus] = useState<ScanTaskType.ScheduleScanStatusResponse | null>(null);
  const [scanTaskConfigVisible, setScanTaskConfigVisible] = useState(false);

  // 获取定时扫描状态
  const fetchScheduleStatus = async () => {
    if (scanDetail.type === undefined) return;
    console.log(scanDetail.type, scanDetail.schedule_id, scanDetail.id, 'scanDetail.type');
    const currentId = scanDetail.type === 2 ? scanDetail.schedule_id : scanDetail.id;
    const currentType = scanDetail.type === 2 ? 2 : 0;
    console.log(currentId, currentType, 'currentId, currentType');
    try {
      const response = await getScheduleScanStatus(currentId, currentType);
      setScheduleStatus(response);
    } catch (error) {
      console.error('Failed to get schedule status:', error);
    }
  };

  // 当visible变化时，重新获取数据
  useEffect(() => {
    if (visible && scanDetail.type !== undefined) {
      fetchScheduleStatus();
    }
  }, [visible, scanDetail.schedule_id, scanDetail.type]);

  // 处理编辑点击
  const handleEditClick = () => {
    // 当type != 1时显示编辑按钮
    if (scanDetail.type === 2) {
      setScanTaskConfigVisible(true);
    }
  };

  // 处理扫描任务配置关闭
  const handleScanTaskConfigClose = (isOk?: boolean) => {
    setScanTaskConfigVisible(false);
    if (isOk) {
      // 如果保存成功，重新获取数据
      fetchScheduleStatus();
    }
  };

  // 构建DetailDrawer数据

  const detailDrawerData = useMemo((): DataItem[] | null => {
    if (!visible) return null;

    // 扫描目标和任务状态、任务类型从scanDetail获取
    // 其他从scheduleStatus获取，scheduleStatus为空时使用默认值
    const scanTarget = scanDetail.name || '--';
    const scanTargetDisplay = scanTarget === '--' ? '--' : getTableType(scanDetail.type === 1 ? '1' : scanDetail.ds_type, scanTarget);
    // 重复规则显示
    const cronExpression = scheduleStatus?.cron_expression;
    const repeatRuleDisplay = cronExpression?.expression || '--';
    const repeatRuleDisplayCom =
      scanDetail.type === 2
        ? [
            {
              name: '重复规则',
              value: repeatRuleDisplay,
            },
          ]
        : [];
    // Map scan strategy values to internationalized display names
    const getStrategyDisplayName = (strategy: string) => {
      switch (strategy) {
        case 'insert':
          return intl.get('DataConnect.onlyScanNewTables');
        case 'update':
          return intl.get('DataConnect.onlyScanChangedTables');
        case 'delete':
          return intl.get('DataConnect.onlyCleanInvalidTables');
        default:
          return strategy;
      }
    };

    const scanStrategyCom =
      scanDetail.type != 1
        ? [
            {
              name: '扫描策略',
              value: scheduleStatus?.scan_strategy?.length
                ? scheduleStatus?.scan_strategy?.map(getStrategyDisplayName).join(', ')
                : intl.get('DataConnect.fullScan'),
            },
          ]
        : [];

    return [
      {
        title: '基本信息',
        isOpen: true,
        extra:
          scanDetail.type === 2 && isEdit ? (
            <Button type="text" icon={<FormOutlined />} onClick={handleEditClick} style={{ color: '#165DFF', padding: 0 }}></Button>
          ) : undefined,
        content: [
          {
            name: '扫描目标',
            value: scanTargetDisplay,
          },
          {
            name: '任务类型',
            value: scanDetail.type === 2 ? intl.get('Global.scheduleScan') : intl.get('Global.immediateScan'),
          },
          {
            name: '任务状态',
            value: scheduleStatus?.task_status === 'open' ? '开启' : '关闭',
          },
          ...scanStrategyCom,
          ...repeatRuleDisplayCom,
        ],
      },
      {
        title: '任务执行',
        isOpen: true,
        content: [
          {
            isOneLine: true,
            value: <TaskExecution scheduleStatus={scheduleStatus} scanDetail={scanDetail} visible={visible} />,
          },
        ],
      },
    ];
    // eslint-disable-next-line react-hooks/use-memo
  }, [visible, JSON.stringify(scanDetail), JSON.stringify(scheduleStatus), getTableType]);

  return (
    <>
      <DetailDrawer data={detailDrawerData} title="扫描详情" width={1040} onClose={onClose} open={visible} />
      <ScanTaskConfig open={scanTaskConfigVisible} onClose={handleScanTaskConfigClose} isEdit={true} scanDetail={scanDetail} />
    </>
  );
};

export default ScanDetail;
