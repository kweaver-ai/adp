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
  const [hoveredEdge, setHoveredEdge] = useState<string | null>(null);
  const [isInitialized, setIsInitialized] = useState(false);
  const [clearSearchTrigger, setClearSearchTrigger] = useState(0);

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
        // 清空搜索内容
        setClearSearchTrigger((prev) => prev + 1);

        return new Promise((resolve, reject) => {
          if (localDataProperties.length === 0) {
            setAlertMessage(intl.get('Object.dataPropertiesRequired'));
            reject(new Error(intl.get('Object.dataPropertiesRequired')));
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
          const hasDisplayKey = localDataProperties.some((p) => p.display_key);

          if (!hasPrimaryKey || !hasDisplayKey) {
            setAlertMessage(intl.get('Object.dataAttributeConfigIncomplete'));
            reject(new Error(intl.get('Object.dataAttributeConfigIncomplete')));
            return;
          }

          if (!hasPrimaryKey) {
            setAlertMessage(intl.get('Object.primaryKeyRequired'));
            reject(new Error(intl.get('Object.primaryKeyRequired')));
            return;
          }

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

  const getDataViewDetail = async (dataViewId: string, shouldAutoInitialize = false) => {
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
        setFields(resetFields);

        // 只有手动选择数据源时才自动初始化
        if (shouldAutoInitialize) {
          const newDataProperties = resetFields.map((field: OntologyObjectType.Field) => ({
            name: field.name,
            display_name: field.display_name,
            type: field.type,
            comment: field.comment,
            primary_key: primaryKeys.includes(field.name) || field.primary_key || false,
            display_key: field.name === displayKey || field.display_key || false,
            incremental_key: field.name === incrementalKey || field.incremental_key || false,
            mapped_field: {
              name: field.name,
              display_name: field.display_name,
              type: field.type,
            },
          }));

          // 合并去重:保留原有的 localDataProperties,只添加新字段
          setLocalDataProperties((prev) => {
            const existingNames = new Set(prev.map((p) => p.name));
            const newProperties = newDataProperties.filter((prop: OntologyObjectType.DataProperty) => !existingNames.has(prop.name));
            return [...prev, ...newProperties];
          });

          // 自动创建连接线(只为新添加的字段创建)
          const existingEdgeTargets = new Set(edges.map((e) => e.targetHandle));
          const newEdges = resetFields
            .filter((field: OntologyObjectType.Field) => !existingEdgeTargets.has(`view-${field.name}`))
            .map((field: OntologyObjectType.Field) => ({
              id: `${field.name}&&${field.name}`,
              type: 'customEdge',
              source: 'data',
              sourceHandle: `data-${field.name}`,
              target: 'view',
              targetHandle: `view-${field.name}`,
              data: { deletable: true },
            }));

          if (newEdges.length > 0) {
            setTimeout(() => {
              setEdges((prev) => [...prev, ...(newEdges as Edge[])]);
            }, EDGE_RENDER_DELAY);
          }
        }
      }
    } catch (event) {
      console.log('getDataViewDetail event: ', event);
    }
  };

  useEffect(() => {
    const initializeData = async () => {
      const hasExistingData = dataProperties?.length > 0;

      if (hasExistingData) {
        const initializedProperties = dataProperties.map((prop) => ({
          ...prop,
          primary_key: primaryKeys.includes(prop.name) || prop.primary_key || false,
          display_key: prop.name === displayKey || prop.display_key || false,
          incremental_key: prop.name === incrementalKey || prop.incremental_key || false,
        }));
        setLocalDataProperties(initializedProperties);

        // 编辑模式下,根据 mapped_field 恢复连线
        if (dataSource?.id) {
          const edgesFromMappedField = initializedProperties
            .filter((prop) => prop.mapped_field)
            .map((prop) => ({
              id: `${prop.name}&&${prop.mapped_field!.name}`,
              type: 'customEdge',
              source: 'data',
              sourceHandle: `data-${prop.name}`,
              target: 'view',
              targetHandle: `view-${prop.mapped_field!.name}`,
              data: { deletable: true },
            }));

          if (edgesFromMappedField.length > 0) {
            setTimeout(() => {
              setEdges(edgesFromMappedField as Edge[]);
            }, EDGE_RENDER_DELAY);
          }
        }
      }
      if (dataSource) {
        setDataViewInfo(dataSource);
        if (dataSource.id) {
          // 组件初始化时只加载 fields,不自动初始化属性
          await getDataViewDetail(dataSource.id, false);
        }
      }
      setIsInitialized(true);
    };

    initializeData();
  }, []);

  const handleOpenDataViewSource = () => {
    setClearSearchTrigger((prev) => prev + 1);
    setOpen(true);
  };

  const handlePickAttribute = () => {
    setClearSearchTrigger((prev) => prev + 1);
    setPickAttributeVisible(true);
  };

  const handleAutoLine = () => {
    setClearSearchTrigger((prev) => prev + 1);
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
        // message.info(intl.get('Object.noMatchingAttributesToConnect'));
        return currentEdges;
      });

      return currentNodes;
    });
  };

  const handleDeleteDataViewSource = () => {
    setClearSearchTrigger((prev) => prev + 1);
    setDataViewInfo(undefined);
    setFields([]);
    setEdges([]);
  };

  const handleAddDataAttribute = () => {
    setClearSearchTrigger((prev) => prev + 1);
    setEditAttrData(undefined);
    setAddAttrVisible(true);
  };

  const handleAddAttrOk = (data: OntologyObjectType.Field) => {
    if (editAttrData) {
      const nameExistsInLocal = localDataProperties.some((p) => p.name === data.name && p.name !== editAttrData.name);
      if (nameExistsInLocal) {
        return {
          name: `${intl.get('Global.attributeName')}「${data.name}」${intl.get('Global.alreadyExists')}`,
        };
      }

      if (data.display_name) {
        const displayNameExistsInLocal = localDataProperties.some((p) => p.display_name === data.display_name && p.name !== editAttrData.name);
        if (displayNameExistsInLocal) {
          return {
            display_name: `${intl.get('Global.displayName')}「${data.display_name}」${intl.get('Global.alreadyExists')}`,
          };
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
              // 保留 mapped_field，因为视图字段不会变
              mapped_field: p.mapped_field,
              index_config: p.index_config,
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

      // 处理边的更新或删除
      if (editAttrData.name !== data.name || editAttrData.type !== data.type) {
        setEdges((prev) => {
          if (editAttrData.type !== data.type) {
            // 类型改变，删除相关的边
            return prev.filter((edge) => !edge.sourceHandle?.includes(editAttrData.name) && !edge.targetHandle?.includes(editAttrData.name));
          }
          // 仅名称改变，更新边的 sourceHandle 和 id
          return prev.map((edge) => {
            const isSourceChanged = edge.sourceHandle?.includes(`data-${editAttrData.name}`);
            const isTargetChanged = edge.targetHandle?.includes(`data-${editAttrData.name}`);

            if (!isSourceChanged && !isTargetChanged) return edge;

            // 更新 sourceHandle（data 侧的属性名变了）
            const newSourceHandle = isSourceChanged ? `data-${data.name}` : edge.sourceHandle;
            // targetHandle 不变（view 侧的字段名没变）
            const newTargetHandle = edge.targetHandle;

            // 从 targetHandle 中提取视图字段名
            const viewFieldName = newTargetHandle?.replace('view-', '') || '';

            return {
              ...edge,
              id: `${data.name}&&${viewFieldName}`,
              sourceHandle: newSourceHandle,
              targetHandle: newTargetHandle,
            };
          });
        });
      }
    } else {
      const nameExistsInLocal = localDataProperties.some((p) => p.name === data.name);

      if (nameExistsInLocal) {
        return {
          name: `${intl.get('Global.attributeName')}「${data.name}」${intl.get('Global.alreadyExists')}`,
        };
      }

      if (data.display_name) {
        const displayNameExistsInFields = fields.some((f) => f.display_name === data.display_name);
        const displayNameExistsInLocal = localDataProperties.some((p) => p.display_name === data.display_name);

        if (displayNameExistsInFields || displayNameExistsInLocal) {
          return {
            display_name: `${intl.get('Global.displayName')}「${data.display_name}」${intl.get('Global.alreadyExists')}`,
          };
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

  const handleDeleteAttribute = (attrName: string) => {
    setLocalDataProperties((prev) => prev.filter((p) => p.name !== attrName));
    setEdges((prev) => prev.filter((edge) => !edge.sourceHandle?.includes(attrName) && !edge.targetHandle?.includes(attrName)));
  };

  const handleTogglePrimaryKey = (attrName: string) => {
    setLocalDataProperties((prev) =>
      prev.map((p) => {
        if (p.name === attrName) {
          return { ...p, primary_key: !p.primary_key };
        }
        return p;
      })
    );
  };

  const handleToggleDisplayKey = (attrName: string) => {
    setLocalDataProperties((prev) =>
      prev.map((p) => {
        if (p.name === attrName) {
          const newDisplayKeyValue = !p.display_key;
          return {
            ...p,
            display_key: newDisplayKeyValue,
          };
        }
        if (p.display_key) {
          return { ...p, display_key: false };
        }
        return p;
      })
    );
  };

  const handleToggleIncrementalKey = (attrName: string) => {
    setLocalDataProperties((prev) =>
      prev.map((p) => {
        if (p.name === attrName) {
          const newIncrementalKeyValue = !p.incremental_key;
          return {
            ...p,
            incremental_key: newIncrementalKeyValue,
          };
        }
        if (p.incremental_key) {
          return { ...p, incremental_key: false };
        }
        return p;
      })
    );
  };

  const handleClearAllAttributes = () => {
    setClearSearchTrigger((prev) => prev + 1);
    setLocalDataProperties([]);
    setEdges([]);
    message.success(intl.get('Global.clearSuccess'));
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
      deleteAttribute: handleDeleteAttribute,
      togglePrimaryKey: handleTogglePrimaryKey,
      toggleDisplayKey: handleToggleDisplayKey,
      toggleIncrementalKey: handleToggleIncrementalKey,
      clearAllAttributes: handleClearAllAttributes,
      clearSearchTrigger,
    });

    // 保留当前节点的位置信息
    setNodes((currentNodes) => {
      return newNodes.map((newNode) => {
        const existingNode = currentNodes.find((node) => node.id === newNode.id);
        if (existingNode) {
          // 如果节点已存在，保留其当前的位置
          return {
            ...newNode,
            position: existingNode.position,
          };
        }
        return newNode;
      });
    });

    if (!isInitialized && edges.length === 0 && newEdges.length > 0) {
      const timer = setTimeout(() => {
        setEdges(newEdges);
      }, EDGE_RENDER_DELAY);
      return () => clearTimeout(timer);
    }
  }, [localDataProperties, fields, logicProperties, dataViewInfo?.id, basicValue, isInitialized, clearSearchTrigger]);

  // 处理节点和连接线高亮效果
  useEffect(() => {
    if (!hoveredEdge) {
      // 清除节点高亮
      setNodes((nds) =>
        nds.map((node) => ({
          ...node,
          data: {
            ...node.data,
            highlightedAttributes: [],
          },
        }))
      );
      // 清除连接线高亮
      setEdges((eds) =>
        eds.map((edge) => ({
          ...edge,
          data: {
            ...edge.data,
            isHovered: false,
          },
        }))
      );
      return;
    }

    // 更新节点高亮
    setNodes((nds) => {
      const edge = edges.find((e) => e.id === hoveredEdge);
      if (!edge) return nds;

      const sourceAttr = edge.sourceHandle?.replace(/^(view|data)-/, '');
      const targetAttr = edge.targetHandle?.replace(/^(view|data)-/, '');

      return nds.map((node) => {
        const highlightedAttributes: string[] = [];
        if (node.id === edge.source && sourceAttr) {
          highlightedAttributes.push(sourceAttr);
        }
        if (node.id === edge.target && targetAttr) {
          highlightedAttributes.push(targetAttr);
        }
        return {
          ...node,
          data: {
            ...node.data,
            highlightedAttributes,
          },
        };
      });
    });

    // 更新连接线高亮
    setEdges((eds) =>
      eds.map((e) => ({
        ...e,
        data: {
          ...e.data,
          isHovered: e.id === hoveredEdge,
        },
      }))
    );
  }, [hoveredEdge]);

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
        const exists = prev.some((edge) => edge.id === `${dataAttr}&&${viewAttr}`);
        if (exists) {
          // message.error(intl.get('Object.onlyOneConnection'));
          return prev;
        }

        const targetAlreadyConnected = prev.some((edge) => edge.sourceHandle === targetHandle || edge.targetHandle === targetHandle);
        const sourceAlreadyConnected = prev.some((edge) => edge.sourceHandle === sourceHandle || edge.targetHandle === sourceHandle);

        if (targetAlreadyConnected && sourceAlreadyConnected) {
          message.error(intl.get('Object.attributeAlreadyConnected'));
          return prev;
        }

        let updatedEdges = prev;
        let oldDataAttrToRemove: string | undefined;

        if (sourceAlreadyConnected) {
          const oldEdge = prev.find((edge) => edge.sourceHandle === sourceHandle || edge.targetHandle === sourceHandle);
          if (oldEdge) {
            const oldConnectionNames = extractConnectionNames(oldEdge.sourceHandle!, oldEdge.targetHandle!);
            oldDataAttrToRemove = oldConnectionNames.dataAttr;
            updatedEdges = prev.filter((edge) => edge.id !== oldEdge.id);
          }
        } else if (targetAlreadyConnected) {
          const oldEdge = prev.find((edge) => edge.sourceHandle === targetHandle || edge.targetHandle === targetHandle);
          if (oldEdge) {
            const oldConnectionNames = extractConnectionNames(oldEdge.sourceHandle!, oldEdge.targetHandle!);
            oldDataAttrToRemove = oldConnectionNames.dataAttr;
            updatedEdges = prev.filter((edge) => edge.id !== oldEdge.id);
          }
        }

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
            if (oldDataAttrToRemove && p.name === oldDataAttrToRemove) {
              const { mapped_field, ...rest } = p;
              return rest;
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
          updatedEdges
        );
      });
    },
    [nodes, setEdges, message]
  );

  const handleChooseOk = (e: OntologyObjectType.DataSource[]) => {
    const dataView = e?.[0] || {};
    setDataViewInfo(dataView);
    // 手动选择数据源时,自动初始化属性和连线
    getDataViewDetail(dataView.id, true);
  };

  const onNodesChange: OnNodesChange = (changes) =>
    setNodes((nds) => {
      const updatedNodes = applyNodeChanges(changes, nds);
      return updatedNodes.map((node) => ({
        ...node,
        type: node.type || 'customNode',
      }));
    });

  const handleEdgeMouseEnter = useCallback((_: React.MouseEvent, edge: Edge) => {
    setHoveredEdge(edge.id);
  }, []);

  const handleEdgeMouseLeave = useCallback(() => {
    setHoveredEdge(null);
  }, []);

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
        mapped_field: {
          name: f.name,
          display_name: f.display_name,
          type: f.type,
        },
      }));
    setLocalDataProperties((prev) => [...newAttributes, ...prev]);

    // 自动创建连接线
    const newEdges = newAttributes.map((attr) => ({
      id: `${attr.name}&&${attr.name}`,
      type: 'customEdge',
      source: 'data',
      sourceHandle: `data-${attr.name}`,
      target: 'view',
      targetHandle: `view-${attr.name}`,
      data: { deletable: true },
    }));
    if (newEdges.length > 0) {
      setTimeout(() => {
        setEdges((prev) => [...prev, ...(newEdges as Edge[])]);
      }, EDGE_RENDER_DELAY);
    }
  };

  const handlePickAttributeClose = () => {
    setPickAttributeVisible(false);
  };

  const availableFields = useMemo(() => {
    const existingPropertyNames = new Set(localDataProperties.map((p) => p.name));
    const existingDisplayNames = new Set(localDataProperties.map((p) => p.display_name).filter(Boolean));
    const connectedFieldNames = new Set(edges.map((edge) => edge.targetHandle?.replace('view-', '')).filter(Boolean));
    return fields.filter((f) => !existingPropertyNames.has(f.name) && !existingDisplayNames.has(f.display_name) && !connectedFieldNames.has(f.name));
  }, [fields, localDataProperties, edges]);

  const selectedRowKeys = useMemo(() => (dataViewInfo?.id ? [dataViewInfo.id] : []), [dataViewInfo?.id]);

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
          onEdgeMouseEnter={handleEdgeMouseEnter}
          onEdgeMouseLeave={handleEdgeMouseLeave}
          nodeTypes={nodeTypes as any}
          edgeTypes={edgeTypes}
          proOptions={{ hideAttribution: true }}
          nodesDraggable={true}
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
        selectedRowKeys={selectedRowKeys}
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
