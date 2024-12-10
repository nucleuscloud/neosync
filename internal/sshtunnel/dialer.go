package sshtunnel

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

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
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration // Max allowed backoff time
	BackoffFactor  float64       // backoff multiplier

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
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
		BackoffFactor:  2,

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
		s.logger.Info(fmt.Sprintf("SSH client was dead, closing and attempting to re-created: %s", err.Error()))
		s.client.Close()
		s.client = nil
	}

	var client *ssh.Client
	var err error
	backoff := s.dialCfg.InitialBackoff

	for i := 0; i < s.dialCfg.MaxRetries; i++ {
		client, err = ssh.Dial("tcp", s.addr, s.ccfg)
		if err == nil {
			s.startKeepAlive(client)
			break
		}
		s.logger.Error(fmt.Sprintf("failed to dial SSH Server on attempt %d/%d: %s", i, s.dialCfg.MaxRetries, err.Error()))
		if i < s.dialCfg.MaxRetries-1 {
			s.logger.Debug(fmt.Sprintf("waiting %.1f seconds until attempting to re-connect to SSH Server", backoff.Seconds()))
			err = sleepContext(ctx, backoff)
			if err != nil {
				break
			}
			nextBackoff := time.Duration(float64(backoff) * s.dialCfg.BackoffFactor)
			if nextBackoff > s.dialCfg.MaxBackoff {
				nextBackoff = s.dialCfg.MaxBackoff
			}
			backoff = nextBackoff
		}
	}

	if err != nil {
		return nil, fmt.Errorf("unable to dial ssh server after %d attempts: %w", s.dialCfg.MaxRetries, err)
	}
	s.client = client
	return client, nil
}

func (s *SSHDialer) startKeepAlive(client *ssh.Client) {
	go func() {
		s.logger.Info("keepalive started for ssh client")
		t := time.NewTicker(s.dialCfg.KeepAliveInterval)
		defer t.Stop()

		for range t.C {
			s.clientmu.Lock()
			if s.client != client {
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

			if s.client == nil {
				return
			}
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

func sleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}
