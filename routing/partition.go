package routing

import (
	"mycelia/protocol"
	"sync"
	"sync/atomic"
)

type partition struct {
	in           chan *protocol.Command
	transformers *atomic.Pointer[[]transformer]
	subscribers  *atomic.Pointer[[]subscriber]
	wg           sync.WaitGroup
}

func (p *partition) start() { p.wg.Add(1); go p.loop() }
func (p *partition) stop()  { close(p.in); p.wg.Wait() }

func (p *partition) loop() {
	defer p.wg.Done()
	for m := range p.in {
		if m == nil {
			continue
		}

		ts := *p.transformers.Load()
		ss := *p.subscribers.Load()
		var err error

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
	}
}
