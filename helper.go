package sandy

import (
	"encoding/binary"
	"strings"
	"time"
)

const (
	hangUPEOF      = `E"@i桐O*4🈚️'F`
	handshakeSize  = 16
	bufSize        = 1024
	readUDPTimeout = 8 * time.Second
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

func pad2Size(s string, size int) string {
	sSize := len(s)
	if sSize >= size {
		return s[:size]
	}
	text := ".pad"
	padding := strings.Repeat(text, (size-sSize)/len(text)+1)
	return (s + padding)[:size]
}
