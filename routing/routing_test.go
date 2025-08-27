package routing_test

import (
	"testing"
	"time"

	"mycelia/boot"
	"mycelia/commands"
	"mycelia/routing"
	"mycelia/test"
)

// TestRouting_TransformThenFanout ensures that a delivery sent through a route
// hits all transformers in insertion order and then is delivered to all
// subscribers with the final transformed body.
func TestRouting_TransformThenFanout(t *testing.T) {
	// Keep transformer read deadlines reasonable.
	boot.RuntimeCfg.TransformTimeout = 1 // seconds

	// 1) Spin up two echo transformers with distinct prefixes.
	t1Addr, stopT1 := test.MockTwoWayServer(t, "A:")
	t2Addr, stopT2 := test.MockTwoWayServer(t, "B:")
	defer stopT1()
	defer stopT2()

	// 2) Spin up two one-way subscribers to capture delivered bodies.
	s1Addr, s1BodyCh, stopS1 := test.MockOneWayServer(t)
	s2Addr, s2BodyCh, stopS2 := test.MockOneWayServer(t)
	defer stopS1()
	defer stopS2()

	// 3) Build the topology using the fluent API.
	b := routing.NewBroker()
	r := b.Route("default")
	ch := r.Channel("my_new_channel")

	ch.AddTransformer(*routing.NewTransformer(t1Addr)) // first A:
	ch.AddTransformer(*routing.NewTransformer(t2Addr)) // then B:

	ch.AddSubscriber(*routing.NewSubscriber(s1Addr))
	ch.AddSubscriber(*routing.NewSubscriber(s2Addr))

	// 4) Send a delivery through the route.
	msg := &commands.Delivery{
		ID:    "test-uuid-123",
		Route: "default",
		Body:  []byte("hello"),
	}
	r.ProcessDelivery(msg)

	// 5) Expect transformed body to be "B:A:hello".
	want := "B:A:hello"

	select {
	case got := <-s1BodyCh:
		if got != want {
			t.Fatalf("subscriber 1 got %q; want %q", got, want)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for subscriber 1 to receive")
	}

	select {
	case got := <-s2BodyCh:
		if got != want {
			t.Fatalf("subscriber 2 got %q; want %q", got, want)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for subscriber 2 to receive")
	}
}
