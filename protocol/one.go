package protocol

import (
	"bytes"
	"io"
	"mycelia/commands"
	"mycelia/global"
	"mycelia/str"
)

// -----------------------------------------------------------------------------
// Version 1 command decoding.
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
		return nil, ParseCommandErr
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
		cmd, err = nil, ParseCommandErr
	}

	if r.Len() != 0 {
		cmd = nil
		err = ParseCommandErr
	}

	return cmd, err
}

// Parses the header after version: obj_type, cmd_type, uid, route
func parseBaseHeader(r io.Reader, msg *Message) (*Message, error) {
	if err := readU8(r, &msg.ObjType); err != nil {
		return nil, err
	}

	if err := readU8(r, &msg.CmdType); err != nil {
		return nil, err
	}

	uid, err := readString(r)
	if err != nil {
		return nil, err
	}
	msg.UID = uid

	route, err := readString(r)
	if err != nil {
		return nil, err
	}
	msg.Route = route

	return msg, nil
}

func parseSubscriberMessage(r io.Reader, msg *Message) (commands.Command, error) {
	msg, err := parseRoutedMessage(r, msg)
	if err != nil {
		return nil, ParseCommandErr
	}
	switch msg.CmdType {
	case global.CMD_ADD:
		cmd := commands.NewSubscriber(
			msg.CmdType,
			msg.UID,
			msg.Route,
			msg.Channel,
			msg.Address,
		)
		return cmd, nil
	case global.CMD_REMOVE:
		str.WarningPrint("SUBSCRIBER CMD_REMOVE not yet implemented")
		return nil, ParseCommandErr
	default:
		return nil, ParseCommandErr
	}
}

func parseTransformerMessage(r io.Reader, msg *Message) (commands.Command, error) {
	msg, err := parseRoutedMessage(r, msg)
	if err != nil {
		return nil, ParseCommandErr
	}
	switch msg.CmdType {
	case global.CMD_ADD:
		cmd := commands.NewTransformer(
			msg.CmdType,
			msg.UID,
			msg.Route,
			msg.Channel,
			msg.Address,
		)
		return cmd, nil
	case global.CMD_REMOVE:
		str.WarningPrint("TRANSFORMER CMD_REMOVE not yet implemented")
		return nil, ParseCommandErr
	default:
		return nil, ParseCommandErr
	}
}

func parseRoutedMessage(r io.Reader, msg *Message) (*Message, error) {
	ch, err := readString(r)
	if err != nil {
		return nil, err
	}
	msg.Channel = ch

	addr, err := readString(r)
	if err != nil {
		return nil, err
	}
	msg.Address = addr

	return msg, nil
}

func parseSendMessage(r io.Reader, msg *Message) (commands.Command, error) {
	payload, err := readBytes(r)
	if err != nil {
		return nil, ParseCommandErr
	}
	if payload == nil {
		msg.Payload = []byte{}
	} else {
		msg.Payload = payload
	}
	if msg.CmdType != uint8(1) {
		return nil, ParseCommandErr
	}

	cmd := commands.NewDelivery(
		msg.CmdType,
		msg.UID,
		msg.Route,
		msg.Payload,
	)
	return cmd, nil
}
