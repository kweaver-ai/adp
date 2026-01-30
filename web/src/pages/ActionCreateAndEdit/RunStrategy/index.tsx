import React, { FC } from 'react';
import { Form } from 'antd';
import ScheduleExpression from '@/components/ScheduleExpression';

interface RunStrategyProps {
  form: any;
  value?: any;
}

const RunStrategy: FC<RunStrategyProps> = ({ form, value }) => {
  const initialValues = {
    'schedule.type': 'FIX_RATE',
    ...value,
  };

  return (
    <div style={{ padding: '24px', minWidth: 500, maxWidth: 900 }}>
      <Form form={form} layout="vertical" initialValues={initialValues}>
        <ScheduleExpression
          form={form}
          showSwitch={true}
          switchName="scheduleEnabled"
          typeFieldName="schedule.type"
          fixExpressionFieldName="schedule.FIX_RATE.expression"
          cronExpressionFieldName="schedule.CRON.expression"
        />
      </Form>
    </div>
  );
};

export default RunStrategy;
