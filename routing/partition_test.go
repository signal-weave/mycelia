package routing

import (
	"testing"
	"time"

	"mycelia/globals"
	"mycelia/protocol"
	"mycelia/test"
)

// Verifies that a partition reads from its input channel, applies transformers,
// and forwards the transformed message to the next channel in the route.
func TestPartition_TransformAndForwardToNextChannel(t *testing.T) {
	// Keep timings tight but reliable for CI.
	globals.DefaultNumPartitions = 1
	globals.TransformTimeout = 1 * time.Second
	globals.AutoConsolidate = false
	globals.PrintTree = false

	// Transformer for the FIRST channel (c1): echoes prefix + body over TCP.
	const prefix = "P:"
	xAddr, stopX := test.MockTwoWayServer(t, prefix)
	defer stopX()

	// Subscriber for the SECOND channel (c2): captures bodies it receives.
	subAddr, gotBody, stopSub := test.MockOneWayServer(t)
	defer stopSub()

	// Build a route with TWO channels so we can verify forwarding:
	// c1 has transformer; c2 has subscriber.
	r := &route{name: "r", channels: []*channel{}}
	c1 := r.channel("c1")
	c2 := r.channel("c2")

	// Wire c1 with the transformer; c2 with the subscriber.
	c1.addTransformer(transformer{Address: xAddr})
	c2.addSubscriber(subscriber{Address: subAddr})

	// Grab the first (and only) partition of c1 and push directly into its input.
	// This exercises partition.loop without going through channel.enqueue.
	if len(c1.partitions) != 1 {
		t.Fatalf("expected 1 partition, got %d", len(c1.partitions))
	}
	p := c1.partitions[0]

	// Build a delivery (Arg3 is used as the hash key by channels; for partition
	// we directly feed the command).
	body := "data"
	cmd := protocol.NewCommand(
		globals.OBJ_DELIVERY,
		globals.CMD_SEND,
		"client", "uid-partition",
		r.name, c1.name, "any-key", "",
		[]byte(body),
	)

	// Send straight to the partition's input channel.
	select {
	case p.in <- cmd:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout sending to partition input")
	}

	// Expect the subscriber on c2 (next channel) to receive the transformed body.
	want := prefix + body
	select {
	case got := <-gotBody:
		if got != want {
			t.Fatalf("subscriber on next channel got %q, want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for subscriber on next channel")
	}
}

// Optional: also verify that a partition with local subscribers fans out before forwarding.
// (Useful if you want to assert partition waits for fan-out via wg before enqueueing next.)
func TestPartition_FanoutThenForward(t *testing.T) {
	globals.DefaultNumPartitions = 1
	globals.TransformTimeout = 1 * time.Second
	globals.AutoConsolidate = false
	globals.PrintTree = false

	const prefix = "X:"
	xAddr, stopX := test.MockTwoWayServer(t, prefix)
	defer stopX()

	// Local subscriber on c1 (to ensure fan-out happens), and a downstream subscriber on c2
	// to ensure forwarding occurs as well.
	s1Addr, got1, stop1 := test.MockOneWayServer(t)
	defer stop1()
	s2Addr, got2, stop2 := test.MockOneWayServer(t)
	defer stop2()

	r := &route{name: "r2", channels: []*channel{}}
	c1 := r.channel("c1")
	c2 := r.channel("c2")

	c1.addTransformer(transformer{Address: xAddr})
	c1.addSubscriber(subscriber{Address: s1Addr})
	c2.addSubscriber(subscriber{Address: s2Addr})

	if len(c1.partitions) != 1 {
		t.Fatalf("expected 1 partition, got %d", len(c1.partitions))
	}
	p := c1.partitions[0]

	body := "m"
	cmd := protocol.NewCommand(
		globals.OBJ_DELIVERY,
		globals.CMD_SEND,
		"client", "uid-fanout",
		r.name, c1.name, "key", "",
		[]byte(body),
	)

	// Feed into partition input directly.
	select {
	case p.in <- cmd:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout sending to partition input")
	}

	want := prefix + body

	// Both subscribers should get the transformed payload.
	deadline := time.After(2 * time.Second)
	gotLocal := false
	gotDownstream := false
	for !(gotLocal && gotDownstream) {
		select {
		case v := <-got1:
			if v != want {
				t.Fatalf("local subscriber got %q, want %q", v, want)
			}
			gotLocal = true
		case v := <-got2:
			if v != want {
				t.Fatalf("downstream subscriber got %q, want %q", v, want)
			}
			gotDownstream = true
		case <-deadline:
			t.Fatalf("timed out; gotLocal=%v gotDownstream=%v", gotLocal, gotDownstream)
		}
	}
}
