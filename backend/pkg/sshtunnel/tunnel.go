package sshtunnel

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"

	"golang.org/x/crypto/ssh"
)

type Sshtunnel struct {
	logger *slog.Logger

	Local  *Endpoint
	Server *Endpoint
	Remote *Endpoint

	Config *ssh.ClientConfig

	maxConnectionAttempts uint
	close                 chan any
	isOpen                bool

	connections       []net.Conn
	serverConnections []*ssh.Client
}

func New(
	tunnel *Endpoint,
	auth ssh.AuthMethod,
	destination *Endpoint,
	local *Endpoint,
	maxConnectionAttempts uint,
	serverPublicKey ssh.PublicKey,
	logger *slog.Logger,
) *Sshtunnel {
	return &Sshtunnel{
		logger: logger,
		close:  make(chan any),

		Local:  local,
		Server: tunnel,
		Remote: destination,

		maxConnectionAttempts: maxConnectionAttempts,

		Config: &ssh.ClientConfig{
			User:            tunnel.User,
			Auth:            []ssh.AuthMethod{auth},
			HostKeyCallback: getHostKeyCallback(serverPublicKey),
		},
	}
}

func getHostKeyCallback(key ssh.PublicKey) ssh.HostKeyCallback {
	if key == nil {
		return ssh.InsecureIgnoreHostKey() //nolint
	}
	return ssh.FixedHostKey(key)
}

func (t *Sshtunnel) Start() (chan any, error) {
	listener, err := net.Listen("tcp", t.Local.String())
	if err != nil {
		return nil, fmt.Errorf("unable to listen to local endpoint: %w", err)
	}
	ready := make(chan any)
	go t.Serve(listener, ready)
	return ready, nil
}

func (t *Sshtunnel) Serve(listener net.Listener, ready chan<- any) {
	t.Local.Port = listener.Addr().(*net.TCPAddr).Port
	t.isOpen = true
	hasSignaledReady := false

	for {
		if !t.isOpen {
			break
		}

		c := make(chan net.Conn)
		go newConnectionWaiter(listener, c, ready, hasSignaledReady, t.logger) // beings accepting connections and sends the connection onto the channel
		hasSignaledReady = true
		t.logger.Info(fmt.Sprintf("listening for new connections on %s", t.Local.String()))

		select {
		case <-t.close:
			t.logger.Info("received close signal...")
			t.isOpen = false
			go func() {
				if err := listener.Close(); err != nil {
					t.logger.Error(err.Error())
				}
			}()
		case conn := <-c:
			t.connections = append(t.connections, conn)
			t.logger.Info(fmt.Sprintf("accepted connection from %s", conn.RemoteAddr().String()))
			go t.forward(conn)
		}
	}
	total := len(t.connections)
	t.logger.Info(fmt.Sprintf("attempting to close %d connections", total))
	for i, conn := range t.connections {
		t.logger.Info(fmt.Sprintf("closing the netConn (%d of %d)", i+1, total))
		err := conn.Close()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				continue
			}
			t.logger.Error(fmt.Sprintf("failed to close connection: %s", err.Error()))
		}
	}
	total = len(t.serverConnections)
	t.logger.Info(fmt.Sprintf("attempting to close %d server connections", total))
	for i, conn := range t.serverConnections {
		t.logger.Info(fmt.Sprintf("closing the serverConn (%d of %d)", i+1, total))
		err := conn.Close()
		if err != nil {
			t.logger.Error(fmt.Sprintf("failed to close server connection: %s", err.Error()))
		}
	}

	t.logger.Info("tunnel closed")
}

func (t *Sshtunnel) forward(localConnection net.Conn) {
	var serverConn *ssh.Client
	var err error
	var attemptsLeft = t.maxConnectionAttempts

	for {
		serverConn, err = ssh.Dial("tcp", t.Server.String(), t.Config)
		if err != nil {
			attemptsLeft--

			if attemptsLeft <= 0 {
				t.logger.Info(fmt.Sprintf("server dial error: %v: exceeded %d attempts", err, t.maxConnectionAttempts))

				if err := localConnection.Close(); err != nil {
					t.logger.Error(fmt.Sprintf("failed to close local connection: %v", err))
					return
				}
				t.logger.Info("closed local connection")
				return
			}
			t.logger.Error(fmt.Sprintf("server dial error: %v: attempt %d/%d", err, t.maxConnectionAttempts-attemptsLeft, t.maxConnectionAttempts))
		} else {
			break
		}
	}

	t.logger.Info(fmt.Sprintf("connected to %s (1 of 2)", t.Server.String()))
	t.serverConnections = append(t.serverConnections, serverConn)

	remoteConnection, err := serverConn.Dial("tcp", t.Remote.String())
	if err != nil {
		t.logger.Error(fmt.Sprintf("remote dial error: %s", err))
		if err := serverConn.Close(); err != nil {
			t.logger.Error(fmt.Sprintf("failed to close server connection: %v", err))
		}
		if err := localConnection.Close(); err != nil {
			t.logger.Error(fmt.Sprintf("failed to close local connection: %v", err))
		}
		return
	}
	t.connections = append(t.connections, remoteConnection)
	t.logger.Info(fmt.Sprintf("connected to %s (2 of 2)", t.Remote.String()))
	go copyConnection(localConnection, remoteConnection, t.logger)
	go copyConnection(remoteConnection, localConnection, t.logger)
}

func copyConnection(writer, reader net.Conn, logger *slog.Logger) {
	_, err := io.Copy(writer, reader)
	if err != nil {
		logger.Error(fmt.Sprintf("io.Copy error: %s", err))
	}
}

func newConnectionWaiter(listener net.Listener, c chan<- net.Conn, ready chan<- any, hasSignaledReady bool, logger *slog.Logger) {
	go func() {
		if !hasSignaledReady {
			logger.Info("notifying ready channel")
			ready <- struct{}{}
		}
	}()
	conn, err := listener.Accept()
	if err != nil {
		if !errors.Is(err, net.ErrClosed) {
			logger.Error(fmt.Sprintf("unable to accept new connection: %v", err))
		}
		return
	}
	logger.Info("sending connection to channel")
	c <- conn
}

func (t *Sshtunnel) Close() {
	t.close <- struct{}{}
}
