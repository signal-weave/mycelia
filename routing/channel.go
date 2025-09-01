package routing

import (
	"fmt"
	"hash/fnv"
	"slices"
	"sync"

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
	partitions   []*partition
}

func newChannel(r *route, name string, numPartitions int) *channel {
	partitions := make([]*partition, numPartitions)

	hash := func(b []byte) uint32 {
		h := fnv.New32a()
		_, _ = h.Write(b)
		return h.Sum32()
	}

	return &channel{
		route: r,
		name:  name,

		hash: hash,

		transformers: []transformer{},
		subscribers:  []subscriber{},
		partitions:   partitions,
	}
}

func (ch *channel) addTransformer(t transformer) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	for _, existing := range ch.transformers {
		if existing.Address == t.Address {
			return
		}
	}
	ch.transformers = append(ch.transformers, t)
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
	defer ch.mutex.Unlock()
	for _, existing := range ch.subscribers {
		if existing.Address == s.Address {
			return
		}
	}
	ch.subscribers = append(ch.subscribers, s)
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
		ch.route.removeChannel(ch.name)
	}
}

func (c *channel) enqueue(m *protocol.Command) {
	key := []byte(m.Arg3) // address on messages
	idx := int(c.hash(key)) % len(c.partitions)
	c.partitions[idx].in <- m
}

func (c *channel) deliver(m *protocol.Command) *protocol.Command {
	result := m

	c.mutex.RLock() // Copy transform slice for minimal mutex lock time
	transformers := slices.Clone(c.transformers)
	c.mutex.RUnlock()

	// First, run delivery through all transformers in order
	for _, transformer := range transformers {
		transformedMsg, err := transformer.apply(result)
		if err != nil {
			continue
		}
		result = transformedMsg
	}

	c.mutex.RLock() // Copy subscriber slice for minimal mutex lock time
	subscribers := slices.Clone(c.subscribers)
	c.mutex.RUnlock()

	// Second, run transformed delivery through all subscribers.
	// --- fan-out delivery ---
	var wg sync.WaitGroup
	wg.Add(len(subscribers))

	for _, sub := range subscribers {
		s := sub // capture cause loops use pointers for tracking
		msg := result

		go func() {
			defer wg.Done()
			s.deliver(msg)
		}()
	}

	wg.Wait()
	return result
}
