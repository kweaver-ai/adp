import { useEffect } from 'react';
import { Select } from '@/web-library/common';

const DateCurrent = (props: any) => {
  const { value, onChange } = props;

  useEffect(() => {
    if (!value) onChange('day');
  }, []);

  const options = [
    { value: 'year', label: '年' },
    { value: 'month', label: '月' },
    { value: 'week', label: '周' },
    { value: 'day', label: '天' },
    { value: 'hour', label: '小时' },
    { value: 'minute', label: '分钟' },
  ];

  return <Select defaultValue="day" options={options} value={value} onChange={onChange} />;
};

export default DateCurrent;
