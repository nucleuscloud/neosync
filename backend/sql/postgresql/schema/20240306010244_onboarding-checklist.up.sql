ALTER TABLE
  neosync_api.accounts
ADD COLUMN IF NOT EXISTS onboarding_config jsonb NOT NULL DEFAULT '{}'::jsonb;
