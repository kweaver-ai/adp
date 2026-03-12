import { useState, useEffect, useRef } from 'react';
import intl from 'react-intl-universal';
import { useHistory, useLocation, useParams } from 'react-router-dom';
import { EllipsisOutlined } from '@ant-design/icons';
import { Table, Dropdown, Segmented, Input, TableColumnProps, Form, Tooltip } from 'antd';
import DetailPageHeader from '@/components/DetailPageHeader';
import DetailSummaryCard from '@/components/DetailSummaryCard';
import FieldTypeIcon from '@/components/FieldTypeIcon';
import ObjectIcon from '@/components/ObjectIcon';
import { formatKeyOfObjectToLine } from '@/utils/format-objectkey-structure';
import formatFileSize from '@/utils/formatFileSize';
import objectApi from '@/services/object';
import * as OntologyObjectType from '@/services/object/type';
import { listObjects } from '@/services/ontologyQuery';
import * as OntologyQuery from '@/services/ontologyQuery/type';
import { baseConfig } from '@/services/request';
import HOOKS from '@/hooks';
import SERVICE, { ObjectType } from '@/services';
import { Title, Button, IconFont } from '@/web-library/common';
import DataFilter from '@/web-library/components/DataFilter';
import DeleteConfirmModal from '../DeleteConfirmModal';
import styles from './index.module.less';

const INIT_FILTER = { field: undefined, value: null, operation: '==', value_from: 'const' };

const Detail = () => {
  const history = useHistory();
  const location = useLocation<{ isPermission?: boolean }>();
  const { id = '' } = useParams<{ id: string }>();
  const isPermission = location.state?.isPermission ?? true;
  const [source, setSource] = useState<ObjectType.Detail>();
  const { message } = HOOKS.useGlobalContext();
  const { id: objectId, tags, comment, kn_id, data_source, status } = source || {};
  const knId = localStorage.getItem('KnowledgeNetwork.id') || sessionStorage.getItem('knId') || kn_id || '';
  const [form] = Form.useForm();
  const { OBJECT_PROPERTY_TYPES, OBJECT_PROPERTY_TYPE_OPTIONS } = HOOKS.useConstants();
  const [type, setType] = useState(OBJECT_PROPERTY_TYPES.DATA_PROPERTY);
  const [searchText, setSearchText] = useState('');
  const [dataSearchText, setDataSearchText] = useState('');
  const [data, setData] = useState<OntologyQuery.ObjectDataResponse['datas']>([]);
  const [dataLoading, setDataLoading] = useState(false);
  const [dataColumns, setDataColumns] = useState<TableColumnProps<OntologyQuery.ObjectDataResponse['object_type']>[]>([]);
  const [switchFilter, setSwitchFilter] = useState<boolean>(false);
  const dataFilterRef = useRef<any>(null);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [deleteConfirmLoading, setDeleteConfirmLoading] = useState(false);

  const goBack = () => {
    history.goBack();
  };

  const getDetail = async () => {
    if (!knId || !id) return;
    try {
      const result = await SERVICE.object.getDetail(knId, [id]);
      setSource(result?.[0]);
    } catch (error) {
      console.error('getObjectDetail error:', error);
    }
  };

  const deleteObject = async (forceDelete?: boolean) => {
    if (!source?.id || !knId) return;
    try {
      await objectApi.deleteObjectTypes(knId, [source.id], forceDelete);
      message.success(intl.get('Global.deleteSuccess'));
      goBack();
    } catch (error) {
      console.error('deleteObject error:', error);
    }
  };

  const onDeleteConfirm = () => {
    setDeleteConfirmOpen(true);
  };

  const onDeleteConfirmOk = async () => {
    setDeleteConfirmLoading(true);
    try {
      await deleteObject(true);
      setDeleteConfirmOpen(false);
    } finally {
      setDeleteConfirmLoading(false);
    }
  };

  const onDeleteConfirmCancel = () => {
    setDeleteConfirmOpen(false);
    setDeleteConfirmLoading(false);
  };

  const fetchData = async (filters?: any) => {
    if (!objectId || !data_source?.id || !knId) return;
    setDataLoading(true);
    try {
      const params: OntologyQuery.ListObjectsRequest = {
        knId,
        otId: objectId,
        includeTypeInfo: true,
        includeLogicParams: false,
        body: { ...filters, limit: 100 },
      };
      const res = await listObjects(params);
      if (res) {
        setData(res.datas?.map((item, index) => ({ ...item, _tempId: item.id || index })) || []);
        setDataColumns(
          res.object_type?.data_properties.map((item) => ({
            title: item.name,
            dataIndex: item.name,
            key: item.name,
            width: 150,
            ellipsis: true,
          })) || []
        );
      }
    } catch (error) {
      console.error('fetchObjectData error:', error);
      message.error(intl.get('Global.loadDataFailed'));
    } finally {
      setDataLoading(false);
    }
  };

  useEffect(() => {
    baseConfig?.toggleSideBarShow(false);
    return () => {
      baseConfig?.toggleSideBarShow(true);
    };
  }, []);

  useEffect(() => {
    if (!id || !knId) return;
    getDetail();
  }, [id, knId]);

  useEffect(() => {
    if (!objectId) return;
    fetchData();
  }, [objectId, data_source?.id, knId]);

  useEffect(() => {
    form.setFieldValue('dataFilter', INIT_FILTER);
  }, [form]);

  const handleFilterSearch = () => {
    if (switchFilter && dataFilterRef.current) {
      const filters = form.getFieldValue('dataFilter');
      const validate = dataFilterRef.current?.validate();
      if (!validate) {
        fetchData({ condition: formatKeyOfObjectToLine(filters) });
      }
      return;
    }
    fetchData();
  };

  const onChange = (data: any) => {
    if (data.key === 'delete') onDeleteConfirm();
  };

  const columns: TableColumnProps<ObjectType.Field>[] = [
    {
      title: intl.get('Global.name'),
      dataIndex: 'name',
      key: 'name',
      render: (value: string, record: any) => (
        <div className={styles.propertyTitle}>
          <div className="g-flex-align-center" style={{ gap: 8 }}>
            <FieldTypeIcon type={record.type} />
            <span>{value}</span>
          </div>
          {record.primary_key && (
            <Tooltip title={intl.get('Global.primaryKey')}>
              <IconFont type="icon-dip-color-primary-key" />
            </Tooltip>
          )}
          {record.display_key && (
            <Tooltip title={intl.get('Global.title')}>
              <IconFont type="icon-dip-color-star" />
            </Tooltip>
          )}
          {record.incremental_key && (
            <Tooltip title={intl.get('Global.incrementalKey')}>
              <IconFont type="icon-dip-color-increment" />
            </Tooltip>
          )}
        </div>
      ),
    },
    {
      title: intl.get('Global.displayName'),
      dataIndex: 'display_name',
      key: 'display_name',
      ellipsis: true,
      render: (value: string) => value || '--',
    },
    {
      title: intl.get('Global.description'),
      dataIndex: 'comment',
      key: 'comment',
      ellipsis: true,
      render: (value: string) => value || '--',
    },
  ];

  if (type === OBJECT_PROPERTY_TYPES.DATA_PROPERTY) {
    columns.push({
      title: intl.get('Global.dataViewField'),
      dataIndex: 'mapped_field',
      key: 'mapped_field',
      ellipsis: true,
      render: (_value: string, record: any) => {
        const mapped = (record as any)?.mapped_field;
        if (mapped?.display_name) {
          return (
            <div className={styles['source-cell']}>
              <FieldTypeIcon type={mapped.type} />
              <span>{mapped.display_name}</span>
            </div>
          );
        }
        return '--';
      },
    });
  }

  if (type === OBJECT_PROPERTY_TYPES.LOGIC_PROPERTY) {
    columns.push({
      title: intl.get('Global.resource'),
      dataIndex: 'data_source',
      key: 'data_source',
      ellipsis: true,
      render: (_value: string, record: any) => {
        const dataSource = record?.data_source || {};
        if (dataSource?.name) {
          return (
            <div className={styles['source-cell']}>
              {dataSource.type === OntologyObjectType.LogicAttributeType.METRIC && <IconFont type="icon-dip-color-MCP" style={{ fontSize: '24px' }} />}
              {dataSource.type === OntologyObjectType.LogicAttributeType.OPERATOR && <IconFont type="icon-dip-color-tool" style={{ fontSize: '24px' }} />}
              <span>{dataSource.name}</span>
            </div>
          );
        }
        return '--';
      },
    });
  }

  const { primary_keys = [], display_key = '', incremental_key } = source || {};
  const dataProperties =
    source?.data_properties?.map((item) => ({
      ...item,
      primary_key: primary_keys.includes(item.name),
      display_key: item.name === display_key,
      incremental_key: item.name === incremental_key,
    })) || [];
  const logicProperties =
    source?.logic_properties?.map((item) => ({
      ...item,
      primary_key: primary_keys.includes(item.name),
      display_key: item.name === display_key,
    })) || [];
  const currentData: any[] = type === OBJECT_PROPERTY_TYPES.DATA_PROPERTY ? dataProperties : logicProperties;
  const dataSource: ObjectType.Field[] = currentData.filter((val: { name: string }) => val.name.includes(searchText));

  const fieldList =
    source?.data_properties?.map((item) => ({
      displayName: item.display_name || item.name,
      name: item.name,
      type: item.type || '',
    })) || [];

  const dataSearchableKeys = dataColumns.map((column) => String(column.dataIndex || '')).filter(Boolean);
  const keyword = dataSearchText.trim().toLowerCase();
  const filteredData = !keyword
    ? (data as any[])
    : (data || []).filter((row: any) =>
        dataSearchableKeys.some((key) =>
          String(row?.[key] ?? '')
            .toLowerCase()
            .includes(keyword)
        )
      );
  if (!source) {
    return null;
  }

  return (
    <div className={styles['object-detail-page']}>
      <DetailPageHeader
        onBack={goBack}
        icon={<ObjectIcon icon={source.icon} color={source.color} />}
        title={source.name}
        actions={
          <>
            {isPermission && (
              <Button
                className={styles['top-edit-btn']}
                icon={<IconFont type="icon-dip-bianji" />}
                onClick={() => history.push(`/ontology/object/edit/${source.id}`)}
              >
                {intl.get('Global.edit')}
              </Button>
            )}
            {isPermission && (
              <Dropdown trigger={['click']} menu={{ items: [{ key: 'delete', label: intl.get('Global.delete') }], onClick: onChange }}>
                <Button className={styles['top-more-btn']} icon={<EllipsisOutlined style={{ fontSize: 20 }} />} />
              </Dropdown>
            )}
          </>
        }
      />

      <DetailSummaryCard
        id={source.id}
        icon={source.icon ? <ObjectIcon size={32} iconSize={20} icon={source.icon} color={source.color} /> : null}
        name={source.name}
        tags={tags}
        comment={comment}
        modifier={source.updater?.name}
        updateTime={source.update_time}
        metaLeftItems={[
          {
            icon: <IconFont type="icon-dip-gnfz" />,
            label: intl.get('ConceptGroup.conceptGroup'),
            value: source.concept_groups?.length || 0,
            valueType: 'badge',
          },
          {
            icon: <IconFont type="icon-dip-sydx" />,
            label: intl.get('Global.indexedObject'),
            value: intl.get('Global.indexedObjectStats', {
              count: status?.doc_count || 0,
              size: formatFileSize(status?.storage_size || 0) || '0 B',
            }),
            emphasize: true,
          },
        ]}
      />

      <div className={styles['section-card']}>
        <div className={styles['section-toolbar']}>
          <Segmented<string> value={type} options={OBJECT_PROPERTY_TYPE_OPTIONS} onChange={(value) => setType(value)} />
          <Input.Search
            placeholder={intl.get('Global.searchProperty')}
            size="middle"
            value={searchText}
            style={{ width: 280 }}
            allowClear
            onChange={(e) => setSearchText(e.target.value)}
          />
        </div>
        <Table size="small" rowKey="name" columns={columns as any} scroll={{ y: 400 }} dataSource={dataSource as any[]} />
      </div>

      {data_source?.id && (
        <div className={styles['section-card']}>
          <Title className={styles['section-title']}>{intl.get('Global.data')}</Title>
          <div className={styles['section-toolbar']}>
            <div />
            <div className={styles['data-tools']}>
              <Input.Search
                placeholder={intl.get('Global.search')}
                size="middle"
                value={dataSearchText}
                style={{ width: 240 }}
                allowClear
                onChange={(e) => setDataSearchText(e.target.value)}
              />

              <IconFont
                className={styles['filter-icon']}
                onClick={() => {
                  const next = !switchFilter;
                  setSwitchFilter(next);
                  if (!next) {
                    form.setFieldValue('dataFilter', INIT_FILTER);
                    fetchData();
                  }
                }}
                type="icon-dip-filter-field"
              />
            </div>
          </div>
          {switchFilter && (
            <div className={styles['filter-panel']}>
              <Form form={form}>
                <Form.Item name="dataFilter" style={{ marginBottom: 0 }}>
                  <DataFilter ref={dataFilterRef} fieldList={fieldList} required={true} defaultValue={INIT_FILTER} maxCount={[10, 10, 10]} level={3} isFirst />
                </Form.Item>
              </Form>
              <Button type="primary" className="g-mt-2" onClick={handleFilterSearch}>
                {intl.get('Global.search')}
              </Button>
            </div>
          )}
          <Table size="small" rowKey="_tempId" columns={dataColumns} scroll={{ y: 400, x: '100%' }} dataSource={filteredData} loading={dataLoading} />
        </div>
      )}
      <DeleteConfirmModal
        open={deleteConfirmOpen}
        loading={deleteConfirmLoading}
        objects={source ? [source] : []}
        onCancel={onDeleteConfirmCancel}
        onConfirm={onDeleteConfirmOk}
      />
    </div>
  );
};

export default Detail;
