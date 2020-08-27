package sandy

import (
	"encoding/binary"
	"time"
)

const (
	magicEOF    = `"@4'`
	firstLen    = 8
	bufSize     = 1024
	readTimeout = 5 * time.Second
)

func filesAccessor() (chan<- *request, chan<- *gettter, chan<- string) {
	reqs := make(chan *request)
	filesPool := make(map[string]*request)
	getStream := make(chan *gettter)
	delStream := make(chan string)
	go func() {
		for {
			select {
			case f, ok := <-reqs:
				if !ok {
					return
				}
				filesPool[f.id] = f
				//log.Info().Int("size", len(filesPool)).Msg("current pool size")
				//go func() {
				f.ready <- struct{}{}
				//	}()
			case g, ok := <-getStream:
				if !ok {
					return
				}
				if r, ok := filesPool[g.id]; ok {
					g.result <- r.file
				}
				g.result <- nil
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
