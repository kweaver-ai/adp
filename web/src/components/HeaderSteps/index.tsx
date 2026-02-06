import { useEffect, useState } from 'react';
import intl from 'react-intl-universal';
import { useHistory } from 'react-router-dom';
import { LeftOutlined } from '@ant-design/icons';
import { Divider } from 'antd';
import HOOKS from '@/hooks';
import { Text, Title, Steps, Button } from '@/web-library/common';
import styles from './index.module.less';
import locales from './locales';

interface ActionButton {
  text: string;
  onClick: () => void;
  type?: 'default' | 'primary';
  loading?: boolean;
  disabled?: boolean;
}

interface TProps {
  title?: string;
  stepsCurrent: number;
  goBack?: () => void;
  items?: { title: string }[];
  onStepChange?: (current: number) => void;
  actions?: {
    prev?: ActionButton;
    next?: ActionButton;
    save?: ActionButton;
  };
}

const HeaderSteps = (props: TProps) => {
  const { title = '', stepsCurrent, goBack: goBackProp, items, onStepChange, actions } = props;
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
      {actions && (
        <div className={styles['header-actions']}>
          {actions.prev && (
            <Button onClick={actions.prev.onClick} loading={actions.prev.loading} disabled={actions.prev.disabled}>
              {actions.prev.text}
            </Button>
          )}
          {actions.save && (
            <Button className="g-ml-2" type="primary" loading={actions.save.loading} disabled={actions.save.disabled} onClick={actions.save.onClick}>
              {actions.save.text}
            </Button>
          )}
          {actions.next && (
            <Button className="g-ml-2" type="primary" loading={actions.next.loading} disabled={actions.next.disabled} onClick={actions.next.onClick}>
              {actions.next.text}
            </Button>
          )}
        </div>
      )}
    </div>
  );
};

export default HeaderSteps;
export type { ActionButton };
