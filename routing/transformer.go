package routing

import (
	"fmt"
	"net"
	"time"

	"mycelia/errgo"
	"mycelia/globals"
	"mycelia/logging"
	"mycelia/protocol"
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
func (t *transformer) apply(obj *protocol.Object) (*protocol.Object, error) {
	logging.LogObjectAction(
		fmt.Sprintf("Transforming delivery via %s", t.Address), obj.UID,
	)

	conn, err := net.Dial("tcp", t.Address)
	if err != nil {
		wMsg := fmt.Sprintf("Could not dial transformer %s", t.Address)
		wErr := errgo.NewError(wMsg, globals.VERB_WRN)
		return obj, wErr // Return original delivery on failure
	}

	// Send the delivery body to transformer
	_, err = conn.Write([]byte(obj.Payload))
	if err != nil {
		wMsg := fmt.Sprintf("Could not send data to transformer %s", t.Address)
		wErr := errgo.NewError(wMsg, globals.VERB_WRN)
		return obj, wErr
	}

	// Read the transformed response with a timeout
	conn.SetReadDeadline(time.Now().Add(globals.TransformTimeout))

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		wMsg := fmt.Sprintf("Error reading from transformer %s", t.Address)
		wErr := errgo.NewError(wMsg, globals.VERB_WRN)
		return obj, wErr
	}

	// Create new delivery with transformed body
	transformedDelivery := protocol.NewObject(
		obj.ObjType, obj.CmdType, obj.AckPlcy,
		obj.UID,
		obj.Arg1, obj.Arg2, obj.Arg3, obj.Arg4,
		buffer[:n],
	)
	transformedDelivery.Responder = obj.Responder
	transformedDelivery.Response = obj.Response

	logging.LogObjectAction(
		fmt.Sprintf("Transformed delivery at: %s", t.Address), obj.UID,
	)

	return transformedDelivery, nil
}
