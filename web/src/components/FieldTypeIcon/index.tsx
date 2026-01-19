import { IconFont } from '@/web-library/common';
import { getIconByType } from './utils';

const FieldTypeIcon: React.FC<{ type: string }> = ({ type }) => {
  if (!type) return null;
  return <IconFont type={getIconByType(type)} />;
};
export default FieldTypeIcon;
