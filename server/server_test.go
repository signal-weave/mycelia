// server_test.go
package server

import (
	"encoding/binary"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"mycelia/globals"
)

// --- Helpers ---

// readFrameU32 reads a big-endian u32 length-prefixed frame (payload only) from c.
func readFrameU32(t *testing.T, c net.Conn) ([]byte, error) {
	t.Helper()
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	var hdr [4]byte
	if _, err := io.ReadFull(c, hdr[:]); err != nil {
		return nil, err
	}
	n := binary.BigEndian.Uint32(hdr[:])
	if n == 0 {
		return []byte{}, nil
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(c, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// --- Tests ---

func TestUpdateListener_OpensAndReplacesListener(t *testing.T) {
	// Arrange: ask OS for an ephemeral port twice
	globals.Address = "127.0.0.1"
	globals.Port = 0

	s := NewServer("", 0)
	if s.listener != nil {
		t.Fatalf("expected nil listener before UpdateListener, got non-nil")
	}

	// Act: first bind
	s.UpdateListener() // should open a new listener and mirror globals.Address/Port
	if s.listener == nil {
		t.Fatalf("listener was not created")
	}
	firstAddr := s.listener.Addr().String()
	if firstAddr == "" {
		t.Fatalf("listener Addr() is empty")
	}

	// Act: second bind (new ephemeral port), should replace old listener
	globals.Port = 0
	s.UpdateListener()
	if s.listener == nil {
		t.Fatalf("second listener was not created")
	}
	secondAddr := s.listener.Addr().String()
	if secondAddr == "" {
		t.Fatalf("second listener Addr() is empty")
	}
	if firstAddr == secondAddr {
		t.Fatalf("expected listener to be replaced with a different address, got same: %s", firstAddr)
	}

	// Cleanup
	s.Shutdown()
}

func TestHandleConnection_WritesERR_OnShortHeader(t *testing.T) {
	// Spin up a real TCP listener just for this test
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	s := NewServer("", 0)

	// Accept exactly one connection and hand it to the handler
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, aerr := ln.Accept()
		if aerr != nil {
			t.Logf("accept error: %v", aerr)
			return
		}
		s.HandleConnection(conn)
	}()

	// Client side: connect
	cconn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer cconn.Close()

	// Send fewer than 4 bytes to trigger short read on the server
	if _, err := cconn.Write([]byte{0x00, 0x01}); err != nil {
		t.Fatalf("client write: %v", err)
	}

	// Half-close the write side so the server sees EOF on header read
	if tcp, ok := cconn.(*net.TCPConn); ok {
		if err := tcp.CloseWrite(); err != nil {
			t.Fatalf("close write: %v", err)
		}
	} else {
		t.Fatalf("expected *net.TCPConn, got %T", cconn)
	}

	// Read the server's framed response and assert it starts with "ERR:"
	payload, rerr := readFrameU32(t, cconn)
	if rerr != nil {
		t.Fatalf("failed reading response frame: %v", rerr)
	}
	got := string(payload)
	if got == "" || (len(got) >= 4 && got[:4] != "ERR:") {
		t.Fatalf("expected payload starting with 'ERR:', got %q", got)
	}

	wg.Wait()
}

func TestShutdown_Idempotent(t *testing.T) {
	globals.Address = "127.0.0.1"
	globals.Port = 0

	s := NewServer("", 0)
	s.UpdateListener()
	if s.listener == nil {
		t.Fatalf("listener not created prior to Shutdown")
	}

	// First shutdown should close listener and jobs channel (if any).
	s.Shutdown()

	// Second shutdown should be safe (idempotent).
	s.Shutdown()
}
