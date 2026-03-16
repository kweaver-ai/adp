import { useState } from 'react';
import { Input, Button, Popover, InputNumber } from 'antd';
import { SettingOutlined } from '@ant-design/icons';
import intl from 'react-intl-universal';
import styles from '../index.module.less';

interface KNNItemProps {
  value: any;
  disabled?: boolean;
  onChange: (value: any) => void;
}

const KNNItem = ({ value, disabled = false, onChange }: KNNItemProps) => {
  // 确保 value 结构正确
  const currentValue = {
    value: value?.value || '',
    limit_value: value?.limit_value || 3000,
    limit_key: 'k',
  };

  const handleValueChange = (e: any) => {
    onChange({
      ...currentValue,
      value: e.target.value,
    });
  };

  const handleLimitValueChange = (limit_value: any) => {
    onChange({
      ...currentValue,
      limit_value,
    });
  };

  return (
    <div className={styles['knn-wrapper']}>
      <Input
        value={currentValue.value}
        disabled={disabled}
        onChange={handleValueChange}
        placeholder={intl.get('DataFilterNew.pleaseInputValue')}
        style={{ width: '180px', marginRight: '8px' }}
      />
      <Popover
        content={
          <div className={styles['knn-popover']}>
            <InputNumber value={currentValue.limit_value} onChange={handleLimitValueChange} min={1} max={10000} style={{ width: '160px' }} />
          </div>
        }
        title={intl.get('DataFilterNew.top_k_results')}
        trigger="click"
      >
        <Button type="text" icon={<SettingOutlined />} disabled={disabled} size="small" />
      </Popover>
    </div>
  );
};

export default KNNItem;