package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"strings"
	"testing"

	"mycelia/commands"
	"mycelia/global"
	"mycelia/test"
)

func TestDecodeV1_SendMessage_Success(t *testing.T) {
	uid := "uid-123"
	route := "topic.route"
	payload := []byte{0xDE, 0xAD, 0xBE, 0xEF}

	var buf bytes.Buffer
	// Header: obj_type, cmd_type, uid, route
	test.WriteU8(&buf, global.OBJ_DELIVERY, t)
	test.WriteU8(&buf, global.CMD_SEND, t)
	test.WriteString(&buf, uid, t)
	test.WriteString(&buf, route, t)
	// Body: payload
	test.WriteBytes(&buf, payload, t)

	cmd, err := decodeV1(buf.Bytes())
	if err != nil {
		t.Fatalf("decodeV1 returned error: %v", err)
	}
	if cmd == nil {
		t.Fatalf("decodeV1 returned nil command")
	}

	tn := test.TypeName(cmd)
	// Be flexible in case the exact package/type name differs slightly.
	if !strings.Contains(tn, "Delivery") {
		t.Fatalf("expected command type to contain 'Delivery', got %q", tn)
	}
}

func TestDecodeV1_AddSubscriber_Success(t *testing.T) {
	uid := "user-42"
	route := "router.alpha"
	channel := "events"
	address := "127.0.0.1:5555"

	var buf bytes.Buffer
	// Header: obj_type, cmd_type, uid, route
	test.WriteU8(&buf, global.OBJ_SUBSCRIBER, t)
	test.WriteU8(&buf, global.CMD_ADD, t)
	test.WriteString(&buf, uid, t)
	test.WriteString(&buf, route, t)
	// Body: channel, address
	test.WriteString(&buf, channel, t)
	test.WriteString(&buf, address, t)

	cmd, err := decodeV1(buf.Bytes())
	if err != nil {
		t.Fatalf("decodeV1 returned error: %v", err)
	}
	if cmd == nil {
		t.Fatalf("decodeV1 returned nil command")
	}

	tn := test.TypeName(cmd)
	if !strings.Contains(tn, "Subscriber") {
		t.Fatalf("expected command type to contain 'Subscriber', got %q", tn)
	}
}

func TestDecodeV1_UnknownObjType_Error(t *testing.T) {
	uid := "abc"
	route := "foo"

	var buf bytes.Buffer
	test.WriteU32(&buf, 99, t) // unknown obj_type
	test.WriteU8(&buf, global.CMD_ADD, t)
	test.WriteString(&buf, uid, t)
	test.WriteString(&buf, route, t)

	cmd, err := decodeV1(buf.Bytes())
	if err == nil {
		t.Fatalf("expected error, got nil (cmd=%T)", cmd)
	}
	if !strings.Contains(err.Error(), "Unable to parse string UID field") {
		t.Fatalf("expected ParseCommandErr, got %v", err)
	}
	if cmd != nil {
		t.Fatalf("expected nil cmd on error, got %T", cmd)
	}
}

func TestDecodeV1_TrailingBytes_Error(t *testing.T) {
	uid := "uid-123"
	route := "topic.route"
	payload := []byte{0x01, 0x02, 0x03}

	var buf bytes.Buffer
	test.WriteU8(&buf, global.OBJ_DELIVERY, t)
	test.WriteU8(&buf, global.CMD_SEND, t)
	test.WriteString(&buf, uid, t)
	test.WriteString(&buf, route, t)
	test.WriteBytes(&buf, payload, t)

	// Append a stray byte to trigger trailing-bytes check.
	buf.WriteByte(0x00)

	cmd, err := decodeV1(buf.Bytes())
	if err == nil {
		t.Fatalf("expected error due to trailing bytes, got nil (cmd=%T)", cmd)
	}
	if cmd != nil {
		t.Fatalf("expected nil cmd on error, got %T", cmd)
	}
}

func TestDecodeV1_AddTransformer_Success(t *testing.T) {
	uid := "u-t"
	route := "r-t"
	channel := "ch-t"
	address := "addr-t"

	var buf bytes.Buffer
	test.WriteU8(&buf, global.OBJ_TRANSFORMER, t)
	test.WriteU8(&buf, global.CMD_ADD, t)
	test.WriteString(&buf, uid, t)
	test.WriteString(&buf, route, t)
	test.WriteString(&buf, channel, t)
	test.WriteString(&buf, address, t)

	cmd, err := decodeV1(buf.Bytes())
	if err != nil {
		t.Fatalf("decodeV1 returned error: %v", err)
	}
	if cmd == nil {
		t.Fatalf("decodeV1 returned nil command")
	}
	if !strings.Contains(test.TypeName(cmd), "Transformer") {
		t.Fatalf("expected transformer add command, got %T", cmd)
	}
}

func TestDecodeV1_Globals(t *testing.T) {
	// Build a Globals struct and JSON encode it.
	want := commands.NewGlobals(
		global.CMD_UPDATE,
		"uid-123",
		"0.0.0.0",
		"30s",
		8080,
		2,
		true,
	)
	payload, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	// Helper: write a length-prefixed string.
	WriteString := func(buf *bytes.Buffer, s string) {
		binary.Write(buf, binary.BigEndian, uint32(len(s)))
		buf.WriteString(s)
	}
	// Helper: write a length-prefixed []byte.
	WriteBytes := func(buf *bytes.Buffer, b []byte) {
		binary.Write(buf, binary.BigEndian, uint32(len(b)))
		buf.Write(b)
	}

	buf := new(bytes.Buffer)
	// Base header: ObjType, CmdType
	buf.WriteByte(byte(global.OBJ_GLOBALS))
	buf.WriteByte(byte(global.CMD_UPDATE))

	WriteString(buf, want.ID) // Sub-header: UID
	WriteBytes(buf, payload)  // Body: payload

	cmd, err := decodeV1(buf.Bytes())
	if err != nil {
		t.Fatalf("decodeV1 returned error: %v", err)
	}
	globalsCmd, ok := cmd.(*commands.Globals)
	if !ok {
		t.Fatalf("decodeV1 returned %T, want *commands.Globals", cmd)
	}

	if globalsCmd.Port != want.Port ||
		globalsCmd.Address != want.Address ||
		globalsCmd.Verbosity != want.Verbosity ||
		globalsCmd.PrintTree != want.PrintTree ||
		globalsCmd.TransformTimeout != want.TransformTimeout {
		t.Errorf("decoded Globals %+v does not match want %+v", globalsCmd, want)
	}
}
