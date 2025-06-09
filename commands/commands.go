package commands

import (
	"encoding/json"
)

// A temporary object that contains the packet type and the data to be handled
// by the server.
type Envelope struct {
	Type PacketType      `json:"type"`
	Data json.RawMessage `json:"data"`
}

// -----------------------------------------------------------------------------

// The Message itself with fields from the incoming tcp stream and various
// Mycelia used fields.
type SendMessage struct {
	// Unsure if I want this - There should be some primary field.
	Route string `json:"route"`

	Status Status `json:"status"`

	// The primary payload to send to the consumer.
	Body map[string]any `json:"body"`
}

// -----------------------------------------------------------------------------

// Command to add a new route to the router. Routes are the boxes where channels
// are organized.
type RegisterRoute struct {
	Name string `json:"name"`
}

// -----------------------------------------------------------------------------

// A subscriber is a message that informs the system of a client that wishes
// to have all messages traveling along a route forwarded to it.
// From there a Consumer is made and registered to the end channel of a route.
type AddSubscriber struct {
	Route   string `json:"route"`   // Which route to subscribe to.
	Channel string `json:"channel"` // Which chnl on the route to subscribe to.
	Address string `json:"address"` // Where to forward the message.
}

// -----------------------------------------------------------------------------

type AddChannel struct {
	Route string `json:"route"`
	Name  string `json:"name"`
}
