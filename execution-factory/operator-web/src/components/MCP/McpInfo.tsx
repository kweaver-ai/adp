import { useMemo } from 'react';
import { Collapse } from 'antd';
import './style.less';
import { dereference, getTableData } from '@/utils/operator';
import JsonschemaTab from '../MyOperator/JsonschemaTab';
import { EditOutlined, InteractionOutlined, ProfileOutlined } from '@ant-design/icons';

const { Panel } = Collapse;

export default function McpInfo({ selectedTool }: any) {
  // 处理json schema成table数据，供JsonschemaTab使用
  const tableData = useMemo(() => {
    if (selectedTool?.inputSchema) {
      const tableData: any[] = [];

      for (const location in selectedTool?.inputSchema?.properties) {
        const params = dereference(selectedTool?.inputSchema?.properties[location], selectedTool?.inputSchema);
        tableData.push(getTableData(params, location, location));
      }

      return tableData.flat();
    }
  }, [selectedTool?.inputSchema]);

  return (
    <div className="operator-info">
      <Collapse ghost defaultActiveKey={''} expandIconPosition="end" className="operator-details-collapse">
        <Panel
          key="1"
          header={
            <span>
              <ProfileOutlined /> 工具信息 <EditOutlined />
            </span>
          }
        >
          <div style={{ padding: '0 16px' }}>
            <div className="operator-info-title">工具名称</div>
            <div className="operator-info-desc">{selectedTool?.name}</div>
            <div className="operator-info-title">工具描述</div>
            <div className="operator-info-desc">{selectedTool?.description || '暂无描述'}</div>
            {/* <div className='operator-info-title'>
                MCP规则
              </div> */}
            {/* <div className='operator-info-desc'>
                {selectedTool?.name}
              </div> */}
          </div>
        </Panel>
        <Panel
          key="2"
          header={
            <span>
              <InteractionOutlined /> 输入
            </span>
          }
        >
          <JsonschemaTab data={tableData} type="Inputs" />
        </Panel>
      </Collapse>
    </div>
  );
}
