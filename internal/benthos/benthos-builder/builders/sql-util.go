package benthosbuilder_builders

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
)

const (
	jobmappingSubsetErrMsg     = "job mappings are not equal to or a subset of the database schema found in the source connection"
	haltOnSchemaAdditionErrMsg = "job mappings does not contain a column mapping for all " +
		"columns found in the source connection for the selected schemas and tables"
)

type sqlJobSourceOpts struct {
	HaltOnNewColumnAddition       bool
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

func areMappingsSubsetOfSchemas(
	groupedSchemas map[string]map[string]*sqlmanager_shared.ColumnInfo,
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

func shouldHaltOnSchemaAddition(
	groupedSchemas map[string]map[string]*sqlmanager_shared.ColumnInfo,
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

func buildPlainInsertArgs(cols []string) string {
	if len(cols) == 0 {
		return ""
	}
	pieces := make([]string, len(cols))
	for idx := range cols {
		pieces[idx] = fmt.Sprintf("this.%q", cols[idx])
	}
	return fmt.Sprintf("root = [%s]", strings.Join(pieces, ", "))
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
	colInfoMap map[string]map[string]*sqlmanager_shared.ColumnInfo,
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
				if !fk.NotNullable[i] && (!ok || t.GetSource() == mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL) {
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
	colInfo map[string]*sqlmanager_shared.ColumnInfo,
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

		transformer, ok := colTransformers[cName]
		if !ok {
			return nil, fmt.Errorf("transformer missing for column: %s", cName)
		}
		var hasDefaultTransformer bool
		if transformer != nil && transformer.Source == mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT {
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
		return &sqlJobSourceOpts{
			HaltOnNewColumnAddition:       jobSourceConfig.Postgres.HaltOnNewColumnAddition,
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
