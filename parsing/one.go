package parsing

import (
	"mycelia/commands"
)

// Version 1 of the command API does not support sub-command parsing.
// The <object>.<action> syntax is the conform to future version feature syntax.
const (
	CMD_MESSAGE_SEND    = "MESSAGE.SEND"
	CMD_SUBSCRIBER_ADD  = "SUBSCRIBER.ADD"
	CMD_TRANSFORMER_ADD = "TRANSFORMER.ADD"
)

// -----------------------------------------------------------------------------
// Version 1 command decoding.
// Currently handles arrays of string tokens - will convert to byte array and
// buffer management in the future.
// -----------------------------------------------------------------------------

func parseDataV1(tokens []string) (string, commands.Command) {
	cmdType := tokens[0]
	cmdTokens := tokens[1:] // prune off command type token.
	var s string
	var cmd commands.Command

	switch cmdType {
	case CMD_MESSAGE_SEND:
		s, cmd = parseSendMsgV1(cmdTokens)
	case CMD_SUBSCRIBER_ADD:
		s, cmd = parseAddSubscriberV1(cmdTokens)
	case CMD_TRANSFORMER_ADD:
		s, cmd = parseAddTransformerV1(cmdTokens)
	}

	return s, cmd
}

// -----------------------------------------------------------------------------
// A fair amount of command generation shoves the decoded tokens straigt into a
// var declared command. This is done rather than useing the NewCommand() funcs
// so that we can pass the ID in along with and the code looks uniform.
// -----------------------------------------------------------------------------

func parseSendMsgV1(tokens []string) (string, commands.Command) {
	if !verifyTokenLength(tokens, 3, CMD_MESSAGE_SEND) {
		return CMD_MESSAGE_SEND, nil
	}

	sm := commands.NewSendMessage(
		tokens[0], // ID
		tokens[1], // Route
		tokens[2], // Body
	)
	return CMD_MESSAGE_SEND, sm
}

func parseAddSubscriberV1(tokens []string) (string, commands.Command) {
	if !verifyTokenLength(tokens, 4, CMD_SUBSCRIBER_ADD) {
		return CMD_SUBSCRIBER_ADD, nil
	}

	as := commands.NewAddSubscriber(
		tokens[0], // ID
		tokens[1], // Route
		tokens[2], // Channel
		tokens[3], // Address
	)
	return CMD_SUBSCRIBER_ADD, as
}

func parseAddTransformerV1(tokens []string) (string, commands.Command) {
	if !verifyTokenLength(tokens, 4, CMD_TRANSFORMER_ADD) {
		return CMD_TRANSFORMER_ADD, nil
	}

	at := commands.NewAddTransformer(
		tokens[0], // ID
		tokens[1], // Route
		tokens[2], // Channel
		tokens[3], // Address
	)
	return CMD_TRANSFORMER_ADD, at
}
