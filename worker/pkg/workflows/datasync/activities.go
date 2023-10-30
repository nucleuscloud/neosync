package datasync

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"go.temporal.io/sdk/activity"
	"golang.org/x/sync/errgroup"

	_ "github.com/benthosdev/benthos/v4/public/components/aws"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	_ "github.com/benthosdev/benthos/v4/public/components/pure"
	_ "github.com/benthosdev/benthos/v4/public/components/pure/extended"
	_ "github.com/benthosdev/benthos/v4/public/components/sql"
	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	_ "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers"
	dbschemas_mysql "github.com/nucleuscloud/neosync/worker/internal/dbschemas/mysql"
	dbschemas_postgres "github.com/nucleuscloud/neosync/worker/internal/dbschemas/postgres"
)

const nullString = "null"

type GenerateBenthosConfigsRequest struct {
	JobId      string
	BackendUrl string
	WorkflowId string
}
type GenerateBenthosConfigsResponse struct {
	BenthosConfigs []*benthosConfigResponse
}

type benthosConfigResponse struct {
	Name      string
	DependsOn []string
	Config    *neosync_benthos.BenthosConfig
}

type Activities struct{}

func (a *Activities) GenerateBenthosConfigs(
	ctx context.Context,
	req *GenerateBenthosConfigsRequest,
) (*GenerateBenthosConfigsResponse, error) {
	logger := activity.GetLogger(ctx)
	_ = logger
	go func() {
		for {
			select {
			case <-time.After(1 * time.Second):
				activity.RecordHeartbeat(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	pgpoolmap := map[string]*pgxpool.Pool{}
	mysqlPoolMap := map[string]*sql.DB{}

	job, err := a.getJobById(ctx, req.BackendUrl, req.JobId)
	if err != nil {
		return nil, err
	}
	responses := []*benthosConfigResponse{}

	sourceConnection, err := a.getConnectionById(ctx, req.BackendUrl, job.Source.ConnectionId)
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

		if _, ok := pgpoolmap[dsn]; !ok {
			pool, err := pgxpool.New(ctx, dsn)
			if err != nil {
				return nil, err
			}
			defer pool.Close()
			pgpoolmap[dsn] = pool
		}
		pool := pgpoolmap[dsn]

		// validate job mappings align with sql connections
		dbschemas, err := dbschemas_postgres.GetDatabaseSchemas(ctx, pool)
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

		allConstraints, err := a.getAllPostgresFkConstraintsFromMappings(ctx, pool, job.Mappings)
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

		if _, ok := mysqlPoolMap[dsn]; !ok {
			pool, err := sql.Open("mysql", dsn)
			if err != nil {
				return nil, err
			}
			defer pool.Close()
			mysqlPoolMap[dsn] = pool
		}
		pool := mysqlPoolMap[dsn]

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

		allConstraints, err := a.getAllMysqlFkConstraintsFromMappings(ctx, pool, job.Mappings)
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
		destinationConnection, err := a.getConnectionById(ctx, req.BackendUrl, destination.ConnectionId)
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

				pool := pgpoolmap[resp.Config.Input.SqlSelect.Dsn]
				// todo: make this more efficient to reduce amount of times we have to connect to the source database
				schema, table := splitTableKey(resp.Config.Input.SqlSelect.Table)
				initStmt, err := a.getInitStatementFromPostgres(
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

				pool := mysqlPoolMap[resp.Config.Input.SqlSelect.Dsn]
				// todo: make this more efficient to reduce amount of times we have to connect to the source database
				schema, table := splitTableKey(resp.Config.Input.SqlSelect.Table)
				initStmt, err := a.getInitStatementFromMysql(
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

type sourceTableOptions struct {
	WhereClause *string
}

func buildBenthosSourceConfigReponses(
	mappings []*TableMapping,
	dsn string,
	driver string,
	sourceTableOpts map[string]*sourceTableOptions,
) ([]*benthosConfigResponse, error) {
	responses := []*benthosConfigResponse{}

	for i := range mappings {
		tableMapping := mappings[i]
		cols := buildPlainColumns(tableMapping.Mappings)
		if areAllColsNull(tableMapping.Mappings) {
			// skipping table as no columns are mapped
			continue
		}

		var where string
		tableOpt := sourceTableOpts[neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table)]
		if tableOpt != nil && tableOpt.WhereClause != nil {
			where = *tableOpt.WhereClause
		}

		bc := &neosync_benthos.BenthosConfig{
			StreamConfig: neosync_benthos.StreamConfig{
				Input: &neosync_benthos.InputConfig{
					Inputs: neosync_benthos.Inputs{
						SqlSelect: &neosync_benthos.SqlSelect{
							Driver: driver,
							Dsn:    dsn,

							Table:   neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table),
							Where:   where,
							Columns: cols,
						},
					},
				},
				Pipeline: &neosync_benthos.PipelineConfig{
					Threads:    -1,
					Processors: []neosync_benthos.ProcessorConfig{},
				},
				Output: &neosync_benthos.OutputConfig{
					Outputs: neosync_benthos.Outputs{
						Broker: &neosync_benthos.OutputBrokerConfig{
							Pattern: "fan_out",
							Outputs: []neosync_benthos.Outputs{},
						},
					},
				},
			},
		}
		mutation, err := buildProcessorMutation(tableMapping.Mappings)
		if err != nil {
			return nil, err
		}
		if mutation != "" {
			bc.StreamConfig.Pipeline.Processors = append(bc.StreamConfig.Pipeline.Processors, neosync_benthos.ProcessorConfig{
				Mutation: mutation,
			})
		}
		responses = append(responses, &benthosConfigResponse{
			Name:      neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table), // todo: may need to expand on this
			Config:    bc,
			DependsOn: []string{},
		})
	}

	return responses, nil
}

func buildBenthosS3Credentials(mgmtCreds *mgmtv1alpha1.AwsS3Credentials) *neosync_benthos.AwsCredentials {
	if mgmtCreds == nil {
		return nil
	}
	creds := &neosync_benthos.AwsCredentials{}
	if mgmtCreds.Profile != nil {
		creds.Profile = *mgmtCreds.Profile
	}
	if mgmtCreds.AccessKeyId != nil {
		creds.Id = *mgmtCreds.AccessKeyId
	}
	if mgmtCreds.SecretAccessKey != nil {
		creds.Secret = *mgmtCreds.SecretAccessKey
	}
	if mgmtCreds.SessionToken != nil {
		creds.Token = *mgmtCreds.SessionToken
	}
	if mgmtCreds.FromEc2Role != nil {
		creds.FromEc2Role = *mgmtCreds.FromEc2Role
	}
	if mgmtCreds.RoleArn != nil {
		creds.Role = *mgmtCreds.RoleArn
	}
	if mgmtCreds.RoleExternalId != nil {
		creds.RoleExternalId = *mgmtCreds.RoleExternalId
	}

	return creds
}

const (
	maxPgParamLimit = 65535
)

func computeMaxPgBatchCount(numCols int) int {
	if numCols < 1 {
		return maxPgParamLimit
	}
	return clampInt(maxPgParamLimit/numCols, 1, maxPgParamLimit) // automatically rounds down
}

// clamps the input between low, high
func clampInt(input, low, high int) int {
	if input < low {
		return low
	}
	if input > high {
		return high
	}
	return input
}

func areMappingsSubsetOfSchemas(
	groupedSchemas map[string]map[string]struct{},
	mappings []*mgmtv1alpha1.JobMapping,
) bool {
	tableColMappings := getUniqueColMappingsMap(mappings)

	for key := range groupedSchemas {
		// For this method, we only care about the schemas+tables that we currently have mappings for
		if _, ok := tableColMappings[key]; !ok {
			delete(groupedSchemas, key)
		}
	}

	if len(tableColMappings) != len(groupedSchemas) {
		return false
	}

	// tests to make sure that every column in the col mappings is present in the db schema
	for table, cols := range tableColMappings {
		schemaCols, ok := groupedSchemas[table]
		if !ok {
			return false
		}
		// job mappings has more columns than the schema
		if len(cols) > len(schemaCols) {
			return false
		}
		for col := range cols {
			if _, ok := schemaCols[col]; !ok {
				return false
			}
		}
	}
	return true
}

func getUniqueColMappingsMap(
	mappings []*mgmtv1alpha1.JobMapping,
) map[string]map[string]struct{} {
	tableColMappings := map[string]map[string]struct{}{}
	for _, mapping := range mappings {
		key := neosync_benthos.BuildBenthosTable(mapping.Schema, mapping.Table)
		if _, ok := tableColMappings[key]; ok {
			tableColMappings[key][mapping.Column] = struct{}{}
		} else {
			tableColMappings[key] = map[string]struct{}{
				mapping.Column: {},
			}
		}
	}
	return tableColMappings
}

func shouldHaltOnSchemaAddition(
	groupedSchemas map[string]map[string]struct{},
	mappings []*mgmtv1alpha1.JobMapping,
) bool {
	tableColMappings := getUniqueColMappingsMap(mappings)

	if len(tableColMappings) != len(groupedSchemas) {
		return true
	}

	for table, cols := range groupedSchemas {
		mappingCols, ok := tableColMappings[table]
		if !ok {
			return true
		}
		if len(cols) > len(mappingCols) {
			return true
		}
		for col := range cols {
			if _, ok := mappingCols[col]; !ok {
				return true
			}
		}
	}
	return false
}

func (a *Activities) getAllPostgresFkConstraintsFromMappings(
	ctx context.Context,
	conn DBTX,
	mappings []*mgmtv1alpha1.JobMapping,
) ([]*dbschemas_postgres.ForeignKeyConstraint, error) {
	uniqueSchemas := getUniqueSchemasFromMappings(mappings)
	holder := make([][]*dbschemas_postgres.ForeignKeyConstraint, len(uniqueSchemas))
	errgrp, errctx := errgroup.WithContext(ctx)
	for idx := range uniqueSchemas {
		idx := idx
		schema := uniqueSchemas[idx]
		errgrp.Go(func() error {
			constraints, err := dbschemas_postgres.GetForeignKeyConstraints(errctx, conn, schema)
			if err != nil {
				return err
			}
			holder[idx] = constraints
			return nil
		})
	}

	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	output := []*dbschemas_postgres.ForeignKeyConstraint{}
	for _, schemas := range holder {
		output = append(output, schemas...)
	}
	return output, nil
}

func (a *Activities) getAllMysqlFkConstraintsFromMappings(
	ctx context.Context,
	conn *sql.DB,
	mappings []*mgmtv1alpha1.JobMapping,
) ([]*dbschemas_mysql.ForeignKeyConstraint, error) {
	uniqueSchemas := getUniqueSchemasFromMappings(mappings)
	holder := make([][]*dbschemas_mysql.ForeignKeyConstraint, len(uniqueSchemas))
	errgrp, errctx := errgroup.WithContext(ctx)
	for idx := range uniqueSchemas {
		idx := idx
		schema := uniqueSchemas[idx]
		errgrp.Go(func() error {
			constraints, err := dbschemas_mysql.GetForeignKeyConstraints(errctx, conn, schema)
			if err != nil {
				return err
			}
			holder[idx] = constraints
			return nil
		})
	}

	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	output := []*dbschemas_mysql.ForeignKeyConstraint{}
	for _, schemas := range holder {
		output = append(output, schemas...)
	}
	return output, nil
}

type initStatementOpts struct {
	TruncateBeforeInsert bool
	TruncateCascade      bool // only applied if truncatebeforeinsert is true
	InitSchema           bool
}

type DBTX interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func (a *Activities) getInitStatementFromPostgres(
	ctx context.Context,
	conn DBTX,
	schema string,
	table string,
	opts *initStatementOpts,
) (string, error) {

	statements := []string{}
	if opts != nil && opts.InitSchema {
		stmt, err := dbschemas_postgres.GetTableCreateStatement(ctx, conn, &dbschemas_postgres.GetTableCreateStatementRequest{
			Schema: schema,
			Table:  table,
		})
		if err != nil {
			return "", err
		}
		statements = append(statements, stmt)
	}
	if opts != nil && opts.TruncateBeforeInsert {
		if opts.TruncateCascade {
			statements = append(statements, fmt.Sprintf("TRUNCATE TABLE %s.%s CASCADE;", schema, table))
		} else {
			statements = append(statements, fmt.Sprintf("TRUNCATE TABLE %s.%s;", schema, table))
		}
	}
	return strings.Join(statements, "\n"), nil
}

func (a *Activities) getInitStatementFromMysql(
	ctx context.Context,
	conn *sql.DB,
	schema string,
	table string,
	opts *initStatementOpts,
) (string, error) {

	statements := []string{}
	if opts != nil && opts.InitSchema {
		stmt, err := dbschemas_mysql.GetTableCreateStatement(ctx, conn, &dbschemas_mysql.GetTableCreateStatementRequest{
			Schema: schema,
			Table:  table,
		})
		if err != nil {
			return "", err
		}
		statements = append(statements, stmt)
	}
	if opts != nil && opts.TruncateBeforeInsert {
		statements = append(statements, fmt.Sprintf("TRUNCATE TABLE %s.%s;", schema, table))
	}
	return strings.Join(statements, "\n"), nil
}

func areAllColsNull(mappings []*mgmtv1alpha1.JobMapping) bool {
	for _, col := range mappings {
		if col.Transformer.Value != nullString {
			return false
		}
	}
	return true
}

func buildPlainColumns(mappings []*mgmtv1alpha1.JobMapping) []string {
	columns := []string{}

	for _, col := range mappings {
		columns = append(columns, col.Column)
	}

	return columns
}

func splitTableKey(key string) (schema, table string) {
	pieces := strings.Split(key, ".")
	if len(pieces) == 1 {
		return "public", pieces[0]
	}
	return pieces[0], pieces[1]
}

// used to record metadata in activity event history
type SyncMetadata struct {
	Schema string
	Table  string
}
type SyncRequest struct {
	BenthosConfig string
}
type SyncResponse struct{}

func (a *Activities) Sync(ctx context.Context, req *SyncRequest, metadata *SyncMetadata) (*SyncResponse, error) {
	logger := activity.GetLogger(ctx)
	var benthosStream *service.Stream
	go func() {
		for {
			select {
			case <-time.After(1 * time.Second):
				activity.RecordHeartbeat(ctx)
			case <-ctx.Done():
				if benthosStream != nil {
					// this must be here because stream.Run(ctx) doesn't seem to fully obey a canceled context when
					// a sink is in an error state. We want to explicitly call stop here because the workflow has been canceled.
					err := benthosStream.Stop(ctx)
					if err != nil {
						logger.Error(err.Error())
					}
				}
				return
			}
		}
	}()

	streambldr := service.NewStreamBuilder()
	// would ideally use the activity logger here but can't convert it into a slog.
	benthoslogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	streambldr.SetLogger(benthoslogger.With(
		"metadata", metadata,
		"benthos", "true",
	))

	err := streambldr.SetYAML(req.BenthosConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to convert benthos config to yaml for stream builder: %w", err)
	}

	stream, err := streambldr.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to build benthos stream: %w", err)
	}
	benthosStream = stream

	err = stream.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to run benthos stream: %w", err)
	}
	benthosStream = nil
	return &SyncResponse{}, nil
}

func groupPostgresSourceOptionsByTable(
	schemaOptions []*mgmtv1alpha1.PostgresSourceSchemaOption,
) map[string]*sourceTableOptions {
	groupedMappings := map[string]*sourceTableOptions{}

	for idx := range schemaOptions {
		schemaOpt := schemaOptions[idx]
		for tidx := range schemaOpt.Tables {
			tableOpt := schemaOpt.Tables[tidx]
			key := neosync_benthos.BuildBenthosTable(schemaOpt.Schema, tableOpt.Table)
			groupedMappings[key] = &sourceTableOptions{
				WhereClause: tableOpt.WhereClause,
			}
		}
	}

	return groupedMappings
}

func groupMysqlSourceOptionsByTable(
	schemaOptions []*mgmtv1alpha1.MysqlSourceSchemaOption,
) map[string]*sourceTableOptions {
	groupedMappings := map[string]*sourceTableOptions{}

	for idx := range schemaOptions {
		schemaOpt := schemaOptions[idx]
		for tidx := range schemaOpt.Tables {
			tableOpt := schemaOpt.Tables[tidx]
			key := neosync_benthos.BuildBenthosTable(schemaOpt.Schema, tableOpt.Table)
			groupedMappings[key] = &sourceTableOptions{
				WhereClause: tableOpt.WhereClause,
			}
		}
	}

	return groupedMappings
}

func groupMappingsByTable(
	mappings []*mgmtv1alpha1.JobMapping,
) []*TableMapping {
	groupedMappings := map[string][]*mgmtv1alpha1.JobMapping{}

	for _, mapping := range mappings {
		key := neosync_benthos.BuildBenthosTable(mapping.Schema, mapping.Table)
		groupedMappings[key] = append(groupedMappings[key], mapping)
	}

	output := make([]*TableMapping, 0, len(groupedMappings))
	for key, mappings := range groupedMappings {
		schema, table := splitTableKey(key)
		output = append(output, &TableMapping{
			Schema:   schema,
			Table:    table,
			Mappings: mappings,
		})
	}
	return output
}

type TableMapping struct {
	Schema   string
	Table    string
	Mappings []*mgmtv1alpha1.JobMapping
}

func getUniqueSchemasFromMappings(mappings []*mgmtv1alpha1.JobMapping) []string {
	schemas := map[string]struct{}{}
	for _, mapping := range mappings {
		schemas[mapping.Schema] = struct{}{}
	}

	output := make([]string, 0, len(schemas))

	for schema := range schemas {
		output = append(output, schema)
	}
	return output
}

func (a *Activities) getJobById(ctx context.Context, backendurl, jobId string) (*mgmtv1alpha1.Job, error) {
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		http.DefaultClient,
		backendurl,
	)

	getjobResp, err := jobclient.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: jobId,
	}))
	if err != nil {
		return nil, err
	}

	return getjobResp.Msg.Job, nil
}

func (a *Activities) getConnectionById(ctx context.Context, backendurl, connectionId string) (*mgmtv1alpha1.Connection, error) {
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(
		http.DefaultClient,
		backendurl,
	)

	getConnResp, err := connclient.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: connectionId,
	}))
	if err != nil {
		return nil, err
	}
	return getConnResp.Msg.Connection, nil
}

func getPgDsn(
	config *mgmtv1alpha1.PostgresConnectionConfig,
) (string, error) {
	switch cfg := config.ConnectionConfig.(type) {
	case *mgmtv1alpha1.PostgresConnectionConfig_Connection:
		dburl := fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s",
			cfg.Connection.User,
			cfg.Connection.Pass,
			cfg.Connection.Host,
			cfg.Connection.Port,
			cfg.Connection.Name,
		)
		if cfg.Connection.SslMode != nil && *cfg.Connection.SslMode != "" {
			dburl = fmt.Sprintf("%s?sslmode=%s", dburl, *cfg.Connection.SslMode)
		}
		return dburl, nil
	case *mgmtv1alpha1.PostgresConnectionConfig_Url:
		return cfg.Url, nil
	default:
		return "", fmt.Errorf("unsupported postgres connection config type")
	}
}

func getMysqlDsn(
	config *mgmtv1alpha1.MysqlConnectionConfig,
) (string, error) {
	switch cfg := config.ConnectionConfig.(type) {
	case *mgmtv1alpha1.MysqlConnectionConfig_Connection:
		dburl := fmt.Sprintf(
			"%s:%s@%s(%s:%d)/%s",
			cfg.Connection.User,
			cfg.Connection.Pass,
			cfg.Connection.Protocol,
			cfg.Connection.Host,
			cfg.Connection.Port,
			cfg.Connection.Name,
		)
		return dburl, nil
	case *mgmtv1alpha1.MysqlConnectionConfig_Url:
		return cfg.Url, nil
	default:
		return "", fmt.Errorf("unsupported mysql connection config type")
	}
}

func buildProcessorMutation(cols []*mgmtv1alpha1.JobMapping) (string, error) {
	pieces := []string{}

	for _, col := range cols {
		if col.Transformer.Value != "" && col.Transformer.Value != "passthrough" {
			mutation, err := computeMutationFunction(col)
			if err != nil {
				return "", fmt.Errorf("%s is not a supported transformer: %w", col.Transformer, err)
			}
			pieces = append(pieces, fmt.Sprintf("root.%s = %s", col.Column, mutation))
		}
	}
	return strings.Join(pieces, "\n"), nil
}

func buildPlainInsertArgs(cols []string) string {
	if len(cols) == 0 {
		return ""
	}
	pieces := make([]string, len(cols))
	for idx, col := range cols {
		pieces[idx] = fmt.Sprintf("this.%s", col)
	}
	return fmt.Sprintf("root = [%s]", strings.Join(pieces, ", "))
}

/*
method transformers
root.{destination_col} = this.{source_col}.transformermethod(args)

function transformers
root.{destination_col} = transformerfunction(args)
*/

func computeMutationFunction(col *mgmtv1alpha1.JobMapping) (string, error) {
	switch col.Transformer.Value {
	case "latitude":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "longitude":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "date":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "time_string":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "month_name":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "year_string":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "day_of_week":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "day_of_month":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "century":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "timezone":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "time_period":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "email":
		pd := col.Transformer.Config.GetEmailConfig().PreserveDomain
		pl := col.Transformer.Config.GetEmailConfig().PreserveLength
		return fmt.Sprintf("this.%s.emailtransformer(%t, %t)", col.Transformer.Value, pd, pl), nil
	case "mac_address":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "domain_name":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "url":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "username":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "ipv4":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "ipv6":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "password":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "jwt":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "word":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "sentence":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "paragraph":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "cc_type":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "cc_number":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "currency":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "amount_with_currency":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "title_male":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "title_female":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "first_name":
		pl := col.Transformer.Config.GetFirstNameConfig().PreserveLength
		return fmt.Sprintf("this.%s.firstnametransformer(%t)", col.Column, pl), nil
	case "first_name_female":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "last_name":
		pl := col.Transformer.Config.GetLastNameConfig().PreserveLength
		return fmt.Sprintf("this.%s.lastnametransformer(%t)", col.Column, pl), nil
	case "full_name":
		pl := col.Transformer.Config.GetFullNameConfig().PreserveLength
		return fmt.Sprintf("this.%s.fullnametransformer(%t)", col.Column, pl), nil
	case "chinese_first_name":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "chinese_last_name":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "chinese_name":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "phone_number":
		pl := col.Transformer.Config.GetPhoneNumberConfig().PreserveLength
		ef := col.Transformer.Config.GetPhoneNumberConfig().E164Format
		ih := col.Transformer.Config.GetPhoneNumberConfig().IncludeHyphens
		return fmt.Sprintf("this.%s.phonetransformer(%t, %t, %t)", col.Column, pl, ef, ih), nil
	case "int_phone_number":
		pl := col.Transformer.Config.GetIntPhoneNumberConfig().PreserveLength
		return fmt.Sprintf("this.%s.intphonetransformer(%t)", col.Column, pl), nil
	case "uuid":
		ih := col.Transformer.Config.GetUuidConfig().IncludeHyphen
		return fmt.Sprintf("this.%s.uuidtransformer(%t)", col.Column, ih), nil
	case "null":
		return "transformernull()", nil
	case "random_bool":
		return "randombooltransformer()", nil
	case "random_string":
		pl := col.Transformer.Config.GetRandomStringConfig().PreserveLength
		sl := col.Transformer.Config.GetRandomStringConfig().StrLength
		return fmt.Sprintf(`this.%s.randomstringtransformer(%t, %d)`, col.Column, pl, sl), nil
	case "random_int":
		pl := col.Transformer.Config.GetRandomIntConfig().PreserveLength
		sl := col.Transformer.Config.GetRandomIntConfig().IntLength
		return fmt.Sprintf(`this.%s.randominttransformer(%t, %d)`, col.Column, pl, sl), nil
	case "random_float":
		pl := col.Transformer.Config.GetRandomFloatConfig().PreserveLength
		bd := col.Transformer.Config.GetRandomFloatConfig().DigitsBeforeDecimal
		ad := col.Transformer.Config.GetRandomFloatConfig().DigitsAfterDecimal
		return fmt.Sprintf(`this.%s.randomfloattransformer(%t, %d, %d)`, col.Column, pl, bd, ad), nil
	case "gender":
		ab := col.Transformer.Config.GetGenderConfig().Abbreviate
		return fmt.Sprintf(`gendertransformer(%t)`, ab), nil
	case "utc_timestamp":
		return "utctimestamptransformer()", nil
	case "unix_timestamp":
		return "unixtimestamptransformer()", nil
	case "street_address":
		return "streetaddresstransformer()", nil
	case "city":
		return "citytransformer()", nil
	case "zipcode":
		return "zipcodetransformer()", nil
	case "state":
		return "statetransformer()", nil
	case "full_address":
		return "fulladdresstransformer()", nil
	default:
		return "", fmt.Errorf("unsupported transformer")
	}
}
