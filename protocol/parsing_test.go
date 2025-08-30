package protocol

import (
	"bytes"
	"strings"
	"testing"
)

// -------helpers to build a V1 frame (after version byte)----------------------

func buildV1AfterVersion(
	obj, cmd uint8,
	uid, sender, a1, a2, a3, a4, payload string) []byte {
	var data []byte
	data = append(data, obj)                   // ObjType
	data = append(data, cmd)                   // CmdType
	data = append(data, encStrU8(uid)...)      // UID (u8-len)
	data = append(data, encStrU16(sender)...)  // Sender (u16-len)
	data = append(data, encStrU8(a1)...)       // Arg1 (u8-len)
	data = append(data, encStrU8(a2)...)       // Arg2
	data = append(data, encStrU8(a3)...)       // Arg3
	data = append(data, encStrU8(a4)...)       // Arg4
	data = append(data, encStrU16(payload)...) // Payload (u16-len)
	return data
}

// -------parseProtoVer---------------------------------------------------------

func TestParseProtoVer_Valid(t *testing.T) {
	line := []byte{1, 0xAA, 0xBB, 0xCC}
	ver, rest, err := parseProtoVer(line)
	if err != nil {
		t.Fatalf("parseProtoVer error: %v", err)
	}
	if ver != 1 {
		t.Fatalf("want version 1, got %d", ver)
	}
	if !bytes.Equal(rest, []byte{0xAA, 0xBB, 0xCC}) {
		t.Fatalf("rest mismatch: %v", rest)
	}
}

func TestParseProtoVer_Short(t *testing.T) {
	_, _, err := parseProtoVer(nil)
	if err == nil {
		t.Fatalf("expected error on empty input, got nil")
	}
}

// -------ParseLine-------------------------------------------------------------

func TestParseLine_V1_Success(t *testing.T) {
	// Build full line: [version=1][v1 payload...]
	body := buildV1AfterVersion(
		5, 9,
		"uid-123",
		"127.0.0.1:5500",
		"a1", "a2", "a3", "a4",
		"hello",
	)
	line := append([]byte{1}, body...)

	cmd, err := ParseLine(line)
	if err != nil {
		t.Fatalf("ParseLine returned error: %v", err)
	}
	if cmd == nil {
		t.Fatalf("ParseLine returned nil *Command")
	}

	if cmd.ObjType != 5 || cmd.CmdType != 9 {
		t.Fatalf("header mismatch: obj=%d cmd=%d", cmd.ObjType, cmd.CmdType)
	}
	if cmd.UID != "uid-123" {
		t.Fatalf("UID mismatch: %q", cmd.UID)
	}
	if cmd.Sender != "127.0.0.1:5500" {
		t.Fatalf("Sender mismatch: %q", cmd.Sender)
	}
	if cmd.Arg1 != "a1" || cmd.Arg2 != "a2" || cmd.Arg3 != "a3" ||
		cmd.Arg4 != "a4" {
		t.Fatalf(
			"args mismatch: %q %q %q %q", cmd.Arg1, cmd.Arg2,
			cmd.Arg3, cmd.Arg4,
		)
	}
	if string(cmd.Payload) != "hello" {
		t.Fatalf("payload mismatch: %q", cmd.Payload)
	}
}

func TestParseLine_UnknownVersion(t *testing.T) {
	// Version 2 is not supported; any trailing bytes are irrelevant.
	line := []byte{2, 0xDE, 0xAD, 0xBE, 0xEF}
	cmd, err := ParseLine(line)
	if err == nil {
		t.Fatalf("expected error for unknown version, got nil")
	}
	if cmd != nil {
		t.Fatalf("expected nil *Command on error, got %#v", cmd)
	}
	if !strings.Contains(err.Error(), "Unable to parse command!") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseLine_VersionReadErrorWrapped(t *testing.T) {
	// Empty input → parseProtoVer returns io.ErrUnexpectedEOF;
	// ParseLine wraps it.
	_, err := ParseLine(nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Read protocol version") {
		t.Fatalf("expected wrapper message, got: %v", err)
	}
}
