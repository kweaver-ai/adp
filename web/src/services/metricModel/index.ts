import { pickBy, cloneDeep } from 'lodash-es';
import { formatKeyOfObjectToCamel, formatKeyOfObjectToLine, formatCamelToLine } from '@/utils/format-objectkey-structure';
import * as MetricModelType from './type';
import Request from '../request';
import { getObjectTags } from '../tag';

const BASE_URL = '/api/mdl-data-model/v1/metric-models';
const GROUP_BASE_URL = '/api/mdl-data-model/v1/metric-model-groups';
const PREVIEW_BASE_URL = '/api/mdl-uniquery/v1/metric-models';
const INDEX_BASE_URL = '/api/mdl-index-base/v1/index_bases';

export const getMetricModelList = async (params: MetricModelType.MetricModelListParams): Promise<any> => {
  const {
    limit = -1,
    offset = 0,
    sort = 'update_time',
    direction = 'desc',
    name_pattern = '',
    query_type = [],
    tag = '',
    group_id,
    metric_type = [],
    simple_info = false,
  } = params;

  const group_id_val = group_id === undefined || group_id === null ? '__all' : group_id || '';

  const queryParams = {
    name_pattern,
    sort: formatCamelToLine(sort),
    query_type: query_type.join(','),
    direction,
    offset,
    limit,
    tag,
    group_id: group_id_val,
    simple_info,
    metric_type: metric_type.join(','),
  };

  const res = await Request.get<MetricModelType.MetricModelList>(BASE_URL, queryParams);

  return formatKeyOfObjectToCamel(res);
};

export const getMetricModelExportGroup = async (groupId: string): Promise<any> => {
  return await Request.get<MetricModelType.MetricModelList>(`${GROUP_BASE_URL}/${groupId}/metric-models`);
};

export const getMetricModelById = async (id: string): Promise<any> => {
  const res = await Request.get<MetricModelType.MetricModelItem[]>(`${BASE_URL}/${id}`);
  const { formula, query_type: queryTypeValue, ...restValues } = res[0] || {};

  return res[0]
    ? {
        [`${queryTypeValue}Formula`]: queryTypeValue === MetricModelType.QueryType.Dsl ? JSON.parse(formula) : formula,
        queryType: queryTypeValue,
        formula,
        ...formatKeyOfObjectToCamel(restValues),
      }
    : {};
};

export const getMetricModelByIds = async (ids: string[]): Promise<any> => {
  return Request.get<MetricModelType.MetricModelItem[]>(`${BASE_URL}/${ids.join(',')}`);
};

export const createMetricModel = async (values: MetricModelType.CreateMetricModelRequest): Promise<{ id: string }[]> => {
  return await Request.post(BASE_URL, [formatKeyOfObjectToLine(cloneDeep(values))]);
};

export const updateMetricModel = async (values: MetricModelType.UpdateMetricModelRequest, id: string): Promise<any> => {
  return Request.put(`${BASE_URL}/${id}`, formatKeyOfObjectToLine(values));
};

export const batchCreateMetricModel = async (values: Record<string, any>, importMode?: string): Promise<any> => {
  const url = importMode ? `${BASE_URL}?import_mode=${importMode}` : BASE_URL;
  return Request.post(url, values, { isNoHint: true });
};

export const deleteMetricModel = async (id: string): Promise<any> => {
  return Request.delete(`${BASE_URL}/${id}`);
};

export const fetchMetricPreviewData = async (values: Record<string, any>, id?: string): Promise<any> => {
  const url = id ? `${PREVIEW_BASE_URL}/${id}?include_model=true` : PREVIEW_BASE_URL;
  return Request.postOverrideGet(url, formatKeyOfObjectToLine(values));
};

export const getAllTags = async (): Promise<MetricModelType.TagList> => {
  return getObjectTags({
    name_pattern: '',
    offset: 0,
  }) as unknown as Promise<MetricModelType.TagList>;
};

export const getMetricModelTags = async (): Promise<MetricModelType.TagList> => {
  return getObjectTags({
    name_pattern: '',
    offset: 0,
    module: 'metric-model',
  }) as unknown as Promise<MetricModelType.TagList>;
};

export const getFields = async (id: string): Promise<any> => {
  return Request.get(`manager/loggroup/${id}/allFields`);
};

export const getGroupList = async (): Promise<any> => {
  const params = {
    limit: -1,
  };

  return Request.get(GROUP_BASE_URL, params);
};

export const createGroup = async (name: string, comment = ''): Promise<any> => {
  const data: MetricModelType.CreateGroupRequest = {
    name,
    comment,
  };

  return Request.post(GROUP_BASE_URL, data);
};

export const updateGroup = async (groupId: string, name: string, comment = ''): Promise<any> => {
  const data: MetricModelType.UpdateGroupRequest = {
    name,
    comment,
  };

  return Request.put(`${GROUP_BASE_URL}/${groupId}`, data);
};

export const deleteGroup = async (groupId: string, force: boolean): Promise<any> => {
  const params = {
    force,
  };

  return Request.delete(`${GROUP_BASE_URL}/${groupId}`, params);
};

export const batchChangeMetricModelGroup = async (metricModelIds: string[], groupName: string): Promise<any> => {
  const ids = metricModelIds.join(',');
  const data = {
    group_name: groupName,
  };

  return Request.put(`${BASE_URL}/${ids}/attributes`, data);
};

export const getMetricOrderFields = async (modelIds: string[]): Promise<MetricModelType.MetricOrderField[][]> => {
  return Request.get(`${BASE_URL}/${modelIds.join(',')}/order_fields`);
};

export const getIndexBaseList = async (params: MetricModelType.GetIndexBaseListParams): Promise<any> => {
  const { limit = -1, offset = 0, direction = 'asc', sort = 'name', process_status = '', name_pattern = '' } = params;

  // 过滤空值，保持与 axiosGetEncode 行为一致
  const queryParams: Record<string, any> = {
    limit,
    offset,
    direction,
    sort,
    process_status,
    name_pattern,
  };

  const filteredParams = pickBy(queryParams, (value) => value !== '' && value !== null && value !== undefined);

  return await Request.get(INDEX_BASE_URL, filteredParams);
};

export default {
  getMetricModelList,
  getMetricModelExportGroup,
  fetchMetricPreviewData,
  getMetricModelById,
  getMetricModelByIds,
  createMetricModel,
  updateMetricModel,
  batchCreateMetricModel,
  deleteMetricModel,
  getAllTags,
  getMetricModelTags,
  getFields,
  getGroupList,
  createGroup,
  updateGroup,
  deleteGroup,
  batchChangeMetricModelGroup,
  getMetricOrderFields,
  getIndexBaseList,
};
