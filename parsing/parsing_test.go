package parsing

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"mycelia/global"
)

// -------Test helpers (writers)------------------------------------------------

func wrU8(buf *bytes.Buffer, v uint8, t *testing.T) {
	t.Helper()
	if err := binary.Write(buf, binary.BigEndian, v); err != nil {
		t.Fatalf("wrU32 failed: %v", err)
	}
}

func wrU32(buf *bytes.Buffer, v uint32, t *testing.T) {
	t.Helper()
	if err := binary.Write(buf, binary.BigEndian, v); err != nil {
		t.Fatalf("wrU32 failed: %v", err)
	}
}

func wrString32(buf *bytes.Buffer, s string, t *testing.T) {
	t.Helper()
	wrU32(buf, uint32(len(s)), t)
	if _, err := buf.Write([]byte(s)); err != nil {
		t.Fatalf("wrString32 failed: %v", err)
	}
}

func wrBytes32(buf *bytes.Buffer, b []byte, t *testing.T) {
	t.Helper()
	wrU32(buf, uint32(len(b)), t)
	if _, err := buf.Write(b); err != nil {
		t.Fatalf("wrBytes32 failed: %v", err)
	}
}

// write an unsigned varint (LEB128) into buf.
func wrUvarint(buf *bytes.Buffer, x uint64) {
	var tmp [10]byte
	n := binary.PutUvarint(tmp[:], x)
	buf.Write(tmp[:n])
}

// -------parseProtoVer---------------------------------------------------------

func TestParseProtoVer_OK(t *testing.T) {
	var buf bytes.Buffer
	wrU8(&buf, 1, t)                    // version
	buf.Write([]byte{0xAA, 0xBB, 0xCC}) // rest

	ver, rest, err := parseProtoVer(buf.Bytes())
	if err != nil {
		t.Fatalf("parseProtoVer error: %v", err)
	}
	if ver != 1 {
		t.Fatalf("expected ver=1, got %d", ver)
	}
	if !bytes.Equal(rest, []byte{0xAA, 0xBB, 0xCC}) {
		t.Fatalf("unexpected rest: %v", rest)
	}
}

func TestParseProtoVer_ShortBuffer(t *testing.T) {
	data := []byte{} // < 1 byte needed for uint8 version
	_, _, err := parseProtoVer(data)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}
}

// -------ParseLine (dispatch to v1)--------------------------------------------

func TestParseLine_V1_Success_SendMessage(t *testing.T) {
	const uid = "uid-123"
	const route = "topic.route"
	payload := []byte{0xDE, 0xAD, 0xBE, 0xEF}

	// Build v1 body that decodeV1 understands:
	// [obj_type][cmd_type][uid][route][payload]
	var body bytes.Buffer
	wrU8(&body, global.OBJ_DELIVERY, t)
	wrU8(&body, global.CMD_SEND, t)
	wrString32(&body, uid, t)
	wrString32(&body, route, t)
	wrBytes32(&body, payload, t)

	// Prefix with version
	var packet bytes.Buffer
	wrU8(&packet, 1, t)
	packet.Write(body.Bytes())

	cmd, err := ParseLine(packet.Bytes())
	if err != nil {
		t.Fatalf("ParseLine error: %v", err)
	}
	if cmd == nil {
		t.Fatalf("ParseLine returned nil command")
	}

	// We don't depend on concrete type;
	// just ensure it's the SendMessage variant.
	typeName := reflect.TypeOf(cmd).String()
	if !strings.Contains(typeName, "Delivery") {
		t.Fatalf(
			"expected command type to contain 'Delivery', got %q", typeName,
		)
	}
}

func TestParseLine_UnsupportedVersion(t *testing.T) {
	// version 999, empty body
	var packet bytes.Buffer
	wrU32(&packet, 999, t)

	cmd, err := ParseLine(packet.Bytes())
	if !errors.Is(err, ParseCommandErr) {
		t.Fatalf("expected ParseCommandErr, got %v", err)
	}
	if cmd != nil {
		t.Fatalf("expected nil cmd on error, got %T", cmd)
	}
}

func TestParseLine_VersionReadError(t *testing.T) {
	// <4 bytes total
	data := []byte{0x01, 0x02, 0x03}
	cmd, err := ParseLine(data)
	if !errors.Is(err, ParseCommandErr) {
		t.Fatalf("expected ParseCommandErr, got %v", err)
	}
	if cmd != nil {
		t.Fatalf("expected nil cmd, got %T", cmd)
	}
}

// -------parseTokens-----------------------------------------------------------

func TestParseTokens_Normal(t *testing.T) {
	var buf bytes.Buffer
	wrUvarint(&buf, 3)
	buf.WriteString("foo")
	wrUvarint(&buf, 3)
	buf.WriteString("bar")

	out, err := parseTokens(buf.Bytes())
	if err != nil {
		t.Fatalf("parseTokens error: %v", err)
	}
	want := []string{"foo", "bar"}
	if !reflect.DeepEqual(out, want) {
		t.Fatalf("expected %v, got %v", want, out)
	}
}

func TestParseTokens_SkipZeroLength(t *testing.T) {
	var buf bytes.Buffer
	wrUvarint(&buf, 3)
	buf.WriteString("foo")
	wrUvarint(&buf, 0) // should be skipped
	wrUvarint(&buf, 3)
	buf.WriteString("bar")

	out, err := parseTokens(buf.Bytes())
	if err != nil {
		t.Fatalf("parseTokens error: %v", err)
	}
	want := []string{"foo", "bar"}
	if !reflect.DeepEqual(out, want) {
		t.Fatalf("expected %v, got %v", want, out)
	}
}

func TestParseTokens_TruncatedFieldBody(t *testing.T) {
	var buf bytes.Buffer
	// Declare length 5, provide only 4 bytes.
	wrUvarint(&buf, 5)
	buf.Write([]byte{'h', 'e', 'l', 'l'})

	_, err := parseTokens(buf.Bytes())
	if err == nil || !strings.Contains(err.Error(), "truncated field body") {
		t.Fatalf("expected truncated body error, got %v", err)
	}
}

func TestParseTokens_OverflowVarint(t *testing.T) {
	var buf bytes.Buffer
	// Craft an overflowing varint (>10 bytes with continuation bits set)
	for i := 0; i < 11; i++ {
		buf.WriteByte(0xFF)
	}

	_, err := parseTokens(buf.Bytes())
	if err == nil {
		t.Fatalf("expected overflow/varint error, got nil")
	}
}

// -------verifyTokenLength-----------------------------------------------------

func TestVerifyTokenLength(t *testing.T) {
	ok := verifyTokenLength([]string{"a", "b"}, 2, "CMD")
	if !ok {
		t.Fatalf("expected true for matching length")
	}
	ok = verifyTokenLength([]string{"a"}, 2, "CMD")
	if ok {
		t.Fatalf("expected false for mismatched length")
	}
}
