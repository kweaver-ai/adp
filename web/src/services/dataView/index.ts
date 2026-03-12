import { formatKeyOfObjectToLine } from '@/utils/format-objectkey-structure';
import UTILS from '@/utils';
import Request from '../request';
import * as DataViewType from './type';

// API URL 常量
const MDL_BASE_URL = '/api/mdl-data-model/v1';
const DATA_CONNECTION_BASE_URL = '/api/data-connection/v1';
const UNIQUERY_BASE_URL = '/api/mdl-uniquery/v1';

const API_URLS = {
  DATA_VIEWS: `${MDL_BASE_URL}/data-views`,
  DATA_SOURCES: `${MDL_BASE_URL}/data-sources`,
  DATA_VIEW_GROUPS: `${MDL_BASE_URL}/data-view-groups`,
  DATASOURCE_CONNECTION: `${DATA_CONNECTION_BASE_URL}/datasource`,
  EXCEL_FILES: `${DATA_CONNECTION_BASE_URL}/gateway/excel/files`,
  PREVIEW: `${UNIQUERY_BASE_URL}/data-views`,
};

/**
 * 获取数据源列表
 */
const getDataSourceList = async (params: DataViewType.GetDataSourceListParams = {}): Promise<DataViewType.List<DataViewType.DataSource>> => {
  return await Request.get(API_URLS.DATA_SOURCES, params);
};

/**
 * 获取数据视图详情
 * @param id 视图ID
 */
const getDataViewDetail = async (id: string): Promise<any> => {
  return await Request.get(`${API_URLS.DATA_VIEWS}/${id}`);
};

/**
 * 获取数据连接的数据源信息
 * @param params 查询参数
 */
const getDatasourceConnection = async (params: Record<string, any>): Promise<any> => {
  return await Request.get(API_URLS.DATASOURCE_CONNECTION, params);
};

/**
 * 查询 Excel 数据源文件
 * @param fileName 文件名
 */
const getExcelFiles = async (fileName: string): Promise<{ data: string[] }> => {
  return await Request.get(`${API_URLS.EXCEL_FILES}/${fileName}`, {}, { timeout: 60000 });
};

/**
 * 获取原子视图列表
 */
const getAtomViewList = async (params: DataViewType.GetAtomViewListParams): Promise<DataViewType.List<DataViewType.DataView>> => {
  const {
    excelFileName,
    dataSourceType,
    dataSourceId,
    offset = 0,
    limit = 10,
    sort = 'update_time',
    direction = 'desc',
    name = '',
    tag = '',
    queryType,
  } = params;

  const queryParams = UTILS.filterEmptyFields({
    keyword: name,
    sort,
    direction,
    offset,
    limit,
    tag,
    file_name: excelFileName,
    data_source_type: dataSourceType,
    data_source_id: dataSourceId,
    query_type: queryType,
  });

  return await Request.get(API_URLS.DATA_VIEWS, queryParams);
};

/**
 * 数据预览
 * @param ids 视图ID或ID列表
 * @param values 预览参数
 */
const getViewDataPreview = async (ids: string | string[], values: Record<string, any>): Promise<any> => {
  const idStr = Array.isArray(ids) ? ids.join(',') : ids;
  return await Request.postOverrideGet(`${API_URLS.PREVIEW}/${idStr}?include_view=true`, formatKeyOfObjectToLine(values));
};

/**
 * 获取数据视图列表
 */
const getDataViewList = async (params: DataViewType.GetDataViewListParams): Promise<DataViewType.List<DataViewType.DataView>> => {
  const queryParams = UTILS.filterEmptyFields({
    ...params,
    sort: params.sort || 'update_time',
    direction: params.direction || 'desc',
  });
  return await Request.get(API_URLS.DATA_VIEWS, queryParams);
};

/**
 * 获取数据视图分组
 */
const getDataViewGroup = async (): Promise<DataViewType.List<DataViewType.Group>> => {
  const params = { limit: -1 };
  return await Request.get(API_URLS.DATA_VIEW_GROUPS, params);
};

export default {
  getDataSourceList,
  getDataViewDetail,
  getDatasourceConnection,
  getExcelFiles,
  getAtomViewList,
  getViewDataPreview,
  getDataViewList,
  getDataViewGroup,
};
