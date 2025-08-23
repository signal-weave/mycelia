package transform

import (
	"net"
	"testing"
	"time"

	"mycelia/boot"
	"mycelia/commands"
	"mycelia/test"
)

func TestTransformMessage_Success(t *testing.T) {
	// Keep timeouts short in tests.
	old := boot.RuntimeCfg.TransformTimeout
	boot.RuntimeCfg.TransformTimeout = 1
	t.Cleanup(func() { boot.RuntimeCfg.TransformTimeout = old })

	addr, stop := test.MockTwoWayServer(t, "transformed:")
	t.Cleanup(stop)

	tr := NewTransformer(addr)
	in := &commands.SendMessage{
		ID:    "abc",
		Route: "orders.created",
		Body:  "hello world",
	}

	out, err := tr.TransformMessage(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// New message should be returned with same ID/Route, transformed Body.
	if out == in {
		t.Fatalf("expected a new message on success, but got original pointer")
	}
	if out.ID != in.ID || out.Route != in.Route {
		t.Fatalf("metadata not preserved: got ID=%q Route=%q",
			out.ID, out.Route)
	}
	wantBody := "transformed:hello world"
	if out.Body != wantBody {
		t.Fatalf("body mismatch: want %q, got %q", wantBody, out.Body)
	}
}

func TestTransformMessage_DialError_ReturnsOriginal(t *testing.T) {
	// No server on port 0 -> Dial should fail immediately.
	old := boot.RuntimeCfg.TransformTimeout
	boot.RuntimeCfg.TransformTimeout = 1
	t.Cleanup(func() { boot.RuntimeCfg.TransformTimeout = old })

	tr := NewTransformer("127.0.0.1:0")
	in := &commands.SendMessage{ID: "x", Route: "r", Body: "payload"}

	out, err := tr.TransformMessage(in)
	if err == nil {
		t.Fatalf("expected dial error, got nil")
	}
	// On failure, function should return the *same* original message pointer.
	if out != in {
		t.Fatalf("expected original message on error; got new message")
	}
	if out.Body != "payload" {
		t.Fatalf("original message should be unchanged; got %q", out.Body)
	}
}

func TestTransformMessage_ReadTimeout_ReturnsOriginal(t *testing.T) {
	// Force immediate timeout by setting TransformTimeout to 0 seconds.
	old := boot.RuntimeCfg.TransformTimeout
	boot.RuntimeCfg.TransformTimeout = 0
	t.Cleanup(func() { boot.RuntimeCfg.TransformTimeout = old })

	// Server accepts and reads but does NOT write a response.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	addr := ln.Addr().String()
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 4096)

		// read client payload, then intentionally never write
		_, _ = conn.Read(buf)
		time.Sleep(200 * time.Millisecond)
	}()

	tr := NewTransformer(addr)
	in := &commands.SendMessage{ID: "y", Route: "r", Body: "payload"}

	out, err := tr.TransformMessage(in)
	if err == nil {
		t.Fatalf("expected read timeout/error, got nil")
	}
	// Should return the original message on read failure.
	if out != in {
		t.Fatalf("expected original message on read error; got new message")
	}

	_ = ln.Close()
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
	}
}
