/*remove the constraint for name in table 'notification_policy'*/
ALTER TABLE notification_policy DROP CONSTRAINT notification_policy_name_key;
/*add union unique constraint for name and project_id in table 'notification_policy'*/
ALTER TABLE notification_policy ADD UNIQUE(name,project_id);