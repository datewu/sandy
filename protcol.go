package sandy

import (
	"io"
	"net"
)

type request struct {
	id    string
	file  io.WriteCloser
	ready chan struct{}
	done  chan struct{}
}

type gettter struct {
	id     string
	result chan io.WriteCloser
}

type handlerOpts struct {
	addr    *net.UDPAddr
	buf     [firstLen]byte
	storage chan<- *request
	get     chan<- *gettter
	del     chan<- string
	done    chan struct{}
}
