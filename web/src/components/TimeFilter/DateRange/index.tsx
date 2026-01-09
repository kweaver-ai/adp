/** 时间段选择 */
import { useState, useEffect } from 'react';
import intl from 'react-intl-universal';
import { DatePicker } from 'antd';
import dayjs from 'dayjs';

const DATE_FORMAT = 'YYYY-MM-DD HH:mm:ss'; // 日期格式化

const DateRange = (props: any) => {
  const { timeRange, onFilterChange } = props;
  const [timeValue, setTimeValue] = useState<any>([null, null]);
  const [i18nLoaded, setI18nLoaded] = useState(false);

  useEffect(() => {
    // 加载国际化文件，完成后更新状态触发重新渲染
    intl.load(require('../locales').default);
    setI18nLoaded(true);
  }, []);

  useEffect(() => {
    if (timeRange?.label?.indexOf('-') < 0) {
      setTimeValue([null, null]);
    } else {
      setTimeValue([timeRange?.value?.[0], timeRange?.value?.[1]]);
    }
  }, [timeRange?.value?.[0], timeRange?.value?.[1]]);

  // 应用日期范围
  const applyDateRange = (date: any) => {
    const startTime = dayjs(date[0]).format(DATE_FORMAT);
    const endTime = dayjs(date[1]).format(DATE_FORMAT);

    const timeRange = { label: `${startTime} - ${endTime}`, value: date };
    setTimeValue(date);
    if (!!date[0] && !!date[1]) onFilterChange(timeRange);
  };

  return (
    <DatePicker.RangePicker
      value={timeValue}
      format={DATE_FORMAT}
      showTime
      allowClear
      placeholder={[intl.get('Global.startTime'), intl.get('Global.endTime')]}
      onOk={applyDateRange}
    />
  );
};

export default DateRange;
