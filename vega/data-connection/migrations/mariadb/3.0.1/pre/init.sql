USE vega;

CREATE TABLE IF NOT EXISTS `data_source` (
    `id` char(36) NOT NULL COMMENT '主键，生成规则:36位uuid',
    `name` varchar(128) NOT NULL COMMENT '数据源展示名称',
    `type_name` varchar(30) NOT NULL COMMENT '数据库类型',
    `bin_data` blob NOT NULL COMMENT '数据源配置信息',
    `comment` varchar(255) DEFAULT NULL COMMENT '描述',
    `created_by_uid` char(36) NOT NULL COMMENT '创建人',
    `created_at` datetime(3) NOT NULL COMMENT '创建时间',
    `updated_by_uid` char(36) DEFAULT NULL COMMENT '修改人',
    `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`)
);
