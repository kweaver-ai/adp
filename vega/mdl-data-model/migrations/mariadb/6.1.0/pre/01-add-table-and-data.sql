USE dip_mdl;

-- 视图行列规则表
CREATE TABLE IF NOT EXISTS t_data_view_row_column_rule (
  f_rule_id varchar(40) NOT NULL DEFAULT '' COMMENT '视图行列规则 id',
  f_rule_name varchar(255) NOT NULL COMMENT '视图行列规则名称',
  f_view_id varchar(40) NOT NULL COMMENT '视图 id',
  f_tags varchar(255) NOT NULL DEFAULT '' COMMENT '标签',
  f_comment varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  f_fields longtext NOT NULL COMMENT '列',
  f_row_filters text NOT NULL COMMENT '行过滤规则',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间', 
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_updater varchar(40) NOT NULL DEFAULT '' COMMENT '更新者id',
  f_updater_type varchar(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
  PRIMARY KEY (f_rule_id),
  UNIQUE KEY uk_f_rule_name (f_rule_name, f_view_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据视图行列规则';