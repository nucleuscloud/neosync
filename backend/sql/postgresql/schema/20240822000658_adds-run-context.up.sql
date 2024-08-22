CREATE TABLE neosync_api.runcontexts (
    workflow_id TEXT NOT NULL,
    external_id TEXT NOT NULL,
    account_id UUID NOT NULL,
    value bytea NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by_id UUID NOT NULL,
    updated_by_id UUID NOT NULL,
    PRIMARY KEY (workflow_id, external_id, account_id),
    FOREIGN KEY (account_id) REFERENCES neosync_api.accounts(id)
);

CREATE TRIGGER update_neosync_api_runcontexts_updated_at
BEFORE UPDATE ON neosync_api.runcontexts
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
