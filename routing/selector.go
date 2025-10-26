package routing

import (
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
		}

	case globals.SelStratRoundRobin:
		return &roundRobinSelector{
			strategy: strat,
			channel:  ch,
		}

	case globals.SelStratPubSub:
		return &pubSubSelector{
			strategy: strat,
			channel:  ch,
		}

	default:
		return nil
	}
}

// -------Random Selector-------------------------------------------------------

type randomSelector struct {
	strategy globals.SelectionStrategy
	channel  *channel
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
	strategy  globals.SelectionStrategy
	channel   *channel
	lastIndex int
}

func (rrs *roundRobinSelector) GetStrategyName() string {
	return rrs.strategy.String()
}

func (rrs *roundRobinSelector) Select() []subscriber {
	subscribers := rrs.channel.loadSubscribers()
	if len(subscribers) == 0 {
		return []subscriber{}
	}

	nextIndex := rrs.lastIndex + 1
	if nextIndex == len(subscribers) {
		nextIndex = 0
	}
	rrs.lastIndex = nextIndex

	chosen := subscribers[nextIndex]
	return []subscriber{chosen}
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
