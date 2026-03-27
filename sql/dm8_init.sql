-- Source: dataflow/coderunner/migrations/dm8/0.1.0/init.sql
SET SCHEMA adp;

CREATE TABLE IF NOT EXISTS "t_python_package" (
  "f_id" VARCHAR(32 CHAR) NOT NULL,
  "f_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_oss_id" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_creator_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_creator_name" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  "f_created_at" BIGINT NOT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_python_package_uk_t_python_package_name ON t_python_package("f_name");


-- Source: dataflow/flow-automation/migrations/dm8/0.5.0/init.sql
SET SCHEMA adp;


CREATE TABLE IF NOT EXISTS "t_model" (
  "f_id" BIGINT  NOT NULL,
  "f_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_description" VARCHAR(300 CHAR) NOT NULL DEFAULT '',
  "f_train_status" VARCHAR(16 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL,
  "f_rule" TEXT DEFAULT NULL,
  "f_userid" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" TINYINT NOT NULL DEFAULT -1,
  "f_created_at" BIGINT DEFAULT NULL,
  "f_updated_at" BIGINT DEFAULT NULL,
  "f_scope" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_model_idx_t_model_f_name ON t_model(f_name);

CREATE INDEX IF NOT EXISTS t_model_idx_t_model_f_userid_status ON t_model(f_userid, f_status);

CREATE INDEX IF NOT EXISTS t_model_idx_t_model_f_status_type ON t_model(f_status, f_type);

CREATE TABLE IF NOT EXISTS "t_train_file" (
  "f_id" BIGINT  NOT NULL,
  "f_train_id" BIGINT  NOT NULL,
  "f_oss_id" VARCHAR(36 CHAR) DEFAULT '',
  "f_key" VARCHAR(36 CHAR) DEFAULT '',
  "f_created_at" BIGINT DEFAULT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_train_file_idx_t_train_file_f_train_id ON t_train_file(f_train_id);

CREATE TABLE IF NOT EXISTS "t_automation_executor" (
  "f_id" BIGINT  NOT NULL,
  "f_name" VARCHAR(256 CHAR) NOT NULL DEFAULT '',
  "f_description" VARCHAR(1024 CHAR) NOT NULL DEFAULT '',
  "f_creator_id" VARCHAR(40 CHAR) NOT NULL,
  "f_status" TINYINT NOT NULL,
  "f_created_at" BIGINT DEFAULT NULL,
  "f_updated_at" BIGINT DEFAULT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_automation_executor_idx_t_automation_executor_name ON t_automation_executor("f_name");

CREATE INDEX IF NOT EXISTS t_automation_executor_idx_t_automation_executor_creator_id ON t_automation_executor("f_creator_id");

CREATE INDEX IF NOT EXISTS t_automation_executor_idx_t_automation_executor_status ON t_automation_executor("f_status");

CREATE TABLE IF NOT EXISTS "t_automation_executor_accessor" (
  "f_id" BIGINT  NOT NULL,
  "f_executor_id" BIGINT  NOT NULL,
  "f_accessor_id" VARCHAR(40 CHAR) NOT NULL,
  "f_accessor_type" VARCHAR(20 CHAR) NOT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_automation_executor_accessor_idx_t_automation_executor_accessor ON t_automation_executor_accessor("f_executor_id", "f_accessor_id", "f_accessor_type");

CREATE UNIQUE INDEX IF NOT EXISTS t_automation_executor_accessor_uk_executor_accessor ON t_automation_executor_accessor("f_executor_id", "f_accessor_id", "f_accessor_type");

CREATE TABLE IF NOT EXISTS "t_automation_executor_action" (
  "f_id" BIGINT  NOT NULL,
  "f_executor_id" BIGINT  NOT NULL,
  "f_operator" VARCHAR(64 CHAR) NOT NULL,
  "f_name" VARCHAR(256 CHAR) NOT NULL DEFAULT '',
  "f_description" VARCHAR(1024 CHAR) NOT NULL DEFAULT '',
  "f_group" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(16 CHAR) NOT NULL DEFAULT 'python',
  "f_inputs" text,
  "f_outputs" text,
  "f_config" text,
  "f_created_at" BIGINT DEFAULT NULL,
  "f_updated_at" BIGINT DEFAULT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_automation_executor_action_idx_t_automation_executor_action_executor_id ON t_automation_executor_action("f_executor_id");

CREATE INDEX IF NOT EXISTS t_automation_executor_action_idx_t_automation_executor_action_operator ON t_automation_executor_action("f_operator");

CREATE INDEX IF NOT EXISTS t_automation_executor_action_idx_t_automation_executor_action_name ON t_automation_executor_action("f_name");

CREATE TABLE IF NOT EXISTS "t_content_admin" (
  "f_id" BIGINT  NOT NULL,
  "f_user_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_user_name" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_content_admin_uk_f_user_id ON t_content_admin("f_user_id");

CREATE TABLE IF NOT EXISTS "t_audio_segments" (
  "f_id" BIGINT  NOT NULL,
  "f_task_id" VARCHAR(32 CHAR) NOT NULL,
  "f_object" VARCHAR(1024 CHAR) NOT NULL,
  "f_summary_type" VARCHAR(12 CHAR) NOT NULL,
  "f_max_segments" TINYINT NOT NULL,
  "f_max_segments_type" VARCHAR(12 CHAR) NOT NULL,
  "f_need_abstract" TINYINT NOT NULL,
  "f_abstract_type" VARCHAR(12 CHAR) NOT NULL,
  "f_callback" VARCHAR(1024 CHAR) NOT NULL,
  "f_created_at" BIGINT DEFAULT NULL,
  "f_updated_at" BIGINT DEFAULT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE TABLE IF NOT EXISTS "t_automation_conf" (
  "f_key" VARCHAR(32 CHAR) NOT NULL,
  "f_value" VARCHAR(255 CHAR) NOT NULL,
  CLUSTER PRIMARY KEY ("f_key")
);

CREATE TABLE IF NOT EXISTS "t_automation_agent" (
  "f_id" BIGINT  NOT NULL,
  "f_name" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  "f_agent_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_version" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_automation_agent_idx_t_automation_agent_agent_id ON t_automation_agent("f_agent_id");

CREATE UNIQUE INDEX IF NOT EXISTS t_automation_agent_uk_t_automation_agent_name ON t_automation_agent("f_name");

CREATE TABLE IF NOT EXISTS "t_alarm_rule" (
  "f_id" BIGINT  NOT NULL,
  "f_rule_id" BIGINT  NOT NULL,
  "f_dag_id" BIGINT  NOT NULL,
  "f_frequency" SMALLINT  NOT NULL,
  "f_threshold" INT  NOT NULL,
  "f_created_at" BIGINT DEFAULT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_alarm_rule_idx_t_alarm_rule_rule_id ON t_alarm_rule("f_rule_id");

CREATE TABLE IF NOT EXISTS "t_alarm_user" (
  "f_id" BIGINT  NOT NULL,
  "f_rule_id" BIGINT  NOT NULL,
  "f_user_id" VARCHAR(36 CHAR) NOT NULL,
  "f_user_name" VARCHAR(128 CHAR) NOT NULL,
  "f_user_type" VARCHAR(10 CHAR) NOT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_alarm_user_idx_t_alarm_user_rule_id ON t_alarm_user("f_rule_id");

CREATE TABLE IF NOT EXISTS "t_automation_dag_instance_ext_data" (
    "f_id" VARCHAR(64 CHAR) NOT NULL,
    "f_created_at" BIGINT DEFAULT NULL,
    "f_updated_at" BIGINT DEFAULT NULL,
    "f_dag_id" VARCHAR(64 CHAR),
    "f_dag_ins_id" VARCHAR(64 CHAR),
    "f_field" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
    "f_oss_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
    "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
    "f_size" BIGINT  DEFAULT NULL,
    "f_removed" TINYINT NOT NULL DEFAULT 1,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_automation_dag_instance_ext_data_idx_t_automation_dag_instance_ext_data_dag_ins_id ON t_automation_dag_instance_ext_data("f_dag_ins_id");

CREATE TABLE IF NOT EXISTS "t_task_cache_0" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_0_uk_hash ON t_task_cache_0("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_0_idx_expire_time ON t_task_cache_0("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_1" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_1_uk_hash ON t_task_cache_1("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_1_idx_expire_time ON t_task_cache_1("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_2" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_2_uk_hash ON t_task_cache_2("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_2_idx_expire_time ON t_task_cache_2("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_3" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_3_uk_hash ON t_task_cache_3("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_3_idx_expire_time ON t_task_cache_3("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_4" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_4_uk_hash ON t_task_cache_4("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_4_idx_expire_time ON t_task_cache_4("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_5" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_5_uk_hash ON t_task_cache_5("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_5_idx_expire_time ON t_task_cache_5("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_6" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_6_uk_hash ON t_task_cache_6("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_6_idx_expire_time ON t_task_cache_6("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_7" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_7_uk_hash ON t_task_cache_7("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_7_idx_expire_time ON t_task_cache_7("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_8" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_8_uk_hash ON t_task_cache_8("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_8_idx_expire_time ON t_task_cache_8("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_9" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_9_uk_hash ON t_task_cache_9("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_9_idx_expire_time ON t_task_cache_9("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_a" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_a_uk_hash ON t_task_cache_a("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_a_idx_expire_time ON t_task_cache_a("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_b" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_b_uk_hash ON t_task_cache_b("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_b_idx_expire_time ON t_task_cache_b("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_c" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_c_uk_hash ON t_task_cache_c("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_c_idx_expire_time ON t_task_cache_c("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_d" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_d_uk_hash ON t_task_cache_d("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_d_idx_expire_time ON t_task_cache_d("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_e" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_e_uk_hash ON t_task_cache_e("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_e_idx_expire_time ON t_task_cache_e("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_f" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_f_uk_hash ON t_task_cache_f("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_f_idx_expire_time ON t_task_cache_f("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_dag_instance_event" (
  "f_id" BIGINT  NOT NULL,
  "f_type" TINYINT NOT NULL DEFAULT '0',
  "f_instance_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_operator" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  "f_task_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_status" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_name" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  "f_data" TEXT NOT NULL,
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_inline" TINYINT NOT NULL DEFAULT '0',
  "f_visibility" TINYINT NOT NULL DEFAULT '0',
  "f_timestamp" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_dag_instance_event_idx_instance_id ON t_dag_instance_event("f_instance_id", "f_id");

CREATE INDEX IF NOT EXISTS t_dag_instance_event_idx_instance_type_vis ON t_dag_instance_event("f_instance_id", "f_type", "f_visibility", "f_id");

CREATE INDEX IF NOT EXISTS t_dag_instance_event_idx_instance_name_type ON t_dag_instance_event("f_instance_id", "f_name", "f_type", "f_id");

INSERT INTO "t_automation_conf" (f_key, f_value) SELECT 'process_template', 1 FROM DUAL WHERE NOT EXISTS(SELECT "f_key", "f_value" FROM "t_automation_conf" WHERE "f_key"='process_template');

INSERT INTO "t_automation_conf" (f_key, f_value) SELECT 'ai_capabilities', 1 FROM DUAL WHERE NOT EXISTS(SELECT "f_key", "f_value" FROM "t_automation_conf" WHERE "f_key"='ai_capabilities');

CREATE TABLE IF NOT EXISTS "t_cron_job"
(
    "f_key_id" BIGINT NOT NULL IDENTITY(1, 1),
    "f_job_id" VARCHAR(36 CHAR) NOT NULL,
    "f_job_name" VARCHAR(64 CHAR) NOT NULL,
    "f_job_cron_time" VARCHAR(32 CHAR) NOT NULL,
    "f_job_type" TINYINT NOT NULL,
    "f_job_context" VARCHAR(10240 CHAR),
    "f_tenant_id" VARCHAR(36 CHAR),
    "f_enabled" TINYINT NOT NULL DEFAULT 1,
    "f_remarks" VARCHAR(256 CHAR),
    "f_create_time" BIGINT NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    CLUSTER PRIMARY KEY ("f_key_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_cron_job_index_job_id ON t_cron_job("f_job_id");
CREATE UNIQUE INDEX IF NOT EXISTS t_cron_job_index_job_name ON t_cron_job("f_job_name", "f_tenant_id");
CREATE INDEX IF NOT EXISTS t_cron_job_index_tenant_id ON t_cron_job("f_tenant_id");
CREATE INDEX IF NOT EXISTS t_cron_job_index_time ON t_cron_job("f_create_time", "f_update_time");



CREATE TABLE IF NOT EXISTS "t_cron_job_status"
(
    "f_key_id" BIGINT NOT NULL IDENTITY(1, 1),
    "f_execute_id" VARCHAR(36 CHAR) NOT NULL,
    "f_job_id" VARCHAR(36 CHAR) NOT NULL,
    "f_job_type" TINYINT NOT NULL,
    "f_job_name" VARCHAR(64 CHAR) NOT NULL,
    "f_job_status" TINYINT NOT NULL,
    "f_begin_time" BIGINT,
    "f_end_time" BIGINT,
    "f_executor" VARCHAR(1024 CHAR),
    "f_execute_times" INT,
    "f_ext_info" VARCHAR(1024 CHAR),
    CLUSTER PRIMARY KEY ("f_key_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_cron_job_status_index_execute_id ON t_cron_job_status("f_execute_id");
CREATE INDEX IF NOT EXISTS t_cron_job_status_index_job_id ON t_cron_job_status("f_job_id");
CREATE INDEX IF NOT EXISTS t_cron_job_status_index_job_status ON t_cron_job_status("f_job_status");
CREATE INDEX IF NOT EXISTS t_cron_job_status_index_time ON t_cron_job_status("f_begin_time","f_end_time");

CREATE TABLE IF NOT EXISTS "t_flow_dag" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_user_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
 "f_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_desc" VARCHAR(310 CHAR) NOT NULL DEFAULT '',
 "f_trigger" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
 "f_cron" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_vars" TEXT DEFAULT NULL,
 "f_status" VARCHAR(16 CHAR) NOT NULL DEFAULT '',
 "f_tasks" TEXT DEFAULT NULL,
 "f_steps" TEXT DEFAULT NULL,
 "f_description" VARCHAR(310 CHAR) NOT NULL DEFAULT '',
 "f_shortcuts" TEXT DEFAULT NULL,
 "f_accessors" TEXT DEFAULT NULL,
 "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
 "f_policy_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
 "f_appinfo" TEXT DEFAULT NULL,
 "f_priority" VARCHAR(16 CHAR) NOT NULL DEFAULT '',
 "f_removed" TINYINT NOT NULL DEFAULT 0,
 "f_emails" TEXT DEFAULT NULL,
 "f_template" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
 "f_published" TINYINT NOT NULL DEFAULT 0,
 "f_trigger_config" TEXT DEFAULT NULL,
 "f_sub_ids" TEXT DEFAULT NULL,
 "f_exec_mode" VARCHAR(8 CHAR) NOT NULL DEFAULT '',
 "f_category" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_outputs" TEXT DEFAULT NULL,
 "f_instructions" TEXT DEFAULT NULL,
 "f_operator_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
 "f_inc_values" VARCHAR(4096 CHAR) DEFAULT NULL,
 "f_version" VARCHAR(64 CHAR) DEFAULT NULL,
 "f_version_id" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
 "f_modify_by" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
 "f_is_debug" TINYINT NOT NULL DEFAULT 0,
 "f_debug_id" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
 "f_biz_domain_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_dag_user_id" ON "t_flow_dag" ("f_user_id");
CREATE INDEX IF NOT EXISTS "idx_dag_type" ON "t_flow_dag" ("f_type");
CREATE INDEX IF NOT EXISTS "idx_dag_trigger" ON "t_flow_dag" ("f_trigger");
CREATE INDEX IF NOT EXISTS "idx_dag_name" ON "t_flow_dag" ("f_name");
CREATE INDEX IF NOT EXISTS "idx_dag_biz_domain" ON "t_flow_dag" ("f_biz_domain_id");

CREATE TABLE IF NOT EXISTS "t_flow_dag_var" (
 "f_id" BIGINT NOT NULL,
 "f_dag_id" BIGINT NOT NULL DEFAULT 0,
 "f_var_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_default_value" TEXT DEFAULT NULL,
 "f_var_type" VARCHAR(16 CHAR) NOT NULL DEFAULT '',
 "f_description" TEXT DEFAULT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_dag_vars_dag_id" ON "t_flow_dag_var" ("f_dag_id");

CREATE TABLE IF NOT EXISTS "t_flow_dag_instance_keyword" (
 "f_id" BIGINT NOT NULL,
 "f_dag_ins_id" BIGINT NOT NULL DEFAULT 0,
 "f_keyword" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_dag_ins_kw" ON "t_flow_dag_instance_keyword" ("f_dag_ins_id", "f_keyword");

CREATE TABLE IF NOT EXISTS "t_flow_dag_step" (
 "f_id" BIGINT NOT NULL DEFAULT 0,
 "f_dag_id" BIGINT NOT NULL DEFAULT 0,
 "f_operator" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_source_id" TEXT NOT NULL,
 "f_has_datasource" TINYINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_dag_step_op" ON "t_flow_dag_step" ("f_operator");
CREATE INDEX IF NOT EXISTS "idx_dag_step_op_dag" ON "t_flow_dag_step" ("f_dag_id", "f_operator");
CREATE INDEX IF NOT EXISTS "idx_dag_step_has_ds_dag" ON "t_flow_dag_step" ("f_dag_id", "f_has_datasource");

CREATE TABLE IF NOT EXISTS "t_flow_dag_accessor" (
 "f_id" BIGINT NOT NULL,
 "f_dag_id" BIGINT NOT NULL DEFAULT 0,
 "f_accessor_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_dag_accessor_id_dag" ON "t_flow_dag_accessor" ("f_accessor_id", "f_dag_id");

CREATE TABLE IF NOT EXISTS "t_flow_dag_version" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_dag_id" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
 "f_user_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
 "f_version" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_version_id" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
 "f_change_log" VARCHAR(512 CHAR) DEFAULT NULL,
 "f_config" TEXT DEFAULT NULL,
 "f_sort_time" BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_dag_versions_dag_version" ON "t_flow_dag_version" ("f_version_id", "f_dag_id");
CREATE INDEX IF NOT EXISTS "idx_dag_versions_dag_sort" ON "t_flow_dag_version" ("f_dag_id", "f_sort_time");

CREATE TABLE IF NOT EXISTS "t_flow_dag_instance" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_dag_id" BIGINT NOT NULL DEFAULT 0,
 "f_trigger" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
 "f_worker" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
 "f_source" TEXT DEFAULT NULL,
 "f_vars" TEXT DEFAULT NULL,
 "f_keywords" TEXT DEFAULT NULL,
 "f_event_persistence" TINYINT NOT NULL DEFAULT 0,
 "f_event_oss_path" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_share_data" TEXT DEFAULT NULL,
 "f_share_data_ext" TEXT DEFAULT NULL,
 "f_status" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
 "f_reason" TEXT DEFAULT NULL,
 "f_cmd" TEXT DEFAULT NULL,
 "f_has_cmd" TINYINT NOT NULL DEFAULT 0,
 "f_batch_run_id" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
 "f_user_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
 "f_ended_at" BIGINT NOT NULL DEFAULT 0,
 "f_dag_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
 "f_policy_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
 "f_appinfo" TEXT DEFAULT NULL,
 "f_priority" VARCHAR(16 CHAR) NOT NULL DEFAULT '',
 "f_mode" TINYINT NOT NULL DEFAULT 0,
 "f_dump" TEXT DEFAULT NULL,
 "f_dump_ext" TEXT DEFAULT NULL,
 "f_success_callback" VARCHAR(1024 CHAR) DEFAULT NULL,
 "f_error_callback" VARCHAR(1024 CHAR) DEFAULT NULL,
 "f_call_chain" TEXT DEFAULT NULL,
 "f_resume_data" TEXT DEFAULT NULL,
 "f_resume_status" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_version" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_version_id" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
 "f_biz_domain_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_dag_ins_dag_status" ON "t_flow_dag_instance" ("f_dag_id", "f_status");
CREATE INDEX IF NOT EXISTS "idx_dag_ins_status_upd" ON "t_flow_dag_instance" ("f_status", "f_updated_at");
CREATE INDEX IF NOT EXISTS "idx_dag_ins_status_user_pri" ON "t_flow_dag_instance" ("f_status", "f_user_id", "f_priority");
CREATE INDEX IF NOT EXISTS "idx_dag_ins_user_id" ON "t_flow_dag_instance" ("f_user_id");
CREATE INDEX IF NOT EXISTS "idx_dag_ins_batch_run" ON "t_flow_dag_instance" ("f_batch_run_id");
CREATE INDEX IF NOT EXISTS "idx_dag_ins_worker" ON "t_flow_dag_instance" ("f_worker");

CREATE TABLE IF NOT EXISTS "t_flow_inbox" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_msg" TEXT DEFAULT NULL,
 "f_topic" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
 "f_docid" VARCHAR(512 CHAR) NOT NULL DEFAULT '',
 "f_dag" TEXT DEFAULT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_inbox_docid" ON "t_flow_inbox" ("f_docid");
CREATE INDEX IF NOT EXISTS "idx_inbox_topic_created" ON "t_flow_inbox" ("f_topic", "f_created_at");

CREATE TABLE IF NOT EXISTS "t_flow_outbox" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_msg" TEXT DEFAULT NULL,
 "f_topic" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_outbox_created" ON "t_flow_outbox" ("f_created_at");

CREATE TABLE IF NOT EXISTS "t_flow_task_instance" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_expired_at" BIGINT NOT NULL DEFAULT 0,
 "f_task_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_dag_ins_id" BIGINT NOT NULL DEFAULT 0,
 "f_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_depend_on" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_action_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_timeout_secs" BIGINT NOT NULL DEFAULT 0,
 "f_params" TEXT DEFAULT NULL,
 "f_traces" TEXT DEFAULT NULL,
 "f_status" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
 "f_reason" TEXT DEFAULT NULL,
 "f_pre_checks" TEXT DEFAULT NULL,
 "f_results" TEXT DEFAULT NULL,
 "f_steps" TEXT DEFAULT NULL,
 "f_last_modified_at" BIGINT NOT NULL DEFAULT 0,
 "f_rendered_params" TEXT DEFAULT NULL,
 "f_hash" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_settings" TEXT DEFAULT NULL,
 "f_metadata" TEXT DEFAULT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_task_ins_dag_ins_id" ON "t_flow_task_instance" ("f_dag_ins_id");
CREATE INDEX IF NOT EXISTS "idx_task_ins_hash" ON "t_flow_task_instance" ("f_hash");
CREATE INDEX IF NOT EXISTS "idx_task_ins_action" ON "t_flow_task_instance" ("f_action_name");
CREATE INDEX IF NOT EXISTS "idx_task_ins_status_expire" ON "t_flow_task_instance" ("f_status", "f_expired_at");
CREATE INDEX IF NOT EXISTS "idx_task_ins_status_upd_id" ON "t_flow_task_instance" ("f_status", "f_updated_at", "f_id");

CREATE TABLE IF NOT EXISTS "t_flow_token" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_user_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
 "f_user_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_refresh_token" TEXT DEFAULT NULL,
 "f_token" TEXT DEFAULT NULL,
 "f_expires_in" INT NOT NULL DEFAULT 0,
 "f_login_ip" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_is_app" TINYINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_token_user_id" ON "t_flow_token" ("f_user_id");

CREATE TABLE IF NOT EXISTS "t_flow_client" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_client_name" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_client_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
 "f_client_secret" VARCHAR(16 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_client_name" ON "t_flow_client" ("f_client_name");

CREATE TABLE IF NOT EXISTS "t_flow_switch" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_status" TINYINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_switch_name" ON "t_flow_switch" ("f_name");

CREATE TABLE IF NOT EXISTS "t_flow_log" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_ossid" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_key" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
 "f_filename" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE TABLE IF NOT EXISTS "t_flow_storage" (
  "f_id" BIGINT NOT NULL,
  "f_oss_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_object_key" VARCHAR(512 CHAR) NOT NULL DEFAULT '',
  "f_name" VARCHAR(256 CHAR) NOT NULL DEFAULT '',
  "f_content_type" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT 0,
  "f_etag" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT 1,
  "f_created_at" BIGINT NOT NULL DEFAULT 0,
  "f_updated_at" BIGINT NOT NULL DEFAULT 0,
  "f_deleted_at" BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS "uk_flow_storage_oss_id_object_key" ON "t_flow_storage" ("f_oss_id", "f_object_key");
CREATE INDEX IF NOT EXISTS "idx_flow_storage_status" ON "t_flow_storage" ("f_status");
CREATE INDEX IF NOT EXISTS "idx_flow_storage_created_at" ON "t_flow_storage" ("f_created_at");

CREATE TABLE IF NOT EXISTS "t_flow_file" (
  "f_id" BIGINT NOT NULL,
  "f_dag_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_dag_instance_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_storage_id" BIGINT NOT NULL DEFAULT 0,
  "f_status" TINYINT NOT NULL DEFAULT 1,
  "f_name" VARCHAR(256 CHAR) NOT NULL DEFAULT '',
  "f_expires_at" BIGINT NOT NULL DEFAULT 0,
  "f_created_at" BIGINT NOT NULL DEFAULT 0,
  "f_updated_at" BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_flow_file_dag_id" ON "t_flow_file" ("f_dag_id");
CREATE INDEX IF NOT EXISTS "idx_flow_file_dag_instance_id" ON "t_flow_file" ("f_dag_instance_id");
CREATE INDEX IF NOT EXISTS "idx_flow_file_storage_id" ON "t_flow_file" ("f_storage_id");
CREATE INDEX IF NOT EXISTS "idx_flow_file_status" ON "t_flow_file" ("f_status");
CREATE INDEX IF NOT EXISTS "idx_flow_file_expires_at" ON "t_flow_file" ("f_expires_at");

CREATE TABLE IF NOT EXISTS "t_flow_file_download_job" (
  "f_id" BIGINT NOT NULL,
  "f_file_id" BIGINT NOT NULL,
  "f_status" TINYINT NOT NULL DEFAULT 1,
  "f_retry_count" INT NOT NULL DEFAULT 0,
  "f_max_retry" INT NOT NULL DEFAULT 3,
  "f_next_retry_at" BIGINT NOT NULL DEFAULT 0,
  "f_error_code" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_error_message" VARCHAR(1024 CHAR) NOT NULL DEFAULT '',
  "f_download_url" VARCHAR(2048 CHAR) NOT NULL DEFAULT '',
  "f_started_at" BIGINT NOT NULL DEFAULT 0,
  "f_finished_at" BIGINT NOT NULL DEFAULT 0,
  "f_created_at" BIGINT NOT NULL DEFAULT 0,
  "f_updated_at" BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS "uk_flow_file_download_job_file_id" ON "t_flow_file_download_job" ("f_file_id");
CREATE INDEX IF NOT EXISTS "idx_flow_file_download_job_status_retry" ON "t_flow_file_download_job" ("f_status", "f_next_retry_at");

CREATE TABLE IF NOT EXISTS "t_flow_task_resume" (
  "f_id" BIGINT NOT NULL,
  "f_task_instance_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_dag_instance_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_resource_type" VARCHAR(32 CHAR) NOT NULL DEFAULT 'file',
  "f_resource_id" BIGINT NOT NULL DEFAULT 0,
  "f_created_at" BIGINT NOT NULL DEFAULT 0,
  "f_updated_at" BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS "uk_flow_task_resume_task_instance_id" ON "t_flow_task_resume" ("f_task_instance_id");
CREATE INDEX IF NOT EXISTS "idx_flow_task_resume_resource" ON "t_flow_task_resume" ("f_resource_type", "f_resource_id");
-- Source: dataflow/flow-stream-data-pipeline/migrations/dm8/0.1.0/init.sql
SET SCHEMA adp;

CREATE TABLE IF NOT EXISTS t_internal_app (
  f_app_id VARCHAR(40 CHAR) NOT NULL,
  f_app_name VARCHAR(40 CHAR) NOT NULL,
  f_app_secret VARCHAR(40 CHAR) NOT NULL,
  f_create_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_app_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_internal_app_uk_app_name ON t_internal_app(f_app_name);

CREATE TABLE IF NOT EXISTS t_stream_data_pipeline (
  f_pipeline_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_pipeline_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) NOT NULL,
  f_comment VARCHAR(255 CHAR),
  f_builtin TINYINT DEFAULT 0,
  f_output_type VARCHAR(20 CHAR) NOT NULL,
  f_index_base VARCHAR(255 CHAR) NOT NULL,
  f_use_index_base_in_data TINYINT DEFAULT 0,
  f_pipeline_status VARCHAR(10 CHAR) NOT NULL,
  f_pipeline_status_details text NOT NULL,
  f_deployment_config text NOT NULL,
  f_create_time BIGINT NOT NULL default 0,
  f_update_time BIGINT NOT NULL default 0,
  "f_creator" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    "f_updater" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    "f_creator_type" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    "f_updater_type" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    CLUSTER PRIMARY KEY (f_pipeline_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_stream_data_pipeline_uk_name ON t_stream_data_pipeline(f_pipeline_name);


-- Source: bkn/bkn-backend/migrations/dm8/0.5.0/init.sql
-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

SET SCHEMA adp;


-- 业务知识网络
CREATE TABLE IF NOT EXISTS t_knowledge_network (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT DEFAULT NULL,
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_business_domain VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_id,f_branch)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_t_knowledge_network_kn_name ON t_knowledge_network(f_name,f_branch);


-- 对象类
CREATE TABLE IF NOT EXISTS t_object_type (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT DEFAULT NULL,
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_data_source VARCHAR(255 CHAR) NOT NULL,
  f_data_properties TEXT DEFAULT NULL,
  f_logic_properties TEXT DEFAULT NULL,
  f_primary_keys VARCHAR(8192 CHAR) DEFAULT NULL,
  f_display_key VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_incremental_key VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id,f_branch,f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_t_object_type_ot_name ON t_object_type(f_kn_id,f_branch,f_name);


-- 对象类状态
CREATE TABLE IF NOT EXISTS t_object_type_status (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_incremental_key VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_incremental_value VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_index VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_index_available BIT NOT NULL DEFAULT 0,
  f_doc_count BIGINT NOT NULL DEFAULT 0,
  f_storage_size BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id,f_branch,f_id)
);


-- 关系类
CREATE TABLE IF NOT EXISTS t_relation_type (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT DEFAULT NULL,
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_source_object_type_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_target_object_type_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_mapping_rules text DEFAULT NULL,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id,f_branch,f_id)
);


-- 行动类
CREATE TABLE IF NOT EXISTS t_action_type (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT DEFAULT NULL,
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_action_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_object_type_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_condition TEXT DEFAULT NULL,
  f_affect TEXT DEFAULT NULL,
  f_action_source VARCHAR(255 CHAR) NOT NULL,
  f_parameters TEXT DEFAULT NULL,
  f_schedule VARCHAR(255 CHAR) DEFAULT NULL,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id,f_branch,f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_t_action_type_at_name ON t_action_type(f_kn_id,f_branch,f_name);


-- 任务管理
CREATE TABLE IF NOT EXISTS t_kn_job (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_job_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_job_concept_config TEXT DEFAULT NULL,
  f_state VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_state_detail TEXT DEFAULT NULL,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_finish_time BIGINT NOT NULL DEFAULT 0,
  f_time_cost BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_id)
);


-- 子任务管理
CREATE TABLE IF NOT EXISTS t_kn_task (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_job_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_concept_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_concept_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_index VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_doc_count BIGINT NOT NULL DEFAULT 0,
  f_state VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_state_detail TEXT DEFAULT NULL,
  f_start_time BIGINT NOT NULL DEFAULT 0,
  f_finish_time BIGINT NOT NULL DEFAULT 0,
  f_time_cost BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_id)
);


-- 概念分组
CREATE TABLE IF NOT EXISTS t_concept_group (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT DEFAULT NULL,
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id,f_branch,f_id)
) ;

CREATE UNIQUE INDEX IF NOT EXISTS uk_concept_group_name ON t_concept_group(f_kn_id,f_branch,f_name);


-- 分组与概念对应表
CREATE TABLE IF NOT EXISTS t_concept_group_relation (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_group_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_concept_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_concept_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_id)
) ;

CREATE UNIQUE INDEX IF NOT EXISTS uk_concept_group_relation ON t_concept_group_relation(f_kn_id,f_branch,f_group_id,f_concept_type,f_concept_id);


-- Action Schedule Management
-- Supports cron-based scheduled action execution with distributed locking
CREATE TABLE IF NOT EXISTS t_action_schedule (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(100 CHAR) NOT NULL DEFAULT '',
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_action_type_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_cron_expression VARCHAR(100 CHAR) NOT NULL DEFAULT '',
  f_instance_identities TEXT DEFAULT NULL,
  f_dynamic_params TEXT DEFAULT NULL,
  f_status VARCHAR(20 CHAR) NOT NULL DEFAULT 'inactive',
  f_last_run_time BIGINT NOT NULL DEFAULT 0,
  f_next_run_time BIGINT NOT NULL DEFAULT 0,
  f_lock_holder VARCHAR(64 CHAR) DEFAULT NULL,
  f_lock_time BIGINT NOT NULL DEFAULT 0,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_id)
);

CREATE INDEX IF NOT EXISTS idx_action_schedule_kn_branch ON t_action_schedule(f_kn_id, f_branch);
CREATE INDEX IF NOT EXISTS idx_action_schedule_status_next_run ON t_action_schedule(f_status, f_next_run_time);
CREATE INDEX IF NOT EXISTS idx_action_schedule_action_type ON t_action_schedule(f_action_type_id);


-- Risk Type
CREATE TABLE IF NOT EXISTS t_risk_type (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id, f_branch, f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_risk_type_name ON t_risk_type(f_kn_id, f_branch, f_name);

-- Source: vega/vega-backend/migrations/dm8/0.5.0/init.sql
-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- VEGA Catalog 表结构定义
-- ==========================================

-- ==========================================
-- Schema定义说明（f_schema_definition字段JSON格式）
-- ==========================================
-- f_schema_definition 字段使用JSON数组格式存储所有字段信息，每个字段包含以下属性：
--
-- 基础属性：
--   - name: 字段名称
--   - type: VEGA统一类型 (integer, unsigned_integer, float, decimal, string, text, date, datetime, time, boolean, binary, json, vector)
--   - description: 字段描述
--   - type_config: 类型配置对象 (如 {"max_length": 128}, {"dimension": 768})
--
-- 源端映射：
--   - source_name: 源端字段名（可能与name不同）
--   - source_type: 源端字段类型
--   - is_native: 是否为系统自动同步的字段
--
-- 字段属性：
--   - is_primary: 是否为主键
--   - is_nullable: 是否可为空
--   - default_value: 默认值
--   - ordinal_position: 字段顺序位置
--
-- 字段特征（features数组，可选，用于扩展字段能力）：
--   - feature_type: 特征类型 (keyword, fulltext, vector)
--   - feature_config: 特征配置对象 (如分词器、向量空间类型等)
--   - ref_field_name: 引用的字段名称（用于借用其他字段的能力）
--   - enabled: 是否启用
--
-- 示例：
-- [
--   {
--     "name": "id",
--     "type": "integer",
--     "description": "主键ID",
--     "type_config": {"length": 11},
--     "source_name": "id",
--     "source_type": "int(11)",
--     "is_native": true,
--     "is_primary": true,
--     "is_nullable": false,
--     "default_value": "",
--     "ordinal_position": 1
--   },
--   {
--     "name": "content",
--     "type": "text",
--     "description": "文章内容",
--     "type_config": {},
--     "source_name": "content",
--     "source_type": "text",
--     "is_native": true,
--     "is_primary": false,
--     "is_nullable": true,
--     "default_value": "",
--     "ordinal_position": 2,
--     "features": [
--       {
--         "feature_type": "fulltext",
--         "feature_config": {"analyzer": "ik_max_word"},
--         "ref_field_name": "",
--         "is_default": true
--       }
--     ]
--   },
--   {
--     "name": "embedding",
--     "type": "vector",
--     "description": "向量嵌入",
--     "type_config": {"dimension": 768},
--     "source_name": "",
--     "source_type": "",
--     "is_native": false,
--     "is_primary": false,
--     "is_nullable": true,
--     "default_value": "",
--     "ordinal_position": 3,
--     "features": [
--       {
--         "feature_type": "vector",
--         "feature_config": {"space_type": "cosinesimil", "m": 16, "ef_construction": 200},
--         "ref_field_name": "",
--         "is_default": true
--       }
--     ]
--   }
-- ]
-- ==========================================
SET SCHEMA adp;
-- ==========================================
-- 1. t_catalog 主表
-- ==========================================
CREATE TABLE IF NOT EXISTS t_catalog (
    -- 主键与基础信息
    f_id                      VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_name                    VARCHAR(255 CHAR) NOT NULL DEFAULT '',
    f_tags                    VARCHAR(255 CHAR) NOT NULL DEFAULT '[]',
    f_description             VARCHAR(1000 CHAR) NOT NULL DEFAULT '',

    f_type                    VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_enabled                 TINYINT NOT NULL DEFAULT 1,

    -- Physical Catalog 专属字段
    f_connector_type          VARCHAR(50 CHAR) NOT NULL DEFAULT '',
    f_connector_config        TEXT NOT NULL,
    f_metadata                TEXT NOT NULL,

    -- 状态管理
    f_health_check_enabled    TINYINT NOT NULL DEFAULT 1,
    f_health_check_status     VARCHAR(20 CHAR) NOT NULL DEFAULT 'healthy',
    f_last_check_time         BIGINT NOT NULL DEFAULT 0,
    f_health_check_result     TEXT NOT NULL,

    -- 审计字段
    f_creator                 VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    f_creator_type            VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_create_time             BIGINT NOT NULL DEFAULT 0,
    f_updater                 VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    f_updater_type            VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_update_time             BIGINT NOT NULL DEFAULT 0,

    -- 索引
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_t_catalog_name ON t_catalog(f_name);

CREATE INDEX IF NOT EXISTS idx_t_catalog_type ON t_catalog(f_type);

CREATE INDEX IF NOT EXISTS idx_t_catalog_connector_type ON t_catalog(f_connector_type);

CREATE INDEX IF NOT EXISTS idx_t_catalog_health_check_status ON t_catalog(f_health_check_status);

-- ==========================================
-- 2. t_catalog_discover_policy 发现与变更策略表
-- ==========================================
CREATE TABLE IF NOT EXISTS t_catalog_discover_policy (
    f_id                      VARCHAR(40 CHAR) NOT NULL DEFAULT '',

    -- 状态
    f_enabled                 TINYINT NOT NULL DEFAULT 0,

    -- 发现策略配置
    f_discover_mode          VARCHAR(20 CHAR) NOT NULL DEFAULT 'manual',
    f_discover_cron          VARCHAR(100 CHAR) NOT NULL DEFAULT '',
    f_discover_config        TEXT NOT NULL,

    -- 变更处理策略
    f_on_resource_added       VARCHAR(20 CHAR) NOT NULL DEFAULT 'auto_register',
    f_on_resource_removed     VARCHAR(20 CHAR) NOT NULL DEFAULT 'mark_stale',
    f_on_schema_changed       VARCHAR(20 CHAR) NOT NULL DEFAULT 'auto_update',
    f_on_file_content_changed VARCHAR(20 CHAR) NOT NULL DEFAULT 'pending_review',

    -- 策略详细配置
    f_change_policy_config    TEXT NOT NULL,

    -- 索引
    CLUSTER PRIMARY KEY (f_id)
);

CREATE INDEX IF NOT EXISTS idx_t_catalog_discover_policy_discover_mode ON t_catalog_discover_policy(f_discover_mode);

CREATE INDEX IF NOT EXISTS idx_t_catalog_discover_policy_enabled ON t_catalog_discover_policy(f_enabled);

-- ==========================================
-- 3. t_resource 数据资源主表
-- ==========================================
CREATE TABLE IF NOT EXISTS t_resource (
    -- 主键与基础信息
    f_id                      VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_catalog_id              VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_name                    VARCHAR(255 CHAR) NOT NULL DEFAULT '',
    f_tags                    VARCHAR(255 CHAR) NOT NULL DEFAULT '[]',
    f_description             VARCHAR(1000 CHAR) NOT NULL DEFAULT '',

    f_category                VARCHAR(20 CHAR) NOT NULL DEFAULT '',

    -- 状态管理
    f_status                  VARCHAR(20 CHAR) NOT NULL DEFAULT 'active',
    f_status_message          VARCHAR(500 CHAR) NOT NULL DEFAULT '',

    -- 物理数据资源专属字段
    f_database                VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    f_source_identifier       VARCHAR(500 CHAR) NOT NULL DEFAULT '',
    f_source_metadata         TEXT NOT NULL,

    -- Schema相关
    f_schema_definition       TEXT NOT NULL,

    -- LogicView 专属字段
    f_logic_type              VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_logic_definition        TEXT NOT NULL,

    -- Local查询配置（物化）
    f_local_enabled           TINYINT NOT NULL DEFAULT 0,
    f_local_storage_engine    VARCHAR(50 CHAR) NOT NULL DEFAULT '',
    f_local_storage_config    TEXT NOT NULL,
    f_local_index_name        VARCHAR(255 CHAR) NOT NULL DEFAULT '',

    -- 同步配置
    f_sync_strategy           VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_sync_config             TEXT NOT NULL,
    f_sync_status             VARCHAR(20 CHAR) NOT NULL DEFAULT 'not_synced',
    f_last_sync_time          BIGINT NOT NULL DEFAULT 0,
    f_sync_error_message      TEXT NOT NULL,

    -- 审计字段
    f_creator                 VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    f_creator_type            VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_create_time             BIGINT NOT NULL DEFAULT 0,
    f_updater                 VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    f_updater_type            VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_update_time             BIGINT NOT NULL DEFAULT 0,

    -- 索引
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_t_resource_catalog_name ON t_resource(f_catalog_id, f_name);

CREATE INDEX IF NOT EXISTS idx_t_resource_category ON t_resource(f_category);

CREATE INDEX IF NOT EXISTS idx_t_resource_status ON t_resource(f_status);

-- ==========================================
-- 4. t_resource_schema_history Schema历史表
-- ==========================================
CREATE TABLE IF NOT EXISTS t_resource_schema_history (
    f_id                      VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_resource_id             VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_schema_version          VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_schema_definition       TEXT NOT NULL,

    -- 变更信息
    f_change_type             VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_change_summary          VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
    f_schema_inferred         TINYINT NOT NULL DEFAULT 0,
    f_change_time             BIGINT NOT NULL DEFAULT 0,

    -- 索引
    CLUSTER PRIMARY KEY (f_id)
);

CREATE INDEX IF NOT EXISTS idx_t_resource_schema_history_resource_id ON t_resource_schema_history(f_resource_id);

-- ==========================================
-- 5. t_connector_type Connector 类型注册表
-- ==========================================
CREATE TABLE IF NOT EXISTS t_connector_type (
    -- 主键与基础信息
    f_type                    VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_name                    VARCHAR(255 CHAR) NOT NULL DEFAULT '',
    f_tags                    VARCHAR(255 CHAR) NOT NULL DEFAULT '[]',
    f_description             VARCHAR(1000 CHAR) NOT NULL DEFAULT '',

    -- 类型分类
    f_mode                    VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_category                VARCHAR(32 CHAR) NOT NULL DEFAULT '',

    -- Remote 模式专用字段
    f_endpoint                VARCHAR(512 CHAR) NOT NULL DEFAULT '',

    -- 字段配置列表（JSON数组格式）
    f_field_config            TEXT NOT NULL,

    -- 状态
    f_enabled                 TINYINT NOT NULL DEFAULT 1,

    -- 索引
    CLUSTER PRIMARY KEY (f_type)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_t_connector_type_name ON t_connector_type(f_name);

CREATE INDEX IF NOT EXISTS idx_t_connector_type_mode ON t_connector_type(f_mode);

CREATE INDEX IF NOT EXISTS idx_t_connector_type_category ON t_connector_type(f_category);

CREATE INDEX IF NOT EXISTS idx_t_connector_type_enabled ON t_connector_type(f_enabled);



-- ==========================================
-- 6. 初始化内置 Local Connector
-- ==========================================
INSERT INTO t_connector_type (f_type, f_name, f_description, f_mode, f_category, f_field_config, f_enabled)
SELECT 'mariadb', 'mariadb', 'MariaDB 关系型数据库连接器', 'local', 'table',
    '{
        "host":      {"name":"主机地址","type":"string","description":"数据库服务器主机地址","required":true,"encrypted":false},
        "port":      {"name":"端口号","type":"integer","description":"数据库服务器端口","required":true,"encrypted":false},
        "username":  {"name":"用户名","type":"string","description":"数据库用户名","required":true,"encrypted":false},
        "password":  {"name":"密码","type":"string","description":"数据库密码","required":true,"encrypted":true},
        "databases": {"name":"数据库列表","type":"array","description":"数据库名称列表（可选，为空则连接实例级别）","required":false,"encrypted":false},
        "options":   {"name":"连接参数","type":"object","description":"连接参数（如 charset, timeout 等）","required":false,"encrypted":false}
    }',
    1
FROM DUAL WHERE NOT EXISTS ( SELECT f_type FROM t_connector_type WHERE f_type = 'mariadb' );

INSERT INTO t_connector_type (f_type, f_name, f_description, f_mode, f_category, f_field_config, f_enabled)
SELECT 'mysql', 'mysql', 'MySQL 关系型数据库连接器', 'local', 'table',
       '{
           "host":      {"name":"主机地址","type":"string","description":"数据库服务器主机地址","required":true,"encrypted":false},
           "port":      {"name":"端口号","type":"integer","description":"数据库服务器端口","required":true,"encrypted":false},
           "username":  {"name":"用户名","type":"string","description":"数据库用户名","required":true,"encrypted":false},
           "password":  {"name":"密码","type":"string","description":"数据库密码","required":true,"encrypted":true},
           "databases": {"name":"数据库列表","type":"array","description":"数据库名称列表（可选，为空则连接实例级别）","required":false,"encrypted":false},
           "options":   {"name":"连接参数","type":"object","description":"连接参数（如 charset, timeout 等）","required":false,"encrypted":false}
       }',
       1
FROM DUAL WHERE NOT EXISTS ( SELECT f_type FROM t_connector_type WHERE f_type = 'mysql' );

INSERT INTO t_connector_type (f_type, f_name, f_description, f_mode, f_category, f_field_config, f_enabled)
SELECT 'opensearch', 'opensearch', 'OpenSearch 搜索引擎连接器', 'local', 'index',
    '{
        "host":          {"name":"主机地址","type":"string","description":"OpenSearch 服务器主机地址","required":true,"encrypted":false},
        "port":          {"name":"端口号","type":"integer","description":"OpenSearch 服务器端口","required":true,"encrypted":false},
        "username":      {"name":"用户名","type":"string","description":"认证用户名","required":false,"encrypted":false},
        "password":      {"name":"密码","type":"string","description":"认证密码","required":false,"encrypted":true},
        "index_pattern": {"name":"索引模式","type":"string","description":"索引匹配模式（可选，如 log-*）","required":false,"encrypted":false}
    }',
    1
FROM DUAL WHERE NOT EXISTS ( SELECT f_type FROM t_connector_type WHERE f_type = 'opensearch' );

INSERT INTO t_connector_type (f_type, f_name, f_description, f_mode, f_category, f_field_config, f_enabled)
SELECT 'postgresql', 'postgresql', 'PostgreSQL 关系型数据库连接器', 'local', 'table',
       '{
           "host":      {"name":"主机地址","type":"string","description":"数据库服务器主机地址","required":true,"encrypted":false},
           "port":      {"name":"端口号","type":"integer","description":"数据库服务器端口","required":true,"encrypted":false},
           "username":  {"name":"用户名","type":"string","description":"数据库用户名","required":true,"encrypted":false},
           "password":  {"name":"密码","type":"string","description":"数据库密码","required":true,"encrypted":true},
           "database":  {"name":"数据库名","type":"string","description":"PostgreSQL 连接目标 database","required":true,"encrypted":false},
           "schemas":   {"name":"Schema 列表","type":"array","description":"可选；为空则扫描当前库下除系统 schema 外的用户 schema；非空则仅扫描列出的 schema","required":false,"encrypted":false},
           "options":   {"name":"连接参数","type":"object","description":"连接参数（如 sslmode、connect_timeout 等）","required":false,"encrypted":false}
       }',
       1
FROM DUAL WHERE NOT EXISTS ( SELECT f_type FROM t_connector_type WHERE f_type = 'postgresql' );

-- ==========================================
-- 7. t_discover_task 发现任务表
-- ==========================================
CREATE TABLE IF NOT EXISTS t_discover_task (
    -- 主键与关联信息
    f_id                      VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_catalog_id              VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_trigger_type            VARCHAR(20 CHAR) NOT NULL DEFAULT 'manual',

    -- 任务状态
    f_status                  VARCHAR(20 CHAR) NOT NULL DEFAULT 'pending',
    f_progress                INT NOT NULL DEFAULT 0,
    f_message                 VARCHAR(1000 CHAR) NOT NULL DEFAULT '',

    -- 时间信息
    f_start_time              BIGINT NOT NULL DEFAULT 0,
    f_finish_time             BIGINT NOT NULL DEFAULT 0,

    -- 执行结果
    f_result                  TEXT NOT NULL,

    -- 审计字段
    f_creator                 VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    f_creator_type            VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_create_time             BIGINT NOT NULL DEFAULT 0,

    -- 索引
    CLUSTER PRIMARY KEY (f_id)
);

CREATE INDEX IF NOT EXISTS idx_t_discover_task_catalog_id ON t_discover_task (f_catalog_id);

CREATE INDEX IF NOT EXISTS idx_t_discover_task_status ON t_discover_task (f_status);

-- ==========================================
-- 8. t_build_task 构建任务表
-- ==========================================
CREATE TABLE IF NOT EXISTS t_build_task (
    -- 主键与关联信息
    f_id                      VARCHAR(40 CHAR) NOT NULL,
    f_resource_id             VARCHAR(40 CHAR) NOT NULL,

    -- 任务状态
    f_status                  VARCHAR(20 CHAR) NOT NULL,
    f_mode                    VARCHAR(20 CHAR) NOT NULL,
    f_total_count             BIGINT NOT NULL DEFAULT 0,
    f_synced_count            BIGINT NOT NULL DEFAULT 0,
    f_vectorized_count        BIGINT NOT NULL DEFAULT 0,
    f_synced_mark             VARCHAR(100 CHAR) DEFAULT NULL,
    f_error_msg               TEXT DEFAULT NULL,

    -- 审计字段
    f_creator_id              VARCHAR(40 CHAR) NOT NULL,
    f_creator_type            VARCHAR(20 CHAR) NOT NULL,
    f_create_time             BIGINT NOT NULL,
    f_updater_id              VARCHAR(40 CHAR) NOT NULL,
    f_updater_type            VARCHAR(20 CHAR) NOT NULL,
    f_update_time             BIGINT NOT NULL,

    -- 索引
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_t_build_task_resource_id ON t_build_task(f_resource_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_t_build_task_status ON t_build_task(f_status);


-- Source: vega/data-connection/migrations/dm8/0.2.0/init.sql
-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

SET SCHEMA adp;

CREATE TABLE IF NOT EXISTS "data_source" (
    "id" varchar(36 char) NOT NULL COMMENT '主键，生成规则:36位uuid',
    "name" varchar(128 char) NOT NULL COMMENT '数据源展示名称',
    "type_name" varchar(30 char) NOT NULL COMMENT '数据库类型',
    "bin_data" blob NOT NULL COMMENT '数据源配置信息',
    "comment" varchar(255 char) DEFAULT NULL COMMENT '描述',
    "created_by_uid" varchar(36 char) NOT NULL COMMENT '创建人',
    "created_at" timestamp NOT NULL COMMENT '创建时间',
    "updated_by_uid" varchar(36 char) DEFAULT NULL COMMENT '修改人',
    "updated_at" timestamp DEFAULT NULL COMMENT '更新时间',
    CLUSTER PRIMARY KEY ("id")
    );

CREATE TABLE IF NOT EXISTS "t_data_source_info" (
    "f_id" varchar(36 char) NOT NULL COMMENT '主键，生成规则:36位uuid',
    "f_name" varchar(128 char) NOT NULL COMMENT '数据源展示名称',
    "f_type" varchar(30 char) NOT NULL COMMENT '数据库类型',
    "f_catalog" varchar(50 char) COMMENT '数据源catalog名称',
    "f_database" varchar(100 char) COMMENT '数据库名称',
    "f_schema" varchar(100 char) COMMENT '数据库模式',
    "f_connect_protocol" varchar(30 char) NOT NULL COMMENT '连接方式',
    "f_host" varchar(128 char) NOT NULL COMMENT '地址',
    "f_port" int NOT NULL COMMENT '端口',
    "f_account" varchar(128 char) COMMENT '账户',
    "f_password" varchar(1024 char) COMMENT '密码',
    "f_storage_protocol" varchar(30 char) COMMENT '存储介质',
    "f_storage_base" varchar(1024 char) COMMENT '存储路径',
    "f_token" varchar(100 char) COMMENT 'token认证',
    "f_replica_set" varchar(100 char) COMMENT '副本集名称',
    "f_is_built_in" tinyint NOT NULL DEFAULT '0' COMMENT '是否为内置数据源（0 特殊 1 非内置 2 内置），默认为0',
    "f_comment" varchar(255 char) COMMENT '描述',
    "f_created_by_uid" varchar(36 char) COMMENT '创建人',
    "f_created_at" datetime(3) COMMENT '创建时间',
    "f_updated_by_uid" varchar(36 char) COMMENT '修改人',
    "f_updated_at" datetime(3) COMMENT '更新时间',
    CLUSTER PRIMARY KEY ("f_id")
);

INSERT INTO "t_data_source_info" ("f_id","f_name","f_type","f_connect_protocol","f_host","f_port","f_is_built_in","f_created_at","f_updated_at")
SELECT 'cedb5294-07c3-45b1-a273-17baefa62800','索引库','index_base','http','mdl-index-base-svc',13013,2,current_timestamp(),current_timestamp()
FROM DUAL WHERE NOT EXISTS( SELECT "f_id" FROM "t_data_source_info" WHERE "f_id" = 'cedb5294-07c3-45b1-a273-17baefa62800' AND "f_type" = 'index_base');

CREATE TABLE IF NOT EXISTS "t_task_scan" (
  "id" varchar(36 char) NOT NULL COMMENT '唯一id，雪花算法',
  "type" tinyint NOT NULL DEFAULT 0 COMMENT '扫描任务：0 :即时-数据源;1 :即时-数据表;2: 定时-数据源;3: 定时-数据表',
  "name" varchar(128 char) NOT NULL COMMENT '任务名称',
  "ds_id" varchar(36 char) DEFAULT NULL COMMENT '数据源唯一标识',
  "scan_status" tinyint DEFAULT NULL COMMENT '任务状态',
  "start_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '任务开始时间',
  "end_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '任务结束时间',
  "create_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
  "task_params_info" text DEFAULT NULL COMMENT '任务执行参数信息',
  "task_process_info" text DEFAULT NULL COMMENT '任务执行进度信息',
  "task_result_info" text DEFAULT NULL COMMENT '任务执行结果信息',
  "schedule_id" varchar(36 char) DEFAULT NULL COMMENT '定时任务配置id',
  CLUSTER PRIMARY KEY ("id")
);
CREATE INDEX IF NOT EXISTS "t_task_scan_ds_id_IDX" on "t_task_scan" ("ds_id");

CREATE TABLE IF NOT EXISTS "t_task_scan_table" (
  "id" varchar(36 char) NOT NULL COMMENT '唯一id，雪花算法',
  "task_id" varchar(36 char) NOT NULL COMMENT '关联任务id',
  "ds_id" varchar(36 char) NOT NULL COMMENT '数据源唯一标识',
  "ds_name" varchar(128 char) NOT NULL COMMENT '数据源名称',
  "table_id" varchar(36 char) NOT NULL COMMENT 'table的唯一id',
  "table_name" varchar(128 char) NOT NULL COMMENT 'table的name',
  "schema_name" varchar(128 char) NOT NULL COMMENT 'schema的name',
  "scan_status" tinyint DEFAULT NULL COMMENT '任务状态：0 等待;1 进行中;2 成功;3 失败',
  "start_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '任务开始时间',
  "end_time" datetime DEFAULT NULL COMMENT '任务结束时间',
  "create_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
  "scan_params" text DEFAULT NULL COMMENT '任务执行参数信息',
  "scan_result_info" text DEFAULT NULL COMMENT '任务执行结果：',
  "error_stack" text DEFAULT NULL COMMENT '异常堆栈信息',
  "operation_type" tinyint DEFAULT NULL COMMENT '操作类型:0:insert 1:delete 2:update 3:unknown',
  CLUSTER PRIMARY KEY ("id")
);
CREATE INDEX IF NOT EXISTS "t_task_scan_table_task_id_IDX" on "t_task_scan_table" ("task_id");
CREATE INDEX IF NOT EXISTS "t_task_scan_table_table_id_IDX" on "t_task_scan_table" ("table_id");


CREATE TABLE IF NOT EXISTS "t_table_scan" (
  "f_id" varchar(36 char) NOT NULL COMMENT '唯一id，雪花算法',
  "f_name" varchar(128 char) NOT NULL COMMENT '表名称',
  "f_advanced_params" text DEFAULT NULL COMMENT '高级参数，格式为"{key(1): value(1), ... , key(n): value(n)}"',
  "f_description" varchar(2048 char) DEFAULT NULL COMMENT '表注释',
  "f_table_rows" bigint NOT NULL DEFAULT 0 COMMENT '表数据量，默认0',
  "f_data_source_id" varchar(36 char) NOT NULL COMMENT '数据源唯一标识',
  "f_data_source_name" varchar(128 char) NOT NULL COMMENT '冗余字段，数据源名称',
  "f_schema_name" varchar(128 char) NOT NULL COMMENT '冗余字段，schema名称',
  "f_task_id" varchar(36 char) NOT NULL COMMENT '关联任务id',
  "f_version" int NOT NULL DEFAULT 1 COMMENT '版本号',
  "f_create_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  "f_create_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
  "f_operation_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '修改时间，默认当前时间',
  "f_operation_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '修改用户（ID），默认空字符串',
  "f_operation_type" tinyint NOT NULL DEFAULT 0 COMMENT '状态：0新增1删除2更新',
  "f_status" tinyint NOT NULL DEFAULT 3 COMMENT '任务状态：0 等待;1 进行中;2 成功;3 失败',
  "f_status_change" tinyint NOT NULL DEFAULT 0 COMMENT '状态是否发生变化：0 否1 是',
  "f_scan_source" tinyint DEFAULT NULL COMMENT '扫描来源',
CLUSTER PRIMARY KEY ("f_id")
);
CREATE INDEX IF NOT EXISTS "t_table_scan_f_data_source_id_IDX" on "t_table_scan" ("f_task_id");

CREATE TABLE IF NOT EXISTS "t_table_field_scan" (
  "f_id" varchar(36 char) NOT NULL COMMENT '唯一id，雪花算法',
  "f_field_name" varchar(128 char) NOT NULL COMMENT '字段名',
  "f_table_id" varchar(36 char) NOT NULL COMMENT 'Table唯一标识',
  "f_table_name" varchar(128 char) NOT NULL COMMENT '表名',
  "f_field_type" varchar(128 char) DEFAULT NULL COMMENT '字段类型',
  "f_field_length" int DEFAULT NULL COMMENT '字段长度',
  "f_field_precision" int DEFAULT NULL COMMENT '字段精度',
  "f_field_comment" varchar(2048 char) DEFAULT NULL COMMENT '字段注释',
  "f_field_order_no" int DEFAULT NULL,
  "f_advanced_params" varchar(2048 char) DEFAULT NULL COMMENT '字段高级参数',
  "f_version" int NOT NULL DEFAULT 1 COMMENT '版本号',
  "f_create_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  "f_create_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
  "f_operation_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '修改时间，默认当前时间',
  "f_operation_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '修改用户（ID），默认空字符串',
  "f_operation_type" tinyint NOT NULL DEFAULT 0 COMMENT '状态：0新增1删除2更新',
  "f_status_change" tinyint NOT NULL DEFAULT 0 COMMENT '状态是否发生变化：0 否1 是',
  CLUSTER PRIMARY KEY ("f_id")
);
CREATE INDEX IF NOT EXISTS "t_table_field_scan_f_table_id_IDX" on "t_table_field_scan" ("f_table_id");

CREATE TABLE IF NOT EXISTS "t_task_scan_schedule" (
  "id" varchar(36 char) NOT NULL COMMENT '唯一id，雪花算法',
  "type" tinyint NOT NULL DEFAULT 0 COMMENT '扫描任务：0 :即时-数据源;1 :即时-数据表;2: 定时-数据源',
  "name" varchar(128 char) NOT NULL COMMENT '任务名称',
  "cron_expression" varchar(64 char) NOT NULL COMMENT 'cron表达式',
  "scan_strategy" varchar(64 char) DEFAULT NULL COMMENT '快速扫描策略',
  "task_status" tinyint DEFAULT 0 COMMENT '定时扫描任务:0 close 1 open',
  "ds_id" varchar(36 char) DEFAULT NULL COMMENT '数据源唯一标识',
  "create_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  "create_user" varchar(64 char) DEFAULT NULL COMMENT '创建用户',
  "operation_time" datetime DEFAULT NULL,
  "operation_user" varchar(64 char) DEFAULT NULL,
  "operation_type" tinyint NOT NULL DEFAULT 0 COMMENT '状态：0新增1删除2更新',
  CLUSTER PRIMARY KEY ("id")
);

CREATE INDEX IF NOT EXISTS "t_task_scan_ds_id_IDX" on "t_task_scan_schedule" ("ds_id");

-- metadata tables


CREATE TABLE IF NOT EXISTS "t_data_quality_model" (
                                                      "f_id" bigint NOT NULL COMMENT '主键id，唯一标识',
                                                      "f_ds_id" bigint NOT NULL COMMENT '数据源ID',
                                                      "f_dolphinscheduler_ds_id" bigint NOT NULL COMMENT 'dolphinscheduler数据源ID',
                                                      "f_db_type" varchar(50 char) NOT NULL COMMENT '数据库类型',
    "f_tb_name" varchar(512 char) NOT NULL COMMENT '表名称',
    "f_process_definition_code" bigint NOT NULL COMMENT '工作流定义ID',
    "f_crontab" varchar(128 char) DEFAULT NULL COMMENT '定时任务表达式',
    CLUSTER PRIMARY KEY ("f_id")
    );


CREATE TABLE IF NOT EXISTS "t_data_quality_rule" (
                                                     "f_id" bigint NOT NULL COMMENT '主键id，唯一标识',
                                                     "f_field_name" varchar(512 char) NOT NULL COMMENT '字段名称',
    "f_rule_id" tinyint NOT NULL COMMENT '质量规则ID：1-空值检测，，2-自定义SQL，5-字段长度校验，6-唯一性校验，7-正则表达式，9-枚举值校验，10-表行数校验',
    "f_threshold" double DEFAULT NULL COMMENT '阈值，默认0',
    "f_check_val" varchar(10240 char) DEFAULT NULL COMMENT '1、自定义sql：填写sql语句；2、字段长度校验：填写字段长度；3、正则表达式：填写正则表达式；4、枚举值校验：填写枚举值，逗号分割；5、表行数校验：填写表行数。',
    "f_check_val_name" varchar(128 char) DEFAULT NULL COMMENT '自定义sql时，填写的实际值名',
    "f_model_id" bigint NOT NULL COMMENT '质量模型ID',
    CLUSTER PRIMARY KEY ("f_id")
    );


CREATE TABLE IF NOT EXISTS "t_data_source" (
    "f_id" varchar(36 char) NOT NULL COMMENT '唯一id，雪花算法',
    "f_name" varchar(128 char) NOT NULL COMMENT '数据源名称',
    "f_data_source_type" tinyint NOT NULL COMMENT '类型，关联字典表f_dict_type为1时的f_dict_key',
    "f_data_source_type_name" varchar(256 char) NOT NULL COMMENT '类型名称，对应字典表f_dict_type为1时的f_dict_value',
    "f_user_name" varchar(128 char) NOT NULL COMMENT '用户名',
    "f_password" varchar(1024 char) NOT NULL COMMENT '密码',
    "f_description" varchar(255 char) NOT NULL DEFAULT '' COMMENT '描述',
    "f_extend_property" varchar(255 char) NOT NULL DEFAULT '' COMMENT '扩展属性，默认为空字符串',
    "f_host" varchar(128 char) NOT NULL COMMENT 'HOST',
    "f_port" int NOT NULL COMMENT '端口',
    "f_enable_status" tinyint NOT NULL DEFAULT 1 COMMENT '禁用/启用状态，1 启用，2 停用，默认为启用',
    "f_connect_status" tinyint NOT NULL DEFAULT 1 COMMENT '连接状态，1 成功，2 失败，默认为成功',
    "f_authority_id" bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
    "f_create_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    "f_create_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
    "f_update_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '修改时间，默认当前时间',
    "f_update_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '修改用户（ID），默认空字符串',
    "f_database" varchar(100 char) DEFAULT NULL COMMENT '数据库名称',
    "f_info_system_id" varchar(128 char) DEFAULT NULL COMMENT '信息系统id',
    "f_dolphin_id" bigint DEFAULT NULL COMMENT 'dolphin数据元id',
    "f_delete_code" bigint DEFAULT 0 COMMENT '逻辑删除标识码',
    "f_live_update_status" tinyint NOT NULL DEFAULT 0 COMMENT '实时更新标识（0无需更新，1待更新，2更新中，3连接不可用，4无权限，5待广播',
    "f_live_update_benchmark" varchar(255 char) DEFAULT NULL COMMENT '实时更新基准',
    "f_live_update_time" datetime DEFAULT current_timestamp() COMMENT '实时更新时间',
    CLUSTER PRIMARY KEY ("f_id")
    );
CREATE UNIQUE INDEX IF NOT EXISTS "t_data_source_un" on "t_data_source" ("f_name","f_create_user","f_info_system_id","f_delete_code");


CREATE TABLE IF NOT EXISTS "t_dict" (
                                        "f_id" int NOT NULL IDENTITY(1,1) COMMENT '唯一id，自增ID',
    "f_dict_type" tinyint NOT NULL COMMENT '字典类型\n1：数据源类型\n2：Oracle字段类型\n3：MySQL字段类型\n4：PostgreSQL字段类型\n5：SqlServer字段类型\n6：Hive字段类型\n7：HBase字段类型\n8：MongoDB字段类型\n9：FTP字段类型\n10：HDFS字段类型\n11：SFTP字段类型\n12：CMQ字段类型\n13：Kafka字段类型\n14：API字段类型',
    "f_dict_key" tinyint NOT NULL COMMENT '枚举值',
    "f_dict_value" varchar(256 char) NOT NULL COMMENT '枚举对应描述',
    "f_extend_property" varchar(1024 char) NOT NULL COMMENT '扩展属性',
    "f_enable_status" tinyint NOT NULL DEFAULT 2 COMMENT '启用状态，1 启用，2 停用，默认为停用',
    CLUSTER PRIMARY KEY ("f_id")
    );
CREATE UNIQUE INDEX IF NOT EXISTS "t_dict_un" on "t_dict" ("f_dict_type","f_dict_key");


CREATE TABLE IF NOT EXISTS "t_indicator" (
                                             "f_id" bigint NOT NULL COMMENT '唯一id，雪花算法',
                                             "f_indicator_name" varchar(128 char) NOT NULL COMMENT '指标名称',
    "f_indicator_type" varchar(128 char) NOT NULL COMMENT '指标类型',
    "f_indicator_value" bigint NOT NULL COMMENT '指标数值',
    "f_create_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    "f_indicator_object_id" bigint DEFAULT NULL COMMENT '关联对象ID',
    "f_authority_id" bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
    "f_advanced_params" varchar(255 char) NOT NULL DEFAULT '[]' COMMENT '指标高级参数',
    CLUSTER PRIMARY KEY ("f_id","f_create_time")
    );


CREATE TABLE IF NOT EXISTS "t_lineage_edge_column" (
    "f_id" varchar(64 char) NOT NULL COMMENT '主键ID，根据f_table_id和f_column_id值MD5计算得到',
    "f_parent_id" varchar(64 char) NOT NULL COMMENT '源字段ID',
    "f_child_id" varchar(64 char) NOT NULL COMMENT '目标字段ID',
    "f_create_type" varchar(20 char) DEFAULT NULL COMMENT '创建类型： HIVE/DATAX/SPARK/USER_REPORT',
    "f_query_text" text DEFAULT NULL COMMENT '生成血缘的sql或者脚本说明',
    "created_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "deleted_at" bigint DEFAULT 0 COMMENT '删除时间（逻辑删除）',
    "f_create_time" timestamp NULL DEFAULT NULL COMMENT '创建时间，时间戳',
    CLUSTER PRIMARY KEY ("f_id")
    );


CREATE TABLE IF NOT EXISTS "t_lineage_edge_column_table_relation" (
    "f_id" varchar(64 char) NOT NULL COMMENT '主键ID，根据f_table_id和f_column_id值MD5计算得到',
    "f_table_id" varchar(64 char) NOT NULL COMMENT '表ID',
    "f_column_id" varchar(64 char) NOT NULL COMMENT '字段ID',
    "created_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "deleted_at" bigint DEFAULT 0 COMMENT '删除时间（逻辑删除）',
    CLUSTER PRIMARY KEY ("f_id")
    );


CREATE TABLE IF NOT EXISTS "t_lineage_edge_table" (
    "f_id" varchar(64 char) NOT NULL COMMENT '主键ID，根据f_table_id和f_column_id值MD5计算得到',
    "f_parent_id" varchar(64 char) NOT NULL COMMENT '源ID',
    "f_child_id" varchar(64 char) NOT NULL COMMENT '目标ID',
    "f_create_type" varchar(20 char) DEFAULT NULL COMMENT '创建类型： HIVE/DATAX/SPARK/USER_REPORT',
    "f_query_text" text DEFAULT NULL COMMENT '生成血缘的sql或者脚本说明',
    "created_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "deleted_at" bigint DEFAULT 0 COMMENT '删除时间（逻辑删除）',
    "f_create_time" timestamp NULL DEFAULT NULL COMMENT '创建时间，时间戳',
    CLUSTER PRIMARY KEY ("f_id")
    );


CREATE TABLE IF NOT EXISTS "t_lineage_graph_info" (
    "app_id" varchar(20 char) NOT NULL COMMENT '图谱appId',
    "graph_id" bigint DEFAULT NULL COMMENT '图谱graphId',
    CLUSTER PRIMARY KEY ("app_id")
    );


CREATE TABLE IF NOT EXISTS "t_lineage_log" (
    "id" varchar(36 char) NOT NULL DEFAULT LOWER(RAWTOHEX(SYS_GUID())),
    "class_id" varchar(36 char) NOT NULL COMMENT '实体的主键id',
    "class_type" varchar(36 char) NOT NULL COMMENT '实体类型',
    "action_type" varchar(10 char) NOT NULL COMMENT '操作类型：insert update delete',
    "class_data" text NOT NULL COMMENT '血缘实体json',
    "created_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    CLUSTER PRIMARY KEY ("id")
    );


CREATE TABLE IF NOT EXISTS "t_lineage_relation" (
    "unique_id" varchar(255 char) NOT NULL COMMENT '实体ID',
    "class_type" tinyint DEFAULT NULL COMMENT '类型，1:column,2:indicator',
    "parent" text DEFAULT NULL COMMENT '上一个节点',
    "child" text DEFAULT NULL COMMENT '下一个节点',
    "created_at" datetime(3) DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) DEFAULT current_timestamp(3) COMMENT '更新时间',
    CLUSTER PRIMARY KEY ("unique_id")
    );


CREATE TABLE IF NOT EXISTS "t_lineage_tag_column" (
    "f_id" varchar(64 char) NOT NULL COMMENT '主键ID，根据f_table_id和f_column值MD5计算得到',
    "f_table_id" varchar(64 char) NOT NULL COMMENT 't_lineage_tag_table表ID',
    "f_column" varchar(255 char) NOT NULL COMMENT '字段名称',
    "created_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "deleted_at" bigint DEFAULT 0 COMMENT '删除时间（逻辑删除）',
    CLUSTER PRIMARY KEY ("f_id")
    );


CREATE TABLE IF NOT EXISTS "t_lineage_tag_table" (
    "f_id" varchar(64 char) NOT NULL COMMENT '主键ID，根据f_db_type、f_ds_id、f_jdbc_url、f_jdbc_user、f_db_name、f_db_schema、f_tb_name值MD5计算得到',
    "f_db_type" varchar(64 char) NOT NULL COMMENT '数据库类型',
    "f_ds_id" varchar(64 char) DEFAULT NULL COMMENT '数据源ID',
    "f_jdbc_url" varchar(255 char) DEFAULT NULL COMMENT '数据库连接URL',
    "f_jdbc_user" varchar(255 char) DEFAULT NULL COMMENT '数据库JDBC 用户名',
    "f_db_name" varchar(255 char) DEFAULT NULL COMMENT '数据库名称',
    "f_db_schema" varchar(255 char) DEFAULT NULL COMMENT '模式名称',
    "f_tb_name" varchar(255 char) NOT NULL COMMENT '表名称',
    "created_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "deleted_at" bigint DEFAULT 0 COMMENT '删除时间（逻辑删除）',
    CLUSTER PRIMARY KEY ("f_id")
    );


CREATE TABLE IF NOT EXISTS "t_indicator2" (
                                              "f_id" bigint NOT NULL COMMENT '唯一id，雪花算法',
                                              "f_indicator_name" varchar(128 char) NOT NULL COMMENT '指标名称',
    "f_indicator_type" varchar(128 char) NOT NULL COMMENT '指标类型',
    "f_indicator_value" bigint NOT NULL COMMENT '指标数值',
    "f_create_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    "f_indicator_object_id" bigint DEFAULT NULL COMMENT '关联对象ID',
    "f_authority_id" bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
    "f_advanced_params" varchar(255 char) NOT NULL DEFAULT '[]' COMMENT '指标高级参数',
    CLUSTER PRIMARY KEY ("f_id","f_create_time")
    );


CREATE TABLE IF NOT EXISTS "t_lineage_tag_column2" (
    "unique_id" varchar(255 char) NOT NULL COMMENT '列的唯一id',
    "uuid" varchar(36 char) DEFAULT NULL COMMENT '字段的uuid',
    "technical_name" varchar(255 char) DEFAULT NULL COMMENT '列技术名称',
    "business_name" varchar(255 char) DEFAULT NULL COMMENT '列业务名称',
    "comment" varchar(300 char) DEFAULT NULL COMMENT '字段注释',
    "data_type" varchar(255 char) DEFAULT NULL COMMENT '字段的数据类型',
    "primary_key" tinyint DEFAULT NULL COMMENT '是否主键',
    "table_unique_id" varchar(36 char) DEFAULT NULL COMMENT '属于血缘表的uuid',
    "expression_name" text DEFAULT NULL COMMENT 'column的生成表达式',
    "column_unique_ids" varchar(1024 char) DEFAULT '' COMMENT 'column的生成依赖的column的uid',
    "action_type" varchar(10 char) DEFAULT NULL COMMENT '操作类型:insertupdatedelete',
    "created_at" datetime(3) DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) DEFAULT current_timestamp(3) COMMENT '更新时间',
    CLUSTER PRIMARY KEY ("unique_id")
    );


CREATE TABLE IF NOT EXISTS "t_lineage_tag_indicator2" (
    "uuid" varchar(36 char) NOT NULL COMMENT '指标的uuid',
    "name" varchar(128 char) NOT NULL COMMENT '指标名称',
    "description" varchar(300 char) DEFAULT NULL COMMENT '指标名称描述',
    "code" varchar(128 char) NOT NULL COMMENT '指标编号',
    "indicator_type" varchar(10 char) NOT NULL COMMENT '指标类型:atomic原子derived衍生composite复合',
    "expression" text DEFAULT NULL COMMENT '指标表达式，如果指标是原子或复合指标时',
    "indicator_uuids" varchar(1024 char) DEFAULT '' COMMENT '引用的指标uuid',
    "time_restrict" text DEFAULT NULL COMMENT '时间限定表达式，如果指标是衍生指标时',
    "modifier_restrict" text DEFAULT NULL COMMENT '普通限定表达式，如果指标是衍生指标时',
    "owner_uid" varchar(50 char) DEFAULT NULL COMMENT '数据ownerID',
    "owner_name" varchar(128 char) DEFAULT NULL COMMENT '数据owner名称',
    "department_id" varchar(36 char) DEFAULT NULL COMMENT '所属部门id',
    "department_name" varchar(128 char) DEFAULT NULL COMMENT '所属部门名称',
    "column_unique_ids" varchar(1024 char) DEFAULT '' COMMENT '依赖的字段的unique_id',
    "action_type" varchar(10 char) NOT NULL COMMENT '操作类型:insertupdatedelete',
    "created_at" datetime(3) DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) DEFAULT current_timestamp(3) COMMENT '更新时间',
    CLUSTER PRIMARY KEY ("uuid")
    );


CREATE TABLE IF NOT EXISTS "t_lineage_tag_table2" (
    "unique_id" varchar(255 char) NOT NULL COMMENT '唯一id',
    "uuid" varchar(36 char) NOT NULL COMMENT '表的uuid',
    "technical_name" varchar(255 char) NOT NULL COMMENT '表技术名称',
    "business_name" varchar(255 char) DEFAULT NULL COMMENT '表业务名称',
    "comment" varchar(300 char) DEFAULT NULL COMMENT '表注释',
    "table_type" varchar(36 char) NOT NULL COMMENT '表类型',
    "datasource_id" varchar(36 char) DEFAULT NULL COMMENT '数据源id',
    "datasource_name" varchar(255 char) DEFAULT NULL COMMENT '数据源名称',
    "owner_id" varchar(36 char) DEFAULT NULL COMMENT '数据Ownerid',
    "owner_name" varchar(128 char) DEFAULT NULL COMMENT '数据OwnerName',
    "department_id" varchar(36 char) DEFAULT NULL COMMENT '所属部门id',
    "department_name" varchar(128 char) DEFAULT NULL COMMENT '所属部门mame',
    "info_system_id" varchar(36 char) DEFAULT NULL COMMENT '信息系统id',
    "info_system_name" varchar(128 char) DEFAULT NULL COMMENT '信息系统名称',
    "database_name" varchar(128 char) NOT NULL COMMENT '数据库名称',
    "catalog_name" varchar(255 char) NOT NULL DEFAULT '' COMMENT '数据源catalog名称',
    "catalog_addr" varchar(1024 char) NOT NULL DEFAULT '' COMMENT '数据源地址',
    "catalog_type" varchar(128 char) NOT NULL COMMENT '数据库类型名称',
    "task_execution_info" varchar(128 char) DEFAULT NULL COMMENT '表加工任务的相关名称',
    "action_type" varchar(10 char) NOT NULL COMMENT '操作类型:insertupdatedelete',
    "created_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '更新时间',
    CLUSTER PRIMARY KEY ("unique_id")
    );


CREATE TABLE IF NOT EXISTS "t_live_ddl" (
                                            "f_id" bigint NOT NULL IDENTITY(1,1) COMMENT '唯一标识',
    "f_data_source_id" bigint NOT NULL DEFAULT 0 COMMENT '数据源ID',
    "f_data_source_name" varchar(255 char) NOT NULL DEFAULT '' COMMENT '数据源名称',
    "f_origin_catalog" varchar(255 char) DEFAULT NULL COMMENT '物理catalog',
    "f_virtual_catalog" varchar(255 char) DEFAULT NULL COMMENT '虚拟化catalog',
    "f_schema_id" bigint DEFAULT NULL COMMENT 'schemaID',
    "f_schema_name" varchar(255 char) DEFAULT NULL COMMENT 'schema名称',
    "f_table_id" bigint DEFAULT NULL COMMENT 'tableID',
    "f_table_name" varchar(255 char) DEFAULT NULL COMMENT 'table名称',
    "f_sql_type" varchar(100 char) DEFAULT NULL COMMENT 'sql类型(AlterTable,AlterColumn,CreateTable,CommentTable,CommentColumn,DropTable,RenameTable)',
    "f_sql_text" text NOT NULL COMMENT 'sql文本',
    "f_live_update_benchmark" varchar(255 char) NOT NULL DEFAULT '' COMMENT '实时更新基准',
    "f_monitor_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '监听时间，默认当前时间',
    "f_update_status" tinyint DEFAULT NULL COMMENT '更新状态（0全量更新，1增量更新，2忽略更新，3待更新，4解析失败，5更新失败）',
    "f_update_message" varchar(2000 char) DEFAULT NULL COMMENT '更新信息',
    "f_push_status" tinyint DEFAULT NULL COMMENT '0不推送,1待推送,2已推送',
    CLUSTER PRIMARY KEY ("f_id")
    );


CREATE TABLE IF NOT EXISTS "t_schema" (
                                          "f_id" bigint NOT NULL COMMENT '唯一id，雪花算法',
                                          "f_name" varchar(128 char) NOT NULL COMMENT 'schema名称',
    "f_data_source_id" varchar(36 char) NOT NULL COMMENT '数据源唯一标识',
    "f_data_source_name" varchar(128 char) NOT NULL COMMENT '冗余字段，数据源名称',
    "f_data_source_type" tinyint NOT NULL COMMENT '冗余字段，数据源类型，关联字典表f_dict_type为1时的f_dict_key',
    "f_data_source_type_name" varchar(256 char) NOT NULL COMMENT '冗余字段，数据源类型名称，对应字典表f_dict_type为1时的f_dict_value',
    "f_authority_id" bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
    "f_create_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    "f_create_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
    "f_update_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '修改时间，默认当前时间',
    "f_update_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '修改用户（ID），默认空字符串',
    CLUSTER PRIMARY KEY ("f_id")
    );
CREATE UNIQUE INDEX IF NOT EXISTS "t_schema_un" on "t_schema" ("f_data_source_id","f_name");


CREATE TABLE IF NOT EXISTS "t_table" (
                                         "f_id" bigint NOT NULL COMMENT '唯一id，雪花算法',
                                         "f_name" varchar(128 char) NOT NULL COMMENT '表名称',
    "f_advanced_params" text NOT NULL COMMENT '高级参数，默认为"{}"，格式为"{key(1): value(1), ... , key(n): value(n)}"',
    "f_description" varchar(2048 char) DEFAULT NULL COMMENT '表注释',
    "f_table_rows" bigint NOT NULL DEFAULT 0 COMMENT '表数据量，默认0',
    "f_schema_id" bigint NOT NULL COMMENT 'schema唯一标识',
    "f_schema_name" varchar(128 char) NOT NULL COMMENT '冗余字段，schema名称',
    "f_data_source_id" varchar(36 char) NOT NULL COMMENT '数据源唯一标识',
    "f_data_source_name" varchar(128 char) NOT NULL COMMENT '冗余字段，数据源名称',
    "f_data_source_type" tinyint NOT NULL COMMENT '冗余字段，数据源类型，关联字典表f_dict_type为1时的f_dict_key',
    "f_data_source_type_name" varchar(256 char) NOT NULL COMMENT '冗余字段，数据源类型名称，对应字典表f_dict_type为1时的f_dict_value',
    "f_version" int NOT NULL DEFAULT 1 COMMENT '版本号',
    "f_authority_id" varchar(100 char) NOT NULL DEFAULT '' COMMENT '权限域（目前为预留字段），默认0',
    "f_create_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    "f_create_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
    "f_update_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '修改时间，默认当前时间',
    "f_update_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '修改用户（ID），默认空字符串',
    "f_delete_flag" tinyint NOT NULL DEFAULT 0 COMMENT '逻辑删除标识',
    "f_delete_time" datetime DEFAULT NULL COMMENT '逻辑删除时间',
    "f_scan_source" tinyint DEFAULT NULL  COMMENT '扫描来源',
    CLUSTER PRIMARY KEY ("f_id")
    );
CREATE UNIQUE INDEX IF NOT EXISTS "t_table_un" on "t_table" ("f_data_source_id","f_schema_id","f_name");


CREATE TABLE IF NOT EXISTS "t_table_field" (
                                               "f_table_id" bigint NOT NULL COMMENT 'Table唯一标识',
                                               "f_field_name" varchar(128 char) NOT NULL COMMENT '字段名',
    "f_field_type" varchar(128 char) DEFAULT NULL COMMENT '字段类型',
    "f_field_length" int DEFAULT NULL COMMENT '字段长度',
    "f_field_precision" int DEFAULT NULL COMMENT '字段精度',
    "f_field_comment" varchar(2048 char) DEFAULT NULL COMMENT '字段注释',
    "f_advanced_params" varchar(2048 char) NOT NULL DEFAULT '[]' COMMENT '字段高级参数',
    "f_update_flag" tinyint NOT NULL DEFAULT 0 COMMENT '更新标识',
    "f_update_time" datetime DEFAULT NULL COMMENT '更新时间',
    "f_delete_flag" tinyint NOT NULL DEFAULT 0 COMMENT '逻辑删除标识',
    "f_delete_time" datetime DEFAULT NULL COMMENT '逻辑删除时间',
    CLUSTER PRIMARY KEY ("f_table_id","f_field_name")
    );


CREATE TABLE IF NOT EXISTS "t_table_field_his" (
                                                   "f_id" bigint NOT NULL COMMENT '唯一id，雪花算法',
                                                   "f_field_name" varchar(128 char) NOT NULL COMMENT '字段名',
    "f_field_type" varchar(128 char) DEFAULT NULL COMMENT '字段类型',
    "f_field_length" int DEFAULT NULL COMMENT '字段长度',
    "f_field_precision" int DEFAULT NULL COMMENT '字段精度',
    "f_field_comment" varchar(2048 char) DEFAULT NULL COMMENT '字段注释',
    "f_table_id" bigint NOT NULL COMMENT 'Table唯一标识',
    "f_version" int NOT NULL DEFAULT 1 COMMENT '版本号',
    "f_advanced_params" varchar(255 char) NOT NULL DEFAULT '[]' COMMENT '字段高级参数',
    CLUSTER PRIMARY KEY ("f_id","f_version")
    );


CREATE TABLE IF NOT EXISTS "t_task" (
                                        "f_id" bigint NOT NULL COMMENT '唯一id，雪花算法',
                                        "f_object_id" varchar(36 char) DEFAULT NULL COMMENT '任务对象id',
    "f_object_type" tinyint DEFAULT NULL COMMENT '任务对象类型1数据源、2数据表',
    "f_name" varchar(255 char) DEFAULT NULL COMMENT '任务名称',
    "f_status" tinyint NOT NULL COMMENT '任务状态：0成功，1失败，2进行中',
    "f_start_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '任务开始时间',
    "f_end_time" datetime DEFAULT NULL COMMENT '任务结束时间',
    "f_create_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '创建用户',
    "f_authority_id" bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
    "f_advanced_params" varchar(255 char) NOT NULL DEFAULT '[]' COMMENT '任务高级参数',
    CLUSTER PRIMARY KEY ("f_id")
    );


CREATE TABLE IF NOT EXISTS "t_task_log" (
                                            "f_id" bigint NOT NULL COMMENT '唯一id，雪花算法',
                                            "f_task_id" bigint DEFAULT NULL COMMENT '任务id',
                                            "f_log" text DEFAULT NULL COMMENT '任务日志文本',
                                            "f_authority_id" bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
                                            CLUSTER PRIMARY KEY ("f_id")
    );



-- 添加虚拟化数据源类型
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,1,'Oracle','{"dbCatalogName": 关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,2,'MySQL','{"dbCatalogName": 关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,3,'PostgreSQL','{"dbCatalogName": 关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,4,'SqlServer','{"dbCatalogName": 关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,5,'Hive','{"dbCatalogName": 关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,6,'HBase','{"dbCatalogName": 非关系型数据库}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,7,'MongoDB','{"dbCatalogName": 非关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 7);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,9,'HDFS','{"dbCatalogName": 文件系统}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 9);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,10,'SFTP','{"dbCatalogName": 文件系统}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 10);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,11,'CMQ','{"dbCatalogName": 消息队列}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 11);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,12,'Kafka','{"dbCatalogName": 消息队列}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 12);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,13,'API','{"dbCatalogName": 其他}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 13);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,14,'CLICKHOUSE', '{"dbCatalogName": 非关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 14);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,15,'doris','{"dbCatalogName": 非关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 15);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,16,'mariadb','{"dbCatalogName": 关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 16);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,17,'dm','{"dbCatalogName": 关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 17);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,18,'maxcompute','{"dbCatalogName": 关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 18);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,1,'VARCHAR2','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,2,'NUMBER','{"jdbcType":2,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,3,'NVARCHAR2','{"jdbcType":-9,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,4,'DATE','{"jdbcType":91,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,5,'NCLOB','{"jdbcType":2011,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,6,'TIMESTAMP','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,7,'CLOB','{"jdbcType":2005,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 7);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,8,'TIMESTAMP(6)','{"jdbcType":10,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 8);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,9,'CHAR','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 9);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,10,'VARCHAR','{"jdbcType":11,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 10);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,11,'FLOAT','{"jdbcType":6,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 11);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,12,'BLOB','{"jdbcType":2004,"javaColumnType":"7","sparkColumnType":"8","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 12);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,13,'DECIMAL','{"jdbcType":3,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 13);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,14,'LONG','{"jdbcType":-1,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 14);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,15,'INT','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 15);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,16,'RAW','{"jdbcType":-3,"javaColumnType":"7","sparkColumnType":"8","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 16);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,17,'TIMESTAMP(4)','{"jdbcType":3,"javaColumnType":"8","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 17);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,18,'ROWID','{"jdbcType":-8,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 18);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,19,'AQ$_SUBSCRIBERS','{"jdbcType":2003,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 19);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,20,'LONG RAW','{"jdbcType":-4,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 20);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,21,'NCHAR','{"jdbcType":-15,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 21);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,1,'DOUBLE','{"jdbcType":8,"javaColumnType":"3","sparkColumnType":"4","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,2,'TINYINT','{"jdbcType":-6,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,3,'BOOLEAN','{"jdbcType":16,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,4,'INTEGER','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,5,'VARCHAR','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,6,'CHAR','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,7,'BLOB','{"jdbcType":-4,"javaColumnType":"7","sparkColumnType":"8","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 7);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,8,'SMALLINT','{"jdbcType":5,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 8);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,9,'MEDIUMINT','{"jdbcType":4,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 9);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,10,'BIT','{"jdbcType":-7,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 10);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,11,'FLOAT','{"jdbcType":7,"javaColumnType":"3","sparkColumnType":"4","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 11);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,12,'DECIMAL','{"jdbcType":3,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 12);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,13,'DATE','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 13);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,14,'TIME','{"jdbcType":92,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 14);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,15,'DATETIME','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 15);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,16,'TIMESTAMP','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 16);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,17,'YEAR','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 17);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,18,'INT','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 18);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,19,'BIGINT','{"jdbcType":-5,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 19);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,20,'LONGTEXT','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 20);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,21,'TEXT','{"jdbcType":-1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 21);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,22,'JSON','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 22);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,23,'MEDIUMTEXT','{"jdbcType":13,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 23);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,24,'SET','{"jdbcType":11,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 24);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,25,'MEDIUMBLOB','{"jdbcType":11,"javaColumnType":"7","sparkColumnType":"8","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 25);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,26,'LONGBLOB','{"jdbcType":11,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 26);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,27,'VARBINARY','{"jdbcType":-3,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 27);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,28,'BINARY','{"jdbcType":-2,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 28);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,29,'ENUM','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 29);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,30,'TINYTEXT','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 30);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,1,'int2','{"jdbcType":12,"javaColumnType":"3","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,2,'int4','{"jdbcType":12,"javaColumnType":"3","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,3,'int8','{"jdbcType":-5,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,4,'varchar','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,5,'date','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,6,'timestamp','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,7,'timestampwithouttimezone','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 7);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,8,'bigint','{"jdbcType":11,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 8);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,9,'decimal','{"jdbcType":3,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 9);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,10,'TEXT','{"jdbcType":5,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 10);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,11,'bigserial','{"jdbcType":11,"javaColumnType":"2","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 11);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,12,'timestampwithtimezone','{"jdbcType":99,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 12);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,13,'numeric','{"jdbcType":3,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 13);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,14,'float8','{"jdbcType":12,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 14);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,15,'float4','{"jdbcType":12,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 15);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,16,'bpchar','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 16);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,17,'serial','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 17);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,18,'json','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 18);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,19,'bytea','{"jdbcType":2004,"javaColumnType":"7","sparkColumnType":"8","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 19);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,20,'bool','{"jdbcType":16,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 20);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,21,'array','{"jdbcType":2003}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 21);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,22,'numeric','{"jdbcType":2,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 22);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,1,'varchar','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,2,'int','{"jdbcType":4,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,3,'datetime','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,4,'smallint','{"jdbcType":5,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,5,'sysname','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,6,'date','{"jdbcType":93,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,7,'datetime2','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 7);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,8,'nvarchar','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 8);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,9,'intidentity','{"jdbcType":11,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 9);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,10,'decimal','{"jdbcType":11,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 10);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,11,'char','{"jdbcType":11,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 11);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,12,'bigint','{"jdbcType":3,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 12);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,13,'uniqueidentifier','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 13);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,14,'tinyint','{"jdbcType":4,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 14);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,15,'real','{"jdbcType":11,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 15);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,16,'nchar','{"jdbcType":11,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 16);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,1,'INT','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,2,'STRING','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,3,'BOOLEAN','{"jdbcType":16,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,4,'DECIMAL','{"jdbcType":3,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,5,'TIMESTAMP','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,6,'SMALLINT','{"jdbcType":5,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,7,'TINYINT','{"jdbcType":-6,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 7);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,8,'BINARY','{"jdbcType":-2,"javaColumnType":"7","sparkColumnType":"8","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 8);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,9,'VARCHAR','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 9);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,10,'CHAR','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 10);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,11,'DATE','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 11);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,12,'DOUBLE','{"jdbcType":8,"javaColumnType":"3","sparkColumnType":"4","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 12);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,13,'BIGINT','{"jdbcType":-5,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 13);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,14,'FLOAT','{"jdbcType":6,"javaColumnType":"3","sparkColumnType":"4","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 14);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,1,'hbase-int','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,2,'hbase-string','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,3,'hbase-date','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,4,'hbase_long','{"jdbcType":-5,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,5,'hbase_bytes','{"jdbcType":-2,"javaColumnType":"7","sparkColumnType":"8","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,6,'hbase_boolean','{"jdbcType":16,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,7,'hbase_double','{"jdbcType":8,"javaColumnType":"3","sparkColumnType":"4","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 7);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,8,'hbase_timestamp','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 8);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,1,'DOUBLE','{"jdbcType":8,"javaColumnType":"3","sparkColumnType":"4","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,2,'STRING','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,3,'OBJECT','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,4,'ARRAY','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,5,'DATE','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,6,'INT','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,7,'OBJECTID','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 7);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,8,'LONG','{"jdbcType":-5,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 8);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,9,'BASICDBOBJECT','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 9);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,10,'INTERGER','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 10);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,11,'NULL','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 11);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,12,'INTEGER','{"jdbcType":11,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 12);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,13,'BASICDBLIST','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 13);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 9,4,'FTPLONG','{"jdbcType":-5,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 9 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 9,5,'FTPDATE','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 9 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 9,6,'FTPTIMESTAMP','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 9 AND f_dict_key = 6);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 10,1,'HDFSSTRING','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 10 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 10,2,'HDFSINT','{"jdbcType":-5,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 10 AND f_dict_key = 2);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 11,1,'SFTPLONG','{"jdbcType":-5,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 11 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 11,2,'SFTPTIMESTAMP','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 11 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 11,3,'SFTPDATE','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 11 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 11,4,'SFTPDOUBLE','{"jdbcType":8,"javaColumnType":"3","sparkColumnType":"4","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 11 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 11,5,'SFTPSTRING','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 11 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 11,6,'FTPINT','{"jdbcType":6,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 11 AND f_dict_key = 6);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 13,1,'kafkaString','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"消息队列"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 13 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 13,2,'kafkaInteger','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"消息队列"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 13 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 13,3,'kafkaLong','{"jdbcType":-5,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"消息队列"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 13 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 13,4,'kafkaDouble','{"jdbcType":8,"javaColumnType":"3","sparkColumnType":"4","flinkColumnType":"4","dbCatalogName":"消息队列"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 13 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 13,5,'kafkaBool','{"jdbcType":16,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"消息队列"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 13 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 13,6,'kafkaDate','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"消息队列"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 13 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 13,7,'kafkaTimestamp','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"消息队列"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 13 AND f_dict_key = 7);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 14,1,'api_boolean','{"jdbcType":16,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"其他"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 14 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 14,2,'api_date','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"其他"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 14 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 14,3,'api_timestamp','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"其他"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 14 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 14,4,'api_long','{"jdbcType":4,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"其他"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 14 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 14,5,'api_string','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"其他"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 14 AND f_dict_key = 5);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 15,1,'oracle.jdbc.OracleDriver','oracle驱动类',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 15 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 15,2,'com.mysql.cj.jdbc.Driver','mysql驱动类',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 15 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 15,3,'org.postgresql.Driver','postgresql驱动类',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 15 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 15,4,'com.microsoft.sqlserver.jdbc.SQLServerDriver','sqlserver驱动类',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 15 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 15,5,'org.apache.hive.jdbc.HiveDriver','hive驱动类',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 15 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 15,14,'jdbc:clickhouse://', 'clickhouse-jdbc前缀', 1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 15 AND f_dict_key = 14);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 16,1,'jdbc:oracle:thin:@//','oracle-jdbc前缀',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 16 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 16,2,'jdbc:mysql://','mysql-jdbc前缀',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 16 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 16,3,'jdbc:postgresql://','postgresql-jdbc前缀',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 16 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 16,4,'jdbc:sqlserver://','sqlserver-jdbc前缀',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 16 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 16,5,'jdbc:hive2://','hive2-jdbc前缀',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 16 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 16,14,'com.clickhouse.jdbc.ClickHouseDriver','clickhouse驱动类',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 16 AND f_dict_key = 14);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 17,1,'select 1 from dual','oracle-有效性检测',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 17 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 17,2,'select 1','mysql-有效性检测',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 17 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 17,3,'select version()','postgresql-有效性检测',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 17 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 17,4,'select 1','sqlserver-有效性检测',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 17 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 17,5,'select 1','hive2-有效性检测',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 17 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 17,14,'select 1','clickhouse有效性检测',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 17 AND f_dict_key = 14);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,0,'UNKNOWN','',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 0);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,0,'UNKNOWN','',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 0);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,0,'UNKNOWN','',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 0);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,0,'UNKNOWN','',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 0);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,0,'UNKNOWN','',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 0);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,0,'UNKNOWN','',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 0);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,0,'UNKNOWN','',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 0);


-- 当前血缘初始化依赖于 AnyFabric，引擎从 AnyFabric 剥离后，临时跳过血缘的初始化逻辑。后续考虑整体方案。
INSERT INTO t_lineage_graph_info (app_id,graph_id) SELECT '1',0 FROM DUAL WHERE NOT EXISTS ( SELECT app_id from t_lineage_graph_info where app_id = 1);
INSERT INTO t_lineage_graph_info (app_id,graph_id) SELECT '2',0 FROM DUAL WHERE NOT EXISTS ( SELECT app_id from t_lineage_graph_info where app_id = 2);
INSERT INTO t_lineage_graph_info (app_id,graph_id) SELECT '3',0 FROM DUAL WHERE NOT EXISTS ( SELECT app_id from t_lineage_graph_info where app_id = 3);
INSERT INTO t_lineage_graph_info (app_id,graph_id) SELECT '4',0 FROM DUAL WHERE NOT EXISTS ( SELECT app_id from t_lineage_graph_info where app_id = 4);
INSERT INTO t_lineage_graph_info (app_id,graph_id) SELECT '5',0 FROM DUAL WHERE NOT EXISTS ( SELECT app_id from t_lineage_graph_info where app_id = 5);
INSERT INTO t_lineage_graph_info (app_id,graph_id) SELECT '6',0 FROM DUAL WHERE NOT EXISTS ( SELECT app_id from t_lineage_graph_info where app_id = 6);


-- Source: vega/mdl-data-model/migrations/dm8/0.1.0/init.sql
-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

SET SCHEMA adp;

CREATE TABLE IF NOT EXISTS t_metric_model (
  f_model_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_model_name VARCHAR(40 CHAR) NOT NULL,
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(255 CHAR) DEFAULT NULL,
  f_catalog_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_catalog_content TEXT DEFAULT NULL,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_measure_name VARCHAR(50 CHAR) NOT NULL DEFAULT '',
  f_metric_type VARCHAR(20 CHAR) NOT NULL,
  f_data_source VARCHAR(255 CHAR) NOT NULL,
  f_query_type VARCHAR(20 CHAR) NOT NULL,
  f_formula TEXT NOT NULL,
  f_formula_config TEXT DEFAULT NULL,
  f_analysis_dimessions VARCHAR(8192 CHAR) DEFAULT NULL,
  f_order_by_fields VARCHAR(4096 CHAR) DEFAULT NULL,
  f_having_condition VARCHAR(2048 CHAR) DEFAULT NULL,
  f_date_field VARCHAR(255 CHAR) DEFAULT NULL,
  f_measure_field VARCHAR(255 CHAR) NOT NULL,
  f_unit_type VARCHAR(40 CHAR) NOT NULL,
  f_unit VARCHAR(20 CHAR) NOT NULL,
  f_group_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_builtin TINYINT DEFAULT 0,
  f_calendar_interval TINYINT DEFAULT 0,
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
  f_variables TEXT DEFAULT NULL,
  f_time_windows VARCHAR(1024 CHAR) DEFAULT NULL,
  f_steps VARCHAR(255 CHAR) NOT NULL DEFAULT '[]',
  f_plan_time BIGINT NOT NULL DEFAULT 0,
  f_index_base VARCHAR(40 CHAR) NOT NULL,
  f_retrace_duration VARCHAR(20 CHAR) DEFAULT NULL,
  f_schedule_sync_status TINYINT NOT NULL,
  f_execute_status TINYINT DEFAULT 0,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
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
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
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
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
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
  f_technical_name VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_group_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_type  VARCHAR(10 CHAR) NOT NULL DEFAULT '',
  f_query_type VARCHAR(10 CHAR) NOT NULL DEFAULT '',
  f_builtin TINYINT DEFAULT 0,
  f_tags VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_comment VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_data_source_type  VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_data_source_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_file_name VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  f_excel_config TEXT DEFAULT NULL,
  f_data_scope TEXT DEFAULT NULL,
  f_fields TEXT DEFAULT NULL,
  f_status VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_metadata_form_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_primary_keys VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_sql TEXT DEFAULT NULL,
  f_meta_table_name VARCHAR(1024 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_delete_time BIGINT NOT NULL DEFAULT 0,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_data_source TEXT DEFAULT NULL,
  f_field_scope TINYINT NOT NULL DEFAULT '0',
  f_filters TEXT DEFAULT NULL,
  f_open_streaming TINYINT NOT NULL DEFAULT 0,
  f_job_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_loggroup_filters TEXT DEFAULT NULL,
  CLUSTER PRIMARY KEY (f_view_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_view_uk_f_view_name ON t_data_view(f_group_id, f_view_name, f_delete_time);


CREATE TABLE IF NOT EXISTS t_data_view_group (
  f_group_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_group_name VARCHAR(40 CHAR) NOT NULL,
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_delete_time BIGINT NOT NULL DEFAULT 0,
  f_builtin TINYINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_group_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS "t_data_view_group_uk_f_group_name" ON "t_data_view_group"(f_builtin, f_group_name, f_delete_time);


CREATE TABLE IF NOT EXISTS t_data_view_row_column_rule (
  f_rule_id VARCHAR(40 CHAR) NOT NULL DEFAULT '' COMMENT '视图行列规则 id',
  f_rule_name VARCHAR(255 CHAR) NOT NULL COMMENT '视图行列规则名称',
  f_view_id VARCHAR(40 CHAR) NOT NULL COMMENT '视图 id',
  f_tags VARCHAR(255 CHAR) NOT NULL DEFAULT '' COMMENT '标签',
  f_comment VARCHAR(255 CHAR) NOT NULL DEFAULT '' COMMENT '备注',
  f_fields TEXT NOT NULL COMMENT '列',
  f_row_filters TEXT NOT NULL COMMENT '行过滤规则',
  f_create_time BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间', 
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '' COMMENT '更新者id',
  f_updater_type VARCHAR(20 CHAR) NOT NULL DEFAULT '' COMMENT '更新者类型',
  CLUSTER PRIMARY KEY (f_rule_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS "t_data_view_row_column_rule_uk_f_rule_name" ON "t_data_view_row_column_rule" (f_rule_name, f_view_id);


CREATE TABLE IF NOT EXISTS t_data_dict (
  f_dict_id VARCHAR(40 CHAR) NOT NULL,
  f_dict_name VARCHAR(255 CHAR) NOT NULL,
  f_tags VARCHAR(255 CHAR) NOT NULL,
  f_comment VARCHAR(255 CHAR) DEFAULT NULL,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_dict_type VARCHAR(20 CHAR) NOT NULL DEFAULT 'kv_dict',
  f_dict_store VARCHAR(255 CHAR) NOT NULL,
  f_dimension VARCHAR(1500 CHAR) NOT NULL,
  f_unique_key TINYINT NOT NULL DEFAULT 1,
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
  f_config TEXT NOT NULL,
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
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_span_source_type VARCHAR(40 CHAR) NOT NULL,
  f_span_config TEXT NOT NULL,
  f_enabled_related_log TINYINT NOT NULL,
  f_related_log_source_type VARCHAR(40 CHAR) NOT NULL,
  f_related_log_config TEXT NOT NULL,
  CLUSTER PRIMARY KEY (f_model_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_trace_model_uk_f_model_name ON t_trace_model(f_model_name);

CREATE INDEX IF NOT EXISTS t_trace_model_idx_f_span_source_type ON t_trace_model(f_span_source_type);

CREATE TABLE IF NOT EXISTS t_data_model_job (
  f_job_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_job_type VARCHAR(40 CHAR) NOT NULL,
  f_job_config TEXT,
  f_job_status VARCHAR(20 CHAR) NOT NULL,
  f_job_status_details TEXT NOT NULL,
  CLUSTER PRIMARY KEY (f_job_id)
);

CREATE TABLE IF NOT EXISTS t_objective_model (
  f_model_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_model_name VARCHAR(40 CHAR) NOT NULL,
  f_tags VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_comment VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_objective_type VARCHAR(20 CHAR) NOT NULL,
  f_objective_config TEXT NOT NULL,
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

-- Source: vega/vega-gateway/migrations/dm8/0.2.0/init.sql
-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

SET SCHEMA adp;


CREATE TABLE IF NOT EXISTS "cache_table" (
  "id" varchar(36 char) NOT NULL COMMENT '主键',
  "catalog_name" varchar(36 char) NOT NULL COMMENT '对应的逻辑视图的catalog名称',
  "schema_name" varchar(36 char) NOT NULL COMMENT '对应的逻辑视图的schema名称',
  "table_name" varchar(36 char) NOT NULL COMMENT '对应的逻辑视图的table名称',
  "cts_sql" text DEFAULT NULL COMMENT '表的建表sql',
  "source_create_sql" text DEFAULT NULL COMMENT '样例数据查询sql',
  "current_view_original_text" text DEFAULT NULL COMMENT '最近一次的原始加密sql',
  "status" varchar(36 char) NOT NULL COMMENT '可用；异常；正在初始化',
  "mid_status" varchar(36 char) DEFAULT NULL COMMENT '在FSM任务的时候的中间状态',
  "deps" varchar(255 char) DEFAULT '' COMMENT '生成的结果缓存表的id用,分隔',
  "create_time" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '更新时间',
  "update_time" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '更新时间',
  CLUSTER PRIMARY KEY ("id")
);


CREATE TABLE IF NOT EXISTS "client_id" (
  "id" int NOT NULL COMMENT '主键id',
  "client_name" varchar(128 char) DEFAULT NULL COMMENT '客户端名称',
  "client_id" varchar(64 char) DEFAULT NULL COMMENT '客户端id',
  "client_secret" varchar(64 char) DEFAULT NULL COMMENT '客户端密码',
  "create_time" datetime DEFAULT NULL COMMENT '创建时间',
  "update_time" datetime DEFAULT NULL COMMENT '更新时间',
  CLUSTER PRIMARY KEY ("id")
);


CREATE TABLE IF NOT EXISTS "excel_column_type" (
  "id" bigint NOT NULL IDENTITY(1,1) COMMENT '主键id',
  "catalog" varchar(256 char) NOT NULL COMMENT '数据源',
  "vdm_catalog" varchar(256 char) DEFAULT NULL COMMENT 'vdm数据源',
  "schema_name" varchar(256 char) NOT NULL COMMENT '库名',
  "table_name" varchar(512 char) NOT NULL COMMENT '表名',
  "column_name" varchar(128 char) NOT NULL COMMENT '列名',
  "column_comment" varchar(512 char) DEFAULT NULL COMMENT '列注释',
  "type" varchar(128 char) NOT NULL COMMENT '字段类型',
  "order_no" int NOT NULL COMMENT '列序号',
  "create_time" timestamp NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  "update_time" timestamp NOT NULL DEFAULT current_timestamp() COMMENT '更新时间',
  CLUSTER PRIMARY KEY ("id")
);


CREATE TABLE IF NOT EXISTS "excel_table_config" (
  "id" bigint NOT NULL IDENTITY(1,1) COMMENT '主键id',
  "catalog" varchar(256 char) NOT NULL COMMENT '数据源',
  "vdm_catalog" varchar(256 char) DEFAULT NULL COMMENT 'vdm数据源',
  "schema_name" varchar(256 char) NOT NULL COMMENT '库名',
  "file_name" varchar(512 char) NOT NULL COMMENT 'excel文件名',
  "table_name" varchar(512 char) NOT NULL COMMENT '表名',
  "table_comment" varchar(512 char) DEFAULT NULL COMMENT '表注释',
  "sheet" varchar(128 char) DEFAULT NULL COMMENT 'sheet名称',
  "all_sheet" tinyint NOT NULL DEFAULT 0 COMMENT '是否加载所有sheet',
  "sheet_as_new_column" tinyint NOT NULL DEFAULT 0 COMMENT 'sheet是否作为列 1:是 0:否',
  "start_cell" varchar(32 char) DEFAULT NULL COMMENT '起始单元格',
  "end_cell" varchar(32 char) DEFAULT NULL COMMENT '结束单元格',
  "has_headers" tinyint NOT NULL DEFAULT 1 COMMENT '是否有表头  1：有； 0：没有',
  "create_time" timestamp NULL DEFAULT current_timestamp() COMMENT '创建时间',
  "update_time" timestamp NOT NULL DEFAULT current_timestamp() COMMENT '更新时间',
  CLUSTER PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "excel_table_config_vdm_table_uindex" on "excel_table_config" ("catalog","table_name");


CREATE TABLE IF NOT EXISTS "query_info" (
  "query_id" varchar(30 char) NOT NULL COMMENT 'query id',
  "result" text DEFAULT NULL COMMENT '查询结果集',
  "msg" varchar(500 char) DEFAULT NULL COMMENT '错误详情',
  "task_id" varchar(200 char) NOT NULL COMMENT '任务Id',
  "state" varchar(30 char) NOT NULL COMMENT '状态',
  "create_time" varchar(30 char) NOT NULL COMMENT '创建时间',
  "update_time" varchar(30 char) NOT NULL COMMENT '更新时间',
  CLUSTER PRIMARY KEY ("query_id")
);


CREATE TABLE IF NOT EXISTS "task_info" (
  "task_id" varchar(200 char) NOT NULL COMMENT '主键taskid',
  "state" varchar(30 char) DEFAULT NULL COMMENT 'task状态',
  "query" text DEFAULT NULL,
  "create_time" varchar(30 char) DEFAULT NULL COMMENT '创建时间',
  "update_time" varchar(30 char) DEFAULT NULL COMMENT '修改时间',
  "topic" varchar(100 char) DEFAULT NULL COMMENT 'topic名称',
  "sub_task_id" varchar(200 char) NOT NULL COMMENT '子任务Id',
  "type" int NOT NULL DEFAULT 1 COMMENT '类型,0:异步查询,1:字段探查',
  "elapsed_time" varchar(30 char) NOT NULL COMMENT '总耗时',
  "update_count" text NOT NULL COMMENT '结果集大小,只针对insert into或create table as记录大小',
  "schedule_time" varchar(30 char) NOT NULL COMMENT '调度耗时',
  "queued_time" varchar(30 char) NOT NULL COMMENT '队列耗时',
  "cpu_time" varchar(30 char) NOT NULL COMMENT 'cpu耗时',
  CLUSTER PRIMARY KEY ("task_id","sub_task_id")
);

