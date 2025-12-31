#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import secrets
import string
import rdsdriver

ViewScanStatus_Delete   = "delete"

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


def upgrade_view_fields(conn_cursor):
    sql = """
    SELECT f_view_id, f_view_name, f_type, f_status FROM t_data_view WHERE f_type = "atomic" AND f_query_type = "SQL" AND f_status != "delete";
    """
    conn_cursor.execute(sql)
    results = conn_cursor.fetchall()

    for res in results:
        view_id = res[0] 
        view_name = res[1]
        view_type = res[2]
        view_status = res[3]

        # 生成6位安全的随机字符串（字母+数字）
        random_suffix = ''.join(secrets.choice(string.ascii_letters + string.digits) for _ in range(6))
        new_view_name = f"{view_name}_{random_suffix}"
        sql = """
        UPDATE t_data_view SET f_view_name='{}', f_status='{}' WHERE f_view_id='{}';
        """.format(new_view_name, ViewScanStatus_Delete, view_id)
        conn_cursor.execute(sql)
        
        print(f"{view_type} view '{view_id}' view_name '{view_name}' -> '{new_view_name}', view_status: {view_status} -> {ViewScanStatus_Delete}")

    print(f"Upgrade {len(results)} atomic SQL view completed.")


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

