package genbenthosconfigs_activity

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"gopkg.in/yaml.v3"
)

const (
	jobmappingSubsetErrMsg     = "job mappings are not equal to or a subset of the database schema found in the source connection"
	haltOnSchemaAdditionErrMsg = "job mappings does not contain a column mapping for all " +
		"columns found in the source connection for the selected schemas and tables"
)

type benthosBuilder struct {
	sqlmanagerclient sqlmanager.SqlManagerClient

	jobclient         mgmtv1alpha1connect.JobServiceClient
	connclient        mgmtv1alpha1connect.ConnectionServiceClient
	transformerclient mgmtv1alpha1connect.TransformersServiceClient

	jobId      string
	workflowId string
	runId      string

	redisConfig *shared.RedisConfig

	metricsEnabled bool
}

func newBenthosBuilder(
	sqlmanagerclient sqlmanager.SqlManagerClient,

	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,

	jobId, workflowId string, runId string,

	redisConfig *shared.RedisConfig,

	metricsEnabled bool,
) *benthosBuilder {
	return &benthosBuilder{
		sqlmanagerclient:  sqlmanagerclient,
		jobclient:         jobclient,
		connclient:        connclient,
		transformerclient: transformerclient,
		jobId:             jobId,
		workflowId:        workflowId,
		runId:             runId,
		redisConfig:       redisConfig,
		metricsEnabled:    metricsEnabled,
	}
}

type workflowMetadata struct {
	WorkflowId string
}

func (b *benthosBuilder) GenerateBenthosConfigs(
	ctx context.Context,
	req *GenerateBenthosConfigsRequest,
	wfmetadata *workflowMetadata,
	slogger *slog.Logger,
) (*GenerateBenthosConfigsResponse, error) {
	job, err := b.getJobById(ctx, req.JobId)
	if err != nil {
		return nil, fmt.Errorf("unable to get job by id: %w", err)
	}
	responses := []*BenthosConfigResponse{}

	// reverse of table dependency
	// map of foreign key to source table + column
	var primaryKeyToForeignKeysMap map[string]map[string][]*referenceKey                 // schema.table -> column -> ForeignKey
	var colTransformerMap map[string]map[string]*mgmtv1alpha1.JobMappingTransformer      // schema.table -> column -> transformer
	var sqlSourceSchemaColumnInfoMap map[string]map[string]*sqlmanager_shared.ColumnInfo // schema.table -> column -> column info struct
	var aiGroupedTableCols map[string][]string                                           // map of table key to columns for AI Generated schemas

	switch job.Source.Options.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_AiGenerate:
		sourceResponses, aimappings, err := b.getAiGenerateBenthosConfigResponses(ctx, job, slogger)
		if err != nil {
			return nil, fmt.Errorf("unable to build benthos AI Generate source config responses: %w", err)
		}
		aiGroupedTableCols = aimappings
		responses = append(responses, sourceResponses...)
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		sourceResponses, err := b.getGenerateBenthosConfigResponses(ctx, job)
		if err != nil {
			return nil, fmt.Errorf("unable to build benthos Generate source config responses: %w", err)
		}
		responses = append(responses, sourceResponses...)
	case *mgmtv1alpha1.JobSourceOptions_Postgres, *mgmtv1alpha1.JobSourceOptions_Mysql, *mgmtv1alpha1.JobSourceOptions_Mssql:
		resp, err := b.getSqlSyncBenthosConfigResponses(ctx, job, slogger)
		if err != nil {
			return nil, fmt.Errorf("unable to build benthos sql sync source config responses: %w", err)
		}
		primaryKeyToForeignKeysMap = resp.primaryKeyToForeignKeysMap
		colTransformerMap = resp.ColumnTransformerMap
		sqlSourceSchemaColumnInfoMap = resp.SchemaColumnInfoMap
		responses = append(responses, resp.BenthosConfigs...)
	case *mgmtv1alpha1.JobSourceOptions_Mongodb:
		resp, err := b.getMongoDbSyncBenthosConfigResponses(ctx, job, slogger)
		if err != nil {
			return nil, fmt.Errorf("unable to build benthos mongo sync source config responses: %w", err)
		}
		responses = append(responses, resp.BenthosConfigs...)
	case *mgmtv1alpha1.JobSourceOptions_Dynamodb:
		resp, err := b.getDynamoDbSyncBenthosConfigResponses(ctx, job, slogger)
		if err != nil {
			return nil, fmt.Errorf("unable to build benthos dynamodb sync source config responses: %w", err)
		}
		responses = append(responses, resp.BenthosConfigs...)
	default:
		return nil, fmt.Errorf("unsupported job source: %T", job.GetSource().GetOptions().GetConfig())
	}

	for destIdx, destination := range job.Destinations {
		destinationConnection, err := shared.GetConnectionById(ctx, b.connclient, destination.ConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get destination connection (%s) by id: %w", destination.ConnectionId, err)
		}
		sqlSchemaColMap := b.GetSqlSchemaColumnMap(ctx, destination, destinationConnection, sqlSourceSchemaColumnInfoMap, slogger)
		for _, resp := range responses {
			dstEnvVarKey := fmt.Sprintf("DESTINATION_%d_CONNECTION_DSN", destIdx)
			dsn := fmt.Sprintf("${%s}", dstEnvVarKey)

			switch connection := destinationConnection.ConnectionConfig.Config.(type) {
			case *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MysqlConfig, *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
				driver, err := getSqlDriverFromConnection(destinationConnection)
				if err != nil {
					return nil, err
				}
				resp.BenthosDsns = append(resp.BenthosDsns, &shared.BenthosDsn{EnvVarKey: dstEnvVarKey, ConnectionId: destinationConnection.Id})
				if isSyncConfig(resp.Config.Input) {
					// SQL sync output
					var colInfoMap map[string]*sqlmanager_shared.ColumnInfo
					colMap, ok := sqlSchemaColMap[neosync_benthos.BuildBenthosTable(resp.TableSchema, resp.TableName)]
					if ok {
						colInfoMap = colMap
					}

					outputs, err := b.getSqlSyncBenthosOutput(driver, destination, resp, dsn, primaryKeyToForeignKeysMap, colTransformerMap, colInfoMap)
					if err != nil {
						return nil, err
					}
					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, outputs...)
				} else if resp.Config.Input.Generate != nil {
					// SQL generate output
					outputs := b.getSqlGenerateOutput(driver, resp, destination, dsn)
					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, outputs...)
				} else if resp.Config.Input.OpenAiGenerate != nil {
					// SQL AI generate output
					outputs, err := b.getSqlAiGenerateOutput(driver, resp, destination, dsn, aiGroupedTableCols)
					if err != nil {
						return nil, err
					}
					resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, outputs...)
				} else {
					return nil, errors.New("unable to build destination connection due to unsupported source connection")
				}
			case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
				if resp.RunType == tabledependency.RunTypeUpdate {
					continue
				}
				outputs := b.getAwsS3SyncBenthosOutput(connection, resp, wfmetadata.WorkflowId)
				resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, outputs...)
			case *mgmtv1alpha1.ConnectionConfig_GcpCloudstorageConfig:
				if resp.RunType == tabledependency.RunTypeUpdate {
					continue
				}
				output := b.getGcpCloudStorageSyncBenthosOutput(connection, resp, wfmetadata.WorkflowId)
				resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, output...)
			case *mgmtv1alpha1.ConnectionConfig_MongoConfig:
				resp.BenthosDsns = append(resp.BenthosDsns, &shared.BenthosDsn{EnvVarKey: dstEnvVarKey, ConnectionId: destinationConnection.GetId()})
				if resp.Config.Input.PooledMongoDB != nil || resp.Config.Input.MongoDB != nil {
					resp.Config.Output.PooledMongoDB = &neosync_benthos.OutputMongoDb{
						Url: dsn,

						Database:   resp.TableSchema,
						Collection: resp.TableName,
						Operation:  "update-one",
						Upsert:     true,
						DocumentMap: `
						  root = {
								"$set": this
							}
						`,
						FilterMap: `
						  root._id = this._id
						`,
						WriteConcern: &neosync_benthos.MongoWriteConcern{
							W: "1",
						},
					}
				} else {
					return nil, errors.New("unable to build destination connection due to unsupported source connection")
				}
			case *mgmtv1alpha1.ConnectionConfig_DynamodbConfig:
				if resp.Config.Input.AwsDynamoDB == nil {
					return nil, errors.New("unable to build destination connection due to unsupported source connection for dynamodb")
				}
				dynamoDestinationOpts := destination.GetOptions().GetDynamodbOptions()
				if dynamoDestinationOpts == nil {
					return nil, errors.New("destination must have configured dyanmodb options")
				}
				tableMap := map[string]string{}
				for _, tm := range dynamoDestinationOpts.GetTableMappings() {
					tableMap[tm.GetSourceTable()] = tm.GetDestinationTable()
				}
				mappedTable, ok := tableMap[resp.TableName]
				if !ok {
					return nil, fmt.Errorf("did not find table map for %q when building dynamodb destination config", resp.TableName)
				}
				resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
					AwsDynamoDB: &neosync_benthos.OutputAwsDynamoDB{
						Table: mappedTable,
						JsonMapColumns: map[string]string{
							"": ".",
						},

						Batching: &neosync_benthos.Batching{
							// https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_BatchWriteItem.html
							// A single call to BatchWriteItem can transmit up to 16MB of data over the network, consisting of up to 25 item put or delete operations
							// Specifying the count here may not be enough if the overall data is above 16MB.
							// Benthos will fall back on error to single writes however
							Period: "5s",
							Count:  25,
						},

						Region:      connection.DynamodbConfig.GetRegion(),
						Endpoint:    connection.DynamodbConfig.GetEndpoint(),
						Credentials: buildBenthosS3Credentials(connection.DynamodbConfig.GetCredentials()),
					},
				})
			default:
				return nil, fmt.Errorf("unsupported destination connection config: %T", destinationConnection.GetConnectionConfig().GetConfig())
			}
		}
	}

	if b.metricsEnabled {
		labels := metrics.MetricLabels{
			metrics.NewEqLabel(metrics.AccountIdLabel, job.AccountId),
			metrics.NewEqLabel(metrics.JobIdLabel, job.Id),
			metrics.NewEqLabel(metrics.TemporalWorkflowId, withEnvInterpolation(metrics.TemporalWorkflowIdEnvKey)),
			metrics.NewEqLabel(metrics.TemporalRunId, withEnvInterpolation(metrics.TemporalRunIdEnvKey)),
			metrics.NewEqLabel(metrics.NeosyncDateLabel, withEnvInterpolation(metrics.NeosyncDateEnvKey)),
		}
		for _, resp := range responses {
			joinedLabels := append(labels, resp.metriclabels...) //nolint:gocritic
			resp.Config.Metrics = &neosync_benthos.Metrics{
				OtelCollector: &neosync_benthos.MetricsOtelCollector{},
				Mapping:       joinedLabels.ToBenthosMeta(),
			}
		}
	}

	var outputConfigs []*BenthosConfigResponse
	// hack to remove update configs when only syncing to s3
	if isOnlyBucketDestinations(job.Destinations) {
		for _, r := range responses {
			if r.RunType == tabledependency.RunTypeInsert {
				outputConfigs = append(outputConfigs, r)
			}
		}
	} else {
		outputConfigs = responses
	}

	outputConfigs, err = b.setRunContexts(ctx, outputConfigs, job.GetAccountId())
	if err != nil {
		return nil, fmt.Errorf("unable to set all run contexts for benthos configs: %w", err)
	}

	slogger.Info(fmt.Sprintf("successfully built %d benthos configs", len(outputConfigs)))
	return &GenerateBenthosConfigsResponse{
		BenthosConfigs: outputConfigs,
		AccountId:      job.GetAccountId(),
	}, nil
}

func withEnvInterpolation(input string) string {
	return fmt.Sprintf("${%s}", input)
}

// tries to get destination schema column info map
// if not uses source destination schema column info map
func (b *benthosBuilder) GetSqlSchemaColumnMap(
	ctx context.Context,
	destination *mgmtv1alpha1.JobDestination,
	destinationConnection *mgmtv1alpha1.Connection,
	sourceSchemaColumnInfoMap map[string]map[string]*sqlmanager_shared.ColumnInfo,
	slogger *slog.Logger,
) map[string]map[string]*sqlmanager_shared.ColumnInfo {
	schemaColMap := sourceSchemaColumnInfoMap
	destOpts, err := shared.GetSqlJobDestinationOpts(destination.GetOptions())
	if err != nil || destOpts.InitSchema {
		return schemaColMap
	}
	switch destinationConnection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MysqlConfig, *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		destDb, err := b.sqlmanagerclient.NewPooledSqlDb(ctx, slogger, destinationConnection)
		if err != nil {
			destDb.Db.Close()
			return schemaColMap
		}
		destColMap, err := destDb.Db.GetSchemaColumnMap(ctx)
		if err != nil {
			destDb.Db.Close()
			return schemaColMap
		}
		if len(destColMap) != 0 {
			schemaColMap = destColMap
		}
		destDb.Db.Close()
	}
	return schemaColMap
}

func isSyncConfig(input *neosync_benthos.InputConfig) bool {
	return input.SqlSelect != nil || input.PooledSqlRaw != nil
}

// this method modifies the input responses by nilling out the benthos config. it returns the same slice for convenience
func (b *benthosBuilder) setRunContexts(
	ctx context.Context,
	responses []*BenthosConfigResponse,
	accountId string,
) ([]*BenthosConfigResponse, error) {
	rcstream := b.jobclient.SetRunContexts(ctx)

	for _, config := range responses {
		bits, err := yaml.Marshal(config.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal benthos config: %w", err)
		}
		err = rcstream.Send(&mgmtv1alpha1.SetRunContextsRequest{
			Id: &mgmtv1alpha1.RunContextKey{
				JobRunId:   b.workflowId,
				ExternalId: shared.GetBenthosConfigExternalId(config.Name),
				AccountId:  accountId,
			},
			Value: bits,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send run context: %w", err)
		}
		config.Config = nil // nilling this out so that it does not persist in temporal
	}

	_, err := rcstream.CloseAndReceive()
	if err != nil {
		return nil, fmt.Errorf("unable to receive response from benthos runcontext request: %w", err)
	}
	return responses, nil
}

func isOnlyBucketDestinations(destinations []*mgmtv1alpha1.JobDestination) bool {
	for _, dest := range destinations {
		if dest.GetOptions().GetAwsS3Options() == nil && dest.GetOptions().GetGcpCloudstorageOptions() == nil {
			return false
		}
	}
	return true
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

func groupGenerateSourceOptionsByTable(
	schemaOptions []*mgmtv1alpha1.GenerateSourceSchemaOption,
) map[string]*generateSourceTableOptions {
	groupedMappings := map[string]*generateSourceTableOptions{}

	for idx := range schemaOptions {
		schemaOpt := schemaOptions[idx]
		for tidx := range schemaOpt.Tables {
			tableOpt := schemaOpt.Tables[tidx]
			key := neosync_benthos.BuildBenthosTable(schemaOpt.Schema, tableOpt.Table)
			groupedMappings[key] = &generateSourceTableOptions{
				Count: int(tableOpt.RowCount), // todo: probably need to update rowcount int64 to int32
			}
		}
	}

	return groupedMappings
}

func getSqlDriverFromConnection(conn *mgmtv1alpha1.Connection) (string, error) {
	switch conn.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return sqlmanager_shared.PostgresDriver, nil
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		return sqlmanager_shared.MysqlDriver, nil
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		return sqlmanager_shared.MssqlDriver, nil
	default:
		return "", fmt.Errorf("unsupported sql connection config")
	}
}

func groupSqlJobSourceOptionsByTable(
	sqlSourceOpts *sqlJobSourceOpts,
) map[string]*sqlSourceTableOptions {
	groupedMappings := map[string]*sqlSourceTableOptions{}
	for _, schemaOpt := range sqlSourceOpts.SchemaOpt {
		for tidx := range schemaOpt.Tables {
			tableOpt := schemaOpt.Tables[tidx]
			key := neosync_benthos.BuildBenthosTable(schemaOpt.Schema, tableOpt.Table)
			groupedMappings[key] = &sqlSourceTableOptions{
				WhereClause: tableOpt.WhereClause,
			}
		}
	}
	return groupedMappings
}

type tableMapping struct {
	Schema   string
	Table    string
	Mappings []*mgmtv1alpha1.JobMapping
}

func groupMappingsByTable(
	mappings []*mgmtv1alpha1.JobMapping,
) []*tableMapping {
	groupedMappings := map[string][]*mgmtv1alpha1.JobMapping{}

	for _, mapping := range mappings {
		key := neosync_benthos.BuildBenthosTable(mapping.Schema, mapping.Table)
		groupedMappings[key] = append(groupedMappings[key], mapping)
	}

	output := make([]*tableMapping, 0, len(groupedMappings))
	for key, mappings := range groupedMappings {
		schema, table := sqlmanager_shared.SplitTableKey(key)
		output = append(output, &tableMapping{
			Schema:   schema,
			Table:    table,
			Mappings: mappings,
		})
	}
	return output
}

func getTableMappingsMap(groupedMappings []*tableMapping) map[string]*tableMapping {
	groupedTableMapping := map[string]*tableMapping{}
	for _, tm := range groupedMappings {
		groupedTableMapping[neosync_benthos.BuildBenthosTable(tm.Schema, tm.Table)] = tm
	}
	return groupedTableMapping
}

func getColumnTransformerMap(tableMappingMap map[string]*tableMapping) map[string]map[string]*mgmtv1alpha1.JobMappingTransformer {
	colTransformerMap := map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{} // schema.table ->  column -> transformer
	for table, mapping := range tableMappingMap {
		colTransformerMap[table] = map[string]*mgmtv1alpha1.JobMappingTransformer{}
		for _, m := range mapping.Mappings {
			colTransformerMap[table][m.Column] = m.Transformer
		}
	}
	return colTransformerMap
}
