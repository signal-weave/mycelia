package routing

import (
	"fmt"
	"sync"

	"mycelia/comm"
	"mycelia/errgo"
	"mycelia/globals"
	"mycelia/protocol"
	"mycelia/str"
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

// The primary object orchestrator.
// Takes the incoming byte stream and runs it through the object parser where
// a generated object is created and then runs the object through the route
// structure.
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

// Handles the raw byte form of a object, hot off a socket, converts it to an
// object, and forwards it to the object handler.
func (b *Broker) HandleBytes(input []byte, resp *comm.ConnResponder) {
	// Parse byte stream -> object.
	obj, err := protocol.DecodeFrame(input, resp)
	if err != nil {
		return
	}

	b.HandleObject(obj)
}

// -------Route Management------------------------------------------------------

// route returns existing or creates if missing.
func (b *Broker) route(name string) *route {
	b.mutex.RLock()
	r := b.routes[name]
	b.mutex.RUnlock()
	if r != nil {
		return r
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()
	if r = b.routes[name]; r == nil {
		r = &route{broker: b, name: name, channels: []*channel{}}
		b.routes[name] = r
		str.ActionPrint(fmt.Sprintf("Created route: %s", name))
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

// -------Object Handling------------------------------------------------------

// Handles the object generated from the incoming byte stream.
// Is exported for boot to load PreInit.json structures into.
func (b *Broker) HandleObject(cmd *protocol.Object) error {
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

func (b *Broker) handleDelivery(cmd *protocol.Object) {
	switch cmd.CmdType {

	case globals.CMD_SEND:
		b.route(cmd.Arg1).enqueue(cmd)

	default:
		str.WarningPrint(
			fmt.Sprintf("Unknown command type for delivery from %s",
				cmd.Responder.C.RemoteAddr().String(),
			),
		)
		return
	}
}

func (b *Broker) handleTransformer(cmd *protocol.Object) {
	switch cmd.CmdType {

	case globals.CMD_ADD:
		// Args: route, channel, address, nil
		t := newTransformer(cmd.Arg3)
		b.route(cmd.Arg1).channel(cmd.Arg2).addTransformer(*t)

	case globals.CMD_REMOVE:
		// Args: route, channel, address, nil
		t := newTransformer(cmd.Arg3)
		b.route(cmd.Arg1).channel(cmd.Arg2).removeTransformer(*t)

	default:
		str.WarningPrint(
			fmt.Sprintf("Unknown command type for transformer from %s",
				cmd.Responder.C.RemoteAddr().String(),
			),
		)
		return
	}

	b.printStructure()
}

func (b *Broker) handleSubscriber(cmd *protocol.Object) {
	switch cmd.CmdType {

	case globals.CMD_ADD:
		// Args: route, channel, address, nil
		s := newSubscriber(cmd.Arg3)
		b.route(cmd.Arg1).channel(cmd.Arg2).addSubscriber(*s)

	case globals.CMD_REMOVE:
		// Args: route, channel, address, nil
		s := newSubscriber(cmd.Arg3)
		b.route(cmd.Arg1).channel(cmd.Arg2).removeSubscriber(*s)

	default:
		str.WarningPrint(
			fmt.Sprintf("Unknown command type for subscriber from %s",
				cmd.Responder.C.RemoteAddr().String(),
			),
		)
		return
	}

	b.printStructure()
}

func (b *Broker) handleGlobals(cmd *protocol.Object) {
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

	default:
		str.WarningPrint(
			fmt.Sprintf("Unknown command type for globals from %s",
				cmd.Responder.C.RemoteAddr().String(),
			),
		)
		return
	}
}

func (b *Broker) handleActions(cmd *protocol.Object) {
	switch cmd.CmdType {

	case globals.CMD_SIGTERM:
		b.ManagingServer.Shutdown()

	default:
		str.WarningPrint(
			fmt.Sprintf("Unknown command type for action from %s",
				cmd.Responder.C.RemoteAddr().String(),
			),
		)
		return
	}
}

// -------Utils-----------------------------------------------------------------

// PrintStructure pretty-prints the broker structure.
func (b *Broker) printStructure() {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	str.PrintCenteredHeader("Broker Shape")
	fmt.Println("[broker]")
	for _, r := range b.routes {
		fmt.Printf("  | - [route] %s\n", r.name)

		r.mutex.RLock()
		for _, ch := range r.channels {
			fmt.Printf("        | - [channel] %s\n", ch.name)

			// Transformers
			for _, t := range ch.loadTransformers() {
				fmt.Printf("              | - [transformer] %s\n", t.Address)
			}

			// Subscribers
			for _, s := range ch.loadSubscribers() {
				fmt.Printf("              | - [subscriber] %s\n", s.Address)
			}
		}
		r.mutex.RUnlock()
	}
	str.PrintAsciiLine()
}
