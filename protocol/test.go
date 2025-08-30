package protocol

import "encoding/binary"

// helper: u8-length-prefixed string
func encStrU8(s string) []byte {
	b := make([]byte, 1+len(s))
	b[0] = uint8(len(s))
	copy(b[1:], []byte(s))
	return b
}

// helper: u16-length-prefixed string (big-endian)
func encStrU16(s string) []byte {
	b := make([]byte, 2+len(s))
	binary.BigEndian.PutUint16(b[0:2], uint16(len(s)))
	copy(b[2:], []byte(s))
	return b
}
