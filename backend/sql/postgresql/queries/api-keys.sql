-- name: GetAccountApiKeyById :one
SELECT * from neosync_api.account_api_keys WHERE id = $1;

-- name: GetAccountApiKeys :many
SELECT aak.* from neosync_api.account_api_keys aak
INNER JOIN neosync_api.accounts a on a.id = aak.account_id
WHERE a.id = sqlc.arg('accountId');

-- name: RemoveAccountApiKey :exec
DELETE FROM neosync_api.account_api_keys WHERE id = $1;

-- name: GetAccountApiKeyByKeyValue :one
SELECT * from neosync_api.account_api_keys WHERE key_value = $1;

-- name: CreateAccountApiKey :one
INSERT INTO neosync_api.account_api_keys (
  key_name, key_value, account_id, expires_at, created_by_id, updated_by_id
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;
