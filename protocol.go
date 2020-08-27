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

type handlerOpts struct {
	addr    *net.UDPAddr
	buf     [handshakeSize]byte
	storage chan<- *udpClient
	get     chan<- *streamToken
	del     chan<- string
	done    chan struct{}
}
