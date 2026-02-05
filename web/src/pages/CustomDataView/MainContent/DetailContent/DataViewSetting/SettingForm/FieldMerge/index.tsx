import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import intl from 'react-intl-universal';
import { ExclamationCircleOutlined } from '@ant-design/icons';
import { Input, Popover, Radio, RadioChangeEvent, Table, Tooltip, message } from 'antd';
import { ColumnsType } from 'antd/es/table';
import { DataViewQueryType } from '@/components/CustomDataViewSource';
import FieldSelect from '@/components/FieldSelect';
import { getObjectArrayIntersectionByKeys, nanoId } from '@/utils/dataView';
import HOOKS from '@/hooks';
import { IconFont } from '@/web-library/common';
import FormHeader from '../FormHeader';
import styles from './index.module.less';
import { useDataViewContext } from '../../../context';

enum OutType {
  DISTINCT = 'distinct',
  ALL = 'all',
}

const FiledMerge = () => {
  const { dataViewTotalInfo, setDataViewTotalInfo, selectedDataView, setSelectedDataView } = useDataViewContext();
  const [outType, setOutType] = useState<OutType>(OutType.ALL);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const mainContainerRef = useRef<HTMLDivElement>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const tableBoxRef = useRef<HTMLDivElement>(null);
  const tableBoxSize = HOOKS.useSize(tableBoxRef);
  const tableScrollY = tableBoxSize?.height ? tableBoxSize.height - 54 : 500;

  const { updateDataViewNode } = HOOKS.useDataView({
    dataViewTotalInfo,
    setDataViewTotalInfo,
    setSelectedDataView,
  });

  // 从 dataViewTotalInfo 派生 nodeList
  const nodeList = useMemo(() => {
    return dataViewTotalInfo?.data_scope?.filter((item: any) => selectedDataView?.input_nodes?.includes(item.id)) || [];
  }, [dataViewTotalInfo?.data_scope, selectedDataView?.input_nodes]);

  // 从 selectedDataView 和 nodeList 派生初始的 dataSource 和 selectedFields
  const initialData = useMemo(() => {
    if (dataViewTotalInfo?.query_type === DataViewQueryType.SQL) {
      const unionFields = selectedDataView?.config?.union_fields || [];
      const outputFields = selectedDataView?.output_fields || [];
      const newDataSource: any[] = [];
      const newSelectedFields: any[] = [];

      if (outputFields.length > 0) {
        outputFields.forEach((item: any, index: number) => {
          const rowId = nanoId();
          const dataSourceItem: any = { rowId, outputName: item.name, type: item.type };
          nodeList.forEach((nodeItem: any, nodeIndex: number) => {
            const fieldName = unionFields[nodeIndex][index]?.field;
            dataSourceItem[nodeItem.id] = fieldName || undefined;
            if (fieldName) {
              newSelectedFields.push({ rowId, nodeId: nodeItem.id, fieldName });
            }
          });
          newDataSource.push(dataSourceItem);
        });
      } else {
        const inputFiledArr = nodeList.map((nodeItem: any) => nodeItem.output_fields);
        const intersectionFields = getObjectArrayIntersectionByKeys(inputFiledArr, ['name', 'type']);

        intersectionFields.forEach((item: any) => {
          const rowId = nanoId();
          const dataSourceItem: any = { rowId, outputName: item.name, type: item.type };
          nodeList.forEach((nodeItem: any) => {
            dataSourceItem[nodeItem.id] = item.name;
            newSelectedFields.push({ rowId, nodeId: nodeItem.id, fieldName: item.name });
          });
          newDataSource.push(dataSourceItem);
        });
      }

      return { dataSource: newDataSource, selectedFields: newSelectedFields };
    } else {
      const fieldMergeMap = new Map<string, any[]>();

      nodeList.forEach((nodeItem: any) => {
        nodeItem.output_fields.forEach((fieldItem: any) => {
          if (!fieldMergeMap.has(fieldItem.name)) {
            fieldMergeMap.set(fieldItem.name, []);
          }
          fieldMergeMap.get(fieldItem.name)?.push(fieldItem);
        });
      });

      const outputFieldsMerge = Array.from(fieldMergeMap.entries()).map(([fieldName, fields]) => {
        const firstField = fields[0];
        const allTypesMatch = fields.every((f) => f.type === firstField.type);
        const allCommentsMatch = fields.every((f) => f.comment === firstField.comment);
        const firstFeatures = firstField.features || [];
        const allFeaturesMatch = fields.every((f) => {
          const currentFeatures = f.features || [];
          return JSON.stringify(currentFeatures) === JSON.stringify(firstFeatures);
        });

        return {
          name: fieldName,
          type: firstField.type,
          original_name: firstField.original_name,
          display_name: firstField.display_name,
          comment: allCommentsMatch ? firstField.comment : '',
          features: allFeaturesMatch ? firstField.features : [],
          error: !allTypesMatch,
        };
      });

      return { dataSource: outputFieldsMerge, selectedFields: [] };
    }
  }, [dataViewTotalInfo?.query_type, selectedDataView?.config?.union_fields, selectedDataView?.output_fields, nodeList]);

  // 使用 state 存储可编辑的数据
  const [dataSource, setDataSource] = useState<any[]>(initialData.dataSource);
  const [selectedFields, setSelectedFields] = useState<any[]>(initialData.selectedFields);

  // 当 initialData 变化时同步更新 state
  useEffect(() => {
    setDataSource(initialData.dataSource);
    setSelectedFields(initialData.selectedFields);
  }, [initialData]);

  const handleOutTypeChange = (e: RadioChangeEvent) => {
    setOutType(e.target.value);
  };

  const handleAddRow = () => {
    setDataSource((prev) => [{ rowId: nanoId() }, ...prev]);
  };

  const handleDeleteRow = (rowId: string) => {
    setDataSource((prev) => prev.filter((item: any) => item.rowId !== rowId));
  };

  const handleFieldSelect = useCallback(
    (nodeId: string, value: string, record: any) => {
      const updatedRecord = { ...record, [nodeId]: value };

      if (!value) {
        setSelectedFields((prev) => prev.filter((item) => !(item.rowId === record.rowId && item.nodeId === nodeId)));
      } else {
        if (!updatedRecord.type) {
          const fieldType = nodeList.find((item: any) => item.id === nodeId)?.output_fields?.find((field: any) => field.name === value)?.type;
          if (fieldType) {
            updatedRecord.type = fieldType;
          }
        }

        setSelectedFields((prev) => [
          ...prev.filter((item) => !(item.rowId === record.rowId && item.nodeId === nodeId)),
          { rowId: record.rowId, nodeId, fieldName: value },
        ]);
      }

      setDataSource((prev) => prev.map((item) => (item.rowId === record.rowId ? updatedRecord : item)));
    },
    [nodeList]
  );

  // 检查一行中所有字段的类型是否一致
  const checkTypeConsistency = useCallback(
    (record: any) => {
      const types: string[] = [];
      nodeList.forEach((node: any) => {
        const fieldName = record[node.id];
        if (fieldName) {
          const field = node.output_fields?.find((f: any) => f.name === fieldName);
          if (field?.type) {
            types.push(field.type);
          }
        }
      });
      // 如果有多个不同的类型,返回 false
      return types.length === 0 || types.every((t) => t === types[0]);
    },
    [nodeList]
  );

  const radioOptions = [
    {
      label: (
        <div className={styles.radioItem}>
          <span>{intl.get('CustomDataView.FieldMerge.keepAllRows')}</span>
          <Tooltip title={intl.get('CustomDataView.FieldMerge.keepAllRowsTooltip')}>
            <IconFont type="icon-dip-tishi" />
          </Tooltip>
        </div>
      ),
      value: OutType.ALL,
    },
    {
      label: (
        <div className={styles.radioItem}>
          <span>{intl.get('CustomDataView.FieldMerge.removeDuplicates')}</span>
          <Tooltip title={intl.get('CustomDataView.FieldMerge.removeDuplicatesTooltip')}>
            <IconFont type="icon-dip-tishi" />
          </Tooltip>
        </div>
      ),
      value: OutType.DISTINCT,
    },
  ];

  const validateParams = () => {
    const newErrors: Record<string, string> = {};

    dataSource.forEach((record) => {
      if (!record.outputName?.trim()) {
        newErrors[`${record.rowId}_outputName`] = intl.get('CustomDataView.FieldMerge.outputFieldNameRequired');
      }

      nodeList.forEach((node: any) => {
        if (!record[node.id]?.trim()) {
          newErrors[`${record.rowId}_${node.id}`] = intl.get('CustomDataView.FieldMerge.nodeFieldRequired');
        }
      });
    });

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = () => {
    if (dataViewTotalInfo?.query_type === DataViewQueryType.SQL) {
      if (dataSource.length === 0) {
        message.error(intl.get('CustomDataView.FieldMerge.addMergeRule'));
        return;
      }
      const valid = validateParams();

      if (!valid) {
        message.error(intl.get('CustomDataView.FieldMerge.completeMergeConfig'));
        return;
      }

      // 检查类型一致性
      const hasTypeError = dataSource.some((record) => !checkTypeConsistency(record));
      if (hasTypeError) {
        message.error(intl.get('CustomDataView.FieldMerge.completeMergeConfig'));
        return;
      }

      const union_fields = nodeList.map((node: any) =>
        dataSource.map((record) => (record[node.id] ? { field: record[node.id], value_from: 'field' } : null)).filter(Boolean)
      );

      const outputFields = dataSource.map((record) => {
        const firstNodeField = nodeList[0]?.output_fields?.find((f: any) => f.name === record[nodeList[0].id]);
        return {
          name: record.outputName,
          original_name: record.outputName,
          display_name: record.outputName,
          type: record.type,
          comment: firstNodeField?.comment || '',
          features: firstNodeField?.features || [],
        };
      });

      const newNodeData = {
        ...selectedDataView,
        config: {
          ...selectedDataView?.config,
          union_type: outType,
          union_fields: union_fields,
        },
        output_fields: outputFields,
        node_status: 'success',
      };
      setLoading(true);
      updateDataViewNode(newNodeData, selectedDataView.id).finally(() => {
        setLoading(false);
      });
    } else {
      const newNodeData = {
        ...selectedDataView,
        config: {
          ...selectedDataView?.config,
          union_type: outType,
        },
        output_fields: dataSource.filter((item) => !item.error),
        node_status: 'success',
      };
      setLoading(true);
      updateDataViewNode(newNodeData, selectedDataView.id).finally(() => {
        setLoading(false);
      });
    }
  };

  const columns: ColumnsType<any> = useMemo(() => {
    if (dataViewTotalInfo?.query_type === DataViewQueryType.SQL) {
      return [
        {
          title: intl.get('CustomDataView.FieldMerge.outputFieldName'),
          dataIndex: 'outputName',
          key: 'outputName',
          fixed: 'left',
          width: 280,
          render: (_text: any, record: any, _index: any) => {
            const hasTypeError = !checkTypeConsistency(record);
            return (
              <div className={styles.resultItem}>
                <Tooltip title={hasTypeError ? intl.get('CustomDataView.FieldMerge.fieldTypeInconsistent') : undefined}>
                  <Input
                    placeholder={intl.get('CustomDataView.FieldMerge.outputFieldName')}
                    maxLength={255}
                    style={{ width: '100%' }}
                    value={record.outputName}
                    onChange={(e) => {
                      const updatedRecord = { ...record, outputName: e.target.value };
                      setDataSource((prev) => prev.map((item) => (item.rowId === record.rowId ? updatedRecord : item)));
                    }}
                    status={errors[`${record.rowId}_outputName`] || hasTypeError ? 'error' : undefined}
                  />
                </Tooltip>
                <div className={styles.rowNameSplit} />
              </div>
            );
          },
        },
        ...nodeList.map((nodeItem: any, nodeIndex: any) => ({
          title: (
            <div>
              <IconFont type="icon-dip-color-zuzhijiegou2" style={{ marginRight: 4 }} />
              {nodeItem.title}
            </div>
          ),
          dataIndex: `${nodeItem.id}`,
          key: `${nodeItem.id}`,
          ellipsis: true,
          width: 280,
          render: (_: any, record: any, index: any) => {
            return (
              <div className={styles.resultItem}>
                <FieldSelect
                  value={record[`${nodeItem.id}`]}
                  onChange={(value) => {
                    handleFieldSelect(nodeItem.id, value, record);
                  }}
                  style={{ width: '100%' }}
                  placeholder={intl.get('Global.pleaseSelect')}
                  allowClear
                  fields={nodeItem?.output_fields || []}
                  getOptionDisabled={(field) => selectedFields.some((f) => f.nodeId === nodeItem.id && f.fieldName === field.name && f.rowId !== record.rowId)}
                  status={errors[`${record.rowId}_${nodeItem.id}`] ? 'error' : undefined}
                  getPopupContainer={(): HTMLElement => mainContainerRef.current as HTMLElement}
                />
                <div className={styles.rowSelectRightSplit} hidden={nodeIndex === nodeList.length - 1} />
              </div>
            );
          },
        })),
        {
          title: '',
          fixed: 'right',
          key: 'action',
          width: 60,
          render: (_text: any, record: any, _index: any) => {
            return (
              <div className={styles.actionBox}>
                <IconFont type="icon-dip-trash" onClick={() => handleDeleteRow(record.rowId)} />
                <IconFont type="icon-dip-add" onClick={() => handleAddRow()} />
              </div>
            );
          },
        },
      ];
    } else {
      return [
        {
          title: intl.get('Global.fieldName'),
          dataIndex: 'name',
          key: 'name',
          ellipsis: true,
          render: (text: string, record: any) => {
            return (
              <div>
                {record.error ? (
                  <div className={styles.errorBox}>
                    <span>{text}</span>
                    <Popover
                      placement="topRight"
                      getPopupContainer={() => mainContainerRef.current as HTMLElement}
                      title={intl.get('Global.tipTitle')}
                      content={intl.get('CustomDataView.FieldMerge.fieldTypeConflictTip')}
                    >
                      <ExclamationCircleOutlined />
                    </Popover>
                  </div>
                ) : (
                  <span>{text}</span>
                )}
              </div>
            );
          },
        },
        {
          title: intl.get('Global.fieldDisplayName'),
          dataIndex: 'display_name',
          key: 'display_name',
          ellipsis: true,
          render: (text: string) => (
            <Tooltip title={text?.length > 20 ? text : undefined}>
              <div className={styles.fieldName}>{text}</div>
            </Tooltip>
          ),
        },
        {
          title: intl.get('Global.fieldType'),
          dataIndex: 'type',
          key: 'type',
        },
      ];
    }
  }, [dataViewTotalInfo, nodeList, selectedFields, dataSource, errors, checkTypeConsistency, handleFieldSelect]);

  return (
    <div className={styles.mainBox} ref={mainContainerRef}>
      <FormHeader
        title={intl.get('CustomDataView.OperateBox.dataMerge')}
        icon="icon-dip-color-shujuhebingsuanzi"
        onSubmit={handleSubmit}
        onCancel={() => setSelectedDataView(null)}
        loading={loading}
      />
      <div className={styles.contentBox}>
        {dataViewTotalInfo?.query_type === DataViewQueryType.SQL && (
          <div className={styles.topBox}>
            <div className={styles.addItemBtn}>
              <div className={styles.addItemBtnContent} onClick={handleAddRow}>
                <IconFont type="icon-dip-add" />
                <span>{intl.get('CustomDataView.FieldMerge.addMergeRule')}</span>
              </div>
              <div className={styles.tipBox}>
                <IconFont type="icon-dip-color-guizetishi" />
                <span>{intl.get('CustomDataView.FieldMerge.sameTypeFieldsOnly')}</span>
              </div>
            </div>
            <div className={styles.outTypeBox}>
              <span>{intl.get('CustomDataView.FieldMerge.mergeOutputType')}：</span>
              <Radio.Group className={styles.radioGroup} onChange={handleOutTypeChange} value={outType} options={radioOptions}></Radio.Group>
            </div>
          </div>
        )}

        <div className={styles.mergeTableBox} ref={tableBoxRef}>
          {dataSource?.length > 0 ? (
            <Table rowKey="rowId" columns={columns} scroll={{ y: tableScrollY }} dataSource={dataSource} pagination={false} />
          ) : (
            <div className={styles.emptyText}>{intl.get('CustomDataView.FieldMerge.noMergeRules')}</div>
          )}
        </div>
      </div>
    </div>
  );
};

export default FiledMerge;
