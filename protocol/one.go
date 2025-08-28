package protocol

import (
	"bytes"
	"fmt"
	"io"
	"mycelia/commands"
	"mycelia/errgo"
	"mycelia/global"
)

// -----------------------------------------------------------------------------
// Version 1 command decoding.
// -----------------------------------------------------------------------------
// *Note that this is a messaging protocol, not a file transfer protocol
// -----------------------------------------------------------------------------
//           fields
// --------------------------
// protocol_ver  |  u8
// obj_type      |  u8
// obj_cmd       |  u8
// uid           |  u32 + len
// route         |  u32 + len
// --------------------------
// channel       |  u32 + len
// address       |  u32 + len
// --------------------------
// payload       |  u32 + len
// -----------------------------------------------------------------------------
// The version 1 protocol looks as follows:

// # Fixed field sized header
// +---------+--------+-------------+-------------+
// | u32 len | u8 ver | u8 obj_type | u8 cmd_type |
// +---------+--------+-------------+-------------+

// which is then followed by a variable field sized sub-header

// # Routing Sub-header
// +-------------+---------------+
// | u32 len uid | u32 len route |
// +-------------+---------------+

// which is then followed by one of the following bodies:

// # Subscriber + Transformer Body
// +--------------+--------------+
// | u32 len chan | u32 len addr |
// +--------------+--------------+

// # Message Body
// +-----------------+
// | u32 len payload |
// +-----------------+
// With the Message Body payload being the data finally forwarded to
// subscribers.
// -----------------------------------------------------------------------------

type Message struct {
	ObjType uint8
	CmdType uint8
	UID     string
	Route   string

	// Optional fields depending on ObjType
	Channel string // Subscriber + Transformer
	Address string // Subscriber + Transformer
	Payload []byte // Message
}

func decodeV1(data []byte) (commands.Command, error) {
	r := bytes.NewReader(data)
	msg := &Message{}
	msg, err := parseBaseHeader(r, msg)
	if err != nil {
		return nil, err
	}
	msg, err = parseSubHeader(r, msg)
	if err != nil {
		return nil, err
	}

	var cmd commands.Command

	switch msg.ObjType {
	case global.OBJ_TRANSFORMER:
		cmd, err = parseTransformerMessage(r, msg)
	case global.OBJ_SUBSCRIBER:
		cmd, err = parseSubscriberMessage(r, msg)
	case global.OBJ_DELIVERY:
		cmd, err = parseSendMessage(r, msg)
	default:
		wMsg := fmt.Sprintf("Unknown object yet %s", string(msg.ObjType))
		wErr := errgo.NewError(wMsg, global.VERB_WRN)
		cmd, err = nil, wErr
	}

	if r.Len() != 0 {
		cmd = nil
		err = errgo.NewError("Unaccounted data in reader", global.VERB_WRN)
	}

	return cmd, err
}

// Parses the header after version: obj_type, cmd_type, uid, route
func parseBaseHeader(r io.Reader, msg *Message) (*Message, error) {
	if err := readU8(r, &msg.ObjType); err != nil {
		wMsg := fmt.Sprintf(
			"Unable to parse u8 ObjType field from %s.", msg.Address,
		)
		wErr := errgo.NewError(wMsg, global.VERB_WRN)
		return nil, wErr
	}

	if err := readU8(r, &msg.CmdType); err != nil {
		wMsg := fmt.Sprintf(
			"Unable to parse u8 CmdType field from %s.", msg.Address,
		)
		wErr := errgo.NewError(wMsg, global.VERB_WRN)
		return nil, wErr
	}

	return msg, nil
}

func parseSubHeader(r io.Reader, msg *Message) (*Message, error) {
	uid, err := readString(r)
	if err != nil {
		wMsg := fmt.Sprintf(
			"Unable to parse string UID field from %s.", msg.Address,
		)
		wErr := errgo.NewError(wMsg, global.VERB_WRN)
		return nil, wErr
	}
	msg.UID = uid

	route, err := readString(r)
	if err != nil {
		wMsg := fmt.Sprintf(
			"Unable to parse string Route field from %s.", msg.Address,
		)
		wErr := errgo.NewError(wMsg, global.VERB_WRN)
		return nil, wErr
	}
	msg.Route = route

	return msg, nil
}

func parseSubscriberMessage(r io.Reader, msg *Message) (commands.Command, error) {
	msg, err := parseRoutedMessage(r, msg)
	if err != nil {
		return nil, err
	}
	if msg.CmdType != global.CMD_ADD && msg.CmdType != global.CMD_REMOVE {
		wErr := errgo.NewError("Invalid command code!", global.VERB_WRN)
		return nil, wErr
	}

	cmd := commands.NewSubscriber(
		msg.CmdType,
		msg.UID,
		msg.Route,
		msg.Channel,
		msg.Address,
	)

	return cmd, nil
}

func parseTransformerMessage(r io.Reader, msg *Message) (commands.Command, error) {
	msg, err := parseRoutedMessage(r, msg)
	if err != nil {
		return nil, err
	}
	if msg.CmdType != global.CMD_ADD && msg.CmdType != global.CMD_REMOVE {
		wErr := errgo.NewError("Invalid command code!", global.VERB_WRN)
		return nil, wErr
	}

	cmd := commands.NewTransformer(
		msg.CmdType,
		msg.UID,
		msg.Route,
		msg.Channel,
		msg.Address,
	)

	return cmd, nil
}

func parseRoutedMessage(r io.Reader, msg *Message) (*Message, error) {
	ch, err := readString(r)
	if err != nil {
		wMsg := fmt.Sprintf("Unable to parse channel name from %s", msg.Address)
		wErr := errgo.NewError(wMsg, global.VERB_WRN)
		return nil, wErr
	}
	msg.Channel = ch

	addr, err := readString(r)
	if err != nil {
		wMsg := fmt.Sprintf("Unable to parse address from %s", msg.Address)
		wErr := errgo.NewError(wMsg, global.VERB_WRN)
		return nil, wErr
	}
	msg.Address = addr

	return msg, nil
}

func parseSendMessage(r io.Reader, msg *Message) (commands.Command, error) {
	payload, err := readBytes(r)
	if err != nil {
		wMsg := fmt.Sprintf("Unable to parse payload len for %s", msg.Address)
		wErr := errgo.NewError(wMsg, global.VERB_WRN)
		return nil, wErr
	}
	if payload == nil {
		msg.Payload = []byte{}
	} else {
		msg.Payload = payload
	}
	if msg.CmdType != global.CMD_SEND {
		wErr := errgo.NewError("Invalid command code!", global.VERB_WRN)
		return nil, wErr
	}

	cmd := commands.NewDelivery(
		msg.CmdType,
		msg.UID,
		msg.Route,
		msg.Payload,
	)
	return cmd, nil
}
