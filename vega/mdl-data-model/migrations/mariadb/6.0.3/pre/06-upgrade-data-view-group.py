# 更新扫描原子视图时建的分组f_builtin属性，原子视图的分组f_builtin属性应为1

#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import uuid
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

def upgrade_data_view_group(conn_cursor):
    # 获取数据源列表
    data_source_map = list_data_sources(conn_cursor)

    # 获取全部分组
    group_sql = """
    SELECT f_group_id, f_group_name FROM t_data_view_group;
    """

    conn_cursor.execute(group_sql)
    results = conn_cursor.fetchall()

    atomic_view_group_ids = []
    for res in results:
        group_id = res[0]
        group_name = res[1]

        if group_id in data_source_map:
            print(f"Group {group_name} is atomic view group")
            atomic_view_group_ids.append(group_id)
        else:
            try:
                uuid.UUID(group_id)
                print(f"Group {group_name} is atomic view group, not in data source list, may be data source deleted")
                atomic_view_group_ids.append(group_id)
            except ValueError:
                # 如果不是有效UUID，则跳过，不做任何处理
                print(f"Group {group_name} is not atomic view group, ignore")
            
    
    print(f"Need to update atomic view group count: {len(atomic_view_group_ids)}")
    # 批量更新分组的f_builtin字段为1    
    if len(atomic_view_group_ids) > 0:
        # 更新分组信息
        # 使用参数绑定机制，不要混合两种字符串格式化方式
        placeholders = ','.join(['%s'] * len(atomic_view_group_ids))  # 使用 %s 作为占位符
        update_group_sql = f"""
        UPDATE t_data_view_group SET f_builtin=1 WHERE f_group_id in ({placeholders});
        """
        print(f"Execute sql: {update_group_sql}")
        conn_cursor.execute(update_group_sql, atomic_view_group_ids)
        
        print(f"Group {atomic_view_group_ids} update success")


def list_data_sources(conn_cursor):
    # 获取数据源列表
    data_source_sql = """
    SELECT id, name FROM vega.data_source;
    """
    conn_cursor.execute(data_source_sql)
    data_source_results = conn_cursor.fetchall()

    # 遍历数据源
    data_source_map = {}
    for data_source in data_source_results:
        data_source_id = data_source[0]
        data_source_name = data_source[1]

        data_source_map[data_source_id] = data_source_name

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
            upgrade_data_view_group(conn_cursor)
    except Exception:
        raise Exception()
    finally:
        conn_cursor.close()
        conn.close()


