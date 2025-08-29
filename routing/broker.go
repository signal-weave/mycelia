package routing

import (
	"fmt"
	"mycelia/commands"
	"mycelia/errgo"
	"mycelia/global"
	"mycelia/protocol"
	"sync"
)

type server interface {
	Run()
	UpdateListener()
	GetAddress() string
	GetPort() int
}

// The primary route orchestrator.
// Takes the incoming byte stream and runs it through the command parser where
// a generated command object is created and then runs the command through the
// route structure.
type Broker struct {
	ManagingServer server
	mutex          sync.RWMutex
	routes         map[string]*Route
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
			broker:   b,
			name:     name,
			channels: map[string]*Channel{},
		}
		b.routes[name] = r
	}
	return r
}

func (b *Broker) removeEmptyRoute(name string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	_, exists := b.routes[name]
	if exists {
		delete(b.routes, name)
	}
}

// Handles the raw byte form of a command, hot off a socket, converts it to a
// command object, and forwards it to the command handler.
func (b *Broker) HandleBytes(input []byte) {
	// Parse byte stream -> command object.
	cmd, err := protocol.ParseLine(input)
	if err != nil {
		return
	}

	// Handle command object
	b.HandleCommand(cmd)
}

// Handles the command object generated from the incoming byte stream.
// Is exported for boot to load PreInit.json structures into.
func (b *Broker) HandleCommand(cmd commands.Command) error {
	switch t := cmd.(type) {
	case *commands.Delivery:
		b.handleDelivery(t)
	case *commands.Transformer:
		b.handleTransformer(t)
	case *commands.Subscriber:
		b.handleSubscriber(t)
	case *commands.Globals:
		b.handleGlobals(t)
	default:
		wErr := errgo.NewError("Unknown command type!", global.VERB_WRN)
		return wErr
	}

	return nil
}

func (b *Broker) handleDelivery(cmd *commands.Delivery) {
	switch cmd.Cmd {
	case global.CMD_SEND:
		b.Route(cmd.Route).ProcessDelivery(cmd)
	}
}

func (b *Broker) handleTransformer(cmd *commands.Transformer) {
	switch cmd.Cmd {
	case global.CMD_ADD:
		transformer := NewTransformer(cmd.Address)
		b.Route(cmd.Route).Channel(cmd.Channel).AddTransformer(*transformer)
	case global.CMD_REMOVE:
		transformer := NewTransformer(cmd.Address)
		b.Route(cmd.Route).Channel(cmd.Channel).RemoveTransformer(*transformer)
	}
	b.PrintBrokerStructure()
}

func (b *Broker) handleSubscriber(cmd *commands.Subscriber) {
	switch cmd.Cmd {
	case global.CMD_ADD:
		subscriber := NewSubscriber(cmd.Address)
		b.Route(cmd.Route).Channel(cmd.Channel).AddSubscriber(*subscriber)
	case global.CMD_REMOVE:
		subscriber := NewSubscriber(cmd.Address)
		b.Route(cmd.Route).Channel(cmd.Channel).RemoveSubscriber(*subscriber)
	}
	b.PrintBrokerStructure()
}

func (b *Broker) handleGlobals(cmd *commands.Globals) {
	global.UpdateGlobalsByMessage(cmd)

	switch cmd.Cmd {
	case global.CMD_UPDATE:
		if b.ManagingServer.GetAddress() != global.Address ||
			b.ManagingServer.GetPort() != global.Port {
			b.ManagingServer.UpdateListener()
		}
	}
}

// PrintBrokerStructure prints the broker, routes, channels, and subscribers.
func (b *Broker) PrintBrokerStructure() {
	if !global.PrintTree {
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
