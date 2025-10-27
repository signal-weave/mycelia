package routing

import (
	"fmt"
	"mycelia/logging"
	"sync"

	"mycelia/globals"

	"github.com/signal-weave/rhizome"
)

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

		ts := p.channel.loadTransformers()
		result := m
		for _, t := range ts {
			result, err = t.apply(result)
			if err != nil {
				continue
			}
		}
		if result == nil {
			continue
		}

		ss := p.channel.selectSubscribers()
		var wg sync.WaitGroup
		wg.Add(len(ss))

		for _, sub := range ss {
			s := sub // capture - loops use pointers for tracking
			msg := result

			go func() {
				defer wg.Done()
				s.deliver(msg)
			}()
		}

		wg.Wait()

		// pass to next channel
		if next := p.route.getNextChannel(p.channel); next != nil {
			next.enqueue(result)
		} else {
			// If no remaining channels, inform sender the message was sent.
			if result.AckPlcy == globals.AckPlcyOnsent {
				result.Response.Ack = globals.AckSent
				payload, err := rhizome.EncodeResponse(result)
				if err != nil {
					logging.LogSystemError(
						fmt.Sprintf("could not encode msg from %s", result.Responder.RemoteAddr()),
					)
				}
				err = result.Responder.Write(payload)
				if err != nil {
					m := fmt.Sprintf("Unable to write to %s: %s", result.Responder.RemoteAddr(), err)
					logging.LogObjectWarning(m, result.UID)
				}
			}
		}
	}
}
