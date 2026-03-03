use adp;

CREATE TABLE IF NOT EXISTS `t_dag_step_index` (
  `f_id` bigint(20) NOT NULL,
  `f_dag_id` bigint(20) NOT NULL,
  `f_operator` varchar(255) DEFAULT NULL,
  `f_source_id` varchar(255) DEFAULT NULL,
  `f_has_datasource` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`f_id`),
  KEY `idx_step_op_src_dag` (`f_operator`, `f_source_id`, `f_dag_id`) USING BTREE,
  KEY `idx_step_op_dag` (`f_operator`, `f_dag_id`) USING BTREE,
  KEY `idx_step_has_ds_dag` (`f_has_datasource`, `f_dag_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE IF NOT EXISTS `t_dag_trigger_config_index` (
  `f_id` bigint(20) NOT NULL,
  `f_dag_id` bigint(20) NOT NULL,
  `f_operator` varchar(255) DEFAULT NULL,
  `f_source_id` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`f_id`),
  KEY `idx_tc_op_src_dag` (`f_operator`, `f_source_id`, `f_dag_id`) USING BTREE,
  KEY `idx_tc_op_dag` (`f_operator`, `f_dag_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE IF NOT EXISTS `t_dag_accessor_index` (
  `f_id` bigint(20) NOT NULL,
  `f_dag_id` bigint(20) NOT NULL,
  `f_accessor_id` varchar(255) NOT NULL,
  PRIMARY KEY (`f_id`),
  KEY `idx_accessor_id_dag` (`f_accessor_id`, `f_dag_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
