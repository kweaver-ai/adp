SET SCHEMA dip_mdl;

CREATE TABLE IF NOT EXISTS t_metric_model (
  f_model_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_model_name VARCHAR(40 CHAR) NOT NULL,
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(255 CHAR) DEFAULT NULL,
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_measure_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_metric_type VARCHAR(20 CHAR) NOT NULL,
  f_query_type VARCHAR(20 CHAR) NOT NULL,
  f_formula text NOT NULL,
  f_date_field VARCHAR(255 CHAR) DEFAULT NULL,
  f_measure_field VARCHAR(255 CHAR) NOT NULL,
  f_unit_type VARCHAR(40 CHAR) NOT NULL,
  f_unit VARCHAR(20 CHAR) NOT NULL,
  f_group_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_builtin TINYINT DEFAULT 0,
  f_calendar_interval TINYINT DEFAULT 0,
  "f_data_source" VARCHAR(255 CHAR) NOT NULL,
    "f_formula_config" text DEFAULT NULL,
    "f_analysis_dimessions" VARCHAR(8192 CHAR) DEFAULT NULL,
    "f_creator" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    "f_catalog_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    "f_catalog_content" text DEFAULT NULL,
    CLUSTER PRIMARY KEY (f_model_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_metric_model_uk_model_name ON t_metric_model(f_group_id, f_model_name);

CREATE TABLE IF NOT EXISTS t_metric_model_group (
  f_group_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_group_name VARCHAR(40 CHAR) NOT NULL,
  f_comment VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_builtin TINYINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_group_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_metric_model_group_uk_f_group_name ON t_metric_model_group(f_group_name);

CREATE TABLE IF NOT EXISTS t_metric_model_task(
  f_task_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_task_name VARCHAR(40 CHAR) NOT NULL,
  f_comment VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_module_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_model_id VARCHAR(40 CHAR) NOT NULL,
  f_schedule VARCHAR(255 CHAR) NOT NULL,
  f_variables text DEFAULT NULL,
  f_time_windows VARCHAR(1024 CHAR) DEFAULT NULL,
  f_steps VARCHAR(255 CHAR) NOT NULL DEFAULT '[]',
  f_plan_time BIGINT NOT NULL DEFAULT 0,
  f_index_base VARCHAR(40 CHAR) NOT NULL,
  f_retrace_duration VARCHAR(20 CHAR) DEFAULT NULL,
  f_schedule_sync_status TINYINT NOT NULL,
  f_execute_status TINYINT DEFAULT 0,
  "f_creator" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    CLUSTER PRIMARY KEY (f_task_id)
);

CREATE TABLE IF NOT EXISTS t_static_metric_index (
  f_id INT IDENTITY(1, 1),
  f_base_type VARCHAR(40 CHAR) NOT NULL,
  f_split_time datetime(0) DEFAULT current_timestamp(),
  CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_static_metric_index_uk_f_index_base_type ON t_static_metric_index(f_base_type);

CREATE TABLE if not exists t_event_model_aggregate_rules (
  f_aggregate_rule_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_aggregate_rule_type VARCHAR(40 CHAR) NOT NULL,
  f_aggregate_algo VARCHAR(900 CHAR) NOT NULL,
  f_rule_priority INT NOT NULL,
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_group_fields VARCHAR(255 CHAR) DEFAULT '[]',
  f_aggregate_analysis_algo VARCHAR(1024 CHAR) DEFAULT '{}',
  CLUSTER PRIMARY KEY (f_aggregate_rule_id)
);

CREATE TABLE if not exists t_event_models (
  f_event_model_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_event_model_name VARCHAR(255 CHAR) NOT NULL,
  f_event_model_group_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_event_model_type VARCHAR(40 CHAR) NOT NULL,
  f_event_model_tags VARCHAR(255 CHAR) NOT NULL,
  f_event_model_comment VARCHAR(255 CHAR) DEFAULT NULL,
  f_data_source_type VARCHAR(40 CHAR) NOT NULL,
  f_data_source VARCHAR(900 CHAR) DEFAULT NULL,
  f_detect_rule_id VARCHAR(40 CHAR) NOT NULL,
  f_aggregate_rule_id VARCHAR(40 CHAR) NOT NULL,
  f_default_time_window VARCHAR(40 CHAR) NOT NULL,
  f_is_active TINYINT DEFAULT 0,
  f_enable_subscribe TINYINT DEFAULT 0,
  f_status TINYINT DEFAULT 0,
  f_downstream_dependent_model VARCHAR(1024 CHAR) DEFAULT '',
  f_is_custom TINYINT NOT NULL,
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  "f_creator" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    CLUSTER PRIMARY KEY (f_event_model_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_event_models_uk_f_model_name ON t_event_models(f_event_model_name);

CREATE TABLE if not exists t_event_model_detect_rules (
  f_detect_rule_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_detect_rule_type VARCHAR(40 CHAR) NOT NULL,
  f_formula VARCHAR(2014 CHAR) DEFAULT NULL,
  f_detect_algo VARCHAR(40 CHAR) DEFAULT NULL,
  f_detect_analysis_algo VARCHAR(1024 CHAR) DEFAULT '{}',
  f_rule_priority INT NOT NULL,
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_detect_rule_id)
);

CREATE TABLE IF NOT EXISTS t_event_model_task (
  f_task_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_model_id VARCHAR(40 CHAR) NOT NULL,
  f_storage_config VARCHAR(255 CHAR) NOT NULL,
  f_schedule VARCHAR(255 CHAR) NOT NULL,
  f_dispatch_config VARCHAR(255 CHAR) NOT NULL,
  f_execute_parameter VARCHAR(255 CHAR) NOT NULL,
  f_task_status TINYINT NOT NULL,
  f_error_details VARCHAR(2048 CHAR) NOT NULL,
  f_status_update_time BIGINT NOT NULL DEFAULT 0,
  f_schedule_sync_status TINYINT NOT NULL,
  f_downstream_dependent_task VARCHAR(1024 CHAR) DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  "f_creator" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    CLUSTER PRIMARY KEY (f_task_id)
);

CREATE TABLE IF NOT EXISTS t_event_model_task_execution_records (
  f_run_id BIGINT  NOT NULL,
  f_run_type VARCHAR(40 CHAR) NOT NULL,
  f_execute_parameter VARCHAR(2048 CHAR) NOT NULL,
  f_status VARCHAR(40 CHAR) DEFAULT '0',
  f_error_details VARCHAR(1024 CHAR) NOT NULL,
  f_update_time datetime(0) NOT NULL,
  f_create_time datetime(0) NOT NULL,
  CLUSTER PRIMARY KEY (f_run_id)
);

CREATE TABLE IF NOT EXISTS t_data_view (
  f_view_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_view_name VARCHAR(255 CHAR) NOT NULL,
  f_tags VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_comment VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_data_source text DEFAULT NULL,
  f_field_scope TINYINT NOT NULL DEFAULT '0',
  f_fields text NOT NULL,
  f_filters text DEFAULT NULL,
  f_builtin TINYINT DEFAULT 0,
  f_open_streaming TINYINT NOT NULL DEFAULT 0,
  f_job_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_loggroup_filters text DEFAULT NULL,
  f_group_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_creator" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    "f_technical_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
    "f_type"  VARCHAR(10 CHAR) NOT NULL DEFAULT '',
    "f_query_type" VARCHAR(10 CHAR) NOT NULL DEFAULT '',
    "f_data_source_type"  VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    "f_data_source_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    "f_file_name" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    "f_excel_config" text DEFAULT NULL,
    "f_data_scope" text DEFAULT NULL,
    "f_status" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    "f_metadata_form_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    "f_primary_keys" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
    "f_sql" text DEFAULT NULL,
    "f_updater" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    CLUSTER PRIMARY KEY (f_view_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_view_uk_f_view_name ON t_data_view(f_group_id, f_view_name);

CREATE TABLE IF NOT EXISTS t_data_view_group (
  f_group_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_group_name VARCHAR(40 CHAR) NOT NULL,
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_builtin TINYINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_group_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_view_group_uk_f_group_name ON t_data_view_group(f_group_name);

CREATE TABLE IF NOT EXISTS t_data_dict (
  f_dict_id VARCHAR(40 CHAR) NOT NULL,
  f_dict_name VARCHAR(255 CHAR) NOT NULL,
  f_tags VARCHAR(255 CHAR) NOT NULL,
  f_comment VARCHAR(255 CHAR) DEFAULT NULL,
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_dict_type VARCHAR(20 CHAR) NOT NULL DEFAULT 'kv_dict',
  f_dict_store VARCHAR(255 CHAR) NOT NULL,
  f_dimension VARCHAR(1500 CHAR) NOT NULL,
  f_unique_key TINYINT NOT NULL DEFAULT 1,
  "f_creator" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    CLUSTER PRIMARY KEY (f_dict_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_dict_uk_dict_name ON t_data_dict(f_dict_name);

CREATE TABLE IF NOT EXISTS t_data_dict_item (
  f_item_id VARCHAR(40 CHAR) NOT NULL,
  f_dict_id VARCHAR(40 CHAR) NOT NULL,
  f_item_key VARCHAR(3000 CHAR) NOT NULL,
  f_item_value VARCHAR(3000 CHAR) NOT NULL,
  f_comment VARCHAR(255 CHAR),
  CLUSTER PRIMARY KEY (f_item_id)
);

CREATE INDEX IF NOT EXISTS t_data_dict_item_idx_dict_id ON t_data_dict_item(f_dict_id);

CREATE TABLE IF NOT EXISTS t_data_connection (
  f_connection_id VARCHAR(40 CHAR) NOT NULL,
  f_connection_name VARCHAR(40 CHAR) NOT NULL,
  f_tags VARCHAR(255 CHAR) DEFAULT '',
  f_comment VARCHAR(255 CHAR) DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_data_source_type VARCHAR(40 CHAR) NOT NULL,
  f_config text NOT NULL,
  f_config_md5 VARCHAR(32 CHAR) DEFAULT '',
  CLUSTER PRIMARY KEY (f_connection_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_connection_uk_f_connection_name ON t_data_connection(f_connection_name);

CREATE INDEX IF NOT EXISTS t_data_connection_idx_f_data_source_type ON t_data_connection(f_data_source_type);

CREATE INDEX IF NOT EXISTS t_data_connection_idx_f_config_md5 ON t_data_connection(f_config_md5);

CREATE TABLE IF NOT EXISTS t_data_connection_status (
  f_connection_id VARCHAR(40 CHAR) NOT NULL,
  f_status VARCHAR(5 CHAR) NOT NULL,
  f_detection_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_connection_id)
);

CREATE TABLE IF NOT EXISTS t_trace_model (
  f_model_id VARCHAR(40 CHAR) NOT NULL,
  f_model_name VARCHAR(40 CHAR) NOT NULL,
  f_tags VARCHAR(255 CHAR) DEFAULT '',
  f_comment VARCHAR(255 CHAR) DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_span_source_type VARCHAR(40 CHAR) NOT NULL,
  f_span_config text NOT NULL,
  f_enabled_related_log TINYINT NOT NULL,
  f_related_log_source_type VARCHAR(40 CHAR) NOT NULL,
  f_related_log_config text NOT NULL,
  "f_creator" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    CLUSTER PRIMARY KEY (f_model_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_trace_model_uk_f_model_name ON t_trace_model(f_model_name);

CREATE INDEX IF NOT EXISTS t_trace_model_idx_f_span_source_type ON t_trace_model(f_span_source_type);

CREATE TABLE IF NOT EXISTS t_data_model_job (
  f_job_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_job_type VARCHAR(40 CHAR) NOT NULL,
  f_job_config text,
  f_job_status VARCHAR(20 CHAR) NOT NULL,
  f_job_status_details text NOT NULL,
  "f_creator" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    CLUSTER PRIMARY KEY (f_job_id)
);

CREATE TABLE IF NOT EXISTS t_objective_model (
  f_model_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_model_name VARCHAR(40 CHAR) NOT NULL,
  f_tags VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_comment VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_objective_type VARCHAR(20 CHAR) NOT NULL,
  f_objective_config text NOT NULL,
  "f_creator" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    CLUSTER PRIMARY KEY (f_model_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_objective_model_uk_t_objective_model ON t_objective_model(f_model_name);

CREATE TABLE IF NOT EXISTS t_scan_record (
    f_record_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_data_source_id VARCHAR(40 CHAR) NOT NULL,
    f_scanner VARCHAR(40 CHAR) NOT NULL,
    f_scan_time BIGINT NOT NULL DEFAULT 0,
    f_data_source_status VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_metadata_task_id VARCHAR(128 CHAR)  DEFAULT NULL,
    CLUSTER PRIMARY KEY (f_record_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_scan_record_uk_scan_record ON t_scan_record(f_data_source_id, f_scanner);

INSERT INTO t_data_view_group (

  f_group_id,

  f_group_name,

  f_create_time,

  f_update_time,

  f_builtin

)

SELECT '', '', 1733903782147, 1733903782147, 0

FROM DUAL

WHERE NOT EXISTS(

  SELECT f_group_id

  FROM t_data_view_group

  WHERE f_group_id = ''

);

INSERT INTO t_data_view_group (

  f_group_id,

  f_group_name,

  f_create_time,

  f_update_time,

  f_builtin

)

SELECT '__index_base', 'index_base', 1733903782147, 1733903782147, 1

FROM DUAL

WHERE NOT EXISTS(

  SELECT f_group_id

  FROM t_data_view_group

  WHERE f_group_id = '__index_base'

);

INSERT INTO t_metric_model_group (

  f_group_id,

  f_group_name,

  f_create_time,

  f_update_time,

  f_builtin

)

SELECT '', '', 1733903782147, 1733903782147, 0

FROM DUAL

WHERE NOT EXISTS(

  SELECT f_group_id

  FROM t_metric_model_group

  WHERE f_group_id = ''

);

