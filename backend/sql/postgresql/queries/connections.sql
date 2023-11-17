-- name: GetConnectionById :one
SELECT * from neosync_api.connections WHERE id = $1;

-- name: GetConnectionByNameAndAccount :one
SELECT c.* from neosync_api.connections c
INNER JOIN neosync_api.accounts a ON a.id = c.account_id
WHERE a.id = sqlc.arg('accountId') AND c.name = sqlc.arg('connectionName');

-- name: GetConnectionsByAccount :many
SELECT c.* from neosync_api.connections c
INNER JOIN neosync_api.accounts a ON a.id = c.account_id
WHERE a.id = sqlc.arg('accountId')
ORDER BY c.created_at DESC;

-- name: RemoveConnectionById :exec
DELETE FROM neosync_api.connections WHERE id = $1;

-- name: RemoveConnectionByNameAndAccount :exec
DELETE FROM neosync_api.connections WHERE name = $1 and account_id = $2;

-- name: CreateConnection :one
INSERT INTO neosync_api.connections (
  name, account_id, connection_config, created_by_id, updated_by_id
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateConnection :one
UPDATE neosync_api.connections
SET name = $1, connection_config = $2,
updated_by_id = $3
WHERE id = $4
RETURNING *;

-- name: IsConnectionNameAvailable :one
SELECT count(c.id) from neosync_api.connections c
INNER JOIN neosync_api.accounts a ON a.id = c.account_id
WHERE a.id = sqlc.arg('accountId') and c.name = sqlc.arg('connectionName');

-- name: IsConnectionInAccount :one
SELECT count(c.id) from neosync_api.connections c
INNER JOIN neosync_api.accounts a ON a.id = c.account_id
WHERE a.id = sqlc.arg('accountId') and c.id = sqlc.arg('connectionId');
