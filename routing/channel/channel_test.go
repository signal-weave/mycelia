package channel

import (
	"testing"
	"time"

	"mycelia/boot"
	"mycelia/commands"
	"mycelia/routing/consumer"
	"mycelia/routing/transform"
	"mycelia/test"
)

func TestChannel_RegisterSubscriber_DeduplicatesByAddress(t *testing.T) {
	ch := NewChannel("test")
	addr, _, stop := test.MockOneWayServer(t)
	t.Cleanup(stop)

	s1 := consumer.NewConsumer(addr)
	s2 := consumer.NewConsumer(addr) // same address -> should dedupe

	ch.RegisterSubscriber(s1)
	ch.RegisterSubscriber(s2)

	if got, want := len(ch.Subscribers), 1; got != want {
		t.Fatalf("want %d subscriber, got %d", want, got)
	}
}

func TestChannel_RegisterTransformer_DeduplicatesByAddress(t *testing.T) {
	ch := NewChannel("test")
	addr, stop := test.MockTwoWayServer(t, "A:")
	t.Cleanup(stop)

	tf1 := transform.NewTransformer(addr)
	tf2 := transform.NewTransformer(addr) // same address -> should dedupe

	ch.RegisterTransformer(tf1)
	ch.RegisterTransformer(tf2)

	if got, want := len(ch.Transformers), 1; got != want {
		t.Fatalf("want %d transformer, got %d", want, got)
	}
}

func TestChannel_ProcessMessage_AppliesTransformersInOrder_AndFansOut(t *testing.T) {
	// Keep transformer reads snappy.
	oldTO := boot.RuntimeCfg.TransformTimeout
	boot.RuntimeCfg.TransformTimeout = 1
	t.Cleanup(func() { boot.RuntimeCfg.TransformTimeout = oldTO })

	// Two transformers, applied in order: payload -> "A:payload" -> "B:A:payload"
	addrA, stopA := test.MockTwoWayServer(t, "A:")
	addrB, stopB := test.MockTwoWayServer(t, "B:")
	t.Cleanup(stopA)
	t.Cleanup(stopB)

	// Two subscribers (consumers) that record the bytes they receive.
	addr1, got1, stop1 := test.MockOneWayServer(t)
	addr2, got2, stop2 := test.MockOneWayServer(t)
	t.Cleanup(stop1)
	t.Cleanup(stop2)

	ch := NewChannel("orders")
	ch.RegisterTransformer(transform.NewTransformer(addrA))
	ch.RegisterTransformer(transform.NewTransformer(addrB))
	ch.RegisterSubscriber(consumer.NewConsumer(addr1))
	ch.RegisterSubscriber(consumer.NewConsumer(addr2))

	msg := &commands.SendMessage{
		ID:    "id-1",
		Route: "orders.created",
		Body:  "payload",
	}

	out := ch.ProcessMessage(msg)

	// Expect both consumers to receive the fully transformed body.
	want := "B:A:payload"

	select {
	case b := <-got1:
		if b != want {
			t.Fatalf("subscriber 1 body: want %q, got %q", want, b)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatalf("subscriber 1 did not receive body in time")
	}

	select {
	case b := <-got2:
		if b != want {
			t.Fatalf("subscriber 2 body: want %q, got %q", want, b)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatalf("subscriber 2 did not receive body in time")
	}

	// The returned message should reflect the final transform output.
	if out.Body != want {
		t.Fatalf("returned message body: want %q, got %q", want, out.Body)
	}

	// Consumers set StatusResolved; since both operate on the same pointer,
	// status should be resolved when ProcessMessage returns.
	if out.Status != commands.StatusResolved {
		t.Fatalf("status not set to resolved, got %v", out.Status)
	}
}

func TestChannel_ProcessMessage_ContinuesOnTransformerFailure(t *testing.T) {
	// Short timeout so failing transformer doesn't hang.
	oldTO := boot.RuntimeCfg.TransformTimeout
	boot.RuntimeCfg.TransformTimeout = 1
	t.Cleanup(func() { boot.RuntimeCfg.TransformTimeout = oldTO })

	// First transformer will fail to dial (port 0), second will succeed and prefix "OK:".
	badAddr := "127.0.0.1:0"
	okAddr, stopOK := test.MockTwoWayServer(t, "OK:")
	t.Cleanup(stopOK)

	recvAddr, got, stopRecv := test.MockOneWayServer(t)
	t.Cleanup(stopRecv)

	ch := NewChannel("orders")
	ch.RegisterTransformer(transform.NewTransformer(badAddr)) // will error and be skipped
	ch.RegisterTransformer(transform.NewTransformer(okAddr))  // will apply

	ch.RegisterSubscriber(consumer.NewConsumer(recvAddr))

	msg := &commands.SendMessage{
		ID:    "id-2",
		Route: "orders.created",
		Body:  "orig",
	}

	out := ch.ProcessMessage(msg)

	// Subscriber should receive the message transformed only by the successful transformer.
	want := "OK:orig"

	select {
	case b := <-got:
		if b != want {
			t.Fatalf("subscriber body: want %q, got %q", want, b)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatalf("subscriber did not receive body in time")
	}

	if out.Body != want {
		t.Fatalf("returned message body: want %q, got %q", want, out.Body)
	}
	if out.Status != commands.StatusResolved {
		t.Fatalf("status not set to resolved, got %v", out.Status)
	}
}
