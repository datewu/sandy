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
	streams := newClientsStream()
	defer func() {
		close(streams.stations)
		close(streams.get)
		close(streams.del)
	}()

	for {
		streams.addr = nil
		streams.buf = [handshakeSize]byte{}
		go handleUPload(conn, streams, getStorage)
		globalForWG.Add(1)
		globalForWG.Wait()
	}
}

func handleUPload(conn *net.UDPConn, stream *clientsStream, getStorage FaceMouther) {
	defer func() {
		globalForWG.Done()
	}()
	var err error
	if stream.addr == nil {
		conn.SetReadDeadline(time.Now().Add(5 * 12 * readUDPTimeout))
		_, stream.addr, err = conn.ReadFromUDP(stream.buf[:])
		if err != nil {
			log.Println("readfromUDP failed")
			return
		}
	}
	id := string(stream.buf[:]) + "." + stream.addr.String() + ".debug"
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
	stream.stations <- &udpClient{
		id:    stream.addr.String(),
		file:  storage,
		ready: ready,
		done:  done,
	}
	<-ready
	_, err = conn.WriteToUDP(stream.buf[:], stream.addr)
	if err != nil {
		log.Println("server writetoUDP failed")
		return
	}
	var fBuf [bufSize]byte
	token := &streamToken{
		result: make(chan io.WriteCloser, 1),
	}
	defer close(token.result) // TODO should close on send side
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
		token.id = b.String()
		stream.get <- token
		if dest := <-token.result; dest != nil {
			if n == len(hangUPEOF) && string(fBuf[:len(hangUPEOF)]) == hangUPEOF {
				log.Println("got magicEOF finishd handle")
				stream.del <- b.String()
				if b.String() == stream.addr.String() {
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
		newOpts := &clientsStream{
			addr:     b,
			buf:      backup,
			stations: stream.stations,
			get:      stream.get,
			del:      stream.del,
		}
		globalForWG.Add(1)
		go handleUPload(conn, newOpts, getStorage)
	}
}
