package routing

import (
	"errors"
	"fmt"
	"net"
	"slices"
	"sync"
	"time"

	"mycelia/boot"
	"mycelia/commands"
	"mycelia/parsing"
	"mycelia/str"
)

// -----------------------------------------------------------------------------
// Herein are the broker components for routing messages.
// This package contains the broker, routes, channels, transformers and
// subscribers.

// See project readme for greater detail of component interactions.
// -----------------------------------------------------------------------------

// -------Transformer-----------------------------------------------------------

// Transformer intercepts messages, processes them, and returns modified
// messages.
type Transformer struct {
	Address string
}

func NewTransformer(address string) *Transformer {
	return &Transformer{
		Address: address,
	}
}

// TransformMessage sends the message to the transformer service and waits for
// response.
func (t *Transformer) transformMessage(m *commands.SendMessage) (*commands.SendMessage, error) {
	actionMsg := fmt.Sprintf("Transforming message via %s", t.Address)
	str.ActionPrint(actionMsg)

	conn, err := net.Dial("tcp", t.Address)
	if err != nil {
		wMsg := fmt.Sprintf("Could not dial transformer %s", t.Address)
		str.WarningPrint(wMsg)
		return m, err // Return original message on failure
	}
	defer conn.Close()

	// Send the message body to transformer
	_, err = conn.Write([]byte(m.Body))
	if err != nil {
		eMsg := fmt.Sprintf("Could not send data to transformer %s", t.Address)
		str.ErrorPrint(eMsg)
		return m, err
	}

	// Read the transformed response with a timeout
	conn.SetReadDeadline(time.Now().Add(
		time.Duration(boot.RuntimeCfg.TransformTimeout) * time.Second))

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		eMsg := fmt.Sprintf("Error reading from transformer %s", t.Address)
		str.ErrorPrint(eMsg)
		return m, err
	}

	// Create new message with transformed body
	transformedMessage := &commands.SendMessage{
		ID:     m.ID,
		Route:  m.Route,
		Status: m.Status,
		Body:   string(buffer[:n]),
	}

	return transformedMessage, nil
}

// -------Subscriber------------------------------------------------------------

// Object representing the client subscribed to an endpoint.
type Subscriber struct {
	Address string
}

func NewSubscriber(address string) *Subscriber {
	return &Subscriber{Address: address}
}

// Forwards the message to the client represented by the consumer object.
func (c *Subscriber) ConsumeMessage(m *commands.SendMessage) {
	fmt.Println("Attempting to dial", c.Address)
	conn, err := net.Dial("tcp", c.Address)
	if err != nil {
		wMsg := fmt.Sprintf("Could not dial %s", c.Address)
		str.WarningPrint(wMsg)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(m.Body))
	if err != nil {
		eMsg := fmt.Sprintf("Error sending to %s", c.Address)
		str.ErrorPrint(eMsg)
		return
	}
	m.Status = commands.StatusResolved
}

// -------Broker----------------------------------------------------------------

type Broker struct {
	mutex  sync.RWMutex
	routes map[string]*Route
}

func NewBroker() *Broker {
	return &Broker{
		routes: map[string]*Route{},
	}
}

// Route returns existing or creates if missing.
func (b *Broker) Route(name string) *Route {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	r, ok := b.routes[name]
	if !ok {
		r = &Route{
			name:     name,
			channels: map[string]*Channel{},
		}
		b.routes[name] = r
	}
	return r
}

// Handles the raw byte form of a command, hot off a socket, converts it to a
// command object, and forwards it to the command handler.
func (b *Broker) HandleBytes(input []byte) {
	// Parse byte stream -> command object.
	cmdType, cmd := parsing.ParseLine(input)
	if cmd == nil || cmdType == "err" {
		wMsg := "Error parsing command..."
		str.WarningPrint(wMsg)
		return
	}

	// Handle command object
	err := b.HandleCommand(cmd)
	if err != nil {
		msg := fmt.Sprintf("Unknown command type: %s", cmdType)
		str.WarningPrint(msg)
	}
}

func (b *Broker) HandleCommand(cmd commands.Command) error {
	switch t := cmd.(type) {
	case *commands.SendMessage:
		b.Route(t.Route).ProcessMessage(t)
	case *commands.AddSubscriber:
		subscriber := NewSubscriber(t.Address)
		b.Route(t.Route).Channel(t.Channel).AddSubscriber(*subscriber)
	case *commands.AddTransformer:
		transformer := NewTransformer(t.Address)
		b.Route(t.Route).Channel(t.Channel).AddTransformer(*transformer)
	default:
		return errors.New("Unknown command type")
	}

	b.printBrokerStructure()
	return nil
}

// printBrokerStructure prints the broker, routes, channels, and subscribers.
func (b *Broker) printBrokerStructure() {
	if !boot.RuntimeCfg.PrintTree {
		return
	}

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	routeExpr := "  | - [route] %s\n"
	channelExpr := "        | - [channel] %s\n"
	transformerExpr := "              | - [transformer] %s\n"
	subscriberExpr := "              | - [subscriber] %s\n"

	fmt.Println("\n[broker]")
	for routeName, route := range b.routes {
		fmt.Printf(routeExpr, routeName)
		for _, channel := range route.channels {
			fmt.Printf(channelExpr, channel.name)

			// Print transformers first
			for _, transformer := range channel.transformers {
				fmt.Printf(transformerExpr, transformer.Address)
			}

			// Then print subscribers
			for _, subscriber := range channel.subscribers {
				fmt.Printf(subscriberExpr, subscriber.Address)
			}
		}
	}

	fmt.Println()
}

// -------Route-----------------------------------------------------------------

type Route struct {
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
			name:         name,
			transformers: []Transformer{},
			subscribers:  []Subscriber{},
		}
		r.channels[name] = ch
	}

	return ch
}

// Sends the message down the route with each transformed message being passed
// on to the next channel.
func (r *Route) ProcessMessage(sm *commands.SendMessage) {
	r.mutex.RLock()
	// copy map for minimal mutex lock time
	channels := make([]*Channel, 0, len(r.channels))
	for _, c := range r.channels {
		channels = append(channels, c)
	}
	r.mutex.RUnlock()

	cur := sm
	for _, ch := range channels {
		cur = ch.ProcessMessage(cur)
	}
}

// -------Channel---------------------------------------------------------------

type Channel struct {
	mutex        sync.RWMutex
	name         string
	transformers []Transformer
	subscribers  []Subscriber
}

func (ch *Channel) AddTransformer(t Transformer) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	ch.transformers = append(ch.transformers, t)
}

func (ch *Channel) AddSubscriber(s Subscriber) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	ch.subscribers = append(ch.subscribers, s)
}

func (c *Channel) ProcessMessage(m *commands.SendMessage) *commands.SendMessage {
	result := m

	c.mutex.RLock() // Copy transform slice for minimal mutex lock time
	transformers := slices.Clone(c.transformers)
	c.mutex.RUnlock()

	// First, run message through all transformers in order
	for _, transformer := range transformers {
		transformedMsg, err := transformer.transformMessage(result)
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
	subscribers := slices.Clone(c.subscribers)
	c.mutex.RUnlock()

	// Second, run transformed message through all subscribers.
	// --- fan-out delivery ---
	var wg sync.WaitGroup
	wg.Add(len(subscribers))

	for _, sub := range subscribers {
		s := sub // capture cause loops use pointers for tracking
		msg := result

		go func() {
			defer wg.Done()
			s.ConsumeMessage(msg)
		}()
	}

	wg.Wait()
	return result
}
