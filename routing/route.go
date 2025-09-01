package routing

import (
	"fmt"
	"slices"
	"sync"

	"mycelia/globals"
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

// channel returns existing or creates if missing.
func (r *route) channel(name string) *channel {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	ch := r.getChannel(name)

	if ch == nil {
		ch = newChannel(
			r,
			name,
			globals.DefaultNumPartitions,
		)
		r.channels = append(r.channels, ch)
		str.ActionPrint(fmt.Sprintf("Created channel %s.%s", r.name, name))
	}

	return ch
}

// Checks if a channel exists on the route.
// Returns channel and index if it does, else nil and -1.
func (r *route) getChannel(name string) *channel {
	var ch *channel = nil

	for _, c := range r.channels {
		if c.name == name {
			ch = c
			break
		}
	}
	return ch
}

// Returns the channel in next sequential order after the given channel.
func (r *route) getNextChannel(ch *channel) *channel {
	r.mutex.RLock()
	channels := slices.Clone(r.channels) // For minimal mutex lock time.
	r.mutex.RUnlock()

	encountered := false // If given channel has been encountered
	for _, v := range channels {
		if encountered {
			return v
		}
		if v.name == ch.name {
			encountered = true
		}
	}

	return nil
}

// Removes channel by name from the route.
// Will tell broker to remove channel if channel is empty and globals auto
// consolidate is true.
func (r *route) removeChannel(name string) {
	r.mutex.RLock()

	var ch *channel
	idx := -1

	for i, c := range r.channels {
		if c.name == name {
			ch = c
			idx = i
			break
		}
	}
	r.mutex.RUnlock()

	if ch == nil {
		return
	}

	r.mutex.Lock()

	r.channels = removeAt(r.channels, idx)
	if len(r.channels) == 0 {
		r.mutex.Unlock()
		r.broker.removeEmptyRoute(r.name)
		return
	}
	r.mutex.Unlock()
}

// Sends the message down the channel. The channel's partition will send it to
// the next channel so routes are only concerned about sending to the first
// channel.
func (r *route) enqueue(msg *protocol.Command) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if len(r.channels) == 0 {
		return
	}
	r.channels[0].enqueue(msg)
}
