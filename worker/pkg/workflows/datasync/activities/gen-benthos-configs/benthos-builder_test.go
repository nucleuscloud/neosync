package genbenthosconfigs_activity

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_mssql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mssql"
	sqlmanager_postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	benthosbuilder "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder"
	runconfigs "github.com/nucleuscloud/neosync/internal/runconfigs"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/stretchr/testify/require"

	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
)

func Test_buildPostTableSyncRunCtx(t *testing.T) {
	t.Run("Empty input", func(t *testing.T) {
		result := buildPostTableSyncRunCtx(nil, nil)
		require.Empty(t, result, "Expected empty map for empty input")
	})

	t.Run("No statements generated", func(t *testing.T) {
		benthosConfigs := []*benthosbuilder.BenthosConfigResponse{
			{
				Name:    "config1",
				RunType: runconfigs.RunTypeUpdate,
			},
		}
		destinations := []*mgmtv1alpha1.JobDestination{
			{
				ConnectionId: "dest1",
				Options: &mgmtv1alpha1.JobDestinationOptions{
					Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{},
				},
			},
		}
		result := buildPostTableSyncRunCtx(benthosConfigs, destinations)
		require.Empty(t, result, "Expected empty map when no statements are generated")
	})

	t.Run("Statements generated for Postgres and MSSQL", func(t *testing.T) {
		benthosConfigs := []*benthosbuilder.BenthosConfigResponse{
			{
				Name:    "config1",
				RunType: runconfigs.RunTypeInsert,
				ColumnDefaultProperties: map[string]*neosync_benthos.ColumnDefaultProperties{
					"col1": {NeedsReset: true, HasDefaultTransformer: false},
				},
				TableSchema: "public",
				TableName:   "table1",
			},
			{
				Name:    "config2",
				RunType: runconfigs.RunTypeInsert,
				ColumnDefaultProperties: map[string]*neosync_benthos.ColumnDefaultProperties{
					"col1": {NeedsOverride: true},
				},
				TableSchema: "dbo",
				TableName:   "table2",
			},
		}
		destinations := []*mgmtv1alpha1.JobDestination{
			{
				ConnectionId: "pg_dest",
				Options: &mgmtv1alpha1.JobDestinationOptions{
					Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{},
				},
			},
			{
				ConnectionId: "mssql_dest",
				Options: &mgmtv1alpha1.JobDestinationOptions{
					Config: &mgmtv1alpha1.JobDestinationOptions_MssqlOptions{},
				},
			},
		}

		result := buildPostTableSyncRunCtx(benthosConfigs, destinations)

		expected := map[string]*shared.PostTableSyncConfig{
			"config1": {
				DestinationConfigs: map[string]*shared.PostTableSyncDestConfig{
					"pg_dest": {
						Statements: []string{
							sqlmanager_postgres.BuildPgIdentityColumnResetCurrentSql("public", "table1", "col1"),
						},
					},
				},
			},
			"config2": {
				DestinationConfigs: map[string]*shared.PostTableSyncDestConfig{
					"mssql_dest": {
						Statements: []string{
							sqlmanager_mssql.BuildMssqlIdentityColumnResetCurrent("dbo", "table2"),
						},
					},
				},
			},
		}

		require.Equal(t, expected, result, "Unexpected result when statements are generated")
	})
}

func Test_BuildPgPostTableSyncStatement(t *testing.T) {
	t.Run("Update run type", func(t *testing.T) {
		bcUpdate := &benthosbuilder.BenthosConfigResponse{
			RunType: runconfigs.RunTypeUpdate,
		}
		resultUpdate := buildPgPostTableSyncStatement(bcUpdate)
		require.Empty(t, resultUpdate, "Expected empty slice for Update run type")
	})

	t.Run("No columns need reset", func(t *testing.T) {
		bcNoReset := &benthosbuilder.BenthosConfigResponse{
			RunType: runconfigs.RunTypeInsert,
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
		bcSomeReset := &benthosbuilder.BenthosConfigResponse{
			RunType: runconfigs.RunTypeInsert,
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
		bcUpdate := &benthosbuilder.BenthosConfigResponse{
			RunType: runconfigs.RunTypeUpdate,
		}
		resultUpdate := buildMssqlPostTableSyncStatement(bcUpdate)
		require.Empty(t, resultUpdate, "Expected empty slice for Update run type")
	})

	t.Run("No columns need override", func(t *testing.T) {
		bcNoOverride := &benthosbuilder.BenthosConfigResponse{
			RunType: runconfigs.RunTypeInsert,
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
		bcSomeOverride := &benthosbuilder.BenthosConfigResponse{
			RunType: runconfigs.RunTypeInsert,
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
