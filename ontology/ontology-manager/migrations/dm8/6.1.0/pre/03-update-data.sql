SET SCHEMA dip;

UPDATE t_knowledge_network SET f_creator_type = 'user' WHERE f_creator_type = '';
UPDATE t_knowledge_network SET f_updater_type = 'user' WHERE f_updater_type = '';
UPDATE t_knowledge_network SET f_business_domain = 'bd_public' WHERE f_business_domain = '';
UPDATE t_object_type SET f_creator_type = 'user' WHERE f_creator_type = '';
UPDATE t_object_type SET f_updater_type = 'user' WHERE f_updater_type = '';
UPDATE t_relation_type SET f_creator_type = 'user' WHERE f_creator_type = '';
UPDATE t_relation_type SET f_updater_type = 'user' WHERE f_updater_type = '';
UPDATE t_action_type SET f_creator_type = 'user' WHERE f_creator_type = '';
UPDATE t_action_type SET f_updater_type = 'user' WHERE f_updater_type = '';

INSERT INTO model_management.t_bd_resource_r (f_bd_id, f_resource_id, f_resource_type, f_create_by, created_at, updated_at)
SELECT 'bd_public', f_id, 'knowledge_network', f_creator, current_timestamp(), current_timestamp()
FROM dip.t_knowledge_network AS a WHERE NOT EXISTS (
  SELECT * FROM model_management.t_bd_resource_r AS b
  WHERE a.f_id = b.f_resource_id AND b.f_resource_type = 'knowledge_network'
);
