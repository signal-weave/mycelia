package routing

import (
	"mycelia/protocol"
	"sync"
)

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
