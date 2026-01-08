import { useEffect } from 'react';
import intl from 'react-intl-universal';
import { Divider } from 'antd';
import classNames from './classNames';
import styles from './index.module.less';
import locales from './locales';

const cs = classNames.bind(styles);

interface Props {
  value?: 'and' | 'or';
  onChange?: (value: any) => void;
  className?: string;
  disabled?: boolean;
}

const LogicalOperation = ({ value = 'and', onChange, className, disabled = false }: Props): JSX.Element => {
  useEffect(() => {
    intl.load(locales);
  }, []);

  const onClick = (): void => {
    onChange && !disabled && onChange(value === 'and' ? 'or' : 'and');
  };

  return (
    <div className={cs('logical-content', { 'logical-content-disabled': disabled }, className ?? '')}>
      <Divider type="vertical" />
      <div onClick={onClick} className={cs('logical-select', { 'logical-select-disabled': disabled })}>
        {intl.get(`DataFilter.${value}`)}
      </div>
      <Divider type="vertical" />
    </div>
  );
};

export default LogicalOperation;
