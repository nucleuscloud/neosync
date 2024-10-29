package benthosbuilder_builders

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_mssql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mssql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	querybuilder "github.com/nucleuscloud/neosync/worker/pkg/query-builder2"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

/*
	Sync
*/

type sqlSyncBuilder struct {
	transformerclient mgmtv1alpha1connect.TransformersServiceClient
	sqlmanagerclient  sqlmanager.SqlManagerClient
	redisConfig       *shared.RedisConfig
	// reverse of table dependency
	// map of foreign key to source table + column

	// when using these in building destination output if they don't exist they should be retrieved from destination
	primaryKeyToForeignKeysMap        map[string]map[string][]*bb_internal.ReferenceKey         // schema.table -> column -> ForeignKey
	colTransformerMap                 map[string]map[string]*mgmtv1alpha1.JobMappingTransformer // schema.table -> column -> transformer
	sqlSourceSchemaColumnInfoMap      map[string]map[string]*sqlmanager_shared.ColumnInfo       // schema.table -> column -> column info struct
	sqlDestinationSchemaColumnInfoMap map[string]map[string]*sqlmanager_shared.ColumnInfo       // schema.table -> column -> column info struct
}

func NewSqlSyncBuilder(
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	redisConfig *shared.RedisConfig,
) bb_internal.ConnectionBenthosBuilder {
	return &sqlSyncBuilder{
		transformerclient: transformerclient,
		sqlmanagerclient:  sqlmanagerclient,
		redisConfig:       redisConfig,
	}
}

func (b *sqlSyncBuilder) BuildSourceConfigs(ctx context.Context, params *bb_internal.SourceParams) ([]*bb_internal.BenthosSourceConfig, error) {
	sourceConnection := params.SourceConnection
	job := params.Job
	logger := params.Logger

	sqlSourceOpts, err := getSqlJobSourceOpts(job.Source)
	if err != nil {
		return nil, err
	}
	var sourceTableOpts map[string]*sqlSourceTableOptions
	if sqlSourceOpts != nil {
		sourceTableOpts = groupSqlJobSourceOptionsByTable(sqlSourceOpts)
	}

	db, err := b.sqlmanagerclient.NewPooledSqlDb(ctx, logger, sourceConnection)
	if err != nil {
		return nil, fmt.Errorf("unable to create new sql db: %w", err)
	}
	defer db.Db.Close()

	groupedColumnInfo, err := db.Db.GetSchemaColumnMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get database schema for connection: %w", err)
	}
	b.sqlSourceSchemaColumnInfoMap = groupedColumnInfo
	if !areMappingsSubsetOfSchemas(groupedColumnInfo, job.Mappings) {
		return nil, errors.New(jobmappingSubsetErrMsg)
	}
	if sqlSourceOpts != nil && sqlSourceOpts.HaltOnNewColumnAddition &&
		shouldHaltOnSchemaAddition(groupedColumnInfo, job.Mappings) {
		return nil, errors.New(haltOnSchemaAdditionErrMsg)
	}
	uniqueSchemas := shared.GetUniqueSchemasFromMappings(job.Mappings)

	tableConstraints, err := db.Db.GetTableConstraintsBySchema(ctx, uniqueSchemas)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve database table constraints: %w", err)
	}

	foreignKeysMap, err := mergeVirtualForeignKeys(tableConstraints.ForeignKeyConstraints, job.GetVirtualForeignKeys(), groupedColumnInfo)
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("found %d foreign key constraints for database", getMapValuesCount(tableConstraints.ForeignKeyConstraints)))
	logger.Info(fmt.Sprintf("found %d primary key constraints for database", getMapValuesCount(tableConstraints.PrimaryKeyConstraints)))

	groupedMappings := groupMappingsByTable(job.Mappings)
	groupedTableMapping := getTableMappingsMap(groupedMappings)
	colTransformerMap := getColumnTransformerMap(groupedTableMapping) // schema.table ->  column -> transformer
	b.colTransformerMap = colTransformerMap
	filteredForeignKeysMap := filterForeignKeysMap(colTransformerMap, foreignKeysMap)

	tableSubsetMap := buildTableSubsetMap(sourceTableOpts, groupedTableMapping)
	tableColMap := getTableColMapFromMappings(groupedMappings)
	runConfigs, err := tabledependency.GetRunConfigs(filteredForeignKeysMap, tableSubsetMap, tableConstraints.PrimaryKeyConstraints, tableColMap)
	if err != nil {
		return nil, err
	}
	primaryKeyToForeignKeysMap := getPrimaryKeyDependencyMap(filteredForeignKeysMap)
	b.primaryKeyToForeignKeysMap = primaryKeyToForeignKeysMap

	tableRunTypeQueryMap, err := querybuilder.BuildSelectQueryMap(db.Driver, filteredForeignKeysMap, runConfigs, sqlSourceOpts.SubsetByForeignKeyConstraints, groupedColumnInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to build select queries: %w", err)
	}

	configs := []*bb_internal.BenthosSourceConfig{}

	// build benthos configs

	// map of table constraints that have transformers
	transformedForeignKeyToSourceMap := getTransformedFksMap(filteredForeignKeysMap, colTransformerMap)

	for _, config := range runConfigs {
		mappings, ok := groupedTableMapping[config.Table()]
		if !ok {
			return nil, fmt.Errorf("missing column mappings for table: %s", config.Table())
		}
		query, ok := tableRunTypeQueryMap[config.Table()][config.RunType()]
		if !ok {
			return nil, fmt.Errorf("select query not found for table: %s runType: %s", config.Table(), config.RunType())
		}
		bc := &neosync_benthos.BenthosConfig{
			StreamConfig: neosync_benthos.StreamConfig{
				Input: &neosync_benthos.InputConfig{
					Inputs: neosync_benthos.Inputs{
						PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
							Driver: db.Driver,
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

		columnForeignKeysMap := primaryKeyToForeignKeysMap[config.Table()]
		transformedFktoPkMap := transformedForeignKeyToSourceMap[config.Table()]
		colInfoMap := groupedColumnInfo[config.Table()]
		// tableColTransformers := colTransformerMap[config.Table()]

		processorConfigs, err := buildProcessorConfigsByRunType(
			ctx,
			b.transformerclient,
			config,
			columnForeignKeysMap,
			transformedFktoPkMap,
			job.Id,
			params.RunId,
			b.redisConfig,
			mappings.Mappings,
			colInfoMap,
			nil,
			[]string{},
		)
		if err != nil {
			return nil, err
		}
		for _, pc := range processorConfigs {
			bc.StreamConfig.Pipeline.Processors = append(bc.StreamConfig.Pipeline.Processors, *pc)
		}

		// columnDefaultProperties, err := getColumnDefaultProperties(logger, db.Driver, config.InsertColumns(), colInfoMap, tableColTransformers)
		// if err != nil {
		// 	return nil, err
		// }

		configs = append(configs, &bb_internal.BenthosSourceConfig{
			Name:           fmt.Sprintf("%s.%s", config.Table(), config.RunType()),
			Config:         bc,
			DependsOn:      config.DependsOn(),
			RedisDependsOn: buildRedisDependsOnMap(transformedFktoPkMap, config),
			RunType:        config.RunType(),

			BenthosDsns: []*bb_shared.BenthosDsn{{ConnectionId: sourceConnection.Id, EnvVarKey: "SOURCE_CONNECTION_DSN"}},

			TableSchema: mappings.Schema,
			TableName:   mappings.Table,
			Columns:     config.InsertColumns(),
			// ColumnDefaultProperties: columnDefaultProperties,
			PrimaryKeys: config.PrimaryKeys(),

			Metriclabels: metrics.MetricLabels{
				metrics.NewEqLabel(metrics.TableSchemaLabel, mappings.Schema),
				metrics.NewEqLabel(metrics.TableNameLabel, mappings.Table),
				metrics.NewEqLabel(metrics.JobTypeLabel, "sync"),
			},
		})
	}

	fmt.Println()
	fmt.Println("source sqlSourceSchemaColumnInfoMap", len(b.sqlSourceSchemaColumnInfoMap))
	fmt.Println("source sqlDestinationSchemaColumnInfoMap", len(b.sqlDestinationSchemaColumnInfoMap))
	fmt.Println()

	return configs, nil
}

func (b *sqlSyncBuilder) BuildDestinationConfig(ctx context.Context, params *bb_internal.DestinationParams) (*bb_internal.BenthosDestinationConfig, error) {
	fmt.Println()
	fmt.Println(params.SourceConfig.Name)
	fmt.Println("destination sqlSourceSchemaColumnInfoMap", len(b.sqlSourceSchemaColumnInfoMap))
	fmt.Println("destination sqlDestinationSchemaColumnInfoMap", len(b.sqlDestinationSchemaColumnInfoMap))
	fmt.Println()
	logger := params.Logger
	benthosConfig := params.SourceConfig
	tableKey := neosync_benthos.BuildBenthosTable(benthosConfig.TableSchema, benthosConfig.TableName)

	config := &bb_internal.BenthosDestinationConfig{}
	// should this be configured
	driver, err := getSqlDriverFromConnection(params.DestConnection)
	if err != nil {
		return nil, err
	}

	// this is very inefficient
	if len(b.sqlDestinationSchemaColumnInfoMap) == 0 {
		sqlSchemaColMap := getSqlSchemaColumnMap(ctx, params.DestinationOpts, params.DestConnection, b.sqlSourceSchemaColumnInfoMap, b.sqlmanagerclient, params.Logger)
		b.sqlDestinationSchemaColumnInfoMap = sqlSchemaColMap
	}

	var colInfoMap map[string]*sqlmanager_shared.ColumnInfo
	colMap, ok := b.sqlDestinationSchemaColumnInfoMap[tableKey]
	if ok {
		colInfoMap = colMap
	}

	colTransformerMap := b.colTransformerMap
	if len(colTransformerMap) == 0 {
		groupedMappings := groupMappingsByTable(params.Job.Mappings)
		groupedTableMapping := getTableMappingsMap(groupedMappings)
		colTMap := getColumnTransformerMap(groupedTableMapping) // schema.table ->  column -> transformer
		b.colTransformerMap = colTMap
		colTransformerMap = colTMap
	}

	// fmt.Println()
	// fmt.Println()
	// jsonF, _ := json.MarshalIndent(colInfoMap, "", " ")
	// fmt.Printf("%s \n", string(jsonF))

	tableColTransformers := colTransformerMap[tableKey]

	columnDefaultProperties, err := getColumnDefaultProperties(logger, driver, benthosConfig.Columns, colInfoMap, tableColTransformers)
	if err != nil {
		return nil, err
	}

	// fmt.Println(benthosConfig.Name)
	// jsonF, _ = json.MarshalIndent(columnDefaultProperties, "", " ")
	// fmt.Printf("columnDefaultProperties: %s \n", string(jsonF))
	// fmt.Println()
	// fmt.Println()

	destOpts := params.DestinationOpts
	dstEnvVarKey := fmt.Sprintf("DESTINATION_%d_CONNECTION_DSN", params.DestinationIdx)
	dsn := fmt.Sprintf("${%s}", dstEnvVarKey)
	config.BenthosDsns = append(config.BenthosDsns, &bb_shared.BenthosDsn{EnvVarKey: dstEnvVarKey, ConnectionId: params.DestConnection.Id})
	if benthosConfig.RunType == tabledependency.RunTypeUpdate {
		args := benthosConfig.Columns
		args = append(args, benthosConfig.PrimaryKeys...)
		config.Outputs = append(config.Outputs, neosync_benthos.Outputs{
			Fallback: []neosync_benthos.Outputs{
				{
					PooledSqlUpdate: &neosync_benthos.PooledSqlUpdate{
						Driver: driver, // TODO
						Dsn:    dsn,

						Schema:                   benthosConfig.TableSchema,
						Table:                    benthosConfig.TableName,
						Columns:                  benthosConfig.Columns,
						SkipForeignKeyViolations: destOpts.GetPostgresOptions().GetSkipForeignKeyViolations(),
						WhereColumns:             benthosConfig.PrimaryKeys,
						ArgsMapping:              buildPlainInsertArgs(args),

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
		constraints := b.primaryKeyToForeignKeysMap[tableKey]
		for col := range constraints {
			transformer := b.colTransformerMap[tableKey][col]
			if shouldProcessStrict(transformer) {
				if b.redisConfig == nil {
					return nil, fmt.Errorf("missing redis config. this operation requires redis")
				}
				hashedKey := neosync_benthos.HashBenthosCacheKey(params.Job.GetId(), params.RunId, tableKey, col)
				config.Outputs = append(config.Outputs, neosync_benthos.Outputs{
					RedisHashOutput: &neosync_benthos.RedisHashOutputConfig{
						Url:            b.redisConfig.Url,
						Key:            hashedKey,
						FieldsMapping:  fmt.Sprintf(`root = {meta(%q): json(%q)}`, hashPrimaryKeyMetaKey(benthosConfig.TableSchema, benthosConfig.TableName, col), col), // map of original value to transformed value
						WalkMetadata:   false,
						WalkJsonObject: false,
						Kind:           &b.redisConfig.Kind,
						Master:         b.redisConfig.Master,
						Tls:            shared.BuildBenthosRedisTlsConfig(b.redisConfig),
					},
				})
				benthosConfig.RedisConfig = append(benthosConfig.RedisConfig, &bb_shared.BenthosRedisConfig{
					Key:    hashedKey,
					Table:  tableKey,
					Column: col,
				})
			}
		}

		columnTypes := []string{}
		for _, c := range benthosConfig.Columns {
			colType, ok := colInfoMap[c]
			if ok {
				columnTypes = append(columnTypes, colType.DataType)
			} else {
				columnTypes = append(columnTypes, "")
			}
		}

		prefix, suffix := getInsertPrefixAndSuffix(driver, benthosConfig.TableSchema, benthosConfig.TableName, benthosConfig.ColumnDefaultProperties)
		config.Outputs = append(config.Outputs, neosync_benthos.Outputs{
			Fallback: []neosync_benthos.Outputs{
				{
					PooledSqlInsert: &neosync_benthos.PooledSqlInsert{
						Driver: driver,
						Dsn:    dsn,

						Schema:                   benthosConfig.TableSchema,
						Table:                    benthosConfig.TableName,
						Columns:                  benthosConfig.Columns,
						ColumnsDataTypes:         columnTypes,
						ColumnDefaultProperties:  columnDefaultProperties,
						OnConflictDoNothing:      destOpts.GetPostgresOptions().GetOnConflict().GetDoNothing(),
						SkipForeignKeyViolations: destOpts.GetPostgresOptions().GetSkipForeignKeyViolations(),
						TruncateOnRetry:          destOpts.GetPostgresOptions().GetTruncateTable().GetTruncateBeforeInsert(),
						ArgsMapping:              buildPlainInsertArgs(benthosConfig.Columns),
						Prefix:                   prefix,
						Suffix:                   suffix,

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

	return config, nil
}

func getInsertPrefixAndSuffix(
	driver, schema, table string,
	columnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties,
) (prefix, suffix *string) {
	var pre, suff *string
	if len(columnDefaultProperties) == 0 {
		return pre, suff
	}
	switch driver {
	case sqlmanager_shared.MssqlDriver:
		if hasPassthroughIdentityColumn(columnDefaultProperties) {
			enableIdentityInsert := true
			p := sqlmanager_mssql.BuildMssqlSetIdentityInsertStatement(schema, table, enableIdentityInsert)
			pre = &p
			s := sqlmanager_mssql.BuildMssqlSetIdentityInsertStatement(schema, table, !enableIdentityInsert)
			suff = &s
		}
		return pre, suff
	default:
		return pre, suff
	}
}

func hasPassthroughIdentityColumn(columnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties) bool {
	for _, d := range columnDefaultProperties {
		if d.NeedsOverride && d.NeedsReset && !d.HasDefaultTransformer {
			return true
		}
	}
	return false
}

// tries to get destination schema column info map
// if not uses source destination schema column info map
func getSqlSchemaColumnMap(
	ctx context.Context,
	destinationOpts *mgmtv1alpha1.JobDestinationOptions,
	destinationConnection *mgmtv1alpha1.Connection,
	sourceSchemaColumnInfoMap map[string]map[string]*sqlmanager_shared.ColumnInfo,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	slogger *slog.Logger,
) map[string]map[string]*sqlmanager_shared.ColumnInfo {
	schemaColMap := sourceSchemaColumnInfoMap
	destOpts, err := shared.GetSqlJobDestinationOpts(destinationOpts)
	if err != nil || destOpts.InitSchema {
		return schemaColMap
	}
	switch destinationConnection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MysqlConfig, *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		destDb, err := sqlmanagerclient.NewPooledSqlDb(ctx, slogger, destinationConnection)
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

/*
	Generate
*/

type postgresGenerateBuilder struct {
}

func NewPostgresGenerateBuilder() bb_internal.ConnectionBenthosBuilder {
	return &postgresGenerateBuilder{}
}

func (b *postgresGenerateBuilder) BuildSourceConfigs(ctx context.Context, params *bb_internal.SourceParams) ([]*bb_internal.BenthosSourceConfig, error) {
	return []*bb_internal.BenthosSourceConfig{}, nil
}

func (b *postgresGenerateBuilder) BuildDestinationConfig(ctx context.Context, params *bb_internal.DestinationParams) (*bb_internal.BenthosDestinationConfig, error) {
	config := &bb_internal.BenthosDestinationConfig{}

	return config, nil
}

/*
	AI Generate
*/

type postgresAIGenerateBuilder struct {
}

func NewPostgresAIGenerateBuilder() bb_internal.ConnectionBenthosBuilder {
	return &postgresGenerateBuilder{}
}

func (b *postgresAIGenerateBuilder) BuildSourceConfigs(ctx context.Context, params *bb_internal.SourceParams) ([]*bb_internal.BenthosSourceConfig, error) {
	return []*bb_internal.BenthosSourceConfig{}, nil
}

func (b *postgresAIGenerateBuilder) BuildDestinationConfig(ctx context.Context, params *bb_internal.DestinationParams) (*bb_internal.BenthosDestinationConfig, error) {
	config := &bb_internal.BenthosDestinationConfig{}

	return config, nil
}
