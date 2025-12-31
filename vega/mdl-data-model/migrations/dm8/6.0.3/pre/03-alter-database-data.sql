SET SCHEMA dip_mdl;
UPDATE t_data_view SET f_builtin = 1 WHERE f_type='atomic' AND f_group_id != '__index_base';
