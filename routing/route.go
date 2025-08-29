package routing

import (
	"mycelia/commands"
	"sync"
)

// The primary grouping for message traversal. A route can contain multiple
// channels. Subscribers subscribe to a channel, not a route. Because channels
// can contain transformers, routes are a way of grouping transformer
// "categories", in the form of channels, together.
//
// When a delivery is sent through a channel and possibly transformed, the newly
// transformed delivery is sent to the next channel in the route.
type Route struct {
	broker   *Broker
	mutex    sync.RWMutex
	name     string
	channels map[string]*Channel
}

// Channel returns existing or creates if missing
func (r *Route) Channel(name string) *Channel {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	ch, exists := r.channels[name]
	if !exists {
		ch = &Channel{
			route:        r,
			name:         name,
			transformers: []Transformer{},
			subscribers:  []Subscriber{},
		}
		r.channels[name] = ch
	}

	return ch
}

func (r *Route) removeChannel(name string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, exists := r.channels[name]
	if !exists {
		return
	}
	delete(r.channels, name)

	if len(r.channels) == 0 {
		r.broker.removeEmptyRoute(r.name)
	}
}

// Sends the delivery down the route with each transformed delivery being passed
// on to the next channel.
func (r *Route) ProcessDelivery(sm *commands.Delivery) {
	r.mutex.RLock()
	// copy map for minimal mutex lock time
	channels := make([]*Channel, 0, len(r.channels))
	for _, c := range r.channels {
		channels = append(channels, c)
	}
	r.mutex.RUnlock()

	cur := sm
	for _, ch := range channels {
		cur = ch.ProcessDelivery(cur)
	}
}
