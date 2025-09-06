package routing

import (
	"fmt"
	"net"

	"mycelia/protocol"
	"mycelia/str"
)

// Object representation of the client subscribed to an endpoint, ie. the
// distributed machine that a delivery will be forwarded to.
type subscriber struct {
	Address string
}

func newSubscriber(address string) *subscriber {
	return &subscriber{Address: address}
}

// Forwards the delivery to the client represented by the consumer object.
func (c *subscriber) deliver(m *protocol.Object) {
	str.ActionPrint(fmt.Sprintf("Attempting to dial %s", c.Address))

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
