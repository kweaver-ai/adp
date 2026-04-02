-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

SET SCHEMA adp;
ALTER TABLE t_discover_task ADD COLUMN IF NOT EXISTS f_scheduled_id varchar(40 char) DEFAULT NULL;
ALTER TABLE t_discover_task ADD COLUMN IF NOT EXISTS f_strategies varchar(100 char) DEFAULT NULL;