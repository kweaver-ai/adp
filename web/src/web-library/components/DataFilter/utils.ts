import dayjs from 'dayjs';
import { DataFilterValue, FieldList } from './type';
import Fields from '../../utils/fields';

// 类型操作符映射
export const typeOperationMapping = {
  // 全局搜索 (适用于所有字段或特殊搜索)
  'all Fields': ['match', 'match_phrase', 'multi_match'],

  // 字符串类型 (string)
  string: [
    '==',
    '!=',
    'in',
    'not_in',
    'like',
    'not_like',
    'prefix',
    'not_prefix',
    'regex',
    'contain',
    'not_contain',
    'empty',
    'not_empty',
    'exist',
    'not_exist',
    'null',
    'not_null',
    'knn',
  ],

  // 文本类型 (text)
  text: [
    '==',
    '!=',
    'like',
    'not_like',
    'prefix',
    'not_prefix',
    'regex',
    'contain',
    'not_contain',
    'empty',
    'not_empty',
    'match',
    'match_phrase',
    'multi_match',
    'exist',
    'not_exist',
    'null',
    'not_null',
    'knn',
  ],

  // 数字大类 (包含 integer, unsigned integer, float, decimal)
  number: [
    '==',
    '!=',
    '<',
    '<=',
    '>',
    '>=',
    'in',
    'not_in',
    'range',
    'out_range',
    'between',
    'contain',
    'not_contain',
    'exist',
    'not_exist',
    'null',
    'not_null',
  ],

  // 日期时间大类 (包含 date, time, datetime)
  date: ['==', '!=', '<', '<=', '>', '>=', 'range', 'out_range', 'before', 'current', 'between', 'exist', 'not_exist', 'null', 'not_null'],

  // IP地址类型 (ip)
  ip: ['==', '!=', 'in', 'not_in', 'contain', 'not_contain', 'exist', 'not_exist', 'null', 'not_null'],

  // 布尔类型 (boolean)
  boolean: ['==', '!=', 'true', 'false', 'exist', 'not_exist', 'null', 'not_null'],

  // JSON类型 (json)
  json: ['contain', 'not_contain', 'exist', 'not_exist', 'null', 'not_null'],

  // 向量类型 (vector)
  vector: ['knn', 'exist', 'not_exist', 'null', 'not_null'],

  // 二进制和空间类型大类 (包含 binary, point, shape)
  binary: ['exist', 'not_exist', 'null', 'not_null'],

  // 其他类型
  other: ['exist', 'not_exist', 'null', 'not_null'],
};

// 包含like，not_like，regexp, 'not_empty', 'empty'这三种操作符的string类型
const stringFieldTypes = ['text', 'string', 'binary'];

export const transformType = (type: string): string => {
  if (type === 'all Fields') {
    return 'all Fields';
  }

  // 将time、datetime和timestamp类型映射到date大类
  if (['time', 'datetime', 'timestamp'].includes(type)) {
    return 'date';
  }

  // 将integer、unsigned integer、float、decimal类型映射到number大类
  if (['integer', 'unsigned integer', 'float', 'decimal'].includes(type)) {
    return 'number';
  }

  // 将point和shape类型映射到binary大类
  if (['point', 'shape'].includes(type)) {
    return 'binary';
  }

  // 将other类型映射到binary大类，因为它们的操作符相同
  if (type === 'other') {
    return 'binary';
  }

  // 直接返回具体类型，以便与typeOperationMapping对应
  return type;
};

// 使用typeOperationMapping作为默认类型操作符映射
export const defaultTypeOption = typeOperationMapping;

export const transformOperation = (operation: string): string => {
  return operation === 'not_like' ? 'notLike' : operation;
};

export const findTypeByName = (name: string, fields: FieldList[]): string | undefined => {
  const field = fields.find((i: FieldList) => i.name === name);

  return field ? transformType(field.type) : undefined;
};

export const transformFilterFontToBack = (filters: any, fields: FieldList[]): any => {
  const { operation, field, value_from, value, sub_conditions } = filters;

  if (field) {
    const type = findTypeByName(field, fields);

    if ((operation === 'range' || operation === 'out_range') && type === 'number') {
      return {
        field,
        operation,
        value_from,
        value: value,
      };
    }

    if ((operation === 'range' || operation === 'out_range') && type === 'date') {
      return {
        field,
        operation,
        value_from,
        value: [dayjs(value.value[0]).format('YYYY-MM-DDTHH:mm:ss.SSSZ'), dayjs(value.value[1]).format('YYYY-MM-DDTHH:mm:ss.SSSZ')],
      };
    }

    if ((operation === 'match' || operation === 'match_phrase') && type === 'date') {
      return {
        field,
        operation,
        value_from,
        value: dayjs(value).format('YYYY-MM-DDTHH:mm:ss.SSSZ'),
      };
    }

    if (operation === 'multi_match') {
      return {
        field,
        operation,
        value_from,
        fields: value.fields,
        value: value.value,
        match_type: value.match_type,
      };
    }

    return {
      field,
      operation,
      value_from,
      value,
    };
  }

  return { operation, sub_conditions: sub_conditions?.map((item: any) => transformFilterFontToBack(item, fields)) };
};

export const transformFilterBackToFont = (filter: any, fields: any): DataFilterValue => {
  const { operation, value, field, value_from, sub_conditions } = filter;

  if (field) {
    let type = findTypeByName(field, fields);
    let val = value;

    // 当没有字段类型时，根据value的里面的值来转换格式，typeof value[0] ===number 推断类型为number
    // typeof value[0] ===string 推断类型为date
    if (!type && Array.isArray(value) && (operation === 'range' || operation === 'out_range')) {
      type = typeof value[0] === 'number' ? 'number' : 'date';
    }

    if ((operation === 'range' || operation === 'out_range') && type === 'number') {
      val = value;
    }

    if ((operation === 'match' || operation === 'match_phrase') && type === 'date') {
      val = dayjs(value);
    }

    if ((operation === 'range' || operation === 'out_range') && type === 'date') {
      val = {
        label: `${dayjs(value[0]).format('YYYY-MM-DD HH:mm:ss')} - ${dayjs(value[1]).format('YYYY-MM-DD HH:mm:ss')}`,
        value: [dayjs(value[0]), dayjs(value[1])],
      };
    }

    if (operation === 'multi_match') {
      val = {
        fields: filter.fields || [],
        value: filter.value || '',
        match_type: filter.match_type || 'best_fields',
      };
    }

    return { operation, value: val, field, value_from };
  }

  return { operation, sub_conditions: sub_conditions?.map((item: any) => transformFilterBackToFont(item, fields)) };
};
