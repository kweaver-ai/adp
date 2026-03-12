import React, { useEffect, useMemo, useState } from 'react';
import intl from 'react-intl-universal';
import { Empty, Spin } from 'antd';
import ObjectIcon from '@/components/ObjectIcon';
import SERVICE from '@/services';
import { Button, IconFont, Modal } from '@/web-library/common';
import styles from './index.module.less';

export interface ObjectBriefInfo {
  id: string;
  name: string;
  icon?: string;
  color?: string;
}

export interface AffectedResourceItem {
  key: string;
  object: ObjectBriefInfo;
  affectedType: 'relation' | 'action';
  affectedName: string;
}

interface DeleteConfirmModalProps {
  open: boolean;
  loading?: boolean;
  objects: ObjectBriefInfo[];
  onCancel: () => void;
  onConfirm: () => void;
}

const AFFECTED_TYPE_MAP = {
  relation: { icon: 'icon-dip-guanxilei', color: '#5381DF' },
  action: { icon: 'icon-dip-hangdonglei', color: '#1f1f1f' },
};

interface EdgeLike {
  id?: string;
  name?: string;
  source_object_type_id?: string;
  source_type_id?: string;
  target_object_type_id?: string;
  target_type_id?: string;
  source_object_type?: { name?: string; icon?: string; color?: string };
  target_object_type?: { name?: string; icon?: string; color?: string };
}

interface ActionLike {
  id?: string;
  name?: string;
  object_type_id?: string;
  object_type?: { id?: string; name?: string; icon?: string; color?: string };
}

const DeleteConfirmModal: React.FC<DeleteConfirmModalProps> = ({ open, loading, objects, onCancel, onConfirm }) => {
  const [affectedRows, setAffectedRows] = useState<AffectedResourceItem[]>([]);
  const [isImpactLoading, setIsImpactLoading] = useState(false);

  const objectIds = useMemo(() => objects.map((item) => item.id).filter(Boolean), [objects]);

  const getAffectedRowsByObjectIds = async (ids: string[]): Promise<AffectedResourceItem[]> => {
    const knId = localStorage.getItem('KnowledgeNetwork.id');
    if (!knId || !ids.length) return [];

    // TODO: 替换为后端“删除影响检查”接口
    const [edgeResult, actionResult] = await Promise.all([
      SERVICE.edge.getEdgeList(knId, { offset: 0, limit: -1 }),
      SERVICE.action.getActionTypes(knId, { offset: 0, limit: -1 }),
    ]);
    const edgeEntries = edgeResult?.entries || [];
    const actionEntries = actionResult?.entries || [];
    const idSet = new Set(ids);
    const objectMap = objects.reduce<Record<string, ObjectBriefInfo>>((acc, cur) => {
      acc[cur.id] = cur;
      return acc;
    }, {});

    const rows: AffectedResourceItem[] = [];
    const keySet = new Set<string>();

    edgeEntries.forEach((edge: EdgeLike) => {
      const sourceObjectId = edge?.source_object_type_id || edge?.source_type_id;
      const targetObjectId = edge?.target_object_type_id || edge?.target_type_id;
      const relationName = edge?.name || '--';
      const relationId = edge?.id || relationName;

      const hitObjectIds = [sourceObjectId, targetObjectId].filter((id, index, arr): id is string => Boolean(id) && arr.indexOf(id) === index);
      hitObjectIds.forEach((objectId) => {
        if (!idSet.has(objectId)) return;

        const objectInfo = objectMap[objectId] || {
          id: objectId,
          name: edge?.source_object_type?.name || edge?.target_object_type?.name || '--',
          icon: edge?.source_object_type?.icon || edge?.target_object_type?.icon,
          color: edge?.source_object_type?.color || edge?.target_object_type?.color,
        };

        const key = `relation-${objectId}-${relationId}`;
        if (keySet.has(key)) return;
        keySet.add(key);

        rows.push({
          key,
          object: objectInfo,
          affectedType: 'relation',
          affectedName: relationName,
        });
      });
    });

    actionEntries.forEach((action: ActionLike) => {
      const objectId = action?.object_type_id || action?.object_type?.id;
      if (!objectId || !idSet.has(objectId)) return;

      const objectInfo = objectMap[objectId] || {
        id: objectId,
        name: action?.object_type?.name || '--',
        icon: action?.object_type?.icon,
        color: action?.object_type?.color,
      };

      const actionName = action?.name || '--';
      const actionId = action?.id || actionName;
      const key = `action-${objectId}-${actionId}`;
      if (keySet.has(key)) return;
      keySet.add(key);

      rows.push({
        key,
        object: objectInfo,
        affectedType: 'action',
        affectedName: actionName,
      });
    });

    return rows;
  };

  useEffect(() => {
    if (!open || !objectIds.length) {
      setAffectedRows([]);
      return;
    }

    let unmounted = false;
    const fetchAffectedRows = async () => {
      setIsImpactLoading(true);
      try {
        const rows = await getAffectedRowsByObjectIds(objectIds);
        if (!unmounted) setAffectedRows(rows);
      } catch (error) {
        console.error('getAffectedRowsByObjectIds error:', error);
        if (!unmounted) setAffectedRows([]);
      } finally {
        if (!unmounted) setIsImpactLoading(false);
      }
    };

    fetchAffectedRows();
    return () => {
      unmounted = true;
    };
  }, [open, objectIds]);

  return (
    <Modal
      open={open}
      title={intl.get('Global.tipTitle')}
      width={840}
      onCancel={onCancel}
      footer={[
        <Button key="confirm-delete" type="primary" danger loading={loading} onClick={onConfirm}>
          {intl.get('Object.deleteIgnoreAndConfirm')}
        </Button>,
        <Button key="cancel" onClick={onCancel}>
          {intl.get('Global.cancel')}
        </Button>,
      ]}
    >
      <div className={styles['delete-modal-content']}>
        <p className={styles['desc-line']}>{intl.get('Object.deleteObjectUsedByRelationAndAction')}</p>
        <p className={styles['desc-line']}>{intl.get('Object.affectedRelationAndActionList')}</p>

        <div className={styles['table-box']}>
          <div className={styles['table-header']}>
            <div className={styles['col-left']}>{intl.get('Object.objectToDelete')}</div>
            <div className={styles['col-right']}>{intl.get('Object.affectedRelationOrAction')}</div>
          </div>

          <div className={styles['table-body']}>
            {isImpactLoading ? (
              <div className={styles['empty-box']}>
                <Spin />
              </div>
            ) : affectedRows.length ? (
              affectedRows.map((row) => {
                const typeConfig = AFFECTED_TYPE_MAP[row.affectedType];
                return (
                  <div key={row.key} className={styles['table-row']}>
                    <div className={styles['col-left']}>
                      <ObjectIcon icon={row.object.icon} color={row.object.color} />
                      <span className="g-ellipsis-1">{row.object.name}</span>
                    </div>
                    <div className={styles['col-right']}>
                      <div className={styles['affected-icon']} style={{ background: typeConfig.color }}>
                        <IconFont type={typeConfig.icon} style={{ color: '#fff', fontSize: 14 }} />
                      </div>
                      <span className="g-ellipsis-1">{row.affectedName}</span>
                    </div>
                  </div>
                );
              })
            ) : (
              <div className={styles['empty-box']}>
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={intl.get('Global.noData')} />
              </div>
            )}
          </div>
        </div>

        <div className={styles['confirm-tip']}>{intl.get('Object.confirmDeleteObjectClass')}</div>
      </div>
    </Modal>
  );
};

export default DeleteConfirmModal;
