/** 执行频率 */
import intl from 'react-intl-universal';
import { Form, Radio, Tooltip } from 'antd';
import ARInputNumberUnit from '@/components/ARInputNumberUnit';
import { CronSelect } from '@/components/CronSelect';
import { SCHEDULE_TYPE } from '@/hooks/useConstants';
import styles from './index.module.less';

const ScheduleExpression = ({ scheduleType, form }: any): JSX.Element => {
  const renderTabs = (options: any): JSX.Element[] => {
    return options.map((item: any) => {
      return (
        item.title && (
          <Form.Item key={item.title} name="expressionType" label={item.title} style={{ marginBottom: 12 }} initialValue={item.value} required>
            <Radio.Group key={item.title} value={item.value} onChange={item.onChange}>
              {item.options.map((opt: any) => (
                <Radio.Button key={opt.value} value={opt.value}>
                  {opt.label}
                </Radio.Button>
              ))}
            </Radio.Group>
          </Form.Item>
        )
      );
    });
  };

  const config = [
    {
      title: intl.get('MetricModel.persistenceTaskSchedule'),
      value: scheduleType,
      options: [
        { label: intl.get('Global.fixRate'), value: SCHEDULE_TYPE.FIX_RATE },
        { label: intl.get('DataConnect.customFrequency'), value: SCHEDULE_TYPE.CRON },
      ],
    },
  ];

  return (
    <div className={styles['schedule-wrapper']}>
      {renderTabs(config)}
      {scheduleType === SCHEDULE_TYPE.FIX_RATE && (
        <Form.Item
          name="fixExpression"
          preserve={true}
          className={styles['expression-wrapper']}
          extra={
            <span className="g-c-text-sub" style={{ fontSize: 12, whiteSpace: 'pre-line' }}>
              {intl.get('DataConnect.fixRateTip')}
            </span>
          }
          rules={[{ required: scheduleType === SCHEDULE_TYPE.FIX_RATE, message: intl.get('MetricModel.scheduleIsEmpty') }]}
        >
          <ARInputNumberUnit unitType="tmhdwM" isDefaultMax textBefore={intl.get('Global.every')} min={1} precision={0} />
        </Form.Item>
      )}
      {scheduleType === SCHEDULE_TYPE.CRON && (
        <Form.Item
          name="cronExpression"
          preserve={true}
          className={styles['expression-wrapper']}
          extra={
            <div className="g-c-text-sub" style={{ fontSize: 12 }}>
              {intl.get('DataConnect.cronTip')}
              <Tooltip placement="bottom" title={intl.get('DataConnect.cronTipBtnDesc')}>
                <span style={{ color: '#126ee3' }}>{intl.get('DataConnect.cronTipBtn')}</span>
              </Tooltip>
            </div>
          }
          rules={[{ required: scheduleType === SCHEDULE_TYPE.CRON, message: intl.get('DataConnect.customFrequencyRequired') }]}
        >
          <CronSelect inputProps={{ style: { width: 300 } }} />
        </Form.Item>
      )}
    </div>
  );
};

export default ScheduleExpression;
