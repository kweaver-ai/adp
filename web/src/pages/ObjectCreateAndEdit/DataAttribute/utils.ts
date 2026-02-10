import intl from 'react-intl-universal';
import * as OntologyObjectType from '@/services/object/type';

export const makeEdgeId = (dataAttr: string, viewAttr: string) => `${dataAttr}&&${viewAttr}`;

export const parseEdgeId = (id?: string) => {
  const [dataAttr = '', viewAttr = ''] = (id || '').split('&&');
  return { dataAttr, viewAttr };
};

export interface TransformCanvasDataParams {
  dataProperties: OntologyObjectType.DataProperty[];
  logicProperties: OntologyObjectType.LogicProperty[];
  fields: OntologyObjectType.Field[];
  dataSource?: OntologyObjectType.DataSource;
  basicValue: OntologyObjectType.BasicInfo;
  /**
   * 是否需要计算初始连线。初始化阶段需要；后续交互阶段 edges 由 ReactFlow state 维护，
   * 每次重建 edges 会造成不必要的计算。
   */
  includeEdges?: boolean;
  openDataViewSource?: () => void;
  deleteDataViewSource?: () => void;
  addDataAttribute?: () => void;
  attrClick?: (val: OntologyObjectType.Field) => void;
  pickAttribute?: () => void;
  autoLine?: () => void;
  deleteAttribute?: (attrName: string) => void;
  togglePrimaryKey?: (attrName: string) => void;
  toggleDisplayKey?: (attrName: string) => void;
  toggleIncrementalKey?: (attrName: string) => void;
  clearAllAttributes?: () => void;
  clearSearchTrigger?: number;
}

export const transformCanvasData = (
  props: TransformCanvasDataParams
): {
  nodes: OntologyObjectType.TNode[];
  edges: OntologyObjectType.TEdge[];
} => {
  const {
    dataProperties,
    logicProperties,
    fields,
    dataSource,
    basicValue,
    includeEdges = true,
    openDataViewSource,
    deleteDataViewSource,
    addDataAttribute,
    attrClick,
    pickAttribute,
    autoLine,
    deleteAttribute,
    togglePrimaryKey,
    toggleDisplayKey,
    toggleIncrementalKey,
    clearAllAttributes,
    clearSearchTrigger,
  } = props;
  const logics = logicProperties.map((item) => item.name) || [];

  const dataNodes = dataProperties
    .filter((val) => !logics.includes(val.name))
    .map((item) => ({
      ...item,
      id: item.name,
      mapped_field: item.mapped_field,
      primary_key: item.primary_key,
      display_key: item.display_key,
      incremental_key: item.incremental_key,
    }));

  const nodesView: OntologyObjectType.TNode = {
    id: 'view',
    type: 'customNode',
    position: { x: 150, y: 30 },
    data: {
      label: dataSource?.name || intl.get('Global.dataView'),
      bg: '#fff',
      icon: 'icon-dip-usedata',
      openDataViewSource,
      deleteDataViewSource,
      attrClick,
      pickAttribute,
      autoLine,
      deleteAttribute,
      togglePrimaryKey,
      toggleDisplayKey,
      toggleIncrementalKey,
      clearAllAttributes,
      clearSearchTrigger,
      attributes: fields.map((item) => ({
        id: item.name,
        name: item.name,
        display_name: item.display_name,
        type: item.type || '',
        comment: item.comment,
        primary_key: item.primary_key,
        display_key: item.display_key,
        incremental_key: item.incremental_key,
      })),
    },
  };

  const nodesData: any = {
    id: 'data',
    type: 'customNode',
    position: { x: 760, y: 30 },
    data: {
      label: basicValue.name,
      bg: basicValue.color,
      icon: basicValue.icon,
      addDataAttribute,
      attrClick,
      pickAttribute,
      deleteAttribute,
      togglePrimaryKey,
      toggleDisplayKey,
      toggleIncrementalKey,
      clearAllAttributes,
      clearSearchTrigger,
      attributes: dataNodes,
    },
  };

  const edgesData: OntologyObjectType.TEdge[] = [];
  if (includeEdges) {
    const edgesFromMappedField = dataProperties
      .filter((val) => val.mapped_field?.name && val.type == val.mapped_field?.type)
      .map((item) => ({
        id: makeEdgeId(item.name, item.mapped_field?.name || ''),
        type: 'customEdge',
        source: 'data',
        sourceHandle: 'data-' + item.name,
        target: 'view',
        targetHandle: 'view-' + item.mapped_field?.name,
        data: { deletable: true },
      }));

    edgesData.push(...edgesFromMappedField);

    if (!(edgesData.length > 0)) {
      const fieldByName = new Map(fields.map((f) => [f.name, f] as const));
      dataProperties.forEach((val) => {
        const beField = fieldByName.get(val.name);
        if (beField && (beField.type === val.type || !val.type)) {
          edgesData.push({
            id: makeEdgeId(val.name, beField.name),
            type: 'customEdge',
            source: 'data',
            sourceHandle: 'data-' + val.name,
            target: 'view',
            targetHandle: 'view-' + beField.name,
          });
        }
      });
    }
  }
  return {
    nodes: [nodesData, nodesView],
    edges: edgesData,
  };
};
interface TransformAttrDataParams {
  nodes: {
    id: string;
    type: string;
    position: { x: number; y: number };
    data: {
      label: string;
      attributes: {
        name: string;
        display_name: string;
        type: string;
        comment?: string;
        mapped_field?: {
          name: string;
          display_name?: string;
          type?: string;
        };
      }[];
    };
  }[];
  edges: {
    id: string;
    type: string;
    source: string;
    sourceHandle: string;
    target: string;
    targetHandle: string;
  }[];
  logics: string[];
}

export const transformAttrData = (props: TransformAttrDataParams): OntologyObjectType.DataProperty[] => {
  const { nodes, edges, logics } = props;
  const dataNode = nodes.find((val) => val.id === 'data')!;
  const viewNode = nodes.find((val) => val.id === 'view')!;
  const realDataAttr = dataNode.data.attributes.filter((val) => !logics.includes(val.name));
  edges.forEach((val) => {
    const { dataAttr: source, viewAttr: target } = parseEdgeId(val.id);
    const sourceAttr = realDataAttr.find((item) => item.name === source);
    const targetAttr = viewNode.data.attributes.find((item) => item.name === target);
    sourceAttr!.mapped_field = {
      name: target,
      display_name: targetAttr?.display_name,
      type: targetAttr?.type,
    };
  });

  return realDataAttr;
};
