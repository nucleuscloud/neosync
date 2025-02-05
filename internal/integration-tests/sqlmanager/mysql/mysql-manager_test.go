package sqlmanager_mysql

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_EscapeMysqlColumns(t *testing.T) {
	require.Empty(t, EscapeMysqlColumns(nil))
	require.Equal(
		t,
		EscapeMysqlColumns([]string{"foo", "bar", "baz"}),
		[]string{"`foo`", "`bar`", "`baz`"},
	)
}

func Test_BuildMysqlTruncateStatement(t *testing.T) {
	actual, err := BuildMysqlTruncateStatement("public", "users")
	require.NoError(t, err)
	require.Equal(
		t,
		`TRUNCATE "public"."users";`,
		actual,
	)
}
