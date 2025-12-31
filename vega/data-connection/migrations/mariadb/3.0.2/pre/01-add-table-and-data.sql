USE vega;

INSERT INTO `data_source` (`id`, `name`, `type_name`, `bin_data`, `comment`, `created_by_uid`, `created_at`, `updated_by_uid`, `updated_at`)
SELECT 'cedb5294-07c3-45b1-a273-17baefa62800','索引库','index_base','{"connect_protocol":"http","host":"mdl-index-base-svc","port":13013}',NULL,'266c6a42-6131-4d62-8f39-853e7093701c','2025-08-15 00:00:00.000','266c6a42-6131-4d62-8f39-853e7093701c','2025-08-15 00:00:00.000'
FROM DUAL WHERE NOT EXISTS(SELECT `id` FROM `data_source` WHERE `id` = 'cedb5294-07c3-45b1-a273-17baefa62800' AND `type_name` = 'index_base');
