package mssqltunconnector

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		connector, cleanup, err := New(&net.Dialer{}, "sqlserver://sa:myStr0ngP%40assword@localhost?database=master")
		require.NoError(t, err)
		require.NotNil(t, cleanup)
		require.NotNil(t, connector)
	})

	t.Run("invalid dsn", func(t *testing.T) {
		connector, cleanup, err := New(&net.Dialer{}, "sqlserver://sa:myStr0ngP%40assword@localhost:invalidport")
		require.Error(t, err)
		require.Nil(t, cleanup)
		require.Nil(t, connector)
	})
}
