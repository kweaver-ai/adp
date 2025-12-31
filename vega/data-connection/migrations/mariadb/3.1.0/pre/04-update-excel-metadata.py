#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import json
import uuid
import rdsdriver


def get_conn(user, password, host, port, database):
    """获取数据库连接，支持多租户模式"""
    try:

        conn = rdsdriver.connect(host=host,
                                 port=int(port),
                                 user=user,
                                 password=password,
                                 database=database,
                                 autocommit=True)
    except Exception as e:
        print(f"connect database error: {str(e)}")
        raise e
    return conn


def migrate_excel_config_to_table_scan(conn):
    """将excel配置迁移到表扫描配置"""
    cursor = conn.cursor()

    try:
        # 1. 查询所有需要迁移的excel配置
        cursor.execute("""
            SELECT 
                etc.id, etc.catalog, etc.vdm_catalog, etc.schema_name, 
                etc.file_name, etc.table_name, etc.table_comment,
                etc.sheet, etc.all_sheet, etc.sheet_as_new_column,
                etc.start_cell, etc.end_cell, etc.has_headers,
                dsi.f_id as data_source_id,
                etc.create_time, etc.update_time
            FROM excel_table_config etc
            LEFT JOIN t_data_source_info dsi ON etc.catalog = dsi.f_catalog
        """)
        excel_configs = cursor.fetchall()

        # 2. 为每条记录准备迁移数据
        for config in excel_configs:
            # 构建高级参数JSON
            advanced_params = [
                {"key":"sheet","value":config[7]},  # sheet
                {"key":"allSheet","value":bool(config[8])},  # all_sheet
                {"key":"sheetAsNewColumn","value":bool(config[9])},  # sheet_as_new_column
                {"key":"startCell","value":config[10]},  # start_cell
                {"key":"endCell","value":config[11]},  # end_cell
                {"key":"hasHeaders","value":bool(config[12])},  # has_headers
                {"key":"fileName","value":config[4]}  # file_name
            ]

            # 转换为JSON字符串
            advanced_params_str = json.dumps(advanced_params, separators=(',', ':'))

            # 准备插入t_table_scan的数据
            table_scan_data = (
                str(uuid.uuid4()),  # f_id
                config[5],  # f_name (table_name)
                advanced_params_str,
                config[6],  # f_description (table_comment)
                0,  # f_table_rows
                config[13],  # f_data_source_id
                config[1],  # f_data_source_name (catalog)
                config[3] if config[3] else 'default',  # f_schema_name
                '-1',  # f_task_id
                1,  # f_version
                config[14],  # f_create_time
                '',  # f_create_user
                config[15],  # f_operation_time
                '',  # f_operation_user
                0,  # f_operation_type
                0,  # f_status
                0  # f_status_change
            )

            # 3. 插入数据到t_table_scan
            cursor.execute("""
                INSERT INTO t_table_scan (
                    f_id, f_name, f_advanced_params, f_description, f_table_rows,
                    f_data_source_id, f_data_source_name, f_schema_name, f_task_id,
                    f_version, f_create_time, f_create_user, f_operation_time,
                    f_operation_user, f_operation_type, f_status, f_status_change
                ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            """, table_scan_data)

        conn.commit()
        print(f"成功迁移 {len(excel_configs)} 条表记录")



    except Exception as e:
        conn.rollback()
        raise Exception(f"数据迁移失败: {str(e)}")
    finally:
        cursor.close()


def migrate_excel_column_type_to_table_field_scan(conn):
    """将excel列类型配置迁移到表字段扫描配置"""
    cursor = conn.cursor()

    try:
        # 1. 查询所有需要迁移的excel列类型配置
        cursor.execute("""
            SELECT 
                ect.catalog, ect.schema_name, ect.table_name, ect.column_name,
                ect.column_comment, ect.type, ect.order_no, ect.create_time,
                ect.update_time, tts.f_id as table_id
            FROM excel_column_type ect
            LEFT JOIN t_table_scan tts ON 
                ect.catalog = tts.f_data_source_name AND 
                ect.table_name = tts.f_name
        """)
        column_configs = cursor.fetchall()

        for config in column_configs:
            # 构建高级参数JSON
            origin_type = config[5]  # 原始类型
            virtual_type = origin_type  # 默认虚拟类型与原始类型相同

            # 根据不同类型设置对应的虚拟类型
            if origin_type.lower() == 'bigint':
                virtual_type = 'integer'
            elif origin_type.lower() == 'varchar':
                virtual_type = 'string'
            elif origin_type.lower() == 'double':
                virtual_type = 'float'

            advanced_params = [
                {"key": "originFieldType", "value": origin_type},
                {"key": "virtualFieldType", "value": virtual_type}
            ]
            advanced_params_str = json.dumps(advanced_params, separators=(',', ':'))

            # 准备插入t_table_field_scan的数据
            field_scan_data = (
                str(uuid.uuid4()),  # f_id
                config[3],  # f_field_name (column_name)
                config[9],  # f_table_id (table_id)
                config[2],  # f_table_name (table_name)
                config[5],  # f_field_type (type)
                None,  # f_field_length
                None,  # f_field_precision
                config[4],  # f_field_comment (column_comment)
                config[6],  # f_field_order_no (order_no)
                advanced_params_str,  # f_advanced_params
                1,  # f_version
                config[7],  # f_create_time (create_time)
                '',  # f_create_user
                config[8],  # f_operation_time (update_time)
                '',  # f_operation_user
                0,  # f_operation_type
                0  # f_status_change
            )

            # 插入数据到t_table_field_scan
            cursor.execute("""
                INSERT INTO t_table_field_scan (
                    f_id, f_field_name, f_table_id, f_table_name, f_field_type,
                    f_field_length, f_field_precision, f_field_comment, f_field_order_no,
                    f_advanced_params, f_version, f_create_time, f_create_user,
                    f_operation_time, f_operation_user, f_operation_type, f_status_change
                ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            """, field_scan_data)

        conn.commit()
        print(f"成功迁移 {len(column_configs)} 条字段记录")

    except Exception as e:
        conn.rollback()
        raise Exception(f"数据迁移失败: {str(e)}")
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

        migrate_excel_config_to_table_scan(conn)
        migrate_excel_column_type_to_table_field_scan(conn)
    except Exception as e:
        print(f"执行过程中发生错误: {str(e)}")
    finally:
        if 'conn' in locals() and conn:
            conn.close()

