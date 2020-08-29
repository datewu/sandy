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
type FaceMouther func(name string, connID string) (io.WriteCloser, error)

// Serve ready for infinity Peanuts
func Serve(addr string, getStorage FaceMouther) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Println("resolve add failed", err)
		return
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Println("listen failed")
		return
	}
	// only one bookkeeper is more than enouth
	// though there can be more bookkerrpers
	k := newBookKeeper()
	defer k.destroy()
	rootCli := newClient(k)
	for {
		rootCli.addr = nil
		rootCli.buf = make([]byte, maxHandshakeSize)
		go handleUPload(conn, rootCli, getStorage)
		globalForWG.Add(1)
		globalForWG.Wait()
	}
}

func handleUPload(conn *net.UDPConn, cli *udpClient, getStorage FaceMouther) {
	defer globalForWG.Done()
	err := cli.initHandshake(conn, getStorage)
	if err != nil {
		return
	}
	defer cli.file.Close()
	var fBuf [bufSize]byte
	for {
		select {
		case <-cli.done:
			return
		default:
		}
		conn.SetReadDeadline(time.Now().Add(readUDPTimeout))
		n, b, err := conn.ReadFromUDP(fBuf[:])
		if err != nil {
			log.Println("server ReadFromUDP error", err)
			if errors.Is(err, io.EOF) {
				break
			}
			return
		}
		if n == len(hangUPEOF) && string(fBuf[:len(hangUPEOF)]) == hangUPEOF {
			log.Println("got magicEOF finishd handle")
			cli.keeper.remove(b.String())
			if b.String() == cli.addr.String() {
				return
			}
			continue
		}

		if dest := cli.getWriter(b.String()); dest != nil {
			_, err = dest.Write(fBuf[:n])
			if err != nil {
				log.Println("write to file failed", err)
				return
			}
			conn.WriteToUDP(int2Bytes(n), b)
			continue
		}
		if !isHandshake(fBuf[:n]) {
			if n > 80 {
				n = 80
			}
			log.Println("new connection invalid, isHandshake false", string(fBuf[:n]))
			return
		}
		newCli := cli.spawn(b, fBuf[:n])
		globalForWG.Add(1)
		go handleUPload(conn, newCli, getStorage)
	}
}
