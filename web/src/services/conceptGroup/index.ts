import Request from '../request';
import * as ConceptGroupType from './type';

const BASE_URL = '/api/ontology-manager/v1/knowledge-networks';

export const createConceptGroup = (knId: string, data: ConceptGroupType.CreateRequest, mode?: 'ignore' | 'overwrite'): Promise<void> => {
  const validate_dependency = data?.validate_dependency !== undefined ? data.validate_dependency : true;
  const url = mode
    ? `${BASE_URL}/${knId}/concept-groups?validate_dependency=${validate_dependency}&import_mode=${mode}`
    : `${BASE_URL}/${knId}/concept-groups?validate_dependency=${validate_dependency}`;
  return Request.post(url, data, { isNoHint: true });
};

export const updateConceptGroup = (knId: string, groupId: string, data: ConceptGroupType.UpdateRequest): Promise<void> => {
  return Request.put(`${BASE_URL}/${knId}/concept-groups/${groupId}`, data);
};

export const deleteConceptGroup = (knId: string, groupIds: string[]): Promise<void> => {
  return Request.delete(`${BASE_URL}/${knId}/concept-groups/${groupIds.join(',')}`);
};

export const queryConceptGroups = (knId: string, params: ConceptGroupType.ListQuery): Promise<ConceptGroupType.List> => {
  return Request.get<ConceptGroupType.List>(`${BASE_URL}/${knId}/concept-groups`, params);
};

export const detailConceptGroup = (knId: string, groupId: string): Promise<ConceptGroupType.Detail> => {
  return Request.get<ConceptGroupType.Detail>(`${BASE_URL}/${knId}/concept-groups/${groupId}`);
};

export const addObjectTypesToGroup = (knId: string, groupId: string, data: ConceptGroupType.AddObjectTypesRequest): Promise<void> => {
  return Request.post(`${BASE_URL}/${knId}/concept-groups/${groupId}/object-types`, data);
};

export const removeObjectTypesFromGroup = (knId: string, groupId: string, otIds: string[]): Promise<void> => {
  return Request.delete(`${BASE_URL}/${knId}/concept-groups/${groupId}/object-types/${otIds.join(',')}`);
};

export default {
  createConceptGroup,
  updateConceptGroup,
  deleteConceptGroup,
  queryConceptGroups,
  detailConceptGroup,
  addObjectTypesToGroup,
  removeObjectTypesFromGroup,
};
