-- name: GetTransformers :many
SELECT * from neosync_api.transformers;

-- name: GetTransformersById :one
SELECT * from neosync_api.transformers WHERE id = $1;