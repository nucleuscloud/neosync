package sshtunnel

import "fmt"

type Endpoint struct {
	host string
	port int
}

func NewEndpoint(host string, port int) *Endpoint {
	return &Endpoint{host: host, port: port}
}

func (e *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", e.host, e.port)
}

func (e *Endpoint) GetValues() (host string, port int) {
	return e.host, e.port
}

func (e *Endpoint) SetPort(port int) {
	e.port = port
}
