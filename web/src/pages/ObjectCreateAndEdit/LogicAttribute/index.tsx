import { useState, forwardRef, useImperativeHandle, useMemo } from 'react';
import intl from 'react-intl-universal';
import { MinusCircleFilled, ExclamationCircleFilled } from '@ant-design/icons';
import { Empty, Input } from 'antd';
import ObjectIcon from '@/components/ObjectIcon';
import { deduplicateObjects } from '@/utils/object';
import * as OntologyObjectType from '@/services/object/type';
import createImage from '@/assets/images/common/create.svg';
import noSearchResultImage from '@/assets/images/common/no_search_result.svg';
import HOOKS from '@/hooks';
import { Button, IconFont, Table } from '@/web-library/common';
import EditDrawer from './EditDrawer';
import styles from './index.module.less';

interface LogicAttributeProps {
  basicValue: OntologyObjectType.BasicInfo;
  dataProperties: OntologyObjectType.DataProperty[];
  logicProperties?: OntologyObjectType.LogicProperty[];
}

const LogicAttribute = forwardRef((props: LogicAttributeProps, ref: any) => {
  const { dataProperties = [], logicProperties: initialLogicProperties = [], basicValue = { name: '', icon: '', color: '' } } = props;
  const [localLogicProperties, setLocalLogicProperties] = useState<OntologyObjectType.LogicProperty[]>(initialLogicProperties);
  const [open, setOpen] = useState(false);
  const [attrInfo, setAttrInfo] = useState<OntologyObjectType.LogicProperty>({} as OntologyObjectType.LogicProperty);
  const [searchInput, setSearchInput] = useState('');
  const { modal, message } = HOOKS.useGlobalContext();

  useImperativeHandle(ref, () => ({
    validateFields: () => {
      return Promise.resolve({ logicProperties: localLogicProperties });
    },
  }));

  const onClose = () => {
    setOpen(false);
  };

  const handleAdd = () => {
    setAttrInfo({} as OntologyObjectType.LogicProperty);
    setOpen(true);
  };

  const handleEdit = (record: OntologyObjectType.LogicProperty) => {
    setAttrInfo(record);
    setOpen(true);
  };

  const handleDelete = (record: OntologyObjectType.LogicProperty) => {
    if (!record.type || !record.name) {
      return;
    }

    modal.confirm({
      title: intl.get('Global.tipTitle'),
      icon: <ExclamationCircleFilled />,
      content: intl.get('Global.deleteConfirm', { name: `「${record.display_name || record.name}」` }),
      okText: intl.get('Global.ok'),
      okButtonProps: {
        style: { backgroundColor: '#ff4d4f', borderColor: '#ff4d4f' },
      },
      cancelText: intl.get('Global.cancel'),
      onOk: () => {
        setLocalLogicProperties(localLogicProperties.filter((item) => item.name !== record.name));
        message.success(intl.get('Global.deleteSuccess'));
      },
    });
  };

  const handleOk = (data: OntologyObjectType.LogicProperty) => {
    const isAddMode = !attrInfo?.name;

    if (isAddMode) {
      const nameExistsInDataProperties = dataProperties.some((p) => p.name === data.name);
      if (nameExistsInDataProperties) {
        message.error(`${intl.get('Global.attributeName')}「${data.name}」${intl.get('Global.alreadyExists')}`);
        return;
      }
    }

    const currentData = isAddMode ? deduplicateObjects([data, ...localLogicProperties], 'name') : deduplicateObjects([...localLogicProperties, data], 'name');
    setLocalLogicProperties(currentData);
    setOpen(false);
  };

  const filteredDataSource = useMemo(() => {
    if (!searchInput) return localLogicProperties;
    return localLogicProperties.filter((item) => item.name?.includes(searchInput) || item.display_name?.includes(searchInput));
  }, [localLogicProperties, searchInput]);

  const columns: any = [
    {
      title: intl.get('Global.attributeName'),
      dataIndex: 'name',
      width: 260,
      __selected: true,
      render: (value: string, record: OntologyObjectType.LogicProperty) => (
        <div className={styles['name-cell']}>
          <MinusCircleFilled
            style={{ color: 'rgba(0, 0, 0, 0.25)', cursor: 'pointer' }}
            onClick={(e) => {
              e.stopPropagation();
              handleDelete(record);
            }}
          />
          <span className={styles['data-name']}>{value}</span>
        </div>
      ),
    },
    {
      title: intl.get('Global.attributeDisplayName'),
      dataIndex: 'display_name',
      width: 260,
      __selected: true,
      render: (value: string) => <span className={styles['data-name']}>{value}</span>,
    },
    {
      title: intl.get('Global.description'),
      dataIndex: 'comment',
      width: 400,
      __selected: true,
      render: (value: string) => <span className={styles['data-desc']}>{value}</span>,
    },
    {
      title: intl.get('Object.bindResource'),
      dataIndex: 'parameters',
      width: 350,
      __selected: true,
      render: (_: any, record: OntologyObjectType.LogicProperty) => {
        if (record.parameters && record.parameters?.length > 0) {
          return (
            <div className={styles['bind-resource-cell']}>
              <div className={styles['data-resource']}>
                {record.data_source?.type === OntologyObjectType.LogicAttributeType.METRIC && (
                  <IconFont type="icon-dip-color-zhibiaometirc" style={{ fontSize: '24px' }} />
                )}
                {record.data_source?.type === OntologyObjectType.LogicAttributeType.OPERATOR && (
                  <IconFont type="icon-dip-color-suanzitool" style={{ fontSize: '24px' }} />
                )}
                <span className={styles['resource-name']}>{record.data_source?.name || ''}</span>
              </div>
            </div>
          );
        }
      },
    },
  ];

  return (
    <>
      <div className={styles['logic-attribute-container']}>
        <div className={styles['object-info-bar']}>
          <ObjectIcon icon={basicValue?.icon} color={basicValue?.color} size={28} iconSize={20} />
          <span className={styles['object-name']}>{basicValue?.name || ''}</span>
        </div>
        <div className={styles['logic-attribute-content']}>
          <div className={styles['header']}>
            <div className={styles['title-box']}>
              <div className={styles['title']}>{intl.get('Object.logicProperty')}</div>
              <div className={styles['add-button']} onClick={handleAdd}>
                <IconFont type="icon-dip-jia" style={{ fontSize: '14px' }} />
                <span>{intl.get('Global.add')}</span>
              </div>
            </div>
            <div className={styles['actions']}>
              <Input
                className={styles['search-input']}
                placeholder={intl.get('Object.searchAttributeNameOrDisplayName')}
                onChange={(value) => setSearchInput(value.target.value)}
                prefix={<IconFont type="icon-dip-search" style={{ fontSize: '16px', color: 'rgba(0, 0, 0, 0.25)' }} />}
              />
            </div>
          </div>

          <div className={styles['table-container']}>
            <Table.PageTable
              name="logic-attribute"
              rowKey="name"
              columns={columns}
              dataSource={filteredDataSource}
              pagination={false}
              canResize={true}
              autoScroll={true}
              onRow={(record: any) => ({
                onClick: () => handleEdit(record),
              })}
              locale={{
                emptyText: searchInput ? (
                  <Empty image={noSearchResultImage} description={intl.get('Global.emptyNoSearchResult')} />
                ) : (
                  <Empty
                    image={createImage}
                    description={
                      <span>
                        {intl.get('Global.click')}
                        <Button type="link" style={{ padding: 0 }} onClick={handleAdd}>
                          {intl.get('Global.createBtn')}
                        </Button>
                        {intl.get('Global.add')}
                      </span>
                    }
                  />
                ),
              }}
            />
          </div>
        </div>
      </div>

      <EditDrawer
        allData={[...dataProperties, ...localLogicProperties]}
        logicFields={localLogicProperties}
        title={attrInfo?.name ? intl.get('Object.logicAttributeMapping') : `${intl.get('Global.add')}${intl.get('Object.logicProperty')}`}
        open={open}
        onClose={onClose}
        onOk={handleOk}
        attrInfo={attrInfo}
      />
    </>
  );
});

export default LogicAttribute;
