package sshtunnel

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/google/uuid"
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
	authmethods := []ssh.AuthMethod{}
	if auth != nil {
		authmethods = append(authmethods, auth)
	}
	return &Sshtunnel{
		logger: logger,
		close:  make(chan any),

		Local:  local,
		Server: tunnel,
		Remote: destination,

		maxConnectionAttempts: maxConnectionAttempts,

		Config: &ssh.ClientConfig{
			User:            tunnel.User,
			Auth:            authmethods,
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
		go newConnectionWaiter(listener, c, ready, hasSignaledReady, t.logger) // begins accepting connections and sends the connection onto the channel
		hasSignaledReady = true
		t.logger.Debug(fmt.Sprintf("listening for new local connections on %s", t.Local.String()))

		select {
		case <-t.close:
			t.logger.Debug("received close signal from client...")
			t.isOpen = false
			go func() {
				if err := listener.Close(); err != nil {
					t.logger.Error(err.Error())
				}
			}()
		case conn := <-c:
			t.logger.Debug(fmt.Sprintf("accepted connection local: %s, remote: %s", conn.LocalAddr().String(), conn.RemoteAddr().String()))
			go t.forward(conn)
		}
	}

	t.logger.Debug("tunnel closed")
}

// Takes the local connection, dials into the SSH server, connects to the remote host with that connection,
// and then forwards the traffic from the local connection to the remote connection
func (t *Sshtunnel) forward(localConnection net.Conn) {
	var sshClient *ssh.Client
	var err error
	var attemptsLeft = t.maxConnectionAttempts

	for {
		sshClient, err = ssh.Dial("tcp", t.Server.String(), t.Config)
		if err != nil {
			attemptsLeft--

			if attemptsLeft <= 0 {
				t.logger.Error(fmt.Sprintf("server dial error: %v: exceeded %d attempts", err, t.maxConnectionAttempts))

				if err := localConnection.Close(); err != nil {
					t.logger.Error(fmt.Sprintf("failed to close local connection: %v", err))
					return
				}
				t.logger.Debug("closed local connection")
				return
			}
			t.logger.Warn(fmt.Sprintf("server dial error: %v: attempt %d/%d", err, t.maxConnectionAttempts-attemptsLeft, t.maxConnectionAttempts))
		} else {
			break
		}
	}

	t.logger.Debug(fmt.Sprintf("connected to %s (1 of 2)", t.Server.String()))

	remoteConnection, err := sshClient.Dial("tcp", t.Remote.String())
	if err != nil {
		t.logger.Error(fmt.Sprintf("remote dial error: %s", err))
		if err := sshClient.Close(); err != nil {
			t.logger.Error(fmt.Sprintf("failed to close server connection: %v", err))
		}
		if err := localConnection.Close(); err != nil {
			t.logger.Error(fmt.Sprintf("failed to close local connection: %v", err))
		}
		return
	}
	t.logger.Debug(fmt.Sprintf("connected to %s (2 of 2)", t.Remote.String()))

	done := make(chan error)

	cleanup := func() {
		// not really necessary since the connections are done
		// but to be safe, let's clean up anyways
		localConnection.Close()
		remoteConnection.Close()
		sshClient.Close()
	}
	go func() {
		<-done
		close(done) // we only really care about one of the connections closing before we issue a cleanup.
		cleanup()
		t.logger.Debug("connection done, closed local, remote, and ssh connection")
	}()
	go copyConnection(localConnection, remoteConnection, done, t.logger.With("direction", "remote->local"))
	go copyConnection(remoteConnection, localConnection, done, t.logger.With("direction", "local->remote"))
}

// Writer is what receives the input (dst), reader is what the input is read from (src)
func copyConnection(writer, reader net.Conn, done chan<- error, logger *slog.Logger) {
	_, err := io.Copy(writer, reader)
	if err != nil {
		logger.Error(fmt.Sprintf("io.Copy error: %s", err))
	} else {
		logger.Debug("io.Copy returned successfully")
	}
	done <- err
}

func newConnectionWaiter(listener net.Listener, c chan<- net.Conn, ready chan<- any, hasSignaledReady bool, logger *slog.Logger) {
	go func() {
		if !hasSignaledReady {
			sessionId := uuid.NewString()
			logger.Debug("notifying ready channel", "sessionId", sessionId)
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
	logger.Debug("sending connection to channel")
	c <- conn
}

func (t *Sshtunnel) Close() {
	t.close <- struct{}{}
}
