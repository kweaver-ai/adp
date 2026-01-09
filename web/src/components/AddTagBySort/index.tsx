import { useEffect, useState } from 'react';
import intl from 'react-intl-universal';
import { Dropdown } from 'antd';
import classNames from 'classnames';
import { cloneDeep, keyBy, map, filter, includes } from 'lodash-es';
import { Button, IconFont, Input } from '@/web-library/common';
import UTILS from '@/web-library/utils';
import styles from './index.module.less';
import locales from './locales';

const _ = { cloneDeep, keyBy, map, filter, includes };

const TagItem = (props: any) => {
  const { data, onDelete, onSort } = props;
  const { type, displayName, display_name, comment } = data || {};
  const icon = type ? UTILS.formatIconByType(type) : '';

  return (
    <div className={styles['tag-item']}>
      <div className={styles['tag-item-name-box']} title={comment}>
        <IconFont className="g-mr-1" type={icon} />
        <span className={classNames('g-ellipsis-1')}>{displayName || display_name}</span>
      </div>
      <div className={styles['tag-item-operate-box']}>
        <Button.Icon
          size="small"
          onClick={() => onSort('desc')}
          title={intl.get('AddTagBySort.desc')}
          icon={<IconFont type="icon-dip-sort-descending" style={{ fontSize: 16, color: data?.direction === 'desc' ? '#126EE3' : 'rgba(0, 0, 0, 0.65)' }} />}
        />
        <Button.Icon
          size="small"
          onClick={() => onSort('asc')}
          title={intl.get('AddTagBySort.asc')}
          icon={<IconFont type="icon-dip-shengxupaixu" style={{ fontSize: 16, color: data?.direction === 'asc' ? '#126EE3' : 'rgba(0, 0, 0, 0.65)' }} />}
        />
        <div className={styles['tag-item-operate-line']}></div>
        {!!onDelete && (
          <Button.Icon size="small" icon={<IconFont type="icon-dip-trash" style={{ fontSize: 16, color: 'rgba(0, 0, 0, 0.65)' }} />} onClick={onDelete} />
        )}
      </div>
    </div>
  );
};

interface AddTagBySortProps {
  value?: any[];
  onChange?: (value: any[]) => void;
  options: any[];
  onAddBtnClick?: () => void;
}

const AddTagBySort = (props: AddTagBySortProps) => {
  const { value = [], onChange, options, onAddBtnClick } = props;

  const [filterOptions, setFilterOptions] = useState(_.cloneDeep(options));
  const [i18nLoaded, setI18nLoaded] = useState(false);

  useEffect(() => {
    // 加载国际化文件，完成后更新状态触发重新渲染
    intl.load(locales);
    setI18nLoaded(true);
  }, []);

  useEffect(() => {
    setFilterOptions(options);
  }, [options]);

  const optionsKV = _.keyBy(_.cloneDeep(options), 'name');
  const valueString = _.map(value, (item) => item?.name || item);

  /** 筛选待选项 */
  const onSearchChange = (data: any) => {
    const value = data.target.value;
    const newFiletFields = _.filter(_.cloneDeep(options), (item) => _.includes(item.name, value));
    setFilterOptions(newFiletFields);
  };

  /** 添加标签 */
  const onAdd = (data: any) => {
    const newValue = _.cloneDeep(value) || [];
    newValue.push({ name: data.name, type: data.type, display_name: data.displayName, direction: 'desc' });
    onChange?.(newValue);
  };

  /** 删除标签 */
  const onDelete = (data: any) => {
    let newValue = [];
    newValue = _.filter(_.cloneDeep(value), (item: any) => item.name !== data.name);
    onChange?.(newValue);
  };

  const onSort = (index: number, sort: 'asc' | 'desc') => {
    const newValue = value.map((item: any, i: number) => {
      if (i === index) {
        item.direction = sort;
      }
      return item;
    });

    onChange?.(newValue);
  };

  return (
    <div className={styles['vega-form-tag-root']}>
      <div className={styles['vega-form-tag-sort-container']}>
        {_.map(value, (item, index) => {
          const data = { ...item, ...optionsKV[item.name] };
          return <TagItem key={index} data={data} onDelete={() => onDelete(item)} onSort={(sort: 'asc' | 'desc') => onSort(index, sort)} />;
        })}
      </div>

      <Dropdown
        trigger={['click']}
        destroyOnHidden
        getPopupContainer={(triggerNode): HTMLElement => triggerNode.parentNode as HTMLElement}
        popupRender={() => {
          // 国际化未加载完成时不渲染 Dropdown 内容，避免显示空白或key值
          if (!i18nLoaded) {
            return null;
          }
          return (
            <div className="g-dropdown-menu-root" style={{ width: 260 }}>
              <Input.Search className="g-mt-2 g-mb-4" allowClear placeholder={intl.get('AddTagBySort.search')} onChange={onSearchChange} />
              <div className={styles['vega-form-tag-options-container']}>
                {_.map(filterOptions, (item, index) => {
                  const icon = UTILS.formatIconByType(item?.type);
                  const selected = _.includes(valueString, item.name);
                  return (
                    <div
                      key={index}
                      className={classNames(styles['vega-form-tag-options'], { [styles['vega-form-tag-options-disabled']]: selected })}
                      onClick={() => {
                        if (!selected) onAdd(item);
                      }}
                      title={item.comment}
                    >
                      <div style={{ display: 'flex', alignItems: 'flex-start' }}>
                        <IconFont className="g-mr-1 g-mt-1" type={icon} style={{ fontSize: 18 }} />
                        <div>
                          <div className="g-ellipsis-1">{item?.displayName}</div>
                          <div className="g-ellipsis-1 g-c-text-sub">{item?.name}</div>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          );
        }}
      >
        <Button className="g-mr-2 g-mb-2" icon={<IconFont type="icon-add" />} onClick={() => onAddBtnClick?.()}>
          {intl.get('AddTagBySort.add')}
        </Button>
      </Dropdown>
    </div>
  );
};

export default AddTagBySort;
