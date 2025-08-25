package server

import (
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"
)

// writeUvarint is a tiny helper for tests.
func writeUvarint(w io.Writer, n uint64) {
	var buf [binary.MaxVarintLen64]byte
	k := binary.PutUvarint(buf[:], n)
	_, _ = w.Write(buf[:k])
}

func runHandlerAsync(t *testing.T, srv *Server, conn net.Conn) chan struct{} {
	t.Helper()
	done := make(chan struct{})
	go func() {
		srv.handleConnection(conn)
		close(done)
	}()
	return done
}

func waitOrTimeout(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
		// ok
	case <-time.After(1 * time.Second):
		t.Fatal("handler did not return within timeout")
	}
}

// Test: empty frame (msgLen=0) should be skipped, then EOF should exit cleanly.
func TestHandleConnection_EmptyFrameThenEOF(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	defer clientSide.Close() // serverSide closed by handler

	srv := &Server{
		Broker:  nil, // safe: we won't send a non-empty message
		address: "pipe",
		port:    0,
	}

	done := runHandlerAsync(t, srv, serverSide)

	// Send msgLen = 0
	writeUvarint(clientSide, 0)
	// Close to cause EOF on the next read
	_ = clientSide.Close()

	waitOrTimeout(t, done)
}

// Test: corrupt length varint (overflow) should log a warning and return.
func TestHandleConnection_BadLengthVarint(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	defer clientSide.Close()

	srv := &Server{
		Broker:  nil, // no payload should be processed
		address: "pipe",
		port:    0,
	}

	done := runHandlerAsync(t, srv, serverSide)

	// Send 10 bytes with MSB set (0x80) to trigger binary.ErrOverflow.
	// (MaxVarintLen64 is 10; continuation bit set on all 10 -> overflow)
	bad := make([]byte, binary.MaxVarintLen64)
	for i := range bad {
		bad[i] = 0x80
	}
	_, _ = clientSide.Write(bad)
	_ = clientSide.Close()

	waitOrTimeout(t, done)
}

// Test: truncated body should warn and return (length says 5, only 3 bytes arrive).
func TestHandleConnection_TruncatedBody(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	defer clientSide.Close()

	srv := &Server{
		Broker:  nil, // we won't reach HandleBytes due to truncation
		address: "pipe",
		port:    0,
	}

	done := runHandlerAsync(t, srv, serverSide)

	// Declare msgLen = 5 but only send 3 bytes.
	writeUvarint(clientSide, 5)
	_, _ = clientSide.Write([]byte("abc"))
	_ = clientSide.Close() // triggers io.ReadFull error on the server

	waitOrTimeout(t, done)
}

// Optional: immediate EOF (client closes without sending any bytes) should exit cleanly.
func TestHandleConnection_ImmediateEOF(t *testing.T) {
	serverSide, clientSide := net.Pipe()

	srv := &Server{
		Broker:  nil,
		address: "pipe",
		port:    0,
	}

	done := runHandlerAsync(t, srv, serverSide)

	_ = clientSide.Close()

	waitOrTimeout(t, done)
}
