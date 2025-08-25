package parsing

import (
	"testing"

	"mycelia/commands"
)

func TestParseLine_V1_SendMessage_OK(t *testing.T) {
	// Body includes a space (escaped) to ensure unescapeTokens is used.
	line := []byte("1;;send_message;;id-1;;orders.created;;hello%20world\r\n")

	typ, cmd := ParseLine(line)
	if typ != "send_message" {
		t.Fatalf("type mismatch: want %q, got %q", "send_message", typ)
	}
	if cmd == nil {
		t.Fatalf("expected non-nil cmd")
	}

	sm, ok := cmd.(*commands.SendMessage)
	if !ok {
		t.Fatalf("expected *commands.SendMessage, got %T", cmd)
	}
	if sm.ID != "id-1" || sm.Route != "orders.created" || sm.Body != "hello world" {
		t.Fatalf("fields mismatch: %+v", *sm)
	}
}

func TestParseLine_V1_SendMessage_WithEscapedDelims_OK(t *testing.T) {
	// Body contains an escaped delimiter "%3B%3B" which must NOT split.
	line := []byte("1;;send_message;;id-2;;orders.created;;foo%3B%3Bbar\n")

	typ, cmd := ParseLine(line)
	if typ != "send_message" || cmd == nil {
		t.Fatalf("want send_message non-nil, got typ=%q cmd=%v", typ, cmd)
	}
	sm := cmd.(*commands.SendMessage)
	if sm.Body != "foo;;bar" {
		t.Fatalf("decoded body mismatch: %q", sm.Body)
	}
}

func TestParseLine_UnknownVersion_ReturnsErrNil(t *testing.T) {
	line := []byte("2;;send_message;;id-3;;r;;b\n") // version 2 unsupported
	typ, cmd := ParseLine(line)
	if typ != "err" || cmd != nil {
		t.Fatalf("want typ=err and cmd=nil, got typ=%q cmd=%v", typ, cmd)
	}
}

func TestParseLine_InvalidVersion_ReturnsErrNil(t *testing.T) {
	line := []byte("x;;send_message;;id-4;;r;;b\n") // non-numeric version
	typ, cmd := ParseLine(line)
	if typ != "err" || cmd != nil {
		t.Fatalf("want typ=err and cmd=nil, got typ=%q cmd=%v", typ, cmd)
	}
}

func TestParseLine_MissingVersionOrCommand_ReturnsErrNil(t *testing.T) {
	// Only one token ("1") -> len(parts) < 2
	line := []byte("1\n")
	typ, cmd := ParseLine(line)
	if typ != "err" || cmd != nil {
		t.Fatalf("want typ=err and cmd=nil, got typ=%q cmd=%v", typ, cmd)
	}
}

func TestParseLine_UnescapeError_ReturnsErrNil(t *testing.T) {
	// Bad percent-escape in body: "%2G" is invalid hex.
	line := []byte("1;;send_message;;id-5;;r;;bad%2Gvalue\n")
	typ, cmd := ParseLine(line)
	if typ != "err" || cmd != nil {
		t.Fatalf("want typ=err and cmd=nil, got typ=%q cmd=%v", typ, cmd)
	}
}

func TestParseLine_V1_UnknownCommandType_PassesThroughEmptyAndNil(t *testing.T) {
	// parseDataV1 returns "", nil when the command type is unrecognized.
	line := []byte("1;;totally_unknown;;id-6\n")
	typ, cmd := ParseLine(line)
	if typ != "" || cmd != nil {
		t.Fatalf("want empty type and nil cmd, got typ=%q cmd=%v", typ, cmd)
	}
}
