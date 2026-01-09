import { useEffect, useState } from 'react';
import intl from 'react-intl-universal';
import { Select } from '../../../common';
import locales from '../locales';

const DateCurrent = (props: any) => {
  const { value, onChange } = props;
  const [i18nLoaded, setI18nLoaded] = useState(false);

  useEffect(() => {
    // 加载国际化文件，完成后更新状态触发重新渲染
    intl.load(locales);
    setI18nLoaded(true);
  }, []);

  useEffect(() => {
    if (!value) onChange('%Y-%m-%d');
  }, []);

  // 国际化未加载完成时返回空数组，避免选项显示空白
  const options = i18nLoaded
    ? [
        { value: '%Y', label: intl.get('DataFilter.year') },
        { value: '%Y-%m', label: intl.get('DataFilter.month') },
        { value: '%x-%v', label: intl.get('DataFilter.week') },
        { value: '%Y-%m-%d', label: intl.get('DataFilter.day') },
        { value: '%Y-%m-%d %H', label: intl.get('DataFilter.hour') },
        { value: '%Y-%m-%d %H:%i', label: intl.get('DataFilter.minute') },
      ]
    : [];

  return <Select defaultValue="%Y-%m-%d" options={options} value={value} onChange={onChange} />;
};

export default DateCurrent;
