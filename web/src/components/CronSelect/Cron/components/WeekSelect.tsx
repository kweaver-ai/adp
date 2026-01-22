// @ts-nocheck
import { useEffect } from 'react';
import intl from 'react-intl-universal';
import { Select } from 'antd';
import locales from '../../locales';

function WeekSelect(props): JSX.Element {
  const weekOptions = {
    7: intl.get('CronSelect.SUN'),
    1: intl.get('CronSelect.MON'),
    2: intl.get('CronSelect.TUE'),
    3: intl.get('CronSelect.WED'),
    4: intl.get('CronSelect.THU'),
    5: intl.get('CronSelect.FRI'),
    6: intl.get('CronSelect.SAT'),
  };

  useEffect(() => {
    intl.load(locales);
  }, []);

  return (
    <Select size="small" {...props}>
      {Object.entries(weekOptions).map(([weekCode, weekName]) => {
        return (
          <Select.Option key={weekCode} value={weekCode}>
            {weekName}
          </Select.Option>
        );
      })}
    </Select>
  );
}

export default WeekSelect;
