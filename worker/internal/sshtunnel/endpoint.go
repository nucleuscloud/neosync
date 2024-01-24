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
func NewEndpoint(host string, port int, user *string) (*Endpoint, error) {
	endpoint := &Endpoint{
		Host: host,
		Port: port,
	}
	if user != nil {
		endpoint.User = *user
	}
	return endpoint, nil

	// if parts := strings.Split(endpoint.Host, "@"); len(parts) > 1 {
	// 	endpoint.User = parts[0]
	// 	endpoint.Host = parts[1]
	// }

	// host, port, err := net.SplitHostPort(endpoint.Host)
	// if err != nil {
	// 	// if error results from missing port in address, we ignore the error
	// 	// since either we'll use a random port assigned by the OS or set a
	// 	// suitable default directly, e.g. port 22 for SSH. Also worth noting,
	// 	// the host is set to the rest of the string since no port is provided
	// 	if !strings.Contains(err.Error(), "missing port in address") {
	// 		return nil, err
	// 	}
	// } else {
	// 	endpoint.Host = host
	// 	endpoint.Port, _ = strconv.Atoi(port)
	// }

	// return endpoint, nil
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}
