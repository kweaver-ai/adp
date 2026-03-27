import intl from 'react-intl-universal';
import { EllipsisOutlined } from '@ant-design/icons';
import { Dropdown, type MenuProps } from 'antd';
import dayjs from 'dayjs';
import ObjectIcon from '@/components/ObjectIcon';
import * as KnowledgeNetworkType from '@/services/knowledgeNetwork/type';
import { Button } from '@/web-library/common';
import styles from '../index.module.less';

type OperationMenuItem = Required<MenuProps>['items'][number];

type Props = {
  record: KnowledgeNetworkType.KnowledgeNetwork;
  checked: boolean;
  dropdownItems: OperationMenuItem[];
  onClick: (record: KnowledgeNetworkType.KnowledgeNetwork) => void;
  onCheckChange: (record: KnowledgeNetworkType.KnowledgeNetwork, checked: boolean) => void;
  onOperate: (key: string, record: KnowledgeNetworkType.KnowledgeNetwork) => void;
};

const KnowledgeNetWorkCard = ({ record, checked, dropdownItems, onClick, onCheckChange: _onCheckChange, onOperate }: Props) => {
  const hasCustomIcon = !!record.icon;
  const placeholderText = (record.name || '?').trim().charAt(0).toUpperCase() || '?';

  return (
    <div className={`${styles['network-card']} ${checked ? styles['network-card-selected'] : ''}`.trim()} onClick={() => onClick(record)}>
      <div className={styles['network-card-header']}>
        <div className={styles['network-card-title']} title={record.name}>
          {hasCustomIcon ? (
            <ObjectIcon icon={record.icon} color={record.color || '#E5E8EF'} borderRadius={12} size={48} iconSize={22} />
          ) : (
            <div className={styles['network-card-placeholder-icon']}>{placeholderText}</div>
          )}
          <div className={styles['network-card-title-text']}>
            <div className={`g-ellipsis-1 ${styles['network-card-name']}`}>{record.name}</div>
            <div className={styles['network-card-description']} title={record.comment}>
              {record.comment || intl.get('Global.noComment')}
            </div>
          </div>
        </div>

        <div className={styles['network-card-actions']}>
          {dropdownItems.length > 0 && (
            <Dropdown
              trigger={['click']}
              menu={{
                items: dropdownItems,
                onClick: ({ key, domEvent }) => {
                  domEvent.stopPropagation();
                  onOperate(String(key), record);
                },
              }}
            >
              <Button.Icon
                className={styles['network-card-more']}
                icon={<EllipsisOutlined style={{ fontSize: 20 }} />}
                onClick={(event) => event.stopPropagation()}
              />
            </Dropdown>
          )}
        </div>
      </div>
      <div className={styles['network-card-footer']}>
        <span className="g-ellipsis-1">
          {intl.get('Global.modifier')}：{record?.updater?.name || '--'}
        </span>
        <span>
          {intl.get('Global.updateTime')}：{record.update_time ? dayjs(record.update_time).format('YYYY/MM/DD HH:mm') : '--'}
        </span>
      </div>
    </div>
  );
};

export default KnowledgeNetWorkCard;
