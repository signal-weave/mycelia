package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

// --- helpers --------------------------------------------------------------

func beU32(n uint32) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, n)
	return buf.Bytes()
}

func prependLenBE(body []byte) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(body)))
	buf.Write(body)
	return buf.Bytes()
}

// --- readU32 --------------------------------------------------------------

func TestReadU32_OK(t *testing.T) {
	r := bytes.NewReader(beU32(42))
	var out uint32
	if err := readU32(r, &out); err != nil {
		t.Fatalf("readU32 error: %v", err)
	}
	if out != 42 {
		t.Fatalf("expected 42, got %d", out)
	}
}

func TestReadU32_ShortBuffer(t *testing.T) {
	r := bytes.NewReader([]byte{0x00, 0x00}) // < 4 bytes
	var out uint32
	err := readU32(r, &out)
	if err == nil {
		t.Fatalf("expected error for short buffer, got nil")
	}
}

// --- readLen --------------------------------------------------------------

func TestReadLen_OK(t *testing.T) {
	r := bytes.NewReader(beU32(5))
	n, err := readLen(r)
	if err != nil {
		t.Fatalf("readLen error: %v", err)
	}
	if n != 5 {
		t.Fatalf("expected 5, got %d", n)
	}
}

func TestReadLen_ShortBuffer(t *testing.T) {
	r := bytes.NewReader([]byte{0x00, 0x00, 0x00}) // < 4 bytes
	_, err := readLen(r)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	// error is wrapped with "read length: %w"
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}
}

func TestReadLen_TooLarge(t *testing.T) {
	const tooBig = 64*1024*1024 + 1 // 64MB + 1
	r := bytes.NewReader(beU32(tooBig))
	_, err := readLen(r)
	if err == nil {
		t.Fatalf("expected error for too-large length, got nil")
	}
	if got := err.Error(); got != "declared length exceeds 64MB safety limit" {
		t.Fatalf("unexpected error: %q", got)
	}
}

// --- readString -----------------------------------------------------------

func TestReadString_Empty(t *testing.T) {
	// length=0
	r := bytes.NewReader(beU32(0))
	s, err := readString(r)
	if err != nil {
		t.Fatalf("readString error: %v", err)
	}
	if s != "" {
		t.Fatalf("expected empty string, got %q", s)
	}
}

func TestReadString_OK(t *testing.T) {
	body := []byte("test")
	r := bytes.NewReader(prependLenBE(body))
	s, err := readString(r)
	if err != nil {
		t.Fatalf("readString error: %v", err)
	}
	if s != "test" {
		t.Fatalf("expected %q, got %q", "test", s)
	}
}

func TestReadString_Truncated(t *testing.T) {
	// declare len=4, provide only 3 bytes
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, uint32(4))
	buf.Write([]byte("tes"))
	r := bytes.NewReader(buf.Bytes())

	_, err := readString(r)
	if err == nil {
		t.Fatalf("expected error for truncated string, got nil")
	}
	// wrapped with "read string bytes: %w"
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}
}

// --- readBytes ------------------------------------------------------------

func TestReadBytes_Empty(t *testing.T) {
	// length=0
	r := bytes.NewReader(beU32(0))
	b, err := readBytes(r)
	if err != nil {
		t.Fatalf("readBytes error: %v", err)
	}
	if b != nil {
		t.Fatalf("expected nil slice for zero length, got %#v", b)
	}
}

func TestReadBytes_OK(t *testing.T) {
	body := []byte{0x01, 0x02, 0x03}
	r := bytes.NewReader(prependLenBE(body))
	b, err := readBytes(r)
	if err != nil {
		t.Fatalf("readBytes error: %v", err)
	}
	if !bytes.Equal(b, body) {
		t.Fatalf("expected %v, got %v", body, b)
	}
}

func TestReadBytes_Truncated(t *testing.T) {
	// declare len=3, provide only 2 bytes
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, uint32(3))
	buf.Write([]byte{0xAA, 0xBB})
	r := bytes.NewReader(buf.Bytes())

	_, err := readBytes(r)
	if err == nil {
		t.Fatalf("expected error for truncated bytes, got nil")
	}
	// wrapped with "read payload bytes: %w"
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}
}
