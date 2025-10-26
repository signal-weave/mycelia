package comm

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"mycelia/errgo"
	"mycelia/globals"
)

// -----------------------------------------------------------------------------
// Comm is a boilerplate and helper library for communication over the wire.

// Typically these functions do not make Mycelia errors because there is better
// context for the error string in the caller.
// -----------------------------------------------------------------------------

const lenU32 = 4

// ReadFrameU32 reads the frame's byte stream until the message header's worth
// of bytes have been consumed, then return a buffer of those bytes or error.
func ReadFrameU32(conn net.Conn) ([]byte, error) {
	var hdr [lenU32]byte
	if _, err := io.ReadFull(conn, hdr[:]); err != nil {
		return nil, err
	}
	n := binary.BigEndian.Uint32(hdr[:])

	if n == 0 {
		return nil, nil
	}
	if n > globals.BytesInMegabyte {
		// sanity limit of 1MB
		return nil, errgo.NewError(
			fmt.Sprintf("Frame too large: %d bytes", n), globals.VerbWrn,
		)
	}

	p := BufPool.Get().(*[]byte)
	buf := *p
	if cap(buf) < int(n) {
		buf = make([]byte, n) // fallback if frame > pooled buffer
	} else {
		buf = buf[:n]
	}

	if _, err := io.ReadFull(conn, buf); err != nil {
		BufPool.Put(p) // return to pool before leaving
		return nil, err
	}

	out := make([]byte, n)
	copy(out, buf)

	BufPool.Put(p) // return buffer for reuse
	return out, nil
}
