-- 升级操作
USE dip_mdl;

-- 将已存在的索引视图添加 type, query type, data_source_id, data_source_type
update t_data_view set f_technical_name = f_view_name, f_type = 'atomic', f_query_type = 'DSL', f_data_source_id = 'cedb5294-07c3-45b1-a273-17baefa62800', f_data_source_type = 'index_base' where f_group_id = '__index_base';



