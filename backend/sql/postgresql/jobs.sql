-- name: GetJobById :one
SELECT * from neosync_api.jobs WHERE id = $1;

-- name: GetJobByNameAndAccount :one
SELECT j.* from neosync_api.jobs j
INNER JOIN neosync_api.accounts a ON a.id = j.account_id
WHERE a.id = sqlc.arg('accountId') AND j.name = sqlc.arg('jobName');

-- name: GetJobsByAccount :many
SELECT j.* from neosync_api.jobs j
INNER JOIN neosync_api.accounts a ON a.id = j.account_id
WHERE a.id = sqlc.arg('accountId')
ORDER BY j.created_at DESC;

-- name: RemoveJobById :exec
DELETE FROM neosync_api.jobs WHERE id = $1;

-- name: CreateJob :one
INSERT INTO neosync_api.jobs (
  name, account_id, status, connection_source_id, mappings,
  cron_schedule, halt_on_new_column_addition, created_by_id, updated_by_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: IsJobNameAvailable :one
SELECT count(j.id) from neosync_api.jobs j
INNER JOIN neosync_api.accounts a ON a.id = j.account_id
WHERE a.id = sqlc.arg('accountId') AND j.name = sqlc.arg('jobName');

-- name: CreateJobConnectionDestination :one
INSERT INTO neosync_api.job_destination_connection_associations (
  job_id, connection_id
) VALUES (
  $1, $2
)
ON CONFLICT(job_id, connection_id)
DO NOTHING
RETURNING *;

-- name: CreateJobConnectionDestinations :copyfrom
INSERT INTO neosync_api.job_destination_connection_associations (
  job_id, connection_id
) VALUES (
  $1, $2
);

-- name: GetJobConnectionDestinations :many
SELECT jdca.connection_id from neosync_api.job_destination_connection_associations jdca
INNER JOIN neosync_api.jobs j ON j.id = jdca.job_id
WHERE j.id = $1;

-- name: GetJobConnectionDestinationsByJobIds :many
SELECT jdca.* from neosync_api.job_destination_connection_associations jdca
INNER JOIN neosync_api.jobs j ON j.id = jdca.job_id
WHERE j.id = ANY(sqlc.arg('jobIds')::uuid[]);

-- name: RemoveJobConnectionDestinations :exec
DELETE FROM neosync_api.job_destination_connection_associations
WHERE id = ANY(sqlc.arg('jobIds')::uuid[]);
