SET SCHEMA dip_mdl;

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
