package sshtunnel

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

type Sshtunnel struct {
	local  *Endpoint
	server *Endpoint
	remote *Endpoint

	config *ssh.ClientConfig

	maxConnectionAttempts uint
	close                 chan any
	isOpen                atomic.Bool

	shutdowns *sync.Map

	sshclient *ssh.Client
	sshMu     *sync.RWMutex
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
		close: make(chan any),

		local:  local,
		server: tunnel,
		remote: destination,

		maxConnectionAttempts: maxConnectionAttempts,

		config: &ssh.ClientConfig{
			User:            tunnel.User,
			Auth:            authmethods,
			HostKeyCallback: getHostKeyCallback(serverPublicKey),
			Timeout:         30 * time.Second,
		},

		shutdowns: &sync.Map{},

		sshMu: &sync.RWMutex{},
	}
}

// After a tunnel has started, this will return the auto-generated port (if 0 was passed in)
func (t *Sshtunnel) GetLocalHostPort() (host string, port int) {
	return t.local.Host, t.local.Port
}

func getHostKeyCallback(key ssh.PublicKey) ssh.HostKeyCallback {
	if key == nil {
		return ssh.InsecureIgnoreHostKey() //nolint
	}
	return ssh.FixedHostKey(key)
}

func (t *Sshtunnel) Start(logger *slog.Logger) (chan any, error) {
	listener, err := net.Listen("tcp", t.local.String())
	if err != nil {
		return nil, fmt.Errorf("unable to listen to local endpoint: %w", err)
	}
	ready := make(chan any)
	go t.serve(listener, ready, logger)
	return ready, nil
}

func (t *Sshtunnel) serve(listener net.Listener, ready chan<- any, logger *slog.Logger) {
	defer func() {
		if err := listener.Close(); err != nil {
			if !errors.Is(err, net.ErrClosed) {
				logger.Error("failed to close tunnel listener", "error", err)
			}
		}
	}()

	t.local.Port = listener.Addr().(*net.TCPAddr).Port
	t.isOpen.Store(true)
	close(ready)

	for {
		if !t.isOpen.Load() {
			break
		}

		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				logger.Debug("listener closed, stopping serve loop")
				return
			}
			logger.Error("failed to accept local connection for tunnel", "error", err)
			continue
		}

		logger.Debug("accepted new connection for tunnel", "remoteAddr", conn.RemoteAddr().String())
		sessionId := uuid.NewString()
		shutdown := make(chan any)
		t.shutdowns.Store(sessionId, shutdown)

		go func() {
			defer func() {
				t.shutdowns.Delete(sessionId)
				if err := conn.Close(); err != nil {
					if !errors.Is(err, net.ErrClosed) {
						logger.Error("failed to close tunnel connection for session", "error", err, "sessionId", sessionId)
					}
				}
			}()
			select {
			case <-t.close:
				logger.Debug("received close signal, closing connection", "sessionId", sessionId)
			case <-shutdown:
				logger.Debug("received shutdown signal for session", "sessionId", sessionId)
			default:
				t.forward(conn, sessionId, shutdown, logger.With("sessionId", sessionId))
			}
		}()
	}

	logger.Debug("tunnel closed")
}

// Takes the local connection, dials into the SSH server, connects to the remote host with that connection,
// and then forwards the traffic from the local connection to the remote connection
func (t *Sshtunnel) forward(localConnection net.Conn, sessionId string, shutdown <-chan any, logger *slog.Logger) {
	sshClient, err := t.getSshClient(t.server.String(), t.config, t.maxConnectionAttempts, logger)
	if err != nil {
		if err := localConnection.Close(); err != nil {
			logger.Error(fmt.Sprintf("failed to close local connection: %v", err))
			return
		}
		logger.Error(fmt.Sprintf("unable to reach SSH server: %v", err))
		return
	}

	remoteConnection, err := sshClient.Dial("tcp", t.remote.String())
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
	logger.Debug(fmt.Sprintf("connected to %s", t.remote.String()))

	// buffering so that we don't block the copyConnection when it sends its result
	done := make(chan error, 2)
	go func() {
		select {
		case <-shutdown:
			logger.Debug("issued shutdown of tunnel")
			localConnection.Close()
			remoteConnection.Close()
			t.closeSshClient()
			logger.Debug("issued shutdown, closing local, remove, and ssh connections")
		case <-done:
			t.shutdowns.Delete(sessionId)
			localConnection.Close()
			remoteConnection.Close()
			logger.Debug("connection done, closed local and remote connections")
		}
	}()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		done <- copyConnection(localConnection, remoteConnection, logger.With("direction", "remote->local"))
	}()
	go func() {
		defer wg.Done()
		done <- copyConnection(remoteConnection, localConnection, logger.With("direction", "local->remote"))
	}()
	wg.Wait()
	logger.Debug("tunnel forwarding complete for session")
}

func (t *Sshtunnel) closeSshClient() {
	t.sshMu.Lock()
	defer t.sshMu.Unlock()
	if t.sshclient == nil {
		return
	}
	client := t.sshclient
	t.sshclient = nil
	client.Close()
}

func (s *Sshtunnel) getSshClient(
	addr string,
	config *ssh.ClientConfig,
	maxAttempts uint,
	logger *slog.Logger,
) (*ssh.Client, error) {
	s.sshMu.RLock()
	client := s.sshclient
	s.sshMu.RUnlock()
	if client != nil {
		return client, nil
	}
	s.sshMu.Lock()
	defer s.sshMu.Unlock()
	if s.sshclient != nil {
		return s.sshclient, nil
	}
	client, err := getSshClient(addr, config, maxAttempts, logger)
	if err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("conntected to %s", addr))
	s.sshclient = client
	return client, nil
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
func copyConnection(writer, reader net.Conn, logger *slog.Logger) error {
	_, err := io.Copy(writer, reader)
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			// This can be a common error if the thing using the ssh connection was abruptly closed.
			// This is common if a user is trying to test their database connection, but they've given Neosync bad credentials
			// or something else that causes the server to force close the client connection
			logger.Warn("connection was closed before reaching end of input", "error", err)
		} else {
			logger.Error("unexpected error while streaming through tunnel", "error", err)
		}
	} else {
		logger.Debug("ssh tunnel stream completed successfully")
	}
	return err
}

func (t *Sshtunnel) Close() {
	if !t.isOpen.CompareAndSwap(true, false) {
		return
	}
	close(t.close)
	t.shutdowns.Range(func(key, value any) bool {
		if ch, ok := value.(chan any); ok {
			close(ch)
		}
		return true
	})
}
