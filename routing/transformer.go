package routing

import (
	"fmt"
	"net"
	"time"

	"mycelia/comm"
	"mycelia/errgo"
	"mycelia/globals"
	"mycelia/logging"

	"github.com/signal-weave/rhizome"
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
func (t *transformer) apply(obj *rhizome.Object) (*rhizome.Object, error) {
	logging.LogObjectAction(
		fmt.Sprintf("Transforming delivery via %s", t.Address), obj.UID,
	)

	conn, err := net.Dial("tcp", t.Address)
	if err != nil {
		wMsg := fmt.Sprintf("Could not dial transformer %s", t.Address)
		wErr := errgo.NewError(wMsg, globals.VerbWrn)
		return obj, wErr // Return original delivery on failure
	}
	defer comm.CloseConnection(conn)

	// Send the delivery body to transformer
	_, err = conn.Write(obj.Payload)
	if err != nil {
		wMsg := fmt.Sprintf("Could not send data to transformer %s", t.Address)
		wErr := errgo.NewError(wMsg, globals.VerbWrn)
		return obj, wErr
	}

	// Read the transformed response with a timeout
	err = conn.SetReadDeadline(time.Now().Add(globals.TransformTimeout))
	if err != nil {
		m := fmt.Sprintf("Unable to set deadline for conn to %s: %s", conn.RemoteAddr(), err)
		logging.LogSystemWarning(m)
	}

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		wMsg := fmt.Sprintf("Error reading from transformer %s", t.Address)
		wErr := errgo.NewError(wMsg, globals.VerbWrn)
		return obj, wErr
	}

	// Create new delivery with transformed body
	transformedDelivery := rhizome.NewObject(
		obj.ObjType, obj.CmdType, obj.AckPlcy,
		obj.UID,
		obj.Arg1, obj.Arg2, obj.Arg3, obj.Arg4,
		buffer[:n],
	)
	transformedDelivery.Responder = obj.Responder
	transformedDelivery.Response = obj.Response
	transformedDelivery.Version = obj.Version

	logging.LogObjectAction(
		fmt.Sprintf("Transformed delivery at: %s", t.Address), obj.UID,
	)

	return transformedDelivery, nil
}
