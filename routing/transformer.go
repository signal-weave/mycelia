package routing

import (
	"fmt"
	"net"
	"time"

	"mycelia/commands"
	"mycelia/errgo"
	"mycelia/global"
	"mycelia/str"
)

// Transformer intercepts deliveries, processes them, and returns modified
// deliveries.
type Transformer struct {
	Address string
}

func NewTransformer(address string) *Transformer {
	return &Transformer{
		Address: address,
	}
}

// transformDelivery sends the delivery to the transformer service and waits for
// response.
func (t *Transformer) transformDelivery(m *commands.Delivery) (*commands.Delivery, error) {
	actionMsg := fmt.Sprintf("Transforming delivery via %s", t.Address)
	str.ActionPrint(actionMsg)

	conn, err := net.Dial("tcp", t.Address)
	if err != nil {
		wMsg := fmt.Sprintf("Could not dial transformer %s", t.Address)
		wErr := errgo.NewError(wMsg, global.VERB_WRN)
		return m, wErr // Return original delivery on failure
	}
	defer conn.Close()

	// Send the delivery body to transformer
	_, err = conn.Write([]byte(m.Body))
	if err != nil {
		wMsg := fmt.Sprintf("Could not send data to transformer %s", t.Address)
		wErr := errgo.NewError(wMsg, global.VERB_WRN)
		return m, wErr
	}

	// Read the transformed response with a timeout
	conn.SetReadDeadline(time.Now().Add(
		time.Duration(global.TransformTimeout) * time.Second))

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		wMsg := fmt.Sprintf("Error reading from transformer %s", t.Address)
		wErr := errgo.NewError(wMsg, global.VERB_WRN)
		return m, wErr
	}

	// Create new delivery with transformed body
	transformedDelivery := &commands.Delivery{
		ID:     m.ID,
		Route:  m.Route,
		Status: m.Status,
		Body:   buffer[:n],
	}

	return transformedDelivery, nil
}
