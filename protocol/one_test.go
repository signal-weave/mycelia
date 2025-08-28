package protocol

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"strings"
	"testing"

	"mycelia/global"
)

// Helpers to build length-prefixed (uint32 big-endian) strings/bytes.
func writeU8(buf *bytes.Buffer, v uint8, t *testing.T) {
	t.Helper()
	if err := binary.Write(buf, binary.BigEndian, v); err != nil {
		t.Fatalf("writeU32 failed: %v", err)
	}
}

// Helpers to build length-prefixed (uint32 big-endian) strings/bytes.
func writeU32(buf *bytes.Buffer, v uint32, t *testing.T) {
	t.Helper()
	if err := binary.Write(buf, binary.BigEndian, v); err != nil {
		t.Fatalf("writeU32 failed: %v", err)
	}
}

func writeString(buf *bytes.Buffer, s string, t *testing.T) {
	t.Helper()
	writeU32(buf, uint32(len(s)), t)
	if _, err := buf.Write([]byte(s)); err != nil {
		t.Fatalf("writeString failed: %v", err)
	}
}

func writeBytes(buf *bytes.Buffer, b []byte, t *testing.T) {
	t.Helper()
	writeU32(buf, uint32(len(b)), t)
	if _, err := buf.Write(b); err != nil {
		t.Fatalf("writeBytes failed: %v", err)
	}
}

func typeName(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	return reflect.TypeOf(v).String()
}

func TestDecodeV1_SendMessage_Success(t *testing.T) {
	uid := "uid-123"
	route := "topic.route"
	payload := []byte{0xDE, 0xAD, 0xBE, 0xEF}

	var buf bytes.Buffer
	// Header: obj_type, cmd_type, uid, route
	writeU8(&buf, global.OBJ_DELIVERY, t)
	writeU8(&buf, global.CMD_SEND, t)
	writeString(&buf, uid, t)
	writeString(&buf, route, t)
	// Body: payload
	writeBytes(&buf, payload, t)

	cmd, err := decodeV1(buf.Bytes())
	if err != nil {
		t.Fatalf("decodeV1 returned error: %v", err)
	}
	if cmd == nil {
		t.Fatalf("decodeV1 returned nil command")
	}

	tn := typeName(cmd)
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
	writeU8(&buf, global.OBJ_SUBSCRIBER, t)
	writeU8(&buf, global.CMD_ADD, t)
	writeString(&buf, uid, t)
	writeString(&buf, route, t)
	// Body: channel, address
	writeString(&buf, channel, t)
	writeString(&buf, address, t)

	cmd, err := decodeV1(buf.Bytes())
	if err != nil {
		t.Fatalf("decodeV1 returned error: %v", err)
	}
	if cmd == nil {
		t.Fatalf("decodeV1 returned nil command")
	}

	tn := typeName(cmd)
	if !strings.Contains(tn, "Subscriber") {
		t.Fatalf("expected command type to contain 'Subscriber', got %q", tn)
	}
}

func TestDecodeV1_UnknownObjType_Error(t *testing.T) {
	uid := "abc"
	route := "foo"

	var buf bytes.Buffer
	writeU32(&buf, 99, t) // unknown obj_type
	writeU8(&buf, global.CMD_ADD, t)
	writeString(&buf, uid, t)
	writeString(&buf, route, t)

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
	writeU8(&buf, global.OBJ_DELIVERY, t)
	writeU8(&buf, global.CMD_SEND, t)
	writeString(&buf, uid, t)
	writeString(&buf, route, t)
	writeBytes(&buf, payload, t)

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
	writeU8(&buf, global.OBJ_TRANSFORMER, t)
	writeU8(&buf, global.CMD_ADD, t)
	writeString(&buf, uid, t)
	writeString(&buf, route, t)
	writeString(&buf, channel, t)
	writeString(&buf, address, t)

	cmd, err := decodeV1(buf.Bytes())
	if err != nil {
		t.Fatalf("decodeV1 returned error: %v", err)
	}
	if cmd == nil {
		t.Fatalf("decodeV1 returned nil command")
	}
	if !strings.Contains(typeName(cmd), "Transformer") {
		t.Fatalf("expected transformer add command, got %T", cmd)
	}
}
