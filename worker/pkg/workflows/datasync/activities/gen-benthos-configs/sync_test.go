package genbenthosconfigs_activity

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_mssql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mssql"
	sqlmanager_postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/stretchr/testify/require"
)

func TestFilterForeignKeysMap(t *testing.T) {
	tests := []struct {
		name              string
		colTransformerMap map[string]map[string]*mgmtv1alpha1.JobMappingTransformer
		foreignKeysMap    map[string][]*sqlmanager_shared.ForeignConstraint
		expected          map[string][]*sqlmanager_shared.ForeignConstraint
	}{
		{
			name:              "Empty input maps",
			colTransformerMap: map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{},
			foreignKeysMap:    map[string][]*sqlmanager_shared.ForeignConstraint{},
			expected:          map[string][]*sqlmanager_shared.ForeignConstraint{},
		},
		{
			name: "No matching tables",
			colTransformerMap: map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{
				"table1": {"col1": &mgmtv1alpha1.JobMappingTransformer{}},
			},
			foreignKeysMap: map[string][]*sqlmanager_shared.ForeignConstraint{
				"table2": {
					{
						Columns:     []string{"col1"},
						NotNullable: []bool{true},
						ForeignKey:  &sqlmanager_shared.ForeignKey{Table: "ref_table", Columns: []string{"ref_col"}},
					},
				},
			},
			expected: map[string][]*sqlmanager_shared.ForeignConstraint{},
		},
		{
			name: "Filtered composite foreign keys",
			colTransformerMap: map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{
				"table1": {
					"col1": &mgmtv1alpha1.JobMappingTransformer{Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL},
					"col2": &mgmtv1alpha1.JobMappingTransformer{Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL},
					"col3": &mgmtv1alpha1.JobMappingTransformer{Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL},
				},
			},
			foreignKeysMap: map[string][]*sqlmanager_shared.ForeignConstraint{
				"table1": {
					{
						Columns:     []string{"col1", "col2", "col3"},
						NotNullable: []bool{false, true, true},
						ForeignKey:  &sqlmanager_shared.ForeignKey{Table: "ref_table", Columns: []string{"ref_col1", "ref_col2", "ref_col3"}},
					},
				},
			},
			expected: map[string][]*sqlmanager_shared.ForeignConstraint{
				"table1": {
					{
						Columns:     []string{"col2", "col3"},
						NotNullable: []bool{true, true},
						ForeignKey:  &sqlmanager_shared.ForeignKey{Table: "ref_table", Columns: []string{"ref_col2", "ref_col3"}},
					},
				},
			},
		},
		{
			name: "Filtered foreign keys",
			colTransformerMap: map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{
				"table1": {
					"col1": &mgmtv1alpha1.JobMappingTransformer{Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH},
				},
				"table2": {
					"col2": &mgmtv1alpha1.JobMappingTransformer{Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL},
				},
			},
			foreignKeysMap: map[string][]*sqlmanager_shared.ForeignConstraint{
				"table1": {
					{
						Columns:     []string{"col1"},
						NotNullable: []bool{false},
						ForeignKey:  &sqlmanager_shared.ForeignKey{Table: "ref_table", Columns: []string{"ref_col1"}},
					},
				},
				"table2": {
					{
						Columns:     []string{"col2"},
						NotNullable: []bool{false},
						ForeignKey:  &sqlmanager_shared.ForeignKey{Table: "ref_table", Columns: []string{"ref_col2"}},
					},
				},
			},
			expected: map[string][]*sqlmanager_shared.ForeignConstraint{
				"table1": {
					{
						Columns:     []string{"col1"},
						NotNullable: []bool{false},
						ForeignKey:  &sqlmanager_shared.ForeignKey{Table: "ref_table", Columns: []string{"ref_col1"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterForeignKeysMap(tt.colTransformerMap, tt.foreignKeysMap)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_BuildPgPostTableSyncStatement(t *testing.T) {
	t.Run("Update run type", func(t *testing.T) {
		bcUpdate := &BenthosConfigResponse{
			RunType: tabledependency.RunTypeUpdate,
		}
		resultUpdate := buildPgPostTableSyncStatement(bcUpdate)
		require.Empty(t, resultUpdate, "Expected empty slice for Update run type")
	})

	t.Run("No columns need reset", func(t *testing.T) {
		bcNoReset := &BenthosConfigResponse{
			RunType: tabledependency.RunTypeInsert,
			ColumnDefaultProperties: map[string]*neosync_benthos.ColumnDefaultProperties{
				"col1": {NeedsReset: false, HasDefaultTransformer: false},
				"col2": {NeedsReset: false, HasDefaultTransformer: true},
			},
			TableSchema: "public",
			TableName:   "test_table",
		}
		resultNoReset := buildPgPostTableSyncStatement(bcNoReset)
		require.Empty(t, resultNoReset, "Expected empty slice when no columns need reset")
	})

	t.Run("Some columns need reset", func(t *testing.T) {
		bcSomeReset := &BenthosConfigResponse{
			RunType: tabledependency.RunTypeInsert,
			ColumnDefaultProperties: map[string]*neosync_benthos.ColumnDefaultProperties{
				"col1": {NeedsReset: true, HasDefaultTransformer: false},
				"col2": {NeedsReset: false, HasDefaultTransformer: true},
				"col3": {NeedsReset: true, HasDefaultTransformer: false},
			},
			TableSchema: "public",
			TableName:   "test_table",
		}
		resultSomeReset := buildPgPostTableSyncStatement(bcSomeReset)
		expectedSomeReset := []string{
			sqlmanager_postgres.BuildPgIdentityColumnResetCurrentSql("public", "test_table", "col1"),
			sqlmanager_postgres.BuildPgIdentityColumnResetCurrentSql("public", "test_table", "col3"),
		}
		require.ElementsMatch(t, expectedSomeReset, resultSomeReset, "Unexpected result when some columns need reset")
	})
}

func Test_BuildMssqlPostTableSyncStatement(t *testing.T) {
	t.Run("Update run type", func(t *testing.T) {
		bcUpdate := &BenthosConfigResponse{
			RunType: tabledependency.RunTypeUpdate,
		}
		resultUpdate := buildMssqlPostTableSyncStatement(bcUpdate)
		require.Empty(t, resultUpdate, "Expected empty slice for Update run type")
	})

	t.Run("No columns need override", func(t *testing.T) {
		bcNoOverride := &BenthosConfigResponse{
			RunType: tabledependency.RunTypeInsert,
			ColumnDefaultProperties: map[string]*neosync_benthos.ColumnDefaultProperties{
				"col1": {NeedsOverride: false},
				"col2": {NeedsOverride: false},
			},
			TableSchema: "dbo",
			TableName:   "test_table",
		}
		resultNoOverride := buildMssqlPostTableSyncStatement(bcNoOverride)
		require.Empty(t, resultNoOverride, "Expected empty slice when no columns need override")
	})

	t.Run("Some columns need override", func(t *testing.T) {
		bcSomeOverride := &BenthosConfigResponse{
			RunType: tabledependency.RunTypeInsert,
			ColumnDefaultProperties: map[string]*neosync_benthos.ColumnDefaultProperties{
				"col1": {NeedsOverride: true},
				"col2": {NeedsOverride: false},
			},
			TableSchema: "dbo",
			TableName:   "test_table",
		}
		resultSomeOverride := buildMssqlPostTableSyncStatement(bcSomeOverride)
		expectedSomeOverride := []string{
			sqlmanager_mssql.BuildMssqlIdentityColumnResetCurrent("dbo", "test_table"),
		}
		require.Equal(t, expectedSomeOverride, resultSomeOverride, "Unexpected result when some columns need override")
	})
}
