package commands

// Status of the current message describing its life cycle or validity.
type Status int

const (
	StatusCreated  = iota // The initial status of a valid message.
	StatusResolved        // Message has been consumed.
	StatusInvalid         // A msg whose fields couldn't properly be decoded.
)

type Command interface {
	GetID() string
	GetCmd() uint8
}

// -----------------------------------------------------------------------------

// The Delivery itself with fields from the incoming tcp stream and various
// Mycelia used fields.
type Delivery struct {
	Cmd    uint8  `json:"cmd"`
	ID     string `json:"id"` // Should be some form of UUID.
	Route  string `json:"route"`
	Status Status `json:"status"`
	Body   []byte `json:"body"` // The primary payload to send to the consumer.
}

func (cmd *Delivery) GetID() string {
	return cmd.ID
}

func (cmd *Delivery) GetCmd() uint8 {
	return cmd.Cmd
}

func NewDelivery(cmd uint8, id, route string, body []byte) *Delivery {
	return &Delivery{
		Cmd:    cmd,
		ID:     id,
		Route:  route,
		Status: StatusCreated,
		Body:   body,
	}
}

// -----------------------------------------------------------------------------

// A subscriber is a message that informs the system of a client that wishes
// to have all messages traveling along a route forwarded to it.
// From there a Consumer is made and registered to the end channel of a route.
type Subscriber struct {
	Cmd     uint8  `json:"cmd"`     // What to do once this object hits a chnl.
	ID      string `json:"id"`      // Should be some form of UUID.
	Route   string `json:"route"`   // Which route to subscribe to.
	Channel string `json:"channel"` // Which chnl on the route to subscribe to.
	Address string `json:"address"` // Where to forward the message.
}

func (cmd *Subscriber) GetID() string {
	return cmd.ID
}

func (cmd *Subscriber) GetCmd() uint8 {
	return cmd.Cmd
}

func NewSubscriber(cmd uint8, id, route, channel, address string) *Subscriber {
	return &Subscriber{
		Cmd:     cmd,
		ID:      id,
		Route:   route,
		Channel: channel,
		Address: address,
	}
}

// -----------------------------------------------------------------------------

// Command to add a transformer to a channel. Transformers intercept and modify
// messages before they reach subscribers.
type Transformer struct {
	Cmd     uint8  `json:"cmd"`     // What to do once this object hits a chnl.
	ID      string `json:"id"`      // Should be some form of UUID.
	Route   string `json:"route"`   // Which route the channel belongs to.
	Channel string `json:"channel"` // Which channel to add transformer to.
	Address string `json:"address"` // Where to send message for transformation.
}

func (cmd *Transformer) GetID() string {
	return cmd.ID
}

func (cmd *Transformer) GetCmd() uint8 {
	return cmd.Cmd
}

func NewTransformer(cmd uint8, id, route, channel, address string) *Transformer {
	return &Transformer{
		Cmd:     cmd,
		ID:      id,
		Route:   route,
		Channel: channel,
		Address: address,
	}
}
