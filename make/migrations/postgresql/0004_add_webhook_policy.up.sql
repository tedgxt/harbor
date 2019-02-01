create table webhook_policy (
 id SERIAL NOT NULL,
 name varchar(256),
 project_id int NOT NULL,
 target varchar(512),
 hook_types varchar(512),
 enabled boolean NOT NULL DEFAULT true,
 description text,
 filters varchar(1024),
 deleted boolean DEFAULT false NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 PRIMARY KEY (id)
 );

CREATE TRIGGER webhook_policy_update_time_at_modtime BEFORE UPDATE ON webhook_policy FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

create table webhook_job (
 id SERIAL NOT NULL,
 status varchar(64) NOT NULL,
 policy_id int NOT NULL,
 hook_type varchar(64) NOT NULL,
 job_detail varchar(16384),
 job_uuid varchar(64),
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 PRIMARY KEY (id)
 );

CREATE INDEX webhook_job_policy ON webhook_job (policy_id);
CREATE INDEX webhook_job_poid_uptime ON webhook_job (policy_id, update_time);
CREATE INDEX webhook_job_poid_status ON webhook_job (policy_id, status);

CREATE TRIGGER webhook_job_update_time_at_modtime BEFORE UPDATE ON webhook_job FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();
