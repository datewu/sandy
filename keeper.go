package sandy

import (
	"io"
)

type bookKeeper struct {
	pools        map[string]*udpClient
	addStream    chan *udpClient
	getStream    chan string
	writerStream chan io.WriteCloser
	delStream    chan string
}

func newBookKeeper() *bookKeeper {
	b := &bookKeeper{
		pools:        make(map[string]*udpClient),
		addStream:    make(chan *udpClient),
		getStream:    make(chan string),
		writerStream: make(chan io.WriteCloser),
		delStream:    make(chan string),
	}
	go func() {
		for {
			select {
			case f, ok := <-b.addStream:
				if !ok {
					return
				}
				b.pools[f.addr.String()] = f
				close(f.ready)
			case id, ok := <-b.getStream:
				if !ok {
					return
				}
				if r, ok := b.pools[id]; ok {
					b.writerStream <- r.file
				} else {
					b.writerStream <- nil
				}
			case id, ok := <-b.delStream:
				if !ok {
					return
				}
				if r, ok := b.pools[id]; ok {
					close(r.done)
					delete(b.pools, id)
				}
			}
		}
	}()
	return b
}

func (b *bookKeeper) add(cli *udpClient) {
	b.addStream <- cli
}

func (b *bookKeeper) getWriter(id string) io.WriteCloser {
	b.getStream <- id
	return <-b.writerStream
}

func (b *bookKeeper) remove(id string) {
	b.delStream <- id
}

func (b *bookKeeper) destroy() {
	close(b.addStream)
	close(b.getStream)
	close(b.writerStream)
	close(b.delStream)
}
