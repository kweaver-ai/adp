-- 升级操作
USE dip_mdl;

-- 更新已存在的原子视图的f_builtin属性, 扫描上来的原子视图的f_builtin属性应为1
UPDATE t_data_view SET f_builtin = 1 WHERE f_type="atomic" AND f_group_id != '__index_base';



