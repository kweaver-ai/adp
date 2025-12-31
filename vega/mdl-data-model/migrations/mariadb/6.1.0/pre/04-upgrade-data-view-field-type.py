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


def get_virtual_type(origin_type, data_source_types):
    """根据原始类型和数据源类型获取虚拟类型"""
    origin_type_lower = origin_type.lower()

    # 特殊处理某些数据源类型的映射, 支持传入多种数据源类型
    # 先特殊处理多种数据源类型的映射
    if data_source_types:
        for data_source_type in data_source_types:
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


def upgrade_view_fields(conn_cursor):
    sql = """
    SELECT f_view_id, f_view_name, f_type, f_fields, f_data_scope, f_data_source_type FROM t_data_view;
    """
    conn_cursor.execute(sql)
    results = conn_cursor.fetchall()

    for res in results:
        view_id = res[0] 
        view_name = res[1]
        view_type = res[2]
        f_fields = res[3]

        # 确保视图ID不为空
        if not view_id:
            print(f"Skip invalid view record: view_id is empty")
            continue
        try:
            if view_type == "atomic":
                data_source_type = res[5] if res[5] else ""
                print(f"{view_type} view '{view_id}' '{view_name}' data_source_type: {data_source_type} start upgrade")
                try:
                    fields = json.loads(f_fields)
                except json.JSONDecodeError as e:
                    print(f"view '{view_id}' '{view_name}' fields json decode error: {str(e)}, Skip upgrade")
                    print(f"error fields: {f_fields}")
                    continue
                new_view_fields = upgrade_fields_type(fields, [data_source_type])
                fields_str = json.dumps(new_view_fields, ensure_ascii=False)
    
                # 更新视图的f_fields字段
                sql = """
                UPDATE t_data_view SET f_fields='{}' WHERE f_view_id='{}';
                """.format(fields_str, view_id)
                conn_cursor.execute(sql)
                print(f"{view_type} view '{view_id}' '{view_name}' fields type updated")
            # 将自定义视图data_scope字段的输出字段类型修改为vega类型
            elif view_type == "custom":
                data_source_type_list = []
                f_data_scope = res[4]
                print(f"{view_type} view '{view_id}' '{view_name}' start upgrade")
                try:
                    fields = json.loads(f_fields)
                except json.JSONDecodeError as e:
                    print(f"view '{view_id}' '{view_name}' fields json decode error: {str(e)}, Skip upgrade")
                    print(f"error fields: {f_fields}")
                    continue
                try:
                    data_scope = json.loads(f_data_scope)
                except json.JSONDecodeError as e:
                    print(f"view '{view_id}' '{view_name}' data_scope json decode error: {str(e)}, Skip upgrade")
                    print(f"error data_scope: {f_data_scope}")
                    continue
                if isinstance(data_scope, list):
                    # 先汇总 data_source_type_list
                    for item in data_scope:
                        node_type = item.get('type')
                        node_config = item.get('config')
                        output_fields = item.get('output_fields')
                        if node_type == "view":
                            atomic_view_id = node_config.get('view_id')
                            # 查询这个view的data_source_type
                            sql = """
                            SELECT f_data_source_type FROM t_data_view WHERE f_view_id='{}';
                            """.format(atomic_view_id)
                            try:
                                conn_cursor.execute(sql)
                                rows = conn_cursor.fetchall()
                                atomic_view_data_source_type = ""
                                for row in rows:
                                    atomic_view_data_source_type = row[0]
                                    if atomic_view_data_source_type:
                                        data_source_type_list.append(atomic_view_data_source_type)
                                
                                new_output_fields = upgrade_fields_type(output_fields, [atomic_view_data_source_type])
                                item['output_fields'] = new_output_fields
                            except Exception as e:
                                print(f"{view_type} view '{view_id}' '{view_name}' query atomic view '{atomic_view_id}' data_source_type error: {str(e)}, Skip this view node")
                                continue
                        
                    # 再更新其他节点类型的输出字段
                    for item in data_scope:
                        node_type = item.get('type')
                        output_fields = item.get('output_fields')
                        if node_type != "view":
                            new_output_fields = upgrade_fields_type(output_fields, data_source_type_list)
                            item['output_fields'] = new_output_fields

                    
                # 更新自定义视图字段类型为vega类型
                new_view_fields = upgrade_fields_type(fields, data_source_type_list)
                    
                fields_str = json.dumps(new_view_fields, ensure_ascii=False)
                data_scope_str = json.dumps(data_scope, ensure_ascii=False)

                sql = """
                UPDATE t_data_view SET f_fields='{}', f_data_scope='{}' WHERE f_view_id='{}';
                """.format(fields_str,  data_scope_str, view_id)
                conn_cursor.execute(sql)
                
                print(f"{view_type} view '{view_id}' '{view_name}' fields type updated")
            else:
                print(f"{view_type} view '{view_id}' '{view_name}' unsupported type")
        except Exception as e:
            print(f"{view_type} view '{view_id}' '{view_name}' upgrade fields type error: {str(e)}, Skip upgrade")
    
    print(f"{len(results)} views fields type updated")


def upgrade_fields_type(fields, data_source_type_list):
    if isinstance(fields, list):
        for field in fields:
            if isinstance(field, dict) and 'type' in field:
                field_type = field["type"]
                if field_type != "":
                    # 获取新的虚拟类型（传入数据源类型）
                    new_virtual_type = get_virtual_type(field_type, data_source_type_list)
                    if new_virtual_type:
                        field["type"] = new_virtual_type
    
    return fields


if __name__ == "__main__":
    conn = get_conn(os.environ["DB_USER"], os.environ["DB_PASSWD"],
                    os.environ["DB_HOST"], os.environ["DB_PORT"], "dip_mdl")

    conn_cursor = conn.cursor()
    try:
        upgrade_view_fields(conn_cursor)
    except Exception as e:
        print(f"Upgrade failed: {str(e)}")
        raise  
    finally:
        conn_cursor.close()
        conn.close()

