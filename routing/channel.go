package routing

import (
	"fmt"
	"mycelia/commands"
	"mycelia/str"
)

func NewChannel(name string) *Channel {
	channel := Channel{
		Name:        name,
		Subscribers: []Consumer{},
	}
	return &channel
}

// The main route a message travels through. Channels
type Channel struct {
	Name        string
	Subscribers []Consumer
}

func (c *Channel) alreadySubscribed(subscriber *Consumer) bool {
	for _, v := range c.Subscribers {
		if v.Address == subscriber.Address {
			return true
		}
	}
	return false
}

// Stores the consumer as a subscriber of the channel and will forward all
// processed messages to the consumer.
func (c *Channel) RegisterSubscriber(subscriber *Consumer) {
	if c.alreadySubscribed(subscriber) {
		return
	}

	// Temp setup of single array of subscribers.
	c.Subscribers = append(c.Subscribers, *subscriber)
	aMsg := fmt.Sprintf("Added Subscriber %s", subscriber.Address)
	str.ActionPrint(aMsg)
}

func (c *Channel) ProcessMessage(m *commands.SendMessage) *commands.SendMessage {
	// Temp fan out broadcasting - Will be packaged into unique channel later.
	for _, consumer := range c.Subscribers {
		consumer.ConsumeMessage(m)
	}

	// Returning the message is for when message transformers are supported.
	return m
}
