import { forwardRef, useCallback, useEffect, useImperativeHandle, useMemo, useState } from 'react';
import intl from 'react-intl-universal';
import { ReactFlow, addEdge, useNodesState, useEdgesState, Edge, Connection, applyNodeChanges, type OnNodesChange, Controls } from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Alert, Popover, Tooltip } from 'antd';
import { nanoid } from 'nanoid';
import { DataViewSource } from '@/components/DataViewSource';
import FieldTypeIcon from '@/components/FieldTypeIcon';
import ObjectIcon from '@/components/ObjectIcon';
import * as OntologyObjectType from '@/services/object/type';
import HOOKS from '@/hooks';
import SERVICE from '@/services';
import { IconFont } from '@/web-library/common';
import AddDataAttribute from './AddDataAttribute';
import CustomEdge from './customEdge';
import CustomNode from './customNode';
import styles from './index.module.less';
import PickAttribute from './PickAttribute';
import { transformCanvasData } from './utils';

const nodeTypes = {
  customNode: CustomNode,
};

const edgeTypes = {
  customEdge: CustomEdge,
};

const EDGE_RENDER_DELAY = 150;

const getHandleName = (handle: string, prefix: string) => handle.replace(new RegExp(`^${prefix}-`), '');

const isConnectionValid = (sourceHandle: string, targetHandle: string) => {
  const isViewToData = sourceHandle.startsWith('view') && targetHandle.startsWith('data');
  const isDataToView = sourceHandle.startsWith('data') && targetHandle.startsWith('view');
  return isViewToData || isDataToView;
};

const extractConnectionNames = (sourceHandle: string, targetHandle: string) => {
  const isViewToData = sourceHandle.startsWith('view');
  const viewAttr = isViewToData ? getHandleName(sourceHandle, 'view') : getHandleName(targetHandle, 'view');
  const dataAttr = isViewToData ? getHandleName(targetHandle, 'data') : getHandleName(sourceHandle, 'data');
  return { viewAttr, dataAttr };
};

interface TProps {
  dataProperties: OntologyObjectType.DataProperty[];
  logicProperties: OntologyObjectType.LogicProperty[];
  dataSource?: OntologyObjectType.DataSource;
  basicValue?: OntologyObjectType.BasicInfo;
  primaryKeys?: string[];
  displayKey?: string;
  incrementalKey?: string;
}

const DataAttribute = forwardRef((props: TProps, ref) => {
  const {
    dataProperties,
    logicProperties,
    dataSource,
    basicValue = { name: '', id: '', icon: 'icon-color-rectangle', color: '#126EE3' },
    primaryKeys = [],
    displayKey,
    incrementalKey,
  } = props;
  const { message } = HOOKS.useGlobalContext();

  const [nodes, setNodes] = useNodesState<any>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
  const [fields, setFields] = useState<OntologyObjectType.Field[]>([]);
  const [addAttrVisible, setAddAttrVisible] = useState(false);
  const [editAttrData, setEditAttrData] = useState<OntologyObjectType.Field>();
  const [localDataProperties, setLocalDataProperties] = useState<OntologyObjectType.DataProperty[]>([]);
  const [open, setOpen] = useState(false);
  const [dataViewInfo, setDataViewInfo] = useState<OntologyObjectType.DataSource | undefined>(undefined);
  const [pickAttributeVisible, setPickAttributeVisible] = useState(false);
  const [alertMessage, setAlertMessage] = useState<string>('');

  // 自定义边变化处理，同步更新 localDataProperties 的 mapped_field
  const handleEdgesChange = useCallback(
    (changes: any[]) => {
      // 检查是否有边被删除
      const removedEdges = changes.filter((change) => change.type === 'remove');
      if (removedEdges.length > 0) {
        // 获取被删除边的 id 列表
        const removedEdgeIds = removedEdges.map((change) => {
          const edge = edges.find((e) => e.id === change.id);
          return edge?.id;
        });

        // 清除对应属性的 mapped_field
        setLocalDataProperties((prevProps) =>
          prevProps.map((p) => {
            // 检查这个属性的连线是否被删除
            const hasRemovedEdge = removedEdgeIds.some((edgeId) => edgeId && edgeId.startsWith(`${p.name}&&`));
            if (hasRemovedEdge) {
              const { mapped_field, ...rest } = p;
              return rest;
            }
            return p;
          })
        );
      }

      // 调用原始的 onEdgesChange
      onEdgesChange(changes);
    },
    [edges, onEdgesChange]
  );

  useImperativeHandle(
    ref,
    () => ({
      validateFields: () => {
        console.log('localDataProperties', localDataProperties);
        return new Promise((resolve, reject) => {
          if (localDataProperties.length === 0) {
            resolve({
              dataProperties: localDataProperties,
              dataSource: dataViewInfo,
            });
            return;
          }

          const namePattern = /^[a-z0-9][a-z0-9_-]*$/;
          const invalidProperty = localDataProperties.find((p) => !namePattern.test(p.name));
          if (invalidProperty) {
            setAlertMessage(`${intl.get('Global.attributeName')}  ${intl.get('Global.idPatternError')}`);
            reject(new Error(intl.get('Global.idPatternError')));
            return;
          }

          const hasPrimaryKey = localDataProperties.some((p) => p.primary_key);
          if (!hasPrimaryKey) {
            setAlertMessage(intl.get('Object.primaryKeyRequired'));
            reject(new Error(intl.get('Object.primaryKeyRequired')));
            return;
          }

          const hasDisplayKey = localDataProperties.some((p) => p.display_key);
          if (!hasDisplayKey) {
            setAlertMessage(intl.get('Object.displayKeyRequired'));
            reject(new Error(intl.get('Object.displayKeyRequired')));
            return;
          }

          resolve({
            dataProperties: localDataProperties,
            dataSource: dataViewInfo,
          });
        });
      },
      getDataProperties: () => {
        return new Promise((resolve) => {
          resolve({
            dataProperties: localDataProperties,
            dataSource: dataViewInfo,
          });
        });
      },
    }),

    [localDataProperties, edges, logicProperties, dataViewInfo, message]
  );

  const getDataViewDetail = async (dataViewId: string) => {
    if (!dataViewId) return;
    try {
      const result = await SERVICE.dataView.getDataViewDetail(dataViewId);
      if (result?.[0]) {
        const fields = result?.[0]?.fields || [];
        const resetFields = fields
          .filter((item: OntologyObjectType.Field) => !item.name.startsWith('_'))
          .map((item: OntologyObjectType.Field) => ({
            ...item,
            id: nanoid(),
            name: item.name.replace(/\./g, '_'),
          }));
        setEdges([]);
        setFields(resetFields);
      }
    } catch (event) {
      console.log('getDataViewDetail event: ', event);
    }
  };

  useEffect(() => {
    const initializeData = async () => {
      if (dataProperties?.length > 0) {
        const initializedProperties = dataProperties.map((prop) => ({
          ...prop,
          primary_key: primaryKeys.includes(prop.name) || prop.primary_key || false,
          display_key: prop.name === displayKey || prop.display_key || false,
          incremental_key: prop.name === incrementalKey || prop.incremental_key || false,
        }));
        setLocalDataProperties(initializedProperties);
      }
      if (dataSource) {
        setDataViewInfo(dataSource);
        if (dataSource.id) {
          await getDataViewDetail(dataSource.id);
        }
      }
    };

    initializeData();
  }, []);

  const handleOpenDataViewSource = () => {
    setOpen(true);
  };

  const handlePickAttribute = () => {
    setPickAttributeVisible(true);
  };

  const handleAutoLine = () => {
    setNodes((currentNodes) => {
      const viewNode = currentNodes.find((node) => node.id === 'view');
      const dataNode = currentNodes.find((node) => node.id === 'data');

      if (!viewNode || !dataNode) return currentNodes;

      const viewAttrs = viewNode.data.attributes;
      const dataAttrs = dataNode.data.attributes;

      setEdges((currentEdges) => {
        const connectedAttrs = new Set<string>();
        currentEdges.forEach((edge) => {
          if (edge.sourceHandle) {
            connectedAttrs.add(getHandleName(edge.sourceHandle, '(view|data)'));
          }
          if (edge.targetHandle) {
            connectedAttrs.add(getHandleName(edge.targetHandle, '(view|data)'));
          }
        });

        const newEdges: Edge[] = [];
        const newMappings: Array<{ dataAttr: string; viewAttr: any }> = [];

        dataAttrs.forEach((dataAttr: { name: string; display_name: string; type: string }) => {
          if (connectedAttrs.has(dataAttr.name)) return;

          const matchedViewAttr = viewAttrs.find(
            (viewAttr: { name: string; display_name: string; type: string }) =>
              viewAttr.name === dataAttr.name && viewAttr.display_name === dataAttr.display_name && viewAttr.type === dataAttr.type
          );

          if (matchedViewAttr && !connectedAttrs.has(matchedViewAttr.name)) {
            newEdges.push({
              id: `${dataAttr.name}&&${matchedViewAttr.name}`,
              type: 'customEdge',
              source: 'data',
              sourceHandle: `data-${dataAttr.name}`,
              target: 'view',
              targetHandle: `view-${matchedViewAttr.name}`,
              data: { deletable: true },
            } as Edge);
            connectedAttrs.add(dataAttr.name);
            connectedAttrs.add(matchedViewAttr.name);
            newMappings.push({ dataAttr: dataAttr.name, viewAttr: matchedViewAttr });
          }
        });

        if (newEdges.length > 0) {
          // 更新 localDataProperties 中的 mapped_field
          setLocalDataProperties((prevProps) =>
            prevProps.map((p) => {
              const mapping = newMappings.find((m) => m.dataAttr === p.name);
              if (mapping) {
                return {
                  ...p,
                  mapped_field: {
                    name: mapping.viewAttr.name,
                    display_name: mapping.viewAttr.display_name,
                    type: mapping.viewAttr.type,
                  },
                };
              }
              return p;
            })
          );

          message.success(`${intl.get('Global.add')}${newEdges.length}条连线`);
          return [...currentEdges, ...newEdges];
        }
        message.info(intl.get('Object.noMatchingAttributesToConnect'));
        return currentEdges;
      });

      return currentNodes;
    });
  };

  const handleDeleteDataViewSource = () => {
    setDataViewInfo(undefined);
    setFields([]);
    setEdges([]);
  };

  const handleAddDataAttribute = () => {
    setEditAttrData(undefined);
    setAddAttrVisible(true);
  };

  const handleAddAttrOk = (data: OntologyObjectType.Field) => {
    if (editAttrData) {
      const nameExistsInLocal = localDataProperties.some((p) => p.name === data.name && p.name !== editAttrData.name);
      if (nameExistsInLocal) {
        message.error(`${intl.get('Global.attributeName')}「${data.name}」${intl.get('Global.alreadyExists')}`);
        return;
      }

      if (data.display_name) {
        const displayNameExistsInLocal = localDataProperties.some((p) => p.display_name === data.display_name && p.name !== editAttrData.name);
        if (displayNameExistsInLocal) {
          message.error(`${intl.get('Global.displayName')}「${data.display_name}」${intl.get('Global.alreadyExists')}`);
          return;
        }
      }

      setLocalDataProperties((prev) =>
        prev.map((p) => {
          if (p.name === editAttrData.name) {
            return {
              name: data.name,
              display_name: data.display_name,
              type: data.type,
              comment: data.comment,
              primary_key: data.primary_key || false,
              display_key: data.display_key || false,
              incremental_key: data.incremental_key || false,
              mapped_field: p.mapped_field,
            };
          }
          if (data.display_key && p.display_key) {
            return { ...p, display_key: false };
          }
          if (data.incremental_key && p.incremental_key) {
            return { ...p, incremental_key: false };
          }
          return p;
        })
      );

      if (editAttrData.type !== data.type) {
        // 类型改变，删除相关的边
        setEdges((prev) => prev.filter((edge) => !edge.sourceHandle?.includes(editAttrData.name) && !edge.targetHandle?.includes(editAttrData.name)));
      } else if (editAttrData.name !== data.name || editAttrData.display_name !== data.display_name) {
        // 只改名称，不改类型，更新边的引用
        setEdges((prev) =>
          prev.map((edge) => {
            const newEdge = { ...edge };
            if (edge.sourceHandle?.includes(editAttrData.name)) {
              newEdge.sourceHandle = edge.sourceHandle.replace(editAttrData.name, data.name);
            }
            if (edge.targetHandle?.includes(editAttrData.name)) {
              newEdge.targetHandle = edge.targetHandle.replace(editAttrData.name, data.name);
            }
            if (edge.id?.includes(editAttrData.name)) {
              newEdge.id = edge.id.replace(editAttrData.name, data.name);
            }
            return newEdge;
          })
        );
      }

      message.success(intl.get('Global.saveSuccess'));
    } else {
      const nameExistsInFields = fields.some((f) => f.name === data.name);
      const nameExistsInLocal = localDataProperties.some((p) => p.name === data.name);

      if (nameExistsInFields || nameExistsInLocal) {
        message.error(`${intl.get('Global.attributeName')}「${data.name}」${intl.get('Global.alreadyExists')}`);
        return;
      }

      if (data.display_name) {
        const displayNameExistsInFields = fields.some((f) => f.display_name === data.display_name);
        const displayNameExistsInLocal = localDataProperties.some((p) => p.display_name === data.display_name);

        if (displayNameExistsInFields || displayNameExistsInLocal) {
          message.error(`${intl.get('Global.displayName')}「${data.display_name}」${intl.get('Global.alreadyExists')}`);
          return;
        }
      }

      const newProperty: OntologyObjectType.DataProperty = {
        name: data.name,
        display_name: data.display_name,
        type: data.type,
        comment: data.comment,
        primary_key: data.primary_key || false,
        display_key: data.display_key || false,
        incremental_key: data.incremental_key || false,
      };

      setLocalDataProperties((prev) => {
        const updatedPrev = prev.map((p) => {
          const updates: Partial<OntologyObjectType.DataProperty> = {};
          if (data.display_key && p.display_key) {
            updates.display_key = false;
          }
          if (data.incremental_key && p.incremental_key) {
            updates.incremental_key = false;
          }
          return Object.keys(updates).length > 0 ? { ...p, ...updates } : p;
        });
        return [newProperty, ...updatedPrev];
      });

      message.success(intl.get('Global.saveSuccess'));
    }
    setAddAttrVisible(false);
    setEditAttrData(undefined);
  };

  const handleAddAttrClose = () => {
    setAddAttrVisible(false);
    setEditAttrData(undefined);
  };

  const handleAddAttrDelete = (data: OntologyObjectType.Field) => {
    setLocalDataProperties((prev) => prev.filter((p) => p.name !== data.name));
    setEdges((prev) => prev.filter((edge) => !edge.sourceHandle?.includes(data.name) && !edge.targetHandle?.includes(data.name)));
    setAddAttrVisible(false);
    setEditAttrData(undefined);
    message.success(intl.get('Global.deleteSuccess'));
  };

  const attrClick = (val: OntologyObjectType.Field) => {
    const fullPropertyData = localDataProperties.find((p) => p.name === val.name);
    setEditAttrData(
      fullPropertyData
        ? {
            ...fullPropertyData,
            id: val.id,
            error: val.error || {},
          }
        : val
    );
    setAddAttrVisible(true);
  };

  useEffect(() => {
    const { nodes: newNodes, edges: newEdges } = transformCanvasData({
      dataProperties: localDataProperties,
      logicProperties,
      fields,
      dataSource: dataViewInfo,
      basicValue,
      openDataViewSource: handleOpenDataViewSource,
      deleteDataViewSource: handleDeleteDataViewSource,
      addDataAttribute: handleAddDataAttribute,
      pickAttribute: handlePickAttribute,
      autoLine: handleAutoLine,
      attrClick,
    });
    setNodes(newNodes);

    if (edges.length === 0 && newEdges.length > 0) {
      const timer = setTimeout(() => {
        setEdges(newEdges);
      }, EDGE_RENDER_DELAY);
      return () => clearTimeout(timer);
    }
  }, [localDataProperties, fields, logicProperties, dataViewInfo?.id, basicValue]);

  const onConnect = useCallback(
    (params: Connection) => {
      const { sourceHandle, targetHandle } = params;
      if (!sourceHandle || !targetHandle) return;

      if (!isConnectionValid(sourceHandle, targetHandle)) return;

      const { viewAttr, dataAttr } = extractConnectionNames(sourceHandle, targetHandle);

      const viewNode = nodes.find((node) => node.id === 'view');
      const dataNode = nodes.find((node) => node.id === 'data');

      const viewAttrObj = viewNode?.data.attributes.find((attr: { name: string }) => attr.name === viewAttr);
      const dataAttrObj = dataNode?.data.attributes.find((attr: { name: string }) => attr.name === dataAttr);

      if (viewAttrObj?.type !== dataAttrObj?.type) {
        message.error(intl.get('Object.attributeTypeInconsistent'));
        return;
      }

      setEdges((prev) => {
        // 检查是否已存在相同的连接
        const exists = prev.some((edge) => edge.id === `${dataAttr}&&${viewAttr}`);
        if (exists) {
          message.error(intl.get('Object.onlyOneConnection'));
          return prev;
        }

        // 1对1限制：检查 sourceHandle 是否已被使用
        const sourceAlreadyConnected = prev.some((edge) => edge.sourceHandle === sourceHandle || edge.targetHandle === sourceHandle);
        if (sourceAlreadyConnected) {
          message.error(intl.get('Object.attributeAlreadyConnected'));
          return prev;
        }

        // 1对1限制：检查 targetHandle 是否已被使用
        const targetAlreadyConnected = prev.some((edge) => edge.sourceHandle === targetHandle || edge.targetHandle === targetHandle);
        if (targetAlreadyConnected) {
          message.error(intl.get('Object.attributeAlreadyConnected'));
          return prev;
        }

        // 更新 localDataProperties 中的 mapped_field
        setLocalDataProperties((prevProps) =>
          prevProps.map((p) => {
            if (p.name === dataAttr) {
              return {
                ...p,
                mapped_field: {
                  name: viewAttr,
                  display_name: viewAttrObj?.display_name,
                  type: viewAttrObj?.type,
                },
              };
            }
            return p;
          })
        );

        return addEdge(
          {
            ...params,
            id: `${dataAttr}&&${viewAttr}`,
            type: 'customEdge',
            data: { deletable: true },
          },
          prev
        );
      });
    },
    [nodes, setEdges, message]
  );

  const handleChooseOk = (e: OntologyObjectType.DataSource[]) => {
    const dataView = e?.[0] || {};
    setDataViewInfo(dataView);
    getDataViewDetail(dataView.id);
  };

  const onNodesChange: OnNodesChange = (changes) =>
    setNodes((nds) => {
      const updatedNodes = applyNodeChanges(changes, nds);
      return updatedNodes.map((node) => ({
        ...node,
        type: node.type || 'customNode',
      }));
    });

  const handlePickAttributeOk = (targetKeys: string[]) => {
    setPickAttributeVisible(false);
    const existingNames = localDataProperties.map((p) => p.name);
    const newAttributes = fields
      .filter((f) => targetKeys.includes(f.name) && !existingNames.includes(f.name))
      .map((f) => ({
        name: f.name,
        display_name: f.display_name,
        type: f.type,
        comment: f.comment,
      }));
    setLocalDataProperties((prev) => [...newAttributes, ...prev]);
  };

  const handlePickAttributeClose = () => {
    setPickAttributeVisible(false);
  };

  const availableFields = useMemo(() => {
    const connectedFieldNames = new Set<string>();
    edges.forEach((edge) => {
      if (edge.sourceHandle) {
        const fieldName = edge.sourceHandle.replace(/^(view|data)-/, '');
        connectedFieldNames.add(fieldName);
      }
      if (edge.targetHandle) {
        const fieldName = edge.targetHandle.replace(/^(view|data)-/, '');
        connectedFieldNames.add(fieldName);
      }
    });
    return fields.filter((f) => !connectedFieldNames.has(f.name));
  }, [fields, edges]);

  return (
    <div className={styles['data-attribute-root']}>
      {alertMessage && <Alert message={alertMessage} type="error" closable onClose={() => setAlertMessage('')} banner />}
      <div className={styles['object-info-bar']}>
        <div className={styles['object-info-content']}>
          <ObjectIcon icon={basicValue?.icon || 'icon-color-rectangle'} color={basicValue?.color} size={28} iconSize={20} />
          <div className={styles['object-name']}>{basicValue?.name || intl.get('Object.objectName')}</div>
        </div>
        <div className={styles['object-info-right']}>
          <div className={styles['info-item']}>
            <IconFont type="icon-dip-color-primary-key" />
            <span className={styles['info-label']}>{intl.get('Global.primaryKey')}</span>
            <Tooltip title={intl.get('Object.primaryKeyTip')}>
              <IconFont type="icon-dip-color-tip" className={styles.helpIcon} />
            </Tooltip>
            <span className={styles['info-label']}>:</span>
            {localDataProperties.filter((p) => p.primary_key).length > 1 ? (
              <Popover
                content={
                  <div className={styles['key-list-popover']}>
                    {localDataProperties
                      .filter((p) => p.primary_key)
                      .map((item) => (
                        <div key={item.name} className={styles['key-list-item']}>
                          {item.type && <FieldTypeIcon type={item.type} />}
                          <span className={styles['key-item-text']}>{item.display_name}</span>
                        </div>
                      ))}
                  </div>
                }
                trigger="hover"
                placement="bottomRight"
                overlayClassName={styles['key-list-popover-wrapper']}
              >
                <span className={styles['count-badge']}>{localDataProperties.filter((p) => p.primary_key).length}</span>
              </Popover>
            ) : (
              <span className={localDataProperties.find((p) => p.primary_key) ? styles['info-value'] : styles['info-value-empty']}>
                {localDataProperties.find((p) => p.primary_key)?.display_name || intl.get('Global.notConfigured')}
              </span>
            )}
          </div>

          <div className={styles['divider']} />
          <div className={styles['info-item']}>
            <IconFont type="icon-dip-color-star" />
            <span className={styles['info-label']}>{intl.get('Global.title')}</span>
            <Tooltip title={intl.get('Object.displayKeyTip')}>
              <IconFont type="icon-dip-color-tip" className={styles.helpIcon} />
            </Tooltip>
            <span className={styles['info-label']}>:</span>
            <span className={localDataProperties.find((p) => p.display_key) ? styles['info-value'] : styles['info-value-empty']}>
              {localDataProperties.find((p) => p.display_key)?.display_name || intl.get('Global.notConfigured')}
            </span>
          </div>
          <div className={styles['divider']} />
          <div className={styles['info-item']}>
            <IconFont type="icon-dip-color-increment" />
            <span className={styles['info-label']}>{intl.get('Object.incremental')}</span>
            <Tooltip title={intl.get('Object.incrementalKeyTip')}>
              <IconFont type="icon-dip-color-tip" className={styles.helpIcon} />
            </Tooltip>
            <span className={styles['info-label']}>:</span>
            <span className={localDataProperties.find((p) => p.incremental_key) ? styles['info-value'] : styles['info-value-empty']}>
              {localDataProperties.find((p) => p.incremental_key)?.display_name || intl.get('Global.notConfigured')}
            </span>
          </div>
        </div>
      </div>
      <div style={{ flex: 1, position: 'relative' }}>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={handleEdgesChange}
          onConnect={onConnect}
          nodeTypes={nodeTypes as any}
          edgeTypes={edgeTypes}
          proOptions={{ hideAttribution: true }}
          nodesDraggable={false}
          nodesConnectable={true}
          nodesFocusable={false}
          edgesFocusable={true}
          elementsSelectable={true}
          edgesReconnectable={false}
          panOnDrag={true}
          zoomOnScroll={true}
          zoomOnPinch={true}
          zoomOnDoubleClick={false}
          preventScrolling={false}
          minZoom={0.3}
          maxZoom={2}
          defaultViewport={{ x: 0, y: 0, zoom: 1 }}
          onKeyDown={(e) => {
            if (e.target instanceof HTMLInputElement) {
              return;
            }
            if (e.key === 'Backspace' || e.key === 'Delete') {
              e.stopPropagation();
              e.preventDefault();
            }
          }}
          fitViewOptions={{
            minZoom: 0.3,
            maxZoom: 2,
          }}
        >
          <Controls showInteractive={false} position="bottom-right" />
        </ReactFlow>
      </div>
      <AddDataAttribute open={addAttrVisible} data={editAttrData} onClose={handleAddAttrClose} onOk={handleAddAttrOk} onDelete={handleAddAttrDelete} />
      <DataViewSource
        open={open}
        onCancel={() => {
          setOpen(false);
        }}
        selectedRowKeys={[dataViewInfo?.id]}
        maxCheckedCount={1}
        onOk={(checkedList: OntologyObjectType.DataSource[]) => {
          handleChooseOk(checkedList);
          setOpen(false);
        }}
      />
      <PickAttribute visible={pickAttributeVisible} onCancel={handlePickAttributeClose} onOk={handlePickAttributeOk} dataSource={availableFields} />
    </div>
  );
});

export default DataAttribute;
