package routing

import (
	"fmt"
	"slices"
	"sync"

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
	mutex        sync.RWMutex
}

// Stores the consumer as a subscriber of the channel and will forward all
// processed messages to the consumer.
func (c *Channel) RegisterSubscriber(subscriber *Consumer) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// do the check and append atomically
	for _, v := range c.Subscribers {
		if v.Address == subscriber.Address {
			return
		}
	}

	c.Subscribers = append(c.Subscribers, subscriber)
	aMsg := fmt.Sprintf("Added Subscriber %s", subscriber.Address)
	str.ActionPrint(aMsg)
}

// Adds a transformer to the channel. Transformers are sorted by order.
func (c *Channel) RegisterTransformer(transformer *Transformer) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, v := range c.Transformers {
		if v.Address == transformer.Address {
			return
		}
	}

	c.Transformers = append(c.Transformers, transformer)
	aMsg := fmt.Sprintf("Added Transformer %s", transformer.Address)
	str.ActionPrint(aMsg)
}

func (c *Channel) ProcessMessage(m *commands.SendMessage) *commands.SendMessage {
	result := m

	c.mutex.RLock() // Copy transform slice for minimal mutex lock time
	transformers := slices.Clone(c.Transformers)
	c.mutex.RUnlock()

	// First, run message through all transformers in order
	for _, transformer := range transformers {
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

	c.mutex.RLock() // Copy subscriber slice for minimal mutex lock time
	subscribers := slices.Clone(c.Subscribers)
	c.mutex.RUnlock()

	// --- fan-out delivery ---
	var wg sync.WaitGroup
	wg.Add(len(subscribers))

	for _, consumer := range subscribers {
		cons := consumer // capture cause loops use pointers for tracking
		msg := result

		go func() {
			defer wg.Done()
			cons.ConsumeMessage(msg)
		}()
	}

	wg.Wait()
	return result
}
