import { forwardRef, useMemo, useCallback, useImperativeHandle, useEffect } from 'react';
import intl from 'react-intl-universal';
import { MinusOutlined } from '@ant-design/icons';
import { InputNumber, Select } from 'antd';
import { map } from 'lodash-es';
import styles from './index.module.less';
import locales from './locales';

interface ResultFilterProps {
  value?: {
    operation: string;
    value: any;
    field: string;
  };
  onChange?: (value: { operation: string; value: any; field: string }) => void;
  layout?: 'horizontal' | 'vertical';
}

const operateOption = ['==', '!=', '<', '<=', '>', '>=', 'in', 'not_in', 'range', 'out_range'];

// 右侧值为数组的操作符
const aryOperation = ['in', 'not_in'];

// 右侧值为范围的操作符
const rangeOperation = ['range', 'out_range'];

const DEFAULT_VALUE = { operation: '==', value: '', field: '__value' };

const ResultFilter = forwardRef((props: ResultFilterProps, ref) => {
  const { value = DEFAULT_VALUE, onChange, layout = 'horizontal' } = props;

  useEffect(() => {
    intl.load(locales);
  }, []);

  const handleChangeOperation = useCallback(
    (e: string) => {
      onChange?.({ operation: e, value: '', field: value.field });
    },
    [onChange, value.field]
  );

  const handleValueChange = useCallback(
    (e: any) => {
      onChange?.({ operation: value.operation, value: e, field: value.field });
    },
    [onChange, value.operation, value.field]
  );

  const validate = useCallback(() => {
    if (!value?.operation) {
      return false;
    }

    if (aryOperation.includes(value?.operation)) {
      return Array.isArray(value.value) && value.value.length > 0;
    }

    return value.value !== '' && value.value !== null && value.value !== undefined;
  }, [value]);

  useImperativeHandle(ref, () => ({
    validate,
  }));

  const ValueInput = useMemo(() => {
    if (!value?.operation) {
      return null;
    }

    if (aryOperation.includes(value?.operation)) {
      return (
        <Select
          style={{ width: '100%' }}
          mode="tags"
          value={!!value?.value ? value?.value : undefined}
          onChange={(e) => {
            const values = e
              .map((item: any) => {
                return Number.parseFloat(item);
              })
              .filter((item: any) => !isNaN(item));
            handleValueChange(values);
          }}
          placeholder={intl.get('ResultFilter.pleaseInputValue')}
        />
      );
    } else if (rangeOperation.includes(value?.operation)) {
      return (
        <div className={styles['result-filter-range']}>
          <InputNumber
            value={value?.value?.[0]}
            onChange={(e) => {
              handleValueChange([e, ...(value?.value || []).slice(1)]);
            }}
            style={{ width: '50%' }}
          />
          <MinusOutlined style={{ color: 'rgba(0, 0, 0, 0.3)' }} />
          <InputNumber
            value={value?.value?.[1]}
            onChange={(e) => {
              handleValueChange([...(value?.value || []).slice(0, 1), e]);
            }}
            style={{ width: '50%' }}
          />
        </div>
      );
    } else {
      return (
        <InputNumber
          autoFocus
          style={{ width: '100%' }}
          value={value?.value}
          onChange={handleValueChange}
          placeholder={intl.get('ResultFilter.pleaseInputValue')}
        />
      );
    }
  }, [value, handleValueChange]);

  return layout === 'horizontal' ? (
    <div className={styles['result-filter-horizontal']}>
      <Select
        style={{ width: 160, flexShrink: 0 }}
        value={value?.operation}
        placeholder="请选择"
        onChange={handleChangeOperation}
        options={map(operateOption, (item: string) => ({ value: item, label: intl.get(`ResultFilter.${item}`) }))}
      />
      {ValueInput}
    </div>
  ) : (
    <div className={styles['result-filter-vertical']}>
      <Select
        style={{ width: '100%', marginBottom: 8 }}
        value={value?.operation}
        placeholder="请选择"
        onChange={handleChangeOperation}
        options={map(operateOption, (item: string) => ({ value: item, label: intl.get(`ResultFilter.${item}`) }))}
      />
      {ValueInput}
    </div>
  );
});

export default ResultFilter;
