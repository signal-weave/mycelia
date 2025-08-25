package parsing

import (
	"testing"

	"mycelia/boot"
)

var _ = boot.RuntimeCfg // REQUIRED for global config values.

// ------splitTokens------------------------------------------------------------

func TestSplitTokens_Basic(t *testing.T) {
	in := []byte("a;;b;;c")
	got := splitTokens(in)

	if len(got) != 3 {
		t.Fatalf("want 3 tokens, got %d", len(got))
	}
	if string(got[0]) != "a" || string(got[1]) != "b" || string(got[2]) != "c" {
		t.Fatalf("tokens mismatch: %q, %q, %q", got[0], got[1], got[2])
	}
}

func TestSplitTokens_TrimCRLF(t *testing.T) {
	in := []byte("a;;b;;c\r\n")
	got := splitTokens(in)

	if len(got) != 3 {
		t.Fatalf("want 3 tokens, got %d", len(got))
	}
	if string(got[2]) != "c" {
		t.Fatalf("expected last token 'c', got %q", got[2])
	}
}

func TestSplitTokens_EmptyFieldsAndTrailingDelim(t *testing.T) {
	// "a;;;;b;;" -> ["a", "", "b", ""]
	in := []byte("a;;;;b;;")
	got := splitTokens(in)

	if len(got) != 4 {
		t.Fatalf("want 4 tokens, got %d", len(got))
	}
	if string(got[0]) != "a" || string(got[1]) != "" || string(got[2]) != "b" || string(got[3]) != "" {
		t.Fatalf("tokens mismatch: %q", got)
	}
}

func TestSplitTokens_ZeroCopyAlias(t *testing.T) {
	in := []byte("x;;y")
	toks := splitTokens(in)

	// Prove aliasing: the first token's first byte shares memory with 'in[0]'.
	if &toks[0][0] != &in[0] {
		t.Fatalf("expected zero-copy aliasing between tokens and input")
	}
}

// ------unescape & unescapeTokens----------------------------------------------

func TestUnescape_Success(t *testing.T) {
	out, err := unescape([]byte("hello%20world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello world" {
		t.Fatalf("want 'hello world', got %q", out)
	}
}

func TestUnescapeTokens_Success(t *testing.T) {
	args := [][]byte{
		[]byte("a%2Fb"),     // "a/b"
		[]byte("c%20d"),     // "c d"
		[]byte("%E2%98%83"), // unicode snowman "☃"
	}
	out, err := unescapeTokens(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 3 || out[0] != "a/b" || out[1] != "c d" || out[2] != "☃" {
		t.Fatalf("decoded tokens mismatch: %#v", out)
	}
}

func TestUnescapeTokens_ErrorPropagates(t *testing.T) {
	// Invalid percent-escape: '2G' is not hex.
	args := [][]byte{[]byte("bad%2Gvalue")}
	out, err := unescapeTokens(args)
	if err == nil {
		t.Fatalf("expected error for invalid escape, got nil (out=%#v)", out)
	}
}

// ------split + unescape combined sanity---------------------------------------

func TestSplitThenUnescape_PreservesEscapedDelims(t *testing.T) {
	// The literal delimiter is ";;". Here we ensure "%3B%3B" (escaped ";;") does not split.
	in := []byte("foo%3B%3Bbar;;baz")
	raw := splitTokens(in)
	if len(raw) != 2 {
		t.Fatalf("want 2 tokens, got %d", len(raw))
	}
	decoded, err := unescapeTokens(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decoded[0] != "foo;;bar" || decoded[1] != "baz" {
		t.Fatalf("decoded mismatch: %#v", decoded)
	}
}

// ------verifyTokenLength------------------------------------------------------

func TestVerifyTokenLength(t *testing.T) {
	ok := verifyTokenLength([]string{"a", "b", "c"}, 3, "CMD")
	if !ok {
		t.Fatalf("expected true when token length matches")
	}

	ok = verifyTokenLength([]string{"a", "b"}, 3, "CMD")
	if ok {
		t.Fatalf("expected false when token length mismatches")
	}
}
