package ee_sqlmanager_mssql

import (
	"database/sql"
	"testing"

	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	"github.com/stretchr/testify/assert"
)

func Test_orderObjectsByDependency(t *testing.T) {
	t.Run("empty input returns empty output", func(t *testing.T) {
		result := orderObjectsByDependency([]*mssql_queries.GetViewsAndFunctionsBySchemasRow{})
		assert.Empty(t, result)
	})

	t.Run("single object with no dependencies", func(t *testing.T) {
		objects := []*mssql_queries.GetViewsAndFunctionsBySchemasRow{
			{
				SchemaName: "dbo",
				ObjectName: "View1",
				Dependencies: sql.NullString{
					Valid:  false,
					String: "",
				},
			},
		}
		result := orderObjectsByDependency(objects)
		assert.Len(t, result, 1)
		assert.Equal(t, "View1", result[0].ObjectName)
	})

	t.Run("linear dependency chain", func(t *testing.T) {
		objects := []*mssql_queries.GetViewsAndFunctionsBySchemasRow{
			{
				SchemaName: "dbo",
				ObjectName: "View3",
				Dependencies: sql.NullString{
					Valid:  true,
					String: "dbo.View2",
				},
			},
			{
				SchemaName: "dbo",
				ObjectName: "View2",
				Dependencies: sql.NullString{
					Valid:  true,
					String: "dbo.View1",
				},
			},
			{
				SchemaName: "dbo",
				ObjectName: "View1",
				Dependencies: sql.NullString{
					Valid:  false,
					String: "",
				},
			},
		}
		result := orderObjectsByDependency(objects)
		assert.Len(t, result, 3)
		assert.Equal(t, "View1", result[0].ObjectName)
		assert.Equal(t, "View2", result[1].ObjectName)
		assert.Equal(t, "View3", result[2].ObjectName)
	})

	t.Run("multiple independent objects", func(t *testing.T) {
		objects := []*mssql_queries.GetViewsAndFunctionsBySchemasRow{
			{
				SchemaName: "dbo",
				ObjectName: "View1",
				Dependencies: sql.NullString{
					Valid: false,
				},
			},
			{
				SchemaName: "dbo",
				ObjectName: "View2",
				Dependencies: sql.NullString{
					Valid: false,
				},
			},
		}
		result := orderObjectsByDependency(objects)
		assert.Len(t, result, 2)
		names := []string{result[0].ObjectName, result[1].ObjectName}
		assert.Contains(t, names, "View1")
		assert.Contains(t, names, "View2")
	})

	t.Run("circular dependency", func(t *testing.T) {
		objects := []*mssql_queries.GetViewsAndFunctionsBySchemasRow{
			{
				SchemaName: "dbo",
				ObjectName: "View1",
				Dependencies: sql.NullString{
					Valid:  true,
					String: "dbo.View2",
				},
			},
			{
				SchemaName: "dbo",
				ObjectName: "View2",
				Dependencies: sql.NullString{
					Valid:  true,
					String: "dbo.View1",
				},
			},
		}
		result := orderObjectsByDependency(objects)
		assert.Len(t, result, 2)
		names := []string{result[0].ObjectName, result[1].ObjectName}
		assert.Contains(t, names, "View1")
		assert.Contains(t, names, "View2")
	})

	t.Run("complex dependency graph", func(t *testing.T) {
		objects := []*mssql_queries.GetViewsAndFunctionsBySchemasRow{
			{
				SchemaName: "dbo",
				ObjectName: "View4",
				Dependencies: sql.NullString{
					Valid:  true,
					String: "dbo.View2, dbo.View3",
				},
			},
			{
				SchemaName: "dbo",
				ObjectName: "View3",
				Dependencies: sql.NullString{
					Valid:  true,
					String: "dbo.View1",
				},
			},
			{
				SchemaName: "dbo",
				ObjectName: "View2",
				Dependencies: sql.NullString{
					Valid:  true,
					String: "dbo.View1",
				},
			},
			{
				SchemaName: "dbo",
				ObjectName: "View1",
				Dependencies: sql.NullString{
					Valid: false,
				},
			},
		}
		result := orderObjectsByDependency(objects)
		assert.Len(t, result, 4)
		assert.Equal(t, "View1", result[0].ObjectName)
		assert.Equal(t, "View4", result[len(result)-1].ObjectName)
	})

	t.Run("dependency on non-existent object", func(t *testing.T) {
		objects := []*mssql_queries.GetViewsAndFunctionsBySchemasRow{
			{
				SchemaName: "dbo",
				ObjectName: "View1",
				Dependencies: sql.NullString{
					Valid:  true,
					String: "dbo.NonExistentView",
				},
			},
		}
		result := orderObjectsByDependency(objects)
		assert.Len(t, result, 1)
		assert.Equal(t, "View1", result[0].ObjectName)
	})

	t.Run("multiple schemas", func(t *testing.T) {
		objects := []*mssql_queries.GetViewsAndFunctionsBySchemasRow{
			{
				SchemaName: "schema2",
				ObjectName: "View2",
				Dependencies: sql.NullString{
					Valid:  true,
					String: "schema1.View1",
				},
			},
			{
				SchemaName: "schema1",
				ObjectName: "View1",
				Dependencies: sql.NullString{
					Valid: false,
				},
			},
		}
		result := orderObjectsByDependency(objects)
		assert.Len(t, result, 2)
		assert.Equal(t, "schema1.View1", result[0].SchemaName+"."+result[0].ObjectName)
		assert.Equal(t, "schema2.View2", result[1].SchemaName+"."+result[1].ObjectName)
	})
}
