#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import json
import time
import rdsdriver

ViewScanStatus_New      = "new"
ViewScanStatus_Modify   = "modify"
ViewScanStatus_Delete   = "delete"
ViewScanStatus_NoChange = "no_change"

FieldScanStatus_New        = "new"
FieldScanStatus_Modify     = "modify"
FieldScanStatus_Delete     = "delete"
FieldScanStatus_NoChange   = "no_change"
FieldScanStatus_NotSupport = "not_support"


def get_conn(user, password, host, port, database):
    try:
        conn = rdsdriver.connect(host=host,
                                 port=int(port),
                                 user=user,
                                 password=password,
                                 database=database,
                                 autocommit=True)
    except Exception as e:
        print("connect database error: %s", str(e))
        raise e
    return conn

def upgrade_vega_logic_view(conn_cursor):
    # 获取数据源列表
    data_source_map = list_data_sources(conn_cursor)

    vega_view_sql = """
    SELECT fv.id, fv.technical_name, fv.business_name, fv.datasource_id, fv.status, fv.metadata_form_id, fv.description, fv.created_at,
    fv.created_by_uid, fv.updated_at, fv.updated_by_uid, fv.deleted_at, fv.excel_file_name, fv.excel_sheet, fv.start_cell, fv.end_cell, fv.has_headers,
    fv.sheet_as_new_column, fvs.form_view_id, fvs.sql FROM vega.form_view fv left join vega.form_view_sql fvs on fv.id=fvs.form_view_id;
    """

    conn_cursor.execute(vega_view_sql)
    results = conn_cursor.fetchall()
    print(f"vega 视图数量: {len(results)}")

    for res in results:
        view_id = res[0]
        view_technical_name = res[1]
        view_business_name = res[2]
        view_data_source_id = res[3]
        view_status = res[4]
        view_metadata_form_id = res[5]
        view_description = res[6]
        view_created_at = res[7]
        view_created_by_uid = res[8]
        view_updated_at = res[9]
        view_updated_by_uid = res[10]
        view_deleted_at = res[11]
        view_excel_file_name = res[12]
        view_excel_sheet = res[13]
        view_start_cell = res[14]
        view_end_cell = res[15]
        view_has_headers = res[16]
        view_sheet_as_new_column = res[17]
        view_sql = res[19]

        # 如果被删除了，就跳过
        if view_deleted_at != 0:
            print(f"view technical name: {view_technical_name} is deleted, skip")
            continue

        print(f"当前升级的视图是: {view_technical_name}")

        # 获取字段列表
        field_sql = """
        SELECT technical_name, business_name, comment, status, primary_key, data_type, data_length, data_accuracy,
        is_nullable, deleted_at, business_timestamp FROM vega.form_view_field WHERE form_view_id='{}';
        """.format(view_id)

        conn_cursor.execute(field_sql)
        field_results = conn_cursor.fetchall()

        # 转换字段列表
        view_field_list = []
        view_primary_keys = []
        for field in field_results:
            field_technical_name = field[0]
            field_business_name = field[1]
            field_comment = field[2]
            field_status = field[3]
            field_primary_key = field[4]
            field_data_type = field[5]
            field_data_length = field[6]
            field_data_accuracy = field[7]
            field_is_nullable = field[8]
            field_deleted_at = field[9]
            field_business_timestamp = field[10]

            # 被删除的字段打个日志, 只添加没被删除的字段
            if field_deleted_at != 0:
                print(f"field technical name: {field_technical_name} is deleted, skip")
                continue

            if field_status == 0:
                new_field_status = ""
            # 转换 field_status
            elif field_status == 1:
                new_field_status = FieldScanStatus_NoChange
            elif field_status == 2:
                new_field_status = FieldScanStatus_New
            elif field_status == 3:
                new_field_status = FieldScanStatus_Modify
            elif field_status == 4:
                new_field_status = FieldScanStatus_Delete
            else:
                new_field_status = FieldScanStatus_NotSupport

            # 转换 primary_key
            if field_primary_key is not None:
                if field_primary_key == 1:
                    view_primary_keys.append(field_technical_name)

            # 转换 business_timestamp
            if field_business_timestamp is None:
                new_business_timestamp = False
            elif  field_business_timestamp == 1:
                new_business_timestamp = True
            else:
                new_business_timestamp = False

            view_field_list.append({
                "name":  field_technical_name,
                "original_name":  field_technical_name,
                "display_name":  field_business_name,
                "comment":  field_comment,
                "type":  field_data_type,
                "status": new_field_status,
                "data_length":  field_data_length,
                "data_accuracy":  field_data_accuracy,
                "is_nullable":  field_is_nullable,
                "business_timestamp": new_business_timestamp,
            })

        view_fields_str = json.dumps(view_field_list, ensure_ascii=False)

        # 规范备注
        if view_description is None:
            view_description = ""

        # 获取数据源类型
        if view_data_source_id in data_source_map:
            view_data_source = data_source_map[view_data_source_id]
        else:
            print("view '{}' data source id {} not found".format(view_technical_name, view_data_source_id))
            raise Exception("view '{}' data source id {} not found".format(view_technical_name, view_data_source_id))

        view_data_source_name = view_data_source["name"]
        view_data_source_type = view_data_source["type_name"]
        bin_data = view_data_source["bin_data"]
        if bin_data is not None:
            try:
                # 解码BLOB数据为JSON字符串
                json_str = bin_data.decode('utf-8')
                # 转换JSON字符串为字典
                data_map = json.loads(json_str)
                data_view_source = data_map["data_view_source"]
            except Exception as e:
                print(f"Error decode bin_data: {str(e)}")
                raise e
        else:
            print(f"view {view_id} has no bin_data")
            raise Exception("view {} has no bin_data".format(view_technical_name))

        if view_sql is None:
            view_sql = f"select * from {data_view_source}.{view_technical_name}"


        # 转换 view_status
        if view_status == 0:
            new_view_status = ""
        if view_status == 1:
            new_view_status = ViewScanStatus_NoChange
        elif view_status == 2:
            new_view_status = ViewScanStatus_New
        elif view_status == 3:
            new_view_status = ViewScanStatus_Modify
        elif view_status == 4:
            new_view_status = ViewScanStatus_Delete
        else:
            new_view_status = ""

        # 构造 excel config
        if view_excel_file_name is None:
            view_excel_file_name = ""
        if view_excel_file_name != "":
            if view_has_headers == 1:
                new_view_has_headers = True
            else:
                new_view_has_headers = False

            if view_sheet_as_new_column == 1:
                new_view_sheet_as_new_column = True
            else:
                 new_view_sheet_as_new_column = False

            excel_config = {
                "sheet": view_excel_sheet,
                "start_cell": view_start_cell,
                "end_cell": view_end_cell,
                "has_headers": new_view_has_headers,
                "sheet_as_new_column": new_view_sheet_as_new_column
            }
        else:
            excel_config = None

        excel_config_str = json.dumps(excel_config, ensure_ascii=False)
        # view_data_scope = json.dumps(None, ensure_ascii=False)
        primaryKeysStr = tag_slice_2_tag_string(view_primary_keys)

        # 先查询分组是否存在，不存在就新建
        group_id = view_data_source_id
        group_name = view_data_source_name
        current_time = int(time.time() * 1000)

        check_group_sql = """
            SELECT f_group_id FROM t_data_view_group WHERE f_group_id='{}';
        """.format(group_id)
        conn_cursor.execute(check_group_sql)
        group_results = conn_cursor.fetchall()
        if len(group_results) == 0:
             # 插入分组表 sql
            insert_group_sql = """
            INSERT INTO t_data_view_group (f_group_id, f_group_name, f_create_time, f_update_time, f_builtin)
            VALUES ('{}', '{}', {}, {}, {});
            """.format(group_id, group_name, current_time, current_time, 0)

            conn_cursor.execute(insert_group_sql)
            print(f"group_id: {group_id}, group_name: {group_name} 插入成功")
        else:
            print(f"group_id: {group_id}, group_name: {group_name} 已存在, 更新分组信息")
            # 更新分组信息
            update_group_sql = """
            UPDATE t_data_view_group SET f_group_name='{}', f_update_time={} WHERE f_group_id='{}';
            """.format(group_name, current_time, group_id)
            conn_cursor.execute(update_group_sql)
            print(f"group_id: {group_id}, group_name: {group_name} 更新成功")


        # 先查询视图是否存在，不存在就新建
        check_view_sql = """
            SELECT f_view_id FROM t_data_view WHERE f_view_id='{}';
        """.format(view_id)
        conn_cursor.execute(check_view_sql)
        view_results = conn_cursor.fetchall()

        view_create_time = datetime_str_to_ms(view_created_at)
        view_update_time = datetime_str_to_ms(view_updated_at)

        if len(view_results) == 0:
            # 插入视图表 sql
            insert_data_view_sql = """
            INSERT INTO t_data_view (f_view_id, f_view_name, f_technical_name, f_group_id, f_type, f_query_type, f_builtin, f_tags, f_comment,
            f_data_source_type, f_data_source_id, f_file_name, f_excel_config, f_fields, f_status, f_metadata_form_id,
            f_primary_keys, f_sql, f_create_time, f_update_time, f_creator, f_updater) VALUES ('{}', '{}', '{}', '{}', '{}', '{}', {}, '{}', '{}',
            '{}', '{}', '{}', '{}', '{}', '{}', '{}',
            '{}', '{}', {}, {}, '{}', '{}');
            """.format(view_id, view_business_name, view_technical_name, view_data_source_id, "atomic", "SQL", 0, "", view_description,
            view_data_source_type, view_data_source_id, view_excel_file_name, excel_config_str, view_fields_str, new_view_status, view_metadata_form_id,
            primaryKeysStr, view_sql, view_create_time, view_update_time, view_created_by_uid, view_updated_by_uid)

            # print(insert_data_view_sql)
            conn_cursor.execute(insert_data_view_sql)
            print(f"view_technical_name: {view_technical_name} 插入成功")
        else:
            print(f"view_technical_name: {view_technical_name} 已存在, 更新视图信息")
            # 更新视图信息
            update_data_view_sql = """
            UPDATE t_data_view SET f_view_name='{}', f_technical_name='{}', f_group_id='{}', f_type='{}', f_query_type='{}', f_builtin={}, f_tags='{}', f_comment='{}',
            f_data_source_type='{}', f_data_source_id='{}', f_file_name='{}', f_excel_config='{}', f_fields='{}', f_status='{}', f_metadata_form_id='{}',
            f_primary_keys='{}', f_sql='{}', f_update_time={}, f_updater='{}' WHERE f_view_id='{}';
            """.format(view_business_name, view_technical_name, view_data_source_id, "atomic", "SQL", 0, "", view_description,
            view_data_source_type, view_data_source_id, view_excel_file_name, excel_config_str, view_fields_str, new_view_status, view_metadata_form_id,
            primaryKeysStr, view_sql, view_update_time, view_updated_by_uid, view_id)

            conn_cursor.execute(update_data_view_sql)
            print(f"view_technical_name: {view_technical_name} 更新成功")

def upgrade_scan_record(conn_cursor):
    sql = """
    SELECT id, datasource_id, scanner, scan_time, status, metadata_task_id FROM scan_record;
    """
    conn_cursor.execute(sql)
    results = conn_cursor.fetchall()

    for res in results:
        record_id = res[0]
        data_source_id = res[1]
        scanner = res[2]
        scan_time = res[3]
        data_source_status = res[4]
        metadata_task_id = res[5]

        # record id
        new_scan_time = datetime_str_to_ms(scan_time)

        # 转换 data_source_status
        if data_source_status == 1:
            new_data_source_status = "avaliable"
        elif data_source_status == 2:
            new_data_source_status = "scanning"
        else:
            new_data_source_status = ""

        # 检查是否已经存在
        check_sql = """
        SELECT f_record_id FROM t_scan_record WHERE f_record_id='{}';
        """.format(str(record_id))
        conn_cursor.execute(check_sql)
        record_results = conn_cursor.fetchall()
        if len(record_results) == 0:
            insert_sql = """
            INSERT INTO t_scan_record (f_record_id, f_data_source_id, f_scanner, f_scan_time, f_data_source_status, f_metadata_task_id)
            VALUES ('{}', '{}', '{}', {}, '{}', '{}');
            """.format(str(record_id), data_source_id, scanner, new_scan_time, new_data_source_status, metadata_task_id)

            conn_cursor.execute(insert_sql)
        else:
            print(f"record_id: {record_id} 已存在, 更新扫描记录")
            update_sql = """
            UPDATE t_scan_record SET f_data_source_id='{}', f_scanner='{}', f_scan_time={}, f_data_source_status='{}', f_metadata_task_id='{}' WHERE f_record_id='{}';
            """.format(data_source_id, scanner, new_scan_time, new_data_source_status, metadata_task_id, record_id)
            conn_cursor.execute(update_sql)

            print(f"record_id: {record_id} 更新成功")


def list_data_sources(conn_cursor):
    # 获取数据源列表
    data_source_sql = """
    SELECT id, name, type_name, bin_data, comment FROM vega.data_source;
    """
    conn_cursor.execute(data_source_sql)
    data_source_results = conn_cursor.fetchall()

    # 遍历数据源
    data_source_map = {}
    for data_source in data_source_results:
        data_source_id = data_source[0]
        data_source_name = data_source[1]
        data_source_type_name = data_source[2]
        data_source_bin_data = data_source[3]
        data_source_comment = data_source[4]

        data_source_map[data_source_id] = {
            "id": data_source_id,
            "name": data_source_name,
            "type_name": data_source_type_name,
            "bin_data": data_source_bin_data,
            "comment": data_source_comment,
        }

    return data_source_map


def check_vega_db_and_table(conn_cursor):
    """
    检查 vega 数据库和 form_view 表是否存在
    返回 True 如果都存在，否则返回 False
    """
    try:
        # 1. 检查数据库列表中是否存在 vega 数据库
        show_databases_sql = "SHOW DATABASES;"
        conn_cursor.execute(show_databases_sql)
        databases = conn_cursor.fetchall()

        vega_db_exists = False
        for db in databases:
            if db[0] == 'vega':
                vega_db_exists = True
                print("找到 vega 数据库")
                break

        if not vega_db_exists:
            print("未找到 vega 数据库，跳过升级")
            return False

        # 2. 检查 vega 数据库中的 form_view 表是否存在
        show_tables_sql = "SHOW TABLES FROM vega LIKE 'form_view';"
        conn_cursor.execute(show_tables_sql)
        tables = conn_cursor.fetchall()

        if len(tables) > 0:
            print("找到 vega.form_view 表")
            return True
        else:
            print("未找到 vega.form_view 表，跳过升级")
            return False

    except Exception as e:
        print(f"检查 vega 数据库和表时发生错误: {str(e)}")
        return False


# 把primary_keys数组转成数据库存储的字符串的形式，格式为 "a","b","c"
def tag_slice_2_tag_string(strs):
    # 将每个字符串用双引号包裹，然后用逗号连接
    return ",".join([f'"{s}"' for s in strs])

# - 解析datetime字符串, 转化成毫秒时间戳
def datetime_str_to_ms(dt):
    # 转换为毫秒时间戳
    return int(dt.timestamp() * 1000)



if __name__ == "__main__":
    conn = get_conn(os.environ["DB_USER"], os.environ["DB_PASSWD"],
                    os.environ["DB_HOST"], os.environ["DB_PORT"], "dip_mdl")

    conn_cursor = conn.cursor()
    try:
        rslt = check_vega_db_and_table(conn_cursor)
        if rslt:
            upgrade_vega_logic_view(conn_cursor)
            upgrade_scan_record(conn_cursor)
    except Exception:
        raise Exception()
    finally:
        conn_cursor.close()
        conn.close()


