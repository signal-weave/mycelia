package routing

import "mycelia/protocol"

// -------Shared unit test helpers.---------------------------------------------

// minimal helper for a message
func msg(body string) *protocol.Command {
	return &protocol.Command{Payload: []byte(body)}
}
