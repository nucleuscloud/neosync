package sqlmanager_mssql

import (
	"testing"

	"github.com/stretchr/testify/require"
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
