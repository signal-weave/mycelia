package comm

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"mycelia/globals"
	"mycelia/test"
)

func TestWriteThenReadFrameU32_RoundTrip(t *testing.T) {
	test.WithTimeout(t, 2*time.Second, func(t *testing.T) {
		c1, c2 := net.Pipe()
		defer c1.Close()
		defer c2.Close()

		want := []byte("hello, frame")

		// writer goroutine
		go func() {
			_ = WriteFrameU32(c1, want)
			_ = c1.Close()
		}()

		got, err := ReadFrameU32(c2)
		if err != nil {
			t.Fatalf("ReadFrameU32 error: %v", err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("payload mismatch: got %q, want %q", got, want)
		}
	})
}

func TestReadFrameU32_ZeroLengthReturnsNil(t *testing.T) {
	test.WithTimeout(t, 2*time.Second, func(t *testing.T) {
		c1, c2 := net.Pipe()
		defer c1.Close()
		defer c2.Close()

		// Send 4-byte header with length 0
		var hdr [4]byte
		binary.BigEndian.PutUint32(hdr[:], 0)
		go func() {
			_, _ = c1.Write(hdr[:])
			_ = c1.Close()
		}()

		got, err := ReadFrameU32(c2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Fatalf("expected nil body for zero-length frame, got %v", got)
		}
	})
}

func TestReadFrameU32_TooLargeRejected(t *testing.T) {
	test.WithTimeout(t, 2*time.Second, func(t *testing.T) {
		c1, c2 := net.Pipe()
		defer c1.Close()
		defer c2.Close()

		// length > globals.BytesInMegabyte triggers sanity-limit error
		var hdr [4]byte
		binary.BigEndian.PutUint32(hdr[:], uint32(globals.BytesInMegabyte+1))

		go func() {
			// no payload necessary; reader fails afterheader
			_, _ = c1.Write(hdr[:])
			_ = c1.Close()
		}()

		got, err := ReadFrameU32(c2)
		if err == nil {
			t.Fatalf("expected error for oversize frame, got payload=%v", got)
		}
	})
}

func TestReadFrameU32_PartialPayload_EOF(t *testing.T) {
	test.WithTimeout(t, 2*time.Second, func(t *testing.T) {
		c1, c2 := net.Pipe()
		defer c1.Close()
		defer c2.Close()

		var hdr [4]byte
		binary.BigEndian.PutUint32(hdr[:], 5) // claim 5 bytes payload

		go func() {
			_, _ = c1.Write(hdr[:])
			_, _ = c1.Write([]byte{1, 2}) // write only 2 bytes, then close
			_ = c1.Close()
		}()

		got, err := ReadFrameU32(c2)
		if err == nil {
			t.Fatalf("expected io.ReadFull error, got payload=%v", got)
		}
		if err != io.ErrUnexpectedEOF && err != io.EOF {
			// io.ReadFull may wrap; allow either EOF or UnexpectedEOF
			// depending on timing
			t.Fatalf("expected EOF/UnexpectedEOF, got %v", err)
		}
	})
}

func TestWriteFrameU32_WritesLengthPrefix(t *testing.T) {
	test.WithTimeout(t, 2*time.Second, func(t *testing.T) {
		c1, c2 := net.Pipe()
		defer c1.Close()
		defer c2.Close()

		payload := []byte("abcde") // len=5

		// writer
		go func() {
			_ = WriteFrameU32(c1, payload)
			_ = c1.Close()
		}()

		// read raw header then body from the other end
		var hdr [4]byte
		if _, err := io.ReadFull(c2, hdr[:]); err != nil {
			t.Fatalf("reading header failed: %v", err)
		}
		n := binary.BigEndian.Uint32(hdr[:])
		if n != uint32(len(payload)) {
			t.Fatalf("header length = %d, want %d", n, len(payload))
		}

		body := make([]byte, n)
		if _, err := io.ReadFull(c2, body); err != nil {
			t.Fatalf("reading body failed: %v", err)
		}
		if !bytes.Equal(body, payload) {
			t.Fatalf("body mismatch: got %q, want %q", body, payload)
		}
	})
}
