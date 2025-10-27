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
			channel:  ch,
			rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
		}

	case globals.SelStratRoundRobin:
		return &roundRobinSelector{
			strategy: strat,
			channel:  ch,
			last:     -1,
		}

	case globals.SelStratPubSub:
		return &pubSubSelector{
			strategy: strat,
			channel:  ch,
		}

	default:
		return &pubSubSelector{
			strategy: strat,
			channel:  ch,
		}
	}
}

// -------Random Selector-------------------------------------------------------

type randomSelector struct {
	strategy globals.SelectionStrategy
	channel  *channel
	rng      *rand.Rand
}

func (rs *randomSelector) GetStrategyName() string {
	return rs.strategy.String()
}

func (rs *randomSelector) Select() []subscriber {
	subscribers := rs.channel.loadSubscribers()
	if len(subscribers) == 0 {
		return []subscriber{}
	}

	chosen, found := randomElement(subscribers)
	if !found {
		return []subscriber{}
	}

	return []subscriber{chosen}
}

// -------Round-Robin Selector--------------------------------------------------

type roundRobinSelector struct {
	strategy globals.SelectionStrategy
	channel  *channel
	mu       sync.Mutex
	last     int
}

func (rrs *roundRobinSelector) GetStrategyName() string {
	return rrs.strategy.String()
}

func (rrs *roundRobinSelector) Select() []subscriber {
	subscribers := rrs.channel.loadSubscribers()
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
	channel  *channel
}

func (pss *pubSubSelector) GetStrategyName() string {
	return pss.strategy.String()
}

func (pss *pubSubSelector) Select() []subscriber {
	subscribers := pss.channel.loadSubscribers()
	if len(subscribers) == 0 {
		return []subscriber{}
	}
	return subscribers
}
