package benthosbuilder_builders

import (
	"fmt"
	"iter"
	"log/slog"
	"maps"
	"slices"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	job_util "github.com/nucleuscloud/neosync/internal/job"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

type SqlJobMappingBuilder struct {
	schemaSettings *CascadeSchemaSettings

	sourceSchemaRows map[string]map[string][]*sqlmanager_shared.DatabaseSchemaRow
	destSchemaRows   map[string]map[string][]*sqlmanager_shared.DatabaseSchemaRow
}

func NewSqlJobMappingBuilder(
	schemaSettings *CascadeSchemaSettings,
	sourceSchemaRows map[string]map[string][]*sqlmanager_shared.DatabaseSchemaRow,
	destSchemaRows map[string]map[string][]*sqlmanager_shared.DatabaseSchemaRow,
) *SqlJobMappingBuilder {
	return &SqlJobMappingBuilder{
		schemaSettings:   schemaSettings,
		sourceSchemaRows: sourceSchemaRows,
		destSchemaRows:   destSchemaRows,
	}
}

func (b *SqlJobMappingBuilder) BuildJobMappings() ([]*shared.JobTransformationMapping, error) {
	schemas := b.getSchemasByStrategy(func() []string { return slices.Collect(maps.Keys(b.sourceSchemaRows)) })
	tables := b.getTablesByStrategy(schemas, func(schema string) []string { return slices.Collect(maps.Keys(b.sourceSchemaRows[schema])) })

	columnTransformErrors := []error{}
	jobMappings := []*shared.JobTransformationMapping{}
	for _, schema := range schemas {
		tables, ok := tables[schema]
		if !ok {
			continue
		}
		for _, table := range tables {
			columnRows, ok := b.sourceSchemaRows[schema][table]
			if !ok {
				continue
			}
			sourceColumnMap := maps.Collect(getColumnMapFromDbInfo(columnRows))
			destColumnRows, ok := b.destSchemaRows[schema][table]
			if !ok {
				destColumnRows = []*sqlmanager_shared.DatabaseSchemaRow{}
			}
			destColumnMap := maps.Collect(getColumnMapFromDbInfo(destColumnRows))

			columnTransforms, err := b.getColumnTransforms(schema, table, sourceColumnMap, destColumnMap)
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

// GetSchemasByStrategy returns the schemas to process based on the schema strategy
// If MapAll or default, returns input. If MapDefined, returns what is defined in the config, does not consider input
func (c *SqlJobMappingBuilder) getSchemasByStrategy(getDbSchemas func() []string) []string {
	switch c.schemaSettings.GetSchemaStrategy().GetStrategy().(type) {
	case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy_MapAllSchemas_:
		return getDbSchemas()
	case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy_MapDefinedSchemas_:
		return c.schemaSettings.GetDefinedSchemas()
	default:
		return getDbSchemas()
	}
}

func (c *SqlJobMappingBuilder) getTablesByStrategy(schemas []string, getDbTables func(schema string) []string) map[string][]string {
	output := map[string][]string{}
	for _, schema := range schemas {
		switch c.schemaSettings.GetTableStrategy(schema).GetStrategy().(type) {
		case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapAllTables_:
			output[schema] = getDbTables(schema)
		case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapDefinedTables_:
			output[schema] = c.schemaSettings.GetDefinedTables(schema)
		default:
			output[schema] = getDbTables(schema)
		}
	}
	return output
}

func (c *SqlJobMappingBuilder) getColumnTransforms(
	schemaName,
	tableName string,
	sourceColumnMap map[string]*sqlmanager_shared.DatabaseSchemaRow,
	destColumnMap map[string]*sqlmanager_shared.DatabaseSchemaRow,
) (map[string]*mgmtv1alpha1.TransformerConfig, error) {
	colStrategy := c.schemaSettings.GetColumnStrategy(schemaName, tableName)

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

func (c *SqlJobMappingBuilder) getDirectColumnTransforms(schemaName, tableName string) iter.Seq2[string, *mgmtv1alpha1.TransformerConfig] {
	return func(yield func(string, *mgmtv1alpha1.TransformerConfig) bool) {
		for _, schemaMapping := range c.schemaSettings.config.GetSchemaMappings() {
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

type CascadeSchemaSettings struct {
	config *mgmtv1alpha1.JobTypeConfig_JobTypeSync
}

func NewCascadeSchemaSettings(config *mgmtv1alpha1.JobTypeConfig_JobTypeSync) *CascadeSchemaSettings {
	return &CascadeSchemaSettings{
		config: config,
	}
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

type LegacySqlJobMappingBuilder struct {
	job              *mgmtv1alpha1.Job
	sourceColumnInfo map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow
	driver           string
	logger           *slog.Logger
}

func NewLegacySqlJobMappingBuilder(
	job *mgmtv1alpha1.Job,
	sourceColumnInfo map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
	driver string,
	logger *slog.Logger,
) *LegacySqlJobMappingBuilder {
	return &LegacySqlJobMappingBuilder{
		job:              job,
		sourceColumnInfo: sourceColumnInfo,
		driver:           driver,
		logger:           logger,
	}
}

func (b *LegacySqlJobMappingBuilder) BuildJobMappings() ([]*shared.JobTransformationMapping, error) {
	legacyMappings := b.job.GetMappings()
	if len(legacyMappings) == 0 {
		return nil, fmt.Errorf("no mappings found")
	}

	sqlSourceOpts, err := job_util.GetSqlJobSourceOpts(b.job.GetSource())
	if err != nil {
		return nil, fmt.Errorf("unable to get sql job source opts: %w", err)
	}

	if sqlSourceOpts != nil && sqlSourceOpts.HaltOnNewColumnAddition {
		newColumns, shouldHalt := shouldHaltOnSchemaAddition(b.sourceColumnInfo, legacyMappings)
		if shouldHalt {
			return nil, fmt.Errorf("%s: [%s]", haltOnSchemaAdditionErrMsg, strings.Join(newColumns, ", "))
		}
	}

	if sqlSourceOpts != nil && sqlSourceOpts.HaltOnColumnRemoval {
		missing, shouldHalt := isSourceMissingColumnsFoundInMappings(b.sourceColumnInfo, legacyMappings)
		if shouldHalt {
			return nil, fmt.Errorf("%s: [%s]", haltOnSchemaAdditionErrMsg, strings.Join(missing, ", "))
		}
	}

	// remove mappings that are not found in the source
	filteredExistingSourceMappings := removeMappingsNotFoundInSource(legacyMappings, b.sourceColumnInfo)

	if sqlSourceOpts != nil && sqlSourceOpts.GenerateNewColumnTransformers {
		extraMappings, err := getAdditionalJobMappings(b.driver, b.sourceColumnInfo, filteredExistingSourceMappings, splitKeyToTablePieces, b.logger)
		if err != nil {
			return nil, fmt.Errorf("unable to get additional legacy job mappings: %w", err)
		}
		b.logger.Debug(fmt.Sprintf("adding %d extra mappings due to unmapped columns", len(extraMappings)))
		filteredExistingSourceMappings = append(filteredExistingSourceMappings, extraMappings...)
	}
	jobMappings := make([]*shared.JobTransformationMapping, len(filteredExistingSourceMappings))
	for i, mapping := range filteredExistingSourceMappings {
		jobMappings[i] = jobMappingFromLegacyMapping(mapping)
	}
	return jobMappings, nil
}

func jobMappingFromLegacyMapping(mapping *mgmtv1alpha1.JobMapping) *shared.JobTransformationMapping {
	return &shared.JobTransformationMapping{
		JobMapping:        mapping,
		DestinationSchema: mapping.GetSchema(),
		DestinationTable:  mapping.GetTable(),
	}
}

// func (b *LegacySqlJobMappingBuilder) getSourceTableOptions() (map[string]*sqlSourceTableOptions, error) {
// 	sqlSourceOpts, err := job_util.GetSqlJobSourceOpts(b.job.GetSource())
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to get sql job source opts: %w", err)
// 	}
// 	if sqlSourceOpts == nil {
// 		return map[string]*sqlSourceTableOptions{}, nil
// 	}

// 	return groupSqlJobSourceOptionsByTable(sqlSourceOpts), nil
// }
