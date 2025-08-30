package protocol

import (
	"fmt"
	"io"

	"mycelia/errgo"
	"mycelia/globals"
)

// -----------------------------------------------------------------------------
// The primary parsing entry point.
// The main protocol version detection and parsing version handling.
// -----------------------------------------------------------------------------

// A command is an object that is decoded from the incoming byte stream and ran
// through the system.
type Command struct {
	ObjType uint8
	CmdType uint8

	Sender string
	UID    string

	Arg1 string
	Arg2 string
	Arg3 string
	Arg4 string

	Payload []byte
}

func NewCommand(
	objType, cmdType uint8,
	sender, uid, arg1, arg2, arg3, arg4 string,
	payload []byte) *Command {
	return &Command{
		ObjType: objType,
		CmdType: cmdType,
		Sender:  sender,
		UID:     uid,
		Arg1:    arg1,
		Arg2:    arg2,
		Arg3:    arg3,
		Arg4:    arg4,
		Payload: payload,
	}
}

// parseProtoVer extracts only the protocol version and returns it along with
// a slice that starts at the next byte (i.e., the remainder of the message).
func parseProtoVer(data []byte) (uint8, []byte, error) {
	const u8len = 1
	if len(data) < u8len {
		return 0, nil, io.ErrUnexpectedEOF
	}
	ver := uint8(data[0])
	return ver, data[u8len:], nil
}

func ParseLine(line []byte) (*Command, error) {
	version, rest, err := parseProtoVer(line)
	if err != nil {
		wMsg := fmt.Sprintf("Read protocol version: %v", err)
		wErr := errgo.NewError(wMsg, globals.VERB_WRN)
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
		wErr := errgo.NewError("Unable to parse command!", globals.VERB_WRN)
		return nil, wErr
	}
}
