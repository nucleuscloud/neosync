-- name: GetRunContextByKey :one
SELECT * from neosync_api.runcontexts
WHERE workflow_id = sqlc.arg('workflowId')
  AND external_id = sqlc.arg('externalId')
  AND account_id = sqlc.arg('accountId');

-- name: SetRunContext :exec
INSERT INTO neosync_api.runcontexts (
    workflow_id,
    external_id,
    "value",
    created_by_id,
    updated_by_id,
    account_id
)
VALUES (
    $1,  -- workflow_id
    $2,  -- external_id
    $3,  -- value
    $4,  -- created_by_id
    $5,  -- updated_by_id
    $6   -- account_id
)
ON CONFLICT (workflow_id, external_id, account_id)
DO UPDATE SET
    value = EXCLUDED.value,
    updated_by_id = EXCLUDED.updated_by_id;

-- name: GetRunContextsByExternalIdSuffix :many
SELECT * from neosync_api.runcontexts
WHERE workflow_id = sqlc.arg('workflowId')
  AND external_id LIKE '%' || sqlc.arg('externalIdSuffix')::text
  AND account_id = sqlc.arg('accountId');
