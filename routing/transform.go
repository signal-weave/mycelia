package routing

import (
	"fmt"
	"net"
	"time"

	"mycelia/commands"
	"mycelia/cli"
	"mycelia/str"
)

func NewTransformer(address string) *Transformer {
	return &Transformer{
		Address: address,
	}
}

// Transformer intercepts messages, processes them, and returns modified messages.
type Transformer struct {
	Address string
}

// TransformMessage sends the message to the transformer service and waits for response.
func (t *Transformer) TransformMessage(m *commands.SendMessage) (*commands.SendMessage, error) {
	actionMsg := fmt.Sprintf("Transforming message via %s", t.Address)
	str.ActionPrint(actionMsg)

	conn, err := net.Dial("tcp", t.Address)
	if err != nil {
		wMsg := fmt.Sprintf("Could not dial transformer %s", t.Address)
		str.WarningPrint(wMsg)
		return m, err // Return original message on failure
	}
	defer conn.Close()

	// Send the message body to transformer
	_, err = conn.Write([]byte(m.Body))
	if err != nil {
		eMsg := fmt.Sprintf("Could not send data to transformer %s", t.Address)
		str.ErrorPrint(eMsg)
		return m, err
	}

	// Read the transformed response with a timeout
	conn.SetReadDeadline(time.Now().Add(
		time.Duration(cli.RuntimeCfg.TransformTimeout) * time.Second))

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		eMsg := fmt.Sprintf("Error reading from transformer %s", t.Address)
		str.ErrorPrint(eMsg)
		return m, err
	}

	// Create new message with transformed body
	transformedMessage := &commands.SendMessage{
		ID:     m.ID,
		Route:  m.Route,
		Status: m.Status,
		Body:   string(buffer[:n]),
	}

	return transformedMessage, nil
}
