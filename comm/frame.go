package comm

import (
	"encoding/binary"
	"io"
	"net"
)

// -----------------------------------------------------------------------------
// Comm is a boilerplate and helper library for communication over the wire.

// Typically these functions do not make Mycelia errors because there is better
// context for the error string in the caller.
// -----------------------------------------------------------------------------

const lenU32 = 4

// WriteFrame writes a length-prefixed payload back to the client.
// Frame contains a 4 byte header prefix dictating message length.
func WriteFrameU32(conn net.Conn, payload []byte) error {
	var hdr [lenU32]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(payload)))

	if _, err := conn.Write(hdr[:]); err != nil {
		return err
	}
	_, err := conn.Write(payload)

	return err
}

// Read the frame's byte stream until the message header's worth of bytes have
// been consumed, then return a buffer of those bytes or error.
func ReadFrameU32(conn net.Conn) ([]byte, error) {
	var hdr [lenU32]byte
	if _, err := io.ReadFull(conn, hdr[:]); err != nil {
		return nil, err
	}

	n := binary.BigEndian.Uint32(hdr[:])
	if n == 0 {
		return []byte{}, nil
	}

	buf := make([]byte, n)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, err
	}
	return buf, nil
}
