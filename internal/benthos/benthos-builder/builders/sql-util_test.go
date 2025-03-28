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

func Test_cleanMysqlType(t *testing.T) {
	t.Run("simple type without params", func(t *testing.T) {
		result := cleanMysqlType("varchar")
		expected := "varchar"
		require.Equal(t, expected, result)
	})

	t.Run("type with single parameter", func(t *testing.T) {
		result := cleanMysqlType("varchar(255)")
		expected := "varchar"
		require.Equal(t, expected, result)
	})

	t.Run("type with multiple parameters", func(t *testing.T) {
		result := cleanMysqlType("decimal(10,2)")
		expected := "decimal"
		require.Equal(t, expected, result)
	})

	t.Run("enum type with values", func(t *testing.T) {
		result := cleanMysqlType("enum('small','medium','large')")
		expected := "enum"
		require.Equal(t, expected, result)
	})

	t.Run("set type with values", func(t *testing.T) {
		result := cleanMysqlType("set('draft','published','archived')")
		expected := "set"
		require.Equal(t, expected, result)
	})

	t.Run("type with trailing space", func(t *testing.T) {
		result := cleanMysqlType("decimal(10,2) ")
		expected := "decimal"
		require.Equal(t, expected, result)
	})

	t.Run("type with leading space", func(t *testing.T) {
		result := cleanMysqlType(" decimal(10,2)")
		expected := "decimal"
		require.Equal(t, expected, result)
	})

	t.Run("empty string", func(t *testing.T) {
		result := cleanMysqlType("")
		expected := ""
		require.Equal(t, expected, result)
	})

	t.Run("just parentheses", func(t *testing.T) {
		result := cleanMysqlType("()")
		expected := ""
		require.Equal(t, expected, result)
	})

	t.Run("datetime type", func(t *testing.T) {
		result := cleanMysqlType("datetime")
		expected := "datetime"
		require.Equal(t, expected, result)
	})

	t.Run("timestamp with precision", func(t *testing.T) {
		result := cleanMysqlType("timestamp(6)")
		expected := "timestamp"
		require.Equal(t, expected, result)
	})
}

func BenchmarkCleanMysqlType(b *testing.B) {
	inputs := []string{
		"varchar",
		"decimal(10,2)",
		"enum('small','medium','large')",
		"set('draft','published','archived')",
		"datetime",
		"timestamp(6)",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			cleanMysqlType(input)
		}
	}
}

func Test_extractMysqlTypeParams(t *testing.T) {
	t.Run("numeric parameters", func(t *testing.T) {
		t.Run("single parameter", func(t *testing.T) {
			result := extractMysqlTypeParams("CHAR(10)")
			require.Equal(t, []string{"10"}, result)
		})

		t.Run("multiple parameters", func(t *testing.T) {
			result := extractMysqlTypeParams("FLOAT(10,2)")
			require.Equal(t, []string{"10", "2"}, result)
		})

		t.Run("parameters with spaces", func(t *testing.T) {
			result := extractMysqlTypeParams("DECIMAL(10, 2)")
			require.Equal(t, []string{"10", "2"}, result)
		})
	})

	t.Run("enum values", func(t *testing.T) {
		t.Run("simple values", func(t *testing.T) {
			result := extractMysqlTypeParams("ENUM('small','medium','large')")
			require.Equal(t, []string{"small", "medium", "large"}, result)
		})

		t.Run("values with spaces", func(t *testing.T) {
			result := extractMysqlTypeParams("ENUM('small', 'medium', 'large')")
			require.Equal(t, []string{"small", "medium", "large"}, result)
		})

		t.Run("values with special characters", func(t *testing.T) {
			result := extractMysqlTypeParams("ENUM('draft-1','published_2','archived.3')")
			require.Equal(t, []string{"draft-1", "published_2", "archived.3"}, result)
		})
	})

	t.Run("set values", func(t *testing.T) {
		result := extractMysqlTypeParams("SET('draft','published','archived')")
		require.Equal(t, []string{"draft", "published", "archived"}, result)
	})

	t.Run("edge cases", func(t *testing.T) {
		t.Run("no parameters", func(t *testing.T) {
			result := extractMysqlTypeParams("VARCHAR")
			require.Nil(t, result)
		})

		t.Run("empty parameters", func(t *testing.T) {
			result := extractMysqlTypeParams("CHAR()")
			require.Empty(t, result)
		})

		t.Run("missing closing parenthesis", func(t *testing.T) {
			result := extractMysqlTypeParams("CHAR(10")
			require.Nil(t, result)
		})

		t.Run("empty enum values", func(t *testing.T) {
			result := extractMysqlTypeParams("ENUM('','','')")
			require.Empty(t, result)
		})

		t.Run("mixed empty and non-empty enum values", func(t *testing.T) {
			result := extractMysqlTypeParams("ENUM('',  'valid',  '')")
			require.Equal(t, []string{"valid"}, result)
		})
	})
}
