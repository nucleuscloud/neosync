package datasync_activities

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	pg_queries "github.com/nucleuscloud/neosync/worker/gen/go/db/postgresql"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	dbschemas_mysql "github.com/nucleuscloud/neosync/worker/internal/dbschemas/mysql"
	dbschemas_postgres "github.com/nucleuscloud/neosync/worker/internal/dbschemas/postgres"

	"go.temporal.io/sdk/log"
)

type benthosBuilder struct {
	pgpool    map[string]pg_queries.DBTX
	pgquerier pg_queries.Querier

	mysqlpool map[string]dbschemas_mysql.DBTX

	jobclient  mgmtv1alpha1connect.JobServiceClient
	connclient mgmtv1alpha1connect.ConnectionServiceClient
}

func newBenthosBuilder(
	pgpool map[string]pg_queries.DBTX,
	pgquerier pg_queries.Querier,

	mysqlpool map[string]dbschemas_mysql.DBTX,

	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
) *benthosBuilder {
	return &benthosBuilder{
		pgpool:     pgpool,
		pgquerier:  pgquerier,
		mysqlpool:  mysqlpool,
		jobclient:  jobclient,
		connclient: connclient,
	}
}

func (b *benthosBuilder) GenerateBenthosConfigs(
	ctx context.Context,
	req *GenerateBenthosConfigsRequest,
	logger log.Logger,
) (*GenerateBenthosConfigsResponse, error) {
	job, err := b.getJobById(ctx, req.JobId)
	if err != nil {
		return nil, err
	}
	responses := []*BenthosConfigResponse{}

	sourceConnection, err := b.getConnectionById(ctx, job.Source.ConnectionId)
	if err != nil {
		return nil, err
	}

	groupedMappings := groupMappingsByTable(job.Mappings)

	switch connection := sourceConnection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		dsn, err := getPgDsn(connection.PgConfig)
		if err != nil {
			return nil, err
		}

		sqlOpts := job.Source.Options.GetPostgresOptions()
		var sourceTableOpts map[string]*sourceTableOptions
		if sqlOpts != nil {
			sourceTableOpts = groupPostgresSourceOptionsByTable(sqlOpts.Schemas)
		}

		sourceResponses, err := buildBenthosSourceConfigReponses(groupedMappings, dsn, "postgres", sourceTableOpts)
		if err != nil {
			return nil, err
		}
		responses = append(responses, sourceResponses...)

		if _, ok := b.pgpool[dsn]; !ok {
			pool, err := pgxpool.New(ctx, dsn)
			if err != nil {
				return nil, err
			}
			defer pool.Close()
			b.pgpool[dsn] = pool
		}
		pool := b.pgpool[dsn]

		// validate job mappings align with sql connections
		dbschemas, err := b.pgquerier.GetDatabaseSchema(ctx, pool)
		if err != nil {
			return nil, err
		}
		groupedSchemas := dbschemas_postgres.GetUniqueSchemaColMappings(dbschemas)
		if !areMappingsSubsetOfSchemas(groupedSchemas, job.Mappings) {
			return nil, errors.New("job mappings are not equal to or a subset of the database schema found in the source connection")
		}
		if sqlOpts != nil && sqlOpts.HaltOnNewColumnAddition &&
			shouldHaltOnSchemaAddition(groupedSchemas, job.Mappings) {
			msg := "job mappings does not contain a column mapping for all " +
				"columns found in the source connection for the selected schemas and tables"
			return nil, errors.New(msg)
		}

		allConstraints, err := b.getAllPostgresFkConstraintsFromMappings(ctx, pool, job.Mappings)
		if err != nil {
			return nil, err
		}
		td := dbschemas_postgres.GetPostgresTableDependencies(allConstraints)

		for _, resp := range responses {
			dependsOn, ok := td[resp.Name]
			if ok {
				resp.DependsOn = dependsOn
			}
		}
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		dsn, err := getMysqlDsn(connection.MysqlConfig)
		if err != nil {
			return nil, err
		}

		sqlOpts := job.Source.Options.GetMysqlOptions()
		var sourceTableOpts map[string]*sourceTableOptions
		if sqlOpts != nil {
			sourceTableOpts = groupMysqlSourceOptionsByTable(sqlOpts.Schemas)
		}

		sourceResponses, err := buildBenthosSourceConfigReponses(groupedMappings, dsn, "mysql", sourceTableOpts)
		if err != nil {
			return nil, err
		}
		responses = append(responses, sourceResponses...)

		if _, ok := b.mysqlpool[dsn]; !ok {
			pool, err := sql.Open("mysql", dsn)
			if err != nil {
				return nil, err
			}
			defer pool.Close()
			b.mysqlpool[dsn] = pool
		}
		pool := b.mysqlpool[dsn]

		// validate job mappings align with sql connections
		dbschemas, err := dbschemas_mysql.GetDatabaseSchemas(ctx, pool)
		if err != nil {
			return nil, err
		}
		groupedSchemas := dbschemas_mysql.GetUniqueSchemaColMappings(dbschemas)
		if !areMappingsSubsetOfSchemas(groupedSchemas, job.Mappings) {
			return nil, errors.New("job mappings are not equal to or a subset of the database schema found in the source connection")
		}
		if sqlOpts != nil && sqlOpts.HaltOnNewColumnAddition &&
			shouldHaltOnSchemaAddition(groupedSchemas, job.Mappings) {
			msg := "job mappings does not contain a column mapping for all " +
				"columns found in the source connection for the selected schemas and tables"
			return nil, errors.New(msg)
		}

		allConstraints, err := b.getAllMysqlFkConstraintsFromMappings(ctx, pool, job.Mappings)
		if err != nil {
			return nil, err
		}
		td := dbschemas_mysql.GetMysqlTableDependencies(allConstraints)

		for _, resp := range responses {
			dependsOn, ok := td[resp.Name]
			if ok {
				resp.DependsOn = dependsOn
			}
		}

	default:
		return nil, fmt.Errorf("unsupported source connection")
	}

	for _, destination := range job.Destinations {
		destinationConnection, err := b.getConnectionById(ctx, destination.ConnectionId)
		if err != nil {
			return nil, err
		}
		for _, resp := range responses {
			switch connection := destinationConnection.ConnectionConfig.Config.(type) {
			case *mgmtv1alpha1.ConnectionConfig_PgConfig:
				dsn, err := getPgDsn(connection.PgConfig)
				if err != nil {
					return nil, err
				}

				truncateBeforeInsert := false
				truncateCascade := false
				initSchema := false
				sqlOpts := destination.Options.GetPostgresOptions()
				if sqlOpts != nil {
					initSchema = sqlOpts.InitTableSchema
					if sqlOpts.TruncateTable != nil {
						truncateBeforeInsert = sqlOpts.TruncateTable.TruncateBeforeInsert
						truncateCascade = sqlOpts.TruncateTable.Cascade
					}
				}

				pool := b.pgpool[resp.Config.Input.SqlSelect.Dsn]
				// todo: make this more efficient to reduce amount of times we have to connect to the source database
				schema, table := splitTableKey(resp.Config.Input.SqlSelect.Table)
				initStmt, err := b.getInitStatementFromPostgres(
					ctx,
					pool,
					schema,
					table,
					&initStatementOpts{
						TruncateBeforeInsert: truncateBeforeInsert,
						TruncateCascade:      truncateCascade,
						InitSchema:           initSchema,
					},
				)
				if err != nil {
					return nil, err
				}
				logger.Info(fmt.Sprintf("sql batch count: %d", maxPgParamLimit/len(resp.Config.Input.SqlSelect.Columns)))
				resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
					SqlInsert: &neosync_benthos.SqlInsert{
						Driver: "postgres",
						Dsn:    dsn,

						Table:         resp.Config.Input.SqlSelect.Table,
						Columns:       resp.Config.Input.SqlSelect.Columns,
						ArgsMapping:   buildPlainInsertArgs(resp.Config.Input.SqlSelect.Columns),
						InitStatement: initStmt,

						ConnMaxIdle: 2,
						ConnMaxOpen: 2,

						Batching: &neosync_benthos.Batching{
							Period: "1s",
							// max allowed by postgres in a single batch
							Count: computeMaxPgBatchCount(len(resp.Config.Input.SqlSelect.Columns)),
						},
					},
				})
			case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
				dsn, err := getMysqlDsn(connection.MysqlConfig)
				if err != nil {
					return nil, err
				}

				truncateBeforeInsert := false
				initSchema := false
				sqlOpts := destination.Options.GetMysqlOptions()
				if sqlOpts != nil {
					initSchema = sqlOpts.InitTableSchema
					if sqlOpts.TruncateTable != nil {
						truncateBeforeInsert = sqlOpts.TruncateTable.TruncateBeforeInsert
					}
				}

				pool := b.mysqlpool[resp.Config.Input.SqlSelect.Dsn]
				// todo: make this more efficient to reduce amount of times we have to connect to the source database
				schema, table := splitTableKey(resp.Config.Input.SqlSelect.Table)
				initStmt, err := b.getInitStatementFromMysql(
					ctx,
					pool,
					schema,
					table,
					&initStatementOpts{
						TruncateBeforeInsert: truncateBeforeInsert,
						InitSchema:           initSchema,
					},
				)
				if err != nil {
					return nil, err
				}
				logger.Info(fmt.Sprintf("sql batch count: %d", maxPgParamLimit/len(resp.Config.Input.SqlSelect.Columns)))
				resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
					SqlInsert: &neosync_benthos.SqlInsert{
						Driver: "mysql",
						Dsn:    dsn,

						Table:         resp.Config.Input.SqlSelect.Table,
						Columns:       resp.Config.Input.SqlSelect.Columns,
						ArgsMapping:   buildPlainInsertArgs(resp.Config.Input.SqlSelect.Columns),
						InitStatement: initStmt,

						ConnMaxIdle: 2,
						ConnMaxOpen: 2,

						Batching: &neosync_benthos.Batching{
							Period: "1s",
							// max allowed by postgres in a single batch
							Count: computeMaxPgBatchCount(len(resp.Config.Input.SqlSelect.Columns)),
						},
					},
				})

			case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
				s3pathpieces := []string{}
				if connection.AwsS3Config.PathPrefix != nil && *connection.AwsS3Config.PathPrefix != "" {
					s3pathpieces = append(s3pathpieces, strings.Trim(*connection.AwsS3Config.PathPrefix, "/"))
				}
				s3pathpieces = append(
					s3pathpieces,
					"workflows",
					req.WorkflowId,
					"activities",
					resp.Name, // may need to do more here
					"data",
					`${!count("files")}.json.gz}`,
				)

				resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
					AwsS3: &neosync_benthos.AwsS3Insert{
						Bucket:      connection.AwsS3Config.BucketArn,
						MaxInFlight: 64,
						Path:        fmt.Sprintf("/%s", strings.Join(s3pathpieces, "/")),
						Batching: &neosync_benthos.Batching{
							Count:  100,
							Period: "1s",
							Processors: []*neosync_benthos.BatchProcessor{
								{Archive: &neosync_benthos.ArchiveProcessor{Format: "json_array"}},
								{Compress: &neosync_benthos.CompressProcessor{Algorithm: "gzip"}},
							},
						},
						Credentials: buildBenthosS3Credentials(connection.AwsS3Config.Credentials),
						Region:      connection.AwsS3Config.GetRegion(),
						Endpoint:    connection.AwsS3Config.GetEndpoint(),
					},
				})
			default:
				return nil, fmt.Errorf("unsupported destination connection config")
			}
		}
	}

	return &GenerateBenthosConfigsResponse{
		BenthosConfigs: responses,
	}, nil
}

func (b *benthosBuilder) getJobById(
	ctx context.Context,
	jobId string,
) (*mgmtv1alpha1.Job, error) {
	getjobResp, err := b.jobclient.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: jobId,
	}))
	if err != nil {
		return nil, err
	}

	return getjobResp.Msg.Job, nil
}

func (b *benthosBuilder) getConnectionById(
	ctx context.Context,
	connectionId string,
) (*mgmtv1alpha1.Connection, error) {
	getConnResp, err := b.connclient.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: connectionId,
	}))
	if err != nil {
		return nil, err
	}
	return getConnResp.Msg.Connection, nil
}
