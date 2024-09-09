ALTER TABLE neosync_api.accounts
ADD COLUMN max_allowed_records BIGINT NULL;

ALTER TABLE neosync_api.accounts
ADD CONSTRAINT no_negative_max_allowed_records
CHECK (max_allowed_records >= 0);
