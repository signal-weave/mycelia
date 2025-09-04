package protocol

import (
	"fmt"
	"io"

	"mycelia/comm"
	"mycelia/errgo"
	"mycelia/globals"
	"mycelia/str"
)

// -----------------------------------------------------------------------------
// The primary parsing entry point.
// The main protocol version detection and parsing version handling.
// -----------------------------------------------------------------------------

// A response represents the ack value and the corresponding message's UID to
// send back to the producer.
type Response struct {
	AckType uint8
	UID     string
}

// An object is a struct that is decoded from the incoming byte stream and ran
// through the system.
//
// Objects contain a com.ConnResponder for communicating with the sender and a
// protocol.Response for encoding a response with ack status and corresponding
// UID to send to the sender via the object.Responder.
type Object struct {
	Responder *comm.ConnResponder
	Response  *Response

	ObjType uint8
	CmdType uint8
	AckPlcy uint8

	UID string

	Arg1, Arg2 string
	Arg3, Arg4 string

	Payload []byte
}

func NewObject(
	objType, cmdType, AckPlcy uint8,
	uid, arg1, arg2, arg3, arg4 string,
	payload []byte) *Object {

	return &Object{
		Response: &Response{
			UID:     uid,
			AckType: globals.ACK_TYPE_UNKNOWN,
		},

		ObjType: objType,
		CmdType: cmdType,
		AckPlcy: AckPlcy,
		UID:     uid,
		Arg1:    arg1,
		Arg2:    arg2,
		Arg3:    arg3,
		Arg4:    arg4,
		Payload: payload,
	}
}

// Prints each field on the object...
func (obj *Object) PrintValues() {
	str.PrintAsciiLine()
	fmt.Println("ObjType:", obj.ObjType)
	fmt.Println("CmdType:", obj.CmdType)

	fmt.Println("ReturnAddress:", obj.Responder.C.RemoteAddr().String())
	fmt.Println("UID:", obj.UID)

	fmt.Println("Arg1:", obj.Arg1)
	fmt.Println("Arg2:", obj.Arg2)
	fmt.Println("Arg3:", obj.Arg3)
	fmt.Println("Arg4:", obj.Arg4)

	fmt.Println("Payload:", string(obj.Payload))
	str.PrintAsciiLine()
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

func DecodeFrame(line []byte, resp *comm.ConnResponder) (*Object, error) {
	version, rest, err := parseProtoVer(line)
	if err != nil {
		wMsg := fmt.Sprintf("Read protocol version: %v", err)
		wErr := errgo.NewError(wMsg, globals.VERB_WRN)
		return nil, wErr
	}

	// The broker always works off of the same types of objects.
	// Message objects may evolve over time, adding new fields for new
	// functionality, but the broker should remain compatible with previous
	// client side API versions.

	// If a client is using API ver 1 to communicate with Broker ver 2, then the
	// client should be able to still communicate.
	// This first token of a message is the API version, and this switch runs
	// the corresponding parsing logic.

	// This is mainly because early on there was uncertainty if the protocol and
	// object structure was done right, and we reserved the ability to update
	// it as we go.
	switch version {
	case 1:
		return decodeV1(rest, resp)
	default:
		wErr := errgo.NewError("Unable to parse object!", globals.VERB_WRN)
		return nil, wErr
	}
}
