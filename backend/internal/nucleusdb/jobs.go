package nucleusdb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	jsonmodels "github.com/nucleuscloud/neosync/backend/internal/nucleusdb/json-models"
)

func (d *NucleusDb) SetSqlSourceSubsets(
	ctx context.Context,
	jobId pgtype.UUID,
	sqlSourceSchemaOptions []*jsonmodels.SqlSourceSchemaOption,
	userUuid pgtype.UUID,
) error {
	return d.WithTx(ctx, nil, func(q *db_queries.Queries) error {
		dbjob, err := q.GetJobById(ctx, jobId)
		if err != nil {
			return err
		}
		if dbjob.ConnectionOptions.SqlOptions == nil {
			dbjob.ConnectionOptions.SqlOptions = &jsonmodels.SqlSourceOptions{}
		}
		dbjob.ConnectionOptions.SqlOptions.Schemas = sqlSourceSchemaOptions

		dbjob.ConnectionOptions.SqlOptions.Schemas = nil
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
