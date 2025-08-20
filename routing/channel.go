package routing

import (
	"fmt"
	"mycelia/commands"
	"mycelia/str"
)

func NewChannel(name string) *Channel {
	channel := Channel{
		Name:         name,
		Subscribers:  []*Consumer{},
		Transformers: []*Transformer{},
	}
	return &channel
}

// The main route a message travels through. Channels can have transformers
// that modify messages before they reach subscribers.
type Channel struct {
	Name         string
	Subscribers  []*Consumer
	Transformers []*Transformer
}

func (c *Channel) alreadySubscribed(subscriber *Consumer) bool {
	for _, v := range c.Subscribers {
		if v.Address == subscriber.Address {
			return true
		}
	}
	return false
}

func (c *Channel) alreadyHasTransformer(transformer *Transformer) bool {
	for _, v := range c.Transformers {
		if v.Address == transformer.Address {
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
	c.Subscribers = append(c.Subscribers, subscriber)
	aMsg := fmt.Sprintf("Added Subscriber %s", subscriber.Address)
	str.ActionPrint(aMsg)
}

// Adds a transformer to the channel. Transformers are sorted by order.
func (c *Channel) RegisterTransformer(transformer *Transformer) {
	if c.alreadyHasTransformer(transformer) {
		return
	}

	c.Transformers = append(c.Transformers, transformer)
	aMsg := fmt.Sprintf("Added Transformer %s", transformer.Address)
	str.ActionPrint(aMsg)
}

func (c *Channel) ProcessMessage(m *commands.SendMessage) *commands.SendMessage {
	result := m

	// First, run message through all transformers in order
	for _, transformer := range c.Transformers {
		transformedMsg, err := transformer.TransformMessage(result)
		if err != nil {
			// Log error but continue with original message
			eMsg := fmt.Sprintf(
				"Transformer %s failed: %v", transformer.Address, err)
			str.ErrorPrint(eMsg)
			continue
		}
		result = transformedMsg
	}

	// Then send transformed message to all subscribers
	for _, consumer := range c.Subscribers {
		consumer.ConsumeMessage(result)
	}

	return result
}
