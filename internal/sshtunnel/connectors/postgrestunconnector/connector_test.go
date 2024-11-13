package postgrestunconnector

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		connector, cleanup, err := New(&net.Dialer{}, "postgres://postgres:postgres@localhost:5432")
		require.NoError(t, err)
		require.NotNil(t, cleanup)
		require.NotNil(t, connector)
		cleanup()
	})

	t.Run("invalid conn", func(t *testing.T) {
		connector, cleanup, err := New(&net.Dialer{}, "foo:bar@tcp(localhost:3306)")
		require.Error(t, err)
		require.Nil(t, cleanup)
		require.Nil(t, connector)
	})
}

func Test_Connector_Driver(t *testing.T) {
	connector, _, err := New(&net.Dialer{}, "postgres://postgres:postgres@localhost:5432")
	require.NoError(t, err)
	driver := connector.Driver()
	require.NotEmpty(t, driver)
}
