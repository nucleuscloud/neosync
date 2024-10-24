package sshtunnel_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	gssh "github.com/gliderlabs/ssh"
	"github.com/nucleuscloud/neosync/internal/sshtunnel"
	"github.com/nucleuscloud/neosync/internal/sshtunnel/connectors/mssqltunconnector"
	"github.com/nucleuscloud/neosync/internal/sshtunnel/connectors/mysqltunconnector"
	"github.com/nucleuscloud/neosync/internal/sshtunnel/connectors/postgrestunconnector"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcmysql "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/mysql"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"

	testmssql "github.com/testcontainers/testcontainers-go/modules/mssql"
)

func Test_NewLazySSHDialer(t *testing.T) {
	t.Parallel()
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}

	ctx := context.Background()

	addr := ":2222"
	server := newSshForwardServer(t, addr)

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != gssh.ErrServerClosed {
			panic(err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	cconfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	dialer := sshtunnel.NewLazySSHDialer(addr, cconfig)
	defer dialer.Close()

	t.Run("postgres", func(t *testing.T) {
		t.Parallel()

		container, err := tcpostgres.NewPostgresTestContainer(ctx)
		if err != nil {
			panic(err)
		}
		require.NoError(t, err)

		connector, cleanup, err := postgrestunconnector.New(dialer, container.URL)
		require.NoError(t, err)
		defer cleanup()

		requireDbConnects(t, connector)
	})

	t.Run("mysql", func(t *testing.T) {
		t.Parallel()

		container, err := tcmysql.NewMysqlTestContainer(ctx)
		if err != nil {
			panic(err)
		}

		connector, cleanup, err := mysqltunconnector.New(dialer, container.URL)
		require.NoError(t, err)
		defer cleanup()

		requireDbConnects(t, connector)
	})

	t.Run("mssql", func(t *testing.T) {
		t.Parallel()
		container, err := testmssql.Run(ctx,
			"mcr.microsoft.com/mssql/server:2022-latest",
			testmssql.WithAcceptEULA(),
			testmssql.WithPassword("mssqlPASSword1"),
		)
		require.NoError(t, err)
		connstr, err := container.ConnectionString(ctx)
		require.NoError(t, err)

		connector, cleanup, err := mssqltunconnector.New(dialer, connstr)
		require.NoError(t, err)
		defer cleanup()

		requireDbConnects(t, connector)
	})
}

func requireDbConnects(t testing.TB, connector driver.Connector) {
	db := sql.OpenDB(connector)
	defer db.Close()

	err := db.Ping()
	require.NoError(t, err)
}

func newSshForwardServer(t testing.TB, addr string) *gssh.Server {
	forwardHandler := &gssh.ForwardedTCPHandler{}
	return &gssh.Server{
		Addr: addr,
		Handler: gssh.Handler(func(s gssh.Session) {
			select {}
		}),
		LocalPortForwardingCallback: gssh.LocalPortForwardingCallback(func(ctx gssh.Context, destinationHost string, destinationPort uint32) bool {
			t.Logf("Accepted forward %s:%d\n", destinationHost, destinationPort)
			return true
		}),
		ReversePortForwardingCallback: gssh.ReversePortForwardingCallback(func(ctx gssh.Context, destinationHost string, destinationPort uint32) bool {
			t.Logf("attempt to bind %s:%d granted\n", destinationHost, destinationPort)
			return true
		}),
		RequestHandlers: map[string]gssh.RequestHandler{
			"tcpip-forward":        forwardHandler.HandleSSHRequest,
			"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
		},
		ChannelHandlers: map[string]gssh.ChannelHandler{
			"direct-tcpip": gssh.DirectTCPIPHandler,
			"session":      gssh.DefaultSessionHandler,
		},
	}
}
