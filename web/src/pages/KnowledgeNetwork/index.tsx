import { useEffect, useState } from 'react';
import intl from 'react-intl-universal';
import { useHistory } from 'react-router-dom';
import { Empty, Pagination, Spin, message, type MenuProps } from 'antd';
import ContainerIsVisible, { getTypePermissionOperation, matchPermission, PERMISSION_CODES } from '@/components/ContainerIsVisible';
import { showDeleteConfirm } from '@/components/DeleteConfirm';
import downFile from '@/utils/down-file';
import api from '@/services/knowledgeNetwork';
import * as KnowledgeNetworkType from '@/services/knowledgeNetwork/type';
import { baseConfig } from '@/services/request';
import createImage from '@/assets/images/common/create.svg';
import emptyImage from '@/assets/images/common/empty.png';
import noSearchResultImage from '@/assets/images/common/no_search_result.svg';
import HOOKS from '@/hooks';
import { Table, Button, Select, Title } from '@/web-library/common';
import CreateAndEditForm from './CreateAndEditForm';
import styles from './index.module.less';
import KnowledgeNetWorkCard from './KnowledgeNetWorkCard';
import ImportCom from './Operation/import';

type TableQueryState = Partial<KnowledgeNetworkType.GetNetworkListParams & { page: number }>;

type OperationMenuItem = Required<MenuProps>['items'][number];

const KnowledgeNetwork = () => {
  const DEFAULT_PAGE_SIZE = 20;
  const history = useHistory();
  const { modal } = HOOKS.useGlobalContext();
  const { pageState, pagination, onUpdateState } = HOOKS.usePageStateNew({ limit: DEFAULT_PAGE_SIZE }); // 分页信息
  const [tableData, setTableData] = useState<KnowledgeNetworkType.KnowledgeNetwork[]>([]);
  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [filterValues, setFilterValues] = useState<Pick<KnowledgeNetworkType.GetNetworkListParams, 'name_pattern' | 'tag'>>({ name_pattern: '', tag: 'all' }); // 筛选条件
  const [checkId, setCheckId] = useState<string>();
  const [open, setOpen] = useState<boolean>(false);
  const { page, limit, direction, sort } = pageState || {};
  const currentPage = page || 1;
  const currentLimit = limit || DEFAULT_PAGE_SIZE;
  const { name_pattern, tag } = filterValues || {};
  const currentSort = (sort as KnowledgeNetworkType.GetNetworkListParams['sort']) || 'update_time';
  const currentDirection = (direction as KnowledgeNetworkType.GetNetworkListParams['direction']) || 'desc';

  // 使用全局 Hook 获取国际化常量
  const { KN_MENU_SORT_ITEMS } = HOOKS.useConstants();

  /** 获取列表数据 */
  const getTableData = async (val?: TableQueryState): Promise<void> => {
    const nextPage = val?.page || currentPage;
    const nextLimit = val?.limit || currentLimit;
    const postData = {
      limit: nextLimit,
      direction: currentDirection,
      sort: currentSort,
      name_pattern,
      tag,
      offset: typeof val?.offset === 'number' ? val.offset : nextLimit * (nextPage - 1),
      ...val,
    };
    if (!postData.tag || postData.tag === 'all') delete postData.tag;
    if (val?.page) delete postData.page;
    setIsLoading(true);
    try {
      const res = await api.getNetworkList(postData);
      if (!res) return;
      const { total_count, entries } = res;

      onUpdateState({ ...postData, page: nextPage, count: total_count });
      setTableData(entries);
      setSelectedRowKeys([]);
      setIsLoading(false);
    } catch (error) {
      setIsLoading(false);
      console.log('error', error);
    }
  };

  useEffect(() => {
    baseConfig?.toggleSideBarShow(true);

    const timer = window.setTimeout(() => {
      getTableData();
    }, 0);

    return () => window.clearTimeout(timer);
  }, []);

  /** 筛选条件变更 */
  const onChangeTableOperation = (values: Pick<KnowledgeNetworkType.GetNetworkListParams, 'name_pattern' | 'tag'>) => {
    getTableData({ offset: 0, ...values });
    setFilterValues(values);
  };

  const onCancel = () => {
    setCheckId(undefined);
    setOpen(false);
  };

  const changeDel = (row?: KnowledgeNetworkType.KnowledgeNetwork) => {
    const ids = row ? [row.id] : selectedRowKeys;
    const content = row ? intl.get('Global.deleteConfirm', { name: row.name }) : intl.get('Global.deleteConfirmMultiple', { count: ids.length });
    showDeleteConfirm(modal, {
      content,
      async onOk() {
        await api.deleteNetwork(ids);
        message.success(intl.get('Global.deleteSuccess'));
        getTableData({ offset: 0 });
      },
    });
  };

  const exportData = async (id: string): Promise<void> => {
    const res = await api.getNetworkDetail({ knId: id, mode: 'export' });
    downFile(JSON.stringify(res, null, 2), res.name, 'json');
    message.success(intl.get('Global.exportSuccess'));
  };

  /** 操作按钮 */
  const onOperate = (key: string, record: KnowledgeNetworkType.KnowledgeNetwork) => {
    if (key === 'view') {
      localStorage.setItem('KnowledgeNetwork.id', record.id);
      history.push(`/ontology/main/overview?id=${record.id}`);
    }
    if (key === 'edit') {
      setCheckId(record.id);
      setOpen(true);
    }
    if (key === 'export') exportData(record.id);
    if (key === 'delete') changeDel(record);
  };

  const getDropdownItems = (record: KnowledgeNetworkType.KnowledgeNetwork): OperationMenuItem[] =>
    [
      { key: 'view', label: intl.get('Global.view'), visible: matchPermission(PERMISSION_CODES.VIEW, record.operations) },
      { key: 'edit', label: intl.get('Global.edit'), visible: matchPermission(PERMISSION_CODES.MODIFY, record.operations) },
      { key: 'export', label: intl.get('Global.export'), visible: matchPermission(PERMISSION_CODES.EXPORT, record.operations) },
      { key: 'delete', label: intl.get('Global.delete'), visible: matchPermission(PERMISSION_CODES.DELETE, record.operations) },
    ]
      .filter((item) => item.visible)
      .map(({ key, label }) => ({ key, label }));

  const onCheckChange = (record: KnowledgeNetworkType.KnowledgeNetwork, checked: boolean) => {
    setSelectedRowKeys((prev) => {
      if (checked) return [...new Set([...prev, record.id])];
      return prev.filter((id) => id !== record.id);
    });
  };

  /** 排序 */
  const handleSortChange = (val: { key: string }) => {
    const state = {
      sort: val.key as KnowledgeNetworkType.GetNetworkListParams['sort'],
      direction: (val.key !== currentSort ? 'desc' : currentDirection === 'desc' ? 'asc' : 'desc') as KnowledgeNetworkType.GetNetworkListParams['direction'],
    };
    getTableData(state);
  };

  const handlePageChange = (current: number, pageSize: number) => {
    const nextPage = pageSize !== currentLimit ? 1 : current;
    const state = {
      page: nextPage,
      limit: pageSize,
      offset: pageSize * (nextPage - 1),
    };
    onUpdateState(state);
    getTableData(state);
  };

  const emptyContent = (() => {
    if (filterValues.name_pattern || filterValues.tag !== 'all') {
      return <Empty image={noSearchResultImage} description={intl.get('Global.emptyNoSearchResult')} />;
    }

    if (matchPermission(PERMISSION_CODES.CREATE, getTypePermissionOperation('knowledge_network'))) {
      return (
        <Empty
          image={createImage}
          description={
            <span>
              {intl.get('Global.click')}
              <Button type="link" style={{ padding: 0 }} onClick={() => setOpen(true)}>
                {intl.get('Global.createBtn')}
              </Button>
              {intl.get('Global.add')}
            </span>
          }
        />
      );
    }

    return <Empty image={emptyImage} description={intl.get('KnowledgeNetwork.emptyDescription')} />;
  })();

  return (
    <div className={styles['box']}>
      <div id="userCom"></div>
      <Title>{intl.get('KnowledgeNetwork.businessKnowledgeNetwork')}</Title>
      <div className={styles['content']}>
        <div className={styles['operation-bar']}>
          <Table.Operation
            nameConfig={{ key: 'name_pattern', placeholder: intl.get('Global.filterByNameOrId') }}
            sortConfig={{ items: KN_MENU_SORT_ITEMS, order: currentDirection, rule: currentSort, onChange: handleSortChange }}
            initialFilter={filterValues}
            onChange={onChangeTableOperation}
            onRefresh={getTableData}
          >
            <ContainerIsVisible visible={matchPermission(PERMISSION_CODES.CREATE, getTypePermissionOperation('knowledge_network'))}>
              <Button.Create onClick={() => setOpen(true)} />
            </ContainerIsVisible>
            <ContainerIsVisible visible={matchPermission(PERMISSION_CODES.IMPORT, getTypePermissionOperation('knowledge_network'))}>
              <ImportCom callback={getTableData} />
            </ContainerIsVisible>
            <Select.LabelSelect
              key="tag"
              label={intl.get('Global.tag')}
              defaultValue="all"
              style={{ width: 190 }}
              options={[{ value: 'all', label: intl.get('Global.all') }]}
            />
          </Table.Operation>
        </div>

        <Spin spinning={isLoading} wrapperClassName={styles['loading-wrapper']}>
          <div className={styles['list-content']}>
            {tableData.length ? (
              <div className={styles['card-list']}>
                {tableData.map((record) => {
                  const dropdownItems = getDropdownItems(record);
                  return (
                    <KnowledgeNetWorkCard
                      key={record.id}
                      record={record}
                      checked={selectedRowKeys.includes(record.id)}
                      dropdownItems={dropdownItems}
                      onClick={() => onOperate('view', record)}
                      onCheckChange={onCheckChange}
                      onOperate={onOperate}
                    />
                  );
                })}
              </div>
            ) : (
              <div className={styles['empty-content']}>{emptyContent}</div>
            )}
          </div>
        </Spin>

        {!!tableData.length && (
          <div className={styles['pagination']}>
            <Pagination
              {...pagination}
              current={pagination?.current || currentPage}
              pageSize={pagination?.pageSize || currentLimit}
              showSizeChanger
              pageSizeOptions={['20', '50', '100']}
              showTotal={(total) => intl.get('Global.total', { total })}
              onChange={handlePageChange}
            />
          </div>
        )}
        <CreateAndEditForm open={open} onCancel={onCancel} id={checkId} callBack={() => getTableData({ offset: 0 })} />
      </div>
    </div>
  );
};

export default KnowledgeNetwork;
export { default as KnowledgeNetWorkCard } from './KnowledgeNetWorkCard';
