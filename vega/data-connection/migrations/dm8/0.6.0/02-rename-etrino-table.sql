-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：在 kweaver schema 下创建 etrino 相关表，并从 adp schema 复制数据
-- ==========================================


-- 在adp中追加etrino模块数据库初始化，对应0.5.0升级至0.6.0场景
SET SCHEMA adp;
CREATE TABLE IF NOT EXISTS "catalog_rule" (
  "id" bigint NOT NULL IDENTITY(1,1) COMMENT '主键id',
  "catalog_name" varchar(200 char) NOT NULL COMMENT 'catalog名称',
  "datasource_type" varchar(50 char) NOT NULL COMMENT '数据源类型',
  "pushdown_rule" varchar(500 char) DEFAULT NULL COMMENT '下推规则',
  "is_enabled" varchar(20 char) NOT NULL COMMENT '是否启用规则',
  "create_time" varchar(30 char) DEFAULT NULL COMMENT '创建时间',
  "update_time" varchar(30 char) DEFAULT NULL COMMENT '修改时间',
  CLUSTER PRIMARY KEY ("id")
);


CREATE TABLE IF NOT EXISTS "hetu_ctlgs" (
  "id" bigint NOT NULL IDENTITY(1,1),
  "catalog_name" varchar(256 char) NOT NULL,
  "create_time" bigint NOT NULL,
  "owner" varchar(767 char) DEFAULT NULL,
  "comment" varchar(256 char) DEFAULT NULL,
  CLUSTER PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "catalog_name" on "hetu_ctlgs" ("catalog_name");


CREATE TABLE IF NOT EXISTS "hetu_dbs" (
  "id" bigint NOT NULL IDENTITY(1,1),
  "catalog_name" varchar(256 char) NOT NULL,
  "database_name" varchar(256 char) NOT NULL,
  "create_time" bigint NOT NULL,
  "owner" varchar(767 char) DEFAULT NULL,
  "comment" varchar(256 char) DEFAULT NULL,
  CLUSTER PRIMARY KEY ("id"),
  CONSTRAINT "hetu_dbs_ibfk_1" FOREIGN KEY ("catalog_name") REFERENCES "hetu_ctlgs" ("catalog_name")
);
CREATE UNIQUE INDEX IF NOT EXISTS "catalog_name" on "hetu_dbs" ("catalog_name","database_name");


CREATE TABLE IF NOT EXISTS "hetu_tbls" (
  "id" bigint NOT NULL IDENTITY(1,1),
  "database_id" bigint NOT NULL,
  "table_name" varchar(256 char) NOT NULL,
  "type" varchar(128 char) NOT NULL,
  "view_original_text" text DEFAULT NULL,
  "create_time" bigint NOT NULL,
  "owner" varchar(767 char) DEFAULT NULL,
  "comment" varchar(4000 char) DEFAULT NULL,
  CLUSTER PRIMARY KEY ("id"),
  CONSTRAINT "hetu_tbls_ibfk_1" FOREIGN KEY ("database_id") REFERENCES "hetu_dbs" ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "database_id" on "hetu_tbls" ("database_id","table_name");


CREATE TABLE IF NOT EXISTS "hetu_tab_cols" (
  "id" bigint NOT NULL IDENTITY(1,1),
  "table_id" bigint NOT NULL,
  "column_name" varchar(256 char) NOT NULL,
  "type" varchar(128 char) NOT NULL,
  "comment" varchar(4000 char) DEFAULT NULL,
  CLUSTER PRIMARY KEY ("id"),
  CONSTRAINT "hetu_tab_cols_ibfk_1" FOREIGN KEY ("table_id") REFERENCES "hetu_tbls" ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "table_id" on "hetu_tab_cols" ("table_id","column_name");


CREATE TABLE IF NOT EXISTS "hetu_tab_lock" (
  "id" bigint NOT NULL IDENTITY(1,1),
  "resource" int NOT NULL,
  "description" varchar(128 char) NOT NULL,
  CLUSTER PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "resource" on "hetu_tab_lock" ("resource");


CREATE TABLE IF NOT EXISTS "hetu_catalog_params" (
  "catalog_id" bigint NOT NULL,
  "param_key" varchar(256 char) NOT NULL,
  "param_value" text DEFAULT NULL,
  CLUSTER PRIMARY KEY ("catalog_id","param_key"),
  CONSTRAINT "hetu_catalog_params_ibfk_1" FOREIGN KEY ("catalog_id") REFERENCES "hetu_ctlgs" ("id")
);


CREATE TABLE IF NOT EXISTS "hetu_database_params" (
  "database_id" bigint NOT NULL,
  "param_key" varchar(256 char) NOT NULL,
  "param_value" text DEFAULT NULL,
  CLUSTER PRIMARY KEY ("database_id","param_key"),
  CONSTRAINT "hetu_database_params_ibfk_1" FOREIGN KEY ("database_id") REFERENCES "hetu_dbs" ("id")
);


CREATE TABLE IF NOT EXISTS "hetu_table_params" (
  "table_id" bigint NOT NULL,
  "param_key" varchar(256 char) NOT NULL,
  "param_value" text DEFAULT NULL,
  CLUSTER PRIMARY KEY ("table_id","param_key"),
  CONSTRAINT "hetu_table_params_ibfk_1" FOREIGN KEY ("table_id") REFERENCES "hetu_tbls" ("id")
);


CREATE TABLE IF NOT EXISTS "hetu_column_params" (
  "column_id" bigint NOT NULL,
  "param_key" varchar(256 char) NOT NULL,
  "param_value" text DEFAULT NULL,
  CLUSTER PRIMARY KEY ("column_id","param_key"),
  CONSTRAINT "hetu_column_params_ibfk_1" FOREIGN KEY ("column_id") REFERENCES "hetu_tab_cols" ("id")
);


CREATE TABLE IF NOT EXISTS "hetu_favorite" (
  "creationTime" datetime DEFAULT current_timestamp(),
  "user" varchar(20 char) NOT NULL,
  "query" varchar(600 char) NOT NULL,
  "catalog" varchar(30 char) NOT NULL,
  "schemata" varchar(30 char) NOT NULL,
  "id" bigint NOT NULL IDENTITY(1,1),
  CLUSTER PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "favorite_uni" on "hetu_favorite" ("user","query","catalog","schemata");


CREATE TABLE IF NOT EXISTS "hetu_query_history" (
  "id" bigint NOT NULL IDENTITY(1,1),
  "user" varchar(20 char) NOT NULL,
  "source" varchar(30 char) NOT NULL,
  "queryId" varchar(50 char) NOT NULL,
  "resource" varchar(30 char) NOT NULL,
  "query" varchar(10000 char) NOT NULL,
  "state" varchar(30 char) NOT NULL,
  "failed" varchar(30 char) DEFAULT NULL,
  "createTime" varchar(30 char) DEFAULT NULL,
  "elapsedTime" varchar(30 char) NOT NULL,
  "cpuTime" varchar(30 char) NOT NULL,
  "executionTime" varchar(30 char) NOT NULL,
  "catalog" varchar(30 char) NOT NULL,
  "schemata" varchar(30 char) NOT NULL,
  "currentMemory" varchar(30 char) NOT NULL,
  "cumulativeUserMemory" double NOT NULL,
  "jsonString" text NOT NULL,
  "completedDrivers" int NOT NULL,
  "runningDrivers" int NOT NULL,
  "queuedDrivers" int NOT NULL,
  "totalCpuTime" varchar(50 char) NOT NULL,
  "totalMemoryReservation" varchar(50 char) NOT NULL,
  "peakTotalMemoryReservation" varchar(50 char) NOT NULL,
  CLUSTER PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "query_history_queryId_uindex" on "hetu_query_history" ("queryId");




SET SCHEMA kweaver;


CREATE TABLE IF NOT EXISTS "catalog_rule" (
  "id" bigint NOT NULL IDENTITY(1,1) COMMENT '主键id',
  "catalog_name" varchar(200 char) NOT NULL COMMENT 'catalog名称',
  "datasource_type" varchar(50 char) NOT NULL COMMENT '数据源类型',
  "pushdown_rule" varchar(500 char) DEFAULT NULL COMMENT '下推规则',
  "is_enabled" varchar(20 char) NOT NULL COMMENT '是否启用规则',
  "create_time" varchar(30 char) DEFAULT NULL COMMENT '创建时间',
  "update_time" varchar(30 char) DEFAULT NULL COMMENT '修改时间',
  CLUSTER PRIMARY KEY ("id")
);


CREATE TABLE IF NOT EXISTS "hetu_ctlgs" (
  "id" bigint NOT NULL IDENTITY(1,1),
  "catalog_name" varchar(256 char) NOT NULL,
  "create_time" bigint NOT NULL,
  "owner" varchar(767 char) DEFAULT NULL,
  "comment" varchar(256 char) DEFAULT NULL,
  CLUSTER PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "catalog_name" on "hetu_ctlgs" ("catalog_name");


CREATE TABLE IF NOT EXISTS "hetu_dbs" (
  "id" bigint NOT NULL IDENTITY(1,1),
  "catalog_name" varchar(256 char) NOT NULL,
  "database_name" varchar(256 char) NOT NULL,
  "create_time" bigint NOT NULL,
  "owner" varchar(767 char) DEFAULT NULL,
  "comment" varchar(256 char) DEFAULT NULL,
  CLUSTER PRIMARY KEY ("id"),
  CONSTRAINT "hetu_dbs_ibfk_1" FOREIGN KEY ("catalog_name") REFERENCES "hetu_ctlgs" ("catalog_name")
);
CREATE UNIQUE INDEX IF NOT EXISTS "catalog_name" on "hetu_dbs" ("catalog_name","database_name");


CREATE TABLE IF NOT EXISTS "hetu_tbls" (
  "id" bigint NOT NULL IDENTITY(1,1),
  "database_id" bigint NOT NULL,
  "table_name" varchar(256 char) NOT NULL,
  "type" varchar(128 char) NOT NULL,
  "view_original_text" text DEFAULT NULL,
  "create_time" bigint NOT NULL,
  "owner" varchar(767 char) DEFAULT NULL,
  "comment" varchar(4000 char) DEFAULT NULL,
  CLUSTER PRIMARY KEY ("id"),
  CONSTRAINT "hetu_tbls_ibfk_1" FOREIGN KEY ("database_id") REFERENCES "hetu_dbs" ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "database_id" on "hetu_tbls" ("database_id","table_name");


CREATE TABLE IF NOT EXISTS "hetu_tab_cols" (
  "id" bigint NOT NULL IDENTITY(1,1),
  "table_id" bigint NOT NULL,
  "column_name" varchar(256 char) NOT NULL,
  "type" varchar(128 char) NOT NULL,
  "comment" varchar(4000 char) DEFAULT NULL,
  CLUSTER PRIMARY KEY ("id"),
  CONSTRAINT "hetu_tab_cols_ibfk_1" FOREIGN KEY ("table_id") REFERENCES "hetu_tbls" ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "table_id" on "hetu_tab_cols" ("table_id","column_name");


CREATE TABLE IF NOT EXISTS "hetu_tab_lock" (
  "id" bigint NOT NULL IDENTITY(1,1),
  "resource" int NOT NULL,
  "description" varchar(128 char) NOT NULL,
  CLUSTER PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "resource" on "hetu_tab_lock" ("resource");


CREATE TABLE IF NOT EXISTS "hetu_catalog_params" (
  "catalog_id" bigint NOT NULL,
  "param_key" varchar(256 char) NOT NULL,
  "param_value" text DEFAULT NULL,
  CLUSTER PRIMARY KEY ("catalog_id","param_key"),
  CONSTRAINT "hetu_catalog_params_ibfk_1" FOREIGN KEY ("catalog_id") REFERENCES "hetu_ctlgs" ("id")
);


CREATE TABLE IF NOT EXISTS "hetu_database_params" (
  "database_id" bigint NOT NULL,
  "param_key" varchar(256 char) NOT NULL,
  "param_value" text DEFAULT NULL,
  CLUSTER PRIMARY KEY ("database_id","param_key"),
  CONSTRAINT "hetu_database_params_ibfk_1" FOREIGN KEY ("database_id") REFERENCES "hetu_dbs" ("id")
);


CREATE TABLE IF NOT EXISTS "hetu_table_params" (
  "table_id" bigint NOT NULL,
  "param_key" varchar(256 char) NOT NULL,
  "param_value" text DEFAULT NULL,
  CLUSTER PRIMARY KEY ("table_id","param_key"),
  CONSTRAINT "hetu_table_params_ibfk_1" FOREIGN KEY ("table_id") REFERENCES "hetu_tbls" ("id")
);


CREATE TABLE IF NOT EXISTS "hetu_column_params" (
  "column_id" bigint NOT NULL,
  "param_key" varchar(256 char) NOT NULL,
  "param_value" text DEFAULT NULL,
  CLUSTER PRIMARY KEY ("column_id","param_key"),
  CONSTRAINT "hetu_column_params_ibfk_1" FOREIGN KEY ("column_id") REFERENCES "hetu_tab_cols" ("id")
);


CREATE TABLE IF NOT EXISTS "hetu_favorite" (
  "creationTime" datetime DEFAULT current_timestamp(),
  "user" varchar(20 char) NOT NULL,
  "query" varchar(600 char) NOT NULL,
  "catalog" varchar(30 char) NOT NULL,
  "schemata" varchar(30 char) NOT NULL,
  "id" bigint NOT NULL IDENTITY(1,1),
  CLUSTER PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "favorite_uni" on "hetu_favorite" ("user","query","catalog","schemata");


CREATE TABLE IF NOT EXISTS "hetu_query_history" (
  "id" bigint NOT NULL IDENTITY(1,1),
  "user" varchar(20 char) NOT NULL,
  "source" varchar(30 char) NOT NULL,
  "queryId" varchar(50 char) NOT NULL,
  "resource" varchar(30 char) NOT NULL,
  "query" varchar(10000 char) NOT NULL,
  "state" varchar(30 char) NOT NULL,
  "failed" varchar(30 char) DEFAULT NULL,
  "createTime" varchar(30 char) DEFAULT NULL,
  "elapsedTime" varchar(30 char) NOT NULL,
  "cpuTime" varchar(30 char) NOT NULL,
  "executionTime" varchar(30 char) NOT NULL,
  "catalog" varchar(30 char) NOT NULL,
  "schemata" varchar(30 char) NOT NULL,
  "currentMemory" varchar(30 char) NOT NULL,
  "cumulativeUserMemory" double NOT NULL,
  "jsonString" text NOT NULL,
  "completedDrivers" int NOT NULL,
  "runningDrivers" int NOT NULL,
  "queuedDrivers" int NOT NULL,
  "totalCpuTime" varchar(50 char) NOT NULL,
  "totalMemoryReservation" varchar(50 char) NOT NULL,
  "peakTotalMemoryReservation" varchar(50 char) NOT NULL,
  CLUSTER PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "query_history_queryId_uindex" on "hetu_query_history" ("queryId");


-- 迁移 catalog_rule 表数据
INSERT INTO kweaver.catalog_rule (
    "id", "catalog_name", "datasource_type", "pushdown_rule", "is_enabled",
    "create_time", "update_time"
)
SELECT
    "id", "catalog_name", "datasource_type", "pushdown_rule", "is_enabled",
    "create_time", "update_time"
FROM adp.catalog_rule
WHERE NOT EXISTS (
    SELECT 1 FROM kweaver.catalog_rule t
    WHERE t.id = adp.catalog_rule.id
);

-- 迁移 hetu_ctlgs 表数据
INSERT INTO kweaver.hetu_ctlgs (
    "id", "catalog_name", "create_time", "owner", "comment"
)
SELECT
    "id", "catalog_name", "create_time", "owner", "comment"
FROM adp.hetu_ctlgs
WHERE NOT EXISTS (
    SELECT 1 FROM kweaver.hetu_ctlgs t
    WHERE t.id = adp.hetu_ctlgs.id
);

-- 迁移 hetu_dbs 表数据
INSERT INTO kweaver.hetu_dbs (
    "id", "catalog_name", "database_name", "create_time", "owner",
    "comment"
)
SELECT
    "id", "catalog_name", "database_name", "create_time", "owner",
    "comment"
FROM adp.hetu_dbs
WHERE NOT EXISTS (
    SELECT 1 FROM kweaver.hetu_dbs t
    WHERE t.id = adp.hetu_dbs.id
);