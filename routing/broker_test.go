package routing_test

import (
	"net"
	"testing"
	"time"

	"mycelia/boot"
	"mycelia/commands"
	"mycelia/routing"
)

// ------helpers: tiny TCP servers----------------------------------------------

// records exactly one message body sent to it.
func startRecorderServer(t *testing.T) (addr string, gotBody <-chan string, stop func()) {
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

// echoes back prefix+body as a transformer would.
func startTransformerServer(t *testing.T, prefix string) (addr string, stop func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
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
		_, _ = conn.Write([]byte(prefix + string(buf[:n])))
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

// ------tests------------------------------------------------------------------

func TestNewBroker_HasMainRoute(t *testing.T) {
	b := routing.NewBroker()
	if b == nil {
		t.Fatal("NewBroker returned nil")
	}
	if _, ok := b.Routes["main"]; !ok {
		t.Fatalf("expected 'main' route to exist")
	}
	if got := len(b.Routes["main"].Channels); got != 0 {
		t.Fatalf("main route should start with 0 channels, got %d", got)
	}
}

func TestBroker_AddRoute_AddChannel_AddSubscriber_AddTransformer_And_SendMessage_E2E(t *testing.T) {
	// keep IO snappy in tests
	oldTO := boot.RuntimeCfg.TransformTimeout
	oldTree := boot.RuntimeCfg.PrintTree
	boot.RuntimeCfg.TransformTimeout = 1
	boot.RuntimeCfg.PrintTree = false
	t.Cleanup(func() {
		boot.RuntimeCfg.TransformTimeout = oldTO
		boot.RuntimeCfg.PrintTree = oldTree
	})

	b := routing.NewBroker()

	// add a new route
	b.AddRoute(commands.AddRoute{Name: "orders"})

	// add one channel on that route
	b.AddChannel(commands.AddChannel{Route: "orders", Name: "primary"})

	// plug a transformer that prefixes "X:"
	tAddr, tStop := startTransformerServer(t, "X:")
	t.Cleanup(tStop)
	b.AddTransformer(commands.AddTransformer{
		Route:   "orders",
		Channel: "primary",
		Address: tAddr,
	})

	// plug a subscriber that records what it receives
	rcvAddr, gotBody, rStop := startRecorderServer(t)
	t.Cleanup(rStop)
	b.AddSubscriber(commands.AddSubscriber{
		Route:   "orders",
		Channel: "primary",
		Address: rcvAddr,
	})

	// send a message through the broker
	b.SendMessage(commands.SendMessage{
		ID:    "m1",
		Route: "orders",
		Body:  "payload",
	})

	// subscriber should receive transformed body
	select {
	case body := <-gotBody:
		if body != "X:payload" {
			t.Fatalf("got body %q; want %q", body, "X:payload")
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatalf("timeout: subscriber did not receive message")
	}
}

func TestBroker_AddRoute_DuplicateDoesNotCreateAnother(t *testing.T) {
	b := routing.NewBroker()

	b.AddRoute(commands.AddRoute{Name: "orders"})
	before := len(b.Routes)

	// duplicate add
	b.AddRoute(commands.AddRoute{Name: "orders"})
	after := len(b.Routes)

	if after != before {
		t.Fatalf("duplicate AddRoute should not change route count: before=%d after=%d", before, after)
	}
}

func TestBroker_AddChannel_OnMissingRoute_NoPanic(t *testing.T) {
	b := routing.NewBroker()
	before := len(b.Routes)

	// route doesn't exist -> should warn and return, not panic
	b.AddChannel(commands.AddChannel{Route: "does.not.exist", Name: "ch"})

	after := len(b.Routes)
	if after != before {
		t.Fatalf("AddChannel on missing route should not mutate routes; before=%d after=%d", before, after)
	}
}

func TestBroker_AddSubscriber_OnMissingRoute_NoPanic(t *testing.T) {
	b := routing.NewBroker()
	before := len(b.Routes)

	b.AddSubscriber(commands.AddSubscriber{
		Route:   "nope",
		Channel: "primary",
		Address: "127.0.0.1:65000",
	})

	after := len(b.Routes)
	if after != before {
		t.Fatalf("AddSubscriber on missing route should not mutate routes; before=%d after=%d", before, after)
	}
}

func TestBroker_AddTransformer_OnMissingRoute_NoPanic(t *testing.T) {
	b := routing.NewBroker()
	before := len(b.Routes)

	b.AddTransformer(commands.AddTransformer{
		Route:   "nope",
		Channel: "primary",
		Address: "127.0.0.1:65000",
	})

	after := len(b.Routes)
	if after != before {
		t.Fatalf("AddTransformer on missing route should not mutate routes; before=%d after=%d", before, after)
	}
}

func TestBroker_SendMessage_UnknownRoute_NoPanic(t *testing.T) {
	b := routing.NewBroker()
	// should only log a warning and return
	b.SendMessage(commands.SendMessage{
		ID:    "m2",
		Route: "does.not.exist",
		Body:  "x",
	})
	// no assertions needed beyond "did not panic"
}
