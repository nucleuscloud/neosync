package benthosbuilder_builders

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"maps"
	"slices"
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
	neosync_redis "github.com/nucleuscloud/neosync/internal/redis"
	rc "github.com/nucleuscloud/neosync/internal/runconfigs"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

type sqlSyncBuilder struct {
	transformerclient  mgmtv1alpha1connect.TransformersServiceClient
	sqlmanagerclient   sqlmanager.SqlManagerClient
	redisConfig        *neosync_redis.RedisConfig
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
}

func NewSqlSyncBuilder(
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	redisConfig *neosync_redis.RedisConfig,
	databaseDriver string,
	selectQueryBuilder bb_shared.SelectQueryMapBuilder,
	pageLimit int,
) bb_internal.BenthosBuilder {
	return &sqlSyncBuilder{
		transformerclient:  transformerclient,
		sqlmanagerclient:   sqlmanagerclient,
		redisConfig:        redisConfig,
		driver:             databaseDriver,
		selectQueryBuilder: selectQueryBuilder,
		pageLimit:          pageLimit,
	}
}

type CascadeSchemaSettings struct {
	config *mgmtv1alpha1.JobTypeConfig_JobTypeSync
}

func (c *CascadeSchemaSettings) GetSchemaStrategy() *mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy {
	return c.config.GetSchemaChange().GetSchemaStrategy()
}

func (c *CascadeSchemaSettings) GetTableStrategy(schemaName string) *mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy {
	for _, schemaMapping := range c.config.GetSchemaMappings() {
		if schemaMapping.GetSchema() == schemaName {
			ts := schemaMapping.GetTableStrategy()
			if ts != nil {
				return ts
			} else {
				break // fall back to global table strategy
			}
		}
	}
	return c.config.GetSchemaChange().GetTableStrategy()
}

func (c *CascadeSchemaSettings) GetColumnStrategy(schemaName, tableName string) *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy {
	for _, schemaMapping := range c.config.GetSchemaMappings() {
		if schemaMapping.GetSchema() == schemaName {
			for _, tableMapping := range schemaMapping.GetTableMappings() {
				if tableMapping.GetTable() == tableName {
					tableLevelColumnStrategy := tableMapping.GetColumnStrategy()
					if tableLevelColumnStrategy != nil {
						return tableLevelColumnStrategy
					}
					break // fall back to schema level column strategy
				}
			}
			schemaLevelColumnStrategy := schemaMapping.GetColumnStrategy()
			if schemaLevelColumnStrategy != nil {
				return schemaLevelColumnStrategy
			}
			break // fall back to global column strategy
		}
	}
	return c.config.GetSchemaChange().GetColumnStrategy()
}

func (c *CascadeSchemaSettings) GetDefinedSchemas() []string {
	output := map[string]bool{}
	for _, schemaMapping := range c.config.GetSchemaMappings() {
		output[schemaMapping.GetSchema()] = true
	}
	return slices.Collect(maps.Keys(output))
}

func (c *CascadeSchemaSettings) GetDefinedTables(schemaName string) []string {
	output := map[string]bool{}
	for _, schemaMapping := range c.config.GetSchemaMappings() {
		if schemaMapping.GetSchema() == schemaName {
			for _, tableMapping := range schemaMapping.GetTableMappings() {
				output[tableMapping.GetTable()] = true
			}
		}
	}
	return slices.Collect(maps.Keys(output))
}

func (c *CascadeSchemaSettings) GetColumnTransforms(
	schemaName,
	tableName string,
	sourceColumnMap map[string]*sqlmanager_shared.DatabaseSchemaRow,
	destColumnMap map[string]*sqlmanager_shared.DatabaseSchemaRow,
) (map[string]*mgmtv1alpha1.TransformerConfig, error) {
	colStrategy := c.GetColumnStrategy(schemaName, tableName)

	output := map[string]*mgmtv1alpha1.TransformerConfig{}
	maps.Insert(output, c.getDirectColumnTransforms(schemaName, tableName))

	switch colStrat := colStrategy.GetStrategy().(type) {
	case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_MapAllColumns_:
		if colStrat.MapAllColumns == nil {
			colStrat.MapAllColumns = &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_MapAllColumns{}
		}

		for colName := range output {
			if _, ok := sourceColumnMap[colName]; !ok {
				switch colStrat.MapAllColumns.GetColumnMappedNotInSource().GetStrategy().(type) {
				case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnMappedNotInSourceStrategy_Continue_:
					// do nothing
				case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnMappedNotInSourceStrategy_Halt_:
					return nil, fmt.Errorf("column %s in source but not mapped, and halt is set", colName)
				default:
					// do nothing
				}
			}
		}

		for _, columnRow := range sourceColumnMap {
			// column in source, but not mapped
			if _, ok := output[columnRow.ColumnName]; !ok {
				switch colStrat.MapAllColumns.GetColumnInSourceNotMapped().GetStrategy().(type) {
				case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy_AutoMap_:
					// todo: handle this, should not be passthrough
					output[columnRow.ColumnName] = &mgmtv1alpha1.TransformerConfig{
						Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
							PassthroughConfig: &mgmtv1alpha1.Passthrough{},
						},
					}
				case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy_Passthrough_:
					output[columnRow.ColumnName] = &mgmtv1alpha1.TransformerConfig{
						Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
							PassthroughConfig: &mgmtv1alpha1.Passthrough{},
						},
					}

				case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy_Halt_:
					// todo: handle this
					return nil, fmt.Errorf("column %s in source but not mapped, and halt is set", columnRow.ColumnName)
				case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy_Drop_:
					// todo: handle this
				default:
					output[columnRow.ColumnName] = &mgmtv1alpha1.TransformerConfig{
						Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
							PassthroughConfig: &mgmtv1alpha1.Passthrough{},
						},
					}
				}
			}
		}

		missingInDestColumns := []string{}
		for colName := range output {
			if _, ok := destColumnMap[colName]; !ok {
				missingInDestColumns = append(missingInDestColumns, colName)
			}
		}
		if len(missingInDestColumns) > 0 {
			switch colStrat.MapAllColumns.GetColumnInSourceMappedNotInDestination().GetStrategy().(type) {
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceMappedNotInDestinationStrategy_Continue_:
				// do nothing
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceMappedNotInDestinationStrategy_Drop_:
				for _, colName := range missingInDestColumns {
					delete(output, colName)
				}
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceMappedNotInDestinationStrategy_Halt_:
				return nil, fmt.Errorf("columns in source + mapped, but not in destination: %s.%s.[%s]", schemaName, tableName, strings.Join(missingInDestColumns, ", "))
			}
		}

		missingInSourceColumns := []string{}
		for _, columnRow := range destColumnMap {
			if _, ok := sourceColumnMap[columnRow.ColumnName]; !ok {
				missingInSourceColumns = append(missingInSourceColumns, columnRow.ColumnName)
			}
		}
		if len(missingInSourceColumns) > 0 {
			switch colStrat.MapAllColumns.GetColumnInDestinationNoLongerInSource().GetStrategy().(type) {
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInDestinationNotInSourceStrategy_Continue_:
				// do nothing
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInDestinationNotInSourceStrategy_Halt_:
				return nil, fmt.Errorf("columns in destination but not in source: %s.%s.[%s]", schemaName, tableName, strings.Join(missingInSourceColumns, ", "))
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInDestinationNotInSourceStrategy_AutoMap_:
				// todo: handle this, should not be passthrough
				for _, colName := range missingInSourceColumns {
					output[colName] = &mgmtv1alpha1.TransformerConfig{
						Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
							PassthroughConfig: &mgmtv1alpha1.Passthrough{},
						},
					}
				}
			default:
				// do nothing, should be the same as auto map
			}
		}

	case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_MapDefinedColumns_:
		if colStrat.MapDefinedColumns == nil {
			colStrat.MapDefinedColumns = &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_MapDefinedColumns{}
		}
		for colName := range output {
			if _, ok := sourceColumnMap[colName]; !ok {
				switch colStrat.MapDefinedColumns.GetColumnMappedNotInSource().GetStrategy().(type) {
				case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnMappedNotInSourceStrategy_Continue_:
					// do nothing
				case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnMappedNotInSourceStrategy_Halt_:
					return nil, fmt.Errorf("column %s in source but not mapped, and halt is set", colName)
				default:
					// do nothing
				}
			}
		}

		missingInDestColumns := []string{}
		for colName := range output {
			if _, ok := destColumnMap[colName]; !ok {
				missingInDestColumns = append(missingInDestColumns, colName)
			}
		}
		if len(missingInDestColumns) > 0 {
			switch colStrat.MapDefinedColumns.GetColumnInSourceMappedNotInDestination().GetStrategy().(type) {
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceMappedNotInDestinationStrategy_Continue_:
				// do nothing
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceMappedNotInDestinationStrategy_Drop_:
				for _, colName := range missingInDestColumns {
					delete(output, colName)
				}
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceMappedNotInDestinationStrategy_Halt_:
				return nil, fmt.Errorf("columns in source + mapped, but not in destination: %s.%s.[%s]", schemaName, tableName, strings.Join(missingInDestColumns, ", "))
			}
		}
	default:
		// do nothing for now
	}

	return output, nil
}

func (c *CascadeSchemaSettings) getDirectColumnTransforms(schemaName, tableName string) iter.Seq2[string, *mgmtv1alpha1.TransformerConfig] {
	return func(yield func(string, *mgmtv1alpha1.TransformerConfig) bool) {
		for _, schemaMapping := range c.config.GetSchemaMappings() {
			if schemaMapping.GetSchema() == schemaName {
				for _, tableMapping := range schemaMapping.GetTableMappings() {
					if tableMapping.GetTable() == tableName {
						for _, columnMapping := range tableMapping.GetColumnMappings() {
							if !yield(columnMapping.GetColumn(), columnMapping.GetTransformer()) {
								return
							}
						}
					}
				}
			}
		}
	}
}

// assumes the columnRows are already scoped to a table
func getColumnMapFromDbInfo(columnRows []*sqlmanager_shared.DatabaseSchemaRow) iter.Seq2[string, *sqlmanager_shared.DatabaseSchemaRow] {
	return func(yield func(string, *sqlmanager_shared.DatabaseSchemaRow) bool) {
		for _, columnRow := range columnRows {
			if !yield(columnRow.ColumnName, columnRow) {
				return
			}
		}
	}
}

func (b *sqlSyncBuilder) hydrateJobMappings(
	job *mgmtv1alpha1.Job,
	groupedColumnInfo map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
	logger *slog.Logger,
) ([]*shared.JobTransformationMapping, error) {
	legacyMappings := job.GetMappings()
	if len(legacyMappings) > 0 {
		jobMappings := make([]*shared.JobTransformationMapping, len(legacyMappings))
		for i, mapping := range legacyMappings {
			jobMappings[i] = &shared.JobTransformationMapping{
				JobMapping: mapping,
			}
		}
		return jobMappings, nil
	}

	syncConfig := job.GetJobType().GetSync()
	if syncConfig == nil {
		return nil, fmt.Errorf("unable to hydrate job mappings: sync config not found")
	}

	c := &CascadeSchemaSettings{
		config: syncConfig,
	}

	schemaTableColumnDbInfo := getSchemaTableColumnDbInfo(groupedColumnInfo)
	schemas := []string{}
	switch c.GetSchemaStrategy().GetStrategy().(type) {
	case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy_MapAllSchemas_:
		schemas = slices.Collect(maps.Keys(schemaTableColumnDbInfo))
	case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy_MapDefinedSchemas_:
		schemas = append(schemas, c.GetDefinedSchemas()...)
	default:
		schemas = slices.Collect(maps.Keys(schemaTableColumnDbInfo))
	}

	schemaTables := map[string][]string{}
	for _, schema := range schemas {
		tableMap, ok := schemaTableColumnDbInfo[schema]
		if !ok {
			continue
		}
		tableStrat := c.GetTableStrategy(schema)
		switch tableStrat.GetStrategy().(type) {
		case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapAllTables_:
			schemaTables[schema] = slices.Collect(maps.Keys(tableMap))
		case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapDefinedTables_:
			schemaTables[schema] = append(schemaTables[schema], c.GetDefinedTables(schema)...)
		default:
			schemaTables[schema] = slices.Collect(maps.Keys(tableMap)) // same as map all tables
		}
	}

	columnTransformErrors := []error{}
	jobMappings := []*shared.JobTransformationMapping{}
	for _, schema := range schemas {
		tables, ok := schemaTables[schema]
		if !ok {
			continue
		}
		for _, table := range tables {
			columnRows, ok := schemaTableColumnDbInfo[schema][table]
			if !ok {
				continue
			}
			columnMap := maps.Collect(getColumnMapFromDbInfo(columnRows))
			// todo: add separate destination column map
			columnTransforms, err := c.GetColumnTransforms(schema, table, columnMap, columnMap)
			if err != nil {
				columnTransformErrors = append(columnTransformErrors, err)
				continue
			}
			for colName, transformerConfig := range columnTransforms {
				jobMappings = append(jobMappings, &shared.JobTransformationMapping{
					JobMapping: &mgmtv1alpha1.JobMapping{
						Schema: schema,
						Table:  table,
						Column: colName,
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Config: transformerConfig,
						},
					},
				})
			}
		}
	}
	if len(columnTransformErrors) > 0 {
		return nil, fmt.Errorf("unable to build column transforms: %w", columnTransformErrors)
	}

	return jobMappings, nil
}

// outputs schema -> table -> []column info
func getSchemaTableColumnDbInfo(groupedColumnInfo map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow) map[string]map[string][]*sqlmanager_shared.DatabaseSchemaRow {
	output := map[string]map[string][]*sqlmanager_shared.DatabaseSchemaRow{}

	for _, columnMap := range groupedColumnInfo {
		for _, row := range columnMap {
			tableMap, ok := output[row.TableSchema]
			if !ok {
				tableMap = map[string][]*sqlmanager_shared.DatabaseSchemaRow{}
			}
			tableMap[row.TableName] = append(tableMap[row.TableName], row)
			output[row.TableSchema] = tableMap
		}
	}

	return output
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

	jobMappings, err := b.hydrateJobMappings(job, groupedColumnInfo, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to hydrate job mappings: %w", err)
	}

	var existingSourceMappings []*shared.JobTransformationMapping
	if len(job.Mappings) > 0 && job.GetJobType() == nil {
		b.sqlSourceSchemaColumnInfoMap = groupedColumnInfo
		if sqlSourceOpts != nil && sqlSourceOpts.HaltOnNewColumnAddition {
			newColumns, shouldHalt := shouldHaltOnSchemaAddition(groupedColumnInfo, job.Mappings)
			if shouldHalt {
				return nil, fmt.Errorf("%s: [%s]", haltOnSchemaAdditionErrMsg, strings.Join(newColumns, ", "))
			}
		}

		if sqlSourceOpts != nil && sqlSourceOpts.HaltOnColumnRemoval {
			missing, shouldHalt := isSourceMissingColumnsFoundInMappings(groupedColumnInfo, job.Mappings)
			if shouldHalt {
				return nil, fmt.Errorf("%s: [%s]", haltOnSchemaAdditionErrMsg, strings.Join(missing, ", "))
			}
		}

		// remove mappings that are not found in the source
		filteredExistingSourceMappings := removeMappingsNotFoundInSource(job.Mappings, groupedColumnInfo)

		if sqlSourceOpts != nil && sqlSourceOpts.GenerateNewColumnTransformers {
			extraMappings, err := getAdditionalJobMappings(b.driver, groupedColumnInfo, filteredExistingSourceMappings, splitKeyToTablePieces, logger)
			if err != nil {
				return nil, err
			}
			logger.Debug(fmt.Sprintf("adding %d extra mappings due to unmapped columns", len(extraMappings)))
			filteredExistingSourceMappings = append(filteredExistingSourceMappings, extraMappings...)
		}
		for _, mapping := range filteredExistingSourceMappings {
			existingSourceMappings = append(existingSourceMappings, &shared.JobTransformationMapping{
				JobMapping: mapping,
			})
		}
	} else {
		existingSourceMappings = jobMappings
	}
	uniqueSchemas := shared.GetUniqueSchemasFromMappings(existingSourceMappings)

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
		b.redisConfig,
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
	redisConfig *neosync_redis.RedisConfig,
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
				if b.redisConfig == nil {
					return nil, fmt.Errorf("missing redis config. this operation requires redis")
				}
				hashedKey := neosync_benthos.HashBenthosCacheKey(params.Job.GetId(), params.JobRunId, tableKey, col)
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
		sqlProcessor, err := getProcessors(b.driver, benthosConfig.Columns, colInfoMap, columnDefaultProperties)
		if err != nil {
			return nil, err
		}

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
