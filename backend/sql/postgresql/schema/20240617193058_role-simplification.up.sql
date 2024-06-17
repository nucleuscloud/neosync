DO $$
BEGIN
    -- Step 1: Reassign ownership of objects from specific roles to the current user
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_owner') THEN
        REASSIGN OWNED BY neosync_api_owner TO current_user;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_readonly') THEN
        REASSIGN OWNED BY neosync_api_readonly TO current_user;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_readwrite') THEN
        REASSIGN OWNED BY neosync_api_readwrite TO current_user;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_schemauser') THEN
        REASSIGN OWNED BY neosync_api_schemauser TO current_user;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_serviceuser') THEN
        REASSIGN OWNED BY neosync_api_serviceuser TO current_user;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_superuser') THEN
        REASSIGN OWNED BY neosync_api_superuser TO current_user;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'nucleus_root') THEN
        REASSIGN OWNED BY nucleus_root TO current_user;
    END IF;
END
$$;

-- Step 2: Revoke all privileges from the roles
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_owner') THEN
        REVOKE ALL PRIVILEGES ON SCHEMA neosync_api FROM neosync_api_owner;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_readonly') THEN
        REVOKE ALL PRIVILEGES ON SCHEMA neosync_api FROM neosync_api_readonly;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_readwrite') THEN
        REVOKE ALL PRIVILEGES ON SCHEMA neosync_api FROM neosync_api_readwrite;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_schemauser') THEN
        REVOKE ALL PRIVILEGES ON SCHEMA neosync_api FROM neosync_api_schemauser;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_serviceuser') THEN
        REVOKE ALL PRIVILEGES ON SCHEMA neosync_api FROM neosync_api_serviceuser;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_superuser') THEN
        REVOKE ALL PRIVILEGES ON SCHEMA neosync_api FROM neosync_api_superuser;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'nucleus_root') THEN
        REVOKE ALL PRIVILEGES ON SCHEMA neosync_api FROM nucleus_root;
        REVOKE ALL PRIVILEGES ON SCHEMA public FROM nucleus_root;
    END IF;
END
$$;

-- Step 3: Drop the hardcoded roles
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_owner') THEN
        DROP ROLE neosync_api_owner;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_readonly') THEN
        DROP ROLE neosync_api_readonly;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_readwrite') THEN
        DROP ROLE neosync_api_readwrite;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_schemauser') THEN
        DROP ROLE neosync_api_schemauser;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_serviceuser') THEN
        DROP ROLE neosync_api_serviceuser;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'neosync_api_superuser') THEN
        DROP ROLE neosync_api_superuser;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'nucleus_root') THEN
        DROP ROLE nucleus_root;
    END IF;
END
$$;

