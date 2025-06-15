package routing

import (
	"fmt"
	"strings"

	"mycelia/cli"
	"mycelia/commands"
	"mycelia/utils"
)

const (
	CMD_SEND_MESSAGE   = "send_message"
	CMD_ADD_SUBSCRIBER = "add_subscriber"
	CMD_ADD_CHANNEL    = "add_channel"
	CMD_ADD_ROUTE      = "add_route"
)

// What function gets run, passing in the data field of a command envelope.
type CommandHandler func([]string)

// Creates a new router with registered commands. A new router will always
// contain a route named 'main' that contains no channels.
func NewRouter() *Router {
	router := Router{}
	router.Routes = map[string]*Route{
		"main": NewRoute("main"),
	}
	router.commandRegistry = map[string]CommandHandler{
		CMD_SEND_MESSAGE:   router.SendMessage,
		CMD_ADD_SUBSCRIBER: router.AddSubscriber,
		CMD_ADD_CHANNEL:    router.AddChannel,
		CMD_ADD_ROUTE:      router.AddRoute,
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
	commandRegistry map[string]CommandHandler
}

func (r *Router) HandleCommand(input []byte) {
	rawString := string(input)
	tokens := strings.Split(rawString, ";;")
	cmd_type := tokens[0]

	cmd, ok := r.commandRegistry[cmd_type]
	if !ok {
		msg := fmt.Sprintf("Unknown command type: %s", cmd_type)
		utils.WarningPrint(msg)
		return
	}

	cmd(tokens)
}

// -------Message Handlers------------------------------------------------------

func (r *Router) SendMessage(tokens []string) {
	if len(tokens) != 4 {
		msg := "send_message command failed, expected 4 args, got %v"
		errMsg := fmt.Sprintf(msg, len(tokens))
		utils.WarningPrint(errMsg)
		return
	}

	var msg commands.SendMessage
	msg.Status = commands.StatusCreated
	msg.ID = tokens[1]
	msg.Route = tokens[2]
	msg.Body = tokens[3]

	route, exists := r.Routes[msg.Route]
	if !exists {
		wMsg := fmt.Sprintf("Route not found: %s", msg.Route)
		utils.WarningPrint(wMsg)
		return
	}
	route.ProcessMessage(&msg)
}

func (r *Router) AddRoute(tokens []string) {
	if len(tokens) != 3 {
		msg := "add_route command failed, expected 3 args, got %v"
		errMsg := fmt.Sprintf(msg, len(tokens))
		utils.WarningPrint(errMsg)
		return
	}

	var reg commands.RegisterRoute
	reg.ID = tokens[1]
	reg.Name = tokens[2]

	_, exists := r.Routes[reg.Name]
	if !exists {
		route := NewRoute(reg.Name)
		r.Routes[reg.Name] = route
		utils.SprintfLn("Route %s registered!", reg.Name)
		return
	}

	wMsg := fmt.Sprintf("Route %s already exists.", reg.Name)
	utils.WarningPrint(wMsg)
	r.PrintRouterStructure()
}

func (r *Router) AddChannel(tokens []string) {
	if len(tokens) != 4 {
		msg := "add_channel command failed, expected 4 args, got %v"
		errMsg := fmt.Sprintf(msg, len(tokens))
		utils.WarningPrint(errMsg)
		return
	}

	var ch commands.AddChannel
	ch.ID = tokens[1]
	ch.Route = tokens[2]
	ch.Name = tokens[3]

	route, exists := r.Routes[ch.Route]
	if !exists {
		wMsg := fmt.Sprintf("Route not found %s", ch.Route)
		utils.WarningPrint(wMsg)
		return
	}
	route.AddChannel(&ch)
	r.PrintRouterStructure()
}

func (r *Router) AddSubscriber(tokens []string) {
	if len(tokens) != 5 {
		msg := "add_channel command failed, expected 5 args, got %v"
		errMsg := fmt.Sprintf(msg, len(tokens))
		utils.WarningPrint(errMsg)
		return
	}

	var sub commands.AddSubscriber
	sub.ID = tokens[1]
	sub.Route = tokens[2]
	sub.Channel = tokens[3]
	sub.Address = tokens[4]

	route, exists := r.Routes[sub.Route]
	if !exists {
		wMsg := fmt.Sprintf("Route not found %s", sub.Route)
		utils.WarningPrint(wMsg)
		return
	}
	route.AddSubscriber(&sub)
	r.PrintRouterStructure()
}

// -------Debug-----------------------------------------------------------------

// PrintRouterStructure prints the router, routes, channels, and subscribers.
func (r *Router) PrintRouterStructure() {
	if !cli.PrintTree {
		return
	}

	routeExpr := "  | - [route] %s\n"
	channelExpr := "  |     | - [channel] %s\n"
	subscriberExpr := "  |     |     | - [subscriber] %s\n"

	fmt.Println("\n[router]")
	for routeName, route := range r.Routes {
		fmt.Printf(routeExpr, routeName)
		for _, channel := range route.Channels {
			fmt.Printf(channelExpr, channel.Name)
			for _, subscriber := range channel.Subscribers {
				fmt.Printf(subscriberExpr, subscriber.Address)
			}
		}
	}

	fmt.Println()
}
