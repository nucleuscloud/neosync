package benthosbuilder_builders

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/stretchr/testify/require"
)

func Test_cleanPostgresType(t *testing.T) {
	t.Run("simple type without params", func(t *testing.T) {
		result := cleanPostgresType("integer")
		expected := "integer"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("type with single parameter", func(t *testing.T) {
		result := cleanPostgresType("varchar(255)")
		expected := "varchar"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("type with multiple parameters", func(t *testing.T) {
		result := cleanPostgresType("numeric(10,2)")
		expected := "numeric"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("character varying with parameter", func(t *testing.T) {
		result := cleanPostgresType("character varying(50)")
		expected := "character varying"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("type with spaces and parameter", func(t *testing.T) {
		result := cleanPostgresType("bit varying(8)")
		expected := "bit varying"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("array type", func(t *testing.T) {
		result := cleanPostgresType("integer[]")
		expected := "integer[]"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("type with time zone", func(t *testing.T) {
		result := cleanPostgresType("timestamp with time zone")
		expected := "timestamp with time zone"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("type with trailing space", func(t *testing.T) {
		result := cleanPostgresType("numeric(10,2) ")
		expected := "numeric"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("type with leading space", func(t *testing.T) {
		result := cleanPostgresType(" decimal(10,2)")
		expected := "decimal"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		result := cleanPostgresType("")
		expected := ""
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("just parentheses", func(t *testing.T) {
		result := cleanPostgresType("()")
		expected := ""
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("multiple parentheses", func(t *testing.T) {
		result := cleanPostgresType("foo(bar(baz))")
		expected := "foo"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("json type", func(t *testing.T) {
		result := cleanPostgresType("json")
		expected := "json"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("jsonb type", func(t *testing.T) {
		result := cleanPostgresType("jsonb")
		expected := "jsonb"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})
}

// Benchmark remains the same as it's already using discrete cases
func BenchmarkCleanPostgresType(b *testing.B) {
	inputs := []string{
		"integer",
		"character varying(50)",
		"numeric(10,2)",
		"timestamp with time zone",
		"text[]",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			cleanPostgresType(input)
		}
	}
}

func Test_getAdditionalMappings(t *testing.T) {
	t.Run("postgres", func(t *testing.T) {
		t.Run("none", func(t *testing.T) {
			actual, err := getAdditionalJobMappings(
				sqlmanager_shared.PostgresDriver,
				map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
					"public.users": {
						"id": {},
					},
				},
				[]*mgmtv1alpha1.JobMapping{{
					Schema: "public",
					Table:  "users",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{PassthroughConfig: &mgmtv1alpha1.Passthrough{}}},
					},
				}},
				splitKeyToTablePieces,
				testutil.GetTestLogger(t),
			)
			require.NoError(t, err)
			require.Empty(t, actual)
		})
		t.Run("fallbacks", func(t *testing.T) {
			actual, err := getAdditionalJobMappings(
				sqlmanager_shared.PostgresDriver,
				map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
					"public.users": {
						"id": {},
						"first_name": {
							IsNullable: true,
						},
						"last_name": {
							ColumnDefault: "FOO",
						},
						"full_name": {
							GeneratedType: sqlmanager_shared.Ptr("first_name + last_name"),
						},
						"income": {
							DataType: "numeric",
						},
					},
				},
				[]*mgmtv1alpha1.JobMapping{{
					Schema: "public",
					Table:  "users",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{PassthroughConfig: &mgmtv1alpha1.Passthrough{}}},
					},
				}},
				splitKeyToTablePieces,
				testutil.GetTestLogger(t),
			)
			require.NoError(t, err)
			require.Len(t, actual, 4)
		})
	})
}

func Test_isSourceMissingColumns(t *testing.T) {
	t.Run("no missing columns", func(t *testing.T) {
		missing, ok := isSourceMissingColumnsFoundInMappings(
			map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
				"public.users": {
					"id":   {},
					"name": {},
				},
			},
			[]*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "users",
					Column: "id",
				},
				{
					Schema: "public",
					Table:  "users",
					Column: "name",
				},
			},
		)
		require.False(t, ok)
		require.Empty(t, missing)
	})

	t.Run("missing table", func(t *testing.T) {
		missing, ok := isSourceMissingColumnsFoundInMappings(
			map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
				"public.users": {
					"id": {},
				},
			},
			[]*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "accounts", // non-existent table
					Column: "id",
				},
				{
					Schema: "public",
					Table:  "accounts", // non-existent table
					Column: "name",
				},
			},
		)
		require.True(t, ok)
		require.ElementsMatch(t, []string{"public.accounts.id", "public.accounts.name"}, missing)
	})

	t.Run("missing column", func(t *testing.T) {
		missing, ok := isSourceMissingColumnsFoundInMappings(
			map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
				"public.users": {
					"id": {},
				},
			},
			[]*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "users",
					Column: "id",
				},
				{
					Schema: "public",
					Table:  "users",
					Column: "email", // non-existent column
				},
			},
		)
		require.True(t, ok)
		require.Equal(t, []string{"public.users.email"}, missing)
	})
}

func Test_removeMappingsNotFoundInSource(t *testing.T) {
	t.Run("removes mappings for non-existent tables", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{
				Schema: "public",
				Table:  "users",
				Column: "id",
			},
			{
				Schema: "public",
				Table:  "accounts", // non-existent table
				Column: "id",
			},
		}

		groupedSchemas := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"public.users": {
				"id": {},
			},
		}

		result := removeMappingsNotFoundInSource(mappings, groupedSchemas)

		require.Len(t, result, 1)
		require.Equal(t, "public", result[0].Schema)
		require.Equal(t, "users", result[0].Table)
		require.Equal(t, "id", result[0].Column)
	})

	t.Run("removes mappings for non-existent columns", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{
				Schema: "public",
				Table:  "users",
				Column: "id",
			},
			{
				Schema: "public",
				Table:  "users",
				Column: "email", // non-existent column
			},
		}

		groupedSchemas := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"public.users": {
				"id": {},
			},
		}

		result := removeMappingsNotFoundInSource(mappings, groupedSchemas)

		require.Len(t, result, 1)
		require.Equal(t, "public", result[0].Schema)
		require.Equal(t, "users", result[0].Table)
		require.Equal(t, "id", result[0].Column)
	})

	t.Run("keeps all mappings when everything exists", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{
				Schema: "public",
				Table:  "users",
				Column: "id",
			},
			{
				Schema: "public",
				Table:  "users",
				Column: "email",
			},
		}

		groupedSchemas := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"public.users": {
				"id":    {},
				"email": {},
			},
		}

		result := removeMappingsNotFoundInSource(mappings, groupedSchemas)

		require.Len(t, result, 2)
		require.Equal(t, mappings, result)
	})

	t.Run("returns empty slice when no mappings exist in schema", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.JobMapping{
			{
				Schema: "public",
				Table:  "users",
				Column: "id",
			},
		}

		groupedSchemas := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
			"public.accounts": {
				"id": {},
			},
		}

		result := removeMappingsNotFoundInSource(mappings, groupedSchemas)

		require.Empty(t, result)
	})
}
