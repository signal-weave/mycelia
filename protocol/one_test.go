package protocol

import (
	"encoding/binary"
	"strings"
	"testing"
)

func buildV1Bytes(obj, cmd uint8, uid, sender, a1, a2, a3, a4, payload string) []byte {
	var data []byte
	data = append(data, obj)
	data = append(data, cmd)
	data = append(data, encStrU8(uid)...)
	data = append(data, encStrU16(sender)...)
	data = append(data, encStrU8(a1)...)
	data = append(data, encStrU8(a2)...)
	data = append(data, encStrU8(a3)...)
	data = append(data, encStrU8(a4)...)
	data = append(data, encStrU16(payload)...)
	return data
}

func TestDecodeV1_Success(t *testing.T) {
	data := buildV1Bytes(
		1, 2,
		"uid-123",
		"127.0.0.1:5500",
		"a1", "a2", "a3", "a4",
		"hello, world",
	)

	cmd, err := decodeV1(data)
	if err != nil {
		t.Fatalf("decodeV1 returned error: %v", err)
	}
	if cmd == nil {
		t.Fatalf("decodeV1 returned nil command")
	}

	if cmd.ObjType != 1 || cmd.CmdType != 2 {
		t.Fatalf("unexpected header: got obj=%d cmd=%d", cmd.ObjType, cmd.CmdType)
	}
	if cmd.UID != "uid-123" {
		t.Fatalf("UID mismatch: got %q", cmd.UID)
	}
	if cmd.Sender != "127.0.0.1:5500" {
		t.Fatalf("Sender mismatch: got %q", cmd.Sender)
	}
	if cmd.Arg1 != "a1" || cmd.Arg2 != "a2" || cmd.Arg3 != "a3" || cmd.Arg4 != "a4" {
		t.Fatalf("Args mismatch: %q %q %q %q", cmd.Arg1, cmd.Arg2, cmd.Arg3, cmd.Arg4)
	}
	if string(cmd.Payload) != "hello, world" {
		t.Fatalf("Payload mismatch: got %q", cmd.Payload)
	}
}

func TestDecodeV1_Error_UnaccountedData(t *testing.T) {
	data := buildV1Bytes(
		3, 4,
		"u-1",
		"10.0.0.1:1234",
		"x", "y", "z", "w",
		"payload",
	)
	// Append an extra stray byte to force the "Unaccounted data in reader" path.
	data = append(data, 0xFF)

	cmd, err := decodeV1(data)
	if err == nil {
		t.Fatalf("expected error for unaccounted data, got nil")
	}
	if cmd != nil {
		t.Fatalf("expected nil command on error, got: %#v", cmd)
	}
	if !strings.Contains(err.Error(), "Unaccounted data in reader") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDecodeV1_Error_TruncatedPayload(t *testing.T) {
	// Build a valid header/args, then craft a payload with a u16 length
	// that is larger than the bytes we actually provide → should error.
	var data []byte
	data = append(data, uint8(7)) // ObjType
	data = append(data, uint8(8)) // CmdType
	data = append(data, encStrU8("trace-uid")...)
	data = append(data, encStrU16("192.168.0.2:7777")...)
	data = append(data, encStrU8("p1")...)
	data = append(data, encStrU8("p2")...)
	data = append(data, encStrU8("p3")...)
	data = append(data, encStrU8("p4")...)

	// Malformed payload: declare length 5, provide only 2 bytes ("hi").
	decl := make([]byte, 2)
	binary.BigEndian.PutUint16(decl, 5)
	data = append(data, decl...)
	data = append(data, []byte("hi")...)

	cmd, err := decodeV1(data)
	if err == nil {
		t.Fatalf("expected error due to truncated payload, got nil")
	}
	if cmd != nil {
		t.Fatalf("expected nil command on payload parse error, got: %#v", cmd)
	}
	if !strings.Contains(err.Error(), "Unable to parse payload") {
		t.Fatalf("unexpected error: %v", err)
	}
}
