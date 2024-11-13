CREATE TABLE IF NOT EXISTS neosync_api.casbin_rule(
    p_type VARCHAR(32) DEFAULT '' NOT NULL,
    v0     VARCHAR(255) DEFAULT '' NOT NULL,
    v1     VARCHAR(255) DEFAULT '' NOT NULL,
    v2     VARCHAR(255) DEFAULT '' NOT NULL,
    v3     VARCHAR(255) DEFAULT '' NOT NULL,
    v4     VARCHAR(255) DEFAULT '' NOT NULL,
    v5     VARCHAR(255) DEFAULT '' NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_neosync_api_casbin_rule" ON neosync_api.casbin_rule (p_type, v0, v1);

CREATE TRIGGER update_neosync_api_casbin_rule_updated_at
BEFORE UPDATE ON neosync_api.casbin_rule
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
