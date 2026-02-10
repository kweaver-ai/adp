import { forwardRef, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState } from 'react';
import intl from 'react-intl-universal';
import {
  ReactFlow,
  addEdge,
  useNodesState,
  useEdgesState,
  Edge,
  Connection,
  applyNodeChanges,
  type OnNodesChange,
  type EdgeChange,
  type Node,
  type NodeTypes,
  Controls,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Alert, Popover, Tooltip } from 'antd';
import { nanoid } from 'nanoid';
import { DataViewSource } from '@/components/DataViewSource';
import FieldSelect from '@/components/FieldSelect';
import ObjectIcon from '@/components/ObjectIcon';
import * as OntologyObjectType from '@/services/object/type';
import HOOKS from '@/hooks';
import SERVICE from '@/services';
import { IconFont } from '@/web-library/common';
import AddDataAttribute from './AddDataAttribute';
import { canBeDisplayKey, canBeIncrementalKey, canBePrimaryKey } from './constants';
import CustomEdge from './customEdge';
import CustomNode from './customNode';
import { HoveredEdgeIdContext } from './hoverContext';
import styles from './index.module.less';
import PickAttribute from './PickAttribute';
import { makeEdgeId, parseEdgeId, transformCanvasData } from './utils';

type TNodeData = OntologyObjectType.TNode['data'];
type TFlowNode = Node<TNodeData>;

const nodeTypes: NodeTypes = {
  customNode: CustomNode,
};

const edgeTypes = {
  customEdge: CustomEdge,
};

const EDGE_RENDER_DELAY = 150;

const getHandleName = (handle: string, prefix: string) => handle.replace(new RegExp(`^${prefix}-`), '');

type TFieldSelectOption = {
  name: string;
  display_name: string;
  type: string;
  comment?: string;
};

const orderFieldsWithSelectedFirst = (fields: TFieldSelectOption[], selectedNames: string[]) => {
  if (!selectedNames.length || !fields.length) return fields;

  const indexByName = new Map(fields.map((f) => [f.name, f]));
  const selectedSet = new Set(selectedNames);

  const selectedFields = selectedNames.map((name) => indexByName.get(name)).filter(Boolean) as TFieldSelectOption[];
  const restFields = fields.filter((f) => !selectedSet.has(f.name));

  return [...selectedFields, ...restFields];
};

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

  const hashStringFNV1a = (input: string) => {
    let hash = 2166136261;
    for (let i = 0; i < input.length; i++) {
      hash ^= input.charCodeAt(i);
      hash = Math.imul(hash, 16777619);
    }
    return (hash >>> 0).toString(16);
  };

  // 用“内容签名”避免父层同内容新引用导致重复重初始化（排序 + hash，避免超长字符串）
  const initKey = useMemo(() => {
    const tokens: string[] = [];

    tokens.push(basicValue?.id || '');
    tokens.push(dataSource?.id || '');
    tokens.push(displayKey || '');
    tokens.push(incrementalKey || '');

    const primaryKeysSig = [...primaryKeys].sort().join('|');
    tokens.push(primaryKeysSig);

    const sortedProps = [...(dataProperties || [])].sort((a, b) => String(a.name).localeCompare(String(b.name)));
    for (const p of sortedProps) {
      tokens.push(
        [
          p.name,
          p.display_name || '',
          p.type || '',
          p.mapped_field?.name || '',
          p.mapped_field?.type || '',
          p.primary_key ? '1' : '0',
          p.display_key ? '1' : '0',
          p.incremental_key ? '1' : '0',
        ].join(':')
      );
    }

    return hashStringFNV1a(tokens.join('::'));
  }, [basicValue?.id, dataSource?.id, displayKey, incrementalKey, primaryKeys, dataProperties]);

  const [nodes, setNodes] = useNodesState<TFlowNode>([]);
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
  const initSeqRef = useRef(0);
  const [clearSearchTrigger, setClearSearchTrigger] = useState(0);
  const [displayKeyDropdownOpen, setDisplayKeyDropdownOpen] = useState(false);
  const [incrementalKeyDropdownOpen, setIncrementalKeyDropdownOpen] = useState(false);
  const [primaryKeyDropdownOpen, setPrimaryKeyDropdownOpen] = useState(false);

  // 自定义边变化处理，同步更新 localDataProperties 的 mapped_field
  const handleEdgesChange = useCallback(
    (changes: EdgeChange<Edge>[]) => {
      // 检查是否有边被删除
      const removedEdges = changes.filter((change) => change.type === 'remove');
      if (removedEdges.length > 0) {
        // 获取被删除边的 id 列表（edge.id 约定为 `${dataAttr}&&${viewAttr}`）
        const removedEdgeIds = removedEdges.map((change) => change.id).filter(Boolean);
        const removedDataAttrs = new Set(removedEdgeIds.map((id) => parseEdgeId(id).dataAttr).filter(Boolean));

        // 清除对应属性的 mapped_field
        setLocalDataProperties((prevProps) =>
          prevProps.map((p) => {
            // 检查这个属性的连线是否被删除
            const hasRemovedEdge = removedDataAttrs.has(p.name);
            if (hasRemovedEdge) {
              const { mapped_field: _mapped_field, ...rest } = p;
              return rest;
            }
            return p;
          })
        );
      }

      // 调用原始的 onEdgesChange
      onEdgesChange(changes);
    },
    [onEdgesChange]
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

    [localDataProperties, dataViewInfo]
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

          // 自动创建连接线(只为新添加的字段创建) —— 用 setEdges(prev) 避免闭包读到旧 edges
          setTimeout(() => {
            setEdges((prev) => {
              const existingEdgeTargets = new Set(prev.map((e) => e.targetHandle));
              const newEdges = resetFields
                .filter((field: OntologyObjectType.Field) => !existingEdgeTargets.has(`view-${field.name}`))
                .map((field: OntologyObjectType.Field) => ({
                  id: makeEdgeId(field.name, field.name),
                  type: 'customEdge',
                  source: 'data',
                  sourceHandle: `data-${field.name}`,
                  target: 'view',
                  targetHandle: `view-${field.name}`,
                  data: { deletable: true },
                }));
              return newEdges.length > 0 ? [...prev, ...(newEdges as Edge[])] : prev;
            });
          }, EDGE_RENDER_DELAY);
        }
      }
    } catch (event) {
      console.log('getDataViewDetail event: ', event);
    }
  };

  useEffect(() => {
    const initializeData = async () => {
      const seq = ++initSeqRef.current;

      // 重初始化：父层切换对象/数据源时，避免沿用旧状态
      setIsInitialized(false);
      setAlertMessage('');
      setHoveredEdge(null);

      setDataViewInfo(dataSource);
      setFields([]);
      setEdges([]);

      const initializedProperties =
        dataProperties?.length > 0
          ? dataProperties.map((prop) => ({
              ...prop,
              primary_key: primaryKeys.includes(prop.name) || prop.primary_key || false,
              display_key: prop.name === displayKey || prop.display_key || false,
              incremental_key: prop.name === incrementalKey || prop.incremental_key || false,
            }))
          : [];
      setLocalDataProperties(initializedProperties);

      // 编辑模式下,根据 mapped_field 恢复连线
      if (dataSource?.id && initializedProperties.length > 0) {
        const edgesFromMappedField = initializedProperties
          .filter((prop) => prop.mapped_field)
          .map((prop) => ({
            id: makeEdgeId(prop.name, prop.mapped_field!.name),
            type: 'customEdge',
            source: 'data',
            sourceHandle: `data-${prop.name}`,
            target: 'view',
            targetHandle: `view-${prop.mapped_field!.name}`,
            data: { deletable: true },
          }));

        if (edgesFromMappedField.length > 0) {
          setTimeout(() => {
            // 只在仍是最新一轮初始化时生效
            if (seq !== initSeqRef.current) return;
            setEdges(edgesFromMappedField as Edge[]);
          }, EDGE_RENDER_DELAY);
        }
      }

      if (dataSource?.id) {
        // 初始化/重初始化时只加载 fields,不自动初始化属性
        await getDataViewDetail(dataSource.id, false);
      }

      if (seq !== initSeqRef.current) return;
      setIsInitialized(true);
    };

    initializeData();
  }, [initKey]);

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
        type TAttrLite = { name: string; display_name: string; type: string };
        const newMappings: Array<{ dataAttr: string; viewAttr: TAttrLite }> = [];

        (dataAttrs as TAttrLite[]).forEach((dataAttr) => {
          if (connectedAttrs.has(dataAttr.name)) return;

          const matchedViewAttr = viewAttrs.find(
            (viewAttr: TAttrLite) => viewAttr.name === dataAttr.name && viewAttr.display_name === dataAttr.display_name && viewAttr.type === dataAttr.type
          );

          if (matchedViewAttr && !connectedAttrs.has(matchedViewAttr.name)) {
            newEdges.push({
              id: makeEdgeId(dataAttr.name, matchedViewAttr.name),
              type: 'customEdge',
              source: 'data',
              sourceHandle: `data-${dataAttr.name}`,
              target: 'view',
              targetHandle: `view-${matchedViewAttr.name}`,
              data: { deletable: true },
            } as Edge);
            connectedAttrs.add(dataAttr.name);
            connectedAttrs.add(matchedViewAttr.name);
            newMappings.push({ dataAttr: dataAttr.name, viewAttr: matchedViewAttr as TAttrLite });
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
    const buildDuplicateErrors = (ignoreName?: string) => {
      let nameExists = false;
      let displayNameExists = false;

      for (const p of localDataProperties) {
        if (ignoreName && p.name === ignoreName) continue;
        if (p.name === data.name) nameExists = true;
        if (p.display_name === data.display_name) displayNameExists = true;
        if (nameExists && displayNameExists) break;
      }

      if (!nameExists && !displayNameExists) return;

      return {
        ...(nameExists ? { name: `${intl.get('Global.attributeName')}「${data.name}」${intl.get('Global.alreadyExists')}` } : {}),
        ...(displayNameExists ? { display_name: `${intl.get('Global.displayName')}「${data.display_name}」${intl.get('Global.alreadyExists')}` } : {}),
      };
    };

    if (editAttrData) {
      const duplicateErrors = buildDuplicateErrors(editAttrData.name);
      if (duplicateErrors) return duplicateErrors;

      const oldName = editAttrData.name;
      const oldType = editAttrData.type;
      const typeChanged = oldType !== data.type;
      const nameChanged = oldName !== data.name;

      setLocalDataProperties((prev) =>
        prev.map((p) => {
          if (p.name === oldName) {
            const mapped_field = typeChanged ? undefined : p.mapped_field;
            return {
              name: data.name,
              display_name: data.display_name,
              type: data.type,
              comment: data.comment,
              primary_key: data.primary_key || false,
              display_key: data.display_key || false,
              incremental_key: data.incremental_key || false,
              // 类型改变时，清理 mapped_field，避免“无连线但仍有映射”的脏状态
              ...(mapped_field ? { mapped_field } : {}),
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
      if (nameChanged || typeChanged) {
        setEdges((prev) => {
          const isEdgeForOldAttr = (edge: Edge) => {
            if (edge.sourceHandle === `data-${oldName}`) return true;
            return parseEdgeId(edge.id).dataAttr === oldName;
          };

          if (typeChanged) {
            // 类型改变，删除相关的边（精确匹配，避免 includes 误删）
            return prev.filter((edge) => !isEdgeForOldAttr(edge));
          }

          // 仅名称改变，更新边的 sourceHandle 和 id（保持 view 侧字段名不变）
          return prev.map((edge) => {
            if (!isEdgeForOldAttr(edge)) return edge;
            const { viewAttr } = parseEdgeId(edge.id);
            return {
              ...edge,
              id: makeEdgeId(data.name, viewAttr),
              sourceHandle: `data-${data.name}`,
            };
          });
        });
      }
    } else {
      const duplicateErrors = buildDuplicateErrors();
      if (duplicateErrors) return duplicateErrors;

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
    setEdges((prev) => prev.filter((edge) => parseEdgeId(edge.id).dataAttr !== data.name));
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
    setEdges((prev) => prev.filter((edge) => parseEdgeId(edge.id).dataAttr !== attrName));
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

  const handleSelectPrimaryKeys = (attrNames?: string[]) => {
    const selected = new Set(attrNames || []);
    setLocalDataProperties((prev) =>
      prev.map((p) => ({
        ...p,
        primary_key: selected.has(p.name),
      }))
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

  const handleSelectDisplayKey = (attrName?: string) => {
    setLocalDataProperties((prev) =>
      prev.map((p) => {
        if (!attrName) return { ...p, display_key: false };
        if (p.name === attrName) return { ...p, display_key: true };
        if (p.display_key) return { ...p, display_key: false };
        return p;
      })
    );
  };

  const handleSelectIncrementalKey = (attrName?: string) => {
    setLocalDataProperties((prev) =>
      prev.map((p) => {
        if (!attrName) return { ...p, incremental_key: false };
        if (p.name === attrName) return { ...p, incremental_key: true };
        if (p.incremental_key) return { ...p, incremental_key: false };
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
      includeEdges: !isInitialized,
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

    // 保留当前节点的位置信息（O(n)）
    setNodes((currentNodes) => {
      const positionById = new Map(currentNodes.map((node) => [node.id, node.position] as const));
      return newNodes.map((newNode) => {
        const position = positionById.get(newNode.id);
        return position ? { ...newNode, position } : newNode;
      });
    });

    if (!isInitialized && newEdges.length > 0) {
      const timer = setTimeout(() => {
        setEdges((prev) => (prev.length === 0 ? newEdges : prev));
      }, EDGE_RENDER_DELAY);
      return () => clearTimeout(timer);
    }
  }, [localDataProperties, fields, logicProperties, dataViewInfo?.id, basicValue, isInitialized, clearSearchTrigger]);

  // 处理节点和连接线高亮效果
  // 高亮逻辑改为在 CustomNode 内根据 hoveredEdgeId 派生（避免 hover 时 setNodes 引发整批节点更新）

  // 连接校验/映射查找用 Map，避免每次连线都从 nodes 里 find
  const logicNameSet = useMemo(() => new Set(logicProperties.map((p) => p.name)), [logicProperties]);
  const viewFieldByName = useMemo(() => new Map(fields.map((f) => [f.name, f] as const)), [fields]);
  const dataPropertyByName = useMemo(() => {
    const map = new Map<string, OntologyObjectType.DataProperty>();
    for (const p of localDataProperties) {
      if (logicNameSet.has(p.name)) continue;
      map.set(p.name, p);
    }
    return map;
  }, [localDataProperties, logicNameSet]);

  const onConnect = useCallback(
    (params: Connection) => {
      const { sourceHandle, targetHandle } = params;
      if (!sourceHandle || !targetHandle) return;

      if (!isConnectionValid(sourceHandle, targetHandle)) return;

      const { viewAttr, dataAttr } = extractConnectionNames(sourceHandle, targetHandle);

      const viewAttrObj = viewFieldByName.get(viewAttr);
      const dataAttrObj = dataPropertyByName.get(dataAttr);
      if (!viewAttrObj || !dataAttrObj) return;

      if ((viewAttrObj.type || '') !== (dataAttrObj.type || '')) {
        message.error(intl.get('Object.attributeTypeInconsistent'));
        return;
      }

      setEdges((prev) => {
        const exists = prev.some((edge) => edge.id === makeEdgeId(dataAttr, viewAttr));
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
              const { mapped_field: _mapped_field, ...rest } = p;
              return rest;
            }
            return p;
          })
        );

        return addEdge(
          {
            ...params,
            id: makeEdgeId(dataAttr, viewAttr),
            type: 'customEdge',
            data: { deletable: true },
          },
          updatedEdges
        );
      });
    },
    [viewFieldByName, dataPropertyByName, setEdges, message]
  );

  const handleChooseOk = (e: OntologyObjectType.DataSource[]) => {
    const dataView = e?.[0] || {};
    setDataViewInfo(dataView);
    // 手动选择数据源时,自动初始化属性和连线
    getDataViewDetail(dataView.id, true);
  };

  const onNodesChange: OnNodesChange<TFlowNode> = (changes) =>
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

  const connectedViewFieldNames = useMemo(() => {
    const set = new Set<string>();
    for (const edge of edges) {
      const name = edge.targetHandle?.replace('view-', '');
      if (name) set.add(name);
    }
    return set;
  }, [edges]);

  const handlePickAttributeOk = (targetKeys: string[]) => {
    setPickAttributeVisible(false);
    const targetKeySet = new Set(targetKeys);
    const existingNameSet = new Set(localDataProperties.map((p) => p.name));
    const newAttributes = fields
      .filter((f) => targetKeySet.has(f.name) && !existingNameSet.has(f.name))
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
      id: makeEdgeId(attr.name, attr.name),
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
    return fields.filter((f) => !existingPropertyNames.has(f.name) && !existingDisplayNames.has(f.display_name) && !connectedViewFieldNames.has(f.name));
  }, [fields, localDataProperties, connectedViewFieldNames]);

  const selectedRowKeys = useMemo(() => (dataViewInfo?.id ? [dataViewInfo.id] : []), [dataViewInfo]);

  const fieldSelectOptions: TFieldSelectOption[] = useMemo(
    () =>
      localDataProperties.map((p) => ({
        name: p.name,
        display_name: p.display_name || p.name,
        type: p.type || '',
        comment: p.comment,
      })),
    [localDataProperties]
  );

  const keySelection = useMemo(() => {
    const primaryNames: string[] = [];
    let primaryCount = 0;
    let primarySingleDisplayName: string | undefined;
    let displayKeyName: string | undefined;
    let displayKeyDisplayName: string | undefined;
    let incrementalKeyName: string | undefined;
    let incrementalKeyDisplayName: string | undefined;

    for (const p of localDataProperties) {
      if (p.primary_key) {
        primaryNames.push(p.name);
        primaryCount += 1;
        if (primaryCount === 1) primarySingleDisplayName = p.display_name || p.name;
      }
      if (p.display_key) {
        displayKeyName = p.name;
        displayKeyDisplayName = p.display_name || p.name;
      }
      if (p.incremental_key) {
        incrementalKeyName = p.name;
        incrementalKeyDisplayName = p.display_name || p.name;
      }
    }

    return {
      primaryNames,
      primaryCount,
      primarySingleDisplayName,
      displayKeyName,
      displayKeyDisplayName,
      incrementalKeyName,
      incrementalKeyDisplayName,
    };
  }, [localDataProperties]);

  const orderedPrimaryKeyFields = useMemo(
    () => orderFieldsWithSelectedFirst(fieldSelectOptions, keySelection.primaryNames),
    [fieldSelectOptions, keySelection.primaryNames]
  );
  const orderedDisplayKeyFields = useMemo(
    () => orderFieldsWithSelectedFirst(fieldSelectOptions, keySelection.displayKeyName ? [keySelection.displayKeyName] : []),
    [fieldSelectOptions, keySelection.displayKeyName]
  );
  const orderedIncrementalKeyFields = useMemo(
    () => orderFieldsWithSelectedFirst(fieldSelectOptions, keySelection.incrementalKeyName ? [keySelection.incrementalKeyName] : []),
    [fieldSelectOptions, keySelection.incrementalKeyName]
  );

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
            <Popover
              trigger="click"
              open={primaryKeyDropdownOpen}
              onOpenChange={setPrimaryKeyDropdownOpen}
              placement="bottomRight"
              content={
                <FieldSelect
                  style={{ minWidth: 260 }}
                  mode="multiple"
                  maxTagCount={1}
                  allowClear
                  value={keySelection.primaryNames}
                  placeholder={intl.get('Global.notConfigured')}
                  fields={orderedPrimaryKeyFields}
                  getOptionDisabled={(field) => !canBePrimaryKey(field?.type)}
                  onChange={(value) => {
                    handleSelectPrimaryKeys(value as string[] | undefined);
                  }}
                />
              }
            >
              {keySelection.primaryCount > 1 ? (
                <span className={styles['count-badge']}>{keySelection.primaryCount}</span>
              ) : (
                <span className={keySelection.primaryCount === 1 ? styles['info-value'] : styles['info-value-empty']}>
                  {keySelection.primarySingleDisplayName || intl.get('Global.notConfigured')}
                </span>
              )}
            </Popover>
          </div>

          <div className={styles['divider']} />
          <div className={styles['info-item']}>
            <IconFont type="icon-dip-color-star" />
            <span className={styles['info-label']}>{intl.get('Global.title')}</span>
            <Tooltip title={intl.get('Object.displayKeyTip')}>
              <IconFont type="icon-dip-color-tip" className={styles.helpIcon} />
            </Tooltip>
            <span className={styles['info-label']}>:</span>
            <Popover
              trigger="click"
              open={displayKeyDropdownOpen}
              onOpenChange={setDisplayKeyDropdownOpen}
              placement="bottomRight"
              content={
                <FieldSelect
                  style={{ minWidth: 220 }}
                  value={keySelection.displayKeyName}
                  placeholder={intl.get('Global.notConfigured')}
                  allowClear
                  fields={orderedDisplayKeyFields}
                  getOptionDisabled={(field) => !canBeDisplayKey(field?.type)}
                  onChange={(value) => {
                    handleSelectDisplayKey(value as string | undefined);
                    setDisplayKeyDropdownOpen(false);
                  }}
                />
              }
            >
              <span className={keySelection.displayKeyName ? styles['info-value'] : styles['info-value-empty']}>
                {keySelection.displayKeyDisplayName || intl.get('Global.notConfigured')}
              </span>
            </Popover>
          </div>
          <div className={styles['divider']} />
          <div className={styles['info-item']}>
            <IconFont type="icon-dip-color-increment" />
            <span className={styles['info-label']}>{intl.get('Object.incremental')}</span>
            <Tooltip title={intl.get('Object.incrementalKeyTip')}>
              <IconFont type="icon-dip-color-tip" className={styles.helpIcon} />
            </Tooltip>
            <span className={styles['info-label']}>:</span>
            <Popover
              trigger="click"
              open={incrementalKeyDropdownOpen}
              onOpenChange={setIncrementalKeyDropdownOpen}
              placement="bottomRight"
              content={
                <FieldSelect
                  style={{ minWidth: 220 }}
                  value={keySelection.incrementalKeyName}
                  placeholder={intl.get('Global.notConfigured')}
                  allowClear
                  fields={orderedIncrementalKeyFields}
                  getOptionDisabled={(field) => !canBeIncrementalKey(field?.type)}
                  onChange={(value) => {
                    handleSelectIncrementalKey(value as string | undefined);
                    setIncrementalKeyDropdownOpen(false);
                  }}
                />
              }
            >
              <span className={keySelection.incrementalKeyName ? styles['info-value'] : styles['info-value-empty']}>
                {keySelection.incrementalKeyDisplayName || intl.get('Global.notConfigured')}
              </span>
            </Popover>
          </div>
        </div>
      </div>
      <div style={{ flex: 1, position: 'relative' }}>
        <HoveredEdgeIdContext.Provider value={hoveredEdge}>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={handleEdgesChange}
            onConnect={onConnect}
            onEdgeMouseEnter={handleEdgeMouseEnter}
            onEdgeMouseLeave={handleEdgeMouseLeave}
            nodeTypes={nodeTypes}
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
        </HoveredEdgeIdContext.Provider>
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
