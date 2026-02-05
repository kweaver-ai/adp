export const DataType_Configs = [
  { name: 'integer', type: 'number', icon: 'icon-dip-integer' },
  { name: 'unsigned integer', type: 'number', icon: 'icon-dip-integer' },
  { name: 'float', type: 'number', icon: 'icon-dip-float' },
  { name: 'decimal', type: 'number', icon: 'icon-dip-decimal' },
  { name: 'string', type: 'string', icon: 'icon-dip-string' },
  { name: 'text', type: 'string', icon: 'icon-dip-text' },
  { name: 'date', type: 'date', icon: 'icon-dip-date' },
  { name: 'time', type: 'date', icon: 'icon-dip-time' },
  { name: 'datetime', type: 'date', icon: 'icon-dip-datetime' },
  { name: 'timestamp', type: 'date', icon: 'icon-dip-timestamp' },
  { name: 'ip', type: 'string', icon: 'icon-dip-ip' },
  { name: 'boolean', type: 'boolean', icon: 'icon-dip-boolean' },
  { name: 'binary', type: 'string', icon: 'icon-dip-binary' },
  { name: 'json', type: 'string', icon: 'icon-dip-json' },
  { name: 'point', type: 'string', icon: 'icon-dip-point' },
  { name: 'shape', type: 'string', icon: 'icon-dip-shape' },
  { name: 'vector', type: 'string', icon: 'icon-dip-vector' },
] as const;

export type DataType = (typeof DataType_Configs)[number]['name'];
export type DataTypeCategory = (typeof DataType_Configs)[number]['type'];

export const DataType_All = DataType_Configs as unknown as {
  name: string;
  type: string;
  icon: string;
}[];

export const DataType_Date_Types = DataType_Configs.filter((item) => item.type === 'date').map((item) => item.name);
export const DataType_Number_Types = DataType_Configs.filter((item) => item.type === 'number').map((item) => item.name);
export const DataType_Boolean_Types = DataType_Configs.filter((item) => item.type === 'boolean').map((item) => item.name);

export const DataType_All_Name = DataType_Configs.map((item) => item.name);

// 缓存图标映射，提高查询效率
const DataType_Icon_Map: Record<string, string> = DataType_Configs.reduce(
  (acc, cur) => {
    acc[cur.name] = cur.icon;
    return acc;
  },
  {} as Record<string, string>
);

/**
 * 根据数据类型获取对应的图标
 * @param type 数据类型名称
 * @returns 图标类名
 */
export const getIconByType = (type: string) => DataType_Icon_Map[type] || 'icon-dip-unknown';

export default {
  DataType_All,
  DataType_Date_Types,
  DataType_Number_Types,
  DataType_Boolean_Types,
  DataType_All_Name,
  getIconByType,
};
