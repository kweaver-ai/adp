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
--   - comment: 字段描述
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
--     "comment": "主键ID",
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
--     "comment": "文章内容",
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
--     "comment": "向量嵌入",
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

-- ==========================================
-- 1. t_catalog 主表
-- ==========================================
CREATE TABLE t_catalog (
    -- 主键与基础信息
    f_id                      VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'catalog唯一标识',
    f_name                    VARCHAR(128) NOT NULL DEFAULT '' COMMENT '目录名称，系统一级命名空间',
    f_tags                    VARCHAR(255) NOT NULL DEFAULT '[]' COMMENT '标签，逗号分隔，用于分类和检索',
    f_comment                 VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '目录描述',

    f_type                    VARCHAR(20) NOT NULL DEFAULT '' COMMENT '目录类型: physical, logical',
    f_enabled                 BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否启用',

    -- Physical Catalog 专属字段
    f_connector_type          VARCHAR(50) NOT NULL DEFAULT '' COMMENT '数据源类型: mysql, postgresql, s3, kafka, elasticsearch, api, prometheus, etc.',
    f_connector_config       MEDIUMTEXT NOT NULL COMMENT '加密存储的连接配置（JSON格式）',

    -- 状态管理
    f_health_check_enabled    BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否启用健康检查',
    f_health_check_status     VARCHAR(20) NOT NULL DEFAULT 'healthy' COMMENT '连接状态: healthy, degraded, unhealthy, offline, disabled',
    f_last_check_time         BIGINT(20) NOT NULL DEFAULT 0 COMMENT '最后健康检查时间',
    f_health_check_result     TEXT NOT NULL COMMENT '健康检查结果',

    -- 审计字段
    f_creator                 VARCHAR(128) COMMENT '创建者id',
    f_creator_type            VARCHAR(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
    f_create_time             BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
    f_updater                 VARCHAR(128) COMMENT '更新者id',
    f_updater_type            VARCHAR(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
    f_update_time             BIGINT(20) NOT NULL DEFAULT 0 COMMENT '更新时间',

    -- 索引
    PRIMARY KEY (f_id),
    UNIQUE INDEX uk_name (f_name),
    INDEX idx_type (f_type),
    INDEX idx_connector_type (f_connector_type),
    INDEX idx_health_check_status (f_health_check_status)
)  ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT='目录表，管理数据源连接和命名空间';


-- ==========================================
-- 2. t_catalog_discovery_policy 发现与变更策略表
-- ==========================================
CREATE TABLE t_catalog_discovery_policy (
    f_id                      VARCHAR(40) NOT NULL DEFAULT '' COMMENT '所属catalog ID',

    -- 状态
    f_enabled                 BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否启用',

    -- 发现策略配置
    f_discovery_mode          VARCHAR(20) NOT NULL DEFAULT 'manual' COMMENT '数据资源发现模式: manual, scheduled, event_driven',
    f_discovery_cron          VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'scheduled模式的cron表达式',
    f_discovery_config        MEDIUMTEXT NOT NULL COMMENT '发现策略详细配置',

    -- 变更处理策略
    f_on_resource_added       VARCHAR(20) NOT NULL DEFAULT 'auto_register' COMMENT '新增数据资源策略: auto_register, pending_review, ignore',
    f_on_resource_removed     VARCHAR(20) NOT NULL DEFAULT 'mark_stale' COMMENT '删除数据资源策略: auto_remove, mark_stale, ignore',
    f_on_schema_changed       VARCHAR(20) NOT NULL DEFAULT 'auto_update' COMMENT 'Schema变更策略: auto_update, pending_review, ignore',
    f_on_file_content_changed VARCHAR(20) NOT NULL DEFAULT 'pending_review' COMMENT '文件内容变更策略: pending_review, ignore',

    -- 策略详细配置
    f_change_policy_config    MEDIUMTEXT NOT NULL COMMENT '变更策略详细配置（如通知设置、审批流程等）',

    -- 索引
    PRIMARY KEY (f_id),
    INDEX idx_discovery_mode (f_discovery_mode),
    INDEX idx_enabled (f_enabled)
)  ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT='Catalog发现与变更策略配置表';


-- ==========================================
-- 3. t_resource 数据资源主表
-- ==========================================
CREATE TABLE t_resource (
    -- 主键与基础信息
    f_id                      VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'resource唯一标识',
    f_catalog_id              VARCHAR(40) NOT NULL DEFAULT '' COMMENT '所属catalog ID',
    f_name                    VARCHAR(128) NOT NULL DEFAULT '' COMMENT '数据资源名称，catalog下唯一',
    f_comment                 VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '数据资源描述',
    f_tags                    VARCHAR(255) NOT NULL DEFAULT '[]' COMMENT '标签，JSON数组格式',

    f_type                    VARCHAR(20) NOT NULL DEFAULT '' COMMENT '数据资源类型: table, file, fileset, api, metric, topic, index, logicview, dataset',

    -- 物理数据资源专属字段
    f_source_identifier       VARCHAR(500) NOT NULL DEFAULT '' COMMENT '源端标识(表名/文件路径/索引名等)',
    f_source_config           MEDIUMTEXT NOT NULL COMMENT '源端特定配置（JSON格式）',

    -- Schema相关
    f_database                VARCHAR(128) NOT NULL DEFAULT '' COMMENT '所属数据库名称（实例级连接时使用）',
    f_schema_definition       MEDIUMTEXT NOT NULL COMMENT 'Schema定义（JSON数组格式，包含所有字段信息）',
    f_schema_version          VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'Schema版本号',
    f_schema_inferred         BOOLEAN NOT NULL DEFAULT FALSE COMMENT 'Schema是否为自动推导',
    f_schema_last_updated     BIGINT(20) NOT NULL DEFAULT 0 COMMENT 'Schema最后更新时间',

    -- LogicView 专属字段
    f_logic_type              VARCHAR(20) NOT NULL DEFAULT '' COMMENT '逻辑类型: derived(衍生), composite(复合), 仅LogicView使用',
    f_logic_definition        MEDIUMTEXT NOT NULL COMMENT '逻辑定义（SQL/声明式映射/脚本），仅LogicView使用',
    f_logic_definition_type   VARCHAR(20) NOT NULL DEFAULT '' COMMENT '定义类型: sql, mapping, script',

    -- Local查询配置（物化）
    f_local_enabled           BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否启用Local查询（物化）',
    f_local_storage_engine    VARCHAR(50) NOT NULL DEFAULT '' COMMENT '物化存储引擎: elasticsearch, opensearch, lancedb, pgvector',
    f_local_storage_config    MEDIUMTEXT NOT NULL COMMENT '物化存储配置（JSON格式）',
    f_local_index_name        VARCHAR(255) NOT NULL DEFAULT '' COMMENT '物化后的索引名称',

    -- 同步配置
    f_sync_strategy           VARCHAR(20) NOT NULL DEFAULT '' COMMENT '同步策略: cdc, bulk_load, etl_pipeline, polling, micro_batch, reindex, snapshot',
    f_sync_config             MEDIUMTEXT NOT NULL COMMENT '同步配置（JSON格式：调度周期、批次大小等）',
    f_sync_status             VARCHAR(20) NOT NULL DEFAULT 'not_synced' COMMENT '同步状态: not_synced, syncing, synced, failed',
    f_last_sync_time          BIGINT(20) NOT NULL DEFAULT 0 COMMENT '最后同步时间',
    f_sync_error_message      TEXT NOT NULL COMMENT '同步错误信息',

    -- 状态管理
    f_status                  VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '数据资源状态: active, disabled, deprecated, stale',
    f_status_message          VARCHAR(500) NOT NULL DEFAULT '' COMMENT '状态说明',

    -- 审计字段
    f_creator                 VARCHAR(128) COMMENT '创建者id',
    f_creator_type            VARCHAR(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
    f_create_time             BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
    f_updater                 VARCHAR(128) COMMENT '更新者id',
    f_updater_type            VARCHAR(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
    f_update_time             BIGINT(20) NOT NULL DEFAULT 0 COMMENT '更新时间',

    -- 索引
    PRIMARY KEY (f_id),
    UNIQUE INDEX uk_catalog_name (f_catalog_id, f_name),
    INDEX idx_type (f_type),
    INDEX idx_status (f_status)
)  ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT='数据资源主表，管理所有类型的数据资源';


-- ==========================================
-- 4. t_resource_schema_history Schema历史表
-- ==========================================
CREATE TABLE t_resource_schema_history (
    f_id                      VARCHAR(40) NOT NULL DEFAULT '' COMMENT '历史记录唯一标识',
    f_resource_id             VARCHAR(40) NOT NULL DEFAULT '' COMMENT '所属resource ID',
    f_schema_version          VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'Schema版本号',
    f_schema_definition       MEDIUMTEXT NOT NULL COMMENT 'Schema定义快照（JSON数组格式）',

    -- 变更信息
    f_change_type             VARCHAR(20) NOT NULL DEFAULT '' COMMENT '变更类型: created, field_added, field_removed, field_modified, type_changed',
    f_change_summary          VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '变更摘要',
    f_schema_inferred         BOOLEAN NOT NULL DEFAULT FALSE COMMENT 'Schema是否为自动推导',
    f_change_time             BIGINT(20) NOT NULL DEFAULT 0 COMMENT '变更时间',

    -- 索引
    PRIMARY KEY (f_id),
    INDEX idx_resource_id (f_resource_id)
)  ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT='数据资源Schema历史表，记录Schema变更历史';


-- ==========================================
-- 5. t_connector_type Connector 类型注册表
-- ==========================================
CREATE TABLE t_connector_type (
    -- 主键与基础信息
    f_type                    VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'connector类型,唯一标识',
    f_name                    VARCHAR(128) NOT NULL DEFAULT '' COMMENT '类型名称: mysql, postgresql, kafka...',
    f_comment                 VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '类型描述',

    -- 类型分类
    f_mode                    VARCHAR(20) NOT NULL DEFAULT '' COMMENT '模式: local, remote',
    f_category                VARCHAR(32) NOT NULL DEFAULT '' COMMENT '分类: table, index, topic, file, api',

    -- Remote 模式专用字段
    f_endpoint                VARCHAR(512) NOT NULL DEFAULT '' COMMENT '远程服务地址 (仅remote模式)',

    -- 状态
    f_enabled                 BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否启用',

    -- 索引
    PRIMARY KEY (f_type),
    UNIQUE INDEX uk_name (f_name),
    INDEX idx_mode (f_mode),
    INDEX idx_category (f_category),
    INDEX idx_enabled (f_enabled)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT='Connector类型注册表';


-- ==========================================
-- 6. 初始化内置 Local Connector
-- ==========================================
INSERT INTO t_connector_type (f_type, f_name, f_comment, f_mode, f_category, f_enabled)
SELECT 'mysql', 'mysql', 'MySQL 关系型数据库连接器', 'local', 'table', TRUE
FROM DUAL WHERE NOT EXISTS ( SELECT f_type FROM t_connector_type WHERE f_type = 'mysql' );

INSERT INTO t_connector_type (f_type, f_name, f_comment, f_mode, f_category, f_enabled)
SELECT 'opensearch', 'opensearch', 'OpenSearch 搜索引擎连接器', 'local', 'index', TRUE
FROM DUAL WHERE NOT EXISTS ( SELECT f_type FROM t_connector_type WHERE f_type = 'opensearch' );
