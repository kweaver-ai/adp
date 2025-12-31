-- 迁移数据从6.1.0 dip schema 到 6.2.0 adp schema
-- 注意：使用WHERE NOT EXISTS避免重复插入数据
SET SCHEMA adp;

-- 迁移 t_knowledge_network 表数据
INSERT INTO adp.t_knowledge_network (
    f_id, f_name, f_tags, f_comment, f_icon, f_color, f_detail,
    f_branch, f_business_domain,
    f_creator, f_creator_type, f_create_time,
    f_updater, f_updater_type, f_update_time
)
SELECT
    f_id, f_name, f_tags, f_comment, f_icon, f_color, f_detail,
    f_branch, f_business_domain,
    f_creator, f_creator_type, f_create_time,
    f_updater, f_updater_type, f_update_time
FROM dip.t_knowledge_network
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_knowledge_network t
    WHERE t.f_id = dip.t_knowledge_network.f_id
);


-- 迁移 t_object_type 表数据
INSERT INTO adp.t_object_type (
    f_id, f_name, f_tags, f_comment, f_icon, f_color, f_detail,
    f_kn_id, f_branch, f_data_source, f_data_properties,
    f_logic_properties, f_primary_keys, f_display_key,
    f_creator, f_creator_type, f_create_time,
    f_updater, f_updater_type, f_update_time
)
SELECT
    f_id, f_name, f_tags, f_comment, f_icon, f_color, f_detail,
    f_kn_id, f_branch, f_data_source, f_data_properties,
    f_logic_properties, f_primary_keys, f_display_key,
    f_creator, f_creator_type, f_create_time,
    f_updater, f_updater_type, f_update_time
FROM dip.t_object_type
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_object_type t
    WHERE t.f_id = dip.t_object_type.f_id
);


-- 迁移 t_relation_type 表数据
INSERT INTO adp.t_relation_type (
    f_id, f_name, f_tags, f_comment, f_icon, f_color, f_detail,
    f_kn_id, f_branch, f_source_object_type_id,
    f_target_object_type_id, f_type, f_mapping_rules,
    f_creator, f_creator_type, f_create_time,
    f_updater, f_updater_type, f_update_time
)
SELECT
    f_id, f_name, f_tags, f_comment, f_icon, f_color, f_detail,
    f_kn_id, f_branch, f_source_object_type_id,
    f_target_object_type_id, f_type, f_mapping_rules,
    f_creator, f_creator_type, f_create_time,
    f_updater, f_updater_type, f_update_time
FROM dip.t_relation_type
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_relation_type t
    WHERE t.f_id = dip.t_relation_type.f_id
);


-- 迁移 t_action_type 表数据
INSERT INTO adp.t_action_type (
    f_id, f_name, f_tags, f_comment, f_icon, f_color, f_detail,
    f_kn_id, f_branch, f_action_type, f_object_type_id,
    f_condition, f_affect, f_action_source, f_parameters, f_schedule,
    f_creator, f_creator_type, f_create_time,
    f_updater, f_updater_type, f_update_time
)
SELECT
    f_id, f_name, f_tags, f_comment, f_icon, f_color, f_detail,
    f_kn_id, f_branch, f_action_type, f_object_type_id,
    f_condition, f_affect, f_action_source, f_parameters, f_schedule,
    f_creator, f_creator_type, f_create_time,
    f_updater, f_updater_type, f_update_time
FROM dip.t_action_type
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_action_type t
    WHERE t.f_id = dip.t_action_type.f_id
);


-- 迁移 t_kn_job 表数据
INSERT INTO adp.t_kn_job (
    f_id, f_name, f_kn_id, f_branch, f_job_type,
    f_job_concept_config, f_state, f_state_detail,
    f_creator, f_creator_type, f_create_time,
    f_finish_time, f_time_cost
)
SELECT
    f_id, f_name, f_kn_id, f_branch, f_job_type,
    f_job_concept_config, f_state, f_state_detail,
    f_creator, f_creator_type, f_create_time,
    f_finish_time, f_time_cost
FROM dip.t_kn_job
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_kn_job t
    WHERE t.f_id = dip.t_kn_job.f_id
);


-- 迁移 t_kn_task 表数据
INSERT INTO adp.t_kn_task (
    f_id, f_name, f_job_id, f_concept_type, f_concept_id,
    f_index, f_state, f_state_detail,
    f_start_time, f_finish_time, f_time_cost
)
SELECT
    f_id, f_name, f_job_id, f_concept_type, f_concept_id,
    f_index, f_state, f_state_detail,
    f_start_time, f_finish_time, f_time_cost
FROM dip.t_kn_task
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_kn_task t
    WHERE t.f_id = dip.t_kn_task.f_id
);


-- 迁移 t_object_type_status 表数据
INSERT INTO adp.t_object_type_status (
    f_id, f_kn_id, f_branch, f_index, f_index_available
)
SELECT
    f_id, f_kn_id, f_branch, f_index, f_index_available
FROM dip.t_object_type
WHERE NOT EXISTS (
    SELECT 1 FROM adp.t_object_type_status t
    WHERE t.f_id = dip.t_object_type.f_id
);


DROP TABLE IF EXISTS dip.t_knowledge_network CASCADE;
DROP TABLE IF EXISTS dip.t_object_type CASCADE;
DROP TABLE IF EXISTS dip.t_relation_type CASCADE;
DROP TABLE IF EXISTS dip.t_action_type CASCADE;
DROP TABLE IF EXISTS dip.t_kn_job CASCADE;
DROP TABLE IF EXISTS dip.t_kn_task CASCADE;
