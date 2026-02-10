export const PRIMARY_KEY_TYPES = ['integer', 'unsigned integer', 'string'] as const;

export const INCREMENTAL_KEY_TYPES = ['integer', 'unsigned integer', 'datetime', 'timestamp'] as const;

export const DISPLAY_KEY_TYPES = [
  'integer',
  'unsigned integer',
  'string',
  'text',
  'float',
  'decimal',
  'date',
  'time',
  'datetime',
  'timestamp',
  'ip',
  'boolean',
] as const;

export const canBePrimaryKey = (type?: string) => !!type && (PRIMARY_KEY_TYPES as readonly string[]).includes(type);
export const canBeIncrementalKey = (type?: string) => !!type && (INCREMENTAL_KEY_TYPES as readonly string[]).includes(type);
export const canBeDisplayKey = (type?: string) => !!type && (DISPLAY_KEY_TYPES as readonly string[]).includes(type);
