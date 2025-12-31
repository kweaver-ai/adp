# 指标模型 t_metric_model 去掉字段 f_data_view_id

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

def upgrade_metric_model_table_schema(conn_cursor):
    # 删除字段
    drop_column_sql = """
    ALTER TABLE t_metric_model DROP COLUMN f_data_view_id;
    """ 

    print(f"Execute sql: {drop_column_sql}")
    conn_cursor.execute(drop_column_sql)
    
    print(f"Execute sql: {drop_column_sql} success")

def check_metric_model_field(conn_cursor):
    """
    检查 t_metric_model 中的字段 f_data_view_id 是否存在
    返回 True 如果都存在，否则返回 False
    """
    try:
        show_columns_sql = "SHOW COLUMNS FROM t_metric_model;"
        conn_cursor.execute(show_columns_sql)
        columns = conn_cursor.fetchall()

        f_data_view_id_exists = False
        for col in columns:
            if col[0] == 'f_data_view_id':
                print("Found f_data_view_id in t_metric_model")
                return True

        if not f_data_view_id_exists:
            print("column f_data_view_id not found, skip upgrade")
            return False

    except Exception as e:
        print(f"Check f_data_view_id in t_metric_model error: {str(e)}")
        return False


if __name__ == "__main__":
    conn = get_conn(os.environ["DB_USER"], os.environ["DB_PASSWD"],
                    os.environ["DB_HOST"], os.environ["DB_PORT"], "dip_mdl")

    conn_cursor = conn.cursor()
    try:
        rslt = check_metric_model_field(conn_cursor)
        if rslt:
            upgrade_metric_model_table_schema(conn_cursor)
    except Exception:
        raise Exception()
    finally:
        conn_cursor.close()
        conn.close()


