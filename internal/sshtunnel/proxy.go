package sshtunnel

type NetProxy struct {
	source      *Endpoint
	destination *Endpoint
	dailer      Dialer
}

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

func (n *NetProxy) GetSourceValues() (host string, port int) {
	return n.source.GetValues()
}
