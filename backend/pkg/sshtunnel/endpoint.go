package sshtunnel

import (
	"fmt"
)

type Endpoint struct {
	Host string
	Port int
	User string
}

// NewEndpoint creates an Endpoint from a string that contains a user, host and
// port. Both User and Port are optional (depending on context). The host can
// be a domain name, IPv4 address or IPv6 address. If it's an IPv6, it must be
// enclosed in square brackets
func NewEndpointWithUser(host string, port int, user string) *Endpoint {
	return &Endpoint{
		Host: host,
		Port: port,
		User: user,
	}
}

func NewEndpoint(host string, port int) *Endpoint {
	return &Endpoint{Host: host, Port: port}
}

// Returns the stringified endpoint sans user
func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}
