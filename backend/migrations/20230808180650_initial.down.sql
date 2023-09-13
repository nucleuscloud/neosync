DROP EXTENSION IF EXISTS "pgcrypto";
DROP SCHEMA IF EXISTS neosync_api;

DROP ROLE IF EXISTS neosync_api_readonly;
DROP ROLE IF EXISTS neosync_api_readwrite;
DROP ROLE IF EXISTS neosync_api_schemauser;
DROP ROLE IF EXISTS neosync_api_serviceuser;
DROP ROLE IF EXISTS neosync_api_superuser;
DROP ROLE IF EXISTS neosync_api_owner;

REVOKE CREATE, USAGE ON SCHEMA public FROM nucleus_root;
DROP ROLE IF EXISTS nucleus_root;
