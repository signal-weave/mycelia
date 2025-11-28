package routing

import (
	"fmt"
	"mycelia/logging"
	"sync"

	"mycelia/globals"

	"github.com/signal-weave/rhizome"
)

var partLoadTransformers = func(p *partition) []transformer {
	return p.channel.loadTransformers()
}
var partSelectSubscribers = func(p *partition) []subscriber {
	return p.channel.selectSubscribers()
}
var partApply = func(t transformer, o *rhizome.Object) (*rhizome.Object, error) {
	return t.apply(o)
}
var partDeliver = func(s subscriber, o *rhizome.Object) {
	s.deliver(o)
}
var partNext = func(p *partition) *channel {
	return p.route.getNextChannel(p.channel)
}
var partEnqueue = func(ch *channel, o *rhizome.Object) {
	ch.enqueue(o)
}

// Acknowledge the message to the sender
var partAck = func(obj *rhizome.Object) error {
	obj.Response.Ack = globals.AckSent
	payload, err := rhizome.EncodeResponse(obj)
	if err != nil {
		return err
	}
	return obj.Responder.Write(payload)
}

// A worker that manages the communication between transformers and subscribers
// on a channel. These offload the communication so the channel can keep
// funneling messages to its partition workers.
type partition struct {
	route   *route
	channel *channel
	in      chan *rhizome.Object
	wg      sync.WaitGroup
}

func newPartition(r *route, c *channel) *partition {
	return &partition{
		route:   r,
		channel: c,
	}
}

func (p *partition) start() { p.wg.Add(1); go p.loop() }
func (p *partition) stop()  { close(p.in); p.wg.Wait() }

// Should be called as a go routine so the partition worker is always working.
// It can be fed messages through partition.in which will be processed by the
// loop. Remember to call partition.stop() to close loop + channel and shutdown.
func (p *partition) loop() {
	defer p.wg.Done()
	for m := range p.in {
		if m == nil {
			continue
		}

		var err error

		ts := partLoadTransformers(p)
		result := m
		for _, t := range ts {
			result, err = partApply(t, result)
			if err != nil {
				continue
			}
		}
		if result == nil {
			continue
		}

		ss := partSelectSubscribers(p)
		var wg sync.WaitGroup
		wg.Add(len(ss))

		for _, sub := range ss {
			s := sub // capture - loops use pointers for tracking
			msg := result

			go func() {
				defer wg.Done()
				partDeliver(s, msg)
			}()
		}

		wg.Wait()

		// pass to the next channel
		if next := partNext(p); next != nil {
			partEnqueue(next, result)
		} else if result.AckPlcy == globals.AckPlcyOnsent {
			if err := partAck(result); err != nil {
				m := fmt.Sprintf("Unable to write to %s: %s", result.Responder.RemoteAddr(), err)
				logging.LogObjectWarning(m, result.UID)
			}
		}
	}
}
