// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: jobs.sql

package db_queries

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
)

const createJob = `-- name: CreateJob :one
INSERT INTO neosync_api.jobs (
  name, account_id, status, connection_options, mappings,
  cron_schedule, created_by_id, updated_by_id, workflow_options, sync_options,
  virtual_foreign_keys
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING id, created_at, updated_at, name, account_id, status, connection_options, mappings, cron_schedule, created_by_id, updated_by_id, workflow_options, sync_options, virtual_foreign_keys
`

type CreateJobParams struct {
	Name               string
	AccountID          pgtype.UUID
	Status             int16
	ConnectionOptions  *pg_models.JobSourceOptions
	Mappings           []*pg_models.JobMapping
	CronSchedule       pgtype.Text
	CreatedByID        pgtype.UUID
	UpdatedByID        pgtype.UUID
	WorkflowOptions    *pg_models.WorkflowOptions
	SyncOptions        *pg_models.ActivityOptions
	VirtualForeignKeys []*pg_models.VirtualForeignConstraint
}

func (q *Queries) CreateJob(ctx context.Context, db DBTX, arg CreateJobParams) (NeosyncApiJob, error) {
	row := db.QueryRow(ctx, createJob,
		arg.Name,
		arg.AccountID,
		arg.Status,
		arg.ConnectionOptions,
		arg.Mappings,
		arg.CronSchedule,
		arg.CreatedByID,
		arg.UpdatedByID,
		arg.WorkflowOptions,
		arg.SyncOptions,
		arg.VirtualForeignKeys,
	)
	var i NeosyncApiJob
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.AccountID,
		&i.Status,
		&i.ConnectionOptions,
		&i.Mappings,
		&i.CronSchedule,
		&i.CreatedByID,
		&i.UpdatedByID,
		&i.WorkflowOptions,
		&i.SyncOptions,
		&i.VirtualForeignKeys,
	)
	return i, err
}

const createJobConnectionDestination = `-- name: CreateJobConnectionDestination :one
INSERT INTO neosync_api.job_destination_connection_associations (
  job_id, connection_id, options
) VALUES (
  $1, $2, $3
)
ON CONFLICT(job_id, connection_id)
DO NOTHING
RETURNING id, created_at, updated_at, job_id, connection_id, options
`

type CreateJobConnectionDestinationParams struct {
	JobID        pgtype.UUID
	ConnectionID pgtype.UUID
	Options      *pg_models.JobDestinationOptions
}

func (q *Queries) CreateJobConnectionDestination(ctx context.Context, db DBTX, arg CreateJobConnectionDestinationParams) (NeosyncApiJobDestinationConnectionAssociation, error) {
	row := db.QueryRow(ctx, createJobConnectionDestination, arg.JobID, arg.ConnectionID, arg.Options)
	var i NeosyncApiJobDestinationConnectionAssociation
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.JobID,
		&i.ConnectionID,
		&i.Options,
	)
	return i, err
}

type CreateJobConnectionDestinationsParams struct {
	JobID        pgtype.UUID
	ConnectionID pgtype.UUID
	Options      *pg_models.JobDestinationOptions
}

const deleteJob = `-- name: DeleteJob :exec
DELETE FROM neosync_api.jobs WHERE id = $1
`

func (q *Queries) DeleteJob(ctx context.Context, db DBTX, id pgtype.UUID) error {
	_, err := db.Exec(ctx, deleteJob, id)
	return err
}

const getJobById = `-- name: GetJobById :one
SELECT id, created_at, updated_at, name, account_id, status, connection_options, mappings, cron_schedule, created_by_id, updated_by_id, workflow_options, sync_options, virtual_foreign_keys from neosync_api.jobs WHERE id = $1
`

func (q *Queries) GetJobById(ctx context.Context, db DBTX, id pgtype.UUID) (NeosyncApiJob, error) {
	row := db.QueryRow(ctx, getJobById, id)
	var i NeosyncApiJob
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.AccountID,
		&i.Status,
		&i.ConnectionOptions,
		&i.Mappings,
		&i.CronSchedule,
		&i.CreatedByID,
		&i.UpdatedByID,
		&i.WorkflowOptions,
		&i.SyncOptions,
		&i.VirtualForeignKeys,
	)
	return i, err
}

const getJobByNameAndAccount = `-- name: GetJobByNameAndAccount :one
SELECT j.id, j.created_at, j.updated_at, j.name, j.account_id, j.status, j.connection_options, j.mappings, j.cron_schedule, j.created_by_id, j.updated_by_id, j.workflow_options, j.sync_options, j.virtual_foreign_keys from neosync_api.jobs j
INNER JOIN neosync_api.accounts a ON a.id = j.account_id
WHERE a.id = $1 AND j.name = $2
`

type GetJobByNameAndAccountParams struct {
	AccountId pgtype.UUID
	JobName   string
}

func (q *Queries) GetJobByNameAndAccount(ctx context.Context, db DBTX, arg GetJobByNameAndAccountParams) (NeosyncApiJob, error) {
	row := db.QueryRow(ctx, getJobByNameAndAccount, arg.AccountId, arg.JobName)
	var i NeosyncApiJob
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.AccountID,
		&i.Status,
		&i.ConnectionOptions,
		&i.Mappings,
		&i.CronSchedule,
		&i.CreatedByID,
		&i.UpdatedByID,
		&i.WorkflowOptions,
		&i.SyncOptions,
		&i.VirtualForeignKeys,
	)
	return i, err
}

const getJobConnectionDestination = `-- name: GetJobConnectionDestination :one
SELECT jdca.id, jdca.created_at, jdca.updated_at, jdca.job_id, jdca.connection_id, jdca.options from neosync_api.job_destination_connection_associations jdca
WHERE jdca.id = $1
`

func (q *Queries) GetJobConnectionDestination(ctx context.Context, db DBTX, id pgtype.UUID) (NeosyncApiJobDestinationConnectionAssociation, error) {
	row := db.QueryRow(ctx, getJobConnectionDestination, id)
	var i NeosyncApiJobDestinationConnectionAssociation
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.JobID,
		&i.ConnectionID,
		&i.Options,
	)
	return i, err
}

const getJobConnectionDestinations = `-- name: GetJobConnectionDestinations :many
SELECT jdca.id, jdca.created_at, jdca.updated_at, jdca.job_id, jdca.connection_id, jdca.options from neosync_api.job_destination_connection_associations jdca
INNER JOIN neosync_api.jobs j ON j.id = jdca.job_id
WHERE j.id = $1
ORDER BY jdca.created_at
`

func (q *Queries) GetJobConnectionDestinations(ctx context.Context, db DBTX, id pgtype.UUID) ([]NeosyncApiJobDestinationConnectionAssociation, error) {
	rows, err := db.Query(ctx, getJobConnectionDestinations, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []NeosyncApiJobDestinationConnectionAssociation
	for rows.Next() {
		var i NeosyncApiJobDestinationConnectionAssociation
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.JobID,
			&i.ConnectionID,
			&i.Options,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getJobConnectionDestinationsByJobIds = `-- name: GetJobConnectionDestinationsByJobIds :many
SELECT jdca.id, jdca.created_at, jdca.updated_at, jdca.job_id, jdca.connection_id, jdca.options from neosync_api.job_destination_connection_associations jdca
INNER JOIN neosync_api.jobs j ON j.id = jdca.job_id
WHERE j.id = ANY($1::uuid[])
ORDER BY j.created_at, jdca.created_at
`

func (q *Queries) GetJobConnectionDestinationsByJobIds(ctx context.Context, db DBTX, jobids []pgtype.UUID) ([]NeosyncApiJobDestinationConnectionAssociation, error) {
	rows, err := db.Query(ctx, getJobConnectionDestinationsByJobIds, jobids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []NeosyncApiJobDestinationConnectionAssociation
	for rows.Next() {
		var i NeosyncApiJobDestinationConnectionAssociation
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.JobID,
			&i.ConnectionID,
			&i.Options,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getJobsByAccount = `-- name: GetJobsByAccount :many
SELECT j.id, j.created_at, j.updated_at, j.name, j.account_id, j.status, j.connection_options, j.mappings, j.cron_schedule, j.created_by_id, j.updated_by_id, j.workflow_options, j.sync_options, j.virtual_foreign_keys from neosync_api.jobs j
INNER JOIN neosync_api.accounts a ON a.id = j.account_id
WHERE a.id = $1
ORDER BY j.created_at DESC
`

func (q *Queries) GetJobsByAccount(ctx context.Context, db DBTX, accountid pgtype.UUID) ([]NeosyncApiJob, error) {
	rows, err := db.Query(ctx, getJobsByAccount, accountid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []NeosyncApiJob
	for rows.Next() {
		var i NeosyncApiJob
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
			&i.AccountID,
			&i.Status,
			&i.ConnectionOptions,
			&i.Mappings,
			&i.CronSchedule,
			&i.CreatedByID,
			&i.UpdatedByID,
			&i.WorkflowOptions,
			&i.SyncOptions,
			&i.VirtualForeignKeys,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const isJobNameAvailable = `-- name: IsJobNameAvailable :one
SELECT count(j.id) from neosync_api.jobs j
INNER JOIN neosync_api.accounts a ON a.id = j.account_id
WHERE a.id = $1 AND j.name = $2
`

type IsJobNameAvailableParams struct {
	AccountId pgtype.UUID
	JobName   string
}

func (q *Queries) IsJobNameAvailable(ctx context.Context, db DBTX, arg IsJobNameAvailableParams) (int64, error) {
	row := db.QueryRow(ctx, isJobNameAvailable, arg.AccountId, arg.JobName)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const removeJobById = `-- name: RemoveJobById :exec
DELETE FROM neosync_api.jobs WHERE id = $1
`

func (q *Queries) RemoveJobById(ctx context.Context, db DBTX, id pgtype.UUID) error {
	_, err := db.Exec(ctx, removeJobById, id)
	return err
}

const removeJobConnectionDestination = `-- name: RemoveJobConnectionDestination :exec
DELETE FROM neosync_api.job_destination_connection_associations WHERE id = $1
`

func (q *Queries) RemoveJobConnectionDestination(ctx context.Context, db DBTX, id pgtype.UUID) error {
	_, err := db.Exec(ctx, removeJobConnectionDestination, id)
	return err
}

const removeJobConnectionDestinations = `-- name: RemoveJobConnectionDestinations :exec
DELETE FROM neosync_api.job_destination_connection_associations
WHERE id = ANY($1::uuid[])
`

func (q *Queries) RemoveJobConnectionDestinations(ctx context.Context, db DBTX, jobids []pgtype.UUID) error {
	_, err := db.Exec(ctx, removeJobConnectionDestinations, jobids)
	return err
}

const setJobSyncOptions = `-- name: SetJobSyncOptions :one
UPDATE neosync_api.jobs
SET sync_options = $1,
updated_by_id = $2
WHERE id = $3
RETURNING id, created_at, updated_at, name, account_id, status, connection_options, mappings, cron_schedule, created_by_id, updated_by_id, workflow_options, sync_options, virtual_foreign_keys
`

type SetJobSyncOptionsParams struct {
	SyncOptions *pg_models.ActivityOptions
	UpdatedByID pgtype.UUID
	ID          pgtype.UUID
}

func (q *Queries) SetJobSyncOptions(ctx context.Context, db DBTX, arg SetJobSyncOptionsParams) (NeosyncApiJob, error) {
	row := db.QueryRow(ctx, setJobSyncOptions, arg.SyncOptions, arg.UpdatedByID, arg.ID)
	var i NeosyncApiJob
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.AccountID,
		&i.Status,
		&i.ConnectionOptions,
		&i.Mappings,
		&i.CronSchedule,
		&i.CreatedByID,
		&i.UpdatedByID,
		&i.WorkflowOptions,
		&i.SyncOptions,
		&i.VirtualForeignKeys,
	)
	return i, err
}

const setJobWorkflowOptions = `-- name: SetJobWorkflowOptions :one
UPDATE neosync_api.jobs
SET workflow_options = $1,
updated_by_id = $2
WHERE id = $3
RETURNING id, created_at, updated_at, name, account_id, status, connection_options, mappings, cron_schedule, created_by_id, updated_by_id, workflow_options, sync_options, virtual_foreign_keys
`

type SetJobWorkflowOptionsParams struct {
	WorkflowOptions *pg_models.WorkflowOptions
	UpdatedByID     pgtype.UUID
	ID              pgtype.UUID
}

func (q *Queries) SetJobWorkflowOptions(ctx context.Context, db DBTX, arg SetJobWorkflowOptionsParams) (NeosyncApiJob, error) {
	row := db.QueryRow(ctx, setJobWorkflowOptions, arg.WorkflowOptions, arg.UpdatedByID, arg.ID)
	var i NeosyncApiJob
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.AccountID,
		&i.Status,
		&i.ConnectionOptions,
		&i.Mappings,
		&i.CronSchedule,
		&i.CreatedByID,
		&i.UpdatedByID,
		&i.WorkflowOptions,
		&i.SyncOptions,
		&i.VirtualForeignKeys,
	)
	return i, err
}

const updateJobConnectionDestination = `-- name: UpdateJobConnectionDestination :one
UPDATE neosync_api.job_destination_connection_associations
SET options = $1,
connection_id = $2
WHERE id = $3
RETURNING id, created_at, updated_at, job_id, connection_id, options
`

type UpdateJobConnectionDestinationParams struct {
	Options      *pg_models.JobDestinationOptions
	ConnectionID pgtype.UUID
	ID           pgtype.UUID
}

func (q *Queries) UpdateJobConnectionDestination(ctx context.Context, db DBTX, arg UpdateJobConnectionDestinationParams) (NeosyncApiJobDestinationConnectionAssociation, error) {
	row := db.QueryRow(ctx, updateJobConnectionDestination, arg.Options, arg.ConnectionID, arg.ID)
	var i NeosyncApiJobDestinationConnectionAssociation
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.JobID,
		&i.ConnectionID,
		&i.Options,
	)
	return i, err
}

const updateJobMappings = `-- name: UpdateJobMappings :one
UPDATE neosync_api.jobs
SET mappings = $1,
updated_by_id = $2
WHERE id = $3
RETURNING id, created_at, updated_at, name, account_id, status, connection_options, mappings, cron_schedule, created_by_id, updated_by_id, workflow_options, sync_options, virtual_foreign_keys
`

type UpdateJobMappingsParams struct {
	Mappings    []*pg_models.JobMapping
	UpdatedByID pgtype.UUID
	ID          pgtype.UUID
}

func (q *Queries) UpdateJobMappings(ctx context.Context, db DBTX, arg UpdateJobMappingsParams) (NeosyncApiJob, error) {
	row := db.QueryRow(ctx, updateJobMappings, arg.Mappings, arg.UpdatedByID, arg.ID)
	var i NeosyncApiJob
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.AccountID,
		&i.Status,
		&i.ConnectionOptions,
		&i.Mappings,
		&i.CronSchedule,
		&i.CreatedByID,
		&i.UpdatedByID,
		&i.WorkflowOptions,
		&i.SyncOptions,
		&i.VirtualForeignKeys,
	)
	return i, err
}

const updateJobSchedule = `-- name: UpdateJobSchedule :one
UPDATE neosync_api.jobs
SET cron_schedule = $1,
updated_by_id = $2
WHERE id = $3
RETURNING id, created_at, updated_at, name, account_id, status, connection_options, mappings, cron_schedule, created_by_id, updated_by_id, workflow_options, sync_options, virtual_foreign_keys
`

type UpdateJobScheduleParams struct {
	CronSchedule pgtype.Text
	UpdatedByID  pgtype.UUID
	ID           pgtype.UUID
}

func (q *Queries) UpdateJobSchedule(ctx context.Context, db DBTX, arg UpdateJobScheduleParams) (NeosyncApiJob, error) {
	row := db.QueryRow(ctx, updateJobSchedule, arg.CronSchedule, arg.UpdatedByID, arg.ID)
	var i NeosyncApiJob
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.AccountID,
		&i.Status,
		&i.ConnectionOptions,
		&i.Mappings,
		&i.CronSchedule,
		&i.CreatedByID,
		&i.UpdatedByID,
		&i.WorkflowOptions,
		&i.SyncOptions,
		&i.VirtualForeignKeys,
	)
	return i, err
}

const updateJobSource = `-- name: UpdateJobSource :one
UPDATE neosync_api.jobs
SET connection_options = $1,
updated_by_id = $2
WHERE id = $3
RETURNING id, created_at, updated_at, name, account_id, status, connection_options, mappings, cron_schedule, created_by_id, updated_by_id, workflow_options, sync_options, virtual_foreign_keys
`

type UpdateJobSourceParams struct {
	ConnectionOptions *pg_models.JobSourceOptions
	UpdatedByID       pgtype.UUID
	ID                pgtype.UUID
}

func (q *Queries) UpdateJobSource(ctx context.Context, db DBTX, arg UpdateJobSourceParams) (NeosyncApiJob, error) {
	row := db.QueryRow(ctx, updateJobSource, arg.ConnectionOptions, arg.UpdatedByID, arg.ID)
	var i NeosyncApiJob
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.AccountID,
		&i.Status,
		&i.ConnectionOptions,
		&i.Mappings,
		&i.CronSchedule,
		&i.CreatedByID,
		&i.UpdatedByID,
		&i.WorkflowOptions,
		&i.SyncOptions,
		&i.VirtualForeignKeys,
	)
	return i, err
}

const updateJobVirtualForeignKeys = `-- name: UpdateJobVirtualForeignKeys :one
UPDATE neosync_api.jobs
SET virtual_foreign_keys = $1,
updated_by_id = $2
WHERE id = $3
RETURNING id, created_at, updated_at, name, account_id, status, connection_options, mappings, cron_schedule, created_by_id, updated_by_id, workflow_options, sync_options, virtual_foreign_keys
`

type UpdateJobVirtualForeignKeysParams struct {
	VirtualForeignKeys []*pg_models.VirtualForeignConstraint
	UpdatedByID        pgtype.UUID
	ID                 pgtype.UUID
}

func (q *Queries) UpdateJobVirtualForeignKeys(ctx context.Context, db DBTX, arg UpdateJobVirtualForeignKeysParams) (NeosyncApiJob, error) {
	row := db.QueryRow(ctx, updateJobVirtualForeignKeys, arg.VirtualForeignKeys, arg.UpdatedByID, arg.ID)
	var i NeosyncApiJob
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.AccountID,
		&i.Status,
		&i.ConnectionOptions,
		&i.Mappings,
		&i.CronSchedule,
		&i.CreatedByID,
		&i.UpdatedByID,
		&i.WorkflowOptions,
		&i.SyncOptions,
		&i.VirtualForeignKeys,
	)
	return i, err
}