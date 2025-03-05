-- name: GetAccountHooksByAccount :many
SELECT * from neosync_api.account_hooks WHERE account_id = $1;

-- name: GetAccountHookById :one
SELECT * from neosync_api.account_hooks WHERE id = $1;

-- name: RemoveAccountHookById :exec
DELETE FROM neosync_api.account_hooks WHERE id = $1;

-- name: CreateAccountHook :one
INSERT INTO neosync_api.account_hooks (
  name, description, account_id, events, config, created_by_user_id, updated_by_user_id, enabled
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: IsAccountHookNameAvailable :one
SELECT NOT EXISTS(
  SELECT 1
  FROM neosync_api.account_hooks
  WHERE account_id = $1 AND name = $2
);

-- name: SetAccountHookEnabled :one
UPDATE neosync_api.account_hooks
SET enabled = $1,
    updated_by_user_id = $2
WHERE id = $3
RETURNING *;

-- name: UpdateAccountHook :one
UPDATE neosync_api.account_hooks
SET name = $1,
    description = $2,
    events = $3,
    config = $4,
    enabled = $5,
    updated_by_user_id = $6
WHERE id = $7
RETURNING *;

-- name: GetActiveAccountHooksByEvent :many
SELECT * from neosync_api.account_hooks
WHERE account_id = $1
  AND enabled = true
  AND events && sqlc.arg(events)::int[]
ORDER BY created_at ASC;

-- name: GetSlackAccessToken :one
SELECT (oauth_v2_response->>'access_token')::TEXT as access_token
FROM neosync_api.slack_oauth_connections
WHERE account_id = $1;

-- name: CreateSlackOAuthConnection :one
INSERT INTO neosync_api.slack_oauth_connections (account_id, oauth_v2_response, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (account_id) DO UPDATE
SET oauth_v2_response = EXCLUDED.oauth_v2_response,
    updated_at = CURRENT_TIMESTAMP,
    updated_by_user_id = $4
RETURNING *;

-- name: DeleteSlackOAuthConnection :exec
DELETE FROM neosync_api.slack_oauth_connections
WHERE account_id = $1;
