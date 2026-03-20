import { forwardRef, useImperativeHandle, useMemo, useState, useEffect } from 'react';
import intl from 'react-intl-universal';
import { useUpdateEffect } from 'ahooks';
import { DatePicker, Input, Select } from 'antd';
import classNames from 'classnames';
import { groupBy, map, filter } from 'lodash-es';
import ObjectSelector from '@/components/ObjectSelector';
import DateBefore from './components/DateBefore';
import DateBetween from './components/DateBetween';
import DateCurrent from './components/DateCurrent';
import NumberItem from './components/NumberItem';
import MultiMatchItem from './components/MultiMatchItem';
import KNNItem from './components/KNNItem';
import styles from './index.module.less';
import locales from './locales';
import { Item } from './type';
import { defaultTypeOption } from './utils';
import moment from 'dayjs';

const typeLabels: any = {
  number: '数值',
  string: '字符串',
  boolean: '布尔',
  date: '时间类型',
};

// 右侧值为数组的操作符
const aryOperation = ['in', 'not_in', 'contain', 'not_contain'];
/**
 * 有效的值来源
 * @returns 常量  const ，value内容即需要比较的值
 * @returns 字段  field，value内容为字段名称，意思是比较两个字段的内容
 * @returns 用户 user，value内容为当前用户的某个属性字段，意思是取当前用户的某个属性字段的值作为比较的值
 */
const valueForms = ['const'];

interface DataFilterItemProps {
  objectOptions: any[];
  disabled?: boolean;
  value: Item;
  onChange: (Item: any) => void;
  transformType?: (string: any) => string;
  typeOption?: { [key: string]: string[] };
  required?: boolean;
}

interface ObjectType {
  id: string;
  name: string;
  data_properties: Array<{
    name: string;
    display_name: string;
    type: string;
  }>;
}

const DataFilterItem = forwardRef(
  ({ objectOptions, value, disabled = false, transformType, typeOption = defaultTypeOption, onChange, required: defaultRequired = true }: DataFilterItemProps, ref) => {
    // 对象类
    const [objectTarget, setObjectTarget] = useState<ObjectType | undefined>(undefined);

    // 构建属性选项
    const fields = useMemo(() => {
      const dataProperties = objectTarget?.data_properties.map((property: any) => ({
        ...property,
        formateTypeLabel: typeLabels[transformType?.(property.type) || property.type] || '',
      }));
      return groupBy(dataProperties || [], 'formateTypeLabel');
    }, [objectTarget?.data_properties, transformType]);

    const [fieldType, setFieldType] = useState<string>('');
    const [errors, setErrors] = useState<{ name: string; value: string }>({ name: '', value: '' });

    const [i18nLoaded, setI18nLoaded] = useState(false);

    useEffect(() => {
      // 加载国际化文件，完成后更新状态触发重新渲染
      intl.load(locales);
      setI18nLoaded(true);
    }, []);

    const fieldListFilter = (val: any): any => {
      return objectTarget?.data_properties.find((i) => (i.name && i.name === val) || i.display_name === val);
    };

    const formatType = useMemo(() => {
      return transformType ? transformType(fieldType || 'number') : fieldType;
    }, [fieldType]);

    const isEmpty = (value: any): boolean => (typeof value !== 'number' && !value) || (Array.isArray(value) && !value.length);

    const validateValue = (value: any, required = defaultRequired): { value?: string } => {
      const error: { value?: string } = {};

      if (required && (isEmpty(value) || (Array.isArray(value) && value.length === 2 && (isEmpty(value[0]) || isEmpty(value[1]))))) {
        error.value = intl.get('DataFilterNew.valueCannotEmpty');
      } else {
        error.value = '';
      }

      return error;
    };

    const validateField = (value: any, required = defaultRequired): { name?: string } => {
      const error: { name?: string } = {};

      if (!value && required) {
        error.name = intl.get('DataFilterNew.fieldCannotEmpty');
      } else if (value && !objectTarget?.data_properties.find((i) => (i.name || i.display_name) === value)) {
        error.name = intl.get('DataFilterNew.fieldsNotExist');
      } else {
        error.name = '';
      }

      return error;
    };

    const validate = (required: any): boolean => {
      const fieldError = validateField(value.field, required);
      const valueError = validateValue(value.value, required);

      // 存在和不存在, 值空, 值非空 没有 value 字段
      if (
        value.operation === 'exist' ||
        value.operation === 'not_exist' ||
        value.operation === 'not_empty' ||
        value.operation === 'empty' ||
        value.operation === 'null' ||
        value.operation === 'not_null'
      ) {
        setErrors({ name: fieldError.name || '', value: '' });
        return !!fieldError?.name;
      }

      setErrors({ name: fieldError.name || '', value: valueError?.value || '' });

      return !!(fieldError?.name || valueError?.value);
    };

    const validateValueError = (val: any): void => {
      setErrors({ ...errors, ...validateValue(val) });
    };

    useImperativeHandle(ref, () => ({ validate }));

    useUpdateEffect(() => {
      if ((formatType === 'number' && value?.value?.[0]) || (formatType === 'number' && (value?.operation === 'range' || value?.operation === 'out_range'))) {
        onChange({ ...value, value: null });
      }

      if (value?.field === undefined && value?.value === undefined) {
        setFieldType('number');
      }
    }, [value?.operation]);

    /** 更换对象类 */
    const handleChangeObject = (val: string, objectTarget: ObjectType): void => {
      if (value.object_type_id !== val) {
        // 清空属性、操作符、值
        onChange({ ...value, object_type_id: val, operation: undefined, field: undefined, value: undefined });
      } else {
        onChange({ ...value, object_type_id: val });
      }

      setObjectTarget(objectTarget);
    };

    /** 更换属性 */
    const handleChangeField = (val: any): void => {
      setErrors({ ...errors, ...validateField(val) });
      const type = fieldListFilter(val)?.type || '';
      const formatType = transformType?.(type) || type;

      onChange({
        ...value,
        operation: typeOption[formatType].includes(value?.operation) ? value.operation : typeOption[formatType][0],
        field: val,
        value_from: 'const',
        value: type !== fieldType ? undefined : value?.value,
      });

      setFieldType(type);
    };

    /** 更换操作符 */
    const handleChangeOperation = (val: any): void => {
      // 创建新对象，只保留必要字段，清除多余字段
      const { field, object_type_id } = value;
      const newData: any = { field, object_type_id, operation: val, value_from: 'const', value: undefined };
      if (val === 'true') newData.value = true;
      if (val === 'false') newData.value = false;
      onChange(newData);
      if (val === 'exist' || val === 'not_exist' || val === 'not_empty' || val === 'empty' || val === 'null' || val === 'not_null') {
        setErrors((item) => ({ ...item, value: '' }));
      }
    };

    const handleChangeValueFrom = (val: any): void => {
      onChange({ ...value, value_from: val });
    };

    /** 更新属性值 */
    const handleValueChange = (val: any): void => {
      setErrors({ ...errors, ...validateValue(val) });
      if (value.operation === 'multi_match' || value.operation === 'knn' || value.operation === 'before') {
        onChange({ ...value, ...(val || {}) });
      } else {
        onChange({ ...value, value: val });
      }
    };

    /** input 值变化 */
    const handleStringValueChange = (e: any): void => {
      setErrors({ ...errors, ...validateValue(e.target.value) });
      onChange({ ...value, value: e.target.value });
    };

    const renderItem = (formatType: string, operation: any): JSX.Element => {
      if (formatType === 'boolean') {
        return <></>;
      }
      if (
        operation === 'exist' ||
        operation === 'not_exist' ||
        operation === 'not_empty' ||
        operation === 'empty' ||
        operation === 'null' ||
        operation === 'not_null'
      ) {
        return <></>;
      }

      if (aryOperation.includes(operation)) {
        return (
          <Select
            mode="tags"
            value={value.value ?? undefined}
            disabled={disabled}
            onChange={(value) => {
              if (formatType === 'number') {
                value = map(value, (item) => {
                  const match = String(item).match(/-?\d+(\.\d+)?/);
                  if (match) return Number.parseFloat(item);
                  return '';
                });
              }
              value = filter(value, (item) => !!item);
              handleValueChange(value);
            }}
            placeholder={intl.get('DataFilterNew.pleaseInputValue')}
            style={{ width: '180px' }}
          />
        );
      }

      if (formatType === 'number') {
        return <NumberItem value={value} disabled={disabled} validateValueError={validateValueError} onChange={onChange} />;
      }

      if (formatType === 'date' && operation === 'before') {
        const { value: dateValue, unit } = value || {};
        return <DateBefore value={{ value: dateValue, unit }} onChange={handleValueChange} />;
      }
      if (formatType === 'date' && operation === 'current') {
        return <DateCurrent value={value.value ?? undefined} onChange={handleValueChange} />;
      }
      if (formatType === 'date' && (operation === 'between' || operation === 'out_range' || operation === 'range')) {
        return <DateBetween value={value.value ?? undefined} onChange={handleValueChange} />;
      }
      if (formatType === 'date') {
        return (
          <DatePicker
            showTime
            value={value?.value ? moment(value?.value) : undefined}
            onChange={(value) => {
              handleValueChange(moment(value).format('YYYY-MM-DD HH:mm:ss'));
            }}
            onOk={(value) => {
              handleValueChange(moment(value).format('YYYY-MM-DD HH:mm:ss'));
            }}
          />
        );
      }

      if (operation === 'multi_match') {
        const { fields, value: matchValue, match_type = 'best_fields' } = value || {};
        return (
          <MultiMatchItem
            value={{ fields, value: matchValue, match_type }}
            fieldList={objectTarget?.data_properties || []}
            disabled={disabled}
            onChange={handleValueChange}
          />
        );
      }

      if (operation === 'knn') {
        const { value: knnValue, limit_value = 3000, limit_key = 'k' } = value || {};
        return <KNNItem value={{ value: knnValue, limit_value, limit_key }} disabled={disabled} onChange={handleValueChange} />;
      }

      return <Input value={value?.value} disabled={disabled} onChange={handleStringValueChange} placeholder={intl.get('DataFilterNew.pleaseInputValue')} />;
    };

    const hasValue = 
      value.operation !== 'exist' &&
      value.operation !== 'not_exist' &&
      value.operation !== 'not_empty' &&
      value.operation !== 'empty' &&
      value.operation !== 'null' &&
      value.operation !== 'not_null';

    return (
      <div className={classNames(styles['filter-item'])}>
        <div className={styles['object-col']}>
          <ObjectSelector objectOptions={objectOptions} value={value.object_type_id} onChange={handleChangeObject} disabled={disabled} />
        </div>
        <div className={classNames(styles['field-col'], { [styles['error-item']]: !!errors?.name })}>
          <Select
            showSearch
            value={value?.field}
            disabled={disabled}
            placeholder={intl.get('DataFilterNew.pleaseSelectValue')}
            getPopupContainer={(triggerNode): HTMLElement => triggerNode.parentNode as HTMLElement}
            onChange={handleChangeField}
            options={map(Object.keys(fields), (key) => {
              return {
                label: key,
                title: key,
                options: map(fields?.[key], (item) => {
                  const { name, display_name } = item;
                  return { value: name, label: display_name };
                }),
              };
            })}
          />
          {errors?.name ? <div className={styles['error-tip']}>{errors?.name}</div> : <></>}
        </div>

        <div className={classNames(styles['operation-col'])}>
          <Select
            value={value?.operation}
            disabled={disabled}
            placeholder="请选择"
            onChange={handleChangeOperation}
            options={map(typeOption[formatType], (item) => ({ value: item, label: intl.get(`DataFilterNew.${item}`) }))}
          />
        </div>

        {hasValue && (
          <div className={classNames(styles['operation-col'])}>
            <Select
              value={valueForms[0]}
              disabled={disabled}
              onChange={handleChangeValueFrom}
              options={map(valueForms, (item) => ({ value: item, label: intl.get(`DataFilterNew.${item}`) }))}
            />
          </div>
        )}

        {hasValue && (
          <div className={classNames(styles['value-col'], { [styles['error-item']]: !!errors?.value })}>
            {renderItem(formatType, value.operation)}
            {errors?.value ? <div className={styles['error-tip']}>{errors?.value}</div> : <></>}
          </div>
        )}
      </div>
    );
  }
);

export default DataFilterItem;