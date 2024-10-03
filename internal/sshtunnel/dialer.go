package sshtunnel

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type Dialer interface {
	DialContext(ctx context.Context, network, addr string) (net.Conn, error)
	Dial(network, addr string) (net.Conn, error)
}

var _ Dialer = (*SSHDialer)(nil)

type SSHDialer struct {
	addr string
	cfg  *ssh.ClientConfig

	client   *ssh.Client
	clientmu *sync.RWMutex
}

func NewLazySSHDialer(addr string, cfg *ssh.ClientConfig) *SSHDialer {
	return &SSHDialer{addr: addr, cfg: cfg, clientmu: &sync.RWMutex{}}
}

func NewSSHDialer(client *ssh.Client) *SSHDialer {
	return &SSHDialer{client: client, clientmu: &sync.RWMutex{}}
}

func (s *SSHDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}
	conn, err := client.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}
	return &wrappedSshConn{Conn: conn}, nil
}

func (s *SSHDialer) Dial(network, addr string) (net.Conn, error) {
	return s.DialContext(context.Background(), network, addr)
}

func (s *SSHDialer) Close() error {
	s.clientmu.Lock()
	defer s.clientmu.Unlock()
	if s.client == nil {
		return nil
	}
	client := s.client
	s.client = nil
	return client.Close()
}

func (s *SSHDialer) getClient() (*ssh.Client, error) {
	s.clientmu.RLock()
	client := s.client
	s.clientmu.RUnlock()
	if client != nil {
		return client, nil
	}
	s.clientmu.Lock()
	defer s.clientmu.Unlock()
	if s.client != nil {
		return s.client, nil
	}
	// todo: implement retries
	client, err := ssh.Dial("tcp", s.addr, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to dial ssh server: %w", err)
	}
	s.client = client
	return client, nil
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
