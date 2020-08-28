package sandy

import (
	"io"
	"log"
	"net"
	"time"
)

type udpClient struct {
	addr   *net.UDPAddr
	file   io.WriteCloser
	ready  chan struct{}
	done   chan struct{}
	buf    [handshakeSize]byte
	keeper *bookKeeper
}

// newClienthandle all Clients
// root client can spawn a lot
// so call it ONCE is enough
func newClient(k *bookKeeper) *udpClient {
	cli := &udpClient{
		buf:    [handshakeSize]byte{},
		keeper: k,
	}
	return cli
}

func (u *udpClient) initWriter(conn *net.UDPConn, getStorage FaceMouther) error {
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
		addr:   addr,
		buf:    backup,
		keeper: u.keeper,
	}
	return newCli
}

func (u *udpClient) registry() {
	u.ready = make(chan struct{})
	u.done = make(chan struct{})
	u.keeper.add(u)
	<-u.ready
}

func (u *udpClient) getWriter(id string) io.WriteCloser {
	if id == u.addr.String() {
		return u.file
	}
	return u.keeper.getWriter(id)
}
