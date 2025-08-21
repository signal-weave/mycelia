package routing

import (
	"fmt"
	"sync"

	"mycelia/commands"
	"mycelia/str"
)

func NewRoute(name string) *Route {
	return &Route{
		Name:     name,
		Channels: []*Channel{},
	}
}

// Routes are comprised of an array of channels that messages can be sent
// through. The router forwards messages to be sent down a route, sending the
// message through each channel.
type Route struct {
	Name     string
	Channels []*Channel
	mutex    sync.RWMutex
}

// Look up channel by name if one can be found else return nil.
func (r *Route) GetChannel(name string) (*Channel, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, ch := range r.Channels {
		if ch.Name == name {
			return ch, true
		}
	}
	return nil, false
}

// Adds and returns a channel by the given command to the route.
// If the channel already exists on the route, the action is skipped and returns
// the existing channel instead.
func (r *Route) AddChannel(cmd *commands.AddChannel) {
	_, exists := r.GetChannel(cmd.Name)
	if exists {
		return
	}

	ch := NewChannel(cmd.Name)
	r.mutex.Lock()
	r.Channels = append(r.Channels, ch)
	r.mutex.Unlock()
}

// Adds a consumer to a channel from the AddSubscriber command data.
// Creates a new channel if the desired one cannot be found.
func (r *Route) AddSubscriber(s *commands.AddSubscriber) {
	channel, exists := r.GetChannel(s.Channel)
	if !exists {
		msg := "Error adding subscriber on route: [%s], channel not found: [%s]"
		eMsg := fmt.Sprintf(msg, s.Route, s.Channel)
		str.ErrorPrint(eMsg)
		return
	}

	consumer := NewConsumer(s.Address)
	channel.RegisterSubscriber(consumer)
}

// Sends the message down the route with each transformed message being passed
// on to the next channel.
func (r *Route) ProcessMessage(m *commands.SendMessage) {
	r.mutex.RLock()
	// copy slice for minimal mutex lock time
	channels := append([]*Channel(nil), r.Channels...)
	r.mutex.RUnlock()

	for _, ch := range channels {
		ch.ProcessMessage(m)
	}
}

// Adds a transformer to a channel from the AddTransformer command data.
func (r *Route) AddTransformer(t *commands.AddTransformer) {
	channel, exists := r.GetChannel(t.Channel)
	if !exists {
		msg := "Error adding transformer on route: [%s], channel not found: [%s]"
		eMsg := fmt.Sprintf(msg, t.Route, t.Channel)
		str.ErrorPrint(eMsg)

		return
	}

	transformer := NewTransformer(t.Address)
	channel.RegisterTransformer(transformer)
}
