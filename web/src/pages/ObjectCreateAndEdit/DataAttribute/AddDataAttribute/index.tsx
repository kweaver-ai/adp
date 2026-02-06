import { useEffect, useRef, useState, useMemo } from 'react';
import intl from 'react-intl-universal';
import { Form, Input, Select, Switch, Tooltip } from 'antd';
import { useForm } from 'antd/es/form/Form';
import * as OntologyObjectType from '@/services/object/type';
import HOOKS from '@/hooks';
import { Drawer, Button, IconFont } from '@/web-library/common';
import Fields from '@/web-library/utils/fields';
import styles from './index.module.less';

const canAddIncrementalKeys = ['integer', 'unsigned integer', 'datetime', 'timestamp'];
const canPrimaryKeys = ['integer', 'unsigned integer', 'string'];
const canTitleKeys = ['integer', 'unsigned integer', 'string', 'text', 'float', 'decimal', 'date', 'time', 'datetime', 'timestamp', 'ip', 'boolean'];

const canBePrimaryKey = (type: string) => canPrimaryKeys.includes(type);
const canBeDisplayKey = (type: string) => canTitleKeys.includes(type);
const canBeIncrementalKey = (type: string) => canAddIncrementalKeys.includes(type);

export type TAddDataAttributeError = {
  name?: string;
  display_name?: string;
};

export type TAddDataAttribute = {
  data?: OntologyObjectType.Field;
  open: boolean;
  onClose: () => void;
  onOk: (data: OntologyObjectType.Field) => TAddDataAttributeError | void;
  onDelete: (data: OntologyObjectType.Field) => void;
};

const AddDataAttribute = (props: TAddDataAttribute) => {
  const { data, open, onClose, onOk, onDelete } = props;
  const [form] = useForm();
  const isDisplayNameManuallyEdited = useRef(false);
  const [currentType, setCurrentType] = useState<string>('string');
  const [isFormEdited, setIsFormEdited] = useState(false);
  const { modal } = HOOKS.useGlobalContext();

  useEffect(() => {
    const initializeForm = async () => {
      if (open && data) {
        form.setFieldsValue(data);
        setCurrentType(data.type || 'string');
        isDisplayNameManuallyEdited.current = true;
        setIsFormEdited(false);

        // 执行表单校验,检查编辑数据是否有错误
        try {
          await form.validateFields();
        } catch (error) {
          console.log('编辑数据校验失败:', error);
        }
      } else if (open) {
        form.resetFields();
        setCurrentType('string');
        isDisplayNameManuallyEdited.current = false;
        setIsFormEdited(false);
      }
    };

    initializeForm();
  }, [open, data, form]);

  const handleCancel = () => {
    form.resetFields();
    onClose();
  };

  const handleOk = async () => {
    try {
      const values = await form.validateFields();
      const errors = onOk({
        ...data,
        ...values,
        error: {},
      });

      if (errors && (errors.name || errors.display_name)) {
        const fieldErrors = [];
        if (errors.name) {
          fieldErrors.push({ name: 'name', errors: [errors.name] });
        }
        if (errors.display_name) {
          fieldErrors.push({ name: 'display_name', errors: [errors.display_name] });
        }
        form.setFields(fieldErrors);
        return;
      }
    } catch (error) {
      console.log('Validation failed:', error);
    }
  };

  const handleDelete = () => {
    if (!data) return;

    onDelete(data);
  };

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const name = e.target.value;
    if (!isDisplayNameManuallyEdited.current) {
      form.setFieldValue('display_name', name);
    }
  };

  const handleDisplayNameChange = () => {
    isDisplayNameManuallyEdited.current = true;
  };

  const handleTypeChange = (value: string) => {
    setCurrentType(value);
    if (!canBePrimaryKey(value)) {
      form.setFieldValue('primary_key', false);
    }
    if (!canBeDisplayKey(value)) {
      form.setFieldValue('display_key', false);
    }
    if (!canBeIncrementalKey(value)) {
      form.setFieldValue('incremental_key', false);
    }
  };

  const typeOptions = useMemo(
    () =>
      Fields.DataType_All.map((item) => ({
        label: item.name,
        value: item.name,
      })),
    []
  );

  const footer = (
    <div className={styles.footer}>
      <Button onClick={handleCancel}>{intl.get('Global.cancel')}</Button>
      <Button type="primary" onClick={handleOk}>
        {intl.get('Global.ok')}
      </Button>
    </div>
  );

  const title = (
    <div className={styles.title}>
      <div>{intl.get('Global.dataProperty')}</div>
      {data && (
        <>
          <div className={styles.line}></div>
          <div className={styles.delete} onClick={handleDelete}>
            <IconFont type="icon-dip-trash" /> {intl.get('Global.delete')}
          </div>
        </>
      )}
    </div>
  );

  return (
    <Drawer width={400} title={title} open={open} maskClosable={!isFormEdited} onClose={handleCancel} footer={footer}>
      <div className={styles.container}>
        <div className={styles.formContent}>
          <Form form={form} layout="vertical" autoComplete="off" onValuesChange={() => setIsFormEdited(true)}>
            <Form.Item
              name="name"
              label={intl.get('Global.attributeName')}
              rules={[
                { required: true, message: intl.get('Global.pleaseInput') },
                {
                  pattern: /^[a-z0-9][a-z0-9_-]*$/,
                  message: intl.get('Global.idPatternError'),
                },
              ]}
            >
              <Input.TextArea autoSize={{ minRows: 2, maxRows: 4 }} placeholder={intl.get('Global.pleaseInput')} onChange={handleNameChange} />
            </Form.Item>

            <Form.Item name="display_name" label={intl.get('Global.displayName')} rules={[{ required: true, message: intl.get('Global.pleaseInput') }]}>
              <Input.TextArea autoSize={{ minRows: 2, maxRows: 4 }} placeholder={intl.get('Global.pleaseInput')} onChange={handleDisplayNameChange} />
            </Form.Item>

            <Form.Item name="type" label={intl.get('Global.attributeType')} initialValue="string">
              <Select options={typeOptions} placeholder={intl.get('Global.pleaseSelect')} onChange={handleTypeChange} />
            </Form.Item>

            <Form.Item name="comment" label={intl.get('Global.description')}>
              <Input.TextArea autoSize={{ minRows: 3, maxRows: 7 }} placeholder={intl.get('Global.pleaseInput')} maxLength={1000} showCount />
            </Form.Item>

            <Form.Item
              name="primary_key"
              layout="horizontal"
              label={
                <div className={styles.switchItem}>
                  <span className={styles.switchLabel}>{intl.get('Global.primaryKey')}</span>
                  <Tooltip title={intl.get('Object.primaryKeyTip')}>
                    <IconFont type="icon-dip-color-tip" className={styles.helpIcon} />
                  </Tooltip>
                </div>
              }
              valuePropName="checked"
              initialValue={false}
            >
              <Switch size="small" style={{ marginLeft: 260 }} disabled={!canBePrimaryKey(currentType)} />
            </Form.Item>

            <Form.Item
              name="display_key"
              layout="horizontal"
              label={
                <div className={styles.switchItem}>
                  <span className={styles.switchLabel}>{intl.get('Global.title')}</span>
                  <Tooltip title={intl.get('Object.displayKeyTip')}>
                    <IconFont type="icon-dip-color-tip" className={styles.helpIcon} />
                  </Tooltip>
                </div>
              }
              valuePropName="checked"
              initialValue={false}
            >
              <Switch size="small" style={{ marginLeft: 260 }} disabled={!canBeDisplayKey(currentType)} />
            </Form.Item>

            <Form.Item
              name="incremental_key"
              layout="horizontal"
              label={
                <div className={styles.switchItem}>
                  <span className={styles.switchLabel}>{intl.get('Object.incrementalKey')}</span>
                  <Tooltip title={intl.get('Object.incrementalKeyTip')}>
                    <IconFont type="icon-dip-color-tip" className={styles.helpIcon} />
                  </Tooltip>
                </div>
              }
              valuePropName="checked"
              initialValue={false}
            >
              <Switch size="small" style={{ marginLeft: 260 }} disabled={!canBeIncrementalKey(currentType)} />
            </Form.Item>
          </Form>
        </div>
      </div>
    </Drawer>
  );
};

export default AddDataAttribute;
