import { useState, type ReactNode } from 'react';
import intl from 'react-intl-universal';
import { ClockCircleOutlined, UserOutlined } from '@ant-design/icons';
import { Tag } from 'antd';
import dayjs from 'dayjs';
import styles from './index.module.less';

interface MetaItem {
  icon?: ReactNode;
  label: ReactNode;
  value: ReactNode;
  valueType?: 'text' | 'badge';
  emphasize?: boolean;
}

interface DetailSummaryCardProps {
  id?: string;
  icon?: ReactNode;
  name?: string;
  tags?: string[];
  comment?: string;
  metaLeftItems?: MetaItem[];
  modifier?: string;
  updateTime?: string | number;
  commentMaxLength?: number;
}

const DetailSummaryCard = (props: DetailSummaryCardProps) => {
  const { id, icon, name, tags, comment, metaLeftItems = [], modifier, updateTime, commentMaxLength = 140 } = props;
  const [descExpanded, setDescExpanded] = useState(false);

  const isLongComment = (comment || '').length > commentMaxLength;
  const displayComment = isLongComment && !descExpanded ? `${String(comment).slice(0, commentMaxLength)}...` : comment || '暂无描述';

  return (
    <div className={styles['detail-summary-card']}>
      <div className={styles['top-id-row']}>ID: {id || '--'}</div>
      <div className={styles['top-main-row']}>
        <div className={styles['object-item']} title={name || ''}>
          {icon}
          <div>
            <span className={styles['object-name']}>{name || '--'}</span>
            <div className={styles['tag-list']}>{Array.isArray(tags) && tags.length ? tags.map((tag) => <Tag key={tag}>{tag}</Tag>) : ''}</div>
          </div>
        </div>
      </div>
      <div className={styles['top-description']}>
        {displayComment}
        {isLongComment && (
          <span className={styles['more-link']} onClick={() => setDescExpanded((value) => !value)}>
            {descExpanded ? intl.get('Global.collapse') : intl.get('Global.more')}
          </span>
        )}
      </div>
      <div className={styles['top-divider']} />
      <div className={styles['top-meta-row']}>
        <div className={styles['meta-left']}>
          {metaLeftItems.map((item, index) => (
            <div key={`${index}`} className={`${styles['meta-item']} ${item.emphasize ? styles['meta-item-main'] : ''}`}>
              {item.icon}
              <span className={styles['meta-item-key']}>{item.label}:</span>
              {item.valueType === 'badge' ? <span className={styles['meta-value-badge']}>{item.value}</span> : <span>{item.value}</span>}
            </div>
          ))}
        </div>
        <div className={styles['meta-right']}>
          <div className={styles['meta-item']}>
            <UserOutlined className={styles['meta-icon']} />
            <span className={styles['meta-item-key']}>{intl.get('Global.modifier')}:</span>
            <span>{modifier || '--'}</span>
          </div>
          <div className={styles['meta-item']}>
            <ClockCircleOutlined className={styles['meta-icon']} />
            <span className={styles['meta-item-key']}>{intl.get('Global.updateTime')}:</span>
            <span>{updateTime ? dayjs(updateTime).format('YYYY-MM-DD HH:mm:ss') : '--'}</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default DetailSummaryCard;
