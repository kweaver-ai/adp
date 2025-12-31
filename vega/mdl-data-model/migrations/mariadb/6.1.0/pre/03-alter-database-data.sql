-- 升级操作
USE dip_mdl;

-- 更新索引库类型的视图的查询类型为 IndexBase
UPDATE t_data_view SET f_query_type = 'IndexBase' WHERE f_query_type = 'DSL';
UPDATE t_data_view SET f_query_type = 'DSL' WHERE f_data_source_type='opensearch';

-- 更新创建人类型为 user
UPDATE t_data_view SET f_creator_type = 'user' WHERE f_creator_type = '';
UPDATE t_metric_model SET f_creator_type = 'user' WHERE f_creator_type = '';
UPDATE t_metric_model_task SET f_creator_type = 'user' WHERE f_creator_type = '';
UPDATE t_event_models SET f_creator_type = 'user' WHERE f_creator_type = '';
UPDATE t_event_model_task SET f_creator_type = 'user' WHERE f_creator_type = '';
UPDATE t_data_dict SET f_creator_type = 'user' WHERE f_creator_type = '';
UPDATE t_trace_model SET f_creator_type = 'user' WHERE f_creator_type = '';
UPDATE t_data_model_job SET f_creator_type = 'user' WHERE f_creator_type = '';
UPDATE t_objective_model SET f_creator_type = 'user' WHERE f_creator_type = '';
