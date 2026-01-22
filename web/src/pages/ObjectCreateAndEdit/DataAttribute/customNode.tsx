import { useMemo, useState } from 'react';
import intl from 'react-intl-universal';
import { EllipsisOutlined, PlusOutlined, SearchOutlined } from '@ant-design/icons';
import { Handle, Position } from '@xyflow/react';
import { Dropdown, Input, Tooltip } from 'antd';
import FieldTypeIcon from '@/components/FieldTypeIcon';
import ObjectIcon from '@/components/ObjectIcon';
import * as OntologyObjectType from '@/services/object/type';
import { IconFont, Button } from '@/web-library/common';
import styles from './index.module.less';

const CustomNode = ({ data, id }: OntologyObjectType.TNode) => {
  const isViewNode = id === 'view';
  const [searchVal, setSearchVal] = useState('');

  const filteredAttributes = useMemo(
    () => data.attributes.filter((attr) => (searchVal ? attr.name.toLowerCase().includes(searchVal.toLowerCase()) : true)),
    [data, searchVal]
  );

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchVal(e.target.value);
  };

  const onOperate = (key: string) => {
    if (key === 'selectView') {
      data.openDataViewSource?.();
    }
    if (key === 'deleteView') {
      data.deleteDataViewSource?.();
    }
  };

  const dropdownMenu = [
    {
      key: 'selectView',
      label: intl.get('Object.replaceDataView'),
    },
    {
      key: 'deleteView',
      label: intl.get('Object.clearDataView'),
    },
  ];

  return (
    <div className={styles['panel-list']}>
      <div className={styles['panel-header']}>
        <div className={styles['panel-title-box']}>
          {id === 'data' ? (
            <ObjectIcon icon={data.icon} color={data.bg} size={20} iconSize={16} />
          ) : (
            <IconFont type={data.icon} style={{ color: '#000', fontSize: '20px' }} />
          )}
          <div className={`${styles['panel-title']} g-ellipsis-1`}>{data.label}</div>
          <div className={styles['panel-count']}>{data.attributes.length}</div>
        </div>
        <div className={styles['panel-actions']}>
          {id === 'view' && (
            <>
              {data.attributes.length > 0 ? (
                <>
                  <Tooltip title={intl.get('Object.addFieldAsNewDataAttribute')}>
                    <Button.Icon icon={<IconFont type="icon-dip-pickup" style={{ fontSize: 18 }} />} onClick={() => data.pickAttribute?.()} />
                  </Tooltip>
                  <Tooltip title={intl.get('Object.smartMatchingConnection')}>
                    <Button.Icon icon={<IconFont type="icon-dip-auto-line" style={{ fontSize: 18 }} />} onClick={() => data.autoLine?.()} />
                  </Tooltip>
                  <Dropdown
                    menu={{
                      items: dropdownMenu,
                      onClick: (event: any) => {
                        event.domEvent.stopPropagation();
                        onOperate(event?.key);
                      },
                    }}
                  >
                    <Button.Icon icon={<EllipsisOutlined style={{ fontSize: 18 }} />} onClick={(event) => event.stopPropagation()} />
                  </Dropdown>
                </>
              ) : (
                <Tooltip title={intl.get('Global.chooseDataView')}>
                  <PlusOutlined
                    style={{ color: 'rgb(18, 110, 227)', fontSize: '16px', cursor: 'pointer' }}
                    onClick={(e) => {
                      e.stopPropagation();
                      data.openDataViewSource?.();
                    }}
                  />
                </Tooltip>
              )}
            </>
          )}
          {id === 'data' && (
            <Tooltip title={intl.get('Object.addDataAttribute')}>
              <PlusOutlined
                style={{ color: 'rgb(18, 110, 227)', fontSize: '16px', cursor: 'pointer' }}
                onClick={(e) => {
                  e.stopPropagation();
                  data.addDataAttribute?.();
                }}
              />
            </Tooltip>
          )}
        </div>
      </div>
      {filteredAttributes.length > 0 && (
        <div className={styles['panel-search']}>
          <Input placeholder={intl.get('Global.searchProperty')} value={searchVal} suffix={<SearchOutlined />} onChange={handleSearch} allowClear />
        </div>
      )}
      <div className={styles['panel-content']}>
        {filteredAttributes.map((attr) => (
          <div
            key={attr.name}
            className={styles['panel-item']}
            style={{ cursor: isViewNode ? 'default' : 'pointer' }}
            onClick={() => !isViewNode && data.attrClick?.(attr)}
          >
            <div className={styles['item-content']}>
              <FieldTypeIcon type={attr.type} />
              <div>
                <div className={styles['item-name']}>
                  <span className={styles['item-name-text']}>{attr.display_name}</span>
                  {attr.comment && (
                    <Tooltip title={attr.comment}>
                      <IconFont type="icon-dip-color-comment" />
                    </Tooltip>
                  )}
                </div>
                <div className={styles['item-tech-name']}>{attr.name}</div>
              </div>
            </div>

            {!isViewNode && (
              <div className={styles['item-icons']}>
                {attr.primary_key && <IconFont type="icon-dip-color-primary-key" />}
                {attr.display_key && <IconFont type="icon-dip-color-star" />}
                {attr.incremental_key && <IconFont type="icon-dip-color-increment" />}
              </div>
            )}
            {isViewNode ? (
              <>
                <Handle type="source" position={Position.Right} id={`${id}-${attr.name}`} className={styles['panel-handle']} />
                <Handle type="target" position={Position.Right} id={`${id}-${attr.name}`} className={styles['panel-handle']} />
              </>
            ) : (
              <>
                <Handle type="source" position={Position.Left} id={`${id}-${attr.name}`} className={styles['panel-handle']} />
                <Handle type="target" position={Position.Left} id={`${id}-${attr.name}`} className={styles['panel-handle']} />
              </>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};

export default CustomNode;
