package routing

import (
	"encoding/json"
	"fmt"
	"net"

	"mycelia/commands"
	"mycelia/utils"
)

func NewConsumer(address string) *Consumer {
	return &Consumer{Address: address}
}

// Object representing the client subscribed to an endpoint.
type Consumer struct {
	Address string
}

// Forwards the message to the client represented by the consumer object.
func (c *Consumer) ConsumeMessage(m *commands.SendMessage) {
	fmt.Println("Attempting to dial", c.Address)
	conn, err := net.Dial("tcp", c.Address)
	if err != nil {
		utils.SprintfLnIndent("Could not dial %s", 2, c.Address)
		return
	}
	defer conn.Close()

	payload, err := json.Marshal(m.Body)
	if err != nil {
		utils.MessageIfError("Error marshaling mesasge body", err)
		return
	}
	_, err = conn.Write([]byte(payload))
	if err != nil {
		msg := fmt.Sprintf("Error sending to %s", c.Address)
		fmt.Println(msg, ": ", err)
		return
	}
	m.Status = commands.StatusResolved
}
