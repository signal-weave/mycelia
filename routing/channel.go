package routing

import (
	"fmt"
	"hash/fnv"
	"sync"
	"sync/atomic"

	"mycelia/globals"
	"mycelia/protocol"
	"mycelia/str"
)

// Channels are the subscription buckets that fill routes. A subscriber
// specifies a route + channel to subscribe to. The channel will run a delivery
// through each of its transformers before forwarding it to the subscriber.
//
// More than one channel may exist in a route - Subscribers choose which
// transform "checkpoint" they wish to subscribe to. i.e. a delivery transformed
// by one channel will then be sent to the next channel in a route.
type channel struct {
	mutex sync.RWMutex
	route *route
	name  string

	hash func([]byte) uint32

	transformers []transformer
	subscribers  []subscriber

	// The authoritative slices.
	// Load these after editing to get fresh copy.
	tSnap atomic.Value
	sSnap atomic.Value

	partitions []*partition
}

func newChannel(r *route, name string, numPartitions int) *channel {
	hash := func(b []byte) uint32 {
		h := fnv.New32a()
		_, _ = h.Write(b)
		return h.Sum32()
	}

	ch := &channel{
		route: r,
		name:  name,
		hash:  hash,
	}
	ch.tSnap.Store([]transformer{})
	ch.sSnap.Store([]subscriber{})

	partitions := []*partition{}
	for range numPartitions {
		np := newPartition(r, ch)
		partitions = append(partitions, np)
		np.in = make(chan *protocol.Command, globals.PartitionChanSize)
		go np.start()
	}
	ch.partitions = partitions

	return ch
}

func (ch *channel) addTransformer(t transformer) {
	ch.mutex.Lock()
	for _, existing := range ch.transformers {
		if existing.Address == t.Address {
			ch.mutex.Unlock()
			return
		}
	}
	ch.transformers = append(ch.transformers, t)
	snap := append([]transformer(nil), ch.transformers...)
	ch.mutex.Unlock()
	ch.tSnap.Store(snap)

	str.ActionPrint(
		fmt.Sprintf("Added transformer at address: %s", t.Address),
	)
}

func (ch *channel) removeTransformer(t transformer) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	for i, transformer := range ch.transformers {
		if t.Address == transformer.Address {
			ch.transformers = append(
				ch.transformers[:i], ch.transformers[i+1:]...,
			)
			break
		}
	}
	str.ActionPrint(
		fmt.Sprintf("Removed transformer for address: %s", t.Address),
	)
	ch.checkEmptyChannel()
}

func (ch *channel) addSubscriber(s subscriber) {
	ch.mutex.Lock()
	for _, existing := range ch.subscribers {
		if existing.Address == s.Address {
			ch.mutex.Unlock()
			return
		}
	}
	ch.subscribers = append(ch.subscribers, s)
	snap := append([]subscriber(nil), ch.subscribers...)
	ch.mutex.Unlock()
	ch.sSnap.Store(snap)

	str.ActionPrint(
		fmt.Sprintf("Added subscriber at address: %s", s.Address),
	)
}

func (ch *channel) removeSubscriber(s subscriber) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	for i, subscriber := range ch.subscribers {
		if s.Address == subscriber.Address {
			ch.subscribers = append(ch.subscribers[:i], ch.subscribers[i+1:]...)
			break
		}
	}
	str.ActionPrint(
		fmt.Sprintf("Removed subscriber for address: %s", s.Address),
	)
	ch.checkEmptyChannel()
}

func (ch *channel) checkEmptyChannel() {
	if !globals.AutoConsolidate {
		return
	}
	if len(ch.subscribers) == 0 && len(ch.transformers) == 0 {
		ch.mutex.Lock()
		parts := ch.partitions
		ch.partitions = nil
		ch.mutex.Unlock()

		for _, p := range parts {
			p.stop()
		}

		ch.route.removeChannel(ch.name)
	}
}

func (c *channel) enqueue(m *protocol.Command) {
	c.mutex.RLock()
	parts := c.partitions
	c.mutex.RUnlock()

	if len(parts) == 0 {
		return // Channel is closed / removed
	}

	key := []byte(m.Arg3) // address on messages
	idx := int(c.hash(key)) % len(c.partitions)
	c.partitions[idx].in <- m
}
