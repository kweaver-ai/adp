import { useState, useEffect } from 'react';
import intl from 'react-intl-universal';
import { Modal, Transfer, TransferProps } from 'antd';
import FieldTypeIcon from '@/components/FieldTypeIcon';
import * as OntologyObjectType from '@/services/object/type';
import styles from './index.module.less';

interface Props {
  visible: boolean;
  onOk: (targetKeys: string[]) => void;
  onCancel: () => void;
  dataSource?: OntologyObjectType.Field[];
}

const PickAttribute: React.FC<Props> = ({ visible, onOk, onCancel, dataSource = [] }) => {
  const [targetKeys, setTargetKeys] = useState<React.Key[]>([]);
  const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([]);

  useEffect(() => {
    if (visible) {
      setTargetKeys([]);
      setSelectedKeys([]);
    }
  }, [visible]);

  const handleChange: TransferProps['onChange'] = (newTargetKeys) => {
    setTargetKeys(newTargetKeys);
  };

  const handleSelectChange: TransferProps['onSelectChange'] = (sourceSelectedKeys, targetSelectedKeys) => {
    setSelectedKeys([...sourceSelectedKeys, ...targetSelectedKeys]);
  };

  const handleSubmit = () => {
    onOk(targetKeys as string[]);
  };

  const handleClear = () => {
    setTargetKeys([]);
    setSelectedKeys([]);
  };

  return (
    <Modal
      title={intl.get('Object.addDataAttribute')}
      width={740}
      open={visible}
      onOk={handleSubmit}
      onCancel={onCancel}
      maskClosable={false}
      okText={intl.get('Global.ok')}
      cancelText={intl.get('Global.cancel')}
      okButtonProps={{ disabled: targetKeys.length === 0 }}
    >
      <div className={styles.modalContent}>
        <Transfer
          dataSource={dataSource.map((item) => ({ ...item, key: item.name }))}
          titles={[
            intl.get('Global.dataView'),
            <div key="clear" className={styles.clearBtn} onClick={handleClear}>
              {intl.get('Global.clear')}
            </div>,
          ]}
          targetKeys={targetKeys}
          selectedKeys={selectedKeys}
          onChange={handleChange}
          onSelectChange={handleSelectChange}
          render={(item) => (
            <div className={styles.transferItem}>
              <div className={styles.itemIcon}>
                <FieldTypeIcon type={item.type || 'string'} />
              </div>
              <div className={styles.itemContent}>
                <div className={styles.itemTitle}>{item.display_name}</div>
                <div className={styles.itemDesc}>{item.name}</div>
              </div>
            </div>
          )}
          oneWay
          listStyle={{
            width: 336,
            height: 416,
          }}
        />
      </div>
    </Modal>
  );
};

export default PickAttribute;
