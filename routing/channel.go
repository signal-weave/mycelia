package routing

import (
	"fmt"
	"mycelia/commands"
	"mycelia/utils"
)

func NewChannel(name string) *Channel {
	channel := Channel{
		Name:        name,
		subscribers: []Consumer{},
	}
	return &channel
}

// The main route a message travels through. Channels
type Channel struct {
	Name        string
	subscribers []Consumer
}

// Stores the consumer as a subscriber of the channel and will forward all
// processed messages to the consumer.
func (c *Channel) RegisterSubscriber(subscriber *Consumer) {
	// Temp setup of single array of subscribers.
	c.subscribers = append(c.subscribers, *subscriber)
	aMsg := fmt.Sprintf("Added Subscriber %s", subscriber.Address)
	utils.ActionPrint(aMsg)
}

func (c *Channel) ProcessMessage(m *commands.SendMessage) *commands.SendMessage {
	// Temp fan out broadcasting - Will be packaged into unique channel later.
	for _, consumer := range c.subscribers {
		consumer.ConsumeMessage(m)
	}

	// Returning the message is for when message transformers are supported.
	return m
}
