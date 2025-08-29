package test

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"
)

// -----------------------------------------------------------------------------
// Herein is a helper library for testing writing to byte buffers.
// -----------------------------------------------------------------------------

// Helpers to build length-prefixed (uint32 big-endian) strings/bytes.
func WriteU8(buf *bytes.Buffer, v uint8, t *testing.T) {
	t.Helper()
	if err := binary.Write(buf, binary.BigEndian, v); err != nil {
		t.Fatalf("writeU32 failed: %v", err)
	}
}

// Helpers to build length-prefixed (uint32 big-endian) strings/bytes.
func WriteU32(buf *bytes.Buffer, v uint32, t *testing.T) {
	t.Helper()
	if err := binary.Write(buf, binary.BigEndian, v); err != nil {
		t.Fatalf("writeU32 failed: %v", err)
	}
}

func WriteString(buf *bytes.Buffer, s string, t *testing.T) {
	t.Helper()
	WriteU32(buf, uint32(len(s)), t)
	if _, err := buf.Write([]byte(s)); err != nil {
		t.Fatalf("writeString failed: %v", err)
	}
}

func WriteBytes(buf *bytes.Buffer, b []byte, t *testing.T) {
	t.Helper()
	WriteU32(buf, uint32(len(b)), t)
	if _, err := buf.Write(b); err != nil {
		t.Fatalf("writeBytes failed: %v", err)
	}
}

func TypeName(v any) string {
	if v == nil {
		return "<nil>"
	}
	return reflect.TypeOf(v).String()
}
