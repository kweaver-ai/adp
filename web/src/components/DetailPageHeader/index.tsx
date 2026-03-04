import { ReactNode } from 'react';
import { LeftOutlined } from '@ant-design/icons';
import styles from './index.module.less';

interface DetailPageHeaderProps {
  title: ReactNode;
  icon?: ReactNode;
  actions?: ReactNode;
  onBack?: () => void;
}

const DetailPageHeader = (props: DetailPageHeaderProps) => {
  const { title, icon, actions, onBack } = props;

  return (
    <div className={styles['detail-page-header']}>
      <div className={styles['detail-page-header-left']}>
        <LeftOutlined className={styles['detail-page-header-back']} onClick={onBack} />
        {icon}
        <span className={styles['detail-page-header-title']}>{title}</span>
      </div>
      <div className={styles['detail-page-header-actions']}>{actions}</div>
    </div>
  );
};

export default DetailPageHeader;
