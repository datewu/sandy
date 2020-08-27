package sandy

import (
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

var (
	globalForWG sync.WaitGroup
)

// FaceMouther is a face or a mouth, i cannnot tell
type FaceMouther func(id string) (io.WriteCloser, error)

// Server ready for infinity Peanuts
func Server(addr string, getStorage FaceMouther) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Println("resolve add failed")
		return
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Println("listen failed")
		return
	}
	requests, getter, del := globalStreams()
	defer func() {
		close(requests)
		close(getter)
		close(del)
	}()
	opts := &handlerOpts{
		buf:     [handshakeSize]byte{},
		storage: requests,
		get:     getter,
		del:     del,
	}
	for {
		opts.addr = nil
		opts.buf = [handshakeSize]byte{}
		go handleUPload(conn, opts, getStorage)
		globalForWG.Add(1)
		globalForWG.Wait()
	}
}

func handleUPload(conn *net.UDPConn, opts *handlerOpts, getStorage FaceMouther) {
	defer func() {
		globalForWG.Done()
	}()
	var err error
	if opts.addr == nil {
		conn.SetReadDeadline(time.Now().Add(5 * 12 * readUDPTimeout))
		_, opts.addr, err = conn.ReadFromUDP(opts.buf[:])
		if err != nil {
			log.Println("readfromUDP failed")
			return
		}
	}
	id := string(opts.buf[:]) + "." + opts.addr.String() + ".debug"
	storage, err := getStorage(id)
	if err != nil {
		log.Println("create file failed")
		return
	}
	defer storage.Close()
	log.Println("about to receive file")
	ready := make(chan struct{})
	defer close(ready) // TODO should close on send side
	done := make(chan struct{})
	opts.storage <- &udpClient{
		id:    opts.addr.String(),
		file:  storage,
		ready: ready,
		done:  done,
	}
	<-ready
	_, err = conn.WriteToUDP(opts.buf[:], opts.addr)
	if err != nil {
		log.Println("server writetoUDP failed")
		return
	}
	var fBuf [bufSize]byte
	for {
		select {
		case <-done:
			return
		default:
		}
		conn.SetReadDeadline(time.Now().Add(readUDPTimeout))
		n, b, err := conn.ReadFromUDP(fBuf[:])
		if err != nil {
			log.Println("server ReadFromUDP error")
			if errors.Is(err, io.EOF) {
				break
			}
			return
		}
		newfile := &streamToken{ // must be a new getter every time
			id:     b.String(),
			result: make(chan io.WriteCloser, 1),
		}
		defer close(newfile.result) // TODO should close on send side
		opts.get <- newfile
		if dest := <-newfile.result; dest != nil {
			if n == len(hangUPEOF) && string(fBuf[:len(hangUPEOF)]) == hangUPEOF {
				log.Println("got magicEOF finishd handle")
				opts.del <- b.String()
				if b.String() == opts.addr.String() {
					return
				}
				continue
			}
			_, err = dest.Write(fBuf[:n])
			if err != nil {
				log.Println("write to file failed")
				return
			}
			conn.WriteToUDP(int2Bytes(n), b)
			continue
		}
		log.Println("new connection")
		if n != handshakeSize {
			log.Println("new connection invalid, should send firstLen bytes")
			return
		}
		var backup [handshakeSize]byte
		copy(backup[:], fBuf[:handshakeSize])
		newOpts := &handlerOpts{
			addr:    b,
			buf:     backup,
			storage: opts.storage,
			get:     opts.get,
			del:     opts.del,
		}
		globalForWG.Add(1)
		go handleUPload(conn, newOpts, getStorage)
	}
}
