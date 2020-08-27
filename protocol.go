package sandy

import (
	"io"
	"net"
)

type udpClient struct {
	id    string
	file  io.WriteCloser
	ready chan struct{}
	done  chan struct{}
}

type streamToken struct {
	id     string
	result chan io.WriteCloser
}

type clientsStream struct {
	addr     *net.UDPAddr
	buf      [handshakeSize]byte
	stations chan<- *udpClient
	get      chan<- *streamToken
	del      chan<- string
}

// newClientsSteam handle all Clients
// so call it ONCE is enough
func newClientsStream() *clientsStream {
	pools, getter, del := globalStreams()
	ss := &clientsStream{
		buf:      [handshakeSize]byte{},
		stations: pools,
		get:      getter,
		del:      del,
	}
	return ss
}
