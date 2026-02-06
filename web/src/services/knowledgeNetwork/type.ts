// 业务知识网络详情
export interface KnowledgeNetwork {
  id: string; // 业务知识网络ID
  name: string; // 业务知识网络名称
  icon: string;
  color: string;
  tags: string[]; // 标签
  comment?: string; // 备注
  creator: {
    id: string;
    name: string;
    type: string;
  }; // 创建者
  create_time: number; // 创建时间
  updater?: {
    id: string;
    name: string;
    type: string;
  }; // 更新者
  update_time?: number; // 更新时间
  detail?: string; // 说明书的markdown文本内容
  operations: string[];
  statistics?: {
    object_types_total: number;
    relation_types_total: number;
    action_types_total: number;
    concept_groups_total: number;
  };
}

// 列表查询参数
export interface GetNetworkListParams {
  name_pattern?: string; // 名称模糊查询
  tag?: string;
  sort?: 'update_time' | 'kn_name'; // 排序类型
  direction?: 'asc' | 'desc'; // 排序方向
  offset?: number;
  limit?: number;
}

// 创建请求参数
export interface CreateNetworkRequest {
  id: string; // 业务知识网络ID
  name: string; // 业务知识网络名称
  tags: string[]; // 标签
  comment?: string; // 备注
  branch: string; // 分支ID
  base_version: string; // 来源版本
  color?: string; // 颜色
  icon?: string; // 图标
  validate_dependency?: boolean;
}

// 更新请求参数
export interface UpdateNetworkRequest {
  name?: string; // 业务知识网络名称
  tags?: string[]; // 标签
  comment?: string; // 备注
  color?: string; // 颜色
  icon?: string; // 图标
}

// 详情查询参数
export interface GetNetworkDetailParams {
  knId: string;
  mode?: 'export';
  include_detail?: boolean;
  include_statistics?: boolean;
}

// 列表响应
export interface List<T> {
  entries: T[];
  total_count: number;
}
