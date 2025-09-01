package routing

import (
	"fmt"
	"net"
	"time"

	"mycelia/errgo"
	"mycelia/globals"
	"mycelia/protocol"
	"mycelia/str"
)

// transformer intercepts deliveries, processes them, and returns modified
// deliveries.
type transformer struct {
	Address string
}

func newTransformer(address string) *transformer {
	return &transformer{
		Address: address,
	}
}

// apply sends the delivery to the transformer service and waits for
// response.
func (t *transformer) apply(m *protocol.Command) (*protocol.Command, error) {
	actionMsg := fmt.Sprintf("Transforming delivery via %s", t.Address)
	str.ActionPrint(actionMsg)

	conn, err := net.Dial("tcp", t.Address)
	if err != nil {
		wMsg := fmt.Sprintf("Could not dial transformer %s", t.Address)
		wErr := errgo.NewError(wMsg, globals.VERB_WRN)
		return m, wErr // Return original delivery on failure
	}
	defer conn.Close()

	// Send the delivery body to transformer
	_, err = conn.Write([]byte(m.Payload))
	if err != nil {
		wMsg := fmt.Sprintf("Could not send data to transformer %s", t.Address)
		wErr := errgo.NewError(wMsg, globals.VERB_WRN)
		return m, wErr
	}

	// Read the transformed response with a timeout
	conn.SetReadDeadline(time.Now().Add(globals.TransformTimeout))

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		wMsg := fmt.Sprintf("Error reading from transformer %s", t.Address)
		wErr := errgo.NewError(wMsg, globals.VERB_WRN)
		return m, wErr
	}

	// Create new delivery with transformed body
	transformedDelivery := protocol.NewCommand(
		m.ObjType,
		m.CmdType,

		m.ReturnAdress,
		m.UID,

		m.Arg1,
		m.Arg2,
		m.Arg3,
		m.Arg4,
		buffer[:n],
	)

	str.ActionPrint(
		fmt.Sprintf("Transformed delivery at: %s", t.Address),
	)

	return transformedDelivery, nil
}
