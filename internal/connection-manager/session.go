package connectionmanager

import "fmt"

type Session struct {
	name string
	cfg  *sessionConfig
}
type SessionInterface interface {
	Name() string
	Group() string
	String() string
}
type SessionGroupInterface interface {
	Group() string
}

func (s *Session) Group() string {
	return s.cfg.group
}
func (s *Session) Name() string {
	return s.name
}

func (s *Session) String() string {
	if s.cfg.group != "" {
		return fmt.Sprintf("%s:%s", s.cfg.group, s.name)
	}
	return s.name
}

type sessionConfig struct {
	group string
}
type SessionOption func(*sessionConfig)

func WithSessionGroup(group string) SessionOption {
	return func(sc *sessionConfig) {
		sc.group = group
	}
}

func NewSession(name string, opts ...SessionOption) *Session {
	cfg := &sessionConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return &Session{name: name, cfg: cfg}
}
