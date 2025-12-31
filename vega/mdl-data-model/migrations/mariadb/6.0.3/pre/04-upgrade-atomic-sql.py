# 将原子视图的sql去掉视图源vdm的依赖

#!/usr/bin/env python3
# -*- coding: utf-8 -*-


import os
import json
import rdsdriver


def get_conn(user, password, host, port, database):
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

def upgrade_atomic_sql(conn_cursor):
    # 获取数据源列表
    data_source_map = list_data_sources(conn_cursor)

    # 获取查询类型为SQL的原子视图
    atomic_view_sql = """
    SELECT f_view_id, f_view_name, f_technical_name, f_data_source_type, f_data_source_id FROM t_data_view WHERE f_type = "atomic" AND f_query_type = "SQL";
    """

    conn_cursor.execute(atomic_view_sql)
    results = conn_cursor.fetchall()
    print(f"need upgrade atomic sql view count: {len(results)}")

    for res in results:
        view_id = res[0]
        view_name = res[1]
        view_technical_name = res[2]
        view_data_source_type = res[3]
        view_data_source_id = res[4]

        # 获取数据源类型
        if view_data_source_id in data_source_map:
            view_data_source = data_source_map[view_data_source_id]
        else:
            print("view '{}' data source id {} not found".format(view_name, view_data_source_id))
            raise Exception("view '{}' data source id {} not found".format(view_name, view_data_source_id))

        view_data_source_type = view_data_source["type_name"]
        bin_data = view_data_source["bin_data"]
        if bin_data is not None:
            try:
                # 解码BLOB数据为JSON字符串
                json_str = bin_data.decode('utf-8')
                # 转换JSON字符串为字典
                data_map = json.loads(json_str)
                catalog = data_map["catalog_name"]
                if view_data_source_type == "excel":
                    schema = "default"
                else:
                    schema = data_map.get("schema", "")
                    if schema == "":
                        schema = data_map.get("database_name", "")

            except Exception as e:
                print(f"Error decode bin_data: {str(e)}")
                raise e
        else:
            print(f"view {view_id} has no bin_data")
            raise Exception("view {} has no bin_data".format(view_name))

        new_view_sql = f'SELECT * FROM {catalog}."{schema}"."{view_technical_name}"'
        meta_table_name = f'{catalog}."{schema}"."{view_technical_name}"'

        print(f"Upgrade atomic sql view '{view_name}' SQL")
        update_data_view_sql = f"""
        UPDATE t_data_view SET f_meta_table_name='{meta_table_name}', f_sql='{new_view_sql}' WHERE f_view_id='{view_id}';
        """

        conn_cursor.execute(update_data_view_sql)
        print(f"Upgrade atomic sql view '{view_name}' success")


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
                print("Found vega database")
                break

        if not vega_db_exists:
            print("vega database not found, skip upgrade")
            return False

        # 2. 检查 vega 数据库中的 data_source 表是否存在
        show_tables_sql = "SHOW TABLES FROM vega LIKE 'data_source';"
        conn_cursor.execute(show_tables_sql)
        tables = conn_cursor.fetchall()

        if len(tables) > 0:
            print("Found vega.data_source table")
            return True
        else:
            print("vega.data_source table not found, skip upgrade")
            return False

    except Exception as e:
        print(f"Check vega database and table error: {str(e)}")
        return False


if __name__ == "__main__":
    conn = get_conn(os.environ["DB_USER"], os.environ["DB_PASSWD"],
                    os.environ["DB_HOST"], os.environ["DB_PORT"], "dip_mdl")

    conn_cursor = conn.cursor()
    try:
        rslt = check_vega_db_and_table(conn_cursor)
        if rslt:
            upgrade_atomic_sql(conn_cursor)
    except Exception:
        raise Exception()
    finally:
        conn_cursor.close()
        conn.close()


