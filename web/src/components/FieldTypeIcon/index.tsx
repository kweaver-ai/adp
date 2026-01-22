import { IconFont } from '@/web-library/common';
import { getIconByType } from './utils';

const FieldTypeIcon: React.FC<{ type: string; size?: number }> = ({ type, size = 20 }) => {
  if (!type) return null;
  return <IconFont type={getIconByType(type)} style={{ fontSize: size }} />;
};
export default FieldTypeIcon;
