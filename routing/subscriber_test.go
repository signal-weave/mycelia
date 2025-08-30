package routing

import (
	"testing"
	"time"

	"mycelia/protocol"
	"mycelia/test"
)

func TestSubscriber_ConsumeDelivery_SendsPayload(t *testing.T) {
	addr, gotBody, stop := test.MockOneWayServer(t)
	defer stop()

	sub := NewSubscriber(addr)
	msg := &protocol.Command{
		Payload: []byte("hello, world"),
	}

	sub.ConsumeDelivery(msg)

	select {
	case got := <-gotBody:
		if got != "hello, world" {
			t.Fatalf("payload mismatch: got %q", got)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout: mock server did not receive payload")
	}
}

func TestSubscriber_ConsumeDelivery_DialFailure_NoPanic(t *testing.T) {
	// Port 1 is almost always closed locally; dialing should fail quickly.
	sub := NewSubscriber("127.0.0.1:1")
	msg := &protocol.Command{Payload: []byte("ignored")}
	// Should not panic; function handles error internally.
	sub.ConsumeDelivery(msg)
}
