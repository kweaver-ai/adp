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
