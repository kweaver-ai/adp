#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import base64
import json
from Crypto.PublicKey import RSA
from Crypto.Cipher import PKCS1_v1_5
import rdsdriver

# RSA加密配置
RSA_PUBLIC_KEY = """-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA22GOSQ1jeDhpdzxhJddS
f+U10F4Ivut7giYhchFAIJgRonMamDT86MSqQUc8DdTFdPGLm7M3GUKcsG1qbC3S
qk4XJ9NjmQXbs7IMWyWEWQrN7Iv7S2QjDYJI+ppvIN03I0Km3WKsmnrle2bLzT/V
G8e72YX69dfXAeiX6uDhht1va/JxZVFMIV3pHa6AQQ9gn5SAUTX2akEhRfe1bPJj
fVyoM+dfNtvgdfaraqV1rOhVDEqd0NlOWt2RHwETQwU8gIJib2baj2MtyIAY+fQw
KlKWxUs1GcFbECnhVPiVN6BEhXD7OhRt9QE/cuYl5v4a6ypugGaMBK6VKOqFHDvf
mwIDAQAB
-----END PUBLIC KEY-----"""


def rsa_encrypt(data: str) -> str:
    """使用RSA 2048加密数据"""
    if not data:
        return ""
    key = RSA.import_key(RSA_PUBLIC_KEY)
    cipher = PKCS1_v1_5.new(key)
    encrypted = cipher.encrypt(data.encode('utf-8'))
    return base64.b64encode(encrypted).decode('utf-8')


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


def migrate_data_source(conn):
    """执行数据源迁移"""
    cursor = conn.cursor()

    try:
        # 查询所有需要迁移的数据源（排除特定ID）
        cursor.execute("""
            SELECT id, name, type_name, bin_data, comment, 
                   created_by_uid, created_at, updated_by_uid, updated_at
            FROM data_source
            WHERE id != 'cedb5294-07c3-45b1-a273-17baefa62800'
        """)

        for row in cursor.fetchall():
            # 解析bin_data中的字段
            bin_data = json.loads(row[3])

            # 处理密码字段（Base64解码后再RSA加密）
            password = ""
            if 'password' in bin_data and bin_data['password']:
                try:
                    decoded_password = base64.b64decode(bin_data['password']).decode('utf-8')
                    password = rsa_encrypt(decoded_password)
                except Exception as e:
                    print(f"Password encryption failed for {row[0]}: {str(e)}")
                    continue

            # 插入到t_data_source_info表
            cursor.execute("""
                INSERT INTO t_data_source_info (
                    f_id, f_name, f_type, f_catalog, f_database, f_schema,
                    f_connect_protocol, f_host, f_port, f_account, f_password,
                    f_storage_protocol, f_storage_base, f_token, f_replica_set,
                    f_is_built_in, f_comment, f_created_by_uid, f_created_at,
                    f_updated_by_uid, f_updated_at
                ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                ON DUPLICATE KEY UPDATE
                    f_name = VALUES(f_name),
                    f_type = VALUES(f_type),
                    f_catalog = VALUES(f_catalog),
                    f_database = VALUES(f_database),
                    f_schema = VALUES(f_schema),
                    f_connect_protocol = VALUES(f_connect_protocol),
                    f_host = VALUES(f_host),
                    f_port = VALUES(f_port),
                    f_account = VALUES(f_account),
                    f_password = VALUES(f_password),
                    f_storage_protocol = VALUES(f_storage_protocol),
                    f_storage_base = VALUES(f_storage_base),
                    f_token = VALUES(f_token),
                    f_replica_set = VALUES(f_replica_set),
                    f_comment = VALUES(f_comment),
                    f_updated_by_uid = VALUES(f_updated_by_uid),
                    f_updated_at = VALUES(f_updated_at)
            """, (
                row[0], row[1], row[2],
                bin_data.get('catalog_name'),
                bin_data.get('database_name'),
                bin_data.get('schema'),
                bin_data['connect_protocol'],
                bin_data['host'],
                bin_data['port'],
                bin_data.get('account'),
                password,
                bin_data.get('storage_protocol'),
                bin_data.get('storage_base'),
                bin_data.get('token'),
                bin_data.get('replica_set'),
                1,  # f_is_built_in设置为1
                row[4], row[5], row[6], row[7], row[8]
            ))

        conn.commit()
        print("数据迁移完成")
    except Exception as e:
        conn.rollback()
        raise Exception(f"数据迁移失败: {str(e)}")
    finally:
        cursor.close()


if __name__ == "__main__":
    try:
        # 从环境变量获取数据库连接信息
        conn = get_conn(os.environ["DB_USER"],
                        os.environ["DB_PASSWD"],
                        os.environ["DB_HOST"],
                        os.environ["DB_PORT"],
                        "vega")

        migrate_data_source(conn)
    except Exception as e:
        print(f"执行过程中发生错误: {str(e)}")
    finally:
        if 'conn' in locals() and conn:
            conn.close()
