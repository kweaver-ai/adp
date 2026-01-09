import { useEffect, useState } from 'react';
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
  const [i18nLoaded, setI18nLoaded] = useState(false);

  useEffect(() => {
    // 加载国际化文件，完成后更新状态触发重新渲染
    intl.load(locales);
    setI18nLoaded(true);
  }, []);

  return (
    <div className={cs('logical-content-detail', 'logical-content-disabled', className ?? '')}>
      <Divider type="vertical" />
      <div className={cs('logical-select-detail', 'logical-select-disabled')}>{intl.get(`DataFilterNew.${value}`)}</div>
      <Divider type="vertical" />
    </div>
  );
};

export default LogicalOperation;
