package protocol

import (
	"bytes"
	"encoding/binary"
	"math"
	"strings"
	"testing"

	"mycelia/globals"
)

// -------small helpers---------------------------------------------------------

func be16(n uint16) []byte {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], n)
	return b[:]
}

func be32(n uint32) []byte {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], n)
	return b[:]
}

// -------readU8 / readU32 -----------------------------------------------------

func TestReadU8_Success(t *testing.T) {
	var out uint8
	err := readU8(bytes.NewReader([]byte{0xAB}), &out)
	if err != nil {
		t.Fatalf("readU8 error: %v", err)
	}
	if out != 0xAB {
		t.Fatalf("want 0xAB, got 0x%02X", out)
	}
}

func TestReadU8_EOF(t *testing.T) {
	var out uint8
	err := readU8(bytes.NewReader(nil), &out)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "EOF") {
		t.Fatalf("expected EOF in error, got: %v", err)
	}
}

func TestReadU32_Success(t *testing.T) {
	var out uint32
	data := be32(0x01020304)
	err := readU32(bytes.NewReader(data), &out)
	if err != nil {
		t.Fatalf("readU32 error: %v", err)
	}
	if out != 0x01020304 {
		t.Fatalf("want 0x01020304, got 0x%08X", out)
	}
}

// -------readStringU8 / U16 / U32 ---------------------------------------------

func TestReadStringU8_ZeroLength(t *testing.T) {
	// length=0 → returns empty string, nil error
	data := []byte{0x00}
	s, err := readStringU8(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("readStringU8 error: %v", err)
	}
	if s != "" {
		t.Fatalf("want empty string, got %q", s)
	}
}

func TestReadStringU8_Success(t *testing.T) {
	payload := []byte("hello")
	data := append([]byte{uint8(len(payload))}, payload...)
	s, err := readStringU8(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("readStringU8 error: %v", err)
	}
	if s != "hello" {
		t.Fatalf("want %q, got %q", "hello", s)
	}
}

func TestReadStringU8_Truncated(t *testing.T) {
	// declare 5, give only 3
	data := append([]byte{5}, []byte("hey")...)
	_, err := readStringU8(bytes.NewReader(data))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "read string bytes") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadStringU16_ZeroLength(t *testing.T) {
	data := be16(0)
	s, err := readStringU16(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("readStringU16 error: %v", err)
	}
	if s != "" {
		t.Fatalf("want empty string, got %q", s)
	}
}

func TestReadStringU16_TooLarge_IfRepresentable(t *testing.T) {
	// Limit from implementation: 64*globals.BytesInKilobyte - 1
	u16Limit := uint16(64*globals.BytesInKilobyte - 1)
	if u16Limit == math.MaxUint16 {
		// Can't represent a value > limit in uint16; skip this edge-case.
		t.Skip("u16 safety limit equals MaxUint16; cannot construct representable overflow")
	}
	decl := be16(u16Limit + 1)
	_, err := readStringU16(bytes.NewReader(decl))
	if err == nil {
		t.Fatalf("expected limit error, got nil")
	}
	if !strings.Contains(err.Error(), "declared length exceeds 64KB safety limit") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadStringU32_TooLarge(t *testing.T) {
	// Limit from implementation: 64 * globals.BytesInMegabyte
	readLimit := uint32(64 * globals.BytesInMegabyte)
	decl := be32(readLimit + 1)
	_, err := readStringU32(bytes.NewReader(decl))
	if err == nil {
		t.Fatalf("expected limit error, got nil")
	}
	if !strings.Contains(err.Error(), "declared length exceeds 64MB safety limit") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadStringU32_Truncated(t *testing.T) {
	// declare 4, provide only 2
	data := append(be32(4), []byte("hi")...)
	_, err := readStringU32(bytes.NewReader(data))
	if err == nil {
		t.Fatalf("expected truncation error, got nil")
	}
	if !strings.Contains(err.Error(), "read string bytes") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// -------readBytesU16 / U32----------------------------------------------------

func TestReadBytesU16_Success(t *testing.T) {
	body := []byte{0xAA, 0xBB, 0xCC}
	data := append(be16(uint16(len(body))), body...)
	got, err := readBytesU16(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("readBytesU16 error: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("payload mismatch: got %v want %v", got, body)
	}
}

func TestReadBytesU16_Truncated(t *testing.T) {
	body := []byte{0xAA}
	decl := be16(3) // declare 3, give 1
	data := append(decl, body...)
	_, err := readBytesU16(bytes.NewReader(data))
	if err == nil {
		t.Fatalf("expected truncation error, got nil")
	}
	if !strings.Contains(err.Error(), "read payload bytes") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadBytesU32_Success(t *testing.T) {
	body := []byte{1, 2, 3, 4, 5}
	data := append(be32(uint32(len(body))), body...)
	got, err := readBytesU32(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("readBytesU32 error: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("payload mismatch: got %v want %v", got, body)
	}
}

func TestReadBytesU32_TooLarge(t *testing.T) {
	readLimit := uint32(64 * globals.BytesInMegabyte)
	decl := be32(readLimit + 1)
	_, err := readBytesU32(bytes.NewReader(decl))
	if err == nil {
		t.Fatalf("expected limit error, got nil")
	}
	if !strings.Contains(err.Error(), "declared length exceeds 64MB safety limit") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// -------readU8Len / U16Len / U32Len-------------------------------------------

func TestReadU8Len_Success(t *testing.T) {
	data := []byte{200}
	n, err := readU8Len(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("readU8Len error: %v", err)
	}
	if n != 200 {
		t.Fatalf("want 200, got %d", n)
	}
}

func TestReadU8Len_EOF(t *testing.T) {
	_, err := readU8Len(bytes.NewReader(nil))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "read length") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadU16Len_TooLarge_IfRepresentable(t *testing.T) {
	u16Limit := uint16(64*globals.BytesInKilobyte - 1)
	if u16Limit == math.MaxUint16 {
		t.Skip("u16 safety limit equals MaxUint16; cannot construct representable overflow")
	}
	decl := be16(u16Limit + 1)
	_, err := readU16Len(bytes.NewReader(decl))
	if err == nil {
		t.Fatalf("expected limit error, got nil")
	}
	if !strings.Contains(err.Error(), "declared length exceeds 64KB safety limit") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadU32Len_TooLarge(t *testing.T) {
	readLimit := uint32(64 * globals.BytesInMegabyte)
	decl := be32(readLimit + 1)
	_, err := readU32Len(bytes.NewReader(decl))
	if err == nil {
		t.Fatalf("expected limit error, got nil")
	}
	if !strings.Contains(err.Error(), "declared length exceeds 64MB safety limit") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadU32Len_Success(t *testing.T) {
	decl := be32(123456)
	n, err := readU32Len(bytes.NewReader(decl))
	if err != nil {
		t.Fatalf("readU32Len error: %v", err)
	}
	if n != 123456 {
		t.Fatalf("want 123456, got %d", n)
	}
}
