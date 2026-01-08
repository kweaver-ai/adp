import { includes } from 'lodash-es';

const boolean = ['boolean'];
const date = ['timestamp', 'datetime'];
const number = ['short', 'integer', 'long', 'float', 'double', 'decimal'];

const formatTypeNew = (type: any): any => {
  if (!type) return '';
  const value = type?.toLowerCase();

  if (includes(boolean, value)) return 'boolean';
  if (includes(date, value)) return 'date';
  if (includes(number, value)) return 'number';

  return 'string';
};

export default formatTypeNew;
