import { Select } from 'antd';
import FieldTypeIcon from '@/components/FieldTypeIcon';
import { IconFont, Tooltip } from '@/web-library/common';
import styles from './index.module.less';
import type { SelectProps } from 'antd';

interface FieldOption {
  name: string;
  display_name: string;
  type: string;
  [key: string]: any;
}

interface FieldSelectProps extends Omit<SelectProps, 'options'> {
  fields: FieldOption[];
  tooltipLengthThreshold?: number;
  getOptionDisabled?: (field: FieldOption) => boolean;
}

const FieldSelect = ({ fields, tooltipLengthThreshold = 20, getOptionDisabled, getPopupContainer, filterOption, ...restProps }: FieldSelectProps) => {
  const defaultFilterOption = (input: string, option?: any) => {
    const field = fields?.find((f) => f.name === option?.value);
    const inputLower = input.toLowerCase();
    return (field?.display_name ?? '').toLowerCase().includes(inputLower) || (field?.name ?? '').toLowerCase().includes(inputLower);
  };

  return (
    <Select
      showSearch
      {...restProps}
      filterOption={filterOption !== undefined ? filterOption : defaultFilterOption}
      getPopupContainer={getPopupContainer || ((triggerNode): HTMLElement => triggerNode.parentNode)}
      labelRender={(option) => {
        if (!option || !option.value) {
          return null;
        }
        const field = fields?.find((f: any) => f.name === option.value);
        const displayName = field?.display_name || '';
        return (
          <div className={styles.selectItemSingle}>
            <div className={styles.itemIcon}>
              <FieldTypeIcon type={field?.type || ''} />
            </div>
            <Tooltip title={displayName.length > tooltipLengthThreshold ? displayName : undefined}>
              <span className={styles.itemTitle}>{displayName}</span>
            </Tooltip>
            {field?.comment && (
              <Tooltip title={field.comment}>
                <IconFont type="icon-dip-color-comment" style={{ marginLeft: 4 }} />
              </Tooltip>
            )}
          </div>
        );
      }}
      options={
        fields?.map((item: any) => {
          const isDisabled = getOptionDisabled ? getOptionDisabled(item) : false;
          return {
            label: (
              <div className={`${styles.selectItem} ${isDisabled ? styles.disabled : ''}`}>
                <div className={styles.itemIcon}>
                  <FieldTypeIcon type={item.type} />
                </div>
                <div className={styles.itemContent}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                    <Tooltip title={item.display_name?.length > tooltipLengthThreshold ? item.display_name : undefined}>
                      <div className={styles.itemTitle}>{item.display_name}</div>
                    </Tooltip>
                    {item.comment && (
                      <Tooltip title={item.comment}>
                        <IconFont type="icon-dip-color-comment" />
                      </Tooltip>
                    )}
                  </div>
                  <Tooltip title={item.name?.length > tooltipLengthThreshold ? item.name : undefined}>
                    <div className={styles.itemDesc}>{item.name}</div>
                  </Tooltip>
                </div>
              </div>
            ),
            value: item.name,
            disabled: isDisabled,
          };
        }) || []
      }
    />
  );
};

export default FieldSelect;
