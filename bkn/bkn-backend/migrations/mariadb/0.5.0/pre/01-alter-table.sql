-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

USE adp;

-- 移除关系类名称唯一性约束，同一 BKN 内允许同名关系类存在
DROP INDEX IF EXISTS uk_relation_type_name ON adp.t_relation_type;
