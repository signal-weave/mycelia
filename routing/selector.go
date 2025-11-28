package routing

import (
	"math/rand"
	"sync"
	"time"

	"mycelia/globals"
)

// A selector is the channel component that chooses which subscribers a message
// should be sent to using various selection strategies: round-robin, random,
// pub/sub, etc.
type selector interface {
	Select() []subscriber
	GetStrategyName() string
}

func newSelector(ch *channel, strat globals.SelectionStrategy) selector {
	switch strat {

	case globals.SelStratRandom:
		return &randomSelector{
			strategy: strat,

			// Needed to seed randomizer
			rng: rand.New(rand.NewSource(time.Now().UnixNano())),

			loadFn:   ch.loadSubscribers,
			chooseFn: func(s []subscriber) (subscriber, bool) { return randomElement(s) },
		}

	case globals.SelStratRoundRobin:
		return &roundRobinSelector{
			strategy: strat,
			last:     -1,
			loadFn:   ch.loadSubscribers,
		}

	default:
		return &pubSubSelector{
			strategy: strat,
			loadFn:   ch.loadSubscribers,
		}
	}
}

// -------Random Selector-------------------------------------------------------

type randomSelector struct {
	strategy globals.SelectionStrategy
	rng      *rand.Rand
	loadFn   func() []subscriber
	chooseFn func([]subscriber) (subscriber, bool)
}

func (rs *randomSelector) GetStrategyName() string {
	return rs.strategy.String()
}

func (rs *randomSelector) Select() []subscriber {
	subscribers := rs.loadFn()
	if len(subscribers) == 0 {
		return []subscriber{}
	}

	chosen, found := rs.chooseFn(subscribers)
	if !found {
		return []subscriber{}
	}

	return []subscriber{chosen}
}

// -------Round-Robin Selector--------------------------------------------------

type roundRobinSelector struct {
	strategy globals.SelectionStrategy
	mu       sync.Mutex
	last     int
	loadFn   func() []subscriber
}

func (rrs *roundRobinSelector) GetStrategyName() string {
	return rrs.strategy.String()
}

func (rrs *roundRobinSelector) Select() []subscriber {
	subscribers := rrs.loadFn()
	if len(subscribers) == 0 {
		return nil
	}
	rrs.mu.Lock()

	if rrs.last < 0 || rrs.last >= len(subscribers)-1 {
		rrs.last = 0
	} else {
		rrs.last++
	}

	idx := rrs.last
	rrs.mu.Unlock()
	return []subscriber{subscribers[idx]}
}

// -------Pub/Sub Selector------------------------------------------------------

type pubSubSelector struct {
	strategy globals.SelectionStrategy
	loadFn   func() []subscriber
}

func (pss *pubSubSelector) GetStrategyName() string {
	return pss.strategy.String()
}

func (pss *pubSubSelector) Select() []subscriber {
	subscribers := pss.loadFn()
	if len(subscribers) == 0 {
		return []subscriber{}
	}
	return subscribers
}
