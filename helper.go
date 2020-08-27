package sandy

import (
	"encoding/binary"
	"time"
)

const (
	hangUPEOF      = `"@4'`
	handshakeSize  = 8
	bufSize        = 1024
	readUDPTimeout = 8 * time.Second
)

func globalStreams() (chan<- *udpClient, chan<- *streamGetter, chan<- string) {
	reqs := make(chan *udpClient)
	filesPool := make(map[string]*udpClient)
	getStream := make(chan *streamGetter)
	delStream := make(chan string)
	go func() {
		for {
			select {
			case f, ok := <-reqs:
				if !ok {
					return
				}
				filesPool[f.addr.String()] = f
				close(f.ready)
			case s, ok := <-getStream:
				if !ok {
					return
				}
				if r, ok := filesPool[s.key]; ok {
					s.result <- r.file
				} else {
					s.result <- nil
				}
			case id, ok := <-delStream:
				if !ok {
					return
				}
				if r, ok := filesPool[id]; ok {
					close(r.done)
					delete(filesPool, id)
				}
			}
		}
	}()
	return reqs, getStream, delStream
}

func int2Bytes(num int) []byte {
	bs := make([]byte, 2)
	n := uint16(num)
	binary.LittleEndian.PutUint16(bs, n)
	return bs
}

func bytes2Int(bs []byte) int {
	n := binary.LittleEndian.Uint16(bs)
	return int(n)
}
