package sandy

import (
	"io"
	"log"
	"net"
	"time"
)

type udpClient struct {
	addr           *net.UDPAddr
	file           io.WriteCloser
	ready          chan struct{}
	done           chan struct{}
	buf            [handshakeSize]byte
	stations       chan<- *udpClient
	del            chan<- string
	get            chan<- *streamGetter
	siblingWriters chan io.WriteCloser
}

type streamGetter struct {
	key    string
	result chan io.WriteCloser
}

// newClienthandle all Clients
// so call it ONCE is enough
func newClient() *udpClient {
	pools, getter, del := globalStreams()
	cli := &udpClient{
		buf:      [handshakeSize]byte{},
		stations: pools,
		get:      getter,
		del:      del,
	}
	return cli
}

func (u *udpClient) init(conn *net.UDPConn, getStorage FaceMouther) error {
	var err error
	if u.addr == nil {
		conn.SetReadDeadline(time.Now().Add(5 * 12 * readUDPTimeout))
		_, u.addr, err = conn.ReadFromUDP(u.buf[:])
		if err != nil {
			log.Println("readfromUDP failed")
			return err
		}
	}
	id := string(u.buf[:]) + "." + u.addr.String() + ".debug"
	storage, err := getStorage(id)
	if err != nil {
		log.Println("create file failed")
		return err
	}
	u.file = storage
	return nil
}

func (u *udpClient) spawn(addr *net.UDPAddr, bs []byte) *udpClient {
	var backup [handshakeSize]byte
	copy(backup[:], bs)
	newCli := &udpClient{
		addr:     addr,
		buf:      backup,
		stations: u.stations,
		del:      u.del,
		get:      u.get,
	}
	return newCli
}

func (u *udpClient) putWriter() {
	u.ready = make(chan struct{})
	u.done = make(chan struct{})
	u.siblingWriters = make(chan io.WriteCloser)
	u.stations <- u
	<-u.ready
}

func (u *udpClient) getWriter(id string) io.WriteCloser {
	if id == u.addr.String() {
		return u.file
	}
	req := &streamGetter{
		key:    id,
		result: u.siblingWriters,
	}
	u.get <- req
	return <-u.siblingWriters
}

func (u *udpClient) close() {
	close(u.siblingWriters)
}
