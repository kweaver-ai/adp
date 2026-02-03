import { FC, useEffect, useState } from 'react';
import intl from 'react-intl-universal';
import { Spin, Tag, Tooltip } from 'antd';
import dayjs from 'dayjs';
import { map } from 'lodash-es';
import DataFilterNew from '@/components/DataFilterNew';
import { renderObjectTypeLabel } from '@/components/ObjectSelector';
import ToolParamsTable from '@/components/ToolParamsTable';
import * as ActionType from '@/services/action/type';
import api from '@/services/tool';
import SERVICE from '@/services';
import { IconFont } from '@/web-library/common';
import UTILS from '@/web-library/utils';
import styles from './index.module.less';

const getScheduleContent = ({ type, expression }: { type?: ActionType.ActionScheduleTypeEnum; expression?: string }) => {
  if (type === ActionType.ActionScheduleTypeEnum.FixRate && expression) {
    const time = expression.slice(0, -1);
    const unit: string = expression.slice(-1);
    const unitLabels: any = {
      d: intl.get('Global.unitDay'),
      m: intl.get('Global.unitMinute'),
      h: intl.get('Global.unitHour'),
    };
    return intl.get('Action.fixedRate', { time, unit: unitLabels[unit] || '' });
  }

  if (type === ActionType.ActionScheduleTypeEnum.Cron && expression) {
    return intl.get('Action.cronExpression', { expression });
  }

  return '--';
};

interface OverviewProps {
  knId: string;
  atId: string;
  detail?: any;
}

const Overview: FC<OverviewProps> = ({ knId, atId, detail }) => {
  const data = detail;
  const [tool, setTool] = useState<
    | { type?: string; tool_id?: string; tool_name: string; box_id?: string; box_name?: string; tool_description?: string; mcp_id?: string; mcp_name?: string }
    | undefined
  >(undefined);
  const [objectOptions, setObjectOptions] = useState<any[]>([]);

  /** 获取对象列表 */
  const getObjectList = async () => {
    try {
      const result = await SERVICE.object.objectGet(knId, { offset: 0, limit: -1 });
      const objectOptions = map(result?.entries, (item) => {
        const { id, name, icon, data_properties, color } = item;
        return {
          value: id,
          name,
          data_properties,
          label: renderObjectTypeLabel({ icon, name, color }),
          detail: item,
        };
      });
      setObjectOptions(objectOptions);
    } catch (error) {
      console.log('getObjectList error: ', error);
    }
  };

  useEffect(() => {
    getObjectList();
  }, []);

  // 请求工具信息
  const fetchToolDetails = async () => {
    if (!data?.action_source) {
      setTool(undefined);
      return;
    }

    const { type, box_id, tool_id, mcp_id, tool_name } = data.action_source;

    try {
      // 根据工具类型获取不同的工具详情
      if (type === 'tool' && box_id && tool_id) {
        const [{ box_name, tools }]: any = await api.getToolBoxDetail(box_id, ['box_name', 'tools']);
        const findTool = tools.find((tool: any) => tool.tool_id === tool_id);

        setTool({
          type,
          box_id,
          box_name,
          tool_id,
          tool_name: findTool?.name,
          tool_description: findTool?.description,
        });
      } else if (type === 'mcp' && mcp_id && tool_name) {
        const { tools: mcpTools } = await api.getMcpTools(mcp_id, { page: 1, page_size: 100, status: 'enabled', all: true });
        const {
          base_info: { name: mcp_name },
        } = await api.getMcpDetail(mcp_id);
        const findTool = mcpTools.find((tool: any) => tool.name === tool_name);

        setTool({
          type,
          mcp_id,
          mcp_name,
          tool_name: findTool?.name,
          tool_description: findTool?.description,
        });
      } else {
        setTool(undefined);
      }
    } catch (error) {
      console.error('获取工具详情失败:', error);
      setTool(undefined);
    }
  };

  useEffect(() => {
    fetchToolDetails();
  }, [data]);

  if (!detail) {
    return (
      <div className={styles['loading']}>
        <Spin />
      </div>
    );
  }

  const renderItem = (label: any, content: any) => (
    <div className={styles['info-item']} key={label}>
      <div className={styles['label']}>{label}</div>
      <div className={styles['content']}>
        {typeof content === 'string' ? (
          <Tooltip title={content.length > 50 ? content : ''}>
            <div className="g-ellipsis-1">{content}</div>
          </Tooltip>
        ) : (
          content
        )}
      </div>
    </div>
  );

  return (
    <div className={styles['overview-box']}>
      <div className={styles['overview-box-header']}>
        {data?.id && <div className={styles['id-text']}>ID:{data.id}</div>}
        <div className={styles['header-title']}>
          <div className={styles['header-title-left']}>
            <div className={styles['name-icon']} style={{ backgroundColor: '#90c06b' }}>
              <IconFont type="icon-dip-hangdonglei" style={{ color: '#fff', fontSize: 20 }} />
            </div>
            <div className={styles['name-text']}>{detail?.name}</div>
          </div>
          <div className={styles['header-title-right']}></div>
        </div>
        {data?.tags?.length > 0 && (
          <div className={styles['tags']}>
            {data.tags.map((tag: string) => (
              <Tag key={tag}>{tag}</Tag>
            ))}
          </div>
        )}
        <div className={styles['header-comment']}></div>
        <div className={styles['header-footer']}>
          <IconFont type="icon-dip-User" style={{ fontSize: 16 }} />
          <span style={{ padding: '0 5px' }}>{intl.get('Global.modifier')}:</span>
          <span style={{ marginRight: 20 }}>{detail?.updater?.name || detail?.creator?.name || '--'}</span>
          <IconFont type="icon-dip-history" style={{ fontSize: 16 }} />
          <span style={{ padding: '0 5px' }}>{intl.get('Global.updateTime')}:</span>
          <span>{detail?.update_time ? dayjs(detail.update_time).format('YYYY-MM-DD HH:mm:ss') : '--'}</span>
        </div>
      </div>

      {/* Basic Info */}
      <div className={styles['overview-card']}>
        <div className={styles['card-title']}>{intl.get('Global.basicInfo')}</div>
        {renderItem(
          intl.get('Action.actionType'),
          data.action_type === ActionType.ActionTypeEnum.Add
            ? intl.get('Global.add')
            : data.action_type === ActionType.ActionTypeEnum.Delete
              ? intl.get('Global.delete')
              : intl.get('Global.edit')
        )}
        {renderItem(intl.get('Action.boundObjectType'), data.object_type?.name ? renderObjectTypeLabel(data.object_type) : '--')}
        {renderItem(
          intl.get('Action.triggerCondition'),
          data.condition?.operation ? (
            <div style={{ width: 'fit-content' }}>
              <DataFilterNew
                isFirst
                disabled
                objectOptions={objectOptions}
                value={data.condition}
                level={3}
                maxCount={[10, 10, 10]}
                transformType={UTILS.formatType}
              />
            </div>
          ) : (
            '--'
          )
        )}
        {renderItem(intl.get('Action.affectedObjectType'), data.affect?.object_type?.name ? renderObjectTypeLabel(data.affect.object_type) : '--')}
        {renderItem(intl.get('Action.affectDescription'), data.affect?.comment || <span className="g-c-watermark">{intl.get('Global.noDescription')}</span>)}
        <div style={{height: 6}}></div>
      </div>

      {/* Resource Info */}
      <div className={styles['overview-card']}>
        <div className={styles['card-title']}>{intl.get('Action.resourceInfo')}</div>
        {tool ? (
          <div className={styles['tool']}>
            <IconFont type={tool.type === 'tool' ? 'icon-dip-color-suanzitool' : 'icon-dip-color-suanzi'} style={{ fontSize: 22 }} />
            <div className={styles['text-wrapper']}>
              {(() => {
                const toolPath = `${tool.box_name || tool.mcp_name || '--'}/${tool.tool_name || '--'}`;
                return (
                  <div className="g-ellipsis-1" style={{ lineHeight: '20px' }} title={toolPath}>
                    {toolPath}
                  </div>
                );
              })()}

              {tool.tool_description ? (
                <div className="g-ellipsis-1" title={tool.tool_description} style={{ color: 'rgba(0, 0, 0, 0.65)', fontSize: '12px', lineHeight: '18px' }}>
                  {tool.tool_description}
                </div>
              ) : (
                <span className="g-c-watermark">{intl.get('Global.noDescription')}</span>
              )}
            </div>
          </div>
        ) : (
          <div style={{ marginBottom: 10 }}>--</div>
        )}

        {data.action_source && (
          <div style={{ margin: '16px 0 20px' }}>
            <ToolParamsTable disabled actionSource={data.action_source} knId={knId} obId={data?.object_type?.id} value={data?.parameters} />
          </div>
        )}
      </div>

      {/* Run Strategy */}
      <div className={styles['overview-card']}>
        <div className={styles['card-title']}>{intl.get('Action.runStrategy')}</div>
        {renderItem(intl.get('Action.runStrategy'), getScheduleContent({ type: data?.schedule?.type, expression: data?.schedule?.expression }))}
      </div>
    </div>
  );
};

export default Overview;
