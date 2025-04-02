package jobmapping_builder_sql

import (
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	cascade_settings "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/builders/jobmapping-builder/settings"
	jobmapping_builder_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/builders/jobmapping-builder/shared"
	job_util "github.com/nucleuscloud/neosync/internal/job"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

const (
	jobmappingSubsetErrMsg     = "unable to continue: job mappings contain schemas, tables, or columns that were not found in the source connection"
	haltOnSchemaAdditionErrMsg = "unable to continue: HaltOnNewColumnAddition: job mappings are missing columns for the mapped tables found in the source connection"
	haltOnColumnRemovalErrMsg  = "unable to continue: HaltOnColumnRemoval: source database is missing columns for the mapped tables found in the job mappings"
)

type SqlJobMappingBuilder struct {
	schemaSettings *cascade_settings.CascadeSchemaSettings

	sourceSchemaRows map[string]map[string][]*sqlmanager_shared.DatabaseSchemaRow
	destSchemaRows   map[string]map[string][]*sqlmanager_shared.DatabaseSchemaRow
}

func NewSqlJobMappingBuilder(
	schemaSettings *cascade_settings.CascadeSchemaSettings,
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
		return nil, fmt.Errorf("unable to build column transforms: %w", errors.Join(columnTransformErrors...))
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
	maps.Insert(output, c.schemaSettings.GetColumnTransformerConfigByTable(schemaName, tableName))

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
	return jobmapping_builder_shared.JobMappingsFromLegacyMappings(filteredExistingSourceMappings), nil
}

// Based on the source schema, we check each mapped table for newly added columns that are not present in the mappings,
// but are present in the source. If so, halt because this means PII may be leaked.
func shouldHaltOnSchemaAddition(
	groupedSchemas map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
	mappings []*mgmtv1alpha1.JobMapping,
) ([]string, bool) {
	tableColMappings := getUniqueColMappingsMap(mappings)
	newColumns := []string{}
	for table, cols := range groupedSchemas {
		mappingCols, exists := tableColMappings[table]
		if !exists {
			// table not mapped in job mappings, skip
			continue
		}
		for col := range cols {
			if _, exists := mappingCols[col]; !exists {
				newColumns = append(newColumns, fmt.Sprintf("%s.%s", table, col))
			}
		}
	}
	return newColumns, len(newColumns) != 0
}

// Builds a map of <schema.table>->column
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

// checks that the source database has all the columns that are mapped in the job mappings
func isSourceMissingColumnsFoundInMappings(
	groupedSchemas map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
	mappings []*mgmtv1alpha1.JobMapping,
) ([]string, bool) {
	missingColumns := []string{}
	tableColMappings := getUniqueColMappingsMap(mappings)

	for schemaTable, cols := range tableColMappings {
		tableCols := groupedSchemas[schemaTable]
		for col := range cols {
			if _, ok := tableCols[col]; !ok {
				missingColumns = append(missingColumns, fmt.Sprintf("%s.%s", schemaTable, col))
			}
		}
	}
	return missingColumns, len(missingColumns) != 0
}

func removeMappingsNotFoundInSource(
	mappings []*mgmtv1alpha1.JobMapping,
	groupedSchemas map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
) []*mgmtv1alpha1.JobMapping {
	newMappings := make([]*mgmtv1alpha1.JobMapping, 0, len(mappings))
	for _, mapping := range mappings {
		key := sqlmanager_shared.BuildTable(mapping.Schema, mapping.Table)
		if _, ok := groupedSchemas[key]; ok {
			if _, ok := groupedSchemas[key][mapping.Column]; ok {
				newMappings = append(newMappings, mapping)
			}
		}
	}
	return newMappings
}

// Based on the source schema and the provided mappings, we find the missing columns (if any) and generate job mappings for them automatically
func getAdditionalJobMappings(
	driver string,
	groupedSchemas map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
	mappings []*mgmtv1alpha1.JobMapping,
	getTableFromKey func(key string) (schema, table string, err error),
	logger *slog.Logger,
) ([]*mgmtv1alpha1.JobMapping, error) {
	output := []*mgmtv1alpha1.JobMapping{}

	tableColMappings := getUniqueColMappingsMap(mappings)

	for schematable, cols := range groupedSchemas {
		mappedCols, ok := tableColMappings[schematable]
		if !ok {
			// todo: we may want to generate mappings for this entire table? However this may be dead code as we get the grouped schemas based on the mappings
			logger.Warn(
				"table found in schema data that is not present in job mappings",
				"table",
				schematable,
			)
			continue
		}
		if len(cols) == len(mappedCols) {
			continue
		}
		for col, info := range cols {
			if _, ok := mappedCols[col]; !ok {
				schema, table, err := getTableFromKey(schematable)
				if err != nil {
					return nil, err
				}
				// we found a column that is not present in the mappings, let's create a mapping for it
				if info.ColumnDefault != "" || info.IdentityGeneration != nil ||
					info.GeneratedType != nil {
					output = append(output, &mgmtv1alpha1.JobMapping{
						Schema: schema,
						Table:  table,
						Column: col,
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{
									GenerateDefaultConfig: &mgmtv1alpha1.GenerateDefault{},
								},
							},
						},
					})
				} else if info.IsNullable {
					output = append(output, &mgmtv1alpha1.JobMapping{
						Schema: schema,
						Table:  table,
						Column: col,
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
									Nullconfig: &mgmtv1alpha1.Null{},
								},
							},
						},
					})
				} else {
					switch driver {
					case sqlmanager_shared.PostgresDriver:
						transformer, err := getJmTransformerByPostgresDataType(info)
						if err != nil {
							return nil, err
						}
						output = append(output, &mgmtv1alpha1.JobMapping{
							Schema:      schema,
							Table:       table,
							Column:      col,
							Transformer: transformer,
						})
					case sqlmanager_shared.MysqlDriver:
						transformer, err := getJmTransformerByMysqlDataType(info)
						if err != nil {
							return nil, err
						}
						output = append(output, &mgmtv1alpha1.JobMapping{
							Schema:      schema,
							Table:       table,
							Column:      col,
							Transformer: transformer,
						})
					default:
						logger.Warn("this driver is not currently supported for additional job mapping by data type")
						return nil, fmt.Errorf("this driver %q does not currently support additional job mappings by data type. Please provide discrete job mappings for %q.%q.%q to continue: %w",
							driver, info.TableSchema, info.TableName, info.ColumnName, errors.ErrUnsupported,
						)
					}
				}
			}
		}
	}

	return output, nil
}

func getJmTransformerByPostgresDataType(
	colInfo *sqlmanager_shared.DatabaseSchemaRow,
) (*mgmtv1alpha1.JobMappingTransformer, error) {
	cleanedDataType := cleanPostgresType(colInfo.DataType)
	switch cleanedDataType {
	case "smallint":
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
					GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
						Min: shared.Ptr(int64(-32768)),
						Max: shared.Ptr(int64(32767)),
					},
				},
			},
		}, nil
	case "integer":
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
					GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
						Min: shared.Ptr(int64(-2147483648)),
						Max: shared.Ptr(int64(2147483647)),
					},
				},
			},
		}, nil
	case "bigint":
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
					GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
						Min: shared.Ptr(int64(-9223372036854775808)),
						Max: shared.Ptr(int64(9223372036854775807)),
					},
				},
			},
		}, nil
	case "decimal", "numeric":
		var precision *int64
		if colInfo.NumericPrecision > 0 {
			np := int64(colInfo.NumericPrecision)
			precision = &np
		}
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFloat64Config{
					GenerateFloat64Config: &mgmtv1alpha1.GenerateFloat64{
						Precision: precision, // todo: we need to expose scale...
					},
				},
			},
		}, nil
	case "real", "double precision":
		var precision *int64
		if colInfo.NumericPrecision > 0 {
			np := int64(colInfo.NumericPrecision)
			precision = &np
		}
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFloat64Config{
					GenerateFloat64Config: &mgmtv1alpha1.GenerateFloat64{
						Precision: precision,
					},
				},
			},
		}, nil

	case "smallserial", "serial", "bigserial":
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{
					GenerateDefaultConfig: &mgmtv1alpha1.GenerateDefault{},
				},
			},
		}, nil
	case "money":
		var precision *int64
		if colInfo.NumericPrecision > 0 {
			np := int64(colInfo.NumericPrecision)
			precision = &np
		}
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFloat64Config{
					GenerateFloat64Config: &mgmtv1alpha1.GenerateFloat64{
						// todo: to adequately support money, we need to know the scale which is set via the lc_monetary setting (but may be properly populated via our query..)
						Precision: precision,
						Min:       shared.Ptr(float64(-92233720368547758.08)),
						Max:       shared.Ptr(float64(92233720368547758.07)),
					},
				},
			},
		}, nil
	case "text",
		"bpchar",
		"character",
		"character varying": // todo: test to see if this works when (n) has been specified
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
					GenerateStringConfig: &mgmtv1alpha1.GenerateString{}, // todo?
				},
			},
		}, nil
	// case "bytea": // todo https://www.postgresql.org/docs/current/datatype-binary.html
	// case "date":
	// 	return &mgmtv1alpha1.JobMappingTransformer{}

	// case "time without time zone":
	// 	return &mgmtv1alpha1.JobMappingTransformer{}

	// case "time with time zone":
	// 	return &mgmtv1alpha1.JobMappingTransformer{}

	// case "interval":
	// 	return &mgmtv1alpha1.JobMappingTransformer{}

	// case "timestamp without time zone":
	// 	return &mgmtv1alpha1.JobMappingTransformer{}

	// case "timestamp with time zone":
	// 	return &mgmtv1alpha1.JobMappingTransformer{}

	case "boolean":
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{
					GenerateBoolConfig: &mgmtv1alpha1.GenerateBool{},
				},
			},
		}, nil
	case "uuid":
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
					GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
						IncludeHyphens: shared.Ptr(true),
					},
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf(
			"uncountered unsupported data type %q for %q.%q.%q when attempting to generate an auto-mapper. To continue, provide a discrete job mapping for this column.: %w",
			colInfo.DataType,
			colInfo.TableSchema,
			colInfo.TableName,
			colInfo.ColumnName,
			errors.ErrUnsupported,
		)
	}
}

func getJmTransformerByMysqlDataType(
	colInfo *sqlmanager_shared.DatabaseSchemaRow,
) (*mgmtv1alpha1.JobMappingTransformer, error) {
	cleanedDataType := cleanMysqlType(colInfo.MysqlColumnType)
	switch cleanedDataType {
	case "char":
		params := extractMysqlTypeParams(colInfo.MysqlColumnType)
		minLength := int64(0)
		maxLength := int64(255)
		if len(params) > 0 {
			fixedLength, err := strconv.ParseInt(params[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to parse length for type %q: %w",
					colInfo.MysqlColumnType,
					err,
				)
			}
			minLength = fixedLength
			maxLength = fixedLength
		} else if colInfo.CharacterMaximumLength > 0 {
			maxLength = int64(colInfo.CharacterMaximumLength)
		}
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
					GenerateStringConfig: &mgmtv1alpha1.GenerateString{
						Min: shared.Ptr(minLength),
						Max: shared.Ptr(maxLength),
					},
				},
			},
		}, nil

	case "varchar":
		params := extractMysqlTypeParams(colInfo.MysqlColumnType)
		maxLength := int64(65535)
		if len(params) > 0 {
			fixedLength, err := strconv.ParseInt(params[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to parse length for type %q: %w",
					colInfo.MysqlColumnType,
					err,
				)
			}
			maxLength = fixedLength
		} else if colInfo.CharacterMaximumLength > 0 {
			maxLength = int64(colInfo.CharacterMaximumLength)
		}
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
					GenerateStringConfig: &mgmtv1alpha1.GenerateString{Max: shared.Ptr(maxLength)},
				},
			},
		}, nil

	case "tinytext":
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
					GenerateStringConfig: &mgmtv1alpha1.GenerateString{Max: shared.Ptr(int64(255))},
				},
			},
		}, nil

	case "text":
		params := extractMysqlTypeParams(colInfo.MysqlColumnType)
		maxLength := int64(65535)
		if len(params) > 0 {
			length, err := strconv.ParseInt(params[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to parse length for type %q: %w",
					colInfo.MysqlColumnType,
					err,
				)
			}
			maxLength = length
		} else if colInfo.CharacterMaximumLength > 0 {
			maxLength = int64(colInfo.CharacterMaximumLength)
		}
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
					GenerateStringConfig: &mgmtv1alpha1.GenerateString{Max: shared.Ptr(maxLength)},
				},
			},
		}, nil

	case "mediumtext":
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
					GenerateStringConfig: &mgmtv1alpha1.GenerateString{
						Max: shared.Ptr(int64(16_777_215)),
					},
				},
			},
		}, nil
	case "longtext":
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
					GenerateStringConfig: &mgmtv1alpha1.GenerateString{
						Max: shared.Ptr(int64(4_294_967_295)),
					},
				},
			},
		}, nil
	case "enum", "set":
		params := extractMysqlTypeParams(colInfo.MysqlColumnType)
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCategoricalConfig{
					GenerateCategoricalConfig: &mgmtv1alpha1.GenerateCategorical{
						Categories: shared.Ptr(strings.Join(params, ",")),
					},
				},
			},
		}, nil

	case "tinyint":
		isUnsigned := strings.Contains(strings.ToLower(colInfo.MysqlColumnType), "unsigned")
		var minVal, maxVal int64
		if isUnsigned {
			minVal = 0
			maxVal = 255 // 2^8 - 1
		} else {
			minVal = -128 // -2^7
			maxVal = 127  // 2^7 - 1
		}
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
					GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
						Min: shared.Ptr(minVal),
						Max: shared.Ptr(maxVal),
					},
				},
			},
		}, nil

	case "smallint":
		isUnsigned := strings.Contains(strings.ToLower(colInfo.MysqlColumnType), "unsigned")
		var minVal, maxVal int64
		if isUnsigned {
			minVal = 0
			maxVal = 65535 // 2^16 - 1
		} else {
			minVal = -32768 // -2^15
			maxVal = 32767  // 2^15 - 1
		}
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
					GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
						Min: shared.Ptr(minVal),
						Max: shared.Ptr(maxVal),
					},
				},
			},
		}, nil
	case "mediumint":
		isUnsigned := strings.Contains(strings.ToLower(colInfo.MysqlColumnType), "unsigned")
		var minVal, maxVal int64
		if isUnsigned {
			minVal = 0
			maxVal = 16777215 // 2^24 - 1
		} else {
			minVal = -8388608 // -2^23
			maxVal = 8388607  // 2^23 - 1
		}
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
					GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
						Min: shared.Ptr(minVal),
						Max: shared.Ptr(maxVal),
					},
				},
			},
		}, nil
	case "int", "integer":
		isUnsigned := strings.Contains(strings.ToLower(colInfo.MysqlColumnType), "unsigned")
		var minVal, maxVal int64
		if isUnsigned {
			minVal = 0
			maxVal = 4294967295 // 2^32 - 1
		} else {
			minVal = -2147483648 // -2^31
			maxVal = 2147483647  // 2^31 - 1
		}
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
					GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
						Min: shared.Ptr(minVal),
						Max: shared.Ptr(maxVal),
					},
				},
			},
		}, nil
	case "bigint":
		minVal := int64(0)             // -2^63
		maxVal := int64(math.MaxInt64) // 2^63 - 1
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
					GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
						Min: shared.Ptr(minVal),
						Max: shared.Ptr(maxVal),
					},
				},
			},
		}, nil
	case "float":
		precision := int64(colInfo.NumericPrecision)
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFloat64Config{
					GenerateFloat64Config: &mgmtv1alpha1.GenerateFloat64{
						Precision: &precision,
					},
				},
			},
		}, nil
	case "double", "double precision", "decimal", "dec":
		precision := int64(colInfo.NumericPrecision)
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFloat64Config{
					GenerateFloat64Config: &mgmtv1alpha1.GenerateFloat64{
						Precision: &precision, // todo: expose scale
					},
				},
			},
		}, nil

	// case "bit":
	// 	params := extractMysqlTypeParams(colInfo.MysqlColumnType)
	// 	bitLength := int64(1) // default length is 1
	// 	if len(params) > 0 {
	// 		if parsed, err := strconv.ParseInt(params[0], 10, 64); err == nil && parsed > 0 && parsed <= 64 {
	// 			bitLength = parsed
	// 		}
	// 	}
	// 	return &mgmtv1alpha1.JobMappingTransformer{
	// 		Config: &mgmtv1alpha1.TransformerConfig{
	// 			Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
	// 				GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
	// 					Code: fmt.Sprintf(`
	// 						// Generate random bits up to specified length
	// 						const length = %d;
	// 						let value = 0;
	// 						for (let i = 0; i < length; i++) {
	// 							if (Math.random() < 0.5) {
	// 								value |= (1 << i);
	// 							}
	// 						}
	// 						// Convert to binary string padded to the correct length
	// 						return value.toString(2).padStart(length, '0');
	// 					`, bitLength),
	// 				},
	// 			},
	// 		},
	// 	}, nil
	// case "binary", "varbinary":
	// 	params := extractMysqlTypeParams(colInfo.DataType)
	// 	maxLength := int64(255) // default max length
	// 	if len(params) > 0 {
	// 		if parsed, err := strconv.ParseInt(params[0], 10, 64); err == nil && parsed > 0 && parsed <= 255 {
	// 			maxLength = parsed
	// 		}
	// 	}
	// 	return &mgmtv1alpha1.JobMappingTransformer{
	// 		Config: &mgmtv1alpha1.TransformerConfig{
	// 			Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
	// 				GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
	// 					Code: fmt.Sprintf(`
	// 						// Generate random binary data up to maxLength bytes
	// 						const maxLength = %d;
	// 						const length = Math.floor(Math.random() * maxLength) + 1;
	// 						const bytes = new Uint8Array(length);
	// 						for (let i = 0; i < length; i++) {
	// 							bytes[i] = Math.floor(Math.random() * 256);
	// 						}
	// 						// Convert to base64 for safe transport
	// 						return Buffer.from(bytes).toString('base64');
	// 					`, maxLength),
	// 				},
	// 			},
	// 		},
	// 	}, nil
	// case "tinyblob":
	// 	return &mgmtv1alpha1.JobMappingTransformer{
	// 		Config: &mgmtv1alpha1.TransformerConfig{
	// 			Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
	// 				GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
	// 					Code: `
	// 						// Generate random TINYBLOB (max 255 bytes)
	// 						const maxLength = 255;
	// 						const length = Math.floor(Math.random() * maxLength) + 1;
	// 						const bytes = new Uint8Array(length);
	// 						for (let i = 0; i < length; i++) {
	// 							bytes[i] = Math.floor(Math.random() * 256);
	// 						}
	// 						return Buffer.from(bytes).toString('base64');
	// 					`,
	// 				},
	// 			},
	// 		},
	// 	}, nil
	// case "blob":
	// 	return &mgmtv1alpha1.JobMappingTransformer{
	// 		Config: &mgmtv1alpha1.TransformerConfig{
	// 			Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
	// 				GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
	// 					Code: `
	// 						// Generate random BLOB (max 65,535 bytes)
	// 						// Using a smaller max for practical purposes
	// 						const maxLength = 1024; // Using 1KB for reasonable performance
	// 						const length = Math.floor(Math.random() * maxLength) + 1;
	// 						const bytes = new Uint8Array(length);
	// 						for (let i = 0; i < length; i++) {
	// 							bytes[i] = Math.floor(Math.random() * 256);
	// 						}
	// 						return Buffer.from(bytes).toString('base64');
	// 					`,
	// 				},
	// 			},
	// 		},
	// 	}, nil
	// case "mediumblob":
	// 	return &mgmtv1alpha1.JobMappingTransformer{
	// 		Config: &mgmtv1alpha1.TransformerConfig{
	// 			Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
	// 				GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
	// 					Code: `
	// 						// Generate random MEDIUMBLOB (max 16,777,215 bytes)
	// 						// Using a smaller max for practical purposes
	// 						const maxLength = 2048; // Using 2KB for reasonable performance
	// 						const length = Math.floor(Math.random() * maxLength) + 1;
	// 						const bytes = new Uint8Array(length);
	// 						for (let i = 0; i < length; i++) {
	// 							bytes[i] = Math.floor(Math.random() * 256);
	// 						}
	// 						return Buffer.from(bytes).toString('base64');
	// 					`,
	// 				},
	// 			},
	// 		},
	// 	}, nil
	// case "longblob":
	// 	return &mgmtv1alpha1.JobMappingTransformer{
	// 		Config: &mgmtv1alpha1.TransformerConfig{
	// 			Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
	// 				GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
	// 					Code: `
	// 						// Generate random LONGBLOB (max 4,294,967,295 bytes)
	// 						// Using a smaller max for practical purposes
	// 						const maxLength = 4096; // Using 4KB for reasonable performance
	// 						const length = Math.floor(Math.random() * maxLength) + 1;
	// 						const bytes = new Uint8Array(length);
	// 						for (let i = 0; i < length; i++) {
	// 							bytes[i] = Math.floor(Math.random() * 256);
	// 						}
	// 						return Buffer.from(bytes).toString('base64');
	// 					`,
	// 				},
	// 			},
	// 		},
	// 	}, nil

	case "date":
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
					GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
						Code: `
								const date = new Date();
								const year = date.getFullYear();
								const month = String(date.getMonth() + 1).padStart(2, '0');
								const day = String(date.getDate()).padStart(2, '0');
								return year + "-" + month + "-" + day;
							`,
					},
				},
			},
		}, nil
	case "datetime", "timestamp":
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
					GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
						Code: `
								const date = new Date();
								const year = date.getFullYear();
								const month = String(date.getMonth() + 1).padStart(2, '0');
								const day = String(date.getDate()).padStart(2, '0');
								const hours = String(date.getHours()).padStart(2, '0');
								const minutes = String(date.getMinutes()).padStart(2, '0');
								const seconds = String(date.getSeconds()).padStart(2, '0');
								return year + "-" + month + "-" + day + " " + hours + ":" + minutes + ":" + seconds;
							`,
					},
				},
			},
		}, nil
	case "time":
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
					GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
						Code: `
								const date = new Date();
								const hours = String(date.getHours()).padStart(2, '0');
								const minutes = String(date.getMinutes()).padStart(2, '0');
								const seconds = String(date.getSeconds()).padStart(2, '0');
								return hours + ":" + minutes + ":" + seconds;
							`,
					},
				},
			},
		}, nil
	case "year":
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
					GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
						Code: `
								const date = new Date();
								return date.getFullYear();
							`,
					},
				},
			},
		}, nil
	case "boolean", "bool":
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{
					GenerateBoolConfig: &mgmtv1alpha1.GenerateBool{},
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf(
			"uncountered unsupported data type %q for %q.%q.%q when attempting to generate an auto-mapper. To continue, provide a discrete job mapping for this column.: %w",
			colInfo.DataType,
			colInfo.TableSchema,
			colInfo.TableName,
			colInfo.ColumnName,
			errors.ErrUnsupported,
		)
	}
}

func cleanPostgresType(dataType string) string {
	parenIndex := strings.Index(dataType, "(")
	if parenIndex == -1 {
		return dataType
	}
	return strings.TrimSpace(dataType[:parenIndex])
}

func cleanMysqlType(dataType string) string {
	parenIndex := strings.Index(dataType, "(")
	if parenIndex == -1 {
		return dataType
	}
	return strings.TrimSpace(dataType[:parenIndex])
}

// extractMysqlTypeParams extracts the parameters from MySQL data type definitions
// Examples:
// - CHAR(10) -> ["10"]
// - FLOAT(10, 2) -> ["10", "2"]
// - ENUM('val1', 'val2') -> ["val1", "val2"]
func extractMysqlTypeParams(dataType string) []string {
	parenIndex := strings.Index(dataType, "(")
	if parenIndex == -1 {
		return nil
	}

	closingIndex := strings.LastIndex(dataType, ")")
	if closingIndex == -1 {
		return nil
	}

	// Extract content between parentheses
	paramsStr := dataType[parenIndex+1 : closingIndex]

	// Handle ENUM/SET cases which use quotes
	if strings.Contains(paramsStr, "'") {
		// Split by comma and handle quoted values
		params := strings.Split(paramsStr, ",")
		result := make([]string, 0, len(params))
		for _, p := range params {
			// Remove quotes and whitespace
			p = strings.Trim(strings.TrimSpace(p), "'")
			if p != "" {
				result = append(result, p)
			}
		}
		return result
	}

	// Handle regular numeric parameters
	params := strings.Split(paramsStr, ",")
	result := make([]string, 0, len(params))
	for _, p := range params {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func shouldOverrideColumnDefault(
	columnDefaults map[string]*neosync_benthos.ColumnDefaultProperties,
) bool {
	for _, cd := range columnDefaults {
		if cd != nil && !cd.HasDefaultTransformer && cd.NeedsOverride {
			return true
		}
	}
	return false
}

func splitKeyToTablePieces(key string) (schema, table string, err error) {
	pieces := strings.SplitN(key, ".", 2)
	if len(pieces) != 2 {
		return "", "", errors.New("unable to split key to get schema and table, not 2 pieces")
	}
	return pieces[0], pieces[1], nil
}
