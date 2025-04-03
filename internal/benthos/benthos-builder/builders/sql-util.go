package benthosbuilder_builders

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	job_util "github.com/nucleuscloud/neosync/internal/job"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

const (
	jobmappingSubsetErrMsg     = "unable to continue: job mappings contain schemas, tables, or columns that were not found in the source connection"
	haltOnSchemaAdditionErrMsg = "unable to continue: HaltOnNewColumnAddition: job mappings are missing columns for the mapped tables found in the source connection"
	haltOnColumnRemovalErrMsg  = "unable to continue: HaltOnColumnRemoval: source database is missing columns for the mapped tables found in the job mappings"
)

type sqlSourceTableOptions struct {
	WhereClause *string
}

type tableMapping struct {
	Schema   string
	Table    string
	Mappings []*mgmtv1alpha1.JobMapping
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

func getMapValuesCount[K comparable, V any](m map[K][]V) int {
	count := 0
	for _, v := range m {
		count += len(v)
	}
	return count
}

func buildPlainColumns(mappings []*mgmtv1alpha1.JobMapping) []string {
	columns := make([]string, len(mappings))
	for idx := range mappings {
		columns[idx] = mappings[idx].Column
	}
	return columns
}

func buildTableSubsetMap(
	tableOpts map[string]*sqlSourceTableOptions,
	tableMap map[string]*tableMapping,
) map[string]string {
	tableSubsetMap := map[string]string{}
	for table, opts := range tableOpts {
		if _, ok := tableMap[table]; !ok {
			continue
		}
		if opts != nil && opts.WhereClause != nil && *opts.WhereClause != "" {
			tableSubsetMap[table] = *opts.WhereClause
		}
	}
	return tableSubsetMap
}

func groupSqlJobSourceOptionsByTable(
	sqlSourceOpts *job_util.SqlJobSourceOpts,
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

func mergeVirtualForeignKeys(
	dbForeignKeys map[string][]*sqlmanager_shared.ForeignConstraint,
	virtualForeignKeys []*mgmtv1alpha1.VirtualForeignConstraint,
	colInfoMap map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
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

func getColumnTransformerMap(
	tableMappingMap map[string]*tableMapping,
) map[string]map[string]*mgmtv1alpha1.JobMappingTransformer {
	colTransformerMap := map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{} // schema.table ->  column -> transformer
	for table, mapping := range tableMappingMap {
		colTransformerMap[table] = map[string]*mgmtv1alpha1.JobMappingTransformer{}
		for _, m := range mapping.Mappings {
			colTransformerMap[table][m.Column] = m.Transformer
		}
	}
	return colTransformerMap
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

func filterForeignKeysMap(
	colTransformerMap map[string]map[string]*mgmtv1alpha1.JobMappingTransformer,
	foreignKeysMap map[string][]*sqlmanager_shared.ForeignConstraint,
) map[string][]*sqlmanager_shared.ForeignConstraint {
	newFkMap := make(map[string][]*sqlmanager_shared.ForeignConstraint)

	for table, fks := range foreignKeysMap {
		cols, ok := colTransformerMap[table]
		if !ok {
			continue
		}
		for _, fk := range fks {
			newFk := &sqlmanager_shared.ForeignConstraint{
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table: fk.ForeignKey.Table,
				},
			}
			for i, c := range fk.Columns {
				t, ok := cols[c]
				if !fk.NotNullable[i] && (!ok || isNullJobMappingTransformer(t)) {
					continue
				}

				newFk.Columns = append(newFk.Columns, c)
				newFk.NotNullable = append(newFk.NotNullable, fk.NotNullable[i])
				newFk.ForeignKey.Columns = append(
					newFk.ForeignKey.Columns,
					fk.ForeignKey.Columns[i],
				)
			}

			if len(newFk.Columns) > 0 {
				newFkMap[table] = append(newFkMap[table], newFk)
			}
		}
	}
	return newFkMap
}

func isNullJobMappingTransformer(t *mgmtv1alpha1.JobMappingTransformer) bool {
	switch t.GetConfig().GetConfig().(type) {
	case *mgmtv1alpha1.TransformerConfig_Nullconfig:
		return true
	default:
		return false
	}
}

func isDefaultJobMappingTransformer(t *mgmtv1alpha1.JobMappingTransformer) bool {
	switch t.GetConfig().GetConfig().(type) {
	case *mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig:
		return true
	default:
		return false
	}
}

// map of table primary key cols to foreign key cols
func getPrimaryKeyDependencyMap(
	tableDependencies map[string][]*sqlmanager_shared.ForeignConstraint,
) map[string]map[string][]*bb_internal.ReferenceKey {
	tc := map[string]map[string][]*bb_internal.ReferenceKey{} // schema.table -> column -> ForeignKey
	for table, constraints := range tableDependencies {
		for _, c := range constraints {
			_, ok := tc[c.ForeignKey.Table]
			if !ok {
				tc[c.ForeignKey.Table] = map[string][]*bb_internal.ReferenceKey{}
			}
			for idx, col := range c.ForeignKey.Columns {
				tc[c.ForeignKey.Table][col] = append(
					tc[c.ForeignKey.Table][col],
					&bb_internal.ReferenceKey{
						Table:  table,
						Column: c.Columns[idx],
					},
				)
			}
		}
	}
	return tc
}

func findTopForeignKeySource(
	tableName, col string,
	tableDependencies map[string][]*sqlmanager_shared.ForeignConstraint,
) *bb_internal.ReferenceKey {
	// Add the foreign key dependencies of the current table
	if foreignKeys, ok := tableDependencies[tableName]; ok {
		for _, fk := range foreignKeys {
			for idx, c := range fk.Columns {
				if c == col {
					// Recursively add dependent tables and their foreign keys
					return findTopForeignKeySource(
						fk.ForeignKey.Table,
						fk.ForeignKey.Columns[idx],
						tableDependencies,
					)
				}
			}
		}
	}
	return &bb_internal.ReferenceKey{
		Table:  tableName,
		Column: col,
	}
}

// builds schema.table -> FK column ->  PK schema table column
// find top level primary key column if foreign keys are nested
func buildForeignKeySourceMap(
	tableDeps map[string][]*sqlmanager_shared.ForeignConstraint,
) map[string]map[string]*bb_internal.ReferenceKey {
	outputMap := map[string]map[string]*bb_internal.ReferenceKey{}
	for tableName, constraints := range tableDeps {
		if _, ok := outputMap[tableName]; !ok {
			outputMap[tableName] = map[string]*bb_internal.ReferenceKey{}
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

func getTransformedFksMap(
	tabledependencies map[string][]*sqlmanager_shared.ForeignConstraint,
	colTransformerMap map[string]map[string]*mgmtv1alpha1.JobMappingTransformer,
) map[string]map[string][]*bb_internal.ReferenceKey {
	foreignKeyToSourceMap := buildForeignKeySourceMap(tabledependencies)
	// filter this list by table constraints that has transformer
	transformedForeignKeyToSourceMap := map[string]map[string][]*bb_internal.ReferenceKey{} // schema.table -> column -> foreignKey
	for table, constraints := range foreignKeyToSourceMap {
		_, ok := transformedForeignKeyToSourceMap[table]
		if !ok {
			transformedForeignKeyToSourceMap[table] = map[string][]*bb_internal.ReferenceKey{}
		}
		for col, tc := range constraints {
			// only add constraint if foreign key has transformer
			transformer, transformerOk := colTransformerMap[tc.Table][tc.Column]
			if transformerOk && shouldProcessStrict(transformer) {
				transformedForeignKeyToSourceMap[table][col] = append(
					transformedForeignKeyToSourceMap[table][col],
					tc,
				)
			}
		}
	}
	return transformedForeignKeyToSourceMap
}

func getColumnDefaultProperties(
	slogger *slog.Logger,
	driver string,
	cols []string,
	colInfo map[string]*sqlmanager_shared.DatabaseSchemaRow,
	colTransformers map[string]*mgmtv1alpha1.JobMappingTransformer,
) (map[string]*neosync_benthos.ColumnDefaultProperties, error) {
	colDefaults := map[string]*neosync_benthos.ColumnDefaultProperties{}
	for _, cName := range cols {
		info, ok := colInfo[cName]
		if !ok {
			return nil, fmt.Errorf("column default type missing. column: %s", cName)
		}
		needsOverride, needsReset, err := sqlmanager.GetColumnOverrideAndResetProperties(
			driver,
			info,
		)
		if err != nil {
			slogger.Error(
				"unable to determine SQL column default flags",
				"error",
				err,
				"column",
				cName,
			)
			return nil, err
		}

		jmTransformer, ok := colTransformers[cName]
		if !ok {
			return nil, fmt.Errorf("transformer missing for column: %s", cName)
		}

		var hasDefaultTransformer bool
		if jmTransformer != nil && isDefaultJobMappingTransformer(jmTransformer) {
			hasDefaultTransformer = true
		}
		if !needsReset && !needsOverride && !hasDefaultTransformer {
			continue
		}
		colDefaults[cName] = &neosync_benthos.ColumnDefaultProperties{
			NeedsReset:            needsReset,
			NeedsOverride:         needsOverride,
			HasDefaultTransformer: hasDefaultTransformer,
		}
	}
	return colDefaults, nil
}

type destinationOptions struct {
	OnConflictDoNothing      bool
	OnConflictDoUpdate       bool
	Truncate                 bool
	TruncateCascade          bool
	SkipForeignKeyViolations bool
	MaxInFlight              uint32
	BatchCount               int
	BatchPeriod              string
}

func getDestinationOptions(
	destOpts *mgmtv1alpha1.JobDestinationOptions,
) (*destinationOptions, error) {
	if destOpts.GetConfig() == nil {
		return &destinationOptions{}, nil
	}
	switch config := destOpts.GetConfig().(type) {
	case *mgmtv1alpha1.JobDestinationOptions_PostgresOptions:
		if config.PostgresOptions == nil {
			return &destinationOptions{}, nil
		}
		batchingConfig, err := getParsedBatchingConfig(config.PostgresOptions)
		if err != nil {
			return nil, err
		}
		onConflictDoNothing := false
		onConflictDoUpdate := false
		if config.PostgresOptions.GetOnConflict().GetNothing() != nil {
			onConflictDoNothing = true
		} else if config.PostgresOptions.GetOnConflict().GetUpdate() != nil {
			onConflictDoUpdate = true
		}
		if onConflictDoNothing && onConflictDoUpdate {
			return nil, fmt.Errorf("cannot have both on conflict do nothing and on conflict do update")
		}
		return &destinationOptions{
			OnConflictDoNothing:      onConflictDoNothing,
			OnConflictDoUpdate:       onConflictDoUpdate,
			Truncate:                 config.PostgresOptions.GetTruncateTable().GetTruncateBeforeInsert(),
			TruncateCascade:          config.PostgresOptions.GetTruncateTable().GetCascade(),
			SkipForeignKeyViolations: config.PostgresOptions.GetSkipForeignKeyViolations(),
			MaxInFlight:              batchingConfig.MaxInFlight,
			BatchCount:               batchingConfig.BatchCount,
			BatchPeriod:              batchingConfig.BatchPeriod,
		}, nil
	case *mgmtv1alpha1.JobDestinationOptions_MysqlOptions:
		if config.MysqlOptions == nil {
			return &destinationOptions{}, nil
		}
		batchingConfig, err := getParsedBatchingConfig(config.MysqlOptions)
		if err != nil {
			return nil, err
		}
		onConflictDoNothing := false
		onConflictDoUpdate := false
		if config.MysqlOptions.GetOnConflict().GetNothing() != nil {
			onConflictDoNothing = true
		} else if config.MysqlOptions.GetOnConflict().GetUpdate() != nil {
			onConflictDoUpdate = true
		}
		if onConflictDoNothing && onConflictDoUpdate {
			return nil, fmt.Errorf("cannot have both on conflict do nothing and on conflict do update")
		}
		return &destinationOptions{
			OnConflictDoNothing:      onConflictDoNothing,
			OnConflictDoUpdate:       onConflictDoUpdate,
			Truncate:                 config.MysqlOptions.GetTruncateTable().GetTruncateBeforeInsert(),
			SkipForeignKeyViolations: config.MysqlOptions.GetSkipForeignKeyViolations(),
			MaxInFlight:              batchingConfig.MaxInFlight,
			BatchCount:               batchingConfig.BatchCount,
			BatchPeriod:              batchingConfig.BatchPeriod,
		}, nil
	case *mgmtv1alpha1.JobDestinationOptions_MssqlOptions:
		if config.MssqlOptions == nil {
			return &destinationOptions{}, nil
		}
		batchingConfig, err := getParsedBatchingConfig(config.MssqlOptions)
		if err != nil {
			return nil, err
		}
		return &destinationOptions{
			SkipForeignKeyViolations: config.MssqlOptions.GetSkipForeignKeyViolations(),
			MaxInFlight:              batchingConfig.MaxInFlight,
			BatchCount:               batchingConfig.BatchCount,
			BatchPeriod:              batchingConfig.BatchPeriod,
		}, nil
	default:
		return &destinationOptions{}, nil
	}
}

type batchingConfig struct {
	MaxInFlight uint32
	BatchPeriod string
	BatchCount  int
}
type batchDestinationOption interface {
	GetMaxInFlight() uint32
	GetBatch() *mgmtv1alpha1.BatchConfig
}

func getParsedBatchingConfig(destOpt batchDestinationOption) (batchingConfig, error) {
	output := batchingConfig{
		MaxInFlight: 10,
		BatchPeriod: "5s",
		BatchCount:  100,
	}
	if destOpt == nil {
		return output, nil
	}
	if destOpt.GetMaxInFlight() > 0 {
		output.MaxInFlight = destOpt.GetMaxInFlight()
	}

	batchConfig := destOpt.GetBatch()
	if batchConfig != nil {
		output.BatchCount = int(batchConfig.GetCount())

		if batchConfig.GetPeriod() != "" {
			_, err := time.ParseDuration(batchConfig.GetPeriod())
			if err != nil {
				return batchingConfig{}, fmt.Errorf(
					"unable to parse batch period for s3 destination config: %w",
					err,
				)
			}
		}
		output.BatchPeriod = batchConfig.GetPeriod()
	}

	if output.BatchCount == 0 && output.BatchPeriod == "" {
		return batchingConfig{}, fmt.Errorf(
			"must have at least one batch policy configured. Cannot disable both period and count",
		)
	}
	return output, nil
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
	case "time without time zone":
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
	case "time with time zone":
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
					GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
						Code: `
								const date = new Date();
								const hours = String(date.getUTCHours()).padStart(2, '0');
								const minutes = String(date.getUTCMinutes()).padStart(2, '0');
								const seconds = String(date.getUTCSeconds()).padStart(2, '0');
								const timezoneOffset = -date.getTimezoneOffset();
								const absOffset = Math.abs(timezoneOffset);
								const offsetHours = String(Math.floor(absOffset / 60)).padStart(2, '0');
								const offsetMinutes = String(absOffset % 60).padStart(2, '0');
								const offsetSign = timezoneOffset >= 0 ? '+' : '-';
								return hours + ":" + minutes + ":" + seconds + offsetSign + offsetHours + ":" + offsetMinutes;
							`,
					},
				},
			},
		}, nil
	case "interval":
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
					GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
						Code: `
								const date = new Date();
								const hours = String(date.getUTCHours()).padStart(2, '0');
								const minutes = String(date.getUTCMinutes()).padStart(2, '0');
								const seconds = String(date.getUTCSeconds()).padStart(2, '0');
								return hours + ":" + minutes + ":" + seconds;
							`,
					},
				},
			},
		}, nil
	case "timestamp without time zone":
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
	case "timestamp with time zone":
		return &mgmtv1alpha1.JobMappingTransformer{
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
					GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
						Code: `
								const date = new Date();
								const year = date.getUTCFullYear();
								const month = String(date.getUTCMonth() + 1).padStart(2, '0');
								const day = String(date.getUTCDate()).padStart(2, '0');
								const hours = String(date.getUTCHours()).padStart(2, '0');
								const minutes = String(date.getUTCMinutes()).padStart(2, '0');
								const seconds = String(date.getUTCSeconds()).padStart(2, '0');
								const timezoneOffset = -date.getTimezoneOffset();
								const absOffset = Math.abs(timezoneOffset);
								const offsetHours = String(Math.floor(absOffset / 60)).padStart(2, '0');
								const offsetMinutes = String(absOffset % 60).padStart(2, '0');
								const offsetSign = timezoneOffset >= 0 ? '+' : '-';
								return year + "-" + month + "-" + day + " " + hours + ":" + minutes + ":" + seconds + offsetSign + offsetHours + ":" + offsetMinutes;
							`,
					},
				},
			},
		}, nil
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

func getSqlBatchProcessors(
	driver string,
	columns []string,
	columnDataTypes map[string]string,
	columnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties,
) (*neosync_benthos.BatchProcessor, error) {
	switch driver {
	case sqlmanager_shared.PostgresDriver:
		return &neosync_benthos.BatchProcessor{
			NeosyncToPgx: &neosync_benthos.NeosyncToPgxConfig{
				Columns:                 columns,
				ColumnDataTypes:         columnDataTypes,
				ColumnDefaultProperties: columnDefaultProperties,
			},
		}, nil
	case sqlmanager_shared.MysqlDriver:
		return &neosync_benthos.BatchProcessor{
			NeosyncToMysql: &neosync_benthos.NeosyncToMysqlConfig{
				Columns:                 columns,
				ColumnDataTypes:         columnDataTypes,
				ColumnDefaultProperties: columnDefaultProperties,
			},
		}, nil
	case sqlmanager_shared.MssqlDriver:
		return &neosync_benthos.BatchProcessor{
			NeosyncToMssql: &neosync_benthos.NeosyncToMssqlConfig{
				Columns:                 columns,
				ColumnDataTypes:         columnDataTypes,
				ColumnDefaultProperties: columnDefaultProperties,
			},
		}, nil
	default:
		return nil, fmt.Errorf(
			"unsupported driver %q when attempting to get sql batch processors",
			driver,
		)
	}
}
