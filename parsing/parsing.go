package parsing

import (
	"strconv"

	"mycelia/commands"
	"mycelia/str"
)

const unknownCommand = "err"

func ParseLine(line []byte) (string, commands.Command) {
	parts := splitTokens(line)
	if len(parts) < 2 {
		str.WarningPrint("Could not parse data - missing version or command!")
		return unknownCommand, nil
	}

	verStr, _ := unescape(parts[0])
	version, err := strconv.Atoi(verStr)
	if err != nil {
		return unknownCommand, nil
	}

	args, err := unescapeTokens(parts[1:]) // prune off protocol version token.
	if err != nil {
		return unknownCommand, nil
	}

	switch version {
	case 1:
		return parseDataV1(args)
	default:
		return unknownCommand, nil
	}
}
