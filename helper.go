package sandy

import (
	"encoding/binary"
	"time"
)

const (
	hangUPEOF        = `E"@⬆️桐👋*™🈚️'F`
	handshakeTailLen = 1
	magicTail        = 255
	maxHandshakeSize = (255 + 1) * handshakeTailLen
	bufSize          = 1024
	readUDPTimeout   = 8 * time.Second
)

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
