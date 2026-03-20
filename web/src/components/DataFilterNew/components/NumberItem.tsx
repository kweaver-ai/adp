import { memo, useEffect, useState } from 'react';
import intl from 'react-intl-universal';
import { InputNumber } from 'antd';
import styles from '../index.module.less';
import locales from '../locales';
import { Item } from '../type';

interface NumberItemProps {
  value: Item;
  onChange: (item: any) => void;
  validateValueError: (val: any) => void;
  disabled: boolean;
}

const NumberItem = (props: NumberItemProps) => {
  const { value, onChange, validateValueError, disabled } = props;
  const [i18nLoaded, setI18nLoaded] = useState(false);

  useEffect(() => {
    // 加载国际化文件，完成后更新状态触发重新渲染
    intl.load(locales);
    setI18nLoaded(true);
  }, []);

  const handleFromChange = (val: any): void => {
    validateValueError([val, value.value?.[1]]);
    onChange({
      ...value,
      value: [val, value.value?.[1]],
    });
  };

  const handleValueChange = (val: any): void => {
    validateValueError(val);
    onChange({ ...value, value: val });
  };

  const handleToChange = (val: any): void => {
    validateValueError([value.value?.[0], val]);
    onChange({
      ...value,
      value: [value.value?.[0], val],
    });
  };

  return (
    <>
      {value?.operation === 'range' || value?.operation === 'out_range' || value?.operation === 'between' ? (
        <div className={styles['range-wrapper']}>
          <InputNumber value={value?.value?.[0]} onChange={handleFromChange} disabled={disabled} placeholder={intl.get('DataFilterNew.pleaseInputValue')} />
          <span className={styles['split-space']}>-</span>
          <InputNumber value={value?.value?.[1]} onChange={handleToChange} disabled={disabled} placeholder={intl.get('DataFilterNew.pleaseInputValue')} />
        </div>
      ) : (
        <InputNumber onChange={handleValueChange} value={value?.value} disabled={disabled} placeholder={intl.get('DataFilterNew.pleaseInputValue')} />
      )}
    </>
  );
};

export default memo(NumberItem);
