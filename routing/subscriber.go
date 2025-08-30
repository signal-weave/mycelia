package routing

import (
	"fmt"
	"net"

	"mycelia/protocol"
	"mycelia/str"
)

// Object representation of the client subscribed to an endpoint, ie. the
// distributed machine that a delivery will be forwarded to.
type Subscriber struct {
	Address string
}

func NewSubscriber(address string) *Subscriber {
	return &Subscriber{Address: address}
}

// Forwards the delivery to the client represented by the consumer object.
func (c *Subscriber) ConsumeDelivery(m *protocol.Command) {
	fmt.Println("Attempting to dial", c.Address)
	conn, err := net.Dial("tcp", c.Address)
	if err != nil {
		wMsg := fmt.Sprintf("Could not dial %s", c.Address)
		str.WarningPrint(wMsg)
		return
	}
	defer conn.Close()

	_, err = conn.Write(m.Payload)
	if err != nil {
		eMsg := fmt.Sprintf("Error sending to %s", c.Address)
		str.ErrorPrint(eMsg)
		return
	}
	str.ActionPrint(
		fmt.Sprintf("Wrote delivery to: %s", c.Address),
	)
}
