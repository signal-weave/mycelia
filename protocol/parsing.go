package protocol

import (
	"fmt"
	"io"

	"mycelia/boot"
	"mycelia/commands"
	"mycelia/errgo"
	"mycelia/global"
)

var _ = boot.RuntimeCfg // REQUIRED for global config values.

// parseProtoVer extracts only the protocol version and returns it along with
// a slice that starts at the next byte (i.e., the remainder of the message).
func parseProtoVer(data []byte) (uint8, []byte, error) {
	const u8len = 1
	if len(data) < u8len {
		return 0, nil, io.ErrUnexpectedEOF
	}
	ver := data[0]
	return ver, data[u8len:], nil
}

func ParseLine(line []byte) (commands.Command, error) {
	version, rest, err := parseProtoVer(line)
	if err != nil {
		wMsg := fmt.Sprintf("Read protocol version: %v", err)
		wErr := errgo.NewError(wMsg, global.VERB_WRN)
		return nil, wErr
	}

	// The broker always works off of the same types of command objects.
	// Command objects may evolve over time, adding new fields for new
	// functionality, but the broker should remain compatible with previous
	// client side API versions.

	// If a client is using API ver 1 to communicate with Broker ver 2, then the
	// client should be able to still communicate.
	// This first token of a message is the API version, and this switch runs
	// the corresponding parsing logic.

	// This is mainly because early on there was uncertainty if the protocol and
	// command structure was done right, and we reserved the ability to update
	// it as we go.
	switch version {
	case 1:
		return decodeV1(rest)
	default:
		wErr := errgo.NewError("Unable to parse command!", global.VERB_WRN)
		return nil, wErr
	}
}
