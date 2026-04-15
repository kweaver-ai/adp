-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：将 etrino 相关表从 adp 库迁移至 kweaver 库
-- ==========================================
USE kweaver;

RENAME TABLE adp.catalog_rule TO kweaver.catalog_rule;
RENAME TABLE adp.hetu_ctlgs TO kweaver.hetu_ctlgs;
RENAME TABLE adp.hetu_dbs TO kweaver.hetu_dbs;
RENAME TABLE adp.hetu_tbls TO kweaver.hetu_tbls;
RENAME TABLE adp.hetu_tab_cols TO kweaver.hetu_tab_cols;
RENAME TABLE adp.hetu_catalog_params TO kweaver.hetu_catalog_params;
RENAME TABLE adp.hetu_database_params TO kweaver.hetu_database_params;
RENAME TABLE adp.hetu_table_params TO kweaver.hetu_table_params;
RENAME TABLE adp.hetu_column_params TO kweaver.hetu_column_params;
RENAME TABLE adp.hetu_tab_lock TO kweaver.hetu_tab_lock;
RENAME TABLE adp.hetu_favorite TO kweaver.hetu_favorite;
RENAME TABLE adp.hetu_query_history TO kweaver.hetu_query_history;



-- 追加etrino模块数据库初始化
CREATE TABLE IF NOT EXISTS `catalog_rule` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `catalog_name` varchar(200) NOT NULL COMMENT 'catalog名称',
  `datasource_type` varchar(50) NOT NULL COMMENT '数据源类型',
  `pushdown_rule` varchar(500) DEFAULT NULL COMMENT '下推规则',
  `is_enabled` varchar(20) NOT NULL COMMENT '是否启用规则',
  `create_time` varchar(30) DEFAULT NULL COMMENT '创建时间',
  `update_time` varchar(30) DEFAULT NULL COMMENT '修改时间',
  PRIMARY KEY (`id`)
);


CREATE TABLE IF NOT EXISTS `hetu_ctlgs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `catalog_name` varchar(256) NOT NULL,
  `create_time` bigint(20) NOT NULL,
  `owner` varchar(767) DEFAULT NULL,
  `comment` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `catalog_name` (`catalog_name`)
);


CREATE TABLE IF NOT EXISTS `hetu_dbs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `catalog_name` varchar(256) NOT NULL,
  `database_name` varchar(256) NOT NULL,
  `create_time` bigint(20) NOT NULL,
  `owner` varchar(767) DEFAULT NULL,
  `comment` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `catalog_name` (`catalog_name`,`database_name`),
  FOREIGN KEY `hetu_dbs_ibfk_1` (`catalog_name`) REFERENCES `hetu_ctlgs` (`catalog_name`)
);


CREATE TABLE IF NOT EXISTS `hetu_tbls` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `database_id` bigint(20) NOT NULL,
  `table_name` varchar(256) NOT NULL,
  `type` varchar(128) NOT NULL,
  `view_original_text` text DEFAULT NULL,
  `create_time` bigint(20) NOT NULL,
  `owner` varchar(767) DEFAULT NULL,
  `comment` varchar(4000) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `database_id` (`database_id`,`table_name`),
  FOREIGN KEY `hetu_tbls_ibfk_1` (`database_id`) REFERENCES `hetu_dbs` (`id`)
);


CREATE TABLE IF NOT EXISTS `hetu_tab_cols` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `table_id` bigint(20) NOT NULL,
  `column_name` varchar(256) NOT NULL,
  `type` varchar(128) NOT NULL,
  `comment` varchar(4000) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `table_id` (`table_id`,`column_name`),
  FOREIGN KEY `hetu_tab_cols_ibfk_1` (`table_id`) REFERENCES `hetu_tbls` (`id`)
);


CREATE TABLE IF NOT EXISTS `hetu_catalog_params` (
  `catalog_id` bigint(20) NOT NULL,
  `param_key` varchar(256) NOT NULL,
  `param_value` text DEFAULT NULL,
  PRIMARY KEY (`catalog_id`,`param_key`),
  FOREIGN KEY `hetu_catalog_params_ibfk_1` (`catalog_id`) REFERENCES `hetu_ctlgs` (`id`)
);


CREATE TABLE IF NOT EXISTS `hetu_database_params` (
  `database_id` bigint(20) NOT NULL,
  `param_key` varchar(256) NOT NULL,
  `param_value` text DEFAULT NULL,
  PRIMARY KEY (`database_id`,`param_key`),
  FOREIGN KEY `hetu_database_params_ibfk_1` (`database_id`) REFERENCES `hetu_dbs` (`id`)
);


CREATE TABLE IF NOT EXISTS `hetu_table_params` (
  `table_id` bigint(20) NOT NULL,
  `param_key` varchar(256) NOT NULL,
  `param_value` text DEFAULT NULL,
  PRIMARY KEY (`table_id`,`param_key`),
  FOREIGN KEY `hetu_table_params_ibfk_1` (`table_id`) REFERENCES `hetu_tbls` (`id`)
);


CREATE TABLE IF NOT EXISTS `hetu_column_params` (
  `column_id` bigint(20) NOT NULL,
  `param_key` varchar(256) NOT NULL,
  `param_value` text DEFAULT NULL,
  PRIMARY KEY (`column_id`,`param_key`),
  FOREIGN KEY `hetu_column_params_ibfk_1` (`column_id`) REFERENCES `hetu_tab_cols` (`id`)
);


CREATE TABLE IF NOT EXISTS `hetu_tab_lock` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `resource` int(11) NOT NULL,
  `description` varchar(128) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `resource` (`resource`)
);


CREATE TABLE IF NOT EXISTS `hetu_favorite` (
  `creationTime` datetime DEFAULT current_timestamp(),
  `user` varchar(20) NOT NULL,
  `query` varchar(600) NOT NULL,
  `catalog` varchar(30) NOT NULL,
  `schemata` varchar(30) NOT NULL,
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`),
  UNIQUE KEY `favorite_uni` (`user`,`query`,`catalog`,`schemata`)
);


CREATE TABLE IF NOT EXISTS `hetu_query_history` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user` varchar(20) NOT NULL,
  `source` varchar(30) NOT NULL,
  `queryId` varchar(50) NOT NULL,
  `resource` varchar(30) NOT NULL,
  `query` varchar(10000) NOT NULL,
  `state` varchar(30) NOT NULL,
  `failed` varchar(30) DEFAULT NULL,
  `createTime` varchar(30) DEFAULT NULL,
  `elapsedTime` varchar(30) NOT NULL,
  `cpuTime` varchar(30) NOT NULL,
  `executionTime` varchar(30) NOT NULL,
  `catalog` varchar(30) NOT NULL,
  `schemata` varchar(30) NOT NULL,
  `currentMemory` varchar(30) NOT NULL,
  `cumulativeUserMemory` double NOT NULL,
  `jsonString` text NOT NULL,
  `completedDrivers` int(11) NOT NULL,
  `runningDrivers` int(11) NOT NULL,
  `queuedDrivers` int(11) NOT NULL,
  `totalCpuTime` varchar(50) NOT NULL,
  `totalMemoryReservation` varchar(50) NOT NULL,
  `peakTotalMemoryReservation` varchar(50) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `query_history_queryId_uindex` (`queryId`)
);