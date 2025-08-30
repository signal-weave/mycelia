package routing

import (
	"encoding/json"
	"fmt"
	"mycelia/errgo"
	"mycelia/globals"
	"mycelia/protocol"
	"mycelia/str"
	"sync"
	"time"
)

// This is here so the server that spawns the broker can add itself without
// causing a circular dependency.
type server interface {
	Run()
	UpdateListener()
	GetAddress() string
	GetPort() int
}

// -------Base Broker Details---------------------------------------------------

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
	default:
		wErr := errgo.NewError("Unknown command type!", globals.VERB_WRN)
		return wErr
	}

	return nil
}

func (b *Broker) handleDelivery(cmd *protocol.Command) {
	switch cmd.CmdType {
	case globals.CMD_SEND:
		b.Route(cmd.Arg1).ProcessDelivery(cmd)
	}
}

func (b *Broker) handleTransformer(cmd *protocol.Command) {
	switch cmd.CmdType {
	case globals.CMD_ADD:
		transformer := NewTransformer(cmd.Arg3)
		b.Route(cmd.Arg1).Channel(cmd.Arg2).AddTransformer(*transformer)
	case globals.CMD_REMOVE:
		transformer := NewTransformer(cmd.Arg3)
		b.Route(cmd.Arg1).Channel(cmd.Arg2).RemoveTransformer(*transformer)
	}
	b.PrintBrokerStructure()
}

func (b *Broker) handleSubscriber(cmd *protocol.Command) {
	switch cmd.CmdType {
	case globals.CMD_ADD:
		// Args: route, channel, address, nil
		subscriber := NewSubscriber(cmd.Arg3)
		b.Route(cmd.Arg1).Channel(cmd.Arg2).AddSubscriber(*subscriber)
	case globals.CMD_REMOVE:
		// Args: route, channel, address, nil
		subscriber := NewSubscriber(cmd.Arg3)
		b.Route(cmd.Arg1).Channel(cmd.Arg2).RemoveSubscriber(*subscriber)
	}
	b.PrintBrokerStructure()
}

func (b *Broker) handleGlobals(cmd *protocol.Command) {
	switch cmd.CmdType {
	case globals.CMD_UPDATE:
		b.updateGlobals(cmd)
		if b.ManagingServer.GetAddress() != globals.Address ||
		b.ManagingServer.GetPort() != globals.Port {
			b.ManagingServer.UpdateListener()
		}
	}
}

func (b *Broker) updateGlobals(cmd *protocol.Command) {
	values := map[string]any{}
	err := json.Unmarshal(cmd.Payload, &values)
	if err != nil {
		wMsg := fmt.Sprintf(
			"Could not parse payload for globals update from %s",
			cmd.Sender,
		)
		errgo.NewError(wMsg, globals.VERB_WRN)
		return
	}

	address, exists := values["address"].(string)
	if exists {
		globals.Address = address
	}
	port, exists := values["port"].(float64)
	if exists {
		globals.Port = int(port)
	}
	verbosity, exists := values["verbosity"].(float64)
	if exists {
		globals.Verbosity = int(verbosity)
	}
	globals.UpdateVerbosityEnvironVar()
	printTree, exists := values["print_tree"].(bool)
	if exists {
		globals.PrintTree = printTree
	}
	timeout, exists := values["transform_timeout"].(string)
	if exists {
		newTimeout, err := time.ParseDuration(timeout)
		if err != nil {
			wMsg := fmt.Sprintf(
				"Unable to parse transform timeout expr from %s", cmd.Sender,
			)
			str.WarningPrint(wMsg)
		} else {
			globals.TransformTimeout = newTimeout
		}
	}
	globals.PrintDynamicValues()
}

// -------Util------------------------------------------------------------------

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
