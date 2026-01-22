import { useEffect, useState } from 'react';
import intl from 'react-intl-universal';
import { useHistory } from 'react-router-dom';
import { LeftOutlined } from '@ant-design/icons';
import { Divider } from 'antd';
import HOOKS from '@/hooks';
import { Text, Title, Steps } from '@/web-library/common';
import styles from './index.module.less';
import locales from './locales';

interface TProps {
  title?: string;
  stepsCurrent: number;
  goBack?: () => void;
  items?: { title: string }[];
  onStepChange?: (current: number) => void;
}

const HeaderSteps = (props: TProps) => {
  const { title = '', stepsCurrent, goBack: goBackProp, items, onStepChange } = props;
  const history = useHistory();
  const { modal } = HOOKS.useGlobalContext();
  const [i18nLoaded, setI18nLoaded] = useState(false);

  useEffect(() => {
    // 加载国际化文件，完成后更新状态触发重新渲染
    intl.load(locales);
    setI18nLoaded(true);
  }, []);

  const goBack = () => {
    modal.confirm({
      title: intl.get('HeaderSteps.confirmBackTitle'),
      content: intl.get('HeaderSteps.confirmBackContent'),
      onOk: () => {
        goBackProp ? goBackProp() : history.goBack();
      },
    });
  };

  // 国际化未加载完成时不渲染
  if (!i18nLoaded) {
    return null;
  }

  return (
    <div className={styles['header-step-box']}>
      <div className={styles['box-exit']}>
        <div className={styles['back-box']} onClick={goBack}>
          <LeftOutlined />
          <Text>{intl.get('HeaderSteps.exit')}</Text>
        </div>
        <Divider type="vertical" />
        <Title>{title}</Title>
      </div>
      {items && <Steps.GapIcon size="small" current={stepsCurrent} items={items} onChange={onStepChange} />}
    </div>
  );
};

export default HeaderSteps;
