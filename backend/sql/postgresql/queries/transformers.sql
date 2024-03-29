-- name: GetUserDefinedTransformersByAccount :many
SELECT t.* from neosync_api.transformers t
INNER JOIN neosync_api.accounts a ON a.id = t.account_id
WHERE a.id = sqlc.arg('accountId')
ORDER BY t.name ASC;

-- name: GetUserDefinedTransformerById :one
SELECT * from neosync_api.transformers WHERE id = $1;

-- name: CreateUserDefinedTransformer :one
INSERT INTO neosync_api.transformers (
  name, description, source, account_id, transformer_config, created_by_id, updated_by_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: DeleteUserDefinedTransformerById :exec
DELETE FROM neosync_api.transformers WHERE id = $1;


-- name: UpdateUserDefinedTransformer :one
UPDATE neosync_api.transformers
SET
  name = $1,
  description = $2,
  transformer_config = $3,
  updated_by_id = $4
WHERE id = $5
RETURNING *;


-- name: IsTransformerNameAvailable :one
SELECT count(t.id) from neosync_api.transformers t
INNER JOIN neosync_api.accounts a ON a.id = t.account_id
WHERE a.id = sqlc.arg('accountId') and t.name = sqlc.arg('transformerName');
