import { useEffect, useMemo, useState } from 'react';
import intl from 'react-intl-universal';
import { EllipsisOutlined, InfoCircleFilled, PlusOutlined, SearchOutlined } from '@ant-design/icons';
import { Handle, Position } from '@xyflow/react';
import { Dropdown, Empty, Input, Tooltip } from 'antd';
import { showDeleteConfirm } from '@/components/DeleteConfirm';
import FieldTypeIcon from '@/components/FieldTypeIcon';
import ObjectIcon from '@/components/ObjectIcon';
import * as OntologyObjectType from '@/services/object/type';
import noSearchResultImage from '@/assets/images/common/no_search_result.svg';
import HOOKS from '@/hooks';
import { IconFont, Button } from '@/web-library/common';
import styles from './index.module.less';

const canPrimaryKeys = ['integer', 'unsigned integer', 'string'];
const canTitleKeys = ['integer', 'unsigned integer', 'string', 'text', 'float', 'decimal', 'date', 'time', 'datetime', 'timestamp', 'ip', 'boolean'];
const canIncrementalKeys = ['integer', 'unsigned integer', 'datetime', 'timestamp'];

const canBePrimaryKey = (type: string) => canPrimaryKeys.includes(type);
const canBeDisplayKey = (type: string) => canTitleKeys.includes(type);
const canBeIncrementalKey = (type: string) => canIncrementalKeys.includes(type);

const CustomNode = ({ data, id }: OntologyObjectType.TNode) => {
  const { modal } = HOOKS.useGlobalContext();
  const isViewNode = id === 'view';
  const [searchVal, setSearchVal] = useState('');
  const [hoveredAttr, setHoveredAttr] = useState<string | null>(null);
  const highlightedAttributes = data?.highlightedAttributes || [];

  // 监听 clearSearchTrigger 变化，清空搜索内容
  useEffect(() => {
    if (data.clearSearchTrigger !== undefined && data.clearSearchTrigger > 0) {
      setSearchVal('');
    }
  }, [data.clearSearchTrigger]);

  // 校验字段名是否符合规范
  const isValidName = (name: string) => {
    const namePattern = /^[a-z0-9][a-z0-9_-]*$/;
    return namePattern.test(name);
  };

  const filteredAttributes = useMemo(
    () => data.attributes.filter((attr) => (searchVal ? attr.name.toLowerCase().includes(searchVal.toLowerCase()) : true)),
    [data, searchVal]
  );

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchVal(e.target.value);
  };

  const handleDeleteAttr = (e: React.MouseEvent, attr: { name: string; display_name: string }) => {
    e.stopPropagation();
    data.deleteAttribute?.(attr.name);
  };

  const handleTogglePrimaryKey = (e: React.MouseEvent, attrName: string) => {
    e.stopPropagation();
    data.togglePrimaryKey?.(attrName);
  };

  const handleToggleDisplayKey = (e: React.MouseEvent, attrName: string) => {
    e.stopPropagation();
    data.toggleDisplayKey?.(attrName);
  };

  const handleToggleIncrementalKey = (e: React.MouseEvent, attrName: string) => {
    e.stopPropagation();
    data.toggleIncrementalKey?.(attrName);
  };

  const onOperate = (key: string) => {
    if (key === 'selectView') {
      data.openDataViewSource?.();
    }
    if (key === 'deleteView') {
      data.deleteDataViewSource?.();
    }
    if (key === 'add') {
      data.addDataAttribute?.();
    }
    if (key === 'pick') {
      data.pickAttribute?.();
    }
    if (key === 'clearAll') {
      showDeleteConfirm(modal, {
        content: intl.get('Object.clearAllDataAttributesConfirm'),
        okText: intl.get('Global.clear'),
        onOk: () => {
          data.clearAllAttributes?.();
        },
      });
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
            <>
              <Dropdown
                menu={{
                  items: [
                    {
                      key: 'add',
                      label: intl.get('Object.manualCreate'),
                    },
                    {
                      key: 'pick',
                      label: intl.get('Object.syncDataViewFields'),
                    },
                  ],
                  onClick: (event: any) => {
                    event.domEvent.stopPropagation();
                    onOperate(event?.key);
                  },
                }}
              >
                <PlusOutlined
                  style={{ color: 'rgb(18, 110, 227)', fontSize: '16px', cursor: 'pointer' }}
                  onClick={(e) => {
                    e.stopPropagation();
                  }}
                />
              </Dropdown>

              <Dropdown
                menu={{
                  items: [
                    {
                      key: 'clearAll',
                      label: intl.get('Global.clearAll'),
                    },
                  ],
                  onClick: (event: any) => {
                    event.domEvent.stopPropagation();
                    onOperate(event?.key);
                  },
                }}
              >
                <EllipsisOutlined
                  style={{ fontSize: '16px', cursor: 'pointer' }}
                  onClick={(e) => {
                    e.stopPropagation();
                  }}
                />
              </Dropdown>
            </>
          )}
        </div>
      </div>
      {(filteredAttributes.length > 0 || !!searchVal) && (
        <div className={styles['panel-search']}>
          <Input placeholder={intl.get('Global.searchProperty')} value={searchVal} suffix={<SearchOutlined />} onChange={handleSearch} allowClear />
        </div>
      )}
      <div className={styles['panel-content']}>
        {filteredAttributes.length > 0 ? (
          filteredAttributes.map((attr) => {
            const isHighlighted = highlightedAttributes.includes(attr.name);
            return (
              <div
                key={attr.name}
                className={`${styles['panel-item']} ${isHighlighted ? styles['panel-item-highlighted'] : ''}`}
                style={{ cursor: isViewNode ? 'default' : 'pointer' }}
                onClick={() => !isViewNode && data.attrClick?.(attr)}
                onMouseEnter={() => setHoveredAttr(attr.name)}
                onMouseLeave={() => setHoveredAttr(null)}
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
                    <div className={styles['item-tech-name']}>
                      <span className={styles['tech-name-text']}>{attr.name}</span>
                      {!isViewNode && !isValidName(attr.name) && (
                        <Tooltip title={intl.get('Global.idPatternError')}>
                          <InfoCircleFilled style={{ color: '#ff4d4f', fontSize: 12 }} />
                        </Tooltip>
                      )}
                    </div>
                  </div>
                </div>

                {!isViewNode && (
                  <div className={styles['item-icons']}>
                    {hoveredAttr === attr.name ? (
                      <>
                        {canBePrimaryKey(attr.type) && (
                          <Tooltip title={attr.primary_key ? intl.get('Global.cancelPrimaryKey') : intl.get('Global.setPrimaryKey')}>
                            <IconFont
                              type={attr.primary_key ? 'icon-dip-color-primary-key' : 'icon-dip-zhujian'}
                              className={styles['delete-icon']}
                              onClick={(e) => handleTogglePrimaryKey(e, attr.name)}
                            />
                          </Tooltip>
                        )}
                        {canBeDisplayKey(attr.type) && (
                          <Tooltip title={attr.display_key ? intl.get('Global.cancelTitle') : intl.get('Global.setTitle')}>
                            <IconFont
                              type={attr.display_key ? 'icon-dip-color-star' : 'icon-dip-biaoti'}
                              className={styles['delete-icon']}
                              onClick={(e) => handleToggleDisplayKey(e, attr.name)}
                            />
                          </Tooltip>
                        )}
                        {canBeIncrementalKey(attr.type) && (
                          <Tooltip title={attr.incremental_key ? intl.get('Global.cancelIncrementalKey') : intl.get('Global.setIncrementalKey')}>
                            <IconFont
                              type={attr.incremental_key ? 'icon-dip-color-increment' : 'icon-dip-zengliang'}
                              className={styles['delete-icon']}
                              onClick={(e) => handleToggleIncrementalKey(e, attr.name)}
                            />
                          </Tooltip>
                        )}
                        <Tooltip title={intl.get('Global.delete')}>
                          <IconFont type="icon-dip-trash" className={styles['delete-icon']} onClick={(e) => handleDeleteAttr(e, attr)} />
                        </Tooltip>
                      </>
                    ) : (
                      <>
                        {attr.primary_key && <IconFont type="icon-dip-color-primary-key" />}
                        {attr.display_key && <IconFont type="icon-dip-color-star" />}
                        {attr.incremental_key && <IconFont type="icon-dip-color-increment" />}
                      </>
                    )}
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
            );
          })
        ) : searchVal ? (
          <div className={styles['empty-state']}>
            <Empty image={noSearchResultImage} description={intl.get('Global.emptyNoSearchResult')} />
          </div>
        ) : null}
      </div>
    </div>
  );
};

export default CustomNode;
