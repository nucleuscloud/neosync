package sshtunnel_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	gssh "github.com/gliderlabs/ssh"
	"github.com/nucleuscloud/neosync/internal/sshtunnel"
	"github.com/nucleuscloud/neosync/internal/sshtunnel/connectors/mssqltunconnector"
	"github.com/nucleuscloud/neosync/internal/sshtunnel/connectors/mysqltunconnector"
	"github.com/nucleuscloud/neosync/internal/sshtunnel/connectors/postgrestunconnector"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"

	"github.com/testcontainers/testcontainers-go"
	testmssql "github.com/testcontainers/testcontainers-go/modules/mssql"
	testmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	testpg "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/testcontainers/testcontainers-go/wait"
)

func Test_NewLazySSHDialer(t *testing.T) {
	t.Parallel()
	evkey := "INTEGRATION_TESTS_ENABLED"
	shouldRun := os.Getenv(evkey)
	if shouldRun != "1" {
		slog.Warn(fmt.Sprintf("skipping integration tests, set %s=1 to enable", evkey))
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

		container, err := testpg.Run(
			ctx,
			"postgres:15",
			postgres.WithDatabase("postgres"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).WithStartupTimeout(20*time.Second),
			),
		)
		require.NoError(t, err)
		connstr, err := container.ConnectionString(ctx, "sslmode=disable")
		require.NoError(t, err)

		connector, cleanup, err := postgrestunconnector.New(dialer, connstr)
		require.NoError(t, err)
		defer cleanup()

		requireDbConnects(t, connector)
	})

	t.Run("mysql", func(t *testing.T) {
		t.Parallel()

		container, err := testmysql.Run(ctx,
			"mysql:8.0.36",
			testmysql.WithDatabase("mydb"),
			testmysql.WithUsername("root"),
			testmysql.WithPassword("password"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("port: 3306  MySQL Community Server").
					WithOccurrence(1).WithStartupTimeout(20*time.Second),
			),
		)
		require.NoError(t, err)
		connstr, err := container.ConnectionString(ctx)
		require.NoError(t, err)

		connector, cleanup, err := mysqltunconnector.New(dialer, connstr)
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
