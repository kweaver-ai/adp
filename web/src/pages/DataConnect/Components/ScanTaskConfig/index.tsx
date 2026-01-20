import { useEffect, useState } from 'react';
import intl from 'react-intl-universal';
import { Button, Radio, Checkbox, Switch, Modal, Form } from 'antd';
import { SCHEDULE_TYPE } from '@/hooks/useConstants';
import * as DataConnectType from '@/services/dataConnect/type';
import scanManagementApi from '@/services/scanManagement';
import * as ScanTaskType from '@/services/scanManagement/type';
import HOOKS from '@/hooks';
import ScheduleExpression from '../ScheduleExpression';
import styles from './styles.module.less';

interface TScanTaskConfig {
  open: boolean;
  onClose: (isOk?: boolean) => void;
  selectedDataSources?: DataConnectType.DataSource[];
  isEdit?: boolean;
  scanTaskId?: string;
  scanDetail?: ScanTaskType.ScanTaskItem;
}

const ScanTaskConfig = ({ open, onClose, selectedDataSources = [], isEdit = false, scanTaskId, scanDetail }: TScanTaskConfig): JSX.Element => {
  const { message } = HOOKS.useGlobalContext();
  const [form] = Form.useForm<{
    type: number;
    expressionType: string;
    fixExpression: string;
    cronExpression: string;
    scan_strategy: string[];
    status: boolean;
  }>();
  const [isLoading, setIsLoading] = useState(false);
  const [scheduleStatus, setScheduleStatus] = useState<ScanTaskType.ScheduleScanStatusResponse | null>(null);

  const scanType = Form.useWatch('type', form);

  // 初始化表单默认值
  const initialValues = {
    type: 0,
    expressionType: SCHEDULE_TYPE.FIX_RATE,
    fixExpression: '',
    cronExpression: '',
    scan_strategy: [],
    status: true,
  };

  // 获取配置详情
  const fetchScheduleStatus = async () => {
    if (!isEdit || !scanDetail) return;

    setIsLoading(true);
    try {
      const currentId = scanDetail.type === 2 ? scanDetail.schedule_id : scanDetail.id;
      const currentType = scanDetail.type === 2 ? 2 : 0;
      const response = await scanManagementApi.getScheduleScanStatus(currentId, currentType);
      setScheduleStatus(response);
      form.setFieldsValue({ type: scanDetail.type });
      setTimeout(() => {
        if (currentType === 2) {
          form.setFieldsValue({
            expressionType: response.cron_expression.type,
            scan_strategy: response.scan_strategy || [],
            status: scanDetail.task_status === 'open',
            fixExpression: response.cron_expression.type === SCHEDULE_TYPE.FIX_RATE ? response.cron_expression.expression : '',
            cronExpression: response.cron_expression.type === SCHEDULE_TYPE.CRON ? response.cron_expression.expression : '',
          });
        } else {
          form.setFieldsValue({
            scan_strategy: response.scan_strategy || [],
            status: scanDetail.task_status === 'open',
          });
        }
      }, 0);
    } catch (error) {
      console.error('Failed to get schedule status:', error);
      message.error(intl.get('Global.fetchFailed'));
    } finally {
      setIsLoading(false);
    }
  };

  // 当组件打开时，重置表单为初始值
  useEffect(() => {
    if (open) {
      form.resetFields();
    }
  }, [open, form]);

  // 当组件打开且为编辑模式时，获取配置详情
  useEffect(() => {
    if (open && isEdit && scanDetail) {
      fetchScheduleStatus();
    }
  }, [open, isEdit, scanDetail]);

  // 扫描数据源
  const handleOk = async (): Promise<void> => {
    try {
      const values = await form.validateFields();
      console.log(values);

      if (isEdit && scanDetail) {
        // 编辑模式：调用更新定时任务接口
        await scanManagementApi.updateSchedule({
          schedule_id: scanDetail.type === 2 ? scanDetail.schedule_id : scanDetail.id,
          scan_strategy: values.scan_strategy,
          cron_expression: {
            type: values.expressionType,
            expression: values.expressionType === SCHEDULE_TYPE.FIX_RATE ? values.fixExpression : values.cronExpression,
          },
          status: values.status ? 'open' : 'close',
        });
        message.success(intl.get('Global.updateSuccess'));
        onClose(true);
      } else {
        // 创建模式：调用批量创建扫描任务接口
        const scanTasks = selectedDataSources.map((item) => ({
          scan_name: item.name,
          ds_info: { ds_id: item.id, ds_type: item.type, scan_strategy: values.scan_strategy },
          type: values.type, // 0 立即扫描，2 定时扫描
          cron_expression: {
            type: values.expressionType,
            expression: values.expressionType === SCHEDULE_TYPE.FIX_RATE ? values.fixExpression : values.cronExpression,
          },
          status: values.status ? 'open' : 'close', // 任务状态
          // 可以根据需要添加更多配置，如频率、策略等
        }));
        await scanManagementApi.batchCreateScanTask(scanTasks);
        message.success(intl.get('Global.scanTaskSuccess'));
        onClose(true);
      }
    } catch (error) {
      console.error('操作失败:', error);
    }
  };

  return (
    <Modal
      title={intl.get('DataConnect.scanTaskConfig')}
      width={560}
      open={open}
      onCancel={() => onClose()}
      className={styles.modalWrapper}
      maskClosable={false}
      footer={
        <div className={styles.modalFooter}>
          <div className={styles.footerLeft}></div>
          <div className={styles.footerRight}>
            <Button type="primary" onClick={handleOk} className={styles.okButton}>
              {intl.get('Global.ok')}
            </Button>
            <Button onClick={() => onClose()}>{intl.get('Global.cancel')}</Button>
          </div>
        </div>
      }
    >
      <Form form={form} initialValues={initialValues} layout="vertical">
        <Form.Item name="type" label={intl.get('DataConnect.scanType')}>
          <Radio.Group className={styles.radioGroup} disabled={isEdit}>
            <Radio value={0} className={styles.radioItem}>
              {intl.get('DataConnect.ImmediateScan')}
            </Radio>
            <Radio value={2} className={styles.radioItem}>
              {intl.get('DataConnect.ScheduledScan')}
            </Radio>
          </Radio.Group>
        </Form.Item>

        {scanType === 2 && <ScheduleExpression form={form} scheduleType="FIX_RATE" />}

        <Form.Item
          name="scan_strategy"
          label={
            <div>
              {intl.get('DataConnect.scanStrategy')}
              <br />
              <span className={styles.strategyHint}>{intl.get('DataConnect.defaultFullScan')}</span>
            </div>
          }
        >
          {/* <div className={styles.strategyHint}>{intl.get('DataConnect.defaultFullScan')}</div> */}
          <Checkbox.Group className={styles.checkboxGroup}>
            <div className={styles.checkboxItem}>
              <Checkbox value="insert">{intl.get('DataConnect.onlyScanNewTables')}</Checkbox>
              <div className={styles.checkboxDesc}>{intl.get('DataConnect.onlyScanNewTablesDesc')}</div>
            </div>

            <div className={styles.checkboxItem}>
              <Checkbox value="update">{intl.get('DataConnect.onlyScanChangedTables')}</Checkbox>
              <div className={styles.checkboxDesc}>{intl.get('DataConnect.onlyScanChangedTablesDesc')}</div>
            </div>

            <div className={styles.checkboxItem}>
              <Checkbox value="delete">{intl.get('DataConnect.onlyCleanInvalidTables')}</Checkbox>
              <div className={styles.checkboxDesc}>{intl.get('DataConnect.onlyCleanInvalidTablesDesc')}</div>
            </div>
          </Checkbox.Group>
        </Form.Item>

        <Form.Item name="status" label={intl.get('DataConnect.taskStatus')} valuePropName="checked">
          <Switch disabled={scanType === 0} />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default ScanTaskConfig;
