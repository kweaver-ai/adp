// 数据源连接信息
export interface BinData {
  catalog_name?: string; // 数据源catalog名称
  data_view_source?: string; // 数据视图源
  database_name?: string; // 数据库名称
  connect_protocol: string; // 连接方式
  schema?: string; // 数据库模式
  host: string; // 地址
  port: number; // 端口
  account?: string; // 账户/用户名
  password?: string; // 密码
  token?: string; // token认证
  storage_protocol?: string; // 存储介质
  storage_base?: string; // 存储路径
  replica_set?: string; // 副本集名称
}

// 数据源详情
export interface DataSource {
  id: string;
  name: string; // 数据源名称
  type: string; // 数据源类型
  bin_data: BinData; // 数据源连接信息
  comment?: string; // 描述
  created_by_uid: string; // 创建人id
  created_by_username: string; // 创建人名称
  created_at: number; // 创建时间
  updated_at: number; // 更新时间
  updated_by_uid: string; // 修改人id
  updated_by_username: string; // 修改人名称
  allow_multi_table_scan: boolean; // 是否支持多表扫描
  is_built_in: boolean; // 是否内置数据源
  auth_method?: number; // 认证方式
  deploy_method?: number; // 部署方式
  operations?: string[]; // 操作权限
  title: string; // 数据源分类名称
  icon: any; // 数据源分类图标
  key: string;
  paramType: string;
  isLeaf: boolean;
  deployMethod?: number; // 前端转换字段：部署方式
  authMethod?: number; // 前端转换字段：认证方式
  children?: DataSource[];
  metadata_obtain_level: number; // 元数据获取级别
  selectable?: boolean; // 是否可选择
}

// 列表查询参数
export interface GetDataSourceListParams {
  offset?: number;
  limit?: number;
  sort?: string;
  direction?: string;
  keyword?: string;
  type?: string;
}

// 列表响应
export interface List<T> {
  entries: T[];
  total_count: number;
}

// 连接器信息
export interface Connector {
  connect_protocol: string; // 连接方式
  olk_connector_name: string; // 原始数据源类型名称
  show_connector_name: string; // 显示数据源类型名称
  type: string; // 数据源分类
}

// 连接器响应
export interface ConnectorsResponse {
  connectors: Connector[];
}

// 创建/测试连接响应
export interface CreateDataSourceResponse {
  id: number;
}
