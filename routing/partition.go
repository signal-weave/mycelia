package routing

import (
	"mycelia/protocol"
	"sync"
)

// A worker that manages the communication between transformers and subscribers
// on a channel. These offload the communication so the channel can keep
// funneling messages to its partition workers.
type partition struct {
	route   *route
	channel *channel
	in      chan *protocol.Command
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
// It can be fed messages through partition.in which will get processed in the
// loop. Remember to call partition.stop() to close loop + channel and shutdown.
func (p *partition) loop() {
	defer p.wg.Done()
	for m := range p.in {
		if m == nil {
			continue
		}

		var err error

		ts := p.channel.tSnap.Load().([]transformer)
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

		ss := p.channel.sSnap.Load().([]subscriber)
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
		}
	}
}
