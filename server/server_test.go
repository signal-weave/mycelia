package server

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"mycelia/comm"
	"mycelia/test"
)

// Starts a one-shot TCP listener and wires the accepted conn into
// HandleConnection.
// Returns: client-side net.Conn, a done channel that closes when the server
// handler returns, and a cleanup func.
func startTestServer(t *testing.T, s *Server) (net.Conn, <-chan struct{}, func()) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	addr := ln.Addr().String()
	done := make(chan struct{})

	// Accept one connection and hand it to the handler
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			// Listener closed early or accept error; just exit
			return
		}
		s.HandleConnection(conn) // HandleConnection closes conn when done
	}()

	// Client connects to the ephemeral port
	c, err := net.Dial("tcp", addr)
	if err != nil {
		_ = ln.Close()
		t.Fatalf("dial failed: %v", err)
	}

	cleanup := func() {
		_ = c.Close()
		_ = ln.Close()
	}

	return c, done, cleanup
}

// Zero-length frame => framed "ERR: empty frame:" and loop continues.
func TestHandleConnection_ZeroLengthFrame_RepliesError(t *testing.T) {
	test.WithTimeout(t, 3*time.Second, func(t *testing.T) {
		s := &Server{}
		c, done, cleanup := startTestServer(t, s)
		defer cleanup()

		// Send zero-length frame
		if err := comm.WriteFrameU32(c, nil); err != nil {
			t.Fatalf("client write zero frame failed: %v", err)
		}

		// Read server's framed error reply
		reply, err := comm.ReadFrameU32(c)
		if err != nil {
			t.Fatalf("client read reply failed: %v", err)
		}
		if !bytes.HasPrefix(reply, []byte("ERR: empty frame:")) {
			t.Fatalf("unexpected reply: %q", string(reply))
		}

		// Confirm the handler is still running (has not closed)
		select {
		case <-done:
			t.Fatal("handler exited unexpectedly after zero-length frame")
		case <-time.After(100 * time.Millisecond):
			// still alive — good
		}

		// Send a second zero frame to confirm it keeps looping
		if err := comm.WriteFrameU32(c, nil); err != nil {
			t.Fatalf("second zero frame write failed: %v", err)
		}
		reply2, err := comm.ReadFrameU32(c)
		if err != nil {
			t.Fatalf("client read second reply failed: %v", err)
		}
		if !bytes.HasPrefix(reply2, []byte("ERR: empty frame:")) {
			t.Fatalf("unexpected second reply: %q", string(reply2))
		}
	})
}

// Partial payload => ReadFrameU32 errors; server replies "ERR: invalid frame:"
// then returns.
func TestHandleConnection_InvalidFrame_RepliesErrorAndExits(t *testing.T) {
	test.WithTimeout(t, 3*time.Second, func(t *testing.T) {
		s := &Server{}
		c, done, cleanup := startTestServer(t, s)
		defer cleanup()

		// Header claims 5 bytes
		var hdr [4]byte
		binary.BigEndian.PutUint32(hdr[:], 5)

		if _, err := c.Write(hdr[:]); err != nil {
			t.Fatalf("client write header failed: %v", err)
		}
		// Write only 2 bytes of the body
		if _, err := c.Write([]byte{1, 2}); err != nil {
			t.Fatalf("client write partial body failed: %v", err)
		}

		// Half-close the write side to cause io.ReadFull on the server to error
		if tcp, ok := c.(*net.TCPConn); ok {
			_ = tcp.CloseWrite()
		} else {
			// Fallback: full close; we'll still be able to read the server's
			// reply because it writes before return
			// NOTE: We will re-dial to read if the OS closes both ways, but on
			// TCP this usually still allows reads.
		}

		// Read server's framed error reply
		reply, err := comm.ReadFrameU32(c)
		if err != nil {
			t.Fatalf("client read error reply failed: %v", err)
		}
		if !bytes.HasPrefix(reply, []byte("ERR: invalid frame:")) {
			t.Fatalf("unexpected reply: %q", string(reply))
		}

		// Handler should exit after responding to invalid frame
		select {
		case <-done:
			// ok
		case <-time.After(800 * time.Millisecond):
			t.Fatal("server did not exit after invalid frame")
		}
	})
}
