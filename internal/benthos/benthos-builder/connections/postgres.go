package benthosbuilder_connections

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/metrics"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_mssql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mssql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	querybuilder "github.com/nucleuscloud/neosync/worker/pkg/query-builder2"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

func NewPostgresBenthosBuilder(jobType bb_shared.JobType) (bb_shared.ConnectionBenthosBuilder, error) {
	switch jobType {
	case bb_shared.JobTypeSync:
		return NewPostgresSyncBuilder(), nil
	case bb_shared.JobTypeGenerate:
		return NewPostgresGenerateBuilder(), nil
	case bb_shared.JobTypeAIGenerate:
		return NewPostgresAIGenerateBuilder(), nil
	default:
		return nil, fmt.Errorf("unsupported postgres job type: %s", jobType)
	}
}

/*
	Sync
*/

type postgresSyncBuilder struct {
	// reverse of table dependency
	// map of foreign key to source table + column

	// when using these in building destination output if they don't exist they should be retrieved from destination
	primaryKeyToForeignKeysMap   map[string]map[string][]*bb_shared.ReferenceKey           // schema.table -> column -> ForeignKey
	colTransformerMap            map[string]map[string]*mgmtv1alpha1.JobMappingTransformer // schema.table -> column -> transformer
	sqlSourceSchemaColumnInfoMap map[string]map[string]*sqlmanager_shared.ColumnInfo       // schema.table -> column -> column info struct
}

func NewPostgresSyncBuilder() bb_shared.ConnectionBenthosBuilder {
	return &postgresSyncBuilder{}
}

func (b *postgresSyncBuilder) BuildSourceConfigs(ctx context.Context, params *bb_shared.SourceParams) ([]*bb_shared.BenthosSourceConfig, error) {
	sourceConnection := params.SourceConnection
	job := params.Job
	logger := params.Logger

	sqlSourceOpts, err := getPostgresJobSourceOpts(job.Source)
	if err != nil {
		return nil, err
	}
	var sourceTableOpts map[string]*sqlSourceTableOptions
	if sqlSourceOpts != nil {
		sourceTableOpts = groupSqlJobSourceOptionsByTable(sqlSourceOpts)
	}

	db, err := params.SqlManager.NewPooledSqlDb(ctx, logger, sourceConnection)
	if err != nil {
		return nil, fmt.Errorf("unable to create new sql db: %w", err)
	}
	defer db.Db.Close()

	groupedColumnInfo, err := db.Db.GetSchemaColumnMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get database schema for connection: %w", err)
	}
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
	filteredForeignKeysMap := filterForeignKeysMap(colTransformerMap, foreignKeysMap)

	tableSubsetMap := buildTableSubsetMap(sourceTableOpts, groupedTableMapping)
	tableColMap := getTableColMapFromMappings(groupedMappings)
	runConfigs, err := tabledependency.GetRunConfigs(filteredForeignKeysMap, tableSubsetMap, tableConstraints.PrimaryKeyConstraints, tableColMap)
	if err != nil {
		return nil, err
	}
	primaryKeyToForeignKeysMap := getPrimaryKeyDependencyMap(filteredForeignKeysMap)

	tableRunTypeQueryMap, err := querybuilder.BuildSelectQueryMap(db.Driver, filteredForeignKeysMap, runConfigs, sqlSourceOpts.SubsetByForeignKeyConstraints, groupedColumnInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to build select queries: %w", err)
	}

	configs := []*bb_shared.BenthosSourceConfig{}

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
		tableColTransformers := colTransformerMap[config.Table()]

		processorConfigs, err := buildProcessorConfigsByRunType(
			ctx,
			params.TransformerClient,
			config,
			columnForeignKeysMap,
			transformedFktoPkMap,
			job.Id,
			params.RunId,
			params.RedisConfig,
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

		columnDefaultProperties, err := getColumnDefaultProperties(logger, db.Driver, config.InsertColumns(), colInfoMap, tableColTransformers)
		if err != nil {
			return nil, err
		}

		configs = append(configs, &bb_shared.BenthosSourceConfig{
			Name:           fmt.Sprintf("%s.%s", config.Table(), config.RunType()),
			Config:         bc,
			DependsOn:      config.DependsOn(),
			RedisDependsOn: buildRedisDependsOnMap(transformedFktoPkMap, config),
			RunType:        config.RunType(),

			BenthosDsns: []*bb_shared.BenthosDsn{{ConnectionId: sourceConnection.Id, EnvVarKey: "SOURCE_CONNECTION_DSN"}},

			TableSchema:             mappings.Schema,
			TableName:               mappings.Table,
			Columns:                 config.InsertColumns(),
			ColumnDefaultProperties: columnDefaultProperties,
			PrimaryKeys:             config.PrimaryKeys(),

			Metriclabels: metrics.MetricLabels{
				metrics.NewEqLabel(metrics.TableSchemaLabel, mappings.Schema),
				metrics.NewEqLabel(metrics.TableNameLabel, mappings.Table),
				metrics.NewEqLabel(metrics.JobTypeLabel, "sync"),
			},
		})
	}

	return configs, nil
}

func getPostgresJobSourceOpts(
	source *mgmtv1alpha1.JobSource,
) (*sqlJobSourceOpts, error) {
	postgresSourceConfig := source.GetOptions().GetPostgres()
	if postgresSourceConfig == nil {
		return nil, fmt.Errorf("postgrest job source options missing")
	}
	schemaOpt := []*schemaOptions{}
	for _, opt := range postgresSourceConfig.GetSchemas() {
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
		HaltOnNewColumnAddition:       postgresSourceConfig.GetHaltOnNewColumnAddition(),
		SubsetByForeignKeyConstraints: postgresSourceConfig.GetSubsetByForeignKeyConstraints(),
		SchemaOpt:                     schemaOpt,
	}, nil
}

func (b *postgresSyncBuilder) BuildDestinationConfig(ctx context.Context, params *bb_shared.DestinationParams) (*bb_shared.BenthosDestinationConfig, error) {
	config := &bb_shared.BenthosDestinationConfig{}
	// this should not be here
	sqlSchemaColMap := getSqlSchemaColumnMap(ctx, params.DestinationOpts, params.DestConnection, b.sqlSourceSchemaColumnInfoMap, params.SqlManager, params.Logger)
	var colInfoMap map[string]*sqlmanager_shared.ColumnInfo
	colMap, ok := sqlSchemaColMap[neosync_benthos.BuildBenthosTable(params.SourceConfig.TableSchema, params.SourceConfig.TableName)]
	if ok {
		colInfoMap = colMap
	}
	benthosConfig := params.SourceConfig
	tableKey := neosync_benthos.BuildBenthosTable(benthosConfig.TableSchema, benthosConfig.TableName)
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
						Driver: sqlmanager_shared.PostgresDriver, // TODO
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
				if params.RedisConfig == nil {
					return nil, fmt.Errorf("missing redis config. this operation requires redis")
				}
				hashedKey := neosync_benthos.HashBenthosCacheKey(params.Job.GetId(), params.RunId, tableKey, col)
				config.Outputs = append(config.Outputs, neosync_benthos.Outputs{
					RedisHashOutput: &neosync_benthos.RedisHashOutputConfig{
						Url:            params.RedisConfig.Url,
						Key:            hashedKey,
						FieldsMapping:  fmt.Sprintf(`root = {meta(%q): json(%q)}`, hashPrimaryKeyMetaKey(benthosConfig.TableSchema, benthosConfig.TableName, col), col), // map of original value to transformed value
						WalkMetadata:   false,
						WalkJsonObject: false,
						Kind:           &params.RedisConfig.Kind,
						Master:         params.RedisConfig.Master,
						Tls:            shared.BuildBenthosRedisTlsConfig(params.RedisConfig),
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

		prefix, suffix := getInsertPrefixAndSuffix(sqlmanager_shared.PostgresDriver, benthosConfig.TableSchema, benthosConfig.TableName, benthosConfig.ColumnDefaultProperties)
		config.Outputs = append(config.Outputs, neosync_benthos.Outputs{
			Fallback: []neosync_benthos.Outputs{
				{
					PooledSqlInsert: &neosync_benthos.PooledSqlInsert{
						Driver: sqlmanager_shared.PostgresDriver,
						Dsn:    dsn,

						Schema:                   benthosConfig.TableSchema,
						Table:                    benthosConfig.TableName,
						Columns:                  benthosConfig.Columns,
						ColumnsDataTypes:         columnTypes,
						ColumnDefaultProperties:  benthosConfig.ColumnDefaultProperties,
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
	destination *mgmtv1alpha1.JobDestinationOptions,
	destinationConnection *mgmtv1alpha1.Connection,
	sourceSchemaColumnInfoMap map[string]map[string]*sqlmanager_shared.ColumnInfo,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	slogger *slog.Logger,
) map[string]map[string]*sqlmanager_shared.ColumnInfo {
	schemaColMap := sourceSchemaColumnInfoMap
	destOpts, err := shared.GetSqlJobDestinationOpts(destination)
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

func NewPostgresGenerateBuilder() bb_shared.ConnectionBenthosBuilder {
	return &postgresGenerateBuilder{}
}

func (b *postgresGenerateBuilder) BuildSourceConfigs(ctx context.Context, params *bb_shared.SourceParams) ([]*bb_shared.BenthosSourceConfig, error) {
	return []*bb_shared.BenthosSourceConfig{}, nil
}

func (b *postgresGenerateBuilder) BuildDestinationConfig(ctx context.Context, params *bb_shared.DestinationParams) (*bb_shared.BenthosDestinationConfig, error) {
	config := &bb_shared.BenthosDestinationConfig{}

	return config, nil
}

/*
	AI Generate
*/

type postgresAIGenerateBuilder struct {
}

func NewPostgresAIGenerateBuilder() bb_shared.ConnectionBenthosBuilder {
	return &postgresGenerateBuilder{}
}

func (b *postgresAIGenerateBuilder) BuildSourceConfigs(ctx context.Context, params *bb_shared.SourceParams) ([]*bb_shared.BenthosSourceConfig, error) {
	return []*bb_shared.BenthosSourceConfig{}, nil
}

func (b *postgresAIGenerateBuilder) BuildDestinationConfig(ctx context.Context, params *bb_shared.DestinationParams) (*bb_shared.BenthosDestinationConfig, error) {
	config := &bb_shared.BenthosDestinationConfig{}

	return config, nil
}
