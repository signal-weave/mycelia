package routing

import (
	"testing"
	"time"

	"mycelia/globals"
	"mycelia/protocol"
	"mycelia/test"
)

// Verifies transformation + fan-out: one transformer, two subscribers.
// The subscribers should both receive the transformed payload.
func TestChannel_FanoutAndTransform(t *testing.T) {
	// Tune globals for the unit test
	globals.DefaultNumPartitions = 1
	globals.TransformTimeout = 1 * time.Second
	globals.AutoConsolidate = false
	globals.PrintTree = false

	// Transformer (echoes prefix + payload)
	const prefix = "X:"
	xAddr, stopX := test.MockTwoWayServer(t, prefix)
	defer stopX()

	// Two subscribers that capture bodies they receive
	s1Addr, got1, stop1 := test.MockOneWayServer(t)
	defer stop1()
	s2Addr, got2, stop2 := test.MockOneWayServer(t)
	defer stop2()

	// Build a minimal route and create the channel under test
	r := &route{name: "r", channels: []*channel{}}
	ch := r.channel("c") // creates channel if missing

	// Wire a transformer + two subscribers onto the channel
	ch.addTransformer(transformer{Address: xAddr})
	ch.addSubscriber(subscriber{Address: s1Addr})
	ch.addSubscriber(subscriber{Address: s2Addr})

	// Enqueue a single delivery (Arg3 is the hashing key for partitions)
	const body = "hello"
	cmd := protocol.NewObject(
		globals.OBJ_DELIVERY,
		globals.CMD_SEND,
		"client", "uid-1",
		r.name, ch.name, "hash-key", "",
		[]byte(body),
	)
	ch.enqueue(cmd)

	// Expect both subscribers to receive the transformed payload
	want := prefix + body

	select {
	case got := <-got1:
		if got != want {
			t.Fatalf("subscriber1 got %q, want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for subscriber1")
	}

	select {
	case got := <-got2:
		if got != want {
			t.Fatalf("subscriber2 got %q, want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for subscriber2")
	}
}

// Burst test using multiple one-shot subscribers, because MockOneWayServer
// accepts a single connection and single read per instance.
func TestChannel_Burst_AllDelivered_WithManyOneShotSubscribers(t *testing.T) {
	globals.DefaultNumPartitions = 2
	globals.TransformTimeout = 1 * time.Second
	globals.AutoConsolidate = false
	globals.PrintTree = false

	// Build route/channel under test (no transformer for burst delivery)
	r := &route{name: "r2", channels: []*channel{}}
	ch := r.channel("c2")

	// Spin up N one-shot subscribers (each captures exactly one body)
	total := 10
	type subRec struct {
		addr string
		got  <-chan string
		stop func()
	}
	var subs []subRec
	for i := 0; i < total; i++ {
		addr, got, stop := test.MockOneWayServer(t)
		subs = append(subs, subRec{addr: addr, got: got, stop: stop})
		ch.addSubscriber(subscriber{Address: addr})
	}
	defer func() {
		for _, s := range subs {
			s.stop()
		}
	}()

	// Enqueue N messages
	for i := 0; i < total; i++ {
		cmd := protocol.NewObject(
			globals.OBJ_DELIVERY,
			globals.CMD_SEND,
			"client", "uid-burst",
			r.name, ch.name, "same-key", "",
			[]byte("m"),
		)
		ch.enqueue(cmd)
	}

	// Expect each one-shot subscriber to receive exactly one raw payload
	deadline := time.After(3 * time.Second)
	received := 0
	want := "m"
	for received < total {
		select {
		case <-deadline:
			t.Fatalf("received %d/%d deliveries in time", received, total)
		default:
			for _, s := range subs {
				select {
				case got := <-s.got:
					if got != want {
						t.Fatalf("subscriber got %q, want %q", got, want)
					}
					received++
				default:
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}
