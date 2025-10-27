package routing

import (
	"fmt"
	"slices"
	"strconv"
	"sync"

	"mycelia/globals"
	"mycelia/logging"

	"github.com/signal-weave/rhizome"
)

// The primary grouping for message traversal. A route can contain multiple
// channels. Subscribers subscribe to a channel, not a route. Because channels
// can contain transformers, routes are a way of grouping transformer
// "categories", in the form of channels, together.
//
// When a delivery is sent through a channel and possibly transformed, the newly
// transformed delivery is sent to the next channel in the route.
type route struct {
	broker     *Broker
	mutex      sync.RWMutex
	name       string
	channels   []*channel
	deadLetter *channel
}

func newRoute(broker *Broker, name string) *route {
	var channels []*channel

	r := route{
		broker:   broker,
		name:     name,
		channels: channels,
	}

	// Create a channel specifically for undelivered/missrouted/errored messages
	// for cleanup later. Dead-letters have 2 partitions, as there should be far
	// less bad messages going through the route than good messages.
	deadLetter := newChannel(&r, globals.DeadLetter, 2, globals.SelStratPubSub)
	r.deadLetter = deadLetter

	return &r
}

// Checks if a channel exists on the route.
// Returns channel if found else nil.
func (r *route) getChannel(name string) *channel {
	var ch *channel = nil

	r.mutex.RLock()
	for _, c := range r.channels {
		if c.name == name {
			ch = c
			break
		}
	}
	r.mutex.RUnlock()

	return ch
}

// Creates a new channel and adds it to the route from the given obj.
// Arg2 is the channel name, arg3 is the selection strategy.
//
// If the channel already exists, a response is sent with an ack value of
// globals.ACK_TYPE_CHANNEL_ALREADY_EXISTS.
func (r *route) createChannel(obj *rhizome.Object) {
	// Args: route, name, strategy, nil
	ch := r.getChannel(obj.Arg2)
	if ch != nil {
		if obj.Responder != nil {
			obj.Response.Ack = globals.AckChannelAlreadyExists
			payload, err := rhizome.EncodeResponse(obj)
			if err != nil {
				logging.LogSystemError(
					fmt.Sprintf("could not create channel from %s", obj.Responder.RemoteAddr()),
				)
			}
			err = obj.Responder.Write(payload)
			if err != nil {
				m := fmt.Sprintf("Could not write ack for %s: %s", obj.Responder.RemoteAddr(), err)
				logging.LogObjectWarning(m, obj.UID)
			}
		}
		return
	}

	i, err := strconv.Atoi(obj.Arg3)
	if err != nil {
		logging.LogObjectWarning(
			fmt.Sprintf(
				"Unable to parse new channel selection strategy from %s",
				obj.Responder.RemoteAddr(),
			), obj.UID,
		)
	}
	strat := globals.SelectionStrategy(i)

	ch = newChannel(r, obj.Arg2, globals.DefaultNumPartitions, strat)
	r.mutex.Lock()
	r.channels = append(r.channels, ch)
	r.mutex.Unlock()
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
func (r *route) enqueue(msg *rhizome.Object) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if len(r.channels) == 0 {
		r.deadLetter.enqueue(msg)
		return
	}
	r.channels[0].enqueue(msg)
}
