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
	ch := &channel{name: "chanA"}
	ch.addTransformer(*newTransformer(t1Addr))
	ch.addTransformer(*newTransformer(t2Addr))
	ch.addSubscriber(*newSubscriber(subAddr))

	in := msg("hello")
	out := ch.deliver(in)

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

	ch := &channel{name: "chanB"}
	ch.addTransformer(*newTransformer(badAddr))
	ch.addTransformer(*newTransformer(okAddr))
	ch.addSubscriber(*newSubscriber(subAddr))

	in := msg("x")
	out := ch.deliver(in)

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

	ch := &channel{name: "chanC"}
	ch.addSubscriber(*newSubscriber(subAddr))
	// attempt to add duplicate
	ch.addSubscriber(*newSubscriber(subAddr))

	// No transformers; message should be forwarded as-is once.
	in := msg("once")
	_ = ch.deliver(in)

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
