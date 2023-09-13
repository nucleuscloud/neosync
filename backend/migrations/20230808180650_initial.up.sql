-- Land mine: This migration will fail in produciton environments because all of these roles and schemas already exist.
-- The workflow here is: let the db-migration container run. It will create the migrations table and be marked in a dirty state.
-- Make it not dirty, then restart the pod (if it doesnt do it automatically) - and all of the remaining migrations will run
-- At some point we should make this initial script idempotent.
DO $$
BEGIN
	CREATE ROLE neosync_api_owner WITH
		NOSUPERUSER
		NOCREATEDB
		NOCREATEROLE
		INHERIT
		NOLOGIN
		NOREPLICATION
		NOBYPASSRLS
		CONNECTION LIMIT -1
		VALID UNTIL 'infinity';
EXCEPTION WHEN duplicate_object THEN RAISE NOTICE '%, skipping', SQLERRM USING ERRCODE = SQLSTATE;
END
$$;



DO $$
BEGIN
	CREATE ROLE neosync_api_readonly WITH
		NOSUPERUSER
		NOCREATEDB
		NOCREATEROLE
		INHERIT
		NOLOGIN
		NOREPLICATION
		NOBYPASSRLS
		CONNECTION LIMIT -1
		VALID UNTIL 'infinity';
EXCEPTION WHEN duplicate_object THEN RAISE NOTICE '%, skipping', SQLERRM USING ERRCODE = SQLSTATE;
END
$$;

DO $$
BEGIN
	CREATE ROLE neosync_api_readwrite WITH
		NOSUPERUSER
		NOCREATEDB
		NOCREATEROLE
		INHERIT
		NOLOGIN
		NOREPLICATION
		NOBYPASSRLS
		CONNECTION LIMIT -1
		VALID UNTIL 'infinity';
EXCEPTION WHEN duplicate_object THEN RAISE NOTICE '%, skipping', SQLERRM USING ERRCODE = SQLSTATE;
END
$$;

DO $$
BEGIN
	CREATE ROLE neosync_api_schemauser WITH
		NOSUPERUSER
		NOCREATEDB
		NOCREATEROLE
		INHERIT
		LOGIN
		NOREPLICATION
		NOBYPASSRLS
		CONNECTION LIMIT -1
		VALID UNTIL 'infinity';
EXCEPTION WHEN duplicate_object THEN RAISE NOTICE '%, skipping', SQLERRM USING ERRCODE = SQLSTATE;
END
$$;

DO $$
BEGIN
	CREATE ROLE neosync_api_serviceuser WITH
		NOSUPERUSER
		NOCREATEDB
		NOCREATEROLE
		INHERIT
		LOGIN
		NOREPLICATION
		NOBYPASSRLS
		CONNECTION LIMIT -1
		VALID UNTIL 'infinity';
EXCEPTION WHEN duplicate_object THEN RAISE NOTICE '%, skipping', SQLERRM USING ERRCODE = SQLSTATE;
END
$$;

DO $$
BEGIN
	CREATE ROLE neosync_api_superuser WITH
		NOSUPERUSER
		NOCREATEDB
		NOCREATEROLE
		INHERIT
		LOGIN
		NOREPLICATION
		NOBYPASSRLS
		CONNECTION LIMIT -1
		VALID UNTIL 'infinity';
EXCEPTION WHEN duplicate_object THEN RAISE NOTICE '%, skipping', SQLERRM USING ERRCODE = SQLSTATE;
END
$$;

DO $$
BEGIN
	CREATE ROLE nucleus_root WITH
		NOSUPERUSER
		CREATEDB
		CREATEROLE
		INHERIT
		LOGIN
		NOREPLICATION
		NOBYPASSRLS
		CONNECTION LIMIT -1
		VALID UNTIL 'infinity';
EXCEPTION WHEN duplicate_object THEN RAISE NOTICE '%, skipping', SQLERRM USING ERRCODE = SQLSTATE;
END
$$;

GRANT CREATE, USAGE ON SCHEMA public TO nucleus_root;

-- schema will always exist since we have to bootstrap it so golang-migrate has
-- a place to store its data..
CREATE SCHEMA IF NOT EXISTS neosync_api AUTHORIZATION neosync_api_owner;

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

GRANT USAGE ON SCHEMA neosync_api TO neosync_api_readonly;
GRANT ALL ON SCHEMA neosync_api TO neosync_api_owner;
GRANT USAGE ON SCHEMA neosync_api TO neosync_api_readwrite;
