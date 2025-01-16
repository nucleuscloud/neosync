package job

import (
	"fmt"
	"slices"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	"github.com/nucleuscloud/neosync/backend/pkg/utils"
)

type JobMappingsValidator struct {
	databaseErrors []string
	columnErrors   map[string]map[string][]string // schema.table -> column -> errors
	columnWarnings map[string]map[string][]string // schema.table -> column -> errors

	jobSourceOptions *SqlJobSourceOpts
	jobMappings      map[string]map[string]*mgmtv1alpha1.JobMapping // schema.table -> column -> job mapping
}

type JobMappingsValidatorResponse struct {
	DatabaseErrors []string
	ColumnErrors   map[string]map[string][]string
	ColumnWarnings map[string]map[string][]string
}

type Option func(*JobMappingsValidator)

func WithJobSourceOptions(jobSourceOptions *SqlJobSourceOpts) Option {
	return func(jmv *JobMappingsValidator) {
		jmv.jobSourceOptions = jobSourceOptions
	}
}

func NewJobMappingsValidator(jobMappings []*mgmtv1alpha1.JobMapping, opts ...Option) *JobMappingsValidator {
	tableToColumnMappings := map[string]map[string]*mgmtv1alpha1.JobMapping{}
	for _, m := range jobMappings {
		tn := sqlmanager_shared.BuildTable(m.Schema, m.Table)
		if _, ok := tableToColumnMappings[tn]; !ok {
			tableToColumnMappings[tn] = map[string]*mgmtv1alpha1.JobMapping{}
		}
		tableToColumnMappings[tn][m.Column] = m
	}

	jmv := &JobMappingsValidator{
		jobMappings:      tableToColumnMappings,
		databaseErrors:   []string{},
		columnErrors:     map[string]map[string][]string{},
		columnWarnings:   map[string]map[string][]string{},
		jobSourceOptions: &SqlJobSourceOpts{},
	}

	for _, opt := range opts {
		opt(jmv)
	}
	return jmv
}

func (j *JobMappingsValidator) GetDatabaseErrors() []string {
	return j.databaseErrors
}

func (j *JobMappingsValidator) GetColumnErrors() map[string]map[string][]string {
	return j.columnErrors
}

func (j *JobMappingsValidator) GetColumnWarnings() map[string]map[string][]string {
	return j.columnWarnings
}

func (j *JobMappingsValidator) addDatabaseError(err string) {
	j.databaseErrors = append(j.databaseErrors, err)
}

func (j *JobMappingsValidator) addColumnError(table, column, err string) {
	if _, ok := j.columnErrors[table]; !ok {
		j.columnErrors[table] = map[string][]string{}
	}
	j.columnErrors[table][column] = append(j.columnErrors[table][column], err)
}

func (j *JobMappingsValidator) addColumnWarning(table, column, err string) {
	if _, ok := j.columnWarnings[table]; !ok {
		j.columnWarnings[table] = map[string][]string{}
	}
	j.columnWarnings[table][column] = append(j.columnWarnings[table][column], err)
}

func (j *JobMappingsValidator) Validate(
	tableColumnMap map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
	virtualForeignKeys []*mgmtv1alpha1.VirtualForeignConstraint,
	tableConstraints *sqlmanager_shared.TableConstraints,
) (*JobMappingsValidatorResponse, error) {
	j.ValidateJobMappingsExistInSource(tableColumnMap)
	j.ValidateVirtualForeignKeys(virtualForeignKeys, tableColumnMap, tableConstraints)
	err := j.ValidateCircularDependencies(tableConstraints.ForeignKeyConstraints, virtualForeignKeys, tableColumnMap)
	if err != nil {
		return nil, err
	}
	j.ValidateRequiredForeignKeys(tableConstraints.ForeignKeyConstraints)
	j.ValidateRequiredColumns(tableColumnMap)
	return &JobMappingsValidatorResponse{
		DatabaseErrors: j.databaseErrors,
		ColumnErrors:   j.columnErrors,
		ColumnWarnings: j.columnWarnings,
	}, nil
}

// validate that all tables and columns in job mappings exist in source
func (j *JobMappingsValidator) ValidateJobMappingsExistInSource(
	tableColumnMap map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
) {
	// check for job mappings that do not exist in the source
	for table, colMappings := range j.jobMappings {
		if _, ok := tableColumnMap[table]; !ok {
			j.addDatabaseError(fmt.Sprintf("Table does not exist [%s] in source", table))
			continue
		}
		for col := range colMappings {
			if _, ok := tableColumnMap[table][col]; !ok {
				msg := fmt.Sprintf("Column does not exist in source. Remove column from job mappings: %s.%s", table, col)
				if j.jobSourceOptions != nil && !j.jobSourceOptions.HaltOnColumnRemoval {
					j.addColumnWarning(table, col, msg)
				} else {
					j.addColumnError(table, col, msg)
				}
			}
		}
	}

	// check for source columns that do not exist in job mappings
	for table, colMap := range tableColumnMap {
		if _, ok := j.jobMappings[table]; !ok {
			continue
		}
		for col := range colMap {
			if _, ok := j.jobMappings[table][col]; !ok {
				msg := fmt.Sprintf("Column does not exist in job mappings. Add column to job mappings: %s.%s", table, col)
				if j.jobSourceOptions != nil && !j.jobSourceOptions.HaltOnNewColumnAddition {
					j.addColumnWarning(table, col, msg)
				} else {
					j.addColumnError(table, col, msg)
				}
			}
		}
	}
}

// validates that there are no unsupported circular dependencies
func (j *JobMappingsValidator) ValidateCircularDependencies(
	foreignKeys map[string][]*sqlmanager_shared.ForeignConstraint,
	virtualForeignKeys []*mgmtv1alpha1.VirtualForeignConstraint,
	tableColumnMap map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
) error {
	// foreign key dependencies that are in job mappings
	validForeignKeyDependencies := map[string][]string{} // table -> foreign key table
	for table, fks := range foreignKeys {
		colMappings, ok := j.jobMappings[table]
		if !ok {
			// skip. table not in mapping
			continue
		}
		for _, fk := range fks {
			for idx, col := range fk.Columns {
				if _, ok := colMappings[col]; ok {
					fkColMappings, ok := j.jobMappings[fk.ForeignKey.Table]
					if ok {
						fkCol := fk.ForeignKey.Columns[idx]
						if _, ok = fkColMappings[fkCol]; ok {
							validForeignKeyDependencies[table] = append(validForeignKeyDependencies[table], fk.ForeignKey.Table)
						}
					}
				}
			}
		}
	}

	// merge virtual foreign keys with table foreign keys
	allForeignKeys := foreignKeys
	for _, vfk := range virtualForeignKeys {
		tableName := sqlmanager_shared.BuildTable(vfk.Schema, vfk.Table)
		fkTable := sqlmanager_shared.BuildTable(vfk.ForeignKey.Schema, vfk.ForeignKey.Table)
		validForeignKeyDependencies[tableName] = append(validForeignKeyDependencies[tableName], fkTable)

		tableCols, ok := tableColumnMap[tableName]
		if !ok {
			continue
		}
		notNullable := []bool{}
		for _, col := range vfk.GetColumns() {
			colInfo, ok := tableCols[col]
			if !ok {
				j.addColumnError(tableName, col, "Column does not exist in source but required by virtual foreign key")
				return nil
			}
			notNullable = append(notNullable, !colInfo.IsNullable)
		}

		allForeignKeys[tableName] = append(allForeignKeys[tableName], &sqlmanager_shared.ForeignConstraint{
			Columns:     vfk.GetColumns(),
			NotNullable: notNullable,
			ForeignKey: &sqlmanager_shared.ForeignKey{
				Columns: vfk.GetColumns(),
				Table:   fkTable,
			},
		})
	}

	for table, deps := range validForeignKeyDependencies {
		validForeignKeyDependencies[table] = utils.DedupeSlice(deps)
	}

	cycles := tabledependency.FindCircularDependencies(validForeignKeyDependencies)
	startTables, err := tabledependency.DetermineCycleStarts(cycles, map[string]string{}, allForeignKeys)
	if err != nil {
		return err
	}

	containsStart := func(t string) bool {
		return slices.Contains(startTables, t)
	}

	for _, cycle := range cycles {
		if !slices.ContainsFunc(cycle, containsStart) {
			j.addDatabaseError(fmt.Sprintf("Unsupported circular dependency. At least one foreign key in circular dependency must be nullable. Tables: %+v", cycle))
		}
	}
	return nil
}

// validate that all required primary keys are present in job mappings given foreign keys
func (j *JobMappingsValidator) ValidateRequiredForeignKeys(
	foreignkeys map[string][]*sqlmanager_shared.ForeignConstraint,
) {
	for table, fks := range foreignkeys {
		_, ok := j.jobMappings[table]
		if !ok {
			// skip. table not in mapping
			continue
		}
		for _, fk := range fks {
			for idx, notNull := range fk.NotNullable {
				if !notNull {
					// skip. foreign key is nullable
					continue
				}
				fkColMappings, ok := j.jobMappings[fk.ForeignKey.Table]
				fkCol := fk.ForeignKey.Columns[idx]
				if !ok {
					j.addColumnError(fk.ForeignKey.Table, fkCol, fmt.Sprintf("Missing required foreign key. Table: %s  Column: %s", fk.ForeignKey.Table, fkCol))
					continue
				}
				_, ok = fkColMappings[fkCol]
				if !ok {
					j.addColumnError(fk.ForeignKey.Table, fkCol, fmt.Sprintf("Missing required foreign key. Table: %s  Column: %s", fk.ForeignKey.Table, fkCol))
				}
			}
		}
	}
}

// validates that all non-nullable columns are included in the job mappings for each table
func (j *JobMappingsValidator) ValidateRequiredColumns(
	tableColumnMap map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
) {
	for table, colMap := range tableColumnMap {
		cm, ok := j.jobMappings[table]
		if !ok {
			// skip. table not in mapping
			continue
		}
		for col, info := range colMap {
			if info.IsNullable {
				// skip. column is nullable
				continue
			}
			if _, ok := cm[col]; !ok {
				j.addColumnError(table, col, fmt.Sprintf("Violates not-null constraint. Missing required column. Table: %s  Column: %s", table, col))
			}
		}
	}
}

// validates that all virtual foreign keys are valid
func (j *JobMappingsValidator) ValidateVirtualForeignKeys(
	virtualForeignKeys []*mgmtv1alpha1.VirtualForeignConstraint,
	tableColumnMap map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow,
	tableConstraints *sqlmanager_shared.TableConstraints,
) {
	for _, vfk := range virtualForeignKeys {
		sourceTable := sqlmanager_shared.BuildTable(vfk.ForeignKey.Schema, vfk.ForeignKey.Table)
		targetTable := sqlmanager_shared.BuildTable(vfk.Schema, vfk.Table)

		// check that source table exist in job mappings
		sourceColMappings, ok := j.jobMappings[sourceTable]
		if !ok {
			j.addDatabaseError(fmt.Sprintf("Virtual foreign key source table missing in job mappings. Table: %s", sourceTable))
			continue
		}
		sourceCols, ok := tableColumnMap[sourceTable]
		if !ok {
			j.addDatabaseError(fmt.Sprintf("Virtual foreign key source table missing in source database. Table: %s", sourceTable))
			return
		}

		// check that target table exist in job mappings
		targetColMappings, ok := j.jobMappings[targetTable]
		if !ok {
			j.addDatabaseError(fmt.Sprintf("Virtual foreign key target table missing in job mappings. Table: %s", targetTable))
			continue
		}
		targetCols, ok := tableColumnMap[targetTable]
		if !ok {
			j.addDatabaseError(fmt.Sprintf("Virtual foreign key target table missing in source database. Table: %s", targetTable))
			continue
		}

		j.validateVfkTableColumnsExistInSource(sourceTable, vfk, sourceColMappings, sourceCols)
		j.validateVfkSourceColumnHasConstraint(sourceTable, vfk, tableConstraints)
		j.validateCircularVfk(sourceTable, targetTable, vfk, targetColMappings, targetCols)

		if len(vfk.GetColumns()) != len(vfk.GetForeignKey().GetColumns()) {
			j.addDatabaseError(fmt.Sprintf("length of source columns was not equal to length of foreign key cols: %d %d. SourceTable: %s SourceColumn: %+v TargetTable: %s  TargetColumn: %+v", len(vfk.GetColumns()), len(vfk.GetForeignKey().GetColumns()), sourceTable, vfk.GetColumns(), targetTable, vfk.GetForeignKey().GetColumns()))
			continue
		}

		// check that source and target column datatypes are the same
		for idx, srcCol := range vfk.GetForeignKey().GetColumns() {
			tarCol := vfk.GetColumns()[idx]
			srcColInfo, srcColOk := sourceCols[srcCol]
			tarColInfo, tarColOk := targetCols[tarCol]
			if !srcColOk || !tarColOk {
				continue
			}
			if srcColInfo.DataType != tarColInfo.DataType {
				j.addColumnError(targetTable, tarCol, fmt.Sprintf("Column datatype mismatch. Source: %s.%s %s Target: %s.%s %s", sourceTable, srcCol, srcColInfo.DataType, targetTable, tarCol, tarColInfo.DataType))
			}
		}
	}
}

// validate that all columns in the virtual foreign key exist in the source database and job mappings
func (j *JobMappingsValidator) validateVfkTableColumnsExistInSource(
	table string,
	vfk *mgmtv1alpha1.VirtualForeignConstraint,
	colMappings map[string]*mgmtv1alpha1.JobMapping,
	sourceCols map[string]*sqlmanager_shared.DatabaseSchemaRow,
) {
	for _, c := range vfk.GetForeignKey().GetColumns() {
		_, ok := colMappings[c]
		if !ok {
			j.addColumnError(table, c, fmt.Sprintf("Virtual foreign key source column missing in job mappings. Table: %s Column: %s", table, c))
		}
		_, ok = sourceCols[c]
		if !ok {
			j.addColumnError(table, c, fmt.Sprintf("Virtual foreign key source column missing in source database. Table: %s Column: %s", table, c))
		}
	}
}

// validates that all sources of virtual foreign keys are either a primary key or have a unique constraint
func (j *JobMappingsValidator) validateVfkSourceColumnHasConstraint(
	table string,
	vfk *mgmtv1alpha1.VirtualForeignConstraint,
	tableConstraints *sqlmanager_shared.TableConstraints,
) {
	pks := tableConstraints.PrimaryKeyConstraints[table]
	uniqueConstraints := tableConstraints.UniqueConstraints[table]
	isVfkValid := isVirtualForeignKeySourceUnique(vfk, pks, uniqueConstraints)
	if !isVfkValid {
		for _, c := range vfk.GetForeignKey().GetColumns() {
			j.addColumnError(table, c, fmt.Sprintf("Virtual foreign key source must be either a primary key or have a unique constraint. Table: %s  Columns: %+v", table, vfk.GetForeignKey().GetColumns()))
		}
	}
}

// validates that all self referencing virtual foreign keys are on nullable columns
func (j *JobMappingsValidator) validateCircularVfk(
	sourceTable, targetTable string,
	vfk *mgmtv1alpha1.VirtualForeignConstraint,
	targetColMappings map[string]*mgmtv1alpha1.JobMapping,
	targetCols map[string]*sqlmanager_shared.DatabaseSchemaRow,
) {
	for _, c := range vfk.GetColumns() {
		_, ok := targetColMappings[c]
		if !ok {
			j.addColumnError(targetTable, c, fmt.Sprintf("Virtual foreign key target column missing in job mappings. Table: %s Column: %s", targetTable, c))
		}
		colInfo, ok := targetCols[c]
		if !ok {
			j.addColumnError(targetTable, c, fmt.Sprintf("Virtual foreign key target column missing in source database. Table: %s Column: %s", targetTable, c))
			continue
		}
		if sourceTable == targetTable && !colInfo.IsNullable {
			j.addColumnError(targetTable, c, fmt.Sprintf("Self referencing virtual foreign key target column must be nullable. Table: %s  Column: %s", targetTable, c))
		}
	}
}

func isVirtualForeignKeySourceUnique(
	virtualForeignKey *mgmtv1alpha1.VirtualForeignConstraint,
	primaryKeys []string,
	uniqueConstraints [][]string,
) bool {
	if utils.CompareSlices(virtualForeignKey.GetForeignKey().GetColumns(), primaryKeys) {
		return true
	}
	for _, uc := range uniqueConstraints {
		if utils.CompareSlices(virtualForeignKey.GetForeignKey().GetColumns(), uc) {
			return true
		}
	}
	return false
}
