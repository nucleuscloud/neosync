-- name: GetTransformersByAccount :many
SELECT t.* from neosync_api.transformers t
INNER JOIN neosync_api.accounts a ON a.id = t.account_id
WHERE a.id = sqlc.arg('accountId')
ORDER BY t.created_at DESC;

-- name: GetTransformersById :one
SELECT * from neosync_api.transformers WHERE id = $1;
