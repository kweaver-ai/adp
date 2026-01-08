import { useEffect } from 'react';
import intl from 'react-intl-universal';
import { Divider } from 'antd';
import classNames from './classNames';
import styles from './index.module.less';
import locales from './locales';

const cs = classNames.bind(styles);

interface Props {
  value?: 'and' | 'or';
  className?: string;
}

const LogicalOperation = ({ value = 'and', className }: Props): JSX.Element => {
  useEffect(() => {
    intl.load(locales);
  }, []);

  return (
    <div className={cs('logical-content-detail', 'logical-content-disabled', className ?? '')}>
      <Divider type="vertical" />
      <div className={cs('logical-select-detail', 'logical-select-disabled')}>{intl.get(`DataFilter.${value}`)}</div>
      <Divider type="vertical" />
    </div>
  );
};

export default LogicalOperation;
