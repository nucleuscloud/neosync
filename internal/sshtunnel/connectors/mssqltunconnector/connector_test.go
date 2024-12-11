package mssqltunconnector

import (
	"crypto/tls"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		connector, cleanup, err := New("sqlserver://sa:myStr0ngP%40assword@localhost?database=master")
		require.NoError(t, err)
		require.NotNil(t, cleanup)
		require.NotNil(t, connector)
	})

	t.Run("invalid dsn", func(t *testing.T) {
		connector, cleanup, err := New("sqlserver://sa:myStr0ngP%40assword@localhost:invalidport")
		require.Error(t, err)
		require.Nil(t, cleanup)
		require.Nil(t, connector)
	})

	t.Run("WithDialer", func(t *testing.T) {
		connector, cleanup, err := New("sqlserver://sa:myStr0ngP%40assword@localhost?database=master", WithDialer(&net.Dialer{}))
		require.NoError(t, err)
		require.NotNil(t, cleanup)
		require.NotNil(t, connector)
	})
	t.Run("WithTls", func(t *testing.T) {
		connector, cleanup, err := New("sqlserver://sa:myStr0ngP%40assword@localhost?database=master", WithTLSConfig(&tls.Config{}))
		require.NoError(t, err)
		require.NotNil(t, cleanup)
		require.NotNil(t, connector)
	})
}
