-- 移除关系类名称唯一性约束，同一 BKN 内允许同名关系类存在
ALTER TABLE t_relation_type DROP INDEX IF EXISTS uk_relation_type_name;
