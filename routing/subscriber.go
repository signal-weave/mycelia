package routing

import (
	"context"
	"fmt"
	"time"

	"mycelia/logging"

	"github.com/signal-weave/rhizome"
)

// Object representation of the client subscribed to an endpoint, i.e. the
// distributed machine that a delivery will be forwarded to.
type subscriber struct {
	Address string
}

func newSubscriber(address string) *subscriber {
	return &subscriber{Address: address}
}

// Forwards the delivery to the client represented by the consumer object.
func (c *subscriber) deliver(obj *rhizome.Object) {
	logging.LogObjectAction(fmt.Sprintf("Attempting to dial %s", c.Address), obj.UID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	b, err := globalConnPool.Get(ctx, c.Address)
	if err != nil {
		logging.LogObjectWarning(fmt.Sprintf("Could not dial %s", c.Address), obj.UID)
		return
	}
	defer b.Put()

	_ = b.SetWriteDeadline(time.Now().Add(10 * time.Second))

	_, err = b.Conn().Write(obj.Payload)
	if err != nil {
		b.MarkBroken()
		wMsg := fmt.Sprintf("Error sending to %s", c.Address)
		logging.LogObjectWarning(wMsg, obj.UID)
		return
	}
	logging.LogObjectAction(fmt.Sprintf("Wrote delivery to: %s", c.Address), obj.UID)
}
