package protocol

import (
	"bytes"
	"fmt"
	"io"
	"mycelia/errgo"
	"mycelia/globals"
)

// -----------------------------------------------------------------------------
// Version 1 command decoding.
// -----------------------------------------------------------------------------
// *Note that this is a messaging protocol, not a file transfer protocol
// -----------------------------------------------------------------------------
// The version 1 protocol looks as follows:

// # Fixed field sized header
// +---------+--------+-------------+-------------+
// | u32 len | u8 ver | u8 obj_type | u8 cmd_type |
// +---------+--------+-------------+-------------+

// which is then followed by a variable field sized sub-header

// # UID Sub-header
// +-------------+
// | u8 len uid  |
// +-------------+

// which is then followed by 4 uint8 sized byte fields that act as arguments for
// the command type in the fixed header.
// Because these are byte streams, all arguments are considered string types
// unless the executor casts them to another type.

// # Argument Sub-Header
// +---------------+---------------+---------------+---------------+
// |  u8 len arg1  |  u8 len arg2  |  u8 len arg3  |  u8 len arg4  |
// +---------------+---------------+---------------+---------------+

// And finally the message payload that would be delivered to external sources.
// If this is unused because the message is changing the internals of the broker
// at runtime, then the field defaults to a vlaue of 0x00.

// # Globals Body
// +-----------------+
// | u16 len payload |
// +-----------------+
// -----------------------------------------------------------------------------

func decodeV1(data []byte) (*Command, error) {
	r := bytes.NewReader(data)
	cmd := &Command{}

	// ObjType + CmdType
	cmd, err := parseBaseHeader(r, cmd)
	if err != nil {
		return nil, err
	}
	// UID + Source Address
	cmd, err = parseTrackingHeader(r, cmd)
	if err != nil {
		return nil, err
	}
	// Arg fields
	cmd, err = parseArgumentFields(r, cmd)
	if err != nil {
		return nil, err
	}
	// Payload
	payload, err := readBytesU16(r)
	if err != nil {
		wMsg := fmt.Sprintf(
			"Unable to parse payload from %s: %s", cmd.Sender, err,
		)
		wErr := errgo.NewError(wMsg, globals.VERB_WRN)
		return nil, wErr
	}
	cmd.Payload = payload

	if r.Len() != 0 {
		cmd = nil
		err = errgo.NewError("Unaccounted data in reader", globals.VERB_WRN)
	}

	return cmd, err
}

// Parses the header after version: obj_type, and cmd_type from message.
func parseBaseHeader(r io.Reader, cmd *Command) (*Command, error) {
	if err := readU8(r, &cmd.ObjType); err != nil {
		wMsg := fmt.Sprintf(
			"Unable to parse u8 ObjType field from message: %s", err,
		)
		wErr := errgo.NewError(wMsg, globals.VERB_WRN)
		return nil, wErr
	}

	if err := readU8(r, &cmd.CmdType); err != nil {
		wMsg := fmt.Sprintf(
			"Unable to parse u8 CmdType field from message: %s", err,
		)
		wErr := errgo.NewError(wMsg, globals.VERB_WRN)
		return nil, wErr
	}

	return cmd, nil
}

// Parses the UID and sender address from the reader.
func parseTrackingHeader(r io.Reader, cmd *Command) (*Command, error) {
	// UID field comes before sender address field.
	uid, err := readStringU8(r)
	if err != nil {
		wMsg := fmt.Sprintf(
			"Unable to parse string UID field from message: %s", err,
		)
		wErr := errgo.NewError(wMsg, globals.VERB_WRN)
		return nil, wErr
	}
	cmd.UID = uid

	senderAddr, err := readStringU16(r)
	if err != nil {
		wMsg := fmt.Sprintf(
			"Unable to parse string address field from message: %s", err,
		)
		wErr := errgo.NewError(wMsg, globals.VERB_WRN)
		return nil, wErr
	}
	cmd.Sender = senderAddr

	return cmd, nil
}

// Parse the four argument fields from the reader.
func parseArgumentFields(r io.Reader, cmd *Command) (*Command, error) {
	arg1, err := readStringU8(r)
	if err != nil {
		wMsg := fmt.Sprintf("Unable to parse argument position %d for %s: %s",
			1, cmd.Sender, err,
		)
		wErr := errgo.NewError(wMsg, globals.VERB_WRN)
		return nil, wErr
	}
	cmd.Arg1 = arg1

	arg2, err := readStringU8(r)
	if err != nil {
		wMsg := fmt.Sprintf("Unable to parse argument position %d for %s, %s",
			2, cmd.Sender, err,
		)
		wErr := errgo.NewError(wMsg, globals.VERB_WRN)
		return nil, wErr
	}
	cmd.Arg2 = arg2

	arg3, err := readStringU8(r)
	if err != nil {
		wMsg := fmt.Sprintf("Unable to parse argument position %d for %s: %s",
			3, cmd.Sender, err,
		)
		wErr := errgo.NewError(wMsg, globals.VERB_WRN)
		return nil, wErr
	}
	cmd.Arg3 = arg3

	arg4, err := readStringU8(r)
	if err != nil {
		wMsg := fmt.Sprintf("Unable to parse argument position %d for %s: %s",
			4, cmd.Sender, err,
		)
		wErr := errgo.NewError(wMsg, globals.VERB_WRN)
		return nil, wErr
	}
	cmd.Arg4 = arg4

	return cmd, nil
}
