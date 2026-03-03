export type GroupType = {
  id: string;
  name: string;
  data_view_count: number;
  update_time: string;
  builtin: boolean;
};

export interface PaginationParams {
  limit?: number;
  offset?: number;
  sort?: string;
  direction?: 'asc' | 'desc';
}

export interface DataViewListParams extends PaginationParams {
  keyword?: string;
  query_type?: string[];
  tag?: string;
  group_id?: string | null;
  simple_info?: boolean;
  type: string;
}

export interface AtomViewListParams extends PaginationParams {
  excelFileName?: string;
  dataSourceType?: string;
  dataSourceId?: string;
  name?: string;
  tag?: string;
  type?: string;
  queryType?: string;
}

export interface GetGroupListResponse {
  entries: GroupType[];
  total_count: number;
}

export interface CustomDataView {
  id: string;
  name: string;
  type: string;
  group_id: string;
  group_name: string;
  query_type: string;
  tags: string[];
  comment: string;
  create_time: number;
  creator: {
    id: string;
    name: string;
    type: string;
  };
  update_time: number;
  updater: {
    id: string;
    name: string;
    type: string;
  };
  kn_id: string;
  builtin: boolean;
  data_source_id: string;
  data_source_type: string;
  sql: string;
  schema: any;
  fields: any[];
  data_scope: any;
  primary_keys: string[];
}

export interface GetCustomDataViewListResponse {
  entries: CustomDataView[];
  total_count: number;
}

export interface GetTagListResponse {
  entries: Array<{
    tag: string;
    module: string;
  }>;
  total_count: number;
}
