import * as OntologyObjectType from '../object/type';

export interface BasicInfo {
  id: string;
  name: string;
  tags?: string[];
  comment?: string;
  icon?: string;
  color?: string;
  detail?: string;
  kn_id: string;
  branch: string;
  creator?: { id: string; name: string; type: string };
  create_time?: number;
  updater?: { id: string; name: string; type: string };
  update_time?: number;
  validate_dependency?: boolean;
}

export type CreateRequest = BasicInfo;

export interface UpdateRequest {
  name?: string;
  tags?: string[];
  comment?: string;
  icon?: string;
  color?: string;
  detail?: string;
}

export interface Detail extends BasicInfo {
  statistics?: {
    object_types_total?: number;
    relation_types_total?: number;
    action_types_total?: number;
  };
  object_types?: Array<OntologyObjectType.Detail>;
  relation_types?: Array<{ id: string; name: string; tags?: string[] }>;
  action_types?: Array<{ id: string; name: string; tags?: string[] }>;
}

export interface ListQuery {
  name_pattern?: string;
  tag?: string;
  sort?: 'update_time' | 'name';
  direction?: 'asc' | 'desc';
  offset?: number;
  limit?: number;
}

export interface List {
  entries: Detail[];
  total_count: number;
}

export interface AddObjectTypesRequest {
  entries: Array<{ id: string }>;
}
