package sandy

import (
	"time"
)

const (
	hangUPEOF        = `¡"¶⬆️桐👋*™🈚️'¬¢`
	handshakeTailLen = 1
	magicTail        = 255
	maxHandshakeSize = (255 + 1) * handshakeTailLen
	bufSize          = 1024
	readUDPTimeout   = 8 * time.Second
)

func encodeHandshake(bs []byte) []byte {
	if len(bs) < 1 {
		return nil
	}
	tail := byte(len(bs))
	return append(bs, magicTail, tail)
}

func decodeHandshake(bs []byte) []byte {
	if len(bs) < 3 {
		return nil
	}
	size := int(bs[len(bs)-1])
	return bs[0:size]
}

func isHandshake(bs []byte) bool {
	if len(bs) < 3 {
		return false
	}
	total := len(bs)
	size := int(bs[total-1])
	if bs[total-2] != magicTail {
		return false
	}
	return size+2 == total
}
