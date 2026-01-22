SET SEARCH_PATH TO adp;
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
FROM dip_data_operator_hub.t_resource_deploy
WHERE NOT EXISTS (
    SELECT 1
    FROM adp.t_resource_deploy
    WHERE adp.t_resource_deploy.f_resource_id = dip_data_operator_hub.t_resource_deploy.f_resource_id
      AND adp.t_resource_deploy.f_type = dip_data_operator_hub.t_resource_deploy.f_type
      AND adp.t_resource_deploy.f_version = dip_data_operator_hub.t_resource_deploy.f_version
);