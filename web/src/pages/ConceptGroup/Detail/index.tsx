import { useState, useEffect, type ChangeEvent, type Key } from 'react';
import intl from 'react-intl-universal';
import { useHistory, useLocation, useParams } from 'react-router-dom';
import { EllipsisOutlined } from '@ant-design/icons';
import { Table, Tabs, Input, Empty, Dropdown, Tag, Select } from 'antd';
import { showDeleteConfirm } from '@/components/DeleteConfirm';
import DetailPageHeader from '@/components/DetailPageHeader';
import DetailSummaryCard from '@/components/DetailSummaryCard';
import ObjectIcon from '@/components/ObjectIcon';
import downFile from '@/utils/down-file';
import * as ActionType from '@/services/action/type';
import api from '@/services/conceptGroup';
import * as ConceptGroupType from '@/services/conceptGroup/type';
import { baseConfig } from '@/services/request';
import HOOKS from '@/hooks';
import { Title, Button, IconFont } from '@/web-library/common';
import AddObjectTypesModal from '../AddObjectTypesModal';
import CreateAndEditForm from '../CreateAndEditForm';
import styles from './index.module.less';

const { TabPane } = Tabs;

type TabKey = 'object' | 'relation' | 'action';

const Detail = () => {
  const history = useHistory();
  const location = useLocation<{ isPermission?: boolean }>();
  const { id = '' } = useParams<{ id: string }>();
  const isPermission = location.state?.isPermission ?? true;
  const knId = localStorage.getItem('KnowledgeNetwork.id') || sessionStorage.getItem('knId') || '';
  const { modal, message } = HOOKS.useGlobalContext();

  const [source, setSource] = useState<ConceptGroupType.Detail>();
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState<TabKey>('object');
  const [selectedRowKeys, setSelectedRowKeys] = useState<Key[]>([]);
  const [addObjectTypesOpen, setAddObjectTypesOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [searchStates, setSearchStates] = useState<Record<TabKey, { searchText: string; tag: string }>>({
    object: { searchText: '', tag: 'all' },
    relation: { searchText: '', tag: 'all' },
    action: { searchText: '', tag: 'all' },
  });

  const goBack = () => {
    if (knId) {
      history.push(`/ontology/main/concept-group?id=${knId}`);
      return;
    }
    history.push('/ontology/main/concept-group');
  };

  const getDetail = async () => {
    if (!knId || !id) return;
    setLoading(true);
    try {
      const res = await api.detailConceptGroup(knId, id);
      setSource(res);
    } catch (error) {
      console.error('getConceptGroupDetail error:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    baseConfig?.toggleSideBarShow(false);
    return () => {
      baseConfig?.toggleSideBarShow(true);
    };
  }, []);

  useEffect(() => {
    getDetail();
  }, [id, knId]);

  const deleteConceptGroup = async () => {
    if (!knId || !id) return;
    try {
      await api.deleteConceptGroup(knId, [id]);
      message.success(intl.get('Global.deleteSuccess'));
      goBack();
    } catch (error) {
      console.error('deleteConceptGroup error:', error);
    }
  };

  const onDeleteConfirm = () => {
    showDeleteConfirm(modal, {
      content: intl.get('ConceptGroup.confirmDeleteSingle', { name: source?.name || '' }),
      onOk: deleteConceptGroup,
    });
  };

  const exportData = async () => {
    if (!knId || !id) return;
    try {
      const res = await api.detailConceptGroup(knId, id);
      downFile(JSON.stringify(res, null, 2), res.name, 'json');
      message.success(intl.get('Global.exportSuccess'));
    } catch (error) {
      console.error('exportConceptGroup error:', error);
    }
  };

  const onMoreAction = (data: { key: string }) => {
    if (data.key === 'export') exportData();
    if (data.key === 'delete') onDeleteConfirm();
  };

  const actionTypeLabels = {
    [ActionType.ActionTypeEnum.Add]: intl.get('Global.create'),
    [ActionType.ActionTypeEnum.Modify]: intl.get('Global.edit'),
    [ActionType.ActionTypeEnum.Delete]: intl.get('Global.delete'),
  };

  const objectColumns = [
    {
      title: intl.get('Global.name'),
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
      render: (_value: string, record: any) => (
        <div className={styles['table-name-cell']} onClick={() => history.push(`/ontology/object/detail/${record.id}`)}>
          <ObjectIcon icon={record.icon} color={record.color} />
          <span className="g-ellipsis-1">{record.name}</span>
        </div>
      ),
    },
    {
      title: intl.get('Global.tag'),
      dataIndex: 'tags',
      key: 'tags',
      render: (tags: string[]) => (Array.isArray(tags) && tags.length ? tags.map((tag) => <Tag key={tag}>{tag}</Tag>) : '--'),
    },
  ];

  const relationColumns = [
    {
      title: intl.get('Global.name'),
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: intl.get('Global.sourceObjectType'),
      dataIndex: 'source_object_type_id',
      key: 'source_object_type_id',
      render: (_value: string, record: any) => {
        const sourceObject = record?.source_object_type || {};
        if (!sourceObject?.name) return '--';
        return (
          <div className={styles['table-name-cell']}>
            <ObjectIcon icon={sourceObject.icon} color={sourceObject.color} />
            <span className="g-ellipsis-1">{sourceObject.name}</span>
          </div>
        );
      },
    },
    {
      title: intl.get('Global.targetObjectType'),
      dataIndex: 'target_object_type_id',
      key: 'target_object_type_id',
      render: (_value: string, record: any) => {
        const targetObject = record?.target_object_type || {};
        if (!targetObject?.name) return '--';
        return (
          <div className={styles['table-name-cell']}>
            <ObjectIcon icon={targetObject.icon} color={targetObject.color} />
            <span className="g-ellipsis-1">{targetObject.name}</span>
          </div>
        );
      },
    },
    {
      title: intl.get('Global.tag'),
      dataIndex: 'tags',
      key: 'tags',
      render: (tags: string[]) => (Array.isArray(tags) && tags.length ? tags.map((tag) => <Tag key={tag}>{tag}</Tag>) : '--'),
    },
  ];

  const actionColumns = [
    {
      title: intl.get('Global.name'),
      dataIndex: 'name',
      key: 'name',
      render: (_value: string, record: any) => (
        <div className={styles['table-name-cell']}>
          <ObjectIcon icon="icon-dip-hangdonglei" color={record.color} />
          <span className="g-ellipsis-1">{record.name}</span>
        </div>
      ),
    },
    {
      title: intl.get('Global.type'),
      dataIndex: 'action_type',
      key: 'action_type',
      width: 120,
      render: (value: ActionType.ActionTypeEnum) => actionTypeLabels[value] || '--',
    },
    {
      title: intl.get('Global.bindObjectType'),
      dataIndex: 'object_type',
      key: 'object_type',
      render: (value: any) => {
        if (!value?.name) return '--';
        return (
          <div className={styles['table-name-cell']}>
            <ObjectIcon icon={value.icon} color={value.color} />
            <span className="g-ellipsis-1">{value.name}</span>
          </div>
        );
      },
    },
    {
      title: intl.get('Global.tag'),
      dataIndex: 'tags',
      key: 'tags',
      render: (tags: string[]) => (Array.isArray(tags) && tags.length ? tags.map((tag) => <Tag key={tag}>{tag}</Tag>) : '--'),
    },
  ];

  const currentTabData = !source
    ? []
    : activeTab === 'object'
      ? source.object_types || []
      : activeTab === 'relation'
        ? source.relation_types || []
        : source.action_types || [];

  const { searchText, tag } = searchStates[activeTab];
  const keyword = searchText.trim().toLowerCase();
  const filteredData = (currentTabData || []).filter((item: any) => {
    const matchedSearch =
      !keyword ||
      String(item?.name || '')
        .toLowerCase()
        .includes(keyword);
    const matchedTag = tag === 'all' || (Array.isArray(item?.tags) && item.tags.includes(tag));
    return matchedSearch && matchedTag;
  });

  const tagSet = new Set<string>();
  (currentTabData || []).forEach((item: any) => {
    (item?.tags || []).forEach((itemTag: string) => {
      if (itemTag) tagSet.add(itemTag);
    });
  });
  const tagOptions = [{ label: intl.get('Global.all'), value: 'all' }, ...Array.from(tagSet).map((itemTag) => ({ label: itemTag, value: itemTag }))];

  const tableColumns = activeTab === 'object' ? objectColumns : activeTab === 'relation' ? relationColumns : actionColumns;

  const handleSearchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value;
    setSearchStates((prev) => ({
      ...prev,
      [activeTab]: { ...prev[activeTab], searchText: value },
    }));
  };

  const handleTagChange = (value: string) => {
    setSearchStates((prev) => ({
      ...prev,
      [activeTab]: { ...prev[activeTab], tag: value },
    }));
  };

  const handleRemoveObjectTypes = async () => {
    if (!knId || !source?.id) return;
    if (!selectedRowKeys.length) {
      message.warning(intl.get('ConceptGroup.pleaseSelectObjectTypesToRemove'));
      return;
    }
    try {
      await api.removeObjectTypesFromGroup(knId, source.id, selectedRowKeys as string[]);
      message.success(intl.get('Global.removeSuccess'));
      setSelectedRowKeys([]);
      getDetail();
    } catch (error) {
      console.error('removeObjectTypesFromGroup error:', error);
    }
  };

  return (
    <div className={styles['concept-group-detail-page']}>
      <DetailPageHeader
        onBack={goBack}
        icon={<ObjectIcon icon="icon-dip-fenzu" color={source?.color} />}
        title={source?.name || '--'}
        actions={
          <>
            {isPermission && (
              <Button className={styles['top-edit-btn']} icon={<IconFont type="icon-dip-bianji" />} onClick={() => setEditOpen(true)}>
                {intl.get('Global.edit')}
              </Button>
            )}
            <Dropdown
              trigger={['click']}
              menu={{
                items: [{ key: 'export', label: intl.get('Global.export') }, ...(isPermission ? [{ key: 'delete', label: intl.get('Global.delete') }] : [])],
                onClick: onMoreAction,
              }}
            >
              <Button className={styles['top-more-btn']} icon={<EllipsisOutlined style={{ fontSize: 20 }} />} />
            </Dropdown>
          </>
        }
      />

      <DetailSummaryCard
        id={source?.id}
        icon={<ObjectIcon size={32} iconSize={20} icon="icon-dip-fenzu" color={source?.color} />}
        name={source?.name}
        tags={source?.tags}
        comment={source?.comment}
        modifier={source?.updater?.name || source?.creator?.name}
        updateTime={source?.update_time}
      />

      <div className={styles['section-card']}>
        <Title className={styles['section-title']}>{intl.get('ConceptGroup.groupDetail')}</Title>
        <Tabs activeKey={activeTab} onChange={(tab) => setActiveTab(tab as TabKey)}>
          <TabPane tab={intl.get('Global.objectClass')} key="object" />
          <TabPane tab={intl.get('Global.edgeClass')} key="relation" />
          <TabPane tab={intl.get('Global.actionClass')} key="action" />
        </Tabs>
        <div className={styles['section-toolbar']}>
          <div className={styles['toolbar-left']}>
            {activeTab === 'object' && isPermission && (
              <>
                <Button.Create onClick={() => setAddObjectTypesOpen(true)}>{intl.get('Global.add')}</Button.Create>
                <Button.Delete onClick={handleRemoveObjectTypes} disabled={!selectedRowKeys.length}>
                  {intl.get('Global.remove')}
                </Button.Delete>
              </>
            )}
          </div>
          <div className={styles['toolbar-right']}>
            <Input.Search
              placeholder={intl.get('Global.searchName')}
              size="middle"
              value={searchStates[activeTab].searchText}
              style={{ width: 220 }}
              allowClear
              onChange={handleSearchChange}
            />
            <Select style={{ width: 140 }} value={searchStates[activeTab].tag} options={tagOptions} onChange={handleTagChange} />
          </div>
        </div>
        <Table
          size="small"
          rowKey="id"
          columns={tableColumns as any}
          loading={loading}
          dataSource={filteredData}
          scroll={{ y: 420, x: '100%' }}
          locale={{ emptyText: <Empty description={intl.get('Global.noData')} /> }}
          rowSelection={
            activeTab === 'object'
              ? {
                  selectedRowKeys,
                  onChange: (rowKeys: Key[]) => setSelectedRowKeys(rowKeys),
                }
              : undefined
          }
          pagination={{
            showSizeChanger: true,
            pageSizeOptions: ['10', '20', '50'],
            showTotal: (total) => intl.get('Global.total', { total }),
          }}
        />
      </div>

      <CreateAndEditForm open={editOpen} onCancel={() => setEditOpen(false)} id={source?.id} callBack={getDetail} knId={knId} />
      <AddObjectTypesModal
        open={addObjectTypesOpen}
        onCancel={() => setAddObjectTypesOpen(false)}
        onSuccess={() => {
          setAddObjectTypesOpen(false);
          getDetail();
        }}
        knId={knId}
        groupId={source?.id || ''}
        groupName={source?.name || ''}
      />
    </div>
  );
};

export default Detail;
