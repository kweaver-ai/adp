-- 升级操作
USE dip_mdl;

-- 指标模型中的数据视图 data_view_id，升级为data_source字段，值形如： {"type": "data_view", "id": "$data_view_id"}
update t_metric_model set f_data_source = CONCAT('{"type":"data_view","id":"', f_data_view_id, '"}') where f_data_view_id is not null and f_data_view_id != '';
update t_metric_model set f_unit_type = 'percentageUnit' where f_unit = '%';
