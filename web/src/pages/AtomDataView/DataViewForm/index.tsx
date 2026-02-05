import { useState, useEffect } from 'react';
import intl from 'react-intl-universal';
import { LeftOutlined } from '@ant-design/icons';
import { Button, Col, Divider, Form, Input, Row } from 'antd';
import FieldFeatureModal from '@/components/FieldFeatureModal';
import { FORM_LAYOUT } from '@/hooks/useConstants';
import api from '@/services/atomDataView';
import * as AtomDataViewType from '@/services/atomDataView/type';
import * as DataConnectType from '@/services/dataConnect/type';
import HOOKS from '@/hooks';
import { Drawer } from '@/web-library/common';
import FieldTable from './fieldTable';
import styles from './index.module.less';

interface PropType {
  id: string;
  visible: boolean;
  checkDatasource?: DataConnectType.DataSource & { excel_file_name: string };
  onClose: (val?: boolean) => void;
}

const { TextArea } = Input;

const DataViewForm = ({ visible, onClose, id, checkDatasource }: PropType): JSX.Element => {
  const [form] = Form.useForm();
  const { resetFields, setFieldsValue } = form;
  const { message } = HOOKS.useGlobalContext();

  const [fieldData, setFieldData] = useState<AtomDataViewType.Field[]>([]);
  const [detail, setDetail] = useState<AtomDataViewType.Data>();
  const [loading, setLoading] = useState<boolean>(false);
  const [page, setPage] = useState<number>(1);
  const [featureModalVisible, setFeatureModalVisible] = useState<boolean>(false);
  const [currentField, setCurrentField] = useState<AtomDataViewType.Field | null>(null);

  const getDetail = async (): Promise<void> => {
    const resDetail = await api.getDataViewsByIds([id]);

    if (resDetail && resDetail.length > 0) {
      console.log(resDetail[0], 'resDetail[0]');
      const cur = resDetail[0];
      setFieldsValue({ name: cur.name, technical_name: cur.technical_name, comment: cur.comment });
      setDetail(cur);
      setFieldData(cur.fields);
      setPage(1);
    }
  };

  useEffect(() => {
    setFieldData([]);
    setPage(1);
    setDetail(undefined);
    if (id) {
      getDetail();
    }
  }, [id]);

  const onCancel = async (bool?: boolean): Promise<void> => {
    resetFields();
    setDetail(undefined);
    setFieldData([]);
    setPage(1);
    onClose(!!bool);
  };

  const validateFields = (): boolean => {
    // 校验字段数据
    for (const field of fieldData) {
      // 校验字段显示名称
      if (!field.display_name || !field.display_name.trim()) {
        message.error(`${intl.get('Global.fieldDisplayName')}: ${intl.get('Global.cannotBeNull')}`);
        return false;
      }
      if (field.display_name.length > 255) {
        message.error(`${intl.get('Global.fieldDisplayName')}: ${intl.get('DataView.displayNameMaxLength')}`);
        return false;
      }

      // 校验字段名称
      if (!field.name || !field.name.trim()) {
        message.error(`${intl.get('Global.fieldName')}: ${intl.get('Global.cannotBeNull')}`);
        return false;
      }
    }

    // 校验字段显示名称重复
    const displayNames = fieldData.map((f) => f.display_name);
    const displayNameSet = new Set(displayNames);
    if (displayNames.length !== displayNameSet.size) {
      message.error(`${intl.get('Global.fieldDisplayName')}: ${intl.get('Global.nameCannotRepeat')}`);
      return false;
    }

    // 校验字段名称重复
    const names = fieldData.map((f) => f.name);
    const nameSet = new Set(names);
    if (names.length !== nameSet.size) {
      message.error(`${intl.get('Global.fieldName')}: ${intl.get('Global.fieldNameCannotRepeat')}`);
      return false;
    }

    return true;
  };

  const onOk = async (): Promise<void> => {
    // 先校验表单
    form.validateFields().then(async (values) => {
      // 再校验字段数据
      if (!validateFields()) {
        return;
      }

      setLoading(true);
      try {
        const { name, comment } = values;
        const paramsFields = {
          fields: fieldData,
          name,
          comment,
        };

        await api.updateDataViewAttrs(id, paramsFields);

        setLoading(false);
        message.success(intl.get('Global.saveSuccess'));
        onCancel(true);
        return;
      } catch {
        setLoading(false);
      }
    });
  };

  const title = (
    <div className={styles['box-exit']}>
      <div className={styles['exit']} onClick={() => onCancel()}>
        <LeftOutlined />
        <span>{intl.get('Global.exit')}</span>
      </div>
      <Divider type="vertical" />
      <span>
        {intl.get('DataView.editAtomView')}：{detail?.name}
      </span>
    </div>
  );

  return (
    <Drawer title={title} open={visible} onClose={() => onCancel()} width={'100vw'} closable={false} maskClosable={false}>
      <div className={styles['dataview-form-wrapper']}>
        <Form form={form} {...FORM_LAYOUT} labelAlign="left">
          <div className={styles['title']}>
            <span> </span>
            {intl.get('Global.basicConfig')}
          </div>
          <Form.Item
            label={intl.get('Global.name')}
            name="name"
            preserve={true}
            initialValue={detail?.name}
            rules={[
              {
                required: true,
                message: intl.get('Global.nameCannotNull'),
              },
              {
                max: 100,
                message: intl.get('Global.nameCannotOverFourty'),
              },
            ]}
          >
            <Input placeholder={intl.get('Global.pleaseInput')} />
          </Form.Item>
          <Form.Item
            label={intl.get('DataView.technicalName')}
            name="technical_name"
            preserve={true}
            initialValue={detail?.technical_name}
            rules={[
              {
                required: true,
                message: intl.get('DataView.technicalNameCannotNull'),
              },
              {
                validator: (_rule, value: string, callback): void => {
                  if (value && value.length > 255) {
                    callback(intl.get('DataView.technicalNameMaxLength'));
                  }
                  callback();
                },
              },
            ]}
          >
            <Input placeholder={intl.get('Global.pleaseInput')} disabled />
          </Form.Item>
          <Form.Item label={intl.get('Global.comment')} name="comment" preserve={true} initialValue={detail?.comment || ''}>
            <TextArea rows={4} maxLength={255} />
          </Form.Item>
        </Form>

        {(detail?.file_name || checkDatasource?.excel_file_name) && (
          <>
            <div className={styles['title']}>
              <span> </span>
              {intl.get('DataView.dataRange')}
            </div>
            <Row gutter={24} style={{ marginBottom: 12 }}>
              <Col span={12}>
                <b>{intl.get('DataView.sheetSpa')}</b>: {detail?.excel_config?.sheet}
              </Col>
              <Col span={12}>
                <b>{intl.get('DataView.cellRange')}</b>:{detail?.excel_config?.start_cell}-{detail?.excel_config?.end_cell}
              </Col>
            </Row>
            <Row gutter={24} style={{ marginBottom: 12 }}>
              <Col span={12}>
                <b>{intl.get('DataView.hasHeaders')}</b>: {detail?.excel_config?.has_headers ? intl.get('Global.selectFirstRow') : intl.get('Global.custom')}
              </Col>
              <Col span={12}>
                <b>{intl.get('DataView.sheetAsNewColumn')}</b>: {detail?.excel_config?.sheet_as_new_column ? intl.get('Global.yes') : intl.get('Global.no')}
              </Col>
            </Row>
          </>
        )}
        <div className={styles['title']}>
          <span> </span>
          {intl.get('Global.fieldInfo')}
        </div>
        <FieldTable
          dataSource={fieldData}
          onChange={(val): void => setFieldData(val)}
          page={page}
          setPage={setPage}
          isEdit={!!id}
          onFieldFeatureClick={(field): void => {
            setCurrentField(field);
            setFeatureModalVisible(true);
          }}
        />
      </div>
      <div className={styles.footer}>
        <Button onClick={onOk} type="primary" style={{ marginRight: 8 }} loading={loading}>
          {intl.get('Global.save')}
        </Button>
        <Button onClick={(): Promise<void> => onCancel()} disabled={loading}>
          {intl.get('Global.cancel')}
        </Button>
      </div>
      <FieldFeatureModal
        visible={featureModalVisible}
        mode="edit"
        fieldName={currentField?.display_name || currentField?.name}
        data={currentField?.features || []}
        fields={fieldData}
        onCancel={(): void => {
          setFeatureModalVisible(false);
          setCurrentField(null);
        }}
        onOk={(data): void => {
          setFieldData((prevFieldData) =>
            prevFieldData.map((field) => {
              if (currentField && field.original_name === currentField.original_name) {
                return { ...field, features: data };
              }
              return field;
            })
          );
          setFeatureModalVisible(false);
          setCurrentField(null);
        }}
        onPrev={(features) => {
          if (!currentField || !fieldData.length) return false;
          // 先保存当前字段的 features
          setFieldData((prevFieldData) =>
            prevFieldData.map((field) => {
              if (field.original_name === currentField.original_name) {
                return { ...field, features };
              }
              return field;
            })
          );
          // 切换到上一个字段
          const currentIndex = fieldData.findIndex((f) => f.original_name === currentField.original_name);
          if (currentIndex > 0) {
            const prevField = fieldData[currentIndex - 1];
            setCurrentField(prevField);
            return true;
          }
          return false;
        }}
        onNext={(features) => {
          if (!currentField || !fieldData.length) return false;
          // 先保存当前字段的 features
          setFieldData((prevFieldData) =>
            prevFieldData.map((field) => {
              if (field.original_name === currentField.original_name) {
                return { ...field, features };
              }
              return field;
            })
          );
          // 切换到下一个字段
          const currentIndex = fieldData.findIndex((f) => f.original_name === currentField.original_name);
          if (currentIndex < fieldData.length - 1) {
            const nextField = fieldData[currentIndex + 1];
            setCurrentField(nextField);
            return true;
          }
          return false;
        }}
      />
    </Drawer>
  );
};

export default DataViewForm;
