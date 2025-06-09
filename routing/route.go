package routing

import (
	"errors"
	"fmt"
	"mycelia/commands"
	"mycelia/utils"
)

func NewRoute(name string) *Route {
	return &Route{
		Name:     name,
		Channels: []Channel{},
	}
}

// Routes are comprised of an array of channels that messages can be sent
// through. The router forwards messages to be sent down a route, sending the
// message through each channel.
type Route struct {
	Name     string
	Channels []Channel
}

// Look up channel by name if one can be found else return nil.
func (r *Route) GetChannel(name string) (*Channel, error) {
	for i, v := range r.Channels {
		if v.Name == name {
			return &r.Channels[i], nil
		}
	}
	return nil, errors.New("no channel found")
}

// Adds and returns a channel by the given command to the route.
// If the channel already exists on the route, the action is skipped and returns
// the existing channel instead.
func (r *Route) AddChannel(c *commands.AddChannel) *Channel {
	channel, err := r.GetChannel(c.Name)
	if err != nil {
		channel = NewChannel(c.Name)
		r.Channels = append(r.Channels, *channel)
		m := fmt.Sprintf("Adding channel: [%s] to route: [%s]", c.Name, r.Name)
		utils.SprintfLnIndent(m, 2)
	}

	return channel
}

// Adds a consumer to a channel from the AddSubscriber command data.
// Creates a new channel if the desired one cannot be found.
func (r *Route) AddSubscriber(s *commands.AddSubscriber) {
	channel, err := r.GetChannel(s.Channel)
	if err != nil {
		msg := "Error adding subscriber on route: [%s], channel not found: [%s]"
		utils.SprintfLnIndent(msg, 2, s.Route, s.Channel)
		return
	}

	consumer := NewConsumer(s.Address)
	channel.RegisterSubscriber(consumer)
}

// Sends the message down the route with each transformed message being passed
// on to the next channel.
func (r *Route) ProcessMessage(m *commands.SendMessage) {
	result := m
	result.Status = commands.StatusPending
	for _, channel := range r.Channels {
		result = channel.ProcessMessage(result)
	}
	result.Status = commands.StatusResolved
}
