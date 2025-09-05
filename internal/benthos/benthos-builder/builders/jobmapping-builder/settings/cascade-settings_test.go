package cascade_settings

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetSchemaStrategy(t *testing.T) {
	t.Run("returns default when no strategy set", func(t *testing.T) {
		cfg := &mgmtv1alpha1.JobTypeConfig_JobTypeSync{}
		settings := NewCascadeSchemaSettings(cfg, testutil.GetConcurrentTestLogger(t))

		strategy := settings.GetSchemaStrategy()
		require.NotNil(t, strategy)
		require.NotNil(t, strategy.GetMapAllSchemas())
	})

	t.Run("returns configured strategy", func(t *testing.T) {
		cfg := &mgmtv1alpha1.JobTypeConfig_JobTypeSync{
			SchemaChange: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaChange{
				SchemaStrategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy{
					Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy_MapAllSchemas_{
						MapAllSchemas: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy_MapAllSchemas{},
					},
				},
			},
		}
		settings := NewCascadeSchemaSettings(cfg, testutil.GetConcurrentTestLogger(t))

		strategy := settings.GetSchemaStrategy()
		require.NotNil(t, strategy)
		require.NotNil(t, strategy.GetMapAllSchemas())
	})
}

func Test_GetTableStrategy(t *testing.T) {
	t.Run("returns default when no strategy set", func(t *testing.T) {
		cfg := &mgmtv1alpha1.JobTypeConfig_JobTypeSync{}
		settings := NewCascadeSchemaSettings(cfg, testutil.GetConcurrentTestLogger(t))

		strategy := settings.GetTableStrategy("public")
		require.NotNil(t, strategy)
		require.NotNil(t, strategy.GetMapAllTables())
	})

	t.Run("returns global strategy when no schema specific strategy", func(t *testing.T) {
		cfg := &mgmtv1alpha1.JobTypeConfig_JobTypeSync{
			SchemaChange: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaChange{
				TableStrategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy{
					Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapDefinedTables_{
						MapDefinedTables: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapDefinedTables{},
					},
				},
			},
		}
		settings := NewCascadeSchemaSettings(cfg, testutil.GetConcurrentTestLogger(t))

		strategy := settings.GetTableStrategy("public")
		require.NotNil(t, strategy)
		require.NotNil(t, strategy.GetMapDefinedTables())
	})

	t.Run("returns schema level strategy over global", func(t *testing.T) {
		cfg := &mgmtv1alpha1.JobTypeConfig_JobTypeSync{
			SchemaChange: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaChange{
				TableStrategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy{
					Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapAllTables_{
						MapAllTables: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapAllTables{},
					},
				},
			},
			SchemaMappings: []*mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaMapping{
				{
					Schema: "public",
					TableStrategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy{
						Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapDefinedTables_{
							MapDefinedTables: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapDefinedTables{},
						},
					},
				},
			},
		}
		settings := NewCascadeSchemaSettings(cfg, testutil.GetConcurrentTestLogger(t))

		strategy := settings.GetTableStrategy("public")
		require.NotNil(t, strategy)
		require.NotNil(t, strategy.GetMapDefinedTables())
	})
}

func Test_GetColumnStrategy(t *testing.T) {
	t.Run("returns default when no strategy set", func(t *testing.T) {
		cfg := &mgmtv1alpha1.JobTypeConfig_JobTypeSync{}
		settings := NewCascadeSchemaSettings(cfg, testutil.GetConcurrentTestLogger(t))

		strategy := settings.GetColumnStrategy("public", "users")
		require.NotNil(t, strategy)
		require.NotNil(t, strategy.GetMapAllColumns())
	})

	t.Run("returns global strategy when no schema or table specific strategy", func(t *testing.T) {
		cfg := &mgmtv1alpha1.JobTypeConfig_JobTypeSync{
			SchemaChange: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaChange{
				ColumnStrategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy{
					Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_MapAllColumns_{
						MapAllColumns: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_MapAllColumns{
							ColumnInSourceNotMapped: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy{
								Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy_Drop_{},
							},
						},
					},
				},
			},
		}
		settings := NewCascadeSchemaSettings(cfg, testutil.GetConcurrentTestLogger(t))

		strategy := settings.GetColumnStrategy("public", "users")
		require.NotNil(t, strategy)
		require.NotNil(t, strategy.GetMapAllColumns())
		require.NotNil(t, strategy.GetMapAllColumns().GetColumnInSourceNotMapped().GetDrop())
	})

	t.Run("returns table level strategy over schema and global", func(t *testing.T) {
		cfg := &mgmtv1alpha1.JobTypeConfig_JobTypeSync{
			SchemaChange: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaChange{
				ColumnStrategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy{
					Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_MapAllColumns_{
						MapAllColumns: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_MapAllColumns{
							ColumnInSourceNotMapped: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy{
								Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy_Drop_{},
							},
						},
					},
				},
			},
			SchemaMappings: []*mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaMapping{
				{
					Schema: "public",
					TableMappings: []*mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaMapping_TableMapping{
						{
							Table: "users",
							ColumnStrategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy{
								Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_MapAllColumns_{
									MapAllColumns: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_MapAllColumns{
										ColumnInSourceNotMapped: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy{
											Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy_Passthrough_{},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		settings := NewCascadeSchemaSettings(cfg, testutil.GetConcurrentTestLogger(t))

		strategy := settings.GetColumnStrategy("public", "users")
		require.NotNil(t, strategy)
		require.NotNil(t, strategy.GetMapAllColumns())
		require.NotNil(t, strategy.GetMapAllColumns().GetColumnInSourceNotMapped().GetPassthrough())
	})
}

func Test_GetColumnTransformerConfigByTable(t *testing.T) {
	t.Run("returns empty when no transformers configured", func(t *testing.T) {
		cfg := &mgmtv1alpha1.JobTypeConfig_JobTypeSync{}
		settings := NewCascadeSchemaSettings(cfg, testutil.GetConcurrentTestLogger(t))

		var count int
		for range settings.GetColumnTransformerConfigByTable("public", "users") {
			count++
		}
		assert.Equal(t, 0, count)
	})

	t.Run("returns configured transformers", func(t *testing.T) {
		transformer := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
				PassthroughConfig: &mgmtv1alpha1.Passthrough{},
			},
		}
		cfg := &mgmtv1alpha1.JobTypeConfig_JobTypeSync{
			SchemaMappings: []*mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaMapping{
				{
					Schema: "public",
					TableMappings: []*mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaMapping_TableMapping{
						{
							Table: "users",
							ColumnMappings: []*mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaMapping_TableMapping_ColumnMapping{
								{
									Column:      "id",
									Transformer: transformer,
								},
							},
						},
					},
				},
			},
		}
		settings := NewCascadeSchemaSettings(cfg, testutil.GetConcurrentTestLogger(t))

		var transformers []*mgmtv1alpha1.TransformerConfig
		var columns []string
		for col, trans := range settings.GetColumnTransformerConfigByTable("public", "users") {
			columns = append(columns, col)
			transformers = append(transformers, trans)
		}
		require.Len(t, transformers, 1)
		require.Equal(t, transformer, transformers[0])
		require.Equal(t, []string{"id"}, columns)
	})
}

func Test_GetDefinedSchemas(t *testing.T) {
	t.Run("returns empty when no schemas defined", func(t *testing.T) {
		cfg := &mgmtv1alpha1.JobTypeConfig_JobTypeSync{}
		settings := NewCascadeSchemaSettings(cfg, testutil.GetConcurrentTestLogger(t))

		schemas := settings.GetDefinedSchemas()
		assert.Empty(t, schemas)
	})

	t.Run("returns defined schemas", func(t *testing.T) {
		cfg := &mgmtv1alpha1.JobTypeConfig_JobTypeSync{
			SchemaMappings: []*mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaMapping{
				{Schema: "public"},
				{Schema: "private"},
			},
		}
		settings := NewCascadeSchemaSettings(cfg, testutil.GetConcurrentTestLogger(t))

		schemas := settings.GetDefinedSchemas()
		assert.ElementsMatch(t, []string{"public", "private"}, schemas)
	})
}

func Test_GetDefinedTables(t *testing.T) {
	t.Run("returns empty when no tables defined", func(t *testing.T) {
		cfg := &mgmtv1alpha1.JobTypeConfig_JobTypeSync{}
		settings := NewCascadeSchemaSettings(cfg, testutil.GetConcurrentTestLogger(t))

		tables := settings.GetDefinedTables("public")
		assert.Empty(t, tables)
	})

	t.Run("returns defined tables for schema", func(t *testing.T) {
		cfg := &mgmtv1alpha1.JobTypeConfig_JobTypeSync{
			SchemaMappings: []*mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaMapping{
				{
					Schema: "public",
					TableMappings: []*mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaMapping_TableMapping{
						{Table: "users"},
						{Table: "posts"},
					},
				},
				{
					Schema: "private",
					TableMappings: []*mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaMapping_TableMapping{
						{Table: "secrets"},
					},
				},
			},
		}
		settings := NewCascadeSchemaSettings(cfg, testutil.GetConcurrentTestLogger(t))

		tables := settings.GetDefinedTables("public")
		assert.ElementsMatch(t, []string{"users", "posts"}, tables)

		tables = settings.GetDefinedTables("private")
		assert.ElementsMatch(t, []string{"secrets"}, tables)
	})
}
