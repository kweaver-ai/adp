# 对已存在的自建索引视图升级为自定义视图

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

def upgrade_custom_view(conn_cursor):
    # 获取查询类型为DSL的原子视图
    select_atomic_view_sql = """
    SELECT f_view_id, f_technical_name FROM t_data_view WHERE f_type = "atomic" AND f_query_type = "DSL";
    """

    conn_cursor.execute(select_atomic_view_sql)
    atomic_view_results = conn_cursor.fetchall()
    
    atomic_view_map = {}
    for atomic_view in atomic_view_results:
        atomic_view_id = atomic_view[0]
        atomic_view_technical_name = atomic_view[1]

        # view_technical_name 对应 base_type
        atomic_view_map[atomic_view_technical_name] = {
            "id": atomic_view_id,
            "technical_name": atomic_view_technical_name,
        }

    # 获取所有的字段模型的mappings
    field_model_map = get_field_model_mappings(conn_cursor)
    # 获取索引库和其字段的映射
    index_base_mappings = get_index_base_mappings(conn_cursor, field_model_map)

    # 获取所有自定义索引视图
    select_view_sql = """
    SELECT f_view_id, f_view_name, f_data_source, f_field_scope, f_fields, f_filters, f_loggroup_filters FROM t_data_view WHERE f_type != "atomic";
    """
    conn_cursor.execute(select_view_sql)
    results = conn_cursor.fetchall()
    print(f"Number of custom views to upgrade: {len(results)}")

    for res in results:
        view_id = res[0]
        view_name = res[1]
        view_data_source = res[2]
        view_field_scope = res[3]
        view_fields = res[4]
        view_filters = res[5]
        view_loggroup_filters = res[6]
        # 如果存在loggroup_filters，需要提醒脚本升级后用户需要手动升级loggroup_filters为每个视图的condition
        if view_loggroup_filters != "":
            print(f"View {view_name} has loggroup_filters, after upgrade it needs to be manually updated")
            print(f"View {view_name} loggroup_filters: {view_loggroup_filters}")

        if view_fields is not None:
            view_fields_list = json.loads(view_fields)
            # 给每个字段加上 original_name
            if isinstance(view_fields_list, list):
                for vf in view_fields_list:
                    if isinstance(vf, dict) and 'name' in vf:
                        vf["original_name"] = vf["name"]

        # 将 data_source的索引库转成data_scope的结构
        data_scope = []
        if view_data_source is None:
            print(f"View {view_name} data_source is None, skip upgrade")
            continue
        data_source = json.loads(view_data_source)
        index_bases_list = data_source["index_base"]
        view_base_types = []
        for idx, index_base in enumerate(index_bases_list, 1):
            base_type = index_base["base_type"]
            view_base_types.append(base_type)
            if base_type in index_base_mappings:
                index_base_field = index_base_mappings[base_type]
            else:
                raise ValueError(f"View {view_name} index base {base_type} does not have field mappings")

            if view_field_scope == 1:
                # 视图字段范围为1时，只有一个索引库，直接获取全部字段
                node_output_fields = index_base_field
            elif view_field_scope == 0:
                # 字段范围为部分字段，需要根据view_fields获取该索引库的部分字段
                # 将视图字段转为map
                view_fields_map = {vf["name"]: vf for vf in view_fields_list}
                node_output_fields = []
                for ibf in index_base_field:
                    field_name = ibf["name"]
                    if field_name in view_fields_map:
                        node_output_fields.append(ibf)
            else:
                raise ValueError(f"View {view_name} field scope unknown")
                
            if base_type in atomic_view_map:
                atomic_view_info = atomic_view_map[base_type]
            else:
                raise ValueError(f"View {view_name} index base {base_type} does not have atomic view")

            if view_filters is not None:
                ds_filters = json.loads(view_filters)
            else:
                ds_filters = None

            data_scope.append({
                "id": f"node{idx}",
                "title": "",
                "type": "view",
                "input_nodes": [],
                "config": {
                    "view_id": atomic_view_info["id"],
                    "filters": ds_filters,
                    "distinct": {"enable": False},
                },
                "output_fields": node_output_fields,
            })


        # 构造输出字段
        final_output_fields = []
        if view_field_scope == 1:
            if len(view_base_types) != 1:
                raise ValueError(f"View {view_name} field scope is 1, index base count must be 1, but actual is {len(view_base_types)}")
            index_base_type = view_base_types[0]
            if index_base_type in index_base_mappings:
                final_output_fields = index_base_mappings[index_base_type]
            else:
                raise ValueError(f"View {view_name} index base {index_base_type} does not have field mappings")
            
            view_fields_list = final_output_fields

        elif view_field_scope == 0:
            # 视图字段范围为0时,output字段即为存储的视图字段列表
            final_output_fields = view_fields_list
        else:
            raise ValueError(f"View {view_name} field scope unknown")

        if len(index_bases_list) > 1:
            # 添加union节点
            union_node = {
                "id": "node_union",
                "title": "合并视图",
                "type": "union",
                "input_nodes": [f"node{idx}" for idx in range(1, len(index_bases_list) + 1)],
                "config": {
                    "union_type": "all"
                },
                "output_fields": final_output_fields,
            }
            
            data_scope.append(union_node)
            final_input_nodes = ["node_union"]
        else:
            final_input_nodes = [f"node{idx}" for idx in range(1, len(index_bases_list) + 1)]
        
        # 添加输出节点
        data_scope.append({
            "id": "node_output",
            "title": "输出视图",
            "type": "output",
            "input_nodes": final_input_nodes,
            "config": {},
            "output_fields": final_output_fields,
        })

        data_scope_str = json.dumps(data_scope, ensure_ascii=False)
        view_fields_str = json.dumps(view_fields_list, ensure_ascii=False)
       
        print(f"Update custom view {view_name}")
         # 更新视图信息
        update_data_view_sql = """
        UPDATE t_data_view SET f_type='{}', f_query_type='{}', f_data_scope='{}', f_fields='{}' WHERE f_view_id='{}';
        """.format("custom", "DSL",  data_scope_str, view_fields_str, view_id)

        conn_cursor.execute(update_data_view_sql)
        print(f"Custom view {view_name} update success")


# 获取索引库和其字段的映射，返回的字段已经转成了视图格式
def get_index_base_mappings(conn_cursor, field_model_map):
    """
    获取索引库的字段列表
    """
    # 获取元子段mappings
    meta_mappings = get_meta_mappings()

    # 自动映射的字段
    select_index_base_sql = """
    SELECT f_base_type, f_data_type, f_mappings FROM t_index_base;
    """

    conn_cursor.execute(select_index_base_sql)
    index_base_results = conn_cursor.fetchall()

    base_type_to_fields = {}
    for index_base in index_base_results:
        base_type = index_base[0]
        data_type = index_base[1]
        dynamic_mappings = index_base[2]
        if dynamic_mappings is not None:
            dynamic_mappingsArr = json.loads(dynamic_mappings)
        else:
            dynamic_mappingsArr = []

        base_fields = []
        base_fieldsMap = {}
        # 将元字段转为视图字段格式
        print(f"Index base {base_type} meta field count: {len(meta_mappings)}")
        for meta_field in meta_mappings:
            meta_field_name = meta_field["field"]
            # 如果遇到重复的字段，以第一个为准
            if meta_field_name in base_fieldsMap:
                continue

            base_fieldsMap[meta_field_name] = True
            base_fields.append({
                "name": meta_field_name,
                "original_name": meta_field_name,
                "type": meta_field["type"],
                "display_name": meta_field["display_name"],
                "comment": ""
            })

        # 将字段模型转为视图字段格式
        if data_type in field_model_map:
            custom_mappings = field_model_map[data_type]
            print(f"Index base {base_type} field model field count: {len(custom_mappings)}")
            for custom_field in custom_mappings:
                custom_field_name = custom_field["field"]
                # 如果遇到重复的字段，以第一个为准
                if custom_field_name in base_fieldsMap:
                    continue

                base_fields.append({
                    "name": custom_field_name,
                    "original_name": custom_field_name,
                    "type": custom_field["type"],
                    "display_name": custom_field["display_name"],
                    "comment": ""
                })

        # 将索引库的字段转为视图字段格式
        print(f"Index base {base_type} dynamic field count: {len(dynamic_mappingsArr)}")
        for dynamic_field in dynamic_mappingsArr:
            dynamic_field_name = dynamic_field["field"]
            # 如果遇到重复的字段，以第一个为准
            if dynamic_field_name in base_fieldsMap:
                continue

            base_fields.append({
                "name": dynamic_field_name,
                "original_name": dynamic_field_name,
                "type": dynamic_field["type"],
                "display_name": dynamic_field["display_name"],
                "comment": ""
            })
     
        base_type_to_fields[base_type] = base_fields
        print(f"Index base {base_type} total field count: {len(base_fields)}")
        if len(meta_mappings) + len(custom_mappings) + len(dynamic_mappingsArr) != len(base_fields):
            print(f"Index base {base_type} field count not match, meta field count: {len(meta_mappings)}, field model field count: {len(custom_mappings)}, dynamic field count: {len(dynamic_mappingsArr)}, total field count: {len(base_fields)}")
   
    return base_type_to_fields


# 获取字段模型的字段
def get_field_model_mappings(conn_cursor):
    # 字段模型的字段
    select_field_model_sql = """
    SELECT f_data_type, f_mappings FROM t_field_model;
    """

    conn_cursor.execute(select_field_model_sql)
    field_model_results = conn_cursor.fetchall()
    field_model_map = {}
    for field_model in field_model_results:
        data_type = field_model[0]
        mappings = field_model[1]
        if mappings is not None:
            mappingsArr = json.loads(mappings)
            field_model_map[data_type] = mappingsArr
        else:
            mappingsArr = []
            field_model_map[data_type] = mappingsArr

    return field_model_map

def get_meta_mappings():
     # 元字段
    meta_mappings = [
        {
            "field": "@timestamp",
            "type": "date",
            "display_name": "@timestamp",
            "format": "epoch_millis||strict_date_optional_time||yyyy-MM-dd HH:mm:ss"
        },
        {
            "display_name": "__write_time",
            "format": "epoch_millis||strict_date_optional_time||yyyy-MM-dd HH:mm:ss",
            "field": "__write_time",
            "type": "date"
        },
        {
            "field": "__data_type",
            "type": "keyword",
            "display_name": "__data_type"
        },
        {
            "field": "__index_base",
            "type": "keyword",
            "display_name": "__index_base"
        },
        {
            "display_name": "__category",
            "field": "__category",
            "type": "keyword"
        },
        {
            "field": "__id",
            "type": "keyword",
            "display_name": "__id"
        },
        {
            "display_name": "__routing",
            "field": "__routing",
            "type": "keyword"
        },
        {
            "field": "__tsid",
            "type": "keyword",
            "display_name": "__tsid"
        },
        {
            "field": "__pipeline_id",
            "type": "keyword",
            "display_name": "__pipeline_id"
        },
        {
            "field": "tags",
            "type": "keyword",
            "display_name": "tags"
        }
    ]

    return meta_mappings


if __name__ == "__main__":
    if os.environ.get("CI_MODE") == "true":
        print("CI_MODE 为 true，跳过 当前升级文件")
    else:
        conn = get_conn(os.environ["DB_USER"], os.environ["DB_PASSWD"],
                        os.environ["DB_HOST"], os.environ["DB_PORT"], "dip_mdl")

        conn_cursor = conn.cursor()
        try:
            upgrade_custom_view(conn_cursor)
        except Exception as e:
            print(f"Upgrade failed: {str(e)}")
            raise  
        finally:
            conn_cursor.close()
            conn.close()

