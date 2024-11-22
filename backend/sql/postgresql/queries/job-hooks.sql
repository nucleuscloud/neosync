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
  AND hook_timing = 'preSync'
ORDER BY priority, created_at, id ASC;

-- name: GetPostSyncJobHooksToExecute :many
SELECT *
FROM neosync_api.job_hooks
WHERE job_id = $1
  AND enabled = true
  AND hook_timing = 'postSync'
ORDER BY priority, created_at, id ASC;

-- name: IsJobHookNameAvailable :one
SELECT NOT EXISTS(
  SELECT 1
  FROM neosync_api.job_hooks
  WHERE job_id = $1 and name = $2
);

-- name: SetJobHookEnabled :one
UPDATE neosync_api.job_hooks
SET enabled = $1,
    updated_by_user_id = $2
WHERE id = $2
RETURNING *;


-- name: UpdateJobHook :one
UPDATE neosync_api.job_hooks
SET name = $1,
    description = $2,
    config = $3,
    enabled = $4,
    priority = $5,
    updated_by_user_id = $6
WHERE id = $7
RETURNING *;
