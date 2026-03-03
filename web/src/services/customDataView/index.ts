import { formatKeyOfObjectToLine } from '@/utils/format-objectkey-structure';
import Request from '../request';
import * as CustomDataViewType from './type';
import { getObjectTags } from '../tag';

// API URL 常量
const API_BASE_URL = '/api/mdl-data-model/v1';
const GROUP_API_URL = `${API_BASE_URL}/data-view-groups`;
const DATA_VIEW_API_URL = `${API_BASE_URL}/data-views`;
const DATA_VIEW_PREVIEW_API_URL = '/api/mdl-uniquery/v1/data-views';

/**
 * 处理查询参数
 * 1. 过滤 null/undefined
 * 2. 数组转逗号分隔字符串
 * @param params 参数对象
 */
const processQueryParams = (params: Record<string, any>): Record<string, any> => {
  const result: Record<string, any> = {};

  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      if (Array.isArray(value)) {
        result[key] = value.map((item) => String(item)).join(',');
      } else {
        result[key] = value;
      }
    }
  });

  return result;
};

/**
 * 获取分组列表
 * @returns 分组列表数据
 */
const getGroupList = async (): Promise<CustomDataViewType.GetGroupListResponse> => {
  const params = {
    limit: -1, // 不分页
    builtin: false, // 不返回系统内置的指标模型
  };

  return await Request.get(GROUP_API_URL, params);
};

/**
 * 创建新分组
 * @param name 分组名称
 * @returns 创建结果
 */
const createGroup = async (name: string) => {
  return await Request.post(GROUP_API_URL, { name });
};

/**
 * 更新分组名称
 * @param id 分组ID
 * @param name 新名称
 * @returns 更新结果
 */
const updateGroup = async (id: string, name: string) => {
  return await Request.put(`${GROUP_API_URL}/${id}`, { name });
};

/**
 * 删除分组
 * @param id 分组ID
 * @param force 是否强制删除（同时删除分组下的视图）
 * @returns 删除结果
 */
const deleteGroup = async (id: string, force: boolean): Promise<any> => {
  return await Request.delete(`${GROUP_API_URL}/${id}?delete_views=${force}`);
};

/**
 * 导出分组数据
 * @param id 分组ID
 * @returns 导出的数据
 */
const exportGroup = async (id: string): Promise<any> => {
  return await Request.get(`${GROUP_API_URL}/${id}/data-views`);
};

/**
 * 获取自定义数据视图列表
 * @param params 查询参数
 * @returns 数据视图列表
 */
const getCustomDataViewList = async (params: CustomDataViewType.DataViewListParams): Promise<CustomDataViewType.GetCustomDataViewListResponse> => {
  const {
    limit = -1,
    offset = 0,
    sort = 'update_time',
    direction = 'desc',
    keyword = '',
    query_type = [],
    tag = '',
    group_id,
    simple_info = false,
    type = 'custom',
  } = params;

  // 处理groupId特殊值
  const group_id_val = group_id === undefined || group_id === null ? '__all' : group_id || '';

  const queryParams = processQueryParams({
    keyword,
    sort,
    query_type,
    direction,
    offset,
    limit,
    tag,
    group_id: group_id_val,
    simple_info,
    type,
  });

  return await Request.get(DATA_VIEW_API_URL, queryParams);
};

/**
 * 创建自定义数据视图
 * @param data 视图数据
 * @param importMode 导入模式
 * @returns 创建结果
 */
const createCustomDataView = async (data: Record<string, any>, importMode?: string): Promise<any> => {
  const params = importMode ? { import_mode: importMode } : undefined;
  return await Request.post(DATA_VIEW_API_URL, formatKeyOfObjectToLine(data), { params });
};

/**
 * 更新自定义数据视图
 * @param id 视图ID
 * @param data 更新数据
 * @returns 更新结果
 */
const updateCustomDataView = async (id: string, data: Record<string, any>) => {
  return await Request.put(`${DATA_VIEW_API_URL}/${id}`, formatKeyOfObjectToLine(data));
};

/**
 * 删除自定义数据视图
 * @param id 视图ID
 * @returns 删除结果
 */
const deleteCustomDataView = async (id: string): Promise<any> => {
  return await Request.delete(`${DATA_VIEW_API_URL}/${id}`);
};

/**
 * 获取自定义数据视图详情
 * @param ids 视图ID数组
 * @param includeDataScopeViews 是否包含数据作用域视图
 * @returns 视图详情
 */
const getCustomDataViewDetails = async (ids: string[], includeDataScopeViews = false): Promise<CustomDataViewType.CustomDataView[]> => {
  const params = { include_data_scope_views: includeDataScopeViews };
  return await Request.get(`${DATA_VIEW_API_URL}/${ids.join(',')}`, params);
};

/**
 * 批量修改数据视图分组
 * @param ids 视图ID数组
 * @param groupName 新分组名称
 * @returns 修改结果
 */
const changeCustomDataViewGroup = async (ids: string[], groupName: string) => {
  return await Request.put(`${DATA_VIEW_API_URL}/${ids.join(',')}/attrs/group_name`, { group_name: groupName });
};

/**
 * 获取原子视图列表
 * @param params 查询参数
 * @returns 原子视图列表
 */
const getAtomViewList = async (params: CustomDataViewType.AtomViewListParams): Promise<any> => {
  const {
    type = 'atomic',
    excelFileName = '',
    dataSourceType,
    dataSourceId,
    offset = 0,
    limit = -1,
    sort = 'update_time',
    direction = 'desc',
    name = '',
    tag = '',
    queryType = '',
  } = params;

  const queryParams = processQueryParams({
    type,
    keyword: name,
    direction,
    offset,
    limit,
    sort,
    tag,
    file_name: excelFileName,
    data_source_type: dataSourceType,
    data_source_id: dataSourceId,
    query_type: queryType,
  });

  return await Request.get(DATA_VIEW_API_URL, queryParams);
};

/**
 * 获取视图数据预览
 * @param ids 视图ID（单个或多个）
 * @param values 查询参数
 * @returns 预览数据
 */
const getViewDataPreview = async (ids: string | string[], values: Record<string, any>): Promise<any> => {
  const idStr = Array.isArray(ids) ? ids.join(',') : ids;
  const params = { include_view: true };
  const url = `${DATA_VIEW_PREVIEW_API_URL}/${idStr}`;

  // 使用 Request.post 并通过 config 传入 headers 和 params
  return await Request.postOverrideGet(url, formatKeyOfObjectToLine(values), {
    params,
  });
};

/**
 * 获取节点数据预览
 * @param values 查询参数
 * @returns 预览数据
 */
const getNodeDataPreview = async (values: Record<string, any>): Promise<any> => {
  const params = { include_view: false };
  return await Request.postOverrideGet(DATA_VIEW_PREVIEW_API_URL, formatKeyOfObjectToLine(values), {
    params,
  });
};

/**
 * 获取标签列表
 * @returns 标签列表数据
 */
const getTagList = async (): Promise<any> => {
  return await getObjectTags({ module: 'metric-model' });
};

/**
 * 自定义数据视图服务模块
 * 提供数据视图分组、自定义数据视图、原子视图、数据预览和标签相关的API接口
 */
export default {
  // 数据视图分组相关
  getGroupList,
  createGroup,
  updateGroup,
  deleteGroup,
  exportGroup,

  // 自定义数据视图相关
  getCustomDataViewList,
  createCustomDataView,
  updateCustomDataView,
  deleteCustomDataView,
  getCustomDataViewDetails,
  changeCustomDataViewGroup,

  // 原子视图相关
  getAtomViewList,

  // 数据预览相关
  getViewDataPreview,
  getNodeDataPreview,

  // 标签相关
  getTagList,
};
