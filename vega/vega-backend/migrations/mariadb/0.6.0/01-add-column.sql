-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

USE adp;
ALTER TABLE adp.t_discover_task ADD COLUMN  IF NOT EXISTS f_scheduled_id varchar(40) DEFAULT NULL AFTER f_catalog_id;
ALTER TABLE adp.t_discover_task ADD COLUMN  IF NOT EXISTS f_strategies varchar(100) DEFAULT NULL AFTER f_scheduled_id;