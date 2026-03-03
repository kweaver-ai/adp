import { useState, useEffect } from 'react';
import intl from 'react-intl-universal';
import { useHistory, useLocation, useParams } from 'react-router-dom';
import { EllipsisOutlined } from '@ant-design/icons';
import { Table, Dropdown, Tooltip } from 'antd';
import { map, uniqueId } from 'lodash-es';
import { showDeleteConfirm } from '@/components/DeleteConfirm';
import DetailPageHeader from '@/components/DetailPageHeader';
import DetailSummaryCard from '@/components/DetailSummaryCard';
import FieldTypeIcon from '@/components/FieldTypeIcon';
import ObjectIcon from '@/components/ObjectIcon';
import { baseConfig } from '@/services/request';
import ENUMS from '@/enums';
import HOOKS from '@/hooks';
import SERVICE from '@/services';
import { Text, Title, Button, IconFont } from '@/web-library/common';
import styles from './index.module.less';

const ObjectItem = (props: any) => {
  const { value, icon, color } = props;
  return (
    <div className="g-flex-align-center" style={{ gap: '8px' }} title={value}>
      {icon && <ObjectIcon icon={icon} color={color} />}
      <div>
        <Text className="g-ellipsis-1">{value}</Text>
      </div>
    </div>
  );
};

const FieldItem = (props: any) => {
  const { property } = props;
  if (!property) return <span>{intl.get('Global.emptyValue')}</span>;

  const displayName = property.display_name || property.name;
  return (
    <div className="g-flex-align-center" style={{ gap: '8px' }}>
      {property?.type && (
        <div style={{ flexShrink: 0 }}>
          <FieldTypeIcon type={property.type} />
        </div>
      )}
      <Tooltip title={displayName?.length > 20 ? displayName : undefined}>
        <Text className="g-ellipsis-1">{displayName}</Text>
      </Tooltip>
    </div>
  );
};

const uniqueIdPrefix = 'edgeDetail';
const EdgeDetail = () => {
  const history = useHistory();
  const location = useLocation<{ isPermission?: boolean }>();
  const { id: detailId = '' } = useParams<{ id?: string }>();
  const { modal, message } = HOOKS.useGlobalContext();

  const finalKnId = localStorage.getItem('KnowledgeNetwork.id') || sessionStorage.getItem('knId') || '';
  const finalIsPermission = location.state?.isPermission ?? true;
  const [source, setSource] = useState<any>({});
  const { id, tags, name, comment } = source;

  const goBack = () => {
    history.goBack();
  };

  useEffect(() => {
    baseConfig?.toggleSideBarShow(false);
    return () => {
      baseConfig?.toggleSideBarShow(true);
    };
  }, []);

  const getDetail = async () => {
    if (!finalKnId || !detailId) return;
    const result = await SERVICE.edge.getEdgeDetail(finalKnId, detailId);
    if (result[0]) setSource(result[0]);
  };

  useEffect(() => {
    if (!detailId) return;
    getDetail();
  }, [detailId, finalKnId]);

  const deleteInPageMode = async () => {
    if (!finalKnId || !source?.id) return;
    await SERVICE.edge.deleteEdge(finalKnId, [source.id]);
    message.success(intl.get('Global.deleteSuccess'));
    goBack();
  };

  /** 下来菜单变更 */
  const onChange = (data: any) => {
    if (data.key === 'delete') {
      showDeleteConfirm(modal, {
        content: intl.get('Global.deleteConfirm', { name: `「${source?.name || ''}」` }),
        onOk: deleteInPageMode,
      });
    }
  };

  const onEdit = () => {
    if (!source?.id) return;
    history.push(`/ontology/edge/edit/${source.id}`);
  };

  const getDirectData = () => {
    const { mapping_rules, source_object_type, source_object_type_id, target_object_type, target_object_type_id } = source;
    const directColumns: any = [
      { dataIndex: 'name', width: 100, render: (value: string) => <Title>{value}</Title> },
      {
        title: intl.get('Edge.detailSourcePoint'),
        dataIndex: 'source',
        width: 415,
        render: (value: string, data: any) => {
          if (data?.type === 'property') return <FieldItem property={data?.source_property} />;
          if (data?.type === 'object') {
            const { icon, color } = data?.source_object_type || {};
            return <ObjectItem value={value} icon={icon} color={color} />;
          }
        },
      },
      {
        title: intl.get('Edge.detailTargetPoint'),
        dataIndex: 'target',
        width: 415,
        render: (value: string, data: any) => {
          if (data?.type === 'property') return <FieldItem property={data?.target_property} />;
          if (data?.type === 'object') {
            const { icon, color } = data?.target_object_type || {};
            return <ObjectItem value={value} icon={icon} color={color} />;
          }
        },
      },
    ];
    const directDataSource = [
      {
        id: uniqueId(uniqueIdPrefix),
        name: intl.get('Global.objectClass'),
        type: 'object',
        source_object_type,
        target_object_type,
        source: source_object_type?.display_name || source_object_type?.name || source_object_type_id,
        target: target_object_type?.display_name || target_object_type?.name || target_object_type_id,
      },
      ...map(mapping_rules, (item, index) => {
        const { source_property, target_property } = item;
        return {
          id: index,
          type: 'property',
          name: intl.get('Global.dataProperty'),
          source_property,
          target_property,
          source: source_property?.display_name || source_property?.name,
          target: target_property?.display_name || target_property?.name,
        };
      }),
    ];
    return { columns: directColumns, dataSource: directDataSource };
  };

  const getDataViewData = () => {
    const { mapping_rules, source_object_type, source_object_type_id, target_object_type, target_object_type_id } = source;
    const { backing_data_source, source_mapping_rules, target_mapping_rules } = mapping_rules || {};
    const viewColumns: any = [
      {
        title: intl.get('Edge.detailSourceObject'),
        dataIndex: 'source',
        width: 415,
        render: (value: string, data: any) => {
          if (data?.type === 'property') return <FieldItem property={data?.source_property} />;
          if (data?.type === 'object') {
            const { icon, color } = data?.source_object_type || {};
            return <ObjectItem value={value} icon={icon} color={color} />;
          }
        },
      },
      {
        title: intl.get('Global.dataView'),
        dataIndex: 'dataView',
        width: 415,
        render: (value: string, data: any) => {
          if (data?.type === 'object') return value || intl.get('Global.emptyValue');
          if (data?.type === 'property')
            return (
              <div className="g-flex-align-center">
                <div className="g-w-100 g-flex-align-center" title={data?.dataView?.source}>
                  <div className={styles['view-data-option-label']}>{intl.get('Edge.detailDataViewSource')}</div>
                  <Text className="g-ellipsis-1">{data?.dataView?.source}</Text>
                </div>
                <div className="g-w-100 g-flex-align-center" title={data?.dataView?.target}>
                  <div className={styles['view-data-option-label']} style={{ background: '#000' }}>
                    {intl.get('Edge.detailDataViewTarget')}
                  </div>
                  <Text className="g-ellipsis-1">{data?.dataView?.target}</Text>
                </div>
              </div>
            );
        },
      },
      {
        title: intl.get('Edge.detailTargetObject'),
        dataIndex: 'target',
        width: 415,
        render: (value: string, data: any) => {
          if (data?.type === 'property') return <FieldItem property={data?.target_property} />;
          if (data?.type === 'object') {
            const { icon, color } = data?.target_object_type || {};
            return <ObjectItem value={value} icon={icon} color={color} />;
          }
        },
      },
    ];
    const viewDataSource = [
      {
        id: uniqueId(uniqueIdPrefix),
        type: 'object',
        source_object_type,
        target_object_type,
        source: source_object_type?.display_name || source_object_type?.name || source_object_type_id,
        target: target_object_type?.display_name || target_object_type?.name || target_object_type_id,
        dataView: backing_data_source?.display_name || backing_data_source?.name || backing_data_source?.id,
      },
      ...map(source_mapping_rules, (item, index) => {
        const targetMapping = target_mapping_rules?.[index];
        return {
          id: index,
          type: 'property',
          source_property: item?.source_property,
          target_property: targetMapping?.target_property,
          source: item?.source_property?.display_name || item?.source_property?.name,
          target: targetMapping?.target_property?.display_name || targetMapping?.target_property?.name,
          dataView: {
            source: item?.target_property?.display_name || item?.target_property?.name,
            target: targetMapping?.source_property?.display_name || targetMapping?.source_property?.name,
          },
        };
      }),
    ];
    return { columns: viewColumns, dataSource: viewDataSource };
  };

  let detailTableData: any = { columns: [], dataSource: [] };
  if (source?.type === ENUMS.EDGE.TYPE_DIRECT) detailTableData = getDirectData();
  if (source?.type === ENUMS.EDGE.TYPE_DATA_VIEW) detailTableData = getDataViewData();
  const { columns, dataSource } = detailTableData;

  const summaryCard = (
    <DetailSummaryCard
      id={id}
      icon={<ObjectIcon size={32} iconSize={20} icon="icon-dip-guanxilei" color="#5381DF" />}
      name={name}
      tags={tags}
      comment={comment}
      modifier={source?.updater?.name}
      updateTime={source?.update_time}
    />
  );

  const mainContent = (
    <div className={styles['section-card']}>
      <Title className={styles['section-title']}>{intl.get('Global.config')}</Title>
      <Table size="small" rowKey="id" bordered columns={columns} dataSource={dataSource} pagination={false} />
    </div>
  );

  return (
    <div className={styles['edge-detail-page']}>
      <DetailPageHeader
        onBack={goBack}
        icon={<ObjectIcon icon="icon-dip-guanxilei" color="#5381DF" />}
        title={name || '--'}
        actions={
          <>
            {finalIsPermission && (
              <Button className={styles['top-edit-btn']} icon={<IconFont type="icon-dip-bianji" />} onClick={onEdit}>
                {intl.get('Global.edit')}
              </Button>
            )}
            {finalIsPermission && (
              <Dropdown trigger={['click']} menu={{ items: [{ key: 'delete', label: intl.get('Global.delete') }], onClick: onChange }}>
                <Button className={styles['top-more-btn']} icon={<EllipsisOutlined style={{ fontSize: 20 }} />} />
              </Dropdown>
            )}
          </>
        }
      />
      {summaryCard}
      {mainContent}
    </div>
  );
};

export default EdgeDetail;
