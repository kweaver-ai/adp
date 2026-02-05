import { useEffect, useMemo, useState } from 'react';
import intl from 'react-intl-universal';
import { Input, Table, Button } from 'antd';
import classNames from 'classnames';
import { PAGINATION_DEFAULT } from '@/hooks/useConstants';
import * as AtomDataViewType from '@/services/atomDataView/type';
import { IconFont, Tooltip } from '@/web-library/common';
import styles from './index.module.less';

interface TProps {
  data?: AtomDataViewType.Field[];
  onFieldFeatureClick?: (record: AtomDataViewType.Field) => void; // 查看字段特征回调
}

const FieldTable = (props: TProps): JSX.Element => {
  const { data = [], onFieldFeatureClick } = props;
  const [searchValue, setSearchValue] = useState<string>();
  const [pagination, setPagination] = useState<{
    current?: number;
    pageSize?: number;
  }>({
    current: 1,
    pageSize: 10,
  });

  useEffect(() => {
    setSearchValue(undefined);
    setPagination({
      current: PAGINATION_DEFAULT.current,
      pageSize: PAGINATION_DEFAULT.pageSize,
    });
  }, [JSON.stringify(data)]);

  const filterData = useMemo(() => {
    if (searchValue) {
      return data.filter((item) => item.display_name.includes(searchValue) || item.name.includes(searchValue));
    }
    return data;
  }, [data, searchValue]);

  return (
    <>
      <Input.Search
        className={styles['table-field-search']}
        value={searchValue}
        onChange={(value): void => setSearchValue(value.target.value)}
        placeholder={intl.get('Global.search')}
        allowClear
      ></Input.Search>
      <Table
        rowKey="name"
        dataSource={filterData}
        size="small"
        columns={[
          {
            title: intl.get('Global.fieldDisplayName'),
            dataIndex: 'display_name',
            render: (text: string) =>
              text ? (
                <Tooltip title={text}>
                  <span className={styles.commentText}>{text}</span>
                </Tooltip>
              ) : (
                '--'
              ),
          },
          {
            title: intl.get('Global.fieldName'),
            dataIndex: 'name',
            width: 200,
          },
          {
            title: intl.get('Global.fieldType'),
            dataIndex: 'type',
            width: 120,
          },
          {
            title: intl.get('Global.fieldComment'),
            dataIndex: 'comment',
            render: (text: string) =>
              text ? (
                <Tooltip title={text}>
                  <span className={styles.commentText}>{text}</span>
                </Tooltip>
              ) : (
                '--'
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
            align: 'center',
            render: (_: unknown, record: AtomDataViewType.Field) => (
              <Button
                type="link"
                onClick={(): void => {
                  if (onFieldFeatureClick) {
                    onFieldFeatureClick(record);
                  }
                }}
                disabled={!record.features || record.features.length === 0}
              >
                {intl.get('Global.view')}
              </Button>
            ),
          },
        ]}
        onChange={(val): void => setPagination(val)}
        scroll={{
          y: 390,
        }}
        pagination={{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total: filterData.length,
          showTotal: (total) => intl.get('Global.total', { total }),
          pageSizeOptions: [10, 20, 50],
          showSizeChanger: true,
          showQuickJumper: true,
        }}
      />
    </>
  );
};

export default FieldTable;
