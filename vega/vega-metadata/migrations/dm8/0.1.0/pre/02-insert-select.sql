SET SCHEMA adp;
-- 迁移 t_data_quality_model 表数据
INSERT INTO adp.t_data_quality_model (
     "f_id", "f_ds_id", "f_dolphinscheduler_ds_id", "f_db_type", "f_tb_name", "f_process_definition_code", "f_crontab"
)
SELECT "f_id", "f_ds_id", "f_dolphinscheduler_ds_id", "f_db_type", "f_tb_name", "f_process_definition_code", "f_crontab"
  FROM vega.t_data_quality_model
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_data_quality_model t
  WHERE t.f_id = vega.t_data_quality_model.f_id
);

-- 迁移 t_data_quality_rule 表数据
INSERT INTO adp.t_data_quality_rule (
  "f_id", "f_field_name", "f_rule_id", "f_threshold", "f_check_val", "f_check_val_name", "f_model_id"
)
SELECT "f_id", "f_field_name", "f_rule_id", "f_threshold", "f_check_val", "f_check_val_name", "f_model_id"
  FROM vega.t_data_quality_rule
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_data_quality_rule t
  WHERE t.f_id = vega.t_data_quality_rule.f_id
);
-- 迁移 t_data_source 表数据
INSERT INTO adp.t_data_source (
 "f_id","f_name","f_data_source_type","f_data_source_type_name","f_user_name",
 "f_password","f_description","f_extend_property","f_host","f_port","f_enable_status",
 "f_connect_status","f_authority_id","f_create_time","f_create_user","f_update_time",
 "f_update_user","f_database","f_info_system_id","f_dolphin_id","f_delete_code","f_live_update_status","f_live_update_benchmark","f_live_update_time"
)
SELECT
  "f_id","f_name","f_data_source_type","f_data_source_type_name","f_user_name",
  "f_password","f_description","f_extend_property","f_host","f_port","f_enable_status",
  "f_connect_status","f_authority_id","f_create_time","f_create_user","f_update_time",
  "f_update_user","f_database","f_info_system_id","f_dolphin_id","f_delete_code","f_live_update_status","f_live_update_benchmark","f_live_update_time"
  FROM vega.t_data_source
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_data_source t
  WHERE t.f_id = vega.t_data_source.f_id
);
-- 迁移 t_dict 表数据
INSERT INTO adp.t_dict ("f_id", "f_dict_type", "f_dict_key", "f_dict_value", "f_extend_property", "f_enable_status")
  SELECT "f_id", "f_dict_type", "f_dict_key", "f_dict_value", "f_extend_property", "f_enable_status"
  FROM vega.t_dict
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_dict t
  WHERE t.f_id = vega.t_dict.f_id
);
-- 迁移 t_indicator 表数据
INSERT INTO adp.t_indicator ("f_id", "f_indicator_name", "f_indicator_type", "f_indicator_value", "f_create_time", "f_indicator_object_id", "f_authority_id", "f_advanced_params")
  SELECT "f_id", "f_indicator_name", "f_indicator_type", "f_indicator_value", "f_create_time", "f_indicator_object_id", "f_authority_id", "f_advanced_params"
  FROM vega.t_indicator
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_indicator t
  WHERE t.f_id = vega.t_indicator.f_id AND t.f_create_time = vega.t_indicator.f_create_time
);
-- 迁移 t_lineage_edge_column 表数据
INSERT INTO adp.t_lineage_edge_column ("f_id", "f_parent_id", "f_child_id", "f_create_type", "f_query_text", "created_at", "updated_at", "deleted_at", "f_create_time")
  SELECT "f_id", "f_parent_id", "f_child_id", "f_create_type", "f_query_text", "created_at", "updated_at", "deleted_at", "f_create_time"
  FROM vega.t_lineage_edge_column
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_lineage_edge_column t
  WHERE t.f_id = vega.t_lineage_edge_column.f_id
);
-- 迁移 t_lineage_edge_column_table_relation 表数据
INSERT INTO adp.t_lineage_edge_column_table_relation ("f_id", "f_table_id", "f_column_id", "created_at", "updated_at", "deleted_at")
  SELECT "f_id", "f_table_id", "f_column_id", "created_at", "updated_at", "deleted_at"
  FROM vega.t_lineage_edge_column_table_relation
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_lineage_edge_column_table_relation t
  WHERE t.f_id = vega.t_lineage_edge_column_table_relation.f_id
);
-- 迁移 t_lineage_edge_table 表数据
INSERT INTO adp.t_lineage_edge_table ("f_id", "f_parent_id", "f_child_id", "f_create_type", "f_query_text", "created_at", "updated_at", "deleted_at", "f_create_time")
  SELECT "f_id", "f_parent_id", "f_child_id", "f_create_type", "f_query_text", "created_at", "updated_at", "deleted_at", "f_create_time"
  FROM vega.t_lineage_edge_table
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_lineage_edge_table t
  WHERE t.f_id = vega.t_lineage_edge_table.f_id
);
-- 迁移 t_lineage_graph_info 表数据
INSERT INTO adp.t_lineage_graph_info ("app_id", "graph_id")
  SELECT "app_id", "graph_id"
  FROM vega.t_lineage_graph_info
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_lineage_graph_info t
  WHERE t.app_id = vega.t_lineage_graph_info.app_id
);
-- 迁移 t_lineage_log 表数据
INSERT INTO adp.t_lineage_log ("id", "class_id", "class_type", "action_type", "class_data", "created_at")
  SELECT "id", "class_id", "class_type", "action_type", "class_data", "created_at"
  FROM vega.t_lineage_log
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_lineage_log t
  WHERE t.id = vega.t_lineage_log.id
);
-- 迁移 t_lineage_relation 表数据
INSERT INTO adp.t_lineage_relation ("unique_id", "class_type", "parent", "child", "created_at", "updated_at")
  SELECT "unique_id", "class_type", "parent", "child", "created_at", "updated_at"
  FROM vega.t_lineage_relation
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_lineage_relation t
  WHERE t.unique_id = vega.t_lineage_relation.unique_id
);
-- 迁移 t_lineage_tag_column 表数据
INSERT INTO adp.t_lineage_tag_column ("f_id", "f_table_id", "f_column", "created_at", "updated_at", "deleted_at")
  SELECT "f_id", "f_table_id", "f_column", "created_at", "updated_at", "deleted_at"
  FROM vega.t_lineage_tag_column
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_lineage_tag_column t
  WHERE t.f_id = vega.t_lineage_tag_column.f_id
);
-- 迁移 t_lineage_tag_table 表数据
INSERT INTO adp.t_lineage_tag_table ("f_id", "f_db_type", "f_ds_id", "f_jdbc_url", "f_jdbc_user", "f_db_name", "f_db_schema", "f_tb_name", "created_at", "updated_at", "deleted_at")
  SELECT "f_id", "f_db_type", "f_ds_id", "f_jdbc_url", "f_jdbc_user", "f_db_name", "f_db_schema", "f_tb_name", "created_at", "updated_at", "deleted_at"
  FROM vega.t_lineage_tag_table
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_lineage_tag_table t
  WHERE t.f_id = vega.t_lineage_tag_table.f_id
);
-- 迁移 t_indicator2 表数据
INSERT INTO adp.t_indicator2 ("f_id", "f_indicator_name", "f_indicator_type", "f_indicator_value", "f_create_time", "f_indicator_object_id", "f_authority_id", "f_advanced_params")
  SELECT "f_id", "f_indicator_name", "f_indicator_type", "f_indicator_value", "f_create_time", "f_indicator_object_id", "f_authority_id", "f_advanced_params"
  FROM vega.t_indicator2
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_indicator2 t
  WHERE t.f_id = vega.t_indicator2.f_id AND t.f_create_time = vega.t_indicator2.f_create_time
);
-- 迁移 t_lineage_tag_column2 表数据
INSERT INTO adp.t_lineage_tag_column2 ("unique_id", "uuid", "technical_name", "business_name", "comment", "data_type", "primary_key", "table_unique_id", "expression_name", "column_unique_ids", "action_type", "created_at", "updated_at")
  SELECT "unique_id", "uuid", "technical_name", "business_name", "comment", "data_type", "primary_key", "table_unique_id", "expression_name", "column_unique_ids", "action_type", "created_at", "updated_at"
  FROM vega.t_lineage_tag_column2
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_lineage_tag_column2 t
  WHERE t.unique_id = vega.t_lineage_tag_column2.unique_id
);
-- 迁移 t_lineage_tag_indicator2 表数据
INSERT INTO adp.t_lineage_tag_indicator2 ("uuid", "name", "description", "code", "indicator_type","expression", "indicator_uuids", "time_restrict", "modifier_restrict","owner_uid", "owner_name", "department_id", "department_name", "column_unique_ids", "action_type", "created_at", "updated_at")
  SELECT "uuid", "name", "description", "code", "indicator_type",
       "expression", "indicator_uuids", "time_restrict", "modifier_restrict",
       "owner_uid", "owner_name", "department_id", "department_name", "column_unique_ids", "action_type", "created_at", "updated_at"
  FROM vega.t_lineage_tag_indicator2
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_lineage_tag_indicator2 t
  WHERE t.uuid = vega.t_lineage_tag_indicator2.uuid
);
-- 迁移 t_lineage_tag_table2 表数据
INSERT INTO adp.t_lineage_tag_table2 ("unique_id", "uuid", "technical_name", "business_name", "comment", "table_type", "datasource_id", "datasource_name",
	"owner_id", "owner_name", "department_id", "department_name", "info_system_id", "info_system_name", "database_name",
	"catalog_name", "catalog_addr", "catalog_type", "task_execution_info", "action_type", "created_at", "updated_at")
  SELECT "unique_id", "uuid", "technical_name", "business_name", "comment", "table_type", "datasource_id", "datasource_name",
       "owner_id", "owner_name", "department_id", "department_name", "info_system_id", "info_system_name", "database_name",
       "catalog_name", "catalog_addr", "catalog_type", "task_execution_info", "action_type", "created_at", "updated_at"
  FROM vega.t_lineage_tag_table2
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_lineage_tag_table2 t
  WHERE t.unique_id = vega.t_lineage_tag_table2.unique_id
);
-- 迁移 t_live_ddl 表数据
INSERT INTO adp.t_live_ddl ("f_id", "f_data_source_id", "f_data_source_name", "f_origin_catalog", "f_virtual_catalog", "f_schema_id", "f_schema_name", "f_table_id", "f_table_name", "f_sql_type",
  "f_sql_text", "f_live_update_benchmark", "f_monitor_time", "f_update_status", "f_update_message", "f_push_status")
  SELECT "f_id", "f_data_source_id", "f_data_source_name", "f_origin_catalog", "f_virtual_catalog", "f_schema_id", "f_schema_name", "f_table_id", "f_table_name", "f_sql_type",
  "f_sql_text", "f_live_update_benchmark", "f_monitor_time", "f_update_status", "f_update_message", "f_push_status"
  FROM vega.t_live_ddl
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_live_ddl t
  WHERE t.f_id = vega.t_live_ddl.f_id
);
-- 迁移 t_schema 表数据
INSERT INTO adp.t_schema ("f_id", "f_name", "f_data_source_id", "f_data_source_name", "f_data_source_type", "f_data_source_type_name", "f_authority_id", "f_create_time", "f_create_user", "f_update_time", "f_update_user")
  SELECT "f_id", "f_name", "f_data_source_id", "f_data_source_name", "f_data_source_type", "f_data_source_type_name", "f_authority_id", "f_create_time", "f_create_user", "f_update_time", "f_update_user"
  FROM vega.t_schema
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_schema t
  WHERE t.f_id = vega.t_schema.f_id
);
-- 迁移 t_table 表数据
INSERT INTO adp.t_table ("f_id", "f_name", "f_advanced_params", "f_description", "f_table_rows", "f_schema_id", "f_schema_name", "f_data_source_id", "f_data_source_name", "f_data_source_type", "f_data_source_type_name", "f_version", "f_authority_id", "f_create_time", "f_create_user", "f_update_time", "f_update_user", "f_delete_flag", "f_delete_time")
  SELECT "f_id", "f_name", "f_advanced_params", "f_description", "f_table_rows", "f_schema_id", "f_schema_name", "f_data_source_id", "f_data_source_name", "f_data_source_type", "f_data_source_type_name", "f_version", "f_authority_id", "f_create_time", "f_create_user", "f_update_time", "f_update_user", "f_delete_flag", "f_delete_time"
  FROM vega.t_table
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_table t
  WHERE t.f_id = vega.t_table.f_id
);
-- 迁移 t_table_field 表数据
INSERT INTO adp.t_table_field ("f_table_id", "f_field_name", "f_field_type", "f_field_length", "f_field_precision", "f_field_comment", "f_advanced_params", "f_update_flag", "f_update_time", "f_delete_flag", "f_delete_time")
  SELECT "f_table_id", "f_field_name", "f_field_type", "f_field_length", "f_field_precision", "f_field_comment", "f_advanced_params", "f_update_flag", "f_update_time", "f_delete_flag", "f_delete_time"
  FROM vega.t_table_field
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_table_field t
  WHERE t.f_table_id = vega.t_table_field.f_table_id AND t.f_field_name = vega.t_table_field.f_field_name
);
-- 迁移 t_table_field_his 表数据
INSERT INTO adp.t_table_field_his ("f_id", "f_field_name", "f_field_type", "f_field_length", "f_field_precision", "f_field_comment", "f_table_id", "f_version", "f_advanced_params")
  SELECT "f_id", "f_field_name", "f_field_type", "f_field_length", "f_field_precision", "f_field_comment", "f_table_id", "f_version", "f_advanced_params"
  FROM vega.t_table_field_his
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_table_field_his t
  WHERE t.f_id = vega.t_table_field_his.f_id AND t.f_version = vega.t_table_field_his.f_version
);
-- 迁移 t_task 表数据
INSERT INTO adp.t_task ("f_id", "f_object_id", "f_object_type", "f_name", "f_status", "f_start_time", "f_end_time", "f_create_user", "f_authority_id", "f_advanced_params")
  SELECT "f_id", "f_object_id", "f_object_type", "f_name", "f_status", "f_start_time", "f_end_time", "f_create_user", "f_authority_id", "f_advanced_params"
  FROM vega.t_task
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_task t
  WHERE t.f_id = vega.t_task.f_id
);
-- 迁移 t_task_log 表数据
INSERT INTO adp.t_task_log ("f_id", "f_task_id", "f_log", "f_authority_id")
  SELECT "f_id", "f_task_id", "f_log", "f_authority_id"
  FROM vega.t_task_log
  WHERE NOT EXISTS (
  SELECT 1 FROM adp.t_task_log t
  WHERE t.f_id = vega.t_task_log.f_id
);


DROP TABLE IF EXISTS vega.t_data_quality_model CASCADE;
DROP TABLE IF EXISTS vega.t_data_quality_rule CASCADE;
DROP TABLE IF EXISTS vega.t_data_source CASCADE;
DROP TABLE IF EXISTS vega.t_dict CASCADE;
DROP TABLE IF EXISTS vega.t_indicator CASCADE;
DROP TABLE IF EXISTS vega.t_lineage_edge_column CASCADE;
DROP TABLE IF EXISTS vega.t_lineage_edge_column_table_relation CASCADE;
DROP TABLE IF EXISTS vega.t_lineage_edge_table CASCADE;
DROP TABLE IF EXISTS vega.t_lineage_graph_info CASCADE;
DROP TABLE IF EXISTS vega.t_lineage_log CASCADE;
DROP TABLE IF EXISTS vega.t_lineage_relation CASCADE;
DROP TABLE IF EXISTS vega.t_lineage_tag_column CASCADE;
DROP TABLE IF EXISTS vega.t_lineage_tag_table CASCADE;
DROP TABLE IF EXISTS vega.t_indicator2 CASCADE;
DROP TABLE IF EXISTS vega.t_lineage_tag_column2 CASCADE;
DROP TABLE IF EXISTS vega.t_lineage_tag_indicator2 CASCADE;
DROP TABLE IF EXISTS vega.t_lineage_tag_table2 CASCADE;
DROP TABLE IF EXISTS vega.t_live_ddl CASCADE;
DROP TABLE IF EXISTS vega.t_schema CASCADE;
DROP TABLE IF EXISTS vega.t_table CASCADE;
DROP TABLE IF EXISTS vega.t_table_field CASCADE;
DROP TABLE IF EXISTS vega.t_table_field_his CASCADE;
DROP TABLE IF EXISTS vega.t_task CASCADE;
DROP TABLE IF EXISTS vega.t_task_log CASCADE;
