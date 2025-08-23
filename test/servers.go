package test

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// -----------------------------------------------------------------------------
// Herein is a boilerplate server for testing routing components or other
// packages that require talking to an external node.
// -----------------------------------------------------------------------------

// A mock TCP server for testing components that communicate with an external
// machine or source.
// This server only listens, does not respond.
// Listens on 127.0.0.1:0 with an 8192 buffer size.
func MockOneWayServer(t *testing.T) (addr string, gotBody <-chan string, stop func()) {
	t.Helper()
	randAddr := fmt.Sprintf("127.0.0.1:%d", FirstFreeTCPPort(t))
	ln, err := net.Listen("tcp", randAddr)
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	addr = ln.Addr().String()
	bodyCh := make(chan string, 1)
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 8192)
		n, _ := conn.Read(buf)
		bodyCh <- string(buf[:n])
	}()

	stop = func() {
		_ = ln.Close()
		select {
		case <-done:
		case <-time.After(200 * time.Millisecond):
		}
	}

	return addr, bodyCh, stop
}

// A mock TCP server for testing components that communicate with an external
// machine or source.
// To differentiate servers, specify a unique prefix.
// Echoes back with prefix + message body.
// Listens on 127.0.0.1:0 with an 8192 buffer size.
func MockTwoWayServer(t *testing.T, prefix string) (addr string, stop func()) {
	t.Helper()
	randAddr := fmt.Sprintf("127.0.0.1:%d", FirstFreeTCPPort(t))
	ln, err := net.Listen("tcp", randAddr)
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	addr = ln.Addr().String()

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buf := make([]byte, 8192)
		n, _ := conn.Read(buf)
		out := []byte(prefix + string(buf[:n]))
		_, _ = conn.Write(out)
	}()

	stop = func() {
		_ = ln.Close()
		select {
		case <-done:
		case <-time.After(200 * time.Millisecond):
		}
	}
	return addr, stop
}
