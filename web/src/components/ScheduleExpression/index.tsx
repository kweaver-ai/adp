import React, { useState, useEffect } from 'react';
import intl from 'react-intl-universal';
import { Form, Radio, Tooltip, Switch, Row, Col } from 'antd';
import { RadioChangeEvent } from 'antd/lib/radio';
import ARInputNumberUnit from '@/components/ARInputNumberUnit';
import { CronSelect } from '@/components/CronSelect';
import { CronFieldName, validateCronField } from '@/components/CronSelect/Cron/utils';
import { SCHEDULE_TYPE } from '@/hooks/useConstants';
import styles from './index.module.less';

/**
 * 验证完整的Cron表达式
 * @param value Cron表达式字符串
 * @returns 验证结果
 */
const validateCronExpression = (value: string): boolean => {
  if (!value) return false;
  
  const fields = value.trim().split(' ');
  // 标准Cron表达式有6个字段：秒、分、时、日、月、周
  if (fields.length < 6) return false;
  
  const fieldNames: CronFieldName[] = [CronFieldName.SECOND, CronFieldName.MINUTE, CronFieldName.HOUR, CronFieldName.DAY, CronFieldName.MONTH, CronFieldName.WEEK];
  
  for (let i = 0; i < 6; i++) {
    if (!validateCronField(fields[i], fieldNames[i])) {
      return false;
    }
  }
  
  return true;
};

interface ScheduleExpressionProps {
  showSwitch?: boolean;
  switchName?: string;
  typeFieldName?: string;
  fixExpressionFieldName?: string;
  cronExpressionFieldName?: string;
  form?: any;
}

const ScheduleExpression = ({
  showSwitch = false,
  switchName = 'scheduleEnabled',
  typeFieldName = 'schedule.type',
  fixExpressionFieldName = 'schedule.FIX_RATE.expression',
  cronExpressionFieldName = 'schedule.CRON.expression',
  form,
}: ScheduleExpressionProps): JSX.Element => {
  const [scheduleType, setScheduleType] = useState(SCHEDULE_TYPE.FIX_RATE);

  // Watch for external changes to the type field if form is provided
  const formScheduleType = Form.useWatch(typeFieldName, form);
  const isEnabled = Form.useWatch(switchName, form);

  useEffect(() => {
    if (formScheduleType) {
      setScheduleType(formScheduleType);
    }
  }, [formScheduleType]);

  const renderTabs = (options: any): JSX.Element[] => {
    return options.map((item: any) => {
      return (
        <Form.Item
          key={item.title || 'type'}
          name={typeFieldName}
          label={item.title}
          style={showSwitch ? { marginBottom: 0 } : { marginBottom: 12 }}
          // initialValue={item.value}
          rules={[{ required: showSwitch ? isEnabled : true, message: intl.get('Global.required') }]}
        >
          <Radio.Group key={item.title || 'type'} value={item.value} onChange={item.onChange}>
            {item.options.map((opt: any) => (
              <Radio.Button key={opt.value} value={opt.value}>
                {opt.label}
              </Radio.Button>
            ))}
          </Radio.Group>
        </Form.Item>
      );
    });
  };

  const config = [
    {
      title: null, // No title for the radio group itself in the new design? Or maybe "Execution Frequency"?
      // The image shows "Fixed Rate" and "Custom Frequency" as tabs/buttons next to the switch?
      // Wait, the image shows:
      // "Execution Frequency" [Switch] [Fixed Rate] [Custom Frequency]
      // This implies the Radio Group is right next to the switch or below it.
      // If I look at the DataConnect component, it has a title "MetricModel.persistenceTaskSchedule".
      // In the new design, if switch is present, maybe we don't need that title or it's outside.
      value: scheduleType,
      onChange: (e: RadioChangeEvent) => setScheduleType(e.target.value),
      options: [
        { label: intl.get('Global.fixRate'), value: SCHEDULE_TYPE.FIX_RATE },
        { label: intl.get('DataConnect.customFrequency'), value: SCHEDULE_TYPE.CRON },
      ],
    },
  ];

  const content = (
    <div className={styles['schedule-wrapper']}>
      {renderTabs(config)}
      {scheduleType === SCHEDULE_TYPE.FIX_RATE && (
        <Form.Item
          name={fixExpressionFieldName}
          preserve={true}
          className={styles['expression-wrapper']}
          extra={
            <span className="g-c-text-sub" style={{ fontSize: 12, whiteSpace: 'pre-line' }}>
              {intl.get('DataConnect.fixRateTip')}
            </span>
          }
          rules={[{ required: showSwitch ? isEnabled : true, message: intl.get('MetricModel.scheduleIsEmpty') }]}
        >
          <ARInputNumberUnit unitType="tmhdwM" isDefaultMax textBefore={intl.get('Global.every')} min={1} precision={0} />
        </Form.Item>
      )}
      {scheduleType === SCHEDULE_TYPE.CRON && (
        <Form.Item
          name={cronExpressionFieldName}
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
          rules={[
            { required: showSwitch ? isEnabled : true, message: intl.get('DataConnect.customFrequencyRequired') },
            { 
              validator: (_, value) => {
                if (!value) return Promise.resolve();
                if (validateCronExpression(value)) {
                  return Promise.resolve();
                }
                return Promise.reject(new Error(intl.get('DataConnect.cronExpressionInvalid')));
              }
            }
          ]}
        >
          <CronSelect inputProps={{ style: { width: 300 } }} />
        </Form.Item>
      )}
    </div>
  );

  if (showSwitch) {
    return (
      <div>
        <Form.Item label={intl.get('MetricModel.persistenceTaskSchedule')} style={{ marginBottom: 0 }}>
          <Row align="middle">
            <Col>
              <Form.Item name={switchName} valuePropName="checked" noStyle>
                <Switch />
              </Form.Item>
            </Col>
            <Col style={{ marginLeft: 16 }}>{isEnabled && renderTabs(config)}</Col>
          </Row>
        </Form.Item>
        {isEnabled && (
          <div style={{ marginTop: 12 }}>
            {scheduleType === SCHEDULE_TYPE.FIX_RATE && (
              <Form.Item
                name={fixExpressionFieldName}
                preserve={true}
                className={styles['expression-wrapper']}
                extra={
                  <span className="g-c-text-sub" style={{ fontSize: 12, whiteSpace: 'pre-line' }}>
                    {intl.get('DataConnect.fixRateTip')}
                  </span>
                }
                rules={[{ required: true, message: intl.get('MetricModel.scheduleIsEmpty') }]}
              >
                <ARInputNumberUnit unitType="tmhdwM" isDefaultMax textBefore={intl.get('Global.every')} min={1} precision={0} />
              </Form.Item>
            )}
            {scheduleType === SCHEDULE_TYPE.CRON && (
              <Form.Item
                name={cronExpressionFieldName}
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
                rules={[
                  { required: true, message: intl.get('DataConnect.customFrequencyRequired') },
                  { 
                    validator: (_, value) => {
                      if (!value) return Promise.resolve();
                      if (validateCronExpression(value)) {
                        return Promise.resolve();
                      }
                      return Promise.reject(new Error(intl.get('DataConnect.cronExpressionInvalid')));
                    }
                  }
                ]}
              >
                <CronSelect inputProps={{ style: { width: 300 } }} />
              </Form.Item>
            )}
          </div>
        )}
      </div>
    );
  }

  return content;
};

export default ScheduleExpression;
