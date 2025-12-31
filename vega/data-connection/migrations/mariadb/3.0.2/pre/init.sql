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

INSERT INTO `data_source` (`id`, `name`, `type_name`, `bin_data`, `comment`, `created_by_uid`, `created_at`, `updated_by_uid`, `updated_at`)
SELECT 'cedb5294-07c3-45b1-a273-17baefa62800','索引库','index_base','{"connect_protocol":"http","host":"mdl-index-base-svc","port":13013}',NULL,'266c6a42-6131-4d62-8f39-853e7093701c','2025-08-15 00:00:00.000','266c6a42-6131-4d62-8f39-853e7093701c','2025-08-15 00:00:00.000'
FROM DUAL WHERE NOT EXISTS(SELECT `id` FROM `data_source` WHERE `id` = 'cedb5294-07c3-45b1-a273-17baefa62800' AND `type_name` = 'index_base');
