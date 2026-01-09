/**
 * @file 基于 BaseCronSelect 封装的选择器（输入框 + 气泡面板）组合版
 */
import React, { useEffect } from 'react';
import intl from 'react-intl-universal';
import { EditOutlined } from '@ant-design/icons';
import { useControllableValue } from 'ahooks';
import { Input, Popover } from 'antd';
import BaseCronSelect from './BaseCronSelect';
import styles from './index.module.less';
import locales from './locales';
import type { CronSelectProps } from './types';

const CronSelect = React.memo<CronSelectProps>((props): JSX.Element => {
  const { inputProps, cronSelectProps, value, onChange } = props;
  const [i18nLoaded, setI18nLoaded] = React.useState(false);

  useEffect(() => {
    // 加载国际化文件，完成后更新状态触发重新渲染
    intl.load(locales);
    setI18nLoaded(true);
  }, []);

  const [visible, setVisible] = useControllableValue(props, {
    defaultValue: false,
    valuePropName: 'visible',
    trigger: 'onVisibleChange',
  });

  const handleClose = (): void => setVisible(false);

  const content = (
    <BaseCronSelect
      {...cronSelectProps}
      value={value}
      onChange={(value): void => {
        onChange && onChange(value);
        setVisible(false);
      }}
      onClose={handleClose}
    />
  );

  return (
    <Input
      {...inputProps}
      value={value}
      onChange={(e): void => {
        onChange && onChange(e.target.value);
      }}
      allowClear
      placeholder={intl.get('CronSelect.pleaseInput')}
      addonAfter={
        <Popover destroyOnHidden content={i18nLoaded ? content : null} trigger="click" open={visible} overlayClassName={styles['ar-cron-wrapper']}>
          <EditOutlined onClick={() => setVisible(!visible)} />
        </Popover>
      }
    />
  );
});

export default CronSelect;
