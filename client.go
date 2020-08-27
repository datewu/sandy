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

// Upload ...
func Upload(server string, p *Peanut) {
	defer func() {
		close(p.Feedback)
		p.Protein.Close()
	}()

	udpAddr, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		log.Println("resolve server address failed")
		return
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Println("dialUDP failed")
		return
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Println("client failed to close conn")
		}
	}()
	sf := p.Name + "padding"
	_, err = conn.Write([]byte(sf[:firstLen]))
	if err != nil {
		log.Println("conn write failed")
		return
	}

	var rBuf [firstLen]byte // must receive firstLen bytes before send file
	conn.SetReadDeadline(time.Now().Add(2 * readTimeout))
	_, _, err = conn.ReadFromUDP(rBuf[:])
	if err != nil {
		log.Println("readFromUDP failed")
		return
	}
	if string(rBuf[:]) != sf[:firstLen] {
		log.Println("do not get 'fileName' back")
		return
	}

	var progress [2]byte
	var accumulated int64
	size := p.Size
	var fBuf [bufSize]byte
	log.Println("going to send file for loop")
	for {
		n, err := p.Protein.Read(fBuf[:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				//	log.Println("going to close conn")
				break
			}
			log.Println("read file err", err)
			return
		}
		_, err = conn.Write(fBuf[:n])
		if err != nil {
			log.Println("conn write failed")
			return
		}
		conn.SetReadDeadline(time.Now().Add(readTimeout))
		_, _, err = conn.ReadFromUDP(progress[:])
		if err != nil {
			log.Println("readFromUDP failed, cannot show progress")
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
	log.Println("file send finished")
	_, err = conn.Write([]byte(magicEOF))
	if err != nil {
		log.Println("conn read failed")
		return
	}
}
