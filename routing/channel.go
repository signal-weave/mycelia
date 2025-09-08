package routing

import (
	"fmt"
	"hash/fnv"
	"sync"
	"sync/atomic"

	"mycelia/globals"
	"mycelia/logging"
	"mycelia/protocol"
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

	selector   selector
	partitions []*partition
}

func newChannel(
	r *route, name string, numPartitions int, strat globals.SelectionStrategy,
) *channel {
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
		np.in = make(chan *protocol.Object, globals.PartitionChanSize)
		np.start()
	}
	ch.partitions = partitions

	ch.selector = newSelector(ch, strat)

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

	logging.LogSystemAction(
		fmt.Sprintf("Added transformer at address: %s", t.Address),
	)
}

func (ch *channel) removeTransformer(t transformer) {
	ch.mutex.Lock()

	for i, transformer := range ch.transformers {
		if t.Address == transformer.Address {
			ch.transformers = append(
				ch.transformers[:i], ch.transformers[i+1:]...,
			)
			break
		}
	}

	snap := append([]transformer(nil), ch.transformers...)
	ch.mutex.Unlock()
	ch.tSnap.Store(snap)

	logging.LogSystemAction(
		fmt.Sprintf("Removed transformer for address: %s", t.Address),
	)
	ch.checkEmptyChannel()
}

// Atomicly load the transformers list without having to lock the mutex.
func (ch *channel) loadTransformers() []transformer {
	if v := ch.tSnap.Load(); v != nil {
		return v.([]transformer)
	}
	return nil
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

	logging.LogSystemAction(
		fmt.Sprintf("Added subscriber at address: %s", s.Address),
	)
}

func (ch *channel) removeSubscriber(s subscriber) {
	ch.mutex.Lock()

	for i, subscriber := range ch.subscribers {
		if s.Address == subscriber.Address {
			ch.subscribers = append(ch.subscribers[:i], ch.subscribers[i+1:]...)
			break
		}
	}

	snap := append([]subscriber(nil), ch.subscribers...)
	ch.mutex.Unlock()
	ch.sSnap.Store(snap)

	logging.LogSystemAction(
		fmt.Sprintf("Removed subscriber for address: %s", s.Address),
	)
	ch.checkEmptyChannel()
}

// Atomicly load the subscriber list without having to lock the mutex.
func (ch *channel) loadSubscribers() []subscriber {
	if v := ch.sSnap.Load(); v != nil {
		return v.([]subscriber)
	}
	return nil
}

// Get the selected subscribers to forward a message to using channel's
// selection strategy.
func (ch *channel) selectSubscribers() []subscriber {
	return ch.selector.Select()
}

func (ch *channel) checkEmptyChannel() {
	if !globals.AutoConsolidate {
		return
	}

	subs := ch.loadSubscribers()
	trans := ch.loadTransformers()
	if len(subs) == 0 && len(trans) == 0 {
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

func (c *channel) enqueue(m *protocol.Object) {
	c.mutex.RLock()
	parts := c.partitions
	c.mutex.RUnlock()

	if len(parts) == 0 {
		return // Channel is closed / removed
	}

	idx := int(c.hash([]byte(m.Arg3))) % len(parts)
	parts[idx].in <- m
}
