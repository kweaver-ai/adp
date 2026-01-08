/** 导出文件 */
import React, { memo, useEffect } from 'react';
import intl from 'react-intl-universal';
import { message, Tooltip } from 'antd';
import locales from './locales';
import { IconFont } from '../../common';
import downFile from '../../utils/down-file';

const ExportFile = memo(
  ({ customRequest, children, name }: { name: string; customRequest: () => Promise<any>; children?: React.ReactElement; [key: string]: any }) => {
    useEffect(() => {
      intl.load(locales);
    }, []);

    const exportData = async (e: any): Promise<void> => {
      e.stopPropagation();

      customRequest().then((res) => {
        if (res.code || res.error_code) return;
        downFile(JSON.stringify(res, null, 2), name, 'json');
        message.success(intl.get('ExportFile.exportSuccess'));
      });
    };

    return (
      <span onClick={exportData}>
        {children || (
          <Tooltip title={intl.get('ExportFile.export')}>
            <IconFont type="icon-download"></IconFont>
          </Tooltip>
        )}
      </span>
    );
  }
);

export default ExportFile;
