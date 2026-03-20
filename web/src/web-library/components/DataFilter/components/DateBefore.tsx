import { useEffect, useState } from 'react';
import intl from 'react-intl-universal';
import { Space, InputNumber } from 'antd';
import { Select } from '../../../common';
import locales from '../locales';

const DateBefore = (props: any) => {
  const { value, onChange } = props;
  const [i18nLoaded, setI18nLoaded] = useState(false);

  useEffect(() => {
    // 加载国际化文件，完成后更新状态触发重新渲染
    intl.load(locales);
    setI18nLoaded(true);
  }, []);

  // 确保 value 结构正确
  const currentValue = {
    value: value?.value || 1,
    unit: value?.unit || 'minute',
  };

  useEffect(() => {
    if (!value?.value) onChange(currentValue);
  }, []);

  const handleValueChange = (value: any) => {
    onChange({
      ...currentValue,
      value,
    });
  };

  const handleUnitChange = (unit: any) => {
    onChange({
      ...currentValue,
      unit,
    });
  };

  // 国际化未加载完成时返回空数组，避免选项显示空白
  const options = i18nLoaded
    ? [
        // { value: 'millisecond', label: intl.get('DataFilter.millisecond') },
        // { value: 'second', label: intl.get('DataFilter.second') },
        { value: 'minute', label: intl.get('DataFilter.minute') },
        { value: 'hour', label: intl.get('DataFilter.hour') },
        { value: 'day', label: intl.get('DataFilter.day') },
        { value: 'week', label: intl.get('DataFilter.week') },
        { value: 'month', label: intl.get('DataFilter.month') },
        // { value: 'quarter', label: intl.get('DataFilter.quarter') },
        { value: 'year', label: intl.get('DataFilter.year') },
      ]
    : [];

  return (
    <Space.Compact>
      <InputNumber placeholder={intl.get('DataFilter.pleaseInput')} min={0} value={currentValue.value} onChange={handleValueChange} />
      <Select defaultValue="minute" options={options} value={currentValue.unit} onChange={handleUnitChange} />
    </Space.Compact>
  );
};

export default DateBefore;
