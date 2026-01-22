USE adp;

CREATE TABLE IF NOT EXISTS `t_task_scan_schedule` (
  `id` char(36) NOT NULL COMMENT '唯一id，雪花算法',
  `type` tinyint NOT NULL DEFAULT 0 COMMENT '扫描任务：0 :即时-数据源;1 :即时-数据表;2: 定时-数据源',
  `name` varchar(128) NOT NULL COMMENT '任务名称',
  `cron_expression` varchar(64) NOT NULL COMMENT 'cron表达式',
  `scan_strategy` varchar(64) DEFAULT NULL COMMENT '快速扫描策略',
  `task_status` tinyint DEFAULT 0 COMMENT '定时扫描任务:0 close 1 open',
  `ds_id` char(36) DEFAULT NULL COMMENT '数据源唯一标识',
  `create_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  `create_user` varchar(64) DEFAULT NULL COMMENT '创建用户',
  `operation_time` datetime DEFAULT NULL,
  `operation_user` varchar(64) DEFAULT NULL,
  `operation_type` tinyint NOT NULL DEFAULT 0 COMMENT '状态：0新增1删除2更新',
  PRIMARY KEY (`id`),
  KEY `t_task_scan_ds_id_IDX` (`ds_id`)
);
