import { useEffect, useMemo, useRef, useState } from 'react';
import intl from 'react-intl-universal';
import { Button, Input, Table } from 'antd';
import classNames from 'classnames';
import { arNotification } from '@/components/ARNotification';
import FieldFeatureModal from '@/components/FieldFeatureModal';
import { REQUIRED_META_FIELDS } from '@/hooks/useConstants';
import HOOKS from '@/hooks';
import { IconFont, Tooltip } from '@/web-library/common';
import FormHeader from '../FormHeader';
import styles from './index.module.less';
import { useDataViewContext } from '../../../context';
import EmptyForm from '../EmptyForm';

const OutputDataView = () => {
  const { dataViewTotalInfo, setDataViewTotalInfo, selectedDataView, setSelectedDataView } = useDataViewContext();
  const [editingCell, setEditingCell] = useState<{ key: string; field: string } | null>(null);
  const [outputFields, setOutputFields] = useState<any[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [featureModalVisible, setFeatureModalVisible] = useState<boolean>(false);
  const [currentField, setCurrentField] = useState<any>(null);
  const mainBoxRef = useRef<HTMLDivElement>(null);
  const mainBoxSize = HOOKS.useSize(mainBoxRef);
  const tableScrollY = mainBoxSize?.height ? mainBoxSize.height - 160 : 500;

  const filteredDataSource = useMemo(() => {
    if (!searchKeyword) return outputFields;

    const keyword = searchKeyword.toLowerCase();
    return outputFields.filter((item) => item.display_name?.toLowerCase().includes(keyword) || item.name?.toLowerCase().includes(keyword));
  }, [outputFields, searchKeyword]);

  const { updateDataViewNode } = HOOKS.useDataView({
    dataViewTotalInfo,
    setDataViewTotalInfo,
    setSelectedDataView,
  });

  useEffect(() => {
    // 前序节点
    if (selectedDataView?.input_nodes?.length > 0) {
      const nodeList = dataViewTotalInfo?.data_scope || [];
      const outputFields = selectedDataView?.output_fields || [];
      const preNodes = nodeList?.find((item: any) => selectedDataView?.input_nodes?.includes(item.id)) || {};
      setOutputFields([
        ...(outputFields || []).map((item: any) => ({
          ...item,
          selected: true,
        })),
        ...(preNodes?.output_fields || [])
          .filter((item: any) => !outputFields.some((field: any) => field.original_name === item.original_name))
          .map((item: any) => ({
            ...item,
            selected: REQUIRED_META_FIELDS.includes(item.name),
          })),
      ]);
    } else {
      setOutputFields([]);
    }
  }, [selectedDataView, dataViewTotalInfo]);

  const handleFieldChange = (record: any, field: string, value: string) => {
    setOutputFields((prevOutputFields) =>
      prevOutputFields.map((item: any) => (item.original_name === record.original_name ? { ...item, [field]: value } : item))
    );
  };

  const handleSubmit = () => {
    const selectedFields = outputFields.filter((item: any) => item.selected);

    if (!selectedFields?.length) {
      arNotification.error(intl.get('Global.pleaseSelectAtLeastOneField'));
      return;
    }
    const newNodeData = {
      ...selectedDataView,
      output_fields: selectedFields,
      node_status: 'success',
    };
    setLoading(true);
    updateDataViewNode(newNodeData, selectedDataView.id).finally(() => {
      setLoading(false);
    });
  };

  const renderEditableCell = (record: any, field: 'display_name' | 'name' | 'comment') => {
    const isEditing = editingCell?.key === record.original_name && editingCell?.field === field;
    const value = record[field];

    if (isEditing) {
      return (
        <Tooltip title={value}>
          <Input
            value={value}
            onChange={(e) => {
              handleFieldChange(record, field, e.target.value);
            }}
            onBlur={() => setEditingCell(null)}
            autoFocus
            maxLength={field === 'comment' ? 1000 : 255}
          />
        </Tooltip>
      );
    }

    return (
      <Tooltip title={value}>
        <div className={styles.fieldName} onClick={() => setEditingCell({ key: record.original_name, field })}>
          {value || '--'}
        </div>
      </Tooltip>
    );
  };

  const columns = [
    {
      title: intl.get('Global.fieldBusinessName'),
      dataIndex: 'display_name',
      key: 'display_name',
      width: 200,
      render: (_: any, record: any) => renderEditableCell(record, 'display_name'),
    },
    {
      title: intl.get('Global.fieldTechnicalName'),
      dataIndex: 'name',
      key: 'name',
      width: 200,
      render: (_: any, record: any) => renderEditableCell(record, 'name'),
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
      render: (_: any, record: any) => renderEditableCell(record, 'comment'),
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
              <span key={type} className={classNames(styles.featureType, styles[type])}>
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
    selectedRowKeys: outputFields?.filter((item: any) => item.selected).map((item: any) => item.original_name),
    getCheckboxProps: (record: any) => ({
      disabled: REQUIRED_META_FIELDS.includes(record.name),
    }),
    onChange: (selectedRowKeys: React.Key[]) => {
      setOutputFields(
        outputFields?.map((item: any) => ({
          ...item,
          selected: selectedRowKeys.includes(item.original_name),
        })) || []
      );
    },
  };

  return (
    <div className={styles.mainBox} ref={mainBoxRef}>
      <FormHeader
        title={intl.get('CustomDataView.outputView')}
        icon="icon-dip-color-shuchushitu"
        showSubmitButton={outputFields?.length > 0}
        onSubmit={handleSubmit}
        onCancel={() => setSelectedDataView(null)}
        loading={loading}
      />
      <div className={styles.contentBox}>
        <div className={styles.headerBox}>
          <div className={styles.titleBox}>
            <span>{intl.get('Global.fieldList')}</span>
          </div>
          <Input.Search
            style={{ width: '272px' }}
            placeholder={intl.get('Global.searchFieldPlaceholder')}
            onChange={(e) => setSearchKeyword(e.target.value)}
            onSearch={setSearchKeyword}
            allowClear
          />
        </div>
        {outputFields?.length ? (
          <Table
            rowKey={(record) => `${record.original_name}`}
            rowSelection={rowSelection}
            dataSource={filteredDataSource || []}
            columns={columns}
            pagination={false}
            scroll={{ y: tableScrollY }}
          />
        ) : (
          <EmptyForm />
        )}
      </div>
      <FieldFeatureModal
        visible={featureModalVisible}
        mode="view"
        fieldName={currentField?.display_name || currentField?.name}
        data={currentField?.features || []}
        fields={outputFields}
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

export default OutputDataView;
