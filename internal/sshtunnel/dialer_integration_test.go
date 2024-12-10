package sshtunnel_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"net"
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
	testcontainers_sqlserver "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/sqlserver"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
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
	dialerConfig := sshtunnel.DefaultSSHDialerConfig()
	dialerConfig.KeepAliveInterval = 1 * time.Second
	dialer := sshtunnel.NewLazySSHDialer(addr, cconfig, dialerConfig, testutil.GetConcurrentTestLogger(t))
	defer dialer.Close()

	t.Run("postgres", func(t *testing.T) {
		t.Parallel()

		container, err := tcpostgres.NewPostgresTestContainer(ctx)
		if err != nil {
			t.Fatal(err)
		}
		require.NoError(t, err)

		connector, cleanup, err := postgrestunconnector.New(
			container.URL,
			postgrestunconnector.WithDialer(dialer),
		)
		require.NoError(t, err)
		defer cleanup()

		requireDbConnects(t, connector)
	})

	t.Run("mysql", func(t *testing.T) {
		t.Parallel()

		container, err := tcmysql.NewMysqlTestContainer(ctx)
		if err != nil {
			t.Fatal(err)
		}

		connector, cleanup, err := mysqltunconnector.New(
			container.URL,
			mysqltunconnector.WithDialer(dialer),
		)
		require.NoError(t, err)
		defer cleanup()

		requireDbConnects(t, connector)
	})

	t.Run("mssql", func(t *testing.T) {
		t.Parallel()
		container, err := testcontainers_sqlserver.NewMssqlTestContainer(ctx)
		if err != nil {
			t.Fatal(err)
		}

		connector, cleanup, err := mssqltunconnector.New(container.URL, mssqltunconnector.WithDialer(dialer))
		require.NoError(t, err)
		defer cleanup()

		requireDbConnects(t, connector)
	})
}

func Test_NewLazySSHDialer_With_Tls(t *testing.T) {
	t.Parallel()
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}

	ctx := context.Background()

	addr := ":2225"
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
	dialerConfig := sshtunnel.DefaultSSHDialerConfig()
	dialerConfig.KeepAliveInterval = 1 * time.Second
	dialer := sshtunnel.NewLazySSHDialer(addr, cconfig, dialerConfig, testutil.GetConcurrentTestLogger(t))
	defer dialer.Close()

	t.Run("postgres", func(t *testing.T) {
		t.Parallel()

		container, err := tcpostgres.NewPostgresTestContainer(ctx, tcpostgres.WithTls())
		if err != nil {
			t.Fatal(err)
		}
		require.NoError(t, err)

		tlsConfig, err := container.GetClientTlsConfig(ctx)
		require.NoError(t, err)

		connector, cleanup, err := postgrestunconnector.New(
			container.URL,
			postgrestunconnector.WithDialer(dialer),
			postgrestunconnector.WithTLSConfig(tlsConfig),
		)
		require.NoError(t, err)
		defer cleanup()

		requireDbConnects(t, connector)
	})

	t.Run("mysql", func(t *testing.T) {
		t.Parallel()

		container, err := tcmysql.NewMysqlTestContainer(ctx, tcmysql.WithTls())
		require.NoError(t, err)

		tlsConfig, err := container.GetClientTlsConfig(ctx)
		require.NoError(t, err)

		connector, cleanup, err := mysqltunconnector.New(
			container.URL,
			mysqltunconnector.WithDialer(dialer),
			mysqltunconnector.WithTLSConfig(tlsConfig),
		)
		require.NoError(t, err)
		defer cleanup()

		requireDbConnects(t, connector)
	})

	t.Run("mssql", func(t *testing.T) {
		t.Parallel()

		container, err := testcontainers_sqlserver.NewMssqlTestContainer(ctx, testcontainers_sqlserver.WithTls())
		require.NoError(t, err)

		tlsConfig, err := container.GetClientTlsConfig(ctx)
		require.NoError(t, err)

		connector, cleanup, err := mssqltunconnector.New(
			container.URL,
			mssqltunconnector.WithDialer(dialer),
			mssqltunconnector.WithTLSConfig(tlsConfig),
		)
		require.NoError(t, err)
		defer cleanup()

		requireDbConnects(t, connector)
	})
}

func Test_SSHDialerResilience(t *testing.T) {
	t.Parallel()
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}

	ctx := context.Background()

	container, err := tcpostgres.NewPostgresTestContainer(ctx)
	if err != nil {
		t.Fatal((err))
	}

	pgHost, err := container.TestContainer.Host(ctx)
	require.NoError(t, err)
	pgPort, err := container.TestContainer.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	pgAddr := net.JoinHostPort(pgHost, pgPort.Port())

	t.Run("handles server disconnect and reconnects", func(t *testing.T) {
		addr := ":2223"
		server := newSshForwardServer(t, addr)
		serverShutdown := make(chan struct{})

		go func() {
			err := server.ListenAndServe()
			if err != nil && err != gssh.ErrServerClosed {
				panic(err)
			}
			close(serverShutdown)
		}()

		time.Sleep(100 * time.Millisecond)

		cconfig := &ssh.ClientConfig{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         5 * time.Second,
		}

		// Configure short retry intervals for testing
		dialer := sshtunnel.NewLazySSHDialer(addr, cconfig, &sshtunnel.SSHDialerConfig{
			MaxRetries:        3,
			InitialBackoff:    50 * time.Millisecond,
			KeepAliveInterval: 100 * time.Millisecond,
			KeepAliveTimeout:  50 * time.Millisecond,
		}, testutil.GetConcurrentTestLogger(t))
		defer dialer.Close()

		// Test initial connection
		conn, err := dialer.DialContext(ctx, "tcp", pgAddr)
		require.NoError(t, err)
		require.NotNil(t, conn)
		conn.Close()

		// Shutdown server
		err = server.Close()
		require.NoError(t, err)
		<-serverShutdown

		// Start new server
		server = newSshForwardServer(t, addr)
		go func() {
			err := server.ListenAndServe()
			if err != nil && err != gssh.ErrServerClosed {
				panic(err)
			}
		}()

		time.Sleep(100 * time.Millisecond)

		// Test reconnection
		conn, err = dialer.DialContext(ctx, "tcp", pgAddr)
		require.NoError(t, err)
		require.NotNil(t, conn)
		conn.Close()
	})

	t.Run("verifies keepalive maintains connection", func(t *testing.T) {
		addr := ":2224"
		server := newSshForwardServer(t, addr)
		keepaliveReceived := make(chan struct{})

		// Enhance server to detect keepalive requests
		server.RequestHandlers["keepalive@openssh.com"] = func(ctx gssh.Context, srv *gssh.Server, req *ssh.Request) (bool, []byte) {
			select {
			case keepaliveReceived <- struct{}{}:
			default:
			}
			return true, nil
		}

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

		dialer := sshtunnel.NewLazySSHDialer(addr, cconfig, &sshtunnel.SSHDialerConfig{
			MaxRetries:        3,
			InitialBackoff:    50 * time.Millisecond,
			KeepAliveInterval: 100 * time.Millisecond,
			KeepAliveTimeout:  50 * time.Millisecond,
		}, testutil.GetConcurrentTestLogger(t))
		defer dialer.Close()

		// Establish initial connection
		conn, err := dialer.DialContext(ctx, "tcp", pgAddr)
		require.NoError(t, err)
		require.NotNil(t, conn)
		defer conn.Close()

		// Wait for keepalive
		select {
		case <-keepaliveReceived:
			// Success - received keepalive
		case <-time.After(500 * time.Millisecond):
			t.Fatal("no keepalive received within timeout")
		}
	})

	t.Run("handles max retries exhaustion", func(t *testing.T) {
		addr := ":2325"

		cconfig := &ssh.ClientConfig{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         100 * time.Millisecond, // Short timeout for faster test
		}

		dialer := sshtunnel.NewLazySSHDialer(addr, cconfig, &sshtunnel.SSHDialerConfig{
			MaxRetries:        2,
			InitialBackoff:    50 * time.Millisecond,
			KeepAliveInterval: 100 * time.Millisecond,
			KeepAliveTimeout:  50 * time.Millisecond,
		}, testutil.GetConcurrentTestLogger(t))
		defer dialer.Close()

		// Attempt connection - should fail after retries
		conn, err := dialer.DialContext(ctx, "tcp", "localhost:5555") // port that is not being used
		require.Error(t, err)
		require.Nil(t, conn)
		require.Contains(t, err.Error(), "unable to dial ssh server after 2 attempts")
	})

	t.Run("cancels retry attempts on context cancellation", func(t *testing.T) {
		addr := ":2226"

		cconfig := &ssh.ClientConfig{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         100 * time.Millisecond,
		}

		dialer := sshtunnel.NewLazySSHDialer(addr, cconfig, &sshtunnel.SSHDialerConfig{
			MaxRetries:        5,
			InitialBackoff:    1 * time.Second, // Longer backoff to ensure we can cancel
			KeepAliveInterval: 100 * time.Millisecond,
			KeepAliveTimeout:  50 * time.Millisecond,
		}, testutil.GetConcurrentTestLogger(t))
		defer dialer.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		startTime := time.Now()
		conn, err := dialer.DialContext(ctx, "tcp", "localhost:5555")
		elapsed := time.Since(startTime)

		require.Error(t, err)
		require.Nil(t, conn)
		require.Contains(t, err.Error(), context.DeadlineExceeded.Error())
		require.Less(t, elapsed, 1*time.Second) // Ensure we didn't wait for the full backoff
	})

	t.Run("detects and replaces dead client", func(t *testing.T) {
		addr := ":2227"
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

		dialer := sshtunnel.NewLazySSHDialer(addr, cconfig, &sshtunnel.SSHDialerConfig{
			MaxRetries:        3,
			InitialBackoff:    50 * time.Millisecond,
			KeepAliveInterval: 30 * time.Second,
			KeepAliveTimeout:  50 * time.Millisecond,
		}, testutil.GetConcurrentTestLogger(t))
		defer dialer.Close()

		// Establish initial connection
		conn, err := dialer.DialContext(ctx, "tcp", pgAddr)
		require.NoError(t, err)
		require.NotNil(t, conn)
		conn.Close()

		// Abruptly stop the server without clean shutdown
		server.Close()

		// Start new server on same port
		server = newSshForwardServer(t, addr)
		go func() {
			err := server.ListenAndServe()
			if err != nil && err != gssh.ErrServerClosed {
				panic(err)
			}
		}()

		time.Sleep(100 * time.Millisecond)

		// Next connection attempt should detect dead client and establish new one
		conn, err = dialer.DialContext(ctx, "tcp", pgAddr)
		require.NoError(t, err)
		require.NotNil(t, conn)
		conn.Close()
	})
}

func requireDbConnects(t testing.TB, connector driver.Connector) {
	db := sql.OpenDB(connector)
	defer db.Close()

	err := db.Ping()
	require.NoError(t, err)
	_, err = db.Exec("SELECT 1")
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
