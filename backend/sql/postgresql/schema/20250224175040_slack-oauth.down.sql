DROP TRIGGER IF EXISTS update_neosync_api_slack_oauth_connections_updated_at ON neosync_api.slack_oauth_connections;

DROP TABLE IF EXISTS neosync_api.slack_oauth_connections;

ALTER TABLE neosync_api.account_hooks
  ALTER COLUMN hook_type SET GENERATED ALWAYS AS (
    CASE
      WHEN config->'webhook' IS NOT NULL THEN 'webhook'
      ELSE NULL
    END
  ) STORED;
