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

type Command interface {
	GetID() string
}

// -----------------------------------------------------------------------------

// The Message itself with fields from the incoming tcp stream and various
// Mycelia used fields.
type SendMessage struct {
	ID     string `json:"id"` // Should be some form of UUID.
	Route  string `json:"route"`
	Status Status `json:"status"`
	Body   string `json:"body"` // The primary payload to send to the consumer.
}

func (cmd *SendMessage) GetID() string {
	return cmd.ID
}

func NewSendMessage(id, route, body string) *SendMessage {
	return &SendMessage{
		ID: id,
		Route: route,
		Status: StatusCreated,
		Body: body,
	}
}

// -----------------------------------------------------------------------------

// Command to add a new route to the router. Routes are the boxes where channels
// are organized.
type AddRoute struct {
	ID   string `json:"id"` // Should be some form of UUID.
	Name string `json:"name"`
}

func (cmd *AddRoute) GetID() string {
	return cmd.ID
}

func NewAddRoute(id, name string) *AddRoute {
	return &AddRoute{
		ID:   id,
		Name: name,
	}
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

func (cmd *AddSubscriber) GetID() string {
	return cmd.ID
}

func NewAddSubscriber(id, route, channel, address string) *AddSubscriber {
	return &AddSubscriber{
		ID:      id,
		Route:   route,
		Channel: channel,
		Address: address,
	}
}

// -----------------------------------------------------------------------------

type AddChannel struct {
	ID    string `json:"id"` // Should be some form of UUID.
	Route string `json:"route"`
	Name  string `json:"name"`
}

func (cmd *AddChannel) GetID() string {
	return cmd.ID
}

func NewAddChannel(id, route, name string) *AddChannel {
	return &AddChannel{
		ID:    id,
		Route: route,
		Name:  name,
	}
}

// -----------------------------------------------------------------------------

// Command to add a transformer to a channel. Transformers intercept and modify
// messages before they reach subscribers.
type AddTransformer struct {
	ID      string `json:"id"`      // Should be some form of UUID.
	Route   string `json:"route"`   // Which route the channel belongs to.
	Channel string `json:"channel"` // Which channel to add transformer to.
	Address string `json:"address"` // Where to send message for transformation.
}

func (cmd *AddTransformer) GetID() string {
	return cmd.ID
}

func NewAddTransformer(id, route, channel, address string) *AddTransformer {
	return &AddTransformer{
		ID:      id,
		Route:   route,
		Channel: channel,
		Address: address,
	}
}
