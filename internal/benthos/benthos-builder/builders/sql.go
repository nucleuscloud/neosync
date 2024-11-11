package benthosbuilder_builders

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

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
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

type sqlSyncBuilder struct {
	transformerclient  mgmtv1alpha1connect.TransformersServiceClient
	sqlmanagerclient   sqlmanager.SqlManagerClient
	redisConfig        *shared.RedisConfig
	driver             string
	selectQueryBuilder bb_shared.SelectQueryMapBuilder
	options            *SqlSyncOptions

	// reverse of table dependency
	// map of foreign key to source table + column
	primaryKeyToForeignKeysMap   map[string]map[string][]*bb_internal.ReferenceKey          // schema.table -> column -> ForeignKey
	colTransformerMap            map[string]map[string]*mgmtv1alpha1.JobMappingTransformer  // schema.table -> column -> transformer
	sqlSourceSchemaColumnInfoMap map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow // schema.table -> column -> column info struct
	// merged source and destination schema. with preference given to destination schema
	mergedSchemaColumnMap map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow // schema.table -> column -> column info struct
}

type SqlSyncOption func(*SqlSyncOptions)
type SqlSyncOptions struct {
	rawInsertMode bool
}

// WithRawInsertMode inserts data as is
func WithRawInsertMode() SqlSyncOption {
	return func(opts *SqlSyncOptions) {
		opts.rawInsertMode = true
	}
}

func NewSqlSyncBuilder(
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	redisConfig *shared.RedisConfig,
	databaseDriver string,
	selectQueryBuilder bb_shared.SelectQueryMapBuilder,
	opts ...SqlSyncOption,
) bb_internal.BenthosBuilder {
	options := &SqlSyncOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return &sqlSyncBuilder{
		transformerclient:  transformerclient,
		sqlmanagerclient:   sqlmanagerclient,
		redisConfig:        redisConfig,
		driver:             databaseDriver,
		selectQueryBuilder: selectQueryBuilder,
		options:            options,
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
	if sqlSourceOpts != nil && sqlSourceOpts.GenerateNewColumnTransformers {
		extraMappings, err := getAdditionalJobMappings(b.driver, groupedColumnInfo, job.Mappings, splitKeyToTablePieces, logger)
		if err != nil {
			return nil, err
		}
		logger.Debug(fmt.Sprintf("adding %d extra mappings due to unmapped columns", len(extraMappings)))
		job.Mappings = append(job.Mappings, extraMappings...)
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

	tableRunTypeQueryMap, err := b.selectQueryBuilder.BuildSelectQueryMap(db.Driver, filteredForeignKeysMap, runConfigs, sqlSourceOpts.SubsetByForeignKeyConstraints, groupedColumnInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to build select queries: %w", err)
	}

	configs, err := buildBenthosSqlSourceConfigResponses(logger, ctx, b.transformerclient, groupedTableMapping, runConfigs, sourceConnection.Id, db.Driver, tableRunTypeQueryMap, groupedColumnInfo, filteredForeignKeysMap, colTransformerMap, job.Id, params.RunId, b.redisConfig, primaryKeyToForeignKeysMap)
	if err != nil {
		return nil, fmt.Errorf("unable to build benthos sql source config responses: %w", err)
	}

	return configs, nil
}

func splitKeyToTablePieces(key string) (schema, table string, err error) {
	pieces := strings.SplitN(key, ".", 2)
	if len(pieces) != 2 {
		return "", "", errors.New("unable to split key to get schema and table, not 2 pieces")
	}
	return pieces[0], pieces[1], nil
}

func buildBenthosSqlSourceConfigResponses(
	slogger *slog.Logger,
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	groupedTableMapping map[string]*tableMapping,
	runconfigs []*tabledependency.RunConfig,
	dsnConnectionId string,
	driver string,
	tableRunTypeQueryMap map[string]map[tabledependency.RunType]string,
	groupedColumnInfo map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
	tableDependencies map[string][]*sqlmanager_shared.ForeignConstraint,
	colTransformerMap map[string]map[string]*mgmtv1alpha1.JobMappingTransformer,
	jobId, runId string,
	redisConfig *shared.RedisConfig,
	primaryKeyToForeignKeysMap map[string]map[string][]*bb_internal.ReferenceKey,
) ([]*bb_internal.BenthosSourceConfig, error) {
	configs := []*bb_internal.BenthosSourceConfig{}

	// map of table constraints that have transformers
	transformedForeignKeyToSourceMap := getTransformedFksMap(tableDependencies, colTransformerMap)

	for _, config := range runconfigs {
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

		columnForeignKeysMap := primaryKeyToForeignKeysMap[config.Table()]
		transformedFktoPkMap := transformedForeignKeyToSourceMap[config.Table()]
		colInfoMap := groupedColumnInfo[config.Table()]

		slogger.Debug("building processors")
		processorConfigs, err := buildProcessorConfigsByRunType(
			ctx,
			transformerclient,
			config,
			columnForeignKeysMap,
			transformedFktoPkMap,
			jobId,
			runId,
			redisConfig,
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

		configs = append(configs, &bb_internal.BenthosSourceConfig{
			Name:           fmt.Sprintf("%s.%s", config.Table(), config.RunType()),
			Config:         bc,
			DependsOn:      config.DependsOn(),
			RedisDependsOn: buildRedisDependsOnMap(transformedFktoPkMap, config),
			RunType:        config.RunType(),

			BenthosDsns: []*bb_shared.BenthosDsn{{ConnectionId: dsnConnectionId, EnvVarKey: "SOURCE_CONNECTION_DSN"}},

			TableSchema: mappings.Schema,
			TableName:   mappings.Table,
			Columns:     config.InsertColumns(),
			PrimaryKeys: config.PrimaryKeys(),

			Metriclabels: metrics.MetricLabels{
				metrics.NewEqLabel(metrics.TableSchemaLabel, mappings.Schema),
				metrics.NewEqLabel(metrics.TableNameLabel, mappings.Table),
				metrics.NewEqLabel(metrics.JobTypeLabel, "sync"),
			},
		})
	}
	return configs, nil
}

func (b *sqlSyncBuilder) BuildDestinationConfig(ctx context.Context, params *bb_internal.DestinationParams) (*bb_internal.BenthosDestinationConfig, error) {
	logger := params.Logger
	benthosConfig := params.SourceConfig
	tableKey := neosync_benthos.BuildBenthosTable(benthosConfig.TableSchema, benthosConfig.TableName)

	config := &bb_internal.BenthosDestinationConfig{}

	// lazy load
	if len(b.mergedSchemaColumnMap) == 0 {
		sqlSchemaColMap := getSqlSchemaColumnMap(ctx, params.DestConnection, b.sqlSourceSchemaColumnInfoMap, b.sqlmanagerclient, params.Logger)
		b.mergedSchemaColumnMap = sqlSchemaColMap
	}
	if len(b.mergedSchemaColumnMap) == 0 {
		return nil, fmt.Errorf("unable to retrieve schema columns for either source or destination: %s", params.DestConnection.Name)
	}

	var colInfoMap map[string]*sqlmanager_shared.DatabaseSchemaRow
	colMap, ok := b.mergedSchemaColumnMap[tableKey]
	if ok {
		colInfoMap = colMap
	}

	if len(colInfoMap) == 0 {
		return nil, fmt.Errorf("unable to retrieve schema columns for destination: %s table: %s", params.DestConnection.Name, tableKey)
	}

	colTransformerMap := b.colTransformerMap
	// lazy load
	if len(colTransformerMap) == 0 {
		groupedMappings := groupMappingsByTable(params.Job.Mappings)
		groupedTableMapping := getTableMappingsMap(groupedMappings)
		colTMap := getColumnTransformerMap(groupedTableMapping) // schema.table ->  column -> transformer
		b.colTransformerMap = colTMap
		colTransformerMap = colTMap
	}

	tableColTransformers := colTransformerMap[tableKey]
	if len(tableColTransformers) == 0 {
		return nil, fmt.Errorf("column transformer mappings not found for table: %s", tableKey)
	}

	columnDefaultProperties, err := getColumnDefaultProperties(logger, b.driver, benthosConfig.Columns, colInfoMap, tableColTransformers)
	if err != nil {
		return nil, err
	}
	params.SourceConfig.ColumnDefaultProperties = columnDefaultProperties

	destOpts, err := getDestinationOptions(params.DestinationOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to parse destination options: %w", err)
	}

	config.BenthosDsns = append(config.BenthosDsns, &bb_shared.BenthosDsn{EnvVarKey: params.DestEnvVarKey, ConnectionId: params.DestConnection.Id})
	if benthosConfig.RunType == tabledependency.RunTypeUpdate {
		args := benthosConfig.Columns
		args = append(args, benthosConfig.PrimaryKeys...)
		config.Outputs = append(config.Outputs, neosync_benthos.Outputs{
			Fallback: []neosync_benthos.Outputs{
				{
					PooledSqlUpdate: &neosync_benthos.PooledSqlUpdate{
						Driver: b.driver,
						Dsn:    params.DSN,

						Schema:                   benthosConfig.TableSchema,
						Table:                    benthosConfig.TableName,
						Columns:                  benthosConfig.Columns,
						SkipForeignKeyViolations: destOpts.SkipForeignKeyViolations,
						MaxInFlight:              int(destOpts.MaxInFlight),
						WhereColumns:             benthosConfig.PrimaryKeys,
						ArgsMapping:              buildPlainInsertArgs(args),

						Batching: &neosync_benthos.Batching{
							Period: destOpts.BatchPeriod,
							Count:  destOpts.BatchCount,
						},
					},
				},
				// kills activity depending on error
				{Error: &neosync_benthos.ErrorOutputConfig{
					ErrorMsg: `${! meta("fallback_error")}`,
					Batching: &neosync_benthos.Batching{
						Period: destOpts.BatchPeriod,
						Count:  destOpts.BatchCount,
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

		columnTypes := []string{} // use map going forward
		columnDataTypes := map[string]string{}
		for _, c := range benthosConfig.Columns {
			colType, ok := colInfoMap[c]
			if ok {
				columnDataTypes[c] = colType.DataType
				columnTypes = append(columnTypes, colType.DataType)
			} else {
				columnTypes = append(columnTypes, "")
			}
		}

		batchProcessors := []*neosync_benthos.BatchProcessor{}
		if benthosConfig.Config.Input.Inputs.NeosyncConnectionData != nil {
			batchProcessors = append(batchProcessors, &neosync_benthos.BatchProcessor{JsonToSql: &neosync_benthos.JsonToSqlConfig{ColumnDataTypes: columnDataTypes}})
		}

		prefix, suffix := getInsertPrefixAndSuffix(b.driver, benthosConfig.TableSchema, benthosConfig.TableName, columnDefaultProperties)
		config.Outputs = append(config.Outputs, neosync_benthos.Outputs{
			Fallback: []neosync_benthos.Outputs{
				{
					PooledSqlInsert: &neosync_benthos.PooledSqlInsert{
						Driver: b.driver,
						Dsn:    params.DSN,

						Schema:                   benthosConfig.TableSchema,
						Table:                    benthosConfig.TableName,
						Columns:                  benthosConfig.Columns,
						ColumnsDataTypes:         columnTypes,
						ColumnDefaultProperties:  columnDefaultProperties,
						OnConflictDoNothing:      destOpts.OnConflictDoNothing,
						SkipForeignKeyViolations: destOpts.SkipForeignKeyViolations,
						RawInsertMode:            b.options.rawInsertMode,
						TruncateOnRetry:          destOpts.Truncate,
						ArgsMapping:              buildPlainInsertArgs(benthosConfig.Columns),
						Prefix:                   prefix,
						Suffix:                   suffix,

						Batching: &neosync_benthos.Batching{
							Period:     destOpts.BatchPeriod,
							Count:      destOpts.BatchCount,
							Processors: batchProcessors,
						},
					},
				},
				// kills activity depending on error
				{Error: &neosync_benthos.ErrorOutputConfig{
					ErrorMsg: `${! meta("fallback_error")}`,
					Batching: &neosync_benthos.Batching{
						Period: destOpts.BatchPeriod,
						Count:  destOpts.BatchCount,
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
	destinationConnection *mgmtv1alpha1.Connection,
	sourceSchemaColumnInfoMap map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	slogger *slog.Logger,
) map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow {
	schemaColMap := sourceSchemaColumnInfoMap
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
			return mergeSourceDestinationColumnInfo(sourceSchemaColumnInfoMap, destColMap)
		}
		destDb.Db.Close()
	}
	return schemaColMap
}

// Merges source db column info with destination db col info
// Destination db col info take precedence
func mergeSourceDestinationColumnInfo(
	sourceCols map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
	destCols map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
) map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow {
	mergedCols := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{}

	for schemaTable, tableCols := range sourceCols {
		mergedCols[schemaTable] = tableCols
	}

	for schemaTable, tableCols := range destCols {
		if _, ok := mergedCols[schemaTable]; !ok {
			mergedCols[schemaTable] = make(map[string]*sqlmanager_shared.DatabaseSchemaRow)
		}
		for colName, colInfo := range tableCols {
			mergedCols[schemaTable][colName] = colInfo
		}
	}

	return mergedCols
}
