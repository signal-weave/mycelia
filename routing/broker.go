package routing

import (
	"fmt"
	"sync"

	"mycelia/errgo"
	"mycelia/globals"
	"mycelia/protocol"
	"mycelia/str"
	"mycelia/system/cache"
)

// This is here so the server that spawns the broker can add itself without
// causing a circular dependency.
type server interface {
	UpdateListener()
	GetAddress() string
	GetPort() int
	Shutdown()
}

// -------Base Broker Details---------------------------------------------------

// The primary route orchestrator.
// Takes the incoming byte stream and runs it through the command parser where
// a generated command object is created and then runs the command through the
// route structure.
type Broker struct {
	ManagingServer server
	mutex          sync.RWMutex
	routes         map[string]*route
}

func NewBroker(s server) *Broker {
	return &Broker{
		ManagingServer: s,
		routes:         map[string]*route{},
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

// -------Route Management------------------------------------------------------

// Route returns existing or creates if missing.
func (b *Broker) Route(name string) *route {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	r, ok := b.routes[name]
	if !ok {
		r = &route{
			broker:   b,
			name:     name,
			channels: []*channel{},
		}
		b.routes[name] = r
		str.ActionPrint(
			fmt.Sprintf("Created route: %s", name),
		)
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

// -------Command Handling------------------------------------------------------

// Handles the command object generated from the incoming byte stream.
// Is exported for boot to load PreInit.json structures into.
func (b *Broker) HandleCommand(cmd *protocol.Command) error {
	switch cmd.ObjType {
	case globals.OBJ_DELIVERY:
		b.handleDelivery(cmd)
	case globals.OBJ_TRANSFORMER:
		b.handleTransformer(cmd)
	case globals.OBJ_SUBSCRIBER:
		b.handleSubscriber(cmd)
	case globals.OBJ_GLOBALS:
		b.handleGlobals(cmd)
	case globals.OBJ_Action:
		b.handleActions(cmd)
	default:
		wErr := errgo.NewError("Unknown object type!", globals.VERB_WRN)
		return wErr
	}

	return nil
}

func (b *Broker) handleDelivery(cmd *protocol.Command) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	switch cmd.CmdType {
	case globals.CMD_SEND:
		b.Route(cmd.Arg1).deliver(cmd)
	default:
		str.WarningPrint(
			fmt.Sprintf("Unknown command type for delivery from %s",
				cmd.ReturnAdress,
			),
		)
		return
	}
}

func (b *Broker) handleTransformer(cmd *protocol.Command) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	switch cmd.CmdType {

	case globals.CMD_ADD:
		t := newTransformer(cmd.Arg3)
		b.Route(cmd.Arg1).channel(cmd.Arg2).addTransformer(*t)
		cache.BrokerShape.Route(cmd.Arg1).Channel(cmd.Arg2).AddTransformer(t.Address)

	case globals.CMD_REMOVE:
		t := newTransformer(cmd.Arg3)
		b.Route(cmd.Arg1).channel(cmd.Arg2).removeTransformer(*t)
		cache.BrokerShape.Route(cmd.Arg1).Channel(cmd.Arg2).RemoveTransformer(t.Address)

	default:
		str.WarningPrint(
			fmt.Sprintf("Unknown command type for transformer from %s",
				cmd.ReturnAdress,
			),
		)
		return
	}

	cache.WriteSnapshot()
	b.PrintBrokerStructure()
}

func (b *Broker) handleSubscriber(cmd *protocol.Command) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	switch cmd.CmdType {

	case globals.CMD_ADD:
		// Args: route, channel, address, nil
		s := newSubscriber(cmd.Arg3)
		b.Route(cmd.Arg1).channel(cmd.Arg2).addSubscriber(*s)
		cache.BrokerShape.Route(cmd.Arg1).Channel(cmd.Arg2).AddSubscriber(s.Address)

	case globals.CMD_REMOVE:
		// Args: route, channel, address, nil
		s := newSubscriber(cmd.Arg3)
		b.Route(cmd.Arg1).channel(cmd.Arg2).removeSubscriber(*s)
		cache.BrokerShape.Route(cmd.Arg1).Channel(cmd.Arg2).RemoveSubscriber(s.Address)

	default:
		str.WarningPrint(
			fmt.Sprintf("Unknown command type for subscriber from %s",
				cmd.ReturnAdress,
			),
		)
		return
	}

	cache.WriteSnapshot()
	b.PrintBrokerStructure()
}

func (b *Broker) handleGlobals(cmd *protocol.Command) {
	switch cmd.CmdType {
	case globals.CMD_UPDATE:
		hasPermission := updateGlobals(cmd)
		if !hasPermission {
			return
		}
		if b.ManagingServer.GetAddress() != globals.Address ||
			b.ManagingServer.GetPort() != globals.Port {
			b.ManagingServer.UpdateListener()
		}
		cache.WriteSnapshot()
	default:
		str.WarningPrint(
			fmt.Sprintf("Unknown command type for globals from %s",
				cmd.ReturnAdress,
			),
		)
		return
	}
}

func (b *Broker) handleActions(cmd *protocol.Command) {
	switch cmd.CmdType {
	case globals.CMD_SIGTERM:
		b.ManagingServer.Shutdown()
	default:
		str.WarningPrint(
			fmt.Sprintf("Unknown command type for action from %s",
				cmd.ReturnAdress,
			),
		)
		return
	}
}

// -------Broker Util-----------------------------------------------------------

// PrintBrokerStructure prints the broker, routes, channels, and subscribers.
func (b *Broker) PrintBrokerStructure() {
	if !globals.PrintTree {
		return
	}

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	routeExpr := "  | - [route] %s\n"
	channelExpr := "        | - [channel] %s\n"
	transformerExpr := "              | - [transformer] %s\n"
	subscriberExpr := "              | - [subscriber] %s\n"

	str.PrintCenteredHeader("Broker Shape")
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

	str.PrintAsciiLine()
	fmt.Println()
}
