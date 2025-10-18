package routing

import (
	"fmt"
	"sync"

	"mycelia/errgo"
	"mycelia/globals"
	"mycelia/logging"
	"mycelia/str"

	"github.com/signal-weave/rhizome"
)

// This is here so the server that spawns the broker can add itself without
// causing a circular dependency.
type server interface {
	UpdateListener() error
	GetAddress() string
	GetPort() int
	Shutdown()
}

// -------Base Broker Details---------------------------------------------------

// Broker is the primary object orchestrator.
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

// HandleBytes handles the raw byte form of a object, hot off a socket, converts
// it to an object, and forwards it to the object handler.
func (b *Broker) HandleBytes(input []byte, resp *rhizome.ConnResponder) {
	// Parse byte stream -> object.
	obj, err := rhizome.DecodeFrame(input, resp)
	if err != nil {
		return
	}

	b.HandleObject(obj)
}

// -------Route Management------------------------------------------------------

// getRoute returns ptr to existing or nil.
func (b *Broker) getRoute(obj *rhizome.Object) *route {
	b.mutex.RLock()
	r := b.routes[obj.Arg1]
	b.mutex.RUnlock()

	if r == nil {
		obj.ResponeWithAck(globals.ACK_ROUTE_NOT_FOUND)
		return nil
	}

	return r
}

// Creates a route from the given object whose Arg1 is the route name.
// If the route already exists, the existing route is returned instead.
func (b *Broker) createRoute(obj *rhizome.Object) *route {
	r := b.getRoute(obj)
	if r != nil {
		return r
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()
	if r = b.routes[obj.Arg1]; r == nil {
		r = newRoute(b, obj.Arg1)
		b.routes[obj.Arg1] = r
		logging.LogObjectAction(
			fmt.Sprintf("Created route: %s", obj.Arg1), obj.UID,
		)
	}
	return r
}

// Called by routes to cull themselves if empty.
func (b *Broker) removeEmptyRoute(name string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	_, exists := b.routes[name]
	if exists {
		delete(b.routes, name)
	}
}

// Gets a channel from a rhizome object whose arg1 is the route and arg2 is the
// channel.
// Will return a pointer if found or nil if not.
// Will write response with nack code for which item wasn't found.
func (b *Broker) getChannel(obj *rhizome.Object) *channel {
	r := b.getRoute(obj)
	if r == nil {
		obj.ResponeWithAck(globals.ACK_ROUTE_NOT_FOUND)
		return nil
	}
	c := r.getChannel(obj.Arg2)
	if c == nil {
		obj.ResponeWithAck(globals.ACK_CHANNEL_NOT_FOUND)
		return nil
	}
	return c
}

// -------Object Handling------------------------------------------------------

// HandleObject routes the object generated from the incoming byte stream.
// Is exported for boot to load PreInit.json structures into.
func (b *Broker) HandleObject(obj *rhizome.Object) error {
	switch obj.ObjType {
	case globals.OBJ_DELIVERY:
		b.handleDelivery(obj)
	case globals.OBJ_CHANNEL:
		b.handleChannel(obj)
	case globals.OBJ_TRANSFORMER:
		b.handleTransformer(obj)
	case globals.OBJ_SUBSCRIBER:
		b.handleSubscriber(obj)
	case globals.OBJ_GLOBALS:
		b.handleGlobals(obj)
	case globals.OBJ_Action:
		b.handleActions(obj)
	default:
		wErr := errgo.NewError("Unknown object type!", globals.VERB_WRN)
		return wErr
	}

	return nil
}

func (b *Broker) handleDelivery(obj *rhizome.Object) {
	switch obj.CmdType {

	case globals.CMD_SEND:
		r := b.getRoute(obj)
		if r == nil {
			obj.ResponeWithAck(globals.ACK_ROUTE_NOT_FOUND)
			return
		}
		r.enqueue(obj) // no channels means route will send to dead letter.

	default:
		logging.LogObjectWarning(
			fmt.Sprintf("Unknown command type for delivery from %s",
				obj.Responder.RemoteAddr(),
			), obj.UID,
		)
		return
	}
}

func (b *Broker) handleChannel(obj *rhizome.Object) {
	switch obj.CmdType {

	case globals.CMD_ADD:
		// Args: route, name, strategy, nil
		r := b.createRoute(obj)
		r.createChannel(obj)

	case globals.CMD_REMOVE:
		// Args: route, name, nil, nil
		fmt.Println("CHANNEL.REMOVE not yet implemented.")
		return

	default:
		logging.LogObjectWarning(
			fmt.Sprintf("Unknown command type for channel from %s",
				obj.Responder.RemoteAddr(),
			), obj.UID,
		)
		return
	}

	b.printStructure()
}

func (b *Broker) handleTransformer(obj *rhizome.Object) {
	switch obj.CmdType {

	case globals.CMD_ADD:
		// Args: route, channel, address, nil
		t := newTransformer(obj.Arg3)
		c := b.getChannel(obj)
		if c == nil {
			return
		}
		c.addTransformer(*t)

	case globals.CMD_REMOVE:
		// Args: route, channel, address, nil
		t := newTransformer(obj.Arg3)
		c := b.getChannel(obj)
		if c == nil {
			return
		}
		c.removeTransformer(*t)

	default:
		logging.LogObjectWarning(
			fmt.Sprintf("Unknown command type for transformer from %s",
				obj.Responder.RemoteAddr(),
			), obj.UID,
		)
		return
	}

	b.printStructure()
}

func (b *Broker) handleSubscriber(obj *rhizome.Object) {
	switch obj.CmdType {

	case globals.CMD_ADD:
		// Args: route, channel, address, nil
		s := newSubscriber(obj.Arg3)
		c := b.getChannel(obj)
		if c == nil {
			return
		}
		c.addSubscriber(*s)

	case globals.CMD_REMOVE:
		// Args: route, channel, address, nil
		s := newSubscriber(obj.Arg3)
		c := b.getChannel(obj)
		if c == nil {
			return
		}
		c.removeSubscriber(*s)

	default:
		logging.LogObjectWarning(
			fmt.Sprintf("Unknown command type for subscriber from %s",
				obj.Responder.RemoteAddr(),
			), obj.UID,
		)
		return
	}

	b.printStructure()
}

func (b *Broker) handleGlobals(obj *rhizome.Object) {
	switch obj.CmdType {

	case globals.CMD_UPDATE:
		hasPermission := updateGlobals(obj)
		if !hasPermission {
			return
		}
		if b.ManagingServer.GetAddress() != globals.Address ||
			b.ManagingServer.GetPort() != globals.Port {
			b.ManagingServer.UpdateListener()
		}

	default:
		logging.LogObjectWarning(
			fmt.Sprintf("Unknown command type for globals from %s",
				obj.Responder.RemoteAddr(),
			), obj.UID,
		)
		return
	}
}

func (b *Broker) handleActions(obj *rhizome.Object) {
	switch obj.CmdType {

	case globals.CMD_SIGTERM:
		b.ManagingServer.Shutdown()

	default:
		logging.LogObjectWarning(
			fmt.Sprintf("Unknown command type for action from %s",
				obj.Responder.RemoteAddr(),
			), obj.UID,
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
