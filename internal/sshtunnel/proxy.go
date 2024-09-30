package sshtunnel

type NetProxy struct {
	source      *Endpoint
	destination *Endpoint
	dailer      Dialer
}

// Create a new proxy that when started will bind to the source endpoint
func NewProxy(
	source *Endpoint,
	destination *Endpoint,
	dialer Dialer,
) *NetProxy {
	return &NetProxy{
		source:      source,
		destination: destination,
		dailer:      dialer,
	}
}

// The proxy binds to localhost on a random port
func NewHostProxy(
	destination *Endpoint,
	dialer Dialer,
) *NetProxy {
	return &NetProxy{
		source:      &Endpoint{host: "localhost", port: 0},
		destination: destination,
		dailer:      dialer,
	}
}

// I think this thing was supposed to act as a traditional proxy.
// It takes in a source, destination, and a tunnel dialer.
// When the proxy is started, it will start a TCP server
// and forward incoming traffic to the destination
// I think the source is useful if you want to bind to a specific local port
// Port 0 is the special case that allows you to bind to any port

func (n *NetProxy) GetSourceValues() (host string, port int) {
	return n.source.GetValues()
}
