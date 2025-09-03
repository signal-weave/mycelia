package routing

import (
	"testing"
	"time"

	"mycelia/globals"
	"mycelia/protocol"
	"mycelia/test"
)

// route.enqueue should push into the *first* channel; partition forwards to the next channel.
func TestRoute_Enqueue_TransformsAndForwardsToNextChannel(t *testing.T) {
	// Test tuning
	globals.DefaultNumPartitions = 1
	globals.TransformTimeout = 1 * time.Second
	globals.AutoConsolidate = false
	globals.PrintTree = false

	// Transformer on c1: echoes prefix + body
	const prefix = "R:"
	xAddr, stopX := test.MockTwoWayServer(t, prefix)
	defer stopX()

	// Subscriber on c2: captures bodies
	subAddr, got, stopSub := test.MockOneWayServer(t)
	defer stopSub()

	// Build route with two channels: c1 (transformer), c2 (subscriber)
	r := &route{name: "r", channels: []*channel{}}
	c1 := r.channel("c1")
	c2 := r.channel("c2")

	c1.addTransformer(transformer{Address: xAddr})
	c2.addSubscriber(subscriber{Address: subAddr})

	// Send a delivery via route.enqueue (must enter c1, then forward to c2)
	body := "hello"
	cmd := protocol.NewObject(
		globals.OBJ_DELIVERY,
		globals.CMD_SEND,
		"client", "uid-route",
		r.name, c1.name, "key", "",
		[]byte(body),
	)
	r.enqueue(cmd)

	// Expect subscriber on c2 to receive transformed payload
	want := prefix + body
	select {
	case gotBody := <-got:
		if gotBody != want {
			t.Fatalf("subscriber got %q, want %q", gotBody, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for subscriber on next channel")
	}
}

// getNextChannel should return the next channel, or nil when at the end.
func TestRoute_GetNextChannel(t *testing.T) {
	globals.DefaultNumPartitions = 1
	globals.AutoConsolidate = false

	r := &route{name: "r2", channels: []*channel{}}
	c1 := r.channel("c1")
	c2 := r.channel("c2")
	c3 := r.channel("c3")

	if next := r.getNextChannel(c1); next == nil || next.name != c2.name {
		t.Fatalf("getNextChannel(c1) = %v, want c2", next)
	}
	if next := r.getNextChannel(c2); next == nil || next.name != c3.name {
		t.Fatalf("getNextChannel(c2) = %v, want c3", next)
	}
	if next := r.getNextChannel(c3); next != nil {
		t.Fatalf("getNextChannel(c3) = %v, want nil", next)
	}
}

// removeChannel should remove by name and maintain order of remaining channels.
func TestRoute_RemoveChannel(t *testing.T) {
	globals.DefaultNumPartitions = 1
	globals.AutoConsolidate = false

	r := &route{name: "r3", channels: []*channel{}}
	c1 := r.channel("c1")
	c2 := r.channel("c2")
	_ = c1
	_ = c2

	// Remove the first channel
	r.removeChannel("c1")

	// The only remaining channel should be c2, and it should now be first
	if len(r.channels) != 1 {
		t.Fatalf("len(channels) = %d, want 1", len(r.channels))
	}
	if r.channels[0].name != "c2" {
		t.Fatalf("remaining channel = %s, want c2", r.channels[0].name)
	}

	// Removing a non-existent channel should be a no-op (no panic, no change)
	r.removeChannel("does-not-exist")
	if len(r.channels) != 1 || r.channels[0].name != "c2" {
		t.Fatalf("route changed unexpectedly after removing non-existent channel")
	}
}
