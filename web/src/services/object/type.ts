import { ValueFrom } from '../action/type';

export interface ConceptGroup {
  id: string;
  name?: string;
}

export interface DataSource {
  type: 'data_view';
  id: string;
  name?: string;
}

export interface ViewField {
  name: string;
  display_name?: string;
  type?: string;
  comment?: string;
}

export type TokenizerType = 'standard' | 'english' | 'ik_max_word' | 'hanlp_standard' | 'hanlp_index';

export interface FulltextConfig {
  analyzer: TokenizerType;
  field_keyword: boolean;
}

export interface VectorConfig {
  dimension: number;
}

export interface IndexConfig {
  keyword_config: {
    enabled: boolean;
    ignore_above_len: number;
  };
  fulltext_config: {
    enabled: boolean;
    analyzer: string;
  };
  vector_config: {
    enabled: boolean;
    model_id: string;
  };
}

export interface DataProperty {
  name: string;
  display_name: string;
  original_name?: string;
  type?: string;
  comment?: string;
  mapped_field?: ViewField;
  index?: boolean;
  fulltext_config?: FulltextConfig;
  vector_config?: VectorConfig;
  index_config?: IndexConfig;
  // 对象配置信息
  primary_key?: boolean;
  display_key?: boolean;
  incremental_key?: boolean;
}

export enum LogicAttributeType {
  METRIC = 'metric',
  OPERATOR = 'operator',
}

export interface LogicSource {
  type: LogicAttributeType;
  id: string;
  name?: string;
}

export interface Parameter {
  name: string;
  value_from: ValueFrom;
  value?: string;
  operation?: string;
  type?: string;
  source?: string;
  description?: string;
  if_system_generate?: boolean;
  id: string;
}

export interface LogicProperty {
  name: string;
  display_name?: string;
  type?: string;
  comment?: string;
  index?: boolean;
  data_source: LogicSource | null;
  parameters: Parameter[] | null;
}

export interface BasicInfo {
  name: string;
  tags?: string[];
  comment?: string;
  concept_groups?: ConceptGroup[];
  concept_groupIds?: string[];
  id: string;
  icon: string;
  color: string;
}

export interface ReqObjectType extends BasicInfo {
  branch: string;
  base_branch?: string;
  data_source?: DataSource;
  data_properties: DataProperty[];
  logic_properties?: LogicProperty[];
  primary_keys: string[];
  display_key: string;
  incremental_key: string;
}

export type UpdateObjectType = ReqObjectType;

export interface Detail extends ReqObjectType {
  concept_type: 'object_type';
  id: string;
  kn_id: string;
  creator: {
    id: string;
    name: string;
    type: string;
  };
  create_time: number;
  updater?: {
    id: string;
    name: string;
    type: string;
  };
  update_time: number;
  detail?: string;
  status?: {
    index_available: boolean;
    doc_count: number;
    incremental_key: string;
    incremental_value: string;
    index: string;
    storage_size: number;
    update_time: number;
  };
}

export interface ListObjectTypes {
  entries: Detail[];
  total_count: number;
}

export interface ListQuery {
  name_pattern?: string;
  sort?: 'update_time' | 'name';
  direction?: 'asc' | 'desc';
  offset?: number;
  limit?: number;
  tag?: string;
  group_id?: string;
}

export interface PathKnId {
  kn_id: string;
}

export interface PathObId {
  ob_id: string;
}

export interface PathObIds {
  ob_ids: string[];
}

export interface DetailQuery {
  include_detail?: boolean;
}

export interface Field {
  name: string;
  type?: string;
  display_name: string;
  comment?: string;
  primary_key?: boolean;
  display_key?: boolean;
  incremental_key?: boolean;
  id?: string;
  displayNameAdd?: boolean;
  error: Record<string, string>;
}

export interface TNode {
  id: string;
  type: string;
  position: { x: number; y: number };
  selected?: boolean;
  data: {
    toggleIncrementalKey?: (attrName: string) => unknown;
    toggleDisplayKey?: (attrName: string) => unknown;
    togglePrimaryKey?: (attrName: string) => unknown;
    deleteAttribute?: (name: string) => unknown;
    clearAllAttributes?: () => void;
    clearSearchTrigger?: number;
    label: string;
    bg: string;
    icon: string;
    attrClick?: (val: any) => void;
    openDataViewSource?: () => void;
    deleteDataViewSource?: () => void;
    addDataAttribute?: () => void;
    pickAttribute?: () => void;
    autoLine?: () => void;
    highlightedAttributes?: string[];
    attributes: {
      name: string;
      display_name: string;
      type: string;
      comment?: string;
      primary_key?: boolean;
      display_key?: boolean;
      incremental_key?: boolean;
      mapped_field?: {
        name: string;
        display_name?: string;
        type?: string;
      };
    }[];
  };
}

export interface TEdge {
  id: string;
  type: string;
  source: string;
  sourceHandle: string;
  target: string;
  targetHandle: string;
}

export interface MetricModelItem {
  id: string;
  name: string;
  group_name: string;
  analysis_dimensions: {
    name: string;
    display_name: string;
    type: string;
  }[];
}

export interface MetricModelList {
  total_count: number;
  entries: MetricModelItem[];
}

export interface OperatorItem {
  name: string;
  operator_id: string;
  metadata: {
    api_spec: any;
  };
  execution_mode: 'sync' | 'async';
}

export interface OperatorList {
  data: OperatorItem[];
}

export interface MetricModelField {
  name: string;
  display_name: string;
  type: string;
  comment?: string;
}

export interface SmallModelItem {
  model_id: string;
  model_name: string;
  batch_size: number;
  embedding_dim: number;
  max_tokens: number;
}

export interface SmallModelList {
  count: number;
  data: SmallModelItem[];
}
