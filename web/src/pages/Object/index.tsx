import { useEffect, useState, useCallback } from 'react';
import intl from 'react-intl-universal';
import { useHistory } from 'react-router-dom';
import { EllipsisOutlined } from '@ant-design/icons';
import { Dropdown, Empty, message, Tooltip } from 'antd';
import { SorterResult } from 'antd/es/table/interface';
import { TableProps } from 'antd/lib/table';
import dayjs from 'dayjs';
import { map } from 'lodash-es';
import { showDeleteConfirm } from '@/components/DeleteConfirm';
import ObjectIcon from '@/components/ObjectIcon';
import Tags from '@/components/Tags';
import api from '@/services/object';
import * as ObjectType from '@/services/object/type';
import createImage from '@/assets/images/common/create.svg';
import emptyImage from '@/assets/images/common/empty.png';
import noSearchResultImage from '@/assets/images/common/no_search_result.svg';
import ENUMS from '@/enums';
import HOOKS from '@/hooks';
import { KnowledgeNetworkType } from '@/services';
import { Table, Button, Select, Title, IconFont } from '@/web-library/common';
import Detail from './Detail';
import styles from './index.module.less';

interface TProps {
  detail?: KnowledgeNetworkType.KnowledgeNetwork;
  isPermission: boolean;
}

const KnowledgeNetwork = (props: TProps) => {
  const history = useHistory();
  const { detail, isPermission } = props;
  const knId = detail?.id || localStorage.getItem('KnowledgeNetwork.id');
  const { modal } = HOOKS.useGlobalContext();
  const { pageState, pagination, onUpdateState } = HOOKS.usePageStateNew();
  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([]);
  const [selectedRows, setSelectedRows] = useState<ObjectType.Detail[]>([]);
  const [tableData, setTableData] = useState<ObjectType.Detail[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [filterValues, setFilterValues] = useState<Pick<ObjectType.ListQuery, 'name_pattern' | 'tag'>>({ name_pattern: '', tag: 'all' });
  const [objectDetail, setObjectDetail] = useState<ObjectType.Detail>();

  const { page, limit, direction, sort } = pageState || {};
  const { name_pattern, tag } = filterValues || {};
  const { OBJECT_MENU_SORT_ITEMS } = HOOKS.useConstants();

  const getTableData = useCallback(
    async (val?: any): Promise<void> => {
      const postData = {
        offset: val?.page ? limit * (val?.page - 1) : limit * (page - 1),
        limit,
        direction,
        sort,
        name_pattern,
        tag,
        ...val,
      };
      if (!postData.tag || postData.tag === 'all') delete postData.tag;
      const curPage = val?.page || page;
      if (val?.page) delete postData.page;
      setIsLoading(true);
      try {
        const res = await api.objectGet(detail?.id as string, postData);
        if (!res) return;
        const { total_count, entries } = res;

        onUpdateState({ ...postData, page: curPage, count: total_count });
        setTableData(entries);
        setSelectedRowKeys([]);
        setSelectedRows([]);
      } catch (error) {
        console.error('getTableData error:', error);
      } finally {
        setIsLoading(false);
      }
    },
    [detail?.id, limit, page, direction, sort, name_pattern, tag, onUpdateState]
  );

  useEffect(() => {
    if (detail?.id) {
      getTableData();
    }
  }, [detail?.id]);

  const onChangeTableOperation = (values: Pick<ObjectType.ListQuery, 'name_pattern' | 'tag'>) => {
    getTableData({ offset: 0, ...values });
    setFilterValues(values);
  };

  const handleTableChange: TableProps['onChange'] = useCallback(
    async (pagination: any, _filters: any, sorter: any): Promise<void> => {
      const { field, order } = sorter as SorterResult;
      const { current, pageSize } = pagination;
      const stateOrder = ENUMS.SORT_ENUM[order as keyof typeof ENUMS.SORT_ENUM] || 'desc';
      const state = { page: current, limit: pageSize, sort: (field as string) || 'update_time', direction: stateOrder };
      onUpdateState(state);
      getTableData(state);
    },
    [onUpdateState, getTableData]
  );

  const onDelete = useCallback(
    async (items: ObjectType.Detail[], isBatch?: boolean) => {
      try {
        const objectIds = map(items, (item) => item?.id);
        await api.deleteObjectTypes(knId as string, objectIds);
        getTableData();
        message.success(intl.get('Global.deleteSuccess'));
        if (isBatch) setSelectedRowKeys([]);
      } catch (error) {
        console.error('onDelete error:', error);
      }
    },
    [knId, getTableData]
  );

  const onDeleteConfirm = useCallback(
    (items: ObjectType.Detail[], isBatch?: boolean, callBack?: () => void) => {
      const name = map(items, (item) => `「${item?.name}」`).join('、');
      const length = items.length || 0;
      showDeleteConfirm(modal, {
        content: length > 1 ? intl.get('Global.deleteConfirmMultiple', { count: length }) : intl.get('Global.deleteConfirm', { name }),
        onOk: async () => {
          await onDelete(items, isBatch);
          if (callBack) callBack();
        },
      });
    },
    [modal, onDelete]
  );

  const toCreateOrEdit = (objId?: string) => {
    if (objId) {
      history.push(`/ontology/object/edit/${objId}`);
      return;
    }
    history.push(`/ontology/object/create`);
  };

  const onOpenDetail = (val: ObjectType.Detail) => setObjectDetail(val);
  const onCloseDetail = () => setObjectDetail(undefined);

  const onOperate = (key: string, record: ObjectType.Detail) => {
    if (key === 'view') {
      onOpenDetail(record);
    }
    if (key === 'edit') {
      toCreateOrEdit(record.id);
    }
    if (key === 'delete') onDeleteConfirm([record]);
    if (key === 'index') {
      history.push(`/ontology/object/settting/${record.id}`);
    }
  };

  const columns: any = [
    {
      title: intl.get('Global.name'),
      dataIndex: 'name',
      fixed: 'left',
      sorter: true,
      width: 350,
      __fixed: true,
      __selected: true,
      render: (value: string, record: ObjectType.Detail) => (
        <div className={styles['object-title-box']} title={value} onClick={() => onOpenDetail(record)}>
          <ObjectIcon icon={record.icon} color={record.color} />
          <span>{record.name}</span>
        </div>
      ),
    },
    {
      title: intl.get('Global.operation'),
      dataIndex: 'operation',
      fixed: 'left',
      width: 80,
      __fixed: true,
      __selected: true,
      render: (_value: any, record: any) => {
        const allOperations = [
          { key: 'view', label: intl.get('Global.view'), visible: true },
          { key: 'edit', label: intl.get('Global.edit'), visible: isPermission },
          { key: 'index', label: intl.get('Object.indexConfiguration'), visible: isPermission },
          { key: 'delete', label: intl.get('Global.delete'), visible: isPermission },
        ];
        const dropdownMenu: any = allOperations.filter((item) => item.visible).map(({ key, label }: any) => ({ key, label }));
        return (
          <Dropdown
            trigger={['click']}
            menu={{
              items: dropdownMenu,
              onClick: (event: any) => {
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
      title: () => (
        <div className={styles['has-index']}>
          <span>{intl.get('Object.hasIndex')}</span>
          <Tooltip title={intl.get('Object.hasIndexTip')}>
            <IconFont type="icon-dip-color-tip" />
          </Tooltip>
        </div>
      ),
      dataIndex: 'status',
      width: 150,
      __selected: true,
      render: (value: any) => (value?.index_available === true ? intl.get('Global.yes') : intl.get('Global.no')),
    },
    {
      title: intl.get('Global.tag'),
      dataIndex: 'tags',
      width: 150,
      __selected: true,
      render: (value: string[]) => <Tags value={value} />,
    },
    {
      title: intl.get('Global.modifier'),
      dataIndex: 'updater',
      width: 150,
      __selected: true,
      render: (value: any, record: any) => record?.updater?.name || '--',
    },
    {
      title: intl.get('Global.updateTime'),
      dataIndex: 'update_time',
      width: 200,
      __selected: true,
      render: (value: string) => (value ? dayjs(value).format('YYYY/MM/DD HH:mm:ss') : '--'),
    },
  ];

  const rowSelection = {
    selectedRowKeys,
    onChange: (selectedRowKeys: any, selectedRows: any): void => {
      setSelectedRowKeys(selectedRowKeys);
      setSelectedRows(selectedRows);
    },
    onSelectAll: (selected: any): void => {
      const newSelectedRowKeys = selected ? tableData.map((item) => item.id) : [];
      const newSelectedRows = selected ? tableData : [];

      setSelectedRowKeys(newSelectedRowKeys);
      setSelectedRows(newSelectedRows);
    },
    getCheckboxProps: (row: any): Record<string, any> => ({
      disabled: row.builtin,
    }),
  };

  const handleSortChange = (val: { key: string }) => {
    const state = {
      sort: val.key,
      direction: val.key !== sort ? 'desc' : direction === 'desc' ? 'asc' : 'desc',
    };
    getTableData(state);
  };

  return (
    <div className={styles['box']}>
      <Title>{intl.get('Global.objectClass')}</Title>
      <Table.PageTable
        name="object-model"
        rowKey="id"
        columns={columns}
        loading={isLoading}
        dataSource={tableData}
        rowSelection={rowSelection}
        pagination={pagination}
        onChange={handleTableChange}
        locale={{
          emptyText:
            filterValues.name_pattern || filterValues.tag !== 'all' ? (
              <Empty image={noSearchResultImage} description={intl.get('Global.emptyNoSearchResult')} />
            ) : isPermission ? (
              <Empty
                image={createImage}
                description={
                  <span>
                    {intl.get('Global.click')}
                    <Button type="link" style={{ padding: 0 }} onClick={() => toCreateOrEdit()}>
                      {intl.get('Global.createBtn')}
                    </Button>
                    {intl.get('Global.add')}
                  </span>
                }
              />
            ) : (
              <Empty image={emptyImage} description={intl.get('Global.noData')} />
            ),
        }}
      >
        <Table.Operation
          nameConfig={{ key: 'name_pattern', placeholder: intl.get('Global.searchNameId') }}
          sortConfig={{ items: OBJECT_MENU_SORT_ITEMS, order: direction, rule: sort, onChange: handleSortChange }}
          initialFilter={filterValues}
          onChange={onChangeTableOperation}
          onRefresh={getTableData}
          isControlFilter
        >
          {isPermission && <Button.Create onClick={() => toCreateOrEdit()} />}
          {isPermission && <Button.Delete onClick={() => onDeleteConfirm(selectedRows, true)} disabled={!selectedRows?.length} />}
          <Select.LabelSelect
            key="tag"
            label={intl.get('Global.tag')}
            defaultValue="all"
            style={{ width: 190 }}
            options={[{ value: 'all', label: intl.get('Global.all') }]}
          />
        </Table.Operation>
      </Table.PageTable>
      <Detail
        open={!!objectDetail}
        sourceData={objectDetail}
        onClose={onCloseDetail}
        onDeleteConfirm={onDeleteConfirm}
        goToCreateAndEditPage={toCreateOrEdit}
        isPermission={isPermission}
      />
    </div>
  );
};

export default KnowledgeNetwork;
