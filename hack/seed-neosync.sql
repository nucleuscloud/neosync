INSERT INTO neosync_api.users (id)
VALUES
  ('00000000-0000-0000-0000-000000000000')
ON CONFLICT DO NOTHING;
--
INSERT INTO neosync_api.accounts (id, account_type, account_slug)
VALUES
  ('43c71652-b3f7-4dbb-87c2-508075496054', 0, 'personal')
ON CONFLICT DO NOTHING;
--
INSERT INTO neosync_api.account_user_associations (account_id, user_id)
VALUES
  ('43c71652-b3f7-4dbb-87c2-508075496054', '00000000-0000-0000-0000-000000000000')
ON CONFLICT DO NOTHING;
--
INSERT INTO neosync_api.connections (id, "name", account_id, connection_config, created_by_id, updated_by_id)
VALUES
  ('e53ff29d-f079-4a4d-8db5-359fe480e740', 'test-prod-db', '43c71652-b3f7-4dbb-87c2-508075496054', '{"pgConfig": {"url": "postgres://postgres:postgres@test-prod-db:5432/postgres?sslmode=disable", "connectionOptions": {"maxConnectionLimit": 20}}}'::jsonb, '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000'),
  ('3c592989-c672-46ea-b709-2e338cd1daff', 'test-stage-db', '43c71652-b3f7-4dbb-87c2-508075496054', '{"pgConfig": {"url": "postgres://postgres:postgres@test-stage-db:5432/postgres?sslmode=disable", "connectionOptions": {"maxConnectionLimit": 20}}}'::jsonb, '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000')
ON CONFLICT DO NOTHING;
--

-- Generate Job
INSERT INTO neosync_api.jobs
(id, "name", account_id, status, connection_options, mappings, cron_schedule, created_by_id, updated_by_id, workflow_options, sync_options)
VALUES('6638b7bf-7905-468f-8287-5c2df4732bf0', 'seed-test-prod-db', '43c71652-b3f7-4dbb-87c2-508075496054', 1, '{"generateOptions": {"schemas": [{"schema": "public", "tables": [{"table": "users", "rowCount": 1000}]}], "fkSourceConnectionId": "e53ff29d-f079-4a4d-8db5-359fe480e740"}}'::jsonb, '[{"table": "users", "column": "id", "schema": "public", "jobMappingTransformerModel": {"config": {"generateUuid": {"includeHyphens": true}}, "source": 29}}, {"table": "users", "column": "age", "schema": "public", "jobMappingTransformerModel": {"config": {"GenerateInt64": {"max": 90, "min": 21, "randomizeSign": false}}, "source": 16}}, {"table": "users", "column": "current_salary", "schema": "public", "jobMappingTransformerModel": {"config": {"generateFloat64": {"max": 150000, "min": 50000, "precision": 6, "randomizeSign": false}}, "source": 11}}, {"table": "users", "column": "first_name", "schema": "public", "jobMappingTransformerModel": {"config": {"generateFirstName": {}}, "source": 10}}, {"table": "users", "column": "last_name", "schema": "public", "jobMappingTransformerModel": {"config": {"generateLastName": {}}, "source": 18}}]'::jsonb, '0 0 1 1 *', '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000', '{}'::jsonb, '{}'::jsonb)
ON CONFLICT DO NOTHING;
--
INSERT INTO neosync_api.job_destination_connection_associations
(id, job_id, connection_id, "options")
VALUES('30c56207-2c9e-4cef-a903-1cede68fd0f0', '6638b7bf-7905-468f-8287-5c2df4732bf0', 'e53ff29d-f079-4a4d-8db5-359fe480e740', '{"postgresOptions": {"initTableSchema": false, "onConflictConfig": {"doNothing": true}, "truncateTableconfig": {"truncateCascade": true, "truncateBeforeInsert": true}}}'::jsonb)
ON CONFLICT DO NOTHING;
-- End Generate Job

-- Sync Job
INSERT INTO neosync_api.jobs
(id, "name", account_id, status, connection_options, mappings, cron_schedule, created_by_id, updated_by_id, workflow_options, sync_options)
VALUES('2a5d5caa-7f09-4fdf-a4a7-6a2e341aa600', 'prod-to-stage-sync', '43c71652-b3f7-4dbb-87c2-508075496054', 1, '{"postgresOptions": {"schemas": [], "connectionId": "e53ff29d-f079-4a4d-8db5-359fe480e740", "haltOnNewColumnAddition": false, "subsetByForeignKeyConstraints": true}}'::jsonb, '[{"table": "users", "column": "id", "schema": "public", "jobMappingTransformerModel": {"config": {"passthrough": {}}, "source": 1}}, {"table": "users", "column": "age", "schema": "public", "jobMappingTransformerModel": {"config": {"passthrough": {}}, "source": 1}}, {"table": "users", "column": "current_salary", "schema": "public", "jobMappingTransformerModel": {"config": {"passthrough": {}}, "source": 1}}, {"table": "users", "column": "first_name", "schema": "public", "jobMappingTransformerModel": {"config": {"transformFirstName": {"preserveLength": false}}, "source": 32}}, {"table": "users", "column": "last_name", "schema": "public", "jobMappingTransformerModel": {"config": {"transformLastName": {"preserveLength": false}}, "source": 37}}]'::jsonb, '0 0 1 1 *', '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000', '{}'::jsonb, '{}'::jsonb)
ON CONFLICT DO NOTHING;

INSERT INTO neosync_api.job_destination_connection_associations
(id, job_id, connection_id, "options")
VALUES('15d0f1c0-55c3-4340-9bbb-5cd55c1baabe', '2a5d5caa-7f09-4fdf-a4a7-6a2e341aa600', '3c592989-c672-46ea-b709-2e338cd1daff', '{"postgresOptions": {"initTableSchema": true, "onConflictConfig": {"doNothing": true}, "truncateTableconfig": {"truncateCascade": true, "truncateBeforeInsert": true}}}'::jsonb)
ON CONFLICT DO NOTHING;
-- End Sync Job
