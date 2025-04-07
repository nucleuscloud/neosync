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
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	job_util "github.com/nucleuscloud/neosync/internal/job"
	rc "github.com/nucleuscloud/neosync/internal/runconfigs"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

type sqlSyncBuilder struct {
	transformerclient  mgmtv1alpha1connect.TransformersServiceClient
	sqlmanagerclient   sqlmanager.SqlManagerClient
	driver             string
	selectQueryBuilder bb_shared.SelectQueryMapBuilder
	pageLimit          int // default page limit for queries

	// reverse of table dependency
	// map of foreign key to source table + column
	primaryKeyToForeignKeysMap   map[string]map[string][]*bb_internal.ReferenceKey          // schema.table -> column -> ForeignKey
	colTransformerMap            map[string]map[string]*mgmtv1alpha1.JobMappingTransformer  // schema.table -> column -> transformer
	sqlSourceSchemaColumnInfoMap map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow // schema.table -> column -> column info struct
	// merged source and destination schema. with preference given to destination schema
	mergedSchemaColumnMap map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow // schema.table -> column -> column info struct
	configQueryMap        map[string]*sqlmanager_shared.SelectQuery                  // config id -> query info
	tableDeferrableMap    map[string]bool                                            // schema.table -> true if table has at least one deferrable constraint
}

func NewSqlSyncBuilder(
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	databaseDriver string,
	selectQueryBuilder bb_shared.SelectQueryMapBuilder,
	pageLimit int,
) bb_internal.BenthosBuilder {
	return &sqlSyncBuilder{
		transformerclient:  transformerclient,
		sqlmanagerclient:   sqlmanagerclient,
		driver:             databaseDriver,
		selectQueryBuilder: selectQueryBuilder,
		pageLimit:          pageLimit,
	}
}

func (b *sqlSyncBuilder) BuildSourceConfigs(
	ctx context.Context,
	params *bb_internal.SourceParams,
) ([]*bb_internal.BenthosSourceConfig, error) {
	sourceConnection := params.SourceConnection
	job := params.Job
	logger := params.Logger

	sqlSourceOpts, err := job_util.GetSqlJobSourceOpts(job.Source)
	if err != nil {
		return nil, err
	}
	var sourceTableOpts map[string]*sqlSourceTableOptions
	if sqlSourceOpts != nil {
		sourceTableOpts = groupSqlJobSourceOptionsByTable(sqlSourceOpts)
	}

	db, err := b.sqlmanagerclient.NewSqlConnection(
		ctx,
		connectionmanager.NewUniqueSession(connectionmanager.WithSessionGroup(params.JobRunId)),
		sourceConnection,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create new sql db: %w", err)
	}
	defer db.Db().Close()

	groupedColumnInfo, err := db.Db().GetSchemaColumnMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get database schema for connection: %w", err)
	}

	b.sqlSourceSchemaColumnInfoMap = groupedColumnInfo
	if sqlSourceOpts != nil && sqlSourceOpts.HaltOnNewColumnAddition {
		newColumns, shouldHalt := shouldHaltOnSchemaAddition(groupedColumnInfo, job.Mappings)
		if shouldHalt {
			return nil, fmt.Errorf(
				"%s: [%s]",
				haltOnSchemaAdditionErrMsg,
				strings.Join(newColumns, ", "),
			)
		}
	}

	if sqlSourceOpts != nil && sqlSourceOpts.HaltOnColumnRemoval {
		missing, shouldHalt := isSourceMissingColumnsFoundInMappings(
			groupedColumnInfo,
			job.Mappings,
		)
		if shouldHalt {
			return nil, fmt.Errorf(
				"%s: [%s]",
				haltOnSchemaAdditionErrMsg,
				strings.Join(missing, ", "),
			)
		}
	}

	// remove mappings that are not found in the source
	existingSourceMappings := removeMappingsNotFoundInSource(job.Mappings, groupedColumnInfo)

	if sqlSourceOpts != nil && sqlSourceOpts.PassthroughOnNewColumnAddition {
		extraMappings, err := getAdditionalPassthroughJobMappings(
			groupedColumnInfo,
			existingSourceMappings,
			splitKeyToTablePieces,
			logger,
		)
		if err != nil {
			return nil, err
		}
		logger.Debug(
			fmt.Sprintf("adding %d extra passthrough mappings due to unmapped columns", len(extraMappings)),
		)
		existingSourceMappings = append(existingSourceMappings, extraMappings...)
	}

	if sqlSourceOpts != nil && sqlSourceOpts.GenerateNewColumnTransformers {
		extraMappings, err := getAdditionalJobMappings(
			b.driver,
			groupedColumnInfo,
			existingSourceMappings,
			splitKeyToTablePieces,
			logger,
		)
		if err != nil {
			return nil, err
		}
		logger.Debug(
			fmt.Sprintf("adding %d extra mappings due to unmapped columns", len(extraMappings)),
		)
		existingSourceMappings = append(existingSourceMappings, extraMappings...)
	}
	uniqueSchemas := shared.GetUniqueSchemasFromMappings(existingSourceMappings)

	schemaTablesMap := shared.GetSchemaTablesMapFromMappings(existingSourceMappings)
	tableDeferrableMap, err := getTableDeferrableMap(ctx, db, sourceConnection, schemaTablesMap)
	if err != nil {
		return nil, fmt.Errorf("unable to get table deferrable map: %w", err)
	}
	b.tableDeferrableMap = tableDeferrableMap

	tableConstraints, err := db.Db().GetTableConstraintsBySchema(ctx, uniqueSchemas)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve database table constraints: %w", err)
	}

	foreignKeysMap, err := mergeVirtualForeignKeys(
		tableConstraints.ForeignKeyConstraints,
		job.GetVirtualForeignKeys(),
		groupedColumnInfo,
	)
	if err != nil {
		return nil, err
	}

	logger.Info(
		fmt.Sprintf(
			"found %d foreign key constraints for database",
			getMapValuesCount(tableConstraints.ForeignKeyConstraints),
		),
	)
	logger.Info(
		fmt.Sprintf(
			"found %d primary key constraints for database",
			getMapValuesCount(tableConstraints.PrimaryKeyConstraints),
		),
	)

	groupedMappings := groupMappingsByTable(existingSourceMappings)
	groupedTableMapping := getTableMappingsMap(groupedMappings)
	colTransformerMap := getColumnTransformerMap(
		groupedTableMapping,
	) // schema.table ->  column -> transformer
	b.colTransformerMap = colTransformerMap
	// include virtual foreign keys and removes fks that have null transformers
	filteredForeignKeysMap := filterForeignKeysMap(colTransformerMap, foreignKeysMap)

	tableSubsetMap := buildTableSubsetMap(sourceTableOpts, groupedTableMapping)
	tableColMap := getTableColMapFromMappings(groupedMappings)
	runConfigs, err := rc.BuildRunConfigs(
		filteredForeignKeysMap,
		tableSubsetMap,
		tableConstraints.PrimaryKeyConstraints,
		tableColMap,
		tableConstraints.UniqueIndexes,
		tableConstraints.UniqueConstraints,
	)
	if err != nil {
		return nil, err
	}

	primaryKeyToForeignKeysMap := getPrimaryKeyDependencyMap(filteredForeignKeysMap)
	b.primaryKeyToForeignKeysMap = primaryKeyToForeignKeysMap

	configQueryMap, err := b.selectQueryBuilder.BuildSelectQueryMap(
		db.Driver(),
		runConfigs,
		sqlSourceOpts.SubsetByForeignKeyConstraints,
		b.pageLimit,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to build select queries: %w", err)
	}
	b.configQueryMap = configQueryMap

	configs, err := buildBenthosSqlSourceConfigResponses(
		logger,
		ctx,
		b.transformerclient,
		groupedTableMapping,
		runConfigs,
		sourceConnection.Id,
		configQueryMap,
		groupedColumnInfo,
		filteredForeignKeysMap,
		colTransformerMap,
		job.Id,
		params.JobRunId,
		primaryKeyToForeignKeysMap,
	)
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

// TODO: remove tableDependencies and use runconfig's foreign keys
func buildBenthosSqlSourceConfigResponses(
	slogger *slog.Logger,
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	groupedTableMapping map[string]*tableMapping,
	runconfigs []*rc.RunConfig,
	dsnConnectionId string,
	configQueryMap map[string]*sqlmanager_shared.SelectQuery,
	groupedColumnInfo map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
	tableDependencies map[string][]*sqlmanager_shared.ForeignConstraint,
	colTransformerMap map[string]map[string]*mgmtv1alpha1.JobMappingTransformer,
	jobId, runId string,
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
		query, ok := configQueryMap[config.Id()]
		if !ok {
			return nil, fmt.Errorf("query info not found for id: %s", config.Id())
		}

		bc := &neosync_benthos.BenthosConfig{
			StreamConfig: neosync_benthos.StreamConfig{
				Input: &neosync_benthos.InputConfig{
					Inputs: neosync_benthos.Inputs{
						PooledSqlRaw: &neosync_benthos.InputPooledSqlRaw{
							ConnectionId: dsnConnectionId,

							Query:             query.Query,
							PagedQuery:        query.PageQuery,
							OrderByColumns:    config.OrderByColumns(),
							ExpectedTotalRows: &query.PageLimit,
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
			mappings.Mappings,
			colInfoMap,
			nil,
			[]string{},
		)
		if err != nil {
			return nil, err
		}
		for _, pc := range processorConfigs {
			bc.Pipeline.Processors = append(bc.Pipeline.Processors, *pc)
		}

		cursors, err := buildIdentityCursors(ctx, transformerclient, mappings.Mappings)
		if err != nil {
			return nil, fmt.Errorf("unable to build identity cursors: %w", err)
		}

		configs = append(configs, &bb_internal.BenthosSourceConfig{
			Name:      config.Id(),
			Config:    bc,
			DependsOn: config.DependsOn(),
			RunType:   config.RunType(),

			BenthosDsns: []*bb_shared.BenthosDsn{{ConnectionId: dsnConnectionId}},

			TableSchema: mappings.Schema,
			TableName:   mappings.Table,
			Columns:     config.InsertColumns(),
			PrimaryKeys: config.PrimaryKeys(),

			ColumnIdentityCursors: cursors,

			Metriclabels: metrics.MetricLabels{
				metrics.NewEqLabel(metrics.TableSchemaLabel, mappings.Schema),
				metrics.NewEqLabel(metrics.TableNameLabel, mappings.Table),
				metrics.NewEqLabel(metrics.JobTypeLabel, "sync"),
			},
		})
	}
	return configs, nil
}

func (b *sqlSyncBuilder) BuildDestinationConfig(
	ctx context.Context,
	params *bb_internal.DestinationParams,
) (*bb_internal.BenthosDestinationConfig, error) {
	logger := params.Logger
	benthosConfig := params.SourceConfig
	tableKey := neosync_benthos.BuildBenthosTable(
		benthosConfig.TableSchema,
		benthosConfig.TableName,
	)

	config := &bb_internal.BenthosDestinationConfig{}

	// lazy load
	if len(b.mergedSchemaColumnMap) == 0 {
		sqlSchemaColMap := getSqlSchemaColumnMap(
			ctx,
			connectionmanager.NewUniqueSession(connectionmanager.WithSessionGroup(params.JobRunId)),
			params.DestConnection,
			b.sqlSourceSchemaColumnInfoMap,
			b.sqlmanagerclient,
			params.Logger,
		)
		b.mergedSchemaColumnMap = sqlSchemaColMap
	}
	if len(b.mergedSchemaColumnMap) == 0 {
		return nil, fmt.Errorf(
			"unable to retrieve schema columns for either source or destination: %s",
			params.DestConnection.Name,
		)
	}

	var colInfoMap map[string]*sqlmanager_shared.DatabaseSchemaRow
	colMap, ok := b.mergedSchemaColumnMap[tableKey]
	if ok {
		colInfoMap = colMap
	}

	if len(colInfoMap) == 0 {
		return nil, fmt.Errorf(
			"unable to retrieve schema columns for destination: %s table: %s",
			params.DestConnection.Name,
			tableKey,
		)
	}

	colTransformerMap := b.colTransformerMap
	// lazy load
	if len(colTransformerMap) == 0 {
		groupedMappings := groupMappingsByTable(params.Job.Mappings)
		groupedTableMapping := getTableMappingsMap(groupedMappings)
		colTMap := getColumnTransformerMap(
			groupedTableMapping,
		) // schema.table ->  column -> transformer
		b.colTransformerMap = colTMap
		colTransformerMap = colTMap
	}

	tableColTransformers := colTransformerMap[tableKey]
	if len(tableColTransformers) == 0 {
		return nil, fmt.Errorf("column transformer mappings not found for table: %s", tableKey)
	}

	columnDefaultProperties, err := getColumnDefaultProperties(
		logger,
		b.driver,
		benthosConfig.Columns,
		colInfoMap,
		tableColTransformers,
	)
	if err != nil {
		return nil, err
	}
	params.SourceConfig.ColumnDefaultProperties = columnDefaultProperties

	destOpts, err := getDestinationOptions(params.DestinationOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to parse destination options: %w", err)
	}

	columnUpdatesDisallowed := []string{}
	for _, col := range colInfoMap {
		if !col.UpdateAllowed {
			columnUpdatesDisallowed = append(columnUpdatesDisallowed, col.ColumnName)
		}
	}

	// this will be nil if coming from CLI sync
	query := b.configQueryMap[benthosConfig.Name]

	// skip foreign key violations if the query could return rows that violate foreign key constraints
	skipForeignKeyViolations := destOpts.SkipForeignKeyViolations ||
		(query != nil && query.IsNotForeignKeySafeSubset)

	config.BenthosDsns = append(
		config.BenthosDsns,
		&bb_shared.BenthosDsn{ConnectionId: params.DestConnection.Id},
	)
	if benthosConfig.RunType == rc.RunTypeUpdate {
		processorColumns := benthosConfig.Columns
		processorColumns = append(processorColumns, benthosConfig.PrimaryKeys...)
		sqlProcessor, err := getProcessors(
			b.driver,
			processorColumns,
			colInfoMap,
			columnDefaultProperties,
		)
		if err != nil {
			return nil, err
		}
		config.Outputs = append(config.Outputs, neosync_benthos.Outputs{
			Fallback: []neosync_benthos.Outputs{
				{
					PooledSqlUpdate: &neosync_benthos.PooledSqlUpdate{
						ConnectionId: params.DestConnection.GetId(),

						Schema:                   benthosConfig.TableSchema,
						Table:                    benthosConfig.TableName,
						Columns:                  benthosConfig.Columns,
						SkipForeignKeyViolations: skipForeignKeyViolations,
						MaxInFlight:              int(destOpts.MaxInFlight),
						WhereColumns:             benthosConfig.PrimaryKeys,

						Batching: &neosync_benthos.Batching{
							Period:     destOpts.BatchPeriod,
							Count:      destOpts.BatchCount,
							Processors: []*neosync_benthos.BatchProcessor{sqlProcessor},
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
				hashedKey := neosync_benthos.HashBenthosCacheKey(params.Job.GetId(), params.JobRunId, tableKey, col)
				config.Outputs = append(config.Outputs, neosync_benthos.Outputs{
					Fallback: []neosync_benthos.Outputs{
						{
							RedisHashOutput: &neosync_benthos.RedisHashOutputConfig{
								Key:            hashedKey,
								FieldsMapping:  fmt.Sprintf(`root = {meta(%q): json(%q)}`, hashPrimaryKeyMetaKey(benthosConfig.TableSchema, benthosConfig.TableName, col), col), // map of original value to transformed value
								WalkMetadata:   false,
								WalkJsonObject: false,
							},
						},
						// kills activity depending on error
						{Error: &neosync_benthos.ErrorOutputConfig{
							ErrorMsg: `${! meta("fallback_error")}`,
						}},
					},
				})
				benthosConfig.RedisConfig = append(benthosConfig.RedisConfig, &bb_shared.BenthosRedisConfig{
					Key:    hashedKey,
					Table:  tableKey,
					Column: col,
				})
			}
		}
		sqlProcessor, err := getProcessors(b.driver, benthosConfig.Columns, colInfoMap, columnDefaultProperties)
		if err != nil {
			return nil, err
		}

		hasDeferrableConstraint := b.tableDeferrableMap[tableKey]
		prefix, suffix := getInsertPrefixAndSuffix(b.driver, benthosConfig.TableSchema, benthosConfig.TableName, columnDefaultProperties)
		config.Outputs = append(config.Outputs, neosync_benthos.Outputs{
			Fallback: []neosync_benthos.Outputs{
				{
					PooledSqlInsert: &neosync_benthos.PooledSqlInsert{
						ConnectionId: params.DestConnection.GetId(),

						Schema:                      benthosConfig.TableSchema,
						Table:                       benthosConfig.TableName,
						PrimaryKeyColumns:           benthosConfig.PrimaryKeys,
						ColumnUpdatesDisallowed:     columnUpdatesDisallowed,
						OnConflictDoNothing:         destOpts.OnConflictDoNothing,
						OnConflictDoUpdate:          destOpts.OnConflictDoUpdate,
						HasDeferrableConstraint:     hasDeferrableConstraint, // postgres only
						SkipForeignKeyViolations:    skipForeignKeyViolations,
						ShouldOverrideColumnDefault: shouldOverrideColumnDefault(columnDefaultProperties),
						TruncateOnRetry:             destOpts.Truncate,
						Prefix:                      prefix,
						Suffix:                      suffix,

						Batching: &neosync_benthos.Batching{
							Period:     destOpts.BatchPeriod,
							Count:      destOpts.BatchCount,
							Processors: []*neosync_benthos.BatchProcessor{sqlProcessor},
						},
						MaxInFlight: int(destOpts.MaxInFlight),
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

func getProcessors(
	driver string,
	columns []string,
	colInfoMap map[string]*sqlmanager_shared.DatabaseSchemaRow,
	columnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties,
) (*neosync_benthos.BatchProcessor, error) {
	columnDataTypes := map[string]string{}
	for _, c := range columns {
		colType, ok := colInfoMap[c]
		if ok {
			columnDataTypes[c] = colType.DataType
		}
	}

	return getSqlBatchProcessors(driver, columns, columnDataTypes, columnDefaultProperties)
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
			p := sqlmanager_mssql.BuildMssqlSetIdentityInsertStatement(
				schema,
				table,
				enableIdentityInsert,
			)
			pre = &p
			s := sqlmanager_mssql.BuildMssqlSetIdentityInsertStatement(
				schema,
				table,
				!enableIdentityInsert,
			)
			suff = &s
		}
		return pre, suff
	default:
		return pre, suff
	}
}

func hasPassthroughIdentityColumn(
	columnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties,
) bool {
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
	session connectionmanager.SessionInterface,
	destinationConnection *mgmtv1alpha1.Connection,
	sourceSchemaColumnInfoMap map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	slogger *slog.Logger,
) map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow {
	schemaColMap := sourceSchemaColumnInfoMap
	switch destinationConnection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MysqlConfig, *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		destDb, err := sqlmanagerclient.NewSqlConnection(ctx, session, destinationConnection, slogger)
		defer destDb.Db().Close()
		if err != nil {
			return schemaColMap
		}
		destColMap, err := destDb.Db().GetSchemaColumnMap(ctx)
		if err != nil {
			return schemaColMap
		}
		if len(destColMap) != 0 {
			return mergeSourceDestinationColumnInfo(sourceSchemaColumnInfoMap, destColMap)
		}
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
