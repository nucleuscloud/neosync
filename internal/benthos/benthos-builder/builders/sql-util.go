package benthosbuilder_builders

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

const (
	jobmappingSubsetErrMsg     = "unable to continue: job mappings contain schemas, tables, or columns that were not found in the source connection"
	haltOnSchemaAdditionErrMsg = "unable to continue: HaltOnNewColumnAddition: job mappings are missing columns for the mapped tables found in the source connection"
)

type sqlJobSourceOpts struct {
	// Determines if the job should halt if a new column is detected that is not present in the job mappings
	HaltOnNewColumnAddition bool
	// Newly detected columns are automatically transformed
	GenerateNewColumnTransformers bool
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

type sqlSourceTableOptions struct {
	WhereClause *string
}

type tableMapping struct {
	Schema   string
	Table    string
	Mappings []*mgmtv1alpha1.JobMapping
}

// Based on the source schema and the provided job mappings, the job mappings must be at least a subset of the source schema
// Otherwise, the sync is doomed for failure
func areMappingsSubsetOfSchemas(
	groupedSchemas map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
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

func buildTableSubsetMap(tableOpts map[string]*sqlSourceTableOptions, tableMap map[string]*tableMapping) map[string]string {
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
	sqlSourceOpts *sqlJobSourceOpts,
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

func getColumnTransformerMap(tableMappingMap map[string]*tableMapping) map[string]map[string]*mgmtv1alpha1.JobMappingTransformer {
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
				newFk.ForeignKey.Columns = append(newFk.ForeignKey.Columns, fk.ForeignKey.Columns[i])
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
func getPrimaryKeyDependencyMap(tableDependencies map[string][]*sqlmanager_shared.ForeignConstraint) map[string]map[string][]*bb_internal.ReferenceKey {
	tc := map[string]map[string][]*bb_internal.ReferenceKey{} // schema.table -> column -> ForeignKey
	for table, constraints := range tableDependencies {
		for _, c := range constraints {
			_, ok := tc[c.ForeignKey.Table]
			if !ok {
				tc[c.ForeignKey.Table] = map[string][]*bb_internal.ReferenceKey{}
			}
			for idx, col := range c.ForeignKey.Columns {
				tc[c.ForeignKey.Table][col] = append(tc[c.ForeignKey.Table][col], &bb_internal.ReferenceKey{
					Table:  table,
					Column: c.Columns[idx],
				})
			}
		}
	}
	return tc
}

func findTopForeignKeySource(tableName, col string, tableDependencies map[string][]*sqlmanager_shared.ForeignConstraint) *bb_internal.ReferenceKey {
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
	return &bb_internal.ReferenceKey{
		Table:  tableName,
		Column: col,
	}
}

// builds schema.table -> FK column ->  PK schema table column
// find top level primary key column if foreign keys are nested
func buildForeignKeySourceMap(tableDeps map[string][]*sqlmanager_shared.ForeignConstraint) map[string]map[string]*bb_internal.ReferenceKey {
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
				transformedForeignKeyToSourceMap[table][col] = append(transformedForeignKeyToSourceMap[table][col], tc)
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
		needsOverride, needsReset, err := sqlmanager.GetColumnOverrideAndResetProperties(driver, info)
		if err != nil {
			slogger.Error("unable to determine SQL column default flags", "error", err, "column", cName)
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

func buildRedisDependsOnMap(transformedForeignKeyToSourceMap map[string][]*bb_internal.ReferenceKey, runconfig *tabledependency.RunConfig) map[string][]string {
	redisDependsOnMap := map[string][]string{}
	for col, fks := range transformedForeignKeyToSourceMap {
		if !slices.Contains(runconfig.InsertColumns(), col) {
			continue
		}
		for _, fk := range fks {
			if _, exists := redisDependsOnMap[fk.Table]; !exists {
				redisDependsOnMap[fk.Table] = []string{}
			}
			redisDependsOnMap[fk.Table] = append(redisDependsOnMap[fk.Table], fk.Column)
		}
	}
	if runconfig.RunType() == tabledependency.RunTypeUpdate && len(redisDependsOnMap) != 0 {
		redisDependsOnMap[runconfig.Table()] = runconfig.PrimaryKeys()
	}
	return redisDependsOnMap
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
		shouldHalt := false
		shouldGenerateNewColTransforms := false
		switch jobSourceConfig.Postgres.GetNewColumnAdditionStrategy().GetStrategy().(type) {
		case *mgmtv1alpha1.PostgresSourceConnectionOptions_NewColumnAdditionStrategy_HaltJob_:
			shouldHalt = true
		case *mgmtv1alpha1.PostgresSourceConnectionOptions_NewColumnAdditionStrategy_AutoMap_:
			shouldGenerateNewColTransforms = true
		}
		// deprecated fallback if no strategy has been defined
		if !shouldHalt && !shouldGenerateNewColTransforms {
			shouldHalt = jobSourceConfig.Postgres.GetHaltOnNewColumnAddition()
		}

		return &sqlJobSourceOpts{
			HaltOnNewColumnAddition:       shouldHalt,
			GenerateNewColumnTransformers: shouldGenerateNewColTransforms,
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
	case *mgmtv1alpha1.JobSourceOptions_Mssql:
		if jobSourceConfig.Mssql == nil {
			return nil, nil
		}
		schemaOpt := []*schemaOptions{}
		for _, opt := range jobSourceConfig.Mssql.Schemas {
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
			HaltOnNewColumnAddition:       jobSourceConfig.Mssql.HaltOnNewColumnAddition,
			SubsetByForeignKeyConstraints: jobSourceConfig.Mssql.SubsetByForeignKeyConstraints,
			SchemaOpt:                     schemaOpt,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported job source options type for sql job source: %T", jobSourceConfig)
	}
}

type destinationOptions struct {
	OnConflictDoNothing      bool
	Truncate                 bool
	TruncateCascade          bool
	SkipForeignKeyViolations bool
	MaxInFlight              uint32
	BatchCount               int
	BatchPeriod              string
}

func getDestinationOptions(destOpts *mgmtv1alpha1.JobDestinationOptions) (*destinationOptions, error) {
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
		return &destinationOptions{
			OnConflictDoNothing:      config.PostgresOptions.GetOnConflict().GetDoNothing(),
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
		return &destinationOptions{
			OnConflictDoNothing:      config.MysqlOptions.GetOnConflict().GetDoNothing(),
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
		MaxInFlight: 64,
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
				return batchingConfig{}, fmt.Errorf("unable to parse batch period for s3 destination config: %w", err)
			}
		}
		output.BatchPeriod = batchConfig.GetPeriod()
	}

	if output.BatchCount == 0 && output.BatchPeriod == "" {
		return batchingConfig{}, fmt.Errorf("must have at least one batch policy configured. Cannot disable both period and count")
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
			logger.Warn("table found in schema data that is not present in job mappings", "table", schematable)
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
				if info.ColumnDefault != "" || info.IdentityGeneration != nil || info.GeneratedType != nil {
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
func getJmTransformerByPostgresDataType(colInfo *sqlmanager_shared.DatabaseSchemaRow) (*mgmtv1alpha1.JobMappingTransformer, error) {
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
	case "text", "bpchar", "character", "character varying": // todo: test to see if this works when (n) has been specified
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
					GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{IncludeHyphens: shared.Ptr(true)},
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("uncountered unsupported data type %q for %q.%q.%q when attempting to generate an auto-mapper. To continue, provide a discrete job mapping for this column.: %w",
			colInfo.DataType, colInfo.TableSchema, colInfo.TableName, colInfo.ColumnName, errors.ErrUnsupported,
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

func shouldOverrideColumnDefault(columnDefaults map[string]*neosync_benthos.ColumnDefaultProperties) bool {
	for _, cd := range columnDefaults {
		if cd != nil && !cd.HasDefaultTransformer && cd.NeedsOverride {
			return true
		}
	}
	return false
}

func getSqlBatchProcessors(driver string, columns []string, columnDataTypes map[string]string, columnDefaultProperties map[string]*neosync_benthos.ColumnDefaultProperties) (*neosync_benthos.BatchProcessor, error) {
	switch driver {
	case sqlmanager_shared.PostgresDriver:
		return &neosync_benthos.BatchProcessor{NeosyncToPgx: &neosync_benthos.NeosyncToPgxConfig{Columns: columns, ColumnDataTypes: columnDataTypes, ColumnDefaultProperties: columnDefaultProperties}}, nil
	case sqlmanager_shared.MysqlDriver:
		return &neosync_benthos.BatchProcessor{NeosyncToMysql: &neosync_benthos.NeosyncToMysqlConfig{Columns: columns, ColumnDataTypes: columnDataTypes, ColumnDefaultProperties: columnDefaultProperties}}, nil
	case sqlmanager_shared.MssqlDriver:
		return &neosync_benthos.BatchProcessor{NeosyncToMssql: &neosync_benthos.NeosyncToMssqlConfig{Columns: columns, ColumnDataTypes: columnDataTypes, ColumnDefaultProperties: columnDefaultProperties}}, nil
	default:
		return nil, fmt.Errorf("unsupported driver %q when attempting to get sql batch processors", driver)
	}
}
