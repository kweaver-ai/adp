import { useEffect } from 'react';
import intl from 'react-intl-universal';
import { Select } from '../../../common';
import locales from '../locales';

const DateCurrent = (props: any) => {
  const { value, onChange } = props;

  useEffect(() => {
    intl.load(locales);
  }, []);

  useEffect(() => {
    if (!value) onChange('%Y-%m-%d');
  }, []);

  const options = [
    { value: '%Y', label: intl.get('DataFilter.year') },
    { value: '%Y-%m', label: intl.get('DataFilter.month') },
    { value: '%x-%v', label: intl.get('DataFilter.week') },
    { value: '%Y-%m-%d', label: intl.get('DataFilter.day') },
    { value: '%Y-%m-%d %H', label: intl.get('DataFilter.hour') },
    { value: '%Y-%m-%d %H:%i', label: intl.get('DataFilter.minute') },
  ];

  return <Select defaultValue="%Y-%m-%d" options={options} value={value} onChange={onChange} />;
};

export default DateCurrent;
