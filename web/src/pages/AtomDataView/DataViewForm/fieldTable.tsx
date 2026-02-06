import { useState } from 'react';
import intl from 'react-intl-universal';
import { Table, Input, Button } from 'antd';
import classNames from 'classnames';
import * as AtomDataViewType from '@/services/atomDataView/type';
import HOOKS from '@/hooks';
import { IconFont, Tooltip } from '@/web-library/common';
import styles from './index.module.less';

interface FieldError {
  [key: string]: {
    display_name?: string;
    name?: string;
  };
}

interface TProps {
  dataSource?: AtomDataViewType.Field[];
  page: number;
  setPage: (val: number) => void;
  isEdit: boolean;
  onChange: (val: AtomDataViewType.Field[]) => void;
  onFieldFeatureClick?: (field: AtomDataViewType.Field) => void;
}

const EditableTable = ({ isEdit, page, setPage, dataSource = [], onChange, onFieldFeatureClick }: TProps): JSX.Element => {
  const { message } = HOOKS.useGlobalContext();
  const { EXCEL_DATA_TYPES } = HOOKS.useConstants();
  const [fieldErrors, setFieldErrors] = useState<FieldError>({});

  // 清除字段错误
  const clearFieldError = (originalName: string, field: 'display_name' | 'name') => {
    setFieldErrors((prev) => {
      const newErrors = { ...prev };
      if (newErrors[originalName]) {
        delete newErrors[originalName][field];
        if (Object.keys(newErrors[originalName]).length === 0) {
          delete newErrors[originalName];
        }
      }
      return newErrors;
    });
  };

  // 设置字段错误
  const setFieldError = (originalName: string, field: 'display_name' | 'name', errorMsg: string) => {
    setFieldErrors((prev) => {
      const newErrors = { ...prev };
      if (!newErrors[originalName]) {
        newErrors[originalName] = {};
      }
      newErrors[originalName][field] = errorMsg;
      return newErrors;
    });
  };

  // 处理字段值变化
  const handleFieldChange = (record: AtomDataViewType.Field, field: string, value: any): void => {
    // 清除该字段的错误状态
    if (field === 'display_name' || field === 'name') {
      clearFieldError(record.original_name, field);
    }

    const newData = dataSource.map((item) => {
      if (item.original_name === record.original_name) {
        return { ...item, [field]: value };
      }
      return item;
    });
    onChange(newData);
  };

  // 校验字段显示名称
  const validateDisplayName = (record: AtomDataViewType.Field, value: string): string => {
    if (!value || !value.trim()) {
      return intl.get('Global.cannotBeNull');
    }
    if (value.length > 255) {
      return intl.get('DataView.displayNameMaxLength');
    }
    const isDuplicate = dataSource.some((item) => item.original_name !== record.original_name && item.display_name === value);
    if (isDuplicate) {
      return intl.get('Global.nameCannotRepeat');
    }
    return '';
  };

  // 校验字段名称
  const validateName = (record: AtomDataViewType.Field, value: string): string => {
    if (!value || !value.trim()) {
      return intl.get('Global.cannotBeNull');
    }
    if (value.length > 100) {
      return intl.get('DataView.fieldNameMaxLength');
    }
    if (!/^[a-z_][a-z0-9_]{0,}$/.test(value)) {
      return intl.get('Global.idSpecialVerification');
    }
    const isDuplicate = dataSource.some((item) => item.original_name !== record.original_name && item.name === value);
    if (isDuplicate) {
      return intl.get('Global.fieldNameCannotRepeat');
    }
    return '';
  };

  // 校验字段
  const validateField = (record: AtomDataViewType.Field, field: string, value: any): void => {
    let errorMsg = '';

    if (field === 'display_name') {
      errorMsg = validateDisplayName(record, value);
    } else if (field === 'name') {
      errorMsg = validateName(record, value);
    }

    // 更新错误状态
    if (errorMsg) {
      setFieldError(record.original_name, field as 'display_name' | 'name', errorMsg);
    } else {
      clearFieldError(record.original_name, field as 'display_name' | 'name');
    }
  };

  // 表格列定义
  const columns: any = [
    {
      title: intl.get('Global.fieldDisplayName'),
      dataIndex: 'display_name',
      key: 'display_name',
      width: 150,
      render: (text: string, record: AtomDataViewType.Field) => {
        const error = fieldErrors[record.original_name]?.display_name;
        return (
          <Tooltip title={error || text} getPopupContainer={(): HTMLElement => document.body}>
            <Input
              value={text}
              status={error ? 'error' : ''}
              onChange={(e): void => handleFieldChange(record, 'display_name', e.target.value)}
              onBlur={(e): void => validateField(record, 'display_name', e.target.value)}
              placeholder={intl.get('Global.pleaseInput')}
            />
          </Tooltip>
        );
      },
    },
    {
      title: intl.get('Global.fieldName'),
      dataIndex: 'name',
      key: 'name',
      width: 150,
    },
    {
      title: intl.get('Global.fieldType'),
      dataIndex: 'type',
      key: 'type',
      width: 100,
    },
    {
      title: intl.get('Global.fieldComment'),
      dataIndex: 'comment',
      key: 'comment',
      width: 200,
      render: (text: string, record: AtomDataViewType.Field): JSX.Element => (
        <Tooltip placement="left" title={text}>
          <Input value={text} onChange={(e): void => handleFieldChange(record, 'comment', e.target.value)} placeholder={intl.get('Global.pleaseInput')} />
        </Tooltip>
      ),
    },
    {
      title: intl.get('Global.fieldFeatureType'),
      dataIndex: 'features',
      key: 'features_type',
      width: 150,
      render: (features: any[]) => {
        if (!features || features.length === 0) {
          return <span style={{ color: 'rgba(0, 0, 0, 0.25)' }}>{intl.get('Global.unset')}</span>;
        }
        const uniqueTypes = Array.from(new Set(features.map((item) => item.type)));
        return (
          <div className={styles.featureTypeContainer}>
            {uniqueTypes.map((type) => (
              <span key={type} className={classNames(styles.featureType, styles[type])}>
                {type}
              </span>
            ))}
          </div>
        );
      },
    },
    {
      title: () => (
        <div>
          <span style={{ marginRight: 8 }}>{intl.get('Global.fieldFeature')}</span>
          <Tooltip title={intl.get('Global.fieldFeatureTip')}>
            <IconFont type="icon-dip-color-tip" className={styles.helpIcon} />
          </Tooltip>
        </div>
      ),
      dataIndex: 'features',
      key: 'features',
      width: 100,
      render: (_: unknown, record: AtomDataViewType.Field) => (
        <Button
          type="link"
          onClick={(): void => {
            if (onFieldFeatureClick) {
              onFieldFeatureClick(record);
            }
          }}
        >
          {intl.get('Global.setting')}
        </Button>
      ),
    },
  ];

  return (
    <Table
      rowKey="original_name"
      size="small"
      columns={columns}
      dataSource={dataSource}
      className={styles['dict-box']}
      scroll={{ y: 360 }}
      onChange={(pagination): void => setPage(pagination.current ?? 1)}
      pagination={
        dataSource.length <= 10
          ? false
          : {
              current: page,
              total: dataSource.length,
              size: 'small',
            }
      }
    />
  );
};

export default EditableTable;
