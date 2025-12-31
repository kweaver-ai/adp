SET SCHEMA dip_mdl;
UPDATE t_metric_model set f_data_source = CONCAT('{"type":"data_view","id":"', f_data_view_id, '"}') where f_data_view_id is not null and f_data_view_id != '';
UPDATE t_metric_model set f_unit_type = 'percentageUnit' where f_unit = '%';
