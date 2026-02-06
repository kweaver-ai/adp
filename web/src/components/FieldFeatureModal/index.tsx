import { useState, useEffect, FC } from 'react';
import intl from 'react-intl-universal';
import { PlusOutlined, CloseOutlined, MinusCircleFilled } from '@ant-design/icons';
import { Modal, Select, Input, Switch, Button } from 'antd';
import { nanoId } from '@/utils/dataView';
import { Tooltip } from '@/web-library/common';
import styles from './index.module.less';
import locales from './locales';

export interface FieldFeature {
  id: string;
  name: string;
  type: string;
  ref_field: string;
  is_default: boolean;
  is_native: boolean;
  comment: string;
  config?: object;
}

interface FieldFeatureModalProps {
  visible: boolean;
  title?: string;
  data?: FieldFeature[];
  fields?: any[];
  mode?: 'edit' | 'view'; // 新增模式属性:编辑模式或查看模式
  fieldName?: string; // 查看模式时显示的字段名称
  onCancel: () => void;
  onOk: (data: FieldFeature[]) => void;
  onPrev?: (data: FieldFeature[]) => boolean; // 上一个字段，编辑模式需要传递当前数据并返回是否允许切换
  onNext?: (data: FieldFeature[]) => boolean; // 下一个字段，编辑模式需要传递当前数据并返回是否允许切换
}

const FieldFeatureModal: FC<FieldFeatureModalProps> = ({
  visible,
  title,
  data = [],
  fields = [],
  mode = 'edit',
  fieldName,
  onCancel,
  onOk,
  onPrev,
  onNext,
}) => {
  const [fieldList, setFieldList] = useState<FieldFeature[]>([]);
  const [i18nLoaded, setI18nLoaded] = useState(false);
  const [errors, setErrors] = useState<Record<string, { name?: string; ref_field?: string }>>({});

  useEffect(() => {
    // 加载国际化文件,完成后更新状态触发重新渲染
    intl.load(locales);
    setI18nLoaded(true);
  }, []);

  // 当弹窗打开时，重置数据为传入的 data
  useEffect(() => {
    if (visible) {
      setFieldList(data.map((item) => ({ ...item, id: `field_${nanoId()}` })));
      setErrors({});
    }
  }, [visible]);

  const typeOptions = [
    { label: 'fulltext', value: 'fulltext' },
    { label: 'keyword', value: 'keyword' },
    { label: 'vector', value: 'vector' },
  ];

  const handleAdd = () => {
    if (fieldList.length >= 10) {
      return;
    }
    const newField: FieldFeature = {
      id: `field_${nanoId()}`,
      name: '',
      type: 'fulltext',
      ref_field: '',
      is_default: false,
      is_native: false,
      comment: '',
      config: undefined,
    };
    setFieldList((prev) => [...prev, newField]);
  };

  const handleDelete = (id: string) => {
    setFieldList((prev) => prev.filter((item) => item.id !== id));
  };

  // 根据类型过滤字段选项
  const getFilteredFieldsByType = (type: string) => {
    if (!fields || fields.length === 0) return [];

    const typeMapping: Record<string, string> = {
      fulltext: 'text',
      keyword: 'string',
      vector: 'vector',
    };

    const targetType = typeMapping[type];
    if (!targetType) return fields.map((field) => ({ label: field.name, value: field.name }));

    return fields.filter((field) => field.type === targetType).map((field) => ({ label: field.name, value: field.name }));
  };

  const handleFieldChange = (id: string, field: keyof FieldFeature, value: any) => {
    setFieldList((prev) => {
      const targetItem = prev.find((f) => f.id === id);

      return prev.map((item) => {
        if (item.id === id) {
          // 当类型改变时，检查当前 ref_field 是否在新过滤的选项中，如果不在则清空
          if (field === 'type') {
            const filteredFields = getFilteredFieldsByType(value);
            const isRefFieldValid = filteredFields.some((opt) => opt.value === item.ref_field);
            if (!isRefFieldValid) {
              return { ...item, [field]: value, ref_field: '' };
            }
          }
          return { ...item, [field]: value };
        }
        // 当某个字段设置为默认时，将同类型的其他字段设置为非默认
        if (field === 'is_default' && value === true && item.type === targetItem?.type) {
          return { ...item, [field]: false };
        }
        return item;
      });
    });
    // 清除该字段的错误状态
    if ((field === 'name' || field === 'ref_field') && errors[id]) {
      setErrors((prev) => {
        const newErrors = { ...prev };
        if (newErrors[id]) {
          delete newErrors[id][field];
          if (Object.keys(newErrors[id]).length === 0) {
            delete newErrors[id];
          }
        }
        return newErrors;
      });
    }
  };

  const validateFieldList = () => {
    const newErrors: Record<string, { name?: string; ref_field?: string }> = {};
    const nameSet = new Set<string>();

    fieldList.forEach((item) => {
      const itemErrors: { name?: string; ref_field?: string } = {};

      // 非空校验 - name
      if (!item.name || item.name.trim() === '') {
        itemErrors.name = intl.get('Global.cannotBeNull');
      }
      // 重复校验 - name
      else if (nameSet.has(item.name)) {
        itemErrors.name = intl.get('Global.nameCannotRepeat');
      } else {
        nameSet.add(item.name);
      }

      // 非空校验 - ref_field
      if (!item.ref_field || item.ref_field.trim() === '') {
        itemErrors.ref_field = intl.get('Global.cannotBeNull');
      }

      // 如果有错误，添加到 newErrors
      if (Object.keys(itemErrors).length > 0) {
        newErrors[item.id] = itemErrors;
      }
    });

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleCancel = () => {
    onCancel();
  };

  const handleOk = () => {
    if (mode === 'view') {
      onCancel();
      return;
    }
    if (validateFieldList()) {
      onOk(fieldList);
    }
  };

  const handlePrev = () => {
    if (!onPrev) return;
    if (mode === 'edit') {
      // 编辑模式需要校验并传递数据
      if (validateFieldList()) {
        onPrev(fieldList);
      }
    } else {
      // 查看模式直接切换
      onPrev(fieldList);
    }
  };

  const handleNext = () => {
    if (!onNext) return;
    if (mode === 'edit') {
      // 编辑模式需要校验并传递数据
      if (validateFieldList()) {
        onNext(fieldList);
      }
    } else {
      // 查看模式直接切换
      onNext(fieldList);
    }
  };

  // 获取状态文本
  const getStatusText = (isDefault: boolean) => {
    return isDefault ? intl.get('FieldFeatureModal.enabled') : intl.get('FieldFeatureModal.disabled');
  };

  const renderTitle = () => {
    if (fieldName) {
      const text = mode === 'view' ? `${intl.get('FieldFeatureModal.viewTitle')}${fieldName}` : `${intl.get('Global.setFieldFeature')}-${fieldName}`;
      return (
        <Tooltip title={text?.length > 20 ? text : undefined}>
          <div className={styles['modal-title-ellipsis']}>{text}</div>
        </Tooltip>
      );
    }
    return title || intl.get('FieldFeatureModal.title');
  };

  const footer = (
    <div key="navigation" className={styles['navigation-container']}>
      {mode === 'view' ? (
        <Button type="primary" onClick={handleOk}>
          {intl.get('FieldFeatureModal.close')}
        </Button>
      ) : (
        <div style={{ display: 'flex', gap: 8 }}>
          <Button type="primary" onClick={handleOk}>
            {intl.get('FieldFeatureModal.ok')}
          </Button>
          <Button onClick={handleCancel}>{intl.get('FieldFeatureModal.cancel')}</Button>
        </div>
      )}
    </div>
  );

  return (
    <Modal
      className={styles['field-feature-modal-container']}
      open={visible}
      title={renderTitle()}
      width={1200}
      centered
      maskClosable={false}
      closeIcon={<CloseOutlined />}
      onCancel={handleCancel}
      onOk={handleOk}
      okText={mode === 'view' ? intl.get('FieldFeatureModal.close') : intl.get('FieldFeatureModal.ok')}
      cancelText={mode === 'view' ? null : intl.get('FieldFeatureModal.cancel')}
      footer={footer}
    >
      {/* 国际化未加载完成时不渲染内容,避免显示空白或key值 */}
      {!i18nLoaded ? null : (
        <div className={styles['field-feature-modal']}>
          <div className={styles['table-header']}>
            <div className={styles['header-item']} style={{ width: 170 }}>
              {mode === 'edit' && <span className={styles['required']}>*</span>}
              {intl.get('FieldFeatureModal.featureName')}
            </div>
            <div className={styles['header-item']} style={{ width: 100 }}>
              {intl.get('FieldFeatureModal.type')}
            </div>
            <div className={styles['header-item']} style={{ width: 170 }}>
              {intl.get('FieldFeatureModal.viewField')}
            </div>
            <div className={styles['header-item']} style={{ width: 60 }}>
              {intl.get('FieldFeatureModal.status')}
            </div>
            <div className={styles['header-item']} style={{ width: 60 }}>
              {intl.get('FieldFeatureModal.isNative')}
            </div>
            <div className={styles['header-item']} style={{ width: 170 }}>
              {intl.get('FieldFeatureModal.remark')}
            </div>
            <div className={styles['header-item']} style={{ width: 170 }}>
              {intl.get('FieldFeatureModal.config')}
            </div>
            {mode === 'edit' && (
              <div className={styles['header-item']} style={{ width: 60 }}>
                {intl.get('FieldFeatureModal.delete')}
              </div>
            )}
          </div>

          <div className={styles['table-body']}>
            {fieldList.map((item) => (
              <div key={item.id} className={styles['table-row']}>
                {mode === 'edit' ? (
                  // 编辑模式:使用表单控件
                  <>
                    <div className={styles['row-item']} style={{ width: 170 }}>
                      <Tooltip title={errors[item.id]?.name || (item.name?.length > 20 ? item.name : undefined)}>
                        <Input
                          value={item.name}
                          onChange={(e) => handleFieldChange(item.id, 'name', e.target.value)}
                          placeholder={intl.get('FieldFeatureModal.placeholder')}
                          status={errors[item.id]?.name ? 'error' : ''}
                          maxLength={255}
                        />
                      </Tooltip>
                    </div>
                    <div className={styles['row-item']} style={{ width: 100 }}>
                      <Select
                        value={item.type}
                        onChange={(value) => handleFieldChange(item.id, 'type', value)}
                        options={typeOptions}
                        style={{ width: '100%' }}
                        disabled={item.is_native}
                      />
                    </div>
                    <div className={styles['row-item']} style={{ width: 170 }}>
                      <Tooltip title={errors[item.id]?.ref_field || (item.ref_field?.length > 20 ? item.ref_field : undefined)}>
                        <Select
                          value={item.ref_field || undefined}
                          onChange={(value) => handleFieldChange(item.id, 'ref_field', value)}
                          style={{ width: '100%' }}
                          options={getFilteredFieldsByType(item.type)}
                          status={errors[item.id]?.ref_field ? 'error' : ''}
                          disabled={item.is_native}
                          placeholder={intl.get('FieldFeatureModal.refFieldPlaceholder')}
                        />
                      </Tooltip>
                    </div>
                    <div className={styles['row-item']} style={{ width: 60 }}>
                      <Switch checked={item.is_default} onChange={(checked) => handleFieldChange(item.id, 'is_default', checked)} size="small" />
                    </div>
                    <div className={styles['row-item']} style={{ width: 60 }}>
                      <span className={styles['status-text']}>{item.is_native ? intl.get('FieldFeatureModal.yes') : intl.get('FieldFeatureModal.no')}</span>
                    </div>
                    <div className={styles['row-item']} style={{ width: 170 }}>
                      <Tooltip title={item.comment?.length > 20 ? item.comment : undefined}>
                        <Input
                          value={item.comment}
                          onChange={(e) => handleFieldChange(item.id, 'comment', e.target.value)}
                          placeholder={intl.get('FieldFeatureModal.placeholder')}
                          maxLength={1000}
                        />
                      </Tooltip>
                    </div>
                    <div className={styles['row-item']} style={{ width: 170 }}>
                      <Tooltip title={item.config ? JSON.stringify(item.config) : undefined}>
                        <span className={styles['text-value']}>{item.config ? JSON.stringify(item.config) : '—'}</span>
                      </Tooltip>
                    </div>
                    <div className={styles['row-item']} style={{ width: 60 }}>
                      <MinusCircleFilled
                        className={styles['delete-icon']}
                        style={{
                          cursor: item.is_native ? 'not-allowed' : 'pointer',
                          color: item.is_native ? 'rgba(0, 0, 0, 0.1)' : 'rgba(0, 0, 0, 0.25)',
                        }}
                        onClick={() => {
                          if (!item.is_native) {
                            handleDelete(item.id);
                          }
                        }}
                      />
                    </div>
                  </>
                ) : (
                  // 查看模式:纯文本展示
                  <>
                    <div className={styles['row-item']} style={{ width: 170 }}>
                      <Tooltip title={item.name?.length > 20 ? item.name : undefined}>
                        <span className={styles['text-value']}>{item.name || '—'}</span>
                      </Tooltip>
                    </div>
                    <div className={styles['row-item']} style={{ width: 100 }}>
                      <span className={styles['text-value']}>{item.type || '—'}</span>
                    </div>
                    <div className={styles['row-item']} style={{ width: 170 }}>
                      <Tooltip title={item.ref_field?.length > 20 ? item.ref_field : undefined}>
                        <span className={styles['text-value']}>{item.ref_field || '—'}</span>
                      </Tooltip>
                    </div>
                    <div className={styles['row-item']} style={{ width: 60 }}>
                      <span className={styles['text-value']}>{getStatusText(item.is_default)}</span>
                    </div>
                    <div className={styles['row-item']} style={{ width: 60 }}>
                      <span className={styles['text-value']}>{item.is_native ? intl.get('FieldFeatureModal.yes') : intl.get('FieldFeatureModal.no')}</span>
                    </div>
                    <div className={styles['row-item']} style={{ width: 170 }}>
                      <Tooltip title={item.comment?.length > 20 ? item.comment : undefined}>
                        <span className={styles['text-value']}>{item.comment || '—'}</span>
                      </Tooltip>
                    </div>
                    <div className={styles['row-item']} style={{ width: 170 }}>
                      <Tooltip title={item.config ? JSON.stringify(item.config) : undefined}>
                        <span className={styles['text-value']}>{item.config ? JSON.stringify(item.config) : '—'}</span>
                      </Tooltip>
                    </div>
                  </>
                )}
              </div>
            ))}
          </div>

          {mode === 'edit' && (
            <div className={styles['footer']}>
              <Tooltip title={fieldList.length >= 10 ? intl.get('FieldFeatureModal.maxLimit') : undefined}>
                <Button className={styles['add-button']} icon={<PlusOutlined />} type="text" onClick={handleAdd} disabled={fieldList.length >= 10}>
                  {intl.get('FieldFeatureModal.add')}
                </Button>
              </Tooltip>
            </div>
          )}
        </div>
      )}
    </Modal>
  );
};

export default FieldFeatureModal;
