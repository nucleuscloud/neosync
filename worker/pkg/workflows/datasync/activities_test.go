package datasync

import (
	"math"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	dbschemas_postgres "github.com/nucleuscloud/neosync/worker/internal/dbschemas/postgres"
	"github.com/stretchr/testify/assert"
)

func TestAreMappingsSubsetFoSchemas(t *testing.T) {
	ok := areMappingsSubsetOfSchemas(
		[]*dbschemas_postgres.DatabaseSchema{
			{TableSchema: "public", TableName: "users", ColumnName: "id"},
			{TableSchema: "public", TableName: "users", ColumnName: "created_by"},
			{TableSchema: "public", TableName: "users", ColumnName: "updated_by"},

			{TableSchema: "neosync_api", TableName: "accounts", ColumnName: "id"},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "created_by"},
		},
	)
	assert.True(t, ok, "job mappings are a subset of the present database schemas")

	ok = areMappingsSubsetOfSchemas(
		[]*dbschemas_postgres.DatabaseSchema{
			{TableSchema: "public", TableName: "users", ColumnName: "id"},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id2"},
		},
	)
	assert.False(t, ok, "job mappings contain mapping that is not in the source schema")

	ok = areMappingsSubsetOfSchemas(
		[]*dbschemas_postgres.DatabaseSchema{
			{TableSchema: "public", TableName: "users", ColumnName: "id"},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "created_by"},
		},
	)
	assert.False(t, ok, "job mappings contain more mappings than are present in the source schema")
}

func TestShouldHaltOnSchemaAddition(t *testing.T) {
	ok := shouldHaltOnSchemaAddition(
		[]*dbschemas_postgres.DatabaseSchema{
			{TableSchema: "public", TableName: "users", ColumnName: "id"},
			{TableSchema: "public", TableName: "users", ColumnName: "created_by"},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "created_by"},
		},
	)
	assert.False(t, ok, "job mappings are valid set of database schemas")

	ok = shouldHaltOnSchemaAddition(
		[]*dbschemas_postgres.DatabaseSchema{
			{TableSchema: "public", TableName: "users", ColumnName: "id"},
			{TableSchema: "public", TableName: "users", ColumnName: "created_by"},

			{TableSchema: "neosync_api", TableName: "accounts", ColumnName: "id"},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "created_by"},
		},
	)
	assert.True(t, ok, "job mappings are missing database schema mappings")

	ok = shouldHaltOnSchemaAddition(
		[]*dbschemas_postgres.DatabaseSchema{
			{TableSchema: "public", TableName: "users", ColumnName: "id"},
			{TableSchema: "public", TableName: "users", ColumnName: "created_by"},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
		},
	)
	assert.True(t, ok, "job mappings are missing table column")

	ok = shouldHaltOnSchemaAddition(
		[]*dbschemas_postgres.DatabaseSchema{
			{TableSchema: "public", TableName: "users", ColumnName: "id"},
			{TableSchema: "public", TableName: "users", ColumnName: "created_by"},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "updated_by"},
		},
	)
	assert.True(t, ok, "job mappings have same column count, but missing specific column")
}

func TestClampUInt(t *testing.T) {
	assert.Equal(t, clampUint(0, 1, 2), 1)
	assert.Equal(t, clampUint(1, 1, 2), 1)
	assert.Equal(t, clampUint(2, 1, 2), 2)
	assert.Equal(t, clampUint(3, 1, 2), 2)
	assert.Equal(t, clampUint(1, 1, 1), 1)

	assert.Equal(t, clampUint(1, 3, 2), 3, "low is evaluated first, order is relevant")

}

func TestComputeMaxPgBatchCount(t *testing.T) {
	assert.Equal(t, computeMaxPgBatchCount(65535), 1)
	assert.Equal(t, computeMaxPgBatchCount(65536), 1, "anything over max should clamp to 1")
	assert.Equal(t, computeMaxPgBatchCount(math.MaxUint), 1, "anything over pgmax should clamp to 1")
	assert.Equal(t, computeMaxPgBatchCount(1), 65535)
	assert.Equal(t, computeMaxPgBatchCount(0), 65535)

}
