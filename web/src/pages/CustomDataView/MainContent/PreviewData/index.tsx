import { useEffect, useRef, useState } from 'react';
import intl from 'react-intl-universal';
import { Button, Form, Switch, Table } from 'antd';
import { INIT_FILTER } from '@/hooks/useConstants';
import api from '@/services/customDataView/index';
import noData from '@/assets/images/no-right.svg';
import { Drawer, Tooltip } from '@/web-library/common';
import DataFilter from '@/web-library/components/DataFilter';
import UTILS from '@/web-library/utils';
import styles from './index.module.less';

interface PreviewDataProps {
  open: boolean;
  id: string;
  name?: string;
  params?: any;
  onClose: () => void;
}

interface FieldItem {
  displayName: string;
  name: string;
  type: string;
}

const PreviewData: React.FC<PreviewDataProps> = ({ open, id, name, params = {}, onClose }) => {
  const [form] = Form.useForm();
  const [columns, setColumns] = useState<any[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [tableData, setTableData] = useState<any[]>([]);
  const [title, setTitle] = useState<string>('');
  const [switchFilter, setSwitchFilter] = useState<boolean>(false);
  const [fieldList, setFieldList] = useState<FieldItem[]>([]);
  const dataFilterRef = useRef<any>(null);
  const [isForbidden, setIsForbidden] = useState<boolean>(false);

  // 重置数据状态
  const resetData = () => {
    setTableData([]);
    setColumns([]);
    setTitle('');
    setFieldList([]);
    setIsForbidden(false);
  };

  // 渲染单元格内容
  const renderCellContent = (text: any) => {
    const content = typeof text !== 'string' ? JSON.stringify(text) : text;
    return (
      <Tooltip title={content}>
        <span style={{ display: 'inline-block', maxWidth: '100%', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{content}</span>
      </Tooltip>
    );
  };

  const getDetail = async (values?: any): Promise<void> => {
    setLoading(true);
    resetData();
    try {
      const resData = await api.getViewDataPreview(id, { limit: 1000, offset: 0, ...values, ...params });

      if (resData?.view) {
        // 过滤字段
        let viewFields = resData.view.fields;
        if (params?.output_fields) {
          viewFields = viewFields.filter((item: any) => params.output_fields.includes(item.name));
        }

        // 构建列配置和字段列表
        const cols = viewFields.map((item: any) => ({
          title: item.display_name,
          dataIndex: item.name,
          width: 180,
          ellipsis: true,
          render: renderCellContent,
        }));

        const fields: FieldItem[] = viewFields.map((item: any) => ({
          displayName: item.display_name,
          name: item.name,
          type: item.type,
        }));

        setFieldList(fields);
        setColumns(cols);
        setTableData(resData.entries);
        setTitle(resData.view.name);
      }
    } catch (error: any) {
      resetData();
      if (error?.status === 403) {
        setIsForbidden(true);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (open) {
      getDetail();
    } else {
      resetData();
      setSwitchFilter(false);
      form.setFieldValue('dataFilter', INIT_FILTER);
    }
  }, [open]);

  const handleSearch = () => {
    if (switchFilter) {
      const validate = dataFilterRef.current?.validate();
      if (!validate) {
        const filters = form.getFieldValue('dataFilter');
        getDetail({ filters });
      }
    } else {
      getDetail();
    }
  };

  return (
    <Drawer size="large" title={title || name || intl.get('Global.dataPreview')} width={'100%'} onClose={onClose} open={open}>
      {isForbidden ? (
        <>
          <div className="g-flex-center g-c-text-sub" style={{ flexDirection: 'column', height: 100, marginTop: 220 }}>
            <img src={noData} />
            <div style={{ marginTop: 8, color: 'rgba(0, 0, 0, 0.65)' }}>{intl.get('Global.noPreviewPermission', { title: title || name })}</div>
          </div>
        </>
      ) : (
        <>
          <div className={styles['search-box']}>
            <div className={styles['switch-box']}>
              <div className={styles['config-title']}>{intl.get('Global.dataFilter')}</div>
              <Switch
                size="small"
                value={switchFilter}
                onChange={(e) => {
                  setSwitchFilter(e);
                }}
              />
            </div>
            <Button type="primary" onClick={handleSearch}>
              {intl.get('Global.search')}
            </Button>
          </div>
          <div style={{ marginBottom: 12, display: switchFilter ? 'block' : 'none' }}>
            <Form form={form}>
              <Form.Item name="dataFilter">
                <DataFilter ref={dataFilterRef} fieldList={fieldList} required transformType={UTILS.formatType} maxCount={[10, 10, 10]} level={3} isFirst />
              </Form.Item>
            </Form>
          </div>
          <Table loading={loading} scroll={{ x: 1000, y: 'calc(100vh - 220px)' }} size="small" columns={columns} dataSource={tableData} />
        </>
      )}
    </Drawer>
  );
};

export default PreviewData;
