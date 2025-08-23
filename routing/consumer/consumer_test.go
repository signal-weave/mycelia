package consumer

import (
	"net"
	"testing"
	"time"

	"mycelia/boot"
	"mycelia/commands"
)

var _ = boot.RuntimeCfg // REQUIRED for global config values.

// startConsumerServer starts a tiny TCP server that records exactly one message body.
func startConsumerServer(t *testing.T) (addr string, gotBody <-chan string, stop func()) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
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

		buf := make([]byte, 4096)
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

func TestConsumer_ConsumeMessage_SetsResolvedAndSendsBody(t *testing.T) {
	addr, gotBody, stop := startConsumerServer(t)
	t.Cleanup(stop)

	c := NewConsumer(addr)
	msg := &commands.SendMessage{
		ID:    "abc",
		Route: "orders.created",
		Body:  "hello world",
		// Status starts non-resolved; we only assert it becomes StatusResolved
	}

	c.ConsumeMessage(msg)

	// Assert status flipped to resolved.
	if msg.Status != commands.StatusResolved {
		t.Fatalf("status not set to resolved, got: %v", msg.Status)
	}

	// Assert the body was actually sent over the wire.
	select {
	case b := <-gotBody:
		if b != "hello world" {
			t.Fatalf("sent body mismatch: want %q, got %q", "hello world", b)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("server did not receive body in time")
	}
}

func TestConsumer_ConsumeMessage_DialFailure_DoesNotSetResolved(t *testing.T) {
	// Port 0 is not dialable; this should fail fast and NOT set StatusResolved.
	c := NewConsumer("127.0.0.1:0")
	msg := &commands.SendMessage{
		ID:    "x",
		Route: "r",
		Body:  "payload",
	}

	c.ConsumeMessage(msg)

	if msg.Status == commands.StatusResolved {
		t.Fatalf("status should not be resolved on dial failure")
	}
}
