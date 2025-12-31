USE adp;

-- 对象类状态
CREATE TABLE IF NOT EXISTS t_object_type_status (
  f_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '对象类id',
  f_kn_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务知识网络id',
  f_branch VARCHAR(40) NOT NULL DEFAULT '' COMMENT '分支',
  f_incremental_key VARCHAR(40) NOT NULL DEFAULT '' COMMENT '对象类增量键',
  f_incremental_value VARCHAR(40) NOT NULL DEFAULT '' COMMENT '对象类当前增量值',
  f_index VARCHAR(255) NOT NULL DEFAULT '' COMMENT '索引名称',
  f_index_available BOOLEAN NOT NULL DEFAULT 0 COMMENT '索引是否可用',
  f_doc_count BIGINT(20) NOT NULL DEFAULT 0 COMMENT '文档数量',
  f_storage_size BIGINT(20) NOT NULL DEFAULT 0 COMMENT '存储大小',
  f_update_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (f_kn_id,f_branch,f_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '对象类状态';

INSERT INTO t_object_type_status(f_id, f_kn_id, f_branch, f_index, f_index_available)
SELECT t1.f_id, t1.f_kn_id, t1.f_branch, t1.f_index, t1.f_index_available from t_object_type as t1
LEFT JOIN t_object_type_status as t2 ON t1.f_id = t2.f_id
WHERE t2.f_id IS NULL;

-- 概念分组
CREATE TABLE IF NOT EXISTS t_concept_group (
  f_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '概念分组id',
  f_name VARCHAR(40) NOT NULL DEFAULT '' COMMENT '概念分组名称',
  f_tags VARCHAR(255) DEFAULT NULL COMMENT '标签',
  f_comment VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  f_icon VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图标',
  f_color VARCHAR(40) NOT NULL DEFAULT '' COMMENT '颜色',
  f_detail MEDIUMTEXT DEFAULT NULL COMMENT '概览',
  f_kn_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务知识网络id',
  f_branch VARCHAR(40) NOT NULL DEFAULT '' COMMENT '分支',
  f_creator VARCHAR(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_updater VARCHAR(40) NOT NULL DEFAULT '' COMMENT '更新者id',
  f_updater_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
  f_update_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (f_kn_id,f_branch,f_id),
  UNIQUE KEY uk_concept_group_name (f_kn_id,f_branch,f_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '概念分组';
 
-- 分组与概念对应表
CREATE TABLE IF NOT EXISTS t_concept_group_relation (
  f_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '主键id',
  f_kn_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务知识网络id',
  f_branch VARCHAR(40) NOT NULL DEFAULT '' COMMENT '分支',
  f_group_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '概念分组id',
  f_concept_type VARCHAR(40) NOT NULL DEFAULT '' COMMENT '概念类型',
  f_concept_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '概念id',
  f_create_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (f_id),
  UNIQUE KEY uk_concept_group_relation (f_kn_id,f_branch,f_group_id,f_concept_type,f_concept_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '分组与概念对应表';
