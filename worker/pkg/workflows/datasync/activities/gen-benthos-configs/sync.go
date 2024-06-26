package genbenthosconfigs_activity

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	querybuilder "github.com/nucleuscloud/neosync/worker/pkg/query-builder"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

type sqlSyncResp struct {
	BenthosConfigs             []*BenthosConfigResponse
	primaryKeyToForeignKeysMap map[string]map[string][]*referenceKey
	ColumnTransformerMap       map[string]map[string]*mgmtv1alpha1.JobMappingTransformer
}

func (b *benthosBuilder) getSqlSyncBenthosConfigResponses(
	ctx context.Context,
	job *mgmtv1alpha1.Job,
	slogger *slog.Logger,
) (*sqlSyncResp, error) {
	sourceConnection, err := shared.GetJobSourceConnection(ctx, job.GetSource(), b.connclient)
	if err != nil {
		return nil, fmt.Errorf("unable to get connection by id: %w", err)
	}

	sqlSourceOpts, err := getSqlJobSourceOpts(job.Source)
	if err != nil {
		return nil, err
	}
	var sourceTableOpts map[string]*sqlSourceTableOptions
	if sqlSourceOpts != nil {
		sourceTableOpts = groupJobSourceOptionsByTable(sqlSourceOpts)
	}

	db, err := b.sqlmanagerclient.NewPooledSqlDb(ctx, slogger, sourceConnection)
	if err != nil {
		return nil, fmt.Errorf("unable to create new sql db: %w", err)
	}
	defer db.Db.Close()

	groupedSchemas, err := db.Db.GetSchemaColumnMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get database schema for connection: %w", err)
	}
	if !areMappingsSubsetOfSchemas(groupedSchemas, job.Mappings) {
		return nil, errors.New(jobmappingSubsetErrMsg)
	}
	if sqlSourceOpts != nil && sqlSourceOpts.HaltOnNewColumnAddition &&
		shouldHaltOnSchemaAddition(groupedSchemas, job.Mappings) {
		return nil, errors.New(haltOnSchemaAdditionErrMsg)
	}
	uniqueSchemas := shared.GetUniqueSchemasFromMappings(job.Mappings)

	tableConstraints, err := db.Db.GetTableConstraintsBySchema(ctx, uniqueSchemas)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve database table constraints: %w", err)
	}

	foreignKeysMap, err := mergeVirtualForeignKeys(tableConstraints.ForeignKeyConstraints, job.GetVirtualForeignKeys(), groupedSchemas)
	if err != nil {
		return nil, err
	}

	slogger.Info(fmt.Sprintf("found %d foreign key constraints for database", getMapValuesCount(tableConstraints.ForeignKeyConstraints)))
	slogger.Info(fmt.Sprintf("found %d primary key constraints for database", getMapValuesCount(tableConstraints.PrimaryKeyConstraints)))

	groupedMappings := groupMappingsByTable(job.Mappings)
	groupedTableMapping := getTableMappingsMap(groupedMappings)
	colTransformerMap := getColumnTransformerMap(groupedTableMapping) // schema.table ->  column -> transformer

	tableSubsetMap := buildTableSubsetMap(sourceTableOpts)
	tableColMap := getTableColMapFromMappings(groupedMappings)
	runConfigs, err := tabledependency.GetRunConfigs(foreignKeysMap, tableSubsetMap, tableConstraints.PrimaryKeyConstraints, tableColMap)
	if err != nil {
		return nil, err
	}
	primaryKeyToForeignKeysMap := getPrimaryKeyDependencyMap(foreignKeysMap)

	// reverse of table dependency
	// map of foreign key to source table + column
	tableRunTypeQueryMap, err := querybuilder.BuildSelectQueryMap(db.Driver, foreignKeysMap, runConfigs, sqlSourceOpts.SubsetByForeignKeyConstraints, groupedSchemas)
	if err != nil {
		return nil, fmt.Errorf("unable to build select queries: %w", err)
	}

	sourceResponses, err := buildBenthosSqlSourceConfigResponses(ctx, b.transformerclient, groupedTableMapping, runConfigs, sourceConnection.Id, db.Driver, tableRunTypeQueryMap, groupedSchemas, foreignKeysMap, colTransformerMap, b.jobId, b.runId, b.redisConfig, primaryKeyToForeignKeysMap)
	if err != nil {
		return nil, fmt.Errorf("unable to build benthos sql source config responses: %w", err)
	}

	return &sqlSyncResp{
		BenthosConfigs:             sourceResponses,
		primaryKeyToForeignKeysMap: primaryKeyToForeignKeysMap,
		ColumnTransformerMap:       colTransformerMap,
	}, nil
}

func mergeVirtualForeignKeys(
	dbForeignKeys map[string][]*sqlmanager_shared.ForeignConstraint,
	virtualForeignKeys []*mgmtv1alpha1.VirtualForeignConstraint,
	colInfoMap map[string]map[string]*sqlmanager_shared.ColumnInfo,
) (map[string][]*sqlmanager_shared.ForeignConstraint, error) {
	fks := map[string][]*sqlmanager_shared.ForeignConstraint{}

	for table, fk := range dbForeignKeys {
		fks[table] = fk
	}

	for _, fk := range virtualForeignKeys {
		tn := sqlmanager_shared.BuildTable(fk.Schema, fk.Table)
		fkTable := sqlmanager_shared.BuildTable(fk.GetForeignKey().Schema, fk.GetForeignKey().Table)
		notNullable := []bool{}
		for _, c := range fk.GetColumns() {
			colMap, ok := colInfoMap[tn]
			if !ok {
				return nil, fmt.Errorf("virtual foreign key source table not found: %s", tn)
			}
			colInfo, ok := colMap[c]
			if !ok {
				return nil, fmt.Errorf("virtual foreign key source column not found: %s.%s", tn, c)
			}
			notNullable = append(notNullable, !colInfo.IsNullable)
		}
		fks[tn] = append(fks[tn], &sqlmanager_shared.ForeignConstraint{
			Columns:     fk.GetColumns(),
			NotNullable: notNullable,
			ForeignKey: &sqlmanager_shared.ForeignKey{
				Table:   fkTable,
				Columns: fk.GetForeignKey().GetColumns(),
			},
		})
	}

	return fks, nil
}

func buildBenthosSqlSourceConfigResponses(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	groupedTableMapping map[string]*tableMapping,
	runconfigs []*tabledependency.RunConfig,
	dsnConnectionId string,
	driver string,
	tableRunTypeQueryMap map[string]map[tabledependency.RunType]string,
	groupedColumnInfo map[string]map[string]*sqlmanager_shared.ColumnInfo,
	tableDependencies map[string][]*sqlmanager_shared.ForeignConstraint,
	colTransformerMap map[string]map[string]*mgmtv1alpha1.JobMappingTransformer,
	jobId, runId string,
	redisConfig *shared.RedisConfig,
	primaryKeyToForeignKeysMap map[string]map[string][]*referenceKey,
) ([]*BenthosConfigResponse, error) {
	responses := []*BenthosConfigResponse{}

	// map of table constraints that have transformers
	transformedForeignKeyToSourceMap := getTransformedFksMap(tableDependencies, colTransformerMap)

	for _, config := range runconfigs {
		mappings, ok := groupedTableMapping[config.Table]
		if !ok {
			return nil, fmt.Errorf("missing column mappings for table: %s", config.Table)
		}
		query, ok := tableRunTypeQueryMap[config.Table][config.RunType]
		if !ok {
			return nil, fmt.Errorf("select query not found for table: %s runType: %s", config.Table, config.RunType)
		}
		bc := &neosync_benthos.BenthosConfig{
			StreamConfig: neosync_benthos.StreamConfig{
				Input: &neosync_benthos.InputConfig{
					Inputs: neosync_benthos.Inputs{
						PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
							Driver: driver,
							Dsn:    "${SOURCE_CONNECTION_DSN}",

							Query: query,
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

		columnForeignKeysMap := primaryKeyToForeignKeysMap[config.Table]
		transformedFktoPkMap := transformedForeignKeyToSourceMap[config.Table]
		colTransformers := colTransformerMap[config.Table]
		colInfoMap := groupedColumnInfo[config.Table]

		processorConfigs, err := buildProcessorConfigsByRunType(
			ctx,
			transformerclient,
			config,
			columnForeignKeysMap,
			transformedFktoPkMap,
			jobId,
			runId,
			redisConfig,
			colTransformers,
			mappings.Mappings,
			colInfoMap,
		)
		if err != nil {
			return nil, err
		}
		for _, pc := range processorConfigs {
			bc.StreamConfig.Pipeline.Processors = append(bc.StreamConfig.Pipeline.Processors, *pc)
		}

		responses = append(responses, &BenthosConfigResponse{
			Name:           fmt.Sprintf("%s.%s", config.Table, config.RunType),
			Config:         bc,
			DependsOn:      config.DependsOn,
			RedisDependsOn: buildRedisDependsOnMap(transformedFktoPkMap),
			RunType:        config.RunType,

			BenthosDsns: []*shared.BenthosDsn{{ConnectionId: dsnConnectionId, EnvVarKey: "SOURCE_CONNECTION_DSN"}},

			TableSchema: mappings.Schema,
			TableName:   mappings.Table,
			Columns:     config.InsertColumns,
			primaryKeys: config.PrimaryKeys,

			metriclabels: metrics.MetricLabels{
				metrics.NewEqLabel(metrics.TableSchemaLabel, mappings.Schema),
				metrics.NewEqLabel(metrics.TableNameLabel, mappings.Table),
				metrics.NewEqLabel(metrics.JobTypeLabel, "sync"),
			},
		})
	}

	return responses, nil
}

func buildRedisDependsOnMap(transformedForeignKeyToSourceMap map[string][]*referenceKey) map[string][]string {
	redisDependsOnMap := map[string][]string{}
	for _, fks := range transformedForeignKeyToSourceMap {
		for _, fk := range fks {
			if _, exists := redisDependsOnMap[fk.Table]; !exists {
				redisDependsOnMap[fk.Table] = []string{}
			}
			redisDependsOnMap[fk.Table] = append(redisDependsOnMap[fk.Table], fk.Column)
		}
	}
	return redisDependsOnMap
}

func getTransformedFksMap(
	tabledependencies map[string][]*sqlmanager_shared.ForeignConstraint,
	colTransformerMap map[string]map[string]*mgmtv1alpha1.JobMappingTransformer,
) map[string]map[string][]*referenceKey {
	foreignKeyToSourceMap := buildForeignKeySourceMap(tabledependencies)
	// filter this list by table constraints that has transformer
	transformedForeignKeyToSourceMap := map[string]map[string][]*referenceKey{} // schema.table -> column -> foreignKey
	for table, constraints := range foreignKeyToSourceMap {
		_, ok := transformedForeignKeyToSourceMap[table]
		if !ok {
			transformedForeignKeyToSourceMap[table] = map[string][]*referenceKey{}
		}
		for col, tc := range constraints {
			// only add constraint if foreign key has transformer
			transformer, transformerOk := colTransformerMap[tc.Table][tc.Column]
			if transformerOk && shouldProcessStrict(transformer) {
				transformedForeignKeyToSourceMap[table][col] = append(transformedForeignKeyToSourceMap[table][col], tc)
			}
		}
	}
	return transformedForeignKeyToSourceMap
}

func buildProcessorConfigsByRunType(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	config *tabledependency.RunConfig,
	columnForeignKeysMap map[string][]*referenceKey,
	transformedFktoPkMap map[string][]*referenceKey,
	jobId, runId string,
	redisConfig *shared.RedisConfig,
	colTransformers map[string]*mgmtv1alpha1.JobMappingTransformer,
	mappings []*mgmtv1alpha1.JobMapping,
	columnInfoMap map[string]*sqlmanager_shared.ColumnInfo,
) ([]*neosync_benthos.ProcessorConfig, error) {
	if config.RunType == tabledependency.RunTypeUpdate {
		// sql update processor configs
		processorConfigs, err := buildSqlUpdateProcessorConfigs(config, redisConfig, jobId, runId, colTransformers, columnForeignKeysMap)
		if err != nil {
			return nil, err
		}
		return processorConfigs, nil
	} else {
		// sql insert processor configs
		fkSourceCols := []string{}
		for col := range columnForeignKeysMap {
			fkSourceCols = append(fkSourceCols, col)
		}
		processorConfigs, err := buildProcessorConfigs(ctx, transformerclient, mappings, columnInfoMap, transformedFktoPkMap, fkSourceCols, jobId, runId, redisConfig)
		if err != nil {
			return nil, err
		}
		return processorConfigs, nil
	}
}

func (b *benthosBuilder) getSqlSyncBenthosOutput(
	driver string,
	destination *mgmtv1alpha1.JobDestination,
	benthosConfig *BenthosConfigResponse,
	dsn string,
	primaryKeyToForeignKeysMap map[string]map[string][]*referenceKey,
	colTransformerMap map[string]map[string]*mgmtv1alpha1.JobMappingTransformer,
) ([]neosync_benthos.Outputs, error) {
	outputs := []neosync_benthos.Outputs{}
	tableKey := neosync_benthos.BuildBenthosTable(benthosConfig.TableSchema, benthosConfig.TableName)
	destOpts := getDestinationOptions(destination)
	if benthosConfig.RunType == tabledependency.RunTypeUpdate {
		args := benthosConfig.Columns
		args = append(args, benthosConfig.primaryKeys...)
		outputs = append(outputs, neosync_benthos.Outputs{
			Fallback: []neosync_benthos.Outputs{
				{
					PooledSqlUpdate: &neosync_benthos.PooledSqlUpdate{
						Driver: driver,
						Dsn:    dsn,

						Schema:       benthosConfig.TableSchema,
						Table:        benthosConfig.TableName,
						Columns:      benthosConfig.Columns,
						WhereColumns: benthosConfig.primaryKeys,
						ArgsMapping:  buildPlainInsertArgs(args),

						Batching: &neosync_benthos.Batching{
							Period: "5s",
							Count:  100,
						},
					},
				},
				// kills activity depending on error
				{Error: &neosync_benthos.ErrorOutputConfig{
					ErrorMsg: `${! meta("fallback_error")}`,
					Batching: &neosync_benthos.Batching{
						Period: "5s",
						Count:  100,
					},
				}},
			},
		})
	} else {
		// adds redis hash output for transformed primary keys
		constraints := primaryKeyToForeignKeysMap[tableKey]
		for col := range constraints {
			transformer := colTransformerMap[tableKey][col]
			if shouldProcessStrict(transformer) {
				if b.redisConfig == nil {
					return nil, fmt.Errorf("missing redis config. this operation requires redis")
				}
				hashedKey := neosync_benthos.HashBenthosCacheKey(b.jobId, b.runId, tableKey, col)
				outputs = append(outputs, neosync_benthos.Outputs{
					RedisHashOutput: &neosync_benthos.RedisHashOutputConfig{
						Url:            b.redisConfig.Url,
						Key:            hashedKey,
						FieldsMapping:  fmt.Sprintf(`root = {meta("neosync_%s_%s_%s"): json(%q)}`, benthosConfig.TableSchema, benthosConfig.TableName, col, col), // map of original value to transformed value
						WalkMetadata:   false,
						WalkJsonObject: false,
						Kind:           &b.redisConfig.Kind,
						Master:         b.redisConfig.Master,
						Tls:            shared.BuildBenthosRedisTlsConfig(b.redisConfig),
					},
				})
				benthosConfig.RedisConfig = append(benthosConfig.RedisConfig, &BenthosRedisConfig{
					Key:    hashedKey,
					Table:  tableKey,
					Column: col,
				})
			}
		}
		outputs = append(outputs, neosync_benthos.Outputs{
			Fallback: []neosync_benthos.Outputs{
				{
					PooledSqlInsert: &neosync_benthos.PooledSqlInsert{
						Driver: driver,
						Dsn:    dsn,

						Schema:              benthosConfig.TableSchema,
						Table:               benthosConfig.TableName,
						Columns:             benthosConfig.Columns,
						OnConflictDoNothing: destOpts.OnConflictDoNothing,
						TruncateOnRetry:     destOpts.Truncate,
						ArgsMapping:         buildPlainInsertArgs(benthosConfig.Columns),

						Batching: &neosync_benthos.Batching{
							Period: "5s",
							Count:  100,
						},
					},
				},
				// kills activity depending on error
				{Error: &neosync_benthos.ErrorOutputConfig{
					ErrorMsg: `${! meta("fallback_error")}`,
					Batching: &neosync_benthos.Batching{
						Period: "5s",
						Count:  100,
					},
				}},
			},
		})
	}

	return outputs, nil
}

func (b *benthosBuilder) getAwsS3SyncBenthosOutput(
	connection *mgmtv1alpha1.ConnectionConfig_AwsS3Config,
	benthosConfig *BenthosConfigResponse,
	workflowId string,
) []neosync_benthos.Outputs {
	outputs := []neosync_benthos.Outputs{}

	s3pathpieces := []string{}
	if connection.AwsS3Config.PathPrefix != nil && *connection.AwsS3Config.PathPrefix != "" {
		s3pathpieces = append(s3pathpieces, strings.Trim(*connection.AwsS3Config.PathPrefix, "/"))
	}

	s3pathpieces = append(
		s3pathpieces,
		"workflows",
		workflowId,
		"activities",
		neosync_benthos.BuildBenthosTable(benthosConfig.TableSchema, benthosConfig.TableName),
		"data",
		`${!count("files")}.txt.gz`,
	)

	outputs = append(outputs, neosync_benthos.Outputs{
		Fallback: []neosync_benthos.Outputs{
			{
				AwsS3: &neosync_benthos.AwsS3Insert{
					Bucket:      connection.AwsS3Config.Bucket,
					MaxInFlight: 64,
					Path:        strings.Join(s3pathpieces, "/"),
					Batching: &neosync_benthos.Batching{
						Count:  100,
						Period: "5s",
						Processors: []*neosync_benthos.BatchProcessor{
							{Archive: &neosync_benthos.ArchiveProcessor{Format: "lines"}},
							{Compress: &neosync_benthos.CompressProcessor{Algorithm: "gzip"}},
						},
					},
					Credentials: buildBenthosS3Credentials(connection.AwsS3Config.Credentials),
					Region:      connection.AwsS3Config.GetRegion(),
					Endpoint:    connection.AwsS3Config.GetEndpoint(),
				},
			},
			// kills activity depending on error
			{Error: &neosync_benthos.ErrorOutputConfig{
				ErrorMsg: `${! meta("fallback_error")}`,
				Batching: &neosync_benthos.Batching{
					Period: "5s",
					Count:  100,
				},
			}},
		},
	})
	return outputs
}

func (b *benthosBuilder) getGcpCloudStorageSyncBenthosOutput(
	connection *mgmtv1alpha1.ConnectionConfig_GcpCloudstorageConfig,
	benthosConfig *BenthosConfigResponse,
	workflowId string,
) []neosync_benthos.Outputs {
	outputs := []neosync_benthos.Outputs{}

	pathpieces := []string{}
	if connection.GcpCloudstorageConfig.GetPathPrefix() != "" {
		pathpieces = append(pathpieces, strings.Trim(connection.GcpCloudstorageConfig.GetPathPrefix(), "/"))
	}

	pathpieces = append(
		pathpieces,
		"workflows",
		workflowId,
		"activities",
		neosync_benthos.BuildBenthosTable(benthosConfig.TableSchema, benthosConfig.TableName),
		"data",
		`${!count("files")}.txt.gz`,
	)

	outputs = append(outputs, neosync_benthos.Outputs{
		Fallback: []neosync_benthos.Outputs{
			{
				GcpCloudStorage: &neosync_benthos.GcpCloudStorageOutput{
					Bucket:          connection.GcpCloudstorageConfig.GetBucket(),
					MaxInFlight:     64,
					Path:            strings.Join(pathpieces, "/"),
					ContentType:     shared.Ptr("txt/plain"),
					ContentEncoding: shared.Ptr("gzip"),
					Batching: &neosync_benthos.Batching{
						Count:  100,
						Period: "5s",
						Processors: []*neosync_benthos.BatchProcessor{
							{Archive: &neosync_benthos.ArchiveProcessor{Format: "lines"}},
							{Compress: &neosync_benthos.CompressProcessor{Algorithm: "gzip"}},
						},
					},
				},
			},
			// kills activity depending on error
			{Error: &neosync_benthos.ErrorOutputConfig{
				ErrorMsg: `${! meta("fallback_error")}`,
				Batching: &neosync_benthos.Batching{
					Period: "5s",
					Count:  100,
				},
			}},
		},
	})
	return outputs
}

func getTableColMapFromMappings(mappings []*tableMapping) map[string][]string {
	tableColMap := map[string][]string{}
	for _, m := range mappings {
		cols := []string{}
		for _, c := range m.Mappings {
			cols = append(cols, c.Column)
		}
		tn := sqlmanager_shared.BuildTable(m.Schema, m.Table)
		tableColMap[tn] = cols
	}
	return tableColMap
}

type referenceKey struct {
	Table  string
	Column string
}

// map of table primary key cols to foreign key cols
func getPrimaryKeyDependencyMap(tableDependencies map[string][]*sqlmanager_shared.ForeignConstraint) map[string]map[string][]*referenceKey {
	tc := map[string]map[string][]*referenceKey{} // schema.table -> column -> ForeignKey
	for table, constraints := range tableDependencies {
		for _, c := range constraints {
			_, ok := tc[c.ForeignKey.Table]
			if !ok {
				tc[c.ForeignKey.Table] = map[string][]*referenceKey{}
			}
			for idx, col := range c.ForeignKey.Columns {
				tc[c.ForeignKey.Table][col] = append(tc[c.ForeignKey.Table][col], &referenceKey{
					Table:  table,
					Column: c.Columns[idx],
				})
			}
		}
	}
	return tc
}

func findTopForeignKeySource(tableName, col string, tableDependencies map[string][]*sqlmanager_shared.ForeignConstraint) *referenceKey {
	// Add the foreign key dependencies of the current table
	if foreignKeys, ok := tableDependencies[tableName]; ok {
		for _, fk := range foreignKeys {
			for idx, c := range fk.Columns {
				if c == col {
					// Recursively add dependent tables and their foreign keys
					return findTopForeignKeySource(fk.ForeignKey.Table, fk.ForeignKey.Columns[idx], tableDependencies)
				}
			}
		}
	}
	return &referenceKey{
		Table:  tableName,
		Column: col,
	}
}

// builds schema.table -> FK column ->  PK schema table column
// find top level primary key column if foreign keys are nested
func buildForeignKeySourceMap(tableDeps map[string][]*sqlmanager_shared.ForeignConstraint) map[string]map[string]*referenceKey {
	outputMap := map[string]map[string]*referenceKey{}
	for tableName, constraints := range tableDeps {
		if _, ok := outputMap[tableName]; !ok {
			outputMap[tableName] = map[string]*referenceKey{}
		}
		for _, con := range constraints {
			for _, col := range con.Columns {
				fk := findTopForeignKeySource(tableName, col, tableDeps)
				outputMap[tableName][col] = fk
			}
		}
	}
	return outputMap
}

type destinationOptions struct {
	OnConflictDoNothing bool
	Truncate            bool
	TruncateCascade     bool
}

func getDestinationOptions(dest *mgmtv1alpha1.JobDestination) *destinationOptions {
	if dest == nil || dest.Options == nil || dest.Options.Config == nil {
		return &destinationOptions{}
	}
	switch config := dest.Options.Config.(type) {
	case *mgmtv1alpha1.JobDestinationOptions_PostgresOptions:
		return &destinationOptions{
			OnConflictDoNothing: config.PostgresOptions.GetOnConflict().GetDoNothing(),
			Truncate:            config.PostgresOptions.GetTruncateTable().GetTruncateBeforeInsert(),
			TruncateCascade:     config.PostgresOptions.GetTruncateTable().GetCascade(),
		}
	case *mgmtv1alpha1.JobDestinationOptions_MysqlOptions:
		return &destinationOptions{
			OnConflictDoNothing: config.MysqlOptions.GetOnConflict().GetDoNothing(),
			Truncate:            config.MysqlOptions.GetTruncateTable().GetTruncateBeforeInsert(),
		}
	default:
		return &destinationOptions{}
	}
}

type sqlJobSourceOpts struct {
	HaltOnNewColumnAddition       bool
	SubsetByForeignKeyConstraints bool
	SchemaOpt                     []*schemaOptions
}
type schemaOptions struct {
	Schema string
	Tables []*tableOptions
}
type tableOptions struct {
	Table       string
	WhereClause *string
}

func getSqlJobSourceOpts(
	source *mgmtv1alpha1.JobSource,
) (*sqlJobSourceOpts, error) {
	switch jobSourceConfig := source.GetOptions().GetConfig().(type) {
	case *mgmtv1alpha1.JobSourceOptions_Postgres:
		if jobSourceConfig.Postgres == nil {
			return nil, nil
		}
		schemaOpt := []*schemaOptions{}
		for _, opt := range jobSourceConfig.Postgres.Schemas {
			tableOpts := []*tableOptions{}
			for _, t := range opt.GetTables() {
				tableOpts = append(tableOpts, &tableOptions{
					Table:       t.Table,
					WhereClause: t.WhereClause,
				})
			}
			schemaOpt = append(schemaOpt, &schemaOptions{
				Schema: opt.GetSchema(),
				Tables: tableOpts,
			})
		}
		return &sqlJobSourceOpts{
			HaltOnNewColumnAddition:       jobSourceConfig.Postgres.HaltOnNewColumnAddition,
			SubsetByForeignKeyConstraints: jobSourceConfig.Postgres.SubsetByForeignKeyConstraints,
			SchemaOpt:                     schemaOpt,
		}, nil
	case *mgmtv1alpha1.JobSourceOptions_Mysql:
		if jobSourceConfig.Mysql == nil {
			return nil, nil
		}
		schemaOpt := []*schemaOptions{}
		for _, opt := range jobSourceConfig.Mysql.Schemas {
			tableOpts := []*tableOptions{}
			for _, t := range opt.GetTables() {
				tableOpts = append(tableOpts, &tableOptions{
					Table:       t.Table,
					WhereClause: t.WhereClause,
				})
			}
			schemaOpt = append(schemaOpt, &schemaOptions{
				Schema: opt.GetSchema(),
				Tables: tableOpts,
			})
		}
		return &sqlJobSourceOpts{
			HaltOnNewColumnAddition:       jobSourceConfig.Mysql.HaltOnNewColumnAddition,
			SubsetByForeignKeyConstraints: jobSourceConfig.Mysql.SubsetByForeignKeyConstraints,
			SchemaOpt:                     schemaOpt,
		}, nil
	default:
		return nil, errors.New("unsupported job source options type")
	}
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

func areMappingsSubsetOfSchemas(
	groupedSchemas map[string]map[string]*sqlmanager_shared.ColumnInfo,
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
	groupedSchemas map[string]map[string]*sqlmanager_shared.ColumnInfo,
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

type sqlSourceTableOptions struct {
	WhereClause *string
}

func buildTableSubsetMap(tableOpts map[string]*sqlSourceTableOptions) map[string]string {
	tableSubsetMap := map[string]string{}
	for table, opts := range tableOpts {
		if opts != nil && opts.WhereClause != nil && *opts.WhereClause != "" {
			tableSubsetMap[table] = *opts.WhereClause
		}
	}
	return tableSubsetMap
}
