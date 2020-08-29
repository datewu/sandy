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
	buf    []byte
	keeper *bookKeeper
}

// newClienthandle all Clients
// root client can spawn a lot
// so call it ONCE is enough
func newClient(k *bookKeeper) *udpClient {
	cli := &udpClient{
		buf:    make([]byte, maxHandshakeSize),
		keeper: k,
	}
	return cli
}

func (u *udpClient) initHandshake(conn *net.UDPConn, getStorage FaceMouther) error {
	var err error
	var n int
	if u.addr == nil {
		conn.SetReadDeadline(time.Now().Add(5 * 12 * readUDPTimeout))
		n, u.addr, err = conn.ReadFromUDP(u.buf[:])
		if err != nil {
			log.Println("readfromUDP failed", err)
			return err
		}
	}
	if n == 0 {
		n = len(u.buf)
	}
	id := string(decodeHandshake(u.buf[:n])) + "." + u.addr.String() + ".debug"
	storage, err := getStorage(id)
	if err != nil {
		log.Println("create file failed", err)
		return err
	}
	u.file = storage
	u.registry()

	_, err = conn.WriteToUDP(u.buf[:n], u.addr)
	if err != nil {
		log.Println("server writetoUDP failed", err)
	}
	return err
}

func (u *udpClient) spawn(addr *net.UDPAddr, bs []byte) *udpClient {
	backup := make([]byte, len(bs))
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
