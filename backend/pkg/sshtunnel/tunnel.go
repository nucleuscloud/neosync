package sshtunnel

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

type Sshtunnel struct {
	Local  *Endpoint
	Server *Endpoint
	Remote *Endpoint

	Config *ssh.ClientConfig

	maxConnectionAttempts uint
	close                 chan any
	isOpen                bool

	shutdowns *sync.Map
}

func New(
	tunnel *Endpoint,
	auth ssh.AuthMethod,
	destination *Endpoint,
	local *Endpoint,
	maxConnectionAttempts uint,
	serverPublicKey ssh.PublicKey,
) *Sshtunnel {
	authmethods := []ssh.AuthMethod{}
	if auth != nil {
		authmethods = append(authmethods, auth)
	}
	return &Sshtunnel{
		close: make(chan any, 1),

		Local:  local,
		Server: tunnel,
		Remote: destination,

		maxConnectionAttempts: maxConnectionAttempts,

		Config: &ssh.ClientConfig{
			User:            tunnel.User,
			Auth:            authmethods,
			HostKeyCallback: getHostKeyCallback(serverPublicKey),
		},

		shutdowns: &sync.Map{},
	}
}

func getHostKeyCallback(key ssh.PublicKey) ssh.HostKeyCallback {
	if key == nil {
		return ssh.InsecureIgnoreHostKey() //nolint
	}
	return ssh.FixedHostKey(key)
}

func (t *Sshtunnel) Start(logger *slog.Logger) (chan any, error) {
	listener, err := net.Listen("tcp", t.Local.String())
	if err != nil {
		return nil, fmt.Errorf("unable to listen to local endpoint: %w", err)
	}
	ready := make(chan any)
	go t.serve(listener, ready, logger)
	return ready, nil
}

func (t *Sshtunnel) serve(listener net.Listener, ready chan<- any, logger *slog.Logger) {
	t.Local.Port = listener.Addr().(*net.TCPAddr).Port
	t.isOpen = true
	hasSignaledReady := false

	for {
		if !t.isOpen {
			break
		}

		c := make(chan net.Conn)

		sessionId := uuid.NewString()
		go newConnectionWaiter(listener, c, ready, hasSignaledReady, logger) // begins accepting connections and sends the connection onto the channel
		hasSignaledReady = true
		logger.Debug(fmt.Sprintf("listening for new local connections on %s", t.Local.String()))
		shutdown := make(chan any, 1)
		t.shutdowns.Store(sessionId, shutdown)
		select {
		case <-t.close:
			logger.Debug("received close signal from client...")
			t.isOpen = false
			go func() {
				t.shutdowns.Range(func(key, value any) bool {
					sd, ok := value.(chan any)
					if ok {
						sd <- struct{}{}
					} else {
						logger.Warn(fmt.Sprintf("was unable to cast shutdown value to chan any. was %T", sd))
					}
					return true
				})
				if err := listener.Close(); err != nil {
					logger.Error(err.Error())
				}
			}()
		case conn := <-c:
			logger.Debug(fmt.Sprintf("accepted connection local: %s, remote: %s", conn.LocalAddr().String(), conn.RemoteAddr().String()))
			go t.forward(conn, sessionId, shutdown, logger.With("tunnelSessionId", sessionId))
		}
	}

	logger.Debug("tunnel closed")
}

// Takes the local connection, dials into the SSH server, connects to the remote host with that connection,
// and then forwards the traffic from the local connection to the remote connection
func (t *Sshtunnel) forward(localConnection net.Conn, sessionId string, shutdown <-chan any, logger *slog.Logger) {
	sshClient, err := getSshClient(t.Server.String(), t.Config, t.maxConnectionAttempts, logger)
	if err != nil {
		if err := localConnection.Close(); err != nil {
			logger.Error(fmt.Sprintf("failed to close local connection: %v", err))
			return
		}
		logger.Error(fmt.Sprintf("unable to reach SSH server: %v", err))
		return
	}

	logger.Debug(fmt.Sprintf("connected to %s (1 of 2)", t.Server.String()))

	remoteConnection, err := sshClient.Dial("tcp", t.Remote.String())
	if err != nil {
		logger.Error(fmt.Sprintf("remote dial error: %s", err))
		if err := sshClient.Close(); err != nil {
			logger.Error(fmt.Sprintf("failed to close server connection: %v", err))
		}
		if err := localConnection.Close(); err != nil {
			logger.Error(fmt.Sprintf("failed to close local connection: %v", err))
		}
		return
	}
	logger.Debug(fmt.Sprintf("connected to %s (2 of 2)", t.Remote.String()))

	cleanup := func() {
		// not really necessary since the connections are done
		// but to be safe, let's clean up anyways
		localConnection.Close()
		remoteConnection.Close()
		sshClient.Close()
	}

	// buffering so that we don't block the copyConnection when it sends its result
	done := make(chan error, 2)
	go func() {
		select {
		case <-shutdown:
			logger.Debug("issued shutdown of tunnel")
		case <-done:
			t.shutdowns.Delete(sessionId)
		}
		cleanup()
		logger.Debug("connection done, closed local, remote, and ssh connection")
	}()
	go copyConnection(localConnection, remoteConnection, done, logger.With("direction", "remote->local"))
	go copyConnection(remoteConnection, localConnection, done, logger.With("direction", "local->remote"))
}

func getSshClient(
	addr string,
	config *ssh.ClientConfig,
	maxAttempts uint,
	logger *slog.Logger,
) (*ssh.Client, error) {
	var sshClient *ssh.Client
	var err error
	var attemptsLeft = maxAttempts
	for {
		sshClient, err = ssh.Dial("tcp", addr, config)
		if err != nil {
			attemptsLeft--
			if attemptsLeft <= 0 {
				logger.Error(fmt.Sprintf("server dial error: %v: exceeded %d attempts", err, maxAttempts))
				return nil, err
			}
			logger.Warn(fmt.Sprintf("server dial error: %v: attempt %d/%d", err, maxAttempts-attemptsLeft, maxAttempts))
		} else {
			break
		}
	}
	return sshClient, err
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

func newConnectionWaiter(
	listener net.Listener,
	c chan<- net.Conn,
	ready chan<- any,
	hasSignaledReady bool,
	logger *slog.Logger,
) {
	go func() {
		if !hasSignaledReady {
			logger.Debug("notifying ready channel")
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
