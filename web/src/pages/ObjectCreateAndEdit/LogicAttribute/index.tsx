import { useState, forwardRef, useImperativeHandle, useMemo } from 'react';
import intl from 'react-intl-universal';
import { EllipsisOutlined } from '@ant-design/icons';
import { Dropdown, Empty, Tooltip } from 'antd';
import { showDeleteConfirm } from '@/components/DeleteConfirm';
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
  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([]);
  const [selectedRows, setSelectedRows] = useState<any[]>([]);

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

  const handleDelete = (record: any) => {
    const content = record ? intl.get('Global.deleteConfirm', { name: record.name }) : intl.get('Global.deleteConfirmMultiple', { count: selectedRows.length });
    showDeleteConfirm(modal, {
      content,
      onOk: () => {
        if (record) {
          setLocalLogicProperties(localLogicProperties.filter((item) => item.name !== record.name));
        } else {
          setLocalLogicProperties(localLogicProperties.filter((item) => !selectedRowKeys.includes(item.name)));
        }
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

  const onChangeTableOperation = (value: any) => {
    console.log(value);
    setSearchInput(value.name);
  };

  const onOperate = (key: string, record: any) => {
    if (key === 'edit') {
      handleEdit(record);
    }
    if (key === 'delete') {
      handleDelete(record);
    }
  };

  const columns: any = [
    {
      title: intl.get('Global.attributeName'),
      dataIndex: 'name',
      width: 260,
      __selected: true,
      render: (value: string) => (
        <div className={styles['name-cell']}>
          <span className={styles['data-name']}>{value}</span>
        </div>
      ),
    },
    {
      title: intl.get('Global.operation'),
      width: 100,
      align: 'center',
      __selected: true,
      render: (_text: any, record: any): JSX.Element => {
        const dropdownMenu: any = [
          { key: 'edit', label: intl.get('Global.edit') },
          { key: 'delete', label: intl.get('Global.delete') },
        ];
        return (
          <Dropdown
            trigger={['click']}
            menu={{
              items: dropdownMenu,
              onClick: (event) => {
                event.domEvent.stopPropagation();
                onOperate(event?.key, record);
              },
            }}
          >
            <Button.Icon icon={<EllipsisOutlined style={{ fontSize: 20 }} />} onClick={(event) => event.stopPropagation()} />
          </Dropdown>
        );
      },
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

  const rowSelection = {
    selectedRowKeys,
    onChange: (selectedRowKeys: any, selectedRows: any): void => {
      setSelectedRowKeys(selectedRowKeys);
      setSelectedRows(selectedRows);
    },
    onSelectAll: (selected: any): void => {
      const newSelectedRowKeys = selected ? filteredDataSource.map((item: any) => item.id) : [];
      const newSelectedRows = selected ? filteredDataSource : [];

      setSelectedRowKeys(newSelectedRowKeys);
      setSelectedRows(newSelectedRows);
    },
    getCheckboxProps: (row: any): Record<string, any> => ({
      disabled: row.builtin,
    }),
  };

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
              <Tooltip title={intl.get('Object.logicPropertyTip')}>
                <IconFont type="icon-dip-color-tip" className={styles.helpIcon} />
              </Tooltip>
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
              rowSelection={rowSelection}
              locale={{
                emptyText: searchInput ? (
                  <Empty image={noSearchResultImage} description={intl.get('Global.emptyNoSearchResult')} />
                ) : (
                  <Empty
                    image={createImage}
                    description={
                      <div>
                        <span>
                          {intl.get('Global.click')}
                          <Button type="link" style={{ padding: 0 }} onClick={handleAdd}>
                            {intl.get('Global.createBtn')}
                          </Button>
                          {intl.get('Global.add')}
                        </span>
                        <div className={styles['logic-property-tip']}>{intl.get('Object.skipLogicPropertyTip')}</div>
                      </div>
                    }
                  />
                ),
              }}
            >
              <Table.Operation nameConfig={{ key: 'name', placeholder: intl.get('Global.search') }} onChange={onChangeTableOperation}>
                <Button.Create onClick={handleAdd} />
                <Button.Delete disabled={!selectedRows?.length} onClick={() => handleDelete('')} />
              </Table.Operation>
            </Table.PageTable>
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
