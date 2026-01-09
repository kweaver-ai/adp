/** 时间过滤器组件，用于全局筛选过滤数据 */
import { useEffect, useState } from 'react';
import intl from 'react-intl-universal';
import { Tabs, Dropdown, TabsProps } from 'antd';
import dayjs from 'dayjs';
import DateRange from './DateRange';
import locales from './locales';
import QuickTags from './QuickTags';

const TimeFilter = (props: any) => {
  const { placement, timeRange = { label: '最近24小时', value: [dayjs().subtract(24, 'h'), dayjs()] }, onFilterChange, children } = props;
  const [i18nLoaded, setI18nLoaded] = useState(false);

  useEffect(() => {
    // 加载国际化文件，完成后更新状态触发重新渲染
    intl.load(locales);
    setI18nLoaded(true);
  }, []);

  const items: TabsProps['items'] = i18nLoaded
    ? [
        { key: 'quickSelect', label: intl.get('TimeFilter.quickSelect'), children: <QuickTags timeRange={timeRange} onFilterChange={onFilterChange} /> },
        { key: 'dateRange', label: intl.get('TimeFilter.dateRange'), children: <DateRange timeRange={timeRange} onFilterChange={onFilterChange} /> },
      ]
    : [];

  const isQuickSelect = timeRange.label.indexOf('-') < 0;

  return (
    <Dropdown
      trigger={['click']}
      destroyOnHidden
      placement={placement || 'bottomLeft'}
      popupRender={() => (
        <div className="g-dropdown-menu-root" style={{ width: 450 }}>
          <Tabs size="small" defaultActiveKey={isQuickSelect ? 'quickSelect' : 'dateRange'} items={items}></Tabs>
        </div>
      )}
    >
      {children}
    </Dropdown>
  );
};

export default TimeFilter;
