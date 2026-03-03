import Request from '../request';
import * as OntologyObjectType from './type';

const BASE_URL = '/api/ontology-manager/v1/knowledge-networks';
const OPERATOR_BASE_URL = '/api/agent-operator-integration/v1/operator/market';
const METRIC_MODEL_BASE_URL = '/api/mdl-data-model/v1/metric-models';
const METRIC_MODEL_PREVIEW_BASE_URL = '/api/mdl-uniquery/v1/metric-models';
const SMALL_MODEL_BASE_URL = '/api/mf-model-manager/v1/small-model/list';

export const objectGet = (knId: string, params?: OntologyObjectType.ListQuery): Promise<OntologyObjectType.ListObjectTypes> => {
  return Request.get(`${BASE_URL}/${knId}/object-types`, params);
};

export const createObjectTypes = (knId: string, payload: OntologyObjectType.ReqObjectType[]): Promise<{ id: string }[]> => {
  return Request.post(`${BASE_URL}/${knId}/object-types`, { entries: payload });
};

export const updateObjectType = (knId: string, obId: string, payload: OntologyObjectType.UpdateObjectType): Promise<void> => {
  return Request.put(`${BASE_URL}/${knId}/object-types/${obId}`, payload);
};

export const getDetail = (knId: string, obIds: string[], query?: OntologyObjectType.DetailQuery): Promise<OntologyObjectType.Detail[]> => {
  return Request.get<{ entries: OntologyObjectType.Detail[] }>(`${BASE_URL}/${knId}/object-types/${obIds.join(',')}`, query).then(
    (response) => response.entries
  );
};

export const deleteObjectTypes = (knId: string, obIds: string[], forceDelete?: boolean): Promise<void> => {
  const query = forceDelete !== undefined ? `?force_delete=${forceDelete}` : '';
  return Request.delete(`${BASE_URL}/${knId}/object-types/${obIds.join(',')}${query}`);
};

export const getOperatorList = (params?: { page: number; page_size: number; execution_mode?: 'sync' | 'async' }): Promise<OntologyObjectType.OperatorList> => {
  return Request.get(OPERATOR_BASE_URL, params);
};

export const getMetricModelList = (params?: OntologyObjectType.ListQuery): Promise<OntologyObjectType.MetricModelList> => {
  return Request.get(METRIC_MODEL_BASE_URL, params);
};

export const getMetricModelFields = (id: string): Promise<OntologyObjectType.MetricModelField[]> => {
  return Request.get(`${METRIC_MODEL_PREVIEW_BASE_URL}/${id}/fields`);
};

export const getSmallModelList = (params?: { page?: number; size?: number; model_name?: string }): Promise<OntologyObjectType.SmallModelList> => {
  return Request.get(`${SMALL_MODEL_BASE_URL}?model_type=embedding`, params);
};

export const updateObjectIndex = (knId: string, obId: string, names: string[], payload: OntologyObjectType.DataProperty[]): Promise<void> => {
  return Request.put(`${BASE_URL}/${knId}/object-types/${obId}/data_properties/${names.join(',')}`, { entries: payload });
};

export default {
  objectGet,
  createObjectTypes,
  updateObjectType,
  getDetail,
  deleteObjectTypes,
  getOperatorList,
  getMetricModelList,
  getMetricModelFields,
  getSmallModelList,
  updateObjectIndex,
};
