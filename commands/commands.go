package commands

// Status of the current message describing its life cycle or validity.
type Status int

const (
	// The initial status of a valid message.
	StatusCreated = iota
	// When a message is waiting in queue to be sent to next consumer.
	StatusPending
	// Message has been consumed.
	StatusResolved
	// A message whose source string could not be properly decoded.
	StatusInvalid
)

// -----------------------------------------------------------------------------

// The Message itself with fields from the incoming tcp stream and various
// Mycelia used fields.
type SendMessage struct {
	ID     string `json:"id"` // Should be some form of UUID.
	Route  string `json:"route"`
	Status Status `json:"status"`
	Body   string `json:"body"` // The primary payload to send to the consumer.
}

// -----------------------------------------------------------------------------

// Command to add a new route to the router. Routes are the boxes where channels
// are organized.
type RegisterRoute struct {
	ID   string `json:"id"` // Should be some form of UUID.
	Name string `json:"name"`
}

// -----------------------------------------------------------------------------

// A subscriber is a message that informs the system of a client that wishes
// to have all messages traveling along a route forwarded to it.
// From there a Consumer is made and registered to the end channel of a route.
type AddSubscriber struct {
	ID      string `json:"id"`      // Should be some form of UUID.
	Route   string `json:"route"`   // Which route to subscribe to.
	Channel string `json:"channel"` // Which chnl on the route to subscribe to.
	Address string `json:"address"` // Where to forward the message.
}

// -----------------------------------------------------------------------------

type AddChannel struct {
	ID    string `json:"id"` // Should be some form of UUID.
	Route string `json:"route"`
	Name  string `json:"name"`
}
