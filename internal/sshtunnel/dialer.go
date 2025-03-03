package sshtunnel

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/nucleuscloud/neosync/internal/backoffutil"
	"golang.org/x/crypto/ssh"
)

type Dialer interface {
	ContextDialer
	Dial(network, addr string) (net.Conn, error)
}

type ContextDialer interface {
	DialContext(ctx context.Context, network, addr string) (net.Conn, error)
}

var _ Dialer = (*SSHDialer)(nil)

type SSHDialerConfig struct {
	getRetryOpts func(logger *slog.Logger) []backoff.RetryOption

	KeepAliveInterval time.Duration
	KeepAliveTimeout  time.Duration
}

type SSHDialer struct {
	addr string
	ccfg *ssh.ClientConfig

	dialCfg *SSHDialerConfig

	client   *ssh.Client
	clientmu *sync.Mutex
	logger   *slog.Logger
}

func DefaultSSHDialerConfig() *SSHDialerConfig {
	return &SSHDialerConfig{
		getRetryOpts: func(logger *slog.Logger) []backoff.RetryOption {
			backoffStrategy := backoff.NewExponentialBackOff()
			backoffStrategy.InitialInterval = 200 * time.Millisecond
			backoffStrategy.MaxInterval = 30 * time.Second
			backoffStrategy.Multiplier = 2
			backoffStrategy.RandomizationFactor = 0.3
			return []backoff.RetryOption{
				backoff.WithBackOff(backoffStrategy),
				backoff.WithMaxTries(10),
				backoff.WithMaxElapsedTime(5 * time.Minute),
				backoff.WithNotify(func(err error, d time.Duration) {
					logger.Warn(fmt.Sprintf("ssh error with retry: %s, retrying in %s", err.Error(), d.String()))
				}),
			}
		},

		KeepAliveInterval: 30 * time.Second,
		KeepAliveTimeout:  15 * time.Second,
	}
}

func NewLazySSHDialer(addr string, ccfg *ssh.ClientConfig, dialCfg *SSHDialerConfig, logger *slog.Logger) *SSHDialer {
	if dialCfg == nil {
		dialCfg = DefaultSSHDialerConfig()
	}
	return &SSHDialer{addr: addr, ccfg: ccfg, clientmu: &sync.Mutex{}, dialCfg: dialCfg, logger: logger}
}

func NewSSHDialer(client *ssh.Client, logger *slog.Logger) *SSHDialer {
	return &SSHDialer{client: client, clientmu: &sync.Mutex{}, logger: logger}
}

func (s *SSHDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get or create ssh client during DialContext: %w", err)
	}
	conn, err := client.DialContext(ctx, network, addr)
	if err != nil {
		return nil, fmt.Errorf("unable to dial address: %w", err)
	}
	return &wrappedSshConn{Conn: conn}, nil
}

func (s *SSHDialer) Dial(network, addr string) (net.Conn, error) {
	return s.DialContext(context.Background(), network, addr)
}

func (s *SSHDialer) Close() error {
	s.clientmu.Lock()
	defer s.clientmu.Unlock()
	client := s.client
	s.client = nil

	if client != nil {
		return client.Close()
	}
	return nil
}

const (
	// this is a well known name for the keepalive request that is respected by most ssh servers
	keepaliveName = "keepalive@openssh.com"
)

func (s *SSHDialer) getClient(ctx context.Context) (*ssh.Client, error) {
	s.clientmu.Lock()
	defer s.clientmu.Unlock()

	if s.client != nil {
		wantReply := true
		_, _, err := s.client.SendRequest(keepaliveName, wantReply, nil)
		if err == nil {
			return s.client, nil
		}
		s.logger.Warn(fmt.Sprintf("SSH client was dead, closing and attempting to re-open connection: %s", err.Error()))
		s.client.Close()
		s.client = nil
	}

	operation := func() (*ssh.Client, error) {
		return ssh.Dial("tcp", s.addr, s.ccfg)
	}

	client, err := backoffutil.Retry(
		ctx,
		operation,
		func() []backoff.RetryOption { return s.dialCfg.getRetryOpts(s.logger) },
		isRetryableError,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to dial ssh server after multiple attempts: %w", err)
	}

	s.client = client
	s.startKeepAlive(client)
	return client, nil
}

// Could expand on this more if there are center errors we do not want to retry
func isRetryableError(err error) bool {
	return err != nil
}

func (s *SSHDialer) startKeepAlive(client *ssh.Client) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("recovered from panic in ssh keepalive goroutine",
					"error", r,
					"stack", string(debug.Stack()),
				)
				// Clean up the client connection on panic
				s.clientmu.Lock()
				defer s.clientmu.Unlock()
				if s.client == client {
					s.logger.Info("abandoning keepalive for active ssh client due to panic")
					s.client = nil
				} else {
					s.logger.Info("closing old ssh client due to keepalive panic")
					err := client.Close()
					if err != nil {
						s.logger.Info("error closing old ssh client during keepalive panic", "error", err)
					}
				}
			}
		}()
		s.logger.Info("keepalive started for ssh client")
		t := time.NewTicker(s.dialCfg.KeepAliveInterval)
		defer t.Stop()

		for range t.C {
			s.clientmu.Lock()
			if s.client != client || s.client == nil {
				s.clientmu.Unlock()
				return
			}

			// Create a timeout context for the keepalive request
			ctx, cancel := context.WithTimeout(context.Background(), s.dialCfg.KeepAliveTimeout)
			done := make(chan error, 1)

			go func() {
				wantReply := true
				_, _, err := client.SendRequest(keepaliveName, wantReply, nil)
				done <- err
			}()

			// Wait for either timeout or response
			select {
			case err := <-done:
				if err != nil {
					s.logger.Error("keepalive failed", "error", err)
					s.client = nil
					client.Close()
				}
			case <-ctx.Done():
				s.logger.Error("keepalive timed out")
				s.client = nil
				client.Close()
			}

			cancel()
			s.clientmu.Unlock()
		}
	}()
}

type wrappedSshConn struct {
	net.Conn
}

func (w *wrappedSshConn) SetDeadline(deadline time.Time) error {
	if err := w.SetReadDeadline(deadline); err != nil {
		return err
	}
	return w.SetWriteDeadline(deadline)
}

// SSH net.Conn does not implement this, so we're overriding it to not return an error
func (w *wrappedSshConn) SetReadDeadline(deadline time.Time) error {
	return nil
}

// SSH net.Conn does not implement this, so we're overriding it to not return an error
func (w *wrappedSshConn) SetWriteDeadline(deadline time.Time) error {
	return nil
}
