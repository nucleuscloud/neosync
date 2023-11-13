ALTER TABLE neosync_api.users
ADD COLUMN user_type smallint not null default 0;

ALTER TABLE neosync_api.account_api_keys
ADD COLUMN user_id uuid null;

ALTER TABLE neosync_api.account_api_keys
ADD CONSTRAINT fk_account_api_keys_user_id_users_id FOREIGN KEY (user_id) REFERENCES neosync_api.users(id);
