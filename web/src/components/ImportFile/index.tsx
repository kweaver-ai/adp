import React, { useEffect, useState } from 'react';
import intl from 'react-intl-universal';
import { Upload, Tooltip, Modal } from 'antd';
import { arNotification } from '@/components/ARNotification';
import { IconFont } from '@/web-library/common';
import locales from './locales';

enum File {
  Json = 'json',
}

const ImportFile = ({
  children,
  accept = File.Json,
  customRequest,
  getData,
  confirm,
  ...props
}: {
  children?: React.ReactElement;
  accept?: string;
  getData: () => Promise<void>;
  confirm?: {
    checkedRenameRequest: (data: any) => Promise<string | boolean>;
  };
  customRequest: (param: any, name: any) => Promise<unknown | void>;
  [key: string]: any;
}): JSX.Element => {
  const [i18nLoaded, setI18nLoaded] = useState(false);

  useEffect(() => {
    // 加载国际化文件，完成后更新状态触发重新渲染
    intl.load(locales);
    setI18nLoaded(true);
  }, []);

  const uploadFile = (e: any): void => {
    const strArr = e.file.name.split('.');
    const name = strArr[0];
    const fileType = strArr[strArr.length - 1];

    if (fileType !== accept) {
      arNotification.error(intl.get('ImportFile.fileAccept', { accept }));
      return;
    }

    const reader = new FileReader();

    reader.readAsText(e.file);

    const resolve = (): void => {
      getData();
      arNotification.success(intl.get('ImportFile.importSuccess'));
    };

    const reject = () => {};

    reader.onload = (e): void => {
      const { result = '' } = e.target || {};

      const importData = async (): Promise<void> => {
        try {
          if (accept === File.Json) {
            if (confirm) {
              const error = await confirm.checkedRenameRequest(JSON.parse(result as string));

              if (error) {
                Modal.confirm({
                  getContainer: () => document.getElementById('vega-root') as HTMLElement,
                  onOk: () => customRequest(JSON.parse(result as string), name).then(resolve, reject),
                  title: error,
                });

                return;
              }
            }
            customRequest(JSON.parse(result as string), name).then(resolve, reject);
          } else {
            customRequest(e.target?.result, name).then(resolve, reject);
          }
        } catch (e) {
          arNotification.error(intl.get('ImportFile.fileError'));
        }
      };

      importData();
    };
  };

  return (
    <Upload accept={accept} showUploadList={false} customRequest={uploadFile} {...props}>
      {children || (
        <Tooltip title={intl.get('ImportFile.import')}>
          <IconFont type="icon-upload"></IconFont>
        </Tooltip>
      )}
    </Upload>
  );
};

export default ImportFile;
