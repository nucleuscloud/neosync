package benthosbuilder_builders

import (
	"fmt"
	"log/slog"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	job_util "github.com/nucleuscloud/neosync/internal/job"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

type sqlSourceTableOptions struct {
	WhereClause *string
}

type tableMapping struct {
	Schema   string
	Table    string
	Mappings []*shared.JobTransformationMapping
}

func getMapValuesCount[K comparable, V any](m map[K][]V) int {
	count := 0
	for _, v := range m {
		count += len(v)
	}
	return count
}

func buildPlainColumns(mappings []*shared.JobTransformationMapping) []string {
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
	mappings []*shared.JobTransformationMapping,
) []*tableMapping {
	groupedMappings := map[string][]*shared.JobTransformationMapping{}

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
