import React from 'react';
import intl from 'react-intl-universal';
import { ExclamationCircleFilled } from '@ant-design/icons';
import type { ModalFuncProps } from 'antd';

interface DeleteConfirmOptions extends Partial<ModalFuncProps> {
  name?: string;
  content?: React.ReactNode;
  onOk: () => void;
}

export const showDeleteConfirm = (modal: any, options: DeleteConfirmOptions): void => {
  const { name, content, onOk, title, icon, okText, okButtonProps, footer, ...restOptions } = options;

  modal.confirm({
    title: title || intl.get('Global.tipTitle'),
    icon: icon !== undefined ? icon : <ExclamationCircleFilled />,
    content: content || intl.get('Global.deleteConfirm', { name }),
    okText: okText || intl.get('Global.delete'),
    okButtonProps: okButtonProps || {
      style: { backgroundColor: '#ff4d4f', borderColor: '#ff4d4f' },
    },
    footer:
      footer !== undefined
        ? footer
        : (_: any, { OkBtn, CancelBtn }: { OkBtn: React.ElementType; CancelBtn: React.ElementType }) => (
            <>
              <OkBtn />
              <CancelBtn />
            </>
          ),
    onOk,
    ...restOptions,
  } as ModalFuncProps);
};
