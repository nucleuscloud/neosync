ALTER TABLE neosync_api.accounts
DROP CONSTRAINT no_negative_max_allowed_records;

ALTER TABLE neosync_api.accounts
DROP COLUMN max_allowed_records;
