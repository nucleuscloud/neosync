package mysqltunconnector

import (
	"crypto/tls"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		connector, cleanup, err := New("foo:bar@tcp(localhost:3306)/mydb")
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
		connector, cleanup, err := New("foo:bar@tcp(localhost:3306)/mydb", WithDialer(&net.Dialer{}))
		require.NoError(t, err)
		require.NotNil(t, cleanup)
		require.NotNil(t, connector)
		cleanup()
	})
	t.Run("WithTls", func(t *testing.T) {
		connector, cleanup, err := New("foo:bar@tcp(localhost:3306)/mydb", WithTLSConfig(&tls.Config{}))
		require.NoError(t, err)
		require.NotNil(t, cleanup)
		require.NotNil(t, connector)
		cleanup()
	})
}
