package routing

import (
	"fmt"
	"sync"

	"mycelia/protocol"
	"mycelia/str"
)

// The primary grouping for message traversal. A route can contain multiple
// channels. Subscribers subscribe to a channel, not a route. Because channels
// can contain transformers, routes are a way of grouping transformer
// "categories", in the form of channels, together.
//
// When a delivery is sent through a channel and possibly transformed, the newly
// transformed delivery is sent to the next channel in the route.
type route struct {
	broker   *Broker
	mutex    sync.RWMutex
	name     string
	channels []*channel
}

// Checks if a channel exists on the route.
// Returns channel and index if it does, else nil and -1.
func (r *route) channelExists(name string) (*channel, int) {
	var ch *channel = nil
	idx := -1

	for i, c := range r.channels {
		if c.name == name {
			idx = i
			ch = c
			break
		}
	}
	return ch, idx
}

// channel returns existing or creates if missing
func (r *route) channel(name string) *channel {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	ch, idx := r.channelExists(name)

	if idx < 0 {
		ch = &channel{
			route:        r,
			name:         name,
			transformers: []transformer{},
			subscribers:  []subscriber{},
		}
		r.channels = append(r.channels, ch)
		str.ActionPrint(
			fmt.Sprintf("Created channel %s.%s", r.name, name),
		)
	}

	return ch
}

func (r *route) removeChannel(name string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, idx := r.channelExists(name)
	if idx < 0 {
		return
	}
	removeAt(r.channels, idx)

	if len(r.channels) == 0 {
		r.broker.removeEmptyRoute(r.name)
	}
}

// Sends the delivery down the route with each transformed delivery being passed
// on to the next channel.
func (r *route) deliver(sm *protocol.Command) {
	r.mutex.RLock()
	// copy map for minimal mutex lock time
	channels := make([]*channel, 0, len(r.channels))
	for _, c := range r.channels {
		channels = append(channels, c)
	}
	r.mutex.RUnlock()

	cur := sm
	for _, ch := range channels {
		cur = ch.deliver(cur)
	}
}
