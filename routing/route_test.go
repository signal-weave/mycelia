package routing

import (
	"testing"
	"time"

	"mycelia/globals"
	"mycelia/test"
)

func TestRoute_Channel_CreateAndReuse(t *testing.T) {
	r := &route{
		name:     "r1",
		channels: []*channel{},
	}

	c1 := r.channel("cA")
	if c1 == nil {
		t.Fatalf("expected non-nil channel")
	}
	if c1.route != r || c1.name != "cA" {
		t.Fatalf("channel fields not initialized correctly")
	}
	// Reuse same name ⇒ same pointer
	c2 := r.channel("cA")
	if c1 != c2 {
		t.Fatalf(
			"expected Channel to return the same pointer for existing name",
		)
	}
	// New name ⇒ new channel
	_ = r.channel("cB")
	if len(r.channels) != 2 {
		t.Fatalf("expected 2 channels in map, got %d", len(r.channels))
	}
}

func TestRoute_ProcessDelivery_MultiChannel_OrderAgnostic(t *testing.T) {
	// Two transformers; each echoes back "<PFX>:<body>"
	addrA, stopA := test.MockTwoWayServer(t, "A:")
	t.Cleanup(stopA)
	addrB, stopB := test.MockTwoWayServer(t, "B:")
	t.Cleanup(stopB)

	// Each channel gets its own subscriber sink to observe what it receives
	subAddr1, got1, stopS1 := test.MockOneWayServer(t)
	t.Cleanup(stopS1)
	subAddr2, got2, stopS2 := test.MockOneWayServer(t)
	t.Cleanup(stopS2)

	globals.TransformTimeout = 2 * time.Second

	r := &route{
		name:     "r2",
		channels: []*channel{},
	}

	// Build two channels with one transformer each and a subscriber each.
	// NOTE: Route.ProcessDelivery iterates over map entries in indeterminate
	// order.
	chA := r.channel("chA")
	chA.addTransformer(*newTransformer(addrA))
	chA.addSubscriber(*newSubscriber(subAddr1))

	chB := r.channel("chB")
	chB.addTransformer(*newTransformer(addrB))
	chB.addSubscriber(*newSubscriber(subAddr2))

	in := msg("x")
	r.deliver(in)

	// Because route iterates channels in undefined order, valid outcomes are:
	//   Order A→B: sub1 gets "A:x", sub2 gets "B:A:x"
	//   Order B→A: sub1 gets "A:B:x", sub2 gets "B:x"
	var gotFirst, gotSecond string
	select {
	case gotFirst = <-got1:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for subscriber 1")
	}
	select {
	case gotSecond = <-got2:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for subscriber 2")
	}

	// Accept either valid ordering
	okAB := (gotFirst == "A:x" && gotSecond == "B:A:x")
	okBA := (gotFirst == "A:B:x" && gotSecond == "B:x")
	if !(okAB || okBA) {
		t.Fatalf(
			"unexpected pair of deliveries: sub1=%q sub2=%q",
			gotFirst, gotSecond,
		)
	}
}

func TestRoute_removeChannel_RemovesEntry_NoBrokerCall(t *testing.T) {
	// We avoid triggering broker.removeEmptyRoute by leaving one channel
	// behind.
	r := &route{
		name:     "r3",
		channels: []*channel{},
	}
	_ = r.channel("keep")
	_ = r.channel("drop")

	// Should remove "drop" without panicking (broker may be nil) and without
	// trying to remove the route itself (since one channel remains).
	r.removeChannel("drop")

	if _, idx := r.channelExists("drop"); idx > -1 {
		t.Fatalf("expected channel 'drop' to be removed")
	}
	if _, idx := r.channelExists("keep"); idx < 0 {
		t.Fatalf("expected channel 'keep' to remain")
	}
}
