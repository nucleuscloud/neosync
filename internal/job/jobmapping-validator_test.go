package job

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateJobMappingsExistInSource(t *testing.T) {
	t.Run("should return database error when mapping schema doesn't exist", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{
				Schema: "schema1",
				Table:  "table1",
				Column: "col1",
			},
		}

		sourceCols := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema2.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{},
			},
		}

		jmv := NewJobMappingsValidator(mappings, WithJobSourceOptions(&SqlJobSourceOpts{
			HaltOnNewColumnAddition: false,
			HaltOnColumnRemoval:     false,
		}))
		jmv.ValidateJobMappingsExistInSource(sourceCols)
		assert.Equal(t, []*mgmtv1alpha1.TableError_TableErrorReport{
			{
				Code:    mgmtv1alpha1.TableError_TABLE_ERROR_CODE_TABLE_NOT_FOUND_IN_SOURCE,
				Message: "Table does not exist [schema1.table1] in source",
			},
		}, jmv.GetTableErrors()["schema1.table1"])
	})

	t.Run("should return column errors", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{
				Schema: "schema1",
				Table:  "table1",
				Column: "col1",
			},
			{
				Schema: "schema1",
				Table:  "table1",
				Column: "col2",
			},
		}

		sourceCols := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{},
				"col3": &sqlmanager_shared.DatabaseSchemaRow{},
			},
			"schema2.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{},
				"col3": &sqlmanager_shared.DatabaseSchemaRow{},
			},
		}

		jmv := NewJobMappingsValidator(mappings, WithJobSourceOptions(&SqlJobSourceOpts{
			HaltOnNewColumnAddition: true,
			HaltOnColumnRemoval:     true,
		}))
		jmv.ValidateJobMappingsExistInSource(sourceCols)

		warnings := jmv.GetColumnWarnings()
		assert.Empty(t, warnings)
		errs := jmv.GetColumnErrors()
		require.NotEmpty(t, errs)
		assert.Equal(t, []*mgmtv1alpha1.ColumnError_ColumnErrorReport{
			{
				Code:    mgmtv1alpha1.ColumnError_COLUMN_ERROR_CODE_NOT_FOUND_IN_SOURCE,
				Message: "Column does not exist in source. Remove column from job mappings: schema1.table1.col2",
			},
		}, errs["schema1.table1"]["col2"])
		assert.Equal(t, []*mgmtv1alpha1.ColumnError_ColumnErrorReport{
			{
				Code:    mgmtv1alpha1.ColumnError_COLUMN_ERROR_CODE_NOT_FOUND_IN_MAPPING,
				Message: "Column does not exist in job mappings. Add column to job mappings: schema1.table1.col3",
			},
		}, errs["schema1.table1"]["col3"])
	})

	t.Run("should return column warnings", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{
				Schema: "schema1",
				Table:  "table1",
				Column: "col1",
			},
			{
				Schema: "schema1",
				Table:  "table1",
				Column: "col2",
			},
		}

		sourceCols := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{},
				"col3": &sqlmanager_shared.DatabaseSchemaRow{},
			},
			"schema2.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{},
				"col3": &sqlmanager_shared.DatabaseSchemaRow{},
			},
		}

		jmv := NewJobMappingsValidator(mappings, WithJobSourceOptions(&SqlJobSourceOpts{
			HaltOnNewColumnAddition: false,
			HaltOnColumnRemoval:     false,
		}))
		jmv.ValidateJobMappingsExistInSource(sourceCols)

		errs := jmv.GetColumnErrors()
		assert.Empty(t, errs)
		warnings := jmv.GetColumnWarnings()
		require.NotEmpty(t, warnings)
		assert.Equal(t, []*mgmtv1alpha1.ColumnWarning_ColumnWarningReport{
			{
				Code:    mgmtv1alpha1.ColumnWarning_COLUMN_WARNING_CODE_NOT_FOUND_IN_SOURCE,
				Message: "Column does not exist in source. Remove column from job mappings: schema1.table1.col2",
			},
		}, warnings["schema1.table1"]["col2"])
		assert.Equal(t, []*mgmtv1alpha1.ColumnWarning_ColumnWarningReport{
			{
				Code:    mgmtv1alpha1.ColumnWarning_COLUMN_WARNING_CODE_NOT_FOUND_IN_MAPPING,
				Message: "Column does not exist in job mappings. Add column to job mappings: schema1.table1.col3",
			},
		}, warnings["schema1.table1"]["col3"])
	})
}

func TestJobMappingsValidator_ValidateRequiredForeignKeys(t *testing.T) {
	t.Run("should return error when required foreign key table missing", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{
				Schema: "schema1",
				Table:  "table1",
				Column: "col1",
			},
		}

		foreignKeys := map[string][]*sqlmanager_shared.ForeignConstraint{
			"schema1.table1": {
				{
					Columns:     []string{"col1"},
					NotNullable: []bool{true},
					ForeignKey: &sqlmanager_shared.ForeignKey{
						Table:   "schema1.table2",
						Columns: []string{"id"},
					},
				},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateRequiredForeignKeys(foreignKeys)

		errs := jmv.GetColumnErrors()
		require.NotEmpty(t, errs)
		assert.Equal(t, []*mgmtv1alpha1.ColumnError_ColumnErrorReport{
			{
				Code:    mgmtv1alpha1.ColumnError_COLUMN_ERROR_CODE_REQUIRED_FOREIGN_KEY_NOT_FOUND_IN_MAPPING,
				Message: "Missing required foreign key. Table: schema1.table2  Column: id",
			},
		}, errs["schema1.table2"]["id"])
	})

	t.Run("should return error when required foreign key column missing", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{
				Schema: "schema1",
				Table:  "table1",
				Column: "col1",
			},
			{
				Schema: "schema1",
				Table:  "table2",
				Column: "col1",
			},
		}

		foreignKeys := map[string][]*sqlmanager_shared.ForeignConstraint{
			"schema1.table1": {
				{
					Columns:     []string{"col1"},
					NotNullable: []bool{true},
					ForeignKey: &sqlmanager_shared.ForeignKey{
						Table:   "schema1.table2",
						Columns: []string{"id"},
					},
				},
			},
			"schema2.table2": {
				{
					Columns:     []string{"col1"},
					NotNullable: []bool{true},
					ForeignKey: &sqlmanager_shared.ForeignKey{
						Table:   "schema2.table3",
						Columns: []string{"id"},
					},
				},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateRequiredForeignKeys(foreignKeys)

		errs := jmv.GetColumnErrors()
		require.NotEmpty(t, errs)
		assert.Equal(t, []*mgmtv1alpha1.ColumnError_ColumnErrorReport{
			{
				Code:    mgmtv1alpha1.ColumnError_COLUMN_ERROR_CODE_REQUIRED_FOREIGN_KEY_NOT_FOUND_IN_MAPPING,
				Message: "Missing required foreign key. Table: schema1.table2  Column: id",
			},
		}, errs["schema1.table2"]["id"])
	})

	t.Run("should not return error when foreign key is nullable", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{
				Schema: "schema1",
				Table:  "table1",
				Column: "col1",
			},
		}

		foreignKeys := map[string][]*sqlmanager_shared.ForeignConstraint{
			"schema1.table1": {
				{
					Columns:     []string{"col1"},
					NotNullable: []bool{false},
					ForeignKey: &sqlmanager_shared.ForeignKey{
						Table:   "schema1.table2",
						Columns: []string{"id"},
					},
				},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateRequiredForeignKeys(foreignKeys)

		errs := jmv.GetColumnErrors()
		assert.Empty(t, errs)
	})

	t.Run("should not return error when required foreign key exists", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{
				Schema: "schema1",
				Table:  "table1",
				Column: "col1",
			},
			{
				Schema: "schema1",
				Table:  "table2",
				Column: "id",
			},
		}

		foreignKeys := map[string][]*sqlmanager_shared.ForeignConstraint{
			"schema1.table1": {
				{
					Columns:     []string{"col1"},
					NotNullable: []bool{true},
					ForeignKey: &sqlmanager_shared.ForeignKey{
						Table:   "schema1.table2",
						Columns: []string{"id"},
					},
				},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateRequiredForeignKeys(foreignKeys)

		errs := jmv.GetColumnErrors()
		assert.Empty(t, errs)
	})
}

func TestValidateRequiredColumns(t *testing.T) {
	t.Run("should return error when required column missing", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{
				Schema: "schema1",
				Table:  "table1",
				Column: "col1",
			},
		}

		sourceCols := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: true,
				},
				"col2": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false,
				},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateRequiredColumns(sourceCols)

		errs := jmv.GetColumnErrors()
		require.NotEmpty(t, errs)
		assert.Equal(t, []*mgmtv1alpha1.ColumnError_ColumnErrorReport{
			{
				Code:    mgmtv1alpha1.ColumnError_COLUMN_ERROR_CODE_REQUIRED_COLUMN_NOT_FOUND_IN_MAPPING,
				Message: "Violates not-null constraint. Missing required column. Table: schema1.table1  Column: col2",
			},
		}, errs["schema1.table1"]["col2"])
	})

	t.Run("should not return error when all required columns exist", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{
				Schema: "schema1",
				Table:  "table1",
				Column: "col1",
			},
			{
				Schema: "schema1",
				Table:  "table1",
				Column: "col2",
			},
		}

		sourceCols := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false,
				},
				"col2": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false,
				},
				"col3": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: true,
				},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateRequiredColumns(sourceCols)

		errs := jmv.GetColumnErrors()
		assert.Empty(t, errs)
	})

	t.Run("should skip table not in mappings", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{
				Schema: "schema1",
				Table:  "table1",
				Column: "col1",
			},
		}

		sourceCols := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false,
				},
			},
			"schema1.table2": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false,
				},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateRequiredColumns(sourceCols)

		errs := jmv.GetColumnErrors()
		assert.Empty(t, errs)
	})
}

func TestValidateCircularDependencies(t *testing.T) {
	t.Run("should return error when cycle has no nullable foreign keys", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
			{Schema: "schema1", Table: "table2", Column: "col1"},
		}

		foreignKeys := map[string][]*sqlmanager_shared.ForeignConstraint{
			"schema1.table1": {
				{
					Columns:     []string{"col1"},
					NotNullable: []bool{true},
					ForeignKey: &sqlmanager_shared.ForeignKey{
						Table:   "schema1.table2",
						Columns: []string{"col1"},
					},
				},
			},
			"schema1.table2": {
				{
					Columns:     []string{"col1"},
					NotNullable: []bool{true},
					ForeignKey: &sqlmanager_shared.ForeignKey{
						Table:   "schema1.table1",
						Columns: []string{"col1"},
					},
				},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		err := jmv.ValidateCircularDependencies(foreignKeys, []*mgmtv1alpha1.VirtualForeignConstraint{}, map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{})
		require.NoError(t, err)

		errs := jmv.GetDatabaseErrors()
		require.Len(t, errs, 1)
		require.Len(t, errs, 1)
		assert.Equal(t, mgmtv1alpha1.DatabaseError_DATABASE_ERROR_CODE_UNSUPPORTED_CIRCULAR_DEPENDENCY_AT_LEAST_ONE_NULLABLE, errs[0].Code)
		assert.Contains(t, errs[0].Message, "Unsupported circular dependency. At least one foreign key in circular dependency must be nullable")
	})

	t.Run("should not return error when cycle has nullable foreign key", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
			{Schema: "schema1", Table: "table2", Column: "col1"},
		}

		foreignKeys := map[string][]*sqlmanager_shared.ForeignConstraint{
			"schema1.table1": {
				{
					Columns:     []string{"col1"},
					NotNullable: []bool{false}, // Nullable foreign key
					ForeignKey: &sqlmanager_shared.ForeignKey{
						Table:   "schema1.table2",
						Columns: []string{"col1"},
					},
				},
			},
			"schema1.table2": {
				{
					Columns:     []string{"col1"},
					NotNullable: []bool{true},
					ForeignKey: &sqlmanager_shared.ForeignKey{
						Table:   "schema1.table1",
						Columns: []string{"col1"},
					},
				},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		err := jmv.ValidateCircularDependencies(foreignKeys, []*mgmtv1alpha1.VirtualForeignConstraint{}, map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{})
		require.NoError(t, err)

		errs := jmv.GetDatabaseErrors()
		assert.Empty(t, errs)
	})

	t.Run("should handle virtual foreign keys in cycle detection", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
			{Schema: "schema1", Table: "table2", Column: "col1"},
		}

		virtualForeignKeys := []*mgmtv1alpha1.VirtualForeignConstraint{
			{
				Schema:  "schema1",
				Table:   "table1",
				Columns: []string{"col1"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "schema1",
					Table:   "table2",
					Columns: []string{"col1"},
				},
			},
		}

		tableColumnMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false,
				},
			},
			"schema1.table2": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false,
				},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		err := jmv.ValidateCircularDependencies(map[string][]*sqlmanager_shared.ForeignConstraint{}, virtualForeignKeys, tableColumnMap)
		require.NoError(t, err)
	})

	t.Run("should detect unsupported circular dependency with virtual foreign keys", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
			{Schema: "schema1", Table: "table2", Column: "col1"},
			{Schema: "schema1", Table: "table3", Column: "col1"},
		}

		foreignKeys := map[string][]*sqlmanager_shared.ForeignConstraint{
			"schema1.table2": {
				{
					Columns:     []string{"col1"},
					NotNullable: []bool{true},
					ForeignKey: &sqlmanager_shared.ForeignKey{
						Table:   "schema1.table3",
						Columns: []string{"col1"},
					},
				},
			},
		}

		virtualForeignKeys := []*mgmtv1alpha1.VirtualForeignConstraint{
			{
				Schema:  "schema1",
				Table:   "table3",
				Columns: []string{"col1"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "schema1",
					Table:   "table1",
					Columns: []string{"col1"},
				},
			},
			{
				Schema:  "schema1",
				Table:   "table1",
				Columns: []string{"col1"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "schema1",
					Table:   "table2",
					Columns: []string{"col1"},
				},
			},
		}

		tableColumnMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false,
				},
			},
			"schema1.table2": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false,
				},
			},
			"schema1.table3": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false,
				},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		err := jmv.ValidateCircularDependencies(foreignKeys, virtualForeignKeys, tableColumnMap)
		require.NoError(t, err)

		errs := jmv.GetDatabaseErrors()
		require.NotEmpty(t, errs)
		require.Len(t, errs, 1)
		assert.Equal(t, mgmtv1alpha1.DatabaseError_DATABASE_ERROR_CODE_UNSUPPORTED_CIRCULAR_DEPENDENCY_AT_LEAST_ONE_NULLABLE, errs[0].Code)
		assert.Contains(t, errs[0].Message, "Unsupported circular dependency. At least one foreign key in circular dependency must be nullable")
	})

	t.Run("should skip tables not in mappings", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
		}

		foreignKeys := map[string][]*sqlmanager_shared.ForeignConstraint{
			"schema1.table2": {
				{
					Columns:     []string{"col1"},
					NotNullable: []bool{true},
					ForeignKey: &sqlmanager_shared.ForeignKey{
						Table:   "schema1.table3",
						Columns: []string{"col1"},
					},
				},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		err := jmv.ValidateCircularDependencies(foreignKeys, []*mgmtv1alpha1.VirtualForeignConstraint{}, map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{})
		require.NoError(t, err)

		errs := jmv.GetDatabaseErrors()
		assert.Empty(t, errs)
	})

	t.Run("should return error when virtual foreign key column does not exist in source", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
			{Schema: "schema1", Table: "table2", Column: "col1"},
		}

		virtualForeignKeys := []*mgmtv1alpha1.VirtualForeignConstraint{
			{
				Schema:  "schema1",
				Table:   "table1",
				Columns: []string{"col2"}, // Column that doesn't exist
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "schema1",
					Table:   "table2",
					Columns: []string{"col1"},
				},
			},
		}

		tableColumnMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false,
				},
			},
			"schema1.table2": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false,
				},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		err := jmv.ValidateCircularDependencies(map[string][]*sqlmanager_shared.ForeignConstraint{}, virtualForeignKeys, tableColumnMap)
		require.NoError(t, err)

		errs := jmv.GetColumnErrors()
		require.NotEmpty(t, errs)
		assert.Equal(t, []*mgmtv1alpha1.ColumnError_ColumnErrorReport{
			{
				Code:    mgmtv1alpha1.ColumnError_COLUMN_ERROR_CODE_VFK_SOURCE_COLUMN_NOT_FOUND_IN_SOURCE,
				Message: "Column does not exist in source but required by virtual foreign key",
			},
		}, errs["schema1.table1"]["col2"])
	})
}

func TestValidateVirtualForeignKeys(t *testing.T) {
	t.Run("should return error when source table missing in job mappings", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
		}

		virtualForeignKeys := []*mgmtv1alpha1.VirtualForeignConstraint{
			{
				Schema:  "schema1",
				Table:   "table1",
				Columns: []string{"col1"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "schema1",
					Table:   "table2",
					Columns: []string{"col1"},
				},
			},
		}

		tableColumnMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{},
			},
			"schema1.table2": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateVirtualForeignKeys(virtualForeignKeys, tableColumnMap, &sqlmanager_shared.TableConstraints{})

		errs := jmv.GetTableErrors()
		require.Len(t, errs["schema1.table2"], 1)
		assert.Equal(t, mgmtv1alpha1.TableError_TABLE_ERROR_CODE_VFK_SOURCE_TABLE_NOT_FOUND_IN_MAPPING, errs["schema1.table2"][0].Code)
		assert.Contains(t, errs["schema1.table2"][0].Message, "Virtual foreign key source table missing in job mappings")
	})

	t.Run("should return error when target table missing in job mappings", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table2", Column: "col1"},
		}

		virtualForeignKeys := []*mgmtv1alpha1.VirtualForeignConstraint{
			{
				Schema:  "schema1",
				Table:   "table1",
				Columns: []string{"col1"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "schema1",
					Table:   "table2",
					Columns: []string{"col1"},
				},
			},
		}

		tableColumnMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{},
			},
			"schema1.table2": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateVirtualForeignKeys(virtualForeignKeys, tableColumnMap, &sqlmanager_shared.TableConstraints{})

		errs := jmv.GetTableErrors()
		require.Len(t, errs["schema1.table1"], 1)
		assert.Equal(t, mgmtv1alpha1.TableError_TABLE_ERROR_CODE_VFK_TARGET_TABLE_NOT_FOUND_IN_MAPPING, errs["schema1.table1"][0].Code)
		assert.Contains(t, errs["schema1.table1"][0].Message, "Virtual foreign key target table missing in job mappings")
	})

	t.Run("should return error when source table missing in source database", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
			{Schema: "schema1", Table: "table2", Column: "col1"},
		}

		virtualForeignKeys := []*mgmtv1alpha1.VirtualForeignConstraint{
			{
				Schema:  "schema1",
				Table:   "table1",
				Columns: []string{"col1"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "schema1",
					Table:   "table2",
					Columns: []string{"col1"},
				},
			},
		}

		tableColumnMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateVirtualForeignKeys(virtualForeignKeys, tableColumnMap, &sqlmanager_shared.TableConstraints{})

		errs := jmv.GetTableErrors()
		require.Len(t, errs["schema1.table2"], 1)
		assert.Equal(t, mgmtv1alpha1.TableError_TABLE_ERROR_CODE_VFK_SOURCE_TABLE_NOT_FOUND_IN_SOURCE, errs["schema1.table2"][0].Code)
		assert.Contains(t, errs["schema1.table2"][0].Message, "Virtual foreign key source table missing in source database")
	})

	t.Run("should return error when target table missing in source database", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
			{Schema: "schema1", Table: "table2", Column: "col1"},
		}

		virtualForeignKeys := []*mgmtv1alpha1.VirtualForeignConstraint{
			{
				Schema:  "schema1",
				Table:   "table1",
				Columns: []string{"col1"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "schema1",
					Table:   "table2",
					Columns: []string{"col1"},
				},
			},
		}

		tableColumnMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table2": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateVirtualForeignKeys(virtualForeignKeys, tableColumnMap, &sqlmanager_shared.TableConstraints{})

		errs := jmv.GetTableErrors()
		require.Len(t, errs["schema1.table1"], 1)
		assert.Equal(t, mgmtv1alpha1.TableError_TABLE_ERROR_CODE_VFK_TARGET_TABLE_NOT_FOUND_IN_SOURCE, errs["schema1.table1"][0].Code)
		assert.Contains(t, errs["schema1.table1"][0].Message, "Virtual foreign key target table missing in source database")
	})

	t.Run("should return error when column datatypes don't match", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
			{Schema: "schema1", Table: "table2", Column: "col1"},
		}

		virtualForeignKeys := []*mgmtv1alpha1.VirtualForeignConstraint{
			{
				Schema:  "schema1",
				Table:   "table1",
				Columns: []string{"col1"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "schema1",
					Table:   "table2",
					Columns: []string{"col1"},
				},
			},
		}

		tableColumnMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					DataType: "integer",
				},
			},
			"schema1.table2": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					DataType: "text",
				},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateVirtualForeignKeys(virtualForeignKeys, tableColumnMap, &sqlmanager_shared.TableConstraints{})

		errs := jmv.GetColumnErrors()
		require.NotEmpty(t, errs)
		require.Len(t, errs["schema1.table1"]["col1"], 1)
		assert.Equal(t, mgmtv1alpha1.ColumnError_COLUMN_ERROR_CODE_VFK_COLUMN_DATATYPE_MISMATCH, errs["schema1.table1"]["col1"][0].Code)
		assert.Contains(t, errs["schema1.table1"]["col1"][0].Message, "Column datatype mismatch.")
	})

	t.Run("should return error foreign key source column missing constraint", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
		}

		virtualForeignKeys := []*mgmtv1alpha1.VirtualForeignConstraint{
			{
				Schema:  "schema1",
				Table:   "table1",
				Columns: []string{"col1"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "schema1",
					Table:   "table1",
					Columns: []string{"col1"},
				},
			},
		}

		tableColumnMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false, // Not nullable
				},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateVirtualForeignKeys(virtualForeignKeys, tableColumnMap, &sqlmanager_shared.TableConstraints{})

		errs := jmv.GetColumnErrors()
		require.NotEmpty(t, errs)
		assert.Equal(t, mgmtv1alpha1.ColumnError_COLUMN_ERROR_CODE_VFK_SOURCE_COLUMN_NOT_UNIQUE, errs["schema1.table1"]["col1"][0].Code)
		assert.Contains(t, errs["schema1.table1"]["col1"][0].Message, "Virtual foreign key source must be either a primary key or have a unique constraint")
	})

	t.Run("should return error when self-referencing virtual foreign key is not nullable", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
		}

		virtualForeignKeys := []*mgmtv1alpha1.VirtualForeignConstraint{
			{
				Schema:  "schema1",
				Table:   "table1",
				Columns: []string{"col1"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "schema1",
					Table:   "table1",
					Columns: []string{"col1"},
				},
			},
		}

		tableColumnMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false, // Not nullable
				},
			},
		}

		tableConstraints := &sqlmanager_shared.TableConstraints{
			PrimaryKeyConstraints: map[string][]string{
				"schema1.table1": {"col1"},
			},
		}
		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateVirtualForeignKeys(virtualForeignKeys, tableColumnMap, tableConstraints)

		errs := jmv.GetColumnErrors()
		require.NotEmpty(t, errs)
		require.Len(t, errs["schema1.table1"]["col1"], 1)
		assert.Equal(t, mgmtv1alpha1.ColumnError_COLUMN_ERROR_CODE_UNSUPPORTED_CIRCULAR_DEPENDENCY_AT_LEAST_ONE_NULLABLE, errs["schema1.table1"]["col1"][0].Code)
		assert.Contains(t, errs["schema1.table1"]["col1"][0].Message, "Self referencing virtual foreign key target column must be nullable")
	})

	t.Run("should not return error when self-referencing virtual foreign key is nullable", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
		}

		virtualForeignKeys := []*mgmtv1alpha1.VirtualForeignConstraint{
			{
				Schema:  "schema1",
				Table:   "table1",
				Columns: []string{"col1"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "schema1",
					Table:   "table1",
					Columns: []string{"col1"},
				},
			},
		}

		tableColumnMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: true, // Nullable
				},
			},
		}

		tableConstraints := &sqlmanager_shared.TableConstraints{
			UniqueConstraints: map[string][][]string{
				"schema1.table1": {{"col1"}},
			},
		}
		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateVirtualForeignKeys(virtualForeignKeys, tableColumnMap, tableConstraints)

		errs := jmv.GetColumnErrors()
		assert.Empty(t, errs)
	})

	t.Run("should return error when virtual foreign key source column missing in job mappings and source database", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
			{Schema: "schema1", Table: "table2", Column: "col1"},
			// col2 intentionally missing from mappings
		}

		virtualForeignKeys := []*mgmtv1alpha1.VirtualForeignConstraint{
			{
				Schema:  "schema1",
				Table:   "table2",
				Columns: []string{"col1"},
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "schema1",
					Table:   "table1",
					Columns: []string{"col2"}, // Reference missing column
				},
			},
		}

		tableColumnMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false,
					DataType:   "integer",
				},
				// col2 intentionally missing from source database
			},
			"schema1.table2": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false,
					DataType:   "integer",
				},
			},
		}

		tableConstraints := &sqlmanager_shared.TableConstraints{
			PrimaryKeyConstraints: map[string][]string{
				"schema1.table1": {"col2"},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateVirtualForeignKeys(virtualForeignKeys, tableColumnMap, tableConstraints)

		errs := jmv.GetColumnErrors()
		require.NotEmpty(t, errs)
		assert.Equal(t, []*mgmtv1alpha1.ColumnError_ColumnErrorReport{
			{
				Code:    mgmtv1alpha1.ColumnError_COLUMN_ERROR_CODE_VFK_SOURCE_COLUMN_NOT_FOUND_IN_MAPPING,
				Message: "Virtual foreign key source column missing in job mappings. Table: schema1.table1 Column: col2",
			},
			{
				Code:    mgmtv1alpha1.ColumnError_COLUMN_ERROR_CODE_VFK_SOURCE_COLUMN_NOT_FOUND_IN_SOURCE,
				Message: "Virtual foreign key source column missing in source database. Table: schema1.table1 Column: col2",
			},
		}, errs["schema1.table1"]["col2"])
	})

	t.Run("validates virtual foreign key target column exists in job mappings and source db", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
			{Schema: "schema1", Table: "table2", Column: "col2"},
			// col1 intentionally missing from table2 mappings
		}

		virtualForeignKeys := []*mgmtv1alpha1.VirtualForeignConstraint{
			{
				Schema:  "schema1",
				Table:   "table2",
				Columns: []string{"col1"}, // Reference missing mapping and column
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "schema1",
					Table:   "table1",
					Columns: []string{"col1"},
				},
			},
		}

		tableColumnMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false,
					DataType:   "integer",
				},
			},
			"schema1.table2": {
				"col2": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false,
					DataType:   "integer",
				},
				// col1 intentionally missing from source database
			},
		}

		tableConstraints := &sqlmanager_shared.TableConstraints{
			PrimaryKeyConstraints: map[string][]string{
				"schema1.table1": {"col1"},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateVirtualForeignKeys(virtualForeignKeys, tableColumnMap, tableConstraints)

		errs := jmv.GetColumnErrors()
		require.NotEmpty(t, errs)
		assert.Equal(t, []*mgmtv1alpha1.ColumnError_ColumnErrorReport{
			{
				Code:    mgmtv1alpha1.ColumnError_COLUMN_ERROR_CODE_VFK_TARGET_COLUMN_NOT_FOUND_IN_MAPPING,
				Message: "Virtual foreign key target column missing in job mappings. Table: schema1.table2 Column: col1",
			},
			{
				Code:    mgmtv1alpha1.ColumnError_COLUMN_ERROR_CODE_VFK_TARGET_COLUMN_NOT_FOUND_IN_SOURCE,
				Message: "Virtual foreign key target column missing in source database. Table: schema1.table2 Column: col1",
			},
		}, errs["schema1.table2"]["col1"])
	})

	t.Run("validates self referencing virtual foreign key target column must be nullable", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
		}

		virtualForeignKeys := []*mgmtv1alpha1.VirtualForeignConstraint{
			{
				Schema:  "schema1",
				Table:   "table1",
				Columns: []string{"col1"}, // Self reference to non-nullable column
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "schema1",
					Table:   "table1",
					Columns: []string{"col1"},
				},
			},
		}

		tableColumnMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					IsNullable: false, // Not nullable
					DataType:   "integer",
				},
			},
		}

		tableConstraints := &sqlmanager_shared.TableConstraints{
			PrimaryKeyConstraints: map[string][]string{
				"schema1.table1": {"col1"},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateVirtualForeignKeys(virtualForeignKeys, tableColumnMap, tableConstraints)

		errs := jmv.GetColumnErrors()
		require.NotEmpty(t, errs)
		require.Len(t, errs["schema1.table1"]["col1"], 1)
		assert.Equal(t, mgmtv1alpha1.ColumnError_COLUMN_ERROR_CODE_UNSUPPORTED_CIRCULAR_DEPENDENCY_AT_LEAST_ONE_NULLABLE, errs["schema1.table1"]["col1"][0].Code)
		assert.Contains(t, errs["schema1.table1"]["col1"][0].Message, "Self referencing virtual foreign key target column must be nullable")
	})

	t.Run("validates length of source and foreign key columns must match", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
			{Schema: "schema1", Table: "table1", Column: "col2"},
			{Schema: "schema1", Table: "table2", Column: "col1"},
		}

		virtualForeignKeys := []*mgmtv1alpha1.VirtualForeignConstraint{
			{
				Schema:  "schema1",
				Table:   "table2",
				Columns: []string{"col1"}, // Only one target column
				ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
					Schema:  "schema1",
					Table:   "table1",
					Columns: []string{"col1", "col2"}, // Two source columns - mismatch
				},
			},
		}

		tableColumnMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					DataType: "integer",
				},
				"col2": &sqlmanager_shared.DatabaseSchemaRow{
					DataType: "integer",
				},
			},
			"schema1.table2": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					DataType: "integer",
				},
			},
		}

		tableConstraints := &sqlmanager_shared.TableConstraints{
			PrimaryKeyConstraints: map[string][]string{
				"schema1.table1": {"col1", "col2"},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		jmv.ValidateVirtualForeignKeys(virtualForeignKeys, tableColumnMap, tableConstraints)

		errs := jmv.GetDatabaseErrors()
		require.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Message, "length of source columns was not equal to length of foreign key cols: 1 2")
		assert.Equal(t, mgmtv1alpha1.DatabaseError_DATABASE_ERROR_CODE_VFK_COLUMN_MISMATCH, errs[0].Code)
	})
}

func TestValidate(t *testing.T) {
	t.Run("validates successfully with no errors", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "table1", Column: "col1"},
			{Schema: "schema1", Table: "table1", Column: "col2"},
		}

		tableColumnMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{
					DataType:   "integer",
					IsNullable: false,
				},
				"col2": &sqlmanager_shared.DatabaseSchemaRow{
					DataType:   "varchar",
					IsNullable: true,
				},
			},
		}

		tableConstraints := &sqlmanager_shared.TableConstraints{
			PrimaryKeyConstraints: map[string][]string{
				"schema1.table1": {"col1"},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		resp, err := jmv.Validate(tableColumnMap, nil, tableConstraints)

		require.NoError(t, err)
		assert.Empty(t, resp.DatabaseErrors)
		assert.Empty(t, resp.ColumnErrors)
	})

	t.Run("returns errors when table missing from source", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{Schema: "schema1", Table: "missing_table", Column: "col1"},
		}

		tableColumnMap := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"schema1.table1": {
				"col1": &sqlmanager_shared.DatabaseSchemaRow{},
			},
		}

		jmv := NewJobMappingsValidator(mappings)
		resp, err := jmv.Validate(tableColumnMap, nil, &sqlmanager_shared.TableConstraints{})

		require.NoError(t, err)
		require.NotEmpty(t, resp.TableErrors)
		assert.Equal(t, []*mgmtv1alpha1.TableError_TableErrorReport{
			{
				Code:    mgmtv1alpha1.TableError_TABLE_ERROR_CODE_TABLE_NOT_FOUND_IN_SOURCE,
				Message: "Table does not exist [schema1.missing_table] in source",
			},
		}, resp.TableErrors["schema1.missing_table"])
	})
}
