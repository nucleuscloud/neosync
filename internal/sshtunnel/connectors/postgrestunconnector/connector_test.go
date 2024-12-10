package postgrestunconnector

import (
	"crypto/tls"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		connector, cleanup, err := New("postgres://postgres:postgres@localhost:5432")
		require.NoError(t, err)
		require.NotNil(t, cleanup)
		require.NotNil(t, connector)
		cleanup()
	})

	t.Run("invalid conn", func(t *testing.T) {
		connector, cleanup, err := New("foo:bar@tcp(localhost:3306)")
		require.Error(t, err)
		require.Nil(t, cleanup)
		require.Nil(t, connector)
	})

	t.Run("WithDialer", func(t *testing.T) {
		connector, cleanup, err := New("postgres://postgres:postgres@localhost:5432", WithDialer(&net.Dialer{}))
		require.NoError(t, err)
		require.NotNil(t, cleanup)
		require.NotNil(t, connector)
		cleanup()
	})

	t.Run("WithTls", func(t *testing.T) {
		connector, cleanup, err := New("postgres://postgres:postgres@localhost:5432", WithTLSConfig(&tls.Config{}))
		require.NoError(t, err)
		require.NotNil(t, cleanup)
		require.NotNil(t, connector)
		cleanup()
	})
}

func Test_Connector_Driver(t *testing.T) {
	connector, _, err := New("postgres://postgres:postgres@localhost:5432")
	require.NoError(t, err)
	driver := connector.Driver()
	require.NotEmpty(t, driver)
}
