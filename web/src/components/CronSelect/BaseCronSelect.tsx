/** 基于 Cron 封装的基础组件 */
import React, { useCallback, useRef, useEffect } from 'react';
import intl from 'react-intl-universal';
import { useControllableValue } from 'ahooks';
import { Button } from 'antd';
import classNames from 'classnames';
import { defaultTo } from 'lodash-es';
import Cron from './Cron';
import locales from './locales';
import type { CronFns } from './Cron/types';
import type { BaseCronSelectProps } from './types';

const BaseCronSelect = React.memo<BaseCronSelectProps>((props): JSX.Element => {
  const { className, style, defaultValue, onClose, ...restProps } = props;
  const cronRef = useRef<CronFns | null>(null);
  const [value, setValue] = useControllableValue<string>(props, {
    defaultValue,
  });

  useEffect(() => {
    intl.load(locales);
  }, []);

  const handleOk = useCallback(() => {
    if (cronRef.current) {
      const newValue = cronRef.current.getValue();
      setValue(newValue);
      props.onChange && props.onChange(newValue);
    }
  }, [setValue, props.onChange]);

  const getCronFns: BaseCronSelectProps['getCronFns'] = (fns): void => {
    cronRef.current = fns;
  };

  const footerContent = (): React.ReactNode => (
    <React.Fragment>
      <Button className="g-mr-2" type="default" onClick={onClose}>
        {intl.get('CronSelect.cancel')}
      </Button>
      <Button type="primary" onClick={handleOk}>
        {intl.get('CronSelect.confirm')}
      </Button>
    </React.Fragment>
  );

  return (
    <div className={classNames(className)} style={defaultTo(style, {})}>
      <Cron {...restProps} value={value} getCronFns={getCronFns} footer={footerContent()} />
    </div>
  );
});

export default BaseCronSelect;
