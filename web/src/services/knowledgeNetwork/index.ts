import UTILS from '@/utils';
import Request, { baseConfig } from '../request';
import * as KnowledgeNetworkType from './type';

const BASE_URL = '/api/ontology-manager/v1/knowledge-networks';

/**
 * 获取业务域ID (从 frameworkProps 中获取)
 */
const getBusinessDomainID = () => baseConfig?.businessDomainID;

/**
 * 查询业务知识网络列表
 * @param params 查询参数
 */
export const getNetworkList = async (
  params: KnowledgeNetworkType.GetNetworkListParams
): Promise<KnowledgeNetworkType.List<KnowledgeNetworkType.KnowledgeNetwork>> => {
  const queryParams = UTILS.filterEmptyFields({
    ...params,
    sort: params.sort || 'update_time',
    direction: params.direction || 'desc',
    offset: params.offset || 0,
    limit: params.limit || 10,
  });

  return await Request.get(BASE_URL, queryParams, {
    headers: { 'x-business-domain': getBusinessDomainID() },
  });
};

/**
 * 创建新的业务知识网络
 * @param data 创建数据
 * @param importMode 导入模式 'ignore' | 'overwrite'
 */
export const createNetwork = async (data: KnowledgeNetworkType.CreateNetworkRequest, importMode?: 'ignore' | 'overwrite'): Promise<any> => {
  const validate_dependency = data?.validate_dependency !== undefined ? data.validate_dependency : true;
  const url = importMode
    ? `${BASE_URL}?validate_dependency=${validate_dependency}&import_mode=${importMode}`
    : `${BASE_URL}?validate_dependency=${validate_dependency}`;
  return await Request.post(url, data, {
    isNoHint: true,
    headers: { 'x-business-domain': getBusinessDomainID() },
  });
};

/**
 * 修改业务知识网络
 * @param knId 知识网络ID
 * @param data 更新数据
 */
export const updateNetwork = async (knId: string, data: KnowledgeNetworkType.UpdateNetworkRequest): Promise<KnowledgeNetworkType.KnowledgeNetwork> => {
  return await Request.put(`${BASE_URL}/${knId}`, data);
};

/**
 * 获取业务知识网络详情
 * @param params 详情查询参数
 */
export const getNetworkDetail = async (params: KnowledgeNetworkType.GetNetworkDetailParams): Promise<any> => {
  const { knId, mode, include_detail, include_statistics } = params;
  const queryParams = {
    include_detail: !!include_detail,
    mode,
    include_statistics: include_statistics,
  };
  const res = await Request.get(`${BASE_URL}/${knId}`, queryParams);
  return res;
};

/**
 * 删除业务知识网络
 * @param knIds 知识网络ID列表
 */
export const deleteNetwork = async (knIds: string[]): Promise<any> => {
  return await Request.delete(`${BASE_URL}/${knIds.join(',')}`);
};

export default {
  getNetworkList,
  createNetwork,
  updateNetwork,
  getNetworkDetail,
  deleteNetwork,
};
