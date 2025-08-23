package test

import (
	"net"
	"testing"
)

// -----------------------------------------------------------------------------
// Herein is a boilerplate server for testing communication over a port by
// provided simple helpers for obtaining a free port.
// -----------------------------------------------------------------------------

// FirstFreeTCPPort asks the OS for an ephemeral TCP port and returns it.
func FirstFreeTCPPort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0") // OS picks a free port
	if err != nil {
		t.Fatalf("listen(0) failed: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	return port
}

// FirstFreeUDPPort asks the OS for an ephemeral UDP port and returns it.
func FirstFreeUDPPort(t *testing.T) int {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen(0) failed: %v", err)
	}
	port := pc.LocalAddr().(*net.UDPAddr).Port
	return port
}
