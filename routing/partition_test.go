package routing

import (
	"bytes"
	"mycelia/globals"
	"sync"
	"testing"

	"github.com/signal-weave/rhizome"
)

func TestPartition_Transforms_Delivers_Then_ForwardsToNext(t *testing.T) {
	var applyCalls []string
	partLoadTransformers = func(p *partition) []transformer { return []transformer{{}, {}} }
	partApply = func(_ transformer, o *rhizome.Object) (*rhizome.Object, error) {
		applyCalls = append(applyCalls, string(o.Payload))
		o2 := *o
		o2.Payload = append(append([]byte(nil), o.Payload...), 'a'+byte(len(applyCalls)-1))
		return &o2, nil
	}

	// capture deliveries + forward with a WG so we know when we're done
	var delivered [][]byte
	var delMu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(3) // 2 deliveries + 1 forward

	partSelectSubscribers = func(p *partition) []subscriber {
		return []subscriber{{Address: "s1"}, {Address: "s2"}}
	}
	partDeliver = func(_ subscriber, o *rhizome.Object) {
		delMu.Lock()
		delivered = append(delivered, append([]byte(nil), o.Payload...))
		delMu.Unlock()
		wg.Done()
	}

	var forwarded [][]byte
	partNext = func(p *partition) *channel { return &channel{name: "next"} }
	partEnqueue = func(_ *channel, o *rhizome.Object) {
		forwarded = append(forwarded, append([]byte(nil), o.Payload...))
		wg.Done()
	}

	p := newTestPartition()
	msg := makeFakeObj("X")
	p.in <- msg

	// Wait for 2 deliveries + 1 forward, then stop
	wg.Wait()
	p.stop()

	// Assert transforms
	if len(applyCalls) != 2 {
		t.Fatalf("expected 2 transformer applications, got %d", len(applyCalls))
	}

	// The final payload after, two appends is "Xab"
	want := []byte("Xab")

	if len(delivered) != 2 {
		t.Fatalf("expected delivery to 2 subscribers, got %d", len(delivered))
	}
	for i, d := range delivered {
		if !bytes.Equal(d, want) {
			t.Fatalf("delivery %d: got %q, want %q", i, d, want)
		}
	}

	if len(forwarded) != 1 || !bytes.Equal(forwarded[0], want) {
		t.Fatalf("forwarded got %v, want single %q", forwarded, want)
	}
}

func TestPartition_NoNext_AckOnSent_Path(t *testing.T) {
	// no transforms, no subs, no next → should hit ack-on-sent
	partLoadTransformers = func(p *partition) []transformer { return nil }
	partSelectSubscribers = func(p *partition) []subscriber { return nil }
	partNext = func(p *partition) *channel { return nil }

	var wg sync.WaitGroup
	wg.Add(1)

	ackCalled := false
	partAck = func(o *rhizome.Object) error {
		ackCalled = true
		o.Response.Ack = globals.AckSent
		wg.Done()
		return nil
	}

	p := newTestPartition()
	msg := makeFakeObj("ping")
	msg.AckPlcy = globals.AckPlcyOnsent
	p.in <- msg

	wg.Wait()
	p.stop()

	if !ackCalled {
		t.Fatalf("expected partAck to be called")
	}
}
