SET SCHEMA adp;

CREATE TABLE IF NOT EXISTS "t_metadata_api" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_summary" VARCHAR(256 CHAR) NOT NULL,
    "f_version" VARCHAR(40 CHAR) NOT NULL,
    "f_svc_url" text NOT NULL,
    "f_description" text,
    "f_path" text NOT NULL,
    "f_method" VARCHAR(50 CHAR) NOT NULL,
    "f_api_spec" text DEFAULT NULL,
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_update_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_metadata_api_uk_version ON t_metadata_api(f_version);

CREATE TABLE IF NOT EXISTS "t_op_registry" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_op_id" VARCHAR(40 CHAR) NOT NULL,
    "f_name" VARCHAR(512 CHAR) NOT NULL,
    "f_metadata_version" VARCHAR(40 CHAR) NOT NULL,
    "f_metadata_type" VARCHAR(40 CHAR) NOT NULL,
    "f_status" VARCHAR(10 CHAR) DEFAULT 0,
    "f_operator_type" VARCHAR(10 CHAR) DEFAULT 0,
    "f_execution_mode" VARCHAR(10 CHAR) DEFAULT 0,
    "f_category" VARCHAR(50 CHAR) DEFAULT 0,
    "f_source" VARCHAR(50 CHAR) DEFAULT '',
    "f_execute_control" text DEFAULT NULL,
    "f_extend_info" text DEFAULT NULL,
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_user" VARCHAR(50 CHAR) NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    "f_is_deleted" TINYINT DEFAULT 0,
    "f_is_internal" TINYINT DEFAULT 0,
    "f_is_data_source" TINYINT DEFAULT 0,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_op_registry_uk_op_id_version ON t_op_registry(f_op_id, f_metadata_version);

CREATE INDEX IF NOT EXISTS t_op_registry_idx_name_update ON t_op_registry(f_name, f_update_time);

CREATE INDEX IF NOT EXISTS t_op_registry_idx_status_update ON t_op_registry(f_status, f_update_time);

CREATE INDEX IF NOT EXISTS t_op_registry_idx_category_update ON t_op_registry(f_category, f_update_time);

CREATE INDEX IF NOT EXISTS t_op_registry_idx_create_user_update ON t_op_registry(f_create_user, f_update_time);

CREATE INDEX IF NOT EXISTS t_op_registry_idx_update_time ON t_op_registry(f_update_time);

CREATE TABLE IF NOT EXISTS "t_toolbox" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_box_id" VARCHAR(40 CHAR) NOT NULL,
    "f_name" VARCHAR(50 CHAR) NOT NULL,
    "f_description" text NOT NULL,
    "f_svc_url" text NOT NULL,
    "f_status" VARCHAR(50 CHAR) NOT NULL,
    "f_is_internal" TINYINT DEFAULT 0,
    "f_source" VARCHAR(50 CHAR) DEFAULT '',
    "f_category" VARCHAR(50 CHAR) DEFAULT 0,
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_user" VARCHAR(50 CHAR) NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    "f_release_user" VARCHAR(50 CHAR) NOT NULL,
    "f_release_time" BIGINT NOT NULL,
    "f_metadata_type" VARCHAR(50 CHAR) NOT NULL DEFAULT 'openapi',
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_toolbox_uk_box_id ON t_toolbox(f_box_id);

CREATE INDEX IF NOT EXISTS t_toolbox_idx_name ON t_toolbox(f_name);

CREATE INDEX IF NOT EXISTS t_toolbox_idx_status ON t_toolbox(f_status);

CREATE INDEX IF NOT EXISTS t_toolbox_idx_category ON t_toolbox(f_category);

CREATE INDEX IF NOT EXISTS t_toolbox_idx_creator_status ON t_toolbox(f_create_user, f_status);

CREATE INDEX IF NOT EXISTS t_toolbox_idx_ctime ON t_toolbox(f_create_time);

CREATE INDEX IF NOT EXISTS t_toolbox_idx_utime ON t_toolbox(f_update_time);

CREATE TABLE IF NOT EXISTS "t_tool" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_tool_id" VARCHAR(40 CHAR) NOT NULL,
    "f_box_id" VARCHAR(40 CHAR) NOT NULL,
    "f_name" VARCHAR(256 CHAR) NOT NULL,
    "f_description" text NOT NULL,
    "f_source_type" VARCHAR(50 CHAR) NOT NULL,
    "f_source_id" VARCHAR(40 CHAR) NOT NULL,
    "f_status" VARCHAR(40 CHAR) DEFAULT 0,
    "f_use_count" BIGINT NOT NULL,
    "f_use_rule" text DEFAULT NULL,
    "f_parameters" text DEFAULT NULL,
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_user" VARCHAR(50 CHAR) NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    "f_extend_info" text DEFAULT NULL,
    "f_is_deleted" TINYINT DEFAULT 0,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_tool_uk_tool_id ON t_tool(f_tool_id);

CREATE INDEX IF NOT EXISTS t_tool_idx_box_id ON t_tool(f_box_id);

CREATE INDEX IF NOT EXISTS t_tool_idx_name_update ON t_tool(f_name, f_update_time);

CREATE TABLE IF NOT EXISTS "t_mcp_server_config" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_mcp_id" VARCHAR(40 CHAR) NOT NULL,
    "f_name" VARCHAR(50 CHAR) NOT NULL,
    "f_description" text NOT NULL,
    "f_mode" VARCHAR(32 CHAR) NOT NULL,
    "f_url" text NOT NULL,
    "f_headers" text NOT NULL,
    "f_command" text NOT NULL,
    "f_env" text NOT NULL,
    "f_args" text NOT NULL,
    "f_status" VARCHAR(30 CHAR) NOT NULL DEFAULT 'unpublish',
    "f_is_internal" TINYINT DEFAULT 0,
    "f_source" VARCHAR(50 CHAR) DEFAULT 0,
    "f_category" VARCHAR(50 CHAR) DEFAULT 0,
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_user" VARCHAR(50 CHAR) NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    "f_creation_type" VARCHAR(20 CHAR) NOT NULL DEFAULT 'custom',
    "f_version" INT DEFAULT 1,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "t_mcp_server_config_idx_update_time" ON "t_mcp_server_config"(f_update_time);

CREATE INDEX IF NOT EXISTS "t_mcp_server_config_idx_status" ON "t_mcp_server_config"(f_status);

CREATE UNIQUE INDEX IF NOT EXISTS t_mcp_server_config_uk_mcp_id ON t_mcp_server_config(f_mcp_id);

CREATE INDEX IF NOT EXISTS t_mcp_server_config_idx_name ON t_mcp_server_config("f_name");

CREATE TABLE IF NOT EXISTS "t_mcp_server_release" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_mcp_id" VARCHAR(40 CHAR) NOT NULL,
    "f_name" VARCHAR(50 CHAR) NOT NULL,
    "f_description" text NOT NULL,
    "f_mode" VARCHAR(32 CHAR) NOT NULL,
    "f_url" text NOT NULL,
    "f_headers" text NOT NULL,
    "f_command" text NOT NULL,
    "f_env" text NOT NULL,
    "f_args" text NOT NULL,
    "f_status" VARCHAR(30 CHAR) NOT NULL DEFAULT 'draft',
    "f_is_internal" TINYINT DEFAULT 0,
    "f_source" VARCHAR(50 CHAR) DEFAULT 0,
    "f_category" VARCHAR(50 CHAR) DEFAULT 0,
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_user" VARCHAR(50 CHAR) NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    "f_version" INT NOT NULL,
    "f_release_desc" VARCHAR(50 CHAR) NOT NULL,
    "f_release_user" VARCHAR(50 CHAR) NOT NULL,
    "f_release_time" BIGINT NOT NULL,
    "f_creation_type" VARCHAR(20 CHAR) NOT NULL DEFAULT 'custom',
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "t_mcp_server_release_idx_status_update_time" ON "t_mcp_server_release"(f_status, f_update_time);

CREATE UNIQUE INDEX IF NOT EXISTS t_mcp_server_release_uk_mcp ON t_mcp_server_release(f_mcp_id, f_version);

CREATE INDEX IF NOT EXISTS t_mcp_server_release_idx_mcp_id_create_time ON t_mcp_server_release("f_mcp_id", "f_create_time");

CREATE TABLE IF NOT EXISTS "t_mcp_server_release_history" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_mcp_id" VARCHAR(40 CHAR) NOT NULL,
    "f_mcp_release" text NOT NULL,
    "f_version" INT NOT NULL,
    "f_release_desc" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_user" VARCHAR(50 CHAR) NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_mcp_server_release_history_uk_mcp ON t_mcp_server_release_history(f_mcp_id, f_version);

CREATE INDEX IF NOT EXISTS t_mcp_server_release_history_idx_mcp_id_create_time ON t_mcp_server_release_history("f_mcp_id", "f_create_time");

CREATE TABLE IF NOT EXISTS "t_internal_component_config" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_component_type" VARCHAR(50 CHAR) NOT NULL,
    "f_component_id" VARCHAR(40 CHAR) NOT NULL,
    "f_config_version" VARCHAR(40 CHAR) NOT NULL,
    "f_config_source" VARCHAR(40 CHAR) NOT NULL,
    "f_protected_flag" TINYINT DEFAULT 0,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_internal_component_config_uk_comp_type_id ON t_internal_component_config("f_component_type","f_component_id");

CREATE TABLE IF NOT EXISTS "t_operator_release" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_op_id" VARCHAR(40 CHAR) NOT NULL,
    "f_name" VARCHAR(512 CHAR) NOT NULL,
    "f_metadata_version" VARCHAR(40 CHAR) NOT NULL,
    "f_metadata_type" VARCHAR(40 CHAR) NOT NULL,
    "f_status" VARCHAR(10 CHAR) DEFAULT 0,
    "f_operator_type" VARCHAR(10 CHAR) DEFAULT 0,
    "f_execution_mode" VARCHAR(10 CHAR) DEFAULT 0,
    "f_category" VARCHAR(50 CHAR) DEFAULT 0,
    "f_source" VARCHAR(50 CHAR) DEFAULT '',
    "f_execute_control" text DEFAULT NULL,
    "f_extend_info" text DEFAULT NULL,
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_user" VARCHAR(50 CHAR) NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    "f_tag" INT NOT NULL,
    "f_release_user" VARCHAR(50 CHAR) NOT NULL,
    "f_release_time" BIGINT NOT NULL,
    "f_is_internal" TINYINT DEFAULT 0,
    "f_is_data_source" TINYINT DEFAULT 0,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "t_operator_release_idx_status_update_time" ON "t_operator_release"(f_status, f_update_time);

CREATE UNIQUE INDEX IF NOT EXISTS t_operator_release_uk_op ON t_operator_release(f_op_id, f_tag);

CREATE INDEX IF NOT EXISTS t_operator_release_idx_op_id_create_time ON t_operator_release("f_op_id", "f_create_time");

CREATE TABLE IF NOT EXISTS "t_operator_release_history" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_op_id" VARCHAR(40 CHAR) NOT NULL,
    "f_op_release" text NOT NULL,
    "f_metadata_version" VARCHAR(40 CHAR) NOT NULL,
    "f_metadata_type" VARCHAR(40 CHAR) NOT NULL,
    "f_tag" INT NOT NULL,
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_user" VARCHAR(50 CHAR) NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_operator_release_history_uk_op ON t_operator_release_history(f_op_id, f_tag);

CREATE INDEX IF NOT EXISTS t_operator_release_history_idx_op_id_create_time ON t_operator_release_history("f_op_id", "f_create_time");

CREATE INDEX IF NOT EXISTS t_operator_release_history_idx_op_id_metadata_version ON t_operator_release_history("f_op_id", "f_metadata_version");

CREATE TABLE IF NOT EXISTS "t_category" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_category_id" VARCHAR(40 CHAR) NOT NULL,
    "f_category_name" VARCHAR(50 CHAR) NOT NULL,
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_user" VARCHAR(50 CHAR) NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_category_uk_category_id ON t_category(f_category_id);

CREATE UNIQUE INDEX IF NOT EXISTS t_category_uk_category_name ON t_category(f_category_name);

CREATE TABLE IF NOT EXISTS "t_outbox_message" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_event_id" VARCHAR(40 CHAR) NOT NULL,
    "f_event_type" VARCHAR(40 CHAR) NOT NULL,
    "f_topic" text NOT NULL,
    "f_payload" text NOT NULL,
    "f_status" VARCHAR(40 CHAR) NOT NULL,
    "f_created_at" BIGINT NOT NULL,
    "f_updated_at" BIGINT NOT NULL,
    "f_next_retry_at" BIGINT NOT NULL,
    "f_retry_count" INT NOT NULL,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_outbox_message_uk_event_id ON t_outbox_message(f_event_id);

CREATE INDEX IF NOT EXISTS t_outbox_message_idx_event_type ON t_outbox_message(f_event_type);

CREATE INDEX IF NOT EXISTS t_outbox_message_idx_status_next_retry ON t_outbox_message(f_status, f_next_retry_at);

CREATE TABLE IF NOT EXISTS "t_mcp_tool" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_mcp_tool_id" VARCHAR(40 CHAR) NOT NULL,
    "f_mcp_id" VARCHAR(40 CHAR) NOT NULL,
    "f_mcp_version" INT NOT NULL,
    "f_box_id" VARCHAR(40 CHAR) NOT NULL,
    "f_box_name" VARCHAR(50 CHAR),
    "f_tool_id" VARCHAR(40 CHAR) NOT NULL,
    "f_name" VARCHAR(256 CHAR),
    "f_description" text,
    "f_use_rule" text,
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_user" VARCHAR(50 CHAR) NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_mcp_tool_uk_mcp_tool_id ON t_mcp_tool(f_mcp_tool_id);

CREATE INDEX IF NOT EXISTS t_mcp_tool_idx_mcp_id_version ON t_mcp_tool(f_mcp_id, f_mcp_version);

CREATE INDEX IF NOT EXISTS t_mcp_tool_idx_name_update ON t_mcp_tool(f_name, f_update_time);


CREATE TABLE IF NOT EXISTS "t_metadata_function" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_summary" VARCHAR(256 CHAR) NOT NULL,
    "f_version" VARCHAR(40 CHAR) NOT NULL,
    "f_svc_url" text NOT NULL,
    "f_description" text,
    "f_path" text NOT NULL,
    "f_method" VARCHAR(50 CHAR) NOT NULL,
    "f_code" text NOT NULL,
    "f_script_type" VARCHAR(50 CHAR) NOT NULL,
    "f_dependencies" text DEFAULT NULL,
    "f_api_spec" text DEFAULT NULL,
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_update_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    "f_dependencies_url" text,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_metadata_function_uk_version ON t_metadata_function(f_version);


CREATE TABLE IF NOT EXISTS "t_resource_deploy" (
    "f_id" BIGINT IDENTITY(1, 1) NOT NULL,
    "f_resource_id" VARCHAR(40 CHAR) NOT NULL,
    "f_type" VARCHAR(40 CHAR) NOT NULL,
    "f_version" INT NOT NULL,
    "f_name" VARCHAR(40 CHAR) NOT NULL,
    "f_description" text NOT NULL,
    "f_config" text NOT NULL,
    "f_status" VARCHAR(40 CHAR) NOT NULL,
    "f_create_user" VARCHAR(50 CHAR) NOT NULL,
    "f_create_time" BIGINT NOT NULL,
    "f_update_user" VARCHAR(50 CHAR) NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_resource_deploy_uk_resource_id ON t_resource_deploy(f_resource_id, f_type, f_version);


CREATE TABLE IF NOT EXISTS `t_skill_repository` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_skill_id` VARCHAR(40) NOT NULL COMMENT 'Skill ID',
  `f_name` VARCHAR(255) NOT NULL COMMENT 'Skill 名称',
  `f_description` LONGTEXT NOT NULL COMMENT 'Skill 描述',
  `f_skill_content` LONGTEXT NOT NULL COMMENT 'Skill 指令正文',
  `f_version` VARCHAR(40) NOT NULL COMMENT 'Skill 版本',
  `f_status` VARCHAR(40) NOT NULL COMMENT 'Skill 状态',
  `f_source` VARCHAR(50) NOT NULL DEFAULT '' COMMENT 'Skill 来源',
  `f_extend_info` TEXT DEFAULT NULL COMMENT '扩展信息',
  `f_dependencies` TEXT DEFAULT NULL COMMENT '依赖信息',
  `f_file_manifest` LONGTEXT DEFAULT NULL COMMENT '文件摘要清单',
  `f_create_user` VARCHAR(50) NOT NULL COMMENT '创建者',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_user` VARCHAR(50) NOT NULL COMMENT '编辑者',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '编辑时间',
  `f_delete_user` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '删除者',
  `f_delete_time` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '删除时间',
  `f_category` VARCHAR(50 CHAR) DEFAULT '' COMMENT '工具箱分类, 数据处理/算法模型',
  `f_is_deleted` BOOLEAN DEFAULT 0 COMMENT '是否删除', -- 0: 未删除, 1: 待删除
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_skill_repository_uk_skill_id` (f_skill_id)
);

CREATE INDEX IF NOT EXISTS `idx_t_skill_repository_idx_status_update_time` ON `t_skill_repository` (f_status, f_update_time);
CREATE INDEX IF NOT EXISTS `idx_t_skill_repository_idx_category_update_time` ON `t_skill_repository` (f_category, f_update_time);
CREATE INDEX IF NOT EXISTS `idx_t_skill_repository_idx_create_user_update_time` ON `t_skill_repository` (f_create_user, f_update_time);


CREATE TABLE IF NOT EXISTS `t_skill_file_index` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_skill_id` VARCHAR(40) NOT NULL COMMENT 'Skill ID',
  `f_skill_version` VARCHAR(40) NOT NULL COMMENT 'Skill 版本',
  `f_rel_path` VARCHAR(512) NOT NULL COMMENT '文件相对路径',
  `f_path_hash` VARCHAR(32) NOT NULL COMMENT '相对路径哈希',
  `f_storage_id` VARCHAR(50) NOT NULL COMMENT '对象存储ID',
  `f_storage_key` TEXT NOT NULL COMMENT '对象存储键',
  `f_file_type` VARCHAR(40) NOT NULL COMMENT '文件类型',
  `f_content_sha256` VARCHAR(64) NOT NULL COMMENT '文件内容 SHA256',
  `f_mime_type` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'MIME 类型',
  `f_size` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '文件大小',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '编辑时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_skill_file_index_uk_skill_version_rel_path` (f_skill_id, f_skill_version, f_rel_path),
  UNIQUE KEY `idx_t_skill_file_index_uk_skill_version_path_hash` (f_skill_id, f_skill_version, f_path_hash)
);
