package nucleusdb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	jsonmodels "github.com/nucleuscloud/neosync/backend/internal/nucleusdb/json-models"
)

func (d *NucleusDb) SetSqlSourceSubsets(
	ctx context.Context,
	jobId pgtype.UUID,
	schemas *mgmtv1alpha1.JobSourceSqlSubetSchemas,
	userUuid pgtype.UUID,
) error {
	return d.WithTx(ctx, nil, func(q *db_queries.Queries) error {
		dbjob, err := q.GetJobById(ctx, jobId)
		if err != nil {
			return err
		}
		switch s := schemas.Schemas.(type) {
		case *mgmtv1alpha1.JobSourceSqlSubetSchemas_PostgresSubset:
			if dbjob.ConnectionOptions.PostgresOptions == nil {
				dbjob.ConnectionOptions.PostgresOptions = &jsonmodels.PostgresSourceOptions{}
			}
			dbjob.ConnectionOptions.PostgresOptions.Schemas = jsonmodels.FromDtoPostgresSourceSchemaOptions(s.PostgresSubset.PostgresSchemas)

		case *mgmtv1alpha1.JobSourceSqlSubetSchemas_MysqlSubset:
			if dbjob.ConnectionOptions.MysqlOptions == nil {
				dbjob.ConnectionOptions.MysqlOptions = &jsonmodels.MysqlSourceOptions{}
			}
			dbjob.ConnectionOptions.MysqlOptions.Schemas = jsonmodels.FromDtoMysqlSourceSchemaOptions(s.MysqlSubset.MysqlSchemas)
		default:
			return fmt.Errorf("this connection config is not currently supported")
		}

		_, err = q.UpdateJobSource(ctx, db_queries.UpdateJobSourceParams{
			ID:                 jobId,
			ConnectionSourceID: dbjob.ConnectionSourceID,
			ConnectionOptions:  dbjob.ConnectionOptions,

			UpdatedByID: userUuid,
		})
		if err != nil {
			return fmt.Errorf("unable to update job source with new subsets: %w", err)
		}
		return nil
	})
}
