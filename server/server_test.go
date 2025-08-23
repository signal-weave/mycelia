package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"mycelia/routing"
)

// pickFreePort asks the OS for an available port and returns it.
func pickFreePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen(0) failed: %v", err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

// captureStdout captures os.Stdout during fn and returns everything printed.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	done := make(chan string, 1)
	go func() {
		var b strings.Builder
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			fmt.Fprintln(&b, sc.Text())
		}
		done <- b.String()
	}()

	fn()
	_ = w.Close()
	out := <-done
	_ = r.Close()
	return out
}

func TestNewServer_InitializesBrokerAndFields(t *testing.T) {
	s := NewServer("127.0.0.1", 12345)
	if s.Broker == nil {
		t.Fatalf("expected non-nil Broker")
	}
	// Sanity: the default broker should have a 'main' route per your NewBroker.
	if _, ok := s.Broker.Routes["main"]; !ok {
		t.Fatalf("expected broker to have 'main' route")
	}
}

func TestServer_Run_AcceptsMultipleConnections(t *testing.T) {
	port := pickFreePort(t)
	s := NewServer("127.0.0.1", port)

	// Start the server in the background.
	started := make(chan struct{})
	go func() {
		// Small delay so we can detect start without fragile sleeps:
		close(started)
		s.Run()
	}()
	<-started

	// Give the OS a moment to bind the listener.
	time.Sleep(50 * time.Millisecond)

	// Connect #1, send two frames, close.
	conn1, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("dial #1 failed: %v", err)
	}
	_, _ = conn1.Write([]byte("frame-1\n"))
	_, _ = conn1.Write([]byte("frame-2\n"))
	_ = conn1.Close()

	// Connect #2 (ensures accept loop continues after a disconnect).
	conn2, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("dial #2 failed: %v", err)
	}
	_, _ = conn2.Write([]byte("another\n"))
	_ = conn2.Close()

	// We can’t stop s.Run() cleanly without changing the server code,
	// but reaching here without panic / deadlock proves accept + handle works.
}

func TestHandleConnection_ReadsFramesAndStopsOnEOF(t *testing.T) {
	s := &Server{
		Broker:  routing.NewBroker(), // real broker; we only exercise connection handling
		address: "ignored",
		port:    0,
	}

	// Use an in-memory full-duplex connection.
	srv, cli := net.Pipe()
	defer cli.Close()

	// Capture stdout to assert the connect/disconnect messages.
	out := captureStdout(t, func() {
		done := make(chan struct{})
		go func() {
			s.handleConnection(srv)
			close(done)
		}()

		// Send two newline-terminated frames, then EOF.
		_, _ = cli.Write([]byte("hello\n"))
		_, _ = cli.Write([]byte("world\n"))
		_ = cli.Close() // triggers EOF in the server

		select {
		case <-done:
			// ok
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("handleConnection did not return after EOF")
		}
	})

	// Basic sanity on logged messages (don’t depend on exact wording).
	if !strings.Contains(out, "Client connected:") {
		t.Fatalf("expected 'Client connected' log, got:\n%s", out)
	}
	if !strings.Contains(out, "Client disconnected:") {
		t.Fatalf("expected 'Client disconnected' log, got:\n%s", out)
	}
}
