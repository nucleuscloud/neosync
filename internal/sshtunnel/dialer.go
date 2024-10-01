package sshtunnel

import (
	"context"
	"fmt"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
)

type Dialer interface {
	DialContext(ctx context.Context, network, addr string) (net.Conn, error)
	Dial(network, addr string) (net.Conn, error)
}

type SSHClientDialer interface {
	Dial(network, addr string, cfg *ssh.ClientConfig) (*ssh.Client, error)
}

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
	return client.DialContext(ctx, network, addr)
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
	s.clientmu.Unlock()
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
	return client, nil
}
