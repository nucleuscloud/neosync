package runconfigs

import (
	"testing"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/assert"
)

func TestNewTableConfigsBuilder(t *testing.T) {
	columns := map[string][]string{
		"public.users": {"id", "name", "email"},
	}
	primaryKeys := map[string][]string{
		"public.users": {"id"},
	}
	whereClauses := map[string]string{
		"public.users": "id < 100",
	}
	uniqueIndexes := map[string][][]string{
		"public.users": {{"email"}},
	}
	uniqueConstraints := map[string][][]string{
		"public.users": {{"name"}},
	}
	foreignKeys := map[string][]*sqlmanager_shared.ForeignConstraint{}

	builder := newTableConfigsBuilder(
		columns,
		primaryKeys,
		whereClauses,
		uniqueIndexes,
		uniqueConstraints,
		foreignKeys,
	)

	assert.NotNil(t, builder)
	assert.Equal(t, columns, builder.columns)
	assert.Equal(t, primaryKeys, builder.primaryKeys)
	assert.Equal(t, whereClauses, builder.whereClauses)
	assert.Equal(t, uniqueIndexes, builder.uniqueIndexes)
	assert.Equal(t, uniqueConstraints, builder.uniqueConstraints)
	assert.Equal(t, foreignKeys, builder.foreignKeys)
	assert.NotNil(t, builder.circularDependencies)
	assert.NotNil(t, builder.subsetPaths)
}

func TestBuildDependencyGraph(t *testing.T) {
	foreignKeys := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.orders": {
			{
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table: "public.users",
				},
			},
		},
		"public.order_items": {
			{
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table: "public.orders",
				},
			},
			{
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table: "public.products",
				},
			},
		},
	}

	builder := &tableConfigsBuilder{
		foreignKeys: foreignKeys,
	}

	graph := builder.buildDependencyGraph()

	assert.Len(t, graph, 2)
	assert.Equal(t, []string{"public.users"}, graph["public.orders"])
	assert.Equal(t, []string{"public.orders", "public.products"}, graph["public.order_items"])
}

func TestGetOrderByColumns_WithPrimaryKeys(t *testing.T) {
	builder := &runConfigBuilder{
		primaryKeys: []string{"id"},
	}

	orderByColumns := builder.getOrderByColumns([]string{"id", "name", "email"})

	assert.Equal(t, []string{"id"}, orderByColumns)
}

func TestGetOrderByColumns_WithUniqueConstraints(t *testing.T) {
	builder := &runConfigBuilder{
		primaryKeys:       []string{},
		uniqueConstraints: [][]string{{"email"}, {"name"}},
	}

	orderByColumns := builder.getOrderByColumns([]string{"id", "name", "email"})

	assert.Equal(t, []string{"email"}, orderByColumns)
}

func TestGetOrderByColumns_WithUniqueIndexes(t *testing.T) {
	builder := &runConfigBuilder{
		primaryKeys:       []string{},
		uniqueConstraints: [][]string{},
		uniqueIndexes:     [][]string{{"name"}, {"email"}},
	}

	orderByColumns := builder.getOrderByColumns([]string{"id", "name", "email"})

	assert.Equal(t, []string{"name"}, orderByColumns)
}

func TestGetOrderByColumns_FallbackToSortedColumns(t *testing.T) {
	builder := &runConfigBuilder{
		primaryKeys:       []string{},
		uniqueConstraints: [][]string{},
		uniqueIndexes:     [][]string{},
	}

	orderByColumns := builder.getOrderByColumns([]string{"id", "name", "email"})

	assert.Equal(t, []string{"email", "id", "name"}, orderByColumns)
}

func TestBuildInsertConfig(t *testing.T) {
	tableschema := sqlmanager_shared.SchemaTable{Schema: "public", Table: "users"}
	builder := &runConfigBuilder{
		table:             tableschema,
		columns:           []string{"id", "name", "email"},
		primaryKeys:       []string{"id"},
		whereClause:       stringPtr("id < 100"),
		subsetPaths:       []*SubsetPath{},
		uniqueIndexes:     [][]string{{"email"}},
		uniqueConstraints: [][]string{},
	}

	config := builder.buildInsertConfig()

	assert.Equal(t, "public.users.insert", config.id)
	assert.Equal(t, "public.users", config.table.String())
	assert.Equal(t, RunTypeInsert, config.runType)
	assert.Equal(t, []string{"id", "name", "email"}, config.selectColumns)
	assert.Equal(t, []string{"id", "name", "email"}, config.insertColumns)
	assert.Equal(t, []string{"id"}, config.primaryKeys)
	assert.Equal(t, stringPtr("id < 100"), config.whereClause)
	assert.Equal(t, []string{"id"}, config.orderByColumns)
	assert.Empty(t, config.dependsOn)
	assert.Empty(t, config.subsetPaths)
}

func TestBuild_TableConfigsBuilder(t *testing.T) {
	columns := map[string][]string{
		"public.users": {"id", "name", "email"},
	}
	primaryKeys := map[string][]string{
		"public.users": {"id"},
	}
	whereClauses := map[string]string{
		"public.users": "id < 100",
	}
	uniqueIndexes := map[string][][]string{
		"public.users": {{"email"}},
	}
	uniqueConstraints := map[string][][]string{}
	foreignKeys := map[string][]*sqlmanager_shared.ForeignConstraint{}

	builder := newTableConfigsBuilder(
		columns,
		primaryKeys,
		whereClauses,
		uniqueIndexes,
		uniqueConstraints,
		foreignKeys,
	)

	configs := builder.Build(sqlmanager_shared.SchemaTable{Schema: "public", Table: "users"})

	assert.Len(t, configs, 1)
	assert.Equal(t, "public.users.insert", configs[0].id)
	assert.Equal(t, "public.users", configs[0].table.String())
	assert.Equal(t, RunTypeInsert, configs[0].runType)
}

func TestBuildConstraintHandlingConfigs(t *testing.T) {
	tableschema := sqlmanager_shared.SchemaTable{Schema: "public", Table: "orders"}
	builder := &runConfigBuilder{
		table:       tableschema,
		columns:     []string{"id", "user_id", "product_id", "quantity"},
		primaryKeys: []string{"id"},
		foreignKeys: []*sqlmanager_shared.ForeignConstraint{
			{
				Columns:     []string{"user_id"},
				NotNullable: []bool{true},
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "public.users",
					Columns: []string{"id"},
				},
			},
			{
				Columns:     []string{"product_id"},
				NotNullable: []bool{false},
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "public.products",
					Columns: []string{"id"},
				},
			},
		},
		subsetPaths: []*SubsetPath{},
	}

	configs := builder.buildConstraintHandlingConfigs()

	assert.Len(t, configs, 2)

	// First config should be insert
	assert.Equal(t, "public.orders.insert", configs[0].id)
	assert.Equal(t, RunTypeInsert, configs[0].runType)
	assert.Contains(t, configs[0].insertColumns, "id")
	assert.Contains(t, configs[0].insertColumns, "user_id")
	assert.Contains(t, configs[0].insertColumns, "quantity")
	assert.NotContains(t, configs[0].insertColumns, "product_id") // Nullable FK should not be in insert
	assert.Len(t, configs[0].dependsOn, 1)
	assert.Equal(t, "public.users", configs[0].dependsOn[0].Table)

	// Second config should be update for nullable FK
	assert.Equal(t, "public.orders.update.1", configs[1].id)
	assert.Equal(t, RunTypeUpdate, configs[1].runType)
	assert.Equal(t, []string{"product_id"}, configs[1].insertColumns)
	assert.Len(t, configs[1].dependsOn, 2)
	assert.Equal(t, "public.products", configs[1].dependsOn[0].Table)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
