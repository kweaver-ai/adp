import { useEffect, useState, useRef, useMemo } from 'react';
import intl from 'react-intl-universal';
import { CheckOutlined } from '@ant-design/icons';
import { Button, Input, Table } from 'antd';
import classnames from 'classnames';
import { arNotification } from '@/components/ARNotification';
import FieldFeatureModal from '@/components/FieldFeatureModal';
import FieldSelect from '@/components/FieldSelect';
import HOOKS from '@/hooks';
import { IconFont, Tooltip } from '@/web-library/common';
import FormHeader from '../FormHeader';
import styles from './index.module.less';
import { useDataViewContext } from '../../../context';

enum JoinType {
  LEFT = 'left',
  RIGHT = 'right',
  INNER = 'inner',
  FULL_OUTER = 'full outer',
}

const FieldJoin = () => {
  const { dataViewTotalInfo, setDataViewTotalInfo, selectedDataView, setSelectedDataView } = useDataViewContext();
  const [joinType, setJoinType] = useState(JoinType.LEFT);
  const [dataSource, setDataSource] = useState<any[]>([]);
  const [leftField, setLeftField] = useState<string | undefined>(undefined);
  const [rightField, setRightField] = useState<string | undefined>(undefined);
  const [leftNode, setLeftNode] = useState<any>({});
  const [rightNode, setRightNode] = useState<any>({});
  const [editingKey, setEditingKey] = useState('');
  const [outputFields, setOutputFields] = useState<any[]>([]);
  const [searchKeyword, setSearchKeyword] = useState('');
  const rightBoxRef = useRef<HTMLDivElement>(null);
  const rightBoxSize = HOOKS.useSize(rightBoxRef);
  const tableScrollY = rightBoxSize?.height ? rightBoxSize.height - 100 : 500;
  const [loading, setLoading] = useState<boolean>(false);
  const [featureModalVisible, setFeatureModalVisible] = useState<boolean>(false);
  const [currentField, setCurrentField] = useState<any>(null);

  const joinTypeOptions = [
    {
      label: intl.get('CustomDataView.FieldJoin.leftJoin'),
      type: JoinType.LEFT,
      icon: 'icon-dip-color-zuolianjie',
    },
    {
      label: intl.get('CustomDataView.FieldJoin.rightJoin'),
      type: JoinType.RIGHT,
      icon: 'icon-dip-color-youlianjie',
    },
    {
      label: intl.get('CustomDataView.FieldJoin.innerJoin'),
      type: JoinType.INNER,
      icon: 'icon-dip-color-neilianjie1',
    },
    {
      label: intl.get('CustomDataView.FieldJoin.fullOuterJoin'),
      type: JoinType.FULL_OUTER,
      icon: 'icon-dip-color-quanwailianjie',
    },
  ];

  const { updateDataViewNode } = HOOKS.useDataView({
    dataViewTotalInfo,
    setDataViewTotalInfo,
    setSelectedDataView,
  });

  useEffect(() => {
    // 前序节点
    if (selectedDataView?.input_nodes?.length > 0) {
      const nodeList = dataViewTotalInfo?.data_scope || [];
      const preNodes = nodeList.filter((item: any) => selectedDataView?.input_nodes?.includes(item.id));
      setLeftNode(preNodes?.[0] || {});
      setRightNode(preNodes?.[1] || {});
      const leftDataSource =
        preNodes?.[0]?.output_fields?.map((item: any) => ({
          ...item,
          position: 'left',
          src_node_id: preNodes?.[0]?.id || '',
          src_node_name: preNodes?.[0]?.title || '',
        })) || [];
      const rightDataSource =
        preNodes?.[1]?.output_fields?.map((item: any) => {
          let display_name = item.display_name;
          let name = item.name;

          //display_name有重名的则修改 字段名+源节点名
          if (leftDataSource?.some((item: any) => item.display_name === display_name)) {
            display_name = `${display_name}_${preNodes?.[1]?.title || ''}`;
          }

          // name 有重名的则修改 字段名+源节点名
          if (leftDataSource?.some((item: any) => item.name === name)) {
            name = `${name}_${preNodes?.[1]?.title || ''}`;
          }
          return {
            ...item,
            display_name,
            name,
            position: 'right',
            src_node_id: preNodes?.[1]?.id || '',
            src_node_name: preNodes?.[1]?.title || '',
          };
        }) || [];
      setDataSource([...leftDataSource, ...rightDataSource]);
      const outputFields = selectedDataView?.output_fields || [];
      setOutputFields(outputFields);
      const config = selectedDataView?.config || {};
      setJoinType(config?.join_type || JoinType.LEFT);
      setLeftField(config?.join_on?.[0]?.left_field || undefined);
      setRightField(config?.join_on?.[0]?.right_field || undefined);
    }
  }, [selectedDataView, dataViewTotalInfo]);

  const filteredDataSource = useMemo(() => {
    if (!searchKeyword) return dataSource;

    const keyword = searchKeyword.toLowerCase();
    return dataSource.filter((item) => item.display_name?.toLowerCase().includes(keyword) || item.name?.toLowerCase().includes(keyword));
  }, [outputFields, searchKeyword]);

  const handleFieldChange = (record: any, field: string, value: string) => {
    setDataSource(
      dataSource.map((item: any) =>
        item.original_name === record.original_name && item.src_node_name === record.src_node_name ? { ...item, [field]: value } : item
      ) || []
    );
  };

  const renderEditableCell = (record: any, field: 'display_name' | 'comment') => {
    const isEditing = editingKey === record.original_name + record.src_node_name;
    const value = record[field];

    if (isEditing) {
      return (
        <Tooltip title={value}>
          <Input
            value={value}
            onChange={(e) => {
              handleFieldChange(record, field, e.target.value);
            }}
            onBlur={() => setEditingKey('')}
            autoFocus
            maxLength={255}
          />
        </Tooltip>
      );
    }

    return (
      <Tooltip title={value}>
        <div className={styles.fieldName} onClick={() => setEditingKey(record.original_name + record.src_node_name)}>
          {value || '--'}
        </div>
      </Tooltip>
    );
  };

  const handleSubmit = () => {
    if (!leftField || !rightField) {
      arNotification.error(intl.get('CustomDataView.FieldJoin.selectJoinFields'));
      return;
    }

    const leftFieldType = leftNode?.output_fields?.find((item: any) => item.name === leftField)?.type || '';
    const rightFieldType = rightNode?.output_fields?.find((item: any) => item.name === rightField)?.type || '';
    if (leftFieldType !== rightFieldType) {
      arNotification.error(intl.get('CustomDataView.FieldJoin.joinFieldTypeMismatch'));
      return;
    }

    if (outputFields?.length === 0) {
      arNotification.error(intl.get('Global.pleaseSelectAtLeastOneField'));
      return;
    }

    const newNodeData = {
      ...selectedDataView,
      config: {
        ...selectedDataView?.config,
        join_type: joinType,
        join_on: [
          {
            left_field: leftField,
            right_field: rightField,
            operator: '=',
          },
        ],
      },
      output_fields: outputFields,
      node_status: 'success',
    };
    setLoading(true);
    updateDataViewNode(newNodeData, selectedDataView.id).finally(() => {
      setLoading(false);
    });
  };

  const columns = [
    {
      title: intl.get('Global.fieldBusinessName'),
      dataIndex: 'display_name',
      key: 'display_name',
      width: 200,
      ellipsis: true,
      render: (_: any, record: any) => renderEditableCell(record, 'display_name'),
    },
    {
      title: intl.get('CustomDataView.FieldJoin.sourceNode'),
      dataIndex: 'src_node_name',
      key: 'src_node_name',
      width: 150,
      render: (_: any, record: any) => (
        <div className={styles.srcNodeBox}>
          <IconFont type={record.position === 'left' ? 'icon-dip-color-zuobiao' : 'icon-dip-color-youbiao'} style={{ fontSize: 16 }} />
          <span>{record.src_node_name}</span>
        </div>
      ),
    },
    {
      title: intl.get('Global.fieldTechnicalName'),
      dataIndex: 'original_name',
      key: 'original_name',
      width: 150,
      ellipsis: true,
    },
    {
      title: intl.get('Global.fieldType'),
      dataIndex: 'type',
      key: 'type',
      width: 120,
    },
    {
      title: intl.get('Global.fieldComment'),
      dataIndex: 'comment',
      key: 'comment',
      width: 200,
      ellipsis: true,
      render: (comment: string) => {
        return (
          <Tooltip title={comment}>
            <div className={styles.fieldName}>{comment || '--'}</div>
          </Tooltip>
        );
      },
    },
    {
      title: intl.get('Global.fieldFeatureType'),
      dataIndex: 'features',
      key: 'features_type',
      width: 150,
      render: (features: any[]) => {
        if (!features || features.length === 0) {
          return <span style={{ color: 'rgba(0, 0, 0, 0.25)' }}>{intl.get('Global.unset')}</span>;
        }
        const uniqueTypes = Array.from(new Set(features.map((item) => item.type)));
        return (
          <div className={styles.featureTypeContainer}>
            {uniqueTypes.map((type) => (
              <span key={type} className={classnames(styles.featureType, styles[type])}>
                {type}
              </span>
            ))}
          </div>
        );
      },
    },
    {
      title: () => (
        <div>
          <span style={{ marginRight: 8 }}>{intl.get('Global.fieldFeature')}</span>
          <Tooltip title={intl.get('Global.fieldFeatureTip')}>
            <IconFont type="icon-dip-color-tip" className={styles.helpIcon} />
          </Tooltip>
        </div>
      ),
      dataIndex: 'features',
      key: 'features',
      width: 120,
      align: 'center' as const,
      render: (_: unknown, record: any) => (
        <Button
          type="link"
          onClick={(): void => {
            setCurrentField(record);
            setFeatureModalVisible(true);
          }}
          disabled={!record.features || record.features.length === 0}
        >
          {intl.get('Global.view')}
        </Button>
      ),
    },
  ];

  const rowSelection: any = {
    selectedRowKeys: outputFields.map((item: any) => item.name),
    onChange: (selectedRowKeys: React.Key[]) => {
      setOutputFields(dataSource?.filter((item: any) => selectedRowKeys.includes(item.name)) || []);
    },
  };

  const JoinTypeItem = ({ item, isActive }: { item: { type: JoinType; label: string; icon: string }; isActive: boolean }) => {
    return (
      <div className={classnames(styles.joinTypeItem, { [styles.active]: isActive })} onClick={() => setJoinType(item.type)}>
        {isActive && (
          <div className={styles.joinTypeCheck}>
            <CheckOutlined className={styles.joinTypeCheckIcon} />
          </div>
        )}
        <IconFont type={item.icon} style={{ fontSize: 20 }} />
        <span>{item.label}</span>
      </div>
    );
  };

  return (
    <div className={styles.mainBox}>
      <FormHeader
        title={intl.get('CustomDataView.OperateBox.dataRelation')}
        icon="icon-dip-color-shujuguanliansuanzi"
        onSubmit={handleSubmit}
        onCancel={() => setSelectedDataView(null)}
        loading={loading}
      />
      <div className={styles.configBox}>
        <div className={styles.leftBox}>
          <div className={styles.headerBox}>
            <span>{intl.get('Global.config')}</span>
          </div>
          <div className={styles.joinConfigBox}>
            <div className={styles.configTitle}>
              <span className={styles.required}>*</span>
              <span>{intl.get('CustomDataView.FieldJoin.selectJoinFields')}</span>
            </div>
            <div className={styles.joinSelectTitle}>
              <div className={styles.joinNodeBox}>
                <IconFont type="icon-dip-color-zuobiao" />
                <IconFont type="icon-dip-color-zuzhijiegou2" />
                <span className={styles.joinNodeTitle}>{leftNode?.title || ''}</span>
              </div>
              {/* <IconFont type="icon-dip-qiehuan" /> */}
              <div className={styles.joinNodeBox}>
                <IconFont type="icon-dip-color-youbiao" />
                <IconFont type="icon-dip-color-zuzhijiegou2" />
                <span className={styles.joinNodeTitle}>{rightNode?.title || ''}</span>
              </div>
            </div>
            <div className={styles.joinSelelctBox}>
              <FieldSelect
                placeholder={intl.get('Global.pleaseSelect')}
                value={leftField}
                onChange={setLeftField}
                style={{ width: 215 }}
                fields={leftNode?.output_fields || []}
              />
              <div className={styles.line}></div>
              <FieldSelect
                placeholder={intl.get('Global.pleaseSelect')}
                value={rightField}
                onChange={setRightField}
                style={{ width: 215 }}
                fields={rightNode?.output_fields || []}
              />
            </div>
          </div>
          <div className={styles.joinConfigBox}>
            <div className={styles.configTitle}>
              <span className={styles.required}>*</span>
              <span>{intl.get('CustomDataView.FieldJoin.joinType')}</span>
            </div>
            <div className={styles.joinTypeBox}>
              {joinTypeOptions.map((item) => (
                <JoinTypeItem item={item} isActive={joinType === item.type} key={item.type} />
              ))}
            </div>
          </div>
        </div>
        <div className={styles.rightBox} ref={rightBoxRef}>
          <div className={styles.headerBox}>
            <div className={styles.titleBox}>
              <span>{intl.get('CustomDataView.FieldJoin.fieldList')}</span>
              <span className={styles.helpText}>{intl.get('CustomDataView.FieldJoin.fieldListHelp')}</span>
            </div>
            <Input.Search
              style={{ width: '272px' }}
              placeholder={intl.get('Global.searchFieldPlaceholder')}
              onChange={(e) => setSearchKeyword(e.target.value)}
              onSearch={setSearchKeyword}
              allowClear
            />
          </div>
          <Table rowKey="name" scroll={{ y: tableScrollY }} rowSelection={rowSelection} columns={columns} dataSource={filteredDataSource} pagination={false} />
        </div>
      </div>
      <FieldFeatureModal
        visible={featureModalVisible}
        mode="view"
        fieldName={currentField?.display_name || currentField?.name}
        data={currentField?.features || []}
        fields={dataSource}
        onCancel={(): void => {
          setFeatureModalVisible(false);
          setCurrentField(null);
        }}
        onOk={(): void => {
          setFeatureModalVisible(false);
          setCurrentField(null);
        }}
      />
    </div>
  );
};

export default FieldJoin;
