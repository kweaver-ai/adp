import { useEffect } from 'react';
import intl from 'react-intl-universal';
import { Space, InputNumber } from 'antd';
import { Select } from '../../../common';
import locales from '../locales';

const DateBefore = (props: any) => {
  const { value, onChange } = props;

  useEffect(() => {
    intl.load(locales);
  }, []);

  useEffect(() => {
    if (!value) onChange([1, 'millisecond']);
  }, []);

  const options = [
    { value: 'millisecond', label: intl.get('DataFilter.millisecond') },
    { value: 'second', label: intl.get('DataFilter.second') },
    { value: 'minute', label: intl.get('DataFilter.minute') },
    { value: 'hour', label: intl.get('DataFilter.hour') },
    { value: 'day', label: intl.get('DataFilter.day') },
    { value: 'week', label: intl.get('DataFilter.week') },
    { value: 'month', label: intl.get('DataFilter.month') },
    { value: 'quarter', label: intl.get('DataFilter.quarter') },
    { value: 'year', label: intl.get('DataFilter.year') },
  ];

  return (
    <Space.Compact>
      <InputNumber placeholder={intl.get('DataFilter.pleaseInput')} min={0} value={value?.[0]} onChange={(data) => onChange([data, value?.[1]])} />
      <Select defaultValue="millisecond" options={options} value={value?.[1]} onChange={(data) => onChange([value?.[0], data])} />
    </Space.Compact>
  );
};

export default DateBefore;
