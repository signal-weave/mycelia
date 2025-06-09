package routing

import (
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
		wMsg := fmt.Sprintf("Could not dial %s", c.Address)
		utils.WarningPrint(wMsg)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(m.Body))
	if err != nil {
		eMsg := fmt.Sprintf("Error sending to %s", c.Address)
		utils.ErrorPrint(eMsg)
		return
	}
	m.Status = commands.StatusResolved
}
