package routing

import (
	"encoding/json"
	"fmt"

	"mycelia/commands"
	"mycelia/utils"
)

// What function gets run, passing in the data field of a command envelope.
type CommandHandler func(commands.Envelope)

// Creates a new router with registered commands. A new router will always
// contain a route named 'main' that contains no channels.
func NewRouter() *Router {
	router := Router{}
	router.Routes = map[string]*Route{
		"main": NewRoute("main"),
	}
	router.commandRegistry = map[commands.PacketType]CommandHandler{
		"send_message":   router.SendMessage,
		"add_subscriber": router.AddSubscriber,
		"add_channel":    router.AddChannel,
		"register_route": router.RegisterRoute,
	}

	return &router
}

// The structure responsible for sending message through routes,
// applying crud operations to routes, nad adding consumers to routes as
// subscribers.
// Routes are comprised of sequences of channels.
type Router struct {
	// Route names to list of channels the route is comprised of.
	// Mesasges will travel through all channels on the route.
	Routes map[string]*Route

	// The map of envelope type strings to runnable commands.
	// The data field of the envolope is passed through to the command handler.
	commandRegistry map[commands.PacketType]CommandHandler
}

func (r *Router) HandleEnvelope(input []byte) {
	var env commands.Envelope
	if err := json.Unmarshal(input, &env); err != nil {
		fmt.Println("Invalid envelop:", err)
		return
	}

	cmd, ok := r.commandRegistry[env.Type]
	if !ok {
		msg := fmt.Sprintf("Unknown command type: %s", env.Type)
		utils.SprintfLnIndent(msg, 2)
		return
	}

	cmd(env)
}

// -------Message Handlers------------------------------------------------------

func (r *Router) SendMessage(env commands.Envelope) {
	var msg commands.SendMessage
	if err := json.Unmarshal(env.Data, &msg); err != nil {
		utils.SprintfLnIndent("Invalid message: %s", 2, err.Error())
		return
	}

	msg.Status = commands.StatusCreated
	route, exists := r.Routes[msg.Route]
	if !exists {
		utils.SprintfLnIndent("Route not found: %s", 2, msg.Route)
		return
	}
	route.ProcessMessage(&msg)
}

func (r *Router) RegisterRoute(env commands.Envelope) {
	var reg commands.RegisterRoute
	if err := json.Unmarshal(env.Data, &reg); err != nil {
		utils.SprintfLnIndent("Invalid route: %s", 2, err.Error())
		return
	}
	_, exists := r.Routes[reg.Name]
	if !exists {
		route := NewRoute(reg.Name)
		r.Routes[reg.Name] = route
		utils.SprintfLn("Route %s registered!", reg.Name)
		return
	}
	utils.SprintfLn("Route %s already exists.", reg.Name)
}

func (r *Router) AddChannel(env commands.Envelope) {
	var ch commands.AddChannel
	if err := json.Unmarshal(env.Data, &ch); err != nil {
		utils.SprintfLnIndent("Invalid channel %s", 2, err.Error())
		return
	}
	route, exists := r.Routes[ch.Route]
	if !exists {
		utils.SprintfLnIndent("Route not found %s", 2, ch.Route)
		return
	}
	route.AddChannel(&ch)
}

func (r *Router) AddSubscriber(env commands.Envelope) {
	var sub commands.AddSubscriber
	if err := json.Unmarshal(env.Data, &sub); err != nil {
		utils.SprintfLnIndent("Invalid subscription: %s", 2, err.Error())
		return
	}
	route, exists := r.Routes[sub.Route]
	if !exists {
		utils.SprintfLnIndent("Route not found %s", 2, sub.Route)
		return
	}
	route.AddSubscriber(&sub)
}
