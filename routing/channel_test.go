package routing

import (
	"testing"
	"time"

	"mycelia/globals"
	"mycelia/test"
)

func TestChannel_ProcessDelivery_TransformsInOrder_AndFanout(t *testing.T) {
	// Transformer #1 and #2 echo back "P1:" + body and "P2:" + body respectively.
	t1Addr, stop1 := test.MockTwoWayServer(t, "P1:")
	t.Cleanup(stop1)
	t2Addr, stop2 := test.MockTwoWayServer(t, "P2:")
	t.Cleanup(stop2)

	// Subscriber sink to observe final payload sent by channel fan-out.
	subAddr, gotBody, stopSub := test.MockOneWayServer(t)
	t.Cleanup(stopSub)

	globals.TransformTimeout = 2 * time.Second

	// Build channel (route not required here; we don't call Remove* which
	// triggers checkEmptyChannel).
	ch := &Channel{name: "chanA"}
	ch.AddTransformer(*NewTransformer(t1Addr))
	ch.AddTransformer(*NewTransformer(t2Addr))
	ch.AddSubscriber(*NewSubscriber(subAddr))

	in := msg("hello")
	out := ch.ProcessDelivery(in)

	// Expect sequential transform: out = P2:(P1:(hello))
	want := "P2:P1:hello"
	if string(out.Payload) != want {
		t.Fatalf(
			"unexpected transformed payload: got %q want %q",
			string(out.Payload), want,
		)
	}

	// Ensure subscriber received the transformed payload.
	select {
	case got := <-gotBody:
		if got != want {
			t.Fatalf("subscriber saw wrong body: got %q want %q", got, want)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for subscriber to receive payload")
	}
}

func TestChannel_ProcessDelivery_SkipsFailedTransformer(t *testing.T) {
	// First transformer fails to dial; second works.
	badAddr := "127.0.0.1:1" // likely closed
	okAddr, stop := test.MockTwoWayServer(t, "OK:")
	t.Cleanup(stop)

	subAddr, gotBody, stopSub := test.MockOneWayServer(t)
	t.Cleanup(stopSub)

	globals.TransformTimeout = 1 * time.Second

	ch := &Channel{name: "chanB"}
	ch.AddTransformer(*NewTransformer(badAddr))
	ch.AddTransformer(*NewTransformer(okAddr))
	ch.AddSubscriber(*NewSubscriber(subAddr))

	in := msg("x")
	out := ch.ProcessDelivery(in)

	want := "OK:x"
	if string(out.Payload) != want {
		t.Fatalf(
			"unexpected payload after skipping failed transformer: got %q want %q",
			string(out.Payload), want,
		)
	}

	select {
	case got := <-gotBody:
		if got != want {
			t.Fatalf("subscriber saw wrong body: got %q want %q", got, want)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for subscriber to receive payload")
	}
}

func TestChannel_AddSubscriber_Deduplicates(t *testing.T) {
	subAddr, gotBody, stop := test.MockOneWayServer(t)
	t.Cleanup(stop)

	ch := &Channel{name: "chanC"}
	ch.AddSubscriber(*NewSubscriber(subAddr))
	// attempt to add duplicate
	ch.AddSubscriber(*NewSubscriber(subAddr))

	// No transformers; message should be forwarded as-is once.
	in := msg("once")
	_ = ch.ProcessDelivery(in)

	// Expect exactly one delivery.
	select {
	case got := <-gotBody:
		if got != "once" {
			t.Fatalf("subscriber saw wrong body: got %q want %q", got, "once")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for subscriber to receive payload")
	}

	// Ensure no second delivery sneaks in.
	select {
	case extra := <-gotBody:
		t.Fatalf("received duplicate delivery: %q", extra)
	case <-time.After(200 * time.Millisecond):
		// good: no duplicate
	}
}
