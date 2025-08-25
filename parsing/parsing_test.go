package parsing

import (
	"encoding/binary"
	"testing"
)

// makeStream encodes tokens as [uvarint length][bytes]...
func makeStream(tokens ...string) []byte {
	buf := make([]byte, 0, 64)
	tmp := make([]byte, binary.MaxVarintLen64)
	for _, s := range tokens {
		n := binary.PutUvarint(tmp, uint64(len(s)))
		buf = append(buf, tmp[:n]...)
		buf = append(buf, []byte(s)...)
	}
	return buf
}

func TestParseTokens_Basic(t *testing.T) {
	in := makeStream("1", "SEND", "route", "payload")
	got, err := parseTokens(in)
	if err != nil {
		t.Fatalf("parseTokens returned error: %v", err)
	}
	want := []string{"1", "SEND", "route", "payload"}
	if len(got) != len(want) {
		t.Fatalf("len mismatch: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d mismatch: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestParseTokens_ZeroLength_Ignored(t *testing.T) {
	// Include a zero-length field (should be ignored) between real tokens
	// Stream: ["1", "", "CMD"]
	in := append(makeStream("1"), 0 /* uvarint(0) */)
	in = append(in, makeStream("CMD")...)
	got, err := parseTokens(in)
	if err != nil {
		t.Fatalf("parseTokens returned error: %v", err)
	}
	want := []string{"1", "CMD"}
	if len(got) != len(want) {
		t.Fatalf("len mismatch: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d mismatch: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestParseTokens_TruncatedBody(t *testing.T) {
	// Create a field that declares length 5 but only provide 3 bytes.
	tmp := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(tmp, 5)
	in := append([]byte{}, tmp[:n]...)
	in = append(in, []byte("abc")...) // truncated
	if _, err := parseTokens(in); err == nil {
		t.Fatalf("expected error for truncated body, got nil")
	}
}

func TestParseLine_BadBody_ReturnsUnknown(t *testing.T) {
	// Malformed: declare 3 bytes then provide none -> parseTokens fails.
	tmp := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(tmp, 3)
	in := append([]byte{}, tmp[:n]...)
	cmd, c := ParseLine(in)
	if cmd != unknownCommand || c != nil {
		t.Fatalf("expected (%q,nil) for bad body, got (%q,%v)", unknownCommand, cmd, c)
	}
}

func TestParseLine_MissingVersionOrCommand(t *testing.T) {
	// Only one token => missing command
	in := makeStream("1")
	cmd, c := ParseLine(in)
	if cmd != unknownCommand || c != nil {
		t.Fatalf("expected (%q,nil) for missing command, got (%q,%v)", unknownCommand, cmd, c)
	}
}

func TestParseLine_BadVersionToken(t *testing.T) {
	// First token not an integer
	in := makeStream("x", "ANY")
	cmd, c := ParseLine(in)
	if cmd != unknownCommand || c != nil {
		t.Fatalf("expected (%q,nil) for bad version token, got (%q,%v)", unknownCommand, cmd, c)
	}
}

func TestParseLine_UnsupportedVersion(t *testing.T) {
	// Version 2 should fall through to unknown (avoids calling parseDataV1)
	in := makeStream("2", "CMD")
	cmd, c := ParseLine(in)
	if cmd != unknownCommand || c != nil {
		t.Fatalf("expected (%q,nil) for unsupported version, got (%q,%v)", unknownCommand, cmd, c)
	}
}

func TestVerifyTokenLength(t *testing.T) {
	if !verifyTokenLength([]string{"a", "b"}, 2, "TEST") {
		t.Fatalf("expected true for matching arg length")
	}
	if verifyTokenLength([]string{"a"}, 2, "TEST") {
		t.Fatalf("expected false for mismatched arg length")
	}
}
