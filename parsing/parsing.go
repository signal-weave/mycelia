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

	var msg commands.SendMessage
	msg.Status = commands.StatusCreated
	msg.ID = tokens[0]
	msg.Route = tokens[1]
	msg.Body = tokens[2]

	return "send_message", &msg
}

func parseAddRouteV1(tokens []string) (string, commands.Command) {
	fmt.Println(tokens)
	if !verifyTokenLength(tokens, 2, "add_route") {
		return "add_route", nil
	}

	var route commands.AddRoute
	route.ID = tokens[0]
	route.Name = tokens[1]
	return "add_route", &route
}

func parseAddSubscriberV1(tokens []string) (string, commands.Command) {
	if !verifyTokenLength(tokens, 4, "add_subscriber") {
		return "add_subscriber", nil
	}

	var sub commands.AddSubscriber
	sub.ID = tokens[0]
	sub.Route = tokens[1]
	sub.Channel = tokens[2]
	sub.Address = tokens[3]
	return "add_subscriber", &sub
}

func parseAddChannelV1(tokens []string) (string, commands.Command) {
	if !verifyTokenLength(tokens, 3, "add_channel") {
		return "add_channel", nil
	}

	var channel commands.AddChannel
	channel.ID = tokens[0]
	channel.Route = tokens[1]
	channel.Name = tokens[2]
	return "add_channel", &channel
}

func parseAddTransformerV1(tokens []string) (string, commands.Command) {
	if !verifyTokenLength(tokens, 4, "add_transformer") {
		return "add_transformer", nil
	}

	var transformer commands.AddTransformer
	transformer.ID = tokens[0]
	transformer.Route = tokens[1]
	transformer.Channel = tokens[2]
	transformer.Address = tokens[3]
	return "add_transformer", &transformer
}
