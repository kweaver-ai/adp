import { useState } from 'react';
import { Select, Input, Button, Popover } from 'antd';
import { SettingOutlined } from '@ant-design/icons';
import { map } from 'lodash-es';
import intl from 'react-intl-universal';
import styles from '../index.module.less';
import { FieldList } from '../type';

interface MultiMatchItemProps {
  value: any;
  fieldList: FieldList[];
  disabled?: boolean;
  onChange: (value: any) => void;
}

const MultiMatchItem = ({ value, fieldList, disabled = false, onChange }: MultiMatchItemProps) => {
  // 过滤出 text 和 string 类型的字段
  const textFields = fieldList?.filter((field) => field.type === 'text') || [];

  // 匹配类型选项
  const matchTypeOptions = [
    { value: 'best_fields', label: intl.get('DataFilter.best_fields') },
    { value: 'most_fields', label: intl.get('DataFilter.most_fields') },
    { value: 'cross_fields', label: intl.get('DataFilter.cross_fields') },
    { value: 'phrase', label: intl.get('DataFilter.phrase') },
    { value: 'phrase_prefix', label: intl.get('DataFilter.phrase_prefix') },
    { value: 'bool_prefix', label: intl.get('DataFilter.bool_prefix') },
  ];

  // 确保 value 结构正确
  const currentValue = {
    fields: value?.fields || [],
    value: value?.value || '',
    match_type: value?.match_type || 'best_fields',
  };

  const handleFieldsChange = (fields: any) => {
    onChange({
      ...currentValue,
      fields,
    });
  };

  const handleValueChange = (e: any) => {
    onChange({
      ...currentValue,
      value: e.target.value,
    });
  };

  const handleMatchTypeChange = (match_type: any) => {
    onChange({
      ...currentValue,
      match_type,
    });
  };

  return (
    <div className={styles['multi-match-wrapper']}>
      <Select
        mode="multiple"
        value={currentValue.fields}
        disabled={disabled}
        onChange={handleFieldsChange}
        placeholder={intl.get('DataFilter.pleaseSelectValue')}
        style={{ width: '180px', marginRight: '8px' }}
        maxTagCount={1}
        maxTagTextLength={5}
        options={map(textFields, (item) => ({
          value: item.name || item.display_name,
          label: item.display_name || item.name,
        }))}
      />
      <Input
        value={currentValue.value}
        disabled={disabled}
        onChange={handleValueChange}
        placeholder={intl.get('DataFilter.pleaseInputValue')}
        style={{ width: '80px', marginRight: '8px' }}
      />
      <Popover
        content={
          <div className={styles['match-type-popover']}>
            <Select value={currentValue.match_type} onChange={handleMatchTypeChange} style={{ width: '160px' }} options={matchTypeOptions} />
          </div>
        }
        title={intl.get('DataFilter.match_type')}
        trigger="click"
      >
        <Button type="text" icon={<SettingOutlined />} disabled={disabled} size="small" />
      </Popover>
    </div>
  );
};

export default MultiMatchItem;
