import { memo, useEffect, useState } from 'react';
import intl from 'react-intl-universal';
import { InputNumber } from 'antd';
import styles from '../index.module.less';
import locales from '../locales';
import { Item } from '../type';

const NumberItem = memo(
  ({
    value,
    onChange,
    disabled,
    validateValueError,
  }: {
    value: Item;
    onChange: (item: any) => void;
    validateValueError: (val: any) => void;
    disabled: boolean;
  }) => {
    const [i18nLoaded, setI18nLoaded] = useState(false);

    useEffect(() => {
      // 加载国际化文件，完成后更新状态触发重新渲染
      intl.load(locales);
      setI18nLoaded(true);
    }, []);

    const handleFromChange = (val: any): void => {
      validateValueError(val);
      onChange({
        ...value,
        value: {
          from: val,
          to: value.value?.to,
        },
      });
    };

    const handleValueChange = (val: any): void => {
      validateValueError(val);
      onChange({ ...value, value: val });
    };

    const handleToChange = (val: any): void => {
      validateValueError(val);
      onChange({
        ...value,
        value: {
          to: val,
          from: value.value?.from,
        },
      });
    };

    return (
      <>
        {value?.operation === 'range' || value?.operation === 'out_range' ? (
          <div className={styles['range-wrapper']}>
            <InputNumber value={value?.value?.from} onChange={handleFromChange} disabled={disabled} placeholder={intl.get('DataFilter.pleaseInputValue')} />
            <span className={styles['split-space']}>-</span>
            <InputNumber value={value?.value?.to} onChange={handleToChange} disabled={disabled} placeholder={intl.get('DataFilter.pleaseInputValue')} />
          </div>
        ) : (
          <InputNumber onChange={handleValueChange} value={value?.value} disabled={disabled} placeholder={intl.get('DataFilter.pleaseInputValue')} />
        )}
      </>
    );
  }
);

export default NumberItem;
