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
func Upload(server string, p *Peanut) error {
	defer func() {
		close(p.Feedback)
		p.Protein.Close()
	}()

	udpAddr, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		log.Println("resolve server address failed", err)
		return err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Println("dialUDP failed", err)
		return err
	}
	defer conn.Close()
	hs := encodeHandshake([]byte(p.Name))
	n, err := conn.Write(hs)
	if err != nil {
		log.Println("conn write failed")
		return err
	}

	rBuf := make([]byte, n)
	conn.SetReadDeadline(time.Now().Add(2 * readUDPTimeout))
	m, _, err := conn.ReadFromUDP(rBuf[:])
	if err != nil {
		log.Println("readFromUDP failed", err)
		return err
	}
	if m != n || string(rBuf) != string(hs) {
		log.Println("do not get 'handshake' back, handshake failed")
		return err
	}

	var accumulated int64
	size := p.Size
	var fBuf [bufSize]byte
	var dummy [1]byte
	p.Feedback <- fmt.Sprintf("going to upload: %q\r", p.Name)
	t := time.NewTicker(3 * time.Second)
	defer t.Stop()
	for {
		n, err := p.Protein.Read(fBuf[:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Println("read file err", err)
			return err
		}
		m, err := conn.Write(fBuf[:n])
		if err != nil {
			log.Println("conn write failed", err)
			return err
		}
		accumulated += int64(m)
		select {
		case <-t.C:
			p.Feedback <- fmt.Sprintf("%s: progress: %v/%v kb (%.2f%%)\r",
				p.Name,
				accumulated/1024,
				size/1024,
				100*float64(accumulated)/float64(size))
		default:
			conn.SetReadDeadline(time.Now().Add(readUDPTimeout))
			conn.ReadFrom(dummy[:])
		}
	}
	p.Feedback <- fmt.Sprintf("%s: progress: %v/%v kb (%.2f%%)\r\n",
		p.Name,
		accumulated/1024,
		size/1024,
		100*float64(accumulated)/float64(size))
	_, err = conn.Write([]byte(hangUPEOF))
	return err
}
