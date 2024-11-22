-- name: GetJobHooksByJob :many
SELECT * from neosync_api.job_hooks WHERE job_id = $1;

-- name: GetJobHookById :one
SELECT * from neosync_api.job_hooks WHERE id = $1;

-- name: RemoveJobHookById :exec
DELETE FROM neosync_api.job_hooks WHERE id = $1;

-- name: CreateJobHook :one
INSERT INTO neosync_api.job_hooks (
  name, description, job_id, config, created_by_user_id, updated_by_user_id, enabled, priority
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetPreSyncJobHooksToExecute :many
SELECT *
FROM neosync_api.job_hooks
WHERE job_id = $1
  AND enabled = true
  AND hook_timing = 'pre_sync'
ORDER BY priority, created_at, id ASC;

-- name: GetPostSyncJobHooksToExecute :many
SELECT *
FROM neosync_api.job_hooks
WHERE job_id = $1
  AND enabled = true
  AND hook_timing = 'post_sync'
ORDER BY priority, created_at, id ASC;

-- name: IsJobHookNameAvailable :one
SELECT NOT EXISTS(
  SELECT 1
  FROM neosync_api.job_hooks
  WHERE job_id = $1 and name = $2
);
