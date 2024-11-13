package sqlmanager_mssql

import (
	"database/sql"
	"testing"

	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
)

func Test_BuildMssqlDeleteStatement(t *testing.T) {
	actual, err := BuildMssqlDeleteStatement("public", "users")
	require.NoError(t, err)
	require.Equal(
		t,
		"DELETE FROM \"public\".\"users\";",
		actual,
	)
}

func Test_IsCircularSelfReferencingFk(t *testing.T) {
	t.Run("different tables should return false", func(t *testing.T) {
		row := &mssql_queries.GetTableConstraintsBySchemasRow{
			SchemaName: "dbo",
			TableName:  "Employee",
			ReferencedSchema: sql.NullString{
				String: "dbo",
				Valid:  true,
			},
			ReferencedTable: sql.NullString{
				String: "Department",
				Valid:  true,
			},
		}
		result := isCircularSelfReferencingFk(row, []string{"DepartmentId"}, []string{"Id"})
		assert.False(t, result)
	})

	t.Run("different schemas should return false", func(t *testing.T) {
		row := &mssql_queries.GetTableConstraintsBySchemasRow{
			SchemaName: "dbo",
			TableName:  "Employee",
			ReferencedSchema: sql.NullString{
				String: "hr",
				Valid:  true,
			},
			ReferencedTable: sql.NullString{
				String: "Employee",
				Valid:  true,
			},
		}
		result := isCircularSelfReferencingFk(row, []string{"ManagerId"}, []string{"Id"})
		assert.False(t, result)
	})

	t.Run("same table but different columns should return false", func(t *testing.T) {
		row := &mssql_queries.GetTableConstraintsBySchemasRow{
			SchemaName: "dbo",
			TableName:  "Employee",
			ReferencedSchema: sql.NullString{
				String: "dbo",
				Valid:  true,
			},
			ReferencedTable: sql.NullString{
				String: "Employee",
				Valid:  true,
			},
		}
		result := isCircularSelfReferencingFk(row, []string{"ManagerId"}, []string{"Id"})
		assert.False(t, result)
	})

	t.Run("circular reference with matching columns should return true", func(t *testing.T) {
		row := &mssql_queries.GetTableConstraintsBySchemasRow{
			SchemaName: "dbo",
			TableName:  "Employee",
			ReferencedSchema: sql.NullString{
				String: "dbo",
				Valid:  true,
			},
			ReferencedTable: sql.NullString{
				String: "Employee",
				Valid:  true,
			},
		}
		result := isCircularSelfReferencingFk(row, []string{"Id"}, []string{"Id"})
		assert.True(t, result)
	})

	t.Run("circular reference with multiple matching columns should return true", func(t *testing.T) {
		row := &mssql_queries.GetTableConstraintsBySchemasRow{
			SchemaName: "dbo",
			TableName:  "CompositeKey",
			ReferencedSchema: sql.NullString{
				String: "dbo",
				Valid:  true,
			},
			ReferencedTable: sql.NullString{
				String: "CompositeKey",
				Valid:  true,
			},
		}
		result := isCircularSelfReferencingFk(row,
			[]string{"Id", "SubId"},
			[]string{"SubId", "Id"})
		assert.True(t, result)
	})

	t.Run("circular reference with some non-matching columns should return false", func(t *testing.T) {
		row := &mssql_queries.GetTableConstraintsBySchemasRow{
			SchemaName: "dbo",
			TableName:  "CompositeKey",
			ReferencedSchema: sql.NullString{
				String: "dbo",
				Valid:  true,
			},
			ReferencedTable: sql.NullString{
				String: "CompositeKey",
				Valid:  true,
			},
		}
		result := isCircularSelfReferencingFk(row,
			[]string{"Id", "SubId"},
			[]string{"Id", "DifferentId"})
		assert.False(t, result)
	})

	t.Run("same table with empty columns should return true", func(t *testing.T) {
		row := &mssql_queries.GetTableConstraintsBySchemasRow{
			SchemaName: "dbo",
			TableName:  "EmptyTest",
			ReferencedSchema: sql.NullString{
				String: "dbo",
				Valid:  true,
			},
			ReferencedTable: sql.NullString{
				String: "EmptyTest",
				Valid:  true,
			},
		}
		result := isCircularSelfReferencingFk(row, []string{}, []string{})
		assert.True(t, result)
	})
}
