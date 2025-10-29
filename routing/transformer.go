package routing

import (
	"context"
	"fmt"
	"time"

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
	logging.LogObjectAction(fmt.Sprintf("Transforming delivery via %s", t.Address), obj.UID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	b, err := globalConnPool.Get(ctx, t.Address)
	if err != nil {
		wMsg := fmt.Sprintf("Could not dial transformer %s", t.Address)
		wErr := errgo.NewError(wMsg, globals.VerbWrn)
		return obj, wErr // Return original delivery on failure
	}
	defer b.Put()

	// Send the delivery body to transformer
	_ = b.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err = b.Conn().Write(obj.Payload); err != nil {
		b.MarkBroken()
		wMsg := fmt.Sprintf("Could not send data to transformer %s", t.Address)
		wErr := errgo.NewError(wMsg, globals.VerbWrn)
		return obj, wErr
	}

	// Read the transformed response with timeout
	_ = b.SetReadDeadline(time.Now().Add(globals.TransformTimeout))

	buffer := make([]byte, 4096)
	n, err := b.Conn().Read(buffer)
	if err != nil {
		b.MarkBroken()
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

	logging.LogObjectAction(fmt.Sprintf("Transformed delivery at: %s", t.Address), obj.UID)

	return transformedDelivery, nil
}
