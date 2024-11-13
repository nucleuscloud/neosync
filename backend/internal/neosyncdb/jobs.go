package neosyncdb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
)

type CreateJobConnectionDestination struct {
	ConnectionId pgtype.UUID
	Options      *pg_models.JobDestinationOptions
}

func (d *NeosyncDb) CreateJob(
	ctx context.Context,
	cjParams *db_queries.CreateJobParams,
	destinations []*CreateJobConnectionDestination,
) (*db_queries.NeosyncApiJob, error) {
	var createdJob *db_queries.NeosyncApiJob
	if err := d.WithTx(ctx, nil, func(tx BaseDBTX) error {
		job, err := d.Q.CreateJob(ctx, tx, *cjParams)
		if err != nil {
			return err
		}
		if len(destinations) > 0 {
			destParams := make([]db_queries.CreateJobConnectionDestinationsParams, 0, len(destinations))
			for i := range destinations {
				destParams = append(destParams, db_queries.CreateJobConnectionDestinationsParams{
					JobID:        job.ID,
					ConnectionID: destinations[i].ConnectionId,
					Options:      destinations[i].Options,
				})
			}
			if _, err := d.Q.CreateJobConnectionDestinations(ctx, tx, destParams); err != nil {
				return err
			}
		}
		createdJob = &job
		return nil
	}); err != nil {
		return nil, err
	}
	return createdJob, nil
}

func (d *NeosyncDb) SetSourceSubsets(
	ctx context.Context,
	jobId pgtype.UUID,
	schemas *mgmtv1alpha1.JobSourceSqlSubetSchemas,
	subsetByForeignKeyConstraints bool,
	userUuid pgtype.UUID,
) error {
	return d.WithTx(ctx, nil, func(dbtx BaseDBTX) error {
		dbjob, err := d.Q.GetJobById(ctx, dbtx, jobId)
		if err != nil {
			return err
		}
		switch s := schemas.Schemas.(type) {
		case *mgmtv1alpha1.JobSourceSqlSubetSchemas_PostgresSubset:
			if dbjob.ConnectionOptions.PostgresOptions == nil {
				dbjob.ConnectionOptions.PostgresOptions = &pg_models.PostgresSourceOptions{}
			}
			dbjob.ConnectionOptions.PostgresOptions.Schemas = pg_models.FromDtoPostgresSourceSchemaOptions(s.PostgresSubset.GetPostgresSchemas())
			dbjob.ConnectionOptions.PostgresOptions.SubsetByForeignKeyConstraints = subsetByForeignKeyConstraints

		case *mgmtv1alpha1.JobSourceSqlSubetSchemas_MysqlSubset:
			if dbjob.ConnectionOptions.MysqlOptions == nil {
				dbjob.ConnectionOptions.MysqlOptions = &pg_models.MysqlSourceOptions{}
			}
			dbjob.ConnectionOptions.MysqlOptions.Schemas = pg_models.FromDtoMysqlSourceSchemaOptions(s.MysqlSubset.GetMysqlSchemas())
			dbjob.ConnectionOptions.MysqlOptions.SubsetByForeignKeyConstraints = subsetByForeignKeyConstraints
		case *mgmtv1alpha1.JobSourceSqlSubetSchemas_MssqlSubset:
			if dbjob.ConnectionOptions.MssqlOptions == nil {
				dbjob.ConnectionOptions.MssqlOptions = &pg_models.MssqlSourceOptions{}
			}
			dbjob.ConnectionOptions.MssqlOptions.Schemas = pg_models.FromDtoMssqlSourceSchemaOptions(s.MssqlSubset.GetMssqlSchemas())
			dbjob.ConnectionOptions.MssqlOptions.SubsetByForeignKeyConstraints = subsetByForeignKeyConstraints
		case *mgmtv1alpha1.JobSourceSqlSubetSchemas_DynamodbSubset:
			if dbjob.ConnectionOptions.DynamoDBOptions == nil {
				dbjob.ConnectionOptions.DynamoDBOptions = &pg_models.DynamoDBSourceOptions{}
			}
			dbjob.ConnectionOptions.DynamoDBOptions.Tables = pg_models.FromDtoDynamoDBSourceTableOptions(s.DynamodbSubset.GetTables())
		default:
			return fmt.Errorf("this connection config is not currently supported: %T", s)
		}

		_, err = d.Q.UpdateJobSource(ctx, dbtx, db_queries.UpdateJobSourceParams{
			ID:                jobId,
			ConnectionOptions: dbjob.ConnectionOptions,

			UpdatedByID: userUuid,
		})
		if err != nil {
			return fmt.Errorf("unable to update job source with new subsets: %w", err)
		}
		return nil
	})
}
