import intl from 'react-intl-universal';
import { LeftOutlined } from '@ant-design/icons';
import { Divider } from 'antd';
import { Text, Title, Steps } from '@/web-library/common';
import styles from './index.module.less';

const Header = (props: any) => {
  const { title, stepsCurrent, goBack, onPrev, onNext, actions } = props;

  return (
    <div className={styles['header-root']}>
      <div className={styles['go-back']}>
        <div className="g-pointer g-flex-align-center" onClick={goBack}>
          <LeftOutlined style={{ marginTop: 2, marginRight: 6 }} />
          <Text>{intl.get('Global.exit')}</Text>
        </div>
        <Divider type="vertical" style={{ margin: '0 12px' }} />
        <Title>{title}</Title>
      </div>
      <div style={{ maxWidth: 800 }}>
        <Steps.GapIcon
          size="small"
          current={stepsCurrent}
          items={[
            { title: intl.get('Action.conceptDefinition') },
            { title: intl.get('Action.resourceMapping') }
            // { title: intl.get('Action.runStrategy') || '运行策略' },
          ]}
          onChange={(value) => {
            // Disable click navigation for now as validation is tricky
            // Or implement it if needed. The original code supported it.
            // But with validation logic in parent, direct jump might bypass validation.
            // Original code:
            // if (value === 0) {
            //   onPrev();
            // } else {
            //   onNext();
            // }
            // If we allow jumping back, we can support it. Jumping forward needs validation.
            // Let's keep it simple and just show the steps, maybe disable onChange or handle it carefully.
            // The original code only supported 2 steps (0 and 1).
            // If I am at 2, clicking 0 or 1 is fine (prev).
            // If I am at 0, clicking 1 or 2 is next (needs validation).
            
            // For safety, let's disable direct click navigation for now unless requested, 
            // or just allow going back.
          }}
        />
      </div>
      <div className={styles['actions']}>{actions}</div>
    </div>
  );
};

export default Header;
