/** 快速选择标签组件 */
import { useEffect, useState } from 'react';
import intl from 'react-intl-universal';
import { Row, Col } from 'antd';
import classNames from 'classnames';
import { map } from 'lodash-es';
import styles from './index.module.less';
import quickRange from './quickRange';
import locales from '../locales';

const QuickTags = (props: any) => {
  const { timeRange, onFilterChange } = props;
  const lastindex = quickRange.length - 1;
  const [i18nLoaded, setI18nLoaded] = useState(false);

  useEffect(() => {
    // 加载国际化文件，完成后更新状态触发重新渲染
    intl.load(locales);
    setI18nLoaded(true);
  }, []);

  // 国际化未加载完成时不渲染内容，避免显示空白或key值
  if (!i18nLoaded) {
    return null;
  }

  return (
    <Row>
      {map(quickRange, ({ section, span, list }, index) => {
        const isLast = index === lastindex;
        return (
          <Col className={styles['quick-tags-col']} key={section} span={span}>
            <ul className={classNames(styles['quick-tags-ul'], { [styles['quick-tags-ul-last']]: isLast })}>
              {list.map((item) => {
                const isActive = timeRange.label === item.label;
                return (
                  <li
                    key={item.label}
                    className={classNames(styles['quick-tags-li'], { [styles['quick-tags-li-active']]: isActive })}
                    onClick={() => onFilterChange(item)}
                  >
                    {intl.get(`TimeFilter.quickRangeTime.${item.label}`)}
                  </li>
                );
              })}
            </ul>
          </Col>
        );
      })}
    </Row>
  );
};

export default QuickTags;
