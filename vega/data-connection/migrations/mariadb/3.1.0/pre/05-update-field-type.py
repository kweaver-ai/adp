#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import json
import rdsdriver

# 类型映射字典
TYPE_MAPPING = {
    'integer': [
        'tinyint', 'int1', 'smallint', 'int2', 'mediumint', 'serial', 'int',
        'integer', 'int3', 'int4', 'mediumint', 'middleint', 'year', 'bigint',
        'int8', 'serial8', 'bigserial', 'pls_integer', 'long', 'int16', 'int32',
        'int64', 'tinyint unsigned', 'smallint unsigned', 'mediumint unsigned',
        'int unsigned', 'integer unsigned', 'bigint unsigned', 'uint8', 'uint16',
        'uint32', 'uint64', 'byte'
    ],
    'float': [
        'double', 'double unsigned', 'float8', 'float unsigned', 'real',
        'real unsigned', 'double precision', 'double precision unsigned',
        'float', 'smallfloat', 'float4', 'binary_double', 'binary_float',
        'float64', 'float32', 'money'
    ],
    'decimal': [
        'decimal', 'decimal unsigned', 'numeric', 'numeric unsigned', 'fixed',
        'fixed unsigned', 'dec', 'dec unsigned', 'number', 'desc'
    ],
    'string': [
        'char', 'nchar', 'fixedstring', 'string', 'interval_year_month',
        'interval_day_time', 'varchar', 'varchar2', 'nvarchar', 'nvarchar2',
        'character varying', 'rowid', 'character', 'bpchar', 'interval',
        'character_varying', 'interval year', 'interval month', 'interval day',
        'interval hour', 'interval minute', 'interval second',
        'interval year to month', 'interval day to hour',
        'interval day to minute', 'interval day to second',
        'interval hour to minute', 'interval hour to second',
        'interval minute to second', 'regex'
    ],
    'text': [
        'text', 'tinytext', 'mediumtext', 'longtext', 'longvarchar', 'nclob',
        'ntext', 'clob'
    ],
    'date': ['date'],
    'timestamp': [
        'datetime year to month', 'datetime year to year', 'datetime day to day',
        'datetime day to hour', 'datetime year to minute', 'datetime year to second',
        'datetime year to fraction', 'timestamp', 'timestamp_ntz',
        'timestamp with time zone', 'timestamp with local time zone', 'timestamptz',
        'smalldatetime', 'datetime2', 'datetimeoffset'
    ],
    'time': ['time', 'time with time zone', 'timetz'],
    'datetime': ['datetime'],
    'boolean': ['bool', 'boolean'],
    'binary': [
        'binary', 'varbinary', 'raw', 'long raw', 'image', 'longvarbinary',
        'blob', 'bfile', 'bytea', 'tinyblob', 'mediumblob', 'longblob',
        'bindata', 'bit'
    ],
    'json': ['json', 'jsonb']
}


def get_conn(user, password, host, port, database):
    """获取数据库连接"""
    try:
        conn = rdsdriver.connect(
            host=host,
            port=int(port),
            user=user,
            password=password,
            database=database,
            autocommit=True
        )
    except Exception as e:
        print(f"connect database error: {str(e)}")
        raise e
    return conn


def get_virtual_type(origin_type, data_source_type=None):
    """根据原始类型和数据源类型获取虚拟类型"""
    origin_type_lower = origin_type.lower()

    # 特殊处理某些数据源类型的映射
    if data_source_type:
        data_source_type = data_source_type.lower()
        if data_source_type == 'gbase':
            if origin_type_lower in ['byte']:
                return 'binary'
            elif origin_type_lower in ['money']:
                return 'decimal'
            elif origin_type_lower in ['datetime']:
                return 'timestamp'
        elif data_source_type == 'dameng':
            if origin_type_lower in ['clob']:
                return 'binary'
            elif origin_type_lower in ['long']:
                return 'text'
            elif origin_type_lower in ['bit']:
                return 'boolean'
        elif data_source_type == 'mongodb':
            if origin_type_lower in ['timestamp']:
                return 'date'
            elif origin_type_lower in ['date']:
                return 'timestamp'
        elif data_source_type in ['maxcompute', 'sqlserver', 'clickhouse', 'doris']:
            if origin_type_lower in ['datetime']:
                return 'timestamp'
        elif data_source_type == 'oracle':
            if origin_type_lower in ['date']:
                return 'datetime'
        elif data_source_type == 'sqlserver':
            if origin_type_lower in ['bit']:
                return 'boolean'

    # 默认映射
    for virtual_type, origin_types in TYPE_MAPPING.items():
        if origin_type_lower in origin_types:
            return virtual_type
    return None


def update_virtual_field_types(conn):
    """更新t_table_field表中的virtualFieldType值"""
    cursor = conn.cursor()

    try:
        # 1. 查询所有需要更新的字段（关联数据源信息）
        cursor.execute("""
            SELECT tf.f_table_id, tf.f_field_name, tf.f_advanced_params, tdi.f_type
            FROM t_table_field tf
            JOIN t_table t ON tf.f_table_id = t.f_id
            JOIN t_data_source_info tdi ON t.f_data_source_id = tdi.f_id
            WHERE tf.f_delete_flag != 1 AND tf.f_advanced_params LIKE '%originFieldType%'
        """)
        fields = cursor.fetchall()

        updated_count = 0

        for field in fields:
            table_id, field_name, advanced_params_str, data_source_type = field
            try:
                advanced_params = json.loads(advanced_params_str)
                origin_type = None
                virtual_type_param = None

                # 查找originFieldType和virtualFieldType
                for param in advanced_params:
                    if param.get('key') == 'originFieldType':
                        origin_type = param.get('value')
                    elif param.get('key') == 'virtualFieldType':
                        virtual_type_param = param

                if origin_type:
                    # 获取新的虚拟类型（传入数据源类型）
                    new_virtual_type = get_virtual_type(origin_type, data_source_type)

                    if new_virtual_type and virtual_type_param:
                        # 更新virtualFieldType的值
                        virtual_type_param['value'] = new_virtual_type

                        # 更新数据库
                        new_advanced_params = json.dumps(advanced_params, separators=(',', ':'))
                        cursor.execute("""
                            UPDATE t_table_field 
                            SET f_advanced_params = %s 
                            WHERE f_table_id = %s AND f_field_name = %s
                        """, (new_advanced_params, table_id, field_name))
                        updated_count += 1

                        # 每1000条记录打印一次进度
                        if updated_count % 1000 == 0:
                            print(f"已处理 {updated_count} 条记录...")

            except Exception as e:
                print(f"更新字段{table_id}.{field_name}失败: {str(e)}")
                conn.rollback()  # 任何失败立即回滚
                return  # 直接返回，不再继续处理

        conn.commit()
        print(f"成功更新 {updated_count} 条字段记录")

    except Exception as e:
        conn.rollback()
        raise Exception(f"更新失败: {str(e)}")
    finally:
        cursor.close()


if __name__ == "__main__":
    try:
        # 获取数据库连接
        conn = get_conn(os.environ["DB_USER"],
                        os.environ["DB_PASSWD"],
                        os.environ["DB_HOST"],
                        os.environ["DB_PORT"],
                        "vega")

        # 执行更新
        update_virtual_field_types(conn)

    except Exception as e:
        print(f"执行过程中发生错误: {str(e)}")
    finally:
        if 'conn' in locals() and conn:
            conn.close()
