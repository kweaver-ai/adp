SET SCHEMA adp;
-- 将dip_data_operator_hub库下t_resource_deploy表中的数据迁移到adp库下t_resource_deploy表
INSERT INTO adp.t_resource_deploy (
    f_resource_id,
    f_type,
    f_version,
    f_name,
    f_description,
    f_config,
    f_status,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time
)
SELECT
    f_resource_id,
    f_type,
    f_version,
    f_name,
    f_description,
    f_config,
    f_status,
    f_create_user,
    f_create_time,
    f_update_user,
    f_update_time
FROM dip_data_operator_hub.t_resource_deploy src
WHERE NOT EXISTS (
    SELECT 1
    FROM adp.t_resource_deploy dest
    WHERE dest.f_resource_id = src.f_resource_id
      AND dest.f_type = src.f_type
      AND dest.f_version = src.f_version
);