import { CSSProperties } from 'react';
import { IconFont } from '@/web-library/common';

export interface ObjectIconProps {
  /** 图标类型 */
  icon?: string;
  /** 背景颜色 */
  color?: string;
  /** 容器大小（宽高相同），默认 24px */
  size?: number;
  /** 图标字体大小，默认为 16px */
  iconSize?: number;
  /** 图标样式 */
  iconStyle?: CSSProperties;
  /** 容器样式 */
  style?: CSSProperties;
  /** 容器类名 */
  className?: string;
  /** 边框圆角，默认 0 */
  borderRadius?: number;
}

const ObjectIcon = (props: ObjectIconProps) => {
  const { icon = 'icon-color-rectangle', color = '#1890ff', size = 24, iconSize = 16, iconStyle, style, className, borderRadius = 4 } = props;

  // 图标大小默认为容器大小的 0.7 倍
  const finalIconSize = iconSize;

  const containerStyle: CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    width: size,
    height: size,
    background: color,
    borderRadius,
    ...style,
  };

  const defaultIconStyle: CSSProperties = {
    color: '#fff',
    fontSize: finalIconSize,
    ...iconStyle,
  };

  return (
    <div className={className} style={containerStyle}>
      <IconFont type={icon} style={defaultIconStyle} />
    </div>
  );
};

export default ObjectIcon;
