package parsing

import (
	"fmt"
	"strconv"
	"strings"

	"mycelia/commands"
	"mycelia/str"
)

func ParseData(data []byte) (string, commands.Command) {
	rawString := string(data)
	tokens := strings.Split(rawString, ";;")
	version, err := strconv.Atoi(tokens[0])
	if err != nil {
		str.WarningPrint("Could not parse data - Missing API version token!")
		return "err", nil
	}

	cmdTokens := tokens[1:] // prune off protocol version token.
	var s string
	var cmd commands.Command

	switch version {
	case 1:
		s, cmd = parseDataV1(cmdTokens)
	}

	return s, cmd
}

func verifyTokenLength(tokens []string, length int, cmdName string) bool {
	if len(tokens) != length {
		msg := "%s command failed, expected %d args, got %d isntead"
		wMsg := fmt.Sprintf(msg, cmdName, length, len(tokens))
		str.WarningPrint(wMsg)
		return false
	}
	return true
}

// ------Version 1--------------------------------------------------------------

func parseDataV1(tokens []string) (string, commands.Command) {
	cmdType := tokens[0]
	cmdTokens := tokens[1:] // prune off command type token
	var s string
	var cmd commands.Command

	switch cmdType {
	case "send_message":
		s, cmd = parseSendMsgV1(cmdTokens)
	case "add_route":
		s, cmd = parseAddRouteV1(cmdTokens)
	case "add_subscriber":
		s, cmd = parseAddSubscriberV1(cmdTokens)
	case "add_channel":
		s, cmd = parseAddChannelV1(cmdTokens)
	case "add_transformer":
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
	if !verifyTokenLength(tokens, 3, "send_message") {
		return "send_message", nil
	}

	sm := commands.NewSendMessage(
		tokens[0], // ID
		tokens[1], // Route
		tokens[2], // Body
	)
	return "send_message", sm
}

func parseAddRouteV1(tokens []string) (string, commands.Command) {
	fmt.Println(tokens)
	if !verifyTokenLength(tokens, 2, "add_route") {
		return "add_route", nil
	}

	ar := commands.NewAddRoute(
		tokens[0], // ID
		tokens[1], // Name
	)
	return "add_route", ar
}

func parseAddSubscriberV1(tokens []string) (string, commands.Command) {
	if !verifyTokenLength(tokens, 4, "add_subscriber") {
		return "add_subscriber", nil
	}

	as := commands.NewAddSubscriber(
		tokens[0], // ID
		tokens[1], // Route
		tokens[2], // Channel
		tokens[3], // Address
	)
	return "add_subscriber", as
}

func parseAddChannelV1(tokens []string) (string, commands.Command) {
	if !verifyTokenLength(tokens, 3, "add_channel") {
		return "add_channel", nil
	}

	ac := commands.NewAddChannel(
		tokens[0], // ID
		tokens[1], // Route
		tokens[2], // Name
	)
	return "add_channel", ac
}

func parseAddTransformerV1(tokens []string) (string, commands.Command) {
	if !verifyTokenLength(tokens, 4, "add_transformer") {
		return "add_transformer", nil
	}

	at := commands.NewAddTransformer(
		tokens[0], // ID
		tokens[1], // Route
		tokens[2], // Channel
		tokens[3], // Address
	)
	return "add_transformer", at
}
