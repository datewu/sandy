package sandy

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// Peanut for sandy
type Peanut struct {
	Protein  io.ReadCloser
	Name     string
	Size     int64
	Feedback chan string
}

// Upload feed sandy
func Upload(server string, p *Peanut) {
	defer func() {
		close(p.Feedback)
		p.Protein.Close()
	}()

	udpAddr, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		log.Println("resolve server address failed", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Println("dialUDP failed", err)
		return
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Println("client failed to close conn", err)
		}
	}()
	hs := encodeHandshake([]byte(p.Name))
	n, err := conn.Write(hs)
	if err != nil {
		log.Println("conn write failed")
		return
	}

	rBuf := make([]byte, n)
	conn.SetReadDeadline(time.Now().Add(2 * readUDPTimeout))
	m, _, err := conn.ReadFromUDP(rBuf[:])
	if err != nil {
		log.Println("readFromUDP failed", err)
		return
	}
	if m != n || string(rBuf) != string(hs) {
		log.Println("do not get 'handshake' back, handshake failed")
		return
	}

	var progress [2]byte
	var accumulated int64
	size := p.Size
	var fBuf [bufSize]byte
	for {
		n, err := p.Protein.Read(fBuf[:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Println("read file err", err)
			return
		}
		_, err = conn.Write(fBuf[:n])
		if err != nil {
			log.Println("conn write failed", err)
			return
		}
		conn.SetReadDeadline(time.Now().Add(readUDPTimeout))
		_, _, err = conn.ReadFromUDP(progress[:])
		if err != nil {
			log.Println("readFromUDP failed, cannot show progress", err)
			return
		}
		m := bytes2Int(progress[:])
		accumulated += int64(m)
		p.Feedback <- fmt.Sprintf("%s: progress: %v/%v kb (%.2f%%)\r",
			p.Name,
			accumulated/1024,
			size/1024,
			100*float64(accumulated)/float64(size))
	}
	p.Feedback <- fmt.Sprintf("%s: progress: %v/%v kb (%.2f%%)\r\n",
		p.Name,
		accumulated/1024,
		size/1024,
		100*float64(accumulated)/float64(size))
	_, err = conn.Write([]byte(hangUPEOF))
	if err != nil {
		log.Println("conn read failed", err)
		return
	}
}
