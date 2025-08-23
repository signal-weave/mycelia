package route

import (
	"testing"
	"time"

	"mycelia/boot"
	"mycelia/commands"
	"mycelia/test"
)

func TestRoute_AddChannel_DeduplicatesByName(t *testing.T) {
	rt := NewRoute("orders")

	rt.AddChannel(&commands.AddChannel{Name: "primary"})
	rt.AddChannel(&commands.AddChannel{Name: "primary"}) // duplicate
	rt.AddChannel(&commands.AddChannel{Name: "audit"})

	if got, want := len(rt.Channels), 2; got != want {
		t.Fatalf("want %d channels, got %d", want, got)
	}

	if ch, ok := rt.GetChannel("primary"); !ok || ch == nil {
		t.Fatalf("expected to find channel 'primary'")
	}
	if _, ok := rt.GetChannel("missing"); ok {
		t.Fatalf("did not expect to find channel 'missing'")
	}
}

func TestRoute_AddSubscriber_ToExistingChannel_AndProcess(t *testing.T) {
	// Short transform timeouts (even though we don't use transformers here).
	old := boot.RuntimeCfg.TransformTimeout
	boot.RuntimeCfg.TransformTimeout = 1
	t.Cleanup(func() { boot.RuntimeCfg.TransformTimeout = old })

	rt := NewRoute("orders")
	rt.AddChannel(&commands.AddChannel{Name: "primary"})

	recvAddr, got, stop := test.MockOneWayServer(t)
	t.Cleanup(stop)

	// Add the subscriber to the existing channel.
	rt.AddSubscriber(&commands.AddSubscriber{
		Route:   "orders",
		Channel: "primary",
		Address: recvAddr,
	})

	// Send message through the route.
	msg := &commands.SendMessage{
		ID:    "id-1",
		Route: "orders.created",
		Body:  "hello",
	}
	rt.ProcessMessage(msg)

	select {
	case b := <-got:
		if b != "hello" {
			t.Fatalf("subscriber received %q; want %q", b, "hello")
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatalf("subscriber did not receive body in time")
	}
}

func TestRoute_AddSubscriber_ChannelMissing_NoPanic(t *testing.T) {
	rt := NewRoute("orders")
	// No channels added.

	// Should log and return; nothing added, no panic.
	rt.AddSubscriber(&commands.AddSubscriber{
		Route:   "orders",
		Channel: "does-not-exist",
		Address: "127.0.0.1:65000",
	})

	// Still zero channels.
	if got := len(rt.Channels); got != 0 {
		t.Fatalf("expected 0 channels; got %d", got)
	}
}

func TestRoute_AddTransformer_ChannelMissing_NoPanic(t *testing.T) {
	rt := NewRoute("orders")
	// Should log and return; nothing added, no panic.
	rt.AddTransformer(&commands.AddTransformer{
		Route:   "orders",
		Channel: "does-not-exist",
		Address: "127.0.0.1:65000",
	})
	// Still zero channels.
	if got := len(rt.Channels); got != 0 {
		t.Fatalf("expected 0 channels; got %d", got)
	}
}

func TestRoute_ProcessMessage_ChainsAcrossChannels_AndFansOut(t *testing.T) {
	// Keep transformer reads snappy.
	old := boot.RuntimeCfg.TransformTimeout
	boot.RuntimeCfg.TransformTimeout = 1
	t.Cleanup(func() { boot.RuntimeCfg.TransformTimeout = old })

	rt := NewRoute("orders")

	// Add two channels to the route.
	rt.AddChannel(&commands.AddChannel{Name: "c1"})
	rt.AddChannel(&commands.AddChannel{Name: "c2"})

	// Transformers: c1 adds "A:", c2 adds "B:".
	addrA, stopA := test.MockTwoWayServer(t, "A:")
	addrB, stopB := test.MockTwoWayServer(t, "B:")
	t.Cleanup(stopA)
	t.Cleanup(stopB)

	rt.AddTransformer(&commands.AddTransformer{
		Route:   "orders",
		Channel: "c1",
		Address: addrA,
	})
	rt.AddTransformer(&commands.AddTransformer{
		Route:   "orders",
		Channel: "c2",
		Address: addrB,
	})

	// Subscribers: one on each channel, capture exactly what arrives.
	addr1, got1, stop1 := test.MockOneWayServer(t)
	addr2, got2, stop2 := test.MockOneWayServer(t)
	t.Cleanup(stop1)
	t.Cleanup(stop2)

	rt.AddSubscriber(&commands.AddSubscriber{
		Route:   "orders",
		Channel: "c1",
		Address: addr1,
	})
	rt.AddSubscriber(&commands.AddSubscriber{
		Route:   "orders",
		Channel: "c2",
		Address: addr2,
	})

	// Send through route.
	msg := &commands.SendMessage{
		ID:    "id-2",
		Route: "orders.created",
		Body:  "payload",
	}
	rt.ProcessMessage(msg)

	// Expect c1 subscriber to see "A:payload".
	select {
	case b := <-got1:
		if b != "A:payload" {
			t.Fatalf("c1 subscriber body: want %q, got %q", "A:payload", b)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatalf("c1 subscriber did not receive body in time")
	}

	// Expect c2 subscriber to see **B:A:payload** (i.e., channel 2 receives the
	// output of channel 1). If this assertion fails and you see "B:payload", it
	// means Route.ProcessMessage is not forwarding the transformed output to
	// the next channel.
	select {
	case b := <-got2:
		want := "B:A:payload"
		if b != want {
			t.Fatalf("c2 subscriber body: want %q, got %q", want, b)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatalf("c2 subscriber did not receive body in time")
	}
}
