import { Tooltip as AntdTooltip, TooltipProps } from 'antd';

interface CustomTooltipProps extends Omit<TooltipProps, 'styles'> {
  maxHeight?: number;
}

const Tooltip = ({ maxHeight = 400, getPopupContainer, ...restProps }: CustomTooltipProps): JSX.Element => {
  return (
    <AntdTooltip
      {...restProps}
      getPopupContainer={getPopupContainer || (() => document.body)}
      styles={{
        body: {
          maxHeight,
          overflow: 'auto',
        },
      }}
    />
  );
};

export default Tooltip;
