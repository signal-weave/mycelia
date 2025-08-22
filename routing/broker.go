package routing

import (
	"fmt"
	"sync"
	"errors"

	"mycelia/boot"
	"mycelia/commands"
	"mycelia/parsing"
	"mycelia/str"
)

// Creates a new broker with registered commands. A new broker will always
// contain a route named 'main' that contains no channels.
func NewBroker() *Broker {
	broker := Broker{}
	broker.Routes = map[string]*Route{
		"main": NewRoute("main"),
	}

	return &broker
}

// The structure responsible for sending message through routes,
// applying crud operations to routes, nad adding consumers to routes as
// subscribers.
// Routes are comprised of sequences of channels.
type Broker struct {
	// Route names to list of channels the route is comprised of.
	// Mesasges will travel through all channels on the route.
	Routes map[string]*Route

	mutex sync.RWMutex
}

// Thread-safe route lookup
func (b *Broker) getRoute(name string) (*Route, bool) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	route, exists := b.Routes[name]
	return route, exists
}

// Thread-safe route addition
func (b *Broker) addRoute(name string, route *Route) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.Routes[name] = route
}

// Handles the raw byte form of a command, hot off a socket.
func (b *Broker) HandleBytes(input []byte) {
	cmdType, cmd := parsing.ParseData(input)
	if cmd == nil || cmdType == "err" {
		wMsg := "Error parsing command..."
		str.WarningPrint(wMsg)
		return
	}

	err := b.HandleCommand(cmd)
	if err != nil {
		msg := fmt.Sprintf("Unknown command type: %s", cmdType)
		str.WarningPrint(msg)
	}
}

// Handle the decoded command object.
func (b *Broker) HandleCommand(cmd commands.Command) error {
	switch t := cmd.(type) {
	case *commands.SendMessage:
		b.SendMessage(*t)
	case *commands.AddRoute:
		b.AddRoute(*t)
	case *commands.AddChannel:
		b.AddChannel(*t)
	case *commands.AddSubscriber:
		b.AddSubscriber(*t)
	case *commands.AddTransformer:
		b.AddTransformer(*t)
	default:
		return errors.New("Unknown command type")
	}
	return nil
}

// -------Message Handlers------------------------------------------------------

func (b *Broker) SendMessage(cmd commands.SendMessage) {
	route, exists := b.getRoute(cmd.Route)
	if !exists {
		wMsg := fmt.Sprintf("Route not found: %s", cmd.Route)
		str.WarningPrint(wMsg)
		return
	}
	route.ProcessMessage(&cmd)
}

func (b *Broker) AddRoute(cmd commands.AddRoute) {
	_, exists := b.getRoute(cmd.Name)
	if exists {
		wMsg := fmt.Sprintf("Route %s already exists.", cmd.Name)
		str.WarningPrint(wMsg)
		b.PrintBrokerStructure()
		return
	}

	route := NewRoute(cmd.Name)
	b.addRoute(cmd.Name, route)
	str.SprintfLn("Route %s registered!", cmd.Name)
}

func (b *Broker) AddChannel(cmd commands.AddChannel) {
	route, exists := b.getRoute(cmd.Route)
	if !exists {
		wMsg := fmt.Sprintf("Route not found %s", cmd.Route)
		str.WarningPrint(wMsg)
		return
	}
	route.AddChannel(&cmd)
	b.PrintBrokerStructure()
}

func (b *Broker) AddSubscriber(cmd commands.AddSubscriber) {
	route, exists := b.getRoute(cmd.Route)
	if !exists {
		wMsg := fmt.Sprintf("Route not found %s", cmd.Route)
		str.WarningPrint(wMsg)
		return
	}
	route.AddSubscriber(&cmd)
	b.PrintBrokerStructure()
}

func (b *Broker) AddTransformer(cmd commands.AddTransformer) {
	route, exists := b.getRoute(cmd.Route)
	if !exists {
		wMsg := fmt.Sprintf("Route not found %s", cmd.Route)
		str.WarningPrint(wMsg)
		return
	}
	route.AddTransformer(&cmd)
	b.PrintBrokerStructure()
}

// -------Debug-----------------------------------------------------------------

// PrintBrokerStructure prints the broker, routes, channels, and subscribers.
func (b *Broker) PrintBrokerStructure() {
	if !boot.RuntimeCfg.PrintTree {
		return
	}

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	routeExpr := "  | - [route] %s\n"
	channelExpr := "        | - [channel] %s\n"
	transformerExpr := "              | - [transformer] %s (order: %d)\n"
	subscriberExpr := "              | - [subscriber] %s\n"

	fmt.Println("\n[broker]")
	for routeName, route := range b.Routes {
		fmt.Printf(routeExpr, routeName)
		for _, channel := range route.Channels {
			fmt.Printf(channelExpr, channel.Name)

			// Print transformers first
			for _, transformer := range channel.Transformers {
				fmt.Printf(transformerExpr, transformer.Address)
			}

			// Then print subscribers
			for _, subscriber := range channel.Subscribers {
				fmt.Printf(subscriberExpr, subscriber.Address)
			}
		}
	}

	fmt.Println()
}
